package httpapi

import (
	"crypto/subtle"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"ponpes-saas/internal/authz"
	"ponpes-saas/internal/questions"
	"ponpes-saas/internal/repository"
)

func eventInt64(value any) int64 {
	switch item := value.(type) {
	case float64:
		return int64(item)
	case int64:
		return item
	case string:
		parsed, _ := strconv.ParseInt(item, 10, 64)
		return parsed
	default:
		parsed, _ := strconv.ParseInt(fmt.Sprint(item), 10, 64)
		return parsed
	}
}

// hermesEvent logs a Telegram question (agent:start) or answer (agent:end) sent by the Hermes container hook.
func (s *Server) hermesEvent(c *fiber.Ctx) error {
	if s.cfg.HermesAPIKey == "" || subtle.ConstantTimeCompare([]byte(c.Get("X-Hermes-Event-Key")), []byte(s.cfg.HermesAPIKey)) != 1 {
		return fiber.NewError(http.StatusUnauthorized, "INVALID_HERMES_EVENT_KEY")
	}
	var input struct {
		ProfileKey string         `json:"profile_key"`
		Event      string         `json:"event"`
		Context    map[string]any `json:"context"`
	}
	if err := c.BodyParser(&input); err != nil || input.ProfileKey == "" || input.Context == nil {
		return fiber.NewError(http.StatusBadRequest, "INVALID_HERMES_EVENT")
	}
	platform, _ := input.Context["platform"].(string)
	if (platform != "telegram" && platform != "whatsapp") || (input.Event != "agent:start" && input.Event != "agent:end") {
		return c.SendStatus(http.StatusNoContent)
	}
	tenantID, err := s.repo.TenantByHermesProfileKey(c.Context(), input.ProfileKey)
	if err != nil {
		return fiber.NewError(http.StatusNotFound, "HERMES_PROFILE_NOT_FOUND")
	}
	channel := "TELEGRAM"
	chatID := eventInt64(input.Context["chat_id"])
	userID := eventInt64(input.Context["user_id"])
	username, _ := input.Context["username"].(string)
	if platform == "whatsapp" {
		// WhatsApp IDs look like "62812...@s.whatsapp.net"; the digits identify the chat and name the visitor.
		channel = "WHATSAPP"
		chatDigits, userDigits := jidDigits(fmt.Sprint(input.Context["chat_id"])), jidDigits(fmt.Sprint(input.Context["user_id"]))
		chatID, userID = eventInt64(chatDigits), eventInt64(userDigits)
		if username == "" && userDigits != "" {
			username = "+" + userDigits
		}
	}
	if chatID == 0 {
		return c.SendStatus(http.StatusNoContent)
	}
	conversation, err := s.repo.UpsertConversation(c.Context(), tenantID, channel, chatID, userID, username)
	if err != nil {
		return err
	}
	text, _ := input.Context["message"].(string)
	sender := "USER"
	if input.Event == "agent:end" {
		text, _ = input.Context["response"].(string)
		sender = "BOT"
	}
	if strings.TrimSpace(text) == "" {
		return c.SendStatus(http.StatusNoContent)
	}
	if _, err := s.repo.AddTelegramMessage(c.Context(), conversation.ID, sender, text, nil); err != nil {
		return err
	}
	// The assistant says a fixed sentence when it has no answer; file the visitor's question for an admin to answer.
	if sender == "BOT" && questions.IsUnanswered(text) {
		if question := s.repo.LastUserMessage(c.Context(), conversation.ID); question != "" {
			if err := s.repo.RecordUnanswered(c.Context(), tenantID, question); err != nil {
				slog.Warn("could not record unanswered question", "tenant_id", tenantID, "error", err)
			}
		}
	}
	return c.SendStatus(http.StatusNoContent)
}

// devTelegramMessage injects a fake user message; it only exists in development.
func (s *Server) devTelegramMessage(c *fiber.Ctx) error {
	if os.Getenv("APP_ENV") != "development" {
		return fiber.NewError(http.StatusNotFound, "NOT_FOUND")
	}
	id, err := tenantIDParam(c)
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "INVALID_TENANT_ID")
	}
	var input struct {
		ChatID    int64  `json:"chat_id"`
		UserID    int64  `json:"user_id"`
		Username  string `json:"username"`
		Text      string `json:"text"`
		MessageID *int64 `json:"message_id"`
	}
	if err = c.BodyParser(&input); err != nil || input.ChatID == 0 || strings.TrimSpace(input.Text) == "" {
		return fiber.NewError(http.StatusBadRequest, "INVALID_TELEGRAM_MESSAGE")
	}
	conversation, err := s.repo.UpsertConversation(c.Context(), id, "TELEGRAM", input.ChatID, input.UserID, strings.TrimSpace(input.Username))
	if err != nil {
		return err
	}
	message, err := s.repo.AddTelegramMessage(c.Context(), conversation.ID, "USER", strings.TrimSpace(input.Text), input.MessageID)
	if err != nil {
		return err
	}
	return c.Status(http.StatusCreated).JSON(fiber.Map{"conversation": conversation, "message": message, "mode": "development"})
}

func (s *Server) exportTelegramLog(c *fiber.Ctx) error {
	tenantID, err := tenantIDParam(c)
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "INVALID_TENANT_ID")
	}
	filter, err := telegramExportFilter(c)
	if err != nil {
		return err
	}
	rows, err := s.repo.TelegramExportRows(c.Context(), tenantID, filter.rows)
	if err != nil {
		return err
	}
	file, err := telegramWorkbook(filterExportTurns(pairTelegramExportRows(rows), filter.from, filter.until))
	if err != nil {
		return fmt.Errorf("telegram export failed: %w", err)
	}
	c.Set(fiber.HeaderContentType, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set(fiber.HeaderContentDisposition, fmt.Sprintf(`attachment; filename="log-telegram-%s.xlsx"`, tenantID.String()))
	return c.Send(file)
}

// telegramConversationGroups serves the grouped Telegram log: one page of conversations with question counts.
func (s *Server) telegramConversationGroups(c *fiber.Ctx) error {
	tenantID, err := tenantIDParam(c)
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "INVALID_TENANT_ID")
	}
	size := c.QueryInt("size", 10)
	if size != 10 && size != 20 && size != 50 {
		size = 10
	}
	page := max(c.QueryInt("page", 1), 1)
	items, total, statuses, err := s.repo.ConversationGroups(c.Context(), tenantID, repository.ConversationGroupQuery{
		Search: c.Query("q"), Status: c.Query("status"), Sort: c.Query("sort"), Limit: size, Offset: (page - 1) * size,
	})
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"data": items, "total": total, "statuses": statuses})
}

// scopedTenant parses the optional tenant_id query and checks the caller may read it.
func (s *Server) scopedTenant(c *fiber.Ctx) (*uuid.UUID, repository.User, error) {
	user, userID, err := s.sessionUser(c)
	if err != nil {
		return nil, user, err
	}
	value := c.Query("tenant_id")
	if value == "" {
		return nil, user, nil
	}
	id, err := uuid.Parse(value)
	if err != nil {
		return nil, user, fiber.NewError(http.StatusBadRequest, "INVALID_TENANT_ID")
	}
	if !user.Superadmin {
		if _, roleErr := s.repo.TenantRole(c.Context(), userID, id); roleErr != nil {
			return nil, user, fiber.NewError(http.StatusForbidden, "FORBIDDEN")
		}
	}
	return &id, user, nil
}

func (s *Server) listConversations(c *fiber.Ctx) error {
	tenantID, _, err := s.scopedTenant(c)
	if err != nil {
		return err
	}
	items, err := s.repo.Conversations(c.Context(), tenantID, c.Query("q"))
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"data": items})
}

func (s *Server) conversationStats(c *fiber.Ctx) error {
	scope, user, err := s.scopedTenant(c)
	if err != nil {
		return err
	}
	// Without tenant_id the stats cover every pesantren, which only a superadmin may see.
	if scope == nil && !user.Superadmin {
		return fiber.NewError(http.StatusForbidden, "FORBIDDEN")
	}
	days := c.QueryInt("days", 14)
	if days < 1 || days > 90 {
		days = 14
	}
	stats, err := s.repo.ConversationStats(c.Context(), scope, days)
	if err != nil {
		return err
	}
	return c.JSON(stats)
}

func (s *Server) setConversationStatus(c *fiber.Ctx) error {
	conversationID, err := uuid.Parse(c.Params("conversation_id"))
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "INVALID_CONVERSATION_ID")
	}
	var input struct {
		Status string `json:"status"`
	}
	if err := c.BodyParser(&input); err != nil || (input.Status != "ACTIVE" && input.Status != "HANDOFF" && input.Status != "RESOLVED") {
		return fiber.NewError(http.StatusBadRequest, "INVALID_CONVERSATION_STATUS")
	}
	conversationTenant, err := s.repo.ConversationTenant(c.Context(), conversationID)
	if err != nil {
		return fiber.NewError(http.StatusNotFound, "CONVERSATION_NOT_FOUND")
	}
	user, userID, err := s.sessionUser(c)
	if err != nil {
		return err
	}
	if !user.Superadmin {
		role, roleErr := s.repo.TenantRole(c.Context(), userID, conversationTenant)
		if roleErr != nil || !authz.Allows(role, "conversation:manage") {
			return fiber.NewError(http.StatusForbidden, "FORBIDDEN")
		}
	}
	item, err := s.repo.SetConversationStatus(c.Context(), conversationID, input.Status)
	if err != nil {
		return fiber.NewError(http.StatusNotFound, "CONVERSATION_NOT_FOUND")
	}
	label := "Pengguna Telegram"
	if item.Username != "" {
		label = "@" + item.Username
	}
	s.repo.RecordAudit(c.Context(), item.TenantID, &userID, "CONVERSATION_"+input.Status, "CONVERSATION", item.ID.String(), label)
	return c.JSON(item)
}

func (s *Server) conversationMessages(c *fiber.Ctx) error {
	conversationID, err := uuid.Parse(c.Params("conversation_id"))
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "INVALID_CONVERSATION_ID")
	}
	conversationTenant, err := s.repo.ConversationTenant(c.Context(), conversationID)
	if err != nil {
		return fiber.NewError(http.StatusNotFound, "CONVERSATION_NOT_FOUND")
	}
	user, userID, err := s.sessionUser(c)
	if err != nil {
		return fiber.NewError(http.StatusForbidden, "FORBIDDEN")
	}
	if !user.Superadmin {
		if _, roleErr := s.repo.TenantRole(c.Context(), userID, conversationTenant); roleErr != nil {
			return fiber.NewError(http.StatusForbidden, "FORBIDDEN")
		}
	}
	items, err := s.repo.ConversationMessages(c.Context(), conversationID)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"data": items})
}

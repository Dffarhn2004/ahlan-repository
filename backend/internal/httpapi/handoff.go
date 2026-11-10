package httpapi

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"ponpes-saas/internal/repository"
)

// handoffRequest is what the Hermes containers send to take a WhatsApp conversation over to a human.
//
//	action "start": the assistant routed the visitor to an admin; the conversation becomes HANDOFF.
//	action "check": a visitor wrote again; the answer says whether the assistant must stay silent. While the
//	               conversation is in HANDOFF the message is logged here, because the assistant never sees it.
type handoffRequest struct {
	ProfileKey string `json:"profile_key"`
	Action     string `json:"action"`
	ChatID     string `json:"chat_id"`
	UserID     string `json:"user_id"`
	Text       string `json:"text"`
}

// handoffIdleLimit is how long a hand-off stays in force without any visitor message.
const handoffIdleLimit = 12 * time.Hour

// hermesHandoff is called by the notify hook and the handoff plugin inside the Hermes container. It is
// authenticated like hermesEvent, with the X-Hermes-Event-Key secret.
func (s *Server) hermesHandoff(c *fiber.Ctx) error {
	if s.cfg.HermesAPIKey == "" || subtle.ConstantTimeCompare([]byte(c.Get("X-Hermes-Event-Key")), []byte(s.cfg.HermesAPIKey)) != 1 {
		return fiber.NewError(http.StatusUnauthorized, "INVALID_HERMES_EVENT_KEY")
	}
	var in handoffRequest
	if err := c.BodyParser(&in); err != nil || in.ProfileKey == "" || in.ChatID == "" {
		return fiber.NewError(http.StatusBadRequest, "INVALID_HERMES_EVENT")
	}
	tenantID, err := s.repo.TenantByHermesProfileKey(c.Context(), in.ProfileKey)
	if err != nil {
		return fiber.NewError(http.StatusNotFound, "HERMES_PROFILE_NOT_FOUND")
	}
	chatDigits, userDigits := jidDigits(in.ChatID), jidDigits(in.UserID)
	chatID, userID := eventInt64(chatDigits), eventInt64(userDigits)
	if chatID == 0 {
		return fiber.NewError(http.StatusBadRequest, "INVALID_HERMES_EVENT")
	}
	switch in.Action {
	case "start":
		username := ""
		if userDigits != "" {
			username = "+" + userDigits
		}
		conversation, err := s.repo.UpsertConversation(c.Context(), tenantID, "WHATSAPP", chatID, userID, username)
		if err != nil {
			return err
		}
		if _, err := s.repo.SetConversationStatus(c.Context(), conversation.ID, "HANDOFF"); err != nil {
			return err
		}
		return c.JSON(fiber.Map{"handoff": true})
	case "end":
		// The visitor started a new conversation (/new): release the hand-off so the assistant answers again.
		conversation, err := s.repo.FindConversation(c.Context(), tenantID, "WHATSAPP", chatID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return c.JSON(fiber.Map{"handoff": false})
			}
			return err
		}
		if conversation.Status == "HANDOFF" {
			if _, err := s.repo.SetConversationStatus(c.Context(), conversation.ID, "ACTIVE"); err != nil {
				return err
			}
		}
		return c.JSON(fiber.Map{"handoff": false})
	case "check":
		conversation, err := s.repo.FindConversation(c.Context(), tenantID, "WHATSAPP", chatID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return c.JSON(fiber.Map{"handoff": false})
			}
			return err
		}
		if conversation.Status != "HANDOFF" {
			return c.JSON(fiber.Map{"handoff": false})
		}
		// A forgotten hand-off must not silence the assistant forever: after handoffIdleLimit without a visitor
			// message the conversation goes back to ACTIVE and the assistant answers again.
			if time.Since(conversation.LastMessageAt) > handoffIdleLimit {
				if _, err := s.repo.SetConversationStatus(c.Context(), conversation.ID, "ACTIVE"); err != nil {
					return err
				}
				return c.JSON(fiber.Map{"handoff": false})
			}
			if err := s.repo.TouchConversation(c.Context(), conversation.ID); err != nil {
				slog.Warn("could not refresh a hand-off", "tenant_id", tenantID, "error", err)
			}
			if text := strings.TrimSpace(in.Text); text != "" {
			if _, err := s.repo.AddTelegramMessage(c.Context(), conversation.ID, "USER", text, nil); err != nil {
				slog.Warn("could not log a handed-off message", "tenant_id", tenantID, "error", err)
			}
		}
		return c.JSON(fiber.Map{"handoff": true})
	}
	return fiber.NewError(http.StatusBadRequest, "INVALID_HANDOFF_ACTION")
}

// handoffURL is where the Hermes container reports and checks hand-offs: next to the events endpoint.
func (s *Server) handoffURL() string {
	return fmt.Sprintf("%s/handoff", strings.TrimSuffix(strings.TrimRight(s.eventsURL(), "/"), "/events"))
}

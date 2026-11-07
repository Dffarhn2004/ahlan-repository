package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"ponpes-saas/internal/contacts"
	"ponpes-saas/internal/repository"
)

// maxWhatsAppAdmins keeps the routing section of the persona short.
const maxWhatsAppAdmins = 10

func (s *Server) listWhatsAppAdmins(c *fiber.Ctx) error {
	id, err := tenantIDParam(c)
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "INVALID_TENANT_ID")
	}
	items, err := s.repo.ListWhatsAppAdmins(c.Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"data": items})
}

// saveWhatsAppAdmins replaces the admin list, then rewrites the persona and restarts a running assistant.
func (s *Server) saveWhatsAppAdmins(c *fiber.Ctx) error {
	id, err := tenantIDParam(c)
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "INVALID_TENANT_ID")
	}
	var in struct {
		Data []repository.WhatsAppAdmin `json:"data"`
	}
	if err := c.BodyParser(&in); err != nil {
		return fiber.NewError(http.StatusBadRequest, "INVALID_REQUEST")
	}
	if len(in.Data) > maxWhatsAppAdmins {
		return fiber.NewError(http.StatusBadRequest, "WHATSAPP_ADMIN_LIMIT")
	}
	clean := make([]repository.WhatsAppAdmin, 0, len(in.Data))
	seen := map[string]bool{}
	for _, item := range in.Data {
		item.Name = strings.TrimSpace(item.Name)
		item.NotifyWhen = strings.TrimSpace(item.NotifyWhen)
		phone, phoneErr := contacts.NormalizePhone(item.Phone)
		if phoneErr != nil {
			return fiber.NewError(http.StatusBadRequest, "WHATSAPP_ADMIN_PHONE_INVALID")
		}
		if n := utf8.RuneCountInString(item.Name); n < 1 || n > 80 {
			return fiber.NewError(http.StatusBadRequest, "WHATSAPP_ADMIN_NAME_INVALID")
		}
		if n := utf8.RuneCountInString(item.NotifyWhen); n < 1 || n > 200 {
			return fiber.NewError(http.StatusBadRequest, "WHATSAPP_ADMIN_WHEN_INVALID")
		}
		if seen[phone] {
			return fiber.NewError(http.StatusBadRequest, "WHATSAPP_ADMIN_DUPLICATE")
		}
		seen[phone] = true
		item.Phone = phone
		clean = append(clean, item)
	}
	previous, _ := s.repo.ListWhatsAppAdmins(c.Context(), id)
	if err := s.repo.ReplaceWhatsAppAdmins(c.Context(), id, clean); err != nil {
		return err
	}
	if profileKey, keyErr := s.repo.HermesProfileKey(c.Context(), id); keyErr == nil {
		s.refreshPersona(c.Context(), profileKey)
		// The notify hook is only installed once the first admin exists; the gateway reads hooks at start-up only.
		s.applyHookChange(c.Context(), id, profileKey, len(previous) == 0 && len(clean) > 0)
	}
	s.repo.RecordAudit(c.Context(), id, auditActor(c), "WHATSAPP_ADMINS_SAVED", "WHATSAPP_CHANNEL", id.String(), "")
	return c.JSON(fiber.Map{"data": clean})
}

// applyHookChange makes a running assistant pick up a change. A newly installed hook only loads when the shared
// gateway restarts (every tenant's bot is briefly offline, and the bridge limits how often it may happen); anything
// else just restarts this tenant's own assistant.
func (s *Server) applyHookChange(ctx context.Context, tenantID uuid.UUID, profileKey string, hookInstalled bool) {
	if hookInstalled && s.assistantActive(ctx, tenantID) {
		if err := s.docker().RestartGateway(ctx); err != nil {
			slog.Warn("could not restart the gateway for the new WhatsApp notify hook; restart it manually", "profile", profileKey, "error", err)
		}
		return
	}
	s.restartIfActive(ctx, tenantID, profileKey)
}

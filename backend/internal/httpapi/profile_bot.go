package httpapi

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"ponpes-saas/internal/crawler"
	"ponpes-saas/internal/links"
	"ponpes-saas/internal/repository"
)

func (s *Server) getProfile(c *fiber.Ctx) error {
	id, err := tenantIDParam(c)
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "INVALID_TENANT_ID")
	}
	profile, err := s.repo.GetProfile(c.Context(), id)
	if err != nil {
		return fiber.NewError(http.StatusNotFound, "PROFILE_NOT_FOUND")
	}
	return c.JSON(profile)
}

func (s *Server) saveProfile(c *fiber.Ctx) error {
	id, err := tenantIDParam(c)
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "INVALID_TENANT_ID")
	}
	// Start from the stored profile so fields the client omits are kept, not blanked.
	in, _ := s.repo.GetProfile(c.Context(), id)
	if err := c.BodyParser(&in); err != nil || strings.TrimSpace(in.DisplayName) == "" {
		return fiber.NewError(http.StatusBadRequest, "INVALID_REQUEST")
	}
	if err := crawler.ValidateMapsURL(in.GoogleMapsURL); err != nil {
		return fiber.NewError(http.StatusUnprocessableEntity, "MAPS_URL_INVALID")
	}
	// Each social page must be an https address on that network's own domain; empty clears it.
	for _, social := range []struct {
		kind  string
		value *string
	}{
		{"INSTAGRAM", &in.InstagramURL}, {"FACEBOOK", &in.FacebookURL}, {"TIKTOK", &in.TiktokURL}, {"YOUTUBE", &in.YoutubeURL},
	} {
		if strings.TrimSpace(*social.value) == "" {
			*social.value = ""
			continue
		}
		clean, err := links.Validate(social.kind, *social.value)
		if err != nil {
			return fiber.NewError(http.StatusUnprocessableEntity, "SOCIAL_URL_INVALID")
		}
		*social.value = clean
	}
	in.TenantID = id
	actor := c.Locals("user_id").(uuid.UUID)
	if err := s.repo.UpsertProfile(c.Context(), in, actor); err != nil {
		return err
	}
	if key, keyErr := s.repo.HermesProfileKey(c.Context(), id); keyErr == nil {
		s.refreshPersona(c.Context(), key)
	}
	s.repo.RecordAudit(c.Context(), id, &actor, "PROFILE_UPDATED", "TENANT", id.String(), "")
	return c.JSON(in)
}

func (s *Server) getBot(c *fiber.Ctx) error {
	id, err := tenantIDParam(c)
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "INVALID_TENANT_ID")
	}
	bot, err := s.repo.GetTelegramBot(c.Context(), id)
	if err != nil {
		return fiber.NewError(http.StatusNotFound, "BOT_NOT_FOUND")
	}
	bot.CredentialRef = ""
	return c.JSON(bot)
}

func (s *Server) saveBot(c *fiber.Ctx) error {
	id, err := tenantIDParam(c)
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "INVALID_TENANT_ID")
	}
	var in repository.TelegramBot
	if err := c.BodyParser(&in); err != nil || in.BotID <= 0 || strings.TrimSpace(in.Username) == "" || strings.TrimSpace(in.CredentialRef) == "" {
		return fiber.NewError(http.StatusBadRequest, "INVALID_REQUEST")
	}
	in.AccessMode = strings.ToUpper(strings.TrimSpace(in.AccessMode))
	if in.AccessMode == "" {
		in.AccessMode = "TESTING"
	}
	if in.AccessMode != "TESTING" && in.AccessMode != "PUBLIC" {
		return fiber.NewError(http.StatusBadRequest, "INVALID_ACCESS_MODE")
	}
	if in.AccessMode == "PUBLIC" {
		in.AllowedUserIDs = []int64{}
	} else {
		seen := map[int64]bool{}
		ids := []int64{}
		for _, userID := range in.AllowedUserIDs {
			if userID <= 0 {
				return fiber.NewError(http.StatusBadRequest, "INVALID_TELEGRAM_USER_ID")
			}
			if !seen[userID] {
				seen[userID] = true
				ids = append(ids, userID)
			}
		}
		if len(ids) == 0 {
			return fiber.NewError(http.StatusBadRequest, "TELEGRAM_USER_IDS_REQUIRED")
		}
		in.AllowedUserIDs = ids
	}
	in.TenantID = id
	if err := s.repo.UpsertTelegramBot(c.Context(), in, c.Locals("user_id").(uuid.UUID)); err != nil {
		return err
	}
	in.CredentialRef = ""
	s.repo.RecordAudit(c.Context(), id, auditActor(c), "TELEGRAM_BOT_CONNECTED", "TELEGRAM_BOT", id.String(), "@"+strings.TrimPrefix(in.Username, "@")+" ("+strings.ToLower(in.AccessMode)+")")
	in.Status = "PENDING"
	return c.JSON(in)
}

func (s *Server) deleteBot(c *fiber.Ctx) error {
	id, err := tenantIDParam(c)
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "INVALID_TENANT_ID")
	}
	// Clear the token from the profile too, so a restart cannot bring the removed bot back.
	if profileKey, keyErr := s.repo.HermesProfileKey(c.Context(), id); keyErr == nil {
		if err := s.docker().RemoveChannelEnv(c.Context(), profileKey, "TELEGRAM_"); err != nil {
			slog.Warn("could not clear Telegram settings", "profile", profileKey, "error", err)
		}
	}
	if err := s.repo.DeleteTelegramBot(c.Context(), id); err != nil {
		return err
	}
	s.repo.RecordAudit(c.Context(), id, auditActor(c), "TELEGRAM_BOT_REMOVED", "TELEGRAM_BOT", id.String(), "")
	return c.SendStatus(http.StatusNoContent)
}

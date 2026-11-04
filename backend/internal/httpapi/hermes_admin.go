package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"ponpes-saas/internal/repository"
)

func (s *Server) hermesOverview(c *fiber.Ctx) error {
	overview, err := s.repo.HermesOverview(c.Context())
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"data": overview})
}

func (s *Server) hermesSkills(c *fiber.Ctx) error {
	skills, err := s.hermes.ListSkills(c.Context())
	if err != nil {
		slog.Warn("Hermes skills catalog unavailable", "error", err)
		return c.JSON(fiber.Map{"data": []map[string]any{}, "available": false})
	}
	return c.JSON(fiber.Map{"data": skills, "available": true})
}

func (s *Server) hermesConnection(c *fiber.Ctx) error {
	capabilities, err := s.hermes.Check(c.Context())
	if err != nil {
		return fiber.NewError(http.StatusBadGateway, "HERMES_UNAVAILABLE")
	}
	return c.JSON(fiber.Map{"status": "connected", "data": capabilities})
}

func (s *Server) getHermesSettings(c *fiber.Ctx) error {
	settings, err := s.repo.GetGlobalHermesSettings(c.Context())
	if err != nil {
		return err
	}
	return c.JSON(settings)
}

func (s *Server) saveHermesSettings(c *fiber.Ctx) error {
	var input repository.HermesSettings
	if err := c.BodyParser(&input); err != nil || len(input.Description) > 2000 || len(input.SoulMD) > 50000 || len(input.EnabledSkills) > 100 {
		return fiber.NewError(http.StatusBadRequest, "INVALID_HERMES_SETTINGS")
	}
	keys, err := s.repo.HermesProfileKeys(c.Context())
	if err != nil {
		return err
	}
	client := s.docker()
	applied := 0
	for _, key := range keys {
		if err := client.ApplyProfileSettings(c.Context(), key, s.profileSettings(c.Context(), key, input)); err != nil {
			return fiber.NewError(http.StatusBadGateway, "HERMES_SETTINGS_APPLY_FAILED")
		}
		applied++
	}
	if err := s.repo.SaveGlobalHermesSettings(c.Context(), input, c.Locals("user_id").(uuid.UUID)); err != nil {
		return err
	}
	// Running assistants keep the old persona until restarted, so restart them in the background.
	if profiles, listErr := s.repo.HermesProfiles(c.Context()); listErr == nil {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			defer cancel()
			for _, profile := range profiles {
				s.restartIfActive(ctx, profile.TenantID, profile.ProfileKey)
			}
		}()
	}
	return c.JSON(fiber.Map{"status": "APPLIED", "profiles_applied": applied, "data": input})
}

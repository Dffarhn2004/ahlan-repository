package httpapi

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"ponpes-saas/internal/repository"
)

// publicLink is the unauthenticated link-page payload. A button's URL is empty when it is unavailable.
type publicLink struct {
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MapsURL     string `json:"maps_url"`
	TelegramURL string `json:"telegram_url"`
	WhatsAppURL string `json:"whatsapp_url"`
	Address     string `json:"address"`
	// Socials holds website, instagram, youtube and tiktok URLs; a missing key means the pesantren has none.
	Socials map[string]string `json:"socials"`
	// LogoURL is empty when the pesantren has not uploaded a logo. Folders are the link categories shown publicly.
	LogoURL string                        `json:"logo_url"`
	Folders []repository.PublicLinkFolder `json:"folders"`
}

// getPublicLink serves GET /api/v1/public/p/:slug. It exposes only fields meant for visitors.
func (s *Server) getPublicLink(c *fiber.Ctx) error {
	slug := strings.ToLower(strings.TrimSpace(c.Params("slug")))
	if slug == "" || len(slug) > 100 {
		return fiber.NewError(http.StatusNotFound, "LINK_NOT_FOUND")
	}
	p, err := s.repo.GetPublicLinkBySlug(c.Context(), slug)
	if errors.Is(err, repository.ErrNotFound) {
		return fiber.NewError(http.StatusNotFound, "LINK_NOT_FOUND")
	}
	if err != nil {
		return err
	}
	out := publicLink{Slug: p.Slug, Name: p.Name, Description: p.Description, MapsURL: p.GoogleMapsURL, Address: p.Address, Socials: p.Socials, LogoURL: publicLogoURL(p)}
	if out.Folders, err = s.repo.ListPublicFolders(c.Context(), slug); err != nil {
		return err
	}
	if user := strings.TrimPrefix(strings.TrimSpace(p.TelegramUsername), "@"); p.TelegramReady && user != "" {
		out.TelegramURL = "https://t.me/" + user
	}
	if digits := nonDigits.ReplaceAllString(p.WhatsAppNumber, ""); p.WhatsAppReady && digits != "" {
		out.WhatsAppURL = "https://wa.me/" + digits
	}
	c.Set("Cache-Control", "public, max-age=60")
	return c.JSON(out)
}

// listPublicLinks serves GET /api/v1/public/p: slugs only, for the sitemap.
func (s *Server) listPublicLinks(c *fiber.Ctx) error {
	slugs, err := s.repo.ListPublicLinkSlugs(c.Context())
	if err != nil {
		return err
	}
	c.Set("Cache-Control", "public, max-age=300")
	return c.JSON(fiber.Map{"data": slugs})
}

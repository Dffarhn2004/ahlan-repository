package httpapi

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2"
	"ponpes-saas/internal/repository"
)

const maxLogoBytes = 2 << 20

// logoTypes maps the sniffed content type (never the client's claim) to the extension stored on disk.
var logoTypes = map[string]string{"image/png": ".png", "image/jpeg": ".jpg", "image/webp": ".webp"}

var kindParam = regexp.MustCompile(`^[A-Z]{3,20}$`)

// uploadLogo serves PUT /tenants/:id/logo. The file type is detected from its bytes; SVG is refused on purpose.
func (s *Server) uploadLogo(c *fiber.Ctx) error {
	id, err := tenantIDParam(c)
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "INVALID_TENANT_ID")
	}
	header, err := c.FormFile("file")
	if err != nil || header.Size > maxLogoBytes {
		return fiber.NewError(http.StatusBadRequest, "LOGO_INVALID")
	}
	src, err := header.Open()
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "LOGO_INVALID")
	}
	defer src.Close()
	head := make([]byte, 512)
	n, _ := io.ReadFull(src, head)
	ext, ok := logoTypes[http.DetectContentType(head[:n])]
	if !ok {
		return fiber.NewError(http.StatusUnsupportedMediaType, "LOGO_TYPE")
	}
	key, _, _, err := s.store.Put(id.String(), "logo"+ext, io.MultiReader(bytes.NewReader(head[:n]), src), maxLogoBytes)
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "LOGO_INVALID")
	}
	old, err := s.repo.SetProfileLogo(c.Context(), id, key)
	if err != nil {
		_ = s.store.Delete(id.String(), key)
		if errors.Is(err, repository.ErrNotFound) {
			return fiber.NewError(http.StatusConflict, "PROFILE_REQUIRED")
		}
		return err
	}
	if old != "" {
		_ = s.store.Delete(id.String(), old)
	}
	s.repo.RecordAudit(c.Context(), id, auditActor(c), "LOGO_UPDATED", "TENANT", id.String(), "")
	return c.JSON(fiber.Map{"has_logo": true})
}

func (s *Server) deleteLogo(c *fiber.Ctx) error {
	id, err := tenantIDParam(c)
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "INVALID_TENANT_ID")
	}
	old, err := s.repo.SetProfileLogo(c.Context(), id, "")
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return err
	}
	if old != "" {
		_ = s.store.Delete(id.String(), old)
		s.repo.RecordAudit(c.Context(), id, auditActor(c), "LOGO_REMOVED", "TENANT", id.String(), "")
	}
	return c.SendStatus(http.StatusNoContent)
}

// getPublicLogo serves GET /api/v1/public/p/:slug/logo. The URL carries the file key as ?v=, so it can be cached for long.
func (s *Server) getPublicLogo(c *fiber.Ctx) error {
	slug := strings.ToLower(strings.TrimSpace(c.Params("slug")))
	tenantID, key, err := s.repo.PublicLogoBySlug(c.Context(), slug)
	if err != nil {
		return fiber.NewError(http.StatusNotFound, "LOGO_NOT_FOUND")
	}
	path, err := s.store.Path(tenantID.String(), key)
	if err != nil {
		return fiber.NewError(http.StatusNotFound, "LOGO_NOT_FOUND")
	}
	c.Set("X-Content-Type-Options", "nosniff")
	c.Set("Cache-Control", "public, max-age=86400")
	return c.SendFile(path)
}

// getPublicFolderLinks serves GET /api/v1/public/p/:slug/links/:kind: the links a pesantren chose to show in one folder.
func (s *Server) getPublicFolderLinks(c *fiber.Ctx) error {
	slug := strings.ToLower(strings.TrimSpace(c.Params("slug")))
	kind := strings.ToUpper(c.Params("kind"))
	if slug == "" || len(slug) > 100 || !kindParam.MatchString(kind) {
		return fiber.NewError(http.StatusNotFound, "LINK_NOT_FOUND")
	}
	items, err := s.repo.ListPublicFolderLinks(c.Context(), slug, kind)
	if err != nil {
		return err
	}
	if len(items) == 0 {
		return fiber.NewError(http.StatusNotFound, "LINK_NOT_FOUND")
	}
	p, err := s.repo.GetPublicLinkBySlug(c.Context(), slug)
	if err != nil {
		return fiber.NewError(http.StatusNotFound, "LINK_NOT_FOUND")
	}
	c.Set("Cache-Control", "public, max-age=60")
	return c.JSON(fiber.Map{"slug": p.Slug, "name": p.Name, "logo_url": publicLogoURL(p), "kind": kind, "items": items})
}

func publicLogoURL(p repository.PublicLink) string {
	if p.LogoKey == "" {
		return ""
	}
	return "/api/v1/public/p/" + p.Slug + "/logo?v=" + strings.TrimSuffix(p.LogoKey, filepath.Ext(p.LogoKey))
}

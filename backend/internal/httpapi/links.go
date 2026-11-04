package httpapi

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"ponpes-saas/internal/crawler"
	"ponpes-saas/internal/links"
	"ponpes-saas/internal/repository"
)

// learnIntervalMinutes is how often a learned link is re-read: daily, the dashboard's default.
const learnIntervalMinutes = 1440

type linkInput struct {
	Kind       *string `json:"kind"`
	Label      *string `json:"label"`
	URL        *string `json:"url"`
	ShareHint  *string `json:"share_hint"`
	Share      *bool   `json:"share"`
	ShowPublic *bool   `json:"show_public"`
	Learn      *bool   `json:"learn"`
}

func (s *Server) listLinks(c *fiber.Ctx) error {
	id, err := tenantIDParam(c)
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "INVALID_TENANT_ID")
	}
	items, err := s.repo.ListTenantLinks(c.Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"data": items})
}

func linkError(err error) error {
	switch {
	case errors.Is(err, links.ErrInvalidKind), errors.Is(err, links.ErrInvalidURL), errors.Is(err, links.ErrWrongDomain):
		return fiber.NewError(http.StatusUnprocessableEntity, err.Error())
	}
	return err
}

// learnSource registers the link's address as a crawl source, or adopts the tenant's existing source for it.
func (s *Server) learnSource(ctx context.Context, tenantID, actor uuid.UUID, link string) (uuid.UUID, error) {
	target, err := crawler.ParseSource(link)
	if err != nil {
		return uuid.Nil, fiber.NewError(http.StatusUnprocessableEntity, "CRAWL_URL_FORBIDDEN")
	}
	if existing, err := s.repo.FindCrawlSourceByURL(ctx, tenantID, target.Start.String()); err == nil {
		return existing.ID, nil
	}
	if !target.IsGoogle() {
		checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		if err := crawler.CheckHost(checkCtx, target.Start.Hostname()); err != nil {
			code := "CRAWL_URL_FORBIDDEN"
			if errors.Is(err, crawler.ErrUnreachable) {
				code = "CRAWL_URL_UNREACHABLE"
			}
			return uuid.Nil, fiber.NewError(http.StatusUnprocessableEntity, code)
		}
	}
	interval := learnIntervalMinutes
	created, err := s.repo.CreateCrawlSource(ctx, repository.CrawlSource{TenantID: tenantID, URL: target.Start.String(), Kind: string(target.Kind), MaxDepth: crawler.DefaultMaxDepth, MaxPages: crawler.DefaultMaxPages,
		ScheduleMode: crawler.ModeInterval, IntervalMinutes: &interval}, actor)
	switch {
	case errors.Is(err, repository.ErrCrawlLimit):
		return uuid.Nil, fiber.NewError(http.StatusConflict, "CRAWL_SOURCE_LIMIT")
	case err != nil:
		return uuid.Nil, err
	}
	s.wakeCrawler()
	return created.ID, nil
}

// refreshLinksInPersona rewrites the assistant's instructions so a changed link list takes effect.
func (s *Server) refreshLinksInPersona(ctx context.Context, tenantID uuid.UUID) {
	if key, err := s.repo.HermesProfileKey(ctx, tenantID); err == nil {
		s.refreshPersona(ctx, key)
	}
}

func (s *Server) createLink(c *fiber.Ctx) error {
	tenantID, err := tenantIDParam(c)
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "INVALID_TENANT_ID")
	}
	var in linkInput
	if err := c.BodyParser(&in); err != nil || in.Kind == nil || in.URL == nil || in.Label == nil {
		return fiber.NewError(http.StatusBadRequest, "INVALID_REQUEST")
	}
	link := repository.TenantLink{TenantID: tenantID, Kind: *in.Kind, Label: links.Clean(*in.Label, 80), Share: in.Share == nil || *in.Share, ShowPublic: in.ShowPublic != nil && *in.ShowPublic}
	if in.ShareHint != nil {
		link.ShareHint = links.Clean(*in.ShareHint, 160)
	}
	if link.Label == "" {
		return fiber.NewError(http.StatusBadRequest, "INVALID_REQUEST")
	}
	if link.URL, err = links.Validate(link.Kind, *in.URL); err != nil {
		return linkError(err)
	}
	learn := in.Learn != nil && *in.Learn
	if !link.Share && !learn && !link.ShowPublic {
		return fiber.NewError(http.StatusBadRequest, "LINK_NEEDS_PURPOSE")
	}
	actor := c.Locals("user_id").(uuid.UUID)
	if learn {
		if !links.CanLearn(link.Kind, link.URL) {
			return fiber.NewError(http.StatusUnprocessableEntity, links.ErrCannotLearn.Error())
		}
		sourceID, err := s.learnSource(c.Context(), tenantID, actor, link.URL)
		if err != nil {
			return err
		}
		link.CrawlSourceID = &sourceID
	}
	created, err := s.repo.CreateTenantLink(c.Context(), link)
	switch {
	case errors.Is(err, repository.ErrLinkLimit):
		return fiber.NewError(http.StatusConflict, "LINK_LIMIT")
	case errors.Is(err, repository.ErrConflict):
		return fiber.NewError(http.StatusConflict, "LINK_EXISTS")
	case err != nil:
		return err
	}
	s.repo.RecordAudit(c.Context(), tenantID, &actor, "LINK_CREATED", "TENANT_LINK", created.ID.String(), created.Label)
	s.refreshLinksInPersona(c.Context(), tenantID)
	return c.Status(http.StatusCreated).JSON(created)
}

func (s *Server) updateLink(c *fiber.Ctx) error {
	tenantID, linkID, err := tenantAndParam(c, "lid", "LINK_NOT_FOUND")
	if err != nil {
		return err
	}
	var in linkInput
	if err := c.BodyParser(&in); err != nil {
		return fiber.NewError(http.StatusBadRequest, "INVALID_REQUEST")
	}
	current, err := s.repo.GetTenantLink(c.Context(), tenantID, linkID)
	if err != nil {
		return fiber.NewError(http.StatusNotFound, "LINK_NOT_FOUND")
	}
	next := current
	if in.Kind != nil {
		next.Kind = *in.Kind
	}
	if in.Label != nil {
		next.Label = links.Clean(*in.Label, 80)
	}
	if in.ShareHint != nil {
		next.ShareHint = links.Clean(*in.ShareHint, 160)
	}
	if in.Share != nil {
		next.Share = *in.Share
	}
	if in.ShowPublic != nil {
		next.ShowPublic = *in.ShowPublic
	}
	rawURL := current.URL
	if in.URL != nil {
		rawURL = *in.URL
	}
	if next.Label == "" {
		return fiber.NewError(http.StatusBadRequest, "INVALID_REQUEST")
	}
	if next.URL, err = links.Validate(next.Kind, rawURL); err != nil {
		return linkError(err)
	}
	wantLearn := current.Learn
	if in.Learn != nil {
		wantLearn = *in.Learn
	}
	if !next.Share && !wantLearn && !next.ShowPublic {
		return fiber.NewError(http.StatusBadRequest, "LINK_NEEDS_PURPOSE")
	}
	if wantLearn && !links.CanLearn(next.Kind, next.URL) {
		return fiber.NewError(http.StatusUnprocessableEntity, links.ErrCannotLearn.Error())
	}

	actor := c.Locals("user_id").(uuid.UUID)
	changedAddress := next.URL != current.URL
	// Stop learning, or learn from a new address: the old source (and what it produced) goes away.
	if current.Learn && (!wantLearn || changedAddress) {
		if err := s.removeCrawlSource(c.Context(), tenantID, *current.CrawlSourceID); err != nil && !errors.Is(err, repository.ErrNotFound) {
			return err
		}
		next.CrawlSourceID = nil
	}
	if wantLearn && next.CrawlSourceID == nil {
		sourceID, err := s.learnSource(c.Context(), tenantID, actor, next.URL)
		if err != nil {
			return err
		}
		next.CrawlSourceID = &sourceID
	}
	updated, err := s.repo.UpdateTenantLink(c.Context(), next)
	switch {
	case errors.Is(err, repository.ErrConflict):
		return fiber.NewError(http.StatusConflict, "LINK_EXISTS")
	case errors.Is(err, repository.ErrNotFound):
		return fiber.NewError(http.StatusNotFound, "LINK_NOT_FOUND")
	case err != nil:
		return err
	}
	s.repo.RecordAudit(c.Context(), tenantID, &actor, "LINK_UPDATED", "TENANT_LINK", linkID.String(), updated.Label)
	s.refreshLinksInPersona(c.Context(), tenantID)
	return c.JSON(updated)
}

func (s *Server) deleteLink(c *fiber.Ctx) error {
	tenantID, linkID, err := tenantAndParam(c, "lid", "LINK_NOT_FOUND")
	if err != nil {
		return err
	}
	current, err := s.repo.GetTenantLink(c.Context(), tenantID, linkID)
	if err != nil {
		return fiber.NewError(http.StatusNotFound, "LINK_NOT_FOUND")
	}
	if current.Learn {
		if err := s.removeCrawlSource(c.Context(), tenantID, *current.CrawlSourceID); err != nil && !errors.Is(err, repository.ErrNotFound) {
			return err
		}
	}
	if err := s.repo.DeleteTenantLink(c.Context(), tenantID, linkID); err != nil {
		return err
	}
	s.repo.RecordAudit(c.Context(), tenantID, auditActor(c), "LINK_DELETED", "TENANT_LINK", linkID.String(), current.Label)
	s.refreshLinksInPersona(c.Context(), tenantID)
	return c.SendStatus(http.StatusNoContent)
}

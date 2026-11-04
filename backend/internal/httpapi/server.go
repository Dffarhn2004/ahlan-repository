// Package httpapi holds the HTTP layer: one Server, with its handlers split per feature (auth.go, tenants.go, ...)
// and every route listed in routes.go.
package httpapi

import (
	"log/slog"
	"net/http"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"ponpes-saas/internal/authz"
	"ponpes-saas/internal/config"
	"ponpes-saas/internal/crawler"
	"ponpes-saas/internal/hermes"
	"ponpes-saas/internal/repository"
	"ponpes-saas/internal/storage"
)

// Server carries the dependencies shared by every handler.
type Server struct {
	cfg    config.Config
	repo   repository.Repository
	store  storage.Store
	hermes *hermes.HTTPClient
	// control manages profiles and their files inside the Hermes container (Docker CLI or bridge).
	control   hermes.Control
	knowledge knowledgeJobs
	crawler   *crawler.Crawler
	// crawlWake nudges the crawl worker when a run is queued, so a click does not wait for the next tick.
	crawlWake chan struct{}
	// whatsappPairings holds the device-linking attempt per tenant (key: tenant ID).
	whatsappPairings sync.Map
	// syncJobs tracks the background "sync all assistants" runs.
	syncJobs syncJobStore
}

func NewServer(cfg config.Config, repo repository.Repository, store storage.Store) *Server {
	control, err := hermes.NewControl(hermes.ControlConfig{Mode: cfg.HermesExecMode, Container: cfg.HermesContainer, BridgeSocket: cfg.HermesBridgeSocket})
	if err != nil {
		slog.Error("Hermes control is misconfigured; assistant management is disabled", "error", err)
		control = hermes.Unavailable(err)
	}
	return &Server{cfg: cfg, repo: repo, store: store, hermes: hermes.NewHTTPClient(cfg.HermesBaseURL, cfg.HermesAPIKey), control: control,
		crawler: crawler.New(crawler.NewFetcher(crawler.FetcherConfig{UserAgent: "PonpesAICrawler/1.0"})), crawlWake: make(chan struct{}, 1)}
}

// docker returns the control that manages assistant profiles inside the Hermes container. The name is historical:
// it is a hermes.Control, which may be the Docker CLI (development) or the bridge (production).
func (s *Server) docker() hermes.Control { return s.control }

// eventsURL is where the Hermes container posts conversation events.
func (s *Server) eventsURL() string {
	if s.cfg.HermesEventsURL != "" {
		return s.cfg.HermesEventsURL
	}
	return config.DefaultHermesEventsURL
}

func tenantIDParam(c *fiber.Ctx) (uuid.UUID, error) {
	return uuid.Parse(c.Params("id"))
}

func auditActor(c *fiber.Ctx) *uuid.UUID {
	if id, ok := c.Locals("user_id").(uuid.UUID); ok {
		return &id
	}
	return nil
}

// sessionUser loads the signed-in user, answering 401 when there is none.
func (s *Server) sessionUser(c *fiber.Ctx) (repository.User, uuid.UUID, error) {
	id, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return repository.User{}, uuid.Nil, fiber.NewError(http.StatusUnauthorized, "UNAUTHENTICATED")
	}
	user, err := s.repo.FindUserByID(c.Context(), id)
	if err != nil {
		return repository.User{}, uuid.Nil, fiber.NewError(http.StatusUnauthorized, "UNAUTHENTICATED")
	}
	return user, id, nil
}

// authenticate resolves the session cookie into the "user_id" local.
func (s *Server) authenticate(c *fiber.Ctx) error {
	id, err := s.repo.SessionUser(c.Context(), c.Cookies("session"))
	if err != nil {
		return fiber.NewError(http.StatusUnauthorized, "UNAUTHENTICATED")
	}
	c.Locals("user_id", id)
	return c.Next()
}

// checkTenant lets superadmins through and requires everyone else to belong to the tenant in :id.
// A non-empty permission must also be granted to their role there.
func (s *Server) checkTenant(c *fiber.Ctx, permission string) error {
	user, userID, err := s.sessionUser(c)
	if err != nil {
		return err
	}
	tenantID, err := tenantIDParam(c)
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "INVALID_TENANT_ID")
	}
	if user.Superadmin {
		return nil
	}
	role, err := s.repo.TenantRole(c.Context(), userID, tenantID)
	if err != nil || (permission != "" && !authz.Allows(role, permission)) {
		return fiber.NewError(http.StatusForbidden, "FORBIDDEN")
	}
	return nil
}

// tenantAccess guards a /tenants/:id route. An empty permission only requires membership.
func (s *Server) tenantAccess(permission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if err := s.checkTenant(c, permission); err != nil {
			return err
		}
		return c.Next()
	}
}

func (s *Server) requireSuperadmin(c *fiber.Ctx) error {
	user, _, err := s.sessionUser(c)
	if err != nil {
		return err
	}
	if !user.Superadmin {
		return fiber.NewError(http.StatusForbidden, "FORBIDDEN")
	}
	return c.Next()
}

func notImplemented(*fiber.Ctx) error {
	return fiber.NewError(http.StatusNotImplemented, "DOMAIN_ROUTE_NOT_IMPLEMENTED")
}

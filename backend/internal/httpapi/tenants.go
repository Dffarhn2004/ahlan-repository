package httpapi

import (
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"ponpes-saas/internal/authz"
	"ponpes-saas/internal/repository"
)

func (s *Server) listTenants(c *fiber.Ctx) error {
	user, userID, err := s.sessionUser(c)
	if err != nil {
		return err
	}
	var tenants []repository.Tenant
	if user.Superadmin {
		tenants, err = s.repo.ListTenants(c.Context())
	} else {
		tenants, err = s.repo.ListMemberTenants(c.Context(), userID)
	}
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"data": tenants})
}

func (s *Server) createTenant(c *fiber.Ctx) error {
	var in struct{ Name, Slug string }
	if err := c.BodyParser(&in); err != nil {
		return fiber.NewError(http.StatusBadRequest, "INVALID_REQUEST")
	}
	actor, _ := c.Locals("user_id").(uuid.UUID)
	tenant, err := s.repo.CreateTenant(c.Context(), in.Name, in.Slug, actor)
	if err != nil {
		return err
	}
	return c.Status(http.StatusAccepted).JSON(tenant)
}

func (s *Server) getTenant(c *fiber.Ctx) error {
	id, err := tenantIDParam(c)
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "INVALID_TENANT_ID")
	}
	tenant, err := s.repo.GetTenant(c.Context(), id)
	if err != nil {
		return fiber.NewError(http.StatusNotFound, "TENANT_NOT_FOUND")
	}
	return c.JSON(tenant)
}

func (s *Server) updateTenant(c *fiber.Ctx) error {
	var in struct{ Name, Slug string }
	if err := c.BodyParser(&in); err != nil || strings.TrimSpace(in.Name) == "" || strings.TrimSpace(in.Slug) == "" {
		return fiber.NewError(http.StatusBadRequest, "INVALID_REQUEST")
	}
	id, err := tenantIDParam(c)
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "INVALID_TENANT_ID")
	}
	actor := c.Locals("user_id").(uuid.UUID)
	if err = s.repo.UpdateTenant(c.Context(), id, strings.TrimSpace(in.Name), strings.TrimSpace(in.Slug), actor); err != nil {
		return err
	}
	return c.JSON(fiber.Map{"status": "UPDATED"})
}

// setTenantStatus suspends a tenant or puts it back to onboarding.
func (s *Server) setTenantStatus(status string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := tenantIDParam(c)
		if err != nil {
			return fiber.NewError(http.StatusBadRequest, "INVALID_TENANT_ID")
		}
		if err := s.repo.SetTenantStatus(c.Context(), id, status, c.Locals("user_id").(uuid.UUID)); err != nil {
			return err
		}
		return c.SendStatus(http.StatusNoContent)
	}
}

func (s *Server) auditLogs(c *fiber.Ctx) error {
	var scope *uuid.UUID
	if value := c.Query("tenant_id"); value != "" {
		parsed, err := uuid.Parse(value)
		if err != nil {
			return fiber.NewError(http.StatusBadRequest, "INVALID_TENANT_ID")
		}
		scope = &parsed
	}
	user, userID, err := s.sessionUser(c)
	if err != nil {
		return err
	}
	if !user.Superadmin {
		// Tenant users only ever see their own pesantren's log.
		if scope == nil {
			return fiber.NewError(http.StatusBadRequest, "TENANT_REQUIRED")
		}
		role, roleErr := s.repo.TenantRole(c.Context(), userID, *scope)
		if roleErr != nil || !authz.Allows(role, "audit:read") {
			return fiber.NewError(http.StatusForbidden, "FORBIDDEN")
		}
	}
	logs, err := s.repo.AuditLogs(c.Context(), scope)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"data": logs})
}

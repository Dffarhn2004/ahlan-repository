package httpapi

import (
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"ponpes-saas/internal/authz"
)

func validMemberRole(role string) bool {
	return role == authz.Admin || role == authz.Staff || role == authz.Viewer
}

func (s *Server) listMembers(c *fiber.Ctx) error {
	id, err := tenantIDParam(c)
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "INVALID_TENANT_ID")
	}
	members, err := s.repo.TenantMembers(c.Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"data": members})
}

func (s *Server) createMember(c *fiber.Ctx) error {
	id, err := tenantIDParam(c)
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "INVALID_TENANT_ID")
	}
	var in struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	if err := c.BodyParser(&in); err != nil || !strings.Contains(in.Email, "@") || len(in.Password) < 10 {
		return fiber.NewError(http.StatusBadRequest, "INVALID_MEMBER")
	}
	if in.Role == "" {
		in.Role = authz.Admin
	}
	if !validMemberRole(in.Role) {
		return fiber.NewError(http.StatusBadRequest, "INVALID_MEMBER_ROLE")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	memberID, err := s.repo.CreateMember(c.Context(), id, strings.ToLower(strings.TrimSpace(in.Email)), hash, in.Role)
	if err != nil {
		return fiber.NewError(http.StatusConflict, "MEMBER_CREATE_FAILED")
	}
	return c.Status(http.StatusCreated).JSON(fiber.Map{"user_id": memberID, "role": in.Role})
}

func (s *Server) updateMemberRole(c *fiber.Ctx) error {
	id, err := tenantIDParam(c)
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "INVALID_TENANT_ID")
	}
	userID, err := uuid.Parse(c.Params("user_id"))
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "INVALID_USER_ID")
	}
	var in struct {
		Role string `json:"role"`
	}
	if err = c.BodyParser(&in); err != nil || !validMemberRole(in.Role) {
		return fiber.NewError(http.StatusBadRequest, "INVALID_MEMBER_ROLE")
	}
	if err = s.repo.UpdateMemberRole(c.Context(), id, userID, in.Role, c.Locals("user_id").(uuid.UUID)); err != nil {
		return fiber.NewError(http.StatusNotFound, "MEMBER_NOT_FOUND")
	}
	return c.JSON(fiber.Map{"status": "UPDATED", "role": in.Role})
}

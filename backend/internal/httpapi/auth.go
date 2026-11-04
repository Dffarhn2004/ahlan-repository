package httpapi

import (
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"ponpes-saas/internal/authz"
)

func (s *Server) health(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok"})
}

func (s *Server) login(c *fiber.Ctx) error {
	var in struct{ Email, Password string }
	if err := c.BodyParser(&in); err != nil {
		return fiber.NewError(http.StatusBadRequest, "INVALID_REQUEST")
	}
	user, err := s.repo.FindUser(c.Context(), in.Email)
	if err != nil || bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(in.Password)) != nil {
		return fiber.NewError(http.StatusUnauthorized, "INVALID_CREDENTIALS")
	}
	raw := uuid.NewString() + uuid.NewString()
	if err = s.repo.CreateSession(c.Context(), user.ID, raw, time.Now().Add(24*time.Hour)); err != nil {
		return err
	}
	c.Cookie(&fiber.Cookie{Name: "session", Value: raw, HTTPOnly: true, SameSite: "Lax", MaxAge: 86400, Path: "/"})
	role := "VIEWER"
	if user.Superadmin {
		role = "SUPERADMIN"
	}
	return c.JSON(fiber.Map{"user": fiber.Map{"id": user.ID, "email": user.Email, "superadmin": user.Superadmin}, "role": role})
}

func (s *Server) logout(c *fiber.Ctx) error {
	_ = s.repo.DeleteSession(c.Context(), c.Cookies("session"))
	c.ClearCookie("session")
	return c.SendStatus(http.StatusNoContent)
}

func (s *Server) me(c *fiber.Ctx) error {
	id, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return fiber.NewError(http.StatusUnauthorized, "UNAUTHENTICATED")
	}
	user, err := s.repo.FindUserByID(c.Context(), id)
	if err != nil {
		return fiber.NewError(http.StatusNotFound, "USER_NOT_FOUND")
	}
	role := authz.Viewer
	if user.Superadmin {
		role = authz.Superadmin
	}
	memberships, err := s.repo.Memberships(c.Context(), id)
	if err != nil {
		return err
	}
	if !user.Superadmin && len(memberships) > 0 {
		role = memberships[0].Role
	}
	return c.JSON(fiber.Map{"user": fiber.Map{"id": user.ID, "email": user.Email}, "role": role, "permissions": authz.Permissions(role), "memberships": memberships})
}

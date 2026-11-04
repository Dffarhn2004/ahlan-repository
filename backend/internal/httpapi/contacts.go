package httpapi

import (
	"errors"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"ponpes-saas/internal/contacts"
	"ponpes-saas/internal/links"
	"ponpes-saas/internal/repository"
)

type contactInput struct {
	Name      *string `json:"name"`
	Role      *string `json:"role"`
	Phone     *string `json:"phone"`
	Email     *string `json:"email"`
	Hours     *string `json:"hours"`
	ShareHint *string `json:"share_hint"`
}

func (s *Server) listContacts(c *fiber.Ctx) error {
	id, err := tenantIDParam(c)
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "INVALID_TENANT_ID")
	}
	items, err := s.repo.ListTenantContacts(c.Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"data": items})
}

func contactError(err error) error {
	if errors.Is(err, contacts.ErrInvalidPhone) || errors.Is(err, contacts.ErrInvalidEmail) {
		return fiber.NewError(http.StatusUnprocessableEntity, err.Error())
	}
	return err
}

// apply copies the supplied fields onto the contact, cleaned and validated. Fields left out are kept.
func (in contactInput) apply(c *repository.TenantContact) error {
	if in.Name != nil {
		c.Name = links.Clean(*in.Name, 80)
	}
	if in.Role != nil {
		c.Role = links.Clean(*in.Role, 80)
	}
	if in.Hours != nil {
		c.Hours = links.Clean(*in.Hours, 80)
	}
	if in.ShareHint != nil {
		c.ShareHint = links.Clean(*in.ShareHint, 160)
	}
	if in.Phone != nil {
		phone, err := contacts.NormalizePhone(*in.Phone)
		if err != nil {
			return err
		}
		c.Phone = phone
	}
	if in.Email != nil {
		email, err := contacts.NormalizeEmail(*in.Email)
		if err != nil {
			return err
		}
		c.Email = email
	}
	if c.Name == "" || c.Phone == "" {
		return fiber.NewError(http.StatusBadRequest, "INVALID_REQUEST")
	}
	return nil
}

func (s *Server) createContact(c *fiber.Ctx) error {
	tenantID, err := tenantIDParam(c)
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "INVALID_TENANT_ID")
	}
	var in contactInput
	if err := c.BodyParser(&in); err != nil || in.Name == nil || in.Phone == nil {
		return fiber.NewError(http.StatusBadRequest, "INVALID_REQUEST")
	}
	contact := repository.TenantContact{TenantID: tenantID}
	if err := in.apply(&contact); err != nil {
		return contactError(err)
	}
	created, err := s.repo.CreateTenantContact(c.Context(), contact)
	switch {
	case errors.Is(err, repository.ErrContactLimit):
		return fiber.NewError(http.StatusConflict, "CONTACT_LIMIT")
	case err != nil:
		return err
	}
	s.repo.RecordAudit(c.Context(), tenantID, auditActor(c), "CONTACT_CREATED", "TENANT_CONTACT", created.ID.String(), created.Name)
	s.refreshLinksInPersona(c.Context(), tenantID)
	return c.Status(http.StatusCreated).JSON(created)
}

func (s *Server) updateContact(c *fiber.Ctx) error {
	tenantID, contactID, err := tenantAndParam(c, "cid", "CONTACT_NOT_FOUND")
	if err != nil {
		return err
	}
	var in contactInput
	if err := c.BodyParser(&in); err != nil {
		return fiber.NewError(http.StatusBadRequest, "INVALID_REQUEST")
	}
	current, err := s.repo.GetTenantContact(c.Context(), tenantID, contactID)
	if err != nil {
		return fiber.NewError(http.StatusNotFound, "CONTACT_NOT_FOUND")
	}
	if err := in.apply(&current); err != nil {
		return contactError(err)
	}
	updated, err := s.repo.UpdateTenantContact(c.Context(), current)
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return fiber.NewError(http.StatusNotFound, "CONTACT_NOT_FOUND")
	case err != nil:
		return err
	}
	s.repo.RecordAudit(c.Context(), tenantID, auditActor(c), "CONTACT_UPDATED", "TENANT_CONTACT", contactID.String(), updated.Name)
	s.refreshLinksInPersona(c.Context(), tenantID)
	return c.JSON(updated)
}

func (s *Server) deleteContact(c *fiber.Ctx) error {
	tenantID, contactID, err := tenantAndParam(c, "cid", "CONTACT_NOT_FOUND")
	if err != nil {
		return err
	}
	current, err := s.repo.GetTenantContact(c.Context(), tenantID, contactID)
	if err != nil {
		return fiber.NewError(http.StatusNotFound, "CONTACT_NOT_FOUND")
	}
	if err := s.repo.DeleteTenantContact(c.Context(), tenantID, contactID); err != nil {
		return err
	}
	s.repo.RecordAudit(c.Context(), tenantID, auditActor(c), "CONTACT_DELETED", "TENANT_CONTACT", contactID.String(), current.Name)
	s.refreshLinksInPersona(c.Context(), tenantID)
	return c.SendStatus(http.StatusNoContent)
}

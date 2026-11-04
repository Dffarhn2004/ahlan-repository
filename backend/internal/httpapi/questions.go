package httpapi

import (
	"errors"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"ponpes-saas/internal/links"
	"ponpes-saas/internal/questions"
	"ponpes-saas/internal/repository"
)

type questionInput struct {
	Status      *string `json:"status"`
	FAQQuestion *string `json:"faq_question"`
	Answer      *string `json:"answer"`
}

func validQuestionStatus(status string) bool {
	switch status {
	case questions.StatusOpen, questions.StatusAnswered, questions.StatusResolved, questions.StatusIgnored:
		return true
	}
	return false
}

func (s *Server) listQuestions(c *fiber.Ctx) error {
	id, err := tenantIDParam(c)
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "INVALID_TENANT_ID")
	}
	status := c.Query("status", questions.StatusOpen)
	if !validQuestionStatus(status) {
		return fiber.NewError(http.StatusBadRequest, "INVALID_REQUEST")
	}
	items, err := s.repo.ListTenantQuestions(c.Context(), id, status)
	if err != nil {
		return err
	}
	counts, err := s.repo.CountTenantQuestions(c.Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"data": items, "counts": counts})
}

// updateQuestion answers a question (an answer makes it ANSWERED and hands it to the assistant) or sets its status:
// RESOLVED (fixed another way, for example by uploading a document), IGNORED, or OPEN again.
func (s *Server) updateQuestion(c *fiber.Ctx) error {
	tenantID, questionID, err := tenantAndParam(c, "qid", "QUESTION_NOT_FOUND")
	if err != nil {
		return err
	}
	var in questionInput
	if err := c.BodyParser(&in); err != nil {
		return fiber.NewError(http.StatusBadRequest, "INVALID_REQUEST")
	}
	current, err := s.repo.GetTenantQuestion(c.Context(), tenantID, questionID)
	if err != nil {
		return fiber.NewError(http.StatusNotFound, "QUESTION_NOT_FOUND")
	}
	wasAnswered := current.Status == questions.StatusAnswered
	next := current
	if in.Answer != nil && links.Clean(*in.Answer, 1500) != "" {
		next.Status = questions.StatusAnswered
		next.Answer = links.Clean(*in.Answer, 1500)
		next.FAQQuestion = current.FAQQuestion
		if in.FAQQuestion != nil {
			next.FAQQuestion = links.Clean(*in.FAQQuestion, 200)
		}
		if next.FAQQuestion == "" {
			next.FAQQuestion = links.Clean(current.Question, 200)
		}
	} else {
		if in.Status == nil || !validQuestionStatus(*in.Status) || *in.Status == questions.StatusAnswered {
			return fiber.NewError(http.StatusBadRequest, "INVALID_REQUEST")
		}
		next.Status = *in.Status
	}
	updated, err := s.repo.UpdateTenantQuestion(c.Context(), next)
	switch {
	case errors.Is(err, repository.ErrQuestionLimit):
		return fiber.NewError(http.StatusConflict, "QUESTION_LIMIT")
	case errors.Is(err, repository.ErrNotFound):
		return fiber.NewError(http.StatusNotFound, "QUESTION_NOT_FOUND")
	case err != nil:
		return err
	}
	action := "QUESTION_" + updated.Status
	s.repo.RecordAudit(c.Context(), tenantID, auditActor(c), action, "TENANT_QUESTION", questionID.String(), links.Clean(updated.Question, 80))
	if wasAnswered || updated.Status == questions.StatusAnswered {
		s.refreshLinksInPersona(c.Context(), tenantID)
	}
	return c.JSON(updated)
}

func (s *Server) deleteQuestion(c *fiber.Ctx) error {
	tenantID, questionID, err := tenantAndParam(c, "qid", "QUESTION_NOT_FOUND")
	if err != nil {
		return err
	}
	current, err := s.repo.GetTenantQuestion(c.Context(), tenantID, questionID)
	if err != nil {
		return fiber.NewError(http.StatusNotFound, "QUESTION_NOT_FOUND")
	}
	if err := s.repo.DeleteTenantQuestion(c.Context(), tenantID, questionID); err != nil {
		return err
	}
	s.repo.RecordAudit(c.Context(), tenantID, auditActor(c), "QUESTION_DELETED", "TENANT_QUESTION", questionID.String(), links.Clean(current.Question, 80))
	if current.Status == questions.StatusAnswered {
		s.refreshLinksInPersona(c.Context(), tenantID)
	}
	return c.SendStatus(http.StatusNoContent)
}

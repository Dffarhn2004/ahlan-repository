package httpapi

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"ponpes-saas/internal/hermes"
)

var assistantActions = map[string]bool{"start": true, "stop": true, "restart": true, "update": true, "delete": true}

func (s *Server) provisionAssistant(c *fiber.Ctx) error {
	id, err := tenantIDParam(c)
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "INVALID_TENANT_ID")
	}
	if err = s.repo.QueueHermesProvisioning(c.Context(), id); err != nil {
		return err
	}
	s.repo.RecordAudit(c.Context(), id, auditActor(c), "HERMES_PROVISIONING_QUEUED", "TENANT", id.String(), "")
	return c.Status(http.StatusAccepted).JSON(fiber.Map{"status": "QUEUED"})
}

// assistantAction starts, stops, restarts, updates or deletes a tenant's assistant.
func (s *Server) assistantAction(c *fiber.Ctx) error {
	id, err := tenantIDParam(c)
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "INVALID_TENANT_ID")
	}
	var input struct {
		Action string `json:"action"`
	}
	if err := c.BodyParser(&input); err != nil || !assistantActions[input.Action] {
		return fiber.NewError(http.StatusBadRequest, "INVALID_HERMES_ACTION")
	}
	if input.Action == "delete" {
		// Removing the assistant is a platform-level action.
		if err := s.checkTenant(c, "tenant:lifecycle"); err != nil {
			return err
		}
	}
	profileKey, err := s.repo.HermesProfileKey(c.Context(), id)
	if err != nil {
		return fiber.NewError(http.StatusConflict, "HERMES_PROFILE_REQUIRED")
	}
	client := s.docker()
	// A restart re-applies the lockdown too, so an assistant started before it existed cannot stay exposed.
	if input.Action == "start" || input.Action == "restart" {
		if err := s.configureChannels(c.Context(), id, profileKey); err != nil {
			slog.Error("could not configure assistant channels", "profile", profileKey, "error", err)
			return fiber.NewError(http.StatusBadGateway, "CHANNEL_CONFIG_FAILED")
		}
		// Fail closed: never start a visitor-facing assistant that kept its powerful tools.
		if err := s.lockDown(c.Context(), profileKey); err != nil {
			slog.Error("could not lock down assistant", "profile", profileKey, "error", err)
			return fiber.NewError(http.StatusBadGateway, "ASSISTANT_LOCKDOWN_FAILED")
		}
		s.refreshPersona(c.Context(), profileKey)
		if strings.TrimSpace(s.cfg.HermesAPIKey) == "" {
			return fiber.NewError(http.StatusFailedDependency, "HERMES_EVENT_KEY_NOT_CONFIGURED")
		}
		if err := client.InstallTelegramLogHook(c.Context(), profileKey, s.eventsURL(), s.cfg.HermesAPIKey); err != nil {
			return fiber.NewError(http.StatusBadGateway, "TELEGRAM_LOG_HOOK_FAILED")
		}
	}
	if err := client.ProfileAction(c.Context(), profileKey, input.Action); err != nil {
		return fiber.NewError(http.StatusBadGateway, "HERMES_ACTION_FAILED")
	}
	switch input.Action {
	case "start":
		_ = s.repo.SetTelegramBotStatus(c.Context(), id, "ACTIVE")
		_ = s.repo.SetWhatsAppChannelStatus(c.Context(), id, "ACTIVE")
	case "stop":
		_ = s.repo.SetTelegramBotStatus(c.Context(), id, "SUSPENDED")
		_ = s.repo.SetWhatsAppChannelStatus(c.Context(), id, "SUSPENDED")
	case "delete":
		if err := s.repo.DeleteHermesProfile(c.Context(), id); err != nil {
			return err
		}
	}
	s.repo.RecordAudit(c.Context(), id, auditActor(c), "HERMES_"+strings.ToUpper(input.Action), "TENANT", id.String(), "")
	switch input.Action {
	case "start":
		s.repo.RecordAudit(c.Context(), id, auditActor(c), "TELEGRAM_ACTIVATED", "TENANT", id.String(), "")
		if _, waErr := s.repo.GetWhatsAppChannel(c.Context(), id); waErr == nil {
			s.repo.RecordAudit(c.Context(), id, auditActor(c), "WHATSAPP_ACTIVATED", "TENANT", id.String(), "")
		}
	case "stop":
		s.repo.RecordAudit(c.Context(), id, auditActor(c), "TELEGRAM_DEACTIVATED", "TENANT", id.String(), "")
		if _, waErr := s.repo.GetWhatsAppChannel(c.Context(), id); waErr == nil {
			s.repo.RecordAudit(c.Context(), id, auditActor(c), "WHATSAPP_DEACTIVATED", "TENANT", id.String(), "")
		}
	}
	return c.JSON(fiber.Map{"status": "OK", "action": input.Action})
}

// syncKnowledge copies the tenant's documents to the assistant and starts reading them.
func (s *Server) syncKnowledge(c *fiber.Ctx) error {
	id, err := tenantIDParam(c)
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "INVALID_TENANT_ID")
	}
	profileKey, err := s.repo.HermesProfileKey(c.Context(), id)
	if err != nil {
		return fiber.NewError(http.StatusConflict, "HERMES_PROFILE_REQUIRED")
	}
	count, images, err := s.syncTenantKnowledge(c.Context(), id, profileKey)
	if errors.Is(err, errHermesSync) {
		return fiber.NewError(http.StatusBadGateway, "HERMES_SYNC_FAILED")
	}
	if err != nil {
		return err
	}
	s.repo.RecordAudit(c.Context(), id, auditActor(c), "KNOWLEDGE_SYNCED", "TENANT", id.String(), "")
	return c.JSON(fiber.Map{"status": "SYNCED", "files": count, "profile_key": profileKey, "images": images})
}

// errHermesSync marks a failure to copy documents into the assistant, as opposed to a local storage or database fault.
var errHermesSync = errors.New("hermes knowledge sync failed")

// syncTenantKnowledge copies every readable document of a tenant to its assistant and starts reading them. It is shared
// by the dashboard's "sync" button and by the crawler, which has no HTTP request to hang on.
func (s *Server) syncTenantKnowledge(ctx context.Context, id uuid.UUID, profileKey string) (count, images int, err error) {
	files, err := s.repo.ListFiles(ctx, id)
	if err != nil {
		return 0, 0, err
	}
	payload := make([]hermes.KnowledgeFile, 0, len(files))
	opened := make([]*os.File, 0, len(files))
	defer func() {
		for _, file := range opened {
			_ = file.Close()
		}
	}()
	for _, file := range files {
		if strings.ToUpper(file.Status) == "FAILED" {
			continue
		}
		openedFile, openErr := s.store.Open(id.String(), file.StorageKey)
		if openErr != nil {
			return 0, 0, openErr
		}
		opened = append(opened, openedFile)
		payload = append(payload, hermes.KnowledgeFile{Name: knowledgeName(file), Path: file.StorageKey, Body: openedFile})
	}
	if err := s.docker().SyncKnowledge(ctx, profileKey, payload); err != nil {
		return 0, 0, errHermesSync
	}
	if err := s.repo.MarkFilesReady(ctx, id); err != nil {
		return 0, 0, err
	}
	return len(payload), s.startKnowledgeReading(id, profileKey), nil
}

// knowledgeStatus reports, per document, whether the assistant has finished reading it.
func (s *Server) knowledgeStatus(c *fiber.Ctx) error {
	id, err := tenantIDParam(c)
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "INVALID_TENANT_ID")
	}
	profileKey, err := s.repo.HermesProfileKey(c.Context(), id)
	if err != nil {
		return c.JSON(fiber.Map{"prepared": false, "reading": false, "total": 0, "ready": 0, "failed": 0, "files": fiber.Map{}})
	}
	files, err := s.repo.ListFiles(c.Context(), id)
	if err != nil {
		return err
	}
	ready, _ := s.docker().ReadyDocuments(c.Context(), profileKey)
	_, reading := s.knowledge.running.Load(profileKey)
	failures, _ := s.knowledge.failures.Load(id)
	states := fiber.Map{}
	total, done, failed := 0, 0, 0
	for _, file := range files {
		// Files not yet handed to the assistant are reported by their own status, not here.
		if status := strings.ToUpper(file.Status); status == "FAILED" || status == "PENDING_SYNC" {
			continue
		}
		name := knowledgeName(file)
		total++
		state := "UNREAD"
		if ready[name] {
			state, done = "READY", done+1
		} else if reading {
			state = "READING"
		}
		if state != "READY" && failures != nil {
			if _, bad := failures.(*sync.Map).Load(name); bad {
				state = "FAILED"
				if !reading {
					failed++
				}
			}
		}
		states[file.ID.String()] = state
	}
	return c.JSON(fiber.Map{"prepared": true, "reading": reading, "total": total, "ready": done, "failed": failed, "files": states})
}

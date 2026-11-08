package httpapi

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"ponpes-saas/internal/hermes"
	"ponpes-saas/internal/repository"
)

// knowledgeJobs tracks the background reading of tenant documents.
type knowledgeJobs struct {
	// running holds the profile keys that have a reading job in progress; there is one job per assistant.
	running sync.Map
	// failures holds, per tenant, the documents the last reading run could not read (*sync.Map of name -> true).
	failures sync.Map
}

func (s *Server) profileSettings(ctx context.Context, profileKey string, settings repository.HermesSettings) hermes.ProfileSettings {
	out := hermes.ProfileSettings{Description: settings.Description, SoulMD: settings.SoulMD, EnabledSkills: settings.EnabledSkills, TenantInfo: s.repo.ProfileInfoByKey(ctx, profileKey)}
	// A shared WhatsApp line routes non-PSB needs to the tenant's admins.
	if tenantID, err := s.repo.TenantByHermesProfileKey(ctx, profileKey); err == nil {
		if channel, err := s.repo.GetWhatsAppChannel(ctx, tenantID); err == nil && channel.ServiceMode == "GENERAL" {
			if admins, err := s.repo.ListWhatsAppAdmins(ctx, tenantID); err == nil && len(admins) > 0 {
				out.WhatsAppGeneral = true
				out.HandoffURL, out.HandoffKey = s.handoffURL(), s.cfg.HermesAPIKey
				for _, a := range admins {
					out.WhatsAppAdmins = append(out.WhatsAppAdmins, hermes.WhatsAppAdmin{Name: a.Name, Phone: a.Phone, NotifyWhen: a.NotifyWhen})
				}
			}
		}
	}
	return out
}

// refreshPersona rewrites an assistant's persona from the global settings and the pesantren profile.
func (s *Server) refreshPersona(ctx context.Context, profileKey string) {
	settings, err := s.repo.GetGlobalHermesSettings(ctx)
	if err != nil {
		return
	}
	if err := s.docker().ApplyProfileSettings(ctx, profileKey, s.profileSettings(ctx, profileKey, settings)); err != nil {
		slog.Warn("could not refresh assistant persona", "profile", profileKey, "error", err)
	}
}

// lockDown makes a profile safe and welcoming for the public: restricted tools and menu, plain wording, greeting.
func (s *Server) lockDown(ctx context.Context, profileKey string) error {
	client := s.docker()
	if err := client.ApplyClientExperience(ctx, profileKey); err != nil {
		return err
	}
	return client.InstallWelcomeHook(ctx, profileKey, s.repo.PesantrenNameByKey(ctx, profileKey))
}

// restartIfActive re-locks and restarts a running assistant: it only sees an updated persona after a restart,
// and a restart must never bring back a profile that still holds its powerful tools.
func (s *Server) restartIfActive(ctx context.Context, tenantID uuid.UUID, profileKey string) {
	if !s.assistantActive(ctx, tenantID) {
		return
	}
	if err := s.lockDown(ctx, profileKey); err != nil {
		slog.Error("could not lock down assistant before restart", "profile", profileKey, "error", err)
		return
	}
	_ = s.docker().ProfileAction(ctx, profileKey, "restart")
}

// assistantActive reports whether any channel (Telegram or WhatsApp) of the tenant is serving visitors.
func (s *Server) assistantActive(ctx context.Context, tenantID uuid.UUID) bool {
	if bot, err := s.repo.GetTelegramBot(ctx, tenantID); err == nil && bot.Status == "ACTIVE" {
		return true
	}
	channel, err := s.repo.GetWhatsAppChannel(ctx, tenantID)
	return err == nil && channel.Status == "ACTIVE"
}

// configureChannels writes the gateway settings of every connected channel into the profile.
func (s *Server) configureChannels(ctx context.Context, tenantID uuid.UUID, profileKey string) error {
	client := s.docker()
	if bot, err := s.repo.GetTelegramBot(ctx, tenantID); err == nil && bot.CredentialRef != "" {
		if err := client.ConfigureTelegram(ctx, profileKey, bot.CredentialRef, bot.AccessMode, bot.AllowedUserIDs); err != nil {
			return fmt.Errorf("telegram: %w", err)
		}
	}
	if channel, err := s.repo.GetWhatsAppChannel(ctx, tenantID); err == nil && channel.PhoneNumber != "" {
		if err := client.ConfigureWhatsApp(ctx, profileKey, channel.PhoneNumber, channel.AccessMode, channel.AllowedNumbers); err != nil {
			return fmt.Errorf("whatsapp: %w", err)
		}
	}
	return nil
}

// SyncProfilesOnStartup re-applies the global settings and the Telegram log hook to every assistant.
func (s *Server) SyncProfilesOnStartup(ctx context.Context) {
	settings, err := s.repo.GetGlobalHermesSettings(ctx)
	if err != nil {
		slog.Warn("could not load global Hermes settings", "error", err)
		return
	}
	keys, err := s.repo.HermesProfileKeys(ctx)
	if err != nil {
		slog.Warn("could not list Hermes profiles for global settings sync", "error", err)
		return
	}
	client := s.docker()
	for _, key := range keys {
		if err := client.ApplyProfileSettings(ctx, key, s.profileSettings(ctx, key, settings)); err != nil {
			slog.Warn("could not apply global Hermes settings", "profile", key, "error", err)
		}
		if err := client.InstallTelegramLogHook(ctx, key, s.eventsURL(), s.cfg.HermesAPIKey); err != nil {
			slog.Warn("could not install Hermes Telegram log hook", "profile", key, "error", err)
		}
	}
}

// startKnowledgeReading turns the tenant's documents into text for the assistant. Quick formats are ready
// at once; pictures and office files are read in the background. It returns how many need the slow path.
func (s *Server) startKnowledgeReading(tenantID uuid.UUID, profileKey string) int {
	files, err := s.repo.ListFiles(context.Background(), tenantID)
	if err != nil {
		return 0
	}
	fast, slow := []string{}, []string{}
	for _, file := range files {
		name := knowledgeName(file)
		if hermes.KnowledgeKind(file.Filename) == "text" {
			fast = append(fast, name)
		} else {
			slow = append(slow, name)
		}
	}
	if _, running := s.knowledge.running.LoadOrStore(profileKey, true); running {
		return len(slow)
	}
	go func() {
		defer s.knowledge.running.Delete(profileKey)
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Minute)
		defer cancel()
		client := s.docker()
		// Remember what failed this run so the dashboard can offer a retry; success is read from disk.
		failures := &sync.Map{}
		s.knowledge.failures.Store(tenantID, failures)
		track := func(name string, ok bool) {
			if !ok {
				failures.Store(name, true)
			}
		}
		createdFast, failedFast := client.ExtractKnowledgeText(ctx, profileKey, fast, track)
		if createdFast > 0 || len(slow) == 0 {
			s.refreshPersona(ctx, profileKey)
		}
		created, failed := client.ExtractKnowledgeText(ctx, profileKey, slow, track)
		slog.Info("knowledge read", "profile", profileKey, "created", createdFast+created, "failed", len(failedFast)+len(failed))
		if created > 0 {
			s.refreshPersona(ctx, profileKey)
			s.restartIfActive(ctx, tenantID, profileKey)
		}
		if created > 0 || len(failed) > 0 || len(failedFast) > 0 {
			s.repo.RecordAudit(context.Background(), tenantID, nil, "KNOWLEDGE_IMAGES_READ", "TENANT", tenantID.String(), "")
		}
	}()
	return len(slow)
}

// knowledgeName is the document's name inside the assistant's knowledge folder.
func knowledgeName(file repository.File) string {
	if file.FolderID != nil {
		return "folder_" + file.FolderID.String() + "/" + file.Filename
	}
	return file.Filename
}

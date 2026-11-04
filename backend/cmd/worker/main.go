package main

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"ponpes-saas/internal/config"
	"ponpes-saas/internal/db"
	"ponpes-saas/internal/hermes"
	"ponpes-saas/internal/jobs"
	"ponpes-saas/internal/repository"
)

func main() {
	cfg := config.Load()
	if cfg.DatabaseURL == "" {
		slog.Warn("provisioning worker stopped", "reason", "DATABASE_URL is not configured")
		return
	}
	pool, err := db.Open(context.Background(), cfg.DatabaseURL)
	if err != nil {
		slog.Error("worker database open failed", "error", err)
		return
	}
	defer pool.Close()
	hermesClient, err := hermes.NewControl(hermes.ControlConfig{Mode: cfg.HermesExecMode, Container: cfg.HermesContainer, BridgeSocket: cfg.HermesBridgeSocket})
	if err != nil {
		slog.Error("worker Hermes control misconfigured", "error", err)
		return
	}
	slog.Info("Hermes control ready", "mode", cfg.HermesExecMode)
	repo := jobs.Repository{DB: pool, Lease: time.Minute}
	slog.Info("provisioning worker ready")
	for {
		job, err := repo.Claim(context.Background())
		if errors.Is(err, jobs.ErrNoJob) {
			time.Sleep(2 * time.Second)
			continue
		}
		if err != nil {
			slog.Error("provisioning job claim failed", "error", err)
			time.Sleep(2 * time.Second)
			continue
		}
		if _, err := hermesClient.CreateParkedProfile(context.Background(), job.TenantID, job.ProfileKey); err != nil {
			slog.Error("provisioning profile failed", "job_id", job.ID, "tenant_id", job.TenantID, "attempt", job.Attempt, "error", err)
			_ = repo.Fail(context.Background(), job.ID, "PROFILE_CREATE_FAILED")
			continue
		}
		if settings, settingsErr := (repository.Repository{DB: pool}).GetGlobalHermesSettings(context.Background()); settingsErr == nil {
			if applyErr := hermesClient.ApplyProfileSettings(context.Background(), job.ProfileKey, hermes.ProfileSettings{Description: settings.Description, SoulMD: settings.SoulMD, EnabledSkills: settings.EnabledSkills, TenantInfo: (repository.Repository{DB: pool}).ProfileInfoByKey(context.Background(), job.ProfileKey)}); applyErr != nil {
				slog.Error("global Hermes settings apply failed", "job_id", job.ID, "tenant_id", job.TenantID, "error", applyErr)
				_ = repo.Fail(context.Background(), job.ID, "GLOBAL_SETTINGS_APPLY_FAILED")
				continue
			}
		} else {
			slog.Warn("global Hermes settings could not be loaded", "error", settingsErr)
		}
		if err := hermesClient.InstallTelegramLogHook(context.Background(), job.ProfileKey, cfg.HermesEventsURL, cfg.HermesAPIKey); err != nil {
			slog.Warn("Telegram log hook installation failed", "profile", job.ProfileKey, "error", err)
		}
		slog.Info("provisioning profile created", "job_id", job.ID, "tenant_id", job.TenantID, "attempt", job.Attempt)
		if err := repo.Complete(context.Background(), job.ID, job.ProfileKey, "HEALTHY"); err != nil {
			slog.Error("provisioning job completion failed", "job_id", job.ID, "tenant_id", job.TenantID, "error", err)
		} else {
			slog.Info("provisioning job completed", "job_id", job.ID, "tenant_id", job.TenantID)
		}
	}
}

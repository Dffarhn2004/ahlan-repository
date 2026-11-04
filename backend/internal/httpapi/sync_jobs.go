package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"ponpes-saas/internal/repository"
)

// Syncing every assistant takes minutes (several steps per profile plus a gateway restart), far longer than a
// browser, a mobile connection or a reverse proxy should be asked to hold one request open. So "sync all" runs as a
// background job: POST answers 202 with a job id at once and the page polls the job for its progress.

const (
	syncJobRunning = "RUNNING"
	syncJobDone    = "DONE"
	syncJobFailed  = "FAILED"

	// syncJobTimeout bounds the whole run, so a hung Hermes cannot keep a job (and the "already running" lock) forever.
	syncJobTimeout = 20 * time.Minute
	// syncJobsKept is how many finished jobs stay available for polling.
	syncJobsKept = 10
)

// syncResult is the outcome for one assistant.
type syncResult struct {
	TenantID   uuid.UUID `json:"tenant_id"`
	TenantName string    `json:"tenant_name"`
	Status     string    `json:"status"`
	Failed     []string  `json:"failed,omitempty"`
	Restarted  bool      `json:"restarted"`
}

// syncJob is a snapshot-able record of one sync-all run.
type syncJob struct {
	ID         string       `json:"job_id"`
	Status     string       `json:"status"`
	Phase      string       `json:"phase"`
	Total      int          `json:"total"`
	Done       int          `json:"done"`
	Results    []syncResult `json:"data"`
	Error      string       `json:"error,omitempty"`
	StartedAt  time.Time    `json:"started_at"`
	FinishedAt *time.Time   `json:"finished_at,omitempty"`
}

// syncJobStore keeps the running job and the last few finished ones in memory. A restart of the API loses them,
// which is acceptable: the run itself is idempotent and can simply be started again.
type syncJobStore struct {
	mu      sync.Mutex
	jobs    map[string]*syncJob
	order   []string
	running string
}

// begin registers a new job, or returns the one already running (only one sync-all may run at a time).
func (st *syncJobStore) begin() (job syncJob, started bool) {
	st.mu.Lock()
	defer st.mu.Unlock()
	if st.running != "" {
		return st.snapshotLocked(st.running), false
	}
	if st.jobs == nil {
		st.jobs = map[string]*syncJob{}
	}
	id := uuid.NewString()
	st.jobs[id] = &syncJob{ID: id, Status: syncJobRunning, Phase: "Menyiapkan", Results: []syncResult{}, StartedAt: time.Now().UTC()}
	st.order = append(st.order, id)
	st.running = id
	for len(st.order) > syncJobsKept {
		delete(st.jobs, st.order[0])
		st.order = st.order[1:]
	}
	return st.snapshotLocked(id), true
}

func (st *syncJobStore) snapshotLocked(id string) syncJob {
	job := *st.jobs[id]
	job.Results = append([]syncResult(nil), job.Results...)
	return job
}

func (st *syncJobStore) get(id string) (syncJob, bool) {
	st.mu.Lock()
	defer st.mu.Unlock()
	if _, ok := st.jobs[id]; !ok {
		return syncJob{}, false
	}
	return st.snapshotLocked(id), true
}

// update changes a job under the lock; finish ends it and frees the "already running" slot.
func (st *syncJobStore) update(id string, change func(*syncJob)) {
	st.mu.Lock()
	defer st.mu.Unlock()
	if job, ok := st.jobs[id]; ok {
		change(job)
	}
}

func (st *syncJobStore) finish(id, status, message string) {
	st.mu.Lock()
	defer st.mu.Unlock()
	if job, ok := st.jobs[id]; ok {
		now := time.Now().UTC()
		job.Status, job.Error, job.FinishedAt, job.Phase = status, message, &now, "Selesai"
	}
	if st.running == id {
		st.running = ""
	}
}

// startSyncAll answers 202 immediately and runs the sync in the background.
func (s *Server) startSyncAll(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	var input struct {
		Restart bool `json:"restart"`
	}
	_ = c.BodyParser(&input)
	job, started := s.syncJobs.begin()
	if started {
		// The request context ends with this response, so the run gets its own bounded context.
		go s.runSyncAll(job.ID, userID, input.Restart)
	}
	return c.Status(http.StatusAccepted).JSON(fiber.Map{"job_id": job.ID, "status": job.Status, "already_running": !started})
}

// syncAllStatus reports the progress and, once finished, the per-assistant results of a job.
func (s *Server) syncAllStatus(c *fiber.Ctx) error {
	job, ok := s.syncJobs.get(c.Params("job_id"))
	if !ok {
		return fiber.NewError(http.StatusNotFound, "SYNC_JOB_NOT_FOUND")
	}
	return c.JSON(job)
}

// restartCandidate remembers which sync result belongs to which profile until the shared restart decision is made.
type restartCandidate struct {
	index   int
	profile repository.HermesProfileRef
	wanted  bool
}

// runSyncAll re-applies lockdown, persona, channels and the log hook to every assistant, optionally restarting them.
func (s *Server) runSyncAll(jobID string, userID uuid.UUID, restart bool) {
	ctx, cancel := context.WithTimeout(context.Background(), syncJobTimeout)
	defer cancel()
	defer func() {
		if recovered := recover(); recovered != nil {
			slog.Error("sync-all crashed", "job", jobID, "panic", recovered)
			s.syncJobs.finish(jobID, syncJobFailed, "SYNC_CRASHED")
		}
	}()

	profiles, err := s.repo.HermesProfiles(ctx)
	if err != nil {
		slog.Error("sync-all could not list assistants", "job", jobID, "error", err)
		s.syncJobs.finish(jobID, syncJobFailed, "PROFILES_UNAVAILABLE")
		return
	}
	settings, err := s.repo.GetGlobalHermesSettings(ctx)
	if err != nil {
		slog.Error("sync-all could not load global settings", "job", jobID, "error", err)
		s.syncJobs.finish(jobID, syncJobFailed, "SETTINGS_UNAVAILABLE")
		return
	}
	client := s.docker()
	s.syncJobs.update(jobID, func(job *syncJob) { job.Total, job.Phase = len(profiles), "Menerapkan pengaturan" })

	results := make([]syncResult, 0, len(profiles))
	toRestart := make([]restartCandidate, 0, len(profiles))
	for _, profile := range profiles {
		failed := []string{}
		if err := s.lockDown(ctx, profile.ProfileKey); err != nil {
			failed = append(failed, "pengaturan layanan")
		}
		if err := client.ApplyProfileSettings(ctx, profile.ProfileKey, s.profileSettings(ctx, profile.ProfileKey, settings)); err != nil {
			failed = append(failed, "persona asisten")
		}
		_, botErr := s.repo.GetTelegramBot(ctx, profile.TenantID)
		_, waErr := s.repo.GetWhatsAppChannel(ctx, profile.TenantID)
		hasChannel := botErr == nil || waErr == nil
		if hasChannel {
			if err := s.configureChannels(ctx, profile.TenantID, profile.ProfileKey); err != nil {
				failed = append(failed, "koneksi saluran")
			}
			if strings.TrimSpace(s.cfg.HermesAPIKey) != "" {
				if err := client.InstallTelegramLogHook(ctx, profile.ProfileKey, s.eventsURL(), s.cfg.HermesAPIKey); err != nil {
					failed = append(failed, "pencatatan percakapan")
				}
			}
		}
		// Running assistants only pick the new settings up after a restart; that happens below, once for everyone.
		toRestart = append(toRestart, restartCandidate{index: len(results), profile: profile,
			wanted: restart && hasChannel && s.assistantActive(ctx, profile.TenantID) && len(failed) == 0})
		s.startKnowledgeReading(profile.TenantID, profile.ProfileKey)
		s.repo.RecordAudit(ctx, profile.TenantID, &userID, "ASSISTANT_SYNCED", "TENANT", profile.TenantID.String(), "")
		results = append(results, syncResult{TenantID: profile.TenantID, TenantName: profile.TenantName, Status: "OK", Failed: failed})
		snapshot := append([]syncResult(nil), results...)
		s.syncJobs.update(jobID, func(job *syncJob) { job.Done, job.Results = len(snapshot), snapshot })
	}

	// Hermes keeps the locale overlays it has read in memory until the shared gateway process restarts, so
	// restarting individual profiles does not make a new overlay visible. One restart of the shared gateway does,
	// and it brings every profile back, so it replaces the per-profile restarts. If it is refused (too soon after
	// the previous one) or fails, fall back to restarting each profile on its own.
	anyWanted := false
	for _, candidate := range toRestart {
		anyWanted = anyWanted || candidate.wanted
	}
	if anyWanted {
		s.syncJobs.update(jobID, func(job *syncJob) { job.Phase = "Memulai ulang asisten" })
		if err := client.RestartGateway(ctx); err != nil {
			slog.Warn("could not restart the shared gateway; restarting profiles one by one", "error", err)
			for _, candidate := range toRestart {
				if !candidate.wanted {
					continue
				}
				if err := client.ProfileAction(ctx, candidate.profile.ProfileKey, "restart"); err != nil {
					results[candidate.index].Failed = append(results[candidate.index].Failed, "mulai ulang")
				} else {
					results[candidate.index].Restarted = true
				}
			}
		} else {
			for _, candidate := range toRestart {
				results[candidate.index].Restarted = candidate.wanted
			}
			s.repo.RecordGlobalAudit(ctx, &userID, "HERMES_GATEWAY_RESTART", "HERMES_GATEWAY", "shared")
		}
	}
	for i := range results {
		if len(results[i].Failed) > 0 {
			results[i].Status = "FAILED"
		}
	}
	s.syncJobs.update(jobID, func(job *syncJob) { job.Done, job.Results = len(results), results })
	s.syncJobs.finish(jobID, syncJobDone, "")
}

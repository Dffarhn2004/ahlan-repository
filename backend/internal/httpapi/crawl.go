package httpapi

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"ponpes-saas/internal/crawler"
	"ponpes-saas/internal/repository"
)

// crawlFolderName is where fetched pages appear in the pesantren's file manager.
const crawlFolderName = "Sumber Web"

const (
	crawlTick       = 30 * time.Second
	crawlRunTimeout = 15 * time.Minute
	crawlMaxStore   = 21 << 20
)

type crawlSchedule struct {
	Mode            string `json:"mode"`
	IntervalMinutes int    `json:"interval_minutes"`
}

func (s *Server) listCrawlSources(c *fiber.Ctx) error {
	id, err := tenantIDParam(c)
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "INVALID_TENANT_ID")
	}
	sources, err := s.repo.ListCrawlSources(c.Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"data": sources})
}

func scheduleError(err error) error {
	if errors.Is(err, crawler.ErrScheduleTooFrequent) {
		return fiber.NewError(http.StatusUnprocessableEntity, "CRAWL_SCHEDULE_TOO_FREQUENT")
	}
	return fiber.NewError(http.StatusBadRequest, "INVALID_REQUEST")
}

func (s *Server) createCrawlSource(c *fiber.Ctx) error {
	id, err := tenantIDParam(c)
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "INVALID_TENANT_ID")
	}
	var in struct {
		URL      string        `json:"url"`
		Schedule crawlSchedule `json:"schedule"`
	}
	if err := c.BodyParser(&in); err != nil {
		return fiber.NewError(http.StatusBadRequest, "INVALID_REQUEST")
	}
	if in.Schedule.Mode == "" {
		in.Schedule.Mode = crawler.ModeManual
	}
	if err := crawler.ValidateSchedule(in.Schedule.Mode, in.Schedule.IntervalMinutes); err != nil {
		return scheduleError(err)
	}
	// The kind is derived here from the URL; whatever the browser claims is ignored.
	target, err := crawler.ParseSource(in.URL)
	if err != nil {
		return fiber.NewError(http.StatusUnprocessableEntity, "CRAWL_URL_FORBIDDEN")
	}
	if !target.IsGoogle() {
		ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
		defer cancel()
		if err := crawler.CheckHost(ctx, target.Start.Hostname()); err != nil {
			code := "CRAWL_URL_FORBIDDEN"
			if errors.Is(err, crawler.ErrUnreachable) {
				code = "CRAWL_URL_UNREACHABLE"
			}
			return fiber.NewError(http.StatusUnprocessableEntity, code)
		}
	}
	source := repository.CrawlSource{TenantID: id, URL: target.Start.String(), Kind: string(target.Kind), MaxDepth: crawler.DefaultMaxDepth, MaxPages: crawler.DefaultMaxPages, ScheduleMode: in.Schedule.Mode}
	if source.ScheduleMode == crawler.ModeInterval {
		source.IntervalMinutes = &in.Schedule.IntervalMinutes
	}
	actor := c.Locals("user_id").(uuid.UUID)
	created, err := s.repo.CreateCrawlSource(c.Context(), source, actor)
	switch {
	case errors.Is(err, repository.ErrCrawlLimit):
		return fiber.NewError(http.StatusConflict, "CRAWL_SOURCE_LIMIT")
	case errors.Is(err, repository.ErrConflict):
		return fiber.NewError(http.StatusConflict, "CRAWL_SOURCE_EXISTS")
	case err != nil:
		return err
	}
	s.repo.RecordAudit(c.Context(), id, &actor, "CRAWL_SOURCE_CREATED", "CRAWL_SOURCE", created.ID.String(), created.URL)
	s.wakeCrawler()
	return c.Status(http.StatusCreated).JSON(created)
}

func crawlSourceID(c *fiber.Ctx) (uuid.UUID, uuid.UUID, error) {
	tenantID, err := tenantIDParam(c)
	if err != nil {
		return uuid.Nil, uuid.Nil, fiber.NewError(http.StatusBadRequest, "INVALID_TENANT_ID")
	}
	sourceID, err := uuid.Parse(c.Params("sid"))
	if err != nil {
		return uuid.Nil, uuid.Nil, fiber.NewError(http.StatusNotFound, "CRAWL_SOURCE_NOT_FOUND")
	}
	return tenantID, sourceID, nil
}

func (s *Server) updateCrawlSource(c *fiber.Ctx) error {
	tenantID, sourceID, err := crawlSourceID(c)
	if err != nil {
		return err
	}
	var in struct {
		Enabled  *bool          `json:"enabled"`
		Schedule *crawlSchedule `json:"schedule"`
	}
	if err := c.BodyParser(&in); err != nil {
		return fiber.NewError(http.StatusBadRequest, "INVALID_REQUEST")
	}
	current, err := s.repo.GetCrawlSource(c.Context(), tenantID, sourceID)
	if err != nil {
		return fiber.NewError(http.StatusNotFound, "CRAWL_SOURCE_NOT_FOUND")
	}
	if current.ForcedDisabled && in.Enabled != nil && *in.Enabled {
		return fiber.NewError(http.StatusForbidden, "CRAWL_FORCED_DISABLED")
	}
	patch := repository.CrawlSourcePatch{Enabled: in.Enabled}
	if in.Schedule != nil {
		if err := crawler.ValidateSchedule(in.Schedule.Mode, in.Schedule.IntervalMinutes); err != nil {
			return scheduleError(err)
		}
		patch.ScheduleMode = &in.Schedule.Mode
		if in.Schedule.Mode == crawler.ModeInterval {
			patch.IntervalMinutes = &in.Schedule.IntervalMinutes
		}
	}
	updated, err := s.repo.UpdateCrawlSource(c.Context(), tenantID, sourceID, patch)
	if err != nil {
		return fiber.NewError(http.StatusNotFound, "CRAWL_SOURCE_NOT_FOUND")
	}
	s.repo.RecordAudit(c.Context(), tenantID, auditActor(c), "CRAWL_SOURCE_UPDATED", "CRAWL_SOURCE", sourceID.String(), updated.URL)
	s.wakeCrawler()
	return c.JSON(updated)
}

func (s *Server) deleteCrawlSource(c *fiber.Ctx) error {
	tenantID, sourceID, err := crawlSourceID(c)
	if err != nil {
		return err
	}
	source, err := s.repo.GetCrawlSource(c.Context(), tenantID, sourceID)
	if err != nil {
		return fiber.NewError(http.StatusNotFound, "CRAWL_SOURCE_NOT_FOUND")
	}
	if err := s.removeCrawlSource(c.Context(), tenantID, sourceID); err != nil {
		return err
	}
	s.repo.RecordAudit(c.Context(), tenantID, auditActor(c), "CRAWL_SOURCE_DELETED", "CRAWL_SOURCE", sourceID.String(), source.URL)
	return c.SendStatus(http.StatusNoContent)
}

// removeCrawlSource deletes a source with the files it produced, and tells the assistant to forget them.
func (s *Server) removeCrawlSource(ctx context.Context, tenantID, sourceID uuid.UUID) error {
	files, err := s.repo.DeleteCrawlSource(ctx, tenantID, sourceID)
	if err != nil {
		return err
	}
	for _, f := range files {
		_ = s.store.Delete(tenantID.String(), f.StorageKey)
	}
	if len(files) > 0 {
		go s.syncAfterCrawl(tenantID) // so the assistant stops answering from the removed pages
	}
	return nil
}

func (s *Server) runCrawlSource(c *fiber.Ctx) error {
	tenantID, sourceID, err := crawlSourceID(c)
	if err != nil {
		return err
	}
	source, err := s.repo.GetCrawlSource(c.Context(), tenantID, sourceID)
	if err != nil {
		return fiber.NewError(http.StatusNotFound, "CRAWL_SOURCE_NOT_FOUND")
	}
	if !source.Enabled || source.ForcedDisabled {
		return fiber.NewError(http.StatusUnprocessableEntity, "CRAWL_SOURCE_DISABLED")
	}
	actor := c.Locals("user_id").(uuid.UUID)
	runID, err := s.repo.EnqueueCrawlRun(c.Context(), tenantID, sourceID, "MANUAL", &actor)
	switch {
	case errors.Is(err, repository.ErrConflict):
		return fiber.NewError(http.StatusConflict, "CRAWL_ALREADY_RUNNING")
	case errors.Is(err, repository.ErrCrawlTooSoon):
		return fiber.NewError(http.StatusTooManyRequests, "CRAWL_RUN_TOO_SOON")
	case errors.Is(err, repository.ErrNotFound):
		return fiber.NewError(http.StatusNotFound, "CRAWL_SOURCE_NOT_FOUND")
	case err != nil:
		return err
	}
	s.repo.RecordAudit(c.Context(), tenantID, &actor, "CRAWL_RUN_MANUAL", "CRAWL_SOURCE", sourceID.String(), source.URL)
	s.wakeCrawler()
	return c.Status(http.StatusAccepted).JSON(fiber.Map{"run_id": runID, "status": "QUEUED"})
}

// forceCrawlSource lets a superadmin switch a tenant's source off (or back on) for abuse or cost reasons.
func (s *Server) forceCrawlSource(c *fiber.Ctx) error {
	sourceID, err := uuid.Parse(c.Params("sid"))
	if err != nil {
		return fiber.NewError(http.StatusNotFound, "CRAWL_SOURCE_NOT_FOUND")
	}
	var in struct {
		Forced bool `json:"forced_disabled"`
	}
	if err := c.BodyParser(&in); err != nil {
		return fiber.NewError(http.StatusBadRequest, "INVALID_REQUEST")
	}
	source, err := s.repo.GetCrawlSourceByID(c.Context(), sourceID)
	if err != nil {
		return fiber.NewError(http.StatusNotFound, "CRAWL_SOURCE_NOT_FOUND")
	}
	if err := s.repo.SetCrawlSourceForced(c.Context(), sourceID, in.Forced); err != nil {
		return err
	}
	s.repo.RecordAudit(c.Context(), source.TenantID, auditActor(c), "CRAWL_SOURCE_FORCED", "CRAWL_SOURCE", sourceID.String(), fmt.Sprintf("forced_disabled=%t", in.Forced))
	return c.JSON(fiber.Map{"id": sourceID, "forced_disabled": in.Forced})
}

// wakeCrawler nudges the worker loop so a click does not wait for the next tick.
func (s *Server) wakeCrawler() {
	select {
	case s.crawlWake <- struct{}{}:
	default:
	}
}

// StartCrawlWorker runs the scheduler and the crawl queue until ctx ends. State lives in PostgreSQL (rows are claimed with
// FOR UPDATE SKIP LOCKED), so running several API instances never fetches a source twice.
func (s *Server) StartCrawlWorker(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(crawlTick)
		defer ticker.Stop()
		for {
			s.crawlOnce(ctx)
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			case <-s.crawlWake:
			}
		}
	}()
}

func (s *Server) crawlOnce(ctx context.Context) {
	if n, err := s.repo.ScheduleDueCrawls(ctx); err != nil {
		slog.Error("crawl scheduling failed", "error", err)
	} else if n > 0 {
		slog.Info("crawl runs scheduled", "count", n)
	}
	for ctx.Err() == nil {
		run, err := s.repo.ClaimCrawlRun(ctx)
		if errors.Is(err, repository.ErrNotFound) {
			return
		}
		if err != nil {
			slog.Error("crawl claim failed", "error", err)
			return
		}
		s.executeCrawl(ctx, run)
	}
}

// crawlFailure turns a crawler error into the run status and the plain-language message the dashboard shows.
func crawlFailure(err error) (status, code, message string) {
	switch {
	case errors.Is(err, crawler.ErrNeedsPublicSharing):
		return "NEEDS_PUBLIC_SHARING", "NEEDS_PUBLIC_SHARING", "Dokumen belum dibagikan sebagai \"Siapa saja yang memiliki link\"."
	case errors.Is(err, crawler.ErrForbiddenURL):
		return "FAILED", "CRAWL_URL_FORBIDDEN", "Tautan ditolak: tidak aman, tidak dikenali, atau dilarang oleh robots.txt situs."
	case errors.Is(err, crawler.ErrTooLarge):
		return "FAILED", "CRAWL_TOO_LARGE", "Isi terlalu besar untuk diambil."
	case errors.Is(err, crawler.ErrUnsupportedType):
		return "FAILED", "CRAWL_UNSUPPORTED_TYPE", "Jenis file tidak didukung. Gunakan PDF, gambar, atau teks."
	}
	return "FAILED", "CRAWL_UNREACHABLE", "Situs tidak dapat dijangkau atau tidak merespons."
}

func (s *Server) executeCrawl(parent context.Context, run repository.CrawlRun) {
	ctx, cancel := context.WithTimeout(parent, crawlRunTimeout)
	defer cancel()
	finish := func(status string, counts repository.CrawlCounts, code, message string) {
		// Use a fresh context so a cancelled run still records how it ended.
		if err := s.repo.FinishCrawlRun(context.Background(), run, status, counts, code, message); err != nil {
			slog.Error("crawl run could not be recorded", "run_id", run.ID, "error", err)
		}
	}
	defer func() {
		if r := recover(); r != nil {
			slog.Error("crawl run panicked", "run_id", run.ID, "panic", r)
			finish("FAILED", repository.CrawlCounts{}, "CRAWL_INTERNAL", "Terjadi kesalahan internal saat mengambil sumber.")
		}
	}()

	source, err := s.repo.GetCrawlSourceByID(ctx, run.SourceID)
	if err != nil {
		finish("FAILED", repository.CrawlCounts{}, "CRAWL_SOURCE_MISSING", "Sumber tidak ditemukan.")
		return
	}
	if !source.Enabled || source.ForcedDisabled {
		finish("FAILED", repository.CrawlCounts{}, "CRAWL_SOURCE_DISABLED", "Sumber sedang dijeda.")
		return
	}
	target, err := crawler.ParseSource(source.URL)
	if err != nil {
		finish("FAILED", repository.CrawlCounts{}, "CRAWL_URL_FORBIDDEN", "Tautan ditolak karena tidak aman.")
		return
	}

	started := time.Now()
	result, err := s.crawler.Run(ctx, target, crawler.Options{MaxDepth: source.MaxDepth, MaxPages: source.MaxPages, Delay: time.Second})
	if err != nil {
		status, code, message := crawlFailure(err)
		slog.Warn("crawl run failed", "run_id", run.ID, "source_id", run.SourceID, "error", err)
		finish(status, repository.CrawlCounts{Failed: result.Failed}, code, message)
		return
	}

	counts := repository.CrawlCounts{Failed: result.Failed}
	folderID, err := s.repo.EnsureRootFolder(ctx, run.TenantID, crawlFolderName)
	if err != nil {
		finish("FAILED", counts, "CRAWL_INTERNAL", "Folder penyimpanan tidak dapat disiapkan.")
		return
	}
	now := time.Now()
	for _, page := range result.Pages {
		outcome, err := s.storeCrawledPage(ctx, run, source, folderID, page, now)
		if err != nil {
			slog.Error("crawled page could not be stored", "run_id", run.ID, "url", page.URL, "error", err)
			counts.Failed++
			continue
		}
		switch outcome {
		case "new":
			counts.New++
		case "changed":
			counts.Changed++
		default:
			counts.Unchanged++
		}
	}

	removed := 0
	// A page is only retired after a complete run: a truncated or partly failed crawl proves nothing about pages it never reached.
	if !result.Truncated && counts.Failed == 0 && result.Blocked == 0 {
		gone, err := s.repo.MarkUnseenCrawledPages(ctx, run.TenantID, run.SourceID, run.ID, started)
		if err != nil {
			slog.Error("could not retire missing pages", "run_id", run.ID, "error", err)
		}
		for _, f := range gone {
			_ = s.store.Delete(run.TenantID.String(), f.StorageKey)
		}
		removed = len(gone)
	}

	status, message := "SUCCEEDED", ""
	switch {
	case counts.New+counts.Changed+counts.Unchanged == 0:
		status, message = "FAILED", "Tidak ada halaman yang dapat disimpan."
	case counts.Failed > 0 || result.Blocked > 0:
		status, message = "PARTIAL", fmt.Sprintf("%d halaman gagal diambil, %d dilarang robots.txt.", counts.Failed, result.Blocked)
	case result.Truncated:
		status, message = "PARTIAL", fmt.Sprintf("Dibatasi %d halaman per pengambilan.", source.MaxPages)
	}
	finish(status, counts, "", message)
	s.repo.RecordAudit(context.Background(), run.TenantID, nil, "CRAWL_RUN_FINISHED", "CRAWL_SOURCE", run.SourceID.String(),
		fmt.Sprintf("%s: baru %d, berubah %d, dihapus %d", source.URL, counts.New, counts.Changed, removed))
	slog.Info("crawl run finished", "run_id", run.ID, "status", status, "new", counts.New, "changed", counts.Changed, "unchanged", counts.Unchanged, "removed", removed)

	if counts.New+counts.Changed+removed > 0 {
		s.syncAfterCrawl(run.TenantID)
	}
}

// syncAfterCrawl hands the updated documents to the assistant. Without an assistant yet the files simply wait as
// PENDING_SYNC, and the dashboard's own sync picks them up later.
func (s *Server) syncAfterCrawl(tenantID uuid.UUID) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	profileKey, err := s.repo.HermesProfileKey(ctx, tenantID)
	if err != nil {
		return
	}
	if _, _, err := s.syncTenantKnowledge(ctx, tenantID, profileKey); err != nil {
		slog.Warn("could not sync crawled knowledge to the assistant", "tenant_id", tenantID, "error", err)
	}
}

func crawlContentType(ext string) string {
	switch ext {
	case ".csv":
		return "text/csv; charset=utf-8"
	case ".pdf":
		return "application/pdf"
	case ".png":
		return "image/png"
	case ".jpg":
		return "image/jpeg"
	}
	return "text/plain; charset=utf-8"
}

var nonSlug = regexp.MustCompile(`[^a-z0-9]+`)

// crawlFilename is chosen once, when the page is first stored, and never changes: the assistant addresses documents by
// name, so a title edit on the website must not make the old document look deleted. The URL hash keeps names unique.
func crawlFilename(page crawler.Page) string {
	return buildCrawlFilename(page.URL, page.Title, page.Ext)
}

func crawlFilenameFor(page repository.CrawledPage) string {
	return buildCrawlFilename(page.URL, page.Title, page.Ext)
}

func buildCrawlFilename(pageURL, title, ext string) string {
	label := title
	if u, err := url.Parse(pageURL); err == nil && label == "" {
		label = u.Hostname() + " " + u.Path
	}
	slug := strings.Trim(nonSlug.ReplaceAllString(strings.ToLower(label), "-"), "-")
	if len(slug) > 50 {
		slug = strings.Trim(slug[:50], "-")
	}
	if slug == "" {
		slug = "halaman"
	}
	sum := sha1.Sum([]byte(pageURL))
	return "web-" + slug + "-" + hex.EncodeToString(sum[:])[:8] + ext
}

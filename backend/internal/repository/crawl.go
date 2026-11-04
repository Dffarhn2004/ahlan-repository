package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	// ErrConflict is a unique-constraint clash or an already-active crawl run.
	ErrConflict = errors.New("conflict")
	// ErrCrawlLimit means the tenant already has the maximum number of crawl sources.
	ErrCrawlLimit = errors.New("crawl source limit reached")
	// ErrCrawlTooSoon means a manual crawl was requested again before the cool-down passed.
	ErrCrawlTooSoon = errors.New("manual crawl too soon")
)

const (
	MaxCrawlSourcesPerTenant = 5
	ManualCrawlCooldown      = 10 * time.Minute
	crawlStaleAfter          = 20 * time.Minute
)

type CrawlSource struct {
	ID                 uuid.UUID  `json:"id"`
	TenantID           uuid.UUID  `json:"tenant_id"`
	URL                string     `json:"url"`
	Kind               string     `json:"kind"`
	MaxDepth           int        `json:"-"`
	MaxPages           int        `json:"-"`
	ScheduleMode       string     `json:"schedule_mode"`
	IntervalMinutes    *int       `json:"interval_minutes,omitempty"`
	Enabled            bool       `json:"enabled"`
	ForcedDisabled     bool       `json:"forced_disabled"`
	NextRunAt          *time.Time `json:"next_run_at,omitempty"`
	LastRunAt          *time.Time `json:"last_run_at,omitempty"`
	LastStatus         string     `json:"last_status,omitempty"`
	ErrorPublicMessage string     `json:"error_public_message,omitempty"`
	PagesCount         int        `json:"pages_count"`
}

type CrawlRun struct {
	ID       uuid.UUID
	TenantID uuid.UUID
	SourceID uuid.UUID
	Trigger  string
}

type CrawlCounts struct{ New, Changed, Unchanged, Failed int }

const crawlSourceColumns = `s.id, s.tenant_id, s.url, s.kind, s.max_depth, s.max_pages, s.schedule_mode, s.interval_minutes, s.enabled, s.forced_disabled,
  s.next_run_at, s.last_run_at, coalesce(s.last_status,''), coalesce(s.error_public_message,''),
  (SELECT count(*) FROM crawled_pages p WHERE p.source_id=s.id AND p.status='ACTIVE')`

func scanCrawlSource(row pgx.Row) (CrawlSource, error) {
	var s CrawlSource
	err := row.Scan(&s.ID, &s.TenantID, &s.URL, &s.Kind, &s.MaxDepth, &s.MaxPages, &s.ScheduleMode, &s.IntervalMinutes, &s.Enabled, &s.ForcedDisabled,
		&s.NextRunAt, &s.LastRunAt, &s.LastStatus, &s.ErrorPublicMessage, &s.PagesCount)
	if errors.Is(err, pgx.ErrNoRows) {
		return s, ErrNotFound
	}
	return s, err
}

func isUniqueViolation(err error) bool {
	var pg *pgconn.PgError
	return errors.As(err, &pg) && pg.Code == "23505"
}

func (r Repository) ListCrawlSources(ctx context.Context, tenantID uuid.UUID) ([]CrawlSource, error) {
	rows, err := r.DB.Query(ctx, `SELECT `+crawlSourceColumns+` FROM crawl_sources s WHERE s.tenant_id=$1 ORDER BY s.created_at`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []CrawlSource{}
	for rows.Next() {
		s, err := scanCrawlSource(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r Repository) GetCrawlSource(ctx context.Context, tenantID, id uuid.UUID) (CrawlSource, error) {
	return scanCrawlSource(r.DB.QueryRow(ctx, `SELECT `+crawlSourceColumns+` FROM crawl_sources s WHERE s.id=$1 AND s.tenant_id=$2`, id, tenantID))
}

// GetCrawlSourceByID is for the worker, which already trusts the run it claimed.
func (r Repository) GetCrawlSourceByID(ctx context.Context, id uuid.UUID) (CrawlSource, error) {
	return scanCrawlSource(r.DB.QueryRow(ctx, `SELECT `+crawlSourceColumns+` FROM crawl_sources s WHERE s.id=$1`, id))
}

// CreateCrawlSource inserts a source unless the tenant is at its limit (ErrCrawlLimit) or already has the URL (ErrConflict).
// An interval source is due immediately so its first fetch happens without a manual click.
func (r Repository) CreateCrawlSource(ctx context.Context, s CrawlSource, actor uuid.UUID) (CrawlSource, error) {
	var id uuid.UUID
	err := r.DB.QueryRow(ctx, `INSERT INTO crawl_sources(tenant_id,url,kind,max_depth,max_pages,schedule_mode,interval_minutes,next_run_at,created_by)
		SELECT $1,$2,$3,$4,$5,$6,$7,CASE WHEN $6='INTERVAL' THEN now() END,$8
		WHERE (SELECT count(*) FROM crawl_sources WHERE tenant_id=$1) < $9
		RETURNING id`, s.TenantID, s.URL, s.Kind, s.MaxDepth, s.MaxPages, s.ScheduleMode, s.IntervalMinutes, actor, MaxCrawlSourcesPerTenant).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return s, ErrCrawlLimit
	}
	if isUniqueViolation(err) {
		return s, ErrConflict
	}
	if err != nil {
		return s, err
	}
	return r.GetCrawlSource(ctx, s.TenantID, id)
}

type CrawlSourcePatch struct {
	Enabled         *bool
	ScheduleMode    *string
	IntervalMinutes *int
}

func (r Repository) UpdateCrawlSource(ctx context.Context, tenantID, id uuid.UUID, p CrawlSourcePatch) (CrawlSource, error) {
	tag, err := r.DB.Exec(ctx, `UPDATE crawl_sources SET
		enabled=COALESCE($3,enabled),
		schedule_mode=COALESCE($4,schedule_mode),
		interval_minutes=CASE WHEN COALESCE($4,schedule_mode)='MANUAL' THEN NULL ELSE COALESCE($5,interval_minutes) END,
		updated_at=now()
		WHERE id=$1 AND tenant_id=$2`, id, tenantID, p.Enabled, p.ScheduleMode, p.IntervalMinutes)
	if err != nil {
		return CrawlSource{}, err
	}
	if tag.RowsAffected() == 0 {
		return CrawlSource{}, ErrNotFound
	}
	// A source that is (re)enabled on a schedule becomes due now; one that is paused or manual has no next run.
	if _, err = r.DB.Exec(ctx, `UPDATE crawl_sources SET next_run_at=CASE WHEN schedule_mode='INTERVAL' AND enabled THEN COALESCE(next_run_at,now()) END WHERE id=$1`, id); err != nil {
		return CrawlSource{}, err
	}
	return r.GetCrawlSource(ctx, tenantID, id)
}

// SetCrawlSourceForced lets a superadmin switch a source off regardless of what the pesantren admin does.
func (r Repository) SetCrawlSourceForced(ctx context.Context, id uuid.UUID, forced bool) error {
	tag, err := r.DB.Exec(ctx, `UPDATE crawl_sources SET forced_disabled=$2, updated_at=now() WHERE id=$1`, id, forced)
	if err == nil && tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return err
}

// StoredFile identifies a file whose bytes the caller must remove from disk after the row is soft-deleted.
type StoredFile struct {
	ID         uuid.UUID
	StorageKey string
}

// DeleteCrawlSource removes the source and soft-deletes the files it produced.
func (r Repository) DeleteCrawlSource(ctx context.Context, tenantID, id uuid.UUID) ([]StoredFile, error) {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	rows, err := tx.Query(ctx, `UPDATE files SET deleted_at=now(), updated_at=now()
		WHERE tenant_id=$1 AND deleted_at IS NULL AND id IN (SELECT file_id FROM crawled_pages WHERE source_id=$2 AND file_id IS NOT NULL)
		RETURNING id, storage_key`, tenantID, id)
	if err != nil {
		return nil, err
	}
	var files []StoredFile
	for rows.Next() {
		var f StoredFile
		if err := rows.Scan(&f.ID, &f.StorageKey); err != nil {
			rows.Close()
			return nil, err
		}
		files = append(files, f)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	tag, err := tx.Exec(ctx, `DELETE FROM crawl_sources WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, ErrNotFound
	}
	return files, tx.Commit(ctx)
}

// EnqueueCrawlRun queues a run. ErrConflict means one is already queued or running; ErrCrawlTooSoon rate-limits manual clicks.
func (r Repository) EnqueueCrawlRun(ctx context.Context, tenantID, sourceID uuid.UUID, trigger string, actor *uuid.UUID) (uuid.UUID, error) {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	defer tx.Rollback(ctx)
	var locked uuid.UUID
	if err = tx.QueryRow(ctx, `SELECT id FROM crawl_sources WHERE id=$1 AND tenant_id=$2 FOR UPDATE`, sourceID, tenantID).Scan(&locked); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, ErrNotFound
		}
		return uuid.Nil, err
	}
	var active bool
	if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM crawl_runs WHERE source_id=$1 AND status IN ('QUEUED','RUNNING'))`, sourceID).Scan(&active); err != nil {
		return uuid.Nil, err
	}
	if active {
		return uuid.Nil, ErrConflict
	}
	if trigger == "MANUAL" {
		var recent bool
		if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM crawl_runs WHERE source_id=$1 AND trigger='MANUAL' AND created_at > now() - $2::interval)`, sourceID, ManualCrawlCooldown.String()).Scan(&recent); err != nil {
			return uuid.Nil, err
		}
		if recent {
			return uuid.Nil, ErrCrawlTooSoon
		}
	}
	var runID uuid.UUID
	if err = tx.QueryRow(ctx, `INSERT INTO crawl_runs(tenant_id,source_id,trigger,triggered_by) VALUES($1,$2,$3,$4) RETURNING id`, tenantID, sourceID, trigger, actor).Scan(&runID); err != nil {
		if isUniqueViolation(err) {
			return uuid.Nil, ErrConflict
		}
		return uuid.Nil, err
	}
	if _, err = tx.Exec(ctx, `UPDATE crawl_sources SET last_status='QUEUED', error_public_message=NULL WHERE id=$1`, sourceID); err != nil {
		return uuid.Nil, err
	}
	return runID, tx.Commit(ctx)
}

// ScheduleDueCrawls queues a run for every enabled interval source whose time has come and which is not already running.
func (r Repository) ScheduleDueCrawls(ctx context.Context) (int, error) {
	tag, err := r.DB.Exec(ctx, `WITH due AS (
			SELECT s.id, s.tenant_id FROM crawl_sources s JOIN tenants t ON t.id=s.tenant_id
			WHERE s.enabled AND NOT s.forced_disabled AND s.schedule_mode='INTERVAL' AND s.next_run_at <= now()
			  AND t.lifecycle_status NOT IN ('SUSPENDED','ARCHIVED','DELETING','DELETED')
			  AND NOT EXISTS (SELECT 1 FROM crawl_runs r WHERE r.source_id=s.id AND r.status IN ('QUEUED','RUNNING'))
			FOR UPDATE OF s SKIP LOCKED
		), ins AS (
			INSERT INTO crawl_runs(tenant_id,source_id,trigger) SELECT tenant_id,id,'SCHEDULED' FROM due ON CONFLICT DO NOTHING RETURNING source_id
		)
		UPDATE crawl_sources SET last_status='QUEUED', error_public_message=NULL WHERE id IN (SELECT source_id FROM ins)`)
	if err != nil {
		return 0, err
	}
	return int(tag.RowsAffected()), nil
}

// ClaimCrawlRun takes the oldest queued run, or one stuck in RUNNING after a crash. ErrNotFound means the queue is empty.
func (r Repository) ClaimCrawlRun(ctx context.Context) (CrawlRun, error) {
	var run CrawlRun
	err := r.DB.QueryRow(ctx, `UPDATE crawl_runs SET status='RUNNING', started_at=now() WHERE id=(
			SELECT id FROM crawl_runs WHERE status='QUEUED' OR (status='RUNNING' AND started_at < now() - $1::interval)
			ORDER BY created_at FOR UPDATE SKIP LOCKED LIMIT 1)
		RETURNING id, tenant_id, source_id, trigger`, crawlStaleAfter.String()).Scan(&run.ID, &run.TenantID, &run.SourceID, &run.Trigger)
	if errors.Is(err, pgx.ErrNoRows) {
		return run, ErrNotFound
	}
	if err == nil {
		_, err = r.DB.Exec(ctx, `UPDATE crawl_sources SET last_status='RUNNING' WHERE id=$1`, run.SourceID)
	}
	return run, err
}

// FinishCrawlRun records the outcome and schedules the next run from now, so a slow run never causes a backlog.
func (r Repository) FinishCrawlRun(ctx context.Context, run CrawlRun, status string, c CrawlCounts, code, publicMessage string) error {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `UPDATE crawl_runs SET status=$2, pages_new=$3, pages_changed=$4, pages_unchanged=$5, pages_failed=$6, error_code=NULLIF($7,''), error_public_message=NULLIF($8,''), finished_at=now() WHERE id=$1`,
		run.ID, status, c.New, c.Changed, c.Unchanged, c.Failed, code, publicMessage); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `UPDATE crawl_sources SET last_status=$2, last_run_at=now(), error_public_message=NULLIF($3,''),
		next_run_at=CASE WHEN schedule_mode='INTERVAL' AND enabled THEN now() + make_interval(mins => interval_minutes) END, updated_at=now()
		WHERE id=$1`, run.SourceID, status, publicMessage); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func softDeleteFiles(ctx context.Context, tx pgx.Tx, tenantID uuid.UUID, ids []uuid.UUID) ([]StoredFile, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	rows, err := tx.Query(ctx, `UPDATE files SET deleted_at=now(), updated_at=now() WHERE tenant_id=$1 AND id=ANY($2) AND deleted_at IS NULL RETURNING id, storage_key`, tenantID, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []StoredFile
	for rows.Next() {
		var f StoredFile
		if err := rows.Scan(&f.ID, &f.StorageKey); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

// ReplaceFileContent points an existing file at newly stored bytes and queues it for the next sync to Hermes.
func (r Repository) ReplaceFileContent(ctx context.Context, tenantID, fileID uuid.UUID, storageKey, contentType string, size int64, sha string) error {
	tag, err := r.DB.Exec(ctx, `UPDATE files SET storage_key=$3, content_type=$4, size_bytes=$5, sha256=$6, processing_status='PENDING_SYNC', updated_at=now() WHERE id=$1 AND tenant_id=$2 AND deleted_at IS NULL`,
		fileID, tenantID, storageKey, contentType, size, sha)
	if err == nil && tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return err
}

// EnsureRootFolder returns the id of the tenant's root-level folder with this name, creating it when missing.
func (r Repository) EnsureRootFolder(ctx context.Context, tenantID uuid.UUID, name string) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.DB.QueryRow(ctx, `SELECT id FROM folders WHERE tenant_id=$1 AND parent_id IS NULL AND name=$2 ORDER BY created_at LIMIT 1`, tenantID, name).Scan(&id)
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, err
	}
	folder, err := r.CreateFolder(ctx, tenantID, nil, name)
	return folder.ID, err
}

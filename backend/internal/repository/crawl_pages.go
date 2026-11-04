package repository

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// CrawledPage is one page a source produced. Body is what the assistant reads; SourceText is what the web said last time.
type CrawledPage struct {
	ID            uuid.UUID
	SourceID      uuid.UUID
	URL           string
	Title         string
	Hash          string // hash of the source text Body is based on
	SourceHash    string // hash of the newest fetch
	FileID        *uuid.UUID
	Status        string
	Ext           string
	Body          string
	SourceText    string
	UserEdited    bool
	EditedAt      *time.Time
	LastSeenAt    *time.Time
	LastChangedAt *time.Time
}

// UpdateAvailable is true when the site changed after the pesantren edited the page, so a choice is waiting.
func (p CrawledPage) UpdateAvailable() bool {
	return p.UserEdited && p.SourceHash != "" && p.SourceHash != p.Hash
}

// MaxEditableBytes bounds what the dashboard will load into an editor.
const MaxEditableBytes = 512 << 10

// Editable pages are text or CSV with a stored file; PDFs and images can only be viewed as files.
func (p CrawledPage) Editable() bool {
	return p.Status == "ACTIVE" && p.FileID != nil && (p.Ext == ".txt" || p.Ext == ".csv") && len(p.Body) <= MaxEditableBytes
}

const crawledPageColumns = `id, source_id, url, coalesce(title,''), coalesce(content_sha256,''), coalesce(source_hash,''), file_id, status, ext,
  coalesce(body_text,''), coalesce(source_text,''), user_edited, edited_at, last_seen_at, last_changed_at`

func scanCrawledPage(row pgx.Row) (CrawledPage, error) {
	var p CrawledPage
	err := row.Scan(&p.ID, &p.SourceID, &p.URL, &p.Title, &p.Hash, &p.SourceHash, &p.FileID, &p.Status, &p.Ext, &p.Body, &p.SourceText, &p.UserEdited, &p.EditedAt, &p.LastSeenAt, &p.LastChangedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return p, ErrNotFound
	}
	return p, err
}

func (r Repository) GetCrawledPage(ctx context.Context, sourceID uuid.UUID, pageURL string) (CrawledPage, error) {
	return scanCrawledPage(r.DB.QueryRow(ctx, `SELECT `+crawledPageColumns+` FROM crawled_pages WHERE source_id=$1 AND url=$2`, sourceID, pageURL))
}

func (r Repository) GetCrawledPageByID(ctx context.Context, tenantID, pageID uuid.UUID) (CrawledPage, error) {
	return scanCrawledPage(r.DB.QueryRow(ctx, `SELECT `+crawledPageColumns+` FROM crawled_pages WHERE id=$1 AND tenant_id=$2`, pageID, tenantID))
}

// ListCrawledPages returns the pages of one source. Texts are left out; an oversized body is flagged by its length only.
func (r Repository) ListCrawledPages(ctx context.Context, tenantID, sourceID uuid.UUID) ([]CrawledPage, error) {
	rows, err := r.DB.Query(ctx, `SELECT id, source_id, url, coalesce(title,''), coalesce(content_sha256,''), coalesce(source_hash,''), file_id, status, ext,
		CASE WHEN length(coalesce(body_text,'')) <= $3 THEN '' ELSE repeat('x', $3 + 1) END, '', user_edited, edited_at, last_seen_at, last_changed_at
		FROM crawled_pages WHERE tenant_id=$1 AND source_id=$2 ORDER BY last_changed_at DESC NULLS LAST, url`, tenantID, sourceID, MaxEditableBytes)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []CrawledPage{}
	for rows.Next() {
		p, err := scanCrawledPage(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// UpsertCrawledPage stores the complete state of a page after a fetch. An excluded page stays excluded.
func (r Repository) UpsertCrawledPage(ctx context.Context, tenantID uuid.UUID, p CrawledPage, changed bool) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.DB.QueryRow(ctx, `INSERT INTO crawled_pages(tenant_id,source_id,url,title,content_sha256,source_hash,file_id,ext,body_text,source_text,user_edited,last_changed_at)
		VALUES($1,$2,$3,$4,$5,$5,$6,$7,$8,$9,$10,CASE WHEN $11 THEN now() END)
		ON CONFLICT (source_id,url) DO UPDATE SET title=EXCLUDED.title, content_sha256=EXCLUDED.content_sha256, source_hash=EXCLUDED.source_hash,
			file_id=EXCLUDED.file_id, ext=EXCLUDED.ext, body_text=EXCLUDED.body_text, source_text=EXCLUDED.source_text, user_edited=EXCLUDED.user_edited,
			edited_at=CASE WHEN EXCLUDED.user_edited THEN crawled_pages.edited_at END,
			status=CASE WHEN crawled_pages.status='EXCLUDED' THEN 'EXCLUDED' ELSE 'ACTIVE' END, miss_count=0, last_seen_at=now(),
			last_changed_at=CASE WHEN $11 THEN now() ELSE crawled_pages.last_changed_at END
		RETURNING id`, tenantID, p.SourceID, p.URL, p.Title, p.Hash, p.FileID, p.Ext, p.Body, p.SourceText, p.UserEdited, changed).Scan(&id)
	return id, err
}

// TouchCrawledPage marks a page as seen again without touching what the assistant reads. sourceText, when given, is the
// newest fetch of a page whose edited body must be left alone.
func (r Repository) TouchCrawledPage(ctx context.Context, pageID uuid.UUID, sourceHash string, sourceText *string) error {
	_, err := r.DB.Exec(ctx, `UPDATE crawled_pages SET last_seen_at=now(), miss_count=0, source_hash=$2, source_text=COALESCE($3,source_text) WHERE id=$1`, pageID, sourceHash, sourceText)
	return err
}

// UpdatePageContent applies an edit (userEdited) or a switch back to the site's text (ackSource), updating what the assistant reads.
func (r Repository) UpdatePageContent(ctx context.Context, tenantID, pageID uuid.UUID, body string, userEdited, ackSource bool, actor *uuid.UUID) error {
	tag, err := r.DB.Exec(ctx, `UPDATE crawled_pages SET body_text=$3, user_edited=$4,
		content_sha256=CASE WHEN $5 THEN source_hash ELSE content_sha256 END,
		edited_at=CASE WHEN $4 THEN now() END, edited_by=CASE WHEN $4 THEN $6::uuid END, last_changed_at=now()
		WHERE id=$1 AND tenant_id=$2`, pageID, tenantID, body, userEdited, ackSource, actor)
	if err == nil && tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return err
}

// AcknowledgeSourceChange keeps the pesantren's edited text and dismisses the "site changed" notice for the newest fetch.
func (r Repository) AcknowledgeSourceChange(ctx context.Context, tenantID, pageID uuid.UUID) error {
	tag, err := r.DB.Exec(ctx, `UPDATE crawled_pages SET content_sha256=source_hash WHERE id=$1 AND tenant_id=$2`, pageID, tenantID)
	if err == nil && tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return err
}

// ExcludeCrawledPage retires a page by hand: its file is soft-deleted so the assistant forgets it, and later crawls leave it out.
func (r Repository) ExcludeCrawledPage(ctx context.Context, tenantID, pageID uuid.UUID) ([]StoredFile, error) {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	var fileID *uuid.UUID
	if err = tx.QueryRow(ctx, `SELECT file_id FROM crawled_pages WHERE id=$1 AND tenant_id=$2 FOR UPDATE`, pageID, tenantID).Scan(&fileID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if _, err = tx.Exec(ctx, `UPDATE crawled_pages SET status='EXCLUDED', file_id=NULL WHERE id=$1`, pageID); err != nil {
		return nil, err
	}
	var files []StoredFile
	if fileID != nil {
		if files, err = softDeleteFiles(ctx, tx, tenantID, []uuid.UUID{*fileID}); err != nil {
			return nil, err
		}
	}
	return files, tx.Commit(ctx)
}

// IncludeCrawledPage puts an excluded page back. fileID is a freshly stored file, or nil when the next crawl must recreate it.
func (r Repository) IncludeCrawledPage(ctx context.Context, tenantID, pageID uuid.UUID, fileID *uuid.UUID) error {
	tag, err := r.DB.Exec(ctx, `UPDATE crawled_pages SET status='ACTIVE', file_id=$3, content_sha256=CASE WHEN $3::uuid IS NULL THEN NULL ELSE content_sha256 END, miss_count=0 WHERE id=$1 AND tenant_id=$2`, pageID, tenantID, fileID)
	if err == nil && tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return err
}

// MarkUnseenCrawledPages counts a miss against every active page this run did not see and retires those missed twice in
// a row, returning their files for removal and recording a GONE event for each. Call it only after a complete run.
func (r Repository) MarkUnseenCrawledPages(ctx context.Context, tenantID, sourceID, runID uuid.UUID, runStart time.Time) ([]StoredFile, error) {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	rows, err := tx.Query(ctx, `UPDATE crawled_pages SET miss_count=miss_count+1, status=CASE WHEN miss_count+1 >= 2 THEN 'GONE' ELSE status END
		WHERE source_id=$1 AND status='ACTIVE' AND last_seen_at < $2
		RETURNING id, status, file_id, url, coalesce(title,''), coalesce(body_text,'')`, sourceID, runStart)
	if err != nil {
		return nil, err
	}
	type gonePage struct {
		id               uuid.UUID
		url, title, body string
	}
	var fileIDs []uuid.UUID
	var gone []gonePage
	for rows.Next() {
		var g gonePage
		var status string
		var fileID *uuid.UUID
		if err := rows.Scan(&g.id, &status, &fileID, &g.url, &g.title, &g.body); err != nil {
			rows.Close()
			return nil, err
		}
		if status == "GONE" {
			gone = append(gone, g)
			if fileID != nil {
				fileIDs = append(fileIDs, *fileID)
			}
		}
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for _, g := range gone {
		removed := 0
		if strings.TrimSpace(g.body) != "" {
			removed = len(strings.Split(strings.TrimSpace(g.body), "\n"))
		}
		if _, err := tx.Exec(ctx, `INSERT INTO crawl_page_events(tenant_id,source_id,run_id,page_id,url,title,event,lines_removed,before_text) VALUES($1,$2,$3,$4,$5,$6,'GONE',$7,NULLIF($8,''))`,
			tenantID, sourceID, runID, g.id, g.url, g.title, removed, capText(g.body)); err != nil {
			return nil, err
		}
	}
	files, err := softDeleteFiles(ctx, tx, tenantID, fileIDs)
	if err != nil {
		return nil, err
	}
	return files, tx.Commit(ctx)
}

const maxEventText = 100 << 10

func capText(s string) string {
	if len(s) > maxEventText {
		s = s[:maxEventText]
		for !utf8.ValidString(s) {
			s = s[:len(s)-1]
		}
	}
	return s
}

// CrawlEvent records one visible change from a run: a new page, a changed page, a site change waiting on an edited page,
// or a removed page.
type CrawlEvent struct {
	ID           uuid.UUID  `json:"id"`
	RunID        uuid.UUID  `json:"run_id"`
	PageID       *uuid.UUID `json:"page_id,omitempty"`
	URL          string     `json:"url"`
	Title        string     `json:"title"`
	Event        string     `json:"event"`
	LinesAdded   int        `json:"lines_added"`
	LinesRemoved int        `json:"lines_removed"`
	BeforeText   string     `json:"before_text,omitempty"`
	AfterText    string     `json:"after_text,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

func (r Repository) RecordCrawlEvent(ctx context.Context, tenantID, sourceID uuid.UUID, e CrawlEvent) error {
	_, err := r.DB.Exec(ctx, `INSERT INTO crawl_page_events(tenant_id,source_id,run_id,page_id,url,title,event,lines_added,lines_removed,before_text,after_text) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,NULLIF($10,''),NULLIF($11,''))`,
		tenantID, sourceID, e.RunID, e.PageID, e.URL, e.Title, e.Event, e.LinesAdded, e.LinesRemoved, capText(e.BeforeText), capText(e.AfterText))
	return err
}

// ListCrawlEvents is the "what changed" feed, newest first, without the texts.
func (r Repository) ListCrawlEvents(ctx context.Context, tenantID, sourceID uuid.UUID, limit int) ([]CrawlEvent, error) {
	rows, err := r.DB.Query(ctx, `SELECT id, coalesce(run_id,'00000000-0000-0000-0000-000000000000'), page_id, url, coalesce(title,''), event, lines_added, lines_removed, created_at
		FROM crawl_page_events WHERE tenant_id=$1 AND source_id=$2 ORDER BY created_at DESC LIMIT $3`, tenantID, sourceID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []CrawlEvent{}
	for rows.Next() {
		var e CrawlEvent
		if err := rows.Scan(&e.ID, &e.RunID, &e.PageID, &e.URL, &e.Title, &e.Event, &e.LinesAdded, &e.LinesRemoved, &e.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// GetCrawlEvent returns one event with the before/after texts for the comparison view.
func (r Repository) GetCrawlEvent(ctx context.Context, tenantID, eventID uuid.UUID) (CrawlEvent, error) {
	var e CrawlEvent
	err := r.DB.QueryRow(ctx, `SELECT id, coalesce(run_id,'00000000-0000-0000-0000-000000000000'), page_id, url, coalesce(title,''), event, lines_added, lines_removed, coalesce(before_text,''), coalesce(after_text,''), created_at
		FROM crawl_page_events WHERE id=$1 AND tenant_id=$2`, eventID, tenantID).Scan(&e.ID, &e.RunID, &e.PageID, &e.URL, &e.Title, &e.Event, &e.LinesAdded, &e.LinesRemoved, &e.BeforeText, &e.AfterText, &e.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return e, ErrNotFound
	}
	return e, err
}

// CrawlRunInfo is a finished or active run as shown in the history.
type CrawlRunInfo struct {
	ID                 uuid.UUID  `json:"id"`
	Trigger            string     `json:"trigger"`
	Status             string     `json:"status"`
	PagesNew           int        `json:"pages_new"`
	PagesChanged       int        `json:"pages_changed"`
	PagesUnchanged     int        `json:"pages_unchanged"`
	PagesFailed        int        `json:"pages_failed"`
	ErrorPublicMessage string     `json:"error_public_message,omitempty"`
	StartedAt          *time.Time `json:"started_at,omitempty"`
	FinishedAt         *time.Time `json:"finished_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
}

func (r Repository) ListCrawlRuns(ctx context.Context, tenantID, sourceID uuid.UUID, limit int) ([]CrawlRunInfo, error) {
	rows, err := r.DB.Query(ctx, `SELECT id, trigger, status, pages_new, pages_changed, pages_unchanged, pages_failed, coalesce(error_public_message,''), started_at, finished_at, created_at
		FROM crawl_runs WHERE tenant_id=$1 AND source_id=$2 ORDER BY created_at DESC LIMIT $3`, tenantID, sourceID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []CrawlRunInfo{}
	for rows.Next() {
		var x CrawlRunInfo
		if err := rows.Scan(&x.ID, &x.Trigger, &x.Status, &x.PagesNew, &x.PagesChanged, &x.PagesUnchanged, &x.PagesFailed, &x.ErrorPublicMessage, &x.StartedAt, &x.FinishedAt, &x.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

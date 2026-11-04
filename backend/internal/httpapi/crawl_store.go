package httpapi

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"ponpes-saas/internal/crawler"
	"ponpes-saas/internal/repository"
)

// storeCrawledPage records one fetched page and returns "new", "changed" or "unchanged".
//
// The rule that protects a pesantren's corrections: a page they edited is never overwritten by a crawl. The newest text
// from the site is only remembered (source_text), and a SOURCE_CHANGED event tells them to choose which version to keep.
func (s *Server) storeCrawledPage(ctx context.Context, run repository.CrawlRun, source repository.CrawlSource, folderID uuid.UUID, page crawler.Page, now time.Time) (string, error) {
	existing, lookupErr := s.repo.GetCrawledPage(ctx, run.SourceID, page.URL)
	if lookupErr != nil && !errors.Is(lookupErr, repository.ErrNotFound) {
		return "", lookupErr
	}
	known := lookupErr == nil
	isText := !page.Binary
	body := string(page.Body)

	if known && existing.Status == "EXCLUDED" {
		return "unchanged", s.repo.TouchCrawledPage(ctx, existing.ID, page.Hash, nil)
	}
	var currentFile *repository.File
	if known && existing.FileID != nil {
		// The pesantren may have deleted the file by hand; then the page is recreated below.
		if f, err := s.repo.FindFile(ctx, run.TenantID, *existing.FileID); err == nil {
			currentFile = &f
		}
	}

	if currentFile != nil && existing.UserEdited && isText {
		if page.Hash == existing.SourceHash {
			return "unchanged", s.repo.TouchCrawledPage(ctx, existing.ID, page.Hash, nil)
		}
		if err := s.repo.TouchCrawledPage(ctx, existing.ID, page.Hash, &body); err != nil {
			return "", err
		}
		added, removed := crawler.DiffLines(existing.SourceText, body)
		id := existing.ID
		err := s.repo.RecordCrawlEvent(ctx, run.TenantID, run.SourceID, repository.CrawlEvent{RunID: run.ID, PageID: &id, URL: page.URL, Title: page.Title, Event: "SOURCE_CHANGED",
			LinesAdded: added, LinesRemoved: removed, BeforeText: existing.SourceText, AfterText: body})
		return "changed", err
	}
	if currentFile != nil && existing.Hash == page.Hash {
		return "unchanged", s.repo.TouchCrawledPage(ctx, existing.ID, page.Hash, nil)
	}

	name := crawlFilename(page)
	key, size, sum, err := s.store.Put(run.TenantID.String(), name, bytes.NewReader(page.Document(now)), crawlMaxStore)
	if err != nil {
		return "", err
	}
	record := repository.CrawledPage{SourceID: run.SourceID, URL: page.URL, Title: page.Title, Hash: page.Hash, Ext: page.Ext}
	if isText {
		record.Body, record.SourceText = body, body
	}
	outcome, before := "new", ""
	if currentFile != nil {
		if err := s.repo.ReplaceFileContent(ctx, run.TenantID, currentFile.ID, key, crawlContentType(page.Ext), size, sum); err != nil {
			_ = s.store.Delete(run.TenantID.String(), key)
			return "", err
		}
		_ = s.store.Delete(run.TenantID.String(), currentFile.StorageKey)
		record.FileID = &currentFile.ID
		outcome, before = "changed", existing.Body
	} else {
		file := repository.File{ID: uuid.New(), TenantID: run.TenantID, FolderID: &folderID, Filename: name, StorageKey: key, ContentType: crawlContentType(page.Ext), Size: size, SHA256: sum, Status: "PENDING_SYNC"}
		if err := s.repo.InsertFile(ctx, file); err != nil {
			_ = s.store.Delete(run.TenantID.String(), key)
			return "", err
		}
		record.FileID = &file.ID
	}
	pageID, err := s.repo.UpsertCrawledPage(ctx, run.TenantID, record, true)
	if err != nil {
		return "", err
	}
	added, removed := crawler.DiffLines(before, body)
	event := "NEW"
	if outcome == "changed" {
		event = "CHANGED"
	}
	return outcome, s.repo.RecordCrawlEvent(ctx, run.TenantID, run.SourceID, repository.CrawlEvent{RunID: run.ID, PageID: &pageID, URL: page.URL, Title: page.Title, Event: event,
		LinesAdded: added, LinesRemoved: removed, BeforeText: before, AfterText: body})
}

// writeCrawlFile stores a text page's current body as a new file version and returns where it went.
func (s *Server) writeCrawlFile(tenantID uuid.UUID, page repository.CrawledPage, body string, editedAt *time.Time) (key string, size int64, sum string, err error) {
	fetched := time.Now()
	if page.LastSeenAt != nil {
		fetched = *page.LastSeenAt
	}
	doc := crawler.RenderDocument(page.URL, page.Title, body, fetched, editedAt)
	return s.store.Put(tenantID.String(), "page"+page.Ext, bytes.NewReader(doc), crawlMaxStore)
}

// validPageText rejects empty, oversized or binary-looking text before it is stored and read by the assistant.
func validPageText(text string) bool {
	trimmed := strings.TrimSpace(text)
	return trimmed != "" && len(text) <= repository.MaxEditableBytes && !strings.ContainsRune(text, 0)
}

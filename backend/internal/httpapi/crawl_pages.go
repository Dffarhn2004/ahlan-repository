package httpapi

import (
	"errors"
	"net/http"
	"time"
	"unicode/utf8"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"ponpes-saas/internal/repository"
)

func pageView(p repository.CrawledPage, withText bool) fiber.Map {
	view := fiber.Map{
		"id": p.ID, "url": p.URL, "title": p.Title, "status": p.Status, "ext": p.Ext,
		"editable": p.Editable(), "user_edited": p.UserEdited, "update_available": p.UpdateAvailable(),
		"edited_at": p.EditedAt, "last_seen_at": p.LastSeenAt, "last_changed_at": p.LastChangedAt,
	}
	if withText {
		view["body"] = p.Body
		if p.UpdateAvailable() {
			view["source_text"] = p.SourceText
		}
	}
	return view
}

func tenantAndParam(c *fiber.Ctx, param, notFound string) (uuid.UUID, uuid.UUID, error) {
	tenantID, err := tenantIDParam(c)
	if err != nil {
		return uuid.Nil, uuid.Nil, fiber.NewError(http.StatusBadRequest, "INVALID_TENANT_ID")
	}
	id, err := uuid.Parse(c.Params(param))
	if err != nil {
		return uuid.Nil, uuid.Nil, fiber.NewError(http.StatusNotFound, notFound)
	}
	return tenantID, id, nil
}

func (s *Server) listCrawlRuns(c *fiber.Ctx) error {
	tenantID, sourceID, err := crawlSourceID(c)
	if err != nil {
		return err
	}
	runs, err := s.repo.ListCrawlRuns(c.Context(), tenantID, sourceID, 30)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"data": runs})
}

func (s *Server) listCrawlEvents(c *fiber.Ctx) error {
	tenantID, sourceID, err := crawlSourceID(c)
	if err != nil {
		return err
	}
	events, err := s.repo.ListCrawlEvents(c.Context(), tenantID, sourceID, 100)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"data": events})
}

func (s *Server) getCrawlEvent(c *fiber.Ctx) error {
	tenantID, eventID, err := tenantAndParam(c, "eid", "CRAWL_EVENT_NOT_FOUND")
	if err != nil {
		return err
	}
	event, err := s.repo.GetCrawlEvent(c.Context(), tenantID, eventID)
	if err != nil {
		return fiber.NewError(http.StatusNotFound, "CRAWL_EVENT_NOT_FOUND")
	}
	return c.JSON(event)
}

func (s *Server) listCrawlPages(c *fiber.Ctx) error {
	tenantID, sourceID, err := crawlSourceID(c)
	if err != nil {
		return err
	}
	pages, err := s.repo.ListCrawledPages(c.Context(), tenantID, sourceID)
	if err != nil {
		return err
	}
	out := make([]fiber.Map, 0, len(pages))
	for _, p := range pages {
		out = append(out, pageView(p, false))
	}
	return c.JSON(fiber.Map{"data": out})
}

func (s *Server) getCrawlPage(c *fiber.Ctx) error {
	tenantID, pageID, err := tenantAndParam(c, "pid", "CRAWL_PAGE_NOT_FOUND")
	if err != nil {
		return err
	}
	page, err := s.repo.GetCrawledPageByID(c.Context(), tenantID, pageID)
	if err != nil {
		return fiber.NewError(http.StatusNotFound, "CRAWL_PAGE_NOT_FOUND")
	}
	return c.JSON(pageView(page, true))
}

// saveCrawlPageContent lets the pesantren correct what the assistant reads for one page.
func (s *Server) saveCrawlPageContent(c *fiber.Ctx) error {
	tenantID, pageID, err := tenantAndParam(c, "pid", "CRAWL_PAGE_NOT_FOUND")
	if err != nil {
		return err
	}
	var in struct {
		Text string `json:"text"`
	}
	if err := c.BodyParser(&in); err != nil || !utf8.ValidString(in.Text) || !validPageText(in.Text) {
		return fiber.NewError(http.StatusBadRequest, "INVALID_REQUEST")
	}
	page, err := s.repo.GetCrawledPageByID(c.Context(), tenantID, pageID)
	if err != nil {
		return fiber.NewError(http.StatusNotFound, "CRAWL_PAGE_NOT_FOUND")
	}
	if !page.Editable() {
		return fiber.NewError(http.StatusUnprocessableEntity, "CRAWL_PAGE_NOT_EDITABLE")
	}
	file, err := s.repo.FindFile(c.Context(), tenantID, *page.FileID)
	if err != nil {
		return fiber.NewError(http.StatusConflict, "CRAWL_PAGE_FILE_MISSING")
	}
	now := time.Now()
	key, size, sum, err := s.writeCrawlFile(tenantID, page, in.Text, &now)
	if err != nil {
		return err
	}
	if err := s.repo.ReplaceFileContent(c.Context(), tenantID, file.ID, key, crawlContentType(page.Ext), size, sum); err != nil {
		_ = s.store.Delete(tenantID.String(), key)
		return err
	}
	_ = s.store.Delete(tenantID.String(), file.StorageKey)
	actor := auditActor(c)
	if err := s.repo.UpdatePageContent(c.Context(), tenantID, pageID, in.Text, true, false, actor); err != nil {
		return err
	}
	s.repo.RecordAudit(c.Context(), tenantID, actor, "CRAWL_PAGE_EDITED", "CRAWL_PAGE", pageID.String(), page.URL)
	go s.syncAfterCrawl(tenantID)
	updated, err := s.repo.GetCrawledPageByID(c.Context(), tenantID, pageID)
	if err != nil {
		return err
	}
	return c.JSON(pageView(updated, true))
}

// resolveCrawlPage answers the "the site changed after you edited this" notice, or reverts an edit.
func (s *Server) resolveCrawlPage(c *fiber.Ctx) error {
	tenantID, pageID, err := tenantAndParam(c, "pid", "CRAWL_PAGE_NOT_FOUND")
	if err != nil {
		return err
	}
	var in struct {
		Action string `json:"action"`
	}
	if err := c.BodyParser(&in); err != nil || (in.Action != "use_source" && in.Action != "keep_mine") {
		return fiber.NewError(http.StatusBadRequest, "INVALID_REQUEST")
	}
	page, err := s.repo.GetCrawledPageByID(c.Context(), tenantID, pageID)
	if err != nil {
		return fiber.NewError(http.StatusNotFound, "CRAWL_PAGE_NOT_FOUND")
	}
	if !page.UserEdited || !page.Editable() {
		return fiber.NewError(http.StatusUnprocessableEntity, "CRAWL_PAGE_NOT_EDITABLE")
	}
	actor := auditActor(c)
	if in.Action == "keep_mine" {
		if err := s.repo.AcknowledgeSourceChange(c.Context(), tenantID, pageID); err != nil {
			return err
		}
		s.repo.RecordAudit(c.Context(), tenantID, actor, "CRAWL_PAGE_KEPT", "CRAWL_PAGE", pageID.String(), page.URL)
	} else {
		if page.SourceText == "" {
			return fiber.NewError(http.StatusUnprocessableEntity, "CRAWL_PAGE_NO_SOURCE")
		}
		file, err := s.repo.FindFile(c.Context(), tenantID, *page.FileID)
		if err != nil {
			return fiber.NewError(http.StatusConflict, "CRAWL_PAGE_FILE_MISSING")
		}
		key, size, sum, err := s.writeCrawlFile(tenantID, page, page.SourceText, nil)
		if err != nil {
			return err
		}
		if err := s.repo.ReplaceFileContent(c.Context(), tenantID, file.ID, key, crawlContentType(page.Ext), size, sum); err != nil {
			_ = s.store.Delete(tenantID.String(), key)
			return err
		}
		_ = s.store.Delete(tenantID.String(), file.StorageKey)
		if err := s.repo.UpdatePageContent(c.Context(), tenantID, pageID, page.SourceText, false, true, nil); err != nil {
			return err
		}
		s.repo.RecordAudit(c.Context(), tenantID, actor, "CRAWL_PAGE_REVERTED", "CRAWL_PAGE", pageID.String(), page.URL)
		go s.syncAfterCrawl(tenantID)
	}
	updated, err := s.repo.GetCrawledPageByID(c.Context(), tenantID, pageID)
	if err != nil {
		return err
	}
	return c.JSON(pageView(updated, true))
}

// patchCrawlPage excludes a page from, or returns it to, what the assistant reads.
func (s *Server) patchCrawlPage(c *fiber.Ctx) error {
	tenantID, pageID, err := tenantAndParam(c, "pid", "CRAWL_PAGE_NOT_FOUND")
	if err != nil {
		return err
	}
	var in struct {
		Excluded *bool `json:"excluded"`
	}
	if err := c.BodyParser(&in); err != nil || in.Excluded == nil {
		return fiber.NewError(http.StatusBadRequest, "INVALID_REQUEST")
	}
	page, err := s.repo.GetCrawledPageByID(c.Context(), tenantID, pageID)
	if err != nil {
		return fiber.NewError(http.StatusNotFound, "CRAWL_PAGE_NOT_FOUND")
	}
	actor := auditActor(c)
	if *in.Excluded {
		files, err := s.repo.ExcludeCrawledPage(c.Context(), tenantID, pageID)
		if err != nil {
			return err
		}
		for _, f := range files {
			_ = s.store.Delete(tenantID.String(), f.StorageKey)
		}
		s.repo.RecordAudit(c.Context(), tenantID, actor, "CRAWL_PAGE_EXCLUDED", "CRAWL_PAGE", pageID.String(), page.URL)
		go s.syncAfterCrawl(tenantID)
	} else {
		if page.Status != "EXCLUDED" {
			return fiber.NewError(http.StatusUnprocessableEntity, "CRAWL_PAGE_NOT_EXCLUDED")
		}
		var fileID *uuid.UUID
		if (page.Ext == ".txt" || page.Ext == ".csv") && page.Body != "" {
			var editedAt *time.Time
			if page.UserEdited {
				editedAt = page.EditedAt
			}
			key, size, sum, err := s.writeCrawlFile(tenantID, page, page.Body, editedAt)
			if err != nil {
				return err
			}
			folderID, err := s.repo.EnsureRootFolder(c.Context(), tenantID, crawlFolderName)
			if err != nil {
				_ = s.store.Delete(tenantID.String(), key)
				return err
			}
			file := repository.File{ID: uuid.New(), TenantID: tenantID, FolderID: &folderID, Filename: crawlFilenameFor(page), StorageKey: key, ContentType: crawlContentType(page.Ext), Size: size, SHA256: sum, Status: "PENDING_SYNC"}
			if err := s.repo.InsertFile(c.Context(), file); err != nil {
				_ = s.store.Delete(tenantID.String(), key)
				return err
			}
			fileID = &file.ID
		}
		if err := s.repo.IncludeCrawledPage(c.Context(), tenantID, pageID, fileID); err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return fiber.NewError(http.StatusNotFound, "CRAWL_PAGE_NOT_FOUND")
			}
			return err
		}
		s.repo.RecordAudit(c.Context(), tenantID, actor, "CRAWL_PAGE_INCLUDED", "CRAWL_PAGE", pageID.String(), page.URL)
		if fileID != nil {
			go s.syncAfterCrawl(tenantID)
		}
	}
	updated, err := s.repo.GetCrawledPageByID(c.Context(), tenantID, pageID)
	if err != nil {
		return err
	}
	return c.JSON(pageView(updated, false))
}

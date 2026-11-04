package httpapi

import (
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"ponpes-saas/internal/repository"
)

const maxUploadBytes = 20 << 20

// tenantAndFile reads :id and :file_id, answering 404 when either is malformed.
func tenantAndFile(c *fiber.Ctx) (uuid.UUID, uuid.UUID, error) {
	tenantID, tenantErr := tenantIDParam(c)
	fileID, fileErr := uuid.Parse(c.Params("file_id"))
	if tenantErr != nil || fileErr != nil {
		return uuid.Nil, uuid.Nil, fiber.NewError(http.StatusNotFound, "FILE_NOT_FOUND")
	}
	return tenantID, fileID, nil
}

func (s *Server) listFiles(c *fiber.Ctx) error {
	id, err := tenantIDParam(c)
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "INVALID_TENANT_ID")
	}
	files, err := s.repo.ListFiles(c.Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"data": files})
}

func (s *Server) uploadFile(c *fiber.Ctx) error {
	id, err := tenantIDParam(c)
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "INVALID_TENANT_ID")
	}
	header, err := c.FormFile("file")
	if err != nil || header.Size > maxUploadBytes {
		return fiber.NewError(http.StatusBadRequest, "INVALID_FILE")
	}
	src, err := header.Open()
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "FILE_READ_FAILED")
	}
	defer src.Close()
	key, size, sum, err := s.store.Put(id.String(), header.Filename, src, maxUploadBytes)
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "INVALID_FILE")
	}
	// Files can be uploaded before Hermes is approved; the worker can sync them later.
	var folderID *uuid.UUID
	if value := c.FormValue("folder_id"); value != "" {
		parsed, parseErr := uuid.Parse(value)
		if parseErr != nil {
			return fiber.NewError(http.StatusBadRequest, "INVALID_FOLDER_ID")
		}
		folderID = &parsed
	}
	item := repository.File{ID: uuid.New(), TenantID: id, FolderID: folderID, Filename: header.Filename, StorageKey: key, ContentType: header.Header.Get("Content-Type"), Size: size, SHA256: sum, Status: "PENDING_SYNC"}
	if err := s.repo.InsertFile(c.Context(), item); err != nil {
		_ = s.store.Delete(id.String(), key)
		return err
	}
	s.repo.RecordAudit(c.Context(), id, auditActor(c), "FILE_UPLOADED", "FILE", item.ID.String(), item.Filename)
	return c.Status(http.StatusCreated).JSON(item)
}

func (s *Server) downloadFile(c *fiber.Ctx) error {
	id, fileID, err := tenantAndFile(c)
	if err != nil {
		return err
	}
	item, err := s.repo.FindFile(c.Context(), id, fileID)
	if err != nil {
		return fiber.NewError(http.StatusNotFound, "FILE_NOT_FOUND")
	}
	path, err := s.store.Path(id.String(), item.StorageKey)
	if err != nil {
		return fiber.NewError(http.StatusNotFound, "FILE_NOT_FOUND")
	}
	return c.SendFile(path)
}

func (s *Server) renameFile(c *fiber.Ctx) error {
	id, fileID, err := tenantAndFile(c)
	if err != nil {
		return err
	}
	var input struct {
		Filename string `json:"filename"`
	}
	if err = c.BodyParser(&input); err != nil || strings.TrimSpace(input.Filename) == "" {
		return fiber.NewError(http.StatusBadRequest, "INVALID_FILENAME")
	}
	if err = s.repo.RenameFile(c.Context(), id, fileID, strings.TrimSpace(input.Filename)); err != nil {
		return err
	}
	item, err := s.repo.FindFile(c.Context(), id, fileID)
	if err != nil {
		return fiber.NewError(http.StatusNotFound, "FILE_NOT_FOUND")
	}
	s.repo.RecordAudit(c.Context(), id, auditActor(c), "FILE_RENAMED", "FILE", fileID.String(), item.Filename)
	return c.JSON(item)
}

func (s *Server) deleteFile(c *fiber.Ctx) error {
	id, fileID, err := tenantAndFile(c)
	if err != nil {
		return err
	}
	item, err := s.repo.FindFile(c.Context(), id, fileID)
	if err != nil {
		return fiber.NewError(http.StatusNotFound, "FILE_NOT_FOUND")
	}
	if err = s.repo.DeleteFile(c.Context(), id, fileID); err != nil {
		return err
	}
	_ = s.store.Delete(id.String(), item.StorageKey)
	s.repo.RecordAudit(c.Context(), id, auditActor(c), "FILE_DELETED", "FILE", fileID.String(), item.Filename)
	return c.SendStatus(http.StatusNoContent)
}

func (s *Server) listFolders(c *fiber.Ctx) error {
	id, err := tenantIDParam(c)
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "INVALID_TENANT_ID")
	}
	var parent *uuid.UUID
	if value := c.Query("parent_id"); value != "" {
		parsed, parseErr := uuid.Parse(value)
		if parseErr != nil {
			return fiber.NewError(http.StatusBadRequest, "INVALID_FOLDER_ID")
		}
		parent = &parsed
	}
	folders, err := s.repo.ListFolders(c.Context(), id, parent)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"data": folders})
}

func (s *Server) createFolder(c *fiber.Ctx) error {
	id, err := tenantIDParam(c)
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "INVALID_TENANT_ID")
	}
	var input struct {
		Name     string `json:"name"`
		ParentID string `json:"parent_id"`
	}
	if err = c.BodyParser(&input); err != nil || strings.TrimSpace(input.Name) == "" {
		return fiber.NewError(http.StatusBadRequest, "INVALID_FOLDER")
	}
	var parent *uuid.UUID
	if input.ParentID != "" {
		parsed, parseErr := uuid.Parse(input.ParentID)
		if parseErr != nil {
			return fiber.NewError(http.StatusBadRequest, "INVALID_FOLDER_ID")
		}
		parent = &parsed
	}
	folder, err := s.repo.CreateFolder(c.Context(), id, parent, strings.TrimSpace(input.Name))
	if err != nil {
		return err
	}
	s.repo.RecordAudit(c.Context(), id, auditActor(c), "FOLDER_CREATED", "FOLDER", folder.ID.String(), folder.Name)
	return c.Status(http.StatusCreated).JSON(folder)
}

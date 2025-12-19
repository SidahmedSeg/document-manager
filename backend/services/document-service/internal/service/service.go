package service

import (
	"context"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/SidahmedSeg/document-manager/backend/pkg/cache"
	"github.com/SidahmedSeg/document-manager/backend/pkg/errors"
	"github.com/SidahmedSeg/document-manager/backend/pkg/logger"
	"github.com/SidahmedSeg/document-manager/backend/pkg/middleware"
	"github.com/SidahmedSeg/document-manager/backend/services/document-service/internal/models"
	"github.com/SidahmedSeg/document-manager/backend/services/document-service/internal/repository"
	"go.uber.org/zap"
)

const (
	documentCacheTTL = 30 * time.Minute
	folderCacheTTL   = 1 * time.Hour
)

// Service handles document business logic
type Service struct {
	repo   *repository.Repository
	cache  *cache.Cache
	logger *zap.Logger
}

// NewService creates a new document service
func NewService(repo *repository.Repository, cache *cache.Cache, logger *zap.Logger) *Service {
	return &Service{
		repo:   repo,
		cache:  cache,
		logger: logger,
	}
}

// Document operations

// CreateDocument creates a new document (metadata only, file upload handled separately)
func (s *Service) CreateDocument(ctx context.Context, req *models.CreateDocumentRequest, fileInfo FileInfo) (*models.Document, error) {
	tenantID := getTenantID(ctx)
	userID := middleware.GetUserID(ctx)

	// Validate folder ownership if provided
	if req.FolderID != "" {
		folderUUID, _ := uuid.Parse(req.FolderID)
		if _, err := s.repo.GetFolder(ctx, tenantID, folderUUID); err != nil {
			return nil, errors.Validationf("invalid folder_id")
		}
	}

	// Validate category ownership if provided
	if req.CategoryID != "" {
		// TODO: Validate category exists and belongs to tenant
	}

	// Create document
	doc := &models.Document{
		ID:            uuid.New(),
		TenantID:      tenantID,
		Name:          req.Name,
		FileType:      fileInfo.Extension,
		FileSize:      fileInfo.Size,
		MimeType:      fileInfo.MimeType,
		StoragePath:   fileInfo.StoragePath,
		Status:        "active",
		UploadedBy:    userID,
		OCRStatus:     "pending",
		Version:       1,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if req.Description != "" {
		doc.Description.String = req.Description
		doc.Description.Valid = true
	}

	if req.FolderID != "" {
		doc.FolderID.String = req.FolderID
		doc.FolderID.Valid = true
	}

	if req.CategoryID != "" {
		doc.CategoryID.String = req.CategoryID
		doc.CategoryID.Valid = true
	}

	if err := s.repo.CreateDocument(ctx, doc); err != nil {
		return nil, err
	}

	// Add tags if provided
	if len(req.Tags) > 0 {
		for _, tagName := range req.Tags {
			// TODO: Find or create tag, then associate with document
			_ = tagName
		}
	}

	logger.InfoContext(ctx, "document created",
		zap.String("document_id", doc.ID.String()),
		zap.String("name", doc.Name),
	)

	return doc, nil
}

// GetDocument retrieves a document by ID
func (s *Service) GetDocument(ctx context.Context, docID uuid.UUID) (*models.Document, error) {
	tenantID := getTenantID(ctx)

	// Try cache first
	cacheKey := cache.TenantKey(tenantID.String(), "document", docID.String())
	var doc models.Document
	if err := s.cache.Get(ctx, cacheKey, &doc); err == nil {
		return &doc, nil
	}

	// Fetch from database
	docPtr, err := s.repo.GetDocument(ctx, tenantID, docID)
	if err != nil {
		return nil, err
	}

	// Cache for future requests
	_ = s.cache.Set(ctx, cacheKey, docPtr, documentCacheTTL)

	return docPtr, nil
}

// ListDocuments retrieves documents with filtering
func (s *Service) ListDocuments(ctx context.Context, params *models.ListDocumentsParams) ([]models.Document, int64, error) {
	tenantID := getTenantID(ctx)

	params.Normalize()

	documents, total, err := s.repo.ListDocuments(ctx, tenantID, params)
	if err != nil {
		return nil, 0, err
	}

	return documents, total, nil
}

// UpdateDocument updates a document
func (s *Service) UpdateDocument(ctx context.Context, docID uuid.UUID, req *models.UpdateDocumentRequest) error {
	tenantID := getTenantID(ctx)

	// Verify document exists and belongs to tenant
	if _, err := s.repo.GetDocument(ctx, tenantID, docID); err != nil {
		return err
	}

	// Validate folder if provided
	if req.FolderID != nil && *req.FolderID != "" {
		folderUUID, _ := uuid.Parse(*req.FolderID)
		if _, err := s.repo.GetFolder(ctx, tenantID, folderUUID); err != nil {
			return errors.Validationf("invalid folder_id")
		}
	}

	// Update document
	if err := s.repo.UpdateDocument(ctx, tenantID, docID, req); err != nil {
		return err
	}

	// Invalidate cache
	cacheKey := cache.TenantKey(tenantID.String(), "document", docID.String())
	_ = s.cache.Delete(ctx, cacheKey)

	// Update tags if provided
	if len(req.Tags) > 0 {
		// TODO: Update document tags
	}

	logger.InfoContext(ctx, "document updated", zap.String("document_id", docID.String()))

	return nil
}

// DeleteDocument deletes a document
func (s *Service) DeleteDocument(ctx context.Context, docID uuid.UUID) error {
	tenantID := getTenantID(ctx)

	// Verify document exists
	doc, err := s.repo.GetDocument(ctx, tenantID, docID)
	if err != nil {
		return err
	}

	// Delete from database
	if err := s.repo.DeleteDocument(ctx, tenantID, docID); err != nil {
		return err
	}

	// Invalidate cache
	cacheKey := cache.TenantKey(tenantID.String(), "document", docID.String())
	_ = s.cache.Delete(ctx, cacheKey)

	// TODO: Delete file from storage service
	_ = doc.StoragePath

	logger.InfoContext(ctx, "document deleted", zap.String("document_id", docID.String()))

	return nil
}

// Folder operations

// CreateFolder creates a new folder
func (s *Service) CreateFolder(ctx context.Context, req *models.CreateFolderRequest) (*models.Folder, error) {
	tenantID := getTenantID(ctx)
	userID := middleware.GetUserID(ctx)

	// Build folder path
	var path string
	if req.ParentID != "" {
		parentUUID, _ := uuid.Parse(req.ParentID)
		parent, err := s.repo.GetFolder(ctx, tenantID, parentUUID)
		if err != nil {
			return nil, errors.Validationf("invalid parent_id")
		}
		path = parent.Path + "/" + sanitizeFolderName(req.Name)
	} else {
		path = "/" + sanitizeFolderName(req.Name)
	}

	folder := &models.Folder{
		ID:        uuid.New(),
		TenantID:  tenantID,
		Name:      req.Name,
		Path:      path,
		CreatedBy: userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if req.ParentID != "" {
		folder.ParentID.String = req.ParentID
		folder.ParentID.Valid = true
	}

	if req.Description != "" {
		folder.Description.String = req.Description
		folder.Description.Valid = true
	}

	if req.Color != "" {
		folder.Color.String = req.Color
		folder.Color.Valid = true
	}

	if req.Icon != "" {
		folder.Icon.String = req.Icon
		folder.Icon.Valid = true
	}

	if err := s.repo.CreateFolder(ctx, folder); err != nil {
		return nil, err
	}

	logger.InfoContext(ctx, "folder created",
		zap.String("folder_id", folder.ID.String()),
		zap.String("name", folder.Name),
	)

	return folder, nil
}

// GetFolder retrieves a folder by ID
func (s *Service) GetFolder(ctx context.Context, folderID uuid.UUID) (*models.Folder, error) {
	tenantID := getTenantID(ctx)

	folder, err := s.repo.GetFolder(ctx, tenantID, folderID)
	if err != nil {
		return nil, err
	}

	return folder, nil
}

// ListFolders retrieves folders
func (s *Service) ListFolders(ctx context.Context, parentID *string) ([]models.Folder, error) {
	tenantID := getTenantID(ctx)

	folders, err := s.repo.ListFolders(ctx, tenantID, parentID)
	if err != nil {
		return nil, err
	}

	return folders, nil
}

// DeleteFolder deletes a folder
func (s *Service) DeleteFolder(ctx context.Context, folderID uuid.UUID) error {
	tenantID := getTenantID(ctx)

	// TODO: Check if folder has documents or subfolders

	if err := s.repo.DeleteFolder(ctx, tenantID, folderID); err != nil {
		return err
	}

	logger.InfoContext(ctx, "folder deleted", zap.String("folder_id", folderID.String()))

	return nil
}

// Tag operations

// CreateTag creates a new tag
func (s *Service) CreateTag(ctx context.Context, req *models.CreateTagRequest) (*models.Tag, error) {
	tenantID := getTenantID(ctx)
	userID := middleware.GetUserID(ctx)

	tag := &models.Tag{
		ID:        uuid.New(),
		TenantID:  tenantID,
		Name:      strings.TrimSpace(req.Name),
		Color:     req.Color,
		CreatedBy: userID,
		CreatedAt: time.Now(),
	}

	if err := s.repo.CreateTag(ctx, tag); err != nil {
		return nil, err
	}

	logger.InfoContext(ctx, "tag created", zap.String("tag", tag.Name))

	return tag, nil
}

// ListTags retrieves all tags
func (s *Service) ListTags(ctx context.Context) ([]models.Tag, error) {
	tenantID := getTenantID(ctx)

	tags, err := s.repo.ListTags(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	return tags, nil
}

// Category operations

// CreateCategory creates a new category
func (s *Service) CreateCategory(ctx context.Context, req *models.CreateCategoryRequest) (*models.Category, error) {
	tenantID := getTenantID(ctx)

	category := &models.Category{
		ID:          uuid.New(),
		TenantID:    tenantID,
		Name:        req.Name,
		Description: req.Description,
		Color:       req.Color,
		Icon:        req.Icon,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.CreateCategory(ctx, category); err != nil {
		return nil, err
	}

	logger.InfoContext(ctx, "category created", zap.String("category", category.Name))

	return category, nil
}

// ListCategories retrieves all categories
func (s *Service) ListCategories(ctx context.Context) ([]models.Category, error) {
	tenantID := getTenantID(ctx)

	categories, err := s.repo.ListCategories(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	return categories, nil
}

// Helper functions

func getTenantID(ctx context.Context) uuid.UUID {
	tenantIDStr := middleware.GetTenantID(ctx)
	tenantID, _ := uuid.Parse(tenantIDStr)
	return tenantID
}

func sanitizeFolderName(name string) string {
	// Remove any path separators and special characters
	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, "\\", "-")
	return strings.TrimSpace(name)
}

// FileInfo represents uploaded file information
type FileInfo struct {
	Extension   string
	Size        int64
	MimeType    string
	StoragePath string
}

// GetFileExtension extracts file extension from filename
func GetFileExtension(filename string) string {
	ext := filepath.Ext(filename)
	if ext != "" {
		ext = strings.ToLower(ext[1:]) // Remove leading dot
	}
	return ext
}

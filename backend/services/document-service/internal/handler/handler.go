package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/SidahmedSeg/document-manager/backend/pkg/response"
	"github.com/SidahmedSeg/document-manager/backend/pkg/validator"
	"github.com/SidahmedSeg/document-manager/backend/services/document-service/internal/models"
	"github.com/SidahmedSeg/document-manager/backend/services/document-service/internal/service"
	"go.uber.org/zap"
)

// Handler handles HTTP requests for document operations
type Handler struct {
	service *service.Service
	logger  *zap.Logger
}

// NewHandler creates a new document handler
func NewHandler(svc *service.Service, logger *zap.Logger) *Handler {
	return &Handler{
		service: svc,
		logger:  logger,
	}
}

// Document handlers

// CreateDocument handles POST /api/documents
func (h *Handler) CreateDocument(w http.ResponseWriter, r *http.Request) {
	var req models.CreateDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	// Validate request
	if err := validator.Validate(&req); err != nil {
		response.ValidationError(w, err)
		return
	}

	// TODO: Handle file upload and get file info
	// For now, return error indicating file upload needed
	response.BadRequest(w, "file upload not implemented yet")
}

// GetDocument handles GET /api/documents/:id
func (h *Handler) GetDocument(w http.ResponseWriter, r *http.Request) {
	docIDStr := r.PathValue("id")
	docID, err := uuid.Parse(docIDStr)
	if err != nil {
		response.BadRequest(w, "invalid document ID")
		return
	}

	doc, err := h.service.GetDocument(r.Context(), docID)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, doc)
}

// ListDocuments handles GET /api/documents
func (h *Handler) ListDocuments(w http.ResponseWriter, r *http.Request) {
	params := &models.ListDocumentsParams{
		FolderID:   r.URL.Query().Get("folder_id"),
		CategoryID: r.URL.Query().Get("category_id"),
		Tags:       r.URL.Query().Get("tags"),
		Status:     r.URL.Query().Get("status"),
		Search:     r.URL.Query().Get("search"),
		SortBy:     r.URL.Query().Get("sort_by"),
		SortOrder:  r.URL.Query().Get("sort_order"),
	}

	// Parse page and limit
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil {
			params.Page = page
		}
	}
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			params.Limit = limit
		}
	}

	// Validate params
	if err := validator.Validate(params); err != nil {
		response.ValidationError(w, err)
		return
	}

	documents, total, err := h.service.ListDocuments(r.Context(), params)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Paginated(w, documents, params.Page, params.Limit, total)
}

// UpdateDocument handles PUT /api/documents/:id
func (h *Handler) UpdateDocument(w http.ResponseWriter, r *http.Request) {
	docIDStr := r.PathValue("id")
	docID, err := uuid.Parse(docIDStr)
	if err != nil {
		response.BadRequest(w, "invalid document ID")
		return
	}

	var req models.UpdateDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	// Validate request
	if err := validator.Validate(&req); err != nil {
		response.ValidationError(w, err)
		return
	}

	if err := h.service.UpdateDocument(r.Context(), docID, &req); err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, map[string]string{"message": "document updated successfully"})
}

// DeleteDocument handles DELETE /api/documents/:id
func (h *Handler) DeleteDocument(w http.ResponseWriter, r *http.Request) {
	docIDStr := r.PathValue("id")
	docID, err := uuid.Parse(docIDStr)
	if err != nil {
		response.BadRequest(w, "invalid document ID")
		return
	}

	if err := h.service.DeleteDocument(r.Context(), docID); err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, map[string]string{"message": "document deleted successfully"})
}

// Folder handlers

// CreateFolder handles POST /api/folders
func (h *Handler) CreateFolder(w http.ResponseWriter, r *http.Request) {
	var req models.CreateFolderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	// Validate request
	if err := validator.Validate(&req); err != nil {
		response.ValidationError(w, err)
		return
	}

	folder, err := h.service.CreateFolder(r.Context(), &req)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Created(w, folder)
}

// GetFolder handles GET /api/folders/:id
func (h *Handler) GetFolder(w http.ResponseWriter, r *http.Request) {
	folderIDStr := r.PathValue("id")
	folderID, err := uuid.Parse(folderIDStr)
	if err != nil {
		response.BadRequest(w, "invalid folder ID")
		return
	}

	folder, err := h.service.GetFolder(r.Context(), folderID)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, folder)
}

// ListFolders handles GET /api/folders
func (h *Handler) ListFolders(w http.ResponseWriter, r *http.Request) {
	var parentID *string
	if parentIDStr := r.URL.Query().Get("parent_id"); parentIDStr != "" {
		parentID = &parentIDStr
	}

	folders, err := h.service.ListFolders(r.Context(), parentID)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, folders)
}

// DeleteFolder handles DELETE /api/folders/:id
func (h *Handler) DeleteFolder(w http.ResponseWriter, r *http.Request) {
	folderIDStr := r.PathValue("id")
	folderID, err := uuid.Parse(folderIDStr)
	if err != nil {
		response.BadRequest(w, "invalid folder ID")
		return
	}

	if err := h.service.DeleteFolder(r.Context(), folderID); err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, map[string]string{"message": "folder deleted successfully"})
}

// Tag handlers

// CreateTag handles POST /api/tags
func (h *Handler) CreateTag(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	// Validate request
	if err := validator.Validate(&req); err != nil {
		response.ValidationError(w, err)
		return
	}

	tag, err := h.service.CreateTag(r.Context(), &req)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Created(w, tag)
}

// ListTags handles GET /api/tags
func (h *Handler) ListTags(w http.ResponseWriter, r *http.Request) {
	tags, err := h.service.ListTags(r.Context())
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, tags)
}

// Category handlers

// CreateCategory handles POST /api/categories
func (h *Handler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var req models.CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	// Validate request
	if err := validator.Validate(&req); err != nil {
		response.ValidationError(w, err)
		return
	}

	category, err := h.service.CreateCategory(r.Context(), &req)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Created(w, category)
}

// ListCategories handles GET /api/categories
func (h *Handler) ListCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.service.ListCategories(r.Context())
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, categories)
}

// Health check handlers

// HealthCheck handles GET /health
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response.Success(w, map[string]string{
		"status":  "healthy",
		"service": "document-service",
	})
}

// ReadyCheck handles GET /health/ready
func (h *Handler) ReadyCheck(w http.ResponseWriter, r *http.Request) {
	// TODO: Check database and cache connectivity
	response.Success(w, map[string]string{
		"status":  "ready",
		"service": "document-service",
	})
}

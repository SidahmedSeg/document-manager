package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/SidahmedSeg/document-manager/backend/pkg/response"
	"github.com/SidahmedSeg/document-manager/backend/pkg/validator"
	"github.com/SidahmedSeg/document-manager/backend/services/storage-service/internal/models"
	"github.com/SidahmedSeg/document-manager/backend/services/storage-service/internal/service"
	"go.uber.org/zap"
)

const (
	maxUploadSize = 100 * 1024 * 1024 // 100MB
)

// Handler handles HTTP requests for storage operations
type Handler struct {
	service *service.Service
	logger  *zap.Logger
}

// NewHandler creates a new storage handler
func NewHandler(svc *service.Service, logger *zap.Logger) *Handler {
	return &Handler{
		service: svc,
		logger:  logger,
	}
}

// UploadFile handles POST /api/storage/upload
func (h *Handler) UploadFile(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		response.BadRequest(w, "file too large or invalid multipart form")
		return
	}

	// Get file from form
	file, header, err := r.FormFile("file")
	if err != nil {
		response.BadRequest(w, "missing file in request")
		return
	}
	defer file.Close()

	// Get request data
	documentID := r.FormValue("document_id")
	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	req := &models.UploadFileRequest{
		DocumentID: documentID,
		FileName:   header.Filename,
		MimeType:   mimeType,
		FileSize:   header.Size,
	}

	// Validate request
	if err := validator.Validate(req); err != nil {
		response.ValidationError(w, err)
		return
	}

	// Upload file
	uploadResp, err := h.service.UploadFile(r.Context(), req, file)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Created(w, uploadResp)
}

// GetPresignedUploadURL handles POST /api/storage/presigned-upload
func (h *Handler) GetPresignedUploadURL(w http.ResponseWriter, r *http.Request) {
	var req models.UploadFileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	// Validate request
	if err := validator.Validate(&req); err != nil {
		response.ValidationError(w, err)
		return
	}

	presignedURL, err := h.service.GetPresignedUploadURL(r.Context(), &req)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, presignedURL)
}

// DownloadFile handles GET /api/storage/download/:id
func (h *Handler) DownloadFile(w http.ResponseWriter, r *http.Request) {
	fileIDStr := r.PathValue("id")
	fileID, err := uuid.Parse(fileIDStr)
	if err != nil {
		response.BadRequest(w, "invalid file ID")
		return
	}

	// Parse query parameters
	inline := r.URL.Query().Get("inline") == "true"
	expiryTime := 0
	if expiryStr := r.URL.Query().Get("expiry"); expiryStr != "" {
		if expiry, err := strconv.Atoi(expiryStr); err == nil {
			expiryTime = expiry
		}
	}

	downloadResp, err := h.service.DownloadFile(r.Context(), fileID, inline, expiryTime)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, downloadResp)
}

// DeleteFile handles DELETE /api/storage/:id
func (h *Handler) DeleteFile(w http.ResponseWriter, r *http.Request) {
	fileIDStr := r.PathValue("id")
	fileID, err := uuid.Parse(fileIDStr)
	if err != nil {
		response.BadRequest(w, "invalid file ID")
		return
	}

	// Parse query parameter for hard delete
	hardDelete := r.URL.Query().Get("hard") == "true"

	if err := h.service.DeleteFile(r.Context(), fileID, hardDelete); err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, map[string]string{"message": "file deleted successfully"})
}

// GetFileMetadata handles GET /api/storage/:id/metadata
func (h *Handler) GetFileMetadata(w http.ResponseWriter, r *http.Request) {
	fileIDStr := r.PathValue("id")
	fileID, err := uuid.Parse(fileIDStr)
	if err != nil {
		response.BadRequest(w, "invalid file ID")
		return
	}

	metadata, err := h.service.GetFileMetadata(r.Context(), fileID)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, metadata)
}

// ListFiles handles GET /api/storage
func (h *Handler) ListFiles(w http.ResponseWriter, r *http.Request) {
	params := &models.ListFilesParams{
		DocumentID: r.URL.Query().Get("document_id"),
		FileType:   r.URL.Query().Get("file_type"),
		MimeType:   r.URL.Query().Get("mime_type"),
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

	files, total, err := h.service.ListFiles(r.Context(), params)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Paginated(w, files, params.Page, params.Limit, total)
}

// GetStats handles GET /api/storage/stats
func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetFileStats(r.Context())
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, stats)
}

// HealthCheck handles GET /health
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response.Success(w, map[string]string{
		"status":  "healthy",
		"service": "storage-service",
	})
}

// ReadyCheck handles GET /health/ready
func (h *Handler) ReadyCheck(w http.ResponseWriter, r *http.Request) {
	// TODO: Check database, cache, and MinIO connectivity
	response.Success(w, map[string]string{
		"status":  "ready",
		"service": "storage-service",
	})
}

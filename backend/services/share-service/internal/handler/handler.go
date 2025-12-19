package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/SidahmedSeg/document-manager/backend/pkg/response"
	"github.com/SidahmedSeg/document-manager/backend/pkg/validator"
	"github.com/SidahmedSeg/document-manager/backend/services/share-service/internal/models"
	"github.com/SidahmedSeg/document-manager/backend/services/share-service/internal/service"
	"go.uber.org/zap"
)

// Handler handles HTTP requests for share operations
type Handler struct {
	service *service.Service
	logger  *zap.Logger
}

// NewHandler creates a new share handler
func NewHandler(svc *service.Service, logger *zap.Logger) *Handler {
	return &Handler{
		service: svc,
		logger:  logger,
	}
}

// CreateShare handles POST /api/shares
func (h *Handler) CreateShare(w http.ResponseWriter, r *http.Request) {
	var req models.CreateShareRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	// Validate request
	if err := validator.Validate(&req); err != nil {
		response.ValidationError(w, err)
		return
	}

	share, err := h.service.CreateShare(r.Context(), &req)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Created(w, share)
}

// GetShare handles GET /api/shares/:id
func (h *Handler) GetShare(w http.ResponseWriter, r *http.Request) {
	shareIDStr := r.PathValue("id")
	shareID, err := uuid.Parse(shareIDStr)
	if err != nil {
		response.BadRequest(w, "invalid share ID")
		return
	}

	share, err := h.service.GetShare(r.Context(), shareID)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, share)
}

// AccessShare handles POST /api/shares/access
func (h *Handler) AccessShare(w http.ResponseWriter, r *http.Request) {
	var req models.AccessShareRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	// Validate request
	if err := validator.Validate(&req); err != nil {
		response.ValidationError(w, err)
		return
	}

	// Get IP address and user agent
	ipAddress := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		ipAddress = forwarded
	}
	userAgent := r.Header.Get("User-Agent")

	accessResp, err := h.service.AccessShare(r.Context(), &req, ipAddress, userAgent)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, accessResp)
}

// ListShares handles GET /api/shares
func (h *Handler) ListShares(w http.ResponseWriter, r *http.Request) {
	params := &models.ListSharesParams{
		DocumentID: r.URL.Query().Get("document_id"),
		ShareType:  r.URL.Query().Get("share_type"),
		SharedWith: r.URL.Query().Get("shared_with"),
		IsActive:   r.URL.Query().Get("is_active"),
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

	shares, total, err := h.service.ListShares(r.Context(), params)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Paginated(w, shares, params.Page, params.Limit, total)
}

// UpdateShare handles PUT /api/shares/:id
func (h *Handler) UpdateShare(w http.ResponseWriter, r *http.Request) {
	shareIDStr := r.PathValue("id")
	shareID, err := uuid.Parse(shareIDStr)
	if err != nil {
		response.BadRequest(w, "invalid share ID")
		return
	}

	var req models.UpdateShareRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	// Validate request
	if err := validator.Validate(&req); err != nil {
		response.ValidationError(w, err)
		return
	}

	if err := h.service.UpdateShare(r.Context(), shareID, &req); err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, map[string]string{"message": "share updated successfully"})
}

// RevokeShare handles POST /api/shares/:id/revoke
func (h *Handler) RevokeShare(w http.ResponseWriter, r *http.Request) {
	shareIDStr := r.PathValue("id")
	shareID, err := uuid.Parse(shareIDStr)
	if err != nil {
		response.BadRequest(w, "invalid share ID")
		return
	}

	if err := h.service.RevokeShare(r.Context(), shareID); err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, map[string]string{"message": "share revoked successfully"})
}

// DeleteShare handles DELETE /api/shares/:id
func (h *Handler) DeleteShare(w http.ResponseWriter, r *http.Request) {
	shareIDStr := r.PathValue("id")
	shareID, err := uuid.Parse(shareIDStr)
	if err != nil {
		response.BadRequest(w, "invalid share ID")
		return
	}

	if err := h.service.DeleteShare(r.Context(), shareID); err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, map[string]string{"message": "share deleted successfully"})
}

// GetShareAccessLogs handles GET /api/shares/:id/access-logs
func (h *Handler) GetShareAccessLogs(w http.ResponseWriter, r *http.Request) {
	shareIDStr := r.PathValue("id")
	shareID, err := uuid.Parse(shareIDStr)
	if err != nil {
		response.BadRequest(w, "invalid share ID")
		return
	}

	// Parse limit
	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	logs, err := h.service.GetShareAccessLogs(r.Context(), shareID, limit)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, logs)
}

// GetStats handles GET /api/shares/stats
func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetShareStats(r.Context())
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, stats)
}

// VerifyToken handles POST /api/shares/verify
func (h *Handler) VerifyToken(w http.ResponseWriter, r *http.Request) {
	var req models.VerifyShareTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	// Validate request
	if err := validator.Validate(&req); err != nil {
		response.ValidationError(w, err)
		return
	}

	verifyResp, err := h.service.VerifyShareToken(r.Context(), req.Token, req.Password)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, verifyResp)
}

// HealthCheck handles GET /health
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response.Success(w, map[string]string{
		"status":  "healthy",
		"service": "share-service",
	})
}

// ReadyCheck handles GET /health/ready
func (h *Handler) ReadyCheck(w http.ResponseWriter, r *http.Request) {
	// TODO: Check database and cache connectivity
	response.Success(w, map[string]string{
		"status":  "ready",
		"service": "share-service",
	})
}

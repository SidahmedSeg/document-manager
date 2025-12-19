package handler

import (
	"encoding/json"
	"net/http"

	"github.com/SidahmedSeg/document-manager/backend/pkg/response"
	"github.com/SidahmedSeg/document-manager/backend/pkg/validator"
	"github.com/SidahmedSeg/document-manager/backend/services/quota-service/internal/models"
	"github.com/SidahmedSeg/document-manager/backend/services/quota-service/internal/service"
	"go.uber.org/zap"
)

// Handler handles HTTP requests for quota operations
type Handler struct {
	service *service.Service
	logger  *zap.Logger
}

// NewHandler creates a new quota handler
func NewHandler(svc *service.Service, logger *zap.Logger) *Handler {
	return &Handler{
		service: svc,
		logger:  logger,
	}
}

// CreateQuota handles POST /api/quotas
func (h *Handler) CreateQuota(w http.ResponseWriter, r *http.Request) {
	var req models.CreateQuotaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	// Validate request
	if err := validator.Validate(&req); err != nil {
		response.ValidationError(w, err)
		return
	}

	quota, err := h.service.CreateQuota(r.Context(), &req)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Created(w, quota)
}

// GetQuota handles GET /api/quotas/me
func (h *Handler) GetQuota(w http.ResponseWriter, r *http.Request) {
	quota, err := h.service.GetQuota(r.Context())
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, quota)
}

// UpdateQuota handles PUT /api/quotas/me
func (h *Handler) UpdateQuota(w http.ResponseWriter, r *http.Request) {
	var req models.UpdateQuotaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	// Validate request
	if err := validator.Validate(&req); err != nil {
		response.ValidationError(w, err)
		return
	}

	if err := h.service.UpdateQuota(r.Context(), &req); err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, map[string]string{"message": "quota updated successfully"})
}

// GetUsage handles GET /api/quotas/usage
func (h *Handler) GetUsage(w http.ResponseWriter, r *http.Request) {
	usage, err := h.service.GetUsage(r.Context())
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, usage)
}

// GetOverview handles GET /api/quotas/overview
func (h *Handler) GetOverview(w http.ResponseWriter, r *http.Request) {
	overview, err := h.service.GetQuotaUsageOverview(r.Context())
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, overview)
}

// CheckQuota handles POST /api/quotas/check
func (h *Handler) CheckQuota(w http.ResponseWriter, r *http.Request) {
	var req models.CheckQuotaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	// Validate request
	if err := validator.Validate(&req); err != nil {
		response.ValidationError(w, err)
		return
	}

	checkResp, err := h.service.CheckQuota(r.Context(), &req)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, checkResp)
}

// IncrementUsage handles POST /api/quotas/usage/increment
func (h *Handler) IncrementUsage(w http.ResponseWriter, r *http.Request) {
	var req models.IncrementUsageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	// Validate request
	if err := validator.Validate(&req); err != nil {
		response.ValidationError(w, err)
		return
	}

	if err := h.service.IncrementUsage(r.Context(), &req); err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, map[string]string{"message": "usage incremented successfully"})
}

// DecrementUsage handles POST /api/quotas/usage/decrement
func (h *Handler) DecrementUsage(w http.ResponseWriter, r *http.Request) {
	var req models.DecrementUsageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	// Validate request
	if err := validator.Validate(&req); err != nil {
		response.ValidationError(w, err)
		return
	}

	if err := h.service.DecrementUsage(r.Context(), &req); err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, map[string]string{"message": "usage decremented successfully"})
}

// GetUsageStats handles GET /api/quotas/stats
func (h *Handler) GetUsageStats(w http.ResponseWriter, r *http.Request) {
	params := &models.UsageStatsParams{
		StartDate: r.URL.Query().Get("start_date"),
		EndDate:   r.URL.Query().Get("end_date"),
		Resource:  r.URL.Query().Get("resource"),
		Action:    r.URL.Query().Get("action"),
	}

	// Validate params
	if err := validator.Validate(params); err != nil {
		response.ValidationError(w, err)
		return
	}

	stats, err := h.service.GetUsageStats(r.Context(), params)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, stats)
}

// GetUsageLogs handles GET /api/quotas/logs
func (h *Handler) GetUsageLogs(w http.ResponseWriter, r *http.Request) {
	params := &models.UsageStatsParams{
		StartDate: r.URL.Query().Get("start_date"),
		EndDate:   r.URL.Query().Get("end_date"),
		Resource:  r.URL.Query().Get("resource"),
		Action:    r.URL.Query().Get("action"),
	}

	// Validate params
	if err := validator.Validate(params); err != nil {
		response.ValidationError(w, err)
		return
	}

	logs, err := h.service.GetUsageLogs(r.Context(), params)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, logs)
}

// GetPredefinedPlans handles GET /api/quotas/plans
func (h *Handler) GetPredefinedPlans(w http.ResponseWriter, r *http.Request) {
	plans := h.service.GetPredefinedPlans()
	response.Success(w, plans)
}

// HealthCheck handles GET /health
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response.Success(w, map[string]string{
		"status":  "healthy",
		"service": "quota-service",
	})
}

// ReadyCheck handles GET /health/ready
func (h *Handler) ReadyCheck(w http.ResponseWriter, r *http.Request) {
	// TODO: Check database and cache connectivity
	response.Success(w, map[string]string{
		"status":  "ready",
		"service": "quota-service",
	})
}

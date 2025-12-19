package handler

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/SidahmedSeg/document-manager/backend/pkg/response"
	"github.com/SidahmedSeg/document-manager/backend/pkg/validator"
	"github.com/SidahmedSeg/document-manager/backend/services/tenant-service/internal/models"
	"github.com/SidahmedSeg/document-manager/backend/services/tenant-service/internal/service"
	"go.uber.org/zap"
)

// Handler handles HTTP requests for tenant operations
type Handler struct {
	service *service.Service
	logger  *zap.Logger
}

// NewHandler creates a new tenant handler
func NewHandler(svc *service.Service, logger *zap.Logger) *Handler {
	return &Handler{
		service: svc,
		logger:  logger,
	}
}

// CreateTenant handles POST /api/tenants
func (h *Handler) CreateTenant(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	// Validate request
	if err := validator.Validate(&req); err != nil {
		response.ValidationError(w, err)
		return
	}

	// Validate and normalize slug
	slug, err := service.ValidateSlug(req.Slug)
	if err != nil {
		response.ValidationError(w, err)
		return
	}
	req.Slug = slug

	// Create tenant
	tenant, err := h.service.CreateTenant(r.Context(), &req)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Created(w, tenant)
}

// GetTenant handles GET /api/tenants/:id
func (h *Handler) GetTenant(w http.ResponseWriter, r *http.Request) {
	// Extract tenant ID from URL path
	tenantIDStr := r.PathValue("id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		response.BadRequest(w, "invalid tenant ID")
		return
	}

	tenant, err := h.service.GetTenant(r.Context(), tenantID)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, tenant)
}

// UpdateTenant handles PUT /api/tenants/:id
func (h *Handler) UpdateTenant(w http.ResponseWriter, r *http.Request) {
	tenantIDStr := r.PathValue("id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		response.BadRequest(w, "invalid tenant ID")
		return
	}

	var req models.UpdateTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	// Validate request
	if err := validator.Validate(&req); err != nil {
		response.ValidationError(w, err)
		return
	}

	if err := h.service.UpdateTenant(r.Context(), tenantID, &req); err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, map[string]string{"message": "tenant updated successfully"})
}

// GetTenantUsers handles GET /api/tenants/:id/users
func (h *Handler) GetTenantUsers(w http.ResponseWriter, r *http.Request) {
	tenantIDStr := r.PathValue("id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		response.BadRequest(w, "invalid tenant ID")
		return
	}

	users, err := h.service.GetTenantUsers(r.Context(), tenantID)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, users)
}

// InviteUser handles POST /api/tenants/:id/users/invite
func (h *Handler) InviteUser(w http.ResponseWriter, r *http.Request) {
	tenantIDStr := r.PathValue("id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		response.BadRequest(w, "invalid tenant ID")
		return
	}

	var req models.InviteUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	// Validate request
	if err := validator.Validate(&req); err != nil {
		response.ValidationError(w, err)
		return
	}

	invitation, err := h.service.InviteUser(r.Context(), tenantID, &req)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Created(w, invitation)
}

// RemoveUser handles DELETE /api/tenants/:id/users/:userId
func (h *Handler) RemoveUser(w http.ResponseWriter, r *http.Request) {
	tenantIDStr := r.PathValue("id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		response.BadRequest(w, "invalid tenant ID")
		return
	}

	userID := r.PathValue("userId")
	if userID == "" {
		response.BadRequest(w, "user ID is required")
		return
	}

	if err := h.service.RemoveUser(r.Context(), tenantID, userID); err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, map[string]string{"message": "user removed successfully"})
}

// GetUserTenants handles GET /api/tenants/me
func (h *Handler) GetUserTenants(w http.ResponseWriter, r *http.Request) {
	tenants, err := h.service.GetUserTenants(r.Context())
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, tenants)
}

// GetPendingInvitations handles GET /api/tenants/:id/invitations
func (h *Handler) GetPendingInvitations(w http.ResponseWriter, r *http.Request) {
	tenantIDStr := r.PathValue("id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		response.BadRequest(w, "invalid tenant ID")
		return
	}

	invitations, err := h.service.GetPendingInvitations(r.Context(), tenantID)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, invitations)
}

// HealthCheck handles GET /health
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response.Success(w, map[string]string{
		"status":  "healthy",
		"service": "tenant-service",
	})
}

// ReadyCheck handles GET /health/ready
func (h *Handler) ReadyCheck(w http.ResponseWriter, r *http.Request) {
	// TODO: Check database and cache connectivity
	response.Success(w, map[string]string{
		"status":  "ready",
		"service": "tenant-service",
	})
}

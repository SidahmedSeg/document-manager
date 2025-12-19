package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/SidahmedSeg/document-manager/backend/pkg/response"
	"github.com/SidahmedSeg/document-manager/backend/pkg/validator"
	"github.com/SidahmedSeg/document-manager/backend/services/rbac-service/internal/models"
	"github.com/SidahmedSeg/document-manager/backend/services/rbac-service/internal/service"
	"go.uber.org/zap"
)

// Handler handles HTTP requests for RBAC operations
type Handler struct {
	service *service.Service
	logger  *zap.Logger
}

// NewHandler creates a new RBAC handler
func NewHandler(svc *service.Service, logger *zap.Logger) *Handler {
	return &Handler{
		service: svc,
		logger:  logger,
	}
}

// Role handlers

// CreateRole handles POST /api/roles
func (h *Handler) CreateRole(w http.ResponseWriter, r *http.Request) {
	var req models.CreateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	// Validate request
	if err := validator.Validate(&req); err != nil {
		response.ValidationError(w, err)
		return
	}

	role, err := h.service.CreateRole(r.Context(), &req)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Created(w, role)
}

// GetRole handles GET /api/roles/:id
func (h *Handler) GetRole(w http.ResponseWriter, r *http.Request) {
	roleIDStr := r.PathValue("id")
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		response.BadRequest(w, "invalid role ID")
		return
	}

	role, err := h.service.GetRole(r.Context(), roleID)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, role)
}

// GetRoleWithPermissions handles GET /api/roles/:id/permissions
func (h *Handler) GetRoleWithPermissions(w http.ResponseWriter, r *http.Request) {
	roleIDStr := r.PathValue("id")
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		response.BadRequest(w, "invalid role ID")
		return
	}

	roleWithPerms, err := h.service.GetRoleWithPermissions(r.Context(), roleID)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, roleWithPerms)
}

// ListRoles handles GET /api/roles
func (h *Handler) ListRoles(w http.ResponseWriter, r *http.Request) {
	params := &models.ListRolesParams{
		IsSystem:  r.URL.Query().Get("is_system"),
		IsDefault: r.URL.Query().Get("is_default"),
		SortBy:    r.URL.Query().Get("sort_by"),
		SortOrder: r.URL.Query().Get("sort_order"),
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

	roles, total, err := h.service.ListRoles(r.Context(), params)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Paginated(w, roles, params.Page, params.Limit, total)
}

// UpdateRole handles PUT /api/roles/:id
func (h *Handler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	roleIDStr := r.PathValue("id")
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		response.BadRequest(w, "invalid role ID")
		return
	}

	var req models.UpdateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	// Validate request
	if err := validator.Validate(&req); err != nil {
		response.ValidationError(w, err)
		return
	}

	if err := h.service.UpdateRole(r.Context(), roleID, &req); err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, map[string]string{"message": "role updated successfully"})
}

// DeleteRole handles DELETE /api/roles/:id
func (h *Handler) DeleteRole(w http.ResponseWriter, r *http.Request) {
	roleIDStr := r.PathValue("id")
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		response.BadRequest(w, "invalid role ID")
		return
	}

	if err := h.service.DeleteRole(r.Context(), roleID); err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, map[string]string{"message": "role deleted successfully"})
}

// Permission handlers

// CreatePermission handles POST /api/permissions
func (h *Handler) CreatePermission(w http.ResponseWriter, r *http.Request) {
	var req models.CreatePermissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	// Validate request
	if err := validator.Validate(&req); err != nil {
		response.ValidationError(w, err)
		return
	}

	permission, err := h.service.CreatePermission(r.Context(), &req)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Created(w, permission)
}

// GetPermission handles GET /api/permissions/:id
func (h *Handler) GetPermission(w http.ResponseWriter, r *http.Request) {
	permIDStr := r.PathValue("id")
	permID, err := uuid.Parse(permIDStr)
	if err != nil {
		response.BadRequest(w, "invalid permission ID")
		return
	}

	permission, err := h.service.GetPermission(r.Context(), permID)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, permission)
}

// ListPermissions handles GET /api/permissions
func (h *Handler) ListPermissions(w http.ResponseWriter, r *http.Request) {
	params := &models.ListPermissionsParams{
		Resource:  r.URL.Query().Get("resource"),
		Action:    r.URL.Query().Get("action"),
		SortBy:    r.URL.Query().Get("sort_by"),
		SortOrder: r.URL.Query().Get("sort_order"),
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

	permissions, total, err := h.service.ListPermissions(r.Context(), params)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Paginated(w, permissions, params.Page, params.Limit, total)
}

// User Role handlers

// AssignRole handles POST /api/user-roles
func (h *Handler) AssignRole(w http.ResponseWriter, r *http.Request) {
	var req models.AssignRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	// Validate request
	if err := validator.Validate(&req); err != nil {
		response.ValidationError(w, err)
		return
	}

	if err := h.service.AssignRole(r.Context(), &req); err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, map[string]string{"message": "role assigned successfully"})
}

// BulkAssignRole handles POST /api/user-roles/bulk
func (h *Handler) BulkAssignRole(w http.ResponseWriter, r *http.Request) {
	var req models.BulkAssignRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	// Validate request
	if err := validator.Validate(&req); err != nil {
		response.ValidationError(w, err)
		return
	}

	result, err := h.service.BulkAssignRole(r.Context(), &req)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, result)
}

// RemoveRole handles DELETE /api/user-roles/:userId/roles/:roleId
func (h *Handler) RemoveRole(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userId")
	if userID == "" {
		response.BadRequest(w, "user ID is required")
		return
	}

	roleIDStr := r.PathValue("roleId")
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		response.BadRequest(w, "invalid role ID")
		return
	}

	if err := h.service.RemoveRole(r.Context(), userID, roleID); err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, map[string]string{"message": "role removed successfully"})
}

// GetUserRoles handles GET /api/user-roles/:userId
func (h *Handler) GetUserRoles(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userId")
	if userID == "" {
		response.BadRequest(w, "user ID is required")
		return
	}

	roles, err := h.service.GetUserRoles(r.Context(), userID)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, roles)
}

// GetUserPermissions handles GET /api/user-roles/:userId/permissions
func (h *Handler) GetUserPermissions(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userId")
	if userID == "" {
		response.BadRequest(w, "user ID is required")
		return
	}

	permissions, err := h.service.GetUserPermissions(r.Context(), userID)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, permissions)
}

// CheckPermission handles POST /api/permissions/check
func (h *Handler) CheckPermission(w http.ResponseWriter, r *http.Request) {
	var req models.CheckPermissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	// Validate request
	if err := validator.Validate(&req); err != nil {
		response.ValidationError(w, err)
		return
	}

	checkResp, err := h.service.CheckPermission(r.Context(), &req)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, checkResp)
}

// GetStats handles GET /api/rbac/stats
func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetRBACStats(r.Context())
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
		"service": "rbac-service",
	})
}

// ReadyCheck handles GET /health/ready
func (h *Handler) ReadyCheck(w http.ResponseWriter, r *http.Request) {
	// TODO: Check database and cache connectivity
	response.Success(w, map[string]string{
		"status":  "ready",
		"service": "rbac-service",
	})
}

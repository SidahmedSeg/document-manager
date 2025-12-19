package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/SidahmedSeg/document-manager/backend/pkg/cache"
	"github.com/SidahmedSeg/document-manager/backend/pkg/errors"
	"github.com/SidahmedSeg/document-manager/backend/pkg/logger"
	"github.com/SidahmedSeg/document-manager/backend/pkg/middleware"
	"github.com/SidahmedSeg/document-manager/backend/services/rbac-service/internal/models"
	"github.com/SidahmedSeg/document-manager/backend/services/rbac-service/internal/repository"
	"go.uber.org/zap"
)

const (
	roleCacheTTL       = 1 * time.Hour
	permissionCacheTTL = 2 * time.Hour
	userRoleCacheTTL   = 30 * time.Minute
)

// Service handles RBAC business logic
type Service struct {
	repo   *repository.Repository
	cache  *cache.Cache
	logger *zap.Logger
}

// NewService creates a new RBAC service
func NewService(repo *repository.Repository, cache *cache.Cache, logger *zap.Logger) *Service {
	return &Service{
		repo:   repo,
		cache:  cache,
		logger: logger,
	}
}

// Role operations

// CreateRole creates a new role
func (s *Service) CreateRole(ctx context.Context, req *models.CreateRoleRequest) (*models.Role, error) {
	tenantID := getTenantID(ctx)
	userID := middleware.GetUserID(ctx)

	// Check if role name already exists
	existing, _ := s.repo.GetRoleByName(ctx, tenantID, req.Name)
	if existing != nil {
		return nil, errors.Conflictf("role with name '%s' already exists", req.Name)
	}

	// Create role
	role := &models.Role{
		ID:        uuid.New(),
		TenantID:  tenantID,
		Name:      req.Name,
		IsSystem:  false,
		IsDefault: req.IsDefault,
		CreatedBy: userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if req.Description != "" {
		role.Description.String = req.Description
		role.Description.Valid = true
	}

	if err := s.repo.CreateRole(ctx, role); err != nil {
		return nil, err
	}

	// Assign permissions if provided
	if len(req.Permissions) > 0 {
		permIDs := make([]uuid.UUID, 0, len(req.Permissions))
		for _, permIDStr := range req.Permissions {
			permID, err := uuid.Parse(permIDStr)
			if err != nil {
				continue
			}
			permIDs = append(permIDs, permID)
		}
		if len(permIDs) > 0 {
			_ = s.repo.AssignPermissionsToRole(ctx, role.ID, permIDs)
		}
	}

	logger.InfoContext(ctx, "role created",
		zap.String("role_id", role.ID.String()),
		zap.String("name", role.Name),
	)

	return role, nil
}

// GetRole retrieves a role by ID
func (s *Service) GetRole(ctx context.Context, roleID uuid.UUID) (*models.Role, error) {
	tenantID := getTenantID(ctx)

	// Try cache first
	cacheKey := cache.TenantKey(tenantID.String(), "role", roleID.String())
	var role models.Role
	if err := s.cache.Get(ctx, cacheKey, &role); err == nil {
		return &role, nil
	}

	// Fetch from database
	rolePtr, err := s.repo.GetRole(ctx, tenantID, roleID)
	if err != nil {
		return nil, err
	}

	// Cache for future requests
	_ = s.cache.Set(ctx, cacheKey, rolePtr, roleCacheTTL)

	return rolePtr, nil
}

// GetRoleWithPermissions retrieves a role with its permissions
func (s *Service) GetRoleWithPermissions(ctx context.Context, roleID uuid.UUID) (*models.RoleWithPermissions, error) {
	tenantID := getTenantID(ctx)

	// Get role
	role, err := s.repo.GetRole(ctx, tenantID, roleID)
	if err != nil {
		return nil, err
	}

	// Get permissions
	permissions, err := s.repo.GetRolePermissions(ctx, roleID)
	if err != nil {
		return nil, err
	}

	return &models.RoleWithPermissions{
		Role:        *role,
		Permissions: permissions,
	}, nil
}

// ListRoles retrieves roles with filtering
func (s *Service) ListRoles(ctx context.Context, params *models.ListRolesParams) ([]models.Role, int64, error) {
	tenantID := getTenantID(ctx)

	params.Normalize()

	roles, total, err := s.repo.ListRoles(ctx, tenantID, params)
	if err != nil {
		return nil, 0, err
	}

	return roles, total, nil
}

// UpdateRole updates a role
func (s *Service) UpdateRole(ctx context.Context, roleID uuid.UUID, req *models.UpdateRoleRequest) error {
	tenantID := getTenantID(ctx)

	// Verify role exists
	role, err := s.repo.GetRole(ctx, tenantID, roleID)
	if err != nil {
		return err
	}

	// System roles cannot be modified
	if role.IsSystem {
		return errors.Forbiddenf("cannot modify system role")
	}

	// Build updates map
	updates := make(map[string]interface{})

	if req.Name != "" {
		// Check if new name already exists
		existing, _ := s.repo.GetRoleByName(ctx, tenantID, req.Name)
		if existing != nil && existing.ID != roleID {
			return errors.Conflictf("role with name '%s' already exists", req.Name)
		}
		updates["name"] = req.Name
	}

	if req.Description != "" {
		updates["description"] = req.Description
	}

	if req.IsDefault != nil {
		updates["is_default"] = *req.IsDefault
	}

	// Update role
	if len(updates) > 0 {
		if err := s.repo.UpdateRole(ctx, tenantID, roleID, updates); err != nil {
			return err
		}
	}

	// Update permissions if provided
	if len(req.Permissions) > 0 {
		permIDs := make([]uuid.UUID, 0, len(req.Permissions))
		for _, permIDStr := range req.Permissions {
			permID, err := uuid.Parse(permIDStr)
			if err != nil {
				continue
			}
			permIDs = append(permIDs, permID)
		}
		if err := s.repo.AssignPermissionsToRole(ctx, roleID, permIDs); err != nil {
			return err
		}
	}

	// Invalidate cache
	cacheKey := cache.TenantKey(tenantID.String(), "role", roleID.String())
	_ = s.cache.Delete(ctx, cacheKey)

	logger.InfoContext(ctx, "role updated", zap.String("role_id", roleID.String()))

	return nil
}

// DeleteRole deletes a role
func (s *Service) DeleteRole(ctx context.Context, roleID uuid.UUID) error {
	tenantID := getTenantID(ctx)

	// Verify role exists and is not a system role
	role, err := s.repo.GetRole(ctx, tenantID, roleID)
	if err != nil {
		return err
	}

	if role.IsSystem {
		return errors.Forbiddenf("cannot delete system role")
	}

	if err := s.repo.DeleteRole(ctx, tenantID, roleID); err != nil {
		return err
	}

	// Invalidate cache
	cacheKey := cache.TenantKey(tenantID.String(), "role", roleID.String())
	_ = s.cache.Delete(ctx, cacheKey)

	logger.InfoContext(ctx, "role deleted", zap.String("role_id", roleID.String()))

	return nil
}

// Permission operations

// CreatePermission creates a new permission
func (s *Service) CreatePermission(ctx context.Context, req *models.CreatePermissionRequest) (*models.Permission, error) {
	permission := &models.Permission{
		ID:        uuid.New(),
		Name:      req.Name,
		Resource:  req.Resource,
		Action:    req.Action,
		CreatedAt: time.Now(),
	}

	if req.Description != "" {
		permission.Description.String = req.Description
		permission.Description.Valid = true
	}

	if err := s.repo.CreatePermission(ctx, permission); err != nil {
		return nil, err
	}

	logger.InfoContext(ctx, "permission created",
		zap.String("permission_id", permission.ID.String()),
		zap.String("name", permission.Name),
	)

	return permission, nil
}

// GetPermission retrieves a permission by ID
func (s *Service) GetPermission(ctx context.Context, permissionID uuid.UUID) (*models.Permission, error) {
	return s.repo.GetPermission(ctx, permissionID)
}

// ListPermissions retrieves permissions with filtering
func (s *Service) ListPermissions(ctx context.Context, params *models.ListPermissionsParams) ([]models.Permission, int64, error) {
	params.Normalize()

	permissions, total, err := s.repo.ListPermissions(ctx, params)
	if err != nil {
		return nil, 0, err
	}

	return permissions, total, nil
}

// User Role operations

// AssignRole assigns a role to a user
func (s *Service) AssignRole(ctx context.Context, req *models.AssignRoleRequest) error {
	tenantID := getTenantID(ctx)
	assignedBy := middleware.GetUserID(ctx)

	// Parse role ID
	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		return errors.Validationf("invalid role_id")
	}

	// Verify role exists
	if _, err := s.repo.GetRole(ctx, tenantID, roleID); err != nil {
		return err
	}

	// Create user role assignment
	userRole := &models.UserRole{
		ID:         uuid.New(),
		TenantID:   tenantID,
		UserID:     req.UserID,
		RoleID:     roleID,
		AssignedBy: assignedBy,
		CreatedAt:  time.Now(),
	}

	if err := s.repo.AssignRoleToUser(ctx, userRole); err != nil {
		return err
	}

	// Invalidate user permissions cache
	userPermCacheKey := cache.TenantKey(tenantID.String(), "user_permissions", req.UserID)
	_ = s.cache.Delete(ctx, userPermCacheKey)

	logger.InfoContext(ctx, "role assigned to user",
		zap.String("user_id", req.UserID),
		zap.String("role_id", req.RoleID),
	)

	return nil
}

// BulkAssignRole assigns a role to multiple users
func (s *Service) BulkAssignRole(ctx context.Context, req *models.BulkAssignRoleRequest) (*models.BulkAssignRoleResponse, error) {
	tenantID := getTenantID(ctx)
	assignedBy := middleware.GetUserID(ctx)

	// Parse role ID
	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		return nil, errors.Validationf("invalid role_id")
	}

	// Verify role exists
	if _, err := s.repo.GetRole(ctx, tenantID, roleID); err != nil {
		return nil, err
	}

	response := &models.BulkAssignRoleResponse{
		Errors: []string{},
	}

	for _, userID := range req.UserIDs {
		userRole := &models.UserRole{
			ID:         uuid.New(),
			TenantID:   tenantID,
			UserID:     userID,
			RoleID:     roleID,
			AssignedBy: assignedBy,
			CreatedAt:  time.Now(),
		}

		if err := s.repo.AssignRoleToUser(ctx, userRole); err != nil {
			response.Failed++
			response.Errors = append(response.Errors, userID+": "+err.Error())
		} else {
			response.Assigned++
			// Invalidate cache
			userPermCacheKey := cache.TenantKey(tenantID.String(), "user_permissions", userID)
			_ = s.cache.Delete(ctx, userPermCacheKey)
		}
	}

	return response, nil
}

// RemoveRole removes a role from a user
func (s *Service) RemoveRole(ctx context.Context, userID string, roleID uuid.UUID) error {
	tenantID := getTenantID(ctx)

	if err := s.repo.RemoveRoleFromUser(ctx, tenantID, userID, roleID); err != nil {
		return err
	}

	// Invalidate user permissions cache
	userPermCacheKey := cache.TenantKey(tenantID.String(), "user_permissions", userID)
	_ = s.cache.Delete(ctx, userPermCacheKey)

	logger.InfoContext(ctx, "role removed from user",
		zap.String("user_id", userID),
		zap.String("role_id", roleID.String()),
	)

	return nil
}

// GetUserRoles retrieves all roles for a user
func (s *Service) GetUserRoles(ctx context.Context, userID string) ([]models.Role, error) {
	tenantID := getTenantID(ctx)

	roles, err := s.repo.GetUserRoles(ctx, tenantID, userID)
	if err != nil {
		return nil, err
	}

	return roles, nil
}

// CheckPermission checks if a user has a specific permission
func (s *Service) CheckPermission(ctx context.Context, req *models.CheckPermissionRequest) (*models.CheckPermissionResponse, error) {
	tenantID := getTenantID(ctx)

	// Try cache first
	cacheKey := cache.TenantKey(tenantID.String(), "permission_check", req.UserID, req.Resource, req.Action)
	var response models.CheckPermissionResponse
	if err := s.cache.Get(ctx, cacheKey, &response); err == nil {
		return &response, nil
	}

	// Check permission
	allowed, err := s.repo.CheckUserPermission(ctx, tenantID, req.UserID, req.Resource, req.Action)
	if err != nil {
		return nil, err
	}

	response = models.CheckPermissionResponse{
		Allowed:  allowed,
		UserID:   req.UserID,
		Resource: req.Resource,
		Action:   req.Action,
	}

	// Get user roles and permissions for context
	if allowed {
		roles, _ := s.repo.GetUserRoles(ctx, tenantID, req.UserID)
		roleNames := make([]string, len(roles))
		for i, role := range roles {
			roleNames[i] = role.Name
		}
		response.Roles = roleNames

		permissions, _ := s.repo.GetUserPermissions(ctx, tenantID, req.UserID)
		permNames := make([]string, len(permissions))
		for i, perm := range permissions {
			permNames[i] = perm.Name
		}
		response.Permissions = permNames
	}

	// Cache result
	_ = s.cache.Set(ctx, cacheKey, &response, userRoleCacheTTL)

	return &response, nil
}

// GetUserPermissions retrieves all permissions for a user
func (s *Service) GetUserPermissions(ctx context.Context, userID string) ([]models.Permission, error) {
	tenantID := getTenantID(ctx)

	// Try cache first
	cacheKey := cache.TenantKey(tenantID.String(), "user_permissions", userID)
	var permissions []models.Permission
	if err := s.cache.Get(ctx, cacheKey, &permissions); err == nil {
		return permissions, nil
	}

	// Fetch from database
	permissions, err := s.repo.GetUserPermissions(ctx, tenantID, userID)
	if err != nil {
		return nil, err
	}

	// Cache for future requests
	_ = s.cache.Set(ctx, cacheKey, permissions, userRoleCacheTTL)

	return permissions, nil
}

// GetRBACStats retrieves RBAC statistics
func (s *Service) GetRBACStats(ctx context.Context) (*models.RBACStats, error) {
	tenantID := getTenantID(ctx)

	stats, err := s.repo.GetRBACStats(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// Helper functions

func getTenantID(ctx context.Context) uuid.UUID {
	tenantIDStr := middleware.GetTenantID(ctx)
	tenantID, _ := uuid.Parse(tenantIDStr)
	return tenantID
}

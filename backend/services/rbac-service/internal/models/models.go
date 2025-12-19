package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// Role represents a role in the system
type Role struct {
	ID          uuid.UUID      `json:"id" db:"id"`
	TenantID    uuid.UUID      `json:"tenant_id" db:"tenant_id"`
	Name        string         `json:"name" db:"name"`
	Description sql.NullString `json:"description,omitempty" db:"description"`
	IsSystem    bool           `json:"is_system" db:"is_system"` // System roles can't be deleted
	IsDefault   bool           `json:"is_default" db:"is_default"` // Default role for new users
	CreatedBy   string         `json:"created_by" db:"created_by"`
	CreatedAt   time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at" db:"updated_at"`
}

// Permission represents a permission in the system
type Permission struct {
	ID          uuid.UUID      `json:"id" db:"id"`
	Name        string         `json:"name" db:"name"`
	Resource    string         `json:"resource" db:"resource"` // e.g., document, folder, share
	Action      string         `json:"action" db:"action"`     // e.g., create, read, update, delete
	Description sql.NullString `json:"description,omitempty" db:"description"`
	CreatedAt   time.Time      `json:"created_at" db:"created_at"`
}

// RolePermission represents the association between roles and permissions
type RolePermission struct {
	RoleID       uuid.UUID `json:"role_id" db:"role_id"`
	PermissionID uuid.UUID `json:"permission_id" db:"permission_id"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// UserRole represents a user's role assignment
type UserRole struct {
	ID        uuid.UUID `json:"id" db:"id"`
	TenantID  uuid.UUID `json:"tenant_id" db:"tenant_id"`
	UserID    string    `json:"user_id" db:"user_id"`
	RoleID    uuid.UUID `json:"role_id" db:"role_id"`
	AssignedBy string   `json:"assigned_by" db:"assigned_by"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// RoleWithPermissions includes role with its permissions
type RoleWithPermissions struct {
	Role
	Permissions []Permission `json:"permissions"`
}

// UserRoleWithDetails includes user role with role and user details
type UserRoleWithDetails struct {
	UserRole
	RoleName        string `json:"role_name"`
	RoleDescription string `json:"role_description,omitempty"`
}

// CreateRoleRequest represents role creation request
type CreateRoleRequest struct {
	Name        string   `json:"name" validate:"required,min=2,max=50"`
	Description string   `json:"description,omitempty" validate:"omitempty,max=255"`
	IsDefault   bool     `json:"is_default,omitempty"`
	Permissions []string `json:"permissions,omitempty"` // Permission IDs
}

// UpdateRoleRequest represents role update request
type UpdateRoleRequest struct {
	Name        string   `json:"name,omitempty" validate:"omitempty,min=2,max=50"`
	Description string   `json:"description,omitempty" validate:"omitempty,max=255"`
	IsDefault   *bool    `json:"is_default,omitempty"`
	Permissions []string `json:"permissions,omitempty"` // Permission IDs to replace existing
}

// AssignRoleRequest represents role assignment request
type AssignRoleRequest struct {
	UserID string `json:"user_id" validate:"required"`
	RoleID string `json:"role_id" validate:"required,uuid"`
}

// CheckPermissionRequest represents permission check request
type CheckPermissionRequest struct {
	UserID   string `json:"user_id" validate:"required"`
	Resource string `json:"resource" validate:"required"`
	Action   string `json:"action" validate:"required"`
}

// CheckPermissionResponse represents permission check response
type CheckPermissionResponse struct {
	Allowed     bool     `json:"allowed"`
	UserID      string   `json:"user_id"`
	Resource    string   `json:"resource"`
	Action      string   `json:"action"`
	Roles       []string `json:"roles,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
}

// CreatePermissionRequest represents permission creation request
type CreatePermissionRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Resource    string `json:"resource" validate:"required,min=2,max=50"`
	Action      string `json:"action" validate:"required,oneof=create read update delete manage share"`
	Description string `json:"description,omitempty" validate:"omitempty,max=255"`
}

// ListRolesParams represents query parameters for listing roles
type ListRolesParams struct {
	IsSystem  string `json:"is_system,omitempty" form:"is_system"`
	IsDefault string `json:"is_default,omitempty" form:"is_default"`
	Page      int    `json:"page" form:"page" validate:"omitempty,gte=1"`
	Limit     int    `json:"limit" form:"limit" validate:"omitempty,gte=1,lte=100"`
	SortBy    string `json:"sort_by,omitempty" form:"sort_by"`
	SortOrder string `json:"sort_order,omitempty" form:"sort_order" validate:"omitempty,oneof=asc desc"`
}

// Normalize sets default values for list parameters
func (p *ListRolesParams) Normalize() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.Limit < 1 {
		p.Limit = 20
	}
	if p.Limit > 100 {
		p.Limit = 100
	}
	if p.SortBy == "" {
		p.SortBy = "name"
	}
	if p.SortOrder == "" {
		p.SortOrder = "asc"
	}
}

// GetOffset calculates the database offset
func (p *ListRolesParams) GetOffset() int {
	return (p.Page - 1) * p.Limit
}

// ListPermissionsParams represents query parameters for listing permissions
type ListPermissionsParams struct {
	Resource  string `json:"resource,omitempty" form:"resource"`
	Action    string `json:"action,omitempty" form:"action"`
	Page      int    `json:"page" form:"page" validate:"omitempty,gte=1"`
	Limit     int    `json:"limit" form:"limit" validate:"omitempty,gte=1,lte=100"`
	SortBy    string `json:"sort_by,omitempty" form:"sort_by"`
	SortOrder string `json:"sort_order,omitempty" form:"sort_order" validate:"omitempty,oneof=asc desc"`
}

// Normalize sets default values for list parameters
func (p *ListPermissionsParams) Normalize() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.Limit < 1 {
		p.Limit = 20
	}
	if p.Limit > 100 {
		p.Limit = 100
	}
	if p.SortBy == "" {
		p.SortBy = "name"
	}
	if p.SortOrder == "" {
		p.SortOrder = "asc"
	}
}

// GetOffset calculates the database offset
func (p *ListPermissionsParams) GetOffset() int {
	return (p.Page - 1) * p.Limit
}

// RBACStats represents RBAC statistics
type RBACStats struct {
	TotalRoles       int64            `json:"total_roles"`
	SystemRoles      int64            `json:"system_roles"`
	CustomRoles      int64            `json:"custom_roles"`
	TotalPermissions int64            `json:"total_permissions"`
	TotalUserRoles   int64            `json:"total_user_roles"`
	RoleDistribution map[string]int64 `json:"role_distribution"` // role_name -> count
}

// BulkAssignRoleRequest represents bulk role assignment
type BulkAssignRoleRequest struct {
	UserIDs []string `json:"user_ids" validate:"required,min=1,max=100"`
	RoleID  string   `json:"role_id" validate:"required,uuid"`
}

// BulkAssignRoleResponse represents bulk assignment response
type BulkAssignRoleResponse struct {
	Assigned int      `json:"assigned"`
	Failed   int      `json:"failed"`
	Errors   []string `json:"errors,omitempty"`
}

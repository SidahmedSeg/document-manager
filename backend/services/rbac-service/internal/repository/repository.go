package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/SidahmedSeg/document-manager/backend/pkg/database"
	"github.com/SidahmedSeg/document-manager/backend/pkg/errors"
	"github.com/SidahmedSeg/document-manager/backend/services/rbac-service/internal/models"
	"go.uber.org/zap"
)

// Repository handles RBAC database operations
type Repository struct {
	db     *database.DB
	logger *zap.Logger
}

// NewRepository creates a new RBAC repository
func NewRepository(db *database.DB, logger *zap.Logger) *Repository {
	return &Repository{
		db:     db,
		logger: logger,
	}
}

// Role operations

// CreateRole creates a new role
func (r *Repository) CreateRole(ctx context.Context, role *models.Role) error {
	query := `
		INSERT INTO roles (
			id, tenant_id, name, description, is_system,
			is_default, created_by, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := r.db.ExecContext(ctx, query,
		role.ID,
		role.TenantID,
		role.Name,
		role.Description,
		role.IsSystem,
		role.IsDefault,
		role.CreatedBy,
		role.CreatedAt,
		role.UpdatedAt,
	)

	if err != nil {
		r.logger.Error("failed to create role", zap.Error(err))
		return errors.New(errors.ErrCodeInternal, "failed to create role")
	}

	return nil
}

// GetRole retrieves a role by ID
func (r *Repository) GetRole(ctx context.Context, tenantID, roleID uuid.UUID) (*models.Role, error) {
	query := `
		SELECT id, tenant_id, name, description, is_system,
			is_default, created_by, created_at, updated_at
		FROM roles
		WHERE id = $1 AND tenant_id = $2`

	var role models.Role
	err := r.db.QueryRowContext(ctx, query, roleID, tenantID).Scan(
		&role.ID,
		&role.TenantID,
		&role.Name,
		&role.Description,
		&role.IsSystem,
		&role.IsDefault,
		&role.CreatedBy,
		&role.CreatedAt,
		&role.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.NotFoundf("role not found")
	}
	if err != nil {
		r.logger.Error("failed to get role", zap.Error(err))
		return nil, errors.New(errors.ErrCodeInternal, "failed to get role")
	}

	return &role, nil
}

// GetRoleByName retrieves a role by name
func (r *Repository) GetRoleByName(ctx context.Context, tenantID uuid.UUID, name string) (*models.Role, error) {
	query := `
		SELECT id, tenant_id, name, description, is_system,
			is_default, created_by, created_at, updated_at
		FROM roles
		WHERE name = $1 AND tenant_id = $2`

	var role models.Role
	err := r.db.QueryRowContext(ctx, query, name, tenantID).Scan(
		&role.ID,
		&role.TenantID,
		&role.Name,
		&role.Description,
		&role.IsSystem,
		&role.IsDefault,
		&role.CreatedBy,
		&role.CreatedAt,
		&role.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.NotFoundf("role not found")
	}
	if err != nil {
		r.logger.Error("failed to get role by name", zap.Error(err))
		return nil, errors.New(errors.ErrCodeInternal, "failed to get role")
	}

	return &role, nil
}

// ListRoles retrieves roles with filtering
func (r *Repository) ListRoles(ctx context.Context, tenantID uuid.UUID, params *models.ListRolesParams) ([]models.Role, int64, error) {
	// Build WHERE clause
	where := []string{"tenant_id = $1"}
	args := []interface{}{tenantID}
	argPos := 2

	if params.IsSystem != "" {
		isSystem := params.IsSystem == "true"
		where = append(where, fmt.Sprintf("is_system = $%d", argPos))
		args = append(args, isSystem)
		argPos++
	}

	if params.IsDefault != "" {
		isDefault := params.IsDefault == "true"
		where = append(where, fmt.Sprintf("is_default = $%d", argPos))
		args = append(args, isDefault)
		argPos++
	}

	whereClause := strings.Join(where, " AND ")

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM roles WHERE %s", whereClause)
	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		r.logger.Error("failed to count roles", zap.Error(err))
		return nil, 0, errors.New(errors.ErrCodeInternal, "failed to count roles")
	}

	// Get roles
	query := fmt.Sprintf(`
		SELECT id, tenant_id, name, description, is_system,
			is_default, created_by, created_at, updated_at
		FROM roles
		WHERE %s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d`,
		whereClause,
		params.SortBy,
		params.SortOrder,
		argPos,
		argPos+1,
	)

	args = append(args, params.Limit, params.GetOffset())

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("failed to list roles", zap.Error(err))
		return nil, 0, errors.New(errors.ErrCodeInternal, "failed to list roles")
	}
	defer rows.Close()

	var roles []models.Role
	for rows.Next() {
		var role models.Role
		err := rows.Scan(
			&role.ID,
			&role.TenantID,
			&role.Name,
			&role.Description,
			&role.IsSystem,
			&role.IsDefault,
			&role.CreatedBy,
			&role.CreatedAt,
			&role.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("failed to scan role", zap.Error(err))
			continue
		}
		roles = append(roles, role)
	}

	return roles, total, nil
}

// UpdateRole updates a role
func (r *Repository) UpdateRole(ctx context.Context, tenantID, roleID uuid.UUID, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	// Build SET clause
	setClauses := []string{}
	args := []interface{}{}
	argPos := 1

	for key, value := range updates {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", key, argPos))
		args = append(args, value)
		argPos++
	}

	// Add updated_at
	setClauses = append(setClauses, fmt.Sprintf("updated_at = $%d", argPos))
	args = append(args, time.Now())
	argPos++

	// Add WHERE conditions
	args = append(args, roleID, tenantID)

	query := fmt.Sprintf(`
		UPDATE roles
		SET %s
		WHERE id = $%d AND tenant_id = $%d`,
		strings.Join(setClauses, ", "),
		argPos,
		argPos+1,
	)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("failed to update role", zap.Error(err))
		return errors.New(errors.ErrCodeInternal, "failed to update role")
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.NotFoundf("role not found")
	}

	return nil
}

// DeleteRole deletes a role
func (r *Repository) DeleteRole(ctx context.Context, tenantID, roleID uuid.UUID) error {
	query := `DELETE FROM roles WHERE id = $1 AND tenant_id = $2 AND is_system = false`

	result, err := r.db.ExecContext(ctx, query, roleID, tenantID)
	if err != nil {
		r.logger.Error("failed to delete role", zap.Error(err))
		return errors.New(errors.ErrCodeInternal, "failed to delete role")
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.NotFoundf("role not found or is a system role")
	}

	return nil
}

// Permission operations

// CreatePermission creates a new permission
func (r *Repository) CreatePermission(ctx context.Context, permission *models.Permission) error {
	query := `
		INSERT INTO permissions (id, name, resource, action, description, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := r.db.ExecContext(ctx, query,
		permission.ID,
		permission.Name,
		permission.Resource,
		permission.Action,
		permission.Description,
		permission.CreatedAt,
	)

	if err != nil {
		r.logger.Error("failed to create permission", zap.Error(err))
		return errors.New(errors.ErrCodeInternal, "failed to create permission")
	}

	return nil
}

// GetPermission retrieves a permission by ID
func (r *Repository) GetPermission(ctx context.Context, permissionID uuid.UUID) (*models.Permission, error) {
	query := `
		SELECT id, name, resource, action, description, created_at
		FROM permissions
		WHERE id = $1`

	var perm models.Permission
	err := r.db.QueryRowContext(ctx, query, permissionID).Scan(
		&perm.ID,
		&perm.Name,
		&perm.Resource,
		&perm.Action,
		&perm.Description,
		&perm.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.NotFoundf("permission not found")
	}
	if err != nil {
		r.logger.Error("failed to get permission", zap.Error(err))
		return nil, errors.New(errors.ErrCodeInternal, "failed to get permission")
	}

	return &perm, nil
}

// ListPermissions retrieves permissions with filtering
func (r *Repository) ListPermissions(ctx context.Context, params *models.ListPermissionsParams) ([]models.Permission, int64, error) {
	// Build WHERE clause
	where := []string{}
	args := []interface{}{}
	argPos := 1

	if params.Resource != "" {
		where = append(where, fmt.Sprintf("resource = $%d", argPos))
		args = append(args, params.Resource)
		argPos++
	}

	if params.Action != "" {
		where = append(where, fmt.Sprintf("action = $%d", argPos))
		args = append(args, params.Action)
		argPos++
	}

	whereClause := "TRUE"
	if len(where) > 0 {
		whereClause = strings.Join(where, " AND ")
	}

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM permissions WHERE %s", whereClause)
	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		r.logger.Error("failed to count permissions", zap.Error(err))
		return nil, 0, errors.New(errors.ErrCodeInternal, "failed to count permissions")
	}

	// Get permissions
	query := fmt.Sprintf(`
		SELECT id, name, resource, action, description, created_at
		FROM permissions
		WHERE %s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d`,
		whereClause,
		params.SortBy,
		params.SortOrder,
		argPos,
		argPos+1,
	)

	args = append(args, params.Limit, params.GetOffset())

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("failed to list permissions", zap.Error(err))
		return nil, 0, errors.New(errors.ErrCodeInternal, "failed to list permissions")
	}
	defer rows.Close()

	var permissions []models.Permission
	for rows.Next() {
		var perm models.Permission
		err := rows.Scan(
			&perm.ID,
			&perm.Name,
			&perm.Resource,
			&perm.Action,
			&perm.Description,
			&perm.CreatedAt,
		)
		if err != nil {
			r.logger.Error("failed to scan permission", zap.Error(err))
			continue
		}
		permissions = append(permissions, perm)
	}

	return permissions, total, nil
}

// Role-Permission operations

// AssignPermissionsToRole assigns permissions to a role
func (r *Repository) AssignPermissionsToRole(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error {
	// First, remove existing permissions
	deleteQuery := `DELETE FROM role_permissions WHERE role_id = $1`
	_, err := r.db.ExecContext(ctx, deleteQuery, roleID)
	if err != nil {
		r.logger.Error("failed to delete existing permissions", zap.Error(err))
		return errors.New(errors.ErrCodeInternal, "failed to update permissions")
	}

	// Then, add new permissions
	if len(permissionIDs) > 0 {
		query := `INSERT INTO role_permissions (role_id, permission_id, created_at) VALUES ($1, $2, $3)`
		for _, permID := range permissionIDs {
			_, err := r.db.ExecContext(ctx, query, roleID, permID, time.Now())
			if err != nil {
				r.logger.Error("failed to assign permission", zap.Error(err))
				continue
			}
		}
	}

	return nil
}

// GetRolePermissions retrieves all permissions for a role
func (r *Repository) GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]models.Permission, error) {
	query := `
		SELECT p.id, p.name, p.resource, p.action, p.description, p.created_at
		FROM permissions p
		INNER JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = $1
		ORDER BY p.resource, p.action`

	rows, err := r.db.QueryContext(ctx, query, roleID)
	if err != nil {
		r.logger.Error("failed to get role permissions", zap.Error(err))
		return nil, errors.New(errors.ErrCodeInternal, "failed to get permissions")
	}
	defer rows.Close()

	var permissions []models.Permission
	for rows.Next() {
		var perm models.Permission
		err := rows.Scan(
			&perm.ID,
			&perm.Name,
			&perm.Resource,
			&perm.Action,
			&perm.Description,
			&perm.CreatedAt,
		)
		if err != nil {
			r.logger.Error("failed to scan permission", zap.Error(err))
			continue
		}
		permissions = append(permissions, perm)
	}

	return permissions, nil
}

// User Role operations

// AssignRoleToUser assigns a role to a user
func (r *Repository) AssignRoleToUser(ctx context.Context, userRole *models.UserRole) error {
	query := `
		INSERT INTO user_roles (id, tenant_id, user_id, role_id, assigned_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := r.db.ExecContext(ctx, query,
		userRole.ID,
		userRole.TenantID,
		userRole.UserID,
		userRole.RoleID,
		userRole.AssignedBy,
		userRole.CreatedAt,
	)

	if err != nil {
		r.logger.Error("failed to assign role to user", zap.Error(err))
		return errors.New(errors.ErrCodeInternal, "failed to assign role")
	}

	return nil
}

// RemoveRoleFromUser removes a role from a user
func (r *Repository) RemoveRoleFromUser(ctx context.Context, tenantID uuid.UUID, userID string, roleID uuid.UUID) error {
	query := `DELETE FROM user_roles WHERE tenant_id = $1 AND user_id = $2 AND role_id = $3`

	result, err := r.db.ExecContext(ctx, query, tenantID, userID, roleID)
	if err != nil {
		r.logger.Error("failed to remove role from user", zap.Error(err))
		return errors.New(errors.ErrCodeInternal, "failed to remove role")
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.NotFoundf("user role assignment not found")
	}

	return nil
}

// GetUserRoles retrieves all roles for a user
func (r *Repository) GetUserRoles(ctx context.Context, tenantID uuid.UUID, userID string) ([]models.Role, error) {
	query := `
		SELECT r.id, r.tenant_id, r.name, r.description, r.is_system,
			r.is_default, r.created_by, r.created_at, r.updated_at
		FROM roles r
		INNER JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.tenant_id = $1 AND ur.user_id = $2
		ORDER BY r.name`

	rows, err := r.db.QueryContext(ctx, query, tenantID, userID)
	if err != nil {
		r.logger.Error("failed to get user roles", zap.Error(err))
		return nil, errors.New(errors.ErrCodeInternal, "failed to get user roles")
	}
	defer rows.Close()

	var roles []models.Role
	for rows.Next() {
		var role models.Role
		err := rows.Scan(
			&role.ID,
			&role.TenantID,
			&role.Name,
			&role.Description,
			&role.IsSystem,
			&role.IsDefault,
			&role.CreatedBy,
			&role.CreatedAt,
			&role.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("failed to scan role", zap.Error(err))
			continue
		}
		roles = append(roles, role)
	}

	return roles, nil
}

// GetUserPermissions retrieves all permissions for a user (via their roles)
func (r *Repository) GetUserPermissions(ctx context.Context, tenantID uuid.UUID, userID string) ([]models.Permission, error) {
	query := `
		SELECT DISTINCT p.id, p.name, p.resource, p.action, p.description, p.created_at
		FROM permissions p
		INNER JOIN role_permissions rp ON p.id = rp.permission_id
		INNER JOIN user_roles ur ON rp.role_id = ur.role_id
		WHERE ur.tenant_id = $1 AND ur.user_id = $2
		ORDER BY p.resource, p.action`

	rows, err := r.db.QueryContext(ctx, query, tenantID, userID)
	if err != nil {
		r.logger.Error("failed to get user permissions", zap.Error(err))
		return nil, errors.New(errors.ErrCodeInternal, "failed to get user permissions")
	}
	defer rows.Close()

	var permissions []models.Permission
	for rows.Next() {
		var perm models.Permission
		err := rows.Scan(
			&perm.ID,
			&perm.Name,
			&perm.Resource,
			&perm.Action,
			&perm.Description,
			&perm.CreatedAt,
		)
		if err != nil {
			r.logger.Error("failed to scan permission", zap.Error(err))
			continue
		}
		permissions = append(permissions, perm)
	}

	return permissions, nil
}

// CheckUserPermission checks if a user has a specific permission
func (r *Repository) CheckUserPermission(ctx context.Context, tenantID uuid.UUID, userID, resource, action string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1
			FROM permissions p
			INNER JOIN role_permissions rp ON p.id = rp.permission_id
			INNER JOIN user_roles ur ON rp.role_id = ur.role_id
			WHERE ur.tenant_id = $1
				AND ur.user_id = $2
				AND p.resource = $3
				AND p.action = $4
		)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, tenantID, userID, resource, action).Scan(&exists)
	if err != nil {
		r.logger.Error("failed to check user permission", zap.Error(err))
		return false, errors.New(errors.ErrCodeInternal, "failed to check permission")
	}

	return exists, nil
}

// GetRBACStats retrieves RBAC statistics for a tenant
func (r *Repository) GetRBACStats(ctx context.Context, tenantID uuid.UUID) (*models.RBACStats, error) {
	stats := &models.RBACStats{
		RoleDistribution: make(map[string]int64),
	}

	// Get role counts
	roleQuery := `
		SELECT
			COUNT(*) as total_roles,
			COUNT(*) FILTER (WHERE is_system = true) as system_roles,
			COUNT(*) FILTER (WHERE is_system = false) as custom_roles
		FROM roles
		WHERE tenant_id = $1`

	err := r.db.QueryRowContext(ctx, roleQuery, tenantID).Scan(
		&stats.TotalRoles,
		&stats.SystemRoles,
		&stats.CustomRoles,
	)
	if err != nil {
		r.logger.Error("failed to get role stats", zap.Error(err))
		return nil, errors.New(errors.ErrCodeInternal, "failed to get stats")
	}

	// Get total permissions
	permQuery := `SELECT COUNT(*) FROM permissions`
	err = r.db.QueryRowContext(ctx, permQuery).Scan(&stats.TotalPermissions)
	if err != nil {
		r.logger.Error("failed to get permission count", zap.Error(err))
	}

	// Get total user roles
	userRoleQuery := `SELECT COUNT(*) FROM user_roles WHERE tenant_id = $1`
	err = r.db.QueryRowContext(ctx, userRoleQuery, tenantID).Scan(&stats.TotalUserRoles)
	if err != nil {
		r.logger.Error("failed to get user role count", zap.Error(err))
	}

	// Get role distribution
	distQuery := `
		SELECT r.name, COUNT(ur.id) as count
		FROM roles r
		LEFT JOIN user_roles ur ON r.id = ur.role_id AND ur.tenant_id = $1
		WHERE r.tenant_id = $1
		GROUP BY r.name`

	rows, err := r.db.QueryContext(ctx, distQuery, tenantID)
	if err != nil {
		r.logger.Error("failed to get role distribution", zap.Error(err))
		return stats, nil
	}
	defer rows.Close()

	for rows.Next() {
		var roleName string
		var count int64
		if err := rows.Scan(&roleName, &count); err != nil {
			continue
		}
		stats.RoleDistribution[roleName] = count
	}

	return stats, nil
}

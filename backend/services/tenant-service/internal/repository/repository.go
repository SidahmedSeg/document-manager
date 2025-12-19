package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/SidahmedSeg/document-manager/backend/pkg/database"
	"github.com/SidahmedSeg/document-manager/backend/pkg/errors"
	"github.com/SidahmedSeg/document-manager/backend/services/tenant-service/internal/models"
	"go.uber.org/zap"
)

// Repository handles database operations for tenants
type Repository struct {
	db     *database.DB
	logger *zap.Logger
}

// NewRepository creates a new tenant repository
func NewRepository(db *database.DB, logger *zap.Logger) *Repository {
	return &Repository{
		db:     db,
		logger: logger,
	}
}

// CreateTenant creates a new tenant
func (r *Repository) CreateTenant(ctx context.Context, tenant *models.Tenant) error {
	query := `
		INSERT INTO tenants (id, name, slug, domain, subscription_plan, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(ctx, query,
		tenant.ID,
		tenant.Name,
		tenant.Slug,
		tenant.Domain,
		tenant.SubscriptionPlan,
		tenant.IsActive,
		tenant.CreatedAt,
		tenant.UpdatedAt,
	)

	if err != nil {
		r.logger.Error("failed to create tenant", zap.Error(err))
		return errors.Wrap(errors.ErrCodeDatabase, "failed to create tenant", err)
	}

	return nil
}

// GetTenantByID retrieves a tenant by ID
func (r *Repository) GetTenantByID(ctx context.Context, id uuid.UUID) (*models.Tenant, error) {
	query := `
		SELECT id, name, slug, domain, subscription_plan, is_active, created_at, updated_at
		FROM tenants
		WHERE id = $1
	`

	var tenant models.Tenant
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&tenant.ID,
		&tenant.Name,
		&tenant.Slug,
		&tenant.Domain,
		&tenant.SubscriptionPlan,
		&tenant.IsActive,
		&tenant.CreatedAt,
		&tenant.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.NotFoundf("tenant not found")
	}
	if err != nil {
		r.logger.Error("failed to get tenant", zap.Error(err))
		return nil, errors.Wrap(errors.ErrCodeDatabase, "failed to get tenant", err)
	}

	return &tenant, nil
}

// GetTenantBySlug retrieves a tenant by slug
func (r *Repository) GetTenantBySlug(ctx context.Context, slug string) (*models.Tenant, error) {
	query := `
		SELECT id, name, slug, domain, subscription_plan, is_active, created_at, updated_at
		FROM tenants
		WHERE slug = $1
	`

	var tenant models.Tenant
	err := r.db.QueryRowContext(ctx, query, slug).Scan(
		&tenant.ID,
		&tenant.Name,
		&tenant.Slug,
		&tenant.Domain,
		&tenant.SubscriptionPlan,
		&tenant.IsActive,
		&tenant.CreatedAt,
		&tenant.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.NotFoundf("tenant not found")
	}
	if err != nil {
		r.logger.Error("failed to get tenant by slug", zap.Error(err))
		return nil, errors.Wrap(errors.ErrCodeDatabase, "failed to get tenant", err)
	}

	return &tenant, nil
}

// UpdateTenant updates a tenant
func (r *Repository) UpdateTenant(ctx context.Context, id uuid.UUID, req *models.UpdateTenantRequest) error {
	query := `
		UPDATE tenants
		SET name = COALESCE(NULLIF($1, ''), name),
		    domain = COALESCE(NULLIF($2, ''), domain),
		    is_active = COALESCE($3, is_active),
		    updated_at = $4
		WHERE id = $5
	`

	_, err := r.db.ExecContext(ctx, query,
		req.Name,
		req.Domain,
		req.IsActive,
		time.Now(),
		id,
	)

	if err != nil {
		r.logger.Error("failed to update tenant", zap.Error(err))
		return errors.Wrap(errors.ErrCodeDatabase, "failed to update tenant", err)
	}

	return nil
}

// AddTenantUser adds a user to a tenant
func (r *Repository) AddTenantUser(ctx context.Context, tu *models.TenantUser) error {
	query := `
		INSERT INTO tenant_users (id, tenant_id, user_id, user_email, role, is_owner, joined_at, invited_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(ctx, query,
		tu.ID,
		tu.TenantID,
		tu.UserID,
		tu.UserEmail,
		tu.Role,
		tu.IsOwner,
		tu.JoinedAt,
		tu.InvitedBy,
	)

	if err != nil {
		r.logger.Error("failed to add tenant user", zap.Error(err))
		return errors.Wrap(errors.ErrCodeDatabase, "failed to add user to tenant", err)
	}

	return nil
}

// GetTenantUsers retrieves all users in a tenant
func (r *Repository) GetTenantUsers(ctx context.Context, tenantID uuid.UUID) ([]models.TenantUser, error) {
	query := `
		SELECT id, tenant_id, user_id, user_email, role, is_owner, joined_at, invited_by
		FROM tenant_users
		WHERE tenant_id = $1
		ORDER BY joined_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		r.logger.Error("failed to get tenant users", zap.Error(err))
		return nil, errors.Wrap(errors.ErrCodeDatabase, "failed to get tenant users", err)
	}
	defer rows.Close()

	var users []models.TenantUser
	for rows.Next() {
		var user models.TenantUser
		err := rows.Scan(
			&user.ID,
			&user.TenantID,
			&user.UserID,
			&user.UserEmail,
			&user.Role,
			&user.IsOwner,
			&user.JoinedAt,
			&user.InvitedBy,
		)
		if err != nil {
			r.logger.Error("failed to scan tenant user", zap.Error(err))
			continue
		}
		users = append(users, user)
	}

	return users, nil
}

// RemoveTenantUser removes a user from a tenant
func (r *Repository) RemoveTenantUser(ctx context.Context, tenantID uuid.UUID, userID string) error {
	query := `
		DELETE FROM tenant_users
		WHERE tenant_id = $1 AND user_id = $2 AND is_owner = false
	`

	result, err := r.db.ExecContext(ctx, query, tenantID, userID)
	if err != nil {
		r.logger.Error("failed to remove tenant user", zap.Error(err))
		return errors.Wrap(errors.ErrCodeDatabase, "failed to remove user from tenant", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.Forbiddenf("cannot remove owner or user not found")
	}

	return nil
}

// CreateInvitation creates a new tenant invitation
func (r *Repository) CreateInvitation(ctx context.Context, inv *models.TenantInvitation) error {
	query := `
		INSERT INTO tenant_invitations (id, tenant_id, email, role, invited_by, token, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(ctx, query,
		inv.ID,
		inv.TenantID,
		inv.Email,
		inv.Role,
		inv.InvitedBy,
		inv.Token,
		inv.ExpiresAt,
		inv.CreatedAt,
	)

	if err != nil {
		r.logger.Error("failed to create invitation", zap.Error(err))
		return errors.Wrap(errors.ErrCodeDatabase, "failed to create invitation", err)
	}

	return nil
}

// GetPendingInvitations retrieves pending invitations for a tenant
func (r *Repository) GetPendingInvitations(ctx context.Context, tenantID uuid.UUID) ([]models.TenantInvitation, error) {
	query := `
		SELECT id, tenant_id, email, role, invited_by, expires_at, created_at
		FROM tenant_invitations
		WHERE tenant_id = $1 AND accepted_at IS NULL AND expires_at > NOW()
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		r.logger.Error("failed to get pending invitations", zap.Error(err))
		return nil, errors.Wrap(errors.ErrCodeDatabase, "failed to get invitations", err)
	}
	defer rows.Close()

	var invitations []models.TenantInvitation
	for rows.Next() {
		var inv models.TenantInvitation
		err := rows.Scan(
			&inv.ID,
			&inv.TenantID,
			&inv.Email,
			&inv.Role,
			&inv.InvitedBy,
			&inv.ExpiresAt,
			&inv.CreatedAt,
		)
		if err != nil {
			r.logger.Error("failed to scan invitation", zap.Error(err))
			continue
		}
		invitations = append(invitations, inv)
	}

	return invitations, nil
}

// GetUserTenants retrieves all tenants a user belongs to
func (r *Repository) GetUserTenants(ctx context.Context, userID string) ([]models.Tenant, error) {
	query := `
		SELECT t.id, t.name, t.slug, t.domain, t.subscription_plan, t.is_active, t.created_at, t.updated_at
		FROM tenants t
		INNER JOIN tenant_users tu ON t.id = tu.tenant_id
		WHERE tu.user_id = $1 AND t.is_active = true
		ORDER BY tu.joined_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		r.logger.Error("failed to get user tenants", zap.Error(err))
		return nil, errors.Wrap(errors.ErrCodeDatabase, "failed to get user tenants", err)
	}
	defer rows.Close()

	var tenants []models.Tenant
	for rows.Next() {
		var tenant models.Tenant
		err := rows.Scan(
			&tenant.ID,
			&tenant.Name,
			&tenant.Slug,
			&tenant.Domain,
			&tenant.SubscriptionPlan,
			&tenant.IsActive,
			&tenant.CreatedAt,
			&tenant.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("failed to scan tenant", zap.Error(err))
			continue
		}
		tenants = append(tenants, tenant)
	}

	return tenants, nil
}

// IsUserInTenant checks if a user belongs to a tenant
func (r *Repository) IsUserInTenant(ctx context.Context, tenantID uuid.UUID, userID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM tenant_users WHERE tenant_id = $1 AND user_id = $2)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, tenantID, userID).Scan(&exists)
	if err != nil {
		return false, errors.Wrap(errors.ErrCodeDatabase, "failed to check user membership", err)
	}

	return exists, nil
}

// GetUserRole retrieves a user's role in a tenant
func (r *Repository) GetUserRole(ctx context.Context, tenantID uuid.UUID, userID string) (string, error) {
	query := `SELECT role FROM tenant_users WHERE tenant_id = $1 AND user_id = $2`

	var role string
	err := r.db.QueryRowContext(ctx, query, tenantID, userID).Scan(&role)
	if err == sql.ErrNoRows {
		return "", errors.NotFoundf("user not found in tenant")
	}
	if err != nil {
		return "", errors.Wrap(errors.ErrCodeDatabase, "failed to get user role", err)
	}

	return role, nil
}

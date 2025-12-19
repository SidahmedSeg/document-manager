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
	"github.com/SidahmedSeg/document-manager/backend/services/share-service/internal/models"
	"go.uber.org/zap"
)

// Repository handles share database operations
type Repository struct {
	db     *database.DB
	logger *zap.Logger
}

// NewRepository creates a new share repository
func NewRepository(db *database.DB, logger *zap.Logger) *Repository {
	return &Repository{
		db:     db,
		logger: logger,
	}
}

// CreateShare creates a new share
func (r *Repository) CreateShare(ctx context.Context, share *models.Share) error {
	query := `
		INSERT INTO shares (
			id, tenant_id, document_id, share_type, shared_by,
			shared_with, permission, share_token, expires_at,
			password, max_access, access_count, is_active,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		)`

	_, err := r.db.ExecContext(ctx, query,
		share.ID,
		share.TenantID,
		share.DocumentID,
		share.ShareType,
		share.SharedBy,
		share.SharedWith,
		share.Permission,
		share.ShareToken,
		share.ExpiresAt,
		share.Password,
		share.MaxAccess,
		share.AccessCount,
		share.IsActive,
		share.CreatedAt,
		share.UpdatedAt,
	)

	if err != nil {
		r.logger.Error("failed to create share", zap.Error(err))
		return errors.New(errors.ErrCodeInternal, "failed to create share")
	}

	return nil
}

// GetShare retrieves a share by ID
func (r *Repository) GetShare(ctx context.Context, tenantID, shareID uuid.UUID) (*models.Share, error) {
	query := `
		SELECT id, tenant_id, document_id, share_type, shared_by,
			shared_with, permission, share_token, expires_at,
			password, max_access, access_count, is_active,
			created_at, updated_at
		FROM shares
		WHERE id = $1 AND tenant_id = $2`

	var share models.Share
	err := r.db.QueryRowContext(ctx, query, shareID, tenantID).Scan(
		&share.ID,
		&share.TenantID,
		&share.DocumentID,
		&share.ShareType,
		&share.SharedBy,
		&share.SharedWith,
		&share.Permission,
		&share.ShareToken,
		&share.ExpiresAt,
		&share.Password,
		&share.MaxAccess,
		&share.AccessCount,
		&share.IsActive,
		&share.CreatedAt,
		&share.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.NotFoundf("share not found")
	}
	if err != nil {
		r.logger.Error("failed to get share", zap.Error(err))
		return nil, errors.New(errors.ErrCodeInternal, "failed to get share")
	}

	return &share, nil
}

// GetShareByToken retrieves a share by token
func (r *Repository) GetShareByToken(ctx context.Context, token string) (*models.Share, error) {
	query := `
		SELECT id, tenant_id, document_id, share_type, shared_by,
			shared_with, permission, share_token, expires_at,
			password, max_access, access_count, is_active,
			created_at, updated_at
		FROM shares
		WHERE share_token = $1`

	var share models.Share
	err := r.db.QueryRowContext(ctx, query, token).Scan(
		&share.ID,
		&share.TenantID,
		&share.DocumentID,
		&share.ShareType,
		&share.SharedBy,
		&share.SharedWith,
		&share.Permission,
		&share.ShareToken,
		&share.ExpiresAt,
		&share.Password,
		&share.MaxAccess,
		&share.AccessCount,
		&share.IsActive,
		&share.CreatedAt,
		&share.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.NotFoundf("share not found")
	}
	if err != nil {
		r.logger.Error("failed to get share by token", zap.Error(err))
		return nil, errors.New(errors.ErrCodeInternal, "failed to get share")
	}

	return &share, nil
}

// ListShares retrieves shares with filtering
func (r *Repository) ListShares(ctx context.Context, tenantID uuid.UUID, params *models.ListSharesParams) ([]models.Share, int64, error) {
	// Build WHERE clause
	where := []string{"tenant_id = $1"}
	args := []interface{}{tenantID}
	argPos := 2

	if params.DocumentID != "" {
		docID, err := uuid.Parse(params.DocumentID)
		if err == nil {
			where = append(where, fmt.Sprintf("document_id = $%d", argPos))
			args = append(args, docID)
			argPos++
		}
	}

	if params.ShareType != "" {
		where = append(where, fmt.Sprintf("share_type = $%d", argPos))
		args = append(args, params.ShareType)
		argPos++
	}

	if params.SharedWith != "" {
		where = append(where, fmt.Sprintf("shared_with = $%d", argPos))
		args = append(args, params.SharedWith)
		argPos++
	}

	if params.IsActive != "" {
		isActive := params.IsActive == "true"
		where = append(where, fmt.Sprintf("is_active = $%d", argPos))
		args = append(args, isActive)
		argPos++
	}

	whereClause := strings.Join(where, " AND ")

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM shares WHERE %s", whereClause)
	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		r.logger.Error("failed to count shares", zap.Error(err))
		return nil, 0, errors.New(errors.ErrCodeInternal, "failed to count shares")
	}

	// Get shares
	query := fmt.Sprintf(`
		SELECT id, tenant_id, document_id, share_type, shared_by,
			shared_with, permission, share_token, expires_at,
			password, max_access, access_count, is_active,
			created_at, updated_at
		FROM shares
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
		r.logger.Error("failed to list shares", zap.Error(err))
		return nil, 0, errors.New(errors.ErrCodeInternal, "failed to list shares")
	}
	defer rows.Close()

	var shares []models.Share
	for rows.Next() {
		var share models.Share
		err := rows.Scan(
			&share.ID,
			&share.TenantID,
			&share.DocumentID,
			&share.ShareType,
			&share.SharedBy,
			&share.SharedWith,
			&share.Permission,
			&share.ShareToken,
			&share.ExpiresAt,
			&share.Password,
			&share.MaxAccess,
			&share.AccessCount,
			&share.IsActive,
			&share.CreatedAt,
			&share.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("failed to scan share", zap.Error(err))
			continue
		}
		shares = append(shares, share)
	}

	return shares, total, nil
}

// UpdateShare updates a share
func (r *Repository) UpdateShare(ctx context.Context, tenantID, shareID uuid.UUID, updates map[string]interface{}) error {
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
	args = append(args, shareID, tenantID)

	query := fmt.Sprintf(`
		UPDATE shares
		SET %s
		WHERE id = $%d AND tenant_id = $%d`,
		strings.Join(setClauses, ", "),
		argPos,
		argPos+1,
	)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("failed to update share", zap.Error(err))
		return errors.New(errors.ErrCodeInternal, "failed to update share")
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.NotFoundf("share not found")
	}

	return nil
}

// DeleteShare deletes a share
func (r *Repository) DeleteShare(ctx context.Context, tenantID, shareID uuid.UUID) error {
	query := `DELETE FROM shares WHERE id = $1 AND tenant_id = $2`

	result, err := r.db.ExecContext(ctx, query, shareID, tenantID)
	if err != nil {
		r.logger.Error("failed to delete share", zap.Error(err))
		return errors.New(errors.ErrCodeInternal, "failed to delete share")
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.NotFoundf("share not found")
	}

	return nil
}

// IncrementAccessCount increments the access count for a share
func (r *Repository) IncrementAccessCount(ctx context.Context, shareID uuid.UUID) error {
	query := `
		UPDATE shares
		SET access_count = access_count + 1, updated_at = $1
		WHERE id = $2`

	_, err := r.db.ExecContext(ctx, query, time.Now(), shareID)
	if err != nil {
		r.logger.Error("failed to increment access count", zap.Error(err))
		return errors.New(errors.ErrCodeInternal, "failed to update access count")
	}

	return nil
}

// CreateShareAccess logs share access
func (r *Repository) CreateShareAccess(ctx context.Context, access *models.ShareAccess) error {
	query := `
		INSERT INTO share_access (
			id, share_id, accessed_by, ip_address,
			user_agent, action, accessed_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.db.ExecContext(ctx, query,
		access.ID,
		access.ShareID,
		access.AccessedBy,
		access.IPAddress,
		access.UserAgent,
		access.Action,
		access.AccessedAt,
	)

	if err != nil {
		r.logger.Error("failed to create share access log", zap.Error(err))
		return errors.New(errors.ErrCodeInternal, "failed to log access")
	}

	return nil
}

// GetShareAccessLogs retrieves access logs for a share
func (r *Repository) GetShareAccessLogs(ctx context.Context, shareID uuid.UUID, limit int) ([]models.ShareAccess, error) {
	query := `
		SELECT id, share_id, accessed_by, ip_address,
			user_agent, action, accessed_at
		FROM share_access
		WHERE share_id = $1
		ORDER BY accessed_at DESC
		LIMIT $2`

	rows, err := r.db.QueryContext(ctx, query, shareID, limit)
	if err != nil {
		r.logger.Error("failed to get share access logs", zap.Error(err))
		return nil, errors.New(errors.ErrCodeInternal, "failed to get access logs")
	}
	defer rows.Close()

	var logs []models.ShareAccess
	for rows.Next() {
		var log models.ShareAccess
		err := rows.Scan(
			&log.ID,
			&log.ShareID,
			&log.AccessedBy,
			&log.IPAddress,
			&log.UserAgent,
			&log.Action,
			&log.AccessedAt,
		)
		if err != nil {
			r.logger.Error("failed to scan access log", zap.Error(err))
			continue
		}
		logs = append(logs, log)
	}

	return logs, nil
}

// GetShareStats retrieves share statistics for a tenant
func (r *Repository) GetShareStats(ctx context.Context, tenantID uuid.UUID) (*models.ShareStats, error) {
	stats := &models.ShareStats{
		SharesByType:       make(map[string]int64),
		SharesByPermission: make(map[string]int64),
	}

	// Get overall stats
	query := `
		SELECT
			COUNT(*) as total_shares,
			COUNT(*) FILTER (WHERE is_active = true) as active_shares,
			COUNT(*) FILTER (WHERE expires_at IS NOT NULL AND expires_at < NOW()) as expired_shares,
			COALESCE(SUM(access_count), 0) as total_access
		FROM shares
		WHERE tenant_id = $1`

	err := r.db.QueryRowContext(ctx, query, tenantID).Scan(
		&stats.TotalShares,
		&stats.ActiveShares,
		&stats.ExpiredShares,
		&stats.TotalAccess,
	)
	if err != nil {
		r.logger.Error("failed to get share stats", zap.Error(err))
		return nil, errors.New(errors.ErrCodeInternal, "failed to get share stats")
	}

	// Get stats by type
	typeQuery := `
		SELECT share_type, COUNT(*) as count
		FROM shares
		WHERE tenant_id = $1
		GROUP BY share_type`

	rows, err := r.db.QueryContext(ctx, typeQuery, tenantID)
	if err != nil {
		r.logger.Error("failed to get share type stats", zap.Error(err))
		return stats, nil
	}
	defer rows.Close()

	for rows.Next() {
		var shareType string
		var count int64
		if err := rows.Scan(&shareType, &count); err != nil {
			continue
		}
		stats.SharesByType[shareType] = count
	}

	// Get stats by permission
	permQuery := `
		SELECT permission, COUNT(*) as count
		FROM shares
		WHERE tenant_id = $1
		GROUP BY permission`

	rows2, err := r.db.QueryContext(ctx, permQuery, tenantID)
	if err != nil {
		r.logger.Error("failed to get share permission stats", zap.Error(err))
		return stats, nil
	}
	defer rows2.Close()

	for rows2.Next() {
		var permission string
		var count int64
		if err := rows2.Scan(&permission, &count); err != nil {
			continue
		}
		stats.SharesByPermission[permission] = count
	}

	return stats, nil
}

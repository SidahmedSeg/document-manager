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
	"github.com/SidahmedSeg/document-manager/backend/services/quota-service/internal/models"
	"go.uber.org/zap"
)

// Repository handles quota database operations
type Repository struct {
	db     *database.DB
	logger *zap.Logger
}

// NewRepository creates a new quota repository
func NewRepository(db *database.DB, logger *zap.Logger) *Repository {
	return &Repository{
		db:     db,
		logger: logger,
	}
}

// Quota operations

// CreateQuota creates a new quota
func (r *Repository) CreateQuota(ctx context.Context, quota *models.Quota) error {
	query := `
		INSERT INTO quotas (
			id, tenant_id, plan_name, max_storage, max_documents,
			max_users, max_api_calls_per_day, max_file_size, max_bandwidth,
			features, is_active, valid_from, valid_until, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`

	_, err := r.db.ExecContext(ctx, query,
		quota.ID,
		quota.TenantID,
		quota.PlanName,
		quota.MaxStorage,
		quota.MaxDocuments,
		quota.MaxUsers,
		quota.MaxAPICallsPerDay,
		quota.MaxFileSize,
		quota.MaxBandwidth,
		quota.Features,
		quota.IsActive,
		quota.ValidFrom,
		quota.ValidUntil,
		quota.CreatedAt,
		quota.UpdatedAt,
	)

	if err != nil {
		r.logger.Error("failed to create quota", zap.Error(err))
		return errors.New(errors.ErrCodeInternal, "failed to create quota")
	}

	return nil
}

// GetQuota retrieves quota for a tenant
func (r *Repository) GetQuota(ctx context.Context, tenantID uuid.UUID) (*models.Quota, error) {
	query := `
		SELECT id, tenant_id, plan_name, max_storage, max_documents,
			max_users, max_api_calls_per_day, max_file_size, max_bandwidth,
			features, is_active, valid_from, valid_until, created_at, updated_at
		FROM quotas
		WHERE tenant_id = $1 AND is_active = true
		ORDER BY created_at DESC
		LIMIT 1`

	var quota models.Quota
	err := r.db.QueryRowContext(ctx, query, tenantID).Scan(
		&quota.ID,
		&quota.TenantID,
		&quota.PlanName,
		&quota.MaxStorage,
		&quota.MaxDocuments,
		&quota.MaxUsers,
		&quota.MaxAPICallsPerDay,
		&quota.MaxFileSize,
		&quota.MaxBandwidth,
		&quota.Features,
		&quota.IsActive,
		&quota.ValidFrom,
		&quota.ValidUntil,
		&quota.CreatedAt,
		&quota.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.NotFoundf("quota not found")
	}
	if err != nil {
		r.logger.Error("failed to get quota", zap.Error(err))
		return nil, errors.New(errors.ErrCodeInternal, "failed to get quota")
	}

	return &quota, nil
}

// UpdateQuota updates a quota
func (r *Repository) UpdateQuota(ctx context.Context, tenantID uuid.UUID, updates map[string]interface{}) error {
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

	// Add WHERE condition
	args = append(args, tenantID)

	query := fmt.Sprintf(`
		UPDATE quotas
		SET %s
		WHERE tenant_id = $%d AND is_active = true`,
		strings.Join(setClauses, ", "),
		argPos,
	)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("failed to update quota", zap.Error(err))
		return errors.New(errors.ErrCodeInternal, "failed to update quota")
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.NotFoundf("quota not found")
	}

	return nil
}

// Usage operations

// CreateUsage creates a new usage record
func (r *Repository) CreateUsage(ctx context.Context, usage *models.Usage) error {
	query := `
		INSERT INTO usage (
			id, tenant_id, storage_used, document_count, user_count,
			api_calls_today, bandwidth_month, last_api_call, last_reset_date, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := r.db.ExecContext(ctx, query,
		usage.ID,
		usage.TenantID,
		usage.StorageUsed,
		usage.DocumentCount,
		usage.UserCount,
		usage.APICallsToday,
		usage.BandwidthMonth,
		usage.LastAPICall,
		usage.LastResetDate,
		usage.UpdatedAt,
	)

	if err != nil {
		r.logger.Error("failed to create usage", zap.Error(err))
		return errors.New(errors.ErrCodeInternal, "failed to create usage")
	}

	return nil
}

// GetUsage retrieves usage for a tenant
func (r *Repository) GetUsage(ctx context.Context, tenantID uuid.UUID) (*models.Usage, error) {
	query := `
		SELECT id, tenant_id, storage_used, document_count, user_count,
			api_calls_today, bandwidth_month, last_api_call, last_reset_date, updated_at
		FROM usage
		WHERE tenant_id = $1`

	var usage models.Usage
	err := r.db.QueryRowContext(ctx, query, tenantID).Scan(
		&usage.ID,
		&usage.TenantID,
		&usage.StorageUsed,
		&usage.DocumentCount,
		&usage.UserCount,
		&usage.APICallsToday,
		&usage.BandwidthMonth,
		&usage.LastAPICall,
		&usage.LastResetDate,
		&usage.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.NotFoundf("usage not found")
	}
	if err != nil {
		r.logger.Error("failed to get usage", zap.Error(err))
		return nil, errors.New(errors.ErrCodeInternal, "failed to get usage")
	}

	return &usage, nil
}

// IncrementStorage increments storage usage
func (r *Repository) IncrementStorage(ctx context.Context, tenantID uuid.UUID, amount int64) error {
	query := `
		UPDATE usage
		SET storage_used = storage_used + $1, updated_at = $2
		WHERE tenant_id = $3`

	_, err := r.db.ExecContext(ctx, query, amount, time.Now(), tenantID)
	if err != nil {
		r.logger.Error("failed to increment storage", zap.Error(err))
		return errors.New(errors.ErrCodeInternal, "failed to update usage")
	}

	return nil
}

// DecrementStorage decrements storage usage
func (r *Repository) DecrementStorage(ctx context.Context, tenantID uuid.UUID, amount int64) error {
	query := `
		UPDATE usage
		SET storage_used = GREATEST(0, storage_used - $1), updated_at = $2
		WHERE tenant_id = $3`

	_, err := r.db.ExecContext(ctx, query, amount, time.Now(), tenantID)
	if err != nil {
		r.logger.Error("failed to decrement storage", zap.Error(err))
		return errors.New(errors.ErrCodeInternal, "failed to update usage")
	}

	return nil
}

// IncrementDocumentCount increments document count
func (r *Repository) IncrementDocumentCount(ctx context.Context, tenantID uuid.UUID, amount int) error {
	query := `
		UPDATE usage
		SET document_count = document_count + $1, updated_at = $2
		WHERE tenant_id = $3`

	_, err := r.db.ExecContext(ctx, query, amount, time.Now(), tenantID)
	if err != nil {
		r.logger.Error("failed to increment document count", zap.Error(err))
		return errors.New(errors.ErrCodeInternal, "failed to update usage")
	}

	return nil
}

// DecrementDocumentCount decrements document count
func (r *Repository) DecrementDocumentCount(ctx context.Context, tenantID uuid.UUID, amount int) error {
	query := `
		UPDATE usage
		SET document_count = GREATEST(0, document_count - $1), updated_at = $2
		WHERE tenant_id = $3`

	_, err := r.db.ExecContext(ctx, query, amount, time.Now(), tenantID)
	if err != nil {
		r.logger.Error("failed to decrement document count", zap.Error(err))
		return errors.New(errors.ErrCodeInternal, "failed to update usage")
	}

	return nil
}

// IncrementAPICallCount increments API call count
func (r *Repository) IncrementAPICallCount(ctx context.Context, tenantID uuid.UUID) error {
	query := `
		UPDATE usage
		SET api_calls_today = api_calls_today + 1, last_api_call = $1, updated_at = $2
		WHERE tenant_id = $3`

	now := time.Now()
	_, err := r.db.ExecContext(ctx, query, now, now, tenantID)
	if err != nil {
		r.logger.Error("failed to increment API call count", zap.Error(err))
		return errors.New(errors.ErrCodeInternal, "failed to update usage")
	}

	return nil
}

// IncrementBandwidth increments bandwidth usage
func (r *Repository) IncrementBandwidth(ctx context.Context, tenantID uuid.UUID, amount int64) error {
	query := `
		UPDATE usage
		SET bandwidth_month = bandwidth_month + $1, updated_at = $2
		WHERE tenant_id = $3`

	_, err := r.db.ExecContext(ctx, query, amount, time.Now(), tenantID)
	if err != nil {
		r.logger.Error("failed to increment bandwidth", zap.Error(err))
		return errors.New(errors.ErrCodeInternal, "failed to update usage")
	}

	return nil
}

// ResetDailyAPICallCount resets daily API call count
func (r *Repository) ResetDailyAPICallCount(ctx context.Context, tenantID uuid.UUID) error {
	query := `
		UPDATE usage
		SET api_calls_today = 0, last_reset_date = $1, updated_at = $2
		WHERE tenant_id = $3`

	now := time.Now()
	_, err := r.db.ExecContext(ctx, query, now, now, tenantID)
	if err != nil {
		r.logger.Error("failed to reset API call count", zap.Error(err))
		return errors.New(errors.ErrCodeInternal, "failed to reset usage")
	}

	return nil
}

// ResetMonthlyBandwidth resets monthly bandwidth
func (r *Repository) ResetMonthlyBandwidth(ctx context.Context, tenantID uuid.UUID) error {
	query := `
		UPDATE usage
		SET bandwidth_month = 0, updated_at = $1
		WHERE tenant_id = $2`

	_, err := r.db.ExecContext(ctx, query, time.Now(), tenantID)
	if err != nil {
		r.logger.Error("failed to reset bandwidth", zap.Error(err))
		return errors.New(errors.ErrCodeInternal, "failed to reset usage")
	}

	return nil
}

// Usage log operations

// CreateUsageLog creates a usage log entry
func (r *Repository) CreateUsageLog(ctx context.Context, log *models.UsageLog) error {
	query := `
		INSERT INTO usage_logs (id, tenant_id, user_id, action, resource, amount, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.db.ExecContext(ctx, query,
		log.ID,
		log.TenantID,
		log.UserID,
		log.Action,
		log.Resource,
		log.Amount,
		log.Metadata,
		log.CreatedAt,
	)

	if err != nil {
		r.logger.Error("failed to create usage log", zap.Error(err))
		return errors.New(errors.ErrCodeInternal, "failed to create usage log")
	}

	return nil
}

// GetUsageLogs retrieves usage logs for a tenant
func (r *Repository) GetUsageLogs(ctx context.Context, tenantID uuid.UUID, params *models.UsageStatsParams) ([]models.UsageLog, error) {
	// Build WHERE clause
	where := []string{"tenant_id = $1"}
	args := []interface{}{tenantID}
	argPos := 2

	// Parse dates
	if params.StartDate != "" {
		startTime, err := time.Parse(time.RFC3339, params.StartDate)
		if err == nil {
			where = append(where, fmt.Sprintf("created_at >= $%d", argPos))
			args = append(args, startTime)
			argPos++
		}
	}

	if params.EndDate != "" {
		endTime, err := time.Parse(time.RFC3339, params.EndDate)
		if err == nil {
			where = append(where, fmt.Sprintf("created_at <= $%d", argPos))
			args = append(args, endTime)
			argPos++
		}
	}

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

	whereClause := strings.Join(where, " AND ")

	query := fmt.Sprintf(`
		SELECT id, tenant_id, user_id, action, resource, amount, metadata, created_at
		FROM usage_logs
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d`,
		whereClause,
		argPos,
	)

	args = append(args, params.Limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("failed to get usage logs", zap.Error(err))
		return nil, errors.New(errors.ErrCodeInternal, "failed to get usage logs")
	}
	defer rows.Close()

	var logs []models.UsageLog
	for rows.Next() {
		var log models.UsageLog
		err := rows.Scan(
			&log.ID,
			&log.TenantID,
			&log.UserID,
			&log.Action,
			&log.Resource,
			&log.Amount,
			&log.Metadata,
			&log.CreatedAt,
		)
		if err != nil {
			r.logger.Error("failed to scan usage log", zap.Error(err))
			continue
		}
		logs = append(logs, log)
	}

	return logs, nil
}

// GetUsageStats retrieves aggregated usage statistics
func (r *Repository) GetUsageStats(ctx context.Context, tenantID uuid.UUID, params *models.UsageStatsParams) (*models.UsageStats, error) {
	stats := &models.UsageStats{
		TenantID:       tenantID,
		StorageByDay:   make(map[string]int64),
		APICallsByDay:  make(map[string]int),
		BandwidthByDay: make(map[string]int64),
	}

	// Parse dates
	startTime, _ := time.Parse(time.RFC3339, params.StartDate)
	endTime, _ := time.Parse(time.RFC3339, params.EndDate)

	stats.Period = fmt.Sprintf("%s to %s", startTime.Format("2006-01-02"), endTime.Format("2006-01-02"))

	// Get current totals
	usage, err := r.GetUsage(ctx, tenantID)
	if err == nil {
		stats.TotalStorage = usage.StorageUsed
		stats.TotalDocuments = usage.DocumentCount
		stats.TotalUsers = usage.UserCount
		stats.TotalAPICall = usage.APICallsToday
		stats.TotalBandwidth = usage.BandwidthMonth
	}

	return stats, nil
}

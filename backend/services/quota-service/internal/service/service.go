package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/SidahmedSeg/document-manager/backend/pkg/cache"
	"github.com/SidahmedSeg/document-manager/backend/pkg/errors"
	"github.com/SidahmedSeg/document-manager/backend/pkg/logger"
	"github.com/SidahmedSeg/document-manager/backend/pkg/middleware"
	"github.com/SidahmedSeg/document-manager/backend/services/quota-service/internal/models"
	"github.com/SidahmedSeg/document-manager/backend/services/quota-service/internal/repository"
	"go.uber.org/zap"
)

const (
	quotaCacheTTL = 1 * time.Hour
	usageCacheTTL = 5 * time.Minute
)

// Service handles quota business logic
type Service struct {
	repo   *repository.Repository
	cache  *cache.Cache
	logger *zap.Logger
}

// NewService creates a new quota service
func NewService(repo *repository.Repository, cache *cache.Cache, logger *zap.Logger) *Service {
	return &Service{
		repo:   repo,
		cache:  cache,
		logger: logger,
	}
}

// CreateQuota creates a new quota for a tenant
func (s *Service) CreateQuota(ctx context.Context, req *models.CreateQuotaRequest) (*models.Quota, error) {
	tenantID := getTenantID(ctx)

	// Parse valid_until if provided
	var validUntil *time.Time
	if req.ValidUntil != "" {
		parsed, err := time.Parse(time.RFC3339, req.ValidUntil)
		if err != nil {
			return nil, errors.Validationf("invalid valid_until format")
		}
		validUntil = &parsed
	}

	// Create quota
	quota := &models.Quota{
		ID:                uuid.New(),
		TenantID:          tenantID,
		PlanName:          req.PlanName,
		MaxStorage:        req.MaxStorage,
		MaxDocuments:      req.MaxDocuments,
		MaxUsers:          req.MaxUsers,
		MaxAPICallsPerDay: req.MaxAPICallsPerDay,
		MaxFileSize:       req.MaxFileSize,
		MaxBandwidth:      req.MaxBandwidth,
		IsActive:          true,
		ValidFrom:         time.Now(),
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if validUntil != nil {
		quota.ValidUntil.Time = *validUntil
		quota.ValidUntil.Valid = true
	}

	if len(req.Features) > 0 {
		featuresJSON, _ := json.Marshal(req.Features)
		quota.Features.String = string(featuresJSON)
		quota.Features.Valid = true
	}

	if err := s.repo.CreateQuota(ctx, quota); err != nil {
		return nil, err
	}

	// Create initial usage record
	usage := &models.Usage{
		ID:             uuid.New(),
		TenantID:       tenantID,
		StorageUsed:    0,
		DocumentCount:  0,
		UserCount:      1, // The tenant creator
		APICallsToday:  0,
		BandwidthMonth: 0,
		LastAPICall:    time.Now(),
		LastResetDate:  time.Now(),
		UpdatedAt:      time.Now(),
	}

	_ = s.repo.CreateUsage(ctx, usage)

	// Invalidate cache
	cacheKey := cache.TenantKey(tenantID.String(), "quota")
	_ = s.cache.Delete(ctx, cacheKey)

	logger.InfoContext(ctx, "quota created",
		zap.String("tenant_id", tenantID.String()),
		zap.String("plan", req.PlanName),
	)

	return quota, nil
}

// GetQuota retrieves quota for current tenant
func (s *Service) GetQuota(ctx context.Context) (*models.Quota, error) {
	tenantID := getTenantID(ctx)

	// Try cache first
	cacheKey := cache.TenantKey(tenantID.String(), "quota")
	var quota models.Quota
	if err := s.cache.Get(ctx, cacheKey, &quota); err == nil {
		return &quota, nil
	}

	// Fetch from database
	quotaPtr, err := s.repo.GetQuota(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	// Cache for future requests
	_ = s.cache.Set(ctx, cacheKey, quotaPtr, quotaCacheTTL)

	return quotaPtr, nil
}

// UpdateQuota updates quota for current tenant
func (s *Service) UpdateQuota(ctx context.Context, req *models.UpdateQuotaRequest) error {
	tenantID := getTenantID(ctx)

	// Build updates map
	updates := make(map[string]interface{})

	if req.MaxStorage != nil {
		updates["max_storage"] = *req.MaxStorage
	}

	if req.MaxDocuments != nil {
		updates["max_documents"] = *req.MaxDocuments
	}

	if req.MaxUsers != nil {
		updates["max_users"] = *req.MaxUsers
	}

	if req.MaxAPICallsPerDay != nil {
		updates["max_api_calls_per_day"] = *req.MaxAPICallsPerDay
	}

	if req.MaxFileSize != nil {
		updates["max_file_size"] = *req.MaxFileSize
	}

	if req.MaxBandwidth != nil {
		updates["max_bandwidth"] = *req.MaxBandwidth
	}

	if len(req.Features) > 0 {
		featuresJSON, _ := json.Marshal(req.Features)
		updates["features"] = string(featuresJSON)
	}

	if req.ValidUntil != "" {
		parsed, err := time.Parse(time.RFC3339, req.ValidUntil)
		if err != nil {
			return errors.Validationf("invalid valid_until format")
		}
		updates["valid_until"] = parsed
	}

	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	if len(updates) == 0 {
		return nil
	}

	if err := s.repo.UpdateQuota(ctx, tenantID, updates); err != nil {
		return err
	}

	// Invalidate cache
	cacheKey := cache.TenantKey(tenantID.String(), "quota")
	_ = s.cache.Delete(ctx, cacheKey)

	logger.InfoContext(ctx, "quota updated", zap.String("tenant_id", tenantID.String()))

	return nil
}

// GetUsage retrieves usage for current tenant
func (s *Service) GetUsage(ctx context.Context) (*models.Usage, error) {
	tenantID := getTenantID(ctx)

	// Try cache first
	cacheKey := cache.TenantKey(tenantID.String(), "usage")
	var usage models.Usage
	if err := s.cache.Get(ctx, cacheKey, &usage); err == nil {
		return &usage, nil
	}

	// Fetch from database
	usagePtr, err := s.repo.GetUsage(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	// Check if we need to reset daily/monthly counters
	s.checkAndResetCounters(ctx, usagePtr)

	// Cache for future requests
	_ = s.cache.Set(ctx, cacheKey, usagePtr, usageCacheTTL)

	return usagePtr, nil
}

// GetQuotaUsageOverview retrieves quota and usage overview
func (s *Service) GetQuotaUsageOverview(ctx context.Context) (*models.QuotaUsageOverview, error) {
	quota, err := s.GetQuota(ctx)
	if err != nil {
		return nil, err
	}

	usage, err := s.GetUsage(ctx)
	if err != nil {
		return nil, err
	}

	overview := &models.QuotaUsageOverview{
		Quota: *quota,
		Usage: *usage,
	}

	// Calculate percentages
	if quota.MaxStorage > 0 {
		overview.StoragePercent = float64(usage.StorageUsed) / float64(quota.MaxStorage) * 100
	}

	if quota.MaxDocuments > 0 {
		overview.DocumentsPercent = float64(usage.DocumentCount) / float64(quota.MaxDocuments) * 100
	}

	if quota.MaxUsers > 0 {
		overview.UsersPercent = float64(usage.UserCount) / float64(quota.MaxUsers) * 100
	}

	if quota.MaxAPICallsPerDay > 0 {
		overview.APICallsPercent = float64(usage.APICallsToday) / float64(quota.MaxAPICallsPerDay) * 100
	}

	if quota.MaxBandwidth > 0 {
		overview.BandwidthPercent = float64(usage.BandwidthMonth) / float64(quota.MaxBandwidth) * 100
	}

	// Check if limits are exceeded
	overview.IsStorageExceeded = usage.StorageUsed >= quota.MaxStorage
	overview.IsLimitReached = usage.DocumentCount >= quota.MaxDocuments ||
		usage.UserCount >= quota.MaxUsers ||
		usage.APICallsToday >= quota.MaxAPICallsPerDay ||
		usage.BandwidthMonth >= quota.MaxBandwidth

	return overview, nil
}

// CheckQuota checks if a resource usage is within quota
func (s *Service) CheckQuota(ctx context.Context, req *models.CheckQuotaRequest) (*models.CheckQuotaResponse, error) {
	quota, err := s.GetQuota(ctx)
	if err != nil {
		return nil, err
	}

	usage, err := s.GetUsage(ctx)
	if err != nil {
		return nil, err
	}

	response := &models.CheckQuotaResponse{
		Resource:        req.Resource,
		RequestedAmount: req.Amount,
	}

	switch req.Resource {
	case "storage":
		response.CurrentUsage = usage.StorageUsed
		response.MaxAllowed = quota.MaxStorage
		response.Remaining = quota.MaxStorage - usage.StorageUsed
		response.Allowed = (usage.StorageUsed + req.Amount) <= quota.MaxStorage

	case "documents":
		response.CurrentUsage = int64(usage.DocumentCount)
		response.MaxAllowed = int64(quota.MaxDocuments)
		response.Remaining = int64(quota.MaxDocuments - usage.DocumentCount)
		response.Allowed = (usage.DocumentCount + int(req.Amount)) <= quota.MaxDocuments

	case "users":
		response.CurrentUsage = int64(usage.UserCount)
		response.MaxAllowed = int64(quota.MaxUsers)
		response.Remaining = int64(quota.MaxUsers - usage.UserCount)
		response.Allowed = (usage.UserCount + int(req.Amount)) <= quota.MaxUsers

	case "api_calls":
		response.CurrentUsage = int64(usage.APICallsToday)
		response.MaxAllowed = int64(quota.MaxAPICallsPerDay)
		response.Remaining = int64(quota.MaxAPICallsPerDay - usage.APICallsToday)
		response.Allowed = (usage.APICallsToday + int(req.Amount)) <= quota.MaxAPICallsPerDay

	case "bandwidth":
		response.CurrentUsage = usage.BandwidthMonth
		response.MaxAllowed = quota.MaxBandwidth
		response.Remaining = quota.MaxBandwidth - usage.BandwidthMonth
		response.Allowed = (usage.BandwidthMonth + req.Amount) <= quota.MaxBandwidth

	case "file_size":
		response.CurrentUsage = 0 // Single file check
		response.MaxAllowed = quota.MaxFileSize
		response.Remaining = quota.MaxFileSize
		response.Allowed = req.Amount <= quota.MaxFileSize

	default:
		return nil, errors.Validationf("invalid resource type")
	}

	if !response.Allowed {
		response.Message = "Quota limit exceeded"
	}

	return response, nil
}

// IncrementUsage increments usage for a resource
func (s *Service) IncrementUsage(ctx context.Context, req *models.IncrementUsageRequest) error {
	tenantID := getTenantID(ctx)

	var err error
	switch req.Resource {
	case "storage":
		err = s.repo.IncrementStorage(ctx, tenantID, req.Amount)
	case "documents":
		err = s.repo.IncrementDocumentCount(ctx, tenantID, int(req.Amount))
	case "api_calls":
		err = s.repo.IncrementAPICallCount(ctx, tenantID)
	case "bandwidth":
		err = s.repo.IncrementBandwidth(ctx, tenantID, req.Amount)
	default:
		return errors.Validationf("invalid resource type")
	}

	if err != nil {
		return err
	}

	// Log usage
	usageLog := &models.UsageLog{
		ID:        uuid.New(),
		TenantID:  tenantID,
		Action:    "increment",
		Resource:  req.Resource,
		Amount:    req.Amount,
		CreatedAt: time.Now(),
	}

	if req.UserID != "" {
		usageLog.UserID.String = req.UserID
		usageLog.UserID.Valid = true
	}

	if req.Metadata != "" {
		usageLog.Metadata.String = req.Metadata
		usageLog.Metadata.Valid = true
	}

	_ = s.repo.CreateUsageLog(ctx, usageLog)

	// Invalidate cache
	cacheKey := cache.TenantKey(tenantID.String(), "usage")
	_ = s.cache.Delete(ctx, cacheKey)

	return nil
}

// DecrementUsage decrements usage for a resource
func (s *Service) DecrementUsage(ctx context.Context, req *models.DecrementUsageRequest) error {
	tenantID := getTenantID(ctx)

	var err error
	switch req.Resource {
	case "storage":
		err = s.repo.DecrementStorage(ctx, tenantID, req.Amount)
	case "documents":
		err = s.repo.DecrementDocumentCount(ctx, tenantID, int(req.Amount))
	default:
		return errors.Validationf("invalid resource type")
	}

	if err != nil {
		return err
	}

	// Log usage
	usageLog := &models.UsageLog{
		ID:        uuid.New(),
		TenantID:  tenantID,
		Action:    "decrement",
		Resource:  req.Resource,
		Amount:    -req.Amount, // Negative for decrement
		CreatedAt: time.Now(),
	}

	if req.UserID != "" {
		usageLog.UserID.String = req.UserID
		usageLog.UserID.Valid = true
	}

	_ = s.repo.CreateUsageLog(ctx, usageLog)

	// Invalidate cache
	cacheKey := cache.TenantKey(tenantID.String(), "usage")
	_ = s.cache.Delete(ctx, cacheKey)

	return nil
}

// GetUsageStats retrieves usage statistics
func (s *Service) GetUsageStats(ctx context.Context, params *models.UsageStatsParams) (*models.UsageStats, error) {
	tenantID := getTenantID(ctx)

	params.Normalize()

	stats, err := s.repo.GetUsageStats(ctx, tenantID, params)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// GetUsageLogs retrieves usage logs
func (s *Service) GetUsageLogs(ctx context.Context, params *models.UsageStatsParams) ([]models.UsageLog, error) {
	tenantID := getTenantID(ctx)

	params.Normalize()

	logs, err := s.repo.GetUsageLogs(ctx, tenantID, params)
	if err != nil {
		return nil, err
	}

	return logs, nil
}

// GetPredefinedPlans returns predefined quota plans
func (s *Service) GetPredefinedPlans() []models.QuotaPlan {
	return models.GetPredefinedPlans()
}

// Helper functions

func getTenantID(ctx context.Context) uuid.UUID {
	tenantIDStr := middleware.GetTenantID(ctx)
	tenantID, _ := uuid.Parse(tenantIDStr)
	return tenantID
}

func (s *Service) checkAndResetCounters(ctx context.Context, usage *models.Usage) {
	tenantID := usage.TenantID
	now := time.Now()

	// Reset daily API calls if last reset was yesterday or earlier
	if usage.LastResetDate.Before(now.Truncate(24 * time.Hour)) {
		_ = s.repo.ResetDailyAPICallCount(ctx, tenantID)
	}

	// Reset monthly bandwidth if we're in a new month
	lastMonth := usage.LastResetDate.Month()
	currentMonth := now.Month()
	if lastMonth != currentMonth {
		_ = s.repo.ResetMonthlyBandwidth(ctx, tenantID)
	}
}

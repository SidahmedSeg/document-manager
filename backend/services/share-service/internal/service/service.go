package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"github.com/SidahmedSeg/document-manager/backend/pkg/cache"
	"github.com/SidahmedSeg/document-manager/backend/pkg/errors"
	"github.com/SidahmedSeg/document-manager/backend/pkg/logger"
	"github.com/SidahmedSeg/document-manager/backend/pkg/middleware"
	"github.com/SidahmedSeg/document-manager/backend/services/share-service/internal/models"
	"github.com/SidahmedSeg/document-manager/backend/services/share-service/internal/repository"
	"go.uber.org/zap"
)

const (
	shareCacheTTL = 30 * time.Minute
	tokenLength   = 32
	baseURL       = "https://app.docmanager.com/share" // TODO: Make configurable
)

// Service handles share business logic
type Service struct {
	repo   *repository.Repository
	cache  *cache.Cache
	logger *zap.Logger
}

// NewService creates a new share service
func NewService(repo *repository.Repository, cache *cache.Cache, logger *zap.Logger) *Service {
	return &Service{
		repo:   repo,
		cache:  cache,
		logger: logger,
	}
}

// CreateShare creates a new share
func (s *Service) CreateShare(ctx context.Context, req *models.CreateShareRequest) (*models.CreateShareResponse, error) {
	tenantID := getTenantID(ctx)
	userID := middleware.GetUserID(ctx)

	// Parse document ID
	documentID, err := uuid.Parse(req.DocumentID)
	if err != nil {
		return nil, errors.Validationf("invalid document_id")
	}

	// Parse expiration time if provided
	var expiresAt *time.Time
	if req.ExpiresAt != "" {
		parsed, err := time.Parse(time.RFC3339, req.ExpiresAt)
		if err != nil {
			return nil, errors.Validationf("invalid expires_at format")
		}
		if parsed.Before(time.Now()) {
			return nil, errors.Validationf("expires_at must be in the future")
		}
		expiresAt = &parsed
	}

	// Create share
	share := &models.Share{
		ID:          uuid.New(),
		TenantID:    tenantID,
		DocumentID:  documentID,
		ShareType:   req.ShareType,
		SharedBy:    userID,
		Permission:  req.Permission,
		AccessCount: 0,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Set shared_with for user shares
	if req.ShareType == "user" || req.ShareType == "email" {
		share.SharedWith.String = req.SharedWith
		share.SharedWith.Valid = true
	}

	// Generate token for public shares
	if req.ShareType == "public" {
		token, err := generateSecureToken(tokenLength)
		if err != nil {
			s.logger.Error("failed to generate share token", zap.Error(err))
			return nil, errors.New(errors.ErrCodeInternal, "failed to generate share token")
		}
		share.ShareToken.String = token
		share.ShareToken.Valid = true
	}

	// Set expiration
	if expiresAt != nil {
		share.ExpiresAt.Time = *expiresAt
		share.ExpiresAt.Valid = true
	}

	// Hash password if provided
	if req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			s.logger.Error("failed to hash password", zap.Error(err))
			return nil, errors.New(errors.ErrCodeInternal, "failed to secure password")
		}
		share.Password.String = string(hashedPassword)
		share.Password.Valid = true
	}

	// Set max access
	if req.MaxAccess > 0 {
		share.MaxAccess.Int64 = int64(req.MaxAccess)
		share.MaxAccess.Valid = true
	}

	// Create share in database
	if err := s.repo.CreateShare(ctx, share); err != nil {
		return nil, err
	}

	logger.InfoContext(ctx, "share created",
		zap.String("share_id", share.ID.String()),
		zap.String("document_id", documentID.String()),
		zap.String("share_type", req.ShareType),
	)

	// Build response
	response := &models.CreateShareResponse{
		ID:         share.ID,
		DocumentID: documentID,
		ShareType:  req.ShareType,
		Permission: req.Permission,
		CreatedAt:  share.CreatedAt,
	}

	if share.ShareToken.Valid {
		response.ShareToken = &share.ShareToken.String
		shareURL := fmt.Sprintf("%s/%s", baseURL, share.ShareToken.String)
		response.ShareURL = &shareURL
	}

	if share.ExpiresAt.Valid {
		response.ExpiresAt = &share.ExpiresAt.Time
	}

	return response, nil
}

// GetShare retrieves a share by ID
func (s *Service) GetShare(ctx context.Context, shareID uuid.UUID) (*models.Share, error) {
	tenantID := getTenantID(ctx)

	// Try cache first
	cacheKey := cache.TenantKey(tenantID.String(), "share", shareID.String())
	var share models.Share
	if err := s.cache.Get(ctx, cacheKey, &share); err == nil {
		return &share, nil
	}

	// Fetch from database
	sharePtr, err := s.repo.GetShare(ctx, tenantID, shareID)
	if err != nil {
		return nil, err
	}

	// Cache for future requests
	_ = s.cache.Set(ctx, cacheKey, sharePtr, shareCacheTTL)

	return sharePtr, nil
}

// AccessShare accesses a share using a token
func (s *Service) AccessShare(ctx context.Context, req *models.AccessShareRequest, ipAddress, userAgent string) (*models.AccessShareResponse, error) {
	// Get share by token
	share, err := s.repo.GetShareByToken(ctx, req.ShareToken)
	if err != nil {
		return nil, errors.NotFoundf("share link not found")
	}

	// Verify share is active
	if !share.IsActive {
		return nil, errors.Forbiddenf("share link has been revoked")
	}

	// Check expiration
	if share.ExpiresAt.Valid && share.ExpiresAt.Time.Before(time.Now()) {
		return nil, errors.Forbiddenf("share link has expired")
	}

	// Check max access limit
	if share.MaxAccess.Valid && share.AccessCount >= int(share.MaxAccess.Int64) {
		return nil, errors.Forbiddenf("share link has reached maximum access limit")
	}

	// Verify password if required
	if share.Password.Valid {
		if req.Password == "" {
			return nil, errors.Unauthorizedf("password required")
		}
		if err := bcrypt.CompareHashAndPassword([]byte(share.Password.String), []byte(req.Password)); err != nil {
			return nil, errors.Unauthorizedf("invalid password")
		}
	}

	// Increment access count
	if err := s.repo.IncrementAccessCount(ctx, share.ID); err != nil {
		s.logger.Error("failed to increment access count", zap.Error(err))
	}

	// Log access
	userID := middleware.GetUserID(ctx)
	accessLog := &models.ShareAccess{
		ID:         uuid.New(),
		ShareID:    share.ID,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		Action:     "view",
		AccessedAt: time.Now(),
	}
	if userID != "" {
		accessLog.AccessedBy.String = userID
		accessLog.AccessedBy.Valid = true
	}

	if err := s.repo.CreateShareAccess(ctx, accessLog); err != nil {
		s.logger.Error("failed to log share access", zap.Error(err))
	}

	// TODO: Get document name and download URL from document service
	response := &models.AccessShareResponse{
		DocumentID: share.DocumentID,
		DocumentName: "Document", // Placeholder
		Permission: share.Permission,
		ExpiresAt:  time.Now().Add(1 * time.Hour), // Placeholder
	}

	if share.Permission == "download" {
		response.DownloadURL = "https://storage.docmanager.com/download/placeholder" // Placeholder
	}

	return response, nil
}

// ListShares retrieves shares with filtering
func (s *Service) ListShares(ctx context.Context, params *models.ListSharesParams) ([]models.Share, int64, error) {
	tenantID := getTenantID(ctx)

	params.Normalize()

	shares, total, err := s.repo.ListShares(ctx, tenantID, params)
	if err != nil {
		return nil, 0, err
	}

	return shares, total, nil
}

// UpdateShare updates a share
func (s *Service) UpdateShare(ctx context.Context, shareID uuid.UUID, req *models.UpdateShareRequest) error {
	tenantID := getTenantID(ctx)

	// Verify share exists
	if _, err := s.repo.GetShare(ctx, tenantID, shareID); err != nil {
		return err
	}

	// Build updates map
	updates := make(map[string]interface{})

	if req.Permission != "" {
		updates["permission"] = req.Permission
	}

	if req.ExpiresAt != "" {
		parsed, err := time.Parse(time.RFC3339, req.ExpiresAt)
		if err != nil {
			return errors.Validationf("invalid expires_at format")
		}
		if parsed.Before(time.Now()) {
			return errors.Validationf("expires_at must be in the future")
		}
		updates["expires_at"] = parsed
	}

	if req.MaxAccess != nil {
		updates["max_access"] = *req.MaxAccess
	}

	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	if len(updates) == 0 {
		return nil
	}

	// Update share
	if err := s.repo.UpdateShare(ctx, tenantID, shareID, updates); err != nil {
		return err
	}

	// Invalidate cache
	cacheKey := cache.TenantKey(tenantID.String(), "share", shareID.String())
	_ = s.cache.Delete(ctx, cacheKey)

	logger.InfoContext(ctx, "share updated", zap.String("share_id", shareID.String()))

	return nil
}

// RevokeShare revokes a share
func (s *Service) RevokeShare(ctx context.Context, shareID uuid.UUID) error {
	tenantID := getTenantID(ctx)

	// Update share to inactive
	updates := map[string]interface{}{
		"is_active": false,
	}

	if err := s.repo.UpdateShare(ctx, tenantID, shareID, updates); err != nil {
		return err
	}

	// Invalidate cache
	cacheKey := cache.TenantKey(tenantID.String(), "share", shareID.String())
	_ = s.cache.Delete(ctx, cacheKey)

	logger.InfoContext(ctx, "share revoked", zap.String("share_id", shareID.String()))

	return nil
}

// DeleteShare deletes a share
func (s *Service) DeleteShare(ctx context.Context, shareID uuid.UUID) error {
	tenantID := getTenantID(ctx)

	if err := s.repo.DeleteShare(ctx, tenantID, shareID); err != nil {
		return err
	}

	// Invalidate cache
	cacheKey := cache.TenantKey(tenantID.String(), "share", shareID.String())
	_ = s.cache.Delete(ctx, cacheKey)

	logger.InfoContext(ctx, "share deleted", zap.String("share_id", shareID.String()))

	return nil
}

// GetShareAccessLogs retrieves access logs for a share
func (s *Service) GetShareAccessLogs(ctx context.Context, shareID uuid.UUID, limit int) ([]models.ShareAccess, error) {
	tenantID := getTenantID(ctx)

	// Verify share exists and belongs to tenant
	if _, err := s.repo.GetShare(ctx, tenantID, shareID); err != nil {
		return nil, err
	}

	if limit <= 0 || limit > 100 {
		limit = 50
	}

	logs, err := s.repo.GetShareAccessLogs(ctx, shareID, limit)
	if err != nil {
		return nil, err
	}

	return logs, nil
}

// GetShareStats retrieves share statistics
func (s *Service) GetShareStats(ctx context.Context) (*models.ShareStats, error) {
	tenantID := getTenantID(ctx)

	stats, err := s.repo.GetShareStats(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// VerifyShareToken verifies a share token
func (s *Service) VerifyShareToken(ctx context.Context, token string, password string) (*models.VerifyShareTokenResponse, error) {
	// Get share by token
	share, err := s.repo.GetShareByToken(ctx, token)
	if err != nil {
		return &models.VerifyShareTokenResponse{Valid: false}, nil
	}

	// Check if active
	if !share.IsActive {
		return &models.VerifyShareTokenResponse{Valid: false}, nil
	}

	// Check expiration
	if share.ExpiresAt.Valid && share.ExpiresAt.Time.Before(time.Now()) {
		return &models.VerifyShareTokenResponse{Valid: false}, nil
	}

	// Check max access
	if share.MaxAccess.Valid && share.AccessCount >= int(share.MaxAccess.Int64) {
		return &models.VerifyShareTokenResponse{Valid: false}, nil
	}

	// Verify password if required
	if share.Password.Valid {
		if password == "" {
			return &models.VerifyShareTokenResponse{Valid: false}, nil
		}
		if err := bcrypt.CompareHashAndPassword([]byte(share.Password.String), []byte(password)); err != nil {
			return &models.VerifyShareTokenResponse{Valid: false}, nil
		}
	}

	response := &models.VerifyShareTokenResponse{
		Valid:      true,
		ShareID:    share.ID,
		DocumentID: share.DocumentID,
		Permission: share.Permission,
	}

	if share.ExpiresAt.Valid {
		response.ExpiresAt = &share.ExpiresAt.Time
	}

	return response, nil
}

// Helper functions

func getTenantID(ctx context.Context) uuid.UUID {
	tenantIDStr := middleware.GetTenantID(ctx)
	tenantID, _ := uuid.Parse(tenantIDStr)
	return tenantID
}

func generateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}

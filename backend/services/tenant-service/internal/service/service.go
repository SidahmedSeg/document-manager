package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/SidahmedSeg/document-manager/backend/pkg/cache"
	"github.com/SidahmedSeg/document-manager/backend/pkg/errors"
	"github.com/SidahmedSeg/document-manager/backend/pkg/logger"
	"github.com/SidahmedSeg/document-manager/backend/pkg/middleware"
	"github.com/SidahmedSeg/document-manager/backend/services/tenant-service/internal/models"
	"github.com/SidahmedSeg/document-manager/backend/services/tenant-service/internal/repository"
	"go.uber.org/zap"
)

const (
	invitationTokenLength = 32
	invitationExpiry      = 7 * 24 * time.Hour // 7 days
	tenantCacheTTL        = 1 * time.Hour
)

// Service handles tenant business logic
type Service struct {
	repo   *repository.Repository
	cache  *cache.Cache
	logger *zap.Logger
}

// NewService creates a new tenant service
func NewService(repo *repository.Repository, cache *cache.Cache, logger *zap.Logger) *Service {
	return &Service{
		repo:   repo,
		cache:  cache,
		logger: logger,
	}
}

// CreateTenant creates a new tenant
func (s *Service) CreateTenant(ctx context.Context, req *models.CreateTenantRequest) (*models.Tenant, error) {
	userID := middleware.GetUserID(ctx)
	userEmail := middleware.GetUserEmail(ctx)

	if userID == "" {
		return nil, errors.ErrUnauthorized
	}

	// Check if slug is already taken
	existing, err := s.repo.GetTenantBySlug(ctx, req.Slug)
	if err == nil && existing != nil {
		return nil, errors.Conflictf("tenant slug '%s' is already taken", req.Slug)
	}

	// Create tenant
	tenant := &models.Tenant{
		ID:               uuid.New(),
		Name:             req.Name,
		Slug:             strings.ToLower(req.Slug),
		SubscriptionPlan: "free", // Default to free plan
		IsActive:         true,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if req.Domain != "" {
		tenant.Domain.String = req.Domain
		tenant.Domain.Valid = true
	}

	// Create tenant in database
	if err := s.repo.CreateTenant(ctx, tenant); err != nil {
		return nil, err
	}

	// Add creator as owner
	tenantUser := &models.TenantUser{
		ID:        uuid.New(),
		TenantID:  tenant.ID,
		UserID:    userID,
		UserEmail: userEmail,
		Role:      "admin",
		IsOwner:   true,
		JoinedAt:  time.Now(),
	}

	if err := s.repo.AddTenantUser(ctx, tenantUser); err != nil {
		s.logger.Error("failed to add tenant owner", zap.Error(err))
		return nil, err
	}

	// Cache tenant
	cacheKey := cache.BuildKey("tenant", tenant.ID.String())
	_ = s.cache.Set(ctx, cacheKey, tenant, tenantCacheTTL)

	logger.InfoContext(ctx, "tenant created",
		zap.String("tenant_id", tenant.ID.String()),
		zap.String("name", tenant.Name),
		zap.String("slug", tenant.Slug),
	)

	return tenant, nil
}

// GetTenant retrieves a tenant by ID
func (s *Service) GetTenant(ctx context.Context, tenantID uuid.UUID) (*models.Tenant, error) {
	userID := middleware.GetUserID(ctx)

	// Check if user has access to this tenant
	hasAccess, err := s.repo.IsUserInTenant(ctx, tenantID, userID)
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, errors.ErrForbidden
	}

	// Try cache first
	cacheKey := cache.BuildKey("tenant", tenantID.String())
	var tenant models.Tenant
	if err := s.cache.Get(ctx, cacheKey, &tenant); err == nil {
		return &tenant, nil
	}

	// Fetch from database
	tenantPtr, err := s.repo.GetTenantByID(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	// Cache for future requests
	_ = s.cache.Set(ctx, cacheKey, tenantPtr, tenantCacheTTL)

	return tenantPtr, nil
}

// UpdateTenant updates a tenant
func (s *Service) UpdateTenant(ctx context.Context, tenantID uuid.UUID, req *models.UpdateTenantRequest) error {
	userID := middleware.GetUserID(ctx)

	// Check if user is admin or owner
	role, err := s.repo.GetUserRole(ctx, tenantID, userID)
	if err != nil {
		return err
	}
	if role != "admin" {
		return errors.Forbiddenf("only admins can update tenant settings")
	}

	// Update tenant
	if err := s.repo.UpdateTenant(ctx, tenantID, req); err != nil {
		return err
	}

	// Invalidate cache
	cacheKey := cache.BuildKey("tenant", tenantID.String())
	_ = s.cache.Delete(ctx, cacheKey)

	logger.InfoContext(ctx, "tenant updated", zap.String("tenant_id", tenantID.String()))

	return nil
}

// GetTenantUsers retrieves all users in a tenant
func (s *Service) GetTenantUsers(ctx context.Context, tenantID uuid.UUID) ([]models.TenantUser, error) {
	userID := middleware.GetUserID(ctx)

	// Check if user has access to this tenant
	hasAccess, err := s.repo.IsUserInTenant(ctx, tenantID, userID)
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, errors.ErrForbidden
	}

	users, err := s.repo.GetTenantUsers(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	return users, nil
}

// InviteUser invites a user to join a tenant
func (s *Service) InviteUser(ctx context.Context, tenantID uuid.UUID, req *models.InviteUserRequest) (*models.TenantInvitation, error) {
	userID := middleware.GetUserID(ctx)

	// Check if inviter is admin
	role, err := s.repo.GetUserRole(ctx, tenantID, userID)
	if err != nil {
		return nil, err
	}
	if role != "admin" {
		return nil, errors.Forbiddenf("only admins can invite users")
	}

	// Check if user is already a member
	users, err := s.repo.GetTenantUsers(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	for _, u := range users {
		if strings.EqualFold(u.UserEmail, req.Email) {
			return nil, errors.Conflictf("user is already a member of this tenant")
		}
	}

	// Generate invitation token
	token, err := generateToken(invitationTokenLength)
	if err != nil {
		return nil, errors.Internalf(err, "failed to generate invitation token")
	}

	// Create invitation
	invitation := &models.TenantInvitation{
		ID:        uuid.New(),
		TenantID:  tenantID,
		Email:     strings.ToLower(req.Email),
		Role:      req.Role,
		InvitedBy: userID,
		Token:     token,
		ExpiresAt: time.Now().Add(invitationExpiry),
		CreatedAt: time.Now(),
	}

	if err := s.repo.CreateInvitation(ctx, invitation); err != nil {
		return nil, err
	}

	logger.InfoContext(ctx, "user invited to tenant",
		zap.String("tenant_id", tenantID.String()),
		zap.String("email", req.Email),
		zap.String("role", req.Role),
	)

	// TODO: Send invitation email via notification service

	return invitation, nil
}

// RemoveUser removes a user from a tenant
func (s *Service) RemoveUser(ctx context.Context, tenantID uuid.UUID, targetUserID string) error {
	userID := middleware.GetUserID(ctx)

	// Check if remover is admin
	role, err := s.repo.GetUserRole(ctx, tenantID, userID)
	if err != nil {
		return err
	}
	if role != "admin" {
		return errors.Forbiddenf("only admins can remove users")
	}

	// Cannot remove yourself
	if userID == targetUserID {
		return errors.Forbiddenf("cannot remove yourself from the tenant")
	}

	if err := s.repo.RemoveTenantUser(ctx, tenantID, targetUserID); err != nil {
		return err
	}

	logger.InfoContext(ctx, "user removed from tenant",
		zap.String("tenant_id", tenantID.String()),
		zap.String("removed_user_id", targetUserID),
	)

	return nil
}

// GetUserTenants retrieves all tenants a user belongs to
func (s *Service) GetUserTenants(ctx context.Context) ([]models.Tenant, error) {
	userID := middleware.GetUserID(ctx)

	tenants, err := s.repo.GetUserTenants(ctx, userID)
	if err != nil {
		return nil, err
	}

	return tenants, nil
}

// GetPendingInvitations retrieves pending invitations for a tenant
func (s *Service) GetPendingInvitations(ctx context.Context, tenantID uuid.UUID) ([]models.TenantInvitation, error) {
	userID := middleware.GetUserID(ctx)

	// Check if user is admin
	role, err := s.repo.GetUserRole(ctx, tenantID, userID)
	if err != nil {
		return nil, err
	}
	if role != "admin" {
		return nil, errors.Forbiddenf("only admins can view invitations")
	}

	invitations, err := s.repo.GetPendingInvitations(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	return invitations, nil
}

// generateToken generates a random token
func generateToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}

// ValidateSlug validates and normalizes a tenant slug
func ValidateSlug(slug string) (string, error) {
	slug = strings.ToLower(strings.TrimSpace(slug))

	// Check length
	if len(slug) < 2 || len(slug) > 50 {
		return "", errors.Validationf("slug must be between 2 and 50 characters")
	}

	// Check format (alphanumeric and hyphens only)
	for _, char := range slug {
		if !(char >= 'a' && char <= 'z') && !(char >= '0' && char <= '9') && char != '-' {
			return "", errors.Validationf("slug can only contain lowercase letters, numbers, and hyphens")
		}
	}

	// Cannot start or end with hyphen
	if slug[0] == '-' || slug[len(slug)-1] == '-' {
		return "", errors.Validationf("slug cannot start or end with a hyphen")
	}

	// Reserved slugs
	reserved := []string{"admin", "api", "www", "app", "dashboard", "system", "internal"}
	for _, r := range reserved {
		if slug == r {
			return "", errors.Validationf("slug '%s' is reserved", slug)
		}
	}

	return slug, nil
}

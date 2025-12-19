package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// Tenant represents a tenant in the system
type Tenant struct {
	ID               uuid.UUID      `json:"id" db:"id"`
	Name             string         `json:"name" db:"name"`
	Slug             string         `json:"slug" db:"slug"`
	Domain           sql.NullString `json:"domain,omitempty" db:"domain"`
	SubscriptionPlan string         `json:"subscription_plan" db:"subscription_plan"`
	IsActive         bool           `json:"is_active" db:"is_active"`
	CreatedAt        time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at" db:"updated_at"`
}

// TenantUser represents a user's membership in a tenant
type TenantUser struct {
	ID        uuid.UUID      `json:"id" db:"id"`
	TenantID  uuid.UUID      `json:"tenant_id" db:"tenant_id"`
	UserID    string         `json:"user_id" db:"user_id"` // Kratos user ID
	UserEmail string         `json:"user_email" db:"user_email"`
	Role      string         `json:"role" db:"role"`
	IsOwner   bool           `json:"is_owner" db:"is_owner"`
	JoinedAt  time.Time      `json:"joined_at" db:"joined_at"`
	InvitedBy sql.NullString `json:"invited_by,omitempty" db:"invited_by"`
}

// TenantInvitation represents a pending invitation to join a tenant
type TenantInvitation struct {
	ID         uuid.UUID      `json:"id" db:"id"`
	TenantID   uuid.UUID      `json:"tenant_id" db:"tenant_id"`
	Email      string         `json:"email" db:"email"`
	Role       string         `json:"role" db:"role"`
	InvitedBy  string         `json:"invited_by" db:"invited_by"`
	Token      string         `json:"-" db:"token"` // Don't expose in API
	ExpiresAt  time.Time      `json:"expires_at" db:"expires_at"`
	AcceptedAt sql.NullTime   `json:"accepted_at,omitempty" db:"accepted_at"`
	CreatedAt  time.Time      `json:"created_at" db:"created_at"`
}

// TenantSettings represents tenant-specific settings
type TenantSettings struct {
	ID        uuid.UUID `json:"id" db:"id"`
	TenantID  uuid.UUID `json:"tenant_id" db:"tenant_id"`
	Settings  string    `json:"settings" db:"settings"` // JSONB stored as string
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// CreateTenantRequest represents the request to create a new tenant
type CreateTenantRequest struct {
	Name   string `json:"name" validate:"required,min=2,max=100"`
	Slug   string `json:"slug" validate:"required,min=2,max=50,alphanum"`
	Domain string `json:"domain,omitempty" validate:"omitempty,url"`
}

// UpdateTenantRequest represents the request to update a tenant
type UpdateTenantRequest struct {
	Name     string `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Domain   string `json:"domain,omitempty" validate:"omitempty,url"`
	IsActive *bool  `json:"is_active,omitempty"`
}

// InviteUserRequest represents the request to invite a user to a tenant
type InviteUserRequest struct {
	Email string `json:"email" validate:"required,email"`
	Role  string `json:"role" validate:"required,oneof=admin user guest"`
}

// TenantWithStats includes tenant with additional statistics
type TenantWithStats struct {
	Tenant
	UserCount     int   `json:"user_count"`
	DocumentCount int   `json:"document_count"`
	StorageUsed   int64 `json:"storage_used"`
}

// TenantUserWithDetails includes user details from Kratos
type TenantUserWithDetails struct {
	TenantUser
	UserName string `json:"user_name,omitempty"`
}

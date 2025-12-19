package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// Share represents a document share
type Share struct {
	ID          uuid.UUID      `json:"id" db:"id"`
	TenantID    uuid.UUID      `json:"tenant_id" db:"tenant_id"`
	DocumentID  uuid.UUID      `json:"document_id" db:"document_id"`
	ShareType   string         `json:"share_type" db:"share_type"` // user, public, email
	SharedBy    string         `json:"shared_by" db:"shared_by"`
	SharedWith  sql.NullString `json:"shared_with,omitempty" db:"shared_with"` // user_id or email
	Permission  string         `json:"permission" db:"permission"`             // view, edit, download
	ShareToken  sql.NullString `json:"share_token,omitempty" db:"share_token"` // for public links
	ExpiresAt   sql.NullTime   `json:"expires_at,omitempty" db:"expires_at"`
	Password    sql.NullString `json:"-" db:"password"`                    // hashed password for protected links
	MaxAccess   sql.NullInt64  `json:"max_access,omitempty" db:"max_access"` // max access count
	AccessCount int            `json:"access_count" db:"access_count"`
	IsActive    bool           `json:"is_active" db:"is_active"`
	CreatedAt   time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at" db:"updated_at"`
}

// ShareAccess represents share access log
type ShareAccess struct {
	ID         uuid.UUID      `json:"id" db:"id"`
	ShareID    uuid.UUID      `json:"share_id" db:"share_id"`
	AccessedBy sql.NullString `json:"accessed_by,omitempty" db:"accessed_by"` // user_id if authenticated
	IPAddress  string         `json:"ip_address" db:"ip_address"`
	UserAgent  string         `json:"user_agent" db:"user_agent"`
	Action     string         `json:"action" db:"action"` // view, download
	AccessedAt time.Time      `json:"accessed_at" db:"accessed_at"`
}

// ShareWithDetails includes share with document and user details
type ShareWithDetails struct {
	Share
	DocumentName string `json:"document_name"`
	SharedByName string `json:"shared_by_name"`
}

// CreateShareRequest represents share creation request
type CreateShareRequest struct {
	DocumentID string `json:"document_id" validate:"required,uuid"`
	ShareType  string `json:"share_type" validate:"required,oneof=user public email"`
	SharedWith string `json:"shared_with,omitempty" validate:"required_if=ShareType user,omitempty,email"`
	Permission string `json:"permission" validate:"required,oneof=view edit download"`
	ExpiresAt  string `json:"expires_at,omitempty" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	Password   string `json:"password,omitempty" validate:"omitempty,min=8,max=100"`
	MaxAccess  int    `json:"max_access,omitempty" validate:"omitempty,gte=1,lte=1000"`
}

// CreateShareResponse represents share creation response
type CreateShareResponse struct {
	ID         uuid.UUID  `json:"id"`
	DocumentID uuid.UUID  `json:"document_id"`
	ShareType  string     `json:"share_type"`
	Permission string     `json:"permission"`
	ShareToken *string    `json:"share_token,omitempty"`
	ShareURL   *string    `json:"share_url,omitempty"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// UpdateShareRequest represents share update request
type UpdateShareRequest struct {
	Permission string `json:"permission,omitempty" validate:"omitempty,oneof=view edit download"`
	ExpiresAt  string `json:"expires_at,omitempty" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	MaxAccess  *int   `json:"max_access,omitempty" validate:"omitempty,gte=1,lte=1000"`
	IsActive   *bool  `json:"is_active,omitempty"`
}

// AccessShareRequest represents share access request
type AccessShareRequest struct {
	ShareToken string `json:"share_token" validate:"required"`
	Password   string `json:"password,omitempty"`
}

// AccessShareResponse represents share access response
type AccessShareResponse struct {
	DocumentID   uuid.UUID `json:"document_id"`
	DocumentName string    `json:"document_name"`
	Permission   string    `json:"permission"`
	DownloadURL  string    `json:"download_url,omitempty"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// ListSharesParams represents query parameters for listing shares
type ListSharesParams struct {
	DocumentID string `json:"document_id,omitempty" form:"document_id"`
	ShareType  string `json:"share_type,omitempty" form:"share_type"`
	SharedWith string `json:"shared_with,omitempty" form:"shared_with"`
	IsActive   string `json:"is_active,omitempty" form:"is_active"`
	Page       int    `json:"page" form:"page" validate:"omitempty,gte=1"`
	Limit      int    `json:"limit" form:"limit" validate:"omitempty,gte=1,lte=100"`
	SortBy     string `json:"sort_by,omitempty" form:"sort_by"`
	SortOrder  string `json:"sort_order,omitempty" form:"sort_order" validate:"omitempty,oneof=asc desc"`
}

// Normalize sets default values for list parameters
func (p *ListSharesParams) Normalize() {
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
		p.SortBy = "created_at"
	}
	if p.SortOrder == "" {
		p.SortOrder = "desc"
	}
}

// GetOffset calculates the database offset
func (p *ListSharesParams) GetOffset() int {
	return (p.Page - 1) * p.Limit
}

// ShareStats represents share statistics
type ShareStats struct {
	TotalShares     int64 `json:"total_shares"`
	ActiveShares    int64 `json:"active_shares"`
	ExpiredShares   int64 `json:"expired_shares"`
	TotalAccess     int64 `json:"total_access"`
	SharesByType    map[string]int64 `json:"shares_by_type"`
	SharesByPermission map[string]int64 `json:"shares_by_permission"`
}

// RevokeShareRequest represents share revocation request
type RevokeShareRequest struct {
	ShareID uuid.UUID `json:"share_id" validate:"required,uuid"`
}

// VerifyShareTokenRequest represents token verification request
type VerifyShareTokenRequest struct {
	Token    string `json:"token" validate:"required"`
	Password string `json:"password,omitempty"`
}

// VerifyShareTokenResponse represents token verification response
type VerifyShareTokenResponse struct {
	Valid      bool       `json:"valid"`
	ShareID    uuid.UUID  `json:"share_id,omitempty"`
	DocumentID uuid.UUID  `json:"document_id,omitempty"`
	Permission string     `json:"permission,omitempty"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
}

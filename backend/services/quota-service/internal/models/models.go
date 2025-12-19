package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// Quota represents tenant quota limits
type Quota struct {
	ID                uuid.UUID      `json:"id" db:"id"`
	TenantID          uuid.UUID      `json:"tenant_id" db:"tenant_id"`
	PlanName          string         `json:"plan_name" db:"plan_name"` // free, basic, pro, enterprise
	MaxStorage        int64          `json:"max_storage" db:"max_storage"` // bytes
	MaxDocuments      int            `json:"max_documents" db:"max_documents"`
	MaxUsers          int            `json:"max_users" db:"max_users"`
	MaxAPICallsPerDay int            `json:"max_api_calls_per_day" db:"max_api_calls_per_day"`
	MaxFileSize       int64          `json:"max_file_size" db:"max_file_size"` // bytes
	MaxBandwidth      int64          `json:"max_bandwidth" db:"max_bandwidth"` // bytes per month
	Features          sql.NullString `json:"features,omitempty" db:"features"` // JSON array of enabled features
	IsActive          bool           `json:"is_active" db:"is_active"`
	ValidFrom         time.Time      `json:"valid_from" db:"valid_from"`
	ValidUntil        sql.NullTime   `json:"valid_until,omitempty" db:"valid_until"`
	CreatedAt         time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at" db:"updated_at"`
}

// Usage represents current usage for a tenant
type Usage struct {
	ID               uuid.UUID `json:"id" db:"id"`
	TenantID         uuid.UUID `json:"tenant_id" db:"tenant_id"`
	StorageUsed      int64     `json:"storage_used" db:"storage_used"` // bytes
	DocumentCount    int       `json:"document_count" db:"document_count"`
	UserCount        int       `json:"user_count" db:"user_count"`
	APICallsToday    int       `json:"api_calls_today" db:"api_calls_today"`
	BandwidthMonth   int64     `json:"bandwidth_month" db:"bandwidth_month"` // bytes
	LastAPICall      time.Time `json:"last_api_call" db:"last_api_call"`
	LastResetDate    time.Time `json:"last_reset_date" db:"last_reset_date"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

// UsageLog represents detailed usage logging
type UsageLog struct {
	ID         uuid.UUID      `json:"id" db:"id"`
	TenantID   uuid.UUID      `json:"tenant_id" db:"tenant_id"`
	UserID     sql.NullString `json:"user_id,omitempty" db:"user_id"`
	Action     string         `json:"action" db:"action"` // upload, download, api_call, etc.
	Resource   string         `json:"resource" db:"resource"` // document, storage, api
	Amount     int64          `json:"amount" db:"amount"` // bytes, count, etc.
	Metadata   sql.NullString `json:"metadata,omitempty" db:"metadata"` // JSON
	CreatedAt  time.Time      `json:"created_at" db:"created_at"`
}

// QuotaUsageOverview combines quota and usage information
type QuotaUsageOverview struct {
	Quota             Quota   `json:"quota"`
	Usage             Usage   `json:"usage"`
	StoragePercent    float64 `json:"storage_percent"`
	DocumentsPercent  float64 `json:"documents_percent"`
	UsersPercent      float64 `json:"users_percent"`
	APICallsPercent   float64 `json:"api_calls_percent"`
	BandwidthPercent  float64 `json:"bandwidth_percent"`
	IsStorageExceeded bool    `json:"is_storage_exceeded"`
	IsLimitReached    bool    `json:"is_limit_reached"`
}

// CreateQuotaRequest represents quota creation request
type CreateQuotaRequest struct {
	PlanName          string   `json:"plan_name" validate:"required,oneof=free basic pro enterprise"`
	MaxStorage        int64    `json:"max_storage" validate:"required,gt=0"`
	MaxDocuments      int      `json:"max_documents" validate:"required,gt=0"`
	MaxUsers          int      `json:"max_users" validate:"required,gt=0"`
	MaxAPICallsPerDay int      `json:"max_api_calls_per_day" validate:"required,gt=0"`
	MaxFileSize       int64    `json:"max_file_size" validate:"required,gt=0"`
	MaxBandwidth      int64    `json:"max_bandwidth" validate:"required,gt=0"`
	Features          []string `json:"features,omitempty"`
	ValidUntil        string   `json:"valid_until,omitempty" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
}

// UpdateQuotaRequest represents quota update request
type UpdateQuotaRequest struct {
	MaxStorage        *int64   `json:"max_storage,omitempty" validate:"omitempty,gt=0"`
	MaxDocuments      *int     `json:"max_documents,omitempty" validate:"omitempty,gt=0"`
	MaxUsers          *int     `json:"max_users,omitempty" validate:"omitempty,gt=0"`
	MaxAPICallsPerDay *int     `json:"max_api_calls_per_day,omitempty" validate:"omitempty,gt=0"`
	MaxFileSize       *int64   `json:"max_file_size,omitempty" validate:"omitempty,gt=0"`
	MaxBandwidth      *int64   `json:"max_bandwidth,omitempty" validate:"omitempty,gt=0"`
	Features          []string `json:"features,omitempty"`
	ValidUntil        string   `json:"valid_until,omitempty" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	IsActive          *bool    `json:"is_active,omitempty"`
}

// CheckQuotaRequest represents quota check request
type CheckQuotaRequest struct {
	Resource string `json:"resource" validate:"required,oneof=storage documents users api_calls bandwidth file_size"`
	Amount   int64  `json:"amount" validate:"required,gt=0"`
}

// CheckQuotaResponse represents quota check response
type CheckQuotaResponse struct {
	Allowed       bool   `json:"allowed"`
	Resource      string `json:"resource"`
	RequestedAmount int64  `json:"requested_amount"`
	CurrentUsage  int64  `json:"current_usage"`
	MaxAllowed    int64  `json:"max_allowed"`
	Remaining     int64  `json:"remaining"`
	Message       string `json:"message,omitempty"`
}

// IncrementUsageRequest represents usage increment request
type IncrementUsageRequest struct {
	Resource string `json:"resource" validate:"required,oneof=storage documents users api_calls bandwidth"`
	Amount   int64  `json:"amount" validate:"required"`
	UserID   string `json:"user_id,omitempty"`
	Metadata string `json:"metadata,omitempty"`
}

// DecrementUsageRequest represents usage decrement request
type DecrementUsageRequest struct {
	Resource string `json:"resource" validate:"required,oneof=storage documents users"`
	Amount   int64  `json:"amount" validate:"required,gt=0"`
	UserID   string `json:"user_id,omitempty"`
}

// UsageStatsParams represents query parameters for usage statistics
type UsageStatsParams struct {
	StartDate string `json:"start_date,omitempty" form:"start_date"`
	EndDate   string `json:"end_date,omitempty" form:"end_date"`
	Resource  string `json:"resource,omitempty" form:"resource"`
	Action    string `json:"action,omitempty" form:"action"`
	Limit     int    `json:"limit" form:"limit" validate:"omitempty,gte=1,lte=1000"`
}

// Normalize sets default values for usage stats parameters
func (p *UsageStatsParams) Normalize() {
	if p.Limit < 1 {
		p.Limit = 100
	}
	if p.Limit > 1000 {
		p.Limit = 1000
	}
	if p.StartDate == "" {
		// Default to 30 days ago
		p.StartDate = time.Now().AddDate(0, 0, -30).Format(time.RFC3339)
	}
	if p.EndDate == "" {
		// Default to now
		p.EndDate = time.Now().Format(time.RFC3339)
	}
}

// UsageStats represents aggregated usage statistics
type UsageStats struct {
	TenantID          uuid.UUID              `json:"tenant_id"`
	Period            string                 `json:"period"`
	TotalStorage      int64                  `json:"total_storage"`
	TotalDocuments    int                    `json:"total_documents"`
	TotalUsers        int                    `json:"total_users"`
	TotalAPICall      int                    `json:"total_api_calls"`
	TotalBandwidth    int64                  `json:"total_bandwidth"`
	StorageByDay      map[string]int64       `json:"storage_by_day,omitempty"`
	APICallsByDay     map[string]int         `json:"api_calls_by_day,omitempty"`
	BandwidthByDay    map[string]int64       `json:"bandwidth_by_day,omitempty"`
	TopUsers          []UserUsageStats       `json:"top_users,omitempty"`
}

// UserUsageStats represents per-user usage statistics
type UserUsageStats struct {
	UserID        string `json:"user_id"`
	StorageUsed   int64  `json:"storage_used"`
	DocumentCount int    `json:"document_count"`
	APICallCount  int    `json:"api_call_count"`
}

// QuotaPlan represents a predefined quota plan
type QuotaPlan struct {
	Name              string `json:"name"`
	DisplayName       string `json:"display_name"`
	MaxStorage        int64  `json:"max_storage"`
	MaxDocuments      int    `json:"max_documents"`
	MaxUsers          int    `json:"max_users"`
	MaxAPICallsPerDay int    `json:"max_api_calls_per_day"`
	MaxFileSize       int64  `json:"max_file_size"`
	MaxBandwidth      int64  `json:"max_bandwidth"`
	Features          []string `json:"features"`
	PriceMonthly      float64  `json:"price_monthly"`
}

// GetPredefinedPlans returns predefined quota plans
func GetPredefinedPlans() []QuotaPlan {
	return []QuotaPlan{
		{
			Name:              "free",
			DisplayName:       "Free",
			MaxStorage:        1 * 1024 * 1024 * 1024, // 1 GB
			MaxDocuments:      100,
			MaxUsers:          3,
			MaxAPICallsPerDay: 1000,
			MaxFileSize:       10 * 1024 * 1024, // 10 MB
			MaxBandwidth:      5 * 1024 * 1024 * 1024, // 5 GB
			Features:          []string{"basic_storage", "basic_sharing"},
			PriceMonthly:      0,
		},
		{
			Name:              "basic",
			DisplayName:       "Basic",
			MaxStorage:        10 * 1024 * 1024 * 1024, // 10 GB
			MaxDocuments:      1000,
			MaxUsers:          10,
			MaxAPICallsPerDay: 10000,
			MaxFileSize:       50 * 1024 * 1024, // 50 MB
			MaxBandwidth:      50 * 1024 * 1024 * 1024, // 50 GB
			Features:          []string{"basic_storage", "basic_sharing", "ocr", "search"},
			PriceMonthly:      9.99,
		},
		{
			Name:              "pro",
			DisplayName:       "Professional",
			MaxStorage:        100 * 1024 * 1024 * 1024, // 100 GB
			MaxDocuments:      10000,
			MaxUsers:          50,
			MaxAPICallsPerDay: 100000,
			MaxFileSize:       500 * 1024 * 1024, // 500 MB
			MaxBandwidth:      500 * 1024 * 1024 * 1024, // 500 GB
			Features:          []string{"basic_storage", "basic_sharing", "ocr", "search", "advanced_sharing", "categorization", "audit"},
			PriceMonthly:      49.99,
		},
		{
			Name:              "enterprise",
			DisplayName:       "Enterprise",
			MaxStorage:        1024 * 1024 * 1024 * 1024, // 1 TB
			MaxDocuments:      100000,
			MaxUsers:          500,
			MaxAPICallsPerDay: 1000000,
			MaxFileSize:       2 * 1024 * 1024 * 1024, // 2 GB
			MaxBandwidth:      5 * 1024 * 1024 * 1024 * 1024, // 5 TB
			Features:          []string{"basic_storage", "basic_sharing", "ocr", "search", "advanced_sharing", "categorization", "audit", "sso", "priority_support"},
			PriceMonthly:      199.99,
		},
	}
}

package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// Document represents a document in the system
type Document struct {
	ID             uuid.UUID      `json:"id" db:"id"`
	TenantID       uuid.UUID      `json:"tenant_id" db:"tenant_id"`
	FolderID       sql.NullString `json:"folder_id,omitempty" db:"folder_id"`
	Name           string         `json:"name" db:"name"`
	Description    sql.NullString `json:"description,omitempty" db:"description"`
	FileType       string         `json:"file_type" db:"file_type"`
	FileSize       int64          `json:"file_size" db:"file_size"`
	MimeType       string         `json:"mime_type" db:"mime_type"`
	StoragePath    string         `json:"-" db:"storage_path"` // Don't expose storage path
	ThumbnailPath  sql.NullString `json:"-" db:"thumbnail_path"`
	Status         string         `json:"status" db:"status"`
	UploadedBy     string         `json:"uploaded_by" db:"uploaded_by"`
	CategoryID     sql.NullString `json:"category_id,omitempty" db:"category_id"`
	OCRStatus      string         `json:"ocr_status" db:"ocr_status"`
	SearchVector   sql.NullString `json:"-" db:"search_vector"` // PostgreSQL tsvector
	Version        int            `json:"version" db:"version"`
	CreatedAt      time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at" db:"updated_at"`
}

// DocumentVersion represents a version of a document
type DocumentVersion struct {
	ID            uuid.UUID `json:"id" db:"id"`
	DocumentID    uuid.UUID `json:"document_id" db:"document_id"`
	VersionNumber int       `json:"version_number" db:"version_number"`
	FileSize      int64     `json:"file_size" db:"file_size"`
	StoragePath   string    `json:"-" db:"storage_path"`
	UploadedBy    string    `json:"uploaded_by" db:"uploaded_by"`
	Comment       string    `json:"comment,omitempty" db:"comment"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

// Folder represents a folder/directory
type Folder struct {
	ID          uuid.UUID      `json:"id" db:"id"`
	TenantID    uuid.UUID      `json:"tenant_id" db:"tenant_id"`
	ParentID    sql.NullString `json:"parent_id,omitempty" db:"parent_id"`
	Name        string         `json:"name" db:"name"`
	Path        string         `json:"path" db:"path"`
	Description sql.NullString `json:"description,omitempty" db:"description"`
	Color       sql.NullString `json:"color,omitempty" db:"color"`
	Icon        sql.NullString `json:"icon,omitempty" db:"icon"`
	CreatedBy   string         `json:"created_by" db:"created_by"`
	CreatedAt   time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at" db:"updated_at"`
}

// Tag represents a document tag
type Tag struct {
	ID         uuid.UUID `json:"id" db:"id"`
	TenantID   uuid.UUID `json:"tenant_id" db:"tenant_id"`
	Name       string    `json:"name" db:"name"`
	Color      string    `json:"color" db:"color"`
	UsageCount int       `json:"usage_count" db:"usage_count"`
	CreatedBy  string    `json:"created_by" db:"created_by"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// DocumentTag represents the association between documents and tags
type DocumentTag struct {
	DocumentID uuid.UUID `json:"document_id" db:"document_id"`
	TagID      uuid.UUID `json:"tag_id" db:"tag_id"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// Category represents a document category
type Category struct {
	ID            uuid.UUID `json:"id" db:"id"`
	TenantID      uuid.UUID `json:"tenant_id" db:"tenant_id"`
	Name          string    `json:"name" db:"name"`
	Description   string    `json:"description,omitempty" db:"description"`
	Color         string    `json:"color" db:"color"`
	Icon          string    `json:"icon,omitempty" db:"icon"`
	DocumentCount int       `json:"document_count" db:"document_count"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// CreateDocumentRequest represents document creation request
type CreateDocumentRequest struct {
	Name        string   `json:"name" validate:"required,min=1,max=255"`
	Description string   `json:"description,omitempty" validate:"omitempty,max=1000"`
	FolderID    string   `json:"folder_id,omitempty" validate:"omitempty,uuid"`
	CategoryID  string   `json:"category_id,omitempty" validate:"omitempty,uuid"`
	Tags        []string `json:"tags,omitempty"`
}

// UpdateDocumentRequest represents document update request
type UpdateDocumentRequest struct {
	Name        string   `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Description string   `json:"description,omitempty" validate:"omitempty,max=1000"`
	FolderID    *string  `json:"folder_id,omitempty" validate:"omitempty,uuid"`
	CategoryID  *string  `json:"category_id,omitempty" validate:"omitempty,uuid"`
	Tags        []string `json:"tags,omitempty"`
}

// CreateFolderRequest represents folder creation request
type CreateFolderRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=100"`
	ParentID    string `json:"parent_id,omitempty" validate:"omitempty,uuid"`
	Description string `json:"description,omitempty" validate:"omitempty,max=500"`
	Color       string `json:"color,omitempty" validate:"omitempty,hexcolor"`
	Icon        string `json:"icon,omitempty" validate:"omitempty,max=50"`
}

// UpdateFolderRequest represents folder update request
type UpdateFolderRequest struct {
	Name        string `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	ParentID    string `json:"parent_id,omitempty" validate:"omitempty,uuid"`
	Description string `json:"description,omitempty" validate:"omitempty,max=500"`
	Color       string `json:"color,omitempty" validate:"omitempty,hexcolor"`
	Icon        string `json:"icon,omitempty" validate:"omitempty,max=50"`
}

// CreateTagRequest represents tag creation request
type CreateTagRequest struct {
	Name  string `json:"name" validate:"required,min=1,max=50"`
	Color string `json:"color" validate:"required,hexcolor"`
}

// CreateCategoryRequest represents category creation request
type CreateCategoryRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=100"`
	Description string `json:"description,omitempty" validate:"omitempty,max=500"`
	Color       string `json:"color" validate:"required,hexcolor"`
	Icon        string `json:"icon,omitempty" validate:"omitempty,max=50"`
}

// DocumentWithDetails includes document with related data
type DocumentWithDetails struct {
	Document
	Tags         []Tag     `json:"tags,omitempty"`
	Category     *Category `json:"category,omitempty"`
	FolderName   string    `json:"folder_name,omitempty"`
	UploadedByName string  `json:"uploaded_by_name,omitempty"`
}

// FolderWithContents includes folder with children and documents
type FolderWithContents struct {
	Folder
	SubFolders    []Folder   `json:"sub_folders,omitempty"`
	Documents     []Document `json:"documents,omitempty"`
	DocumentCount int        `json:"document_count"`
}

// ListDocumentsParams represents query parameters for listing documents
type ListDocumentsParams struct {
	FolderID   string `json:"folder_id,omitempty" form:"folder_id"`
	CategoryID string `json:"category_id,omitempty" form:"category_id"`
	Tags       string `json:"tags,omitempty" form:"tags"` // Comma-separated tag IDs
	Status     string `json:"status,omitempty" form:"status"`
	Search     string `json:"search,omitempty" form:"search"`
	Page       int    `json:"page" form:"page" validate:"omitempty,gte=1"`
	Limit      int    `json:"limit" form:"limit" validate:"omitempty,gte=1,lte=100"`
	SortBy     string `json:"sort_by,omitempty" form:"sort_by"`
	SortOrder  string `json:"sort_order,omitempty" form:"sort_order" validate:"omitempty,oneof=asc desc"`
}

// Normalize sets default values for list parameters
func (p *ListDocumentsParams) Normalize() {
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
func (p *ListDocumentsParams) GetOffset() int {
	return (p.Page - 1) * p.Limit
}

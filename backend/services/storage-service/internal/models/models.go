package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// FileMetadata represents file metadata stored in database
type FileMetadata struct {
	ID            uuid.UUID      `json:"id" db:"id"`
	TenantID      uuid.UUID      `json:"tenant_id" db:"tenant_id"`
	DocumentID    uuid.UUID      `json:"document_id" db:"document_id"`
	FileName      string         `json:"file_name" db:"file_name"`
	OriginalName  string         `json:"original_name" db:"original_name"`
	FileSize      int64          `json:"file_size" db:"file_size"`
	MimeType      string         `json:"mime_type" db:"mime_type"`
	FileType      string         `json:"file_type" db:"file_type"`
	BucketName    string         `json:"-" db:"bucket_name"`
	ObjectKey     string         `json:"-" db:"object_key"`
	ThumbnailKey  sql.NullString `json:"-" db:"thumbnail_key"`
	StoragePath   string         `json:"-" db:"storage_path"`
	Checksum      string         `json:"checksum" db:"checksum"`
	UploadedBy    string         `json:"uploaded_by" db:"uploaded_by"`
	IsEncrypted   bool           `json:"is_encrypted" db:"is_encrypted"`
	EncryptionKey sql.NullString `json:"-" db:"encryption_key"`
	CreatedAt     time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at" db:"updated_at"`
}

// UploadFileRequest represents file upload request
type UploadFileRequest struct {
	DocumentID  string `json:"document_id" validate:"required,uuid"`
	FileName    string `json:"file_name" validate:"required,min=1,max=255"`
	MimeType    string `json:"mime_type" validate:"required"`
	FileSize    int64  `json:"file_size" validate:"required,gt=0"`
	IsEncrypted bool   `json:"is_encrypted,omitempty"`
}

// UploadFileResponse represents file upload response
type UploadFileResponse struct {
	FileID       uuid.UUID `json:"file_id"`
	DocumentID   uuid.UUID `json:"document_id"`
	UploadURL    string    `json:"upload_url"`
	FileName     string    `json:"file_name"`
	ExpiresAt    time.Time `json:"expires_at"`
	StoragePath  string    `json:"storage_path"`
	ThumbnailURL string    `json:"thumbnail_url,omitempty"`
}

// DownloadFileRequest represents file download request
type DownloadFileRequest struct {
	FileID     uuid.UUID `json:"file_id"`
	Inline     bool      `json:"inline,omitempty"` // true for inline viewing, false for download
	ExpiryTime int       `json:"expiry_time,omitempty" validate:"omitempty,gte=60,lte=604800"` // seconds, default 3600 (1 hour)
}

// DownloadFileResponse represents file download response
type DownloadFileResponse struct {
	DownloadURL string    `json:"download_url"`
	FileName    string    `json:"file_name"`
	FileSize    int64     `json:"file_size"`
	MimeType    string    `json:"mime_type"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// PresignedURLRequest represents presigned URL generation request
type PresignedURLRequest struct {
	FileID     uuid.UUID `json:"file_id"`
	Operation  string    `json:"operation" validate:"required,oneof=upload download"` // upload or download
	ExpiryTime int       `json:"expiry_time,omitempty" validate:"omitempty,gte=60,lte=604800"` // seconds
}

// PresignedURLResponse represents presigned URL response
type PresignedURLResponse struct {
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expires_at"`
}

// DeleteFileRequest represents file deletion request
type DeleteFileRequest struct {
	FileID       uuid.UUID `json:"file_id"`
	DocumentID   uuid.UUID `json:"document_id"`
	HardDelete   bool      `json:"hard_delete,omitempty"` // true to delete from storage, false for soft delete
}

// ThumbnailRequest represents thumbnail generation/retrieval request
type ThumbnailRequest struct {
	FileID uuid.UUID `json:"file_id"`
	Width  int       `json:"width,omitempty" validate:"omitempty,gte=50,lte=1000"`
	Height int       `json:"height,omitempty" validate:"omitempty,gte=50,lte=1000"`
}

// ThumbnailResponse represents thumbnail response
type ThumbnailResponse struct {
	ThumbnailURL string `json:"thumbnail_url"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
}

// FileStats represents storage statistics
type FileStats struct {
	TotalFiles     int64 `json:"total_files"`
	TotalSize      int64 `json:"total_size"`
	TotalDocuments int64 `json:"total_documents"`
	ByFileType     map[string]FileTypeStats `json:"by_file_type"`
}

// FileTypeStats represents statistics by file type
type FileTypeStats struct {
	Count     int64 `json:"count"`
	TotalSize int64 `json:"total_size"`
}

// ListFilesParams represents query parameters for listing files
type ListFilesParams struct {
	DocumentID string `json:"document_id,omitempty" form:"document_id"`
	FileType   string `json:"file_type,omitempty" form:"file_type"`
	MimeType   string `json:"mime_type,omitempty" form:"mime_type"`
	Page       int    `json:"page" form:"page" validate:"omitempty,gte=1"`
	Limit      int    `json:"limit" form:"limit" validate:"omitempty,gte=1,lte=100"`
	SortBy     string `json:"sort_by,omitempty" form:"sort_by"`
	SortOrder  string `json:"sort_order,omitempty" form:"sort_order" validate:"omitempty,oneof=asc desc"`
}

// Normalize sets default values for list parameters
func (p *ListFilesParams) Normalize() {
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
func (p *ListFilesParams) GetOffset() int {
	return (p.Page - 1) * p.Limit
}

// BucketInfo represents MinIO bucket information
type BucketInfo struct {
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	Size      int64     `json:"size"`
	FileCount int64     `json:"file_count"`
}

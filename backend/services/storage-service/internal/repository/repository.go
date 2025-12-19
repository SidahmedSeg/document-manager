package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/SidahmedSeg/document-manager/backend/pkg/database"
	"github.com/SidahmedSeg/document-manager/backend/pkg/errors"
	"github.com/SidahmedSeg/document-manager/backend/services/storage-service/internal/models"
	"go.uber.org/zap"
)

// Repository handles file metadata database operations
type Repository struct {
	db     *database.DB
	logger *zap.Logger
}

// NewRepository creates a new file metadata repository
func NewRepository(db *database.DB, logger *zap.Logger) *Repository {
	return &Repository{
		db:     db,
		logger: logger,
	}
}

// CreateFileMetadata creates file metadata record
func (r *Repository) CreateFileMetadata(ctx context.Context, metadata *models.FileMetadata) error {
	query := `
		INSERT INTO file_metadata (
			id, tenant_id, document_id, file_name, original_name,
			file_size, mime_type, file_type, bucket_name, object_key,
			thumbnail_key, storage_path, checksum, uploaded_by,
			is_encrypted, encryption_key, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15, $16, $17, $18
		)`

	_, err := r.db.ExecContext(ctx, query,
		metadata.ID,
		metadata.TenantID,
		metadata.DocumentID,
		metadata.FileName,
		metadata.OriginalName,
		metadata.FileSize,
		metadata.MimeType,
		metadata.FileType,
		metadata.BucketName,
		metadata.ObjectKey,
		metadata.ThumbnailKey,
		metadata.StoragePath,
		metadata.Checksum,
		metadata.UploadedBy,
		metadata.IsEncrypted,
		metadata.EncryptionKey,
		metadata.CreatedAt,
		metadata.UpdatedAt,
	)

	if err != nil {
		r.logger.Error("failed to create file metadata", zap.Error(err))
		return errors.New(errors.ErrCodeInternal, "failed to create file metadata")
	}

	return nil
}

// GetFileMetadata retrieves file metadata by ID
func (r *Repository) GetFileMetadata(ctx context.Context, tenantID, fileID uuid.UUID) (*models.FileMetadata, error) {
	query := `
		SELECT id, tenant_id, document_id, file_name, original_name,
			file_size, mime_type, file_type, bucket_name, object_key,
			thumbnail_key, storage_path, checksum, uploaded_by,
			is_encrypted, encryption_key, created_at, updated_at
		FROM file_metadata
		WHERE id = $1 AND tenant_id = $2`

	var metadata models.FileMetadata
	err := r.db.QueryRowContext(ctx, query, fileID, tenantID).Scan(
		&metadata.ID,
		&metadata.TenantID,
		&metadata.DocumentID,
		&metadata.FileName,
		&metadata.OriginalName,
		&metadata.FileSize,
		&metadata.MimeType,
		&metadata.FileType,
		&metadata.BucketName,
		&metadata.ObjectKey,
		&metadata.ThumbnailKey,
		&metadata.StoragePath,
		&metadata.Checksum,
		&metadata.UploadedBy,
		&metadata.IsEncrypted,
		&metadata.EncryptionKey,
		&metadata.CreatedAt,
		&metadata.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.NotFoundf("file not found")
	}
	if err != nil {
		r.logger.Error("failed to get file metadata", zap.Error(err))
		return nil, errors.New(errors.ErrCodeInternal,"failed to get file metadata")
	}

	return &metadata, nil
}

// GetFileMetadataByDocumentID retrieves file metadata by document ID
func (r *Repository) GetFileMetadataByDocumentID(ctx context.Context, tenantID, documentID uuid.UUID) (*models.FileMetadata, error) {
	query := `
		SELECT id, tenant_id, document_id, file_name, original_name,
			file_size, mime_type, file_type, bucket_name, object_key,
			thumbnail_key, storage_path, checksum, uploaded_by,
			is_encrypted, encryption_key, created_at, updated_at
		FROM file_metadata
		WHERE document_id = $1 AND tenant_id = $2
		ORDER BY created_at DESC
		LIMIT 1`

	var metadata models.FileMetadata
	err := r.db.QueryRowContext(ctx, query, documentID, tenantID).Scan(
		&metadata.ID,
		&metadata.TenantID,
		&metadata.DocumentID,
		&metadata.FileName,
		&metadata.OriginalName,
		&metadata.FileSize,
		&metadata.MimeType,
		&metadata.FileType,
		&metadata.BucketName,
		&metadata.ObjectKey,
		&metadata.ThumbnailKey,
		&metadata.StoragePath,
		&metadata.Checksum,
		&metadata.UploadedBy,
		&metadata.IsEncrypted,
		&metadata.EncryptionKey,
		&metadata.CreatedAt,
		&metadata.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.NotFoundf("file not found for document")
	}
	if err != nil {
		r.logger.Error("failed to get file metadata by document ID", zap.Error(err))
		return nil, errors.New(errors.ErrCodeInternal,"failed to get file metadata")
	}

	return &metadata, nil
}

// ListFileMetadata retrieves files with filtering and pagination
func (r *Repository) ListFileMetadata(ctx context.Context, tenantID uuid.UUID, params *models.ListFilesParams) ([]models.FileMetadata, int64, error) {
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

	if params.FileType != "" {
		where = append(where, fmt.Sprintf("file_type = $%d", argPos))
		args = append(args, params.FileType)
		argPos++
	}

	if params.MimeType != "" {
		where = append(where, fmt.Sprintf("mime_type = $%d", argPos))
		args = append(args, params.MimeType)
		argPos++
	}

	whereClause := strings.Join(where, " AND ")

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM file_metadata WHERE %s", whereClause)
	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		r.logger.Error("failed to count files", zap.Error(err))
		return nil, 0, errors.New(errors.ErrCodeInternal,"failed to count files")
	}

	// Get files
	query := fmt.Sprintf(`
		SELECT id, tenant_id, document_id, file_name, original_name,
			file_size, mime_type, file_type, bucket_name, object_key,
			thumbnail_key, storage_path, checksum, uploaded_by,
			is_encrypted, encryption_key, created_at, updated_at
		FROM file_metadata
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
		r.logger.Error("failed to list files", zap.Error(err))
		return nil, 0, errors.New(errors.ErrCodeInternal,"failed to list files")
	}
	defer rows.Close()

	var files []models.FileMetadata
	for rows.Next() {
		var metadata models.FileMetadata
		err := rows.Scan(
			&metadata.ID,
			&metadata.TenantID,
			&metadata.DocumentID,
			&metadata.FileName,
			&metadata.OriginalName,
			&metadata.FileSize,
			&metadata.MimeType,
			&metadata.FileType,
			&metadata.BucketName,
			&metadata.ObjectKey,
			&metadata.ThumbnailKey,
			&metadata.StoragePath,
			&metadata.Checksum,
			&metadata.UploadedBy,
			&metadata.IsEncrypted,
			&metadata.EncryptionKey,
			&metadata.CreatedAt,
			&metadata.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("failed to scan file metadata", zap.Error(err))
			continue
		}
		files = append(files, metadata)
	}

	return files, total, nil
}

// UpdateFileMetadata updates file metadata
func (r *Repository) UpdateFileMetadata(ctx context.Context, tenantID, fileID uuid.UUID, updates map[string]interface{}) error {
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
	setClauses = append(setClauses, fmt.Sprintf("updated_at = NOW()"))

	// Add WHERE conditions
	args = append(args, fileID, tenantID)

	query := fmt.Sprintf(`
		UPDATE file_metadata
		SET %s
		WHERE id = $%d AND tenant_id = $%d`,
		strings.Join(setClauses, ", "),
		argPos,
		argPos+1,
	)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("failed to update file metadata", zap.Error(err))
		return errors.New(errors.ErrCodeInternal,"failed to update file metadata")
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.NotFoundf("file not found")
	}

	return nil
}

// DeleteFileMetadata deletes file metadata
func (r *Repository) DeleteFileMetadata(ctx context.Context, tenantID, fileID uuid.UUID) error {
	query := `DELETE FROM file_metadata WHERE id = $1 AND tenant_id = $2`

	result, err := r.db.ExecContext(ctx, query, fileID, tenantID)
	if err != nil {
		r.logger.Error("failed to delete file metadata", zap.Error(err))
		return errors.New(errors.ErrCodeInternal,"failed to delete file metadata")
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.NotFoundf("file not found")
	}

	return nil
}

// GetFileStats retrieves storage statistics for a tenant
func (r *Repository) GetFileStats(ctx context.Context, tenantID uuid.UUID) (*models.FileStats, error) {
	stats := &models.FileStats{
		ByFileType: make(map[string]models.FileTypeStats),
	}

	// Get overall stats
	query := `
		SELECT
			COUNT(*) as total_files,
			COALESCE(SUM(file_size), 0) as total_size,
			COUNT(DISTINCT document_id) as total_documents
		FROM file_metadata
		WHERE tenant_id = $1`

	err := r.db.QueryRowContext(ctx, query, tenantID).Scan(
		&stats.TotalFiles,
		&stats.TotalSize,
		&stats.TotalDocuments,
	)
	if err != nil {
		r.logger.Error("failed to get file stats", zap.Error(err))
		return nil, errors.New(errors.ErrCodeInternal,"failed to get file stats")
	}

	// Get stats by file type
	typeQuery := `
		SELECT
			file_type,
			COUNT(*) as count,
			COALESCE(SUM(file_size), 0) as total_size
		FROM file_metadata
		WHERE tenant_id = $1
		GROUP BY file_type`

	rows, err := r.db.QueryContext(ctx, typeQuery, tenantID)
	if err != nil {
		r.logger.Error("failed to get file type stats", zap.Error(err))
		return stats, nil // Return partial stats
	}
	defer rows.Close()

	for rows.Next() {
		var fileType string
		var typeStats models.FileTypeStats
		if err := rows.Scan(&fileType, &typeStats.Count, &typeStats.TotalSize); err != nil {
			continue
		}
		stats.ByFileType[fileType] = typeStats
	}

	return stats, nil
}

// UpdateThumbnailKey updates the thumbnail key for a file
func (r *Repository) UpdateThumbnailKey(ctx context.Context, tenantID, fileID uuid.UUID, thumbnailKey string) error {
	query := `
		UPDATE file_metadata
		SET thumbnail_key = $1, updated_at = NOW()
		WHERE id = $2 AND tenant_id = $3`

	result, err := r.db.ExecContext(ctx, query, thumbnailKey, fileID, tenantID)
	if err != nil {
		r.logger.Error("failed to update thumbnail key", zap.Error(err))
		return errors.New(errors.ErrCodeInternal,"failed to update thumbnail key")
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.NotFoundf("file not found")
	}

	return nil
}

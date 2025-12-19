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
	"github.com/SidahmedSeg/document-manager/backend/services/document-service/internal/models"
	"go.uber.org/zap"
)

// Repository handles database operations for documents
type Repository struct {
	db     *database.DB
	logger *zap.Logger
}

// NewRepository creates a new document repository
func NewRepository(db *database.DB, logger *zap.Logger) *Repository {
	return &Repository{
		db:     db,
		logger: logger,
	}
}

// Document operations

// CreateDocument creates a new document
func (r *Repository) CreateDocument(ctx context.Context, doc *models.Document) error {
	query := `
		INSERT INTO documents (
			id, tenant_id, folder_id, name, description, file_type, file_size,
			mime_type, storage_path, thumbnail_path, status, uploaded_by,
			category_id, ocr_status, version, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
	`

	_, err := r.db.ExecContext(ctx, query,
		doc.ID, doc.TenantID, doc.FolderID, doc.Name, doc.Description,
		doc.FileType, doc.FileSize, doc.MimeType, doc.StoragePath,
		doc.ThumbnailPath, doc.Status, doc.UploadedBy, doc.CategoryID,
		doc.OCRStatus, doc.Version, doc.CreatedAt, doc.UpdatedAt,
	)

	if err != nil {
		r.logger.Error("failed to create document", zap.Error(err))
		return errors.Wrap(errors.ErrCodeDatabase, "failed to create document", err)
	}

	return nil
}

// GetDocument retrieves a document by ID
func (r *Repository) GetDocument(ctx context.Context, tenantID, docID uuid.UUID) (*models.Document, error) {
	query := `
		SELECT id, tenant_id, folder_id, name, description, file_type, file_size,
		       mime_type, storage_path, thumbnail_path, status, uploaded_by,
		       category_id, ocr_status, version, created_at, updated_at
		FROM documents
		WHERE id = $1 AND tenant_id = $2
	`

	var doc models.Document
	err := r.db.QueryRowContext(ctx, query, docID, tenantID).Scan(
		&doc.ID, &doc.TenantID, &doc.FolderID, &doc.Name, &doc.Description,
		&doc.FileType, &doc.FileSize, &doc.MimeType, &doc.StoragePath,
		&doc.ThumbnailPath, &doc.Status, &doc.UploadedBy, &doc.CategoryID,
		&doc.OCRStatus, &doc.Version, &doc.CreatedAt, &doc.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.NotFoundf("document not found")
	}
	if err != nil {
		r.logger.Error("failed to get document", zap.Error(err))
		return nil, errors.Wrap(errors.ErrCodeDatabase, "failed to get document", err)
	}

	return &doc, nil
}

// ListDocuments retrieves documents with filtering and pagination
func (r *Repository) ListDocuments(ctx context.Context, tenantID uuid.UUID, params *models.ListDocumentsParams) ([]models.Document, int64, error) {
	// Build WHERE clause
	whereClauses := []string{"tenant_id = $1"}
	args := []interface{}{tenantID}
	argPos := 2

	if params.FolderID != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("folder_id = $%d", argPos))
		args = append(args, params.FolderID)
		argPos++
	}

	if params.CategoryID != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("category_id = $%d", argPos))
		args = append(args, params.CategoryID)
		argPos++
	}

	if params.Status != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("status = $%d", argPos))
		args = append(args, params.Status)
		argPos++
	}

	if params.Search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("(name ILIKE $%d OR description ILIKE $%d)", argPos, argPos))
		args = append(args, "%"+params.Search+"%")
		argPos++
	}

	whereClause := strings.Join(whereClauses, " AND ")

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM documents WHERE %s", whereClause)
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, errors.Wrap(errors.ErrCodeDatabase, "failed to count documents", err)
	}

	// Get documents
	query := fmt.Sprintf(`
		SELECT id, tenant_id, folder_id, name, description, file_type, file_size,
		       mime_type, storage_path, thumbnail_path, status, uploaded_by,
		       category_id, ocr_status, version, created_at, updated_at
		FROM documents
		WHERE %s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, params.SortBy, params.SortOrder, argPos, argPos+1)

	args = append(args, params.Limit, params.GetOffset())

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("failed to list documents", zap.Error(err))
		return nil, 0, errors.Wrap(errors.ErrCodeDatabase, "failed to list documents", err)
	}
	defer rows.Close()

	var documents []models.Document
	for rows.Next() {
		var doc models.Document
		err := rows.Scan(
			&doc.ID, &doc.TenantID, &doc.FolderID, &doc.Name, &doc.Description,
			&doc.FileType, &doc.FileSize, &doc.MimeType, &doc.StoragePath,
			&doc.ThumbnailPath, &doc.Status, &doc.UploadedBy, &doc.CategoryID,
			&doc.OCRStatus, &doc.Version, &doc.CreatedAt, &doc.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("failed to scan document", zap.Error(err))
			continue
		}
		documents = append(documents, doc)
	}

	return documents, total, nil
}

// UpdateDocument updates a document
func (r *Repository) UpdateDocument(ctx context.Context, tenantID, docID uuid.UUID, req *models.UpdateDocumentRequest) error {
	query := `
		UPDATE documents
		SET name = COALESCE(NULLIF($1, ''), name),
		    description = COALESCE(NULLIF($2, ''), description),
		    folder_id = COALESCE($3, folder_id),
		    category_id = COALESCE($4, category_id),
		    updated_at = $5
		WHERE id = $6 AND tenant_id = $7
	`

	var folderID, categoryID interface{}
	if req.FolderID != nil {
		folderID = *req.FolderID
	}
	if req.CategoryID != nil {
		categoryID = *req.CategoryID
	}

	_, err := r.db.ExecContext(ctx, query,
		req.Name, req.Description, folderID, categoryID,
		time.Now(), docID, tenantID,
	)

	if err != nil {
		r.logger.Error("failed to update document", zap.Error(err))
		return errors.Wrap(errors.ErrCodeDatabase, "failed to update document", err)
	}

	return nil
}

// DeleteDocument deletes a document
func (r *Repository) DeleteDocument(ctx context.Context, tenantID, docID uuid.UUID) error {
	query := `DELETE FROM documents WHERE id = $1 AND tenant_id = $2`

	result, err := r.db.ExecContext(ctx, query, docID, tenantID)
	if err != nil {
		r.logger.Error("failed to delete document", zap.Error(err))
		return errors.Wrap(errors.ErrCodeDatabase, "failed to delete document", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.NotFoundf("document not found")
	}

	return nil
}

// Folder operations

// CreateFolder creates a new folder
func (r *Repository) CreateFolder(ctx context.Context, folder *models.Folder) error {
	query := `
		INSERT INTO folders (id, tenant_id, parent_id, name, path, description, color, icon, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.db.ExecContext(ctx, query,
		folder.ID, folder.TenantID, folder.ParentID, folder.Name, folder.Path,
		folder.Description, folder.Color, folder.Icon, folder.CreatedBy,
		folder.CreatedAt, folder.UpdatedAt,
	)

	if err != nil {
		r.logger.Error("failed to create folder", zap.Error(err))
		return errors.Wrap(errors.ErrCodeDatabase, "failed to create folder", err)
	}

	return nil
}

// GetFolder retrieves a folder by ID
func (r *Repository) GetFolder(ctx context.Context, tenantID, folderID uuid.UUID) (*models.Folder, error) {
	query := `
		SELECT id, tenant_id, parent_id, name, path, description, color, icon, created_by, created_at, updated_at
		FROM folders
		WHERE id = $1 AND tenant_id = $2
	`

	var folder models.Folder
	err := r.db.QueryRowContext(ctx, query, folderID, tenantID).Scan(
		&folder.ID, &folder.TenantID, &folder.ParentID, &folder.Name, &folder.Path,
		&folder.Description, &folder.Color, &folder.Icon, &folder.CreatedBy,
		&folder.CreatedAt, &folder.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.NotFoundf("folder not found")
	}
	if err != nil {
		r.logger.Error("failed to get folder", zap.Error(err))
		return nil, errors.Wrap(errors.ErrCodeDatabase, "failed to get folder", err)
	}

	return &folder, nil
}

// ListFolders retrieves all folders in a tenant
func (r *Repository) ListFolders(ctx context.Context, tenantID uuid.UUID, parentID *string) ([]models.Folder, error) {
	var query string
	var args []interface{}

	if parentID != nil && *parentID != "" {
		query = `
			SELECT id, tenant_id, parent_id, name, path, description, color, icon, created_by, created_at, updated_at
			FROM folders
			WHERE tenant_id = $1 AND parent_id = $2
			ORDER BY name ASC
		`
		args = []interface{}{tenantID, *parentID}
	} else {
		query = `
			SELECT id, tenant_id, parent_id, name, path, description, color, icon, created_by, created_at, updated_at
			FROM folders
			WHERE tenant_id = $1 AND parent_id IS NULL
			ORDER BY name ASC
		`
		args = []interface{}{tenantID}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("failed to list folders", zap.Error(err))
		return nil, errors.Wrap(errors.ErrCodeDatabase, "failed to list folders", err)
	}
	defer rows.Close()

	var folders []models.Folder
	for rows.Next() {
		var folder models.Folder
		err := rows.Scan(
			&folder.ID, &folder.TenantID, &folder.ParentID, &folder.Name, &folder.Path,
			&folder.Description, &folder.Color, &folder.Icon, &folder.CreatedBy,
			&folder.CreatedAt, &folder.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("failed to scan folder", zap.Error(err))
			continue
		}
		folders = append(folders, folder)
	}

	return folders, nil
}

// DeleteFolder deletes a folder
func (r *Repository) DeleteFolder(ctx context.Context, tenantID, folderID uuid.UUID) error {
	query := `DELETE FROM folders WHERE id = $1 AND tenant_id = $2`

	result, err := r.db.ExecContext(ctx, query, folderID, tenantID)
	if err != nil {
		r.logger.Error("failed to delete folder", zap.Error(err))
		return errors.Wrap(errors.ErrCodeDatabase, "failed to delete folder", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.NotFoundf("folder not found")
	}

	return nil
}

// Tag operations

// CreateTag creates a new tag
func (r *Repository) CreateTag(ctx context.Context, tag *models.Tag) error {
	query := `
		INSERT INTO tags (id, tenant_id, name, color, created_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.ExecContext(ctx, query,
		tag.ID, tag.TenantID, tag.Name, tag.Color, tag.CreatedBy, tag.CreatedAt,
	)

	if err != nil {
		r.logger.Error("failed to create tag", zap.Error(err))
		return errors.Wrap(errors.ErrCodeDatabase, "failed to create tag", err)
	}

	return nil
}

// ListTags retrieves all tags in a tenant
func (r *Repository) ListTags(ctx context.Context, tenantID uuid.UUID) ([]models.Tag, error) {
	query := `
		SELECT id, tenant_id, name, color, usage_count, created_by, created_at
		FROM tags
		WHERE tenant_id = $1
		ORDER BY name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		r.logger.Error("failed to list tags", zap.Error(err))
		return nil, errors.Wrap(errors.ErrCodeDatabase, "failed to list tags", err)
	}
	defer rows.Close()

	var tags []models.Tag
	for rows.Next() {
		var tag models.Tag
		err := rows.Scan(&tag.ID, &tag.TenantID, &tag.Name, &tag.Color, &tag.UsageCount, &tag.CreatedBy, &tag.CreatedAt)
		if err != nil {
			r.logger.Error("failed to scan tag", zap.Error(err))
			continue
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

// AddTagToDocument adds a tag to a document
func (r *Repository) AddTagToDocument(ctx context.Context, documentID, tagID uuid.UUID) error {
	query := `
		INSERT INTO document_tags (document_id, tag_id, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (document_id, tag_id) DO NOTHING
	`

	_, err := r.db.ExecContext(ctx, query, documentID, tagID, time.Now())
	if err != nil {
		r.logger.Error("failed to add tag to document", zap.Error(err))
		return errors.Wrap(errors.ErrCodeDatabase, "failed to add tag", err)
	}

	return nil
}

// RemoveTagFromDocument removes a tag from a document
func (r *Repository) RemoveTagFromDocument(ctx context.Context, documentID, tagID uuid.UUID) error {
	query := `DELETE FROM document_tags WHERE document_id = $1 AND tag_id = $2`

	_, err := r.db.ExecContext(ctx, query, documentID, tagID)
	if err != nil {
		r.logger.Error("failed to remove tag from document", zap.Error(err))
		return errors.Wrap(errors.ErrCodeDatabase, "failed to remove tag", err)
	}

	return nil
}

// GetDocumentTags retrieves all tags for a document
func (r *Repository) GetDocumentTags(ctx context.Context, documentID uuid.UUID) ([]models.Tag, error) {
	query := `
		SELECT t.id, t.tenant_id, t.name, t.color, t.usage_count, t.created_by, t.created_at
		FROM tags t
		INNER JOIN document_tags dt ON t.id = dt.tag_id
		WHERE dt.document_id = $1
		ORDER BY t.name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, documentID)
	if err != nil {
		r.logger.Error("failed to get document tags", zap.Error(err))
		return nil, errors.Wrap(errors.ErrCodeDatabase, "failed to get document tags", err)
	}
	defer rows.Close()

	var tags []models.Tag
	for rows.Next() {
		var tag models.Tag
		err := rows.Scan(&tag.ID, &tag.TenantID, &tag.Name, &tag.Color, &tag.UsageCount, &tag.CreatedBy, &tag.CreatedAt)
		if err != nil {
			r.logger.Error("failed to scan tag", zap.Error(err))
			continue
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

// Category operations

// CreateCategory creates a new category
func (r *Repository) CreateCategory(ctx context.Context, category *models.Category) error {
	query := `
		INSERT INTO categories (id, tenant_id, name, description, color, icon, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(ctx, query,
		category.ID, category.TenantID, category.Name, category.Description,
		category.Color, category.Icon, category.CreatedAt, category.UpdatedAt,
	)

	if err != nil {
		r.logger.Error("failed to create category", zap.Error(err))
		return errors.Wrap(errors.ErrCodeDatabase, "failed to create category", err)
	}

	return nil
}

// ListCategories retrieves all categories in a tenant
func (r *Repository) ListCategories(ctx context.Context, tenantID uuid.UUID) ([]models.Category, error) {
	query := `
		SELECT id, tenant_id, name, description, color, icon, document_count, created_at, updated_at
		FROM categories
		WHERE tenant_id = $1
		ORDER BY name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		r.logger.Error("failed to list categories", zap.Error(err))
		return nil, errors.Wrap(errors.ErrCodeDatabase, "failed to list categories", err)
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var cat models.Category
		err := rows.Scan(&cat.ID, &cat.TenantID, &cat.Name, &cat.Description, &cat.Color, &cat.Icon, &cat.DocumentCount, &cat.CreatedAt, &cat.UpdatedAt)
		if err != nil {
			r.logger.Error("failed to scan category", zap.Error(err))
			continue
		}
		categories = append(categories, cat)
	}

	return categories, nil
}

package service

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/SidahmedSeg/document-manager/backend/pkg/cache"
	"github.com/SidahmedSeg/document-manager/backend/pkg/config"
	"github.com/SidahmedSeg/document-manager/backend/pkg/errors"
	"github.com/SidahmedSeg/document-manager/backend/pkg/logger"
	"github.com/SidahmedSeg/document-manager/backend/pkg/middleware"
	"github.com/SidahmedSeg/document-manager/backend/services/storage-service/internal/models"
	"github.com/SidahmedSeg/document-manager/backend/services/storage-service/internal/repository"
	"go.uber.org/zap"
)

const (
	fileCacheTTL         = 30 * time.Minute
	presignedURLExpiry   = 1 * time.Hour
	defaultThumbnailSize = 300
	maxFileSize          = 100 * 1024 * 1024 // 100MB
)

// Service handles storage business logic
type Service struct {
	repo        *repository.Repository
	cache       *cache.Cache
	minioClient *minio.Client
	bucketName  string
	logger      *zap.Logger
}

// NewService creates a new storage service
func NewService(repo *repository.Repository, cache *cache.Cache, cfg config.MinIOConfig, logger *zap.Logger) (*Service, error) {
	// Initialize MinIO client
	minioClient, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize MinIO client: %w", err)
	}

	return &Service{
		repo:        repo,
		cache:       cache,
		minioClient: minioClient,
		bucketName:  cfg.BucketName,
		logger:      logger,
	}, nil
}

// EnsureBucket ensures the bucket exists, creates if not
func (s *Service) EnsureBucket(ctx context.Context) error {
	exists, err := s.minioClient.BucketExists(ctx, s.bucketName)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		err = s.minioClient.MakeBucket(ctx, s.bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
		logger.InfoContext(ctx, "created MinIO bucket", zap.String("bucket", s.bucketName))
	}

	return nil
}

// UploadFile handles file upload
func (s *Service) UploadFile(ctx context.Context, req *models.UploadFileRequest, file io.Reader) (*models.UploadFileResponse, error) {
	tenantID := getTenantID(ctx)
	userID := middleware.GetUserID(ctx)

	// Validate file size
	if req.FileSize > maxFileSize {
		return nil, errors.Validationf("file size exceeds maximum allowed size of %d bytes", maxFileSize)
	}

	// Parse document ID
	documentID, err := uuid.Parse(req.DocumentID)
	if err != nil {
		return nil, errors.Validationf("invalid document_id")
	}

	// Generate unique file ID and object key
	fileID := uuid.New()
	ext := filepath.Ext(req.FileName)
	fileType := getFileType(req.MimeType)
	objectKey := fmt.Sprintf("%s/%s/%s%s", tenantID.String(), documentID.String(), fileID.String(), ext)

	// Calculate checksum while uploading
	hasher := sha256.New()
	teeReader := io.TeeReader(file, hasher)

	// Upload to MinIO
	uploadInfo, err := s.minioClient.PutObject(
		ctx,
		s.bucketName,
		objectKey,
		teeReader,
		req.FileSize,
		minio.PutObjectOptions{
			ContentType: req.MimeType,
			UserMetadata: map[string]string{
				"tenant-id":   tenantID.String(),
				"document-id": documentID.String(),
				"uploaded-by": userID,
			},
		},
	)
	if err != nil {
		s.logger.Error("failed to upload file to MinIO", zap.Error(err))
		return nil, errors.New(errors.ErrCodeInternal,"failed to upload file")
	}

	// Calculate checksum
	checksum := fmt.Sprintf("%x", hasher.Sum(nil))

	// Save file metadata
	metadata := &models.FileMetadata{
		ID:           fileID,
		TenantID:     tenantID,
		DocumentID:   documentID,
		FileName:     fmt.Sprintf("%s%s", fileID.String(), ext),
		OriginalName: req.FileName,
		FileSize:     uploadInfo.Size,
		MimeType:     req.MimeType,
		FileType:     fileType,
		BucketName:   s.bucketName,
		ObjectKey:    objectKey,
		StoragePath:  objectKey,
		Checksum:     checksum,
		UploadedBy:   userID,
		IsEncrypted:  req.IsEncrypted,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.repo.CreateFileMetadata(ctx, metadata); err != nil {
		// Rollback: delete file from MinIO
		_ = s.minioClient.RemoveObject(ctx, s.bucketName, objectKey, minio.RemoveObjectOptions{})
		return nil, err
	}

	// Generate presigned URL for download
	presignedURL, err := s.minioClient.PresignedGetObject(
		ctx,
		s.bucketName,
		objectKey,
		presignedURLExpiry,
		nil,
	)
	if err != nil {
		s.logger.Error("failed to generate presigned URL", zap.Error(err))
	}

	logger.InfoContext(ctx, "file uploaded",
		zap.String("file_id", fileID.String()),
		zap.String("document_id", documentID.String()),
		zap.Int64("size", uploadInfo.Size),
	)

	return &models.UploadFileResponse{
		FileID:      fileID,
		DocumentID:  documentID,
		UploadURL:   presignedURL.String(),
		FileName:    metadata.FileName,
		ExpiresAt:   time.Now().Add(presignedURLExpiry),
		StoragePath: objectKey,
	}, nil
}

// GetPresignedUploadURL generates a presigned URL for direct upload
func (s *Service) GetPresignedUploadURL(ctx context.Context, req *models.UploadFileRequest) (*models.PresignedURLResponse, error) {
	tenantID := getTenantID(ctx)

	// Parse document ID
	documentID, err := uuid.Parse(req.DocumentID)
	if err != nil {
		return nil, errors.Validationf("invalid document_id")
	}

	// Generate object key
	fileID := uuid.New()
	ext := filepath.Ext(req.FileName)
	objectKey := fmt.Sprintf("%s/%s/%s%s", tenantID.String(), documentID.String(), fileID.String(), ext)

	// Generate presigned URL for upload
	presignedURL, err := s.minioClient.PresignedPutObject(
		ctx,
		s.bucketName,
		objectKey,
		presignedURLExpiry,
	)
	if err != nil {
		s.logger.Error("failed to generate presigned upload URL", zap.Error(err))
		return nil, errors.New(errors.ErrCodeInternal,"failed to generate upload URL")
	}

	return &models.PresignedURLResponse{
		URL:       presignedURL.String(),
		ExpiresAt: time.Now().Add(presignedURLExpiry),
	}, nil
}

// DownloadFile generates a download URL for a file
func (s *Service) DownloadFile(ctx context.Context, fileID uuid.UUID, inline bool, expiryTime int) (*models.DownloadFileResponse, error) {
	tenantID := getTenantID(ctx)

	// Get file metadata
	metadata, err := s.repo.GetFileMetadata(ctx, tenantID, fileID)
	if err != nil {
		return nil, err
	}

	// Set expiry time (default 1 hour)
	if expiryTime == 0 {
		expiryTime = 3600
	}
	expiry := time.Duration(expiryTime) * time.Second

	// Generate presigned URL
	reqParams := make(url.Values)
	if inline {
		reqParams.Set("response-content-disposition", fmt.Sprintf("inline; filename=\"%s\"", metadata.OriginalName))
	} else {
		reqParams.Set("response-content-disposition", fmt.Sprintf("attachment; filename=\"%s\"", metadata.OriginalName))
	}

	presignedURL, err := s.minioClient.PresignedGetObject(
		ctx,
		s.bucketName,
		metadata.ObjectKey,
		expiry,
		reqParams,
	)
	if err != nil {
		s.logger.Error("failed to generate download URL", zap.Error(err))
		return nil, errors.New(errors.ErrCodeInternal,"failed to generate download URL")
	}

	return &models.DownloadFileResponse{
		DownloadURL: presignedURL.String(),
		FileName:    metadata.OriginalName,
		FileSize:    metadata.FileSize,
		MimeType:    metadata.MimeType,
		ExpiresAt:   time.Now().Add(expiry),
	}, nil
}

// DeleteFile deletes a file
func (s *Service) DeleteFile(ctx context.Context, fileID uuid.UUID, hardDelete bool) error {
	tenantID := getTenantID(ctx)

	// Get file metadata
	metadata, err := s.repo.GetFileMetadata(ctx, tenantID, fileID)
	if err != nil {
		return err
	}

	// Delete from MinIO if hard delete
	if hardDelete {
		err = s.minioClient.RemoveObject(ctx, s.bucketName, metadata.ObjectKey, minio.RemoveObjectOptions{})
		if err != nil {
			s.logger.Error("failed to delete file from MinIO", zap.Error(err))
			return errors.New(errors.ErrCodeInternal,"failed to delete file from storage")
		}

		// Delete thumbnail if exists
		if metadata.ThumbnailKey.Valid {
			_ = s.minioClient.RemoveObject(ctx, s.bucketName, metadata.ThumbnailKey.String, minio.RemoveObjectOptions{})
		}
	}

	// Delete metadata from database
	if err := s.repo.DeleteFileMetadata(ctx, tenantID, fileID); err != nil {
		return err
	}

	// Invalidate cache
	cacheKey := cache.TenantKey(tenantID.String(), "file", fileID.String())
	_ = s.cache.Delete(ctx, cacheKey)

	logger.InfoContext(ctx, "file deleted",
		zap.String("file_id", fileID.String()),
		zap.Bool("hard_delete", hardDelete),
	)

	return nil
}

// GetFileMetadata retrieves file metadata
func (s *Service) GetFileMetadata(ctx context.Context, fileID uuid.UUID) (*models.FileMetadata, error) {
	tenantID := getTenantID(ctx)

	// Try cache first
	cacheKey := cache.TenantKey(tenantID.String(), "file", fileID.String())
	var metadata models.FileMetadata
	if err := s.cache.Get(ctx, cacheKey, &metadata); err == nil {
		return &metadata, nil
	}

	// Fetch from database
	metadataPtr, err := s.repo.GetFileMetadata(ctx, tenantID, fileID)
	if err != nil {
		return nil, err
	}

	// Cache for future requests
	_ = s.cache.Set(ctx, cacheKey, metadataPtr, fileCacheTTL)

	return metadataPtr, nil
}

// ListFiles retrieves files with filtering
func (s *Service) ListFiles(ctx context.Context, params *models.ListFilesParams) ([]models.FileMetadata, int64, error) {
	tenantID := getTenantID(ctx)

	params.Normalize()

	files, total, err := s.repo.ListFileMetadata(ctx, tenantID, params)
	if err != nil {
		return nil, 0, err
	}

	return files, total, nil
}

// GetFileStats retrieves storage statistics
func (s *Service) GetFileStats(ctx context.Context) (*models.FileStats, error) {
	tenantID := getTenantID(ctx)

	stats, err := s.repo.GetFileStats(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// Helper functions

func getTenantID(ctx context.Context) uuid.UUID {
	tenantIDStr := middleware.GetTenantID(ctx)
	tenantID, _ := uuid.Parse(tenantIDStr)
	return tenantID
}

func getFileType(mimeType string) string {
	parts := strings.Split(mimeType, "/")
	if len(parts) > 0 {
		return parts[0]
	}
	return "application"
}

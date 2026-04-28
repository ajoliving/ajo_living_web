/*
 * Upload and media business logic.
 * 1. Validate media upload requests and create presign targets.
 * 2. Persist uploaded media metadata for later listing binding.
 */
package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/model"
	"ajoliving_web/http_service/internal/utils"
	"gorm.io/gorm"
)

// 1. UploadService handles media upload flow.
type UploadService struct {
	runtime *Runtime
}

// 2. PresignParams defines the upload presign input.
type PresignParams struct {
	UserID       int64
	FileName     string
	MimeType     string
	FileSize     int64
	ObjectPrefix string
}

// 3. CompleteUploadParams defines the upload complete input.
type CompleteUploadParams struct {
	UserID         int64
	ObjectKey      string
	MimeType       string
	FileSize       int64
	Width          *int
	Height         *int
	ChecksumSHA256 string
}

// 4. CompleteUploadResult defines the upload complete output.
type CompleteUploadResult struct {
	MediaAssetID    string    `json:"media_asset_id"`
	StorageProvider string    `json:"storage_provider"`
	BucketName      string    `json:"bucket_name"`
	ObjectKey       string    `json:"object_key"`
	MimeType        string    `json:"mime_type"`
	Width           *int      `json:"width"`
	Height          *int      `json:"height"`
	FileSize        int64     `json:"file_size"`
	ChecksumSHA256  string    `json:"checksum_sha256"`
	URL             string    `json:"url"`
	InUse           bool      `json:"in_use"`
	CreatedAt       time.Time `json:"created_at"`
}

// 5. NewUploadService creates an upload service instance.
func NewUploadService(runtime *Runtime) *UploadService {
	return &UploadService{runtime: runtime}
}

// 6. Presign creates an upload target through the configured storage provider.
func (s *UploadService) Presign(ctx context.Context, params PresignParams) (*PresignUploadResult, error) {
	if strings.TrimSpace(params.FileName) == "" || strings.TrimSpace(params.MimeType) == "" || params.FileSize <= 0 {
		return nil, errcode.New(errcode.CodeValidationError, "invalid upload payload")
	}

	if !strings.HasPrefix(strings.ToLower(params.MimeType), "image/") {
		return nil, errcode.New(errcode.CodeValidationError, "only image uploads are supported")
	}

	return s.runtime.StorageProvider.PresignUpload(ctx, PresignUploadInput{
		FileName:     params.FileName,
		MimeType:     params.MimeType,
		FileSize:     params.FileSize,
		ObjectPrefix: params.ObjectPrefix,
	})
}

// 7. CompleteUpload persists media asset metadata after upload.
func (s *UploadService) CompleteUpload(ctx context.Context, params CompleteUploadParams) (*CompleteUploadResult, error) {
	if strings.TrimSpace(params.ObjectKey) == "" || strings.TrimSpace(params.MimeType) == "" || params.FileSize <= 0 {
		return nil, errcode.New(errcode.CodeValidationError, "invalid upload complete payload")
	}
	if !strings.HasPrefix(strings.TrimSpace(params.ObjectKey), mediaObjectPrefix) {
		return nil, errcode.New(errcode.CodeValidationError, "object_key is invalid")
	}

	objectInfo, err := s.runtime.StorageProvider.HeadObject(ctx, params.ObjectKey)
	if err != nil {
		if errors.Is(err, errStorageObjectNotFound) {
			return nil, errcode.New(errcode.CodeValidationError, "uploaded object not found in storage")
		}
		return nil, errcode.New(errcode.CodeInternalError, "failed to verify uploaded object")
	}
	if objectInfo.ContentLength > 0 && objectInfo.ContentLength != params.FileSize {
		return nil, errcode.New(errcode.CodeValidationError, "file_size does not match uploaded object")
	}
	if contentType := strings.TrimSpace(strings.ToLower(objectInfo.ContentType)); contentType != "" && contentType != strings.ToLower(params.MimeType) {
		return nil, errcode.New(errcode.CodeValidationError, "mime_type does not match uploaded object")
	}

	var existing model.MediaAsset
	err = s.runtime.DB.WithContext(ctx).
		Where("bucket_name = ? AND object_key = ? AND created_by = ?", s.runtime.Config.StorageBucket, params.ObjectKey, params.UserID).
		First(&existing).Error
	if err == nil {
		return s.buildMediaAssetResult(ctx, &existing)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load media asset")
	}

	media := model.MediaAsset{
		PublicID:        utils.NewPublicID(),
		StorageProvider: s.runtime.Config.StorageProvider,
		BucketName:      s.runtime.Config.StorageBucket,
		ObjectKey:       params.ObjectKey,
		MimeType:        params.MimeType,
		Width:           params.Width,
		Height:          params.Height,
		FileSize:        params.FileSize,
		ChecksumSHA256:  params.ChecksumSHA256,
		CreatedBy:       &params.UserID,
	}

	if err := s.runtime.DB.WithContext(ctx).Create(&media).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to save media asset")
	}

	return s.buildMediaAssetResult(ctx, &media)
}

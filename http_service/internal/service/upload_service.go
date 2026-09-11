/*
 * Upload and media business logic.
 * 1. Validate media upload requests and create presign targets.
 * 2. Persist uploaded media metadata for later listing binding.
 */
package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strconv"
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
	UploadToken    string
	MimeType       string
	FileSize       int64
	Width          *int
	Height         *int
	ChecksumSHA256 string
}

// 4. UploadCallbackParams defines the OSS upload callback payload.
type UploadCallbackParams struct {
	BucketName    string
	ObjectKey     string
	MimeType      string
	FileSize      int64
	ETag          string
	CallbackToken string
}

// 5. UploadCallbackResult defines the callback acknowledgement payload.
type UploadCallbackResult struct {
	BucketName string `json:"bucket_name"`
	ObjectKey  string `json:"object_key"`
	Accepted   bool   `json:"accepted"`
}

// 6. CompleteUploadResult defines the upload complete output.
type CompleteUploadResult struct {
	MediaAssetID     string    `json:"media_asset_id"`
	StorageProvider  string    `json:"storage_provider"`
	BucketName       string    `json:"bucket_name"`
	ObjectKey        string    `json:"object_key"`
	MimeType         string    `json:"mime_type"`
	Width            *int      `json:"width"`
	Height           *int      `json:"height"`
	FileSize         int64     `json:"file_size"`
	ChecksumSHA256   string    `json:"checksum_sha256"`
	URL              string    `json:"url"`
	InUse            bool      `json:"in_use"`
	CreatedAt        time.Time `json:"created_at"`
	ProcessingStatus string    `json:"processing_status"`
	ScanStatus       string    `json:"scan_status"`
	RejectionReason  string    `json:"rejection_reason,omitempty"`
}

type uploadTokenClaims struct {
	UserID    int64  `json:"user_id"`
	ObjectKey string `json:"object_key"`
	MimeType  string `json:"mime_type"`
	FileSize  int64  `json:"file_size"`
	UploadID  string `json:"upload_id,omitempty"`
	ExpiresAt int64  `json:"exp"`
}

const maxChatAttachmentSize int64 = 50 * 1024 * 1024

// 7. NewUploadService creates an upload service instance.
func NewUploadService(runtime *Runtime) *UploadService {
	return &UploadService{runtime: runtime}
}

// 8. Presign creates an upload target through the configured storage provider.
func (s *UploadService) Presign(ctx context.Context, params PresignParams) (*PresignUploadResult, error) {
	if strings.TrimSpace(params.FileName) == "" || strings.TrimSpace(params.MimeType) == "" || params.FileSize <= 0 {
		return nil, errcode.New(errcode.CodeValidationError, "invalid upload payload")
	}
	if err := validateChatAttachmentSize(params.FileSize, params.ObjectPrefix); err != nil {
		return nil, err
	}

	if !isSupportedPresignUpload(params.MimeType, params.ObjectPrefix) {
		return nil, errcode.New(errcode.CodeValidationError, "only image or video uploads are supported")
	}

	result, err := s.runtime.StorageProvider.PresignUpload(ctx, PresignUploadInput{
		FileName:     params.FileName,
		MimeType:     params.MimeType,
		FileSize:     params.FileSize,
		ObjectPrefix: params.ObjectPrefix,
	})
	if err != nil {
		return nil, err
	}

	expiresAt := s.runtime.Now().Add(s.uploadTokenTTL()).Unix()
	token, err := s.signUploadToken(uploadTokenClaims{
		UserID:    params.UserID,
		ObjectKey: strings.TrimSpace(result.ObjectKey),
		MimeType:  strings.TrimSpace(params.MimeType),
		FileSize:  params.FileSize,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to prepare upload token")
	}
	result.UploadToken = token

	return result, nil
}

// 8.1 validateChatAttachmentSize enforces the chat attachment size limit for all upload paths.
func validateChatAttachmentSize(fileSize int64, objectPrefix string) error {
	if normalizeMediaObjectPrefix(objectPrefix) == chatMediaObjectPrefix && fileSize > maxChatAttachmentSize {
		return errcode.New(errcode.CodeValidationError, "chat attachment exceeds 50 MB")
	}
	return nil
}

// 9. isSupportedPresignUpload validates browser direct upload media types and target directory.
func isSupportedPresignUpload(mimeType string, objectPrefix string) bool {
	normalized := strings.ToLower(strings.TrimSpace(mimeType))
	if strings.HasPrefix(normalized, "image/") {
		return true
	}
	if normalizeMediaObjectPrefix(objectPrefix) == chatMediaObjectPrefix {
		return isSupportedChatMime(normalized)
	}
	return strings.HasPrefix(normalized, "video/") && normalizeMediaObjectPrefix(objectPrefix) == advertisementVideoObjectPrefix
}

// 10. CompleteUpload persists media asset metadata after upload.
func (s *UploadService) CompleteUpload(ctx context.Context, params CompleteUploadParams) (*CompleteUploadResult, error) {
	if strings.TrimSpace(params.ObjectKey) == "" || strings.TrimSpace(params.MimeType) == "" || params.FileSize <= 0 {
		return nil, errcode.New(errcode.CodeValidationError, "invalid upload complete payload")
	}
	if !isSupportedCompleteUpload(params.MimeType, params.ObjectKey) {
		return nil, errcode.New(errcode.CodeValidationError, "only image or video uploads are supported")
	}
	if !strings.HasPrefix(strings.TrimSpace(params.ObjectKey), mediaObjectPrefix) {
		return nil, errcode.New(errcode.CodeValidationError, "object_key is invalid")
	}
	if err := s.verifyUploadToken(params); err != nil {
		return nil, err
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
		Where("bucket_name = ? AND object_key = ?", s.runtime.Config.StorageBucket, params.ObjectKey).
		First(&existing).Error
	if err == nil {
		if existing.CreatedBy == nil || *existing.CreatedBy != params.UserID {
			return nil, errcode.New(errcode.CodeAuthForbidden, "media object belongs to another account")
		}
		return s.buildMediaAssetResult(ctx, &existing)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load media asset")
	}

	media := model.MediaAsset{
		PublicID:         utils.NewPublicID(),
		StorageProvider:  s.runtime.Config.StorageProvider,
		BucketName:       s.runtime.Config.StorageBucket,
		ObjectKey:        params.ObjectKey,
		MimeType:         params.MimeType,
		Width:            params.Width,
		Height:           params.Height,
		FileSize:         params.FileSize,
		ChecksumSHA256:   params.ChecksumSHA256,
		CreatedBy:        &params.UserID,
		ProcessingStatus: mediaProcessingReady,
		ScanStatus:       mediaScanPending,
	}
	if strings.HasPrefix(strings.ToLower(params.MimeType), "video/") && strings.HasPrefix(params.ObjectKey, chatMediaObjectPrefix) {
		media.ProcessingStatus = mediaProcessingProcessing
	}

	if err := s.runtime.DB.WithContext(ctx).Create(&media).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to save media asset")
	}

	return s.buildMediaAssetResult(ctx, &media)
}

// 11. AcceptUploadCallback validates and acknowledges an OSS upload callback.
func (s *UploadService) AcceptUploadCallback(_ context.Context, params UploadCallbackParams) (*UploadCallbackResult, error) {
	if !s.runtime.Config.OSSCallbackEnabled {
		return nil, errcode.New(errcode.CodeAuthForbidden, "upload callback is disabled")
	}
	if strings.TrimSpace(s.runtime.Config.OSSCallbackSecret) == "" {
		return nil, errcode.New(errcode.CodeAuthForbidden, "upload callback secret is required")
	}
	expectedToken := ossCallbackToken(s.runtime.Config.OSSCallbackSecret, params.BucketName, params.ObjectKey)
	if !hmac.Equal([]byte(strings.TrimSpace(params.CallbackToken)), []byte(expectedToken)) {
		return nil, errcode.New(errcode.CodeAuthForbidden, "upload callback token is invalid")
	}
	if strings.TrimSpace(params.BucketName) != strings.TrimSpace(s.runtime.Config.StorageBucket) {
		return nil, errcode.New(errcode.CodeValidationError, "bucket_name is invalid")
	}
	if strings.TrimSpace(params.ObjectKey) == "" || strings.TrimSpace(params.MimeType) == "" || params.FileSize <= 0 {
		return nil, errcode.New(errcode.CodeValidationError, "invalid upload callback payload")
	}
	if !isSupportedCompleteUpload(params.MimeType, params.ObjectKey) {
		return nil, errcode.New(errcode.CodeValidationError, "only image or video callbacks are supported")
	}
	if !strings.HasPrefix(strings.TrimSpace(params.ObjectKey), mediaObjectPrefix) {
		return nil, errcode.New(errcode.CodeValidationError, "object_key is invalid")
	}

	return &UploadCallbackResult{
		BucketName: strings.TrimSpace(params.BucketName),
		ObjectKey:  strings.TrimSpace(params.ObjectKey),
		Accepted:   true,
	}, nil
}

// 12. isSupportedCompleteUpload validates completed upload media types and object key.
func isSupportedCompleteUpload(mimeType string, objectKey string) bool {
	normalized := strings.ToLower(strings.TrimSpace(mimeType))
	if strings.HasPrefix(normalized, "image/") {
		return true
	}
	if strings.HasPrefix(strings.TrimSpace(objectKey), chatMediaObjectPrefix) {
		return isSupportedChatMime(normalized)
	}
	return strings.HasPrefix(normalized, "video/") && strings.HasPrefix(strings.TrimSpace(objectKey), advertisementVideoObjectPrefix)
}

// 12.1 isSupportedChatMime limits chat attachments to browser-friendly media and documents.
func isSupportedChatMime(mimeType string) bool {
	return strings.HasPrefix(mimeType, "video/") || mimeType == "application/pdf" || mimeType == "text/plain"
}

// 13. signUploadToken signs the expected upload completion values.
func (s *UploadService) signUploadToken(claims uploadTokenClaims) (string, error) {
	raw, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	payload := base64.RawURLEncoding.EncodeToString(raw)
	signature := s.signUploadTokenPayload(payload)
	return payload + "." + signature, nil
}

// 14. verifyUploadToken verifies ownership and object metadata before completion.
func (s *UploadService) verifyUploadToken(params CompleteUploadParams) error {
	parts := strings.Split(strings.TrimSpace(params.UploadToken), ".")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return errcode.New(errcode.CodeValidationError, "upload_token is invalid")
	}
	if !hmac.Equal([]byte(parts[1]), []byte(s.signUploadTokenPayload(parts[0]))) {
		return errcode.New(errcode.CodeValidationError, "upload_token is invalid")
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return errcode.New(errcode.CodeValidationError, "upload_token is invalid")
	}

	var claims uploadTokenClaims
	if err := json.Unmarshal(raw, &claims); err != nil {
		return errcode.New(errcode.CodeValidationError, "upload_token is invalid")
	}
	if claims.ExpiresAt <= s.runtime.Now().Unix() {
		return errcode.New(errcode.CodeExpired, "upload_token is expired")
	}
	if claims.UserID != params.UserID ||
		strings.TrimSpace(claims.ObjectKey) != strings.TrimSpace(params.ObjectKey) ||
		!strings.EqualFold(strings.TrimSpace(claims.MimeType), strings.TrimSpace(params.MimeType)) ||
		claims.FileSize != params.FileSize {
		return errcode.New(errcode.CodeValidationError, "upload_token does not match upload payload")
	}

	return nil
}

// 15. signUploadTokenPayload signs one encoded upload token payload.
func (s *UploadService) signUploadTokenPayload(payload string) string {
	mac := hmac.New(sha256.New, []byte(s.uploadTokenSecret()))
	mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

// 16. uploadTokenSecret returns stable key material for upload tokens.
func (s *UploadService) uploadTokenSecret() string {
	for _, value := range []string{s.runtime.Config.JWTSecret, s.runtime.Config.EncryptionKey, strconv.FormatInt(s.runtime.Config.SystemUserID, 10)} {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return "ajo-living-upload-token"
}

// 17. uploadTokenTTL returns the presign token lifetime.
func (s *UploadService) uploadTokenTTL() time.Duration {
	ttl := s.runtime.Config.StoragePresignExpires
	if ttl <= 0 {
		return 15 * time.Minute
	}

	return ttl
}

// 18. ossCallbackToken signs the expected OSS callback object identity.
func ossCallbackToken(secret string, bucketName string, objectKey string) string {
	mac := hmac.New(sha256.New, []byte(strings.TrimSpace(secret)))
	mac.Write([]byte(strings.TrimSpace(bucketName)))
	mac.Write([]byte("\n"))
	mac.Write([]byte(strings.TrimSpace(objectKey)))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

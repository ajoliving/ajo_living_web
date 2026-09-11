/*
 * Multipart upload business logic.
 * 1. Create and authorize OSS multipart upload sessions.
 * 2. Sign individual part URLs without proxying binary data.
 * 3. Complete or abort uploads and persist media metadata.
 */
package service

import (
	"context"
	"crypto/hmac"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math"
	"path"
	"sort"
	"strings"
	"time"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/model"
	"ajoliving_web/http_service/internal/utils"
	"gorm.io/gorm"
)

const defaultMultipartPartSize int64 = 10 * 1024 * 1024

// 1. MultipartInitiateParams defines a resumable upload request.
type MultipartInitiateParams struct {
	UserID       int64
	FileName     string
	MimeType     string
	FileSize     int64
	ObjectPrefix string
}

// 2. MultipartInitiateResult contains the upload session metadata.
type MultipartInitiateResult struct {
	UploadID    string    `json:"upload_id"`
	ObjectKey   string    `json:"object_key"`
	UploadToken string    `json:"upload_token"`
	PartSize    int64     `json:"part_size"`
	PartCount   int       `json:"part_count"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// 3. MultipartPartParams identifies one part URL request.
type MultipartPartParams struct {
	UserID      int64
	ObjectKey   string
	UploadID    string
	UploadToken string
	MimeType    string
	FileSize    int64
	PartNumber  int32
}

// 4. MultipartPartResult contains a signed part URL.
type MultipartPartResult struct {
	UploadURL  string            `json:"upload_url"`
	ObjectKey  string            `json:"object_key"`
	UploadID   string            `json:"upload_id"`
	PartNumber int32             `json:"part_number"`
	Headers    map[string]string `json:"headers"`
}

// 5. MultipartCompleteParams defines the completion request.
type MultipartCompleteParams struct {
	UserID         int64
	ObjectKey      string
	UploadID       string
	UploadToken    string
	MimeType       string
	FileSize       int64
	Parts          []MultipartPart
	Width          *int
	Height         *int
	ChecksumSHA256 string
}

// 5.1 MultipartListParams identifies an upload whose parts should be listed.
type MultipartListParams struct {
	UserID      int64
	ObjectKey   string
	UploadID    string
	UploadToken string
	MimeType    string
	FileSize    int64
}

// 6. MultipartAbortParams defines the cancellation request.
type MultipartAbortParams struct {
	UserID      int64
	ObjectKey   string
	UploadID    string
	UploadToken string
	MimeType    string
	FileSize    int64
}

// 7. InitiateMultipart creates a provider upload session and signed token.
func (s *UploadService) InitiateMultipart(ctx context.Context, params MultipartInitiateParams) (*MultipartInitiateResult, error) {
	if strings.TrimSpace(params.FileName) == "" || strings.TrimSpace(params.MimeType) == "" || params.FileSize <= 0 {
		return nil, errcode.New(errcode.CodeValidationError, "invalid multipart upload payload")
	}
	if err := validateChatAttachmentSize(params.FileSize, params.ObjectPrefix); err != nil {
		return nil, err
	}
	if !isSupportedPresignUpload(params.MimeType, params.ObjectPrefix) {
		return nil, errcode.New(errcode.CodeValidationError, "unsupported multipart upload type")
	}
	provider, ok := s.runtime.StorageProvider.(MultipartStorageProvider)
	if !ok {
		return nil, errcode.New(errcode.CodeInternalError, "multipart upload is not supported")
	}
	partSize := s.multipartPartSize()
	partCount := int(math.Ceil(float64(params.FileSize) / float64(partSize)))
	if partCount <= 0 || partCount > 10000 {
		return nil, errcode.New(errcode.CodeValidationError, "file is too large for multipart upload")
	}
	objectKey := normalizeMediaObjectPrefix(params.ObjectPrefix) + strings.ToLower(utils.NewPublicID()) + path.Ext(params.FileName)
	started, err := provider.InitiateMultipart(ctx, MultipartInitiateInput{ObjectKey: objectKey, MimeType: strings.TrimSpace(params.MimeType)})
	if err != nil || started == nil || strings.TrimSpace(started.UploadID) == "" {
		return nil, errcode.New(errcode.CodeInternalError, "failed to initiate multipart upload")
	}
	expiresAt := s.runtime.Now().Add(s.uploadTokenTTL())
	token, err := s.signUploadToken(uploadTokenClaims{UserID: params.UserID, ObjectKey: objectKey, MimeType: strings.TrimSpace(params.MimeType), FileSize: params.FileSize, UploadID: strings.TrimSpace(started.UploadID), ExpiresAt: expiresAt.Unix()})
	if err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to prepare upload token")
	}
	return &MultipartInitiateResult{UploadID: strings.TrimSpace(started.UploadID), ObjectKey: objectKey, UploadToken: token, PartSize: partSize, PartCount: partCount, ExpiresAt: expiresAt}, nil
}

// 8. PresignMultipartPart returns a signed URL for one part.
func (s *UploadService) PresignMultipartPart(ctx context.Context, params MultipartPartParams) (*MultipartPartResult, error) {
	if strings.TrimSpace(params.ObjectKey) == "" || strings.TrimSpace(params.UploadID) == "" || params.PartNumber <= 0 || params.PartNumber > 10000 || params.FileSize <= 0 {
		return nil, errcode.New(errcode.CodeValidationError, "invalid multipart part payload")
	}
	claims, err := s.verifyMultipartToken(params.UserID, params.ObjectKey, params.UploadID, params.MimeType, params.FileSize, params.UploadToken)
	if err != nil {
		return nil, err
	}
	partCount := int(math.Ceil(float64(claims.FileSize) / float64(s.multipartPartSize())))
	if int(params.PartNumber) > partCount {
		return nil, errcode.New(errcode.CodeValidationError, "part_number exceeds expected part count")
	}
	provider, ok := s.runtime.StorageProvider.(MultipartStorageProvider)
	if !ok {
		return nil, errcode.New(errcode.CodeInternalError, "multipart upload is not supported")
	}
	result, err := provider.PresignMultipartPart(ctx, MultipartPartInput{ObjectKey: params.ObjectKey, UploadID: params.UploadID, PartNumber: params.PartNumber}, s.uploadTokenTTL())
	if err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to presign multipart part")
	}
	return &MultipartPartResult{UploadURL: result.UploadURL, ObjectKey: params.ObjectKey, UploadID: params.UploadID, PartNumber: params.PartNumber, Headers: result.Headers}, nil
}

// 9. CompleteMultipart commits parts and records the resulting media asset.
func (s *UploadService) CompleteMultipart(ctx context.Context, params MultipartCompleteParams) (*CompleteUploadResult, error) {
	if strings.TrimSpace(params.ObjectKey) == "" || strings.TrimSpace(params.UploadID) == "" || params.FileSize <= 0 {
		return nil, errcode.New(errcode.CodeValidationError, "invalid multipart completion payload")
	}
	if err := validateMultipartParts(params.Parts); err != nil {
		return nil, err
	}
	if _, err := s.verifyMultipartToken(params.UserID, params.ObjectKey, params.UploadID, params.MimeType, params.FileSize, params.UploadToken); err != nil {
		return nil, err
	}
	var existing model.MediaAsset
	if err := s.runtime.DB.WithContext(ctx).Where("bucket_name = ? AND object_key = ?", s.runtime.Config.StorageBucket, params.ObjectKey).First(&existing).Error; err == nil {
		if existing.CreatedBy == nil || *existing.CreatedBy != params.UserID {
			return nil, errcode.New(errcode.CodeAuthForbidden, "media object belongs to another account")
		}
		return s.buildMediaAssetResult(ctx, &existing)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load media asset")
	}
	provider, ok := s.runtime.StorageProvider.(MultipartStorageProvider)
	if !ok {
		return nil, errcode.New(errcode.CodeInternalError, "multipart upload is not supported")
	}
	parts := append([]MultipartPart(nil), params.Parts...)
	sort.Slice(parts, func(i, j int) bool { return parts[i].PartNumber < parts[j].PartNumber })
	if _, err := provider.CompleteMultipart(ctx, MultipartCompleteInput{ObjectKey: params.ObjectKey, UploadID: params.UploadID, Parts: parts}); err != nil {
		return nil, errcode.New(errcode.CodeValidationError, "failed to complete multipart upload")
	}
	return s.CompleteUpload(ctx, CompleteUploadParams{UserID: params.UserID, ObjectKey: params.ObjectKey, UploadToken: params.UploadToken, MimeType: params.MimeType, FileSize: params.FileSize, Width: params.Width, Height: params.Height, ChecksumSHA256: params.ChecksumSHA256})
}

// 9.1 ListMultipartParts returns uploaded parts so clients can resume safely.
func (s *UploadService) ListMultipartParts(ctx context.Context, params MultipartListParams) ([]MultipartPart, error) {
	if _, err := s.verifyMultipartToken(params.UserID, params.ObjectKey, params.UploadID, params.MimeType, params.FileSize, params.UploadToken); err != nil {
		return nil, err
	}
	provider, ok := s.runtime.StorageProvider.(MultipartStorageProvider)
	if !ok {
		return nil, errcode.New(errcode.CodeInternalError, "multipart upload is not supported")
	}
	parts, err := provider.ListMultipartParts(ctx, MultipartListInput{ObjectKey: params.ObjectKey, UploadID: params.UploadID})
	if err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to list multipart parts")
	}
	sort.Slice(parts, func(i, j int) bool { return parts[i].PartNumber < parts[j].PartNumber })
	return parts, nil
}

// 10. AbortMultipart cancels an in-progress upload after token validation.
func (s *UploadService) AbortMultipart(ctx context.Context, params MultipartAbortParams) error {
	if _, err := s.verifyMultipartToken(params.UserID, params.ObjectKey, params.UploadID, params.MimeType, params.FileSize, params.UploadToken); err != nil {
		return err
	}
	provider, ok := s.runtime.StorageProvider.(MultipartStorageProvider)
	if !ok {
		return errcode.New(errcode.CodeInternalError, "multipart upload is not supported")
	}
	if err := provider.AbortMultipart(ctx, MultipartAbortInput{ObjectKey: params.ObjectKey, UploadID: params.UploadID}); err != nil {
		return errcode.New(errcode.CodeInternalError, "failed to abort multipart upload")
	}
	return nil
}

// 11. verifyMultipartToken validates ownership and the upload session ID.
func (s *UploadService) verifyMultipartToken(userID int64, objectKey string, uploadID string, mimeType string, fileSize int64, token string) (uploadTokenClaims, error) {
	parts := strings.Split(strings.TrimSpace(token), ".")
	if len(parts) != 2 || !hmac.Equal([]byte(parts[1]), []byte(s.signUploadTokenPayload(parts[0]))) {
		return uploadTokenClaims{}, errcode.New(errcode.CodeValidationError, "upload_token is invalid")
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return uploadTokenClaims{}, errcode.New(errcode.CodeValidationError, "upload_token is invalid")
	}
	var claims uploadTokenClaims
	if err := json.Unmarshal(raw, &claims); err != nil || claims.ExpiresAt <= s.runtime.Now().Unix() {
		if err != nil {
			return uploadTokenClaims{}, errcode.New(errcode.CodeValidationError, "upload_token is invalid")
		}
		return uploadTokenClaims{}, errcode.New(errcode.CodeExpired, "upload_token is expired")
	}
	if claims.UploadID == "" || claims.UserID != userID || claims.UploadID != strings.TrimSpace(uploadID) || claims.ObjectKey != strings.TrimSpace(objectKey) || !strings.EqualFold(claims.MimeType, strings.TrimSpace(mimeType)) || claims.FileSize != fileSize {
		return uploadTokenClaims{}, errcode.New(errcode.CodeValidationError, "upload_token does not match upload payload")
	}
	return claims, nil
}

// 12. validateMultipartParts rejects malformed or duplicate part metadata.
func validateMultipartParts(parts []MultipartPart) error {
	if len(parts) == 0 || len(parts) > 10000 {
		return errcode.New(errcode.CodeValidationError, "parts are required")
	}
	seen := make(map[int32]struct{}, len(parts))
	for _, part := range parts {
		if part.PartNumber <= 0 || part.PartNumber > 10000 || strings.TrimSpace(part.ETag) == "" {
			return errcode.New(errcode.CodeValidationError, "invalid multipart part")
		}
		if _, exists := seen[part.PartNumber]; exists {
			return errcode.New(errcode.CodeValidationError, "duplicate multipart part")
		}
		seen[part.PartNumber] = struct{}{}
	}
	return nil
}

// 13. multipartPartSize returns the configured part size with a safe fallback.
func (s *UploadService) multipartPartSize() int64 {
	if s.runtime != nil && s.runtime.Config != nil && s.runtime.Config.StorageMultipartPartSize >= 100*1024 {
		return s.runtime.Config.StorageMultipartPartSize
	}
	return defaultMultipartPartSize
}

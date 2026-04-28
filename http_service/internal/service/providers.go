/*
 * External provider adapters.
 * 1. Define OTP and storage provider boundaries.
 * 2. Supply mock implementations for local development and tests.
 * 3. Provide Alibaba Cloud OSS upload presign support.
 */
package service

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"

	"ajoliving_web/http_service/internal/config"
	"ajoliving_web/http_service/internal/utils"
)

// 1. OTPProvider sends verification codes through an external channel.
type OTPProvider interface {
	SendCode(ctx context.Context, phone string, scene string, code string) error
}

// 2. StorageProvider creates upload targets for media assets.
type StorageProvider interface {
	PresignUpload(ctx context.Context, input PresignUploadInput) (*PresignUploadResult, error)
	HeadObject(ctx context.Context, objectKey string) (*StorageObjectInfo, error)
	DeleteObject(ctx context.Context, objectKey string) error
}

// 3. PresignUploadInput defines input for upload presign.
type PresignUploadInput struct {
	FileName     string
	MimeType     string
	FileSize     int64
	ObjectPrefix string
}

// 4. PresignUploadResult defines the upload target payload.
type PresignUploadResult struct {
	UploadURL string            `json:"upload_url"`
	ObjectKey string            `json:"object_key"`
	Headers   map[string]string `json:"headers"`
}

// 5. StorageObjectInfo defines the storage object metadata payload.
type StorageObjectInfo struct {
	ObjectKey     string
	ContentType   string
	ContentLength int64
	ETag          string
}

const mediaObjectPrefix = "ajo_living/"
const accountMediaObjectPrefix = "ajo_living/account/"

var errStorageObjectNotFound = errors.New("storage object not found")

// 6. MockOTPProvider prints OTP details to stdout through loggerless stub behavior.
type MockOTPProvider struct{}

// 7. SendCode accepts mock OTP delivery without external side effects.
func (p *MockOTPProvider) SendCode(_ context.Context, _ string, _ string, _ string) error {
	return nil
}

// 8. MockStorageProvider creates deterministic local-style upload targets.
type MockStorageProvider struct {
	config *config.Config
}

// 9. OSSStorageProvider creates presigned upload URLs through Alibaba Cloud OSS.
type OSSStorageProvider struct {
	config *config.Config
	client *oss.Client
}

// 10. NewOTPProvider returns the configured OTP provider.
func NewOTPProvider(_ *config.Config) OTPProvider {
	return &MockOTPProvider{}
}

// 11. NewStorageProvider returns the configured storage provider.
func NewStorageProvider(cfg *config.Config) (StorageProvider, error) {
	switch strings.ToLower(strings.TrimSpace(cfg.StorageProvider)) {
	case "", "mock":
		return &MockStorageProvider{config: cfg}, nil
	case "oss":
		provider, err := newOSSStorageProvider(cfg)
		if err != nil {
			return nil, err
		}
		return provider, nil
	default:
		return nil, fmt.Errorf("unsupported storage provider %q", cfg.StorageProvider)
	}
}

// 12. PresignUpload returns a fake upload URL and object key.
func (p *MockStorageProvider) PresignUpload(_ context.Context, input PresignUploadInput) (*PresignUploadResult, error) {
	extension := path.Ext(input.FileName)
	objectKey := fmt.Sprintf("%s%s%s", normalizeMediaObjectPrefix(input.ObjectPrefix), strings.ToLower(utils.NewPublicID()), extension)
	uploadURL := strings.TrimRight(p.config.StorageEndpoint, "/") + "/upload/" + url.PathEscape(objectKey)

	return &PresignUploadResult{
		UploadURL: uploadURL,
		ObjectKey: objectKey,
		Headers: map[string]string{
			"Content-Type": input.MimeType,
			"X-Mock-Until": time.Now().Add(15 * time.Minute).UTC().Format(time.RFC3339),
		},
	}, nil
}

// 13. HeadObject returns mock metadata for the target object key.
func (p *MockStorageProvider) HeadObject(_ context.Context, objectKey string) (*StorageObjectInfo, error) {
	if strings.TrimSpace(objectKey) == "" {
		return nil, errStorageObjectNotFound
	}

	return &StorageObjectInfo{
		ObjectKey: objectKey,
	}, nil
}

// 14. DeleteObject accepts mock object deletion without external side effects.
func (p *MockStorageProvider) DeleteObject(_ context.Context, objectKey string) error {
	if strings.TrimSpace(objectKey) == "" {
		return errStorageObjectNotFound
	}

	return nil
}

// 15. newOSSStorageProvider validates config and creates the OSS-backed provider.
func newOSSStorageProvider(cfg *config.Config) (*OSSStorageProvider, error) {
	if err := validateOSSConfig(cfg); err != nil {
		return nil, err
	}

	client := newOSSClient(cfg)
	return &OSSStorageProvider{
		config: cfg,
		client: client,
	}, nil
}

// 16. PresignUpload returns a PUT presign URL for direct browser upload to OSS.
func (p *OSSStorageProvider) PresignUpload(ctx context.Context, input PresignUploadInput) (*PresignUploadResult, error) {
	extension := path.Ext(input.FileName)
	objectKey := fmt.Sprintf("%s%s%s", normalizeMediaObjectPrefix(input.ObjectPrefix), strings.ToLower(utils.NewPublicID()), extension)

	result, err := p.client.Presign(ctx, &oss.PutObjectRequest{
		Bucket:      oss.Ptr(p.config.StorageBucket),
		Key:         oss.Ptr(objectKey),
		ContentType: oss.Ptr(input.MimeType),
	}, oss.PresignExpires(p.config.StoragePresignExpires))
	if err != nil {
		return nil, fmt.Errorf("presign oss upload: %w", err)
	}

	headers := map[string]string{}
	for key, value := range result.SignedHeaders {
		headers[key] = value
	}
	if len(headers) == 0 {
		headers["Content-Type"] = input.MimeType
	}

	return &PresignUploadResult{
		UploadURL: result.URL,
		ObjectKey: objectKey,
		Headers:   headers,
	}, nil
}

// 17. normalizeMediaObjectPrefix keeps upload object keys in approved media directories.
func normalizeMediaObjectPrefix(prefix string) string {
	switch strings.Trim(strings.TrimSpace(prefix), "/") {
	case "ajo_living/account", "account":
		return accountMediaObjectPrefix
	default:
		return mediaObjectPrefix
	}
}

// 18. HeadObject returns object metadata from OSS for server-side verification.
func (p *OSSStorageProvider) HeadObject(ctx context.Context, objectKey string) (*StorageObjectInfo, error) {
	result, err := p.client.HeadObject(ctx, &oss.HeadObjectRequest{
		Bucket: oss.Ptr(p.config.StorageBucket),
		Key:    oss.Ptr(objectKey),
	})
	if err != nil {
		if isOSSObjectNotFoundError(err) {
			return nil, errStorageObjectNotFound
		}
		return nil, fmt.Errorf("head oss object: %w", err)
	}

	return &StorageObjectInfo{
		ObjectKey:     objectKey,
		ContentType:   strings.TrimSpace(oss.ToString(result.ContentType)),
		ContentLength: result.ContentLength,
		ETag:          strings.TrimSpace(oss.ToString(result.ETag)),
	}, nil
}

// 19. DeleteObject removes the target object from OSS storage.
func (p *OSSStorageProvider) DeleteObject(ctx context.Context, objectKey string) error {
	_, err := p.client.DeleteObject(ctx, &oss.DeleteObjectRequest{
		Bucket: oss.Ptr(p.config.StorageBucket),
		Key:    oss.Ptr(objectKey),
	})
	if err != nil && !isOSSObjectNotFoundError(err) {
		return fmt.Errorf("delete oss object: %w", err)
	}

	return nil
}

// 20. validateOSSConfig enforces the minimum configuration required for OSS uploads.
func validateOSSConfig(cfg *config.Config) error {
	switch {
	case strings.TrimSpace(cfg.StorageBucket) == "":
		return fmt.Errorf("storage bucket is required when OSS_PROVIDER=oss")
	case strings.TrimSpace(cfg.StorageRegion) == "":
		return fmt.Errorf("storage region is required when OSS_PROVIDER=oss")
	case strings.TrimSpace(cfg.StorageAccessKeyID) == "":
		return fmt.Errorf("storage access key id is required when OSS_PROVIDER=oss")
	case strings.TrimSpace(cfg.StorageAccessKeySecret) == "":
		return fmt.Errorf("storage access key secret is required when OSS_PROVIDER=oss")
	case cfg.StoragePresignExpires <= 0:
		return fmt.Errorf("storage presign expires must be greater than zero")
	case cfg.StoragePresignExpires > 7*24*time.Hour:
		return fmt.Errorf("storage presign expires must be less than or equal to seven days")
	default:
		return nil
	}
}

// 21. newOSSClient builds the Alibaba Cloud OSS SDK client.
func newOSSClient(cfg *config.Config) *oss.Client {
	ossConfig := oss.LoadDefaultConfig().
		WithRegion(strings.TrimSpace(cfg.StorageRegion)).
		WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			strings.TrimSpace(cfg.StorageAccessKeyID),
			strings.TrimSpace(cfg.StorageAccessKeySecret),
			strings.TrimSpace(cfg.StorageSessionToken),
		))

	if endpoint := strings.TrimSpace(cfg.StorageEndpoint); endpoint != "" {
		ossConfig = ossConfig.WithEndpoint(endpoint)
	}
	if cfg.StorageUseCName {
		ossConfig = ossConfig.WithUseCName(true)
	}
	if cfg.StorageDisableSSL {
		ossConfig = ossConfig.WithDisableSSL(true)
	}

	return oss.NewClient(ossConfig)
}

// 22. isOSSObjectNotFoundError checks whether OSS returned an object-not-found response.
func isOSSObjectNotFoundError(err error) bool {
	var serviceErr *oss.ServiceError
	if !errors.As(err, &serviceErr) {
		return false
	}

	return serviceErr.StatusCode == 404 ||
		strings.EqualFold(serviceErr.Code, "NoSuchKey") ||
		strings.EqualFold(serviceErr.Code, "NotFound")
}

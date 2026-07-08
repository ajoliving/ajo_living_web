/*
 * External provider adapters.
 * 1. Define OTP and storage provider boundaries.
 * 2. Supply mock implementations for local development and tests.
 * 3. Provide Alibaba Cloud OSS upload presign support.
 */
package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
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
	PutObject(ctx context.Context, input PutObjectInput) (*StorageObjectInfo, error)
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
	UploadURL   string            `json:"upload_url"`
	ObjectKey   string            `json:"object_key"`
	UploadToken string            `json:"upload_token"`
	Headers     map[string]string `json:"headers"`
}

// 5. PutObjectInput defines a direct server-side object upload.
type PutObjectInput struct {
	ObjectKey string
	MimeType  string
	Body      []byte
}

// 6. StorageObjectInfo defines the storage object metadata payload.
type StorageObjectInfo struct {
	ObjectKey     string
	ContentType   string
	ContentLength int64
	ETag          string
}

const mediaObjectPrefix = "ajo_living/"
const accountMediaObjectPrefix = "ajo_living/account/"
const listingMediaObjectPrefix = "ajo_living/listings/"
const homeEngMediaObjectPrefix = "ajo_living/eng/home-carousel/"
const loginBagMediaObjectPrefix = "ajo_living/login_bag/"
const advertisementImageObjectPrefix = "ajo_living/advertisements/images/"
const advertisementVideoObjectPrefix = "ajo_living/advertisements/video/"

var errStorageObjectNotFound = errors.New("storage object not found")

// 6. MockOTPProvider prints OTP details to stdout through loggerless stub behavior.
type MockOTPProvider struct{}

// 7. SendCode accepts mock OTP delivery without external side effects.
func (p *MockOTPProvider) SendCode(_ context.Context, _ string, _ string, _ string) error {
	return nil
}

// 8. DisabledOTPProvider fails closed when a real provider is not implemented.
type DisabledOTPProvider struct {
	name string
}

// 9. SendCode rejects OTP delivery for unsupported provider names.
func (p *DisabledOTPProvider) SendCode(context.Context, string, string, string) error {
	return fmt.Errorf("OTP provider %q is not available", p.name)
}

// 10. MockStorageProvider creates deterministic local-style upload targets.
type MockStorageProvider struct {
	config *config.Config
}

// 11. OSSStorageProvider creates presigned upload URLs through Alibaba Cloud OSS.
type OSSStorageProvider struct {
	config *config.Config
	client *oss.Client
}

// 12. NewOTPProvider returns the configured OTP provider.
func NewOTPProvider(cfg *config.Config) OTPProvider {
	provider := strings.ToLower(strings.TrimSpace(cfg.OTPProvider))
	switch provider {
	case "", "mock":
		return &MockOTPProvider{}
	default:
		return &DisabledOTPProvider{name: provider}
	}
}

// 13. NewStorageProvider returns the configured storage provider.
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

// 13. PutObject accepts a mock server-side object upload.
func (p *MockStorageProvider) PutObject(_ context.Context, input PutObjectInput) (*StorageObjectInfo, error) {
	if strings.TrimSpace(input.ObjectKey) == "" {
		return nil, errStorageObjectNotFound
	}

	return &StorageObjectInfo{
		ObjectKey:     input.ObjectKey,
		ContentType:   input.MimeType,
		ContentLength: int64(len(input.Body)),
	}, nil
}

// 14. HeadObject returns mock metadata for the target object key.
func (p *MockStorageProvider) HeadObject(_ context.Context, objectKey string) (*StorageObjectInfo, error) {
	if strings.TrimSpace(objectKey) == "" {
		return nil, errStorageObjectNotFound
	}

	return &StorageObjectInfo{
		ObjectKey: objectKey,
	}, nil
}

// 15. DeleteObject accepts mock object deletion without external side effects.
func (p *MockStorageProvider) DeleteObject(_ context.Context, objectKey string) error {
	if strings.TrimSpace(objectKey) == "" {
		return errStorageObjectNotFound
	}

	return nil
}

// 16. newOSSStorageProvider validates config and creates the OSS-backed provider.
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

// 17. PresignUpload returns a PUT presign URL for direct browser upload to OSS.
func (p *OSSStorageProvider) PresignUpload(ctx context.Context, input PresignUploadInput) (*PresignUploadResult, error) {
	extension := path.Ext(input.FileName)
	objectKey := fmt.Sprintf("%s%s%s", normalizeMediaObjectPrefix(input.ObjectPrefix), strings.ToLower(utils.NewPublicID()), extension)

	request := &oss.PutObjectRequest{
		Bucket:      oss.Ptr(p.config.StorageBucket),
		Key:         oss.Ptr(objectKey),
		ContentType: oss.Ptr(input.MimeType),
	}
	if callback := buildOSSUploadCallback(p.config, input, objectKey); callback != "" {
		request.Callback = oss.Ptr(callback)
	}

	result, err := p.client.Presign(ctx, request, oss.PresignExpires(p.config.StoragePresignExpires))
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

// 18. buildOSSUploadCallback returns the optional OSS upload callback payload.
func buildOSSUploadCallback(cfg *config.Config, input PresignUploadInput, objectKey string) string {
	if cfg == nil || !cfg.OSSCallbackEnabled {
		return ""
	}
	callbackURL := strings.TrimSpace(cfg.OSSCallbackURL)
	callbackSecret := strings.TrimSpace(cfg.OSSCallbackSecret)
	if callbackURL == "" || callbackSecret == "" {
		return ""
	}
	callbackBody := `{"object_key":"${object}","bucket_name":"${bucket}","mime_type":"${mimeType}","file_size":${size},"etag":"${etag}"}`
	token, err := json.Marshal(ossCallbackToken(callbackSecret, cfg.StorageBucket, objectKey))
	if err != nil {
		return ""
	}
	callbackBody = strings.TrimSuffix(callbackBody, "}") + `,"callback_token":` + string(token) + `}`

	payload := map[string]string{
		"callbackUrl":      callbackURL,
		"callbackHost":     callbackHost(callbackURL),
		"callbackBodyType": "application/json",
		"callbackBody":     callbackBody,
	}
	if strings.TrimSpace(input.MimeType) != "" {
		payload["callbackBody"] = strings.ReplaceAll(payload["callbackBody"], "${mimeType}", input.MimeType)
	}

	encoded, err := json.Marshal(payload)
	if err != nil {
		return ""
	}

	return base64.StdEncoding.EncodeToString(encoded)
}

// 19. callbackHost returns the host value expected by OSS callback config.
func callbackHost(callbackURL string) string {
	parsed, err := url.Parse(callbackURL)
	if err != nil || parsed.Host == "" {
		return ""
	}

	return parsed.Host
}

// 20. PutObject uploads bytes directly to OSS from the server.
func (p *OSSStorageProvider) PutObject(ctx context.Context, input PutObjectInput) (*StorageObjectInfo, error) {
	if strings.TrimSpace(input.ObjectKey) == "" || strings.TrimSpace(input.MimeType) == "" || len(input.Body) == 0 {
		return nil, fmt.Errorf("invalid object upload payload")
	}

	_, err := p.client.PutObject(ctx, &oss.PutObjectRequest{
		Bucket:      oss.Ptr(p.config.StorageBucket),
		Key:         oss.Ptr(input.ObjectKey),
		Body:        bytes.NewReader(input.Body),
		ContentType: oss.Ptr(input.MimeType),
	})
	if err != nil {
		return nil, fmt.Errorf("put oss object: %w", err)
	}

	return &StorageObjectInfo{
		ObjectKey:     input.ObjectKey,
		ContentType:   input.MimeType,
		ContentLength: int64(len(input.Body)),
	}, nil
}

// 21. normalizeMediaObjectPrefix keeps upload object keys in approved media directories.
func normalizeMediaObjectPrefix(prefix string) string {
	normalizedPrefix := strings.Trim(strings.TrimSpace(prefix), "/")
	switch normalizedPrefix {
	case "ajo_living/account", "account":
		return accountMediaObjectPrefix
	case "ajo_living/listings", "listings":
		return listingMediaObjectPrefix
	case "ajo_living/eng/home-carousel", "eng/home-carousel":
		return homeEngMediaObjectPrefix
	case "ajo_living/login_bag", "login_bag":
		return loginBagMediaObjectPrefix
	case "ajo_living/advertisements/images", "advertisements/images":
		return advertisementImageObjectPrefix
	case "ajo_living/advertisements/video", "advertisements/video":
		return advertisementVideoObjectPrefix
	default:
		if listingPrefix, ok := normalizeListingMediaObjectPrefix(normalizedPrefix); ok {
			return listingPrefix
		}
		return mediaObjectPrefix
	}
}

// 22. normalizeListingMediaObjectPrefix keeps listing uploads under one listing.
func normalizeListingMediaObjectPrefix(prefix string) (string, bool) {
	cleanedPrefix := strings.TrimPrefix(prefix, "ajo_living/")
	if !strings.HasPrefix(cleanedPrefix, "listings/") {
		return "", false
	}

	segments := strings.Split(strings.Trim(cleanedPrefix, "/"), "/")
	if len(segments) < 2 {
		return "", false
	}
	for _, segment := range segments {
		if !isSafeObjectPrefixSegment(segment) {
			return "", false
		}
	}

	return mediaObjectPrefix + strings.Join(segments, "/") + "/", true
}

// 23. isSafeObjectPrefixSegment validates one object prefix segment.
func isSafeObjectPrefixSegment(segment string) bool {
	if segment == "" {
		return false
	}

	for _, char := range segment {
		if char >= 'a' && char <= 'z' {
			continue
		}
		if char >= 'A' && char <= 'Z' {
			continue
		}
		if char >= '0' && char <= '9' {
			continue
		}
		if char == '-' || char == '_' {
			continue
		}
		return false
	}

	return true
}

// 24. HeadObject returns object metadata from OSS for server-side verification.
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

// 23. DeleteObject removes the target object from OSS storage.
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

// 24. validateOSSConfig enforces the minimum configuration required for OSS uploads.
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

// 25. newOSSClient builds the Alibaba Cloud OSS SDK client.
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

// 26. isOSSObjectNotFoundError checks whether OSS returned an object-not-found response.
func isOSSObjectNotFoundError(err error) bool {
	var serviceErr *oss.ServiceError
	if !errors.As(err, &serviceErr) {
		return false
	}

	return serviceErr.StatusCode == 404 ||
		strings.EqualFold(serviceErr.Code, "NoSuchKey") ||
		strings.EqualFold(serviceErr.Code, "NotFound")
}

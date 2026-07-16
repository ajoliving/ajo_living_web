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
	PresignDownload(ctx context.Context, objectKey string, expires time.Duration) (string, error)
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
const agencyIndividualAvatarObjectPrefix = "ajo_living/agency_individual/avatar/"
const agencyIndividualEAAObjectPrefix = "ajo_living/agency_individual/eaa/"
const agencyIndividualCompanyCardObjectPrefix = "ajo_living/agency_individual/company-card/"
const agencyIndividualWechatQRObjectPrefix = "ajo_living/agency_individual/wechat-qr/"
const agencyCompanyLogoObjectPrefix = "ajo_living/agency_company/logo/"
const agencyCompanyEAAObjectPrefix = "ajo_living/agency_company/eaa/"
const agencyCompanyBusinessRegistrationObjectPrefix = "ajo_living/agency_company/business-registration/"
const agencyCompanyCardObjectPrefix = "ajo_living/agency_company/company-card/"

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
func NewOTPProvider(cfg *config.Config) (OTPProvider, error) {
	provider := strings.ToLower(strings.TrimSpace(cfg.OTPProvider))
	switch provider {
	case "", "mock":
		return &MockOTPProvider{}, nil
	case "aliyun_sms":
		return newAliyunSMSOTPProvider(cfg)
	default:
		return nil, fmt.Errorf("OTP provider %q is not available", provider)
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
	headers := map[string]string{"Content-Type": input.MimeType, "X-Mock-Until": time.Now().Add(15 * time.Minute).UTC().Format(time.RFC3339)}
	if acl := privateAgencyObjectACL(input.ObjectPrefix); acl != "" {
		headers["x-oss-object-acl"] = acl
	}

	return &PresignUploadResult{
		UploadURL: uploadURL,
		ObjectKey: objectKey,
		Headers:   headers,
	}, nil
}

// 12.1 PresignDownload returns a short-lived mock download URL.
func (p *MockStorageProvider) PresignDownload(_ context.Context, objectKey string, expires time.Duration) (string, error) {
	if strings.TrimSpace(objectKey) == "" || expires <= 0 {
		return "", fmt.Errorf("invalid download presign payload")
	}
	baseURL := strings.TrimRight(p.config.StorageEndpoint, "/")
	return baseURL + "/download/" + url.PathEscape(objectKey) + "?expires=" + fmt.Sprintf("%d", expires/time.Second), nil
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
	if isPrivateAgencyEvidenceObjectKey(objectKey) {
		request.Acl = oss.ObjectACLPrivate
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

// 17.1 PresignDownload returns a short-lived signed GET URL for private objects.
func (p *OSSStorageProvider) PresignDownload(ctx context.Context, objectKey string, expires time.Duration) (string, error) {
	if strings.TrimSpace(objectKey) == "" || expires <= 0 {
		return "", fmt.Errorf("invalid download presign payload")
	}
	result, err := p.client.Presign(ctx, &oss.GetObjectRequest{Bucket: oss.Ptr(p.config.StorageBucket), Key: oss.Ptr(strings.TrimSpace(objectKey))}, oss.PresignExpires(expires))
	if err != nil {
		return "", fmt.Errorf("presign oss download: %w", err)
	}
	return result.URL, nil
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
	case "ajo_living/agency_individual/avatar", "agency_individual/avatar":
		return agencyIndividualAvatarObjectPrefix
	case "ajo_living/agency_individual/eaa", "agency_individual/eaa":
		return agencyIndividualEAAObjectPrefix
	case "ajo_living/agency_individual/company-card", "agency_individual/company-card":
		return agencyIndividualCompanyCardObjectPrefix
	case "ajo_living/agency_individual/wechat-qr", "agency_individual/wechat-qr":
		return agencyIndividualWechatQRObjectPrefix
	case "ajo_living/agency_company/logo", "agency_company/logo":
		return agencyCompanyLogoObjectPrefix
	case "ajo_living/agency_company/eaa", "agency_company/eaa":
		return agencyCompanyEAAObjectPrefix
	case "ajo_living/agency_company/business-registration", "agency_company/business-registration":
		return agencyCompanyBusinessRegistrationObjectPrefix
	case "ajo_living/agency_company/company-card", "agency_company/company-card":
		return agencyCompanyCardObjectPrefix
	default:
		if listingPrefix, ok := normalizeListingMediaObjectPrefix(normalizedPrefix); ok {
			return listingPrefix
		}
		return mediaObjectPrefix
	}
}

// 21.1 ResolveAgencyProfileUploadPrefix resolves fixed OSS roots for agency evidence.
func ResolveAgencyProfileUploadPrefix(purpose string) (string, bool) {
	switch strings.TrimSpace(purpose) {
	case "agency_individual_avatar":
		return agencyIndividualAvatarObjectPrefix, true
	case "agency_individual_eaa":
		return agencyIndividualEAAObjectPrefix, true
	case "agency_individual_company_card":
		return agencyIndividualCompanyCardObjectPrefix, true
	case "agency_individual_wechat_qr":
		return agencyIndividualWechatQRObjectPrefix, true
	case "agency_company_logo":
		return agencyCompanyLogoObjectPrefix, true
	case "agency_company_eaa":
		return agencyCompanyEAAObjectPrefix, true
	case "agency_company_business_registration":
		return agencyCompanyBusinessRegistrationObjectPrefix, true
	case "agency_company_company_card":
		return agencyCompanyCardObjectPrefix, true
	default:
		return "", false
	}
}

// 21.2 AgencyProfilePurposeMatchesAccount prevents restricted sessions from crossing profile roots.
func AgencyProfilePurposeMatchesAccount(purpose string, accountType string) bool {
	purpose = strings.TrimSpace(purpose)
	if accountType == AccountTypeIndividualAgent {
		return strings.HasPrefix(purpose, "agency_individual_")
	}
	if accountType == AccountTypeAgencyCompany {
		return strings.HasPrefix(purpose, "agency_company_")
	}
	return false
}

// 21.3 IsAgencyProfileObjectKey reports whether a key belongs to an agency evidence root.
func IsAgencyProfileObjectKey(objectKey string) bool {
	for _, purpose := range []string{"agency_individual_avatar", "agency_individual_eaa", "agency_individual_company_card", "agency_individual_wechat_qr", "agency_company_logo", "agency_company_eaa", "agency_company_business_registration", "agency_company_company_card"} {
		prefix, _ := ResolveAgencyProfileUploadPrefix(purpose)
		if strings.HasPrefix(strings.TrimSpace(objectKey), prefix) {
			return true
		}
	}
	return false
}

// 21.4 isPrivateAgencyEvidenceObjectKey identifies documents that must never be public-read.
func isPrivateAgencyEvidenceObjectKey(objectKey string) bool {
	cleaned := strings.TrimSpace(objectKey)
	return strings.HasPrefix(cleaned, agencyIndividualEAAObjectPrefix) || strings.HasPrefix(cleaned, agencyCompanyEAAObjectPrefix) || strings.HasPrefix(cleaned, agencyCompanyBusinessRegistrationObjectPrefix)
}

// 21.5 privateAgencyObjectACL returns the mock header value for evidence roots.
func privateAgencyObjectACL(objectPrefix string) string {
	if isPrivateAgencyEvidenceObjectKey(normalizeMediaObjectPrefix(objectPrefix)) {
		return "private"
	}
	return ""
}

// 21.6 agencyProfileAssetPrefixes returns semantic roots for one profile type.
func agencyProfileAssetPrefixes(profileType string) map[string]string {
	if profileType == AgencyProfileTypeIndividual {
		return map[string]string{"avatar": agencyIndividualAvatarObjectPrefix, "eaa": agencyIndividualEAAObjectPrefix, "company_card": agencyIndividualCompanyCardObjectPrefix, "wechat_qr": agencyIndividualWechatQRObjectPrefix}
	}
	return map[string]string{"logo": agencyCompanyLogoObjectPrefix, "eaa": agencyCompanyEAAObjectPrefix, "business_registration": agencyCompanyBusinessRegistrationObjectPrefix, "company_card": agencyCompanyCardObjectPrefix}
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

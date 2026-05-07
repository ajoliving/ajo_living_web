/*
 * Provider adapter tests.
 * 1. Verify OSS provider validation fails fast on missing required config.
 * 2. Verify OSS upload presign returns a signed PUT URL and expected headers.
 * 3. Verify media object prefixes stay inside approved directories.
 */
package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"ajoliving_web/http_service/internal/config"
)

// 1. TestNewStorageProviderRejectsInvalidOSSConfig verifies required OSS config validation.
func TestNewStorageProviderRejectsInvalidOSSConfig(t *testing.T) {
	provider, err := NewStorageProvider(&config.Config{
		StorageProvider:        "oss",
		StorageBucket:          "media-bucket",
		StorageAccessKeyID:     "ak",
		StorageAccessKeySecret: "sk",
		StoragePresignExpires:  15 * time.Minute,
	})
	if err == nil {
		t.Fatalf("expected missing storage region to fail")
	}
	if provider != nil {
		t.Fatalf("expected provider to be nil on validation failure")
	}
	if !strings.Contains(err.Error(), "storage region is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// 3. TestNormalizeMediaObjectPrefixAllowsListingDirectories verifies listing scoped uploads.
func TestNormalizeMediaObjectPrefixAllowsListingDirectories(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "account",
			input:    "ajo_living/account/",
			expected: accountMediaObjectPrefix,
		},
		{
			name:     "listing root",
			input:    "ajo_living/listings/",
			expected: listingMediaObjectPrefix,
		},
		{
			name:     "listing detail",
			input:    "ajo_living/listings/01KLISTINGABC/",
			expected: "ajo_living/listings/01KLISTINGABC/",
		},
		{
			name:     "unsafe listing detail",
			input:    "ajo_living/listings/../account/",
			expected: mediaObjectPrefix,
		},
	}

	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			if actual := normalizeMediaObjectPrefix(item.input); actual != item.expected {
				t.Fatalf("expected %q, got %q", item.expected, actual)
			}
		})
	}
}

// 2. TestOSSStorageProviderPresignUpload verifies OSS presign behavior without external IO.
func TestOSSStorageProviderPresignUpload(t *testing.T) {
	provider, err := NewStorageProvider(&config.Config{
		StorageProvider:        "oss",
		StorageBucket:          "media-bucket",
		StorageEndpoint:        "oss-cn-hongkong.aliyuncs.com",
		StorageRegion:          "cn-hongkong",
		StorageAccessKeyID:     "ak",
		StorageAccessKeySecret: "sk",
		StoragePresignExpires:  15 * time.Minute,
	})
	if err != nil {
		t.Fatalf("create oss storage provider: %v", err)
	}

	result, err := provider.PresignUpload(context.Background(), PresignUploadInput{
		FileName: "cover.webp",
		MimeType: "image/webp",
		FileSize: 182736,
	})
	if err != nil {
		t.Fatalf("presign oss upload: %v", err)
	}

	if !strings.HasPrefix(result.ObjectKey, mediaObjectPrefix) {
		t.Fatalf("expected object key to use %s prefix, got %q", mediaObjectPrefix, result.ObjectKey)
	}
	if !strings.HasSuffix(result.ObjectKey, ".webp") {
		t.Fatalf("expected object key to keep extension, got %q", result.ObjectKey)
	}
	if result.Headers["Content-Type"] != "image/webp" {
		t.Fatalf("expected content type header, got %#v", result.Headers)
	}
	if !strings.Contains(result.UploadURL, "https://media-bucket.oss-cn-hongkong.aliyuncs.com/") {
		t.Fatalf("expected OSS host in upload URL, got %q", result.UploadURL)
	}
	if !strings.Contains(result.UploadURL, "x-oss-signature=") {
		t.Fatalf("expected v4 signature in upload URL, got %q", result.UploadURL)
	}
}

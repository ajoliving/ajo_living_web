/*
 * OSS CORS 設定工具。
 * 1. 讀取後端 .env 中的 OSS 與公開站點設定。
 * 2. 寫入允許前端直傳的 bucket CORS 規則。
 * 3. 驗證 OPTIONS 預檢是否會返回允許來源。
 */
package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"

	"ajoliving_web/http_service/internal/config"
)

// 1. main applies OSS CORS settings for browser direct upload.
func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	cfg := config.Load()
	if err := validateConfig(cfg); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	client := newClient(cfg)
	origins := allowedOrigins(cfg)
	maxAgeSeconds := int64(3600)
	responseVary := true

	_, err := client.PutBucketCors(ctx, &oss.PutBucketCorsRequest{
		Bucket: oss.Ptr(cfg.StorageBucket),
		CORSConfiguration: &oss.CORSConfiguration{
			ResponseVary: oss.Ptr(responseVary),
			CORSRules: []oss.CORSRule{
				{
					AllowedOrigins: origins,
					AllowedMethods: []string{"GET", "HEAD", "PUT"},
					AllowedHeaders: []string{"*"},
					ExposeHeaders:  []string{"ETag", "x-oss-request-id"},
					MaxAgeSeconds:  oss.Ptr(maxAgeSeconds),
				},
			},
		},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "put bucket cors: %v\n", err)
		os.Exit(1)
	}

	if err := verifyPreflight(ctx, client, cfg, origins[0]); err != nil {
		fmt.Fprintf(os.Stderr, "verify bucket cors: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("OSS CORS updated for bucket %s\n", cfg.StorageBucket)
	fmt.Printf("Allowed origins: %s\n", strings.Join(origins, ", "))
}

// 2. validateConfig checks the minimum OSS configuration.
func validateConfig(cfg *config.Config) error {
	switch {
	case !strings.EqualFold(strings.TrimSpace(cfg.StorageProvider), "oss"):
		return fmt.Errorf("OSS_PROVIDER must be oss")
	case strings.TrimSpace(cfg.StorageBucket) == "":
		return fmt.Errorf("OSS_BUCKET is required")
	case strings.TrimSpace(cfg.StorageRegion) == "":
		return fmt.Errorf("OSS_REGION is required")
	case strings.TrimSpace(cfg.StorageAccessKeyID) == "":
		return fmt.Errorf("OSS_ACCESS_KEY_ID is required")
	case strings.TrimSpace(cfg.StorageAccessKeySecret) == "":
		return fmt.Errorf("OSS_ACCESS_KEY_SECRET is required")
	default:
		return nil
	}
}

// 3. newClient builds an OSS SDK client from runtime config.
func newClient(cfg *config.Config) *oss.Client {
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

// 4. allowedOrigins returns local and configured frontend origins.
func allowedOrigins(cfg *config.Config) []string {
	candidates := []string{
		cfg.AppPublicBaseURL,
		"http://localhost:5173",
		"http://127.0.0.1:5173",
	}
	if strings.TrimSpace(cfg.OSSAllowedOrigins) != "" {
		candidates = append(candidates, strings.Split(cfg.OSSAllowedOrigins, ",")...)
	}

	seen := map[string]bool{}
	origins := []string{}
	for _, candidate := range candidates {
		origin := strings.TrimRight(strings.TrimSpace(candidate), "/")
		if origin == "" || seen[origin] {
			continue
		}
		seen[origin] = true
		origins = append(origins, origin)
	}

	return origins
}

// 5. verifyPreflight checks whether OSS accepts the browser PUT preflight.
func verifyPreflight(ctx context.Context, client *oss.Client, cfg *config.Config, origin string) error {
	result, err := client.OptionObject(ctx, &oss.OptionObjectRequest{
		Bucket:                      oss.Ptr(cfg.StorageBucket),
		Key:                         oss.Ptr("ajo_living/account/cors-check.txt"),
		Origin:                      oss.Ptr(origin),
		AccessControlRequestMethod:  oss.Ptr("PUT"),
		AccessControlRequestHeaders: oss.Ptr("content-type"),
	})
	if err != nil {
		return err
	}

	if strings.TrimSpace(oss.ToString(result.AccessControlAllowOrigin)) == "" {
		return fmt.Errorf("missing Access-Control-Allow-Origin")
	}

	return nil
}

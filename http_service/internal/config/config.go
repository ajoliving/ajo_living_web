/*
 * Application configuration loader.
 * 1. Load environment variables and optional .env file.
 * 2. Normalize runtime defaults for local development.
 * 3. Expose typed configuration for server, database, auth, and providers.
 */
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// 1. Config defines the runtime configuration set.
type Config struct {
	AppEnv                 string
	AppPort                string
	AppPublicBaseURL       string
	DBDriver               string
	DBDSN                  string
	JWTSecret              string
	EncryptionKey          string
	BootstrapStaffPhones   string
	OTPProvider            string
	OTPMockCode            string
	MailEnabled            bool
	SMTPHost               string
	SMTPPort               int
	SMTPUsername           string
	SMTPPassword           string
	SMTPFrom               string
	SMTPHelloName          string
	StorageProvider        string
	StorageBucket          string
	StorageEndpoint        string
	StorageRegion          string
	StorageAccessKeyID     string
	StorageAccessKeySecret string
	StorageSessionToken    string
	StorageUseCName        bool
	StorageDisableSSL      bool
	StoragePresignExpires  time.Duration
	MediaBaseURL           string
	SeedCommunities        bool
	EnableExpireTicker     bool
	ExpireTickerInterval   time.Duration
}

// 2. Load reads environment variables and returns normalized config.
func Load() *Config {
	_ = godotenv.Load()

	cfg := &Config{
		AppEnv:                 getEnv("APP_ENV", "development"),
		AppPort:                getEnv("APP_PORT", "8080"),
		AppPublicBaseURL:       getEnv("APP_PUBLIC_BASE_URL", "http://localhost:5173"),
		DBDriver:               getEnv("DB_DRIVER", "postgres"),
		DBDSN:                  getEnv("DB_DSN", "host=127.0.0.1 user=postgres password=postgres dbname=ajoliving port=5432 sslmode=disable TimeZone=Asia/Shanghai"),
		JWTSecret:              getEnv("JWT_SECRET", "dev-secret-key"),
		EncryptionKey:          getEnv("ENCRYPTION_KEY", "dev-encryption-key"),
		BootstrapStaffPhones:   getEnv("BOOTSTRAP_STAFF_PHONES", ""),
		OTPProvider:            getEnv("OTP_PROVIDER", "mock"),
		OTPMockCode:            getEnv("OTP_MOCK_CODE", "123456"),
		MailEnabled:            getBoolEnv("MAIL_ENABLED", false),
		SMTPHost:               getEnv("SMTP_HOST", ""),
		SMTPPort:               getIntEnv("SMTP_PORT", 25),
		SMTPUsername:           strings.TrimSpace(os.Getenv("SMTP_USERNAME")),
		SMTPPassword:           os.Getenv("SMTP_PASSWORD"),
		SMTPFrom:               strings.TrimSpace(os.Getenv("SMTP_FROM")),
		SMTPHelloName:          getEnv("SMTP_HELLO_NAME", "localhost"),
		StorageProvider:        getEnvWithLegacy("OSS_PROVIDER", "STORAGE_PROVIDER", "mock"),
		StorageBucket:          getEnvWithLegacy("OSS_BUCKET", "STORAGE_BUCKET", "local-bucket"),
		StorageEndpoint:        getEnvWithLegacy("OSS_ENDPOINT", "STORAGE_ENDPOINT", "http://localhost:9000"),
		StorageRegion:          getEnvWithLegacy("OSS_REGION", "STORAGE_REGION", ""),
		StorageAccessKeyID:     getEnvWithLegacy("OSS_ACCESS_KEY_ID", "STORAGE_ACCESS_KEY_ID", ""),
		StorageAccessKeySecret: getEnvWithLegacy("OSS_ACCESS_KEY_SECRET", "STORAGE_ACCESS_KEY_SECRET", ""),
		StorageSessionToken:    getEnvWithLegacy("OSS_SESSION_TOKEN", "STORAGE_SESSION_TOKEN", ""),
		StorageUseCName:        getBoolEnvWithLegacy("OSS_USE_CNAME", "STORAGE_USE_CNAME", false),
		StorageDisableSSL:      getBoolEnvWithLegacy("OSS_DISABLE_SSL", "STORAGE_DISABLE_SSL", false),
		StoragePresignExpires:  getDurationEnvWithLegacy("OSS_PRESIGN_EXPIRES", "STORAGE_PRESIGN_EXPIRES", 15*time.Minute),
		SeedCommunities:        getBoolEnv("SEED_COMMUNITIES", true),
		EnableExpireTicker:     getBoolEnv("ENABLE_EXPIRE_TICKER", true),
		ExpireTickerInterval:   getDurationEnv("EXPIRE_TICKER_INTERVAL", time.Hour),
	}

	cfg.MediaBaseURL = getEnvWithLegacy("OSS_MEDIA_BASE_URL", "MEDIA_BASE_URL", defaultMediaBaseURL(cfg))

	return cfg
}

// 3. getEnv returns env value or fallback.
func getEnv(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

// 4. getEnvWithLegacy returns new env value, legacy env value, or fallback.
func getEnvWithLegacy(key string, legacyKey string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	if value := os.Getenv(legacyKey); value != "" {
		return value
	}

	return fallback
}

// 5. getIntEnv parses integer env value or fallback.
func getIntEnv(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}

	return value
}

// 6. getBoolEnv parses bool env value or fallback.
func getBoolEnv(key string, fallback bool) bool {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}

	value, err := strconv.ParseBool(raw)
	if err != nil {
		return fallback
	}

	return value
}

// 7. getBoolEnvWithLegacy parses new or legacy bool env value.
func getBoolEnvWithLegacy(key string, legacyKey string, fallback bool) bool {
	if os.Getenv(key) != "" {
		return getBoolEnv(key, fallback)
	}

	return getBoolEnv(legacyKey, fallback)
}

// 8. getDurationEnv parses duration env value or fallback.
func getDurationEnv(key string, fallback time.Duration) time.Duration {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}

	value, err := time.ParseDuration(raw)
	if err != nil {
		return fallback
	}

	return value
}

// 9. getDurationEnvWithLegacy parses new or legacy duration env value.
func getDurationEnvWithLegacy(key string, legacyKey string, fallback time.Duration) time.Duration {
	if os.Getenv(key) != "" {
		return getDurationEnv(key, fallback)
	}

	return getDurationEnv(legacyKey, fallback)
}

// 10. SMTPAddress builds the outbound SMTP server address.
func (c *Config) SMTPAddress() string {
	return fmt.Sprintf("%s:%d", c.SMTPHost, c.SMTPPort)
}

// 11. defaultMediaBaseURL builds the default public media base URL.
func defaultMediaBaseURL(cfg *Config) string {
	if strings.EqualFold(strings.TrimSpace(cfg.StorageProvider), "oss") {
		return defaultOSSMediaBaseURL(cfg)
	}

	return strings.TrimRight(cfg.StorageEndpoint, "/") + "/" + strings.TrimLeft(cfg.StorageBucket, "/")
}

// 12. defaultOSSMediaBaseURL derives the public base URL for OSS-backed media.
func defaultOSSMediaBaseURL(cfg *Config) string {
	bucket := strings.TrimSpace(cfg.StorageBucket)
	if bucket == "" {
		return ""
	}

	endpoint := trimEndpointScheme(cfg.StorageEndpoint)
	if endpoint == "" && strings.TrimSpace(cfg.StorageRegion) != "" {
		endpoint = fmt.Sprintf("oss-%s.aliyuncs.com", strings.TrimSpace(cfg.StorageRegion))
	}
	if endpoint == "" {
		return ""
	}

	scheme := "https"
	if cfg.StorageDisableSSL {
		scheme = "http"
	}

	if cfg.StorageUseCName {
		return scheme + "://" + endpoint
	}

	return fmt.Sprintf("%s://%s.%s", scheme, bucket, endpoint)
}

// 13. trimEndpointScheme removes scheme and surrounding slashes from endpoint values.
func trimEndpointScheme(endpoint string) string {
	value := strings.TrimSpace(endpoint)
	value = strings.TrimPrefix(value, "https://")
	value = strings.TrimPrefix(value, "http://")
	return strings.Trim(value, "/")
}

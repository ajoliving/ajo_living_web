/*
 * Application configuration loader.
 * 1. Load environment variables and optional .env file.
 * 2. Normalize runtime defaults for local development.
 * 3. Expose typed configuration for server, database, auth, and providers.
 */
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// 1. Config defines the runtime configuration set.
type Config struct {
	AppEnv                      string
	AppPort                     string
	AppPublicBaseURL            string
	AppAPIPublicBaseURL         string
	CORSAllowedOrigins          string
	StrictProductionConfig      bool
	DBDriver                    string
	DBDSN                       string
	JWTSecret                   string
	EncryptionKey               string
	BootstrapStaffPhones        string
	POSLoginURL                 string
	POSAPIBaseURL               string
	POSAPIUsername              string
	POSAPIPassword              string
	POSAPILoginType             string
	POSLoginUsernameField       string
	POSLoginPasswordField       string
	POSLoginTimeout             time.Duration
	POSDirectoryCacheTTL        time.Duration
	POSPaymentServiceURL        string
	IsmartExternalAppBaseURL    string
	IsmartExternalAppAPIBaseURL string
	IsmartIntegrationAPIBaseURL string
	IsmartServiceCaseAPIBaseURL string
	IsmartSubaccountAPIBaseURL  string
	IsmartExternalAppTimeout    time.Duration
	RedisEnabled                bool
	RedisAddr                   string
	RedisPassword               string
	RedisDB                     int
	IsmartBuildingCacheTTL      time.Duration
	ICCTVAPIBaseURL             string
	ICCTVAdminUsername          string
	ICCTVAdminPassword          string
	ICCTVRequestTimeout         time.Duration
	ICCTVStreamProxyBaseURL     string
	POSTerminalProxyEnabled     bool
	POSTerminalType             string
	POSTerminalURL              string
	POSTerminalAllowedBuildings string
	GoodPriceAPIBaseURL         string
	GoodPriceRequestTimeout     time.Duration
	GoodPriceCacheTTL           time.Duration
	GoodPriceSummaryCacheTTL    time.Duration
	GoodPriceSearchCacheTTL     time.Duration
	GoodPriceDetailCacheTTL     time.Duration
	GoodPriceAlertEnabled       bool
	GoodPriceAlertInterval      time.Duration
	TranslationProvider         string
	DeepLAPIBaseURL             string
	DeepLAuthKey                string
	TranslationRequestTimeout   time.Duration
	OTPProvider                 string
	OTPMockCode                 string
	OTPResendCooldown           time.Duration
	AliyunSMSAccessKeyID        string
	AliyunSMSAccessKeySecret    string
	AliyunSMSRegionID           string
	AliyunSMSSignName           string
	AliyunSMSTemplateCode       string
	AliyunSMSTemplateParamName  string
	AliyunSMSRequestTimeout     time.Duration
	MailEnabled                 bool
	SMTPHost                    string
	SMTPPort                    int
	SMTPUsername                string
	SMTPPassword                string
	SMTPFrom                    string
	SMTPHelloName               string
	StorageProvider             string
	StorageBucket               string
	StorageEndpoint             string
	StorageRegion               string
	StorageAccessKeyID          string
	StorageAccessKeySecret      string
	StorageSessionToken         string
	StorageUseCName             bool
	StorageDisableSSL           bool
	StoragePresignExpires       time.Duration
	StorageMultipartPartSize    int64
	OSSCallbackEnabled          bool
	OSSCallbackURL              string
	OSSCallbackSecret           string
	OSSAllowedOrigins           string
	MediaBaseURL                string
	MediaProcessingProvider     string
	MediaProcessingTimeout      time.Duration
	MediaFFmpegBinary           string
	MediaFFmpegArgs             string
	MediaScanProvider           string
	MediaScanTimeout            time.Duration
	MediaClamAVBinary           string
	MediaClamAVArgs             string
	SeedCommunities             bool
	SeedHomeContent             bool
	WebPublicDir                string
	SystemUserID                int64
	EnableExpireTicker          bool
	ExpireTickerInterval        time.Duration
	WalletRechargeEnabled       bool
	PaymentGatewayEnv           string
	PaymentGatewayBaseURL       string
	PaymentMerchantNo           string
	PaymentReturnBaseURL        string
	PaymentNotifyURL            string
	PaymentRequestTimeout       time.Duration
	PaymentExpireSeconds        int
	PaymentChannelConfigs       map[string]PaymentChannelConfig
}

// 2. PaymentChannelConfig stores EasyLink app credentials for one channel.
type PaymentChannelConfig struct {
	AppID        string
	AppSecret    string
	ChannelExtra string
}

// 3. Load reads environment variables and returns normalized config.
func Load() *Config {
	_ = godotenv.Load()
	easylinkEnv := getEnv("EASYLINK_ENV", "sandbox")
	sharedH5AppID := getEnv("EASYLINK_H5_APP_ID", "")
	sharedH5AppSecret := getEnv("EASYLINK_H5_APP_SECRET", "")
	posAPIBaseURL := getEnv("POS_API_BASE_URL", defaultPOSAPIBaseURL())

	cfg := &Config{
		AppEnv:                      getEnv("APP_ENV", "development"),
		AppPort:                     getEnv("APP_PORT", "8081"),
		AppPublicBaseURL:            getEnv("APP_PUBLIC_BASE_URL", "http://localhost:5173"),
		AppAPIPublicBaseURL:         getEnv("APP_API_PUBLIC_BASE_URL", "http://localhost:8081"),
		CORSAllowedOrigins:          getEnv("CORS_ALLOWED_ORIGINS", ""),
		StrictProductionConfig:      getBoolEnv("STRICT_PRODUCTION_CONFIG", false),
		DBDriver:                    getEnv("DB_DRIVER", "postgres"),
		DBDSN:                       getEnv("DB_DSN", "host=127.0.0.1 user=postgres password=postgres dbname=ajoliving port=5432 sslmode=disable TimeZone=Asia/Shanghai"),
		JWTSecret:                   getEnv("JWT_SECRET", "dev-secret-key"),
		EncryptionKey:               getEnv("ENCRYPTION_KEY", "dev-encryption-key"),
		BootstrapStaffPhones:        getEnv("BOOTSTRAP_STAFF_PHONES", ""),
		POSLoginURL:                 getEnv("POS_LOGIN_URL", defaultPOSLoginURL(posAPIBaseURL)),
		POSAPIBaseURL:               posAPIBaseURL,
		POSAPIUsername:              getEnv("POS_API_USERNAME", "testowner02"),
		POSAPIPassword:              getEnv("POS_API_PASSWORD", "test02test02"),
		POSAPILoginType:             getEnv("POS_API_LOGIN_TYPE", "username"),
		POSLoginUsernameField:       getEnv("POS_LOGIN_USERNAME_FIELD", "login_name"),
		POSLoginPasswordField:       getEnv("POS_LOGIN_PASSWORD_FIELD", "password"),
		POSLoginTimeout:             getDurationEnv("POS_LOGIN_TIMEOUT", 10*time.Second),
		POSDirectoryCacheTTL:        getDurationEnv("POS_DIRECTORY_CACHE_TTL", 5*time.Minute),
		POSPaymentServiceURL:        getEnv("POS_PAYMENT_SERVICE_URL", defaultPOSPaymentServiceURL()),
		IsmartExternalAppBaseURL:    getEnv("ISMART_EXTERNAL_APP_BASE_URL", defaultIsmartExternalAppBaseURL()),
		IsmartExternalAppAPIBaseURL: getEnv("ISMART_EXTERNAL_APP_API_BASE_URL", defaultIsmartExternalAppAPIBaseURL()),
		IsmartIntegrationAPIBaseURL: getEnv("ISMART_INTEGRATION_API_BASE_URL", defaultIsmartIntegrationAPIBaseURL()),
		IsmartServiceCaseAPIBaseURL: getEnv("ISMART_SERVICE_CASE_API_BASE_URL", defaultIsmartServiceCaseAPIBaseURL()),
		IsmartSubaccountAPIBaseURL:  getEnv("ISMART_SUBACCOUNT_API_BASE_URL", defaultIsmartSubaccountAPIBaseURL()),
		IsmartExternalAppTimeout:    getDurationEnv("ISMART_EXTERNAL_APP_TIMEOUT", 10*time.Second),
		RedisEnabled:                getBoolEnv("REDIS_ENABLED", true),
		RedisAddr:                   strings.TrimSpace(getEnv("REDIS_ADDR", "127.0.0.1:6382")),
		RedisPassword:               os.Getenv("REDIS_PASSWORD"),
		RedisDB:                     getIntEnv("REDIS_DB", 0),
		IsmartBuildingCacheTTL:      getDurationEnv("ISMART_BUILDING_CACHE_TTL", 5*time.Minute),
		ICCTVAPIBaseURL:             getEnv("ICCTV_API_BASE_URL", defaultICCTVAPIBaseURL()),
		ICCTVAdminUsername:          strings.TrimSpace(getEnv("ICCTV_ADMIN_USERNAME", "")),
		ICCTVAdminPassword:          os.Getenv("ICCTV_ADMIN_PASSWORD"),
		ICCTVRequestTimeout:         getDurationEnv("ICCTV_REQUEST_TIMEOUT", 10*time.Second),
		ICCTVStreamProxyBaseURL:     strings.TrimRight(strings.TrimSpace(getEnv("ICCTV_STREAM_PROXY_BASE_URL", "")), "/"),
		POSTerminalProxyEnabled:     getBoolEnv("POS_TERMINAL_PROXY_ENABLED", false),
		GoodPriceAPIBaseURL:         getEnv("GOOD_PRICE_API_BASE_URL", "https://good.price.skylinedances.com/api"),
		GoodPriceRequestTimeout:     getDurationEnv("GOOD_PRICE_REQUEST_TIMEOUT", 10*time.Second),
		GoodPriceCacheTTL:           getDurationEnv("GOOD_PRICE_CACHE_TTL", 5*time.Minute),
		GoodPriceSummaryCacheTTL:    getDurationEnv("GOOD_PRICE_SUMMARY_CACHE_TTL", 5*time.Minute),
		GoodPriceSearchCacheTTL:     getDurationEnv("GOOD_PRICE_SEARCH_CACHE_TTL", 2*time.Minute),
		GoodPriceDetailCacheTTL:     getDurationEnv("GOOD_PRICE_DETAIL_CACHE_TTL", 3*time.Minute),
		GoodPriceAlertEnabled:       getBoolEnv("GOOD_PRICE_ALERT_ENABLED", true),
		GoodPriceAlertInterval:      getDurationEnv("GOOD_PRICE_ALERT_INTERVAL", time.Hour),
		TranslationProvider:         getEnv("TRANSLATION_PROVIDER", "disabled"),
		DeepLAPIBaseURL:             strings.TrimRight(strings.TrimSpace(getEnv("DEEPL_API_BASE_URL", "https://api-free.deepl.com")), "/"),
		DeepLAuthKey:                strings.TrimSpace(os.Getenv("DEEPL_AUTH_KEY")),
		TranslationRequestTimeout:   getDurationEnv("TRANSLATION_REQUEST_TIMEOUT", 10*time.Second),
		POSTerminalType:             getEnv("POS_TERMINAL_TYPE", "allinpay"),
		POSTerminalURL:              getEnv("POS_TERMINAL_URL", ""),
		POSTerminalAllowedBuildings: getEnv("POS_TERMINAL_ALLOWED_BUILDINGS", ""),
		OTPProvider:                 getEnv("OTP_PROVIDER", "mock"),
		OTPMockCode:                 getEnv("OTP_MOCK_CODE", "123456"),
		OTPResendCooldown:           getDurationEnv("OTP_RESEND_COOLDOWN", time.Minute),
		AliyunSMSAccessKeyID:        strings.TrimSpace(os.Getenv("ALIYUN_SMS_ACCESS_KEY_ID")),
		AliyunSMSAccessKeySecret:    os.Getenv("ALIYUN_SMS_ACCESS_KEY_SECRET"),
		AliyunSMSRegionID:           getEnv("ALIYUN_SMS_REGION_ID", "cn-hangzhou"),
		AliyunSMSSignName:           strings.TrimSpace(getEnv("ALIYUN_SMS_SIGN_NAME", "")),
		AliyunSMSTemplateCode:       strings.TrimSpace(getEnv("ALIYUN_SMS_TEMPLATE_CODE_CN", "")),
		AliyunSMSTemplateParamName:  getEnv("ALIYUN_SMS_TEMPLATE_PARAM_NAME", "code"),
		AliyunSMSRequestTimeout:     getDurationEnv("ALIYUN_SMS_REQUEST_TIMEOUT", 10*time.Second),
		MailEnabled:                 getBoolEnv("MAIL_ENABLED", false),
		SMTPHost:                    getEnv("SMTP_HOST", ""),
		SMTPPort:                    getIntEnv("SMTP_PORT", 25),
		SMTPUsername:                strings.TrimSpace(os.Getenv("SMTP_USERNAME")),
		SMTPPassword:                os.Getenv("SMTP_PASSWORD"),
		SMTPFrom:                    strings.TrimSpace(os.Getenv("SMTP_FROM")),
		SMTPHelloName:               getEnv("SMTP_HELLO_NAME", "localhost"),
		StorageProvider:             getEnvWithLegacy("OSS_PROVIDER", "STORAGE_PROVIDER", "mock"),
		StorageBucket:               getEnvWithLegacy("OSS_BUCKET", "STORAGE_BUCKET", "local-bucket"),
		StorageEndpoint:             getEnvWithLegacy("OSS_ENDPOINT", "STORAGE_ENDPOINT", "http://localhost:9000"),
		StorageRegion:               getEnvWithLegacy("OSS_REGION", "STORAGE_REGION", ""),
		StorageAccessKeyID:          getEnvWithLegacy("OSS_ACCESS_KEY_ID", "STORAGE_ACCESS_KEY_ID", ""),
		StorageAccessKeySecret:      getEnvWithLegacy("OSS_ACCESS_KEY_SECRET", "STORAGE_ACCESS_KEY_SECRET", ""),
		StorageSessionToken:         getEnvWithLegacy("OSS_SESSION_TOKEN", "STORAGE_SESSION_TOKEN", ""),
		StorageUseCName:             getBoolEnvWithLegacy("OSS_USE_CNAME", "STORAGE_USE_CNAME", false),
		StorageDisableSSL:           getBoolEnvWithLegacy("OSS_DISABLE_SSL", "STORAGE_DISABLE_SSL", false),
		StoragePresignExpires:       getDurationEnvWithLegacy("OSS_PRESIGN_EXPIRES", "STORAGE_PRESIGN_EXPIRES", 15*time.Minute),
		StorageMultipartPartSize:    getInt64EnvWithLegacy("OSS_MULTIPART_PART_SIZE", "STORAGE_MULTIPART_PART_SIZE", 10*1024*1024),
		OSSCallbackEnabled:          getBoolEnv("OSS_CALLBACK_ENABLED", false),
		OSSCallbackSecret:           strings.TrimSpace(os.Getenv("OSS_CALLBACK_SECRET")),
		OSSAllowedOrigins:           getEnv("OSS_ALLOWED_ORIGINS", ""),
		SeedCommunities:             getBoolEnv("SEED_COMMUNITIES", true),
		SeedHomeContent:             getBoolEnv("SEED_HOME_CONTENT", true),
		WebPublicDir:                getEnv("WEB_PUBLIC_DIR", "../web/public"),
		SystemUserID:                getInt64Env("SYSTEM_USER_ID", 1),
		EnableExpireTicker:          getBoolEnv("ENABLE_EXPIRE_TICKER", true),
		ExpireTickerInterval:        getDurationEnv("EXPIRE_TICKER_INTERVAL", time.Hour),
		WalletRechargeEnabled:       getBoolEnv("WALLET_RECHARGE_ENABLED", true),
		PaymentGatewayEnv:           normalizePaymentGatewayEnv(easylinkEnv),
		PaymentGatewayBaseURL:       getEnv("EASYLINK_BASE_URL", defaultPaymentGatewayBaseURL(easylinkEnv)),
		PaymentMerchantNo:           getEnv("EASYLINK_MCH_NO", ""),
		PaymentReturnBaseURL:        getEnv("WALLET_RECHARGE_RETURN_BASE_URL", getEnv("AJO_EASYLINK_RETURN_BASE_URL", getEnv("EASYLINK_RETURN_BASE_URL", defaultWalletRechargeReturnBaseURL(easylinkEnv)))),
		PaymentNotifyURL:            getEnv("WALLET_RECHARGE_NOTIFY_URL", getEnv("AJO_EASYLINK_NOTIFY_URL", defaultWalletRechargeNotifyURL(easylinkEnv))),
		PaymentRequestTimeout:       time.Duration(getIntEnv("EASYLINK_TIMEOUT_MS", 15000)) * time.Millisecond,
		PaymentExpireSeconds:        getIntEnv("EASYLINK_EXPIRED_SECONDS", 180),
		PaymentChannelConfigs: map[string]PaymentChannelConfig{
			"WX_H5": {
				AppID:        strings.TrimSpace(firstConfigValue(os.Getenv("EASYLINK_WX_H5_APP_ID"), sharedH5AppID)),
				AppSecret:    strings.TrimSpace(firstConfigValue(os.Getenv("EASYLINK_WX_H5_APP_SECRET"), sharedH5AppSecret)),
				ChannelExtra: buildPaymentChannelExtra("appid", os.Getenv("EASYLINK_WX_H5_CHANNEL_APP_ID")),
			},
			"ALI_H5": {
				AppID:        strings.TrimSpace(firstConfigValue(os.Getenv("EASYLINK_ALI_H5_APP_ID"), sharedH5AppID)),
				AppSecret:    strings.TrimSpace(firstConfigValue(os.Getenv("EASYLINK_ALI_H5_APP_SECRET"), sharedH5AppSecret)),
				ChannelExtra: buildPaymentChannelExtra("walletType", os.Getenv("EASYLINK_ALI_H5_WALLET_TYPE")),
			},
			"WX_QR": {
				AppID: strings.TrimSpace(firstConfigValue(
					os.Getenv("EASYLINK_WX_QR_APP_ID"),
					os.Getenv("EASYLINK_WX_H5_APP_ID"),
					sharedH5AppID,
				)),
				AppSecret: strings.TrimSpace(firstConfigValue(
					os.Getenv("EASYLINK_WX_QR_APP_SECRET"),
					os.Getenv("EASYLINK_WX_H5_APP_SECRET"),
					sharedH5AppSecret,
				)),
				ChannelExtra: buildPaymentChannelExtra("payDataType", firstConfigValue(os.Getenv("EASYLINK_WX_QR_PAY_DATA_TYPE"), "codeUrl")),
			},
			"ALI_QR": {
				AppID: strings.TrimSpace(firstConfigValue(
					os.Getenv("EASYLINK_ALI_QR_APP_ID"),
					os.Getenv("EASYLINK_ALI_H5_APP_ID"),
					sharedH5AppID,
				)),
				AppSecret: strings.TrimSpace(firstConfigValue(
					os.Getenv("EASYLINK_ALI_QR_APP_SECRET"),
					os.Getenv("EASYLINK_ALI_H5_APP_SECRET"),
					sharedH5AppSecret,
				)),
				ChannelExtra: buildPaymentChannelExtra("payDataType", firstConfigValue(os.Getenv("EASYLINK_ALI_QR_PAY_DATA_TYPE"), "codeUrl")),
			},
			"YSF_QR": {
				AppID: strings.TrimSpace(firstConfigValue(
					os.Getenv("EASYLINK_YSF_QR_APP_ID"),
					os.Getenv("EASYLINK_UP_OP_APP_ID"),
					os.Getenv("EASYLINK_UP_APP_ID"),
				)),
				AppSecret: strings.TrimSpace(firstConfigValue(
					os.Getenv("EASYLINK_YSF_QR_APP_SECRET"),
					os.Getenv("EASYLINK_UP_OP_APP_SECRET"),
					os.Getenv("EASYLINK_UP_APP_SECRET"),
				)),
				ChannelExtra: buildPaymentChannelExtra("payDataType", firstConfigValue(
					os.Getenv("EASYLINK_YSF_QR_PAY_DATA_TYPE"),
					os.Getenv("EASYLINK_UP_OP_PAY_DATA_TYPE"),
					"codeImgUrl",
				)),
			},
		},
	}

	cfg.MediaBaseURL = getEnvWithLegacy("OSS_MEDIA_BASE_URL", "MEDIA_BASE_URL", defaultMediaBaseURL(cfg))
	cfg.OSSCallbackURL = getEnv("OSS_CALLBACK_URL", defaultOSSCallbackURL(cfg))
	cfg.MediaProcessingProvider = getEnv("MEDIA_PROCESSING_PROVIDER", defaultMediaProvider(cfg.AppEnv))
	cfg.MediaProcessingTimeout = getDurationEnv("MEDIA_PROCESSING_TIMEOUT", 10*time.Minute)
	cfg.MediaFFmpegBinary = getEnv("MEDIA_FFMPEG_BINARY", "ffmpeg")
	cfg.MediaFFmpegArgs = getEnv("MEDIA_FFMPEG_ARGS", "-y -i {source_path} -ss 00:00:01 -frames:v 1 {thumbnail_path} -c:v libx264 -movflags +faststart {playback_path}")
	cfg.MediaScanProvider = getEnv("MEDIA_SCAN_PROVIDER", defaultMediaProvider(cfg.AppEnv))
	cfg.MediaScanTimeout = getDurationEnv("MEDIA_SCAN_TIMEOUT", 2*time.Minute)
	cfg.MediaClamAVBinary = getEnv("MEDIA_CLAMAV_BINARY", "clamscan")
	cfg.MediaClamAVArgs = getEnv("MEDIA_CLAMAV_ARGS", "--no-summary {source_path}")

	return cfg
}

// 3. Validate always enforces production media scanning and applies strict production checks when enabled.
func (c *Config) Validate() error {
	if c == nil {
		return nil
	}
	if isProductionLikeEnv(c.AppEnv) {
		switch strings.ToLower(strings.TrimSpace(c.MediaProcessingProvider)) {
		case "ffmpeg", "command":
			if strings.TrimSpace(c.MediaFFmpegBinary) == "" || strings.TrimSpace(c.MediaFFmpegArgs) == "" {
				return fmt.Errorf("MEDIA_FFMPEG_BINARY and MEDIA_FFMPEG_ARGS are required when MEDIA_PROCESSING_PROVIDER=%s", c.MediaProcessingProvider)
			}
			if _, err := exec.LookPath(strings.TrimSpace(c.MediaFFmpegBinary)); err != nil {
				return fmt.Errorf("MEDIA_FFMPEG_BINARY is not executable: %w", err)
			}
		default:
			return fmt.Errorf("MEDIA_PROCESSING_PROVIDER must use a real ffmpeg-compatible processor when APP_ENV=%s", c.AppEnv)
		}
		switch strings.ToLower(strings.TrimSpace(c.MediaScanProvider)) {
		case "clamav", "clamscan", "command":
			if strings.TrimSpace(c.MediaClamAVBinary) == "" || strings.TrimSpace(c.MediaClamAVArgs) == "" {
				return fmt.Errorf("MEDIA_CLAMAV_BINARY and MEDIA_CLAMAV_ARGS are required when MEDIA_SCAN_PROVIDER=%s", c.MediaScanProvider)
			}
			if _, err := exec.LookPath(strings.TrimSpace(c.MediaClamAVBinary)); err != nil {
				return fmt.Errorf("MEDIA_CLAMAV_BINARY is not executable: %w", err)
			}
		default:
			return fmt.Errorf("MEDIA_SCAN_PROVIDER must use a real ClamAV-compatible scanner when APP_ENV=%s", c.AppEnv)
		}
	}
	if !c.StrictProductionConfig || !isProductionLikeEnv(c.AppEnv) {
		return nil
	}
	if isUnsafeSecret(c.JWTSecret, "dev-secret-key") {
		return fmt.Errorf("JWT_SECRET must be set to a strong production value")
	}
	if isUnsafeSecret(c.EncryptionKey, "dev-encryption-key") {
		return fmt.Errorf("ENCRYPTION_KEY must be set to a strong production value")
	}
	switch strings.ToLower(strings.TrimSpace(c.OTPProvider)) {
	case "", "mock":
		return fmt.Errorf("OTP_PROVIDER=mock is not allowed when APP_ENV=%s", c.AppEnv)
	case "aliyun_sms":
		if strings.TrimSpace(c.AliyunSMSAccessKeyID) == "" || strings.TrimSpace(c.AliyunSMSAccessKeySecret) == "" || strings.TrimSpace(c.AliyunSMSSignName) == "" || strings.TrimSpace(c.AliyunSMSTemplateCode) == "" {
			return fmt.Errorf("Alibaba Cloud SMS credentials, sign name, and mainland template code are required when OTP_PROVIDER=aliyun_sms")
		}
		if c.OTPResendCooldown <= 0 || c.AliyunSMSRequestTimeout <= 0 {
			return fmt.Errorf("OTP_RESEND_COOLDOWN and ALIYUN_SMS_REQUEST_TIMEOUT must be greater than zero")
		}
	default:
		return fmt.Errorf("unsupported OTP_PROVIDER %q", c.OTPProvider)
	}
	if !c.MailEnabled {
		return fmt.Errorf("MAIL_ENABLED=false is not allowed when APP_ENV=%s", c.AppEnv)
	}
	if c.OSSCallbackEnabled && strings.TrimSpace(c.OSSCallbackSecret) == "" {
		return fmt.Errorf("OSS_CALLBACK_SECRET is required when OSS_CALLBACK_ENABLED=true")
	}
	if strings.EqualFold(strings.TrimSpace(c.MediaProcessingProvider), "ffmpeg") || strings.EqualFold(strings.TrimSpace(c.MediaProcessingProvider), "command") {
		if strings.TrimSpace(c.MediaFFmpegBinary) == "" || strings.TrimSpace(c.MediaFFmpegArgs) == "" {
			return fmt.Errorf("MEDIA_FFMPEG_BINARY and MEDIA_FFMPEG_ARGS are required when MEDIA_PROCESSING_PROVIDER=%s", c.MediaProcessingProvider)
		}
	}

	return nil
}

// 4. isProductionLikeEnv reports whether strict runtime safety checks apply.
func isProductionLikeEnv(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "production", "prod", "staging":
		return true
	default:
		return false
	}
}

// 4.1 defaultMediaProvider enables deterministic local media checks only in development and tests.
func defaultMediaProvider(appEnv string) string {
	switch strings.ToLower(strings.TrimSpace(appEnv)) {
	case "development", "dev", "test", "testing":
		return "mock"
	default:
		return "disabled"
	}
}

// 5. isUnsafeSecret rejects empty, default, and short production secrets.
func isUnsafeSecret(value string, defaultValue string) bool {
	trimmed := strings.TrimSpace(value)
	return trimmed == "" || trimmed == defaultValue || len(trimmed) < 32
}

// 6. getEnv returns env value or fallback.
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

// 6. getInt64Env parses integer env value or fallback.
func getInt64Env(key string, fallback int64) int64 {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}

	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return fallback
	}

	return value
}

// 6.1 getInt64EnvWithLegacy parses new or legacy integer env values.
func getInt64EnvWithLegacy(key string, legacyKey string, fallback int64) int64 {
	if os.Getenv(key) != "" {
		return getInt64Env(key, fallback)
	}
	if os.Getenv(legacyKey) != "" {
		return getInt64Env(legacyKey, fallback)
	}
	return fallback
}

// 7. getBoolEnv parses bool env value or fallback.
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

// 8. getBoolEnvWithLegacy parses new or legacy bool env value.
func getBoolEnvWithLegacy(key string, legacyKey string, fallback bool) bool {
	if os.Getenv(key) != "" {
		return getBoolEnv(key, fallback)
	}

	return getBoolEnv(legacyKey, fallback)
}

// 9. getDurationEnv parses duration env value or fallback.
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

// 10. getDurationEnvWithLegacy parses new or legacy duration env value.
func getDurationEnvWithLegacy(key string, legacyKey string, fallback time.Duration) time.Duration {
	if os.Getenv(key) != "" {
		return getDurationEnv(key, fallback)
	}

	return getDurationEnv(legacyKey, fallback)
}

// 11. SMTPAddress builds the outbound SMTP server address.
func (c *Config) SMTPAddress() string {
	return fmt.Sprintf("%s:%d", c.SMTPHost, c.SMTPPort)
}

// 12. defaultMediaBaseURL builds the default public media base URL.
func defaultMediaBaseURL(cfg *Config) string {
	if strings.EqualFold(strings.TrimSpace(cfg.StorageProvider), "oss") {
		return defaultOSSMediaBaseURL(cfg)
	}

	return strings.TrimRight(cfg.StorageEndpoint, "/") + "/" + strings.TrimLeft(cfg.StorageBucket, "/")
}

// 13. defaultOSSMediaBaseURL derives the public base URL for OSS-backed media.
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

// 14. defaultOSSCallbackURL builds the public upload callback endpoint.
func defaultOSSCallbackURL(cfg *Config) string {
	baseURL := strings.TrimRight(strings.TrimSpace(cfg.AppAPIPublicBaseURL), "/")
	if baseURL == "" {
		return ""
	}

	return baseURL + "/api/v1/oss/callback"
}

// 15. trimEndpointScheme removes scheme and surrounding slashes from endpoint values.
func trimEndpointScheme(endpoint string) string {
	value := strings.TrimSpace(endpoint)
	value = strings.TrimPrefix(value, "https://")
	value = strings.TrimPrefix(value, "http://")
	return strings.Trim(value, "/")
}

// 16. normalizePaymentGatewayEnv normalizes EasyLink environment selection.
func normalizePaymentGatewayEnv(value string) string {
	if strings.EqualFold(strings.TrimSpace(value), "production") {
		return "production"
	}

	return "sandbox"
}

// 17. defaultPaymentGatewayBaseURL returns the EasyLink gateway base URL.
func defaultPaymentGatewayBaseURL(env string) string {
	if normalizePaymentGatewayEnv(env) == "production" {
		return "https://api-pay.gnete.com.hk"
	}

	return "https://ts-api-pay.gnete.com.hk"
}

// 18. defaultWalletRechargeReturnBaseURL returns the deployed member web URL.
func defaultWalletRechargeReturnBaseURL(env string) string {
	if normalizePaymentGatewayEnv(env) == "production" {
		return "https://ajoliving.skylinedances.com"
	}

	return getEnv("APP_PUBLIC_BASE_URL", "http://localhost:5173")
}

// 19. defaultWalletRechargeNotifyURL returns the deployed payment callback URL.
func defaultWalletRechargeNotifyURL(env string) string {
	if normalizePaymentGatewayEnv(env) == "production" {
		return "https://ajoliving.server.skylinedances.com/api/v1/payments/easylink/notify"
	}

	return strings.TrimRight(getEnv("APP_API_PUBLIC_BASE_URL", "http://localhost:8081"), "/") + "/api/v1/payments/easylink/notify"
}

// 20. firstConfigValue returns the first non-empty config value.
func firstConfigValue(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}

	return ""
}

// 21. buildPaymentChannelExtra serializes one optional channelExtra field.
func buildPaymentChannelExtra(key string, value string) string {
	key = strings.TrimSpace(key)
	value = strings.TrimSpace(value)
	if key == "" || value == "" {
		return ""
	}

	raw, err := json.Marshal(map[string]string{key: value})
	if err != nil {
		return ""
	}

	return string(raw)
}

// 22. defaultPOSAPIBaseURL returns the deployed POS relay API base URL.
func defaultPOSAPIBaseURL() string {
	return "https://pos.ismart.skylinedances.com/api"
}

// 23. defaultPOSLoginURL builds the POS Web login endpoint from the API base URL.
func defaultPOSLoginURL(baseURL string) string {
	value := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if value == "" {
		value = defaultPOSAPIBaseURL()
	}

	return value + "/poslogin"
}

// 24. defaultPOSPaymentServiceURL returns the deployed POS H5 payment API base URL.
func defaultPOSPaymentServiceURL() string {
	return "https://easy.payment.skylinedances.com/api/payments"
}

// 25. defaultIsmartExternalAppBaseURL returns the deployed iSmart external app base URL.
func defaultIsmartExternalAppBaseURL() string {
	return "https://ismart.ajoliving.com"
}

// 26. defaultIsmartExternalAppAPIBaseURL returns the deployed iSmart external app API base URL.
func defaultIsmartExternalAppAPIBaseURL() string {
	return strings.TrimRight(getEnv("ISMART_EXTERNAL_APP_BASE_URL", defaultIsmartExternalAppBaseURL()), "/") + "/api/v1/external"
}

// 27. defaultIsmartIntegrationAPIBaseURL returns the deployed iSmart integration API base URL.
func defaultIsmartIntegrationAPIBaseURL() string {
	return strings.TrimRight(getEnv("ISMART_EXTERNAL_APP_BASE_URL", defaultIsmartExternalAppBaseURL()), "/") + "/api/v1/integration"
}

// 28. defaultIsmartServiceCaseAPIBaseURL returns the iSmart service-case API base URL.
func defaultIsmartServiceCaseAPIBaseURL() string {
	return "https://clouddev.ismart.ajoliving.com/api/v1/integration"
}

// 28.1 defaultIsmartSubaccountAPIBaseURL returns the iSmart subaccount API base URL.
func defaultIsmartSubaccountAPIBaseURL() string {
	return "https://clouddev.ismart.ajoliving.com/api/v1/integration"
}

// 29. defaultICCTVAPIBaseURL returns the deployed iCCTV API base URL.
func defaultICCTVAPIBaseURL() string {
	return "https://icctv.skylinedances.com/api"
}

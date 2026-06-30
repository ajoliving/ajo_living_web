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
	POSPaymentServiceURL        string
	IsmartExternalAppBaseURL    string
	IsmartExternalAppAPIBaseURL string
	IsmartExternalAppTimeout    time.Duration
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
	OTPProvider                 string
	OTPMockCode                 string
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
	OSSCallbackEnabled          bool
	OSSCallbackURL              string
	OSSAllowedOrigins           string
	MediaBaseURL                string
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
		POSPaymentServiceURL:        getEnv("POS_PAYMENT_SERVICE_URL", defaultPOSPaymentServiceURL()),
		IsmartExternalAppBaseURL:    getEnv("ISMART_EXTERNAL_APP_BASE_URL", defaultIsmartExternalAppBaseURL()),
		IsmartExternalAppAPIBaseURL: getEnv("ISMART_EXTERNAL_APP_API_BASE_URL", defaultIsmartExternalAppAPIBaseURL()),
		IsmartExternalAppTimeout:    getDurationEnv("ISMART_EXTERNAL_APP_TIMEOUT", 10*time.Second),
		POSTerminalProxyEnabled:     getBoolEnv("POS_TERMINAL_PROXY_ENABLED", false),
		GoodPriceAPIBaseURL:         getEnv("GOOD_PRICE_API_BASE_URL", "https://good.price.skylinedances.com/api"),
		GoodPriceRequestTimeout:     getDurationEnv("GOOD_PRICE_REQUEST_TIMEOUT", 10*time.Second),
		GoodPriceCacheTTL:           getDurationEnv("GOOD_PRICE_CACHE_TTL", 5*time.Minute),
		GoodPriceSummaryCacheTTL:    getDurationEnv("GOOD_PRICE_SUMMARY_CACHE_TTL", 5*time.Minute),
		GoodPriceSearchCacheTTL:     getDurationEnv("GOOD_PRICE_SEARCH_CACHE_TTL", 2*time.Minute),
		GoodPriceDetailCacheTTL:     getDurationEnv("GOOD_PRICE_DETAIL_CACHE_TTL", 3*time.Minute),
		GoodPriceAlertEnabled:       getBoolEnv("GOOD_PRICE_ALERT_ENABLED", true),
		GoodPriceAlertInterval:      getDurationEnv("GOOD_PRICE_ALERT_INTERVAL", time.Hour),
		POSTerminalType:             getEnv("POS_TERMINAL_TYPE", "allinpay"),
		POSTerminalURL:              getEnv("POS_TERMINAL_URL", ""),
		POSTerminalAllowedBuildings: getEnv("POS_TERMINAL_ALLOWED_BUILDINGS", ""),
		OTPProvider:                 getEnv("OTP_PROVIDER", "mock"),
		OTPMockCode:                 getEnv("OTP_MOCK_CODE", "123456"),
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
		OSSCallbackEnabled:          getBoolEnv("OSS_CALLBACK_ENABLED", false),
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

/*
 * CORS middleware.
 * 1. Allows browser clients from configured frontend origins.
 * 2. Handles OPTIONS preflight before auth and route matching.
 * 3. Keeps credential and authorization headers available for API calls.
 */
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"ajoliving_web/http_service/internal/config"
)

// 1. CORS returns API CORS headers for configured frontend origins.
func CORS(cfg *config.Config) gin.HandlerFunc {
	allowedOrigins := buildAllowedOrigins(cfg)

	return func(c *gin.Context) {
		origin := strings.TrimSpace(c.GetHeader("Origin"))
		if _, allowed := allowedOrigins[origin]; allowed {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Authorization,Content-Type,X-Request-ID")
			c.Header("Access-Control-Expose-Headers", "X-Request-ID")
			c.Header("Access-Control-Max-Age", "600")
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// 2. buildAllowedOrigins creates an exact-match origin allowlist.
func buildAllowedOrigins(cfg *config.Config) map[string]struct{} {
	origins := map[string]struct{}{}
	if cfg == nil {
		return origins
	}

	appendOrigin(origins, cfg.AppPublicBaseURL)
	appendOrigin(origins, cfg.AppAPIPublicBaseURL)
	for _, origin := range strings.Split(cfg.CORSAllowedOrigins, ",") {
		appendOrigin(origins, origin)
	}
	for _, origin := range []string{
		"http://localhost:5173",
		"http://127.0.0.1:5173",
	} {
		appendOrigin(origins, origin)
	}

	return origins
}

// 3. appendOrigin stores one normalized origin.
func appendOrigin(origins map[string]struct{}, origin string) {
	trimmed := strings.TrimRight(strings.TrimSpace(origin), "/")
	if trimmed == "" {
		return
	}
	origins[trimmed] = struct{}{}
}

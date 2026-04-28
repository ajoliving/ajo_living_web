/*
 * Access logging middleware.
 * 1. Record request path, method, latency, and status.
 * 2. Keep logging centralized and consistent.
 */
package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"

	"ajoliving_web/http_service/internal/errcode"
)

// 1. AccessLog logs every incoming request.
func AccessLog(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		startedAt := time.Now()
		c.Next()

		log.Info("http request",
			slog.String("request_id", errcode.RequestID(c)),
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status", c.Writer.Status()),
			slog.Duration("latency", time.Since(startedAt)),
			slog.String("client_ip", c.ClientIP()),
		)
	}
}

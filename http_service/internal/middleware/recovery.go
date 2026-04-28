/*
 * Panic recovery middleware.
 * 1. Recover from unexpected panics.
 * 2. Log the panic with request context.
 * 3. Return the unified internal error response.
 */
package middleware

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	"ajoliving_web/http_service/internal/errcode"
)

// 1. Recovery converts panics to unified error responses.
func Recovery(log *slog.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		log.Error("panic recovered",
			slog.Any("panic", recovered),
			slog.String("path", c.Request.URL.Path),
			slog.String("method", c.Request.Method),
			slog.String("request_id", errcode.RequestID(c)),
		)
		errcode.WriteError(c, errcode.New(errcode.CodeInternalError, "internal server error"))
	})
}

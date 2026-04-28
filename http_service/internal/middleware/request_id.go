/*
 * Request metadata middleware.
 * 1. Ensure every request has a request ID.
 * 2. Propagate request ID through headers and gin context.
 */
package middleware

import (
	"github.com/gin-gonic/gin"

	"ajoliving_web/http_service/internal/utils"
)

// 1. RequestID injects request id into every request context.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = utils.NewPublicID()
		}

		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

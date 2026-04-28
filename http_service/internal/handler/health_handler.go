/*
 * Health check handler.
 * 1. Expose the service health endpoint.
 */
package handler

import (
	"github.com/gin-gonic/gin"

	"ajoliving_web/http_service/internal/errcode"
)

// 1. HealthHandler handles service health checks.
type HealthHandler struct{}

// 2. NewHealthHandler creates a health handler instance.
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// 3. Check returns the basic health status response.
func (h *HealthHandler) Check(c *gin.Context) {
	errcode.Success(c, gin.H{
		"status": "ok",
	})
}

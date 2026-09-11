/*
 * Health check handler.
 * 1. Expose liveness and database readiness endpoints.
 */
package handler

import (
	"github.com/gin-gonic/gin"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/service"
)

// 1. HealthHandler handles service health checks.
type HealthHandler struct {
	runtime *service.Runtime
}

// 2. NewHealthHandler creates a health handler instance.
func NewHealthHandler(runtime ...*service.Runtime) *HealthHandler {
	var value *service.Runtime
	if len(runtime) > 0 {
		value = runtime[0]
	}
	return &HealthHandler{runtime: value}
}

// 3. Check returns the basic health status response.
func (h *HealthHandler) Check(c *gin.Context) {
	errcode.Success(c, gin.H{
		"status": "ok",
	})
}

// 4. Ready verifies the database dependency before accepting traffic.
func (h *HealthHandler) Ready(c *gin.Context) {
	if h == nil || h.runtime == nil || h.runtime.DB == nil {
		c.JSON(503, gin.H{"status": "not_ready"})
		return
	}
	if err := h.runtime.DB.WithContext(c.Request.Context()).Exec("SELECT 1").Error; err != nil {
		c.JSON(503, gin.H{"status": "not_ready"})
		return
	}
	errcode.Success(c, gin.H{"status": "ready", "dependencies": gin.H{"database": "ok"}})
}

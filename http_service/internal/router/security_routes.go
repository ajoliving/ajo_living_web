/*
 * Security route registration.
 * 1. Register authenticated iCCTV camera lookup routes.
 * 2. Keep device and security integrations outside marketplace routes.
 */
package router

import (
	"github.com/gin-gonic/gin"

	"ajoliving_web/http_service/internal/handler"
)

// 1. registerSecurityRoutes registers authenticated security module routes.
func registerSecurityRoutes(
	api *gin.RouterGroup,
	securityICCTVHandler *handler.SecurityICCTVHandler,
	requireAuth gin.HandlerFunc,
) {
	api.GET("/me/security/icctv/public-cameras", requireAuth, securityICCTVHandler.GetPublicCameras)
}

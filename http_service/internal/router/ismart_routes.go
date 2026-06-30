/*
 * iSmart proxy route registration.
 * 1. Keep all iSmart external API proxy routes in one file.
 * 2. Require AJO authentication before proxying to old iSmart services.
 */
package router

import (
	"github.com/gin-gonic/gin"

	"ajoliving_web/http_service/internal/handler"
)

// 1. registerIsmartRoutes registers authenticated iSmart proxy routes.
func registerIsmartRoutes(
	api *gin.RouterGroup,
	ismartHandler *handler.IsmartExternalHandler,
	requireAuth gin.HandlerFunc,
) {
	api.GET("/me/ismart/buildings", requireAuth, ismartHandler.ListBuildings)
	api.GET("/me/ismart/building-info", requireAuth, ismartHandler.GetBuildingInfo)
	api.POST("/me/ismart/building-comments", requireAuth, ismartHandler.SubmitBuildingComment)
	api.GET("/me/ismart/building-access", requireAuth, ismartHandler.GetBuildingAccess)
	api.POST("/me/ismart/building-access/open-door", requireAuth, ismartHandler.OpenDoor)
	api.POST("/me/ismart/building-access/qrcode", requireAuth, ismartHandler.GenerateQRCode)
	api.POST("/me/ismart/pos-payment-to-ismart", requireAuth, ismartHandler.SubmitPOSPayment)
}

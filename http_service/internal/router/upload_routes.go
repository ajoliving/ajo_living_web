/*
 * Upload route registration.
 * 1. Register OSS callback, presign, complete, and asset management routes.
 * 2. Keep upload routing separate from marketplace and property publishing routes.
 */
package router

import (
	"github.com/gin-gonic/gin"

	"ajoliving_web/http_service/internal/handler"
)

// 1. registerUploadRoutes registers OSS and upload routes.
func registerUploadRoutes(api *gin.RouterGroup, uploadHandler *handler.UploadHandler, requireAuth gin.HandlerFunc) {
	api.POST("/oss/callback", uploadHandler.UploadCallback)
	api.POST("/oss/presign", requireAuth, uploadHandler.Presign)
	api.POST("/oss/complete", requireAuth, uploadHandler.CompleteUpload)
	api.GET("/oss/assets", requireAuth, uploadHandler.ListAssets)
	api.GET("/oss/assets/:mediaAssetId", requireAuth, uploadHandler.GetAsset)
	api.DELETE("/oss/assets/:mediaAssetId", requireAuth, uploadHandler.DeleteAsset)
	api.POST("/uploads/presign", requireAuth, uploadHandler.Presign)
	api.POST("/uploads/complete", requireAuth, uploadHandler.CompleteUpload)
}

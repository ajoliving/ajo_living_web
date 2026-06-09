/*
 * Marketplace order route registration.
 * 1. Register secondhand listing order lifecycle routes.
 * 2. Keep marketplace orders separate from POS payment orders.
 */
package router

import (
	"github.com/gin-gonic/gin"

	"ajoliving_web/http_service/internal/handler"
)

// 1. registerOrderRoutes registers marketplace order routes.
func registerOrderRoutes(api *gin.RouterGroup, orderHandler *handler.OrderHandler, requireAuth gin.HandlerFunc) {
	api.POST("/listings/:listingId/orders", requireAuth, orderHandler.Create)
	api.GET("/orders/:orderId", requireAuth, orderHandler.GetDetail)
	api.POST("/orders/:orderId/confirm", requireAuth, orderHandler.Confirm)
	api.POST("/orders/:orderId/cancel", requireAuth, orderHandler.Cancel)
	api.POST("/orders/:orderId/complete", requireAuth, orderHandler.Complete)
}

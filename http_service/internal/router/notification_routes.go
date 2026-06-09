/*
 * Notification route registration.
 * 1. Register member notification list, unread count, and read state routes.
 * 2. Keep notification routes separate from communication message routes.
 */
package router

import (
	"github.com/gin-gonic/gin"

	"ajoliving_web/http_service/internal/handler"
)

// 1. registerNotificationRoutes registers authenticated notification routes.
func registerNotificationRoutes(api *gin.RouterGroup, notificationHandler *handler.NotificationHandler, requireAuth gin.HandlerFunc) {
	api.GET("/notifications", requireAuth, notificationHandler.List)
	api.GET("/notifications/unread-count", requireAuth, notificationHandler.UnreadCount)
	api.POST("/notifications/:notificationId/read", requireAuth, notificationHandler.MarkRead)
	api.POST("/notifications/read-all", requireAuth, notificationHandler.MarkAllRead)
}

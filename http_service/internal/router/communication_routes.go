/*
 * Communication route registration.
 * 1. Register chat, message, and read state routes.
 * 2. Keep platform communication routes separate from marketplace listing routes.
 */
package router

import (
	"time"

	"github.com/gin-gonic/gin"

	"ajoliving_web/http_service/internal/handler"
	"ajoliving_web/http_service/internal/middleware"
)

// 1. registerChatRoutes registers authenticated chat routes.
func registerChatRoutes(
	api *gin.RouterGroup,
	chatHandler *handler.ChatHandler,
	requireAuth gin.HandlerFunc,
	limiter *middleware.InMemoryLimiter,
) {
	api.GET("/chats", requireAuth, chatHandler.ListChats)
	api.GET("/chats/:chatId", requireAuth, chatHandler.GetChat)
	api.GET("/chats/:chatId/messages", requireAuth, chatHandler.ListMessages)
	api.POST("/chats/:chatId/messages",
		requireAuth,
		limiter.Limit(30, time.Minute, func(c *gin.Context) string {
			return "send_message:" + c.ClientIP()
		}),
		chatHandler.SendMessage,
	)
	api.POST("/chats/:chatId/read", requireAuth, chatHandler.MarkRead)
}

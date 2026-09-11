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

// 1. registerRealtimeRoutes registers the ticket and WebSocket endpoints.
func registerRealtimeRoutes(api *gin.RouterGroup, realtimeHandler *handler.RealtimeHandler, requireActive gin.HandlerFunc) {
	api.GET("/realtime/ticket", requireActive, realtimeHandler.IssueTicket)
	api.GET("/realtime/ws", realtimeHandler.HandleWebSocket)
}

// 2. registerChatRoutes registers authenticated chat routes.
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
	api.GET("/building-chats", requireAuth, chatHandler.ListBuildingChats)
	api.POST("/buildings/:buildingId/chat", requireAuth, chatHandler.EnsureBuildingChat)
	api.GET("/building-chats/:chatId", requireAuth, chatHandler.GetBuildingChat)
	api.GET("/building-chats/:chatId/members", requireAuth, chatHandler.ListBuildingChatMembers)
	api.GET("/building-chats/:chatId/join-requests", requireAuth, chatHandler.ListBuildingChatJoinRequests)
	api.GET("/building-chats/:chatId/messages", requireAuth, chatHandler.ListBuildingChatMessages)
	api.POST("/building-chats/:chatId/messages",
		requireAuth,
		limiter.Limit(30, time.Minute, func(c *gin.Context) string {
			return "send_building_message:" + c.ClientIP()
		}),
		chatHandler.SendBuildingChatMessage,
	)
	api.POST("/building-chats/:chatId/read", requireAuth, chatHandler.MarkBuildingChatRead)
	api.POST("/building-chats/:chatId/leave", requireAuth, chatHandler.LeaveBuildingChat)
	api.POST("/building-chats/:chatId/join-requests", requireAuth, chatHandler.RequestBuildingChatJoin)
	api.POST("/building-chat-join-requests/:requestId/review", requireAuth, chatHandler.ReviewBuildingChatJoin)
	api.POST("/building-chats/:chatId/members/moderate", requireAuth, chatHandler.ModerateBuildingChatMember)
}

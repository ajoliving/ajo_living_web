/*
 * Chat HTTP handlers.
 * 1. Bind chat creation, message, and read payloads.
 * 2. Delegate chat flows to the service layer.
 */
package handler

import (
	"strings"

	"github.com/gin-gonic/gin"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/service"
)

// 1. ChatHandler handles chat endpoints.
type ChatHandler struct {
	chatService *service.ChatService
}

// 2. messageRequest defines the send message payload.
type messageRequest struct {
	Content string `json:"content" binding:"required"`
}

// 3. NewChatHandler creates a chat handler instance.
func NewChatHandler(chatService *service.ChatService) *ChatHandler {
	return &ChatHandler{chatService: chatService}
}

// 4. CreateOrReuse creates or reuses a listing chat.
func (h *ChatHandler) CreateOrReuse(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	result, err := h.chatService.CreateOrReuseChat(c.Request.Context(), user.UserID, user.PrimaryCommunityID, strings.TrimSpace(c.Param("listingId")))
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, result)
}

// 5. ListChats returns chats owned by the current member.
func (h *ChatHandler) ListChats(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	page, pageSize := parsePagination(c)
	items, pagination, err := h.chatService.ListChats(c.Request.Context(), user.UserID, page, pageSize)
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, gin.H{"items": items, "pagination": pagination})
}

// 6. GetChat returns one chat.
func (h *ChatHandler) GetChat(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	result, err := h.chatService.GetChat(c.Request.Context(), user.UserID, strings.TrimSpace(c.Param("chatId")))
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, result)
}

// 7. ListMessages returns chat messages.
func (h *ChatHandler) ListMessages(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	page, pageSize := parsePagination(c)
	items, pagination, err := h.chatService.ListMessages(c.Request.Context(), user.UserID, strings.TrimSpace(c.Param("chatId")), page, pageSize)
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, gin.H{"items": items, "pagination": pagination})
}

// 8. SendMessage sends a new chat message.
func (h *ChatHandler) SendMessage(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	var request messageRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid request payload"))
		return
	}

	result, err := h.chatService.SendMessage(c.Request.Context(), user.UserID, strings.TrimSpace(c.Param("chatId")), strings.TrimSpace(request.Content))
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, result)
}

// 9. MarkRead marks a chat as read.
func (h *ChatHandler) MarkRead(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	if err := h.chatService.MarkRead(c.Request.Context(), user.UserID, strings.TrimSpace(c.Param("chatId"))); err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, gin.H{"chat_id": strings.TrimSpace(c.Param("chatId")), "read": true})
}

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

// 4. CreateOrReuse creates or reuses a secondhand listing chat.
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

// 5. CreateOrReusePropertySale creates or reuses a property sale chat.
func (h *ChatHandler) CreateOrReusePropertySale(c *gin.Context) {
	h.createOrReuseProperty(c, service.PropertyChannelSale)
}

// 6. CreateOrReuseServicedApartment creates or reuses a serviced apartment chat.
func (h *ChatHandler) CreateOrReuseServicedApartment(c *gin.Context) {
	h.createOrReuseProperty(c, service.PropertyChannelServiced)
}

// 7. createOrReuseProperty creates or reuses a property listing chat.
func (h *ChatHandler) createOrReuseProperty(c *gin.Context, channel service.PropertyChannel) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	result, err := h.chatService.CreateOrReusePropertyChat(c.Request.Context(), channel, user.UserID, strings.TrimSpace(c.Param("listingId")))
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, result)
}

// 8. ListChats returns chats owned by the current member.
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

// 9. GetChat returns one chat.
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

// 10. ListMessages returns chat messages.
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

// 11. SendMessage sends a new chat message.
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

// 12. MarkRead marks a chat as read.
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

/*
 * Chat HTTP handlers.
 * 1. Bind chat creation, message, and read payloads.
 * 2. Delegate chat flows to the service layer.
 */
package handler

import (
	"strconv"
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

// 2. buildingChatJoinRequestPayload defines a join request payload.
type buildingChatJoinRequestPayload struct {
	Reason string `json:"reason"`
}

// 3. buildingChatReviewPayload defines an administrator review payload.
type buildingChatReviewPayload struct {
	Status string `json:"status" binding:"required"`
}

// 4. buildingChatModerationPayload defines a member moderation payload.
type buildingChatModerationPayload struct {
	UserID          int64  `json:"user_id" binding:"required"`
	Action          string `json:"action" binding:"required"`
	DurationMinutes int    `json:"duration_minutes"`
	Reason          string `json:"reason"`
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

// 13. ListBuildingChats returns groups available to the current member.
func (h *ChatHandler) ListBuildingChats(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}
	items, err := h.chatService.ListBuildingChats(c.Request.Context(), user.UserID)
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, gin.H{"items": items})
}

// 14. EnsureBuildingChat creates or joins the group for one visible building.
func (h *ChatHandler) EnsureBuildingChat(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}
	result, err := h.chatService.EnsureBuildingChat(c.Request.Context(), user.UserID, strings.TrimSpace(c.Param("buildingId")))
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 15. GetBuildingChat returns one building group detail.
func (h *ChatHandler) GetBuildingChat(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}
	result, err := h.chatService.GetBuildingChat(c.Request.Context(), user.UserID, strings.TrimSpace(c.Param("chatId")))
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 16. ListBuildingChatMembers returns group members.
func (h *ChatHandler) ListBuildingChatMembers(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}
	items, err := h.chatService.ListBuildingChatMembers(c.Request.Context(), user.UserID, strings.TrimSpace(c.Param("chatId")))
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, gin.H{"items": items})
}

// 17. ListBuildingChatJoinRequests returns the administrator request queue.
func (h *ChatHandler) ListBuildingChatJoinRequests(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}
	items, err := h.chatService.ListBuildingChatJoinRequests(c.Request.Context(), user.UserID, strings.TrimSpace(c.Param("chatId")))
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, gin.H{"items": items})
}

// 18. ListBuildingChatMessages returns group messages.
func (h *ChatHandler) ListBuildingChatMessages(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}
	page, pageSize := parsePagination(c)
	items, pagination, err := h.chatService.ListBuildingChatMessages(c.Request.Context(), user.UserID, strings.TrimSpace(c.Param("chatId")), page, pageSize)
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, gin.H{"items": items, "pagination": pagination})
}

// 19. SendBuildingChatMessage sends a text message to a group.
func (h *ChatHandler) SendBuildingChatMessage(c *gin.Context) {
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
	result, err := h.chatService.SendBuildingChatMessage(c.Request.Context(), user.UserID, strings.TrimSpace(c.Param("chatId")), strings.TrimSpace(request.Content))
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 20. MarkBuildingChatRead marks a group as read.
func (h *ChatHandler) MarkBuildingChatRead(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}
	if err := h.chatService.MarkBuildingChatRead(c.Request.Context(), user.UserID, strings.TrimSpace(c.Param("chatId"))); err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, gin.H{"chat_id": strings.TrimSpace(c.Param("chatId")), "read": true})
}

// 21. LeaveBuildingChat leaves a group while preserving the membership record.
func (h *ChatHandler) LeaveBuildingChat(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}
	if err := h.chatService.LeaveBuildingChat(c.Request.Context(), user.UserID, strings.TrimSpace(c.Param("chatId"))); err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, gin.H{"chat_id": strings.TrimSpace(c.Param("chatId")), "left": true})
}

// 22. RequestBuildingChatJoin submits an approval-required join request.
func (h *ChatHandler) RequestBuildingChatJoin(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}
	var request buildingChatJoinRequestPayload
	if err := c.ShouldBindJSON(&request); err != nil && err.Error() != "EOF" {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid request payload"))
		return
	}
	result, err := h.chatService.RequestBuildingChatJoin(c.Request.Context(), user.UserID, strings.TrimSpace(c.Param("chatId")), strings.TrimSpace(request.Reason))
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 23. ReviewBuildingChatJoin approves or rejects a pending request.
func (h *ChatHandler) ReviewBuildingChatJoin(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}
	var request buildingChatReviewPayload
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid request payload"))
		return
	}
	result, err := h.chatService.ReviewBuildingChatJoin(c.Request.Context(), user.UserID, strings.TrimSpace(c.Param("requestId")), strings.TrimSpace(request.Status))
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 24. ModerateBuildingChatMember applies an administrator member action.
func (h *ChatHandler) ModerateBuildingChatMember(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}
	var request buildingChatModerationPayload
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid request payload"))
		return
	}
	if request.UserID <= 0 {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "user_id is required"))
		return
	}
	if err := h.chatService.ModerateBuildingChatMember(c.Request.Context(), user.UserID, strings.TrimSpace(c.Param("chatId")), request.UserID, strings.TrimSpace(request.Action), request.DurationMinutes, strings.TrimSpace(request.Reason)); err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, gin.H{"chat_id": strings.TrimSpace(c.Param("chatId")), "user_id": strconv.FormatInt(request.UserID, 10), "action": request.Action})
}

// 25. ListAllBuildingChatsForStaff returns all building chats for staff management.
func (h *ChatHandler) ListAllBuildingChatsForStaff(c *gin.Context) {
	items, err := h.chatService.ListAllBuildingChatsForStaff(c.Request.Context())
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, gin.H{"items": items})
}

/*
 * Notification HTTP handlers.
 * 1. Bind inbox list and read actions.
 * 2. Delegate notification flows to the service layer.
 */
package handler

import (
	"strings"

	"github.com/gin-gonic/gin"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/service"
)

// 1. NotificationHandler handles inbox endpoints.
type NotificationHandler struct {
	notificationService *service.NotificationService
}

// 2. systemNoticeRequest defines the staff notice broadcast payload.
type systemNoticeRequest struct {
	Title string `json:"title" binding:"required"`
	Body  string `json:"body" binding:"required"`
}

// 3. NewNotificationHandler creates a notification handler instance.
func NewNotificationHandler(notificationService *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{notificationService: notificationService}
}

// 4. List returns the current member's notifications.
func (h *NotificationHandler) List(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	page, pageSize := parsePagination(c)
	items, pagination, unreadCount, err := h.notificationService.ListNotifications(c.Request.Context(), user.UserID, service.NotificationListFilters{
		Page:       page,
		PageSize:   pageSize,
		OnlyUnread: parseBoolQuery(c, "only_unread"),
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, gin.H{"items": items, "pagination": pagination, "unread_count": unreadCount})
}

// 5. UnreadCount returns the current member's unread notification count.
func (h *NotificationHandler) UnreadCount(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	count, err := h.notificationService.UnreadCount(c.Request.Context(), user.UserID)
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, gin.H{"unread_count": count})
}

// 6. MarkRead marks one notification as read.
func (h *NotificationHandler) MarkRead(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	if err := h.notificationService.MarkRead(c.Request.Context(), user.UserID, strings.TrimSpace(c.Param("notificationId"))); err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, gin.H{"notification_id": strings.TrimSpace(c.Param("notificationId")), "read": true})
}

// 7. MarkAllRead marks all unread notifications as read.
func (h *NotificationHandler) MarkAllRead(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	rows, err := h.notificationService.MarkAllRead(c.Request.Context(), user.UserID)
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, gin.H{"updated": rows})
}

// 8. PublishSystemNotice broadcasts one notification-center item to all members.
func (h *NotificationHandler) PublishSystemNotice(c *gin.Context) {
	var request systemNoticeRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid request payload"))
		return
	}

	result, err := h.notificationService.PublishSystemNotice(c.Request.Context(), service.SystemNoticePublishParams{
		Title: request.Title,
		Body:  request.Body,
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, result)
}

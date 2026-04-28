/*
 * Notification business logic.
 * 1. Manage the member-facing notification inbox.
 * 2. Provide reusable helpers for order and chat event delivery.
 */
package service

import (
	"context"
	"strings"
	"time"

	"gorm.io/gorm"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/model"
	"ajoliving_web/http_service/internal/utils"
)

// 1. NotificationService manages inbox data.
type NotificationService struct {
	runtime *Runtime
}

// 2. NotificationItem defines a notification list item payload.
type NotificationItem struct {
	NotificationID  string  `json:"notification_id"`
	Category        string  `json:"category"`
	Title           string  `json:"title"`
	Body            string  `json:"body"`
	RelatedType     string  `json:"related_type"`
	RelatedPublicID string  `json:"related_public_id"`
	IsRead          bool    `json:"is_read"`
	ReadAt          *string `json:"read_at,omitempty"`
	CreatedAt       string  `json:"created_at"`
}

// 3. NotificationListFilters defines inbox list filters.
type NotificationListFilters struct {
	Page       int
	PageSize   int
	OnlyUnread bool
}

// 4. CreateNotificationParams defines the reusable create payload.
type CreateNotificationParams struct {
	UserID          int64
	Category        string
	Title           string
	Body            string
	RelatedType     string
	RelatedPublicID string
}

// 5. NewNotificationService creates a notification service instance.
func NewNotificationService(runtime *Runtime) *NotificationService {
	return &NotificationService{runtime: runtime}
}

// 6. ListNotifications returns one user's inbox items.
func (s *NotificationService) ListNotifications(ctx context.Context, userID int64, filters NotificationListFilters) ([]NotificationItem, *model.Pagination, int64, error) {
	page, pageSize := normalizePagination(filters.Page, filters.PageSize)
	query := s.runtime.DB.WithContext(ctx).Model(&model.Notification{}).Where("user_id = ?", userID)
	if filters.OnlyUnread {
		query = query.Where("is_read = ?", false)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, nil, 0, errcode.New(errcode.CodeInternalError, "failed to count notifications")
	}

	var unreadCount int64
	if err := s.runtime.DB.WithContext(ctx).Model(&model.Notification{}).Where("user_id = ? AND is_read = ?", userID, false).Count(&unreadCount).Error; err != nil {
		return nil, nil, 0, errcode.New(errcode.CodeInternalError, "failed to count unread notifications")
	}

	var notifications []model.Notification
	if err := query.Order("created_at desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&notifications).Error; err != nil {
		return nil, nil, 0, errcode.New(errcode.CodeInternalError, "failed to load notifications")
	}

	items := make([]NotificationItem, 0, len(notifications))
	for _, notification := range notifications {
		items = append(items, toNotificationItem(&notification))
	}

	return items, &model.Pagination{Page: page, PageSize: pageSize, Total: total}, unreadCount, nil
}

// 7. MarkRead marks a notification as read for the target user.
func (s *NotificationService) MarkRead(ctx context.Context, userID int64, notificationPublicID string) error {
	result := s.runtime.DB.WithContext(ctx).Model(&model.Notification{}).
		Where("public_id = ? AND user_id = ?", strings.TrimSpace(notificationPublicID), userID).
		Updates(map[string]any{
			"is_read": true,
			"read_at": s.runtime.Now(),
		})
	if result.Error != nil {
		return errcode.New(errcode.CodeInternalError, "failed to update notification")
	}
	if result.RowsAffected == 0 {
		return errcode.New(errcode.CodeNotFound, "notification not found")
	}

	return nil
}

// 8. MarkAllRead marks all unread notifications as read for the target user.
func (s *NotificationService) MarkAllRead(ctx context.Context, userID int64) (int64, error) {
	result := s.runtime.DB.WithContext(ctx).Model(&model.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Updates(map[string]any{
			"is_read": true,
			"read_at": s.runtime.Now(),
		})
	if result.Error != nil {
		return 0, errcode.New(errcode.CodeInternalError, "failed to update notifications")
	}

	return result.RowsAffected, nil
}

// 9. CreateNotification persists one notification inside an existing transaction.
func (s *NotificationService) CreateNotification(ctx context.Context, tx *gorm.DB, params CreateNotificationParams) error {
	if params.UserID <= 0 {
		return errcode.New(errcode.CodeValidationError, "notification user is required")
	}
	if strings.TrimSpace(params.Category) == "" || strings.TrimSpace(params.Title) == "" || strings.TrimSpace(params.Body) == "" {
		return errcode.New(errcode.CodeValidationError, "notification content is required")
	}

	record := model.Notification{
		PublicID:        utils.NewPublicID(),
		UserID:          params.UserID,
		Category:        strings.TrimSpace(params.Category),
		Title:           strings.TrimSpace(params.Title),
		Body:            strings.TrimSpace(params.Body),
		RelatedType:     strings.TrimSpace(params.RelatedType),
		RelatedPublicID: strings.TrimSpace(params.RelatedPublicID),
		IsRead:          false,
		CreatedAt:       s.runtime.Now(),
	}
	if err := tx.WithContext(ctx).Create(&record).Error; err != nil {
		return errcode.New(errcode.CodeInternalError, "failed to create notification")
	}

	return nil
}

// 10. toNotificationItem maps one model notification into the response payload.
func toNotificationItem(notification *model.Notification) NotificationItem {
	result := NotificationItem{
		NotificationID:  notification.PublicID,
		Category:        notification.Category,
		Title:           notification.Title,
		Body:            notification.Body,
		RelatedType:     notification.RelatedType,
		RelatedPublicID: notification.RelatedPublicID,
		IsRead:          notification.IsRead,
		CreatedAt:       notification.CreatedAt.UTC().Format(time.RFC3339),
	}
	if notification.ReadAt != nil {
		value := notification.ReadAt.UTC().Format(time.RFC3339)
		result.ReadAt = &value
	}

	return result
}

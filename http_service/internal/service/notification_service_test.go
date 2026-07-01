/*
 * Notification service tests.
 * 1. Validate staff system notice publishing creates notification-center entries only.
 * 2. Keep broadcast recipient filtering independent from chat conversations.
 */
package service

import (
	"context"
	"testing"
	"time"

	"ajoliving_web/http_service/internal/database"
	"ajoliving_web/http_service/internal/model"
)

// 1. TestPublishSystemNoticeCreatesInboxNotifications validates member inbox delivery.
func TestPublishSystemNoticeCreatesInboxNotifications(t *testing.T) {
	ctx := context.Background()
	runtimeValue := newAuthTestRuntime(
		t,
		nil,
		&model.User{},
		&model.UserProfile{},
		&model.Chat{},
		&model.ChatParticipant{},
		&model.Message{},
		&model.Notification{},
	)
	runtimeValue.Now = func() time.Time {
		return time.Date(2026, 6, 29, 10, 0, 0, 0, time.UTC)
	}
	if err := database.SeedSystemNotificationAccount(ctx, runtimeValue.DB); err != nil {
		t.Fatalf("seed system notification account: %v", err)
	}

	activeUser := createNoticeTestUser(t, runtimeValue, "+852", "61230001", "active")
	inactiveUser := createNoticeTestUser(t, runtimeValue, "+852", "61230002", "inactive")
	notificationService := NewNotificationService(runtimeValue)

	result, err := notificationService.PublishSystemNotice(ctx, SystemNoticePublishParams{
		Title: "系統維護通知",
		Body:  "今晚 10 時系統會進行例行維護。",
	})
	if err != nil {
		t.Fatalf("publish system notice: %v", err)
	}
	if result.DeliveredCount != 1 {
		t.Fatalf("expected one active recipient, got %d", result.DeliveredCount)
	}

	var notification model.Notification
	if err := runtimeValue.DB.Where("user_id = ?", activeUser.ID).First(&notification).Error; err != nil {
		t.Fatalf("load active user notification: %v", err)
	}
	if notification.Category != "system_notice" || notification.Title != "系統維護通知" || notification.Body != "今晚 10 時系統會進行例行維護。" {
		t.Fatalf("unexpected notification payload: %+v", notification)
	}
	if notification.RelatedType != "notification" || notification.RelatedPublicID != "" || notification.IsRead {
		t.Fatalf("expected unread notification-center item, got %+v", notification)
	}

	var chatCount int64
	if err := runtimeValue.DB.Model(&model.Chat{}).Where("chat_type = ?", "system_notice").Count(&chatCount).Error; err != nil {
		t.Fatalf("count system notice chats: %v", err)
	}
	if chatCount != 0 {
		t.Fatalf("expected no system notice chat, got %d", chatCount)
	}

	var inactiveCount int64
	if err := runtimeValue.DB.Model(&model.Notification{}).Where("user_id = ?", inactiveUser.ID).Count(&inactiveCount).Error; err != nil {
		t.Fatalf("count inactive notifications: %v", err)
	}
	if inactiveCount != 0 {
		t.Fatalf("expected inactive user to receive no notifications, got %d", inactiveCount)
	}
}

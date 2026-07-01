/*
 * Chat service tests.
 * 1. Validate direct listing chats between different members.
 * 2. Keep message delivery and legacy notice-chat blocking covered by service-level tests.
 */
package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"ajoliving_web/http_service/internal/model"
	"ajoliving_web/http_service/internal/utils"
)

// 1. TestDirectListingChatAllowsDifferentUsersToExchangeMessages validates two member messaging.
func TestDirectListingChatAllowsDifferentUsersToExchangeMessages(t *testing.T) {
	ctx := context.Background()
	runtimeValue := newAuthTestRuntime(
		t,
		nil,
		&model.User{},
		&model.UserProfile{},
		&model.Listing{},
		&model.ListingContact{},
		&model.SecondhandListing{},
		&model.Chat{},
		&model.ChatParticipant{},
		&model.Message{},
		&model.Notification{},
	)
	runtimeValue.Now = func() time.Time {
		return time.Date(2026, 6, 29, 11, 0, 0, 0, time.UTC)
	}

	publisher := createChatTestUser(t, runtimeValue, "+852", "61231001", "餐桌發布者")
	inquirer := createChatTestUser(t, runtimeValue, "+852", "61231002", "查詢會員")
	listingPublicID := createChatTestSecondhandListing(t, runtimeValue, publisher.ID)
	chatService := NewChatService(runtimeValue, NewSecondhandService(runtimeValue), nil)

	result, err := chatService.CreateOrReuseChat(ctx, inquirer.ID, nil, listingPublicID)
	if err != nil {
		t.Fatalf("create direct listing chat: %v", err)
	}
	chatID, ok := result["chat_id"].(string)
	if !ok || chatID == "" {
		t.Fatalf("expected chat public id, got %#v", result)
	}

	var chat model.Chat
	if err := runtimeValue.DB.Where("public_id = ?", chatID).First(&chat).Error; err != nil {
		t.Fatalf("load created chat: %v", err)
	}
	assertChatParticipantRole(t, runtimeValue, chat.ID, inquirer.ID, "inquirer")
	assertChatParticipantRole(t, runtimeValue, chat.ID, publisher.ID, "publisher")

	firstMessage, err := chatService.SendMessage(ctx, inquirer.ID, chatID, "請問餐桌還在嗎？")
	if err != nil {
		t.Fatalf("send inquirer message: %v", err)
	}
	if firstMessage.SenderUserID != fmtInt64(inquirer.ID) || firstMessage.Content != "請問餐桌還在嗎？" {
		t.Fatalf("unexpected first message payload: %#v", firstMessage)
	}
	assertChatUnreadCount(t, runtimeValue, chat.ID, inquirer.ID, 0)
	assertChatUnreadCount(t, runtimeValue, chat.ID, publisher.ID, 1)
	assertChatNotification(t, runtimeValue, publisher.ID, chatID, "請問餐桌還在嗎？")

	messagesForPublisher, _, err := chatService.ListMessages(ctx, publisher.ID, chatID, 1, 20)
	if err != nil {
		t.Fatalf("list messages as publisher: %v", err)
	}
	if len(messagesForPublisher) != 1 || messagesForPublisher[0].Content != "請問餐桌還在嗎？" {
		t.Fatalf("expected publisher to read inquirer message, got %#v", messagesForPublisher)
	}

	reply, err := chatService.SendMessage(ctx, publisher.ID, chatID, "還在，可以約時間。")
	if err != nil {
		t.Fatalf("send publisher reply: %v", err)
	}
	if reply.SenderUserID != fmtInt64(publisher.ID) || reply.Content != "還在，可以約時間。" {
		t.Fatalf("unexpected reply payload: %#v", reply)
	}
	assertChatUnreadCount(t, runtimeValue, chat.ID, publisher.ID, 0)
	assertChatUnreadCount(t, runtimeValue, chat.ID, inquirer.ID, 1)
	assertChatNotification(t, runtimeValue, inquirer.ID, chatID, "還在，可以約時間。")

	messagesForInquirer, _, err := chatService.ListMessages(ctx, inquirer.ID, chatID, 1, 20)
	if err != nil {
		t.Fatalf("list messages as inquirer: %v", err)
	}
	if len(messagesForInquirer) != 2 {
		t.Fatalf("expected two exchanged messages, got %#v", messagesForInquirer)
	}
}

// 2. TestSendMessageRejectsTooLongContent validates message length protection.
func TestSendMessageRejectsTooLongContent(t *testing.T) {
	ctx := context.Background()
	runtimeValue := newAuthTestRuntime(
		t,
		nil,
		&model.User{},
		&model.UserProfile{},
		&model.Listing{},
		&model.ListingContact{},
		&model.SecondhandListing{},
		&model.Chat{},
		&model.ChatParticipant{},
		&model.Message{},
	)
	runtimeValue.Now = func() time.Time {
		return time.Date(2026, 6, 29, 12, 0, 0, 0, time.UTC)
	}

	publisher := createChatTestUser(t, runtimeValue, "+852", "61232001", "餐桌發布者")
	inquirer := createChatTestUser(t, runtimeValue, "+852", "61232002", "查詢會員")
	listingPublicID := createChatTestSecondhandListing(t, runtimeValue, publisher.ID)
	chatService := NewChatService(runtimeValue, NewSecondhandService(runtimeValue), nil)

	result, err := chatService.CreateOrReuseChat(ctx, inquirer.ID, nil, listingPublicID)
	if err != nil {
		t.Fatalf("create direct listing chat: %v", err)
	}
	chatID, ok := result["chat_id"].(string)
	if !ok || chatID == "" {
		t.Fatalf("expected chat public id, got %#v", result)
	}

	if _, err := chatService.SendMessage(ctx, inquirer.ID, chatID, strings.Repeat("長", messageContentMaxRunes+1)); err == nil {
		t.Fatal("expected too long message to be rejected")
	}

	var count int64
	if err := runtimeValue.DB.Model(&model.Message{}).Count(&count).Error; err != nil {
		t.Fatalf("count messages: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected no message to be stored, got %d", count)
	}
}

// 3. TestSystemNoticeChatsAreHiddenFromMessageManagement validates legacy notice chats stay out of chat APIs.
func TestSystemNoticeChatsAreHiddenFromMessageManagement(t *testing.T) {
	ctx := context.Background()
	runtimeValue := newAuthTestRuntime(
		t,
		nil,
		&model.User{},
		&model.UserProfile{},
		&model.Listing{},
		&model.SecondhandListing{},
		&model.Chat{},
		&model.ChatParticipant{},
		&model.Message{},
	)
	now := time.Date(2026, 6, 29, 13, 0, 0, 0, time.UTC)
	runtimeValue.Now = func() time.Time {
		return now
	}

	member := createNoticeTestUser(t, runtimeValue, "+852", "61233001", "active")
	legacyChat := model.Chat{
		PublicID:           utils.NewPublicID(),
		BizModule:          "system",
		ListingID:          0,
		ChatType:           chatTypeSystemNotice,
		CreatedBy:          member.ID,
		LastMessagePreview: "系統通知",
		LastMessageAt:      &now,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	if err := runtimeValue.DB.Create(&legacyChat).Error; err != nil {
		t.Fatalf("create legacy notice chat: %v", err)
	}
	participant := model.ChatParticipant{
		ChatID:      legacyChat.ID,
		UserID:      member.ID,
		RoleInChat:  "recipient",
		UnreadCount: 1,
		JoinedAt:    now,
	}
	if err := runtimeValue.DB.Create(&participant).Error; err != nil {
		t.Fatalf("create legacy notice participant: %v", err)
	}

	chatService := NewChatService(runtimeValue, nil, nil)
	items, pagination, err := chatService.ListChats(ctx, member.ID, 1, 20)
	if err != nil {
		t.Fatalf("list chats: %v", err)
	}
	if len(items) != 0 || pagination.Total != 0 {
		t.Fatalf("expected legacy system notice chat to be hidden, got items=%#v pagination=%#v", items, pagination)
	}
	if _, err := chatService.GetChat(ctx, member.ID, legacyChat.PublicID); err == nil {
		t.Fatal("expected legacy system notice chat detail to be blocked")
	}
	if _, _, err := chatService.ListMessages(ctx, member.ID, legacyChat.PublicID, 1, 20); err == nil {
		t.Fatal("expected legacy system notice messages to be blocked")
	}
	if err := chatService.MarkRead(ctx, member.ID, legacyChat.PublicID); err == nil {
		t.Fatal("expected legacy system notice read state to be blocked")
	}
}

// 4. createNoticeTestUser inserts a member with the target status.
func createNoticeTestUser(t *testing.T, runtimeValue *Runtime, phoneCountryCode string, phoneNumber string, status string) model.User {
	t.Helper()
	user := model.User{
		PublicID:         utils.NewPublicID(),
		PhoneCountryCode: phoneCountryCode,
		PhoneNumber:      phoneNumber,
		MemberStatus:     status,
		MemberType:       "user",
		IsVerifiedPhone:  true,
	}
	if err := runtimeValue.DB.Create(&user).Error; err != nil {
		t.Fatalf("create notice test user: %v", err)
	}

	return user
}

// 5. createChatTestUser inserts an active member and profile for direct chat tests.
func createChatTestUser(t *testing.T, runtimeValue *Runtime, phoneCountryCode string, phoneNumber string, displayName string) model.User {
	t.Helper()
	user := createNoticeTestUser(t, runtimeValue, phoneCountryCode, phoneNumber, "active")
	if err := runtimeValue.DB.Create(&model.UserProfile{
		UserID:                user.ID,
		DisplayName:           displayName,
		PublisherIdentityType: "owner",
		DistrictCode:          "eastern",
	}).Error; err != nil {
		t.Fatalf("create chat test user profile: %v", err)
	}

	return user
}

// 6. createChatTestSecondhandListing inserts a chat-enabled public listing.
func createChatTestSecondhandListing(t *testing.T, runtimeValue *Runtime, ownerUserID int64) string {
	t.Helper()
	publishedAt := runtimeValue.Now()
	listing := model.Listing{
		PublicID:              utils.NewPublicID(),
		Module:                "secondhand",
		OwnerUserID:           ownerUserID,
		Title:                 "實木餐桌",
		Summary:               "保養良好，可約時間交收。",
		Description:           "餐桌狀態良好，適合家庭使用。",
		DistrictCode:          "eastern",
		PublisherIdentityType: "owner",
		PublicationStatus:     "active",
		ModerationStatus:      "approved",
		BusinessStatus:        "available",
		PublishedAt:           &publishedAt,
		IsDeleted:             false,
	}
	if err := runtimeValue.DB.Create(&listing).Error; err != nil {
		t.Fatalf("create chat test listing: %v", err)
	}

	deliveryTags, err := marshalJSON([]string{"self_pickup"})
	if err != nil {
		t.Fatalf("marshal chat test delivery tags: %v", err)
	}
	price := float64(800)
	if err := runtimeValue.DB.Create(&model.SecondhandListing{
		ListingID:          listing.ID,
		CategoryCode:       "home_furniture",
		PriceMode:          "fixed",
		PriceHKD:           &price,
		ConditionLevel:     "used_good",
		PickupRegionCode:   "eastern",
		PickupLocationText: "屋苑大堂",
		DeliveryTags:       deliveryTags,
		VisibilityScope:    "public",
		ContactMethod:      "chat",
	}).Error; err != nil {
		t.Fatalf("create chat test secondhand listing: %v", err)
	}

	if err := runtimeValue.DB.Create(&model.ListingContact{
		ListingID:   listing.ID,
		ShowChat:    true,
		ContactMode: "chat",
	}).Error; err != nil {
		t.Fatalf("create chat test contact: %v", err)
	}

	return listing.PublicID
}

// 7. assertChatParticipantRole validates one chat participant role.
func assertChatParticipantRole(t *testing.T, runtimeValue *Runtime, chatID int64, userID int64, expectedRole string) {
	t.Helper()
	var participant model.ChatParticipant
	if err := runtimeValue.DB.Where("chat_id = ? AND user_id = ?", chatID, userID).First(&participant).Error; err != nil {
		t.Fatalf("load chat participant: %v", err)
	}
	if participant.RoleInChat != expectedRole {
		t.Fatalf("expected participant role %s, got %s", expectedRole, participant.RoleInChat)
	}
}

// 8. assertChatUnreadCount validates the participant unread count.
func assertChatUnreadCount(t *testing.T, runtimeValue *Runtime, chatID int64, userID int64, expectedCount int) {
	t.Helper()
	var participant model.ChatParticipant
	if err := runtimeValue.DB.Where("chat_id = ? AND user_id = ?", chatID, userID).First(&participant).Error; err != nil {
		t.Fatalf("load chat unread participant: %v", err)
	}
	if participant.UnreadCount != expectedCount {
		t.Fatalf("expected unread count %d for user %d, got %d", expectedCount, userID, participant.UnreadCount)
	}
}

// 9. assertChatNotification validates one chat notification was delivered.
func assertChatNotification(t *testing.T, runtimeValue *Runtime, userID int64, chatPublicID string, expectedBody string) {
	t.Helper()
	var notification model.Notification
	if err := runtimeValue.DB.
		Where("user_id = ? AND related_type = ? AND related_public_id = ? AND body = ?", userID, "chat", chatPublicID, expectedBody).
		First(&notification).
		Error; err != nil {
		t.Fatalf("load chat notification: %v", err)
	}
	if notification.Category != "chat_message" || notification.IsRead {
		t.Fatalf("unexpected chat notification: %+v", notification)
	}
}

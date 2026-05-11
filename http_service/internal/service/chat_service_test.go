/*
 * Chat service tests.
 * 1. Create chats, send messages, and validate unread counts.
 * 2. Marking a chat as read should clear unread state.
 * 3. Validate seeded admin accounts can exchange listing-scoped messages.
 */
package service

import (
	"context"
	"testing"

	"ajoliving_web/http_service/internal/database"
	"ajoliving_web/http_service/internal/model"
)

// 1. TestChatSendAndRead validates chat creation, messaging, and read flow.
func TestChatSendAndRead(t *testing.T) {
	runtime := newTestRuntime(t)
	secondhandService := NewSecondhandService(runtime)
	chatService := NewChatService(runtime, secondhandService)
	communityA, _ := mustGetCommunities(t, runtime)
	owner := mustCreateUser(t, runtime, "+852", "90000101", &communityA.ID)
	buyer := mustCreateUser(t, runtime, "+852", "90000102", &communityA.ID)
	listingID := mustCreatePublishedListing(t, runtime, owner, "public", communityA.PublicID)

	chatCreate, err := chatService.CreateOrReuseChat(context.Background(), buyer.ID, viewerProfileCommunityID(&buyer, runtime), listingID)
	if err != nil {
		t.Fatalf("create or reuse chat: %v", err)
	}

	chatID, _ := chatCreate["chat_id"].(string)
	if chatID == "" {
		t.Fatalf("expected chat id")
	}

	message, err := chatService.SendMessage(context.Background(), buyer.ID, chatID, "Hello, is this still available?")
	if err != nil {
		t.Fatalf("send message: %v", err)
	}
	if message.MessageID == "" {
		t.Fatalf("expected message id")
	}

	notificationService := NewNotificationService(runtime)
	ownerNotifications, _, ownerUnreadCount, err := notificationService.ListNotifications(context.Background(), owner.ID, NotificationListFilters{
		Page:     1,
		PageSize: 20,
	})
	if err != nil {
		t.Fatalf("list owner notifications: %v", err)
	}
	if ownerUnreadCount != 1 || len(ownerNotifications) != 1 || ownerNotifications[0].RelatedPublicID != chatID {
		t.Fatalf("expected one unread chat notification for owner, got %+v unread=%d", ownerNotifications, ownerUnreadCount)
	}
	buyerUnreadCount, err := notificationService.UnreadCount(context.Background(), buyer.ID)
	if err != nil {
		t.Fatalf("count buyer notifications: %v", err)
	}
	if buyerUnreadCount != 0 {
		t.Fatalf("expected no notification for message sender, got %d", buyerUnreadCount)
	}

	chats, _, err := chatService.ListChats(context.Background(), owner.ID, 1, 20)
	if err != nil {
		t.Fatalf("list chats: %v", err)
	}
	ownerChat := findChatSummaryByID(chats, chatID)
	if ownerChat == nil || ownerChat.UnreadCount != 1 {
		t.Fatalf("expected one unread message for the owner")
	}
	if ownerChat.Peer == nil || ownerChat.Peer.PublicID != buyer.PublicID || ownerChat.Peer.DisplayName == "" {
		t.Fatalf("expected peer display info in chat list response")
	}
	if ownerChat.Listing == nil || ownerChat.Listing.ListingID != listingID || ownerChat.Listing.CoverImage == nil {
		t.Fatalf("expected listing summary in chat list response")
	}

	detail, err := chatService.GetChat(context.Background(), owner.ID, chatID)
	if err != nil {
		t.Fatalf("get chat: %v", err)
	}
	if detail.Peer == nil || detail.Peer.PublicID != buyer.PublicID || detail.Peer.DisplayName == "" {
		t.Fatalf("expected peer display info in chat detail response")
	}
	if detail.Listing == nil || detail.Listing.ListingID != listingID || detail.Listing.Title == "" {
		t.Fatalf("expected listing summary in chat detail response")
	}
	if len(detail.Participants) != 2 || detail.Participants[0].DisplayName == "" || detail.Participants[0].PublicID == "" {
		t.Fatalf("expected participant display info in chat detail response")
	}

	if err := chatService.MarkRead(context.Background(), owner.ID, chatID); err != nil {
		t.Fatalf("mark read: %v", err)
	}

	chats, _, err = chatService.ListChats(context.Background(), owner.ID, 1, 20)
	if err != nil {
		t.Fatalf("list chats after read: %v", err)
	}
	ownerChat = findChatSummaryByID(chats, chatID)
	if ownerChat == nil || ownerChat.UnreadCount != 0 {
		t.Fatalf("expected unread count to be cleared after mark read")
	}

	if err := notificationService.MarkRead(context.Background(), owner.ID, ownerNotifications[0].NotificationID); err != nil {
		t.Fatalf("mark owner notification read: %v", err)
	}
	ownerUnreadCount, err = notificationService.UnreadCount(context.Background(), owner.ID)
	if err != nil {
		t.Fatalf("count owner notifications after mark read: %v", err)
	}
	if ownerUnreadCount != 0 {
		t.Fatalf("expected owner notification unread count cleared, got %d", ownerUnreadCount)
	}

	if _, err := chatService.SendMessage(context.Background(), owner.ID, chatID, "Yes, still available."); err != nil {
		t.Fatalf("send owner reply: %v", err)
	}
	buyerNotifications, _, buyerUnreadCount, err := notificationService.ListNotifications(context.Background(), buyer.ID, NotificationListFilters{
		Page:     1,
		PageSize: 20,
	})
	if err != nil {
		t.Fatalf("list buyer notifications: %v", err)
	}
	if buyerUnreadCount != 1 || len(buyerNotifications) != 1 || buyerNotifications[0].RelatedPublicID != chatID {
		t.Fatalf("expected one unread chat notification for buyer, got %+v unread=%d", buyerNotifications, buyerUnreadCount)
	}
	ownerUnreadCount, err = notificationService.UnreadCount(context.Background(), owner.ID)
	if err != nil {
		t.Fatalf("count owner notifications after reply: %v", err)
	}
	if ownerUnreadCount != 0 {
		t.Fatalf("expected no owner unread self notification after reply, got %d", ownerUnreadCount)
	}

	updated, err := notificationService.MarkAllRead(context.Background(), buyer.ID)
	if err != nil {
		t.Fatalf("mark all buyer notifications read: %v", err)
	}
	if updated != 1 {
		t.Fatalf("expected one buyer notification marked read, got %d", updated)
	}
}

// 2. TestSystemNoticeChatIsCreatedAndReadOnly validates the built-in system notice chat.
func TestSystemNoticeChatIsCreatedAndReadOnly(t *testing.T) {
	runtime := newTestRuntime(t)
	secondhandService := NewSecondhandService(runtime)
	chatService := NewChatService(runtime, secondhandService)
	communityA, _ := mustGetCommunities(t, runtime)
	user := mustCreateUser(t, runtime, "+852", "90000111", &communityA.ID)

	chats, _, err := chatService.ListChats(context.Background(), user.ID, 1, 20)
	if err != nil {
		t.Fatalf("list chats: %v", err)
	}
	if len(chats) != 1 || chats[0].ChatType != chatTypeSystemNotice {
		t.Fatalf("expected system notice chat, got %+v", chats)
	}
	if chats[0].Peer == nil || chats[0].Peer.DisplayName != model.SystemNotificationDisplayName {
		t.Fatalf("expected system notice peer, got %+v", chats[0].Peer)
	}
	if chats[0].UnreadCount != len(defaultSystemNoticeMessages(runtime.Now())) {
		t.Fatalf("expected default system unread count, got %d", chats[0].UnreadCount)
	}

	detail, err := chatService.GetChat(context.Background(), user.ID, chats[0].ChatID)
	if err != nil {
		t.Fatalf("get system notice chat: %v", err)
	}
	if detail.ChatType != chatTypeSystemNotice || detail.Listing != nil || detail.ListingID != "" {
		t.Fatalf("expected system notice detail without listing, got %+v", detail)
	}

	messages, _, err := chatService.ListMessages(context.Background(), user.ID, chats[0].ChatID, 1, 20)
	if err != nil {
		t.Fatalf("list system notice messages: %v", err)
	}
	if len(messages) != len(defaultSystemNoticeMessages(runtime.Now())) || messages[0].MessageType != messageTypeNoticeCard {
		t.Fatalf("expected notice card messages, got %+v", messages)
	}
	if messages[0].ActionLabel == "" || messages[0].ActionURL == "" {
		t.Fatalf("expected notice card action, got %+v", messages[0])
	}

	if _, err := chatService.SendMessage(context.Background(), user.ID, chats[0].ChatID, "hello"); err == nil {
		t.Fatalf("expected system notice chat to reject user message")
	}

	if err := chatService.MarkRead(context.Background(), user.ID, chats[0].ChatID); err != nil {
		t.Fatalf("mark system notice read: %v", err)
	}
	chats, _, err = chatService.ListChats(context.Background(), user.ID, 1, 20)
	if err != nil {
		t.Fatalf("list chats after read: %v", err)
	}
	if chats[0].UnreadCount != 0 {
		t.Fatalf("expected system notice unread count cleared, got %d", chats[0].UnreadCount)
	}

	owner := mustCreateUser(t, runtime, "+852", "90000112", &communityA.ID)
	listingID := mustCreatePublishedListing(t, runtime, owner, "public", communityA.PublicID)
	if _, err := chatService.CreateOrReuseChat(context.Background(), user.ID, viewerProfileCommunityID(&user, runtime), listingID); err != nil {
		t.Fatalf("create direct chat: %v", err)
	}

	chats, _, err = chatService.ListChats(context.Background(), user.ID, 1, 20)
	if err != nil {
		t.Fatalf("list chats with direct chat: %v", err)
	}
	if len(chats) < 2 || chats[0].ChatType != chatTypeSystemNotice {
		t.Fatalf("expected system notice chat to stay first, got %+v", chats)
	}

	result, err := chatService.PublishSystemNotice(context.Background(), SystemNoticePublishParams{
		Title:       "週末市集活動",
		Body:        "本週六開放二手攤位報名。",
		ActionLabel: "立即查看",
		ActionURL:   "/marketplace/discover",
	})
	if err != nil {
		t.Fatalf("publish system notice: %v", err)
	}
	if result.DeliveredCount != 2 {
		t.Fatalf("expected two active users to receive notice, got %d", result.DeliveredCount)
	}

	chats, _, err = chatService.ListChats(context.Background(), user.ID, 1, 20)
	if err != nil {
		t.Fatalf("list chats after notice publish: %v", err)
	}
	if len(chats) < 2 || chats[0].ChatType != chatTypeSystemNotice || chats[0].UnreadCount != 1 {
		t.Fatalf("expected user system notice unread count to increase, got %+v", chats)
	}

	messages, _, err = chatService.ListMessages(context.Background(), user.ID, chats[0].ChatID, 1, 20)
	if err != nil {
		t.Fatalf("list system notice messages after publish: %v", err)
	}
	latestMessage := messages[len(messages)-1]
	if latestMessage.MessageType != messageTypeNoticeCard || latestMessage.ActionLabel != "立即查看" || latestMessage.ActionURL == "" {
		t.Fatalf("expected latest notice card action, got %+v", latestMessage)
	}
}

// 3. TestDefaultAdminAccountsCanExchangeListingMessages validates admin-to-admin chat flow.
func TestDefaultAdminAccountsCanExchangeListingMessages(t *testing.T) {
	runtime := newTestRuntime(t)
	if err := database.SeedDefaultAdminAccount(context.Background(), runtime.DB); err != nil {
		t.Fatalf("seed default admin accounts: %v", err)
	}

	secondhandService := NewSecondhandService(runtime)
	chatService := NewChatService(runtime, secondhandService)
	communityA, _ := mustGetCommunities(t, runtime)
	adminCom := mustLoadSeedAdminUser(t, runtime, "admin@admin.com")
	adminCn := mustLoadSeedAdminUser(t, runtime, "admin@admin.cn")
	if adminCom.ID == adminCn.ID || adminCom.PublicID == adminCn.PublicID {
		t.Fatalf("expected distinct seeded admin users, got %+v and %+v", adminCom, adminCn)
	}

	listingID := mustCreatePublishedListing(t, runtime, adminCom, "public", communityA.PublicID)
	chatCreate, err := chatService.CreateOrReuseChat(context.Background(), adminCn.ID, viewerProfileCommunityID(&adminCn, runtime), listingID)
	if err != nil {
		t.Fatalf("create admin chat: %v", err)
	}
	chatID, _ := chatCreate["chat_id"].(string)
	if chatID == "" {
		t.Fatalf("expected admin chat id")
	}

	firstMessage, err := chatService.SendMessage(context.Background(), adminCn.ID, chatID, "你好，這件商品還在嗎？")
	if err != nil {
		t.Fatalf("send admin cn message: %v", err)
	}
	if firstMessage.SenderUserID != fmtInt64(adminCn.ID) {
		t.Fatalf("expected first sender admin cn, got %+v", firstMessage)
	}

	ownerChats, _, err := chatService.ListChats(context.Background(), adminCom.ID, 1, 20)
	if err != nil {
		t.Fatalf("list admin com chats: %v", err)
	}
	ownerChat := findChatSummaryByID(ownerChats, chatID)
	if ownerChat == nil || ownerChat.UnreadCount != 1 || ownerChat.Peer == nil || ownerChat.Peer.PublicID != adminCn.PublicID {
		t.Fatalf("expected admin com to see unread chat from admin cn, got %+v", ownerChat)
	}

	reply, err := chatService.SendMessage(context.Background(), adminCom.ID, chatID, "還在，可以安排交收。")
	if err != nil {
		t.Fatalf("send admin com reply: %v", err)
	}
	if reply.SenderUserID != fmtInt64(adminCom.ID) {
		t.Fatalf("expected reply sender admin com, got %+v", reply)
	}

	messagesForCn, _, err := chatService.ListMessages(context.Background(), adminCn.ID, chatID, 1, 20)
	if err != nil {
		t.Fatalf("list admin cn messages: %v", err)
	}
	if len(messagesForCn) != 2 || messagesForCn[0].SenderUserID != fmtInt64(adminCn.ID) || messagesForCn[1].SenderUserID != fmtInt64(adminCom.ID) {
		t.Fatalf("expected two admin messages in order, got %+v", messagesForCn)
	}

	buyerChats, _, err := chatService.ListChats(context.Background(), adminCn.ID, 1, 20)
	if err != nil {
		t.Fatalf("list admin cn chats: %v", err)
	}
	buyerChat := findChatSummaryByID(buyerChats, chatID)
	if buyerChat == nil || buyerChat.UnreadCount != 1 || buyerChat.Peer == nil || buyerChat.Peer.PublicID != adminCom.PublicID {
		t.Fatalf("expected admin cn to see unread reply from admin com, got %+v", buyerChat)
	}
}

// 4. mustLoadSeedAdminUser returns a seeded admin user by email.
func mustLoadSeedAdminUser(t *testing.T, runtime *Runtime, email string) model.User {
	t.Helper()

	var credential model.UserCredential
	if err := runtime.DB.Where("email = ?", email).First(&credential).Error; err != nil {
		t.Fatalf("load admin credential %s: %v", email, err)
	}

	var user model.User
	if err := runtime.DB.First(&user, credential.UserID).Error; err != nil {
		t.Fatalf("load admin user %s: %v", email, err)
	}

	return user
}

// 5. findChatSummaryByID returns a chat summary by public ID.
func findChatSummaryByID(items []ChatSummary, chatID string) *ChatSummary {
	for i := range items {
		if items[i].ChatID == chatID {
			return &items[i]
		}
	}

	return nil
}

/*
 * Chat service tests.
 * 1. Create chats, send messages, and validate unread counts.
 * 2. Marking a chat as read should clear unread state.
 */
package service

import (
	"context"
	"testing"
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

	chats, _, err := chatService.ListChats(context.Background(), owner.ID, 1, 20)
	if err != nil {
		t.Fatalf("list chats: %v", err)
	}
	if len(chats) != 1 || chats[0].UnreadCount != 1 {
		t.Fatalf("expected one unread message for the owner")
	}
	if chats[0].Peer == nil || chats[0].Peer.PublicID != buyer.PublicID || chats[0].Peer.DisplayName == "" {
		t.Fatalf("expected peer display info in chat list response")
	}
	if chats[0].Listing == nil || chats[0].Listing.ListingID != listingID || chats[0].Listing.CoverImage == nil {
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
	if len(chats) != 1 || chats[0].UnreadCount != 0 {
		t.Fatalf("expected unread count to be cleared after mark read")
	}
}

/*
 * Order and notification service tests.
 * 1. Verify order creation, confirmation, completion, and notification delivery.
 * 2. Verify confirmed order cancellation restores listing availability.
 */
package service

import (
	"context"
	"testing"

	"ajoliving_web/http_service/internal/model"
)

// 1. TestOrderLifecycleAndNotifications verifies the happy path order lifecycle.
func TestOrderLifecycleAndNotifications(t *testing.T) {
	runtime := newTestRuntime(t)
	notificationService := NewNotificationService(runtime)
	secondhandService := NewSecondhandService(runtime)
	orderService := NewOrderService(runtime, secondhandService, notificationService)
	communityA, _ := mustGetCommunities(t, runtime)
	owner := mustCreateUser(t, runtime, "+852", "91110001", &communityA.ID)
	buyer := mustCreateUser(t, runtime, "+852", "91110002", &communityA.ID)
	listingID := mustCreatePublishedListing(t, runtime, owner, "public", communityA.PublicID)

	created, err := orderService.CreateOrder(context.Background(), CreateOrderParams{
		BuyerUserID:      buyer.ID,
		BuyerCommunityID: viewerProfileCommunityID(&buyer, runtime),
		ListingPublicID:  listingID,
		BuyerNote:        "Can pick up tonight.",
		HandoverMethod:   "lobby_pickup",
	})
	if err != nil {
		t.Fatalf("create order: %v", err)
	}
	if created.OrderStatus != OrderStatusPendingConfirm {
		t.Fatalf("expected pending order, got %+v", created)
	}

	sellerNotifications, _, unreadCount, err := notificationService.ListNotifications(context.Background(), owner.ID, NotificationListFilters{
		Page: 1, PageSize: 20,
	})
	if err != nil {
		t.Fatalf("list seller notifications: %v", err)
	}
	if unreadCount < 1 || len(sellerNotifications) < 1 || sellerNotifications[0].Category != "order_created" {
		t.Fatalf("expected seller order_created notification, got %+v / unread=%d", sellerNotifications, unreadCount)
	}

	confirmed, err := orderService.ConfirmOrder(context.Background(), owner.ID, created.OrderID)
	if err != nil {
		t.Fatalf("confirm order: %v", err)
	}
	if confirmed.OrderStatus != OrderStatusConfirmed {
		t.Fatalf("expected confirmed order, got %+v", confirmed)
	}

	completed, err := orderService.CompleteOrder(context.Background(), buyer.ID, created.OrderID)
	if err != nil {
		t.Fatalf("complete order: %v", err)
	}
	if completed.OrderStatus != OrderStatusCompleted {
		t.Fatalf("expected completed order, got %+v", completed)
	}

	var listing model.Listing
	if err := runtime.DB.Where("public_id = ?", listingID).First(&listing).Error; err != nil {
		t.Fatalf("reload listing: %v", err)
	}
	if listing.BusinessStatus != "sold" {
		t.Fatalf("expected sold listing, got %s", listing.BusinessStatus)
	}

	buyerNotifications, _, buyerUnread, err := notificationService.ListNotifications(context.Background(), buyer.ID, NotificationListFilters{
		Page: 1, PageSize: 20,
	})
	if err != nil {
		t.Fatalf("list buyer notifications: %v", err)
	}
	if buyerUnread < 1 {
		t.Fatalf("expected buyer unread notifications, got %d", buyerUnread)
	}
	if len(buyerNotifications) == 0 || buyerNotifications[0].RelatedPublicID != created.OrderID {
		t.Fatalf("expected buyer notifications for order %s, got %+v", created.OrderID, buyerNotifications)
	}
}

// 2. TestCancelConfirmedOrderRestoresListing verifies the cancellation fallback path.
func TestCancelConfirmedOrderRestoresListing(t *testing.T) {
	runtime := newTestRuntime(t)
	notificationService := NewNotificationService(runtime)
	secondhandService := NewSecondhandService(runtime)
	orderService := NewOrderService(runtime, secondhandService, notificationService)
	communityA, _ := mustGetCommunities(t, runtime)
	owner := mustCreateUser(t, runtime, "+852", "91110011", &communityA.ID)
	buyer := mustCreateUser(t, runtime, "+852", "91110012", &communityA.ID)
	listingID := mustCreatePublishedListing(t, runtime, owner, "public", communityA.PublicID)

	created, err := orderService.CreateOrder(context.Background(), CreateOrderParams{
		BuyerUserID:      buyer.ID,
		BuyerCommunityID: viewerProfileCommunityID(&buyer, runtime),
		ListingPublicID:  listingID,
		HandoverMethod:   "face_to_face",
	})
	if err != nil {
		t.Fatalf("create order: %v", err)
	}
	if _, err := orderService.ConfirmOrder(context.Background(), owner.ID, created.OrderID); err != nil {
		t.Fatalf("confirm order: %v", err)
	}

	cancelled, err := orderService.CancelOrder(context.Background(), CancelOrderParams{
		UserID: buyer.ID,
		Reason: "Need to reschedule.",
	}, created.OrderID)
	if err != nil {
		t.Fatalf("cancel order: %v", err)
	}
	if cancelled.OrderStatus != OrderStatusCancelled {
		t.Fatalf("expected cancelled order, got %+v", cancelled)
	}

	var listing model.Listing
	if err := runtime.DB.Where("public_id = ?", listingID).First(&listing).Error; err != nil {
		t.Fatalf("reload listing: %v", err)
	}
	if listing.BusinessStatus != "available" {
		t.Fatalf("expected restored listing availability, got %s", listing.BusinessStatus)
	}
}

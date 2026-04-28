/*
 * Order service support helpers.
 * 1. Keep order access checks and summary builders separate from transitions.
 * 2. Format order timeline and peer payloads for the HTTP layer.
 */
package service

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/model"
)

// 1. ensureNoOpenBuyerOrder blocks duplicate open orders from the same buyer.
func (s *OrderService) ensureNoOpenBuyerOrder(ctx context.Context, listingID int64, buyerUserID int64) error {
	var count int64
	if err := s.runtime.DB.WithContext(ctx).Model(&model.Order{}).
		Where("listing_id = ? AND buyer_user_id = ? AND order_status IN ?", listingID, buyerUserID, []string{OrderStatusPendingConfirm, OrderStatusConfirmed}).
		Count(&count).Error; err != nil {
		return errcode.New(errcode.CodeInternalError, "failed to validate existing orders")
	}
	if count > 0 {
		return errcode.New(errcode.CodeValidationError, "an active order already exists for this listing")
	}

	return nil
}

// 2. loadAuthorizedOrder loads one order and validates participant access.
func (s *OrderService) loadAuthorizedOrder(ctx context.Context, userID int64, orderPublicID string) (*model.Order, *model.Listing, error) {
	var order model.Order
	if err := s.runtime.DB.WithContext(ctx).Where("public_id = ?", strings.TrimSpace(orderPublicID)).First(&order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, errcode.New(errcode.CodeNotFound, "order not found")
		}
		return nil, nil, errcode.New(errcode.CodeInternalError, "failed to load order")
	}
	if order.BuyerUserID != userID && order.SellerUserID != userID {
		return nil, nil, errcode.New(errcode.CodeAuthForbidden, "order is not accessible")
	}

	var listing model.Listing
	if err := s.runtime.DB.WithContext(ctx).First(&listing, order.ListingID).Error; err != nil {
		return nil, nil, errcode.New(errcode.CodeInternalError, "failed to load order listing")
	}

	return &order, &listing, nil
}

// 3. buildOrderSummaries maps a batch of orders into list payloads.
func (s *OrderService) buildOrderSummaries(ctx context.Context, orders []model.Order, viewerUserID int64) ([]OrderSummary, error) {
	items := make([]OrderSummary, 0, len(orders))
	for i := range orders {
		order := orders[i]
		var listing model.Listing
		if err := s.runtime.DB.WithContext(ctx).First(&listing, order.ListingID).Error; err != nil {
			return nil, errcode.New(errcode.CodeInternalError, "failed to load order listing")
		}
		summary, err := s.buildOrderSummary(ctx, &order, &listing, viewerUserID)
		if err != nil {
			return nil, err
		}
		items = append(items, *summary)
	}

	return items, nil
}

// 4. buildOrderSummary maps one order into a response payload.
func (s *OrderService) buildOrderSummary(ctx context.Context, order *model.Order, listing *model.Listing, viewerUserID int64) (*OrderSummary, error) {
	peer, err := s.loadOrderPeer(ctx, order, viewerUserID)
	if err != nil {
		return nil, err
	}

	images, err := s.secondhandService.loadListingImages(ctx, []int64{listing.ID})
	if err != nil {
		return nil, err
	}

	roleInOrder := "buyer"
	if order.SellerUserID == viewerUserID {
		roleInOrder = "seller"
	}

	return &OrderSummary{
		OrderID:        order.PublicID,
		ListingID:      listing.PublicID,
		ListingTitle:   listing.Title,
		OrderStatus:    order.OrderStatus,
		RoleInOrder:    roleInOrder,
		BuyerNote:      order.BuyerNote,
		HandoverMethod: order.HandoverMethod,
		CancelReason:   order.CancelReason,
		CreatedAt:      order.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:      order.UpdatedAt.UTC().Format(time.RFC3339),
		ConfirmedAt:    timePtrToRFC3339(order.ConfirmedAt),
		CompletedAt:    timePtrToRFC3339(order.CompletedAt),
		CancelledAt:    timePtrToRFC3339(order.CancelledAt),
		CoverImage:     firstListingImage(images[listing.ID]),
		Peer:           *peer,
	}, nil
}

// 5. loadOrderPeer resolves the opposite party display payload.
func (s *OrderService) loadOrderPeer(ctx context.Context, order *model.Order, viewerUserID int64) (*OrderPeer, error) {
	targetUserID := order.SellerUserID
	roleInOrder := "seller"
	if viewerUserID == order.SellerUserID {
		targetUserID = order.BuyerUserID
		roleInOrder = "buyer"
	}

	var user model.User
	if err := s.runtime.DB.WithContext(ctx).First(&user, targetUserID).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load order peer")
	}

	var profile model.UserProfile
	_ = s.runtime.DB.WithContext(ctx).Where("user_id = ?", targetUserID).First(&profile).Error
	displayName := strings.TrimSpace(profile.DisplayName)
	if displayName == "" {
		displayName = user.PublicID
	}

	return &OrderPeer{
		PublicID:    user.PublicID,
		DisplayName: displayName,
		RoleInOrder: roleInOrder,
	}, nil
}

// 6. createOrderLog persists one order event log row.
func (s *OrderService) createOrderLog(ctx context.Context, tx *gorm.DB, order *model.Order, actionType string, fromStatus string, toStatus string, operatorUserID int64, note string) error {
	record := model.OrderLog{
		OrderID:        order.ID,
		ActionType:     actionType,
		FromStatus:     fromStatus,
		ToStatus:       toStatus,
		OperatorUserID: operatorUserID,
		Note:           strings.TrimSpace(note),
		CreatedAt:      s.runtime.Now(),
	}
	if err := tx.WithContext(ctx).Create(&record).Error; err != nil {
		return errcode.New(errcode.CodeInternalError, "failed to create order log")
	}

	return nil
}

// 7. normalizeHandoverMethod normalizes the handover input.
func normalizeHandoverMethod(value string) string {
	switch strings.TrimSpace(value) {
	case "face_to_face", "lobby_pickup", "courier":
		return strings.TrimSpace(value)
	default:
		return "face_to_face"
	}
}

// 8. timePtrToRFC3339 formats one optional timestamp.
func timePtrToRFC3339(value *time.Time) *string {
	if value == nil {
		return nil
	}

	formatted := value.UTC().Format(time.RFC3339)
	return &formatted
}

// 9. firstListingImage returns the first image or nil.
func firstListingImage(images []ListingImageResponse) *ListingImageResponse {
	if len(images) == 0 {
		return nil
	}

	image := images[0]
	return &image
}

// 10. formatInt64 converts int64 to string.
func formatInt64(value int64) string {
	return strconv.FormatInt(value, 10)
}

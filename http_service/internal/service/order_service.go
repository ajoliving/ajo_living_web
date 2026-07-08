/*
 * Order business logic.
 * 1. Manage secondhand listing order creation and state transitions.
 * 2. Keep listing availability and order logs synchronized.
 * 3. Trigger member notifications for major order events.
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

const (
	// 1. OrderStatusPendingConfirm marks a newly created order.
	OrderStatusPendingConfirm = "pending_confirm"
	// 2. OrderStatusConfirmed marks a seller-confirmed order.
	OrderStatusConfirmed = "confirmed"
	// 3. OrderStatusCompleted marks a completed order.
	OrderStatusCompleted = "completed"
	// 4. OrderStatusCancelled marks a cancelled order.
	OrderStatusCancelled = "cancelled"
)

// 5. OrderService manages listing orders.
type OrderService struct {
	runtime             *Runtime
	secondhandService   *SecondhandService
	notificationService *NotificationService
}

// 6. OrderPeer defines the opposite party summary.
type OrderPeer struct {
	PublicID    string `json:"public_id"`
	DisplayName string `json:"display_name"`
	RoleInOrder string `json:"role_in_order"`
}

// 7. OrderEvent defines one order timeline event.
type OrderEvent struct {
	ActionType     string `json:"action_type"`
	FromStatus     string `json:"from_status"`
	ToStatus       string `json:"to_status"`
	OperatorUserID string `json:"operator_user_id"`
	Note           string `json:"note"`
	CreatedAt      string `json:"created_at"`
}

// 8. OrderSummary defines an order list item payload.
type OrderSummary struct {
	OrderID        string                `json:"order_id"`
	ListingID      string                `json:"listing_id"`
	ListingTitle   string                `json:"listing_title"`
	OrderStatus    string                `json:"order_status"`
	RoleInOrder    string                `json:"role_in_order"`
	BuyerNote      string                `json:"buyer_note"`
	HandoverMethod string                `json:"handover_method"`
	CancelReason   string                `json:"cancel_reason"`
	CreatedAt      string                `json:"created_at"`
	UpdatedAt      string                `json:"updated_at"`
	ConfirmedAt    *string               `json:"confirmed_at,omitempty"`
	CompletedAt    *string               `json:"completed_at,omitempty"`
	CancelledAt    *string               `json:"cancelled_at,omitempty"`
	CoverImage     *ListingImageResponse `json:"cover_image,omitempty"`
	Peer           OrderPeer             `json:"peer"`
}

// 9. OrderDetail defines a full order detail payload.
type OrderDetail struct {
	OrderSummary
	Logs []OrderEvent `json:"logs"`
}

// 10. OrderListFilters defines order list filters.
type OrderListFilters struct {
	Page     int
	PageSize int
	Status   string
	Role     string
}

// 11. CreateOrderParams defines order creation input.
type CreateOrderParams struct {
	BuyerUserID      int64
	BuyerCommunityID *int64
	ListingPublicID  string
	BuyerNote        string
	HandoverMethod   string
}

// 12. CancelOrderParams defines order cancellation input.
type CancelOrderParams struct {
	UserID int64
	Reason string
}

// 13. NewOrderService creates an order service instance.
func NewOrderService(runtime *Runtime, secondhandService *SecondhandService, notificationService *NotificationService) *OrderService {
	return &OrderService{
		runtime:             runtime,
		secondhandService:   secondhandService,
		notificationService: notificationService,
	}
}

// 14. CreateOrder creates a new buyer order for one listing.
func (s *OrderService) CreateOrder(ctx context.Context, params CreateOrderParams) (*OrderDetail, error) {
	listing, secondhand, _, err := s.secondhandService.loadListingByPublicID(ctx, strings.TrimSpace(params.ListingPublicID))
	if err != nil {
		return nil, err
	}
	if listing.OwnerUserID == params.BuyerUserID {
		return nil, errcode.New(errcode.CodeValidationError, "seller cannot create an order on the same listing")
	}
	if listing.PublicationStatus != "active" || listing.ModerationStatus != "approved" || listing.BusinessStatus != "available" {
		return nil, errcode.New(errcode.CodeValidationError, "listing is not available for ordering")
	}
	if !s.secondhandService.canViewListing(listing, secondhand, params.BuyerCommunityID) {
		return nil, errcode.New(errcode.CodeVisibilityForbidden, "listing is not visible to the current user")
	}
	if err := s.ensureNoOpenBuyerOrder(ctx, listing.ID, params.BuyerUserID); err != nil {
		return nil, err
	}

	order := model.Order{
		PublicID:       utils.NewPublicID(),
		BizModule:      "secondhand",
		ListingID:      listing.ID,
		BuyerUserID:    params.BuyerUserID,
		SellerUserID:   listing.OwnerUserID,
		OrderStatus:    OrderStatusPendingConfirm,
		BuyerNote:      strings.TrimSpace(params.BuyerNote),
		HandoverMethod: normalizeHandoverMethod(params.HandoverMethod),
	}

	if err := s.runtime.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&order).Error; err != nil {
			return errcode.New(errcode.CodeInternalError, "failed to create order")
		}
		if err := s.createOrderLog(ctx, tx, &order, "create_order", "", OrderStatusPendingConfirm, params.BuyerUserID, order.BuyerNote); err != nil {
			return err
		}
		return s.notificationService.CreateNotification(ctx, tx, CreateNotificationParams{
			UserID:          listing.OwnerUserID,
			Category:        "order_created",
			Title:           "New order request",
			Body:            "A buyer has created an order request for your listing.",
			RelatedType:     "order",
			RelatedPublicID: order.PublicID,
		})
	}); err != nil {
		return nil, err
	}

	return s.GetOrder(ctx, params.BuyerUserID, order.PublicID)
}

// 15. ListMyOrders returns the current user's buyer and seller orders.
func (s *OrderService) ListMyOrders(ctx context.Context, userID int64, filters OrderListFilters) ([]OrderSummary, *model.Pagination, error) {
	page, pageSize := normalizePagination(filters.Page, filters.PageSize)
	query := s.runtime.DB.WithContext(ctx).Model(&model.Order{})

	switch strings.TrimSpace(filters.Role) {
	case "buyer":
		query = query.Where("buyer_user_id = ?", userID)
	case "seller":
		query = query.Where("seller_user_id = ?", userID)
	default:
		query = query.Where("buyer_user_id = ? OR seller_user_id = ?", userID, userID)
	}

	if status := strings.TrimSpace(filters.Status); status != "" {
		query = query.Where("order_status = ?", status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, nil, errcode.New(errcode.CodeInternalError, "failed to count orders")
	}

	var orders []model.Order
	if err := query.Order("updated_at desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&orders).Error; err != nil {
		return nil, nil, errcode.New(errcode.CodeInternalError, "failed to load orders")
	}

	items, err := s.buildOrderSummaries(ctx, orders, userID)
	if err != nil {
		return nil, nil, err
	}

	return items, &model.Pagination{Page: page, PageSize: pageSize, Total: total}, nil
}

// 16. GetOrder returns one order detail payload for a participant.
func (s *OrderService) GetOrder(ctx context.Context, userID int64, orderPublicID string) (*OrderDetail, error) {
	order, listing, err := s.loadAuthorizedOrder(ctx, userID, orderPublicID)
	if err != nil {
		return nil, err
	}

	summary, err := s.buildOrderSummary(ctx, order, listing, userID)
	if err != nil {
		return nil, err
	}

	var logs []model.OrderLog
	if err := s.runtime.DB.WithContext(ctx).Where("order_id = ?", order.ID).Order("created_at asc").Find(&logs).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load order logs")
	}

	result := &OrderDetail{
		OrderSummary: *summary,
		Logs:         make([]OrderEvent, 0, len(logs)),
	}
	for _, item := range logs {
		result.Logs = append(result.Logs, OrderEvent{
			ActionType:     item.ActionType,
			FromStatus:     item.FromStatus,
			ToStatus:       item.ToStatus,
			OperatorUserID: formatInt64(item.OperatorUserID),
			Note:           item.Note,
			CreatedAt:      item.CreatedAt.UTC().Format(time.RFC3339),
		})
	}

	return result, nil
}

// 17. ConfirmOrder confirms a pending order as the listing seller.
func (s *OrderService) ConfirmOrder(ctx context.Context, userID int64, orderPublicID string) (*OrderDetail, error) {
	order, listing, err := s.loadAuthorizedOrder(ctx, userID, orderPublicID)
	if err != nil {
		return nil, err
	}
	if order.SellerUserID != userID {
		return nil, errcode.New(errcode.CodeAuthForbidden, "only the seller can confirm this order")
	}
	if order.OrderStatus != OrderStatusPendingConfirm {
		return nil, errcode.New(errcode.CodeValidationError, "only pending orders can be confirmed")
	}
	if listing.PublicationStatus != "active" || listing.ModerationStatus != "approved" || listing.BusinessStatus != "available" {
		return nil, errcode.New(errcode.CodeValidationError, "listing is not available for ordering")
	}

	now := s.runtime.Now()
	if err := s.runtime.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Order{}).Where("id = ?", order.ID).Updates(map[string]any{
			"order_status": OrderStatusConfirmed,
			"confirmed_at": now,
			"updated_at":   now,
		}).Error; err != nil {
			return errcode.New(errcode.CodeInternalError, "failed to confirm order")
		}
		if err := tx.Model(&model.Listing{}).Where("id = ?", listing.ID).Update("business_status", "reserved").Error; err != nil {
			return errcode.New(errcode.CodeInternalError, "failed to update listing status")
		}
		if err := s.createOrderLog(ctx, tx, order, "confirm_order", OrderStatusPendingConfirm, OrderStatusConfirmed, userID, ""); err != nil {
			return err
		}
		return s.notificationService.CreateNotification(ctx, tx, CreateNotificationParams{
			UserID:          order.BuyerUserID,
			Category:        "order_confirmed",
			Title:           "Order confirmed",
			Body:            "The seller has confirmed your order request.",
			RelatedType:     "order",
			RelatedPublicID: order.PublicID,
		})
	}); err != nil {
		return nil, err
	}

	return s.GetOrder(ctx, userID, order.PublicID)
}

// 18. CancelOrder cancels a pending or confirmed order.
func (s *OrderService) CancelOrder(ctx context.Context, params CancelOrderParams, orderPublicID string) (*OrderDetail, error) {
	order, listing, err := s.loadAuthorizedOrder(ctx, params.UserID, orderPublicID)
	if err != nil {
		return nil, err
	}
	if order.OrderStatus != OrderStatusPendingConfirm && order.OrderStatus != OrderStatusConfirmed {
		return nil, errcode.New(errcode.CodeValidationError, "only pending or confirmed orders can be cancelled")
	}

	now := s.runtime.Now()
	previousStatus := order.OrderStatus
	reason := strings.TrimSpace(params.Reason)
	if err := s.runtime.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Order{}).Where("id = ?", order.ID).Updates(map[string]any{
			"order_status":  OrderStatusCancelled,
			"cancel_reason": reason,
			"cancelled_at":  now,
			"updated_at":    now,
		}).Error; err != nil {
			return errcode.New(errcode.CodeInternalError, "failed to cancel order")
		}
		if previousStatus == OrderStatusConfirmed {
			if err := tx.Model(&model.Listing{}).Where("id = ? AND business_status = ?", listing.ID, "reserved").Update("business_status", "available").Error; err != nil {
				return errcode.New(errcode.CodeInternalError, "failed to restore listing status")
			}
		}
		if err := s.createOrderLog(ctx, tx, order, "cancel_order", previousStatus, OrderStatusCancelled, params.UserID, reason); err != nil {
			return err
		}

		notifyUserID := order.BuyerUserID
		if params.UserID == order.BuyerUserID {
			notifyUserID = order.SellerUserID
		}
		return s.notificationService.CreateNotification(ctx, tx, CreateNotificationParams{
			UserID:          notifyUserID,
			Category:        "order_cancelled",
			Title:           "Order cancelled",
			Body:            "The order has been cancelled.",
			RelatedType:     "order",
			RelatedPublicID: order.PublicID,
		})
	}); err != nil {
		return nil, err
	}

	return s.GetOrder(ctx, params.UserID, order.PublicID)
}

// 19. CompleteOrder completes a confirmed order.
func (s *OrderService) CompleteOrder(ctx context.Context, userID int64, orderPublicID string) (*OrderDetail, error) {
	order, listing, err := s.loadAuthorizedOrder(ctx, userID, orderPublicID)
	if err != nil {
		return nil, err
	}
	if order.OrderStatus != OrderStatusConfirmed {
		return nil, errcode.New(errcode.CodeValidationError, "only confirmed orders can be completed")
	}

	now := s.runtime.Now()
	if err := s.runtime.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Order{}).Where("id = ?", order.ID).Updates(map[string]any{
			"order_status": OrderStatusCompleted,
			"completed_at": now,
			"updated_at":   now,
		}).Error; err != nil {
			return errcode.New(errcode.CodeInternalError, "failed to complete order")
		}
		if err := tx.Model(&model.Listing{}).Where("id = ?", listing.ID).Update("business_status", "sold").Error; err != nil {
			return errcode.New(errcode.CodeInternalError, "failed to update listing status")
		}
		if err := s.createOrderLog(ctx, tx, order, "complete_order", OrderStatusConfirmed, OrderStatusCompleted, userID, ""); err != nil {
			return err
		}

		notifyUserID := order.BuyerUserID
		if userID == order.BuyerUserID {
			notifyUserID = order.SellerUserID
		}
		return s.notificationService.CreateNotification(ctx, tx, CreateNotificationParams{
			UserID:          notifyUserID,
			Category:        "order_completed",
			Title:           "Order completed",
			Body:            "The order has been marked as completed.",
			RelatedType:     "order",
			RelatedPublicID: order.PublicID,
		})
	}); err != nil {
		return nil, err
	}

	return s.GetOrder(ctx, userID, order.PublicID)
}

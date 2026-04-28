/*
 * Order HTTP handlers.
 * 1. Bind order creation and lifecycle payloads.
 * 2. Delegate order flows to the service layer.
 */
package handler

import (
	"strings"

	"github.com/gin-gonic/gin"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/service"
)

// 1. OrderHandler handles member order endpoints.
type OrderHandler struct {
	orderService *service.OrderService
}

// 2. createOrderRequest defines the order creation payload.
type createOrderRequest struct {
	BuyerNote      string `json:"buyer_note"`
	HandoverMethod string `json:"handover_method"`
}

// 3. cancelOrderRequest defines the order cancellation payload.
type cancelOrderRequest struct {
	Reason string `json:"reason"`
}

// 4. NewOrderHandler creates an order handler instance.
func NewOrderHandler(orderService *service.OrderService) *OrderHandler {
	return &OrderHandler{orderService: orderService}
}

// 5. Create creates a new listing order.
func (h *OrderHandler) Create(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	var request createOrderRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid request payload"))
		return
	}

	result, err := h.orderService.CreateOrder(c.Request.Context(), service.CreateOrderParams{
		BuyerUserID:      user.UserID,
		BuyerCommunityID: user.PrimaryCommunityID,
		ListingPublicID:  strings.TrimSpace(c.Param("listingId")),
		BuyerNote:        strings.TrimSpace(request.BuyerNote),
		HandoverMethod:   strings.TrimSpace(request.HandoverMethod),
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, result)
}

// 6. MyOrders returns the current user's orders.
func (h *OrderHandler) MyOrders(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	page, pageSize := parsePagination(c)
	items, pagination, err := h.orderService.ListMyOrders(c.Request.Context(), user.UserID, service.OrderListFilters{
		Page:     page,
		PageSize: pageSize,
		Status:   strings.TrimSpace(c.Query("status")),
		Role:     strings.TrimSpace(c.Query("role")),
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, gin.H{"items": items, "pagination": pagination})
}

// 7. GetDetail returns one order detail.
func (h *OrderHandler) GetDetail(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	result, err := h.orderService.GetOrder(c.Request.Context(), user.UserID, strings.TrimSpace(c.Param("orderId")))
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, result)
}

// 8. Confirm confirms one pending order.
func (h *OrderHandler) Confirm(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	result, err := h.orderService.ConfirmOrder(c.Request.Context(), user.UserID, strings.TrimSpace(c.Param("orderId")))
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, result)
}

// 9. Cancel cancels one order.
func (h *OrderHandler) Cancel(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	var request cancelOrderRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid request payload"))
		return
	}

	result, err := h.orderService.CancelOrder(c.Request.Context(), service.CancelOrderParams{
		UserID: user.UserID,
		Reason: strings.TrimSpace(request.Reason),
	}, strings.TrimSpace(c.Param("orderId")))
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, result)
}

// 10. Complete completes one confirmed order.
func (h *OrderHandler) Complete(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	result, err := h.orderService.CompleteOrder(c.Request.Context(), user.UserID, strings.TrimSpace(c.Param("orderId")))
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, result)
}

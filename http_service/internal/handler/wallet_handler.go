/*
 * Wallet HTTP handlers.
 * 1. Bind member wallet balance, transaction, rewarded ad, and ad metric endpoints.
 * 2. Delegate all AJO Point account and reward logic to the wallet service.
 * 3. Keep wallet responses aligned with the unified API envelope.
 */
package handler

import (
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/service"
)

// 1. WalletHandler handles member wallet endpoints.
type WalletHandler struct {
	walletService *service.WalletService
}

// 2. NewWalletHandler creates a wallet handler instance.
func NewWalletHandler(walletService *service.WalletService) *WalletHandler {
	return &WalletHandler{walletService: walletService}
}

// 3. GetWallet returns the current member wallet overview.
func (h *WalletHandler) GetWallet(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	result, err := h.walletService.GetWalletOverview(c.Request.Context(), user.UserID)
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 4. ListTransactions returns paged wallet transactions.
func (h *WalletHandler) ListTransactions(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	page, pageSize := parsePagination(c)
	items, pagination, err := h.walletService.ListTransactions(c.Request.Context(), user.UserID, page, pageSize)
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, gin.H{"items": items, "pagination": pagination})
}

// 5. ListAdTasks returns available rewarded ad tasks.
func (h *WalletHandler) ListAdTasks(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	items, err := h.walletService.ListRewardAdTasks(c.Request.Context(), user.UserID)
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, gin.H{"items": items})
}

// 6. ListPublicAds returns public display ads for listing pages.
func (h *WalletHandler) ListPublicAds(c *gin.Context) {
	items, err := h.walletService.ListPublicDisplayAds(c.Request.Context(), service.PublicDisplayAdFilters{
		Channel:   strings.TrimSpace(c.Query("channel")),
		Placement: strings.TrimSpace(c.Query("placement")),
		Limit:     parsePositiveIntQuery(c, "limit"),
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, gin.H{"items": items})
}

// 7. StartAdTask starts a rewarded ad watch session.
func (h *WalletHandler) StartAdTask(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	result, err := h.walletService.StartRewardAd(
		c.Request.Context(),
		user.UserID,
		strings.TrimSpace(c.Param("taskId")),
		c.ClientIP(),
		c.Request.UserAgent(),
	)
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 8. ClaimAdTask claims a completed rewarded ad watch session.
func (h *WalletHandler) ClaimAdTask(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	var request struct {
		ClaimID string `json:"claim_id"`
	}
	if err := c.ShouldBindJSON(&request); err != nil || strings.TrimSpace(request.ClaimID) == "" {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid reward claim payload"))
		return
	}

	result, err := h.walletService.ClaimRewardAd(
		c.Request.Context(),
		user.UserID,
		strings.TrimSpace(c.Param("taskId")),
		strings.TrimSpace(request.ClaimID),
		c.ClientIP(),
		c.Request.UserAgent(),
	)
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 9. TrackAdTaskClick records one rewarded ad target link click.
func (h *WalletHandler) TrackAdTaskClick(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	result, err := h.walletService.TrackRewardAdClick(
		c.Request.Context(),
		strings.TrimSpace(c.Param("taskId")),
	)
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 10. CreateRechargeOrder creates a member wallet recharge payment order.
func (h *WalletHandler) CreateRechargeOrder(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	var request struct {
		AmountHKD  float64 `json:"amount_hkd"`
		PayMethod  string  `json:"pay_method"`
		PayRegion  string  `json:"pay_region"`
		DeviceMode string  `json:"device_mode"`
		ReturnPath string  `json:"return_path"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid recharge payload"))
		return
	}

	result, err := h.walletService.CreateRechargeOrder(c.Request.Context(), service.WalletRechargeCreateParams{
		UserID:     user.UserID,
		AmountHKD:  request.AmountHKD,
		PayMethod:  request.PayMethod,
		PayRegion:  request.PayRegion,
		DeviceMode: request.DeviceMode,
		ReturnPath: request.ReturnPath,
		ClientIP:   c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 11. GetRechargeOrder returns one member recharge order status.
func (h *WalletHandler) GetRechargeOrder(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	result, err := h.walletService.GetRechargeOrder(
		c.Request.Context(),
		user.UserID,
		strings.TrimSpace(c.Param("orderId")),
		parseBoolQuery(c, "refresh"),
	)
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 12. HandlePaymentNotify receives EasyLink wallet recharge callbacks.
func (h *WalletHandler) HandlePaymentNotify(c *gin.Context) {
	rawBody, err := io.ReadAll(io.LimitReader(c.Request.Body, 1<<20))
	if err != nil {
		c.String(http.StatusBadRequest, "fail")
		return
	}

	result, handleErr := h.walletService.HandlePaymentNotify(
		c.Request.Context(),
		c.GetHeader("Content-Type"),
		string(rawBody),
	)
	if result == nil {
		result = &service.PaymentNotifyResult{StatusCode: http.StatusInternalServerError, Body: "fail"}
	}
	if handleErr != nil {
		c.String(result.StatusCode, result.Body)
		return
	}

	c.String(result.StatusCode, result.Body)
}

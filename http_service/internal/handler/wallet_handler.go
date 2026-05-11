/*
 * Wallet HTTP handlers.
 * 1. Bind member wallet balance, transaction, and rewarded ad endpoints.
 * 2. Delegate all AJO Point account and reward logic to the wallet service.
 * 3. Keep wallet responses aligned with the unified API envelope.
 */
package handler

import (
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

// 6. StartAdTask starts a rewarded ad watch session.
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

// 7. ClaimAdTask claims a completed rewarded ad watch session.
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

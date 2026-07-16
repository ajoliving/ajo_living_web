/*
 * POS 物業繳費 HTTP 介面。
 * 1. 提供會員 POS 繳費概覽、賬單、訂單與歷史查詢。
 * 2. 提供會員可見屋苑會計資料查詢入口。
 * 3. 所有請求均透過 AJO 登入態與 service 層權限校驗。
 */
package handler

import (
	"strings"

	"github.com/gin-gonic/gin"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/service"
)

// 1. POSPaymentHandler handles POS payment endpoints.
type POSPaymentHandler struct {
	posPaymentService *service.POSPaymentService
}

// 2. NewPOSPaymentHandler creates a POS payment handler instance.
func NewPOSPaymentHandler(posPaymentService *service.POSPaymentService) *POSPaymentHandler {
	return &POSPaymentHandler{posPaymentService: posPaymentService}
}

// 3. Overview returns member POS payment readiness.
func (h *POSPaymentHandler) Overview(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	includeSummary := true
	if summary := parseOptionalBoolQuery(c, "summary"); summary != nil {
		includeSummary = *summary
	}
	result, err := h.posPaymentService.Overview(c.Request.Context(), user.UserID, paymentSelectionFromRequest(c), includeSummary)
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 4. ListBills returns current unit POS bills.
func (h *POSPaymentHandler) ListBills(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	result, err := h.posPaymentService.ListBills(c.Request.Context(), user.UserID, paymentSelectionFromRequest(c))
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 5. ListOrders returns current unit H5 order records.
func (h *POSPaymentHandler) ListOrders(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	result, err := h.posPaymentService.ListOrderRecords(c.Request.Context(), user.UserID, paymentSelectionFromRequest(c))
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 6. CreateOrder creates one H5 POS payment order.
func (h *POSPaymentHandler) CreateOrder(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	var request struct {
		BuildingID              string           `json:"building_id"`
		UnitID                  string           `json:"unit_id"`
		Scene                   string           `json:"scene"`
		PayChannel              string           `json:"pay_channel"`
		ExpireSeconds           int64            `json:"expire_seconds"`
		FinalAmount             int64            `json:"final_amount"`
		HandleFeeAmount         int64            `json:"handle_fee_amount"`
		BillObjs                []map[string]any `json:"bill_objs"`
		HandleFeeObj            []map[string]any `json:"handle_fee_obj"`
		ReturnPath              string           `json:"return_path"`
		Remark                  string           `json:"remark"`
		GatewayRequestOverrides map[string]any   `json:"gateway_request_overrides"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid payment order payload"))
		return
	}

	result, err := h.posPaymentService.CreateH5Order(c.Request.Context(), service.POSPaymentOrderCreateParams{
		UserID: user.UserID,
		Selection: service.POSPaymentSelection{
			BuildingID: strings.TrimSpace(request.BuildingID),
			UnitID:     strings.TrimSpace(request.UnitID),
		},
		Scene:                   strings.TrimSpace(request.Scene),
		PayChannel:              strings.TrimSpace(request.PayChannel),
		ExpireSeconds:           request.ExpireSeconds,
		FinalAmount:             request.FinalAmount,
		HandleFeeAmount:         request.HandleFeeAmount,
		BillObjs:                request.BillObjs,
		HandleFeeObj:            request.HandleFeeObj,
		ReturnPath:              strings.TrimSpace(request.ReturnPath),
		Remark:                  strings.TrimSpace(request.Remark),
		GatewayRequestOverrides: request.GatewayRequestOverrides,
		ClientIP:                c.ClientIP(),
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 7. GetOrder returns one H5 order.
func (h *POSPaymentHandler) GetOrder(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	result, err := h.posPaymentService.GetH5Order(
		c.Request.Context(),
		user.UserID,
		strings.TrimSpace(c.Param("mchOrderNo")),
		parseBoolQuery(c, "detail"),
		parseBoolQuery(c, "refresh"),
		paymentSelectionFromRequest(c),
	)
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 8. ListHistory returns current unit POS transaction history.
func (h *POSPaymentHandler) ListHistory(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	result, err := h.posPaymentService.ListHistoryByQuery(c.Request.Context(), user.UserID, paymentHistoryQueryFromRequest(c))
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 9. ListAccounting returns visible POS accounting entries.
func (h *POSPaymentHandler) ListAccounting(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	result, err := h.posPaymentService.ListAccountingGroups(c.Request.Context(), user.UserID, paymentSelectionFromRequest(c))
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 10. paymentSelectionFromRequest reads the selected POS context.
func paymentSelectionFromRequest(c *gin.Context) service.POSPaymentSelection {
	return service.POSPaymentSelection{
		BuildingID: strings.TrimSpace(c.Query("building_id")),
		UnitID:     strings.TrimSpace(c.Query("unit_id")),
	}
}

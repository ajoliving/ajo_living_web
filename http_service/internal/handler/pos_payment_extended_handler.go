/*
 * POS 物業繳費擴展 HTTP 介面。
 * 1. 提供手續費、銀行賬戶、線下繳費與歷史詳情接口。
 * 2. 提供 H5 訂單查詢、關閉、取消與所屬屋苑清機接口。
 * 3. 保持所有舊 POS 操作經由 AJO 後端登入態與權限校驗。
 */
package handler

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/service"
)

// 1. ListFees returns POS payment method and handling fee settings.
func (h *POSPaymentHandler) ListFees(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	result, err := h.posPaymentService.ListFees(c.Request.Context(), user.UserID, paymentSelectionFromRequest(c))
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 2. ListBankAccounts returns POS bank accounts for the selected building.
func (h *POSPaymentHandler) ListBankAccounts(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	result, err := h.posPaymentService.ListBankAccounts(c.Request.Context(), user.UserID, paymentSelectionFromRequest(c))
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 3. ReportPayment submits an offline POS payment through the AJO proxy.
func (h *POSPaymentHandler) ReportPayment(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	var request map[string]any
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid payment report payload"))
		return
	}

	result, err := h.posPaymentService.ReportPayment(c.Request.Context(), service.POSPaymentReportParams{
		UserID:    user.UserID,
		Selection: paymentSelectionFromPayload(c, request),
		Payload:   request,
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 4. PayByTerminal submits one member POS terminal payment.
func (h *POSPaymentHandler) PayByTerminal(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	var request struct {
		BuildingID   string           `json:"building_id"`
		UnitID       string           `json:"unit_id"`
		PayType      string           `json:"pay_type"`
		FinalAmount  int64            `json:"final_amount"`
		BillObjs     []map[string]any `json:"bill_objs"`
		HandleFeeObj []map[string]any `json:"handle_fee_obj"`
		Remark       string           `json:"remark"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid terminal payment payload"))
		return
	}

	result, err := h.posPaymentService.PayByTerminal(c.Request.Context(), service.POSTerminalPaymentParams{
		UserID: user.UserID,
		Selection: service.POSPaymentSelection{
			BuildingID: strings.TrimSpace(request.BuildingID),
			UnitID:     strings.TrimSpace(request.UnitID),
		},
		PayType:      strings.TrimSpace(request.PayType),
		FinalAmount:  request.FinalAmount,
		BillObjs:     request.BillObjs,
		HandleFeeObj: request.HandleFeeObj,
		Remark:       strings.TrimSpace(request.Remark),
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 5. QueryOrder queries one H5 order by merchant order number or pay order id.
func (h *POSPaymentHandler) QueryOrder(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	var request map[string]any
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid order query payload"))
		return
	}

	result, err := h.posPaymentService.QueryH5Order(c.Request.Context(), user.UserID, request, paymentSelectionFromPayload(c, request))
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 6. CloseOrder closes one H5 order.
func (h *POSPaymentHandler) CloseOrder(c *gin.Context) {
	h.handleOrderAction(c, "close")
}

// 7. CancelOrder cancels one H5 order.
func (h *POSPaymentHandler) CancelOrder(c *gin.Context) {
	h.handleOrderAction(c, "cancel")
}

// 8. SimulateOrder simulates one visible H5 order in non-production.
func (h *POSPaymentHandler) SimulateOrder(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	var request map[string]any
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid simulation payload"))
		return
	}

	result, err := h.posPaymentService.SimulateH5Order(
		c.Request.Context(),
		user.UserID,
		strings.TrimSpace(c.Param("mchOrderNo")),
		request,
		paymentSelectionFromPayload(c, request),
	)
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 9. GetHistoryDetail returns one POS transaction detail.
func (h *POSPaymentHandler) GetHistoryDetail(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	result, err := h.posPaymentService.GetHistoryDetail(
		c.Request.Context(),
		user.UserID,
		strings.TrimSpace(c.Param("paymentId")),
		paymentHistoryQueryFromRequest(c),
	)
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 10. ClearAccounting clears selected cash or cheque transactions.
func (h *POSPaymentHandler) ClearAccounting(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	var request struct {
		BuildingID    string   `json:"building_id"`
		PaymentIDList []string `json:"payment_id_list"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid accounting clear payload"))
		return
	}

	result, err := h.posPaymentService.ClearAccounting(c.Request.Context(), service.POSPaymentAccountingClearParams{
		UserID: user.UserID,
		Selection: service.POSPaymentSelection{
			BuildingID: strings.TrimSpace(request.BuildingID),
		},
		PaymentIDList: request.PaymentIDList,
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 11. ListAccountingRecords returns visible accounting clear records.
func (h *POSPaymentHandler) ListAccountingRecords(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	result, err := h.posPaymentService.ListAccountingRecords(c.Request.Context(), user.UserID, paymentSelectionFromRequest(c))
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 12. GetAccountingRecord returns one visible accounting clear record detail.
func (h *POSPaymentHandler) GetAccountingRecord(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	result, err := h.posPaymentService.GetAccountingRecord(
		c.Request.Context(),
		user.UserID,
		strings.TrimSpace(c.Param("recordId")),
		paymentSelectionFromRequest(c),
	)
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 13. handleOrderAction routes H5 order action requests.
func (h *POSPaymentHandler) handleOrderAction(c *gin.Context, action string) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	var request map[string]any
	if err := c.ShouldBindJSON(&request); err != nil {
		request = map[string]any{}
	}

	result, err := h.posPaymentService.UpdateH5OrderAction(
		c.Request.Context(),
		user.UserID,
		strings.TrimSpace(c.Param("mchOrderNo")),
		action,
		request,
		paymentSelectionFromPayload(c, request),
	)
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 14. paymentHistoryQueryFromRequest reads POS history filters.
func paymentHistoryQueryFromRequest(c *gin.Context) service.POSPaymentHistoryQuery {
	return service.POSPaymentHistoryQuery{
		Selection: paymentSelectionFromRequest(c),
		FromDate:  strings.TrimSpace(c.Query("from_date")),
		ToDate:    strings.TrimSpace(c.Query("to_date")),
		DateType:  strings.TrimSpace(c.DefaultQuery("date_type", "input_date")),
		PayMethod: strings.TrimSpace(c.DefaultQuery("pay_method", "all")),
		UnitIDs:   parseCSVValues(c.QueryArray("unit_id_list"), c.Query("unit_ids")),
	}
}

// 15. paymentSelectionFromPayload reads selected POS context from query or JSON.
func paymentSelectionFromPayload(c *gin.Context, payload map[string]any) service.POSPaymentSelection {
	selection := paymentSelectionFromRequest(c)
	if selection.BuildingID == "" {
		selection.BuildingID = firstPayloadString(payload, "building_id", "BLG_ID", "blg_id")
	}
	if selection.UnitID == "" {
		selection.UnitID = firstPayloadString(payload, "unit_id", "UNIT_ID")
	}
	return selection
}

// 16. firstPayloadString returns the first non-empty string payload field.
func firstPayloadString(payload map[string]any, keys ...string) string {
	for _, key := range keys {
		value := strings.TrimSpace(fmt.Sprint(payload[key]))
		if payload[key] == nil {
			value = ""
		}
		if value != "" {
			return value
		}
	}
	return ""
}

// 15. parseCSVValues normalizes repeated and comma-separated query values.
func parseCSVValues(groups []string, csv string) []string {
	values := append([]string{}, groups...)
	if strings.TrimSpace(csv) != "" {
		values = append(values, strings.Split(csv, ",")...)
	}

	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, raw := range values {
		value := strings.TrimSpace(raw)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

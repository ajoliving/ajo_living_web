/*
 * POS 物業繳費擴展服務。
 * 1. 代理手續費、銀行賬戶、線下繳費、歷史詳情與清機介面。
 * 2. 代理 H5 訂單查詢、關閉與取消介面。
 * 3. 在 AJO 權限模型內校驗 ismart 綁定、可見大廈與可見單位。
 */
package service

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"gorm.io/gorm"

	"ajoliving_web/http_service/internal/errcode"
)

// 1. POSPaymentHistoryQuery defines transaction history filters.
type POSPaymentHistoryQuery struct {
	Selection POSPaymentSelection
	FromDate  string
	ToDate    string
	DateType  string
	PayMethod string
	UnitIDs   []string
}

// 2. POSPaymentReportParams defines offline payment report input.
type POSPaymentReportParams struct {
	UserID    int64
	Selection POSPaymentSelection
	Payload   map[string]any
}

// 3. POSPaymentAccountingResponse defines Staff accounting overview.
type POSPaymentAccountingResponse struct {
	Context         *POSUnitContext  `json:"context,omitempty"`
	BuildingOptions []string         `json:"building_options"`
	UnitOptions     []string         `json:"unit_options"`
	CashItems       []map[string]any `json:"cash_items"`
	ChequeItems     []map[string]any `json:"cheque_items"`
	HistoryItems    []map[string]any `json:"history_items"`
	Items           []map[string]any `json:"items"`
}

// 4. POSPaymentAccountingClearParams defines Staff clear-machine input.
type POSPaymentAccountingClearParams struct {
	UserID        int64
	Selection     POSPaymentSelection
	PaymentIDList []string
}

// 5. ListFees returns POS payment method and handling fee settings.
func (s *POSPaymentService) ListFees(ctx context.Context, userID int64, selection POSPaymentSelection) (*POSPaymentListResponse, error) {
	contextValue, token, err := s.memberPOSAccess(ctx, userID, selection)
	if err != nil {
		return nil, err
	}
	account, err := s.loadIsmartAccount(ctx, userID)
	if err != nil {
		return nil, err
	}

	query := url.Values{}
	query.Set("pos_type", posPaymentClientType(account))

	var rows []map[string]any
	path := "/building/" + url.PathEscape(contextValue.BuildingID) + "/fee"
	if err := s.posRelayJSONWithRefresh(ctx, userID, token, http.MethodGet, path, query, nil, &rows); err != nil {
		return nil, err
	}

	buildingOptions, unitOptions, _ := s.memberPaymentBindings(ctx, userID, nil)
	return &POSPaymentListResponse{Context: contextValue, BuildingOptions: buildingOptions, UnitOptions: unitOptions, Items: rows}, nil
}

// 6. ListBankAccounts returns bank accounts for the selected building.
func (s *POSPaymentService) ListBankAccounts(ctx context.Context, userID int64, selection POSPaymentSelection) (*POSPaymentListResponse, error) {
	contextValue, token, err := s.memberPOSAccess(ctx, userID, selection)
	if err != nil {
		return nil, err
	}

	var rows []map[string]any
	path := "/building/" + url.PathEscape(contextValue.BuildingID) + "/bank_account"
	if err := s.posRelayJSONWithRefresh(ctx, userID, token, http.MethodGet, path, nil, nil, &rows); err != nil {
		return nil, err
	}

	buildingOptions, unitOptions, _ := s.memberPaymentBindings(ctx, userID, nil)
	return &POSPaymentListResponse{Context: contextValue, BuildingOptions: buildingOptions, UnitOptions: unitOptions, Items: rows}, nil
}

// 7. ReportPayment submits offline POS payment through POS relay.
func (s *POSPaymentService) ReportPayment(ctx context.Context, params POSPaymentReportParams) (map[string]any, error) {
	contextValue, token, err := s.memberPOSAccess(ctx, params.UserID, params.Selection)
	if err != nil {
		return nil, err
	}
	payload := paymentCloneMap(params.Payload)
	if len(payload) == 0 {
		return nil, errcode.New(errcode.CodeValidationError, "payment report payload is required")
	}
	if err := ensureReportPayloadContext(payload, contextValue); err != nil {
		return nil, err
	}

	var result map[string]any
	if err := s.posRelayJSONWithRefresh(ctx, params.UserID, token, http.MethodPost, "/bill", nil, payload, &result); err != nil {
		return nil, err
	}
	if err := s.creditPOSReportPaymentReward(ctx, params.UserID, payload, result); err != nil {
		return nil, err
	}
	return result, nil
}

// 8. ListHistoryByQuery returns transaction history with AJO visibility checks.
func (s *POSPaymentService) ListHistoryByQuery(ctx context.Context, userID int64, query POSPaymentHistoryQuery) (*POSPaymentListResponse, error) {
	contextValue, token, unitIDs, err := s.resolveHistoryAccess(ctx, userID, query)
	if err != nil {
		return nil, err
	}

	rows, err := s.fetchHistoryRows(ctx, userID, token, contextValue, unitIDs, query)
	if err != nil {
		return nil, err
	}
	rows = filterPOSHistoryRows(rows, query)

	buildingOptions, unitOptions, _ := s.memberPaymentBindings(ctx, userID, nil)
	return &POSPaymentListResponse{Context: contextValue, BuildingOptions: buildingOptions, UnitOptions: unitOptions, Items: rows}, nil
}

// 9. GetHistoryDetail returns one visible transaction history row.
func (s *POSPaymentService) GetHistoryDetail(ctx context.Context, userID int64, paymentID string, query POSPaymentHistoryQuery) (map[string]any, error) {
	if strings.TrimSpace(paymentID) == "" {
		return nil, errcode.New(errcode.CodeValidationError, "payment id is required")
	}

	result, err := s.ListHistoryByQuery(ctx, userID, query)
	if err != nil {
		return nil, err
	}
	for _, row := range result.Items {
		if posHistoryRowMatchesID(row, paymentID) {
			return row, nil
		}
	}

	return nil, errcode.New(errcode.CodeNotFound, "payment history is not found")
}

// 10. ListAccountingGroups returns Staff cash and cheque pending groups.
func (s *POSPaymentService) ListAccountingGroups(ctx context.Context, userID int64, selection POSPaymentSelection) (*POSPaymentAccountingResponse, error) {
	contextValue, token, err := s.staffBuildingAccess(ctx, userID, selection)
	if err != nil {
		return nil, err
	}

	var pending map[string]any
	path := "/building/" + url.PathEscape(contextValue.BuildingID) + "/accounting"
	if err := s.posRelayJSONWithRefresh(ctx, userID, token, http.MethodGet, path, nil, nil, &pending); err != nil {
		return nil, err
	}

	var history any
	historyPath := "/building/" + url.PathEscape(contextValue.BuildingID) + "/accounting_record"
	if err := s.posRelayJSONWithRefresh(ctx, userID, token, http.MethodGet, historyPath, nil, nil, &history); err != nil {
		return nil, err
	}

	cashItems, chequeItems := splitAccountingRows(pending)
	items := append([]map[string]any{}, cashItems...)
	items = append(items, chequeItems...)
	buildingOptions, unitOptions, _ := s.memberPaymentBindings(ctx, userID, nil)
	return &POSPaymentAccountingResponse{
		Context:         contextValue,
		BuildingOptions: buildingOptions,
		UnitOptions:     unitOptions,
		CashItems:       cashItems,
		ChequeItems:     chequeItems,
		HistoryItems:    normalizeLooseRows(history, "items"),
		Items:           items,
	}, nil
}

// 11. ClearAccounting submits selected Staff cash or cheque transaction ids.
func (s *POSPaymentService) ClearAccounting(ctx context.Context, params POSPaymentAccountingClearParams) (map[string]any, error) {
	contextValue, token, err := s.staffBuildingAccess(ctx, params.UserID, params.Selection)
	if err != nil {
		return nil, err
	}
	paymentIDs := normalizeStringSlice(params.PaymentIDList)
	if len(paymentIDs) == 0 {
		return nil, errcode.New(errcode.CodeValidationError, "payment id list is required")
	}
	if err := s.ensureAccountingPaymentIDs(ctx, params.UserID, token, contextValue, paymentIDs); err != nil {
		return nil, err
	}

	var result map[string]any
	payload := map[string]any{"payment_id_list": paymentIDs}
	if err := s.posRelayJSONWithRefresh(ctx, params.UserID, token, http.MethodPost, "/accounting", nil, payload, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// 12. ListAccountingRecords returns Staff clear-machine history records.
func (s *POSPaymentService) ListAccountingRecords(ctx context.Context, userID int64, selection POSPaymentSelection) (*POSPaymentListResponse, error) {
	contextValue, token, err := s.staffBuildingAccess(ctx, userID, selection)
	if err != nil {
		return nil, err
	}

	var response any
	path := "/building/" + url.PathEscape(contextValue.BuildingID) + "/accounting_record"
	if err := s.posRelayJSONWithRefresh(ctx, userID, token, http.MethodGet, path, nil, nil, &response); err != nil {
		return nil, err
	}

	buildingOptions, unitOptions, _ := s.memberPaymentBindings(ctx, userID, nil)
	return &POSPaymentListResponse{Context: contextValue, BuildingOptions: buildingOptions, UnitOptions: unitOptions, Items: normalizeLooseRows(response, "items")}, nil
}

// 13. GetAccountingRecord returns one Staff clear-machine history detail.
func (s *POSPaymentService) GetAccountingRecord(ctx context.Context, userID int64, recordID string, selection POSPaymentSelection) (map[string]any, error) {
	contextValue, token, err := s.staffBuildingAccess(ctx, userID, selection)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(recordID) == "" {
		return nil, errcode.New(errcode.CodeValidationError, "record id is required")
	}

	var response map[string]any
	path := "/building/" + url.PathEscape(contextValue.BuildingID) + "/accounting_record/" + url.PathEscape(strings.TrimSpace(recordID))
	if err := s.posRelayJSONWithRefresh(ctx, userID, token, http.MethodGet, path, nil, nil, &response); err != nil {
		return nil, err
	}
	normalizePOSAccountingRecordRow(response)
	return response, nil
}

// 14. QueryH5Order queries one H5 payment order and checks visibility.
func (s *POSPaymentService) QueryH5Order(ctx context.Context, userID int64, payload map[string]any, selection POSPaymentSelection) (map[string]any, error) {
	contextValue, _, err := s.memberPOSAccess(ctx, userID, selection)
	if err != nil {
		return nil, err
	}

	var result map[string]any
	if err := s.posPaymentServiceJSON(ctx, http.MethodPost, "/h5/orders/query", payload, &result); err != nil {
		return nil, err
	}
	detail, err := s.h5OrderDetailFromQuery(ctx, result)
	if err != nil {
		return nil, err
	}
	if !h5OrderMatchesContext(detail, contextValue) {
		return nil, errcode.New(errcode.CodeAuthForbidden, "payment order is not visible")
	}
	if err := s.creditPOSPaymentReward(ctx, userID, detail); err != nil {
		return nil, err
	}
	return detail, nil
}

// 15. UpdateH5OrderAction closes or cancels a visible H5 order.
func (s *POSPaymentService) UpdateH5OrderAction(ctx context.Context, userID int64, mchOrderNo string, action string, payload map[string]any, selection POSPaymentSelection) (map[string]any, error) {
	if strings.TrimSpace(mchOrderNo) == "" {
		return nil, errcode.New(errcode.CodeValidationError, "merchant order number is required")
	}
	if action != "close" && action != "cancel" {
		return nil, errcode.New(errcode.CodeValidationError, "invalid order action")
	}
	contextValue, _, err := s.memberPOSAccess(ctx, userID, selection)
	if err != nil {
		return nil, err
	}

	current, err := s.GetH5Order(ctx, userID, mchOrderNo, true, false, selection)
	if err != nil {
		return nil, err
	}
	if !h5OrderMatchesContext(current, contextValue) {
		return nil, errcode.New(errcode.CodeAuthForbidden, "payment order is not visible")
	}

	var result map[string]any
	path := "/h5/orders/" + url.PathEscape(strings.TrimSpace(mchOrderNo)) + "/" + action
	if err := s.posPaymentServiceJSON(ctx, http.MethodPost, path, payload, &result); err != nil {
		return nil, err
	}
	if h5OrderHasContext(result) && !h5OrderMatchesContext(result, contextValue) {
		return nil, errcode.New(errcode.CodeAuthForbidden, "payment order is not visible")
	}
	return result, nil
}

// 16. SimulateH5Order proxies test-only H5 order simulation for Staff.
func (s *POSPaymentService) SimulateH5Order(ctx context.Context, userID int64, mchOrderNo string, payload map[string]any, selection POSPaymentSelection) (map[string]any, error) {
	if !posH5SimulationAllowed(s.runtime.Config.AppEnv) {
		return nil, errcode.New(errcode.CodeAuthForbidden, "payment simulation is disabled")
	}
	account, err := s.loadIsmartAccount(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !account.IsStaff {
		return nil, errcode.New(errcode.CodeAuthForbidden, "staff access required")
	}
	if strings.TrimSpace(mchOrderNo) == "" {
		return nil, errcode.New(errcode.CodeValidationError, "merchant order number is required")
	}

	contextValue, _, err := s.memberPOSAccess(ctx, userID, selection)
	if err != nil {
		return nil, err
	}
	current, err := s.GetH5Order(ctx, userID, mchOrderNo, true, false, selection)
	if err != nil {
		return nil, err
	}
	if !h5OrderMatchesContext(current, contextValue) {
		return nil, errcode.New(errcode.CodeAuthForbidden, "payment order is not visible")
	}

	requestPayload := paymentCloneMap(payload)
	if strings.TrimSpace(paymentStringValue(requestPayload["state"])) == "" {
		return nil, errcode.New(errcode.CodeValidationError, "simulation state is required")
	}

	var result map[string]any
	path := "/h5/orders/" + url.PathEscape(strings.TrimSpace(mchOrderNo)) + "/simulate"
	if err := s.posPaymentServiceJSON(ctx, http.MethodPost, path, requestPayload, &result); err != nil {
		return nil, err
	}
	if h5OrderHasContext(result) && !h5OrderMatchesContext(result, contextValue) {
		return nil, errcode.New(errcode.CodeAuthForbidden, "payment order is not visible")
	}
	return result, nil
}

// 17. posH5SimulationAllowed allows simulation only outside production.
func posH5SimulationAllowed(appEnv string) bool {
	return !strings.EqualFold(strings.TrimSpace(appEnv), "production")
}

// 18. creditPOSReportPaymentReward credits AJO Points once after offline POS payment succeeds.
func (s *POSPaymentService) creditPOSReportPaymentReward(ctx context.Context, userID int64, payload map[string]any, result map[string]any) error {
	if s.runtime.WalletService == nil {
		return nil
	}
	points := posPaymentRewardPoints(payload)
	if points <= 0 {
		return nil
	}
	key := posReportPaymentRewardKey(payload, result)
	if key == "" {
		return nil
	}

	return s.runtime.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		_, err := s.runtime.WalletService.CreditPointsWithTx(ctx, tx, WalletCreditParams{
			UserID:         userID,
			Amount:         points,
			SourceType:     WalletSourcePOSPayment,
			BizModule:      "payment",
			ActionType:     WalletActionPOSReward,
			IdempotencyKey: "pos_payment_reward:" + key,
			Note:           fmt.Sprintf("POS payment reward %s", key),
		})
		return err
	})
}

// 19. posReportPaymentRewardKey builds a stable reward idempotency key.
func posReportPaymentRewardKey(payload map[string]any, result map[string]any) string {
	for _, source := range []map[string]any{result, payload} {
		if source == nil {
			continue
		}
		for _, key := range []string{"receipt_id", "receipt_no", "TRAN_REF_NO", "tran_ref_no", "payment_id"} {
			value := strings.TrimSpace(paymentStringValue(source[key]))
			if value != "" {
				return value
			}
		}
		if nested := paymentMapValue(source["ismart_receipt_no"]); len(nested) > 0 {
			if value := strings.TrimSpace(paymentStringValue(nested["receipt_id"])); value != "" {
				return value
			}
		}
	}

	parts := []string{
		paymentStringValue(payload["BLG_ID"]),
		paymentStringValue(payload["UNIT_ID"]),
		paymentStringValue(payload["TRAN_DATETIME"]),
		paymentStringValue(payload["FINAL_AMOUNT"]),
	}
	return strings.Trim(strings.Join(parts, ":"), ":")
}

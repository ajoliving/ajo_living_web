/*
 * POS 物業繳費擴展輔助。
 * 1. 解析歷史查詢可見範圍與鬆散 POS 回應。
 * 2. 按現金、支票分組會計待清機交易。
 * 3. 對 POS 歷史資料執行日期、方式與明細匹配。
 */
package service

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"ajoliving_web/http_service/internal/errcode"
)

// 1. resolveHistoryAccess resolves visible history units and relay token.
func (s *POSPaymentService) resolveHistoryAccess(ctx context.Context, userID int64, query POSPaymentHistoryQuery) (*POSUnitContext, string, []string, error) {
	account, err := s.loadIsmartAccount(ctx, userID)
	if err != nil {
		return nil, "", nil, err
	}
	requestedUnitIDs := normalizeStringSlice(query.UnitIDs)
	if strings.TrimSpace(query.Selection.UnitID) != "" && len(requestedUnitIDs) == 0 {
		contextValue, token, err := s.memberPOSAccess(ctx, userID, query.Selection)
		if err != nil {
			return nil, "", nil, err
		}
		return contextValue, token, []string{contextValue.UnitID}, nil
	}
	if len(requestedUnitIDs) == 0 {
		contextValue, token, err := s.memberPOSAccess(ctx, userID, query.Selection)
		if err != nil {
			return nil, "", nil, err
		}
		return contextValue, token, []string{contextValue.UnitID}, nil
	}

	buildingSelection := query.Selection
	if strings.TrimSpace(buildingSelection.BuildingID) == "" {
		buildingSelection.UnitID = requestedUnitIDs[0]
	}
	contextValue, token, err := s.memberBuildingAccess(ctx, userID, buildingSelection)
	if err != nil {
		return nil, "", nil, err
	}
	_, unitOptions, err := s.memberPaymentBindings(ctx, userID, account)
	if err != nil {
		return nil, "", nil, err
	}

	if len(unitOptions) > 0 {
		for _, unitID := range requestedUnitIDs {
			if posUnitBuildingID(unitID) != contextValue.BuildingID || !containsString(unitOptions, unitID) {
				return nil, "", nil, errcode.New(errcode.CodeAuthForbidden, "unit is not visible")
			}
		}
	}
	if len(unitOptions) == 0 {
		for _, unitID := range requestedUnitIDs {
			selection := POSPaymentSelection{BuildingID: contextValue.BuildingID, UnitID: unitID}
			if _, _, err := s.memberPOSAccess(ctx, userID, selection); err != nil {
				return nil, "", nil, err
			}
		}
	}

	return contextValue, token, requestedUnitIDs, nil
}

// 2. fetchHistoryRows reads history rows from POS relay.
func (s *POSPaymentService) fetchHistoryRows(ctx context.Context, userID int64, token string, contextValue *POSUnitContext, unitIDs []string, query POSPaymentHistoryQuery) ([]map[string]any, error) {
	if len(unitIDs) > 0 {
		payload := map[string]any{"unit_id_list": unitIDs}
		var response map[string]any
		if err := s.posRelayJSONWithRefresh(ctx, userID, token, http.MethodPost, "/transactions/flat_units", nil, payload, &response); err != nil {
			return nil, err
		}
		return normalizePOSRows(response, "payment_objs"), nil
	}

	requestQuery := url.Values{}
	requestQuery.Set("blg_id", contextValue.BuildingID)
	requestQuery.Set("from_date", strings.TrimSpace(query.FromDate))
	requestQuery.Set("to_date", strings.TrimSpace(query.ToDate))
	requestQuery.Set("date_type", paymentDefaultHistoryDateType(query.DateType))
	requestQuery.Set("pay_method", paymentDefaultHistoryPayMethod(query.PayMethod))

	var response map[string]any
	path := "/building/" + url.PathEscape(contextValue.BuildingID) + "/bill_history"
	if err := s.posRelayJSONWithRefresh(ctx, userID, token, http.MethodGet, path, requestQuery, nil, &response); err != nil {
		return nil, err
	}
	return normalizePOSRows(response, "payment_objs"), nil
}

// 3. filterPOSHistoryRows applies date and pay method filters locally.
func filterPOSHistoryRows(rows []map[string]any, query POSPaymentHistoryQuery) []map[string]any {
	fromDate := strings.TrimSpace(query.FromDate)
	toDate := strings.TrimSpace(query.ToDate)
	payMethod := strings.ToUpper(strings.TrimSpace(query.PayMethod))
	if fromDate == "" && toDate == "" && (payMethod == "" || payMethod == "ALL") {
		return rows
	}

	result := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		if !historyRowMatchesDate(row, query.DateType, fromDate, toDate) {
			continue
		}
		if !historyRowMatchesPayMethod(row, payMethod) {
			continue
		}
		result = append(result, row)
	}
	return result
}

// 4. posHistoryRowMatchesID checks one history row against an id.
func posHistoryRowMatchesID(row map[string]any, paymentID string) bool {
	target := strings.TrimSpace(paymentID)
	for _, key := range []string{"payment_id", "receipt_id", "receiptNo", "ref_no", "id"} {
		if strings.TrimSpace(paymentStringValue(row[key])) == target {
			return true
		}
	}
	return false
}

// 5. splitAccountingRows returns cash and cheque pending rows.
func splitAccountingRows(payload map[string]any) ([]map[string]any, []map[string]any) {
	cashItems := rowsAtKey(payload, "payment_objs_cash")
	chequeItems := rowsAtKey(payload, "payment_objs_cheque")
	if len(cashItems) > 0 || len(chequeItems) > 0 {
		return cashItems, chequeItems
	}

	for _, row := range normalizePOSAccountingRows(payload) {
		method := strings.ToUpper(paymentFirstNonEmpty(
			paymentStringValue(row["pay_type"]),
			paymentStringValue(row["pay_method"]),
			paymentStringValue(row["method"]),
		))
		if strings.Contains(method, "CHEQUE") || strings.Contains(method, "CHECK") {
			chequeItems = append(chequeItems, row)
			continue
		}
		cashItems = append(cashItems, row)
	}
	return cashItems, chequeItems
}

// 6. normalizeLooseRows extracts rows from loose POS relay payloads.
func normalizeLooseRows(value any, key string) []map[string]any {
	result := []map[string]any{}
	if payload := paymentMapValue(value); len(payload) > 0 {
		for _, candidateKey := range []string{key, "items", "records", "payment_objs", "data"} {
			if rows := rowsAtKey(payload, candidateKey); len(rows) > 0 {
				result = rows
				break
			}
		}
	}

	if len(result) == 0 {
		if rows, ok := anyToMapRows(value); ok {
			result = rows
		}
	}

	for index := range result {
		normalizePOSAccountingRecordRow(result[index])
	}
	return result
}

// 7. rowsAtKey extracts a row list from a concrete payload key.
func rowsAtKey(payload map[string]any, key string) []map[string]any {
	if payload == nil {
		return []map[string]any{}
	}
	if rows, ok := anyToMapRows(payload[key]); ok {
		return rows
	}
	return []map[string]any{}
}

// 8. normalizePOSAccountingRecordRow fills stable fields for accounting history rows.
func normalizePOSAccountingRecordRow(row map[string]any) {
	if row == nil {
		return
	}
	if strings.TrimSpace(paymentStringValue(row["id"])) == "" {
		if value := paymentFirstNonEmpty(
			paymentStringValue(row["record_id"]),
			paymentStringValue(row["accounting_record_id"]),
		); value != "" {
			row["id"] = value
		}
	}
	if strings.TrimSpace(paymentStringValue(row["time"])) == "" {
		if value := paymentFirstNonEmpty(
			paymentStringValue(row["record_time"]),
			paymentStringValue(row["created_at"]),
			paymentStringValue(row["input_time"]),
			paymentStringValue(row["date"]),
		); value != "" {
			row["time"] = value
		}
	}
	if strings.TrimSpace(paymentStringValue(row["pay_method"])) == "" {
		if value := paymentFirstNonEmpty(
			paymentStringValue(row["pay_type"]),
			paymentStringValue(row["method"]),
			paymentStringValue(row["PAY_METHOD"]),
		); value != "" {
			row["pay_method"] = value
		}
	}
	if strings.TrimSpace(paymentStringValue(row["amount"])) == "" {
		if value := paymentFirstNonEmpty(
			paymentStringValue(row["trs_val"]),
			paymentStringValue(row["total"]),
			paymentStringValue(row["cash_amount"]),
			paymentStringValue(row["cheque_amount"]),
		); value != "" {
			row["amount"] = value
		}
	}
}

// 9. historyRowMatchesDate checks one history row against date range.
func historyRowMatchesDate(row map[string]any, dateType string, fromDate string, toDate string) bool {
	dateValue := historyRowDate(row, dateType)
	if dateValue == "" {
		return true
	}
	if fromDate != "" && dateValue < fromDate {
		return false
	}
	if toDate != "" && dateValue > toDate {
		return false
	}
	return true
}

// 10. historyRowMatchesPayMethod checks one history row against pay method.
func historyRowMatchesPayMethod(row map[string]any, payMethod string) bool {
	if payMethod == "" || payMethod == "ALL" {
		return true
	}
	value := strings.ToUpper(paymentFirstNonEmpty(
		paymentStringValue(row["pay_type"]),
		paymentStringValue(row["pay_method"]),
		paymentStringValue(row["method"]),
	))
	return value == payMethod
}

// 11. historyRowDate returns the selected history date.
func historyRowDate(row map[string]any, dateType string) string {
	if strings.TrimSpace(dateType) == "tran_date" {
		return normalizeHistoryDate(paymentFirstNonEmpty(
			paymentStringValue(row["tran_time"]),
			paymentStringValue(row["tran_datetime"]),
			paymentStringValue(row["input_time"]),
		))
	}
	return normalizeHistoryDate(paymentFirstNonEmpty(
		paymentStringValue(row["input_time"]),
		paymentStringValue(row["tran_time"]),
		paymentStringValue(row["tran_datetime"]),
	))
}

// 12. normalizeHistoryDate extracts YYYY-MM-DD date prefix.
func normalizeHistoryDate(value string) string {
	text := strings.TrimSpace(value)
	if len(text) >= 10 {
		return text[:10]
	}
	return ""
}

// 13. paymentDefaultHistoryDateType returns POS relay compatible date type.
func paymentDefaultHistoryDateType(value string) string {
	if strings.TrimSpace(value) == "tran_date" {
		return "tran_date"
	}
	return "input_date"
}

// 14. paymentDefaultHistoryPayMethod returns POS relay compatible method filter.
func paymentDefaultHistoryPayMethod(value string) string {
	text := strings.TrimSpace(value)
	if text == "" {
		return "all"
	}
	return text
}

// 15. ensureReportPayloadContext rejects conflicting old POS payload context.
func ensureReportPayloadContext(payload map[string]any, contextValue *POSUnitContext) error {
	if contextValue == nil {
		return errcode.New(errcode.CodeValidationError, "payment context is required")
	}
	finalAmount := paymentStringValue(payload["FINAL_AMOUNT"])
	if strings.TrimSpace(finalAmount) == "" || paymentInt64Value(finalAmount) <= 0 {
		return errcode.New(errcode.CodeValidationError, "valid payment amount is required")
	}
	if strings.TrimSpace(paymentStringValue(payload["ENTRY_DATETIME"])) == "" ||
		strings.TrimSpace(paymentStringValue(payload["TRAN_DATETIME"])) == "" ||
		strings.TrimSpace(paymentStringValue(payload["TRAN_REF_NO"])) == "" ||
		strings.TrimSpace(paymentStringValue(payload["PAY_METHOD"])) == "" {
		return errcode.New(errcode.CodeValidationError, "valid payment report payload is required")
	}
	if rows, ok := anyToMapRows(payload["BILL_OBJS"]); !ok || len(rows) == 0 {
		return errcode.New(errcode.CodeValidationError, "bill object is required")
	}
	if payloadBuildingID := paymentFirstNonEmpty(
		paymentStringValue(payload["BLG_ID"]),
		paymentStringValue(payload["building_id"]),
		paymentStringValue(payload["blg_id"]),
	); payloadBuildingID != "" && payloadBuildingID != contextValue.BuildingID {
		return errcode.New(errcode.CodeAuthForbidden, "building is not visible")
	}
	if payloadUnitID := paymentFirstNonEmpty(
		paymentStringValue(payload["UNIT_ID"]),
		paymentStringValue(payload["unit_id"]),
	); payloadUnitID != "" && payloadUnitID != contextValue.UnitID {
		return errcode.New(errcode.CodeAuthForbidden, "unit is not visible")
	}

	payload["BLG_ID"] = contextValue.BuildingID
	payload["UNIT_ID"] = contextValue.UnitID
	payload["building_id"] = contextValue.BuildingID
	payload["unit_id"] = contextValue.UnitID
	payload["FINAL_AMOUNT"] = finalAmount
	return nil
}

// 16. ensureAccountingPaymentIDs verifies selected clear-machine ids are pending rows.
func (s *POSPaymentService) ensureAccountingPaymentIDs(ctx context.Context, userID int64, token string, contextValue *POSUnitContext, paymentIDs []string) error {
	var pending map[string]any
	path := "/building/" + url.PathEscape(contextValue.BuildingID) + "/accounting"
	if err := s.posRelayJSONWithRefresh(ctx, userID, token, http.MethodGet, path, nil, nil, &pending); err != nil {
		return err
	}

	visibleIDs := make(map[string]struct{})
	cashItems, chequeItems := splitAccountingRows(pending)
	for _, row := range append(cashItems, chequeItems...) {
		for _, key := range []string{"payment_id", "receipt_id", "id"} {
			value := strings.TrimSpace(paymentStringValue(row[key]))
			if value != "" {
				visibleIDs[value] = struct{}{}
			}
		}
	}
	for _, paymentID := range paymentIDs {
		if _, ok := visibleIDs[strings.TrimSpace(paymentID)]; !ok {
			return errcode.New(errcode.CodeAuthForbidden, "payment is not visible")
		}
	}
	return nil
}

// 18. h5OrderDetailFromQuery loads detail data for a queried H5 order.
func (s *POSPaymentService) h5OrderDetailFromQuery(ctx context.Context, queryResult map[string]any) (map[string]any, error) {
	mchOrderNo := paymentFirstNonEmpty(
		paymentStringValue(paymentMapValue(queryResult["data"])["mch_order_no"]),
		paymentStringValue(paymentMapValue(queryResult["data"])["mchOrderNo"]),
		paymentStringValue(queryResult["mch_order_no"]),
		paymentStringValue(queryResult["mchOrderNo"]),
	)
	if strings.TrimSpace(mchOrderNo) == "" {
		return nil, errcode.New(errcode.CodeInternalError, "payment service response is missing order number")
	}

	var detail map[string]any
	path := "/h5/orders/" + url.PathEscape(mchOrderNo) + "/detail?refresh=1&retry_business=1"
	if err := s.posPaymentServiceJSON(ctx, http.MethodGet, path, nil, &detail); err != nil {
		return nil, err
	}
	return detail, nil
}

// 19. ensureH5OrderVisibleByDetail verifies order visibility when normal response lacks context.
func (s *POSPaymentService) ensureH5OrderVisibleByDetail(ctx context.Context, mchOrderNo string, contextValue *POSUnitContext, refresh bool) error {
	query := url.Values{}
	if refresh {
		query.Set("refresh", "1")
		query.Set("retry_business", "1")
	}
	path := "/h5/orders/" + url.PathEscape(strings.TrimSpace(mchOrderNo)) + "/detail"
	if len(query) > 0 {
		path += "?" + query.Encode()
	}

	var detail map[string]any
	if err := s.posPaymentServiceJSON(ctx, http.MethodGet, path, nil, &detail); err != nil {
		return err
	}
	if !h5OrderMatchesContext(detail, contextValue) {
		return errcode.New(errcode.CodeAuthForbidden, "payment order is not visible")
	}
	return nil
}

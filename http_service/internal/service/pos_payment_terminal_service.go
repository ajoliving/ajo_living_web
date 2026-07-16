/*
 * POS 機收款代理服務。
 * 1. 代理會員於可見單位使用櫃台 POS 機銀行卡收款。
 * 2. 使用服務端配置的終端地址，不接受前端傳入設備地址。
 * 3. 收款成功後透過 POS relay 上報物業費入賬。
 */
package service

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"ajoliving_web/http_service/internal/errcode"
)

const (
	posTerminalAllinpayType = "allinpay"
	posTerminalCardPayType  = "POS_CARD"
	posTerminalAliwePayType = "POS_ALIWE"
)

// 1. POSTerminalPaymentParams defines member POS terminal payment input.
type POSTerminalPaymentParams struct {
	UserID       int64
	Selection    POSPaymentSelection
	PayType      string
	FinalAmount  int64
	BillObjs     []map[string]any
	HandleFeeObj []map[string]any
	Remark       string
}

// 2. PayByTerminal charges a configured POS terminal and reports payment.
func (s *POSPaymentService) PayByTerminal(ctx context.Context, params POSTerminalPaymentParams) (map[string]any, error) {
	contextValue, token, err := s.memberPOSAccess(ctx, params.UserID, params.Selection)
	if err != nil {
		return nil, err
	}
	if params.FinalAmount <= 0 || len(params.BillObjs) == 0 {
		return nil, errcode.New(errcode.CodeValidationError, "valid terminal payment payload is required")
	}

	terminalType, terminalURL, err := s.terminalProxyConfig(contextValue.BuildingID)
	if err != nil {
		return nil, err
	}
	terminalPayload, rawResponse, err := s.chargeConfiguredTerminal(ctx, terminalType, terminalURL, contextValue, params)
	if err != nil {
		return nil, err
	}

	reportPayload := s.buildTerminalReportPayload(contextValue, terminalType, terminalPayload, rawResponse, params)
	if err := ensureReportPayloadContext(reportPayload, contextValue); err != nil {
		return nil, err
	}

	var result map[string]any
	if err := s.posRelayJSONWithRefresh(ctx, params.UserID, token, http.MethodPost, "/bill", nil, reportPayload, &result); err != nil {
		return nil, err
	}
	if result == nil {
		result = map[string]any{}
	}
	result["terminal_type"] = terminalType
	result["terminal_pay_type"] = normalizePOSTerminalPayType(params.PayType)
	result["terminal_response"] = terminalPayload
	return result, nil
}

// 3. TerminalProxyEnabled returns whether terminal payment is available.
func (s *POSPaymentService) TerminalProxyEnabled(buildingID string) bool {
	_, _, err := s.terminalProxyConfig(buildingID)
	return err == nil
}

// 4. terminalProxyConfig validates configured terminal proxy settings.
func (s *POSPaymentService) terminalProxyConfig(buildingID string) (string, string, error) {
	if !s.runtime.Config.POSTerminalProxyEnabled {
		return "", "", errcode.New(errcode.CodeValidationError, "pos terminal proxy is not enabled")
	}
	terminalType := strings.ToLower(strings.TrimSpace(s.runtime.Config.POSTerminalType))
	if terminalType == "" {
		terminalType = posTerminalAllinpayType
	}
	if terminalType != posTerminalAllinpayType {
		return "", "", errcode.New(errcode.CodeValidationError, "pos terminal type is not supported")
	}
	if !posTerminalBuildingAllowed(s.runtime.Config.POSTerminalAllowedBuildings, buildingID) {
		return "", "", errcode.New(errcode.CodeAuthForbidden, "building terminal is not available")
	}

	terminalURL := strings.TrimSpace(s.runtime.Config.POSTerminalURL)
	parsed, err := url.Parse(terminalURL)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return "", "", errcode.New(errcode.CodeValidationError, "pos terminal url is not configured")
	}
	return terminalType, terminalURL, nil
}

// 5. chargeConfiguredTerminal routes a payment to the configured terminal type.
func (s *POSPaymentService) chargeConfiguredTerminal(ctx context.Context, terminalType string, terminalURL string, contextValue *POSUnitContext, params POSTerminalPaymentParams) (map[string]any, string, error) {
	if terminalType != posTerminalAllinpayType {
		return nil, "", errcode.New(errcode.CodeValidationError, "pos terminal type is not supported")
	}
	return s.chargeAllinpayTerminal(ctx, terminalURL, contextValue, params)
}

// 6. chargeAllinpayTerminal posts one form payment request to Allinpay terminal.
func (s *POSPaymentService) chargeAllinpayTerminal(ctx context.Context, terminalURL string, contextValue *POSUnitContext, params POSTerminalPaymentParams) (map[string]any, string, error) {
	payType := normalizePOSTerminalPayType(params.PayType)
	businessID := posTerminalBusinessID(payType)
	if businessID == "" {
		return nil, "", errcode.New(errcode.CodeValidationError, "pos terminal pay type is not supported")
	}

	traceNo := terminalTraceNumber(s.runtime.Now(), contextValue.BuildingID)
	form := url.Values{}
	form.Set("BUSINESS_ID", businessID)
	form.Set("AMOUNT", formatPOSTerminalAmount(params.FinalAmount))
	form.Set("TRANS_TRACE_NO", traceNo)
	form.Set("TRANS_ORDER_NO", traceNo)
	form.Set("CURRENCY", posPaymentDefaultCurrency)

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, terminalURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, "", errcode.New(errcode.CodeInternalError, "failed to prepare pos terminal request")
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")

	client := &http.Client{Timeout: s.httpTimeout()}
	response, err := client.Do(request)
	if err != nil {
		return nil, "", errcode.New(errcode.CodeInternalError, "pos terminal request failed")
	}
	defer response.Body.Close()

	raw, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, "", errcode.New(errcode.CodeInternalError, "invalid pos terminal response")
	}
	rawText := strings.TrimSpace(string(raw))
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, rawText, errcode.New(errcode.CodeInternalError, "pos terminal payment failed")
	}

	return parsePOSTerminalResponse(rawText), rawText, nil
}

// 7. buildTerminalReportPayload creates the POS relay payment report body.
func (s *POSPaymentService) buildTerminalReportPayload(contextValue *POSUnitContext, terminalType string, terminalPayload map[string]any, rawResponse string, params POSTerminalPaymentParams) map[string]any {
	nowText := formatPOSTime(s.runtime.Now())
	payload := paymentCloneMap(terminalPayload)
	payload["FINAL_AMOUNT"] = strconv.FormatInt(params.FinalAmount, 10)
	payload["ENTRY_DATETIME"] = paymentFirstNonEmpty(paymentStringValue(payload["ENTRY_DATETIME"]), nowText)
	payload["TRAN_DATETIME"] = paymentFirstNonEmpty(paymentStringValue(payload["TRAN_DATETIME"]), paymentStringValue(payload["TRAN_TIME"]), nowText)
	payload["TRAN_REF_NO"] = paymentFirstNonEmpty(
		paymentStringValue(payload["TRAN_REF_NO"]),
		paymentStringValue(payload["REF_NO"]),
		paymentStringValue(payload["TRANS_TRACE_NO"]),
		paymentStringValue(payload["TRANS_ORDER_NO"]),
		terminalTraceNumber(s.runtime.Now(), contextValue.BuildingID),
	)
	payload["COMMENT"] = strings.TrimSpace(params.Remark)
	payload["BILL_OBJS"] = params.BillObjs
	payload["BILL_OBJ"] = params.BillObjs
	payload["HANDLE_FEE_OBJ"] = params.HandleFeeObj
	payload["BLG_ID"] = contextValue.BuildingID
	payload["UNIT_ID"] = contextValue.UnitID
	payload["building_id"] = contextValue.BuildingID
	payload["unit_id"] = contextValue.UnitID
	payload["PAY_METHOD"] = terminalType
	payload["bank_account_received"] = ""
	if strings.TrimSpace(rawResponse) != "" {
		payload["TERMINAL_RAW_RESPONSE"] = rawResponse
	}
	return payload
}

// 8. normalizePOSTerminalPayType normalizes supported terminal payment types.
func normalizePOSTerminalPayType(value string) string {
	text := strings.ToUpper(strings.TrimSpace(value))
	if text == "" {
		return posTerminalCardPayType
	}
	return text
}

// 9. posTerminalBusinessID returns Allinpay business id.
func posTerminalBusinessID(payType string) string {
	switch normalizePOSTerminalPayType(payType) {
	case posTerminalCardPayType:
		return "100100001"
	case posTerminalAliwePayType:
		return "100300001"
	default:
		return ""
	}
}

// 10. formatPOSTerminalAmount returns a 12-digit cent amount.
func formatPOSTerminalAmount(amountCents int64) string {
	if amountCents < 0 {
		amountCents = 0
	}
	text := strings.TrimSpace(strconv.FormatInt(amountCents, 10))
	if len(text) >= 12 {
		return text
	}
	return strings.Repeat("0", 12-len(text)) + text
}

// 11. terminalTraceNumber creates a stable POS terminal trace number.
func terminalTraceNumber(now time.Time, buildingID string) string {
	return now.Format("20060102150405") + strings.TrimSpace(buildingID)
}

// 12. formatPOSTime formats POS relay datetime values.
func formatPOSTime(value time.Time) string {
	return value.Format("2006-01-02 15:04:05")
}

// 13. parsePOSTerminalResponse parses JSON or form terminal responses.
func parsePOSTerminalResponse(rawText string) map[string]any {
	result := map[string]any{}
	if strings.TrimSpace(rawText) == "" {
		return result
	}
	if err := json.Unmarshal([]byte(rawText), &result); err == nil && result != nil {
		return result
	}
	if values, err := url.ParseQuery(rawText); err == nil && len(values) > 0 {
		for key, item := range values {
			if len(item) > 0 {
				result[key] = item[0]
			}
		}
		return result
	}
	result["raw_response"] = rawText
	return result
}

// 14. posTerminalBuildingAllowed checks optional building allowlist.
func posTerminalBuildingAllowed(rawAllowlist string, buildingID string) bool {
	normalizedBuildingID := strings.TrimSpace(buildingID)
	if normalizedBuildingID == "" {
		return false
	}
	parts := strings.FieldsFunc(rawAllowlist, func(char rune) bool {
		return char == ',' || char == ';' || char == '\n' || char == '\t' || char == ' '
	})
	allowed := normalizeStringSlice(parts)
	if len(allowed) == 0 {
		return true
	}
	return containsString(allowed, normalizedBuildingID)
}

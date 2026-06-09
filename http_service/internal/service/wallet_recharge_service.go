/*
 * Wallet recharge payment service.
 * 1. Create EasyLink wallet recharge orders for WeChat, Alipay, and UnionPay QR.
 * 2. Verify gateway responses and callbacks before updating local order state.
 * 3. Credit AJO Point exactly once after a payment reaches success.
 */
package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"ajoliving_web/http_service/internal/config"
	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/model"
	"ajoliving_web/http_service/internal/utils"
)

const (
	walletRechargeMinCents = int64(1)
	walletRechargeMaxCents = int64(5000000)
)

// 1. WalletRechargeCreateParams defines member recharge input.
type WalletRechargeCreateParams struct {
	UserID     int64
	AmountHKD  float64
	PayMethod  string
	PayRegion  string
	DeviceMode string
	ReturnPath string
	ClientIP   string
	UserAgent  string
}

// 2. WalletRechargeResponse defines one recharge order payload.
type WalletRechargeResponse struct {
	OrderID          string `json:"order_id"`
	MchOrderNo       string `json:"mch_order_no"`
	PayOrderID       string `json:"pay_order_id"`
	PayChannel       string `json:"pay_channel"`
	PayRegion        string `json:"pay_region"`
	PayDataType      string `json:"pay_data_type"`
	PayData          string `json:"pay_data"`
	State            string `json:"state"`
	StateLabel       string `json:"state_label"`
	GatewayStateCode int    `json:"gateway_state_code"`
	GatewayMessage   string `json:"gateway_message"`
	Currency         string `json:"currency"`
	AmountHKD        string `json:"amount_hkd"`
	AmountCents      int64  `json:"amount_cents"`
	PointsAmount     int64  `json:"points_amount"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
	ExpireTime       string `json:"expire_time"`
	PaidAt           string `json:"paid_at"`
	CreditedAt       string `json:"credited_at"`
}

// 3. PaymentNotifyResult defines raw gateway notify handling result.
type PaymentNotifyResult struct {
	StatusCode int
	Body       string
}

// 4. CreateRechargeOrder creates an EasyLink payment order.
func (s *WalletService) CreateRechargeOrder(ctx context.Context, params WalletRechargeCreateParams) (*WalletRechargeResponse, error) {
	if !s.runtime.Config.WalletRechargeEnabled {
		return nil, errcode.New(errcode.CodeWalletActionUnavailable, "wallet recharge is not enabled")
	}
	amountCents := normalizeRechargeAmountCents(params.AmountHKD)
	if amountCents < walletRechargeMinCents || amountCents > walletRechargeMaxCents {
		return nil, errcode.New(errcode.CodeValidationError, "recharge amount is invalid")
	}

	payChannel, payRegion, err := s.resolveRechargePayChannel(params.PayMethod, params.PayRegion, params.DeviceMode)
	if err != nil {
		return nil, err
	}
	channelConfig, err := s.paymentChannelConfig(payChannel)
	if err != nil {
		return nil, err
	}

	mchOrderNo := s.createRechargeMchOrderNo()
	expireSeconds := s.normalizedPaymentExpireSeconds()
	expireTime := s.runtime.Now().Add(time.Duration(expireSeconds) * time.Second)
	returnURL, err := s.buildRechargeReturnURL(params.ReturnPath, mchOrderNo, payChannel)
	if err != nil {
		return nil, errcode.New(errcode.CodeValidationError, "invalid payment return URL")
	}

	pointsAmount := amountCents * WalletRechargePointRate / 100
	requestPayload := map[string]any{
		"amount_hkd":  float64(amountCents) / 100,
		"pay_method":  strings.TrimSpace(params.PayMethod),
		"pay_region":  payRegion,
		"device_mode": strings.TrimSpace(params.DeviceMode),
		"return_path": strings.TrimSpace(params.ReturnPath),
	}
	extParamBytes, _ := json.Marshal(map[string]any{
		"source":       "ajoliving_wallet",
		"scene":        "wallet_recharge",
		"user_id":      params.UserID,
		"mch_order_no": mchOrderNo,
	})
	gatewayPayload := map[string]any{
		"mchNo":       s.runtime.Config.PaymentMerchantNo,
		"appId":       channelConfig.AppID,
		"mchOrderNo":  mchOrderNo,
		"wayCode":     payChannel,
		"amount":      amountCents,
		"currency":    "HKD",
		"subject":     "AJO Point Recharge",
		"body":        fmt.Sprintf("AJO Point recharge %s", mchOrderNo),
		"notifyUrl":   s.runtime.Config.PaymentNotifyURL,
		"expiredTime": expireSeconds,
		"extParam":    string(extParamBytes),
		"reqTime":     s.runtime.Now().UnixMilli(),
		"version":     "1.0",
		"signType":    "MD5",
	}
	if shouldSendRechargeReturnURL(payChannel) {
		gatewayPayload["returnUrl"] = returnURL
	}
	if channelExtra := s.resolveRechargeChannelExtra(payChannel, payRegion, channelConfig); channelExtra != "" {
		gatewayPayload["channelExtra"] = channelExtra
	}
	if strings.TrimSpace(params.ClientIP) != "" {
		gatewayPayload["clientIp"] = strings.TrimSpace(params.ClientIP)
	}
	gatewayPayload["sign"] = paymentSignPayload(gatewayPayload, channelConfig.AppSecret)

	gatewayResponse, err := s.postPaymentGateway(ctx, "/api/pay/unifiedOrder", gatewayPayload)
	if err != nil {
		return nil, err
	}
	if err := verifyPaymentGatewayResponse(gatewayResponse, channelConfig.AppSecret); err != nil {
		return nil, errcode.New(errcode.CodeInternalError, err.Error())
	}

	gatewayData := paymentMapValue(gatewayResponse["data"])
	gatewayStateCode := paymentIntValue(paymentFirstNonNil(gatewayData["orderState"], gatewayData["state"], 1))
	order := model.WalletRechargeOrder{
		PublicID:              utils.NewPublicID(),
		UserID:                params.UserID,
		MchOrderNo:            mchOrderNo,
		PayOrderID:            paymentStringValue(gatewayData["payOrderId"]),
		PayChannel:            payChannel,
		PayRegion:             payRegion,
		PayDataType:           paymentStringValue(gatewayData["payDataType"]),
		PayData:               paymentStringValue(gatewayData["payData"]),
		Currency:              "HKD",
		AmountCents:           amountCents,
		PointsAmount:          pointsAmount,
		State:                 normalizePaymentGatewayState(gatewayStateCode),
		GatewayStateCode:      gatewayStateCode,
		GatewayMessage:        paymentFirstNonEmpty(paymentStringValue(gatewayData["errMsg"]), paymentStringValue(gatewayResponse["msg"])),
		RequestPayload:        mustMarshalJSON(requestPayload),
		GatewayCreateRequest:  mustMarshalJSON(gatewayPayload),
		GatewayCreateResponse: mustMarshalJSON(gatewayResponse),
		ClientIP:              strings.TrimSpace(params.ClientIP),
		UserAgent:             strings.TrimSpace(params.UserAgent),
		ExpireTime:            &expireTime,
	}

	if order.State == PaymentStateSuccess {
		now := s.runtime.Now()
		order.PaidAt = &now
	}

	if err := s.runtime.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&order).Error; err != nil {
			return errcode.New(errcode.CodeInternalError, "failed to save recharge order")
		}
		if order.State == PaymentStateSuccess {
			return s.creditRechargeOrderWithTx(ctx, tx, &order)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return s.toRechargeResponse(order), nil
}

// 5. GetRechargeOrder returns one member recharge order and optionally refreshes it.
func (s *WalletService) GetRechargeOrder(ctx context.Context, userID int64, orderPublicID string, refresh bool) (*WalletRechargeResponse, error) {
	order, err := s.loadRechargeOrderForMember(ctx, userID, orderPublicID)
	if err != nil {
		return nil, err
	}
	if refresh && order.State == PaymentStatePaying {
		if order, err = s.refreshRechargeOrder(ctx, order); err != nil {
			return nil, err
		}
	}

	return s.toRechargeResponse(order), nil
}

// 6. HandlePaymentNotify verifies EasyLink callback and credits successful recharge.
func (s *WalletService) HandlePaymentNotify(ctx context.Context, contentType string, rawBody string) (*PaymentNotifyResult, error) {
	payload, err := parsePaymentNotifyBody(contentType, rawBody)
	if err != nil {
		return &PaymentNotifyResult{StatusCode: http.StatusBadRequest, Body: "fail"}, err
	}

	payOrderID := paymentFirstNonEmpty(
		paymentStringValue(payload["payOrderId"]),
		paymentStringValue(payload["pay_order_id"]),
		paymentStringValue(payload["paymentOrderId"]),
		paymentStringValue(payload["payment_order_id"]),
	)
	order, err := s.loadRechargeOrderByGatewayKeys(ctx, paymentStringValue(payload["mchOrderNo"]), payOrderID)
	if err != nil {
		return &PaymentNotifyResult{StatusCode: http.StatusBadRequest, Body: "fail"}, err
	}
	channelConfig, err := s.paymentChannelConfig(order.PayChannel)
	if err != nil {
		return &PaymentNotifyResult{StatusCode: http.StatusInternalServerError, Body: "fail"}, err
	}
	if !paymentVerifySignature(payload, paymentStringValue(payload["sign"]), channelConfig.AppSecret) {
		return &PaymentNotifyResult{StatusCode: http.StatusBadRequest, Body: "fail"}, errcode.New(errcode.CodeValidationError, "payment notify signature is invalid")
	}

	err = s.runtime.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var locked model.WalletRechargeOrder
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", order.ID).First(&locked).Error; err != nil {
			return errcode.New(errcode.CodeInternalError, "failed to load recharge order")
		}
		s.applyRechargeGatewayPayload(&locked, payload, true)
		if locked.State == PaymentStateSuccess {
			if err := s.creditRechargeOrderWithTx(ctx, tx, &locked); err != nil {
				return err
			}
		}
		return tx.Save(&locked).Error
	})
	if err != nil {
		return &PaymentNotifyResult{StatusCode: http.StatusInternalServerError, Body: "fail"}, err
	}

	return &PaymentNotifyResult{StatusCode: http.StatusOK, Body: "success"}, nil
}

// 7. refreshRechargeOrder queries EasyLink and reconciles local state.
func (s *WalletService) refreshRechargeOrder(ctx context.Context, order model.WalletRechargeOrder) (model.WalletRechargeOrder, error) {
	channelConfig, err := s.paymentChannelConfig(order.PayChannel)
	if err != nil {
		return order, err
	}
	queryPayload := map[string]any{
		"mchNo":      s.runtime.Config.PaymentMerchantNo,
		"appId":      channelConfig.AppID,
		"payOrderId": optionalPaymentString(order.PayOrderID),
		"mchOrderNo": order.MchOrderNo,
		"reqTime":    s.runtime.Now().UnixMilli(),
		"version":    "1.0",
		"signType":   "MD5",
	}
	deletePaymentEmptyValues(queryPayload)
	queryPayload["sign"] = paymentSignPayload(queryPayload, channelConfig.AppSecret)
	gatewayResponse, err := s.postPaymentGateway(ctx, "/api/pay/query", queryPayload)
	if err != nil {
		return order, err
	}
	if err := verifyPaymentGatewayResponse(gatewayResponse, channelConfig.AppSecret); err != nil {
		return order, errcode.New(errcode.CodeInternalError, err.Error())
	}

	err = s.runtime.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var locked model.WalletRechargeOrder
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", order.ID).First(&locked).Error; err != nil {
			return errcode.New(errcode.CodeInternalError, "failed to load recharge order")
		}
		locked.GatewayQueryRequest = mustMarshalJSON(queryPayload)
		locked.GatewayQueryResponse = mustMarshalJSON(gatewayResponse)
		s.applyRechargeGatewayPayload(&locked, paymentMapValue(gatewayResponse["data"]), false)
		if locked.State == PaymentStatePaying && locked.ExpireTime != nil && s.runtime.Now().After(*locked.ExpireTime) {
			locked.State = PaymentStateExpired
			locked.GatewayStateCode = 6
			now := s.runtime.Now()
			locked.ClosedAt = &now
		}
		if locked.State == PaymentStateSuccess {
			if err := s.creditRechargeOrderWithTx(ctx, tx, &locked); err != nil {
				return err
			}
		}
		if err := tx.Save(&locked).Error; err != nil {
			return errcode.New(errcode.CodeInternalError, "failed to update recharge order")
		}
		order = locked
		return nil
	})
	if err != nil {
		return order, err
	}

	return order, nil
}

// 8. creditRechargeOrderWithTx credits points once for a paid recharge order.
func (s *WalletService) creditRechargeOrderWithTx(ctx context.Context, tx *gorm.DB, order *model.WalletRechargeOrder) error {
	if order.WalletTransactionID != nil || order.CreditedAt != nil {
		return nil
	}
	credit, err := s.CreditPointsWithTx(ctx, tx, WalletCreditParams{
		UserID:         order.UserID,
		Amount:         order.PointsAmount,
		SourceType:     WalletSourceRecharge,
		BizModule:      "wallet",
		ActionType:     WalletActionRecharge,
		IdempotencyKey: "wallet_recharge:" + order.MchOrderNo,
		Note:           fmt.Sprintf("wallet recharge HKD %s", formatRechargeHKD(order.AmountCents)),
	})
	if err != nil {
		return err
	}

	var transaction model.WalletTransaction
	if err := tx.WithContext(ctx).Where("public_id = ?", credit.PointsTransactionID).First(&transaction).Error; err != nil {
		return errcode.New(errcode.CodeInternalError, "failed to load recharge transaction")
	}
	now := s.runtime.Now()
	order.WalletTransactionID = &transaction.ID
	order.CreditedAt = &now
	if order.PaidAt == nil {
		order.PaidAt = &now
	}

	return nil
}

// 9. applyRechargeGatewayPayload maps gateway fields to the order.
func (s *WalletService) applyRechargeGatewayPayload(order *model.WalletRechargeOrder, payload map[string]any, isNotify bool) {
	stateCode := paymentIntValue(paymentFirstNonNil(payload["state"], payload["orderState"], order.GatewayStateCode))
	if stateCode <= 0 {
		stateCode = order.GatewayStateCode
	}
	nextState := normalizePaymentGatewayState(stateCode)
	if order.State == PaymentStateSuccess && nextState == PaymentStatePaying {
		nextState = order.State
	}

	order.PayOrderID = paymentFirstNonEmpty(paymentStringValue(payload["payOrderId"]), order.PayOrderID)
	order.PayDataType = paymentFirstNonEmpty(paymentStringValue(payload["payDataType"]), order.PayDataType)
	order.PayData = paymentFirstNonEmpty(paymentStringValue(payload["payData"]), order.PayData)
	order.Currency = paymentFirstNonEmpty(paymentStringValue(payload["currency"]), order.Currency, "HKD")
	order.GatewayMessage = paymentFirstNonEmpty(paymentStringValue(payload["errMsg"]), paymentStringValue(payload["msg"]), order.GatewayMessage)
	order.State = nextState
	order.GatewayStateCode = stateCode
	if nextState == PaymentStateSuccess && order.PaidAt == nil {
		now := s.runtime.Now()
		order.PaidAt = &now
	}
	if nextState == PaymentStateClosed || nextState == PaymentStateExpired || nextState == PaymentStateRevoked || nextState == PaymentStateRefunded || nextState == PaymentStateFailed {
		if order.ClosedAt == nil {
			now := s.runtime.Now()
			order.ClosedAt = &now
		}
	}
	if isNotify {
		order.GatewayNotifyPayload = mustMarshalJSON(payload)
	}
}

// 10. resolveRechargePayChannel maps UI method, region, and device mode to EasyLink wayCode.
func (s *WalletService) resolveRechargePayChannel(payMethod string, payRegion string, deviceMode string) (string, string, error) {
	method := strings.ToLower(strings.TrimSpace(payMethod))
	region := strings.ToUpper(strings.TrimSpace(payRegion))
	if region != "CN" {
		region = "HK"
	}
	device := strings.ToLower(strings.TrimSpace(deviceMode))
	isDesktop := device == "desktop"

	switch method {
	case "wechat":
		if isDesktop {
			return "WX_QR", region, nil
		}
		return "WX_H5", region, nil
	case "alipay":
		if isDesktop {
			return "ALI_QR", region, nil
		}
		return "ALI_H5", region, nil
	case "unionpay":
		return "YSF_QR", region, nil
	default:
		return "", "", errcode.New(errcode.CodeValidationError, "unsupported recharge payment method")
	}
}

// 11. resolveRechargeChannelExtra applies region-specific channel extra fields.
func (s *WalletService) resolveRechargeChannelExtra(payChannel string, payRegion string, channelConfig config.PaymentChannelConfig) string {
	if payChannel == "ALI_H5" && strings.EqualFold(payRegion, "CN") {
		return `{"walletType":"CN"}`
	}
	if channelConfig.ChannelExtra != "" {
		return channelConfig.ChannelExtra
	}
	return ""
}

// 12. paymentChannelConfig returns channel credentials with validation.
func (s *WalletService) paymentChannelConfig(payChannel string) (config.PaymentChannelConfig, error) {
	if strings.TrimSpace(s.runtime.Config.PaymentMerchantNo) == "" {
		return config.PaymentChannelConfig{}, errcode.New(errcode.CodeInternalError, "payment merchant is not configured")
	}
	channelConfig, ok := s.runtime.Config.PaymentChannelConfigs[strings.ToUpper(strings.TrimSpace(payChannel))]
	if !ok {
		return config.PaymentChannelConfig{}, errcode.New(errcode.CodeValidationError, "unsupported payment channel")
	}
	if strings.TrimSpace(channelConfig.AppID) == "" || strings.TrimSpace(channelConfig.AppSecret) == "" {
		return config.PaymentChannelConfig{}, errcode.New(errcode.CodeInternalError, "payment channel is not configured")
	}

	return channelConfig, nil
}

// 13. postPaymentGateway sends a JSON request to EasyLink.
func (s *WalletService) postPaymentGateway(ctx context.Context, pathname string, payload map[string]any) (map[string]any, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to encode payment request")
	}

	requestURL := strings.TrimRight(s.runtime.Config.PaymentGatewayBaseURL, "/") + pathname
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, bytes.NewReader(body))
	if err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to create payment request")
	}
	request.Header.Set("Content-Type", "application/json;charset=UTF-8")

	client := &http.Client{Timeout: s.runtime.Config.PaymentRequestTimeout}
	response, err := client.Do(request)
	if err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "payment gateway request failed")
	}
	defer response.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to read payment gateway response")
	}
	var result map[string]any
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "payment gateway response is invalid")
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, errcode.New(errcode.CodeInternalError, "payment gateway request failed")
	}

	return result, nil
}

// 14. loadRechargeOrderForMember loads one order owned by a member.
func (s *WalletService) loadRechargeOrderForMember(ctx context.Context, userID int64, orderPublicID string) (model.WalletRechargeOrder, error) {
	var order model.WalletRechargeOrder
	orderKey := strings.TrimSpace(orderPublicID)
	if err := s.runtime.DB.WithContext(ctx).
		Where("user_id = ? AND (public_id = ? OR mch_order_no = ?)", userID, orderKey, orderKey).
		First(&order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return order, errcode.New(errcode.CodeNotFound, "recharge order not found")
		}
		return order, errcode.New(errcode.CodeInternalError, "failed to load recharge order")
	}

	return order, nil
}

// 15. loadRechargeOrderByGatewayKeys loads one order from callback keys.
func (s *WalletService) loadRechargeOrderByGatewayKeys(ctx context.Context, mchOrderNo string, payOrderID string) (model.WalletRechargeOrder, error) {
	var order model.WalletRechargeOrder
	query := s.runtime.DB.WithContext(ctx).Model(&model.WalletRechargeOrder{})
	if strings.TrimSpace(mchOrderNo) != "" {
		query = query.Where("mch_order_no = ?", strings.TrimSpace(mchOrderNo))
	} else if strings.TrimSpace(payOrderID) != "" {
		query = query.Where("pay_order_id = ?", strings.TrimSpace(payOrderID))
	} else {
		return order, errcode.New(errcode.CodeValidationError, "payment notify order key is missing")
	}
	if err := query.First(&order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return order, errcode.New(errcode.CodeNotFound, "recharge order not found")
		}
		return order, errcode.New(errcode.CodeInternalError, "failed to load recharge order")
	}

	return order, nil
}

// 16. toRechargeResponse maps a recharge order to public payload.
func (s *WalletService) toRechargeResponse(order model.WalletRechargeOrder) *WalletRechargeResponse {
	return &WalletRechargeResponse{
		OrderID:          order.PublicID,
		MchOrderNo:       order.MchOrderNo,
		PayOrderID:       order.PayOrderID,
		PayChannel:       order.PayChannel,
		PayRegion:        order.PayRegion,
		PayDataType:      order.PayDataType,
		PayData:          order.PayData,
		State:            order.State,
		StateLabel:       rechargeStateLabel(order.State),
		GatewayStateCode: order.GatewayStateCode,
		GatewayMessage:   order.GatewayMessage,
		Currency:         order.Currency,
		AmountHKD:        formatRechargeHKD(order.AmountCents),
		AmountCents:      order.AmountCents,
		PointsAmount:     order.PointsAmount,
		CreatedAt:        order.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        order.UpdatedAt.Format(time.RFC3339),
		ExpireTime:       formatPaymentOptionalTime(order.ExpireTime),
		PaidAt:           formatPaymentOptionalTime(order.PaidAt),
		CreditedAt:       formatPaymentOptionalTime(order.CreditedAt),
	}
}

// 17. buildRechargeReturnURL builds the member-facing return URL.
func (s *WalletService) buildRechargeReturnURL(returnPath string, mchOrderNo string, payChannel string) (string, error) {
	var target *url.URL
	var err error
	if strings.HasPrefix(strings.TrimSpace(returnPath), "http://") || strings.HasPrefix(strings.TrimSpace(returnPath), "https://") {
		target, err = url.Parse(returnPath)
	} else {
		target, err = url.Parse(strings.TrimSpace(s.runtime.Config.PaymentReturnBaseURL))
		if err == nil {
			target, err = target.Parse(paymentFirstNonEmpty(returnPath, "/account/profile/wallet"))
		}
	}
	if err != nil {
		return "", err
	}

	query := target.Query()
	query.Set("mch_order_no", mchOrderNo)
	query.Set("pay_channel", payChannel)
	query.Set("wallet_recharge", "1")
	target.RawQuery = query.Encode()
	return target.String(), nil
}

// 18. createRechargeMchOrderNo creates a readable merchant order number.
func (s *WalletService) createRechargeMchOrderNo() string {
	return fmt.Sprintf("AJO_%d_%s", s.runtime.Now().UnixMilli(), strings.ToUpper(utils.NewNumericCode(6)))
}

// 19. normalizedPaymentExpireSeconds returns a safe gateway expiry.
func (s *WalletService) normalizedPaymentExpireSeconds() int {
	if s.runtime.Config.PaymentExpireSeconds <= 0 {
		return 180
	}
	return s.runtime.Config.PaymentExpireSeconds
}

// 20. verifyPaymentGatewayResponse verifies success and gateway signature.
func verifyPaymentGatewayResponse(response map[string]any, appSecret string) error {
	if response == nil {
		return fmt.Errorf("payment gateway response is empty")
	}
	if paymentIntValue(response["code"]) != 0 {
		return errors.New(paymentFirstNonEmpty(paymentStringValue(response["msg"]), "payment gateway returned failure"))
	}
	data := paymentMapValue(response["data"])
	sign := paymentStringValue(response["sign"])
	if len(data) > 0 && !paymentVerifySignature(data, sign, appSecret) {
		return fmt.Errorf("payment gateway signature verification failed")
	}

	return nil
}

// 21. parsePaymentNotifyBody parses JSON or form callbacks.
func parsePaymentNotifyBody(contentType string, rawBody string) (map[string]any, error) {
	if strings.TrimSpace(rawBody) == "" {
		return map[string]any{}, nil
	}
	if strings.Contains(contentType, "application/x-www-form-urlencoded") {
		values, err := url.ParseQuery(rawBody)
		if err != nil {
			return nil, err
		}
		result := make(map[string]any, len(values))
		for key, items := range values {
			if len(items) > 0 {
				result[key] = items[len(items)-1]
			}
		}
		return result, nil
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(rawBody), &result); err != nil {
		return nil, err
	}
	if result == nil {
		return map[string]any{}, nil
	}
	return result, nil
}

// 22. mustMarshalJSON marshals JSON-compatible snapshots for audit fields.
func mustMarshalJSON(value any) []byte {
	raw, err := json.Marshal(value)
	if err != nil {
		return []byte("{}")
	}
	return raw
}

// 23. optionalPaymentString returns nil for empty strings.
func optionalPaymentString(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}

// 24. deletePaymentEmptyValues removes empty fields from gateway payload.
func deletePaymentEmptyValues(payload map[string]any) {
	for key, value := range payload {
		if paymentIsEmptyValue(value) {
			delete(payload, key)
		}
	}
}

// 25. shouldSendRechargeReturnURL omits returnUrl for QR channels.
func shouldSendRechargeReturnURL(payChannel string) bool {
	return !strings.HasSuffix(strings.ToUpper(strings.TrimSpace(payChannel)), "_QR")
}

// 26. normalizeRechargeAmountCents converts an HKD amount to cents.
func normalizeRechargeAmountCents(amountHKD float64) int64 {
	if math.IsNaN(amountHKD) || math.IsInf(amountHKD, 0) {
		return 0
	}

	return int64(math.Round(amountHKD * 100))
}

// 27. formatRechargeHKD formats cents as a stable decimal HKD string.
func formatRechargeHKD(amountCents int64) string {
	dollars := amountCents / 100
	cents := amountCents % 100
	if cents < 0 {
		cents = -cents
	}

	return fmt.Sprintf("%d.%02d", dollars, cents)
}

// 28. rechargeStateLabel maps local state to visible text.
func rechargeStateLabel(state string) string {
	switch state {
	case PaymentStateSuccess:
		return "支付成功"
	case PaymentStateFailed:
		return "支付失敗"
	case PaymentStateClosed:
		return "訂單已關閉"
	case PaymentStateExpired:
		return "訂單已過期"
	case PaymentStateRevoked:
		return "訂單已撤銷"
	case PaymentStateRefunded:
		return "訂單已退款"
	default:
		return "支付確認中"
	}
}

// 29. formatPaymentOptionalTime formats nullable payment timestamps.
func formatPaymentOptionalTime(value *time.Time) string {
	if value == nil {
		return ""
	}

	return value.Format(time.RFC3339)
}

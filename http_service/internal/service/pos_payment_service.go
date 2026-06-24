/*
 * POS 物業繳費服務。
 * 1. 依會員資料解析 POS 大廈、樓層與單位。
 * 2. 透過 AJO 後端代理 POS 賬單、歷史、會計與 H5 支付接口。
 * 3. 保持舊系統資料在 AJO 權限模型內受控讀取。
 */
package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"gorm.io/gorm"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/model"
	"ajoliving_web/http_service/internal/utils"
)

const (
	posPaymentDefaultCurrency      = "HKD"
	posPaymentDefaultExpireSeconds = 180
	posPaymentRewardPointRate      = int64(1)
)

// 1. POSPaymentService handles POS property payment data.
type POSPaymentService struct {
	runtime *Runtime
}

// 2. POSUnitContext defines the selected POS unit derived from member profile.
type POSUnitContext struct {
	BuildingID   string `json:"building_id"`
	BuildingName string `json:"building_name"`
	UnitID       string `json:"unit_id"`
	Floor        string `json:"floor"`
	Unit         string `json:"unit"`
	UnitLabel    string `json:"unit_label"`
}

// 3. POSPaymentOverview defines the payment hub summary.
type POSPaymentOverview struct {
	Context              *POSUnitContext            `json:"context,omitempty"`
	BuildingOptions      []string                   `json:"building_options"`
	UnitOptions          []string                   `json:"unit_options"`
	IsStaff              bool                       `json:"is_staff"`
	ProfileRequired      bool                       `json:"profile_required"`
	POSLoginRequired     bool                       `json:"pos_login_required"`
	TerminalProxyEnabled bool                       `json:"terminal_proxy_enabled"`
	Summary              *POSPaymentOverviewSummary `json:"summary,omitempty"`
}

// 4. POSPaymentOverviewSummary defines aggregated payment hub indicators.
type POSPaymentOverviewSummary struct {
	PendingBillCount   int     `json:"pending_bill_count"`
	PendingBillAmount  float64 `json:"pending_bill_amount"`
	OrderCount         int     `json:"order_count"`
	PendingOrderCount  int     `json:"pending_order_count"`
	AbnormalOrderCount int     `json:"abnormal_order_count"`
}

// 5. POSPaymentListResponse defines a generic POS row list.
type POSPaymentListResponse struct {
	Context         *POSUnitContext  `json:"context,omitempty"`
	BuildingOptions []string         `json:"building_options"`
	UnitOptions     []string         `json:"unit_options"`
	Items           []map[string]any `json:"items"`
}

// 6. POSPaymentSelection defines a selected POS payment context.
type POSPaymentSelection struct {
	BuildingID string
	UnitID     string
}

// 7. POSPaymentOrderCreateParams defines H5 order creation input.
type POSPaymentOrderCreateParams struct {
	UserID                  int64
	Selection               POSPaymentSelection
	Scene                   string
	PayChannel              string
	ExpireSeconds           int64
	FinalAmount             int64
	HandleFeeAmount         int64
	BillObjs                []map[string]any
	HandleFeeObj            []map[string]any
	ReturnPath              string
	Remark                  string
	GatewayRequestOverrides map[string]any
	ClientIP                string
}

// 8. NewPOSPaymentService creates a POS payment service instance.
func NewPOSPaymentService(runtime *Runtime) *POSPaymentService {
	return &POSPaymentService{runtime: runtime}
}

// 9. Overview returns member POS payment readiness.
func (s *POSPaymentService) Overview(ctx context.Context, userID int64, selection POSPaymentSelection, includeSummary bool) (*POSPaymentOverview, error) {
	account, accountErr := s.loadIsmartAccount(ctx, userID)
	contextValue, contextErr := s.resolveMemberUnitContext(ctx, userID, selection)
	buildingOptions, unitOptions, _ := s.memberPaymentBindings(ctx, userID, account)
	var appErr *errcode.AppError
	profileRequired := false
	if contextErr != nil {
		if !errors.As(contextErr, &appErr) {
			return nil, contextErr
		}
		if appErr.Code != errcode.CodeValidationError {
			return nil, contextErr
		}
		profileRequired = true
	}
	if accountErr != nil {
		if !errors.As(accountErr, &appErr) || appErr.Code != errcode.CodeAuthRequired {
			return nil, accountErr
		}
	}

	isStaff := account != nil && account.IsStaff
	var summary *POSPaymentOverviewSummary
	if includeSummary {
		summary = s.overviewSummary(ctx, userID, contextValue, account != nil && strings.TrimSpace(account.RelayTokenEncrypted) != "")
	}
	return &POSPaymentOverview{
		Context:              contextValue,
		BuildingOptions:      buildingOptions,
		UnitOptions:          unitOptions,
		IsStaff:              isStaff,
		ProfileRequired:      profileRequired,
		POSLoginRequired:     account == nil || strings.TrimSpace(account.RelayTokenEncrypted) == "",
		TerminalProxyEnabled: isStaff && contextValue != nil && s.TerminalProxyEnabled(contextValue.BuildingID),
		Summary:              summary,
	}, nil
}

// 10. overviewSummary returns non-blocking payment hub indicators.
func (s *POSPaymentService) overviewSummary(ctx context.Context, userID int64, contextValue *POSUnitContext, canReadPOS bool) *POSPaymentOverviewSummary {
	if contextValue == nil || !canReadPOS {
		return nil
	}

	summary := &POSPaymentOverviewSummary{}
	token, err := s.loadRelayToken(ctx, userID)
	if err != nil {
		return summary
	}

	var bills []map[string]any
	if err := s.posRelayJSONWithRefresh(ctx, userID, token, http.MethodGet, "/unit/"+url.PathEscape(contextValue.UnitID)+"/bills", nil, nil, &bills); err == nil {
		summary.PendingBillCount = len(bills)
		for _, row := range bills {
			summary.PendingBillAmount += posSummaryAmount(row)
		}
	}

	query := url.Values{}
	query.Set("unit_id", contextValue.UnitID)
	query.Set("limit", "80")
	var orderResult map[string]any
	if err := s.posPaymentServiceJSON(ctx, http.MethodGet, "/order-records?"+query.Encode(), nil, &orderResult); err == nil {
		orders := normalizePaymentServiceRows(orderResult)
		summary.OrderCount = len(orders)
		for _, row := range orders {
			if posSummaryOrderIsPending(row) {
				summary.PendingOrderCount++
			}
			if posSummaryOrderIsAbnormal(row) {
				summary.AbnormalOrderCount++
			}
		}
	}

	return summary
}

// 11. posSummaryAmount reads an amount-like POS field.
func posSummaryAmount(row map[string]any) float64 {
	for _, key := range []string{"paid_amount", "net_amount", "amount", "final_amount", "total", "total_amount", "payable"} {
		raw := strings.TrimSpace(paymentStringValue(row[key]))
		if raw == "" {
			continue
		}
		value, err := strconv.ParseFloat(raw, 64)
		if err == nil && value > 0 {
			return value
		}
	}
	return 0
}

// 12. posSummaryOrderIsPending checks pending order state.
func posSummaryOrderIsPending(row map[string]any) bool {
	state := strings.ToUpper(paymentFirstNonEmpty(
		paymentStringValue(row["state"]),
		paymentStringValue(row["status"]),
		paymentStringValue(row["gateway_state_code"]),
	))
	businessState := strings.ToUpper(paymentStringValue(row["business_state"]))
	return state == "PAYING" || state == "PENDING" || state == "0" || businessState == "PENDING"
}

// 13. posSummaryOrderIsAbnormal checks failed or expired order state.
func posSummaryOrderIsAbnormal(row map[string]any) bool {
	state := strings.ToUpper(paymentFirstNonEmpty(
		paymentStringValue(row["state"]),
		paymentStringValue(row["status"]),
	))
	businessState := strings.ToUpper(paymentStringValue(row["business_state"]))
	return state == "FAILED" || state == "EXPIRED" || state == "REVOKED" || businessState == "FAILED"
}

// 14. ListBills returns selected unit unpaid POS bills.
func (s *POSPaymentService) ListBills(ctx context.Context, userID int64, selection POSPaymentSelection) (*POSPaymentListResponse, error) {
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
	if err := s.posRelayJSONWithRefresh(ctx, userID, token, http.MethodGet, "/unit/"+url.PathEscape(contextValue.UnitID)+"/bills", query, nil, &rows); err != nil {
		return nil, err
	}

	buildingOptions, unitOptions, _ := s.memberPaymentBindings(ctx, userID, nil)
	return &POSPaymentListResponse{Context: contextValue, BuildingOptions: buildingOptions, UnitOptions: unitOptions, Items: rows}, nil
}

// 10. ListHistory returns POS transaction history for the selected unit.
func (s *POSPaymentService) ListHistory(ctx context.Context, userID int64, selection POSPaymentSelection) (*POSPaymentListResponse, error) {
	contextValue, token, err := s.memberPOSAccess(ctx, userID, selection)
	if err != nil {
		return nil, err
	}

	payload := map[string]any{"unit_id_list": []string{contextValue.UnitID}}
	var response map[string]any
	if err := s.posRelayJSONWithRefresh(ctx, userID, token, http.MethodPost, "/transactions/flat_units", nil, payload, &response); err != nil {
		return nil, err
	}

	buildingOptions, unitOptions, _ := s.memberPaymentBindings(ctx, userID, nil)
	return &POSPaymentListResponse{Context: contextValue, BuildingOptions: buildingOptions, UnitOptions: unitOptions, Items: normalizePOSRows(response, "payment_objs")}, nil
}

// 11. ListAccounting returns staff cashier entries for the selected building.
func (s *POSPaymentService) ListAccounting(ctx context.Context, userID int64, selection POSPaymentSelection) (*POSPaymentListResponse, error) {
	contextValue, token, err := s.staffBuildingAccess(ctx, userID, selection)
	if err != nil {
		return nil, err
	}

	var response map[string]any
	path := "/building/" + url.PathEscape(contextValue.BuildingID) + "/accounting"
	if err := s.posRelayJSONWithRefresh(ctx, userID, token, http.MethodGet, path, nil, nil, &response); err != nil {
		return nil, err
	}

	buildingOptions, unitOptions, _ := s.memberPaymentBindings(ctx, userID, nil)
	return &POSPaymentListResponse{Context: contextValue, BuildingOptions: buildingOptions, UnitOptions: unitOptions, Items: normalizePOSAccountingRows(response)}, nil
}

// 12. CreateH5Order creates one POS H5 payment order.
func (s *POSPaymentService) CreateH5Order(ctx context.Context, params POSPaymentOrderCreateParams) (map[string]any, error) {
	contextValue, _, err := s.memberPOSAccess(ctx, params.UserID, params.Selection)
	if err != nil {
		return nil, err
	}
	account, err := s.loadIsmartAccount(ctx, params.UserID)
	if err != nil {
		return nil, err
	}
	if params.FinalAmount <= 0 || strings.TrimSpace(params.PayChannel) == "" || len(params.BillObjs) == 0 {
		return nil, errcode.New(errcode.CodeValidationError, "valid payment order payload is required")
	}

	payload := s.buildH5OrderPayload(contextValue, account, params)
	var result map[string]any
	if err := s.posPaymentServiceJSON(ctx, http.MethodPost, "/h5/orders", payload, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// 13. GetH5Order returns one POS H5 payment order.
func (s *POSPaymentService) GetH5Order(ctx context.Context, userID int64, mchOrderNo string, detail bool, refresh bool, selection POSPaymentSelection) (map[string]any, error) {
	contextValue, _, err := s.memberPOSAccess(ctx, userID, selection)
	if err != nil {
		return nil, err
	}

	path := "/h5/orders/" + url.PathEscape(strings.TrimSpace(mchOrderNo))
	if detail {
		path += "/detail"
	}
	query := url.Values{}
	if refresh {
		query.Set("refresh", "1")
		query.Set("retry_business", "1")
	}

	var result map[string]any
	if err := s.posPaymentServiceJSON(ctx, http.MethodGet, path+"?"+query.Encode(), nil, &result); err != nil {
		return nil, err
	}
	if !h5OrderMatchesContext(result, contextValue) {
		if h5OrderHasContext(result) {
			return nil, errcode.New(errcode.CodeAuthForbidden, "payment order is not visible")
		}
		if err := s.ensureH5OrderVisibleByDetail(ctx, mchOrderNo, contextValue, refresh); err != nil {
			return nil, err
		}
	}
	if err := s.creditPOSPaymentReward(ctx, userID, result); err != nil {
		return nil, err
	}

	return result, nil
}

// 14. ListOrderRecords returns POS payment order snapshots for the selected unit.
func (s *POSPaymentService) ListOrderRecords(ctx context.Context, userID int64, selection POSPaymentSelection) (*POSPaymentListResponse, error) {
	contextValue, _, err := s.memberPOSAccess(ctx, userID, selection)
	if err != nil {
		return nil, err
	}

	query := url.Values{}
	query.Set("unit_id", contextValue.UnitID)
	query.Set("limit", "80")

	var result map[string]any
	if err := s.posPaymentServiceJSON(ctx, http.MethodGet, "/order-records?"+query.Encode(), nil, &result); err != nil {
		return nil, err
	}

	buildingOptions, unitOptions, _ := s.memberPaymentBindings(ctx, userID, nil)
	return &POSPaymentListResponse{Context: contextValue, BuildingOptions: buildingOptions, UnitOptions: unitOptions, Items: normalizePaymentServiceRows(result)}, nil
}

// 15. memberPOSAccess returns selected unit context and POS relay token.
func (s *POSPaymentService) memberPOSAccess(ctx context.Context, userID int64, selection POSPaymentSelection) (*POSUnitContext, string, error) {
	contextValue, err := s.resolveMemberUnitContext(ctx, userID, selection)
	if err != nil {
		return nil, "", err
	}

	token, err := s.loadRelayToken(ctx, userID)
	if err != nil {
		return nil, "", err
	}

	return contextValue, token, nil
}

// 16. staffPOSAccess returns staff context and token.
func (s *POSPaymentService) staffPOSAccess(ctx context.Context, userID int64, selection POSPaymentSelection) (*POSUnitContext, string, error) {
	account, err := s.loadIsmartAccount(ctx, userID)
	if err != nil {
		return nil, "", err
	}
	if !account.IsStaff {
		return nil, "", errcode.New(errcode.CodeAuthForbidden, "staff access required")
	}

	return s.memberPOSAccess(ctx, userID, selection)
}

// 17. staffBuildingAccess returns staff building context and POS relay token.
func (s *POSPaymentService) staffBuildingAccess(ctx context.Context, userID int64, selection POSPaymentSelection) (*POSUnitContext, string, error) {
	account, err := s.loadIsmartAccount(ctx, userID)
	if err != nil {
		return nil, "", err
	}
	if !account.IsStaff {
		return nil, "", errcode.New(errcode.CodeAuthForbidden, "staff access required")
	}

	buildingOptions, _, err := s.memberPaymentBindings(ctx, userID, account)
	if err != nil {
		return nil, "", err
	}
	buildingID := strings.TrimSpace(selection.BuildingID)
	if buildingID == "" && len(buildingOptions) > 0 {
		buildingID = buildingOptions[0]
	}
	if !containsString(buildingOptions, buildingID) {
		return nil, "", errcode.New(errcode.CodeAuthForbidden, "building is not visible")
	}
	token, err := s.loadRelayToken(ctx, userID)
	if err != nil {
		return nil, "", err
	}

	return &POSUnitContext{
		BuildingID:   buildingID,
		BuildingName: buildingID,
	}, token, nil
}

// 18. resolveMemberUnitContext resolves POS unit from the member profile.
func (s *POSPaymentService) resolveMemberUnitContext(ctx context.Context, userID int64, selection POSPaymentSelection) (*POSUnitContext, error) {
	var profile model.UserProfile
	if err := s.runtime.DB.WithContext(ctx).Preload("PrimaryCommunity").Where("user_id = ?", userID).First(&profile).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.New(errcode.CodeValidationError, "member profile is required")
		}
		return nil, errcode.New(errcode.CodeInternalError, "failed to load member profile")
	}
	account, _ := s.loadIsmartAccount(ctx, userID)
	buildingOptions, unitOptions, err := s.memberPaymentBindings(ctx, userID, account)
	if err != nil {
		return nil, err
	}
	isStaff := account != nil && account.IsStaff

	buildingID := strings.TrimSpace(selection.BuildingID)
	profileBuildingID := ""
	if profile.PrimaryCommunity != nil {
		profileBuildingID = strings.TrimSpace(profile.PrimaryCommunity.PublicID)
	}
	if buildingID == "" && containsString(buildingOptions, profileBuildingID) {
		buildingID = profileBuildingID
	}
	unitID := strings.TrimSpace(selection.UnitID)
	if buildingID == "" || (unitID == "" && strings.TrimSpace(profile.ResidenceUnit) == "") {
		return nil, errcode.New(errcode.CodeValidationError, "building and residence unit are required")
	}
	if !containsString(buildingOptions, buildingID) {
		return nil, errcode.New(errcode.CodeAuthForbidden, "building is not visible")
	}
	if unitID != "" && len(unitOptions) > 0 && !containsString(unitOptions, unitID) && !isStaff {
		return nil, errcode.New(errcode.CodeAuthForbidden, "unit is not visible")
	}

	units, err := NewPOSBuildingService(s.runtime).ListUnits(ctx, buildingID)
	if err != nil {
		return nil, err
	}

	floor := strings.TrimSpace(profile.ResidenceFloor)
	unitName := strings.TrimSpace(profile.ResidenceUnit)
	for _, item := range units {
		itemUnitID := posUnitID(item)
		if unitID != "" && itemUnitID != unitID {
			continue
		}
		if unitID == "" && !posUnitMatchesProfile(item, floor, unitName) {
			continue
		}
		if unitID != "" && len(unitOptions) == 0 && !isStaff && !posUnitMatchesProfile(item, floor, unitName) {
			return nil, errcode.New(errcode.CodeAuthForbidden, "unit is not visible")
		}

		selectedUnitID := itemUnitID
		if selectedUnitID == "" {
			continue
		}
		if !isStaff && len(unitOptions) > 0 && !containsString(unitOptions, selectedUnitID) {
			return nil, errcode.New(errcode.CodeAuthForbidden, "unit is not visible")
		}
		selectedFloor := paymentFirstNonEmpty(posUnitFloor(item), floor)
		selectedUnit := paymentFirstNonEmpty(posUnitName(item), unitName)

		return &POSUnitContext{
			BuildingID:   buildingID,
			BuildingName: posPaymentBuildingName(profile.PrimaryCommunity, buildingID),
			UnitID:       selectedUnitID,
			Floor:        selectedFloor,
			Unit:         selectedUnit,
			UnitLabel:    posUnitLabel(selectedFloor, selectedUnit),
		}, nil
	}

	return nil, errcode.New(errcode.CodeValidationError, "selected residence unit is not available")
}

// 19. memberPaymentBindings returns the visible building and unit arrays.
func (s *POSPaymentService) memberPaymentBindings(ctx context.Context, userID int64, account *model.UserIsmartAccount) ([]string, []string, error) {
	var profile model.UserProfile
	err := s.runtime.DB.WithContext(ctx).Preload("PrimaryCommunity").Where("user_id = ?", userID).First(&profile).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, errcode.New(errcode.CodeInternalError, "failed to load member profile")
	}
	buildings := normalizeStringSlice(unmarshalStringSlice(profile.BoundBuildingIDs))
	units := normalizeStringSlice(unmarshalStringSlice(profile.BoundFlatUnitIDs))
	if account != nil {
		message := &IsmartMessage{
			IsStaff:                            account.IsStaff,
			Building:                           unmarshalStringSlice(account.Building),
			StaffBuildingPermissions:           unmarshalStringSlice(account.StaffBuildingPermissions),
			ClientBuildingPermissions:          unmarshalStringSlice(account.ClientBuildingPermissions),
			ClientBuildingFlatUnitsPermissions: unmarshalStringSlice(account.ClientBuildingFlatUnitsPermissions),
		}
		buildings = resolveIsmartBoundBuildings(message)
		units = resolveIsmartBoundUnits(message)
	}
	if account == nil && len(buildings) == 0 && profile.PrimaryCommunity != nil && strings.TrimSpace(profile.PrimaryCommunity.PublicID) != "" {
		buildings = []string{strings.TrimSpace(profile.PrimaryCommunity.PublicID)}
	}

	return buildings, units, nil
}

// 20. loadRelayToken decrypts the linked POS relay token.
func (s *POSPaymentService) loadRelayToken(ctx context.Context, userID int64) (string, error) {
	account, err := s.loadIsmartAccount(ctx, userID)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(account.RelayTokenEncrypted) == "" {
		return "", errcode.New(errcode.CodeAuthRequired, "ismart login is required")
	}

	token, err := utils.DecryptString(s.runtime.Config.EncryptionKey, account.RelayTokenEncrypted)
	if err != nil || strings.TrimSpace(token) == "" {
		return "", errcode.New(errcode.CodeAuthRequired, "ismart login is required")
	}

	return strings.TrimSpace(token), nil
}

// 21. posRelayJSONWithRefresh retries POS relay calls once after token refresh.
func (s *POSPaymentService) posRelayJSONWithRefresh(ctx context.Context, userID int64, token string, method string, path string, query url.Values, payload any, target any) error {
	err := s.posRelayJSON(ctx, method, path, token, query, payload, target)
	if err == nil {
		return nil
	}
	if !isPOSAuthRequiredError(err) {
		return err
	}

	refreshedToken, refreshErr := s.refreshRelayToken(ctx, userID)
	if refreshErr != nil {
		return refreshErr
	}

	return s.posRelayJSON(ctx, method, path, refreshedToken, query, payload, target)
}

// 22. refreshRelayToken uses stored ismart credentials to renew the POS relay token.
func (s *POSPaymentService) refreshRelayToken(ctx context.Context, userID int64) (string, error) {
	account, err := s.loadIsmartAccount(ctx, userID)
	if err != nil {
		return "", err
	}
	password, err := utils.DecryptString(s.runtime.Config.EncryptionKey, account.PasswordEncrypted)
	if err != nil || strings.TrimSpace(password) == "" {
		return "", errcode.New(errcode.CodeAuthRequired, "ismart login is required")
	}

	authService := NewAuthService(s.runtime)
	var lastErr error
	for _, accountName := range buildIsmartLoginAttempts(account.Username, account.Phone, strings.TrimPrefix(account.Phone, "+852")) {
		token, loginErr := authService.callPOSRelayLogin(ctx, accountName, password)
		if loginErr == nil {
			encrypted, encryptErr := utils.EncryptString(s.runtime.Config.EncryptionKey, token)
			if encryptErr != nil {
				return "", errcode.New(errcode.CodeInternalError, "failed to prepare pos token")
			}
			now := s.runtime.Now()
			if err := s.runtime.DB.WithContext(ctx).
				Model(&model.UserIsmartAccount{}).
				Where("user_id = ?", userID).
				Updates(map[string]any{
					"relay_token_encrypted": encrypted,
					"relay_token_synced_at": &now,
				}).Error; err != nil {
				return "", errcode.New(errcode.CodeInternalError, "failed to update pos token")
			}
			return token, nil
		}
		lastErr = loginErr
		if !isIsmartCredentialError(loginErr) {
			break
		}
	}
	if lastErr != nil {
		return "", lastErr
	}

	return "", errcode.New(errcode.CodeAuthRequired, "ismart login is required")
}

// 23. loadIsmartAccount loads the linked POS account.
func (s *POSPaymentService) loadIsmartAccount(ctx context.Context, userID int64) (*model.UserIsmartAccount, error) {
	var account model.UserIsmartAccount
	if err := s.runtime.DB.WithContext(ctx).Where("user_id = ?", userID).First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.New(errcode.CodeAuthRequired, "ismart login is required")
		}
		return nil, errcode.New(errcode.CodeInternalError, "failed to load ismart account")
	}

	return &account, nil
}

// 24. buildH5OrderPayload creates the POS payment service request body.
func (s *POSPaymentService) buildH5OrderPayload(contextValue *POSUnitContext, account *model.UserIsmartAccount, params POSPaymentOrderCreateParams) map[string]any {
	returnPath := strings.TrimSpace(params.ReturnPath)
	if returnPath == "" {
		returnPath = "/payments/orders"
	}
	scene := normalizePOSPaymentScene(params.Scene)
	expireSeconds := params.ExpireSeconds
	if expireSeconds <= 0 {
		expireSeconds = posPaymentDefaultExpireSeconds
	}
	gatewayOverrides := sanitizePOSGatewayRequestOverrides(params.PayChannel, params.GatewayRequestOverrides)
	gatewayOverrides["clientIp"] = strings.TrimSpace(params.ClientIP)
	reportPaymentData := map[string]any{
		"BLG_ID":                contextValue.BuildingID,
		"UNIT_ID":               contextValue.UnitID,
		"UNIT_LABEL":            contextValue.UnitLabel,
		"BILL_OBJ":              params.BillObjs,
		"HANDLE_FEE_OBJ":        params.HandleFeeObj,
		"COMMENT":               strings.TrimSpace(params.Remark),
		"FINAL_AMOUNT":          params.FinalAmount,
		"CURRENCY":              posPaymentDefaultCurrency,
		"bank_account_received": "",
		"USER_ID":               account.IsmartUserID,
	}

	return map[string]any{
		"scene":               scene,
		"pay_channel":         strings.TrimSpace(params.PayChannel),
		"currency":            posPaymentDefaultCurrency,
		"expire_seconds":      expireSeconds,
		"blg_id":              contextValue.BuildingID,
		"unit_id":             contextValue.UnitID,
		"final_amount":        params.FinalAmount,
		"handle_fee_amount":   params.HandleFeeAmount,
		"bill_objs":           params.BillObjs,
		"handle_fee_obj":      params.HandleFeeObj,
		"return_path":         returnPath,
		"remark":              strings.TrimSpace(params.Remark),
		"report_payment_data": reportPaymentData,
		"operator": map[string]any{
			"login_name": account.Username,
			"user_id":    account.IsmartUserID,
		},
		"gateway_request_overrides": gatewayOverrides,
	}
}

// 25. normalizePOSPaymentScene keeps old H5 order scene semantics.
func normalizePOSPaymentScene(scene string) string {
	if strings.EqualFold(strings.TrimSpace(scene), "billing") {
		return "billing"
	}
	return "cart"
}

// 26. sanitizePOSGatewayRequestOverrides keeps only supported gateway overrides.
func sanitizePOSGatewayRequestOverrides(payChannel string, overrides map[string]any) map[string]any {
	result := map[string]any{}
	if !strings.EqualFold(strings.TrimSpace(payChannel), "ALI_H5") {
		return result
	}

	walletType := strings.ToUpper(strings.TrimSpace(paymentStringValue(overrides["walletType"])))
	if walletType == "CN" || walletType == "HK" {
		result["walletType"] = walletType
	}
	return result
}

// 27. creditPOSPaymentReward credits AJO Points once when an H5 order is paid.
func (s *POSPaymentService) creditPOSPaymentReward(ctx context.Context, userID int64, order map[string]any) error {
	if s.runtime.WalletService == nil || !posH5OrderPaid(order) {
		return nil
	}
	orderNo := posH5OrderNo(order)
	if orderNo == "" {
		return nil
	}
	points := posPaymentRewardPoints(order)
	if points <= 0 {
		return nil
	}

	return s.runtime.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		_, err := s.runtime.WalletService.CreditPointsWithTx(ctx, tx, WalletCreditParams{
			UserID:         userID,
			Amount:         points,
			SourceType:     WalletSourcePOSPayment,
			BizModule:      "payment",
			ActionType:     WalletActionPOSReward,
			IdempotencyKey: "pos_payment_reward:" + orderNo,
			Note:           fmt.Sprintf("POS payment reward %s", orderNo),
		})
		return err
	})
}

// 28. posH5OrderPaid checks whether a POS H5 order has reached paid state.
func posH5OrderPaid(order map[string]any) bool {
	state := strings.ToLower(paymentFirstNonEmpty(
		paymentStringValue(order["state"]),
		paymentStringValue(order["status"]),
		paymentStringValue(order["business_state"]),
		paymentStringValue(order["gateway_state"]),
		paymentStringValue(order["gateway_state_code"]),
	))
	return state == "success" || state == "succeeded" || state == "paid" || state == "2"
}

// 29. posH5OrderNo reads the merchant order number.
func posH5OrderNo(order map[string]any) string {
	return paymentFirstNonEmpty(
		paymentStringValue(order["mch_order_no"]),
		paymentStringValue(order["mchOrderNo"]),
		paymentStringValue(order["order_no"]),
		paymentStringValue(order["pay_order_id"]),
		paymentStringValue(order["payOrderId"]),
	)
}

// 30. posPaymentRewardPoints converts paid HKD amount into reward points.
func posPaymentRewardPoints(order map[string]any) int64 {
	for _, key := range []string{"amount_hkd", "paid_amount_hkd"} {
		value, err := strconv.ParseFloat(strings.TrimSpace(paymentStringValue(order[key])), 64)
		if err == nil && value > 0 {
			return int64(value) * posPaymentRewardPointRate
		}
	}
	for _, key := range []string{"final_amount", "amount", "paid_amount", "total_amount"} {
		value, err := strconv.ParseInt(strings.TrimSpace(paymentStringValue(order[key])), 10, 64)
		if err == nil && value > 0 {
			return (value / 100) * posPaymentRewardPointRate
		}
	}
	return 0
}

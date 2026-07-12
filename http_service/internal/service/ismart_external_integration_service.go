/*
 * iSmart 新整合介面代理服務。
 * 1. 代理大廈應收、通告與會員角色綁定 integration API。
 * 2. 使用 AJO 登入態補齊 iSmart user_id，避免前端提交舊系統身份。
 * 3. 保持新舊 iSmart 路徑在讀取場景下可兼容。
 * 4. 代理支付 integration 查詢介面並保留上游 raw payload。
 */
package service

import (
	"context"
	"net/url"
	"strconv"
	"strings"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/model"
)

// 1. IsmartReceivableParams defines building receivable lookup input.
type IsmartReceivableParams struct {
	BuildingID    string
	DataStructure string
}

// 2. IsmartOwnerBindingParams defines owner binding request input.
type IsmartOwnerBindingParams struct {
	BuildingID        string
	OwnedFlat         []string
	Role              string
	OwnerNote         string
	IsReceiveEmail    *bool
	RegistrationTel   string
	RegistrationEmail string
	ClientName        string
	ClientIDCard      string
	ClientTel         string
}

// 3. IsmartSubaccountQuery defines subaccount list input.
type IsmartSubaccountQuery struct {
	UnitID string
}

// 4. IsmartSubaccountMutationParams defines grant and revoke input.
type IsmartSubaccountMutationParams struct {
	UnitID       string
	TargetUserID int64
	Remark       string
}

// 5. IsmartPaymentUnpaidInvoiceParams defines unpaid invoice lookup input.
type IsmartPaymentUnpaidInvoiceParams struct {
	UnitID string
}

// 6. IsmartPaymentTransactionsByUnitParams defines unit transaction lookup input.
type IsmartPaymentTransactionsByUnitParams struct {
	UnitIDList []string
}

// 7. IsmartPaymentTransactionsByDateParams defines date transaction lookup input.
type IsmartPaymentTransactionsByDateParams struct {
	BuildingID string
	FromDate   string
	ToDate     string
	DateType   string
	PayMethod  string
}

// 8. RegisterAccount proxies authenticated iSmart direct account registration.
func (s *IsmartExternalService) RegisterAccount(ctx context.Context, userID int64, payload map[string]any) (map[string]any, error) {
	if userID <= 0 {
		return nil, errcode.New(errcode.CodeAuthRequired, "login required")
	}
	if len(payload) == 0 {
		return nil, errcode.New(errcode.CodeValidationError, "registration payload is required")
	}

	result, err := s.postIntegration(ctx, "/auth/register/", payload)
	if err != nil {
		return nil, err
	}
	return decorateIsmartPayload(result.Payload, "", nil, result.Message, false), nil
}

// 9. ListManagementFees returns visible building management-fee receivables.
func (s *IsmartExternalService) ListManagementFees(ctx context.Context, userID int64, params IsmartReceivableParams) (map[string]any, error) {
	return s.listBuildingReceivables(ctx, userID, params, "/buildings/receivables/management-fees/", "/api/v1/building-mf-table/", "table")
}

// 10. ListOtherFees returns visible building other-fee receivables.
func (s *IsmartExternalService) ListOtherFees(ctx context.Context, userID int64, params IsmartReceivableParams) (map[string]any, error) {
	return s.listBuildingReceivables(ctx, userID, params, "/buildings/receivables/other-fees/", "/api/v1/building-of-list/", "list")
}

// 11. ListBuildingNotices returns visible building notices.
func (s *IsmartExternalService) ListBuildingNotices(ctx context.Context, userID int64, params IsmartBuildingParams) (map[string]any, error) {
	account, buildingID, buildingOptions, err := s.resolveBuildingAccess(ctx, userID, params.BuildingID)
	if err != nil {
		return nil, err
	}

	query := url.Values{}
	query.Set("building_id", buildingID)
	result, err := s.getIntegration(ctx, "/buildings/notices/", query)
	if err != nil && ismartFallbackAllowed(err, false) {
		result, err = s.postRoot(ctx, "/api/v1/building-notices/", map[string]any{"blg_id": buildingID})
	}
	if err != nil {
		return nil, err
	}

	return decorateIsmartPayload(result.Payload, buildingID, buildingOptions, result.Message, account.IsStaff), nil
}

// 12. SubmitOwnerBindingRequest submits an owner-role binding request for current iSmart user.
func (s *IsmartExternalService) SubmitOwnerBindingRequest(ctx context.Context, userID int64, params IsmartOwnerBindingParams) (map[string]any, error) {
	account, err := s.loadIsmartAccount(ctx, userID)
	if err != nil {
		return nil, err
	}
	buildingID := strings.TrimSpace(params.BuildingID)
	ownedFlat := normalizeStringSlice(params.OwnedFlat)
	if buildingID == "" || len(ownedFlat) == 0 {
		return nil, errcode.New(errcode.CodeValidationError, "building id and owned flat are required")
	}

	payload := map[string]any{
		"building_id": buildingID,
		"ownedflat":   ownedFlat,
		"user":        account.IsmartUserID,
	}
	if role := strings.TrimSpace(params.Role); role != "" {
		payload["cli_role"] = role
	}
	if note := strings.TrimSpace(params.OwnerNote); note != "" {
		payload["ownernote"] = note
	}
	if params.IsReceiveEmail != nil {
		payload["is_receive_email"] = *params.IsReceiveEmail
	}
	copyOptionalString(payload, "reg_tel", params.RegistrationTel)
	copyOptionalString(payload, "reg_email", params.RegistrationEmail)
	copyOptionalString(payload, "cli_name", params.ClientName)
	copyOptionalString(payload, "cli_id_card", params.ClientIDCard)
	copyOptionalString(payload, "cli_tel", params.ClientTel)

	result, err := s.postIntegration(ctx, "/buildings/building-flat-owner-binding-requests/", payload)
	if err != nil {
		return nil, err
	}
	return decorateIsmartPayload(result.Payload, buildingID, nil, result.Message, account.IsStaff), nil
}

// 13. ListSubaccounts lists current authorized sub users for owner-controlled units.
func (s *IsmartExternalService) ListSubaccounts(ctx context.Context, userID int64, queryParams IsmartSubaccountQuery) (map[string]any, error) {
	account, err := s.loadIsmartAccount(ctx, userID)
	if err != nil {
		return nil, err
	}
	unitID := strings.TrimSpace(queryParams.UnitID)
	if unitID != "" && !s.unitVisible(account, unitID) {
		return nil, errcode.New(errcode.CodeAuthForbidden, "unit is not visible")
	}

	query := url.Values{}
	query.Set("user_id", strconv.FormatInt(account.IsmartUserID, 10))
	if unitID != "" {
		query.Set("unit_id", unitID)
	}

	result, err := s.getIntegration(ctx, "/buildings/subaccounts/", query)
	if err != nil {
		return nil, err
	}
	return decorateIsmartPayload(result.Payload, "", nil, result.Message, account.IsStaff), nil
}

// 14. GrantSubaccount grants one authorized sub user through iSmart.
func (s *IsmartExternalService) GrantSubaccount(ctx context.Context, userID int64, params IsmartSubaccountMutationParams) (map[string]any, error) {
	return s.mutateSubaccount(ctx, userID, params, "/buildings/subaccounts/grant/")
}

// 15. RevokeSubaccount revokes one authorized sub user through iSmart.
func (s *IsmartExternalService) RevokeSubaccount(ctx context.Context, userID int64, params IsmartSubaccountMutationParams) (map[string]any, error) {
	return s.mutateSubaccount(ctx, userID, params, "/buildings/subaccounts/revoke/")
}

// 16. ListPaymentUnpaidInvoices proxies unit unpaid invoice lookup.
func (s *IsmartExternalService) ListPaymentUnpaidInvoices(ctx context.Context, params IsmartPaymentUnpaidInvoiceParams) (any, error) {
	unitID := strings.TrimSpace(params.UnitID)
	if unitID == "" {
		return nil, errcode.New(errcode.CodeValidationError, "unit_id is required")
	}

	query := url.Values{}
	query.Set("unit_id", unitID)
	result, err := s.getIntegration(ctx, "/payments/unpaid-invoices/", query)
	if err != nil {
		return nil, err
	}
	return result.Payload, nil
}

// 17. ListPaymentTransactionsByUnit proxies unit payment history lookup.
func (s *IsmartExternalService) ListPaymentTransactionsByUnit(ctx context.Context, params IsmartPaymentTransactionsByUnitParams) (any, error) {
	unitIDs := normalizeStringSlice(params.UnitIDList)
	if len(unitIDs) == 0 {
		return nil, errcode.New(errcode.CodeValidationError, "unit_id_list is required")
	}

	query := url.Values{}
	for _, unitID := range unitIDs {
		query.Add("unit_id_list", unitID)
	}
	result, err := s.getIntegration(ctx, "/payments/transactions/by-unit/", query)
	if err != nil {
		return nil, err
	}
	return normalizeIntegrationPaymentTransactionsPayload(result.Payload), nil
}

// 18. ListPaymentTransactionsByDate proxies building payment history date lookup.
func (s *IsmartExternalService) ListPaymentTransactionsByDate(ctx context.Context, params IsmartPaymentTransactionsByDateParams) (any, error) {
	buildingID := strings.TrimSpace(params.BuildingID)
	fromDate := strings.TrimSpace(params.FromDate)
	toDate := strings.TrimSpace(params.ToDate)
	dateType := strings.TrimSpace(params.DateType)
	payMethod := strings.TrimSpace(params.PayMethod)
	if buildingID == "" || fromDate == "" || toDate == "" || dateType == "" {
		return nil, errcode.New(errcode.CodeValidationError, "building_id, from_date, to_date and date_type are required")
	}
	if payMethod == "" {
		payMethod = "all"
	}

	query := url.Values{}
	query.Set("building_id", buildingID)
	query.Set("from_date", fromDate)
	query.Set("to_date", toDate)
	query.Set("date_type", dateType)
	query.Set("pay_method", payMethod)
	result, err := s.getIntegration(ctx, "/payments/transactions/by-date/", query)
	if err != nil {
		return nil, err
	}
	return normalizeIntegrationPaymentTransactionsPayload(result.Payload), nil
}

// 19. listBuildingReceivables proxies one visible building receivable endpoint.
func (s *IsmartExternalService) listBuildingReceivables(ctx context.Context, userID int64, params IsmartReceivableParams, integrationPath string, legacyPath string, defaultStructure string) (map[string]any, error) {
	account, buildingID, buildingOptions, err := s.resolveBuildingAccess(ctx, userID, params.BuildingID)
	if err != nil {
		return nil, err
	}
	dataStructure := strings.TrimSpace(params.DataStructure)
	if dataStructure == "" {
		dataStructure = defaultStructure
	}

	query := url.Values{}
	query.Set("building_id", buildingID)
	query.Set("data_structure", dataStructure)
	payload := map[string]any{"building_id": buildingID, "data_structure": dataStructure}
	result, err := s.getIntegration(ctx, integrationPath, query)
	if err != nil && ismartFallbackAllowed(err, false) {
		result, err = s.postRoot(ctx, legacyPath, payload)
	}
	if err != nil {
		return nil, err
	}

	return decorateIsmartPayload(result.Payload, buildingID, buildingOptions, result.Message, account.IsStaff), nil
}

// 20. mutateSubaccount runs one grant or revoke request.
func (s *IsmartExternalService) mutateSubaccount(ctx context.Context, userID int64, params IsmartSubaccountMutationParams, path string) (map[string]any, error) {
	account, err := s.loadIsmartAccount(ctx, userID)
	if err != nil {
		return nil, err
	}
	unitID := strings.TrimSpace(params.UnitID)
	if unitID == "" || params.TargetUserID <= 0 {
		return nil, errcode.New(errcode.CodeValidationError, "unit id and target user id are required")
	}
	if !s.unitVisible(account, unitID) {
		return nil, errcode.New(errcode.CodeAuthForbidden, "unit is not visible")
	}

	payload := map[string]any{
		"user_id":        account.IsmartUserID,
		"unit_id":        unitID,
		"target_user_id": params.TargetUserID,
	}
	if remark := strings.TrimSpace(params.Remark); remark != "" {
		payload["remark"] = remark
	}

	result, err := s.postIntegration(ctx, path, payload)
	if err != nil {
		return nil, err
	}
	return decorateIsmartPayload(result.Payload, "", nil, result.Message, account.IsStaff), nil
}

// 21. unitVisible checks current iSmart account unit permissions.
func (s *IsmartExternalService) unitVisible(account *model.UserIsmartAccount, unitID string) bool {
	units := s.visibleUnitIDs(account)
	return len(units) == 0 || containsString(units, strings.TrimSpace(unitID))
}

// 22. visibleUnitIDs resolves the iSmart unit permission list.
func (s *IsmartExternalService) visibleUnitIDs(account *model.UserIsmartAccount) []string {
	if account == nil {
		return []string{}
	}
	message := &IsmartMessage{
		ClientBuildingFlatUnitsPermissions: unmarshalStringSlice(account.ClientBuildingFlatUnitsPermissions),
	}
	return resolveIsmartBoundUnits(message)
}

// 23. copyOptionalString copies one non-empty string payload field.
func copyOptionalString(payload map[string]any, key string, value string) {
	if strings.TrimSpace(value) != "" {
		payload[key] = strings.TrimSpace(value)
	}
}

// 24. normalizeIntegrationPaymentTransactionsPayload keeps payment history in documented shape.
func normalizeIntegrationPaymentTransactionsPayload(payload any) map[string]any {
	payloadMap := paymentMapValue(payload)
	rows := normalizePOSRows(payloadMap, "payment_objs")
	if len(rows) > 0 {
		return map[string]any{"payment_objs": rows}
	}
	if rows, ok := anyToMapRows(payload); ok {
		return map[string]any{"payment_objs": rows}
	}
	return map[string]any{"payment_objs": []map[string]any{}}
}

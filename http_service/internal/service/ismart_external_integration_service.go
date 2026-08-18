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
	TargetPhone  string
	TargetEmail  string
	Remark       string
}

// 4.1 IsmartServiceCaseListParams defines the current member's service-case filters.
type IsmartServiceCaseListParams struct {
	BuildingID  string
	Status      string
	RequestType string
}

// 4.2 IsmartServiceCaseDetailParams defines one visible service-case lookup.
type IsmartServiceCaseDetailParams struct {
	BuildingID string
	CaseID     string
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

// 8. IsmartDirectRegistrationParams defines AJO-controlled iSmart account creation input.
type IsmartDirectRegistrationParams struct {
	Phone          string
	Email          string
	EngName        string
	ChiName        string
	LegalEntity    string
	IDCard         string
	Remark         string
	Gender         string
	IsReceiveEmail bool
	Password       string
}

// 9. RegisterDirectAccount creates one iSmart account during AJO registration.
func (s *IsmartExternalService) RegisterDirectAccount(ctx context.Context, params IsmartDirectRegistrationParams) (*IsmartMessage, error) {
	payload := map[string]any{
		"phone":            strings.TrimSpace(params.Phone),
		"email":            normalizeEmail(params.Email),
		"eng_name":         strings.TrimSpace(params.EngName),
		"chi_name":         strings.TrimSpace(params.ChiName),
		"legal_entity":     strings.TrimSpace(params.LegalEntity),
		"id_card":          strings.TrimSpace(params.IDCard),
		"remark":           strings.TrimSpace(params.Remark),
		"gender":           strings.TrimSpace(params.Gender),
		"is_receive_email": params.IsReceiveEmail,
		"password":         params.Password,
	}
	result, err := s.postIntegration(ctx, "/auth/register/", payload)
	if err != nil {
		return nil, err
	}
	data := paymentMapValue(result.Payload)
	message := &IsmartMessage{
		UserID:   paymentInt64Value(data["user_id"]),
		Username: strings.TrimSpace(paymentStringValue(data["username"])),
		Email:    normalizeEmail(paymentStringValue(data["email"])),
		Phone:    strings.TrimSpace(paymentStringValue(data["phone"])),
	}
	if message.UserID <= 0 || message.Username == "" {
		return nil, errcode.New(errcode.CodeInternalError, "invalid ismart registration response")
	}
	if message.Email == "" {
		message.Email = normalizeEmail(params.Email)
	}
	if message.Phone == "" {
		message.Phone = strings.TrimSpace(params.Phone)
	}
	message.RawMessage = sanitizeIsmartRawMessage(data)
	message.ProfileSnapshot = ismartRegistrationProfileSnapshot(data, message, params)
	return message, nil
}

// 10. ismartRegistrationProfileSnapshot builds the member-center-safe iSmart registration snapshot.
func ismartRegistrationProfileSnapshot(data map[string]any, message *IsmartMessage, params IsmartDirectRegistrationParams) map[string]any {
	return map[string]any{
		"account_code":    firstLegacyText(data, message.Username, "account_code", "account_no", "account_number", "username", "memberno"),
		"account_phone":   firstLegacyText(data, message.Phone, "account_phone", "memberphone", "phone", "tel"),
		"account_email":   firstLegacyText(data, message.Email, "account_email", "memberemail", "email", "billing_email"),
		"owner_name_en":   firstLegacyText(data, params.EngName, "owner_name_en", "memberengname", "eng_name", "english_name"),
		"owner_name_zh":   firstLegacyText(data, params.ChiName, "owner_name_zh", "memberchiname", "chi_name", "chinese_name"),
		"identity_number": firstLegacyText(data, params.IDCard, "identity_number", "memberid", "id_number", "hkid"),
		"legal_entity":    firstLegacyText(data, params.LegalEntity, "legal_entity", "member_legalentity", "legalentity"),
		"gender":          firstLegacyText(data, params.Gender, "gender", "membergender"),
		"birth_date":      firstLegacyText(data, "", "birth_date", "birthday", "date_of_birth", "dob"),
		"contact_name":    firstLegacyText(data, "", "contact_name", "contact_person", "contactperson"),
		"contact_phone":   firstLegacyText(data, message.Phone, "contact_phone", "contact_tel", "contactphone"),
		"billing_email":   firstLegacyText(data, message.Email, "billing_email", "bill_email", "memberemail"),
		"billing_address": firstLegacyText(data, "", "billing_address", "bill_address", "address"),
	}
}

// 11. RegisterAccount proxies authenticated iSmart direct account registration.
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

// 12. ListManagementFees returns visible building management-fee receivables.
func (s *IsmartExternalService) ListManagementFees(ctx context.Context, userID int64, params IsmartReceivableParams) (map[string]any, error) {
	return s.listBuildingReceivables(ctx, userID, params, "/buildings/receivables/management-fees/", "/api/v1/building-mf-table/", "table")
}

// 13. ListOtherFees returns visible building other-fee receivables.
func (s *IsmartExternalService) ListOtherFees(ctx context.Context, userID int64, params IsmartReceivableParams) (map[string]any, error) {
	return s.listBuildingReceivables(ctx, userID, params, "/buildings/receivables/other-fees/", "/api/v1/building-of-list/", "list")
}

// 14. ListBuildingNotices returns visible building notices.
func (s *IsmartExternalService) ListBuildingNotices(ctx context.Context, userID int64, params IsmartBuildingParams) (map[string]any, error) {
	account, buildingID, buildingOptions, err := s.resolveAuthorizedBuildingAccess(ctx, userID, params.BuildingID, BuildingPermissionBuildingNotices)
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

// 15. SubmitOwnerBindingRequest submits an owner-role binding request for current iSmart user.
func (s *IsmartExternalService) SubmitOwnerBindingRequest(ctx context.Context, userID int64, params IsmartOwnerBindingParams) (map[string]any, error) {
	account, buildingID, payload, err := s.ownerBindingPayload(ctx, userID, params)
	if err != nil {
		return nil, err
	}

	result, err := s.postIntegration(ctx, "/buildings/building-flat-owner-binding-requests/", payload)
	if err != nil {
		return nil, err
	}
	return decorateIsmartPayload(result.Payload, buildingID, nil, result.Message, account.IsStaff), nil
}

// 16. SubmitOwnerBindingRequestAccepted treats any upstream HTTP 2xx response as an accepted OwnerReg application.
func (s *IsmartExternalService) SubmitOwnerBindingRequestAccepted(ctx context.Context, userID int64, params IsmartOwnerBindingParams) error {
	_, _, payload, err := s.ownerBindingPayload(ctx, userID, params)
	if err != nil {
		return err
	}

	return s.postIntegrationAccepted(ctx, "/buildings/building-flat-owner-binding-requests/", payload)
}

// 17. ownerBindingPayload prepares one OwnerReg request from the current linked iSmart account.
func (s *IsmartExternalService) ownerBindingPayload(ctx context.Context, userID int64, params IsmartOwnerBindingParams) (*model.UserIsmartAccount, string, map[string]any, error) {
	account, err := s.loadIsmartAccount(ctx, userID)
	if err != nil {
		return nil, "", nil, err
	}
	buildingID := strings.TrimSpace(params.BuildingID)
	ownedFlat := normalizeStringSlice(params.OwnedFlat)
	if buildingID == "" || len(ownedFlat) == 0 {
		return nil, "", nil, errcode.New(errcode.CodeValidationError, "building id and owned flat are required")
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

	return account, buildingID, payload, nil
}

// 18. ListSubaccounts lists current authorized sub users for owner-controlled units.
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

// 17. GrantSubaccount grants one authorized sub user through iSmart.
func (s *IsmartExternalService) GrantSubaccount(ctx context.Context, userID int64, params IsmartSubaccountMutationParams) (map[string]any, error) {
	return s.mutateSubaccount(ctx, userID, params, "/buildings/subaccounts/grant/")
}

// 18. RevokeSubaccount revokes one authorized sub user through iSmart.
func (s *IsmartExternalService) RevokeSubaccount(ctx context.Context, userID int64, params IsmartSubaccountMutationParams) (map[string]any, error) {
	return s.mutateSubaccount(ctx, userID, params, "/buildings/subaccounts/revoke/")
}

// 18.1 ListBuildingServiceCases returns only service cases visible to the current member.
func (s *IsmartExternalService) ListBuildingServiceCases(ctx context.Context, userID int64, params IsmartServiceCaseListParams) (map[string]any, error) {
	account, buildingID, buildingOptions, err := s.resolveBuildingAccess(ctx, userID, params.BuildingID)
	if err != nil {
		return nil, err
	}

	query := url.Values{}
	query.Set("user_id", strconv.FormatInt(account.IsmartUserID, 10))
	query.Set("building_id", buildingID)
	query.Set("scope", "mine")
	copyQueryValue(query, "status", params.Status)
	copyQueryValue(query, "request_type", params.RequestType)
	result, err := s.getServiceCaseIntegration(ctx, "/buildings/service-cases/", query)
	if err != nil {
		return nil, err
	}

	return decorateIsmartPayload(result.Payload, buildingID, buildingOptions, result.Message, account.IsStaff), nil
}

// 18.2 GetBuildingServiceCase returns one current member-visible service case and message thread.
func (s *IsmartExternalService) GetBuildingServiceCase(ctx context.Context, userID int64, params IsmartServiceCaseDetailParams) (map[string]any, error) {
	account, buildingID, buildingOptions, err := s.resolveBuildingAccess(ctx, userID, params.BuildingID)
	if err != nil {
		return nil, err
	}
	caseID := strings.TrimSpace(params.CaseID)
	if caseID == "" {
		return nil, errcode.New(errcode.CodeValidationError, "service case id is required")
	}

	query := url.Values{}
	query.Set("user_id", strconv.FormatInt(account.IsmartUserID, 10))
	query.Set("building_id", buildingID)
	result, err := s.getServiceCaseIntegration(ctx, "/buildings/service-cases/"+url.PathEscape(caseID)+"/", query)
	if err != nil {
		return nil, err
	}

	return decorateIsmartPayload(result.Payload, buildingID, buildingOptions, result.Message, account.IsStaff), nil
}

// 19. ListPaymentUnpaidInvoices proxies unit unpaid invoice lookup.
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

// 20. ListPaymentTransactionsByUnit proxies unit payment history lookup.
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

// 21. ListPaymentTransactionsByDate proxies building payment history date lookup.
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

// 22. listBuildingReceivables proxies one visible building receivable endpoint.
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

// 23. mutateSubaccount runs one grant or revoke request.
func (s *IsmartExternalService) mutateSubaccount(ctx context.Context, userID int64, params IsmartSubaccountMutationParams, path string) (map[string]any, error) {
	account, err := s.loadIsmartAccount(ctx, userID)
	if err != nil {
		return nil, err
	}
	unitID := strings.TrimSpace(params.UnitID)
	if unitID == "" {
		return nil, errcode.New(errcode.CodeValidationError, "unit id is required")
	}
	if !s.unitVisible(account, unitID) {
		return nil, errcode.New(errcode.CodeAuthForbidden, "unit is not visible")
	}
	targetUserID := params.TargetUserID
	if strings.HasSuffix(path, "/grant/") && targetUserID <= 0 {
		var resolveErr error
		targetUserID, resolveErr = s.ResolveSubaccountTargetUser(ctx, IsmartSubaccountContactParams{
			Phone: params.TargetPhone,
			Email: params.TargetEmail,
		})
		if resolveErr != nil {
			return nil, resolveErr
		}
	}
	if targetUserID <= 0 {
		return nil, errcode.New(errcode.CodeValidationError, "target user id is required")
	}

	payload := map[string]any{
		"user_id":        account.IsmartUserID,
		"unit_id":        unitID,
		"target_user_id": targetUserID,
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

// 24. unitVisible checks current iSmart account unit permissions.
func (s *IsmartExternalService) unitVisible(account *model.UserIsmartAccount, unitID string) bool {
	units := s.visibleUnitIDs(account)
	return len(units) == 0 || containsString(units, strings.TrimSpace(unitID))
}

// 25. visibleUnitIDs resolves the iSmart unit permission list.
func (s *IsmartExternalService) visibleUnitIDs(account *model.UserIsmartAccount) []string {
	if account == nil {
		return []string{}
	}
	message := &IsmartMessage{
		ClientBuildingFlatUnitsPermissions: unmarshalStringSlice(account.ClientBuildingFlatUnitsPermissions),
	}
	return resolveIsmartBoundUnits(message)
}

// 26. copyOptionalString copies one non-empty string payload field.
func copyOptionalString(payload map[string]any, key string, value string) {
	if strings.TrimSpace(value) != "" {
		payload[key] = strings.TrimSpace(value)
	}
}

// 27. copyQueryValue adds one non-empty integration query parameter.
func copyQueryValue(query url.Values, key string, value string) {
	if strings.TrimSpace(value) != "" {
		query.Set(key, strings.TrimSpace(value))
	}
}

// 27. normalizeIntegrationPaymentTransactionsPayload keeps payment history in documented shape.
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

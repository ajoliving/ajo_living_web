/*
 * iSmart 對外介面代理服務。
 * 1. 依 AJO 登入使用者解析 iSmart 綁定帳戶與可見大廈。
 * 2. 代理大廈資料、意見提交、門禁、二維碼與 POS payment to iSmart。
 * 3. 統一將舊系統回應轉為 AJO 後端受控回應。
 */
package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/model"
)

// 1. IsmartExternalService handles iSmart external API proxy calls.
type IsmartExternalService struct {
	runtime           *Runtime
	buildingInfoGroup singleflight.Group
}

// 2. IsmartBuildingParams defines one selected building request.
type IsmartBuildingParams struct {
	BuildingID string
}

// 3. IsmartBuildingCommentParams defines building comment submission input.
type IsmartBuildingCommentParams struct {
	BuildingID   string
	RequestType  string
	Category     string
	Subcategory  string
	Subject      string
	Content      string
	LocationText string
	UnitID       string
	ContactName  string
	ContactPhone string
	CommentType  string
	Comment      string
}

// 4. IsmartDoorOpenParams defines remote door open input.
type IsmartDoorOpenParams struct {
	BuildingID string
	DoorID     int64
}

// 5. IsmartQRCodeParams defines door QR payload input.
type IsmartQRCodeParams struct {
	BuildingID     string
	QRCodeRecordID int64
	Term           string
}

// 6. NewIsmartExternalService creates an iSmart external proxy service.
func NewIsmartExternalService(runtime *Runtime) *IsmartExternalService {
	return &IsmartExternalService{runtime: runtime}
}

// 7. ListBuildings returns iSmart buildings visible to the current AJO member.
func (s *IsmartExternalService) ListBuildings(ctx context.Context, userID int64) (map[string]any, error) {
	account, err := s.loadIsmartAccount(ctx, userID)
	if err == nil {
		buildings := normalizeStringSlice(append(s.visibleBuildingIDs(account), s.grantedBuildingIDs(ctx, userID)...))
		return map[string]any{
			"building_options": buildings,
			"is_staff":         account.IsStaff,
		}, nil
	}
	grantedBuildings, grantErr := NewBuildingAuthorizationService(s.runtime).GrantedBuildingIDs(ctx, userID)
	if grantErr != nil || len(grantedBuildings) == 0 {
		return nil, err
	}
	return map[string]any{
		"building_options": grantedBuildings,
		"is_staff":         false,
	}, nil
}

// 7.1 grantedBuildingIDs returns local delegated buildings without breaking old schemas.
func (s *IsmartExternalService) grantedBuildingIDs(ctx context.Context, userID int64) []string {
	buildingIDs, err := NewBuildingAuthorizationService(s.runtime).GrantedBuildingIDs(ctx, userID)
	if err != nil {
		return []string{}
	}
	return buildingIDs
}

// 8. GetBuildingInfo proxies the merged building info and form files API.
func (s *IsmartExternalService) GetBuildingInfo(ctx context.Context, userID int64, params IsmartBuildingParams) (map[string]any, error) {
	account, buildingID, buildingOptions, err := s.resolveBuildingAccess(ctx, userID, params.BuildingID)
	if err != nil {
		return nil, err
	}

	result, err := s.loadBuildingInfo(ctx, buildingID)
	if err != nil {
		return nil, err
	}

	return decorateIsmartPayload(result.Payload, buildingID, buildingOptions, result.Message, account.IsStaff), nil
}

// 9. SubmitBuildingComment proxies a repair or feedback service case submission.
func (s *IsmartExternalService) SubmitBuildingComment(ctx context.Context, userID int64, params IsmartBuildingCommentParams) (map[string]any, error) {
	account, buildingID, buildingOptions, err := s.resolveBuildingAccess(ctx, userID, params.BuildingID)
	if err != nil {
		return nil, err
	}
	requestType := strings.TrimSpace(params.RequestType)
	if requestType != "" && requestType != "repair" && requestType != "feedback" {
		return nil, errcode.New(errcode.CodeValidationError, "request type must be repair or feedback")
	}
	unitID := strings.TrimSpace(params.UnitID)
	if unitID != "" && !s.unitVisible(account, unitID) {
		return nil, errcode.New(errcode.CodeAuthForbidden, "unit is not visible")
	}
	content := strings.TrimSpace(params.Content)
	commentType := strings.TrimSpace(params.CommentType)
	comment := strings.TrimSpace(params.Comment)
	if content == "" {
		content = comment
	}
	if content == "" {
		return nil, errcode.New(errcode.CodeValidationError, "service case content is required")
	}

	payload := map[string]any{
		"user_id":     account.IsmartUserID,
		"building_id": buildingID,
		"content":     content,
	}
	copyOptionalString(payload, "request_type", requestType)
	copyOptionalString(payload, "category", params.Category)
	copyOptionalString(payload, "subcategory", params.Subcategory)
	copyOptionalString(payload, "subject", params.Subject)
	copyOptionalString(payload, "location_text", params.LocationText)
	copyOptionalString(payload, "unit_id", unitID)
	copyOptionalString(payload, "contact_name", params.ContactName)
	copyOptionalString(payload, "contact_phone", params.ContactPhone)
	if requestType == "" && commentType != "" {
		payload["comment_type"] = commentType
		payload["comment"] = content
	}
	result, err := s.postIntegration(ctx, "/buildings/comments/", payload)
	if err != nil && ismartFallbackAllowed(err, true) {
		legacyCommentType := commentType
		if legacyCommentType == "" {
			legacyCommentType = "其他事宜"
		}
		legacyPayload := map[string]any{
			"user_id":      account.IsmartUserID,
			"building_id":  buildingID,
			"comment_type": legacyCommentType,
			"comment":      content,
		}
		result, err = s.postExternal(ctx, "/blg-cs/submit/", legacyPayload)
	}
	if err != nil {
		return nil, err
	}

	return decorateIsmartPayload(result.Payload, buildingID, buildingOptions, result.Message, account.IsStaff), nil
}

// 10. GetBuildingAccess proxies building door access summary.
func (s *IsmartExternalService) GetBuildingAccess(ctx context.Context, userID int64, params IsmartBuildingParams) (map[string]any, error) {
	account, buildingID, buildingOptions, err := s.resolveBuildingAccess(ctx, userID, params.BuildingID)
	if err != nil {
		return nil, err
	}

	payload := map[string]any{
		"user_id":     account.IsmartUserID,
		"building_id": buildingID,
	}
	result, err := s.postIntegration(ctx, "/access/buildings/", payload)
	if err != nil && ismartFallbackAllowed(err, false) {
		result, err = s.postExternal(ctx, "/building-access/", payload)
	}
	if err != nil {
		return nil, err
	}

	return decorateIsmartPayload(result.Payload, buildingID, buildingOptions, result.Message, account.IsStaff), nil
}

// 11. OpenDoor proxies remote door open.
func (s *IsmartExternalService) OpenDoor(ctx context.Context, userID int64, params IsmartDoorOpenParams) (map[string]any, error) {
	account, buildingID, buildingOptions, err := s.resolveAuthorizedBuildingAccess(ctx, userID, params.BuildingID, BuildingPermissionRemoteDoorOpen)
	if err != nil {
		return nil, err
	}
	if params.DoorID <= 0 {
		return nil, errcode.New(errcode.CodeValidationError, "door id is required")
	}

	payload := map[string]any{
		"user_id":     account.IsmartUserID,
		"building_id": buildingID,
		"door_id":     params.DoorID,
	}
	result, err := s.postIntegration(ctx, "/access/open-door/", payload)
	if err != nil && ismartFallbackAllowed(err, true) {
		result, err = s.postExternal(ctx, "/building-access/open-door/", payload)
	}
	if err != nil {
		return nil, err
	}

	return decorateIsmartPayload(result.Payload, buildingID, buildingOptions, result.Message, account.IsStaff), nil
}

// 12.1 resolveAuthorizedBuildingAccess resolves owner access or a local grantee delegation.
func (s *IsmartExternalService) resolveAuthorizedBuildingAccess(ctx context.Context, userID int64, requestedBuildingID string, permission string) (*model.UserIsmartAccount, string, []string, error) {
	var grant model.BuildingAuthorization
	query := s.runtime.DB.WithContext(ctx).Where("grantee_user_id = ? AND status = ?", userID, BuildingAuthorizationStatusActive)
	if strings.TrimSpace(requestedBuildingID) != "" {
		query = query.Where("building_id = ?", strings.TrimSpace(requestedBuildingID))
	}
	grantErr := query.Order("created_at desc").First(&grant).Error
	if grantErr == nil {
		var permissions []string
		if err := json.Unmarshal(grant.Permissions, &permissions); err != nil {
			return nil, "", nil, errcode.New(errcode.CodeInternalError, "failed to load building authorization")
		}
		if !containsString(permissions, permission) {
			return nil, "", nil, errcode.New(errcode.CodeAuthForbidden, "building feature is not authorized")
		}
		ownerAccount, err := s.loadIsmartAccount(ctx, grant.OwnerUserID)
		if err != nil {
			return nil, "", nil, err
		}
		buildingID, buildingOptions, err := s.selectVisibleBuilding(ctx, ownerAccount, grant.BuildingID)
		if err != nil {
			return nil, "", nil, err
		}
		return ownerAccount, buildingID, buildingOptions, nil
	}
	_, accountErr := s.loadIsmartAccount(ctx, userID)
	if grantErr != nil && isBuildingAuthorizationTableMissing(grantErr) && accountErr == nil {
		return s.resolveBuildingAccess(ctx, userID, requestedBuildingID)
	}
	if grantErr != nil && !errors.Is(grantErr, gorm.ErrRecordNotFound) {
		return nil, "", nil, errcode.New(errcode.CodeInternalError, "failed to load building authorization")
	}
	if accountErr == nil {
		return s.resolveBuildingAccess(ctx, userID, requestedBuildingID)
	}
	return nil, "", nil, accountErr
}

// 12.2 isBuildingAuthorizationTableMissing keeps older schemas and rolling deployments compatible.
func isBuildingAuthorizationTableMissing(err error) bool {
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "no such table") || strings.Contains(message, "does not exist")
}

// 12. GenerateQRCode proxies door QR payload generation.
func (s *IsmartExternalService) GenerateQRCode(ctx context.Context, userID int64, params IsmartQRCodeParams) (map[string]any, error) {
	account, buildingID, buildingOptions, err := s.resolveBuildingAccess(ctx, userID, params.BuildingID)
	if err != nil {
		return nil, err
	}
	if params.QRCodeRecordID <= 0 {
		return nil, errcode.New(errcode.CodeValidationError, "qrcode record id is required")
	}
	term := strings.TrimSpace(params.Term)
	if term == "" {
		term = "dynamic"
	}

	payload := map[string]any{
		"user_id":          account.IsmartUserID,
		"qrcode_record_id": params.QRCodeRecordID,
		"term":             term,
	}
	result, err := s.postIntegration(ctx, "/access/qrcode/", payload)
	if err != nil && ismartFallbackAllowed(err, true) {
		result, err = s.postExternal(ctx, "/building-access/qrcode/", payload)
	}
	if err != nil {
		return nil, err
	}

	return decorateIsmartPayload(result.Payload, buildingID, buildingOptions, result.Message, account.IsStaff), nil
}

// 13. SubmitPOSPayment proxies POS payment to iSmart with AJO visibility checks.
func (s *IsmartExternalService) SubmitPOSPayment(ctx context.Context, userID int64, payload map[string]any) (map[string]any, error) {
	account, err := s.loadIsmartAccount(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(payload) == 0 {
		return nil, errcode.New(errcode.CodeValidationError, "payment payload is required")
	}

	requestPayload := paymentCloneMap(payload)
	buildingID, buildingOptions, err := s.selectVisibleBuilding(ctx, account, paymentStringValue(requestPayload["BLG_ID"]))
	if err != nil {
		return nil, err
	}
	requestPayload["BLG_ID"] = buildingID
	requestPayload["USER_ID"] = account.IsmartUserID

	result, err := s.postIntegration(ctx, "/payments/pos/", requestPayload)
	if err != nil && ismartFallbackAllowed(err, true) {
		result, err = s.postRoot(ctx, "/api/v1/pos-payment-to-ismart", requestPayload)
	}
	if err != nil {
		return nil, err
	}

	return decorateIsmartPayload(result.Payload, buildingID, buildingOptions, result.Message, account.IsStaff), nil
}

// 14. resolveBuildingAccess loads the account and checks one visible building.
func (s *IsmartExternalService) resolveBuildingAccess(ctx context.Context, userID int64, requestedBuildingID string) (*model.UserIsmartAccount, string, []string, error) {
	account, err := s.loadIsmartAccount(ctx, userID)
	if err != nil {
		return nil, "", nil, err
	}
	buildingID, buildingOptions, err := s.selectVisibleBuilding(ctx, account, requestedBuildingID)
	if err != nil {
		return nil, "", nil, err
	}

	return account, buildingID, buildingOptions, nil
}

// 15. loadIsmartAccount loads the current user's linked iSmart account.
func (s *IsmartExternalService) loadIsmartAccount(ctx context.Context, userID int64) (*model.UserIsmartAccount, error) {
	var account model.UserIsmartAccount
	if err := s.runtime.DB.WithContext(ctx).Where("user_id = ?", userID).First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.New(errcode.CodeAuthRequired, "ismart login is required")
		}
		return nil, errcode.New(errcode.CodeInternalError, "failed to load ismart account")
	}

	return &account, nil
}

// 16. selectVisibleBuilding returns a requested or profile-default visible building.
func (s *IsmartExternalService) selectVisibleBuilding(ctx context.Context, account *model.UserIsmartAccount, requestedBuildingID string) (string, []string, error) {
	buildingOptions := s.visibleBuildingIDs(account)
	if len(buildingOptions) == 0 {
		return "", buildingOptions, errcode.New(errcode.CodeAuthForbidden, "building is not visible")
	}

	buildingID := strings.TrimSpace(requestedBuildingID)
	if buildingID == "" {
		profileBuildingID, hasProfileBuilding := s.profileBuildingID(ctx, account.UserID, buildingOptions)
		if profileBuildingID != "" {
			return profileBuildingID, buildingOptions, nil
		}
		if hasProfileBuilding {
			return "", buildingOptions, errcode.New(errcode.CodeAuthForbidden, "profile building is not visible")
		}
		return buildingOptions[0], buildingOptions, nil
	}
	if !containsString(buildingOptions, buildingID) {
		return "", buildingOptions, errcode.New(errcode.CodeAuthForbidden, "building is not visible")
	}

	return buildingID, buildingOptions, nil
}

// 17. profileBuildingID returns the member center selected building when visible.
func (s *IsmartExternalService) profileBuildingID(ctx context.Context, userID int64, buildingOptions []string) (string, bool) {
	var profile model.UserProfile
	if err := s.runtime.DB.WithContext(ctx).Preload("PrimaryCommunity").Where("user_id = ?", userID).First(&profile).Error; err != nil {
		return "", false
	}

	boundBuildingIDs := normalizeStringSlice(unmarshalStringSlice(profile.BoundBuildingIDs))
	for _, value := range boundBuildingIDs {
		if containsString(buildingOptions, value) {
			return value, true
		}
	}
	if len(boundBuildingIDs) > 0 {
		return "", true
	}
	if profile.ResidenceBindingStatus == residenceBindingStatusPending {
		return "", false
	}
	if profile.PrimaryCommunity != nil {
		value := strings.TrimSpace(profile.PrimaryCommunity.PublicID)
		if value != "" && containsString(buildingOptions, value) {
			return value, true
		}
		return "", value != ""
	}

	return "", false
}

// 18. visibleBuildingIDs resolves the iSmart building permission list.
func (s *IsmartExternalService) visibleBuildingIDs(account *model.UserIsmartAccount) []string {
	if account == nil {
		return []string{}
	}
	message := &IsmartMessage{
		IsStaff:                   account.IsStaff,
		Building:                  unmarshalStringSlice(account.Building),
		StaffBuildingPermissions:  unmarshalStringSlice(account.StaffBuildingPermissions),
		ClientBuildingPermissions: unmarshalStringSlice(account.ClientBuildingPermissions),
	}

	return resolveIsmartBoundBuildings(message)
}

// 19. normalizeBuildingInfoPayload keeps old and new building document groups readable.
func normalizeBuildingInfoPayload(payload any) any {
	result := paymentMapValue(payload)
	if len(result) == 0 {
		return payload
	}

	documents := paymentMapValue(result["documents"])
	if documents == nil {
		documents = map[string]any{}
	}
	documents["forms"] = buildingDocumentGroup(documents, "forms")
	documents["building_info_files"] = buildingDocumentGroup(documents, "building_info_files")
	documents["floorplans"] = buildingDocumentGroup(documents, "floorplans", "floorplan")
	documents["audit_reports"] = buildingDocumentGroup(documents, "audit_reports", "auditreport", "audition", "audit_report", "auditreports", "auditions")
	documents["financial_reports"] = buildingDocumentGroup(documents, "financial_reports", "mfinreport")
	result["documents"] = documents

	return result
}

// 20. buildingDocumentGroup reads one canonical document group from compatible keys.
func buildingDocumentGroup(documents map[string]any, keys ...string) any {
	var fallback any
	for _, key := range keys {
		if value, ok := documents[key]; ok && value != nil {
			if fallback == nil {
				fallback = value
			}
			if !isEmptyDocumentGroup(value) {
				return value
			}
		}
	}
	if fallback != nil {
		return fallback
	}

	return []any{}
}

// 21. isEmptyDocumentGroup checks whether one document group is an empty list.
func isEmptyDocumentGroup(value any) bool {
	switch rows := value.(type) {
	case []any:
		return len(rows) == 0
	case []map[string]any:
		return len(rows) == 0
	case []string:
		return len(rows) == 0
	default:
		return false
	}
}

// 22. decorateIsmartPayload adds AJO visibility metadata to upstream data.
func decorateIsmartPayload(payload any, buildingID string, buildingOptions []string, message string, isStaff bool) map[string]any {
	result := paymentCloneMap(paymentMapValue(payload))
	if len(result) == 0 && payload != nil {
		result["result"] = payload
	}
	result["building_options"] = buildingOptions
	result["is_staff"] = isStaff
	if strings.TrimSpace(buildingID) != "" {
		result["selected_building_id"] = strings.TrimSpace(buildingID)
	}
	if strings.TrimSpace(message) != "" {
		result["upstream_message"] = strings.TrimSpace(message)
	}

	return result
}

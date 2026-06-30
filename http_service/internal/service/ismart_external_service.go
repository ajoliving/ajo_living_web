/*
 * iSmart 對外介面代理服務。
 * 1. 依 AJO 登入使用者解析 iSmart 綁定帳戶與可見大廈。
 * 2. 代理大廈資料、意見提交、門禁、二維碼與 POS payment to iSmart。
 * 3. 統一將舊系統回應轉為 AJO 後端受控回應。
 */
package service

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/model"
)

// 1. IsmartExternalService handles iSmart external API proxy calls.
type IsmartExternalService struct {
	runtime *Runtime
}

// 2. IsmartBuildingParams defines one selected building request.
type IsmartBuildingParams struct {
	BuildingID string
}

// 3. IsmartBuildingCommentParams defines building comment submission input.
type IsmartBuildingCommentParams struct {
	BuildingID  string
	CommentType string
	Comment     string
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
	if err != nil {
		return nil, err
	}

	buildings := s.visibleBuildingIDs(account)
	return map[string]any{
		"building_options": buildings,
		"is_staff":         account.IsStaff,
	}, nil
}

// 8. GetBuildingInfo proxies the merged building info and form files API.
func (s *IsmartExternalService) GetBuildingInfo(ctx context.Context, userID int64, params IsmartBuildingParams) (map[string]any, error) {
	account, buildingID, buildingOptions, err := s.resolveBuildingAccess(ctx, userID, params.BuildingID)
	if err != nil {
		return nil, err
	}

	result, err := s.postExternal(ctx, "/building-info/", map[string]any{
		"building_id": buildingID,
	})
	if err != nil {
		return nil, err
	}

	return decorateIsmartPayload(result.Payload, buildingID, buildingOptions, result.Message, account.IsStaff), nil
}

// 9. SubmitBuildingComment proxies building comment submission.
func (s *IsmartExternalService) SubmitBuildingComment(ctx context.Context, userID int64, params IsmartBuildingCommentParams) (map[string]any, error) {
	account, buildingID, buildingOptions, err := s.resolveBuildingAccess(ctx, userID, params.BuildingID)
	if err != nil {
		return nil, err
	}
	commentType := strings.TrimSpace(params.CommentType)
	comment := strings.TrimSpace(params.Comment)
	if commentType == "" || comment == "" {
		return nil, errcode.New(errcode.CodeValidationError, "comment type and content are required")
	}

	result, err := s.postExternal(ctx, "/blg-cs/submit/", map[string]any{
		"user_id":      account.IsmartUserID,
		"building_id":  buildingID,
		"comment_type": commentType,
		"comment":      comment,
	})
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

	result, err := s.postExternal(ctx, "/building-access/", map[string]any{
		"user_id":     account.IsmartUserID,
		"building_id": buildingID,
	})
	if err != nil {
		return nil, err
	}

	return decorateIsmartPayload(result.Payload, buildingID, buildingOptions, result.Message, account.IsStaff), nil
}

// 11. OpenDoor proxies remote door open.
func (s *IsmartExternalService) OpenDoor(ctx context.Context, userID int64, params IsmartDoorOpenParams) (map[string]any, error) {
	account, buildingID, buildingOptions, err := s.resolveBuildingAccess(ctx, userID, params.BuildingID)
	if err != nil {
		return nil, err
	}
	if params.DoorID <= 0 {
		return nil, errcode.New(errcode.CodeValidationError, "door id is required")
	}

	result, err := s.postExternal(ctx, "/building-access/open-door/", map[string]any{
		"user_id":     account.IsmartUserID,
		"building_id": buildingID,
		"door_id":     params.DoorID,
	})
	if err != nil {
		return nil, err
	}

	return decorateIsmartPayload(result.Payload, buildingID, buildingOptions, result.Message, account.IsStaff), nil
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

	result, err := s.postExternal(ctx, "/building-access/qrcode/", map[string]any{
		"user_id":          account.IsmartUserID,
		"qrcode_record_id": params.QRCodeRecordID,
		"term":             term,
	})
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

	result, err := s.postRoot(ctx, "/api/v1/pos-payment-to-ismart", requestPayload)
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

	hasProfileBuilding := false
	if profile.PrimaryCommunity != nil {
		value := strings.TrimSpace(profile.PrimaryCommunity.PublicID)
		hasProfileBuilding = value != ""
		if value != "" && containsString(buildingOptions, value) {
			return value, true
		}
	}
	for _, value := range normalizeStringSlice(unmarshalStringSlice(profile.BoundBuildingIDs)) {
		hasProfileBuilding = true
		if containsString(buildingOptions, value) {
			return value, true
		}
	}

	return "", hasProfileBuilding
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

// 19. decorateIsmartPayload adds AJO visibility metadata to upstream data.
func decorateIsmartPayload(payload map[string]any, buildingID string, buildingOptions []string, message string, isStaff bool) map[string]any {
	result := paymentCloneMap(payload)
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

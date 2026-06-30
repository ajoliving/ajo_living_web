/*
 * iSmart 對外介面代理服務。
 * 1. 依 AJO 登入使用者解析 iSmart 綁定帳戶與可見大廈。
 * 2. 代理大廈資料、意見提交、門禁、二維碼與 POS payment to iSmart。
 * 3. 統一將舊系統回應轉為 AJO 後端受控回應。
 */
package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

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
	buildingID, buildingOptions, err := s.selectVisibleBuilding(account, paymentStringValue(requestPayload["BLG_ID"]))
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
	buildingID, buildingOptions, err := s.selectVisibleBuilding(account, requestedBuildingID)
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

// 16. selectVisibleBuilding returns a requested or default visible building.
func (s *IsmartExternalService) selectVisibleBuilding(account *model.UserIsmartAccount, requestedBuildingID string) (string, []string, error) {
	buildingOptions := s.visibleBuildingIDs(account)
	if len(buildingOptions) == 0 {
		return "", buildingOptions, errcode.New(errcode.CodeAuthForbidden, "building is not visible")
	}

	buildingID := strings.TrimSpace(requestedBuildingID)
	if buildingID == "" {
		return buildingOptions[0], buildingOptions, nil
	}
	if !containsString(buildingOptions, buildingID) {
		return "", buildingOptions, errcode.New(errcode.CodeAuthForbidden, "building is not visible")
	}

	return buildingID, buildingOptions, nil
}

// 17. visibleBuildingIDs resolves the iSmart building permission list.
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

// 18. postExternal posts to the iSmart external app API base URL.
func (s *IsmartExternalService) postExternal(ctx context.Context, path string, payload map[string]any) (*ismartProxyResult, error) {
	return s.postJSON(ctx, s.externalURL(path), payload)
}

// 19. postRoot posts to an iSmart root API path outside the external app base path.
func (s *IsmartExternalService) postRoot(ctx context.Context, path string, payload map[string]any) (*ismartProxyResult, error) {
	return s.postJSON(ctx, s.rootURL(path), payload)
}

// 20. postJSON sends one JSON request to iSmart and decodes the response.
func (s *IsmartExternalService) postJSON(ctx context.Context, requestURL string, payload map[string]any) (*ismartProxyResult, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to prepare ismart request")
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, bytes.NewReader(raw))
	if err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to prepare ismart request")
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")

	response, err := s.httpClient().Do(request)
	if err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to call ismart service")
	}
	defer response.Body.Close()

	body, err := io.ReadAll(io.LimitReader(response.Body, 8*1024*1024))
	if err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to read ismart response")
	}

	return decodeIsmartProxyResponse(response.StatusCode, body)
}

// 21. externalURL builds an iSmart external app API URL.
func (s *IsmartExternalService) externalURL(path string) string {
	baseURL := strings.TrimRight(strings.TrimSpace(s.runtime.Config.IsmartExternalAppAPIBaseURL), "/")
	if baseURL == "" {
		baseURL = "https://ismart.ajoliving.com/api/v1/external"
	}

	return baseURL + "/" + strings.TrimLeft(path, "/")
}

// 22. rootURL builds an iSmart root API URL.
func (s *IsmartExternalService) rootURL(path string) string {
	baseURL := strings.TrimRight(strings.TrimSpace(s.runtime.Config.IsmartExternalAppBaseURL), "/")
	if baseURL == "" {
		baseURL = "https://ismart.ajoliving.com"
	}

	return baseURL + "/" + strings.TrimLeft(path, "/")
}

// 23. httpClient returns the iSmart HTTP client.
func (s *IsmartExternalService) httpClient() *http.Client {
	timeout := s.runtime.Config.IsmartExternalAppTimeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	return &http.Client{Timeout: timeout}
}

// 24. ismartProxyResult stores a normalized iSmart response.
type ismartProxyResult struct {
	Payload map[string]any
	Message string
}

// 25. decodeIsmartProxyResponse normalizes iSmart success and error payloads.
func decodeIsmartProxyResponse(statusCode int, body []byte) (*ismartProxyResult, error) {
	payload := map[string]any{}
	if len(strings.TrimSpace(string(body))) > 0 {
		decoder := json.NewDecoder(bytes.NewReader(body))
		decoder.UseNumber()
		if err := decoder.Decode(&payload); err != nil {
			return nil, errcode.New(errcode.CodeInternalError, "invalid ismart service response")
		}
	}

	message := strings.TrimSpace(paymentStringValue(payload["message"]))
	if statusCode < http.StatusOK || statusCode >= http.StatusMultipleChoices {
		if message == "" {
			message = "ismart service request failed"
		}
		return nil, errcode.New(ismartHTTPErrorCode(statusCode), message)
	}

	if strings.EqualFold(paymentStringValue(payload["status"]), "error") {
		if message == "" {
			message = "ismart service request failed"
		}
		return nil, errcode.New(errcode.CodeInternalError, message)
	}
	if code := strings.TrimSpace(paymentStringValue(payload["code"])); code != "" && code != "200" && !strings.EqualFold(code, errcode.CodeOK) {
		if message == "" {
			message = "ismart service request failed"
		}
		return nil, errcode.New(errcode.CodeInternalError, message)
	}

	if data, ok := payload["data"]; ok {
		if dataMap := paymentMapValue(data); len(dataMap) > 0 {
			return &ismartProxyResult{Payload: dataMap, Message: message}, nil
		}
		return &ismartProxyResult{Payload: map[string]any{"result": data}, Message: message}, nil
	}

	return &ismartProxyResult{Payload: payload, Message: message}, nil
}

// 26. ismartHTTPErrorCode maps upstream status into AJO error codes.
func ismartHTTPErrorCode(statusCode int) string {
	switch statusCode {
	case http.StatusBadRequest:
		return errcode.CodeValidationError
	case http.StatusForbidden:
		return errcode.CodeAuthForbidden
	case http.StatusNotFound:
		return errcode.CodeNotFound
	default:
		return errcode.CodeInternalError
	}
}

// 27. decorateIsmartPayload adds AJO visibility metadata to upstream data.
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

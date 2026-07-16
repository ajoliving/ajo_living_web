/*
 * iCCTV 視像監控服務。
 * 1. 依 AJO 登入使用者解析可見大廈。
 * 2. 代理 iCCTV public 授權接口取得 Orange Pi 與鏡頭 URL。
 * 3. 將舊系統回應轉為 AJO 視像監控穩定資料結構。
 */
package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"gorm.io/gorm"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/model"
)

// 1. SecurityICCTVService handles iCCTV camera lookup through AJO access checks.
type SecurityICCTVService struct {
	runtime *Runtime
}

// 2. ICCTVBuildingParams defines one selected building request.
type ICCTVBuildingParams struct {
	BuildingID string
}

// 3. ICCTVPublicResponse defines the member-facing camera response.
type ICCTVPublicResponse struct {
	SelectedBuildingID string                 `json:"selected_building_id"`
	BuildingOptions    []string               `json:"building_options"`
	IsStaff            bool                   `json:"is_staff"`
	OrangePis          []ICCTVOrangePiSummary `json:"orangepis"`
	Cameras            []ICCTVCameraSummary   `json:"cameras"`
}

// 4. ICCTVOrangePiSummary defines one Orange Pi source.
type ICCTVOrangePiSummary struct {
	OrangePiID   int64  `json:"orangepi_id"`
	OrangePiName string `json:"orangepi_name"`
	IsActive     bool   `json:"is_active"`
	CameraCount  int    `json:"camera_count"`
}

// 5. ICCTVCameraSummary defines one playable camera URL.
type ICCTVCameraSummary struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Channel      string `json:"channel"`
	URL          string `json:"url"`
	OrangePiID   int64  `json:"orangepi_id"`
	OrangePiName string `json:"orangepi_name"`
	IsActive     bool   `json:"is_active"`
}

type icctvPublicEnvelope struct {
	Success bool            `json:"success"`
	Data    icctvPublicData `json:"data"`
	Error   string          `json:"error"`
	Message string          `json:"message"`
	Code    string          `json:"code"`
}

type icctvPublicData struct {
	OrangePis []icctvPublicOrangePi `json:"orangepis"`
}

type icctvPublicOrangePi struct {
	OrangePiID   int64    `json:"orangepi_id"`
	OrangePiName string   `json:"orangepi_name"`
	IsActive     bool     `json:"is_active"`
	Token        string   `json:"token"`
	URLs         []string `json:"urls"`
}

// 6. NewSecurityICCTVService creates an iCCTV service.
func NewSecurityICCTVService(runtime *Runtime) *SecurityICCTVService {
	return &SecurityICCTVService{runtime: runtime}
}

// 7. GetPublicCameras returns camera URLs for one visible building.
func (s *SecurityICCTVService) GetPublicCameras(ctx context.Context, userID int64, params ICCTVBuildingParams) (*ICCTVPublicResponse, error) {
	account, buildingID, buildingOptions, err := s.resolveBuildingAccess(ctx, userID, params.BuildingID)
	if err != nil {
		return nil, err
	}

	result, err := s.postPublicAuth(ctx, buildingID, account.IsStaff)
	if err != nil {
		return nil, err
	}

	orangePis, cameras := s.normalizeICCTVPublicResult(result)
	return &ICCTVPublicResponse{
		SelectedBuildingID: buildingID,
		BuildingOptions:    buildingOptions,
		IsStaff:            account.IsStaff,
		OrangePis:          orangePis,
		Cameras:            cameras,
	}, nil
}

// 8. resolveBuildingAccess loads the iSmart account and checks one visible building.
func (s *SecurityICCTVService) resolveBuildingAccess(ctx context.Context, userID int64, requestedBuildingID string) (*model.UserIsmartAccount, string, []string, error) {
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

// 9. loadIsmartAccount loads the current user's linked iSmart account.
func (s *SecurityICCTVService) loadIsmartAccount(ctx context.Context, userID int64) (*model.UserIsmartAccount, error) {
	var account model.UserIsmartAccount
	if err := s.runtime.DB.WithContext(ctx).Where("user_id = ?", userID).First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.New(errcode.CodeAuthRequired, "ismart login is required")
		}
		return nil, errcode.New(errcode.CodeInternalError, "failed to load ismart account")
	}

	return &account, nil
}

// 10. selectVisibleBuilding returns a requested or profile-default visible building.
func (s *SecurityICCTVService) selectVisibleBuilding(ctx context.Context, account *model.UserIsmartAccount, requestedBuildingID string) (string, []string, error) {
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

// 11. profileBuildingID returns the member center selected building when visible.
func (s *SecurityICCTVService) profileBuildingID(ctx context.Context, userID int64, buildingOptions []string) (string, bool) {
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
	if profile.PrimaryCommunity != nil {
		value := strings.TrimSpace(profile.PrimaryCommunity.PublicID)
		if value != "" && containsString(buildingOptions, value) {
			return value, true
		}
		return "", value != ""
	}

	return "", false
}

// 12. visibleBuildingIDs resolves the iSmart building permission list.
func (s *SecurityICCTVService) visibleBuildingIDs(account *model.UserIsmartAccount) []string {
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

// 13. postPublicAuth calls iCCTV public auth for one building.
func (s *SecurityICCTVService) postPublicAuth(ctx context.Context, buildingID string, isStaff bool) (*icctvPublicEnvelope, error) {
	payload := map[string]any{
		"ismartid": buildingID,
		"is_staff": isStaff,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to prepare icctv request")
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, s.publicAuthURL(), bytes.NewReader(raw))
	if err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to prepare icctv request")
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")

	response, err := s.httpClient().Do(request)
	if err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to call icctv service")
	}
	defer response.Body.Close()

	body, err := io.ReadAll(io.LimitReader(response.Body, 8*1024*1024))
	if err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to read icctv response")
	}

	return decodeICCTVPublicResponse(response.StatusCode, body)
}

// 14. publicAuthURL builds the iCCTV public auth URL.
func (s *SecurityICCTVService) publicAuthURL() string {
	baseURL := strings.TrimRight(strings.TrimSpace(s.runtime.Config.ICCTVAPIBaseURL), "/")
	if baseURL == "" {
		baseURL = "https://icctv.skylinedances.com/api"
	}

	return baseURL + "/auth/public"
}

// 15. httpClient returns the iCCTV HTTP client.
func (s *SecurityICCTVService) httpClient() *http.Client {
	timeout := s.runtime.Config.ICCTVRequestTimeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	return &http.Client{Timeout: timeout}
}

// 16. decodeICCTVPublicResponse normalizes iCCTV success and error payloads.
func decodeICCTVPublicResponse(statusCode int, body []byte) (*icctvPublicEnvelope, error) {
	var envelope icctvPublicEnvelope
	if len(strings.TrimSpace(string(body))) > 0 {
		if err := json.Unmarshal(body, &envelope); err != nil {
			return nil, errcode.New(errcode.CodeInternalError, "invalid icctv service response")
		}
	}

	if statusCode < http.StatusOK || statusCode >= http.StatusMultipleChoices {
		return nil, errcode.New(icctvHTTPErrorCode(statusCode), icctvErrorMessage(envelope, "icctv service request failed"))
	}
	if !envelope.Success {
		return nil, errcode.New(errcode.CodeInternalError, icctvErrorMessage(envelope, "icctv service request failed"))
	}

	return &envelope, nil
}

// 17. normalizeICCTVPublicResult flattens Orange Pi URLs into camera rows.
func (s *SecurityICCTVService) normalizeICCTVPublicResult(result *icctvPublicEnvelope) ([]ICCTVOrangePiSummary, []ICCTVCameraSummary) {
	if result == nil {
		return []ICCTVOrangePiSummary{}, []ICCTVCameraSummary{}
	}
	orangePis := make([]ICCTVOrangePiSummary, 0, len(result.Data.OrangePis))
	cameras := make([]ICCTVCameraSummary, 0)
	for _, item := range result.Data.OrangePis {
		orangePis = append(orangePis, ICCTVOrangePiSummary{
			OrangePiID:   item.OrangePiID,
			OrangePiName: strings.TrimSpace(item.OrangePiName),
			IsActive:     item.IsActive,
			CameraCount:  len(item.URLs),
		})
		for index, cameraURL := range item.URLs {
			urlValue := strings.TrimSpace(cameraURL)
			if urlValue == "" {
				continue
			}
			channel := icctvChannelName(urlValue, index)
			playbackURL := ""
			if item.IsActive {
				playbackURL = s.icctvProxyURL(urlValue)
			}
			cameras = append(cameras, ICCTVCameraSummary{
				ID:           icctvCameraID(item.OrangePiID, channel, index),
				Title:        icctvCameraTitle(item.OrangePiName, channel, index),
				Channel:      channel,
				URL:          playbackURL,
				OrangePiID:   item.OrangePiID,
				OrangePiName: strings.TrimSpace(item.OrangePiName),
				IsActive:     item.IsActive,
			})
		}
	}

	return orangePis, cameras
}

// 18. icctvChannelName derives channel text from one camera URL.
func icctvChannelName(rawURL string, index int) string {
	parsed, err := url.Parse(rawURL)
	if err == nil {
		value := strings.Trim(strings.TrimSpace(parsed.Path), "/")
		if value != "" {
			parts := strings.Split(value, "/")
			candidate := strings.TrimSpace(parts[len(parts)-1])
			if candidate != "" {
				return candidate
			}
		}
	}

	return "channel" + paymentStringValue(index+1)
}

// 19. icctvCameraID returns a stable camera id.
func icctvCameraID(orangePiID int64, channel string, index int) string {
	return paymentStringValue(orangePiID) + "-" + strings.TrimSpace(channel) + "-" + paymentStringValue(index+1)
}

// 20. icctvCameraTitle returns a display title for one camera.
func icctvCameraTitle(orangePiName string, channel string, index int) string {
	name := strings.TrimSpace(orangePiName)
	channelText := strings.TrimSpace(channel)
	if channelText == "" {
		channelText = "channel" + paymentStringValue(index+1)
	}
	if name == "" {
		return channelText
	}

	return name + " / " + channelText
}

// 21. icctvErrorMessage returns a stable upstream error message.
func icctvErrorMessage(envelope icctvPublicEnvelope, fallback string) string {
	for _, value := range []string{envelope.Error, envelope.Message, envelope.Code} {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}

	return fallback
}

// 22. icctvHTTPErrorCode maps upstream status into AJO error codes.
func icctvHTTPErrorCode(statusCode int) string {
	switch statusCode {
	case http.StatusBadRequest:
		return errcode.CodeValidationError
	case http.StatusForbidden, http.StatusUnauthorized:
		return errcode.CodeAuthForbidden
	case http.StatusNotFound:
		return errcode.CodeNotFound
	default:
		return errcode.CodeInternalError
	}
}

/*
 * POS 樓宇只讀服務。
 * 1. 使用服務端 POS 憑據取得 relay token。
 * 2. 讀取 POS 大廈清單與指定大廈單位清單。
 * 3. 為公開註冊頁提供不暴露 POS 憑據的資料來源。
 */
package service

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"ajoliving_web/http_service/internal/errcode"
)

// 1. POSBuildingService handles read-only POS building lookups.
type POSBuildingService struct {
	runtime *Runtime
	mu      sync.Mutex
	token   string
}

// 2. POSBuildingSummary defines POS building list rows.
type POSBuildingSummary struct {
	BuildingID   string `json:"building_id,omitempty"`
	BuildnameCHI string `json:"buildname_chi,omitempty"`
	Buildname    string `json:"buildname,omitempty"`
	ID           string `json:"id,omitempty"`
	Name         string `json:"name,omitempty"`
}

// 3. POSUnitSummary defines POS flat unit list rows.
type POSUnitSummary struct {
	UnitID   string `json:"unit_id,omitempty"`
	ID       string `json:"id,omitempty"`
	Floor    string `json:"floor,omitempty"`
	Unit     string `json:"unit,omitempty"`
	UnitName string `json:"unit_name,omitempty"`
	Name     string `json:"name,omitempty"`
}

// 4. NewPOSBuildingService creates a POS building service instance.
func NewPOSBuildingService(runtime *Runtime) *POSBuildingService {
	return &POSBuildingService{runtime: runtime}
}

// 5. ListBuildings returns the POS building list.
func (s *POSBuildingService) ListBuildings(ctx context.Context) ([]POSBuildingSummary, error) {
	var buildings []POSBuildingSummary
	if err := s.getPOS(ctx, "/building", &buildings); err != nil {
		return nil, err
	}

	return buildings, nil
}

// 6. ListUnits returns POS units for one building.
func (s *POSBuildingService) ListUnits(ctx context.Context, buildingID string) ([]POSUnitSummary, error) {
	value := strings.TrimSpace(buildingID)
	if value == "" {
		return nil, errcode.New(errcode.CodeValidationError, "building id is required")
	}

	var units []POSUnitSummary
	if err := s.getPOS(ctx, "/building/"+url.PathEscape(value)+"/units", &units); err != nil {
		return nil, err
	}

	return units, nil
}

// 7. getPOS sends an authenticated GET request to POS relay.
func (s *POSBuildingService) getPOS(ctx context.Context, path string, target any) error {
	token, err := s.ensureToken(ctx)
	if err != nil {
		return err
	}

	if err := s.sendPOSGet(ctx, path, token, target); err == nil {
		return nil
	}

	s.clearToken()
	token, err = s.ensureToken(ctx)
	if err != nil {
		return err
	}

	return s.sendPOSGet(ctx, path, token, target)
}

// 8. sendPOSGet performs a POS GET call with the current token.
func (s *POSBuildingService) sendPOSGet(ctx context.Context, path string, token string, target any) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, s.posURL(path), nil)
	if err != nil {
		return errcode.New(errcode.CodeInternalError, "failed to prepare pos request")
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Authorization", "Bearer "+token)

	response, err := s.httpClient().Do(request)
	if err != nil {
		return errcode.New(errcode.CodeInternalError, "failed to call pos service")
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusUnauthorized {
		return errcode.New(errcode.CodeAuthRequired, "pos token expired")
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return errcode.New(errcode.CodeInternalError, "pos service request failed")
	}
	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		return errcode.New(errcode.CodeInternalError, "invalid pos service response")
	}

	return nil
}

// 9. ensureToken logs into POS relay when no cached token exists.
func (s *POSBuildingService) ensureToken(ctx context.Context) (string, error) {
	s.mu.Lock()
	token := s.token
	s.mu.Unlock()
	if token != "" {
		return token, nil
	}

	nextToken, err := s.login(ctx)
	if err != nil {
		return "", err
	}

	s.mu.Lock()
	s.token = nextToken
	s.mu.Unlock()
	return nextToken, nil
}

// 10. login obtains a POS relay token with service credentials.
func (s *POSBuildingService) login(ctx context.Context) (string, error) {
	username := strings.TrimSpace(s.runtime.Config.POSAPIUsername)
	password := strings.TrimSpace(s.runtime.Config.POSAPIPassword)
	loginType := strings.TrimSpace(s.runtime.Config.POSAPILoginType)
	if username == "" || password == "" {
		return "", errcode.New(errcode.CodeInternalError, "pos api credential is not configured")
	}
	if loginType == "" {
		loginType = "username"
	}

	payload := map[string]string{
		"login_name": username,
		"password":   password,
		"login_type": loginType,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", errcode.New(errcode.CodeInternalError, "failed to prepare pos login request")
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, s.posURL("/login"), bytes.NewReader(body))
	if err != nil {
		return "", errcode.New(errcode.CodeInternalError, "failed to prepare pos login request")
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")

	response, err := s.httpClient().Do(request)
	if err != nil {
		return "", errcode.New(errcode.CodeInternalError, "failed to call pos login")
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return "", errcode.New(errcode.CodeInternalError, "pos login failed")
	}

	var result struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return "", errcode.New(errcode.CodeInternalError, "invalid pos login response")
	}
	if strings.TrimSpace(result.Token) == "" {
		return "", errcode.New(errcode.CodeInternalError, "pos login token is empty")
	}

	return strings.TrimSpace(result.Token), nil
}

// 11. clearToken removes the cached POS token.
func (s *POSBuildingService) clearToken() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.token = ""
}

// 12. httpClient returns the shared timeout behavior for POS calls.
func (s *POSBuildingService) httpClient() *http.Client {
	timeout := s.runtime.Config.POSLoginTimeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	return &http.Client{Timeout: timeout}
}

// 13. posURL builds a POS relay URL from config.
func (s *POSBuildingService) posURL(path string) string {
	baseURL := strings.TrimRight(strings.TrimSpace(s.runtime.Config.POSAPIBaseURL), "/")
	if baseURL == "" {
		baseURL = "https://pos.ismart.skylinedances.com/api"
	}

	return baseURL + "/" + strings.TrimLeft(path, "/")
}

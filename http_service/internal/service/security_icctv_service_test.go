/*
 * iCCTV 視像監控服務測試。
 * 1. 驗證可見大廈會代理到 iCCTV public 授權接口。
 * 2. 驗證攝像頭 URL 會保留授權參數並改寫為 HTTPS 代理入口。
 * 3. 驗證離線攝像頭不會提供播放 URL。
 * 4. 驗證不可見大廈不會向上游請求。
 */
package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ajoliving_web/http_service/internal/config"
	"ajoliving_web/http_service/internal/model"
	"ajoliving_web/http_service/internal/utils"
)

// 1. TestSecurityICCTVGetPublicCameras verifies public camera lookup.
func TestSecurityICCTVGetPublicCameras(t *testing.T) {
	var upstreamPayload map[string]any
	upstream := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/api/auth/public" {
			response.WriteHeader(http.StatusNotFound)
			return
		}
		if err := json.NewDecoder(request.Body).Decode(&upstreamPayload); err != nil {
			t.Fatalf("decode upstream payload: %v", err)
		}
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`{"success":true,"data":{"orangepis":[{"orangepi_id":7,"orangepi_name":"192.168.72.174","is_active":true,"token":"hidden","urls":["http://47.83.21.100:29005/channel1?token=hidden","http://47.83.21.100:29005/channel2?token=hidden"]}]}}`))
	}))
	defer upstream.Close()

	runtimeValue, user := newSecurityICCTVTestRuntime(t, upstream.URL+"/api", true, []string{"0999900"})
	result, err := NewSecurityICCTVService(runtimeValue).GetPublicCameras(context.Background(), user.ID, ICCTVBuildingParams{
		BuildingID: "0999900",
	})
	if err != nil {
		t.Fatalf("get public cameras: %v", err)
	}
	if upstreamPayload["ismartid"] != "0999900" || upstreamPayload["is_staff"] != true {
		t.Fatalf("unexpected upstream payload: %#v", upstreamPayload)
	}
	if result.SelectedBuildingID != "0999900" || len(result.Cameras) != 2 || len(result.OrangePis) != 1 {
		t.Fatalf("unexpected camera result: %#v", result)
	}
	if result.Cameras[0].Channel != "channel1" || result.Cameras[0].URL == "" {
		t.Fatalf("unexpected first camera: %#v", result.Cameras[0])
	}
}

// 2. TestSecurityICCTVRewritesCameraURLs verifies HTTPS proxy URL rewriting.
func TestSecurityICCTVRewritesCameraURLs(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`{"success":true,"data":{"orangepis":[{"orangepi_id":7,"orangepi_name":"192.168.72.174","is_active":true,"token":"hidden","urls":["http://47.83.21.100:29005/channel1?token=hidden","http://47.83.21.100:20042/channel1?token=hidden"]}]}}`))
	}))
	defer upstream.Close()

	runtimeValue, user := newSecurityICCTVTestRuntime(t, upstream.URL+"/api", true, []string{"0999900"})
	runtimeValue.Config.ICCTVStreamProxyBaseURL = "https://icctv.skylinedances.com"
	result, err := NewSecurityICCTVService(runtimeValue).GetPublicCameras(context.Background(), user.ID, ICCTVBuildingParams{
		BuildingID: "0999900",
	})
	if err != nil {
		t.Fatalf("get public cameras: %v", err)
	}
	if result.Cameras[0].URL != "https://icctv.skylinedances.com/opi/29005/channel1?token=hidden" {
		t.Fatalf("unexpected rewritten camera url: %s", result.Cameras[0].URL)
	}
	if result.Cameras[1].URL != "http://47.83.21.100:20042/channel1?token=hidden" {
		t.Fatalf("unexpected non-stream camera url: %s", result.Cameras[1].URL)
	}
}

// 3. TestSecurityICCTVHidesOfflineCameraURLs verifies offline cameras cannot be opened.
func TestSecurityICCTVHidesOfflineCameraURLs(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`{"success":true,"data":{"orangepis":[{"orangepi_id":7,"orangepi_name":"offline","is_active":false,"token":"hidden","urls":["http://47.83.21.100:29003/channel1?token=hidden"]}]}}`))
	}))
	defer upstream.Close()

	runtimeValue, user := newSecurityICCTVTestRuntime(t, upstream.URL+"/api", true, []string{"0999900"})
	runtimeValue.Config.ICCTVStreamProxyBaseURL = "https://icctv.skylinedances.com"
	result, err := NewSecurityICCTVService(runtimeValue).GetPublicCameras(context.Background(), user.ID, ICCTVBuildingParams{
		BuildingID: "0999900",
	})
	if err != nil {
		t.Fatalf("get public cameras: %v", err)
	}
	if len(result.Cameras) != 1 || result.Cameras[0].IsActive || result.Cameras[0].URL != "" {
		t.Fatalf("unexpected offline camera result: %#v", result.Cameras)
	}
}

// 4. TestSecurityICCTVRejectsInvisibleBuilding verifies visibility checks before upstream calls.
func TestSecurityICCTVRejectsInvisibleBuilding(t *testing.T) {
	called := false
	upstream := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		called = true
		response.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	runtimeValue, user := newSecurityICCTVTestRuntime(t, upstream.URL+"/api", false, []string{"0999900"})
	_, err := NewSecurityICCTVService(runtimeValue).GetPublicCameras(context.Background(), user.ID, ICCTVBuildingParams{
		BuildingID: "0888800",
	})
	if err == nil {
		t.Fatal("expected invisible building to be rejected")
	}
	if called {
		t.Fatal("upstream should not be called for invisible building")
	}
}

// 5. newSecurityICCTVTestRuntime creates a member and linked iSmart account.
func newSecurityICCTVTestRuntime(t *testing.T, upstreamURL string, isStaff bool, buildingIDs []string) (*Runtime, model.User) {
	t.Helper()
	runtimeValue := newAuthTestRuntime(
		t,
		&config.Config{
			ICCTVAPIBaseURL:     upstreamURL,
			ICCTVRequestTimeout: 2_000_000_000,
		},
		&model.User{},
		&model.UserProfile{},
		&model.UserIsmartAccount{},
		&model.Community{},
	)
	user := model.User{
		PublicID:         utils.NewPublicID(),
		PhoneCountryCode: "+852",
		PhoneNumber:      utils.NewPublicID(),
		MemberStatus:     "active",
		MemberType:       MemberTypeUser,
		IsVerifiedPhone:  true,
	}
	if err := runtimeValue.DB.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	buildingJSON, err := marshalJSON(buildingIDs)
	if err != nil {
		t.Fatalf("marshal building json: %v", err)
	}
	account := model.UserIsmartAccount{
		UserID:                             user.ID,
		IsmartUserID:                       user.ID + 1000,
		Username:                           "icctv-user",
		IsStaff:                            isStaff,
		Building:                           buildingJSON,
		StaffBuildingPermissions:           []byte("[]"),
		ClientBuildingPermissions:          buildingJSON,
		ClientBuildingFlatUnitsPermissions: []byte("[]"),
		RawMessage:                         []byte("{}"),
	}
	if isStaff {
		account.StaffBuildingPermissions = buildingJSON
	}
	if err := runtimeValue.DB.Create(&account).Error; err != nil {
		t.Fatalf("create ismart account: %v", err)
	}

	return runtimeValue, user
}

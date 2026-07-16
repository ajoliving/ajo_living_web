/*
 * iSmart 對外介面代理服務測試。
 * 1. 驗證會員中心綁定大廈優先於 iSmart 可見大廈預設順序。
 * 2. 驗證未指定大廈時仍保留 iSmart 可見範圍校驗。
 * 3. 驗證大廈資料快取不重複回源且不繞過會員權限。
 */
package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"ajoliving_web/http_service/internal/config"
	"ajoliving_web/http_service/internal/model"
	"ajoliving_web/http_service/internal/utils"
)

// 1. TestIsmartSelectVisibleBuildingPrefersProfileBuilding verifies member-center binding priority.
func TestIsmartSelectVisibleBuildingPrefersProfileBuilding(t *testing.T) {
	runtimeValue := newAuthTestRuntime(
		t,
		nil,
		&model.User{},
		&model.UserProfile{},
		&model.UserIsmartAccount{},
		&model.Community{},
	)
	user := model.User{
		PublicID:         utils.NewPublicID(),
		PhoneCountryCode: "+852",
		PhoneNumber:      "61234567",
		MemberStatus:     "active",
		MemberType:       MemberTypeUser,
		IsVerifiedPhone:  true,
	}
	if err := runtimeValue.DB.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	boundBuilding := model.Community{
		PublicID:      "BLG-002",
		CommunityType: "building",
		NameZH:        "仁英大廈",
		DistrictCode:  "unknown",
	}
	if err := runtimeValue.DB.Create(&boundBuilding).Error; err != nil {
		t.Fatalf("create building: %v", err)
	}
	profileBuildingJSON, err := marshalJSON([]string{"BLG-002"})
	if err != nil {
		t.Fatalf("marshal profile building json: %v", err)
	}
	if err := runtimeValue.DB.Create(&model.UserProfile{
		UserID:                 user.ID,
		PrimaryCommunityID:     &boundBuilding.ID,
		BoundBuildingIDs:       profileBuildingJSON,
		ResidenceFloor:         "01",
		ResidenceUnit:          "B",
		ResidenceBindingStatus: residenceBindingStatusPending,
	}).Error; err != nil {
		t.Fatalf("create profile: %v", err)
	}
	visibleBuildingJSON, err := marshalJSON([]string{"BLG-001", "BLG-002"})
	if err != nil {
		t.Fatalf("marshal visible building json: %v", err)
	}
	account := &model.UserIsmartAccount{
		UserID:                    user.ID,
		IsmartUserID:              88,
		Username:                  "patrick",
		ClientBuildingPermissions: visibleBuildingJSON,
		Building:                  []byte("[]"),
		StaffBuildingPermissions:  []byte("[]"),
		RawMessage:                []byte("{}"),
	}

	buildingID, buildingOptions, err := NewIsmartExternalService(runtimeValue).selectVisibleBuilding(context.Background(), account, "")
	if err != nil {
		t.Fatalf("select building: %v", err)
	}
	if buildingID != "BLG-002" {
		t.Fatalf("expected profile building BLG-002, got %s", buildingID)
	}
	if len(buildingOptions) != 2 || buildingOptions[0] != "BLG-001" || buildingOptions[1] != "BLG-002" {
		t.Fatalf("expected original visible building options, got %#v", buildingOptions)
	}
}

// 2. TestIsmartSelectVisibleBuildingRejectsInvisibleProfileBuilding verifies mismatched bindings fail closed.
func TestIsmartSelectVisibleBuildingRejectsInvisibleProfileBuilding(t *testing.T) {
	runtimeValue := newAuthTestRuntime(
		t,
		nil,
		&model.User{},
		&model.UserProfile{},
		&model.UserIsmartAccount{},
		&model.Community{},
	)
	user := model.User{
		PublicID:         utils.NewPublicID(),
		PhoneCountryCode: "+852",
		PhoneNumber:      "61234568",
		MemberStatus:     "active",
		MemberType:       MemberTypeUser,
		IsVerifiedPhone:  true,
	}
	if err := runtimeValue.DB.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	boundBuilding := model.Community{
		PublicID:      "BLG-002",
		CommunityType: "building",
		NameZH:        "仁英大廈",
		DistrictCode:  "unknown",
	}
	if err := runtimeValue.DB.Create(&boundBuilding).Error; err != nil {
		t.Fatalf("create building: %v", err)
	}
	profileBuildingJSON, err := marshalJSON([]string{"BLG-002"})
	if err != nil {
		t.Fatalf("marshal profile building json: %v", err)
	}
	if err := runtimeValue.DB.Create(&model.UserProfile{
		UserID:             user.ID,
		PrimaryCommunityID: &boundBuilding.ID,
		BoundBuildingIDs:   profileBuildingJSON,
	}).Error; err != nil {
		t.Fatalf("create profile: %v", err)
	}
	visibleBuildingJSON, err := marshalJSON([]string{"BLG-001"})
	if err != nil {
		t.Fatalf("marshal visible building json: %v", err)
	}
	account := &model.UserIsmartAccount{
		UserID:                    user.ID,
		IsmartUserID:              89,
		Username:                  "patrick",
		ClientBuildingPermissions: visibleBuildingJSON,
		Building:                  []byte("[]"),
		StaffBuildingPermissions:  []byte("[]"),
		RawMessage:                []byte("{}"),
	}

	_, _, err = NewIsmartExternalService(runtimeValue).selectVisibleBuilding(context.Background(), account, "")
	if err == nil {
		t.Fatal("expected invisible profile building to be rejected")
	}
}

// 3. TestIsmartListManagementFeesUsesIntegration verifies raw array integration payloads are decorated.
func TestIsmartListManagementFeesUsesIntegration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/api/v1/integration/buildings/receivables/management-fees/" {
			t.Fatalf("unexpected path: %s", request.URL.Path)
		}
		if request.URL.Query().Get("building_id") != "0348200" || request.URL.Query().Get("data_structure") != "table" {
			t.Fatalf("unexpected query: %s", request.URL.RawQuery)
		}
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`[{"unit":"18A","amount":1500}]`))
	}))
	defer server.Close()

	runtimeValue, user := newIsmartIntegrationTestRuntime(t, server.URL)
	result, err := NewIsmartExternalService(runtimeValue).ListManagementFees(context.Background(), user.ID, IsmartReceivableParams{
		BuildingID: "0348200",
	})
	if err != nil {
		t.Fatalf("list management fees: %v", err)
	}
	rows, ok := result["result"].([]any)
	if !ok || len(rows) != 1 {
		t.Fatalf("expected one raw result row, got %#v", result["result"])
	}
	if result["selected_building_id"] != "0348200" {
		t.Fatalf("expected selected building, got %#v", result)
	}
}

// 4. TestIsmartGetBuildingInfoNormalizesDocumentGroups verifies old and new file keys are readable.
func TestIsmartGetBuildingInfoNormalizesDocumentGroups(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/api/v1/integration/buildings/info/" {
			t.Fatalf("unexpected path: %s", request.URL.Path)
		}
		if request.URL.Query().Get("building_id") != "0348200" {
			t.Fatalf("unexpected query: %s", request.URL.RawQuery)
		}
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`{
			"status":"success",
			"data":{
				"building":{"building_id":"0348200"},
				"documents":{
					"floorplans":[],
					"floorplan":[{"id":1,"title":"平面圖"}],
					"audit_reports":[],
					"auditreport":[],
					"audit_report":[{"id":3,"title":"核數報告"}],
					"mfinreport":[{"id":2,"title":"財務報告"}]
				}
			}
		}`))
	}))
	defer server.Close()

	runtimeValue, user := newIsmartIntegrationTestRuntime(t, server.URL)
	result, err := NewIsmartExternalService(runtimeValue).GetBuildingInfo(context.Background(), user.ID, IsmartBuildingParams{
		BuildingID: "0348200",
	})
	if err != nil {
		t.Fatalf("get building info: %v", err)
	}
	documents, ok := result["documents"].(map[string]any)
	if !ok {
		t.Fatalf("expected document map, got %#v", result["documents"])
	}
	floorplans, ok := documents["floorplans"].([]any)
	if !ok || len(floorplans) != 1 {
		t.Fatalf("expected normalized floorplans, got %#v", documents["floorplans"])
	}
	forms, ok := documents["forms"].([]any)
	if !ok || len(forms) != 0 {
		t.Fatalf("expected empty forms array, got %#v", documents["forms"])
	}
	auditReports, ok := documents["audit_reports"].([]any)
	if !ok || len(auditReports) != 1 {
		t.Fatalf("expected normalized audit reports, got %#v", documents["audit_reports"])
	}
	financialReports, ok := documents["financial_reports"].([]any)
	if !ok || len(financialReports) != 1 {
		t.Fatalf("expected normalized financial reports, got %#v", documents["financial_reports"])
	}
}

// 5. TestIsmartListBuildingNoticesFallbackUsesLegacyBuildingKey verifies old notice API payloads.
func TestIsmartListBuildingNoticesFallbackUsesLegacyBuildingKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/api/v1/integration/buildings/notices/":
			response.Header().Set("Content-Type", "application/json")
			response.WriteHeader(http.StatusInternalServerError)
			_, _ = response.Write([]byte(`{"status":"error","message":"missing integration route"}`))
		case "/api/v1/building-notices/":
			var payload map[string]any
			if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
				t.Fatalf("decode payload: %v", err)
			}
			if payload["blg_id"] != "0348200" {
				t.Fatalf("expected legacy blg_id, got %#v", payload)
			}
			if _, ok := payload["building_id"]; ok {
				t.Fatalf("unexpected building_id in legacy payload: %#v", payload)
			}
			response.Header().Set("Content-Type", "application/json")
			_, _ = response.Write([]byte(`[{"id":18,"mess_code":"N1","mess_title":"通告"}]`))
		default:
			t.Fatalf("unexpected path: %s", request.URL.Path)
		}
	}))
	defer server.Close()

	runtimeValue, user := newIsmartIntegrationTestRuntime(t, server.URL)
	result, err := NewIsmartExternalService(runtimeValue).ListBuildingNotices(context.Background(), user.ID, IsmartBuildingParams{
		BuildingID: "0348200",
	})
	if err != nil {
		t.Fatalf("list building notices: %v", err)
	}
	rows, ok := result["result"].([]any)
	if !ok || len(rows) != 1 {
		t.Fatalf("expected one fallback notice, got %#v", result["result"])
	}
}

// 6. TestNormalizeBuildingInfoPayloadUsesAuditionAlias verifies the current upstream audit key remains readable.
func TestNormalizeBuildingInfoPayloadUsesAuditionAlias(t *testing.T) {
	payload := map[string]any{
		"documents": map[string]any{
			"audit_reports": []any{},
			"audition": []any{
				map[string]any{"id": 8, "title": "核數報告"},
			},
		},
	}
	result := paymentMapValue(normalizeBuildingInfoPayload(payload))
	documents := paymentMapValue(result["documents"])
	auditReports, ok := documents["audit_reports"].([]any)
	if !ok || len(auditReports) != 1 {
		t.Fatalf("expected audition alias to normalize audit reports, got %#v", documents["audit_reports"])
	}
}

// 7. TestIsmartOwnerBindingInjectsCurrentUser verifies frontend user_id cannot override iSmart identity.
func TestIsmartOwnerBindingInjectsCurrentUser(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/api/v1/integration/buildings/building-flat-owner-binding-requests/" {
			t.Fatalf("unexpected path: %s", request.URL.Path)
		}
		var payload map[string]any
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		if payload["user"].(float64) != 88 {
			t.Fatalf("expected current ismart user 88, got %#v", payload["user"])
		}
		if payload["building_id"] != "0348200" {
			t.Fatalf("unexpected building id: %#v", payload["building_id"])
		}
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`{"status":"success","data":{"owner_reg_id":45,"status":"pending_approval"}}`))
	}))
	defer server.Close()

	runtimeValue, user := newIsmartIntegrationTestRuntime(t, server.URL)
	result, err := NewIsmartExternalService(runtimeValue).SubmitOwnerBindingRequest(context.Background(), user.ID, IsmartOwnerBindingParams{
		BuildingID: "0348200",
		OwnedFlat:  []string{"0348200001"},
		Role:       "業主",
	})
	if err != nil {
		t.Fatalf("submit owner binding: %v", err)
	}
	if result["owner_reg_id"] != json.Number("45") && result["owner_reg_id"] != float64(45) {
		t.Fatalf("expected owner_reg_id in result, got %#v", result)
	}
}

// 8. TestIsmartGetBuildingInfoUsesSharedCacheAfterPermissionCheck verifies cache reuse and access isolation.
func TestIsmartGetBuildingInfoUsesSharedCacheAfterPermissionCheck(t *testing.T) {
	var upstreamCalls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		upstreamCalls.Add(1)
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`{"status":"success","data":{"building":{"building_id":"0348200","name":"測試大廈"}}}`))
	}))
	defer server.Close()

	runtimeValue, user := newIsmartIntegrationTestRuntime(t, server.URL)
	cacheStore := newMemoryCacheStore()
	runtimeValue.CacheStore = cacheStore
	runtimeValue.Config.IsmartBuildingCacheTTL = 5 * time.Minute
	ismartService := NewIsmartExternalService(runtimeValue)

	for range 2 {
		result, err := ismartService.GetBuildingInfo(context.Background(), user.ID, IsmartBuildingParams{BuildingID: "0348200"})
		if err != nil {
			t.Fatalf("get cached building info: %v", err)
		}
		if result["selected_building_id"] != "0348200" {
			t.Fatalf("expected selected building metadata, got %#v", result)
		}
	}
	if upstreamCalls.Load() != 1 {
		t.Fatalf("expected one upstream request, got %d", upstreamCalls.Load())
	}
	cached := string(cacheStore.value(ismartBuildingInfoCacheKeyPrefix + "0348200"))
	if strings.Contains(cached, "building_options") || strings.Contains(cached, "selected_building_id") {
		t.Fatalf("member metadata must not be stored in shared cache: %s", cached)
	}

	unauthorizedUser := model.User{
		PublicID:         utils.NewPublicID(),
		PhoneCountryCode: "+852",
		PhoneNumber:      "61234570",
		MemberStatus:     "active",
		MemberType:       MemberTypeUser,
		IsVerifiedPhone:  true,
	}
	if err := runtimeValue.DB.Create(&unauthorizedUser).Error; err != nil {
		t.Fatalf("create unauthorized user: %v", err)
	}
	otherBuildings, err := marshalJSON([]string{"0999900"})
	if err != nil {
		t.Fatalf("marshal unauthorized buildings: %v", err)
	}
	if err := runtimeValue.DB.Create(&model.UserIsmartAccount{
		UserID:                             unauthorizedUser.ID,
		IsmartUserID:                       99,
		Username:                           "other-member",
		ClientBuildingPermissions:          otherBuildings,
		ClientBuildingFlatUnitsPermissions: []byte("[]"),
		Building:                           []byte("[]"),
		StaffBuildingPermissions:           []byte("[]"),
		RawMessage:                         []byte("{}"),
	}).Error; err != nil {
		t.Fatalf("create unauthorized ismart account: %v", err)
	}
	if _, err := ismartService.GetBuildingInfo(context.Background(), unauthorizedUser.ID, IsmartBuildingParams{BuildingID: "0348200"}); err == nil {
		t.Fatal("expected unauthorized cached building request to fail")
	}
	if upstreamCalls.Load() != 1 {
		t.Fatalf("unauthorized request must fail before upstream or cache read, got %d calls", upstreamCalls.Load())
	}
}

// 9. newIsmartIntegrationTestRuntime creates a linked iSmart test member.
func newIsmartIntegrationTestRuntime(t *testing.T, baseURL string) (*Runtime, model.User) {
	t.Helper()
	runtimeValue := newAuthTestRuntime(
		t,
		&config.Config{
			IsmartExternalAppBaseURL:    baseURL,
			IsmartExternalAppAPIBaseURL: baseURL + "/api/v1/external",
			IsmartIntegrationAPIBaseURL: baseURL + "/api/v1/integration",
		},
		&model.User{},
		&model.UserProfile{},
		&model.UserIsmartAccount{},
		&model.Community{},
	)
	user := model.User{
		PublicID:         utils.NewPublicID(),
		PhoneCountryCode: "+852",
		PhoneNumber:      "61234569",
		MemberStatus:     "active",
		MemberType:       MemberTypeUser,
		IsVerifiedPhone:  true,
	}
	if err := runtimeValue.DB.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	visibleBuildingJSON, err := marshalJSON([]string{"0348200"})
	if err != nil {
		t.Fatalf("marshal visible building json: %v", err)
	}
	visibleUnitJSON, err := marshalJSON([]string{"0348200001"})
	if err != nil {
		t.Fatalf("marshal visible unit json: %v", err)
	}
	if err := runtimeValue.DB.Create(&model.UserIsmartAccount{
		UserID:                             user.ID,
		IsmartUserID:                       88,
		Username:                           "patrick",
		ClientBuildingPermissions:          visibleBuildingJSON,
		ClientBuildingFlatUnitsPermissions: visibleUnitJSON,
		Building:                           []byte("[]"),
		StaffBuildingPermissions:           []byte("[]"),
		RawMessage:                         []byte("{}"),
	}).Error; err != nil {
		t.Fatalf("create ismart account: %v", err)
	}

	return runtimeValue, user
}

// 10. memoryCacheStore is an in-memory CacheStore used by service tests.
type memoryCacheStore struct {
	mu     sync.Mutex
	values map[string][]byte
}

// 11. newMemoryCacheStore creates an empty test cache.
func newMemoryCacheStore() *memoryCacheStore {
	return &memoryCacheStore{values: make(map[string][]byte)}
}

// 12. Get returns one copied test cache value.
func (s *memoryCacheStore) Get(_ context.Context, key string) ([]byte, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	value, ok := s.values[key]
	return append([]byte(nil), value...), ok, nil
}

// 13. Set stores one copied test cache value.
func (s *memoryCacheStore) Set(_ context.Context, key string, value []byte, _ time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.values[key] = append([]byte(nil), value...)
	return nil
}

// 14. Close completes the CacheStore contract for tests.
func (s *memoryCacheStore) Close() error {
	return nil
}

// 15. value returns one cache value for assertions.
func (s *memoryCacheStore) value(key string) []byte {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]byte(nil), s.values[key]...)
}

/*
 * POS 大廈與單位目錄快取回歸測試。
 * 1. 驗證 Redis 命中與同程序並行請求只回源一次。
 * 2. 驗證不同大廈的單位目錄使用獨立 key。
 * 3. 驗證 Redis 故障時直接回源 POS。
 * 4. 驗證會員權限在快取讀取前即時校驗並過濾目錄。
 */
package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"ajoliving_web/http_service/internal/config"
	"ajoliving_web/http_service/internal/model"
	"ajoliving_web/http_service/internal/utils"
)

// 1. TestPOSBuildingDirectoryCacheCollapsesConcurrentMiss verifies one upstream fill and later hits.
func TestPOSBuildingDirectoryCacheCollapsesConcurrentMiss(t *testing.T) {
	var buildingCalls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("Content-Type", "application/json")
		switch request.URL.Path {
		case "/login":
			_, _ = response.Write([]byte(`{"token":"service-token"}`))
		case "/building":
			buildingCalls.Add(1)
			time.Sleep(20 * time.Millisecond)
			_, _ = response.Write([]byte(`[{"building_id":"0999900","buildname_chi":"測試1大廈"}]`))
		default:
			response.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	cacheStore := newMemoryCacheStore()
	serviceValue := NewPOSBuildingService(newPOSDirectoryTestRuntime(server.URL, cacheStore))
	var group sync.WaitGroup
	errorsChannel := make(chan error, 8)
	for range 8 {
		group.Add(1)
		go func() {
			defer group.Done()
			rows, err := serviceValue.ListBuildings(context.Background())
			if err != nil {
				errorsChannel <- err
				return
			}
			if len(rows) != 1 || rows[0].BuildnameCHI != "測試1大廈" {
				errorsChannel <- errors.New("unexpected building directory")
			}
		}()
	}
	group.Wait()
	close(errorsChannel)
	for err := range errorsChannel {
		t.Fatal(err)
	}
	if _, err := serviceValue.ListBuildings(context.Background()); err != nil {
		t.Fatalf("read cached buildings: %v", err)
	}
	if buildingCalls.Load() != 1 {
		t.Fatalf("expected one upstream building request, got %d", buildingCalls.Load())
	}
}

// 2. TestPOSUnitDirectoryCacheSeparatesBuildingKeys verifies unit caches do not overlap.
func TestPOSUnitDirectoryCacheSeparatesBuildingKeys(t *testing.T) {
	var firstCalls atomic.Int32
	var secondCalls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("Content-Type", "application/json")
		switch request.URL.Path {
		case "/login":
			_, _ = response.Write([]byte(`{"token":"service-token"}`))
		case "/building/0419900/units":
			firstCalls.Add(1)
			_, _ = response.Write([]byte(`[{"unit_id":"04199000112","floor":"01","unit":"12"}]`))
		case "/building/0999900/units":
			secondCalls.Add(1)
			_, _ = response.Write([]byte(`[{"unit_id":"09999000407","floor":"04","unit":"G"}]`))
		default:
			response.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	cacheStore := newMemoryCacheStore()
	serviceValue := NewPOSBuildingService(newPOSDirectoryTestRuntime(server.URL, cacheStore))
	for range 2 {
		if _, err := serviceValue.ListUnits(context.Background(), "0419900"); err != nil {
			t.Fatalf("load first unit directory: %v", err)
		}
		if _, err := serviceValue.ListUnits(context.Background(), "0999900"); err != nil {
			t.Fatalf("load second unit directory: %v", err)
		}
	}
	if firstCalls.Load() != 1 || secondCalls.Load() != 1 {
		t.Fatalf("expected one upstream call per building, got first=%d second=%d", firstCalls.Load(), secondCalls.Load())
	}
	if len(cacheStore.value(posBuildingUnitsCacheKeyPrefix+"0419900")) == 0 || len(cacheStore.value(posBuildingUnitsCacheKeyPrefix+"0999900")) == 0 {
		t.Fatal("expected independent unit cache keys")
	}
}

// 3. TestPOSDirectoryCacheFailureFallsBackToUpstream verifies cache errors never block directory reads.
func TestPOSDirectoryCacheFailureFallsBackToUpstream(t *testing.T) {
	var buildingCalls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("Content-Type", "application/json")
		if request.URL.Path == "/login" {
			_, _ = response.Write([]byte(`{"token":"service-token"}`))
			return
		}
		if request.URL.Path == "/building" {
			buildingCalls.Add(1)
			_, _ = response.Write([]byte(`[{"building_id":"0999900"}]`))
			return
		}
		response.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	serviceValue := NewPOSBuildingService(newPOSDirectoryTestRuntime(server.URL, failingPOSCacheStore{}))
	for range 2 {
		if _, err := serviceValue.ListBuildings(context.Background()); err != nil {
			t.Fatalf("load buildings with unavailable cache: %v", err)
		}
	}
	if buildingCalls.Load() != 2 {
		t.Fatalf("expected both requests to reach POS after cache failure, got %d", buildingCalls.Load())
	}
}

// 4. TestPOSMemberDirectoryChecksPermissionsBeforeSharedCache verifies live local visibility and filtering.
func TestPOSMemberDirectoryChecksPermissionsBeforeSharedCache(t *testing.T) {
	var allowedUnitCalls atomic.Int32
	var forbiddenUnitCalls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("Content-Type", "application/json")
		switch request.URL.Path {
		case "/login":
			_, _ = response.Write([]byte(`{"token":"service-token"}`))
		case "/building":
			_, _ = response.Write([]byte(`[{"building_id":"0999900","buildname_chi":"測試1大廈"},{"building_id":"0888800","buildname_chi":"其他大廈"}]`))
		case "/building/0999900/units":
			allowedUnitCalls.Add(1)
			_, _ = response.Write([]byte(`[{"unit_id":"09999000407","floor":"04","unit":"G"},{"unit_id":"09999000408","floor":"04","unit":"H"}]`))
		case "/building/0888800/units":
			forbiddenUnitCalls.Add(1)
			_, _ = response.Write([]byte(`[{"unit_id":"08888000101","floor":"01","unit":"01"}]`))
		default:
			response.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	cfg := &config.Config{
		POSAPIBaseURL:        server.URL,
		POSAPIUsername:       "service-user",
		POSAPIPassword:       "service-password",
		POSLoginTimeout:      time.Second,
		POSDirectoryCacheTTL: 5 * time.Minute,
	}
	runtimeValue := newAuthTestRuntime(t, cfg, &model.User{}, &model.UserProfile{}, &model.UserIsmartAccount{}, &model.Community{})
	runtimeValue.CacheStore = newMemoryCacheStore()
	user := model.User{
		PublicID:         utils.NewPublicID(),
		PhoneCountryCode: "+852",
		PhoneNumber:      "61234579",
		MemberStatus:     "active",
		MemberType:       MemberTypeUser,
		IsVerifiedPhone:  true,
	}
	if err := runtimeValue.DB.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	if err := runtimeValue.DB.Create(&model.UserProfile{UserID: user.ID}).Error; err != nil {
		t.Fatalf("create profile: %v", err)
	}
	buildingJSON, err := json.Marshal([]string{"0999900"})
	if err != nil {
		t.Fatalf("marshal building permissions: %v", err)
	}
	unitJSON, err := json.Marshal([]string{"09999000407"})
	if err != nil {
		t.Fatalf("marshal unit permissions: %v", err)
	}
	account := model.UserIsmartAccount{
		UserID:                             user.ID,
		IsmartUserID:                       201,
		Username:                           "resident-cache-member",
		Building:                           []byte("[]"),
		StaffBuildingPermissions:           []byte("[]"),
		ClientBuildingPermissions:          buildingJSON,
		ClientBuildingFlatUnitsPermissions: unitJSON,
		RawMessage:                         []byte("{}"),
		ProfileSnapshot:                    []byte("{}"),
	}
	if err := runtimeValue.DB.Create(&account).Error; err != nil {
		t.Fatalf("create iSmart account: %v", err)
	}

	directoryService := NewPOSBuildingService(runtimeValue)
	paymentService := NewPOSPaymentService(runtimeValue, directoryService)
	buildings, err := paymentService.ListMemberBuildings(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("load member buildings: %v", err)
	}
	if len(buildings) != 1 || posBuildingID(buildings[0]) != "0999900" {
		t.Fatalf("expected only visible building, got %#v", buildings)
	}
	units, err := paymentService.ListMemberUnits(context.Background(), user.ID, "0999900")
	if err != nil {
		t.Fatalf("load member units: %v", err)
	}
	if len(units) != 1 || posUnitID(units[0]) != "09999000407" {
		t.Fatalf("expected only visible unit, got %#v", units)
	}
	if _, err := paymentService.ListMemberUnits(context.Background(), user.ID, "0888800"); err == nil {
		t.Fatal("expected unauthorized building to fail before directory read")
	}
	if allowedUnitCalls.Load() != 1 || forbiddenUnitCalls.Load() != 0 {
		t.Fatalf("unexpected unit upstream calls: allowed=%d forbidden=%d", allowedUnitCalls.Load(), forbiddenUnitCalls.Load())
	}

	updatedBuildings, err := json.Marshal([]string{"0888800"})
	if err != nil {
		t.Fatalf("marshal updated building permissions: %v", err)
	}
	updatedUnits, err := json.Marshal([]string{"08888000101"})
	if err != nil {
		t.Fatalf("marshal updated unit permissions: %v", err)
	}
	if err := runtimeValue.DB.Model(&model.UserIsmartAccount{}).Where("user_id = ?", user.ID).Updates(map[string]any{
		"client_building_permissions":            updatedBuildings,
		"client_building_flat_units_permissions": updatedUnits,
	}).Error; err != nil {
		t.Fatalf("update local permissions: %v", err)
	}
	if _, err := paymentService.ListMemberUnits(context.Background(), user.ID, "0999900"); err == nil {
		t.Fatal("expected cached old building to be denied after local permission update")
	}
	if allowedUnitCalls.Load() != 1 {
		t.Fatalf("expected permission denial before cached directory read, got %d upstream calls", allowedUnitCalls.Load())
	}
}

// 5. newPOSDirectoryTestRuntime creates one isolated directory runtime.
func newPOSDirectoryTestRuntime(baseURL string, cacheStore CacheStore) *Runtime {
	return &Runtime{
		Config: &config.Config{
			POSAPIBaseURL:        baseURL,
			POSAPIUsername:       "service-user",
			POSAPIPassword:       "service-password",
			POSLoginTimeout:      time.Second,
			POSDirectoryCacheTTL: 5 * time.Minute,
		},
		CacheStore: cacheStore,
		Now:        time.Now,
	}
}

// 6. failingPOSCacheStore simulates unavailable Redis reads and writes.
type failingPOSCacheStore struct{}

// 7. Get returns one simulated Redis error.
func (failingPOSCacheStore) Get(context.Context, string) ([]byte, bool, error) {
	return nil, false, errors.New("redis unavailable")
}

// 8. Set returns one simulated Redis error.
func (failingPOSCacheStore) Set(context.Context, string, []byte, time.Duration) error {
	return errors.New("redis unavailable")
}

// 9. Close completes the CacheStore contract.
func (failingPOSCacheStore) Close() error {
	return nil
}

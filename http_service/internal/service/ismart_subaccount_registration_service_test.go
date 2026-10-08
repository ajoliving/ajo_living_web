/*
 * iSmart 副戶註冊、授權及臨時密碼郵件流程測試。
 * 1. 使用 httptest 模擬 iSmart check-contact、register 及 grant 接口。
 * 2. 使用內存 SQLite 提供業主的可見單位權限。
 * 3. 驗證既有帳戶授權、新帳戶註冊授權及聯絡資料衝突分支。
 */
package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"ajoliving_web/http_service/internal/config"
	"ajoliving_web/http_service/internal/model"
)

type capturedSubaccountMail struct {
	Recipient string
	Subject   string
	Body      string
}

type capturedSubaccountMailSender struct {
	Messages []capturedSubaccountMail
	Err      error
}

// 1. Send captures one outgoing subaccount notification.
func (s *capturedSubaccountMailSender) Send(_ context.Context, recipient string, subject string, body string) error {
	s.Messages = append(s.Messages, capturedSubaccountMail{
		Recipient: recipient,
		Subject:   subject,
		Body:      body,
	})
	return s.Err
}

type subaccountTestCache struct {
	values map[string][]byte
}

// 2. Get returns one deterministic directory cache value.
func (s *subaccountTestCache) Get(_ context.Context, key string) ([]byte, bool, error) {
	value, ok := s.values[key]
	return value, ok, nil
}

// 3. Set keeps cache writes available for the production cache path.
func (s *subaccountTestCache) Set(_ context.Context, key string, value []byte, _ time.Duration) error {
	if s.values == nil {
		s.values = make(map[string][]byte)
	}
	s.values[key] = value
	return nil
}

// 4. Close implements CacheStore for the in-memory test cache.
func (s *subaccountTestCache) Close() error {
	return nil
}

// 5. TestRegisterAndGrantSubaccountExistingAccount verifies matching contacts grant an existing user.
func TestRegisterAndGrantSubaccountExistingAccount(t *testing.T) {
	var grantPayload map[string]any
	server := newSubaccountIntegrationTestServer(t, func(path string, payload map[string]any) (int, map[string]any) {
		switch path {
		case "/api/v1/integration/auth/check-contact/":
			if _, ok := payload["email"]; ok {
				return http.StatusOK, map[string]any{"status": "success", "data": map[string]any{
					"email_available": false,
					"email_user_id":   456,
				}}
			}
			return http.StatusOK, map[string]any{"status": "success", "data": map[string]any{
				"phone_available": false,
				"phone_user_id":   456,
			}}
		case "/api/v1/integration/buildings/subaccounts/grant/":
			grantPayload = payload
			return http.StatusOK, map[string]any{"status": "success", "data": map[string]any{
				"target_username":  "200456",
				"upstream_message": "granted",
			}}
		default:
			t.Fatalf("unexpected upstream path %s", path)
			return http.StatusInternalServerError, nil
		}
	})
	defer server.Close()

	mailSender := &capturedSubaccountMailSender{}
	service := newSubaccountRegistrationTestService(t, server.URL, mailSender)
	result, err := service.RegisterAndGrantSubaccount(context.Background(), 1, IsmartSubaccountRegisterGrantParams{
		UnitID:    "0348200001",
		Phone:     "91234567",
		Email:     "Existing@Example.com",
		SendEmail: true,
		Remark:    "family member",
	})
	if err != nil {
		t.Fatalf("register and grant existing account: %v", err)
	}

	if result.AccountCreated || result.TargetUserID != 456 || result.TargetUsername != "200456" {
		t.Fatalf("unexpected existing account result: %+v", result)
	}
	if !result.EmailSent || result.TemporaryPasswordSent {
		t.Fatalf("unexpected existing account email state: %+v", result)
	}
	if len(mailSender.Messages) != 1 {
		t.Fatalf("mail count = %d, want 1", len(mailSender.Messages))
	}
	if strings.Contains(mailSender.Messages[0].Body, "臨時密碼") {
		t.Fatalf("existing account email must not contain temporary password: %s", mailSender.Messages[0].Body)
	}
	if grantPayload["user_id"] != float64(111) || grantPayload["target_user_id"] != float64(456) {
		t.Fatalf("unexpected grant payload: %+v", grantPayload)
	}
	if grantPayload["unit_id"] != "0348200001" || grantPayload["remark"] != "family member" {
		t.Fatalf("unexpected grant details: %+v", grantPayload)
	}
}

// 6. TestRegisterAndGrantSubaccountNewAccountSendsTemporaryPassword verifies registration and mail delivery.
func TestRegisterAndGrantSubaccountNewAccountSendsTemporaryPassword(t *testing.T) {
	var registerPayload map[string]any
	var grantPayload map[string]any
	server := newSubaccountIntegrationTestServer(t, func(path string, payload map[string]any) (int, map[string]any) {
		switch path {
		case "/api/v1/integration/auth/check-contact/":
			if _, ok := payload["email"]; ok {
				return http.StatusOK, map[string]any{"status": "success", "data": map[string]any{
					"email_available": true,
				}}
			}
			return http.StatusOK, map[string]any{"status": "success", "data": map[string]any{
				"phone_available": true,
			}}
		case "/api/v1/integration/auth/register/":
			registerPayload = payload
			return http.StatusOK, map[string]any{"status": "success", "data": map[string]any{
				"user_id":  789,
				"username": "200789",
				"email":    "new@example.com",
				"phone":    "+85291234567",
			}}
		case "/api/v1/integration/buildings/subaccounts/grant/":
			grantPayload = payload
			return http.StatusOK, map[string]any{"status": "success", "data": map[string]any{
				"upstream_message": "granted",
			}}
		default:
			t.Fatalf("unexpected upstream path %s", path)
			return http.StatusInternalServerError, nil
		}
	})
	defer server.Close()

	mailSender := &capturedSubaccountMailSender{}
	service := newSubaccountRegistrationTestService(t, server.URL, mailSender)
	result, err := service.RegisterAndGrantSubaccount(context.Background(), 1, IsmartSubaccountRegisterGrantParams{
		UnitID:    "0348200001",
		Phone:     "91234567",
		Email:     "new@example.com",
		EngName:   "CHAN TAI MAN",
		SendEmail: true,
		Remark:    "family member",
	})
	if err != nil {
		t.Fatalf("register and grant new account: %v", err)
	}

	if !result.AccountCreated || result.TargetUserID != 789 || result.TargetUsername != "200789" {
		t.Fatalf("unexpected new account result: %+v", result)
	}
	if !result.EmailSent || !result.TemporaryPasswordSent {
		t.Fatalf("temporary password email was not reported as sent: %+v", result)
	}
	if len(mailSender.Messages) != 1 {
		t.Fatalf("mail count = %d, want 1", len(mailSender.Messages))
	}
	password, ok := registerPayload["password"].(string)
	if !ok || password == "" {
		t.Fatalf("registration password missing: %+v", registerPayload)
	}
	if !strings.Contains(mailSender.Messages[0].Body, "臨時密碼："+password) {
		t.Fatalf("email does not contain the registration password: %s", mailSender.Messages[0].Body)
	}
	if !strings.Contains(mailSender.Messages[0].Body, "測試大廈") || !strings.Contains(mailSender.Messages[0].Body, "12 / A") {
		t.Fatalf("email does not contain unit display: %s", mailSender.Messages[0].Body)
	}
	if grantPayload["user_id"] != float64(111) || grantPayload["target_user_id"] != float64(789) {
		t.Fatalf("unexpected grant payload: %+v", grantPayload)
	}
	if registerPayload["phone"] != "+85291234567" || registerPayload["is_receive_email"] != true {
		t.Fatalf("unexpected registration contacts/settings: %+v", registerPayload)
	}
}

// 7. TestRegisterAndGrantSubaccountRejectsContactConflict verifies mismatched existing contacts stop before grant.
func TestRegisterAndGrantSubaccountRejectsContactConflict(t *testing.T) {
	var grantCalled bool
	server := newSubaccountIntegrationTestServer(t, func(path string, payload map[string]any) (int, map[string]any) {
		switch path {
		case "/api/v1/integration/auth/check-contact/":
			if _, ok := payload["email"]; ok {
				return http.StatusOK, map[string]any{"status": "success", "data": map[string]any{
					"email_available": false,
					"email_user_id":   456,
				}}
			}
			return http.StatusOK, map[string]any{"status": "success", "data": map[string]any{
				"phone_available": false,
				"phone_user_id":   999,
			}}
		case "/api/v1/integration/buildings/subaccounts/grant/":
			grantCalled = true
			return http.StatusOK, map[string]any{"status": "success", "data": map[string]any{}}
		default:
			t.Fatalf("unexpected upstream path %s", path)
			return http.StatusInternalServerError, nil
		}
	})
	defer server.Close()

	service := newSubaccountRegistrationTestService(t, server.URL, &capturedSubaccountMailSender{})
	_, err := service.RegisterAndGrantSubaccount(context.Background(), 1, IsmartSubaccountRegisterGrantParams{
		UnitID:    "0348200001",
		Phone:     "91234567",
		Email:     "conflict@example.com",
		EngName:   "CHAN TAI MAN",
		SendEmail: true,
	})
	if err == nil || !strings.Contains(err.Error(), "same existing ismart account") {
		t.Fatalf("expected contact conflict error, got %v", err)
	}
	if grantCalled {
		t.Fatal("grant must not be called when contacts belong to different users")
	}
}

// 8. newSubaccountIntegrationTestServer creates a deterministic upstream API simulator.
func newSubaccountIntegrationTestServer(t *testing.T, handler func(path string, payload map[string]any) (int, map[string]any)) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		defer request.Body.Close()
		payload := map[string]any{}
		if request.Method == http.MethodPost {
			if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
				t.Fatalf("decode upstream payload: %v", err)
			}
		}
		status, response := handler(request.URL.Path, payload)
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(status)
		if response != nil {
			if err := json.NewEncoder(writer).Encode(response); err != nil {
				t.Fatalf("encode upstream response: %v", err)
			}
		}
	}))
}

// 9. newSubaccountRegistrationTestService builds an isolated service with one visible owner unit.
func newSubaccountRegistrationTestService(t *testing.T, upstreamURL string, mailSender MailSender) *IsmartExternalService {
	t.Helper()
	databaseName := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	db, err := gorm.Open(sqlite.Open("file:ismart-subaccount-registration-test-"+databaseName+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.UserIsmartAccount{}); err != nil {
		t.Fatalf("migrate ismart account: %v", err)
	}
	account := model.UserIsmartAccount{
		UserID:                             1,
		IsmartUserID:                       111,
		Username:                           "200111",
		ClientBuildingFlatUnitsPermissions: datatypes.JSON([]byte(`["0348200001"]`)),
		Building:                           datatypes.JSON([]byte(`[]`)),
		StaffBuildingPermissions:           datatypes.JSON([]byte(`[]`)),
		ClientBuildingPermissions:          datatypes.JSON([]byte(`[]`)),
		RawMessage:                         datatypes.JSON([]byte(`{}`)),
		ProfileSnapshot:                    datatypes.JSON([]byte(`{}`)),
	}
	if err := db.Create(&account).Error; err != nil {
		t.Fatalf("create owner ismart account: %v", err)
	}

	cfg := &config.Config{
		AppPublicBaseURL:            "https://test.ajoliving.skylinedances.com",
		IsmartIntegrationAPIBaseURL: upstreamURL + "/api/v1/integration",
		IsmartSubaccountAPIBaseURL:  upstreamURL + "/api/v1/integration",
		IsmartExternalAppTimeout:    2 * time.Second,
		POSLoginTimeout:             2 * time.Second,
	}
	cache := &subaccountTestCache{values: map[string][]byte{
		posBuildingDirectoryCacheKey:               []byte(`[{"building_id":"0348200","buildname_chi":"測試大廈"}]`),
		posBuildingUnitsCacheKeyPrefix + "0348200": []byte(`[{"unit_id":"0348200001","floor":"12","unit":"A"}]`),
	}}
	return NewIsmartExternalService(&Runtime{
		Config:     cfg,
		DB:         db,
		MailSender: mailSender,
		CacheStore: cache,
	})
}

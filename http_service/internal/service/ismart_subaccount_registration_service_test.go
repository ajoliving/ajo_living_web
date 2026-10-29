/*
 * iSmart 副戶註冊、授權及普通通知電郵流程測試。
 * 1. 使用 httptest 模擬 iSmart check-contact、register 及 grant 接口。
 * 2. 使用內存 SQLite 提供業主的可見單位權限。
 * 3. 驗證手動密碼、新舊帳戶授權、聯絡資料衝突及非阻斷郵件分支。
 */
package service

import (
	"context"
	"encoding/json"
	"errors"
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

// 1.1 SendHTML captures one outgoing HTML subaccount notification.
func (s *capturedSubaccountMailSender) SendHTML(ctx context.Context, recipient string, subject string, body string) error {
	return s.Send(ctx, recipient, subject, body)
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
			return http.StatusOK, map[string]any{"status": "success", "data": map[string]any{
				"email_available": false,
				"email_user_id":   456,
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
	if !result.EmailSent {
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
	result, err = service.RegisterAndGrantSubaccount(context.Background(), 1, IsmartSubaccountRegisterGrantParams{
		UnitID:   "0348200001",
		Phone:    "91234567",
		Email:    "existing@example.com",
		Password: "x",
	})
	if err != nil || result.AccountCreated || result.EmailSent {
		t.Fatalf("existing account must ignore registration password: %+v, %v", result, err)
	}
}

// 6. TestRegisterAndGrantSubaccountNewAccount verifies manual passwords and optional non-blocking mail.
func TestRegisterAndGrantSubaccountNewAccount(t *testing.T) {
	for _, tc := range []struct {
		name          string
		password      string
		sendEmail     bool
		noMailSender  bool
		mailErr       error
		wantEmailSent bool
		wantError     bool
	}{
		{name: "manual password with mail", password: "ManualPass 123!", sendEmail: true, wantEmailSent: true},
		{name: "manual password without mail", password: "ManualPass 123!"},
		{name: "mail failure", password: "ManualPass 123!", sendEmail: true, mailErr: errors.New("SMTP unavailable")},
		{name: "mail sender unavailable", password: "ManualPass 123!", sendEmail: true, noMailSender: true},
		{name: "missing password", wantError: true},
		{name: "short password", password: "1234567", wantError: true},
		{name: "eight characters", password: "Abcd1234"},
		{name: "short unicode password", password: "密碼不足八字", wantError: true},
		{name: "blank password", password: "        ", wantError: true},
		{name: "leading whitespace", password: " ManualPass123!", wantError: true},
		{name: "trailing whitespace", password: "ManualPass123!\t", wantError: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var registerPayload map[string]any
			var grantPayload map[string]any
			server := newSubaccountIntegrationTestServer(t, func(path string, payload map[string]any) (int, map[string]any) {
				switch path {
				case "/api/v1/integration/auth/check-contact/":
					return http.StatusOK, map[string]any{"status": "success", "data": map[string]any{
						"email_available": true,
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

			mailSender := &capturedSubaccountMailSender{Err: tc.mailErr}
			var sender MailSender = mailSender
			if tc.noMailSender {
				sender = nil
			}
			service := newSubaccountRegistrationTestService(t, server.URL, sender)
			result, err := service.RegisterAndGrantSubaccount(context.Background(), 1, IsmartSubaccountRegisterGrantParams{
				UnitID:    "0348200001",
				Phone:     "91234567",
				Email:     "new@example.com",
				EngName:   "CHAN TAI MAN",
				Password:  tc.password,
				SendEmail: tc.sendEmail,
				Remark:    "family member",
			})
			if tc.wantError {
				if err == nil || !strings.Contains(err.Error(), "password must contain at least 8 characters") {
					t.Fatalf("expected password validation failure, got %v", err)
				}
				if registerPayload != nil || grantPayload != nil || len(mailSender.Messages) != 0 {
					t.Fatal("invalid passwords must not register, grant or send mail")
				}
				return
			}
			if err != nil {
				t.Fatalf("register and grant new account: %v", err)
			}

			if !result.AccountCreated || result.TargetUserID != 789 || result.TargetUsername != "200789" {
				t.Fatalf("unexpected new account result: %+v", result)
			}
			if result.EmailSent != tc.wantEmailSent {
				t.Fatalf("email sent = %v, want %v", result.EmailSent, tc.wantEmailSent)
			}
			wantMailCount := 0
			if tc.sendEmail && !tc.noMailSender {
				wantMailCount = 1
			}
			if len(mailSender.Messages) != wantMailCount {
				t.Fatalf("mail count = %d, want %d", len(mailSender.Messages), wantMailCount)
			}
			if registerPayload["password"] != tc.password {
				t.Fatal("registration password must be forwarded unchanged")
			}
			for _, message := range mailSender.Messages {
				if strings.Contains(message.Body, tc.password) || strings.Contains(message.Body, "臨時密碼") {
					t.Fatal("notification must not contain the registration password")
				}
				if !strings.Contains(message.Body, "測試大廈") || !strings.Contains(message.Body, "12 / A") ||
					!strings.Contains(message.Body, "自行設定的密碼登入") {
					t.Fatalf("email does not contain account and unit guidance: %s", message.Body)
				}
			}
			encodedResult, err := json.Marshal(result)
			if err != nil || strings.Contains(string(encodedResult), "temporary_password_sent") ||
				strings.Contains(string(encodedResult), tc.password) {
				t.Fatal("response must not contain a password or obsolete delivery state")
			}
			if grantPayload["user_id"] != float64(111) || grantPayload["target_user_id"] != float64(789) {
				t.Fatalf("unexpected grant payload: %+v", grantPayload)
			}
			if registerPayload["phone"] != "+85291234567" || registerPayload["is_receive_email"] != true {
				t.Fatalf("unexpected registration contacts/settings: %+v", registerPayload)
			}
		})
	}
}

// 7. TestRegisterAndGrantSubaccountRejectsContactConflict verifies mismatched existing contacts stop before grant.
func TestRegisterAndGrantSubaccountRejectsContactConflict(t *testing.T) {
	var grantCalled bool
	server := newSubaccountIntegrationTestServer(t, func(path string, payload map[string]any) (int, map[string]any) {
		switch path {
		case "/api/v1/integration/auth/check-contact/":
			return http.StatusOK, map[string]any{"status": "success", "data": map[string]any{
				"phone_available": false,
				"phone_user_id":   999,
				"email_available": false,
				"email_user_id":   456,
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

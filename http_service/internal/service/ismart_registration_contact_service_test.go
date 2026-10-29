/*
 * iSmart 聯絡資料檢查回歸測試。
 * 1. 驗證單次聯絡檢查與帳戶存在狀態。
 * 2. 驗證雙項帳戶比對、授權門檻與上游錯誤保真。
 * 3. 驗證通知電郵包含登入及單位資料，不包含密碼。
 */
package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ajoliving_web/http_service/internal/config"
)

// 1. 驗證正式上游回應的單項查詢與雙項帳戶比對
func TestIsmartSubaccountContactResultStates(t *testing.T) {
	for _, tc := range []struct {
		name     string
		phone    string
		email    string
		phoneID  int64
		emailID  int64
		sameUser bool
		conflict bool
	}{
		{name: "phone registered", phone: "+85261234567", phoneID: 12},
		{name: "phone unregistered", phone: "+85261234567"},
		{name: "email registered", email: "resident@example.com", emailID: 12},
		{name: "email unregistered", email: "resident@example.com"},
		{name: "same account", phone: "+85261234567", email: "resident@example.com", phoneID: 12, emailID: 12, sameUser: true},
		{name: "different accounts", phone: "+85261234567", email: "resident@example.com", phoneID: 12, emailID: 34, conflict: true},
		{name: "only email registered", phone: "+85261234567", email: "resident@example.com", emailID: 12, conflict: true},
		{name: "only phone registered", phone: "+85261234567", email: "resident@example.com", phoneID: 12, conflict: true},
		{name: "both unregistered", phone: "+85261234567", email: "resident@example.com"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			requestCount := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost || r.URL.Path != "/auth/check-contact/" {
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
				}
				requestCount++
				var input map[string]string
				if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
					t.Errorf("decode request: %v", err)
				}
				data := map[string]any{}
				if tc.phone != "" {
					if input["phone"] != tc.phone {
						t.Errorf("unexpected phone lookup: %v", input)
					}
					data["phone_available"], data["phone_user_id"] = tc.phoneID == 0, tc.phoneID
				}
				if tc.email != "" {
					if input["email"] != tc.email {
						t.Errorf("unexpected email lookup: %v", input)
					}
					data["email_available"], data["email_user_id"] = tc.emailID == 0, tc.emailID
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "data": data})
			}))
			defer server.Close()
			service := NewIsmartExternalService(&Runtime{Config: &config.Config{IsmartSubaccountAPIBaseURL: server.URL}})
			result, err := service.CheckSubaccountContacts(context.Background(), IsmartSubaccountContactParams{Phone: tc.phone, Email: tc.email})
			if err != nil {
				t.Fatalf("contact check: %v", err)
			}
			if result.PhoneExists != (tc.phoneID > 0) || result.EmailExists != (tc.emailID > 0) ||
				result.SameUser != tc.sameUser || result.CanAuthorize != tc.sameUser ||
				(result.ConflictReason != "") != tc.conflict {
				t.Fatalf("unexpected result: %+v", result)
			}
			if requestCount != 1 {
				t.Fatalf("contact lookup count = %d, want 1", requestCount)
			}
		})
	}
}

// 2. 驗證單次聯絡檢查使用規範化電話
func TestIsmartSubaccountContactPhoneUsesOneNormalizedLookup(t *testing.T) {
	phoneRequests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var input map[string]string
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		phoneRequests++
		if input["phone"] != "+85291234567" {
			t.Fatalf("phone = %q, want normalized E.164", input["phone"])
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status": "success",
			"data": map[string]any{
				"phone_available": true,
			},
		})
	}))
	defer server.Close()

	service := NewIsmartExternalService(&Runtime{Config: &config.Config{IsmartSubaccountAPIBaseURL: server.URL}})
	result, err := service.CheckSubaccountContacts(context.Background(), IsmartSubaccountContactParams{Phone: "91234567"})
	if err != nil {
		t.Fatalf("contact check: %v", err)
	}
	if result.PhoneExists {
		t.Fatal("unregistered phone should not be marked as existing")
	}
	if phoneRequests != 1 {
		t.Fatalf("phone lookup count = %d, want 1", phoneRequests)
	}
}

// 3. 上游檢查失敗不得回傳未註冊結果
func TestIsmartSubaccountContactFailureIsNotUnregistered(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "error", "message": "upstream unavailable"})
	}))
	defer server.Close()
	service := NewIsmartExternalService(&Runtime{Config: &config.Config{IsmartSubaccountAPIBaseURL: server.URL}})
	for _, params := range []IsmartSubaccountContactParams{{Phone: "+85261234567"}, {Email: "resident@example.com"}} {
		result, err := service.CheckSubaccountContacts(context.Background(), params)
		if err == nil || result != nil {
			t.Fatalf("upstream error must not become a lookup result: %+v, %v", result, err)
		}
	}
}

// 4. 驗證正式郵件包含帳戶及單位資料，不包含密碼
func TestBuildSubaccountEmailBodyExcludesPassword(t *testing.T) {
	body := buildSubaccountEmailBody(&IsmartSubaccountRegisterGrantResult{
		AccountCreated: true,
		TargetUsername: "200123",
		BuildingName:   "測試大廈",
		UnitDisplay:    "12 / A",
	}, "user@example.com")
	for _, value := range []string{"測試大廈", "12 / A", "200123", "user@example.com", "自行設定的密碼登入"} {
		if !strings.Contains(body, value) {
			t.Fatalf("email body does not contain %q: %s", value, body)
		}
	}
	if strings.Contains(body, "臨時密碼") || strings.Contains(body, "密碼：") {
		t.Fatal("notification must not contain a password")
	}
}

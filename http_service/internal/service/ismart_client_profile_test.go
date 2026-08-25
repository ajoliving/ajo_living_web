/*
 * iSmart ClientTbl profile synchronization tests.
 * 1. Verify the linked upstream identity is used for the client request.
 * 2. Verify ClientTbl fields update the local snapshot and /me response.
 * 3. Verify sensitive upstream values are excluded from stored member data.
 */
package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ajoliving_web/http_service/internal/model"
)

// 1. TestGetMeRefreshesIsmartClientProfile verifies /me refreshes the linked ClientTbl snapshot.
func TestGetMeRefreshesIsmartClientProfile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet || request.URL.Path != "/api/v1/integration/auth/client/" {
			t.Fatalf("unexpected client request: %s %s", request.Method, request.URL.String())
		}
		if request.URL.Query().Get("user_id") != "88" {
			t.Fatalf("expected linked iSmart user id, got %q", request.URL.Query().Get("user_id"))
		}
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`{"status":"success","data":{"user_id":88,"username":"03141139","email":"977052wk@gmail.com","phone":"+85296022060","client":{"cli_id":"03141139","cli_legalentity":"NA","cli_type":"personal","cli_contact_person":"CHAN CONTACT","cli_urgent_contact_person":"CHAN EMERGENCY","cli_name":"CHAN TAI MAN","cli_chi_name":"陳大文","cli_acname":"CHAN TAI MAN","cli_id_card":"A1234567","cli_tel":"+85296022060","cli_tel2":"+85291230000","cli_addr":"1 Example Road","cli_chiadd":"時安大廈","cli_email":"977052wk@gmail.com","cli_birthday":"1990-01-01","cli_sex":"M","password":"must-not-persist"}}}`))
	}))
	defer server.Close()

	runtimeValue, user := newIsmartIntegrationTestRuntime(t, server.URL)
	result, err := NewUserService(runtimeValue).GetMe(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("get member profile: %v", err)
	}
	if result.IsmartProfileSyncStatus != "synced" || result.IsmartAccount == nil {
		t.Fatalf("expected refreshed iSmart account profile, got %#v", result)
	}
	profile := result.IsmartAccount
	if profile.AccountCode != "03141139" ||
		profile.AccountPhone != "+85296022060" ||
		profile.AccountEmail != "977052wk@gmail.com" ||
		profile.OwnerNameEN != "CHAN TAI MAN" ||
		profile.OwnerNameZH != "陳大文" ||
		profile.AccountName != "CHAN TAI MAN" ||
		profile.IdentityNumber != "A1234567" ||
		profile.LegalEntity != "NA" ||
		profile.ClientType != "personal" ||
		profile.BirthDate != "1990-01-01" ||
		profile.ContactName != "CHAN CONTACT" ||
		profile.ContactPhone != "+85296022060" ||
		profile.EmergencyContactName != "CHAN EMERGENCY" ||
		profile.EmergencyContactPhone != "+85291230000" ||
		profile.BillingPhone != "+85296022060" ||
		profile.BillingEmail != "977052wk@gmail.com" ||
		profile.BillingAddress != "1 Example Road" ||
		profile.BillingAddressEN != "1 Example Road" ||
		profile.BillingAddressZH != "時安大廈" ||
		profile.Gender != "M" {
		t.Fatalf("unexpected refreshed iSmart profile: %#v", profile)
	}

	var account model.UserIsmartAccount
	if err := runtimeValue.DB.Where("user_id = ?", user.ID).First(&account).Error; err != nil {
		t.Fatalf("load refreshed iSmart account: %v", err)
	}
	if account.Username != "03141139" || account.Email != "977052wk@gmail.com" || account.Phone != "+85296022060" || account.ClientProfileSyncedAt == nil {
		t.Fatalf("unexpected saved iSmart account: %#v", account)
	}
	var raw map[string]any
	if err := json.Unmarshal(account.RawMessage, &raw); err != nil {
		t.Fatalf("decode saved iSmart payload: %v", err)
	}
	client := paymentMapValue(raw["client"])
	if _, exists := client["password"]; exists {
		t.Fatalf("client snapshot must not retain password: %#v", client)
	}
}

// 2. TestUpdateClientProfileUsesLinkedIdentity verifies AJO writes only to the linked iSmart account.
func TestUpdateClientProfileUsesLinkedIdentity(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPatch || request.URL.Path != "/api/v1/integration/auth/client/" {
			t.Fatalf("unexpected client update request: %s %s", request.Method, request.URL.String())
		}
		var payload map[string]any
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			t.Fatalf("decode client update request: %v", err)
		}
		if paymentInt64Value(payload["user_id"]) != 88 {
			t.Fatalf("expected linked iSmart user id, got %#v", payload["user_id"])
		}
		if payload["cli_chi_name"] != "陳更新" || payload["cli_id"] != nil {
			t.Fatalf("unexpected client update payload: %#v", payload)
		}
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`{"status":"success","data":{"user_id":88,"username":"03141139","email":"member@example.com","phone":"+85296022060","client":{"cli_id":"03141139","cli_name":"CHAN UPDATED","cli_chi_name":"陳更新","cli_acname":"CHAN UPDATED","cli_id_card":"A1234567","cli_tel":"+85296022060","cli_email":"member@example.com"}}}`))
	}))
	defer server.Close()

	runtimeValue, user := newIsmartIntegrationTestRuntime(t, server.URL)
	result, err := NewIsmartExternalService(runtimeValue).UpdateClientProfile(context.Background(), user.ID, IsmartClientProfileUpdate{
		OwnerNameZH: pointerToString("陳更新"),
	})
	if err != nil {
		t.Fatalf("update client profile: %v", err)
	}
	if paymentStringValue(result["username"]) != "03141139" {
		t.Fatalf("unexpected update result: %#v", result)
	}

	profile := NewUserService(runtimeValue).loadIsmartAccountProfile(context.Background(), user.ID)
	if profile == nil || profile.OwnerNameZH != "陳更新" {
		t.Fatalf("expected updated iSmart profile snapshot, got %#v", profile)
	}
}

// 3. pointerToString creates an optional string test value.
func pointerToString(value string) *string {
	return &value
}

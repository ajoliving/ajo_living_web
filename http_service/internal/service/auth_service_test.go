/*
 * Authentication service tests.
 * 1. Validate resident registration credential persistence.
 * 2. Validate username password sign-in uses the same token flow.
 * 3. Validate registration profile fields and credentials.
 * 4. Validate email password reset updates stored credentials.
 * 5. Validate account type registration and derived publisher identity compatibility.
 */
package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"ajoliving_web/http_service/internal/config"
	"ajoliving_web/http_service/internal/model"
	"ajoliving_web/http_service/internal/utils"
)

// 1. TestAuthServiceRegistersAccountTypes verifies registration status for all account categories.
func TestAuthServiceRegistersAccountTypes(t *testing.T) {
	runtimeValue := newAuthTestRuntime(t, nil)
	auth := NewAuthService(runtimeValue)
	tests := []struct{ accountType, phone, status string }{{AccountTypePersonal, "61110001", "active"}, {AccountTypeIndividualAgent, "61110002", "pending_profile"}, {AccountTypeAgencyCompany, "61110003", "pending_profile"}}
	for _, item := range tests {
		result, err := auth.RegisterWithEmail(context.Background(), EmailPasswordParams{Email: "user" + item.phone + "@example.com", Password: "password123", EngName: "Test User", Username: "user" + item.phone, PhoneCountryCode: "+852", PhoneNumber: item.phone, AccountType: item.accountType})
		if err != nil {
			t.Fatalf("register %s: %v", item.accountType, err)
		}
		if result.User.AccountType != item.accountType || result.User.MemberStatus != item.status {
			t.Fatalf("unexpected %s registration: %#v", item.accountType, result.User)
		}
	}
}

// 2. TestAuthServiceRejectsRegistrationWithoutEmail verifies required credentials are validated before duplicate checks.
func TestAuthServiceRejectsRegistrationWithoutEmail(t *testing.T) {
	runtimeValue := newAuthTestRuntime(t, nil)
	auth := NewAuthService(runtimeValue)

	_, err := auth.RegisterWithEmail(context.Background(), EmailPasswordParams{
		Password:         "password123",
		EngName:          "Test User",
		Username:         "existing-user",
		PhoneCountryCode: "+852",
		PhoneNumber:      "61110004",
		AccountType:      AccountTypePersonal,
	})
	if err == nil || err.Error() != "valid email is required" {
		t.Fatalf("expected missing email validation error, got %v", err)
	}
}

// 2.1 TestAuthServiceAllowsIndividualAgentRegistrationWithoutEmail verifies individual agents may use their licence username without an email address.
func TestAuthServiceAllowsIndividualAgentRegistrationWithoutEmail(t *testing.T) {
	runtimeValue := newAuthTestRuntime(t, nil)
	result, err := NewAuthService(runtimeValue).RegisterWithEmail(context.Background(), EmailPasswordParams{
		Password: "password123", EngName: "CHAN TAI MAN", Username: "E-123456",
		PhoneCountryCode: "+852", PhoneNumber: "61110014", AccountType: AccountTypeIndividualAgent,
	})
	if err != nil || result.User.MemberStatus != "pending_profile" {
		t.Fatalf("expected pending individual agent registration without email: %#v %v", result, err)
	}
	var credential model.UserCredential
	if err := runtimeValue.DB.Where("user_id = ?", 1).First(&credential).Error; err != nil || credential.Email != nil {
		t.Fatalf("expected empty email credential: %#v %v", credential, err)
	}
}

// 2.2 TestAuthServiceChecksRegistrationAvailability verifies email and phone are checked independently before registration.
func TestAuthServiceChecksRegistrationAvailability(t *testing.T) {
	runtimeValue := newAuthTestRuntime(t, nil)
	auth := NewAuthService(runtimeValue)
	if _, err := auth.RegisterWithEmail(context.Background(), EmailPasswordParams{
		Email: "registered@example.com", Password: "password123", EngName: "Registered User", Username: "registered-user",
		PhoneCountryCode: "+852", PhoneNumber: "61110015", AccountType: AccountTypePersonal,
	}); err != nil {
		t.Fatalf("register account: %v", err)
	}

	availability, err := auth.CheckRegistrationAvailability(context.Background(), RegistrationAvailabilityParams{
		Email: "REGISTERED@example.com", PhoneCountryCode: "+852", PhoneNumber: "61110015",
	})
	if err != nil {
		t.Fatalf("check used identities: %v", err)
	}
	if availability.EmailAvailable || availability.PhoneAvailable {
		t.Fatalf("expected used identities to be unavailable: %#v", availability)
	}

	availability, err = auth.CheckRegistrationAvailability(context.Background(), RegistrationAvailabilityParams{
		Email: "available@example.com", PhoneCountryCode: "+852", PhoneNumber: "61110016",
	})
	if err != nil {
		t.Fatalf("check available identities: %v", err)
	}
	if !availability.EmailAvailable || !availability.PhoneAvailable {
		t.Fatalf("expected unused identities to be available: %#v", availability)
	}
}

// 2.3 TestAuthServiceChecksIsmartRegistrationAvailability verifies iSmart contact conflicts block the AJO precheck.
func TestAuthServiceChecksIsmartRegistrationAvailability(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/api/v1/integration/auth/check-contact/" {
			t.Fatalf("unexpected iSmart path: %s", request.URL.Path)
		}
		var payload map[string]any
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			t.Fatalf("decode iSmart contact request: %v", err)
		}
		if payload["email"] != "used-by-ismart@example.com" || payload["phone"] != "+85261110017" {
			t.Fatalf("unexpected iSmart contact payload: %#v", payload)
		}
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`{"status":"success","data":{"email_available":false,"phone_available":true}}`))
	}))
	defer server.Close()

	runtimeValue := newAuthTestRuntime(t, &config.Config{IsmartIntegrationAPIBaseURL: server.URL + "/api/v1/integration"})
	availability, err := NewAuthService(runtimeValue).CheckRegistrationAvailability(context.Background(), RegistrationAvailabilityParams{
		Email: "used-by-ismart@example.com", PhoneCountryCode: "+852", PhoneNumber: "61110017",
	})
	if err != nil {
		t.Fatalf("check iSmart availability: %v", err)
	}
	if availability.EmailAvailable || !availability.PhoneAvailable {
		t.Fatalf("expected iSmart email conflict only: %#v", availability)
	}
}

// 3. TestAuthServiceRegistersAndLinksIsmartAccount verifies registration uses the direct iSmart API and saves its identity.
func TestAuthServiceRegistersAndLinksIsmartAccount(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/api/v1/integration/auth/client/" {
			if request.URL.Query().Get("user_id") != "88" {
				t.Fatalf("unexpected iSmart client user id: %s", request.URL.Query().Get("user_id"))
			}
			response.Header().Set("Content-Type", "application/json")
			_, _ = response.Write([]byte(`{"status":"success","data":{"user_id":88,"username":"200123","phone":"61110005","email":"agent@example.com","client":{"cli_id":"200123","cli_legalentity":"LE","cli_name":"CHAN T. M.","cli_chi_name":"陳大文","cli_id_card":"A1234567","cli_tel":"61110005","cli_email":"agent@example.com","cli_sex":"M"}}}`))
			return
		}
		if request.URL.Path != "/api/v1/integration/auth/register/" {
			t.Fatalf("unexpected iSmart path: %s", request.URL.Path)
		}
		var payload map[string]any
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			t.Fatalf("decode iSmart registration payload: %v", err)
		}
		if payload["phone"] != "61110005" || payload["email"] != "agent@example.com" || payload["eng_name"] != "CHAN TAI MAN" || payload["chi_name"] != "陳大文" {
			t.Fatalf("unexpected required iSmart payload: %#v", payload)
		}
		if payload["legal_entity"] != "LE" || payload["id_card"] != "A1234567" || payload["remark"] != "AJO registration" || payload["gender"] != "M" || payload["is_receive_email"] != false || payload["password"] != "password123" {
			t.Fatalf("unexpected optional iSmart payload: %#v", payload)
		}
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`{"status":"success","data":{"user_id":88,"username":"200123","phone":"61110005","email":"agent@example.com","owner_name_en":"CHAN T. M."}}`))
	}))
	defer server.Close()

	runtimeValue := newAuthTestRuntime(t, &config.Config{IsmartIntegrationAPIBaseURL: server.URL + "/api/v1/integration"}, &model.User{}, &model.UserCredential{}, &model.UserProfile{}, &model.UserIsmartAccount{}, &model.Community{})
	auth := NewAuthService(runtimeValue)
	registered, err := auth.RegisterWithEmail(context.Background(), EmailPasswordParams{
		Email:            "agent@example.com",
		Password:         "password123",
		EngName:          "CHAN TAI MAN",
		ChiName:          "陳大文",
		Username:         "ajo-agent",
		PhoneCountryCode: "+852",
		PhoneNumber:      "61110005",
		AccountType:      AccountTypeIndividualAgent,
		IDCard:           "A1234567",
		Remark:           "AJO registration",
		Gender:           "m",
		IsReceiveEmail:   boolPointer(false),
	})
	if err != nil {
		t.Fatalf("register account: %v", err)
	}
	if registered.User.MemberStatus != "pending_profile" {
		t.Fatalf("expected pending agent account, got %#v", registered.User)
	}

	var ismartAccount model.UserIsmartAccount
	if err := runtimeValue.DB.Where("user_id = ?", 1).First(&ismartAccount).Error; err != nil {
		t.Fatalf("load linked iSmart account: %v", err)
	}
	if ismartAccount.IsmartUserID != 88 || ismartAccount.Username != "200123" {
		t.Fatalf("unexpected linked iSmart account: %#v", ismartAccount)
	}
	var rawProfile map[string]any
	if err := json.Unmarshal(ismartAccount.RawMessage, &rawProfile); err != nil {
		t.Fatalf("decode iSmart profile snapshot: %v", err)
	}
	if _, exists := rawProfile["password"]; exists {
		t.Fatalf("iSmart profile snapshot must not expose a password: %#v", rawProfile)
	}
	memberProfile, err := NewUserService(runtimeValue).GetMe(context.Background(), 1)
	if err != nil || memberProfile.IsmartAccount == nil {
		t.Fatalf("load iSmart member profile: %#v %v", memberProfile, err)
	}
	ismartProfile := memberProfile.IsmartAccount
	if ismartProfile.AccountCode != "200123" ||
		ismartProfile.AccountPhone != "61110005" ||
		ismartProfile.AccountEmail != "agent@example.com" ||
		ismartProfile.OwnerNameEN != "CHAN T. M." ||
		ismartProfile.OwnerNameZH != "陳大文" ||
		ismartProfile.IdentityNumber != "A1234567" ||
		ismartProfile.LegalEntity != "LE" ||
		ismartProfile.Gender != "M" ||
		ismartProfile.ContactPhone != "61110005" ||
		ismartProfile.BillingEmail != "agent@example.com" {
		t.Fatalf("unexpected iSmart member profile: %#v", ismartProfile)
	}
	if err := runtimeValue.DB.Transaction(func(tx *gorm.DB) error {
		return auth.saveIsmartAccount(context.Background(), tx, 1, &IsmartMessage{UserID: 88, Username: "200123", Email: "agent@example.com", Phone: "61110005"}, "", "")
	}); err != nil {
		t.Fatalf("save POS login snapshot: %v", err)
	}
	memberProfile, err = NewUserService(runtimeValue).GetMe(context.Background(), 1)
	if err != nil || memberProfile.IsmartAccount == nil || memberProfile.IsmartAccount.OwnerNameZH != "陳大文" || memberProfile.IsmartAccount.IdentityNumber != "A1234567" {
		t.Fatalf("POS login must preserve the registered iSmart profile: %#v %v", memberProfile, err)
	}
	loggedIn, err := auth.LoginWithUsername(context.Background(), EmailPasswordParams{Username: "ajo-agent", Password: "password123"})
	if err != nil || loggedIn.User.PublicID != registered.User.PublicID {
		t.Fatalf("local login after iSmart registration failed: %#v %v", loggedIn, err)
	}
}

// 4. TestAuthServiceRejectsRegistrationWhenIsmartFails verifies failed upstream registration leaves no local account.
func TestAuthServiceRejectsRegistrationWhenIsmartFails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("Content-Type", "application/json")
		response.WriteHeader(http.StatusServiceUnavailable)
		_, _ = response.Write([]byte(`{"message":"iSmart unavailable"}`))
	}))
	defer server.Close()

	runtimeValue := newAuthTestRuntime(t, &config.Config{IsmartIntegrationAPIBaseURL: server.URL + "/api/v1/integration"}, &model.User{}, &model.UserCredential{}, &model.UserProfile{}, &model.UserIsmartAccount{}, &model.Community{})
	_, err := NewAuthService(runtimeValue).RegisterWithEmail(context.Background(), EmailPasswordParams{
		Email:            "failed@example.com",
		Password:         "password123",
		EngName:          "Failed Account",
		Username:         "failed-account",
		PhoneCountryCode: "+852",
		PhoneNumber:      "61110006",
		AccountType:      AccountTypePersonal,
	})
	if err == nil {
		t.Fatal("expected iSmart registration failure")
	}
	var count int64
	if err := runtimeValue.DB.Model(&model.User{}).Count(&count).Error; err != nil {
		t.Fatalf("count users: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected no local user after iSmart failure, got %d", count)
	}
}

// 4.1 TestAuthServiceRegistrationSubmitsPendingResidenceBinding verifies OwnerReg accepts HTTP 200 without waiting for approval.
func TestAuthServiceRegistrationSubmitsPendingResidenceBinding(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("Content-Type", "application/json")
		switch request.URL.Path {
		case "/api/v1/integration/auth/register/":
			_, _ = response.Write([]byte(`{"status":"success","data":{"user_id":91,"username":"200191","phone":"61110011","email":"binding@example.com"}}`))
		case "/pos/login":
			_, _ = response.Write([]byte(`{"token":"pos-service-token"}`))
		case "/pos/building/0348200/units":
			_, _ = response.Write([]byte(`[{"unit_id":"0348200001","floor":"12","unit":"A"}]`))
		case "/api/v1/integration/buildings/building-flat-owner-binding-requests/":
			var payload map[string]any
			if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
				t.Fatalf("decode owner binding payload: %v", err)
			}
			if payload["building_id"] != "0348200" || !reflect.DeepEqual(payload["ownedflat"], []any{"0348200001"}) || payload["user"] != float64(91) {
				t.Fatalf("unexpected owner binding payload: %#v", payload)
			}
			if payload["cli_role"] != "業主" || payload["cli_name"] != "CHAN TAI MAN" || payload["cli_id_card"] != "A1234567" {
				t.Fatalf("missing applicant data in owner binding payload: %#v", payload)
			}
			// OwnerReg's business status can still be pending; registration only waits for HTTP acceptance.
			_, _ = response.Write([]byte(`{"status":"error","message":"pending approval"}`))
		default:
			response.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	runtimeValue := newAuthTestRuntime(t, &config.Config{
		IsmartIntegrationAPIBaseURL: server.URL + "/api/v1/integration",
		POSAPIBaseURL:               server.URL + "/pos",
		POSAPIUsername:              "service-account",
		POSAPIPassword:              "service-password",
	}, &model.User{}, &model.UserCredential{}, &model.UserProfile{}, &model.UserIsmartAccount{}, &model.Community{})
	registered, err := NewAuthService(runtimeValue).RegisterWithEmail(context.Background(), EmailPasswordParams{
		Email:                "binding@example.com",
		Password:             "password123",
		EngName:              "CHAN TAI MAN",
		PhoneCountryCode:     "+852",
		PhoneNumber:          "61110011",
		IDCard:               "A1234567",
		PrimaryCommunityID:   "0348200",
		PrimaryCommunityName: "示例大廈",
		ResidenceFloor:       "12",
		ResidenceUnit:        "A",
	})
	if err != nil {
		t.Fatalf("register account with property request: %v", err)
	}
	if registered.User.ProfileCompleted {
		t.Fatalf("pending residence application must not complete the profile: %#v", registered.User)
	}

	var profile model.UserProfile
	if err := runtimeValue.DB.Preload("PrimaryCommunity").Where("user_id = ?", 1).First(&profile).Error; err != nil {
		t.Fatalf("load pending profile: %v", err)
	}
	if profile.PrimaryCommunity == nil || profile.PrimaryCommunity.PublicID != "0348200" || profile.ResidenceFloor != "12" || profile.ResidenceUnit != "A" || profile.ResidenceBindingStatus != residenceBindingStatusPending {
		t.Fatalf("expected saved pending property request, got %#v", profile)
	}
	if len(unmarshalStringSlice(profile.BoundBuildingIDs)) != 0 || len(unmarshalStringSlice(profile.BoundFlatUnitIDs)) != 0 {
		t.Fatalf("pending property request must not create approved bindings: %#v", profile)
	}
	me, err := NewUserService(runtimeValue).GetMe(context.Background(), 1)
	if err != nil {
		t.Fatalf("load pending member response: %v", err)
	}
	if me.ResidenceBindingStatus != residenceBindingStatusPending || len(me.BoundBuildingIDs) != 0 || len(me.BoundFlatUnitIDs) != 0 {
		t.Fatalf("member response must not report a pending request as bound: %#v", me)
	}
}

// 4.2 TestAuthServiceRegistrationKeepsAccountWhenResidenceBindingFails verifies a failed request remains recoverable through later login.
func TestAuthServiceRegistrationKeepsAccountWhenResidenceBindingFails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("Content-Type", "application/json")
		switch request.URL.Path {
		case "/api/v1/integration/auth/register/":
			_, _ = response.Write([]byte(`{"status":"success","data":{"user_id":92,"username":"200192","phone":"61110012","email":"binding-failed@example.com"}}`))
		case "/pos/login":
			_, _ = response.Write([]byte(`{"token":"pos-service-token"}`))
		case "/pos/building/0348200/units":
			_, _ = response.Write([]byte(`[{"unit_id":"0348200002","floor":"12","unit":"B"}]`))
		case "/api/v1/integration/buildings/building-flat-owner-binding-requests/":
			response.WriteHeader(http.StatusBadGateway)
			_, _ = response.Write([]byte(`{"message":"upstream unavailable"}`))
		default:
			response.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	runtimeValue := newAuthTestRuntime(t, &config.Config{
		IsmartIntegrationAPIBaseURL: server.URL + "/api/v1/integration",
		POSAPIBaseURL:               server.URL + "/pos",
		POSAPIUsername:              "service-account",
		POSAPIPassword:              "service-password",
	}, &model.User{}, &model.UserCredential{}, &model.UserProfile{}, &model.UserIsmartAccount{}, &model.Community{})
	_, err := NewAuthService(runtimeValue).RegisterWithEmail(context.Background(), EmailPasswordParams{
		Email:              "binding-failed@example.com",
		Password:           "password123",
		EngName:            "CHAN TAI MAN",
		PhoneCountryCode:   "+852",
		PhoneNumber:        "61110012",
		PrimaryCommunityID: "0348200",
		ResidenceFloor:     "12",
		ResidenceUnit:      "B",
	})
	if err == nil || err.Error() != "account created but property binding request failed" {
		t.Fatalf("expected explicit recoverable binding failure, got %v", err)
	}

	var profile model.UserProfile
	if err := runtimeValue.DB.Where("user_id = ?", 1).First(&profile).Error; err != nil {
		t.Fatalf("load retained profile: %v", err)
	}
	if profile.PrimaryCommunityID != nil || profile.ResidenceFloor != "" || profile.ResidenceUnit != "" || profile.ResidenceBindingStatus != "" {
		t.Fatalf("failed request must not save a false property binding: %#v", profile)
	}
	var ismartAccount model.UserIsmartAccount
	if err := runtimeValue.DB.Where("user_id = ?", 1).First(&ismartAccount).Error; err != nil || ismartAccount.IsmartUserID != 92 {
		t.Fatalf("local account and iSmart relation must remain for retry: %#v %v", ismartAccount, err)
	}
	if _, err := NewAuthService(runtimeValue).LoginWithEmail(context.Background(), EmailPasswordParams{Email: "binding-failed@example.com", Password: "password123"}); err != nil {
		t.Fatalf("retained account must be able to sign in for a later binding retry: %v", err)
	}
}

// 5. boolPointer returns one boolean pointer for optional request tests.
func boolPointer(value bool) *bool {
	return &value
}

// 6. TestAuthServiceAllowsRejectedAgencyStatusLogin verifies rejected agents can inspect review progress.
func TestAuthServiceAllowsRejectedAgencyStatusLogin(t *testing.T) {
	runtimeValue := newAuthTestRuntime(t, nil)
	auth := NewAuthService(runtimeValue)
	registered, err := auth.RegisterWithEmail(context.Background(), EmailPasswordParams{
		Email:            "rejected-agent@example.com",
		Password:         "password123",
		EngName:          "Rejected Agent",
		Username:         "rejected-agent",
		PhoneCountryCode: "+852",
		PhoneNumber:      "61110009",
		AccountType:      AccountTypeIndividualAgent,
	})
	if err != nil {
		t.Fatal(err)
	}
	var user model.User
	if err := runtimeValue.DB.Where("public_id = ?", registered.User.PublicID).First(&user).Error; err != nil {
		t.Fatal(err)
	}
	if err := runtimeValue.DB.Model(&user).Update("member_status", "rejected").Error; err != nil {
		t.Fatal(err)
	}

	results := []struct {
		name  string
		login func() (*VerifyOTPResult, error)
	}{
		{"email", func() (*VerifyOTPResult, error) {
			return auth.LoginWithEmail(context.Background(), EmailPasswordParams{Email: "rejected-agent@example.com", Password: "password123"})
		}},
		{"username", func() (*VerifyOTPResult, error) {
			return auth.LoginWithUsername(context.Background(), EmailPasswordParams{Username: "rejected-agent", Password: "password123"})
		}},
		{"phone", func() (*VerifyOTPResult, error) {
			return auth.LoginWithPhone(context.Background(), PhonePasswordParams{PhoneCountryCode: "+852", PhoneNumber: "61110009", Password: "password123"})
		}},
	}
	for _, item := range results {
		result, err := item.login()
		if err != nil || result.AccessToken == "" || result.User.MemberStatus != "rejected" {
			t.Fatalf("%s rejected agency login failed: %#v %v", item.name, result, err)
		}
	}
}

// 1. newAuthTestRuntime creates an isolated auth runtime.
func newAuthTestRuntime(t *testing.T, cfg *config.Config, models ...any) *Runtime {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+utils.NewPublicID()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if len(models) == 0 {
		models = []any{&model.User{}, &model.UserCredential{}, &model.UserProfile{}, &model.Community{}}
	}
	if err := db.AutoMigrate(models...); err != nil {
		t.Fatalf("migrate db: %v", err)
	}
	if cfg == nil {
		cfg = &config.Config{}
	}
	if cfg.JWTSecret == "" {
		cfg.JWTSecret = "test-jwt-secret"
	}
	if cfg.EncryptionKey == "" {
		cfg.EncryptionKey = "test-encryption-key"
	}

	return &Runtime{
		Config:   cfg,
		DB:       db,
		OTPStore: NewOTPStore(),
		Now:      time.Now,
	}
}

// 2. newAuthServiceTestServer creates POS login and relay stubs.
func newAuthServiceTestServer(t *testing.T, relayStatus int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("Content-Type", "application/json")
		if request.URL.Path == "/poslogin" {
			var payload map[string]string
			if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
				t.Fatalf("decode pos login payload: %v", err)
			}
			if strings.TrimSpace(payload["login_name"]) != "patrick" || payload["password"] != "s61980774" {
				response.WriteHeader(http.StatusUnauthorized)
				_, _ = response.Write([]byte(`{"code":0,"message":"invalid"}`))
				return
			}
			_, _ = response.Write([]byte(`{"code":1,"message":"success","msg":{"user_id":88,"username":"patrick","email":"patrick@example.com","phone":"+85261234567","is_staff":false,"client_building_permissions":["BLG-001"],"client_building_flat_units_permissions":["BLG-0010000101"]}}`))
			return
		}
		if request.URL.Path == "/login" {
			response.WriteHeader(relayStatus)
			if relayStatus >= http.StatusOK && relayStatus < http.StatusMultipleChoices {
				_, _ = response.Write([]byte(`{"token":"relay-token"}`))
				return
			}
			_, _ = response.Write([]byte(`{"message":"relay unavailable"}`))
			return
		}
		response.WriteHeader(http.StatusNotFound)
	}))
}

// 3. TestAuthServiceRegisterAndLoginWithUsername validates local resident username login.
func TestAuthServiceRegisterAndLoginWithUsername(t *testing.T) {
	runtimeValue := newAuthTestRuntime(t, nil)
	authService := NewAuthService(runtimeValue)

	registered, err := authService.RegisterWithEmail(context.Background(), EmailPasswordParams{
		Email:            "resident@example.com",
		Password:         "password123",
		DisplayName:      "Resident One",
		Username:         "ResidentOne",
		PhoneCountryCode: "+852",
		PhoneNumber:      "61234567",
	})
	if err != nil {
		t.Fatalf("register account: %v", err)
	}
	if registered.AccessToken == "" || registered.RefreshToken == "" {
		t.Fatalf("expected tokens after registration")
	}
	if _, err := authService.AuthenticateToken(context.Background(), registered.AccessToken); err != nil {
		t.Fatalf("access token should authenticate: %v", err)
	}
	if _, err := authService.AuthenticateToken(context.Background(), registered.RefreshToken); err == nil {
		t.Fatal("refresh token must not authenticate as an access token")
	}
	var registeredUser model.User
	if err := runtimeValue.DB.Where("public_id = ?", registered.User.PublicID).First(&registeredUser).Error; err != nil {
		t.Fatalf("load registered user: %v", err)
	}
	var profile model.UserProfile
	if err := runtimeValue.DB.Where("user_id = ?", registeredUser.ID).First(&profile).Error; err != nil {
		t.Fatalf("load registered profile: %v", err)
	}
	if profile.PublisherIdentityType != "owner" {
		t.Fatalf("expected default publisher identity owner, got %q", profile.PublisherIdentityType)
	}
	if profile.PrimaryCommunityID != nil || profile.ResidenceFloor != "" || profile.ResidenceUnit != "" {
		t.Fatalf("expected optional residence fields to remain empty, got %+v", profile)
	}

	loggedIn, err := authService.LoginWithUsername(context.Background(), EmailPasswordParams{
		Username: "residentone",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("login with username: %v", err)
	}
	if loggedIn.AccessToken == "" || loggedIn.User.PublicID != registered.User.PublicID {
		t.Fatalf("expected username login to return same user, got %+v", loggedIn.User)
	}
}

// 4. TestAuthServicePublisherIdentityUpdate verifies owner and agent self-service selection.
func TestAuthServicePublisherIdentityUpdate(t *testing.T) {
	runtimeValue := newAuthTestRuntime(t, nil)
	authService := NewAuthService(runtimeValue)

	registered, err := authService.RegisterWithEmail(context.Background(), EmailPasswordParams{
		Email:                 "agent@example.com",
		Password:              "password123",
		DisplayName:           "Agent One",
		Username:              "AgentOne",
		PhoneCountryCode:      "+852",
		PhoneNumber:           "62345678",
		PublisherIdentityType: "agent",
	})
	if err != nil {
		t.Fatalf("register agent account: %v", err)
	}

	var registeredUser model.User
	if err := runtimeValue.DB.Where("public_id = ?", registered.User.PublicID).First(&registeredUser).Error; err != nil {
		t.Fatalf("load registered agent: %v", err)
	}
	updated, err := NewUserService(runtimeValue).UpdateProfile(context.Background(), registeredUser.ID, UpdateProfileParams{
		DisplayName:           "Agent Updated",
		PublisherIdentityType: "owner",
	})
	if err != nil {
		t.Fatalf("update agent profile: %v", err)
	}
	if updated.PublisherIdentity != "owner" {
		t.Fatalf("expected updated identity owner, got %q", updated.PublisherIdentity)
	}

	var profile model.UserProfile
	if err := runtimeValue.DB.Where("user_id = ?", registeredUser.ID).First(&profile).Error; err != nil {
		t.Fatalf("load updated agent profile: %v", err)
	}
	if profile.PublisherIdentityType != "owner" {
		t.Fatalf("expected stored identity owner, got %q", profile.PublisherIdentityType)
	}
	if _, err := NewUserService(runtimeValue).UpdateProfile(context.Background(), registeredUser.ID, UpdateProfileParams{
		DisplayName:           "Agent Updated",
		PublisherIdentityType: "tenant",
	}); err == nil {
		t.Fatal("expected unsupported publisher identity update to fail")
	}
	if err := runtimeValue.DB.Model(&model.UserProfile{}).Where("user_id = ?", registeredUser.ID).
		Update("publisher_identity_type", "tenant").Error; err != nil {
		t.Fatalf("prepare legacy publisher identity: %v", err)
	}
	legacyProfile, err := NewUserService(runtimeValue).GetMe(context.Background(), registeredUser.ID)
	if err != nil {
		t.Fatalf("load legacy member profile: %v", err)
	}
	if legacyProfile.PublisherIdentity != "owner" {
		t.Fatalf("expected legacy publisher identity to default to owner, got %q", legacyProfile.PublisherIdentity)
	}
}

// 5. TestAuthServiceRegisterWithEngNameAllowsSpaces validates iSmart-aligned registration names.
func TestAuthServiceRegisterWithEngNameAllowsSpaces(t *testing.T) {
	runtimeValue := newAuthTestRuntime(t, nil)
	authService := NewAuthService(runtimeValue)

	registered, err := authService.RegisterWithEmail(context.Background(), EmailPasswordParams{
		Email:            "eng-name@example.com",
		Password:         "password123",
		EngName:          "CHAN TAI MAN",
		PhoneCountryCode: "+852",
		PhoneNumber:      "61239876",
	})
	if err != nil {
		t.Fatalf("register account: %v", err)
	}

	var registeredUser model.User
	if err := runtimeValue.DB.Where("public_id = ?", registered.User.PublicID).First(&registeredUser).Error; err != nil {
		t.Fatalf("load registered user: %v", err)
	}
	var profile model.UserProfile
	if err := runtimeValue.DB.Where("user_id = ?", registeredUser.ID).First(&profile).Error; err != nil {
		t.Fatalf("load registered profile: %v", err)
	}
	if profile.DisplayName != "CHAN TAI MAN" {
		t.Fatalf("expected eng_name as profile display name, got %q", profile.DisplayName)
	}

	var credential model.UserCredential
	if err := runtimeValue.DB.Where("user_id = ?", registeredUser.ID).First(&credential).Error; err != nil {
		t.Fatalf("load credential: %v", err)
	}
	if credential.Username == nil || *credential.Username != "61239876" {
		t.Fatalf("expected phone fallback username, got %+v", credential.Username)
	}
}

// 6. TestAuthServiceResetPasswordWithEmail validates email reset code flow.
func TestAuthServiceResetPasswordWithEmail(t *testing.T) {
	runtimeValue := newAuthTestRuntime(t, &config.Config{OTPMockCode: "123456"})
	runtimeValue.MailSender = &MockMailSender{}
	authService := NewAuthService(runtimeValue)

	if _, err := authService.RegisterWithEmail(context.Background(), EmailPasswordParams{
		Email:            "reset@example.com",
		Password:         "password123",
		DisplayName:      "Reset User",
		Username:         "ResetUser",
		PhoneCountryCode: "+852",
		PhoneNumber:      "61230000",
	}); err != nil {
		t.Fatalf("register account: %v", err)
	}

	requested, err := authService.RequestEmailPasswordReset(context.Background(), EmailOTPParams{
		Email: "RESET@example.com",
	})
	if err != nil {
		t.Fatalf("request password reset: %v", err)
	}
	if requested.MockCode != "123456" {
		t.Fatalf("expected mock code 123456, got %q", requested.MockCode)
	}

	resetResult, err := authService.ResetPasswordWithEmail(context.Background(), PasswordResetParams{
		Email:    "reset@example.com",
		Code:     "123456",
		Password: "newpass123",
	})
	if err != nil {
		t.Fatalf("reset password: %v", err)
	}
	if !resetResult.PasswordReset {
		t.Fatalf("expected password reset result")
	}

	if _, err := authService.LoginWithEmail(context.Background(), EmailPasswordParams{
		Email:    "reset@example.com",
		Password: "password123",
	}); err == nil {
		t.Fatalf("expected old password login to fail")
	}

	loggedIn, err := authService.LoginWithEmail(context.Background(), EmailPasswordParams{
		Email:    "reset@example.com",
		Password: "newpass123",
	})
	if err != nil {
		t.Fatalf("login with new password: %v", err)
	}
	if loggedIn.AccessToken == "" {
		t.Fatalf("expected token after password reset login")
	}
}

// 7. TestAuthServiceIsmartLoginCreatesLocalAccountWithoutRelay validates iSmart-first login.
func TestAuthServiceIsmartLoginCreatesLocalAccountWithoutRelay(t *testing.T) {
	server := newAuthServiceTestServer(t, http.StatusBadGateway)
	defer server.Close()
	runtimeValue := newAuthTestRuntime(
		t,
		&config.Config{
			POSLoginURL:           server.URL + "/poslogin",
			POSAPIBaseURL:         server.URL,
			POSLoginUsernameField: "login_name",
			POSLoginPasswordField: "password",
			POSLoginTimeout:       time.Second,
		},
		&model.User{},
		&model.UserCredential{},
		&model.UserProfile{},
		&model.UserIsmartAccount{},
		&model.Community{},
	)
	authService := NewAuthService(runtimeValue)

	result, err := authService.LoginWithIsmart(context.Background(), IsmartLoginParams{
		Account:  "patrick",
		Password: "s61980774",
	})
	if err != nil {
		t.Fatalf("login with ismart: %v", err)
	}
	if result.AccessToken == "" || result.User.PublicID == "" || !result.User.ProfileCompleted {
		t.Fatalf("expected signed-in completed member, got %+v", result.User)
	}
	if result.User.IsmartMsg == nil ||
		result.User.IsmartMsg.UserID != 88 ||
		result.User.IsmartMsg.Username != "patrick" ||
		result.User.IsmartMsg.Email != "patrick@example.com" ||
		result.User.IsmartMsg.Phone != "+85261234567" ||
		result.User.IsmartMsg.IsStaff ||
		len(result.User.IsmartMsg.Building) != 0 ||
		len(result.User.IsmartMsg.StaffBuildingPermissions) != 0 ||
		!reflect.DeepEqual(result.User.IsmartMsg.ClientBuildingPermissions, []string{"BLG-001"}) ||
		!reflect.DeepEqual(result.User.IsmartMsg.ClientBuildingFlatUnitsPermissions, []string{"BLG-0010000101"}) {
		t.Fatalf("expected POS login data in auth user response, got %#v", result.User.IsmartMsg)
	}

	var ismartAccount model.UserIsmartAccount
	if err := runtimeValue.DB.First(&ismartAccount, "ismart_user_id = ?", int64(88)).Error; err != nil {
		t.Fatalf("load ismart account: %v", err)
	}
	if ismartAccount.Username != "patrick" || ismartAccount.Email != "patrick@example.com" || ismartAccount.Phone != "+85261234567" {
		t.Fatalf("expected persisted ismart identity, got %+v", ismartAccount)
	}
	if ismartAccount.RelayTokenEncrypted != "" {
		t.Fatalf("expected failed relay login not to block or persist token")
	}
	ismartPassword, err := utils.DecryptString(runtimeValue.Config.EncryptionKey, ismartAccount.PasswordEncrypted)
	if err != nil || ismartPassword != "s61980774" {
		t.Fatalf("expected encrypted ismart password, got password=%q err=%v", ismartPassword, err)
	}

	var credential model.UserCredential
	if err := runtimeValue.DB.First(&credential, "user_id = ?", ismartAccount.UserID).Error; err != nil {
		t.Fatalf("load local credential: %v", err)
	}
	if credential.Username == nil || *credential.Username != "patrick" {
		t.Fatalf("expected local username credential, got %+v", credential)
	}
	if credential.Email == nil || *credential.Email != "patrick@example.com" {
		t.Fatalf("expected local email credential, got %+v", credential)
	}
	if !utils.VerifyPassword("s61980774", credential.PasswordHash) {
		t.Fatalf("expected local credential password to match iSmart password")
	}

	var profile model.UserProfile
	if err := runtimeValue.DB.Preload("PrimaryCommunity").First(&profile, "user_id = ?", ismartAccount.UserID).Error; err != nil {
		t.Fatalf("load synced profile: %v", err)
	}
	if profile.PrimaryCommunity == nil || profile.PrimaryCommunity.PublicID != "BLG-001" {
		t.Fatalf("expected primary community from iSmart building, got %+v", profile.PrimaryCommunity)
	}
	if profile.ResidenceFloor != "1" || profile.ResidenceUnit != "01" {
		t.Fatalf("expected POS unit split into floor and unit, got floor=%q unit=%q", profile.ResidenceFloor, profile.ResidenceUnit)
	}
}

// 8. TestAuthServiceIsmartLoginPreservesMemberCenterBinding validates login does not overwrite saved residence.
func TestAuthServiceIsmartLoginPreservesMemberCenterBinding(t *testing.T) {
	server := newAuthServiceTestServer(t, http.StatusBadGateway)
	defer server.Close()
	runtimeValue := newAuthTestRuntime(
		t,
		&config.Config{
			POSLoginURL:           server.URL + "/poslogin",
			POSAPIBaseURL:         server.URL,
			POSLoginUsernameField: "login_name",
			POSLoginPasswordField: "password",
			POSLoginTimeout:       time.Second,
		},
		&model.User{},
		&model.UserCredential{},
		&model.UserProfile{},
		&model.UserIsmartAccount{},
		&model.Community{},
	)
	authService := NewAuthService(runtimeValue)

	if _, err := authService.LoginWithIsmart(context.Background(), IsmartLoginParams{
		Account:  "patrick",
		Password: "s61980774",
	}); err != nil {
		t.Fatalf("initial login with ismart: %v", err)
	}

	var ismartAccount model.UserIsmartAccount
	if err := runtimeValue.DB.First(&ismartAccount, "ismart_user_id = ?", int64(88)).Error; err != nil {
		t.Fatalf("load ismart account: %v", err)
	}
	manualCommunity := model.Community{
		PublicID:      "0999900",
		CommunityType: "building",
		NameZH:        "測試1大廈",
		DistrictCode:  "HK",
	}
	if err := runtimeValue.DB.Create(&manualCommunity).Error; err != nil {
		t.Fatalf("create manual community: %v", err)
	}
	buildingJSON, err := marshalJSON([]string{"0999900"})
	if err != nil {
		t.Fatalf("marshal building binding: %v", err)
	}
	unitJSON, err := marshalJSON([]string{"099990000001A"})
	if err != nil {
		t.Fatalf("marshal unit binding: %v", err)
	}
	if err := runtimeValue.DB.Model(&model.UserProfile{}).Where("user_id = ?", ismartAccount.UserID).Updates(map[string]any{
		"primary_community_id": manualCommunity.ID,
		"bound_building_ids":   buildingJSON,
		"bound_flat_unit_ids":  unitJSON,
		"residence_floor":      "01",
		"residence_unit":       "A",
		"district_code":        manualCommunity.DistrictCode,
	}).Error; err != nil {
		t.Fatalf("save manual profile binding: %v", err)
	}

	secondResult, err := authService.LoginWithIsmart(context.Background(), IsmartLoginParams{
		Account:  "patrick",
		Password: "s61980774",
	})
	if err != nil {
		t.Fatalf("second login with ismart: %v", err)
	}
	if !secondResult.User.ProfileCompleted {
		t.Fatalf("expected profile to remain completed")
	}

	var profile model.UserProfile
	if err := runtimeValue.DB.Preload("PrimaryCommunity").First(&profile, "user_id = ?", ismartAccount.UserID).Error; err != nil {
		t.Fatalf("load preserved profile: %v", err)
	}
	if profile.PrimaryCommunity == nil || profile.PrimaryCommunity.PublicID != "0999900" {
		t.Fatalf("expected manual community to survive login, got %+v", profile.PrimaryCommunity)
	}
	if profile.ResidenceFloor != "01" || profile.ResidenceUnit != "A" {
		t.Fatalf("expected manual unit to survive login, got floor=%q unit=%q", profile.ResidenceFloor, profile.ResidenceUnit)
	}
	if buildings := normalizeStringSlice(unmarshalStringSlice(profile.BoundBuildingIDs)); len(buildings) != 1 || buildings[0] != "0999900" {
		t.Fatalf("expected manual building bindings to survive login, got %v", buildings)
	}
	if units := normalizeStringSlice(unmarshalStringSlice(profile.BoundFlatUnitIDs)); len(units) != 1 || units[0] != "099990000001A" {
		t.Fatalf("expected manual unit bindings to survive login, got %v", units)
	}

	me, err := NewUserService(runtimeValue).GetMe(context.Background(), ismartAccount.UserID)
	if err != nil {
		t.Fatalf("load me response: %v", err)
	}
	if me.PrimaryCommunity == nil || me.PrimaryCommunity.PublicID != "0999900" {
		t.Fatalf("expected me response to use manual community, got %+v", me.PrimaryCommunity)
	}
	if len(me.BoundBuildingIDs) != 1 || me.BoundBuildingIDs[0] != "0999900" {
		t.Fatalf("expected me response to use manual building binding, got %v", me.BoundBuildingIDs)
	}
	if len(me.BoundFlatUnitIDs) != 1 || me.BoundFlatUnitIDs[0] != "099990000001A" {
		t.Fatalf("expected me response to use manual unit binding, got %v", me.BoundFlatUnitIDs)
	}
}

// 9. TestAuthServiceDoesNotPromoteIsmartStaffToAJOStaff verifies external staff status stays business metadata.
func TestAuthServiceDoesNotPromoteIsmartStaffToAJOStaff(t *testing.T) {
	runtimeValue := newAuthTestRuntime(
		t,
		nil,
		&model.User{},
		&model.UserCredential{},
		&model.UserProfile{},
		&model.UserIsmartAccount{},
		&model.Community{},
	)
	authService := NewAuthService(runtimeValue)
	result, err := authService.upsertUserByIsmart(
		context.Background(),
		&IsmartMessage{
			UserID:                             188,
			Username:                           "ismart-staff-member",
			Email:                              "ismart-staff@example.com",
			Phone:                              "+85261234570",
			IsStaff:                            true,
			StaffBuildingPermissions:           []string{"0419900"},
			ClientBuildingPermissions:          []string{"0999900"},
			ClientBuildingFlatUnitsPermissions: []string{"09999000012"},
		},
		"+85261234570",
		"ismart-staff@example.com",
		"",
		"password123",
	)
	if err != nil {
		t.Fatalf("upsert iSmart account: %v", err)
	}
	if result.User.IsStaff || result.User.Role != MemberTypeUser {
		t.Fatalf("expected regular AJO member access, got %#v", result.User)
	}

	var user model.User
	if err := runtimeValue.DB.First(&user, "public_id = ?", result.User.PublicID).Error; err != nil {
		t.Fatalf("load AJO user: %v", err)
	}
	if user.IsStaff {
		t.Fatal("expected iSmart staff status not to grant AJO staff access")
	}

	var account model.UserIsmartAccount
	if err := runtimeValue.DB.First(&account, "user_id = ?", user.ID).Error; err != nil {
		t.Fatalf("load iSmart account: %v", err)
	}
	if !account.IsStaff {
		t.Fatal("expected upstream iSmart staff status to remain available as business metadata")
	}
	if buildings := resolveIsmartBoundBuildings(authService.loadIsmartMessage(context.Background(), user.ID)); len(buildings) != 1 || buildings[0] != "0999900" {
		t.Fatalf("expected resident buildings only, got %v", buildings)
	}
}

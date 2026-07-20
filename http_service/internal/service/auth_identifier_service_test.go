/*
 * Unified password login service tests.
 * 1. Validate local identifier classification and authentication.
 * 2. Validate controlled iSmart fallback and generic credential errors.
 * 3. Validate database and upstream failures stop further fallback attempts.
 */
package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"ajoliving_web/http_service/internal/config"
	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/model"
	"ajoliving_web/http_service/internal/utils"
)

// 1. TestAuthServiceLoginWithIdentifierUsesLocalAccounts validates all local identifier forms.
func TestAuthServiceLoginWithIdentifierUsesLocalAccounts(t *testing.T) {
	var loginCalls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		loginCalls.Add(1)
		response.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	runtimeValue := newIdentifierLoginTestRuntime(t, server.URL)
	createIdentifierLoginUser(t, runtimeValue, "resident", "resident@example.com", "+852", "61234567", "password123")
	authService := NewAuthService(runtimeValue)

	tests := []string{"resident@example.com", "Resident", "+852 (6123) 4567"}
	for _, identifier := range tests {
		result, err := authService.LoginWithIdentifier(context.Background(), IdentifierLoginParams{
			Identifier: identifier,
			Password:   "password123",
		})
		if err != nil || result == nil || result.AccessToken == "" {
			t.Fatalf("login with %q failed: %#v %v", identifier, result, err)
		}
	}
	if loginCalls.Load() != 0 {
		t.Fatalf("expected local login to avoid POS, got %d calls", loginCalls.Load())
	}
}

// 2. TestAuthServiceLoginWithIdentifierFallsBackToIsmart validates one-request legacy login.
func TestAuthServiceLoginWithIdentifierFallsBackToIsmart(t *testing.T) {
	var loginCalls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("Content-Type", "application/json")
		if request.URL.Path == "/poslogin" {
			loginCalls.Add(1)
			var payload map[string]string
			if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
				t.Errorf("decode pos login payload: %v", err)
				response.WriteHeader(http.StatusBadRequest)
				return
			}
			if payload["login_name"] != "patrick" || payload["password"] != "legacy-password" {
				response.WriteHeader(http.StatusUnauthorized)
				return
			}
			_, _ = response.Write([]byte(`{"code":1,"message":"success","msg":{"user_id":388,"username":"patrick","email":"patrick@example.com","phone":"+85261234567"}}`))
			return
		}
		response.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	runtimeValue := newIdentifierLoginTestRuntime(t, server.URL)
	result, err := NewAuthService(runtimeValue).LoginWithIdentifier(context.Background(), IdentifierLoginParams{
		Identifier: "patrick",
		Password:   "legacy-password",
	})
	if err != nil || result == nil || result.User.IsmartMsg == nil || result.User.IsmartMsg.UserID != 388 {
		t.Fatalf("iSmart fallback failed: %#v %v", result, err)
	}
	if loginCalls.Load() != 1 {
		t.Fatalf("expected one POS credential attempt, got %d", loginCalls.Load())
	}
}

// 3. TestAuthServiceLoginWithIdentifierReturnsGenericCredentialError hides account existence.
func TestAuthServiceLoginWithIdentifierReturnsGenericCredentialError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	_, err := NewAuthService(newIdentifierLoginTestRuntime(t, server.URL)).LoginWithIdentifier(context.Background(), IdentifierLoginParams{
		Identifier: "unknown-user",
		Password:   "wrong-password",
	})
	assertIdentifierLoginError(t, err, errcode.CodeValidationError, "account or password is incorrect")
}

// 4. TestAuthServiceLoginWithIdentifierPreservesPOSFailure validates upstream failures are not credential errors.
func TestAuthServiceLoginWithIdentifierPreservesPOSFailure(t *testing.T) {
	var loginCalls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		loginCalls.Add(1)
		response.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	_, err := NewAuthService(newIdentifierLoginTestRuntime(t, server.URL)).LoginWithIdentifier(context.Background(), IdentifierLoginParams{
		Identifier: "+852 (6123) 4567",
		Password:   "wrong-password",
	})
	assertIdentifierLoginError(t, err, errcode.CodeInternalError, "failed to call pos login")
	if loginCalls.Load() != 1 {
		t.Fatalf("expected upstream failure to stop fallback, got %d calls", loginCalls.Load())
	}
}

// 5. TestAuthServiceLoginWithIdentifierStopsOnDatabaseFailure validates local failures do not reach iSmart.
func TestAuthServiceLoginWithIdentifierStopsOnDatabaseFailure(t *testing.T) {
	var loginCalls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		loginCalls.Add(1)
		response.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	runtimeValue := newIdentifierLoginTestRuntime(t, server.URL)
	sqlDB, err := runtimeValue.DB.DB()
	if err != nil {
		t.Fatalf("load sql db: %v", err)
	}
	if err := sqlDB.Close(); err != nil {
		t.Fatalf("close sql db: %v", err)
	}

	_, err = NewAuthService(runtimeValue).LoginWithIdentifier(context.Background(), IdentifierLoginParams{
		Identifier: "resident",
		Password:   "password123",
	})
	assertIdentifierLoginError(t, err, errcode.CodeInternalError, "failed to load username account")
	if loginCalls.Load() != 0 {
		t.Fatalf("expected database failure to stop iSmart fallback, got %d calls", loginCalls.Load())
	}
}

// 6. TestParseLoginPhoneIdentifier validates supported phone normalization.
func TestParseLoginPhoneIdentifier(t *testing.T) {
	tests := []struct {
		input       string
		countryCode string
		phoneNumber string
		valid       bool
	}{
		{input: "6123 4567", countryCode: "+852", phoneNumber: "61234567", valid: true},
		{input: "+852 (6123) 4567", countryCode: "+852", phoneNumber: "61234567", valid: true},
		{input: "+86 (138) 0013-8000", countryCode: "+86", phoneNumber: "13800138000", valid: true},
		{input: "+1 (380) 013-8000", valid: false},
		{input: "++852 6123 4567", valid: false},
		{input: "patrick", valid: false},
	}
	for _, item := range tests {
		countryCode, phoneNumber, valid := parseLoginPhoneIdentifier(item.input)
		if countryCode != item.countryCode || phoneNumber != item.phoneNumber || valid != item.valid {
			t.Fatalf("unexpected phone parse for %q: %q %q %t", item.input, countryCode, phoneNumber, valid)
		}
	}
}

// 7. newIdentifierLoginTestRuntime creates the full local and iSmart persistence runtime.
func newIdentifierLoginTestRuntime(t *testing.T, serverURL string) *Runtime {
	return newAuthTestRuntime(t, &config.Config{
		POSLoginURL:     serverURL + "/poslogin",
		POSAPIBaseURL:   serverURL,
		POSLoginTimeout: time.Second,
	}, &model.User{}, &model.UserCredential{}, &model.UserProfile{}, &model.UserIsmartAccount{}, &model.Community{})
}

// 8. createIdentifierLoginUser saves one local password account for classification tests.
func createIdentifierLoginUser(t *testing.T, runtimeValue *Runtime, username string, email string, countryCode string, phone string, password string) {
	t.Helper()
	passwordHash, err := utils.HashPassword(password)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	user := model.User{
		PublicID:         utils.NewPublicID(),
		PhoneCountryCode: countryCode,
		PhoneNumber:      phone,
		MemberStatus:     "active",
		MemberType:       MemberTypeUser,
		IsVerifiedPhone:  true,
	}
	if err := runtimeValue.DB.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	credential := model.UserCredential{UserID: user.ID, Username: &username, Email: &email, PasswordHash: passwordHash, IsVerified: true}
	if err := runtimeValue.DB.Create(&credential).Error; err != nil {
		t.Fatalf("create credential: %v", err)
	}
	if err := runtimeValue.DB.Create(&model.UserProfile{UserID: user.ID, DisplayName: username, AccountType: AccountTypePersonal}).Error; err != nil {
		t.Fatalf("create profile: %v", err)
	}
}

// 9. assertIdentifierLoginError validates public error code and message.
func assertIdentifierLoginError(t *testing.T, err error, code string, message string) {
	t.Helper()
	var appErr *errcode.AppError
	if !errors.As(err, &appErr) || appErr.Code != code || appErr.Message != message {
		t.Fatalf("unexpected login error: %#v", err)
	}
}

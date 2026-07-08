/*
 * Authentication service tests.
 * 1. Validate resident registration credential persistence.
 * 2. Validate username password sign-in uses the same token flow.
 * 3. Validate optional registration profile fields stay optional.
 * 4. Validate email password reset updates stored credentials.
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

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"ajoliving_web/http_service/internal/config"
	"ajoliving_web/http_service/internal/model"
	"ajoliving_web/http_service/internal/utils"
)

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
		Email:                 "resident@example.com",
		Password:              "password123",
		DisplayName:           "Resident One",
		Username:              "ResidentOne",
		PhoneCountryCode:      "+852",
		PhoneNumber:           "61234567",
		PublisherIdentityType: "tenant",
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
	if profile.PublisherIdentityType != "tenant" {
		t.Fatalf("expected publisher identity tenant, got %q", profile.PublisherIdentityType)
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

// 4. TestAuthServiceResetPasswordWithEmail validates email reset code flow.
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

// 5. TestAuthServiceIsmartLoginCreatesLocalAccountWithoutRelay validates iSmart-first login.
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

// 6. TestAuthServiceIsmartLoginPreservesMemberCenterBinding validates login does not overwrite saved residence.
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

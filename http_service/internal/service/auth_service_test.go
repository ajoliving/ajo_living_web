/*
 * Auth and profile service tests.
 * 1. Verify OTP should create a user and return tokens.
 * 2. Profile update should store the selected community.
 */
package service

import (
	"context"
	"errors"
	"testing"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/model"
	"ajoliving_web/http_service/internal/utils"
)

// 1. TestVerifyOTPAndUpdateProfile covers OTP verify and profile update flow.
func TestVerifyOTPAndUpdateProfile(t *testing.T) {
	runtime := newTestRuntime(t)
	authService := NewAuthService(runtime)
	userService := NewUserService(runtime)
	communityA, _ := mustGetCommunities(t, runtime)

	if _, err := authService.RequestOTP(context.Background(), RequestOTPParams{
		PhoneCountryCode: "+852",
		PhoneNumber:      "91230001",
		Scene:            "login",
	}); err != nil {
		t.Fatalf("request otp: %v", err)
	}

	result, err := authService.VerifyOTP(context.Background(), VerifyOTPParams{
		PhoneCountryCode: "+852",
		PhoneNumber:      "91230001",
		Scene:            "login",
		Code:             "123456",
	})
	if err != nil {
		t.Fatalf("verify otp: %v", err)
	}

	if result.AccessToken == "" || result.RefreshToken == "" {
		t.Fatalf("expected non-empty tokens")
	}
	if result.User.MemberType != MemberTypeUser || result.User.IsStaff || result.User.Role != MemberTypeUser {
		t.Fatalf("expected default role to be user, got %+v", result.User)
	}
	if result.User.ProfileCompleted {
		t.Fatalf("expected profile to be incomplete before community binding")
	}

	identity, err := authService.AuthenticateToken(context.Background(), result.AccessToken)
	if err != nil {
		t.Fatalf("authenticate token: %v", err)
	}
	if identity.MemberType != MemberTypeUser || identity.IsStaff || identity.Role != MemberTypeUser {
		t.Fatalf("expected default authenticated identity to be user, got %+v", identity)
	}

	updated, err := userService.UpdateProfile(context.Background(), identity.UserID, UpdateProfileParams{
		DisplayName:           "Neighbour User",
		PublisherIdentityType: "owner",
		PrimaryCommunityID:    communityA.PublicID,
	})
	if err != nil {
		t.Fatalf("update profile: %v", err)
	}

	if !updated.ProfileCompleted {
		t.Fatalf("expected profile to be completed after community update")
	}
	if updated.PrimaryCommunity == nil || updated.PrimaryCommunity.PublicID != communityA.PublicID {
		t.Fatalf("expected primary community to match the selected community")
	}
	if updated.MemberType != MemberTypeUser || updated.IsStaff || updated.Role != MemberTypeUser {
		t.Fatalf("expected updated profile to keep default role fields, got %+v", updated)
	}
}

// 2. TestUpdateProfilePhoneAndAvatarCharge covers profile phone update and avatar point charge.
func TestUpdateProfilePhoneAndAvatarCharge(t *testing.T) {
	runtime := newTestRuntime(t)
	userService := NewUserService(runtime)
	communityA, _ := mustGetCommunities(t, runtime)
	user := mustCreateUser(t, runtime, "+852", "91230011", &communityA.ID)
	mustGrantPoints(t, runtime, user.ID, 60)
	avatarAssetID := mustCreateAccountAvatarAsset(t, runtime, user.ID)

	updated, err := userService.UpdateProfile(context.Background(), user.ID, UpdateProfileParams{
		DisplayName:      "Avatar Member",
		PhoneCountryCode: "853",
		PhoneNumber:      "66889900",
		AvatarAssetID:    avatarAssetID,
	})
	if err != nil {
		t.Fatalf("update profile avatar: %v", err)
	}
	if updated.PhoneCountryCode != "+853" || updated.PhoneNumber != "66889900" {
		t.Fatalf("expected updated phone, got %+v", updated)
	}
	if updated.AJOBalance != 10 {
		t.Fatalf("expected 10 points remaining, got %d", updated.AJOBalance)
	}

	var transaction model.WalletTransaction
	if err := runtime.DB.Where("user_id = ? AND biz_module = ? AND action_type = ?", user.ID, "profile", WalletActionAvatar).First(&transaction).Error; err != nil {
		t.Fatalf("load avatar charge transaction: %v", err)
	}
	if transaction.Amount != 50 || transaction.SourceType != WalletSourceProfileCharge {
		t.Fatalf("unexpected avatar charge transaction: %#v", transaction)
	}

	var storedUser model.User
	if err := runtime.DB.First(&storedUser, user.ID).Error; err != nil {
		t.Fatalf("load stored user: %v", err)
	}
	if storedUser.IsVerifiedPhone {
		t.Fatalf("expected changed phone to be marked unverified")
	}
}

// 3. TestUpdateProfileAvatarRequiresPoints keeps avatar changes blocked without enough balance.
func TestUpdateProfileAvatarRequiresPoints(t *testing.T) {
	runtime := newTestRuntime(t)
	userService := NewUserService(runtime)
	communityA, _ := mustGetCommunities(t, runtime)
	user := mustCreateUser(t, runtime, "+852", "91230012", &communityA.ID)
	avatarAssetID := mustCreateAccountAvatarAsset(t, runtime, user.ID)

	_, err := userService.UpdateProfile(context.Background(), user.ID, UpdateProfileParams{
		DisplayName:   "No Point Member",
		AvatarAssetID: avatarAssetID,
	})
	var appErr *errcode.AppError
	if !errors.As(err, &appErr) || appErr.Code != errcode.CodePointsInsufficient {
		t.Fatalf("expected insufficient points, got %#v", err)
	}

	var profile model.UserProfile
	if err := runtime.DB.Where("user_id = ?", user.ID).First(&profile).Error; err != nil {
		t.Fatalf("load profile: %v", err)
	}
	if profile.AvatarAssetID != nil {
		t.Fatalf("avatar should not be changed when charge fails")
	}
}

// 4. TestUpdateProfileRejectsDuplicatePhone keeps phone numbers unique.
func TestUpdateProfileRejectsDuplicatePhone(t *testing.T) {
	runtime := newTestRuntime(t)
	userService := NewUserService(runtime)
	communityA, _ := mustGetCommunities(t, runtime)
	user := mustCreateUser(t, runtime, "+852", "91230013", &communityA.ID)
	_ = mustCreateUser(t, runtime, "+852", "91230014", &communityA.ID)

	_, err := userService.UpdateProfile(context.Background(), user.ID, UpdateProfileParams{
		PhoneCountryCode: "+852",
		PhoneNumber:      "91230014",
	})
	var appErr *errcode.AppError
	if !errors.As(err, &appErr) || appErr.Code != errcode.CodeValidationError {
		t.Fatalf("expected duplicate phone validation error, got %#v", err)
	}
}

// 5. mustCreateAccountAvatarAsset creates an account directory image asset.
func mustCreateAccountAvatarAsset(t *testing.T, runtime *Runtime, userID int64) string {
	t.Helper()

	publicID := utils.NewPublicID()
	asset := model.MediaAsset{
		PublicID:        publicID,
		StorageProvider: runtime.Config.StorageProvider,
		BucketName:      runtime.Config.StorageBucket,
		ObjectKey:       accountMediaObjectPrefix + publicID + ".webp",
		MimeType:        "image/webp",
		FileSize:        1024,
		CreatedBy:       &userID,
		CreatedAt:       runtime.Now(),
	}
	if err := runtime.DB.Create(&asset).Error; err != nil {
		t.Fatalf("create avatar asset: %v", err)
	}

	return asset.PublicID
}

// 6. TestEmailPasswordRegisterAndLogin covers email account creation and sign-in.
func TestEmailPasswordRegisterAndLogin(t *testing.T) {
	runtime := newTestRuntime(t)
	authService := NewAuthService(runtime)

	registered, err := authService.RegisterWithEmail(context.Background(), EmailPasswordParams{
		Email:            "member@example.com",
		Password:         "safe-password-123",
		DisplayName:      "Email Member",
		PhoneCountryCode: "+852",
		PhoneNumber:      "91234567",
	})
	if err != nil {
		t.Fatalf("register email: %v", err)
	}
	if registered.AccessToken == "" || registered.RefreshToken == "" {
		t.Fatalf("expected register tokens")
	}

	loggedIn, err := authService.LoginWithEmail(context.Background(), EmailPasswordParams{
		Email:    "member@example.com",
		Password: "safe-password-123",
	})
	if err != nil {
		t.Fatalf("login email: %v", err)
	}
	if loggedIn.AccessToken == "" || loggedIn.User.MemberType != MemberTypeUser {
		t.Fatalf("expected user login result, got %+v", loggedIn)
	}

	phoneLoggedIn, err := authService.LoginWithPhone(context.Background(), PhonePasswordParams{
		PhoneCountryCode: "+852",
		PhoneNumber:      "91234567",
		Password:         "safe-password-123",
	})
	if err != nil {
		t.Fatalf("login phone: %v", err)
	}
	if phoneLoggedIn.AccessToken == "" || phoneLoggedIn.User.MemberType != MemberTypeUser {
		t.Fatalf("expected phone login result, got %+v", phoneLoggedIn)
	}

	if _, err := authService.LoginWithEmail(context.Background(), EmailPasswordParams{
		Email:    "member@example.com",
		Password: "wrong-password",
	}); err == nil {
		t.Fatalf("expected wrong password to fail")
	}

	if _, err := authService.LoginWithPhone(context.Background(), PhonePasswordParams{
		PhoneCountryCode: "+852",
		PhoneNumber:      "91234567",
		Password:         "wrong-password",
	}); err == nil {
		t.Fatalf("expected wrong phone password to fail")
	}
}

// 3. TestEmailOTPLoginCreatesAndReusesUser covers email code sign-in.
func TestEmailOTPLoginCreatesAndReusesUser(t *testing.T) {
	runtime := newTestRuntime(t)
	authService := NewAuthService(runtime)

	requested, err := authService.RequestEmailOTP(context.Background(), EmailOTPParams{
		Email: "otp-member@example.com",
		Scene: "login",
	})
	if err != nil {
		t.Fatalf("request email otp: %v", err)
	}
	if requested.MockCode != "123456" {
		t.Fatalf("expected mock code, got %+v", requested)
	}

	verified, err := authService.VerifyEmailOTP(context.Background(), EmailOTPParams{
		Email:       "otp-member@example.com",
		Scene:       "login",
		Code:        "123456",
		DisplayName: "OTP Member",
	})
	if err != nil {
		t.Fatalf("verify email otp: %v", err)
	}
	if verified.AccessToken == "" || verified.User.MemberType != MemberTypeUser {
		t.Fatalf("expected email otp login result, got %+v", verified)
	}

	if _, err := authService.RequestEmailOTP(context.Background(), EmailOTPParams{Email: "otp-member@example.com"}); err != nil {
		t.Fatalf("request second email otp: %v", err)
	}
	secondLogin, err := authService.VerifyEmailOTP(context.Background(), EmailOTPParams{
		Email: "otp-member@example.com",
		Code:  "123456",
	})
	if err != nil {
		t.Fatalf("verify second email otp: %v", err)
	}
	if secondLogin.AccessToken == "" {
		t.Fatalf("expected second login token")
	}
}

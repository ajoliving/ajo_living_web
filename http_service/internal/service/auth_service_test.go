/*
 * Auth and profile service tests.
 * 1. Verify OTP should create a user and return tokens.
 * 2. Profile update should store the selected community.
 */
package service

import (
	"context"
	"testing"
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

// 2. TestEmailPasswordRegisterAndLogin covers email account creation and sign-in.
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

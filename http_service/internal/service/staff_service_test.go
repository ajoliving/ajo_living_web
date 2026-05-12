/*
 * Staff service tests.
 * 1. Bootstrap a staff account through OTP verify.
 * 2. Promote a member to pro_user and validate staff user listing.
 */
package service

import (
	"context"
	"testing"

	"ajoliving_web/http_service/internal/model"
)

// 1. TestBootstrapStaffCreateUserAndUpdateUserRole validates staff account management.
func TestBootstrapStaffAndUpdateUserRole(t *testing.T) {
	runtime := newTestRuntime(t)
	runtime.Config.BootstrapStaffPhones = "+85291238888"
	authService := NewAuthService(runtime)
	staffService := NewStaffService(runtime)
	communityA, _ := mustGetCommunities(t, runtime)

	if _, err := authService.RequestOTP(context.Background(), RequestOTPParams{
		PhoneCountryCode: "+852",
		PhoneNumber:      "91238888",
		Scene:            "login",
	}); err != nil {
		t.Fatalf("request staff otp: %v", err)
	}

	result, err := authService.VerifyOTP(context.Background(), VerifyOTPParams{
		PhoneCountryCode: "+852",
		PhoneNumber:      "91238888",
		Scene:            "login",
		Code:             "123456",
	})
	if err != nil {
		t.Fatalf("verify staff otp: %v", err)
	}

	if !result.User.IsStaff || result.User.Role != RoleStaff {
		t.Fatalf("expected bootstrap account to become staff, got %+v", result.User)
	}

	identity, err := authService.AuthenticateToken(context.Background(), result.AccessToken)
	if err != nil {
		t.Fatalf("authenticate staff token: %v", err)
	}

	me, err := staffService.GetStaffMe(context.Background(), identity.UserID)
	if err != nil {
		t.Fatalf("get staff me: %v", err)
	}
	if !me.IsStaff || me.Role != RoleStaff {
		t.Fatalf("expected staff me payload, got %+v", me)
	}
	if len(me.Roles) != 2 || me.Roles[0] != model.RoleCodeSuperAdmin || me.Roles[1] != model.RoleCodeMember {
		t.Fatalf("expected compact staff role set, got %+v", me.Roles)
	}

	target := mustCreateUser(t, runtime, "+852", "91239999", &communityA.ID)
	proUser := MemberTypeProUser
	notStaff := false
	updated, err := staffService.UpdateUserRole(context.Background(), identity.UserID, target.PublicID, StaffUserRoleUpdateParams{
		MemberType: &proUser,
		IsStaff:    &notStaff,
	})
	if err != nil {
		t.Fatalf("update user role: %v", err)
	}

	if updated.MemberType != MemberTypeProUser || updated.IsStaff || updated.Role != MemberTypeProUser {
		t.Fatalf("expected target user to become pro_user, got %+v", updated)
	}
	if len(updated.Roles) != 1 || updated.Roles[0] != model.RoleCodeMember {
		t.Fatalf("expected compact member role set, got %+v", updated.Roles)
	}

	items, pagination, err := staffService.ListUsers(context.Background(), StaffUserListFilters{
		Page:       1,
		PageSize:   20,
		MemberType: MemberTypeProUser,
	})
	if err != nil {
		t.Fatalf("list users: %v", err)
	}

	if pagination.Total < 1 {
		t.Fatalf("expected at least one pro_user in pagination, got %+v", pagination)
	}

	found := false
	for _, item := range items {
		if item.PublicID == target.PublicID && item.Role == MemberTypeProUser {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected promoted user to appear in staff list")
	}

	created, err := staffService.CreateUser(context.Background(), identity.UserID, StaffUserCreateParams{
		Email:            "staff-created@example.com",
		Password:         "staffpass123",
		DisplayName:      "Created Staff",
		PhoneCountryCode: "+852",
		PhoneNumber:      "91237777",
		MemberType:       MemberTypeUser,
		RoleCodes:        []string{model.RoleCodeStaff},
	})
	if err != nil {
		t.Fatalf("create staff user: %v", err)
	}
	if !created.IsStaff || created.Role != RoleStaff || created.DisplayName != "Created Staff" {
		t.Fatalf("expected created staff account, got %+v", created)
	}
	if created.Email != "staff-created@example.com" {
		t.Fatalf("expected created staff email, got %q", created.Email)
	}
	if len(created.Roles) != 2 || created.Roles[0] != model.RoleCodeStaff || created.Roles[1] != model.RoleCodeMember {
		t.Fatalf("expected staff and member roles, got %+v", created.Roles)
	}

	loggedIn, err := authService.LoginWithEmail(context.Background(), EmailPasswordParams{
		Email:    "staff-created@example.com",
		Password: "staffpass123",
	})
	if err != nil {
		t.Fatalf("login created staff user: %v", err)
	}
	if !loggedIn.User.IsStaff || loggedIn.User.Role != RoleStaff {
		t.Fatalf("expected created account to login as staff, got %+v", loggedIn.User)
	}
}

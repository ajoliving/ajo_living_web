/*
 * 大廈副戶授權服務測試。
 * 1. 驗證未註冊電話及電郵會建立待啟用帳戶並可接受邀請。
 * 2. 驗證已註冊帳戶會直接取得功能授權。
 */
package service

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"
	"testing"

	"ajoliving_web/http_service/internal/config"
	"ajoliving_web/http_service/internal/model"
	"ajoliving_web/http_service/internal/utils"
)

type buildingAuthorizationTestMail struct{ body string }

// 1. Send captures the invitation email body.
func (m *buildingAuthorizationTestMail) Send(_ context.Context, _ string, _ string, body string) error {
	m.body = body
	return nil
}

// 2. TestBuildingAuthorizationCreatesAndAcceptsInvitation verifies the unregistered flow.
func TestBuildingAuthorizationCreatesAndAcceptsInvitation(t *testing.T) {
	runtime := newAuthTestRuntime(t, &config.Config{AppPublicBaseURL: "https://ajo.test"}, &model.User{}, &model.UserCredential{}, &model.UserProfile{}, &model.UserIsmartAccount{}, &model.BuildingAuthorization{})
	owner := createBuildingAuthorizationTestOwner(t, runtime, "61230001", "BLG-001")
	mail := &buildingAuthorizationTestMail{}
	runtime.MailSender = mail
	service := NewBuildingAuthorizationService(runtime)
	result, err := service.Create(context.Background(), owner.ID, BuildingAuthorizationCreateParams{BuildingID: "BLG-001", PhoneCountryCode: "+852", PhoneNumber: "61230002", Email: "invite@example.com", Permissions: []string{BuildingPermissionRemoteDoorOpen, BuildingPermissionBuildingNotices}})
	if err != nil || result["account_created"] != true || result["status"] != BuildingAuthorizationStatusPending {
		t.Fatalf("unexpected invitation result: %#v %v", result, err)
	}
	parsed, err := url.Parse(strings.TrimSpace(strings.Split(mail.body, "\n\n")[1]))
	if err != nil || parsed.Query().Get("token") == "" {
		t.Fatalf("invitation token missing from email: %q %v", mail.body, err)
	}
	if err := service.Accept(context.Background(), BuildingAuthorizationAcceptParams{Token: parsed.Query().Get("token"), Username: "invited-user", Password: "password123"}); err != nil {
		t.Fatalf("accept invitation: %v", err)
	}
	var grant model.BuildingAuthorization
	if err := runtime.DB.First(&grant).Error; err != nil || grant.Status != BuildingAuthorizationStatusActive {
		t.Fatalf("grant not activated: %#v %v", grant, err)
	}
	if err := service.RequirePermission(context.Background(), *grant.GranteeUserID, "BLG-001", BuildingPermissionBuildingNotices); err != nil {
		t.Fatalf("notice permission missing: %v", err)
	}
}

// 3. TestBuildingAuthorizationDirectGrant verifies the registered flow.
func TestBuildingAuthorizationDirectGrant(t *testing.T) {
	runtime := newAuthTestRuntime(t, &config.Config{}, &model.User{}, &model.UserCredential{}, &model.UserProfile{}, &model.UserIsmartAccount{}, &model.BuildingAuthorization{})
	owner := createBuildingAuthorizationTestOwner(t, runtime, "61230003", "BLG-002")
	grantee := model.User{PublicID: utils.NewPublicID(), PhoneCountryCode: "+852", PhoneNumber: "61230004", MemberStatus: "active", MemberType: MemberTypeUser, IsVerifiedPhone: true}
	if err := runtime.DB.Create(&grantee).Error; err != nil {
		t.Fatal(err)
	}
	email := "member@example.com"
	if err := runtime.DB.Create(&model.UserCredential{UserID: grantee.ID, Email: &email, IsVerified: true}).Error; err != nil {
		t.Fatal(err)
	}
	result, err := NewBuildingAuthorizationService(runtime).Create(context.Background(), owner.ID, BuildingAuthorizationCreateParams{BuildingID: "BLG-002", PhoneCountryCode: "+852", PhoneNumber: "61230004", Email: email, Permissions: []string{BuildingPermissionBuildingNotices}})
	if err != nil || result["status"] != BuildingAuthorizationStatusActive {
		t.Fatalf("unexpected direct grant: %#v %v", result, err)
	}
}

// 4. createBuildingAuthorizationTestOwner creates an owner with visible iSmart building data.
func createBuildingAuthorizationTestOwner(t *testing.T, runtime *Runtime, phone string, buildingID string) *model.User {
	t.Helper()
	owner := model.User{PublicID: utils.NewPublicID(), PhoneCountryCode: "+852", PhoneNumber: phone, MemberStatus: "active", MemberType: MemberTypeUser, IsVerifiedPhone: true}
	if err := runtime.DB.Create(&owner).Error; err != nil {
		t.Fatal(err)
	}
	if err := runtime.DB.Create(&model.UserProfile{UserID: owner.ID}).Error; err != nil {
		t.Fatal(err)
	}
	buildingJSON, _ := json.Marshal([]string{buildingID})
	if err := runtime.DB.Create(&model.UserIsmartAccount{UserID: owner.ID, IsmartUserID: owner.ID + 100, Username: "owner", ClientBuildingPermissions: buildingJSON, Building: []byte("[]"), StaffBuildingPermissions: []byte("[]"), ClientBuildingFlatUnitsPermissions: []byte("[]"), RawMessage: []byte("{}"), ProfileSnapshot: []byte("{}")}).Error; err != nil {
		t.Fatal(err)
	}
	return &owner
}

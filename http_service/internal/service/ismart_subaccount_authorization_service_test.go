/*
 * iSmart 單位級副戶授權流程測試。
 * 1. 未註冊對象發送註冊邀請，已註冊對象發送授權確認。
 * 2. 兩項聯絡資料未命中同一帳戶時阻止提交。
 * 3. 註冊命中待授權記錄後轉為等待確認，確認後才呼叫 iSmart 授予。
 * 4. 逾期失效、撤銷與列表只回傳待處理記錄。
 */
package service

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"ajoliving_web/http_service/internal/config"
	"ajoliving_web/http_service/internal/model"
)

// 1. newSubaccountAuthorizationTestRuntime builds an isolated runtime with one visible owner unit.
func newSubaccountAuthorizationTestRuntime(t *testing.T, upstreamURL string, mailSender MailSender) *Runtime {
	t.Helper()
	databaseName := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	db, err := gorm.Open(sqlite.Open("file:ismart-subaccount-auth-test-"+databaseName+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&model.UserIsmartAccount{},
		&model.IsmartSubaccountAuthorization{},
		&model.User{},
		&model.UserCredential{},
		&model.Notification{},
	); err != nil {
		t.Fatalf("migrate authorization tables: %v", err)
	}
	if err := db.Create(&model.UserIsmartAccount{
		UserID:                             1,
		IsmartUserID:                       111,
		Username:                           "200111",
		ClientBuildingFlatUnitsPermissions: datatypes.JSON([]byte(`["0348200001"]`)),
		Building:                           datatypes.JSON([]byte(`[]`)),
		StaffBuildingPermissions:           datatypes.JSON([]byte(`[]`)),
		ClientBuildingPermissions:          datatypes.JSON([]byte(`[]`)),
		RawMessage:                         datatypes.JSON([]byte(`{}`)),
		ProfileSnapshot:                    datatypes.JSON([]byte(`{}`)),
	}).Error; err != nil {
		t.Fatalf("create owner ismart account: %v", err)
	}

	return &Runtime{
		Config: &config.Config{
			AppPublicBaseURL:            "https://test.ajoliving.skylinedances.com",
			IsmartIntegrationAPIBaseURL: upstreamURL + "/api/v1/integration",
			IsmartSubaccountAPIBaseURL:  upstreamURL + "/api/v1/integration",
			IsmartExternalAppTimeout:    2 * time.Second,
			POSLoginTimeout:             2 * time.Second,
		},
		DB:         db,
		MailSender: mailSender,
		CacheStore: &subaccountTestCache{values: map[string][]byte{
			posBuildingDirectoryCacheKey:               []byte(`[{"building_id":"0348200","buildname_chi":"測試大廈"}]`),
			posBuildingUnitsCacheKeyPrefix + "0348200": []byte(`[{"unit_id":"0348200001","floor":"12","unit":"A"}]`),
		}},
		Now: time.Now,
	}
}

// 2. Upstream routes used by the authorization flow tests.
const (
	subaccountAuthCheckContactPath = "/api/v1/integration/auth/check-contact/"
	subaccountAuthGrantPath        = "/api/v1/integration/buildings/subaccounts/grant/"
)

// 3. TestCreateInvitationSendsRegistrationInvitationForUnregisteredTarget verifies the invitation path.
func TestCreateInvitationSendsRegistrationInvitationForUnregisteredTarget(t *testing.T) {
	var grantCalled bool
	server := newSubaccountIntegrationTestServer(t, func(path string, _ map[string]any) (int, map[string]any) {
		switch path {
		case subaccountAuthCheckContactPath:
			return http.StatusOK, map[string]any{"status": "success", "data": map[string]any{
				"email_available": true,
				"phone_available": true,
			}}
		case subaccountAuthGrantPath:
			grantCalled = true
			return http.StatusOK, map[string]any{"status": "success", "data": map[string]any{}}
		default:
			t.Fatalf("unexpected upstream path %s", path)
			return http.StatusInternalServerError, nil
		}
	})
	defer server.Close()

	mailSender := &capturedSubaccountMailSender{}
	authService := NewIsmartSubaccountAuthorizationService(newSubaccountAuthorizationTestRuntime(t, server.URL, mailSender))
	result, err := authService.CreateInvitation(context.Background(), 1, IsmartSubaccountAuthorizationParams{
		UnitID:    "0348200001",
		Phone:     "+85291234567",
		Email:     "New@Example.com",
		SendEmail: true,
	})
	if err != nil {
		t.Fatalf("create invitation: %v", err)
	}
	if result.Status != IsmartSubaccountAuthStatusPendingRegistration || !result.RequiresRegistration {
		t.Fatalf("unexpected invitation result: %+v", result)
	}
	if !result.EmailSent {
		t.Fatalf("registration invitation email must be sent: %+v", result)
	}
	if grantCalled {
		t.Fatal("unregistered target must not be granted before registration")
	}
	if len(mailSender.Messages) != 1 {
		t.Fatalf("mail count = %d, want 1", len(mailSender.Messages))
	}
	body := mailSender.Messages[0].Body
	if !strings.Contains(body, "/login?") || !strings.Contains(body, "action=register") {
		t.Fatalf("invitation must contain the registration link: %s", body)
	}
	if !strings.Contains(body, "new%40example.com") && !strings.Contains(body, "New%40Example.com") {
		t.Fatalf("invitation link must prefill the target email: %s", body)
	}

	items, err := authService.ListAuthorizations(context.Background(), 1, "0348200001")
	if err != nil {
		t.Fatalf("list authorizations: %v", err)
	}
	if len(items) != 1 || items[0].Status != IsmartSubaccountAuthStatusPendingRegistration {
		t.Fatalf("unexpected pending items: %+v", items)
	}
	if items[0].Phone != "+85291234567" {
		t.Fatalf("unexpected normalized phone: %+v", items[0])
	}
}

// 4. TestCreateInvitationRequiresEmailForUnregisteredTarget verifies the email precondition.
func TestCreateInvitationRequiresEmailForUnregisteredTarget(t *testing.T) {
	server := newSubaccountIntegrationTestServer(t, func(path string, _ map[string]any) (int, map[string]any) {
		if path != subaccountAuthCheckContactPath {
			t.Fatalf("unexpected upstream path %s", path)
		}
		return http.StatusOK, map[string]any{"status": "success", "data": map[string]any{
			"phone_available": true,
		}}
	})
	defer server.Close()

	authService := NewIsmartSubaccountAuthorizationService(newSubaccountAuthorizationTestRuntime(t, server.URL, &capturedSubaccountMailSender{}))
	_, err := authService.CreateInvitation(context.Background(), 1, IsmartSubaccountAuthorizationParams{
		UnitID:    "0348200001",
		Phone:     "+85291234567",
		SendEmail: true,
	})
	if err == nil || !strings.Contains(err.Error(), "email is required") {
		t.Fatalf("expected missing email error, got %v", err)
	}
}

// 5. TestCreateInvitationForRegisteredTargetSendsConfirmationOnly verifies the confirmation path.
func TestCreateInvitationForRegisteredTargetSendsConfirmationOnly(t *testing.T) {
	var grantCalled bool
	server := newSubaccountIntegrationTestServer(t, func(path string, _ map[string]any) (int, map[string]any) {
		switch path {
		case subaccountAuthCheckContactPath:
			return http.StatusOK, map[string]any{"status": "success", "data": map[string]any{
				"email_available": false,
				"email_user_id":   456,
				"phone_available": false,
				"phone_user_id":   456,
			}}
		case subaccountAuthGrantPath:
			grantCalled = true
			return http.StatusOK, map[string]any{"status": "success", "data": map[string]any{}}
		default:
			t.Fatalf("unexpected upstream path %s", path)
			return http.StatusInternalServerError, nil
		}
	})
	defer server.Close()

	mailSender := &capturedSubaccountMailSender{}
	authService := NewIsmartSubaccountAuthorizationService(newSubaccountAuthorizationTestRuntime(t, server.URL, mailSender))
	result, err := authService.CreateInvitation(context.Background(), 1, IsmartSubaccountAuthorizationParams{
		UnitID:    "0348200001",
		Phone:     "+85291234567",
		Email:     "existing@example.com",
		SendEmail: true,
	})
	if err != nil {
		t.Fatalf("create authorization: %v", err)
	}
	if result.Status != IsmartSubaccountAuthStatusAwaitingConfirmation || result.RequiresRegistration {
		t.Fatalf("unexpected confirmation result: %+v", result)
	}
	if result.TargetUserID != 456 {
		t.Fatalf("unexpected target user id: %+v", result)
	}
	if grantCalled {
		t.Fatal("ismart grant must wait for the target confirmation")
	}
	if len(mailSender.Messages) != 1 || !strings.Contains(mailSender.Messages[0].Body, "/auth/ismart-subaccount-authorizations/confirm?token=") {
		t.Fatalf("confirmation email must contain the confirmation link: %+v", mailSender.Messages)
	}
}

// 6. TestCreateInvitationRejectsMismatchedRegisteredContacts verifies the both-fields rule.
func TestCreateInvitationRejectsMismatchedRegisteredContacts(t *testing.T) {
	server := newSubaccountIntegrationTestServer(t, func(path string, _ map[string]any) (int, map[string]any) {
		if path != subaccountAuthCheckContactPath {
			t.Fatalf("unexpected upstream path %s", path)
		}
		return http.StatusOK, map[string]any{"status": "success", "data": map[string]any{
			"email_available": false,
			"email_user_id":   456,
			"phone_available": false,
			"phone_user_id":   999,
		}}
	})
	defer server.Close()

	runtime := newSubaccountAuthorizationTestRuntime(t, server.URL, &capturedSubaccountMailSender{})
	authService := NewIsmartSubaccountAuthorizationService(runtime)
	_, err := authService.CreateInvitation(context.Background(), 1, IsmartSubaccountAuthorizationParams{
		UnitID:    "0348200001",
		Phone:     "+85291234567",
		Email:     "conflict@example.com",
		SendEmail: true,
	})
	if err == nil || !strings.Contains(err.Error(), "same ismart account") {
		t.Fatalf("expected same-account validation error, got %v", err)
	}
	var count int64
	if err := runtime.DB.Model(&model.IsmartSubaccountAuthorization{}).Count(&count).Error; err != nil {
		t.Fatalf("count authorizations: %v", err)
	}
	if count != 0 {
		t.Fatalf("mismatched contacts must not create a record, count = %d", count)
	}
}

// 7. TestConfirmByTokenGrantsExactlyOnce verifies confirmation-driven grant and idempotency.
func TestConfirmByTokenGrantsExactlyOnce(t *testing.T) {
	var grantPayloads []map[string]any
	server := newSubaccountIntegrationTestServer(t, func(path string, payload map[string]any) (int, map[string]any) {
		switch path {
		case subaccountAuthCheckContactPath:
			return http.StatusOK, map[string]any{"status": "success", "data": map[string]any{
				"email_available": false,
				"email_user_id":   456,
				"phone_available": false,
				"phone_user_id":   456,
			}}
		case subaccountAuthGrantPath:
			grantPayloads = append(grantPayloads, payload)
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
	runtime := newSubaccountAuthorizationTestRuntime(t, server.URL, mailSender)
	if err := runtime.DB.Create(&model.User{
		ID:               2,
		PublicID:         "member-456",
		PhoneCountryCode: "+852",
		PhoneNumber:      "91234567",
		MemberStatus:     "active",
	}).Error; err != nil {
		t.Fatalf("create grantee member: %v", err)
	}

	authService := NewIsmartSubaccountAuthorizationService(runtime)
	if _, err := authService.ResolveToken(context.Background(), "unknown-token"); err == nil {
		t.Fatal("unknown token must not resolve")
	}

	if _, err := authService.CreateInvitation(context.Background(), 1, IsmartSubaccountAuthorizationParams{
		UnitID:    "0348200001",
		Phone:     "+85291234567",
		Email:     "existing@example.com",
		SendEmail: true,
	}); err != nil {
		t.Fatalf("create authorization: %v", err)
	}
	token := subaccountAuthTokenFromMail(t, mailSender)

	item, err := authService.ResolveToken(context.Background(), token)
	if err != nil {
		t.Fatalf("resolve token: %v", err)
	}
	if item.Status != IsmartSubaccountAuthStatusAwaitingConfirmation {
		t.Fatalf("unexpected resolved status: %+v", item)
	}

	result, err := authService.ConfirmByToken(context.Background(), token)
	if err != nil {
		t.Fatalf("confirm by token: %v", err)
	}
	if !result.Granted || result.Status != IsmartSubaccountAuthStatusActive {
		t.Fatalf("unexpected confirm result: %+v", result)
	}
	if len(grantPayloads) != 1 {
		t.Fatalf("grant call count = %d, want 1", len(grantPayloads))
	}
	if grantPayloads[0]["user_id"] != float64(111) || grantPayloads[0]["target_user_id"] != float64(456) {
		t.Fatalf("unexpected grant payload: %+v", grantPayloads[0])
	}
	if grantPayloads[0]["unit_id"] != "0348200001" {
		t.Fatalf("grant must target the invited unit: %+v", grantPayloads[0])
	}

	// 以站內入口再次確認時必須視為已生效，且不得重複呼叫 iSmart。
	again, err := authService.ConfirmByID(context.Background(), 2, item.AuthorizationID)
	if err != nil {
		t.Fatalf("confirm by id after grant: %v", err)
	}
	if again.Granted || again.Status != IsmartSubaccountAuthStatusActive {
		t.Fatalf("repeated confirmation must be idempotent: %+v", again)
	}
	if len(grantPayloads) != 1 {
		t.Fatalf("grant call count after repeated confirm = %d, want 1", len(grantPayloads))
	}

	// 已使用過的令牌不得再次確認。
	if _, err := authService.ConfirmByToken(context.Background(), token); err == nil {
		t.Fatal("a consumed confirmation token must not be reusable")
	}
}

// 8. TestProcessRegistrationPromotesPendingInvitation verifies the registration hook.
func TestProcessRegistrationPromotesPendingInvitation(t *testing.T) {
	var grantPayloads []map[string]any
	server := newSubaccountIntegrationTestServer(t, func(path string, payload map[string]any) (int, map[string]any) {
		switch path {
		case subaccountAuthCheckContactPath:
			return http.StatusOK, map[string]any{"status": "success", "data": map[string]any{
				"email_available": true,
				"phone_available": true,
			}}
		case subaccountAuthGrantPath:
			grantPayloads = append(grantPayloads, payload)
			return http.StatusOK, map[string]any{"status": "success", "data": map[string]any{}}
		default:
			t.Fatalf("unexpected upstream path %s", path)
			return http.StatusInternalServerError, nil
		}
	})
	defer server.Close()

	mailSender := &capturedSubaccountMailSender{}
	runtime := newSubaccountAuthorizationTestRuntime(t, server.URL, mailSender)
	authService := NewIsmartSubaccountAuthorizationService(runtime)
	if _, err := authService.CreateInvitation(context.Background(), 1, IsmartSubaccountAuthorizationParams{
		UnitID:    "0348200001",
		Phone:     "+85291234567",
		Email:     "newmember@example.com",
		SendEmail: true,
	}); err != nil {
		t.Fatalf("create invitation: %v", err)
	}
	if len(mailSender.Messages) != 1 {
		t.Fatalf("invitation mail count = %d, want 1", len(mailSender.Messages))
	}

	grantee := model.User{
		ID:               2,
		PublicID:         "member-789",
		PhoneCountryCode: "+852",
		PhoneNumber:      "91234567",
		MemberStatus:     "active",
	}
	if err := runtime.DB.Create(&grantee).Error; err != nil {
		t.Fatalf("create grantee: %v", err)
	}
	email := "newmember@example.com"
	if err := runtime.DB.Create(&model.UserCredential{UserID: grantee.ID, Email: &email, PasswordHash: "x"}).Error; err != nil {
		t.Fatalf("create grantee credential: %v", err)
	}
	if err := runtime.DB.Create(&model.UserIsmartAccount{
		UserID:                             grantee.ID,
		IsmartUserID:                       789,
		Username:                           "200789",
		ClientBuildingFlatUnitsPermissions: datatypes.JSON([]byte(`[]`)),
		Building:                           datatypes.JSON([]byte(`[]`)),
		StaffBuildingPermissions:           datatypes.JSON([]byte(`[]`)),
		ClientBuildingPermissions:          datatypes.JSON([]byte(`[]`)),
		RawMessage:                         datatypes.JSON([]byte(`{}`)),
		ProfileSnapshot:                    datatypes.JSON([]byte(`{}`)),
	}).Error; err != nil {
		t.Fatalf("create grantee ismart account: %v", err)
	}

	if err := authService.ProcessRegistrationForPendingAuthorizations(context.Background(), grantee.ID); err != nil {
		t.Fatalf("process registration: %v", err)
	}

	items, err := authService.ListAuthorizations(context.Background(), 1, "0348200001")
	if err != nil {
		t.Fatalf("list authorizations: %v", err)
	}
	if len(items) != 1 || items[0].Status != IsmartSubaccountAuthStatusAwaitingConfirmation {
		t.Fatalf("registration must promote the record to awaiting confirmation: %+v", items)
	}
	if items[0].TargetUserID != 789 {
		t.Fatalf("unexpected target user id after promotion: %+v", items[0])
	}
	if len(mailSender.Messages) != 2 {
		t.Fatalf("confirmation mail count = %d, want 2", len(mailSender.Messages))
	}
	if !strings.Contains(mailSender.Messages[1].Body, "confirm?token=") {
		t.Fatalf("promotion must send a confirmation link: %s", mailSender.Messages[1].Body)
	}

	result, err := authService.ConfirmByID(context.Background(), grantee.ID, items[0].AuthorizationID)
	if err != nil {
		t.Fatalf("confirm by id: %v", err)
	}
	if !result.Granted {
		t.Fatalf("registered grantee must be granted after confirmation: %+v", result)
	}
	if len(grantPayloads) != 1 || grantPayloads[0]["target_user_id"] != float64(789) {
		t.Fatalf("unexpected grant payload: %+v", grantPayloads)
	}

	// 非本人不得確認。
	other := model.User{ID: 3, PublicID: "member-999", PhoneCountryCode: "+852", PhoneNumber: "60000000", MemberStatus: "active"}
	if err := runtime.DB.Create(&other).Error; err != nil {
		t.Fatalf("create other member: %v", err)
	}
	if _, err := authService.ConfirmByID(context.Background(), other.ID, items[0].AuthorizationID); err == nil {
		t.Fatal("another member must not confirm someone else's authorization")
	}
	// 已生效的記錄不得走本地撤銷，必須透過 iSmart 撤銷。
	if err := authService.Revoke(context.Background(), 1, items[0].AuthorizationID); err == nil {
		t.Fatal("active authorization must not be revoked locally")
	}
}

// 9. TestExpireAndRevokeAuthorizations verifies expiry and revocation rules.
func TestExpireAndRevokeAuthorizations(t *testing.T) {
	server := newSubaccountIntegrationTestServer(t, func(path string, _ map[string]any) (int, map[string]any) {
		if path != subaccountAuthCheckContactPath {
			t.Fatalf("unexpected upstream path %s", path)
		}
		return http.StatusOK, map[string]any{"status": "success", "data": map[string]any{
			"email_available": true,
			"phone_available": true,
		}}
	})
	defer server.Close()

	runtime := newSubaccountAuthorizationTestRuntime(t, server.URL, &capturedSubaccountMailSender{})
	authService := NewIsmartSubaccountAuthorizationService(runtime)
	created, err := authService.CreateInvitation(context.Background(), 1, IsmartSubaccountAuthorizationParams{
		UnitID:    "0348200001",
		Phone:     "+85291234567",
		Email:     "expire@example.com",
		SendEmail: false,
	})
	if err != nil {
		t.Fatalf("create invitation: %v", err)
	}
	if created.EmailSent {
		t.Fatal("send_email=false must not send an invitation")
	}

	if err := runtime.DB.Model(&model.IsmartSubaccountAuthorization{}).
		Where("public_id = ?", created.AuthorizationID).
		Update("invitation_expires_at", time.Now().Add(-time.Hour)).Error; err != nil {
		t.Fatalf("expire invitation: %v", err)
	}
	expired, err := authService.ExpireOverdueAuthorizations(context.Background())
	if err != nil {
		t.Fatalf("expire overdue authorizations: %v", err)
	}
	if expired != 1 {
		t.Fatalf("expired count = %d, want 1", expired)
	}
	if _, err := authService.ResendNotification(context.Background(), 1, created.AuthorizationID, true); err == nil {
		t.Fatal("expired authorization must not be resent")
	}

	revoked, err := authService.CreateInvitation(context.Background(), 1, IsmartSubaccountAuthorizationParams{
		UnitID:    "0348200001",
		Phone:     "+85291234567",
		Email:     "revoke@example.com",
		SendEmail: false,
	})
	if err != nil {
		t.Fatalf("create revocable invitation: %v", err)
	}
	if err := authService.Revoke(context.Background(), 1, revoked.AuthorizationID); err != nil {
		t.Fatalf("revoke authorization: %v", err)
	}
	items, err := authService.ListAuthorizations(context.Background(), 1, "0348200001")
	if err != nil {
		t.Fatalf("list authorizations: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("revoked and expired records must leave the pending list: %+v", items)
	}
	// 重複撤銷為冪等操作，但其他主戶不得撤銷不屬於自己的記錄。
	if err := authService.Revoke(context.Background(), 1, revoked.AuthorizationID); err != nil {
		t.Fatalf("repeated revoke must stay idempotent: %v", err)
	}
	if err := authService.Revoke(context.Background(), 2, revoked.AuthorizationID); err == nil {
		t.Fatal("another owner must not revoke someone else's authorization")
	}
}

// 10. subaccountAuthTokenFromMail extracts the confirmation token from the latest notice.
func subaccountAuthTokenFromMail(t *testing.T, sender *capturedSubaccountMailSender) string {
	t.Helper()
	if len(sender.Messages) == 0 {
		return ""
	}
	body := sender.Messages[len(sender.Messages)-1].Body
	marker := "confirm?token="
	index := strings.Index(body, marker)
	if index < 0 {
		t.Fatalf("confirmation link missing from mail body: %s", body)
	}
	token := body[index+len(marker):]
	if end := strings.IndexAny(token, " \n\r\t\"<>"); end >= 0 {
		token = token[:end]
	}
	return strings.TrimSpace(token)
}

// 11. TestSubaccountAuthPhonePartsKeepsLocalNumber verifies country-code splitting stays compatible.
func TestSubaccountAuthPhonePartsKeepsLocalNumber(t *testing.T) {
	for _, tc := range []struct {
		phone       string
		countryCode string
		number      string
	}{
		{phone: "+85291234567", countryCode: "+852", number: "91234567"},
		{phone: "91234567", countryCode: "", number: "91234567"},
		{phone: "+8613800000000", countryCode: "+86", number: "13800000000"},
		{phone: "", countryCode: "", number: ""},
	} {
		countryCode, number := subaccountAuthPhoneParts(tc.phone)
		if countryCode != tc.countryCode || number != tc.number {
			t.Fatalf("subaccountAuthPhoneParts(%q) = (%q, %q), want (%q, %q)", tc.phone, countryCode, number, tc.countryCode, tc.number)
		}
	}
}

// 12. TestSubaccountConfirmRejectsUnknownToken verifies missing records stay errors.
func TestSubaccountConfirmRejectsUnknownToken(t *testing.T) {
	runtime := newSubaccountAuthorizationTestRuntime(t, "http://127.0.0.1:1", &capturedSubaccountMailSender{})
	authService := NewIsmartSubaccountAuthorizationService(runtime)
	if _, err := authService.ConfirmByToken(context.Background(), "missing-token"); err == nil {
		t.Fatal("unknown confirmation token must fail")
	}
	if _, err := authService.ConfirmByToken(context.Background(), ""); err == nil {
		t.Fatal("blank confirmation token must fail")
	}
}

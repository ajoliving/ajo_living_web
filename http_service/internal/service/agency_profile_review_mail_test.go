/*
 * 代理資料審核郵件測試。
 * 1. 驗證拒絕郵件包含審核原因與會員中心入口。
 * 2. 驗證郵件發送失敗不會回滾已完成的審核結果。
 */
package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"ajoliving_web/http_service/internal/config"
	"ajoliving_web/http_service/internal/database"
	"ajoliving_web/http_service/internal/model"
	"ajoliving_web/http_service/internal/utils"
)

type agencyReviewMailSender struct {
	recipient string
	subject   string
	body      string
	err       error
}

// 1. Send captures review email delivery for assertions.
func (s *agencyReviewMailSender) Send(_ context.Context, recipient string, subject string, body string) error {
	s.recipient, s.subject, s.body = recipient, subject, body
	return s.err
}

// 2. TestAgencyProfileReviewSendsRejectionEmail verifies the registered email receives the reason.
func TestAgencyProfileReviewSendsRejectionEmail(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open("file:"+utils.NewPublicID()+"?mode=memory&cache=shared"), &gorm.Config{})
	_ = database.Migrate(db)
	sender := &agencyReviewMailSender{}
	runtime := &Runtime{
		Config:     &config.Config{EncryptionKey: "review-mail-test", AppPublicBaseURL: "https://ajo.test"},
		DB:         db,
		MailSender: sender,
		Now:        time.Now,
	}
	user := createAgencyProfileTestUser(t, db, AccountTypeIndividualAgent, "pending_profile", "65550001")
	email := "agent@example.com"
	if err := db.Create(&model.UserCredential{UserID: user.ID, Email: &email, PasswordHash: "test", PasswordEncrypted: "test"}).Error; err != nil {
		t.Fatal(err)
	}
	service := NewAgencyCompanyService(runtime)
	params := AgencyProfileUpsertParams{ProfileType: AgencyProfileTypeIndividual, NameZH: "測試代理", NameEN: "Test Agent", Phone1CountryCode: "+852", Phone1Number: "65550001", DefaultAvatar: "male", IsOverseas: true}
	if _, err := service.CreateMemberAgencyProfile(context.Background(), user.ID, params); err != nil {
		t.Fatal(err)
	}
	pending, err := service.SubmitMemberAgencyProfile(context.Background(), user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.ReviewAgencyProfile(context.Background(), pending.Revision.ProfileID, AgencyProfileReviewParams{ReviewerUserID: user.ID, ReviewNote: "牌照圖片不清晰", Approved: false}); err != nil {
		t.Fatal(err)
	}
	if sender.recipient != email || !strings.Contains(sender.subject, "未通過") || !strings.Contains(sender.body, "牌照圖片不清晰") || !strings.Contains(sender.body, "/account/profile/agency-profile") {
		t.Fatalf("unexpected review email: %#v", sender)
	}
}

// 3. TestAgencyProfileReviewMailFailureDoesNotRollback verifies review remains committed.
func TestAgencyProfileReviewMailFailureDoesNotRollback(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open("file:"+utils.NewPublicID()+"?mode=memory&cache=shared"), &gorm.Config{})
	_ = database.Migrate(db)
	runtime := &Runtime{
		Config:     &config.Config{EncryptionKey: "review-mail-failure"},
		DB:         db,
		MailSender: &agencyReviewMailSender{err: errors.New("smtp unavailable")},
		Now:        time.Now,
	}
	user := createAgencyProfileTestUser(t, db, AccountTypeIndividualAgent, "pending_profile", "65550002")
	email := "failure@example.com"
	_ = db.Create(&model.UserCredential{UserID: user.ID, Email: &email, PasswordHash: "test", PasswordEncrypted: "test"}).Error
	service := NewAgencyCompanyService(runtime)
	params := AgencyProfileUpsertParams{ProfileType: AgencyProfileTypeIndividual, NameZH: "電郵失敗代理", NameEN: "Mail Failure Agent", Phone1CountryCode: "+852", Phone1Number: "65550002", DefaultAvatar: "male", IsOverseas: true}
	_, _ = service.CreateMemberAgencyProfile(context.Background(), user.ID, params)
	pending, _ := service.SubmitMemberAgencyProfile(context.Background(), user.ID)
	if _, err := service.ReviewAgencyProfile(context.Background(), pending.Revision.ProfileID, AgencyProfileReviewParams{ReviewerUserID: user.ID, Approved: true}); err != nil {
		t.Fatalf("mail failure rolled back review: %v", err)
	}
	var stored model.User
	if err := db.First(&stored, user.ID).Error; err != nil || stored.MemberStatus != "active" {
		t.Fatalf("review was not committed: %#v %v", stored, err)
	}
}

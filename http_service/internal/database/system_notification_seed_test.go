/*
 * System notification account seed tests.
 * 1. Validate built-in notification sender account creation.
 * 2. Validate profile display name synchronization.
 */
package database

import (
	"context"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"ajoliving_web/http_service/internal/model"
	"ajoliving_web/http_service/internal/utils"
)

// 1. TestSeedSystemNotificationAccountCreatesProfile validates the system sender seed.
func TestSeedSystemNotificationAccountCreatesProfile(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+utils.NewPublicID()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	if err := Migrate(db); err != nil {
		t.Fatalf("migrate db: %v", err)
	}
	if err := SeedSystemNotificationAccount(context.Background(), db); err != nil {
		t.Fatalf("seed system notification account: %v", err)
	}

	var user model.User
	if err := db.Where("phone_country_code = ? AND phone_number = ?", model.SystemNotificationPhoneCountryCode, model.SystemNotificationPhoneNumber).First(&user).Error; err != nil {
		t.Fatalf("load system notification user: %v", err)
	}
	if user.MemberStatus != "active" || user.IsStaff {
		t.Fatalf("expected active non-staff system notification user, got %+v", user)
	}

	var profile model.UserProfile
	if err := db.Where("user_id = ?", user.ID).First(&profile).Error; err != nil {
		t.Fatalf("load system notification profile: %v", err)
	}
	if profile.DisplayName != model.SystemNotificationDisplayName || profile.PublisherIdentityType != "system" {
		t.Fatalf("expected system notification profile, got %+v", profile)
	}
}

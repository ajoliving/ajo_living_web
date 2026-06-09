/*
 * Default admin account seed tests.
 * 1. Validate default admin email credential creation.
 * 2. Validate staff flag and password reset behavior.
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

// 1. TestSeedDefaultAdminAccountCreatesLoginCredential validates the default admin seed.
func TestSeedDefaultAdminAccountCreatesLoginCredential(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+utils.NewPublicID()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	if err := Migrate(db); err != nil {
		t.Fatalf("migrate db: %v", err)
	}
	if err := SeedDefaultAdminAccount(context.Background(), db); err != nil {
		t.Fatalf("seed default admin: %v", err)
	}

	for _, email := range []string{defaultAdminEmail, defaultAdminSecondEmail} {
		var credential model.UserCredential
		if err := db.Where("email = ?", email).First(&credential).Error; err != nil {
			t.Fatalf("load default admin credential %s: %v", email, err)
		}
		if !utils.VerifyPassword(defaultAdminPassword, credential.PasswordHash) {
			t.Fatalf("expected default admin password to verify for %s", email)
		}

		var user model.User
		if err := db.First(&user, credential.UserID).Error; err != nil {
			t.Fatalf("load default admin user %s: %v", email, err)
		}
		if !user.IsStaff || user.MemberStatus != "active" {
			t.Fatalf("expected active staff user for %s, got %+v", email, user)
		}

	}
}

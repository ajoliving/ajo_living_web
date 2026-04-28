/*
 * Default admin account seed tests.
 * 1. Validate default admin email credential creation.
 * 2. Validate staff role binding and password reset behavior.
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
	if err := SeedAccessControl(context.Background(), db); err != nil {
		t.Fatalf("seed access control: %v", err)
	}
	if err := SeedDefaultAdminAccount(context.Background(), db); err != nil {
		t.Fatalf("seed default admin: %v", err)
	}

	var credential model.UserCredential
	if err := db.Where("email = ?", defaultAdminEmail).First(&credential).Error; err != nil {
		t.Fatalf("load default admin credential: %v", err)
	}
	if !utils.VerifyPassword(defaultAdminPassword, credential.PasswordHash) {
		t.Fatalf("expected default admin password to verify")
	}

	var user model.User
	if err := db.First(&user, credential.UserID).Error; err != nil {
		t.Fatalf("load default admin user: %v", err)
	}
	if !user.IsStaff || user.MemberStatus != "active" {
		t.Fatalf("expected active staff user, got %+v", user)
	}

	var binding model.UserRoleBinding
	err = db.
		Joins("JOIN roles ON roles.id = user_role_bindings.role_id").
		Where("user_role_bindings.user_id = ? AND roles.code = ?", user.ID, model.RoleCodeSuperAdmin).
		First(&binding).
		Error
	if err != nil {
		t.Fatalf("load default admin role binding: %v", err)
	}
}

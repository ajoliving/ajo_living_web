/*
 * Default admin account seed utilities.
 * 1. Ensure the local default admin email account exists.
 * 2. Keep the default password and super admin role synchronized at startup.
 */
package database

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"ajoliving_web/http_service/internal/model"
	"ajoliving_web/http_service/internal/utils"
)

const (
	defaultAdminEmail       = "admin@admin.com"
	defaultAdminPassword    = "admin123"
	defaultAdminDisplayName = "Admin"
)

// 1. SeedDefaultAdminAccount upserts the default local admin account.
func SeedDefaultAdminAccount(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		user, err := upsertDefaultAdminUser(tx)
		if err != nil {
			return err
		}

		if err := upsertDefaultAdminCredential(tx, user.ID); err != nil {
			return err
		}

		if err := upsertDefaultAdminProfile(tx, user.ID); err != nil {
			return err
		}

		if err := upsertDefaultAdminRole(tx, user.ID); err != nil {
			return err
		}

		return nil
	})
}

// 2. upsertDefaultAdminUser creates or promotes the default admin user.
func upsertDefaultAdminUser(tx *gorm.DB) (*model.User, error) {
	var credential model.UserCredential
	if err := tx.Where("email = ?", defaultAdminEmail).Limit(1).Find(&credential).Error; err != nil {
		return nil, err
	}

	user := model.User{}
	if credential.UserID > 0 {
		if err := tx.First(&user, credential.UserID).Error; err != nil {
			return nil, err
		}
	}

	if user.ID == 0 {
		user = model.User{
			PublicID:         utils.NewPublicID(),
			PhoneCountryCode: "email",
			PhoneNumber:      utils.NewPublicID(),
			MemberStatus:     "active",
			MemberType:       "user",
			IsStaff:          true,
			IsVerifiedPhone:  false,
		}
		if err := tx.Create(&user).Error; err != nil {
			return nil, err
		}
		return &user, nil
	}

	if err := tx.Model(&user).Updates(map[string]any{
		"member_status": "active",
		"member_type":   "user",
		"is_staff":      true,
	}).Error; err != nil {
		return nil, err
	}
	user.MemberStatus = "active"
	user.MemberType = "user"
	user.IsStaff = true

	return &user, nil
}

// 3. upsertDefaultAdminCredential creates or resets the default admin credential.
func upsertDefaultAdminCredential(tx *gorm.DB, userID int64) error {
	passwordHash, err := utils.HashPassword(defaultAdminPassword)
	if err != nil {
		return fmt.Errorf("hash default admin password: %w", err)
	}

	var credential model.UserCredential
	if err := tx.Where("email = ?", defaultAdminEmail).Limit(1).Find(&credential).Error; err != nil {
		return err
	}

	if credential.UserID == 0 {
		return tx.Create(&model.UserCredential{
			UserID:       userID,
			Email:        defaultAdminEmail,
			PasswordHash: passwordHash,
			IsVerified:   true,
		}).Error
	}

	return tx.Model(&credential).Updates(map[string]any{
		"user_id":       userID,
		"password_hash": passwordHash,
		"is_verified":   true,
	}).Error
}

// 4. upsertDefaultAdminProfile creates the default admin profile when absent.
func upsertDefaultAdminProfile(tx *gorm.DB, userID int64) error {
	var profile model.UserProfile
	if err := tx.Where("user_id = ?", userID).Limit(1).Find(&profile).Error; err != nil {
		return err
	}

	if profile.UserID == 0 {
		return tx.Create(&model.UserProfile{
			UserID:                userID,
			DisplayName:           defaultAdminDisplayName,
			PublisherIdentityType: "staff",
		}).Error
	}

	if profile.DisplayName != "" {
		return nil
	}

	return tx.Model(&profile).Update("display_name", defaultAdminDisplayName).Error
}

// 5. upsertDefaultAdminRole binds the default admin user to the super admin role.
func upsertDefaultAdminRole(tx *gorm.DB, userID int64) error {
	var role model.Role
	if err := tx.Where("code = ?", model.RoleCodeSuperAdmin).First(&role).Error; err != nil {
		return err
	}

	var binding model.UserRoleBinding
	if err := tx.Where("user_id = ? AND role_id = ?", userID, role.ID).Limit(1).Find(&binding).Error; err != nil {
		return err
	}
	if binding.ID > 0 {
		return nil
	}

	return tx.Create(&model.UserRoleBinding{
		UserID:     userID,
		RoleID:     role.ID,
		AssignedAt: time.Now(),
	}).Error
}

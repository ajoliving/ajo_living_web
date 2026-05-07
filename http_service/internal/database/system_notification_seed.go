/*
 * System notification account seed utilities.
 * 1. Ensure the built-in notification sender account exists.
 * 2. Keep its member profile stable for read-only system conversations.
 */
package database

import (
	"context"

	"gorm.io/gorm"

	"ajoliving_web/http_service/internal/model"
	"ajoliving_web/http_service/internal/utils"
)

// 1. SeedSystemNotificationAccount upserts the built-in notification sender.
func SeedSystemNotificationAccount(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		user, err := upsertSystemNotificationUser(tx)
		if err != nil {
			return err
		}

		return upsertSystemNotificationProfile(tx, user.ID)
	})
}

// 2. upsertSystemNotificationUser creates the system notification user.
func upsertSystemNotificationUser(tx *gorm.DB) (*model.User, error) {
	var user model.User
	if err := tx.Where("phone_country_code = ? AND phone_number = ?", model.SystemNotificationPhoneCountryCode, model.SystemNotificationPhoneNumber).Limit(1).Find(&user).Error; err != nil {
		return nil, err
	}

	if user.ID == 0 {
		user = model.User{
			PublicID:         utils.NewPublicID(),
			PhoneCountryCode: model.SystemNotificationPhoneCountryCode,
			PhoneNumber:      model.SystemNotificationPhoneNumber,
			MemberStatus:     "active",
			MemberType:       "user",
			IsStaff:          false,
			IsVerifiedPhone:  false,
		}
		if err := tx.Create(&user).Error; err != nil {
			return nil, err
		}
		return &user, nil
	}

	if err := tx.Model(&user).Updates(map[string]any{
		"member_status":     "active",
		"member_type":       "user",
		"is_staff":          false,
		"is_verified_phone": false,
	}).Error; err != nil {
		return nil, err
	}
	user.MemberStatus = "active"
	user.MemberType = "user"
	user.IsStaff = false
	user.IsVerifiedPhone = false

	return &user, nil
}

// 3. upsertSystemNotificationProfile creates the system notification profile.
func upsertSystemNotificationProfile(tx *gorm.DB, userID int64) error {
	var profile model.UserProfile
	if err := tx.Where("user_id = ?", userID).Limit(1).Find(&profile).Error; err != nil {
		return err
	}

	if profile.UserID == 0 {
		return tx.Create(&model.UserProfile{
			UserID:                userID,
			DisplayName:           model.SystemNotificationDisplayName,
			PublisherIdentityType: "system",
		}).Error
	}

	return tx.Model(&profile).Updates(map[string]any{
		"display_name":            model.SystemNotificationDisplayName,
		"publisher_identity_type": "system",
	}).Error
}

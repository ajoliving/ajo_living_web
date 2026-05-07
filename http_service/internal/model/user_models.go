/*
 * User and community data models.
 * 1. Define member account, profile, and community tables.
 * 2. Keep public IDs and core indexes explicit for GORM migration.
 */
package model

import "time"

const (
	SystemNotificationPhoneCountryCode = "system"
	SystemNotificationPhoneNumber      = "notification"
	SystemNotificationDisplayName      = "通知"
)

// 1. User stores member account records.
type User struct {
	ID               int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	PublicID         string `gorm:"type:varchar(26);not null;uniqueIndex" json:"public_id"`
	PhoneCountryCode string `gorm:"type:varchar(8);not null;uniqueIndex:uk_user_phone" json:"phone_country_code"`
	PhoneNumber      string `gorm:"type:varchar(32);not null;uniqueIndex:uk_user_phone" json:"phone_number"`
	MemberStatus     string `gorm:"type:varchar(32);not null;index" json:"member_status"`
	MemberType       string `gorm:"type:varchar(32);not null;default:user;index" json:"member_type"`
	IsStaff          bool   `gorm:"not null;default:false;index" json:"is_staff"`
	IsVerifiedPhone  bool   `gorm:"not null" json:"is_verified_phone"`
	TimestampModel
}

// 2. UserCredential stores email and password login credentials.
type UserCredential struct {
	UserID       int64     `gorm:"primaryKey" json:"user_id"`
	Email        string    `gorm:"type:varchar(255);not null;uniqueIndex" json:"email"`
	PasswordHash string    `gorm:"type:text" json:"password_hash"`
	IsVerified   bool      `gorm:"not null;default:true" json:"is_verified"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	User         *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// 3. UserProfile stores profile details and community linkage.
type UserProfile struct {
	UserID                int64      `gorm:"primaryKey" json:"user_id"`
	DisplayName           string     `gorm:"type:varchar(120)" json:"display_name"`
	PublisherIdentityType string     `gorm:"type:varchar(32)" json:"publisher_identity_type"`
	PrimaryCommunityID    *int64     `gorm:"index" json:"primary_community_id"`
	DistrictCode          string     `gorm:"type:varchar(32)" json:"district_code"`
	AvatarAssetID         *int64     `json:"avatar_asset_id"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
	PrimaryCommunity      *Community `gorm:"foreignKey:PrimaryCommunityID" json:"primary_community,omitempty"`
}

// 4. Community stores estate and building records.
type Community struct {
	ID                int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	PublicID          string     `gorm:"type:varchar(26);not null;uniqueIndex" json:"public_id"`
	CommunityType     string     `gorm:"type:varchar(32);not null" json:"community_type"`
	NameZH            string     `gorm:"type:varchar(200);not null;index" json:"name_zh"`
	NameEN            string     `gorm:"type:varchar(200)" json:"name_en"`
	DistrictCode      string     `gorm:"type:varchar(32);not null;index" json:"district_code"`
	ParentCommunityID *int64     `gorm:"index" json:"parent_community_id"`
	AddressText       string     `gorm:"type:varchar(500)" json:"address_text"`
	CreatedAt         time.Time  `json:"created_at"`
	ParentCommunity   *Community `gorm:"foreignKey:ParentCommunityID" json:"parent_community,omitempty"`
}

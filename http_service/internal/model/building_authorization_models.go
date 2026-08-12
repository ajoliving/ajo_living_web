/*
 * 大廈副戶授權資料模型。
 * 1. 保存主戶按大廈授予副戶的功能權限。
 * 2. 保存未註冊副戶的一次性啟用邀請狀態。
 */
package model

import "time"

// 1. BuildingAuthorization stores one owner-to-member building authorization.
type BuildingAuthorization struct {
	ID                  int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	PublicID            string     `gorm:"type:varchar(26);not null;uniqueIndex" json:"public_id"`
	BuildingID          string     `gorm:"type:varchar(120);not null;index:idx_building_auth_building_owner" json:"building_id"`
	OwnerUserID         int64      `gorm:"not null;index:idx_building_auth_building_owner" json:"owner_user_id"`
	GranteeUserID       *int64     `gorm:"index" json:"grantee_user_id,omitempty"`
	PhoneCountryCode    string     `gorm:"type:varchar(8);not null" json:"phone_country_code"`
	PhoneNumber         string     `gorm:"type:varchar(32);not null" json:"phone_number"`
	Email               string     `gorm:"type:varchar(255);not null" json:"email"`
	Permissions         []byte     `gorm:"type:jsonb;not null" json:"permissions"`
	Status              string     `gorm:"type:varchar(24);not null;index" json:"status"`
	InvitationTokenHash string     `gorm:"type:varchar(64);index" json:"-"`
	InvitationExpiresAt *time.Time `json:"invitation_expires_at,omitempty"`
	AcceptedAt          *time.Time `json:"accepted_at,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
	Owner               *User      `gorm:"foreignKey:OwnerUserID" json:"owner,omitempty"`
	Grantee             *User      `gorm:"foreignKey:GranteeUserID" json:"grantee,omitempty"`
}

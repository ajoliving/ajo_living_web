/*
 * iSmart 單位級副戶授權資料模型。
 * 1. 保存主戶對某一單位發起的副戶授權，涵蓋待註冊、待確認、已生效三個主狀態。
 * 2. 保存註冊邀請與授權確認的令牌雜湊及有效期，令牌只以 SHA-256 雜湊落庫。
 * 3. 保存被綁定人的 iSmart 目標用戶資訊、通知重發次數與確認、生效時間。
 */
package model

import "time"

// 1. IsmartSubaccountAuthorization stores one unit-level iSmart subaccount authorization
// waiting for the invitee registration and/or explicit confirmation.
type IsmartSubaccountAuthorization struct {
	ID                  int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	PublicID            string     `gorm:"type:varchar(26);not null;uniqueIndex" json:"public_id"`
	OwnerUserID         int64      `gorm:"not null;index:idx_ismart_subaccount_auth_owner_status,priority:1" json:"owner_user_id"`
	GranteeUserID       *int64     `gorm:"index" json:"grantee_user_id,omitempty"`
	BuildingID          string     `gorm:"type:varchar(120);index" json:"building_id"`
	UnitID              string     `gorm:"type:varchar(120);not null;index" json:"unit_id"`
	TargetUserID        int64      `gorm:"not null;default:0" json:"target_user_id"`
	TargetUsername      string     `gorm:"type:varchar(120)" json:"target_username"`
	PhoneCountryCode    string     `gorm:"type:varchar(8);not null;default:''" json:"phone_country_code"`
	PhoneNumber         string     `gorm:"type:varchar(32);index" json:"phone_number"`
	Email               string     `gorm:"type:varchar(255);index" json:"email"`
	Status              string     `gorm:"type:varchar(32);not null;index:idx_ismart_subaccount_auth_owner_status,priority:2" json:"status"`
	ConfirmTokenHash    string     `gorm:"type:varchar(64);index" json:"-"`
	ConfirmExpiresAt    *time.Time `json:"confirm_expires_at,omitempty"`
	InvitationExpiresAt *time.Time `json:"invitation_expires_at,omitempty"`
	ConfirmedAt         *time.Time `json:"confirmed_at,omitempty"`
	GrantedAt           *time.Time `json:"granted_at,omitempty"`
	ResendCount         int        `gorm:"not null;default:0" json:"resend_count"`
	LastSentAt          *time.Time `json:"last_sent_at,omitempty"`
	Remark              string     `gorm:"type:varchar(500)" json:"remark"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`

	Owner   *User `gorm:"foreignKey:OwnerUserID" json:"owner,omitempty"`
	Grantee *User `gorm:"foreignKey:GranteeUserID" json:"grantee,omitempty"`
}

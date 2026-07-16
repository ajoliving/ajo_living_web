/*
 * Agency account data models.
 * 1. Keep one active approved profile and one editable review revision per agency account.
 * 2. Store individual and company evidence as explicit media references.
 * 3. Bind company subaccounts to one approved company owner.
 */
package model

import (
	"time"

	"gorm.io/datatypes"
)

// 1. AgencyProfileBinding keeps active and working profile versions for one account.
type AgencyProfileBinding struct {
	UserID            int64      `gorm:"primaryKey" json:"user_id"`
	ActiveProfileID   *int64     `gorm:"index" json:"active_profile_id"`
	RevisionProfileID *int64     `gorm:"index" json:"revision_profile_id"`
	LastEditedAt      *time.Time `json:"last_edited_at"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// 2. AgencyProfile stores one approved snapshot or one review revision.
type AgencyProfile struct {
	ID                          int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	PublicID                    string     `gorm:"type:varchar(26);not null;uniqueIndex" json:"public_id"`
	UserID                      int64      `gorm:"not null;index" json:"user_id"`
	ProfileType                 string     `gorm:"type:varchar(32);not null;index" json:"profile_type"`
	NameZH                      string     `gorm:"type:varchar(120)" json:"name_zh"`
	NameEN                      string     `gorm:"type:varchar(200)" json:"name_en"`
	AddressZH                   string     `gorm:"type:varchar(500)" json:"address_zh"`
	AddressEN                   string     `gorm:"type:varchar(500)" json:"address_en"`
	LicenseNumber               string     `gorm:"type:varchar(120);index" json:"license_number"`
	IsOverseas                  bool       `gorm:"not null;default:false" json:"is_overseas"`
	IsBigFour                   bool       `gorm:"not null;default:false" json:"is_big_four"`
	Phone1CountryCode           string     `gorm:"type:varchar(16)" json:"phone_1_country_code"`
	Phone1Number                string     `gorm:"type:varchar(32)" json:"phone_1_number"`
	Phone1WhatsApp              bool       `gorm:"not null;default:false" json:"phone_1_whatsapp"`
	Phone2CountryCode           string     `gorm:"type:varchar(16)" json:"phone_2_country_code"`
	Phone2Number                string     `gorm:"type:varchar(32)" json:"phone_2_number"`
	Phone2WhatsApp              bool       `gorm:"not null;default:false" json:"phone_2_whatsapp"`
	WechatID                    string     `gorm:"type:varchar(120)" json:"wechat_id"`
	WechatURL                   string     `gorm:"type:varchar(1000)" json:"wechat_url"`
	SignatureZH                 string     `gorm:"type:varchar(300)" json:"signature_zh"`
	SignatureEN                 string     `gorm:"type:varchar(1000)" json:"signature_en"`
	DefaultAvatar               string     `gorm:"type:varchar(16)" json:"default_avatar"`
	AvatarAssetID               *int64     `gorm:"index" json:"avatar_asset_id"`
	WechatQRAssetID             *int64     `gorm:"index" json:"wechat_qr_asset_id"`
	LogoAssetID                 *int64     `gorm:"index" json:"logo_asset_id"`
	EAALicenseAssetID           *int64     `gorm:"index" json:"eaa_license_asset_id"`
	BusinessRegistrationAssetID *int64     `gorm:"index" json:"business_registration_asset_id"`
	CompanyCardAssetID          *int64     `gorm:"index" json:"company_card_asset_id"`
	Status                      string     `gorm:"type:varchar(32);not null;index" json:"status"`
	ReviewNote                  string     `gorm:"type:varchar(1000)" json:"review_note"`
	SubmittedAt                 *time.Time `json:"submitted_at"`
	ReviewedAt                  *time.Time `json:"reviewed_at"`
	ReviewedByUserID            *int64     `gorm:"index" json:"reviewed_by_user_id"`
	CreatedAt                   time.Time  `json:"created_at"`
	UpdatedAt                   time.Time  `json:"updated_at"`
}

// 3. AgencyCompanySubaccount binds one child login to one company owner.
type AgencyCompanySubaccount struct {
	ID                 int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	PublicID           string         `gorm:"type:varchar(26);not null;uniqueIndex" json:"public_id"`
	CompanyOwnerUserID int64          `gorm:"not null;index" json:"company_owner_user_id"`
	ChildUserID        int64          `gorm:"not null;uniqueIndex" json:"child_user_id"`
	DisplayName        string         `gorm:"type:varchar(120);not null" json:"display_name"`
	Status             string         `gorm:"type:varchar(32);not null;index" json:"status"`
	Permissions        datatypes.JSON `gorm:"type:jsonb" json:"permissions"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
}

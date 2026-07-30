/*
 * ismart account data models.
 * 1. Store the read-only identity returned by the POS login service.
 * 2. Preserve staff and client building permission arrays for account-level access decisions.
 * 3. Store encrypted POS relay access used by AJO-controlled payment queries.
 */
package model

import (
	"time"

	"gorm.io/datatypes"
)

// 1. UserIsmartAccount stores one linked ismart account snapshot.
type UserIsmartAccount struct {
	UserID                             int64          `gorm:"primaryKey" json:"user_id"`
	IsmartUserID                       int64          `gorm:"not null;uniqueIndex" json:"ismart_user_id"`
	Username                           string         `gorm:"type:varchar(120);not null;index" json:"username"`
	Email                              string         `gorm:"type:varchar(255);index" json:"email"`
	Phone                              string         `gorm:"type:varchar(64)" json:"phone"`
	IsStaff                            bool           `gorm:"not null;default:false;index" json:"is_staff"`
	Building                           datatypes.JSON `gorm:"type:jsonb;not null" json:"building"`
	StaffBuildingPermissions           datatypes.JSON `gorm:"type:jsonb;not null" json:"staff_building_permissions"`
	ClientBuildingPermissions          datatypes.JSON `gorm:"type:jsonb;not null" json:"client_building_permissions"`
	ClientBuildingFlatUnitsPermissions datatypes.JSON `gorm:"type:jsonb;not null" json:"client_building_flat_units_permissions"`
	RawMessage                         datatypes.JSON `gorm:"type:jsonb;not null" json:"raw_message"`
	ProfileSnapshot                    datatypes.JSON `gorm:"type:jsonb;not null;default:'{}'" json:"profile_snapshot"`
	ClientProfileSyncedAt              *time.Time     `json:"client_profile_synced_at,omitempty"`
	RelayTokenEncrypted                string         `gorm:"type:text" json:"-"`
	RelayTokenSyncedAt                 *time.Time     `json:"relay_token_synced_at,omitempty"`
	PasswordEncrypted                  string         `gorm:"type:text" json:"-"`
	CreatedAt                          time.Time      `json:"created_at"`
	UpdatedAt                          time.Time      `json:"updated_at"`
	User                               *User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

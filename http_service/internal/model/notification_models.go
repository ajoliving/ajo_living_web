/*
 * Notification data models.
 * 1. Store member-facing notification inbox entries.
 * 2. Keep order and chat related events queryable by user.
 */
package model

import "time"

// 1. Notification stores one user-facing inbox item.
type Notification struct {
	ID              int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	PublicID        string     `gorm:"type:varchar(26);not null;uniqueIndex" json:"public_id"`
	UserID          int64      `gorm:"not null;index:idx_notifications_user_read_created,priority:1" json:"user_id"`
	Category        string     `gorm:"type:varchar(64);not null;index" json:"category"`
	Title           string     `gorm:"type:varchar(200);not null" json:"title"`
	Body            string     `gorm:"type:varchar(1000);not null" json:"body"`
	RelatedType     string     `gorm:"type:varchar(64)" json:"related_type"`
	RelatedPublicID string     `gorm:"type:varchar(64)" json:"related_public_id"`
	IsRead          bool       `gorm:"not null;default:false;index:idx_notifications_user_read_created,priority:2" json:"is_read"`
	ReadAt          *time.Time `json:"read_at"`
	CreatedAt       time.Time  `gorm:"index:idx_notifications_user_read_created,priority:3" json:"created_at"`
}

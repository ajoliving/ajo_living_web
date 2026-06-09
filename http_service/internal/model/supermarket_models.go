/*
 * Supermarket offer member models.
 * 1. Store AJO-owned favorites for good-price products.
 * 2. Store AJO-owned price alert rules and sent alert events.
 * 3. Keep product price data owned by the deployed good-price service.
 */
package model

import "time"

// 1. SupermarketFavorite stores one saved good-price product for an AJO member.
type SupermarketFavorite struct {
	ID          int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      int64  `gorm:"not null;uniqueIndex:uk_supermarket_favorite_user_product,priority:1;index" json:"user_id"`
	ProductCode string `gorm:"type:varchar(64);not null;uniqueIndex:uk_supermarket_favorite_user_product,priority:2;index" json:"product_code"`
	ProductName string `gorm:"type:varchar(255);not null;default:''" json:"product_name"`
	Brand       string `gorm:"type:varchar(255);not null;default:''" json:"brand"`
	TimestampModel
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// 2. SupermarketPriceAlert stores one AJO member price alert rule.
type SupermarketPriceAlert struct {
	ID              int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID          int64      `gorm:"not null;uniqueIndex:uk_supermarket_alert_user_product,priority:1;index" json:"user_id"`
	ProductCode     string     `gorm:"type:varchar(64);not null;uniqueIndex:uk_supermarket_alert_user_product,priority:2;index" json:"product_code"`
	ProductName     string     `gorm:"type:varchar(255);not null;default:''" json:"product_name"`
	TargetPrice     *float64   `gorm:"type:numeric(12,2)" json:"target_price,omitempty"`
	PriceMode       string     `gorm:"type:varchar(16);not null;default:effective" json:"price_mode"`
	OfferRequired   bool       `gorm:"not null;default:false" json:"offer_required"`
	Enabled         bool       `gorm:"not null;default:true;index" json:"enabled"`
	LastTriggeredAt *time.Time `json:"last_triggered_at,omitempty"`
	TimestampModel
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// 3. SupermarketPriceAlertEvent stores one delivered or deduplicated alert event.
type SupermarketPriceAlertEvent struct {
	ID           int64                  `gorm:"primaryKey;autoIncrement" json:"id"`
	RuleID       int64                  `gorm:"not null;uniqueIndex:uk_supermarket_alert_event,priority:1;index" json:"rule_id"`
	UserID       int64                  `gorm:"not null;index" json:"user_id"`
	ProductCode  string                 `gorm:"type:varchar(64);not null" json:"product_code"`
	StoreCode    string                 `gorm:"type:varchar(64);not null;uniqueIndex:uk_supermarket_alert_event,priority:2" json:"store_code"`
	SnapshotDate string                 `gorm:"type:varchar(16);not null;uniqueIndex:uk_supermarket_alert_event,priority:3" json:"snapshot_date"`
	MatchedPrice float64                `gorm:"type:numeric(12,2);not null;uniqueIndex:uk_supermarket_alert_event,priority:4" json:"matched_price"`
	OfferText    string                 `gorm:"type:text;not null;default:''" json:"offer_text"`
	Message      string                 `gorm:"type:text;not null;default:''" json:"message"`
	NotifiedAt   *time.Time             `json:"notified_at,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	Rule         *SupermarketPriceAlert `gorm:"foreignKey:RuleID" json:"rule,omitempty"`
	User         *User                  `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

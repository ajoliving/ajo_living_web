/*
 * Order and handover data models.
 * 1. Define listing-linked order records for the marketplace flow.
 * 2. Persist status transition logs for auditability.
 */
package model

import "time"

// 1. Order stores one buyer-to-seller marketplace order.
type Order struct {
	ID             int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	PublicID       string     `gorm:"type:varchar(26);not null;uniqueIndex" json:"public_id"`
	BizModule      string     `gorm:"type:varchar(32);not null;index" json:"biz_module"`
	ListingID      int64      `gorm:"not null;index" json:"listing_id"`
	BuyerUserID    int64      `gorm:"not null;index" json:"buyer_user_id"`
	SellerUserID   int64      `gorm:"not null;index" json:"seller_user_id"`
	OrderStatus    string     `gorm:"type:varchar(32);not null;index" json:"order_status"`
	BuyerNote      string     `gorm:"type:varchar(500)" json:"buyer_note"`
	CancelReason   string     `gorm:"type:varchar(500)" json:"cancel_reason"`
	HandoverMethod string     `gorm:"type:varchar(64)" json:"handover_method"`
	ConfirmedAt    *time.Time `json:"confirmed_at"`
	CompletedAt    *time.Time `json:"completed_at"`
	CancelledAt    *time.Time `json:"cancelled_at"`
	TimestampModel
}

// 2. OrderLog stores a single order status transition event.
type OrderLog struct {
	ID             int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderID        int64     `gorm:"not null;index" json:"order_id"`
	ActionType     string    `gorm:"type:varchar(64);not null" json:"action_type"`
	FromStatus     string    `gorm:"type:varchar(32)" json:"from_status"`
	ToStatus       string    `gorm:"type:varchar(32)" json:"to_status"`
	OperatorUserID int64     `gorm:"not null;index" json:"operator_user_id"`
	Note           string    `gorm:"type:varchar(500)" json:"note"`
	CreatedAt      time.Time `json:"created_at"`
}

/*
 * Wallet and AJO Point data models.
 * 1. Store member point balances and immutable point transactions.
 * 2. Store operator-managed reward and display ads, engagement metrics, and member ad claim records.
 * 3. Keep all wallet audit fields explicit for production traceability.
 */
package model

import (
	"time"

	"gorm.io/datatypes"
)

// 1. WalletAccount stores one AJO Point account per member.
type WalletAccount struct {
	ID          int64 `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      int64 `gorm:"not null;uniqueIndex" json:"user_id"`
	Balance     int64 `gorm:"not null;default:0" json:"balance"`
	TotalEarned int64 `gorm:"not null;default:0" json:"total_earned"`
	TotalSpent  int64 `gorm:"not null;default:0" json:"total_spent"`
	Version     int64 `gorm:"not null;default:1" json:"version"`
	TimestampModel
}

// 2. WalletTransaction stores an immutable balance change record.
type WalletTransaction struct {
	ID             int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	PublicID       string    `gorm:"type:varchar(26);not null;uniqueIndex" json:"public_id"`
	UserID         int64     `gorm:"not null;index:idx_wallet_transactions_user_created,priority:1" json:"user_id"`
	Direction      string    `gorm:"type:varchar(16);not null;index" json:"direction"`
	Amount         int64     `gorm:"not null" json:"amount"`
	BalanceBefore  int64     `gorm:"not null" json:"balance_before"`
	BalanceAfter   int64     `gorm:"not null" json:"balance_after"`
	SourceType     string    `gorm:"type:varchar(64);not null;index" json:"source_type"`
	BizModule      string    `gorm:"type:varchar(64);not null;index" json:"biz_module"`
	ActionType     string    `gorm:"type:varchar(64);not null;index" json:"action_type"`
	ListingID      *int64    `gorm:"index" json:"listing_id"`
	RewardAdID     *int64    `gorm:"index" json:"reward_ad_id"`
	ClaimID        *int64    `gorm:"index" json:"claim_id"`
	IdempotencyKey string    `gorm:"type:varchar(160);not null;uniqueIndex" json:"idempotency_key"`
	OperatorUserID *int64    `gorm:"index" json:"operator_user_id"`
	Note           string    `gorm:"type:varchar(500)" json:"note"`
	CreatedAt      time.Time `gorm:"index:idx_wallet_transactions_user_created,priority:2" json:"created_at"`
}

// 3. RewardAd stores an operator-managed reward or display ad task.
type RewardAd struct {
	ID                int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	PublicID          string     `gorm:"type:varchar(26);not null;uniqueIndex" json:"public_id"`
	AdType            string     `gorm:"type:varchar(32);not null;default:'reward';index" json:"ad_type"`
	Title             string     `gorm:"type:varchar(160);not null" json:"title"`
	Summary           string     `gorm:"type:varchar(500)" json:"summary"`
	CoverURL          string     `gorm:"type:varchar(800)" json:"cover_url"`
	MediaURL          string     `gorm:"type:varchar(800)" json:"media_url"`
	MediaType         string     `gorm:"type:varchar(32);not null;default:'image';index" json:"media_type"`
	TargetURL         string     `gorm:"type:varchar(800)" json:"target_url"`
	DisplayChannel    string     `gorm:"type:varchar(64);not null;default:'';index:idx_reward_ads_display,priority:1" json:"display_channel"`
	DisplayPlacement  string     `gorm:"type:varchar(64);not null;default:'';index:idx_reward_ads_display,priority:2" json:"display_placement"`
	DisplayLayout     string     `gorm:"type:varchar(64);not null;default:'image_text'" json:"display_layout"`
	SortOrder         int        `gorm:"not null;default:0;index:idx_reward_ads_display,priority:3" json:"sort_order"`
	RewardPoints      int64      `gorm:"not null" json:"reward_points"`
	WatchSeconds      int        `gorm:"not null;default:30" json:"watch_seconds"`
	DailyUserLimit    int        `gorm:"not null;default:1" json:"daily_user_limit"`
	TotalBudget       int64      `gorm:"not null;default:0" json:"total_budget"`
	TotalGranted      int64      `gorm:"not null;default:0" json:"total_granted"`
	WatchCount        int64      `gorm:"not null;default:0" json:"watch_count"`
	TotalWatchSeconds int64      `gorm:"not null;default:0" json:"total_watch_seconds"`
	LinkClickCount    int64      `gorm:"not null;default:0" json:"link_click_count"`
	IsActive          bool       `gorm:"not null;default:true;index" json:"is_active"`
	StartsAt          *time.Time `gorm:"index" json:"starts_at"`
	EndsAt            *time.Time `gorm:"index" json:"ends_at"`
	CreatedBy         *int64     `gorm:"index" json:"created_by"`
	UpdatedBy         *int64     `gorm:"index" json:"updated_by"`
	TimestampModel
}

// 4. RewardAdClaim stores one member ad watch and claim attempt.
type RewardAdClaim struct {
	ID                  int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	PublicID            string     `gorm:"type:varchar(26);not null;uniqueIndex" json:"public_id"`
	RewardAdID          int64      `gorm:"not null;index:idx_reward_claim_user_ad,priority:2" json:"reward_ad_id"`
	UserID              int64      `gorm:"not null;index:idx_reward_claim_user_ad,priority:1" json:"user_id"`
	ClaimDate           string     `gorm:"type:varchar(10);not null;index:idx_reward_claim_daily,priority:1" json:"claim_date"`
	Status              string     `gorm:"type:varchar(32);not null;index" json:"status"`
	WatchStartedAt      time.Time  `json:"watch_started_at"`
	ClaimedAt           *time.Time `json:"claimed_at"`
	RewardPoints        int64      `gorm:"not null" json:"reward_points"`
	IPAddress           string     `gorm:"type:varchar(64)" json:"ip_address"`
	UserAgent           string     `gorm:"type:varchar(500)" json:"user_agent"`
	FailureReason       string     `gorm:"type:varchar(300)" json:"failure_reason"`
	WalletTransactionID *int64     `gorm:"index" json:"wallet_transaction_id"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
	RewardAd            *RewardAd  `gorm:"foreignKey:RewardAdID" json:"reward_ad,omitempty"`
}

// 5. WalletRechargeOrder stores one EasyLink-backed wallet recharge order.
type WalletRechargeOrder struct {
	ID                    int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	PublicID              string         `gorm:"type:varchar(26);not null;uniqueIndex" json:"public_id"`
	UserID                int64          `gorm:"not null;index" json:"user_id"`
	MchOrderNo            string         `gorm:"type:varchar(64);not null;uniqueIndex" json:"mch_order_no"`
	PayOrderID            string         `gorm:"type:varchar(80);index" json:"pay_order_id"`
	PayChannel            string         `gorm:"type:varchar(32);not null;index" json:"pay_channel"`
	PayRegion             string         `gorm:"type:varchar(16);not null;default:'HK';index" json:"pay_region"`
	PayDataType           string         `gorm:"type:varchar(32)" json:"pay_data_type"`
	PayData               string         `gorm:"type:text" json:"pay_data"`
	Currency              string         `gorm:"type:varchar(8);not null;default:'HKD'" json:"currency"`
	AmountCents           int64          `gorm:"not null" json:"amount_cents"`
	PointsAmount          int64          `gorm:"not null" json:"points_amount"`
	State                 string         `gorm:"type:varchar(32);not null;index" json:"state"`
	GatewayStateCode      int            `gorm:"not null;default:1" json:"gateway_state_code"`
	GatewayMessage        string         `gorm:"type:varchar(500)" json:"gateway_message"`
	RequestPayload        datatypes.JSON `gorm:"type:jsonb" json:"request_payload"`
	GatewayCreateRequest  datatypes.JSON `gorm:"type:jsonb" json:"gateway_create_request"`
	GatewayCreateResponse datatypes.JSON `gorm:"type:jsonb" json:"gateway_create_response"`
	GatewayQueryRequest   datatypes.JSON `gorm:"type:jsonb" json:"gateway_query_request"`
	GatewayQueryResponse  datatypes.JSON `gorm:"type:jsonb" json:"gateway_query_response"`
	GatewayNotifyPayload  datatypes.JSON `gorm:"type:jsonb" json:"gateway_notify_payload"`
	WalletTransactionID   *int64         `gorm:"index" json:"wallet_transaction_id"`
	ClientIP              string         `gorm:"type:varchar(64)" json:"client_ip"`
	UserAgent             string         `gorm:"type:varchar(500)" json:"user_agent"`
	ExpireTime            *time.Time     `gorm:"index" json:"expire_time"`
	PaidAt                *time.Time     `gorm:"index" json:"paid_at"`
	CreditedAt            *time.Time     `gorm:"index" json:"credited_at"`
	ClosedAt              *time.Time     `gorm:"index" json:"closed_at"`
	TimestampModel
}

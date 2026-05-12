/*
 * Wallet and AJO Point data models.
 * 1. Store member point balances and immutable point transactions.
 * 2. Store operator-managed reward ads, engagement metrics, and member ad claim records.
 * 3. Keep all wallet audit fields explicit for production traceability.
 */
package model

import "time"

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

// 3. RewardAd stores an operator-managed rewarded ad task.
type RewardAd struct {
	ID                int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	PublicID          string     `gorm:"type:varchar(26);not null;uniqueIndex" json:"public_id"`
	Title             string     `gorm:"type:varchar(160);not null" json:"title"`
	Summary           string     `gorm:"type:varchar(500)" json:"summary"`
	CoverURL          string     `gorm:"type:varchar(800)" json:"cover_url"`
	MediaURL          string     `gorm:"type:varchar(800)" json:"media_url"`
	MediaType         string     `gorm:"type:varchar(32);not null;default:'image';index" json:"media_type"`
	TargetURL         string     `gorm:"type:varchar(800)" json:"target_url"`
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

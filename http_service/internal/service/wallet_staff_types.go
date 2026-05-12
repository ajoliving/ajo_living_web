/*
 * Staff wallet API types.
 * 1. Define staff-visible wallet ledger payloads.
 * 2. Define operator rewarded ad management payloads.
 * 3. Keep staff wallet constants in one small file.
 */
package service

import "time"

const maxOperatorGrantPoints = int64(1000000)
const defaultRewardAdRetentionDays = 30
const maxRewardAdRetentionDays = 365
const rewardAdMediaTypeImage = "image"
const rewardAdMediaTypeVideo = "video"

// 1. StaffWalletUserResponse defines staff-visible wallet user metadata.
type StaffWalletUserResponse struct {
	UserID           string `json:"user_id"`
	PhoneCountryCode string `json:"phone_country_code"`
	PhoneNumber      string `json:"phone_number"`
	DisplayName      string `json:"display_name"`
}

// 2. StaffWalletTransactionResponse defines a staff-visible ledger row.
type StaffWalletTransactionResponse struct {
	TransactionID string                   `json:"transaction_id"`
	Direction     string                   `json:"direction"`
	Amount        int64                    `json:"amount"`
	BalanceBefore int64                    `json:"balance_before"`
	BalanceAfter  int64                    `json:"balance_after"`
	SourceType    string                   `json:"source_type"`
	BizModule     string                   `json:"biz_module"`
	ActionType    string                   `json:"action_type"`
	Note          string                   `json:"note"`
	TargetUser    StaffWalletUserResponse  `json:"target_user"`
	OperatorUser  *StaffWalletUserResponse `json:"operator_user,omitempty"`
	CreatedAt     string                   `json:"created_at"`
}

// 3. StaffWalletTransactionFilters defines staff ledger search filters.
type StaffWalletTransactionFilters struct {
	Page       int
	PageSize   int
	UserID     string
	Direction  string
	SourceType string
	BizModule  string
}

// 4. StaffWalletGrantResponse defines an operator grant result.
type StaffWalletGrantResponse struct {
	Charge     PointsChargeResponse    `json:"charge"`
	TargetUser StaffWalletUserResponse `json:"target_user"`
	Operator   StaffWalletUserResponse `json:"operator"`
}

// 5. StaffRewardAdResponse defines an operator-visible rewarded ad task.
type StaffRewardAdResponse struct {
	TaskID            string  `json:"task_id"`
	Title             string  `json:"title"`
	Summary           string  `json:"summary"`
	CoverURL          string  `json:"cover_url"`
	MediaURL          string  `json:"media_url"`
	MediaType         string  `json:"media_type"`
	TargetURL         string  `json:"target_url"`
	RewardPoints      int64   `json:"reward_points"`
	WatchSeconds      int     `json:"watch_seconds"`
	TotalBudget       int64   `json:"total_budget"`
	TotalGranted      int64   `json:"total_granted"`
	RemainingBudget   int64   `json:"remaining_budget"`
	WatchCount        int64   `json:"watch_count"`
	TotalWatchSeconds int64   `json:"total_watch_seconds"`
	LinkClickCount    int64   `json:"link_click_count"`
	LinkClickRate     float64 `json:"link_click_rate"`
	IsActive          bool    `json:"is_active"`
	StartsAt          string  `json:"starts_at,omitempty"`
	EndsAt            string  `json:"ends_at,omitempty"`
	CreatedAt         string  `json:"created_at"`
	UpdatedAt         string  `json:"updated_at"`
}

// 6. StaffRewardAdFilters defines rewarded ad search filters.
type StaffRewardAdFilters struct {
	Page     int
	PageSize int
	Keyword  string
	IsActive *bool
}

// 7. RewardAdCreateParams defines operator rewarded ad creation input.
type RewardAdCreateParams struct {
	Title         string
	Summary       string
	CoverURL      string
	MediaURL      string
	MediaType     string
	TargetURL     string
	RewardPoints  int64
	WatchSeconds  int
	TotalBudget   int64
	IsActive      bool
	StartsAt      *time.Time
	EndsAt        *time.Time
	RetentionDays int
}

// 8. RewardAdUpdateParams defines operator rewarded ad update input.
type RewardAdUpdateParams struct {
	Title         *string
	Summary       *string
	CoverURL      *string
	MediaURL      *string
	MediaType     *string
	TargetURL     *string
	RewardPoints  *int64
	WatchSeconds  *int
	TotalBudget   *int64
	IsActive      *bool
	StartsAtSet   bool
	StartsAt      *time.Time
	EndsAtSet     bool
	EndsAt        *time.Time
	RetentionDays *int
}

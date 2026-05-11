/*
 * AJO Point wallet business logic.
 * 1. Manage point account reads, credits, debits, and immutable transaction records.
 * 2. Provide rewarded ad watch and claim workflows with server-side timing checks.
 * 3. Expose shared listing charge helpers for secondhand and property modules.
 */
package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/model"
	"ajoliving_web/http_service/internal/utils"
)

const (
	WalletDirectionCredit = "credit"
	WalletDirectionDebit  = "debit"

	WalletSourceListingCharge = "listing_charge"
	WalletSourceRewardAd      = "reward_ad"
	WalletSourceOperatorGrant = "operator_grant"

	WalletActionPublish   = "publish"
	WalletActionEdit      = "edit"
	WalletActionRepublish = "republish"
	WalletActionAdReward  = "ad_reward"

	RewardClaimStatusStarted = "started"
	RewardClaimStatusClaimed = "claimed"
	RewardClaimStatusFailed  = "failed"

	WalletDailyAdRewardLimit = int64(1000)
)

// 1. WalletService handles AJO Point account and reward ad workflows.
type WalletService struct {
	runtime *Runtime
}

// 2. WalletAccountResponse defines the wallet account payload.
type WalletAccountResponse struct {
	Balance             int64 `json:"balance"`
	TotalEarned         int64 `json:"total_earned"`
	TotalSpent          int64 `json:"total_spent"`
	TodayAdRewardPoints int64 `json:"today_ad_reward_points"`
	DailyAdRewardLimit  int64 `json:"daily_ad_reward_limit"`
}

// 3. WalletTransactionResponse defines one transaction payload.
type WalletTransactionResponse struct {
	TransactionID string `json:"transaction_id"`
	Direction     string `json:"direction"`
	Amount        int64  `json:"amount"`
	BalanceBefore int64  `json:"balance_before"`
	BalanceAfter  int64  `json:"balance_after"`
	SourceType    string `json:"source_type"`
	BizModule     string `json:"biz_module"`
	ActionType    string `json:"action_type"`
	Note          string `json:"note"`
	CreatedAt     string `json:"created_at"`
}

// 4. WalletOverviewResponse defines the wallet page payload.
type WalletOverviewResponse struct {
	Account            WalletAccountResponse       `json:"account"`
	RecentTransactions []WalletTransactionResponse `json:"recent_transactions"`
	ChargeRules        []WalletChargeRuleResponse  `json:"charge_rules"`
	RechargeEnabled    bool                        `json:"recharge_enabled"`
}

// 5. WalletChargeRuleResponse defines one visible charge rule.
type WalletChargeRuleResponse struct {
	BizModule string `json:"biz_module"`
	Label     string `json:"label"`
	Publish   int64  `json:"publish"`
	Edit      int64  `json:"edit"`
	Republish int64  `json:"republish"`
}

// 6. PointsChargeResponse defines charge metadata attached to listing mutations.
type PointsChargeResponse struct {
	PointsCharged       int64  `json:"points_charged"`
	PointsBalanceAfter  int64  `json:"points_balance_after"`
	PointsTransactionID string `json:"points_transaction_id"`
}

// 7. RewardAdTaskResponse defines a member-visible rewarded ad task.
type RewardAdTaskResponse struct {
	TaskID          string `json:"task_id"`
	Title           string `json:"title"`
	Summary         string `json:"summary"`
	CoverURL        string `json:"cover_url"`
	MediaURL        string `json:"media_url"`
	MediaType       string `json:"media_type"`
	TargetURL       string `json:"target_url"`
	RewardPoints    int64  `json:"reward_points"`
	WatchSeconds    int    `json:"watch_seconds"`
	CanClaimToday   bool   `json:"can_claim_today"`
	ClaimedToday    bool   `json:"claimed_today"`
	RemainingBudget int64  `json:"remaining_budget"`
}

// 8. RewardAdSessionResponse defines a watch session payload.
type RewardAdSessionResponse struct {
	ClaimID      string `json:"claim_id"`
	TaskID       string `json:"task_id"`
	WatchSeconds int    `json:"watch_seconds"`
	StartedAt    string `json:"started_at"`
	AvailableAt  string `json:"available_at"`
	RewardPoints int64  `json:"reward_points"`
}

// 9. WalletSpendParams defines a point debit request.
type WalletSpendParams struct {
	UserID         int64
	Amount         int64
	BizModule      string
	ActionType     string
	ListingID      *int64
	IdempotencyKey string
	Note           string
}

// 10. WalletCreditParams defines a point credit request.
type WalletCreditParams struct {
	UserID         int64
	Amount         int64
	SourceType     string
	BizModule      string
	ActionType     string
	RewardAdID     *int64
	ClaimID        *int64
	OperatorUserID *int64
	IdempotencyKey string
	Note           string
}

// 11. NewWalletService creates a wallet service instance.
func NewWalletService(runtime *Runtime) *WalletService {
	return &WalletService{runtime: runtime}
}

// 12. ListingActionCost returns the configured cost for one module action.
func ListingActionCost(module string, action string) int64 {
	switch strings.TrimSpace(module) {
	case "secondhand":
		return 100
	case "property_sale":
		return 1000
	case "serviced_apartment":
		return 800
	default:
		return 0
	}
}

// 13. shouldChargeListingEdit returns whether an edit is billable.
func shouldChargeListingEdit(listing *model.Listing) bool {
	if listing == nil {
		return false
	}
	return listing.PublicationStatus != "draft" || listing.PublishedAt != nil
}

// 14. WalletChargeRules returns all member-facing listing charge rules.
func WalletChargeRules() []WalletChargeRuleResponse {
	return []WalletChargeRuleResponse{
		{BizModule: "secondhand", Label: "二手交易", Publish: 100, Edit: 100, Republish: 100},
		{BizModule: "property_sale", Label: "樓盤放售", Publish: 1000, Edit: 1000, Republish: 1000},
		{BizModule: "serviced_apartment", Label: "服務式住宅", Publish: 800, Edit: 800, Republish: 800},
	}
}

// 15. GetWalletOverview returns account, recent transactions, and charge rules.
func (s *WalletService) GetWalletOverview(ctx context.Context, userID int64) (*WalletOverviewResponse, error) {
	account, err := s.ensureAccount(ctx, s.runtime.DB, userID)
	if err != nil {
		return nil, err
	}

	todayReward, err := s.todayAdRewardPoints(ctx, s.runtime.DB, userID)
	if err != nil {
		return nil, err
	}
	transactions, _, err := s.ListTransactions(ctx, userID, 1, 8)
	if err != nil {
		return nil, err
	}

	return &WalletOverviewResponse{
		Account: WalletAccountResponse{
			Balance:             account.Balance,
			TotalEarned:         account.TotalEarned,
			TotalSpent:          account.TotalSpent,
			TodayAdRewardPoints: todayReward,
			DailyAdRewardLimit:  WalletDailyAdRewardLimit,
		},
		RecentTransactions: transactions,
		ChargeRules:        WalletChargeRules(),
		RechargeEnabled:    false,
	}, nil
}

// 16. ListTransactions returns paged wallet transactions.
func (s *WalletService) ListTransactions(ctx context.Context, userID int64, page int, pageSize int) ([]WalletTransactionResponse, *model.Pagination, error) {
	page, pageSize = normalizePagination(page, pageSize)

	query := s.runtime.DB.WithContext(ctx).Model(&model.WalletTransaction{}).Where("user_id = ?", userID)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, nil, errcode.New(errcode.CodeInternalError, "failed to count wallet transactions")
	}

	var transactions []model.WalletTransaction
	if err := query.Order("created_at desc, id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&transactions).Error; err != nil {
		return nil, nil, errcode.New(errcode.CodeInternalError, "failed to load wallet transactions")
	}

	items := make([]WalletTransactionResponse, 0, len(transactions))
	for _, transaction := range transactions {
		items = append(items, toWalletTransactionResponse(transaction))
	}

	return items, &model.Pagination{Page: page, PageSize: pageSize, Total: total}, nil
}

// 17. SpendPointsWithTx debits points inside an existing business transaction.
func (s *WalletService) SpendPointsWithTx(ctx context.Context, tx *gorm.DB, params WalletSpendParams) (*PointsChargeResponse, error) {
	if params.Amount <= 0 {
		return nil, errcode.New(errcode.CodeValidationError, "points amount must be positive")
	}
	if strings.TrimSpace(params.IdempotencyKey) == "" {
		params.IdempotencyKey = utils.NewPublicID()
	}
	if existing, ok, err := s.findTransactionByIdempotencyKey(ctx, tx, params.IdempotencyKey); err != nil {
		return nil, err
	} else if ok {
		return &PointsChargeResponse{
			PointsCharged:       existing.Amount,
			PointsBalanceAfter:  existing.BalanceAfter,
			PointsTransactionID: existing.PublicID,
		}, nil
	}

	account, err := s.ensureAccountForUpdate(ctx, tx, params.UserID)
	if err != nil {
		return nil, err
	}
	if account.Balance < params.Amount {
		return nil, errcode.New(errcode.CodePointsInsufficient, "insufficient AJO points")
	}

	before := account.Balance
	after := before - params.Amount
	transaction := model.WalletTransaction{
		PublicID:       utils.NewPublicID(),
		UserID:         params.UserID,
		Direction:      WalletDirectionDebit,
		Amount:         params.Amount,
		BalanceBefore:  before,
		BalanceAfter:   after,
		SourceType:     WalletSourceListingCharge,
		BizModule:      strings.TrimSpace(params.BizModule),
		ActionType:     strings.TrimSpace(params.ActionType),
		ListingID:      params.ListingID,
		IdempotencyKey: params.IdempotencyKey,
		Note:           strings.TrimSpace(params.Note),
		CreatedAt:      s.runtime.Now(),
	}
	if err := tx.WithContext(ctx).Create(&transaction).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to record wallet transaction")
	}

	if err := tx.WithContext(ctx).Model(&model.WalletAccount{}).Where("id = ?", account.ID).Updates(map[string]any{
		"balance":     after,
		"total_spent": account.TotalSpent + params.Amount,
		"version":     account.Version + 1,
	}).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to update wallet account")
	}

	return &PointsChargeResponse{
		PointsCharged:       params.Amount,
		PointsBalanceAfter:  after,
		PointsTransactionID: transaction.PublicID,
	}, nil
}

// 18. CreditPointsWithTx credits points inside an existing transaction.
func (s *WalletService) CreditPointsWithTx(ctx context.Context, tx *gorm.DB, params WalletCreditParams) (*PointsChargeResponse, error) {
	if params.Amount <= 0 {
		return nil, errcode.New(errcode.CodeValidationError, "points amount must be positive")
	}
	if strings.TrimSpace(params.IdempotencyKey) == "" {
		params.IdempotencyKey = utils.NewPublicID()
	}
	if existing, ok, err := s.findTransactionByIdempotencyKey(ctx, tx, params.IdempotencyKey); err != nil {
		return nil, err
	} else if ok {
		return &PointsChargeResponse{
			PointsCharged:       existing.Amount,
			PointsBalanceAfter:  existing.BalanceAfter,
			PointsTransactionID: existing.PublicID,
		}, nil
	}

	account, err := s.ensureAccountForUpdate(ctx, tx, params.UserID)
	if err != nil {
		return nil, err
	}

	before := account.Balance
	after := before + params.Amount
	transaction := model.WalletTransaction{
		PublicID:       utils.NewPublicID(),
		UserID:         params.UserID,
		Direction:      WalletDirectionCredit,
		Amount:         params.Amount,
		BalanceBefore:  before,
		BalanceAfter:   after,
		SourceType:     strings.TrimSpace(params.SourceType),
		BizModule:      strings.TrimSpace(params.BizModule),
		ActionType:     strings.TrimSpace(params.ActionType),
		RewardAdID:     params.RewardAdID,
		ClaimID:        params.ClaimID,
		OperatorUserID: params.OperatorUserID,
		IdempotencyKey: params.IdempotencyKey,
		Note:           strings.TrimSpace(params.Note),
		CreatedAt:      s.runtime.Now(),
	}
	if err := tx.WithContext(ctx).Create(&transaction).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to record wallet transaction")
	}

	if err := tx.WithContext(ctx).Model(&model.WalletAccount{}).Where("id = ?", account.ID).Updates(map[string]any{
		"balance":      after,
		"total_earned": account.TotalEarned + params.Amount,
		"version":      account.Version + 1,
	}).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to update wallet account")
	}

	return &PointsChargeResponse{
		PointsCharged:       params.Amount,
		PointsBalanceAfter:  after,
		PointsTransactionID: transaction.PublicID,
	}, nil
}

// 19. GrantOperatorPoints credits points from an operator adjustment.
func (s *WalletService) GrantOperatorPoints(ctx context.Context, operatorUserID int64, userID int64, amount int64, note string) (*PointsChargeResponse, error) {
	var result *PointsChargeResponse
	err := s.runtime.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		credit, err := s.CreditPointsWithTx(ctx, tx, WalletCreditParams{
			UserID:         userID,
			Amount:         amount,
			SourceType:     WalletSourceOperatorGrant,
			BizModule:      "wallet",
			ActionType:     "operator_grant",
			OperatorUserID: &operatorUserID,
			IdempotencyKey: fmt.Sprintf("operator_grant:%d:%d:%s", operatorUserID, userID, utils.NewPublicID()),
			Note:           note,
		})
		if err != nil {
			return err
		}
		result = credit
		return nil
	})
	return result, err
}

// 20. ListRewardAdTasks returns active ad tasks and member claim state.
func (s *WalletService) ListRewardAdTasks(ctx context.Context, userID int64) ([]RewardAdTaskResponse, error) {
	now := s.runtime.Now()
	var ads []model.RewardAd
	if err := s.runtime.DB.WithContext(ctx).
		Where("is_active = ?", true).
		Where("(starts_at IS NULL OR starts_at <= ?)", now).
		Where("(ends_at IS NULL OR ends_at >= ?)", now).
		Order("updated_at desc, id desc").
		Find(&ads).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load reward ads")
	}

	claimDate := walletDate(now)
	todayReward, err := s.todayAdRewardPoints(ctx, s.runtime.DB, userID)
	if err != nil {
		return nil, err
	}

	items := make([]RewardAdTaskResponse, 0, len(ads))
	for _, ad := range ads {
		claimed, err := s.hasClaimedRewardAdToday(ctx, s.runtime.DB, userID, ad.ID, claimDate)
		if err != nil {
			return nil, err
		}
		remainingBudget := ad.TotalBudget - ad.TotalGranted
		if ad.TotalBudget <= 0 {
			remainingBudget = ad.RewardPoints
		}
		items = append(items, RewardAdTaskResponse{
			TaskID:          ad.PublicID,
			Title:           ad.Title,
			Summary:         ad.Summary,
			CoverURL:        ad.CoverURL,
			MediaURL:        ad.MediaURL,
			MediaType:       normalizedRewardAdMediaType(ad.MediaType),
			TargetURL:       ad.TargetURL,
			RewardPoints:    ad.RewardPoints,
			WatchSeconds:    normalizedWatchSeconds(ad.WatchSeconds),
			ClaimedToday:    claimed,
			CanClaimToday:   !claimed && todayReward+ad.RewardPoints <= WalletDailyAdRewardLimit && remainingBudget >= ad.RewardPoints,
			RemainingBudget: remainingBudget,
		})
	}

	return items, nil
}

// 21. StartRewardAd starts a rewarded ad watch session.
func (s *WalletService) StartRewardAd(ctx context.Context, userID int64, taskPublicID string, ipAddress string, userAgent string) (*RewardAdSessionResponse, error) {
	ad, err := s.loadActiveRewardAd(ctx, s.runtime.DB, taskPublicID)
	if err != nil {
		return nil, err
	}
	now := s.runtime.Now()
	claimDate := walletDate(now)
	if claimed, err := s.hasClaimedRewardAdToday(ctx, s.runtime.DB, userID, ad.ID, claimDate); err != nil {
		return nil, err
	} else if claimed {
		return nil, errcode.New(errcode.CodeWalletActionUnavailable, "reward already claimed today")
	}

	claim := model.RewardAdClaim{
		PublicID:       utils.NewPublicID(),
		RewardAdID:     ad.ID,
		UserID:         userID,
		ClaimDate:      claimDate,
		Status:         RewardClaimStatusStarted,
		WatchStartedAt: now,
		RewardPoints:   ad.RewardPoints,
		IPAddress:      strings.TrimSpace(ipAddress),
		UserAgent:      strings.TrimSpace(userAgent),
	}
	if err := s.runtime.DB.WithContext(ctx).Create(&claim).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to start reward ad")
	}

	watchSeconds := normalizedWatchSeconds(ad.WatchSeconds)
	return &RewardAdSessionResponse{
		ClaimID:      claim.PublicID,
		TaskID:       ad.PublicID,
		WatchSeconds: watchSeconds,
		StartedAt:    now.Format(time.RFC3339),
		AvailableAt:  now.Add(time.Duration(watchSeconds) * time.Second).Format(time.RFC3339),
		RewardPoints: ad.RewardPoints,
	}, nil
}

// 22. ClaimRewardAd grants points for a completed rewarded ad session.
func (s *WalletService) ClaimRewardAd(ctx context.Context, userID int64, taskPublicID string, claimPublicID string, ipAddress string, userAgent string) (*PointsChargeResponse, error) {
	var result *PointsChargeResponse
	err := s.runtime.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ad, err := s.loadActiveRewardAd(ctx, tx, taskPublicID)
		if err != nil {
			return err
		}
		var claim model.RewardAdClaim
		if err := tx.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("public_id = ? AND user_id = ? AND reward_ad_id = ?", claimPublicID, userID, ad.ID).
			First(&claim).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errcode.New(errcode.CodeNotFound, "reward claim not found")
			}
			return errcode.New(errcode.CodeInternalError, "failed to load reward claim")
		}
		if claim.Status == RewardClaimStatusClaimed {
			return errcode.New(errcode.CodeWalletActionUnavailable, "reward already claimed")
		}
		now := s.runtime.Now()
		if now.Before(claim.WatchStartedAt.Add(time.Duration(normalizedWatchSeconds(ad.WatchSeconds)) * time.Second)) {
			return errcode.New(errcode.CodeWalletActionUnavailable, "reward ad watch time is not complete")
		}
		todayReward, err := s.todayAdRewardPoints(ctx, tx, userID)
		if err != nil {
			return err
		}
		if todayReward+ad.RewardPoints > WalletDailyAdRewardLimit {
			return errcode.New(errcode.CodeWalletActionUnavailable, "daily ad reward limit reached")
		}
		if ad.TotalBudget > 0 && ad.TotalGranted+ad.RewardPoints > ad.TotalBudget {
			return errcode.New(errcode.CodeWalletActionUnavailable, "reward ad budget reached")
		}
		if claimed, err := s.hasClaimedRewardAdToday(ctx, tx, userID, ad.ID, walletDate(now)); err != nil {
			return err
		} else if claimed {
			return errcode.New(errcode.CodeWalletActionUnavailable, "reward already claimed today")
		}

		claim.Status = RewardClaimStatusClaimed
		claim.ClaimedAt = &now
		claim.IPAddress = strings.TrimSpace(ipAddress)
		claim.UserAgent = strings.TrimSpace(userAgent)
		if err := tx.WithContext(ctx).Save(&claim).Error; err != nil {
			return errcode.New(errcode.CodeInternalError, "failed to update reward claim")
		}

		credit, err := s.CreditPointsWithTx(ctx, tx, WalletCreditParams{
			UserID:         userID,
			Amount:         ad.RewardPoints,
			SourceType:     WalletSourceRewardAd,
			BizModule:      "wallet",
			ActionType:     WalletActionAdReward,
			RewardAdID:     &ad.ID,
			ClaimID:        &claim.ID,
			IdempotencyKey: fmt.Sprintf("reward_ad:%d:%d:%s", userID, ad.ID, claim.PublicID),
			Note:           "rewarded ad completed",
		})
		if err != nil {
			return err
		}
		result = credit

		if err := tx.WithContext(ctx).Model(&model.RewardAd{}).Where("id = ?", ad.ID).Update("total_granted", ad.TotalGranted+ad.RewardPoints).Error; err != nil {
			return errcode.New(errcode.CodeInternalError, "failed to update reward ad budget")
		}
		var transaction model.WalletTransaction
		if err := tx.WithContext(ctx).Where("public_id = ?", credit.PointsTransactionID).First(&transaction).Error; err != nil {
			return errcode.New(errcode.CodeInternalError, "failed to load reward transaction")
		}
		if err := tx.WithContext(ctx).Model(&model.RewardAdClaim{}).Where("id = ?", claim.ID).Update("wallet_transaction_id", transaction.ID).Error; err != nil {
			return errcode.New(errcode.CodeInternalError, "failed to bind reward transaction")
		}
		return nil
	})
	return result, err
}

// 23. ensureAccount creates a wallet account when missing.
func (s *WalletService) ensureAccount(ctx context.Context, tx *gorm.DB, userID int64) (*model.WalletAccount, error) {
	var account model.WalletAccount
	if err := tx.WithContext(ctx).Where("user_id = ?", userID).First(&account).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.New(errcode.CodeInternalError, "failed to load wallet account")
		}
		account = model.WalletAccount{UserID: userID, Balance: 0, TotalEarned: 0, TotalSpent: 0, Version: 1}
		if err := tx.WithContext(ctx).Create(&account).Error; err != nil {
			return nil, errcode.New(errcode.CodeInternalError, "failed to create wallet account")
		}
	}
	return &account, nil
}

// 24. ensureAccountForUpdate creates and locks a wallet account.
func (s *WalletService) ensureAccountForUpdate(ctx context.Context, tx *gorm.DB, userID int64) (*model.WalletAccount, error) {
	if _, err := s.ensureAccount(ctx, tx, userID); err != nil {
		return nil, err
	}
	var account model.WalletAccount
	if err := tx.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ?", userID).First(&account).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to lock wallet account")
	}
	return &account, nil
}

// 25. findTransactionByIdempotencyKey loads a previous wallet transaction.
func (s *WalletService) findTransactionByIdempotencyKey(ctx context.Context, tx *gorm.DB, key string) (*model.WalletTransaction, bool, error) {
	var transaction model.WalletTransaction
	if err := tx.WithContext(ctx).Where("idempotency_key = ?", strings.TrimSpace(key)).First(&transaction).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, nil
		}
		return nil, false, errcode.New(errcode.CodeInternalError, "failed to load wallet idempotency record")
	}
	return &transaction, true, nil
}

// 26. todayAdRewardPoints sums claimed ad rewards for the current wallet day.
func (s *WalletService) todayAdRewardPoints(ctx context.Context, tx *gorm.DB, userID int64) (int64, error) {
	var total int64
	if err := tx.WithContext(ctx).Model(&model.WalletTransaction{}).
		Where("user_id = ? AND direction = ? AND source_type = ? AND action_type = ? AND created_at >= ?",
			userID,
			WalletDirectionCredit,
			WalletSourceRewardAd,
			WalletActionAdReward,
			walletDayStart(s.runtime.Now()),
		).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error; err != nil {
		return 0, errcode.New(errcode.CodeInternalError, "failed to sum ad reward points")
	}
	return total, nil
}

// 27. hasClaimedRewardAdToday checks completed claims for one ad and date.
func (s *WalletService) hasClaimedRewardAdToday(ctx context.Context, tx *gorm.DB, userID int64, adID int64, claimDate string) (bool, error) {
	var count int64
	if err := tx.WithContext(ctx).Model(&model.RewardAdClaim{}).
		Where("user_id = ? AND reward_ad_id = ? AND claim_date = ? AND status = ?", userID, adID, claimDate, RewardClaimStatusClaimed).
		Count(&count).Error; err != nil {
		return false, errcode.New(errcode.CodeInternalError, "failed to load reward claim state")
	}
	return count > 0, nil
}

// 28. loadActiveRewardAd loads a currently claimable reward ad.
func (s *WalletService) loadActiveRewardAd(ctx context.Context, tx *gorm.DB, publicID string) (*model.RewardAd, error) {
	now := s.runtime.Now()
	var ad model.RewardAd
	if err := tx.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("public_id = ? AND is_active = ?", strings.TrimSpace(publicID), true).
		Where("(starts_at IS NULL OR starts_at <= ?)", now).
		Where("(ends_at IS NULL OR ends_at >= ?)", now).
		First(&ad).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.New(errcode.CodeNotFound, "reward ad not found")
		}
		return nil, errcode.New(errcode.CodeInternalError, "failed to load reward ad")
	}
	return &ad, nil
}

// 29. toWalletTransactionResponse maps a transaction model to API shape.
func toWalletTransactionResponse(transaction model.WalletTransaction) WalletTransactionResponse {
	return WalletTransactionResponse{
		TransactionID: transaction.PublicID,
		Direction:     transaction.Direction,
		Amount:        transaction.Amount,
		BalanceBefore: transaction.BalanceBefore,
		BalanceAfter:  transaction.BalanceAfter,
		SourceType:    transaction.SourceType,
		BizModule:     transaction.BizModule,
		ActionType:    transaction.ActionType,
		Note:          transaction.Note,
		CreatedAt:     transaction.CreatedAt.Format(time.RFC3339),
	}
}

// 30. walletDate returns the server-side wallet date string.
func walletDate(value time.Time) string {
	return value.Format("2006-01-02")
}

// 31. walletDayStart returns the start of the current wallet day.
func walletDayStart(value time.Time) time.Time {
	year, month, day := value.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, value.Location())
}

// 32. normalizedWatchSeconds enforces the minimum watch duration.
func normalizedWatchSeconds(value int) int {
	if value < 1 {
		return 1
	}
	return value
}

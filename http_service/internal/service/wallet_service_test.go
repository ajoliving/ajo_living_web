/*
 * Wallet service tests.
 * 1. Validate AJO Point debit, insufficient balance, and transaction records.
 * 2. Validate rewarded ad timing, duplicate claim, and balance updates.
 */
package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"gorm.io/gorm"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/model"
	"ajoliving_web/http_service/internal/utils"
)

// 1. TestWalletSpendPoints validates debit and insufficient balance behavior.
func TestWalletSpendPoints(t *testing.T) {
	runtime := newTestRuntime(t)
	walletService := NewWalletService(runtime)
	communityA, _ := mustGetCommunities(t, runtime)
	user := mustCreateUser(t, runtime, "+852", "93000001", &communityA.ID)
	mustGrantPoints(t, runtime, user.ID, 100)

	var charge *PointsChargeResponse
	err := runtime.DB.Transaction(func(tx *gorm.DB) error {
		var spendErr error
		charge, spendErr = walletService.SpendPointsWithTx(context.Background(), tx, WalletSpendParams{
			UserID:         user.ID,
			Amount:         80,
			BizModule:      "secondhand",
			ActionType:     WalletActionEdit,
			IdempotencyKey: "test-spend-80",
			Note:           "unit test spend",
		})
		return spendErr
	})
	if err != nil {
		t.Fatalf("spend points: %v", err)
	}
	if charge.PointsCharged != 80 || charge.PointsBalanceAfter != 20 {
		t.Fatalf("unexpected charge result: %#v", charge)
	}

	err = runtime.DB.Transaction(func(tx *gorm.DB) error {
		_, spendErr := walletService.SpendPointsWithTx(context.Background(), tx, WalletSpendParams{
			UserID:         user.ID,
			Amount:         30,
			BizModule:      "secondhand",
			ActionType:     WalletActionEdit,
			IdempotencyKey: "test-spend-30",
		})
		return spendErr
	})
	var appErr *errcode.AppError
	if !errors.As(err, &appErr) || appErr.Code != errcode.CodePointsInsufficient {
		t.Fatalf("expected insufficient points, got %#v", err)
	}
}

// 2. TestRewardAdClaimRequiresThirtySeconds validates rewarded ad claim safeguards.
func TestRewardAdClaimRequiresThirtySeconds(t *testing.T) {
	runtime := newTestRuntime(t)
	walletService := NewWalletService(runtime)
	communityA, _ := mustGetCommunities(t, runtime)
	user := mustCreateUser(t, runtime, "+852", "93000002", &communityA.ID)
	ad := model.RewardAd{
		PublicID:     utils.NewPublicID(),
		Title:        "AJO Living Reward",
		Summary:      "Watch to earn points.",
		RewardPoints: 50,
		WatchSeconds: 30,
		TotalBudget:  100,
		IsActive:     true,
	}
	if err := runtime.DB.Create(&ad).Error; err != nil {
		t.Fatalf("create reward ad: %v", err)
	}

	session, err := walletService.StartRewardAd(context.Background(), user.ID, ad.PublicID, "127.0.0.1", "unit-test")
	if err != nil {
		t.Fatalf("start reward ad: %v", err)
	}
	_, err = walletService.ClaimRewardAd(context.Background(), user.ID, ad.PublicID, session.ClaimID, "127.0.0.1", "unit-test")
	var appErr *errcode.AppError
	if !errors.As(err, &appErr) || appErr.Code != errcode.CodeWalletActionUnavailable {
		t.Fatalf("expected early claim rejection, got %#v", err)
	}

	runtime.Now = func() time.Time { return fixedNow.Add(31 * time.Second) }
	charge, err := walletService.ClaimRewardAd(context.Background(), user.ID, ad.PublicID, session.ClaimID, "127.0.0.1", "unit-test")
	if err != nil {
		t.Fatalf("claim reward ad: %v", err)
	}
	if charge.PointsCharged != 50 || charge.PointsBalanceAfter != 50 {
		t.Fatalf("unexpected reward charge: %#v", charge)
	}

	_, err = walletService.ClaimRewardAd(context.Background(), user.ID, ad.PublicID, session.ClaimID, "127.0.0.1", "unit-test")
	if !errors.As(err, &appErr) || appErr.Code != errcode.CodeWalletActionUnavailable {
		t.Fatalf("expected duplicate claim rejection, got %#v", err)
	}
}

// 3. TestRewardAdDailyLimitAllowsOneThousandPoints validates the daily ad reward cap.
func TestRewardAdDailyLimitAllowsOneThousandPoints(t *testing.T) {
	runtime := newTestRuntime(t)
	walletService := NewWalletService(runtime)
	communityA, _ := mustGetCommunities(t, runtime)
	user := mustCreateUser(t, runtime, "+852", "93000006", &communityA.ID)

	firstAd := mustCreateRewardAd(t, runtime, 500, 1000)
	secondAd := mustCreateRewardAd(t, runtime, 500, 1000)
	thirdAd := mustCreateRewardAd(t, runtime, 1, 1000)

	mustClaimRewardAd(t, runtime, walletService, user.ID, firstAd.PublicID)
	secondCharge := mustClaimRewardAd(t, runtime, walletService, user.ID, secondAd.PublicID)
	if secondCharge.PointsBalanceAfter != WalletDailyAdRewardLimit {
		t.Fatalf("expected balance to reach daily cap, got %#v", secondCharge)
	}

	session, err := walletService.StartRewardAd(context.Background(), user.ID, thirdAd.PublicID, "127.0.0.1", "unit-test")
	if err != nil {
		t.Fatalf("start third reward ad: %v", err)
	}
	runtime.Now = func() time.Time { return fixedNow.Add(3*time.Minute + 1*time.Second) }
	_, err = walletService.ClaimRewardAd(context.Background(), user.ID, thirdAd.PublicID, session.ClaimID, "127.0.0.1", "unit-test")
	var appErr *errcode.AppError
	if !errors.As(err, &appErr) || appErr.Code != errcode.CodeWalletActionUnavailable {
		t.Fatalf("expected daily reward cap rejection, got %#v", err)
	}

	overview, err := walletService.GetWalletOverview(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("get wallet overview: %v", err)
	}
	if overview.Account.TodayAdRewardPoints != WalletDailyAdRewardLimit || overview.Account.DailyAdRewardLimit != WalletDailyAdRewardLimit {
		t.Fatalf("unexpected daily reward overview: %#v", overview.Account)
	}
}

// 4. TestStaffWalletGrantAndLedger validates operator grants and staff ledger search.
func TestStaffWalletGrantAndLedger(t *testing.T) {
	runtime := newTestRuntime(t)
	walletService := NewWalletService(runtime)
	communityA, _ := mustGetCommunities(t, runtime)
	operator := mustCreateUser(t, runtime, "+852", "93000003", &communityA.ID)
	target := mustCreateUser(t, runtime, "+852", "93000004", &communityA.ID)

	grant, err := walletService.GrantOperatorPointsByUserID(context.Background(), operator.ID, target.PublicID, 300, "Launch credit")
	if err != nil {
		t.Fatalf("grant operator points: %v", err)
	}
	if grant.Charge.PointsCharged != 300 || grant.Charge.PointsBalanceAfter != 300 {
		t.Fatalf("unexpected grant result: %#v", grant)
	}
	if grant.TargetUser.UserID != target.PublicID || grant.Operator.UserID != operator.PublicID {
		t.Fatalf("unexpected grant users: %#v", grant)
	}

	rows, pagination, err := walletService.ListStaffWalletTransactions(context.Background(), StaffWalletTransactionFilters{
		UserID:     target.PublicID,
		SourceType: WalletSourceOperatorGrant,
	})
	if err != nil {
		t.Fatalf("list staff wallet transactions: %v", err)
	}
	if pagination.Total != 1 || len(rows) != 1 {
		t.Fatalf("unexpected ledger rows: total=%d len=%d", pagination.Total, len(rows))
	}
	if rows[0].Amount != 300 || rows[0].TargetUser.UserID != target.PublicID {
		t.Fatalf("unexpected ledger row: %#v", rows[0])
	}
	if rows[0].OperatorUser == nil || rows[0].OperatorUser.UserID != operator.PublicID {
		t.Fatalf("missing operator metadata: %#v", rows[0])
	}
}

// 5. TestStaffRewardAdManagement validates operator reward ad create and update.
func TestStaffRewardAdManagement(t *testing.T) {
	runtime := newTestRuntime(t)
	walletService := NewWalletService(runtime)
	communityA, _ := mustGetCommunities(t, runtime)
	operator := mustCreateUser(t, runtime, "+852", "93000005", &communityA.ID)

	created, err := walletService.CreateStaffRewardAd(context.Background(), operator.ID, RewardAdCreateParams{
		Title:         "AJO Living Reward",
		Summary:       "Watch official video.",
		MediaURL:      "https://example.com/reward.mp4",
		MediaType:     rewardAdMediaTypeVideo,
		RewardPoints:  60,
		WatchSeconds:  15,
		IsActive:      true,
		RetentionDays: 7,
	})
	if err != nil {
		t.Fatalf("create staff reward ad: %v", err)
	}
	if created.WatchSeconds != 15 || created.MediaType != rewardAdMediaTypeVideo || created.TotalBudget != 0 || created.RemainingBudget != 0 || created.EndsAt == "" {
		t.Fatalf("unexpected created ad: %#v", created)
	}

	isActive := false
	rewardPoints := int64(80)
	updated, err := walletService.UpdateStaffRewardAd(context.Background(), operator.ID, created.TaskID, RewardAdUpdateParams{
		IsActive:     &isActive,
		RewardPoints: &rewardPoints,
	})
	if err != nil {
		t.Fatalf("update staff reward ad: %v", err)
	}
	if updated.IsActive || updated.RewardPoints != 80 {
		t.Fatalf("unexpected updated ad: %#v", updated)
	}

	items, pagination, err := walletService.ListStaffRewardAds(context.Background(), StaffRewardAdFilters{
		IsActive: &isActive,
	})
	if err != nil {
		t.Fatalf("list staff reward ads: %v", err)
	}
	if pagination.Total != 1 || len(items) != 1 || items[0].TaskID != created.TaskID {
		t.Fatalf("unexpected ad list: total=%d items=%#v", pagination.Total, items)
	}
}

// 6. mustCreateRewardAd creates an active rewarded ad fixture.
func mustCreateRewardAd(t *testing.T, runtime *Runtime, rewardPoints int64, totalBudget int64) model.RewardAd {
	t.Helper()

	ad := model.RewardAd{
		PublicID:     utils.NewPublicID(),
		Title:        "AJO Living Reward",
		Summary:      "Watch to earn points.",
		RewardPoints: rewardPoints,
		WatchSeconds: 30,
		TotalBudget:  totalBudget,
		IsActive:     true,
	}
	if err := runtime.DB.Create(&ad).Error; err != nil {
		t.Fatalf("create reward ad: %v", err)
	}

	return ad
}

// 7. mustClaimRewardAd completes one rewarded ad claim.
func mustClaimRewardAd(t *testing.T, runtime *Runtime, walletService *WalletService, userID int64, taskID string) *PointsChargeResponse {
	t.Helper()

	startedAt := runtime.Now()
	session, err := walletService.StartRewardAd(context.Background(), userID, taskID, "127.0.0.1", "unit-test")
	if err != nil {
		t.Fatalf("start reward ad: %v", err)
	}
	runtime.Now = func() time.Time { return startedAt.Add(31 * time.Second) }
	charge, err := walletService.ClaimRewardAd(context.Background(), userID, taskID, session.ClaimID, "127.0.0.1", "unit-test")
	if err != nil {
		t.Fatalf("claim reward ad: %v", err)
	}

	return charge
}

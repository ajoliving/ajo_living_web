/*
 * Staff rewarded ad business logic.
 * 1. List operator-managed rewarded ad tasks.
 * 2. Create and update rewarded ad tasks with media and retention validation.
 * 3. Map rewarded ad records into staff API responses.
 */
package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/model"
	"ajoliving_web/http_service/internal/utils"
)

// 1. ListStaffRewardAds returns rewarded ad tasks for operators.
func (s *WalletService) ListStaffRewardAds(ctx context.Context, filters StaffRewardAdFilters) ([]StaffRewardAdResponse, *model.Pagination, error) {
	page, pageSize := normalizePagination(filters.Page, filters.PageSize)
	query := s.runtime.DB.WithContext(ctx).Model(&model.RewardAd{})

	if keyword := strings.TrimSpace(filters.Keyword); keyword != "" {
		pattern := "%" + keyword + "%"
		query = query.Where("title LIKE ? OR summary LIKE ?", pattern, pattern)
	}
	if filters.IsActive != nil {
		query = query.Where("is_active = ?", *filters.IsActive)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, nil, errcode.New(errcode.CodeInternalError, "failed to count reward ads")
	}

	var ads []model.RewardAd
	if err := query.Order("updated_at desc, id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&ads).Error; err != nil {
		return nil, nil, errcode.New(errcode.CodeInternalError, "failed to load reward ads")
	}

	items := make([]StaffRewardAdResponse, 0, len(ads))
	for _, ad := range ads {
		items = append(items, toStaffRewardAdResponse(ad))
	}
	return items, &model.Pagination{Page: page, PageSize: pageSize, Total: total}, nil
}

// 2. GetStaffRewardAd returns one operator-managed rewarded ad task.
func (s *WalletService) GetStaffRewardAd(ctx context.Context, taskPublicID string) (*StaffRewardAdResponse, error) {
	ad, err := s.loadStaffRewardAd(ctx, taskPublicID)
	if err != nil {
		return nil, err
	}

	result := toStaffRewardAdResponse(*ad)
	return &result, nil
}

// 3. CreateStaffRewardAd creates an operator-managed rewarded ad task.
func (s *WalletService) CreateStaffRewardAd(ctx context.Context, operatorUserID int64, params RewardAdCreateParams) (*StaffRewardAdResponse, error) {
	if err := validateRewardAdCreate(params); err != nil {
		return nil, err
	}

	now := s.runtime.Now()
	startsAt := params.StartsAt
	endsAt := params.EndsAt
	if startsAt == nil {
		startsAt = &now
	}
	if params.RetentionDays > 0 {
		calculatedEnd := now.AddDate(0, 0, params.RetentionDays)
		endsAt = &calculatedEnd
	}
	ad := model.RewardAd{
		PublicID:       utils.NewPublicID(),
		Title:          strings.TrimSpace(params.Title),
		Summary:        strings.TrimSpace(params.Summary),
		CoverURL:       "",
		MediaURL:       strings.TrimSpace(params.MediaURL),
		MediaType:      normalizedRewardAdMediaType(params.MediaType),
		TargetURL:      strings.TrimSpace(params.TargetURL),
		RewardPoints:   params.RewardPoints,
		WatchSeconds:   normalizedWatchSeconds(params.WatchSeconds),
		DailyUserLimit: 1,
		TotalBudget:    0,
		IsActive:       params.IsActive,
		StartsAt:       startsAt,
		EndsAt:         endsAt,
		CreatedBy:      &operatorUserID,
		UpdatedBy:      &operatorUserID,
	}
	ad.CreatedAt = now
	ad.UpdatedAt = now
	if err := s.runtime.DB.WithContext(ctx).Create(&ad).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to create reward ad")
	}

	result := toStaffRewardAdResponse(ad)
	return &result, nil
}

// 4. UpdateStaffRewardAd updates an operator-managed rewarded ad task.
func (s *WalletService) UpdateStaffRewardAd(ctx context.Context, operatorUserID int64, taskPublicID string, params RewardAdUpdateParams) (*StaffRewardAdResponse, error) {
	ad, err := s.loadStaffRewardAd(ctx, taskPublicID)
	if err != nil {
		return nil, err
	}
	if err := applyRewardAdUpdates(ad, params, operatorUserID, s.runtime.Now()); err != nil {
		return nil, err
	}
	if err := validateRewardAdState(*ad); err != nil {
		return nil, err
	}
	if err := s.runtime.DB.WithContext(ctx).Save(ad).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to update reward ad")
	}

	result := toStaffRewardAdResponse(*ad)
	return &result, nil
}

// 5. loadStaffRewardAd loads a reward ad by public id.
func (s *WalletService) loadStaffRewardAd(ctx context.Context, taskPublicID string) (*model.RewardAd, error) {
	var ad model.RewardAd
	if err := s.runtime.DB.WithContext(ctx).Where("public_id = ?", strings.TrimSpace(taskPublicID)).First(&ad).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.New(errcode.CodeNotFound, "reward ad not found")
		}
		return nil, errcode.New(errcode.CodeInternalError, "failed to load reward ad")
	}
	return &ad, nil
}

// 6. validateRewardAdCreate validates rewarded ad creation input.
func validateRewardAdCreate(params RewardAdCreateParams) error {
	ad := model.RewardAd{
		Title:        strings.TrimSpace(params.Title),
		Summary:      strings.TrimSpace(params.Summary),
		MediaURL:     strings.TrimSpace(params.MediaURL),
		MediaType:    normalizedRewardAdMediaType(params.MediaType),
		RewardPoints: params.RewardPoints,
		WatchSeconds: normalizedWatchSeconds(params.WatchSeconds),
		StartsAt:     params.StartsAt,
		EndsAt:       params.EndsAt,
	}
	if params.RetentionDays <= 0 {
		return errcode.New(errcode.CodeValidationError, "reward ad retention days must be positive")
	}
	if params.RetentionDays > maxRewardAdRetentionDays {
		return errcode.New(errcode.CodeValidationError, "reward ad retention days is too long")
	}
	return validateRewardAdState(ad)
}

// 7. validateRewardAdState validates a rewarded ad before persistence.
func validateRewardAdState(ad model.RewardAd) error {
	if strings.TrimSpace(ad.Title) == "" {
		return errcode.New(errcode.CodeValidationError, "reward ad title is required")
	}
	if strings.TrimSpace(ad.Summary) == "" {
		return errcode.New(errcode.CodeValidationError, "reward ad summary is required")
	}
	if strings.TrimSpace(ad.MediaURL) == "" {
		return errcode.New(errcode.CodeValidationError, "reward ad media is required")
	}
	if ad.RewardPoints <= 0 {
		return errcode.New(errcode.CodeValidationError, "reward points must be positive")
	}
	if normalizedWatchSeconds(ad.WatchSeconds) < 1 {
		return errcode.New(errcode.CodeValidationError, "watch seconds must be positive")
	}
	if normalizedRewardAdMediaType(ad.MediaType) == "" {
		return errcode.New(errcode.CodeValidationError, "reward ad media type is invalid")
	}
	if ad.StartsAt != nil && ad.EndsAt != nil && !ad.EndsAt.After(*ad.StartsAt) {
		return errcode.New(errcode.CodeValidationError, "reward ad end time must be later than start time")
	}
	return nil
}

// 8. applyRewardAdUpdates applies patch fields to a rewarded ad.
func applyRewardAdUpdates(ad *model.RewardAd, params RewardAdUpdateParams, operatorUserID int64, now time.Time) error {
	if ad == nil {
		return errcode.New(errcode.CodeNotFound, "reward ad not found")
	}
	if params.Title != nil {
		ad.Title = strings.TrimSpace(*params.Title)
	}
	if params.Summary != nil {
		ad.Summary = strings.TrimSpace(*params.Summary)
	}
	if params.CoverURL != nil {
		ad.CoverURL = ""
	}
	if params.MediaURL != nil {
		ad.MediaURL = strings.TrimSpace(*params.MediaURL)
	}
	if params.MediaType != nil {
		ad.MediaType = normalizedRewardAdMediaType(*params.MediaType)
	}
	if params.TargetURL != nil {
		ad.TargetURL = strings.TrimSpace(*params.TargetURL)
	}
	if params.RewardPoints != nil {
		ad.RewardPoints = *params.RewardPoints
	}
	if params.WatchSeconds != nil {
		ad.WatchSeconds = normalizedWatchSeconds(*params.WatchSeconds)
	}
	if params.TotalBudget != nil {
		ad.TotalBudget = 0
	}
	if params.IsActive != nil {
		ad.IsActive = *params.IsActive
	}
	if params.StartsAtSet {
		ad.StartsAt = params.StartsAt
	}
	if params.EndsAtSet {
		ad.EndsAt = params.EndsAt
	}
	if params.RetentionDays != nil {
		if *params.RetentionDays <= 0 {
			return errcode.New(errcode.CodeValidationError, "reward ad retention days must be positive")
		}
		if *params.RetentionDays > maxRewardAdRetentionDays {
			return errcode.New(errcode.CodeValidationError, "reward ad retention days is too long")
		}
		endsAt := now.AddDate(0, 0, *params.RetentionDays)
		ad.EndsAt = &endsAt
	}
	ad.UpdatedBy = &operatorUserID
	ad.UpdatedAt = now
	return nil
}

// 9. toStaffRewardAdResponse maps a rewarded ad to staff API shape.
func toStaffRewardAdResponse(ad model.RewardAd) StaffRewardAdResponse {
	remainingBudget := ad.TotalBudget - ad.TotalGranted
	if ad.TotalBudget <= 0 {
		remainingBudget = 0
	}
	result := StaffRewardAdResponse{
		TaskID:          ad.PublicID,
		Title:           ad.Title,
		Summary:         ad.Summary,
		CoverURL:        ad.CoverURL,
		MediaURL:        ad.MediaURL,
		MediaType:       normalizedRewardAdMediaType(ad.MediaType),
		TargetURL:       ad.TargetURL,
		RewardPoints:    ad.RewardPoints,
		WatchSeconds:    normalizedWatchSeconds(ad.WatchSeconds),
		TotalBudget:     ad.TotalBudget,
		TotalGranted:    ad.TotalGranted,
		RemainingBudget: remainingBudget,
		IsActive:        ad.IsActive,
		CreatedAt:       ad.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       ad.UpdatedAt.Format(time.RFC3339),
	}
	if ad.StartsAt != nil {
		result.StartsAt = ad.StartsAt.Format(time.RFC3339)
	}
	if ad.EndsAt != nil {
		result.EndsAt = ad.EndsAt.Format(time.RFC3339)
	}
	return result
}

// 10. normalizedRewardAdMediaType returns a supported reward ad media type.
func normalizedRewardAdMediaType(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", rewardAdMediaTypeImage:
		return rewardAdMediaTypeImage
	case rewardAdMediaTypeVideo:
		return rewardAdMediaTypeVideo
	default:
		return ""
	}
}

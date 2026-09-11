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
	"sort"
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
	// 1. 空用途表示不篩選；不可經過 normalizedRewardAdType 後誤變成 reward。
	if rawAdType := strings.TrimSpace(filters.AdType); rawAdType != "" {
		adType := normalizedRewardAdType(rawAdType)
		if adType == "" {
			return nil, nil, errcode.New(errcode.CodeValidationError, "ad type is invalid")
		}
		if adType == rewardAdTypeReward {
			query = query.Where("(ad_type = ? OR ad_type = '')", rewardAdTypeReward)
		} else if adType == rewardAdTypeDisplay {
			query = query.Where("ad_type IN ?", displayAdTypes())
		} else {
			query = query.Where("ad_type = ?", adType)
		}
	}
	if channel := normalizedDisplayAdChannel(filters.DisplayChannel); channel != "" {
		query = query.Where("display_channel = ?", channel)
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

// 2.1 ListDisplayAdSettings returns configured listing-side display ad slots.
func (s *WalletService) ListDisplayAdSettings(ctx context.Context, channel string) (*DisplayAdChannelSettingsResponse, error) {
	normalizedChannel := normalizedDisplayAdChannel(channel)
	if normalizedChannel == "" {
		return nil, errcode.New(errcode.CodeValidationError, "display ad channel is invalid")
	}

	var assignments []model.DisplayAdSlotAssignment
	if err := s.runtime.DB.WithContext(ctx).
		Where("display_channel = ? AND display_placement = ?", normalizedChannel, displayAdPlacementListingSide).
		Order("slot_index asc, sort_order asc, id asc").
		Find(&assignments).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load display ad settings")
	}

	adIDs := make([]int64, 0, len(assignments))
	for _, assignment := range assignments {
		adIDs = append(adIDs, assignment.RewardAdID)
	}

	adMap, err := s.loadDisplayAdMap(ctx, normalizedChannel, adIDs)
	if err != nil {
		return nil, err
	}

	return buildDisplayAdSettingsResponse(normalizedChannel, assignments, adMap), nil
}

// 2.2 SaveDisplayAdSettings replaces one channel's listing-side slot settings.
func (s *WalletService) SaveDisplayAdSettings(ctx context.Context, operatorUserID int64, channel string, inputs []DisplayAdSlotInput) (*DisplayAdChannelSettingsResponse, error) {
	normalizedChannel := normalizedDisplayAdChannel(channel)
	if normalizedChannel == "" {
		return nil, errcode.New(errcode.CodeValidationError, "display ad channel is invalid")
	}
	normalizedInputs, adTaskIDs, err := normalizeDisplayAdSlotInputs(inputs)
	if err != nil {
		return nil, err
	}
	adMap, err := s.loadDisplayAdPublicIDMap(ctx, normalizedChannel, adTaskIDs)
	if err != nil {
		return nil, err
	}

	err = s.runtime.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("display_channel = ? AND display_placement = ?", normalizedChannel, displayAdPlacementListingSide).Delete(&model.DisplayAdSlotAssignment{}).Error; err != nil {
			return errcode.New(errcode.CodeInternalError, "failed to clear display ad settings")
		}
		for _, input := range normalizedInputs {
			for index, slotAd := range input.Ads {
				ad := adMap[slotAd.AdTaskID]
				if !displayAdTypeMatchesSlot(ad.AdType, input.SlotIndex) || normalizedDisplayAdLayout(ad.DisplayLayout) != displayAdLayoutForSlot(input.SlotIndex) {
					return errcode.New(errcode.CodeValidationError, "display ad layout does not match slot")
				}
				assignment := model.DisplayAdSlotAssignment{
					DisplayChannel:   normalizedChannel,
					DisplayPlacement: displayAdPlacementListingSide,
					SlotIndex:        input.SlotIndex,
					RewardAdID:       ad.ID,
					SortOrder:        index + 1,
					DisplayTitle:     slotAd.DisplayTitle,
					DisplayText:      slotAd.DisplayText,
					TargetURL:        slotAd.TargetURL,
					CreatedBy:        &operatorUserID,
					UpdatedBy:        &operatorUserID,
				}
				if err := tx.Create(&assignment).Error; err != nil {
					return errcode.New(errcode.CodeInternalError, "failed to save display ad settings")
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return s.ListDisplayAdSettings(ctx, normalizedChannel)
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
		PublicID:         utils.NewPublicID(),
		AdType:           normalizedRewardAdType(params.AdType),
		Title:            strings.TrimSpace(params.Title),
		Summary:          strings.TrimSpace(params.Summary),
		CoverURL:         "",
		MediaURL:         strings.TrimSpace(params.MediaURL),
		MediaType:        normalizedRewardAdMediaType(params.MediaType),
		TargetURL:        strings.TrimSpace(params.TargetURL),
		DisplayChannel:   normalizedDisplayAdChannel(params.DisplayChannel),
		DisplayPlacement: normalizedDisplayAdPlacement(params.DisplayPlacement),
		DisplayLayout:    normalizedDisplayAdLayout(params.DisplayLayout),
		SortOrder:        params.SortOrder,
		RewardPoints:     params.RewardPoints,
		WatchSeconds:     normalizedRewardAdWatchSeconds(normalizedRewardAdType(params.AdType), params.WatchSeconds),
		DailyUserLimit:   1,
		TotalBudget:      0,
		IsActive:         params.IsActive,
		StartsAt:         startsAt,
		EndsAt:           endsAt,
		CreatedBy:        &operatorUserID,
		UpdatedBy:        &operatorUserID,
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
		AdType:           normalizedRewardAdType(params.AdType),
		Title:            strings.TrimSpace(params.Title),
		Summary:          strings.TrimSpace(params.Summary),
		MediaURL:         strings.TrimSpace(params.MediaURL),
		MediaType:        normalizedRewardAdMediaType(params.MediaType),
		DisplayChannel:   normalizedDisplayAdChannel(params.DisplayChannel),
		DisplayPlacement: normalizedDisplayAdPlacement(params.DisplayPlacement),
		DisplayLayout:    normalizedDisplayAdLayout(params.DisplayLayout),
		RewardPoints:     params.RewardPoints,
		WatchSeconds:     normalizedRewardAdWatchSeconds(normalizedRewardAdType(params.AdType), params.WatchSeconds),
		StartsAt:         params.StartsAt,
		EndsAt:           params.EndsAt,
	}
	if ad.AdType == rewardAdTypeReward {
		if params.RetentionDays <= 0 {
			return errcode.New(errcode.CodeValidationError, "reward ad retention days must be positive")
		}
		if params.RetentionDays > maxRewardAdRetentionDays {
			return errcode.New(errcode.CodeValidationError, "reward ad retention days is too long")
		}
	}
	return validateRewardAdState(ad)
}

// 7. validateRewardAdState validates a rewarded ad before persistence.
func validateRewardAdState(ad model.RewardAd) error {
	if strings.TrimSpace(ad.Title) == "" {
		return errcode.New(errcode.CodeValidationError, "reward ad title is required")
	}
	adType := normalizedRewardAdType(ad.AdType)
	if adType == "" {
		return errcode.New(errcode.CodeValidationError, "ad type is invalid")
	}
	if strings.TrimSpace(ad.Summary) == "" && adType == rewardAdTypeReward {
		return errcode.New(errcode.CodeValidationError, "reward ad summary is required")
	}
	if strings.TrimSpace(ad.MediaURL) == "" {
		return errcode.New(errcode.CodeValidationError, "reward ad media is required")
	}
	if adType == rewardAdTypeReward && ad.RewardPoints <= 0 {
		return errcode.New(errcode.CodeValidationError, "reward points must be positive")
	}
	if adType == rewardAdTypeReward && normalizedWatchSeconds(ad.WatchSeconds) < 1 {
		return errcode.New(errcode.CodeValidationError, "watch seconds must be positive")
	}
	if normalizedRewardAdMediaType(ad.MediaType) == "" {
		return errcode.New(errcode.CodeValidationError, "reward ad media type is invalid")
	}
	if normalizedDisplayAdType(adType) != "" {
		if normalizedDisplayAdLayout(ad.DisplayLayout) == "" {
			return errcode.New(errcode.CodeValidationError, "display ad layout is invalid")
		}
		if !displayAdTypeMatchesLayout(adType, ad.DisplayLayout) {
			return errcode.New(errcode.CodeValidationError, "display ad layout does not match ad type")
		}
		if normalizedRewardAdMediaType(ad.MediaType) != rewardAdMediaTypeImage {
			return errcode.New(errcode.CodeValidationError, "display ad media must be image")
		}
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
	if params.AdType != nil {
		ad.AdType = normalizedRewardAdType(*params.AdType)
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
	if params.DisplayChannel != nil {
		ad.DisplayChannel = normalizedDisplayAdChannel(*params.DisplayChannel)
	}
	if params.DisplayPlacement != nil {
		ad.DisplayPlacement = normalizedDisplayAdPlacement(*params.DisplayPlacement)
	}
	if params.DisplayLayout != nil {
		ad.DisplayLayout = normalizedDisplayAdLayout(*params.DisplayLayout)
	}
	if params.SortOrder != nil {
		ad.SortOrder = *params.SortOrder
	}
	if params.RewardPoints != nil {
		ad.RewardPoints = *params.RewardPoints
	}
	if params.WatchSeconds != nil {
		ad.WatchSeconds = normalizedRewardAdWatchSeconds(ad.AdType, *params.WatchSeconds)
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
		TaskID:            ad.PublicID,
		AdType:            normalizedRewardAdType(ad.AdType),
		Title:             ad.Title,
		Summary:           ad.Summary,
		CoverURL:          ad.CoverURL,
		MediaURL:          ad.MediaURL,
		MediaType:         normalizedRewardAdMediaType(ad.MediaType),
		TargetURL:         ad.TargetURL,
		DisplayChannel:    normalizedDisplayAdChannel(ad.DisplayChannel),
		DisplayPlacement:  normalizedDisplayAdPlacement(ad.DisplayPlacement),
		DisplayLayout:     normalizedDisplayAdLayout(ad.DisplayLayout),
		SlotDisplayTitle:  "",
		DisplayText:       "",
		SlotTargetURL:     "",
		SortOrder:         ad.SortOrder,
		RewardPoints:      ad.RewardPoints,
		WatchSeconds:      normalizedRewardAdWatchSeconds(ad.AdType, ad.WatchSeconds),
		TotalBudget:       ad.TotalBudget,
		TotalGranted:      ad.TotalGranted,
		RemainingBudget:   remainingBudget,
		WatchCount:        ad.WatchCount,
		TotalWatchSeconds: ad.TotalWatchSeconds,
		LinkClickCount:    ad.LinkClickCount,
		LinkClickRate:     rewardAdClickRate(ad.WatchCount, ad.LinkClickCount),
		IsActive:          ad.IsActive,
		CreatedAt:         ad.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         ad.UpdatedAt.Format(time.RFC3339),
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

// 10.1 normalizedRewardAdWatchSeconds returns watch seconds for reward ads only.
func normalizedRewardAdWatchSeconds(adType string, value int) int {
	if normalizedRewardAdType(adType) != rewardAdTypeReward {
		return 0
	}
	return normalizedWatchSeconds(value)
}

// 11. normalizedRewardAdType returns a supported ad task type.
func normalizedRewardAdType(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", rewardAdTypeReward:
		return rewardAdTypeReward
	case rewardAdTypeDisplay:
		return rewardAdTypeDisplay
	case rewardAdTypeDisplayShort:
		return rewardAdTypeDisplayShort
	case rewardAdTypeDisplayLong:
		return rewardAdTypeDisplayLong
	default:
		return ""
	}
}

// 11.1 normalizedDisplayAdType returns a supported display ad type.
func normalizedDisplayAdType(value string) string {
	switch normalizedRewardAdType(value) {
	case rewardAdTypeDisplay, rewardAdTypeDisplayShort, rewardAdTypeDisplayLong:
		return normalizedRewardAdType(value)
	default:
		return ""
	}
}

// 12. normalizedDisplayAdChannel returns a supported listing display channel.
func normalizedDisplayAdChannel(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case displayAdChannelPropertySale:
		return displayAdChannelPropertySale
	case displayAdChannelServicedApartment:
		return displayAdChannelServicedApartment
	case displayAdChannelFurniture:
		return displayAdChannelFurniture
	default:
		return ""
	}
}

// 13. normalizedDisplayAdPlacement returns a supported display ad placement.
func normalizedDisplayAdPlacement(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case displayAdPlacementListingSide:
		return displayAdPlacementListingSide
	default:
		return ""
	}
}

// 14. normalizedDisplayAdLayout returns a supported display ad layout.
func normalizedDisplayAdLayout(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", displayAdLayoutImageText:
		return displayAdLayoutImageText
	case displayAdLayoutImageFull:
		return displayAdLayoutImageFull
	case displayAdLayoutTextCompact:
		return displayAdLayoutTextCompact
	default:
		return ""
	}
}

// 15. displayAdLayoutForSlot returns the fixed visual layout for one listing-side slot.
func displayAdLayoutForSlot(slotIndex int) string {
	switch slotIndex {
	case 1, 2, 3:
		return displayAdLayoutImageText
	case 4, 5:
		return displayAdLayoutImageFull
	default:
		return ""
	}
}

// 15.1 displayAdTypes returns all display-compatible ad types.
func displayAdTypes() []string {
	return []string{rewardAdTypeDisplay, rewardAdTypeDisplayShort, rewardAdTypeDisplayLong}
}

// 15.2 displayAdTypeMatchesSlot validates short and long display ad placement.
func displayAdTypeMatchesSlot(adType string, slotIndex int) bool {
	normalizedType := normalizedDisplayAdType(adType)
	switch slotIndex {
	case 1, 2, 3:
		return normalizedType == rewardAdTypeDisplayShort || normalizedType == rewardAdTypeDisplay
	case 4, 5:
		return normalizedType == rewardAdTypeDisplayLong || normalizedType == rewardAdTypeDisplay
	default:
		return false
	}
}

// 15.3 displayAdTypeMatchesLayout validates display ad type and visual layout.
func displayAdTypeMatchesLayout(adType string, layout string) bool {
	normalizedType := normalizedDisplayAdType(adType)
	normalizedLayout := normalizedDisplayAdLayout(layout)
	switch normalizedType {
	case rewardAdTypeDisplayLong:
		return normalizedLayout == displayAdLayoutImageFull
	case rewardAdTypeDisplayShort:
		return normalizedLayout == displayAdLayoutImageText
	case rewardAdTypeDisplay:
		return normalizedLayout != ""
	default:
		return false
	}
}

// 16. normalizeDisplayAdSlotInputs validates fixed listing-side ad slots.
func normalizeDisplayAdSlotInputs(inputs []DisplayAdSlotInput) ([]DisplayAdSlotInput, []string, error) {
	slotMap := map[int]DisplayAdSlotInput{}
	adTaskIDs := []string{}
	for _, input := range inputs {
		if input.SlotIndex < 1 || input.SlotIndex > displayAdSlotCount {
			return nil, nil, errcode.New(errcode.CodeValidationError, "display ad slot is invalid")
		}
		seenTaskIDs := map[string]bool{}
		sourceAds := input.Ads
		if len(sourceAds) == 0 {
			for _, taskID := range input.AdTaskIDs {
				sourceAds = append(sourceAds, DisplayAdSlotAdInput{AdTaskID: taskID})
			}
		}
		normalizedAds := make([]DisplayAdSlotAdInput, 0, len(sourceAds))
		for _, slotAd := range sourceAds {
			trimmedTaskID := strings.TrimSpace(slotAd.AdTaskID)
			if trimmedTaskID == "" || seenTaskIDs[trimmedTaskID] {
				continue
			}
			seenTaskIDs[trimmedTaskID] = true
			normalizedAds = append(normalizedAds, DisplayAdSlotAdInput{
				AdTaskID:     trimmedTaskID,
				DisplayTitle: trimDisplayAdSlotTitle(slotAd.DisplayTitle),
				DisplayText:  trimDisplayAdSlotText(slotAd.DisplayText),
				TargetURL:    trimDisplayAdSlotURL(slotAd.TargetURL),
			})
			adTaskIDs = append(adTaskIDs, trimmedTaskID)
		}
		slotMap[input.SlotIndex] = DisplayAdSlotInput{
			SlotIndex: input.SlotIndex,
			Ads:       normalizedAds,
		}
	}

	result := make([]DisplayAdSlotInput, 0, displayAdSlotCount)
	for slotIndex := 1; slotIndex <= displayAdSlotCount; slotIndex++ {
		input, exists := slotMap[slotIndex]
		if !exists {
			input = DisplayAdSlotInput{SlotIndex: slotIndex, Ads: []DisplayAdSlotAdInput{}}
		}
		result = append(result, input)
	}

	return result, adTaskIDs, nil
}

// 17. loadDisplayAdPublicIDMap loads valid display ads by public task IDs.
func (s *WalletService) loadDisplayAdPublicIDMap(ctx context.Context, channel string, taskIDs []string) (map[string]model.RewardAd, error) {
	result := map[string]model.RewardAd{}
	if len(taskIDs) == 0 {
		return result, nil
	}

	var ads []model.RewardAd
	if err := s.runtime.DB.WithContext(ctx).
		Where("public_id IN ?", taskIDs).
		Where("ad_type IN ?", displayAdTypes()).
		Where("media_type = ?", rewardAdMediaTypeImage).
		Where("display_channel = ?", channel).
		Find(&ads).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load image ads")
	}
	for _, ad := range ads {
		result[ad.PublicID] = ad
	}
	for _, taskID := range taskIDs {
		if _, exists := result[taskID]; !exists {
			return nil, errcode.New(errcode.CodeValidationError, "image ad task not found")
		}
	}
	return result, nil
}

// 18. loadDisplayAdMap loads display ads by numeric IDs.
func (s *WalletService) loadDisplayAdMap(ctx context.Context, channel string, adIDs []int64) (map[int64]model.RewardAd, error) {
	result := map[int64]model.RewardAd{}
	if len(adIDs) == 0 {
		return result, nil
	}

	var ads []model.RewardAd
	if err := s.runtime.DB.WithContext(ctx).
		Where("id IN ?", adIDs).
		Where("ad_type IN ?", displayAdTypes()).
		Where("media_type = ?", rewardAdMediaTypeImage).
		Find(&ads).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load image ads")
	}
	for _, ad := range ads {
		result[ad.ID] = ad
	}
	return result, nil
}

// 19. buildDisplayAdSettingsResponse maps slot assignments into fixed listing-side slots.
func buildDisplayAdSettingsResponse(channel string, assignments []model.DisplayAdSlotAssignment, adMap map[int64]model.RewardAd) *DisplayAdChannelSettingsResponse {
	slotAssignments := map[int][]model.DisplayAdSlotAssignment{}
	for _, assignment := range assignments {
		slotAssignments[assignment.SlotIndex] = append(slotAssignments[assignment.SlotIndex], assignment)
	}

	slots := make([]DisplayAdSlotResponse, 0, displayAdSlotCount)
	for slotIndex := 1; slotIndex <= displayAdSlotCount; slotIndex++ {
		assignmentsForSlot := slotAssignments[slotIndex]
		sort.SliceStable(assignmentsForSlot, func(left int, right int) bool {
			return assignmentsForSlot[left].SortOrder < assignmentsForSlot[right].SortOrder
		})
		ads := make([]StaffRewardAdResponse, 0, len(assignmentsForSlot))
		for _, assignment := range assignmentsForSlot {
			if ad, exists := adMap[assignment.RewardAdID]; exists {
				response := toStaffRewardAdResponse(ad)
				response.SlotDisplayTitle = assignment.DisplayTitle
				response.DisplayText = assignment.DisplayText
				response.SlotTargetURL = assignment.TargetURL
				ads = append(ads, response)
			}
		}
		slots = append(slots, DisplayAdSlotResponse{
			SlotIndex: slotIndex,
			Layout:    displayAdLayoutForSlot(slotIndex),
			Ads:       ads,
		})
	}

	return &DisplayAdChannelSettingsResponse{Channel: channel, Slots: slots}
}

// 20. trimDisplayAdSlotTitle normalizes optional slot display title.
func trimDisplayAdSlotTitle(value string) string {
	trimmed := strings.TrimSpace(value)
	if len([]rune(trimmed)) <= 160 {
		return trimmed
	}
	return string([]rune(trimmed)[:160])
}

// 21. trimDisplayAdSlotText normalizes optional slot display text.
func trimDisplayAdSlotText(value string) string {
	trimmed := strings.TrimSpace(value)
	if len([]rune(trimmed)) <= 255 {
		return trimmed
	}
	return string([]rune(trimmed)[:255])
}

// 22. trimDisplayAdSlotURL normalizes optional slot target URL.
func trimDisplayAdSlotURL(value string) string {
	trimmed := strings.TrimSpace(value)
	if len([]rune(trimmed)) <= 1024 {
		return trimmed
	}
	return string([]rune(trimmed)[:1024])
}

/*
 * Secondhand discover placement business logic.
 * 1. Manage manually curated discover page hero and category slots.
 * 2. Build public discover payloads from valid public secondhand listings.
 */
package service

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"gorm.io/gorm"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/model"
)

const (
	// 1. discoverSceneHero marks the top recommendation slots.
	discoverSceneHero = "discover_hero"
	// 2. discoverSceneCategory marks category carousel slots.
	discoverSceneCategory = "discover_category_carousel"
	// 3. discoverHeroSlotLimit defines the hero slot count.
	discoverHeroSlotLimit = 4
	// 4. discoverCategorySlotLimit defines each category slot count.
	discoverCategorySlotLimit = 10
)

// 1. GetPublicDiscover returns configured discover placements for public display.
func (s *SecondhandService) GetPublicDiscover(ctx context.Context) (*DiscoverPayload, error) {
	placements, err := s.loadDiscoverPlacements(ctx)
	if err != nil {
		return nil, err
	}

	summaryMap, err := s.loadPlacementListingSummaryMap(ctx, placements, true)
	if err != nil {
		return nil, err
	}

	payload := &DiscoverPayload{
		Hero:       []DiscoverPlacementResponse{},
		Categories: map[string][]DiscoverPlacementResponse{},
	}
	for _, placement := range placements {
		summary, ok := summaryMap[placement.ListingID]
		if !ok {
			continue
		}
		response := DiscoverPlacementResponse{
			Scene:        placement.Scene,
			CategoryCode: placement.CategoryCode,
			SlotIndex:    placement.SlotIndex,
			Listing:      &summary,
		}
		if placement.Scene == discoverSceneHero {
			payload.Hero = append(payload.Hero, response)
			continue
		}
		payload.Categories[placement.CategoryCode] = append(payload.Categories[placement.CategoryCode], response)
	}

	return payload, nil
}

// 2. GetDiscoverSettings returns the full discover slot matrix.
func (s *SecondhandService) GetDiscoverSettings(ctx context.Context) (*DiscoverSettingsPayload, error) {
	placements, err := s.loadDiscoverPlacements(ctx)
	if err != nil {
		return nil, err
	}

	summaryMap, err := s.loadPlacementListingSummaryMap(ctx, placements, false)
	if err != nil {
		return nil, err
	}

	placementMap := make(map[string]model.DiscoverPlacement, len(placements))
	for _, placement := range placements {
		placementMap[discoverPlacementKey(placement.Scene, placement.CategoryCode, placement.SlotIndex)] = placement
	}

	payload := &DiscoverSettingsPayload{
		Hero:       make([]DiscoverPlacementResponse, 0, discoverHeroSlotLimit),
		Categories: make(map[string][]DiscoverPlacementResponse, len(allowedSecondhandCategoryCodes)),
	}
	for slotIndex := 1; slotIndex <= discoverHeroSlotLimit; slotIndex++ {
		payload.Hero = append(payload.Hero, s.buildDiscoverSlotResponse(placementMap, summaryMap, discoverSceneHero, "", slotIndex))
	}
	for _, categoryCode := range allowedSecondhandCategoryCodes {
		payload.Categories[categoryCode] = make([]DiscoverPlacementResponse, 0, discoverCategorySlotLimit)
		for slotIndex := 1; slotIndex <= discoverCategorySlotLimit; slotIndex++ {
			payload.Categories[categoryCode] = append(
				payload.Categories[categoryCode],
				s.buildDiscoverSlotResponse(placementMap, summaryMap, discoverSceneCategory, categoryCode, slotIndex),
			)
		}
	}

	return payload, nil
}

// 3. SaveDiscoverPlacements persists a batch of discover slot assignments.
func (s *SecondhandService) SaveDiscoverPlacements(ctx context.Context, operatorUserID int64, inputs []DiscoverPlacementInput) (*DiscoverSettingsPayload, error) {
	if len(inputs) == 0 {
		return s.GetDiscoverSettings(ctx)
	}
	if err := validateDiscoverPlacementInputs(inputs); err != nil {
		return nil, err
	}

	err := s.runtime.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, input := range inputs {
			scene, categoryCode := normalizeDiscoverPlacementPosition(input.Scene, input.CategoryCode)
			listingPublicID := strings.TrimSpace(input.ListingID)
			if listingPublicID == "" {
				if err := tx.Where("scene = ? AND category_code = ? AND slot_index = ?", scene, categoryCode, input.SlotIndex).
					Delete(&model.DiscoverPlacement{}).Error; err != nil {
					return errcode.New(errcode.CodeInternalError, "failed to clear discover placement")
				}
				continue
			}

			listing, secondhand, _, err := s.loadListingByPublicIDWithDB(ctx, tx, listingPublicID)
			if err != nil {
				return err
			}
			if scene == discoverSceneCategory && secondhand.CategoryCode != categoryCode {
				return errcode.New(errcode.CodeValidationError, "listing category does not match discover slot")
			}

			if err := s.upsertDiscoverPlacement(ctx, tx, operatorUserID, scene, categoryCode, input.SlotIndex, listing.ID); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return s.GetDiscoverSettings(ctx)
}

// 4. ListSettingsSecondhand returns all secondhand listings for the settings page.
func (s *SecondhandService) ListSettingsSecondhand(ctx context.Context, filters SecondhandSettingsListFilters) ([]SecondhandListingSummary, *model.Pagination, error) {
	page, pageSize := normalizePagination(filters.Page, filters.PageSize)
	baseQuery := s.runtime.DB.WithContext(ctx).Table("listings").
		Select("listings.*, secondhand_listings.category_code, secondhand_listings.price_mode, secondhand_listings.price_hkd, secondhand_listings.condition_level, secondhand_listings.visibility_scope, secondhand_listings.contact_method, secondhand_listings.is_free_giveaway").
		Joins("JOIN secondhand_listings ON secondhand_listings.listing_id = listings.id").
		Where("listings.module = ? AND listings.is_deleted = ?", "secondhand", false)

	baseQuery = applySettingsSecondhandFilters(baseQuery, filters)
	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, nil, errcode.New(errcode.CodeInternalError, "failed to count settings listings")
	}

	var rows []secondhandListingRow
	if err := baseQuery.Order("listings.updated_at desc").Offset((page - 1) * pageSize).Limit(pageSize).Scan(&rows).Error; err != nil {
		return nil, nil, errcode.New(errcode.CodeInternalError, "failed to load settings listings")
	}

	items, err := s.buildListingSummaries(ctx, rows, nil)
	if err != nil {
		return nil, nil, err
	}

	return items, &model.Pagination{Page: page, PageSize: pageSize, Total: total}, nil
}

// 5. MarkSoldForSettings marks any secondhand listing as sold from settings.
func (s *SecondhandService) MarkSoldForSettings(ctx context.Context, listingPublicID string) error {
	listing, _, _, err := s.loadListingByPublicID(ctx, listingPublicID)
	if err != nil {
		return err
	}

	return s.runtime.DB.WithContext(ctx).Model(&model.Listing{}).
		Where("id = ?", listing.ID).
		Update("business_status", "sold").Error
}

// 6. DeactivateForSettings hides any secondhand listing from settings.
func (s *SecondhandService) DeactivateForSettings(ctx context.Context, listingPublicID string) error {
	listing, _, _, err := s.loadListingByPublicID(ctx, listingPublicID)
	if err != nil {
		return err
	}

	return s.runtime.DB.WithContext(ctx).Model(&model.Listing{}).
		Where("id = ?", listing.ID).
		Update("publication_status", "hidden").Error
}

// 7. loadDiscoverPlacements loads saved discover slots in display order.
func (s *SecondhandService) loadDiscoverPlacements(ctx context.Context) ([]model.DiscoverPlacement, error) {
	var placements []model.DiscoverPlacement
	if err := s.runtime.DB.WithContext(ctx).
		Order("scene asc, category_code asc, slot_index asc").
		Find(&placements).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load discover placements")
	}

	return placements, nil
}

// 8. loadPlacementListingSummaryMap loads summaries for placement listing IDs.
func (s *SecondhandService) loadPlacementListingSummaryMap(ctx context.Context, placements []model.DiscoverPlacement, publicOnly bool) (map[int64]SecondhandListingSummary, error) {
	result := make(map[int64]SecondhandListingSummary)
	listingIDs := make([]int64, 0, len(placements))
	seen := map[int64]bool{}
	for _, placement := range placements {
		if placement.ListingID <= 0 || seen[placement.ListingID] {
			continue
		}
		seen[placement.ListingID] = true
		listingIDs = append(listingIDs, placement.ListingID)
	}
	if len(listingIDs) == 0 {
		return result, nil
	}

	query := s.runtime.DB.WithContext(ctx).Table("listings").
		Select("listings.*, secondhand_listings.category_code, secondhand_listings.price_mode, secondhand_listings.price_hkd, secondhand_listings.condition_level, secondhand_listings.visibility_scope, secondhand_listings.contact_method, secondhand_listings.is_free_giveaway").
		Joins("JOIN secondhand_listings ON secondhand_listings.listing_id = listings.id").
		Where("listings.id IN ? AND listings.module = ? AND listings.is_deleted = ?", listingIDs, "secondhand", false)
	if publicOnly {
		query = query.Where("listings.publication_status = ? AND listings.moderation_status = ? AND listings.business_status = ? AND secondhand_listings.visibility_scope = ?", "active", "approved", "available", "public")
	}

	var rows []secondhandListingRow
	if err := query.Scan(&rows).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load discover listings")
	}

	summaries, err := s.buildListingSummaries(ctx, rows, nil)
	if err != nil {
		return nil, err
	}
	for index, row := range rows {
		result[row.ID] = summaries[index]
	}

	return result, nil
}

// 9. buildDiscoverSlotResponse builds one settings slot response.
func (s *SecondhandService) buildDiscoverSlotResponse(
	placementMap map[string]model.DiscoverPlacement,
	summaryMap map[int64]SecondhandListingSummary,
	scene string,
	categoryCode string,
	slotIndex int,
) DiscoverPlacementResponse {
	response := DiscoverPlacementResponse{
		Scene:        scene,
		CategoryCode: categoryCode,
		SlotIndex:    slotIndex,
	}
	if placement, ok := placementMap[discoverPlacementKey(scene, categoryCode, slotIndex)]; ok {
		if summary, exists := summaryMap[placement.ListingID]; exists {
			response.Listing = &summary
		}
	}

	return response
}

// 10. upsertDiscoverPlacement creates or updates one discover slot.
func (s *SecondhandService) upsertDiscoverPlacement(ctx context.Context, tx *gorm.DB, operatorUserID int64, scene string, categoryCode string, slotIndex int, listingID int64) error {
	var placement model.DiscoverPlacement
	err := tx.WithContext(ctx).
		Where("scene = ? AND category_code = ? AND slot_index = ?", scene, categoryCode, slotIndex).
		First(&placement).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return errcode.New(errcode.CodeInternalError, "failed to load discover placement")
	}
	if placement.ID == 0 {
		return tx.Create(&model.DiscoverPlacement{
			Scene:        scene,
			CategoryCode: categoryCode,
			SlotIndex:    slotIndex,
			ListingID:    listingID,
			CreatedBy:    &operatorUserID,
			UpdatedBy:    &operatorUserID,
		}).Error
	}

	return tx.Model(&placement).Updates(map[string]any{
		"listing_id":    listingID,
		"updated_by":    operatorUserID,
		"category_code": categoryCode,
	}).Error
}

// 11. validateDiscoverPlacementInputs validates a batch of placement payloads.
func validateDiscoverPlacementInputs(inputs []DiscoverPlacementInput) error {
	seen := map[string]bool{}
	for _, input := range inputs {
		scene, categoryCode := normalizeDiscoverPlacementPosition(input.Scene, input.CategoryCode)
		if !isValidDiscoverPlacementPosition(scene, categoryCode, input.SlotIndex) {
			return errcode.New(errcode.CodeValidationError, "invalid discover placement position")
		}

		key := discoverPlacementKey(scene, categoryCode, input.SlotIndex)
		if seen[key] {
			return errcode.New(errcode.CodeValidationError, "duplicated discover placement position")
		}
		seen[key] = true
	}

	return nil
}

// 12. applySettingsSecondhandFilters applies settings list filters.
func applySettingsSecondhandFilters(query *gorm.DB, filters SecondhandSettingsListFilters) *gorm.DB {
	if keyword := strings.TrimSpace(filters.Keyword); keyword != "" {
		pattern := "%" + keyword + "%"
		query = query.Where("(listings.title LIKE ? OR listings.summary LIKE ? OR listings.description LIKE ? OR listings.public_id LIKE ?)", pattern, pattern, pattern, pattern)
	}
	if categoryCode := strings.TrimSpace(filters.CategoryCode); categoryCode != "" {
		query = query.Where("secondhand_listings.category_code = ?", categoryCode)
	}
	switch strings.TrimSpace(filters.Status) {
	case "sold":
		query = query.Where("listings.business_status = ?", "sold")
	case "draft", "active", "hidden", "expired":
		query = query.Where("listings.publication_status = ? AND listings.business_status <> ?", filters.Status, "sold")
	}

	return query
}

// 13. normalizeDiscoverPlacementPosition normalizes scene and category position fields.
func normalizeDiscoverPlacementPosition(scene string, categoryCode string) (string, string) {
	normalizedScene := strings.TrimSpace(scene)
	normalizedCategory := strings.TrimSpace(categoryCode)
	if normalizedScene == discoverSceneHero {
		normalizedCategory = ""
	}

	return normalizedScene, normalizedCategory
}

// 14. isValidDiscoverPlacementPosition checks scene, category, and slot bounds.
func isValidDiscoverPlacementPosition(scene string, categoryCode string, slotIndex int) bool {
	if scene == discoverSceneHero {
		return categoryCode == "" && slotIndex >= 1 && slotIndex <= discoverHeroSlotLimit
	}
	if scene == discoverSceneCategory {
		return isAllowedSecondhandValue(categoryCode, allowedSecondhandCategoryCodes) && slotIndex >= 1 && slotIndex <= discoverCategorySlotLimit
	}

	return false
}

// 15. discoverPlacementKey returns a stable map key for one slot position.
func discoverPlacementKey(scene string, categoryCode string, slotIndex int) string {
	return scene + "|" + categoryCode + "|" + strconv.Itoa(slotIndex)
}

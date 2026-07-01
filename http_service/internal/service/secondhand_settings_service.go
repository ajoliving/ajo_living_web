/*
 * Secondhand settings business logic.
 * 1. Provide operator secondhand listing queries.
 * 2. Provide owner-independent sold and deactivate actions.
 */
package service

import (
	"context"
	"strings"

	"gorm.io/gorm"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/model"
)

// 1. ListSettingsSecondhand returns all secondhand listings for the settings page.
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

// 2. MarkSoldForSettings marks any secondhand listing as sold from settings.
func (s *SecondhandService) MarkSoldForSettings(ctx context.Context, listingPublicID string) error {
	listing, _, _, err := s.loadListingByPublicID(ctx, listingPublicID)
	if err != nil {
		return err
	}

	return s.runtime.DB.WithContext(ctx).Model(&model.Listing{}).
		Where("id = ?", listing.ID).
		Update("business_status", "sold").Error
}

// 3. DeactivateForSettings hides any secondhand listing from settings.
func (s *SecondhandService) DeactivateForSettings(ctx context.Context, listingPublicID string) error {
	listing, _, _, err := s.loadListingByPublicID(ctx, listingPublicID)
	if err != nil {
		return err
	}

	return s.runtime.DB.WithContext(ctx).Model(&model.Listing{}).
		Where("id = ?", listing.ID).
		Update("publication_status", "hidden").Error
}

// 4. applySettingsSecondhandFilters applies settings list filters.
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

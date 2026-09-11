/*
 * Secondhand staff management business logic.
 * 1. Provide staff-only secondhand listing queries.
 * 2. Provide owner-independent publish, deactivate, sold, and renew actions.
 */
package service

import (
	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/model"
	"context"
	"gorm.io/gorm"
	"strings"
	"time"
)

const (
	defaultStaffSecondhandRenewalDays = 14
	maxStaffSecondhandRenewalDays     = 365
)

// 1. ListStaffSecondhand returns all secondhand listings for staff management.
func (s *SecondhandService) ListStaffSecondhand(ctx context.Context, filters SecondhandSettingsListFilters) ([]SecondhandListingSummary, *model.Pagination, error) {
	return s.ListSettingsSecondhand(ctx, filters)
}

// 2. PublishForStaff publishes any non-deleted secondhand listing without owner checks.
func (s *SecondhandService) PublishForStaff(ctx context.Context, listingPublicID string) error {
	listing, _, _, err := s.loadListingByPublicID(ctx, strings.TrimSpace(listingPublicID))
	if err != nil {
		return err
	}

	now := s.runtime.Now()
	updates := map[string]any{
		"publication_status": "active",
		"moderation_status":  "approved",
		"business_status":    "available",
		"sort_refreshed_at":  now,
		"expire_at":          now.Add(time.Duration(defaultStaffSecondhandRenewalDays) * 24 * time.Hour),
	}
	if listing.PublishedAt == nil {
		updates["published_at"] = now
	}

	if err := s.runtime.DB.WithContext(ctx).Model(&model.Listing{}).Where("id = ?", listing.ID).Updates(updates).Error; err != nil {
		return errcode.New(errcode.CodeInternalError, "failed to publish listing")
	}

	return nil
}

// 3. DeactivateForStaff hides any secondhand listing.
func (s *SecondhandService) DeactivateForStaff(ctx context.Context, listingPublicID string) error {
	return s.DeactivateForSettings(ctx, listingPublicID)
}

// 4. RenewForStaff refreshes any secondhand listing validity.
func (s *SecondhandService) RenewForStaff(ctx context.Context, listingPublicID string, renewalDays int) error {
	listing, _, _, err := s.loadListingByPublicID(ctx, strings.TrimSpace(listingPublicID))
	if err != nil {
		return err
	}

	renewalDuration, err := resolveStaffSecondhandRenewalDuration(renewalDays)
	if err != nil {
		return err
	}

	now := s.runtime.Now()
	if err := s.runtime.DB.WithContext(ctx).Model(&model.Listing{}).Where("id = ?", listing.ID).Updates(map[string]any{
		"publication_status": "active",
		"moderation_status":  "approved",
		"business_status":    "available",
		"sort_refreshed_at":  now,
		"expire_at":          now.Add(renewalDuration),
	}).Error; err != nil {
		return errcode.New(errcode.CodeInternalError, "failed to renew listing")
	}

	return nil
}

// 5. resolveStaffSecondhandRenewalDuration validates renewal days.
func resolveStaffSecondhandRenewalDuration(renewalDays int) (time.Duration, error) {
	if renewalDays == 0 {
		renewalDays = defaultStaffSecondhandRenewalDays
	}
	if renewalDays < 0 {
		return 0, errcode.New(errcode.CodeValidationError, "renewal_days must be positive")
	}
	if renewalDays > maxStaffSecondhandRenewalDays {
		return 0, errcode.New(errcode.CodeValidationError, "renewal_days is too long")
	}

	return time.Duration(renewalDays) * 24 * time.Hour, nil
}

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

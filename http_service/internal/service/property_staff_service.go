/*
 * Property staff management business logic.
 * 1. Provide staff-only property listing queries.
 * 2. Provide owner-independent publish, deactivate, and renew actions.
 */
package service

import (
	"context"
	"strings"

	"gorm.io/gorm"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/model"
)

// 1. ListStaffProperties returns all property listings for one channel.
func (s *PropertyService) ListStaffProperties(ctx context.Context, channel PropertyChannel, filters PropertyListFilters) ([]PropertyListingSummary, *model.Pagination, error) {
	page, pageSize := normalizePagination(filters.Page, filters.PageSize)
	baseQuery := s.basePropertyListQuery(ctx, channel).
		Where("listings.module = ? AND listings.is_deleted = ?", string(channel), false)
	baseQuery = s.applyPropertyFilters(baseQuery, channel, filters)
	baseQuery = applyStaffPropertyStatusFilter(baseQuery, filters.Status)

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, nil, errcode.New(errcode.CodeInternalError, "failed to count staff listings")
	}

	var rows []propertyListingRow
	if err := baseQuery.Order("listings.updated_at desc").Offset((page - 1) * pageSize).Limit(pageSize).Scan(&rows).Error; err != nil {
		return nil, nil, errcode.New(errcode.CodeInternalError, "failed to load staff listings")
	}

	items, err := s.buildPropertySummaries(ctx, channel, rows)
	if err != nil {
		return nil, nil, err
	}

	return items, &model.Pagination{Page: page, PageSize: pageSize, Total: total}, nil
}

// 2. PublishPropertyForStaff publishes any property listing without owner checks.
func (s *PropertyService) PublishPropertyForStaff(ctx context.Context, channel PropertyChannel, listingPublicID string) error {
	return s.activatePropertyForStaff(ctx, channel, listingPublicID, false)
}

// 3. DeactivatePropertyForStaff hides any property listing.
func (s *PropertyService) DeactivatePropertyForStaff(ctx context.Context, channel PropertyChannel, listingPublicID string) error {
	listing, _, err := s.loadPropertyListingByPublicID(ctx, channel, strings.TrimSpace(listingPublicID))
	if err != nil {
		return err
	}

	if err := s.runtime.DB.WithContext(ctx).Model(&model.Listing{}).Where("id = ?", listing.ID).Update("publication_status", "hidden").Error; err != nil {
		return errcode.New(errcode.CodeInternalError, "failed to deactivate listing")
	}

	return nil
}

// 4. RenewPropertyForStaff refreshes any property listing validity.
func (s *PropertyService) RenewPropertyForStaff(ctx context.Context, channel PropertyChannel, listingPublicID string) error {
	return s.activatePropertyForStaff(ctx, channel, listingPublicID, true)
}

// 5. activatePropertyForStaff publishes or renews one listing.
func (s *PropertyService) activatePropertyForStaff(ctx context.Context, channel PropertyChannel, listingPublicID string, forcePublishedAt bool) error {
	listing, _, err := s.loadPropertyListingByPublicID(ctx, channel, strings.TrimSpace(listingPublicID))
	if err != nil {
		return err
	}

	now := s.runtime.Now()
	updates := map[string]any{
		"publication_status": "active",
		"moderation_status":  "approved",
		"business_status":    "available",
		"sort_refreshed_at":  now,
		"expire_at":          now.Add(propertyChannelTTL(channel)),
	}
	if listing.PublishedAt == nil || forcePublishedAt {
		updates["published_at"] = now
	}

	if err := s.runtime.DB.WithContext(ctx).Model(&model.Listing{}).Where("id = ?", listing.ID).Updates(updates).Error; err != nil {
		return errcode.New(errcode.CodeInternalError, "failed to activate listing")
	}

	return nil
}

// 6. applyStaffPropertyStatusFilter applies staff list status filters.
func applyStaffPropertyStatusFilter(query *gorm.DB, status string) *gorm.DB {
	normalizedStatus := strings.TrimSpace(status)
	switch normalizedStatus {
	case "sold":
		return query.Where("listings.business_status = ?", "sold")
	case "draft", "active", "hidden", "expired":
		return query.Where("listings.publication_status = ? AND listings.business_status <> ?", normalizedStatus, "sold")
	default:
		return query
	}
}

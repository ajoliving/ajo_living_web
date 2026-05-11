/*
 * Secondhand staff management business logic.
 * 1. Provide staff-only secondhand listing queries.
 * 2. Provide owner-independent publish, deactivate, and renew actions.
 */
package service

import (
	"context"
	"strings"
	"time"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/model"
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
		"expire_at":          now.Add(14 * 24 * time.Hour),
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
func (s *SecondhandService) RenewForStaff(ctx context.Context, listingPublicID string) error {
	listing, _, _, err := s.loadListingByPublicID(ctx, strings.TrimSpace(listingPublicID))
	if err != nil {
		return err
	}

	now := s.runtime.Now()
	if err := s.runtime.DB.WithContext(ctx).Model(&model.Listing{}).Where("id = ?", listing.ID).Updates(map[string]any{
		"publication_status": "active",
		"moderation_status":  "approved",
		"business_status":    "available",
		"sort_refreshed_at":  now,
		"expire_at":          now.Add(14 * 24 * time.Hour),
	}).Error; err != nil {
		return errcode.New(errcode.CodeInternalError, "failed to renew listing")
	}

	return nil
}

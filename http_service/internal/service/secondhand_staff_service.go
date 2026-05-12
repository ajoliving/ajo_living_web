/*
 * Secondhand staff management business logic.
 * 1. Provide staff-only secondhand listing queries.
 * 2. Provide owner-independent publish, deactivate, and configurable renew actions.
 */
package service

import (
	"context"
	"strings"
	"time"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/model"
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

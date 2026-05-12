/*
 * Secondhand favorite business logic.
 * 1. Manage member saved secondhand listings.
 * 2. Reuse public listing visibility checks for favorite eligibility.
 * 3. Return listing summaries for the member favorites page.
 */
package service

import (
	"context"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/model"
)

// 1. AddFavoriteSecondhand saves one visible secondhand listing for a member.
func (s *SecondhandService) AddFavoriteSecondhand(ctx context.Context, userID int64, communityID *int64, listingPublicID string) (*FavoriteResult, error) {
	listing, secondhand, _, err := s.loadListingByPublicID(ctx, listingPublicID)
	if err != nil {
		return nil, err
	}
	if err := s.validateFavoriteTarget(listing, secondhand, communityID); err != nil {
		return nil, err
	}

	favorite := model.ListingFavorite{
		UserID:    userID,
		ListingID: listing.ID,
	}
	err = s.runtime.DB.WithContext(ctx).
		Where("user_id = ? AND listing_id = ?", userID, listing.ID).
		FirstOrCreate(&favorite).Error
	if err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to save favorite")
	}

	return &FavoriteResult{ListingID: listing.PublicID, IsFavorited: true}, nil
}

// 2. RemoveFavoriteSecondhand removes one saved secondhand listing for a member.
func (s *SecondhandService) RemoveFavoriteSecondhand(ctx context.Context, userID int64, listingPublicID string) (*FavoriteResult, error) {
	listing, _, _, err := s.loadListingByPublicID(ctx, listingPublicID)
	if err != nil {
		return nil, err
	}

	if err := s.runtime.DB.WithContext(ctx).
		Where("user_id = ? AND listing_id = ?", userID, listing.ID).
		Delete(&model.ListingFavorite{}).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to remove favorite")
	}

	return &FavoriteResult{ListingID: listing.PublicID, IsFavorited: false}, nil
}

// 3. ListFavoriteSecondhand returns saved listings for the current member.
func (s *SecondhandService) ListFavoriteSecondhand(ctx context.Context, userID int64, communityID *int64, filters MySecondhandFilters) ([]SecondhandListingSummary, *model.Pagination, error) {
	page, pageSize := normalizePagination(filters.Page, filters.PageSize)
	baseQuery := s.runtime.DB.WithContext(ctx).Table("listing_favorites").
		Joins("JOIN listings ON listings.id = listing_favorites.listing_id").
		Joins("JOIN secondhand_listings ON secondhand_listings.listing_id = listings.id").
		Where("listing_favorites.user_id = ? AND listings.module = ? AND listings.publication_status = ? AND listings.moderation_status = ? AND listings.business_status = ? AND listings.is_deleted = ?", userID, "secondhand", "active", "approved", "available", false)
	if communityID != nil {
		baseQuery = baseQuery.Where("(secondhand_listings.visibility_scope = ? OR (secondhand_listings.visibility_scope = ? AND secondhand_listings.visible_community_id = ?))", "public", "building_only", *communityID)
	} else {
		baseQuery = baseQuery.Where("secondhand_listings.visibility_scope = ?", "public")
	}

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, nil, errcode.New(errcode.CodeInternalError, "failed to count favorite listings")
	}

	var rows []secondhandListingRow
	if err := baseQuery.
		Select("listings.*, secondhand_listings.category_code, secondhand_listings.price_mode, secondhand_listings.price_hkd, secondhand_listings.condition_level, secondhand_listings.visibility_scope, secondhand_listings.contact_method, secondhand_listings.is_free_giveaway").
		Order("listing_favorites.created_at desc").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&rows).Error; err != nil {
		return nil, nil, errcode.New(errcode.CodeInternalError, "failed to load favorite listings")
	}

	items, err := s.buildListingSummaries(ctx, rows, &userID)
	if err != nil {
		return nil, nil, err
	}

	return items, &model.Pagination{Page: page, PageSize: pageSize, Total: total}, nil
}

// 4. validateFavoriteTarget ensures only currently visible public listings are saved.
func (s *SecondhandService) validateFavoriteTarget(listing *model.Listing, secondhand *model.SecondhandListing, communityID *int64) error {
	if listing.PublicationStatus != "active" || listing.ModerationStatus != "approved" {
		return errcode.New(errcode.CodeHidden, "listing is not active")
	}
	if listing.BusinessStatus != "available" {
		return errcode.New(errcode.CodeAuthForbidden, "listing is not available")
	}
	if listing.IsDeleted {
		return errcode.New(errcode.CodeNotFound, "listing not found")
	}
	if !s.canViewListing(listing, secondhand, communityID) {
		return errcode.New(errcode.CodeVisibilityForbidden, "listing is not visible to the current user")
	}

	return nil
}

/*
 * Secondhand listing business logic.
 * 1. Manage secondhand draft, publish, republish, and owner flows.
 * 2. Enforce visibility, ownership, and listing lifecycle rules.
 */
package service

import (
	"context"
	"strings"
	"time"

	"gorm.io/gorm"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/model"
	"ajoliving_web/http_service/internal/utils"
)

// 1. SecondhandService handles secondhand listing workflows.
type SecondhandService struct {
	runtime *Runtime
}

// 2. NewSecondhandService creates a secondhand service instance.
func NewSecondhandService(runtime *Runtime) *SecondhandService {
	return &SecondhandService{runtime: runtime}
}

// 3. CreateSecondhandListing creates a secondhand draft.
func (s *SecondhandService) CreateSecondhandListing(ctx context.Context, params UpsertSecondhandParams) (*SecondhandListingDetail, error) {
	listingID, err := s.upsertSecondhand(ctx, params, true)
	if err != nil {
		return nil, err
	}

	return s.GetSecondhandDetail(ctx, listingID, &params.OwnerUserID, nil)
}

// 4. UpdateSecondhandListing updates a secondhand listing owned by the current user.
func (s *SecondhandService) UpdateSecondhandListing(ctx context.Context, params UpsertSecondhandParams) (*SecondhandListingDetail, error) {
	listingID, err := s.upsertSecondhand(ctx, params, false)
	if err != nil {
		return nil, err
	}

	return s.GetSecondhandDetail(ctx, listingID, &params.OwnerUserID, nil)
}

// 5. PublishSecondhandListing publishes a prepared secondhand draft.
func (s *SecondhandService) PublishSecondhandListing(ctx context.Context, ownerUserID int64, listingPublicID string) (*SecondhandListingDetail, error) {
	listing, _, _, err := s.loadOwnedListing(ctx, ownerUserID, listingPublicID)
	if err != nil {
		return nil, err
	}

	if listing.PublicationStatus != "draft" {
		return nil, errcode.New(errcode.CodeValidationError, "only draft listings can be published")
	}

	if err := s.validateListingReady(ctx, ownerUserID, listing.ID); err != nil {
		return nil, err
	}

	now := s.runtime.Now()
	expireAt := now.Add(14 * 24 * time.Hour)
	if err := s.runtime.DB.WithContext(ctx).Model(&model.Listing{}).
		Where("id = ?", listing.ID).
		Updates(map[string]any{
			"publication_status": "active",
			"published_at":       now,
			"sort_refreshed_at":  now,
			"expire_at":          expireAt,
		}).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to publish listing")
	}

	return s.GetSecondhandDetail(ctx, listing.PublicID, &ownerUserID, nil)
}

// 6. RepublishSecondhandListing republishes an expired listing and refreshes sort order.
func (s *SecondhandService) RepublishSecondhandListing(ctx context.Context, ownerUserID int64, listingPublicID string) (*SecondhandListingDetail, error) {
	listing, _, _, err := s.loadOwnedListing(ctx, ownerUserID, listingPublicID)
	if err != nil {
		return nil, err
	}

	if listing.PublicationStatus != "expired" {
		return nil, errcode.New(errcode.CodeValidationError, "only expired listings can be republished")
	}

	now := s.runtime.Now()
	expireAt := now.Add(14 * 24 * time.Hour)
	if err := s.runtime.DB.WithContext(ctx).Model(&model.Listing{}).
		Where("id = ?", listing.ID).
		Updates(map[string]any{
			"publication_status": "active",
			"sort_refreshed_at":  now,
			"expire_at":          expireAt,
			"business_status":    "available",
		}).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to republish listing")
	}

	return s.GetSecondhandDetail(ctx, listing.PublicID, &ownerUserID, nil)
}

// 7. MarkSold marks a listing as sold and removes it from active discovery.
func (s *SecondhandService) MarkSold(ctx context.Context, ownerUserID int64, listingPublicID string) error {
	listing, _, _, err := s.loadOwnedListing(ctx, ownerUserID, listingPublicID)
	if err != nil {
		return err
	}

	return s.runtime.DB.WithContext(ctx).Model(&model.Listing{}).
		Where("id = ?", listing.ID).
		Update("business_status", "sold").Error
}

// 8. Deactivate hides a listing from public discovery.
func (s *SecondhandService) Deactivate(ctx context.Context, ownerUserID int64, listingPublicID string) error {
	listing, _, _, err := s.loadOwnedListing(ctx, ownerUserID, listingPublicID)
	if err != nil {
		return err
	}

	return s.runtime.DB.WithContext(ctx).Model(&model.Listing{}).
		Where("id = ?", listing.ID).
		Update("publication_status", "hidden").Error
}

// 9. ListPublicSecondhand returns public secondhand listings with visibility filtering.
func (s *SecondhandService) ListPublicSecondhand(ctx context.Context, filters SecondhandListFilters) ([]SecondhandListingSummary, *model.Pagination, error) {
	page, pageSize := normalizePagination(filters.Page, filters.PageSize)
	baseQuery := s.runtime.DB.WithContext(ctx).Table("listings").
		Select("listings.*, secondhand_listings.category_code, secondhand_listings.price_mode, secondhand_listings.price_hkd, secondhand_listings.condition_level, secondhand_listings.visibility_scope, secondhand_listings.contact_method, secondhand_listings.is_free_giveaway").
		Joins("JOIN secondhand_listings ON secondhand_listings.listing_id = listings.id").
		Where("listings.module = ? AND listings.publication_status = ? AND listings.moderation_status = ? AND listings.business_status = ? AND listings.is_deleted = ?", "secondhand", "active", "approved", "available", false)

	baseQuery = s.applyPublicFilters(baseQuery, filters)
	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, nil, errcode.New(errcode.CodeInternalError, "failed to count listings")
	}

	var rows []secondhandListingRow
	query := baseQuery
	switch filters.SortBy {
	case "price_asc":
		query = query.Order("secondhand_listings.price_hkd asc")
	case "price_desc":
		query = query.Order("secondhand_listings.price_hkd desc")
	default:
		query = query.Order("listings.sort_refreshed_at desc NULLS LAST, listings.created_at desc")
	}

	if err := query.Offset((page - 1) * pageSize).Limit(pageSize).Scan(&rows).Error; err != nil {
		return nil, nil, errcode.New(errcode.CodeInternalError, "failed to load listings")
	}

	items, err := s.buildListingSummaries(ctx, rows)
	if err != nil {
		return nil, nil, err
	}

	return items, &model.Pagination{Page: page, PageSize: pageSize, Total: total}, nil
}

// 10. ListMySecondhand returns listings owned by the current user.
func (s *SecondhandService) ListMySecondhand(ctx context.Context, ownerUserID int64, filters MySecondhandFilters) ([]SecondhandListingSummary, *model.Pagination, error) {
	page, pageSize := normalizePagination(filters.Page, filters.PageSize)
	baseQuery := s.runtime.DB.WithContext(ctx).Table("listings").
		Select("listings.*, secondhand_listings.category_code, secondhand_listings.price_mode, secondhand_listings.price_hkd, secondhand_listings.condition_level, secondhand_listings.visibility_scope, secondhand_listings.contact_method, secondhand_listings.is_free_giveaway").
		Joins("JOIN secondhand_listings ON secondhand_listings.listing_id = listings.id").
		Where("listings.module = ? AND listings.owner_user_id = ? AND listings.is_deleted = ?", "secondhand", ownerUserID, false)

	if filters.Status != "" {
		if filters.Status == "sold" {
			baseQuery = baseQuery.Where("listings.business_status = ?", "sold")
		} else {
			baseQuery = baseQuery.Where("listings.publication_status = ?", filters.Status).Where("listings.business_status <> ?", "sold")
		}
	}

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, nil, errcode.New(errcode.CodeInternalError, "failed to count owner listings")
	}

	var rows []secondhandListingRow
	if err := baseQuery.Order("listings.updated_at desc").Offset((page - 1) * pageSize).Limit(pageSize).Scan(&rows).Error; err != nil {
		return nil, nil, errcode.New(errcode.CodeInternalError, "failed to load owner listings")
	}

	items, err := s.buildListingSummaries(ctx, rows)
	if err != nil {
		return nil, nil, err
	}

	return items, &model.Pagination{Page: page, PageSize: pageSize, Total: total}, nil
}

// 11. GetSecondhandDetail returns a single secondhand listing detail payload.
func (s *SecondhandService) GetSecondhandDetail(ctx context.Context, listingPublicID string, viewerUserID *int64, viewerCommunityID *int64) (*SecondhandListingDetail, error) {
	listing, secondhand, contact, err := s.loadListingByPublicID(ctx, listingPublicID)
	if err != nil {
		return nil, err
	}

	if listing.PublicationStatus == "hidden" {
		return nil, errcode.New(errcode.CodeHidden, "listing is hidden")
	}

	if viewerUserID == nil || *viewerUserID != listing.OwnerUserID {
		if !s.canViewListing(listing, secondhand, viewerCommunityID) {
			return nil, errcode.New(errcode.CodeVisibilityForbidden, "listing is not visible to the current user")
		}
	}

	detail, err := s.buildListingDetail(ctx, listing, secondhand, contact)
	if err != nil {
		return nil, err
	}

	if viewerUserID == nil && detail.VisibilityScope == "building_only" {
		return nil, errcode.New(errcode.CodeVisibilityForbidden, "listing is not visible to the current user")
	}

	return detail, nil
}

// 12. upsertSecondhand creates or updates the listing aggregate.
func (s *SecondhandService) upsertSecondhand(ctx context.Context, params UpsertSecondhandParams, creating bool) (string, error) {
	if err := s.validateUpsertParams(params); err != nil {
		return "", err
	}

	communityID, err := s.resolveCommunityID(ctx, params.OwnerUserID, params.CommunityID, params.VisibilityScope)
	if err != nil {
		return "", err
	}

	returnPublicID := params.ListingPublicID
	err = s.runtime.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var listing model.Listing
		var secondhand model.SecondhandListing
		if creating {
			listing = model.Listing{
				PublicID:              utils.NewPublicID(),
				Module:                "secondhand",
				OwnerUserID:           params.OwnerUserID,
				Title:                 params.Title,
				Summary:               params.Summary,
				Description:           params.Description,
				DistrictCode:          params.DistrictCode,
				CommunityID:           communityID,
				PublisherIdentityType: s.fallbackPublisherIdentity(params.PublisherIdentityType),
				PublicationStatus:     "draft",
				ModerationStatus:      "approved",
				BusinessStatus:        "available",
				IsDeleted:             false,
			}
			if err := tx.Create(&listing).Error; err != nil {
				return err
			}
			returnPublicID = listing.PublicID
		} else {
			current, currentSecondhand, _, err := s.loadOwnedListingWithTx(ctx, tx, params.OwnerUserID, params.ListingPublicID)
			if err != nil {
				return err
			}
			listing = *current
			secondhand = *currentSecondhand
		}

		requestedBusinessStatus := strings.TrimSpace(params.BusinessStatus)

		listing.Title = params.Title
		listing.Summary = params.Summary
		listing.Description = params.Description
		listing.DistrictCode = params.DistrictCode
		listing.CommunityID = communityID
		listing.PublisherIdentityType = s.fallbackPublisherIdentity(params.PublisherIdentityType)
		if !creating && requestedBusinessStatus != "" {
			listing.BusinessStatus = requestedBusinessStatus
		}
		if err := tx.Save(&listing).Error; err != nil {
			return err
		}
		// 12.1 明確保存交易狀態，支援已售出帖子重新改回未售出
		if !creating && requestedBusinessStatus != "" {
			if err := tx.Model(&model.Listing{}).
				Where("id = ?", listing.ID).
				Update("business_status", requestedBusinessStatus).Error; err != nil {
				return errcode.New(errcode.CodeInternalError, "failed to update listing business status")
			}
		}

		deliveryTags, err := marshalJSON(params.DeliveryTags)
		if err != nil {
			return err
		}

		secondhand.ListingID = listing.ID
		secondhand.CategoryCode = params.CategoryCode
		secondhand.PriceMode = params.PriceMode
		secondhand.PriceHKD = normalizePrice(params.PriceMode, params.PriceHKD)
		secondhand.ConditionLevel = params.ConditionLevel
		secondhand.DimensionText = params.DimensionText
		secondhand.PickupRegionCode = params.PickupRegionCode
		secondhand.PickupLocationText = params.PickupLocationText
		secondhand.DeliveryTags = deliveryTags
		secondhand.VisibilityScope = params.VisibilityScope
		secondhand.VisibleCommunityID = visibleCommunityID(params.VisibilityScope, communityID)
		secondhand.ContactMethod = params.ContactMethod
		secondhand.IsFreeGiveaway = params.PriceMode == "free"
		if err := tx.Save(&secondhand).Error; err != nil {
			return err
		}

		contact, err := s.buildListingContact(listing.ID, params.Contact, params.ContactMethod)
		if err != nil {
			return err
		}
		if err := tx.Save(contact).Error; err != nil {
			return err
		}

		if err := tx.Where("listing_id = ?", listing.ID).Delete(&model.ListingImage{}).Error; err != nil {
			return err
		}

		images, err := s.resolveListingImages(ctx, tx, listing.ID, params.OwnerUserID, params.Images)
		if err != nil {
			return err
		}
		if len(images) > 0 {
			if err := tx.Create(&images).Error; err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	return returnPublicID, nil
}

// 13. validateUpsertParams validates create and update payloads.
func (s *SecondhandService) validateUpsertParams(params UpsertSecondhandParams) error {
	if strings.TrimSpace(params.Title) == "" || strings.TrimSpace(params.CategoryCode) == "" || strings.TrimSpace(params.PriceMode) == "" || strings.TrimSpace(params.ConditionLevel) == "" {
		return errcode.New(errcode.CodeValidationError, "missing required listing fields")
	}

	if strings.TrimSpace(params.PickupRegionCode) == "" || strings.TrimSpace(params.VisibilityScope) == "" || strings.TrimSpace(params.ContactMethod) == "" {
		return errcode.New(errcode.CodeValidationError, "missing required pickup or contact fields")
	}

	if !isAllowedSecondhandValue(params.CategoryCode, allowedSecondhandCategoryCodes) {
		return errcode.New(errcode.CodeValidationError, "invalid category code")
	}

	if !isAllowedSecondhandDistrict(params.DistrictCode) || !isAllowedSecondhandDistrict(params.PickupRegionCode) {
		return errcode.New(errcode.CodeValidationError, "invalid district code")
	}

	if params.VisibilityScope == "building_only" && strings.TrimSpace(params.CommunityID) == "" {
		return errcode.New(errcode.CodeValidationError, "community is required for building_only visibility")
	}

	if !isAllowedSecondhandValue(params.PriceMode, []string{"fixed", "negotiable", "free"}) {
		return errcode.New(errcode.CodeValidationError, "invalid price mode")
	}

	if !isAllowedSecondhandValue(params.VisibilityScope, []string{"public", "building_only"}) {
		return errcode.New(errcode.CodeValidationError, "invalid visibility scope")
	}

	if !isAllowedSecondhandValue(params.ConditionLevel, []string{"brand_new", "used_excellent", "used_good", "used_fair"}) {
		return errcode.New(errcode.CodeValidationError, "invalid condition level")
	}

	if !isAllowedSecondhandValue(params.ContactMethod, []string{"phone", "whatsapp", "chat", "both", "chat_or_whatsapp"}) {
		return errcode.New(errcode.CodeValidationError, "invalid contact method")
	}

	if strings.TrimSpace(params.BusinessStatus) != "" && !isAllowedSecondhandValue(params.BusinessStatus, []string{"available", "sold"}) {
		return errcode.New(errcode.CodeValidationError, "invalid business status")
	}

	if !params.Contact.ShowPhone && !params.Contact.ShowWhatsApp && !params.Contact.ShowChat {
		return errcode.New(errcode.CodeValidationError, "at least one contact channel must be enabled")
	}

	return nil
}

// 14. applyPublicFilters applies search and visibility filters to the list query.
func (s *SecondhandService) applyPublicFilters(query *gorm.DB, filters SecondhandListFilters) *gorm.DB {
	if filters.Keyword != "" {
		likeKeyword := "%" + strings.TrimSpace(filters.Keyword) + "%"
		query = query.Where("(listings.title LIKE ? OR listings.summary LIKE ? OR listings.description LIKE ?)", likeKeyword, likeKeyword, likeKeyword)
	}
	if filters.CategoryCode != "" {
		query = query.Where("secondhand_listings.category_code = ?", filters.CategoryCode)
	}
	if filters.RegionCode != "" {
		districts, ok := secondhandDistrictsForRegion(filters.RegionCode)
		if !ok {
			return query.Where("1 = 0")
		}
		query = query.Where("listings.district_code IN ?", districts)
	}
	if filters.DistrictCode != "" {
		query = query.Where("listings.district_code = ?", filters.DistrictCode)
	}
	if filters.PriceMode != "" {
		query = query.Where("secondhand_listings.price_mode = ?", filters.PriceMode)
	}
	if filters.ConditionLevel != "" {
		query = query.Where("secondhand_listings.condition_level = ?", filters.ConditionLevel)
	}
	if filters.VisibilityScope != "" {
		query = query.Where("secondhand_listings.visibility_scope = ?", filters.VisibilityScope)
	}
	if filters.MinPriceHKD != nil {
		query = query.Where("secondhand_listings.price_hkd >= ?", *filters.MinPriceHKD)
	}
	if filters.MaxPriceHKD != nil {
		query = query.Where("secondhand_listings.price_hkd <= ?", *filters.MaxPriceHKD)
	}
	if filters.OnlyFree {
		query = query.Where("secondhand_listings.is_free_giveaway = ?", true)
	}
	if filters.ExcludeFree {
		query = query.Where("secondhand_listings.is_free_giveaway = ?", false)
	}
	if filters.OnlyBuilding {
		if filters.CommunityID == nil {
			return query.Where("1 = 0")
		}
		return query.Where("secondhand_listings.visibility_scope = ? AND secondhand_listings.visible_community_id = ?", "building_only", *filters.CommunityID)
	}
	if filters.CommunityID != nil {
		return query.Where("(secondhand_listings.visibility_scope = ? OR (secondhand_listings.visibility_scope = ? AND secondhand_listings.visible_community_id = ?))", "public", "building_only", *filters.CommunityID)
	}

	return query.Where("secondhand_listings.visibility_scope = ?", "public")
}

// 15. fallbackPublisherIdentity falls back to the stored profile identity when missing.
func (s *SecondhandService) fallbackPublisherIdentity(identity string) string {
	if strings.TrimSpace(identity) == "" {
		return "owner"
	}
	return strings.TrimSpace(identity)
}

// 16. isAllowedSecondhandValue validates a normalized enum-like input.
func isAllowedSecondhandValue(value string, allowedValues []string) bool {
	normalizedValue := strings.TrimSpace(value)
	for _, allowedValue := range allowedValues {
		if normalizedValue == allowedValue {
			return true
		}
	}

	return false
}

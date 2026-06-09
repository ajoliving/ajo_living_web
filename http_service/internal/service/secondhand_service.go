/*
 * Secondhand listing business logic.
 * 1. Manage secondhand draft, publish, republish, and owner flows.
 * 2. Enforce visibility, ownership, and listing lifecycle rules.
 */
package service

import (
	"context"
	"fmt"
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
	listingID, charge, err := s.createSecondhandWithCharge(ctx, params)
	if err != nil {
		return nil, err
	}

	result, err := s.GetSecondhandDetail(ctx, listingID, &params.OwnerUserID, nil)
	if err != nil {
		return nil, err
	}
	attachSecondhandCharge(result, charge)
	return result, nil
}

// 4. UpdateSecondhandListing updates a secondhand listing owned by the current user.
func (s *SecondhandService) UpdateSecondhandListing(ctx context.Context, params UpsertSecondhandParams) (*SecondhandListingDetail, error) {
	listingID, charge, err := s.updateSecondhandWithCharge(ctx, params)
	if err != nil {
		return nil, err
	}

	result, err := s.GetSecondhandDetail(ctx, listingID, &params.OwnerUserID, nil)
	if err != nil {
		return nil, err
	}
	attachSecondhandCharge(result, charge)
	return result, nil
}

// 5. PublishSecondhandListing publishes a prepared secondhand draft.
func (s *SecondhandService) PublishSecondhandListing(ctx context.Context, ownerUserID int64, listingPublicID string) (*SecondhandListingDetail, error) {
	return s.publishSecondhandWithCharge(ctx, ownerUserID, listingPublicID, WalletActionPublish)
}

// 6. RepublishSecondhandListing republishes an expired listing and refreshes sort order.
func (s *SecondhandService) RepublishSecondhandListing(ctx context.Context, ownerUserID int64, listingPublicID string) (*SecondhandListingDetail, error) {
	return s.publishSecondhandWithCharge(ctx, ownerUserID, listingPublicID, WalletActionRepublish)
}

// 7. RenewSecondhandListing renews an active listing and charges the owner.
func (s *SecondhandService) RenewSecondhandListing(ctx context.Context, ownerUserID int64, listingPublicID string) (*SecondhandListingDetail, error) {
	var listingPublicIDResult string
	var charge *PointsChargeResponse
	err := s.runtime.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		listing, _, _, err := s.loadOwnedListingWithTx(ctx, tx, ownerUserID, listingPublicID)
		if err != nil {
			return err
		}
		if listing.PublicationStatus != "active" {
			return errcode.New(errcode.CodeValidationError, "only active listings can be renewed")
		}
		if listing.BusinessStatus == "sold" {
			return errcode.New(errcode.CodeValidationError, "sold listings cannot be renewed")
		}

		chargeResult, err := s.chargeListingAction(ctx, tx, listing, WalletActionRenew)
		if err != nil {
			return err
		}
		charge = chargeResult

		now := s.runtime.Now()
		if err := tx.Model(&model.Listing{}).Where("id = ?", listing.ID).Updates(map[string]any{
			"publication_status": "active",
			"business_status":    "available",
			"sort_refreshed_at":  now,
			"expire_at":          now.Add(14 * 24 * time.Hour),
		}).Error; err != nil {
			return errcode.New(errcode.CodeInternalError, "failed to renew listing")
		}
		listingPublicIDResult = listing.PublicID
		return nil
	})
	if err != nil {
		return nil, err
	}

	result, err := s.GetSecondhandDetail(ctx, listingPublicIDResult, &ownerUserID, nil)
	if err != nil {
		return nil, err
	}
	attachSecondhandCharge(result, charge)
	return result, nil
}

// 8. MarkSold marks a listing as sold and removes it from active discovery.
func (s *SecondhandService) MarkSold(ctx context.Context, ownerUserID int64, listingPublicID string) error {
	listing, _, _, err := s.loadOwnedListing(ctx, ownerUserID, listingPublicID)
	if err != nil {
		return err
	}

	return s.runtime.DB.WithContext(ctx).Model(&model.Listing{}).
		Where("id = ?", listing.ID).
		Update("business_status", "sold").Error
}

// 9. Deactivate hides a listing from public discovery.
func (s *SecondhandService) Deactivate(ctx context.Context, ownerUserID int64, listingPublicID string) error {
	listing, _, _, err := s.loadOwnedListing(ctx, ownerUserID, listingPublicID)
	if err != nil {
		return err
	}

	return s.runtime.DB.WithContext(ctx).Model(&model.Listing{}).
		Where("id = ?", listing.ID).
		Update("publication_status", "hidden").Error
}

// 10. ListPublicSecondhand returns public secondhand listings with visibility filtering.
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

	items, err := s.buildListingSummaries(ctx, rows, filters.ViewerUserID)
	if err != nil {
		return nil, nil, err
	}

	return items, &model.Pagination{Page: page, PageSize: pageSize, Total: total}, nil
}

// 11. ListMySecondhand returns listings owned by the current user.
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

	items, err := s.buildListingSummaries(ctx, rows, &ownerUserID)
	if err != nil {
		return nil, nil, err
	}

	return items, &model.Pagination{Page: page, PageSize: pageSize, Total: total}, nil
}

// 12. GetSecondhandDetail returns a single secondhand listing detail payload.
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

	detail, err := s.buildListingDetail(ctx, listing, secondhand, contact, viewerUserID)
	if err != nil {
		return nil, err
	}

	if viewerUserID == nil && detail.VisibilityScope == "building_only" {
		return nil, errcode.New(errcode.CodeVisibilityForbidden, "listing is not visible to the current user")
	}

	return detail, nil
}

// 13. upsertSecondhand creates or updates the listing aggregate.
func (s *SecondhandService) upsertSecondhand(ctx context.Context, params UpsertSecondhandParams, creating bool) (string, error) {
	if err := s.validateUpsertParams(params); err != nil {
		return "", err
	}

	communityID, err := s.resolveCommunityID(ctx, params.OwnerUserID, params.CommunityID, params.CommunityName, params.VisibilityScope)
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

// 14. createSecondhandAggregate creates the listing aggregate inside a transaction.
func (s *SecondhandService) createSecondhandAggregate(ctx context.Context, tx *gorm.DB, params UpsertSecondhandParams, communityID *int64) (*model.Listing, error) {
	listing := model.Listing{
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
		return nil, err
	}

	deliveryTags, err := marshalJSON(params.DeliveryTags)
	if err != nil {
		return nil, err
	}
	secondhand := model.SecondhandListing{
		ListingID:          listing.ID,
		CategoryCode:       params.CategoryCode,
		PriceMode:          params.PriceMode,
		PriceHKD:           normalizePrice(params.PriceMode, params.PriceHKD),
		ConditionLevel:     params.ConditionLevel,
		DimensionText:      params.DimensionText,
		PickupRegionCode:   params.PickupRegionCode,
		PickupLocationText: params.PickupLocationText,
		DeliveryTags:       deliveryTags,
		VisibilityScope:    params.VisibilityScope,
		VisibleCommunityID: visibleCommunityID(params.VisibilityScope, communityID),
		ContactMethod:      params.ContactMethod,
		IsFreeGiveaway:     params.PriceMode == "free",
	}
	if err := tx.Create(&secondhand).Error; err != nil {
		return nil, err
	}

	contact, err := s.buildListingContact(listing.ID, params.Contact, params.ContactMethod)
	if err != nil {
		return nil, err
	}
	if err := tx.Create(contact).Error; err != nil {
		return nil, err
	}

	images, err := s.resolveListingImages(ctx, tx, listing.ID, params.OwnerUserID, params.Images)
	if err != nil {
		return nil, err
	}
	if len(images) > 0 {
		if err := tx.Create(&images).Error; err != nil {
			return nil, err
		}
	}

	return &listing, nil
}

// 15. createSecondhandWithCharge creates a draft and optionally charges draft-save points.
func (s *SecondhandService) createSecondhandWithCharge(ctx context.Context, params UpsertSecondhandParams) (string, *PointsChargeResponse, error) {
	if !params.ChargeDraftSave {
		listingID, err := s.upsertSecondhand(ctx, params, true)
		return listingID, nil, err
	}
	if err := s.validateUpsertParams(params); err != nil {
		return "", nil, err
	}

	communityID, err := s.resolveCommunityID(ctx, params.OwnerUserID, params.CommunityID, params.CommunityName, params.VisibilityScope)
	if err != nil {
		return "", nil, err
	}

	var returnPublicID string
	var charge *PointsChargeResponse
	err = s.runtime.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		listing, err := s.createSecondhandAggregate(ctx, tx, params, communityID)
		if err != nil {
			return err
		}
		chargeResult, err := s.chargeListingAction(ctx, tx, listing, WalletActionSaveDraft)
		if err != nil {
			return err
		}
		returnPublicID = listing.PublicID
		charge = chargeResult
		return nil
	})
	if err != nil {
		return "", nil, err
	}

	return returnPublicID, charge, nil
}

// 16. updateSecondhandWithCharge updates a listing and charges edit points.
func (s *SecondhandService) updateSecondhandWithCharge(ctx context.Context, params UpsertSecondhandParams) (string, *PointsChargeResponse, error) {
	if err := s.validateUpsertParams(params); err != nil {
		return "", nil, err
	}

	communityID, err := s.resolveCommunityID(ctx, params.OwnerUserID, params.CommunityID, params.CommunityName, params.VisibilityScope)
	if err != nil {
		return "", nil, err
	}

	var returnPublicID string
	var charge *PointsChargeResponse
	err = s.runtime.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		listing, secondhand, _, err := s.loadOwnedListingWithTx(ctx, tx, params.OwnerUserID, params.ListingPublicID)
		if err != nil {
			return err
		}
		returnPublicID = listing.PublicID

		if shouldChargeSecondhandUpdate(listing, params.ChargeDraftSave) {
			action := secondhandUpdateChargeAction(listing, params.ChargeDraftSave)
			chargeResult, err := s.chargeListingAction(ctx, tx, listing, action)
			if err != nil {
				return err
			}
			charge = chargeResult
		}

		requestedBusinessStatus := strings.TrimSpace(params.BusinessStatus)
		listing.Title = params.Title
		listing.Summary = params.Summary
		listing.Description = params.Description
		listing.DistrictCode = params.DistrictCode
		listing.CommunityID = communityID
		listing.PublisherIdentityType = s.fallbackPublisherIdentity(params.PublisherIdentityType)
		if requestedBusinessStatus != "" {
			listing.BusinessStatus = requestedBusinessStatus
		}
		if err := tx.Save(listing).Error; err != nil {
			return err
		}
		if requestedBusinessStatus != "" {
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
		if err := tx.Save(secondhand).Error; err != nil {
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
		return "", nil, err
	}

	return returnPublicID, charge, nil
}

// 17. publishSecondhandWithCharge publishes or republishes and charges points.
func (s *SecondhandService) publishSecondhandWithCharge(ctx context.Context, ownerUserID int64, listingPublicID string, action string) (*SecondhandListingDetail, error) {
	var listingPublicIDResult string
	var charge *PointsChargeResponse
	err := s.runtime.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		listing, _, _, err := s.loadOwnedListingWithTx(ctx, tx, ownerUserID, listingPublicID)
		if err != nil {
			return err
		}
		if action == WalletActionPublish && listing.PublicationStatus != "draft" {
			return errcode.New(errcode.CodeValidationError, "only draft listings can be published")
		}
		if action == WalletActionRepublish && listing.PublicationStatus != "expired" {
			return errcode.New(errcode.CodeValidationError, "only expired listings can be republished")
		}
		if err := s.validateListingReadyWithTx(ctx, tx, ownerUserID, listing.ID); err != nil {
			return err
		}
		chargeResult, err := s.chargeListingAction(ctx, tx, listing, action)
		if err != nil {
			return err
		}
		charge = chargeResult

		now := s.runtime.Now()
		expireAt := now.Add(14 * 24 * time.Hour)
		updates := map[string]any{
			"publication_status": "active",
			"sort_refreshed_at":  now,
			"expire_at":          expireAt,
		}
		if action == WalletActionPublish {
			updates["published_at"] = now
		}
		if action == WalletActionRepublish {
			updates["business_status"] = "available"
		}
		if err := tx.Model(&model.Listing{}).Where("id = ?", listing.ID).Updates(updates).Error; err != nil {
			if action == WalletActionPublish {
				return errcode.New(errcode.CodeInternalError, "failed to publish listing")
			}
			return errcode.New(errcode.CodeInternalError, "failed to republish listing")
		}
		listingPublicIDResult = listing.PublicID
		return nil
	})
	if err != nil {
		return nil, err
	}

	result, err := s.GetSecondhandDetail(ctx, listingPublicIDResult, &ownerUserID, nil)
	if err != nil {
		return nil, err
	}
	attachSecondhandCharge(result, charge)
	return result, nil
}

// 18. shouldChargeSecondhandUpdate returns whether a listing update should charge points.
func shouldChargeSecondhandUpdate(listing *model.Listing, chargeDraftSave bool) bool {
	if shouldChargeListingEdit(listing) {
		return true
	}
	return chargeDraftSave
}

// 19. secondhandUpdateChargeAction returns the wallet action for one update.
func secondhandUpdateChargeAction(listing *model.Listing, chargeDraftSave bool) string {
	if shouldChargeListingEdit(listing) {
		return WalletActionEdit
	}
	if chargeDraftSave {
		return WalletActionSaveDraft
	}
	return ""
}

// 20. chargeListingAction charges the configured secondhand action cost.
func (s *SecondhandService) chargeListingAction(ctx context.Context, tx *gorm.DB, listing *model.Listing, action string) (*PointsChargeResponse, error) {
	if s.runtime.WalletService == nil {
		return nil, errcode.New(errcode.CodeInternalError, "wallet service is not configured")
	}
	amount := ListingActionCost("secondhand", action)
	return s.runtime.WalletService.SpendPointsWithTx(ctx, tx, WalletSpendParams{
		UserID:         listing.OwnerUserID,
		Amount:         amount,
		BizModule:      "secondhand",
		ActionType:     action,
		ListingID:      &listing.ID,
		IdempotencyKey: fmt.Sprintf("listing:%s:%s:%d", listing.PublicID, action, s.runtime.Now().UnixNano()),
		Note:           "secondhand listing " + action,
	})
}

// 21. attachSecondhandCharge attaches wallet charge metadata to a listing detail.
func attachSecondhandCharge(detail *SecondhandListingDetail, charge *PointsChargeResponse) {
	if detail == nil || charge == nil {
		return
	}
	detail.PointsCharged = charge.PointsCharged
	detail.PointsBalanceAfter = &charge.PointsBalanceAfter
	detail.PointsTransactionID = charge.PointsTransactionID
}

// 22. validateUpsertParams validates create and update payloads.
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

// 23. applyPublicFilters applies search and visibility filters to the list query.
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
	if filters.HasMedia {
		query = query.Where("EXISTS (SELECT 1 FROM listing_images WHERE listing_images.listing_id = listings.id)")
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

// 24. fallbackPublisherIdentity falls back to the stored profile identity when missing.
func (s *SecondhandService) fallbackPublisherIdentity(identity string) string {
	if strings.TrimSpace(identity) == "" {
		return "owner"
	}
	return strings.TrimSpace(identity)
}

// 25. isAllowedSecondhandValue validates a normalized enum-like input.
func isAllowedSecondhandValue(value string, allowedValues []string) bool {
	normalizedValue := strings.TrimSpace(value)
	for _, allowedValue := range allowedValues {
		if normalizedValue == allowedValue {
			return true
		}
	}

	return false
}

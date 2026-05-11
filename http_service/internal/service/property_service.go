/*
 * Property channel business logic.
 * 1. Manage property sale and serviced apartment draft, publish, and owner flows.
 * 2. Reuse listing lifecycle, media, contact, and visibility rules across property modules.
 */
package service

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/model"
	"ajoliving_web/http_service/internal/utils"
)

const (
	propertySaleTTL      = 30 * 24 * time.Hour
	servicedApartmentTTL = 28 * 24 * time.Hour
)

// 1. PropertyService handles sale and serviced apartment workflows.
type PropertyService struct {
	runtime *Runtime
}

// 2. propertyListingRow defines common list query fields.
type propertyListingRow struct {
	model.Listing
	SalePropertyType             string
	SaleEstateName               string
	SaleAddressText              string
	SaleAskingPriceHKD           float64
	SaleUsableAreaSqft           int
	SaleGrossAreaSqft            *int
	SaleBedroomCount             int
	SaleLivingRoomCount          int
	SaleBathroomCount            int
	SaleFloorLevel               string
	SaleDirection                string
	SaleBuildingAge              string
	SaleFeatureTags              []byte
	SaleContactMethod            string
	SalePublisherRoleLabel       string
	ServicedProjectName          string
	ServicedAddressText          string
	ServicedLowestMonthlyRentHKD float64
	ServicedMinLeaseMonths       int
	ServicedFacilityTags         []byte
	ServicedServiceTags          []byte
	ServicedRoomTypes            []byte
	ServicedContactMethod        string
	ServicedPublisherRoleLabel   string
}

// 3. NewPropertyService creates a property service instance.
func NewPropertyService(runtime *Runtime) *PropertyService {
	return &PropertyService{runtime: runtime}
}

// 4. CreatePropertySale creates a property sale draft.
func (s *PropertyService) CreatePropertySale(ctx context.Context, params UpsertPropertySaleParams) (*PropertyListingDetail, error) {
	listingID, err := s.upsertPropertySale(ctx, params, true)
	if err != nil {
		return nil, err
	}

	return s.GetPropertyDetail(ctx, PropertyChannelSale, listingID, &params.OwnerUserID)
}

// 5. UpdatePropertySale updates an owned property sale listing.
func (s *PropertyService) UpdatePropertySale(ctx context.Context, params UpsertPropertySaleParams) (*PropertyListingDetail, error) {
	listingID, charge, err := s.updatePropertySaleWithCharge(ctx, params)
	if err != nil {
		return nil, err
	}

	result, err := s.GetPropertyDetail(ctx, PropertyChannelSale, listingID, &params.OwnerUserID)
	if err != nil {
		return nil, err
	}
	attachPropertyCharge(result, charge)
	return result, nil
}

// 6. CreateServicedApartment creates a serviced apartment draft.
func (s *PropertyService) CreateServicedApartment(ctx context.Context, params UpsertServicedApartmentParams) (*PropertyListingDetail, error) {
	listingID, err := s.upsertServicedApartment(ctx, params, true)
	if err != nil {
		return nil, err
	}

	return s.GetPropertyDetail(ctx, PropertyChannelServiced, listingID, &params.OwnerUserID)
}

// 7. UpdateServicedApartment updates an owned serviced apartment listing.
func (s *PropertyService) UpdateServicedApartment(ctx context.Context, params UpsertServicedApartmentParams) (*PropertyListingDetail, error) {
	listingID, charge, err := s.updateServicedApartmentWithCharge(ctx, params)
	if err != nil {
		return nil, err
	}

	result, err := s.GetPropertyDetail(ctx, PropertyChannelServiced, listingID, &params.OwnerUserID)
	if err != nil {
		return nil, err
	}
	attachPropertyCharge(result, charge)
	return result, nil
}

// 8. PublishProperty publishes a draft property listing.
func (s *PropertyService) PublishProperty(ctx context.Context, channel PropertyChannel, ownerUserID int64, listingPublicID string) (*PropertyListingDetail, error) {
	return s.publishPropertyWithCharge(ctx, channel, ownerUserID, listingPublicID, WalletActionPublish)
}

// 9. RepublishProperty republishes an expired property listing.
func (s *PropertyService) RepublishProperty(ctx context.Context, channel PropertyChannel, ownerUserID int64, listingPublicID string) (*PropertyListingDetail, error) {
	return s.publishPropertyWithCharge(ctx, channel, ownerUserID, listingPublicID, WalletActionRepublish)
}

// 10. MarkPropertySold marks a sale listing as sold.
func (s *PropertyService) MarkPropertySold(ctx context.Context, channel PropertyChannel, ownerUserID int64, listingPublicID string) error {
	listing, _, err := s.loadOwnedPropertyListing(ctx, channel, ownerUserID, listingPublicID)
	if err != nil {
		return err
	}

	return s.runtime.DB.WithContext(ctx).Model(&model.Listing{}).
		Where("id = ?", listing.ID).
		Update("business_status", "sold").Error
}

// 11. DeactivateProperty hides a property listing.
func (s *PropertyService) DeactivateProperty(ctx context.Context, channel PropertyChannel, ownerUserID int64, listingPublicID string) error {
	listing, _, err := s.loadOwnedPropertyListing(ctx, channel, ownerUserID, listingPublicID)
	if err != nil {
		return err
	}

	return s.runtime.DB.WithContext(ctx).Model(&model.Listing{}).
		Where("id = ?", listing.ID).
		Update("publication_status", "hidden").Error
}

// 12. ListPublicProperties returns public listings for a property channel.
func (s *PropertyService) ListPublicProperties(ctx context.Context, channel PropertyChannel, filters PropertyListFilters) ([]PropertyListingSummary, *model.Pagination, error) {
	page, pageSize := normalizePagination(filters.Page, filters.PageSize)
	baseQuery := s.basePropertyListQuery(ctx, channel).
		Where("listings.module = ? AND listings.publication_status = ? AND listings.moderation_status = ? AND listings.business_status = ? AND listings.is_deleted = ?", string(channel), "active", "approved", "available", false)
	baseQuery = s.applyPropertyFilters(baseQuery, channel, filters)

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, nil, errcode.New(errcode.CodeInternalError, "failed to count listings")
	}

	query := baseQuery
	switch filters.SortBy {
	case "price_asc":
		query = query.Order(propertyPriceColumn(channel) + " asc")
	case "price_desc":
		query = query.Order(propertyPriceColumn(channel) + " desc")
	default:
		query = query.Order("listings.sort_refreshed_at desc NULLS LAST, listings.created_at desc")
	}

	var rows []propertyListingRow
	if err := query.Offset((page - 1) * pageSize).Limit(pageSize).Scan(&rows).Error; err != nil {
		return nil, nil, errcode.New(errcode.CodeInternalError, "failed to load listings")
	}

	items, err := s.buildPropertySummaries(ctx, channel, rows)
	if err != nil {
		return nil, nil, err
	}

	return items, &model.Pagination{Page: page, PageSize: pageSize, Total: total}, nil
}

// 13. ListMyProperties returns owner listings for a property channel.
func (s *PropertyService) ListMyProperties(ctx context.Context, channel PropertyChannel, ownerUserID int64, filters PropertyListFilters) ([]PropertyListingSummary, *model.Pagination, error) {
	page, pageSize := normalizePagination(filters.Page, filters.PageSize)
	baseQuery := s.basePropertyListQuery(ctx, channel).
		Where("listings.module = ? AND listings.owner_user_id = ? AND listings.is_deleted = ?", string(channel), ownerUserID, false)
	baseQuery = s.applyPropertyFilters(baseQuery, channel, filters)
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

	var rows []propertyListingRow
	if err := baseQuery.Order("listings.updated_at desc").Offset((page - 1) * pageSize).Limit(pageSize).Scan(&rows).Error; err != nil {
		return nil, nil, errcode.New(errcode.CodeInternalError, "failed to load owner listings")
	}

	items, err := s.buildPropertySummaries(ctx, channel, rows)
	if err != nil {
		return nil, nil, err
	}

	return items, &model.Pagination{Page: page, PageSize: pageSize, Total: total}, nil
}

// 14. GetPropertyDetail returns one property detail payload.
func (s *PropertyService) GetPropertyDetail(ctx context.Context, channel PropertyChannel, listingPublicID string, viewerUserID *int64) (*PropertyListingDetail, error) {
	listing, contact, err := s.loadPropertyListingByPublicID(ctx, channel, listingPublicID)
	if err != nil {
		return nil, err
	}
	if listing.PublicationStatus == "hidden" && (viewerUserID == nil || *viewerUserID != listing.OwnerUserID) {
		return nil, errcode.New(errcode.CodeHidden, "listing is hidden")
	}
	if viewerUserID == nil || *viewerUserID != listing.OwnerUserID {
		if listing.PublicationStatus != "active" || listing.BusinessStatus != "available" {
			return nil, errcode.New(errcode.CodeNotFound, "listing not found")
		}
	}

	rows, err := s.loadPropertyRowsByIDs(ctx, channel, []int64{listing.ID})
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, errcode.New(errcode.CodeNotFound, "listing not found")
	}

	summaries, err := s.buildPropertySummaries(ctx, channel, rows)
	if err != nil {
		return nil, err
	}

	images, err := s.loadListingImages(ctx, []int64{listing.ID})
	if err != nil {
		return nil, err
	}

	return &PropertyListingDetail{
		PropertyListingSummary: summaries[0],
		Description:            listing.Description,
		Images:                 images[listing.ID],
		ContactSummary: ListingContactSummary{
			ShowPhone:    contact.ShowPhone,
			ShowWhatsApp: contact.ShowWhatsApp,
			ShowChat:     contact.ShowChat,
			ShowInquiry:  contact.ShowInquiryForm,
		},
	}, nil
}

// 15. publishPropertyWithCharge publishes or republishes and charges points.
func (s *PropertyService) publishPropertyWithCharge(ctx context.Context, channel PropertyChannel, ownerUserID int64, listingPublicID string, action string) (*PropertyListingDetail, error) {
	var returnPublicID string
	var charge *PointsChargeResponse
	err := s.runtime.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		listing, _, err := s.loadOwnedPropertyListingWithTx(ctx, tx, channel, ownerUserID, listingPublicID)
		if err != nil {
			return err
		}
		if action == WalletActionPublish && listing.PublicationStatus != "draft" {
			return errcode.New(errcode.CodeValidationError, "only draft listings can be published")
		}
		if action == WalletActionRepublish && listing.PublicationStatus != "expired" {
			return errcode.New(errcode.CodeValidationError, "only expired listings can be republished")
		}
		if err := s.validatePropertyReadyWithTx(ctx, tx, listing.ID); err != nil {
			return err
		}
		chargeResult, err := s.chargePropertyAction(ctx, tx, channel, listing, action)
		if err != nil {
			return err
		}
		charge = chargeResult

		now := s.runtime.Now()
		expireAt := now.Add(propertyChannelTTL(channel))
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
		returnPublicID = listing.PublicID
		return nil
	})
	if err != nil {
		return nil, err
	}

	result, err := s.GetPropertyDetail(ctx, channel, returnPublicID, &ownerUserID)
	if err != nil {
		return nil, err
	}
	attachPropertyCharge(result, charge)
	return result, nil
}

// 16. chargePropertyAction charges the configured property channel action cost.
func (s *PropertyService) chargePropertyAction(ctx context.Context, tx *gorm.DB, channel PropertyChannel, listing *model.Listing, action string) (*PointsChargeResponse, error) {
	if s.runtime.WalletService == nil {
		return nil, errcode.New(errcode.CodeInternalError, "wallet service is not configured")
	}
	module := string(channel)
	return s.runtime.WalletService.SpendPointsWithTx(ctx, tx, WalletSpendParams{
		UserID:         listing.OwnerUserID,
		Amount:         ListingActionCost(module, action),
		BizModule:      module,
		ActionType:     action,
		ListingID:      &listing.ID,
		IdempotencyKey: fmt.Sprintf("listing:%s:%s:%d", listing.PublicID, action, s.runtime.Now().UnixNano()),
		Note:           module + " listing " + action,
	})
}

// 17. attachPropertyCharge attaches wallet charge metadata to a property detail.
func attachPropertyCharge(detail *PropertyListingDetail, charge *PointsChargeResponse) {
	if detail == nil || charge == nil {
		return
	}
	detail.PointsCharged = charge.PointsCharged
	detail.PointsBalanceAfter = &charge.PointsBalanceAfter
	detail.PointsTransactionID = charge.PointsTransactionID
}

// 18. GrantPropertyContactAccess returns allowed contact payload for logged-in users.
func (s *PropertyService) GrantPropertyContactAccess(ctx context.Context, channel PropertyChannel, userID int64, listingPublicID string, requestIP string, userAgent string) (*ContactAccessResult, error) {
	listing, contact, err := s.loadPropertyListingByPublicID(ctx, channel, listingPublicID)
	if err != nil {
		return nil, err
	}
	if listing.PublicationStatus != "active" {
		return nil, errcode.New(errcode.CodeExpired, "listing is not active")
	}
	if listing.BusinessStatus != "available" {
		return nil, errcode.New(errcode.CodeAuthForbidden, "listing is not available")
	}

	payload := map[string]string{}
	channels := map[string]bool{
		"phone":        false,
		"whatsapp":     false,
		"chat":         contact.ShowChat,
		"inquiry_form": contact.ShowInquiryForm,
	}

	if contact.ShowPhone && contact.PhoneEncrypted != "" {
		phone, err := utils.DecryptString(s.runtime.Config.EncryptionKey, contact.PhoneEncrypted)
		if err != nil {
			return nil, errcode.New(errcode.CodeInternalError, "failed to decrypt phone")
		}
		payload["phone"] = phone
		channels["phone"] = true
	}

	if contact.ShowWhatsApp && contact.WhatsAppEncrypted != "" {
		whatsApp, err := utils.DecryptString(s.runtime.Config.EncryptionKey, contact.WhatsAppEncrypted)
		if err != nil {
			return nil, errcode.New(errcode.CodeInternalError, "failed to decrypt whatsapp")
		}
		payload["whatsapp_url"] = buildPropertyWhatsAppURL(whatsApp, listing.Title, listing.PublicID, s.runtime.Config.AppPublicBaseURL, channel)
		channels["whatsapp"] = true
	}

	grantedChannels, err := marshalJSON(channels)
	if err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to encode access audit")
	}

	if err := s.runtime.DB.WithContext(ctx).Create(&model.ContactAccessLog{
		ListingID:       listing.ID,
		RequestUserID:   userID,
		GrantedChannels: grantedChannels,
		RequestIP:       requestIP,
		UserAgent:       userAgent,
	}).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to store contact access log")
	}

	return &ContactAccessResult{
		ListingID:       listing.PublicID,
		AllowedChannels: channels,
		ContactPayload:  payload,
	}, nil
}

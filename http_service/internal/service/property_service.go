/*
 * Property channel business logic.
 * 1. Manage property sale and serviced apartment draft, publish, and owner flows.
 * 2. Reuse listing lifecycle, media, contact, and visibility rules across property modules.
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
	SalePropertyNo                string
	SaleTransactionType           string
	SaleLocationScope             string
	SaleListingCategory           string
	SaleMultiUnitProject          bool
	SalePropertyType              string
	SaleRentalType                string
	SalePropertyAttributes        []byte
	SaleRenovationType            string
	SaleAgencyCompanyName         string
	SaleEstateName                string
	SaleAddressText               string
	SaleAddressTextEn             string
	SaleBlockName                 string
	SaleUnitName                  string
	SaleShowUnit                  bool
	SaleLatitude                  *float64
	SaleLongitude                 *float64
	SaleAskingPriceHKD            float64
	SaleMonthlyRentHKD            float64
	SalePriceReferenceOnly        bool
	SalePriceNegotiable           bool
	SaleAnnualPrepayDiscount      bool
	SaleAnnualPrepayOption        string
	SaleLeaseStartDate            string
	SaleRentIncluded              string
	SaleAreaMode                  string
	SaleUsableAreaSqft            int
	SaleGrossAreaSqft             *int
	SaleBedroomCount              int
	SaleLivingRoomCount           int
	SaleBathroomCount             int
	SaleFloorLevel                string
	SaleFloorRaw                  string
	SaleFloorZone                 string
	SaleFloorDisplayRange         string
	SaleTotalFloors               int
	SalePublicLocationText        string
	SaleDirection                 string
	SaleBuildingAge               string
	SaleCompletionYear            int
	SaleBuildingTotalFloors       int
	SaleManagementCompany         string
	SaleKitchenType               string
	SaleCookingMode               string
	SaleManagementFeeHKD          float64
	SaleVideoURL                  string
	SaleVRURL                     string
	SalePrivateNote               string
	SaleTitleEn                   string
	SaleDescriptionEn             string
	SaleAdPackageCode             string
	SaleAdWeight                  int
	SaleAdPriceHKD                float64
	SaleAdPricePoints             int64
	SaleAdDurationDays            int
	SaleAdExpiresAt               *time.Time
	SaleFeatureTags               []byte
	SaleContactMethod             string
	SalePublisherRoleLabel        string
	SaleViewCount                 int64
	SaleInquiryCount              int64
	ServicedProjectName           string
	ServicedProjectNameEn         string
	ServicedProjectAttributes     []byte
	ServicedAddressText           string
	ServicedAddressTextEn         string
	ServicedWebsiteURL            string
	ServicedWhatsApp              string
	ServicedFax                   string
	ServicedDescriptionEn         string
	ServicedServiceIntro          string
	ServicedServiceIntroEn        string
	ServicedBenefitsText          string
	ServicedBenefitsTextEn        string
	ServicedExtraChargesText      string
	ServicedExtraChargesTextEn    string
	ServicedLowestMonthlyRentHKD  float64
	ServicedHighestMonthlyRentHKD float64
	ServicedLowestDailyRentHKD    float64
	ServicedPriceReferenceOnly    bool
	ServicedPriceNegotiable       bool
	ServicedMinUsableAreaSqft     int
	ServicedMaxUsableAreaSqft     int
	ServicedMinLeaseMonths        int
	ServicedMinStayValue          int
	ServicedMinStayUnit           string
	ServicedLocationScope         string
	ServicedListingCategory       string
	ServicedMultiUnitProject      bool
	ServicedFacilityTags          []byte
	ServicedServiceTags           []byte
	ServicedRoomTypes             []byte
	ServicedAdPackageCode         string
	ServicedAdWeight              int
	ServicedAdPriceHKD            float64
	ServicedAdPricePoints         int64
	ServicedAdDurationDays        int
	ServicedAdExpiresAt           *time.Time
	ServicedContactMethod         string
	ServicedPublisherRoleLabel    string
}

// 3. NewPropertyService creates a property service instance.
func NewPropertyService(runtime *Runtime) *PropertyService {
	return &PropertyService{runtime: runtime}
}

// 4. CreatePropertySale creates a property sale draft.
func (s *PropertyService) CreatePropertySale(ctx context.Context, params UpsertPropertySaleParams) (*PropertyListingDetail, error) {
	publisher, err := s.resolvePropertySalePublisher(ctx, s.runtime.DB, params.OwnerUserID, params.PublisherIdentityType, "property_publish")
	if err != nil {
		return nil, err
	}
	if err := validateAgencyPublisherLocation(publisher, params.LocationScope); err != nil {
		return nil, err
	}
	applyPropertySalePublisher(&params, publisher)

	listingID, charge, err := s.upsertPropertySale(ctx, params, true, params.ChargeDraftSave)
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

// 5. UpdatePropertySale updates an owned property sale listing.
func (s *PropertyService) UpdatePropertySale(ctx context.Context, params UpsertPropertySaleParams) (*PropertyListingDetail, error) {
	publisher, err := s.resolvePropertySalePublisher(ctx, s.runtime.DB, params.OwnerUserID, params.PublisherIdentityType, "property_manage")
	if err != nil {
		return nil, err
	}
	if err := validateAgencyPublisherLocation(publisher, params.LocationScope); err != nil {
		return nil, err
	}
	applyPropertySalePublisher(&params, publisher)

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
	if err := s.requireSubaccountPropertyPermission(ctx, s.runtime.DB, params.OwnerUserID, "property_publish"); err != nil {
		return nil, err
	}
	listingID, err := s.upsertServicedApartment(ctx, params, true)
	if err != nil {
		return nil, err
	}

	return s.GetPropertyDetail(ctx, PropertyChannelServiced, listingID, &params.OwnerUserID)
}

// 7. UpdateServicedApartment updates an owned serviced apartment listing.
func (s *PropertyService) UpdateServicedApartment(ctx context.Context, params UpsertServicedApartmentParams) (*PropertyListingDetail, error) {
	if err := s.requireSubaccountPropertyPermission(ctx, s.runtime.DB, params.OwnerUserID, "property_manage"); err != nil {
		return nil, err
	}
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
	if err := s.requireSubaccountPropertyPermission(ctx, s.runtime.DB, ownerUserID, "property_publish"); err != nil {
		return nil, err
	}
	return s.publishPropertyWithCharge(ctx, channel, ownerUserID, listingPublicID, WalletActionPublish)
}

// 9. RepublishProperty republishes an expired or hidden property listing.
func (s *PropertyService) RepublishProperty(ctx context.Context, channel PropertyChannel, ownerUserID int64, listingPublicID string) (*PropertyListingDetail, error) {
	if err := s.requireSubaccountPropertyPermission(ctx, s.runtime.DB, ownerUserID, "property_publish"); err != nil {
		return nil, err
	}
	return s.publishPropertyWithCharge(ctx, channel, ownerUserID, listingPublicID, WalletActionRepublish)
}

// 10. MarkPropertySold marks a sale listing as sold.
func (s *PropertyService) MarkPropertySold(ctx context.Context, channel PropertyChannel, ownerUserID int64, listingPublicID string) error {
	if err := s.requireSubaccountPropertyPermission(ctx, s.runtime.DB, ownerUserID, "property_manage"); err != nil {
		return err
	}
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
	if err := s.requireSubaccountPropertyPermission(ctx, s.runtime.DB, ownerUserID, "property_manage"); err != nil {
		return err
	}
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
		query = query.Order(propertyDefaultSortPrefix(channel) + propertyPriceColumn(channel, filters.PriceMode) + " asc")
	case "price_desc":
		query = query.Order(propertyDefaultSortPrefix(channel) + propertyPriceColumn(channel, filters.PriceMode) + " desc")
	default:
		query = query.Order(propertyDefaultSortPrefix(channel) + "listings.sort_refreshed_at desc NULLS LAST, listings.created_at desc")
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
	ownerUserIDs, err := s.companyManagedPropertyOwnerIDs(ctx, ownerUserID)
	if err != nil {
		return nil, nil, err
	}
	baseQuery := s.basePropertyListQuery(ctx, channel).
		Where("listings.module = ? AND listings.owner_user_id IN ? AND listings.is_deleted = ?", string(channel), ownerUserIDs, false)
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

// 13.1 companyManagedPropertyOwnerIDs includes linked children for approved company owners.
func (s *PropertyService) companyManagedPropertyOwnerIDs(ctx context.Context, userID int64) ([]int64, error) {
	result := []int64{userID}
	if err := NewAgencyCompanyService(s.runtime).requireApprovedCompanyOwner(ctx, s.runtime.DB, userID); err != nil {
		return result, nil
	}
	var links []model.AgencyCompanySubaccount
	if err := s.runtime.DB.WithContext(ctx).Where("company_owner_user_id = ?", userID).Find(&links).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load company listing members")
	}
	for _, link := range links {
		result = append(result, link.ChildUserID)
	}
	return result, nil
}

// 14. GetPropertyDetail returns one property detail payload.
func (s *PropertyService) GetPropertyDetail(ctx context.Context, channel PropertyChannel, listingPublicID string, viewerUserID *int64) (*PropertyListingDetail, error) {
	listing, contact, err := s.loadPropertyListingByPublicID(ctx, channel, listingPublicID)
	if err != nil {
		return nil, err
	}
	viewerCanManage := viewerUserID != nil && *viewerUserID == listing.OwnerUserID
	if viewerUserID != nil && !viewerCanManage {
		viewerCanManage, err = s.canCompanyOwnerManageChildListing(ctx, s.runtime.DB, *viewerUserID, listing.OwnerUserID)
		if err != nil {
			return nil, err
		}
	}
	if listing.PublicationStatus == "hidden" && !viewerCanManage {
		return nil, errcode.New(errcode.CodeHidden, "listing is hidden")
	}
	if !viewerCanManage {
		if listing.PublicationStatus != "active" || listing.ModerationStatus != "approved" || listing.BusinessStatus != "available" {
			return nil, errcode.New(errcode.CodeNotFound, "listing not found")
		}
		if channel == PropertyChannelSale {
			_ = s.incrementPropertySaleCounter(ctx, listing.ID, "view_count")
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
	if channel == PropertyChannelSale && len(summaries) > 0 && summaries[0].PropertySale != nil && viewerCanManage {
		summaries[0].PropertySale.UnitName = rows[0].SaleUnitName
		summaries[0].PropertySale.FloorRaw = rows[0].SaleFloorRaw
		summaries[0].PropertySale.PrivateNote = rows[0].SalePrivateNote
		summaries[0].PropertySale.PropertyAttributes = decodeStringMapBytes(rows[0].SalePropertyAttributes)
	}
	if channel == PropertyChannelSale && viewerUserID != nil && len(summaries) > 0 {
		summaries[0].IsFavorite = s.isFavoriteProperty(ctx, *viewerUserID, listing.ID)
	}

	images, err := s.loadListingImages(ctx, []int64{listing.ID})
	if err != nil {
		return nil, err
	}

	contactSummary := ListingContactSummary{
		ShowPhone:    contact.ShowPhone,
		ShowWhatsApp: contact.ShowWhatsApp,
		ShowChat:     contact.ShowChat,
		ShowInquiry:  contact.ShowInquiryForm,
	}
	if viewerCanManage {
		contactSummary.ContactAttributes = decodeStringMapBytes(contact.ContactAttributes)
	}

	return &PropertyListingDetail{
		PropertyListingSummary: summaries[0],
		Description:            listing.Description,
		Images:                 images[listing.ID],
		ContactSummary:         contactSummary,
	}, nil
}

// 15. publishPropertyWithCharge publishes or republishes and charges points.
func (s *PropertyService) publishPropertyWithCharge(ctx context.Context, channel PropertyChannel, ownerUserID int64, listingPublicID string, action string) (*PropertyListingDetail, error) {
	var returnPublicID string
	var charge *PointsChargeResponse
	err := s.runtime.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		listing, contact, err := s.loadOwnedPropertyListingWithTx(ctx, tx, channel, ownerUserID, listingPublicID)
		if err != nil {
			return err
		}
		if action == WalletActionPublish && listing.PublicationStatus != "draft" {
			return errcode.New(errcode.CodeValidationError, "only draft listings can be published")
		}
		if action == WalletActionRepublish && listing.PublicationStatus != "expired" && listing.PublicationStatus != "hidden" {
			return errcode.New(errcode.CodeValidationError, "only expired or hidden listings can be republished")
		}
		if channel == PropertyChannelSale {
			publisher, err := s.resolvePropertySalePublisher(ctx, tx, ownerUserID, listing.PublisherIdentityType, "property_publish")
			if err != nil {
				return err
			}
			var sale model.PropertySaleListing
			if err := tx.Where("listing_id = ?", listing.ID).First(&sale).Error; err != nil {
				return errcode.New(errcode.CodeInternalError, "failed to load property sale publisher scope")
			}
			if err := validateAgencyPublisherLocation(publisher, sale.LocationScope); err != nil {
				return err
			}
			if listing.PublisherIdentityType != publisher.Identity {
				if err := tx.Model(&model.Listing{}).Where("id = ?", listing.ID).Update("publisher_identity_type", publisher.Identity).Error; err != nil {
					return errcode.New(errcode.CodeInternalError, "failed to update property publisher identity")
				}
				if err := tx.Model(&model.PropertySaleListing{}).Where("listing_id = ?", listing.ID).Update("publisher_role_label", publisherRoleLabel(publisher.Identity)).Error; err != nil {
					return errcode.New(errcode.CodeInternalError, "failed to update property publisher role")
				}
				listing.PublisherIdentityType = publisher.Identity
			}
			if err := tx.Model(&model.PropertySaleListing{}).Where("listing_id = ?", listing.ID).Update("agency_company_name", publisher.AgencyCompanyName).Error; err != nil {
				return errcode.New(errcode.CodeInternalError, "failed to update property agency company")
			}
			if publisher.Identity == "agent" {
				params := UpsertPropertySaleParams{Contact: PropertyContactInput{ContactAttributes: decodeStringMapBytes(contact.ContactAttributes), ShowPhone: contact.ShowPhone, ShowWhatsApp: contact.ShowWhatsApp, ShowChat: contact.ShowChat, ShowInquiryForm: contact.ShowInquiryForm}}
				applyPropertySalePublisher(&params, publisher)
				nextContact, err := s.buildPropertyContact(listing.ID, params.Contact, contact.ContactMode)
				if err != nil {
					return err
				}
				if err := tx.Save(nextContact).Error; err != nil {
					return errcode.New(errcode.CodeInternalError, "failed to update approved agency contact snapshot")
				}
				contact = nextContact
			}
			if err := s.validatePropertySalePublicationWithTx(ctx, tx, listing, contact); err != nil {
				return err
			}
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
		if channel == PropertyChannelSale {
			var sale model.PropertySaleListing
			if err := tx.WithContext(ctx).Where("listing_id = ?", listing.ID).First(&sale).Error; err != nil {
				return errcode.New(errcode.CodeInternalError, "failed to load property sale listing")
			}
			durationDays := sale.AdDurationDays
			if durationDays <= 0 {
				durationDays = propertyAdPackage(sale.AdPackageCode).DurationDays
			}
			expireAt = now.Add(time.Duration(durationDays) * 24 * time.Hour)
			updates["expire_at"] = expireAt
			if err := tx.Model(&model.PropertySaleListing{}).Where("listing_id = ?", listing.ID).Update("ad_expires_at", expireAt).Error; err != nil {
				return errcode.New(errcode.CodeInternalError, "failed to update property ad expiry")
			}
		}
		if channel == PropertyChannelServiced {
			var serviced model.ServicedApartmentProject
			if err := tx.WithContext(ctx).Where("listing_id = ?", listing.ID).First(&serviced).Error; err != nil {
				return errcode.New(errcode.CodeInternalError, "failed to load serviced apartment listing")
			}
			durationDays := serviced.AdDurationDays
			if durationDays <= 0 {
				durationDays = servicedApartmentAdPackage(serviced.AdPackageCode).DurationDays
			}
			expireAt = now.Add(time.Duration(durationDays) * 24 * time.Hour)
			updates["expire_at"] = expireAt
			if err := tx.Model(&model.ServicedApartmentProject{}).Where("listing_id = ?", listing.ID).Update("ad_expires_at", expireAt).Error; err != nil {
				return errcode.New(errcode.CodeInternalError, "failed to update serviced apartment ad expiry")
			}
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
	amount := ListingActionCost(module, action)
	if channel == PropertyChannelSale && (action == WalletActionPublish || action == WalletActionRepublish) {
		var sale model.PropertySaleListing
		if err := tx.WithContext(ctx).Where("listing_id = ?", listing.ID).First(&sale).Error; err != nil {
			return nil, errcode.New(errcode.CodeInternalError, "failed to load property sale listing")
		}
		amount = propertyAdPackage(sale.AdPackageCode).PricePoints
	}
	if channel == PropertyChannelServiced && (action == WalletActionPublish || action == WalletActionRepublish) {
		var serviced model.ServicedApartmentProject
		if err := tx.WithContext(ctx).Where("listing_id = ?", listing.ID).First(&serviced).Error; err != nil {
			return nil, errcode.New(errcode.CodeInternalError, "failed to load serviced apartment listing")
		}
		amount = servicedApartmentAdPackage(serviced.AdPackageCode).PricePoints
	}
	return s.runtime.WalletService.SpendPointsWithTx(ctx, tx, WalletSpendParams{
		UserID:         listing.OwnerUserID,
		Amount:         amount,
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
	if listing.ModerationStatus != "approved" {
		return nil, errcode.New(errcode.CodeNotFound, "listing not found")
	}
	if listing.BusinessStatus != "available" {
		return nil, errcode.New(errcode.CodeAuthForbidden, "listing is not available")
	}

	payload := map[string]string{}
	contactAttributes := decodeStringMapBytes(contact.ContactAttributes)
	phone1WhatsAppEnabled := contactAttributes["phone_whatsapp_enabled"] == "yes"
	phone2WhatsAppEnabled := contactAttributes["phone_2_whatsapp_enabled"] == "yes"
	phone2CountryCode := strings.TrimSpace(contactAttributes["phone_2_country_code"])
	if phone2CountryCode == "" {
		phone2CountryCode = contactAttributes["phone_country_code"]
	}
	channels := map[string]bool{
		"phone":        false,
		"phone_2":      false,
		"whatsapp":     false,
		"wechat":       false,
		"chat":         contact.ShowChat,
		"inquiry_form": contact.ShowInquiryForm,
	}
	if strings.TrimSpace(contact.ContactNameZH) != "" {
		payload["contact_name_zh"] = strings.TrimSpace(contact.ContactNameZH)
	}
	if strings.TrimSpace(contact.ContactNameEN) != "" {
		payload["contact_name_en"] = strings.TrimSpace(contact.ContactNameEN)
	}

	phone := ""
	if (contact.ShowPhone || phone1WhatsAppEnabled) && contact.PhoneEncrypted != "" {
		phone, err = utils.DecryptString(s.runtime.Config.EncryptionKey, contact.PhoneEncrypted)
		if err != nil {
			return nil, errcode.New(errcode.CodeInternalError, "failed to decrypt phone")
		}
		if contact.ShowPhone {
			payload["phone"] = phone
			channels["phone"] = true
		}
	}
	phone2 := ""
	if (contact.ShowPhone || phone2WhatsAppEnabled) && contact.Phone2Encrypted != "" {
		phone2, err = utils.DecryptString(s.runtime.Config.EncryptionKey, contact.Phone2Encrypted)
		if err != nil {
			return nil, errcode.New(errcode.CodeInternalError, "failed to decrypt secondary phone")
		}
		if contact.ShowPhone {
			payload["phone_2"] = phone2
			channels["phone_2"] = true
		}
	}

	if phone1WhatsAppEnabled && phone != "" {
		payload["phone_whatsapp_url"] = buildPropertyWhatsAppURL(
			propertyWhatsAppPhoneNumber(phone, contactAttributes["phone_country_code"]),
			listing.Title,
			listing.PublicID,
			s.runtime.Config.AppPublicBaseURL,
			channel,
		)
		payload["whatsapp_url"] = payload["phone_whatsapp_url"]
		channels["whatsapp"] = true
	}
	if phone2WhatsAppEnabled && phone2 != "" {
		payload["phone_2_whatsapp_url"] = buildPropertyWhatsAppURL(
			propertyWhatsAppPhoneNumber(phone2, phone2CountryCode),
			listing.Title,
			listing.PublicID,
			s.runtime.Config.AppPublicBaseURL,
			channel,
		)
		if payload["whatsapp_url"] == "" {
			payload["whatsapp_url"] = payload["phone_2_whatsapp_url"]
		}
		channels["whatsapp"] = true
	}
	if contact.ShowWhatsApp && contact.WhatsAppEncrypted != "" {
		whatsApp, err := utils.DecryptString(s.runtime.Config.EncryptionKey, contact.WhatsAppEncrypted)
		if err != nil {
			return nil, errcode.New(errcode.CodeInternalError, "failed to decrypt whatsapp")
		}
		if payload["whatsapp_url"] == "" {
			payload["whatsapp_url"] = buildPropertyWhatsAppURL(
				propertyWhatsAppPhoneNumber(whatsApp, contactAttributes["phone_country_code"]),
				listing.Title,
				listing.PublicID,
				s.runtime.Config.AppPublicBaseURL,
				channel,
			)
		}
		channels["whatsapp"] = true
	} else if phone := strings.TrimSpace(payload["phone"]); payload["whatsapp_url"] == "" && phone != "" {
		payload["whatsapp_url"] = buildPropertyWhatsAppURL(phone, listing.Title, listing.PublicID, s.runtime.Config.AppPublicBaseURL, channel)
		channels["whatsapp"] = true
	}
	if contact.WeChatEncrypted != "" {
		wechat, err := utils.DecryptString(s.runtime.Config.EncryptionKey, contact.WeChatEncrypted)
		if err != nil {
			return nil, errcode.New(errcode.CodeInternalError, "failed to decrypt wechat")
		}
		payload["wechat"] = wechat
		channels["wechat"] = true
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
	if channel == PropertyChannelSale {
		_ = s.incrementPropertySaleCounter(ctx, listing.ID, "inquiry_count")
	}

	return &ContactAccessResult{
		ListingID:       listing.PublicID,
		AllowedChannels: channels,
		ContactPayload:  payload,
	}, nil
}

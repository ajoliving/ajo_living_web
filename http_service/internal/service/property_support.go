/*
 * Property channel support logic.
 * 1. Persist property channel aggregates and extension records.
 * 2. Build response payloads, contacts, media, and filter helpers.
 */
package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"gorm.io/gorm"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/model"
	"ajoliving_web/http_service/internal/utils"
)

// 1. upsertPropertySale creates or updates a sale listing aggregate.
func (s *PropertyService) upsertPropertySale(ctx context.Context, params UpsertPropertySaleParams, creating bool) (string, error) {
	if err := s.validateSaleParams(params); err != nil {
		return "", err
	}

	communityID, err := s.resolveCommunityID(ctx, params.CommunityID)
	if err != nil {
		return "", err
	}

	returnPublicID := params.ListingPublicID
	err = s.runtime.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		listing, err := s.preparePropertyRoot(ctx, tx, PropertyChannelSale, creating, propertyRootInput{
			OwnerUserID:           params.OwnerUserID,
			ListingPublicID:       params.ListingPublicID,
			Title:                 params.Title,
			Summary:               params.Summary,
			Description:           params.Description,
			DistrictCode:          params.DistrictCode,
			CommunityID:           communityID,
			PublisherIdentityType: params.PublisherIdentityType,
			BusinessStatus:        params.BusinessStatus,
		})
		if err != nil {
			return err
		}
		returnPublicID = listing.PublicID

		featureTags, err := marshalJSON(params.FeatureTags)
		if err != nil {
			return err
		}

		sale := model.PropertySaleListing{
			ListingID:          listing.ID,
			PropertyType:       strings.TrimSpace(params.PropertyType),
			EstateName:         strings.TrimSpace(params.EstateName),
			AddressText:        strings.TrimSpace(params.AddressText),
			AskingPriceHKD:     params.AskingPriceHKD,
			UsableAreaSqft:     params.UsableAreaSqft,
			GrossAreaSqft:      params.GrossAreaSqft,
			BedroomCount:       params.BedroomCount,
			LivingRoomCount:    params.LivingRoomCount,
			BathroomCount:      params.BathroomCount,
			FloorLevel:         strings.TrimSpace(params.FloorLevel),
			Direction:          strings.TrimSpace(params.Direction),
			BuildingAge:        strings.TrimSpace(params.BuildingAge),
			FeatureTags:        featureTags,
			ContactMethod:      strings.TrimSpace(params.ContactMethod),
			PublisherRoleLabel: publisherRoleLabel(params.PublisherIdentityType),
		}
		if err := tx.Save(&sale).Error; err != nil {
			return err
		}

		return s.savePropertyContactAndImages(ctx, tx, listing.ID, params.OwnerUserID, params.Contact, params.ContactMethod, params.Images)
	})
	if err != nil {
		return "", err
	}

	return returnPublicID, nil
}

// 2. updatePropertySaleWithCharge updates a sale listing and charges edit points.
func (s *PropertyService) updatePropertySaleWithCharge(ctx context.Context, params UpsertPropertySaleParams) (string, *PointsChargeResponse, error) {
	if err := s.validateSaleParams(params); err != nil {
		return "", nil, err
	}

	communityID, err := s.resolveCommunityID(ctx, params.CommunityID)
	if err != nil {
		return "", nil, err
	}

	var returnPublicID string
	var charge *PointsChargeResponse
	err = s.runtime.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		listing, err := s.preparePropertyRoot(ctx, tx, PropertyChannelSale, false, propertyRootInput{
			OwnerUserID:           params.OwnerUserID,
			ListingPublicID:       params.ListingPublicID,
			Title:                 params.Title,
			Summary:               params.Summary,
			Description:           params.Description,
			DistrictCode:          params.DistrictCode,
			CommunityID:           communityID,
			PublisherIdentityType: params.PublisherIdentityType,
			BusinessStatus:        params.BusinessStatus,
		})
		if err != nil {
			return err
		}
		returnPublicID = listing.PublicID

		if shouldChargeListingEdit(listing) {
			chargeResult, err := s.chargePropertyAction(ctx, tx, PropertyChannelSale, listing, WalletActionEdit)
			if err != nil {
				return err
			}
			charge = chargeResult
		}

		featureTags, err := marshalJSON(params.FeatureTags)
		if err != nil {
			return err
		}

		sale := model.PropertySaleListing{
			ListingID:          listing.ID,
			PropertyType:       strings.TrimSpace(params.PropertyType),
			EstateName:         strings.TrimSpace(params.EstateName),
			AddressText:        strings.TrimSpace(params.AddressText),
			AskingPriceHKD:     params.AskingPriceHKD,
			UsableAreaSqft:     params.UsableAreaSqft,
			GrossAreaSqft:      params.GrossAreaSqft,
			BedroomCount:       params.BedroomCount,
			LivingRoomCount:    params.LivingRoomCount,
			BathroomCount:      params.BathroomCount,
			FloorLevel:         strings.TrimSpace(params.FloorLevel),
			Direction:          strings.TrimSpace(params.Direction),
			BuildingAge:        strings.TrimSpace(params.BuildingAge),
			FeatureTags:        featureTags,
			ContactMethod:      strings.TrimSpace(params.ContactMethod),
			PublisherRoleLabel: publisherRoleLabel(params.PublisherIdentityType),
		}
		if err := tx.Save(&sale).Error; err != nil {
			return err
		}

		return s.savePropertyContactAndImages(ctx, tx, listing.ID, params.OwnerUserID, params.Contact, params.ContactMethod, params.Images)
	})
	if err != nil {
		return "", nil, err
	}

	return returnPublicID, charge, nil
}

// 3. upsertServicedApartment creates or updates a serviced apartment aggregate.
func (s *PropertyService) upsertServicedApartment(ctx context.Context, params UpsertServicedApartmentParams, creating bool) (string, error) {
	if err := s.validateServicedParams(params); err != nil {
		return "", err
	}

	communityID, err := s.resolveCommunityID(ctx, params.CommunityID)
	if err != nil {
		return "", err
	}

	returnPublicID := params.ListingPublicID
	err = s.runtime.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		listing, err := s.preparePropertyRoot(ctx, tx, PropertyChannelServiced, creating, propertyRootInput{
			OwnerUserID:           params.OwnerUserID,
			ListingPublicID:       params.ListingPublicID,
			Title:                 params.Title,
			Summary:               params.Summary,
			Description:           params.Description,
			DistrictCode:          params.DistrictCode,
			CommunityID:           communityID,
			PublisherIdentityType: params.PublisherIdentityType,
			BusinessStatus:        params.BusinessStatus,
		})
		if err != nil {
			return err
		}
		returnPublicID = listing.PublicID

		facilityTags, err := marshalJSON(params.FacilityTags)
		if err != nil {
			return err
		}
		serviceTags, err := marshalJSON(params.ServiceTags)
		if err != nil {
			return err
		}
		roomTypes, err := marshalJSON(params.RoomTypes)
		if err != nil {
			return err
		}

		serviced := model.ServicedApartmentProject{
			ListingID:            listing.ID,
			ProjectName:          strings.TrimSpace(params.ProjectName),
			AddressText:          strings.TrimSpace(params.AddressText),
			LowestMonthlyRentHKD: params.LowestMonthlyRentHKD,
			MinLeaseMonths:       params.MinLeaseMonths,
			FacilityTags:         facilityTags,
			ServiceTags:          serviceTags,
			RoomTypes:            roomTypes,
			ContactMethod:        strings.TrimSpace(params.ContactMethod),
			PublisherRoleLabel:   publisherRoleLabel(params.PublisherIdentityType),
		}
		if err := tx.Save(&serviced).Error; err != nil {
			return err
		}

		return s.savePropertyContactAndImages(ctx, tx, listing.ID, params.OwnerUserID, params.Contact, params.ContactMethod, params.Images)
	})
	if err != nil {
		return "", err
	}

	return returnPublicID, nil
}

// 4. updateServicedApartmentWithCharge updates a serviced apartment and charges edit points.
func (s *PropertyService) updateServicedApartmentWithCharge(ctx context.Context, params UpsertServicedApartmentParams) (string, *PointsChargeResponse, error) {
	if err := s.validateServicedParams(params); err != nil {
		return "", nil, err
	}

	communityID, err := s.resolveCommunityID(ctx, params.CommunityID)
	if err != nil {
		return "", nil, err
	}

	var returnPublicID string
	var charge *PointsChargeResponse
	err = s.runtime.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		listing, err := s.preparePropertyRoot(ctx, tx, PropertyChannelServiced, false, propertyRootInput{
			OwnerUserID:           params.OwnerUserID,
			ListingPublicID:       params.ListingPublicID,
			Title:                 params.Title,
			Summary:               params.Summary,
			Description:           params.Description,
			DistrictCode:          params.DistrictCode,
			CommunityID:           communityID,
			PublisherIdentityType: params.PublisherIdentityType,
			BusinessStatus:        params.BusinessStatus,
		})
		if err != nil {
			return err
		}
		returnPublicID = listing.PublicID

		if shouldChargeListingEdit(listing) {
			chargeResult, err := s.chargePropertyAction(ctx, tx, PropertyChannelServiced, listing, WalletActionEdit)
			if err != nil {
				return err
			}
			charge = chargeResult
		}

		facilityTags, err := marshalJSON(params.FacilityTags)
		if err != nil {
			return err
		}
		serviceTags, err := marshalJSON(params.ServiceTags)
		if err != nil {
			return err
		}
		roomTypes, err := marshalJSON(params.RoomTypes)
		if err != nil {
			return err
		}

		serviced := model.ServicedApartmentProject{
			ListingID:            listing.ID,
			ProjectName:          strings.TrimSpace(params.ProjectName),
			AddressText:          strings.TrimSpace(params.AddressText),
			LowestMonthlyRentHKD: params.LowestMonthlyRentHKD,
			MinLeaseMonths:       params.MinLeaseMonths,
			FacilityTags:         facilityTags,
			ServiceTags:          serviceTags,
			RoomTypes:            roomTypes,
			ContactMethod:        strings.TrimSpace(params.ContactMethod),
			PublisherRoleLabel:   publisherRoleLabel(params.PublisherIdentityType),
		}
		if err := tx.Save(&serviced).Error; err != nil {
			return err
		}

		return s.savePropertyContactAndImages(ctx, tx, listing.ID, params.OwnerUserID, params.Contact, params.ContactMethod, params.Images)
	})
	if err != nil {
		return "", nil, err
	}

	return returnPublicID, charge, nil
}

// 5. propertyRootInput defines shared root listing input.
type propertyRootInput struct {
	OwnerUserID           int64
	ListingPublicID       string
	Title                 string
	Summary               string
	Description           string
	DistrictCode          string
	CommunityID           *int64
	PublisherIdentityType string
	BusinessStatus        string
}

// 6. preparePropertyRoot creates or updates the common listing root.
func (s *PropertyService) preparePropertyRoot(ctx context.Context, tx *gorm.DB, channel PropertyChannel, creating bool, input propertyRootInput) (*model.Listing, error) {
	var listing model.Listing
	if creating {
		listing = model.Listing{
			PublicID:          utils.NewPublicID(),
			Module:            string(channel),
			OwnerUserID:       input.OwnerUserID,
			PublicationStatus: "draft",
			ModerationStatus:  "approved",
			BusinessStatus:    "available",
			IsDeleted:         false,
		}
	} else {
		current, _, err := s.loadOwnedPropertyListingWithTx(ctx, tx, channel, input.OwnerUserID, input.ListingPublicID)
		if err != nil {
			return nil, err
		}
		listing = *current
	}

	listing.Title = strings.TrimSpace(input.Title)
	listing.Summary = strings.TrimSpace(input.Summary)
	listing.Description = strings.TrimSpace(input.Description)
	listing.DistrictCode = strings.TrimSpace(input.DistrictCode)
	listing.CommunityID = input.CommunityID
	listing.PublisherIdentityType = fallbackPropertyPublisherIdentity(input.PublisherIdentityType)
	if !creating && strings.TrimSpace(input.BusinessStatus) != "" {
		listing.BusinessStatus = strings.TrimSpace(input.BusinessStatus)
	}

	if err := tx.Save(&listing).Error; err != nil {
		return nil, err
	}

	return &listing, nil
}

// 7. savePropertyContactAndImages saves encrypted contacts and listing images.
func (s *PropertyService) savePropertyContactAndImages(ctx context.Context, tx *gorm.DB, listingID int64, ownerUserID int64, contactInput PropertyContactInput, contactMethod string, imageInputs []ListingImageInput) error {
	contact, err := s.buildPropertyContact(listingID, contactInput, contactMethod)
	if err != nil {
		return err
	}
	s.preserveExistingPropertyContact(ctx, tx, listingID, contact)
	if err := tx.Save(contact).Error; err != nil {
		return err
	}

	if err := tx.Where("listing_id = ?", listingID).Delete(&model.ListingImage{}).Error; err != nil {
		return err
	}
	images, err := s.resolveListingImages(ctx, tx, listingID, ownerUserID, imageInputs)
	if err != nil {
		return err
	}
	if len(images) > 0 {
		return tx.Create(&images).Error
	}

	return nil
}

// 8. preserveExistingPropertyContact keeps encrypted contact values when edit payload omits them.
func (s *PropertyService) preserveExistingPropertyContact(ctx context.Context, tx *gorm.DB, listingID int64, contact *model.ListingContact) {
	var existing model.ListingContact
	if err := tx.WithContext(ctx).Where("listing_id = ?", listingID).First(&existing).Error; err != nil {
		return
	}

	if contact.PhoneEncrypted == "" {
		contact.PhoneEncrypted = existing.PhoneEncrypted
		contact.PhoneMasked = existing.PhoneMasked
	}
	if contact.WhatsAppEncrypted == "" {
		contact.WhatsAppEncrypted = existing.WhatsAppEncrypted
		contact.WhatsAppMasked = existing.WhatsAppMasked
	}
	if contact.EmailEncrypted == "" {
		contact.EmailEncrypted = existing.EmailEncrypted
	}
}

// 9. validateSaleParams validates sale listing payloads.
func (s *PropertyService) validateSaleParams(params UpsertPropertySaleParams) error {
	if strings.TrimSpace(params.Title) == "" || strings.TrimSpace(params.DistrictCode) == "" || strings.TrimSpace(params.PropertyType) == "" || strings.TrimSpace(params.AddressText) == "" {
		return errcode.New(errcode.CodeValidationError, "missing required property fields")
	}
	if params.AskingPriceHKD <= 0 || params.UsableAreaSqft <= 0 {
		return errcode.New(errcode.CodeValidationError, "valid price and usable area are required")
	}
	if strings.TrimSpace(params.ContactMethod) == "" || !isAllowedPropertyValue(params.ContactMethod, []string{"phone", "whatsapp", "chat", "both", "chat_or_whatsapp"}) {
		return errcode.New(errcode.CodeValidationError, "invalid contact method")
	}
	if strings.TrimSpace(params.BusinessStatus) != "" && !isAllowedPropertyValue(params.BusinessStatus, []string{"available", "sold"}) {
		return errcode.New(errcode.CodeValidationError, "invalid business status")
	}
	if !isAllowedPropertyDistrict(params.DistrictCode) {
		return errcode.New(errcode.CodeValidationError, "invalid district code")
	}
	if !params.Contact.ShowPhone && !params.Contact.ShowWhatsApp && !params.Contact.ShowChat {
		return errcode.New(errcode.CodeValidationError, "at least one contact channel must be enabled")
	}

	return nil
}

// 10. validateServicedParams validates serviced apartment payloads.
func (s *PropertyService) validateServicedParams(params UpsertServicedApartmentParams) error {
	if strings.TrimSpace(params.Title) == "" || strings.TrimSpace(params.ProjectName) == "" || strings.TrimSpace(params.DistrictCode) == "" || strings.TrimSpace(params.AddressText) == "" {
		return errcode.New(errcode.CodeValidationError, "missing required serviced apartment fields")
	}
	if params.LowestMonthlyRentHKD <= 0 || params.MinLeaseMonths <= 0 {
		return errcode.New(errcode.CodeValidationError, "valid rent and lease term are required")
	}
	if strings.TrimSpace(params.ContactMethod) == "" || !isAllowedPropertyValue(params.ContactMethod, []string{"phone", "whatsapp", "chat", "both", "chat_or_whatsapp"}) {
		return errcode.New(errcode.CodeValidationError, "invalid contact method")
	}
	if strings.TrimSpace(params.BusinessStatus) != "" && !isAllowedPropertyValue(params.BusinessStatus, []string{"available", "sold"}) {
		return errcode.New(errcode.CodeValidationError, "invalid business status")
	}
	if !isAllowedPropertyDistrict(params.DistrictCode) {
		return errcode.New(errcode.CodeValidationError, "invalid district code")
	}
	if !params.Contact.ShowPhone && !params.Contact.ShowWhatsApp && !params.Contact.ShowChat {
		return errcode.New(errcode.CodeValidationError, "at least one contact channel must be enabled")
	}

	return nil
}

// 11. basePropertyListQuery builds the channel-specific select query.
func (s *PropertyService) basePropertyListQuery(ctx context.Context, channel PropertyChannel) *gorm.DB {
	query := s.runtime.DB.WithContext(ctx).Table("listings")
	if channel == PropertyChannelSale {
		return query.
			Select(`listings.*,
				property_sale_listings.property_type AS sale_property_type,
				property_sale_listings.estate_name AS sale_estate_name,
				property_sale_listings.address_text AS sale_address_text,
				property_sale_listings.asking_price_hkd AS sale_asking_price_hkd,
				property_sale_listings.usable_area_sqft AS sale_usable_area_sqft,
				property_sale_listings.gross_area_sqft AS sale_gross_area_sqft,
				property_sale_listings.bedroom_count AS sale_bedroom_count,
				property_sale_listings.living_room_count AS sale_living_room_count,
				property_sale_listings.bathroom_count AS sale_bathroom_count,
				property_sale_listings.floor_level AS sale_floor_level,
				property_sale_listings.direction AS sale_direction,
				property_sale_listings.building_age AS sale_building_age,
				property_sale_listings.feature_tags AS sale_feature_tags,
				property_sale_listings.contact_method AS sale_contact_method,
				property_sale_listings.publisher_role_label AS sale_publisher_role_label`).
			Joins("JOIN property_sale_listings ON property_sale_listings.listing_id = listings.id")
	}

	return query.
		Select(`listings.*,
			serviced_apartment_projects.project_name AS serviced_project_name,
			serviced_apartment_projects.address_text AS serviced_address_text,
			serviced_apartment_projects.lowest_monthly_rent_hkd AS serviced_lowest_monthly_rent_hkd,
			serviced_apartment_projects.min_lease_months AS serviced_min_lease_months,
			serviced_apartment_projects.facility_tags AS serviced_facility_tags,
			serviced_apartment_projects.service_tags AS serviced_service_tags,
			serviced_apartment_projects.room_types AS serviced_room_types,
			serviced_apartment_projects.contact_method AS serviced_contact_method,
			serviced_apartment_projects.publisher_role_label AS serviced_publisher_role_label`).
		Joins("JOIN serviced_apartment_projects ON serviced_apartment_projects.listing_id = listings.id")
}

// 10. applyPropertyFilters applies shared public filters.
func (s *PropertyService) applyPropertyFilters(query *gorm.DB, channel PropertyChannel, filters PropertyListFilters) *gorm.DB {
	if filters.Keyword != "" {
		likeKeyword := "%" + strings.TrimSpace(filters.Keyword) + "%"
		query = query.Where("(listings.title LIKE ? OR listings.summary LIKE ? OR listings.description LIKE ?)", likeKeyword, likeKeyword, likeKeyword)
	}
	if filters.DistrictCode != "" {
		query = query.Where("listings.district_code = ?", filters.DistrictCode)
	}
	if filters.MinPriceHKD != nil {
		query = query.Where(propertyPriceColumn(channel)+" >= ?", *filters.MinPriceHKD)
	}
	if filters.MaxPriceHKD != nil {
		query = query.Where(propertyPriceColumn(channel)+" <= ?", *filters.MaxPriceHKD)
	}

	return query
}

// 11. loadPropertyRowsByIDs loads channel rows by listing IDs.
func (s *PropertyService) loadPropertyRowsByIDs(ctx context.Context, channel PropertyChannel, listingIDs []int64) ([]propertyListingRow, error) {
	if len(listingIDs) == 0 {
		return []propertyListingRow{}, nil
	}

	var rows []propertyListingRow
	if err := s.basePropertyListQuery(ctx, channel).Where("listings.id IN ?", listingIDs).Scan(&rows).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load listing")
	}

	return rows, nil
}

// 12. buildPropertySummaries maps listing rows to response payloads.
func (s *PropertyService) buildPropertySummaries(ctx context.Context, channel PropertyChannel, rows []propertyListingRow) ([]PropertyListingSummary, error) {
	listingIDs := make([]int64, 0, len(rows))
	ownerUserIDs := make([]int64, 0, len(rows))
	communityIDs := make([]int64, 0, len(rows))
	for _, item := range rows {
		listingIDs = append(listingIDs, item.ID)
		ownerUserIDs = append(ownerUserIDs, item.OwnerUserID)
		if item.CommunityID != nil {
			communityIDs = append(communityIDs, *item.CommunityID)
		}
	}

	imageMap, err := s.loadListingImages(ctx, listingIDs)
	if err != nil {
		return nil, err
	}
	ownerMap, err := s.loadUserPreviewMap(ctx, ownerUserIDs)
	if err != nil {
		return nil, err
	}
	communityMap, err := s.loadCommunityMap(ctx, communityIDs)
	if err != nil {
		return nil, err
	}

	result := make([]PropertyListingSummary, 0, len(rows))
	for _, item := range rows {
		result = append(result, s.toPropertySummary(channel, item, imageMap, ownerMap, communityMap))
	}

	return result, nil
}

// 13. toPropertySummary builds one summary payload.
func (s *PropertyService) toPropertySummary(channel PropertyChannel, item propertyListingRow, imageMap map[int64][]ListingImageResponse, ownerMap map[int64]*UserPreviewResponse, communityMap map[int64]*CommunityResponse) PropertyListingSummary {
	var cover *ListingImageResponse
	if images := imageMap[item.ID]; len(images) > 0 {
		cover = &images[0]
	}

	var community *CommunityResponse
	if item.CommunityID != nil {
		community = communityMap[*item.CommunityID]
	}

	summary := PropertyListingSummary{
		ListingID:             item.PublicID,
		Module:                item.Module,
		Title:                 item.Title,
		Summary:               item.Summary,
		DistrictCode:          item.DistrictCode,
		PublisherIdentityType: item.PublisherIdentityType,
		PublicationStatus:     item.PublicationStatus,
		BusinessStatus:        item.BusinessStatus,
		ExpireAt:              formatOptionalTime(item.ExpireAt),
		PublishedAt:           formatOptionalTime(item.PublishedAt),
		UpdatedAt:             item.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		Community:             community,
		Owner:                 ownerMap[item.OwnerUserID],
		CoverImage:            cover,
	}
	if channel == PropertyChannelSale {
		summary.PropertySale = &PropertySalePayload{
			PropertyType:       item.SalePropertyType,
			EstateName:         item.SaleEstateName,
			AddressText:        item.SaleAddressText,
			AskingPriceHKD:     item.SaleAskingPriceHKD,
			UsableAreaSqft:     item.SaleUsableAreaSqft,
			GrossAreaSqft:      item.SaleGrossAreaSqft,
			BedroomCount:       item.SaleBedroomCount,
			LivingRoomCount:    item.SaleLivingRoomCount,
			BathroomCount:      item.SaleBathroomCount,
			FloorLevel:         item.SaleFloorLevel,
			Direction:          item.SaleDirection,
			BuildingAge:        item.SaleBuildingAge,
			FeatureTags:        decodeStringSliceBytes(item.SaleFeatureTags),
			ContactMethod:      item.SaleContactMethod,
			PublisherRoleLabel: item.SalePublisherRoleLabel,
		}
		return summary
	}

	summary.ServicedApartment = &ServicedApartmentPayload{
		ProjectName:          item.ServicedProjectName,
		AddressText:          item.ServicedAddressText,
		LowestMonthlyRentHKD: item.ServicedLowestMonthlyRentHKD,
		MinLeaseMonths:       item.ServicedMinLeaseMonths,
		FacilityTags:         decodeStringSliceBytes(item.ServicedFacilityTags),
		ServiceTags:          decodeStringSliceBytes(item.ServicedServiceTags),
		RoomTypes:            decodeRoomTypeBytes(item.ServicedRoomTypes),
		ContactMethod:        item.ServicedContactMethod,
		PublisherRoleLabel:   item.ServicedPublisherRoleLabel,
	}
	return summary
}

// 14. loadPropertyListingByPublicID loads listing root and contact.
func (s *PropertyService) loadPropertyListingByPublicID(ctx context.Context, channel PropertyChannel, listingPublicID string) (*model.Listing, *model.ListingContact, error) {
	var listing model.Listing
	if err := s.runtime.DB.WithContext(ctx).Where("public_id = ? AND module = ? AND is_deleted = ?", listingPublicID, string(channel), false).First(&listing).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, errcode.New(errcode.CodeNotFound, "listing not found")
		}
		return nil, nil, errcode.New(errcode.CodeInternalError, "failed to load listing")
	}

	var contact model.ListingContact
	if err := s.runtime.DB.WithContext(ctx).Where("listing_id = ?", listing.ID).First(&contact).Error; err != nil {
		return nil, nil, errcode.New(errcode.CodeInternalError, "failed to load listing contact")
	}

	return &listing, &contact, nil
}

// 15. loadOwnedPropertyListing loads an owned property listing.
func (s *PropertyService) loadOwnedPropertyListing(ctx context.Context, channel PropertyChannel, ownerUserID int64, listingPublicID string) (*model.Listing, *model.ListingContact, error) {
	return s.loadOwnedPropertyListingWithTx(ctx, s.runtime.DB, channel, ownerUserID, listingPublicID)
}

// 16. loadOwnedPropertyListingWithTx loads an owned property listing with tx.
func (s *PropertyService) loadOwnedPropertyListingWithTx(ctx context.Context, tx *gorm.DB, channel PropertyChannel, ownerUserID int64, listingPublicID string) (*model.Listing, *model.ListingContact, error) {
	var listing model.Listing
	if err := tx.WithContext(ctx).Where("public_id = ? AND module = ? AND is_deleted = ?", listingPublicID, string(channel), false).First(&listing).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, errcode.New(errcode.CodeNotFound, "listing not found")
		}
		return nil, nil, errcode.New(errcode.CodeInternalError, "failed to load listing")
	}
	if listing.OwnerUserID != ownerUserID {
		return nil, nil, errcode.New(errcode.CodeAuthForbidden, "listing does not belong to the current user")
	}

	var contact model.ListingContact
	if err := tx.WithContext(ctx).Where("listing_id = ?", listing.ID).First(&contact).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, errcode.New(errcode.CodeInternalError, "failed to load listing contact")
	}

	return &listing, &contact, nil
}

// 17. buildPropertyContact builds the encrypted contact record.
func (s *PropertyService) buildPropertyContact(listingID int64, input PropertyContactInput, contactMethod string) (*model.ListingContact, error) {
	contact := &model.ListingContact{
		ListingID:       listingID,
		ShowPhone:       input.ShowPhone,
		ShowWhatsApp:    input.ShowWhatsApp,
		ShowChat:        input.ShowChat,
		ShowInquiryForm: input.ShowInquiryForm,
		ContactMode:     strings.TrimSpace(contactMethod),
	}
	if input.Phone != "" {
		encrypted, err := utils.EncryptString(s.runtime.Config.EncryptionKey, input.Phone)
		if err != nil {
			return nil, err
		}
		contact.PhoneEncrypted = encrypted
		contact.PhoneMasked = utils.MaskPhone(input.Phone)
	}
	if input.WhatsApp != "" {
		encrypted, err := utils.EncryptString(s.runtime.Config.EncryptionKey, input.WhatsApp)
		if err != nil {
			return nil, err
		}
		contact.WhatsAppEncrypted = encrypted
		contact.WhatsAppMasked = utils.MaskPhone(input.WhatsApp)
	}
	if input.Email != "" {
		encrypted, err := utils.EncryptString(s.runtime.Config.EncryptionKey, input.Email)
		if err != nil {
			return nil, err
		}
		contact.EmailEncrypted = encrypted
	}

	return contact, nil
}

// 18. resolveCommunityID resolves an optional community id.
func (s *PropertyService) resolveCommunityID(ctx context.Context, communityPublicID string) (*int64, error) {
	if strings.TrimSpace(communityPublicID) == "" {
		return nil, nil
	}

	var community model.Community
	if err := s.runtime.DB.WithContext(ctx).Where("public_id = ?", strings.TrimSpace(communityPublicID)).First(&community).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.New(errcode.CodeValidationError, "community not found")
		}
		return nil, errcode.New(errcode.CodeInternalError, "failed to load community")
	}

	return &community.ID, nil
}

// 19. loadListingImages loads listing images and media assets.
func (s *PropertyService) loadListingImages(ctx context.Context, listingIDs []int64) (map[int64][]ListingImageResponse, error) {
	result := make(map[int64][]ListingImageResponse)
	if len(listingIDs) == 0 {
		return result, nil
	}

	var images []model.ListingImage
	if err := s.runtime.DB.WithContext(ctx).Where("listing_id IN ?", listingIDs).Order("sort_order asc").Find(&images).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load listing images")
	}

	assetIDs := make([]int64, 0, len(images))
	for _, image := range images {
		assetIDs = append(assetIDs, image.MediaAssetID)
	}

	var assets []model.MediaAsset
	if len(assetIDs) > 0 {
		if err := s.runtime.DB.WithContext(ctx).Where("id IN ?", assetIDs).Find(&assets).Error; err != nil {
			return nil, errcode.New(errcode.CodeInternalError, "failed to load media assets")
		}
	}

	assetMap := make(map[int64]model.MediaAsset, len(assets))
	for _, asset := range assets {
		assetMap[asset.ID] = asset
	}
	for _, image := range images {
		asset := assetMap[image.MediaAssetID]
		result[image.ListingID] = append(result[image.ListingID], ListingImageResponse{
			MediaAssetID: asset.PublicID,
			URL:          s.mediaURL(&asset),
			SortOrder:    image.SortOrder,
			IsCover:      image.IsCover,
		})
	}

	return result, nil
}

// 20. resolveListingImages validates media ownership and builds image records.
func (s *PropertyService) resolveListingImages(ctx context.Context, tx *gorm.DB, listingID int64, ownerUserID int64, inputs []ListingImageInput) ([]model.ListingImage, error) {
	if len(inputs) == 0 {
		return []model.ListingImage{}, nil
	}

	assetPublicIDs := make([]string, 0, len(inputs))
	for _, input := range inputs {
		assetPublicIDs = append(assetPublicIDs, input.MediaAssetID)
	}

	var assets []model.MediaAsset
	if err := tx.WithContext(ctx).Where("public_id IN ? AND created_by = ?", assetPublicIDs, ownerUserID).Find(&assets).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load media assets")
	}
	assetMap := make(map[string]model.MediaAsset, len(assets))
	for _, asset := range assets {
		assetMap[asset.PublicID] = asset
	}

	images := make([]model.ListingImage, 0, len(inputs))
	for index, input := range inputs {
		asset, exists := assetMap[input.MediaAssetID]
		if !exists {
			return nil, errcode.New(errcode.CodeValidationError, "media asset not found")
		}
		sortOrder := input.SortOrder
		if sortOrder <= 0 {
			sortOrder = index + 1
		}
		images = append(images, model.ListingImage{
			ListingID:    listingID,
			MediaAssetID: asset.ID,
			SortOrder:    sortOrder,
			IsCover:      input.IsCover || index == 0,
		})
	}

	return images, nil
}

// 21. loadUserPreviewMap loads lightweight owner data.
func (s *PropertyService) loadUserPreviewMap(ctx context.Context, userIDs []int64) (map[int64]*UserPreviewResponse, error) {
	result := make(map[int64]*UserPreviewResponse)
	if len(userIDs) == 0 {
		return result, nil
	}

	var rows []secondhandUserPreviewRow
	if err := s.runtime.DB.WithContext(ctx).Table("users").
		Select("users.id, users.public_id, user_profiles.display_name, user_profiles.publisher_identity_type, media_assets.object_key AS avatar_object_key").
		Joins("LEFT JOIN user_profiles ON user_profiles.user_id = users.id").
		Joins("LEFT JOIN media_assets ON media_assets.id = user_profiles.avatar_asset_id").
		Where("users.id IN ?", userIDs).
		Scan(&rows).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load user previews")
	}

	for _, item := range rows {
		displayName := strings.TrimSpace(item.DisplayName)
		if displayName == "" {
			displayName = item.PublicID
		}
		result[item.ID] = &UserPreviewResponse{
			UserID:                fmt.Sprintf("%d", item.ID),
			PublicID:              item.PublicID,
			DisplayName:           displayName,
			AvatarURL:             buildMediaURL(s.runtime.Config.MediaBaseURL, item.AvatarObjectKey),
			PublisherIdentityType: item.PublisherIdentityType,
		}
	}

	return result, nil
}

// 22. loadCommunityMap loads lightweight communities.
func (s *PropertyService) loadCommunityMap(ctx context.Context, communityIDs []int64) (map[int64]*CommunityResponse, error) {
	result := make(map[int64]*CommunityResponse)
	if len(communityIDs) == 0 {
		return result, nil
	}

	var communities []model.Community
	if err := s.runtime.DB.WithContext(ctx).Where("id IN ?", communityIDs).Find(&communities).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load listing communities")
	}
	for i := range communities {
		community := communities[i]
		result[community.ID] = toCommunityResponse(&community)
	}

	return result, nil
}

// 23. validatePropertyReady ensures listing can be published.
func (s *PropertyService) validatePropertyReady(ctx context.Context, listingID int64) error {
	return s.validatePropertyReadyWithTx(ctx, s.runtime.DB, listingID)
}

// 24. validatePropertyReadyWithTx ensures listing can be published inside a transaction.
func (s *PropertyService) validatePropertyReadyWithTx(ctx context.Context, tx *gorm.DB, listingID int64) error {
	var imageCount int64
	if err := tx.WithContext(ctx).Model(&model.ListingImage{}).Where("listing_id = ?", listingID).Count(&imageCount).Error; err != nil {
		return errcode.New(errcode.CodeInternalError, "failed to validate listing images")
	}
	if imageCount == 0 {
		return errcode.New(errcode.CodeValidationError, "at least one image is required before publish")
	}

	var contact model.ListingContact
	if err := tx.WithContext(ctx).Where("listing_id = ?", listingID).First(&contact).Error; err != nil {
		return errcode.New(errcode.CodeInternalError, "failed to validate listing contacts")
	}
	if !contact.ShowPhone && !contact.ShowWhatsApp && !contact.ShowChat {
		return errcode.New(errcode.CodeValidationError, "at least one contact channel must be enabled")
	}

	return nil
}

// 25. decodeStringSliceBytes decodes JSON bytes into string slice.
func decodeStringSliceBytes(value []byte) []string {
	if len(value) == 0 {
		return []string{}
	}

	var result []string
	if err := json.Unmarshal(value, &result); err != nil {
		return []string{}
	}

	return result
}

// 26. decodeRoomTypeBytes decodes JSON bytes into room type slice.
func decodeRoomTypeBytes(value []byte) []ServicedApartmentRoomTypeInput {
	if len(value) == 0 {
		return []ServicedApartmentRoomTypeInput{}
	}

	var result []ServicedApartmentRoomTypeInput
	if err := json.Unmarshal(value, &result); err != nil {
		return []ServicedApartmentRoomTypeInput{}
	}

	return result
}

// 27. formatOptionalTime formats a time pointer as RFC3339.
func formatOptionalTime(value *time.Time) *string {
	if value == nil {
		return nil
	}

	formatted := value.UTC().Format("2006-01-02T15:04:05Z07:00")
	return &formatted
}

// 28. propertyChannelTTL returns listing validity duration.
func propertyChannelTTL(channel PropertyChannel) time.Duration {
	if channel == PropertyChannelServiced {
		return servicedApartmentTTL
	}

	return propertySaleTTL
}

// 29. propertyPriceColumn returns sortable price column.
func propertyPriceColumn(channel PropertyChannel) string {
	if channel == PropertyChannelServiced {
		return "serviced_apartment_projects.lowest_monthly_rent_hkd"
	}

	return "property_sale_listings.asking_price_hkd"
}

// 30. fallbackPropertyPublisherIdentity returns a stable identity.
func fallbackPropertyPublisherIdentity(identity string) string {
	if strings.TrimSpace(identity) == "" {
		return "owner"
	}

	return strings.TrimSpace(identity)
}

// 31. publisherRoleLabel returns display label for owner type.
func publisherRoleLabel(identity string) string {
	switch strings.TrimSpace(identity) {
	case "agent", "professional_seller":
		return "代理人"
	default:
		return "業主"
	}
}

// 32. isAllowedPropertyValue validates enum-like input.
func isAllowedPropertyValue(value string, allowedValues []string) bool {
	normalizedValue := strings.TrimSpace(value)
	for _, allowedValue := range allowedValues {
		if normalizedValue == allowedValue {
			return true
		}
	}

	return false
}

// 33. isAllowedPropertyDistrict validates district input with existing marketplace set.
func isAllowedPropertyDistrict(value string) bool {
	return isAllowedSecondhandDistrict(value)
}

// 34. buildPropertyWhatsAppURL creates a prefilled WhatsApp deep link.
func buildPropertyWhatsAppURL(phone string, title string, listingPublicID string, baseURL string, channel PropertyChannel) string {
	digits := strings.NewReplacer("+", "", " ", "", "-", "", "(", "", ")", "").Replace(phone)
	pathPrefix := "/properties/"
	if channel == PropertyChannelServiced {
		pathPrefix = "/serviced-residences/"
	}
	message := fmt.Sprintf("I am interested in %s %s%s%s", title, strings.TrimRight(baseURL, "/"), pathPrefix, listingPublicID)
	return "https://wa.me/" + digits + "?text=" + url.QueryEscape(message)
}

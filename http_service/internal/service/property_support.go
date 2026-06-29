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
			ListingID:            listing.ID,
			PropertyNo:           strings.TrimSpace(firstNonBlank(params.PropertyNo, listing.PublicID)),
			TransactionType:      normalizePropertyTransactionType(params.TransactionType),
			LocationScope:        normalizePropertyLocationScope(params.LocationScope),
			ListingCategory:      normalizePropertyListingCategory(params.ListingCategory),
			MultiUnitProject:     params.MultiUnitProject,
			PropertyType:         strings.TrimSpace(params.PropertyType),
			RentalType:           strings.TrimSpace(params.RentalType),
			RenovationType:       normalizePropertyRenovationType(params.RenovationType, params.FeatureTags),
			AgencyCompanyName:    strings.TrimSpace(params.AgencyCompanyName),
			EstateName:           strings.TrimSpace(params.EstateName),
			AddressText:          strings.TrimSpace(params.AddressText),
			AddressTextEn:        strings.TrimSpace(firstNonBlank(params.AddressTextEn, params.AddressText)),
			BlockName:            strings.TrimSpace(params.BlockName),
			UnitName:             strings.TrimSpace(params.UnitName),
			ShowUnit:             params.ShowUnit,
			Latitude:             normalizeCoordinate(params.Latitude, -90, 90),
			Longitude:            normalizeCoordinate(params.Longitude, -180, 180),
			AskingPriceHKD:       params.AskingPriceHKD,
			MonthlyRentHKD:       params.MonthlyRentHKD,
			PriceReferenceOnly:   params.PriceReferenceOnly,
			PriceNegotiable:      params.PriceNegotiable,
			AnnualPrepayDiscount: params.AnnualPrepayDiscount,
			AnnualPrepayOption:   normalizeAnnualPrepayOption(params.AnnualPrepayOption, params.AnnualPrepayDiscount),
			LeaseStartDate:       strings.TrimSpace(params.LeaseStartDate),
			RentIncluded:         strings.TrimSpace(params.RentIncluded),
			AreaMode:             normalizePropertyAreaMode(params.AreaMode),
			UsableAreaSqft:       params.UsableAreaSqft,
			GrossAreaSqft:        params.GrossAreaSqft,
			BedroomCount:         params.BedroomCount,
			LivingRoomCount:      params.LivingRoomCount,
			BathroomCount:        params.BathroomCount,
			FloorLevel:           buildPublicFloorLabel(params.FloorRaw, params.FloorZone, params.TotalFloors),
			FloorRaw:             strings.TrimSpace(firstNonBlank(params.FloorRaw, params.FloorLevel)),
			FloorZone:            resolveFloorZone(params.FloorRaw, params.FloorZone, params.TotalFloors),
			FloorDisplayRange:    buildFloorDisplayRange(params.FloorRaw, params.FloorZone, params.TotalFloors),
			TotalFloors:          params.TotalFloors,
			PublicLocationText:   buildPublicLocationText(params.BlockName, params.UnitName, params.ShowUnit, params.FloorRaw, params.FloorZone, params.TotalFloors),
			Direction:            strings.TrimSpace(params.Direction),
			BuildingAge:          strings.TrimSpace(params.BuildingAge),
			CompletionYear:       params.CompletionYear,
			BuildingTotalFloors:  params.BuildingTotalFloors,
			ManagementCompany:    strings.TrimSpace(params.ManagementCompany),
			KitchenType:          strings.TrimSpace(params.KitchenType),
			CookingMode:          strings.TrimSpace(params.CookingMode),
			ManagementFeeHKD:     params.ManagementFeeHKD,
			VideoURL:             strings.TrimSpace(params.VideoURL),
			VRURL:                strings.TrimSpace(params.VRURL),
			PrivateNote:          strings.TrimSpace(params.PrivateNote),
			TitleEn:              strings.TrimSpace(firstNonBlank(params.TitleEn, params.Title)),
			DescriptionEn:        strings.TrimSpace(firstNonBlank(params.DescriptionEn, params.Description)),
			AdPackageCode:        propertyAdPackage(params.AdPackageCode).Code,
			AdWeight:             propertyAdPackage(params.AdPackageCode).Weight,
			AdPriceHKD:           propertyAdPackage(params.AdPackageCode).PriceHKD,
			AdPricePoints:        propertyAdPackage(params.AdPackageCode).PricePoints,
			AdDurationDays:       propertyAdPackage(params.AdPackageCode).DurationDays,
			FeatureTags:          featureTags,
			ContactMethod:        strings.TrimSpace(params.ContactMethod),
			PublisherRoleLabel:   publisherRoleLabel(params.PublisherIdentityType),
		}
		if err := tx.Save(&sale).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.PropertySaleListing{}).Where("listing_id = ?", listing.ID).Update("show_unit", params.ShowUnit).Error; err != nil {
			return err
		}

		return s.savePropertyContactAndImages(ctx, tx, listing.ID, params.OwnerUserID, params.Contact, params.ContactMethod, params.Images)
	})
	if err != nil {
		return "", err
	}

	return returnPublicID, nil
}

// 2. normalizeAnnualPrepayOption normalizes annual prepay discount options.
func normalizeAnnualPrepayOption(value string, enabled bool) string {
	switch strings.TrimSpace(value) {
	case "95_off":
		return "95_off"
	case "90_off":
		return "90_off"
	}
	if enabled {
		return "95_off"
	}
	return "none"
}

// 3. updatePropertySaleWithCharge updates a sale listing and charges edit points.
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
			ListingID:            listing.ID,
			PropertyNo:           strings.TrimSpace(firstNonBlank(params.PropertyNo, listing.PublicID)),
			TransactionType:      normalizePropertyTransactionType(params.TransactionType),
			LocationScope:        normalizePropertyLocationScope(params.LocationScope),
			ListingCategory:      normalizePropertyListingCategory(params.ListingCategory),
			MultiUnitProject:     params.MultiUnitProject,
			PropertyType:         strings.TrimSpace(params.PropertyType),
			RentalType:           strings.TrimSpace(params.RentalType),
			RenovationType:       normalizePropertyRenovationType(params.RenovationType, params.FeatureTags),
			AgencyCompanyName:    strings.TrimSpace(params.AgencyCompanyName),
			EstateName:           strings.TrimSpace(params.EstateName),
			AddressText:          strings.TrimSpace(params.AddressText),
			AddressTextEn:        strings.TrimSpace(firstNonBlank(params.AddressTextEn, params.AddressText)),
			BlockName:            strings.TrimSpace(params.BlockName),
			UnitName:             strings.TrimSpace(params.UnitName),
			ShowUnit:             params.ShowUnit,
			Latitude:             normalizeCoordinate(params.Latitude, -90, 90),
			Longitude:            normalizeCoordinate(params.Longitude, -180, 180),
			AskingPriceHKD:       params.AskingPriceHKD,
			MonthlyRentHKD:       params.MonthlyRentHKD,
			PriceReferenceOnly:   params.PriceReferenceOnly,
			PriceNegotiable:      params.PriceNegotiable,
			AnnualPrepayDiscount: params.AnnualPrepayDiscount,
			AnnualPrepayOption:   normalizeAnnualPrepayOption(params.AnnualPrepayOption, params.AnnualPrepayDiscount),
			LeaseStartDate:       strings.TrimSpace(params.LeaseStartDate),
			RentIncluded:         strings.TrimSpace(params.RentIncluded),
			AreaMode:             normalizePropertyAreaMode(params.AreaMode),
			UsableAreaSqft:       params.UsableAreaSqft,
			GrossAreaSqft:        params.GrossAreaSqft,
			BedroomCount:         params.BedroomCount,
			LivingRoomCount:      params.LivingRoomCount,
			BathroomCount:        params.BathroomCount,
			FloorLevel:           buildPublicFloorLabel(params.FloorRaw, params.FloorZone, params.TotalFloors),
			FloorRaw:             strings.TrimSpace(firstNonBlank(params.FloorRaw, params.FloorLevel)),
			FloorZone:            resolveFloorZone(params.FloorRaw, params.FloorZone, params.TotalFloors),
			FloorDisplayRange:    buildFloorDisplayRange(params.FloorRaw, params.FloorZone, params.TotalFloors),
			TotalFloors:          params.TotalFloors,
			PublicLocationText:   buildPublicLocationText(params.BlockName, params.UnitName, params.ShowUnit, params.FloorRaw, params.FloorZone, params.TotalFloors),
			Direction:            strings.TrimSpace(params.Direction),
			BuildingAge:          strings.TrimSpace(params.BuildingAge),
			CompletionYear:       params.CompletionYear,
			BuildingTotalFloors:  params.BuildingTotalFloors,
			ManagementCompany:    strings.TrimSpace(params.ManagementCompany),
			KitchenType:          strings.TrimSpace(params.KitchenType),
			CookingMode:          strings.TrimSpace(params.CookingMode),
			ManagementFeeHKD:     params.ManagementFeeHKD,
			VideoURL:             strings.TrimSpace(params.VideoURL),
			VRURL:                strings.TrimSpace(params.VRURL),
			PrivateNote:          strings.TrimSpace(params.PrivateNote),
			TitleEn:              strings.TrimSpace(firstNonBlank(params.TitleEn, params.Title)),
			DescriptionEn:        strings.TrimSpace(firstNonBlank(params.DescriptionEn, params.Description)),
			AdPackageCode:        propertyAdPackage(params.AdPackageCode).Code,
			AdWeight:             propertyAdPackage(params.AdPackageCode).Weight,
			AdPriceHKD:           propertyAdPackage(params.AdPackageCode).PriceHKD,
			AdPricePoints:        propertyAdPackage(params.AdPackageCode).PricePoints,
			AdDurationDays:       propertyAdPackage(params.AdPackageCode).DurationDays,
			FeatureTags:          featureTags,
			ContactMethod:        strings.TrimSpace(params.ContactMethod),
			PublisherRoleLabel:   publisherRoleLabel(params.PublisherIdentityType),
		}
		if err := tx.Save(&sale).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.PropertySaleListing{}).Where("listing_id = ?", listing.ID).Update("show_unit", params.ShowUnit).Error; err != nil {
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
	derived := deriveServicedApartmentFields(params)

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
		adPackage := servicedApartmentAdPackage(params.AdPackageCode)

		serviced := model.ServicedApartmentProject{
			ListingID:            listing.ID,
			ProjectName:          strings.TrimSpace(params.ProjectName),
			ProjectNameEn:        strings.TrimSpace(params.ProjectNameEn),
			AddressText:          strings.TrimSpace(params.AddressText),
			AddressTextEn:        strings.TrimSpace(params.AddressTextEn),
			WebsiteURL:           strings.TrimSpace(params.WebsiteURL),
			WhatsApp:             strings.TrimSpace(params.WhatsApp),
			Fax:                  strings.TrimSpace(params.Fax),
			DescriptionEn:        strings.TrimSpace(params.DescriptionEn),
			ServiceIntro:         strings.TrimSpace(params.ServiceIntro),
			BenefitsText:         strings.TrimSpace(params.BenefitsText),
			ExtraChargesText:     strings.TrimSpace(params.ExtraChargesText),
			LowestMonthlyRentHKD: derived.LowestMonthlyRentHKD,
			LowestDailyRentHKD:   derived.LowestDailyRentHKD,
			PriceReferenceOnly:   params.PriceReferenceOnly,
			PriceNegotiable:      params.PriceNegotiable,
			MinUsableAreaSqft:    derived.MinUsableAreaSqft,
			MinLeaseMonths:       derived.MinLeaseMonths,
			MinStayValue:         derived.MinStayValue,
			MinStayUnit:          derived.MinStayUnit,
			LocationScope:        normalizePropertyLocationScope(params.LocationScope),
			ListingCategory:      normalizePropertyListingCategory(params.ListingCategory),
			MultiUnitProject:     params.MultiUnitProject,
			FacilityTags:         facilityTags,
			ServiceTags:          serviceTags,
			RoomTypes:            roomTypes,
			AdPackageCode:        adPackage.Code,
			AdWeight:             adPackage.Weight,
			AdPriceHKD:           adPackage.PriceHKD,
			AdPricePoints:        adPackage.PricePoints,
			AdDurationDays:       adPackage.DurationDays,
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
	derived := deriveServicedApartmentFields(params)

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
		adPackage := servicedApartmentAdPackage(params.AdPackageCode)

		serviced := model.ServicedApartmentProject{
			ListingID:            listing.ID,
			ProjectName:          strings.TrimSpace(params.ProjectName),
			ProjectNameEn:        strings.TrimSpace(params.ProjectNameEn),
			AddressText:          strings.TrimSpace(params.AddressText),
			AddressTextEn:        strings.TrimSpace(params.AddressTextEn),
			WebsiteURL:           strings.TrimSpace(params.WebsiteURL),
			WhatsApp:             strings.TrimSpace(params.WhatsApp),
			Fax:                  strings.TrimSpace(params.Fax),
			DescriptionEn:        strings.TrimSpace(params.DescriptionEn),
			ServiceIntro:         strings.TrimSpace(params.ServiceIntro),
			BenefitsText:         strings.TrimSpace(params.BenefitsText),
			ExtraChargesText:     strings.TrimSpace(params.ExtraChargesText),
			LowestMonthlyRentHKD: derived.LowestMonthlyRentHKD,
			LowestDailyRentHKD:   derived.LowestDailyRentHKD,
			PriceReferenceOnly:   params.PriceReferenceOnly,
			PriceNegotiable:      params.PriceNegotiable,
			MinUsableAreaSqft:    derived.MinUsableAreaSqft,
			MinLeaseMonths:       derived.MinLeaseMonths,
			MinStayValue:         derived.MinStayValue,
			MinStayUnit:          derived.MinStayUnit,
			LocationScope:        normalizePropertyLocationScope(params.LocationScope),
			ListingCategory:      normalizePropertyListingCategory(params.ListingCategory),
			MultiUnitProject:     params.MultiUnitProject,
			FacilityTags:         facilityTags,
			ServiceTags:          serviceTags,
			RoomTypes:            roomTypes,
			AdPackageCode:        adPackage.Code,
			AdWeight:             adPackage.Weight,
			AdPriceHKD:           adPackage.PriceHKD,
			AdPricePoints:        adPackage.PricePoints,
			AdDurationDays:       adPackage.DurationDays,
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
	if strings.TrimSpace(params.Title) == "" || strings.TrimSpace(params.Description) == "" || strings.TrimSpace(params.DistrictCode) == "" || strings.TrimSpace(params.PropertyType) == "" || strings.TrimSpace(params.EstateName) == "" || strings.TrimSpace(params.AddressText) == "" {
		return errcode.New(errcode.CodeValidationError, "missing required property fields")
	}
	rawTransactionType := strings.TrimSpace(params.TransactionType)
	if rawTransactionType != "" && !isAllowedPropertyValue(rawTransactionType, []string{"sale", "rent"}) {
		return errcode.New(errcode.CodeValidationError, "invalid transaction type")
	}
	if !isAllowedPropertyValue(normalizePropertyLocationScope(params.LocationScope), []string{"local", "overseas"}) {
		return errcode.New(errcode.CodeValidationError, "invalid location scope")
	}
	if !isAllowedPropertyValue(normalizePropertyListingCategory(params.ListingCategory), []string{"standard", "new_development", "developer_project", "multi_unit"}) {
		return errcode.New(errcode.CodeValidationError, "invalid listing category")
	}
	transactionType := normalizePropertyTransactionType(params.TransactionType)
	if transactionType == "sale" && !params.PriceNegotiable && params.AskingPriceHKD <= 0 {
		return errcode.New(errcode.CodeValidationError, "valid asking price is required")
	}
	if transactionType == "rent" && !params.PriceNegotiable && params.MonthlyRentHKD <= 0 {
		return errcode.New(errcode.CodeValidationError, "valid monthly rent is required")
	}
	if params.UsableAreaSqft <= 0 {
		return errcode.New(errcode.CodeValidationError, "valid usable area is required")
	}
	if strings.TrimSpace(params.FloorRaw) == "" && strings.TrimSpace(params.FloorLevel) == "" {
		return errcode.New(errcode.CodeValidationError, "valid floor is required")
	}
	rawAreaMode := strings.TrimSpace(params.AreaMode)
	if rawAreaMode != "" && !isAllowedPropertyValue(rawAreaMode, []string{"usable", "gross"}) {
		return errcode.New(errcode.CodeValidationError, "invalid area mode")
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
	if !isAllowedPropertyValue(normalizePropertyAdPackageCode(params.AdPackageCode), []string{"basic", "featured", "premium", "fast_sale"}) {
		return errcode.New(errcode.CodeValidationError, "invalid ad package")
	}

	return nil
}

// 10. validateServicedParams validates serviced apartment payloads.
func (s *PropertyService) validateServicedParams(params UpsertServicedApartmentParams) error {
	if strings.TrimSpace(params.Title) == "" || strings.TrimSpace(params.ProjectName) == "" || strings.TrimSpace(params.DistrictCode) == "" || strings.TrimSpace(params.AddressText) == "" {
		return errcode.New(errcode.CodeValidationError, "missing required serviced apartment fields")
	}
	if !isAllowedPropertyValue(normalizePropertyLocationScope(params.LocationScope), []string{"local", "overseas"}) {
		return errcode.New(errcode.CodeValidationError, "invalid location scope")
	}
	if !isAllowedPropertyValue(normalizePropertyListingCategory(params.ListingCategory), []string{"standard", "new_development", "developer_project", "multi_unit"}) {
		return errcode.New(errcode.CodeValidationError, "invalid listing category")
	}
	if !isAllowedPropertyValue(normalizePropertyAdPackageCode(params.AdPackageCode), []string{"basic", "featured", "premium", "fast_sale"}) {
		return errcode.New(errcode.CodeValidationError, "invalid ad package")
	}
	if len(params.RoomTypes) == 0 {
		return errcode.New(errcode.CodeValidationError, "at least one room type is required")
	}
	for _, room := range params.RoomTypes {
		monthlyMin := firstPositive(room.MonthlyRentMinHKD, room.MonthlyRentHKD)
		if strings.TrimSpace(room.Name) == "" {
			return errcode.New(errcode.CodeValidationError, "room type name is required")
		}
		if !params.PriceReferenceOnly && !params.PriceNegotiable && monthlyMin <= 0 && room.DailyRentMinHKD <= 0 {
			return errcode.New(errcode.CodeValidationError, "room type requires a monthly or daily price")
		}
		if room.MonthlyRentMaxHKD > 0 && monthlyMin > 0 && room.MonthlyRentMaxHKD < monthlyMin {
			return errcode.New(errcode.CodeValidationError, "monthly maximum rent cannot be lower than minimum rent")
		}
		if room.DailyRentMaxHKD > 0 && room.DailyRentMinHKD > 0 && room.DailyRentMaxHKD < room.DailyRentMinHKD {
			return errcode.New(errcode.CodeValidationError, "daily maximum rent cannot be lower than minimum rent")
		}
		if room.UsableAreaSqft < 0 {
			return errcode.New(errcode.CodeValidationError, "invalid room type area")
		}
		if strings.TrimSpace(room.MinStayUnit) != "" && normalizeStayUnit(room.MinStayUnit) == "" {
			return errcode.New(errcode.CodeValidationError, "invalid minimum stay unit")
		}
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
				property_sale_listings.property_no AS sale_property_no,
				property_sale_listings.transaction_type AS sale_transaction_type,
				property_sale_listings.location_scope AS sale_location_scope,
				property_sale_listings.listing_category AS sale_listing_category,
				property_sale_listings.multi_unit_project AS sale_multi_unit_project,
				property_sale_listings.property_type AS sale_property_type,
				property_sale_listings.rental_type AS sale_rental_type,
				property_sale_listings.renovation_type AS sale_renovation_type,
				property_sale_listings.agency_company_name AS sale_agency_company_name,
				property_sale_listings.estate_name AS sale_estate_name,
				property_sale_listings.address_text AS sale_address_text,
				property_sale_listings.address_text_en AS sale_address_text_en,
				property_sale_listings.block_name AS sale_block_name,
				property_sale_listings.unit_name AS sale_unit_name,
				property_sale_listings.show_unit AS sale_show_unit,
				property_sale_listings.latitude AS sale_latitude,
				property_sale_listings.longitude AS sale_longitude,
				property_sale_listings.asking_price_hkd AS sale_asking_price_hkd,
				property_sale_listings.monthly_rent_hkd AS sale_monthly_rent_hkd,
				property_sale_listings.price_reference_only AS sale_price_reference_only,
				property_sale_listings.price_negotiable AS sale_price_negotiable,
				property_sale_listings.annual_prepay_discount AS sale_annual_prepay_discount,
				property_sale_listings.annual_prepay_option AS sale_annual_prepay_option,
				property_sale_listings.lease_start_date AS sale_lease_start_date,
				property_sale_listings.rent_included AS sale_rent_included,
				property_sale_listings.area_mode AS sale_area_mode,
				property_sale_listings.usable_area_sqft AS sale_usable_area_sqft,
				property_sale_listings.gross_area_sqft AS sale_gross_area_sqft,
				property_sale_listings.bedroom_count AS sale_bedroom_count,
				property_sale_listings.living_room_count AS sale_living_room_count,
				property_sale_listings.bathroom_count AS sale_bathroom_count,
				property_sale_listings.floor_level AS sale_floor_level,
				property_sale_listings.floor_raw AS sale_floor_raw,
				property_sale_listings.floor_zone AS sale_floor_zone,
				property_sale_listings.floor_display_range AS sale_floor_display_range,
				property_sale_listings.total_floors AS sale_total_floors,
				property_sale_listings.public_location_text AS sale_public_location_text,
				property_sale_listings.direction AS sale_direction,
				property_sale_listings.building_age AS sale_building_age,
				property_sale_listings.completion_year AS sale_completion_year,
				property_sale_listings.building_total_floors AS sale_building_total_floors,
				property_sale_listings.management_company AS sale_management_company,
				property_sale_listings.kitchen_type AS sale_kitchen_type,
				property_sale_listings.cooking_mode AS sale_cooking_mode,
				property_sale_listings.management_fee_hkd AS sale_management_fee_hkd,
				property_sale_listings.video_url AS sale_video_url,
				property_sale_listings.vr_url AS sale_vr_url,
				property_sale_listings.private_note AS sale_private_note,
				property_sale_listings.title_en AS sale_title_en,
				property_sale_listings.description_en AS sale_description_en,
				property_sale_listings.ad_package_code AS sale_ad_package_code,
				property_sale_listings.ad_weight AS sale_ad_weight,
				property_sale_listings.ad_price_hkd AS sale_ad_price_hkd,
				property_sale_listings.ad_price_points AS sale_ad_price_points,
				property_sale_listings.ad_duration_days AS sale_ad_duration_days,
				property_sale_listings.ad_expires_at AS sale_ad_expires_at,
				property_sale_listings.feature_tags AS sale_feature_tags,
				property_sale_listings.contact_method AS sale_contact_method,
				property_sale_listings.publisher_role_label AS sale_publisher_role_label,
				property_sale_listings.view_count AS sale_view_count,
				property_sale_listings.inquiry_count AS sale_inquiry_count`).
			Joins("JOIN property_sale_listings ON property_sale_listings.listing_id = listings.id")
	}

	return query.
		Select(`listings.*,
			serviced_apartment_projects.project_name AS serviced_project_name,
			serviced_apartment_projects.project_name_en AS serviced_project_name_en,
			serviced_apartment_projects.address_text AS serviced_address_text,
			serviced_apartment_projects.address_text_en AS serviced_address_text_en,
			serviced_apartment_projects.website_url AS serviced_website_url,
			serviced_apartment_projects.whatsapp AS serviced_whats_app,
			serviced_apartment_projects.fax AS serviced_fax,
			serviced_apartment_projects.description_en AS serviced_description_en,
			serviced_apartment_projects.service_intro AS serviced_service_intro,
			serviced_apartment_projects.benefits_text AS serviced_benefits_text,
			serviced_apartment_projects.extra_charges_text AS serviced_extra_charges_text,
			serviced_apartment_projects.lowest_monthly_rent_hkd AS serviced_lowest_monthly_rent_hkd,
			serviced_apartment_projects.lowest_daily_rent_hkd AS serviced_lowest_daily_rent_hkd,
			serviced_apartment_projects.price_reference_only AS serviced_price_reference_only,
			serviced_apartment_projects.price_negotiable AS serviced_price_negotiable,
			serviced_apartment_projects.min_usable_area_sqft AS serviced_min_usable_area_sqft,
			serviced_apartment_projects.min_lease_months AS serviced_min_lease_months,
			serviced_apartment_projects.min_stay_value AS serviced_min_stay_value,
			serviced_apartment_projects.min_stay_unit AS serviced_min_stay_unit,
			serviced_apartment_projects.location_scope AS serviced_location_scope,
			serviced_apartment_projects.listing_category AS serviced_listing_category,
			serviced_apartment_projects.multi_unit_project AS serviced_multi_unit_project,
			serviced_apartment_projects.facility_tags AS serviced_facility_tags,
			serviced_apartment_projects.service_tags AS serviced_service_tags,
			serviced_apartment_projects.room_types AS serviced_room_types,
			serviced_apartment_projects.ad_package_code AS serviced_ad_package_code,
			serviced_apartment_projects.ad_weight AS serviced_ad_weight,
			serviced_apartment_projects.ad_price_hkd AS serviced_ad_price_hkd,
			serviced_apartment_projects.ad_price_points AS serviced_ad_price_points,
			serviced_apartment_projects.ad_duration_days AS serviced_ad_duration_days,
			serviced_apartment_projects.ad_expires_at AS serviced_ad_expires_at,
			serviced_apartment_projects.contact_method AS serviced_contact_method,
			serviced_apartment_projects.publisher_role_label AS serviced_publisher_role_label`).
		Joins("JOIN serviced_apartment_projects ON serviced_apartment_projects.listing_id = listings.id")
}

// 10. applyPropertyFilters applies shared public filters.
func (s *PropertyService) applyPropertyFilters(query *gorm.DB, channel PropertyChannel, filters PropertyListFilters) *gorm.DB {
	if filters.Keyword != "" {
		likeKeyword := "%" + strings.TrimSpace(filters.Keyword) + "%"
		condition, argCount := propertyKeywordSearchCondition(channel)
		args := make([]any, argCount)
		for index := range args {
			args[index] = likeKeyword
		}
		query = query.Where(condition, args...)
	}
	if filters.RegionCode != "" {
		districts, ok := propertyDistrictsForRegion(filters.RegionCode)
		if !ok {
			return query.Where("1 = 0")
		}
		query = query.Where("listings.district_code IN ?", districts)
	}
	if filters.DistrictCode != "" {
		query = query.Where("listings.district_code = ?", filters.DistrictCode)
	}
	if filters.PublisherIdentityType != "" {
		query = query.Where("listings.publisher_identity_type = ?", filters.PublisherIdentityType)
	}
	if channel == PropertyChannelSale {
		if filters.TransactionType != "" {
			query = query.Where("property_sale_listings.transaction_type = ?", filters.TransactionType)
		}
		if filters.PropertyType != "" {
			if filters.PropertyType == "residential" {
				query = query.Where("property_sale_listings.property_type IN ?", []string{"residential", "private_flat", "estate", "house"})
			} else {
				query = query.Where("property_sale_listings.property_type = ?", filters.PropertyType)
			}
		}
		if filters.RentalType != "" {
			query = query.Where("property_sale_listings.rental_type = ?", filters.RentalType)
		}
		if filters.RenovationType != "" {
			query = query.Where("(property_sale_listings.renovation_type = ? OR property_sale_listings.feature_tags @> ?)", filters.RenovationType, fmt.Sprintf(`["%s"]`, strings.ReplaceAll(filters.RenovationType, `"`, `\"`)))
		}
		if filters.BedroomCount != nil {
			if *filters.BedroomCount >= 4 {
				query = query.Where("property_sale_listings.bedroom_count >= ?", *filters.BedroomCount)
			} else {
				query = query.Where("property_sale_listings.bedroom_count = ?", *filters.BedroomCount)
			}
		}
		if filters.MinAreaSqft != nil {
			query = query.Where(propertyAreaColumn(filters.AreaMode)+" >= ?", *filters.MinAreaSqft)
		}
		if filters.MaxAreaSqft != nil {
			query = query.Where(propertyAreaColumn(filters.AreaMode)+" <= ?", *filters.MaxAreaSqft)
		}
		if filters.IsNew != nil {
			if *filters.IsNew {
				query = query.Where("property_sale_listings.feature_tags @> ?", `["brand_new"]`)
			} else {
				query = query.Where("NOT (property_sale_listings.feature_tags @> ?)", `["brand_new"]`)
			}
		}
		for _, tag := range filters.FeatureTags {
			query = query.Where("property_sale_listings.feature_tags @> ?", fmt.Sprintf(`["%s"]`, strings.ReplaceAll(tag, `"`, `\"`)))
		}
	} else {
		if filters.MinAreaSqft != nil {
			query = query.Where("serviced_apartment_projects.min_usable_area_sqft >= ?", *filters.MinAreaSqft)
		}
		if filters.MaxAreaSqft != nil {
			query = query.Where("serviced_apartment_projects.min_usable_area_sqft <= ?", *filters.MaxAreaSqft)
		}
		for _, tag := range filters.FacilityTags {
			query = query.Where("serviced_apartment_projects.facility_tags @> ?", fmt.Sprintf(`["%s"]`, strings.ReplaceAll(tag, `"`, `\"`)))
		}
		for _, tag := range filters.ServiceTags {
			query = query.Where("serviced_apartment_projects.service_tags @> ?", fmt.Sprintf(`["%s"]`, strings.ReplaceAll(tag, `"`, `\"`)))
		}
	}
	if filters.HasMedia {
		query = query.Where("EXISTS (SELECT 1 FROM listing_images WHERE listing_images.listing_id = listings.id)")
	}
	if filters.MinPriceHKD != nil {
		query = query.Where(propertyPriceColumn(channel, filters.PriceMode)+" >= ?", *filters.MinPriceHKD)
	}
	if filters.MaxPriceHKD != nil {
		query = query.Where(propertyPriceColumn(channel, filters.PriceMode)+" <= ?", *filters.MaxPriceHKD)
	}

	return query
}

// 10.1 propertyKeywordSearchCondition returns channel keyword search SQL.
func propertyKeywordSearchCondition(channel PropertyChannel) (string, int) {
	if channel == PropertyChannelServiced {
		return `(
			listings.title ILIKE ?
			OR listings.summary ILIKE ?
			OR listings.description ILIKE ?
			OR serviced_apartment_projects.project_name ILIKE ?
			OR serviced_apartment_projects.project_name_en ILIKE ?
			OR serviced_apartment_projects.address_text ILIKE ?
			OR serviced_apartment_projects.address_text_en ILIKE ?
		)`, 7
	}

	return `(
		listings.title ILIKE ?
		OR listings.summary ILIKE ?
		OR listings.description ILIKE ?
		OR property_sale_listings.property_no ILIKE ?
		OR property_sale_listings.estate_name ILIKE ?
		OR property_sale_listings.address_text ILIKE ?
		OR property_sale_listings.address_text_en ILIKE ?
		OR property_sale_listings.block_name ILIKE ?
		OR property_sale_listings.unit_name ILIKE ?
		OR property_sale_listings.public_location_text ILIKE ?
	)`, 10
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
			PropertyNo:           item.SalePropertyNo,
			TransactionType:      normalizePropertyTransactionType(item.SaleTransactionType),
			LocationScope:        normalizePropertyLocationScope(item.SaleLocationScope),
			ListingCategory:      normalizePropertyListingCategory(item.SaleListingCategory),
			MultiUnitProject:     item.SaleMultiUnitProject,
			PropertyType:         item.SalePropertyType,
			RentalType:           item.SaleRentalType,
			RenovationType:       item.SaleRenovationType,
			AgencyCompanyName:    item.SaleAgencyCompanyName,
			EstateName:           item.SaleEstateName,
			AddressText:          item.SaleAddressText,
			AddressTextEn:        item.SaleAddressTextEn,
			BlockName:            item.SaleBlockName,
			UnitName:             publicUnitName(item.SaleUnitName, item.SaleShowUnit),
			ShowUnit:             item.SaleShowUnit,
			Latitude:             item.SaleLatitude,
			Longitude:            item.SaleLongitude,
			AskingPriceHKD:       item.SaleAskingPriceHKD,
			MonthlyRentHKD:       item.SaleMonthlyRentHKD,
			PriceReferenceOnly:   item.SalePriceReferenceOnly,
			PriceNegotiable:      item.SalePriceNegotiable,
			AnnualPrepayDiscount: item.SaleAnnualPrepayDiscount,
			AnnualPrepayOption:   normalizeAnnualPrepayOption(item.SaleAnnualPrepayOption, item.SaleAnnualPrepayDiscount),
			LeaseStartDate:       item.SaleLeaseStartDate,
			RentIncluded:         item.SaleRentIncluded,
			AreaMode:             normalizePropertyAreaMode(item.SaleAreaMode),
			UsableAreaSqft:       item.SaleUsableAreaSqft,
			GrossAreaSqft:        item.SaleGrossAreaSqft,
			BedroomCount:         item.SaleBedroomCount,
			LivingRoomCount:      item.SaleLivingRoomCount,
			BathroomCount:        item.SaleBathroomCount,
			FloorLevel:           item.SaleFloorLevel,
			FloorZone:            item.SaleFloorZone,
			FloorDisplayRange:    item.SaleFloorDisplayRange,
			TotalFloors:          item.SaleTotalFloors,
			PublicLocationText:   item.SalePublicLocationText,
			Direction:            item.SaleDirection,
			BuildingAge:          item.SaleBuildingAge,
			CompletionYear:       item.SaleCompletionYear,
			BuildingTotalFloors:  item.SaleBuildingTotalFloors,
			ManagementCompany:    item.SaleManagementCompany,
			KitchenType:          item.SaleKitchenType,
			CookingMode:          item.SaleCookingMode,
			ManagementFeeHKD:     item.SaleManagementFeeHKD,
			VideoURL:             item.SaleVideoURL,
			VRURL:                item.SaleVRURL,
			TitleEn:              item.SaleTitleEn,
			DescriptionEn:        item.SaleDescriptionEn,
			AdPackageCode:        normalizePropertyAdPackageCode(item.SaleAdPackageCode),
			AdWeight:             item.SaleAdWeight,
			AdPriceHKD:           item.SaleAdPriceHKD,
			AdPricePoints:        item.SaleAdPricePoints,
			AdDurationDays:       item.SaleAdDurationDays,
			AdExpiresAt:          formatOptionalTime(item.SaleAdExpiresAt),
			FeatureTags:          decodeStringSliceBytes(item.SaleFeatureTags),
			ContactMethod:        item.SaleContactMethod,
			PublisherRoleLabel:   item.SalePublisherRoleLabel,
			ViewCount:            item.SaleViewCount,
			InquiryCount:         item.SaleInquiryCount,
		}
		return summary
	}

	summary.ServicedApartment = &ServicedApartmentPayload{
		ProjectName:          item.ServicedProjectName,
		ProjectNameEn:        item.ServicedProjectNameEn,
		AddressText:          item.ServicedAddressText,
		AddressTextEn:        item.ServicedAddressTextEn,
		WebsiteURL:           item.ServicedWebsiteURL,
		WhatsApp:             item.ServicedWhatsApp,
		Fax:                  item.ServicedFax,
		DescriptionEn:        item.ServicedDescriptionEn,
		ServiceIntro:         item.ServicedServiceIntro,
		BenefitsText:         item.ServicedBenefitsText,
		ExtraChargesText:     item.ServicedExtraChargesText,
		LowestMonthlyRentHKD: item.ServicedLowestMonthlyRentHKD,
		LowestDailyRentHKD:   item.ServicedLowestDailyRentHKD,
		PriceReferenceOnly:   item.ServicedPriceReferenceOnly,
		PriceNegotiable:      item.ServicedPriceNegotiable,
		MinUsableAreaSqft:    item.ServicedMinUsableAreaSqft,
		MinLeaseMonths:       item.ServicedMinLeaseMonths,
		MinStayValue:         item.ServicedMinStayValue,
		MinStayUnit:          item.ServicedMinStayUnit,
		LocationScope:        normalizePropertyLocationScope(item.ServicedLocationScope),
		ListingCategory:      normalizePropertyListingCategory(item.ServicedListingCategory),
		MultiUnitProject:     item.ServicedMultiUnitProject,
		AdPackageCode:        normalizePropertyAdPackageCode(item.ServicedAdPackageCode),
		AdWeight:             item.ServicedAdWeight,
		AdPriceHKD:           item.ServicedAdPriceHKD,
		AdPricePoints:        item.ServicedAdPricePoints,
		AdDurationDays:       item.ServicedAdDurationDays,
		AdExpiresAt:          formatOptionalTime(item.ServicedAdExpiresAt),
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
	for index := range result {
		normalizeServicedRoomType(&result[index])
	}

	return result
}

// 27. servicedApartmentDerivedFields stores calculated project values.
type servicedApartmentDerivedFields struct {
	LowestMonthlyRentHKD float64
	LowestDailyRentHKD   float64
	MinUsableAreaSqft    int
	MinLeaseMonths       int
	MinStayValue         int
	MinStayUnit          string
}

// 28. deriveServicedApartmentFields calculates searchable values from room types.
func deriveServicedApartmentFields(params UpsertServicedApartmentParams) servicedApartmentDerivedFields {
	result := servicedApartmentDerivedFields{
		LowestMonthlyRentHKD: params.LowestMonthlyRentHKD,
		LowestDailyRentHKD:   params.LowestDailyRentHKD,
		MinUsableAreaSqft:    params.MinUsableAreaSqft,
		MinLeaseMonths:       firstPositiveInt(params.MinLeaseMonths, 1),
		MinStayValue:         firstPositiveInt(params.MinStayValue, params.MinLeaseMonths, 1),
		MinStayUnit:          normalizeStayUnit(params.MinStayUnit),
	}
	for index := range params.RoomTypes {
		normalizeServicedRoomType(&params.RoomTypes[index])
		if params.RoomTypes[index].MonthlyRentMinHKD > 0 && (result.LowestMonthlyRentHKD <= 0 || params.RoomTypes[index].MonthlyRentMinHKD < result.LowestMonthlyRentHKD) {
			result.LowestMonthlyRentHKD = params.RoomTypes[index].MonthlyRentMinHKD
		}
		if params.RoomTypes[index].DailyRentMinHKD > 0 && (result.LowestDailyRentHKD <= 0 || params.RoomTypes[index].DailyRentMinHKD < result.LowestDailyRentHKD) {
			result.LowestDailyRentHKD = params.RoomTypes[index].DailyRentMinHKD
		}
		if params.RoomTypes[index].UsableAreaSqft > 0 && (result.MinUsableAreaSqft <= 0 || params.RoomTypes[index].UsableAreaSqft < result.MinUsableAreaSqft) {
			result.MinUsableAreaSqft = params.RoomTypes[index].UsableAreaSqft
		}
		if params.RoomTypes[index].MinStayValue > 0 && (result.MinStayValue <= 0 || compareStay(params.RoomTypes[index].MinStayValue, params.RoomTypes[index].MinStayUnit, result.MinStayValue, result.MinStayUnit) < 0) {
			result.MinStayValue = params.RoomTypes[index].MinStayValue
			result.MinStayUnit = params.RoomTypes[index].MinStayUnit
		}
		if params.RoomTypes[index].MinLeaseMonths > 0 && (result.MinLeaseMonths <= 0 || params.RoomTypes[index].MinLeaseMonths < result.MinLeaseMonths) {
			result.MinLeaseMonths = params.RoomTypes[index].MinLeaseMonths
		}
	}
	if result.MinStayUnit == "" {
		result.MinStayUnit = "month"
	}

	return result
}

// 29. normalizeServicedRoomType fills legacy room values.
func normalizeServicedRoomType(room *ServicedApartmentRoomTypeInput) {
	if room == nil {
		return
	}
	room.Name = strings.TrimSpace(room.Name)
	room.RoomCategory = strings.TrimSpace(room.RoomCategory)
	room.MonthlyRentMinHKD = firstPositive(room.MonthlyRentMinHKD, room.MonthlyRentHKD)
	room.MonthlyRentHKD = room.MonthlyRentMinHKD
	if room.MonthlyRentMaxHKD <= 0 {
		room.MonthlyRentMaxHKD = room.MonthlyRentMinHKD
	}
	room.MinStayUnit = normalizeStayUnit(room.MinStayUnit)
	if room.MinStayUnit == "" {
		room.MinStayUnit = "month"
	}
	room.MinStayValue = firstPositiveInt(room.MinStayValue, room.MinLeaseMonths, 1)
	room.MinLeaseMonths = firstPositiveInt(room.MinLeaseMonths, room.MinStayValue, 1)
}

// 30. normalizeStayUnit keeps supported minimum stay units.
func normalizeStayUnit(value string) string {
	switch strings.TrimSpace(value) {
	case "day", "month":
		return strings.TrimSpace(value)
	default:
		return ""
	}
}

// 31. compareStay compares stay values by approximate days.
func compareStay(leftValue int, leftUnit string, rightValue int, rightUnit string) int {
	leftDays := leftValue
	if leftUnit == "month" {
		leftDays = leftValue * 30
	}
	rightDays := rightValue
	if rightUnit == "month" {
		rightDays = rightValue * 30
	}

	return leftDays - rightDays
}

// 32. firstPositive returns the first positive float.
func firstPositive(values ...float64) float64 {
	for _, value := range values {
		if value > 0 {
			return value
		}
	}

	return 0
}

// 33. firstPositiveInt returns the first positive int.
func firstPositiveInt(values ...int) int {
	for _, value := range values {
		if value > 0 {
			return value
		}
	}

	return 0
}

// 34. formatOptionalTime formats a time pointer as RFC3339.
func formatOptionalTime(value *time.Time) *string {
	if value == nil {
		return nil
	}

	formatted := value.UTC().Format("2006-01-02T15:04:05Z07:00")
	return &formatted
}

// 35. propertyChannelTTL returns listing validity duration.
func propertyChannelTTL(channel PropertyChannel) time.Duration {
	if channel == PropertyChannelServiced {
		return servicedApartmentTTL
	}

	return propertySaleTTL
}

// 36. propertyPriceColumn returns sortable price column.
func propertyPriceColumn(channel PropertyChannel, priceMode string) string {
	if channel == PropertyChannelServiced {
		if strings.TrimSpace(priceMode) == "daily" {
			return "serviced_apartment_projects.lowest_daily_rent_hkd"
		}
		return "serviced_apartment_projects.lowest_monthly_rent_hkd"
	}

	return "CASE WHEN property_sale_listings.transaction_type = 'rent' THEN property_sale_listings.monthly_rent_hkd ELSE property_sale_listings.asking_price_hkd END"
}

// 37. propertyDefaultSortPrefix returns ad priority ordering.
func propertyDefaultSortPrefix(channel PropertyChannel) string {
	if channel == PropertyChannelServiced {
		return "serviced_apartment_projects.ad_weight desc, "
	}

	return "property_sale_listings.ad_weight desc, "
}

// 38. propertyAreaColumn returns the selected sale area column.
func propertyAreaColumn(areaMode string) string {
	if strings.TrimSpace(areaMode) == "gross" {
		return "COALESCE(property_sale_listings.gross_area_sqft, property_sale_listings.usable_area_sqft)"
	}

	return "property_sale_listings.usable_area_sqft"
}

// 39. propertyDistrictsForRegion returns accepted district codes for property regions.
func propertyDistrictsForRegion(value string) ([]string, bool) {
	switch strings.TrimSpace(value) {
	case "hong_kong_island":
		return []string{"hong_kong_island", "central_western", "wan_chai", "eastern", "southern"}, true
	case "kowloon":
		return []string{"kowloon", "yau_tsim_mong", "sham_shui_po", "kowloon_city", "wong_tai_sin", "kwun_tong"}, true
	case "new_territories":
		return []string{"new_territories", "kwai_tsing", "tsuen_wan", "tuen_mun", "yuen_long", "north", "tai_po", "sha_tin", "sai_kung"}, true
	case "outlying_islands":
		return []string{"outlying_islands", "islands"}, true
	default:
		return nil, false
	}
}

// 40. normalizePropertyTransactionType defaults legacy listings to sale.
func normalizePropertyTransactionType(value string) string {
	if strings.TrimSpace(value) == "rent" {
		return "rent"
	}

	return "sale"
}

// 33. normalizePropertyAreaMode defaults area filtering to usable area.
func normalizePropertyAreaMode(value string) string {
	if strings.TrimSpace(value) == "gross" {
		return "gross"
	}

	return "usable"
}

// 34. normalizePropertyLocationScope defaults listings to local.
func normalizePropertyLocationScope(value string) string {
	if strings.TrimSpace(value) == "overseas" {
		return "overseas"
	}

	return "local"
}

// 35. normalizePropertyListingCategory defaults listings to standard.
func normalizePropertyListingCategory(value string) string {
	switch strings.TrimSpace(value) {
	case "new_development", "developer_project", "multi_unit":
		return strings.TrimSpace(value)
	default:
		return "standard"
	}
}

// 36. propertyAdPackageConfig defines persisted advertising package metadata.
type propertyAdPackageConfig struct {
	Code         string
	Weight       int
	PriceHKD     float64
	PricePoints  int64
	DurationDays int
}

// 37. propertyAdPackage returns configured advertising package values.
func propertyAdPackage(code string) propertyAdPackageConfig {
	switch normalizePropertyAdPackageCode(code) {
	case "featured":
		return propertyAdPackageConfig{Code: "featured", Weight: 1, PriceHKD: 800, PricePoints: 800, DurationDays: 30}
	case "premium":
		return propertyAdPackageConfig{Code: "premium", Weight: 2, PriceHKD: 1500, PricePoints: 1500, DurationDays: 30}
	case "fast_sale":
		return propertyAdPackageConfig{Code: "fast_sale", Weight: 3, PriceHKD: 1200, PricePoints: 1200, DurationDays: 15}
	default:
		return propertyAdPackageConfig{Code: "basic", Weight: 0, PriceHKD: 600, PricePoints: 600, DurationDays: 30}
	}
}

// 38. servicedApartmentAdPackage returns serviced apartment advertising values.
func servicedApartmentAdPackage(code string) propertyAdPackageConfig {
	config := propertyAdPackage(code)
	config.PricePoints += 200
	return config
}

// 39. normalizePropertyAdPackageCode defaults package code to basic.
func normalizePropertyAdPackageCode(value string) string {
	switch strings.TrimSpace(value) {
	case "featured", "premium", "fast_sale":
		return strings.TrimSpace(value)
	default:
		return "basic"
	}
}

// 40. buildPublicFloorLabel returns a safe floor label.
func buildPublicFloorLabel(rawFloor string, floorZone string, totalFloors int) string {
	zone := resolveFloorZone(rawFloor, floorZone, totalFloors)
	switch zone {
	case "high":
		return "高層"
	case "middle":
		return "中層"
	default:
		return "低層"
	}
}

// 40. resolveFloorZone calculates low, middle, or high floor group.
func resolveFloorZone(rawFloor string, floorZone string, totalFloors int) string {
	normalizedZone := strings.TrimSpace(floorZone)
	if isAllowedPropertyValue(normalizedZone, []string{"low", "middle", "high"}) {
		return normalizedZone
	}
	floor := parsePositiveFloor(rawFloor)
	if floor <= 0 || totalFloors <= 0 {
		return "middle"
	}
	lowMax := totalFloors / 3
	highMin := (totalFloors * 2 / 3) + 1
	if floor <= lowMax {
		return "low"
	}
	if floor >= highMin {
		return "high"
	}
	return "middle"
}

// 41. buildFloorDisplayRange returns a public floor range.
func buildFloorDisplayRange(rawFloor string, floorZone string, totalFloors int) string {
	if totalFloors <= 0 {
		return ""
	}
	zone := resolveFloorZone(rawFloor, floorZone, totalFloors)
	start := 1
	end := totalFloors / 3
	if zone == "middle" {
		start = totalFloors/3 + 1
		end = totalFloors * 2 / 3
	}
	if zone == "high" {
		start = totalFloors*2/3 + 1
		end = totalFloors
	}
	floor := parsePositiveFloor(rawFloor)
	if floor > 0 {
		start = floor - 4
		if start < 1 {
			start = 1
		}
		end = floor
		if end > totalFloors {
			end = totalFloors
		}
	}
	return fmt.Sprintf("%d-%d|%d/F", start, end, totalFloors)
}

// 42. buildPublicLocationText builds the public location snippet.
func buildPublicLocationText(blockName string, unitName string, showUnit bool, rawFloor string, floorZone string, totalFloors int) string {
	parts := []string{}
	if strings.TrimSpace(blockName) != "" {
		parts = append(parts, strings.TrimSpace(blockName))
	}
	parts = append(parts, buildPublicFloorLabel(rawFloor, floorZone, totalFloors))
	if showUnit && strings.TrimSpace(unitName) != "" {
		parts = append(parts, strings.TrimSpace(unitName))
	} else if !showUnit {
		displayRange := buildFloorDisplayRange(rawFloor, floorZone, totalFloors)
		if displayRange != "" {
			parts[len(parts)-1] = parts[len(parts)-1] + "(" + displayRange + ")"
		}
	}
	return strings.Join(parts, "")
}

// 43. publicUnitName returns unit name only when owner allows public display.
func publicUnitName(unitName string, showUnit bool) string {
	if !showUnit {
		return ""
	}
	return strings.TrimSpace(unitName)
}

// 44. parsePositiveFloor extracts a positive numeric floor.
func parsePositiveFloor(value string) int {
	var digits strings.Builder
	for _, item := range value {
		if item >= '0' && item <= '9' {
			digits.WriteRune(item)
		}
	}
	var floor int
	if _, err := fmt.Sscanf(digits.String(), "%d", &floor); err != nil {
		return 0
	}
	return floor
}

// 45. firstNonBlank returns the first non-empty string.
func firstNonBlank(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

// 46. fallbackPropertyPublisherIdentity returns a stable identity.
func fallbackPropertyPublisherIdentity(identity string) string {
	if strings.TrimSpace(identity) == "" {
		return "owner"
	}

	return strings.TrimSpace(identity)
}

// 46. publisherRoleLabel returns display label for owner type.
func publisherRoleLabel(identity string) string {
	switch strings.TrimSpace(identity) {
	case "agent", "professional_seller":
		return "代理人"
	default:
		return "業主"
	}
}

// 47. isAllowedPropertyValue validates enum-like input.
func isAllowedPropertyValue(value string, allowedValues []string) bool {
	normalizedValue := strings.TrimSpace(value)
	for _, allowedValue := range allowedValues {
		if normalizedValue == allowedValue {
			return true
		}
	}

	return false
}

// 48. isAllowedPropertyDistrict validates district input with existing marketplace set.
func isAllowedPropertyDistrict(value string) bool {
	return isAllowedSecondhandDistrict(value)
}

// 49. buildPropertyWhatsAppURL creates a WhatsApp deep link from an international phone number.
func buildPropertyWhatsAppURL(phone string, title string, listingPublicID string, baseURL string, channel PropertyChannel) string {
	digits := strings.NewReplacer("+", "", " ", "", "-", "", "(", "", ")", "").Replace(phone)
	return "https://wa.me/" + digits
}

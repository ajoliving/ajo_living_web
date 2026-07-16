/*
 * Property handler support methods.
 * 1. Map HTTP payloads and queries into property service params.
 * 2. Share lifecycle and contact operations across property channels.
 */
package handler

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/service"
)

// 1. listPublic returns public listings for a channel.
func (h *PropertyHandler) listPublic(c *gin.Context, channel service.PropertyChannel) {
	page, pageSize := parsePagination(c)
	filters := service.PropertyListFilters{
		Page:                  page,
		PageSize:              pageSize,
		Keyword:               strings.TrimSpace(c.Query("keyword")),
		RegionCode:            strings.TrimSpace(c.Query("region_code")),
		DistrictCode:          strings.TrimSpace(c.Query("district_code")),
		TransactionType:       strings.TrimSpace(c.Query("transaction_type")),
		PropertyType:          strings.TrimSpace(c.Query("property_type")),
		RentalType:            strings.TrimSpace(c.Query("rental_type")),
		RenovationType:        strings.TrimSpace(c.Query("renovation_type")),
		AreaMode:              strings.TrimSpace(c.Query("area_mode")),
		PriceMode:             strings.TrimSpace(c.DefaultQuery("price_mode", "monthly")),
		FeatureTags:           parseCSVQuery(c.Query("feature_tags")),
		FacilityTags:          parseCSVQuery(c.Query("facility_tags")),
		ServiceTags:           parseCSVQuery(c.Query("service_tags")),
		PublisherIdentityType: strings.TrimSpace(c.Query("publisher_identity_type")),
		HasMedia:              parseBoolQuery(c, "has_media"),
		IsNew:                 parseOptionalBoolQuery(c, "is_new"),
		SortBy:                strings.TrimSpace(c.DefaultQuery("sort_by", "latest")),
	}
	if value, err := strconv.ParseFloat(strings.TrimSpace(c.Query("min_price_hkd")), 64); err == nil {
		filters.MinPriceHKD = &value
	}
	if value, err := strconv.ParseFloat(strings.TrimSpace(c.Query("max_price_hkd")), 64); err == nil {
		filters.MaxPriceHKD = &value
	}
	if value, err := strconv.Atoi(strings.TrimSpace(c.Query("min_area_sqft"))); err == nil {
		filters.MinAreaSqft = &value
	}
	if value, err := strconv.Atoi(strings.TrimSpace(c.Query("max_area_sqft"))); err == nil {
		filters.MaxAreaSqft = &value
	}
	if value, err := strconv.Atoi(strings.TrimSpace(c.Query("bedroom_count"))); err == nil {
		filters.BedroomCount = &value
	}

	items, pagination, err := h.propertyService.ListPublicProperties(c.Request.Context(), channel, filters)
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, gin.H{"items": items, "pagination": pagination})
}

// 2. myListings returns current user's listings for a channel.
func (h *PropertyHandler) myListings(c *gin.Context, channel service.PropertyChannel) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	page, pageSize := parsePagination(c)
	filters := service.PropertyListFilters{
		Page:         page,
		PageSize:     pageSize,
		Status:       strings.TrimSpace(c.Query("status")),
		Keyword:      strings.TrimSpace(c.Query("keyword")),
		DistrictCode: strings.TrimSpace(c.Query("district_code")),
		SortBy:       strings.TrimSpace(c.DefaultQuery("sort_by", "latest")),
	}

	items, pagination, err := h.propertyService.ListMyProperties(c.Request.Context(), channel, user.UserID, filters)
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, gin.H{"items": items, "pagination": pagination})
}

// 3. parseCSVQuery normalizes comma-separated query values.
func parseCSVQuery(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		normalized := strings.TrimSpace(part)
		if normalized != "" {
			result = append(result, normalized)
		}
	}

	return result
}

// 4. getDetail returns a property detail.
func (h *PropertyHandler) getDetail(c *gin.Context, channel service.PropertyChannel) {
	current := currentUser(c)
	var userID *int64
	if current != nil {
		userID = &current.UserID
	}

	result, err := h.propertyService.GetPropertyDetail(c.Request.Context(), channel, strings.TrimSpace(c.Param("listingId")), userID)
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, result)
}

// 6. publish publishes a draft listing.
func (h *PropertyHandler) publish(c *gin.Context, channel service.PropertyChannel) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	result, err := h.propertyService.PublishProperty(c.Request.Context(), channel, user.UserID, strings.TrimSpace(c.Param("listingId")))
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, result)
}

// 7. republish republishes an expired listing.
func (h *PropertyHandler) republish(c *gin.Context, channel service.PropertyChannel) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	result, err := h.propertyService.RepublishProperty(c.Request.Context(), channel, user.UserID, strings.TrimSpace(c.Param("listingId")))
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, result)
}

// 8. deactivate hides a listing.
func (h *PropertyHandler) deactivate(c *gin.Context, channel service.PropertyChannel) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}
	listingID := strings.TrimSpace(c.Param("listingId"))
	if err := h.propertyService.DeactivateProperty(c.Request.Context(), channel, user.UserID, listingID); err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, gin.H{"listing_id": listingID, "publication_status": "hidden"})
}

// 9. contactAccess grants contact access to logged-in users.
func (h *PropertyHandler) contactAccess(c *gin.Context, channel service.PropertyChannel) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	result, err := h.propertyService.GrantPropertyContactAccess(
		c.Request.Context(),
		channel,
		user.UserID,
		strings.TrimSpace(c.Param("listingId")),
		c.ClientIP(),
		c.Request.UserAgent(),
	)
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, result)
}

// 10. bindPropertySaleRequest maps a sale request body.
func (h *PropertyHandler) bindPropertySaleRequest(c *gin.Context, listingID string) (service.UpsertPropertySaleParams, bool) {
	return bindPropertySaleRequestFromContext(c, listingID)
}

// 10.1 bindPropertySaleRequestFromContext maps a sale request body.
func bindPropertySaleRequestFromContext(c *gin.Context, listingID string) (service.UpsertPropertySaleParams, bool) {
	var request propertySaleRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid request payload"))
		return service.UpsertPropertySaleParams{}, false
	}
	showUnit := true
	if request.ShowUnit != nil {
		showUnit = *request.ShowUnit
	}

	return service.UpsertPropertySaleParams{
		ListingPublicID:       listingID,
		PropertyNo:            strings.TrimSpace(request.PropertyNo),
		Title:                 strings.TrimSpace(request.Title),
		TitleEn:               strings.TrimSpace(request.TitleEn),
		Summary:               strings.TrimSpace(request.Summary),
		Description:           strings.TrimSpace(request.Description),
		DescriptionEn:         strings.TrimSpace(request.DescriptionEn),
		DistrictCode:          strings.TrimSpace(request.DistrictCode),
		CommunityID:           strings.TrimSpace(request.CommunityID),
		PublisherIdentityType: strings.TrimSpace(request.PublisherIdentityType),
		TransactionType:       strings.TrimSpace(request.TransactionType),
		LocationScope:         strings.TrimSpace(request.LocationScope),
		ListingCategory:       strings.TrimSpace(request.ListingCategory),
		MultiUnitProject:      request.MultiUnitProject,
		PropertyType:          strings.TrimSpace(request.PropertyType),
		RentalType:            strings.TrimSpace(request.RentalType),
		PropertyAttributes:    normalizePropertyAttributeMap(request.PropertyAttributes),
		RenovationType:        strings.TrimSpace(request.RenovationType),
		AgencyCompanyName:     strings.TrimSpace(request.AgencyCompanyName),
		EstateName:            strings.TrimSpace(request.EstateName),
		AddressText:           strings.TrimSpace(request.AddressText),
		AddressTextEn:         strings.TrimSpace(request.AddressTextEn),
		BlockName:             strings.TrimSpace(request.BlockName),
		UnitName:              strings.TrimSpace(request.UnitName),
		ShowUnit:              showUnit,
		Latitude:              request.Latitude,
		Longitude:             request.Longitude,
		AskingPriceHKD:        request.AskingPriceHKD,
		MonthlyRentHKD:        request.MonthlyRentHKD,
		PriceReferenceOnly:    request.PriceReferenceOnly,
		PriceNegotiable:       request.PriceNegotiable,
		AnnualPrepayDiscount:  request.AnnualPrepayDiscount,
		AnnualPrepayOption:    strings.TrimSpace(request.AnnualPrepayOption),
		LeaseStartDate:        strings.TrimSpace(request.LeaseStartDate),
		RentIncluded:          strings.TrimSpace(request.RentIncluded),
		AreaMode:              strings.TrimSpace(request.AreaMode),
		UsableAreaSqft:        request.UsableAreaSqft,
		GrossAreaSqft:         request.GrossAreaSqft,
		BedroomCount:          request.BedroomCount,
		LivingRoomCount:       request.LivingRoomCount,
		BathroomCount:         request.BathroomCount,
		FloorLevel:            strings.TrimSpace(request.FloorLevel),
		FloorRaw:              strings.TrimSpace(request.FloorRaw),
		FloorZone:             strings.TrimSpace(request.FloorZone),
		TotalFloors:           request.TotalFloors,
		Direction:             strings.TrimSpace(request.Direction),
		BuildingAge:           strings.TrimSpace(request.BuildingAge),
		CompletionYear:        request.CompletionYear,
		BuildingTotalFloors:   request.BuildingTotalFloors,
		ManagementCompany:     strings.TrimSpace(request.ManagementCompany),
		KitchenType:           strings.TrimSpace(request.KitchenType),
		CookingMode:           strings.TrimSpace(request.CookingMode),
		ManagementFeeHKD:      request.ManagementFeeHKD,
		VideoURL:              strings.TrimSpace(request.VideoURL),
		VRURL:                 strings.TrimSpace(request.VRURL),
		PrivateNote:           strings.TrimSpace(request.PrivateNote),
		AdPackageCode:         strings.TrimSpace(request.AdPackageCode),
		FeatureTags:           request.FeatureTags,
		ContactMethod:         strings.TrimSpace(request.ContactMethod),
		BusinessStatus:        strings.TrimSpace(request.BusinessStatus),
		ChargeDraftSave:       c.Query("charge_draft") == "true",
		Images:                request.Images,
		Contact:               toPropertyContactInput(request.Contact),
	}, true
}

// 11. normalizePropertyAttributeMap trims dynamic sale attributes.
func normalizePropertyAttributeMap(input map[string]string) map[string]string {
	result := make(map[string]string)
	for key, value := range input {
		cleanKey := strings.TrimSpace(key)
		cleanValue := strings.TrimSpace(value)
		if cleanKey == "" || cleanValue == "" || len(cleanKey) > 80 || len(cleanValue) > 500 {
			continue
		}
		result[cleanKey] = cleanValue
	}

	return result
}

// 12. bindServicedApartmentRequest maps a serviced apartment request body.
func (h *PropertyHandler) bindServicedApartmentRequest(c *gin.Context, listingID string) (service.UpsertServicedApartmentParams, bool) {
	return bindServicedApartmentRequestFromContext(c, listingID)
}

// 11.1 bindServicedApartmentRequestFromContext maps a serviced apartment request body.
func bindServicedApartmentRequestFromContext(c *gin.Context, listingID string) (service.UpsertServicedApartmentParams, bool) {
	var request servicedApartmentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid request payload"))
		return service.UpsertServicedApartmentParams{}, false
	}

	return service.UpsertServicedApartmentParams{
		ListingPublicID:       listingID,
		Title:                 strings.TrimSpace(request.Title),
		Summary:               strings.TrimSpace(request.Summary),
		Description:           strings.TrimSpace(request.Description),
		DescriptionEn:         strings.TrimSpace(request.DescriptionEn),
		DistrictCode:          strings.TrimSpace(request.DistrictCode),
		CommunityID:           strings.TrimSpace(request.CommunityID),
		PublisherIdentityType: strings.TrimSpace(request.PublisherIdentityType),
		ProjectName:           strings.TrimSpace(request.ProjectName),
		ProjectNameEn:         strings.TrimSpace(request.ProjectNameEn),
		ProjectAttributes:     normalizePropertyAttributeMap(request.ProjectAttributes),
		AddressText:           strings.TrimSpace(request.AddressText),
		AddressTextEn:         strings.TrimSpace(request.AddressTextEn),
		WebsiteURL:            strings.TrimSpace(request.WebsiteURL),
		WhatsApp:              strings.TrimSpace(request.WhatsApp),
		Fax:                   strings.TrimSpace(request.Fax),
		ServiceIntro:          strings.TrimSpace(request.ServiceIntro),
		ServiceIntroEn:        strings.TrimSpace(request.ServiceIntroEn),
		BenefitsText:          strings.TrimSpace(request.BenefitsText),
		BenefitsTextEn:        strings.TrimSpace(request.BenefitsTextEn),
		ExtraChargesText:      strings.TrimSpace(request.ExtraChargesText),
		ExtraChargesTextEn:    strings.TrimSpace(request.ExtraChargesTextEn),
		LowestMonthlyRentHKD:  request.LowestMonthlyRentHKD,
		HighestMonthlyRentHKD: request.HighestMonthlyRentHKD,
		LowestDailyRentHKD:    request.LowestDailyRentHKD,
		PriceReferenceOnly:    request.PriceReferenceOnly,
		PriceNegotiable:       request.PriceNegotiable,
		MinUsableAreaSqft:     request.MinUsableAreaSqft,
		MaxUsableAreaSqft:     request.MaxUsableAreaSqft,
		MinLeaseMonths:        request.MinLeaseMonths,
		MinStayValue:          request.MinStayValue,
		MinStayUnit:           strings.TrimSpace(request.MinStayUnit),
		LocationScope:         strings.TrimSpace(request.LocationScope),
		ListingCategory:       strings.TrimSpace(request.ListingCategory),
		MultiUnitProject:      request.MultiUnitProject,
		AdPackageCode:         strings.TrimSpace(request.AdPackageCode),
		FacilityTags:          request.FacilityTags,
		ServiceTags:           request.ServiceTags,
		RoomTypes:             request.RoomTypes,
		ContactMethod:         strings.TrimSpace(request.ContactMethod),
		BusinessStatus:        strings.TrimSpace(request.BusinessStatus),
		Images:                request.Images,
		Contact:               toPropertyContactInput(request.Contact),
	}, true
}

// 13. toPropertyContactInput maps contact fields.
func toPropertyContactInput(request propertyContactRequest) service.PropertyContactInput {
	return service.PropertyContactInput{
		ContactNameZH:     strings.TrimSpace(request.ContactNameZH),
		ContactNameEN:     strings.TrimSpace(request.ContactNameEN),
		Phone:             strings.TrimSpace(request.Phone),
		Phone2:            strings.TrimSpace(request.Phone2),
		WhatsApp:          strings.TrimSpace(request.WhatsApp),
		WeChat:            strings.TrimSpace(request.WeChat),
		Email:             strings.TrimSpace(request.Email),
		ContactAttributes: normalizePropertyAttributeMap(request.ContactAttributes),
		ShowPhone:         request.ShowPhone,
		ShowWhatsApp:      request.ShowWhatsApp,
		ShowChat:          request.ShowChat,
		ShowInquiryForm:   request.ShowInquiryForm,
	}
}

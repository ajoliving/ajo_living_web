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
		Page:         page,
		PageSize:     pageSize,
		Keyword:      strings.TrimSpace(c.Query("keyword")),
		DistrictCode: strings.TrimSpace(c.Query("district_code")),
		SortBy:       strings.TrimSpace(c.DefaultQuery("sort_by", "latest")),
	}
	if value, err := strconv.ParseFloat(strings.TrimSpace(c.Query("min_price_hkd")), 64); err == nil {
		filters.MinPriceHKD = &value
	}
	if value, err := strconv.ParseFloat(strings.TrimSpace(c.Query("max_price_hkd")), 64); err == nil {
		filters.MaxPriceHKD = &value
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

// 3. getDetail returns a property detail.
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

// 4. publish publishes a draft listing.
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

// 5. republish republishes an expired listing.
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

// 6. deactivate hides a listing.
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

// 7. contactAccess grants contact access to logged-in users.
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

// 8. bindPropertySaleRequest maps a sale request body.
func (h *PropertyHandler) bindPropertySaleRequest(c *gin.Context, listingID string) (service.UpsertPropertySaleParams, bool) {
	var request propertySaleRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid request payload"))
		return service.UpsertPropertySaleParams{}, false
	}

	return service.UpsertPropertySaleParams{
		ListingPublicID:       listingID,
		Title:                 strings.TrimSpace(request.Title),
		Summary:               strings.TrimSpace(request.Summary),
		Description:           strings.TrimSpace(request.Description),
		DistrictCode:          strings.TrimSpace(request.DistrictCode),
		CommunityID:           strings.TrimSpace(request.CommunityID),
		PublisherIdentityType: strings.TrimSpace(request.PublisherIdentityType),
		PropertyType:          strings.TrimSpace(request.PropertyType),
		EstateName:            strings.TrimSpace(request.EstateName),
		AddressText:           strings.TrimSpace(request.AddressText),
		AskingPriceHKD:        request.AskingPriceHKD,
		UsableAreaSqft:        request.UsableAreaSqft,
		GrossAreaSqft:         request.GrossAreaSqft,
		BedroomCount:          request.BedroomCount,
		LivingRoomCount:       request.LivingRoomCount,
		BathroomCount:         request.BathroomCount,
		FloorLevel:            strings.TrimSpace(request.FloorLevel),
		Direction:             strings.TrimSpace(request.Direction),
		BuildingAge:           strings.TrimSpace(request.BuildingAge),
		FeatureTags:           request.FeatureTags,
		ContactMethod:         strings.TrimSpace(request.ContactMethod),
		BusinessStatus:        strings.TrimSpace(request.BusinessStatus),
		Images:                request.Images,
		Contact:               toPropertyContactInput(request.Contact),
	}, true
}

// 9. bindServicedApartmentRequest maps a serviced apartment request body.
func (h *PropertyHandler) bindServicedApartmentRequest(c *gin.Context, listingID string) (service.UpsertServicedApartmentParams, bool) {
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
		DistrictCode:          strings.TrimSpace(request.DistrictCode),
		CommunityID:           strings.TrimSpace(request.CommunityID),
		PublisherIdentityType: strings.TrimSpace(request.PublisherIdentityType),
		ProjectName:           strings.TrimSpace(request.ProjectName),
		AddressText:           strings.TrimSpace(request.AddressText),
		LowestMonthlyRentHKD:  request.LowestMonthlyRentHKD,
		MinLeaseMonths:        request.MinLeaseMonths,
		FacilityTags:          request.FacilityTags,
		ServiceTags:           request.ServiceTags,
		RoomTypes:             request.RoomTypes,
		ContactMethod:         strings.TrimSpace(request.ContactMethod),
		BusinessStatus:        strings.TrimSpace(request.BusinessStatus),
		Images:                request.Images,
		Contact:               toPropertyContactInput(request.Contact),
	}, true
}

// 10. toPropertyContactInput maps contact fields.
func toPropertyContactInput(request propertyContactRequest) service.PropertyContactInput {
	return service.PropertyContactInput{
		Phone:           strings.TrimSpace(request.Phone),
		WhatsApp:        strings.TrimSpace(request.WhatsApp),
		Email:           strings.TrimSpace(request.Email),
		ShowPhone:       request.ShowPhone,
		ShowWhatsApp:    request.ShowWhatsApp,
		ShowChat:        request.ShowChat,
		ShowInquiryForm: request.ShowInquiryForm,
	}
}

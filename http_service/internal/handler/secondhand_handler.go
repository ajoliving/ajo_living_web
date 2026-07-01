/*
 * Secondhand listing HTTP handlers.
 * 1. Bind public browse and owner management requests.
 * 2. Delegate secondhand listing logic to the service layer.
 */
package handler

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/service"
)

// 1. SecondhandHandler handles secondhand listing endpoints.
type SecondhandHandler struct {
	secondhandService *service.SecondhandService
}

// 2. secondhandRequest defines the create and update payload.
type secondhandRequest struct {
	Title                 string                      `json:"title" binding:"required"`
	Summary               string                      `json:"summary"`
	Description           string                      `json:"description"`
	DistrictCode          string                      `json:"district_code" binding:"required"`
	CommunityID           string                      `json:"community_id"`
	CommunityName         string                      `json:"community_name"`
	PublisherIdentityType string                      `json:"publisher_identity_type"`
	CategoryCode          string                      `json:"category_code" binding:"required"`
	PriceMode             string                      `json:"price_mode" binding:"required"`
	PriceHKD              *float64                    `json:"price_hkd"`
	ConditionLevel        string                      `json:"condition_level" binding:"required"`
	DimensionText         string                      `json:"dimension_text"`
	PickupRegionCode      string                      `json:"pickup_region_code" binding:"required"`
	PickupLocationText    string                      `json:"pickup_location_text"`
	DeliveryTags          []string                    `json:"delivery_tags"`
	VisibilityScope       string                      `json:"visibility_scope" binding:"required"`
	ContactMethod         string                      `json:"contact_method" binding:"required"`
	BusinessStatus        string                      `json:"business_status"`
	Images                []service.ListingImageInput `json:"images"`
	Contact               secondhandContactRequest    `json:"contact"`
}

// 3. secondhandContactRequest defines the contact payload for create and update.
type secondhandContactRequest struct {
	Phone           string `json:"phone"`
	WhatsApp        string `json:"whatsapp"`
	Email           string `json:"email"`
	ShowPhone       bool   `json:"show_phone"`
	ShowWhatsApp    bool   `json:"show_whatsapp"`
	ShowChat        bool   `json:"show_chat"`
	ShowInquiryForm bool   `json:"show_inquiry_form"`
}

// 4. NewSecondhandHandler creates a secondhand handler instance.
func NewSecondhandHandler(secondhandService *service.SecondhandService) *SecondhandHandler {
	return &SecondhandHandler{secondhandService: secondhandService}
}

// 5. ListPublic returns public secondhand listings.
func (h *SecondhandHandler) ListPublic(c *gin.Context) {
	page, pageSize := parsePagination(c)
	current := currentUser(c)

	var minPrice *float64
	var maxPrice *float64
	if value, err := strconv.ParseFloat(strings.TrimSpace(c.Query("min_price_hkd")), 64); err == nil {
		minPrice = &value
	}
	if value, err := strconv.ParseFloat(strings.TrimSpace(c.Query("max_price_hkd")), 64); err == nil {
		maxPrice = &value
	}

	filters := service.SecondhandListFilters{
		Page:            page,
		PageSize:        pageSize,
		Keyword:         strings.TrimSpace(c.Query("keyword")),
		CategoryCode:    strings.TrimSpace(c.Query("category_code")),
		RegionCode:      strings.TrimSpace(c.Query("region_code")),
		DistrictCode:    strings.TrimSpace(c.Query("district_code")),
		PriceMode:       strings.TrimSpace(c.Query("price_mode")),
		ConditionLevel:  strings.TrimSpace(c.Query("condition_level")),
		VisibilityScope: strings.TrimSpace(c.Query("visibility_scope")),
		MinPriceHKD:     minPrice,
		MaxPriceHKD:     maxPrice,
		HasMedia:        parseBoolQuery(c, "has_media"),
		OnlyBuilding:    parseBoolQuery(c, "only_building"),
		OnlyFree:        parseBoolQuery(c, "only_free"),
		ExcludeFree:     parseBoolQuery(c, "exclude_free"),
		SortBy:          strings.TrimSpace(c.DefaultQuery("sort_by", "latest")),
	}
	if current != nil {
		filters.CommunityID = current.PrimaryCommunityID
		filters.ViewerUserID = &current.UserID
	}

	items, pagination, err := h.secondhandService.ListPublicSecondhand(c.Request.Context(), filters)
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, gin.H{"items": items, "pagination": pagination})
}

// 6. GetDetail returns a secondhand detail payload.
func (h *SecondhandHandler) GetDetail(c *gin.Context) {
	current := currentUser(c)
	var userID *int64
	var communityID *int64
	if current != nil {
		userID = &current.UserID
		communityID = current.PrimaryCommunityID
	}

	result, err := h.secondhandService.GetSecondhandDetail(c.Request.Context(), strings.TrimSpace(c.Param("listingId")), userID, communityID)
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, result)
}

// 9. Create creates a secondhand draft listing.
func (h *SecondhandHandler) Create(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	params, ok := h.bindSecondhandRequest(c, "")
	if !ok {
		return
	}
	params.OwnerUserID = user.UserID

	result, err := h.secondhandService.CreateSecondhandListing(c.Request.Context(), params)
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, result)
}

// 10. Update updates an owned secondhand listing.
func (h *SecondhandHandler) Update(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	params, ok := h.bindSecondhandRequest(c, strings.TrimSpace(c.Param("listingId")))
	if !ok {
		return
	}
	params.OwnerUserID = user.UserID

	result, err := h.secondhandService.UpdateSecondhandListing(c.Request.Context(), params)
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, result)
}

// 11. Publish publishes a secondhand draft.
func (h *SecondhandHandler) Publish(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	result, err := h.secondhandService.PublishSecondhandListing(c.Request.Context(), user.UserID, strings.TrimSpace(c.Param("listingId")))
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, result)
}

// 12. Republish republishes an expired secondhand listing.
func (h *SecondhandHandler) Republish(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	result, err := h.secondhandService.RepublishSecondhandListing(c.Request.Context(), user.UserID, strings.TrimSpace(c.Param("listingId")))
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, result)
}

// 13. Renew renews an active secondhand listing.
func (h *SecondhandHandler) Renew(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	result, err := h.secondhandService.RenewSecondhandListing(c.Request.Context(), user.UserID, strings.TrimSpace(c.Param("listingId")))
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, result)
}

// 14. MarkSold marks a secondhand listing as sold.
func (h *SecondhandHandler) MarkSold(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	if err := h.secondhandService.MarkSold(c.Request.Context(), user.UserID, strings.TrimSpace(c.Param("listingId"))); err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, gin.H{"listing_id": strings.TrimSpace(c.Param("listingId")), "business_status": "sold"})
}

// 15. Deactivate hides a secondhand listing.
func (h *SecondhandHandler) Deactivate(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	if err := h.secondhandService.Deactivate(c.Request.Context(), user.UserID, strings.TrimSpace(c.Param("listingId"))); err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, gin.H{"listing_id": strings.TrimSpace(c.Param("listingId")), "publication_status": "hidden"})
}

// 16. MyListings returns owned secondhand listings.
func (h *SecondhandHandler) MyListings(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	page, pageSize := parsePagination(c)
	items, pagination, err := h.secondhandService.ListMySecondhand(c.Request.Context(), user.UserID, service.MySecondhandFilters{
		Page:     page,
		PageSize: pageSize,
		Status:   strings.TrimSpace(c.Query("status")),
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, gin.H{"items": items, "pagination": pagination})
}

// 17. MyFavorites returns saved secondhand listings.
func (h *SecondhandHandler) MyFavorites(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	page, pageSize := parsePagination(c)
	items, pagination, err := h.secondhandService.ListFavoriteSecondhand(c.Request.Context(), user.UserID, user.PrimaryCommunityID, service.MySecondhandFilters{
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, gin.H{"items": items, "pagination": pagination})
}

// 18. Favorite adds a secondhand listing to current member favorites.
func (h *SecondhandHandler) Favorite(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	result, err := h.secondhandService.AddFavoriteSecondhand(c.Request.Context(), user.UserID, user.PrimaryCommunityID, strings.TrimSpace(c.Param("listingId")))
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, result)
}

// 19. Unfavorite removes a secondhand listing from current member favorites.
func (h *SecondhandHandler) Unfavorite(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	result, err := h.secondhandService.RemoveFavoriteSecondhand(c.Request.Context(), user.UserID, strings.TrimSpace(c.Param("listingId")))
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, result)
}

// 20. SettingsListings returns all secondhand listings for the settings page.
func (h *SecondhandHandler) SettingsListings(c *gin.Context) {
	if currentUser(c) == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	page, pageSize := parsePagination(c)
	items, pagination, err := h.secondhandService.ListSettingsSecondhand(c.Request.Context(), service.SecondhandSettingsListFilters{
		Page:         page,
		PageSize:     pageSize,
		Keyword:      strings.TrimSpace(c.Query("keyword")),
		CategoryCode: strings.TrimSpace(c.Query("category_code")),
		Status:       strings.TrimSpace(c.Query("status")),
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, gin.H{"items": items, "pagination": pagination})
}

// 21. SettingsMarkSold marks any secondhand listing as sold.
func (h *SecondhandHandler) SettingsMarkSold(c *gin.Context) {
	if currentUser(c) == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	listingID := strings.TrimSpace(c.Param("listingId"))
	if err := h.secondhandService.MarkSoldForSettings(c.Request.Context(), listingID); err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, gin.H{"listing_id": listingID, "business_status": "sold"})
}

// 22. SettingsDeactivate hides any secondhand listing.
func (h *SecondhandHandler) SettingsDeactivate(c *gin.Context) {
	if currentUser(c) == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	listingID := strings.TrimSpace(c.Param("listingId"))
	if err := h.secondhandService.DeactivateForSettings(c.Request.Context(), listingID); err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, gin.H{"listing_id": listingID, "publication_status": "hidden"})
}

// 25. ContactAccess validates contact access and returns allowed payload.
func (h *SecondhandHandler) ContactAccess(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	result, err := h.secondhandService.GrantContactAccess(
		c.Request.Context(),
		user.UserID,
		user.PrimaryCommunityID,
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

// 26. bindSecondhandRequest binds and normalizes secondhand create or update payloads.
func (h *SecondhandHandler) bindSecondhandRequest(c *gin.Context, listingPublicID string) (service.UpsertSecondhandParams, bool) {
	var request secondhandRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid request payload"))
		return service.UpsertSecondhandParams{}, false
	}

	return service.UpsertSecondhandParams{
		ListingPublicID:       listingPublicID,
		Title:                 strings.TrimSpace(request.Title),
		Summary:               strings.TrimSpace(request.Summary),
		Description:           strings.TrimSpace(request.Description),
		DistrictCode:          strings.TrimSpace(request.DistrictCode),
		CommunityID:           strings.TrimSpace(request.CommunityID),
		CommunityName:         strings.TrimSpace(request.CommunityName),
		PublisherIdentityType: strings.TrimSpace(request.PublisherIdentityType),
		CategoryCode:          strings.TrimSpace(request.CategoryCode),
		PriceMode:             strings.TrimSpace(request.PriceMode),
		PriceHKD:              request.PriceHKD,
		ConditionLevel:        strings.TrimSpace(request.ConditionLevel),
		DimensionText:         strings.TrimSpace(request.DimensionText),
		PickupRegionCode:      strings.TrimSpace(request.PickupRegionCode),
		PickupLocationText:    strings.TrimSpace(request.PickupLocationText),
		DeliveryTags:          request.DeliveryTags,
		VisibilityScope:       strings.TrimSpace(request.VisibilityScope),
		ContactMethod:         strings.TrimSpace(request.ContactMethod),
		BusinessStatus:        strings.TrimSpace(request.BusinessStatus),
		ChargeDraftSave:       c.Query("charge_draft") == "true",
		Images:                request.Images,
		Contact: service.ListingContactInput{
			Phone:           strings.TrimSpace(request.Contact.Phone),
			WhatsApp:        strings.TrimSpace(request.Contact.WhatsApp),
			Email:           strings.TrimSpace(request.Contact.Email),
			ShowPhone:       request.Contact.ShowPhone,
			ShowWhatsApp:    request.Contact.ShowWhatsApp,
			ShowChat:        request.Contact.ShowChat,
			ShowInquiryForm: request.Contact.ShowInquiryForm,
		},
	}, true
}

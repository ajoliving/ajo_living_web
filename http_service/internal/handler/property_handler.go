/*
 * Property channel HTTP handlers.
 * 1. Bind property sale and serviced apartment requests.
 * 2. Delegate listing, lifecycle, and contact access flows to the service layer.
 */
package handler

import (
	"strings"

	"github.com/gin-gonic/gin"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/service"
)

// 1. PropertyHandler handles property sale and serviced apartment endpoints.
type PropertyHandler struct {
	propertyService *service.PropertyService
}

// 2. propertyContactRequest defines contact payload.
type propertyContactRequest struct {
	Phone           string `json:"phone"`
	WhatsApp        string `json:"whatsapp"`
	Email           string `json:"email"`
	ShowPhone       bool   `json:"show_phone"`
	ShowWhatsApp    bool   `json:"show_whatsapp"`
	ShowChat        bool   `json:"show_chat"`
	ShowInquiryForm bool   `json:"show_inquiry_form"`
}

// 3. propertySaleRequest defines sale listing create and update payload.
type propertySaleRequest struct {
	Title                 string                      `json:"title" binding:"required"`
	Summary               string                      `json:"summary"`
	Description           string                      `json:"description"`
	DistrictCode          string                      `json:"district_code" binding:"required"`
	CommunityID           string                      `json:"community_id"`
	PublisherIdentityType string                      `json:"publisher_identity_type"`
	PropertyType          string                      `json:"property_type" binding:"required"`
	EstateName            string                      `json:"estate_name"`
	AddressText           string                      `json:"address_text" binding:"required"`
	AskingPriceHKD        float64                     `json:"asking_price_hkd" binding:"required"`
	UsableAreaSqft        int                         `json:"usable_area_sqft" binding:"required"`
	GrossAreaSqft         *int                        `json:"gross_area_sqft"`
	BedroomCount          int                         `json:"bedroom_count"`
	LivingRoomCount       int                         `json:"living_room_count"`
	BathroomCount         int                         `json:"bathroom_count"`
	FloorLevel            string                      `json:"floor_level"`
	Direction             string                      `json:"direction"`
	BuildingAge           string                      `json:"building_age"`
	FeatureTags           []string                    `json:"feature_tags"`
	ContactMethod         string                      `json:"contact_method" binding:"required"`
	BusinessStatus        string                      `json:"business_status"`
	Images                []service.ListingImageInput `json:"images"`
	Contact               propertyContactRequest      `json:"contact"`
}

// 4. servicedApartmentRequest defines serviced apartment create and update payload.
type servicedApartmentRequest struct {
	Title                 string                                   `json:"title" binding:"required"`
	Summary               string                                   `json:"summary"`
	Description           string                                   `json:"description"`
	DistrictCode          string                                   `json:"district_code" binding:"required"`
	CommunityID           string                                   `json:"community_id"`
	PublisherIdentityType string                                   `json:"publisher_identity_type"`
	ProjectName           string                                   `json:"project_name" binding:"required"`
	AddressText           string                                   `json:"address_text" binding:"required"`
	LowestMonthlyRentHKD  float64                                  `json:"lowest_monthly_rent_hkd" binding:"required"`
	MinLeaseMonths        int                                      `json:"min_lease_months" binding:"required"`
	FacilityTags          []string                                 `json:"facility_tags"`
	ServiceTags           []string                                 `json:"service_tags"`
	RoomTypes             []service.ServicedApartmentRoomTypeInput `json:"room_types"`
	ContactMethod         string                                   `json:"contact_method" binding:"required"`
	BusinessStatus        string                                   `json:"business_status"`
	Images                []service.ListingImageInput              `json:"images"`
	Contact               propertyContactRequest                   `json:"contact"`
}

// 5. NewPropertyHandler creates a property handler instance.
func NewPropertyHandler(propertyService *service.PropertyService) *PropertyHandler {
	return &PropertyHandler{propertyService: propertyService}
}

// 6. ListPropertySales returns public sale listings.
func (h *PropertyHandler) ListPropertySales(c *gin.Context) {
	h.listPublic(c, service.PropertyChannelSale)
}

// 7. GetPropertySale returns one sale listing detail.
func (h *PropertyHandler) GetPropertySale(c *gin.Context) {
	h.getDetail(c, service.PropertyChannelSale)
}

// 8. CreatePropertySale creates a sale draft.
func (h *PropertyHandler) CreatePropertySale(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	params, ok := h.bindPropertySaleRequest(c, "")
	if !ok {
		return
	}
	params.OwnerUserID = user.UserID

	result, err := h.propertyService.CreatePropertySale(c.Request.Context(), params)
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 9. UpdatePropertySale updates an owned sale listing.
func (h *PropertyHandler) UpdatePropertySale(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	params, ok := h.bindPropertySaleRequest(c, strings.TrimSpace(c.Param("listingId")))
	if !ok {
		return
	}
	params.OwnerUserID = user.UserID

	result, err := h.propertyService.UpdatePropertySale(c.Request.Context(), params)
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 10. MyPropertySales returns listings owned by the current user.
func (h *PropertyHandler) MyPropertySales(c *gin.Context) {
	h.myListings(c, service.PropertyChannelSale)
}

// 11. ListServicedApartments returns public serviced apartment listings.
func (h *PropertyHandler) ListServicedApartments(c *gin.Context) {
	h.listPublic(c, service.PropertyChannelServiced)
}

// 12. GetServicedApartment returns one serviced apartment detail.
func (h *PropertyHandler) GetServicedApartment(c *gin.Context) {
	h.getDetail(c, service.PropertyChannelServiced)
}

// 13. CreateServicedApartment creates a serviced apartment draft.
func (h *PropertyHandler) CreateServicedApartment(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	params, ok := h.bindServicedApartmentRequest(c, "")
	if !ok {
		return
	}
	params.OwnerUserID = user.UserID

	result, err := h.propertyService.CreateServicedApartment(c.Request.Context(), params)
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 14. UpdateServicedApartment updates an owned serviced apartment listing.
func (h *PropertyHandler) UpdateServicedApartment(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	params, ok := h.bindServicedApartmentRequest(c, strings.TrimSpace(c.Param("listingId")))
	if !ok {
		return
	}
	params.OwnerUserID = user.UserID

	result, err := h.propertyService.UpdateServicedApartment(c.Request.Context(), params)
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 15. MyServicedApartments returns serviced apartment listings owned by the current user.
func (h *PropertyHandler) MyServicedApartments(c *gin.Context) {
	h.myListings(c, service.PropertyChannelServiced)
}

// 16. PublishPropertySale publishes a sale draft.
func (h *PropertyHandler) PublishPropertySale(c *gin.Context) {
	h.publish(c, service.PropertyChannelSale)
}

// 17. RepublishPropertySale republishes an expired sale listing.
func (h *PropertyHandler) RepublishPropertySale(c *gin.Context) {
	h.republish(c, service.PropertyChannelSale)
}

// 18. MarkPropertySaleSold marks a sale listing as sold.
func (h *PropertyHandler) MarkPropertySaleSold(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}
	listingID := strings.TrimSpace(c.Param("listingId"))
	if err := h.propertyService.MarkPropertySold(c.Request.Context(), service.PropertyChannelSale, user.UserID, listingID); err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, gin.H{"listing_id": listingID, "business_status": "sold"})
}

// 19. DeactivatePropertySale hides a sale listing.
func (h *PropertyHandler) DeactivatePropertySale(c *gin.Context) {
	h.deactivate(c, service.PropertyChannelSale)
}

// 20. ContactAccessPropertySale grants sale contact access.
func (h *PropertyHandler) ContactAccessPropertySale(c *gin.Context) {
	h.contactAccess(c, service.PropertyChannelSale)
}

// 21. PublishServicedApartment publishes a serviced apartment draft.
func (h *PropertyHandler) PublishServicedApartment(c *gin.Context) {
	h.publish(c, service.PropertyChannelServiced)
}

// 22. RepublishServicedApartment republishes an expired serviced apartment listing.
func (h *PropertyHandler) RepublishServicedApartment(c *gin.Context) {
	h.republish(c, service.PropertyChannelServiced)
}

// 23. DeactivateServicedApartment hides a serviced apartment listing.
func (h *PropertyHandler) DeactivateServicedApartment(c *gin.Context) {
	h.deactivate(c, service.PropertyChannelServiced)
}

// 24. ContactAccessServicedApartment grants serviced apartment contact access.
func (h *PropertyHandler) ContactAccessServicedApartment(c *gin.Context) {
	h.contactAccess(c, service.PropertyChannelServiced)
}

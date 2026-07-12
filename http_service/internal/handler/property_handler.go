/*
 * Property channel HTTP handlers.
 * 1. Bind property sale and serviced apartment requests.
 * 2. Delegate listing, lifecycle, and contact access flows to the service layer.
 */
package handler

import (
	"strconv"
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
	ContactNameZH     string            `json:"contact_name_zh"`
	ContactNameEN     string            `json:"contact_name_en"`
	Phone             string            `json:"phone"`
	Phone2            string            `json:"phone_2"`
	WhatsApp          string            `json:"whatsapp"`
	WeChat            string            `json:"wechat"`
	Email             string            `json:"email"`
	ContactAttributes map[string]string `json:"contact_attributes"`
	ShowPhone         bool              `json:"show_phone"`
	ShowWhatsApp      bool              `json:"show_whatsapp"`
	ShowChat          bool              `json:"show_chat"`
	ShowInquiryForm   bool              `json:"show_inquiry_form"`
}

// 3. propertySaleRequest defines sale listing create and update payload.
type propertySaleRequest struct {
	Title                 string                      `json:"title"`
	TitleEn               string                      `json:"title_en"`
	Summary               string                      `json:"summary"`
	Description           string                      `json:"description"`
	DescriptionEn         string                      `json:"description_en"`
	DistrictCode          string                      `json:"district_code"`
	CommunityID           string                      `json:"community_id"`
	PublisherIdentityType string                      `json:"publisher_identity_type"`
	PropertyNo            string                      `json:"property_no"`
	TransactionType       string                      `json:"transaction_type"`
	LocationScope         string                      `json:"location_scope"`
	ListingCategory       string                      `json:"listing_category"`
	MultiUnitProject      bool                        `json:"multi_unit_project"`
	PropertyType          string                      `json:"property_type"`
	RentalType            string                      `json:"rental_type"`
	PropertyAttributes    map[string]string           `json:"property_attributes"`
	RenovationType        string                      `json:"renovation_type"`
	AgencyCompanyName     string                      `json:"agency_company_name"`
	EstateName            string                      `json:"estate_name"`
	AddressText           string                      `json:"address_text"`
	AddressTextEn         string                      `json:"address_text_en"`
	BlockName             string                      `json:"block_name"`
	UnitName              string                      `json:"unit_name"`
	ShowUnit              *bool                       `json:"show_unit"`
	Latitude              *float64                    `json:"latitude"`
	Longitude             *float64                    `json:"longitude"`
	AskingPriceHKD        float64                     `json:"asking_price_hkd"`
	MonthlyRentHKD        float64                     `json:"monthly_rent_hkd"`
	PriceReferenceOnly    bool                        `json:"price_reference_only"`
	PriceNegotiable       bool                        `json:"price_negotiable"`
	AnnualPrepayDiscount  bool                        `json:"annual_prepay_discount"`
	AnnualPrepayOption    string                      `json:"annual_prepay_option"`
	LeaseStartDate        string                      `json:"lease_start_date"`
	RentIncluded          string                      `json:"rent_included"`
	AreaMode              string                      `json:"area_mode"`
	UsableAreaSqft        int                         `json:"usable_area_sqft"`
	GrossAreaSqft         *int                        `json:"gross_area_sqft"`
	BedroomCount          int                         `json:"bedroom_count"`
	LivingRoomCount       int                         `json:"living_room_count"`
	BathroomCount         int                         `json:"bathroom_count"`
	FloorLevel            string                      `json:"floor_level"`
	FloorRaw              string                      `json:"floor_raw"`
	FloorZone             string                      `json:"floor_zone"`
	TotalFloors           int                         `json:"total_floors"`
	Direction             string                      `json:"direction"`
	BuildingAge           string                      `json:"building_age"`
	CompletionYear        int                         `json:"completion_year"`
	BuildingTotalFloors   int                         `json:"building_total_floors"`
	ManagementCompany     string                      `json:"management_company"`
	KitchenType           string                      `json:"kitchen_type"`
	CookingMode           string                      `json:"cooking_mode"`
	ManagementFeeHKD      float64                     `json:"management_fee_hkd"`
	VideoURL              string                      `json:"video_url"`
	VRURL                 string                      `json:"vr_url"`
	PrivateNote           string                      `json:"private_note"`
	AdPackageCode         string                      `json:"ad_package_code"`
	FeatureTags           []string                    `json:"feature_tags"`
	ContactMethod         string                      `json:"contact_method"`
	BusinessStatus        string                      `json:"business_status"`
	Images                []service.ListingImageInput `json:"images"`
	Contact               propertyContactRequest      `json:"contact"`
}

// 4. propertyAppointmentRequest defines viewing appointment payload.
type propertyAppointmentRequest struct {
	ContactName     string `json:"contact_name" binding:"required"`
	ContactPhone    string `json:"contact_phone" binding:"required"`
	PreferredTime   string `json:"preferred_time"`
	Message         string `json:"message"`
	AppointmentType string `json:"appointment_type"`
}

// 5. propertyReportRequest defines public report payload.
type propertyReportRequest struct {
	Reason  string `json:"reason" binding:"required"`
	Message string `json:"message"`
}

// 6. servicedApartmentRequest defines serviced apartment create and update payload.
type servicedApartmentRequest struct {
	Title                 string                                   `json:"title" binding:"required"`
	TitleEn               string                                   `json:"title_en"`
	Summary               string                                   `json:"summary"`
	Description           string                                   `json:"description"`
	DescriptionEn         string                                   `json:"description_en"`
	DistrictCode          string                                   `json:"district_code" binding:"required"`
	CommunityID           string                                   `json:"community_id"`
	PublisherIdentityType string                                   `json:"publisher_identity_type"`
	ProjectName           string                                   `json:"project_name" binding:"required"`
	ProjectNameEn         string                                   `json:"project_name_en"`
	ProjectAttributes     map[string]string                        `json:"project_attributes"`
	AddressText           string                                   `json:"address_text" binding:"required"`
	AddressTextEn         string                                   `json:"address_text_en"`
	WebsiteURL            string                                   `json:"website_url"`
	WhatsApp              string                                   `json:"whatsapp"`
	Fax                   string                                   `json:"fax"`
	ServiceIntro          string                                   `json:"service_intro"`
	ServiceIntroEn        string                                   `json:"service_intro_en"`
	BenefitsText          string                                   `json:"benefits_text"`
	BenefitsTextEn        string                                   `json:"benefits_text_en"`
	ExtraChargesText      string                                   `json:"extra_charges_text"`
	ExtraChargesTextEn    string                                   `json:"extra_charges_text_en"`
	LowestMonthlyRentHKD  float64                                  `json:"lowest_monthly_rent_hkd"`
	HighestMonthlyRentHKD float64                                  `json:"highest_monthly_rent_hkd"`
	LowestDailyRentHKD    float64                                  `json:"lowest_daily_rent_hkd"`
	PriceReferenceOnly    bool                                     `json:"price_reference_only"`
	PriceNegotiable       bool                                     `json:"price_negotiable"`
	MinUsableAreaSqft     int                                      `json:"min_usable_area_sqft"`
	MaxUsableAreaSqft     int                                      `json:"max_usable_area_sqft"`
	MinLeaseMonths        int                                      `json:"min_lease_months"`
	MinStayValue          int                                      `json:"min_stay_value"`
	MinStayUnit           string                                   `json:"min_stay_unit"`
	LocationScope         string                                   `json:"location_scope"`
	ListingCategory       string                                   `json:"listing_category"`
	MultiUnitProject      bool                                     `json:"multi_unit_project"`
	AdPackageCode         string                                   `json:"ad_package_code"`
	FacilityTags          []string                                 `json:"facility_tags"`
	ServiceTags           []string                                 `json:"service_tags"`
	RoomTypes             []service.ServicedApartmentRoomTypeInput `json:"room_types"`
	ContactMethod         string                                   `json:"contact_method" binding:"required"`
	BusinessStatus        string                                   `json:"business_status"`
	Images                []service.ListingImageInput              `json:"images"`
	Contact               propertyContactRequest                   `json:"contact"`
}

// 7. NewPropertyHandler creates a property handler instance.
func NewPropertyHandler(propertyService *service.PropertyService) *PropertyHandler {
	return &PropertyHandler{propertyService: propertyService}
}

// 8. ListPropertySales returns public sale listings.
func (h *PropertyHandler) ListPropertySales(c *gin.Context) {
	h.listPublic(c, service.PropertyChannelSale)
}

// 9. GetPropertySale returns one sale listing detail.
func (h *PropertyHandler) GetPropertySale(c *gin.Context) {
	h.getDetail(c, service.PropertyChannelSale)
}

// 10. CreatePropertySale creates a sale draft.
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

// 11. UpdatePropertySale updates an owned sale listing.
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

// 12. MyPropertySales returns listings owned by the current user.
func (h *PropertyHandler) MyPropertySales(c *gin.Context) {
	h.myListings(c, service.PropertyChannelSale)
}

// 13. MyFavoritePropertySales returns saved property sale listings.
func (h *PropertyHandler) MyFavoritePropertySales(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}
	page, pageSize := parsePagination(c)
	items, pagination, err := h.propertyService.ListFavoriteProperties(c.Request.Context(), user.UserID, service.PropertyListFilters{
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, gin.H{"items": items, "pagination": pagination})
}

// 14. ListServicedApartments returns public serviced apartment listings.
func (h *PropertyHandler) ListServicedApartments(c *gin.Context) {
	h.listPublic(c, service.PropertyChannelServiced)
}

// 15. GetServicedApartment returns one serviced apartment detail.
func (h *PropertyHandler) GetServicedApartment(c *gin.Context) {
	h.getDetail(c, service.PropertyChannelServiced)
}

// 16. CreateServicedApartment creates a serviced apartment draft.
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

// 17. UpdateServicedApartment updates an owned serviced apartment listing.
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

// 18. MyServicedApartments returns serviced apartment listings owned by the current user.
func (h *PropertyHandler) MyServicedApartments(c *gin.Context) {
	h.myListings(c, service.PropertyChannelServiced)
}

// 19. PublishPropertySale publishes a sale draft.
func (h *PropertyHandler) PublishPropertySale(c *gin.Context) {
	h.publish(c, service.PropertyChannelSale)
}

// 20. RepublishPropertySale republishes an expired sale listing.
func (h *PropertyHandler) RepublishPropertySale(c *gin.Context) {
	h.republish(c, service.PropertyChannelSale)
}

// 21. MarkPropertySaleSold marks a sale listing as sold.
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

// 22. DeactivatePropertySale hides a sale listing.
func (h *PropertyHandler) DeactivatePropertySale(c *gin.Context) {
	h.deactivate(c, service.PropertyChannelSale)
}

// 23. ContactAccessPropertySale grants sale contact access.
func (h *PropertyHandler) ContactAccessPropertySale(c *gin.Context) {
	h.contactAccess(c, service.PropertyChannelSale)
}

// 24. FavoritePropertySale saves one property sale listing.
func (h *PropertyHandler) FavoritePropertySale(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}
	result, err := h.propertyService.AddFavoriteProperty(c.Request.Context(), user.UserID, strings.TrimSpace(c.Param("listingId")))
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 25. UnfavoritePropertySale removes one property sale listing from favorites.
func (h *PropertyHandler) UnfavoritePropertySale(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}
	result, err := h.propertyService.RemoveFavoriteProperty(c.Request.Context(), user.UserID, strings.TrimSpace(c.Param("listingId")))
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 26. SimilarPropertySales returns related public sale listings.
func (h *PropertyHandler) SimilarPropertySales(c *gin.Context) {
	limit, _ := strconv.Atoi(strings.TrimSpace(c.DefaultQuery("limit", "8")))
	items, err := h.propertyService.ListSimilarProperties(c.Request.Context(), strings.TrimSpace(c.Param("listingId")), limit)
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, gin.H{"items": items})
}

// 27. CreatePropertyAppointment creates a viewing request.
func (h *PropertyHandler) CreatePropertyAppointment(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}
	var request propertyAppointmentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid request payload"))
		return
	}
	result, err := h.propertyService.CreatePropertyAppointment(c.Request.Context(), service.PropertyAppointmentParams{
		UserID:          user.UserID,
		ListingPublicID: strings.TrimSpace(c.Param("listingId")),
		ContactName:     strings.TrimSpace(request.ContactName),
		ContactPhone:    strings.TrimSpace(request.ContactPhone),
		PreferredTime:   strings.TrimSpace(request.PreferredTime),
		Message:         strings.TrimSpace(request.Message),
		AppointmentType: strings.TrimSpace(request.AppointmentType),
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 28. ReportPropertySale creates a public report.
func (h *PropertyHandler) ReportPropertySale(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}
	var request propertyReportRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid request payload"))
		return
	}
	result, err := h.propertyService.ReportProperty(c.Request.Context(), service.PropertyReportParams{
		UserID:          user.UserID,
		ListingPublicID: strings.TrimSpace(c.Param("listingId")),
		Reason:          strings.TrimSpace(request.Reason),
		Message:         strings.TrimSpace(request.Message),
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 29. SearchPropertyAddresses returns building address autocomplete suggestions.
func (h *PropertyHandler) SearchPropertyAddresses(c *gin.Context) {
	limit, _ := strconv.Atoi(strings.TrimSpace(c.Query("limit")))
	result, err := h.propertyService.SearchPropertyAddresses(
		c.Request.Context(),
		strings.TrimSpace(c.Query("keyword")),
		strings.TrimSpace(c.Query("district_code")),
		limit,
	)
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 30. PublishServicedApartment publishes a serviced apartment draft.
func (h *PropertyHandler) PublishServicedApartment(c *gin.Context) {
	h.publish(c, service.PropertyChannelServiced)
}

// 31. RepublishServicedApartment republishes an expired serviced apartment listing.
func (h *PropertyHandler) RepublishServicedApartment(c *gin.Context) {
	h.republish(c, service.PropertyChannelServiced)
}

// 32. DeactivateServicedApartment hides a serviced apartment listing.
func (h *PropertyHandler) DeactivateServicedApartment(c *gin.Context) {
	h.deactivate(c, service.PropertyChannelServiced)
}

// 33. ContactAccessServicedApartment grants serviced apartment contact access.
func (h *PropertyHandler) ContactAccessServicedApartment(c *gin.Context) {
	h.contactAccess(c, service.PropertyChannelServiced)
}

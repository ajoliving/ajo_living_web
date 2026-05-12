/*
 * Staff listing HTTP handlers.
 * 1. Bind staff-only listing management queries.
 * 2. Delegate listing status operations to secondhand and property services.
 */
package handler

import (
	"errors"
	"io"
	"strings"

	"github.com/gin-gonic/gin"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/service"
)

// 1. StaffListingHandler handles staff-only listing management endpoints.
type StaffListingHandler struct {
	secondhandService *service.SecondhandService
	propertyService   *service.PropertyService
}

// 2. staffRenewRequest defines the configurable renewal payload.
type staffRenewRequest struct {
	RenewalDays *int `json:"renewal_days"`
}

// 3. NewStaffListingHandler creates a staff listing handler instance.
func NewStaffListingHandler(secondhandService *service.SecondhandService, propertyService *service.PropertyService) *StaffListingHandler {
	return &StaffListingHandler{
		secondhandService: secondhandService,
		propertyService:   propertyService,
	}
}

// 4. ListSecondhand returns staff-visible secondhand listings.
func (h *StaffListingHandler) ListSecondhand(c *gin.Context) {
	page, pageSize := parsePagination(c)
	items, pagination, err := h.secondhandService.ListStaffSecondhand(c.Request.Context(), service.SecondhandSettingsListFilters{
		Page:     page,
		PageSize: pageSize,
		Keyword:  strings.TrimSpace(c.Query("keyword")),
		Status:   strings.TrimSpace(c.Query("status")),
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, gin.H{"items": items, "pagination": pagination})
}

// 5. PublishSecondhand publishes one secondhand listing.
func (h *StaffListingHandler) PublishSecondhand(c *gin.Context) {
	listingID := strings.TrimSpace(c.Param("listingId"))
	if err := h.secondhandService.PublishForStaff(c.Request.Context(), listingID); err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, gin.H{"listing_id": listingID, "publication_status": "active", "business_status": "available"})
}

// 6. DeactivateSecondhand hides one secondhand listing.
func (h *StaffListingHandler) DeactivateSecondhand(c *gin.Context) {
	listingID := strings.TrimSpace(c.Param("listingId"))
	if err := h.secondhandService.DeactivateForStaff(c.Request.Context(), listingID); err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, gin.H{"listing_id": listingID, "publication_status": "hidden"})
}

// 7. RenewSecondhand renews one secondhand listing.
func (h *StaffListingHandler) RenewSecondhand(c *gin.Context) {
	listingID := strings.TrimSpace(c.Param("listingId"))
	renewalDays, ok := bindStaffRenewalDays(c)
	if !ok {
		return
	}
	if err := h.secondhandService.RenewForStaff(c.Request.Context(), listingID, renewalDays); err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, gin.H{"listing_id": listingID, "publication_status": "active", "business_status": "available"})
}

// 8. ListPropertySales returns staff-visible property sale listings.
func (h *StaffListingHandler) ListPropertySales(c *gin.Context) {
	h.listProperties(c, service.PropertyChannelSale)
}

// 9. ListServicedApartments returns staff-visible serviced apartment listings.
func (h *StaffListingHandler) ListServicedApartments(c *gin.Context) {
	h.listProperties(c, service.PropertyChannelServiced)
}

// 10. PublishPropertySale publishes one property sale listing.
func (h *StaffListingHandler) PublishPropertySale(c *gin.Context) {
	h.publishProperty(c, service.PropertyChannelSale)
}

// 11. DeactivatePropertySale hides one property sale listing.
func (h *StaffListingHandler) DeactivatePropertySale(c *gin.Context) {
	h.deactivateProperty(c, service.PropertyChannelSale)
}

// 12. RenewPropertySale renews one property sale listing.
func (h *StaffListingHandler) RenewPropertySale(c *gin.Context) {
	h.renewProperty(c, service.PropertyChannelSale)
}

// 13. RenewServicedApartment renews one serviced apartment listing.
func (h *StaffListingHandler) RenewServicedApartment(c *gin.Context) {
	h.renewProperty(c, service.PropertyChannelServiced)
}

// 14. listProperties returns staff-visible property listings for one channel.
func (h *StaffListingHandler) listProperties(c *gin.Context, channel service.PropertyChannel) {
	page, pageSize := parsePagination(c)
	items, pagination, err := h.propertyService.ListStaffProperties(c.Request.Context(), channel, service.PropertyListFilters{
		Page:     page,
		PageSize: pageSize,
		Keyword:  strings.TrimSpace(c.Query("keyword")),
		Status:   strings.TrimSpace(c.Query("status")),
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, gin.H{"items": items, "pagination": pagination})
}

// 15. publishProperty publishes one property listing.
func (h *StaffListingHandler) publishProperty(c *gin.Context, channel service.PropertyChannel) {
	listingID := strings.TrimSpace(c.Param("listingId"))
	if err := h.propertyService.PublishPropertyForStaff(c.Request.Context(), channel, listingID); err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, gin.H{"listing_id": listingID, "publication_status": "active", "business_status": "available"})
}

// 16. deactivateProperty hides one property listing.
func (h *StaffListingHandler) deactivateProperty(c *gin.Context, channel service.PropertyChannel) {
	listingID := strings.TrimSpace(c.Param("listingId"))
	if err := h.propertyService.DeactivatePropertyForStaff(c.Request.Context(), channel, listingID); err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, gin.H{"listing_id": listingID, "publication_status": "hidden"})
}

// 17. renewProperty renews one property listing.
func (h *StaffListingHandler) renewProperty(c *gin.Context, channel service.PropertyChannel) {
	listingID := strings.TrimSpace(c.Param("listingId"))
	if err := h.propertyService.RenewPropertyForStaff(c.Request.Context(), channel, listingID); err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, gin.H{"listing_id": listingID, "publication_status": "active", "business_status": "available"})
}

// 18. bindStaffRenewalDays binds optional renewal days from request body.
func bindStaffRenewalDays(c *gin.Context) (int, bool) {
	var request staffRenewRequest
	if c.Request.Body == nil || c.Request.ContentLength == 0 {
		return 0, true
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		if errors.Is(err, io.EOF) {
			return 0, true
		}
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid request payload"))
		return 0, false
	}
	if request.RenewalDays == nil {
		return 0, true
	}
	if *request.RenewalDays <= 0 {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "renewal_days must be positive"))
		return 0, false
	}

	return *request.RenewalDays, true
}

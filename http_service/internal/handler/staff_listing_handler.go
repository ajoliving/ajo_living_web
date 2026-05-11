/*
 * Staff listing HTTP handlers.
 * 1. Bind staff-only listing management queries.
 * 2. Delegate listing status operations to secondhand and property services.
 */
package handler

import (
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

// 2. NewStaffListingHandler creates a staff listing handler instance.
func NewStaffListingHandler(secondhandService *service.SecondhandService, propertyService *service.PropertyService) *StaffListingHandler {
	return &StaffListingHandler{
		secondhandService: secondhandService,
		propertyService:   propertyService,
	}
}

// 3. ListSecondhand returns staff-visible secondhand listings.
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

// 4. PublishSecondhand publishes one secondhand listing.
func (h *StaffListingHandler) PublishSecondhand(c *gin.Context) {
	listingID := strings.TrimSpace(c.Param("listingId"))
	if err := h.secondhandService.PublishForStaff(c.Request.Context(), listingID); err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, gin.H{"listing_id": listingID, "publication_status": "active", "business_status": "available"})
}

// 5. DeactivateSecondhand hides one secondhand listing.
func (h *StaffListingHandler) DeactivateSecondhand(c *gin.Context) {
	listingID := strings.TrimSpace(c.Param("listingId"))
	if err := h.secondhandService.DeactivateForStaff(c.Request.Context(), listingID); err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, gin.H{"listing_id": listingID, "publication_status": "hidden"})
}

// 6. RenewSecondhand renews one secondhand listing.
func (h *StaffListingHandler) RenewSecondhand(c *gin.Context) {
	listingID := strings.TrimSpace(c.Param("listingId"))
	if err := h.secondhandService.RenewForStaff(c.Request.Context(), listingID); err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, gin.H{"listing_id": listingID, "publication_status": "active", "business_status": "available"})
}

// 7. ListPropertySales returns staff-visible property sale listings.
func (h *StaffListingHandler) ListPropertySales(c *gin.Context) {
	h.listProperties(c, service.PropertyChannelSale)
}

// 8. ListServicedApartments returns staff-visible serviced apartment listings.
func (h *StaffListingHandler) ListServicedApartments(c *gin.Context) {
	h.listProperties(c, service.PropertyChannelServiced)
}

// 9. PublishPropertySale publishes one property sale listing.
func (h *StaffListingHandler) PublishPropertySale(c *gin.Context) {
	h.publishProperty(c, service.PropertyChannelSale)
}

// 10. DeactivatePropertySale hides one property sale listing.
func (h *StaffListingHandler) DeactivatePropertySale(c *gin.Context) {
	h.deactivateProperty(c, service.PropertyChannelSale)
}

// 11. RenewPropertySale renews one property sale listing.
func (h *StaffListingHandler) RenewPropertySale(c *gin.Context) {
	h.renewProperty(c, service.PropertyChannelSale)
}

// 12. RenewServicedApartment renews one serviced apartment listing.
func (h *StaffListingHandler) RenewServicedApartment(c *gin.Context) {
	h.renewProperty(c, service.PropertyChannelServiced)
}

// 13. listProperties returns staff-visible property listings for one channel.
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

// 14. publishProperty publishes one property listing.
func (h *StaffListingHandler) publishProperty(c *gin.Context, channel service.PropertyChannel) {
	listingID := strings.TrimSpace(c.Param("listingId"))
	if err := h.propertyService.PublishPropertyForStaff(c.Request.Context(), channel, listingID); err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, gin.H{"listing_id": listingID, "publication_status": "active", "business_status": "available"})
}

// 15. deactivateProperty hides one property listing.
func (h *StaffListingHandler) deactivateProperty(c *gin.Context, channel service.PropertyChannel) {
	listingID := strings.TrimSpace(c.Param("listingId"))
	if err := h.propertyService.DeactivatePropertyForStaff(c.Request.Context(), channel, listingID); err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, gin.H{"listing_id": listingID, "publication_status": "hidden"})
}

// 16. renewProperty renews one property listing.
func (h *StaffListingHandler) renewProperty(c *gin.Context, channel service.PropertyChannel) {
	listingID := strings.TrimSpace(c.Param("listingId"))
	if err := h.propertyService.RenewPropertyForStaff(c.Request.Context(), channel, listingID); err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, gin.H{"listing_id": listingID, "publication_status": "active", "business_status": "available"})
}

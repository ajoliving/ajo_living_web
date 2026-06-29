/*
 * Property route registration.
 * 1. Register property sale and serviced residence write routes.
 * 2. Keep public property listing routes in public route registration.
 */
package router

import (
	"github.com/gin-gonic/gin"

	"ajoliving_web/http_service/internal/handler"
)

// 1. registerPropertyRoutes registers authenticated property listing routes.
func registerPropertyRoutes(api *gin.RouterGroup, propertyHandler *handler.PropertyHandler, requireAuth gin.HandlerFunc) {
	api.POST("/property-sales", requireAuth, propertyHandler.CreatePropertySale)
	api.PATCH("/property-sales/:listingId", requireAuth, propertyHandler.UpdatePropertySale)
	api.POST("/property-sales/:listingId/publish", requireAuth, propertyHandler.PublishPropertySale)
	api.POST("/property-sales/:listingId/republish", requireAuth, propertyHandler.RepublishPropertySale)
	api.POST("/property-sales/:listingId/mark-sold", requireAuth, propertyHandler.MarkPropertySaleSold)
	api.POST("/property-sales/:listingId/deactivate", requireAuth, propertyHandler.DeactivatePropertySale)
	api.POST("/property-sales/:listingId/favorite", requireAuth, propertyHandler.FavoritePropertySale)
	api.DELETE("/property-sales/:listingId/favorite", requireAuth, propertyHandler.UnfavoritePropertySale)
	api.POST("/property-sales/:listingId/appointments", requireAuth, propertyHandler.CreatePropertyAppointment)
	api.POST("/property-sales/:listingId/reports", requireAuth, propertyHandler.ReportPropertySale)
	api.POST("/serviced-apartments", requireAuth, propertyHandler.CreateServicedApartment)
	api.PATCH("/serviced-apartments/:listingId", requireAuth, propertyHandler.UpdateServicedApartment)
	api.POST("/serviced-apartments/:listingId/publish", requireAuth, propertyHandler.PublishServicedApartment)
	api.POST("/serviced-apartments/:listingId/republish", requireAuth, propertyHandler.RepublishServicedApartment)
	api.POST("/serviced-apartments/:listingId/deactivate", requireAuth, propertyHandler.DeactivateServicedApartment)
}

/*
 * Contact route registration.
 * 1. Register rate-limited contact access routes for listing and property modules.
 * 2. Register source listing chat entry routes.
 */
package router

import (
	"time"

	"github.com/gin-gonic/gin"

	"ajoliving_web/http_service/internal/handler"
	"ajoliving_web/http_service/internal/middleware"
)

// 1. registerContactRoutes registers contact access and chat entry routes.
func registerContactRoutes(
	api *gin.RouterGroup,
	secondhandHandler *handler.SecondhandHandler,
	propertyHandler *handler.PropertyHandler,
	chatHandler *handler.ChatHandler,
	requireAuth gin.HandlerFunc,
	limiter *middleware.InMemoryLimiter,
) {
	api.POST("/listings/:listingId/contact-access",
		requireAuth,
		limiter.Limit(20, time.Minute, func(c *gin.Context) string {
			return "contact_access:" + c.ClientIP()
		}),
		secondhandHandler.ContactAccess,
	)
	api.POST("/secondhand/listings/:listingId/contact-access",
		requireAuth,
		limiter.Limit(20, time.Minute, func(c *gin.Context) string {
			return "contact_access:" + c.ClientIP()
		}),
		secondhandHandler.ContactAccess,
	)
	api.POST("/property-sales/:listingId/contact-access",
		requireAuth,
		limiter.Limit(20, time.Minute, func(c *gin.Context) string {
			return "property_sale_contact_access:" + c.ClientIP()
		}),
		propertyHandler.ContactAccessPropertySale,
	)
	api.POST("/serviced-apartments/:listingId/contact-access",
		requireAuth,
		limiter.Limit(20, time.Minute, func(c *gin.Context) string {
			return "serviced_apartment_contact_access:" + c.ClientIP()
		}),
		propertyHandler.ContactAccessServicedApartment,
	)
	api.POST("/listings/:listingId/chats", requireAuth, chatHandler.CreateOrReuse)
	api.POST("/property-sales/:listingId/chats", requireAuth, chatHandler.CreateOrReusePropertySale)
	api.POST("/serviced-apartments/:listingId/chats", requireAuth, chatHandler.CreateOrReuseServicedApartment)
}

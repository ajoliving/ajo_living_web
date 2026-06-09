/*
 * Public route registration.
 * 1. Register health, home content, public marketplace, property, and supermarket routes.
 * 2. Keep unauthenticated route registration separate from member and staff routes.
 */
package router

import (
	"github.com/gin-gonic/gin"

	"ajoliving_web/http_service/internal/handler"
)

// 1. registerPublicRoutes registers health checks and public-facing routes.
func registerPublicRoutes(
	api *gin.RouterGroup,
	healthHandler *handler.HealthHandler,
	userHandler *handler.UserHandler,
	homeContentHandler *handler.HomeContentHandler,
	walletHandler *handler.WalletHandler,
	secondhandHandler *handler.SecondhandHandler,
	propertyHandler *handler.PropertyHandler,
	supermarketOfferHandler *handler.SupermarketOfferHandler,
	optionalAuth gin.HandlerFunc,
) {
	// 1.1 Health and channel metadata routes.
	api.GET("/health", healthHandler.Check)
	api.GET("/channel-home/overview", optionalAuth, userHandler.ChannelHomeOverview)
	api.GET("/meta/communities", userHandler.ListCommunities)

	// 1.2 Home content routes.
	api.GET("/home/content", homeContentHandler.PublicContent)
	api.GET("/home/login-hero", homeContentHandler.LoginHero)

	// 1.3 Public wallet advertisement routes.
	api.GET("/public/ads", walletHandler.ListPublicAds)

	// 1.4 Public secondhand listing routes.
	api.GET("/listings", optionalAuth, secondhandHandler.ListPublic)
	api.GET("/listings/:listingId", optionalAuth, secondhandHandler.GetDetail)
	api.GET("/secondhand/listings", optionalAuth, secondhandHandler.ListPublic)
	api.GET("/secondhand/listings/:listingId", optionalAuth, secondhandHandler.GetDetail)

	// 1.5 Public property and serviced residence routes.
	api.GET("/property-sales", optionalAuth, propertyHandler.ListPropertySales)
	api.GET("/property-sales/:listingId", optionalAuth, propertyHandler.GetPropertySale)
	api.GET("/property-addresses/search", propertyHandler.SearchPropertyAddresses)
	api.GET("/serviced-apartments", optionalAuth, propertyHandler.ListServicedApartments)
	api.GET("/serviced-apartments/:listingId", optionalAuth, propertyHandler.GetServicedApartment)

	// 1.6 Public supermarket offer routes.
	api.GET("/supermarket-offers/summary", supermarketOfferHandler.Summary)
	api.GET("/supermarket-offers/search", supermarketOfferHandler.Search)
	api.GET("/supermarket-offers/products/:code", optionalAuth, supermarketOfferHandler.ProductDetail)
}

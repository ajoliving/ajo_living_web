/*
 * Staff route registration.
 * 1. Register staff user, wallet, listing, property, serviced residence, and settings routes.
 * 2. Keep staff management routes separate from member-facing routes.
 */
package router

import (
	"github.com/gin-gonic/gin"

	"ajoliving_web/http_service/internal/handler"
)

// 1. registerStaffRoutes registers staff-only management endpoints.
func registerStaffRoutes(
	api *gin.RouterGroup,
	staffHandler *handler.StaffHandler,
	staffWalletHandler *handler.StaffWalletHandler,
	staffListingHandler *handler.StaffListingHandler,
	homeContentHandler *handler.HomeContentHandler,
	secondhandHandler *handler.SecondhandHandler,
	chatHandler *handler.ChatHandler,
	requireStaff gin.HandlerFunc,
) {
	// 1.1 Staff account and user management routes.
	api.GET("/staff/me", requireStaff, staffHandler.GetMe)
	api.GET("/staff/roles", requireStaff, staffHandler.ListRoles)
	api.GET("/staff/users", requireStaff, staffHandler.ListUsers)
	api.POST("/staff/users", requireStaff, staffHandler.CreateUser)
	api.PATCH("/staff/users/:userId/role", requireStaff, staffHandler.UpdateUserRole)

	// 1.2 Staff wallet and reward advertisement routes.
	api.GET("/staff/wallet/transactions", requireStaff, staffWalletHandler.ListTransactions)
	api.POST("/staff/wallet/grants", requireStaff, staffWalletHandler.GrantPoints)
	api.GET("/staff/wallet/reward-ads", requireStaff, staffWalletHandler.ListRewardAds)
	api.GET("/staff/wallet/reward-ads/:taskId", requireStaff, staffWalletHandler.GetRewardAd)
	api.POST("/staff/wallet/reward-ads", requireStaff, staffWalletHandler.CreateRewardAd)
	api.PATCH("/staff/wallet/reward-ads/:taskId", requireStaff, staffWalletHandler.UpdateRewardAd)
	api.GET("/staff/wallet/display-ad-settings", requireStaff, staffWalletHandler.GetDisplayAdSettings)
	api.PUT("/staff/wallet/display-ad-settings", requireStaff, staffWalletHandler.SaveDisplayAdSettings)

	// 1.3 Staff secondhand listing routes.
	api.GET("/staff/secondhand/listings", requireStaff, staffListingHandler.ListSecondhand)
	api.POST("/staff/secondhand/listings/:listingId/publish", requireStaff, staffListingHandler.PublishSecondhand)
	api.POST("/staff/secondhand/listings/:listingId/deactivate", requireStaff, staffListingHandler.DeactivateSecondhand)
	api.POST("/staff/secondhand/listings/:listingId/renew", requireStaff, staffListingHandler.RenewSecondhand)

	// 1.4 Staff property sale routes.
	api.GET("/staff/property-sales", requireStaff, staffListingHandler.ListPropertySales)
	api.GET("/staff/property-sales/:listingId", requireStaff, staffListingHandler.GetPropertySale)
	api.PATCH("/staff/property-sales/:listingId", requireStaff, staffListingHandler.UpdatePropertySale)
	api.POST("/staff/property-sales/:listingId/publish", requireStaff, staffListingHandler.PublishPropertySale)
	api.POST("/staff/property-sales/:listingId/deactivate", requireStaff, staffListingHandler.DeactivatePropertySale)
	api.POST("/staff/property-sales/:listingId/renew", requireStaff, staffListingHandler.RenewPropertySale)

	// 1.5 Staff serviced residence routes.
	api.GET("/staff/serviced-apartments", requireStaff, staffListingHandler.ListServicedApartments)
	api.GET("/staff/serviced-apartments/:listingId", requireStaff, staffListingHandler.GetServicedApartment)
	api.PATCH("/staff/serviced-apartments/:listingId", requireStaff, staffListingHandler.UpdateServicedApartment)
	api.POST("/staff/serviced-apartments/:listingId/renew", requireStaff, staffListingHandler.RenewServicedApartment)

	// 1.6 Staff home content setting routes.
	api.GET("/home/settings/carousel", requireStaff, homeContentHandler.SettingsCarousel)
	api.PUT("/home/settings/carousel", requireStaff, homeContentHandler.SaveSettingsCarousel)
	api.GET("/home/settings/module-cards", requireStaff, homeContentHandler.SettingsModuleCards)
	api.PUT("/home/settings/module-cards", requireStaff, homeContentHandler.SaveSettingsModuleCards)
	api.GET("/home/settings/login-hero", requireStaff, homeContentHandler.SettingsLoginHero)
	api.PUT("/home/settings/login-hero", requireStaff, homeContentHandler.SaveSettingsLoginHero)

	// 1.7 Staff secondhand setting routes.
	api.GET("/secondhand/settings/listings", requireStaff, secondhandHandler.SettingsListings)
	api.POST("/secondhand/settings/listings/:listingId/mark-sold", requireStaff, secondhandHandler.SettingsMarkSold)

	// 1.8 Staff building chat management routes.
	api.GET("/staff/building-chats", requireStaff, chatHandler.ListAllBuildingChatsForStaff)
}

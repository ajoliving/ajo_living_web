/*
 * Member route registration.
 * 1. Register member profile, ismart binding, and member overview routes.
 * 2. Keep member center routes separate from payment, wallet, and staff routes.
 */
package router

import (
	"github.com/gin-gonic/gin"

	"ajoliving_web/http_service/internal/handler"
)

// 1. registerMemberRoutes registers member profile and member center routes.
func registerMemberRoutes(
	api *gin.RouterGroup,
	userHandler *handler.UserHandler,
	secondhandHandler *handler.SecondhandHandler,
	propertyHandler *handler.PropertyHandler,
	requireAuth gin.HandlerFunc,
	requireActive gin.HandlerFunc,
) {
	api.GET("/me", requireAuth, userHandler.GetMe)
	api.PATCH("/me/profile", requireActive, userHandler.UpdateProfile)
	api.POST("/me/ismart/bind", requireActive, userHandler.BindIsmart)
	api.GET("/me/secondhand/listings", requireActive, secondhandHandler.MyListings)
	api.GET("/me/secondhand/favorites", requireActive, secondhandHandler.MyFavorites)
	api.GET("/me/property-sales", requireActive, propertyHandler.MyPropertySales)
	api.GET("/me/property-sales/favorites", requireActive, propertyHandler.MyFavoritePropertySales)
	api.GET("/me/serviced-apartments", requireActive, propertyHandler.MyServicedApartments)
}

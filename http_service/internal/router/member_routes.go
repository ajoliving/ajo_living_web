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
	orderHandler *handler.OrderHandler,
	requireAuth gin.HandlerFunc,
) {
	api.GET("/me", requireAuth, userHandler.GetMe)
	api.PATCH("/me/profile", requireAuth, userHandler.UpdateProfile)
	api.POST("/me/ismart/bind", requireAuth, userHandler.BindIsmart)
	api.GET("/me/secondhand/listings", requireAuth, secondhandHandler.MyListings)
	api.GET("/me/secondhand/favorites", requireAuth, secondhandHandler.MyFavorites)
	api.GET("/me/property-sales", requireAuth, propertyHandler.MyPropertySales)
	api.GET("/me/serviced-apartments", requireAuth, propertyHandler.MyServicedApartments)
	api.GET("/me/orders", requireAuth, orderHandler.MyOrders)
}

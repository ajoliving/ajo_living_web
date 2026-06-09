/*
 * Supermarket offer route registration.
 * 1. Register member favorite and price alert routes.
 * 2. Keep public supermarket offer routes in public route registration.
 */
package router

import (
	"github.com/gin-gonic/gin"

	"ajoliving_web/http_service/internal/handler"
)

// 1. registerSupermarketMemberRoutes registers member supermarket offer routes.
func registerSupermarketMemberRoutes(api *gin.RouterGroup, supermarketOfferHandler *handler.SupermarketOfferHandler, requireAuth gin.HandlerFunc) {
	api.GET("/me/supermarket-offers/favorites", requireAuth, supermarketOfferHandler.ListFavorites)
	api.POST("/me/supermarket-offers/favorites", requireAuth, supermarketOfferHandler.AddFavorite)
	api.DELETE("/me/supermarket-offers/favorites/:code", requireAuth, supermarketOfferHandler.RemoveFavorite)
	api.GET("/me/supermarket-offers/price-alerts", requireAuth, supermarketOfferHandler.ListPriceAlerts)
	api.POST("/me/supermarket-offers/price-alerts", requireAuth, supermarketOfferHandler.CreatePriceAlert)
	api.PATCH("/me/supermarket-offers/price-alerts/:id", requireAuth, supermarketOfferHandler.UpdatePriceAlert)
	api.DELETE("/me/supermarket-offers/price-alerts/:id", requireAuth, supermarketOfferHandler.DeletePriceAlert)
}

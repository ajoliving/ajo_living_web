/*
 * Secondhand route registration.
 * 1. Register authenticated secondhand listing write routes.
 * 2. Keep public secondhand listing routes in public route registration.
 */
package router

import (
	"github.com/gin-gonic/gin"

	"ajoliving_web/http_service/internal/handler"
)

// 1. registerSecondhandRoutes registers authenticated secondhand listing routes.
func registerSecondhandRoutes(api *gin.RouterGroup, secondhandHandler *handler.SecondhandHandler, requireAuth gin.HandlerFunc) {
	api.POST("/listings", requireAuth, secondhandHandler.Create)
	api.PATCH("/listings/:listingId", requireAuth, secondhandHandler.Update)
	api.POST("/listings/:listingId/publish", requireAuth, secondhandHandler.Publish)
	api.POST("/listings/:listingId/republish", requireAuth, secondhandHandler.Republish)
	api.POST("/listings/:listingId/renew", requireAuth, secondhandHandler.Renew)
	api.POST("/listings/:listingId/mark-sold", requireAuth, secondhandHandler.MarkSold)
	api.POST("/listings/:listingId/deactivate", requireAuth, secondhandHandler.Deactivate)
	api.POST("/listings/:listingId/favorite", requireAuth, secondhandHandler.Favorite)
	api.DELETE("/listings/:listingId/favorite", requireAuth, secondhandHandler.Unfavorite)
	api.POST("/secondhand/listings", requireAuth, secondhandHandler.Create)
	api.PATCH("/secondhand/listings/:listingId", requireAuth, secondhandHandler.Update)
	api.POST("/secondhand/listings/:listingId/publish", requireAuth, secondhandHandler.Publish)
	api.POST("/secondhand/listings/:listingId/republish", requireAuth, secondhandHandler.Republish)
	api.POST("/secondhand/listings/:listingId/renew", requireAuth, secondhandHandler.Renew)
	api.POST("/secondhand/listings/:listingId/mark-sold", requireAuth, secondhandHandler.MarkSold)
	api.POST("/secondhand/listings/:listingId/deactivate", requireAuth, secondhandHandler.Deactivate)
	api.POST("/secondhand/listings/:listingId/favorite", requireAuth, secondhandHandler.Favorite)
	api.DELETE("/secondhand/listings/:listingId/favorite", requireAuth, secondhandHandler.Unfavorite)
}

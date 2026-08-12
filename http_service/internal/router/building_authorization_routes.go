/*
 * 大廈副戶授權路由。
 * 1. 集中註冊主戶授權管理與受邀帳戶啟用入口。
 * 2. 受邀帳戶啟用不要求先登入，主戶管理要求登入。
 */
package router

import (
	"github.com/gin-gonic/gin"

	"ajoliving_web/http_service/internal/handler"
)

// 1. registerBuildingAuthorizationRoutes registers local building authorization endpoints.
func registerBuildingAuthorizationRoutes(api *gin.RouterGroup, authorizationHandler *handler.BuildingAuthorizationHandler, requireAuth gin.HandlerFunc) {
	api.POST("/auth/building-authorizations/accept", authorizationHandler.Accept)
	api.GET("/me/building-authorizations/permissions", requireAuth, authorizationHandler.AvailablePermissions)
	api.GET("/me/building-authorizations", requireAuth, authorizationHandler.List)
	api.POST("/me/building-authorizations", requireAuth, authorizationHandler.Create)
	api.DELETE("/me/building-authorizations/:authorizationId", requireAuth, authorizationHandler.Revoke)
}

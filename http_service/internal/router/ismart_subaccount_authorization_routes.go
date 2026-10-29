/*
 * iSmart 單位級副戶授權路由註冊。
 * 1. 主戶端點需登入且為有效會員。
 * 2. 確認令牌解析與令牌確認端點為公開端點，供電郵連結直接使用。
 */
package router

import (
	"github.com/gin-gonic/gin"

	"ajoliving_web/http_service/internal/handler"
)

// 1. registerIsmartSubaccountAuthorizationRoutes registers authenticated subaccount authorization routes.
func registerIsmartSubaccountAuthorizationRoutes(
	api *gin.RouterGroup,
	authorizationHandler *handler.IsmartSubaccountAuthorizationHandler,
	requireAuth gin.HandlerFunc,
) {
	api.POST("/me/ismart/subaccount-authorizations", requireAuth, authorizationHandler.CreateAuthorization)
	api.GET("/me/ismart/subaccount-authorizations", requireAuth, authorizationHandler.ListAuthorizations)
	api.POST("/me/ismart/subaccount-authorizations/:authorizationId/resend", requireAuth, authorizationHandler.ResendNotification)
	api.POST("/me/ismart/subaccount-authorizations/:authorizationId/confirm", requireAuth, authorizationHandler.ConfirmAuthorization)
	api.POST("/me/ismart/subaccount-authorizations/:authorizationId/revoke", requireAuth, authorizationHandler.RevokeAuthorization)
}

// 2. registerPublicIsmartSubaccountAuthorizationRoutes registers public confirmation routes.
func registerPublicIsmartSubaccountAuthorizationRoutes(
	api *gin.RouterGroup,
	authorizationHandler *handler.IsmartSubaccountAuthorizationHandler,
) {
	api.GET("/auth/ismart-subaccount-authorizations/:token", authorizationHandler.ResolveAuthorizationToken)
	api.POST("/auth/ismart-subaccount-authorizations/confirm", authorizationHandler.ConfirmAuthorizationByToken)
}

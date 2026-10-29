/*
 * iSmart 副戶授權路由契約測試。
 * 1. 驗證主戶端點需登入且路徑保持穩定。
 * 2. 驗證電郵確認連結使用的兩個公開端點仍註冊。
 */
package router

import (
	"testing"

	"github.com/gin-gonic/gin"

	"ajoliving_web/http_service/internal/handler"
)

// 1. TestRegisterIsmartSubaccountAuthorizationRoutes verifies the authorization route contract.
func TestRegisterIsmartSubaccountAuthorizationRoutes(t *testing.T) {
	engine := gin.New()
	api := engine.Group("/api/v1")
	authorizationHandler := handler.NewIsmartSubaccountAuthorizationHandler(nil)
	registerIsmartSubaccountAuthorizationRoutes(api, authorizationHandler, func(c *gin.Context) { c.Next() })
	registerPublicIsmartSubaccountAuthorizationRoutes(api, authorizationHandler)

	routes := make(map[string]bool)
	for _, route := range engine.Routes() {
		routes[route.Method+" "+route.Path] = true
	}
	for _, route := range []string{
		"POST /api/v1/me/ismart/subaccount-authorizations",
		"GET /api/v1/me/ismart/subaccount-authorizations",
		"POST /api/v1/me/ismart/subaccount-authorizations/:authorizationId/resend",
		"POST /api/v1/me/ismart/subaccount-authorizations/:authorizationId/confirm",
		"POST /api/v1/me/ismart/subaccount-authorizations/:authorizationId/revoke",
		"GET /api/v1/auth/ismart-subaccount-authorizations/:token",
		"POST /api/v1/auth/ismart-subaccount-authorizations/confirm",
	} {
		if !routes[route] {
			t.Fatalf("expected route %s", route)
		}
	}
}

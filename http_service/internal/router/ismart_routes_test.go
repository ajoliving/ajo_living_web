/*
 * iSmart proxy route contract tests.
 * 1. Verify the member binding query routes stay registered.
 * 2. Keep the owner binding submission route available for compatibility.
 */
package router

import (
	"testing"

	"github.com/gin-gonic/gin"

	"ajoliving_web/http_service/internal/handler"
)

// 1. TestRegisterIsmartRoutesIncludesBindingQueries verifies the binding route contract.
func TestRegisterIsmartRoutesIncludesBindingQueries(t *testing.T) {
	engine := gin.New()
	registerIsmartRoutes(
		engine.Group("/api/v1"),
		handler.NewIsmartExternalHandler(nil),
		func(c *gin.Context) { c.Next() },
	)

	routes := make(map[string]bool)
	for _, route := range engine.Routes() {
		routes[route.Method+" "+route.Path] = true
	}
	for _, route := range []string{
		"POST /api/v1/me/ismart/owner-binding-requests",
		"GET /api/v1/me/ismart/owner-binding-requests",
		"GET /api/v1/me/ismart/flat-bindings",
		"POST /api/v1/me/ismart/subaccounts/check-contacts",
		"POST /api/v1/me/ismart/subaccounts/register-and-grant",
	} {
		if !routes[route] {
			t.Fatalf("expected route %s", route)
		}
	}
}

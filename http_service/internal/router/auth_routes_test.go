/*
 * Authentication route contract tests.
 * 1. Verify the unified login endpoint remains registered.
 * 2. Keep legacy authentication routes available for compatibility.
 */
package router

import (
	"testing"

	"github.com/gin-gonic/gin"

	"ajoliving_web/http_service/internal/handler"
	"ajoliving_web/http_service/internal/middleware"
)

// 1. TestRegisterAuthRoutesIncludesUnifiedLogin verifies the public route contract.
func TestRegisterAuthRoutesIncludesUnifiedLogin(t *testing.T) {
	engine := gin.New()
	registerAuthRoutes(
		engine.Group("/api/v1"),
		handler.NewAuthHandler(nil),
		func(c *gin.Context) { c.Next() },
		middleware.NewInMemoryLimiter(),
	)

	routes := make(map[string]bool)
	for _, route := range engine.Routes() {
		routes[route.Method+" "+route.Path] = true
	}
	for _, route := range []string{
		"POST /api/v1/auth/registration/availability",
		"POST /api/v1/auth/login",
		"POST /api/v1/auth/email/login",
		"POST /api/v1/auth/username/login",
		"POST /api/v1/auth/phone/login",
		"POST /api/v1/auth/ismart/login",
		"POST /api/v1/auth/refresh",
	} {
		if !routes[route] {
			t.Fatalf("expected route %s", route)
		}
	}
}

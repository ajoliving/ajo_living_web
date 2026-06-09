/*
 * Authentication route registration.
 * 1. Register OTP, email, phone, ismart login, and logout routes.
 * 2. Keep authentication limiter rules colocated with authentication routes.
 */
package router

import (
	"time"

	"github.com/gin-gonic/gin"

	"ajoliving_web/http_service/internal/handler"
	"ajoliving_web/http_service/internal/middleware"
)

// 1. registerAuthRoutes registers authentication endpoints.
func registerAuthRoutes(
	api *gin.RouterGroup,
	authHandler *handler.AuthHandler,
	requireAuth gin.HandlerFunc,
	limiter *middleware.InMemoryLimiter,
) {
	api.POST("/auth/otp/request", limiter.Limit(5, 10*time.Minute, func(c *gin.Context) string {
		return "otp_request:" + c.ClientIP()
	}), authHandler.RequestOTP)
	api.POST("/auth/otp/verify", limiter.Limit(10, 10*time.Minute, func(c *gin.Context) string {
		return "otp_verify:" + c.ClientIP()
	}), authHandler.VerifyOTP)
	api.POST("/auth/email/otp/request", limiter.Limit(5, 10*time.Minute, func(c *gin.Context) string {
		return "email_otp_request:" + c.ClientIP()
	}), authHandler.RequestEmailOTP)
	api.POST("/auth/email/otp/verify", limiter.Limit(10, 10*time.Minute, func(c *gin.Context) string {
		return "email_otp_verify:" + c.ClientIP()
	}), authHandler.VerifyEmailOTP)
	api.POST("/auth/email/register", limiter.Limit(10, 10*time.Minute, func(c *gin.Context) string {
		return "email_register:" + c.ClientIP()
	}), authHandler.RegisterEmail)
	api.POST("/auth/email/login", limiter.Limit(20, 10*time.Minute, func(c *gin.Context) string {
		return "email_login:" + c.ClientIP()
	}), authHandler.LoginEmail)
	api.POST("/auth/phone/login", limiter.Limit(20, 10*time.Minute, func(c *gin.Context) string {
		return "phone_login:" + c.ClientIP()
	}), authHandler.LoginPhone)
	api.POST("/auth/ismart/login", limiter.Limit(20, 10*time.Minute, func(c *gin.Context) string {
		return "ismart_login:" + c.ClientIP()
	}), authHandler.LoginIsmart)
	api.POST("/auth/logout", requireAuth, authHandler.Logout)
}

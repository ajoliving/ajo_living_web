/*
 * HTTP router registration.
 * 1. Compose handlers, middleware, and module-based route groups.
 * 2. Keep routing concerns isolated from business services.
 * 3. Keep module boundaries in code aligned with the API document.
 */
package router

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"

	"ajoliving_web/http_service/internal/config"
	"ajoliving_web/http_service/internal/handler"
	"ajoliving_web/http_service/internal/middleware"
	"ajoliving_web/http_service/internal/service"
)

// 1. Dependencies groups objects required to build the router.
type Dependencies struct {
	Config              *config.Config
	Logger              *slog.Logger
	AuthService         *service.AuthService
	UserService         *service.UserService
	StaffService        *service.StaffService
	UploadService       *service.UploadService
	HomeContentService  *service.HomeContentService
	WalletService       *service.WalletService
	SecondhandService   *service.SecondhandService
	PropertyService     *service.PropertyService
	ChatService         *service.ChatService
	OrderService        *service.OrderService
	NotificationService *service.NotificationService
}

// 2. New builds and returns the gin engine.
func New(deps *Dependencies) *gin.Engine {
	engine := gin.New()
	limiter := middleware.NewInMemoryLimiter()
	requireAuth := middleware.RequireAuth(deps.AuthService)
	requireStaff := middleware.RequireStaff(deps.AuthService)
	optionalAuth := middleware.OptionalAuth(deps.AuthService)

	engine.Use(middleware.CORS(deps.Config))
	engine.Use(middleware.RequestID())
	engine.Use(middleware.Recovery(deps.Logger))
	engine.Use(middleware.AccessLog(deps.Logger))

	healthHandler := handler.NewHealthHandler()
	authHandler := handler.NewAuthHandler(deps.AuthService)
	userHandler := handler.NewUserHandler(deps.UserService)
	staffHandler := handler.NewStaffHandler(deps.StaffService)
	uploadHandler := handler.NewUploadHandler(deps.UploadService)
	homeContentHandler := handler.NewHomeContentHandler(deps.HomeContentService)
	walletHandler := handler.NewWalletHandler(deps.WalletService)
	staffWalletHandler := handler.NewStaffWalletHandler(deps.WalletService)
	secondhandHandler := handler.NewSecondhandHandler(deps.SecondhandService)
	propertyHandler := handler.NewPropertyHandler(deps.PropertyService)
	staffListingHandler := handler.NewStaffListingHandler(deps.SecondhandService, deps.PropertyService)
	chatHandler := handler.NewChatHandler(deps.ChatService)
	orderHandler := handler.NewOrderHandler(deps.OrderService)
	notificationHandler := handler.NewNotificationHandler(deps.NotificationService)

	api := engine.Group("/api/v1")
	registerPublicRoutes(api, healthHandler, userHandler, homeContentHandler, secondhandHandler, propertyHandler, optionalAuth)
	registerAppRoutes(
		api,
		authHandler,
		userHandler,
		uploadHandler,
		homeContentHandler,
		walletHandler,
		secondhandHandler,
		propertyHandler,
		chatHandler,
		orderHandler,
		notificationHandler,
		requireAuth,
		limiter,
	)
	registerStaffRoutes(api, staffHandler, staffWalletHandler, staffListingHandler, homeContentHandler, secondhandHandler, requireStaff)
	registerStaffNoticeRoutes(api, chatHandler, requireStaff)

	return engine
}

// 3. registerPublicRoutes registers health checks and public-facing routes.
func registerPublicRoutes(
	api *gin.RouterGroup,
	healthHandler *handler.HealthHandler,
	userHandler *handler.UserHandler,
	homeContentHandler *handler.HomeContentHandler,
	secondhandHandler *handler.SecondhandHandler,
	propertyHandler *handler.PropertyHandler,
	optionalAuth gin.HandlerFunc,
) {
	api.GET("/health", healthHandler.Check)
	api.GET("/channel-home/overview", optionalAuth, userHandler.ChannelHomeOverview)
	api.GET("/home/content", homeContentHandler.PublicContent)
	api.GET("/home/login-hero", homeContentHandler.LoginHero)
	api.GET("/listings", optionalAuth, secondhandHandler.ListPublic)
	api.GET("/listings/:listingId", optionalAuth, secondhandHandler.GetDetail)
	api.GET("/secondhand/discover", optionalAuth, secondhandHandler.PublicDiscover)
	api.GET("/secondhand/listings", optionalAuth, secondhandHandler.ListPublic)
	api.GET("/secondhand/listings/:listingId", optionalAuth, secondhandHandler.GetDetail)
	api.GET("/property-sales", optionalAuth, propertyHandler.ListPropertySales)
	api.GET("/property-sales/:listingId", optionalAuth, propertyHandler.GetPropertySale)
	api.GET("/serviced-apartments", optionalAuth, propertyHandler.ListServicedApartments)
	api.GET("/serviced-apartments/:listingId", optionalAuth, propertyHandler.GetServicedApartment)
}

// 4. registerAppRoutes registers non-public application routes.
func registerAppRoutes(
	api *gin.RouterGroup,
	authHandler *handler.AuthHandler,
	userHandler *handler.UserHandler,
	uploadHandler *handler.UploadHandler,
	homeContentHandler *handler.HomeContentHandler,
	walletHandler *handler.WalletHandler,
	secondhandHandler *handler.SecondhandHandler,
	propertyHandler *handler.PropertyHandler,
	chatHandler *handler.ChatHandler,
	orderHandler *handler.OrderHandler,
	notificationHandler *handler.NotificationHandler,
	requireAuth gin.HandlerFunc,
	limiter *middleware.InMemoryLimiter,
) {
	api.GET("/meta/communities", requireAuth, userHandler.ListCommunities)
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
	api.POST("/auth/logout", requireAuth, authHandler.Logout)
	api.GET("/me", requireAuth, userHandler.GetMe)
	api.PATCH("/me/profile", requireAuth, userHandler.UpdateProfile)
	api.GET("/me/wallet", requireAuth, walletHandler.GetWallet)
	api.GET("/me/wallet/transactions", requireAuth, walletHandler.ListTransactions)
	api.GET("/me/wallet/ad-tasks", requireAuth, walletHandler.ListAdTasks)
	api.POST("/me/wallet/ad-tasks/:taskId/start", requireAuth, walletHandler.StartAdTask)
	api.POST("/me/wallet/ad-tasks/:taskId/claim", requireAuth, walletHandler.ClaimAdTask)
	api.POST("/me/wallet/ad-tasks/:taskId/click", requireAuth, walletHandler.TrackAdTaskClick)
	api.GET("/me/secondhand/listings", requireAuth, secondhandHandler.MyListings)
	api.GET("/me/secondhand/favorites", requireAuth, secondhandHandler.MyFavorites)
	api.GET("/me/property-sales", requireAuth, propertyHandler.MyPropertySales)
	api.GET("/me/serviced-apartments", requireAuth, propertyHandler.MyServicedApartments)
	api.GET("/me/orders", requireAuth, orderHandler.MyOrders)
	api.POST("/oss/callback", uploadHandler.UploadCallback)
	api.POST("/oss/presign", requireAuth, uploadHandler.Presign)
	api.POST("/oss/complete", requireAuth, uploadHandler.CompleteUpload)
	api.GET("/oss/assets", requireAuth, uploadHandler.ListAssets)
	api.GET("/oss/assets/:mediaAssetId", requireAuth, uploadHandler.GetAsset)
	api.DELETE("/oss/assets/:mediaAssetId", requireAuth, uploadHandler.DeleteAsset)
	api.POST("/uploads/presign", requireAuth, uploadHandler.Presign)
	api.POST("/uploads/complete", requireAuth, uploadHandler.CompleteUpload)

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
	api.POST("/property-sales", requireAuth, propertyHandler.CreatePropertySale)
	api.PATCH("/property-sales/:listingId", requireAuth, propertyHandler.UpdatePropertySale)
	api.POST("/property-sales/:listingId/publish", requireAuth, propertyHandler.PublishPropertySale)
	api.POST("/property-sales/:listingId/republish", requireAuth, propertyHandler.RepublishPropertySale)
	api.POST("/property-sales/:listingId/mark-sold", requireAuth, propertyHandler.MarkPropertySaleSold)
	api.POST("/property-sales/:listingId/deactivate", requireAuth, propertyHandler.DeactivatePropertySale)
	api.POST("/serviced-apartments", requireAuth, propertyHandler.CreateServicedApartment)
	api.PATCH("/serviced-apartments/:listingId", requireAuth, propertyHandler.UpdateServicedApartment)
	api.POST("/serviced-apartments/:listingId/publish", requireAuth, propertyHandler.PublishServicedApartment)
	api.POST("/serviced-apartments/:listingId/republish", requireAuth, propertyHandler.RepublishServicedApartment)
	api.POST("/serviced-apartments/:listingId/deactivate", requireAuth, propertyHandler.DeactivateServicedApartment)

	api.POST("/listings/:listingId/contact-access",
		requireAuth,
		limiter.Limit(20, time.Minute, func(c *gin.Context) string {
			return "contact_access:" + c.ClientIP()
		}),
		secondhandHandler.ContactAccess,
	)
	api.POST("/secondhand/listings/:listingId/contact-access",
		requireAuth,
		limiter.Limit(20, time.Minute, func(c *gin.Context) string {
			return "contact_access:" + c.ClientIP()
		}),
		secondhandHandler.ContactAccess,
	)
	api.POST("/property-sales/:listingId/contact-access",
		requireAuth,
		limiter.Limit(20, time.Minute, func(c *gin.Context) string {
			return "property_sale_contact_access:" + c.ClientIP()
		}),
		propertyHandler.ContactAccessPropertySale,
	)
	api.POST("/serviced-apartments/:listingId/contact-access",
		requireAuth,
		limiter.Limit(20, time.Minute, func(c *gin.Context) string {
			return "serviced_apartment_contact_access:" + c.ClientIP()
		}),
		propertyHandler.ContactAccessServicedApartment,
	)
	api.POST("/listings/:listingId/chats", requireAuth, chatHandler.CreateOrReuse)
	api.GET("/chats", requireAuth, chatHandler.ListChats)
	api.GET("/chats/:chatId", requireAuth, chatHandler.GetChat)
	api.GET("/chats/:chatId/messages", requireAuth, chatHandler.ListMessages)
	api.POST("/chats/:chatId/messages",
		requireAuth,
		limiter.Limit(30, time.Minute, func(c *gin.Context) string {
			return "send_message:" + c.ClientIP()
		}),
		chatHandler.SendMessage,
	)
	api.POST("/chats/:chatId/read", requireAuth, chatHandler.MarkRead)

	api.POST("/listings/:listingId/orders", requireAuth, orderHandler.Create)
	api.GET("/orders/:orderId", requireAuth, orderHandler.GetDetail)
	api.POST("/orders/:orderId/confirm", requireAuth, orderHandler.Confirm)
	api.POST("/orders/:orderId/cancel", requireAuth, orderHandler.Cancel)
	api.POST("/orders/:orderId/complete", requireAuth, orderHandler.Complete)

	api.GET("/notifications", requireAuth, notificationHandler.List)
	api.GET("/notifications/unread-count", requireAuth, notificationHandler.UnreadCount)
	api.POST("/notifications/:notificationId/read", requireAuth, notificationHandler.MarkRead)
	api.POST("/notifications/read-all", requireAuth, notificationHandler.MarkAllRead)
}

// 5. registerStaffRoutes registers staff-only management endpoints.
func registerStaffRoutes(
	api *gin.RouterGroup,
	staffHandler *handler.StaffHandler,
	staffWalletHandler *handler.StaffWalletHandler,
	staffListingHandler *handler.StaffListingHandler,
	homeContentHandler *handler.HomeContentHandler,
	secondhandHandler *handler.SecondhandHandler,
	requireStaff gin.HandlerFunc,
) {
	api.GET("/staff/me", requireStaff, staffHandler.GetMe)
	api.GET("/staff/roles", requireStaff, staffHandler.ListRoles)
	api.GET("/staff/users", requireStaff, staffHandler.ListUsers)
	api.POST("/staff/users", requireStaff, staffHandler.CreateUser)
	api.PATCH("/staff/users/:userId/role", requireStaff, staffHandler.UpdateUserRole)
	api.GET("/staff/wallet/transactions", requireStaff, staffWalletHandler.ListTransactions)
	api.POST("/staff/wallet/grants", requireStaff, staffWalletHandler.GrantPoints)
	api.GET("/staff/wallet/reward-ads", requireStaff, staffWalletHandler.ListRewardAds)
	api.GET("/staff/wallet/reward-ads/:taskId", requireStaff, staffWalletHandler.GetRewardAd)
	api.POST("/staff/wallet/reward-ads", requireStaff, staffWalletHandler.CreateRewardAd)
	api.PATCH("/staff/wallet/reward-ads/:taskId", requireStaff, staffWalletHandler.UpdateRewardAd)
	api.GET("/staff/secondhand/listings", requireStaff, staffListingHandler.ListSecondhand)
	api.POST("/staff/secondhand/listings/:listingId/publish", requireStaff, staffListingHandler.PublishSecondhand)
	api.POST("/staff/secondhand/listings/:listingId/deactivate", requireStaff, staffListingHandler.DeactivateSecondhand)
	api.POST("/staff/secondhand/listings/:listingId/renew", requireStaff, staffListingHandler.RenewSecondhand)
	api.GET("/staff/property-sales", requireStaff, staffListingHandler.ListPropertySales)
	api.POST("/staff/property-sales/:listingId/publish", requireStaff, staffListingHandler.PublishPropertySale)
	api.POST("/staff/property-sales/:listingId/deactivate", requireStaff, staffListingHandler.DeactivatePropertySale)
	api.POST("/staff/property-sales/:listingId/renew", requireStaff, staffListingHandler.RenewPropertySale)
	api.GET("/staff/serviced-apartments", requireStaff, staffListingHandler.ListServicedApartments)
	api.POST("/staff/serviced-apartments/:listingId/renew", requireStaff, staffListingHandler.RenewServicedApartment)
	api.GET("/home/settings/carousel", requireStaff, homeContentHandler.SettingsCarousel)
	api.PUT("/home/settings/carousel", requireStaff, homeContentHandler.SaveSettingsCarousel)
	api.GET("/home/settings/module-cards", requireStaff, homeContentHandler.SettingsModuleCards)
	api.PUT("/home/settings/module-cards", requireStaff, homeContentHandler.SaveSettingsModuleCards)
	api.GET("/home/settings/login-hero", requireStaff, homeContentHandler.SettingsLoginHero)
	api.PUT("/home/settings/login-hero", requireStaff, homeContentHandler.SaveSettingsLoginHero)
	api.GET("/secondhand/settings/listings", requireStaff, secondhandHandler.SettingsListings)
	api.GET("/secondhand/settings/discover-placements", requireStaff, secondhandHandler.SettingsDiscoverPlacements)
	api.PUT("/secondhand/settings/discover-placements", requireStaff, secondhandHandler.SaveSettingsDiscoverPlacements)
	api.POST("/secondhand/settings/listings/:listingId/mark-sold", requireStaff, secondhandHandler.SettingsMarkSold)
	api.POST("/secondhand/settings/listings/:listingId/deactivate", requireStaff, secondhandHandler.SettingsDeactivate)
}

// 6. registerStaffNoticeRoutes registers staff-only system notice publishing.
func registerStaffNoticeRoutes(api *gin.RouterGroup, chatHandler *handler.ChatHandler, requireStaff gin.HandlerFunc) {
	api.POST("/staff/system-notices", requireStaff, chatHandler.PublishSystemNotice)
}

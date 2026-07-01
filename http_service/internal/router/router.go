/*
 * HTTP router registration.
 * 1. Compose handlers, middleware, and module-based route groups.
 * 2. Keep routing concerns isolated from business services.
 * 3. Keep module boundaries in code aligned with the API document.
 */
package router

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	"ajoliving_web/http_service/internal/config"
	"ajoliving_web/http_service/internal/handler"
	"ajoliving_web/http_service/internal/middleware"
	"ajoliving_web/http_service/internal/service"
)

// 1. Dependencies groups objects required to build the router.
type Dependencies struct {
	Config                  *config.Config
	Logger                  *slog.Logger
	AuthService             *service.AuthService
	UserService             *service.UserService
	StaffService            *service.StaffService
	UploadService           *service.UploadService
	HomeContentService      *service.HomeContentService
	POSBuildingService      *service.POSBuildingService
	POSPaymentService       *service.POSPaymentService
	IsmartExternalService   *service.IsmartExternalService
	SecurityICCTVService    *service.SecurityICCTVService
	WalletService           *service.WalletService
	SecondhandService       *service.SecondhandService
	PropertyService         *service.PropertyService
	ChatService             *service.ChatService
	OrderService            *service.OrderService
	NotificationService     *service.NotificationService
	SupermarketOfferService *service.SupermarketOfferService
	MarketTrendService      *service.MarketTrendService
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
	posBuildingHandler := handler.NewPOSBuildingHandler(deps.POSBuildingService, deps.POSPaymentService)
	posPaymentHandler := handler.NewPOSPaymentHandler(deps.POSPaymentService)
	ismartHandler := handler.NewIsmartExternalHandler(deps.IsmartExternalService)
	securityICCTVHandler := handler.NewSecurityICCTVHandler(deps.SecurityICCTVService)
	walletHandler := handler.NewWalletHandler(deps.WalletService)
	staffWalletHandler := handler.NewStaffWalletHandler(deps.WalletService)
	secondhandHandler := handler.NewSecondhandHandler(deps.SecondhandService)
	propertyHandler := handler.NewPropertyHandler(deps.PropertyService)
	staffListingHandler := handler.NewStaffListingHandler(deps.SecondhandService, deps.PropertyService)
	chatHandler := handler.NewChatHandler(deps.ChatService)
	orderHandler := handler.NewOrderHandler(deps.OrderService)
	notificationHandler := handler.NewNotificationHandler(deps.NotificationService)
	supermarketOfferHandler := handler.NewSupermarketOfferHandler(deps.SupermarketOfferService)
	marketTrendHandler := handler.NewMarketTrendHandler(deps.MarketTrendService)

	api := engine.Group("/api/v1")
	registerPublicRoutes(api, healthHandler, userHandler, homeContentHandler, walletHandler, secondhandHandler, propertyHandler, supermarketOfferHandler, marketTrendHandler, optionalAuth)
	registerPublicPOSPaymentRoutes(api, posBuildingHandler, walletHandler)
	registerAuthRoutes(api, authHandler, requireAuth, limiter)
	registerMemberRoutes(api, userHandler, secondhandHandler, propertyHandler, orderHandler, requireAuth)
	registerPOSPaymentRoutes(api, posBuildingHandler, posPaymentHandler, requireAuth)
	registerIsmartRoutes(api, ismartHandler, requireAuth)
	registerSecurityRoutes(api, securityICCTVHandler, requireAuth)
	registerWalletRoutes(api, walletHandler, requireAuth)
	registerUploadRoutes(api, uploadHandler, requireAuth)
	registerSecondhandRoutes(api, secondhandHandler, requireAuth)
	registerPropertyRoutes(api, propertyHandler, requireAuth)
	registerContactRoutes(api, secondhandHandler, propertyHandler, chatHandler, requireAuth, limiter)
	registerChatRoutes(api, chatHandler, requireAuth, limiter)
	registerOrderRoutes(api, orderHandler, requireAuth)
	registerNotificationRoutes(api, notificationHandler, requireAuth)
	registerSupermarketMemberRoutes(api, supermarketOfferHandler, requireAuth)
	registerStaffRoutes(api, staffHandler, staffWalletHandler, staffListingHandler, homeContentHandler, secondhandHandler, requireStaff)
	registerStaffNoticeRoutes(api, notificationHandler, requireStaff)

	return engine
}

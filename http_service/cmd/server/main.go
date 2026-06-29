/*
 * HTTP service entrypoint.
 * 1. Load configuration, logger, database, and providers.
 * 2. Migrate schema, seed data, register routes, and start background tasks.
 * 3. Run the HTTP server with graceful shutdown.
 */
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ajoliving_web/http_service/internal/config"
	"ajoliving_web/http_service/internal/database"
	"ajoliving_web/http_service/internal/logger"
	"ajoliving_web/http_service/internal/router"
	"ajoliving_web/http_service/internal/service"
)

// 1. main wires the application dependencies and starts the HTTP server.
func main() {
	cfg := config.Load()
	logg := logger.New(cfg)

	db, err := database.Open(cfg)
	if err != nil {
		log.Fatal(err)
	}

	if err := database.Migrate(db); err != nil {
		log.Fatal(err)
	}

	if err := database.SeedDefaultAdminAccount(context.Background(), db); err != nil {
		log.Fatal(err)
	}

	if err := database.SeedSystemNotificationAccount(context.Background(), db); err != nil {
		log.Fatal(err)
	}

	if cfg.SeedCommunities {
		if err := database.SeedCommunities(context.Background(), db); err != nil {
			log.Fatal(err)
		}
	}

	storageProvider, err := service.NewStorageProvider(cfg)
	if err != nil {
		log.Fatal(err)
	}

	runtime := &service.Runtime{
		Config:          cfg,
		DB:              db,
		Logger:          logg,
		OTPProvider:     service.NewOTPProvider(cfg),
		MailSender:      service.NewMailSender(cfg),
		StorageProvider: storageProvider,
		OTPStore:        service.NewOTPStore(),
		Now:             time.Now,
	}

	authService := service.NewAuthService(runtime)
	walletService := service.NewWalletService(runtime)
	runtime.WalletService = walletService
	userService := service.NewUserService(runtime)
	staffService := service.NewStaffService(runtime)
	uploadService := service.NewUploadService(runtime)
	homeContentService := service.NewHomeContentService(runtime)
	posBuildingService := service.NewPOSBuildingService(runtime)
	posPaymentService := service.NewPOSPaymentService(runtime)
	secondhandService := service.NewSecondhandService(runtime)
	propertyService := service.NewPropertyService(runtime)
	notificationService := service.NewNotificationService(runtime)
	supermarketOfferService := service.NewSupermarketOfferService(runtime)
	chatService := service.NewChatService(runtime, secondhandService, propertyService)
	orderService := service.NewOrderService(runtime, secondhandService, notificationService)
	lifecycleService := service.NewLifecycleService(runtime)

	if cfg.SeedHomeContent {
		if err := homeContentService.SeedDefaultHomeContent(context.Background(), service.HomeContentSeedOptions{
			WebPublicDir:   cfg.WebPublicDir,
			OperatorUserID: cfg.SystemUserID,
		}); err != nil {
			log.Fatal(err)
		}
	}

	if cfg.EnableExpireTicker {
		go startExpireTicker(lifecycleService, cfg.ExpireTickerInterval)
	}
	if cfg.GoodPriceAlertEnabled {
		go startSupermarketAlertTicker(supermarketOfferService, cfg.GoodPriceAlertInterval)
	}

	engine := router.New(&router.Dependencies{
		Config:                  cfg,
		Logger:                  logg,
		AuthService:             authService,
		UserService:             userService,
		StaffService:            staffService,
		UploadService:           uploadService,
		HomeContentService:      homeContentService,
		POSBuildingService:      posBuildingService,
		POSPaymentService:       posPaymentService,
		WalletService:           walletService,
		SecondhandService:       secondhandService,
		PropertyService:         propertyService,
		ChatService:             chatService,
		OrderService:            orderService,
		NotificationService:     notificationService,
		SupermarketOfferService: supermarketOfferService,
	})

	server := &http.Server{
		Addr:              ":" + cfg.AppPort,
		Handler:           engine,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	waitForShutdown(server)
}

// 2. startExpireTicker periodically expires overdue listings.
func startExpireTicker(lifecycleService *service.LifecycleService, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		_, _ = lifecycleService.ExpireListings(context.Background())
	}
}

// 3. startSupermarketAlertTicker periodically evaluates supermarket price alerts.
func startSupermarketAlertTicker(supermarketOfferService *service.SupermarketOfferService, interval time.Duration) {
	if interval <= 0 {
		interval = time.Hour
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		_, _ = supermarketOfferService.EvaluatePriceAlerts(context.Background())
	}
}

// 4. waitForShutdown gracefully stops the HTTP server.
func waitForShutdown(server *http.Server) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = server.Shutdown(ctx)
}

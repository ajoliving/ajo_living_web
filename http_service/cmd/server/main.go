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
	appCtx, cancelApp := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancelApp()
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatal(err)
	}
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
	mediaProcessor, err := service.NewMediaProcessor(cfg)
	if err != nil {
		log.Fatal(err)
	}
	mediaScanner, err := service.NewMediaScanner(cfg)
	if err != nil {
		log.Fatal(err)
	}
	otpProvider, err := service.NewOTPProvider(cfg)
	if err != nil {
		log.Fatal(err)
	}
	cacheStore, err := service.NewRedisCacheStore(context.Background(), cfg)
	if err != nil {
		logg.Warn("redis cache unavailable; requests will use upstream services", "error", err)
	}
	if cacheStore != nil {
		defer cacheStore.Close()
	}
	realtimeBroker, err := service.NewRedisRealtimeBroker(context.Background(), cfg)
	if err != nil {
		logg.Warn("redis realtime broker unavailable at startup; realtime workers will retry", "error", err)
	}
	if realtimeBroker != nil {
		defer realtimeBroker.Close()
	}

	runtime := &service.Runtime{
		Config:          cfg,
		DB:              db,
		Logger:          logg,
		OTPProvider:     otpProvider,
		MailSender:      service.NewMailSender(cfg),
		StorageProvider: storageProvider,
		MediaProcessor:  mediaProcessor,
		MediaScanner:    mediaScanner,
		CacheStore:      cacheStore,
		OTPStore:        service.NewOTPStore(),
		Now:             time.Now,
	}
	if realtimeBroker != nil {
		go service.NewRealtimeOutboxWorker(runtime, realtimeBroker).Run(appCtx)
	}

	authService := service.NewAuthService(runtime)
	walletService := service.NewWalletService(runtime)
	runtime.WalletService = walletService
	userService := service.NewUserService(runtime)
	staffService := service.NewStaffService(runtime)
	uploadService := service.NewUploadService(runtime)
	homeContentService := service.NewHomeContentService(runtime)
	posBuildingService := service.NewPOSBuildingService(runtime)
	posPaymentService := service.NewPOSPaymentService(runtime, posBuildingService)
	ismartExternalService := service.NewIsmartExternalService(runtime)
	buildingAuthorizationService := service.NewBuildingAuthorizationService(runtime)
	securityICCTVService := service.NewSecurityICCTVService(runtime)
	secondhandService := service.NewSecondhandService(runtime)
	propertyService := service.NewPropertyService(runtime)
	agencyCompanyService := service.NewAgencyCompanyService(runtime)
	notificationService := service.NewNotificationService(runtime)
	supermarketOfferService := service.NewSupermarketOfferService(runtime)
	marketTrendService := service.NewMarketTrendService(runtime)
	chatService := service.NewChatService(runtime, secondhandService, propertyService)
	go service.NewMediaProcessingWorker(runtime).Run(appCtx)
	go startChatMediaCleanupTicker(appCtx, chatService)
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
		Config:                       cfg,
		Runtime:                      runtime,
		Logger:                       logg,
		AuthService:                  authService,
		UserService:                  userService,
		StaffService:                 staffService,
		UploadService:                uploadService,
		HomeContentService:           homeContentService,
		POSBuildingService:           posBuildingService,
		POSPaymentService:            posPaymentService,
		IsmartExternalService:        ismartExternalService,
		BuildingAuthorizationService: buildingAuthorizationService,
		SecurityICCTVService:         securityICCTVService,
		WalletService:                walletService,
		SecondhandService:            secondhandService,
		PropertyService:              propertyService,
		AgencyCompanyService:         agencyCompanyService,
		ChatService:                  chatService,
		NotificationService:          notificationService,
		SupermarketOfferService:      supermarketOfferService,
		MarketTrendService:           marketTrendService,
		RealtimeBroker:               realtimeBroker,
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

	waitForShutdown(appCtx, server)
}

// 2.1 startChatMediaCleanupTicker removes stale unbound chat uploads.
func startChatMediaCleanupTicker(ctx context.Context, chatService *service.ChatService) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_, _ = chatService.CleanupOrphanChatMedia(ctx, 24*time.Hour, 500)
		}
	}
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

// 4. waitForShutdown gracefully stops the HTTP server after context cancellation.
func waitForShutdown(ctx context.Context, server *http.Server) {
	<-ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = server.Shutdown(ctx)
}

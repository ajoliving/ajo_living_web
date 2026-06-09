/*
 * 預設首頁內容種子命令。
 * 1. 連接資料庫並初始化 OSS provider。
 * 2. 將本地預設首頁與登入圖片上傳至 OSS。
 * 3. 幂等寫入首頁輪播、首頁三大圖與登入背景圖設定。
 */
package main

import (
	"context"
	"log"
	"time"

	"ajoliving_web/http_service/internal/config"
	"ajoliving_web/http_service/internal/database"
	"ajoliving_web/http_service/internal/logger"
	"ajoliving_web/http_service/internal/service"
)

// 1. main runs the default content seed once.
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

	homeContentService := service.NewHomeContentService(runtime)
	if err := homeContentService.SeedDefaultHomeContent(context.Background(), service.HomeContentSeedOptions{
		WebPublicDir:   cfg.WebPublicDir,
		OperatorUserID: cfg.SystemUserID,
	}); err != nil {
		log.Fatal(err)
	}
}

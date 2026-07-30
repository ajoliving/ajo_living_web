/*
 * 樓盤草稿扣費更正命令。
 * 1. 預覽指定草稿的歷史重複扣費退款。
 * 2. 以 apply 參數建立可追溯的更正流水。
 * 3. 僅處理仍處於草稿狀態的放售樓盤。
 */
package main

import (
	"context"
	"flag"
	"log"
	"strings"
	"time"

	"ajoliving_web/http_service/internal/config"
	"ajoliving_web/http_service/internal/database"
	"ajoliving_web/http_service/internal/service"
)

// 1. main previews or applies one safe legacy draft-charge correction.
func main() {
	listingID := flag.String("listing-id", "", "property sale draft public ID")
	apply := flag.Bool("apply", false, "create the correction transaction")
	flag.Parse()
	if strings.TrimSpace(*listingID) == "" {
		log.Fatal("listing-id is required")
	}

	cfg := config.Load()
	db, err := database.Open(cfg)
	if err != nil {
		log.Fatal(err)
	}
	runtime := &service.Runtime{Config: cfg, DB: db, Now: time.Now}
	walletService := service.NewWalletService(runtime)
	result, err := walletService.CorrectPropertySaleDraftCharge(context.Background(), strings.TrimSpace(*listingID), *apply)
	if err != nil {
		log.Fatal(err)
	}
	if !result.CorrectionNeeded {
		log.Printf("no correction needed: listing=%s draft_points_paid=%d", result.ListingID, result.DraftPointsPaid)
		return
	}
	if !*apply {
		log.Printf("preview: listing=%s draft_points_paid=%d refund_points=%d", result.ListingID, result.DraftPointsPaid, result.RefundPoints)
		return
	}
	log.Printf("corrected: listing=%s draft_points_paid=%d refund_points=%d", result.ListingID, result.DraftPointsPaid, result.RefundPoints)
}

/*
 * Supermarket product image report service tests.
 * 1. Verify a product image report is stored once globally.
 * 2. Verify repeated reports return success without duplicate records.
 */
package service

import (
	"context"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"ajoliving_web/http_service/internal/model"
	"ajoliving_web/http_service/internal/utils"
)

// 1. TestCreateImageReportDeduplicatesProductCode verifies repeated reports create one database record.
func TestCreateImageReportDeduplicatesProductCode(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+utils.NewPublicID()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.SupermarketImageReport{}); err != nil {
		t.Fatalf("migrate image reports: %v", err)
	}

	supermarketService := NewSupermarketOfferService(&Runtime{DB: db})
	first, err := supermarketService.CreateImageReport(context.Background(), " p000000001 ")
	if err != nil {
		t.Fatalf("create first image report: %v", err)
	}
	if first["productCode"] != "P000000001" || first["created"] != true {
		t.Fatalf("expected normalized first report result, got %#v", first)
	}

	second, err := supermarketService.CreateImageReport(context.Background(), "P000000001")
	if err != nil {
		t.Fatalf("create duplicate image report: %v", err)
	}
	if second["created"] != false {
		t.Fatalf("expected duplicate report to succeed without creation, got %#v", second)
	}

	var count int64
	if err := db.Model(&model.SupermarketImageReport{}).Where("product_code = ?", "P000000001").Count(&count).Error; err != nil {
		t.Fatalf("count image reports: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected one image report, got %d", count)
	}
}

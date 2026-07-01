/*
 * 二手帖子服務回歸測試。
 * 1. 驗證家具地區篩選與前端區份 code 對齊。
 * 2. 驗證公開詳情狀態限制與業主可見性。
 * 3. 驗證編輯時未重填聯絡資料會保留已加密資料。
 */
package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"ajoliving_web/http_service/internal/config"
	"ajoliving_web/http_service/internal/database"
	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/model"
	"ajoliving_web/http_service/internal/utils"
)

// 1. TestSecondhandDistrictOptionsAlignWithFurnitureFrontend verifies accepted furniture districts.
func TestSecondhandDistrictOptionsAlignWithFurnitureFrontend(t *testing.T) {
	if !isAllowedSecondhandDistrict("eastern") || !isAllowedSecondhandDistrict("kwun_tong") {
		t.Fatal("expected frontend district codes to be accepted")
	}
	districts, ok := secondhandDistrictsForRegion("kowloon")
	if !ok {
		t.Fatal("expected kowloon region to be supported")
	}
	for _, expected := range []string{"kowloon", "yau_tsim_mong", "kwun_tong"} {
		found := false
		for _, district := range districts {
			if district == expected {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected district %s in kowloon region: %#v", expected, districts)
		}
	}
}

// 2. TestSecondhandPublicDetailRequiresPublishableState verifies public detail visibility.
func TestSecondhandPublicDetailRequiresPublishableState(t *testing.T) {
	secondhandService, db, ownerID := newSecondhandTestService(t)
	listingID := createSecondhandTestListing(t, db, ownerID, "draft", "approved", "available")

	if _, err := secondhandService.GetSecondhandDetail(context.Background(), listingID, nil, nil); !hasAppErrorCode(err, errcode.CodeNotFound) {
		t.Fatalf("expected public draft detail to be hidden as not found, got %v", err)
	}
	if _, err := secondhandService.GetSecondhandDetail(context.Background(), listingID, &ownerID, nil); err != nil {
		t.Fatalf("expected owner to read own draft detail: %v", err)
	}

	if err := db.Model(&model.Listing{}).Where("public_id = ?", listingID).Updates(map[string]any{
		"publication_status": "active",
		"business_status":    "sold",
	}).Error; err != nil {
		t.Fatalf("mark listing sold: %v", err)
	}
	if _, err := secondhandService.GetSecondhandDetail(context.Background(), listingID, nil, nil); !hasAppErrorCode(err, errcode.CodeNotFound) {
		t.Fatalf("expected sold public detail to be hidden as not found, got %v", err)
	}
}

// 3. TestUpdateSecondhandRetainsExistingDirectContact verifies edit payload does not erase encrypted contacts.
func TestUpdateSecondhandRetainsExistingDirectContact(t *testing.T) {
	secondhandService, db, ownerID := newSecondhandTestService(t)
	listingID := createSecondhandTestListing(t, db, ownerID, "draft", "approved", "available")

	if _, err := secondhandService.UpdateSecondhandListing(context.Background(), UpsertSecondhandParams{
		OwnerUserID:           ownerID,
		ListingPublicID:       listingID,
		Title:                 "更新餐桌",
		Summary:               "保留原有聯絡方式",
		Description:           "尺寸與交收資料已更新。",
		DistrictCode:          "eastern",
		PublisherIdentityType: "owner",
		CategoryCode:          "home_furniture",
		PriceMode:             "fixed",
		PriceHKD:              ptrFloat64(900),
		ConditionLevel:        "used_good",
		DimensionText:         "L 120 cm / W 60 cm",
		PickupRegionCode:      "eastern",
		PickupLocationText:    "屋苑大堂",
		VisibilityScope:       "public",
		ContactMethod:         "whatsapp",
		Contact: ListingContactInput{
			ShowWhatsApp: true,
		},
	}); err != nil {
		t.Fatalf("update secondhand listing: %v", err)
	}
	if err := db.Model(&model.Listing{}).Where("public_id = ?", listingID).Update("publication_status", "active").Error; err != nil {
		t.Fatalf("activate listing for contact access: %v", err)
	}

	result, err := secondhandService.GrantContactAccess(context.Background(), ownerID, nil, listingID, "127.0.0.1", "test")
	if err != nil {
		t.Fatalf("grant contact access: %v", err)
	}
	if result.ContactPayload["whatsapp_url"] == "" {
		t.Fatalf("expected retained whatsapp url in contact payload: %#v", result.ContactPayload)
	}
}

// 4. newSecondhandTestService creates an isolated sqlite-backed secondhand service.
func newSecondhandTestService(t *testing.T) (*SecondhandService, *gorm.DB, int64) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+utils.NewPublicID()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		t.Fatalf("migrate sqlite: %v", err)
	}

	runtime := &Runtime{
		Config: &config.Config{
			EncryptionKey:    "secondhand-service-test-key",
			MediaBaseURL:     "https://cdn.test",
			AppPublicBaseURL: "https://www.test",
		},
		DB:  db,
		Now: func() time.Time { return time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC) },
	}
	secondhandService := NewSecondhandService(runtime)

	owner := model.User{
		PublicID:         utils.NewPublicID(),
		PhoneCountryCode: "+852",
		PhoneNumber:      "61234567",
		MemberStatus:     "active",
		MemberType:       "user",
		IsVerifiedPhone:  true,
	}
	if err := db.Create(&owner).Error; err != nil {
		t.Fatalf("create owner: %v", err)
	}
	if err := db.Create(&model.UserProfile{
		UserID:                owner.ID,
		DisplayName:           "家具業主",
		PublisherIdentityType: "owner",
		DistrictCode:          "eastern",
	}).Error; err != nil {
		t.Fatalf("create owner profile: %v", err)
	}

	return secondhandService, db, owner.ID
}

// 5. createSecondhandTestListing inserts one complete listing aggregate.
func createSecondhandTestListing(t *testing.T, db *gorm.DB, ownerID int64, publicationStatus string, moderationStatus string, businessStatus string) string {
	t.Helper()
	listing := model.Listing{
		PublicID:              utils.NewPublicID(),
		Module:                "secondhand",
		OwnerUserID:           ownerID,
		Title:                 "實木餐桌",
		Summary:               "保養良好",
		Description:           "可於屋苑大堂交收。",
		DistrictCode:          "eastern",
		PublisherIdentityType: "owner",
		PublicationStatus:     publicationStatus,
		ModerationStatus:      moderationStatus,
		BusinessStatus:        businessStatus,
		IsDeleted:             false,
	}
	if err := db.Create(&listing).Error; err != nil {
		t.Fatalf("create listing: %v", err)
	}
	deliveryTags, err := marshalJSON([]string{"self_pickup"})
	if err != nil {
		t.Fatalf("marshal delivery tags: %v", err)
	}
	secondhand := model.SecondhandListing{
		ListingID:          listing.ID,
		CategoryCode:       "home_furniture",
		PriceMode:          "fixed",
		PriceHKD:           ptrFloat64(1200),
		ConditionLevel:     "used_good",
		PickupRegionCode:   "eastern",
		PickupLocationText: "屋苑大堂",
		DeliveryTags:       deliveryTags,
		VisibilityScope:    "public",
		ContactMethod:      "whatsapp",
	}
	if err := db.Create(&secondhand).Error; err != nil {
		t.Fatalf("create secondhand listing: %v", err)
	}
	whatsApp, err := utils.EncryptString("secondhand-service-test-key", "+852 6123 4567")
	if err != nil {
		t.Fatalf("encrypt whatsapp: %v", err)
	}
	contact := model.ListingContact{
		ListingID:         listing.ID,
		WhatsAppEncrypted: whatsApp,
		WhatsAppMasked:    utils.MaskPhone("+852 6123 4567"),
		ShowWhatsApp:      true,
		ContactMode:       "whatsapp",
	}
	if err := db.Create(&contact).Error; err != nil {
		t.Fatalf("create contact: %v", err)
	}

	return listing.PublicID
}

// 6. ptrFloat64 returns a float64 pointer.
func ptrFloat64(value float64) *float64 {
	return &value
}

// 7. hasAppErrorCode checks app error code equality.
func hasAppErrorCode(err error, code string) bool {
	var appErr *errcode.AppError
	return errors.As(err, &appErr) && appErr.Code == code
}

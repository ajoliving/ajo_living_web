/*
 * Property publishing flow tests.
 * 1. Build a complete property sale draft with image, contact, and wallet data.
 * 2. Publish the draft and verify public listing visibility and privacy fields.
 */
package service

import (
	"context"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"ajoliving_web/http_service/internal/config"
	"ajoliving_web/http_service/internal/database"
	"ajoliving_web/http_service/internal/model"
	"ajoliving_web/http_service/internal/utils"
)

// 1. TestCreateAndPublishPropertySaleListing verifies the full sale publish path.
func TestCreateAndPublishPropertySaleListing(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+utils.NewPublicID()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		t.Fatalf("migrate sqlite: %v", err)
	}

	now := time.Date(2026, 5, 25, 10, 0, 0, 0, time.UTC)
	runtime := &Runtime{
		Config: &config.Config{
			EncryptionKey:    "property-publish-flow-test-key",
			MediaBaseURL:     "https://cdn.test",
			AppPublicBaseURL: "https://www.test",
		},
		DB:  db,
		Now: func() time.Time { return now },
	}
	runtime.WalletService = NewWalletService(runtime)
	propertyService := NewPropertyService(runtime)

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
		DisplayName:           "測試業主",
		PublisherIdentityType: "owner",
		DistrictCode:          "hong_kong_island",
	}).Error; err != nil {
		t.Fatalf("create owner profile: %v", err)
	}
	if err := db.Create(&model.WalletAccount{
		UserID:  owner.ID,
		Balance: 5000,
	}).Error; err != nil {
		t.Fatalf("create wallet: %v", err)
	}

	cover := model.MediaAsset{
		PublicID:        utils.NewPublicID(),
		StorageProvider: "oss",
		BucketName:      "test",
		ObjectKey:       "property/cover.jpg",
		MimeType:        "image/jpeg",
		FileSize:        1024,
		ChecksumSHA256:  "property-cover-checksum",
		CreatedBy:       &owner.ID,
	}
	if err := db.Create(&cover).Error; err != nil {
		t.Fatalf("create media asset: %v", err)
	}

	detail, err := propertyService.CreatePropertySale(context.Background(), UpsertPropertySaleParams{
		OwnerUserID:           owner.ID,
		PropertyNo:            "TEST-PROP-001",
		Title:                 "仁英大廈高層放售",
		TitleEn:               "Yen Ying Building high floor for sale",
		Summary:               "灣仔核心地段實用兩房",
		Description:           "實用間隔，交通方便，適合自住或投資。",
		DescriptionEn:         "Practical layout in Wan Chai with convenient transport.",
		DistrictCode:          "hong_kong_island",
		PublisherIdentityType: "owner",
		TransactionType:       "sale",
		LocationScope:         "local",
		ListingCategory:       "standard",
		PropertyType:          "private_residential",
		EstateName:            "仁英大廈",
		AddressText:           "謝斐道221號",
		AddressTextEn:         "No.221 Jaffe Road",
		BlockName:             "仁英大廈",
		UnitName:              "A",
		ShowUnit:              false,
		AskingPriceHKD:        6800000,
		PriceReferenceOnly:    true,
		AreaMode:              "usable",
		UsableAreaSqft:        420,
		GrossAreaSqft:         ptrInt(560),
		BedroomCount:          2,
		LivingRoomCount:       1,
		BathroomCount:         1,
		FloorRaw:              "25",
		FloorZone:             "high",
		TotalFloors:           40,
		Direction:             "東南",
		BuildingAge:           "49",
		KitchenType:           "開放式廚房",
		CookingMode:           "明火",
		ManagementFeeHKD:      1200,
		PrivateNote:           "內部測試記事",
		AdPackageCode:         "premium",
		FeatureTags:           []string{"near_mtr", "high_floor"},
		ContactMethod:         "phone",
		Images: []ListingImageInput{
			{MediaAssetID: cover.PublicID, SortOrder: 1, IsCover: true},
		},
		Contact: PropertyContactInput{
			Phone:           "61234567",
			ShowPhone:       true,
			ShowInquiryForm: true,
		},
	})
	if err != nil {
		t.Fatalf("create property sale: %v", err)
	}
	if detail.ListingID == "" || detail.PublicationStatus != "draft" {
		t.Fatalf("unexpected draft detail: %#v", detail)
	}
	if detail.PropertySale == nil || detail.PropertySale.UnitName != "A" || detail.PropertySale.PrivateNote == "" {
		t.Fatalf("owner draft detail did not retain private fields: %#v", detail.PropertySale)
	}
	if !detail.PropertySale.PriceReferenceOnly || detail.PropertySale.PriceNegotiable {
		t.Fatalf("unexpected draft price flags: %#v", detail.PropertySale)
	}

	published, err := propertyService.PublishProperty(context.Background(), PropertyChannelSale, owner.ID, detail.ListingID)
	if err != nil {
		t.Fatalf("publish property sale: %v", err)
	}
	if published.PublicationStatus != "active" || published.PointsCharged != 1500 {
		t.Fatalf("unexpected published detail: %#v", published)
	}
	if published.PropertySale == nil || published.PropertySale.AdPackageCode != "premium" || published.PropertySale.AdWeight != 2 {
		t.Fatalf("unexpected published sale payload: %#v", published.PropertySale)
	}

	items, pagination, err := propertyService.ListPublicProperties(context.Background(), PropertyChannelSale, PropertyListFilters{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("list public properties: %v", err)
	}
	if pagination.Total != 1 || len(items) != 1 {
		t.Fatalf("expected one public listing, total=%d len=%d", pagination.Total, len(items))
	}
	publicSale := items[0].PropertySale
	if publicSale == nil {
		t.Fatal("expected public sale payload")
	}
	if publicSale.UnitName != "" || publicSale.FloorRaw != "" || publicSale.PrivateNote != "" {
		t.Fatalf("public payload leaked private fields: %#v", publicSale)
	}
	if !publicSale.PriceReferenceOnly || publicSale.PriceNegotiable {
		t.Fatalf("unexpected public price flags: %#v", publicSale)
	}
	if publicSale.PublicLocationText != "仁英大廈高層(21-25|40/F)" {
		t.Fatalf("unexpected public location text: %q", publicSale.PublicLocationText)
	}
}

// 2. ptrInt returns an int pointer for optional area fields.
func ptrInt(value int) *int {
	return &value
}

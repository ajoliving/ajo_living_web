/*
 * Property publishing flow tests.
 * 1. Build a complete property sale draft with image, contact, and wallet data.
 * 2. Publish the draft and verify public listing visibility and privacy fields.
 * 3. Verify serviced apartment field derivation for xlsx range fields.
 * 4. Verify residential owner and agent form validation rules.
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
		Title:                 "仁英大廈高層放售",
		TitleEn:               "High floor unit in Yan Yee Building",
		Summary:               "灣仔核心地段實用兩房",
		Description:           "實用間隔，交通方便，適合自住或投資。",
		DescriptionEn:         "Practical layout with convenient transport, suitable for self-use or investment.",
		DistrictCode:          "hong_kong_island",
		PublisherIdentityType: "owner",
		TransactionType:       "sale",
		LocationScope:         "local",
		ListingCategory:       "standard",
		PropertyType:          "private_residential",
		EstateName:            "仁英大廈",
		AddressText:           "謝斐道221號",
		AddressTextEn:         "221 Jaffe Road",
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
		FeatureTags:           []string{"residential_private_estate", "near_mtr", "high_floor"},
		ContactMethod:         "phone",
		Images: []ListingImageInput{
			{MediaAssetID: cover.PublicID, SortOrder: 1, IsCover: true},
		},
		Contact: PropertyContactInput{
			ContactNameZH:   "測試業主",
			ContactNameEN:   "Test Owner",
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
	if detail.PropertySale.PropertyNo != detail.ListingID || detail.PropertySale.TitleEn != "High floor unit in Yan Yee Building" || detail.PropertySale.AddressTextEn != "221 Jaffe Road" {
		t.Fatalf("draft detail did not retain english fields: %#v", detail.PropertySale)
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

	if err := propertyService.DeactivateProperty(context.Background(), PropertyChannelSale, owner.ID, detail.ListingID); err != nil {
		t.Fatalf("deactivate property sale: %v", err)
	}
	hiddenDetail, err := propertyService.GetPropertyDetail(context.Background(), PropertyChannelSale, detail.ListingID, &owner.ID)
	if err != nil {
		t.Fatalf("load hidden property sale: %v", err)
	}
	if hiddenDetail.PublicationStatus != "hidden" {
		t.Fatalf("expected hidden property sale, got %#v", hiddenDetail)
	}
	republished, err := propertyService.RepublishProperty(context.Background(), PropertyChannelSale, owner.ID, detail.ListingID)
	if err != nil {
		t.Fatalf("republish hidden property sale: %v", err)
	}
	if republished.PublicationStatus != "active" || republished.BusinessStatus != "available" || republished.PointsCharged != 1500 {
		t.Fatalf("unexpected republished detail: %#v", republished)
	}
}

// 2. TestValidateResidentialSaleModes verifies owner and agent residential form rules.
func TestValidateResidentialSaleModes(t *testing.T) {
	propertyService := &PropertyService{}
	base := UpsertPropertySaleParams{
		Title:                 "康怡花園住宅放租",
		TitleEn:               "Residential listing in Kornhill",
		Description:           "實用住宅單位。",
		DescriptionEn:         "Practical residential unit.",
		DistrictCode:          "hong_kong_island",
		PublisherIdentityType: "owner",
		TransactionType:       "rent",
		LocationScope:         "local",
		ListingCategory:       "standard",
		PropertyType:          "residential",
		EstateName:            "康怡花園",
		AddressText:           "康山道1號",
		AddressTextEn:         "1 Kornhill Road",
		MonthlyRentHKD:        18000,
		AreaMode:              "usable",
		UsableAreaSqft:        420,
		FeatureTags:           []string{"residential_private_estate", "feature_student_friendly"},
		ContactMethod:         "phone",
		AdPackageCode:         "basic",
		Contact: PropertyContactInput{
			ContactNameZH: "陳先生",
			ContactNameEN: "Mr Chan",
			Phone:         "61234567",
			ShowPhone:     true,
		},
	}
	if err := propertyService.validateSaleParams(base); err != nil {
		t.Fatalf("owner residential rent should pass without floor: %v", err)
	}

	agentSale := base
	agentSale.PublisherIdentityType = "agent"
	agentSale.TransactionType = "sale"
	agentSale.PropertyNo = "AJO-1001"
	agentSale.MonthlyRentHKD = 0
	agentSale.AskingPriceHKD = 6800000
	agentSale.FeatureTags = []string{"residential_private_estate"}
	agentSale.ContactMethod = "chat"
	agentSale.Contact = PropertyContactInput{
		ContactAttributes: map[string]string{
			"agency_company_profile": "AJO Agency",
			"agency_contact_profile": "Agent Chan",
		},
		ShowChat: true,
	}
	if err := propertyService.validateSaleParams(agentSale); err != nil {
		t.Fatalf("agent residential sale should pass with agency profiles: %v", err)
	}

	missingCategory := base
	missingCategory.FeatureTags = []string{"feature_mtr"}
	if err := propertyService.validateSaleParams(missingCategory); err == nil {
		t.Fatal("expected missing residential category tag to fail")
	}

	saleWithStudentTag := agentSale
	saleWithStudentTag.FeatureTags = []string{"residential_private_estate", "feature_student_friendly"}
	if err := propertyService.validateSaleParams(saleWithStudentTag); err == nil {
		t.Fatal("expected residential sale with student friendly tag to fail")
	}
}

// 3. TestValidateNonResidentialSaleModes verifies xlsx category and area rules.
func TestValidateNonResidentialSaleModes(t *testing.T) {
	propertyService := &PropertyService{}
	base := UpsertPropertySaleParams{
		Title:                 "車位放租",
		TitleEn:               "Car park rental",
		Description:           "實用車位。",
		DescriptionEn:         "Practical car park.",
		DistrictCode:          "hong_kong_island",
		PublisherIdentityType: "owner",
		TransactionType:       "rent",
		LocationScope:         "local",
		ListingCategory:       "standard",
		PropertyType:          "car_park",
		EstateName:            "中環大廈",
		AddressText:           "皇后大道中1號",
		AddressTextEn:         "1 Queen's Road Central",
		MonthlyRentHKD:        3200,
		FloorRaw:              "N/A",
		FeatureTags:           []string{"car_park_residential"},
		ContactMethod:         "phone",
		AdPackageCode:         "basic",
		Contact: PropertyContactInput{
			ContactNameZH: "陳先生",
			ContactNameEN: "Mr Chan",
			Phone:         "61234567",
			ShowPhone:     true,
		},
	}
	if err := propertyService.validateSaleParams(base); err != nil {
		t.Fatalf("car park listing should pass: %v", err)
	}

	missingCarParkCategory := base
	missingCarParkCategory.FeatureTags = nil
	if err := propertyService.validateSaleParams(missingCarParkCategory); err == nil {
		t.Fatal("expected missing car park category tag to fail")
	}

	industrial := base
	industrial.Title = "工商放租"
	industrial.TitleEn = "Commercial rental"
	industrial.PropertyType = "industrial"
	industrial.GrossAreaSqft = ptrInt(800)
	industrial.UsableAreaSqft = 600
	industrial.FeatureTags = nil
	if err := propertyService.validateSaleParams(industrial); err != nil {
		t.Fatalf("industrial category should remain optional: %v", err)
	}

	shop := industrial
	shop.Title = "店鋪放租"
	shop.TitleEn = "Shop rental"
	shop.PropertyType = "shop"
	shop.FeatureTags = []string{"shop_street"}
	if err := propertyService.validateSaleParams(shop); err != nil {
		t.Fatalf("shop listing should pass with shop category: %v", err)
	}
	shopMissingGross := shop
	shopMissingGross.GrossAreaSqft = nil
	if err := propertyService.validateSaleParams(shopMissingGross); err == nil {
		t.Fatal("expected shop missing gross area to fail")
	}

	land := industrial
	land.Title = "土地放售"
	land.TitleEn = "Land sale"
	land.TransactionType = "sale"
	land.PropertyType = "land"
	land.PropertyNo = "LAND-001"
	land.MonthlyRentHKD = 0
	land.AskingPriceHKD = 3800000
	land.FeatureTags = []string{"land_farmland"}
	if err := propertyService.validateSaleParams(land); err != nil {
		t.Fatalf("land listing should pass with land category: %v", err)
	}
}

// 4. TestValidateServicedApartmentSheetFields verifies serviced apartment xlsx field rules.
func TestValidateServicedApartmentSheetFields(t *testing.T) {
	propertyService := &PropertyService{}
	base := UpsertServicedApartmentParams{
		Title:                 "灣仔服務式住宅",
		Summary:               "交通方便。",
		Description:           "交通方便。",
		DistrictCode:          "hong_kong_island",
		PublisherIdentityType: "operator",
		ProjectName:           "灣仔服務式住宅",
		AddressText:           "灣仔道1號",
		LowestMonthlyRentHKD:  18000,
		MinUsableAreaSqft:     180,
		MaxUsableAreaSqft:     260,
		MinStayValue:          1,
		MinStayUnit:           "month",
		LocationScope:         "local",
		ListingCategory:       "standard",
		AdPackageCode:         "basic",
		ContactMethod:         "phone",
		RoomTypes: []ServicedApartmentRoomTypeInput{
			{
				Name:              "開放式",
				RoomCategory:      "studio",
				UsableAreaMinSqft: 180,
				UsableAreaMaxSqft: 260,
				MonthlyRentMinHKD: 18000,
				MonthlyRentMaxHKD: 22000,
				RentUnit:          "month",
				MinStayValue:      1,
				MinStayUnit:       "month",
				IncludedFeeItems:  []string{"appliance_tv"},
				ImageMediaAssetID: "MEDIA-001",
				PageURL:           "https://example.com/rooms/studio",
			},
		},
		Contact: PropertyContactInput{
			Phone:     "61234567",
			ShowPhone: true,
		},
	}
	if err := propertyService.validateServicedParams(base); err != nil {
		t.Fatalf("serviced apartment should pass: %v", err)
	}

	missingContact := base
	missingContact.Contact = PropertyContactInput{ShowChat: true}
	if err := propertyService.validateServicedParams(missingContact); err == nil {
		t.Fatal("expected missing serviced contact to fail")
	}

	invalidAreaRange := base
	invalidAreaRange.MaxUsableAreaSqft = 100
	if err := propertyService.validateServicedParams(invalidAreaRange); err == nil {
		t.Fatal("expected invalid serviced area range to fail")
	}
}

// 5. ptrInt returns an int pointer for optional area fields.
func ptrInt(value int) *int {
	return &value
}

// 6. TestDeriveServicedApartmentRentRange verifies project-level room rent range fields.
func TestDeriveServicedApartmentRentRange(t *testing.T) {
	derived := deriveServicedApartmentFields(UpsertServicedApartmentParams{
		LowestMonthlyRentHKD:  20000,
		HighestMonthlyRentHKD: 26000,
		RoomTypes: []ServicedApartmentRoomTypeInput{
			{
				Name:              "開放式",
				UsableAreaMinSqft: 180,
				UsableAreaMaxSqft: 240,
				MonthlyRentMinHKD: 18000,
				MonthlyRentMaxHKD: 22000,
				MinStayValue:      1,
				MinStayUnit:       "month",
			},
			{
				Name:              "一房",
				UsableAreaMinSqft: 260,
				UsableAreaMaxSqft: 360,
				MonthlyRentMinHKD: 23000,
				MonthlyRentMaxHKD: 30000,
				MinStayValue:      2,
				MinStayUnit:       "week",
			},
		},
	})

	if derived.LowestMonthlyRentHKD != 18000 || derived.HighestMonthlyRentHKD != 30000 {
		t.Fatalf("unexpected serviced rent range: %#v", derived)
	}
	if derived.MinUsableAreaSqft != 180 || derived.MaxUsableAreaSqft != 360 {
		t.Fatalf("unexpected serviced area range: %#v", derived)
	}
}

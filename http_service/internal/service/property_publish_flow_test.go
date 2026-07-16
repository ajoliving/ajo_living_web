/*
 * Property publishing flow tests.
 * 1. Build a complete property sale draft with image, contact, and wallet data.
 * 2. Publish the draft and verify public listing visibility and privacy fields.
 * 3. Verify serviced apartment field derivation for xlsx range fields.
 * 4. Verify residential owner and agent form validation rules.
 * 5. Verify sale publisher identity comes from the registered account.
 */
package service

import (
	"context"
	"strings"
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

	agent := model.User{
		PublicID:         utils.NewPublicID(),
		PhoneCountryCode: "+852",
		PhoneNumber:      "62345678",
		MemberStatus:     "active",
		MemberType:       "user",
		IsVerifiedPhone:  true,
	}
	if err := db.Create(&agent).Error; err != nil {
		t.Fatalf("create agent: %v", err)
	}
	if err := db.Create(&model.UserProfile{UserID: agent.ID, DisplayName: "測試代理", PublisherIdentityType: "agent"}).Error; err != nil {
		t.Fatalf("create agent profile: %v", err)
	}
	agentDraft, err := propertyService.CreatePropertySale(context.Background(), UpsertPropertySaleParams{
		OwnerUserID:           agent.ID,
		PublisherIdentityType: "owner",
	})
	if err != nil {
		t.Fatalf("create agent property draft: %v", err)
	}
	if agentDraft.PublisherIdentityType != "owner" {
		t.Fatalf("expected requested owner identity to override profile, got %q", agentDraft.PublisherIdentityType)
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

	incompleteDraft, err := propertyService.CreatePropertySale(context.Background(), UpsertPropertySaleParams{
		OwnerUserID:           owner.ID,
		PublisherIdentityType: "owner",
	})
	if err != nil {
		t.Fatalf("create incomplete property draft: %v", err)
	}
	if incompleteDraft.ListingID == "" || incompleteDraft.PublicationStatus != "draft" {
		t.Fatalf("unexpected incomplete draft detail: %#v", incompleteDraft)
	}
	updatedIncompleteDraft, err := propertyService.UpdatePropertySale(context.Background(), UpsertPropertySaleParams{
		OwnerUserID:           owner.ID,
		ListingPublicID:       incompleteDraft.ListingID,
		Title:                 "未完成草稿",
		PublisherIdentityType: "owner",
	})
	if err != nil {
		t.Fatalf("update incomplete property draft: %v", err)
	}
	if updatedIncompleteDraft.Title != "未完成草稿" || updatedIncompleteDraft.PublicationStatus != "draft" || updatedIncompleteDraft.PublisherIdentityType != "owner" {
		t.Fatalf("unexpected updated incomplete draft detail: %#v", updatedIncompleteDraft)
	}
	if _, err := propertyService.PublishProperty(context.Background(), PropertyChannelSale, owner.ID, incompleteDraft.ListingID); err == nil {
		t.Fatal("expected incomplete draft publication to fail")
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
			ContactNameZH: "測試業主",
			ContactNameEN: "Test Owner",
			Phone:         "61234567",
			Phone2:        "62345678",
			WhatsApp:      "61234567",
			ContactAttributes: map[string]string{
				"phone_country_code":       "+852",
				"phone_2_country_code":     "+86",
				"phone_whatsapp_enabled":   "yes",
				"phone_2_whatsapp_enabled": "yes",
			},
			ShowPhone:       true,
			ShowWhatsApp:    true,
			ShowInquiryForm: true,
		},
	})
	if err != nil {
		t.Fatalf("create property sale: %v", err)
	}
	if detail.ListingID == "" || detail.PublicationStatus != "draft" {
		t.Fatalf("unexpected draft detail: %#v", detail)
	}
	if detail.PublisherIdentityType != "owner" {
		t.Fatalf("expected profile owner identity to override request, got %q", detail.PublisherIdentityType)
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
	contactAccess, err := propertyService.GrantPropertyContactAccess(
		context.Background(),
		PropertyChannelSale,
		owner.ID,
		detail.ListingID,
		"127.0.0.1",
		"property-test",
	)
	if err != nil {
		t.Fatalf("grant property contact access: %v", err)
	}
	if !strings.Contains(contactAccess.ContactPayload["phone_whatsapp_url"], "wa.me/85261234567") {
		t.Fatalf("unexpected phone 1 whatsapp url: %q", contactAccess.ContactPayload["phone_whatsapp_url"])
	}
	if !strings.Contains(contactAccess.ContactPayload["phone_2_whatsapp_url"], "wa.me/8662345678") {
		t.Fatalf("unexpected phone 2 whatsapp url: %q", contactAccess.ContactPayload["phone_2_whatsapp_url"])
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

	if err := db.Model(&model.UserProfile{}).Where("user_id = ?", owner.ID).Update("publisher_identity_type", "tenant").Error; err != nil {
		t.Fatalf("change account identity for publication check: %v", err)
	}
	if err := propertyService.DeactivateProperty(context.Background(), PropertyChannelSale, owner.ID, detail.ListingID); err != nil {
		t.Fatalf("deactivate property for identity check: %v", err)
	}
	legacyRepublished, err := propertyService.RepublishProperty(context.Background(), PropertyChannelSale, owner.ID, detail.ListingID)
	if err != nil {
		t.Fatalf("republish property with legacy account identity: %v", err)
	}
	if legacyRepublished.PublisherIdentityType != "owner" {
		t.Fatalf("expected legacy account identity to default to owner, got %q", legacyRepublished.PublisherIdentityType)
	}
	ownerDraft, err := propertyService.CreatePropertySale(context.Background(), UpsertPropertySaleParams{
		OwnerUserID:           owner.ID,
		PublisherIdentityType: "owner",
	})
	if err != nil {
		t.Fatalf("create property draft with selected owner identity: %v", err)
	}
	if ownerDraft.PublisherIdentityType != "owner" {
		t.Fatalf("expected selected owner draft identity, got %q", ownerDraft.PublisherIdentityType)
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
	land.EstateName = ""
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

// 5. TestPropertySaleDraftSaveChargesPoints verifies draft persistence and billing are atomic.
func TestPropertySaleDraftSaveChargesPoints(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+utils.NewPublicID()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		t.Fatalf("migrate sqlite: %v", err)
	}

	now := time.Date(2026, 7, 15, 10, 0, 0, 0, time.UTC)
	runtime := &Runtime{
		Config: &config.Config{EncryptionKey: "property-draft-charge-test-key"},
		DB:     db,
		Now:    func() time.Time { return now },
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
	if err := db.Create(&model.UserProfile{UserID: owner.ID, PublisherIdentityType: "owner"}).Error; err != nil {
		t.Fatalf("create owner profile: %v", err)
	}
	if err := db.Create(&model.WalletAccount{UserID: owner.ID, Balance: 2500}).Error; err != nil {
		t.Fatalf("create wallet: %v", err)
	}

	draft, err := propertyService.CreatePropertySale(context.Background(), UpsertPropertySaleParams{
		OwnerUserID:     owner.ID,
		Title:           "首次草稿",
		ChargeDraftSave: true,
	})
	if err != nil {
		t.Fatalf("create charged property draft: %v", err)
	}
	if draft.PublicationStatus != "draft" || draft.PointsCharged != 1000 || draft.PointsBalanceAfter == nil || *draft.PointsBalanceAfter != 1500 {
		t.Fatalf("unexpected charged draft detail: %#v", draft)
	}

	now = now.Add(time.Second)
	updated, err := propertyService.UpdatePropertySale(context.Background(), UpsertPropertySaleParams{
		OwnerUserID:     owner.ID,
		ListingPublicID: draft.ListingID,
		Title:           "第二次草稿",
		ChargeDraftSave: true,
	})
	if err != nil {
		t.Fatalf("update charged property draft: %v", err)
	}
	if updated.Title != "第二次草稿" || updated.PointsCharged != 1000 || updated.PointsBalanceAfter == nil || *updated.PointsBalanceAfter != 500 {
		t.Fatalf("unexpected updated charged draft: %#v", updated)
	}

	now = now.Add(time.Second)
	if _, err := propertyService.UpdatePropertySale(context.Background(), UpsertPropertySaleParams{
		OwnerUserID:     owner.ID,
		ListingPublicID: draft.ListingID,
		Title:           "不應保存的草稿",
		ChargeDraftSave: true,
	}); err == nil {
		t.Fatal("expected insufficient points draft save to fail")
	}
	retained, err := propertyService.GetPropertyDetail(context.Background(), PropertyChannelSale, draft.ListingID, &owner.ID)
	if err != nil {
		t.Fatalf("load retained property draft: %v", err)
	}
	if retained.Title != "第二次草稿" {
		t.Fatalf("insufficient points changed draft title: %q", retained.Title)
	}
	var wallet model.WalletAccount
	if err := db.Where("user_id = ?", owner.ID).First(&wallet).Error; err != nil {
		t.Fatalf("load wallet: %v", err)
	}
	if wallet.Balance != 500 || wallet.TotalSpent != 2000 {
		t.Fatalf("unexpected wallet after draft saves: %#v", wallet)
	}
}

// 6. ptrInt returns an int pointer for optional area fields.
func ptrInt(value int) *int {
	return &value
}

// 7. TestDeriveServicedApartmentRentRange verifies project-level room rent range fields.
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

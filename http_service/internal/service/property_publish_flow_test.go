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
	"errors"
	"fmt"
	"strings"
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

	createParams := UpsertPropertySaleParams{
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
		ChargeDraftSave:       true,
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
	}
	detail, err := propertyService.CreatePropertySale(context.Background(), createParams)
	if err != nil {
		t.Fatalf("create property sale: %v", err)
	}
	if detail.ListingID == "" || detail.PublicationStatus != "draft" {
		t.Fatalf("unexpected draft detail: %#v", detail)
	}
	if detail.PointsCharged != PropertySaleDraftCost || detail.PointsBalanceAfter == nil || *detail.PointsBalanceAfter != 4400 {
		t.Fatalf("unexpected property draft prepayment: %#v", detail)
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
	if detail.ContactSummary.EditableContact == nil ||
		detail.ContactSummary.EditableContact.ContactNameZH != "測試業主" ||
		detail.ContactSummary.EditableContact.Phone != "61234567" ||
		detail.ContactSummary.EditableContact.Phone2 != "62345678" ||
		detail.ContactSummary.EditableContact.WhatsApp != "61234567" {
		t.Fatalf("owner draft detail did not return editable contact: %#v", detail.ContactSummary.EditableContact)
	}
	myItems, _, err := propertyService.ListMyProperties(context.Background(), PropertyChannelSale, owner.ID, PropertyListFilters{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("list owner property drafts: %v", err)
	}
	foundChargedDraft := false
	for _, item := range myItems {
		if item.ListingID != detail.ListingID {
			continue
		}
		if item.DraftPointsPaid == nil || item.PublishPointsTotal == nil || item.PublishPointsDue == nil || *item.DraftPointsPaid != PropertySaleDraftCost || *item.PublishPointsTotal != PropertySalePublishCost || *item.PublishPointsDue != 400 {
			t.Fatalf("unexpected owner publish charge summary: %#v", item)
		}
		foundChargedDraft = true
		break
	}
	if !foundChargedDraft {
		t.Fatal("expected charged draft in owner property list")
	}

	published, err := propertyService.PublishProperty(context.Background(), PropertyChannelSale, owner.ID, detail.ListingID)
	if err != nil {
		t.Fatalf("publish property sale: %v", err)
	}
	if published.PublicationStatus != "active" || published.PointsCharged != 400 {
		t.Fatalf("unexpected published detail: %#v", published)
	}
	if published.PropertySale == nil || published.PropertySale.AdPackageCode != "premium" || published.PropertySale.AdWeight != 2 {
		t.Fatalf("unexpected published sale payload: %#v", published.PropertySale)
	}
	var publishedListing model.Listing
	if err := db.Where("public_id = ?", detail.ListingID).First(&publishedListing).Error; err != nil {
		t.Fatalf("load published property sale: %v", err)
	}
	if publishedListing.ExpireAt == nil {
		t.Fatal("expected published property sale expiry")
	}
	publishedExpiry := *publishedListing.ExpireAt

	now = now.Add(time.Hour)
	renewed, err := propertyService.RenewPropertySale(context.Background(), owner.ID, detail.ListingID)
	if err != nil {
		t.Fatalf("renew active property sale: %v", err)
	}
	if renewed.PointsCharged != 750 {
		t.Fatalf("expected premium renewal to charge 750 points, got %#v", renewed)
	}
	if err := db.Where("public_id = ?", detail.ListingID).First(&publishedListing).Error; err != nil {
		t.Fatalf("load renewed property sale: %v", err)
	}
	if publishedListing.ExpireAt == nil || !publishedListing.ExpireAt.Equal(publishedExpiry.AddDate(0, 1, 0)) {
		t.Fatalf("expected calendar-month renewal from %s, got %#v", publishedExpiry, publishedListing.ExpireAt)
	}

	createParams.ListingPublicID = detail.ListingID
	createParams.Title = "仁英大廈高層放售更新"
	updatedActive, err := propertyService.UpdatePropertySale(context.Background(), createParams)
	if err != nil {
		t.Fatalf("update active property sale: %v", err)
	}
	if updatedActive.PointsCharged != 0 || updatedActive.Title != createParams.Title {
		t.Fatalf("active update should save without charging before republish: %#v", updatedActive)
	}

	now = now.Add(time.Hour)
	activeRepublished, err := propertyService.RepublishProperty(context.Background(), PropertyChannelSale, owner.ID, detail.ListingID)
	if err != nil {
		t.Fatalf("republish active property sale: %v", err)
	}
	if activeRepublished.PointsCharged != 750 || activeRepublished.Title != createParams.Title {
		t.Fatalf("active republish should charge the premium half price: %#v", activeRepublished)
	}
	if err := db.Where("public_id = ?", detail.ListingID).First(&publishedListing).Error; err != nil {
		t.Fatalf("load active republished property sale: %v", err)
	}
	if publishedListing.ExpireAt == nil || !publishedListing.ExpireAt.Equal(now.Add(30*24*time.Hour)) {
		t.Fatalf("unexpected active republish expiry: %#v", publishedListing.ExpireAt)
	}
	var activeRepublishTransaction model.WalletTransaction
	if err := db.Where("listing_id = ? AND action_type = ?", publishedListing.ID, WalletActionRepublish).Order("id desc").First(&activeRepublishTransaction).Error; err != nil {
		t.Fatalf("load active republish transaction: %v", err)
	}
	if activeRepublishTransaction.Amount != 750 || activeRepublishTransaction.ActionType != WalletActionRepublish {
		t.Fatalf("unexpected active republish transaction: %#v", activeRepublishTransaction)
	}
	publicDetail, err := propertyService.GetPropertyDetail(context.Background(), PropertyChannelSale, detail.ListingID, nil)
	if err != nil {
		t.Fatalf("load public property detail: %v", err)
	}
	if publicDetail.ContactSummary.EditableContact != nil {
		t.Fatalf("public detail exposed editable contact: %#v", publicDetail.ContactSummary.EditableContact)
	}
	if publicDetail.PropertySale == nil || publicDetail.PropertySale.FloorRaw != "25" || publicDetail.PropertySale.FloorZone != "high" {
		t.Fatalf("public detail should expose actual and public floor values: %#v", publicDetail.PropertySale)
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
	if publicSale.UnitName != "" || publicSale.PrivateNote != "" {
		t.Fatalf("public payload leaked private fields: %#v", publicSale)
	}
	if publicSale.FloorRaw != "25" || publicSale.FloorZone != "high" {
		t.Fatalf("public payload should expose actual and public floor values: %#v", publicSale)
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
	if republished.PublicationStatus != "active" || republished.BusinessStatus != "available" || republished.PointsCharged != 750 {
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
	} else {
		var validationError *errcode.AppError
		if !errors.As(err, &validationError) || len(validationError.Errors) != 1 || validationError.Errors[0].Field != "feature_tags" {
			t.Fatalf("expected feature_tags field validation error, got %#v", err)
		}
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
	if draft.PublicationStatus != "draft" || draft.PointsCharged != PropertySaleDraftCost || draft.PointsBalanceAfter == nil || *draft.PointsBalanceAfter != 1900 {
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
	if updated.Title != "第二次草稿" || updated.PointsCharged != 0 || updated.PointsBalanceAfter != nil {
		t.Fatalf("unexpected updated charged draft: %#v", updated)
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
	if wallet.Balance != 1900 || wallet.TotalSpent != PropertySaleDraftCost {
		t.Fatalf("unexpected wallet after draft saves: %#v", wallet)
	}
}

// 6. TestCorrectPropertySaleDraftCharge verifies legacy draft overcharges are refunded once and credited correctly.
func TestCorrectPropertySaleDraftCharge(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+utils.NewPublicID()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		t.Fatalf("migrate sqlite: %v", err)
	}

	now := time.Date(2026, 7, 28, 10, 0, 0, 0, time.UTC)
	runtime := &Runtime{Config: &config.Config{}, DB: db, Now: func() time.Time { return now }}
	walletService := NewWalletService(runtime)
	runtime.WalletService = walletService
	propertyService := NewPropertyService(runtime)
	owner := model.User{PublicID: utils.NewPublicID(), PhoneCountryCode: "+852", PhoneNumber: "61234567", MemberStatus: "active", MemberType: "user", IsVerifiedPhone: true}
	if err := db.Create(&owner).Error; err != nil {
		t.Fatalf("create owner: %v", err)
	}
	if err := db.Create(&model.WalletAccount{UserID: owner.ID, Balance: 2500}).Error; err != nil {
		t.Fatalf("create wallet: %v", err)
	}
	listing := model.Listing{PublicID: utils.NewPublicID(), Module: string(PropertyChannelSale), OwnerUserID: owner.ID, PublicationStatus: "draft", ModerationStatus: "approved", BusinessStatus: "available"}
	if err := db.Create(&listing).Error; err != nil {
		t.Fatalf("create property draft: %v", err)
	}

	for index := 0; index < 2; index++ {
		if err := db.Transaction(func(tx *gorm.DB) error {
			_, err := walletService.SpendPointsWithTx(context.Background(), tx, WalletSpendParams{
				UserID:         owner.ID,
				Amount:         1000,
				BizModule:      string(PropertyChannelSale),
				ActionType:     WalletActionSaveDraft,
				ListingID:      &listing.ID,
				IdempotencyKey: fmt.Sprintf("legacy-draft-charge:%d", index),
			})
			return err
		}); err != nil {
			t.Fatalf("create legacy draft debit %d: %v", index, err)
		}
	}

	preview, err := walletService.CorrectPropertySaleDraftCharge(context.Background(), listing.PublicID, false)
	if err != nil {
		t.Fatalf("preview draft correction: %v", err)
	}
	if !preview.CorrectionNeeded || preview.DraftPointsPaid != 2000 || preview.RefundPoints != 1400 {
		t.Fatalf("unexpected correction preview: %#v", preview)
	}
	corrected, err := walletService.CorrectPropertySaleDraftCharge(context.Background(), listing.PublicID, true)
	if err != nil {
		t.Fatalf("apply draft correction: %v", err)
	}
	if !corrected.CorrectionNeeded || corrected.DraftPointsPaid != PropertySaleDraftCost || corrected.RefundPoints != 1400 {
		t.Fatalf("unexpected correction result: %#v", corrected)
	}
	repeated, err := walletService.CorrectPropertySaleDraftCharge(context.Background(), listing.PublicID, true)
	if err != nil {
		t.Fatalf("repeat draft correction: %v", err)
	}
	if repeated.CorrectionNeeded || repeated.DraftPointsPaid != PropertySaleDraftCost || repeated.RefundPoints != 0 {
		t.Fatalf("unexpected repeated correction result: %#v", repeated)
	}
	paid, err := propertyService.propertyDraftPointsPaid(context.Background(), db, listing.ID)
	if err != nil {
		t.Fatalf("load corrected draft payment: %v", err)
	}
	if paid != PropertySaleDraftCost || propertyPublishPointsDue(PropertySalePublishCost, paid) != 400 {
		t.Fatalf("unexpected corrected publish settlement: paid=%d", paid)
	}
	var wallet model.WalletAccount
	if err := db.Where("user_id = ?", owner.ID).First(&wallet).Error; err != nil {
		t.Fatalf("load corrected wallet: %v", err)
	}
	if wallet.Balance != 1900 || wallet.TotalSpent != 2000 || wallet.TotalEarned != 1400 {
		t.Fatalf("unexpected corrected wallet: %#v", wallet)
	}
}

// 7. TestPropertyPublishPointsDue verifies draft prepayment credits never exceed the total publish fee.
func TestPropertyPublishPointsDue(t *testing.T) {
	if due := propertyPublishPointsDue(PropertySalePublishCost, PropertySaleDraftCost); due != 400 {
		t.Fatalf("expected 400 points due after draft prepayment, got %d", due)
	}
	if due := propertyPublishPointsDue(PropertySalePublishCost, PropertySalePublishCost); due != 0 {
		t.Fatalf("expected fully prepaid publish to be free, got %d", due)
	}
	if due := propertyPublishPointsDue(PropertySalePublishCost, 1400); due != 0 {
		t.Fatalf("expected legacy overpayment to avoid a second charge, got %d", due)
	}
}

// 8. ptrInt returns an int pointer for optional area fields.
func ptrInt(value int) *int {
	return &value
}

// 9. TestDeriveServicedApartmentRentRange verifies project-level room rent range fields.
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

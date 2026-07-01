/*
 * 測試列表資料種子命令。
 * 1. 連接資料庫並確保基礎帳戶與屋苑資料存在。
 * 2. 幂等建立公開可見的樓盤租售、服務式住宅與家具帖子。
 * 3. 供本地與測試環境快速補齊可預覽內容。
 */
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"

	"ajoliving_web/http_service/internal/config"
	"ajoliving_web/http_service/internal/database"
	"ajoliving_web/http_service/internal/model"
	"ajoliving_web/http_service/internal/utils"
)

type sampleUserInput struct {
	phoneNumber           string
	displayName           string
	publisherIdentityType string
	districtCode          string
	primaryCommunityID    *int64
}

type sampleListingSeedResult struct {
	module string
	title  string
	id     string
}

type sampleImageInput struct {
	objectKey string
	mimeType  string
	sortOrder int
	isCover   bool
}

// 1. main seeds sample public listings once.
func main() {
	cfg := config.Load()

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
	if err := database.SeedCommunities(context.Background(), db); err != nil {
		log.Fatal(err)
	}

	results, err := seedSampleListings(context.Background(), db, cfg)
	if err != nil {
		log.Fatal(err)
	}

	for _, item := range results {
		log.Printf("[%s] %s -> %s", item.module, item.title, item.id)
	}
}

// 2. seedSampleListings inserts or updates sample listings.
func seedSampleListings(ctx context.Context, db *gorm.DB, cfg *config.Config) ([]sampleListingSeedResult, error) {
	results := make([]sampleListingSeedResult, 0, 8)
	now := time.Now()
	expireAt := now.AddDate(0, 3, 0)
	adExpireAt := now.AddDate(0, 1, 0)

	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		communities, err := loadSeedCommunities(tx)
		if err != nil {
			return err
		}
		lagunaCityID := communities["麗港城"].ID
		kornhillID := communities["康怡花園"].ID
		taikooVenusID := communities["太古城金星閣"].ID

		ownerUserID, err := ensureSampleUser(tx, sampleUserInput{
			phoneNumber:           "90000011",
			displayName:           "陳先生",
			publisherIdentityType: "owner",
			districtCode:          communities["麗港城"].DistrictCode,
			primaryCommunityID:    &lagunaCityID,
		})
		if err != nil {
			return err
		}

		agentUserID, err := ensureSampleUser(tx, sampleUserInput{
			phoneNumber:           "90000022",
			displayName:           "AJO 代理",
			publisherIdentityType: "agent",
			districtCode:          communities["康怡花園"].DistrictCode,
			primaryCommunityID:    &kornhillID,
		})
		if err != nil {
			return err
		}

		saleOne, err := upsertListing(tx, model.Listing{
			Module:                "property_sale",
			OwnerUserID:           ownerUserID,
			Title:                 "麗港城 2 房高層放售",
			Summary:               "已翻新兩房，景觀開揚，適合自住。",
			Description:           "單位已完成基本翻新，客飯廳方正，鄰近港鐵與商場，方便睇樓及比較樓盤篩選效果。",
			DistrictCode:          communities["麗港城"].DistrictCode,
			CommunityID:           &lagunaCityID,
			PublisherIdentityType: "owner",
			PublicationStatus:     "active",
			ModerationStatus:      "approved",
			BusinessStatus:        "available",
			PublishedAt:           &now,
			SortRefreshedAt:       &now,
			ExpireAt:              &expireAt,
			IsDeleted:             false,
		})
		if err != nil {
			return err
		}
		if err := upsertPropertySale(tx, saleOne.ID, model.PropertySaleListing{
			ListingID:            saleOne.ID,
			PropertyNo:           "PS-DEMO-001",
			TransactionType:      "sale",
			LocationScope:        "local",
			ListingCategory:      "standard",
			MultiUnitProject:     false,
			PropertyType:         "residential",
			RentalType:           "",
			EstateName:           "麗港城",
			AddressText:          "九龍藍田麗港街 1 號",
			AddressTextEn:        "1 Laguna Street, Lam Tin, Kowloon",
			BlockName:            "第 12 座",
			UnitName:             "H 室",
			ShowUnit:             true,
			AskingPriceHKD:       8380000,
			MonthlyRentHKD:       0,
			PriceReferenceOnly:   false,
			PriceNegotiable:      false,
			AnnualPrepayDiscount: false,
			AnnualPrepayOption:   "none",
			LeaseStartDate:       "",
			RentIncluded:         "",
			AreaMode:             "usable",
			UsableAreaSqft:       517,
			GrossAreaSqft:        intPtr(668),
			BedroomCount:         2,
			LivingRoomCount:      1,
			BathroomCount:        1,
			FloorLevel:           "高層",
			FloorRaw:             "高層",
			FloorZone:            "high",
			FloorDisplayRange:    "高層",
			TotalFloors:          30,
			PublicLocationText:   "藍田",
			Direction:            "東南",
			BuildingAge:          "約 30 年",
			KitchenType:          "獨立廚房",
			CookingMode:          "明火",
			ManagementFeeHKD:     1680,
			VideoURL:             "",
			VRURL:                "",
			PrivateNote:          "sample-seed",
			TitleEn:              "Laguna City 2-bedroom high floor for sale",
			DescriptionEn:        "Renovated two-bedroom flat with open view.",
			AdPackageCode:        "featured",
			AdWeight:             1,
			AdPriceHKD:           800,
			AdPricePoints:        800,
			AdDurationDays:       30,
			AdExpiresAt:          &adExpireAt,
			FeatureTags:          mustJSON([]string{"renovated", "appliances", "view"}),
			ContactMethod:        "both",
			PublisherRoleLabel:   "業主",
		}); err != nil {
			return err
		}
		if err := upsertListingContact(tx, cfg.EncryptionKey, saleOne.ID, "91234567", "91234567", "owner.demo@ajo.local", true, true, true, true, "both"); err != nil {
			return err
		}
		if err := replaceListingImages(tx, cfg, saleOne.ID, ownerUserID, []sampleImageInput{
			{objectKey: "images/pexels-jimmy-teoh-294331-35774007.jpg", mimeType: "image/jpeg", sortOrder: 1, isCover: true},
			{objectKey: "home-stage/property-sale.webp", mimeType: "image/webp", sortOrder: 2, isCover: false},
		}); err != nil {
			return err
		}
		results = append(results, sampleListingSeedResult{module: "property_sale", title: saleOne.Title, id: saleOne.PublicID})

		saleTwo, err := upsertListing(tx, model.Listing{
			Module:                "property_sale",
			OwnerUserID:           agentUserID,
			Title:                 "康怡花園 3 房放租",
			Summary:               "三房連傢俬，可即住，適合家庭租客。",
			Description:           "代理示範盤，保留真實租盤字段，方便測試租售切換、租盤類別與價格範圍篩選。",
			DistrictCode:          communities["康怡花園"].DistrictCode,
			CommunityID:           &kornhillID,
			PublisherIdentityType: "agent",
			PublicationStatus:     "active",
			ModerationStatus:      "approved",
			BusinessStatus:        "available",
			PublishedAt:           &now,
			SortRefreshedAt:       &now,
			ExpireAt:              &expireAt,
			IsDeleted:             false,
		})
		if err != nil {
			return err
		}
		if err := upsertPropertySale(tx, saleTwo.ID, model.PropertySaleListing{
			ListingID:            saleTwo.ID,
			PropertyNo:           "PS-DEMO-002",
			TransactionType:      "rent",
			LocationScope:        "local",
			ListingCategory:      "standard",
			MultiUnitProject:     false,
			PropertyType:         "residential",
			RentalType:           "short_term",
			EstateName:           "康怡花園",
			AddressText:          "港島鰂魚涌康山道 2 號",
			AddressTextEn:        "2 Kornhill Road, Quarry Bay, Hong Kong",
			BlockName:            "M 座",
			UnitName:             "09 室",
			ShowUnit:             true,
			AskingPriceHKD:       0,
			MonthlyRentHKD:       32800,
			PriceReferenceOnly:   false,
			PriceNegotiable:      false,
			AnnualPrepayDiscount: true,
			AnnualPrepayOption:   "95",
			LeaseStartDate:       "2026-07-01",
			RentIncluded:         "管理費",
			AreaMode:             "usable",
			UsableAreaSqft:       688,
			GrossAreaSqft:        intPtr(802),
			BedroomCount:         3,
			LivingRoomCount:      1,
			BathroomCount:        2,
			FloorLevel:           "中高層",
			FloorRaw:             "中高層",
			FloorZone:            "middle",
			FloorDisplayRange:    "中高層",
			TotalFloors:          32,
			PublicLocationText:   "鰂魚涌",
			Direction:            "東",
			BuildingAge:          "約 36 年",
			KitchenType:          "獨立廚房",
			CookingMode:          "明火",
			ManagementFeeHKD:     2100,
			VideoURL:             "",
			VRURL:                "",
			PrivateNote:          "sample-seed",
			TitleEn:              "Kornhill 3-bedroom for rent",
			DescriptionEn:        "Furnished three-bedroom rental listing for demo use.",
			AdPackageCode:        "premium",
			AdWeight:             2,
			AdPriceHKD:           1500,
			AdPricePoints:        1500,
			AdDurationDays:       30,
			AdExpiresAt:          &adExpireAt,
			FeatureTags:          mustJSON([]string{"brand_new", "furnished", "pet_friendly"}),
			ContactMethod:        "chat_or_whatsapp",
			PublisherRoleLabel:   "代理",
		}); err != nil {
			return err
		}
		if err := upsertListingContact(tx, cfg.EncryptionKey, saleTwo.ID, "92345678", "92345678", "agent.demo@ajo.local", true, true, true, true, "chat_or_whatsapp"); err != nil {
			return err
		}
		if err := replaceListingImages(tx, cfg, saleTwo.ID, agentUserID, []sampleImageInput{
			{objectKey: "images/pexels-kseniya-kobi-3624194-7820979.jpg", mimeType: "image/jpeg", sortOrder: 1, isCover: true},
			{objectKey: "home-stage/gradient-slider/img03.webp", mimeType: "image/webp", sortOrder: 2, isCover: false},
		}); err != nil {
			return err
		}
		results = append(results, sampleListingSeedResult{module: "property_sale", title: saleTwo.Title, id: saleTwo.PublicID})

		saleThree, err := upsertListing(tx, model.Listing{
			Module:                "property_sale",
			OwnerUserID:           agentUserID,
			Title:                 "太古城海景 4 房放售",
			Summary:               "高層四房，附連車位，適合換樓家庭。",
			Description:           "示範大型放售單位，用於測試高價、四房以上、連車位和大面積篩選。",
			DistrictCode:          communities["太古城金星閣"].DistrictCode,
			CommunityID:           &taikooVenusID,
			PublisherIdentityType: "agent",
			PublicationStatus:     "active",
			ModerationStatus:      "approved",
			BusinessStatus:        "available",
			PublishedAt:           &now,
			SortRefreshedAt:       &now,
			ExpireAt:              &expireAt,
			IsDeleted:             false,
		})
		if err != nil {
			return err
		}
		if err := upsertPropertySale(tx, saleThree.ID, model.PropertySaleListing{
			ListingID:          saleThree.ID,
			PropertyNo:         "PS-DEMO-003",
			TransactionType:    "sale",
			LocationScope:      "local",
			ListingCategory:    "standard",
			MultiUnitProject:   false,
			PropertyType:       "residential",
			EstateName:         "太古城",
			AddressText:        "港島太古城道 18 號",
			AddressTextEn:      "18 Taikoo Shing Road, Hong Kong",
			BlockName:          "金星閣",
			UnitName:           "31A",
			ShowUnit:           true,
			AskingPriceHKD:     22800000,
			MonthlyRentHKD:     0,
			AreaMode:           "usable",
			UsableAreaSqft:     1120,
			GrossAreaSqft:      intPtr(1350),
			BedroomCount:       4,
			LivingRoomCount:    2,
			BathroomCount:      2,
			FloorLevel:         "高層",
			FloorRaw:           "高層",
			FloorZone:          "high",
			FloorDisplayRange:  "高層",
			TotalFloors:        35,
			PublicLocationText: "太古",
			Direction:          "南",
			BuildingAge:        "約 32 年",
			KitchenType:        "獨立廚房",
			CookingMode:        "明火",
			ManagementFeeHKD:   3200,
			TitleEn:            "Taikoo Shing sea view 4-bedroom for sale",
			DescriptionEn:      "Large family sale listing with parking included.",
			AdPackageCode:      "premium",
			AdWeight:           2,
			AdPriceHKD:         1500,
			AdPricePoints:      1500,
			AdDurationDays:     30,
			AdExpiresAt:        &adExpireAt,
			FeatureTags:        mustJSON([]string{"view", "parking_included", "exclusive"}),
			ContactMethod:      "both",
			PublisherRoleLabel: "代理",
		}); err != nil {
			return err
		}
		if err := upsertListingContact(tx, cfg.EncryptionKey, saleThree.ID, "95670001", "95670001", "sales03@ajo.local", true, true, true, true, "both"); err != nil {
			return err
		}
		if err := replaceListingImages(tx, cfg, saleThree.ID, agentUserID, []sampleImageInput{
			{objectKey: "home-stage/gradient-slider/img07.webp", mimeType: "image/webp", sortOrder: 1, isCover: true},
			{objectKey: "home-stage/gradient-slider/img08.webp", mimeType: "image/webp", sortOrder: 2, isCover: false},
		}); err != nil {
			return err
		}
		results = append(results, sampleListingSeedResult{module: "property_sale", title: saleThree.Title, id: saleThree.PublicID})

		saleFour, err := upsertListing(tx, model.Listing{
			Module:                "property_sale",
			OwnerUserID:           ownerUserID,
			Title:                 "藍田車位放售",
			Summary:               "單邊易泊車位，適合區內住戶。",
			Description:           "示範非住宅類型，方便測試車位類型篩選和較低總價篩選。",
			DistrictCode:          communities["麗港城"].DistrictCode,
			CommunityID:           &lagunaCityID,
			PublisherIdentityType: "owner",
			PublicationStatus:     "active",
			ModerationStatus:      "approved",
			BusinessStatus:        "available",
			PublishedAt:           &now,
			SortRefreshedAt:       &now,
			ExpireAt:              &expireAt,
			IsDeleted:             false,
		})
		if err != nil {
			return err
		}
		if err := upsertPropertySale(tx, saleFour.ID, model.PropertySaleListing{
			ListingID:          saleFour.ID,
			PropertyNo:         "PS-DEMO-004",
			TransactionType:    "sale",
			LocationScope:      "local",
			ListingCategory:    "standard",
			MultiUnitProject:   false,
			PropertyType:       "car_park",
			EstateName:         "麗港城停車場",
			AddressText:        "九龍藍田麗港街停車場",
			AddressTextEn:      "Laguna City Car Park, Lam Tin, Kowloon",
			BlockName:          "P2",
			UnitName:           "B128",
			ShowUnit:           true,
			AskingPriceHKD:     1680000,
			MonthlyRentHKD:     0,
			AreaMode:           "gross",
			UsableAreaSqft:     0,
			GrossAreaSqft:      intPtr(160),
			BedroomCount:       0,
			LivingRoomCount:    0,
			BathroomCount:      0,
			FloorLevel:         "B2",
			FloorRaw:           "B2",
			FloorZone:          "low",
			FloorDisplayRange:  "B2",
			TotalFloors:        3,
			PublicLocationText: "藍田",
			Direction:          "",
			BuildingAge:        "約 30 年",
			KitchenType:        "",
			CookingMode:        "",
			ManagementFeeHKD:   280,
			TitleEn:            "Lam Tin car park for sale",
			DescriptionEn:      "Demo car park sale listing.",
			AdPackageCode:      "basic",
			AdWeight:           0,
			AdPriceHKD:         600,
			AdPricePoints:      600,
			AdDurationDays:     30,
			AdExpiresAt:        &adExpireAt,
			FeatureTags:        mustJSON([]string{"special_unit"}),
			ContactMethod:      "phone",
			PublisherRoleLabel: "業主",
		}); err != nil {
			return err
		}
		if err := upsertListingContact(tx, cfg.EncryptionKey, saleFour.ID, "95670002", "95670002", "parking@ajo.local", true, true, false, false, "phone"); err != nil {
			return err
		}
		if err := replaceListingImages(tx, cfg, saleFour.ID, ownerUserID, []sampleImageInput{
			{objectKey: "home-stage/gradient-slider/img01.webp", mimeType: "image/webp", sortOrder: 1, isCover: true},
		}); err != nil {
			return err
		}
		results = append(results, sampleListingSeedResult{module: "property_sale", title: saleFour.Title, id: saleFour.PublicID})

		servicedOne, err := upsertListing(tx, model.Listing{
			Module:                "serviced_apartment",
			OwnerUserID:           agentUserID,
			Title:                 "太古城服務式住宅示範盤",
			Summary:               "月租及日租均可查閱，適合短住及企業住宿。",
			Description:           "服務式住宅示範資料，方便前端測試設施標籤、面積篩選與房型價格顯示。",
			DistrictCode:          communities["太古城金星閣"].DistrictCode,
			CommunityID:           &taikooVenusID,
			PublisherIdentityType: "agent",
			PublicationStatus:     "active",
			ModerationStatus:      "approved",
			BusinessStatus:        "available",
			PublishedAt:           &now,
			SortRefreshedAt:       &now,
			ExpireAt:              &expireAt,
			IsDeleted:             false,
		})
		if err != nil {
			return err
		}
		if err := upsertServicedApartment(tx, servicedOne.ID, model.ServicedApartmentProject{
			ListingID:            servicedOne.ID,
			ProjectName:          "太古城服務式住宅",
			ProjectNameEn:        "Taikoo Serviced Residences",
			AddressText:          "港島太古城道 18 號",
			AddressTextEn:        "18 Taikoo Shing Road, Hong Kong",
			WebsiteURL:           "https://example.com/serviced-demo",
			WhatsApp:             "95556677",
			Fax:                  "28881234",
			DescriptionEn:        "Serviced apartment demo for public listing preview.",
			ServiceIntro:         "提供短住、中短租及企業住宿安排。",
			BenefitsText:         "每週房務、Wi-Fi、水電煤。",
			ExtraChargesText:     "額外清潔及加床按需要收費。",
			LowestMonthlyRentHKD: 19800,
			LowestDailyRentHKD:   980,
			PriceReferenceOnly:   false,
			PriceNegotiable:      false,
			MinUsableAreaSqft:    320,
			MinLeaseMonths:       1,
			MinStayValue:         3,
			MinStayUnit:          "day",
			LocationScope:        "local",
			ListingCategory:      "standard",
			MultiUnitProject:     true,
			FacilityTags:         mustJSON([]string{"gym", "laundry", "front_desk_24h", "private_kitchen"}),
			ServiceTags:          mustJSON([]string{"housekeeping", "wifi", "utilities"}),
			RoomTypes:            mustJSON([]map[string]any{{"name": "Studio", "room_category": "studio", "usable_area_sqft": 320, "monthly_rent_min_hkd": 19800, "monthly_rent_max_hkd": 22800, "daily_rent_min_hkd": 980, "daily_rent_max_hkd": 1180, "monthly_rent_hkd": 19800, "included_fees": true, "included_fee_items": []string{"wifi", "utilities"}, "min_lease_months": 1, "min_stay_value": 3, "min_stay_unit": "day", "feature_tags": []string{"wifi", "utilities"}}, {"name": "One Bedroom", "room_category": "one_bedroom", "usable_area_sqft": 480, "monthly_rent_min_hkd": 25800, "monthly_rent_max_hkd": 29800, "daily_rent_min_hkd": 1280, "daily_rent_max_hkd": 1480, "monthly_rent_hkd": 25800, "included_fees": true, "included_fee_items": []string{"wifi", "utilities", "housekeeping"}, "min_lease_months": 1, "min_stay_value": 3, "min_stay_unit": "day", "feature_tags": []string{"housekeeping", "wifi"}}}),
			AdPackageCode:        "featured",
			AdWeight:             1,
			AdPriceHKD:           800,
			AdPricePoints:        1000,
			AdDurationDays:       30,
			AdExpiresAt:          &adExpireAt,
			ContactMethod:        "whatsapp",
			PublisherRoleLabel:   "代理",
		}); err != nil {
			return err
		}
		if err := upsertListingContact(tx, cfg.EncryptionKey, servicedOne.ID, "95556677", "95556677", "stay.demo@ajo.local", true, true, true, true, "whatsapp"); err != nil {
			return err
		}
		if err := replaceListingImages(tx, cfg, servicedOne.ID, agentUserID, []sampleImageInput{
			{objectKey: "home-stage/serviced-apartment.webp", mimeType: "image/webp", sortOrder: 1, isCover: true},
			{objectKey: "home-stage/gradient-slider/img10.webp", mimeType: "image/webp", sortOrder: 2, isCover: false},
		}); err != nil {
			return err
		}
		results = append(results, sampleListingSeedResult{module: "serviced_apartment", title: servicedOne.Title, id: servicedOne.PublicID})

		servicedTwo, err := upsertListing(tx, model.Listing{
			Module:                "serviced_apartment",
			OwnerUserID:           ownerUserID,
			Title:                 "銅鑼灣短住服務住宅",
			Summary:               "支援日租，鄰近商業區，方便商務住宿。",
			Description:           "測試日租價、較小面積和短住條件。",
			DistrictCode:          "hong_kong_island",
			PublisherIdentityType: "owner",
			PublicationStatus:     "active",
			ModerationStatus:      "approved",
			BusinessStatus:        "available",
			PublishedAt:           &now,
			SortRefreshedAt:       &now,
			ExpireAt:              &expireAt,
			IsDeleted:             false,
		})
		if err != nil {
			return err
		}
		if err := upsertServicedApartment(tx, servicedTwo.ID, model.ServicedApartmentProject{
			ListingID:            servicedTwo.ID,
			ProjectName:          "銅鑼灣短住服務住宅",
			ProjectNameEn:        "Causeway Bay Short Stay Residence",
			AddressText:          "港島銅鑼灣告士打道",
			AddressTextEn:        "Gloucester Road, Causeway Bay, Hong Kong",
			WebsiteURL:           "https://example.com/shortstay-demo",
			WhatsApp:             "95556678",
			Fax:                  "28881235",
			DescriptionEn:        "Short stay serviced apartment demo.",
			ServiceIntro:         "適合商務客及短期住宿。",
			BenefitsText:         "包 Wi-Fi 及每週房務。",
			ExtraChargesText:     "按金及附加清潔另計。",
			LowestMonthlyRentHKD: 16800,
			LowestDailyRentHKD:   780,
			MinUsableAreaSqft:    220,
			MinLeaseMonths:       1,
			MinStayValue:         2,
			MinStayUnit:          "day",
			LocationScope:        "local",
			ListingCategory:      "standard",
			MultiUnitProject:     true,
			FacilityTags:         mustJSON([]string{"front_desk_24h", "broadband"}),
			ServiceTags:          mustJSON([]string{"wifi", "front_desk"}),
			RoomTypes:            mustJSON([]map[string]any{{"name": "Compact Studio", "room_category": "studio", "usable_area_sqft": 220, "monthly_rent_min_hkd": 16800, "monthly_rent_max_hkd": 18800, "daily_rent_min_hkd": 780, "daily_rent_max_hkd": 880, "monthly_rent_hkd": 16800, "included_fees": true, "included_fee_items": []string{"wifi"}, "min_lease_months": 1, "min_stay_value": 2, "min_stay_unit": "day", "feature_tags": []string{"wifi"}}}),
			AdPackageCode:        "basic",
			AdWeight:             0,
			AdPriceHKD:           600,
			AdPricePoints:        800,
			AdDurationDays:       30,
			AdExpiresAt:          &adExpireAt,
			ContactMethod:        "whatsapp",
			PublisherRoleLabel:   "業主",
		}); err != nil {
			return err
		}
		if err := upsertListingContact(tx, cfg.EncryptionKey, servicedTwo.ID, "95556678", "95556678", "shortstay@ajo.local", true, true, true, true, "whatsapp"); err != nil {
			return err
		}
		if err := replaceListingImages(tx, cfg, servicedTwo.ID, ownerUserID, []sampleImageInput{
			{objectKey: "home-stage/gradient-slider/img04.webp", mimeType: "image/webp", sortOrder: 1, isCover: true},
		}); err != nil {
			return err
		}
		results = append(results, sampleListingSeedResult{module: "serviced_apartment", title: servicedTwo.Title, id: servicedTwo.PublicID})

		servicedThree, err := upsertListing(tx, model.Listing{
			Module:                "serviced_apartment",
			OwnerUserID:           agentUserID,
			Title:                 "尖沙咀長住服務住宅",
			Summary:               "長住月租方案，適合外派及搬屋過渡。",
			Description:           "測試較高月租和大面積服務住宅排序。",
			DistrictCode:          "kowloon",
			PublisherIdentityType: "agent",
			PublicationStatus:     "active",
			ModerationStatus:      "approved",
			BusinessStatus:        "available",
			PublishedAt:           &now,
			SortRefreshedAt:       &now,
			ExpireAt:              &expireAt,
			IsDeleted:             false,
		})
		if err != nil {
			return err
		}
		if err := upsertServicedApartment(tx, servicedThree.ID, model.ServicedApartmentProject{
			ListingID:            servicedThree.ID,
			ProjectName:          "尖沙咀長住服務住宅",
			ProjectNameEn:        "Tsim Sha Tsui Long Stay Residence",
			AddressText:          "九龍尖沙咀梳士巴利道",
			AddressTextEn:        "Salisbury Road, Tsim Sha Tsui, Kowloon",
			WebsiteURL:           "https://example.com/longstay-demo",
			WhatsApp:             "95556679",
			Fax:                  "28881236",
			DescriptionEn:        "Long stay serviced apartment demo.",
			ServiceIntro:         "支援長住及企業住宿。",
			BenefitsText:         "包房務、健身室及會所設施。",
			ExtraChargesText:     "泊車及額外床品另收費。",
			LowestMonthlyRentHKD: 28800,
			LowestDailyRentHKD:   0,
			MinUsableAreaSqft:    520,
			MinLeaseMonths:       3,
			MinStayValue:         1,
			MinStayUnit:          "month",
			LocationScope:        "local",
			ListingCategory:      "standard",
			MultiUnitProject:     true,
			FacilityTags:         mustJSON([]string{"gym", "parking", "restaurant"}),
			ServiceTags:          mustJSON([]string{"housekeeping", "utilities", "linen"}),
			RoomTypes:            mustJSON([]map[string]any{{"name": "One Bedroom Suite", "room_category": "one_bedroom", "usable_area_sqft": 520, "monthly_rent_min_hkd": 28800, "monthly_rent_max_hkd": 32800, "daily_rent_min_hkd": 0, "daily_rent_max_hkd": 0, "monthly_rent_hkd": 28800, "included_fees": true, "included_fee_items": []string{"utilities", "housekeeping"}, "min_lease_months": 3, "min_stay_value": 1, "min_stay_unit": "month", "feature_tags": []string{"housekeeping"}}}),
			AdPackageCode:        "featured",
			AdWeight:             1,
			AdPriceHKD:           800,
			AdPricePoints:        1000,
			AdDurationDays:       30,
			AdExpiresAt:          &adExpireAt,
			ContactMethod:        "chat_or_whatsapp",
			PublisherRoleLabel:   "代理",
		}); err != nil {
			return err
		}
		if err := upsertListingContact(tx, cfg.EncryptionKey, servicedThree.ID, "95556679", "95556679", "longstay@ajo.local", true, true, true, true, "chat_or_whatsapp"); err != nil {
			return err
		}
		if err := replaceListingImages(tx, cfg, servicedThree.ID, agentUserID, []sampleImageInput{
			{objectKey: "home-stage/gradient-slider/img05.webp", mimeType: "image/webp", sortOrder: 1, isCover: true},
		}); err != nil {
			return err
		}
		results = append(results, sampleListingSeedResult{module: "serviced_apartment", title: servicedThree.Title, id: servicedThree.PublicID})

		servicedFour, err := upsertListing(tx, model.Listing{
			Module:                "serviced_apartment",
			OwnerUserID:           ownerUserID,
			Title:                 "將軍澳家庭式服務住宅",
			Summary:               "兩房房型，包廚房及基本家電。",
			Description:           "測試家庭房型、設施標籤與較大面積。",
			DistrictCode:          "new_territories",
			PublisherIdentityType: "owner",
			PublicationStatus:     "active",
			ModerationStatus:      "approved",
			BusinessStatus:        "available",
			PublishedAt:           &now,
			SortRefreshedAt:       &now,
			ExpireAt:              &expireAt,
			IsDeleted:             false,
		})
		if err != nil {
			return err
		}
		if err := upsertServicedApartment(tx, servicedFour.ID, model.ServicedApartmentProject{
			ListingID:            servicedFour.ID,
			ProjectName:          "將軍澳家庭式服務住宅",
			ProjectNameEn:        "Tseung Kwan O Family Serviced Residence",
			AddressText:          "新界將軍澳唐德街",
			AddressTextEn:        "Tong Tak Street, Tseung Kwan O, New Territories",
			WebsiteURL:           "https://example.com/family-demo",
			WhatsApp:             "95556680",
			Fax:                  "28881237",
			DescriptionEn:        "Family-friendly serviced apartment demo.",
			ServiceIntro:         "適合家庭及中期入住。",
			BenefitsText:         "提供廚房、洗衣及寵物友善房型。",
			ExtraChargesText:     "寵物入住另需清潔費。",
			LowestMonthlyRentHKD: 23800,
			LowestDailyRentHKD:   0,
			MinUsableAreaSqft:    610,
			MinLeaseMonths:       2,
			MinStayValue:         2,
			MinStayUnit:          "month",
			LocationScope:        "local",
			ListingCategory:      "standard",
			MultiUnitProject:     true,
			FacilityTags:         mustJSON([]string{"private_kitchen", "laundry", "pet_friendly"}),
			ServiceTags:          mustJSON([]string{"wifi", "utilities", "housekeeping"}),
			RoomTypes:            mustJSON([]map[string]any{{"name": "Two Bedroom", "room_category": "two_bedroom", "usable_area_sqft": 610, "monthly_rent_min_hkd": 23800, "monthly_rent_max_hkd": 27800, "daily_rent_min_hkd": 0, "daily_rent_max_hkd": 0, "monthly_rent_hkd": 23800, "included_fees": true, "included_fee_items": []string{"wifi", "utilities"}, "min_lease_months": 2, "min_stay_value": 2, "min_stay_unit": "month", "feature_tags": []string{"utilities", "wifi"}}}),
			AdPackageCode:        "basic",
			AdWeight:             0,
			AdPriceHKD:           600,
			AdPricePoints:        800,
			AdDurationDays:       30,
			AdExpiresAt:          &adExpireAt,
			ContactMethod:        "both",
			PublisherRoleLabel:   "業主",
		}); err != nil {
			return err
		}
		if err := upsertListingContact(tx, cfg.EncryptionKey, servicedFour.ID, "95556680", "95556680", "family@ajo.local", true, true, true, true, "both"); err != nil {
			return err
		}
		if err := replaceListingImages(tx, cfg, servicedFour.ID, ownerUserID, []sampleImageInput{
			{objectKey: "home-stage/gradient-slider/img06.webp", mimeType: "image/webp", sortOrder: 1, isCover: true},
		}); err != nil {
			return err
		}
		results = append(results, sampleListingSeedResult{module: "serviced_apartment", title: servicedFour.Title, id: servicedFour.PublicID})

		furnitureOne, err := upsertListing(tx, model.Listing{
			Module:                "secondhand",
			OwnerUserID:           ownerUserID,
			Title:                 "北歐雙座位梳化",
			Summary:               "米白色雙座位梳化，整體乾淨，適合細客廳。",
			Description:           "方便前端測試家具列表、詳情和聯絡流程。可自取，亦可安排市區交收。",
			DistrictCode:          communities["麗港城"].DistrictCode,
			CommunityID:           &lagunaCityID,
			PublisherIdentityType: "owner",
			PublicationStatus:     "active",
			ModerationStatus:      "approved",
			BusinessStatus:        "available",
			PublishedAt:           &now,
			SortRefreshedAt:       &now,
			ExpireAt:              &expireAt,
			IsDeleted:             false,
		})
		if err != nil {
			return err
		}
		if err := upsertSecondhand(tx, furnitureOne.ID, model.SecondhandListing{
			ListingID:          furnitureOne.ID,
			CategoryCode:       "home_furniture",
			PriceMode:          "fixed",
			PriceHKD:           floatPtr(1200),
			ConditionLevel:     "used_good",
			DimensionText:      "約 150cm x 82cm",
			PickupRegionCode:   communities["麗港城"].DistrictCode,
			PickupLocationText: "藍田麗港城",
			DeliveryTags:       mustJSON([]string{"self_pickup", "local_delivery"}),
			VisibilityScope:    "public",
			VisibleCommunityID: nil,
			ContactMethod:      "both",
			IsFreeGiveaway:     false,
		}); err != nil {
			return err
		}
		if err := upsertListingContact(tx, cfg.EncryptionKey, furnitureOne.ID, "93456789", "93456789", "", true, true, true, false, "both"); err != nil {
			return err
		}
		if err := replaceListingImages(tx, cfg, furnitureOne.ID, ownerUserID, []sampleImageInput{
			{objectKey: "images/pexels-steppewalker-36596073.jpg", mimeType: "image/jpeg", sortOrder: 1, isCover: true},
			{objectKey: "home-stage/secondhand.webp", mimeType: "image/webp", sortOrder: 2, isCover: false},
		}); err != nil {
			return err
		}
		results = append(results, sampleListingSeedResult{module: "secondhand", title: furnitureOne.Title, id: furnitureOne.PublicID})

		furnitureTwo, err := upsertListing(tx, model.Listing{
			Module:                "secondhand",
			OwnerUserID:           agentUserID,
			Title:                 "升降辦公椅",
			Summary:               "網布辦公椅，適合 Home Office 測試展示。",
			Description:           "保留固定價格、分類、成色與交收欄位，方便直接查看家具頁面效果。",
			DistrictCode:          communities["康怡花園"].DistrictCode,
			CommunityID:           &kornhillID,
			PublisherIdentityType: "agent",
			PublicationStatus:     "active",
			ModerationStatus:      "approved",
			BusinessStatus:        "available",
			PublishedAt:           &now,
			SortRefreshedAt:       &now,
			ExpireAt:              &expireAt,
			IsDeleted:             false,
		})
		if err != nil {
			return err
		}
		if err := upsertSecondhand(tx, furnitureTwo.ID, model.SecondhandListing{
			ListingID:          furnitureTwo.ID,
			CategoryCode:       "home_furniture",
			PriceMode:          "negotiable",
			PriceHKD:           floatPtr(680),
			ConditionLevel:     "used_excellent",
			DimensionText:      "標準辦公椅尺寸",
			PickupRegionCode:   communities["康怡花園"].DistrictCode,
			PickupLocationText: "鰂魚涌康怡花園",
			DeliveryTags:       mustJSON([]string{"self_pickup"}),
			VisibilityScope:    "public",
			VisibleCommunityID: nil,
			ContactMethod:      "chat_or_whatsapp",
			IsFreeGiveaway:     false,
		}); err != nil {
			return err
		}
		if err := upsertListingContact(tx, cfg.EncryptionKey, furnitureTwo.ID, "94567890", "94567890", "", true, true, true, false, "chat_or_whatsapp"); err != nil {
			return err
		}
		if err := replaceListingImages(tx, cfg, furnitureTwo.ID, agentUserID, []sampleImageInput{
			{objectKey: "images/pexels-jimmy-teoh-294331-35774007.jpg", mimeType: "image/jpeg", sortOrder: 1, isCover: true},
		}); err != nil {
			return err
		}
		results = append(results, sampleListingSeedResult{module: "secondhand", title: furnitureTwo.Title, id: furnitureTwo.PublicID})

		furnitureThree, err := upsertListing(tx, model.Listing{
			Module:                "secondhand",
			OwnerUserID:           ownerUserID,
			Title:                 "餐桌連四椅",
			Summary:               "胡桃木色餐桌組，適合二人至四人家庭。",
			Description:           "測試家具分類與固定價格篩選。",
			DistrictCode:          communities["太古城金星閣"].DistrictCode,
			CommunityID:           &taikooVenusID,
			PublisherIdentityType: "owner",
			PublicationStatus:     "active",
			ModerationStatus:      "approved",
			BusinessStatus:        "available",
			PublishedAt:           &now,
			SortRefreshedAt:       &now,
			ExpireAt:              &expireAt,
			IsDeleted:             false,
		})
		if err != nil {
			return err
		}
		if err := upsertSecondhand(tx, furnitureThree.ID, model.SecondhandListing{
			ListingID:          furnitureThree.ID,
			CategoryCode:       "home_furniture",
			PriceMode:          "fixed",
			PriceHKD:           floatPtr(980),
			ConditionLevel:     "used_good",
			DimensionText:      "餐桌約 120cm x 75cm",
			PickupRegionCode:   communities["太古城金星閣"].DistrictCode,
			PickupLocationText: "太古城",
			DeliveryTags:       mustJSON([]string{"self_pickup"}),
			VisibilityScope:    "public",
			ContactMethod:      "both",
			IsFreeGiveaway:     false,
		}); err != nil {
			return err
		}
		if err := upsertListingContact(tx, cfg.EncryptionKey, furnitureThree.ID, "94567891", "94567891", "", true, true, true, false, "both"); err != nil {
			return err
		}
		if err := replaceListingImages(tx, cfg, furnitureThree.ID, ownerUserID, []sampleImageInput{
			{objectKey: "images/pexels-kseniya-kobi-3624194-7820979.jpg", mimeType: "image/jpeg", sortOrder: 1, isCover: true},
		}); err != nil {
			return err
		}
		results = append(results, sampleListingSeedResult{module: "secondhand", title: furnitureThree.Title, id: furnitureThree.PublicID})

		furnitureFour, err := upsertListing(tx, model.Listing{
			Module:                "secondhand",
			OwnerUserID:           agentUserID,
			Title:                 "小型雪櫃",
			Summary:               "適合單人住戶或辦公室茶水間。",
			Description:           "測試家電分類、可議價和圖片顯示。",
			DistrictCode:          "kowloon",
			PublisherIdentityType: "agent",
			PublicationStatus:     "active",
			ModerationStatus:      "approved",
			BusinessStatus:        "available",
			PublishedAt:           &now,
			SortRefreshedAt:       &now,
			ExpireAt:              &expireAt,
			IsDeleted:             false,
		})
		if err != nil {
			return err
		}
		if err := upsertSecondhand(tx, furnitureFour.ID, model.SecondhandListing{
			ListingID:          furnitureFour.ID,
			CategoryCode:       "home_appliance",
			PriceMode:          "negotiable",
			PriceHKD:           floatPtr(450),
			ConditionLevel:     "used_excellent",
			DimensionText:      "約 50cm x 50cm x 85cm",
			PickupRegionCode:   "kowloon",
			PickupLocationText: "九龍灣",
			DeliveryTags:       mustJSON([]string{"self_pickup", "local_delivery"}),
			VisibilityScope:    "public",
			ContactMethod:      "chat_or_whatsapp",
			IsFreeGiveaway:     false,
		}); err != nil {
			return err
		}
		if err := upsertListingContact(tx, cfg.EncryptionKey, furnitureFour.ID, "94567892", "94567892", "", true, true, true, false, "chat_or_whatsapp"); err != nil {
			return err
		}
		if err := replaceListingImages(tx, cfg, furnitureFour.ID, agentUserID, []sampleImageInput{
			{objectKey: "home-stage/gradient-slider/img09.webp", mimeType: "image/webp", sortOrder: 1, isCover: true},
		}); err != nil {
			return err
		}
		results = append(results, sampleListingSeedResult{module: "secondhand", title: furnitureFour.Title, id: furnitureFour.PublicID})

		return nil
	})
	if err != nil {
		return nil, err
	}

	return results, nil
}

// 3. loadSeedCommunities loads the baseline seed communities.
func loadSeedCommunities(tx *gorm.DB) (map[string]model.Community, error) {
	var communities []model.Community
	if err := tx.Where("name_zh IN ?", []string{"康怡花園", "太古城金星閣", "麗港城"}).Find(&communities).Error; err != nil {
		return nil, err
	}

	result := make(map[string]model.Community, len(communities))
	for _, community := range communities {
		result[community.NameZH] = community
	}

	for _, name := range []string{"康怡花園", "太古城金星閣", "麗港城"} {
		if _, ok := result[name]; !ok {
			return nil, fmt.Errorf("missing seed community %s", name)
		}
	}

	return result, nil
}

// 4. ensureSampleUser creates or updates one sample member and profile.
func ensureSampleUser(tx *gorm.DB, input sampleUserInput) (int64, error) {
	var user model.User
	if err := tx.Where("phone_country_code = ? AND phone_number = ?", "+852", input.phoneNumber).Limit(1).Find(&user).Error; err != nil {
		return 0, err
	}

	if user.ID == 0 {
		user = model.User{
			PublicID:         utils.NewPublicID(),
			PhoneCountryCode: "+852",
			PhoneNumber:      input.phoneNumber,
			MemberStatus:     "active",
			MemberType:       "user",
			IsStaff:          false,
			IsVerifiedPhone:  true,
		}
		if err := tx.Create(&user).Error; err != nil {
			return 0, err
		}
	} else {
		if err := tx.Model(&user).Updates(map[string]any{
			"member_status":     "active",
			"member_type":       "user",
			"is_staff":          false,
			"is_verified_phone": true,
		}).Error; err != nil {
			return 0, err
		}
	}

	var profile model.UserProfile
	if err := tx.Where("user_id = ?", user.ID).Limit(1).Find(&profile).Error; err != nil {
		return 0, err
	}

	assignments := map[string]any{
		"display_name":            input.displayName,
		"publisher_identity_type": input.publisherIdentityType,
		"district_code":           input.districtCode,
		"primary_community_id":    input.primaryCommunityID,
	}

	if profile.UserID == 0 {
		profile = model.UserProfile{
			UserID:                user.ID,
			DisplayName:           input.displayName,
			PublisherIdentityType: input.publisherIdentityType,
			PrimaryCommunityID:    input.primaryCommunityID,
			DistrictCode:          input.districtCode,
		}
		if err := tx.Create(&profile).Error; err != nil {
			return 0, err
		}
	} else {
		if err := tx.Model(&profile).Updates(assignments).Error; err != nil {
			return 0, err
		}
	}

	return user.ID, nil
}

// 5. upsertListing creates or updates one shared listing root row.
func upsertListing(tx *gorm.DB, input model.Listing) (*model.Listing, error) {
	var existing model.Listing
	if err := tx.Where("module = ? AND title = ?", input.Module, input.Title).Limit(1).Find(&existing).Error; err != nil {
		return nil, err
	}

	if existing.ID == 0 {
		input.PublicID = utils.NewPublicID()
		if err := tx.Create(&input).Error; err != nil {
			return nil, err
		}
		return &input, nil
	}

	updates := map[string]any{
		"owner_user_id":           input.OwnerUserID,
		"summary":                 input.Summary,
		"description":             input.Description,
		"district_code":           input.DistrictCode,
		"community_id":            input.CommunityID,
		"publisher_identity_type": input.PublisherIdentityType,
		"publication_status":      input.PublicationStatus,
		"moderation_status":       input.ModerationStatus,
		"business_status":         input.BusinessStatus,
		"published_at":            input.PublishedAt,
		"sort_refreshed_at":       input.SortRefreshedAt,
		"expire_at":               input.ExpireAt,
		"is_deleted":              input.IsDeleted,
	}
	if err := tx.Model(&existing).Updates(updates).Error; err != nil {
		return nil, err
	}
	existing.OwnerUserID = input.OwnerUserID
	existing.Summary = input.Summary
	existing.Description = input.Description
	existing.DistrictCode = input.DistrictCode
	existing.CommunityID = input.CommunityID
	existing.PublisherIdentityType = input.PublisherIdentityType
	existing.PublicationStatus = input.PublicationStatus
	existing.ModerationStatus = input.ModerationStatus
	existing.BusinessStatus = input.BusinessStatus
	existing.PublishedAt = input.PublishedAt
	existing.SortRefreshedAt = input.SortRefreshedAt
	existing.ExpireAt = input.ExpireAt
	existing.IsDeleted = input.IsDeleted

	return &existing, nil
}

// 6. upsertPropertySale creates or updates one property sale extension row.
func upsertPropertySale(tx *gorm.DB, listingID int64, input model.PropertySaleListing) error {
	var existing model.PropertySaleListing
	if err := tx.Where("listing_id = ?", listingID).Limit(1).Find(&existing).Error; err != nil {
		return err
	}

	input.ListingID = listingID
	if existing.ListingID == 0 {
		return tx.Create(&input).Error
	}

	return tx.Model(&existing).Updates(input).Error
}

// 7. upsertServicedApartment creates or updates one serviced apartment extension row.
func upsertServicedApartment(tx *gorm.DB, listingID int64, input model.ServicedApartmentProject) error {
	var existing model.ServicedApartmentProject
	if err := tx.Where("listing_id = ?", listingID).Limit(1).Find(&existing).Error; err != nil {
		return err
	}

	input.ListingID = listingID
	if existing.ListingID == 0 {
		return tx.Create(&input).Error
	}

	return tx.Model(&existing).Updates(input).Error
}

// 8. upsertSecondhand creates or updates one secondhand extension row.
func upsertSecondhand(tx *gorm.DB, listingID int64, input model.SecondhandListing) error {
	var existing model.SecondhandListing
	if err := tx.Where("listing_id = ?", listingID).Limit(1).Find(&existing).Error; err != nil {
		return err
	}

	input.ListingID = listingID
	if existing.ListingID == 0 {
		return tx.Create(&input).Error
	}

	return tx.Model(&existing).Updates(input).Error
}

// 9. upsertListingContact creates or updates one encrypted contact record.
func upsertListingContact(tx *gorm.DB, encryptionKey string, listingID int64, phone string, whatsapp string, email string, showPhone bool, showWhatsApp bool, showChat bool, showInquiry bool, contactMode string) error {
	phoneEncrypted, err := encryptOptional(encryptionKey, phone)
	if err != nil {
		return err
	}
	whatsAppEncrypted, err := encryptOptional(encryptionKey, whatsapp)
	if err != nil {
		return err
	}
	emailEncrypted, err := encryptOptional(encryptionKey, email)
	if err != nil {
		return err
	}

	payload := model.ListingContact{
		ListingID:         listingID,
		PhoneEncrypted:    phoneEncrypted,
		PhoneMasked:       utils.MaskPhone(phone),
		WhatsAppEncrypted: whatsAppEncrypted,
		WhatsAppMasked:    utils.MaskPhone(whatsapp),
		EmailEncrypted:    emailEncrypted,
		ShowPhone:         showPhone,
		ShowWhatsApp:      showWhatsApp,
		ShowChat:          showChat,
		ShowInquiryForm:   showInquiry,
		ContactMode:       contactMode,
	}

	var existing model.ListingContact
	if err := tx.Where("listing_id = ?", listingID).Limit(1).Find(&existing).Error; err != nil {
		return err
	}
	if existing.ListingID == 0 {
		return tx.Create(&payload).Error
	}

	return tx.Model(&existing).Updates(payload).Error
}

// 10. replaceListingImages replaces listing image bindings.
func replaceListingImages(tx *gorm.DB, cfg *config.Config, listingID int64, ownerUserID int64, images []sampleImageInput) error {
	if err := tx.Where("listing_id = ?", listingID).Delete(&model.ListingImage{}).Error; err != nil {
		return err
	}

	rows := make([]model.ListingImage, 0, len(images))
	for _, image := range images {
		assetID, err := ensureMediaAsset(tx, cfg, ownerUserID, image.objectKey, image.mimeType)
		if err != nil {
			return err
		}
		rows = append(rows, model.ListingImage{
			ListingID:    listingID,
			MediaAssetID: assetID,
			SortOrder:    image.sortOrder,
			IsCover:      image.isCover,
		})
	}
	if len(rows) == 0 {
		return nil
	}

	return tx.Create(&rows).Error
}

// 11. ensureMediaAsset creates or reuses one media asset.
func ensureMediaAsset(tx *gorm.DB, cfg *config.Config, ownerUserID int64, objectKey string, mimeType string) (int64, error) {
	var asset model.MediaAsset
	if err := tx.Where("bucket_name = ? AND object_key = ? AND created_by = ?", cfg.StorageBucket, objectKey, ownerUserID).Limit(1).Find(&asset).Error; err != nil {
		return 0, err
	}
	if asset.ID > 0 {
		return asset.ID, nil
	}

	asset = model.MediaAsset{
		PublicID:        utils.NewPublicID(),
		StorageProvider: cfg.StorageProvider,
		BucketName:      cfg.StorageBucket,
		ObjectKey:       objectKey,
		MimeType:        mimeType,
		FileSize:        0,
		CreatedBy:       &ownerUserID,
	}
	if err := tx.Create(&asset).Error; err != nil {
		return 0, err
	}

	return asset.ID, nil
}

// 12. encryptOptional encrypts one optional contact value.
func encryptOptional(secret string, value string) (string, error) {
	if value == "" {
		return "", nil
	}

	return utils.EncryptString(secret, value)
}

// 13. mustJSON marshals one JSON payload.
func mustJSON(value any) datatypes.JSON {
	encoded, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}

	return datatypes.JSON(encoded)
}

// 14. intPtr returns one integer pointer.
func intPtr(value int) *int {
	return &value
}

// 15. floatPtr returns one float pointer.
func floatPtr(value float64) *float64 {
	return &value
}

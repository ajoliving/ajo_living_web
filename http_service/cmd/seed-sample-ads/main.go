/*
 * 測試廣告資料種子命令。
 * 1. 連接資料庫並確保廣告相關資料表存在。
 * 2. 幂等建立公開列表右側短廣告、長廣告與 AJO 錢包積分廣告。
 * 3. 為樓盤、家具與服務式住宅頻道綁定 5 個固定展示位。
 */
package main

import (
	"context"
	"log"
	"time"

	"gorm.io/gorm"

	"ajoliving_web/http_service/internal/config"
	"ajoliving_web/http_service/internal/database"
	"ajoliving_web/http_service/internal/model"
)

const displayPlacementListingSide = "listing_side"

type sampleAdSeed struct {
	publicID         string
	adType           string
	title            string
	summary          string
	mediaURL         string
	targetURL        string
	displayChannel   string
	displayPlacement string
	displayLayout    string
	sortOrder        int
	rewardPoints     int64
	watchSeconds     int
}

type sampleAdSlotSeed struct {
	channel      string
	slotIndex    int
	adPublicID   string
	displayTitle string
	displayText  string
	targetURL    string
}

// 1. main seeds sample ads once.
func main() {
	cfg := config.Load()

	db, err := database.Open(cfg)
	if err != nil {
		log.Fatal(err)
	}
	if err := database.Migrate(db); err != nil {
		log.Fatal(err)
	}

	result, err := seedSampleAds(context.Background(), db, cfg)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("sample ads seeded: reward_ads=%d display_assignments=%d", result.rewardAds, result.displayAssignments)
}

type sampleAdSeedResult struct {
	rewardAds          int
	displayAssignments int
}

// 2. seedSampleAds writes idempotent sample reward and display ads.
func seedSampleAds(ctx context.Context, db *gorm.DB, cfg *config.Config) (sampleAdSeedResult, error) {
	now := time.Now()
	startsAt := now.Add(-time.Hour)
	endsAt := now.AddDate(0, 3, 0)
	result := sampleAdSeedResult{}

	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, seed := range sampleAdSeeds(cfg) {
			if err := upsertSampleAd(tx, seed, startsAt, endsAt); err != nil {
				return err
			}
			result.rewardAds++
		}
		for _, seed := range sampleAdSlotSeeds() {
			if err := upsertSampleAdSlot(tx, seed); err != nil {
				return err
			}
			result.displayAssignments++
		}
		return nil
	})

	return result, err
}

// 3. sampleAdSeeds returns fixed test ad definitions.
func sampleAdSeeds(cfg *config.Config) []sampleAdSeed {
	baseURL := cfg.AppPublicBaseURL
	return []sampleAdSeed{
		{
			publicID:         "ad-prop-s-1",
			adType:           "display_short",
			title:            "樓盤按揭預審",
			summary:          "快速了解可承造樓價與每月供款。",
			mediaURL:         baseURL + "/home-stage/property-sale.webp",
			targetURL:        baseURL + "/properties",
			displayChannel:   "property_sale",
			displayPlacement: displayPlacementListingSide,
			displayLayout:    "image_text",
			sortOrder:        1,
		},
		{
			publicID:         "ad-prop-s-2",
			adType:           "display_short",
			title:            "物業估價服務",
			summary:          "為放售單位提供參考估值。",
			mediaURL:         baseURL + "/home-stage/serviced-apartment.webp",
			targetURL:        baseURL + "/properties",
			displayChannel:   "property_sale",
			displayPlacement: displayPlacementListingSide,
			displayLayout:    "image_text",
			sortOrder:        2,
		},
		{
			publicID:         "ad-prop-s-3",
			adType:           "display_short",
			title:            "睇樓預約支援",
			summary:          "管理處與代理可集中處理預約。",
			mediaURL:         baseURL + "/home-stage/gradient-slider/img01.webp",
			targetURL:        baseURL + "/properties",
			displayChannel:   "property_sale",
			displayPlacement: displayPlacementListingSide,
			displayLayout:    "image_text",
			sortOrder:        3,
		},
		{
			publicID:         "ad-prop-l-1",
			adType:           "display_long",
			title:            "精選樓盤曝光",
			summary:          "提升公開列表右側曝光位置。",
			mediaURL:         baseURL + "/home-stage/gradient-slider/img02.webp",
			targetURL:        baseURL + "/properties",
			displayChannel:   "property_sale",
			displayPlacement: displayPlacementListingSide,
			displayLayout:    "image_full",
			sortOrder:        4,
		},
		{
			publicID:         "ad-prop-l-2",
			adType:           "display_long",
			title:            "大廈服務推廣",
			summary:          "展示屋苑配套與服務入口。",
			mediaURL:         baseURL + "/home-stage/gradient-slider/img03.webp",
			targetURL:        baseURL + "/properties",
			displayChannel:   "property_sale",
			displayPlacement: displayPlacementListingSide,
			displayLayout:    "image_full",
			sortOrder:        5,
		},
		{
			publicID:         "ad-furn-s-1",
			adType:           "display_short",
			title:            "社區家具回收",
			summary:          "大件家具回收、清拆與轉售安排。",
			mediaURL:         baseURL + "/home-stage/secondhand.webp",
			targetURL:        baseURL + "/furniture",
			displayChannel:   "furniture",
			displayPlacement: displayPlacementListingSide,
			displayLayout:    "image_text",
			sortOrder:        1,
		},
		{
			publicID:         "ad-furn-s-2",
			adType:           "display_short",
			title:            "精選家居用品",
			summary:          "收納、餐桌、燈具與日常設備。",
			mediaURL:         baseURL + "/home-stage/gradient-slider/img04.webp",
			targetURL:        baseURL + "/furniture",
			displayChannel:   "furniture",
			displayPlacement: displayPlacementListingSide,
			displayLayout:    "image_text",
			sortOrder:        2,
		},
		{
			publicID:         "ad-furn-s-3",
			adType:           "display_short",
			title:            "電器保養服務",
			summary:          "洗衣機、雪櫃與冷氣維修預約。",
			mediaURL:         baseURL + "/home-stage/gradient-slider/img06.webp",
			targetURL:        baseURL + "/furniture",
			displayChannel:   "furniture",
			displayPlacement: displayPlacementListingSide,
			displayLayout:    "image_text",
			sortOrder:        3,
		},
		{
			publicID:         "ad-furn-l-1",
			adType:           "display_long",
			title:            "上門安裝服務",
			summary:          "窗簾、層架與燈具安裝。",
			mediaURL:         baseURL + "/home-stage/gradient-slider/img05.webp",
			targetURL:        baseURL + "/furniture",
			displayChannel:   "furniture",
			displayPlacement: displayPlacementListingSide,
			displayLayout:    "image_full",
			sortOrder:        4,
		},
		{
			publicID:         "ad-furn-l-2",
			adType:           "display_long",
			title:            "即日配送安排",
			summary:          "同區交收與預約送貨。",
			mediaURL:         baseURL + "/home-stage/gradient-slider/img06.webp",
			targetURL:        baseURL + "/furniture",
			displayChannel:   "furniture",
			displayPlacement: displayPlacementListingSide,
			displayLayout:    "image_full",
			sortOrder:        5,
		},
		{
			publicID:         "ad-serv-s-1",
			adType:           "display_short",
			title:            "短租入住方案",
			summary:          "服務式住宅月租與靈活入住安排。",
			mediaURL:         baseURL + "/home-stage/serviced-apartment.webp",
			targetURL:        baseURL + "/serviced-residence",
			displayChannel:   "serviced_apartment",
			displayPlacement: displayPlacementListingSide,
			displayLayout:    "image_text",
			sortOrder:        1,
		},
		{
			publicID:         "ad-serv-s-2",
			adType:           "display_short",
			title:            "企業住宿配套",
			summary:          "為員工短期住宿提供集中查詢。",
			mediaURL:         baseURL + "/home-stage/gradient-slider/img07.webp",
			targetURL:        baseURL + "/serviced-residence",
			displayChannel:   "serviced_apartment",
			displayPlacement: displayPlacementListingSide,
			displayLayout:    "image_text",
			sortOrder:        2,
		},
		{
			publicID:         "ad-serv-s-3",
			adType:           "display_short",
			title:            "清潔與維修支援",
			summary:          "入住期間可預約日常服務。",
			mediaURL:         baseURL + "/home-stage/gradient-slider/img08.webp",
			targetURL:        baseURL + "/serviced-residence",
			displayChannel:   "serviced_apartment",
			displayPlacement: displayPlacementListingSide,
			displayLayout:    "image_text",
			sortOrder:        3,
		},
		{
			publicID:         "ad-serv-l-1",
			adType:           "display_long",
			title:            "精選服務式住宅",
			summary:          "展示高曝光長版廣告位。",
			mediaURL:         baseURL + "/home-stage/gradient-slider/img09.webp",
			targetURL:        baseURL + "/serviced-residence",
			displayChannel:   "serviced_apartment",
			displayPlacement: displayPlacementListingSide,
			displayLayout:    "image_full",
			sortOrder:        4,
		},
		{
			publicID:         "ad-serv-l-2",
			adType:           "display_long",
			title:            "生活配套推廣",
			summary:          "配合大廈服務與周邊生活入口。",
			mediaURL:         baseURL + "/home-stage/gradient-slider/img10.webp",
			targetURL:        baseURL + "/serviced-residence",
			displayChannel:   "serviced_apartment",
			displayPlacement: displayPlacementListingSide,
			displayLayout:    "image_full",
			sortOrder:        5,
		},
		{
			publicID:      "ad-reward-1",
			adType:        "reward",
			title:         "瀏覽樓盤服務介紹",
			summary:       "觀看完整介紹後可領取 AJO 積分。",
			mediaURL:      baseURL + "/home-stage/property-sale.webp",
			targetURL:     baseURL + "/properties",
			displayLayout: "image_text",
			sortOrder:     1,
			rewardPoints:  20,
			watchSeconds:  5,
		},
		{
			publicID:      "ad-reward-2",
			adType:        "reward",
			title:         "了解二手家私服務",
			summary:       "完成觀看後可在錢包領取積分。",
			mediaURL:      baseURL + "/home-stage/secondhand.webp",
			targetURL:     baseURL + "/furniture",
			displayLayout: "image_text",
			sortOrder:     2,
			rewardPoints:  15,
			watchSeconds:  5,
		},
		{
			publicID:      "ad-reward-3",
			adType:        "reward",
			title:         "查看服務式住宅介紹",
			summary:       "觀看後可獲得測試積分獎勵。",
			mediaURL:      baseURL + "/home-stage/serviced-apartment.webp",
			targetURL:     baseURL + "/serviced-residence",
			displayLayout: "image_text",
			sortOrder:     3,
			rewardPoints:  25,
			watchSeconds:  8,
		},
	}
}

// 4. sampleAdSlotSeeds returns fixed channel slot bindings.
func sampleAdSlotSeeds() []sampleAdSlotSeed {
	return []sampleAdSlotSeed{
		{channel: "property_sale", slotIndex: 1, adPublicID: "ad-prop-s-1", displayTitle: "樓盤按揭預審", displayText: "快速了解可承造樓價。", targetURL: "/properties"},
		{channel: "property_sale", slotIndex: 2, adPublicID: "ad-prop-s-2", displayTitle: "物業估價服務", displayText: "為放售單位提供參考估值。", targetURL: "/properties"},
		{channel: "property_sale", slotIndex: 3, adPublicID: "ad-prop-s-3", displayTitle: "睇樓預約支援", displayText: "集中處理睇樓預約。", targetURL: "/properties"},
		{channel: "property_sale", slotIndex: 4, adPublicID: "ad-prop-l-1", displayTitle: "精選樓盤曝光", displayText: "提升公開列表右側曝光位置。", targetURL: "/properties"},
		{channel: "property_sale", slotIndex: 5, adPublicID: "ad-prop-l-2", displayTitle: "大廈服務推廣", displayText: "展示屋苑配套與服務入口。", targetURL: "/properties"},
		{channel: "furniture", slotIndex: 1, adPublicID: "ad-furn-s-1", displayTitle: "社區家具回收", displayText: "大件家具回收與轉售安排。", targetURL: "/furniture"},
		{channel: "furniture", slotIndex: 2, adPublicID: "ad-furn-s-2", displayTitle: "精選家居用品", displayText: "收納、餐桌、燈具與日常設備。", targetURL: "/furniture"},
		{channel: "furniture", slotIndex: 3, adPublicID: "ad-furn-s-3", displayTitle: "電器保養服務", displayText: "洗衣機、雪櫃與冷氣維修。", targetURL: "/furniture"},
		{channel: "furniture", slotIndex: 4, adPublicID: "ad-furn-l-1", displayTitle: "上門安裝服務", displayText: "窗簾、層架與燈具安裝。", targetURL: "/furniture"},
		{channel: "furniture", slotIndex: 5, adPublicID: "ad-furn-l-2", displayTitle: "即日配送安排", displayText: "同區交收與預約送貨。", targetURL: "/furniture"},
		{channel: "serviced_apartment", slotIndex: 1, adPublicID: "ad-serv-s-1", displayTitle: "短租入住方案", displayText: "月租與靈活入住安排。", targetURL: "/serviced-residence"},
		{channel: "serviced_apartment", slotIndex: 2, adPublicID: "ad-serv-s-2", displayTitle: "企業住宿配套", displayText: "員工短期住宿集中查詢。", targetURL: "/serviced-residence"},
		{channel: "serviced_apartment", slotIndex: 3, adPublicID: "ad-serv-s-3", displayTitle: "清潔與維修支援", displayText: "入住期間可預約日常服務。", targetURL: "/serviced-residence"},
		{channel: "serviced_apartment", slotIndex: 4, adPublicID: "ad-serv-l-1", displayTitle: "精選服務式住宅", displayText: "展示高曝光長版廣告位。", targetURL: "/serviced-residence"},
		{channel: "serviced_apartment", slotIndex: 5, adPublicID: "ad-serv-l-2", displayTitle: "生活配套推廣", displayText: "配合周邊生活服務入口。", targetURL: "/serviced-residence"},
	}
}

// 5. upsertSampleAd creates or updates one sample ad.
func upsertSampleAd(tx *gorm.DB, seed sampleAdSeed, startsAt time.Time, endsAt time.Time) error {
	var existing model.RewardAd
	if err := tx.Where("public_id = ?", seed.publicID).Limit(1).Find(&existing).Error; err != nil {
		return err
	}

	now := time.Now()
	ad := model.RewardAd{
		PublicID:          seed.publicID,
		AdType:            seed.adType,
		Title:             seed.title,
		Summary:           seed.summary,
		CoverURL:          "",
		MediaURL:          seed.mediaURL,
		MediaType:         "image",
		TargetURL:         seed.targetURL,
		DisplayChannel:    seed.displayChannel,
		DisplayPlacement:  seed.displayPlacement,
		DisplayLayout:     seed.displayLayout,
		SortOrder:         seed.sortOrder,
		RewardPoints:      seed.rewardPoints,
		WatchSeconds:      seed.watchSeconds,
		DailyUserLimit:    1,
		TotalBudget:       0,
		TotalGranted:      0,
		WatchCount:        0,
		TotalWatchSeconds: 0,
		LinkClickCount:    0,
		IsActive:          true,
		StartsAt:          &startsAt,
		EndsAt:            &endsAt,
		TimestampModel:    model.TimestampModel{CreatedAt: now, UpdatedAt: now},
	}
	if ad.AdType != "reward" {
		ad.WatchSeconds = 0
	}
	if ad.AdType == "reward" && ad.WatchSeconds <= 0 {
		ad.WatchSeconds = 5
	}

	if existing.ID == 0 {
		return tx.Create(&ad).Error
	}

	updates := map[string]any{
		"ad_type":           ad.AdType,
		"title":             ad.Title,
		"summary":           ad.Summary,
		"cover_url":         ad.CoverURL,
		"media_url":         ad.MediaURL,
		"media_type":        ad.MediaType,
		"target_url":        ad.TargetURL,
		"display_channel":   ad.DisplayChannel,
		"display_placement": ad.DisplayPlacement,
		"display_layout":    ad.DisplayLayout,
		"sort_order":        ad.SortOrder,
		"reward_points":     ad.RewardPoints,
		"watch_seconds":     ad.WatchSeconds,
		"daily_user_limit":  ad.DailyUserLimit,
		"total_budget":      ad.TotalBudget,
		"is_active":         ad.IsActive,
		"starts_at":         ad.StartsAt,
		"ends_at":           ad.EndsAt,
		"updated_at":        now,
	}
	return tx.Model(&existing).Updates(updates).Error
}

// 6. upsertSampleAdSlot binds one ad to one display slot.
func upsertSampleAdSlot(tx *gorm.DB, seed sampleAdSlotSeed) error {
	var ad model.RewardAd
	if err := tx.Where("public_id = ?", seed.adPublicID).First(&ad).Error; err != nil {
		return err
	}

	var existing model.DisplayAdSlotAssignment
	if err := tx.Where(
		"display_channel = ? AND display_placement = ? AND slot_index = ? AND reward_ad_id = ?",
		seed.channel,
		displayPlacementListingSide,
		seed.slotIndex,
		ad.ID,
	).Limit(1).Find(&existing).Error; err != nil {
		return err
	}

	now := time.Now()
	row := model.DisplayAdSlotAssignment{
		DisplayChannel:   seed.channel,
		DisplayPlacement: displayPlacementListingSide,
		SlotIndex:        seed.slotIndex,
		RewardAdID:       ad.ID,
		SortOrder:        1,
		DisplayTitle:     seed.displayTitle,
		DisplayText:      seed.displayText,
		TargetURL:        seed.targetURL,
		TimestampModel:   model.TimestampModel{CreatedAt: now, UpdatedAt: now},
	}
	if existing.ID == 0 {
		return tx.Create(&row).Error
	}

	return tx.Model(&existing).Updates(map[string]any{
		"sort_order":    row.SortOrder,
		"display_title": row.DisplayTitle,
		"display_text":  row.DisplayText,
		"target_url":    row.TargetURL,
		"updated_at":    now,
	}).Error
}

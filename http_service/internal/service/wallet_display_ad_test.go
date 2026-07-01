/*
 * Public display ad service tests.
 * 1. Verify listing-side display ads are filtered by channel, placement, state, and time range.
 * 2. Verify reward ads stay out of public display ad responses.
 */
package service

import (
	"context"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"ajoliving_web/http_service/internal/config"
	"ajoliving_web/http_service/internal/model"
	"ajoliving_web/http_service/internal/utils"
)

// 1. TestListPublicDisplayAdsFiltersAndOrdersDisplayAds verifies public ad filtering.
func TestListPublicDisplayAdsFiltersAndOrdersDisplayAds(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+utils.NewPublicID()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&model.RewardAd{}, &model.DisplayAdSlotAssignment{}); err != nil {
		t.Fatalf("migrate db: %v", err)
	}

	now := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)
	past := now.Add(-time.Hour)
	future := now.Add(time.Hour)
	expired := now.Add(-time.Minute)
	ads := []model.RewardAd{
		{
			PublicID:         "display-b",
			AdType:           rewardAdTypeDisplayShort,
			Title:            "B",
			MediaURL:         "https://example.com/b.png",
			MediaType:        rewardAdMediaTypeImage,
			DisplayChannel:   displayAdChannelPropertySale,
			DisplayPlacement: displayAdPlacementListingSide,
			DisplayLayout:    displayAdLayoutImageText,
			SortOrder:        2,
			IsActive:         true,
			StartsAt:         &past,
			EndsAt:           &future,
			TimestampModel:   model.TimestampModel{CreatedAt: past, UpdatedAt: past},
		},
		{
			PublicID:         "display-c",
			AdType:           rewardAdTypeDisplayShort,
			Title:            "C",
			MediaURL:         "https://example.com/c.png",
			MediaType:        rewardAdMediaTypeImage,
			DisplayChannel:   displayAdChannelPropertySale,
			DisplayPlacement: displayAdPlacementListingSide,
			DisplayLayout:    displayAdLayoutImageText,
			SortOrder:        3,
			IsActive:         true,
			StartsAt:         &past,
			EndsAt:           &future,
			TimestampModel:   model.TimestampModel{CreatedAt: past, UpdatedAt: past},
		},
		{
			PublicID:         "display-a",
			AdType:           rewardAdTypeDisplayLong,
			Title:            "A",
			MediaURL:         "https://example.com/a.png",
			MediaType:        rewardAdMediaTypeImage,
			DisplayChannel:   displayAdChannelPropertySale,
			DisplayPlacement: displayAdPlacementListingSide,
			DisplayLayout:    displayAdLayoutImageFull,
			SortOrder:        1,
			IsActive:         true,
			StartsAt:         &past,
			EndsAt:           &future,
			TimestampModel:   model.TimestampModel{CreatedAt: future, UpdatedAt: future},
		},
		{
			PublicID:         "display-d",
			AdType:           rewardAdTypeDisplayLong,
			Title:            "D",
			MediaURL:         "https://example.com/d.png",
			MediaType:        rewardAdMediaTypeImage,
			DisplayChannel:   displayAdChannelPropertySale,
			DisplayPlacement: displayAdPlacementListingSide,
			DisplayLayout:    displayAdLayoutImageFull,
			SortOrder:        4,
			IsActive:         true,
			StartsAt:         &past,
			EndsAt:           &future,
			TimestampModel:   model.TimestampModel{CreatedAt: past, UpdatedAt: past},
		},
		{
			PublicID:         "display-other-channel",
			AdType:           rewardAdTypeDisplayShort,
			Title:            "Other",
			MediaURL:         "https://example.com/other.png",
			MediaType:        rewardAdMediaTypeImage,
			DisplayChannel:   displayAdChannelFurniture,
			DisplayPlacement: displayAdPlacementListingSide,
			DisplayLayout:    displayAdLayoutImageText,
			SortOrder:        1,
			IsActive:         true,
			TimestampModel:   model.TimestampModel{CreatedAt: past, UpdatedAt: past},
		},
		{
			PublicID:         "display-hidden",
			AdType:           rewardAdTypeDisplayShort,
			Title:            "Hidden",
			MediaURL:         "https://example.com/hidden.png",
			MediaType:        rewardAdMediaTypeImage,
			DisplayChannel:   displayAdChannelPropertySale,
			DisplayPlacement: "home_banner",
			DisplayLayout:    displayAdLayoutImageText,
			SortOrder:        3,
			IsActive:         false,
			TimestampModel:   model.TimestampModel{CreatedAt: past, UpdatedAt: past},
		},
		{
			PublicID:         "display-expired",
			AdType:           rewardAdTypeDisplayShort,
			Title:            "Expired",
			MediaURL:         "https://example.com/expired.png",
			MediaType:        rewardAdMediaTypeImage,
			DisplayChannel:   displayAdChannelPropertySale,
			DisplayPlacement: displayAdPlacementListingSide,
			DisplayLayout:    displayAdLayoutImageText,
			IsActive:         true,
			EndsAt:           &expired,
			TimestampModel:   model.TimestampModel{CreatedAt: past, UpdatedAt: past},
		},
		{
			PublicID:     "reward",
			AdType:       rewardAdTypeReward,
			Title:        "Reward",
			Summary:      "Reward",
			MediaURL:     "https://example.com/reward.png",
			MediaType:    rewardAdMediaTypeImage,
			RewardPoints: 10,
			WatchSeconds: 5,
			IsActive:     true,
			TimestampModel: model.TimestampModel{
				CreatedAt: past,
				UpdatedAt: past,
			},
		},
	}
	if err := db.Create(&ads).Error; err != nil {
		t.Fatalf("seed ads: %v", err)
	}

	service := NewWalletService(&Runtime{
		Config: &config.Config{},
		DB:     db,
		Now:    func() time.Time { return now },
	})
	items, err := service.ListPublicDisplayAds(context.Background(), PublicDisplayAdFilters{
		Channel:   displayAdChannelPropertySale,
		Placement: displayAdPlacementListingSide,
		Limit:     10,
	})
	if err != nil {
		t.Fatalf("list public display ads: %v", err)
	}
	if len(items) != 4 {
		t.Fatalf("expected 4 items, got %d", len(items))
	}
	if items[0].TaskID != "display-a" || items[0].SlotIndex != 4 || items[1].TaskID != "display-b" || items[1].SlotIndex != 1 {
		t.Fatalf("unexpected order: %#v", items)
	}

	assignments := []model.DisplayAdSlotAssignment{
		{
			DisplayChannel:   displayAdChannelPropertySale,
			DisplayPlacement: displayAdPlacementListingSide,
			SlotIndex:        1,
			RewardAdID:       ads[0].ID,
			SortOrder:        1,
		},
		{
			DisplayChannel:   displayAdChannelPropertySale,
			DisplayPlacement: displayAdPlacementListingSide,
			SlotIndex:        4,
			RewardAdID:       ads[2].ID,
			SortOrder:        1,
		},
		{
			DisplayChannel:   displayAdChannelPropertySale,
			DisplayPlacement: displayAdPlacementListingSide,
			SlotIndex:        5,
			RewardAdID:       ads[3].ID,
			SortOrder:        1,
		},
	}
	if err := db.Create(&assignments).Error; err != nil {
		t.Fatalf("seed display slot assignments: %v", err)
	}

	configuredItems, err := service.ListPublicDisplayAds(context.Background(), PublicDisplayAdFilters{
		Channel:   displayAdChannelPropertySale,
		Placement: displayAdPlacementListingSide,
		Limit:     10,
	})
	if err != nil {
		t.Fatalf("list configured display ads: %v", err)
	}
	if len(configuredItems) != 3 || configuredItems[0].SlotIndex != 1 || configuredItems[1].SlotIndex != 4 || configuredItems[2].SlotIndex != 5 {
		t.Fatalf("unexpected configured slot result: %#v", configuredItems)
	}
}

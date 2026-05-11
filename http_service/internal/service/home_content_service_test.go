/*
 * 首頁內容服務測試。
 * 1. 驗證首頁輪播與三個主模組設定可保存並公開輸出。
 * 2. 驗證首頁圖片必須為 PNG 且位於指定 OSS 目錄。
 */
package service

import (
	"context"
	"strings"
	"testing"

	"ajoliving_web/http_service/internal/model"
	"ajoliving_web/http_service/internal/utils"
)

// 1. TestHomeContentSaveAndPublicOutput verifies saved homepage content output.
func TestHomeContentSaveAndPublicOutput(t *testing.T) {
	runtime := newTestRuntime(t)
	community, _ := mustGetCommunities(t, runtime)
	owner := mustCreateUser(t, runtime, "+852", "61000001", &community.ID)
	service := NewHomeContentService(runtime)
	carouselAssetID := mustCreateHomePNGMediaAsset(t, runtime, owner.ID, "carousel-one.png")
	secondhandAssetID := mustCreateHomePNGMediaAsset(t, runtime, owner.ID, "secondhand.png")
	propertyAssetID := mustCreateHomePNGMediaAsset(t, runtime, owner.ID, "property-sale.png")
	servicedAssetID := mustCreateHomePNGMediaAsset(t, runtime, owner.ID, "serviced-apartment.png")
	loginHeroAssetID := mustCreateLoginHeroPNGMediaAsset(t, runtime, owner.ID, "login-hero.png")

	carousel, err := service.SaveCarouselSettings(context.Background(), owner.ID, []HomeCarouselInput{
		{MediaAssetID: carouselAssetID, SortOrder: 1},
	})
	if err != nil {
		t.Fatalf("save carousel settings: %v", err)
	}
	if len(carousel) != 1 || carousel[0].MediaAssetID != carouselAssetID || !strings.Contains(carousel[0].ObjectKey, homeEngMediaObjectPrefix) {
		t.Fatalf("unexpected carousel payload: %#v", carousel)
	}

	cards, err := service.SaveModuleCardSettings(context.Background(), owner.ID, []HomeModuleCardInput{
		{ModuleCode: "secondhand", MediaAssetID: secondhandAssetID, Title: "Secondhand", Subtitle: "Neighbour market", Body: "Curated items"},
		{ModuleCode: "property_sale", MediaAssetID: propertyAssetID, Title: "Property", Subtitle: "Owner listings", Body: "Homes for sale"},
		{ModuleCode: "serviced_apartment", MediaAssetID: servicedAssetID, Title: "Serviced", Subtitle: "Flexible stays", Body: "Managed residences"},
	})
	if err != nil {
		t.Fatalf("save module card settings: %v", err)
	}
	if len(cards) != 3 || cards[0].ModuleCode != "secondhand" || cards[0].Title != "Secondhand" {
		t.Fatalf("unexpected module cards: %#v", cards)
	}

	publicContent, err := service.GetPublicHomeContent(context.Background())
	if err != nil {
		t.Fatalf("load public content: %v", err)
	}
	if len(publicContent.Carousel) != 1 || len(publicContent.ModuleCards) != 3 {
		t.Fatalf("unexpected public content: %#v", publicContent)
	}

	loginHero, err := service.SaveLoginHeroSettings(context.Background(), owner.ID, []LoginHeroInput{
		{
			MediaAssetID: loginHeroAssetID,
			Author:       "AJO Living",
			Location:     "Hong Kong",
			SortOrder:    1,
		},
	})
	if err != nil {
		t.Fatalf("save login hero settings: %v", err)
	}
	if len(loginHero) != 1 || loginHero[0].MediaAssetID != loginHeroAssetID || loginHero[0].Author != "AJO Living" {
		t.Fatalf("unexpected login hero: %#v", loginHero)
	}
}

// 2. TestHomeContentRejectsInvalidMedia verifies homepage media validation.
func TestHomeContentRejectsInvalidMedia(t *testing.T) {
	runtime := newTestRuntime(t)
	community, _ := mustGetCommunities(t, runtime)
	owner := mustCreateUser(t, runtime, "+852", "61000002", &community.ID)
	service := NewHomeContentService(runtime)
	webpAssetID := mustCreateMediaAsset(t, runtime, owner.ID)

	_, err := service.SaveCarouselSettings(context.Background(), owner.ID, []HomeCarouselInput{
		{MediaAssetID: webpAssetID, SortOrder: 1},
	})
	if err == nil {
		t.Fatalf("expected webp homepage media to be rejected")
	}
}

// 3. mustCreateHomePNGMediaAsset creates a homepage PNG media asset.
func mustCreateHomePNGMediaAsset(t *testing.T, runtime *Runtime, userID int64, fileName string) string {
	t.Helper()

	return mustCreatePNGMediaAssetInPrefix(t, runtime, userID, homeEngMediaObjectPrefix, fileName)
}

// 4. mustCreateLoginHeroPNGMediaAsset creates a login hero PNG media asset.
func mustCreateLoginHeroPNGMediaAsset(t *testing.T, runtime *Runtime, userID int64, fileName string) string {
	t.Helper()

	return mustCreatePNGMediaAssetInPrefix(t, runtime, userID, loginBagMediaObjectPrefix, fileName)
}

// 5. mustCreatePNGMediaAssetInPrefix creates a PNG media asset in a specific directory.
func mustCreatePNGMediaAssetInPrefix(t *testing.T, runtime *Runtime, userID int64, objectPrefix string, fileName string) string {
	t.Helper()

	asset := model.MediaAsset{
		PublicID:        utils.NewPublicID(),
		StorageProvider: runtime.Config.StorageProvider,
		BucketName:      runtime.Config.StorageBucket,
		ObjectKey:       objectPrefix + strings.ToLower(fileName),
		MimeType:        "image/png",
		FileSize:        2048,
		ChecksumSHA256:  "checksum",
		CreatedBy:       &userID,
	}
	if err := runtime.DB.Create(&asset).Error; err != nil {
		t.Fatalf("create homepage media asset: %v", err)
	}

	return asset.PublicID
}

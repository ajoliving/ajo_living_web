/*
 * 首頁與登入圖片預設內容種子。
 * 1. 將前端預設靜態圖片上傳至 OSS。
 * 2. 在資料庫內建立首頁三大圖與登入背景圖配置。
 * 3. 已有配置時保持不覆蓋。
 */
package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/model"
	"ajoliving_web/http_service/internal/utils"
)

// 1. HomeContentSeedOptions defines required filesystem context for default content seed.
type HomeContentSeedOptions struct {
	WebPublicDir   string
	OperatorUserID int64
}

// 2. defaultContentAsset defines one local image to upload and persist.
type defaultContentAsset struct {
	LocalPath string
	ObjectKey string
	MimeType  string
	Width     int
	Height    int
}

// 3. defaultLoginHeroSeed defines login background metadata.
type defaultLoginHeroSeed struct {
	Asset    defaultContentAsset
	Author   string
	Location string
}

// 4. defaultHomeModuleCardSeed defines homepage module card metadata.
type defaultHomeModuleCardSeed struct {
	ModuleCode string
	Asset      defaultContentAsset
	Title      string
	Subtitle   string
	Body       string
}

// 5. SeedDefaultHomeContent uploads and stores default home and login images.
func (s *HomeContentService) SeedDefaultHomeContent(ctx context.Context, options HomeContentSeedOptions) error {
	publicDir := strings.TrimSpace(options.WebPublicDir)
	if publicDir == "" {
		return errcode.New(errcode.CodeValidationError, "web public directory is required")
	}
	operatorUserID := options.OperatorUserID
	if operatorUserID <= 0 {
		operatorUserID = s.runtime.Config.SystemUserID
	}
	if operatorUserID <= 0 {
		operatorUserID = 1
	}

	loginSeeds := defaultLoginHeroSeeds(publicDir)
	if err := s.seedDefaultLoginHeroes(ctx, operatorUserID, loginSeeds); err != nil {
		return err
	}
	if err := s.seedDefaultHomeModuleCards(ctx, operatorUserID, defaultHomeModuleCardSeeds(publicDir)); err != nil {
		return err
	}

	return nil
}

// 6. seedDefaultLoginHeroes stores login background defaults when absent.
func (s *HomeContentService) seedDefaultLoginHeroes(ctx context.Context, operatorUserID int64, seeds []defaultLoginHeroSeed) error {
	items, err := s.ListLoginHeroSettings(ctx)
	if err != nil {
		return err
	}
	if len(items) > 0 {
		return nil
	}

	inputs := make([]LoginHeroInput, 0, len(seeds))
	for index, seed := range seeds {
		asset, err := s.ensureDefaultMediaAsset(ctx, operatorUserID, seed.Asset)
		if err != nil {
			return err
		}
		inputs = append(inputs, LoginHeroInput{
			MediaAssetID: asset.PublicID,
			Author:       seed.Author,
			Location:     seed.Location,
			SortOrder:    index + 1,
		})
	}

	_, err = s.SaveLoginHeroSettings(ctx, operatorUserID, inputs)
	return err
}

// 7. seedDefaultHomeModuleCards stores homepage module card defaults when absent.
func (s *HomeContentService) seedDefaultHomeModuleCards(ctx context.Context, operatorUserID int64, seeds []defaultHomeModuleCardSeed) error {
	items, err := s.ListModuleCardSettings(ctx)
	if err != nil {
		return err
	}
	if len(items) > 0 {
		return nil
	}

	inputs := make([]HomeModuleCardInput, 0, len(seeds))
	for _, seed := range seeds {
		asset, err := s.ensureDefaultMediaAsset(ctx, operatorUserID, seed.Asset)
		if err != nil {
			return err
		}
		inputs = append(inputs, HomeModuleCardInput{
			ModuleCode:   seed.ModuleCode,
			MediaAssetID: asset.PublicID,
			Title:        seed.Title,
			Subtitle:     seed.Subtitle,
			Body:         seed.Body,
		})
	}

	_, err = s.SaveModuleCardSettings(ctx, operatorUserID, inputs)
	return err
}

// 8. ensureDefaultMediaAsset uploads local image to storage and upserts media asset metadata.
func (s *HomeContentService) ensureDefaultMediaAsset(ctx context.Context, operatorUserID int64, seed defaultContentAsset) (*model.MediaAsset, error) {
	body, err := os.ReadFile(seed.LocalPath)
	if err != nil {
		return nil, errcode.New(errcode.CodeInternalError, fmt.Sprintf("failed to read default image %s", seed.LocalPath))
	}

	media, err := s.loadMediaAssetByObjectKey(ctx, seed.ObjectKey)
	if err != nil {
		return nil, err
	}
	if media != nil {
		return media, nil
	}

	info, err := s.runtime.StorageProvider.PutObject(ctx, PutObjectInput{
		ObjectKey: seed.ObjectKey,
		MimeType:  seed.MimeType,
		Body:      body,
	})
	if err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to upload default image")
	}
	if strings.TrimSpace(info.ContentType) != "" && !strings.EqualFold(strings.TrimSpace(info.ContentType), seed.MimeType) {
		return nil, errcode.New(errcode.CodeInternalError, "default image content type mismatch")
	}

	checksum := sha256.Sum256(body)
	width := seed.Width
	height := seed.Height
	media = &model.MediaAsset{
		PublicID:        utils.NewPublicID(),
		StorageProvider: s.runtime.Config.StorageProvider,
		BucketName:      s.runtime.Config.StorageBucket,
		ObjectKey:       seed.ObjectKey,
		MimeType:        seed.MimeType,
		Width:           &width,
		Height:          &height,
		FileSize:        int64(len(body)),
		ChecksumSHA256:  hex.EncodeToString(checksum[:]),
		CreatedBy:       &operatorUserID,
	}
	if err := s.runtime.DB.WithContext(ctx).Create(media).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to save default media asset")
	}

	return media, nil
}

// 9. loadMediaAssetByObjectKey loads one media asset by object key.
func (s *HomeContentService) loadMediaAssetByObjectKey(ctx context.Context, objectKey string) (*model.MediaAsset, error) {
	var media model.MediaAsset
	if err := s.runtime.DB.WithContext(ctx).
		Where("bucket_name = ? AND object_key = ?", s.runtime.Config.StorageBucket, objectKey).
		Limit(1).
		Find(&media).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load default media asset")
	}
	if media.ID == 0 {
		return nil, nil
	}

	return &media, nil
}

// 10. defaultLoginHeroSeeds returns the initial login background set.
func defaultLoginHeroSeeds(publicDir string) []defaultLoginHeroSeed {
	return []defaultLoginHeroSeed{
		{
			Asset: defaultContentAsset{
				LocalPath: filepath.Join(publicDir, "images", "pexels-jimmy-teoh-294331-35774007.jpg"),
				ObjectKey: loginBagMediaObjectPrefix + "pexels-jimmy-teoh-294331-35774007.jpg",
				MimeType:  "image/jpeg",
				Width:     2976,
				Height:    3968,
			},
			Author:   "@jimmy teoh",
			Location: "Victoria Harbour, HK",
		},
		{
			Asset: defaultContentAsset{
				LocalPath: filepath.Join(publicDir, "images", "pexels-kseniya-kobi-3624194-7820979.jpg"),
				ObjectKey: loginBagMediaObjectPrefix + "pexels-kseniya-kobi-3624194-7820979.jpg",
				MimeType:  "image/jpeg",
				Width:     3024,
				Height:    4032,
			},
			Author:   "@kseniya kobi",
			Location: "Hong Kong Residence",
		},
	}
}

// 11. defaultHomeModuleCardSeeds returns the initial homepage module card set.
func defaultHomeModuleCardSeeds(publicDir string) []defaultHomeModuleCardSeed {
	return []defaultHomeModuleCardSeed{
		{
			ModuleCode: "secondhand",
			Asset: defaultContentAsset{
				LocalPath: filepath.Join(publicDir, "home-stage", "secondhand.webp"),
				ObjectKey: homeEngMediaObjectPrefix + "secondhand.webp",
				MimeType:  "image/webp",
				Width:     2816,
				Height:    1536,
			},
			Title:    "二手交易",
			Subtitle: "已接入首頁、列表、篩選與聊天。",
			Body:     "瀏覽屋苑二手帖子，支援篩選、聊天與成交跟進。",
		},
		{
			ModuleCode: "property_sale",
			Asset: defaultContentAsset{
				LocalPath: filepath.Join(publicDir, "home-stage", "property-sale.webp"),
				ObjectKey: homeEngMediaObjectPrefix + "property-sale.webp",
				MimeType:  "image/webp",
				Width:     2816,
				Height:    1536,
			},
			Title:    "樓盤放售",
			Subtitle: "樓盤放售已可瀏覽與發布。",
			Body:     "樓盤放售入口，支援列表、詳情、發布與聯絡方式解鎖。",
		},
		{
			ModuleCode: "serviced_apartment",
			Asset: defaultContentAsset{
				LocalPath: filepath.Join(publicDir, "home-stage", "serviced-apartment.webp"),
				ObjectKey: homeEngMediaObjectPrefix + "serviced-apartment.webp",
				MimeType:  "image/webp",
				Width:     2816,
				Height:    1536,
			},
			Title:    "服務住宅",
			Subtitle: "服務式住宅已可瀏覽與發布。",
			Body:     "短租與月租服務住宅入口，支援列表、詳情、發布與聯絡方式解鎖。",
		},
	}
}

// 14. DetectImageMimeType detects the image MIME type from bytes.
func DetectImageMimeType(body []byte) string {
	contentType := http.DetectContentType(body)
	if contentType == "image/x-png" {
		return "image/png"
	}

	return contentType
}

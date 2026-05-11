/*
 * 首頁內容配置服務。
 * 1. 管理首頁輪播圖片、三個主模組大圖與登入背景圖設定。
 * 2. 驗證圖片媒體資產與指定 OSS 目錄。
 * 3. 輸出公開首頁展示資料。
 */
package service

import (
	"context"
	"sort"
	"strings"

	"gorm.io/gorm"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/model"
)

const (
	// 1. homePlacementTypeCarousel marks homepage carousel image slots.
	homePlacementTypeCarousel = "carousel"
	// 2. homePlacementTypeModuleCard marks homepage module card slots.
	homePlacementTypeModuleCard = "module_card"
	// 3. homePlacementTypeLoginHero marks login page hero image.
	homePlacementTypeLoginHero = "login_hero"
	// 4. homeCarouselSlotLimit defines the maximum carousel image count.
	homeCarouselSlotLimit = 12
	// 5. loginHeroSlotLimit defines the maximum login hero image count.
	loginHeroSlotLimit = 3
)

var allowedHomeModuleCodes = []string{"secondhand", "property_sale", "serviced_apartment"}

// 1. HomeContentService handles homepage configuration.
type HomeContentService struct {
	runtime *Runtime
}

// 2. HomeCarouselInput defines one carousel save item.
type HomeCarouselInput struct {
	MediaAssetID string `json:"media_asset_id"`
	SortOrder    int    `json:"sort_order"`
}

// 3. HomeModuleCardInput defines one module card save item.
type HomeModuleCardInput struct {
	ModuleCode   string `json:"module_code"`
	MediaAssetID string `json:"media_asset_id"`
	Title        string `json:"title"`
	Subtitle     string `json:"subtitle"`
	Body         string `json:"body"`
}

// 4. LoginHeroInput defines login page hero save payload.
type LoginHeroInput struct {
	MediaAssetID string `json:"media_asset_id"`
	Author       string `json:"author"`
	Location     string `json:"location"`
	SortOrder    int    `json:"sort_order"`
}

// 5. HomeImageResponse defines reusable homepage image payload.
type HomeImageResponse struct {
	MediaAssetID string `json:"media_asset_id"`
	URL          string `json:"url"`
	ObjectKey    string `json:"object_key"`
	SortOrder    int    `json:"sort_order"`
}

// 6. HomeModuleCardResponse defines one homepage module card payload.
type HomeModuleCardResponse struct {
	ModuleCode   string `json:"module_code"`
	MediaAssetID string `json:"media_asset_id"`
	URL          string `json:"url"`
	ObjectKey    string `json:"object_key"`
	Title        string `json:"title"`
	Subtitle     string `json:"subtitle"`
	Body         string `json:"body"`
}

// 7. LoginHeroResponse defines login page hero image payload.
type LoginHeroResponse struct {
	MediaAssetID string `json:"media_asset_id"`
	URL          string `json:"url"`
	ObjectKey    string `json:"object_key"`
	Author       string `json:"author"`
	Location     string `json:"location"`
	SortOrder    int    `json:"sort_order"`
}

// 8. HomeContentResponse defines public homepage content.
type HomeContentResponse struct {
	Carousel    []HomeImageResponse      `json:"carousel"`
	ModuleCards []HomeModuleCardResponse `json:"module_cards"`
}

// 9. NewHomeContentService creates a homepage content service instance.
func NewHomeContentService(runtime *Runtime) *HomeContentService {
	return &HomeContentService{runtime: runtime}
}

// 10. GetPublicHomeContent returns the saved public homepage content.
func (s *HomeContentService) GetPublicHomeContent(ctx context.Context) (*HomeContentResponse, error) {
	carousel, err := s.ListCarouselSettings(ctx)
	if err != nil {
		return nil, err
	}
	moduleCards, err := s.ListModuleCardSettings(ctx)
	if err != nil {
		return nil, err
	}

	return &HomeContentResponse{
		Carousel:    carousel,
		ModuleCards: moduleCards,
	}, nil
}

// 11. ListCarouselSettings returns configured homepage carousel images.
func (s *HomeContentService) ListCarouselSettings(ctx context.Context) ([]HomeImageResponse, error) {
	placements, assetMap, err := s.loadHomePlacements(ctx, homePlacementTypeCarousel)
	if err != nil {
		return nil, err
	}

	items := make([]HomeImageResponse, 0, len(placements))
	for _, placement := range placements {
		asset, exists := assetMap[placement.MediaAssetID]
		if !exists {
			continue
		}
		items = append(items, HomeImageResponse{
			MediaAssetID: asset.PublicID,
			URL:          buildMediaURL(s.runtime.Config.MediaBaseURL, asset.ObjectKey),
			ObjectKey:    asset.ObjectKey,
			SortOrder:    placement.SlotIndex,
		})
	}

	return items, nil
}

// 12. SaveCarouselSettings replaces the homepage carousel configuration.
func (s *HomeContentService) SaveCarouselSettings(ctx context.Context, operatorUserID int64, inputs []HomeCarouselInput) ([]HomeImageResponse, error) {
	if len(inputs) > homeCarouselSlotLimit {
		return nil, errcode.New(errcode.CodeValidationError, "too many carousel images")
	}

	normalizedInputs, err := normalizeHomeCarouselInputs(inputs)
	if err != nil {
		return nil, err
	}
	assets, err := s.loadOwnedHomePNGAssets(ctx, operatorUserID, homeCarouselAssetIDs(normalizedInputs))
	if err != nil {
		return nil, err
	}

	err = s.runtime.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("placement_type = ?", homePlacementTypeCarousel).Delete(&model.HomeContentPlacement{}).Error; err != nil {
			return errcode.New(errcode.CodeInternalError, "failed to clear homepage carousel")
		}
		for index, input := range normalizedInputs {
			asset := assets[input.MediaAssetID]
			placement := model.HomeContentPlacement{
				PlacementType: homePlacementTypeCarousel,
				ModuleCode:    "",
				SlotIndex:     index + 1,
				MediaAssetID:  asset.ID,
				CreatedBy:     &operatorUserID,
				UpdatedBy:     &operatorUserID,
			}
			if err := tx.Create(&placement).Error; err != nil {
				return errcode.New(errcode.CodeInternalError, "failed to save homepage carousel")
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return s.ListCarouselSettings(ctx)
}

// 13. ListModuleCardSettings returns configured homepage module cards.
func (s *HomeContentService) ListModuleCardSettings(ctx context.Context) ([]HomeModuleCardResponse, error) {
	placements, assetMap, err := s.loadHomePlacements(ctx, homePlacementTypeModuleCard)
	if err != nil {
		return nil, err
	}

	items := make([]HomeModuleCardResponse, 0, len(placements))
	for _, placement := range placements {
		asset, exists := assetMap[placement.MediaAssetID]
		if !exists {
			continue
		}
		items = append(items, HomeModuleCardResponse{
			ModuleCode:   placement.ModuleCode,
			MediaAssetID: asset.PublicID,
			URL:          buildMediaURL(s.runtime.Config.MediaBaseURL, asset.ObjectKey),
			ObjectKey:    asset.ObjectKey,
			Title:        placement.Title,
			Subtitle:     placement.Subtitle,
			Body:         placement.Body,
		})
	}

	return items, nil
}

// 14. SaveModuleCardSettings replaces the homepage module card configuration.
func (s *HomeContentService) SaveModuleCardSettings(ctx context.Context, operatorUserID int64, inputs []HomeModuleCardInput) ([]HomeModuleCardResponse, error) {
	normalizedInputs, err := normalizeHomeModuleCardInputs(inputs)
	if err != nil {
		return nil, err
	}
	assets, err := s.loadOwnedHomePNGAssets(ctx, operatorUserID, homeModuleCardAssetIDs(normalizedInputs))
	if err != nil {
		return nil, err
	}

	err = s.runtime.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("placement_type = ?", homePlacementTypeModuleCard).Delete(&model.HomeContentPlacement{}).Error; err != nil {
			return errcode.New(errcode.CodeInternalError, "failed to clear homepage module cards")
		}
		for index, input := range normalizedInputs {
			asset := assets[input.MediaAssetID]
			placement := model.HomeContentPlacement{
				PlacementType: homePlacementTypeModuleCard,
				ModuleCode:    input.ModuleCode,
				SlotIndex:     index + 1,
				MediaAssetID:  asset.ID,
				Title:         input.Title,
				Subtitle:      input.Subtitle,
				Body:          input.Body,
				CreatedBy:     &operatorUserID,
				UpdatedBy:     &operatorUserID,
			}
			if err := tx.Create(&placement).Error; err != nil {
				return errcode.New(errcode.CodeInternalError, "failed to save homepage module cards")
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return s.ListModuleCardSettings(ctx)
}

// 15. ListLoginHeroSettings returns configured login hero images.
func (s *HomeContentService) ListLoginHeroSettings(ctx context.Context) ([]LoginHeroResponse, error) {
	return s.listLoginHeroSettings(ctx)
}

// 16. SaveLoginHeroSettings replaces login hero image configuration.
func (s *HomeContentService) SaveLoginHeroSettings(ctx context.Context, operatorUserID int64, inputs []LoginHeroInput) ([]LoginHeroResponse, error) {
	normalizedInputs, err := normalizeLoginHeroInputs(inputs)
	if err != nil {
		return nil, err
	}

	assets, err := s.loadOwnedImageAssetsInPrefix(ctx, operatorUserID, loginHeroAssetIDs(normalizedInputs), loginBagMediaObjectPrefix, "login hero")
	if err != nil {
		return nil, err
	}

	err = s.runtime.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("placement_type = ?", homePlacementTypeLoginHero).Delete(&model.HomeContentPlacement{}).Error; err != nil {
			return errcode.New(errcode.CodeInternalError, "failed to clear login hero")
		}

		for index, input := range normalizedInputs {
			asset := assets[input.MediaAssetID]
			slotIndex := input.SortOrder
			if slotIndex <= 0 {
				slotIndex = index + 1
			}
			placement := model.HomeContentPlacement{
				PlacementType: homePlacementTypeLoginHero,
				ModuleCode:    "",
				SlotIndex:     slotIndex,
				MediaAssetID:  asset.ID,
				Title:         strings.TrimSpace(input.Author),
				Subtitle:      strings.TrimSpace(input.Location),
				CreatedBy:     &operatorUserID,
				UpdatedBy:     &operatorUserID,
			}
			if err := tx.Create(&placement).Error; err != nil {
				return errcode.New(errcode.CodeInternalError, "failed to save login hero")
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return s.ListLoginHeroSettings(ctx)
}

// 17. listLoginHeroSettings returns configured login hero image list.
func (s *HomeContentService) listLoginHeroSettings(ctx context.Context) ([]LoginHeroResponse, error) {
	placements, assetMap, err := s.loadHomePlacements(ctx, homePlacementTypeLoginHero)
	if err != nil {
		return nil, err
	}

	items := make([]LoginHeroResponse, 0, len(placements))
	for _, placement := range placements {
		asset, exists := assetMap[placement.MediaAssetID]
		if !exists {
			continue
		}
		items = append(items, LoginHeroResponse{
			MediaAssetID: asset.PublicID,
			URL:          buildMediaURL(s.runtime.Config.MediaBaseURL, asset.ObjectKey),
			ObjectKey:    asset.ObjectKey,
			Author:       placement.Title,
			Location:     placement.Subtitle,
			SortOrder:    placement.SlotIndex,
		})
	}

	return items, nil
}

// 18. loadHomePlacements loads saved homepage slots and their media assets.
func (s *HomeContentService) loadHomePlacements(ctx context.Context, placementType string) ([]model.HomeContentPlacement, map[int64]model.MediaAsset, error) {
	var placements []model.HomeContentPlacement
	if err := s.runtime.DB.WithContext(ctx).
		Where("placement_type = ?", placementType).
		Order("slot_index asc").
		Find(&placements).Error; err != nil {
		return nil, nil, errcode.New(errcode.CodeInternalError, "failed to load homepage content")
	}

	assetIDs := make([]int64, 0, len(placements))
	for _, placement := range placements {
		if placement.MediaAssetID > 0 {
			assetIDs = append(assetIDs, placement.MediaAssetID)
		}
	}

	assetMap := map[int64]model.MediaAsset{}
	if len(assetIDs) == 0 {
		return placements, assetMap, nil
	}

	var assets []model.MediaAsset
	if err := s.runtime.DB.WithContext(ctx).Where("id IN ?", assetIDs).Find(&assets).Error; err != nil {
		return nil, nil, errcode.New(errcode.CodeInternalError, "failed to load homepage media")
	}
	for _, asset := range assets {
		assetMap[asset.ID] = asset
	}

	return placements, assetMap, nil
}

// 19. loadOwnedHomePNGAssets validates owned PNG media for homepage content.
func (s *HomeContentService) loadOwnedHomePNGAssets(ctx context.Context, operatorUserID int64, mediaAssetIDs []string) (map[string]model.MediaAsset, error) {
	return s.loadOwnedImageAssetsInPrefix(ctx, operatorUserID, mediaAssetIDs, homeEngMediaObjectPrefix, "homepage media")
}

// 20. loadOwnedImageAssetsInPrefix validates owned image media under a required directory.
func (s *HomeContentService) loadOwnedImageAssetsInPrefix(ctx context.Context, operatorUserID int64, mediaAssetIDs []string, objectPrefix string, label string) (map[string]model.MediaAsset, error) {
	result := map[string]model.MediaAsset{}
	if len(mediaAssetIDs) == 0 {
		return result, nil
	}

	var assets []model.MediaAsset
	if err := s.runtime.DB.WithContext(ctx).
		Where("public_id IN ? AND created_by = ?", mediaAssetIDs, operatorUserID).
		Find(&assets).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load "+label)
	}
	for _, asset := range assets {
		result[asset.PublicID] = asset
	}

	for _, mediaAssetID := range mediaAssetIDs {
		asset, exists := result[mediaAssetID]
		if !exists {
			return nil, errcode.New(errcode.CodeValidationError, label+" asset not found")
		}
		if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(asset.MimeType)), "image/") {
			return nil, errcode.New(errcode.CodeValidationError, label+" asset must be an image")
		}
		if !strings.HasPrefix(strings.TrimSpace(asset.ObjectKey), objectPrefix) {
			return nil, errcode.New(errcode.CodeValidationError, label+" asset uploaded under invalid directory")
		}
	}

	return result, nil
}

// 21. normalizeHomeCarouselInputs cleans and orders carousel inputs.
func normalizeHomeCarouselInputs(inputs []HomeCarouselInput) ([]HomeCarouselInput, error) {
	result := make([]HomeCarouselInput, 0, len(inputs))
	seen := map[string]bool{}
	for _, input := range inputs {
		mediaAssetID := strings.TrimSpace(input.MediaAssetID)
		if mediaAssetID == "" {
			continue
		}
		if seen[mediaAssetID] {
			return nil, errcode.New(errcode.CodeValidationError, "duplicated homepage carousel image")
		}
		seen[mediaAssetID] = true
		sortOrder := input.SortOrder
		if sortOrder <= 0 {
			sortOrder = len(result) + 1
		}
		result = append(result, HomeCarouselInput{
			MediaAssetID: mediaAssetID,
			SortOrder:    sortOrder,
		})
	}

	sort.SliceStable(result, func(left int, right int) bool {
		return result[left].SortOrder < result[right].SortOrder
	})

	return result, nil
}

// 22. normalizeHomeModuleCardInputs validates the fixed module card input set.
func normalizeHomeModuleCardInputs(inputs []HomeModuleCardInput) ([]HomeModuleCardInput, error) {
	if len(inputs) != len(allowedHomeModuleCodes) {
		return nil, errcode.New(errcode.CodeValidationError, "homepage module cards must include three modules")
	}

	inputMap := make(map[string]HomeModuleCardInput, len(inputs))
	for _, input := range inputs {
		moduleCode := strings.TrimSpace(input.ModuleCode)
		if !isAllowedHomeModuleCode(moduleCode) {
			return nil, errcode.New(errcode.CodeValidationError, "invalid homepage module code")
		}
		if _, exists := inputMap[moduleCode]; exists {
			return nil, errcode.New(errcode.CodeValidationError, "duplicated homepage module code")
		}
		mediaAssetID := strings.TrimSpace(input.MediaAssetID)
		if mediaAssetID == "" {
			return nil, errcode.New(errcode.CodeValidationError, "homepage module image is required")
		}
		inputMap[moduleCode] = HomeModuleCardInput{
			ModuleCode:   moduleCode,
			MediaAssetID: mediaAssetID,
			Title:        strings.TrimSpace(input.Title),
			Subtitle:     strings.TrimSpace(input.Subtitle),
			Body:         strings.TrimSpace(input.Body),
		}
	}

	result := make([]HomeModuleCardInput, 0, len(allowedHomeModuleCodes))
	for _, moduleCode := range allowedHomeModuleCodes {
		input, exists := inputMap[moduleCode]
		if !exists {
			return nil, errcode.New(errcode.CodeValidationError, "homepage module card is missing")
		}
		result = append(result, input)
	}

	return result, nil
}

// 23. normalizeLoginHeroInputs validates and sorts login hero inputs.
func normalizeLoginHeroInputs(inputs []LoginHeroInput) ([]LoginHeroInput, error) {
	if len(inputs) == 0 {
		return nil, errcode.New(errcode.CodeValidationError, "login hero image is required")
	}
	if len(inputs) > loginHeroSlotLimit {
		return nil, errcode.New(errcode.CodeValidationError, "too many login hero images")
	}

	result := make([]LoginHeroInput, 0, len(inputs))
	seen := map[string]bool{}
	for _, input := range inputs {
		mediaAssetID := strings.TrimSpace(input.MediaAssetID)
		if mediaAssetID == "" {
			continue
		}
		if seen[mediaAssetID] {
			return nil, errcode.New(errcode.CodeValidationError, "duplicated login hero image")
		}
		seen[mediaAssetID] = true
		sortOrder := input.SortOrder
		if sortOrder <= 0 {
			sortOrder = len(result) + 1
		}
		result = append(result, LoginHeroInput{
			MediaAssetID: mediaAssetID,
			Author:       strings.TrimSpace(input.Author),
			Location:     strings.TrimSpace(input.Location),
			SortOrder:    sortOrder,
		})
	}
	if len(result) == 0 {
		return nil, errcode.New(errcode.CodeValidationError, "login hero image is required")
	}

	sort.SliceStable(result, func(left int, right int) bool {
		return result[left].SortOrder < result[right].SortOrder
	})

	return result, nil
}

// 24. homeCarouselAssetIDs returns unique carousel media public IDs.
func homeCarouselAssetIDs(inputs []HomeCarouselInput) []string {
	result := make([]string, 0, len(inputs))
	for _, input := range inputs {
		result = append(result, input.MediaAssetID)
	}

	return result
}

// 25. homeModuleCardAssetIDs returns unique module card media public IDs.
func homeModuleCardAssetIDs(inputs []HomeModuleCardInput) []string {
	result := make([]string, 0, len(inputs))
	for _, input := range inputs {
		result = append(result, input.MediaAssetID)
	}

	return result
}

// 26. loginHeroAssetIDs returns unique login hero media public IDs.
func loginHeroAssetIDs(inputs []LoginHeroInput) []string {
	result := make([]string, 0, len(inputs))
	for _, input := range inputs {
		result = append(result, input.MediaAssetID)
	}

	return result
}

// 27. isAllowedHomeModuleCode checks the fixed homepage module list.
func isAllowedHomeModuleCode(moduleCode string) bool {
	for _, allowed := range allowedHomeModuleCodes {
		if moduleCode == allowed {
			return true
		}
	}

	return false
}

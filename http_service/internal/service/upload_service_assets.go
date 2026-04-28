/*
 * Upload media asset queries and mutations.
 * 1. Provide member-facing media asset list, detail, and delete flows.
 * 2. Keep media asset ownership and listing reference checks centralized.
 */
package service

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/model"
)

// 1. MediaAssetListFilters defines owned media asset list filters.
type MediaAssetListFilters struct {
	Page     int
	PageSize int
}

// 2. ListMediaAssets returns paginated media assets owned by the current user.
func (s *UploadService) ListMediaAssets(ctx context.Context, userID int64, filters MediaAssetListFilters) ([]CompleteUploadResult, *model.Pagination, error) {
	page, pageSize := normalizePagination(filters.Page, filters.PageSize)
	query := s.runtime.DB.WithContext(ctx).Model(&model.MediaAsset{}).Where("created_by = ?", userID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, nil, errcode.New(errcode.CodeInternalError, "failed to count media assets")
	}

	var assets []model.MediaAsset
	if err := query.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&assets).Error; err != nil {
		return nil, nil, errcode.New(errcode.CodeInternalError, "failed to load media assets")
	}

	items, err := s.buildMediaAssetResults(ctx, assets)
	if err != nil {
		return nil, nil, err
	}

	return items, &model.Pagination{Page: page, PageSize: pageSize, Total: total}, nil
}

// 3. GetMediaAsset returns a single owned media asset payload.
func (s *UploadService) GetMediaAsset(ctx context.Context, userID int64, mediaAssetID string) (*CompleteUploadResult, error) {
	asset, err := s.loadOwnedMediaAsset(ctx, userID, mediaAssetID)
	if err != nil {
		return nil, err
	}

	return s.buildMediaAssetResult(ctx, asset)
}

// 4. DeleteMediaAsset removes an owned media asset that is not referenced by listings.
func (s *UploadService) DeleteMediaAsset(ctx context.Context, userID int64, mediaAssetID string) (*CompleteUploadResult, error) {
	asset, err := s.loadOwnedMediaAsset(ctx, userID, mediaAssetID)
	if err != nil {
		return nil, err
	}

	usageMap, err := s.loadMediaUsageMap(ctx, []int64{asset.ID})
	if err != nil {
		return nil, err
	}
	if usageMap[asset.ID] {
		return nil, errcode.New(errcode.CodeValidationError, "media asset is already referenced by a listing")
	}

	result, err := s.buildMediaAssetResult(ctx, asset)
	if err != nil {
		return nil, err
	}

	if err := s.runtime.StorageProvider.DeleteObject(ctx, asset.ObjectKey); err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to delete media object from storage")
	}
	if err := s.runtime.DB.WithContext(ctx).Delete(&model.MediaAsset{}, asset.ID).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to delete media asset")
	}

	return result, nil
}

// 5. buildMediaAssetResult maps a model into a shared API payload.
func (s *UploadService) buildMediaAssetResult(ctx context.Context, asset *model.MediaAsset) (*CompleteUploadResult, error) {
	items, err := s.buildMediaAssetResults(ctx, []model.MediaAsset{*asset})
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, errcode.New(errcode.CodeNotFound, "media asset not found")
	}

	return &items[0], nil
}

// 6. buildMediaAssetResults maps models into shared API payloads with usage flags.
func (s *UploadService) buildMediaAssetResults(ctx context.Context, assets []model.MediaAsset) ([]CompleteUploadResult, error) {
	if len(assets) == 0 {
		return []CompleteUploadResult{}, nil
	}

	assetIDs := make([]int64, 0, len(assets))
	for _, asset := range assets {
		assetIDs = append(assetIDs, asset.ID)
	}

	usageMap, err := s.loadMediaUsageMap(ctx, assetIDs)
	if err != nil {
		return nil, err
	}

	items := make([]CompleteUploadResult, 0, len(assets))
	for _, asset := range assets {
		items = append(items, CompleteUploadResult{
			MediaAssetID:    asset.PublicID,
			StorageProvider: asset.StorageProvider,
			BucketName:      asset.BucketName,
			ObjectKey:       asset.ObjectKey,
			MimeType:        asset.MimeType,
			Width:           asset.Width,
			Height:          asset.Height,
			FileSize:        asset.FileSize,
			ChecksumSHA256:  asset.ChecksumSHA256,
			URL:             buildMediaURL(s.runtime.Config.MediaBaseURL, asset.ObjectKey),
			InUse:           usageMap[asset.ID],
			CreatedAt:       asset.CreatedAt,
		})
	}

	return items, nil
}

// 7. loadOwnedMediaAsset loads a media asset owned by the current user.
func (s *UploadService) loadOwnedMediaAsset(ctx context.Context, userID int64, mediaAssetID string) (*model.MediaAsset, error) {
	var asset model.MediaAsset
	if err := s.runtime.DB.WithContext(ctx).
		Where("public_id = ? AND created_by = ?", mediaAssetID, userID).
		First(&asset).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.New(errcode.CodeNotFound, "media asset not found")
		}
		return nil, errcode.New(errcode.CodeInternalError, "failed to load media asset")
	}

	return &asset, nil
}

// 8. loadMediaUsageMap returns whether each media asset is already referenced by listings.
func (s *UploadService) loadMediaUsageMap(ctx context.Context, assetIDs []int64) (map[int64]bool, error) {
	result := make(map[int64]bool, len(assetIDs))
	if len(assetIDs) == 0 {
		return result, nil
	}

	type usageRow struct {
		MediaAssetID int64
	}

	var rows []usageRow
	if err := s.runtime.DB.WithContext(ctx).
		Model(&model.ListingImage{}).
		Distinct("media_asset_id").
		Where("media_asset_id IN ?", assetIDs).
		Scan(&rows).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load media asset usage")
	}

	for _, row := range rows {
		result[row.MediaAssetID] = true
	}

	return result, nil
}

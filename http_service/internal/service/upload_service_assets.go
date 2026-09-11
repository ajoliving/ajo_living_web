/*
 * Upload media asset queries and mutations.
 * 1. Provide member-facing media asset list, detail, and delete flows.
 * 2. Keep media asset ownership and listing reference checks centralized.
 */
package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

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
		return nil, errcode.New(errcode.CodeValidationError, "media asset is already referenced")
	}
	referenceCount, err := s.mediaObjectReferenceCount(ctx, asset.BucketName, asset.ObjectKey)
	if err != nil {
		return nil, err
	}
	if referenceCount > 1 {
		return nil, errcode.New(errcode.CodeValidationError, "media object is referenced by another asset")
	}

	result, err := s.buildMediaAssetResult(ctx, asset)
	if err != nil {
		return nil, err
	}

	if err := s.runtime.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var lockedAsset model.MediaAsset
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND created_by = ?", asset.ID, userID).First(&lockedAsset).Error; err != nil {
			return err
		}
		var messageReferenceCount int64
		if tx.Migrator().HasTable(&model.MessageAttachment{}) {
			if err := tx.Model(&model.MessageAttachment{}).Where("media_asset_id = ?", lockedAsset.ID).Count(&messageReferenceCount).Error; err != nil {
				return err
			}
		}
		if messageReferenceCount > 0 {
			return errcode.New(errcode.CodeValidationError, "media asset is already referenced")
		}
		if err := s.runtime.StorageProvider.DeleteObject(ctx, lockedAsset.ObjectKey); err != nil {
			return errcode.New(errcode.CodeInternalError, "failed to delete media object from storage")
		}
		return tx.Delete(&model.MediaAsset{}, lockedAsset.ID).Error
	}); err != nil {
		var appErr *errcode.AppError
		if errors.As(err, &appErr) {
			return nil, appErr
		}
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
		assetURL := buildMediaURL(s.runtime.Config.MediaBaseURL, asset.ObjectKey)
		if strings.HasPrefix(asset.ObjectKey, chatMediaObjectPrefix) {
			assetURL = chatAttachmentURL(s.runtime.Config.MediaBaseURL, asset)
		}
		if isPrivateAgencyEvidenceObjectKey(asset.ObjectKey) {
			if s.runtime.StorageProvider == nil {
				return nil, errcode.New(errcode.CodeInternalError, "private media storage is not configured")
			}
			signedURL, err := s.runtime.StorageProvider.PresignDownload(ctx, asset.ObjectKey, 10*time.Minute)
			if err != nil {
				return nil, errcode.New(errcode.CodeInternalError, "failed to authorize private media download")
			}
			assetURL = signedURL
		}
		items = append(items, CompleteUploadResult{
			MediaAssetID:     asset.PublicID,
			StorageProvider:  asset.StorageProvider,
			BucketName:       asset.BucketName,
			ObjectKey:        asset.ObjectKey,
			MimeType:         asset.MimeType,
			Width:            asset.Width,
			Height:           asset.Height,
			FileSize:         asset.FileSize,
			ChecksumSHA256:   asset.ChecksumSHA256,
			URL:              assetURL,
			InUse:            usageMap[asset.ID],
			CreatedAt:        asset.CreatedAt,
			ProcessingStatus: normalizedProcessingStatus(asset),
			ScanStatus:       normalizedScanStatus(asset),
			RejectionReason:  asset.RejectionReason,
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

// 8. loadMediaUsageMap returns whether each media asset is already referenced by listings, profiles, or chat messages.
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

	var profileRows []usageRow
	if err := s.runtime.DB.WithContext(ctx).
		Model(&model.UserProfile{}).
		Distinct("avatar_asset_id AS media_asset_id").
		Where("avatar_asset_id IN ?", assetIDs).
		Scan(&profileRows).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load media asset usage")
	}
	for _, row := range profileRows {
		result[row.MediaAssetID] = true
	}

	if s.runtime.DB.Migrator().HasTable(&model.MessageAttachment{}) {
		var messageRows []usageRow
		if err := s.runtime.DB.WithContext(ctx).
			Model(&model.MessageAttachment{}).
			Distinct("media_asset_id").
			Where("media_asset_id IN ?", assetIDs).
			Scan(&messageRows).Error; err != nil {
			return nil, errcode.New(errcode.CodeInternalError, "failed to load media asset usage")
		}
		for _, row := range messageRows {
			result[row.MediaAssetID] = true
		}
	}

	var homeRows []usageRow
	if err := s.runtime.DB.WithContext(ctx).
		Model(&model.HomeContentPlacement{}).
		Distinct("media_asset_id").
		Where("media_asset_id IN ?", assetIDs).
		Scan(&homeRows).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load media asset usage")
	}
	for _, row := range homeRows {
		result[row.MediaAssetID] = true
	}

	for _, column := range []string{"avatar_asset_id", "wechat_qr_asset_id", "logo_asset_id", "eaa_license_asset_id", "business_registration_asset_id", "company_card_asset_id"} {
		var rows []usageRow
		if err := s.runtime.DB.WithContext(ctx).Model(&model.AgencyProfile{}).
			Distinct(column+" AS media_asset_id").Where(column+" IN ?", assetIDs).Scan(&rows).Error; err != nil {
			return nil, errcode.New(errcode.CodeInternalError, "failed to load agency profile media usage")
		}
		for _, row := range rows {
			result[row.MediaAssetID] = true
		}
	}

	return result, nil
}

// 9. mediaObjectReferenceCount counts database rows pointing at one object key.
func (s *UploadService) mediaObjectReferenceCount(ctx context.Context, bucketName string, objectKey string) (int64, error) {
	var count int64
	if err := s.runtime.DB.WithContext(ctx).
		Model(&model.MediaAsset{}).
		Where("bucket_name = ? AND object_key = ?", bucketName, objectKey).
		Count(&count).Error; err != nil {
		return 0, errcode.New(errcode.CodeInternalError, "failed to load media object references")
	}

	return count, nil
}

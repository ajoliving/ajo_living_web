/*
 * Chat response support helpers.
 * 1. Load peer display payloads for chat list and detail responses.
 * 2. Clean orphan chat media after the retention window.
 */
package service

import (
	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/model"
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strconv"
	"strings"
	"time"
)

// 1. chatPeerRow defines batched peer query fields.
type chatPeerRow struct {
	ChatPublicID string
	UserID       int64
	RoleInChat   string
}

// 2. loadChatPeerMap loads the opposite-side display payload for chats.
func (s *ChatService) loadChatPeerMap(ctx context.Context, userID int64, chatPublicIDs []string) (map[string]*ChatPeerSummary, error) {
	result := make(map[string]*ChatPeerSummary)
	if len(chatPublicIDs) == 0 {
		return result, nil
	}

	var rows []chatPeerRow
	if err := s.runtime.DB.WithContext(ctx).Table("chat_participants").
		Select("chats.public_id AS chat_public_id, chat_participants.user_id, chat_participants.role_in_chat").
		Joins("JOIN chats ON chats.id = chat_participants.chat_id").
		Where("chats.public_id IN ? AND chat_participants.user_id <> ?", chatPublicIDs, userID).
		Scan(&rows).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load chat peers")
	}

	profileMap, err := s.secondhandService.loadUserPreviewMap(ctx, extractPeerUserIDs(rows))
	if err != nil {
		return nil, err
	}

	for _, item := range rows {
		profile := profileMap[item.UserID]
		if item.RoleInChat == "system" {
			result[item.ChatPublicID] = &ChatPeerSummary{
				UserID:      fmtInt64(item.UserID),
				PublicID:    safeUserPublicID(profile),
				DisplayName: model.SystemNotificationDisplayName,
				RoleInChat:  item.RoleInChat,
			}
			continue
		}

		result[item.ChatPublicID] = &ChatPeerSummary{
			UserID:      fmtInt64(item.UserID),
			PublicID:    safeUserPublicID(profile),
			DisplayName: safeUserDisplayName(profile, fmtInt64(item.UserID)),
			RoleInChat:  item.RoleInChat,
		}
	}

	return result, nil
}

// 3. extractListingIDs converts chat rows into listing ID slice.
func extractListingIDs(rows []chatListRow) []int64 {
	result := make([]int64, 0, len(rows))
	for _, item := range rows {
		result = append(result, item.ListingID)
	}
	return result
}

// 4. extractChatIDs converts chat rows into chat public ID slice.
func extractChatIDs(rows []chatListRow) []string {
	result := make([]string, 0, len(rows))
	for _, item := range rows {
		result = append(result, item.ChatPublicID)
	}
	return result
}

// 5. extractParticipantUserIDs converts participants into user ID slice.
func extractParticipantUserIDs(participants []model.ChatParticipant) []int64 {
	result := make([]int64, 0, len(participants))
	for _, item := range participants {
		result = append(result, item.UserID)
	}
	return result
}

// 6. extractPeerUserIDs converts chat peer rows into user ID slice.
func extractPeerUserIDs(rows []chatPeerRow) []int64 {
	result := make([]int64, 0, len(rows))
	for _, item := range rows {
		result = append(result, item.UserID)
	}
	return result
}

// 7. buildChatListingSummary maps listing fields into chat listing payload.
func buildChatListingSummary(row chatListRow, cover *ListingImageResponse) *ChatListingSummary {
	var publishedAt *string
	if row.PublishedAt != nil {
		value := row.PublishedAt.UTC().Format(time.RFC3339)
		publishedAt = &value
	}

	return &ChatListingSummary{
		ListingID:      row.ListingPublicID,
		BizModule:      row.BizModule,
		Title:          row.Title,
		Summary:        row.Summary,
		PublishedAt:    publishedAt,
		BusinessStatus: row.BusinessStatus,
		PriceMode:      row.PriceMode,
		PriceHKD:       row.PriceHKD,
		CoverImage:     cover,
	}
}

// 8. safeUserPublicID returns a preview public ID fallback.
func safeUserPublicID(profile *UserPreviewResponse) string {
	if profile == nil {
		return ""
	}
	return profile.PublicID
}

// 9. safeUserDisplayName returns a preview display name fallback.
func safeUserDisplayName(profile *UserPreviewResponse, fallback string) string {
	if profile == nil {
		return fallback
	}
	if strings.TrimSpace(profile.DisplayName) == "" {
		return fallback
	}
	return profile.DisplayName
}

// 10. fmtInt64 converts int64 to string.
func fmtInt64(value int64) string {
	return strconv.FormatInt(value, 10)
}

// 11. loadChatListingImages loads listing images for all chat modules.
func (s *ChatService) loadChatListingImages(ctx context.Context, listingIDs []int64) (map[int64][]ListingImageResponse, error) {
	result := make(map[int64][]ListingImageResponse)
	if len(listingIDs) == 0 {
		return result, nil
	}

	var images []model.ListingImage
	if err := s.runtime.DB.WithContext(ctx).Where("listing_id IN ?", listingIDs).Order("sort_order asc").Find(&images).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load listing images")
	}

	assetIDs := make([]int64, 0, len(images))
	for _, image := range images {
		assetIDs = append(assetIDs, image.MediaAssetID)
	}

	var assets []model.MediaAsset
	if len(assetIDs) > 0 {
		if err := s.runtime.DB.WithContext(ctx).Where("id IN ?", assetIDs).Find(&assets).Error; err != nil {
			return nil, errcode.New(errcode.CodeInternalError, "failed to load media assets")
		}
	}

	assetMap := make(map[int64]model.MediaAsset, len(assets))
	for _, asset := range assets {
		assetMap[asset.ID] = asset
	}

	for _, image := range images {
		asset := assetMap[image.MediaAssetID]
		result[image.ListingID] = append(result[image.ListingID], ListingImageResponse{
			MediaAssetID: asset.PublicID,
			URL:          buildMediaURL(s.runtime.Config.MediaBaseURL, asset.ObjectKey),
			SortOrder:    image.SortOrder,
			IsCover:      image.IsCover,
		})
	}

	return result, nil
}

// 1. CleanupOrphanChatMedia removes stale unbound chat uploads.
func (s *ChatService) CleanupOrphanChatMedia(ctx context.Context, olderThan time.Duration, limit int) (int, error) {
	if s == nil || s.runtime == nil || s.runtime.DB == nil || s.runtime.StorageProvider == nil || !s.runtime.DB.Migrator().HasTable(&model.MessageAttachment{}) {
		return 0, nil
	}
	if olderThan <= 0 {
		olderThan = 24 * time.Hour
	}
	if limit <= 0 || limit > 500 {
		limit = 500
	}
	cutoff := s.runtime.Now().Add(-olderThan)
	var assets []model.MediaAsset
	if err := s.runtime.DB.WithContext(ctx).Where("object_key LIKE ? AND created_at < ? AND id NOT IN (SELECT media_asset_id FROM message_attachments)", chatMediaObjectPrefix+"%", cutoff).Order("id asc").Limit(limit).Find(&assets).Error; err != nil {
		return 0, err
	}
	removed := 0
	for _, asset := range assets {
		if err := s.runtime.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			var lockedAsset model.MediaAsset
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", asset.ID).First(&lockedAsset).Error; err != nil {
				return err
			}
			var referenceCount int64
			if err := tx.Model(&model.MessageAttachment{}).Where("media_asset_id = ?", lockedAsset.ID).Count(&referenceCount).Error; err != nil {
				return err
			}
			if referenceCount > 0 {
				return gorm.ErrRecordNotFound
			}
			if err := s.runtime.StorageProvider.DeleteObject(ctx, lockedAsset.ObjectKey); err != nil {
				return err
			}
			return tx.Delete(&model.MediaAsset{}, lockedAsset.ID).Error
		}); err != nil {
			continue
		}
		removed++
	}
	return removed, nil
}

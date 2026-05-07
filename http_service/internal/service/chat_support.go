/*
 * Chat response support helpers.
 * 1. Load peer display payloads for chat list and detail responses.
 * 2. Keep chat summary mapping helpers outside the core message workflow file.
 */
package service

import (
	"context"
	"strconv"
	"strings"
	"time"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/model"
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

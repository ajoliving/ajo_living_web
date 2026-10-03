/*
 * Chat attachment helpers.
 * 1. Validate uploaded assets before message creation.
 * 2. Load attachment metadata for history and realtime responses.
 */
package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// 1. optionalString stores nullable idempotency keys without empty-value collisions.
func optionalString(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	cleaned := strings.TrimSpace(value)
	return &cleaned
}

// 2. buildRealtimeMessagePayload serializes the durable realtime event body.
func buildRealtimeMessagePayload(chatID string, message MessageResponse, clientMessageID string) (string, error) {
	payload, err := json.Marshal(map[string]any{"type": "message", "chat_id": chatID, "message": message, "client_message_id": strings.TrimSpace(clientMessageID)})
	if err != nil {
		return "", err
	}
	return string(payload), nil
}

// 2.1 buildRealtimeRecalledPayload serializes a durable recall event without original content.
func buildRealtimeRecalledPayload(chatID string, message MessageResponse) (string, error) {
	payload, err := json.Marshal(map[string]any{"type": realtimeEventMessageRecalled, "chat_id": chatID, "message": message})
	if err != nil {
		return "", err
	}
	return string(payload), nil
}

// 2.2 recalledOutboxAggregateID keeps recall events distinct from the original message outbox row.
func recalledOutboxAggregateID(messagePublicID string) string {
	return strings.TrimSpace(messagePublicID) + ":recalled"
}

// 2.3 recalledMessageResponse hides original text and attachment URLs after recall.
func recalledMessageResponse(message model.Message) MessageResponse {
	return MessageResponse{MessageID: message.PublicID, SenderUserID: fmtInt64(message.SenderUserID), Content: "", MessageType: message.MessageType, Status: messageStatusRecalled, CreatedAt: message.CreatedAt.UTC().Format(time.RFC3339)}
}

// 3. messageResponseFromAssets builds the event response before transaction commit.
func messageResponseFromAssets(message model.Message, assets []model.MediaAsset, baseURL string) MessageResponse {
	attachments := make([]MessageAttachmentResponse, 0, len(assets))
	for _, asset := range assets {
		attachments = append(attachments, MessageAttachmentResponse{MediaAssetID: asset.PublicID, MimeType: asset.MimeType, FileSize: asset.FileSize, Width: asset.Width, Height: asset.Height, URL: chatAttachmentURL(baseURL, asset), ProcessingStatus: normalizedProcessingStatus(asset), ScanStatus: normalizedScanStatus(asset), RejectionReason: asset.RejectionReason})
	}
	return MessageResponse{MessageID: message.PublicID, SenderUserID: fmtInt64(message.SenderUserID), Content: message.ContentText, MessageType: message.MessageType, Status: message.MessageStatus, CreatedAt: message.CreatedAt.UTC().Format(time.RFC3339), Attachments: attachments}
}

// 3.1 realtimeMessagePayload attaches reply preview before outbox serialization.
func realtimeMessagePayload(chatPublicID string, message model.Message, assets []model.MediaAsset, baseURL string, replyTo *MessageReplyPreview, clientMessageID string) (string, error) {
	response := messageResponseFromAssets(message, assets, baseURL)
	response.ReplyTo = replyTo
	return buildRealtimeMessagePayload(chatPublicID, response, clientMessageID)
}

// 4. chatAttachmentURL only exposes media after processing and safety checks pass.
func chatAttachmentURL(baseURL string, asset model.MediaAsset) string {
	// Chat media must pass an explicit processing and scan state. Legacy rows
	// with blank state are not safe to expose until the worker has examined them.
	if strings.HasPrefix(strings.TrimSpace(asset.ObjectKey), chatMediaObjectPrefix) &&
		(strings.TrimSpace(asset.ProcessingStatus) == "" || strings.TrimSpace(asset.ScanStatus) == "") {
		return ""
	}
	if normalizedProcessingStatus(asset) != mediaProcessingReady || normalizedScanStatus(asset) != mediaScanPassed {
		return ""
	}
	objectKey := asset.ObjectKey
	if isVideoMedia(asset.MimeType) && strings.TrimSpace(asset.PlaybackObjectKey) != "" {
		objectKey = asset.PlaybackObjectKey
	}
	return buildMediaURL(baseURL, objectKey)
}

// 5. normalizeAttachmentIDs removes duplicates and rejects messages above the chat limit.
func normalizeAttachmentIDs(values []string) ([]string, error) {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
		if len(result) > messageAttachmentMax {
			return nil, errcode.New(errcode.CodeValidationError, "a message supports at most 5 attachments")
		}
	}
	return result, nil
}

// 6. loadChatAttachmentAssets validates ownership and preserves sender-selected order.
func (s *ChatService) loadChatAttachmentAssets(ctx context.Context, userID int64, publicIDs []string) ([]model.MediaAsset, error) {
	return s.loadChatAttachmentAssetsWithDB(ctx, s.runtime.DB, userID, publicIDs, false)
}

// 6.1 loadChatAttachmentAssetsWithDB locks attachments during message creation so cleanup cannot remove them mid-send.
func (s *ChatService) loadChatAttachmentAssetsWithDB(ctx context.Context, db *gorm.DB, userID int64, publicIDs []string, lockForMessage bool) ([]model.MediaAsset, error) {
	if len(publicIDs) == 0 {
		return nil, nil
	}
	var assets []model.MediaAsset
	query := db.WithContext(ctx)
	if lockForMessage {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if err := query.Where("public_id IN ? AND created_by = ?", publicIDs, userID).Find(&assets).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load chat attachments")
	}
	if len(assets) != len(publicIDs) {
		return nil, errcode.New(errcode.CodeAuthForbidden, "chat attachment is not accessible")
	}
	assetsByPublicID := make(map[string]model.MediaAsset, len(assets))
	for _, asset := range assets {
		mimeType := strings.ToLower(asset.MimeType)
		if !strings.HasPrefix(asset.ObjectKey, chatMediaObjectPrefix) || !(strings.HasPrefix(mimeType, "image/") || isSupportedChatMime(mimeType)) {
			return nil, errcode.New(errcode.CodeValidationError, "unsupported chat attachment")
		}
		assetsByPublicID[asset.PublicID] = asset
	}
	orderedAssets := make([]model.MediaAsset, 0, len(publicIDs))
	for _, publicID := range publicIDs {
		orderedAssets = append(orderedAssets, assetsByPublicID[publicID])
	}
	return orderedAssets, nil
}

// 7. loadMessageAttachmentMap loads attachment metadata for a message batch.
func (s *ChatService) loadMessageAttachmentMap(ctx context.Context, messageIDs []int64) (map[int64][]MessageAttachmentResponse, error) {
	result := make(map[int64][]MessageAttachmentResponse)
	if len(messageIDs) == 0 || !s.runtime.DB.Migrator().HasTable(&model.MessageAttachment{}) {
		return result, nil
	}
	var rows []struct {
		MessageID         int64
		SortOrder         int
		MediaAssetID      string
		MimeType          string
		FileSize          int64
		Width             *int
		Height            *int
		ObjectKey         string
		PlaybackObjectKey string
		ProcessingStatus  string
		ScanStatus        string
		RejectionReason   string
	}
	if err := s.runtime.DB.WithContext(ctx).Table("message_attachments").Select("message_attachments.message_id, message_attachments.sort_order, media_assets.public_id AS media_asset_id, media_assets.mime_type, media_assets.file_size, media_assets.width, media_assets.height, media_assets.object_key, media_assets.playback_object_key, media_assets.processing_status, media_assets.scan_status, media_assets.rejection_reason").Joins("JOIN media_assets ON media_assets.id = message_attachments.media_asset_id").Where("message_attachments.message_id IN ?", messageIDs).Order("message_attachments.message_id asc, message_attachments.sort_order asc").Scan(&rows).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load message attachments")
	}
	for _, row := range rows {
		asset := model.MediaAsset{PublicID: row.MediaAssetID, ObjectKey: row.ObjectKey, PlaybackObjectKey: row.PlaybackObjectKey, MimeType: row.MimeType, FileSize: row.FileSize, Width: row.Width, Height: row.Height, ProcessingStatus: row.ProcessingStatus, ScanStatus: row.ScanStatus, RejectionReason: row.RejectionReason}
		result[row.MessageID] = append(result[row.MessageID], MessageAttachmentResponse{MediaAssetID: row.MediaAssetID, MimeType: row.MimeType, FileSize: row.FileSize, Width: row.Width, Height: row.Height, URL: chatAttachmentURL(s.runtime.Config.MediaBaseURL, asset), ProcessingStatus: normalizedProcessingStatus(asset), ScanStatus: normalizedScanStatus(asset), RejectionReason: row.RejectionReason})
	}
	return result, nil
}

// 8. buildMessageResponses maps persisted messages, attachments, and reply previews.
func (s *ChatService) buildMessageResponses(ctx context.Context, messages []model.Message) ([]MessageResponse, error) {
	messageIDs := make([]int64, 0, len(messages))
	replyIDs := make([]int64, 0, len(messages))
	for _, message := range messages {
		if message.MessageStatus != messageStatusRecalled {
			messageIDs = append(messageIDs, message.ID)
		}
		if message.ReplyToMessageID != nil && *message.ReplyToMessageID > 0 {
			replyIDs = append(replyIDs, *message.ReplyToMessageID)
		}
	}
	attachmentMap, err := s.loadMessageAttachmentMap(ctx, messageIDs)
	if err != nil {
		return nil, err
	}
	replyMap, err := s.loadReplyPreviewMap(ctx, replyIDs)
	if err != nil {
		return nil, err
	}
	items := make([]MessageResponse, 0, len(messages))
	for _, message := range messages {
		var replyTo *MessageReplyPreview
		if message.ReplyToMessageID != nil {
			replyTo = replyMap[*message.ReplyToMessageID]
		}
		if message.MessageStatus == messageStatusRecalled {
			redacted := recalledMessageResponse(message)
			redacted.ReplyTo = replyTo
			items = append(items, redacted)
			continue
		}
		items = append(items, MessageResponse{MessageID: message.PublicID, SenderUserID: fmtInt64(message.SenderUserID), Content: message.ContentText, MessageType: message.MessageType, ActionLabel: message.ActionLabel, ActionURL: message.ActionURL, Status: message.MessageStatus, CreatedAt: message.CreatedAt.UTC().Format(time.RFC3339), Attachments: attachmentMap[message.ID], ReplyTo: replyTo})
	}
	return items, nil
}

// 8.1 resolveReplyTarget loads a same-chat original for a new reply.
func (s *ChatService) resolveReplyTarget(ctx context.Context, chatID int64, replyToPublicID string) (*int64, *MessageReplyPreview, error) {
	replyToPublicID = strings.TrimSpace(replyToPublicID)
	if replyToPublicID == "" {
		return nil, nil, nil
	}
	var target model.Message
	if err := s.runtime.DB.WithContext(ctx).Where("public_id = ?", replyToPublicID).First(&target).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, errcode.New(errcode.CodeValidationError, "reply target not found")
		}
		return nil, nil, errcode.New(errcode.CodeInternalError, "failed to load reply target")
	}
	if target.ChatID != chatID {
		return nil, nil, errcode.New(errcode.CodeValidationError, "reply target is not in this chat")
	}
	previews, err := s.loadReplyPreviewMap(ctx, []int64{target.ID})
	if err != nil {
		return nil, nil, err
	}
	return &target.ID, previews[target.ID], nil
}

// 8.2 loadReplyPreviewMap batches reply originals and sender display names.
func (s *ChatService) loadReplyPreviewMap(ctx context.Context, messageIDs []int64) (map[int64]*MessageReplyPreview, error) {
	result := make(map[int64]*MessageReplyPreview)
	if len(messageIDs) == 0 {
		return result, nil
	}
	seen := make(map[int64]struct{}, len(messageIDs))
	uniqueIDs := make([]int64, 0, len(messageIDs))
	for _, messageID := range messageIDs {
		if messageID <= 0 {
			continue
		}
		if _, exists := seen[messageID]; exists {
			continue
		}
		seen[messageID] = struct{}{}
		uniqueIDs = append(uniqueIDs, messageID)
	}
	if len(uniqueIDs) == 0 {
		return result, nil
	}
	var messages []model.Message
	if err := s.runtime.DB.WithContext(ctx).Where("id IN ?", uniqueIDs).Find(&messages).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load reply targets")
	}
	senderIDs := make([]int64, 0, len(messages))
	senderSeen := make(map[int64]struct{}, len(messages))
	for _, message := range messages {
		if _, exists := senderSeen[message.SenderUserID]; exists {
			continue
		}
		senderSeen[message.SenderUserID] = struct{}{}
		senderIDs = append(senderIDs, message.SenderUserID)
	}
	nameMap := make(map[int64]string, len(senderIDs))
	if len(senderIDs) > 0 {
		var profiles []model.UserProfile
		if err := s.runtime.DB.WithContext(ctx).Where("user_id IN ?", senderIDs).Find(&profiles).Error; err != nil {
			return nil, errcode.New(errcode.CodeInternalError, "failed to load reply authors")
		}
		for _, profile := range profiles {
			nameMap[profile.UserID] = strings.TrimSpace(profile.DisplayName)
		}
	}
	for _, message := range messages {
		preview := buildReplyPreview(message, nameMap[message.SenderUserID])
		result[message.ID] = &preview
	}
	return result, nil
}

// 8.3 buildReplyPreview maps one original message into a quote snippet.
func buildReplyPreview(message model.Message, senderName string) MessageReplyPreview {
	recalled := message.MessageStatus == messageStatusRecalled
	content := recalledMessagePreview
	if !recalled {
		content = strings.TrimSpace(message.ContentText)
		if content == "" {
			content = "[Attachment]"
		} else {
			content = truncateReplyContent(content)
		}
	}
	return MessageReplyPreview{
		MessageID:    message.PublicID,
		SenderUserID: fmtInt64(message.SenderUserID),
		SenderName:   senderName,
		Content:      content,
		Recalled:     recalled,
	}
}

// 8.4 truncateReplyContent keeps quote snippets short for list and bubble display.
func truncateReplyContent(content string) string {
	runes := []rune(content)
	if len(runes) <= replyPreviewMaxRunes {
		return content
	}
	return string(runes[:replyPreviewMaxRunes])
}

/*
 * Chat business logic.
 * 1. Create or reuse listing-scoped chats.
 * 2. Persist messages and unread state.
 * 3. Enforce participant and listing access checks across listing modules.
 */
package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/model"
	"ajoliving_web/http_service/internal/utils"
)

const (
	chatTypeDirectListing  = "direct_listing_chat"
	chatTypeSystemNotice   = "system_notice"
	messageTypeText        = "text"
	messageTypeAttachment  = "attachment"
	messageTypeMixed       = "mixed"
	messageContentMaxRunes = 1000
	messageAttachmentMax   = 5
)

// 1. ChatService handles listing chats and messages.
type ChatService struct {
	runtime           *Runtime
	secondhandService *SecondhandService
	propertyService   *PropertyService
}

// 2. chatListRow defines the internal chat list query row.
type chatListRow struct {
	UnreadCount        int
	ChatPublicID       string
	BizModule          string
	ChatType           string
	LastMessagePreview string
	LastMessageAt      *time.Time
	ListingPublicID    string
	Title              string
	Summary            string
	ListingID          int64
	BusinessStatus     string
	PublishedAt        *time.Time
	PriceMode          string
	PriceHKD           *float64
}

// 3. NewChatService creates a chat service instance.
func NewChatService(runtime *Runtime, secondhandService *SecondhandService, propertyService *PropertyService) *ChatService {
	return &ChatService{
		runtime:           runtime,
		secondhandService: secondhandService,
		propertyService:   propertyService,
	}
}

// 4. CreateOrReuseChat creates or reuses a secondhand listing chat for the current user.
func (s *ChatService) CreateOrReuseChat(ctx context.Context, userID int64, communityID *int64, listingPublicID string) (map[string]any, error) {
	listing, secondhand, contact, err := s.secondhandService.loadListingByPublicID(ctx, listingPublicID)
	if err != nil {
		return nil, err
	}
	if listing.OwnerUserID == userID {
		return nil, errcode.New(errcode.CodeValidationError, "publisher cannot start a chat with the same listing")
	}
	if listing.PublicationStatus != "active" || listing.ModerationStatus != "approved" || listing.BusinessStatus != "available" {
		return nil, errcode.New(errcode.CodeAuthForbidden, "listing is not available for chat")
	}
	if !s.secondhandService.canViewListing(listing, secondhand, communityID) {
		return nil, errcode.New(errcode.CodeVisibilityForbidden, "listing is not visible to the current user")
	}
	if !contact.ShowChat {
		return nil, errcode.New(errcode.CodeAuthForbidden, "chat is not enabled for this listing")
	}

	return s.createOrReuseListingChat(ctx, userID, listing, "secondhand")
}

// 5. CreateOrReusePropertyChat creates or reuses a property listing chat.
func (s *ChatService) CreateOrReusePropertyChat(ctx context.Context, channel PropertyChannel, userID int64, listingPublicID string) (map[string]any, error) {
	listing, _, err := s.propertyService.loadPropertyListingByPublicID(ctx, channel, listingPublicID)
	if err != nil {
		return nil, err
	}
	if listing.OwnerUserID == userID {
		return nil, errcode.New(errcode.CodeValidationError, "publisher cannot start a chat with the same listing")
	}
	if listing.PublicationStatus != "active" || listing.ModerationStatus != "approved" || listing.BusinessStatus != "available" {
		return nil, errcode.New(errcode.CodeAuthForbidden, "listing is not available for chat")
	}
	return s.createOrReuseListingChat(ctx, userID, listing, string(channel))
}

// 6. createOrReuseListingChat creates the shared chat aggregate.
func (s *ChatService) createOrReuseListingChat(ctx context.Context, userID int64, listing *model.Listing, bizModule string) (map[string]any, error) {
	var chat model.Chat
	err := s.runtime.DB.WithContext(ctx).Where("listing_id = ? AND created_by = ?", listing.ID, userID).First(&chat).Error
	if err == nil {
		return map[string]any{"chat_id": chat.PublicID, "is_new": false}, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load chat")
	}

	chat = model.Chat{
		PublicID:  utils.NewPublicID(),
		BizModule: strings.TrimSpace(bizModule),
		ListingID: listing.ID,
		ChatType:  chatTypeDirectListing,
		CreatedBy: userID,
		CreatedAt: s.runtime.Now(),
		UpdatedAt: s.runtime.Now(),
	}

	if err := s.runtime.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&chat).Error; err != nil {
			return err
		}

		participants := []model.ChatParticipant{
			{ChatID: chat.ID, UserID: userID, RoleInChat: "inquirer", UnreadCount: 0, JoinedAt: s.runtime.Now()},
			{ChatID: chat.ID, UserID: listing.OwnerUserID, RoleInChat: "publisher", UnreadCount: 0, JoinedAt: s.runtime.Now()},
		}
		return tx.Create(&participants).Error
	}); err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to create chat")
	}

	return map[string]any{"chat_id": chat.PublicID, "is_new": true}, nil
}

// 7. ListChats returns chats for the current user.
func (s *ChatService) ListChats(ctx context.Context, userID int64, page int, pageSize int) ([]ChatSummary, *model.Pagination, error) {
	page, pageSize = normalizePagination(page, pageSize)
	baseQuery := s.runtime.DB.WithContext(ctx).Table("chat_participants").
		Select("chat_participants.unread_count, chats.public_id AS chat_public_id, chats.biz_module, chats.chat_type, chats.last_message_preview, chats.last_message_at, listings.public_id AS listing_public_id, listings.title, listings.summary, listings.id AS listing_id, listings.business_status, listings.published_at, secondhand_listings.price_mode, secondhand_listings.price_hkd").
		Joins("JOIN chats ON chats.id = chat_participants.chat_id").
		Joins("LEFT JOIN listings ON listings.id = chats.listing_id").
		Joins("LEFT JOIN secondhand_listings ON secondhand_listings.listing_id = listings.id").
		Where("chat_participants.user_id = ? AND chats.chat_type <> ?", userID, chatTypeSystemNotice)

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, nil, errcode.New(errcode.CodeInternalError, "failed to count chats")
	}

	var rows []chatListRow
	if err := baseQuery.
		Order("chats.last_message_at desc NULLS LAST, chats.created_at desc").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&rows).Error; err != nil {
		return nil, nil, errcode.New(errcode.CodeInternalError, "failed to load chats")
	}

	imageMap, err := s.loadChatListingImages(ctx, extractListingIDs(rows))
	if err != nil {
		return nil, nil, err
	}
	peerMap, err := s.loadChatPeerMap(ctx, userID, extractChatIDs(rows))
	if err != nil {
		return nil, nil, err
	}

	items := make([]ChatSummary, 0, len(rows))
	for _, item := range rows {
		var lastMessageAt *string
		if item.LastMessageAt != nil {
			value := item.LastMessageAt.UTC().Format(time.RFC3339)
			lastMessageAt = &value
		}

		var cover *ListingImageResponse
		if item.ListingID > 0 {
			if images := imageMap[item.ListingID]; len(images) > 0 {
				cover = &images[0]
			}
		}

		listingSummary := buildChatListingSummary(item, cover)
		items = append(items, ChatSummary{
			ChatID:             item.ChatPublicID,
			ListingID:          item.ListingPublicID,
			ListingTitle:       item.Title,
			BizModule:          item.BizModule,
			ChatType:           item.ChatType,
			LastMessagePreview: item.LastMessagePreview,
			LastMessageAt:      lastMessageAt,
			UnreadCount:        item.UnreadCount,
			Peer:               peerMap[item.ChatPublicID],
			Listing:            listingSummary,
			CoverImage:         cover,
		})
	}

	return items, &model.Pagination{Page: page, PageSize: pageSize, Total: total}, nil
}

// 8. GetChat returns a single chat summary with participant data.
func (s *ChatService) GetChat(ctx context.Context, userID int64, chatPublicID string) (*ChatDetail, error) {
	chat, listing, _, err := s.loadAuthorizedChat(ctx, userID, chatPublicID)
	if err != nil {
		return nil, err
	}

	var participants []model.ChatParticipant
	if err := s.runtime.DB.WithContext(ctx).Where("chat_id = ?", chat.ID).Find(&participants).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load chat participants")
	}
	profileMap, err := s.secondhandService.loadUserPreviewMap(ctx, extractParticipantUserIDs(participants))
	if err != nil {
		return nil, err
	}
	peerMap, err := s.loadChatPeerMap(ctx, userID, []string{chat.PublicID})
	if err != nil {
		return nil, err
	}

	result := &ChatDetail{
		ChatID:       chat.PublicID,
		BizModule:    chat.BizModule,
		ChatType:     chat.ChatType,
		CreatedAt:    chat.CreatedAt.UTC().Format(time.RFC3339),
		Peer:         peerMap[chat.PublicID],
		Participants: make([]ChatMember, 0, len(participants)),
	}
	images, imageErr := s.loadChatListingImages(ctx, []int64{listing.ID})
	if imageErr != nil {
		return nil, imageErr
	}
	result.ListingID = listing.PublicID
	result.ListingTitle = listing.Title
	result.Listing = buildChatListingSummary(chatListRow{
		ListingPublicID: listing.PublicID,
		BizModule:       chat.BizModule,
		Title:           listing.Title,
		Summary:         listing.Summary,
		BusinessStatus:  listing.BusinessStatus,
		PublishedAt:     listing.PublishedAt,
	}, firstListingImage(images[listing.ID]))
	for _, participant := range participants {
		profile := profileMap[participant.UserID]
		result.Participants = append(result.Participants, ChatMember{
			UserID:      fmtInt64(participant.UserID),
			PublicID:    safeUserPublicID(profile),
			DisplayName: safeUserDisplayName(profile, fmtInt64(participant.UserID)),
			RoleInChat:  participant.RoleInChat,
		})
	}

	return result, nil
}

// 9. ListMessages returns messages for a chat the current user belongs to.
func (s *ChatService) ListMessages(ctx context.Context, userID int64, chatPublicID string, page int, pageSize int) ([]MessageResponse, *model.Pagination, error) {
	return s.listMessages(ctx, userID, chatPublicID, page, pageSize, false)
}

// 9.1 ListRecentMessages returns the most recent message page in chronological order.
func (s *ChatService) ListRecentMessages(ctx context.Context, userID int64, chatPublicID string, pageSize int) ([]MessageResponse, *model.Pagination, error) {
	return s.listMessages(ctx, userID, chatPublicID, 1, pageSize, true)
}

// 9.2 listMessages keeps legacy pagination while supporting the current chat tail.
func (s *ChatService) listMessages(ctx context.Context, userID int64, chatPublicID string, page int, pageSize int, newest bool) ([]MessageResponse, *model.Pagination, error) {
	chat, _, _, err := s.loadAuthorizedChat(ctx, userID, chatPublicID)
	if err != nil {
		return nil, nil, err
	}

	page, pageSize = normalizePagination(page, pageSize)
	baseQuery := s.runtime.DB.WithContext(ctx).Model(&model.Message{}).Where("chat_id = ?", chat.ID)

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, nil, errcode.New(errcode.CodeInternalError, "failed to count messages")
	}

	var messages []model.Message
	messageQuery := baseQuery
	if newest {
		messageQuery = baseQuery.Order("created_at desc, id desc").Limit(pageSize)
	} else {
		messageQuery = baseQuery.Order("created_at asc, id asc").Offset((page - 1) * pageSize).Limit(pageSize)
	}
	if err := messageQuery.Find(&messages).Error; err != nil {
		return nil, nil, errcode.New(errcode.CodeInternalError, "failed to load messages")
	}
	if newest {
		reverseMessages(messages)
	}

	items, err := s.buildMessageResponses(ctx, messages)
	if err != nil {
		return nil, nil, err
	}
	return items, &model.Pagination{Page: page, PageSize: pageSize, Total: total}, nil
}

// 9.3 reverseMessages restores chronological order after querying the latest page descending.
func reverseMessages(messages []model.Message) {
	for left, right := 0, len(messages)-1; left < right; left, right = left+1, right-1 {
		messages[left], messages[right] = messages[right], messages[left]
	}
}

// 10.1 ListMessagesAfter returns messages created after one known message.
func (s *ChatService) ListMessagesAfter(ctx context.Context, userID int64, chatPublicID string, afterMessageID string, pageSize int) ([]MessageResponse, *model.Pagination, error) {
	chat, _, _, err := s.loadAuthorizedChat(ctx, userID, chatPublicID)
	if err != nil {
		return nil, nil, err
	}
	_, pageSize = normalizePagination(1, pageSize)
	var anchor model.Message
	if err := s.runtime.DB.WithContext(ctx).Where("public_id = ? AND chat_id = ?", strings.TrimSpace(afterMessageID), chat.ID).First(&anchor).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &model.Pagination{Page: 1, PageSize: pageSize, Total: 0}, nil
		}
		return nil, nil, errcode.New(errcode.CodeInternalError, "failed to load message cursor")
	}
	query := s.runtime.DB.WithContext(ctx).Model(&model.Message{}).Where("chat_id = ? AND (created_at > ? OR (created_at = ? AND id > ?))", chat.ID, anchor.CreatedAt, anchor.CreatedAt, anchor.ID)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, nil, errcode.New(errcode.CodeInternalError, "failed to count messages")
	}
	var messages []model.Message
	if err := query.Order("created_at asc, id asc").Limit(pageSize).Find(&messages).Error; err != nil {
		return nil, nil, errcode.New(errcode.CodeInternalError, "failed to load messages")
	}
	items, err := s.buildMessageResponses(ctx, messages)
	if err != nil {
		return nil, nil, err
	}
	return items, &model.Pagination{Page: 1, PageSize: pageSize, Total: total}, nil
}

// 11. SendMessage creates a new text message in the target chat.
func (s *ChatService) SendMessage(ctx context.Context, userID int64, chatPublicID string, content string) (*MessageResponse, error) {
	return s.SendMessageWithAttachmentsAndClientID(ctx, userID, chatPublicID, content, nil, "")
}

// 11.1 SendMessageWithAttachments creates a text, attachment, or mixed message.
func (s *ChatService) SendMessageWithAttachments(ctx context.Context, userID int64, chatPublicID string, content string, attachmentIDs []string) (*MessageResponse, error) {
	return s.SendMessageWithAttachmentsAndClientID(ctx, userID, chatPublicID, content, attachmentIDs, "")
}

// 11.2 SendMessageWithAttachmentsAndClientID adds client retry idempotency.
func (s *ChatService) SendMessageWithAttachmentsAndClientID(ctx context.Context, userID int64, chatPublicID string, content string, attachmentIDs []string, clientMessageID string) (*MessageResponse, error) {
	content = strings.TrimSpace(content)
	var attachmentErr error
	attachmentIDs, attachmentErr = normalizeAttachmentIDs(attachmentIDs)
	if attachmentErr != nil {
		return nil, attachmentErr
	}
	clientMessageID = strings.TrimSpace(clientMessageID)
	if len([]rune(clientMessageID)) > 80 {
		return nil, errcode.New(errcode.CodeValidationError, "client_message_id is too long")
	}
	if content == "" && len(attachmentIDs) == 0 {
		return nil, errcode.New(errcode.CodeValidationError, "message content is required")
	}
	if len([]rune(content)) > messageContentMaxRunes {
		return nil, errcode.New(errcode.CodeValidationError, "message content is too long")
	}

	chat, listing, participant, err := s.loadAuthorizedChat(ctx, userID, chatPublicID)
	if err != nil {
		return nil, err
	}
	if listing.PublicationStatus != "active" || listing.ModerationStatus != "approved" || listing.BusinessStatus != "available" {
		return nil, errcode.New(errcode.CodeAuthForbidden, "listing is not available for messaging")
	}
	if clientMessageID != "" {
		var existing model.Message
		if err := s.runtime.DB.WithContext(ctx).Where("chat_id = ? AND sender_user_id = ? AND client_message_id = ?", chat.ID, userID, clientMessageID).First(&existing).Error; err == nil {
			items, buildErr := s.buildMessageResponses(ctx, []model.Message{existing})
			if buildErr != nil {
				return nil, buildErr
			}
			return &items[0], nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.New(errcode.CodeInternalError, "failed to load message retry")
		}
	}
	assets, err := s.loadChatAttachmentAssets(ctx, userID, attachmentIDs)
	if err != nil {
		return nil, err
	}
	messageType := messageTypeText
	if len(assets) > 0 && content == "" {
		messageType = messageTypeAttachment
	} else if len(assets) > 0 {
		messageType = messageTypeMixed
	}
	preview := content
	if preview == "" {
		preview = "[Attachment]"
	}

	message := model.Message{
		PublicID:        utils.NewPublicID(),
		ChatID:          chat.ID,
		SenderUserID:    userID,
		ClientMessageID: optionalString(clientMessageID),
		MessageType:     messageType,
		ContentText:     content,
		MessageStatus:   "sent",
		CreatedAt:       s.runtime.Now(),
	}
	eventPayload, payloadErr := buildRealtimeMessagePayload(chat.PublicID, messageResponseFromAssets(message, assets, s.runtime.Config.MediaBaseURL), clientMessageID)
	if payloadErr != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to prepare realtime event")
	}
	outboxEnabled := s.runtime.DB.Migrator().HasTable(&model.RealtimeOutbox{})

	if err := s.runtime.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		assets, err = s.loadChatAttachmentAssetsWithDB(ctx, tx, userID, attachmentIDs, true)
		if err != nil {
			return err
		}
		eventPayload, payloadErr = buildRealtimeMessagePayload(chat.PublicID, messageResponseFromAssets(message, assets, s.runtime.Config.MediaBaseURL), clientMessageID)
		if payloadErr != nil {
			return payloadErr
		}
		if err := tx.Create(&message).Error; err != nil {
			return err
		}
		for index, asset := range assets {
			if err := tx.Create(&model.MessageAttachment{MessageID: message.ID, MediaAssetID: asset.ID, SortOrder: index}).Error; err != nil {
				return err
			}
		}
		if outboxEnabled {
			if err := tx.Create(&model.RealtimeOutbox{EventType: "message", AggregateID: message.PublicID, Payload: eventPayload, Status: "pending", NextAttemptAt: message.CreatedAt, CreatedAt: message.CreatedAt}).Error; err != nil {
				return err
			}
		}

		now := s.runtime.Now()
		if err := tx.Model(&model.Chat{}).Where("id = ?", chat.ID).Updates(map[string]any{
			"last_message_preview": preview,
			"last_message_at":      now,
			"updated_at":           now,
		}).Error; err != nil {
			return err
		}

		if err := tx.Model(&model.ChatParticipant{}).Where("chat_id = ? AND user_id <> ?", chat.ID, userID).Update("unread_count", gorm.Expr("unread_count + ?", 1)).Error; err != nil {
			return err
		}

		var recipients []model.ChatParticipant
		if err := tx.WithContext(ctx).Where("chat_id = ? AND user_id <> ?", chat.ID, userID).Find(&recipients).Error; err != nil {
			return err
		}
		notificationService := NewNotificationService(s.runtime)
		for _, recipient := range recipients {
			if err := notificationService.CreateNotification(ctx, tx, CreateNotificationParams{
				UserID:          recipient.UserID,
				Category:        "chat_message",
				Title:           "新訊息",
				Body:            preview,
				RelatedType:     "chat",
				RelatedPublicID: chat.PublicID,
			}); err != nil {
				return err
			}
		}

		return tx.Model(&model.ChatParticipant{}).Where("id = ?", participant.ID).Updates(map[string]any{
			"last_read_message_id": message.ID,
			"unread_count":         0,
		}).Error
	}); err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to send message")
	}
	items, err := s.buildMessageResponses(ctx, []model.Message{message})
	if err != nil {
		return nil, err
	}
	return &items[0], nil
}

// 11. MarkRead clears unread count for the current participant.
func (s *ChatService) MarkRead(ctx context.Context, userID int64, chatPublicID string) error {
	chat, _, participant, err := s.loadAuthorizedChat(ctx, userID, chatPublicID)
	if err != nil {
		return err
	}

	var latestMessage model.Message
	err = s.runtime.DB.WithContext(ctx).Where("chat_id = ?", chat.ID).Order("created_at desc").First(&latestMessage).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return errcode.New(errcode.CodeInternalError, "failed to load latest message")
	}

	updates := map[string]any{"unread_count": 0}
	if latestMessage.ID > 0 {
		updates["last_read_message_id"] = latestMessage.ID
	}

	if err := s.runtime.DB.WithContext(ctx).Model(&model.ChatParticipant{}).Where("id = ?", participant.ID).Updates(updates).Error; err != nil {
		return errcode.New(errcode.CodeInternalError, "failed to mark chat as read")
	}

	return nil
}

// 12. ValidateChatAccess validates a realtime connection's chat scope.
func (s *ChatService) ValidateChatAccess(ctx context.Context, userID int64, chatPublicID string) error {
	var chat model.Chat
	if err := s.runtime.DB.WithContext(ctx).Where("public_id = ?", strings.TrimSpace(chatPublicID)).First(&chat).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.New(errcode.CodeNotFound, "chat not found")
		}
		return errcode.New(errcode.CodeInternalError, "failed to load chat")
	}
	if chat.ChatType == chatTypeBuildingGroup {
		_, _, err := s.loadAuthorizedBuildingChat(ctx, userID, chat.PublicID, true)
		return err
	}
	_, _, _, err := s.loadAuthorizedChat(ctx, userID, chat.PublicID)
	return err
}

// 13. SendRealtimeMessage routes one WebSocket text command through the existing chat services.
func (s *ChatService) SendRealtimeMessage(ctx context.Context, userID int64, chatPublicID string, content string) (*MessageResponse, error) {
	return s.SendRealtimeMessageWithAttachmentsAndClientID(ctx, userID, chatPublicID, content, nil, "")
}

// 13.1 SendRealtimeMessageWithAttachments routes a WebSocket message with uploads.
func (s *ChatService) SendRealtimeMessageWithAttachments(ctx context.Context, userID int64, chatPublicID string, content string, attachmentIDs []string) (*MessageResponse, error) {
	return s.SendRealtimeMessageWithAttachmentsAndClientID(ctx, userID, chatPublicID, content, attachmentIDs, "")
}

// 14.2 SendRealtimeMessageWithAttachmentsAndClientID routes an idempotent WebSocket message.
func (s *ChatService) SendRealtimeMessageWithAttachmentsAndClientID(ctx context.Context, userID int64, chatPublicID string, content string, attachmentIDs []string, clientMessageID string) (*MessageResponse, error) {
	var chat model.Chat
	if err := s.runtime.DB.WithContext(ctx).Where("public_id = ?", strings.TrimSpace(chatPublicID)).First(&chat).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.New(errcode.CodeNotFound, "chat not found")
		}
		return nil, errcode.New(errcode.CodeInternalError, "failed to load chat")
	}
	if chat.ChatType == chatTypeBuildingGroup {
		return s.SendBuildingChatMessageWithAttachmentsAndClientID(ctx, userID, chat.PublicID, content, attachmentIDs, clientMessageID)
	}
	return s.SendMessageWithAttachmentsAndClientID(ctx, userID, chat.PublicID, content, attachmentIDs, clientMessageID)
}

// 14. loadAuthorizedChat loads a chat and validates membership.
func (s *ChatService) loadAuthorizedChat(ctx context.Context, userID int64, chatPublicID string) (*model.Chat, *model.Listing, *model.ChatParticipant, error) {
	var chat model.Chat
	if err := s.runtime.DB.WithContext(ctx).Where("public_id = ?", chatPublicID).First(&chat).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, nil, errcode.New(errcode.CodeNotFound, "chat not found")
		}
		return nil, nil, nil, errcode.New(errcode.CodeInternalError, "failed to load chat")
	}
	if chat.ChatType == chatTypeSystemNotice {
		return nil, nil, nil, errcode.New(errcode.CodeNotFound, "chat not found")
	}

	var participant model.ChatParticipant
	if err := s.runtime.DB.WithContext(ctx).Where("chat_id = ? AND user_id = ?", chat.ID, userID).First(&participant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, nil, errcode.New(errcode.CodeAuthForbidden, "chat is not accessible")
		}
		return nil, nil, nil, errcode.New(errcode.CodeInternalError, "failed to load chat participant")
	}

	var listing model.Listing
	if err := s.runtime.DB.WithContext(ctx).First(&listing, chat.ListingID).Error; err != nil {
		return nil, nil, nil, errcode.New(errcode.CodeInternalError, "failed to load chat listing")
	}

	return &chat, &listing, &participant, nil
}

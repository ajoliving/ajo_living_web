/*
 * Chat business logic.
 * 1. Create or reuse listing-scoped chats.
 * 2. Persist messages and unread state.
 * 3. Enforce participant and listing access checks.
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

// 1. ChatService handles listing chats and messages.
type ChatService struct {
	runtime           *Runtime
	secondhandService *SecondhandService
}

// 2. chatListRow defines the internal chat list query row.
type chatListRow struct {
	UnreadCount        int
	ChatPublicID       string
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
func NewChatService(runtime *Runtime, secondhandService *SecondhandService) *ChatService {
	return &ChatService{
		runtime:           runtime,
		secondhandService: secondhandService,
	}
}

// 4. CreateOrReuseChat creates or reuses a listing chat for the current user.
func (s *ChatService) CreateOrReuseChat(ctx context.Context, userID int64, communityID *int64, listingPublicID string) (map[string]any, error) {
	listing, secondhand, _, err := s.secondhandService.loadListingByPublicID(ctx, listingPublicID)
	if err != nil {
		return nil, err
	}
	if listing.OwnerUserID == userID {
		return nil, errcode.New(errcode.CodeValidationError, "seller cannot start a chat with the same listing")
	}
	if listing.PublicationStatus != "active" || listing.BusinessStatus != "available" {
		return nil, errcode.New(errcode.CodeAuthForbidden, "listing is not available for chat")
	}
	if !s.secondhandService.canViewListing(listing, secondhand, communityID) {
		return nil, errcode.New(errcode.CodeVisibilityForbidden, "listing is not visible to the current user")
	}

	var chat model.Chat
	err = s.runtime.DB.WithContext(ctx).Where("listing_id = ? AND created_by = ?", listing.ID, userID).First(&chat).Error
	if err == nil {
		return map[string]any{"chat_id": chat.PublicID, "is_new": false}, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load chat")
	}

	chat = model.Chat{
		PublicID:  utils.NewPublicID(),
		BizModule: "secondhand",
		ListingID: listing.ID,
		ChatType:  "direct_listing_chat",
		CreatedBy: userID,
		CreatedAt: s.runtime.Now(),
		UpdatedAt: s.runtime.Now(),
	}

	if err := s.runtime.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&chat).Error; err != nil {
			return err
		}

		participants := []model.ChatParticipant{
			{ChatID: chat.ID, UserID: userID, RoleInChat: "buyer", UnreadCount: 0, JoinedAt: s.runtime.Now()},
			{ChatID: chat.ID, UserID: listing.OwnerUserID, RoleInChat: "owner", UnreadCount: 0, JoinedAt: s.runtime.Now()},
		}
		return tx.Create(&participants).Error
	}); err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to create chat")
	}

	return map[string]any{"chat_id": chat.PublicID, "is_new": true}, nil
}

// 5. ListChats returns chats for the current user.
func (s *ChatService) ListChats(ctx context.Context, userID int64, page int, pageSize int) ([]ChatSummary, *model.Pagination, error) {
	page, pageSize = normalizePagination(page, pageSize)
	baseQuery := s.runtime.DB.WithContext(ctx).Table("chat_participants").
		Select("chat_participants.unread_count, chats.public_id AS chat_public_id, chats.last_message_preview, chats.last_message_at, listings.public_id AS listing_public_id, listings.title, listings.summary, listings.id AS listing_id, listings.business_status, listings.published_at, secondhand_listings.price_mode, secondhand_listings.price_hkd").
		Joins("JOIN chats ON chats.id = chat_participants.chat_id").
		Joins("JOIN listings ON listings.id = chats.listing_id").
		Joins("JOIN secondhand_listings ON secondhand_listings.listing_id = listings.id").
		Where("chat_participants.user_id = ?", userID)

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, nil, errcode.New(errcode.CodeInternalError, "failed to count chats")
	}

	var rows []chatListRow
	if err := baseQuery.Order("chats.last_message_at desc NULLS LAST, chats.created_at desc").Offset((page - 1) * pageSize).Limit(pageSize).Scan(&rows).Error; err != nil {
		return nil, nil, errcode.New(errcode.CodeInternalError, "failed to load chats")
	}

	imageMap, err := s.secondhandService.loadListingImages(ctx, extractListingIDs(rows))
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
		if images := imageMap[item.ListingID]; len(images) > 0 {
			cover = &images[0]
		}

		listingSummary := buildChatListingSummary(item, cover)
		items = append(items, ChatSummary{
			ChatID:             item.ChatPublicID,
			ListingID:          item.ListingPublicID,
			ListingTitle:       item.Title,
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

// 6. GetChat returns a single chat summary with participant data.
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
	images, err := s.secondhandService.loadListingImages(ctx, []int64{listing.ID})
	if err != nil {
		return nil, err
	}
	peerMap, err := s.loadChatPeerMap(ctx, userID, []string{chat.PublicID})
	if err != nil {
		return nil, err
	}

	result := &ChatDetail{
		ChatID:       chat.PublicID,
		ListingID:    listing.PublicID,
		ListingTitle: listing.Title,
		CreatedAt:    chat.CreatedAt.UTC().Format(time.RFC3339),
		Peer:         peerMap[chat.PublicID],
		Listing: buildChatListingSummary(chatListRow{
			ListingPublicID: listing.PublicID,
			Title:           listing.Title,
			Summary:         listing.Summary,
			BusinessStatus:  listing.BusinessStatus,
			PublishedAt:     listing.PublishedAt,
		}, firstListingImage(images[listing.ID])),
		Participants: make([]ChatMember, 0, len(participants)),
	}
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

// 7. ListMessages returns messages for a chat the current user belongs to.
func (s *ChatService) ListMessages(ctx context.Context, userID int64, chatPublicID string, page int, pageSize int) ([]MessageResponse, *model.Pagination, error) {
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
	if err := baseQuery.Order("created_at asc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&messages).Error; err != nil {
		return nil, nil, errcode.New(errcode.CodeInternalError, "failed to load messages")
	}

	items := make([]MessageResponse, 0, len(messages))
	for _, message := range messages {
		items = append(items, MessageResponse{
			MessageID:    message.PublicID,
			SenderUserID: fmtInt64(message.SenderUserID),
			Content:      message.ContentText,
			MessageType:  message.MessageType,
			Status:       message.MessageStatus,
			CreatedAt:    message.CreatedAt.UTC().Format(time.RFC3339),
		})
	}

	return items, &model.Pagination{Page: page, PageSize: pageSize, Total: total}, nil
}

// 8. SendMessage creates a new text message in the target chat.
func (s *ChatService) SendMessage(ctx context.Context, userID int64, chatPublicID string, content string) (*MessageResponse, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, errcode.New(errcode.CodeValidationError, "message content is required")
	}

	chat, listing, participant, err := s.loadAuthorizedChat(ctx, userID, chatPublicID)
	if err != nil {
		return nil, err
	}
	if listing.PublicationStatus != "active" || listing.BusinessStatus != "available" {
		return nil, errcode.New(errcode.CodeAuthForbidden, "listing is not available for messaging")
	}

	message := model.Message{
		PublicID:      utils.NewPublicID(),
		ChatID:        chat.ID,
		SenderUserID:  userID,
		MessageType:   "text",
		ContentText:   content,
		MessageStatus: "sent",
		CreatedAt:     s.runtime.Now(),
	}

	if err := s.runtime.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&message).Error; err != nil {
			return err
		}

		now := s.runtime.Now()
		if err := tx.Model(&model.Chat{}).Where("id = ?", chat.ID).Updates(map[string]any{
			"last_message_preview": content,
			"last_message_at":      now,
			"updated_at":           now,
		}).Error; err != nil {
			return err
		}

		if err := tx.Model(&model.ChatParticipant{}).Where("chat_id = ? AND user_id <> ?", chat.ID, userID).Update("unread_count", gorm.Expr("unread_count + ?", 1)).Error; err != nil {
			return err
		}

		if err := NewNotificationService(s.runtime).CreateNotification(ctx, tx, CreateNotificationParams{
			UserID:          listing.OwnerUserID,
			Category:        "chat_message",
			Title:           "New chat message",
			Body:            content,
			RelatedType:     "chat",
			RelatedPublicID: chat.PublicID,
		}); err != nil && listing.OwnerUserID != userID {
			return err
		}

		if listing.OwnerUserID == userID {
			if err := NewNotificationService(s.runtime).CreateNotification(ctx, tx, CreateNotificationParams{
				UserID:          chat.CreatedBy,
				Category:        "chat_message",
				Title:           "New chat message",
				Body:            content,
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

	return &MessageResponse{
		MessageID:    message.PublicID,
		SenderUserID: fmtInt64(message.SenderUserID),
		Content:      message.ContentText,
		MessageType:  message.MessageType,
		Status:       message.MessageStatus,
		CreatedAt:    message.CreatedAt.UTC().Format(time.RFC3339),
	}, nil
}

// 9. MarkRead clears unread count for the current participant.
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

// 10. loadAuthorizedChat loads a chat and validates membership.
func (s *ChatService) loadAuthorizedChat(ctx context.Context, userID int64, chatPublicID string) (*model.Chat, *model.Listing, *model.ChatParticipant, error) {
	var chat model.Chat
	if err := s.runtime.DB.WithContext(ctx).Where("public_id = ?", chatPublicID).First(&chat).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, nil, errcode.New(errcode.CodeNotFound, "chat not found")
		}
		return nil, nil, nil, errcode.New(errcode.CodeInternalError, "failed to load chat")
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

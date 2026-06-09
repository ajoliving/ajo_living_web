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

const (
	chatTypeDirectListing = "direct_listing_chat"
	chatTypeSystemNotice  = "system_notice"
	messageTypeText       = "text"
	messageTypeNoticeCard = "notice_card"
	systemNoticeListingID = 0
)

type defaultNoticeMessage struct {
	Content     string
	ActionLabel string
	ActionURL   string
	CreatedAt   time.Time
}

// 1. ChatService handles listing chats and messages.
type ChatService struct {
	runtime           *Runtime
	secondhandService *SecondhandService
}

// 2. chatListRow defines the internal chat list query row.
type chatListRow struct {
	UnreadCount        int
	ChatPublicID       string
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
	if _, err := s.ensureSystemNoticeChat(ctx, userID); err != nil {
		return nil, nil, err
	}

	page, pageSize = normalizePagination(page, pageSize)
	baseQuery := s.runtime.DB.WithContext(ctx).Table("chat_participants").
		Select("chat_participants.unread_count, chats.public_id AS chat_public_id, chats.chat_type, chats.last_message_preview, chats.last_message_at, listings.public_id AS listing_public_id, listings.title, listings.summary, listings.id AS listing_id, listings.business_status, listings.published_at, secondhand_listings.price_mode, secondhand_listings.price_hkd").
		Joins("JOIN chats ON chats.id = chat_participants.chat_id").
		Joins("LEFT JOIN listings ON listings.id = chats.listing_id").
		Joins("LEFT JOIN secondhand_listings ON secondhand_listings.listing_id = listings.id").
		Where("chat_participants.user_id = ?", userID)

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, nil, errcode.New(errcode.CodeInternalError, "failed to count chats")
	}

	var rows []chatListRow
	if err := baseQuery.
		Order("CASE WHEN chats.chat_type = 'system_notice' THEN 0 ELSE 1 END ASC").
		Order("chats.last_message_at desc NULLS LAST, chats.created_at desc").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&rows).Error; err != nil {
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
		if item.ListingID > 0 {
			if images := imageMap[item.ListingID]; len(images) > 0 {
				cover = &images[0]
			}
		}

		var listingSummary *ChatListingSummary
		if item.ChatType != chatTypeSystemNotice {
			listingSummary = buildChatListingSummary(item, cover)
		}
		if item.ChatType == chatTypeSystemNotice {
			item.Title = model.SystemNotificationDisplayName
		}
		items = append(items, ChatSummary{
			ChatID:             item.ChatPublicID,
			ListingID:          item.ListingPublicID,
			ListingTitle:       item.Title,
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
	peerMap, err := s.loadChatPeerMap(ctx, userID, []string{chat.PublicID})
	if err != nil {
		return nil, err
	}

	result := &ChatDetail{
		ChatID:       chat.PublicID,
		ChatType:     chat.ChatType,
		CreatedAt:    chat.CreatedAt.UTC().Format(time.RFC3339),
		Peer:         peerMap[chat.PublicID],
		Participants: make([]ChatMember, 0, len(participants)),
	}
	if chat.ChatType != chatTypeSystemNotice {
		images, imageErr := s.secondhandService.loadListingImages(ctx, []int64{listing.ID})
		if imageErr != nil {
			return nil, imageErr
		}
		result.ListingID = listing.PublicID
		result.ListingTitle = listing.Title
		result.Listing = buildChatListingSummary(chatListRow{
			ListingPublicID: listing.PublicID,
			Title:           listing.Title,
			Summary:         listing.Summary,
			BusinessStatus:  listing.BusinessStatus,
			PublishedAt:     listing.PublishedAt,
		}, firstListingImage(images[listing.ID]))
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
			ActionLabel:  message.ActionLabel,
			ActionURL:    message.ActionURL,
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
	if chat.ChatType == chatTypeSystemNotice {
		return nil, errcode.New(errcode.CodeAuthForbidden, "system notice chat is read-only")
	}
	if listing.PublicationStatus != "active" || listing.BusinessStatus != "available" {
		return nil, errcode.New(errcode.CodeAuthForbidden, "listing is not available for messaging")
	}

	message := model.Message{
		PublicID:      utils.NewPublicID(),
		ChatID:        chat.ID,
		SenderUserID:  userID,
		MessageType:   messageTypeText,
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

		var recipients []model.ChatParticipant
		if err := tx.WithContext(ctx).Where("chat_id = ? AND user_id <> ?", chat.ID, userID).Find(&recipients).Error; err != nil {
			return err
		}
		notificationService := NewNotificationService(s.runtime)
		for _, recipient := range recipients {
			if err := notificationService.CreateNotification(ctx, tx, CreateNotificationParams{
				UserID:          recipient.UserID,
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

// 9. PublishSystemNotice sends one read-only notice card to every active member.
func (s *ChatService) PublishSystemNotice(ctx context.Context, params SystemNoticePublishParams) (*SystemNoticePublishResult, error) {
	title := strings.TrimSpace(params.Title)
	body := strings.TrimSpace(params.Body)
	actionLabel := strings.TrimSpace(params.ActionLabel)
	actionURL := strings.TrimSpace(params.ActionURL)
	if title == "" || body == "" {
		return nil, errcode.New(errcode.CodeValidationError, "notice title and body are required")
	}
	if len([]rune(title)) > 120 || len([]rune(body)) > 1000 || len([]rune(actionLabel)) > 80 || len([]rune(actionURL)) > 500 {
		return nil, errcode.New(errcode.CodeValidationError, "notice content is too long")
	}
	if (actionLabel == "") != (actionURL == "") {
		return nil, errcode.New(errcode.CodeValidationError, "notice action label and url must be provided together")
	}

	systemUser, err := s.loadSystemNotificationUser(ctx)
	if err != nil {
		return nil, err
	}

	var users []model.User
	if err := s.runtime.DB.WithContext(ctx).
		Where("member_status = ? AND id <> ?", "active", systemUser.ID).
		Find(&users).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load notice recipients")
	}

	delivered := 0
	noticeContent := title + "\n" + body
	if err := s.runtime.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, user := range users {
			chat, err := s.ensureSystemNoticeChatWithDB(ctx, tx, user.ID, systemUser)
			if err != nil {
				return err
			}
			if chat == nil {
				continue
			}

			now := s.runtime.Now()
			message := model.Message{
				PublicID:      utils.NewPublicID(),
				ChatID:        chat.ID,
				SenderUserID:  systemUser.ID,
				MessageType:   messageTypeNoticeCard,
				ContentText:   noticeContent,
				ActionLabel:   actionLabel,
				ActionURL:     actionURL,
				MessageStatus: "sent",
				CreatedAt:     now,
			}
			if err := tx.Create(&message).Error; err != nil {
				return err
			}
			if err := tx.Model(&model.Chat{}).Where("id = ?", chat.ID).Updates(map[string]any{
				"last_message_preview": body,
				"last_message_at":      now,
				"updated_at":           now,
			}).Error; err != nil {
				return err
			}
			if err := tx.Model(&model.ChatParticipant{}).
				Where("chat_id = ? AND user_id = ?", chat.ID, user.ID).
				Update("unread_count", gorm.Expr("unread_count + ?", 1)).
				Error; err != nil {
				return err
			}

			delivered++
		}

		return nil
	}); err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to publish system notice")
	}

	return &SystemNoticePublishResult{DeliveredCount: delivered}, nil
}

// 10. MarkRead clears unread count for the current participant.
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

// 11. loadAuthorizedChat loads a chat and validates membership.
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
	if chat.ChatType != chatTypeSystemNotice {
		if err := s.runtime.DB.WithContext(ctx).First(&listing, chat.ListingID).Error; err != nil {
			return nil, nil, nil, errcode.New(errcode.CodeInternalError, "failed to load chat listing")
		}
	}

	return &chat, &listing, &participant, nil
}

// 12. ensureSystemNoticeChat creates a read-only system notice chat for a user.
func (s *ChatService) ensureSystemNoticeChat(ctx context.Context, userID int64) (*model.Chat, error) {
	systemUser, err := s.loadSystemNotificationUser(ctx)
	if err != nil {
		return nil, err
	}

	return s.ensureSystemNoticeChatWithDB(ctx, s.runtime.DB, userID, systemUser)
}

// 13. ensureSystemNoticeChatWithDB creates a system notice chat with the provided DB handle.
func (s *ChatService) ensureSystemNoticeChatWithDB(ctx context.Context, db *gorm.DB, userID int64, systemUser *model.User) (*model.Chat, error) {
	if systemUser.ID == userID {
		return nil, nil
	}

	var chat model.Chat
	err := db.WithContext(ctx).
		Where("chat_type = ? AND created_by = ?", chatTypeSystemNotice, userID).
		First(&chat).
		Error
	if err == nil {
		return &chat, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load system notice chat")
	}

	now := s.runtime.Now()
	chat = model.Chat{
		PublicID:           utils.NewPublicID(),
		BizModule:          "system",
		ListingID:          systemNoticeListingID,
		ChatType:           chatTypeSystemNotice,
		CreatedBy:          userID,
		LastMessagePreview: "最高100幣，可以當錢花",
		LastMessageAt:      &now,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	if err := db.WithContext(ctx).Create(&chat).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to create system notice chat")
	}

	participants := []model.ChatParticipant{
		{ChatID: chat.ID, UserID: userID, RoleInChat: "recipient", UnreadCount: len(defaultSystemNoticeMessages(now)), JoinedAt: now},
		{ChatID: chat.ID, UserID: systemUser.ID, RoleInChat: "system", UnreadCount: 0, JoinedAt: now},
	}
	if err := db.WithContext(ctx).Create(&participants).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to create system notice chat")
	}

	messages := make([]model.Message, 0, len(defaultSystemNoticeMessages(now)))
	for _, notice := range defaultSystemNoticeMessages(now) {
		messages = append(messages, model.Message{
			PublicID:      utils.NewPublicID(),
			ChatID:        chat.ID,
			SenderUserID:  systemUser.ID,
			MessageType:   messageTypeNoticeCard,
			ContentText:   notice.Content,
			ActionLabel:   notice.ActionLabel,
			ActionURL:     notice.ActionURL,
			MessageStatus: "sent",
			CreatedAt:     notice.CreatedAt,
		})
	}
	if err := db.WithContext(ctx).Create(&messages).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to create system notice chat")
	}

	return &chat, nil
}

// 14. loadSystemNotificationUser returns the built-in notification sender.
func (s *ChatService) loadSystemNotificationUser(ctx context.Context) (*model.User, error) {
	var user model.User
	if err := s.runtime.DB.WithContext(ctx).
		Where("phone_country_code = ? AND phone_number = ?", model.SystemNotificationPhoneCountryCode, model.SystemNotificationPhoneNumber).
		First(&user).
		Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load system notification user")
	}

	return &user, nil
}

// 15. defaultSystemNoticeMessages returns initial read-only announcement cards.
func defaultSystemNoticeMessages(now time.Time) []defaultNoticeMessage {
	return []defaultNoticeMessage{
		{
			Content:     "家具精選更新\n查看近期上架家具與家電",
			ActionLabel: "查看家具",
			ActionURL:   "/furniture",
			CreatedAt:   now.Add(-5 * time.Minute),
		},
		{
			Content:     "樓盤租售更新\n瀏覽最新可租售樓盤",
			ActionLabel: "查看樓盤",
			ActionURL:   "/properties",
			CreatedAt:   now.Add(-4 * time.Minute),
		},
		{
			Content:     "會員中心提醒\n查看你的收藏、發布與聊天",
			ActionLabel: "前往會員中心",
			ActionURL:   "/member",
			CreatedAt:   now.Add(-3 * time.Minute),
		},
		{
			Content:     "服務式住宅更新\n查看最新服務式住宅資訊",
			ActionLabel: "查看住宅",
			ActionURL:   "/serviced-residences",
			CreatedAt:   now.Add(-2 * time.Minute),
		},
	}
}

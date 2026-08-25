/*
 * Building group chat business logic.
 * 1. Resolve building membership from local profile and active authorizations.
 * 2. Keep one group conversation for each building and persist member state.
 * 3. Handle join requests and scoped moderation with an audit trail.
 */
package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/model"
	"ajoliving_web/http_service/internal/utils"
)

const (
	chatTypeBuildingGroup      = "building_group"
	chatMemberActive           = "active"
	chatMemberLeft             = "left"
	chatMemberKicked           = "kicked"
	chatMemberBanned           = "banned"
	chatJoinRequestPending     = "pending"
	chatJoinRequestApproved    = "approved"
	chatJoinRequestRejected    = "rejected"
	chatModerationTargetMember = "chat_member"
	chatModerationMute         = "mute"
	chatModerationUnmute       = "unmute"
	chatModerationKick         = "kick"
	chatModerationBan          = "ban"
	chatModerationUnban        = "unban"
)

// 1. BuildingChatSummary defines the building group list payload.
type BuildingChatSummary struct {
	ChatID        string  `json:"chat_id"`
	BuildingID    string  `json:"building_id"`
	ChatType      string  `json:"chat_type"`
	MemberCount   int64   `json:"member_count"`
	UnreadCount   int     `json:"unread_count"`
	LastPreview   string  `json:"last_message_preview"`
	LastMessageAt *string `json:"last_message_at,omitempty"`
}

// 2. BuildingChatMember defines a member management payload.
type BuildingChatMember struct {
	UserID           string  `json:"user_id"`
	PublicID         string  `json:"public_id"`
	DisplayName      string  `json:"display_name"`
	RoleInChat       string  `json:"role_in_chat"`
	MembershipStatus string  `json:"membership_status"`
	MutedUntil       *string `json:"muted_until,omitempty"`
	BannedUntil      *string `json:"banned_until,omitempty"`
	JoinedAt         string  `json:"joined_at"`
}

// 3. BuildingChatJoinRequest defines an approval queue payload.
type BuildingChatJoinRequest struct {
	RequestID  string  `json:"request_id"`
	ChatID     string  `json:"chat_id"`
	UserID     string  `json:"user_id"`
	Status     string  `json:"status"`
	Reason     string  `json:"reason,omitempty"`
	CreatedAt  string  `json:"created_at"`
	ReviewedAt *string `json:"reviewed_at,omitempty"`
}

// 4. ListBuildingChats returns one group for every building visible to a member.
func (s *ChatService) ListBuildingChats(ctx context.Context, userID int64) ([]BuildingChatSummary, error) {
	buildingIDs, err := s.visibleBuildingIDs(ctx, userID)
	if err != nil {
		return nil, err
	}
	items := make([]BuildingChatSummary, 0, len(buildingIDs))
	for _, buildingID := range buildingIDs {
		chat, participant, err := s.ensureBuildingChat(ctx, userID, buildingID)
		if err != nil {
			continue
		}
		items = append(items, s.buildBuildingChatSummary(ctx, chat, participant))
	}
	return items, nil
}

// 5. EnsureBuildingChat creates the building group on first access.
func (s *ChatService) EnsureBuildingChat(ctx context.Context, userID int64, buildingID string) (*BuildingChatSummary, error) {
	chat, participant, err := s.ensureBuildingChat(ctx, userID, buildingID)
	if err != nil {
		return nil, err
	}
	result := s.buildBuildingChatSummary(ctx, chat, participant)
	return &result, nil
}

// 6. GetBuildingChat returns an authorized group detail.
func (s *ChatService) GetBuildingChat(ctx context.Context, userID int64, chatPublicID string) (*ChatDetail, error) {
	chat, participant, err := s.loadAuthorizedBuildingChat(ctx, userID, chatPublicID, true)
	if err != nil {
		return nil, err
	}
	var members []model.ChatParticipant
	if err := s.runtime.DB.WithContext(ctx).Where("chat_id = ? AND membership_status <> ?", chat.ID, chatMemberLeft).Order("joined_at asc").Find(&members).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load building chat members")
	}
	result := &ChatDetail{ChatID: chat.PublicID, BuildingID: chat.BuildingID, BizModule: chat.BizModule, ChatType: chat.ChatType, CreatedAt: chat.CreatedAt.UTC().Format(time.RFC3339), CanManage: s.canManageBuilding(ctx, userID, chat), Participants: make([]ChatMember, 0, len(members))}
	for _, member := range members {
		preview := s.loadUserPreview(ctx, member.UserID)
		result.Participants = append(result.Participants, ChatMember{UserID: fmtInt64(member.UserID), PublicID: preview.PublicID, DisplayName: preview.DisplayName, RoleInChat: member.RoleInChat, MembershipStatus: member.MembershipStatus})
	}
	if participant != nil {
		result.UnreadCount = participant.UnreadCount
	}
	return result, nil
}

// 7. ListBuildingChatMessages returns group messages.
func (s *ChatService) ListBuildingChatMessages(ctx context.Context, userID int64, chatPublicID string, page int, pageSize int) ([]MessageResponse, *model.Pagination, error) {
	chat, _, err := s.loadAuthorizedBuildingChat(ctx, userID, chatPublicID, true)
	if err != nil {
		return nil, nil, err
	}
	page, pageSize = normalizePagination(page, pageSize)
	query := s.runtime.DB.WithContext(ctx).Model(&model.Message{}).Where("chat_id = ?", chat.ID)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, nil, errcode.New(errcode.CodeInternalError, "failed to count messages")
	}
	var messages []model.Message
	if err := query.Order("created_at asc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&messages).Error; err != nil {
		return nil, nil, errcode.New(errcode.CodeInternalError, "failed to load messages")
	}
	items := make([]MessageResponse, 0, len(messages))
	for _, message := range messages {
		items = append(items, MessageResponse{MessageID: message.PublicID, SenderUserID: fmtInt64(message.SenderUserID), Content: message.ContentText, MessageType: message.MessageType, Status: message.MessageStatus, CreatedAt: message.CreatedAt.UTC().Format(time.RFC3339)})
	}
	return items, &model.Pagination{Page: page, PageSize: pageSize, Total: total}, nil
}

// 8. SendBuildingChatMessage persists a text message after group moderation checks.
func (s *ChatService) SendBuildingChatMessage(ctx context.Context, userID int64, chatPublicID string, content string) (*MessageResponse, error) {
	content = strings.TrimSpace(content)
	if content == "" || len([]rune(content)) > messageContentMaxRunes {
		return nil, errcode.New(errcode.CodeValidationError, "message content is required and must be at most 1000 characters")
	}
	chat, participant, err := s.loadAuthorizedBuildingChat(ctx, userID, chatPublicID, true)
	if err != nil {
		return nil, err
	}
	now := s.runtime.Now()
	if participant != nil && participant.MembershipStatus != chatMemberActive {
		return nil, errcode.New(errcode.CodeAuthForbidden, "member cannot send messages in this group")
	}
	if participant != nil && ((participant.MutedUntil != nil && now.Before(*participant.MutedUntil)) || (participant.BannedUntil != nil && now.Before(*participant.BannedUntil))) {
		return nil, errcode.New(errcode.CodeAuthForbidden, "member cannot send messages in this group")
	}
	message := model.Message{PublicID: utils.NewPublicID(), ChatID: chat.ID, SenderUserID: userID, MessageType: messageTypeText, ContentText: content, MessageStatus: "sent", CreatedAt: now}
	if err := s.runtime.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&message).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.Chat{}).Where("id = ?", chat.ID).Updates(map[string]any{"last_message_preview": content, "last_message_at": now, "updated_at": now}).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.ChatParticipant{}).Where("chat_id = ? AND user_id <> ? AND membership_status = ?", chat.ID, userID, chatMemberActive).Update("unread_count", gorm.Expr("unread_count + ?", 1)).Error; err != nil {
			return err
		}
		var recipients []model.ChatParticipant
		if err := tx.Where("chat_id = ? AND user_id <> ? AND membership_status = ?", chat.ID, userID, chatMemberActive).Find(&recipients).Error; err != nil {
			return err
		}
		for _, recipient := range recipients {
			if err := NewNotificationService(s.runtime).CreateNotification(ctx, tx, CreateNotificationParams{UserID: recipient.UserID, Category: "chat_message", Title: "大廈群組新訊息", Body: content, RelatedType: "chat", RelatedPublicID: chat.PublicID}); err != nil {
				return err
			}
		}
		if participant == nil {
			return nil
		}
		return tx.Model(&model.ChatParticipant{}).Where("id = ?", participant.ID).Updates(map[string]any{"last_read_message_id": message.ID, "unread_count": 0}).Error
	}); err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to send message")
	}
	return &MessageResponse{MessageID: message.PublicID, SenderUserID: fmtInt64(userID), Content: content, MessageType: messageTypeText, Status: message.MessageStatus, CreatedAt: now.UTC().Format(time.RFC3339)}, nil
}

// 9. LeaveBuildingChat marks a member as having left the group.
func (s *ChatService) LeaveBuildingChat(ctx context.Context, userID int64, chatPublicID string) error {
	chat, participant, err := s.loadAuthorizedBuildingChat(ctx, userID, chatPublicID, true)
	if err != nil {
		return err
	}
	if participant == nil {
		return errcode.New(errcode.CodeAuthForbidden, "chat membership required")
	}
	if err := s.runtime.DB.WithContext(ctx).Model(&model.ChatParticipant{}).Where("id = ?", participant.ID).Updates(map[string]any{"membership_status": chatMemberLeft, "unread_count": 0}).Error; err != nil {
		return errcode.New(errcode.CodeInternalError, "failed to leave building chat")
	}
	_ = chat
	return nil
}

// 10. RequestBuildingChatJoin creates a pending request requiring administrator approval.
func (s *ChatService) RequestBuildingChatJoin(ctx context.Context, userID int64, chatPublicID string, reason string) (*BuildingChatJoinRequest, error) {
	chat, _, err := s.loadBuildingChat(ctx, chatPublicID)
	if err != nil {
		return nil, err
	}
	if err := s.requireBuildingAccess(ctx, userID, chat.BuildingID); err != nil {
		return nil, err
	}
	var participant model.ChatParticipant
	if err := s.runtime.DB.WithContext(ctx).Where("chat_id = ? AND user_id = ?", chat.ID, userID).First(&participant).Error; err == nil && participant.MembershipStatus == chatMemberActive {
		return nil, errcode.New(errcode.CodeValidationError, "member is already in the group")
	}
	request := model.ChatJoinRequest{PublicID: utils.NewPublicID(), ChatID: chat.ID, UserID: userID, Status: chatJoinRequestPending, Reason: strings.TrimSpace(reason), CreatedAt: s.runtime.Now(), UpdatedAt: s.runtime.Now()}
	if err := s.runtime.DB.WithContext(ctx).Where("chat_id = ? AND user_id = ? AND status = ?", chat.ID, userID, chatJoinRequestPending).FirstOrCreate(&request).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to create join request")
	}
	return mapJoinRequest(&request), nil
}

// 11. ReviewBuildingChatJoin approves or rejects a pending member request.
func (s *ChatService) ReviewBuildingChatJoin(ctx context.Context, operatorID int64, requestPublicID string, status string) (*BuildingChatJoinRequest, error) {
	var request model.ChatJoinRequest
	if err := s.runtime.DB.WithContext(ctx).Where("public_id = ?", strings.TrimSpace(requestPublicID)).First(&request).Error; err != nil {
		return nil, errcode.New(errcode.CodeNotFound, "join request not found")
	}
	chat, _, err := s.loadBuildingChatByID(ctx, request.ChatID)
	if err != nil {
		return nil, err
	}
	if !s.canManageBuilding(ctx, operatorID, chat) {
		return nil, errcode.New(errcode.CodeAuthForbidden, "building administrator access required")
	}
	if status != chatJoinRequestApproved && status != chatJoinRequestRejected {
		return nil, errcode.New(errcode.CodeValidationError, "status must be approved or rejected")
	}
	if status == chatJoinRequestApproved {
		if err := s.requireBuildingAccess(ctx, request.UserID, chat.BuildingID); err != nil {
			return nil, errcode.New(errcode.CodeAuthForbidden, "member no longer has building access")
		}
	}
	now := s.runtime.Now()
	if err := s.runtime.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&request).Updates(map[string]any{"status": status, "reviewed_by": operatorID, "reviewed_at": now, "updated_at": now}).Error; err != nil {
			return err
		}
		if status != chatJoinRequestApproved {
			return nil
		}
		var participant model.ChatParticipant
		findErr := tx.Where("chat_id = ? AND user_id = ?", chat.ID, request.UserID).First(&participant).Error
		if errors.Is(findErr, gorm.ErrRecordNotFound) {
			return tx.Create(&model.ChatParticipant{ChatID: chat.ID, UserID: request.UserID, RoleInChat: "member", MembershipStatus: chatMemberActive, JoinedAt: now}).Error
		}
		if findErr != nil {
			return findErr
		}
		return tx.Model(&participant).Updates(map[string]any{"membership_status": chatMemberActive, "banned_until": nil, "muted_until": nil, "joined_at": now}).Error
	}); err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to review join request")
	}
	request.Status, request.ReviewedBy, request.ReviewedAt = status, &operatorID, &now
	return mapJoinRequest(&request), nil
}

// 12. ListBuildingChatMembers returns active and moderated members in one group.
func (s *ChatService) ListBuildingChatMembers(ctx context.Context, userID int64, chatPublicID string) ([]BuildingChatMember, error) {
	chat, _, err := s.loadAuthorizedBuildingChat(ctx, userID, chatPublicID, true)
	if err != nil {
		return nil, err
	}
	var rows []model.ChatParticipant
	if err := s.runtime.DB.WithContext(ctx).Where("chat_id = ? AND membership_status <> ?", chat.ID, chatMemberLeft).Order("joined_at asc").Find(&rows).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load chat members")
	}
	result := make([]BuildingChatMember, 0, len(rows))
	for _, row := range rows {
		preview := s.loadUserPreview(ctx, row.UserID)
		result = append(result, BuildingChatMember{UserID: fmtInt64(row.UserID), PublicID: preview.PublicID, DisplayName: preview.DisplayName, RoleInChat: row.RoleInChat, MembershipStatus: row.MembershipStatus, MutedUntil: buildingChatFormatOptionalTime(row.MutedUntil), BannedUntil: buildingChatFormatOptionalTime(row.BannedUntil), JoinedAt: row.JoinedAt.UTC().Format(time.RFC3339)})
	}
	return result, nil
}

// 13.1 ListBuildingChatJoinRequests returns pending requests for a building administrator.
func (s *ChatService) ListBuildingChatJoinRequests(ctx context.Context, operatorID int64, chatPublicID string) ([]BuildingChatJoinRequest, error) {
	chat, _, err := s.loadBuildingChat(ctx, chatPublicID)
	if err != nil {
		return nil, err
	}
	if !s.canManageBuilding(ctx, operatorID, chat) {
		return nil, errcode.New(errcode.CodeAuthForbidden, "building administrator access required")
	}
	var rows []model.ChatJoinRequest
	if err := s.runtime.DB.WithContext(ctx).Where("chat_id = ?", chat.ID).Order("created_at desc").Find(&rows).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load join requests")
	}
	result := make([]BuildingChatJoinRequest, 0, len(rows))
	for _, row := range rows {
		result = append(result, *mapJoinRequest(&row))
	}
	return result, nil
}

// 13. ModerateBuildingChatMember applies a scoped member action and writes an audit record.
func (s *ChatService) ModerateBuildingChatMember(ctx context.Context, operatorID int64, chatPublicID string, targetUserID int64, action string, durationMinutes int, reason string) error {
	chat, _, err := s.loadBuildingChat(ctx, chatPublicID)
	if err != nil {
		return err
	}
	if !s.canManageBuilding(ctx, operatorID, chat) {
		return errcode.New(errcode.CodeAuthForbidden, "building administrator access required")
	}
	if targetUserID == operatorID && action != chatModerationUnmute && action != chatModerationUnban {
		return errcode.New(errcode.CodeValidationError, "administrator cannot moderate self")
	}
	var member model.ChatParticipant
	if err := s.runtime.DB.WithContext(ctx).Where("chat_id = ? AND user_id = ?", chat.ID, targetUserID).First(&member).Error; err != nil {
		return errcode.New(errcode.CodeNotFound, "chat member not found")
	}
	var until *time.Time
	if durationMinutes > 0 {
		value := s.runtime.Now().Add(time.Duration(durationMinutes) * time.Minute)
		until = &value
	}
	updates := map[string]any{}
	switch action {
	case chatModerationMute:
		updates["muted_until"] = until
	case chatModerationUnmute:
		updates["muted_until"] = nil
	case chatModerationKick:
		updates["membership_status"] = chatMemberKicked
	case chatModerationBan:
		updates["membership_status"], updates["banned_until"] = chatMemberBanned, until
	case chatModerationUnban:
		updates["membership_status"], updates["banned_until"] = chatMemberActive, nil
	default:
		return errcode.New(errcode.CodeValidationError, "unsupported moderation action")
	}
	if err := s.runtime.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&member).Updates(updates).Error; err != nil {
			return err
		}
		return tx.Create(&model.ModerationAction{TargetType: chatModerationTargetMember, TargetID: member.ID, ActionType: action, ReasonText: strings.TrimSpace(reason), OperatorUserID: operatorID, CreatedAt: s.runtime.Now()}).Error
	}); err != nil {
		return errcode.New(errcode.CodeInternalError, "failed to moderate chat member")
	}
	return nil
}

// 14. MarkBuildingChatRead clears unread state for one group member.
func (s *ChatService) MarkBuildingChatRead(ctx context.Context, userID int64, chatPublicID string) error {
	chat, participant, err := s.loadAuthorizedBuildingChat(ctx, userID, chatPublicID, true)
	if err != nil {
		return err
	}
	if participant == nil {
		return nil
	}
	var latest model.Message
	findErr := s.runtime.DB.WithContext(ctx).Where("chat_id = ?", chat.ID).Order("created_at desc").First(&latest).Error
	if findErr != nil && !errors.Is(findErr, gorm.ErrRecordNotFound) {
		return errcode.New(errcode.CodeInternalError, "failed to load latest message")
	}
	updates := map[string]any{"unread_count": 0}
	if latest.ID > 0 {
		updates["last_read_message_id"] = latest.ID
	}
	return s.runtime.DB.WithContext(ctx).Model(&model.ChatParticipant{}).Where("id = ?", participant.ID).Updates(updates).Error
}

// 15. ensureBuildingChat ensures one group and one active membership.
func (s *ChatService) ensureBuildingChat(ctx context.Context, userID int64, buildingID string) (*model.Chat, *model.ChatParticipant, error) {
	buildingID = strings.TrimSpace(buildingID)
	if buildingID == "" {
		return nil, nil, errcode.New(errcode.CodeValidationError, "building id is required")
	}
	if err := s.requireBuildingAccess(ctx, userID, buildingID); err != nil {
		return nil, nil, err
	}
	var chat model.Chat
	err := s.runtime.DB.WithContext(ctx).Where("chat_type = ? AND building_id = ?", chatTypeBuildingGroup, buildingID).First(&chat).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		chat = model.Chat{PublicID: utils.NewPublicID(), BizModule: "building", BuildingID: buildingID, ChatType: chatTypeBuildingGroup, CreatedBy: userID, CreatedAt: s.runtime.Now(), UpdatedAt: s.runtime.Now()}
		if err := s.runtime.DB.WithContext(ctx).Create(&chat).Error; err != nil {
			// A concurrent first access may have created the unique building group.
			findErr := s.runtime.DB.WithContext(ctx).Where("chat_type = ? AND building_id = ?", chatTypeBuildingGroup, buildingID).First(&chat).Error
			if findErr != nil {
				return nil, nil, errcode.New(errcode.CodeInternalError, "failed to create building chat")
			}
		}
	} else if err != nil {
		return nil, nil, errcode.New(errcode.CodeInternalError, "failed to load building chat")
	}
	var participant model.ChatParticipant
	err = s.runtime.DB.WithContext(ctx).Where("chat_id = ? AND user_id = ?", chat.ID, userID).First(&participant).Error
	now := s.runtime.Now()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		role := "member"
		if s.canManageBuilding(ctx, userID, &chat) {
			role = "admin"
		}
		participant = model.ChatParticipant{ChatID: chat.ID, UserID: userID, RoleInChat: role, MembershipStatus: chatMemberActive, JoinedAt: now}
		if err := s.runtime.DB.WithContext(ctx).Create(&participant).Error; err != nil {
			return nil, nil, errcode.New(errcode.CodeInternalError, "failed to join building chat")
		}
	} else if err != nil {
		return nil, nil, errcode.New(errcode.CodeInternalError, "failed to load building membership")
	} else if participant.MembershipStatus == chatMemberLeft {
		if err := s.runtime.DB.WithContext(ctx).Model(&participant).Updates(map[string]any{"membership_status": chatMemberActive, "joined_at": now}).Error; err != nil {
			return nil, nil, errcode.New(errcode.CodeInternalError, "failed to rejoin building chat")
		}
		participant.MembershipStatus = chatMemberActive
	} else if participant.MembershipStatus == chatMemberBanned {
		if participant.BannedUntil == nil || now.Before(*participant.BannedUntil) {
			return nil, nil, errcode.New(errcode.CodeAuthForbidden, "member is banned from this group")
		}
		if err := s.runtime.DB.WithContext(ctx).Model(&participant).Updates(map[string]any{"membership_status": chatMemberActive, "banned_until": nil}).Error; err != nil {
			return nil, nil, errcode.New(errcode.CodeInternalError, "failed to restore expired group ban")
		}
		participant.MembershipStatus = chatMemberActive
		participant.BannedUntil = nil
	} else if participant.MembershipStatus == chatMemberKicked {
		return nil, nil, errcode.New(errcode.CodeAuthForbidden, "member was removed from this group")
	}
	return &chat, &participant, nil
}

// 16. loadAuthorizedBuildingChat loads a group and checks active membership or staff access.
func (s *ChatService) loadAuthorizedBuildingChat(ctx context.Context, userID int64, chatPublicID string, requireMembership bool) (*model.Chat, *model.ChatParticipant, error) {
	chat, _, err := s.loadBuildingChat(ctx, chatPublicID)
	if err != nil {
		return nil, nil, err
	}
	var participant model.ChatParticipant
	participantErr := s.runtime.DB.WithContext(ctx).Where("chat_id = ? AND user_id = ?", chat.ID, userID).First(&participant).Error
	if errors.Is(participantErr, gorm.ErrRecordNotFound) {
		if requireMembership && !s.isStaff(ctx, userID) {
			return nil, nil, errcode.New(errcode.CodeAuthForbidden, "chat is not accessible")
		}
		return chat, nil, nil
	}
	if participantErr != nil {
		return nil, nil, errcode.New(errcode.CodeInternalError, "failed to load building membership")
	}
	if !requireMembership || s.isStaff(ctx, userID) {
		return chat, &participant, nil
	}
	if participant.MembershipStatus == chatMemberBanned && participant.BannedUntil != nil && !s.runtime.Now().Before(*participant.BannedUntil) {
		if err := s.runtime.DB.WithContext(ctx).Model(&participant).Updates(map[string]any{"membership_status": chatMemberActive, "banned_until": nil}).Error; err != nil {
			return nil, nil, errcode.New(errcode.CodeInternalError, "failed to restore expired group ban")
		}
		participant.MembershipStatus = chatMemberActive
		participant.BannedUntil = nil
	}
	if participant.MembershipStatus != chatMemberActive || (participant.BannedUntil != nil && s.runtime.Now().Before(*participant.BannedUntil)) {
		return nil, nil, errcode.New(errcode.CodeAuthForbidden, "chat is not accessible")
	}
	if participant.MutedUntil != nil && s.runtime.Now().After(*participant.MutedUntil) {
		_ = s.runtime.DB.WithContext(ctx).Model(&participant).Update("muted_until", nil)
		participant.MutedUntil = nil
	}
	return chat, &participant, nil
}

// 17. loadBuildingChat loads only building group records.
func (s *ChatService) loadBuildingChat(ctx context.Context, chatPublicID string) (*model.Chat, *model.ChatParticipant, error) {
	var chat model.Chat
	if err := s.runtime.DB.WithContext(ctx).Where("public_id = ? AND chat_type = ?", strings.TrimSpace(chatPublicID), chatTypeBuildingGroup).First(&chat).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, errcode.New(errcode.CodeNotFound, "building chat not found")
		}
		return nil, nil, errcode.New(errcode.CodeInternalError, "failed to load building chat")
	}
	return &chat, nil, nil
}

// 18. loadBuildingChatByID loads a group by database ID for internal review flows.
func (s *ChatService) loadBuildingChatByID(ctx context.Context, chatID int64) (*model.Chat, *model.ChatParticipant, error) {
	var chat model.Chat
	if err := s.runtime.DB.WithContext(ctx).Where("id = ? AND chat_type = ?", chatID, chatTypeBuildingGroup).First(&chat).Error; err != nil {
		return nil, nil, errcode.New(errcode.CodeNotFound, "building chat not found")
	}
	return &chat, nil, nil
}

// 19. requireBuildingAccess checks bound profile, active authorization, or staff access.
func (s *ChatService) requireBuildingAccess(ctx context.Context, userID int64, buildingID string) error {
	if s.isStaff(ctx, userID) {
		return nil
	}
	ids, err := s.visibleBuildingIDs(ctx, userID)
	if err != nil {
		return err
	}
	for _, id := range ids {
		if id == strings.TrimSpace(buildingID) {
			return nil
		}
	}
	return errcode.New(errcode.CodeAuthForbidden, "building is not accessible")
}

// 20. requireBuildingAdmin limits moderation to AJO staff or building owners.
func (s *ChatService) requireBuildingAdmin(ctx context.Context, userID int64, buildingID string) error {
	if s.isStaff(ctx, userID) {
		return nil
	}
	var count int64
	if err := s.runtime.DB.WithContext(ctx).Model(&model.BuildingAuthorization{}).Where("owner_user_id = ? AND building_id = ? AND status = ?", userID, buildingID, BuildingAuthorizationStatusActive).Count(&count).Error; err != nil {
		return errcode.New(errcode.CodeInternalError, "failed to verify building administrator")
	}
	if count == 0 {
		return errcode.New(errcode.CodeAuthForbidden, "building administrator access required")
	}
	return nil
}

// 20.1 canManageBuilding returns scoped group moderation access.
func (s *ChatService) canManageBuilding(ctx context.Context, userID int64, chat *model.Chat) bool {
	if s.isStaff(ctx, userID) {
		return true
	}
	var count int64
	if err := s.runtime.DB.WithContext(ctx).Model(&model.BuildingAuthorization{}).
		Where("owner_user_id = ? AND building_id = ? AND status = ?", userID, chat.BuildingID, BuildingAuthorizationStatusActive).
		Count(&count).Error; err != nil {
		return false
	}
	return count > 0
}

// 21. visibleBuildingIDs combines current profile bindings and active local grants.
func (s *ChatService) visibleBuildingIDs(ctx context.Context, userID int64) ([]string, error) {
	result := make([]string, 0)
	var profile model.UserProfile
	if err := s.runtime.DB.WithContext(ctx).Where("user_id = ?", userID).First(&profile).Error; err == nil {
		var values []any
		if json.Unmarshal(profile.BoundBuildingIDs, &values) == nil {
			for _, value := range values {
				if id, ok := value.(string); ok && strings.TrimSpace(id) != "" {
					result = appendUniqueBuilding(result, id)
				}
			}
		}
		if profile.ResidenceBindingStatus != "pending" && profile.PrimaryCommunityID != nil {
			var community model.Community
			if s.runtime.DB.WithContext(ctx).First(&community, *profile.PrimaryCommunityID).Error == nil {
				result = appendUniqueBuilding(result, community.PublicID)
			}
		}
	}
	var grants []model.BuildingAuthorization
	if err := s.runtime.DB.WithContext(ctx).Where("(grantee_user_id = ? OR owner_user_id = ?) AND status = ?", userID, userID, BuildingAuthorizationStatusActive).Find(&grants).Error; err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "failed to load building access")
	}
	for _, grant := range grants {
		result = appendUniqueBuilding(result, grant.BuildingID)
	}
	return result, nil
}

// 22. isStaff returns the authoritative AJO staff flag.
func (s *ChatService) isStaff(ctx context.Context, userID int64) bool {
	var user model.User
	return s.runtime.DB.WithContext(ctx).Select("is_staff").First(&user, userID).Error == nil && user.IsStaff
}

// 23. buildBuildingChatSummary maps group metadata for list responses.
func (s *ChatService) buildBuildingChatSummary(ctx context.Context, chat *model.Chat, participant *model.ChatParticipant) BuildingChatSummary {
	var count int64
	_ = s.runtime.DB.WithContext(ctx).Model(&model.ChatParticipant{}).Where("chat_id = ? AND membership_status = ?", chat.ID, chatMemberActive).Count(&count).Error
	result := BuildingChatSummary{ChatID: chat.PublicID, BuildingID: chat.BuildingID, ChatType: chat.ChatType, MemberCount: count, LastPreview: chat.LastMessagePreview, UnreadCount: participant.UnreadCount}
	if chat.LastMessageAt != nil {
		value := chat.LastMessageAt.UTC().Format(time.RFC3339)
		result.LastMessageAt = &value
	}
	return result
}

// 24. loadUserPreview loads a minimal display payload without exposing credentials.
func (s *ChatService) loadUserPreview(ctx context.Context, userID int64) ChatPeerSummary {
	var user model.User
	_ = s.runtime.DB.WithContext(ctx).First(&user, userID).Error
	var profile model.UserProfile
	_ = s.runtime.DB.WithContext(ctx).Where("user_id = ?", userID).First(&profile).Error
	name := strings.TrimSpace(profile.DisplayName)
	if name == "" {
		name = fmtInt64(userID)
	}
	return ChatPeerSummary{UserID: fmtInt64(userID), PublicID: user.PublicID, DisplayName: name}
}

// 25. mapJoinRequest maps a persisted request to a stable response.
func mapJoinRequest(value *model.ChatJoinRequest) *BuildingChatJoinRequest {
	result := &BuildingChatJoinRequest{RequestID: value.PublicID, ChatID: fmtInt64(value.ChatID), UserID: fmtInt64(value.UserID), Status: value.Status, Reason: value.Reason, CreatedAt: value.CreatedAt.UTC().Format(time.RFC3339)}
	if value.ReviewedAt != nil {
		formatted := value.ReviewedAt.UTC().Format(time.RFC3339)
		result.ReviewedAt = &formatted
	}
	return result
}

// 26. buildingChatFormatOptionalTime maps an optional timestamp to RFC3339.
func buildingChatFormatOptionalTime(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := value.UTC().Format(time.RFC3339)
	return &formatted
}

// 27. appendUniqueBuilding adds a non-empty building identifier once.
func appendUniqueBuilding(values []string, value string) []string {
	value = strings.TrimSpace(value)
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	if value != "" {
		values = append(values, value)
	}
	return values
}

// 28. ListAllBuildingChatsForStaff returns all building chats for staff management.
func (s *ChatService) ListAllBuildingChatsForStaff(ctx context.Context) ([]BuildingChatSummary, error) {
	var chats []model.Chat
	if err := s.runtime.DB.WithContext(ctx).
		Where("chat_type = ?", chatTypeBuildingGroup).
		Order("created_at DESC").
		Find(&chats).Error; err != nil {
		return nil, errcode.New(errcode.CodeDatabaseError, "failed to load building chats")
	}

	items := make([]BuildingChatSummary, 0, len(chats))
	for _, chat := range chats {
		var count int64
		s.runtime.DB.WithContext(ctx).
			Model(&model.ChatParticipant{}).
			Where("chat_id = ? AND membership_status = ?", chat.ID, chatMemberActive).
			Count(&count)

		summary := BuildingChatSummary{
			ChatID:       chat.PublicID,
			BuildingID:   chat.BuildingID,
			ChatType:     chat.ChatType,
			MemberCount:  int(count),
			LastPreview:  chat.LastMessagePreview,
			UnreadCount:  0,
		}
		if chat.LastMessageAt != nil {
			value := chat.LastMessageAt.UTC().Format(time.RFC3339)
			summary.LastMessageAt = &value
		}
		items = append(items, summary)
	}
	return items, nil
}

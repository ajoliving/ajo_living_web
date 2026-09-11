/*
 * Chat and moderation data models.
 * 1. Define chat sessions, participants, messages, and moderation records.
 * 2. Support secondhand contact and unread state tracking.
 */
package model

import "time"

// 1. Chat stores a listing-scoped conversation.
type Chat struct {
	ID                 int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	PublicID           string     `gorm:"type:varchar(26);not null;uniqueIndex" json:"public_id"`
	BizModule          string     `gorm:"type:varchar(32);not null" json:"biz_module"`
	BuildingID         string     `gorm:"type:varchar(120);index:idx_chats_building_type" json:"building_id"`
	ListingID          int64      `gorm:"not null;index:idx_chats_listing_updated,priority:1" json:"listing_id"`
	ChatType           string     `gorm:"type:varchar(32);not null" json:"chat_type"`
	CreatedBy          int64      `gorm:"not null;index" json:"created_by"`
	LastMessagePreview string     `gorm:"type:varchar(500)" json:"last_message_preview"`
	LastMessageAt      *time.Time `json:"last_message_at"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `gorm:"index:idx_chats_listing_updated,priority:2" json:"updated_at"`
}

// 2. ChatParticipant stores a chat membership record.
type ChatParticipant struct {
	ID                int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	ChatID            int64      `gorm:"not null;uniqueIndex:uk_chat_participants_chat_user,priority:1;index" json:"chat_id"`
	UserID            int64      `gorm:"not null;uniqueIndex:uk_chat_participants_chat_user,priority:2;index" json:"user_id"`
	RoleInChat        string     `gorm:"type:varchar(32);not null" json:"role_in_chat"`
	MembershipStatus  string     `gorm:"type:varchar(24);not null;default:active;index" json:"membership_status"`
	MutedUntil        *time.Time `json:"muted_until,omitempty"`
	BannedUntil       *time.Time `json:"banned_until,omitempty"`
	LastReadMessageID *int64     `json:"last_read_message_id"`
	UnreadCount       int        `gorm:"not null" json:"unread_count"`
	JoinedAt          time.Time  `json:"joined_at"`
}

// 3. ChatJoinRequest stores a member request to enter a building group.
type ChatJoinRequest struct {
	ID         int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	PublicID   string     `gorm:"type:varchar(26);not null;uniqueIndex" json:"public_id"`
	ChatID     int64      `gorm:"not null;uniqueIndex:uk_chat_join_requests_chat_user,priority:1;index" json:"chat_id"`
	UserID     int64      `gorm:"not null;uniqueIndex:uk_chat_join_requests_chat_user,priority:2;index" json:"user_id"`
	Status     string     `gorm:"type:varchar(24);not null;index" json:"status"`
	Reason     string     `gorm:"type:varchar(500)" json:"reason"`
	ReviewedBy *int64     `gorm:"index" json:"reviewed_by,omitempty"`
	ReviewedAt *time.Time `json:"reviewed_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// 4. Message stores chat messages.
type Message struct {
	ID              int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	PublicID        string    `gorm:"type:varchar(26);not null;uniqueIndex" json:"public_id"`
	ChatID          int64     `gorm:"not null;index:idx_messages_chat_time,priority:1;uniqueIndex:uk_messages_client_id,priority:1" json:"chat_id"`
	SenderUserID    int64     `gorm:"not null;index;uniqueIndex:uk_messages_client_id,priority:2" json:"sender_user_id"`
	ClientMessageID *string   `gorm:"type:varchar(80);uniqueIndex:uk_messages_client_id,priority:3" json:"client_message_id,omitempty"`
	MessageType     string    `gorm:"type:varchar(32);not null" json:"message_type"`
	ContentText     string    `gorm:"type:text;not null" json:"content_text"`
	ActionLabel     string    `gorm:"type:varchar(80)" json:"action_label"`
	ActionURL       string    `gorm:"type:varchar(500)" json:"action_url"`
	MessageStatus   string    `gorm:"type:varchar(32);not null" json:"message_status"`
	CreatedAt       time.Time `gorm:"index:idx_messages_chat_time,priority:2" json:"created_at"`
}

// 5. RealtimeOutbox stores committed realtime events until publication succeeds.
type RealtimeOutbox struct {
	ID            int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	EventType     string     `gorm:"type:varchar(32);not null;index" json:"event_type"`
	AggregateID   string     `gorm:"type:varchar(80);not null;index" json:"aggregate_id"`
	Payload       string     `gorm:"type:text;not null" json:"payload"`
	Status        string     `gorm:"type:varchar(20);not null;index" json:"status"`
	Attempts      int        `gorm:"not null" json:"attempts"`
	NextAttemptAt time.Time  `gorm:"index" json:"next_attempt_at"`
	PublishedAt   *time.Time `json:"published_at,omitempty"`
	LastError     string     `gorm:"type:varchar(500)" json:"last_error,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

// 6. MessageAttachment binds an uploaded media asset to one chat message.
type MessageAttachment struct {
	ID           int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	MessageID    int64     `gorm:"not null;uniqueIndex:uk_message_attachments_message_asset,priority:1;index" json:"message_id"`
	MediaAssetID int64     `gorm:"not null;uniqueIndex:uk_message_attachments_message_asset,priority:2;index" json:"media_asset_id"`
	SortOrder    int       `gorm:"not null" json:"sort_order"`
	CreatedAt    time.Time `json:"created_at"`
}

// 7. ModerationAction stores manual moderation records.
type ModerationAction struct {
	ID             int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TargetType     string    `gorm:"type:varchar(32);not null" json:"target_type"`
	TargetID       int64     `gorm:"not null;index" json:"target_id"`
	ActionType     string    `gorm:"type:varchar(32);not null" json:"action_type"`
	ReasonCode     string    `gorm:"type:varchar(64)" json:"reason_code"`
	ReasonText     string    `gorm:"type:varchar(500)" json:"reason_text"`
	OperatorUserID int64     `gorm:"not null;index" json:"operator_user_id"`
	CreatedAt      time.Time `json:"created_at"`
}

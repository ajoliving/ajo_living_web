/*
 * Chat response DTOs.
 * 1. Define chat list, detail, member, peer, and message payloads.
 * 2. Keep transport-facing structures separate from chat persistence logic.
 */
package service

// 1. ChatSummary defines the chat list payload.
type ChatSummary struct {
	ChatID             string                `json:"chat_id"`
	ListingID          string                `json:"listing_id"`
	ListingTitle       string                `json:"listing_title"`
	BizModule          string                `json:"biz_module"`
	ChatType           string                `json:"chat_type"`
	LastMessagePreview string                `json:"last_message_preview"`
	LastMessageAt      *string               `json:"last_message_at,omitempty"`
	UnreadCount        int                   `json:"unread_count"`
	Peer               *ChatPeerSummary      `json:"peer,omitempty"`
	Listing            *ChatListingSummary   `json:"listing,omitempty"`
	CoverImage         *ListingImageResponse `json:"cover_image,omitempty"`
}

// 2. ChatDetail defines the chat detail payload.
type ChatDetail struct {
	ChatID       string              `json:"chat_id"`
	ListingID    string              `json:"listing_id"`
	ListingTitle string              `json:"listing_title"`
	BizModule    string              `json:"biz_module"`
	ChatType     string              `json:"chat_type"`
	CreatedAt    string              `json:"created_at"`
	Peer         *ChatPeerSummary    `json:"peer,omitempty"`
	Listing      *ChatListingSummary `json:"listing,omitempty"`
	Participants []ChatMember        `json:"participants"`
}

// 3. ChatPeerSummary defines the opposite-side display payload.
type ChatPeerSummary struct {
	UserID      string `json:"user_id"`
	PublicID    string `json:"public_id"`
	DisplayName string `json:"display_name"`
	RoleInChat  string `json:"role_in_chat"`
}

// 4. ChatListingSummary defines the listing snippet shown in chat payloads.
type ChatListingSummary struct {
	ListingID      string                `json:"listing_id"`
	BizModule      string                `json:"biz_module"`
	Title          string                `json:"title"`
	Summary        string                `json:"summary"`
	PublishedAt    *string               `json:"published_at,omitempty"`
	BusinessStatus string                `json:"business_status"`
	PriceMode      string                `json:"price_mode"`
	PriceHKD       *float64              `json:"price_hkd,omitempty"`
	CoverImage     *ListingImageResponse `json:"cover_image,omitempty"`
}

// 5. ChatMember defines a chat participant response.
type ChatMember struct {
	UserID      string `json:"user_id"`
	PublicID    string `json:"public_id"`
	DisplayName string `json:"display_name"`
	RoleInChat  string `json:"role_in_chat"`
}

// 6. MessageResponse defines a chat message payload.
type MessageResponse struct {
	MessageID    string `json:"message_id"`
	SenderUserID string `json:"sender_user_id"`
	Content      string `json:"content"`
	MessageType  string `json:"message_type"`
	ActionLabel  string `json:"action_label,omitempty"`
	ActionURL    string `json:"action_url,omitempty"`
	Status       string `json:"status"`
	CreatedAt    string `json:"created_at"`
}

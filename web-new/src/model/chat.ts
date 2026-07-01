/*
 * 站內聊天資料型別。
 * 1. 描述會話列表與訊息內容。
 * 2. 為聊天原型與後續 API 對接預留一致結構。
 */
import type { Listing } from '@/model/listing';
import type { UserSummary } from '@/model/user';

// 1. 定義訊息方向
export type MessageSenderRole = 'self' | 'peer';

// 2. 定義單則訊息
export interface Message {
  id: string;
  chat_id: string;
  sender_role: MessageSenderRole;
  body: string;
  sent_at: string;
}

// 3. 定義聊天會話摘要
export interface ChatConversation {
  id: string;
  listing: Pick<Listing, 'id' | 'title' | 'price_hkd' | 'images' | 'status'>;
  peer: UserSummary;
  unread_count: number;
  last_message: string;
  last_message_at: string;
}

// 4. 定義聊天頁訊息方向
export type ChatMessageRole = 'self' | 'peer';

// 5. 定義聊天頁會話類型
export type ChatConversationType = 'direct_listing_chat';

// 6. 定義聊天頁訊息
export interface ChatMessageView {
  id: string;
  chat_id: string;
  sender_role: ChatMessageRole;
  body: string;
  message_type: string;
  action_label?: string;
  action_url?: string;
  sent_at: string;
}

// 7. 定義聊天列表商品摘要
export interface ChatListingView {
  id: string;
  title: string;
  summary: string;
  price_hkd: number;
  status: string;
  published_at?: string | null;
  cover_image_url?: string;
}

// 8. 定義聊天對方摘要
export interface ChatPeerView {
  user_id: string;
  public_id: string;
  display_name: string;
  avatar_url: string;
  role_in_chat: string;
}

// 9. 定義聊天會話摘要
export interface ChatConversationView {
  id: string;
  type: ChatConversationType;
  listing: ChatListingView;
  peer: ChatPeerView;
  unread_count: number;
  last_message: string;
  last_message_at: string;
}

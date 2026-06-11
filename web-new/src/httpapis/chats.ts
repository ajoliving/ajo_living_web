/*
 * 聊天會話 API。
 * 1. 對齊後端 listing-chat 建立與會話列表契約。
 * 2. 提供聊天詳情與已讀能力。
 */
import httpClient from '@/httpapis';
import type { ApiResponse, PaginatedResult } from '@/model/api';

export interface ChatPeerSummaryResponse {
  user_id: string;
  public_id?: string;
  display_name?: string;
  role_in_chat?: string;
}

export interface ChatListingSummaryResponse {
  listing_id: string;
  title: string;
  summary: string;
  published_at?: string | null;
  business_status: string;
  price_mode: string;
  cover_image?: {
    media_asset_id: string;
    url: string;
    sort_order: number;
    is_cover: boolean;
  };
  price_hkd?: number | null;
}

export interface ChatSummaryResponse {
  chat_id: string;
  listing_id: string;
  listing_title: string;
  chat_type: 'direct_listing_chat' | 'system_notice';
  last_message_preview: string;
  last_message_at?: string | null;
  unread_count: number;
  cover_image?: {
    media_asset_id: string;
    url: string;
    sort_order: number;
    is_cover: boolean;
  };
  peer?: ChatPeerSummaryResponse | null;
  listing?: ChatListingSummaryResponse | null;
}

export interface ChatDetailResponse {
  chat_id: string;
  listing_id: string;
  listing_title: string;
  chat_type: 'direct_listing_chat' | 'system_notice';
  created_at: string;
  participants: Array<{
    user_id: string;
    role_in_chat: string;
    public_id?: string;
    display_name?: string;
  }>;
  peer?: ChatPeerSummaryResponse | null;
  listing?: ChatListingSummaryResponse | null;
}

interface FetchChatsParams {
  page?: number;
  page_size?: number;
}

export interface PublishSystemNoticePayload {
  title: string;
  body: string;
  action_label?: string;
  action_url?: string;
}

export interface PublishSystemNoticeResponse {
  delivered_count: number;
}

// 1. 建立或復用一個帖子聊天
export const createOrReuseChat = (listingId: string) =>
  httpClient.post<ApiResponse<{ chat_id: string; is_new: boolean }>>(`/listings/${listingId}/chats`);

// 2. 取得聊天會話列表
export const fetchChats = (params: FetchChatsParams = {}) =>
  httpClient.get<ApiResponse<PaginatedResult<ChatSummaryResponse>>>('/chats', { params });

// 3. 取得單個聊天詳情
export const fetchChatDetail = (chatId: string) =>
  httpClient.get<ApiResponse<ChatDetailResponse>>(`/chats/${chatId}`);

// 4. 標記聊天已讀
export const markChatRead = (chatId: string) =>
  httpClient.post<ApiResponse<{ chat_id: string; read: boolean }>>(`/chats/${chatId}/read`);

// 5. 發布系統通知
export const publishSystemNotice = (payload: PublishSystemNoticePayload) =>
  httpClient.post<ApiResponse<PublishSystemNoticeResponse>>('/staff/system-notices', payload);

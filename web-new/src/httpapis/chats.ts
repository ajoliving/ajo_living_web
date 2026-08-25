/*
 * 聊天會話 API。
 * 1. 對齊後端 listing-chat 建立與會話列表契約。
 * 2. 提供聊天詳情與已讀能力。
 */
import httpClient from '@/httpapis';
import type { ApiResponse, PaginatedResult } from '@/model/api';
import type { MessageResponse } from '@/httpapis/messages';

export interface ChatPeerSummaryResponse {
  user_id: string;
  public_id?: string;
  display_name?: string;
  role_in_chat?: string;
}

export interface ChatListingSummaryResponse {
  listing_id: string;
  biz_module?: 'secondhand' | 'property_sale' | 'serviced_apartment' | 'system';
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
  biz_module?: 'secondhand' | 'property_sale' | 'serviced_apartment' | 'system';
  chat_type: string;
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

export interface BuildingChatSummaryResponse {
  chat_id: string;
  building_id: string;
  chat_type: 'building_group';
  member_count: number;
  unread_count: number;
  last_message_preview: string;
  last_message_at?: string | null;
}

export interface BuildingChatMemberResponse {
  user_id: string;
  public_id: string;
  display_name: string;
  role_in_chat: string;
  membership_status: string;
  muted_until?: string | null;
  banned_until?: string | null;
  joined_at: string;
}

export interface BuildingChatJoinRequestResponse {
  request_id: string;
  chat_id: string;
  user_id: string;
  status: string;
  reason?: string;
  created_at: string;
  reviewed_at?: string | null;
}

export interface ChatDetailResponse {
  chat_id: string;
  listing_id: string;
  listing_title: string;
  biz_module?: 'secondhand' | 'property_sale' | 'serviced_apartment' | 'system';
  chat_type: string;
  created_at: string;
  building_id?: string;
  unread_count?: number;
  can_manage?: boolean;
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

export type PropertyChatChannel = 'sale' | 'serviced';

// 1. 建立或復用一個二手帖子聊天
export const createOrReuseChat = (listingId: string) =>
  httpClient.post<ApiResponse<{ chat_id: string; is_new: boolean }>>(`/listings/${listingId}/chats`);

// 2. 建立或復用一個物業頻道聊天
export const createOrReusePropertyChat = (channel: PropertyChatChannel, listingId: string) => {
  const resource = channel === 'serviced' ? 'serviced-apartments' : 'property-sales';

  return httpClient.post<ApiResponse<{ chat_id: string; is_new: boolean }>>(
    `/${resource}/${listingId}/chats`,
  );
};

// 3. 取得聊天會話列表
export const fetchChats = (params: FetchChatsParams = {}) =>
  httpClient.get<ApiResponse<PaginatedResult<ChatSummaryResponse>>>('/chats', { params });

// 4. 取得單個聊天詳情
export const fetchChatDetail = (chatId: string) =>
  httpClient.get<ApiResponse<ChatDetailResponse>>(`/chats/${chatId}`);

// 5. 標記聊天已讀
export const markChatRead = (chatId: string) =>
  httpClient.post<ApiResponse<{ chat_id: string; read: boolean }>>(`/chats/${chatId}/read`);

// 6. 取得會員可見的大廈群聊
export const fetchBuildingChats = () =>
  httpClient.get<ApiResponse<{ items: BuildingChatSummaryResponse[] }>>('/building-chats');

// 7. 首次進入時建立或加入指定大廈群聊
export const ensureBuildingChat = (buildingId: string) =>
  httpClient.post<ApiResponse<BuildingChatSummaryResponse>>(`/buildings/${buildingId}/chat`);

// 8. 取得大廈群聊詳情
export const fetchBuildingChatDetail = (chatId: string) =>
  httpClient.get<ApiResponse<ChatDetailResponse>>(`/building-chats/${chatId}`);

// 9. 取得大廈群聊訊息
export const fetchBuildingChatMessages = (chatId: string, params: FetchChatsParams = {}) =>
  httpClient.get<ApiResponse<PaginatedResult<MessageResponse>>>(
    `/building-chats/${chatId}/messages`,
    { params },
  );

// 10. 發送大廈群聊訊息
export const sendBuildingChatMessage = (chatId: string, payload: { content: string }) =>
  httpClient.post<ApiResponse<MessageResponse>>(
    `/building-chats/${chatId}/messages`,
    payload,
  );

// 11. 標記大廈群聊已讀
export const markBuildingChatRead = (chatId: string) =>
  httpClient.post<ApiResponse<{ chat_id: string; read: boolean }>>(`/building-chats/${chatId}/read`);

// 12. 取得大廈群聊成員
export const fetchBuildingChatMembers = (chatId: string) =>
  httpClient.get<ApiResponse<{ items: BuildingChatMemberResponse[] }>>(`/building-chats/${chatId}/members`);

// 13. 取得大廈群聊申請隊列
export const fetchBuildingChatJoinRequests = (chatId: string) =>
  httpClient.get<ApiResponse<{ items: BuildingChatJoinRequestResponse[] }>>(`/building-chats/${chatId}/join-requests`);

// 14. 提交加入大廈群聊申請
export const requestBuildingChatJoin = (chatId: string, reason = '') =>
  httpClient.post<ApiResponse<BuildingChatJoinRequestResponse>>(`/building-chats/${chatId}/join-requests`, { reason });

// 15. 審核大廈群聊加入申請
export const reviewBuildingChatJoin = (requestId: string, status: 'approved' | 'rejected') =>
  httpClient.post<ApiResponse<BuildingChatJoinRequestResponse>>(`/building-chat-join-requests/${requestId}/review`, { status });

// 16. 管理大廈群聊成員
export const moderateBuildingChatMember = (
  chatId: string,
  payload: { user_id: number; action: string; duration_minutes?: number; reason?: string },
) => httpClient.post<ApiResponse<{ chat_id: string; user_id: string; action: string }>>(
  `/building-chats/${chatId}/members/moderate`,
  payload,
);

// 17. 離開大廈群聊
export const leaveBuildingChat = (chatId: string) =>
  httpClient.post<ApiResponse<{ chat_id: string; left: boolean }>>(`/building-chats/${chatId}/leave`);

// 18. Staff 查看所有大廈群聊
export const fetchStaffBuildingChats = () =>
  httpClient.get<ApiResponse<PaginatedResult<BuildingChatSummaryResponse>>>('/staff/building-chats');

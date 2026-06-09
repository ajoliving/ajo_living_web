/*
 * 訊息 API。
 * 1. 對齊後端訊息列表與發送契約。
 * 2. 提供聊天頁接入真實資料所需能力。
 */
import httpClient from '@/shared/utils/http';
import type { ApiResponse, PaginatedResult } from '@/shared/utils/http/model';

export interface MessageResponse {
  message_id: string;
  sender_user_id: string;
  sender_display_name?: string;
  sender_role?: 'self' | 'peer';
  content: string;
  message_type: string;
  action_label?: string;
  action_url?: string;
  status: string;
  created_at: string;
}

interface FetchMessagesParams {
  page?: number;
  page_size?: number;
}

// 1. 取得指定聊天訊息
export const fetchMessages = (chatId: string, params: FetchMessagesParams = {}) =>
  httpClient.get<ApiResponse<PaginatedResult<MessageResponse>>>(`/chats/${chatId}/messages`, {
    params,
  });

// 2. 發送聊天訊息
export const sendMessage = (chatId: string, payload: { content: string }) =>
  httpClient.post<ApiResponse<MessageResponse>>(`/chats/${chatId}/messages`, payload);

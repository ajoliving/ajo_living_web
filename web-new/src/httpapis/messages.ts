/*
 * 訊息 API。
 * 1. 對齊後端訊息列表與發送契約。
 * 2. 提供聊天頁接入真實資料所需能力。
 */
import httpClient from '@/httpapis';
import type { ApiResponse, PaginatedResult } from '@/model/api';

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
  attachments?: MessageAttachmentResponse[];
}

export interface MessageAttachmentResponse {
  media_asset_id: string;
  mime_type: string;
  file_size: number;
  width?: number | null;
  height?: number | null;
  url: string;
  processing_status: string;
  scan_status: string;
  rejection_reason?: string;
}

interface FetchMessagesParams {
  page?: number;
  page_size?: number;
  after_message_id?: string;
  latest?: boolean;
}

// 1. 取得指定聊天訊息
export const fetchMessages = (chatId: string, params: FetchMessagesParams = {}) =>
  httpClient.get<ApiResponse<PaginatedResult<MessageResponse>>>(`/chats/${chatId}/messages`, {
    params,
  });

// 2. 發送聊天訊息
export const sendMessage = (chatId: string, payload: { content: string; attachment_ids?: string[]; client_message_id?: string }) =>
  httpClient.post<ApiResponse<MessageResponse>>(`/chats/${chatId}/messages`, payload);

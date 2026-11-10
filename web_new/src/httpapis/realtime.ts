/*
 * 即時聊天連線 API。
 * 1. 以登入態取得短時 WebSocket ticket。
 * 2. WebSocket 只承載聊天 JSON 事件，媒體仍走 OSS 上傳 API。
 */
import httpClient from '@/httpapis';
import type { ApiResponse } from '@/model/api';

// 1. RealtimeTicketResponse defines the short-lived handshake credential.
export interface RealtimeTicketResponse {
  ticket: string;
  expires_in: number;
}

// 2. fetchRealtimeTicket requests a short-lived WebSocket ticket.
export const fetchRealtimeTicket = () =>
  httpClient.get<ApiResponse<RealtimeTicketResponse>>('/realtime/ticket');

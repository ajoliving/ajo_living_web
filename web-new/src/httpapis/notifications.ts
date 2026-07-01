/*
 * 通知 API。
 * 1. 取得會員通知列表。
 * 2. 取得未讀數並提供已讀操作。
 */
import httpClient from '@/httpapis';
import type { ApiResponse } from '@/model/api';
import type { NotificationListPayload } from '@/model/notification';

interface FetchNotificationsParams {
  page?: number;
  page_size?: number;
  only_unread?: boolean;
}

export interface PublishSystemNoticePayload {
  title: string;
  body: string;
}

export interface PublishSystemNoticeResponse {
  delivered_count: number;
}

// 1. 取得通知列表
export const fetchNotifications = (params: FetchNotificationsParams = {}) =>
  httpClient.get<ApiResponse<NotificationListPayload>>('/notifications', { params });

// 2. 取得未讀通知數
export const fetchNotificationUnreadCount = () =>
  httpClient.get<ApiResponse<{ unread_count: number }>>('/notifications/unread-count');

// 3. 標記單一通知已讀
export const markNotificationRead = (notificationId: string) =>
  httpClient.post<ApiResponse<{ notification_id: string; read: boolean }>>(
    `/notifications/${notificationId}/read`,
  );

// 4. 標記全部通知已讀
export const markAllNotificationsRead = () =>
  httpClient.post<ApiResponse<{ updated: number }>>('/notifications/read-all');

// 5. 發布系統通知
export const publishSystemNotice = (payload: PublishSystemNoticePayload) =>
  httpClient.post<ApiResponse<PublishSystemNoticeResponse>>('/staff/system-notices', payload);

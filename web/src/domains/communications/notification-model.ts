/*
 * 通知資料型別。
 * 1. 對齊會員通知列表與已讀操作。
 * 2. 支援訂單與聊天事件提示。
 */

import type { PaginationMeta } from '@/shared/utils/http/model';

// 1. 通知列表項
export interface NotificationItem {
  notification_id: string;
  category: string;
  title: string;
  body: string;
  related_type: string;
  related_public_id: string;
  is_read: boolean;
  read_at?: string;
  created_at: string;
}

// 2. 通知列表回應
export interface NotificationListPayload {
  items: NotificationItem[];
  pagination: PaginationMeta;
  unread_count: number;
}

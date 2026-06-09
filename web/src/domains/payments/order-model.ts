/*
 * 訂單資料型別。
 * 1. 對齊會員中心訂單列表與詳情資料。
 * 2. 支援一期線下交收訂單流程。
 */

import type { PaginationMeta } from '@/shared/utils/http/model';

// 1. 訂單對方摘要
export interface OrderPeer {
  public_id: string;
  display_name: string;
  role_in_order: 'buyer' | 'seller';
}

// 2. 訂單事件
export interface OrderEvent {
  action_type: string;
  from_status: string;
  to_status: string;
  operator_user_id: string;
  note: string;
  created_at: string;
}

// 3. 訂單列表項
export interface OrderSummary {
  order_id: string;
  listing_id: string;
  listing_title: string;
  order_status: 'pending_confirm' | 'confirmed' | 'completed' | 'cancelled';
  role_in_order: 'buyer' | 'seller';
  buyer_note: string;
  handover_method: string;
  cancel_reason: string;
  created_at: string;
  updated_at: string;
  confirmed_at?: string;
  completed_at?: string;
  cancelled_at?: string;
  cover_image?: {
    media_asset_id: string;
    url: string;
    sort_order: number;
    is_cover: boolean;
  };
  peer: OrderPeer;
}

// 4. 訂單詳情
export interface OrderDetail extends OrderSummary {
  logs: OrderEvent[];
}

// 5. 訂單列表回應
export interface OrderListPayload {
  items: OrderSummary[];
  pagination: PaginationMeta;
}

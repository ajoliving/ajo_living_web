/*
 * 訂單 API。
 * 1. 取得會員中心訂單列表與詳情。
 * 2. 預留後續訂單流轉操作擴展點。
 */
import httpClient from '@/httpapis';
import type { ApiResponse } from '@/model/api';
import type { OrderDetail, OrderListPayload } from '@/model/payments';

interface FetchOrdersParams {
  page?: number;
  page_size?: number;
  status?: string;
  role?: string;
}

// 1. 取得我的訂單列表
export const fetchMyOrders = (params: FetchOrdersParams = {}) =>
  httpClient.get<ApiResponse<OrderListPayload>>('/me/orders', { params });

// 2. 取得單一訂單詳情
export const fetchOrderDetail = (orderId: string) =>
  httpClient.get<ApiResponse<OrderDetail>>(`/orders/${orderId}`);

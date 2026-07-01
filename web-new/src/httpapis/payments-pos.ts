/*
 * POS 物業繳費 API。
 * 1. 串接 AJO 後端聚合的 POS 賬單、付款設定與訂單介面。
 * 2. 保持 AJO Pay 前端只依賴 AJO API，不直接連接舊 POS 服務。
 */
import httpClient from '@/httpapis';
import type { ApiResponse } from '@/model/api';
import type {
  POSPaymentListPayload,
  POSPaymentOrderCreatePayload,
  POSPaymentReportPayload,
} from '@/model/payment-pos';

export interface POSPaymentQuery {
  building_id?: string;
  unit_id?: string;
}

// 1. 取得 POS 賬單
export const fetchPOSPaymentBills = (params: POSPaymentQuery = {}) =>
  httpClient.get<ApiResponse<POSPaymentListPayload>>('/me/payments/pos/bills', { params });

// 2. 取得 POS 手續費與付款方式
export const fetchPOSPaymentFees = (params: POSPaymentQuery = {}) =>
  httpClient.get<ApiResponse<POSPaymentListPayload>>('/me/payments/pos/fees', { params });

// 3. 取得 POS 銀行賬戶
export const fetchPOSPaymentBankAccounts = (params: POSPaymentQuery = {}) =>
  httpClient.get<ApiResponse<POSPaymentListPayload>>('/me/payments/pos/bank-accounts', { params });

// 4. 取得 POS H5 訂單
export const fetchPOSPaymentOrders = (params: POSPaymentQuery = {}) =>
  httpClient.get<ApiResponse<POSPaymentListPayload>>('/me/payments/pos/orders', { params });

// 5. 建立 POS H5 訂單
export const createPOSPaymentOrder = (payload: POSPaymentOrderCreatePayload) =>
  httpClient.post<ApiResponse<Record<string, unknown>>>('/me/payments/pos/orders', payload);

// 6. 取得單一 POS H5 訂單
export const fetchPOSPaymentOrder = (
  mchOrderNo: string,
  params: { detail?: boolean; refresh?: boolean } & POSPaymentQuery = {},
) =>
  httpClient.get<ApiResponse<Record<string, unknown>>>(`/me/payments/pos/orders/${encodeURIComponent(mchOrderNo)}`, { params });

// 7. 按訂單號或支付單號查詢 POS H5 訂單
export const queryPOSPaymentOrder = (payload: Record<string, unknown>) =>
  httpClient.post<ApiResponse<Record<string, unknown>>>('/me/payments/pos/orders/query', payload);

// 8. 上報 POS 線下繳費
export const reportPOSPayment = (payload: POSPaymentReportPayload) =>
  httpClient.post<ApiResponse<Record<string, unknown>>>('/me/payments/pos/payments/report', payload);

/*
 * POS 物業繳費 API。
 * 1. 串接 AJO 後端聚合的 POS 賬單、訂單、會計與歷史介面。
 * 2. 保持前端只依賴 AJO API，不直接連接舊 POS 服務。
 */
import httpClient from '@/shared/utils/http';
import type { ApiResponse } from '@/shared/utils/http/model';
import type {
  POSPaymentAccountingPayload,
  POSPaymentHistoryQuery,
  POSPaymentListPayload,
  POSPaymentOrderCreatePayload,
  POSPaymentOverview,
  POSPaymentReportPayload,
  POSPaymentRow,
  POSTerminalPaymentPayload,
} from '@/domains/payments/pos-model';

export interface POSPaymentQuery {
  building_id?: string;
  unit_id?: string;
}

export interface POSPaymentOverviewQuery extends POSPaymentQuery {
  summary?: boolean;
}

// 1. 取得 POS 繳費概覽
export const fetchPOSPaymentOverview = (params: POSPaymentOverviewQuery = {}) =>
  httpClient.get<ApiResponse<POSPaymentOverview>>('/me/payments/pos/overview', { params });

// 2. 取得 POS 賬單
export const fetchPOSPaymentBills = (params: POSPaymentQuery = {}) =>
  httpClient.get<ApiResponse<POSPaymentListPayload>>('/me/payments/pos/bills', { params });

// 3. 取得 POS 手續費與付款方式
export const fetchPOSPaymentFees = (params: POSPaymentQuery = {}) =>
  httpClient.get<ApiResponse<POSPaymentListPayload>>('/me/payments/pos/fees', { params });

// 4. 取得 POS 銀行賬戶
export const fetchPOSPaymentBankAccounts = (params: POSPaymentQuery = {}) =>
  httpClient.get<ApiResponse<POSPaymentListPayload>>('/me/payments/pos/bank-accounts', { params });

// 5. 取得 POS H5 訂單
export const fetchPOSPaymentOrders = (params: POSPaymentQuery = {}) =>
  httpClient.get<ApiResponse<POSPaymentListPayload>>('/me/payments/pos/orders', { params });

// 6. 建立 POS H5 訂單
export const createPOSPaymentOrder = (payload: POSPaymentOrderCreatePayload) =>
  httpClient.post<ApiResponse<Record<string, unknown>>>('/me/payments/pos/orders', payload);

// 7. 取得單一 POS H5 訂單
export const fetchPOSPaymentOrder = (
  mchOrderNo: string,
  params: { detail?: boolean; refresh?: boolean; retry_business?: boolean; real_gateway?: boolean } & POSPaymentQuery = {},
) =>
  httpClient.get<ApiResponse<Record<string, unknown>>>(`/me/payments/pos/orders/${encodeURIComponent(mchOrderNo)}`, { params });

// 8. 按訂單號或支付單號查詢 POS H5 訂單
export const queryPOSPaymentOrder = (payload: Record<string, unknown>) =>
  httpClient.post<ApiResponse<Record<string, unknown>>>('/me/payments/pos/orders/query', payload);

// 9. 關閉 POS H5 訂單
export const closePOSPaymentOrder = (mchOrderNo: string, payload: Record<string, unknown> = {}) =>
  httpClient.post<ApiResponse<Record<string, unknown>>>(`/me/payments/pos/orders/${encodeURIComponent(mchOrderNo)}/close`, payload);

// 10. 取消 POS H5 訂單
export const cancelPOSPaymentOrder = (mchOrderNo: string, payload: Record<string, unknown> = {}) =>
  httpClient.post<ApiResponse<Record<string, unknown>>>(`/me/payments/pos/orders/${encodeURIComponent(mchOrderNo)}/cancel`, payload);

// 11. 模擬 POS H5 訂單
export const simulatePOSPaymentOrder = (mchOrderNo: string, payload: Record<string, unknown> = {}) =>
  httpClient.post<ApiResponse<Record<string, unknown>>>(`/me/payments/pos/orders/${encodeURIComponent(mchOrderNo)}/simulate`, payload);

// 12. 上報 POS 線下繳費
export const reportPOSPayment = (payload: POSPaymentReportPayload) =>
  httpClient.post<ApiResponse<Record<string, unknown>>>('/me/payments/pos/payments/report', payload);

// 13. POS 機收款
export const payPOSTerminal = (payload: POSTerminalPaymentPayload) =>
  httpClient.post<ApiResponse<Record<string, unknown>>>('/me/payments/pos/terminal/pay', payload);

// 14. 取得 POS 交易歷史
export const fetchPOSPaymentHistory = (params: POSPaymentHistoryQuery = {}) =>
  httpClient.get<ApiResponse<POSPaymentListPayload>>('/me/payments/pos/history', { params });

// 15. 取得 POS 交易歷史詳情
export const fetchPOSPaymentHistoryDetail = (paymentId: string, params: POSPaymentHistoryQuery = {}) =>
  httpClient.get<ApiResponse<POSPaymentRow>>(`/me/payments/pos/history/${encodeURIComponent(paymentId)}`, { params });

// 16. 取得 POS 會計資料
export const fetchPOSPaymentAccounting = (params: POSPaymentQuery = {}) =>
  httpClient.get<ApiResponse<POSPaymentAccountingPayload>>('/me/payments/pos/accounting', { params });

// 17. 清機 POS 現金或支票交易
export const clearPOSPaymentAccounting = (payload: { building_id?: string; payment_id_list: string[] }) =>
  httpClient.post<ApiResponse<Record<string, unknown>>>('/me/payments/pos/accounting/clear', payload);

// 18. 取得 POS 清機歷史
export const fetchPOSPaymentAccountingRecords = (params: POSPaymentQuery = {}) =>
  httpClient.get<ApiResponse<POSPaymentListPayload>>('/me/payments/pos/accounting/records', { params });

// 19. 取得 POS 清機詳情
export const fetchPOSPaymentAccountingRecord = (recordId: string, params: POSPaymentQuery = {}) =>
  httpClient.get<ApiResponse<POSPaymentRow>>(`/me/payments/pos/accounting/records/${encodeURIComponent(recordId)}`, { params });

// 20. 輸出可復用的 row 型別
export type { POSPaymentRow };

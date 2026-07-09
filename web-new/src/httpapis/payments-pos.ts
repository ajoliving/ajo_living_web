/*
 * POS 物業繳費 API。
 * 1. 串接 AJO 後端聚合的 POS 賬單、付款設定與訂單介面。
 * 2. 保持 AJO Pay 前端只依賴 AJO API，不直接連接舊 POS 服務。
 * 3. 串接 AJO 受控會員態 POS 未繳賬單與交易紀錄查詢介面。
 */
import type { AxiosResponse } from 'axios';

import httpClient from '@/httpapis';
import type { ApiResponse } from '@/model/api';
import type {
  POSIntegrationPaymentTransaction,
  POSIntegrationTransactionDateQuery,
  POSIntegrationTransactionsPayload,
  POSIntegrationUnpaidInvoice,
  POSPaymentListPayload,
  POSPaymentOrderCreatePayload,
  POSPaymentReportPayload,
} from '@/model/payment-pos';

export interface POSPaymentQuery {
  building_id?: string;
  unit_id?: string;
}

// 1. 將 AJO POS 列表資料轉成 iSmart 帳目資料列
const toIntegrationTransactionRows = (items: POSPaymentListPayload['items'] | undefined): POSIntegrationPaymentTransaction[] =>
  (items ?? []) as POSIntegrationPaymentTransaction[];

// 1.1 彙總多個單位歷史，避免單一無效單位拖垮整頁
const collectPOSHistoryResponses = async (
  requests: Array<Promise<AxiosResponse<ApiResponse<POSPaymentListPayload>>>>,
): Promise<POSIntegrationTransactionsPayload> => {
  const results = await Promise.allSettled(requests);
  const fulfilled = results.filter((result): result is PromiseFulfilledResult<AxiosResponse<ApiResponse<POSPaymentListPayload>>> =>
    result.status === 'fulfilled',
  );

  if (fulfilled.length === 0) {
    const rejected = results.find((result): result is PromiseRejectedResult => result.status === 'rejected');
    throw rejected?.reason ?? new Error('payment history request failed');
  }

  return {
    payment_objs: fulfilled.flatMap(({ value }) => toIntegrationTransactionRows(value.data.data.items)),
  };
};

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
  params: { detail?: boolean; refresh?: boolean } & POSPaymentQuery = {},
) =>
  httpClient.get<ApiResponse<Record<string, unknown>>>(`/me/payments/pos/orders/${encodeURIComponent(mchOrderNo)}`, { params });

// 8. 按訂單號或支付單號查詢 POS H5 訂單
export const queryPOSPaymentOrder = (payload: Record<string, unknown>) =>
  httpClient.post<ApiResponse<Record<string, unknown>>>('/me/payments/pos/orders/query', payload);

// 9. 上報 POS 線下繳費
export const reportPOSPayment = (payload: POSPaymentReportPayload) =>
  httpClient.post<ApiResponse<Record<string, unknown>>>('/me/payments/pos/payments/report', payload);

// 10. 取得單位未繳賬單列表
export const fetchPOSIntegrationUnpaidInvoices = async (unitID: string): Promise<POSIntegrationUnpaidInvoice[]> => {
  const { data } = await httpClient.get<ApiResponse<POSPaymentListPayload>>('/me/payments/pos/bills', {
    params: { unit_id: unitID },
  });
  return (data.data.items ?? []) as unknown as POSIntegrationUnpaidInvoice[];
};

// 11. 按單位查詢交易
export const fetchPOSIntegrationTransactionsByUnit = async (
  unitIDs: string[],
): Promise<POSIntegrationTransactionsPayload> => {
  if (unitIDs.length === 0) {
    return { payment_objs: [] };
  }

  return collectPOSHistoryResponses(
    unitIDs.map((unitID) =>
      httpClient.get<ApiResponse<POSPaymentListPayload>>('/me/payments/pos/history', {
        params: { unit_id: unitID },
      }),
    ),
  );
};

// 12. 按日期查詢交易
export const fetchPOSIntegrationTransactionsByDate = async (
  query: POSIntegrationTransactionDateQuery,
  unitIDs: string[] = [],
): Promise<POSIntegrationTransactionsPayload> => {
  const buildParams = (unitID?: string) => {
    const params: POSPaymentQuery & Omit<POSIntegrationTransactionDateQuery, 'building_id'> = {
      from_date: query.from_date,
      to_date: query.to_date,
      date_type: query.date_type,
      pay_method: query.pay_method ?? 'all',
    };
    if (unitID) {
      params.unit_id = unitID;
    } else {
      params.building_id = query.building_id;
    }
    return params;
  };

  const normalizedUnitIDs = unitIDs.map((item) => item.trim()).filter(Boolean);
  if (normalizedUnitIDs.length > 0) {
    return collectPOSHistoryResponses(
      normalizedUnitIDs.map((unitID) =>
        httpClient.get<ApiResponse<POSPaymentListPayload>>('/me/payments/pos/history', {
          params: buildParams(unitID),
        }),
      ),
    );
  }

  const { data } = await httpClient.get<ApiResponse<POSPaymentListPayload>>('/me/payments/pos/history', {
    params: buildParams(),
  });
  return { payment_objs: toIntegrationTransactionRows(data.data.items) };
};

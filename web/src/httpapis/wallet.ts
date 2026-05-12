/*
 * AJO Point 錢包 API。
 * 1. 串接會員錢包、積分流水與廣告任務接口。
 * 2. 保持扣費與廣告領取結果由後端可信返回。
 */
import httpClient from '@/httpapis';
import type { ApiListData, ApiResponse, PaginatedResult } from '@/model/api';
import type {
  PointsChargeResponse,
  RewardAdClickResponse,
  RewardAdSessionResponse,
  RewardAdTaskResponse,
  StaffRewardAdPayload,
  StaffRewardAdResponse,
  StaffWalletGrantResponse,
  StaffWalletTransactionResponse,
  WalletOverviewResponse,
  WalletTransactionResponse,
} from '@/model/wallet';

// 1. 取得錢包總覽
export const fetchWalletOverview = () =>
  httpClient.get<ApiResponse<WalletOverviewResponse>>('/me/wallet');

// 2. 取得積分流水
export const fetchWalletTransactions = (params: { page?: number; page_size?: number } = {}) =>
  httpClient.get<ApiResponse<PaginatedResult<WalletTransactionResponse>>>(
    '/me/wallet/transactions',
    { params },
  );

// 3. 取得廣告任務
export const fetchRewardAdTasks = () =>
  httpClient.get<ApiResponse<ApiListData<RewardAdTaskResponse>>>('/me/wallet/ad-tasks');

// 4. 開始廣告觀看任務
export const startRewardAdTask = (taskId: string) =>
  httpClient.post<ApiResponse<RewardAdSessionResponse>>(`/me/wallet/ad-tasks/${taskId}/start`);

// 5. 領取廣告積分
export const claimRewardAdTask = (taskId: string, claimId: string) =>
  httpClient.post<ApiResponse<PointsChargeResponse>>(`/me/wallet/ad-tasks/${taskId}/claim`, {
    claim_id: claimId,
  });

// 6. 記錄廣告連結點擊
export const trackRewardAdClick = (taskId: string) =>
  httpClient.post<ApiResponse<RewardAdClickResponse>>(`/me/wallet/ad-tasks/${taskId}/click`);

// 7. Staff 查詢積分流水
export const fetchStaffWalletTransactions = (params: {
  page?: number;
  page_size?: number;
  user_id?: string;
  direction?: string;
  source_type?: string;
  biz_module?: string;
} = {}) =>
  httpClient.get<ApiResponse<PaginatedResult<StaffWalletTransactionResponse>>>(
    '/staff/wallet/transactions',
    { params },
  );

// 8. Staff 發放積分
export const grantStaffWalletPoints = (payload: {
  user_id: string;
  amount: number;
  note: string;
}) => httpClient.post<ApiResponse<StaffWalletGrantResponse>>('/staff/wallet/grants', payload);

// 9. Staff 查詢廣告任務
export const fetchStaffRewardAds = (params: {
  page?: number;
  page_size?: number;
  keyword?: string;
  is_active?: boolean;
} = {}) =>
  httpClient.get<ApiResponse<PaginatedResult<StaffRewardAdResponse>>>(
    '/staff/wallet/reward-ads',
    { params },
  );

// 10. Staff 建立廣告任務
export const createStaffRewardAd = (payload: StaffRewardAdPayload) =>
  httpClient.post<ApiResponse<StaffRewardAdResponse>>('/staff/wallet/reward-ads', payload);

// 11. Staff 取得單個廣告任務
export const fetchStaffRewardAd = (taskId: string) =>
  httpClient.get<ApiResponse<StaffRewardAdResponse>>(`/staff/wallet/reward-ads/${taskId}`);

// 12. Staff 更新廣告任務
export const updateStaffRewardAd = (taskId: string, payload: StaffRewardAdPayload) =>
  httpClient.patch<ApiResponse<StaffRewardAdResponse>>(`/staff/wallet/reward-ads/${taskId}`, payload);

// 13. Staff 續期廣告任務
export const renewStaffRewardAd = (taskId: string, payload: { retention_days: number }) =>
  httpClient.patch<ApiResponse<StaffRewardAdResponse>>(`/staff/wallet/reward-ads/${taskId}`, {
    retention_days: payload.retention_days,
    is_active: true,
  });

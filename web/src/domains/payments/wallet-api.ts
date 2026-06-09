/*
 * AJO Point 錢包 API。
 * 1. 串接會員錢包、積分流水與廣告任務接口。
 * 2. 保持扣費與廣告領取結果由後端可信返回。
 */
import httpClient from '@/shared/utils/http';
import type { ApiListData, ApiResponse, PaginatedResult } from '@/shared/utils/http/model';
import type {
  DisplayAdChannelSettingsResponse,
  DisplayAdSlotSaveItem,
  PointsChargeResponse,
  PublicDisplayAdResponse,
  RewardAdClickResponse,
  RewardAdSessionResponse,
  RewardAdTaskResponse,
  StaffRewardAdPayload,
  StaffRewardAdResponse,
  StaffWalletGrantResponse,
  StaffWalletTransactionResponse,
  WalletOverviewResponse,
  WalletRechargeDeviceMode,
  WalletRechargeOrderResponse,
  WalletRechargePayMethod,
  WalletRechargePayRegion,
  WalletTransactionResponse,
} from '@/domains/payments/model';

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
  ad_type?: string;
  display_channel?: string;
} = {}) =>
  httpClient.get<ApiResponse<PaginatedResult<StaffRewardAdResponse>>>(
    '/staff/wallet/reward-ads',
    { params },
  );

// 10. Staff 查詢展示廣告位設定
export const fetchDisplayAdSettings = (channel: 'property_sale' | 'serviced_apartment' | 'furniture') =>
  httpClient.get<ApiResponse<DisplayAdChannelSettingsResponse>>('/staff/wallet/display-ad-settings', {
    params: { channel },
  });

// 11. Staff 儲存展示廣告位設定
export const saveDisplayAdSettings = (
  channel: 'property_sale' | 'serviced_apartment' | 'furniture',
  slots: DisplayAdSlotSaveItem[],
) =>
  httpClient.put<ApiResponse<DisplayAdChannelSettingsResponse>>(
    '/staff/wallet/display-ad-settings',
    { slots },
    { params: { channel } },
  );

// 12. Staff 建立廣告任務
export const createStaffRewardAd = (payload: StaffRewardAdPayload) =>
  httpClient.post<ApiResponse<StaffRewardAdResponse>>('/staff/wallet/reward-ads', payload);

// 13. Staff 取得單個廣告任務
export const fetchStaffRewardAd = (taskId: string) =>
  httpClient.get<ApiResponse<StaffRewardAdResponse>>(`/staff/wallet/reward-ads/${taskId}`);

// 14. Staff 更新廣告任務
export const updateStaffRewardAd = (taskId: string, payload: StaffRewardAdPayload) =>
  httpClient.patch<ApiResponse<StaffRewardAdResponse>>(`/staff/wallet/reward-ads/${taskId}`, payload);

// 15. 取得公開展示廣告
export const fetchPublicDisplayAds = (params: {
  channel: 'property_sale' | 'serviced_apartment' | 'furniture';
  placement?: 'listing_side';
  limit?: number;
}) =>
  httpClient.get<ApiResponse<ApiListData<PublicDisplayAdResponse>>>('/public/ads', {
    params: {
      placement: 'listing_side',
      limit: 10,
      ...params,
    },
  });

// 16. 建立錢包充值訂單
export const createWalletRechargeOrder = (payload: {
  amount_hkd: number;
  pay_method: WalletRechargePayMethod;
  pay_region: WalletRechargePayRegion;
  device_mode: WalletRechargeDeviceMode;
  return_path: string;
}) => httpClient.post<ApiResponse<WalletRechargeOrderResponse>>('/me/wallet/recharge-orders', payload);

// 17. 取得錢包充值訂單
export const fetchWalletRechargeOrder = (orderId: string, params: { refresh?: boolean } = {}) =>
  httpClient.get<ApiResponse<WalletRechargeOrderResponse>>(`/me/wallet/recharge-orders/${orderId}`, {
    params,
  });

// 18. Staff 續期廣告任務
export const renewStaffRewardAd = (taskId: string, payload: { retention_days: number }) =>
  httpClient.patch<ApiResponse<StaffRewardAdResponse>>(`/staff/wallet/reward-ads/${taskId}`, {
    retention_days: payload.retention_days,
    is_active: true,
  });

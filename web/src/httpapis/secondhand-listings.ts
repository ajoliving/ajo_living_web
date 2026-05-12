/*
 * 二手帖子 API。
 * 1. 對齊後端 `/listings`、`/me/secondhand/listings` 與聯絡權限契約。
 * 2. 提供列表、詳情、建立、更新與狀態流轉接口。
 */
import httpClient from '@/httpapis';
import type { ApiResponse, PaginatedResult } from '@/model/api';
import type {
  ContactAccessResult,
  DiscoverPayloadResponse,
  DiscoverPlacementPayload,
  FavoriteResult,
  ListingListParams,
  MyListingListParams,
  SettingsListingListParams,
  SecondhandListingDetailResponse,
  SecondhandListingSummaryResponse,
  UpsertSecondhandListingPayload,
} from '@/model/marketplace';

// 1. 查詢公開二手帖子列表
export const fetchSecondhandListings = (params: ListingListParams = {}) =>
  httpClient.get<ApiResponse<PaginatedResult<SecondhandListingSummaryResponse>>>('/listings', {
    params,
  });

// 2. 取得二手發現頁配置
export const fetchSecondhandDiscover = () =>
  httpClient.get<ApiResponse<DiscoverPayloadResponse>>('/secondhand/discover');

// 3. 取得二手帖子詳情
export const fetchSecondhandListingDetail = (listingId: string) =>
  httpClient.get<ApiResponse<SecondhandListingDetailResponse>>(`/listings/${listingId}`);

// 4. 建立二手帖子草稿
export const createSecondhandListing = (
  payload: UpsertSecondhandListingPayload,
  params: { charge_draft?: boolean } = {},
) => httpClient.post<ApiResponse<SecondhandListingDetailResponse>>('/listings', payload, { params });

// 5. 更新二手帖子
export const updateSecondhandListing = (
  listingId: string,
  payload: UpsertSecondhandListingPayload,
  params: { charge_draft?: boolean } = {},
) => httpClient.patch<ApiResponse<SecondhandListingDetailResponse>>(`/listings/${listingId}`, payload, { params });

// 6. 發布二手帖子
export const publishSecondhandListing = (listingId: string) =>
  httpClient.post<ApiResponse<SecondhandListingDetailResponse>>(`/listings/${listingId}/publish`);

// 7. 重新發布二手帖子
export const republishSecondhandListing = (listingId: string) =>
  httpClient.post<ApiResponse<SecondhandListingDetailResponse>>(`/listings/${listingId}/republish`);

// 8. 續期二手帖子
export const renewSecondhandListing = (listingId: string) =>
  httpClient.post<ApiResponse<SecondhandListingDetailResponse>>(`/listings/${listingId}/renew`);

// 9. 標記帖子售出
export const markSecondhandListingSold = (listingId: string) =>
  httpClient.post<ApiResponse<{ listing_id: string; business_status: string }>>(
    `/listings/${listingId}/mark-sold`,
  );

// 10. 下架帖子
export const deactivateSecondhandListing = (listingId: string) =>
  httpClient.post<ApiResponse<{ listing_id: string; publication_status: string }>>(
    `/listings/${listingId}/deactivate`,
  );

// 11. 取得我的二手帖子
export const fetchMySecondhandListings = (params: MyListingListParams = {}) =>
  httpClient.get<ApiResponse<PaginatedResult<SecondhandListingSummaryResponse>>>(
    '/me/secondhand/listings',
    { params },
  );

// 12. 取得我的收藏二手帖子
export const fetchMyFavoriteSecondhandListings = (params: MyListingListParams = {}) =>
  httpClient.get<ApiResponse<PaginatedResult<SecondhandListingSummaryResponse>>>(
    '/me/secondhand/favorites',
    { params },
  );

// 13. 收藏二手帖子
export const favoriteSecondhandListing = (listingId: string) =>
  httpClient.post<ApiResponse<FavoriteResult>>(`/listings/${listingId}/favorite`);

// 14. 取消收藏二手帖子
export const unfavoriteSecondhandListing = (listingId: string) =>
  httpClient.delete<ApiResponse<FavoriteResult>>(`/listings/${listingId}/favorite`);

// 15. 取得設定頁全部二手帖子
export const fetchSettingsSecondhandListings = (params: SettingsListingListParams = {}) =>
  httpClient.get<ApiResponse<PaginatedResult<SecondhandListingSummaryResponse>>>(
    '/secondhand/settings/listings',
    { params },
  );

// 16. 取得設定頁發現廣告位
export const fetchSettingsDiscoverPlacements = () =>
  httpClient.get<ApiResponse<DiscoverPayloadResponse>>('/secondhand/settings/discover-placements');

// 17. 儲存設定頁發現廣告位
export const saveSettingsDiscoverPlacements = (placements: DiscoverPlacementPayload[]) =>
  httpClient.put<ApiResponse<DiscoverPayloadResponse>>('/secondhand/settings/discover-placements', {
    placements,
  });

// 18. 設定頁標記任意帖子售出
export const markSettingsSecondhandListingSold = (listingId: string) =>
  httpClient.post<ApiResponse<{ listing_id: string; business_status: string }>>(
    `/secondhand/settings/listings/${listingId}/mark-sold`,
  );

// 19. 設定頁下架任意帖子
export const deactivateSettingsSecondhandListing = (listingId: string) =>
  httpClient.post<ApiResponse<{ listing_id: string; publication_status: string }>>(
    `/secondhand/settings/listings/${listingId}/deactivate`,
  );

// 20. 驗證聯絡方式權限
export const fetchListingContactAccess = (listingId: string) =>
  httpClient.post<ApiResponse<ContactAccessResult>>(`/listings/${listingId}/contact-access`);

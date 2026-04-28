/*
 * 二手帖子 API。
 * 1. 對齊後端 `/listings`、`/me/secondhand/listings` 與聯絡權限契約。
 * 2. 提供列表、詳情、建立、更新與狀態流轉接口。
 */
import httpClient from '@/httpapis';
import type { ApiResponse, PaginatedResult } from '@/model/api';
import type {
  ContactAccessResult,
  ListingListParams,
  MyListingListParams,
  SecondhandListingDetailResponse,
  SecondhandListingSummaryResponse,
  UpsertSecondhandListingPayload,
} from '@/model/marketplace';

// 1. 查詢公開二手帖子列表
export const fetchSecondhandListings = (params: ListingListParams = {}) =>
  httpClient.get<ApiResponse<PaginatedResult<SecondhandListingSummaryResponse>>>('/listings', {
    params,
  });

// 2. 取得二手帖子詳情
export const fetchSecondhandListingDetail = (listingId: string) =>
  httpClient.get<ApiResponse<SecondhandListingDetailResponse>>(`/listings/${listingId}`);

// 3. 建立二手帖子草稿
export const createSecondhandListing = (payload: UpsertSecondhandListingPayload) =>
  httpClient.post<ApiResponse<SecondhandListingDetailResponse>>('/listings', payload);

// 4. 更新二手帖子
export const updateSecondhandListing = (listingId: string, payload: UpsertSecondhandListingPayload) =>
  httpClient.patch<ApiResponse<SecondhandListingDetailResponse>>(`/listings/${listingId}`, payload);

// 5. 發布二手帖子
export const publishSecondhandListing = (listingId: string) =>
  httpClient.post<ApiResponse<SecondhandListingDetailResponse>>(`/listings/${listingId}/publish`);

// 6. 重新發布二手帖子
export const republishSecondhandListing = (listingId: string) =>
  httpClient.post<ApiResponse<SecondhandListingDetailResponse>>(`/listings/${listingId}/republish`);

// 7. 標記帖子售出
export const markSecondhandListingSold = (listingId: string) =>
  httpClient.post<ApiResponse<{ listing_id: string; business_status: string }>>(
    `/listings/${listingId}/mark-sold`,
  );

// 8. 下架帖子
export const deactivateSecondhandListing = (listingId: string) =>
  httpClient.post<ApiResponse<{ listing_id: string; publication_status: string }>>(
    `/listings/${listingId}/deactivate`,
  );

// 9. 取得我的二手帖子
export const fetchMySecondhandListings = (params: MyListingListParams = {}) =>
  httpClient.get<ApiResponse<PaginatedResult<SecondhandListingSummaryResponse>>>(
    '/me/secondhand/listings',
    { params },
  );

// 10. 驗證聯絡方式權限
export const fetchListingContactAccess = (listingId: string) =>
  httpClient.post<ApiResponse<ContactAccessResult>>(`/listings/${listingId}/contact-access`);

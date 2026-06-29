/*
 * 物業頻道 API。
 * 1. 串接樓盤放售與服務式住宅列表、詳情與發布接口。
 * 2. 提供我的列表、狀態操作與聯絡方式授權。
 */
import httpClient from '@/httpapis';
import type { ApiResponse, PaginatedResult } from '@/model/api';
import type {
  ContactAccessResult,
  PropertyActionResult,
  PropertyAddressSuggestion,
  PropertyAppointmentPayload,
  PropertyAppointmentResponse,
  PropertyListParams,
  PropertyListingDetailResponse,
  PropertyListingSummaryResponse,
  PropertyReportPayload,
  PropertyReportResponse,
  UpsertPropertySalePayload,
  UpsertServicedApartmentPayload,
} from '@/model/property';

// 1. 查詢公開樓盤放售列表
export const fetchPropertySaleListings = (params: PropertyListParams = {}) =>
  httpClient.get<ApiResponse<PaginatedResult<PropertyListingSummaryResponse>>>(
    '/property-sales',
    { params },
  );

// 2. 取得樓盤放售詳情
export const fetchPropertySaleDetail = (listingId: string) =>
  httpClient.get<ApiResponse<PropertyListingDetailResponse>>(`/property-sales/${listingId}`);

// 2.1 查詢相似樓盤
export const fetchSimilarPropertySales = (listingId: string, params: { limit?: number } = {}) =>
  httpClient.get<ApiResponse<{ items: PropertyListingSummaryResponse[] }>>(
    `/property-sales/${listingId}/similar`,
    { params },
  );

// 3. 建立樓盤放售草稿
export const createPropertySale = (payload: UpsertPropertySalePayload) =>
  httpClient.post<ApiResponse<PropertyListingDetailResponse>>('/property-sales', payload);

// 4. 更新樓盤放售
export const updatePropertySale = (listingId: string, payload: UpsertPropertySalePayload) =>
  httpClient.patch<ApiResponse<PropertyListingDetailResponse>>(
    `/property-sales/${listingId}`,
    payload,
  );

// 5. 發布樓盤放售
export const publishPropertySale = (listingId: string) =>
  httpClient.post<ApiResponse<PropertyListingDetailResponse>>(
    `/property-sales/${listingId}/publish`,
  );

// 6. 重新發布樓盤放售
export const republishPropertySale = (listingId: string) =>
  httpClient.post<ApiResponse<PropertyListingDetailResponse>>(
    `/property-sales/${listingId}/republish`,
  );

// 7. 標記樓盤成交
export const markPropertySaleSold = (listingId: string) =>
  httpClient.post<ApiResponse<{ listing_id: string; business_status: string }>>(
    `/property-sales/${listingId}/mark-sold`,
  );

// 8. 下架樓盤放售
export const deactivatePropertySale = (listingId: string) =>
  httpClient.post<ApiResponse<{ listing_id: string; publication_status: string }>>(
    `/property-sales/${listingId}/deactivate`,
  );

// 8.1 收藏樓盤放售
export const favoritePropertySale = (listingId: string) =>
  httpClient.post<ApiResponse<PropertyActionResult>>(
    `/property-sales/${listingId}/favorite`,
  );

// 8.2 取消收藏樓盤放售
export const unfavoritePropertySale = (listingId: string) =>
  httpClient.delete<ApiResponse<PropertyActionResult>>(
    `/property-sales/${listingId}/favorite`,
  );

// 8.3 預約睇樓
export const createPropertyAppointment = (
  listingId: string,
  payload: PropertyAppointmentPayload,
) =>
  httpClient.post<ApiResponse<PropertyAppointmentResponse>>(
    `/property-sales/${listingId}/appointments`,
    payload,
  );

// 8.4 舉報樓盤
export const reportPropertySale = (
  listingId: string,
  payload: PropertyReportPayload,
) =>
  httpClient.post<ApiResponse<PropertyReportResponse>>(
    `/property-sales/${listingId}/reports`,
    payload,
  );

// 9. 取得我的樓盤放售
export const fetchMyPropertySaleListings = (params: PropertyListParams = {}) =>
  httpClient.get<ApiResponse<PaginatedResult<PropertyListingSummaryResponse>>>(
    '/me/property-sales',
    { params },
  );

// 9.1 取得我的樓盤收藏
export const fetchMyFavoritePropertySales = (params: PropertyListParams = {}) =>
  httpClient.get<ApiResponse<PaginatedResult<PropertyListingSummaryResponse>>>(
    '/me/property-sales/favorites',
    { params },
  );

// 10. 解鎖樓盤放售聯絡方式
export const fetchPropertySaleContactAccess = (listingId: string) =>
  httpClient.post<ApiResponse<ContactAccessResult>>(
    `/property-sales/${listingId}/contact-access`,
  );

// 11. 查詢樓盤地址聯想
export const searchPropertyAddresses = (params: { keyword: string; district_code?: string; limit?: number }) =>
  httpClient.get<ApiResponse<PropertyAddressSuggestion[]>>(
    '/property-addresses/search',
    { params },
  );

// 12. 查詢公開服務式住宅列表
export const fetchServicedApartmentListings = (params: PropertyListParams = {}) =>
  httpClient.get<ApiResponse<PaginatedResult<PropertyListingSummaryResponse>>>(
    '/serviced-apartments',
    { params },
  );

// 13. 取得服務式住宅詳情
export const fetchServicedApartmentDetail = (listingId: string) =>
  httpClient.get<ApiResponse<PropertyListingDetailResponse>>(
    `/serviced-apartments/${listingId}`,
  );

// 14. 建立服務式住宅草稿
export const createServicedApartment = (payload: UpsertServicedApartmentPayload) =>
  httpClient.post<ApiResponse<PropertyListingDetailResponse>>(
    '/serviced-apartments',
    payload,
  );

// 15. 更新服務式住宅
export const updateServicedApartment = (
  listingId: string,
  payload: UpsertServicedApartmentPayload,
) =>
  httpClient.patch<ApiResponse<PropertyListingDetailResponse>>(
    `/serviced-apartments/${listingId}`,
    payload,
  );

// 16. 發布服務式住宅
export const publishServicedApartment = (listingId: string) =>
  httpClient.post<ApiResponse<PropertyListingDetailResponse>>(
    `/serviced-apartments/${listingId}/publish`,
  );

// 17. 重新發布服務式住宅
export const republishServicedApartment = (listingId: string) =>
  httpClient.post<ApiResponse<PropertyListingDetailResponse>>(
    `/serviced-apartments/${listingId}/republish`,
  );

// 18. 下架服務式住宅
export const deactivateServicedApartment = (listingId: string) =>
  httpClient.post<ApiResponse<{ listing_id: string; publication_status: string }>>(
    `/serviced-apartments/${listingId}/deactivate`,
  );

// 19. 取得我的服務式住宅
export const fetchMyServicedApartmentListings = (params: PropertyListParams = {}) =>
  httpClient.get<ApiResponse<PaginatedResult<PropertyListingSummaryResponse>>>(
    '/me/serviced-apartments',
    { params },
  );

// 20. 解鎖服務式住宅聯絡方式
export const fetchServicedApartmentContactAccess = (listingId: string) =>
  httpClient.post<ApiResponse<ContactAccessResult>>(
    `/serviced-apartments/${listingId}/contact-access`,
  );

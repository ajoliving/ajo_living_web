/*
 * Staff 管理 API。
 * 1. 串接 Staff 會員、帖子、樓盤與服務式住宅管理列表接口。
 * 2. 提供 Staff-only 狀態操作。
 */
import httpClient from '@/httpapis';
import type { ApiResponse, PaginatedResult } from '@/model/api';
import type { SecondhandListingSummaryResponse } from '@/model/marketplace';
import type { PropertyListingSummaryResponse } from '@/model/property';
import type { RoleCatalogItem, StaffUserCreatePayload, StaffUserSummary } from '@/model/user';

export interface StaffListParams {
  page?: number;
  page_size?: number;
  keyword?: string;
  status?: string;
}

export interface StaffUserListParams extends StaffListParams {
  member_type?: string;
  role_code?: string;
  is_staff?: boolean;
}

export interface StaffListingActionResult {
  listing_id: string;
  publication_status?: string;
  business_status?: string;
}

export interface StaffRenewSecondhandListingPayload {
  renewal_days: number;
}

// 1. Staff 查詢會員列表
export const fetchStaffUsers = (params: StaffUserListParams = {}) =>
  httpClient.get<ApiResponse<PaginatedResult<StaffUserSummary>>>('/staff/users', {
    params,
  });

// 2. Staff 查詢角色列表
export const fetchStaffRoles = () =>
  httpClient.get<ApiResponse<{ items: RoleCatalogItem[] }>>('/staff/roles');

// 3. Staff 建立會員帳戶
export const createStaffUser = (payload: StaffUserCreatePayload) =>
  httpClient.post<ApiResponse<StaffUserSummary>>('/staff/users', payload);

// 4. Staff 查詢二手帖子列表
export const fetchStaffSecondhandListings = (params: StaffListParams = {}) =>
  httpClient.get<ApiResponse<PaginatedResult<SecondhandListingSummaryResponse>>>(
    '/staff/secondhand/listings',
    { params },
  );

// 5. Staff 查詢樓盤放售列表
export const fetchStaffPropertySales = (params: StaffListParams = {}) =>
  httpClient.get<ApiResponse<PaginatedResult<PropertyListingSummaryResponse>>>(
    '/staff/property-sales',
    { params },
  );

// 6. Staff 查詢服務式住宅列表
export const fetchStaffServicedApartments = (params: StaffListParams = {}) =>
  httpClient.get<ApiResponse<PaginatedResult<PropertyListingSummaryResponse>>>(
    '/staff/serviced-apartments',
    { params },
  );

// 7. Staff 上架二手帖子
export const publishStaffSecondhandListing = (listingId: string) =>
  httpClient.post<ApiResponse<StaffListingActionResult>>(
    `/staff/secondhand/listings/${listingId}/publish`,
  );

// 8. Staff 下架二手帖子
export const deactivateStaffSecondhandListing = (listingId: string) =>
  httpClient.post<ApiResponse<StaffListingActionResult>>(
    `/staff/secondhand/listings/${listingId}/deactivate`,
  );

// 9. Staff 續期二手帖子
export const renewStaffSecondhandListing = (
  listingId: string,
  payload?: StaffRenewSecondhandListingPayload,
) =>
  httpClient.post<ApiResponse<StaffListingActionResult>>(
    `/staff/secondhand/listings/${listingId}/renew`,
    payload,
  );

// 10. Staff 上架樓盤放售
export const publishStaffPropertySale = (listingId: string) =>
  httpClient.post<ApiResponse<StaffListingActionResult>>(
    `/staff/property-sales/${listingId}/publish`,
  );

// 11. Staff 下架樓盤放售
export const deactivateStaffPropertySale = (listingId: string) =>
  httpClient.post<ApiResponse<StaffListingActionResult>>(
    `/staff/property-sales/${listingId}/deactivate`,
  );

// 12. Staff 續期樓盤放售
export const renewStaffPropertySale = (listingId: string) =>
  httpClient.post<ApiResponse<StaffListingActionResult>>(
    `/staff/property-sales/${listingId}/renew`,
  );

// 13. Staff 續期服務式住宅
export const renewStaffServicedApartment = (listingId: string) =>
  httpClient.post<ApiResponse<StaffListingActionResult>>(
    `/staff/serviced-apartments/${listingId}/renew`,
  );

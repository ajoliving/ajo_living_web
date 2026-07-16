/*
 * 地產代理資料 API。
 * 1. 提供會員代理資料版本與公司子帳戶接口。
 * 2. 提供 Staff 代理資料審核列表、詳情及審核接口。
 */
import httpClient from '@/httpapis';
import type { ApiResponse, PaginatedResult } from '@/model/api';
import type {
  AgencyProfileUpsertPayload,
  AgencySubaccount,
  AgencySubaccountStatus,
  CreateAgencySubaccountPayload,
  CurrentAgencyProfileResponse,
  ReviewAgencyProfilePayload,
  StaffAgencyProfile,
  StaffAgencyProfileListParams,
} from '@/model/agency-profile';

// 1. 會員代理資料
export const fetchMyAgencyProfile = () =>
  httpClient.get<ApiResponse<CurrentAgencyProfileResponse>>('/me/agency-profile');

export const createMyAgencyProfile = (payload: AgencyProfileUpsertPayload) =>
  httpClient.post<ApiResponse<CurrentAgencyProfileResponse>>('/me/agency-profile', payload);

export const updateMyAgencyProfile = (payload: AgencyProfileUpsertPayload) =>
  httpClient.patch<ApiResponse<CurrentAgencyProfileResponse>>('/me/agency-profile', payload);

export const submitMyAgencyProfile = () =>
  httpClient.post<ApiResponse<CurrentAgencyProfileResponse>>('/me/agency-profile/submit');

// 2. 公司子帳戶
export const fetchAgencySubaccounts = () =>
  httpClient.get<ApiResponse<{ items: AgencySubaccount[] }>>('/me/agency-profile/subaccounts');

export const createAgencySubaccount = (payload: CreateAgencySubaccountPayload) =>
  httpClient.post<ApiResponse<AgencySubaccount>>('/me/agency-profile/subaccounts', payload);

export const updateAgencySubaccountStatus = (publicId: string, status: AgencySubaccountStatus) =>
  httpClient.patch<ApiResponse<AgencySubaccount>>(`/me/agency-profile/subaccounts/${publicId}/status`, { status });

export const removeAgencySubaccount = (publicId: string) =>
  httpClient.delete<ApiResponse<{ deleted: boolean }>>(`/me/agency-profile/subaccounts/${publicId}`);

// 3. Staff 審核
export const fetchStaffAgencyProfiles = (params: StaffAgencyProfileListParams = {}) =>
  httpClient.get<ApiResponse<PaginatedResult<StaffAgencyProfile>>>('/staff/agency-profiles', { params });

export const fetchStaffAgencyProfileDetail = (profileId: string) =>
  httpClient.get<ApiResponse<StaffAgencyProfile>>(`/staff/agency-profiles/${profileId}`);

export const reviewStaffAgencyProfile = (profileId: string, payload: ReviewAgencyProfilePayload) =>
  httpClient.post<ApiResponse<StaffAgencyProfile>>(`/staff/agency-profiles/${profileId}/review`, payload);

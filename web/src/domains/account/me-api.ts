/*
 * 會員 API。
 * 1. 串接當前會員資料查詢與更新接口。
 * 2. 對齊後端 `/me` 與 `/me/profile` 契約。
 */
import httpClient from '@/shared/utils/http';
import type { ApiResponse } from '@/shared/utils/http/model';
import type { CurrentMemberProfile, IsmartLoginPayload } from '@/domains/account/model';

// 1. 取得當前會員資料
export const fetchMe = () => httpClient.get<ApiResponse<CurrentMemberProfile>>('/me');

// 2. 更新當前會員資料
export const updateMe = (payload: {
  display_name?: string;
  email?: string;
  email_otp_code?: string;
  phone_country_code?: string;
  phone_number?: string;
  password?: string;
  publisher_identity_type?: string;
  primary_community_id?: string;
  primary_community_name?: string;
  bound_building_ids?: string[];
  bound_flat_unit_ids?: string[];
  residence_floor?: string;
  residence_unit?: string;
  district_code?: string;
  avatar_asset_id?: string;
}) => httpClient.patch<ApiResponse<CurrentMemberProfile>>('/me/profile', payload);

// 3. 綁定目前會員的 ismart 帳戶
export const bindCurrentUserIsmart = (payload: IsmartLoginPayload) =>
  httpClient.post<ApiResponse<CurrentMemberProfile>>('/me/ismart/bind', payload);

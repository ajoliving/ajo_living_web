/*
 * 會員 API。
 * 1. 串接當前會員資料查詢與更新接口。
 * 2. 對齊後端 `/me` 與 `/me/profile` 契約。
 */
import httpClient from '@/httpapis';
import type { ApiResponse } from '@/model/api';
import type { IsmartLoginPayload } from '@/model/auth';
import type { CurrentMemberProfile } from '@/model/user';

// 2.1 iSmart ClientTbl 更新資料
export interface UpdateMemberIsmartProfilePayload {
  account_email?: string;
  account_phone?: string;
  contact_name?: string;
  emergency_contact_name?: string;
  owner_name_en?: string;
  owner_name_zh?: string;
  account_name?: string;
  identity_number?: string;
  contact_phone?: string;
  emergency_contact_phone?: string;
  contact_email?: string;
  birth_date?: string;
  gender?: string;
  billing_address_en?: string;
  billing_address_zh?: string;
}

// 1. 取得當前會員資料
export const fetchMe = () => httpClient.get<ApiResponse<CurrentMemberProfile>>('/me');

// 2. 更新當前會員資料
export const updateMe = (payload: {
  display_name?: string;
  phone_country_code?: string;
  phone_number?: string;
  primary_community_id?: string;
  primary_community_name?: string;
  bound_building_ids?: string[];
  bound_flat_unit_ids?: string[];
  district_code?: string;
  avatar_asset_id?: string;
  residence_floor?: string;
  residence_unit?: string;
}) => httpClient.patch<ApiResponse<CurrentMemberProfile>>('/me/profile', payload);

// 3. 綁定目前會員的 ismart 帳戶
export const bindCurrentUserIsmart = (payload: IsmartLoginPayload) =>
  httpClient.post<ApiResponse<CurrentMemberProfile>>('/me/ismart/bind', payload);

// 4. 更新目前會員可修改的 iSmart 資料
export const updateMemberIsmartProfile = (payload: UpdateMemberIsmartProfilePayload) =>
  httpClient.patch<ApiResponse<CurrentMemberProfile>>('/me/ismart/profile', payload);

// 5. 修改目前會員的 iSmart 登入密碼
export const changeMemberIsmartPassword = (payload: {
  old_password: string;
  new_password: string;
  new_password_confirm: string;
}) => httpClient.post<ApiResponse<Record<string, unknown>>>('/me/ismart/change-password', payload);

// 6. 取得目前會員的 iSmart 通知電郵設定
export const fetchMemberIsmartNotificationSettings = () =>
  httpClient.get<ApiResponse<{ blg_notice_email: boolean; is_receive_email: boolean }>>('/me/ismart/notification-settings');

// 7. 更新目前會員的 iSmart 通知電郵設定
export const updateMemberIsmartNotificationSettings = (isReceiveEmail: boolean) =>
  httpClient.patch<ApiResponse<{ blg_notice_email: boolean; is_receive_email: boolean }>>(
    '/me/ismart/notification-settings',
    { is_receive_email: isReceiveEmail },
  );

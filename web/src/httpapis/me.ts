/*
 * 會員 API。
 * 1. 串接當前會員資料查詢與更新接口。
 * 2. 對齊後端 `/me` 與 `/me/profile` 契約。
 */
import httpClient from '@/httpapis';
import type { ApiResponse } from '@/model/api';
import type { CurrentMemberProfile } from '@/model/user';

// 1. 取得當前會員資料
export const fetchMe = () => httpClient.get<ApiResponse<CurrentMemberProfile>>('/me');

// 2. 更新當前會員資料
export const updateMe = (payload: {
  display_name?: string;
  phone_country_code?: string;
  phone_number?: string;
  publisher_identity_type?: string;
  primary_community_id?: string;
  district_code?: string;
  avatar_asset_id?: string;
}) => httpClient.patch<ApiResponse<CurrentMemberProfile>>('/me/profile', payload);

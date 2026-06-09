/*
 * 社區 API。
 * 1. 串接真實社區清單查詢接口。
 * 2. 供登入後選擇所屬屋苑或大廈流程使用。
 */
import httpClient from '@/shared/utils/http';
import type { ApiListData, ApiResponse } from '@/shared/utils/http/model';
import type { MetaCommunity } from '@/domains/building/model';

// 1. 取得社區清單
export const fetchCommunities = () =>
  httpClient.get<ApiResponse<ApiListData<MetaCommunity>>>('/meta/communities');

// 2. 匯出 POS 大廈與單位 API
export * from './pos-api';

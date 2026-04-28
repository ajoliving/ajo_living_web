/*
 * 社區 API。
 * 1. 串接真實社區清單查詢接口。
 * 2. 供登入後選擇所屬屋苑或大廈流程使用。
 */
import httpClient from '@/httpapis';
import type { ApiListData, ApiResponse } from '@/model/api';
import type { MetaCommunity } from '@/model/community';

// 1. 取得社區清單
export const fetchCommunities = () =>
  httpClient.get<ApiResponse<ApiListData<MetaCommunity>>>('/meta/communities');

/*
 * 首頁 API。
 * 1. 串接頻道首頁摘要介面。
 * 2. 讀取香港或目前大廈的環境資料。
 */
import httpClient from '@/httpapis';
import type { ApiResponse } from '@/model/api';
import type { ChannelHomeOverview, HomeEnvironment } from '@/model/home';

// 1. 取得首頁摘要
export const fetchChannelHomeOverview = () =>
  httpClient.get<ApiResponse<ChannelHomeOverview>>('/channel-home/overview');

// 2. 只提交語系與目前大廈 ID，位置及權限由後端解析。
export const fetchHomeEnvironment = (
  params: { lang: 'tc' | 'en'; building_id?: string; estate?: 'seaview' | 'summit' | 'fleur'; detail?: 'forecast' },
  signal: AbortSignal,
) => httpClient.get<ApiResponse<HomeEnvironment>>('/home/environment', { params, signal });

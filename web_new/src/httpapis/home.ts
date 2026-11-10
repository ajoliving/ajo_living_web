/*
 * 首頁 API。
 * 1. 串接頻道首頁摘要介面。
 * 2. 提供首頁入口與精選二手清單資料。
 */
import httpClient from '@/httpapis';
import type { ApiResponse } from '@/model/api';
import type { ChannelHomeOverview } from '@/model/home';

// 1. 取得首頁摘要
export const fetchChannelHomeOverview = () =>
  httpClient.get<ApiResponse<ChannelHomeOverview>>('/channel-home/overview');

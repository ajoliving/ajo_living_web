/*
 * 首頁 API。
 * 1. 串接頻道首頁摘要介面。
 * 2. 提供首頁入口與精選二手清單資料。
 */
import httpClient from '@/shared/utils/http';
import type { ApiResponse } from '@/shared/utils/http/model';
import type { ChannelHomeOverview } from '@/domains/home/model';

// 1. 取得首頁摘要
export const fetchChannelHomeOverview = () =>
  httpClient.get<ApiResponse<ChannelHomeOverview>>('/channel-home/overview');

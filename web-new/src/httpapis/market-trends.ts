/*
 * 市場走勢 API。
 * 1. 讀取後端聚合的公開租金走勢。
 * 2. 避免前端直接依賴外部公開 CSV。
 */
import httpClient from '@/httpapis';
import type { ApiResponse } from '@/model/api';
import type { MarketRentTrendResponse } from '@/model/market-trend';

// 1. 取得香港私人住宅租金走勢
export const fetchMarketRentTrend = () =>
  httpClient.get<ApiResponse<MarketRentTrendResponse>>('/market-trends/rent');

/*
 * 超市優惠 API。
 * 1. 串接 AJO 後端公開 good-price 代理介面。
 * 2. 串接 AJO 會員收藏介面。
 * 3. 保持前端不直接連接舊 good-price 服務。
 */
import httpClient from '@/httpapis';
import type { ApiResponse } from '@/model/api';
import type { PaginationMeta } from '@/model/api';
import type {
  SupermarketProduct,
  SupermarketProductDetail,
  SupermarketSearchParams,
  SupermarketSearchResult,
  SupermarketSummary,
} from '@/model/supermarket-offers';

// 1. 取得超市優惠摘要
export const fetchSupermarketSummary = () =>
  httpClient.get<ApiResponse<SupermarketSummary>>('/supermarket-offers/summary');

// 2. 搜尋超市商品
export const searchSupermarketProducts = (params: SupermarketSearchParams) =>
  httpClient.get<ApiResponse<SupermarketSearchResult>>('/supermarket-offers/search', { params });

// 3. 取得商品詳情
export const fetchSupermarketProductDetail = (code: string, days = 90) =>
  httpClient.get<ApiResponse<SupermarketProductDetail>>(`/supermarket-offers/products/${encodeURIComponent(code)}`, {
    params: { days },
  });

// 4. 新增目前會員收藏
export const fetchSupermarketFavorites = (params: { page?: number; pageSize?: number } = {}) =>
  httpClient.get<ApiResponse<{ items: SupermarketProduct[]; pagination: PaginationMeta }>>(
    '/me/supermarket-offers/favorites',
    { params },
  );

// 5. 新增目前會員收藏
export const addSupermarketFavorite = (productCode: string) =>
  httpClient.post<ApiResponse<SupermarketProduct>>('/me/supermarket-offers/favorites', { productCode });

// 6. 移除目前會員收藏
export const removeSupermarketFavorite = (productCode: string) =>
  httpClient.delete<ApiResponse<{ productCode: string; isFavorite: boolean }>>(
    `/me/supermarket-offers/favorites/${encodeURIComponent(productCode)}`,
  );

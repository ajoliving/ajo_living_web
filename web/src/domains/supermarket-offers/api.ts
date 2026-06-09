/*
 * 超市優惠 API。
 * 1. 串接 AJO 後端公開 good-price 代理介面。
 * 2. 串接 AJO 會員收藏與價格提示介面。
 * 3. 保持前端不直接使用 good-price 舊帳戶 API。
 */
import httpClient from '@/shared/utils/http';
import type { ApiResponse } from '@/shared/utils/http/model';
import type {
  SupermarketPriceAlert,
  SupermarketProduct,
  SupermarketProductDetail,
  SupermarketSearchParams,
  SupermarketSearchResult,
  SupermarketSummary,
} from '@/domains/supermarket-offers/model';

// 1. 取得超市優惠首頁摘要
export const fetchSupermarketSummary = () =>
  httpClient.get<ApiResponse<SupermarketSummary>>('/supermarket-offers/summary');

// 2. 搜尋超市商品
export const searchSupermarketProducts = (params: SupermarketSearchParams) =>
  httpClient.get<ApiResponse<SupermarketSearchResult>>('/supermarket-offers/search', { params });

// 3. 取得超市商品詳情
export const fetchSupermarketProductDetail = (code: string, days = 90) =>
  httpClient.get<ApiResponse<SupermarketProductDetail>>(`/supermarket-offers/products/${encodeURIComponent(code)}`, {
    params: { days },
  });

// 4. 取得目前會員收藏
export const fetchSupermarketFavorites = (params: { page?: number; pageSize?: number } = {}) =>
  httpClient.get<ApiResponse<{ items: SupermarketProduct[]; pagination: { page: number; page_size: number; total: number } }>>(
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

// 7. 取得目前會員價格提示
export const fetchSupermarketPriceAlerts = () =>
  httpClient.get<ApiResponse<{ items: SupermarketPriceAlert[] }>>('/me/supermarket-offers/price-alerts');

// 8. 建立或更新目前會員價格提示
export const saveSupermarketPriceAlert = (payload: {
  productCode: string;
  targetPrice?: number;
  priceMode: 'list' | 'effective';
  offerRequired: boolean;
  enabled: boolean;
}) => httpClient.post<ApiResponse<SupermarketPriceAlert>>('/me/supermarket-offers/price-alerts', payload);

// 9. 更新目前會員價格提示
export const updateSupermarketPriceAlert = (
  id: number,
  payload: {
    targetPrice?: number;
    priceMode: 'list' | 'effective';
    offerRequired: boolean;
    enabled: boolean;
  },
) => httpClient.patch<ApiResponse<SupermarketPriceAlert>>(`/me/supermarket-offers/price-alerts/${id}`, payload);

// 10. 刪除目前會員價格提示
export const deleteSupermarketPriceAlert = (id: number) =>
  httpClient.delete<ApiResponse<{ id: number; deleted: boolean }>>(`/me/supermarket-offers/price-alerts/${id}`);

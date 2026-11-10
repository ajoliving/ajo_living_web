/*
 * API 共用型別。
 * 1. 定義統一回應格式。
 * 2. 定義常用列表包裝資料結構。
 * 3. 為後續真實 API 串接提供型別骨架。
 */

// 1. 定義標準 API 回應包裝
export interface ApiResponse<T> {
  code: string;
  message: string;
  data: T;
  request_id?: string;
  timestamp?: string;
}

// 2. 定義列表包裝資料
export interface ApiListData<T> {
  items: T[];
}

// 3. 定義分頁資訊
export interface PaginationMeta {
  page: number;
  page_size: number;
  total: number;
}

// 4. 定義分頁回應結構
export interface PaginatedResult<T> {
  items: T[];
  pagination: PaginationMeta;
}

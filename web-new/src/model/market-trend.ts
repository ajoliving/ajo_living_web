/*
 * 市場走勢資料型別。
 * 1. 定義公開租金走勢 API 回應。
 * 2. 保持前端圖表資料與後端欄位一致。
 */

// 1. 月度租金資料點
export interface MarketRentTrendPoint {
  month: string;
  label: string;
  value_hkd_per_sqft: number;
  source_value_hkd_per_sqm: number;
}

// 2. 區域租金走勢
export interface MarketRentTrendRegion {
  key: string;
  label: string;
  latest_hkd_per_sqft: number;
  monthly_change_percent: number;
  direction: 'up' | 'down' | 'flat';
  points: MarketRentTrendPoint[];
}

// 3. 公開租金走勢回應
export interface MarketRentTrendResponse {
  source: string;
  source_url: string;
  dataset: string;
  unit: string;
  display_unit: string;
  updated_month: string;
  method: string;
  months: string[];
  regions: MarketRentTrendRegion[];
}

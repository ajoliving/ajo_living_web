/*
 * 首頁摘要型別。
 * 1. 對齊 `/channel-home/overview` 回傳結構。
 * 2. 提供首頁頻道入口與精選二手資料型別。
 * 3. 定義首頁社區環境、交通及生活提示的資料範圍。
 */

// 1. 定義首頁頻道入口
export interface HomeChannelEntry {
  code: string;
  title: string;
  description: string;
}

// 2. 定義首頁精選二手摘要
export interface HomeFeaturedSecondhand {
  listing_id: string;
  title: string;
  summary: string;
  district_code: string;
  category_code: string;
  price_mode: string;
  price_hkd: number;
  condition_level: string;
  visibility_scope: string;
  contact_method: string;
  is_free_giveaway: boolean;
  publisher_identity_type: string;
  publication_status: string;
  business_status: string;
  updated_at: string;
}

// 3. 定義首頁摘要資料
export interface ChannelHomeOverview {
  channels: HomeChannelEntry[];
  featured_secondhand: HomeFeaturedSecondhand[];
}

// 4. 各項讀數獨立回傳；缺失資料不以零值替代。
export interface HomeEnvironment {
  scope: 'hong_kong' | 'building' | 'reference';
  building_name?: string;
  temperature?: { value: number; station: string; observed_at: string };
  forecast?: { date: string; min: number; max: number; description: string };
  forecast_days?: Array<{ date: string; min: number; max: number; description: string }>;
  humidity?: number;
  uv?: number;
  rainfall?: { value: number; district: string; observed_at: string };
  air_quality?: { value: string; station: string; observed_at: string };
  transport?: HomeTransport;
  holiday?: { name: string; date: string; is_today: boolean };
  parking?: { name: string; spaces?: number; observed_at?: string; closed?: boolean };
  emergency?: { hospital: string; wait: string; observed_at?: string; category: '3' | '45'; nearby: boolean };
}

// 5. 到站時間使用完整時間戳，不以固定倒數取代。
export interface HomeTransport {
  station: string;
  line: string;
  nearby: boolean;
  observed_at?: string;
  arrivals: Array<{ destination: string; arrival_at: string }>;
  delayed?: boolean;
}

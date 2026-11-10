/*
 * 首頁摘要型別。
 * 1. 對齊 `/channel-home/overview` 回傳結構。
 * 2. 提供首頁頻道入口與精選二手資料型別。
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

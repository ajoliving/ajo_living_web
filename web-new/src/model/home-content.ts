/*
 * 首頁內容設定型別。
 * 1. 對齊首頁三個主模組與登入背景圖設定 API。
 * 2. 提供首頁、登入頁與設定頁共用資料結構。
 */

// 1. 定義首頁主模組代碼
export type HomeContentModuleCode = 'secondhand' | 'property_sale' | 'serviced_apartment';

// 2. 定義首頁三大圖設定
export interface HomeModuleCard {
  module_code: HomeContentModuleCode;
  media_asset_id: string;
  url: string;
  object_key: string;
  title: string;
  subtitle: string;
  body: string;
}

// 3. 定義登入頁背景圖
export interface LoginHeroImageSetting {
  media_asset_id: string;
  url: string;
  object_key: string;
  author: string;
  location: string;
  sort_order: number;
}

// 4. 定義首頁公開內容
export interface HomeContentResponse {
  module_cards: HomeModuleCard[];
}

// 5. 定義首頁三大圖保存項
export interface HomeModuleCardSaveItem {
  module_code: HomeContentModuleCode;
  media_asset_id: string;
  title: string;
  subtitle: string;
  body: string;
}

// 6. 定義登入頁背景圖保存項
export interface LoginHeroSaveItem {
  media_asset_id: string;
  author: string;
  location: string;
  sort_order: number;
}

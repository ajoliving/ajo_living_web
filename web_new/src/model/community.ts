/*
 * 社區資料型別。
 * 1. 對應原型社區與真實社區中繼資料。
 * 2. 供列表、詳情、會員資料與聊天摘要共用。
 * 3. 補充 POS relay 的大廈與單位選擇結構。
 */

// 1. 定義地區分區
export type CommunityRegion =
  | 'hong-kong-island'
  | 'kowloon'
  | 'new-territories'
  | 'outlying-islands';

// 2. 定義社區資料結構
export interface Community {
  id: number;
  slug: string;
  name: string;
  district: string;
  region: CommunityRegion;
  building_count: number;
}

// 3. 定義真實 API 社區資料
export interface MetaCommunity {
  public_id: string;
  community_type: string;
  name_zh: string;
  name_en: string;
  district_code: string;
  address_text: string;
}

// 4. POS 樓宇摘要
export interface PosBuilding {
  building_id?: string;
  buildname_chi?: string;
  buildname?: string;
  name?: string;
  id?: string;
}

// 5. POS 單位摘要
export interface PosBuildingUnit {
  unit_id?: string;
  id?: string;
  floor?: string;
  unit?: string;
  unit_name?: string;
  name?: string;
}

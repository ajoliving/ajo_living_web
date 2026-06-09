/*
 * POS 樓宇資料型別。
 * 1. 對齊 POS relay 的樓宇清單與單位清單回應。
 * 2. 供住戶註冊頁級聯選擇大廈、樓層與單位使用。
 */

// 1. POS 樓宇摘要
export interface PosBuilding {
  building_id?: string;
  buildname_chi?: string;
  buildname?: string;
  name?: string;
  id?: string;
}

// 2. POS 單位摘要
export interface PosBuildingUnit {
  unit_id?: string;
  id?: string;
  floor?: string;
  unit?: string;
  unit_name?: string;
  name?: string;
}

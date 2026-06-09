/*
 * POS 樓宇 API。
 * 1. 讀取 POS relay 樓宇清單。
 * 2. 依大廈讀取可選單位清單。
 */
import httpClient from '@/shared/utils/http';
import type { ApiListData, ApiResponse } from '@/shared/utils/http/model';
import type { PosBuilding, PosBuildingUnit } from '@/domains/building/model';

// 1. 取得 POS 樓宇清單
export const fetchPosBuildings = async (): Promise<PosBuilding[]> => {
  const { data } = await httpClient.get<ApiResponse<ApiListData<PosBuilding>>>('/pos/buildings');
  return data.data.items;
};

// 2. 取得 POS 大廈單位清單
export const fetchPosBuildingUnits = async (buildingID: string): Promise<PosBuildingUnit[]> => {
  const { data } = await httpClient.get<ApiResponse<ApiListData<PosBuildingUnit>>>(`/pos/buildings/${encodeURIComponent(buildingID)}/units`);
  return data.data.items;
};

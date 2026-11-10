/*
 * 會員中心 - 住戶單位顯示資料。
 * 1. 以 POS 單位詳情對齊權限碼中的真實樓層與單位名稱。
 * 2. 單位詳情缺失時保留權限碼拆分結果作為回退。
 */
import type { PosBuildingUnit } from '@/model/community';

export interface ResidentUnitDisplay {
  unitID: string;
  buildingID: string;
  floor: string;
  unit: string;
}

// 1. 正規化 POS 識別碼
const digitsOnly = (value: unknown): string =>
  String(value ?? '').replace(/\D/g, '');

// 2. 取得 POS 單位詳情識別碼
const getUnitID = (item: PosBuildingUnit): string =>
  digitsOnly(item.unit_id ?? item.id).slice(0, 11);

// 3. 取得 POS 單位顯示名稱
const getUnitName = (item: PosBuildingUnit): string =>
  String(item.unit ?? item.unit_name ?? item.name ?? '').trim();

// 4. 以單位詳情建立住戶單位顯示資料
export const resolveResidentUnitDisplays = (
  permissionUnitIDs: string[],
  unitDetails: PosBuildingUnit[],
): ResidentUnitDisplay[] => {
  const unitDetailMap = new Map(
    unitDetails
      .map((item) => [getUnitID(item), item] as const)
      .filter(([unitID]) => unitID.length === 11),
  );

  return permissionUnitIDs.map((rawUnitID) => {
    const unitID = digitsOnly(rawUnitID).slice(0, 11);
    const buildingID = unitID.slice(0, 7);
    const detail = unitDetailMap.get(unitID);
    if (detail) {
      return {
        unitID,
        buildingID,
        floor: String(detail.floor ?? '').trim(),
        unit: getUnitName(detail),
      };
    }

    const floorCode = unitID.slice(7, 9);
    const unitCode = unitID.slice(9, 11);
    return {
      unitID,
      buildingID,
      floor: floorCode === '00' ? '' : floorCode,
      unit: unitCode.replace(/^0+/, '') || unitCode,
    };
  });
};

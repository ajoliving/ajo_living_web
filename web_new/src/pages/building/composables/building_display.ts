/*
 * 我的大廈顯示資料解析。
 * 1. 解析會員目前選中的所屬大廈。
 * 2. 建立 POS 大廈名稱索引並排除 ID 佔位名稱。
 */
import type { PosBuilding } from '@/model/community';
import type { CurrentMemberCommunity, CurrentMemberProfile } from '@/model/user';

// 1. 建立可兼容補零格式的大廈 ID 查找鍵
export const buildingIDKeys = (value: string): string[] => {
  const rawValue = value.trim();
  const result = rawValue ? [rawValue] : [];
  if (!/^\d+$/.test(rawValue)) {
    return result;
  }

  [rawValue.padStart(7, '0'), rawValue.replace(/^0+/, '') || '0'].forEach((item) => {
    if (!result.includes(item)) {
      result.push(item);
    }
  });
  return result;
};

// 2. 判斷名稱是否只是大廈 ID 佔位值
export const usableBuildingName = (value: unknown, buildingID: string): string => {
  const name = String(value ?? '').trim();
  if (!name) {
    return '';
  }
  return buildingIDKeys(buildingID).includes(name) ? '' : name;
};

// 3. 讀取 POS 大廈 ID
export const posBuildingID = (item: PosBuilding): string =>
  String(item.building_id ?? item.id ?? '').trim();

// 4. 讀取 POS 大廈正式名稱
export const posBuildingName = (item: PosBuilding, locale: string): string => {
  const buildingID = posBuildingID(item);
  const candidates = locale === 'en'
    ? [item.buildname, item.name, item.buildname_chi]
    : [item.buildname_chi, item.buildname, item.name];
  for (const candidate of candidates) {
    const name = usableBuildingName(candidate, buildingID);
    if (name) {
      return name;
    }
  }
  return '';
};

// 5. 讀取會員目前保存的大廈 ID
export const preferredMemberBuildingID = (member: CurrentMemberProfile | null): string => {
  const boundBuildingID = member?.bound_building_ids?.find((item) => String(item ?? '').trim())?.trim() ?? '';
  return boundBuildingID || member?.primary_community?.public_id?.trim() || '';
};

// 6. 讀取會員社區正式名稱
export const memberCommunityName = (community: CurrentMemberCommunity | undefined, locale: string): string => {
  const buildingID = community?.public_id?.trim() ?? '';
  const candidates = locale === 'en'
    ? [community?.name_en, community?.name_zh, community?.address_text]
    : [community?.name_zh, community?.name_en, community?.address_text];
  for (const candidate of candidates) {
    const name = usableBuildingName(candidate, buildingID);
    if (name) {
      return name;
    }
  }
  return '';
};

// 7. 建立大廈名稱索引
export const buildBuildingNameMap = (
  buildings: PosBuilding[],
  community: CurrentMemberCommunity | undefined,
  locale: string,
): Map<string, string> => {
  const result = new Map<string, string>();
  buildings.forEach((item) => {
    const buildingID = posBuildingID(item);
    const name = posBuildingName(item, locale);
    if (!buildingID || !name) {
      return;
    }
    buildingIDKeys(buildingID).forEach((key) => result.set(key, name));
  });

  const communityID = community?.public_id?.trim() ?? '';
  const communityName = memberCommunityName(community, locale);
  if (communityID && communityName) {
    buildingIDKeys(communityID).forEach((key) => {
      if (!result.has(key)) {
        result.set(key, communityName);
      }
    });
  }
  return result;
};

// 8. 按大廈 ID 讀取正式名稱
export const indexedBuildingName = (names: Map<string, string>, buildingID: string): string => {
  const key = buildingIDKeys(buildingID).find((item) => names.has(item));
  return key ? names.get(key) ?? '' : '';
};

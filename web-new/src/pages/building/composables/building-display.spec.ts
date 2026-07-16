/*
 * 我的大廈顯示資料測試。
 * 1. 驗證已保存綁定優先於舊主要社區。
 * 2. 驗證正式大廈名稱優先且不顯示 ID 佔位值。
 */
import { describe, expect, it } from 'vitest';

import type { CurrentMemberProfile } from '@/model/user';

import {
  buildBuildingNameMap,
  indexedBuildingName,
  preferredMemberBuildingID,
  usableBuildingName,
} from './building-display';

describe('building display data', () => {
  it('prefers the saved building binding over an older primary community', () => {
    const member = {
      bound_building_ids: ['0999900'],
      primary_community: { public_id: '0419900' },
    } as CurrentMemberProfile;

    expect(preferredMemberBuildingID(member)).toBe('0999900');
  });

  it('uses the POS building name and ignores ID placeholder names', () => {
    const names = buildBuildingNameMap([
      { building_id: '0999900', buildname_chi: '測試1大廈', buildname: 'Test Building 1' },
    ], {
      public_id: '0999900',
      community_type: 'building',
      name_zh: '0999900',
      name_en: '0999900',
      district_code: 'unknown',
      address_text: '0999900',
    }, 'zh-HK');

    expect(indexedBuildingName(names, '999900')).toBe('測試1大廈');
    expect(usableBuildingName('0999900', '0999900')).toBe('');
  });
});

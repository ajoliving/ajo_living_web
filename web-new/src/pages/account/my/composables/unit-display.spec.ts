/*
 * 會員中心住戶單位顯示測試。
 * 1. 驗證權限碼會採用 POS 單位詳情中的真實樓層與單位名稱。
 * 2. 驗證 POS 單位詳情缺失時仍可使用權限碼回退。
 */
import { describe, expect, it } from 'vitest';

import { resolveResidentUnitDisplays } from './unit-display';

describe('resolveResidentUnitDisplays', () => {
  it('uses POS unit details for the resident unit labels', () => {
    expect(resolveResidentUnitDisplays(
      ['09999000012', '09999000111', '09999000211'],
      [
        { unit_id: '09999000012', floor: 'G', unit: 'B' },
        { unit_id: '09999000111', floor: '01', unit: 'A' },
        { unit_id: '09999000211', floor: '02', unit: 'A' },
      ],
    )).toEqual([
      { unitID: '09999000012', buildingID: '0999900', floor: 'G', unit: 'B' },
      { unitID: '09999000111', buildingID: '0999900', floor: '01', unit: 'A' },
      { unitID: '09999000211', buildingID: '0999900', floor: '02', unit: 'A' },
    ]);
  });

  it('falls back to permission code segments when unit details are unavailable', () => {
    expect(resolveResidentUnitDisplays(['09999000012'], [])).toEqual([
      { unitID: '09999000012', buildingID: '0999900', floor: '', unit: '12' },
    ]);
  });
});

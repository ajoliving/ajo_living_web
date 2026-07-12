/*
 * 物業頻道常量測試。
 * 1. 驗證服務式住宅設施代碼顯示為本地化標籤。
 * 2. 保留英文介面的對應標籤。
 */
import { describe, expect, it } from 'vitest';

import { getPropertyTagLabel } from '@/constants/property';

// 1. 驗證服務式住宅設施與服務標籤
describe('property serviced residence tags', () => {
  it.each([
    ['front_desk_24h', '24小時接待'],
    ['broadband', '寬頻上網'],
    ['wifi', '無線網絡'],
    ['private_kitchen', '獨立廚房'],
    ['business_center', '商業中心'],
    ['child_care', '幼兒護理'],
    ['pay_tv', '收費電視'],
    ['pet_friendly', '可養貓狗'],
    ['shuttle_bus', '接駁巴士'],
    ['pool', '泳池'],
  ])('resolves %s to a Traditional Chinese label', (value, label) => {
    expect(getPropertyTagLabel(value, 'zh-HK')).toBe(label);
  });

  it('keeps the English Wi-Fi label for the English locale', () => {
    expect(getPropertyTagLabel('wifi', 'en')).toBe('Wi-Fi');
  });
});

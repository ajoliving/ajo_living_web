/*
 * 物業展示工具測試。
 * 1. 驗證公開卡片只顯示一次位置資料。
 * 2. 驗證卡片資料列顯示有效實際樓層與公開樓層並移除座向。
 * 3. 驗證住宅分類優先顯示於特色標籤。
 * 4. 驗證繁中卡片按可靠資料並列屋苑 English 名稱。
 */
import { describe, expect, it } from 'vitest';

import type { PropertyListingSummaryResponse } from '@/model/property';
import {
  formatPropertyUnit,
  resolvePropertyCardCommunityName,
  resolvePropertyListingFacts,
  resolvePropertyTagLabels,
} from './property';

const listing = {
  listing_id: 'property-preview',
  display_number: 1,
  module: 'property_sale',
  title: '測試樓盤',
  summary: '',
  district_code: 'sham_shui_po',
  publisher_identity_type: 'owner',
  publication_status: 'draft',
  business_status: 'available',
  updated_at: '2026-07-30T00:00:00Z',
  community: {
    public_id: 'community-1',
    community_type: 'estate',
    name_zh: '美寧中心',
    name_en: 'Merlin Centre',
    district_code: 'sham_shui_po',
    address_text: '長沙灣順寧道',
  },
  property_sale: {
    property_type: 'residential',
    estate_name: '美寧中心',
    block_name: 'Block D',
    floor_level: '25/F',
    floor_raw: '25',
    floor_zone: 'middle',
    unit_name: '3',
    show_unit: true,
    bedroom_count: 1,
    bathroom_count: 1,
    direction: 'N/A',
  },
} as PropertyListingSummaryResponse;

describe('resolvePropertyListingFacts', () => {
  // 1. 使用座數、實際／公開樓層及單位建立公開資訊列
  it('keeps the requested property location facts', () => {
    expect(resolvePropertyListingFacts(listing, 'zh-HK')).toEqual([
      'Block D',
      '25/F / 中層',
      '3單位',
    ]);
    expect(resolvePropertyListingFacts(listing, 'en')).toEqual([
      'Block D',
      '25/F / Middle floor',
      'Unit 3',
    ]);
  });

  // 2. 單獨的樓層單位字母不當作有效實際樓層
  it('ignores a standalone floor suffix', () => {
    const listingWithInvalidRawFloor = {
      ...listing,
      property_sale: { ...listing.property_sale, floor_raw: 'F' },
    } as PropertyListingSummaryResponse;

    expect(resolvePropertyListingFacts(listingWithInvalidRawFloor, 'zh-HK')).toEqual([
      'Block D',
      '中層',
      '3單位',
    ]);
  });

  // 2. 繁中卡片並列匹配屋苑的 English 名稱
  it('combines matching Chinese and English estate names', () => {
    expect(resolvePropertyCardCommunityName(listing, 'zh-HK')).toBe('美寧中心 Merlin Centre');
    expect(resolvePropertyCardCommunityName(listing, 'en')).toBe('Merlin Centre');

    const differentEstate = {
      ...listing,
      property_sale: { ...listing.property_sale, estate_name: '其他屋苑' },
    } as PropertyListingSummaryResponse;
    expect(resolvePropertyCardCommunityName(differentEstate, 'zh-HK')).toBe('其他屋苑');
  });

  // 3. 單位名稱按語系補足可讀標籤
  it('formats unit names for Chinese and English cards', () => {
    expect(formatPropertyUnit('3', 'zh-HK')).toBe('3單位');
    expect(formatPropertyUnit('09 室', 'zh-HK')).toBe('09 室');
    expect(formatPropertyUnit('3', 'en')).toBe('Unit 3');
    expect(formatPropertyUnit('Flat 3', 'en')).toBe('Flat 3');
  });
});

describe('resolvePropertyTagLabels', () => {
  // 1. 住宅分類即使在細則標籤之後儲存，仍優先顯示於卡片
  it('prioritizes residential category labels before detail labels', () => {
    const listingWithTags = {
      ...listing,
      property_sale: {
        ...listing.property_sale,
        feature_tags: ['view_mountain', 'fitout_luxury', 'residential_private_estate'],
      },
    } as PropertyListingSummaryResponse;

    expect(resolvePropertyTagLabels(listingWithTags, 'zh-HK', 4)).toEqual([
      '私人屋苑',
      '望山景',
      '豪華裝修',
    ]);
  });
});

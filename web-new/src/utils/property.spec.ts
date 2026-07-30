/*
 * 物業展示工具測試。
 * 1. 驗證公開卡片只顯示一次位置資料。
 * 2. 驗證未指定座向不會顯示 N/A。
 * 3. 驗證住宅分類優先顯示於特色標籤。
 */
import { describe, expect, it } from 'vitest';

import type { PropertyListingSummaryResponse } from '@/model/property';
import { resolvePropertyListingFacts, resolvePropertyTagLabels } from './property';

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
  property_sale: {
    property_type: 'residential',
    block_name: 'Block D',
    floor_level: '中層',
    unit_name: '3',
    show_unit: true,
    bedroom_count: 1,
    bathroom_count: 1,
    direction: 'N/A',
  },
} as PropertyListingSummaryResponse;

describe('resolvePropertyListingFacts', () => {
  // 1. 使用公開位置、房間及浴室資料建立單一中間資訊列
  it('does not duplicate the public location or show N/A direction', () => {
    expect(resolvePropertyListingFacts(listing, 'zh-HK')).toEqual([
      'Block D',
      '中層',
      '3',
      '1房 · 1浴室',
    ]);
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

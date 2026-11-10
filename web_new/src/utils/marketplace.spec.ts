/*
 * 二手交易展示工具測試。
 * 1. 驗證社區名稱及地區後備文案跟隨目前語系。
 */
import { describe, expect, it } from 'vitest';

import type { SecondhandListingSummaryResponse } from '@/model/marketplace';

import { resolveListingCommunityName } from './marketplace';

// 1. 建立最小可用帖子資料
const createListing = (): SecondhandListingSummaryResponse => ({
  listing_id: 'listing-1',
  title: 'Dining table',
  summary: '',
  district_code: 'wan_chai',
  category_code: 'home_furniture',
  price_mode: 'fixed',
  price_hkd: 100,
  condition_level: 'used_good',
  visibility_scope: 'public',
  contact_method: 'chat',
  is_free_giveaway: false,
  publisher_identity_type: 'member',
  publication_status: 'active',
  business_status: 'available',
  updated_at: '2026-07-12T00:00:00Z',
  is_favorited: false,
  community: {
    public_id: 'community-1',
    community_type: 'estate',
    name_zh: '灣仔花園',
    name_en: 'Wan Chai Garden',
    district_code: 'wan_chai',
    address_text: '灣仔',
  },
});

describe('marketplace display helpers', () => {
  it('selects the community name for the requested locale', () => {
    const listing = createListing();

    expect(resolveListingCommunityName(listing, 'zh-HK')).toBe('灣仔花園');
    expect(resolveListingCommunityName(listing, 'en')).toBe('Wan Chai Garden');
  });

  it('localizes the district fallback when community data is absent', () => {
    const listing = createListing();
    listing.community = null;

    expect(resolveListingCommunityName(listing, 'zh-HK')).toBe('灣仔區');
    expect(resolveListingCommunityName(listing, 'en')).toBe('Wan Chai');
  });
});

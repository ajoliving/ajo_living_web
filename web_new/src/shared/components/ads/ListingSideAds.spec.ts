/*
 * 列表右側展示廣告欄測試。
 * 1. 驗證公開廣告 API 參數。
 * 2. 驗證已配置廣告與空廣告位渲染。
 */
import { flushPromises, mount } from '@vue/test-utils';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { fetchPublicDisplayAds } from '@/httpapis/wallet';

import ListingSideAds from './ListingSideAds.vue';

vi.mock('@/httpapis/wallet', () => ({
  fetchPublicDisplayAds: vi.fn(),
}));

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string): string => ({
      'common.ad.label': '廣告',
      'common.ad.slot': '廣告位',
    })[key] ?? key,
  }),
}));

const mockedFetchPublicDisplayAds = vi.mocked(fetchPublicDisplayAds);

// 1. 建立公開廣告 API 回應
const createPublicAdsResponse = (items: unknown[]) => ({
  data: {
    data: {
      items,
    },
  },
}) as unknown as Awaited<ReturnType<typeof fetchPublicDisplayAds>>;

describe('ListingSideAds', () => {
  beforeEach(() => {
    mockedFetchPublicDisplayAds.mockReset();
  });

  it('loads and renders listing side ads for the given channel', async () => {
    mockedFetchPublicDisplayAds.mockResolvedValue(createPublicAdsResponse([
      {
        task_id: 'ad-1',
        title: '搬屋優惠',
        display_title: '全港搬屋優惠',
        summary: '一站式服務',
        display_text: '即日報價',
        cover_url: '',
        media_url: 'https://cdn.example.com/ad.png',
        media_type: 'image',
        target_url: 'https://example.com',
        display_channel: 'property_sale',
        display_placement: 'listing_side',
        display_layout: 'image_full',
        sort_order: 1,
        slot_index: 1,
      },
    ]));

    const wrapper = mount(ListingSideAds, {
      props: {
        channel: 'property_sale',
      },
    });
    await flushPromises();

    expect(mockedFetchPublicDisplayAds).toHaveBeenCalledWith({
      channel: 'property_sale',
      placement: 'listing_side',
      limit: 5,
    });
    expect(wrapper.text()).toContain('全港搬屋優惠');
    expect(wrapper.text()).toContain('即日報價');
    expect(wrapper.find('a[href="https://example.com"]').exists()).toBe(true);
  });

  it('renders fixed placeholders when no ads are configured', async () => {
    mockedFetchPublicDisplayAds.mockResolvedValue(createPublicAdsResponse([]));

    const wrapper = mount(ListingSideAds, {
      props: {
        channel: 'furniture',
      },
    });
    await flushPromises();

    expect(wrapper.findAll('.listing-side-ads__slot')).toHaveLength(5);
    expect(wrapper.findAll('.listing-side-ads__slot--short')).toHaveLength(3);
    expect(wrapper.findAll('.listing-side-ads__slot--long')).toHaveLength(2);
    expect(wrapper.text()).toContain('16:9');
    expect(wrapper.text()).toContain('9:16');
  });
});

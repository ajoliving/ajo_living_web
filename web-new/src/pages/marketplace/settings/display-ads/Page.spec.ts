/*
 * 展示廣告位設定頁測試。
 * 1. 驗證頻道設定載入。
 * 2. 驗證保存廣告位 payload。
 */
import { flushPromises, mount } from '@vue/test-utils';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { fetchDisplayAdSettings, fetchStaffRewardAds, saveDisplayAdSettings } from '@/httpapis/wallet';

import Page from './Page.vue';

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string, params?: Record<string, unknown>) => {
      if (key === 'marketplace.settings.displayAdSlotLabel') {
        return `廣告位 ${String(params?.index ?? '')}`;
      }

      const messages: Record<string, string> = {
        'common.status.loading': '載入中',
        'marketplace.settings.displayAdChannelPropertySale': '樓盤放售',
        'marketplace.settings.displayAdChannelFurniture': '二手家私',
        'marketplace.settings.displayAdChannelServicedApartment': '服務式住宅',
        'marketplace.settings.emptySlot': '未設定',
        'marketplace.settings.emptySlotHint': '選擇廣告後顯示',
        'marketplace.settings.saving': '儲存中',
      };

      return messages[key] ?? key;
    },
  }),
}));

vi.mock('@/stores/feedback', () => ({
  useFeedbackStore: () => ({
    pushToast: vi.fn(),
  }),
}));

vi.mock('@/httpapis/wallet', () => ({
  fetchDisplayAdSettings: vi.fn(),
  fetchStaffRewardAds: vi.fn(),
  saveDisplayAdSettings: vi.fn(),
}));

const mockedFetchDisplayAdSettings = vi.mocked(fetchDisplayAdSettings);
const mockedFetchStaffRewardAds = vi.mocked(fetchStaffRewardAds);
const mockedSaveDisplayAdSettings = vi.mocked(saveDisplayAdSettings);

// 1. 建立固定廣告位設定回應
const createSettingsResponse = () => ({
  data: {
    data: {
      channel: 'property_sale',
      slots: Array.from({ length: 10 }, (_, index) => ({
        slot_index: index + 1,
        layout: index < 3 ? 'text_compact' : 'image_text',
        ads: [],
      })),
    },
  },
}) as unknown as Awaited<ReturnType<typeof fetchDisplayAdSettings>>;

// 2. 建立廣告列表回應
const createAdsResponse = () => ({
  data: {
    data: {
      items: [
        {
          task_id: 'ad-1',
          ad_type: 'display',
          title: '搬屋優惠',
          summary: '即日報價',
          cover_url: '',
          media_url: 'https://cdn.example.com/ad.png',
          media_type: 'image',
          target_url: 'https://example.com',
          display_channel: 'property_sale',
          display_placement: 'listing_side',
          display_layout: 'image_text',
          slot_display_title: '',
          display_text: '',
          slot_target_url: '',
          sort_order: 1,
          reward_points: 0,
          watch_seconds: 0,
          total_budget: 0,
          total_granted: 0,
          remaining_budget: 0,
          watch_count: 0,
          total_watch_seconds: 0,
          link_click_count: 0,
          link_click_rate: 0,
          is_active: true,
          created_at: '',
          updated_at: '',
        },
      ],
      pagination: {
        page: 1,
        page_size: 50,
        total: 1,
      },
    },
  },
}) as unknown as Awaited<ReturnType<typeof fetchStaffRewardAds>>;

describe('DisplayAdsSettingsPage', () => {
  beforeEach(() => {
    mockedFetchDisplayAdSettings.mockReset();
    mockedFetchStaffRewardAds.mockReset();
    mockedSaveDisplayAdSettings.mockReset();
    mockedFetchDisplayAdSettings.mockResolvedValue(createSettingsResponse());
    mockedFetchStaffRewardAds.mockResolvedValue(createAdsResponse());
    mockedSaveDisplayAdSettings.mockResolvedValue(createSettingsResponse());
  });

  it('saves selected display ad into the first listing side slot', async () => {
    const wrapper = mount(Page);
    await flushPromises();

    const firstSlotSelect = wrapper.findAll('select')[0];
    await firstSlotSelect.setValue('ad-1');

    await wrapper.find('.display-ads-button--primary').trigger('click');
    await flushPromises();

    expect(mockedFetchStaffRewardAds).toHaveBeenCalledWith({
      page: 1,
      page_size: 50,
      ad_type: 'display',
      display_channel: 'property_sale',
      is_active: true,
    });
    expect(mockedSaveDisplayAdSettings).toHaveBeenCalledWith('property_sale', [
      {
        slot_index: 1,
        ad_task_ids: ['ad-1'],
        ads: [{
          ad_task_id: 'ad-1',
          display_title: '搬屋優惠',
          display_text: '即日報價',
          target_url: 'https://example.com',
        }],
      },
      {
        slot_index: 2,
        ad_task_ids: [],
        ads: [],
      },
    ]);
  });
});

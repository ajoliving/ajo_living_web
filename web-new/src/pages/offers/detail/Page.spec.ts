/*
 * 商品詳情關聯卡測試。
 * 1. 移除單一及多個其他門店價格的比較回退。
 * 2. 保留勝出門店的實際原價及完整門店價格表。
 */
import { flushPromises, mount } from '@vue/test-utils';
import { AxiosHeaders } from 'axios';
import { createPinia } from 'pinia';
import { afterEach, describe, expect, it, vi } from 'vitest';

import { fetchSupermarketProductDetail } from '@/httpapis/supermarket-offers';
import i18n from '@/i18n';
import type { SupermarketProduct, SupermarketStorePrice } from '@/model/supermarket-offers';
import Page from './Page.vue';

vi.mock('vue-router', () => ({
  useRoute: () => ({ params: { code: 'P001' }, query: {} }),
  useRouter: () => ({ push: vi.fn(), replace: vi.fn() }),
}));

vi.mock('@/httpapis/auth-session', () => ({
  readStoredAccessToken: () => '',
}));

vi.mock('@/httpapis/supermarket-offers', () => ({
  fetchSupermarketProductDetail: vi.fn(),
  addSupermarketFavorite: vi.fn(),
  removeSupermarketFavorite: vi.fn(),
  deleteSupermarketPriceAlert: vi.fn(),
  reportSupermarketProductImage: vi.fn(),
  saveSupermarketPriceAlert: vi.fn(),
}));

// 1. 建立真實原價與跨店價格可分辨的資料。
const makeProduct = (code: string, originalPrice: number, otherPrices: number[]): SupermarketProduct => {
  const prices: SupermarketStorePrice[] = [12, ...otherPrices].map((price, index) => ({
    store: ['AEON', 'PARKNSHOP', 'WELLCOME'][index]!,
    listPrice: index === 0 ? originalPrice : price,
    effectiveUnitPrice: price,
    offer: '',
    parseStatus: 'none',
    snapshotDate: '2026-10-20',
  }));
  return {
    code, name: '測試商品', brand: '測試品牌',
    category1: '個人護理', category2: '', category3: '',
    minPrice: 12, maxPrice: 60, priceDiff: 48, diffPercent: 80,
    listPrice: originalPrice, effectiveUnitPrice: 12, bestEffectiveUnitPrice: 12,
    discountAmount: originalPrice - 12, discountRate: 0,
    stores: prices.map((price) => price.store), storePrices: prices,
    hasOffer: originalPrice > 12, bestStore: 'AEON',
    offerType: 'none', offerPattern: 'none', parseStatus: 'none',
  };
};

// 2. 驗證繁中及 English 關聯商品不再呈現跨店比較行。
describe.each(['zh-HK', 'en'] as const)('Product detail related prices (%s)', (locale) => {
  afterEach(() => {
    i18n.global.locale.value = 'zh-HK';
  });

  it('hides other-store fallback while keeping actual original prices and the store table', async () => {
    i18n.global.locale.value = locale;
    const product = makeProduct('P001', 12, [26, 60]);
    vi.mocked(fetchSupermarketProductDetail).mockResolvedValue({
      status: 200,
      statusText: 'OK',
      headers: {},
      config: { headers: new AxiosHeaders() },
      data: {
        code: 'OK',
        message: 'success',
        data: {
          product,
          summary: { days: 90, lowestList: 12, highestList: 60, lowestDeal: 12, highestDeal: 60 },
          stores: product.storePrices!,
          history: [],
          sameBrand: [],
          sameCategory: [
            makeProduct('P002', 12, [26]),
            makeProduct('P003', 12, [26, 60]),
            makeProduct('P004', 20, [26, 60]),
          ],
          isFavorite: false,
        },
      },
    });

    const wrapper = mount(Page, {
      global: { plugins: [createPinia(), i18n], stubs: { AppIcon: true, Teleport: true } },
    });
    try {
      await flushPromises();
      const cards = wrapper.findAll('.gp-related-card');
      expect(cards).toHaveLength(3);
      expect(cards[0]!.find('.gp-related-compare').exists()).toBe(false);
      expect(cards[1]!.find('.gp-related-compare').exists()).toBe(false);
      expect(cards[2]!.find('.gp-related-compare').text()).toBe(
        locale === 'zh-HK' ? '原價 $20.00' : 'Original $20.00',
      );
      expect(wrapper.find('.gp-related-block').text()).not.toMatch(/他店|其他店鋪|Others|Other stores/);
      expect(wrapper.findAll('.gp-price-table tbody tr')).toHaveLength(3);
    } finally {
      wrapper.unmount();
    }
  });
});

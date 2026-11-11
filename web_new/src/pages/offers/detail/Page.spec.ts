/*
 * 商品詳情關聯卡測試。
 * 1. 移除單一及多個其他門店價格的比較回退。
 * 2. 保留勝出門店的實際原價及完整門店價格表。
 */
import { flushPromises, mount } from '@vue/test-utils';
import { AxiosHeaders } from 'axios';
import { createPinia } from 'pinia';
import { afterEach, describe, expect, it, vi } from 'vitest';

import { readStoredAccessToken } from '@/httpapis/auth_session';
import { addSupermarketFavorite, fetchSupermarketProductDetail, saveSupermarketPriceAlert } from '@/httpapis/supermarket_offers';
import i18n from '@/i18n';
import type { SupermarketProduct, SupermarketStorePrice } from '@/model/supermarket_offers';
import Page from './Page.vue';

vi.mock('vue-router', () => ({
  useRoute: () => ({ params: { code: 'P001' }, query: {} }),
  useRouter: () => ({ push: vi.fn(), replace: vi.fn() }),
}));

vi.mock('@/httpapis/auth_session', () => ({
  readStoredAccessToken: vi.fn(() => ''),
}));

vi.mock('@/httpapis/supermarket_offers', () => ({
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
    product.storePrices![0]!.offer = '每件慳$0';
    const zeroOfferProduct = makeProduct('P002', 12, [26]);
    zeroOfferProduct.storePrices![0]!.offer = '每件慳$0';
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
            zeroOfferProduct,
            makeProduct('P003', 12, [26, 60]),
            makeProduct('P004', 20, [26, 60]),
            ...Array.from({ length: 9 }, (_, index) => makeProduct(`P${index + 5}`, 20, [26])),
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
      expect(cards).toHaveLength(10);
      expect(cards[0]!.find('.gp-related-compare').exists()).toBe(false);
      expect(cards[0]!.find('.gp-related-saving').text()).toContain('$14');
      expect(wrapper.find('.gp-best-hero p').text()).toContain('$14');
      expect(cards[1]!.find('.gp-related-compare').exists()).toBe(false);
      expect(cards[2]!.find('.gp-related-compare').text()).toBe(
        locale === 'zh-HK' ? '原價 $20.00' : 'Original $20.00',
      );
      expect(wrapper.find('.gp-related-block').text()).not.toMatch(/他店|其他店鋪|Others|Other stores/);
      expect(wrapper.findAll('.gp-price-table tbody tr')).toHaveLength(3);
      expect(wrapper.findAll('.gp-related-header .gp-related-arrow')).toHaveLength(0);
      expect(wrapper.findAll('.gp-related-rail > .gp-related-arrow')).toHaveLength(2);
      const scrollBy = vi.fn();
      const track = wrapper.get('.gp-related-scroll');
      Object.defineProperty(track.element, 'clientWidth', { configurable: true, value: 400 });
      Object.defineProperty(track.element, 'scrollBy', { configurable: true, value: scrollBy });
      await wrapper.get('.gp-related-arrow.is-next').trigger('click');
      expect(scrollBy).toHaveBeenCalledWith({ left: 320, behavior: 'smooth' });
      await wrapper.get('.gp-related-arrow.is-previous').trigger('click');
      expect(scrollBy).toHaveBeenLastCalledWith({ left: -320, behavior: 'smooth' });
    } finally {
      wrapper.unmount();
    }
  });
});

// 3. 驗證提醒與收藏保持一致，零節省金額不顯示。
describe('Product detail actions', () => {
  afterEach(() => {
    vi.clearAllMocks();
    vi.unstubAllGlobals();
    i18n.global.locale.value = 'zh-HK';
  });

  it('hides zero savings and automatically favorites a product when saving an alert', async () => {
    vi.mocked(readStoredAccessToken).mockReturnValue('token');
    const product = makeProduct('P001', 12, []);
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
          summary: { days: 90, lowestList: 12, highestList: 12, lowestDeal: 12, highestDeal: 12 },
          stores: product.storePrices!,
          history: [],
          sameBrand: [],
          sameCategory: [makeProduct('P002', 12, [])],
          isFavorite: false,
        },
      },
    });
    vi.mocked(addSupermarketFavorite).mockResolvedValue({} as never);
    vi.mocked(saveSupermarketPriceAlert).mockResolvedValue({
      data: { data: { id: 1, productCode: 'P001', priceMode: 'effective', offerRequired: false, enabled: true } },
    } as never);
    vi.stubGlobal('navigator', {
      geolocation: {
        getCurrentPosition: (success: PositionCallback) => success({
          coords: { latitude: 22.3, longitude: 114.17 },
        } as GeolocationPosition),
      },
    });
    const open = vi.spyOn(window, 'open').mockImplementation(() => null);

    const wrapper = mount(Page, {
      global: { plugins: [createPinia(), i18n], stubs: { AppIcon: true, Teleport: true } },
    });
    try {
      await flushPromises();
      expect(wrapper.find('.gp-best-hero p').exists()).toBe(false);
      expect(wrapper.find('.gp-related-saving').exists()).toBe(false);
      const buttons = wrapper.findAll('.gp-detail-action');
      expect(buttons[0]!.classes()).not.toContain('gp-detail-action-primary');
      expect(buttons[2]!.text()).toBe('設定提醒');
      const branch = wrapper.find<HTMLSelectElement>('.gp-store-branch-select');
      const initialRoute = wrapper.find('.gp-store-route-button').attributes('href');
      const anotherBranch = branch.findAll('option')[2];
      expect(anotherBranch).toBeDefined();
      await branch.setValue(anotherBranch!.element.value);
      expect(open).not.toHaveBeenCalled();
      expect(wrapper.find('.gp-store-route-button').attributes('href')).not.toBe(initialRoute);
      await buttons[2]!.trigger('click');
      await wrapper.find('.gp-alert-form').trigger('submit');
      await flushPromises();
      expect(addSupermarketFavorite).toHaveBeenCalledWith('P001');
      expect(saveSupermarketPriceAlert).toHaveBeenCalledOnce();
      expect(buttons[1]!.text()).toBe(i18n.global.t('offers.list.favorited'));
    } finally {
      wrapper.unmount();
    }
  });
});

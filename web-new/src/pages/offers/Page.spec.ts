/*
 * 綜合優惠頁搜尋測試。
 * 1. 確認一般搜尋不會排除無優惠商品。
 * 2. 確認輸入文字會顯示商品搜尋建議。
 */
import { flushPromises, mount } from '@vue/test-utils';
import { createPinia, setActivePinia } from 'pinia';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { fetchSupermarketSummary, searchSupermarketProducts } from '@/httpapis/supermarket-offers';
import i18n from '@/i18n';

import Page from './Page.vue';

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: vi.fn() }),
}));

vi.mock('@/httpapis/auth-session', () => ({
  readStoredAccessToken: () => '',
}));

vi.mock('@/httpapis/supermarket-offers', () => ({
  fetchSupermarketSummary: vi.fn(),
  searchSupermarketProducts: vi.fn(),
  fetchSupermarketFavorites: vi.fn(),
  addSupermarketFavorite: vi.fn(),
  removeSupermarketFavorite: vi.fn(),
}));

const mockedFetchSummary = vi.mocked(fetchSupermarketSummary);
const mockedSearchProducts = vi.mocked(searchSupermarketProducts);

// 1. 建立最小可渲染的商品搜尋回應。
const createSearchResponse = (query = '') => ({
  data: {
    data: {
      items: query
        ? [{
            code: 'P000003682',
            name: '花生醬 - 幼滑裝 340克',
            brand: 'Meadows',
            category1: '',
            category2: '',
            category3: '',
            minPrice: 24.9,
            maxPrice: 24.9,
            priceDiff: 0,
            diffPercent: 0,
            listPrice: 24.9,
            effectiveUnitPrice: 24.9,
            bestEffectiveUnitPrice: 24.9,
            discountAmount: 0,
            discountRate: 0,
            stores: [],
            hasOffer: false,
            bestStore: '',
            offerType: 'none',
            offerPattern: 'none',
            parseStatus: 'none',
          }]
        : [],
      total: query ? 1 : 0,
      page: 1,
      pageSize: 20,
      stats: {
        records: 0,
        products: 0,
        stores: 0,
        offers: 0,
        parsedOffers: 0,
        unsupportedOffers: 0,
        ambiguousOffers: 0,
        uncalculatedOffers: 0,
      },
      categories: [],
      brands: [],
      stores: [],
    },
  },
}) as unknown as Awaited<ReturnType<typeof searchSupermarketProducts>>;

describe('SupermarketOffersPage search', () => {
  beforeEach(() => {
    vi.useFakeTimers();
    setActivePinia(createPinia());
    mockedFetchSummary.mockReset();
    mockedSearchProducts.mockReset();
    mockedFetchSummary.mockResolvedValue({
      data: {
        data: {
          stats: { records: 0, products: 0, stores: 0, offers: 0, parsedOffers: 0, unsupportedOffers: 0, ambiguousOffers: 0, uncalculatedOffers: 0 },
          categories: [],
          stores: [],
          cheapest: [],
          bestDiscounts: [],
          biggestDiffs: [],
          offers: [],
        },
      },
    } as unknown as Awaited<ReturnType<typeof fetchSupermarketSummary>>);
    mockedSearchProducts.mockImplementation((params) => Promise.resolve(createSearchResponse(params.q)));
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  // 2. 一般列表查詢必須包含沒有優惠的商品。
  it('requests all matching products instead of offer-only products', async () => {
    mount(Page, { global: { plugins: [createPinia(), i18n], stubs: { AppIcon: true } } });
    await flushPromises();

    expect(mockedSearchProducts).toHaveBeenCalledWith(expect.objectContaining({ offerOnly: false }));
  });

  // 3. 輸入文字後應顯示可直接選取的搜尋建議。
  it('shows suggestions for the typed product query', async () => {
    const wrapper = mount(Page, { global: { plugins: [createPinia(), i18n], stubs: { AppIcon: true } } });
    await flushPromises();

    await wrapper.find('.gp-sinput').setValue('花生醬');
    await vi.advanceTimersByTimeAsync(180);
    await flushPromises();

    expect(wrapper.find('.gp-search-suggestion').text()).toContain('花生醬 - 幼滑裝 340克');
    expect(mockedSearchProducts).toHaveBeenLastCalledWith(expect.objectContaining({
      q: '花生醬',
      offerOnly: false,
      pageSize: 8,
    }));

    await wrapper.find('.gp-search-suggestion').trigger('mousedown');
    await flushPromises();

    expect(wrapper.find<HTMLInputElement>('.gp-sinput').element.value).toBe('花生醬 - 幼滑裝 340克');
    expect(mockedSearchProducts).toHaveBeenLastCalledWith(expect.objectContaining({
      q: '花生醬 - 幼滑裝 340克',
      offerOnly: false,
      pageSize: 20,
    }));
  });
});

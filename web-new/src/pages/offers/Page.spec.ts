/*
 * 綜合優惠頁搜尋測試。
 * 1. 確認一般搜尋不會排除無優惠商品。
 * 2. 確認輸入文字會顯示品牌及包含規格的商品名稱。
 */
import { flushPromises, mount } from '@vue/test-utils';
import { createPinia, setActivePinia } from 'pinia';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { fetchSupermarketSummary, searchSupermarketProducts } from '@/httpapis/supermarket-offers';
import i18n from '@/i18n';

import Page from './Page.vue';

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: vi.fn() }),
  useRoute: () => ({ query: {} }),
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
const createSearchResponse = (query = '', itemCount = 1) => ({
  data: {
    data: {
      items: query
        ? Array.from({ length: itemCount }, (_, index) => ({
            code: `P00000368${index}`,
            name: index === 0 ? '花生醬 - 幼滑裝 340克' : `花生醬 - 幼滑裝 340克 ${index + 1}`,
            brand: 'Meadows',
            subtitle: '340克',
            category1: '個人護理',
            category2: '女士衛生用品',
            category3: '衛生巾',
            minPrice: 24.9,
            maxPrice: 24.9,
            priceDiff: 0,
            diffPercent: 0,
            listPrice: 24.9,
            effectiveUnitPrice: 24.9,
            bestEffectiveUnitPrice: 24.9,
            discountAmount: 0,
            discountRate: 0,
            stores: ['PARKNSHOP', 'WELLCOME'],
            storePrices: [
              {
                store: 'PARKNSHOP',
                listPrice: 64,
                effectiveUnitPrice: 39,
                offer: '買2件慳$50.00',
                parseStatus: 'parsed',
                snapshotDate: '2026-09-12',
              },
              {
                store: 'WELLCOME',
                listPrice: 164.5,
                effectiveUnitPrice: 164.5,
                offer: '',
                parseStatus: 'none',
                snapshotDate: '2026-09-12',
              },
            ],
            hasOffer: true,
            bestStore: 'PARKNSHOP',
            offerType: 'buy_n_save_amount',
            offerPattern: 'none',
            parseStatus: 'parsed',
          }))
        : [],
      total: query ? itemCount : 0,
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

    expect(wrapper.find('.gp-search-suggestion-name').text()).toBe('花生醬 - 幼滑裝 340克');
    expect(wrapper.find('.gp-search-suggestion-chip').text()).toBe('Meadows');
    expect(wrapper.find('.gp-search-suggestion-category').text()).toBe('個人護理 / 女士衛生用品 / 衛生巾');
    expect(mockedSearchProducts).toHaveBeenLastCalledWith(expect.objectContaining({
      q: '花生醬',
      offerOnly: false,
      pageSize: 12,
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

  // 4. 商品卡片必須在左側圖片旁顯示品牌、含單位名稱、最低價主卡及其他門店。
  it('shows brand, unit-inclusive name, and a featured lowest-price panel', async () => {
    const wrapper = mount(Page, { global: { plugins: [createPinia(), i18n], stubs: { AppIcon: true } } });
    await flushPromises();

    await wrapper.find('.gp-sinput').setValue('花生醬');
    await wrapper.find('.gp-search-row').trigger('submit');
    await flushPromises();

    expect(wrapper.find('.gp-card-name').text().replace(/\s+/g, ' ').trim()).toContain('花生醬 - 幼滑裝');
    expect(wrapper.find('.gp-card-unit').text()).toBe('340克');
    expect(wrapper.find('.gp-card-brand-chip').text()).toBe('Meadows');
    expect(wrapper.find('.gp-card-best-badge').text()).toContain('全場最低');
    expect(wrapper.find('.gp-card-best-store').text()).toBe('百佳');
    expect(wrapper.find('.gp-card-best-compare').text()).toContain('原價');
    expect(wrapper.find('.gp-card-other-row').text()).toContain('惠康');
  });

  // 5. 搜尋建議預設顯示 6 項，點選後可展開至 12 項。
  it('expands search suggestions from six to twelve items', async () => {
    mockedSearchProducts.mockImplementation((params) => Promise.resolve(createSearchResponse(params.q, 12)));
    const wrapper = mount(Page, { global: { plugins: [createPinia(), i18n], stubs: { AppIcon: true } } });
    await flushPromises();

    await wrapper.find('.gp-sinput').setValue('花生醬');
    await vi.advanceTimersByTimeAsync(180);
    await flushPromises();

    expect(wrapper.findAll('.gp-search-suggestion')).toHaveLength(6);
    await wrapper.find('.gp-search-suggestions-more').trigger('mousedown');
    expect(wrapper.findAll('.gp-search-suggestion')).toHaveLength(12);
  });
});

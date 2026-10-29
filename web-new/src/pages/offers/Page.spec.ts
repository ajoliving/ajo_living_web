/*
 * 綜合優惠頁搜尋測試。
 * 1. 確認一般搜尋不會排除無優惠商品。
 * 2. 確認輸入文字會顯示品牌及包含規格的商品名稱。
 * 3. 確認熱門分類標題及桌面、手機篩選入口。
 * 4. 確認商品卡只顯示實際原價，不顯示跨店比較 fallback。
 */
import { flushPromises, mount } from '@vue/test-utils';
import { createPinia, setActivePinia } from 'pinia';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { fetchSupermarketSummary, searchSupermarketProducts } from '@/httpapis/supermarket-offers';
import i18n from '@/i18n';

import Page from './Page.vue';
import categoryIconsSvg from './category-icons.svg?raw';

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: vi.fn() }),
  useRoute: () => ({ query: {} }),
}));

vi.mock('@/httpapis/auth-session', () => ({
  readStoredAccessToken: () => '',
}));

vi.mock('@/httpapis/wallet', () => ({
  fetchPublicDisplayAds: vi.fn().mockResolvedValue({ data: { data: { items: [] } } }),
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
          categories: ['個人護理', '保健食品', '家居用品', '米油雜貨', '奶類乳品', '零食飲品']
            .map((value) => ({ value, count: 1 })),
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

  it('shows category, store, and discount sorting in the mobile filter rail', async () => {
    const wrapper = mount(Page, { global: { plugins: [createPinia(), i18n], stubs: { AppIcon: true } } });
    await flushPromises();

    const mobileSelects = wrapper.findAll('.gp-mobile-filter-rail select');
    expect(mobileSelects).toHaveLength(3);
    expect(mobileSelects.map((select) => select.attributes('aria-label'))).toEqual([
      '商品分類',
      '商店',
      '排序方式',
    ]);
  });

  it('keeps desktop filters but removes desktop sorting and view switching', async () => {
    const wrapper = mount(Page, { global: { plugins: [createPinia(), i18n], stubs: { AppIcon: true } } });
    await flushPromises();

    expect(wrapper.find('.gp-toolbar .gp-category-select').exists()).toBe(true);
    expect(wrapper.find('.gp-toolbar .gp-store-select').exists()).toBe(true);
    expect(wrapper.find('.gp-toolbar .gp-brand-select').exists()).toBe(true);
    expect(wrapper.find('.gp-toolbar .gp-order-select').exists()).toBe(false);
    expect(wrapper.find('.gp-view-switch').exists()).toBe(false);
    expect(wrapper.find('.gp-mobile-sort-select').exists()).toBe(true);
  });

  // 2.1 每個快捷分類均有可渲染、同色描邊的 Lucide 圖示。
  it('renders six local category SVG icons matching the category labels', async () => {
    const wrapper = mount(Page, { global: { plugins: [createPinia(), i18n], stubs: { AppIcon: true } } });
    await flushPromises();

    const tiles = wrapper.findAll('.gp-cat-tile');
    const expectedIcons = ['soap-dispenser-droplet', 'pill', 'spray-can', 'wheat', 'milk', 'cookie'];
    const sprite = new DOMParser().parseFromString(categoryIconsSvg, 'image/svg+xml');

    expect(sprite.querySelector('parsererror')).toBeNull();
    expect(sprite.querySelectorAll('symbol')).toHaveLength(6);
    expect(wrapper.find('.gp-cat-nav-title').text()).toBe('熱門分類');
    expect(tiles.map((tile) => tile.text())).toEqual([
      '個人護理', '保健食品', '家居用品', '米油雜貨', '奶類乳品', '零食飲品',
    ]);
    for (const [index, icon] of expectedIcons.entries()) {
      const svg = tiles[index]!.find('svg');
      expect(svg.attributes()).toMatchObject({
        width: '22',
        height: '22',
        stroke: 'currentColor',
        'aria-hidden': 'true',
      });
      const iconUrl = new URL(svg.find('use').attributes('href')!, 'http://localhost');
      expect(iconUrl.pathname).toContain('category-icons.svg');
      expect(iconUrl.hash).toBe(`#${icon}`);
      expect(sprite.getElementById(icon)?.querySelector('path')?.getAttribute('d')).toBeTruthy();
    }
  });

  it('scrolls the selected category rail in either direction', async () => {
    mockedSearchProducts.mockImplementation((params) =>
      Promise.resolve(createSearchResponse(params.category ? '分類商品' : params.q)),
    );
    const wrapper = mount(Page, { global: { plugins: [createPinia(), i18n], stubs: { AppIcon: true } } });
    await flushPromises();
    await flushPromises();

    const rails = wrapper.findAll('.gp-product-track.is-rail');
    expect(rails).toHaveLength(6);
    const targetRail = rails[1]!;
    const scrollBy = vi.fn();
    Object.defineProperty(targetRail.element, 'clientWidth', { configurable: true, value: 400 });
    Object.defineProperty(targetRail.element, 'scrollBy', { configurable: true, value: scrollBy });
    const section = wrapper.findAll('.gp-product-section')[1]!;

    await section.find('.gp-product-rail-control[aria-label="向右捲動商品"]').trigger('click');
    expect(scrollBy).toHaveBeenLastCalledWith({ left: 340, behavior: 'smooth' });

    await section.find('.gp-product-rail-control[aria-label="向左捲動商品"]').trigger('click');
    expect(scrollBy).toHaveBeenLastCalledWith({ left: -340, behavior: 'smooth' });
  });

  // 2.2 分類入口進入分類頁後隱藏，改由分類標題的返回入口回到首頁。
  it('enters a category page and returns to the home sections', async () => {
    const wrapper = mount(Page, { global: { plugins: [createPinia(), i18n], stubs: { AppIcon: true } } });
    await flushPromises();

    const tile = wrapper.find('.gp-cat-tile');
    await tile.trigger('click');
    await flushPromises();

    expect(mockedSearchProducts).toHaveBeenCalledWith(expect.objectContaining({ category: '個人護理' }));
    expect(wrapper.find('.gp-cat-nav').exists()).toBe(false);
    expect(wrapper.find('.gp-category-heading h2').text()).toBe('個人護理');

    await wrapper.find('.gp-category-heading button').trigger('click');
    await flushPromises();

    expect(mockedSearchProducts).toHaveBeenCalledWith(expect.objectContaining({ category: '' }));
    expect(wrapper.find('.gp-cat-nav').exists()).toBe(true);
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
    expect(mockedSearchProducts).toHaveBeenCalledWith(expect.objectContaining({
      q: '花生醬',
      offerOnly: false,
      pageSize: 12,
    }));

    await wrapper.find('.gp-search-suggestion').trigger('mousedown');
    await flushPromises();

    expect(wrapper.find<HTMLInputElement>('.gp-sinput').element.value).toBe('花生醬 - 幼滑裝 340克');
    expect(mockedSearchProducts).toHaveBeenCalledWith(expect.objectContaining({
      q: '花生醬 - 幼滑裝 340克',
      offerOnly: false,
      pageSize: 20,
    }));
  });

  // 4. 商品卡片顯示品牌、含規格名稱、最低價門店、實付價與節省金額。
  it('shows brand, unit-inclusive name, and a featured lowest-price panel', async () => {
    const wrapper = mount(Page, { global: { plugins: [createPinia(), i18n], stubs: { AppIcon: true } } });
    await flushPromises();

    await wrapper.find('.gp-sinput').setValue('花生醬');
    await wrapper.find('.gp-search-row').trigger('submit');
    await flushPromises();

    expect(wrapper.find('.gp-card-name').text().replace(/\s+/g, ' ').trim()).toContain('花生醬 - 幼滑裝');
    expect(wrapper.find('.gp-card-brand-chip').text()).toBe('Meadows');
    expect(wrapper.find('.gp-card-best-store').text()).toBe('百佳');
    expect(wrapper.find('.gp-card-best-price').text()).toContain('HK$ 39.00');
    expect(wrapper.find('.gp-card-best-compare').text()).toContain('原價');
    expect(wrapper.find('.gp-card-best-compare').text()).toContain('$64.00');
    expect(wrapper.find('.gp-card-best-offer').text()).toContain('每件慳');
    expect(wrapper.find('.gp-card-other-row').exists()).toBe(false);
  });

  it('does not use another store as a comparison-price fallback', async () => {
    const response = createSearchResponse('花生醬');
    const product = response.data.data.items[0]!;
    const firstStorePrice = product.storePrices?.[0];
    expect(firstStorePrice).toBeDefined();
    product.listPrice = 39;
    firstStorePrice!.listPrice = 39;
    firstStorePrice!.effectiveUnitPrice = 39;
    mockedSearchProducts.mockResolvedValue(response);

    const wrapper = mount(Page, { global: { plugins: [createPinia(), i18n], stubs: { AppIcon: true } } });
    await flushPromises();
    await wrapper.find('.gp-sinput').setValue('花生醬');
    await wrapper.find('.gp-search-row').trigger('submit');
    await flushPromises();

    const card = wrapper.find('.gp-card');
    expect(card.find('.gp-card-name-text').text()).toContain('花生醬 - 幼滑裝 340克');
    expect(card.find('.gp-card-brand-chip').text()).toBe('Meadows');
    expect(card.find('.gp-card-best-compare').exists()).toBe(false);
    expect(card.text()).not.toMatch(/他店|其他店鋪/);
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

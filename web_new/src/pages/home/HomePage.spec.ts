/*
 * 首頁入口與公開資料回歸測試。
 * 1. 驗證既有功能連結及優惠商品排序與正式詳情路徑。
 * 2. 驗證兩個資料區獨立載入、失敗重試及空態。
 * 3. 驗證圖片失效回退及繁體中文、English 切換。
 */
import { enableAutoUnmount, flushPromises, mount } from '@vue/test-utils';
import { createPinia, setActivePinia } from 'pinia';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { createMemoryHistory, createRouter } from 'vue-router';

import { fetchPropertySaleListings } from '@/httpapis/properties';
import { fetchSupermarketSummary } from '@/httpapis/supermarket_offers';
import i18n, { applyLocale } from '@/i18n';
import type { PropertyListingSummaryResponse } from '@/model/property';
import type { SupermarketProduct } from '@/model/supermarket_offers';
import { usePreferenceStore } from '@/stores/preferences';

import HomePage from './HomePage.vue';

vi.mock('@/httpapis/properties', () => ({
  fetchPropertySaleListings: vi.fn(),
}));

vi.mock('@/httpapis/supermarket_offers', () => ({
  fetchSupermarketSummary: vi.fn(),
}));

const mockedFetchProperties = vi.mocked(fetchPropertySaleListings);
const mockedFetchSummary = vi.mocked(fetchSupermarketSummary);
type PropertyResponse = Awaited<ReturnType<typeof fetchPropertySaleListings>>;
type SummaryResponse = Awaited<ReturnType<typeof fetchSupermarketSummary>>;

enableAutoUnmount(afterEach);

// 1. 建立正式商品欄位，價格順序刻意不同以確認不在首頁重排。
const createProduct = (code: string, price: number, bestStore = 'LUNGFUNG'): SupermarketProduct => ({
  code,
  name: `商品 ${code}`,
  brand: 'Test Brand',
  subtitle: '400克',
  imageUrl: `/products/${code}.jpg`,
  category1: '食品',
  category2: '',
  category3: '',
  minPrice: price,
  maxPrice: price + 50,
  priceDiff: 50,
  diffPercent: 20,
  listPrice: price + 20,
  effectiveUnitPrice: price + 1,
  bestEffectiveUnitPrice: price,
  discountAmount: 20,
  discountRate: 20,
  stores: [bestStore],
  hasOffer: true,
  bestStore,
  offerType: 'discount',
  offerPattern: '',
  parseStatus: 'parsed',
});

// 2. 建立具有中英文標題及封面的真實樓盤回應形狀。
const createProperty = (): PropertyListingSummaryResponse => ({
  listing_id: 'property-live-7',
  display_number: 7,
  module: 'property_sale',
  title: '海景單位',
  summary: '兩房單位',
  district_code: 'sha_tin',
  publisher_identity_type: 'owner',
  publication_status: 'active',
  business_status: 'available',
  updated_at: '2026-11-19T09:00:00Z',
  cover_image: {
    media_asset_id: 'property-cover',
    url: '/properties/cover.jpg',
    is_cover: true,
    sort_order: 0,
  },
  property_sale: {
    transaction_type: 'sale',
    property_type: 'residential',
    estate_name: '海景大廈',
    address_text: '海景街一號',
    title_en: 'Seaview Home',
    asking_price_hkd: 6000000,
    usable_area_sqft: 520,
    bedroom_count: 2,
    living_room_count: 1,
    bathroom_count: 1,
    floor_level: 'middle',
    direction: 'south',
    building_age: '20',
    feature_tags: [],
    contact_method: 'phone',
    publisher_role_label: '業主',
  },
});

// 3. 包裝 API 回應；Axios 傳輸欄位不屬於此元件的測試邊界。
const summaryResponse = (products: SupermarketProduct[] = []): SummaryResponse => ({
  data: {
    data: { bestDiscounts: products },
  },
}) as SummaryResponse;

const propertyResponse = (items: PropertyListingSummaryResponse[] = []): PropertyResponse => ({
  data: {
    data: { items },
  },
}) as PropertyResponse;

// 4. 掛載正式 RouterLink，確認可導覽而非僅驗證按鈕文字。
const mountHome = async () => {
  const pinia = createPinia();
  setActivePinia(pinia);
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [{ path: '/:pathMatch(.*)*', component: { template: '<div />' } }],
  });
  await router.push('/');
  await router.isReady();
  const wrapper = mount(HomePage, {
    global: { plugins: [pinia, i18n, router] },
  });
  return { wrapper, router, preferences: usePreferenceStore(pinia) };
};

describe('HomePage', () => {
  beforeEach(() => {
    applyLocale('zh-HK');
    mockedFetchProperties.mockReset().mockResolvedValue(propertyResponse([createProperty()]));
    mockedFetchSummary.mockReset().mockResolvedValue(summaryResponse([
      createProduct('P003', 178),
      createProduct('P001', 12, 'AEON'),
      createProduct('P002', 38),
      createProduct('P004', 1),
    ]));
  });

  afterEach(() => {
    applyLocale('zh-HK');
    window.localStorage.clear();
  });

  it('opens only existing services and keeps My Building as the main action', async () => {
    const { wrapper, router } = await mountHome();
    await flushPromises();

    expect(wrapper.findAll('a.home-shortcut').map((link) => link.attributes('href'))).toEqual([
      '/building',
      '/supermarket-offers',
      '/properties',
      '/serviced-residences',
      '/furniture',
      '/trend',
    ]);
    const buildingAction = wrapper.findAll('a[href="/building"]')
      .find((link) => !link.classes().includes('home-shortcut'));
    expect(buildingAction).toBeDefined();
    await buildingAction!.trigger('click');
    await flushPromises();
    expect(router.currentRoute.value.path).toBe('/building');
    expect(wrapper.text()).not.toMatch(/即將開放|待補路由|租務管理|智能家居產品/);
  });

  it('keeps the first three upstream offers and links cards to their actual details', async () => {
    const { wrapper } = await mountHome();
    await flushPromises();

    const offers = wrapper.findAll('a.home-offer-card');
    expect(offers.map((card) => card.attributes('href'))).toEqual([
      '/supermarket-offers/products/P003',
      '/supermarket-offers/products/P001',
      '/supermarket-offers/products/P002',
    ]);
    expect(offers[0]!.text()).toContain('龍豐');
    expect(offers[1]!.text()).toContain('AEON');
    expect(offers[0]!.get('.home-offer-price').text()).toContain('HK$ 178.00');
    expect(wrapper.get('a.home-property-card').attributes('href')).toBe('/properties/property-live-7');
    expect(wrapper.get('a.home-property-card').text()).toContain('海景單位');
    expect(mockedFetchProperties).toHaveBeenCalledWith({ page: 1, page_size: 3, sort_by: 'latest' });
  });

  it('encodes reserved characters in product detail URLs', async () => {
    mockedFetchSummary.mockResolvedValue(summaryResponse([createProduct('P/003 ?A', 12)]));
    const { wrapper } = await mountHome();
    await flushPromises();

    expect(wrapper.get('a.home-offer-card').attributes('href'))
      .toBe('/supermarket-offers/products/P%2F003%20%3FA');
  });

  it('renders working navigation while both data requests are still pending', async () => {
    let finishOffers!: (value: SummaryResponse) => void;
    let finishProperties!: (value: PropertyResponse) => void;
    mockedFetchSummary.mockReturnValue(new Promise((resolve) => { finishOffers = resolve; }));
    mockedFetchProperties.mockReturnValue(new Promise((resolve) => { finishProperties = resolve; }));
    const { wrapper, router } = await mountHome();

    expect(mockedFetchSummary).toHaveBeenCalledTimes(1);
    expect(mockedFetchProperties).toHaveBeenCalledTimes(1);
    expect(wrapper.get('section.home-offers').attributes('aria-busy')).toBe('true');
    expect(wrapper.get('section.home-properties').attributes('aria-busy')).toBe('true');
    await wrapper.get('a.home-shortcut[href="/furniture"]').trigger('click');
    await flushPromises();
    expect(router.currentRoute.value.path).toBe('/furniture');

    finishOffers(summaryResponse([createProduct('P001', 12)]));
    await flushPromises();
    expect(wrapper.findAll('a.home-offer-card')).toHaveLength(1);
    expect(wrapper.get('section.home-properties').attributes('aria-busy')).toBe('true');
    finishProperties(propertyResponse([]));
    await flushPromises();
  });

  it('retries failed offers without refetching or hiding successful properties', async () => {
    mockedFetchSummary.mockRejectedValueOnce(new Error('summary unavailable'));
    const { wrapper } = await mountHome();
    await flushPromises();

    expect(wrapper.get('section.home-offers').attributes('aria-busy')).toBe('false');
    expect(wrapper.findAll('a.home-offer-card')).toHaveLength(0);
    expect(wrapper.findAll('a.home-property-card')).toHaveLength(1);
    await wrapper.get('section.home-offers button').trigger('click');
    await flushPromises();
    expect(mockedFetchSummary).toHaveBeenCalledTimes(2);
    expect(mockedFetchProperties).toHaveBeenCalledTimes(1);
    expect(wrapper.findAll('a.home-offer-card')).toHaveLength(3);
  });

  it('treats an upstream-unavailable summary as retryable instead of a successful empty result', async () => {
    const unavailable = summaryResponse();
    unavailable.data.data.metadata = { upstreamAvailable: false };
    mockedFetchSummary.mockResolvedValueOnce(unavailable);
    const { wrapper } = await mountHome();
    await flushPromises();

    expect(wrapper.get('section.home-offers').text()).toContain('暫時無法取得');
    expect(wrapper.get('section.home-offers').text()).not.toContain('暫時沒有超市優惠');
    expect(wrapper.findAll('a.home-property-card')).toHaveLength(1);
    await wrapper.get('section.home-offers button').trigger('click');
    await flushPromises();
    expect(mockedFetchSummary).toHaveBeenCalledTimes(2);
    expect(mockedFetchProperties).toHaveBeenCalledTimes(1);
    expect(wrapper.findAll('a.home-offer-card')).toHaveLength(3);
  });

  it('retries failed properties without affecting successful offers', async () => {
    mockedFetchProperties.mockRejectedValueOnce(new Error('properties unavailable'));
    const { wrapper } = await mountHome();
    await flushPromises();

    expect(wrapper.findAll('a.home-offer-card')).toHaveLength(3);
    expect(wrapper.findAll('a.home-property-card')).toHaveLength(0);
    await wrapper.get('section.home-properties button').trigger('click');
    await flushPromises();
    expect(mockedFetchProperties).toHaveBeenCalledTimes(2);
    expect(mockedFetchSummary).toHaveBeenCalledTimes(1);
    expect(wrapper.findAll('a.home-property-card')).toHaveLength(1);
  });

  it('shows empty states without inventing products or property listings', async () => {
    mockedFetchSummary.mockResolvedValue(summaryResponse());
    mockedFetchProperties.mockResolvedValue(propertyResponse());
    const { wrapper } = await mountHome();
    await flushPromises();

    expect(wrapper.findAll('a.home-offer-card, a.home-property-card')).toHaveLength(0);
    expect(wrapper.get('section.home-offers').text()).toMatch(/暫|沒有/);
    expect(wrapper.get('section.home-properties').text()).toMatch(/暫|沒有/);
    expect(wrapper.findAll('a.home-shortcut')).toHaveLength(6);
  });

  it('replaces broken product and property images while preserving usable cards', async () => {
    const { wrapper } = await mountHome();
    await flushPromises();

    const offer = wrapper.get('a.home-offer-card');
    const property = wrapper.get('a.home-property-card');
    await offer.get('img').trigger('error');
    await property.get('img').trigger('error');
    expect(offer.find('img').exists()).toBe(false);
    expect(property.find('img').exists()).toBe(false);
    expect(offer.attributes('href')).toBe('/supermarket-offers/products/P003');
    expect(property.attributes('href')).toBe('/properties/property-live-7');
    expect(offer.text()).toContain('178');
    expect(property.text()).toContain('海景單位');
  });

  it('switches visible copy, store names and property titles without reloading APIs', async () => {
    const { wrapper, preferences } = await mountHome();
    await flushPromises();

    const chineseShortcut = wrapper.get('a.home-shortcut[href="/building"]').text();
    preferences.setLocale('en');
    applyLocale('en');
    await flushPromises();
    expect(wrapper.get('a.home-shortcut[href="/building"]').text()).not.toBe(chineseShortcut);
    expect(wrapper.get('a.home-offer-card').text()).toContain('Lung Fung');
    expect(wrapper.get('a.home-property-card').text()).toContain('Seaview Home');
    expect(wrapper.text()).not.toContain('home.dashboard.');
    expect(mockedFetchSummary).toHaveBeenCalledTimes(1);
    expect(mockedFetchProperties).toHaveBeenCalledTimes(1);
  });
});

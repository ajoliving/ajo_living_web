/*
 * 首頁入口與公開資料回歸測試。
 * 1. 驗證既有功能連結及優惠商品排序與正式詳情路徑。
 * 2. 驗證兩個資料區獨立載入、失敗重試及空態。
 * 3. 驗證圖片失效回退及繁體中文、English 切換。
 * 4. 驗證天氣範圍、真實零雨量、切換後的舊回應，以及獨立重試。
 */
import { enableAutoUnmount, flushPromises, mount } from '@vue/test-utils';
import { createPinia, setActivePinia } from 'pinia';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { createMemoryHistory, createRouter } from 'vue-router';

import { fetchHomeEnvironment } from '@/httpapis/home';
import { fetchPropertySaleListings } from '@/httpapis/properties';
import { addSupermarketFavorite, fetchSupermarketSummary, removeSupermarketFavorite } from '@/httpapis/supermarket_offers';
import i18n, { applyLocale } from '@/i18n';
import type { HomeEnvironment } from '@/model/home';
import type { PropertyListingSummaryResponse } from '@/model/property';
import type { SupermarketProduct } from '@/model/supermarket_offers';
import type { CurrentMemberProfile } from '@/model/user';
import { usePreferenceStore } from '@/stores/preferences';
import { useSessionStore } from '@/stores/session';

import HomePage from './HomePage.vue';

vi.mock('@/httpapis/home', () => ({
  fetchHomeEnvironment: vi.fn(),
}));

vi.mock('@/httpapis/properties', () => ({
  fetchPropertySaleListings: vi.fn(),
}));

vi.mock('@/httpapis/supermarket_offers', () => ({
  fetchSupermarketSummary: vi.fn(),
  addSupermarketFavorite: vi.fn(),
  removeSupermarketFavorite: vi.fn(),
}));

const mockedFetchProperties = vi.mocked(fetchPropertySaleListings);
const mockedFetchSummary = vi.mocked(fetchSupermarketSummary);
const mockedFetchEnvironment = vi.mocked(fetchHomeEnvironment);
type PropertyResponse = Awaited<ReturnType<typeof fetchPropertySaleListings>>;
type SummaryResponse = Awaited<ReturnType<typeof fetchSupermarketSummary>>;
type EnvironmentResponse = Awaited<ReturnType<typeof fetchHomeEnvironment>>;

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
  storePrices: [{
    store: bestStore,
    listPrice: price + 20,
    effectiveUnitPrice: price,
    offer: '',
    parseStatus: 'parsed',
    snapshotDate: '2026-11-23',
  }],
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

const cityEnvironment: HomeEnvironment = {
  scope: 'hong_kong',
  temperature: { value: 20, station: '香港天文台', observed_at: '2026-11-20T03:45:00Z' },
  forecast: { date: '2026-11-21', min: 16, max: 24, description: '大致天晴' },
};

const environmentResponse = (payload: HomeEnvironment = cityEnvironment): EnvironmentResponse => ({
  data: { data: payload },
}) as EnvironmentResponse;

// 4. 掛載正式 RouterLink，確認可導覽而非僅驗證按鈕文字。
const mountHome = async (member?: Pick<CurrentMemberProfile, 'member_status' | 'bound_building_ids'>) => {
  const pinia = createPinia();
  setActivePinia(pinia);
  const session = useSessionStore(pinia);
  if (member) {
    session.accessToken = 'test-session';
    session.me = member as CurrentMemberProfile;
  }
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [{ path: '/:pathMatch(.*)*', component: { template: '<div />' } }],
  });
  await router.push('/');
  await router.isReady();
  const wrapper = mount(HomePage, {
    global: { plugins: [pinia, i18n, router] },
  });
  return { wrapper, router, preferences: usePreferenceStore(pinia), session };
};

describe('HomePage', () => {
  beforeEach(() => {
    applyLocale('zh-HK');
    vi.mocked(addSupermarketFavorite).mockReset();
    vi.mocked(removeSupermarketFavorite).mockReset();
    mockedFetchProperties.mockReset().mockResolvedValue(propertyResponse([createProperty()]));
    mockedFetchEnvironment.mockReset().mockResolvedValue(environmentResponse());
    mockedFetchSummary.mockReset().mockResolvedValue(summaryResponse([
      createProduct('P003', 178),
      createProduct('P001', 12, 'AEON'),
      createProduct('P002', 38),
      createProduct('P004', 1),
      createProduct('P005', 0.5),
    ]));
  });

  afterEach(() => {
    applyLocale('zh-HK');
    window.localStorage.clear();
  });

  it('opens only existing services and keeps My Building as the main action', async () => {
    const { wrapper, router } = await mountHome();
    await flushPromises();

    expect(wrapper.get('#page-home').element.firstElementChild?.classList.contains('home-cover')).toBe(true);
    expect(wrapper.find('.home-toolbar').exists()).toBe(false);
    expect(wrapper.findAll('a.home-shortcut').map((link) => link.attributes('href'))).toEqual([
      '/building',
      '/properties',
      '/building?tab=access',
      '/supermarket-offers',
      '/trend',
      '/notifications?tab=conversations',
    ]);
    const buildingAction = wrapper.findAll('a[href="/building"]')
      .find((link) => !link.classes().includes('home-shortcut'));
    expect(buildingAction).toBeDefined();
    await buildingAction!.trigger('click');
    await flushPromises();
    expect(router.currentRoute.value.path).toBe('/building');
    expect(wrapper.text()).not.toMatch(/即將開放|待補路由|租務管理|智能家居產品/);
  });

  it('keeps the first four upstream offers and links cards to their actual details', async () => {
    const { wrapper } = await mountHome();
    await flushPromises();

    const offers = wrapper.findAll('.home-offer-card');
    expect(offers.map((card) => card.get('a').attributes('href'))).toEqual([
      '/supermarket-offers/products/P003',
      '/supermarket-offers/products/P001',
      '/supermarket-offers/products/P002',
      '/supermarket-offers/products/P004',
    ]);
    expect(offers[0]!.text()).toContain('龍豐');
    expect(offers[1]!.text()).toContain('AEON');
    expect(offers[0]!.get('.gp-card-best-price').text()).toContain('HK$ 178.00');
    expect(offers[0]!.get('.gp-card-brand-chip').text()).toBe('Test Brand');
    expect(offers[0]!.get('.gp-card-best-compare').text()).toContain('$198.00');
    expect(offers[0]!.get('.gp-card-best-offer').text()).toContain('$20');
    expect(wrapper.get('a.home-property-card').attributes('href')).toBe('/properties/property-live-7');
    expect(wrapper.get('a.home-property-card').text()).toContain('海景單位');
    expect(mockedFetchProperties).toHaveBeenCalledWith({ page: 1, page_size: 3, sort_by: 'latest' });
  });

  it('encodes reserved characters in product detail URLs', async () => {
    mockedFetchSummary.mockResolvedValue(summaryResponse([createProduct('P/003 ?A', 12)]));
    const { wrapper } = await mountHome();
    await flushPromises();

    expect(wrapper.get('.home-offer-card a').attributes('href'))
      .toBe('/supermarket-offers/products/P%2F003%20%3FA');
  });

  it('opens the actual product and sends guests to login only when favoriting', async () => {
    const { wrapper, router } = await mountHome();
    await flushPromises();
    await wrapper.get('.home-offer-card a').trigger('click');
    await flushPromises();
    expect(router.currentRoute.value.path).toBe('/supermarket-offers/products/P003');
    await wrapper.get('.home-offer-card .gp-card-fav').trigger('click');
    await flushPromises();
    expect(router.currentRoute.value.path).toBe('/login');
    expect(router.currentRoute.value.query.redirect).toBe('/supermarket-offers/products/P003');
    expect(addSupermarketFavorite).not.toHaveBeenCalled();
  });

  it('updates member favorites only after success and preserves the state on failure', async () => {
    const { wrapper, router } = await mountHome({ member_status: 'active', bound_building_ids: [] });
    await flushPromises();
    const button = wrapper.get('.home-offer-card .gp-card-fav');
    await button.trigger('click');
    await flushPromises();
    expect(addSupermarketFavorite).toHaveBeenCalledWith('P003');
    expect(button.attributes('aria-pressed')).toBe('true');
    expect(router.currentRoute.value.path).toBe('/');
    vi.mocked(removeSupermarketFavorite).mockRejectedValueOnce(new Error('unavailable'));
    await button.trigger('click');
    await flushPromises();
    expect(button.attributes('aria-pressed')).toBe('true');
    expect(wrapper.get('.home-offer-feedback').text()).toBe(i18n.global.t('offers.list.favoriteError'));
    await button.trigger('click');
    await flushPromises();
    expect(removeSupermarketFavorite).toHaveBeenCalledWith('P003');
    expect(button.attributes('aria-pressed')).toBe('false');
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
    expect(wrapper.findAll('.home-offer-skeleton')).toHaveLength(4);
    expect(wrapper.get('section.home-properties').attributes('aria-busy')).toBe('true');
    await wrapper.get('a.home-shortcut[href="/properties"]').trigger('click');
    await flushPromises();
    expect(router.currentRoute.value.path).toBe('/properties');

    finishOffers(summaryResponse([createProduct('P001', 12)]));
    await flushPromises();
    expect(wrapper.findAll('.home-offer-card')).toHaveLength(1);
    expect(wrapper.get('section.home-properties').attributes('aria-busy')).toBe('true');
    finishProperties(propertyResponse([]));
    await flushPromises();
  });

  it('retries failed offers without refetching or hiding successful properties', async () => {
    mockedFetchSummary.mockRejectedValueOnce(new Error('summary unavailable'));
    const { wrapper } = await mountHome();
    await flushPromises();

    expect(wrapper.get('section.home-offers').attributes('aria-busy')).toBe('false');
    expect(wrapper.findAll('.home-offer-card')).toHaveLength(0);
    expect(wrapper.findAll('a.home-property-card')).toHaveLength(1);
    await wrapper.get('section.home-offers button').trigger('click');
    await flushPromises();
    expect(mockedFetchSummary).toHaveBeenCalledTimes(2);
    expect(mockedFetchProperties).toHaveBeenCalledTimes(1);
    expect(wrapper.findAll('.home-offer-card')).toHaveLength(4);
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
    expect(wrapper.findAll('.home-offer-card')).toHaveLength(4);
  });

  it('retries failed properties without affecting successful offers', async () => {
    mockedFetchProperties.mockRejectedValueOnce(new Error('properties unavailable'));
    const { wrapper } = await mountHome();
    await flushPromises();

    expect(wrapper.findAll('.home-offer-card')).toHaveLength(4);
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

    expect(wrapper.findAll('.home-offer-card, a.home-property-card')).toHaveLength(0);
    expect(wrapper.get('section.home-offers').text()).toMatch(/暫|沒有/);
    expect(wrapper.get('section.home-properties').text()).toMatch(/暫|沒有/);
    expect(wrapper.findAll('a.home-shortcut')).toHaveLength(6);
  });

  it('replaces broken product and property images while preserving usable cards', async () => {
    const { wrapper } = await mountHome();
    await flushPromises();

    const offer = wrapper.get('.home-offer-card');
    const property = wrapper.get('a.home-property-card');
    await offer.get('img').trigger('error');
    await property.get('img').trigger('error');
    expect(offer.find('img').exists()).toBe(false);
    expect(property.find('img').exists()).toBe(false);
    expect(offer.get('a').attributes('href')).toBe('/supermarket-offers/products/P003');
    expect(property.attributes('href')).toBe('/properties/property-live-7');
    expect(offer.text()).toContain('178');
    expect(property.text()).toContain('海景單位');
  });

  it('switches visible copy, offers and listings without refetching either catalog', async () => {
    const { wrapper, preferences } = await mountHome();
    await flushPromises();

    const chineseShortcut = wrapper.get('a.home-shortcut[href="/building"]').text();
    preferences.setLocale('en');
    applyLocale('en');
    await flushPromises();
    expect(wrapper.get('a.home-shortcut[href="/building"]').text()).not.toBe(chineseShortcut);
    expect(wrapper.get('.home-offer-card').text()).toContain('Lung Fung');
    expect(wrapper.get('a.home-property-card').text()).toContain('Seaview Home');
    expect(wrapper.text()).not.toContain('home.dashboard.');
    expect(mockedFetchSummary).toHaveBeenCalledTimes(1);
    expect(mockedFetchProperties).toHaveBeenCalledTimes(1);
    expect(mockedFetchEnvironment).toHaveBeenLastCalledWith({ lang: 'en', estate: 'seaview' }, expect.any(AbortSignal));
  });

  it('shows clearly labelled city weather between offers and properties for guests', async () => {
    const { wrapper } = await mountHome();
    await flushPromises();

    const panel = wrapper.get('section.home-environment');
    expect(mockedFetchEnvironment).toHaveBeenCalledWith({ lang: 'tc', estate: 'seaview' }, expect.any(AbortSignal));
    expect(wrapper.findAll('section.home-offers, section.home-environment, section.home-transport, section.home-life, section.home-properties')
      .map((section) => section.attributes('aria-labelledby'))).toEqual([
        'home-offers-title',
        'home-environment-title',
        'home-transport-title',
        'home-life-title',
        'home-properties-title',
      ]);
    expect(panel.text()).toContain('香港天氣');
    expect(panel.text()).toContain('明日全港預報');
    expect(panel.text()).toContain('16–24 °C');
    expect(panel.get('time').text()).toBe('11:45');
    expect(panel.find('[data-metric="rainfall"]').exists()).toBe(false);
    expect(panel.find('[data-metric="air-quality"]').exists()).toBe(false);
  });

  it('shows the active property and preserves a measured zero rainfall', async () => {
    mockedFetchEnvironment.mockResolvedValue(environmentResponse({
      ...cityEnvironment,
      scope: 'building',
      building_name: '柏蔚山',
      temperature: { value: 19, station: '跑馬地', observed_at: '2026-11-20T11:45:00+08:00' },
      rainfall: { value: 0, district: '東區', observed_at: '2026-11-20T11:30:00+08:00' },
      air_quality: { value: '3', station: '東區監測站', observed_at: '2026-11-20T11:00:00+08:00' },
    }));
    const { wrapper } = await mountHome({ member_status: 'active', bound_building_ids: ['owner-building'] });
    await flushPromises();

    const panel = wrapper.get('section.home-environment');
    expect(mockedFetchEnvironment).toHaveBeenCalledWith(
      { lang: 'tc', building_id: 'owner-building' },
      expect.any(AbortSignal),
    );
    expect(panel.text()).toContain('社區環境');
    expect(panel.text()).toContain('柏蔚山');
    expect(panel.get('[data-metric="temperature"]').text()).toContain('跑馬地');
    expect(panel.get('[data-metric="rainfall"]').text()).toContain('0 毫米');
    expect(panel.get('[data-metric="air-quality"]').text()).toContain('東區監測站');
    expect(panel.findAll('[data-metric]')).toHaveLength(4);
  });

  it('does not submit saved property context for a restricted account', async () => {
    await mountHome({ member_status: 'pending_review', bound_building_ids: ['stored-building'] });
    await flushPromises();

    expect(mockedFetchEnvironment).toHaveBeenCalledWith({ lang: 'tc', estate: 'seaview' }, expect.any(AbortSignal));
  });

  it('ignores a late response after switching or clearing the current property', async () => {
    let finishOld!: (value: EnvironmentResponse) => void;
    mockedFetchEnvironment
      .mockReturnValueOnce(new Promise((resolve) => { finishOld = resolve; }))
      .mockResolvedValueOnce(environmentResponse({ ...cityEnvironment, scope: 'building', building_name: '目前物業' }));
    const { wrapper, session } = await mountHome({ member_status: 'active', bound_building_ids: ['previous'] });

    const oldSignal = mockedFetchEnvironment.mock.calls[0]![1];
    session.me!.bound_building_ids = ['current'];
    await flushPromises();
    expect(oldSignal.aborted).toBe(true);
    expect(wrapper.get('.home-environment').text()).toContain('目前物業');
    finishOld(environmentResponse({ ...cityEnvironment, scope: 'building', building_name: '舊物業' }));
    await flushPromises();
    expect(wrapper.get('.home-environment').text()).not.toContain('舊物業');
    session.clearSession();
    await flushPromises();
    expect(wrapper.get('.home-environment').text()).not.toContain('目前物業');
    expect(mockedFetchEnvironment).toHaveBeenLastCalledWith({ lang: 'tc', estate: 'seaview' }, expect.any(AbortSignal));
  });

  it('retries an empty or failed environment response without disturbing other sections', async () => {
    mockedFetchEnvironment
      .mockResolvedValueOnce(environmentResponse({ scope: 'hong_kong' }))
      .mockRejectedValueOnce(new Error('weather unavailable'));
    const { wrapper } = await mountHome();
    await flushPromises();

    const panel = wrapper.get('.home-environment');
    expect(panel.text()).toContain('暫時無法取得環境資料');
    await panel.get('button').trigger('click');
    await flushPromises();
    expect(panel.text()).toContain('暫時無法取得環境資料');
    await panel.get('button').trigger('click');
    await flushPromises();
    expect(panel.findAll('[data-metric]')).toHaveLength(2);
    expect(mockedFetchEnvironment).toHaveBeenCalledTimes(3);
    expect(mockedFetchSummary).toHaveBeenCalledTimes(1);
    expect(mockedFetchProperties).toHaveBeenCalledTimes(1);
  });

  it('renders a public station and life information without suggesting a nearby hospital', async () => {
    mockedFetchEnvironment.mockResolvedValue(environmentResponse({
      ...cityEnvironment,
      transport: {
        station: '北角',
        line: '港島綫',
        nearby: false,
        delayed: true,
        arrivals: [{ destination: '堅尼地城', arrival_at: new Date(Date.now() + 60000).toISOString() }],
      },
      holiday: { name: '聖誕節', date: '2026-12-25', is_today: false },
      emergency: { hospital: '東區尤德夫人那打素醫院', wait: '1 小時', category: '45', nearby: false },
    }));
    const { wrapper } = await mountHome();
    await flushPromises();

    const transport = wrapper.get('.home-transport');
    const life = wrapper.get('.home-life');
    expect(transport.text()).toContain('公開車站');
    expect(transport.text()).toContain('往 堅尼地城');
    expect(transport.text()).toContain('列車服務有延誤');
    expect(transport.text()).not.toContain('附近港鐵站');
    expect(life.text()).toContain('下一個公眾假期');
    expect(life.text()).toContain('聖誕節');
    expect(life.text()).toContain('次緊急及非緊急');
    expect(life.text()).toContain('公開輪候資料');
    expect(life.text()).toContain('暫無附近車位資料');
  });

  it('links to the four public dashboard sections even when readings are unavailable', async () => {
    mockedFetchEnvironment.mockRejectedValue(new Error('public readings unavailable'));
    const { wrapper } = await mountHome();
    await flushPromises();

    const links = wrapper.findAll('.home-transport a, .home-life a');
    expect(links.map((link) => link.attributes('href'))).toEqual([
      'https://apis.hk.skylinedances.com/?page=transport',
      'https://apis.hk.skylinedances.com/?page=holidays',
      'https://apis.hk.skylinedances.com/?page=parking',
      'https://apis.hk.skylinedances.com/?page=health',
    ]);
    for (const link of links) {
      expect(link.attributes('target')).toBe('_blank');
      expect(link.attributes('rel')).toBe('noopener noreferrer');
    }
  });

  it('marks reference-area data and preserves a measured zero-space result', async () => {
    mockedFetchEnvironment.mockResolvedValue(environmentResponse({
      ...cityEnvironment,
      scope: 'reference',
      building_name: '柏蔚山',
      transport: { station: '北角', line: '港島綫', nearby: true, arrivals: [] },
      parking: { name: '北角停車場', spaces: 0, observed_at: '2026-11-20T11:30:00+08:00' },
      emergency: { hospital: '東區尤德夫人那打素醫院', wait: '20 分鐘', category: '3', nearby: true },
    }));
    const { wrapper } = await mountHome();
    await flushPromises();

    expect(wrapper.get('.home-environment').text()).toContain('柏蔚山');
    expect(wrapper.get('.home-environment').text()).toContain('參考區域');
    expect(wrapper.get('.home-transport').text()).toContain('參考區域');
    expect(wrapper.get('.home-life').text()).toContain('0 個空位');
    expect(wrapper.get('.home-life').text()).toContain('緊急（第 III 類）輪候中位數');
  });

  it('removes expired train arrivals and refreshes only the public information request', async () => {
    vi.useFakeTimers({ toFake: ['Date', 'setInterval', 'clearInterval'] });
    try {
      vi.setSystemTime(new Date('2026-11-20T11:30:00+08:00'));
      mockedFetchEnvironment.mockResolvedValueOnce(environmentResponse({
        ...cityEnvironment,
        transport: {
          station: '北角',
          line: '港島綫',
          nearby: false,
          arrivals: [
            { destination: '堅尼地城', arrival_at: '2026-11-20T11:30:04+08:00' },
            { destination: '柴灣', arrival_at: '2026-11-20T11:29:00+08:00' },
          ],
        },
      }));
      const { wrapper } = await mountHome();
      await flushPromises();

      const transport = wrapper.get('.home-transport');
      expect(transport.findAll('li')).toHaveLength(1);
      expect(transport.get('li').text()).toContain('1 分鐘');
      await vi.advanceTimersByTimeAsync(5000);
      expect(transport.findAll('li')).toHaveLength(0);
      expect(transport.text()).toContain('暫無預計到站班次');
      await transport.get('button').trigger('click');
      await flushPromises();
      expect(mockedFetchEnvironment).toHaveBeenCalledTimes(2);
      expect(mockedFetchSummary).toHaveBeenCalledTimes(1);
      expect(mockedFetchProperties).toHaveBeenCalledTimes(1);
    } finally {
      vi.useRealTimers();
    }
  });
});

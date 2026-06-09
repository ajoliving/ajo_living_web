<!--
 * 超市優惠主頁。
 * 1. 提供首頁、篩選、優惠、收藏、數據說明與商品詳情分頁。
 * 2. 透過 AJO 後端讀取 good-price 已部署服務資料。
 * 3. 商品詳情留在本頁左側分頁內，不再跳轉獨立路由。
-->
<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import { useSessionStore } from '@/app/stores/session';
import {
  addSupermarketFavorite,
  fetchSupermarketFavorites,
  fetchSupermarketProductDetail,
  fetchSupermarketSummary,
  removeSupermarketFavorite,
  saveSupermarketPriceAlert,
  searchSupermarketProducts,
} from '@/domains/supermarket-offers/api';
import type {
  SupermarketDailyStorePrice,
  SupermarketPriceAlert,
  SupermarketProduct,
  SupermarketProductDetail,
  SupermarketSearchResult,
  SupermarketStorePrice,
  SupermarketSummary,
} from '@/domains/supermarket-offers/model';
import AppIcon from '@/shared/components/base/AppIcon.vue';
import {
  formatSupermarketDistanceKm,
  nearestSupermarketStoreLocations,
  normalizeSupermarketStoreCode,
  SUPERMARKET_GEOLOCATION_CACHE_MAX_AGE_MS,
  SUPERMARKET_NEARBY_STORE_LIMIT,
  supermarketDirectionsUrl,
  supermarketStoreAddressSourceLinks,
} from './constants/store-locations';
import type { SupermarketNearbyStoreLocation } from './constants/store-locations';

type SupermarketTab = 'home' | 'filter' | 'offers' | 'favorites' | 'detail' | 'data';
type ProductLayout = 'list' | 'grid';
type PriceChartMode = 'effective' | 'list';
type NearbyStoreStatus = 'idle' | 'requesting' | 'ready' | 'error';

interface SupermarketNavItem {
  key: SupermarketTab;
  label: string;
  icon: 'home' | 'search' | 'star' | 'wallet' | 'view' | 'inbox';
}

interface SupermarketFilterState {
  q: string;
  category: string;
  brand: string;
  store: string;
  offerOnly: boolean;
  sort: string;
  page: number;
}

interface TrendChartPoint {
  key: string;
  x: number;
  y: number;
  date: string;
  store: string;
  value: number;
  listPrice: number;
  effectiveUnitPrice: number;
  offer?: string;
}

interface TrendChartSeries {
  store: string;
  color: string;
  points: string;
  dots: TrendChartPoint[];
}

interface TrendChartHoverItem extends TrendChartPoint {
  color: string;
}

const route = useRoute();
const router = useRouter();
const sessionStore = useSessionStore();
const pageSize = 20;
const summary = ref<SupermarketSummary | null>(null);
const summaryLoading = ref(false);
const summaryError = ref('');
const homeQuery = ref('');
const homeSuggestions = ref<SupermarketProduct[]>([]);
const homeSuggestionsOpen = ref(false);
const homeLayout = ref<ProductLayout>('list');
const searchLayout = ref<ProductLayout>('list');
const favoritesLayout = ref<ProductLayout>('grid');
const searchResult = ref<SupermarketSearchResult | null>(null);
const searchLoading = ref(false);
const searchError = ref('');
const favorites = ref<SupermarketProduct[]>([]);
const favoritesLoading = ref(false);
const favoritesError = ref('');
const detail = ref<SupermarketProductDetail | null>(null);
const detailLoading = ref(false);
const detailError = ref('');
const savingFavorite = ref(false);
const savingAlert = ref(false);
const actionMessage = ref('');
const alertFormOpen = ref(false);
const detailChartMode = ref<PriceChartMode>('effective');
const trendHoverDate = ref('');
const bestDealNearbyStores = ref<SupermarketNearbyStoreLocation[]>([]);
const bestDealNearbyStatus = ref<NearbyStoreStatus>('idle');
const bestDealNearbyMessage = ref('');
const isAuthenticated = computed(() => sessionStore.isAuthenticated);
let suggestionTimer: number | undefined;

const navItems: SupermarketNavItem[] = [
  { key: 'home', label: '首頁', icon: 'home' },
  { key: 'filter', label: '商品篩選', icon: 'search' },
  { key: 'offers', label: '優惠商品', icon: 'star' },
  { key: 'favorites', label: '我的收藏', icon: 'wallet' },
  { key: 'detail', label: '商品詳情', icon: 'view' },
  { key: 'data', label: '數據說明', icon: 'inbox' },
];

const filters = reactive<SupermarketFilterState>({
  q: '',
  category: '',
  brand: '',
  store: '',
  offerOnly: false,
  sort: 'discount',
  page: 1,
});
const offerFilters = reactive<SupermarketFilterState>({
  q: '',
  category: '',
  brand: '',
  store: '',
  offerOnly: true,
  sort: 'discount',
  page: 1,
});
const alertForm = reactive({
  targetPrice: '',
  priceMode: 'effective' as 'list' | 'effective',
  offerRequired: false,
  enabled: true,
});
const dataFields = [
  ['貨品分類1', '最高層級商品分類。'],
  ['貨品分類2', '第二層商品分類。'],
  ['貨品分類3', '第三層商品分類。'],
  ['貨品編號', '商品唯一代號。'],
  ['品牌', '商品品牌或供應商標示。'],
  ['貨品名稱', '商品名稱與規格。'],
  ['超市代號', '商店或連鎖超市代碼。'],
  ['價格', '公開價格資料中的原價或標示價格。'],
  ['優惠', '商戶提供的優惠文案，可能為空。'],
] as const;

const activeTab = computed<SupermarketTab>(() => resolveTab(route.query.tab));
const routeProductCode = computed(() => {
  const code = route.query.code;
  return typeof code === 'string' ? code.trim() : '';
});
const activeSearchFilters = computed(() => (activeTab.value === 'offers' ? offerFilters : filters));
const featuredProducts = computed(() => summary.value?.bestDiscounts ?? summary.value?.offers ?? []);
const hotCategories = computed(() => (summary.value?.categories ?? []).slice(0, 8));
const filterCategories = computed(() => searchResult.value?.categories ?? summary.value?.categories.map((item) => item.value) ?? []);
const filterBrands = computed(() => searchResult.value?.brands ?? []);
const filterStores = computed(() => searchResult.value?.stores ?? summary.value?.stores.map((item) => item.value) ?? []);
const totalPages = computed(() => {
  if (!searchResult.value) {
    return 1;
  }
  return Math.max(1, Math.ceil(searchResult.value.total / searchResult.value.pageSize));
});
const summarySnapshotDateText = computed(() =>
  formatDisplayDate(summary.value?.metadata?.latestSnapshotDate || ''),
);
const summaryLatestUpdatedAtText = computed(() =>
  formatDisplayDateTime(summary.value?.metadata?.latestUpdatedAt || ''),
);
const detailTrendChart = computed(() => buildTrendChart(detail.value?.history ?? [], detailChartMode.value));
const detailTrendHover = computed(() => {
  const date = trendHoverDate.value;
  const chart = detailTrendChart.value;
  if (!date || !chart.hasData) {
    return null;
  }

  const items: TrendChartHoverItem[] = chart.series
    .flatMap((series) => series.dots
      .filter((point) => point.date === date)
      .map((point) => ({ ...point, color: series.color })))
    .sort((a, b) => a.value - b.value);
  if (items.length === 0) {
    return null;
  }

  const x = items[0].x;
  return {
    date,
    items,
    x,
    tooltipLeft: `${Math.min(88, Math.max(12, (x / chart.width) * 100))}%`,
  };
});

// 1. 解析目前 tab
function resolveTab(value: unknown): SupermarketTab {
  const tab = typeof value === 'string' ? value : '';
  if (['home', 'filter', 'offers', 'favorites', 'detail', 'data'].includes(tab)) {
    return tab as SupermarketTab;
  }
  return 'home';
}

// 2. 切換超市優惠分頁
const setActiveTab = async (tab: SupermarketTab): Promise<void> => {
  const query = tab === 'detail' && (detail.value?.product.code || routeProductCode.value)
    ? { tab, code: detail.value?.product.code || routeProductCode.value }
    : { tab };
  await router.push({ path: '/supermarket-offers', query });
};

// 3. 載入首頁摘要
const loadSummary = async (): Promise<void> => {
  summaryLoading.value = true;
  summaryError.value = '';
  try {
    const { data } = await fetchSupermarketSummary();
    summary.value = data.data;
  } catch {
    summaryError.value = '暫時無法載入超市優惠資料。';
  } finally {
    summaryLoading.value = false;
  }
};

// 4. 載入首頁搜尋建議
const loadHomeSuggestions = async (q: string): Promise<void> => {
  if (!q) {
    homeSuggestions.value = [];
    return;
  }
  try {
    const { data } = await searchSupermarketProducts({
      q,
      sort: 'discount',
      page: 1,
      pageSize: 6,
    });
    if (homeQuery.value.trim() === q) {
      homeSuggestions.value = data.data.items;
    }
  } catch {
    homeSuggestions.value = [];
  }
};

// 5. 載入商品搜尋
const loadSearch = async (offersOnly: boolean): Promise<void> => {
  const source = offersOnly ? offerFilters : filters;
  searchLoading.value = true;
  searchError.value = '';
  try {
    const { data } = await searchSupermarketProducts({
      q: source.q,
      category: source.category,
      brand: source.brand,
      store: source.store,
      offerOnly: offersOnly ? true : source.offerOnly,
      sort: source.sort,
      page: source.page,
      pageSize,
    });
    searchResult.value = data.data;
  } catch {
    searchError.value = '暫時無法載入商品資料。';
  } finally {
    searchLoading.value = false;
  }
};

// 6. 重設篩選條件
const resetFilters = (offersOnly: boolean): void => {
  const source = offersOnly ? offerFilters : filters;
  source.q = '';
  source.category = '';
  source.brand = '';
  source.store = '';
  source.offerOnly = offersOnly;
  source.sort = 'discount';
  source.page = 1;
  void loadSearch(offersOnly);
};

// 7. 提交首頁搜尋
const submitHomeSearch = async (): Promise<void> => {
  filters.q = homeQuery.value.trim();
  filters.category = '';
  filters.page = 1;
  homeSuggestionsOpen.value = false;
  await setActiveTab('filter');
};

// 8. 使用首頁分類搜尋
const searchByCategory = async (category: string): Promise<void> => {
  filters.q = '';
  filters.category = category;
  filters.page = 1;
  await setActiveTab('filter');
};

// 9. 提交目前分頁搜尋
const submitActiveSearch = (): void => {
  activeSearchFilters.value.page = 1;
  void loadSearch(activeTab.value === 'offers');
};

// 10. 套用目前篩選條件
const applyActiveFilters = (): void => {
  activeSearchFilters.value.page = 1;
  void loadSearch(activeTab.value === 'offers');
};

// 11. 切換搜尋頁碼
const changeSearchPage = (direction: -1 | 1): void => {
  const source = activeSearchFilters.value;
  source.page = Math.min(totalPages.value, Math.max(1, source.page + direction));
  void loadSearch(activeTab.value === 'offers');
};

// 12. 載入目前會員收藏
const loadFavorites = async (): Promise<void> => {
  if (!isAuthenticated.value) {
    favorites.value = [];
    return;
  }
  favoritesLoading.value = true;
  favoritesError.value = '';
  try {
    const { data } = await fetchSupermarketFavorites({ page: 1, pageSize: 60 });
    favorites.value = data.data.items;
  } catch {
    favoritesError.value = '暫時無法載入收藏。';
  } finally {
    favoritesLoading.value = false;
  }
};

// 13. 載入商品詳情
const loadDetail = async (code: string): Promise<void> => {
  const productCode = code.trim();
  if (!productCode) {
    detail.value = null;
    detailError.value = '';
    return;
  }
  detailLoading.value = true;
  detailError.value = '';
  actionMessage.value = '';
  try {
    const { data } = await fetchSupermarketProductDetail(productCode, 90);
    if (routeProductCode.value === productCode || activeTab.value === 'detail') {
      detail.value = data.data;
      trendHoverDate.value = '';
      resetBestDealLocation();
      fillAlertForm(data.data.alertRule ?? null);
      alertFormOpen.value = false;
    }
  } catch {
    detailError.value = '暫時無法載入商品詳情。';
  } finally {
    detailLoading.value = false;
  }
};

// 14. 打開商品詳情分頁
const openProduct = async (product: SupermarketProduct): Promise<void> => {
  await openProductCode(product.code);
};

// 15. 使用商品編號打開詳情分頁
const openProductCode = async (code: string): Promise<void> => {
  const productCode = code.trim();
  if (!productCode) {
    return;
  }
  homeSuggestionsOpen.value = false;
  await router.push({
    path: '/supermarket-offers',
    query: { tab: 'detail', code: productCode },
  });
};

// 16. 重設最優惠門店定位狀態
const resetBestDealLocation = (): void => {
  bestDealNearbyStores.value = [];
  bestDealNearbyStatus.value = 'idle';
  bestDealNearbyMessage.value = '';
};

// 17. 切換收藏狀態
const toggleFavorite = async (): Promise<void> => {
  if (!detail.value) {
    return;
  }
  if (!isAuthenticated.value) {
    await openLogin(`/supermarket-offers?tab=detail&code=${encodeURIComponent(detail.value.product.code)}`);
    return;
  }
  savingFavorite.value = true;
  actionMessage.value = '';
  try {
    if (detail.value.isFavorite) {
      await removeSupermarketFavorite(detail.value.product.code);
      detail.value.isFavorite = false;
      detail.value.product.isFavorite = false;
      favorites.value = favorites.value.filter((item) => item.code !== detail.value?.product.code);
      actionMessage.value = '已取消收藏。';
    } else {
      await addSupermarketFavorite(detail.value.product.code);
      detail.value.isFavorite = true;
      detail.value.product.isFavorite = true;
      actionMessage.value = '已加入收藏。';
    }
  } catch {
    actionMessage.value = '收藏操作失敗。';
  } finally {
    savingFavorite.value = false;
  }
};

// 18. 移除收藏商品
const removeFavorite = async (product: SupermarketProduct): Promise<void> => {
  try {
    await removeSupermarketFavorite(product.code);
    favorites.value = favorites.value.filter((item) => item.code !== product.code);
    if (detail.value?.product.code === product.code) {
      detail.value.isFavorite = false;
      detail.value.product.isFavorite = false;
    }
  } catch {
    favoritesError.value = '暫時無法移除收藏。';
  }
};

// 19. 開啟價格提示表單
const openAlertForm = async (): Promise<void> => {
  if (!detail.value) {
    return;
  }
  if (!isAuthenticated.value) {
    await openLogin(`/supermarket-offers?tab=detail&code=${encodeURIComponent(detail.value.product.code)}`);
    return;
  }
  alertFormOpen.value = true;
};

// 20. 儲存價格提示
const submitAlert = async (): Promise<void> => {
  if (!detail.value) {
    return;
  }
  const trimmedTarget = alertForm.targetPrice.trim();
  const targetPrice = trimmedTarget ? Number(trimmedTarget) : undefined;
  if (targetPrice !== undefined && (!Number.isFinite(targetPrice) || targetPrice < 0)) {
    actionMessage.value = '請輸入有效價格。';
    return;
  }
  savingAlert.value = true;
  actionMessage.value = '';
  try {
    const { data } = await saveSupermarketPriceAlert({
      productCode: detail.value.product.code,
      targetPrice,
      priceMode: alertForm.priceMode,
      offerRequired: alertForm.offerRequired,
      enabled: alertForm.enabled,
    });
    detail.value.alertRule = data.data;
    alertFormOpen.value = false;
    actionMessage.value = '價格提示已儲存。';
  } catch {
    actionMessage.value = '價格提示儲存失敗，請確認會員已綁定 Email。';
  } finally {
    savingAlert.value = false;
  }
};

// 21. 前往 AJO 登入
const openLogin = async (redirect = '/supermarket-offers?tab=favorites'): Promise<void> => {
  await router.push({ path: '/login', query: { redirect } });
};

// 22. 用現有提示填入表單
const fillAlertForm = (rule: SupermarketPriceAlert | null): void => {
  alertForm.targetPrice = rule?.targetPrice !== undefined && rule?.targetPrice !== null ? String(rule.targetPrice) : '';
  alertForm.priceMode = rule?.priceMode ?? 'effective';
  alertForm.offerRequired = rule?.offerRequired ?? false;
  alertForm.enabled = rule?.enabled ?? true;
};

// 23. 取得商品分類文字
const categoryText = (product: SupermarketProduct): string =>
  [product.category1, product.category2, product.category3].filter(Boolean).join(' / ');

// 24. 格式化港幣
const formatHKPrice = (value: number | undefined): string =>
  `HK$ ${(value ?? 0).toLocaleString('zh-HK', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;

// 25. 格式化節省金額
const formatSaving = (value: number): string =>
  `$${Math.max(0, value).toLocaleString('zh-HK', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;

// 26. 格式化顯示日期
const formatDisplayDate = (value: string): string => {
  if (!value) {
    return '';
  }
  const parts = value.split('-');
  if (parts.length === 3) {
    return `${parts[0]}年${parts[1]}月${parts[2]}日`;
  }
  return value;
};

// 27. 格式化顯示日期時間
const formatDisplayDateTime = (value: string): string => {
  if (!value) {
    return '';
  }

  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }

  return date.toLocaleString('zh-HK', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  });
};

// 28. 顯示超市名稱
const displayStore = (value: string): string => {
  const labels: Record<string, string> = {
    AEON: 'AEON',
    WELLCOME: '惠康',
    PARKNSHOP: '百佳',
    JASONS: 'Market Place',
    LUNGFUNG: '龍豐',
    DCHFOOD: '大昌食品',
    WATSONS: '屈臣氏',
    MANNINGS: '萬寧',
    SASA: '莎莎',
  };
  return labels[normalizeSupermarketStoreCode(value)] ?? value.trim();
};

// 29. 顯示分類名稱
const displayCategory = (value: string): string => {
  const normalized = value.split('/')[0]?.trim() ?? value;
  if (normalized.includes('個人護理')) {
    return '個人護理';
  }
  if (normalized.includes('飲品')) {
    return '飲品 / 水';
  }
  if (normalized.includes('糖果')) {
    return '零食 / 食品';
  }
  if (normalized.includes('奶粉')) {
    return '奶粉嬰兒';
  }
  if (normalized.includes('家居')) {
    return '家居用品';
  }
  if (normalized.includes('米')) {
    return '米油雜貨';
  }
  if (normalized.includes('粉麵')) {
    return '粉麵食品';
  }
  return normalized;
};

// 30. 標準化商品各商店價格
const normalizedStorePrices = (product: SupermarketProduct): SupermarketStorePrice[] => {
  if (product.storePrices && product.storePrices.length > 0) {
    return [...product.storePrices].sort((a, b) => {
      if (a.effectiveUnitPrice === b.effectiveUnitPrice) {
        return displayStore(a.store).localeCompare(displayStore(b.store), 'zh-HK');
      }
      return a.effectiveUnitPrice - b.effectiveUnitPrice;
    });
  }
  const stores = product.stores.length > 0 ? product.stores : [product.bestStore || ''];
  return stores.filter(Boolean).map((store) => ({
    store,
    listPrice: store === product.bestStore ? product.listPrice : product.maxPrice,
    effectiveUnitPrice: store === product.bestStore ? product.effectiveUnitPrice : product.maxPrice,
    offer: store === product.bestStore ? product.bestOffer ?? '' : '',
    parseStatus: store === product.bestStore ? product.parseStatus : 'none',
    snapshotDate: '',
  }));
};

// 31. 取得商品最低價商店
const bestStorePrices = (product: SupermarketProduct): SupermarketStorePrice[] => {
  const prices = normalizedStorePrices(product);
  if (prices.length === 0) {
    return [fallbackStorePrice(product)];
  }
  const lowest = Math.min(...prices.map((item) => item.effectiveUnitPrice));
  return prices.filter((item) => item.effectiveUnitPrice === lowest);
};

// 32. 取得商品主要價格
const primaryPrice = (product: SupermarketProduct): SupermarketStorePrice => bestStorePrices(product)[0] ?? fallbackStorePrice(product);

// 33. 建立商品後備價格
const fallbackStorePrice = (product: SupermarketProduct): SupermarketStorePrice => ({
  store: product.bestStore || product.stores[0] || '未提供商店',
  listPrice: product.listPrice ?? product.maxPrice ?? 0,
  effectiveUnitPrice: product.effectiveUnitPrice ?? product.bestEffectiveUnitPrice ?? product.minPrice ?? 0,
  offer: product.bestOffer ?? '',
  parseStatus: product.parseStatus,
  snapshotDate: '',
});

// 34. 取得商品優惠文字
const offerTexts = (product: SupermarketProduct): string[] => {
  const prices = bestStorePrices(product);
  const values = prices.map((price) => price.offer).filter(Boolean);
  if (values.length === 0 && product.bestOffer) {
    values.push(product.bestOffer);
  }
  return uniqueText(values);
};

// 35. 去除重複文字
const uniqueText = (values: string[]): string[] =>
  Array.from(new Set(values.map((value) => value.trim()).filter(Boolean)));

// 36. 計算優惠力度
const priceDiscountRate = (price: SupermarketStorePrice): number =>
  price.listPrice > 0 ? Math.max(0, ((price.listPrice - price.effectiveUnitPrice) / price.listPrice) * 100) : 0;

// 37. 取得目前商店價格排序
const currentStorePrices = (stores: SupermarketStorePrice[]): SupermarketStorePrice[] =>
  [...stores].sort((a, b) => {
    if (a.effectiveUnitPrice === b.effectiveUnitPrice) {
      return displayStore(a.store).localeCompare(displayStore(b.store), 'zh-HK');
    }
    return a.effectiveUnitPrice - b.effectiveUnitPrice;
  });

// 38. 取得有優惠的商店排序
const discountStorePrices = (stores: SupermarketStorePrice[]): SupermarketStorePrice[] =>
  currentStorePrices(stores)
    .filter((item) => priceDiscountRate(item) > 0)
    .sort((a, b) => {
      const rateDiff = priceDiscountRate(b) - priceDiscountRate(a);
      if (rateDiff !== 0) {
        return rateDiff;
      }
      const savingDiff = (b.listPrice - b.effectiveUnitPrice) - (a.listPrice - a.effectiveUnitPrice);
      if (savingDiff !== 0) {
        return savingDiff;
      }
      return a.effectiveUnitPrice - b.effectiveUnitPrice;
    });

// 39. 取得最優惠商店
const bestDealStorePrices = (stores: SupermarketStorePrice[]): SupermarketStorePrice[] => {
  const discounted = discountStorePrices(stores);
  const primary = discounted[0];
  if (!primary) {
    return currentStorePrices(stores).slice(0, 1);
  }
  const bestRate = priceDiscountRate(primary);
  return discounted.filter((item) => Math.abs(priceDiscountRate(item) - bestRate) < 0.0001);
};

// 40. 按目前位置查找最優惠門店
const requestBestDealLocation = (): void => {
  if (!detail.value) {
    return;
  }
  if (!navigator.geolocation) {
    bestDealNearbyStatus.value = 'error';
    bestDealNearbyMessage.value = '此瀏覽器不支援定位功能。';
    return;
  }

  const bestDealStores = bestDealStorePrices(detail.value.stores).map((item) => item.store);
  bestDealNearbyStores.value = [];
  bestDealNearbyStatus.value = 'requesting';
  bestDealNearbyMessage.value = bestDealStores.length > 0
    ? '正在詢問定位權限並整理最優惠門店。'
    : '未有可計算的優惠商戶。';
  if (bestDealStores.length === 0) {
    bestDealNearbyStatus.value = 'ready';
    return;
  }

  navigator.geolocation.getCurrentPosition(
    (position) => {
      bestDealNearbyStores.value = bestDealStores
        .flatMap((store) =>
          nearestSupermarketStoreLocations(
            position.coords.latitude,
            position.coords.longitude,
            [store],
            SUPERMARKET_NEARBY_STORE_LIMIT,
          ),
        )
        .sort((a, b) => a.distanceKm - b.distanceKm);
      bestDealNearbyStatus.value = 'ready';
      bestDealNearbyMessage.value = bestDealNearbyStores.value.length > 0
        ? '已按你目前位置排序。'
        : '暫時沒有最優惠商戶的可定位門店資料。';
    },
    (error) => {
      bestDealNearbyStores.value = [];
      bestDealNearbyStatus.value = 'error';
      bestDealNearbyMessage.value = geolocationErrorMessage(error);
    },
    { enableHighAccuracy: true, timeout: 10000, maximumAge: SUPERMARKET_GEOLOCATION_CACHE_MAX_AGE_MS },
  );
};

// 41. 取得定位錯誤文案
const geolocationErrorMessage = (error: GeolocationPositionError): string => {
  if (error.code === error.PERMISSION_DENIED) {
    return '未取得定位權限，請允許瀏覽器定位後再試。';
  }
  if (error.code === error.POSITION_UNAVAILABLE) {
    return '暫時無法取得目前位置。';
  }
  if (error.code === error.TIMEOUT) {
    return '定位逾時，請稍後再試。';
  }
  return '定位失敗。';
};

// 42. 按滑鼠位置更新走勢圖浮層
const handleTrendHover = (event: MouseEvent): void => {
  const chart = detailTrendChart.value;
  if (!chart.hasData || chart.dates.length === 0) {
    trendHoverDate.value = '';
    return;
  }

  const target = event.currentTarget;
  if (!(target instanceof SVGSVGElement)) {
    return;
  }

  const rect = target.getBoundingClientRect();
  const rawX = ((event.clientX - rect.left) / Math.max(1, rect.width)) * chart.width;
  const plotStart = chart.padding.left;
  const plotEnd = chart.width - chart.padding.right;
  const ratio = Math.min(1, Math.max(0, (rawX - plotStart) / Math.max(1, plotEnd - plotStart)));
  const index = Math.round(ratio * Math.max(0, chart.dates.length - 1));
  trendHoverDate.value = chart.dates[index] ?? '';
};

// 43. 清除走勢圖浮層
const clearTrendHover = (): void => {
  trendHoverDate.value = '';
};

// 44. 取得圖表價格
const priceForMode = (item: SupermarketDailyStorePrice, mode: PriceChartMode): number =>
  mode === 'effective' ? item.effectiveUnitPrice : item.listPrice;

// 45. 建立 90 日多商店趨勢圖
const buildTrendChart = (history: SupermarketDailyStorePrice[], mode: PriceChartMode) => {
  const width = 720;
  const height = 320;
  const padding = { top: 22, right: 26, bottom: 44, left: 56 };
  const dates = Array.from(new Set(history.map((item) => item.date))).sort();
  const stores = Array.from(new Set(history.map((item) => item.store))).sort();
  const values = history.map((item) => priceForMode(item, mode));
  const minValue = values.length > 0 ? Math.min(...values) : 0;
  const maxValue = values.length > 0 ? Math.max(...values) : 1;
  const span = Math.max(1, maxValue - minValue);
  const colors = ['#004482', '#006973', '#974301', '#ba1a1a', '#3f6212', '#6d28d9', '#0f766e', '#a16207'];
  const xFor = (date: string): number => {
    const index = Math.max(0, dates.indexOf(date));
    return padding.left + (index / Math.max(1, dates.length - 1)) * (width - padding.left - padding.right);
  };
  const yFor = (value: number): number =>
    padding.top + ((maxValue - value) / span) * (height - padding.top - padding.bottom);
  const series: TrendChartSeries[] = stores.map((store, index) => {
    const dots = history
      .filter((item) => item.store === store)
      .sort((a, b) => a.date.localeCompare(b.date))
      .map((item) => ({
        key: `${item.store}-${item.date}`,
        x: xFor(item.date),
        y: yFor(priceForMode(item, mode)),
        date: item.date,
        store: item.store,
        value: priceForMode(item, mode),
        listPrice: item.listPrice,
        effectiveUnitPrice: item.effectiveUnitPrice,
        offer: item.offer,
      }));
    return {
      store,
      color: colors[index % colors.length],
      points: dots.map((item) => `${item.x},${item.y}`).join(' '),
      dots,
    };
  });
  const yTicks = [0, 0.25, 0.5, 0.75, 1].map((tick) => {
    const y = padding.top + tick * (height - padding.top - padding.bottom);
    const value = maxValue - tick * span;
    return { key: String(tick), y, label: value.toFixed(1) };
  });
  const xTicks = pickDateTicks(dates, 6).map((date) => ({
    key: date,
    x: xFor(date),
    label: formatDateTick(date),
  }));
  return {
    width,
    height,
    padding,
    dates,
    stores,
    series,
    yTicks,
    xTicks,
    hasData: history.length > 0,
  };
};

// 46. 選取日期刻度
const pickDateTicks = (dates: string[], maxTicks: number): string[] => {
  if (dates.length <= maxTicks) {
    return dates;
  }
  const lastIndex = dates.length - 1;
  const indexes = new Set<number>();
  for (let index = 0; index < maxTicks; index += 1) {
    indexes.add(Math.round((index / Math.max(1, maxTicks - 1)) * lastIndex));
  }
  return [...indexes].sort((a, b) => a - b).map((index) => dates[index]).filter(Boolean);
};

// 47. 格式化日期刻度
const formatDateTick = (date: string): string => {
  const parts = date.split('-');
  if (parts.length === 3) {
    return `${parts[1]}/${parts[2]}`;
  }
  return date;
};

// 48. 初始化公開摘要
onMounted(() => {
  void loadSummary();
});

// 49. 根據首頁搜尋字載入建議
watch(homeQuery, (q) => {
  homeSuggestionsOpen.value = true;
  if (suggestionTimer !== undefined) {
    window.clearTimeout(suggestionTimer);
  }
  const trimmed = q.trim();
  if (!trimmed) {
    homeSuggestions.value = [];
    return;
  }
  suggestionTimer = window.setTimeout(() => {
    void loadHomeSuggestions(trimmed);
  }, 220);
});

// 50. 按目前分頁與商品編號載入資料
watch(
  () => [activeTab.value, routeProductCode.value] as const,
  ([tab, code], previous) => {
    const previousTab = previous?.[0] ?? '';
    const previousCode = previous?.[1] ?? '';
    if (tab === 'filter') {
      void loadSearch(false);
    }
    if (tab === 'offers') {
      void loadSearch(true);
    }
    if (tab === 'favorites' && isAuthenticated.value) {
      void loadFavorites();
    }
    if (tab === 'detail' && code && (tab !== previousTab || code !== previousCode)) {
      void loadDetail(code);
    }
  },
  { immediate: true },
);

// 51. 登入狀態變更時刷新會員資料
watch(isAuthenticated, (authenticated) => {
  if (activeTab.value === 'favorites' && authenticated) {
    void loadFavorites();
  }
  if (activeTab.value === 'detail' && routeProductCode.value) {
    void loadDetail(routeProductCode.value);
  }
});
</script>

<template>
  <main class="supermarket-page">
    <aside class="supermarket-page__sidebar">
      <div>
        <p class="supermarket-page__kicker">AJO Living</p>
        <h1>超市優惠</h1>
        <p class="supermarket-page__description">
          香港超市價格與優惠查詢。
        </p>
      </div>

      <nav class="supermarket-page__nav">
        <button
          v-for="item in navItems"
          :key="item.key"
          type="button"
          class="supermarket-page__nav-item"
          :class="activeTab === item.key ? 'supermarket-page__nav-item--active' : ''"
          @click="setActiveTab(item.key)"
        >
          <AppIcon
            :name="item.icon"
            :size="17"
          />
          <span>{{ item.label }}</span>
        </button>
      </nav>

      <div
        v-if="summary"
        class="supermarket-page__sidebar-stats"
      >
        <span>{{ summary.stats.products.toLocaleString('zh-HK') }} 件商品</span>
        <span>{{ summary.stats.offers.toLocaleString('zh-HK') }} 個優惠</span>
        <span
          v-if="summarySnapshotDateText"
          class="supermarket-page__updated-at"
        >
          最新資料日期 {{ summarySnapshotDateText }}
        </span>
      </div>
    </aside>

    <section class="supermarket-page__content">
      <section
        v-if="activeTab === 'home'"
        class="supermarket-home"
      >
        <section class="supermarket-home__search">
          <form
            class="supermarket-home__search-form"
            @submit.prevent="submitHomeSearch"
          >
            <AppIcon
              name="search"
              :size="22"
            />
            <input
              v-model="homeQuery"
              type="search"
              placeholder="搜尋商品、品牌或分類"
              @focus="homeSuggestionsOpen = true"
            >
            <button type="submit">
              搜尋
            </button>
            <div
              v-if="homeSuggestionsOpen && homeSuggestions.length > 0"
              class="supermarket-home__suggestions"
            >
              <button
                v-for="product in homeSuggestions"
                :key="product.code"
                type="button"
                @mousedown.prevent
                @click="openProduct(product)"
              >
                <span>
                  <strong>{{ product.name }}</strong>
                  <small>{{ product.brand || '未提供品牌' }}</small>
                </span>
                <em>{{ formatHKPrice(primaryPrice(product).effectiveUnitPrice) }}</em>
              </button>
            </div>
          </form>

          <div
            v-if="hotCategories.length > 0"
            class="supermarket-hot-categories"
          >
            <span>熱門分類</span>
            <button
              v-for="category in hotCategories"
              :key="category.value"
              type="button"
              @click="searchByCategory(category.value)"
            >
              {{ displayCategory(category.value) }}
            </button>
          </div>
        </section>

        <p
          v-if="summaryError"
          class="supermarket-state supermarket-state--error"
        >
          {{ summaryError }}
        </p>
        <p
          v-else-if="summaryLoading"
          class="supermarket-state"
        >
          正在載入資料。
        </p>

        <section
          v-if="summary"
          class="supermarket-stats"
        >
          <article>
            <span class="supermarket-stats__icon">品</span>
            <div>
              <strong>{{ summary.stats.products.toLocaleString('zh-HK') }}</strong>
              <p>監測商品</p>
            </div>
          </article>
          <article>
            <span class="supermarket-stats__icon">表</span>
            <div>
              <strong>{{ summary.stats.records.toLocaleString('zh-HK') }}</strong>
              <p>今日價格紀錄</p>
            </div>
          </article>
          <article>
            <span class="supermarket-stats__icon">店</span>
            <div>
              <strong>{{ summary.stats.stores.toLocaleString('zh-HK') }}</strong>
              <p>涵蓋連鎖商店</p>
            </div>
          </article>
          <article class="supermarket-stats__deal">
            <span class="supermarket-stats__icon">%</span>
            <div>
              <strong>{{ summary.stats.offers.toLocaleString('zh-HK') }}</strong>
              <p>活躍優惠</p>
            </div>
          </article>
        </section>

        <section
          v-if="summary"
          class="supermarket-section"
        >
          <div class="supermarket-section__title">
            <div>
              <p>優惠參考</p>
              <h2>最高優惠力度</h2>
            </div>
            <div class="supermarket-section__actions">
              <div class="supermarket-layout-toggle">
                <button
                  type="button"
                  :class="homeLayout === 'list' ? 'supermarket-layout-toggle__active' : ''"
                  @click="homeLayout = 'list'"
                >
                  列表
                </button>
                <button
                  type="button"
                  :class="homeLayout === 'grid' ? 'supermarket-layout-toggle__active' : ''"
                  @click="homeLayout = 'grid'"
                >
                  網格
                </button>
              </div>
              <button
                type="button"
                class="supermarket-link-button"
                @click="setActiveTab('filter')"
              >
                查看更多
              </button>
            </div>
          </div>

          <div
            v-if="homeLayout === 'list'"
            class="supermarket-product-table"
          >
            <table>
              <thead>
                <tr>
                  <th>商品名稱</th>
                  <th>優惠後</th>
                  <th>原價</th>
                  <th>優惠</th>
                  <th>優惠力度</th>
                  <th class="supermarket-product-table__stores">商店</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="product in featuredProducts.slice(0, 8)"
                  :key="product.code"
                  @click="openProduct(product)"
                >
                  <td>
                    <strong>{{ product.name }}</strong>
                    <span>{{ product.brand || '未提供品牌' }}</span>
                  </td>
                  <td>
                    <strong>{{ formatHKPrice(primaryPrice(product).effectiveUnitPrice) }}</strong>
                    <span>優惠後</span>
                  </td>
                  <td>{{ formatHKPrice(primaryPrice(product).listPrice) }}</td>
                  <td>
                    <em v-if="offerTexts(product).length > 0">
                      {{ offerTexts(product).join(' / ') }}
                    </em>
                    <span v-else>-</span>
                  </td>
                  <td>
                    <strong class="supermarket-discount">
                      -{{ priceDiscountRate(primaryPrice(product)).toFixed(0) }}%
                    </strong>
                    <span>{{ formatSaving(primaryPrice(product).listPrice - primaryPrice(product).effectiveUnitPrice) }}</span>
                  </td>
                  <td class="supermarket-product-table__stores">
                    <span class="supermarket-store-chips">
                      <span
                        v-for="store in bestStorePrices(product).map((price) => price.store)"
                        :key="store"
                      >
                        {{ displayStore(store) }}
                      </span>
                    </span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <div
            v-else
            class="supermarket-product-grid"
          >
            <article
              v-for="product in featuredProducts.slice(0, 6)"
              :key="product.code"
              class="supermarket-product-card"
            >
              <button
                type="button"
                @click="openProduct(product)"
              >
                <span>{{ product.brand || '未提供品牌' }}</span>
                <strong>{{ product.name }}</strong>
                <div class="supermarket-store-list">
                  <article
                    v-for="(price, index) in normalizedStorePrices(product).slice(0, 3)"
                    :key="`${price.store}-${index}`"
                  >
                    <div>
                      <strong>{{ displayStore(price.store) }}</strong>
                      <span v-if="price.offer">{{ price.offer }}</span>
                    </div>
                    <div>
                      <strong>{{ formatHKPrice(price.effectiveUnitPrice) }} / 件</strong>
                      <span v-if="price.listPrice > price.effectiveUnitPrice">
                        {{ formatHKPrice(price.listPrice) }}(-{{ priceDiscountRate(price).toFixed(0) }}%)
                      </span>
                    </div>
                  </article>
                </div>
              </button>
            </article>
          </div>
        </section>
      </section>

      <section
        v-else-if="activeTab === 'filter' || activeTab === 'offers'"
        class="supermarket-search-shell"
      >
        <aside class="supermarket-filter-panel">
          <div>
            <h2>篩選條件</h2>
            <p>精確搜尋</p>
          </div>
          <div class="supermarket-filter-panel__card">
            <label>
              <span>商品分類</span>
              <select
                v-model="activeSearchFilters.category"
                class="supermarket-control-select"
                @change="applyActiveFilters"
              >
                <option value="">全部</option>
                <option
                  v-for="category in filterCategories"
                  :key="category"
                  :value="category"
                >
                  {{ displayCategory(category) }}
                </option>
              </select>
            </label>
            <label>
              <span>熱門品牌</span>
              <select
                v-model="activeSearchFilters.brand"
                class="supermarket-control-select"
                @change="applyActiveFilters"
              >
                <option value="">全部</option>
                <option
                  v-for="brand in filterBrands"
                  :key="brand"
                  :value="brand"
                >
                  {{ brand }}
                </option>
              </select>
            </label>
            <label>
              <span>超級市場</span>
              <select
                v-model="activeSearchFilters.store"
                class="supermarket-control-select"
                @change="applyActiveFilters"
              >
                <option value="">全部</option>
                <option
                  v-for="store in filterStores"
                  :key="store"
                  :value="store"
                >
                  {{ displayStore(store) }}
                </option>
              </select>
            </label>
            <label class="supermarket-filter-panel__checkbox">
              <span>只顯示優惠</span>
              <input
                v-model="activeSearchFilters.offerOnly"
                type="checkbox"
                :disabled="activeTab === 'offers'"
                @change="applyActiveFilters"
              >
            </label>
            <button
              type="button"
              @click="resetFilters(activeTab === 'offers')"
            >
              重設篩選
            </button>
          </div>
        </aside>

        <section class="supermarket-search-main">
          <div class="supermarket-search-header">
            <div>
              <p>{{ activeTab === 'offers' ? '優惠參考' : '商品格價' }}</p>
              <h2>{{ activeTab === 'offers' ? '優惠商品' : '商品篩選' }}</h2>
            </div>
            <span>價格只供參考</span>
          </div>

          <form
            class="supermarket-search-controls"
            @submit.prevent="submitActiveSearch"
          >
            <div class="supermarket-search-controls__input">
              <AppIcon
                name="search"
                :size="18"
              />
              <input
                v-model="activeSearchFilters.q"
                type="search"
                placeholder="搜尋商品、品牌或分類"
              >
              <button type="submit">
                搜尋
              </button>
            </div>
            <div class="supermarket-search-controls__meta">
              <span>{{ searchLoading ? '載入中' : `找到 ${(searchResult?.total ?? 0).toLocaleString('zh-HK')} 個商品` }}</span>
              <div class="supermarket-layout-toggle supermarket-layout-toggle--small">
                <button
                  type="button"
                  :class="searchLayout === 'list' ? 'supermarket-layout-toggle__active' : ''"
                  @click="searchLayout = 'list'"
                >
                  列表
                </button>
                <button
                  type="button"
                  :class="searchLayout === 'grid' ? 'supermarket-layout-toggle__active' : ''"
                  @click="searchLayout = 'grid'"
                >
                  網格
                </button>
              </div>
              <label class="supermarket-search-controls__sort">
                排序
                <select
                  v-model="activeSearchFilters.sort"
                  class="supermarket-control-select"
                  @change="applyActiveFilters"
                >
                  <option value="discount">優惠力度</option>
                  <option value="effective">優惠後價格</option>
                  <option value="diff">價差</option>
                  <option value="name">商品名稱</option>
                  <option value="brand">品牌</option>
                </select>
              </label>
            </div>
          </form>

          <div
            v-if="searchResult"
            class="supermarket-stats-strip"
          >
            <article>
              <span>商品</span>
              <strong>{{ searchResult.stats.products.toLocaleString('zh-HK') }}</strong>
            </article>
            <article>
              <span>價格</span>
              <strong>{{ searchResult.stats.records.toLocaleString('zh-HK') }}</strong>
            </article>
            <article>
              <span>商店</span>
              <strong>{{ searchResult.stats.stores.toLocaleString('zh-HK') }}</strong>
            </article>
            <article>
              <span>優惠</span>
              <strong>{{ searchResult.stats.offers.toLocaleString('zh-HK') }}</strong>
            </article>
          </div>

          <p
            v-if="searchError"
            class="supermarket-state supermarket-state--error"
          >
            {{ searchError }}
          </p>
          <p
            v-else-if="searchLoading && !searchResult"
            class="supermarket-state"
          >
            正在載入商品。
          </p>
          <p
            v-else-if="searchResult && searchResult.items.length === 0"
            class="supermarket-state"
          >
            沒有找到符合條件的商品。
          </p>

          <div
            v-if="searchResult && searchResult.items.length > 0 && searchLayout === 'list'"
            class="supermarket-product-table"
          >
            <table>
              <thead>
                <tr>
                  <th>商品名稱</th>
                  <th>優惠後</th>
                  <th>原價</th>
                  <th>優惠</th>
                  <th>優惠力度</th>
                  <th class="supermarket-product-table__stores">商店</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="product in searchResult.items"
                  :key="product.code"
                  @click="openProduct(product)"
                >
                  <td>
                    <strong>{{ product.name }}</strong>
                    <span>{{ product.brand || '未提供品牌' }}</span>
                  </td>
                  <td>
                    <strong>{{ formatHKPrice(primaryPrice(product).effectiveUnitPrice) }}</strong>
                    <span>優惠後</span>
                  </td>
                  <td>{{ formatHKPrice(primaryPrice(product).listPrice) }}</td>
                  <td>
                    <em v-if="offerTexts(product).length > 0">
                      {{ offerTexts(product).join(' / ') }}
                    </em>
                    <span v-else>-</span>
                  </td>
                  <td>
                    <strong class="supermarket-discount">
                      -{{ priceDiscountRate(primaryPrice(product)).toFixed(0) }}%
                    </strong>
                    <span>{{ formatSaving(primaryPrice(product).listPrice - primaryPrice(product).effectiveUnitPrice) }}</span>
                  </td>
                  <td class="supermarket-product-table__stores">
                    <span class="supermarket-store-chips">
                      <span
                        v-for="store in bestStorePrices(product).map((price) => price.store)"
                        :key="store"
                      >
                        {{ displayStore(store) }}
                      </span>
                    </span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <div
            v-if="searchResult && searchResult.items.length > 0 && searchLayout === 'grid'"
            class="supermarket-product-grid"
          >
            <article
              v-for="product in searchResult.items"
              :key="product.code"
              class="supermarket-product-card"
            >
              <button
                type="button"
                @click="openProduct(product)"
              >
                <span>{{ product.brand || '未提供品牌' }}</span>
                <strong>{{ product.name }}</strong>
                <div class="supermarket-store-list">
                  <article
                    v-for="(price, index) in normalizedStorePrices(product).slice(0, 3)"
                    :key="`${price.store}-${index}`"
                  >
                    <div>
                      <strong>{{ displayStore(price.store) }}</strong>
                      <span v-if="price.offer">{{ price.offer }}</span>
                    </div>
                    <div>
                      <strong>{{ formatHKPrice(price.effectiveUnitPrice) }} / 件</strong>
                      <span v-if="price.listPrice > price.effectiveUnitPrice">
                        {{ formatHKPrice(price.listPrice) }}(-{{ priceDiscountRate(price).toFixed(0) }}%)
                      </span>
                    </div>
                  </article>
                </div>
              </button>
            </article>
          </div>

          <div
            v-if="searchResult && searchResult.total > searchResult.pageSize"
            class="supermarket-pagination"
          >
            <button
              type="button"
              :disabled="activeSearchFilters.page <= 1"
              @click="changeSearchPage(-1)"
            >
              上一頁
            </button>
            <span>第 {{ searchResult.page }} / {{ totalPages }} 頁</span>
            <button
              type="button"
              :disabled="activeSearchFilters.page >= totalPages"
              @click="changeSearchPage(1)"
            >
              下一頁
            </button>
          </div>
        </section>
      </section>

      <section
        v-else-if="activeTab === 'favorites'"
        class="supermarket-section"
      >
        <div class="supermarket-section__title">
          <div>
            <p>會員功能</p>
            <h2>我的超市收藏</h2>
          </div>
          <div
            v-if="isAuthenticated"
            class="supermarket-section__actions"
          >
            <div class="supermarket-layout-toggle">
              <button
                type="button"
                :class="favoritesLayout === 'list' ? 'supermarket-layout-toggle__active' : ''"
                @click="favoritesLayout = 'list'"
              >
                列表
              </button>
              <button
                type="button"
                :class="favoritesLayout === 'grid' ? 'supermarket-layout-toggle__active' : ''"
                @click="favoritesLayout = 'grid'"
              >
                網格
              </button>
            </div>
            <button
              type="button"
              class="supermarket-link-button"
              @click="loadFavorites"
            >
              重新整理
            </button>
          </div>
        </div>

        <div
          v-if="!isAuthenticated"
          class="supermarket-login"
        >
          <h3>登入後使用收藏與價格提示</h3>
          <p>收藏和價格提示會綁定你的 AJO 會員身份。</p>
          <button
            type="button"
            @click="openLogin()"
          >
            前往登入
          </button>
        </div>

        <p
          v-if="favoritesError"
          class="supermarket-state supermarket-state--error"
        >
          {{ favoritesError }}
        </p>
        <p
          v-else-if="favoritesLoading"
          class="supermarket-state"
        >
          正在載入收藏。
        </p>
        <p
          v-else-if="isAuthenticated && favorites.length === 0"
          class="supermarket-state"
        >
          尚未收藏商品。
        </p>

        <div
          v-if="favorites.length > 0 && favoritesLayout === 'grid'"
          class="supermarket-product-grid"
        >
          <article
            v-for="product in favorites"
            :key="product.code"
            class="supermarket-product-card"
          >
            <button
              type="button"
              @click="openProduct(product)"
            >
              <span>{{ product.brand || '未提供品牌' }}</span>
              <strong>{{ product.name }}</strong>
              <small>{{ categoryText(product) || '收藏商品' }}</small>
              <div class="supermarket-store-list">
                <article
                  v-for="(price, index) in normalizedStorePrices(product).slice(0, 2)"
                  :key="`${price.store}-${index}`"
                >
                  <div>
                    <strong>{{ displayStore(price.store) }}</strong>
                    <span v-if="price.offer">{{ price.offer }}</span>
                  </div>
                  <div>
                    <strong>{{ formatHKPrice(price.effectiveUnitPrice) }} / 件</strong>
                    <span v-if="price.listPrice > price.effectiveUnitPrice">
                      {{ formatHKPrice(price.listPrice) }}(-{{ priceDiscountRate(price).toFixed(0) }}%)
                    </span>
                  </div>
                </article>
              </div>
            </button>
            <footer>
              <button
                type="button"
                @click="openProduct(product)"
              >
                詳情
              </button>
              <button
                type="button"
                @click.stop="removeFavorite(product)"
              >
                移除
              </button>
            </footer>
          </article>
        </div>

        <div
          v-if="favorites.length > 0 && favoritesLayout === 'list'"
          class="supermarket-product-table"
        >
          <table>
            <thead>
              <tr>
                <th>商品名稱</th>
                <th>優惠後</th>
                <th>原價</th>
                <th>優惠</th>
                <th class="supermarket-product-table__stores">商店</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="product in favorites"
                :key="product.code"
                @click="openProduct(product)"
              >
                <td>
                  <strong>{{ product.name }}</strong>
                  <span>{{ product.brand || '未提供品牌' }}</span>
                </td>
                <td>
                  <strong>{{ formatHKPrice(primaryPrice(product).effectiveUnitPrice) }}</strong>
                  <span>優惠後</span>
                </td>
                <td>{{ formatHKPrice(primaryPrice(product).listPrice) }}</td>
                <td>
                  <em v-if="offerTexts(product).length > 0">
                    {{ offerTexts(product).join(' / ') }}
                  </em>
                  <span v-else>-</span>
                </td>
                <td class="supermarket-product-table__stores">
                  <span class="supermarket-store-chips">
                    <span
                      v-for="store in bestStorePrices(product).map((price) => price.store)"
                      :key="store"
                    >
                      {{ displayStore(store) }}
                    </span>
                  </span>
                </td>
                <td>
                  <button
                    type="button"
                    class="supermarket-table-action"
                    @click.stop="removeFavorite(product)"
                  >
                    移除
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      <section
        v-else-if="activeTab === 'data'"
        class="supermarket-section"
      >
        <div class="supermarket-section__title">
          <div>
            <p>DATA SOURCE</p>
            <h2>數據說明</h2>
          </div>
        </div>
        <p class="supermarket-section__note">
          本頁以公開 CSV 價格資料整理，支援 AJO 會員收藏商品及到價 / 優惠提示；價格、貨量及優惠條款以商戶公布為準。
        </p>
        <p
          v-if="summaryError"
          class="supermarket-state supermarket-state--error"
        >
          {{ summaryError }}
        </p>
        <div class="supermarket-data-grid">
          <article>
            <h3>資料概況</h3>
            <div class="supermarket-data-freshness">
              <div>
                <span>最新資料日期</span>
                <strong>{{ summarySnapshotDateText || '等待資料' }}</strong>
              </div>
              <div>
                <span>最近入庫時間</span>
                <strong>{{ summaryLatestUpdatedAtText || '等待資料' }}</strong>
              </div>
            </div>
            <div
              v-if="summary"
              class="supermarket-data-stats"
            >
              <div>
                <span>商品</span>
                <strong>{{ summary.stats.products.toLocaleString('zh-HK') }}</strong>
              </div>
              <div>
                <span>價格記錄</span>
                <strong>{{ summary.stats.records.toLocaleString('zh-HK') }}</strong>
              </div>
              <div>
                <span>商店</span>
                <strong>{{ summary.stats.stores.toLocaleString('zh-HK') }}</strong>
              </div>
              <div>
                <span>優惠</span>
                <strong>{{ summary.stats.offers.toLocaleString('zh-HK') }}</strong>
              </div>
            </div>
            <p>
              商品詳情頁會使用已入庫的 90 日歷史資料展示超市價格走勢。價格及優惠只供查詢參考。
            </p>
          </article>
          <article>
            <h3>欄位解釋</h3>
            <div class="supermarket-fields">
              <div
                v-for="[name, description] in dataFields"
                :key="name"
              >
                <strong>{{ name }}</strong>
                <span>{{ description }}</span>
              </div>
            </div>
          </article>
        </div>
        <article class="supermarket-data-source">
          <h3>門店地址資料來源</h3>
          <div class="supermarket-source-links">
            <a
              v-for="source in supermarketStoreAddressSourceLinks"
              :key="source.url"
              :href="source.url"
              target="_blank"
              rel="noreferrer"
            >
              <strong>{{ source.name }}</strong>
              <span>{{ source.url }}</span>
            </a>
          </div>
          <p>
            門店地址與營業時間來自上述商戶公開門店頁面，坐標用於商品詳情頁的附近門店排序。
          </p>
        </article>
      </section>

      <section
        v-else-if="activeTab === 'detail'"
        class="supermarket-detail"
      >
        <p
          v-if="!routeProductCode"
          class="supermarket-state"
        >
          請先從首頁、篩選或收藏選擇商品。
        </p>
        <p
          v-else-if="detailError"
          class="supermarket-state supermarket-state--error"
        >
          {{ detailError }}
        </p>
        <p
          v-else-if="detailLoading && !detail"
          class="supermarket-state"
        >
          正在載入商品詳情。
        </p>

        <template v-if="detail">
          <section class="supermarket-detail__hero">
            <div>
              <p>商品基本資訊</p>
              <h1>{{ detail.product.name }}</h1>
              <span>{{ detail.product.brand || '未提供品牌' }} · {{ categoryText(detail.product) || '未提供分類' }}</span>
            </div>
            <div class="supermarket-detail__actions">
              <button
                type="button"
                :disabled="savingFavorite"
                @click="toggleFavorite"
              >
                {{ detail.isFavorite ? '已收藏' : '收藏' }}
              </button>
              <button
                type="button"
                @click="openAlertForm"
              >
                收藏並設定提醒
              </button>
              <span v-if="actionMessage">{{ actionMessage }}</span>
            </div>
          </section>

          <section class="supermarket-detail__store-section">
            <div class="supermarket-detail__panel-title">
              <div>
                <h3>超市優惠與門店地址</h3>
                <p>先看有優惠的最新超市，再查最優惠商戶或最近門店路線。</p>
              </div>
              <span>門店列表可直接開 Google Maps</span>
            </div>

            <section class="supermarket-detail__discount-list">
              <h4>最新超市對比</h4>
              <div
                v-if="currentStorePrices(detail.stores).length > 0"
                class="supermarket-detail__discount-grid"
              >
                <article
                  v-for="store in (discountStorePrices(detail.stores).length > 0 ? discountStorePrices(detail.stores) : currentStorePrices(detail.stores)).slice(0, 4)"
                  :key="store.store"
                >
                  <strong>{{ displayStore(store.store) }}</strong>
                  <span>{{ formatHKPrice(store.effectiveUnitPrice) }}</span>
                  <small>原價 {{ formatHKPrice(store.listPrice) }}</small>
                  <em v-if="store.offer">{{ store.offer }}</em>
                </article>
              </div>
              <p
                v-else
                class="supermarket-state"
              >
                暫時沒有超市價格資料。
              </p>
            </section>

            <section class="supermarket-detail__best-panel">
              <div class="supermarket-detail__best-header">
                <div>
                  <h4>最優惠最近門店地址</h4>
                  <p>
                    {{ bestDealStorePrices(detail.stores).map((store) => `${displayStore(store.store)} ${formatHKPrice(store.effectiveUnitPrice)}`).join(' / ') || '未有可計算的優惠商戶。' }}
                  </p>
                </div>
                <button
                  type="button"
                  :disabled="bestDealStorePrices(detail.stores).length === 0 || bestDealNearbyStatus === 'requesting'"
                  @click="requestBestDealLocation"
                >
                  {{ bestDealNearbyStatus === 'requesting' ? '定位中' : '查找最優惠門店' }}
                </button>
              </div>

              <p
                v-if="bestDealNearbyMessage"
                class="supermarket-detail__location-message"
                :class="bestDealNearbyStatus === 'error' ? 'supermarket-detail__location-message--error' : ''"
              >
                {{ bestDealNearbyStatus === 'requesting' ? '定位中。' : bestDealNearbyMessage }}
              </p>

              <div
                v-if="bestDealNearbyStores.length > 0"
                class="supermarket-detail__nearby-table"
              >
                <table>
                  <thead>
                    <tr>
                      <th>#</th>
                      <th>超市</th>
                      <th>門店</th>
                      <th>地址</th>
                      <th>距離</th>
                      <th>營業時間</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr
                      v-for="(store, index) in bestDealNearbyStores"
                      :key="store.id || `${store.name}-${store.address}`"
                    >
                      <td>
                        <a
                          :href="supermarketDirectionsUrl(store)"
                          target="_blank"
                          rel="noreferrer"
                        >
                          {{ index + 1 }}
                        </a>
                      </td>
                      <td>
                        <a
                          :href="supermarketDirectionsUrl(store)"
                          target="_blank"
                          rel="noreferrer"
                        >
                          {{ displayStore(store.storeCode) }}
                        </a>
                      </td>
                      <td>
                        <a
                          :href="supermarketDirectionsUrl(store)"
                          target="_blank"
                          rel="noreferrer"
                        >
                          {{ store.name }}
                        </a>
                      </td>
                      <td>
                        <a
                          :href="supermarketDirectionsUrl(store)"
                          target="_blank"
                          rel="noreferrer"
                        >
                          {{ store.address }}
                        </a>
                      </td>
                      <td>
                        <a
                          :href="supermarketDirectionsUrl(store)"
                          target="_blank"
                          rel="noreferrer"
                        >
                          <span>{{ formatSupermarketDistanceKm(store.distanceKm) }}</span>
                        </a>
                      </td>
                      <td>
                        <a
                          :href="supermarketDirectionsUrl(store)"
                          target="_blank"
                          rel="noreferrer"
                        >
                          {{ store.hours || '-' }}
                        </a>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
              <p
                v-else-if="bestDealNearbyStatus === 'ready' && !bestDealNearbyMessage"
                class="supermarket-detail__address-note"
              >
                暫時沒有最優惠商戶的可定位門店資料。
              </p>
            </section>
          </section>

          <section class="supermarket-detail__trend-layout">
            <aside class="supermarket-detail__summary-panel">
              <h3>90日價格摘要</h3>
              <div>
                <article>
                  <span>90日最低原價</span>
                  <strong>{{ formatHKPrice(detail.summary.lowestList) }}</strong>
                </article>
                <article>
                  <span>90日最高原價</span>
                  <strong>{{ formatHKPrice(detail.summary.highestList) }}</strong>
                </article>
                <article>
                  <span>90日最低優惠價</span>
                  <strong>{{ formatHKPrice(detail.summary.lowestDeal) }}</strong>
                </article>
                <article>
                  <span>90日最高優惠價</span>
                  <strong>{{ formatHKPrice(detail.summary.highestDeal) }}</strong>
                </article>
              </div>
            </aside>

            <article class="supermarket-detail__chart-panel">
              <div class="supermarket-detail__chart-header">
                <div>
                  <h3>90日超市價格走勢</h3>
                  <p>每條線代表一間超市，可切換原價與優惠後等效單件價。</p>
                </div>
                <div class="supermarket-detail__mode-toggle">
                  <button
                    type="button"
                    :class="detailChartMode === 'effective' ? 'supermarket-detail__mode-toggle-active' : ''"
                    @click="detailChartMode = 'effective'"
                  >
                    優惠價
                  </button>
                  <button
                    type="button"
                    :class="detailChartMode === 'list' ? 'supermarket-detail__mode-toggle-active' : ''"
                    @click="detailChartMode = 'list'"
                  >
                    原價
                  </button>
                </div>
              </div>

              <div
                v-if="detailTrendChart.hasData"
                class="supermarket-detail__chart"
              >
                <svg
                  :viewBox="`0 0 ${detailTrendChart.width} ${detailTrendChart.height}`"
                  preserveAspectRatio="none"
                  role="img"
                  aria-label="90日超市價格走勢"
                  @mousemove="handleTrendHover"
                  @mouseleave="clearTrendHover"
                >
                  <rect
                    x="0"
                    y="0"
                    :width="detailTrendChart.width"
                    :height="detailTrendChart.height"
                  />
                  <g
                    v-for="tick in detailTrendChart.yTicks"
                    :key="tick.key"
                  >
                    <line
                      :x1="detailTrendChart.padding.left"
                      :x2="detailTrendChart.width - detailTrendChart.padding.right"
                      :y1="tick.y"
                      :y2="tick.y"
                    />
                    <text
                      :x="detailTrendChart.padding.left - 8"
                      :y="tick.y + 4"
                      text-anchor="end"
                    >
                      {{ tick.label }}
                    </text>
                  </g>
                  <g
                    v-for="tick in detailTrendChart.xTicks"
                    :key="tick.key"
                  >
                    <line
                      :x1="tick.x"
                      :x2="tick.x"
                      :y1="detailTrendChart.padding.top"
                      :y2="detailTrendChart.height - detailTrendChart.padding.bottom"
                    />
                    <text
                      :x="tick.x"
                      :y="detailTrendChart.height - 16"
                      text-anchor="middle"
                    >
                      {{ tick.label }}
                    </text>
                  </g>
                  <line
                    v-if="detailTrendHover"
                    class="supermarket-detail__hover-line"
                    :x1="detailTrendHover.x"
                    :x2="detailTrendHover.x"
                    :y1="detailTrendChart.padding.top"
                    :y2="detailTrendChart.height - detailTrendChart.padding.bottom"
                  />
                  <polyline
                    v-for="series in detailTrendChart.series"
                    :key="series.store"
                    :points="series.points"
                    :stroke="series.color"
                  />
                  <template
                    v-for="series in detailTrendChart.series"
                    :key="`${series.store}-dots`"
                  >
                    <circle
                      v-for="point in series.dots"
                      :key="point.key"
                      :cx="point.x"
                      :cy="point.y"
                      r="3.2"
                      :fill="series.color"
                    />
                  </template>
                  <circle
                    v-for="point in detailTrendHover?.items ?? []"
                    :key="`${point.key}-hover`"
                    class="supermarket-detail__hover-dot"
                    :cx="point.x"
                    :cy="point.y"
                    :fill="point.color"
                  />
                </svg>
                <div
                  v-if="detailTrendHover"
                  class="supermarket-detail__chart-tooltip"
                  :style="{ left: detailTrendHover.tooltipLeft }"
                >
                  <div class="supermarket-detail__chart-tooltip-head">
                    <strong>{{ formatDateTick(detailTrendHover.date) }}</strong>
                    <span>{{ detailChartMode === 'effective' ? '優惠後等效價' : '原價' }}</span>
                  </div>
                  <ul>
                    <li
                      v-for="item in detailTrendHover.items"
                      :key="`${item.key}-tooltip`"
                    >
                      <i :style="{ background: item.color }" />
                      <span>
                        <strong>{{ displayStore(item.store) }}</strong>
                        <small v-if="item.offer">{{ item.offer }}</small>
                      </span>
                      <em>{{ formatHKPrice(item.value) }}</em>
                    </li>
                  </ul>
                </div>
                <div class="supermarket-detail__legend">
                  <span
                    v-for="series in detailTrendChart.series"
                    :key="`${series.store}-legend`"
                  >
                    <i :style="{ background: series.color }" />
                    {{ displayStore(series.store) }}
                  </span>
                </div>
              </div>
              <p
                v-else
                class="supermarket-state"
              >
                未有歷史資料。
              </p>
            </article>
          </section>

          <section
            v-if="alertFormOpen"
            class="supermarket-detail__panel"
          >
            <div class="supermarket-detail__panel-title">
              <h3>到價 / 優惠提醒</h3>
              <button
                type="button"
                @click="alertFormOpen = false"
              >
                關閉
              </button>
            </div>
            <form
              class="supermarket-alert-form"
              @submit.prevent="submitAlert"
            >
              <input
                v-model="alertForm.targetPrice"
                type="number"
                min="0"
                step="0.1"
                placeholder="目標價格，可留空"
              >
              <select v-model="alertForm.priceMode">
                <option value="effective">優惠後等效價</option>
                <option value="list">原價</option>
              </select>
              <label>
                <input
                  v-model="alertForm.offerRequired"
                  type="checkbox"
                >
                只在有優惠時提醒
              </label>
              <label>
                <input
                  v-model="alertForm.enabled"
                  type="checkbox"
                >
                啟用提示
              </label>
              <button
                type="submit"
                :disabled="savingAlert"
              >
                {{ savingAlert ? '儲存中' : '儲存提示' }}
              </button>
            </form>
          </section>

          <section class="supermarket-detail__related">
            <article
              v-if="detail.sameBrand.length > 0"
              class="supermarket-detail__related-panel"
            >
              <div class="supermarket-detail__panel-title">
                <h3>同品牌產品</h3>
                <span>最多 6 個</span>
              </div>
              <div class="supermarket-detail__related-list">
                <button
                  v-for="product in detail.sameBrand.slice(0, 6)"
                  :key="product.code"
                  type="button"
                  @click="openProduct(product)"
                >
                  <span>{{ product.brand || '未提供品牌' }}</span>
                  <strong>{{ product.name }}</strong>
                  <em>{{ formatHKPrice(primaryPrice(product).effectiveUnitPrice) }}</em>
                </button>
              </div>
            </article>

            <article
              v-if="detail.sameCategory.length > 0"
              class="supermarket-detail__related-panel"
            >
              <div class="supermarket-detail__panel-title">
                <h3>同分類產品</h3>
                <span>最多 6 個</span>
              </div>
              <div class="supermarket-detail__related-list">
                <button
                  v-for="product in detail.sameCategory.slice(0, 6)"
                  :key="product.code"
                  type="button"
                  @click="openProduct(product)"
                >
                  <span>{{ product.brand || '未提供品牌' }}</span>
                  <strong>{{ product.name }}</strong>
                  <em>{{ formatHKPrice(primaryPrice(product).effectiveUnitPrice) }}</em>
                </button>
              </div>
            </article>
          </section>
        </template>
      </section>
    </section>
  </main>
</template>

<style scoped>
.supermarket-page {
  display: grid;
  width: 100%;
  max-width: var(--layout-page-max-width);
  gap: 1.25rem;
  margin: 0 auto;
  padding: 1rem var(--layout-page-padding-inline) 5rem;
  color: rgb(var(--color-text));
}

.supermarket-page__sidebar,
.supermarket-section,
.supermarket-search-main,
.supermarket-filter-panel,
.supermarket-detail {
  border: 1px solid rgb(var(--color-border) / 0.34);
  border-radius: 8px;
  background: rgb(var(--color-topbar-surface) / 0.78);
  box-shadow: 0 16px 40px rgb(15 23 42 / 0.08);
  backdrop-filter: blur(22px) saturate(164%);
}

.supermarket-page__sidebar {
  display: grid;
  align-content: start;
  gap: 1.4rem;
  padding: 1rem;
}

.supermarket-page__kicker,
.supermarket-section__title p,
.supermarket-search-header p {
  margin: 0;
  color: rgb(var(--color-text-muted));
  font-size: 0.74rem;
  font-weight: 800;
  letter-spacing: 0;
  text-transform: uppercase;
}

.supermarket-page__sidebar h1,
.supermarket-section__title h2,
.supermarket-search-header h2 {
  margin: 0.55rem 0 0;
  color: rgb(var(--color-primary));
  font-family: var(--font-display);
  font-size: 1.55rem;
  font-weight: 600;
  line-height: 1.2;
}

.supermarket-page__description {
  margin: 0.7rem 0 0;
  color: rgb(var(--color-text-muted));
  font-size: 0.9rem;
  line-height: 1.55;
}

.supermarket-page__nav {
  display: grid;
  gap: 0.45rem;
}

.supermarket-page__nav-item {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  border-radius: 8px;
  padding: 0.78rem 0.85rem;
  color: rgb(var(--color-text-muted));
  font-size: 0.9rem;
  font-weight: 800;
  text-align: left;
  transition: 0.2s ease;
}

.supermarket-page__nav-item--active,
.supermarket-page__nav-item:hover {
  background: rgb(var(--color-primary) / 0.1);
  color: rgb(var(--color-primary));
}

.supermarket-page__sidebar-stats {
  display: grid;
  gap: 0.35rem;
  border-top: 1px solid rgb(var(--color-border) / 0.46);
  padding-top: 1rem;
  color: rgb(var(--color-text-muted));
  font-size: 0.78rem;
  font-weight: 800;
}

.supermarket-page__updated-at {
  border-top: 1px solid rgb(var(--color-border) / 0.46);
  margin-top: 0.35rem;
  padding-top: 0.65rem;
  color: rgb(var(--color-primary));
  line-height: 1.45;
}

.supermarket-page__content,
.supermarket-search-main {
  min-width: 0;
}

.supermarket-home,
.supermarket-section,
.supermarket-detail {
  display: grid;
  gap: 1rem;
}

.supermarket-home__search {
  display: grid;
  justify-items: center;
  gap: 1rem;
  padding: 1rem 0 0.25rem;
  text-align: center;
}

.supermarket-home__search-form {
  position: relative;
  display: flex;
  align-items: center;
  width: min(100%, 42rem);
  min-height: 3.4rem;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
  box-shadow: 0 12px 30px rgb(15 23 42 / 0.08);
}

.supermarket-home__search-form > .app-icon {
  margin-left: 1rem;
  color: rgb(var(--color-primary));
  flex: 0 0 auto;
}

.supermarket-home__search-form input {
  min-width: 0;
  flex: 1;
  border: 0;
  background: transparent;
  padding: 0 1rem;
  color: rgb(var(--color-text));
  font-size: 1rem;
  outline: none;
}

.supermarket-home__search-form > button,
.supermarket-search-controls__input button,
.supermarket-login button,
.supermarket-alert-form button {
  border-radius: 8px;
  background: rgb(var(--color-primary));
  color: white;
  font-size: 0.84rem;
  font-weight: 900;
  padding: 0.76rem 1.05rem;
}

.supermarket-home__search-form > button {
  margin-right: 0.45rem;
}

.supermarket-home__suggestions {
  position: absolute;
  z-index: 20;
  top: calc(100% + 0.45rem);
  right: 0;
  left: 0;
  overflow: hidden;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
  box-shadow: 0 18px 48px rgb(15 23 42 / 0.16);
  text-align: left;
}

.supermarket-home__suggestions button {
  display: flex;
  width: 100%;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  border-bottom: 1px solid rgb(var(--color-border) / 0.46);
  padding: 0.85rem 1rem;
  text-align: left;
}

.supermarket-home__suggestions button:last-child {
  border-bottom: 0;
}

.supermarket-home__suggestions strong,
.supermarket-home__suggestions small {
  display: block;
}

.supermarket-home__suggestions strong {
  overflow: hidden;
  color: rgb(var(--color-text));
  font-size: 0.9rem;
  font-weight: 900;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.supermarket-home__suggestions small,
.supermarket-home__suggestions em {
  color: rgb(var(--color-text-muted));
  font-size: 0.78rem;
  font-style: normal;
  font-weight: 800;
}

.supermarket-hot-categories {
  display: flex;
  max-width: 62rem;
  flex-wrap: wrap;
  align-items: center;
  justify-content: center;
  gap: 0.45rem;
}

.supermarket-hot-categories > span {
  color: rgb(var(--color-text-muted));
  font-size: 0.78rem;
  font-weight: 900;
}

.supermarket-hot-categories button {
  border: 1px solid rgb(var(--color-border));
  border-radius: 999px;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text-muted));
  font-size: 0.78rem;
  font-weight: 900;
  padding: 0.45rem 0.75rem;
}

.supermarket-stats {
  display: grid;
  gap: 0.8rem;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.supermarket-stats article {
  display: flex;
  align-items: center;
  gap: 0.85rem;
  border: 1px solid rgb(var(--color-border) / 0.42);
  border-radius: 8px;
  background: rgb(var(--color-surface));
  padding: 1rem;
}

.supermarket-stats__icon {
  display: inline-flex;
  width: 2.45rem;
  height: 2.45rem;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  background: rgb(var(--color-primary) / 0.1);
  color: rgb(var(--color-primary));
  font-size: 0.82rem;
  font-weight: 900;
}

.supermarket-stats__deal {
  border-left: 4px solid rgb(var(--color-primary));
}

.supermarket-stats strong {
  display: block;
  color: rgb(var(--color-text));
  font-family: var(--font-display);
  font-size: 1.45rem;
  font-weight: 700;
  line-height: 1;
}

.supermarket-stats p {
  margin: 0.35rem 0 0;
  color: rgb(var(--color-text-muted));
  font-size: 0.78rem;
  font-weight: 800;
}

.supermarket-section,
.supermarket-detail {
  padding: 1rem;
}

.supermarket-section__title,
.supermarket-section__actions,
.supermarket-pagination,
.supermarket-detail__panel-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
}

.supermarket-section__actions {
  flex-wrap: wrap;
  justify-content: flex-end;
}

.supermarket-link-button,
.supermarket-filter-panel button,
.supermarket-pagination button,
.supermarket-table-action {
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  color: rgb(var(--color-primary));
  font-size: 0.82rem;
  font-weight: 900;
  padding: 0.62rem 0.85rem;
}

.supermarket-layout-toggle {
  display: inline-flex;
  overflow: hidden;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
}

.supermarket-layout-toggle button {
  min-height: 2.35rem;
  padding: 0 0.85rem;
  color: rgb(var(--color-text-muted));
  font-size: 0.78rem;
  font-weight: 900;
}

.supermarket-layout-toggle__active {
  background: rgb(var(--color-primary));
  color: white !important;
}

.supermarket-layout-toggle--small button {
  min-height: 2rem;
  padding: 0 0.65rem;
}

.supermarket-state,
.supermarket-login,
.supermarket-section__note {
  border: 1px dashed rgb(var(--color-border));
  border-radius: 8px;
  margin: 0;
  padding: 1rem;
  color: rgb(var(--color-text-muted));
  font-size: 0.9rem;
  line-height: 1.6;
}

.supermarket-state--error {
  border-color: rgb(var(--color-danger) / 0.35);
  color: rgb(var(--color-danger));
}

.supermarket-product-table {
  overflow-x: auto;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
}

.supermarket-product-table table {
  width: 100%;
  min-width: 58rem;
  border-collapse: collapse;
  text-align: left;
}

.supermarket-product-table th {
  border-bottom: 1px solid rgb(var(--color-border));
  background: rgb(var(--color-primary) / 0.07);
  color: rgb(var(--color-text-muted));
  font-size: 0.72rem;
  font-weight: 900;
  letter-spacing: 0;
  padding: 0.8rem;
  text-transform: uppercase;
}

.supermarket-product-table td {
  border-bottom: 1px solid rgb(var(--color-border) / 0.52);
  color: rgb(var(--color-text-muted));
  font-size: 0.82rem;
  font-weight: 800;
  padding: 0.85rem;
  vertical-align: middle;
}

.supermarket-product-table__stores {
  width: 7.5rem;
  min-width: 7.5rem;
  max-width: 7.5rem;
}

.supermarket-product-table tbody tr {
  cursor: pointer;
}

.supermarket-product-table tbody tr:hover {
  background: rgb(var(--color-primary) / 0.06);
}

.supermarket-product-table tbody tr:last-child td {
  border-bottom: 0;
}

.supermarket-product-table td > strong,
.supermarket-product-table td > span {
  display: block;
}

.supermarket-product-table td:first-child strong {
  display: -webkit-box;
  overflow: hidden;
  max-width: 16rem;
  color: rgb(var(--color-text));
  font-size: 0.9rem;
  font-weight: 900;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.supermarket-product-table td:nth-child(2) strong {
  color: rgb(var(--color-primary));
  font-family: var(--font-display);
  font-size: 0.98rem;
  font-weight: 700;
  white-space: nowrap;
}

.supermarket-product-table em {
  display: -webkit-box;
  overflow: hidden;
  max-width: 16rem;
  border-radius: 8px;
  background: rgb(var(--color-primary) / 0.1);
  color: rgb(var(--color-primary));
  font-size: 0.74rem;
  font-style: normal;
  font-weight: 900;
  line-height: 1.45;
  padding: 0.35rem 0.5rem;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.supermarket-discount {
  color: rgb(var(--color-danger));
  font-family: var(--font-display);
  font-size: 0.95rem;
}

.supermarket-store-chips {
  display: flex !important;
  width: 6.6rem;
  flex-direction: column;
  align-items: stretch;
  gap: 0.3rem;
}

.supermarket-store-chips span {
  display: block;
  overflow: hidden;
  width: 100%;
  border-radius: 999px;
  background: rgb(var(--color-primary) / 0.08);
  color: rgb(var(--color-primary));
  font-size: 0.72rem;
  font-weight: 900;
  padding: 0.22rem 0.5rem;
  text-align: center;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.supermarket-product-grid {
  display: grid;
  gap: 1rem;
  grid-template-columns: repeat(auto-fit, minmax(17rem, 1fr));
}

.supermarket-product-card {
  display: grid;
  gap: 0.75rem;
  border: 1px solid rgb(var(--color-border) / 0.42);
  border-radius: 8px;
  background: rgb(var(--color-surface));
  padding: 1rem;
}

.supermarket-product-card > button {
  display: grid;
  gap: 0.7rem;
  min-width: 0;
  text-align: left;
}

.supermarket-product-card > button > span,
.supermarket-product-card small {
  color: rgb(var(--color-text-muted));
  font-size: 0.8rem;
  font-weight: 900;
}

.supermarket-product-card > button > strong {
  display: -webkit-box;
  overflow: hidden;
  min-height: 2.5rem;
  color: rgb(var(--color-text));
  font-size: 0.98rem;
  font-weight: 900;
  line-height: 1.28;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.supermarket-store-list {
  display: grid;
  gap: 0.55rem;
}

.supermarket-store-list article {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 0.8rem;
  border: 1px solid rgb(var(--color-border) / 0.48);
  border-radius: 8px;
  background: rgb(var(--color-primary) / 0.05);
  padding: 0.65rem;
}

.supermarket-store-list strong,
.supermarket-store-list span {
  display: block;
}

.supermarket-store-list div:first-child strong {
  color: rgb(var(--color-text));
  font-size: 0.78rem;
  font-weight: 900;
}

.supermarket-store-list div:first-child span {
  display: -webkit-box;
  overflow: hidden;
  margin-top: 0.2rem;
  color: rgb(var(--color-primary));
  font-size: 0.72rem;
  font-weight: 900;
  line-height: 1.35;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.supermarket-store-list div:last-child {
  text-align: right;
}

.supermarket-store-list div:last-child strong {
  color: rgb(var(--color-primary));
  font-family: var(--font-display);
  font-size: 0.82rem;
  font-weight: 700;
  white-space: nowrap;
}

.supermarket-store-list div:last-child span {
  color: rgb(var(--color-danger));
  font-size: 0.72rem;
  font-weight: 900;
  white-space: nowrap;
}

.supermarket-search-shell {
  display: grid;
  gap: 1rem;
}

.supermarket-filter-panel {
  display: grid;
  align-content: start;
  gap: 1rem;
  padding: 1rem;
}

.supermarket-filter-panel__card {
  display: grid;
  gap: 0.9rem;
  border: 1px solid rgb(var(--color-border) / 0.62);
  border-radius: 8px;
  background: rgb(var(--color-surface));
  padding: 1rem;
  box-shadow: 0 10px 28px rgb(15 23 42 / 0.06);
}

.supermarket-filter-panel h2 {
  margin: 0;
  color: rgb(var(--color-text));
  font-size: 1.05rem;
  font-weight: 900;
}

.supermarket-filter-panel p {
  margin: 0.25rem 0 0;
  color: rgb(var(--color-text-muted));
  font-size: 0.74rem;
  font-weight: 900;
  letter-spacing: 0;
  text-transform: uppercase;
}

.supermarket-filter-panel label {
  display: grid;
  gap: 0.45rem;
}

.supermarket-filter-panel label > span {
  color: rgb(var(--color-text-muted));
  font-size: 0.78rem;
  font-weight: 900;
}

.supermarket-filter-panel input[type='search'],
.supermarket-alert-form input,
.supermarket-alert-form select {
  min-height: 2.65rem;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
  padding: 0 0.8rem;
  color: rgb(var(--color-text));
  font-size: 0.88rem;
}

.supermarket-control-select {
  width: 100%;
  min-height: 2.65rem;
  appearance: none;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background-color: rgb(var(--color-surface));
  background-image:
    linear-gradient(45deg, transparent 50%, rgb(var(--color-text-muted)) 50%),
    linear-gradient(135deg, rgb(var(--color-text-muted)) 50%, transparent 50%);
  background-position:
    calc(100% - 18px) calc(50% - 2px),
    calc(100% - 12px) calc(50% - 2px);
  background-repeat: no-repeat;
  background-size:
    6px 6px,
    6px 6px;
  color: rgb(var(--color-text));
  cursor: pointer;
  font-size: 0.88rem;
  font-weight: 800;
  outline: none;
  padding: 0 2.4rem 0 0.8rem;
  transition:
    border-color 0.18s ease,
    box-shadow 0.18s ease;
}

.supermarket-control-select:focus {
  border-color: rgb(var(--color-primary));
  box-shadow: 0 0 0 3px rgb(var(--color-primary) / 0.14);
}

.supermarket-filter-panel__checkbox {
  display: flex !important;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  border: 1px solid rgb(var(--color-border) / 0.52);
  border-radius: 8px;
  background: rgb(var(--color-primary) / 0.06);
  padding: 0.75rem;
}

.supermarket-filter-panel__checkbox input {
  width: 1.1rem;
  height: 1.1rem;
  accent-color: rgb(var(--color-primary));
}

.supermarket-search-main {
  display: grid;
  align-content: start;
  overflow: hidden;
}

.supermarket-search-header,
.supermarket-search-controls,
.supermarket-stats-strip,
.supermarket-search-main > .supermarket-product-table,
.supermarket-search-main > .supermarket-product-grid,
.supermarket-search-main > .supermarket-state,
.supermarket-search-main > .supermarket-pagination {
  margin: 1rem;
}

.supermarket-search-header {
  display: flex;
  align-items: start;
  justify-content: space-between;
  gap: 1rem;
}

.supermarket-search-header > span {
  border-radius: 8px;
  background: rgb(var(--color-primary) / 0.08);
  color: rgb(var(--color-text-muted));
  font-size: 0.74rem;
  font-weight: 900;
  padding: 0.42rem 0.6rem;
  white-space: nowrap;
}

.supermarket-search-controls {
  display: grid;
  gap: 0.85rem;
}

.supermarket-search-controls__input {
  display: flex;
  align-items: center;
  min-height: 2.85rem;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
}

.supermarket-search-controls__input > .app-icon {
  margin-left: 0.85rem;
  color: rgb(var(--color-primary));
  flex: 0 0 auto;
}

.supermarket-search-controls__input input {
  min-width: 0;
  flex: 1;
  border: 0;
  background: transparent;
  padding: 0 0.75rem;
  color: rgb(var(--color-text));
  outline: none;
}

.supermarket-search-controls__input button {
  margin-right: 0.35rem;
  padding-block: 0.58rem;
}

.supermarket-search-controls__meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  color: rgb(var(--color-text-muted));
  font-size: 0.78rem;
  font-weight: 900;
}

.supermarket-search-controls__meta label {
  display: flex;
  align-items: center;
  gap: 0.45rem;
}

.supermarket-search-controls__sort {
  color: rgb(var(--color-text-muted));
  font-size: 0.78rem;
  font-weight: 900;
}

.supermarket-search-controls__sort .supermarket-control-select {
  width: auto;
  min-width: 9.5rem;
  min-height: 2.35rem;
  color: rgb(var(--color-text));
  font-size: 0.82rem;
}

.supermarket-stats-strip {
  display: grid;
  overflow: hidden;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-primary) / 0.05);
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.supermarket-stats-strip article {
  border-right: 1px solid rgb(var(--color-border));
  padding: 0.65rem 0.75rem;
}

.supermarket-stats-strip article:last-child {
  border-right: 0;
}

.supermarket-stats-strip span {
  display: block;
  color: rgb(var(--color-text-muted));
  font-size: 0.68rem;
  font-weight: 900;
}

.supermarket-stats-strip strong {
  display: block;
  margin-top: 0.16rem;
  color: rgb(var(--color-primary));
  font-size: 0.95rem;
  font-weight: 900;
}

.supermarket-pagination {
  justify-content: center;
  color: rgb(var(--color-text-muted));
  font-size: 0.82rem;
  font-weight: 900;
}

.supermarket-pagination button:disabled {
  opacity: 0.45;
}

.supermarket-login {
  display: grid;
  justify-items: start;
  gap: 0.7rem;
}

.supermarket-login h3,
.supermarket-login p {
  margin: 0;
}

.supermarket-product-card footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.55rem;
  border-top: 1px solid rgb(var(--color-border) / 0.48);
  padding-top: 0.75rem;
}

.supermarket-product-card footer button {
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  color: rgb(var(--color-primary));
  font-size: 0.8rem;
  font-weight: 900;
  padding: 0.5rem 0.75rem;
}

.supermarket-data-grid {
  display: grid;
  gap: 1rem;
}

.supermarket-data-grid > article {
  border: 1px solid rgb(var(--color-border) / 0.42);
  border-radius: 8px;
  background: rgb(var(--color-surface));
  padding: 1rem;
}

.supermarket-data-grid h3,
.supermarket-data-source h3 {
  margin: 0 0 0.8rem;
  color: rgb(var(--color-text));
  font-size: 1rem;
  font-weight: 900;
}

.supermarket-data-grid p,
.supermarket-data-source p {
  margin: 1rem 0 0;
  color: rgb(var(--color-text-muted));
  font-size: 0.88rem;
  line-height: 1.65;
}

.supermarket-data-freshness {
  display: grid;
  gap: 0.7rem;
  margin-bottom: 1rem;
}

.supermarket-data-freshness div,
.supermarket-data-source {
  border: 1px solid rgb(var(--color-border) / 0.42);
  border-radius: 8px;
  background: rgb(var(--color-surface));
}

.supermarket-data-freshness div {
  padding: 0.9rem;
}

.supermarket-data-freshness span {
  display: block;
  color: rgb(var(--color-text-muted));
  font-size: 0.78rem;
  font-weight: 900;
}

.supermarket-data-freshness strong {
  display: block;
  margin-top: 0.24rem;
  color: rgb(var(--color-primary));
  font-family: var(--font-display);
  font-size: 1.05rem;
  font-weight: 700;
}

.supermarket-data-stats {
  display: grid;
  overflow: hidden;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.supermarket-data-stats div {
  border-right: 1px solid rgb(var(--color-border));
  border-bottom: 1px solid rgb(var(--color-border));
  padding: 0.8rem;
}

.supermarket-data-stats div:nth-child(2n) {
  border-right: 0;
}

.supermarket-data-stats div:nth-last-child(-n + 2) {
  border-bottom: 0;
}

.supermarket-data-stats span,
.supermarket-fields span {
  color: rgb(var(--color-text-muted));
  font-size: 0.78rem;
  font-weight: 800;
}

.supermarket-data-stats strong {
  display: block;
  margin-top: 0.24rem;
  color: rgb(var(--color-primary));
  font-size: 1.35rem;
  font-weight: 900;
}

.supermarket-fields {
  overflow: hidden;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
}

.supermarket-fields div {
  display: grid;
  grid-template-columns: 7rem minmax(0, 1fr);
  border-bottom: 1px solid rgb(var(--color-border));
}

.supermarket-fields div:last-child {
  border-bottom: 0;
}

.supermarket-fields strong {
  background: rgb(var(--color-primary) / 0.07);
  color: rgb(var(--color-primary));
  font-size: 0.78rem;
  font-weight: 900;
  padding: 0.65rem 0.75rem;
}

.supermarket-fields span {
  padding: 0.65rem 0.75rem;
}

.supermarket-data-source {
  padding: 1rem;
}

.supermarket-source-links {
  display: grid;
  gap: 0.6rem;
}

.supermarket-source-links a {
  min-width: 0;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
  padding: 0.75rem;
  transition: 0.2s ease;
}

.supermarket-source-links a:hover {
  border-color: rgb(var(--color-primary) / 0.42);
  background: rgb(var(--color-primary) / 0.06);
}

.supermarket-source-links strong,
.supermarket-source-links span {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.supermarket-source-links strong {
  color: rgb(var(--color-text));
  font-size: 0.86rem;
  font-weight: 900;
}

.supermarket-source-links span {
  margin-top: 0.22rem;
  color: rgb(var(--color-text-muted));
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, 'Liberation Mono', 'Courier New', monospace;
  font-size: 0.72rem;
  font-weight: 700;
}

.supermarket-detail__hero,
.supermarket-detail__store-section,
.supermarket-detail__summary-panel,
.supermarket-detail__chart-panel,
.supermarket-detail__panel,
.supermarket-detail__related-panel {
  border: 1px solid rgb(var(--color-border) / 0.42);
  border-radius: 8px;
  background: rgb(var(--color-surface));
}

.supermarket-detail__hero {
  display: grid;
  gap: 1rem;
  padding: 1.15rem;
}

.supermarket-detail__hero p,
.supermarket-detail__hero span,
.supermarket-detail__panel-title span,
.supermarket-detail__chart-header p,
.supermarket-detail__best-panel p,
.supermarket-detail__address-note {
  color: rgb(var(--color-text-muted));
  font-size: 0.82rem;
  font-weight: 800;
}

.supermarket-detail__hero p,
.supermarket-detail__hero span {
  display: block;
  margin: 0;
}

.supermarket-detail__hero h1 {
  margin: 0.45rem 0;
  color: rgb(var(--color-text));
  font-size: 1.55rem;
  font-weight: 900;
  line-height: 1.18;
}

.supermarket-detail__actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-start;
  gap: 0.65rem;
}

.supermarket-detail__actions button,
.supermarket-detail__mode-toggle button,
.supermarket-detail__best-header button {
  border-radius: 8px;
  font-size: 0.84rem;
  font-weight: 900;
}

.supermarket-detail__actions button {
  background: rgb(var(--color-primary));
  color: white;
  padding: 0.72rem 1rem;
}

.supermarket-detail__actions button:first-child {
  border: 1px solid rgb(var(--color-primary));
  background: transparent;
  color: rgb(var(--color-primary));
}

.supermarket-detail__actions > span {
  flex-basis: 100%;
  color: rgb(var(--color-text-muted));
  font-size: 0.78rem;
  font-weight: 800;
}

.supermarket-detail__actions button:disabled,
.supermarket-detail__best-header button:disabled,
.supermarket-alert-form button:disabled {
  opacity: 0.6;
}

.supermarket-detail__store-section,
.supermarket-detail__summary-panel,
.supermarket-detail__chart-panel,
.supermarket-detail__panel,
.supermarket-detail__related-panel {
  display: grid;
  gap: 1rem;
  padding: 1rem;
}

.supermarket-detail__panel-title h3,
.supermarket-detail__discount-list h4,
.supermarket-detail__best-panel h4,
.supermarket-detail__summary-panel h3,
.supermarket-detail__chart-header h3 {
  margin: 0;
  color: rgb(var(--color-text));
  font-size: 1rem;
  font-weight: 900;
}

.supermarket-detail__panel-title p,
.supermarket-detail__chart-header p,
.supermarket-detail__best-panel p {
  margin: 0.3rem 0 0;
  line-height: 1.55;
}

.supermarket-detail__panel-title button {
  color: rgb(var(--color-primary));
  font-size: 0.82rem;
  font-weight: 900;
}

.supermarket-detail__discount-list {
  display: grid;
  gap: 0.75rem;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-primary) / 0.04);
  padding: 0.9rem;
}

.supermarket-detail__discount-grid {
  display: grid;
  gap: 0.75rem;
}

.supermarket-detail__discount-grid article,
.supermarket-detail__best-list article,
.supermarket-detail__summary-panel article {
  border: 1px solid rgb(var(--color-border) / 0.46);
  border-radius: 8px;
  background: rgb(var(--color-surface));
  padding: 0.85rem;
}

.supermarket-detail__discount-grid strong,
.supermarket-detail__discount-grid span,
.supermarket-detail__discount-grid small,
.supermarket-detail__discount-grid em,
.supermarket-detail__summary-panel span,
.supermarket-detail__summary-panel strong {
  display: block;
}

.supermarket-detail__discount-grid strong,
.supermarket-detail__best-list strong {
  color: rgb(var(--color-text));
  font-size: 0.88rem;
  font-weight: 900;
}

.supermarket-detail__discount-grid span,
.supermarket-detail__best-list div:nth-child(2) strong,
.supermarket-detail__summary-panel strong {
  margin-top: 0.35rem;
  color: rgb(var(--color-primary));
  font-family: var(--font-display);
  font-size: 1rem;
  font-weight: 700;
}

.supermarket-detail__discount-grid small,
.supermarket-detail__best-list span {
  color: rgb(var(--color-text-muted));
  font-size: 0.74rem;
  font-weight: 800;
}

.supermarket-detail__discount-grid em,
.supermarket-detail__best-list p {
  display: -webkit-box;
  overflow: hidden;
  border-radius: 8px;
  margin-top: 0.65rem;
  background: rgb(var(--color-primary) / 0.1);
  color: rgb(var(--color-primary));
  font-size: 0.76rem;
  font-style: normal;
  font-weight: 900;
  line-height: 1.45;
  padding: 0.45rem 0.55rem;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.supermarket-detail__best-panel {
  display: grid;
  gap: 0.8rem;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
  padding: 1rem;
}

.supermarket-detail__best-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.9rem;
}

.supermarket-detail__best-header button {
  flex: 0 0 auto;
  background: rgb(var(--color-primary));
  color: white;
  padding: 0.72rem 1rem;
}

.supermarket-detail__best-list {
  display: grid;
  gap: 0.7rem;
}

.supermarket-detail__best-list article {
  display: grid;
  gap: 0.65rem;
}

.supermarket-detail__best-list article > div {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
}

.supermarket-detail__address-note {
  border: 1px dashed rgb(var(--color-border));
  border-radius: 8px;
  margin: 0;
  padding: 0.75rem;
}

.supermarket-detail__location-message {
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-primary) / 0.04);
  margin: 0;
  padding: 0.75rem;
}

.supermarket-detail__location-message--error {
  border-color: rgb(var(--color-danger) / 0.35);
  background: rgb(var(--color-danger) / 0.05);
  color: rgb(var(--color-danger)) !important;
}

.supermarket-detail__nearby-table {
  overflow-x: auto;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
}

.supermarket-detail__nearby-table table {
  width: 100%;
  min-width: 48rem;
  border-collapse: collapse;
  text-align: left;
}

.supermarket-detail__nearby-table th {
  border-bottom: 1px solid rgb(var(--color-border));
  background: rgb(var(--color-primary) / 0.06);
  color: rgb(var(--color-text-muted));
  font-size: 0.74rem;
  font-weight: 900;
  padding: 0.72rem;
}

.supermarket-detail__nearby-table td {
  border-bottom: 1px solid rgb(var(--color-border) / 0.48);
  color: rgb(var(--color-text-muted));
  font-size: 0.8rem;
  font-weight: 800;
  vertical-align: top;
}

.supermarket-detail__nearby-table tbody tr:last-child td {
  border-bottom: 0;
}

.supermarket-detail__nearby-table a {
  display: block;
  height: 100%;
  color: inherit;
  padding: 0.8rem 0.72rem;
}

.supermarket-detail__nearby-table td:nth-child(1),
.supermarket-detail__nearby-table td:nth-child(2),
.supermarket-detail__nearby-table td:nth-child(3) {
  color: rgb(var(--color-text));
  font-weight: 900;
}

.supermarket-detail__nearby-table td:nth-child(5),
.supermarket-detail__nearby-table th:nth-child(5) {
  text-align: right;
  white-space: nowrap;
}

.supermarket-detail__nearby-table td:nth-child(5) span {
  border-radius: 999px;
  background: rgb(var(--color-primary) / 0.1);
  color: rgb(var(--color-primary));
  font-size: 0.74rem;
  font-weight: 900;
  padding: 0.24rem 0.5rem;
}

.supermarket-detail__trend-layout {
  display: grid;
  gap: 1rem;
}

.supermarket-detail__summary-panel > div {
  display: grid;
  gap: 0.75rem;
}

.supermarket-detail__summary-panel span {
  color: rgb(var(--color-text-muted));
  font-size: 0.8rem;
  font-weight: 900;
}

.supermarket-detail__summary-panel strong {
  font-size: 1.28rem;
}

.supermarket-detail__chart-panel {
  min-width: 0;
}

.supermarket-detail__chart-header {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 0.9rem;
}

.supermarket-detail__mode-toggle {
  display: inline-grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.25rem;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-primary) / 0.05);
  padding: 0.25rem;
}

.supermarket-detail__mode-toggle button {
  color: rgb(var(--color-text-muted));
  padding: 0.48rem 0.85rem;
}

.supermarket-detail__mode-toggle-active {
  background: rgb(var(--color-primary));
  color: white !important;
}

.supermarket-detail__chart {
  position: relative;
  display: grid;
  gap: 0.8rem;
  min-height: 20rem;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
  padding: 0.85rem;
}

.supermarket-detail__chart svg {
  display: block;
  width: 100%;
  min-height: 18rem;
  cursor: crosshair;
}

.supermarket-detail__chart rect {
  fill: rgb(var(--color-surface));
}

.supermarket-detail__chart line {
  stroke: rgb(var(--color-border));
  stroke-width: 1.2;
}

.supermarket-detail__chart text {
  fill: rgb(var(--color-text-muted));
  font-size: 0.7rem;
  font-weight: 800;
}

.supermarket-detail__chart polyline {
  fill: none;
  stroke-linecap: round;
  stroke-linejoin: round;
  stroke-width: 2.6;
}

.supermarket-detail__chart circle {
  r: 3.2;
  stroke: rgb(var(--color-surface));
  stroke-width: 1.8;
}

.supermarket-detail__hover-line {
  pointer-events: none;
  stroke: rgb(var(--color-primary)) !important;
  stroke-dasharray: 5 5;
  stroke-width: 1.8 !important;
}

.supermarket-detail__hover-dot {
  pointer-events: none;
  r: 5.4;
  stroke: rgb(var(--color-surface));
  stroke-width: 2.4;
}

.supermarket-detail__chart-tooltip {
  position: absolute;
  z-index: 3;
  top: 1.1rem;
  width: min(18rem, calc(100% - 1.5rem));
  transform: translateX(-50%);
  pointer-events: none;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
  box-shadow: 0 18px 44px rgb(15 23 42 / 0.18);
  padding: 0.75rem;
}

.supermarket-detail__chart-tooltip-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.7rem;
  border-bottom: 1px solid rgb(var(--color-border) / 0.5);
  padding-bottom: 0.55rem;
}

.supermarket-detail__chart-tooltip-head strong,
.supermarket-detail__chart-tooltip li span strong,
.supermarket-detail__chart-tooltip li em {
  color: rgb(var(--color-text));
  font-size: 0.78rem;
  font-weight: 900;
}

.supermarket-detail__chart-tooltip-head span,
.supermarket-detail__chart-tooltip li small {
  color: rgb(var(--color-text-muted));
  font-size: 0.72rem;
  font-weight: 800;
}

.supermarket-detail__chart-tooltip ul {
  display: grid;
  gap: 0.52rem;
  margin: 0.65rem 0 0;
  padding: 0;
}

.supermarket-detail__chart-tooltip li {
  display: grid;
  align-items: center;
  gap: 0.55rem;
  grid-template-columns: auto minmax(0, 1fr) auto;
  list-style: none;
}

.supermarket-detail__chart-tooltip li i {
  width: 0.58rem;
  height: 0.58rem;
  border-radius: 999px;
}

.supermarket-detail__chart-tooltip li span {
  display: grid;
  min-width: 0;
  gap: 0.15rem;
}

.supermarket-detail__chart-tooltip li span strong,
.supermarket-detail__chart-tooltip li small {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.supermarket-detail__chart-tooltip li em {
  color: rgb(var(--color-primary));
  font-family: var(--font-display);
  font-style: normal;
  white-space: nowrap;
}

.supermarket-detail__legend {
  display: flex;
  flex-wrap: wrap;
  gap: 0.6rem 1rem;
}

.supermarket-detail__legend span {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  color: rgb(var(--color-text-muted));
  font-size: 0.76rem;
  font-weight: 900;
}

.supermarket-detail__legend i {
  width: 0.62rem;
  height: 0.62rem;
  border-radius: 999px;
}

.supermarket-alert-form {
  display: grid;
  gap: 0.75rem;
}

.supermarket-alert-form label {
  display: flex;
  align-items: center;
  gap: 0.55rem;
  color: rgb(var(--color-text-muted));
  font-size: 0.86rem;
  font-weight: 900;
}

.supermarket-alert-form input[type='checkbox'] {
  width: 1rem;
  height: 1rem;
  accent-color: rgb(var(--color-primary));
}

.supermarket-detail__related {
  display: grid;
  gap: 1rem;
}

.supermarket-detail__related-list {
  display: flex;
  gap: 0.85rem;
  overflow-x: auto;
  padding-bottom: 0.2rem;
}

.supermarket-detail__related-list button {
  display: grid;
  min-width: 17rem;
  align-items: start;
  justify-content: stretch;
  gap: 0.45rem;
  border: 1px solid rgb(var(--color-border) / 0.42);
  border-radius: 8px;
  background: rgb(var(--color-primary) / 0.04);
  padding: 0.85rem;
  text-align: left;
}

.supermarket-detail__related-list button span {
  color: rgb(var(--color-text-muted));
  font-size: 0.76rem;
  font-weight: 900;
}

.supermarket-detail__related-list button strong {
  display: -webkit-box;
  overflow: hidden;
  color: rgb(var(--color-text));
  font-size: 0.9rem;
  font-weight: 900;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.supermarket-detail__related-list button em {
  color: rgb(var(--color-primary));
  font-family: var(--font-display);
  font-size: 0.88rem;
  font-style: normal;
  font-weight: 700;
}

@media (max-width: 640px) {
  .supermarket-section__title,
  .supermarket-search-header,
  .supermarket-search-controls__meta,
  .supermarket-detail__panel-title,
  .supermarket-detail__chart-header,
  .supermarket-detail__best-header,
  .supermarket-detail__best-list article > div {
    align-items: flex-start;
    flex-direction: column;
  }

  .supermarket-home__search-form {
    align-items: stretch;
    flex-wrap: wrap;
    padding-top: 0.65rem;
  }

  .supermarket-home__search-form input {
    flex-basis: calc(100% - 3.5rem);
    min-height: 2.3rem;
  }

  .supermarket-home__search-form > button {
    width: calc(100% - 0.9rem);
    margin: 0 0.45rem 0.45rem;
  }

  .supermarket-store-list article {
    grid-template-columns: 1fr;
  }

  .supermarket-store-list div:last-child {
    text-align: left;
  }

  .supermarket-stats-strip strong {
    font-size: 0.82rem;
  }

  .supermarket-fields div {
    grid-template-columns: 1fr;
  }
}

@media (min-width: 768px) {
  .supermarket-page {
    grid-template-columns: minmax(14rem, 0.24fr) minmax(0, 1fr);
    padding-top: 1.25rem;
  }

  .supermarket-page__sidebar {
    position: sticky;
    top: 1rem;
    min-height: calc(100vh - 2rem);
  }

  .supermarket-stats {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }

  .supermarket-search-shell {
    grid-template-columns: minmax(14rem, 0.28fr) minmax(0, 1fr);
  }

  .supermarket-filter-panel {
    position: sticky;
    top: 1rem;
    min-height: 32rem;
  }

  .supermarket-search-controls {
    grid-template-columns: minmax(18rem, 1fr) auto;
    align-items: center;
  }

  .supermarket-data-grid,
  .supermarket-detail__trend-layout {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .supermarket-data-freshness {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .supermarket-source-links {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .supermarket-detail__discount-grid {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }

  .supermarket-detail__trend-layout {
    grid-template-columns: minmax(18rem, 0.32fr) minmax(0, 0.68fr);
    align-items: stretch;
  }

  .supermarket-detail__hero {
    grid-template-columns: minmax(0, 1fr) auto;
    align-items: center;
  }

  .supermarket-detail__actions {
    justify-content: flex-end;
  }

  .supermarket-detail__hero h1 {
    font-size: 2rem;
  }
}
</style>

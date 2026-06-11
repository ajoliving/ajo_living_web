<!--
 * 超市優惠頻道頁。
 * 1. 接入 AJO 後端超市優惠搜尋、摘要與商品詳情 API。
 * 2. 提供分頁、網格 / 表格切換、分類與商店篩選。
 * 3. 支援會員收藏與商品詳情獨立路由。
-->
<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import {
  addSupermarketFavorite,
  fetchSupermarketFavorites,
  fetchSupermarketProductDetail,
  fetchSupermarketSummary,
  removeSupermarketFavorite,
  searchSupermarketProducts,
} from '@/httpapis/supermarket-offers';
import type {
  SupermarketDailyStorePrice,
  SupermarketProduct,
  SupermarketProductDetail,
  SupermarketSearchResult,
  SupermarketStorePrice,
  SupermarketSummary,
} from '@/model/supermarket-offers';
import { useSessionStore } from '@/stores/session';
import {
  formatSupermarketDistanceKm,
  nearestSupermarketStoreLocations,
  SUPERMARKET_GEOLOCATION_CACHE_MAX_AGE_MS,
  SUPERMARKET_NEARBY_STORE_LIMIT,
  supermarketDirectionsUrl,
} from './constants/store-locations';
import type { SupermarketNearbyStoreLocation } from './constants/store-locations';

type OfferViewMode = 'grid' | 'table';
type OfferListMode = 'all' | 'favorites';
type PriceChartMode = 'effective' | 'list';
type NearbyStoreStatus = 'idle' | 'requesting' | 'ready' | 'error';

interface OfferFilters {
  q: string;
  category: string;
  store: string;
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
  offer?: string;
}

interface TrendChartSeries {
  store: string;
  color: string;
  points: string;
  dots: TrendChartPoint[];
}

const route = useRoute();
const router = useRouter();
const sessionStore = useSessionStore();
const pageSize = 20;
const summary = ref<SupermarketSummary | null>(null);
const searchResult = ref<SupermarketSearchResult | null>(null);
const favorites = ref<SupermarketProduct[]>([]);
const detail = ref<SupermarketProductDetail | null>(null);
const viewMode = ref<OfferViewMode>('grid');
const listMode = ref<OfferListMode>('all');
const loading = ref(false);
const favoritesLoading = ref(false);
const detailLoading = ref(false);
const favoriteSavingCode = ref('');
const errorMessage = ref('');
const detailErrorMessage = ref('');
const suggestions = ref<SupermarketProduct[]>([]);
const suggestionsOpen = ref(false);
const detailChartMode = ref<PriceChartMode>('effective');
const trendHoverDate = ref('');
const bestDealNearbyStores = ref<SupermarketNearbyStoreLocation[]>([]);
const bestDealNearbyStatus = ref<NearbyStoreStatus>('idle');
const bestDealNearbyMessage = ref('');
const filters = reactive<OfferFilters>({
  q: '',
  category: '',
  store: '',
  sort: 'discount',
  page: 1,
});
let suggestionTimer: number | undefined;

const productCode = computed(() => (typeof route.params.code === 'string' ? route.params.code.trim() : ''));
const isDetailPage = computed(() => Boolean(productCode.value));
const visibleItems = computed(() => {
  if (listMode.value === 'all') return searchResult.value?.items ?? [];
  return filteredFavorites.value;
});
const filteredFavorites = computed(() => {
  const keyword = filters.q.trim().toLowerCase();
  return favorites.value.filter((product) => {
    const matchKeyword = !keyword ||
      product.name.toLowerCase().includes(keyword) ||
      product.brand.toLowerCase().includes(keyword) ||
      categoryText(product).toLowerCase().includes(keyword);
    const matchCategory = !filters.category || [product.category1, product.category2, product.category3].includes(filters.category);
    const matchStore = !filters.store || normalizedStorePrices(product).some((price) => price.store === filters.store);
    return matchKeyword && matchCategory && matchStore;
  }).sort((a, b) => compareProducts(a, b, filters.sort));
});
const visibleTotal = computed(() => (listMode.value === 'favorites' ? filteredFavorites.value.length : searchResult.value?.total ?? 0));
const totalPages = computed(() => {
  if (listMode.value === 'favorites') return 1;
  if (!searchResult.value) return 1;
  return Math.max(1, Math.ceil(searchResult.value.total / searchResult.value.pageSize));
});
const pageNumbers = computed(() => {
  const total = totalPages.value;
  const current = Math.min(total, Math.max(1, filters.page));
  const start = Math.max(1, Math.min(current - 2, total - 4));
  const end = Math.min(total, start + 4);
  return Array.from({ length: end - start + 1 }, (_, index) => start + index);
});
const categoryOptions = computed(() => searchResult.value?.categories ?? summary.value?.categories.map((item) => item.value) ?? []);
const storeOptions = computed(() => searchResult.value?.stores ?? summary.value?.stores.map((item) => item.value) ?? []);
const heroStats = computed(() => summary.value?.stats ?? searchResult.value?.stats ?? null);
const highestDiscount = computed(() => {
  const products = summary.value?.bestDiscounts ?? searchResult.value?.items ?? [];
  const rate = products.reduce((max, product) => Math.max(max, discountRate(primaryPrice(product))), 0);
  return rate > 0 ? `-${rate.toFixed(0)}%` : '-';
});
const latestDateText = computed(() => formatDisplayDate(summary.value?.metadata?.latestSnapshotDate ?? ''));
const detailStorePrices = computed(() => currentStorePrices(detail.value?.stores ?? []));
const detailDiscountPrices = computed(() => detailStorePrices.value.filter((price) => discountRate(price) > 0));
const bestDealPrices = computed(() => {
  const prices = detailDiscountPrices.value.length > 0 ? detailDiscountPrices.value : detailStorePrices.value;
  const primary = prices[0];
  if (!primary) return [];
  const bestRate = discountRate(primary);
  if (bestRate <= 0) return prices.slice(0, 1);
  return prices.filter((price) => Math.abs(discountRate(price) - bestRate) < 0.0001);
});
const bestDealSummary = computed(() =>
  bestDealPrices.value.length > 0
    ? bestDealPrices.value.map((price) => `${displayStore(price.store)} ${formatHKPrice(price.effectiveUnitPrice)}`).join(' / ')
    : '未有可計算的優惠商戶。',
);
const detailRecentHistory = computed(() =>
  [...(detail.value?.history ?? [])]
    .sort((a, b) => b.date.localeCompare(a.date))
    .slice(0, 8),
);
const detailTrendChart = computed(() => buildTrendChart(detail.value?.history ?? [], detailChartMode.value));
const detailTrendHover = computed(() => {
  const chart = detailTrendChart.value;
  if (!trendHoverDate.value || !chart.hasData) return null;
  const items = chart.series
    .flatMap((series) => series.dots
      .filter((point) => point.date === trendHoverDate.value)
      .map((point) => ({ ...point, color: series.color })))
    .sort((a, b) => a.value - b.value);
  if (items.length === 0) return null;
  const x = items[0].x;
  return {
    date: trendHoverDate.value,
    items,
    x,
    tooltipLeft: `${Math.min(88, Math.max(12, (x / chart.width) * 100))}%`,
  };
});

// 1. 載入摘要與首屏列表
const loadInitialData = async (): Promise<void> => {
  await Promise.all([loadSummary(), loadProducts()]);
};

// 2. 載入超市優惠摘要
const loadSummary = async (): Promise<void> => {
  try {
    const { data } = await fetchSupermarketSummary();
    summary.value = data.data;
  } catch {
    summary.value = null;
  }
};

// 3. 載入優惠商品列表
const loadProducts = async (): Promise<void> => {
  loading.value = true;
  errorMessage.value = '';
  try {
    const { data } = await searchSupermarketProducts({
      q: filters.q.trim() || undefined,
      category: filters.category || undefined,
      store: filters.store || undefined,
      offerOnly: true,
      sort: filters.sort,
      page: filters.page,
      pageSize,
    });
    searchResult.value = data.data;
  } catch {
    errorMessage.value = '暫時無法載入優惠資料。';
  } finally {
    loading.value = false;
  }
};

// 4. 載入會員收藏
const loadFavorites = async (): Promise<void> => {
  if (!sessionStore.isAuthenticated) {
    favorites.value = [];
    return;
  }
  favoritesLoading.value = true;
  errorMessage.value = '';
  try {
    const { data } = await fetchSupermarketFavorites({ page: 1, pageSize: 200 });
    favorites.value = data.data.items.map((product) => ({ ...product, isFavorite: true }));
  } catch {
    errorMessage.value = '暫時無法載入收藏。';
  } finally {
    favoritesLoading.value = false;
  }
};

// 5. 載入搜尋提示
const loadSuggestions = async (keyword: string): Promise<void> => {
  if (!keyword) {
    suggestions.value = [];
    return;
  }
  try {
    const { data } = await searchSupermarketProducts({
      q: keyword,
      offerOnly: true,
      sort: 'discount',
      page: 1,
      pageSize: 6,
    });
    if (filters.q.trim() === keyword) {
      suggestions.value = data.data.items;
    }
  } catch {
    suggestions.value = [];
  }
};

// 6. 載入商品詳情
const loadDetail = async (code: string): Promise<void> => {
  detailLoading.value = true;
  detailErrorMessage.value = '';
  try {
    const { data } = await fetchSupermarketProductDetail(code, 90);
    detail.value = data.data;
    trendHoverDate.value = '';
    resetBestDealLocation();
  } catch {
    detailErrorMessage.value = '暫時無法載入商品詳情。';
  } finally {
    detailLoading.value = false;
  }
};

// 7. 重設最優惠門店定位
const resetBestDealLocation = (): void => {
  bestDealNearbyStores.value = [];
  bestDealNearbyStatus.value = 'idle';
  bestDealNearbyMessage.value = '';
};

// 8. 提交搜尋
const submitSearch = (): void => {
  filters.page = 1;
  suggestionsOpen.value = false;
  if (listMode.value === 'favorites') return;
  void loadProducts();
};

// 9. 延遲關閉搜尋提示
const closeSuggestionsSoon = (): void => {
  window.setTimeout(() => {
    suggestionsOpen.value = false;
  }, 120);
};

// 10. 套用分類或商店篩選
const applyFilter = (): void => {
  filters.page = 1;
  if (listMode.value === 'favorites') return;
  void loadProducts();
};

// 11. 切換分頁
const changePage = (page: number): void => {
  if (listMode.value === 'favorites') return;
  filters.page = Math.min(totalPages.value, Math.max(1, page));
  void loadProducts();
};

// 12. 切換列表模式
const setListMode = async (mode: OfferListMode): Promise<void> => {
  if (mode === 'favorites' && !sessionStore.isAuthenticated) {
    await router.push({ path: '/login', query: { redirect: route.fullPath } });
    return;
  }
  listMode.value = mode;
  filters.page = 1;
  if (mode === 'favorites') {
    await loadFavorites();
  } else {
    await loadProducts();
  }
};

// 13. 開啟商品詳情
const openProduct = async (product: SupermarketProduct): Promise<void> => {
  suggestionsOpen.value = false;
  await router.push({ path: `/supermarket-offers/products/${encodeURIComponent(product.code)}` });
};

// 14. 返回優惠列表
const backToList = async (): Promise<void> => {
  detail.value = null;
  await router.push({ path: '/supermarket-offers' });
};

// 15. 切換收藏
const toggleFavorite = async (product: SupermarketProduct): Promise<void> => {
  if (!sessionStore.isAuthenticated) {
    await router.push({ path: '/login', query: { redirect: route.fullPath } });
    return;
  }
  if (favoriteSavingCode.value) return;
  favoriteSavingCode.value = product.code;
  try {
    if (product.isFavorite) {
      await removeSupermarketFavorite(product.code);
      product.isFavorite = false;
      favorites.value = favorites.value.filter((item) => item.code !== product.code);
      if (detail.value?.product.code === product.code) {
        detail.value.isFavorite = false;
        detail.value.product.isFavorite = false;
      }
      return;
    }
    await addSupermarketFavorite(product.code);
    product.isFavorite = true;
    if (!favorites.value.some((item) => item.code === product.code)) {
      favorites.value = [{ ...product, isFavorite: true }, ...favorites.value];
    }
    if (detail.value?.product.code === product.code) {
      detail.value.isFavorite = true;
      detail.value.product.isFavorite = true;
    }
  } finally {
    favoriteSavingCode.value = '';
  }
};

// 16. 標準化商店價格
const normalizedStorePrices = (product: SupermarketProduct): SupermarketStorePrice[] => {
  if (product.storePrices && product.storePrices.length > 0) {
    return [...product.storePrices].sort((a, b) => a.effectiveUnitPrice - b.effectiveUnitPrice);
  }
  return [{
    store: product.bestStore || product.stores[0] || '',
    listPrice: product.listPrice || product.maxPrice || 0,
    effectiveUnitPrice: product.effectiveUnitPrice || product.bestEffectiveUnitPrice || product.minPrice || 0,
    offer: product.bestOffer ?? '',
    parseStatus: product.parseStatus,
    snapshotDate: '',
  }];
};

// 17. 排列詳情商店價格
const currentStorePrices = (stores: SupermarketStorePrice[]): SupermarketStorePrice[] =>
  [...stores].sort((a, b) => {
    if (a.effectiveUnitPrice === b.effectiveUnitPrice) {
      return displayStore(a.store).localeCompare(displayStore(b.store), 'zh-HK');
    }
    return a.effectiveUnitPrice - b.effectiveUnitPrice;
  });

// 18. 取得商品主要價格
const primaryPrice = (product: SupermarketProduct): SupermarketStorePrice =>
  normalizedStorePrices(product)[0] ?? {
    store: product.bestStore,
    listPrice: product.listPrice,
    effectiveUnitPrice: product.effectiveUnitPrice,
    offer: product.bestOffer ?? '',
    parseStatus: product.parseStatus,
    snapshotDate: '',
  };

// 19. 計算優惠力度
const discountRate = (price: SupermarketStorePrice): number =>
  price.listPrice > 0 ? Math.max(0, ((price.listPrice - price.effectiveUnitPrice) / price.listPrice) * 100) : 0;

// 20. 比較商品排序
const compareProducts = (a: SupermarketProduct, b: SupermarketProduct, sort: string): number => {
  if (sort === 'effective') return primaryPrice(a).effectiveUnitPrice - primaryPrice(b).effectiveUnitPrice;
  if (sort === 'diff') return b.priceDiff - a.priceDiff;
  if (sort === 'name') return a.name.localeCompare(b.name, 'zh-HK');
  return discountRate(primaryPrice(b)) - discountRate(primaryPrice(a));
};

// 21. 格式化港幣
const formatHKPrice = (value: number | undefined): string =>
  `HK$${(value ?? 0).toLocaleString('zh-HK', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;

// 22. 顯示商店名稱
const displayStore = (store: string): string => {
  const labels: Record<string, string> = {
    AEON: 'AEON',
    WELLCOME: '惠康',
    PARKNSHOP: '百佳',
    JASONS: 'Market Place',
    MARKETPLACE: 'Market Place',
    CITYSUPER: 'CitySuper',
    WATSONS: '屈臣氏',
    MANNINGS: '萬寧',
    SASA: '莎莎',
  };
  return labels[store.trim().toUpperCase()] ?? store;
};

// 23. 顯示分類名稱
const displayCategory = (category: string): string => category.split('/')[0]?.trim() || category;

// 24. 格式化日期
const formatDisplayDate = (value: string): string => {
  const parts = value.split('-');
  return parts.length === 3 ? `${parts[0]}年${parts[1]}月${parts[2]}日` : value;
};

// 25. 格式化日期刻度
const formatDateTick = (date: string): string => {
  const parts = date.split('-');
  return parts.length === 3 ? `${parts[1]}/${parts[2]}` : date;
};

// 26. 建立分類文字
const categoryText = (product: SupermarketProduct): string =>
  [product.category1, product.category2, product.category3].filter(Boolean).join(' / ');

// 27. 查找最優惠最近門店
const requestBestDealLocation = (): void => {
  if (!navigator.geolocation) {
    bestDealNearbyStatus.value = 'error';
    bestDealNearbyMessage.value = '此瀏覽器不支援定位功能。';
    return;
  }
  const stores = bestDealPrices.value.map((price) => price.store);
  bestDealNearbyStores.value = [];
  bestDealNearbyStatus.value = 'requesting';
  bestDealNearbyMessage.value = stores.length > 0 ? '正在詢問定位權限並整理最優惠門店。' : '未有可計算的優惠商戶。';
  if (stores.length === 0) {
    bestDealNearbyStatus.value = 'ready';
    return;
  }
  navigator.geolocation.getCurrentPosition(
    (position) => {
      bestDealNearbyStores.value = stores
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

// 28. 取得定位錯誤文案
const geolocationErrorMessage = (error: GeolocationPositionError): string => {
  if (error.code === error.PERMISSION_DENIED) return '未取得定位權限，請允許瀏覽器定位後再試。';
  if (error.code === error.POSITION_UNAVAILABLE) return '暫時無法取得目前位置。';
  if (error.code === error.TIMEOUT) return '定位逾時，請稍後再試。';
  return '定位失敗。';
};

// 29. 更新走勢圖浮層
const handleTrendHover = (event: MouseEvent): void => {
  const chart = detailTrendChart.value;
  if (!chart.hasData || chart.dates.length === 0) {
    trendHoverDate.value = '';
    return;
  }
  const target = event.currentTarget;
  if (!(target instanceof SVGSVGElement)) return;
  const rect = target.getBoundingClientRect();
  const rawX = ((event.clientX - rect.left) / Math.max(1, rect.width)) * chart.width;
  const ratio = Math.min(1, Math.max(0, (rawX - chart.padding.left) / Math.max(1, chart.width - chart.padding.left - chart.padding.right)));
  const index = Math.round(ratio * Math.max(0, chart.dates.length - 1));
  trendHoverDate.value = chart.dates[index] ?? '';
};

// 30. 清除走勢圖浮層
const clearTrendHover = (): void => {
  trendHoverDate.value = '';
};

// 31. 取得圖表價格
const priceForMode = (item: SupermarketDailyStorePrice, mode: PriceChartMode): number =>
  mode === 'effective' ? item.effectiveUnitPrice : item.listPrice;

// 32. 建立價格走勢圖資料
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
        offer: item.offer,
      }));
    return { store, color: colors[index % colors.length], points: dots.map((item) => `${item.x},${item.y}`).join(' '), dots };
  });
  const yTicks = [0, 0.25, 0.5, 0.75, 1].map((tick) => {
    const y = padding.top + tick * (height - padding.top - padding.bottom);
    return { key: String(tick), y, label: (maxValue - tick * span).toFixed(1) };
  });
  const xTicks = pickDateTicks(dates, 6).map((date) => ({ key: date, x: xFor(date), label: formatDateTick(date) }));
  return { width, height, padding, dates, series, yTicks, xTicks, hasData: history.length > 0 };
};

// 33. 選取日期刻度
const pickDateTicks = (dates: string[], maxTicks: number): string[] => {
  if (dates.length <= maxTicks) return dates;
  const lastIndex = dates.length - 1;
  const indexes = new Set<number>();
  for (let index = 0; index < maxTicks; index += 1) {
    indexes.add(Math.round((index / Math.max(1, maxTicks - 1)) * lastIndex));
  }
  return [...indexes].sort((a, b) => a - b).map((index) => dates[index]).filter(Boolean);
};

onMounted(() => {
  if (isDetailPage.value) {
    void loadDetail(productCode.value);
    return;
  }
  void loadInitialData();
});

watch(productCode, (code) => {
  if (code) {
    void loadDetail(code);
    return;
  }
  if (!searchResult.value) {
    void loadInitialData();
  }
});

watch(() => filters.q, (value) => {
  suggestionsOpen.value = true;
  if (suggestionTimer !== undefined) {
    window.clearTimeout(suggestionTimer);
  }
  const keyword = value.trim();
  if (!keyword) {
    suggestions.value = [];
    return;
  }
  suggestionTimer = window.setTimeout(() => {
    void loadSuggestions(keyword);
  }, 220);
});
</script>

<template>
  <main class="grocery-page">
    <section
      v-if="!isDetailPage"
      class="grocery-hero"
    >
      <div>
        <p>超市格價</p>
        <h1>今日最抵買</h1>
        <span>追蹤全港主要超市即時優惠，比較優惠後價格與折扣力度。</span>
      </div>
      <div class="grocery-hero__stats">
        <div>
          <strong>{{ (heroStats?.offers ?? 0).toLocaleString('zh-HK') }}</strong>
          <small>活躍優惠</small>
        </div>
        <i />
        <div>
          <strong>{{ (heroStats?.stores ?? 0).toLocaleString('zh-HK') }}</strong>
          <small>連鎖商店</small>
        </div>
        <i />
        <div>
          <strong>{{ highestDiscount }}</strong>
          <small>最高折扣</small>
        </div>
      </div>
    </section>

    <template v-if="!isDetailPage">
      <section class="grocery-controls">
        <div class="grocery-search-row">
          <form
            class="grocery-search"
            @submit.prevent="submitSearch"
          >
            <span>⌕</span>
            <input
              v-model="filters.q"
              placeholder="搜尋商品、品牌或分類..."
              @blur="closeSuggestionsSoon"
              @focus="suggestionsOpen = true"
            >
            <div
              v-if="suggestionsOpen && suggestions.length > 0"
              class="grocery-search__suggestions"
            >
              <button
                v-for="product in suggestions"
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
          <button
            type="button"
            class="grocery-search-submit"
            @click="submitSearch"
          >
            搜尋
          </button>
          <button
            type="button"
            class="grocery-favorites-button"
            :class="listMode === 'favorites' ? 'grocery-favorites-button--active' : ''"
            @click="setListMode(listMode === 'favorites' ? 'all' : 'favorites')"
          >
            {{ listMode === 'favorites' ? '全部優惠' : '我的收藏' }}
          </button>
        </div>
        <div class="grocery-pills">
          <button
            type="button"
            class="grocery-pill"
            :class="filters.category === '' ? 'grocery-pill--active' : ''"
            @click="filters.category = ''; applyFilter()"
          >
            全部
          </button>
          <button
            v-for="category in categoryOptions.slice(0, 8)"
            :key="category"
            type="button"
            class="grocery-pill"
            :class="filters.category === category ? 'grocery-pill--active' : ''"
            @click="filters.category = category; applyFilter()"
          >
            {{ displayCategory(category) }}
          </button>
        </div>
        <div class="grocery-store-pills">
          <button
            type="button"
            class="grocery-store-pill"
            :class="filters.store === '' ? 'grocery-store-pill--active' : ''"
            @click="filters.store = ''; applyFilter()"
          >
            全部商店
          </button>
          <button
            v-for="store in storeOptions.slice(0, 8)"
            :key="store"
            type="button"
            class="grocery-store-pill"
            :class="filters.store === store ? 'grocery-store-pill--active' : ''"
            @click="filters.store = store; applyFilter()"
          >
            {{ displayStore(store) }}
          </button>
        </div>
      </section>

      <section class="grocery-content">
        <div class="grocery-content__header">
          <span>
            {{ loading || favoritesLoading ? '載入中' : `${visibleTotal.toLocaleString('zh-HK')} 個${listMode === 'favorites' ? '收藏' : '優惠'}` }}
          </span>
          <div class="grocery-toolbar">
            <div class="grocery-view-toggle">
              <button
                type="button"
                :class="viewMode === 'grid' ? 'grocery-view-toggle__active' : ''"
                @click="viewMode = 'grid'"
              >
                網格
              </button>
              <button
                type="button"
                :class="viewMode === 'table' ? 'grocery-view-toggle__active' : ''"
                @click="viewMode = 'table'"
              >
                表格
              </button>
            </div>
            <select
              v-model="filters.sort"
              @change="applyFilter"
            >
              <option value="discount">優惠力度 ↓</option>
              <option value="effective">優惠後價格</option>
              <option value="diff">價差</option>
              <option value="name">商品名稱</option>
            </select>
          </div>
        </div>

        <p
          v-if="errorMessage"
          class="grocery-state grocery-state--error"
        >
          {{ errorMessage }}
        </p>
        <p
          v-else-if="(loading && !searchResult) || favoritesLoading"
          class="grocery-state"
        >
          正在載入{{ listMode === 'favorites' ? '收藏' : '優惠資料' }}。
        </p>
        <p
          v-else-if="visibleItems.length === 0"
          class="grocery-state"
        >
          沒有找到符合條件的{{ listMode === 'favorites' ? '收藏' : '優惠' }}。
        </p>

        <div
          v-if="visibleItems.length > 0 && viewMode === 'grid'"
          class="grocery-grid"
        >
          <article
            v-for="product in visibleItems"
            :key="product.code"
            class="grocery-card"
            @click="openProduct(product)"
          >
            <button
              type="button"
              class="grocery-card__favorite"
              :disabled="favoriteSavingCode === product.code"
              @click.stop="toggleFavorite(product)"
            >
              {{ product.isFavorite ? '已收藏' : '收藏' }}
            </button>
            <div class="grocery-card__body">
              <div class="grocery-card__heading">
                <div>
                  <p>{{ product.brand || '未提供品牌' }}</p>
                  <h2>{{ product.name }}</h2>
                </div>
                <span>-{{ discountRate(primaryPrice(product)).toFixed(0) }}%</span>
              </div>
              <div class="grocery-card__prices">
                <article
                  v-for="price in normalizedStorePrices(product).slice(0, 4)"
                  :key="`${product.code}-${price.store}-${price.effectiveUnitPrice}`"
                >
                  <div>
                    <strong>{{ displayStore(price.store) }}</strong>
                    <small v-if="price.offer">{{ price.offer }}</small>
                  </div>
                  <div>
                    <b>{{ formatHKPrice(price.effectiveUnitPrice) }} / 件</b>
                    <span v-if="price.listPrice > price.effectiveUnitPrice">
                      {{ formatHKPrice(price.listPrice) }}(-{{ discountRate(price).toFixed(0) }}%)
                    </span>
                  </div>
                </article>
              </div>
            </div>
          </article>
        </div>

        <div
          v-if="visibleItems.length > 0 && viewMode === 'table'"
          class="grocery-table"
        >
          <table>
            <thead>
              <tr>
                <th>商品</th>
                <th>商店</th>
                <th>優惠</th>
                <th>優惠後</th>
                <th>原價</th>
                <th>收藏</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="product in visibleItems"
                :key="product.code"
                @click="openProduct(product)"
              >
                <td>
                  <strong>{{ product.name }}</strong>
                  <span>{{ product.brand || '未提供品牌' }}</span>
                </td>
                <td>{{ displayStore(primaryPrice(product).store) }}</td>
                <td>{{ primaryPrice(product).offer || product.bestOffer || '-' }}</td>
                <td>{{ formatHKPrice(primaryPrice(product).effectiveUnitPrice) }}</td>
                <td>{{ formatHKPrice(primaryPrice(product).listPrice) }}</td>
                <td>
                  <button
                    type="button"
                    class="grocery-table__favorite"
                    :disabled="favoriteSavingCode === product.code"
                    @click.stop="toggleFavorite(product)"
                  >
                    {{ product.isFavorite ? '已收藏' : '收藏' }}
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <div
          v-if="listMode === 'all' && searchResult && searchResult.total > searchResult.pageSize"
          class="grocery-pagination"
        >
          <button
            type="button"
            :disabled="filters.page <= 1"
            @click="changePage(filters.page - 1)"
          >
            上一頁
          </button>
          <button
            v-for="page in pageNumbers"
            :key="page"
            type="button"
            :class="filters.page === page ? 'grocery-pagination__active' : ''"
            @click="changePage(page)"
          >
            {{ page }}
          </button>
          <button
            type="button"
            :disabled="filters.page >= totalPages"
            @click="changePage(filters.page + 1)"
          >
            下一頁
          </button>
        </div>

        <p class="grocery-updated">
          資料更新：{{ latestDateText || '等待資料' }} · 共監測 {{ (heroStats?.products ?? 0).toLocaleString('zh-HK') }} 件商品
        </p>
      </section>
    </template>

    <section
      v-else
      class="grocery-detail"
    >
      <nav
        class="grocery-breadcrumb"
        aria-label="Breadcrumb"
      >
        <button
          type="button"
          @click="backToList"
        >
          綜合優惠
        </button>
        <span>/</span>
        <strong>{{ detail?.product.name || '商品詳情' }}</strong>
      </nav>
      <p
        v-if="detailErrorMessage"
        class="grocery-state grocery-state--error"
      >
        {{ detailErrorMessage }}
      </p>
      <p
        v-else-if="detailLoading && !detail"
        class="grocery-state"
      >
        正在載入商品詳情。
      </p>

      <template v-if="detail">
        <section class="grocery-detail__hero">
          <div>
            <span>商品詳情</span>
            <h1>{{ detail.product.name }}</h1>
            <p>{{ detail.product.brand || '未提供品牌' }} · {{ categoryText(detail.product) || '未提供分類' }}</p>
          </div>
          <button
            type="button"
            :disabled="favoriteSavingCode === detail.product.code"
            @click="toggleFavorite(detail.product)"
          >
            {{ detail.isFavorite || detail.product.isFavorite ? '已收藏' : '收藏' }}
          </button>
        </section>

        <section class="grocery-detail__store-section">
          <div class="grocery-detail__section-title">
            <div>
              <h2>超市優惠與門店地址</h2>
              <p>先看有優惠的最新超市，再查最優惠商戶或最近門店路線。</p>
            </div>
            <span>門店列表可直接開 Google Maps</span>
          </div>
          <div class="grocery-detail__price-list">
            <article
              v-for="price in (detailDiscountPrices.length > 0 ? detailDiscountPrices : detailStorePrices).slice(0, 8)"
              :key="`${price.store}-${price.effectiveUnitPrice}-${price.offer}`"
            >
              <div>
                <strong>{{ displayStore(price.store) }}</strong>
                <small v-if="price.offer">{{ price.offer }}</small>
                <small v-else>未提供優惠文案</small>
              </div>
              <div>
                <b>{{ formatHKPrice(price.effectiveUnitPrice) }} / 件</b>
                <span v-if="price.listPrice > price.effectiveUnitPrice">
                  {{ formatHKPrice(price.listPrice) }}(-{{ discountRate(price).toFixed(0) }}%)
                </span>
                <span v-else>{{ formatHKPrice(price.listPrice) }}</span>
              </div>
            </article>
          </div>

          <section class="grocery-detail__nearby">
            <div class="grocery-detail__nearby-head">
              <div>
                <h3>最優惠最近門店地址</h3>
                <p>{{ bestDealSummary }}</p>
              </div>
              <button
                type="button"
                :disabled="bestDealPrices.length === 0 || bestDealNearbyStatus === 'requesting'"
                @click="requestBestDealLocation"
              >
                {{ bestDealNearbyStatus === 'requesting' ? '定位中' : '查找最優惠門店' }}
              </button>
            </div>
            <p
              v-if="bestDealNearbyMessage"
              class="grocery-detail__nearby-message"
              :class="bestDealNearbyStatus === 'error' ? 'grocery-detail__nearby-message--error' : ''"
            >
              {{ bestDealNearbyStatus === 'requesting' ? '定位中。' : bestDealNearbyMessage }}
            </p>
            <div
              v-if="bestDealNearbyStores.length > 0"
              class="grocery-detail__nearby-table"
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
                      >{{ index + 1 }}</a>
                    </td>
                    <td>
                      <a
                        :href="supermarketDirectionsUrl(store)"
                        target="_blank"
                        rel="noreferrer"
                      >{{ displayStore(store.storeCode) }}</a>
                    </td>
                    <td>
                      <a
                        :href="supermarketDirectionsUrl(store)"
                        target="_blank"
                        rel="noreferrer"
                      >{{ store.name }}</a>
                    </td>
                    <td>
                      <a
                        :href="supermarketDirectionsUrl(store)"
                        target="_blank"
                        rel="noreferrer"
                      >{{ store.address }}</a>
                    </td>
                    <td>
                      <a
                        :href="supermarketDirectionsUrl(store)"
                        target="_blank"
                        rel="noreferrer"
                      >{{ formatSupermarketDistanceKm(store.distanceKm) }}</a>
                    </td>
                    <td>
                      <a
                        :href="supermarketDirectionsUrl(store)"
                        target="_blank"
                        rel="noreferrer"
                      >{{ store.hours || '-' }}</a>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </section>
        </section>

        <section class="grocery-detail__insights">
          <aside class="grocery-detail__summary">
            <article>
              <span>90日最低優惠價</span>
              <strong>{{ formatHKPrice(detail.summary.lowestDeal) }}</strong>
            </article>
            <article>
              <span>90日最高優惠價</span>
              <strong>{{ formatHKPrice(detail.summary.highestDeal) }}</strong>
            </article>
            <article>
              <span>90日最低原價</span>
              <strong>{{ formatHKPrice(detail.summary.lowestList) }}</strong>
            </article>
            <article>
              <span>90日最高原價</span>
              <strong>{{ formatHKPrice(detail.summary.highestList) }}</strong>
            </article>
          </aside>

          <article class="grocery-detail__chart-panel">
            <div class="grocery-detail__section-title">
              <div>
                <h2>90日超市價格走勢</h2>
                <p>每條線代表一間超市，可切換原價與優惠後等效單件價。</p>
              </div>
              <div class="grocery-detail__chart-toggle">
                <button
                  type="button"
                  :class="detailChartMode === 'effective' ? 'grocery-detail__chart-toggle-active' : ''"
                  @click="detailChartMode = 'effective'"
                >
                  優惠價
                </button>
                <button
                  type="button"
                  :class="detailChartMode === 'list' ? 'grocery-detail__chart-toggle-active' : ''"
                  @click="detailChartMode = 'list'"
                >
                  原價
                </button>
              </div>
            </div>
            <div
              v-if="detailTrendChart.hasData"
              class="grocery-detail__chart"
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
                  class="grocery-detail__hover-line"
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
              </svg>
              <div
                v-if="detailTrendHover"
                class="grocery-detail__chart-tooltip"
                :style="{ left: detailTrendHover.tooltipLeft }"
              >
                <strong>{{ formatDateTick(detailTrendHover.date) }}</strong>
                <ul>
                  <li
                    v-for="item in detailTrendHover.items"
                    :key="`${item.key}-tooltip`"
                  >
                    <i :style="{ background: item.color }" />
                    <span>{{ displayStore(item.store) }}</span>
                    <em>{{ formatHKPrice(item.value) }}</em>
                  </li>
                </ul>
              </div>
              <div class="grocery-detail__legend">
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
              class="grocery-state"
            >
              未有歷史資料。
            </p>
          </article>
        </section>

        <section class="grocery-detail__history">
          <div class="grocery-detail__section-title">
            <div>
              <h2>最近價格記錄</h2>
              <p>顯示已入庫的最新價格，方便核對走勢。</p>
            </div>
          </div>
          <div class="grocery-detail__history-list">
            <div
              v-for="item in detailRecentHistory"
              :key="`${item.date}-${item.store}-${item.effectiveUnitPrice}`"
            >
              <span>{{ formatDisplayDate(item.date) }}</span>
              <strong>{{ displayStore(item.store) }}</strong>
              <b>{{ formatHKPrice(item.effectiveUnitPrice) }}</b>
              <small v-if="item.offer">{{ item.offer }}</small>
            </div>
          </div>
        </section>
      </template>
    </section>
  </main>
</template>

<style scoped>
.grocery-page {
  min-height: calc(100vh - var(--app-header-offset, 48px));
  background: #f4f4f4;
  color: #1a1a1a;
}

.grocery-hero {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
  border-bottom: 3px solid rgb(var(--color-primary));
  background: #1a1a1a;
  padding: 28px 32px;
}

.grocery-hero p {
  margin: 0 0 8px;
  color: #aaaaaa;
  font-size: 9px;
  letter-spacing: 3px;
  text-transform: uppercase;
}

.grocery-hero h1 {
  margin: 0 0 8px;
  color: #ffffff;
  font-family: var(--font-display);
  font-size: 32px;
  font-weight: 400;
  line-height: 1.1;
}

.grocery-hero span {
  display: block;
  max-width: 420px;
  color: #bbbbbb;
  font-size: 12px;
  line-height: 1.6;
}

.grocery-hero__stats {
  display: flex;
  align-items: center;
  flex-shrink: 0;
}

.grocery-hero__stats div {
  text-align: center;
  padding: 0 28px;
}

.grocery-hero__stats strong {
  display: block;
  color: #ffffff;
  font-size: 26px;
  font-weight: 300;
  letter-spacing: 0;
  line-height: 1;
}

.grocery-hero__stats small {
  display: block;
  color: #aaaaaa;
  font-size: 10px;
  letter-spacing: 0.3px;
  margin-top: 4px;
}

.grocery-hero__stats i {
  display: block;
  width: 1px;
  height: 40px;
  background: #444444;
}

.grocery-controls {
  display: flex;
  flex-direction: column;
  gap: 10px;
  border-bottom: 1px solid #e4e4e4;
  background: #ffffff;
  padding: 14px 24px;
}

.grocery-search-row {
  display: flex;
  max-width: 820px;
  align-items: stretch;
  gap: 8px;
}

.grocery-search {
  position: relative;
  display: flex;
  align-items: center;
  flex: 1;
  gap: 8px;
  border: 1px solid #e4e4e4;
  border-radius: 3px;
  padding: 0 14px;
}

.grocery-search span {
  color: #777777;
  font-size: 21px;
  line-height: 1;
}

.grocery-search input {
  flex: 1;
  min-width: 0;
  border: 0;
  font: inherit;
  font-size: 13px;
  outline: 0;
  padding: 10px 0;
}

.grocery-search-submit,
.grocery-favorites-button,
.grocery-detail__hero button {
  border-radius: 3px;
  background: rgb(var(--color-primary));
  color: #ffffff;
  font-size: 12px;
  font-weight: 600;
  padding: 0 16px;
}

.grocery-favorites-button {
  border: 1px solid rgb(var(--color-primary));
  background: #ffffff;
  color: rgb(var(--color-primary));
  min-width: 88px;
}

.grocery-favorites-button--active {
  background: rgb(var(--color-primary));
  color: #ffffff;
}

.grocery-search__suggestions {
  position: absolute;
  z-index: 20;
  top: calc(100% + 6px);
  right: 0;
  left: 0;
  overflow: hidden;
  border: 1px solid #e4e4e4;
  border-radius: 4px;
  background: #ffffff;
  box-shadow: 0 16px 40px rgb(0 0 0 / 0.12);
}

.grocery-search__suggestions button {
  display: flex;
  width: 100%;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  border: 0;
  border-bottom: 1px solid #eeeeee;
  border-radius: 0;
  background: #ffffff;
  color: #1a1a1a;
  padding: 10px 12px;
  text-align: left;
}

.grocery-search__suggestions button:last-child {
  border-bottom: 0;
}

.grocery-search__suggestions button:hover {
  background: #fafafa;
}

.grocery-search__suggestions span,
.grocery-search__suggestions strong,
.grocery-search__suggestions small {
  display: block;
  min-width: 0;
}

.grocery-search__suggestions strong {
  overflow: hidden;
  font-size: 12px;
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.grocery-search__suggestions small,
.grocery-search__suggestions em {
  color: #777777;
  font-size: 11px;
  font-style: normal;
}

.grocery-search__suggestions em {
  flex: 0 0 auto;
  color: rgb(var(--color-primary));
  font-weight: 600;
}

.grocery-pills,
.grocery-store-pills {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.grocery-pill,
.grocery-store-pill,
.grocery-view-toggle button,
.grocery-pagination button,
.grocery-breadcrumb button,
.grocery-favorites-button,
.grocery-card__favorite,
.grocery-table__favorite {
  border: 1px solid #e4e4e4;
  background: #ffffff;
  color: #777777;
  cursor: pointer;
  font-size: 11px;
}

.grocery-pill {
  border-radius: 20px;
  padding: 4px 12px;
}

.grocery-store-pill {
  border-radius: 3px;
  padding: 3px 10px;
}

.grocery-pill--active,
.grocery-view-toggle__active,
.grocery-pagination__active {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary));
  color: #ffffff;
}

.grocery-store-pill--active {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary-soft));
  color: rgb(var(--color-border-strong));
  font-weight: 500;
}

.grocery-content,
.grocery-detail {
  padding: 20px 24px 40px;
}

.grocery-content__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 16px;
}

.grocery-content__header span {
  color: #777777;
  font-size: 12px;
}

.grocery-toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
}

.grocery-view-toggle {
  display: flex;
  overflow: hidden;
  border-radius: 3px;
}

.grocery-view-toggle button {
  padding: 6px 10px;
}

.grocery-content__header select {
  border: 1px solid #e4e4e4;
  border-radius: 2px;
  background: #ffffff;
  font: inherit;
  font-size: 11px;
  outline: 0;
  padding: 6px 10px;
}

.grocery-state {
  border: 1px dashed #d8d8d8;
  border-radius: 4px;
  background: #ffffff;
  color: #777777;
  font-size: 13px;
  margin: 0 0 16px;
  padding: 16px;
}

.grocery-state--error {
  border-color: #c2410c;
  color: #c2410c;
}

.grocery-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
}

.grocery-card {
  position: relative;
  overflow: hidden;
  border: 1px solid #e4e4e4;
  border-radius: 3px;
  background: #ffffff;
  cursor: pointer;
  min-height: 238px;
}

.grocery-card:hover {
  border-color: rgb(var(--color-primary));
  box-shadow: 0 4px 18px rgb(240 90 0 / 0.16);
}

.grocery-card__favorite {
  position: absolute;
  z-index: 2;
  top: 10px;
}

.grocery-card__favorite {
  right: 10px;
  border-radius: 2px;
  padding: 3px 7px;
}

.grocery-card__body {
  display: grid;
  gap: 12px;
  padding: 14px;
}

.grocery-card__heading {
  display: grid;
  gap: 10px;
  grid-template-columns: minmax(0, 1fr) auto;
  padding-right: 58px;
}

.grocery-card__heading > div {
  min-width: 0;
}

.grocery-card__heading p {
  margin: 0 0 4px;
  color: #777777;
  font-size: 11px;
}

.grocery-card__heading h2 {
  display: -webkit-box;
  overflow: hidden;
  margin: 0;
  color: #1a1a1a;
  font-size: 15px;
  font-weight: 600;
  line-height: 1.35;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.grocery-card__heading span {
  align-self: start;
  border-radius: 2px;
  background: rgb(var(--color-primary));
  color: #ffffff;
  font-size: 11px;
  font-weight: 600;
  padding: 3px 7px;
}

.grocery-card__prices {
  display: grid;
  gap: 8px;
}

.grocery-card__prices article {
  display: grid;
  align-items: start;
  gap: 10px;
  grid-template-columns: minmax(0, 1fr) auto;
  border-top: 1px solid #f0f0f0;
  padding-top: 8px;
}

.grocery-card__prices article:first-child {
  border-top: 0;
  padding-top: 0;
}

.grocery-card__prices strong,
.grocery-card__prices small,
.grocery-card__prices b,
.grocery-card__prices span {
  display: block;
}

.grocery-card__prices strong {
  color: #1a1a1a;
  font-size: 12px;
  font-weight: 600;
}

.grocery-card__prices small {
  display: -webkit-box;
  overflow: hidden;
  color: #777777;
  font-size: 11px;
  margin-top: 3px;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.grocery-card__prices article > div:last-child {
  text-align: right;
}

.grocery-card__prices b {
  color: #1a1a1a;
  font-size: 13px;
  font-weight: 600;
  white-space: nowrap;
}

.grocery-card__prices span {
  color: rgb(var(--color-primary));
  font-size: 11px;
  margin-top: 3px;
  white-space: nowrap;
}

.grocery-table {
  overflow-x: auto;
  border: 1px solid #e4e4e4;
  border-radius: 3px;
  background: #ffffff;
}

.grocery-table table {
  width: 100%;
  min-width: 760px;
  border-collapse: collapse;
  text-align: left;
}

.grocery-table th,
.grocery-table td {
  border-bottom: 1px solid #eeeeee;
  padding: 10px 12px;
}

.grocery-table th {
  color: #777777;
  font-size: 11px;
  font-weight: 500;
}

.grocery-table td {
  color: #333333;
  font-size: 12px;
}

.grocery-table tbody tr {
  cursor: pointer;
}

.grocery-table tbody tr:hover {
  background: #fafafa;
}

.grocery-table td strong,
.grocery-table td span {
  display: block;
}

.grocery-table td span {
  color: #777777;
  font-size: 11px;
  margin-top: 3px;
}

.grocery-table__favorite {
  border-radius: 3px;
  padding: 4px 8px;
}

.grocery-pagination {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 6px;
  margin-top: 18px;
}

.grocery-pagination button {
  border-radius: 3px;
  min-width: 34px;
  padding: 6px 10px;
}

.grocery-pagination button:disabled,
.grocery-card__favorite:disabled,
.grocery-table__favorite:disabled {
  cursor: not-allowed;
  opacity: 0.45;
}

.grocery-updated {
  margin: 20px 0 0;
  color: #777777;
  font-size: 10px;
  letter-spacing: 0.3px;
  text-align: center;
}

.grocery-detail {
  display: grid;
  gap: 16px;
}

.grocery-breadcrumb {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 8px;
  color: #777777;
  font-size: 12px;
}

.grocery-breadcrumb button {
  border-radius: 3px;
  color: rgb(var(--color-primary));
  padding: 6px 10px;
}

.grocery-breadcrumb span {
  color: #aaaaaa;
}

.grocery-breadcrumb strong {
  overflow: hidden;
  color: #333333;
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.grocery-detail__hero,
.grocery-detail__store-section,
.grocery-detail__summary {
  border: 1px solid #e4e4e4;
  border-radius: 3px;
  background: #ffffff;
}

.grocery-detail__hero {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 22px;
}

.grocery-detail__hero span,
.grocery-detail__hero p,
.grocery-detail__prices small,
.grocery-detail__summary span {
  color: #777777;
  font-size: 12px;
}

.grocery-detail__hero h1 {
  margin: 7px 0;
  color: #1a1a1a;
  font-size: 26px;
  line-height: 1.2;
}

.grocery-detail__hero p {
  margin: 0;
}

.grocery-detail__store-section {
  display: grid;
  gap: 14px;
  padding: 18px;
}

.grocery-detail__section-title {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.grocery-detail__section-title h2 {
  margin: 0;
  color: #1a1a1a;
  font-size: 18px;
  font-weight: 700;
}

.grocery-detail__section-title p {
  margin: 4px 0 0;
  color: #777777;
  font-size: 12px;
  line-height: 1.5;
}

.grocery-detail__section-title > span {
  border-radius: 3px;
  background: #f4f4f4;
  color: #777777;
  flex: 0 0 auto;
  font-size: 11px;
  padding: 5px 8px;
}

.grocery-detail__price-list {
  display: grid;
  gap: 10px;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.grocery-detail__price-list article {
  display: grid;
  align-items: start;
  gap: 12px;
  grid-template-columns: minmax(0, 1fr) auto;
  border: 1px solid #eeeeee;
  border-radius: 4px;
  background: #fafafa;
  padding: 12px;
}

.grocery-detail__price-list strong,
.grocery-detail__price-list small,
.grocery-detail__price-list b,
.grocery-detail__price-list span {
  display: block;
}

.grocery-detail__price-list strong {
  color: #1a1a1a;
  font-size: 13px;
  font-weight: 700;
}

.grocery-detail__price-list small {
  display: -webkit-box;
  overflow: hidden;
  color: #777777;
  font-size: 12px;
  line-height: 1.45;
  margin-top: 4px;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.grocery-detail__price-list article > div:last-child {
  text-align: right;
}

.grocery-detail__price-list b {
  color: #1a1a1a;
  font-size: 15px;
  font-weight: 700;
  white-space: nowrap;
}

.grocery-detail__price-list span {
  color: rgb(var(--color-primary));
  font-size: 12px;
  margin-top: 4px;
  white-space: nowrap;
}

.grocery-detail__nearby {
  border: 1px solid #eeeeee;
  border-radius: 4px;
  background: #ffffff;
  padding: 14px;
}

.grocery-detail__nearby-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.grocery-detail__nearby-head h3 {
  margin: 0;
  color: #1a1a1a;
  font-size: 16px;
  font-weight: 700;
}

.grocery-detail__nearby-head p {
  margin: 4px 0 0;
  color: #777777;
  font-size: 12px;
  line-height: 1.5;
}

.grocery-detail__nearby-head button {
  border-radius: 4px;
  background: rgb(var(--color-primary));
  color: #ffffff;
  flex: 0 0 auto;
  font-size: 12px;
  font-weight: 600;
  padding: 8px 12px;
}

.grocery-detail__nearby-head button:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

.grocery-detail__nearby-message {
  border: 1px solid #eeeeee;
  border-radius: 4px;
  background: #fafafa;
  color: #777777;
  font-size: 12px;
  margin: 12px 0 0;
  padding: 10px 12px;
}

.grocery-detail__nearby-message--error {
  border-color: #c2410c;
  color: #c2410c;
}

.grocery-detail__nearby-table {
  overflow-x: auto;
  border: 1px solid #eeeeee;
  border-radius: 4px;
  margin-top: 12px;
}

.grocery-detail__nearby-table table {
  width: 100%;
  min-width: 760px;
  border-collapse: collapse;
  text-align: left;
}

.grocery-detail__nearby-table th {
  border-bottom: 1px solid #eeeeee;
  background: #fafafa;
  color: #777777;
  font-size: 11px;
  font-weight: 700;
  padding: 9px 10px;
}

.grocery-detail__nearby-table td {
  border-bottom: 1px solid #eeeeee;
  color: #333333;
  font-size: 12px;
  padding: 0;
}

.grocery-detail__nearby-table tbody tr:last-child td {
  border-bottom: 0;
}

.grocery-detail__nearby-table a {
  display: block;
  color: inherit;
  padding: 10px;
}

.grocery-detail__insights {
  display: grid;
  gap: 16px;
  grid-template-columns: minmax(240px, 0.34fr) minmax(0, 0.66fr);
}

.grocery-detail__summary {
  display: grid;
  gap: 0;
  grid-template-columns: 1fr;
}

.grocery-detail__summary article {
  display: grid;
  gap: 5px;
  border-bottom: 1px solid #eeeeee;
  padding: 16px;
}

.grocery-detail__summary article:last-child {
  border-bottom: 0;
}

.grocery-detail__summary strong {
  color: #1a1a1a;
  font-size: 18px;
  font-weight: 500;
}

.grocery-detail__chart-panel,
.grocery-detail__history {
  border: 1px solid #e4e4e4;
  border-radius: 3px;
  background: #ffffff;
  padding: 18px;
}

.grocery-detail__chart-toggle {
  display: inline-flex;
  overflow: hidden;
  border: 1px solid #e4e4e4;
  border-radius: 4px;
}

.grocery-detail__chart-toggle button {
  background: #ffffff;
  color: #777777;
  font-size: 11px;
  font-weight: 600;
  padding: 6px 10px;
}

.grocery-detail__chart-toggle-active {
  background: rgb(var(--color-primary)) !important;
  color: #ffffff !important;
}

.grocery-detail__chart {
  position: relative;
  display: grid;
  gap: 10px;
  min-height: 320px;
  margin-top: 14px;
}

.grocery-detail__chart svg {
  display: block;
  width: 100%;
  min-height: 280px;
  cursor: crosshair;
}

.grocery-detail__chart rect {
  fill: #ffffff;
}

.grocery-detail__chart line {
  stroke: #e4e4e4;
  stroke-width: 1.1;
}

.grocery-detail__chart text {
  fill: #777777;
  font-size: 10px;
  font-weight: 600;
}

.grocery-detail__chart polyline {
  fill: none;
  stroke-linecap: round;
  stroke-linejoin: round;
  stroke-width: 2.4;
}

.grocery-detail__chart circle {
  stroke: #ffffff;
  stroke-width: 1.8;
}

.grocery-detail__hover-line {
  stroke: #1a1a1a !important;
  stroke-dasharray: 5 5;
  stroke-width: 1.5 !important;
}

.grocery-detail__chart-tooltip {
  position: absolute;
  z-index: 3;
  top: 12px;
  width: min(240px, calc(100% - 24px));
  transform: translateX(-50%);
  border: 1px solid #e4e4e4;
  border-radius: 4px;
  background: #ffffff;
  box-shadow: 0 14px 34px rgb(0 0 0 / 0.14);
  padding: 10px;
  pointer-events: none;
}

.grocery-detail__chart-tooltip > strong {
  display: block;
  border-bottom: 1px solid #eeeeee;
  color: #1a1a1a;
  font-size: 12px;
  padding-bottom: 7px;
}

.grocery-detail__chart-tooltip ul {
  display: grid;
  gap: 7px;
  margin: 8px 0 0;
  padding: 0;
}

.grocery-detail__chart-tooltip li {
  display: grid;
  align-items: center;
  gap: 6px;
  grid-template-columns: auto minmax(0, 1fr) auto;
  list-style: none;
}

.grocery-detail__chart-tooltip i,
.grocery-detail__legend i {
  width: 9px;
  height: 9px;
  border-radius: 999px;
}

.grocery-detail__chart-tooltip span {
  overflow: hidden;
  color: #333333;
  font-size: 11px;
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.grocery-detail__chart-tooltip em {
  color: rgb(var(--color-primary));
  font-size: 11px;
  font-style: normal;
  font-weight: 700;
}

.grocery-detail__legend {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 12px;
}

.grocery-detail__legend span {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  color: #777777;
  font-size: 11px;
  font-weight: 600;
}

.grocery-detail__history-list {
  display: grid;
  gap: 0;
  margin-top: 12px;
}

.grocery-detail__history-list div {
  display: grid;
  align-items: center;
  gap: 10px;
  grid-template-columns: 96px minmax(0, 1fr) auto minmax(0, 1.3fr);
  border-top: 1px solid #eeeeee;
  padding: 10px 0;
}

.grocery-detail__history-list div:first-child {
  border-top: 0;
}

.grocery-detail__history-list span,
.grocery-detail__history-list small {
  color: #777777;
  font-size: 11px;
}

.grocery-detail__history-list strong {
  color: #1a1a1a;
  font-size: 12px;
  font-weight: 700;
}

.grocery-detail__history-list b {
  color: rgb(var(--color-primary));
  font-size: 12px;
  font-weight: 700;
  white-space: nowrap;
}

@media (max-width: 1023px) {
  .grocery-hero,
  .grocery-detail__hero,
  .grocery-detail__nearby-head {
    align-items: flex-start;
    flex-direction: column;
  }

  .grocery-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .grocery-detail__price-list,
  .grocery-detail__insights {
    grid-template-columns: repeat(2, 1fr);
  }

  .grocery-detail__insights {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 640px) {
  .grocery-content__header,
  .grocery-toolbar,
  .grocery-search-row {
    align-items: stretch;
    flex-direction: column;
  }

  .grocery-grid,
  .grocery-detail__price-list,
  .grocery-detail__summary,
  .grocery-detail__history-list div {
    grid-template-columns: 1fr;
  }

  .grocery-detail__section-title {
    flex-direction: column;
  }

  .grocery-hero__stats {
    align-items: stretch;
    flex-direction: column;
    gap: 12px;
  }

  .grocery-hero__stats i {
    width: 100%;
    height: 1px;
  }
}
</style>

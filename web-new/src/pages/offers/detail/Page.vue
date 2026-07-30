<!--
 * 綜合優惠商品詳情頁。
 * 1. 使用 AJO 後端商品詳情接口展示即時價格與 90 日走勢。
 * 2. 支援會員收藏、到價提示與最優惠商戶附近門店查找。
 * 3. 對齊 docs 高保真參考的 gp-detail-panel 結構與樣式。
-->
<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import {
  addSupermarketFavorite,
  deleteSupermarketPriceAlert,
  fetchSupermarketProductDetail,
  removeSupermarketFavorite,
  saveSupermarketPriceAlert,
} from '@/httpapis/supermarket-offers';
import { readStoredAccessToken } from '@/httpapis/auth-session';
import type {
  SupermarketDailyStorePrice,
  SupermarketPriceAlert,
  SupermarketProduct,
  SupermarketProductDetail,
  SupermarketStorePrice,
} from '@/model/supermarket-offers';
import { usePreferenceStore } from '@/stores/preferences';
import {
  displaySupermarketCategory,
  displaySupermarketStore,
  formatSupermarketDate,
  formatSupermarketHKPrice,
  supermarketBestDealStorePrices,
  supermarketCurrentStorePrices,
  supermarketDiscountStorePrices,
  supermarketOfferTexts,
  supermarketPriceDiscountRate,
  supermarketPrimaryPrice,
} from '@/utils/supermarket-offers';
import {
  formatSupermarketDistanceKm,
  nearestSupermarketStoreLocations,
  SUPERMARKET_GEOLOCATION_CACHE_MAX_AGE_MS,
  SUPERMARKET_NEARBY_STORE_LIMIT,
  supermarketDirectionsUrl,
} from '../constants/store-locations';
import type { SupermarketNearbyStoreLocation } from '../constants/store-locations';

type PriceChartMode = 'effective' | 'list';
type NearbyStoreStatus = 'idle' | 'requesting' | 'ready' | 'error';

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

interface TrendChartHoverItem extends TrendChartPoint {
  color: string;
}

const route = useRoute();
const router = useRouter();
const { t } = useI18n();
const preferenceStore = usePreferenceStore();
const detail = ref<SupermarketProductDetail | null>(null);
const detailLoading = ref(false);
const detailError = ref('');
const actionMessage = ref('');
const savingFavorite = ref(false);
const savingAlert = ref(false);
const alertFormOpen = ref(false);
const chartMode = ref<PriceChartMode>('effective');
const trendHoverDate = ref('');
const nearbyStores = ref<SupermarketNearbyStoreLocation[]>([]);
const nearbyStatus = ref<NearbyStoreStatus>('idle');
const nearbyMessage = ref('');
const failedImageCodes = ref(new Set<string>());
const alertForm = reactive({
  targetPrice: '',
  priceMode: 'effective' as PriceChartMode,
  offerRequired: false,
  enabled: true,
});

// 1. 取得路由商品編號
const productCode = computed(() => {
  const rawCode = route.params.code;
  return Array.isArray(rawCode) ? rawCode[0] ?? '' : rawCode ?? '';
});

// 2. 取得目前商品
const product = computed<SupermarketProduct | null>(() => detail.value?.product ?? null);

// 3. 取得商品分類文字
const productCategoryText = computed(() => {
  if (!product.value) return '';
  return [product.value.category1, product.value.category2, product.value.category3]
    .filter(Boolean)
    .map((value) => displaySupermarketCategory(value, preferenceStore.locale))
    .join(' / ') || displaySupermarketCategory('', preferenceStore.locale);
});

// 4. 取得最新商店價格
const latestStorePrices = computed<SupermarketStorePrice[]>(() => {
  const stores = detail.value?.stores ?? [];
  const discountedStores = supermarketDiscountStorePrices(stores, preferenceStore.locale);
  return (discountedStores.length > 0
    ? discountedStores
    : supermarketCurrentStorePrices(stores, preferenceStore.locale)).slice(0, 8);
});

// 5. 取得最新商店對比卡（對齊參考三列）
const latestStoreCards = computed<SupermarketStorePrice[]>(() => latestStorePrices.value.slice(0, 3));

// 6. 取得最優惠商店
const bestDealPrices = computed<SupermarketStorePrice[]>(() =>
  supermarketBestDealStorePrices(detail.value?.stores ?? [], preferenceStore.locale),
);

// 7. 建立最優惠門店摘要文字
const bestDealSummaryText = computed(() =>
  bestDealPrices.value.length > 0
    ? bestDealPrices.value
        .map((item) => `${formatStore(item.store)} ${formatOfferPrice(item.effectiveUnitPrice)}`)
        .join(' / ')
    : t('offers.detail.noDeal'),
);

// 8. 取得最新快照日期
const latestSnapshotText = computed(() =>
  formatSupermarketDate(latestStoreCards.value[0]?.snapshotDate, preferenceStore.locale)
  || t('offers.detail.latestData'),
);

// 9. 取得同品牌商品
const sameBrandProducts = computed<SupermarketProduct[]>(() => {
  const direct = detail.value?.sameBrand ?? [];
  if (direct.length > 0) {
    return direct.slice(0, 8);
  }

  const brand = product.value?.brand?.trim() ?? '';
  if (!brand) {
    return [];
  }

  return (detail.value?.sameCategory ?? [])
    .filter((item) => item.brand?.trim() === brand && item.code !== product.value?.code)
    .slice(0, 8);
});

// 10. 取得同分類商品
const sameCategoryProducts = computed<SupermarketProduct[]>(() =>
  (detail.value?.sameCategory ?? [])
    .filter((item) => item.code !== product.value?.code)
    .slice(0, 8),
);

// 11. 取得商品頭圖首字
const productInitial = computed(() => {
  const source = product.value?.brand?.trim()
    || product.value?.name?.trim()
    || t('offers.detail.productFallbackInitial');
  return source.slice(0, 1);
});

// 12. 建立走勢圖資料
const trendChart = computed(() => buildTrendChart(detail.value?.history ?? [], chartMode.value));

// 13. 建立走勢圖浮層資料
const trendHover = computed(() => {
  const date = trendHoverDate.value;
  const chart = trendChart.value;
  if (!date || !chart.hasData) {
    return null;
  }

  const items: TrendChartHoverItem[] = chart.series
    .flatMap((series) =>
      series.dots
        .filter((point) => point.date === date)
        .map((point) => ({ ...point, color: series.color })),
    )
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

// 14. 取得關聯商品價格
const relatedProductPriceText = (item: SupermarketProduct): string =>
  formatOfferPrice(supermarketPrimaryPrice(item, preferenceStore.locale).effectiveUnitPrice);

// 15. 取得關聯商品門店
const relatedProductStoreText = (item: SupermarketProduct): string =>
  formatStore(supermarketPrimaryPrice(item, preferenceStore.locale).store);

// 16. 取得關聯商品優惠文字
const relatedProductOfferText = (item: SupermarketProduct): string => {
  const [firstOffer] = supermarketOfferTexts(item, preferenceStore.locale);
  return firstOffer
    || supermarketPrimaryPrice(item, preferenceStore.locale).offer
    || t('offers.detail.regularPrice');
};

// 17. 取得關聯商品品牌首字
const relatedProductInitial = (item: SupermarketProduct): string => {
  const source = item.brand?.trim()
    || item.name?.trim()
    || t('offers.detail.productFallbackInitial');
  return source.slice(0, 1);
};

// 18. 記錄無法載入的商品圖片，改用既有後備顯示。
const handleProductImageError = (code: string): void => {
  failedImageCodes.value = new Set([...failedImageCodes.value, code]);
};

// 19. 載入商品詳情
const loadDetail = async (code: string): Promise<void> => {
  if (!code.trim()) {
    detail.value = null;
    detailError.value = t('offers.detail.noProductCode');
    return;
  }

  detailLoading.value = true;
  detailError.value = '';
  actionMessage.value = '';
  resetNearbyStores();

  try {
    const { data } = await fetchSupermarketProductDetail(code, 90);
    detail.value = data.data;
    chartMode.value = 'effective';
    trendHoverDate.value = '';
    fillAlertForm(data.data.alertRule ?? null);
  } catch {
    detailError.value = t('offers.detail.loadError');
  } finally {
    detailLoading.value = false;
  }
};

// 20. 返回列表
const backToList = async (): Promise<void> => {
  await router.push('/supermarket-offers');
};

// 21. 開啟相關商品
const openRelatedProduct = async (item: SupermarketProduct): Promise<void> => {
  await router.push(`/supermarket-offers/products/${encodeURIComponent(item.code)}`);
};

// 22. 切換收藏
const toggleFavorite = async (): Promise<void> => {
  if (!detail.value) return;
  if (!readStoredAccessToken()) {
    await openLogin();
    return;
  }

  savingFavorite.value = true;
  actionMessage.value = '';
  try {
    if (detail.value.isFavorite) {
      await removeSupermarketFavorite(detail.value.product.code);
      detail.value.isFavorite = false;
      detail.value.product.isFavorite = false;
      actionMessage.value = t('offers.list.favoriteRemoved');
      return;
    }

    await addSupermarketFavorite(detail.value.product.code);
    detail.value.isFavorite = true;
    detail.value.product.isFavorite = true;
    actionMessage.value = t('offers.list.favoriteAdded');
  } catch {
    actionMessage.value = t('offers.list.favoriteError');
  } finally {
    savingFavorite.value = false;
  }
};

// 23. 開啟提醒表單
const openAlertForm = async (): Promise<void> => {
  if (!readStoredAccessToken()) {
    await openLogin();
    return;
  }
  alertFormOpen.value = true;
};

// 24. 儲存價格提示
const submitAlert = async (): Promise<void> => {
  if (!detail.value) return;

  const trimmedPrice = alertForm.targetPrice.trim();
  const targetPrice = trimmedPrice ? Number(trimmedPrice) : undefined;
  if (targetPrice !== undefined && (!Number.isFinite(targetPrice) || targetPrice < 0)) {
    actionMessage.value = t('offers.detail.priceValidation');
    return;
  }

  const payload: {
    productCode: string;
    targetPrice?: number;
    priceMode: PriceChartMode;
    offerRequired: boolean;
    enabled: boolean;
  } = {
    productCode: detail.value.product.code,
    priceMode: alertForm.priceMode,
    offerRequired: alertForm.offerRequired,
    enabled: alertForm.enabled,
  };
  if (targetPrice !== undefined) {
    payload.targetPrice = targetPrice;
  }

  savingAlert.value = true;
  actionMessage.value = '';
  try {
    const { data } = await saveSupermarketPriceAlert(payload);
    detail.value.alertRule = data.data;
    fillAlertForm(data.data);
    alertFormOpen.value = false;
    actionMessage.value = t('offers.detail.alertSaved');
  } catch {
    actionMessage.value = t('offers.detail.alertSaveError');
  } finally {
    savingAlert.value = false;
  }
};

// 25. 刪除價格提示
const removeAlert = async (): Promise<void> => {
  if (!detail.value?.alertRule) return;

  savingAlert.value = true;
  actionMessage.value = '';
  try {
    await deleteSupermarketPriceAlert(detail.value.alertRule.id);
    detail.value.alertRule = null;
    fillAlertForm(null);
    alertFormOpen.value = false;
    actionMessage.value = t('offers.detail.alertDeleted');
  } catch {
    actionMessage.value = t('offers.detail.alertDeleteError');
  } finally {
    savingAlert.value = false;
  }
};

// 26. 用現有提示填入表單
const fillAlertForm = (rule: SupermarketPriceAlert | null): void => {
  alertForm.targetPrice = rule?.targetPrice !== undefined && rule?.targetPrice !== null ? String(rule.targetPrice) : '';
  alertForm.priceMode = rule?.priceMode ?? 'effective';
  alertForm.offerRequired = rule?.offerRequired ?? false;
  alertForm.enabled = rule?.enabled ?? true;
};

// 27. 前往登入
const openLogin = async (): Promise<void> => {
  await router.push({ path: '/login', query: { redirect: route.fullPath } });
};

// 28. 重設附近門店狀態
const resetNearbyStores = (): void => {
  nearbyStores.value = [];
  nearbyStatus.value = 'idle';
  nearbyMessage.value = '';
};

// 29. 查找最優惠附近門店
const requestBestDealLocation = (): void => {
  if (!navigator.geolocation) {
    nearbyStatus.value = 'error';
    nearbyMessage.value = t('offers.detail.geolocationUnsupported');
    return;
  }

  const storeCodes = bestDealPrices.value.map((item) => item.store);
  if (storeCodes.length === 0) {
    nearbyStatus.value = 'ready';
    nearbyMessage.value = t('offers.detail.noDealStore');
    return;
  }

  nearbyStores.value = [];
  nearbyStatus.value = 'requesting';
  nearbyMessage.value = t('offers.detail.locatingPosition');

  navigator.geolocation.getCurrentPosition(
    (position) => {
      nearbyStores.value = storeCodes
        .flatMap((storeCode) =>
          nearestSupermarketStoreLocations(
            position.coords.latitude,
            position.coords.longitude,
            [storeCode],
            SUPERMARKET_NEARBY_STORE_LIMIT,
          ),
        )
        .sort((a, b) => a.distanceKm - b.distanceKm);
      nearbyStatus.value = 'ready';
      nearbyMessage.value = nearbyStores.value.length > 0
        ? t('offers.detail.locationSorted')
        : t('offers.detail.noLocationData');
    },
    (error) => {
      nearbyStores.value = [];
      nearbyStatus.value = 'error';
      nearbyMessage.value = geolocationErrorMessage(error);
    },
    { enableHighAccuracy: true, timeout: 10000, maximumAge: SUPERMARKET_GEOLOCATION_CACHE_MAX_AGE_MS },
  );
};

// 30. 取得定位錯誤文字
const geolocationErrorMessage = (error: GeolocationPositionError): string => {
  if (error.code === error.PERMISSION_DENIED) return t('offers.detail.permissionDenied');
  if (error.code === error.POSITION_UNAVAILABLE) return t('offers.detail.positionUnavailable');
  if (error.code === error.TIMEOUT) return t('offers.detail.locationTimeout');
  return t('offers.detail.locationError');
};

// 31. 按滑鼠位置更新走勢圖浮層
const handleTrendHover = (event: MouseEvent): void => {
  const chart = trendChart.value;
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

// 32. 清除走勢圖浮層
const clearTrendHover = (): void => {
  trendHoverDate.value = '';
};

// 33. 取得圖表價格
const priceForMode = (item: SupermarketDailyStorePrice, mode: PriceChartMode): number =>
  mode === 'effective' ? item.effectiveUnitPrice : item.listPrice;

// 34. 建立走勢圖
const buildTrendChart = (history: SupermarketDailyStorePrice[], mode: PriceChartMode) => {
  const width = 720;
  const height = 280;
  const padding = { top: 22, right: 26, bottom: 44, left: 56 };
  const dates = Array.from(new Set(history.map((item) => item.date))).sort();
  const stores = Array.from(new Set(history.map((item) => item.store))).sort();
  const values = history.map((item) => priceForMode(item, mode));
  const minValue = values.length > 0 ? Math.min(...values) : 0;
  const maxValue = values.length > 0 ? Math.max(...values) : 1;
  const span = Math.max(1, maxValue - minValue);
  const colors = ['#F05A00', '#1A7A3A', '#1A1A1A', '#006973', '#974301', '#6d28d9', '#0f766e', '#a16207'];
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
    series,
    yTicks,
    xTicks,
    hasData: history.length > 0,
  };
};

// 35. 選取日期刻度
const pickDateTicks = (dates: string[], maxTicks: number): string[] => {
  if (dates.length <= maxTicks) return dates;
  const lastIndex = dates.length - 1;
  const indexes = new Set<number>();
  for (let index = 0; index < maxTicks; index += 1) {
    indexes.add(Math.round((index / Math.max(1, maxTicks - 1)) * lastIndex));
  }
  return [...indexes].sort((a, b) => a - b).map((index) => dates[index]).filter(Boolean);
};

// 36. 格式化日期刻度
const formatDateTick = (date: string): string => {
  const parts = date.split('-');
  if (parts.length === 3) {
    return `${parts[1]}/${parts[2]}`;
  }
  return date;
};

// 37. 格式化商店名稱
const formatStore = (value: string): string =>
  displaySupermarketStore(value, preferenceStore.locale);

// 38. 格式化商品價格
const formatOfferPrice = (value: number | null | undefined): string =>
  formatSupermarketHKPrice(value, preferenceStore.locale);

watch(
  productCode,
  (code) => {
    void loadDetail(code);
  },
  { immediate: true },
);
</script>

<template>
  <main class="offers-detail-page">
    <p
      v-if="detailError"
      class="gp-state gp-state-error"
    >
      {{ detailError }}
    </p>
    <p
      v-else-if="detailLoading && !detail"
      class="gp-state"
    >
      {{ t('offers.detail.loading') }}
    </p>

    <template v-if="detail && product">
      <section class="gp-detail-panel">
        <!-- 1. 麵包屑 -->
        <div class="gp-breadcrumb">
          <button
            type="button"
            class="bc-link"
            @click="backToList"
          >
            {{ t('offers.detail.back') }}
          </button>
          <span class="bc-sep">›</span>
          <strong class="bc-current">{{ t('offers.detail.title') }}</strong>
        </div>

        <!-- 2. 商品頭 -->
        <div class="gp-product-head">
          <div class="gp-product-img">
            <img
              v-if="(product.image_url || product.imageUrl) && !failedImageCodes.has(product.code)"
              :src="product.image_url || product.imageUrl"
              :alt="product.name"
              @error="handleProductImageError(product.code)"
            >
            <span
              v-else
              class="gp-product-img-text"
            >
              {{ productInitial }}
            </span>
          </div>
          <div>
            <div class="gp-product-brand">{{ product.brand || t('offers.list.noBrand') }}</div>
            <div class="gp-detail-title">{{ product.name }}</div>
            <div class="gp-product-cat">{{ productCategoryText }}</div>
          </div>
          <div class="gp-detail-actions">
            <button
              type="button"
              class="gp-detail-action"
              :class="detail.isFavorite ? 'on' : ''"
              :disabled="savingFavorite"
              @click="toggleFavorite"
            >
              {{ detail.isFavorite ? t('offers.list.favorited') : t('offers.list.favorite') }}
            </button>
            <button
              type="button"
              class="gp-detail-action"
              @click="openAlertForm"
            >
              {{ t('offers.detail.favoriteWithAlert') }}
            </button>
          </div>
        </div>

        <p
          v-if="actionMessage"
          class="gp-action-message"
        >
          {{ actionMessage }}
        </p>

        <!-- 3. 到價提醒表單 -->
        <section
          v-if="alertFormOpen"
          class="gp-detail-card gp-alert-card"
        >
          <div class="gp-alert-head">
            <div>
              <h3>{{ t('offers.detail.priceAlert') }}</h3>
              <div class="gp-detail-subtitle">{{ t('offers.detail.alertMemberHint') }}</div>
            </div>
            <div class="gp-alert-actions">
              <button
                v-if="detail.alertRule"
                type="button"
                class="gp-link-btn"
                :disabled="savingAlert"
                @click="removeAlert"
              >
                {{ t('offers.detail.deleteAlert') }}
              </button>
              <button
                type="button"
                class="gp-link-btn"
                @click="alertFormOpen = false"
              >
                {{ t('offers.detail.close') }}
              </button>
            </div>
          </div>
          <form
            class="gp-alert-form"
            @submit.prevent="submitAlert"
          >
            <input
              v-model="alertForm.targetPrice"
              type="number"
              min="0"
              step="0.1"
              :placeholder="t('offers.detail.targetPricePlaceholder')"
            >
            <select v-model="alertForm.priceMode">
              <option value="effective">{{ t('offers.detail.effectiveEquivalent') }}</option>
              <option value="list">{{ t('offers.detail.originalPrice') }}</option>
            </select>
            <label>
              <input
                v-model="alertForm.offerRequired"
                type="checkbox"
              >
              {{ t('offers.detail.alertOfferOnly') }}
            </label>
            <label>
              <input
                v-model="alertForm.enabled"
                type="checkbox"
              >
              {{ t('offers.detail.enableAlert') }}
            </label>
            <button
              type="submit"
              :disabled="savingAlert"
            >
              {{ savingAlert ? t('offers.detail.saving') : t('offers.detail.saveAlert') }}
            </button>
          </form>
        </section>

        <!-- 4. 超市優惠與門店地址 -->
        <div class="gp-detail-card gp-section-block">
          <h3>{{ t('offers.detail.offersAndStores') }}</h3>
          <div class="gp-detail-subtitle">
            {{ t('offers.detail.offersAndStoresHint') }}
          </div>
          <div class="gp-compare-grid">
            <div
              v-for="price in latestStoreCards"
              :key="`${price.store}-${price.effectiveUnitPrice}`"
              class="gp-compare-card"
            >
              <div class="gp-compare-store">{{ formatStore(price.store) }}</div>
              <div class="gp-compare-price">{{ formatOfferPrice(price.effectiveUnitPrice) }}</div>
              <div class="gp-compare-origin">{{ t('offers.detail.originalPriceValue', {
                price: formatOfferPrice(price.listPrice),
              }) }}</div>
              <div
                v-if="price.offer"
                class="gp-compare-offer"
              >
                {{ price.offer }}
              </div>
              <div class="gp-compare-badge">-{{ supermarketPriceDiscountRate(price).toFixed(0) }}%</div>
            </div>
          </div>
          <div class="gp-best-shop">
            <div>
              <strong>{{ t('offers.detail.bestStoreAddress') }}</strong>
              <span>{{ bestDealSummaryText }}</span>
            </div>
            <button
              type="button"
              class="gp-map-btn"
              :disabled="nearbyStatus === 'requesting'"
              @click="requestBestDealLocation"
            >
              {{ nearbyStatus === 'requesting'
                ? t('offers.detail.locating')
                : t('offers.detail.findBestStore') }}
            </button>
          </div>
          <h3 class="gp-shop-block-title">{{ t('offers.detail.nearestBestStoreAddress') }}</h3>
          <div class="gp-detail-subtitle">{{ nearbyMessage || t('offers.detail.latestSnapshot', {
            date: latestSnapshotText,
          }) }}</div>
          <div
            v-if="nearbyStores.length > 0"
            class="gp-shop-table-wrap"
          >
            <table class="gp-shop-table">
              <thead>
                <tr>
                  <th>#</th>
                  <th>{{ t('offers.detail.supermarket') }}</th>
                  <th>{{ t('offers.detail.shop') }}</th>
                  <th>{{ t('offers.detail.address') }}</th>
                  <th>{{ t('offers.detail.distance') }}</th>
                  <th>{{ t('offers.detail.openingHours') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="(store, index) in nearbyStores"
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
                  <td>{{ formatStore(store.storeCode) }}</td>
                  <td>{{ store.name }}</td>
                  <td>{{ store.address }}</td>
                  <td><span class="gp-shop-dist">{{ formatSupermarketDistanceKm(store.distanceKm) }}</span></td>
                  <td>{{ store.hours || '-' }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <!-- 5. 價格摘要 + 走勢 -->
        <div class="gp-price-layout">
          <div class="gp-detail-card">
            <h3>{{ t('offers.detail.summary90Days') }}</h3>
            <div class="gp-summary-grid">
              <div class="gp-summary-item">
                <div class="gp-summary-label">{{ t('offers.detail.lowestList90Days') }}</div>
                <div class="gp-summary-value">{{ formatOfferPrice(detail.summary.lowestList) }}</div>
              </div>
              <div class="gp-summary-item">
                <div class="gp-summary-label">{{ t('offers.detail.highestList90Days') }}</div>
                <div class="gp-summary-value">{{ formatOfferPrice(detail.summary.highestList) }}</div>
              </div>
              <div class="gp-summary-item">
                <div class="gp-summary-label">{{ t('offers.detail.lowestDeal90Days') }}</div>
                <div class="gp-summary-value">{{ formatOfferPrice(detail.summary.lowestDeal) }}</div>
              </div>
              <div class="gp-summary-item">
                <div class="gp-summary-label">{{ t('offers.detail.highestDeal90Days') }}</div>
                <div class="gp-summary-value">{{ formatOfferPrice(detail.summary.highestDeal) }}</div>
              </div>
            </div>
          </div>
          <div class="gp-detail-card">
            <h3>{{ t('offers.detail.trend90Days') }}</h3>
            <div class="gp-detail-subtitle">{{ chartMode === 'effective'
              ? t('offers.detail.effectiveEquivalent')
              : t('offers.detail.originalPrice') }}</div>
            <div class="gp-chart-mode">
              <button
                type="button"
                class="gp-mode-pill"
                :class="chartMode === 'effective' ? 'on' : ''"
                @click="chartMode = 'effective'"
              >
                {{ t('offers.detail.dealPrice') }}
              </button>
              <button
                type="button"
                class="gp-mode-pill"
                :class="chartMode === 'list' ? 'on' : ''"
                @click="chartMode = 'list'"
              >
                {{ t('offers.detail.originalPrice') }}
              </button>
            </div>

            <div
              v-if="trendChart.hasData"
              class="gp-chart"
            >
              <svg
                :viewBox="`0 0 ${trendChart.width} ${trendChart.height}`"
                preserveAspectRatio="none"
                role="img"
                :aria-label="t('offers.detail.trendAria')"
                @mousemove="handleTrendHover"
                @mouseleave="clearTrendHover"
              >
                <rect
                  x="0"
                  y="0"
                  :width="trendChart.width"
                  :height="trendChart.height"
                />
                <g
                  v-for="tick in trendChart.yTicks"
                  :key="tick.key"
                >
                  <line
                    :x1="trendChart.padding.left"
                    :x2="trendChart.width - trendChart.padding.right"
                    :y1="tick.y"
                    :y2="tick.y"
                  />
                  <text
                    :x="trendChart.padding.left - 8"
                    :y="tick.y + 4"
                    text-anchor="end"
                  >
                    {{ tick.label }}
                  </text>
                </g>
                <g
                  v-for="tick in trendChart.xTicks"
                  :key="tick.key"
                >
                  <line
                    :x1="tick.x"
                    :x2="tick.x"
                    :y1="trendChart.padding.top"
                    :y2="trendChart.height - trendChart.padding.bottom"
                  />
                  <text
                    :x="tick.x"
                    :y="trendChart.height - 16"
                    text-anchor="middle"
                  >
                    {{ tick.label }}
                  </text>
                </g>
                <line
                  v-if="trendHover"
                  class="gp-hover-line"
                  :x1="trendHover.x"
                  :x2="trendHover.x"
                  :y1="trendChart.padding.top"
                  :y2="trendChart.height - trendChart.padding.bottom"
                />
                <polyline
                  v-for="series in trendChart.series"
                  :key="series.store"
                  :points="series.points"
                  :stroke="series.color"
                />
                <template
                  v-for="series in trendChart.series"
                  :key="`${series.store}-dots`"
                >
                  <circle
                    v-for="point in series.dots"
                    :key="point.key"
                    :cx="point.x"
                    :cy="point.y"
                    r="3"
                    :fill="series.color"
                  />
                </template>
              </svg>
              <div
                v-if="trendHover"
                class="gp-tooltip"
                :style="{ left: trendHover.tooltipLeft }"
              >
                <strong>{{ formatDateTick(trendHover.date) }}</strong>
                <span
                  v-for="item in trendHover.items"
                  :key="`${item.key}-tooltip`"
                >
                  {{ formatStore(item.store) }} {{ formatOfferPrice(item.value) }}
                </span>
              </div>
              <div class="gp-chart-legend">
                <span
                  v-for="series in trendChart.series"
                  :key="`${series.store}-legend`"
                >
                  <i
                    class="gp-legend-dot"
                    :style="{ background: series.color }"
                  />
                  {{ formatStore(series.store) }}
                </span>
              </div>
            </div>
            <p
              v-else
              class="gp-state"
            >
              {{ t('offers.detail.noHistory') }}
            </p>
          </div>
        </div>

        <!-- 6. 同品牌產品 -->
        <div
          v-if="sameBrandProducts.length > 0"
          class="gp-detail-card gp-related-block"
        >
          <h3>{{ t('offers.detail.sameBrand') }}</h3>
          <div class="gp-related-scroll">
            <button
              v-for="item in sameBrandProducts"
              :key="item.code"
              type="button"
              class="gp-related-card"
              @click="openRelatedProduct(item)"
            >
              <div class="gp-related-img">
                <img
                  v-if="(item.image_url || item.imageUrl) && !failedImageCodes.has(item.code)"
                  :src="item.image_url || item.imageUrl"
                  :alt="item.name"
                  @error="handleProductImageError(item.code)"
                >
                <span
                  v-else
                  class="gp-related-img-text"
                >
                  {{ relatedProductInitial(item) }}
                </span>
              </div>
              <div class="gp-related-body">
                <div class="gp-related-brand">{{ item.brand || t('offers.list.noBrand') }}</div>
                <div class="gp-related-name">{{ item.name }}</div>
                <div class="gp-related-offer">{{ relatedProductStoreText(item) }} {{ relatedProductOfferText(item) }}</div>
                <div class="gp-related-price">{{ relatedProductPriceText(item) }} {{ t('offers.detail.perItem') }}</div>
              </div>
            </button>
          </div>
        </div>

        <!-- 7. 同分類產品 -->
        <div
          v-if="sameCategoryProducts.length > 0"
          class="gp-detail-card gp-related-block"
        >
          <h3>{{ t('offers.detail.sameCategory') }}</h3>
          <div class="gp-related-scroll">
            <button
              v-for="item in sameCategoryProducts"
              :key="item.code"
              type="button"
              class="gp-related-card"
              @click="openRelatedProduct(item)"
            >
              <div class="gp-related-img">
                <img
                  v-if="(item.image_url || item.imageUrl) && !failedImageCodes.has(item.code)"
                  :src="item.image_url || item.imageUrl"
                  :alt="item.name"
                  @error="handleProductImageError(item.code)"
                >
                <span
                  v-else
                  class="gp-related-img-text"
                >
                  {{ relatedProductInitial(item) }}
                </span>
              </div>
              <div class="gp-related-body">
                <div class="gp-related-brand">{{ item.brand || t('offers.list.noBrand') }}</div>
                <div class="gp-related-name">{{ item.name }}</div>
                <div class="gp-related-offer">{{ relatedProductStoreText(item) }} {{ relatedProductOfferText(item) }}</div>
                <div class="gp-related-price">{{ relatedProductPriceText(item) }} {{ t('offers.detail.perItem') }}</div>
              </div>
            </button>
          </div>
        </div>
      </section>
    </template>
  </main>
</template>

<style scoped>
/* 1. 頁面容器 */
.offers-detail-page {
  width: 100%;
  min-height: calc(100svh - 48px);
  background: var(--sur);
  color: var(--ink);
  padding: 20px 0 44px;
}

/* 2. 詳情面板 */
.gp-detail-panel {
  max-width: var(--layout-page-max-width);
  margin: 0 auto;
  border: 1px solid var(--bdr);
  border-radius: 8px;
  background: var(--sur-2);
  padding: 0;
}

/* 3. 麵包屑 */
.gp-breadcrumb {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
  padding: 12px 24px;
  margin: 0;
  background: var(--sur);
  border-bottom: 1px solid var(--bdr);
  color: var(--ink-3);
  font-size: 12px;
  line-height: 1.5;
}

.gp-breadcrumb .bc-link {
  border: 0;
  background: transparent;
  color: var(--ink-3);
  cursor: pointer;
  font-family: inherit;
  font-size: 12px;
  padding: 0;
}

.gp-breadcrumb .bc-link:hover {
  color: var(--brand);
}

.gp-breadcrumb .bc-sep {
  color: var(--ink-4);
  font-size: 12px;
}

.gp-breadcrumb .bc-current {
  color: var(--ink);
  font-weight: 600;
}

/* 4. 商品頭 */
.gp-product-head {
  display: grid;
  grid-template-columns: 138px minmax(0, 1fr) auto;
  gap: 18px;
  align-items: center;
  margin: 0;
  border-bottom: 1px solid var(--bdr);
  background: var(--sur);
  padding: 18px 20px;
}

.gp-product-img {
  display: flex;
  width: 138px;
  height: 118px;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--bdr);
  border-radius: 8px;
  background: linear-gradient(135deg, #fff 0%, var(--sur-2) 100%);
  overflow: hidden;
}

.gp-product-img img {
  width: auto;
  height: auto;
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
  display: block;
}

.gp-product-img-text {
  color: var(--brand);
  font-family: var(--font-serif);
  font-size: 36px;
  font-weight: 700;
}

.gp-product-brand {
  margin-bottom: 6px;
  color: var(--ink-3);
  font-size: 12px;
  font-weight: 700;
}

.gp-detail-title {
  color: var(--ink);
  font-size: 24px;
  font-weight: 700;
  line-height: 1.25;
}

.gp-product-cat {
  margin-top: 8px;
  color: var(--ink-3);
  font-size: 13px;
}

.gp-detail-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  justify-content: flex-end;
}

.gp-detail-action {
  border: 1px solid var(--brand);
  border-radius: 6px;
  background: var(--brand-light);
  color: var(--brand);
  cursor: pointer;
  font-family: inherit;
  font-size: 12px;
  font-weight: 700;
  padding: 9px 12px;
}

.gp-detail-action.on {
  background: var(--brand);
  color: #fff;
}

.gp-detail-action:disabled {
  cursor: not-allowed;
  opacity: 0.52;
}

/* 5. 操作訊息 */
.gp-action-message {
  margin: 12px 20px 0;
  border: 1px solid var(--bdr);
  border-radius: 6px;
  background: var(--sur);
  color: var(--ink-3);
  font-size: 12px;
  padding: 10px 12px;
}

/* 6. 通用卡片 */
.gp-detail-card {
  border: 1px solid var(--bdr);
  border-radius: 8px;
  background: var(--sur);
  padding: 16px;
}

.gp-detail-card h3 {
  margin: 0 0 10px;
  color: var(--ink);
  font-size: 14px;
  font-weight: 700;
}

.gp-detail-subtitle {
  margin: 6px 0 14px;
  color: var(--ink-3);
  font-size: 12px;
  line-height: 1.6;
}

.gp-section-block {
  margin: 16px 20px;
}

/* 7. 到價提醒表單 */
.gp-alert-card {
  margin: 16px 20px;
}

.gp-alert-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
}

.gp-alert-head h3 {
  margin: 0 0 4px;
}

.gp-alert-actions {
  display: flex;
  gap: 8px;
}

.gp-link-btn {
  border: 1px solid var(--bdr);
  border-radius: 6px;
  background: var(--sur);
  color: var(--brand);
  cursor: pointer;
  font-family: inherit;
  font-size: 12px;
  font-weight: 700;
  padding: 7px 10px;
}

.gp-alert-form {
  display: grid;
  grid-template-columns: minmax(160px, 1fr) minmax(160px, 0.6fr) auto auto auto;
  gap: 10px;
  align-items: center;
}

.gp-alert-form input[type='number'],
.gp-alert-form select {
  border: 1px solid var(--bdr);
  border-radius: 6px;
  color: var(--ink);
  font: inherit;
  font-size: 12px;
  padding: 9px 10px;
}

.gp-alert-form label {
  display: flex;
  align-items: center;
  gap: 6px;
  color: var(--ink-3);
  font-size: 12px;
}

.gp-alert-form button {
  border: 1px solid var(--brand);
  border-radius: 6px;
  background: var(--brand);
  color: #fff;
  cursor: pointer;
  font-family: inherit;
  font-size: 12px;
  font-weight: 700;
  padding: 9px 14px;
}

.gp-alert-form button:disabled {
  cursor: not-allowed;
  opacity: 0.52;
}

/* 8. 對比卡 */
.gp-compare-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
  padding: 12px;
  border: 1px solid #b7c9dc;
  border-radius: 6px;
  background: #e8f4ff;
}

.gp-compare-card {
  position: relative;
  border: 1px solid var(--bdr);
  border-radius: 8px;
  background: var(--sur-2);
  padding: 12px;
}

.gp-compare-store {
  color: var(--ink);
  font-size: 13px;
  font-weight: 800;
  padding-right: 46px;
}

.gp-compare-price {
  margin-top: 8px;
  color: var(--brand);
  font-size: 22px;
  font-weight: 800;
  line-height: 1;
}

.gp-compare-origin {
  margin-top: 6px;
  color: var(--ink-3);
  font-size: 11px;
  text-decoration: line-through;
}

.gp-compare-offer {
  display: inline-block;
  margin-top: 8px;
  background: #ffd8c8;
  color: #9a3b12;
  font-size: 12px;
  font-weight: 800;
  line-height: 1.45;
  padding: 5px 7px;
  border-radius: 2px;
}

.gp-compare-badge {
  position: absolute;
  top: 10px;
  right: 10px;
  border-radius: 2px;
  background: var(--brand);
  color: #fff;
  font-size: 11px;
  font-weight: 800;
  line-height: 1;
  padding: 5px 7px;
}

/* 9. 最優惠門店 */
.gp-best-shop {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 12px;
  align-items: center;
  border: 1px solid var(--bdr);
  border-radius: 8px;
  background: var(--brand-light);
  padding: 13px 14px;
  margin-top: 14px;
}

.gp-best-shop strong {
  display: block;
  color: var(--ink);
  font-size: 13px;
  font-weight: 700;
}

.gp-best-shop span {
  display: block;
  margin-top: 4px;
  color: var(--ink-3);
  font-size: 12px;
}

.gp-map-btn {
  border: 1px solid var(--brand);
  border-radius: 6px;
  background: var(--sur);
  color: var(--brand);
  cursor: pointer;
  font-family: inherit;
  font-size: 12px;
  font-weight: 800;
  padding: 8px 11px;
  white-space: nowrap;
}

.gp-map-btn:disabled {
  cursor: not-allowed;
  opacity: 0.52;
}

/* 10. 門店表格 */
.gp-shop-block-title {
  margin-top: 18px !important;
}

.gp-shop-table-wrap {
  overflow: auto;
  border: 1px solid var(--bdr);
  border-radius: 8px;
}

.gp-shop-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 12px;
  min-width: 760px;
}

.gp-shop-table th {
  border-bottom: 1px solid var(--bdr);
  background: var(--sur-2);
  color: var(--ink-3);
  font-weight: 700;
  padding: 10px;
  text-align: left;
}

.gp-shop-table td {
  border-bottom: 1px solid var(--sur-3);
  color: var(--ink-2);
  padding: 10px;
  vertical-align: top;
}

.gp-shop-table tr:last-child td {
  border-bottom: 0;
}

.gp-shop-table a {
  color: var(--brand);
  font-weight: 800;
  text-decoration: none;
}

.gp-shop-dist {
  display: inline-flex;
  border-radius: 2px;
  background: #ffd8c8;
  color: #9a3b12;
  font-weight: 800;
  padding: 4px 7px;
}

/* 11. 價格摘要 + 走勢佈局 */
.gp-price-layout {
  display: grid;
  grid-template-columns: 330px minmax(0, 1fr);
  gap: 16px;
  margin: 0 20px 16px;
}

.gp-summary-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 10px;
}

.gp-summary-item {
  border: 1px solid var(--bdr);
  border-radius: 8px;
  background: var(--sur-2);
  padding: 12px;
}

.gp-summary-label {
  color: var(--ink-3);
  font-size: 11px;
  font-weight: 700;
}

.gp-summary-value {
  margin-top: 6px;
  color: var(--ink);
  font-size: 18px;
  font-weight: 800;
}

/* 12. 走勢圖 */
.gp-chart-mode {
  display: flex;
  gap: 6px;
  margin-bottom: 10px;
}

.gp-mode-pill {
  border: 1px solid var(--bdr);
  border-radius: 999px;
  background: var(--sur);
  color: var(--ink-3);
  cursor: pointer;
  font-family: inherit;
  font-size: 11px;
  font-weight: 800;
  padding: 6px 10px;
}

.gp-mode-pill.on {
  border-color: var(--brand);
  background: var(--brand-light);
  color: var(--brand);
}

.gp-chart {
  position: relative;
  height: 280px;
  border: 1px solid var(--bdr);
  border-radius: 8px;
  background: linear-gradient(180deg, #fff, #fafafa);
  padding: 12px;
}

.gp-chart svg {
  display: block;
  width: 100%;
  height: 100%;
}

.gp-chart rect {
  fill: transparent;
}

.gp-chart line {
  stroke: var(--sur-3);
  stroke-width: 1;
}

.gp-chart text {
  fill: var(--ink-3);
  font-size: 11px;
}

.gp-chart polyline {
  fill: none;
  stroke-width: 2.4;
}

.gp-hover-line {
  stroke: var(--brand);
  stroke-width: 1.4;
}

.gp-tooltip {
  position: absolute;
  top: 24px;
  z-index: 2;
  min-width: 180px;
  transform: translateX(-50%);
  border: 1px solid var(--bdr);
  border-radius: 6px;
  background: var(--sur);
  box-shadow: 0 10px 28px rgba(0, 0, 0, 0.08);
  padding: 10px;
}

.gp-tooltip strong,
.gp-tooltip span {
  display: block;
  color: var(--ink);
  font-size: 12px;
}

.gp-tooltip span {
  margin-top: 5px;
  color: var(--ink-3);
  font-size: 11px;
}

.gp-chart-legend {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
  margin-top: 10px;
  color: var(--ink-3);
  font-size: 11px;
}

.gp-chart-legend span {
  display: inline-flex;
  align-items: center;
}

.gp-legend-dot {
  display: inline-flex;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  margin-right: 5px;
}

/* 13. 同品牌 / 同分類 */
.gp-related-block {
  margin: 0 20px 16px;
}

.gp-related-scroll {
  display: flex;
  gap: 10px;
  overflow-x: auto;
  padding-bottom: 4px;
  scrollbar-width: thin;
}

.gp-related-card {
  flex: 0 0 220px;
  display: block;
  border: 1px solid var(--bdr);
  border-radius: 8px;
  background: var(--sur);
  overflow: hidden;
  cursor: pointer;
  font: inherit;
  padding: 0;
  text-align: left;
  color: var(--ink);
}

.gp-related-img {
  display: flex;
  height: 96px;
  align-items: center;
  justify-content: center;
  border-bottom: 1px solid var(--bdr);
  background: linear-gradient(135deg, #fff 0%, var(--sur-2) 100%);
  overflow: hidden;
}

.gp-related-img img {
  width: auto;
  height: auto;
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
  display: block;
}

.gp-related-img-text {
  color: var(--brand);
  font-family: var(--font-serif);
  font-size: 28px;
  font-weight: 700;
}

.gp-related-body {
  padding: 11px;
}

.gp-related-brand {
  color: var(--ink-3);
  font-size: 11px;
  font-weight: 700;
}

.gp-related-name {
  margin-top: 4px;
  color: var(--ink);
  font-size: 13px;
  font-weight: 800;
  line-height: 1.35;
}

.gp-related-offer {
  margin-top: 8px;
  color: var(--ink-3);
  font-size: 11px;
  line-height: 1.4;
}

.gp-related-price {
  margin-top: 6px;
  color: var(--brand);
  font-size: 13px;
  font-weight: 800;
}

/* 14. 狀態訊息 */
.gp-state {
  max-width: var(--layout-page-max-width);
  margin: 20px auto 0;
  border: 1px solid var(--bdr);
  border-radius: 8px;
  background: var(--sur);
  color: var(--ink-3);
  font-size: 13px;
  padding: 12px 14px;
}

.gp-state-error {
  border-color: rgba(186, 26, 26, 0.24);
  color: #ba1a1a;
}

/* 15. 響應式 */
@media (max-width: 1023px) {
  .offers-detail-page {
    padding: 14px 14px 48px;
  }

  .gp-product-head {
    grid-template-columns: 1fr;
  }

  .gp-product-img {
    width: 100%;
    height: 180px;
  }

  .gp-detail-actions {
    justify-content: flex-start;
  }

  .gp-compare-grid,
  .gp-summary-grid {
    grid-template-columns: 1fr;
  }

  .gp-best-shop {
    grid-template-columns: 1fr;
  }

  .gp-price-layout {
    grid-template-columns: 1fr;
    margin-left: 14px;
    margin-right: 14px;
  }

  .gp-section-block,
  .gp-detail-panel > .gp-detail-card,
  .gp-alert-card,
  .gp-action-message,
  .gp-related-block {
    margin-left: 14px;
    margin-right: 14px;
  }

  .gp-alert-form {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 560px) {
  .gp-detail-title {
    font-size: 20px;
  }
}
</style>

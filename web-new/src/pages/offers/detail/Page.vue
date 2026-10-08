<!--
 * 綜合優惠商品詳情頁。
 * 1. 使用 AJO 後端商品詳情接口展示即時價格與 90 日走勢。
 * 2. 支援會員收藏、到價提示與最優惠商戶附近門店查找。
 * 3. 對齊 docs 高保真參考的 gp-detail-panel 結構與樣式。
-->
<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import {
  addSupermarketFavorite,
  deleteSupermarketPriceAlert,
  fetchSupermarketProductDetail,
  removeSupermarketFavorite,
  reportSupermarketProductImage,
  saveSupermarketPriceAlert,
} from '@/httpapis/supermarket-offers';
import { readStoredAccessToken } from '@/httpapis/auth-session';
import AppIcon from '@/shared/components/base/AppIcon.vue';
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
  resolveSupermarketProductBrand,
  resolveSupermarketProductFullTitle,
  resolveSupermarketProductName,
  resolveSupermarketProductUnit,
  supermarketBestDealStorePrices,
  supermarketCurrentStorePrices,
  supermarketOfferTexts,
  supermarketIsSimpleOffer,
  supermarketOfferDisplayText,
  supermarketPriceDiscountRate,
  normalizeSupermarketStoreCode,
  supermarketSecondPriceAdvantageRate,
  supermarketSecondPriceAdvantageStores,
  supermarketPrimaryPrice,
  supermarketStorePrices,
} from '@/utils/supermarket-offers';
import {
  formatSupermarketDistanceKm,
  nearestSupermarketStoreLocations,
  SUPERMARKET_GEOLOCATION_CACHE_MAX_AGE_MS,
  SUPERMARKET_NEARBY_STORE_LIMIT,
  supermarketDirectionsUrl,
  supermarketMapEmbedUrl,
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
const selectedNearbyStoreId = ref('');
const nearbyStatus = ref<NearbyStoreStatus>('idle');
const nearbyMessage = ref('');
const failedImageCodes = ref(new Set<string>());
const isImagePreviewOpen = ref(false);
const reportingImageIssue = ref(false);
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

// 3. 取得可預覽的商品圖片網址。
const productImageUrl = computed(() => product.value?.image_url || product.value?.imageUrl || '');

// 4. 建立可返回列表篩選的商品分類。
const productCategories = computed(() =>
  [product.value?.category1, product.value?.category2, product.value?.category3]
    .filter((value): value is string => Boolean(value?.trim()))
    .map((value) => ({ value, label: displaySupermarketCategory(value, preferenceStore.locale) })),
);

// 4. 取得最新商店價格
const latestStorePrices = computed<SupermarketStorePrice[]>(() => {
  const pricesByStore = new Map<string, SupermarketStorePrice>();
  const productPrices = product.value
    ? supermarketStorePrices(product.value, preferenceStore.locale)
    : [];
  [...productPrices, ...(detail.value?.stores ?? [])].forEach((price) => {
    const storeCode = price.store.trim().toUpperCase();
    if (storeCode) {
      pricesByStore.set(storeCode, price);
    }
  });
  return supermarketCurrentStorePrices([...pricesByStore.values()], preferenceStore.locale);
});

// 5. 取得返回列表所需的原搜尋參數。
const returnListQuery = computed(() => {
  const { return: _return, ...query } = route.query;
  return query;
});

// 6. 取得所有最新商店對比卡，避免只顯示單一商店。
const latestStoreCards = computed<SupermarketStorePrice[]>(() => latestStorePrices.value);

// 6.1 取得可顯示第二低價優勢的最低價門店代碼。
const secondPriceAdvantageStoreCodes = computed(() => new Set(
  supermarketSecondPriceAdvantageStores(latestStoreCards.value, preferenceStore.locale)
    .map((item) => normalizeSupermarketStoreCode(item.store)),
));

// 6.2 取得最低價相對下一個不同價格的優勢百分比。
const secondPriceAdvantageRate = computed<number>(() =>
  supermarketSecondPriceAdvantageRate(latestStoreCards.value),
);

// 7. 取得純減價或原價貨品的 90 日價格標籤。
const simpleOfferTrendLabel = (price: SupermarketStorePrice): string => {
  if (!supermarketIsSimpleOffer(price.offer)) {
    return '';
  }

  const history = detail.value?.history ?? [];
  const historyDates = [...new Set(history.map((item) => item.date))].sort();
  const today = price.snapshotDate || historyDates[historyDates.length - 1] || '';
  const previousDates = [...new Set(
    history
      .filter((item) => item.store.trim().toUpperCase() === price.store.trim().toUpperCase() && item.date < today)
      .map((item) => item.date),
  )].sort().slice(-89);
  const previousPrices = history
    .filter((item) => item.store.trim().toUpperCase() === price.store.trim().toUpperCase())
    .filter((item) => previousDates.includes(item.date))
    .map((item) => item.effectiveUnitPrice)
    .filter((value) => Number.isFinite(value) && value >= 0);
  if (previousPrices.length === 0) {
    return '';
  }

  const lowestPreviousPrice = Math.min(...previousPrices);
  if (price.effectiveUnitPrice > lowestPreviousPrice * 1.05) {
    return t('offers.detail.nearThreeMonthLow');
  }
  if (price.effectiveUnitPrice < lowestPreviousPrice) {
    return t('offers.detail.nearThreeMonthLowest');
  }
  return '';
};

// 8. 顯示非純減價及非原價貨品的有效優惠文字。
const priceOfferDisplayText = (price: SupermarketStorePrice): string =>
  supermarketOfferDisplayText(price.offer);

// 9. 取得最優惠商店
const bestDealPrices = computed<SupermarketStorePrice[]>(() =>
  supermarketBestDealStorePrices(latestStorePrices.value, preferenceStore.locale),
);

// 9.1 判斷門店是否屬於目前最低實付價。
const isBestDealStore = (price: SupermarketStorePrice): boolean =>
  bestDealPrices.value.some((item) =>
    normalizeSupermarketStoreCode(item.store) === normalizeSupermarketStoreCode(price.store),
  );

// 9.2 取得首屏最低價的優惠說明。
const bestDealOfferText = computed(() => {
  const price = bestDealPrices.value[0];
  if (!price) {
    return '';
  }
  return simpleOfferTrendLabel(price) || priceOfferDisplayText(price) || t('offers.list.currentPrice');
});

// 9.3 取得首屏最低價標籤。
const bestDealBadgeText = computed(() => {
  const price = bestDealPrices.value[0];
  if (!price) {
    return '';
  }
  const label = bestDealPrices.value.length > 1
    ? t('offers.detail.tiedLowest')
    : t('offers.list.lowestAcrossStores');
  return priceOfferDisplayText(price)
    ? `${label} · ${t('offers.list.offerApplied')}`
    : label;
});

// 9.4 取得最近的一間最優惠門市。
const nearestStore = computed(() => nearbyStores.value[0] ?? null);

// 9.5 取得目前選擇的門市。
const selectedNearbyStore = computed(() =>
  nearbyStores.value.find((store) => store.id === selectedNearbyStoreId.value) ?? nearestStore.value,
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

// 11. 取得目前商品可切換的同系列資料。
const seriesProducts = computed<SupermarketProduct[]>(() => {
  const currentSeriesID = product.value?.seriesId?.trim();
  if (!currentSeriesID) {
    return [];
  }

  return [product.value, ...(detail.value?.sameSeries ?? [])]
    .filter((item): item is SupermarketProduct => Boolean(item))
    .filter((item) => item.seriesId === currentSeriesID && item.variantLabels)
    .filter((item, index, items) => items.findIndex((candidate) => candidate.code === item.code) === index);
});

// 12. 取得目前商品的同系列屬性值。
const productVariantLabel = (item: SupermarketProduct, key: 'formula' | 'package'): string =>
  item.variantLabels?.[key]?.trim() ?? '';

// 13. 依目前另一個屬性篩選真正可切換的同系列商品。
const seriesOptionsFor = (key: 'formula' | 'package'): SupermarketProduct[] => {
  const current = product.value;
  if (!current) {
    return [];
  }

  const companionKey = key === 'formula' ? 'package' : 'formula';
  const companionValue = productVariantLabel(current, companionKey);
  const uniqueOptions = new Map<string, SupermarketProduct>();
  seriesProducts.value.forEach((item) => {
    const value = productVariantLabel(item, key);
    if (!value || (companionValue && productVariantLabel(item, companionKey) !== companionValue)) {
      return;
    }
    if (!uniqueOptions.has(value)) {
      uniqueOptions.set(value, item);
    }
  });
  return [...uniqueOptions.values()];
};

// 14. 取得相同包裝下可切換的口味或配方。
const seriesFormulaOptions = computed<SupermarketProduct[]>(() => seriesOptionsFor('formula'));

// 15. 取得相同口味或配方下可切換的包裝規格。
const seriesPackageOptions = computed<SupermarketProduct[]>(() => seriesOptionsFor('package'));

// 16. 判斷是否有值得顯示的同系列選擇。
const hasSeriesOptions = computed(() =>
  seriesFormulaOptions.value.length > 1 || seriesPackageOptions.value.length > 1,
);

// 17. 取得商品頭圖首字
const productInitial = computed(() => {
  const source = product.value?.brand?.trim()
    || product.value?.name?.trim()
    || t('offers.detail.productFallbackInitial');
  return source.slice(0, 1);
});

// 18. 建立走勢圖資料
const trendChart = computed(() => buildTrendChart(detail.value?.history ?? [], chartMode.value));

// 19. 建立走勢圖浮層資料
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

// 20. 取得關聯商品價格
const relatedProductPriceText = (item: SupermarketProduct): string =>
  formatOfferPrice(supermarketPrimaryPrice(item, preferenceStore.locale).effectiveUnitPrice);

// 20.1 顯示關聯商品節省百分比
const relatedProductSavingText = (item: SupermarketProduct): string => {
  const rate = supermarketPriceDiscountRate(supermarketPrimaryPrice(item, preferenceStore.locale));
  return rate >= 1
    ? t('offers.detail.savedPercentage', { rate: rate.toFixed(0) })
    : '';
};

// 21. 取得關聯商品門店
const relatedProductStoreText = (item: SupermarketProduct): string =>
  formatStore(supermarketPrimaryPrice(item, preferenceStore.locale).store);

// 22. 取得關聯商品優惠文字
const relatedProductOfferText = (item: SupermarketProduct): string => {
  const [firstOffer] = supermarketOfferTexts(item, preferenceStore.locale);
  return firstOffer
    || supermarketPrimaryPrice(item, preferenceStore.locale).offer
    || t('offers.detail.regularPrice');
};

// 23. 取得關聯商品品牌首字
const relatedProductInitial = (item: SupermarketProduct): string => {
  const source = item.brand?.trim()
    || item.name?.trim()
    || t('offers.detail.productFallbackInitial');
  return source.slice(0, 1);
};

// 24. 記錄無法載入的商品圖片，改用既有後備顯示。
const handleProductImageError = (code: string): void => {
  failedImageCodes.value = new Set([...failedImageCodes.value, code]);
};

// 25. 開啟商品圖片預覽。
const openImagePreview = (): void => {
  if (productImageUrl.value && !failedImageCodes.value.has(productCode.value)) {
    isImagePreviewOpen.value = true;
  }
};

// 26. 關閉商品圖片預覽。
const closeImagePreview = (): void => {
  isImagePreviewOpen.value = false;
};

// 27. 提交商品圖片問題。
const reportImageIssue = async (): Promise<void> => {
  if (!productCode.value || reportingImageIssue.value) {
    return;
  }

  reportingImageIssue.value = true;
  actionMessage.value = '';
  try {
    await reportSupermarketProductImage(productCode.value);
    actionMessage.value = t('offers.detail.imageIssueSubmitted');
  } catch {
    actionMessage.value = t('offers.detail.imageIssueSubmitError');
  } finally {
    reportingImageIssue.value = false;
  }
};

// 28. 載入商品詳情
const loadDetail = async (code: string): Promise<void> => {
  if (!code.trim()) {
    detail.value = null;
    detailError.value = t('offers.detail.noProductCode');
    return;
  }

  detailLoading.value = true;
  detailError.value = '';
  actionMessage.value = '';
  isImagePreviewOpen.value = false;
  resetNearbyStores();

  try {
    const { data } = await fetchSupermarketProductDetail(code, 90);
    detail.value = data.data;
    chartMode.value = 'effective';
    trendHoverDate.value = '';
    fillAlertForm(data.data.alertRule ?? null);
    requestBestDealLocation();
  } catch {
    detailError.value = t('offers.detail.loadError');
  } finally {
    detailLoading.value = false;
  }
};

// 29. 返回列表並還原原本搜尋記錄。
const backToList = async (): Promise<void> => {
  if (route.query.return === 'supermarket-offers') {
    await router.replace({ path: '/supermarket-offers', query: returnListQuery.value });
    return;
  }
  await router.push('/supermarket-offers');
};

// 30. 開啟列表篩選結果。
const openListFilter = async (field: 'brand' | 'category', value: string): Promise<void> => {
  await router.push({
    path: '/supermarket-offers',
    query: { [field]: value },
  });
};

// 31. 開啟相關商品
const openRelatedProduct = async (item: SupermarketProduct): Promise<void> => {
  await router.push({
    path: `/supermarket-offers/products/${encodeURIComponent(item.code)}`,
    query: route.query,
  });
};

// 32. 切換同系列商品。
const selectSeriesProduct = async (item: SupermarketProduct): Promise<void> => {
  if (item.code === productCode.value) {
    return;
  }
  await openRelatedProduct(item);
};

// 33. 切換收藏
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
  selectedNearbyStoreId.value = '';
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

// 29.1 選擇其他分店並開啟該店路線。
const selectNearbyStore = (event: Event): void => {
  const target = event.target;
  if (!(target instanceof HTMLSelectElement)) {
    return;
  }

  selectedNearbyStoreId.value = target.value;
  const store = selectedNearbyStore.value;
  if (store) {
    window.open(supermarketDirectionsUrl(store), '_blank', 'noopener,noreferrer');
  }
};

// 29.2 定位並帶到最近門市區。
const focusNearestStores = (): void => {
  requestBestDealLocation();
  document.getElementById('offers-nearest-stores')?.scrollIntoView({ block: 'nearest' });
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
  const colors = ['#FF5A00', '#1A7A3A', '#1A1A1A', '#006973', '#974301', '#6d28d9', '#0f766e', '#a16207'];
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

// 39. 格式化商品名稱
const formatProductName = (item: SupermarketProduct): string =>
  [resolveSupermarketProductName(item), resolveSupermarketProductUnit(item)]
    .filter(Boolean)
    .join(' ');

// 40. 格式化商品品牌
const formatProductBrand = (item: SupermarketProduct): string =>
  resolveSupermarketProductBrand(item);

// 41. 格式化完整商品標題
const formatProductTitle = (item: SupermarketProduct): string =>
  resolveSupermarketProductFullTitle(item);

watch(
  productCode,
  (code) => {
    void loadDetail(code);
  },
  { immediate: true },
);

// 44. 以 Esc 關閉圖片預覽。
const handlePreviewKeydown = (event: KeyboardEvent): void => {
  if (event.key === 'Escape') {
    closeImagePreview();
  }
};

onMounted(() => {
  window.addEventListener('keydown', handlePreviewKeydown);
});

onBeforeUnmount(() => {
  window.removeEventListener('keydown', handlePreviewKeydown);
});
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
            <button
              v-if="productImageUrl && !failedImageCodes.has(product.code)"
              type="button"
              class="gp-image-preview-trigger"
              :aria-label="t('offers.detail.openImagePreview')"
              @click="openImagePreview"
            >
              <img
                :src="productImageUrl"
                :alt="formatProductTitle(product)"
                @error="handleProductImageError(product.code)"
              >
            </button>
            <span
              v-else
              class="gp-product-img-text"
            >
              <span>{{ productInitial }}</span>
            </span>
          </div>
          <div>
            <div class="gp-detail-title-row">
              <button
                v-if="formatProductBrand(product)"
                type="button"
                class="gp-detail-brand-chip"
                @click="openListFilter('brand', formatProductBrand(product))"
              >
                {{ formatProductBrand(product) }}
              </button>
              <div class="gp-detail-title">{{ formatProductName(product) }}</div>
            </div>
            <div class="gp-product-cat">
              <button
                v-for="category in productCategories"
                :key="category.value"
                type="button"
                class="gp-filter-link"
                @click="openListFilter('category', category.value)"
              >
                {{ category.label }}
              </button>
            </div>
            <div
              v-if="hasSeriesOptions"
              class="gp-series-selector"
            >
              <div
                v-if="seriesFormulaOptions.length > 1"
                class="gp-series-row"
                role="group"
                :aria-label="t('offers.detail.formula')"
              >
                <span class="gp-series-label">{{ t('offers.detail.formula') }}</span>
                <div class="gp-series-options">
                  <button
                    v-for="item in seriesFormulaOptions"
                    :key="`formula-${item.code}`"
                    type="button"
                    class="gp-series-option"
                    :class="item.code === product.code ? 'on' : ''"
                    :aria-pressed="item.code === product.code"
                    @click="selectSeriesProduct(item)"
                  >
                    {{ productVariantLabel(item, 'formula') }}
                  </button>
                </div>
              </div>
              <div
                v-if="seriesPackageOptions.length > 1"
                class="gp-series-row"
                role="group"
                :aria-label="t('offers.detail.package')"
              >
                <span class="gp-series-label">{{ t('offers.detail.package') }}</span>
                <div class="gp-series-options">
                  <button
                    v-for="item in seriesPackageOptions"
                    :key="`package-${item.code}`"
                    type="button"
                    class="gp-series-option"
                    :class="item.code === product.code ? 'on' : ''"
                    :aria-pressed="item.code === product.code"
                    @click="selectSeriesProduct(item)"
                  >
                    {{ productVariantLabel(item, 'package') }}
                  </button>
                </div>
              </div>
            </div>
            <div
              v-if="bestDealPrices.length > 0"
              class="gp-best-hero"
            >
              <div>
                <div class="gp-best-hero-stores">
                  <strong
                    v-for="price in bestDealPrices"
                    :key="`best-${price.store}`"
                  >{{ formatStore(price.store) }}</strong>
                </div>
                <p>{{ bestDealOfferText }}</p>
                <span>{{ bestDealBadgeText }}</span>
              </div>
              <div class="gp-best-hero-price">{{ formatOfferPrice(bestDealPrices[0]?.effectiveUnitPrice) }}</div>
            </div>
          </div>
          <div class="gp-detail-actions">
            <div class="gp-detail-primary-actions">
              <button
                type="button"
                class="gp-detail-action gp-detail-action-primary"
                :disabled="nearbyStatus === 'requesting' || bestDealPrices.length === 0"
                @click="focusNearestStores"
              >
                {{ nearbyStatus === 'requesting'
                  ? t('offers.detail.locating')
                  : t('offers.detail.nearestStoreAction') }}
              </button>
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
              <button
                type="button"
                class="gp-detail-action"
                :disabled="reportingImageIssue"
                @click="reportImageIssue"
              >
                {{ reportingImageIssue
                  ? t('offers.detail.imageIssueSubmitting')
                  : t('offers.detail.reportImageIssue') }}
              </button>
            </div>
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

        <!-- 4. 實付價與最低價門市 -->
        <div class="gp-detail-split">
          <section class="gp-detail-card">
            <h3>{{ t('offers.detail.offersAndStores') }}</h3>
            <p class="gp-detail-subtitle">{{ t('offers.detail.latestSnapshot', { date: latestSnapshotText }) }}</p>
          <div class="gp-price-table-wrap">
            <table class="gp-price-table">
              <thead>
                <tr>
                  <th>{{ t('offers.detail.supermarket') }}</th>
                  <th>{{ t('offers.detail.priceNote') }}</th>
                  <th>{{ t('offers.detail.paidUnitPrice') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="price in latestStoreCards"
                  :key="`${price.store}-${price.effectiveUnitPrice}`"
                  :class="isBestDealStore(price) ? 'is-best' : ''"
                >
                  <td>
                    {{ formatStore(price.store) }}
                    <span
                      v-if="isBestDealStore(price)"
                      class="gp-best-pill"
                    >{{ t('offers.detail.bestDealPill') }}</span>
                    <span
                      v-if="secondPriceAdvantageStoreCodes.has(normalizeSupermarketStoreCode(price.store))"
                      class="gp-saving-pill"
                    >{{ t('offers.detail.secondPriceAdvantage', { rate: secondPriceAdvantageRate.toFixed(0) }) }}</span>
                  </td>
                  <td>{{ simpleOfferTrendLabel(price) || priceOfferDisplayText(price) || t('offers.list.currentPrice') }}</td>
                  <td class="gp-price-unit">{{ formatOfferPrice(price.effectiveUnitPrice) }}</td>
                </tr>
              </tbody>
            </table>
          </div>
          </section>
          <section
            id="offers-nearest-stores"
            class="gp-detail-card"
          >
          <h3>{{ t('offers.detail.nearestBestStoreAddress') }}</h3>
          <a
            v-if="selectedNearbyStore"
            class="gp-store-map"
            :href="supermarketDirectionsUrl(selectedNearbyStore)"
            target="_blank"
            rel="noreferrer"
            :aria-label="t('offers.detail.openRoute')"
          >
            <iframe
              class="gp-store-map-frame"
              :src="supermarketMapEmbedUrl(selectedNearbyStore)"
              :title="selectedNearbyStore.name"
              loading="lazy"
              referrerpolicy="no-referrer-when-downgrade"
            />
          </a>
          <button
            v-else
            type="button"
            class="gp-store-map gp-store-map-empty"
            :disabled="nearbyStatus === 'requesting'"
            @click="requestBestDealLocation"
          >
            {{ nearbyStatus === 'requesting'
              ? t('offers.detail.locating')
              : t('offers.detail.findBestStore') }}
          </button>
          <div
            v-if="selectedNearbyStore"
            class="gp-store-map-meta"
          >
            <strong>{{ selectedNearbyStore.name }}</strong>
            <span>{{ selectedNearbyStore.address }}</span>
            <span>{{ formatSupermarketDistanceKm(selectedNearbyStore.distanceKm) }} · {{ selectedNearbyStore.hours || '-' }}</span>
          </div>
          <div
            v-if="selectedNearbyStore"
            class="gp-store-map-actions"
          >
            <a
              class="gp-store-route-button"
              :href="supermarketDirectionsUrl(selectedNearbyStore)"
              target="_blank"
              rel="noreferrer"
            >
              {{ t('offers.detail.openRoute') }}
            </a>
            <select
              class="gp-store-branch-select"
              :value="selectedNearbyStoreId"
              :aria-label="t('offers.detail.otherBranches')"
              :disabled="nearbyStores.length < 2"
              @change="selectNearbyStore"
            >
              <option
                value=""
                disabled
              >
                {{ t('offers.detail.otherBranches') }}
              </option>
              <option
                v-for="store in nearbyStores"
                :key="store.id"
                :value="store.id"
              >
                {{ store.name }} · {{ formatSupermarketDistanceKm(store.distanceKm) }}
              </option>
            </select>
          </div>
          <p
            v-else-if="nearbyMessage"
            class="gp-detail-subtitle"
          >{{ nearbyMessage }}</p>
          </section>
        </div>

        <!-- 5. 價格摘要 + 走勢 -->
        <div class="gp-price-layout">
          <div class="gp-detail-card">
            <h3>{{ t('offers.detail.summary90Days') }}</h3>
            <div class="gp-summary-grid">
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
            <div class="gp-detail-subtitle">{{ t('offers.detail.effectiveEquivalent') }}</div>
            <div class="gp-chart-mode">
              <button
                type="button"
                class="gp-mode-pill"
                :class="chartMode === 'effective' ? 'on' : ''"
                @click="chartMode = 'effective'"
              >
                {{ t('offers.detail.dealPrice') }}
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
                  :alt="formatProductTitle(item)"
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
                <div class="gp-related-title-row">
                  <span
                    v-if="formatProductBrand(item)"
                    class="gp-related-brand-chip"
                  >{{ formatProductBrand(item) }}</span>
                  <div class="gp-related-name">{{ formatProductName(item) }}</div>
                </div>
                <div class="gp-related-offer">{{ relatedProductStoreText(item) }} {{ relatedProductOfferText(item) }}</div>
                <div class="gp-related-price">{{ relatedProductPriceText(item) }}{{ t('offers.detail.perItem') }}</div>
                <div
                  v-if="relatedProductSavingText(item)"
                  class="gp-related-saving"
                >{{ relatedProductSavingText(item) }}</div>
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
                  :alt="formatProductTitle(item)"
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
                <div class="gp-related-title-row">
                  <span
                    v-if="formatProductBrand(item)"
                    class="gp-related-brand-chip"
                  >{{ formatProductBrand(item) }}</span>
                  <div class="gp-related-name">{{ formatProductName(item) }}</div>
                </div>
                <div class="gp-related-offer">{{ relatedProductStoreText(item) }} {{ relatedProductOfferText(item) }}</div>
                <div class="gp-related-price">{{ relatedProductPriceText(item) }}{{ t('offers.detail.perItem') }}</div>
                <div
                  v-if="relatedProductSavingText(item)"
                  class="gp-related-saving"
                >{{ relatedProductSavingText(item) }}</div>
              </div>
            </button>
          </div>
        </div>
      </section>

      <Teleport to="body">
        <div
          v-if="isImagePreviewOpen"
          class="gp-image-preview-overlay"
          @click.self="closeImagePreview"
        >
          <section
            class="gp-image-preview-dialog"
            role="dialog"
            aria-modal="true"
            :aria-label="t('offers.detail.openImagePreview')"
          >
            <button
              type="button"
              class="gp-image-preview-close"
              :aria-label="t('offers.detail.closeImagePreview')"
              @click="closeImagePreview"
            >
              <AppIcon
                name="close"
                :size="20"
              />
            </button>
            <img
              :src="productImageUrl"
              :alt="formatProductTitle(product)"
              @error="handleProductImageError(product.code)"
            >
          </section>
        </div>
      </Teleport>
    </template>
  </main>
</template>

<style scoped>
/* 1. 頁面容器 */
.offers-detail-page {
  width: 100%;
  min-height: calc(100svh - 48px);
  background: var(--sur-2);
  color: var(--ink);
  padding: 18px 16px 44px;
}

/* 2. 詳情面板 */
.gp-detail-panel {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  width: 100%;
  min-width: 0;
  max-width: 1120px;
  margin: 0 auto;
  gap: 12px;
  border: 0;
  border-radius: 0;
  background: transparent;
  padding: 0;
}

/* 3. 麵包屑 */
.gp-breadcrumb {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
  padding: 0 2px;
  margin: 0;
  background: transparent;
  border: 0;
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
  grid-template-columns: 200px minmax(0, 1fr);
  gap: 18px;
  align-items: start;
  margin: 0;
  border: 1px solid var(--bdr);
  border-radius: 16px;
  background: var(--sur);
  padding: 18px;
}

.gp-product-img {
  display: flex;
  grid-row: 1 / span 2;
  width: 100%;
  height: 200px;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--bdr);
  border-radius: 6px;
  background: var(--sur-2);
  overflow: hidden;
}

.gp-image-preview-trigger {
  display: flex;
  width: 100%;
  height: 100%;
  align-items: center;
  justify-content: center;
  border: 0;
  background: transparent;
  cursor: zoom-in;
  padding: 0;
}

.gp-image-preview-trigger img {
  width: auto;
  height: auto;
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
  display: block;
}

.gp-image-preview-overlay {
  position: fixed;
  z-index: 1000;
  inset: 0;
  display: grid;
  place-items: center;
  background: rgb(0 0 0 / 0.68);
  padding: 24px;
}

.gp-image-preview-dialog {
  position: relative;
  display: flex;
  max-width: min(92vw, 960px);
  max-height: calc(100svh - 48px);
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  background: var(--sur);
  padding: 12px;
}

.gp-image-preview-dialog img {
  display: block;
  max-width: min(88vw, 920px);
  max-height: calc(100svh - 96px);
  object-fit: contain;
}

.gp-image-preview-close {
  position: absolute;
  top: 8px;
  right: 8px;
  display: inline-flex;
  width: 36px;
  height: 36px;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--bdr);
  border-radius: 6px;
  background: var(--sur);
  color: var(--ink);
  cursor: pointer;
}

.gp-product-img-text {
  display: inline-flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  color: var(--brand);
  font-family: var(--font-serif);
  font-size: 36px;
  font-weight: 700;
}

.gp-detail-title {
  color: var(--ink);
  font-size: 24px;
  font-weight: 700;
  line-height: 1.25;
}

.gp-detail-title-row {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  gap: 10px;
}

.gp-detail-brand-chip {
  border: 1px solid transparent;
  border-radius: 999px;
  background: var(--brand-light);
  color: var(--brand);
  cursor: pointer;
  font: inherit;
  font-size: 24px;
  font-weight: 700;
  line-height: 1.25;
  padding: 0 8px;
}

.gp-product-cat {
  margin-top: 8px;
  color: var(--ink-3);
  font-size: 13px;
}

.gp-best-hero {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-top: 14px;
  border-radius: 8px;
  background: var(--brand-light);
  padding: 12px;
}

.gp-best-hero-stores {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.gp-best-hero p {
  margin: 4px 0 0;
  color: var(--ink-3);
  font-size: 12px;
}

.gp-best-hero span {
  display: inline-block;
  margin-top: 6px;
  color: var(--brand);
  font-size: 12px;
  font-weight: 700;
}

.gp-best-hero-price {
  color: var(--ink);
  font-size: 28px;
  font-weight: 800;
  line-height: 1;
  white-space: nowrap;
}

.gp-filter-link {
  border: 0;
  background: transparent;
  color: var(--brand);
  cursor: pointer;
  font: inherit;
  padding: 0;
  text-decoration: underline;
  text-underline-offset: 2px;
}

.gp-product-cat .gp-filter-link + .gp-filter-link {
  margin-left: 6px;
}

/* 5. 同系列選擇 */
.gp-series-selector {
  display: grid;
  gap: 8px;
  margin-top: 14px;
}

.gp-series-row {
  display: flex;
  align-items: flex-start;
  gap: 10px;
}

.gp-series-label {
  flex: 0 0 70px;
  color: var(--ink-3);
  font-size: 12px;
  line-height: 32px;
}

.gp-series-options {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.gp-series-option {
  min-height: 32px;
  border: 1px solid var(--bdr);
  border-radius: 6px;
  background: var(--sur-2);
  color: var(--ink);
  cursor: pointer;
  font: inherit;
  font-size: 12px;
  line-height: 1.3;
  padding: 7px 10px;
}

.gp-series-option:hover {
  border-color: var(--brand);
  color: var(--brand);
}

.gp-series-option.on {
  border-color: var(--brand);
  background: var(--brand-light);
  color: var(--brand);
  font-weight: 700;
}

.gp-detail-actions {
  display: block;
  grid-column: 2;
  min-width: 0;
  margin-top: 12px;
}

.gp-detail-primary-actions {
  display: flex;
  width: 100%;
  min-width: 0;
  flex-wrap: nowrap;
  gap: 8px;
  justify-content: flex-start;
  overflow-x: auto;
  scrollbar-width: none;
}

.gp-detail-primary-actions::-webkit-scrollbar {
  display: none;
}

.gp-detail-action {
  flex: 0 0 auto;
  border: 1px solid var(--brand);
  border-radius: 6px;
  background: var(--brand-light);
  color: var(--brand);
  cursor: pointer;
  font-family: inherit;
  font-size: 12px;
  font-weight: 700;
  padding: 9px 12px;
  white-space: nowrap;
}

.gp-detail-action.on {
  background: var(--brand);
  color: #fff;
}

.gp-detail-action-primary {
  background: var(--brand);
  color: #fff;
}

.gp-detail-action:disabled {
  cursor: not-allowed;
  opacity: 0.52;
}

/* 5. 操作訊息 */
.gp-action-message {
  margin: 0;
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
  border-radius: 16px;
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
  margin: 0;
}

/* 7. 到價提醒表單 */
.gp-alert-card {
  margin: 0;
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
.gp-price-table-wrap {
  overflow: auto;
  border: 0;
  border-radius: 0;
}

.gp-detail-split {
  display: grid;
  grid-template-columns: minmax(0, 1.15fr) minmax(260px, 0.85fr);
  gap: 12px;
  align-items: stretch;
}

.gp-price-table {
  width: 100%;
  min-width: 0;
  border-collapse: collapse;
  font-size: 13px;
}

.gp-price-table th {
  color: var(--ink-3);
  font-size: 12px;
  font-weight: 700;
  padding: 10px 8px;
  text-align: left;
}

.gp-price-table td {
  border-top: 1px solid var(--bdr);
  padding: 10px 8px;
  vertical-align: middle;
}

.gp-price-table tr.is-best td {
  background: var(--brand-light);
}

.gp-price-unit {
  font-weight: 800;
  text-align: right;
  white-space: nowrap;
}

.gp-best-pill,
.gp-saving-pill {
  display: inline-block;
  margin-left: 6px;
  border-radius: 999px;
  font-size: 11px;
  font-weight: 700;
  line-height: 1.2;
  padding: 2px 6px;
}

.gp-best-pill {
  background: var(--brand);
  color: #fff;
}

.gp-saving-pill {
  border: 1px solid var(--bdr);
  background: var(--sur);
  color: var(--brand);
}

.gp-store-map {
  position: relative;
  display: block;
  height: 180px;
  margin-top: 12px;
  overflow: hidden;
  border: 1px solid var(--bdr);
  border-radius: 12px;
  background: var(--sur-2);
}

.gp-store-map-frame {
  display: block;
  width: 100%;
  height: 100%;
  border: 0;
  pointer-events: none;
}

.gp-store-map-empty {
  width: 100%;
  color: var(--brand);
  cursor: pointer;
  font: inherit;
  font-size: 13px;
  font-weight: 700;
}

.gp-store-map-empty:disabled {
  cursor: not-allowed;
  opacity: 0.6;
}

.gp-store-map-meta {
  display: grid;
  gap: 4px;
  margin-top: 10px;
}

.gp-store-map-meta strong {
  color: var(--ink);
  font-size: 14px;
}

.gp-store-map-meta span {
  color: var(--ink-3);
  font-size: 12px;
  line-height: 1.5;
}

.gp-store-map-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
  margin-top: 12px;
}

.gp-store-route-button,
.gp-store-branch-select {
  min-height: 40px;
  border: 1px solid var(--bdr);
  border-radius: 8px;
  font: inherit;
  font-size: 13px;
  font-weight: 700;
}

.gp-store-route-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: var(--brand);
  color: #fff;
  padding: 0 14px;
  text-decoration: none;
}

.gp-store-branch-select {
  min-width: 150px;
  max-width: 100%;
  background: var(--sur);
  color: var(--ink);
  cursor: pointer;
  padding: 0 34px 0 12px;
}

.gp-store-branch-select:disabled {
  color: var(--ink-3);
  cursor: not-allowed;
  opacity: 0.72;
}

.gp-store-route-button:focus-visible,
.gp-store-branch-select:focus-visible {
  outline: 2px solid var(--brand);
  outline-offset: 2px;
}

.gp-nearest-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-top: 12px;
  border: 1px solid var(--bdr);
  border-radius: 8px;
  background: var(--sur);
  padding: 12px;
}

.gp-nearest-card strong {
  display: block;
  color: var(--ink);
  font-size: 15px;
}

.gp-nearest-card p {
  margin: 4px 0 0;
  color: var(--ink-3);
  font-size: 12px;
}

.gp-nearest-card a {
  border-radius: 8px;
  background: var(--brand);
  color: #fff;
  font-size: 13px;
  font-weight: 700;
  padding: 8px 12px;
  text-decoration: none;
  white-space: nowrap;
}

.gp-compare-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
  padding: 12px;
  border: 1px solid var(--bdr);
  border-radius: 6px;
  background: var(--sur-2);
}

.gp-compare-card {
  position: relative;
  border: 1px solid var(--bdr);
  border-radius: 6px;
  background: var(--sur-2);
  padding: 12px;
}

.gp-compare-store {
  color: var(--ink);
  font-size: 13px;
  font-weight: 800;
  padding-right: 92px;
}

.gp-compare-price {
  margin-top: 8px;
  color: var(--brand);
  font-size: 22px;
  font-weight: 800;
  line-height: 1;
}

.gp-compare-offer {
  display: inline-block;
  margin-top: 8px;
  background: var(--brand-light);
  color: var(--brand);
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
  background: var(--brand-light);
  color: var(--brand);
  font-weight: 800;
  padding: 4px 7px;
}

/* 11. 價格摘要 + 走勢佈局 */
.gp-price-layout {
  display: grid;
  grid-template-columns: 280px minmax(0, 1fr);
  gap: 12px;
  margin: 0;
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
  background: var(--sur-2);
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
  width: 100%;
  min-width: 0;
  margin: 0;
}

.gp-related-scroll {
  display: flex;
  width: 100%;
  min-width: 0;
  max-width: 100%;
  gap: 10px;
  overflow-x: auto;
  overflow-y: hidden;
  padding-bottom: 4px;
  scroll-snap-type: x proximity;
  scrollbar-width: thin;
}

.gp-related-card {
  flex: 0 0 220px;
  min-width: 0;
  display: flex;
  flex-direction: column;
  align-items: stretch;
  justify-content: flex-start;
  border: 1px solid var(--bdr);
  border-radius: 8px;
  background: var(--sur);
  overflow: hidden;
  cursor: pointer;
  font: inherit;
  padding: 0;
  scroll-snap-align: start;
  text-align: left;
  color: var(--ink);
}

.gp-related-img {
  display: flex;
  flex: 0 0 120px;
  height: 120px;
  align-items: flex-start;
  justify-content: center;
  border-bottom: 1px solid var(--bdr);
  background: var(--sur-2);
  overflow: hidden;
  padding: 8px 8px 0;
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
  display: flex;
  flex: 1 1 auto;
  flex-direction: column;
  justify-content: flex-start;
  padding: 11px;
}

.gp-related-name {
  color: var(--ink);
  font-size: 13px;
  font-weight: 800;
  line-height: 1.35;
}

.gp-related-title-row {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  gap: 8px;
}

.gp-related-brand-chip {
  display: inline-flex;
  flex: 0 0 auto;
  border: 1px solid transparent;
  border-radius: 999px;
  background: var(--brand-light);
  color: var(--brand);
  font-size: 13px;
  font-weight: 700;
  line-height: 1.35;
  padding: 0 8px;
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

.gp-related-saving {
  margin-top: 3px;
  color: var(--brand);
  font-size: 11px;
  font-weight: 700;
  line-height: 1.2;
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
    padding: 12px;
  }

  .gp-product-img {
    grid-row: auto;
    width: 100%;
    height: 160px;
  }

  .gp-detail-actions {
    grid-column: auto;
    width: 100%;
  }

  .gp-detail-primary-actions {
    justify-content: flex-start;
  }

  .gp-series-label {
    flex-basis: 68px;
  }

  .gp-compare-grid,
  .gp-summary-grid {
    grid-template-columns: 1fr;
  }

  .gp-best-shop {
    grid-template-columns: 1fr;
  }

  .gp-detail-split {
    grid-template-columns: 1fr;
  }

  .gp-best-hero,
  .gp-nearest-card {
    align-items: flex-start;
    flex-direction: column;
  }

  .gp-price-layout {
    grid-template-columns: 1fr;
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

/* 16. 夜間護眼色 */
:global(html[data-theme="night"]) .gp-product-img,
:global(html[data-theme="night"]) .gp-related-img,
:global(html[data-theme="night"]) .gp-chart,
:global(html[data-theme="night"]) .gp-chart rect {
  background: rgb(42 48 56);
}

:global(html[data-theme="night"]) .gp-product-img img,
:global(html[data-theme="night"]) .gp-related-img img,
:global(html[data-theme="night"]) .gp-image-preview-dialog img {
  filter: brightness(0.88);
}

:global(html[data-theme="night"]) .gp-chart line {
  stroke: rgb(62 70 80);
}
</style>

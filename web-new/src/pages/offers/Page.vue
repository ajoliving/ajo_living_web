<!--
 * 綜合優惠頻道頁。
 * 1. 使用 AJO 後端超市優惠摘要與搜尋接口。
 * 2. 提供商品搜尋、分類、品牌、商店、排序、收藏與分頁。
 * 3. 桌面保留摘要 Hero；手機以搜尋、快捷篩選與底部篩選彈窗呈現。
-->
<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import {
  addSupermarketFavorite,
  fetchSupermarketFavorites,
  fetchSupermarketSummary,
  removeSupermarketFavorite,
  searchSupermarketProducts,
} from '@/httpapis/supermarket-offers';
import { readStoredAccessToken } from '@/httpapis/auth-session';
import AppIcon from '@/shared/components/base/AppIcon.vue';
import { usePreferenceStore } from '@/stores/preferences';
import type {
  SupermarketProduct,
  SupermarketSearchParams,
  SupermarketSearchResult,
  SupermarketStorePrice,
  SupermarketSummary,
  SupermarketValueCount,
} from '@/model/supermarket-offers';
import {
  displaySupermarketCategory,
  displaySupermarketStore,
  formatSupermarketDate,
  formatSupermarketHKPrice,
  supermarketPriceDiscountRate,
  supermarketOfferTexts,
  supermarketOfferDisplayText,
  supermarketIsSimpleOffer,
  supermarketPrimaryPrice,
  resolveSupermarketProductBrand,
  resolveSupermarketProductCategory,
  resolveSupermarketProductFullTitle,
  resolveSupermarketProductName,
  resolveSupermarketProductUnit,
  supermarketStorePrices,
} from '@/utils/supermarket-offers';

type ViewMode = 'grid' | 'table';

interface FilterPill {
  label: string;
  value: string;
}

interface SortOption {
  label: string;
  value: string;
}

interface PageButton {
  label: string | number;
  key: string;
  page: number;
  active: boolean;
  disabled: boolean;
}

const router = useRouter();
const route = useRoute();
const { t } = useI18n();
const preferenceStore = usePreferenceStore();
const pageSize = 20;
const summary = ref<SupermarketSummary | null>(null);
const searchResult = ref<SupermarketSearchResult | null>(null);
const favorites = ref<SupermarketProduct[]>([]);
const summaryLoading = ref(false);
const searchLoading = ref(false);
const favoritesLoading = ref(false);
const summaryError = ref('');
const searchError = ref('');
const actionMessage = ref('');
const searchQuery = ref('');
const searchSuggestions = ref<SupermarketProduct[]>([]);
const suggestionsVisible = ref(false);
const suggestionsLoading = ref(false);
const suggestionsExpanded = ref(false);
const failedImageCodes = ref(new Set<string>());
const activeCategory = ref('');
const activeStore = ref('');
const activeBrand = ref('');
const activeSort = ref('discount');
const showFavoritesOnly = ref(false);
const viewMode = ref<ViewMode>('grid');
const currentPage = ref(1);
const isMobileFilterOpen = ref(false);
let suggestionTimer: ReturnType<typeof setTimeout> | undefined;
let suggestionRequestVersion = 0;

const sortOptions = computed<SortOption[]>(() => [
  { label: t('offers.list.sortDiscount'), value: 'discount' },
  { label: t('offers.list.sortEffective'), value: 'effective' },
  { label: t('offers.list.sortDifference'), value: 'diff' },
  { label: t('offers.list.sortName'), value: 'name' },
  { label: t('offers.list.sortBrand'), value: 'brand' },
]);

// 1. 控制搜尋建議的初始與展開顯示數量。
const visibleSearchSuggestions = computed<SupermarketProduct[]>(() =>
  searchSuggestions.value.slice(0, suggestionsExpanded.value ? 12 : 6),
);
const canExpandSearchSuggestions = computed(() =>
  !suggestionsExpanded.value && searchSuggestions.value.length > 6,
);

// 2. 建立分類篩選項
const categoryPills = computed<FilterPill[]>(() => [
  { label: t('offers.list.all'), value: '' },
  ...valueCountPills(
    summary.value?.categories ?? [],
    (value) => displaySupermarketCategory(value, preferenceStore.locale),
    9,
  ),
]);

// 4. 建立商店篩選項
const storePills = computed<FilterPill[]>(() => [
  { label: t('offers.list.allStores'), value: '' },
  ...valueCountPills(
    summary.value?.stores ?? [],
    (value) => displaySupermarketStore(value, preferenceStore.locale),
    9,
  ),
]);

// 5. 建立品牌篩選項
const brandOptions = computed<string[]>(() => searchResult.value?.brands ?? []);

// 6. 計算目前已套用的手機端篩選數量
const activeFilterCount = computed(() =>
  Number(Boolean(activeCategory.value))
  + Number(Boolean(activeStore.value))
  + Number(Boolean(activeBrand.value))
  + Number(showFavoritesOnly.value),
);

// 7. 取得目前列表商品
const visibleProducts = computed<SupermarketProduct[]>(() => {
  if (showFavoritesOnly.value) {
    const start = (currentPage.value - 1) * pageSize;
    return favorites.value.slice(start, start + pageSize);
  }
  return searchResult.value?.items ?? [];
});

// 8. 取得目前總數
const totalCount = computed(() => (showFavoritesOnly.value ? favorites.value.length : searchResult.value?.total ?? 0));

// 9. 取得目前總頁數
const totalPages = computed(() => Math.max(1, Math.ceil(totalCount.value / pageSize)));

// 10. 建立分頁按鈕
const pageButtons = computed<PageButton[]>(() => {
  const pages = new Set<number>([1, totalPages.value, currentPage.value]);
  if (currentPage.value > 1) pages.add(currentPage.value - 1);
  if (currentPage.value < totalPages.value) pages.add(currentPage.value + 1);
  const numericPages = [...pages].filter((page) => page >= 1 && page <= totalPages.value).sort((a, b) => a - b);

  return [
    {
      label: t('offers.list.previous'),
      key: 'prev',
      page: Math.max(1, currentPage.value - 1),
      active: false,
      disabled: currentPage.value <= 1,
    },
    ...numericPages.map((page) => ({
      label: page,
      key: String(page),
      page,
      active: page === currentPage.value,
      disabled: false,
    })),
    {
      label: t('offers.list.next'),
      key: 'next',
      page: Math.min(totalPages.value, currentPage.value + 1),
      active: false,
      disabled: currentPage.value >= totalPages.value,
    },
  ];
});

// 11. 建立更新資訊文字
const updatedBarText = computed(() => {
  const updatedDate = formatSupermarketDate(
    summary.value?.metadata?.latestSnapshotDate,
    preferenceStore.locale,
  );
  const productCount = formatInteger(summary.value?.stats.products);
  const prefix = updatedDate
    ? t('offers.list.updatedDate', { date: updatedDate })
    : t('offers.list.updatedPending');
  return t('offers.list.updatedSummary', { prefix, count: productCount });
});

// 12. 載入摘要
const loadSummary = async (): Promise<void> => {
  summaryLoading.value = true;
  summaryError.value = '';
  try {
    const { data } = await fetchSupermarketSummary();
    summary.value = data.data;
  } catch {
    summaryError.value = t('offers.list.summaryError');
  } finally {
    summaryLoading.value = false;
  }
};

// 13. 載入搜尋結果
const loadSearch = async (): Promise<void> => {
  searchLoading.value = true;
  searchError.value = '';
  actionMessage.value = '';
  try {
    const params: SupermarketSearchParams = {
      q: searchQuery.value.trim(),
      category: activeCategory.value,
      brand: activeBrand.value,
      store: activeStore.value,
      offerOnly: false,
      sort: activeSort.value,
      page: currentPage.value,
      pageSize,
    };
    const { data } = await searchSupermarketProducts(params);
    searchResult.value = data.data;
  } catch {
    searchError.value = t('offers.list.searchError');
  } finally {
    searchLoading.value = false;
  }
};

// 14. 載入搜尋建議，保留最多 12 個名稱、品牌或分類匹配的商品。
const loadSearchSuggestions = async (): Promise<void> => {
  const query = searchQuery.value.trim();
  const requestVersion = ++suggestionRequestVersion;
  suggestionsExpanded.value = false;
  if (!query) {
    searchSuggestions.value = [];
    suggestionsVisible.value = false;
    return;
  }

  suggestionsLoading.value = true;
  try {
    const { data } = await searchSupermarketProducts({
      q: query,
      offerOnly: false,
      sort: 'name',
      page: 1,
      pageSize: 12,
    });
    if (requestVersion === suggestionRequestVersion) {
      searchSuggestions.value = data.data.items;
      suggestionsVisible.value = true;
    }
  } catch {
    if (requestVersion === suggestionRequestVersion) {
      searchSuggestions.value = [];
      suggestionsVisible.value = false;
    }
  } finally {
    if (requestVersion === suggestionRequestVersion) {
      suggestionsLoading.value = false;
    }
  }
};

// 15. 延遲查詢輸入建議，避免每個字元都立即發送請求。
const scheduleSearchSuggestions = (): void => {
  if (suggestionTimer) {
    clearTimeout(suggestionTimer);
  }
  suggestionTimer = setTimeout(() => {
    void loadSearchSuggestions();
  }, 180);
};

// 16. 顯示目前搜尋字詞的建議。
const showSearchSuggestions = (): void => {
  if (searchQuery.value.trim()) {
    scheduleSearchSuggestions();
  }
};

// 17. 延後收起建議，讓點選結果可以完成。
const hideSearchSuggestions = (): void => {
  window.setTimeout(() => {
    suggestionsVisible.value = false;
  }, 120);
};

// 18. 展開更多搜尋建議。
const expandSearchSuggestions = (): void => {
  suggestionsExpanded.value = true;
};

// 19. 套用一個建議商品並提交搜尋。
const selectSearchSuggestion = (product: SupermarketProduct): void => {
  searchQuery.value = product.name;
  suggestionsVisible.value = false;
  submitSearch();
};

// 20. 記錄無法載入的商品圖片，改用既有後備顯示。
const handleProductImageError = (code: string): void => {
  failedImageCodes.value = new Set([...failedImageCodes.value, code]);
};

// 21. 載入收藏
const loadFavorites = async (): Promise<void> => {
  if (!readStoredAccessToken()) {
    favorites.value = [];
    return;
  }

  favoritesLoading.value = true;
  searchError.value = '';
  try {
    const { data } = await fetchSupermarketFavorites({ page: 1, pageSize: 80 });
    favorites.value = data.data.items;
  } catch {
    searchError.value = t('offers.list.favoritesError');
  } finally {
    favoritesLoading.value = false;
  }
};

// 22. 提交搜尋
const submitSearch = (): void => {
  showFavoritesOnly.value = false;
  currentPage.value = 1;
  suggestionsVisible.value = false;
  void loadSearch();
};

// 23. 切換分類
const selectCategory = (category: string): void => {
  activeCategory.value = category;
  showFavoritesOnly.value = false;
  currentPage.value = 1;
  void loadSearch();
};

// 24. 切換商店
const selectStore = (store: string): void => {
  activeStore.value = store;
  showFavoritesOnly.value = false;
  currentPage.value = 1;
  void loadSearch();
};

// 25. 套用手機端快捷篩選
const handleMobileQuickFilter = (filter: 'category' | 'store', event: Event): void => {
  const value = (event.target as HTMLSelectElement).value;
  if (filter === 'category') {
    selectCategory(value);
    return;
  }
  selectStore(value);
};

// 26. 開關手機端篩選彈窗
const openMobileFilter = (): void => {
  isMobileFilterOpen.value = true;
};

const closeMobileFilter = (): void => {
  isMobileFilterOpen.value = false;
};

// 27. 重設篩選條件並保留搜尋文字
const resetFilters = (): void => {
  activeCategory.value = '';
  activeStore.value = '';
  activeBrand.value = '';
  activeSort.value = 'discount';
  showFavoritesOnly.value = false;
  currentPage.value = 1;
  void loadSearch();
};

// 28. 切換品牌
const selectBrand = (): void => {
  showFavoritesOnly.value = false;
  currentPage.value = 1;
  void loadSearch();
};

// 29. 切換排序
const selectSort = (): void => {
  currentPage.value = 1;
  if (!showFavoritesOnly.value) {
    void loadSearch();
  }
};

// 30. 切換收藏列表
const toggleFavoritesOnly = async (): Promise<void> => {
  if (!showFavoritesOnly.value && !readStoredAccessToken()) {
    await openLogin('/supermarket-offers');
    return;
  }

  showFavoritesOnly.value = !showFavoritesOnly.value;
  currentPage.value = 1;
  if (showFavoritesOnly.value) {
    await loadFavorites();
  } else {
    await loadSearch();
  }
};

// 31. 切換商品收藏
const toggleFavorite = async (product: SupermarketProduct): Promise<void> => {
  if (!readStoredAccessToken()) {
    await openLogin(`/supermarket-offers/products/${encodeURIComponent(product.code)}`);
    return;
  }

  actionMessage.value = '';
  try {
    if (product.isFavorite) {
      await removeSupermarketFavorite(product.code);
      patchFavoriteState(product.code, false);
      favorites.value = favorites.value.filter((item) => item.code !== product.code);
      actionMessage.value = t('offers.list.favoriteRemoved');
      return;
    }

    const { data } = await addSupermarketFavorite(product.code);
    patchFavoriteState(product.code, true);
    favorites.value = [data.data, ...favorites.value.filter((item) => item.code !== product.code)];
    actionMessage.value = t('offers.list.favoriteAdded');
  } catch {
    actionMessage.value = t('offers.list.favoriteError');
  }
};

// 32. 切換視圖模式
const setView = (mode: ViewMode): void => {
  viewMode.value = mode;
};

// 33. 切換分頁
const selectPage = (page: PageButton): void => {
  if (page.disabled || currentPage.value === page.page) {
    return;
  }
  currentPage.value = page.page;
  if (!showFavoritesOnly.value) {
    void loadSearch();
  }
};

// 34. 開啟商品詳情。
const openDetail = (product: SupermarketProduct): void => {
  void router.push({
    path: `/supermarket-offers/products/${encodeURIComponent(product.code)}`,
    query: {
      return: 'supermarket-offers',
      q: searchQuery.value || undefined,
      category: activeCategory.value || undefined,
      store: activeStore.value || undefined,
      brand: activeBrand.value || undefined,
      sort: activeSort.value || undefined,
      page: currentPage.value > 1 ? String(currentPage.value) : undefined,
      view: viewMode.value !== 'grid' ? viewMode.value : undefined,
    },
  });
};

// 35. 前往登入
const openLogin = async (redirect: string): Promise<void> => {
  await router.push({ path: '/login', query: { redirect } });
};

// 36. 更新商品收藏狀態
const patchFavoriteState = (productCode: string, isFavorite: boolean): void => {
  searchResult.value?.items.forEach((item) => {
    if (item.code === productCode) {
      item.isFavorite = isFavorite;
    }
  });
  favorites.value.forEach((item) => {
    if (item.code === productCode) {
      item.isFavorite = isFavorite;
    }
  });
};

// 37. 建立篩選按鈕
const valueCountPills = (
  values: SupermarketValueCount[],
  formatter: (value: string) => string,
  limit: number,
): FilterPill[] =>
  values
    .slice(0, limit)
    .map((item) => ({ label: formatter(item.value), value: item.value }))
    .filter((item) => item.value);

// 38. 格式化整數
const formatInteger = (value: number | undefined): string =>
  (value ?? 0).toLocaleString(preferenceStore.locale);

// 39. 格式化商店名稱
const formatStore = (value: string): string =>
  displaySupermarketStore(value, preferenceStore.locale);

// 40. 格式化商品分類
const formatCategory = (value: string): string =>
  displaySupermarketCategory(value, preferenceStore.locale);

// 41. 格式化商品價格
const formatOfferPrice = (value: number | null | undefined): string =>
  formatSupermarketHKPrice(value, preferenceStore.locale);

// 42. 取得商品商店價格
const productStorePrices = (product: SupermarketProduct) =>
  supermarketStorePrices(product, preferenceStore.locale);

// 43. 取得商品主要價格
const productPrimaryPrice = (product: SupermarketProduct) =>
  supermarketPrimaryPrice(product, preferenceStore.locale);

// 44. 取得商品優惠文字
const productOfferTexts = (product: SupermarketProduct): string[] =>
  supermarketOfferTexts(product, preferenceStore.locale)
    .map((value) => supermarketOfferDisplayText(value))
    .filter(Boolean);

// 45. 顯示商品主標題名稱
const productDisplayName = (product: SupermarketProduct): string =>
  [resolveSupermarketProductName(product), resolveSupermarketProductUnit(product)]
    .filter(Boolean)
    .join(' ');

// 46. 顯示不含單位的商品名稱
const productTitleName = (product: SupermarketProduct): string =>
  resolveSupermarketProductName(product);

// 47. 顯示商品單位
const productDisplayUnit = (product: SupermarketProduct): string =>
  resolveSupermarketProductUnit(product);

// 48. 顯示商品品牌
const productDisplayBrand = (product: SupermarketProduct): string =>
  resolveSupermarketProductBrand(product);

// 49. 顯示商品完整標題
const productDisplayTitle = (product: SupermarketProduct): string =>
  resolveSupermarketProductFullTitle(product);

// 50. 顯示商品完整分類。
const productCategoryText = (product: SupermarketProduct): string =>
  resolveSupermarketProductCategory(product);

// 51. 顯示單位價格
const formatOfferUnitPrice = (value: number | null | undefined): string =>
  `${formatOfferPrice(value)}${t('offers.list.perItem')}`;

// 53. 顯示不含港幣前綴的金額
const formatPlainDollar = (value: number): string =>
  `$${value.toLocaleString(preferenceStore.locale, {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  })}`;

// 54. 取得最低價門店
const productFeaturedPrice = (product: SupermarketProduct): SupermarketStorePrice =>
  productStorePrices(product)[0] ?? productPrimaryPrice(product);

// 55. 取得其餘可對比門店
const productOtherPrices = (product: SupermarketProduct): SupermarketStorePrice[] =>
  productStorePrices(product).slice(1, 3);

// 56. 判斷門店是否已計算優惠
const productHasCalculatedOffer = (price: SupermarketStorePrice): boolean =>
  supermarketPriceDiscountRate(price) > 0 || Boolean(supermarketOfferDisplayText(price.offer));

// 57. 顯示最低價主卡標籤
const productBestBadgeText = (price: SupermarketStorePrice): string => {
  const lowestLabel = t('offers.list.lowestAcrossStores');
  return productHasCalculatedOffer(price)
    ? `${lowestLabel} · ${t('offers.list.offerApplied')}`
    : lowestLabel;
};

// 58. 顯示最低價主卡優惠說明
const productFeaturedOfferText = (product: SupermarketProduct, price: SupermarketStorePrice): string => {
  const offer = supermarketOfferDisplayText(price.offer);
  if (offer) {
    return `${offer} · ${t('offers.list.paidUnitPrice')}`;
  }
  const othersHaveExtra = productStorePrices(product)
    .slice(1)
    .some((item) => Boolean(supermarketOfferDisplayText(item.offer)) && !supermarketIsSimpleOffer(item.offer));
  return othersHaveExtra ? t('offers.list.currentPriceNoExtra') : t('offers.list.currentPrice');
};

// 59. 顯示最低價主卡對照價
const productFeaturedCompareText = (product: SupermarketProduct, price: SupermarketStorePrice): string => {
  if (price.listPrice > price.effectiveUnitPrice) {
    return t('offers.list.originalPrice', { price: formatPlainDollar(price.listPrice) });
  }
  const otherPrices = productStorePrices(product)
    .slice(1)
    .map((item) => item.effectiveUnitPrice)
    .filter((value) => Number.isFinite(value) && value > price.effectiveUnitPrice);
  if (otherPrices.length === 0) {
    return '';
  }
  const maxOther = Math.max(...otherPrices);
  const minOther = Math.min(...otherPrices);
  return minOther === maxOther
    ? t('offers.list.otherStorePrice', { price: formatPlainDollar(maxOther) })
    : t('offers.list.otherStoreUpTo', { price: formatPlainDollar(maxOther) });
};

// 60. 顯示其他門店優惠說明
const productStoreOfferLabel = (price: SupermarketStorePrice): string =>
  supermarketOfferDisplayText(price.offer) || t('offers.list.currentPrice');

// 35. 還原由商品詳情頁帶回的搜尋條件。
const restoreSearchState = (): void => {
  const query = route.query;
  searchQuery.value = typeof query.q === 'string' ? query.q : '';
  activeCategory.value = typeof query.category === 'string' ? query.category : '';
  activeStore.value = typeof query.store === 'string' ? query.store : '';
  activeBrand.value = typeof query.brand === 'string' ? query.brand : '';
  activeSort.value = typeof query.sort === 'string' && sortOptions.value.some((item) => item.value === query.sort)
    ? query.sort
    : 'discount';
  viewMode.value = query.view === 'table' ? 'table' : 'grid';
  const page = typeof query.page === 'string' ? Number(query.page) : 1;
  currentPage.value = Number.isInteger(page) && page > 0 ? page : 1;
};

onMounted(() => {
  restoreSearchState();
  void loadSummary();
  void loadSearch();
});

onBeforeUnmount(() => {
  if (suggestionTimer) {
    clearTimeout(suggestionTimer);
  }
});
</script>

<template>
  <main
    id="page-offers"
    class="page"
  >
    <button
      v-if="isMobileFilterOpen"
      type="button"
      class="gp-filter-backdrop"
      :aria-label="t('offers.list.closeFilters')"
      @click="closeMobileFilter"
    ></button>

    <aside
      id="offers-mobile-filter-sheet"
      class="gp-filter-sheet"
      :class="{ 'is-filter-open': isMobileFilterOpen }"
      :role="isMobileFilterOpen ? 'dialog' : undefined"
      :aria-modal="isMobileFilterOpen ? 'true' : undefined"
      aria-labelledby="offers-mobile-filter-title"
    >
      <header class="gp-filter-sheet-header">
        <h2 id="offers-mobile-filter-title">{{ t('offers.list.filters') }}</h2>
        <button
          type="button"
          class="gp-filter-sheet-close filter-sheet-close"
          :aria-label="t('offers.list.closeFilters')"
          @click="closeMobileFilter"
        >
          <AppIcon
            name="close"
            :size="20"
          />
        </button>
      </header>

      <div class="gp-filter-sheet-body">
        <section class="gp-filter-section">
          <h3>{{ t('offers.list.category') }}</h3>
          <div class="gp-filter-tags">
            <button
              v-for="category in categoryPills"
              :key="`sheet-category-${category.value || 'all'}`"
              type="button"
              class="gp-filter-tag"
              :class="activeCategory === category.value ? 'on' : ''"
              @click="selectCategory(category.value)"
            >
              {{ category.label }}
            </button>
          </div>
        </section>

        <section class="gp-filter-section">
          <h3>{{ t('offers.list.store') }}</h3>
          <div class="gp-filter-tags">
            <button
              v-for="store in storePills"
              :key="`sheet-store-${store.value || 'all'}`"
              type="button"
              class="gp-filter-tag"
              :class="activeStore === store.value ? 'on' : ''"
              @click="selectStore(store.value)"
            >
              {{ store.label }}
            </button>
          </div>
        </section>

        <label class="gp-filter-select-field">
          <span>{{ t('offers.list.brand') }}</span>
          <select
            v-model="activeBrand"
            @change="selectBrand"
          >
            <option value="">{{ t('offers.list.allBrands') }}</option>
            <option
              v-for="brand in brandOptions"
              :key="`sheet-brand-${brand}`"
              :value="brand"
            >
              {{ brand }}
            </option>
          </select>
        </label>

        <label class="gp-filter-select-field">
          <span>{{ t('offers.list.sort') }}</span>
          <select
            v-model="activeSort"
            @change="selectSort"
          >
            <option
              v-for="option in sortOptions"
              :key="`sheet-sort-${option.value}`"
              :value="option.value"
            >
              {{ option.label }}
            </option>
          </select>
        </label>

        <section class="gp-filter-section gp-view-filter-section">
          <h3>{{ t('offers.list.viewMode') }}</h3>
          <div class="gp-view-filter-actions">
            <button
              type="button"
              :class="viewMode === 'grid' ? 'on' : ''"
              @click="setView('grid')"
            >{{ t('offers.list.grid') }}</button>
            <button
              type="button"
              :class="viewMode === 'table' ? 'on' : ''"
              @click="setView('table')"
            >{{ t('offers.list.table') }}</button>
          </div>
        </section>
      </div>

      <footer class="gp-filter-sheet-actions filter-sheet-actions">
        <button
          type="button"
          class="gp-filter-sheet-reset"
          @click="resetFilters"
        >{{ t('offers.list.reset') }}</button>
        <button
          type="button"
          class="gp-filter-sheet-apply"
          @click="closeMobileFilter"
        >{{ t('offers.list.viewOffers', { count: formatInteger(totalCount) }) }}</button>
      </footer>
    </aside>

    <!-- 1. 超市情報標題 -->
    <section class="gp-hero">
      <div class="gp-hero-left">
        <div class="gp-hero-title">{{ t('offers.list.heroTitle') }}</div>
        <div class="gp-hero-sub">{{ t('offers.list.heroSubtitle') }}</div>
      </div>
    </section>

    <!-- 2. 控制欄 -->
    <section class="gp-controls">
      <form
        class="gp-search-row"
        @submit.prevent="submitSearch"
      >
        <div class="gp-search-box">
          <span class="gp-search-ico">⌕</span>
          <input
            v-model="searchQuery"
            class="gp-sinput"
            :placeholder="t('offers.list.searchPlaceholder')"
            autocomplete="off"
            @input="scheduleSearchSuggestions"
            @focus="showSearchSuggestions"
            @blur="hideSearchSuggestions"
          >
          <div
            v-if="suggestionsVisible"
            class="gp-search-suggestions"
          >
            <p
              v-if="suggestionsLoading"
              class="gp-search-suggestion-state"
            >{{ t('offers.list.loading') }}</p>
            <button
              v-for="product in visibleSearchSuggestions"
              :key="product.code"
              type="button"
              class="gp-search-suggestion"
              @mousedown.prevent="selectSearchSuggestion(product)"
            >
              <span class="gp-search-suggestion-head">
                <span
                  v-if="productDisplayBrand(product)"
                  class="gp-search-suggestion-chip"
                >{{ productDisplayBrand(product) }}</span>
                <span class="gp-search-suggestion-name">{{ productDisplayName(product) }}</span>
              </span>
              <span
                v-if="productCategoryText(product)"
                class="gp-search-suggestion-category"
              >{{ productCategoryText(product) }}</span>
            </button>
            <button
              v-if="canExpandSearchSuggestions"
              type="button"
              class="gp-search-suggestions-more"
              @mousedown.prevent="expandSearchSuggestions"
            >{{ t('offers.list.showMoreSuggestions') }}</button>
            <p
              v-if="!suggestionsLoading && searchSuggestions.length === 0"
              class="gp-search-suggestion-state"
            >{{ t('offers.list.noSuggestions') }}</p>
          </div>
        </div>
        <button
          type="submit"
          class="gp-search-btn"
        >
          <span class="gp-search-button-label">{{ t('offers.list.search') }}</span>
          <AppIcon
            class="gp-mobile-search-icon"
            name="search"
            :size="19"
          />
        </button>
        <button
          type="button"
          class="gp-fav-btn"
          :class="showFavoritesOnly ? 'on' : ''"
          :disabled="favoritesLoading"
          @click="toggleFavoritesOnly"
        >
          {{ favoritesLoading ? t('offers.list.loading') : t('offers.list.myFavorites') }}
        </button>
      </form>
      <div
        class="gp-mobile-filter-rail"
        :aria-label="t('offers.list.offerFilters')"
      >
        <select
          v-model="activeSort"
          class="gp-mobile-filter-select gp-mobile-sort-select"
          :aria-label="t('offers.list.sortMethod')"
          @change="selectSort"
        >
          <option
            v-for="option in sortOptions"
            :key="`mobile-sort-${option.value}`"
            :value="option.value"
          >{{ option.label }}</option>
        </select>
        <select
          :value="activeCategory"
          class="gp-mobile-filter-select"
          :aria-label="t('offers.list.category')"
          @change="handleMobileQuickFilter('category', $event)"
        >
          <option value="">{{ t('offers.list.category') }}</option>
          <option
            v-for="category in categoryPills.slice(1)"
            :key="`mobile-category-${category.value}`"
            :value="category.value"
          >{{ category.label }}</option>
        </select>
        <select
          :value="activeStore"
          class="gp-mobile-filter-select"
          :aria-label="t('offers.list.store')"
          @change="handleMobileQuickFilter('store', $event)"
        >
          <option value="">{{ t('offers.list.store') }}</option>
          <option
            v-for="store in storePills.slice(1)"
            :key="`mobile-store-${store.value}`"
            :value="store.value"
          >{{ store.label }}</option>
        </select>
        <button
          type="button"
          class="gp-mobile-filter-button mobile-filter-button"
          aria-controls="offers-mobile-filter-sheet"
          :aria-expanded="isMobileFilterOpen"
          @click="openMobileFilter"
        >
          <AppIcon
            name="filter"
            :size="16"
          />
          <span>{{ t('offers.list.more') }}</span>
          <span
            v-if="activeFilterCount > 0"
            class="gp-mobile-filter-count"
          >{{ activeFilterCount }}</span>
        </button>
        <button
          type="button"
          class="gp-mobile-favorite-filter"
          :class="showFavoritesOnly ? 'on' : ''"
          :aria-pressed="showFavoritesOnly"
          :disabled="favoritesLoading"
          @click="toggleFavoritesOnly"
        >
          <AppIcon
            name="star"
            :size="16"
          />
          <span>{{ t('offers.list.favorite') }}</span>
        </button>
        <button
          v-if="activeFilterCount > 0"
          type="button"
          class="gp-mobile-filter-reset"
          @click="resetFilters"
        >{{ t('offers.list.reset') }}</button>
      </div>
      <div class="gp-filter-pills">
        <button
          v-for="category in categoryPills"
          :key="category.value || 'all-category'"
          type="button"
          class="gp-fpill"
          :class="activeCategory === category.value ? 'on' : ''"
          @click="selectCategory(category.value)"
        >
          {{ category.label }}
        </button>
      </div>
      <div class="gp-store-pills">
        <button
          v-for="store in storePills"
          :key="store.value || 'all-store'"
          type="button"
          class="gp-spill"
          :class="activeStore === store.value ? 'on' : ''"
          @click="selectStore(store.value)"
        >
          {{ store.label }}
        </button>
      </div>
    </section>

    <!-- 3. 內容區 -->
    <section class="gp-content">
      <div class="gp-content-header">
        <span class="gp-count">
          {{ searchLoading
            ? t('offers.list.loading')
            : t('offers.list.offerCount', { count: formatInteger(totalCount) }) }}
        </span>
        <div class="gp-toolbar">
          <select
            v-model="activeBrand"
            class="gp-sort gp-brand-select"
            @change="selectBrand"
          >
            <option value="">{{ t('offers.list.allBrands') }}</option>
            <option
              v-for="brand in brandOptions"
              :key="brand"
              :value="brand"
            >
              {{ brand }}
            </option>
          </select>
          <div class="gp-view-toggle">
            <button
              type="button"
              class="gp-view-btn"
              :class="viewMode === 'grid' ? 'on' : ''"
              @click="setView('grid')"
            >
              {{ t('offers.list.grid') }}
            </button>
            <button
              type="button"
              class="gp-view-btn"
              :class="viewMode === 'table' ? 'on' : ''"
              @click="setView('table')"
            >
              {{ t('offers.list.table') }}
            </button>
          </div>
          <select
            v-model="activeSort"
            class="gp-sort gp-order-select"
            @change="selectSort"
          >
            <option
              v-for="option in sortOptions"
              :key="option.value"
              :value="option.value"
            >
              {{ option.label }}
            </option>
          </select>
        </div>
      </div>

      <p
        v-if="summaryError"
        class="gp-state gp-state-error"
      >
        {{ summaryError }}
      </p>
      <p
        v-if="searchError"
        class="gp-state gp-state-error"
      >
        {{ searchError }}
      </p>
      <p
        v-if="actionMessage"
        class="gp-state"
      >
        {{ actionMessage }}
      </p>
      <p
        v-if="searchLoading && !searchResult"
        class="gp-state"
      >
        {{ t('offers.list.loadingProducts') }}
      </p>
      <p
        v-else-if="!searchLoading && visibleProducts.length === 0"
        class="gp-state"
      >
        {{ t('offers.list.empty') }}
      </p>

      <!-- 3.1 網格視圖 -->
      <div
        v-if="visibleProducts.length > 0 && viewMode === 'grid'"
        class="gp-grid"
      >
        <article
          v-for="product in visibleProducts"
          :key="product.code"
          class="gp-card"
          @click="openDetail(product)"
        >
          <div class="gp-card-media">
            <div class="gp-card-img">
              <button
                type="button"
                class="gp-card-fav"
                :class="product.isFavorite ? 'on' : ''"
                :aria-label="product.isFavorite ? t('offers.list.favorited') : t('offers.list.favorite')"
                @click.stop="toggleFavorite(product)"
              >
                <AppIcon
                  name="heart"
                  :size="16"
                />
              </button>
              <img
                v-if="(product.image_url || product.imageUrl) && !failedImageCodes.has(product.code)"
                :src="product.image_url || product.imageUrl"
                :alt="productDisplayTitle(product)"
                @error="handleProductImageError(product.code)"
              >
              <div
                v-else
                class="gp-card-img-ph"
              />
            </div>
          </div>
          <div class="gp-card-body">
            <div class="gp-card-title-row">
              <span
                v-if="productDisplayBrand(product)"
                class="gp-card-brand-chip"
              >{{ productDisplayBrand(product) }}</span>
              <div class="gp-card-name">
                <span>{{ productTitleName(product) }}</span>
                <span
                  v-if="productDisplayUnit(product)"
                  class="gp-card-unit"
                >{{ productDisplayUnit(product) }}</span>
              </div>
            </div>
            <div class="gp-card-best">
              <div class="gp-card-best-copy">
                <div class="gp-card-best-badge">{{ productBestBadgeText(productFeaturedPrice(product)) }}</div>
                <div class="gp-card-best-store">{{ formatStore(productFeaturedPrice(product).store) }}</div>
                <div class="gp-card-best-offer">{{ productFeaturedOfferText(product, productFeaturedPrice(product)) }}</div>
              </div>
              <div class="gp-card-best-price">
                <div class="gp-card-best-amount">{{ formatOfferPrice(productFeaturedPrice(product).effectiveUnitPrice) }}</div>
                <div
                  v-if="productFeaturedCompareText(product, productFeaturedPrice(product))"
                  class="gp-card-best-compare"
                >{{ productFeaturedCompareText(product, productFeaturedPrice(product)) }}</div>
              </div>
            </div>
            <div
              v-if="productOtherPrices(product).length > 0"
              class="gp-card-others"
            >
              <div
                v-for="price in productOtherPrices(product)"
                :key="`${product.code}-${price.store}-${price.effectiveUnitPrice}`"
                class="gp-card-other-row"
              >
                <span>{{ formatStore(price.store) }}</span>
                <span>{{ productStoreOfferLabel(price) }}</span>
                <strong>{{ formatPlainDollar(price.effectiveUnitPrice) }}</strong>
              </div>
            </div>
          </div>
        </article>
      </div>

      <!-- 3.2 表格視圖 -->
      <div
        v-if="visibleProducts.length > 0 && viewMode === 'table'"
        class="gp-table"
      >
        <table>
          <thead>
            <tr>
              <th>{{ t('offers.list.product') }}</th>
              <th>{{ t('offers.list.store') }}</th>
              <th>{{ t('offers.list.offer') }}</th>
              <th>{{ t('offers.list.effectivePrice') }}</th>
              <th>{{ t('offers.list.favorite') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="product in visibleProducts"
              :key="product.code"
              @click="openDetail(product)"
            >
              <td>
                <div class="gp-table-product">
                  <div class="gp-table-product-head">
                    <span
                      v-if="productDisplayBrand(product)"
                      class="gp-table-brand-chip"
                    >{{ productDisplayBrand(product) }}</span>
                    <strong>{{ productDisplayName(product) }}</strong>
                  </div>
                </div>
              </td>
              <td>{{ formatStore(productPrimaryPrice(product).store) }}</td>
              <td>{{ productOfferTexts(product).join(' / ') || '-' }}</td>
              <td>{{ formatOfferUnitPrice(productPrimaryPrice(product).effectiveUnitPrice) }}</td>
              <td>
                <button
                  type="button"
                  class="gp-table-fav"
                  @click.stop="toggleFavorite(product)"
                >
                  {{ product.isFavorite ? t('offers.list.favorited') : t('offers.list.favorite') }}
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- 3.3 分頁 -->
      <div
        v-if="totalPages > 1"
        class="gp-pagination"
      >
        <button
          v-for="page in pageButtons"
          :key="page.key"
          type="button"
          class="gp-page-btn"
          :class="page.active ? 'on' : ''"
          :disabled="page.disabled"
          @click="selectPage(page)"
        >
          {{ page.label }}
        </button>
      </div>

      <div class="gp-updated-bar">
        {{ updatedBarText }}
      </div>
    </section>
  </main>
</template>

<style scoped>
/* 1. 頁面容器 */
.page {
  width: 100%;
  min-height: calc(100svh - 48px);
  background: var(--sur);
  color: var(--ink);
}

.gp-filter-backdrop,
.gp-filter-sheet,
.gp-mobile-filter-rail,
.gp-mobile-search-icon {
  display: none;
}

/* 2. 超市情報標題 */
.gp-hero {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
  min-height: 132px;
  max-width: 1440px;
  margin: 0 auto;
  padding: 34px 38px;
  border-bottom: 3px solid var(--accent);
  background: #fff;
  color: var(--brand);
}

.gp-hero-left {
  min-width: 0;
}

.gp-hero-label {
  color: rgba(255, 255, 255, 0.58);
  font-size: 12px;
}

.gp-hero-title {
  margin-top: 0;
  font-family: var(--font-serif);
  font-size: 34px;
  font-weight: 400;
  line-height: 1.12;
  color: var(--brand);
}

.gp-hero-sub {
  margin-top: 8px;
  max-width: 480px;
  color: var(--brand);
  font-size: 13px;
  line-height: 1.6;
}

.gp-hero-right {
  display: flex;
  align-items: center;
  gap: 0;
  flex-shrink: 0;
}

.gp-hstat {
  min-width: 118px;
  padding: 0 26px;
  text-align: center;
}

.gp-hnum {
  display: block;
  color: #fff;
  font-size: 28px;
  font-weight: 300;
  line-height: 1;
}

.gp-hlabel {
  display: block;
  margin-top: 6px;
  color: rgba(255, 255, 255, 0.68);
  font-size: 11px;
}

.gp-hdiv {
  width: 1px;
  height: 42px;
  background: rgba(255, 255, 255, 0.26);
}

/* 3. 控制欄 */
.gp-controls {
  max-width: 1440px;
  margin: 0 auto;
  padding: 18px 28px 14px;
  border-bottom: 1px solid var(--bdr);
  background: rgb(var(--color-surface));
}

.gp-search-row {
  display: flex;
  align-items: stretch;
  justify-content: center;
  gap: 8px;
  margin-bottom: 4px;
}

.gp-search-box {
  position: relative;
  display: flex;
  max-width: 520px;
  align-items: center;
  gap: 8px;
  flex: 0 1 520px;
  border: 1px solid var(--bdr);
  border-radius: 3px;
  background: rgb(var(--color-surface));
  padding: 0 13px;
}

.gp-search-suggestions {
  position: absolute;
  z-index: 5;
  top: calc(100% + 4px);
  right: 0;
  left: 0;
  overflow: hidden;
  border: 1px solid var(--bdr);
  border-radius: 3px;
  background: rgb(var(--color-surface));
  box-shadow: 0 8px 20px rgba(0, 0, 0, 0.12);
}

.gp-search-suggestion,
.gp-search-suggestion-state {
  display: block;
  width: 100%;
  box-sizing: border-box;
  margin: 0;
  border: 0;
  border-bottom: 1px solid var(--bdr);
  background: transparent;
  color: var(--ink);
  font: inherit;
  text-align: left;
}

.gp-search-suggestion {
  cursor: pointer;
  padding: 9px 12px;
}

.gp-search-suggestion:hover {
  background: var(--sur-2);
}

.gp-search-suggestions-more {
  display: block;
  width: 100%;
  border: 0;
  border-bottom: 1px solid var(--bdr);
  background: transparent;
  color: var(--brand);
  cursor: pointer;
  font: inherit;
  font-size: 13px;
  font-weight: 600;
  padding: 10px 12px;
  text-align: center;
}

.gp-search-suggestions-more:hover {
  background: var(--sur-2);
}

.gp-search-suggestion:last-child,
.gp-search-suggestion-state:last-child,
.gp-search-suggestions-more:last-child {
  border-bottom: 0;
}

.gp-search-suggestion-head {
  display: flex;
  align-items: baseline;
  gap: 8px;
}

.gp-search-suggestion-name {
  display: block;
  min-width: 0;
  flex: 1 1 auto;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  line-height: 1.2;
}

.gp-search-suggestion-chip {
  display: inline-flex;
  flex: 0 0 auto;
  border: 1px solid var(--bdr);
  border-radius: 999px;
  background: var(--sur-2);
  color: var(--ink-2);
  font-size: 13px;
  font-weight: 600;
  line-height: 1.2;
  padding: 0 8px;
}

.gp-search-suggestion-name {
  font-size: 13px;
  font-weight: 700;
}

.gp-search-suggestion-category,
.gp-search-suggestion-state {
  display: block;
  color: var(--ink-3);
  font-size: 12px;
  line-height: 1.45;
}

.gp-search-suggestion-category {
  margin-top: 2px;
}

.gp-search-suggestion-state {
  padding: 10px 12px;
}

.gp-search-ico {
  color: var(--ink-3);
  font-size: 21px;
  line-height: 1;
}

.gp-sinput {
  flex: 1;
  min-width: 0;
  border: 0;
  box-shadow: none;
  background: transparent;
  font: inherit;
  font-size: 14px;
  color: var(--ink);
  outline: 0;
  padding: 12px 0;
}

.gp-search-btn,
.gp-fav-btn,
.gp-view-btn {
  border: 1px solid var(--bdr);
  border-radius: 3px;
  background: rgb(var(--color-surface));
  color: var(--ink);
  cursor: pointer;
  font-family: inherit;
  font-size: 12px;
  font-weight: 600;
  padding: 0 18px;
}

.gp-search-btn {
  border-color: var(--accent);
  background: var(--accent);
  color: #fff;
}

.gp-fav-btn.on,
.gp-view-btn.on {
  border-color: var(--brand-mid);
  background: var(--brand-light);
  color: var(--accent);
}

.gp-search-btn:disabled,
.gp-fav-btn:disabled,
.gp-page-btn:disabled {
  cursor: not-allowed;
  opacity: 0.52;
}

.gp-filter-pills,
.gp-store-pills {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 8px;
  margin-top: 6px;
}

.gp-fpill,
.gp-spill {
  border: 1px solid var(--bdr);
  border-radius: 999px;
  background: rgb(var(--color-surface));
  color: var(--ink-3);
  cursor: pointer;
  font-family: inherit;
  font-size: 13px;
  padding: 7px 15px;
}

.gp-fpill.on,
.gp-spill.on {
  background: var(--accent);
  color: var(--white);
  border-color: var(--accent);
  font-weight: 700;
}

/* 4. 內容區 */
.gp-content {
  max-width: 1440px;
  margin: 0 auto;
  padding: 26px 0 44px;
  background: var(--sur);
}

.gp-content-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 20px;
  padding: 0 28px;
}

.gp-count {
  color: var(--ink-3);
  font-size: 12px;
  font-weight: 400;
}

.gp-toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
}

.gp-view-toggle {
  display: flex;
  overflow: hidden;
  border: 0;
  gap: 6px;
}

.gp-view-toggle .gp-view-btn {
  border: 1px solid var(--bdr);
  border-radius: 2px;
  padding: 7px 13px;
  background: rgb(var(--color-surface));
}

.gp-view-toggle .gp-view-btn.on {
  background: rgb(var(--color-surface));
  color: var(--ink);
  border-color: var(--bdr);
}

.gp-sort {
  max-width: 180px;
  border: 1px solid var(--bdr);
  border-radius: 2px;
  background: rgb(var(--color-surface));
  color: var(--ink);
  font-family: inherit;
  font-size: 12px;
  padding: 7px 26px 7px 10px;
  outline: 0;
}

.gp-state {
  margin: 0 0 16px;
  border: 1px solid var(--bdr);
  border-radius: 3px;
  background: rgb(var(--color-surface));
  color: var(--ink-3);
  font-size: 13px;
  padding: 12px 14px;
}

.gp-state-error {
  border-color: rgba(186, 26, 26, 0.24);
  color: #ba1a1a;
}

/* 5. 卡片網格 */
.gp-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  align-items: stretch;
  gap: 14px;
}

.gp-card {
  --gp-media-width: 148px;
  --gp-media-hover-width: 248px;
  position: relative;
  z-index: 1;
  display: grid;
  grid-template-columns: var(--gp-media-width) minmax(0, 1fr);
  min-height: 0;
  margin: 0;
  isolation: isolate;
  border: 1px solid var(--bdr);
  border-radius: var(--r-xl);
  overflow: hidden;
  background: rgb(var(--color-surface));
  cursor: pointer;
  transition: border-color 0.15s ease, box-shadow 0.15s ease;
}

.gp-card:hover {
  z-index: 6;
  border-color: var(--accent);
  box-shadow: 0 12px 28px rgb(32 48 61 / 0.16);
}

.gp-card-media {
  position: relative;
  z-index: 2;
  justify-self: start;
  align-self: stretch;
  width: var(--gp-media-width);
  height: 100%;
  overflow: visible;
  background: #fff;
}

.gp-card-img {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  min-height: 168px;
  padding: 28px 12px 16px;
  box-sizing: border-box;
  overflow: hidden;
  background: #fff;
  transition: width 0.22s ease;
}

.gp-card-img img {
  width: auto;
  height: auto;
  max-width: 88%;
  max-height: 132px;
  object-fit: contain;
  display: block;
  transform-origin: center center;
  transition: max-width 0.22s ease, max-height 0.22s ease;
}

@media (hover: hover) and (pointer: fine) {
  .gp-card-img:hover {
    position: absolute;
    top: 0;
    bottom: 0;
    left: 0;
    z-index: 3;
    width: var(--gp-media-hover-width);
    height: 100%;
  }

  .gp-card-img:hover img {
    max-width: 100%;
    max-height: 176px;
  }
}

.gp-card-img-ph {
  width: 64px;
  height: 64px;
  border-radius: 50%;
  background: linear-gradient(135deg, var(--brand-mid), var(--accent));
  opacity: 0.6;
}

.gp-card-fav {
  position: absolute;
  top: 10px;
  right: 10px;
  z-index: 3;
  display: inline-flex;
  width: 32px;
  height: 32px;
  align-items: center;
  justify-content: center;
  border: 0;
  border-radius: 999px;
  background: rgb(var(--color-surface));
  color: var(--ink-3);
  box-shadow: 0 2px 8px rgb(26 26 26 / 0.08);
  cursor: pointer;
  padding: 0;
  pointer-events: auto;
}

.gp-card-fav.on {
  color: var(--brand);
}

.gp-card-fav.on :deep(.app-icon) {
  fill: currentColor;
}

.gp-card-body {
  position: relative;
  z-index: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 14px 16px 14px 14px;
}

.gp-card-title-row {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 8px;
}

.gp-card-name {
  display: -webkit-box;
  min-width: 0;
  flex: 1 1 0;
  margin: 0;
  color: var(--ink);
  font-size: 16px;
  font-weight: 700;
  line-height: 1.35;
  overflow: hidden;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.gp-card-unit {
  margin-left: 6px;
  color: var(--ink-3);
  font-size: 16px;
  font-weight: 500;
}

.gp-card-brand-chip {
  display: inline-flex;
  flex: 0 0 auto;
  width: fit-content;
  margin: 0;
  border: 0;
  border-radius: 6px;
  background: var(--ink);
  color: #fff;
  font-size: 16px;
  font-weight: 700;
  line-height: 1.2;
  padding: 3px 8px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.gp-card-best {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 12px;
  align-items: start;
  border-radius: 14px;
  background: var(--brand-light);
  padding: 12px 14px;
}

.gp-card-best-copy {
  min-width: 0;
}

.gp-card-best-badge {
  display: inline-block;
  max-width: 100%;
  border-radius: 999px;
  background: var(--brand);
  color: #fff;
  font-size: 11px;
  font-weight: 700;
  line-height: 1.3;
  padding: 5px 10px;
}

.gp-card-best-store {
  margin-top: 10px;
  color: var(--ink);
  font-size: 18px;
  font-weight: 800;
  line-height: 1.2;
}

.gp-card-best-offer {
  margin-top: 4px;
  color: var(--ink-2);
  font-size: 12px;
  line-height: 1.4;
}

.gp-card-best-price {
  text-align: right;
  white-space: nowrap;
}

.gp-card-best-amount {
  color: var(--ink);
  font-size: 22px;
  font-weight: 800;
  letter-spacing: -0.04em;
  line-height: 1;
}

.gp-card-best-compare {
  margin-top: 8px;
  color: var(--ink-3);
  font-size: 12px;
  font-weight: 600;
}

.gp-card-others {
  display: grid;
  gap: 8px;
  margin-top: auto;
}

.gp-card-other-row {
  display: grid;
  grid-template-columns: minmax(72px, 0.9fr) minmax(0, 1.4fr) auto;
  gap: 10px;
  align-items: center;
  color: var(--ink-2);
  font-size: 13px;
}

.gp-card-other-row span:nth-child(2) {
  color: var(--ink-3);
}

.gp-card-other-row strong {
  color: var(--ink);
  font-size: 14px;
  font-weight: 700;
  text-align: right;
  white-space: nowrap;
}

.gp-table-product {
  display: grid;
  gap: 4px;
}

.gp-table-product-head {
  display: flex;
  flex-wrap: nowrap;
  align-items: baseline;
  gap: 8px;
}

.gp-table-product-head strong {
  min-width: 0;
  flex: 1 1 auto;
  line-height: 1.4;
}

.gp-table-brand-chip {
  display: inline-flex;
  flex: 0 0 auto;
  border: 1px solid var(--bdr);
  border-radius: 999px;
  background: var(--sur-2);
  color: var(--ink-2);
  font-size: 13px;
  font-weight: 600;
  line-height: 1.2;
  padding: 2px 8px;
}

/* 6. 表格視圖 */
.gp-table {
  overflow: auto;
  border: 1px solid var(--bdr);
  border-radius: 3px;
  background: rgb(var(--color-surface));
}

.gp-table table {
  width: 100%;
  border-collapse: collapse;
  min-width: 760px;
}

.gp-table th {
  border-bottom: 1px solid var(--bdr);
  color: var(--ink-3);
  font-size: 12px;
  font-weight: 600;
  padding: 12px;
  text-align: left;
}

.gp-table td {
  border-bottom: 1px solid var(--sur-3);
  color: var(--ink-2);
  font-size: 12px;
  padding: 12px;
  vertical-align: middle;
}

.gp-table tr {
  cursor: pointer;
}

.gp-table tr:hover td {
  background: #fffaf7;
}

.gp-table strong {
  display: block;
  color: var(--ink);
  font-size: 13px;
}

.gp-table span {
  display: block;
  margin-top: 3px;
  color: var(--ink-3);
  font-size: 11px;
}

.gp-table .gp-table-brand-chip {
  display: inline-flex;
  flex: 0 0 auto;
  align-self: center;
  margin-top: 0;
  font-size: 13px;
  line-height: 1.4;
  padding: 0 8px;
  white-space: nowrap;
}

.gp-table-fav {
  border: 1px solid var(--bdr);
  border-radius: 2px;
  background: rgb(var(--color-surface));
  color: var(--ink-2);
  cursor: pointer;
  font-family: inherit;
  font-size: 11px;
  font-weight: 600;
  padding: 5px 9px;
}

/* 7. 分頁 */
.gp-pagination {
  display: flex;
  justify-content: center;
  gap: 6px;
  margin-top: 22px;
}

.gp-page-btn {
  border: 1px solid var(--bdr);
  border-radius: 3px;
  background: rgb(var(--color-surface));
  color: var(--ink-3);
  cursor: pointer;
  font-family: inherit;
  font-size: 12px;
  min-width: 36px;
  padding: 8px 12px;
}

.gp-page-btn.on {
  border-color: var(--bdr);
  background: rgb(var(--color-surface));
  color: var(--ink);
}

/* 8. 更新資訊列 */
.gp-updated-bar {
  margin-top: 18px;
  color: var(--ink-3);
  font-size: 11px;
  text-align: center;
}

/* 9. 響應式 */
@media (min-width: 1440px) {
  .gp-hero,
  .gp-controls,
  .gp-content {
    max-width: 1480px;
  }

  .gp-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}

@media (max-width: 1023px) {
  .gp-hero,
  .gp-filter-pills,
  .gp-store-pills,
  .gp-search-row .gp-fav-btn,
  .gp-brand-select,
  .gp-view-toggle,
  .gp-toolbar {
    display: none;
  }

  .gp-controls,
  .gp-content {
    max-width: none;
    margin: 0;
  }

  .gp-controls {
    padding: 12px;
  }

  .gp-search-row {
    display: grid;
    grid-template-columns: minmax(0, 1fr) 48px;
    gap: 0;
    margin: 0;
  }

  .gp-search-box {
    width: 100%;
    max-width: none;
    min-height: 48px;
    flex: 1 1 auto;
    border-radius: 8px 0 0 8px;
    padding: 0 13px;
  }

  .gp-sinput {
    min-height: 48px;
    font-size: 16px;
    padding: 0;
  }

  .gp-search-btn {
    display: inline-flex;
    width: 48px;
    min-height: 48px;
    align-items: center;
    justify-content: center;
    border-radius: 0 8px 8px 0;
    padding: 0;
  }

  .gp-search-button-label {
    display: none;
  }

  .gp-mobile-search-icon {
    display: block;
  }

  .gp-mobile-filter-rail {
    display: flex;
    gap: 8px;
    margin: 10px -2px 0;
    overflow-x: auto;
    overscroll-behavior-x: contain;
    padding: 2px;
    scrollbar-width: none;
  }

  .gp-mobile-filter-rail::-webkit-scrollbar {
    display: none;
  }

  .gp-mobile-filter-select,
  .gp-mobile-filter-button,
  .gp-mobile-favorite-filter {
    min-height: 40px;
    border: 1px solid var(--bdr);
    border-radius: 999px;
    background: var(--sur);
    color: var(--ink-2);
    font: inherit;
    font-size: 13px;
  }

  .gp-mobile-filter-select {
    width: auto;
    min-width: 104px;
    flex: 0 0 auto;
    padding: 0 30px 0 13px;
  }

  .gp-mobile-sort-select {
    min-width: 120px;
  }

  .gp-mobile-filter-button,
  .gp-mobile-favorite-filter {
    display: inline-flex;
    flex: 0 0 auto;
    align-items: center;
    gap: 4px;
    padding: 0 12px;
    cursor: pointer;
  }

  .gp-mobile-favorite-filter.on {
    border-color: var(--brand-mid);
    background: var(--brand-light);
    color: var(--accent);
  }

  .gp-mobile-favorite-filter:disabled {
    cursor: not-allowed;
    opacity: 0.52;
  }

  .gp-mobile-filter-count {
    display: inline-flex;
    min-width: 18px;
    height: 18px;
    align-items: center;
    justify-content: center;
    border-radius: 999px;
    background: var(--brand);
    color: #fff;
    font-size: 10px;
    line-height: 1;
  }

  .gp-mobile-filter-reset {
    flex: 0 0 auto;
    min-height: 40px;
    border: 0;
    background: transparent;
    color: var(--ink-3);
    cursor: pointer;
    font: inherit;
    font-size: 13px;
    padding: 0 4px;
  }

  .gp-content {
    padding: 14px 12px calc(32px + var(--app-safe-bottom));
  }

  .gp-content-header {
    align-items: center;
    gap: 8px;
    margin-bottom: 14px;
  }

  .gp-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 10px;
  }

  .gp-card {
    --gp-media-width: 120px;
    --gp-media-hover-width: 200px;
    grid-template-columns: var(--gp-media-width) minmax(0, 1fr);
  }

  .gp-card-img {
    min-height: 148px;
    padding: 24px 12px 12px;
  }

  .gp-card-body {
    padding: 12px;
  }

  .gp-card-best-amount {
    font-size: 24px;
  }

  .gp-filter-backdrop {
    position: fixed;
    z-index: 120;
    inset: 0;
    display: block;
    width: 100%;
    height: 100%;
    border: 0;
    background: rgb(0 0 0 / 0.44);
    cursor: pointer;
  }

  .gp-filter-sheet {
    position: fixed;
    z-index: 121;
    right: 0;
    bottom: 0;
    left: 0;
    display: flex;
    width: 100%;
    max-height: min(82svh, 720px);
    min-height: 0;
    box-sizing: border-box;
    flex-direction: column;
    border-radius: 8px 8px 0 0;
    background: var(--sur);
    box-shadow: 0 -12px 32px rgb(0 0 0 / 0.18);
    opacity: 0;
    overflow-y: auto;
    pointer-events: none;
    touch-action: pan-y;
    transform: translateY(100%);
    transition: transform 0.2s ease, opacity 0.18s ease, visibility 0.2s;
    visibility: hidden;
  }

  .gp-filter-sheet.is-filter-open {
    opacity: 1;
    pointer-events: auto;
    transform: translateY(0);
    visibility: visible;
  }

  .gp-filter-sheet-header {
    position: sticky;
    z-index: 2;
    top: 0;
    display: flex;
    min-height: 58px;
    align-items: center;
    justify-content: space-between;
    border-bottom: 1px solid var(--bdr);
    padding: 8px 12px 8px 14px;
    background: var(--sur);
  }

  .gp-filter-sheet-header h2 {
    margin: 0;
    color: var(--ink);
    font-size: 16px;
    font-weight: 700;
  }

  .gp-filter-sheet-close {
    display: inline-flex;
    width: 44px;
    height: 44px;
    align-items: center;
    justify-content: center;
    border: 0;
    border-radius: 6px;
    background: transparent;
    color: var(--ink);
    cursor: pointer;
  }

  .gp-filter-sheet-body {
    padding: 12px 14px 4px;
  }

  .gp-filter-section {
    margin-bottom: 16px;
  }

  .gp-filter-section h3 {
    margin: 0 0 7px;
    color: var(--ink-3);
    font-size: 12px;
    font-weight: 700;
  }

  .gp-filter-tags {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
  }

  .gp-filter-tag {
    min-height: 36px;
    border: 1px solid var(--bdr);
    border-radius: 999px;
    background: var(--sur);
    color: var(--ink-2);
    cursor: pointer;
    font: inherit;
    font-size: 13px;
    padding: 4px 11px;
  }

  .gp-filter-tag.on {
    border-color: var(--accent);
    background: var(--accent);
    color: #fff;
  }

  .gp-filter-select-field {
    display: grid;
    grid-template-columns: 74px minmax(0, 1fr);
    min-height: 48px;
    align-items: center;
    gap: 12px;
    border-top: 1px solid var(--sur-3);
    color: var(--ink-2);
    font-size: 13px;
  }

  .gp-filter-select-field select {
    min-width: 0;
    min-height: 38px;
    border: 1px solid var(--bdr);
    border-radius: 6px;
    background: var(--sur);
    color: var(--ink);
    font: inherit;
    font-size: 13px;
    padding: 0 10px;
  }

  .gp-view-filter-section {
    margin: 14px 0 10px;
  }

  .gp-view-filter-actions {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 8px;
  }

  .gp-view-filter-actions button {
    min-height: 40px;
    border: 1px solid var(--bdr);
    border-radius: 6px;
    background: var(--sur);
    color: var(--ink-2);
    cursor: pointer;
    font: inherit;
    font-size: 13px;
    font-weight: 700;
  }

  .gp-view-filter-actions button.on {
    border-color: var(--accent);
    background: var(--accent);
    color: #fff;
  }

  .gp-filter-sheet-actions {
    position: sticky;
    z-index: 2;
    bottom: 0;
    display: grid;
    grid-template-columns: minmax(0, 0.8fr) minmax(0, 1.4fr);
    gap: 8px;
    border-top: 1px solid var(--bdr);
    padding: 10px max(14px, calc(14px + var(--app-safe-right))) calc(10px + var(--app-safe-bottom)) max(14px, calc(14px + var(--app-safe-left)));
    background: var(--sur);
  }

  .gp-filter-sheet-reset,
  .gp-filter-sheet-apply {
    min-height: 44px;
    border-radius: 6px;
    cursor: pointer;
    font: inherit;
    font-size: 13px;
    font-weight: 700;
  }

  .gp-filter-sheet-reset {
    border: 1px solid var(--bdr);
    background: var(--sur);
    color: var(--ink-2);
  }

  .gp-filter-sheet-apply {
    border: 1px solid var(--accent);
    background: var(--accent);
    color: #fff;
  }
}

@media (max-width: 560px) {
  .gp-grid {
    grid-template-columns: 1fr;
  }

  .gp-card {
    --gp-media-width: 108px;
    --gp-media-hover-width: 176px;
    grid-template-columns: var(--gp-media-width) minmax(0, 1fr);
  }

  .gp-card-best {
    grid-template-columns: 1fr;
    gap: 8px;
  }

  .gp-card-best-price {
    text-align: left;
  }

  .gp-card-other-row {
    grid-template-columns: minmax(64px, 0.8fr) minmax(0, 1fr) auto;
    gap: 6px;
    font-size: 12px;
  }
}
</style>

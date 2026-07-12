<!--
 * 綜合優惠頻道頁。
 * 1. 使用 AJO 後端超市優惠摘要與搜尋接口。
 * 2. 提供商品搜尋、分類、品牌、商店、排序、收藏與分頁。
 * 3. 桌面保留摘要 Hero；手機以搜尋、快捷篩選與底部篩選彈窗呈現。
-->
<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';

import {
  addSupermarketFavorite,
  fetchSupermarketFavorites,
  fetchSupermarketSummary,
  removeSupermarketFavorite,
  searchSupermarketProducts,
} from '@/httpapis/supermarket-offers';
import { readStoredAccessToken } from '@/httpapis/auth-session';
import AppIcon from '@/shared/components/base/AppIcon.vue';
import type {
  SupermarketProduct,
  SupermarketSearchParams,
  SupermarketSearchResult,
  SupermarketSummary,
  SupermarketValueCount,
} from '@/model/supermarket-offers';
import {
  displaySupermarketCategory,
  displaySupermarketStore,
  formatSupermarketDate,
  formatSupermarketHKPrice,
  supermarketOfferTexts,
  supermarketPriceDiscountRate,
  supermarketPrimaryPrice,
  supermarketStorePrices,
} from '@/utils/supermarket-offers';

type ViewMode = 'grid' | 'table';

interface HeroStat {
  value: string;
  label: string;
}

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
const activeCategory = ref('');
const activeStore = ref('');
const activeBrand = ref('');
const activeSort = ref('discount');
const showFavoritesOnly = ref(false);
const viewMode = ref<ViewMode>('grid');
const currentPage = ref(1);
const isMobileFilterOpen = ref(false);

const sortOptions: SortOption[] = [
  { label: '優惠力度 ↓', value: 'discount' },
  { label: '優惠後價格', value: 'effective' },
  { label: '價差', value: 'diff' },
  { label: '商品名稱', value: 'name' },
  { label: '品牌', value: 'brand' },
];

// 1. 建立 Hero 統計資料
const heroStats = computed<HeroStat[]>(() => {
  const stats = summary.value?.stats;
  const discountSource = summary.value?.bestDiscounts?.length
    ? summary.value.bestDiscounts
    : summary.value?.offers ?? [];
  const maxDiscount = discountSource.reduce((result, product) => {
    const rate = supermarketPriceDiscountRate(supermarketPrimaryPrice(product));
    return Math.max(result, rate);
  }, 0);

  return [
    { value: formatInteger(stats?.offers), label: '活躍優惠' },
    { value: formatInteger(stats?.stores), label: '連鎖商店' },
    { value: maxDiscount > 0 ? `-${maxDiscount.toFixed(0)}%` : '-', label: '最高折扣' },
  ];
});

// 2. 建立分類篩選項
const categoryPills = computed<FilterPill[]>(() => [
  { label: '全部', value: '' },
  ...valueCountPills(summary.value?.categories ?? [], displaySupermarketCategory, 9),
]);

// 3. 建立商店篩選項
const storePills = computed<FilterPill[]>(() => [
  { label: '全部商店', value: '' },
  ...valueCountPills(summary.value?.stores ?? [], displaySupermarketStore, 9),
]);

// 4. 建立品牌篩選項
const brandOptions = computed<string[]>(() => searchResult.value?.brands ?? []);

// 5. 計算目前已套用的手機端篩選數量
const activeFilterCount = computed(() =>
  Number(Boolean(activeCategory.value))
  + Number(Boolean(activeStore.value))
  + Number(Boolean(activeBrand.value))
  + Number(showFavoritesOnly.value),
);

// 6. 取得目前列表商品
const visibleProducts = computed<SupermarketProduct[]>(() => {
  if (showFavoritesOnly.value) {
    const start = (currentPage.value - 1) * pageSize;
    return favorites.value.slice(start, start + pageSize);
  }
  return searchResult.value?.items ?? [];
});

// 7. 取得目前總數
const totalCount = computed(() => (showFavoritesOnly.value ? favorites.value.length : searchResult.value?.total ?? 0));

// 8. 取得目前總頁數
const totalPages = computed(() => Math.max(1, Math.ceil(totalCount.value / pageSize)));

// 9. 建立分頁按鈕
const pageButtons = computed<PageButton[]>(() => {
  const pages = new Set<number>([1, totalPages.value, currentPage.value]);
  if (currentPage.value > 1) pages.add(currentPage.value - 1);
  if (currentPage.value < totalPages.value) pages.add(currentPage.value + 1);
  const numericPages = [...pages].filter((page) => page >= 1 && page <= totalPages.value).sort((a, b) => a - b);

  return [
    {
      label: '上一頁',
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
      label: '下一頁',
      key: 'next',
      page: Math.min(totalPages.value, currentPage.value + 1),
      active: false,
      disabled: currentPage.value >= totalPages.value,
    },
  ];
});

// 10. 建立更新資訊文字
const updatedBarText = computed(() => {
  const updatedDate = formatSupermarketDate(summary.value?.metadata?.latestSnapshotDate);
  const productCount = formatInteger(summary.value?.stats.products);
  const prefix = updatedDate ? `資料更新：${updatedDate}` : '資料更新：等待資料';
  return `${prefix} · 共監測 ${productCount} 件商品 · 價格只供參考，實際售價以商戶公布為準。`;
});

// 11. 載入摘要
const loadSummary = async (): Promise<void> => {
  summaryLoading.value = true;
  summaryError.value = '';
  try {
    const { data } = await fetchSupermarketSummary();
    summary.value = data.data;
  } catch {
    summaryError.value = '暫時無法載入超市優惠摘要。';
  } finally {
    summaryLoading.value = false;
  }
};

// 12. 載入搜尋結果
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
      offerOnly: true,
      sort: activeSort.value,
      page: currentPage.value,
      pageSize,
    };
    const { data } = await searchSupermarketProducts(params);
    searchResult.value = data.data;
  } catch {
    searchError.value = '暫時無法載入優惠商品。';
  } finally {
    searchLoading.value = false;
  }
};

// 13. 載入收藏
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
    searchError.value = '暫時無法載入我的收藏。';
  } finally {
    favoritesLoading.value = false;
  }
};

// 14. 提交搜尋
const submitSearch = (): void => {
  showFavoritesOnly.value = false;
  currentPage.value = 1;
  void loadSearch();
};

// 15. 切換分類
const selectCategory = (category: string): void => {
  activeCategory.value = category;
  showFavoritesOnly.value = false;
  currentPage.value = 1;
  void loadSearch();
};

// 16. 切換商店
const selectStore = (store: string): void => {
  activeStore.value = store;
  showFavoritesOnly.value = false;
  currentPage.value = 1;
  void loadSearch();
};

// 17. 套用手機端快捷篩選
const handleMobileQuickFilter = (filter: 'category' | 'store', event: Event): void => {
  const value = (event.target as HTMLSelectElement).value;
  if (filter === 'category') {
    selectCategory(value);
    return;
  }
  selectStore(value);
};

// 18. 開關手機端篩選彈窗
const openMobileFilter = (): void => {
  isMobileFilterOpen.value = true;
};

const closeMobileFilter = (): void => {
  isMobileFilterOpen.value = false;
};

// 19. 重設篩選條件並保留搜尋文字
const resetFilters = (): void => {
  activeCategory.value = '';
  activeStore.value = '';
  activeBrand.value = '';
  activeSort.value = 'discount';
  showFavoritesOnly.value = false;
  currentPage.value = 1;
  void loadSearch();
};

// 20. 切換品牌
const selectBrand = (): void => {
  showFavoritesOnly.value = false;
  currentPage.value = 1;
  void loadSearch();
};

// 21. 切換排序
const selectSort = (): void => {
  currentPage.value = 1;
  if (!showFavoritesOnly.value) {
    void loadSearch();
  }
};

// 22. 切換收藏列表
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

// 23. 切換商品收藏
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
      actionMessage.value = '已取消收藏。';
      return;
    }

    const { data } = await addSupermarketFavorite(product.code);
    patchFavoriteState(product.code, true);
    favorites.value = [data.data, ...favorites.value.filter((item) => item.code !== product.code)];
    actionMessage.value = '已加入收藏。';
  } catch {
    actionMessage.value = '收藏操作失敗。';
  }
};

// 24. 切換視圖模式
const setView = (mode: ViewMode): void => {
  viewMode.value = mode;
};

// 25. 切換分頁
const selectPage = (page: PageButton): void => {
  if (page.disabled || currentPage.value === page.page) {
    return;
  }
  currentPage.value = page.page;
  if (!showFavoritesOnly.value) {
    void loadSearch();
  }
};

// 26. 開啟商品詳情
const openDetail = (product: SupermarketProduct): void => {
  void router.push({ path: `/supermarket-offers/products/${encodeURIComponent(product.code)}` });
};

// 27. 前往登入
const openLogin = async (redirect: string): Promise<void> => {
  await router.push({ path: '/login', query: { redirect } });
};

// 28. 更新商品收藏狀態
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

// 29. 建立篩選按鈕
const valueCountPills = (
  values: SupermarketValueCount[],
  formatter: (value: string) => string,
  limit: number,
): FilterPill[] =>
  values
    .slice(0, limit)
    .map((item) => ({ label: formatter(item.value), value: item.value }))
    .filter((item) => item.value);

// 30. 格式化整數
const formatInteger = (value: number | undefined): string => (value ?? 0).toLocaleString('zh-HK');

onMounted(() => {
  void loadSummary();
  void loadSearch();
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
      aria-label="關閉篩選"
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
        <h2 id="offers-mobile-filter-title">篩選條件</h2>
        <button
          type="button"
          class="gp-filter-sheet-close filter-sheet-close"
          aria-label="關閉篩選"
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
          <h3>商品分類</h3>
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
          <h3>商店</h3>
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
          <span>品牌</span>
          <select
            v-model="activeBrand"
            @change="selectBrand"
          >
            <option value="">全部品牌</option>
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
          <span>排序</span>
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
          <h3>顯示方式</h3>
          <div class="gp-view-filter-actions">
            <button
              type="button"
              :class="viewMode === 'grid' ? 'on' : ''"
              @click="setView('grid')"
            >網格</button>
            <button
              type="button"
              :class="viewMode === 'table' ? 'on' : ''"
              @click="setView('table')"
            >表格</button>
          </div>
        </section>
      </div>

      <footer class="gp-filter-sheet-actions filter-sheet-actions">
        <button
          type="button"
          class="gp-filter-sheet-reset"
          @click="resetFilters"
        >重設</button>
        <button
          type="button"
          class="gp-filter-sheet-apply"
          @click="closeMobileFilter"
        >查看 {{ totalCount.toLocaleString('zh-HK') }} 個優惠</button>
      </footer>
    </aside>

    <!-- 1. 暗色 HERO -->
    <section class="gp-hero">
      <div class="gp-hero-left">
        <div class="gp-hero-label">超市格價</div>
        <div class="gp-hero-title">今日超市優惠</div>
        <div class="gp-hero-sub">搜尋全港主要超市價格，按優惠後價格、折扣力度與商店快速篩選。</div>
      </div>
      <div class="gp-hero-right">
        <template
          v-for="(stat, index) in heroStats"
          :key="stat.label"
        >
          <div class="gp-hstat">
            <span class="gp-hnum">{{ summaryLoading ? '-' : stat.value }}</span>
            <span class="gp-hlabel">{{ stat.label }}</span>
          </div>
          <div
            v-if="index < heroStats.length - 1"
            class="gp-hdiv"
          />
        </template>
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
            placeholder="搜尋商品、品牌或分類..."
          >
        </div>
        <button
          type="submit"
          class="gp-search-btn"
        >
          <span class="gp-search-button-label">搜尋</span>
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
          {{ favoritesLoading ? '載入中' : '我的收藏' }}
        </button>
      </form>
      <div
        class="gp-mobile-filter-rail"
        aria-label="優惠篩選"
      >
        <select
          v-model="activeSort"
          class="gp-mobile-filter-select gp-mobile-sort-select"
          aria-label="排序方式"
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
          aria-label="商品分類"
          @change="handleMobileQuickFilter('category', $event)"
        >
          <option value="">商品分類</option>
          <option
            v-for="category in categoryPills.slice(1)"
            :key="`mobile-category-${category.value}`"
            :value="category.value"
          >{{ category.label }}</option>
        </select>
        <select
          :value="activeStore"
          class="gp-mobile-filter-select"
          aria-label="商店"
          @change="handleMobileQuickFilter('store', $event)"
        >
          <option value="">商店</option>
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
          <span>更多</span>
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
          <span>收藏</span>
        </button>
        <button
          v-if="activeFilterCount > 0"
          type="button"
          class="gp-mobile-filter-reset"
          @click="resetFilters"
        >重設</button>
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
          {{ searchLoading ? '載入中' : `${totalCount.toLocaleString('zh-HK')} 個優惠` }}
        </span>
        <div class="gp-toolbar">
          <select
            v-model="activeBrand"
            class="gp-sort gp-brand-select"
            @change="selectBrand"
          >
            <option value="">全部品牌</option>
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
              網格
            </button>
            <button
              type="button"
              class="gp-view-btn"
              :class="viewMode === 'table' ? 'on' : ''"
              @click="setView('table')"
            >
              表格
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
        正在載入優惠商品。
      </p>
      <p
        v-else-if="!searchLoading && visibleProducts.length === 0"
        class="gp-state"
      >
        沒有找到符合條件的商品。
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
          <button
            type="button"
            class="gp-card-fav"
            :class="product.isFavorite ? 'on' : ''"
            @click.stop="toggleFavorite(product)"
          >
            {{ product.isFavorite ? '已收藏' : '收藏' }}
          </button>
          <div class="gp-card-img">
            <img
              v-if="product.image_url || product.imageUrl"
              :src="product.image_url || product.imageUrl"
              :alt="product.name"
            >
            <div
              v-else
              class="gp-card-img-ph"
            />
          </div>
          <div class="gp-card-body">
            <div class="gp-card-heading">
              <div>
                <div class="gp-card-name">{{ product.name }}</div>
                <div class="gp-card-brand">{{ product.brand || '未提供品牌' }}</div>
              </div>
              <div class="gp-card-badge">-{{ supermarketPriceDiscountRate(supermarketPrimaryPrice(product)).toFixed(0) }}%</div>
            </div>
            <div class="gp-card-prices">
              <div
                v-for="price in supermarketStorePrices(product).slice(0, 3)"
                :key="`${product.code}-${price.store}-${price.effectiveUnitPrice}`"
                class="gp-price-row"
              >
                <div>
                  <strong>{{ displaySupermarketStore(price.store) }}</strong>
                  <small>{{ price.offer || supermarketOfferTexts(product).join(' / ') || '-' }}</small>
                </div>
                <div>
                  <b>{{ formatSupermarketHKPrice(price.effectiveUnitPrice) }}</b>
                  <span>原價 {{ formatSupermarketHKPrice(price.listPrice) }}</span>
                </div>
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
              v-for="product in visibleProducts"
              :key="product.code"
              @click="openDetail(product)"
            >
              <td>
                <strong>{{ product.name }}</strong>
                <span>{{ product.brand || displaySupermarketCategory(product.category1 || '') }}</span>
              </td>
              <td>{{ displaySupermarketStore(supermarketPrimaryPrice(product).store) }}</td>
              <td>{{ supermarketOfferTexts(product).join(' / ') || '-' }}</td>
              <td>{{ formatSupermarketHKPrice(supermarketPrimaryPrice(product).effectiveUnitPrice) }}</td>
              <td>{{ formatSupermarketHKPrice(supermarketPrimaryPrice(product).listPrice) }}</td>
              <td>
                <button
                  type="button"
                  class="gp-table-fav"
                  @click.stop="toggleFavorite(product)"
                >
                  {{ product.isFavorite ? '已收藏' : '收藏' }}
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
  background: var(--sur-2);
  color: var(--ink);
}

.gp-filter-backdrop,
.gp-filter-sheet,
.gp-mobile-filter-rail,
.gp-mobile-search-icon {
  display: none;
}

/* 2. 暗色 HERO */
.gp-hero {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
  min-height: 164px;
  max-width: 1440px;
  margin: 0 auto;
  padding: 34px 38px;
  border-bottom: 3px solid var(--accent);
  background: #1a1a1a;
  color: #fff;
}

.gp-hero-left {
  min-width: 0;
}

.gp-hero-label {
  color: rgba(255, 255, 255, 0.58);
  font-size: 12px;
}

.gp-hero-title {
  margin-top: 8px;
  font-family: var(--font-serif);
  font-size: 34px;
  font-weight: 400;
  line-height: 1.12;
  color: #fff;
}

.gp-hero-sub {
  margin-top: 8px;
  max-width: 480px;
  color: rgba(255, 255, 255, 0.68);
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
  background: #fff;
}

.gp-search-row {
  display: flex;
  align-items: stretch;
  justify-content: center;
  gap: 8px;
  margin-bottom: 4px;
}

.gp-search-box {
  display: flex;
  max-width: 520px;
  align-items: center;
  gap: 8px;
  flex: 0 1 520px;
  border: 1px solid var(--bdr);
  border-radius: 3px;
  background: #fff;
  padding: 0 13px;
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
  background: #fff;
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
  background: #fff;
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
  padding: 26px 28px 44px;
  background: var(--sur-2);
}

.gp-content-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 20px;
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
  background: #fff;
}

.gp-view-toggle .gp-view-btn.on {
  background: #fff;
  color: var(--ink);
  border-color: var(--bdr);
}

.gp-sort {
  max-width: 180px;
  border: 1px solid var(--bdr);
  border-radius: 2px;
  background: #fff;
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
  background: #fff;
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
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 14px;
}

.gp-card {
  position: relative;
  min-height: 420px;
  margin: 0;
  border: 1px solid var(--bdr);
  border-radius: 3px;
  overflow: hidden;
  background: #fff;
  cursor: pointer;
  transition: border-color 0.15s ease, box-shadow 0.15s ease;
}

.gp-card:hover {
  border-color: var(--accent);
  box-shadow: none;
}

.gp-card-img {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 140px;
  padding: 14px;
  box-sizing: border-box;
  border-bottom: 1px solid var(--bdr);
  background: #fff;
  overflow: hidden;
}

.gp-card-img img {
  width: auto;
  height: auto;
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
  display: block;
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
  top: 14px;
  right: 14px;
  z-index: 2;
  border: 1px solid var(--bdr);
  border-radius: 2px;
  background: #fff;
  color: var(--ink-3);
  cursor: pointer;
  font-family: inherit;
  font-size: 11px;
  font-weight: 400;
  padding: 6px 10px;
}

.gp-card-fav.on {
  border-color: var(--accent);
  background: #fff;
  color: var(--accent);
}

.gp-card-body {
  padding: 14px;
}

.gp-card-heading {
  display: block;
  margin-bottom: 12px;
  padding-right: 0;
}

.gp-card-store {
  display: block;
  width: auto;
  margin: 0 0 6px;
  border-radius: 0;
  background: transparent;
  color: var(--ink-3);
  font-size: 12px;
  font-weight: 400;
  padding: 0;
}

.gp-card-name {
  margin: 0;
  color: var(--ink);
  font-size: 15px;
  font-weight: 700;
  line-height: 1.28;
}

.gp-card-brand {
  margin-top: 4px;
  color: var(--ink-3);
  font-size: 12px;
}

.gp-card-badge {
  position: absolute;
  top: 12px;
  left: 12px;
  right: auto;
  display: inline-flex;
  width: auto;
  max-width: max-content;
  align-items: center;
  border-radius: 2px;
  background: var(--accent);
  color: #fff;
  font-size: 11px;
  font-weight: 700;
  line-height: 1;
  padding: 5px 7px;
  white-space: nowrap;
}

.gp-card-prices {
  display: grid;
  gap: 0;
  margin-top: 10px;
}

.gp-price-row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 10px;
  border-top: 1px solid var(--sur-3);
  padding: 10px 0;
}

.gp-price-row:first-child {
  border-top: 0;
  padding-top: 0;
}

.gp-price-row strong {
  display: block;
  color: var(--ink);
  font-size: 13px;
  font-weight: 700;
}

.gp-price-row small {
  display: block;
  margin-top: 4px;
  color: var(--ink-3);
  font-size: 11px;
  line-height: 1.4;
}

.gp-price-row b {
  display: block;
  color: var(--ink);
  font-size: 14px;
  text-align: right;
  white-space: nowrap;
}

.gp-price-row span {
  display: block;
  margin-top: 3px;
  color: var(--accent);
  font-size: 11px;
  text-align: right;
  text-decoration: none;
  white-space: nowrap;
}

/* 6. 表格視圖 */
.gp-table {
  overflow: auto;
  border: 1px solid var(--bdr);
  border-radius: 3px;
  background: #fff;
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

.gp-table-fav {
  border: 1px solid var(--bdr);
  border-radius: 2px;
  background: #fff;
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
  background: #fff;
  color: var(--ink-3);
  cursor: pointer;
  font-family: inherit;
  font-size: 12px;
  min-width: 36px;
  padding: 8px 12px;
}

.gp-page-btn.on {
  border-color: var(--bdr);
  background: #fff;
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
    min-height: 0;
  }

  .gp-card-img {
    height: 126px;
    padding: 10px;
  }

  .gp-card-body {
    padding: 12px;
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
}
</style>

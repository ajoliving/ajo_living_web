<!--
 * 綜合優惠頻道頁。
 * 1. 使用 AJO 後端超市優惠摘要與搜尋接口。
 * 2. 提供商品搜尋、分類、品牌、商店、排序、收藏與分頁。
 * 3. 保持 web-new 現有暗色 Hero、控制欄、卡片與表格視圖風格。
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

// 5. 取得目前列表商品
const visibleProducts = computed<SupermarketProduct[]>(() => {
  if (showFavoritesOnly.value) {
    const start = (currentPage.value - 1) * pageSize;
    return favorites.value.slice(start, start + pageSize);
  }
  return searchResult.value?.items ?? [];
});

// 6. 取得目前總數
const totalCount = computed(() => (showFavoritesOnly.value ? favorites.value.length : searchResult.value?.total ?? 0));

// 7. 取得目前總頁數
const totalPages = computed(() => Math.max(1, Math.ceil(totalCount.value / pageSize)));

// 8. 建立分頁按鈕
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

// 9. 建立更新資訊文字
const updatedBarText = computed(() => {
  const updatedDate = formatSupermarketDate(summary.value?.metadata?.latestSnapshotDate);
  const productCount = formatInteger(summary.value?.stats.products);
  const prefix = updatedDate ? `資料更新：${updatedDate}` : '資料更新：等待資料';
  return `${prefix} · 共監測 ${productCount} 件商品 · 價格只供參考，實際售價以商戶公布為準。`;
});

// 10. 載入摘要
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

// 11. 載入搜尋結果
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

// 12. 載入收藏
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

// 13. 提交搜尋
const submitSearch = (): void => {
  showFavoritesOnly.value = false;
  currentPage.value = 1;
  void loadSearch();
};

// 14. 切換分類
const selectCategory = (category: string): void => {
  activeCategory.value = category;
  showFavoritesOnly.value = false;
  currentPage.value = 1;
  void loadSearch();
};

// 15. 切換商店
const selectStore = (store: string): void => {
  activeStore.value = store;
  showFavoritesOnly.value = false;
  currentPage.value = 1;
  void loadSearch();
};

// 16. 切換品牌
const selectBrand = (): void => {
  showFavoritesOnly.value = false;
  currentPage.value = 1;
  void loadSearch();
};

// 17. 切換排序
const selectSort = (): void => {
  currentPage.value = 1;
  if (!showFavoritesOnly.value) {
    void loadSearch();
  }
};

// 18. 切換收藏列表
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

// 19. 切換商品收藏
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

// 20. 切換視圖模式
const setView = (mode: ViewMode): void => {
  viewMode.value = mode;
};

// 21. 切換分頁
const selectPage = (page: PageButton): void => {
  if (page.disabled || currentPage.value === page.page) {
    return;
  }
  currentPage.value = page.page;
  if (!showFavoritesOnly.value) {
    void loadSearch();
  }
};

// 22. 開啟商品詳情
const openDetail = (product: SupermarketProduct): void => {
  void router.push({ path: `/supermarket-offers/products/${encodeURIComponent(product.code)}` });
};

// 23. 前往登入
const openLogin = async (redirect: string): Promise<void> => {
  await router.push({ path: '/login', query: { redirect } });
};

// 24. 更新商品收藏狀態
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

// 25. 建立篩選按鈕
const valueCountPills = (
  values: SupermarketValueCount[],
  formatter: (value: string) => string,
  limit: number,
): FilterPill[] =>
  values
    .slice(0, limit)
    .map((item) => ({ label: formatter(item.value), value: item.value }))
    .filter((item) => item.value);

// 26. 格式化整數
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
          搜尋
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
            class="gp-sort"
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
            class="gp-sort"
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
            <div class="gp-card-img-ph" />
          </div>
          <div class="gp-card-body">
            <div class="gp-card-heading">
              <div>
                <div class="gp-card-store">{{ displaySupermarketStore(supermarketPrimaryPrice(product).store) }}</div>
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
  min-height: calc(100vh - 48px);
  background: var(--sur-2);
  color: var(--ink);
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
  background: #fff;
  color: var(--ink-3);
  cursor: pointer;
  font-family: inherit;
  font-size: 13px;
  padding: 7px 15px;
}

.gp-fpill {
  border-radius: 999px;
}

.gp-spill {
  border-radius: 3px;
}

.gp-fpill.on {
  background: var(--accent);
  color: var(--white);
  border-color: var(--accent);
}

.gp-spill.on {
  background: var(--accent-light);
  color: var(--accent-dark);
  border-color: var(--accent);
  font-weight: 500;
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
  border-bottom: 1px solid var(--bdr);
  background: #f6f6f6;
  overflow: hidden;
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

@media (max-width: 900px) {
  .gp-hero,
  .gp-controls,
  .gp-content {
    max-width: none;
  }

  .gp-hero {
    margin: 14px 14px 0;
    align-items: flex-start;
    flex-direction: column;
    padding: 22px;
  }

  .gp-hero-right,
  .gp-content-header,
  .gp-toolbar {
    width: 100%;
  }

  .gp-hero-right {
    justify-content: space-between;
  }

  .gp-hstat {
    min-width: 0;
    flex: 1;
    padding: 0 12px;
  }

  .gp-controls,
  .gp-content {
    margin: 12px 14px 0;
    padding: 14px;
  }

  .gp-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .gp-search-row,
  .gp-content-header {
    align-items: stretch;
    flex-direction: column;
  }

  .gp-toolbar {
    align-items: stretch;
    flex-wrap: wrap;
  }

  .gp-sort {
    max-width: none;
  }
}

@media (max-width: 560px) {
  .gp-grid {
    grid-template-columns: 1fr;
  }
}
</style>

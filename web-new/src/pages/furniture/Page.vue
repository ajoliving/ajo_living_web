<!--
 * 家具市集列表頁。
 * 1. 桌面採左側篩選、中間列表與右側廣告三欄布局。
 * 2. 手機將篩選收進底部彈窗，入口固定在搜尋框左側。
 * 3. 中間含排序欄、商品卡片網格（懸停顯示操作按鈕）與分頁。
 * 4. 右側接入 3 個 16:9 短廣告與 2 個 9:16 長廣告。
 * 5. 接入真實二手帖子 API，保留舊版家具篩選能力與新版 UI 風格。
 * 6. 樣式對齊 HTML 設計稿 ajo_living_desktop_20260624(3)(10).html 的 #page-market 規則。
-->
<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useRouter } from 'vue-router';

import type { MarketplaceCategoryCode, MarketplaceConditionCode } from '@/constants/marketplace';
import AppIcon from '@/shared/components/base/AppIcon.vue';
import FilterTag from '@/shared/components/base/FilterTag.vue';
import ListingSideAds from '@/shared/components/ads/ListingSideAds.vue';

import { useFurniturePage } from './composables/useFurniturePage';

interface FurniturePriceRangeOption {
  value: string;
  label: string;
  min: number;
  max: number;
}

// 1. 路由
const router = useRouter();

const {
  area,
  areaOptions,
  categories,
  clearConditions,
  clearFilters,
  conditionOptions,
  formatListingTime,
  isFavoriteUpdating,
  keyword,
  listings,
  loading,
  maxPrice,
  maxPriceLimit,
  minPrice,
  page,
  selectedCategoryKeys,
  selectedConditions,
  setPage,
  sortBy,
  sortOptions,
  syncFiltersToQuery,
  t,
  toggleCategory,
  toggleCondition,
  toggleFavorite,
  totalPages,
  totalResults,
  visibility,
  visibilityOptions,
  withPhotos,
} = useFurniturePage();

// 2. 價格快捷選項
const selectedPriceRange = ref('');
const isFilterOpen = ref(false);
const furniturePriceRangeOptions = computed<FurniturePriceRangeOption[]>(() => [
  { value: 'under_1000', label: t('channels.furniture.priceUnder', { price: '$1k' }), min: 0, max: 1000 },
  { value: '1000_3000', label: t('channels.furniture.priceBetween', { min: '$1k', max: '$3k' }), min: 1000, max: 3000 },
  { value: '3000_5000', label: t('channels.furniture.priceBetween', { min: '$3k', max: '$5k' }), min: 3000, max: 5000 },
  { value: 'over_10000', label: t('channels.furniture.priceOver', { price: '$10k' }), min: 10000, max: maxPriceLimit },
]);

const selectedPriceRangeFromValues = computed(() => {
  const minValue = Number(minPrice.value || 0);
  const maxValue = Number(maxPrice.value || maxPriceLimit);
  const matched = furniturePriceRangeOptions.value.find((option) =>
    option.min === minValue && option.max === maxValue,
  );

  return matched?.value ?? '';
});

const isAllPriceSelected = computed(() => {
  const minValue = Number(minPrice.value || 0);
  const maxValue = Number(maxPrice.value || maxPriceLimit);

  return minValue <= 0 && maxValue >= maxPriceLimit;
});

const activeFilterCount = computed(() =>
  selectedCategoryKeys.value.length
  + selectedConditions.value.length
  + (isAllPriceSelected.value ? 0 : 1)
  + (area.value === 'all' ? 0 : 1)
  + (visibility.value === 'all' ? 0 : 1)
  + (withPhotos.value ? 1 : 0),
);
const mobileAreaOptions = computed(() =>
  areaOptions.value.filter((option) => option.value !== 'all'),
);

// 3. 分頁按鈕資料
const paginationButtons = computed(() => {
  const buttons: Array<{ label: string | number; key: string | number; pageValue: number; active?: boolean; disabled?: boolean }> = [
    {
      label: t('marketplace.filter.previousPage'),
      key: 'prev',
      pageValue: page.value - 1,
      disabled: page.value <= 1,
    },
  ];
  const start = Math.max(1, page.value - 1);
  const end = Math.min(totalPages.value, start + 2);

  for (let index = start; index <= end; index += 1) {
    buttons.push({
      label: index,
      key: index,
      pageValue: index,
      active: index === page.value,
    });
  }

  buttons.push({
    label: t('marketplace.filter.nextPage'),
    key: 'next',
    pageValue: page.value + 1,
    disabled: page.value >= totalPages.value,
  });

  return buttons;
});

// 4. 套用價格區間
const applyPriceRange = (option?: FurniturePriceRangeOption): void => {
  selectedPriceRange.value = option?.value ?? '';
  minPrice.value = String(option?.min ?? 0);
  maxPrice.value = String(option?.max ?? maxPriceLimit);
  void syncFiltersToQuery();
};

// 5. 切換篩選標籤
const handleFilterToggle = (groupKey: string, optionValue: string): void => {
  if (groupKey === 'category') {
    if (optionValue === 'all') {
      selectedCategoryKeys.value = [];
      void syncFiltersToQuery();
      return;
    }
    void toggleCategory(optionValue as MarketplaceCategoryCode);
    return;
  }

  if (groupKey === 'price') {
    const target = furniturePriceRangeOptions.value.find((option) => option.value === optionValue);
    applyPriceRange(target);
    return;
  }

  if (groupKey === 'condition') {
    if (optionValue === 'all') {
      void clearConditions();
      return;
    }
    void toggleCondition(optionValue as MarketplaceConditionCode);
    return;
  }

  if (groupKey === 'area') {
    area.value = optionValue;
    void syncFiltersToQuery();
    return;
  }

  if (groupKey === 'visibility') {
    visibility.value = optionValue as typeof visibility.value;
    void syncFiltersToQuery();
  }
};

// 6. 搜尋與清除條件
const handleSearch = (): void => {
  void syncFiltersToQuery(1);
};

const handleClearFilters = (): void => {
  selectedPriceRange.value = '';
  void clearFilters();
};

// 7. 套用手機端快捷篩選
const handleMobileFilterChange = (groupKey: 'category' | 'price' | 'condition' | 'area', event: Event): void => {
  const value = (event.target as HTMLSelectElement).value;

  if (groupKey === 'category') {
    selectedCategoryKeys.value = value ? [value as MarketplaceCategoryCode] : [];
    void syncFiltersToQuery();
    return;
  }

  if (groupKey === 'price') {
    applyPriceRange(furniturePriceRangeOptions.value.find((option) => option.value === value));
    return;
  }

  if (groupKey === 'condition') {
    selectedConditions.value = value ? [value as MarketplaceConditionCode] : [];
    void syncFiltersToQuery();
    return;
  }

  area.value = value || 'all';
  void syncFiltersToQuery();
};

// 8. 開關手機端篩選彈窗
const openFilterSheet = (): void => {
  isFilterOpen.value = true;
};

const closeFilterSheet = (): void => {
  isFilterOpen.value = false;
};

// 9. 點擊卡片跳轉詳情
const handleCardClick = (id: string): void => {
  void router.push(`/furniture/${id}`);
};

// 10. 隱藏失效商品圖片
const hideFailedImage = (event: Event): void => {
  (event.currentTarget as HTMLImageElement).style.display = 'none';
};

// 11. 點擊分頁
const handlePageClick = (targetPage: number): void => {
  void setPage(targetPage);
};

// 12. 從 URL 價格條件反推快捷區間狀態
watch(
  selectedPriceRangeFromValues,
  (value) => {
    selectedPriceRange.value = value;
  },
  { immediate: true },
);
</script>

<template>
  <div
    id="page-market"
    class="page"
  >
    <div class="mp">
      <button
        v-if="isFilterOpen"
        type="button"
        class="filter-backdrop"
        :aria-label="t('common.action.close')"
        @click="closeFilterSheet"
      ></button>

      <!-- 左側篩選欄 -->
      <aside
        id="furniture-filter-sheet"
        class="mf"
        :class="{ 'is-filter-open': isFilterOpen }"
        :role="isFilterOpen ? 'dialog' : undefined"
        :aria-modal="isFilterOpen ? 'true' : undefined"
        aria-labelledby="furniture-filter-title"
      >
        <header class="filter-sheet-header">
          <div>
            <span class="filter-sheet-eyebrow">{{ t('channels.furniture.title') }}</span>
            <h2 id="furniture-filter-title">{{ t('marketplace.filter.title') }}</h2>
          </div>
          <button
            type="button"
            class="filter-sheet-close"
            :aria-label="t('common.action.close')"
            @click="closeFilterSheet"
          >
            <AppIcon
              name="close"
              :size="20"
            />
          </button>
        </header>

        <div class="sbar desktop-filter-search">
          <input
            v-model="keyword"
            class="sinput"
            :placeholder="t('channels.furniture.searchPlaceholder')"
            autocomplete="off"
            @keyup.enter="handleSearch"
          />
          <button
            type="button"
            class="sbtn"
            @click="handleSearch"
          >{{ t('marketplace.list.searchAction') }}</button>
        </div>

        <div class="filter-sections">
          <section class="fs">
            <div class="ft-title">{{ t('marketplace.filter.categories') }}</div>
            <div class="ftags">
              <FilterTag
                :label="t('channels.furniture.allFurniture')"
                :active="selectedCategoryKeys.length === 0"
                @toggle="handleFilterToggle('category', 'all')"
              />
              <FilterTag
                v-for="category in categories"
                :key="category.key"
                :label="category.label"
                :active="selectedCategoryKeys.includes(category.key)"
                @toggle="handleFilterToggle('category', category.key)"
              />
            </div>
          </section>

          <section class="fs">
            <div class="ft-title">{{ t('marketplace.filter.priceRange') }}</div>
            <div class="ftags">
              <FilterTag
                :label="t('marketplace.list.priceAll')"
                :active="isAllPriceSelected"
                @toggle="applyPriceRange()"
              />
              <FilterTag
                v-for="option in furniturePriceRangeOptions"
                :key="option.value"
                :label="option.label"
                :active="selectedPriceRange === option.value"
                @toggle="handleFilterToggle('price', option.value)"
              />
            </div>
          </section>

          <section class="fs">
            <div class="ft-title">{{ t('marketplace.filter.condition') }}</div>
            <div class="ftags">
              <FilterTag
                :label="t('marketplace.filter.conditionAll')"
                :active="selectedConditions.length === 0"
                @toggle="handleFilterToggle('condition', 'all')"
              />
              <FilterTag
                v-for="condition in conditionOptions"
                :key="condition.value"
                :label="condition.label"
                :active="selectedConditions.includes(condition.value)"
                @toggle="handleFilterToggle('condition', condition.value)"
              />
            </div>
          </section>

          <section class="fs">
            <div class="ft-title">{{ t('marketplace.filter.area') }}</div>
            <div class="ftags">
              <FilterTag
                v-for="option in areaOptions"
                :key="option.value"
                :label="option.label"
                :active="area === option.value"
                @toggle="handleFilterToggle('area', option.value)"
              />
            </div>
          </section>

          <section class="fs">
            <div class="ft-title">{{ t('common.label.visibility') }}</div>
            <div class="ftags">
              <FilterTag
                v-for="option in visibilityOptions"
                :key="option.value"
                :label="option.label"
                :active="visibility === option.value"
                @toggle="handleFilterToggle('visibility', option.value)"
              />
            </div>
          </section>

          <label class="photo-toggle">
            <input
              v-model="withPhotos"
              type="checkbox"
              @change="syncFiltersToQuery()"
            />
            <span>{{ t('marketplace.filter.onlyPhotos') }}</span>
          </label>
        </div>

        <button
          type="button"
          class="clear-filter-btn desktop-filter-clear"
          @click="handleClearFilters"
        >
          {{ t('marketplace.filter.clearAll') }}
        </button>

        <div class="filter-sheet-actions">
          <button
            type="button"
            class="filter-sheet-reset"
            @click="handleClearFilters"
          >{{ t('marketplace.filter.clearAll') }}</button>
          <button
            type="button"
            class="filter-sheet-apply"
            @click="closeFilterSheet"
          >{{ t('marketplace.filter.viewResults', { count: totalResults }) }}</button>
        </div>
      </aside>

      <!-- 中間列表區 -->
      <main class="mr">
        <div class="mobile-listing-controls">
          <div class="mobile-search-toolbar">
            <form
              class="mobile-search-form"
              @submit.prevent="handleSearch"
            >
              <input
                v-model="keyword"
                class="sinput"
                :placeholder="t('channels.furniture.searchPlaceholder')"
                autocomplete="off"
              />
              <button
                type="submit"
                class="mobile-search-button"
                :aria-label="t('marketplace.list.searchAction')"
              >
                <AppIcon
                  name="search"
                  :size="19"
                />
              </button>
            </form>
          </div>

          <div
            class="mobile-filter-rail"
            :aria-label="t('marketplace.filter.title')"
          >
            <select
              :value="selectedCategoryKeys[0] ?? ''"
              class="mobile-filter-select"
              :aria-label="t('marketplace.filter.categories')"
              @change="handleMobileFilterChange('category', $event)"
            >
              <option value="">{{ t('marketplace.filter.categories') }}</option>
              <option
                v-for="category in categories"
                :key="category.key"
                :value="category.key"
              >{{ category.label }}</option>
            </select>
            <select
              :value="area === 'all' ? '' : area"
              class="mobile-filter-select"
              :aria-label="t('marketplace.filter.area')"
              @change="handleMobileFilterChange('area', $event)"
            >
              <option value="">{{ t('marketplace.filter.area') }}</option>
              <option
                v-for="option in mobileAreaOptions"
                :key="option.value"
                :value="option.value"
              >{{ option.label }}</option>
            </select>
            <select
              :value="selectedConditions[0] ?? ''"
              class="mobile-filter-select"
              :aria-label="t('marketplace.filter.condition')"
              @change="handleMobileFilterChange('condition', $event)"
            >
              <option value="">{{ t('marketplace.filter.condition') }}</option>
              <option
                v-for="condition in conditionOptions"
                :key="condition.value"
                :value="condition.value"
              >{{ condition.label }}</option>
            </select>
            <select
              v-model="sortBy"
              class="mobile-filter-select mobile-sort-select"
              :aria-label="t('marketplace.filter.sortBy')"
              @change="syncFiltersToQuery(page)"
            >
              <option
                v-for="option in sortOptions"
                :key="option.value"
                :value="option.value"
              >{{ option.label }}</option>
            </select>
            <button
              type="button"
              class="mobile-filter-button"
              aria-controls="furniture-filter-sheet"
              :aria-expanded="isFilterOpen"
              @click="openFilterSheet"
            >
              <span>{{ t('marketplace.filter.more') }}</span>
              <AppIcon
                name="chevron-down"
                :size="16"
              />
              <span
                v-if="activeFilterCount > 0"
                class="mobile-filter-count"
              >{{ activeFilterCount }}</span>
            </button>
            <button
              v-if="activeFilterCount > 0"
              type="button"
              class="mobile-filter-reset"
              @click="handleClearFilters"
            >{{ t('marketplace.filter.reset') }}</button>
          </div>
        </div>

        <div class="sort-row">
          <span class="rn">
            {{ loading ? t('common.status.loading') : t('channels.furniture.results', { count: totalResults }) }}
          </span>
          <select
            v-model="sortBy"
            class="ssel"
            @change="syncFiltersToQuery(page)"
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

        <div
          v-if="loading"
          class="market-state"
        >
          {{ t('common.status.loading') }}
        </div>

        <div
          v-else-if="listings.length === 0"
          class="market-state"
        >
          {{ t('channels.furniture.empty') }}
        </div>

        <div
          v-else
          class="mgrid"
        >
          <div
            v-for="listing in listings"
            :key="listing.id"
            class="mc"
            @click="handleCardClick(listing.id)"
          >
            <div
              class="mimg"
              :class="listing.imageUrl ? '' : 'pat'"
            >
              <img
                v-if="listing.imageUrl"
                :src="listing.imageUrl"
                :alt="listing.title"
                @error="hideFailedImage"
              />
              <span v-else>{{ t('marketplace.filter.noImage') }}</span>
            </div>
            <div class="mbody">
              <div class="mname">{{ listing.title }}</div>
              <div class="msub">
                {{ listing.districtLabel }} · {{ listing.conditionLabel }} ·
                {{ listing.visibilityScope === 'building_only' ? t('common.state.buildingOnly') : t('common.state.public') }}
              </div>
              <div class="mprice">{{ listing.price }}</div>
              <div class="mfoot">
                <span>{{ listing.categoryLabel }} · {{ formatListingTime(listing.publishedAt, listing.updatedAt) }}</span>
              </div>
            </div>
            <div class="mc-actions">
              <button
                type="button"
                class="mc-action-btn"
                :disabled="isFavoriteUpdating(listing.id)"
                @click.stop="toggleFavorite(listing.id)"
              >{{ listing.isFavorited ? t('marketplace.detail.favoritedAction') : t('marketplace.detail.favoriteAction') }}</button>
              <button
                type="button"
                class="mc-action-btn primary"
                @click.stop="handleCardClick(listing.id)"
              >{{ t('common.action.viewDetail') }}</button>
            </div>
          </div>
        </div>

        <nav
          v-if="!loading && totalResults > 0"
          class="market-pagination"
          :aria-label="t('channels.furniture.paginationAria')"
        >
          <span class="market-pagination-info">
            {{ t('marketplace.list.pageLabel') }} {{ page }} / {{ totalPages }}
          </span>
          <div class="market-pagination-actions">
            <button
              v-for="button in paginationButtons"
              :key="button.key"
              type="button"
              class="market-page-btn"
              :class="{ on: button.active, disabled: button.disabled }"
              :disabled="button.disabled"
              @click="handlePageClick(button.pageValue)"
            >
              {{ button.label }}
            </button>
          </div>
        </nav>
      </main>

      <!-- 右側廣告欄 -->
      <aside
        class="market-ad-aside"
        :aria-label="t('channels.furniture.advertisingAria')"
      >
        <ListingSideAds channel="furniture" />
      </aside>
    </div>
  </div>
</template>

<style scoped>
/* 1. 頁面容器 */
.page {
  width: 100%;
  min-height: calc(100svh - var(--nav-h, 52px));
  background: var(--sur-2);
}

/* 2. 三欄布局：對齊全局頁面寬度與左右留白 */
.mp {
  display: grid;
  grid-template-columns: 250px minmax(0, 1fr) 360px;
  justify-content: center;
  align-items: stretch;
  width: min(100%, var(--layout-page-max-width));
  margin: 0 auto;
  padding: 0 var(--layout-page-padding-inline);
  background: var(--sur-2);
}

/* 3. 左側篩選欄：對齊 .mf 與 #page-market .mf */
.mf {
  border-right: 1px solid var(--bdr);
  padding: 18px 18px;
  position: sticky;
  top: var(--nav-h, 52px);
  align-self: start;
  height: auto;
  min-height: calc(100svh - var(--nav-h, 52px));
  overflow: visible;
  scrollbar-width: none;
  background: var(--sur);
}

.mf::-webkit-scrollbar {
  display: none;
}

.filter-backdrop,
.filter-sheet-header,
.filter-sheet-actions,
.mobile-listing-controls {
  display: none;
}

.sbar {
  display: flex;
  gap: 6px;
  min-width: 0;
  margin-left: -14px;
  margin-right: -14px;
  margin-top: -18px;
  margin-bottom: 14px;
  padding: 18px 0 14px;
  background: var(--sur);
}

.sinput {
  flex: 1;
  min-width: 0;
  width: 100%;
  border: 1px solid var(--bdr);
  padding: 8px 10px;
  font-size: 12px;
  font-family: inherit;
  outline: none;
  border-radius: var(--r-md);
  background: var(--sur);
  color: var(--ink);
  transition: border-color 0.15s;
}

.sinput:focus {
  border-color: var(--brand);
  outline: none;
  box-shadow: 0 0 0 3px rgba(240, 90, 0, 0.1);
}

.sbtn {
  background: var(--brand);
  color: var(--sur);
  border: none;
  padding: 8px 12px;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  border-radius: var(--r-md);
  font-family: inherit;
}

.fs {
  margin-bottom: 16px;
}

.ft-title {
  font-size: 9px;
  letter-spacing: 2px;
  color: var(--ink-4);
  text-transform: uppercase;
  margin-bottom: 6px;
}

.ftags {
  display: flex;
  flex-wrap: wrap;
  gap: 5px;
  max-width: 100%;
  overflow: hidden;
}

.photo-toggle {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 4px 0 12px;
  color: var(--ink-3);
  cursor: pointer;
  font-size: 11px;
  line-height: 1.5;
}

.photo-toggle input {
  width: 14px;
  height: 14px;
  accent-color: var(--brand);
}

.clear-filter-btn {
  width: 100%;
  border: 1px solid var(--bdr);
  border-radius: 6px;
  background: var(--sur);
  color: var(--ink-2);
  cursor: pointer;
  font-family: inherit;
  font-size: 12px;
  padding: 8px 10px;
}

.clear-filter-btn:hover {
  border-color: var(--brand-mid);
  color: var(--brand);
}

/* 4. 中間列表區：對齊 .mr 與 #page-market .mr */
.mr {
  padding: 14px 14px 36px;
  min-width: 0;
  min-height: calc(100svh - var(--nav-h, 52px));
  background: var(--sur-2);
}

.sort-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
  width: 100%;
  margin-left: 0;
  margin-right: auto;
}

.rn {
  font-size: 11px;
  color: var(--ink-4);
}

.ssel {
  border: 1px solid var(--bdr);
  padding: 5px 8px;
  font-size: 11px;
  font-family: inherit;
  outline: none;
  border-radius: 2px;
  background: var(--sur);
  color: var(--ink);
}

.market-state {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 220px;
  width: 100%;
  border: 1px solid var(--bdr);
  border-radius: 8px;
  background: var(--sur);
  color: var(--ink-3);
  font-size: 13px;
}

/* 5. 商品卡片網格：對齊 #page-market .mgrid、.mc、.mimg */
.mgrid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
  width: 100%;
  margin-left: 0;
  margin-right: auto;
}

.mc {
  position: relative;
  margin-bottom: 0;
  border: 1px solid var(--bdr);
  border-radius: 8px;
  overflow: hidden;
  cursor: pointer;
  background: var(--sur);
  break-inside: avoid;
  transition: box-shadow 0.15s ease, border-color 0.15s ease;
}

.mc:hover {
  border-color: var(--brand-mid);
  box-shadow: var(--shadow-md);
}

.mnew {
  position: absolute;
  top: 7px;
  left: 7px;
  background: var(--ink);
  color: var(--sur);
  font-size: 9px;
  font-weight: 600;
  padding: 1px 5px;
  letter-spacing: 0.3px;
  border-radius: var(--r-sm);
}

.mimg {
  height: 160px;
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--g1);
}

.mimg img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.mimg span {
  position: relative;
  z-index: 1;
  color: var(--ink-4);
  font-size: 11px;
}

/* 6. 斜紋圖案：對齊 .pat */
.pat {
  position: relative;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
}

.pat::after {
  content: '';
  position: absolute;
  inset: 0;
  background: repeating-linear-gradient(
    45deg,
    transparent,
    transparent 5px,
    rgba(0, 0, 0, 0.025) 5px,
    rgba(0, 0, 0, 0.025) 10px
  );
  pointer-events: none;
}

.mbody {
  padding: 9px 11px;
  padding-bottom: 9px;
}

.mname {
  font-size: 12px;
  font-weight: 500;
  margin-bottom: 2px;
  color: var(--ink);
}

.msub {
  font-size: 10px;
  color: var(--ink-4);
  margin-bottom: 5px;
}

.mprice {
  font-size: 14px;
  font-weight: 300;
  color: var(--ink);
}

.mfoot {
  display: none !important;
}

/* 7. 懸停操作按鈕：對齊 .mc-actions、.mc-action-btn */
.mc-actions {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px;
  height: auto;
  max-height: 0;
  overflow: hidden;
  opacity: 0;
  pointer-events: none;
  transform: translateY(-2px);
  padding: 0 12px;
  transition: max-height 0.18s ease, opacity 0.15s ease, transform 0.15s ease, padding 0.18s ease;
}

.mc:hover .mc-actions {
  max-height: 54px;
  opacity: 1;
  pointer-events: auto;
  transform: translateY(0);
  padding: 0 12px 12px;
}

.mc-action-btn {
  min-height: 34px;
  border: 1px solid var(--bdr);
  border-radius: 6px;
  background: var(--sur);
  color: var(--ink-2);
  cursor: pointer;
  font-family: var(--font);
  font-size: 12px;
  font-weight: 500;
}

.mc-action-btn:hover {
  border-color: var(--brand-mid);
  color: var(--brand);
}

.mc-action-btn.primary {
  border-color: var(--brand);
  background: var(--brand);
  color: var(--sur);
}

.mc-action-btn.primary:hover {
  background: var(--brand-dark);
  color: var(--sur);
}

.mc-action-btn:disabled,
.market-page-btn:disabled {
  cursor: not-allowed;
  opacity: 0.55;
}

/* 8. 分頁：對齊 .market-pagination、.market-page-btn */
.market-pagination {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-top: 18px;
  padding: 12px 14px;
  border: 1px solid var(--bdr);
  border-radius: 8px;
  background: var(--sur);
  width: 100%;
  margin-left: 0;
  margin-right: auto;
}

.market-pagination-info {
  font-size: 11px;
  color: var(--ink-3);
}

.market-pagination-actions {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}

.market-page-btn {
  border: 1px solid var(--bdr);
  border-radius: 6px;
  background: var(--sur);
  color: var(--ink-3);
  cursor: pointer;
  font-family: inherit;
  font-size: 11px;
  padding: 7px 11px;
}

.market-page-btn:hover {
  border-color: var(--brand-mid);
  color: var(--brand);
}

.market-page-btn.on {
  border-color: var(--ink);
  color: var(--ink);
  font-weight: 500;
}

.market-page-btn.disabled:hover {
  border-color: var(--bdr);
  color: var(--ink-3);
}

/* 9. 右側廣告欄 */
.market-ad-aside {
  position: sticky;
  top: var(--nav-h, 52px);
  align-self: start;
  overflow: visible;
  border-left: 1px solid var(--bdr);
  background: var(--sur-2);
  padding: 18px 18px 40px;
}

/* 10. 響應式 - 大螢幕：對齊 @media (min-width:1440px) */
@media (min-width: 1440px) {
  .mp {
    grid-template-columns: 260px minmax(0, 1fr) 380px;
  }
}

/* 11. 響應式 - 平板與行動裝置 */
@media (max-width: 1023px) {
  .mp {
    grid-template-columns: 1fr;
  }

  .filter-backdrop {
    position: fixed;
    z-index: 120;
    inset: 0;
    display: block;
    width: 100%;
    height: 100%;
    border: 0;
    background: rgb(0 0 0 / 0.44);
    cursor: pointer;
    touch-action: none;
  }

  .mf {
    position: fixed;
    z-index: 121;
    top: auto;
    right: 0;
    bottom: 0;
    left: 0;
    display: flex;
    width: 100%;
    max-height: min(82svh, 720px);
    min-height: 0;
    flex-direction: column;
    border: 0;
    border-radius: 8px 8px 0 0;
    padding: 0;
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

  .mf.is-filter-open {
    opacity: 1;
    pointer-events: auto;
    transform: translateY(0);
    visibility: visible;
  }

  .desktop-filter-search,
  .desktop-filter-clear {
    display: none;
  }

  .filter-sheet-header {
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

  .filter-sheet-header h2 {
    margin: 1px 0 0;
    color: var(--ink);
    font-size: 16px;
    font-weight: 700;
    letter-spacing: 0;
  }

  .filter-sheet-eyebrow {
    display: block;
    color: var(--ink-3);
    font-size: 10px;
    font-weight: 600;
    letter-spacing: 0;
  }

  .filter-sheet-close {
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

  .filter-sections {
    padding: 10px 14px 2px;
  }

  .filter-sections .fs {
    margin-bottom: 10px;
  }

  .filter-sections .ft-title {
    margin-bottom: 4px;
    color: var(--ink-3);
    font-size: 10px;
    font-weight: 700;
    letter-spacing: 0;
  }

  .filter-sections .ftags {
    gap: 4px;
    overflow: visible;
  }

  .filter-sections :deep(.ft) {
    min-height: 36px;
    padding: 4px 10px;
  }

  .filter-sections .photo-toggle {
    min-height: 36px;
    margin: 0 0 8px;
  }

  .filter-sheet-actions {
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

  .filter-sheet-reset,
  .filter-sheet-apply {
    min-height: 44px;
    border-radius: 6px;
    cursor: pointer;
    font: inherit;
    font-size: 13px;
    font-weight: 700;
  }

  .filter-sheet-reset {
    border: 1px solid var(--bdr);
    background: var(--sur);
    color: var(--ink-2);
  }

  .filter-sheet-apply {
    border: 1px solid var(--brand);
    background: var(--brand);
    color: #fff;
  }

  .mobile-listing-controls {
    display: block;
    box-sizing: border-box;
    margin: -12px calc(-1 * var(--layout-page-padding-inline)) 12px;
    border-bottom: 1px solid var(--bdr);
    padding: 12px var(--layout-page-padding-inline) 10px;
    background: var(--sur);
  }

  .mobile-search-toolbar {
    margin-bottom: 10px;
  }

  .mobile-search-form {
    display: grid;
    grid-template-columns: minmax(0, 1fr) 48px;
    min-width: 0;
  }

  .mobile-search-form .sinput {
    height: 48px;
    border-radius: 8px 0 0 8px;
    background: var(--sur);
    font-size: 16px;
  }

  .mobile-search-button {
    display: inline-flex;
    width: 48px;
    height: 48px;
    align-items: center;
    justify-content: center;
    border: 1px solid var(--brand);
    border-radius: 0 8px 8px 0;
    background: var(--brand);
    color: #fff;
    cursor: pointer;
  }

  .mobile-filter-rail {
    display: flex;
    gap: 8px;
    margin: 0 -2px;
    overflow-x: auto;
    overscroll-behavior-x: contain;
    padding: 2px;
    scrollbar-width: none;
  }

  .mobile-filter-rail::-webkit-scrollbar {
    display: none;
  }

  .mobile-filter-select,
  .mobile-filter-button,
  .mobile-filter-reset {
    min-height: 40px;
    border: 1px solid var(--bdr);
    border-radius: 999px;
    background: var(--sur);
    color: var(--ink-2);
    font: inherit;
    font-size: 13px;
  }

  .mobile-filter-select {
    width: auto;
    min-width: 98px;
    flex: 0 0 auto;
    padding: 0 30px 0 13px;
  }

  .mobile-sort-select {
    min-width: 108px;
  }

  .mobile-filter-button {
    display: inline-flex;
    flex: 0 0 auto;
    align-items: center;
    gap: 4px;
    padding: 0 12px;
    cursor: pointer;
  }

  .mobile-filter-reset {
    flex: 0 0 auto;
    border-color: transparent;
    padding: 0 4px;
    background: transparent;
    color: var(--ink-3);
    cursor: pointer;
  }

  .mobile-filter-count {
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

  .sort-row {
    align-items: stretch;
    flex-direction: column;
    gap: 6px;
    margin: 0 0 10px;
  }

  .sort-row .ssel {
    width: 100%;
    min-height: 42px;
    border: 1px solid var(--bdr);
    border-radius: 8px;
    padding: 0 12px;
    background: var(--sur);
    color: var(--ink);
    font: inherit;
    font-size: 14px;
  }

  .sort-row > .ssel {
    display: none;
  }

  .mr {
    padding: 0 0 36px;
  }

  .mgrid,
  .market-pagination,
  .sort-row {
    max-width: none;
  }

  .mgrid {
    grid-template-columns: 1fr;
  }

  .market-ad-aside {
    position: static;
    height: auto;
    border-left: 0;
    border-top: 1px solid var(--bdr);
    padding: 14px;
  }

  .market-pagination {
    flex-direction: column;
    align-items: flex-start;
  }

}

</style>

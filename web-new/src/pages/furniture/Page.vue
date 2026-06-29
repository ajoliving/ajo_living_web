<!--
 * 家具市集列表頁。
 * 1. 三欄布局：左側篩選欄 + 中間列表區 + 右側廣告欄。
 * 2. 左側含搜尋框與 5 組篩選標籤（分類、價格範圍、成色、地區、可見範圍）。
 * 3. 中間含排序欄、商品卡片網格（懸停顯示操作按鈕）與分頁。
 * 4. 右側接入 3 個 16:9 短廣告與 2 個 9:16 長廣告。
 * 5. 接入真實二手帖子 API，保留舊版家具篩選能力與新版 UI 風格。
 * 6. 樣式對齊 HTML 設計稿 ajo_living_desktop_20260624(3)(10).html 的 #page-market 規則。
-->
<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useRouter } from 'vue-router';

import type { MarketplaceCategoryCode, MarketplaceConditionCode } from '@/constants/marketplace';
import FilterTag from '@/shared/components/base/FilterTag.vue';
import ListingSideAds from '@/shared/components/ads/ListingSideAds.vue';

import { useFurniturePage } from './composables/useFurniturePage';

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
const furniturePriceRangeOptions = [
  { value: 'under_1000', label: '$1,000以下', min: 0, max: 1000 },
  { value: '1000_3000', label: '$1,000-$3,000', min: 1000, max: 3000 },
  { value: '3000_5000', label: '$3,000-$5,000', min: 3000, max: 5000 },
  { value: '5000_10000', label: '$5,000-$10,000', min: 5000, max: 10000 },
  { value: 'over_10000', label: '$10,000以上', min: 10000, max: maxPriceLimit },
];

const selectedPriceRangeFromValues = computed(() => {
  const minValue = Number(minPrice.value || 0);
  const maxValue = Number(maxPrice.value || maxPriceLimit);
  const matched = furniturePriceRangeOptions.find((option) =>
    option.min === minValue && option.max === maxValue,
  );

  return matched?.value ?? '';
});

const isAllPriceSelected = computed(() => {
  const minValue = Number(minPrice.value || 0);
  const maxValue = Number(maxPrice.value || maxPriceLimit);

  return minValue <= 0 && maxValue >= maxPriceLimit;
});

// 3. 分頁按鈕資料
const paginationButtons = computed(() => {
  const buttons: Array<{ label: string | number; key: string | number; pageValue: number; active?: boolean; disabled?: boolean }> = [
    {
      label: '上一頁',
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
    label: '下一頁',
    key: 'next',
    pageValue: page.value + 1,
    disabled: page.value >= totalPages.value,
  });

  return buttons;
});

// 4. 套用價格區間
const applyPriceRange = (option?: (typeof furniturePriceRangeOptions)[number]): void => {
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
    const target = furniturePriceRangeOptions.find((option) => option.value === optionValue);
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

// 7. 點擊卡片跳轉詳情
const handleCardClick = (id: string): void => {
  void router.push(`/furniture/${id}`);
};

// 8. 點擊分頁
const handlePageClick = (targetPage: number): void => {
  void setPage(targetPage);
};

// 9. 從 URL 價格條件反推快捷區間狀態
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
      <!-- 左側篩選欄 -->
      <aside class="mf">
        <div class="sbar">
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

        <button
          type="button"
          class="clear-filter-btn"
          @click="handleClearFilters"
        >
          {{ t('marketplace.filter.clearAll') }}
        </button>
      </aside>

      <!-- 中間列表區 -->
      <main class="mr">
        <button
          type="button"
          class="filter-toggle-btn"
        >篩選條件</button>

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
            <span
              v-if="listing.conditionLevel === 'brand_new'"
              class="mnew"
            >{{ t('marketplace.filter.newBadge') }}</span>
            <div
              class="mimg"
              :class="listing.imageUrl ? '' : 'pat'"
            >
              <img
                v-if="listing.imageUrl"
                :src="listing.imageUrl"
                :alt="listing.title"
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
          aria-label="家具市集分頁"
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
        aria-label="家具市集展示廣告"
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
  min-height: calc(100vh - var(--nav-h, 52px));
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
  min-height: calc(100vh - var(--nav-h, 52px));
  overflow: visible;
  scrollbar-width: none;
  background: var(--sur);
}

.mf::-webkit-scrollbar {
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
  min-height: calc(100vh - var(--nav-h, 52px));
  background: var(--sur-2);
}

.filter-toggle-btn {
  display: none;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: var(--ink);
  background: var(--sur-2);
  border: 1px solid var(--bdr);
  padding: 8px 14px;
  border-radius: 2px;
  cursor: pointer;
  font-family: inherit;
  margin-bottom: 12px;
  width: 100%;
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

/* 11. 響應式 - 平板與行動裝置：對齊 @media (max-width:900px) */
@media (max-width: 900px) {
  .mp {
    grid-template-columns: 1fr;
  }

  .mf {
    position: static;
    height: auto;
    padding: 14px;
  }

  .sbar {
    margin-left: -14px;
    margin-right: -14px;
    margin-top: -14px;
  }

  .filter-toggle-btn {
    display: inline-flex;
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

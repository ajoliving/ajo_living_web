<!--
 * 二手傢俬篩選頁。
 * 1. 提供上方篩選欄、下方商品結果與右側推廣欄。
 * 2. 對齊二手傢俬篩選版面的留白、字體與卡片層級。
 * 3. 保留分類 query、排序、展示模式、成色與照片篩選狀態。
-->
<script setup lang="ts">
import { computed, ref, watch } from 'vue';

import ListingAdRail from '@/shared/components/data-display/ListingAdRail.vue';
import AppGlassSelect from '@/shared/components/base/AppGlassSelect.vue';
import AppIcon from '@/shared/components/base/AppIcon.vue';

import { useFurniturePage } from './furniture';

type ResultViewMode = 'list' | 'grid';
type FurnitureFilterSortValue = 'latest' | 'price_asc' | 'price_desc';

const {
  area,
  areaOptions,
  categories,
  clearConditions,
  clearFilters,
  conditionOptions,
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

// 1. 管理右側結果展示模式
const viewMode = ref<ResultViewMode>('list');
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

// 1. 格式化發布時間
const formatListingTime = (publishedAt?: string | null, updatedAt?: string): string => {
  const source = publishedAt || updatedAt;
  if (!source) {
    return '';
  }
  return new Intl.DateTimeFormat('zh-HK', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  }).format(new Date(source));
};

// 2. 套用價格區間
const applyPriceRange = (option?: (typeof furniturePriceRangeOptions)[number]): void => {
  selectedPriceRange.value = option?.value ?? '';
  minPrice.value = String(option?.min ?? 0);
  maxPrice.value = String(option?.max ?? maxPriceLimit);
  syncFiltersToQuery();
};

// 3. 執行關鍵字搜尋
const handleSearch = (): void => {
  syncFiltersToQuery(1);
};

// 4. 清除篩選並重置本頁工具列狀態
const handleClearFilters = (): void => {
  selectedPriceRange.value = '';
  clearFilters();
};

// 5. 更新排序條件
const handleSortChange = (value: string): void => {
  sortBy.value = value as FurnitureFilterSortValue;
  syncFiltersToQuery(page.value);
};

// 6. 切換圖片篩選
const toggleWithPhotos = (): void => {
  withPhotos.value = !withPhotos.value;
  syncFiltersToQuery();
};

// 7. 從 URL 價格條件反推快捷區間狀態
watch(
  selectedPriceRangeFromValues,
  (value) => {
    selectedPriceRange.value = value;
  },
  { immediate: true },
);
</script>

<template>
  <main class="marketplace-filter-page">
    <section class="marketplace-filter-layout">
      <section class="marketplace-top-layout">
        <section class="marketplace-filter-panel">
          <section class="marketplace-filter-section">
            <h3>{{ t('marketplace.filter.categories') }}</h3>
            <div class="marketplace-option-list marketplace-option-list--inline">
              <button
                type="button"
                class="marketplace-option-row marketplace-option-row--inline"
                :class="{ 'marketplace-option-row--active': selectedCategoryKeys.length === 0 }"
                @click="selectedCategoryKeys = []; syncFiltersToQuery()"
              >
                <span class="marketplace-option-box" />
                <span>{{ t('marketplace.filter.conditionAll') }}</span>
              </button>
              <button
                v-for="category in categories"
                :key="category.key"
                type="button"
                class="marketplace-option-row marketplace-option-row--inline"
                :class="{ 'marketplace-option-row--active': selectedCategoryKeys.includes(category.key) }"
                @click="toggleCategory(category.key)"
              >
                <span class="marketplace-option-box" />
                <span>{{ category.label }}</span>
              </button>
            </div>
          </section>

          <section class="marketplace-filter-section">
            <h3>{{ t('marketplace.filter.priceRange') }}</h3>
            <div class="marketplace-option-list marketplace-option-list--inline">
              <button
                type="button"
                class="marketplace-option-row marketplace-option-row--inline"
                :class="{ 'marketplace-option-row--active': selectedPriceRange === '' }"
                @click="applyPriceRange()"
              >
                <span class="marketplace-option-box" />
                <span>{{ t('marketplace.filter.conditionAll') }}</span>
              </button>
              <button
                v-for="option in furniturePriceRangeOptions"
                :key="option.value"
                type="button"
                class="marketplace-option-row marketplace-option-row--inline"
                :class="{ 'marketplace-option-row--active': selectedPriceRange === option.value }"
                @click="applyPriceRange(option)"
              >
                <span class="marketplace-option-box" />
                <span>{{ option.label }}</span>
              </button>
            </div>
          </section>

          <section class="marketplace-filter-section">
            <h3>{{ t('marketplace.filter.condition') }}</h3>
            <div class="marketplace-option-list marketplace-option-list--inline">
              <button
                type="button"
                class="marketplace-option-row marketplace-option-row--inline"
                :class="{ 'marketplace-option-row--active': selectedConditions.length === 0 }"
                @click="clearConditions"
              >
                <span class="marketplace-option-box" />
                <span>{{ t('marketplace.filter.conditionAll') }}</span>
              </button>
              <button
                v-for="condition in conditionOptions"
                :key="condition.value"
                type="button"
                class="marketplace-option-row marketplace-option-row--inline"
                :class="{ 'marketplace-option-row--active': selectedConditions.includes(condition.value) }"
                @click="toggleCondition(condition.value)"
              >
                <span class="marketplace-option-box" />
                <span>{{ condition.label }}</span>
              </button>
            </div>
          </section>

          <section class="marketplace-filter-section">
            <h3>{{ t('marketplace.filter.visibility') }}</h3>
            <div class="marketplace-option-list marketplace-option-list--inline">
              <button
                v-for="option in visibilityOptions"
                :key="option.value"
                type="button"
                class="marketplace-option-row marketplace-option-row--inline"
                :class="{ 'marketplace-option-row--active': visibility === option.value }"
                @click="visibility = option.value; syncFiltersToQuery()"
              >
                <span class="marketplace-option-box" />
                <span>{{ option.label }}</span>
              </button>
            </div>
          </section>

          <section class="marketplace-filter-section">
            <h3>{{ t('marketplace.filter.area') }}</h3>
            <div class="marketplace-option-list marketplace-option-list--inline">
              <button
                v-for="option in areaOptions"
                :key="option.value"
                type="button"
                class="marketplace-option-row marketplace-option-row--inline"
                :class="{ 'marketplace-option-row--active': area === option.value }"
                @click="area = option.value; syncFiltersToQuery()"
              >
                <span class="marketplace-option-box" />
                <span>{{ option.label }}</span>
              </button>
            </div>
          </section>

          <div class="marketplace-search-toolbar">
            <span>{{ t('marketplace.mine.search') }}</span>
            <div class="marketplace-search-input">
              <AppIcon
                name="search"
                :size="16"
              />
              <input
                v-model="keyword"
                type="search"
                :placeholder="t('marketplace.list.searchPlaceholder')"
                @keyup.enter="handleSearch"
              />
            </div>
            <button
              type="button"
              class="marketplace-toolbar-button marketplace-toolbar-button--primary"
              @click="handleSearch"
            >
              {{ t('marketplace.filter.searchAction') }}
            </button>
            <button
              type="button"
              class="marketplace-toolbar-icon-button"
              :class="{ 'marketplace-toolbar-icon-button--active': withPhotos }"
              :aria-label="t('marketplace.filter.onlyPhotos')"
              :title="t('marketplace.filter.onlyPhotos')"
              :data-tooltip="t('marketplace.filter.onlyPhotos')"
              @click="toggleWithPhotos"
            >
              <AppIcon
                name="picture"
                :size="17"
              />
            </button>
            <button
              type="button"
              class="marketplace-toolbar-icon-button"
              :class="{ 'marketplace-toolbar-icon-button--active': viewMode === 'grid' }"
              :aria-label="t('marketplace.filter.gridView')"
              :title="t('marketplace.filter.gridView')"
              :data-tooltip="t('marketplace.filter.gridView')"
              @click="viewMode = viewMode === 'grid' ? 'list' : 'grid'"
            >
              <AppIcon
                name="layout-grid"
                :size="17"
              />
            </button>
            <button
              type="button"
              class="marketplace-toolbar-icon-button"
              :aria-label="t('marketplace.filter.clearAll')"
              :title="t('marketplace.filter.clearAll')"
              :data-tooltip="t('marketplace.filter.clearAll')"
              @click="handleClearFilters"
            >
              <AppIcon
                name="trash"
                :size="17"
              />
            </button>
            <AppGlassSelect
              class="marketplace-sort-menu"
              :model-value="sortBy"
              :options="sortOptions"
              panel-max-height="14rem"
              @change="handleSortChange"
            />
          </div>
      </section>

        <ListingAdRail
          channel="furniture"
          variant="top"
        />
      </section>

      <section class="marketplace-content-layout">
        <section class="marketplace-filter-results min-w-0">
      <div class="marketplace-results-topline">
        <p>{{ loading ? t('common.status.loading') : t('marketplace.filter.resultsFound', { count: totalResults }) }}</p>
      </div>

      <div
        v-if="loading"
        class="marketplace-empty"
      >
        {{ t('common.status.loading') }}
      </div>

      <div
        v-else-if="listings.length === 0"
        class="marketplace-empty"
      >
        {{ t('marketplace.list.emptyTitle') }}
      </div>

      <div
        v-else
        class="marketplace-results-grid"
        :class="viewMode === 'grid' ? 'marketplace-results-grid--grid' : 'marketplace-results-grid--list'"
      >
        <RouterLink
          v-for="(listing, index) in listings"
          :key="listing.id"
          :to="{ path: `/furniture/listing/${listing.id}`, query: { from: 'filter' } }"
          class="marketplace-result-card group"
          :class="{ 'marketplace-result-card--list': viewMode === 'list' }"
        >
          <div class="marketplace-result-card__media">
            <img
              v-if="listing.imageUrl"
              :src="listing.imageUrl"
              :alt="listing.title"
            />
            <div
              v-else
              class="marketplace-result-card__placeholder"
            >
              <AppIcon
                name="picture"
                :size="42"
              />
            </div>
            <span
              v-if="index === 0 || selectedConditions.includes('brand_new')"
              class="marketplace-result-card__badge"
            >
              {{ t('marketplace.filter.newBadge') }}
            </span>
            <button
              type="button"
              class="favorite-card-button"
              :class="listing.isFavorited ? 'favorite-card-button--active' : ''"
              :aria-label="listing.isFavorited ? t('marketplace.detail.favoritedAction') : t('marketplace.detail.favoriteAction')"
              :disabled="isFavoriteUpdating(listing.id)"
              @click.prevent.stop="toggleFavorite(listing.id)"
            >
              <AppIcon
                name="star"
                :size="16"
              />
            </button>
          </div>

          <div class="marketplace-result-card__body">
            <div class="marketplace-result-card__meta">
              <span>{{ listing.districtLabel }}</span>
              <span>{{ listing.categoryLabel }}</span>
              <span>{{ listing.conditionLabel }}</span>
            </div>
            <h2>{{ listing.title }}</h2>
            <p>{{ listing.summary }}</p>
            <dl class="marketplace-result-card__specs">
              <div>
                <dt>{{ t('marketplace.filter.priceRange') }}</dt>
                <dd>{{ listing.price }}</dd>
              </div>
              <div>
                <dt>{{ t('marketplace.filter.area') }}</dt>
                <dd>{{ listing.districtLabel }}</dd>
              </div>
              <div>
                <dt>{{ t('marketplace.filter.categories') }}</dt>
                <dd>{{ listing.categoryLabel }}</dd>
              </div>
              <div>
                <dt>{{ t('marketplace.filter.condition') }}</dt>
                <dd>{{ listing.conditionLabel }}</dd>
              </div>
            </dl>
            <div class="marketplace-result-card__footer">
              <span>{{ formatListingTime(listing.publishedAt, listing.updatedAt) }}</span>
              <span>{{ t('common.action.viewDetail') }}</span>
            </div>
          </div>
        </RouterLink>
      </div>

      <div class="mt-16 flex items-center justify-center gap-2">
        <button
          type="button"
          class="filter-page-button text-text-muted opacity-50"
          :aria-label="t('marketplace.filter.previousPage')"
          :disabled="page <= 1"
          @click="setPage(page - 1)"
        >
          <AppIcon
            name="arrow-left"
            :size="16"
          />
        </button>
        <button
          type="button"
          class="h-10 w-10 rounded-full bg-primary text-[14px] font-medium leading-none tracking-[0.02em] text-primary-contrast"
        >
          {{ page }}
        </button>
        <span class="px-2 text-text-muted">/</span>
        <span class="px-2 text-text-muted">{{ totalPages }}</span>
        <button
          type="button"
          class="filter-page-button text-text-muted"
          :aria-label="t('marketplace.filter.nextPage')"
          :disabled="page >= totalPages"
          @click="setPage(page + 1)"
        >
          <AppIcon
            name="arrow-right"
            :size="16"
          />
        </button>
      </div>
        </section>

      <ListingAdRail
        channel="furniture"
        variant="bottom"
      />
      </section>
    </section>
  </main>
</template>

<style scoped>
.marketplace-filter-page {
  display: block;
  max-width: var(--layout-page-max-width);
  min-height: calc(100vh - var(--app-header-offset, 0rem));
  margin: 0 auto;
  padding: 1.75rem var(--layout-page-padding-inline) 4rem;
  color: rgb(var(--color-text));
  overflow: visible;
}

.marketplace-filter-layout {
  display: grid;
  gap: 2rem;
}

.marketplace-top-layout,
.marketplace-content-layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr) clamp(13rem, 18vw, 17rem);
  align-items: start;
  gap: 1.5rem;
}

.marketplace-filter-panel {
  display: flex;
  flex-direction: column;
  gap: 0;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface-raised));
  overflow: visible;
  padding: 0.85rem 1rem;
}

.marketplace-filter-panel > .marketplace-filter-section,
.marketplace-search-toolbar {
  display: grid;
  grid-template-columns: minmax(4.75rem, 6.5rem) minmax(0, 1fr);
  align-items: start;
  gap: 0.35rem 0.45rem;
  padding: 0.35rem 0;
}

.marketplace-search-toolbar {
  grid-template-columns: minmax(4.75rem, 6.5rem) minmax(10rem, 1fr) auto auto auto auto minmax(11rem, 15rem);
}

.marketplace-filter-results {
  min-width: 0;
  overflow: visible;
}

.marketplace-filter-section h3,
.marketplace-search-toolbar > span {
  margin: 0;
  color: rgb(var(--color-text-muted));
  font-size: 0.8rem;
  font-weight: 600;
  letter-spacing: 0.1em;
  line-height: 1;
  padding-top: 0.5rem;
  text-transform: uppercase;
}

.marketplace-option-list {
  display: grid;
  gap: 0.22rem;
}

.marketplace-option-list--inline {
  display: flex;
  flex-wrap: wrap;
  gap: 0.12rem;
}

.marketplace-option-row {
  display: flex;
  width: 100%;
  cursor: pointer;
  align-items: center;
  gap: 0.5rem;
  color: rgb(var(--color-text-muted));
  font-size: 0.95rem;
  line-height: 1.35;
  text-align: left;
  transition:
    background-color 0.2s ease,
    border-color 0.2s ease,
    color 0.2s ease;
}

.marketplace-option-row--inline {
  width: auto;
}

.marketplace-option-list--inline .marketplace-option-row {
  min-height: 1.75rem;
  width: auto;
  align-items: center;
  justify-content: center;
  gap: 0;
  border: 1px solid transparent;
  border-radius: 999px;
  padding: 0 0.38rem;
  color: rgb(var(--color-text-muted));
  font-size: 0.84rem;
  font-weight: 600;
  line-height: 1;
}

.marketplace-option-list--inline .marketplace-option-box {
  display: none;
}

.marketplace-option-list--inline .marketplace-option-row:hover {
  color: rgb(var(--color-text));
}

.marketplace-option-list--inline .marketplace-option-row:focus-visible {
  border-color: rgb(var(--color-primary));
  outline: 2px solid rgb(var(--color-primary));
  outline-offset: 2px;
}

.marketplace-option-list--inline .marketplace-option-row--active {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary-soft) / 0.48);
  color: rgb(var(--color-primary));
}

.marketplace-option-box {
  display: inline-flex;
  width: 1rem;
  height: 1rem;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.25rem;
  background: rgb(var(--color-surface-raised));
}

.marketplace-option-row--active .marketplace-option-box {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary));
  box-shadow: inset 0 0 0 3px rgb(var(--color-surface-raised));
}

.marketplace-search-input {
  display: flex;
  min-height: 2rem;
  align-items: center;
  gap: 0.5rem;
  border: 0;
  border-radius: 0.25rem;
  background: rgb(var(--color-surface-muted));
  padding: 0 0.75rem;
  color: rgb(var(--color-text));
  font-size: 0.95rem;
  line-height: 1.35;
}

.marketplace-search-input input {
  width: 100%;
  border: 0;
  background: transparent;
  color: inherit;
  outline: 0;
}

.marketplace-toolbar-button,
.marketplace-toolbar-icon-button {
  display: inline-flex;
  min-height: 2rem;
  align-items: center;
  justify-content: center;
  border-radius: 0.25rem;
  font-size: 0.88rem;
  font-weight: 700;
  line-height: 1;
  white-space: nowrap;
}

.marketplace-toolbar-button {
  padding: 0 0.85rem;
}

.marketplace-toolbar-button--primary {
  background: rgb(var(--color-primary));
  color: rgb(var(--color-primary-contrast));
}

.marketplace-toolbar-icon-button {
  position: relative;
  width: 2rem;
  border: 1px solid rgb(var(--color-border));
  color: rgb(var(--color-text-muted));
}

.marketplace-toolbar-icon-button:hover,
.marketplace-toolbar-icon-button:focus-visible,
.marketplace-toolbar-icon-button--active {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary-soft) / 0.48);
  color: rgb(var(--color-primary));
}

.marketplace-toolbar-icon-button:focus-visible {
  outline: 2px solid rgb(var(--color-primary));
  outline-offset: 2px;
}

.marketplace-toolbar-icon-button::after {
  position: absolute;
  bottom: calc(100% + 0.45rem);
  left: 50%;
  z-index: 10;
  max-width: 8rem;
  border-radius: 0.25rem;
  background: rgb(var(--color-text));
  padding: 0.35rem 0.5rem;
  color: rgb(var(--color-surface-raised));
  content: attr(data-tooltip);
  font-size: 0.72rem;
  font-weight: 700;
  line-height: 1.2;
  opacity: 0;
  pointer-events: none;
  text-align: center;
  transform: translate(-50%, 0.2rem);
  transition:
    opacity 0.16s ease,
    transform 0.16s ease;
  white-space: nowrap;
}

.marketplace-toolbar-icon-button:hover::after,
.marketplace-toolbar-icon-button:focus-visible::after {
  opacity: 1;
  transform: translate(-50%, 0);
}

.marketplace-sort-menu {
  position: relative;
  display: flex;
  min-width: 0;
  align-items: center;
}

.marketplace-sort-menu :deep(.app-glass-select__trigger) {
  width: 100%;
  min-height: 2rem;
  cursor: pointer;
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.25rem;
  background: rgb(var(--color-surface-muted));
  padding: 0 1.9rem 0 0.65rem;
  color: rgb(var(--color-text));
  font-size: 0.88rem;
  font-weight: 700;
  line-height: 1;
  box-shadow: none;
}

.marketplace-sort-menu :deep(.app-glass-select__trigger:hover),
.marketplace-sort-menu :deep(.app-glass-select__trigger[aria-expanded='true']) {
  border-color: rgb(var(--color-border));
  box-shadow: none;
}

.marketplace-sort-menu :deep(.app-glass-select__trigger svg) {
  position: absolute;
  right: 0.55rem;
  color: rgb(var(--color-text-muted));
  pointer-events: none;
}

.marketplace-sort-menu :deep(.app-glass-select__trigger span) {
  font-size: 0.88rem;
  font-weight: 700;
  line-height: 1;
}

.marketplace-results-topline {
  display: grid;
  gap: 1.25rem;
  margin-bottom: 2rem;
}

.marketplace-results-topline p {
  margin: 0;
  color: rgb(var(--color-text-muted));
  font-size: 1.125rem;
  line-height: 1.6;
}

.marketplace-results-grid {
  display: grid;
  gap: 2rem;
}

.marketplace-results-grid--list {
  grid-template-columns: minmax(0, 1fr);
  gap: 1.25rem;
}

.marketplace-result-card {
  display: grid;
  overflow: hidden;
  height: 22.625rem;
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.75rem;
  background: rgb(var(--color-surface-raised));
  color: inherit;
  box-shadow: 0 4px 24px rgb(0 0 0 / 0.03);
  text-decoration: none;
  transition: box-shadow 0.3s ease;
}

.marketplace-result-card:hover {
  box-shadow: 0 8px 32px rgb(0 0 0 / 0.06);
}

.marketplace-result-card__media {
  position: relative;
  aspect-ratio: 4 / 3;
  min-height: 0;
  overflow: hidden;
  background: rgb(var(--color-surface-raised));
}

.marketplace-result-card__media img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 0.35s ease;
}

.marketplace-result-card:hover .marketplace-result-card__media img {
  transform: scale(1.03);
}

.marketplace-result-card__placeholder {
  display: grid;
  height: 100%;
  place-items: center;
  color: rgb(var(--color-text-muted));
}

.marketplace-result-card__badge {
  position: absolute;
  top: 0.65rem;
  left: 0.65rem;
  z-index: 2;
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.35rem;
  background: rgb(var(--color-surface) / 0.92);
  padding: 0.25rem 0.45rem;
  color: rgb(var(--color-text));
  font-size: 0.72rem;
  font-weight: 950;
  line-height: 1;
}

.favorite-card-button {
  position: absolute;
  right: 0.65rem;
  top: 0.65rem;
  z-index: 2;
  display: inline-flex;
  width: 2rem;
  height: 2rem;
  align-items: center;
  justify-content: center;
  border: 1px solid rgb(255 255 255 / 0.72);
  border-radius: 999px;
  background: rgb(255 255 255 / 0.88);
  color: rgb(var(--color-text-muted));
  transition:
    background 0.2s ease,
    color 0.2s ease,
    opacity 0.2s ease;
}

.favorite-card-button--active {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary));
  color: rgb(var(--color-primary-contrast));
}

.favorite-card-button:disabled {
  opacity: 0.55;
}

.marketplace-result-card__body {
  display: grid;
  grid-template-rows: auto auto auto auto minmax(0, 1fr);
  min-width: 0;
  min-height: 0;
  gap: 0.75rem;
  padding: 0.875rem 1rem;
}

.marketplace-result-card__meta,
.marketplace-result-card__footer {
  display: flex;
  flex-wrap: wrap;
  min-width: 0;
  min-height: 0;
  gap: 0.45rem;
  color: rgb(var(--color-text-muted));
  font-size: 0.75rem;
  font-weight: 800;
}

.marketplace-result-card__meta span,
.marketplace-result-card__footer span {
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.marketplace-result-card h2 {
  margin: 0;
  overflow: hidden;
  color: rgb(var(--color-text));
  font-family: var(--font-display);
  font-size: 1.125rem;
  font-weight: 500;
  line-height: 1.28;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.marketplace-result-card p {
  display: -webkit-box;
  min-height: 2.7rem;
  margin: 0;
  overflow: hidden;
  color: rgb(var(--color-text-muted));
  font-size: 0.875rem;
  line-height: 1.45;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.marketplace-result-card__specs {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  min-height: 0;
  gap: 0.65rem;
  margin: 0;
  overflow: hidden;
}

.marketplace-result-card__specs dt {
  color: rgb(var(--color-text-muted));
  font-size: 0.7rem;
  font-weight: 600;
}

.marketplace-result-card__specs dd {
  margin: 0.18rem 0 0;
  overflow: hidden;
  font-size: 0.84rem;
  font-weight: 700;
  line-height: 1.35;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.marketplace-result-card__footer {
  align-self: end;
  justify-content: space-between;
  border-top: 1px solid rgb(var(--color-border));
  padding-top: 0.75rem;
}

.marketplace-empty {
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface-raised));
  padding: 2rem;
  color: rgb(var(--color-text-muted));
  text-align: center;
}

.filter-page-button {
  display: flex;
  height: 2.5rem;
  width: 2.5rem;
  align-items: center;
  justify-content: center;
  border: 1px solid rgb(var(--color-border));
  border-radius: 9999px;
  font-size: 0.875rem;
  font-weight: 500;
  letter-spacing: 0.02em;
  line-height: 1;
  transition:
    border-color 0.2s ease,
    color 0.2s ease;
}

.filter-page-button:hover:not(:disabled) {
  border-color: rgb(var(--color-primary));
  color: rgb(var(--color-primary));
}

.filter-page-button:disabled {
  cursor: not-allowed;
  opacity: 0.45;
}

@media (max-width: 767px) {
  .marketplace-filter-page {
    padding: 1.25rem var(--layout-page-padding-inline) 3rem;
  }

  .marketplace-top-layout,
  .marketplace-content-layout {
    grid-template-columns: minmax(0, 1fr);
  }

  .marketplace-filter-panel > .marketplace-filter-section,
  .marketplace-search-toolbar {
    grid-template-columns: minmax(0, 1fr);
  }

  .marketplace-search-toolbar {
    grid-template-columns: minmax(0, 1fr) repeat(3, 2rem);
  }

  .marketplace-search-toolbar > span,
  .marketplace-search-input,
  .marketplace-toolbar-button,
  .marketplace-sort-menu {
    grid-column: 1 / -1;
  }
}

@media (min-width: 640px) {
  .marketplace-results-grid--grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (min-width: 768px) {
  .marketplace-result-card--list {
    height: auto;
    min-height: 6.875rem;
    grid-template-columns: 10.25rem minmax(0, 1fr);
  }

  .marketplace-result-card--list .marketplace-result-card__media {
    height: 100%;
    min-height: 6.875rem;
    align-self: stretch;
    aspect-ratio: auto;
    border-radius: 0.75rem 0 0 0.75rem;
  }

  .marketplace-result-card--list .marketplace-result-card__body {
    grid-template-rows: auto;
    min-height: 6.875rem;
    gap: 0.24rem;
    padding: 0.42rem 0.62rem;
  }

  .marketplace-result-card--list p {
    display: -webkit-box;
    min-height: 0;
    -webkit-line-clamp: 1;
  }

  .marketplace-result-card--list .marketplace-result-card__specs {
    gap: 0.18rem 0.4rem;
  }

  .marketplace-result-card--list .marketplace-result-card__footer {
    padding-top: 0.28rem;
  }
}

@media (min-width: 1280px) {
  .marketplace-results-grid--grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}
</style>

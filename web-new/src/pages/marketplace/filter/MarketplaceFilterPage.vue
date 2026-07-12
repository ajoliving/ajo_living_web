<!--
 * 二手交易篩選頁。
 * 1. 提供左側篩選欄與右側商品結果卡片。
 * 2. 對齊精品 marketplace 篩選版面的留白、字體與卡片層級。
 * 3. 保留分類 query、排序、展示模式、成色與照片篩選狀態。
-->
<script setup lang="ts">
import { ref } from 'vue';

import ListingSideAds from '@/shared/components/ads/ListingSideAds.vue';
import AppIcon from '@/shared/components/base/AppIcon.vue';

import { useMarketplaceFilterPage } from './filter';

type ResultViewMode = 'list' | 'grid';

const {
  area,
  areaOptions,
  categories,
  clearFilters,
  conditionOptions,
  isFavoriteUpdating,
  keyword,
  listings,
  loading,
  maxPrice,
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
  withPhotos,
} = useMarketplaceFilterPage();

// 1. 管理右側結果展示模式
const viewMode = ref<ResultViewMode>('grid');
</script>

<template>
  <main class="marketplace-filter-page">
    <aside class="marketplace-filter-sidebar">
      <div class="flex items-center justify-between border-b border-border pb-4">
        <h1 class="font-display text-[24px] font-medium leading-[1.4] text-text">
          {{ t('marketplace.filter.title') }}
        </h1>
        <button
          type="button"
          class="text-[14px] font-medium tracking-[0.02em] text-primary transition-colors hover:text-primary/80"
          @click="clearFilters"
        >
          {{ t('marketplace.filter.clearAll') }}
        </button>
      </div>

      <section class="space-y-4">
        <h2 class="filter-section-title">
          {{ t('marketplace.mine.search') }}
        </h2>
        <label class="filter-select-field">
          <AppIcon
            name="search"
            class="left-icon"
            :size="17"
          />
          <input
            v-model="keyword"
            type="search"
            class="filter-text-input pl-9"
            :placeholder="t('marketplace.list.searchPlaceholder')"
            @keyup.enter="syncFiltersToQuery()"
            @change="syncFiltersToQuery()"
          />
        </label>
      </section>

      <section class="space-y-4">
        <h2 class="filter-section-title">
          {{ t('marketplace.filter.categories') }}
        </h2>
        <div class="space-y-3">
          <label
            v-for="category in categories"
            :key="category.key"
            class="filter-checkbox-row group"
          >
            <span class="flex items-center gap-3">
              <input
                type="checkbox"
                class="filter-checkbox"
                :checked="selectedCategoryKeys.includes(category.key)"
                @change="toggleCategory(category.key)"
              />
              <span class="transition-colors group-hover:text-text">{{ category.label }}</span>
            </span>
            <span class="text-sm text-text-muted">{{ category.count }}</span>
          </label>
        </div>
      </section>

      <section class="space-y-4 border-t border-border pt-6">
        <h2 class="filter-section-title">
          {{ t('marketplace.filter.priceRange') }}
        </h2>
        <div class="grid grid-cols-[1fr_auto_1fr] items-center gap-3">
          <label class="filter-price-field">
            <span>$</span>
            <input
              v-model="minPrice"
              type="number"
              :placeholder="t('marketplace.filter.minPrice')"
              class="filter-text-input pl-7"
              @change="syncFiltersToQuery()"
            />
          </label>
          <span class="text-border">-</span>
          <label class="filter-price-field">
            <span>$</span>
            <input
              v-model="maxPrice"
              type="number"
              :placeholder="t('marketplace.filter.maxPrice')"
              class="filter-text-input pl-7"
              @change="syncFiltersToQuery()"
            />
          </label>
        </div>
      </section>

      <section class="space-y-4 border-t border-border pt-6">
        <h2 class="filter-section-title">
          {{ t('marketplace.filter.condition') }}
        </h2>
        <div class="flex flex-wrap gap-2">
          <button
            v-for="condition in conditionOptions"
            :key="condition.value"
            type="button"
            class="filter-chip"
            :class="selectedConditions.includes(condition.value) ? 'filter-chip-active' : 'filter-chip-idle'"
            @click="toggleCondition(condition.value)"
          >
            {{ condition.label }}
          </button>
        </div>
      </section>

      <section class="space-y-4 border-t border-border pt-6">
        <h2 class="filter-section-title">
          {{ t('marketplace.filter.area') }}
        </h2>
        <label class="filter-select-field">
          <AppIcon
            name="location"
            class="left-icon"
            :size="17"
          />
          <select
            v-model="area"
            class="filter-text-input cursor-pointer appearance-none pl-9 pr-8"
            @change="syncFiltersToQuery()"
          >
            <option
              v-for="option in areaOptions"
              :key="option.value"
              :value="option.value"
            >
              {{ option.label }}
            </option>
          </select>
          <AppIcon
            name="chevron-down"
            class="right-icon"
            :size="17"
          />
        </label>
      </section>

      <section class="border-t border-border pt-6">
        <label class="flex cursor-pointer items-center justify-between gap-4 text-[16px] leading-[1.6] text-text-muted transition-colors hover:text-text">
          <span>{{ t('marketplace.filter.onlyPhotos') }}</span>
          <span class="relative inline-block h-6 w-10">
            <input
              v-model="withPhotos"
              type="checkbox"
              class="peer sr-only"
              @change="syncFiltersToQuery()"
            />
            <span class="block h-6 w-10 rounded-full bg-border transition-colors peer-checked:bg-primary" />
            <span class="absolute left-1 top-1 h-4 w-4 rounded-full bg-white shadow-sm transition-transform peer-checked:translate-x-4" />
          </span>
        </label>
      </section>
    </aside>

    <section class="marketplace-filter-results min-w-0 flex-1">
      <div class="mb-8 grid gap-4 md:grid-cols-[minmax(0,1fr)_auto] md:items-center">
        <p class="min-w-0 text-[18px] leading-[1.6] text-text-muted">
          {{ t('marketplace.filter.resultsFound', { count: totalResults }) }}
        </p>
        <div class="sort-control flex min-w-0 flex-wrap items-center gap-3 md:justify-end">
          <div
            class="view-mode-toggle"
            role="group"
            :aria-label="t('marketplace.filter.viewMode')"
          >
            <button
              type="button"
              class="view-mode-button"
              :class="viewMode === 'list' ? 'view-mode-button-active' : 'view-mode-button-idle'"
              :aria-label="t('marketplace.filter.listView')"
              :aria-pressed="viewMode === 'list'"
              @click="viewMode = 'list'"
            >
              <AppIcon
                name="layout-list"
                :size="18"
              />
            </button>
            <button
              type="button"
              class="view-mode-button"
              :class="viewMode === 'grid' ? 'view-mode-button-active' : 'view-mode-button-idle'"
              :aria-label="t('marketplace.filter.gridView')"
              :aria-pressed="viewMode === 'grid'"
              @click="viewMode = 'grid'"
            >
              <AppIcon
                name="layout-grid"
                :size="18"
              />
            </button>
          </div>
          <label class="shrink-0 text-[16px] leading-[1.6] text-text-muted">{{ t('marketplace.filter.sortBy') }}</label>
          <label class="relative w-[9.75rem] max-w-full">
            <select
              v-model="sortBy"
              class="w-full cursor-pointer appearance-none truncate border-none bg-transparent pr-5 text-[14px] font-medium leading-none tracking-[0.02em] text-text outline-none focus:ring-0"
              @change="syncFiltersToQuery()"
            >
              <option
                v-for="option in sortOptions"
                :key="option.value"
                :value="option.value"
              >
                {{ option.label }}
              </option>
            </select>
            <AppIcon
              name="chevron-down"
              class="pointer-events-none absolute right-0 top-1/2 -translate-y-1/2 text-text-muted"
              :size="16"
            />
          </label>
        </div>
      </div>

      <div
        v-if="loading"
        class="filter-loading"
      >
        {{ t('common.status.loading') }}
      </div>

      <div
        v-else
        class="marketplace-results-grid"
        :class="viewMode === 'grid' ? 'marketplace-results-grid--grid' : 'marketplace-results-grid--list'"
      >
        <RouterLink
          v-for="(listing, index) in listings"
          :key="listing.id"
          :to="{ path: `/marketplace/listing/${listing.id}`, query: { from: 'filter' } }"
          class="marketplace-result-card group"
          :class="viewMode === 'list' ? 'marketplace-result-card--list' : ''"
        >
          <div class="marketplace-result-card__media relative overflow-hidden bg-surface-muted">
            <img
              v-if="listing.imageUrl"
              :src="listing.imageUrl"
              :alt="listing.title"
              class="h-full w-full object-cover transition-transform duration-500 group-hover:scale-105"
            />
            <div
              v-else
              class="flex h-full w-full items-center justify-center bg-border"
            >
              <AppIcon
                name="picture"
                class="text-text-muted"
                :size="48"
              />
            </div>
            <button
              type="button"
              class="favorite-card-button"
              :class="listing.isFavorited ? 'favorite-card-button--active' : ''"
              :aria-label="listing.isFavorited ? t('marketplace.detail.favoritedAction') : t('marketplace.detail.favoriteAction')"
              :disabled="isFavoriteUpdating(listing.id)"
              @click.prevent="toggleFavorite(listing.id)"
            >
              <AppIcon
                name="star"
                :size="16"
              />
            </button>
          </div>

          <div class="marketplace-result-card__body">
            <div class="marketplace-result-card__heading">
              <h2 class="marketplace-result-card__title">
                {{ listing.title }}
              </h2>
              <p class="marketplace-result-card__price">
                {{ listing.price }}
              </p>
            </div>

            <p class="marketplace-result-card__summary">
              {{ listing.summary }}
            </p>

            <div class="marketplace-result-card__meta">
              <div class="flex items-center gap-2 text-sm text-text-muted">
                <AppIcon
                  name="location"
                  :size="16"
                />
                <span>{{ listing.districtLabel }}</span>
              </div>
              <span class="text-xs text-text-muted">{{ index < 2 ? `${index + 2}h ago` : `${index - 1}d ago` }}</span>
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
          class="filter-current-page"
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

    <ListingSideAds channel="furniture" />
  </main>
</template>

<style scoped>
.marketplace-filter-page {
  display: flex;
  width: 100%;
  max-width: var(--layout-page-max-width);
  min-height: calc(100svh - var(--app-header-offset, 0rem));
  gap: 16px;
  margin: 0 auto;
  padding: 24px var(--layout-page-padding-inline) 72px;
  color: rgb(var(--color-text));
  overflow: visible;
}

.marketplace-filter-sidebar {
  display: flex;
  width: 220px;
  flex-shrink: 0;
  flex-direction: column;
  gap: 16px;
  border-right: 1px solid rgb(var(--color-border));
  padding-right: 16px;
  overflow: visible;
}

.marketplace-filter-results {
  min-width: 0;
  flex: 1;
  overflow: visible;
}

.view-mode-toggle {
  display: inline-flex;
  overflow: hidden;
  border: 1px solid rgb(var(--color-border));
  border-radius: 2px;
  background: rgb(var(--color-surface));
}

.view-mode-button {
  display: inline-flex;
  height: 34px;
  width: 34px;
  align-items: center;
  justify-content: center;
  transition:
    background-color 0.2s ease,
    color 0.2s ease;
}

.view-mode-button:focus-visible {
  outline: 2px solid rgb(var(--color-primary));
  outline-offset: 2px;
}

.view-mode-button-active {
  background: rgb(var(--color-primary));
  color: rgb(var(--color-primary-contrast));
}

.view-mode-button-idle {
  color: rgb(var(--color-text-muted));
}

.view-mode-button-idle:hover {
  color: rgb(var(--color-primary));
}

.marketplace-results-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 12px;
}

.filter-section-title {
  color: rgb(var(--color-text-muted));
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0.16em;
  line-height: 1;
  text-transform: uppercase;
}

.filter-checkbox-row {
  display: flex;
  cursor: pointer;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  color: rgb(var(--color-text-muted));
  font-size: 12px;
  line-height: 1.45;
}

.filter-checkbox {
  height: 14px;
  width: 14px;
  border-radius: 2px;
  border-color: rgb(var(--color-border));
  background-color: rgb(var(--color-surface-raised));
  color: rgb(var(--color-primary));
}

.filter-checkbox:focus {
  --tw-ring-color: rgb(var(--color-primary));
  --tw-ring-offset-width: 0;
}

.filter-price-field,
.filter-select-field {
  position: relative;
  display: block;
}

.filter-price-field > span,
.filter-select-field .left-icon,
.filter-select-field .right-icon {
  position: absolute;
  top: 50%;
  z-index: 1;
  transform: translateY(-50%);
  color: rgb(var(--color-text-muted));
  pointer-events: none;
}

.filter-price-field > span {
  left: 9px;
  color: rgb(var(--color-border));
}

.filter-select-field .left-icon {
  left: 9px;
}

.filter-select-field .right-icon {
  right: 9px;
}

.filter-text-input {
  height: 36px;
  width: 100%;
  border: 1px solid rgb(var(--color-border));
  border-radius: 2px;
  background: rgb(var(--color-surface-muted));
  padding-top: 6px;
  padding-bottom: 6px;
  color: rgb(var(--color-text));
  font-size: 12px;
  line-height: 1.4;
}

.filter-text-input:focus {
  --tw-ring-color: rgb(var(--color-primary));
  --tw-ring-offset-width: 0;
}

.filter-chip {
  border-radius: 2px;
  border-width: 1px;
  padding: 7px 10px;
  font-size: 11px;
  font-weight: 500;
  letter-spacing: 0;
  line-height: 1;
  transition:
    border-color 0.2s ease,
    background-color 0.2s ease,
    color 0.2s ease;
}

.filter-chip-active {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary));
  color: rgb(var(--color-primary-contrast));
}

.filter-chip-idle {
  border-color: rgb(var(--color-border));
  color: rgb(var(--color-text-muted));
}

.filter-chip-idle:hover {
  border-color: rgb(var(--color-text-muted));
  color: rgb(var(--color-text));
}

.marketplace-results-grid--list {
  gap: 10px;
}

.filter-loading {
  border: 1px solid rgb(var(--color-border));
  border-radius: 2px;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text-muted));
  font-size: 12px;
  margin: 16px 0;
  padding: 16px;
}

.marketplace-result-card {
  display: flex;
  height: 252px;
  flex-direction: column;
  overflow: hidden;
  border: 1px solid rgb(var(--color-border));
  border-radius: 2px;
  background: rgb(var(--color-surface));
  color: inherit;
  text-decoration: none;
  box-shadow: none;
  transition:
    border-color 0.18s ease,
    box-shadow 0.18s ease;
}

.marketplace-result-card:hover {
  border-color: rgb(var(--color-primary));
  box-shadow: var(--shadow-soft);
}

.marketplace-result-card__media {
  position: relative;
  height: 140px;
  flex-shrink: 0;
}

.favorite-card-button {
  position: absolute;
  left: 8px;
  top: 8px;
  z-index: 2;
  display: inline-flex;
  width: 28px;
  height: 28px;
  align-items: center;
  justify-content: center;
  border: 1px solid rgb(255 255 255 / 0.72);
  border-radius: 2px;
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
  display: flex;
  min-width: 0;
  flex: 1;
  flex-direction: column;
  overflow: hidden;
  padding: 14px;
}

.marketplace-result-card__heading {
  display: flex;
  min-width: 0;
  align-items: flex-start;
  justify-content: space-between;
  gap: 8px;
}

.marketplace-result-card__title {
  min-width: 0;
  display: -webkit-box;
  overflow: hidden;
  margin: 0;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 1;
  color: rgb(var(--color-text));
  font-family: var(--font-sans);
  font-size: 13px;
  font-weight: 600;
  line-height: 1.45;
}

.marketplace-result-card__price {
  flex-shrink: 0;
  margin: 0;
  color: rgb(var(--color-primary));
  font-family: var(--font-sans);
  font-size: 14px;
  font-weight: 600;
  line-height: 1.2;
}

.marketplace-result-card__summary {
  display: -webkit-box;
  overflow: hidden;
  margin: 5px 0 0;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  color: rgb(var(--color-text-muted));
  flex: 1;
  font-size: 11px;
  line-height: 1.5;
}

.marketplace-result-card__meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  border-top: 1px solid rgb(var(--color-border));
  padding-top: 10px;
  color: rgb(var(--color-text-muted));
}

.filter-page-button {
  display: flex;
  height: 32px;
  width: 32px;
  align-items: center;
  justify-content: center;
  border: 1px solid rgb(var(--color-border));
  border-radius: 2px;
  font-size: 12px;
  font-weight: 500;
  letter-spacing: 0.02em;
  line-height: 1;
  transition:
    border-color 0.2s ease,
    color 0.2s ease;
}

.filter-current-page {
  width: 32px;
  height: 32px;
  border-radius: 2px;
  background: rgb(var(--color-primary));
  color: rgb(var(--color-primary-contrast));
  font-size: 12px;
  font-weight: 600;
}

.filter-page-button:hover:not(:disabled) {
  border-color: rgb(var(--color-primary));
  color: rgb(var(--color-primary));
}

@media (max-width: 767px) {
  .marketplace-filter-page {
    height: auto;
    min-height: calc(100svh - var(--app-header-offset, 0rem));
    flex-direction: column;
    gap: 18px;
    overflow: visible;
    padding: 18px var(--layout-page-padding-inline) 96px;
  }

  .marketplace-filter-sidebar {
    width: 100%;
    border-right: 0;
    border-bottom: 1px solid rgb(var(--color-border));
    padding-bottom: 16px;
    overflow: visible;
    padding-right: 0;
  }

  .marketplace-filter-results {
    overflow: visible;
  }
}

@media (min-width: 640px) {
  .marketplace-results-grid--grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (min-width: 768px) {
  .marketplace-result-card--list {
    height: 156px;
    min-height: 0;
    flex-direction: row;
  }

  .marketplace-result-card--list .marketplace-result-card__media {
    width: 180px;
    height: 100%;
    min-height: 0;
    flex-shrink: 0;
  }

  .marketplace-result-card--list .marketplace-result-card__summary {
    -webkit-line-clamp: 1;
  }
}

@media (min-width: 1280px) {
  .marketplace-results-grid--grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}
</style>

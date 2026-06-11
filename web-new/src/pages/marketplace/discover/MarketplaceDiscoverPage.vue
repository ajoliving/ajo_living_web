<!--
 * 二手交易 Discover 頁。
 * 1. 讀取推薦位與分類推薦資料。
 * 2. 以緊湊列表與瀑布流卡片承接二手入口。
 * 3. 保留分類跳轉與輪播切換能力。
-->
<script setup lang="ts">
import { RouterLink } from 'vue-router';
import { useI18n } from 'vue-i18n';

import AppIcon from '@/shared/components/base/AppIcon.vue';

import { useMarketplaceDiscoverPage } from './discover';

const { t } = useI18n();
const {
  categoryRows,
  getActiveCategorySlide,
  loading,
  primaryRecommendation,
  recommendedItems,
  showNextCategorySlide,
  showPreviousCategorySlide,
} = useMarketplaceDiscoverPage();
</script>

<template>
  <main class="market-discover">
    <section class="market-discover__hero">
      <div>
        <p class="market-discover__eyebrow">
          AJO Living
        </p>
        <h1>{{ t('nav.marketplace') }}</h1>
        <p>{{ t('home.newShell.modules.secondhand.description') }}</p>
      </div>
      <RouterLink
        to="/marketplace/filter"
        class="market-discover__filter-link"
      >
        {{ t('nav.filter') }}
      </RouterLink>
    </section>

    <section
      v-if="loading"
      class="market-discover__empty"
    >
      {{ t('common.status.loading') }}
    </section>

    <section
      v-if="primaryRecommendation"
      class="market-discover__feature"
    >
      <RouterLink
        :to="primaryRecommendation.to"
        class="market-feature-card market-feature-card--main"
      >
        <div class="market-feature-card__media">
          <img
            v-if="primaryRecommendation.imageUrl"
            :src="primaryRecommendation.imageUrl"
            :alt="primaryRecommendation.imageAlt"
          />
          <AppIcon
            v-else
            name="picture"
            :size="36"
          />
        </div>
        <div class="market-feature-card__body">
          <span>{{ t('marketplace.discover.recommended') }}</span>
          <h2>{{ primaryRecommendation.title }}</h2>
          <p>{{ primaryRecommendation.summary }}</p>
          <strong>{{ primaryRecommendation.price }}</strong>
        </div>
      </RouterLink>

      <div class="market-discover__side-list">
        <RouterLink
          v-for="item in recommendedItems.slice(1, 4)"
          :key="item.key"
          :to="item.to"
          class="market-feature-card"
        >
          <div class="market-feature-card__media">
            <img
              v-if="item.imageUrl"
              :src="item.imageUrl"
              :alt="item.imageAlt"
            />
            <AppIcon
              v-else
              name="picture"
              :size="28"
            />
          </div>
          <div class="market-feature-card__body">
            <h2>{{ item.title }}</h2>
            <strong>{{ item.price }}</strong>
          </div>
        </RouterLink>
      </div>
    </section>

    <section
      v-for="row in categoryRows"
      :key="row.key"
      class="market-category"
    >
      <div class="market-category__header">
        <div>
          <h2>{{ row.label }}</h2>
          <p>{{ t('marketplace.discover.itemCount', { count: row.count }) }}</p>
        </div>
        <RouterLink
          :to="row.filterTo"
          class="market-category__more"
        >
          {{ t('marketplace.discover.more') }}
          <AppIcon
            name="arrow-right"
            :size="15"
          />
        </RouterLink>
      </div>

      <div class="market-category__content">
        <div class="market-category__carousel">
          <img
            v-if="getActiveCategorySlide(row).imageUrl"
            :src="getActiveCategorySlide(row).imageUrl"
            :alt="getActiveCategorySlide(row).imageAlt"
          />
          <AppIcon
            v-else
            name="picture"
            :size="36"
          />
          <div class="market-category__carousel-copy">
            <h3>{{ getActiveCategorySlide(row).title }}</h3>
            <strong>{{ getActiveCategorySlide(row).price }}</strong>
          </div>
          <div class="market-category__carousel-actions">
            <button
              type="button"
              :aria-label="t('marketplace.discover.previousSlide')"
              @click="showPreviousCategorySlide(row)"
            >
              <AppIcon
                name="arrow-left"
                :size="15"
              />
            </button>
            <button
              type="button"
              :aria-label="t('marketplace.discover.nextSlide')"
              @click="showNextCategorySlide(row)"
            >
              <AppIcon
                name="arrow-right"
                :size="15"
              />
            </button>
          </div>
        </div>

        <div class="market-category__grid">
          <RouterLink
            v-for="listing in row.listings"
            :key="listing.key"
            :to="listing.to"
            class="market-item-card"
          >
            <div class="market-item-card__media">
              <img
                v-if="listing.imageUrl"
                :src="listing.imageUrl"
                :alt="listing.imageAlt"
              />
              <AppIcon
                v-else
                name="picture"
                :size="28"
              />
            </div>
            <div class="market-item-card__body">
              <h3>{{ listing.title }}</h3>
              <p>{{ listing.summary }}</p>
              <strong>{{ listing.price }}</strong>
            </div>
          </RouterLink>
        </div>
      </div>
    </section>
  </main>
</template>

<style scoped>
.market-discover {
  width: min(100%, var(--layout-page-max-width));
  margin: 0 auto;
  padding: 12px var(--layout-page-padding-inline) 72px;
}

.market-discover__hero {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 24px;
  border-bottom: 1px solid rgb(var(--color-border));
  padding: 16px 0 22px;
}

.market-discover__eyebrow {
  color: rgb(var(--color-primary));
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0.24em;
  text-transform: uppercase;
}

.market-discover__hero h1 {
  margin: 8px 0;
  color: rgb(var(--color-text));
  font-family: var(--font-display);
  font-size: 42px;
  font-weight: 400;
}

.market-discover__hero p {
  max-width: 520px;
  color: rgb(var(--color-text-muted));
  font-size: 13px;
  line-height: 1.8;
}

.market-discover__filter-link,
.market-category__more {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  border: 1px solid rgb(var(--color-border));
  border-radius: 2px;
  color: rgb(var(--color-text));
  font-size: 12px;
  font-weight: 600;
  padding: 9px 14px;
  white-space: nowrap;
}

.market-discover__empty {
  border: 1px solid rgb(var(--color-border));
  border-radius: 3px;
  color: rgb(var(--color-text-muted));
  font-size: 12px;
  margin-top: 14px;
  padding: 24px;
  text-align: center;
}

.market-discover__feature {
  display: grid;
  grid-template-columns: minmax(0, 1.4fr) minmax(260px, 0.6fr);
  gap: 10px;
  margin-top: 18px;
}

.market-discover__side-list {
  display: grid;
  gap: 10px;
}

.market-feature-card,
.market-item-card {
  overflow: hidden;
  border: 1px solid rgb(var(--color-border));
  border-radius: 3px;
  background: rgb(var(--color-surface));
  transition:
    border-color 0.18s ease,
    box-shadow 0.18s ease;
}

.market-feature-card:hover,
.market-item-card:hover {
  border-color: rgb(var(--color-primary));
  box-shadow: var(--shadow-soft);
}

.market-feature-card {
  display: grid;
  grid-template-columns: 120px minmax(0, 1fr);
}

.market-feature-card--main {
  grid-template-columns: minmax(220px, 0.82fr) minmax(0, 1fr);
}

.market-feature-card__media,
.market-item-card__media {
  display: flex;
  min-height: 116px;
  align-items: center;
  justify-content: center;
  background: rgb(var(--color-surface-muted));
  color: rgb(var(--color-text-muted));
}

.market-feature-card--main .market-feature-card__media {
  min-height: 280px;
}

.market-feature-card__media img,
.market-item-card__media img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.market-feature-card__body {
  display: flex;
  min-width: 0;
  flex-direction: column;
  justify-content: center;
  padding: 12px;
}

.market-feature-card__body span {
  color: rgb(var(--color-primary));
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0.16em;
  text-transform: uppercase;
}

.market-feature-card__body h2,
.market-item-card h3 {
  color: rgb(var(--color-text));
  font-size: 14px;
  font-weight: 600;
  line-height: 1.45;
}

.market-feature-card--main h2 {
  font-size: 20px;
}

.market-feature-card__body p,
.market-item-card p {
  display: -webkit-box;
  color: rgb(var(--color-text-muted));
  font-size: 12px;
  line-height: 1.6;
  margin-top: 5px;
  overflow: hidden;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.market-feature-card__body strong,
.market-item-card strong {
  color: rgb(var(--color-text));
  font-size: 15px;
  font-weight: 500;
  margin-top: 8px;
}

.market-category {
  margin-top: 24px;
}

.market-category__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 12px;
}

.market-category__header h2 {
  border-left: 3px solid rgb(var(--color-primary));
  color: rgb(var(--color-text));
  font-size: 15px;
  font-weight: 600;
  padding-left: 10px;
}

.market-category__header p {
  color: rgb(var(--color-text-muted));
  font-size: 11px;
  margin-top: 4px;
}

.market-category__content {
  display: grid;
  grid-template-columns: 260px minmax(0, 1fr);
  gap: 10px;
}

.market-category__carousel {
  position: relative;
  display: flex;
  min-height: 280px;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  border: 1px solid rgb(var(--color-border));
  border-radius: 3px;
  background: rgb(var(--color-surface-muted));
  color: rgb(var(--color-text-muted));
}

.market-category__carousel img {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.market-category__carousel::after {
  position: absolute;
  inset: 0;
  background: linear-gradient(180deg, transparent, rgb(0 0 0 / 0.68));
  content: '';
}

.market-category__carousel-copy {
  position: absolute;
  z-index: 1;
  right: 0;
  bottom: 0;
  left: 0;
  color: #ffffff;
  padding: 12px;
}

.market-category__carousel-copy h3 {
  font-size: 15px;
  font-weight: 600;
}

.market-category__carousel-copy strong {
  display: block;
  font-size: 18px;
  font-weight: 500;
  margin-top: 5px;
}

.market-category__carousel-actions {
  position: absolute;
  z-index: 2;
  top: 8px;
  right: 8px;
  display: flex;
  gap: 5px;
}

.market-category__carousel-actions button {
  display: inline-flex;
  width: 28px;
  height: 28px;
  align-items: center;
  justify-content: center;
  border: 1px solid rgb(255 255 255 / 0.52);
  border-radius: 2px;
  background: rgb(0 0 0 / 0.24);
  color: #ffffff;
}

.market-category__grid {
  columns: 3;
  column-gap: 10px;
}

.market-item-card {
  display: inline-block;
  width: 100%;
  break-inside: avoid;
  margin-bottom: 10px;
}

.market-item-card__media {
  min-height: 112px;
}

.market-item-card__media img {
  height: auto;
}

.market-item-card__body {
  padding: 12px;
}

@media (max-width: 1023px) {
  .market-discover {
    padding-bottom: 96px;
  }

  .market-discover__hero,
  .market-discover__feature,
  .market-feature-card,
  .market-feature-card--main,
  .market-category__content {
    grid-template-columns: 1fr;
  }

  .market-discover__hero {
    align-items: flex-start;
    flex-direction: column;
  }

  .market-category__grid {
    columns: 1;
  }
}
</style>

<!--
 * 二手交易 Discover 頁。
 * 1. 展示推薦好物首屏。
 * 2. 依主分類輸出輪播圖與橫向帖子列表。
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
  <div class="discover-page space-y-12 py-6 lg:space-y-14 lg:py-8">
    <section
      v-if="loading"
      class="section-shell"
    >
      <div class="rounded-feature border border-border bg-surface p-6 text-sm text-text-muted">
        {{ t('common.status.loading') }}
      </div>
    </section>

    <section
      v-if="primaryRecommendation"
      class="section-shell"
    >
      <div class="recommendation-stage grid gap-5 lg:grid-cols-[minmax(0,1.55fr)_minmax(22rem,0.8fr)]">
        <RouterLink
          :to="primaryRecommendation.to"
          class="recommendation-feature group relative min-h-[32rem] overflow-hidden rounded-feature"
        >
          <img
            v-if="primaryRecommendation.imageUrl"
            :src="primaryRecommendation.imageUrl"
            :alt="primaryRecommendation.imageAlt"
            class="absolute inset-0 h-full w-full object-cover transition duration-700 group-hover:scale-[1.03]"
          />
          <div
            v-else
            class="absolute inset-0 flex items-center justify-center bg-surface-raised text-text-muted"
          >
            <AppIcon
              name="picture"
              :size="44"
            />
          </div>
          <div class="absolute inset-0 bg-gradient-to-t from-black/72 via-black/18 to-transparent" />
          <div class="absolute inset-x-0 bottom-0 max-w-4xl space-y-4 p-6 text-white sm:p-8 lg:p-10">
            <p class="text-kicker text-white/72">
              {{ t('marketplace.discover.recommended') }}
            </p>
            <h1 class="max-w-3xl font-display text-4xl font-semibold leading-tight sm:text-5xl lg:text-6xl">
              {{ primaryRecommendation.title }}
            </h1>
            <p class="line-clamp-2 max-w-2xl text-base leading-7 text-white/82">
              {{ primaryRecommendation.summary }}
            </p>
            <p class="font-display text-4xl">
              {{ primaryRecommendation.price }}
            </p>
          </div>
        </RouterLink>

        <div class="recommendation-sidebar grid gap-5 sm:grid-cols-3 lg:grid-cols-1">
          <RouterLink
            v-for="item in recommendedItems.slice(1)"
            :key="item.key"
            :to="item.to"
            class="recommendation-card marketplace-post-card group relative min-h-[14rem] overflow-hidden rounded-feature bg-surface shadow-soft"
          >
            <img
              v-if="item.imageUrl"
              :src="item.imageUrl"
              :alt="item.imageAlt"
              class="absolute inset-0 h-full w-full object-cover transition duration-500 group-hover:scale-[1.04]"
            />
            <div
              v-else
              class="absolute inset-0 flex items-center justify-center bg-surface-raised text-text-muted"
            >
              <AppIcon
                name="picture"
                :size="34"
              />
            </div>
            <div class="marketplace-post-card__shade" />
            <div class="marketplace-post-card__copy">
              <h2 class="line-clamp-2 text-lg font-semibold leading-snug text-white">
                {{ item.title }}
              </h2>
              <p class="line-clamp-2 text-sm leading-6 text-white/78">
                {{ item.summary }}
              </p>
              <p class="font-display text-2xl text-white">
                {{ item.price }}
              </p>
            </div>
          </RouterLink>
        </div>
      </div>
    </section>

    <section class="section-shell space-y-8">
      <article
        v-for="row in categoryRows"
        :key="row.key"
        class="category-row"
      >
        <div class="mb-4 flex items-center justify-between gap-4">
          <div>
            <h2 class="font-display text-3xl font-semibold text-text sm:text-4xl">
              {{ row.label }}
            </h2>
            <p class="mt-1 text-sm text-text-muted">
              {{ t('marketplace.discover.itemCount', { count: row.count }) }}
            </p>
          </div>
          <RouterLink
            :to="row.filterTo"
            class="more-link inline-flex items-center gap-2 rounded-pill px-4 py-2 text-sm font-semibold"
          >
            {{ t('marketplace.discover.more') }}
            <AppIcon
              name="arrow-right"
              :size="16"
            />
          </RouterLink>
        </div>

        <div class="grid gap-5 lg:grid-cols-[21rem_minmax(0,1fr)] xl:grid-cols-[24rem_minmax(0,1fr)]">
          <div class="category-carousel relative overflow-hidden rounded-feature">
            <img
              v-if="getActiveCategorySlide(row).imageUrl"
              :src="getActiveCategorySlide(row).imageUrl"
              :alt="getActiveCategorySlide(row).imageAlt"
              class="h-full min-h-[23rem] w-full object-cover"
            />
            <div
              v-else
              class="flex h-full min-h-[23rem] w-full items-center justify-center bg-surface-raised text-text-muted"
            >
              <AppIcon
                name="picture"
                :size="44"
              />
            </div>
            <div class="absolute inset-0 bg-gradient-to-t from-black/68 via-black/10 to-transparent" />
            <div class="absolute inset-x-0 bottom-0 space-y-2 p-5 text-white">
              <h3 class="line-clamp-2 text-2xl font-semibold">
                {{ getActiveCategorySlide(row).title }}
              </h3>
              <p class="line-clamp-2 text-sm leading-6 text-white/78">
                {{ getActiveCategorySlide(row).summary }}
              </p>
              <p class="font-display text-3xl">
                {{ getActiveCategorySlide(row).price }}
              </p>
            </div>
            <div class="absolute right-4 top-4 flex gap-2">
              <button
                type="button"
                class="carousel-button"
                :aria-label="t('marketplace.discover.previousSlide')"
                @click="showPreviousCategorySlide(row)"
              >
                <AppIcon
                  name="arrow-left"
                  :size="16"
                />
              </button>
              <button
                type="button"
                class="carousel-button"
                :aria-label="t('marketplace.discover.nextSlide')"
                @click="showNextCategorySlide(row)"
              >
                <AppIcon
                  name="arrow-right"
                  :size="16"
                />
              </button>
            </div>
          </div>

          <div class="listing-lane flex gap-4 overflow-x-auto pb-3">
            <RouterLink
              v-for="listing in row.listings"
              :key="listing.key"
              :to="listing.to"
              class="listing-card marketplace-post-card group relative h-[20rem] w-[15.5rem] shrink-0 overflow-hidden rounded-feature bg-surface shadow-soft"
            >
              <img
                v-if="listing.imageUrl"
                :src="listing.imageUrl"
                :alt="listing.imageAlt"
                class="absolute inset-0 h-full w-full object-cover transition duration-500 group-hover:scale-[1.04]"
              />
              <div
                v-else
                class="absolute inset-0 flex items-center justify-center bg-surface-raised text-text-muted"
              >
                <AppIcon
                  name="picture"
                  :size="34"
                />
              </div>
              <div class="marketplace-post-card__shade" />
              <div class="marketplace-post-card__copy">
                <h3 class="line-clamp-2 text-lg font-semibold leading-snug text-white">
                  {{ listing.title }}
                </h3>
                <p class="line-clamp-2 text-sm leading-6 text-white/78">
                  {{ listing.summary }}
                </p>
                <p class="font-display text-2xl text-white">
                  {{ listing.price }}
                </p>
              </div>
            </RouterLink>
          </div>
        </div>
      </article>
    </section>
  </div>
</template>

<style scoped>
.recommendation-stage {
  min-height: 0;
}

.recommendation-feature,
.recommendation-card,
.category-carousel,
.listing-card {
  border: 1px solid rgb(var(--color-border) / 0.58);
}

.recommendation-card,
.listing-card {
  transition:
    transform 0.24s ease,
    box-shadow 0.24s ease;
}

.recommendation-card:hover,
.listing-card:hover {
  transform: translateY(-3px);
  box-shadow: var(--shadow-floating);
}

.category-row {
  border-top: 1px solid rgb(var(--color-border) / 0.64);
  padding-top: 2rem;
}

.more-link {
  border: 1px solid rgb(var(--color-primary) / 0.22);
  background: rgb(var(--color-primary) / 0.08);
  color: rgb(var(--color-primary));
}

.category-carousel {
  min-height: 23rem;
  background: rgb(var(--color-surface-raised));
}

.marketplace-post-card__shade {
  position: absolute;
  inset: 0;
  background: linear-gradient(180deg, rgb(0 0 0 / 0.08), rgb(0 0 0 / 0.28) 45%, rgb(0 0 0 / 0.72));
}

.marketplace-post-card__copy {
  position: absolute;
  inset-inline: 0;
  bottom: 0;
  display: flex;
  min-height: 10.75rem;
  flex-direction: column;
  gap: 0.7rem;
  justify-content: flex-end;
  padding: 1rem;
}

.carousel-button {
  display: inline-flex;
  width: 2.35rem;
  height: 2.35rem;
  align-items: center;
  justify-content: center;
  border: 1px solid rgb(255 255 255 / 0.42);
  border-radius: 999px;
  background: rgb(255 255 255 / 0.18);
  color: white;
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
}

.listing-lane {
  scrollbar-width: thin;
  scrollbar-color: rgb(var(--color-border)) transparent;
}

.listing-lane::-webkit-scrollbar {
  height: 0.45rem;
}

.listing-lane::-webkit-scrollbar-thumb {
  border-radius: 999px;
  background: rgb(var(--color-border));
}

@media (max-width: 1023px) {
  .recommendation-stage {
    height: auto;
    min-height: auto;
  }
}

@media (min-width: 1024px) {
  .recommendation-stage {
    height: clamp(34rem, calc(100svh - var(--app-header-offset, 0rem) - 4rem), 42rem);
  }

  .recommendation-feature,
  .recommendation-sidebar,
  .recommendation-card {
    min-height: 0;
  }

  .recommendation-sidebar {
    grid-template-rows: repeat(3, minmax(0, 1fr));
  }

  .recommendation-sidebar img {
    min-height: 0;
  }
}
</style>

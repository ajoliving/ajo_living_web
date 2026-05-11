<!--
 * 我的帖子頁。
 * 1. 顯示真實我的帖子列表、搜尋篩選與關鍵操作。
 * 2. 支援草稿發布、過期重發、售出與下架流程。
-->
<script setup lang="ts">
import ListingStatusBadge from '@/shared/components/marketplace/ListingStatusBadge.vue';
import AppIcon from '@/shared/components/base/AppIcon.vue';
import BaseEmpty from '@/shared/components/feedback/BaseEmpty.vue';

import { useMarketplaceMyListingsPage } from './my-listings';

const {
  activeTab,
  filteredItems,
  formatDate,
  formatPrice,
  formatAjoPoints,
  loading,
  openCreate,
  openEditor,
  openManagedDetail,
  openPublicDetail,
  preferenceStore,
  resolveListingCategoryLabel,
  resolveListingCommunityName,
  resolveListingCoverImage,
  resolveListingPrice,
  resolveListingPublishedAt,
  resolveListingStatus,
  resolveListingSummary,
  resolveListingTitle,
  resolveListingVisibility,
  runAction,
  searchQuery,
  secondhandChargeCost,
  setActiveTab,
  t,
  tabOptions,
} = useMarketplaceMyListingsPage();
</script>

<template>
  <main class="marketplace-my-page">
    <aside class="marketplace-my-sidebar">
      <section class="space-y-4">
        <h2 class="my-section-title">
          {{ t('marketplace.mine.search') }}
        </h2>
        <label class="my-search-field">
          <AppIcon
            name="search"
            class="left-icon"
            :size="17"
          />
          <input
            v-model="searchQuery"
            type="search"
            class="my-text-input pl-9"
            :placeholder="t('marketplace.mine.searchPlaceholder')"
          />
        </label>
      </section>

      <section class="space-y-4 border-t border-border pt-5">
        <h2 class="my-section-title">
          {{ t('marketplace.mine.statusFilter') }}
        </h2>
        <div class="flex flex-wrap gap-2">
          <button
            v-for="option in tabOptions"
            :key="option.value"
            type="button"
            class="my-filter-chip"
            :class="activeTab === option.value ? 'my-filter-chip-active' : 'my-filter-chip-idle'"
            @click="setActiveTab(option.value)"
          >
            {{ option.label }}
          </button>
        </div>
      </section>
    </aside>

    <section class="marketplace-my-results min-w-0 flex-1">
      <div class="mb-8 grid gap-4 md:grid-cols-[minmax(0,1fr)_auto] md:items-center">
        <p class="min-w-0 text-[18px] leading-[1.6] text-text-muted">
          {{ t('marketplace.filter.resultsFound', { count: filteredItems.length }) }}
        </p>
        <button
          type="button"
          class="my-create-button"
          @click="openCreate"
        >
          <AppIcon
            name="plus-square"
            :size="17"
          />
          <span>{{ t('marketplace.editor.createTitle') }}</span>
        </button>
      </div>

      <div
        v-if="loading"
        class="my-loading-card"
      >
        {{ t('marketplace.mine.loading') }}
      </div>

      <div
        v-else-if="filteredItems.length > 0"
        class="my-listing-stack"
      >
        <article
          v-for="listing in filteredItems"
          :key="listing.listing_id"
          class="my-listing-card my-listing-card--with-actions group"
        >
          <div class="my-listing-card__media">
            <img
              v-if="resolveListingCoverImage(listing)"
              :src="resolveListingCoverImage(listing)?.url"
              :alt="resolveListingTitle(listing)"
              class="h-full w-full object-cover transition-transform duration-500 group-hover:scale-105"
            />
            <div
              v-else
              class="flex h-full w-full flex-col items-center justify-center gap-3 bg-border text-sm text-text-muted"
            >
              <AppIcon
                name="picture"
                :size="44"
              />
              {{ t('marketplace.mine.noCover') }}
            </div>
          </div>

          <div class="my-listing-card__body">
            <div class="my-listing-card__topline">
              <ListingStatusBadge
                :status="resolveListingStatus(listing)"
                :visibility="resolveListingVisibility(listing)"
              />
              <p class="my-listing-card__price">
                {{ formatPrice(resolveListingPrice(listing), preferenceStore.locale) }}
              </p>
            </div>

            <div>
              <div class="mb-1.5 flex items-start justify-between gap-3">
                <h2 class="my-listing-card__title">
                  {{ resolveListingTitle(listing) }}
                </h2>
              </div>
              <p class="my-listing-card__summary">
                {{ resolveListingSummary(listing) }}
              </p>
            </div>

            <div class="my-listing-meta-grid">
              <div>
                <p class="my-meta-label">{{ t('marketplace.mine.category') }}</p>
                <p class="my-meta-value">
                  {{ resolveListingCategoryLabel(listing, preferenceStore.locale) }}
                </p>
              </div>
              <div>
                <p class="my-meta-label">{{ t('marketplace.mine.published') }}</p>
                <p class="my-meta-value">
                  {{ formatDate(resolveListingPublishedAt(listing), preferenceStore.locale) }}
                </p>
              </div>
              <div>
                <p class="my-meta-label">{{ t('marketplace.mine.community') }}</p>
                <p class="my-meta-value">
                  {{ resolveListingCommunityName(listing) }}
                </p>
              </div>
            </div>
          </div>

          <div
            class="my-listing-actions"
          >
            <button
              type="button"
              class="my-action-button my-action-button-primary"
              @click="openEditor(listing.listing_id)"
            >
              <AppIcon
                name="palette"
                :size="16"
              />
              <span>{{ t('marketplace.editor.editTitle') }}</span>
            </button>

            <button
              type="button"
              class="my-action-button my-action-button-secondary"
              @click="openManagedDetail(listing.listing_id)"
            >
              <AppIcon
                name="view"
                :size="16"
              />
              <span>{{ t('marketplace.mine.manageDetailAction') }}</span>
            </button>

            <button
              type="button"
              class="my-action-button my-action-button-secondary"
              @click="openPublicDetail(listing.listing_id)"
            >
              <AppIcon
                name="picture"
                :size="16"
              />
              <span>{{ t('marketplace.mine.publicDetailAction') }}</span>
            </button>

            <button
              v-if="listing.publication_status === 'draft'"
              type="button"
              class="my-action-button my-action-button-primary"
              @click="runAction('publish', listing.listing_id)"
            >
              <AppIcon
                name="plus-square"
                :size="16"
              />
              <span>
                {{ t('marketplace.mine.publishAction') }} · {{ formatAjoPoints(secondhandChargeCost) }}
              </span>
            </button>

            <button
              v-if="listing.publication_status === 'expired'"
              type="button"
              class="my-action-button my-action-button-primary"
              @click="runAction('republish', listing.listing_id)"
            >
              <AppIcon
                name="reload"
                :size="16"
              />
              <span>
                {{ t('marketplace.mine.republishAction') }} · {{ formatAjoPoints(secondhandChargeCost) }}
              </span>
            </button>

            <button
              v-if="listing.publication_status === 'active' && listing.business_status !== 'sold'"
              type="button"
              class="my-action-button my-action-button-secondary"
              @click="runAction('mark-sold', listing.listing_id)"
            >
              <AppIcon
                name="check-circle"
                :size="16"
              />
              <span>
                {{ t('marketplace.mine.markSoldAction') }}
              </span>
            </button>

            <button
              v-if="listing.publication_status === 'active' && listing.business_status !== 'sold'"
              type="button"
              class="my-action-button my-action-button-secondary"
              @click="runAction('deactivate', listing.listing_id)"
            >
              <AppIcon
                name="close"
                :size="16"
              />
              <span>
                {{ t('marketplace.mine.deactivateAction') }}
              </span>
            </button>
          </div>
        </article>
      </div>

      <BaseEmpty
        v-else
        :title="t('common.empty.myListingsTitle')"
        :description="t('common.empty.myListingsDescription')"
      >
        <button
          type="button"
          class="my-create-button"
          @click="openCreate"
        >
          <AppIcon
            name="plus-square"
            :size="17"
          />
          <span>{{ t('marketplace.editor.createTitle') }}</span>
        </button>
      </BaseEmpty>
    </section>
  </main>
</template>

<style scoped>
.marketplace-my-page {
  display: grid;
  width: 100%;
  gap: 1.5rem;
  margin: 0;
  padding: 0;
  color: rgb(var(--color-text));
  overflow: visible;
}

.marketplace-my-sidebar {
  display: grid;
  width: 100%;
  height: auto;
  gap: 1.25rem;
  overflow: visible;
  padding: 0;
}

.marketplace-my-sidebar::-webkit-scrollbar {
  display: none;
}

.marketplace-my-results {
  height: auto;
  overflow: visible;
  padding-right: 0;
}

.my-section-title {
  color: rgb(var(--color-text-muted));
  font-size: 0.75rem;
  font-weight: 600;
  letter-spacing: 0.1em;
  line-height: 1;
  text-transform: uppercase;
}

.my-search-field {
  position: relative;
  display: block;
}

.my-search-field .left-icon {
  position: absolute;
  top: 50%;
  left: 0.625rem;
  z-index: 1;
  transform: translateY(-50%);
  color: rgb(var(--color-text-muted));
  pointer-events: none;
}

.my-text-input {
  height: 3rem;
  width: 100%;
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.5rem;
  background: rgb(var(--color-surface));
  padding-top: 0.5rem;
  padding-bottom: 0.5rem;
  color: rgb(var(--color-text));
  font-size: 0.9375rem;
  line-height: 1.6;
  outline: none;
  transition:
    border-color 0.2s ease,
    box-shadow 0.2s ease;
}

.my-text-input:focus {
  border-color: rgb(var(--color-primary));
  box-shadow: 0 0 0 3px rgb(0 39 39 / 0.16);
}

.my-filter-chip {
  border-radius: 9999px;
  border-width: 1px;
  min-height: 2.5rem;
  padding: 0.55rem 1rem;
  font-size: 0.875rem;
  font-weight: 700;
  letter-spacing: 0.02em;
  line-height: 1;
  transition:
    border-color 0.2s ease,
    background-color 0.2s ease,
    color 0.2s ease;
}

.my-filter-chip-active {
  border-color: rgb(var(--color-primary));
  background: rgb(0 39 39 / 0.05);
  color: rgb(var(--color-primary));
}

.my-filter-chip-active:hover {
  border-color: rgb(var(--color-primary));
  background: rgb(0 39 39 / 0.05);
  color: rgb(var(--color-primary));
}

.my-filter-chip-idle {
  border-color: rgb(var(--color-border));
  color: rgb(var(--color-text-muted));
}

.my-filter-chip-idle:hover {
  border-color: rgb(var(--color-text-muted));
  color: rgb(var(--color-text));
}

.my-loading-card {
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.75rem;
  background: rgb(var(--color-surface));
  padding: 1.5rem;
  color: rgb(var(--color-text-muted));
  font-size: 0.875rem;
  box-shadow: 0 10px 30px -5px rgb(0 39 39 / 0.05);
}

.my-listing-stack {
  display: grid;
  gap: 1.25rem;
}

.my-create-button {
  display: inline-flex;
  min-height: 3.25rem;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  justify-self: start;
  border: 1px solid rgb(var(--color-primary));
  border-radius: 9999px;
  background: rgb(var(--color-primary));
  padding: 0.65rem 1.1rem;
  color: rgb(var(--color-primary-contrast));
  font-size: 0.875rem;
  font-weight: 800;
  letter-spacing: 0.02em;
  line-height: 1.1;
  transition:
    background-color 0.2s ease,
    border-color 0.2s ease;
}

.my-create-button:hover {
  border-color: rgb(var(--color-text));
  background: rgb(var(--color-text));
}

.my-listing-card {
  display: grid;
  min-height: 13rem;
  overflow: hidden;
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.75rem;
  background: rgb(var(--color-surface));
  box-shadow: 0 10px 30px -5px rgb(0 39 39 / 0.05);
  transition: box-shadow 0.3s ease;
}

.my-listing-card:hover {
  box-shadow: 0 8px 32px rgb(0 0 0 / 0.06);
}

.my-listing-card__media {
  height: 10.875rem;
  overflow: hidden;
  background: rgb(var(--color-surface-muted));
}

.my-listing-card__body {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 0.65rem;
  overflow: hidden;
  padding: 0.875rem 1rem;
}

.my-listing-card__topline {
  display: flex;
  min-width: 0;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
}

.my-listing-card__price {
  flex-shrink: 0;
  color: rgb(var(--color-primary));
  font-family: var(--font-display);
  font-size: 1rem;
  font-weight: 700;
  line-height: 1.2;
}

.my-listing-card__title {
  min-width: 0;
  display: -webkit-box;
  overflow: hidden;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 1;
  color: rgb(var(--color-text));
  font-family: var(--font-display);
  font-size: 1.125rem;
  font-weight: 500;
  line-height: 1.28;
}

.my-listing-card__summary {
  display: -webkit-box;
  overflow: hidden;
  margin: 0;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 1;
  color: rgb(var(--color-text-muted));
  font-size: 0.875rem;
  line-height: 1.45;
}

.my-listing-meta-grid {
  display: grid;
  gap: 0.65rem;
  border-top: 1px solid rgb(var(--color-border));
  padding-top: 0.65rem;
}

.my-meta-label {
  color: rgb(var(--color-text-muted));
  font-size: 0.75rem;
  font-weight: 600;
  letter-spacing: 0.1em;
  line-height: 1;
  text-transform: uppercase;
}

.my-meta-value {
  margin-top: 0.35rem;
  color: rgb(var(--color-text));
  font-size: 0.8125rem;
  font-weight: 600;
  line-height: 1.25;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.my-listing-actions {
  display: grid;
  align-content: center;
  gap: 0.35rem;
  overflow: hidden;
  border-top: 1px solid rgb(var(--color-border));
  padding: 0.75rem 1rem;
}

.my-action-button {
  display: inline-flex;
  min-height: 1.85rem;
  align-items: center;
  justify-content: center;
  gap: 0.35rem;
  border: 1px solid rgb(var(--color-border));
  border-radius: 9999px;
  padding: 0.35rem 0.5rem;
  font-size: 0.6875rem;
  font-weight: 800;
  letter-spacing: 0.02em;
  line-height: 1.1;
  transition:
    border-color 0.2s ease,
    background-color 0.2s ease,
    color 0.2s ease;
}

.my-action-button-primary {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary));
  color: rgb(var(--color-primary-contrast));
}

.my-action-button-primary:hover {
  background: rgb(var(--color-text));
  border-color: rgb(var(--color-text));
}

.my-action-button-secondary {
  background: transparent;
  color: rgb(var(--color-text-muted));
}

.my-action-button-secondary:hover {
  border-color: rgb(var(--color-primary));
  color: rgb(var(--color-primary));
}

@media (max-width: 767px) {
  .marketplace-my-page {
    gap: 2rem;
  }
}

@media (min-width: 640px) {
  .my-listing-meta-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .my-create-button {
    justify-self: end;
  }
}

@media (min-width: 1024px) {
  .my-listing-card {
    height: 10.875rem;
    min-height: 0;
    grid-template-columns: 13.5rem minmax(0, 1fr);
  }

  .my-listing-card--with-actions {
    grid-template-columns: 13.5rem minmax(0, 1fr) 13rem;
  }

  .my-listing-card__media {
    height: 100%;
  }

  .my-listing-actions {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    border-top: 0;
    border-left: 1px solid rgb(var(--color-border));
  }
}
</style>

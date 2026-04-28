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

      <section class="space-y-4 border-t border-[#E2E3E1] pt-5">
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
        <p class="min-w-0 text-[18px] leading-[1.6] text-[#414848]">
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
              class="flex h-full w-full flex-col items-center justify-center gap-3 bg-[#E2E3E1] text-sm text-[#717878]"
            >
              <AppIcon
                name="picture"
                :size="44"
              />
              {{ t('marketplace.mine.noCover') }}
            </div>
          </div>

          <div class="my-listing-card__body">
            <div class="flex flex-wrap items-center justify-between gap-3">
              <ListingStatusBadge
                :status="resolveListingStatus(listing)"
                :visibility="resolveListingVisibility(listing)"
              />
              <p class="font-display text-[20px] font-semibold text-[#002727]">
                {{ formatPrice(resolveListingPrice(listing), preferenceStore.locale) }}
              </p>
            </div>

            <div>
              <div class="mb-2 flex items-start justify-between gap-4">
                <h2 class="min-w-0 line-clamp-1 font-display text-[24px] font-medium leading-[1.4] text-[#1A1C1B]">
                  {{ resolveListingTitle(listing) }}
                </h2>
              </div>
              <p class="line-clamp-2 text-[16px] leading-[1.6] text-[#414848]">
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
                {{ t('marketplace.mine.publishAction') }}
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
                {{ t('marketplace.mine.republishAction') }}
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
  color: #1a1c1b;
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
  color: #717878;
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
  color: #717878;
  pointer-events: none;
}

.my-text-input {
  height: 3rem;
  width: 100%;
  border: 1px solid #e2e3e1;
  border-radius: 0.5rem;
  background: #ffffff;
  padding-top: 0.5rem;
  padding-bottom: 0.5rem;
  color: #1a1c1b;
  font-size: 0.9375rem;
  line-height: 1.6;
  outline: none;
  transition:
    border-color 0.2s ease,
    box-shadow 0.2s ease;
}

.my-text-input:focus {
  border-color: #002727;
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
  border-color: #002727;
  background: rgb(0 39 39 / 0.05);
  color: #002727;
}

.my-filter-chip-active:hover {
  border-color: #002727;
  background: rgb(0 39 39 / 0.05);
  color: #002727;
}

.my-filter-chip-idle {
  border-color: #c1c8c7;
  color: #414848;
}

.my-filter-chip-idle:hover {
  border-color: #717878;
  color: #1a1c1b;
}

.my-loading-card {
  border: 1px solid #e2e3e1;
  border-radius: 0.75rem;
  background: #ffffff;
  padding: 1.5rem;
  color: #717878;
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
  border: 1px solid #002727;
  border-radius: 9999px;
  background: #002727;
  padding: 0.65rem 1.1rem;
  color: #ffffff;
  font-size: 0.875rem;
  font-weight: 800;
  letter-spacing: 0.02em;
  line-height: 1.1;
  transition:
    background-color 0.2s ease,
    border-color 0.2s ease;
}

.my-create-button:hover {
  border-color: #1a1c1b;
  background: #1a1c1b;
}

.my-listing-card {
  display: grid;
  min-height: 16rem;
  overflow: hidden;
  border: 1px solid #e2e3e1;
  border-radius: 0.75rem;
  background: #ffffff;
  box-shadow: 0 10px 30px -5px rgb(0 39 39 / 0.05);
  transition: box-shadow 0.3s ease;
}

.my-listing-card:hover {
  box-shadow: 0 8px 32px rgb(0 0 0 / 0.06);
}

.my-listing-card__media {
  height: 13rem;
  overflow: hidden;
  background: #f4f4f2;
}

.my-listing-card__body {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 1rem;
  overflow: hidden;
  padding: 1.5rem;
}

.my-listing-meta-grid {
  display: grid;
  gap: 1rem;
  border-top: 1px solid #e2e3e1;
  padding-top: 1rem;
}

.my-meta-label {
  color: #717878;
  font-size: 0.75rem;
  font-weight: 600;
  letter-spacing: 0.1em;
  line-height: 1;
  text-transform: uppercase;
}

.my-meta-value {
  margin-top: 0.55rem;
  color: #1a1c1b;
  font-size: 0.875rem;
  font-weight: 600;
  line-height: 1.5;
}

.my-listing-actions {
  display: grid;
  align-content: center;
  gap: 0.5rem;
  overflow: hidden;
  border-top: 1px solid #e2e3e1;
  padding: 1.25rem 1.5rem 1.5rem;
}

.my-action-button {
  display: inline-flex;
  min-height: 2.35rem;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  border: 1px solid #c1c8c7;
  border-radius: 9999px;
  padding: 0.6rem 1rem;
  font-size: 0.8125rem;
  font-weight: 800;
  letter-spacing: 0.02em;
  line-height: 1.1;
  transition:
    border-color 0.2s ease,
    background-color 0.2s ease,
    color 0.2s ease;
}

.my-action-button-primary {
  border-color: #002727;
  background: #002727;
  color: #ffffff;
}

.my-action-button-primary:hover {
  background: #1a1c1b;
  border-color: #1a1c1b;
}

.my-action-button-secondary {
  background: transparent;
  color: #414848;
}

.my-action-button-secondary:hover {
  border-color: #002727;
  color: #002727;
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
    height: 16rem;
    min-height: 0;
    grid-template-columns: 17rem minmax(0, 1fr);
  }

  .my-listing-card--with-actions {
    grid-template-columns: 17rem minmax(0, 1fr) 13rem;
  }

  .my-listing-card__media {
    height: 100%;
  }

  .my-listing-actions {
    border-top: 0;
    border-left: 1px solid #e2e3e1;
    justify-content: center;
    padding: 1.5rem;
  }
}
</style>

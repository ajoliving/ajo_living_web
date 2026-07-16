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
  handleSearch,
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
  secondhandRenewChargeCost,
  setActiveTab,
  t,
  tabOptions,
} = useMarketplaceMyListingsPage();
</script>

<template>
  <main class="marketplace-my-page">
    <section class="marketplace-my-results min-w-0">
      <header class="my-page-heading">
        <div>
          <p class="my-page-kicker">{{ t('marketplace.mine.kicker') }}</p>
          <h1>{{ t('marketplace.mine.title') }}</h1>
          <p>{{ t('marketplace.mine.subtitle') }}</p>
        </div>
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
      </header>

      <section class="my-toolbar">
        <form
          class="my-search-row"
          @submit.prevent="handleSearch"
        >
          <label class="my-search-field">
            <AppIcon
              name="search"
              class="left-icon member-search-input-icon"
              :size="17"
            />
            <input
              v-model="searchQuery"
              type="search"
              class="my-text-input pl-9"
              :placeholder="t('marketplace.mine.searchPlaceholder')"
            />
          </label>
          <button
            type="submit"
            class="my-search-button"
            :disabled="loading"
          >
            <AppIcon
              name="search"
              :size="16"
            />
            {{ t('marketplace.mine.search') }}
          </button>
        </form>
        <div class="my-filter-row">
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

      <div
        v-if="loading"
        class="my-loading-card"
      >
        {{ t('marketplace.mine.loading') }}
      </div>

      <div
        v-else-if="filteredItems.length > 0"
        class="my-listing-panel"
      >
        <div class="my-listing-panel__title">
          <h2>{{ t('marketplace.mine.overviewTitle') }}</h2>
        </div>

        <div class="my-table-wrap">
          <table class="my-table">
            <thead>
              <tr>
                <th>{{ t('marketplace.mine.item') }}</th>
                <th>{{ t('marketplace.mine.category') }}</th>
                <th>{{ t('marketplace.mine.price') }}</th>
                <th>{{ t('common.label.status') }}</th>
                <th>{{ t('marketplace.mine.community') }}</th>
                <th>{{ t('marketplace.mine.published') }}</th>
                <th>{{ t('marketplace.mine.actions') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="listing in filteredItems"
                :key="listing.listing_id"
              >
                <td>
                  <div class="my-item">
                    <div class="my-item__media">
                      <img
                        v-if="resolveListingCoverImage(listing)"
                        :src="resolveListingCoverImage(listing)?.url"
                        :alt="resolveListingTitle(listing)"
                      />
                      <div
                        v-else
                        class="my-item__placeholder"
                      >
                        <AppIcon
                          name="picture"
                          :size="20"
                        />
                      </div>
                    </div>

                    <div class="my-item__content">
                      <strong>{{ resolveListingTitle(listing) }}</strong>
                      <p>{{ resolveListingSummary(listing) }}</p>
                    </div>
                  </div>
                </td>
                <td>{{ resolveListingCategoryLabel(listing, preferenceStore.locale) }}</td>
                <td>{{ formatPrice(resolveListingPrice(listing), preferenceStore.locale) }}</td>
                <td>
                  <ListingStatusBadge
                    :status="resolveListingStatus(listing)"
                    :visibility="resolveListingVisibility(listing)"
                  />
                </td>
                <td>{{ resolveListingCommunityName(listing, preferenceStore.locale) }}</td>
                <td>{{ formatDate(resolveListingPublishedAt(listing), preferenceStore.locale) }}</td>
                <td>
                  <div class="my-table-actions">
                    <button
                      type="button"
                      class="my-action-button my-action-button-primary"
                      @click="openEditor(listing.listing_id)"
                    >
                      {{ t('marketplace.editor.editTitle') }}
                    </button>

                    <button
                      type="button"
                      class="my-action-button my-action-button-secondary"
                      @click="openManagedDetail(listing.listing_id)"
                    >
                      {{ t('marketplace.mine.manageDetailAction') }}
                    </button>

                    <button
                      type="button"
                      class="my-action-button my-action-button-secondary"
                      @click="openPublicDetail(listing.listing_id)"
                    >
                      {{ t('marketplace.mine.publicDetailAction') }}
                    </button>

                    <button
                      v-if="listing.publication_status === 'draft'"
                      type="button"
                      class="my-action-button my-action-button-primary"
                      @click="runAction('publish', listing.listing_id)"
                    >
                      {{ t('marketplace.mine.publishAction') }} · {{ formatAjoPoints(secondhandChargeCost) }}
                    </button>

                    <button
                      v-if="listing.publication_status === 'expired'"
                      type="button"
                      class="my-action-button my-action-button-primary"
                      @click="runAction('republish', listing.listing_id)"
                    >
                      {{ t('marketplace.mine.republishAction') }} · {{ formatAjoPoints(secondhandChargeCost) }}
                    </button>

                    <button
                      v-if="listing.publication_status === 'active' && listing.business_status !== 'sold'"
                      type="button"
                      class="my-action-button my-action-button-primary"
                      @click="runAction('renew', listing.listing_id)"
                    >
                      {{ t('marketplace.mine.renewAction') }} · {{ formatAjoPoints(secondhandRenewChargeCost) }}
                    </button>

                    <button
                      v-if="listing.publication_status === 'active' && listing.business_status !== 'sold'"
                      type="button"
                      class="my-action-button my-action-button-secondary"
                      @click="runAction('mark-sold', listing.listing_id)"
                    >
                      {{ t('marketplace.mine.markSoldAction') }}
                    </button>

                    <button
                      v-if="listing.publication_status === 'active' && listing.business_status !== 'sold'"
                      type="button"
                      class="my-action-button my-action-button-secondary"
                      @click="runAction('deactivate', listing.listing_id)"
                    >
                      {{ t('marketplace.mine.deactivateAction') }}
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
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
  gap: 1rem;
  width: 100%;
  color: rgb(var(--color-text));
}

.marketplace-my-results {
  height: auto;
  overflow: visible;
  padding-right: 0;
}

.my-page-heading {
  display: flex;
  align-items: end;
  justify-content: space-between;
  gap: 16px;
  border-bottom: 1px solid rgb(var(--color-border));
  padding-bottom: 16px;
}

.my-page-kicker {
  margin: 0;
  color: rgb(var(--color-primary));
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 0.14em;
  text-transform: uppercase;
}

.my-page-heading h1 {
  margin: 6px 0 0;
  color: rgb(var(--color-text));
  font-family: var(--font-display);
  font-size: 32px;
  font-weight: 400;
  line-height: 1.15;
}

.my-page-heading p:not(.my-page-kicker) {
  margin: 8px 0 0;
  color: rgb(var(--color-text-muted));
  font-size: 13px;
  line-height: 1.6;
}

.my-toolbar {
  display: grid;
  gap: 10px;
  border: 1px solid rgb(var(--color-border));
  border-radius: 3px;
  background: rgb(var(--color-surface));
  padding: 12px;
}

.my-search-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 8px;
  min-width: 0;
}

.my-filter-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
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

.my-search-button {
  display: inline-flex;
  min-height: 34px;
  align-items: center;
  justify-content: center;
  gap: 6px;
  border: 1px solid rgb(var(--color-border));
  border-radius: 2px;
  background: rgb(var(--color-surface));
  padding: 0 14px;
  color: rgb(var(--color-text));
  cursor: pointer;
  font-size: 12px;
  font-weight: 600;
}

.my-search-button:hover:not(:disabled) {
  border-color: rgb(var(--color-primary));
  color: rgb(var(--color-primary));
}

.my-search-button:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

.my-text-input {
  height: 34px;
  width: 100%;
  border: 1px solid rgb(var(--color-border));
  border-radius: 2px;
  background: rgb(var(--color-surface));
  padding-top: 0.5rem;
  padding-bottom: 0.5rem;
  color: rgb(var(--color-text));
  font-size: 12px;
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
  border-radius: var(--radius-sm);
  border-width: 1px;
  min-height: 32px;
  padding: 0 12px;
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.02em;
  line-height: 1;
  transition:
    border-color 0.2s ease,
    background-color 0.2s ease,
    color 0.2s ease;
}

.my-filter-chip-active {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary));
  color: rgb(var(--color-primary-contrast));
}

.my-filter-chip-active:hover {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary));
  color: rgb(var(--color-primary-contrast));
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
  border-radius: 3px;
  background: rgb(var(--color-surface));
  padding: 20px 16px;
  color: rgb(var(--color-text-muted));
  font-size: 13px;
}

.my-listing-panel {
  display: grid;
  gap: 14px;
  border: 1px solid rgb(var(--color-border));
  border-radius: 3px;
  background: rgb(var(--color-surface));
  padding: 12px;
}

.my-listing-panel__title {
  border-bottom: 1px solid rgb(var(--color-border));
  padding-bottom: 14px;
}

.my-listing-panel__title h2 {
  margin: 0;
  color: rgb(var(--color-primary));
  font-size: 14px;
  font-weight: 600;
}

.my-create-button {
  display: inline-flex;
  min-height: 34px;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  justify-self: start;
  border: 1px solid rgb(var(--color-primary));
  border-radius: 2px;
  background: rgb(var(--color-primary));
  padding: 0 12px;
  color: rgb(var(--color-primary-contrast));
  font-size: 12px;
  font-weight: 600;
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

.my-table-wrap {
  overflow-x: auto;
}

.my-table {
  width: 100%;
  min-width: 1120px;
  border-collapse: collapse;
  font-size: 13px;
}

.my-table th,
.my-table td {
  border-top: 1px solid rgb(var(--color-border));
  padding: 12px 10px;
  text-align: left;
  vertical-align: middle;
}

.my-table thead th {
  border-top: 0;
  color: rgb(var(--color-text-muted));
  font-size: 12px;
  font-weight: 600;
}

.my-item {
  display: grid;
  grid-template-columns: 72px minmax(0, 1fr);
  gap: 12px;
  align-items: start;
}

.my-item__media {
  overflow: hidden;
  border: 1px solid rgb(var(--color-border));
  border-radius: 2px;
  background: rgb(var(--color-surface-muted));
  aspect-ratio: 1;
}

.my-item__media img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.my-item__placeholder {
  display: grid;
  height: 100%;
  place-items: center;
  color: rgb(var(--color-text-muted));
}

.my-item__content {
  display: grid;
  gap: 6px;
  min-width: 0;
}

.my-item__content strong {
  overflow: hidden;
  color: rgb(var(--color-text));
  font-size: 14px;
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.my-item__content p {
  display: -webkit-box;
  overflow: hidden;
  margin: 0;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  color: rgb(var(--color-text-muted));
  font-size: 12px;
  line-height: 1.6;
}

.my-table-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.my-action-button {
  display: inline-flex;
  min-height: 32px;
  align-items: center;
  justify-content: center;
  gap: 6px;
  border: 1px solid rgb(var(--color-border));
  border-radius: 2px;
  padding: 0 10px;
  font-size: 11px;
  font-weight: 600;
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
    gap: 16px;
  }
}
</style>

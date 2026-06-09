<!--
 * 二手帖子頁。
 * 1. 以管理表格方式顯示會員自己的二手帖子。
 * 2. 支援搜尋、狀態下拉篩選、新增帖子與狀態操作。
-->
<script setup lang="ts">
import { ref } from 'vue';

import AppIcon from '@/shared/components/base/AppIcon.vue';
import BaseEmpty from '@/shared/components/feedback/BaseEmpty.vue';

import MarketplaceListingEditorPage from './editor/Page.vue';
import { useMarketplaceMyListingsPage } from './my-listings';

type ListingEditorDialogInstance = InstanceType<typeof MarketplaceListingEditorPage> & {
  requestCloseEditor: () => Promise<void>;
};

const listingEditorDialog = ref<ListingEditorDialogInstance | null>(null);

const {
  activeTab,
  closeCreate,
  createDialogOpen,
  filteredItems,
  formatListingStatus,
  formatOptionalDate,
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
  resolveListingPrice,
  resolveListingSummary,
  resolveListingTitle,
  runAction,
  searchQuery,
  secondhandChargeCost,
  secondhandRenewChargeCost,
  setActiveTab,
  t,
  tabOptions,
  loadMyListings,
} = useMarketplaceMyListingsPage();

// 1. 透過編輯器自身確認流程關閉新增帖子彈窗
const requestCloseCreate = async (): Promise<void> => {
  await listingEditorDialog.value?.requestCloseEditor();
};
</script>

<template>
  <main class="marketplace-my-page">
    <header class="my-list-header">
      <div>
        <p class="my-list-kicker">
          AJO Living
        </p>
        <h1>{{ t('marketplace.mine.title') }}</h1>
        <p>{{ t('marketplace.mine.subtitle') }}</p>
      </div>
      <button
        type="button"
        class="my-list-button my-list-button--primary"
        @click="openCreate"
      >
        <AppIcon
          name="plus-square"
          :size="17"
        />
        <span>{{ t('marketplace.editor.createTitle') }}</span>
      </button>
    </header>

    <article class="my-list-panel">
      <div class="my-list-toolbar">
        <label class="my-list-search">
          <AppIcon
            name="search"
            :size="16"
          />
          <input
            v-model="searchQuery"
            type="search"
            :placeholder="t('marketplace.mine.searchPlaceholder')"
          />
        </label>
        <select
          :value="activeTab"
          class="my-list-select"
          @change="setActiveTab(($event.target as HTMLSelectElement).value)"
        >
          <option
            v-for="option in tabOptions"
            :key="option.value"
            :value="option.value"
          >
            {{ option.label }}
          </option>
        </select>
        <p class="my-list-count">
          {{ t('marketplace.filter.resultsFound', { count: filteredItems.length }) }}
        </p>
      </div>

      <div
        v-if="loading"
        class="my-list-loading"
      >
        {{ t('marketplace.mine.loading') }}
      </div>

      <div
        v-else-if="filteredItems.length > 0"
        class="my-table-wrap"
      >
        <table class="my-table">
          <thead>
            <tr>
              <th>{{ t('marketplace.management.columnTitle') }}</th>
              <th>{{ t('marketplace.mine.category') }}</th>
              <th>{{ t('marketplace.mine.price') }}</th>
              <th>{{ t('common.label.status') }}</th>
              <th>{{ t('marketplace.mine.community') }}</th>
              <th>{{ t('common.label.publishedAt') }}</th>
              <th>{{ t('common.label.expiresAt') }}</th>
              <th class="my-table-actions">{{ t('marketplace.management.columnActions') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="listing in filteredItems"
              :key="listing.listing_id"
            >
              <td>
                <strong>{{ resolveListingTitle(listing) }}</strong>
                <small>{{ resolveListingSummary(listing) }}</small>
                <small>{{ listing.listing_id }}</small>
              </td>
              <td>{{ resolveListingCategoryLabel(listing, preferenceStore.locale) }}</td>
              <td>{{ formatPrice(resolveListingPrice(listing), preferenceStore.locale) }}</td>
              <td>
                <span class="my-status-pill">{{ formatListingStatus(listing) }}</span>
              </td>
              <td>{{ resolveListingCommunityName(listing) }}</td>
              <td>{{ formatOptionalDate(listing.published_at) }}</td>
              <td>{{ formatOptionalDate(listing.expire_at) }}</td>
              <td class="my-table-actions">
                <div class="my-action-group">
                  <button
                    type="button"
                    class="my-list-action"
                    @click="openEditor(listing.listing_id)"
                  >
                    {{ t('marketplace.editor.editTitle') }}
                  </button>
                  <button
                    type="button"
                    class="my-list-action"
                    @click="openManagedDetail(listing.listing_id)"
                  >
                    {{ t('marketplace.mine.manageDetailAction') }}
                  </button>
                  <button
                    type="button"
                    class="my-list-action"
                    @click="openPublicDetail(listing.listing_id)"
                  >
                    {{ t('marketplace.mine.publicDetailAction') }}
                  </button>
                  <button
                    v-if="listing.publication_status === 'draft'"
                    type="button"
                    class="my-list-action"
                    @click="runAction('publish', listing.listing_id)"
                  >
                    {{ t('marketplace.mine.publishAction') }} · {{ formatAjoPoints(secondhandChargeCost) }}
                  </button>
                  <button
                    v-if="listing.publication_status === 'expired'"
                    type="button"
                    class="my-list-action"
                    @click="runAction('republish', listing.listing_id)"
                  >
                    {{ t('marketplace.mine.republishAction') }} · {{ formatAjoPoints(secondhandChargeCost) }}
                  </button>
                  <button
                    v-if="listing.publication_status === 'active' && listing.business_status !== 'sold'"
                    type="button"
                    class="my-list-action"
                    @click="runAction('renew', listing.listing_id)"
                  >
                    {{ t('marketplace.mine.renewAction') }} · {{ formatAjoPoints(secondhandRenewChargeCost) }}
                  </button>
                  <button
                    v-if="listing.publication_status === 'active' && listing.business_status !== 'sold'"
                    type="button"
                    class="my-list-action"
                    @click="runAction('mark-sold', listing.listing_id)"
                  >
                    {{ t('marketplace.mine.markSoldAction') }}
                  </button>
                  <button
                    v-if="listing.publication_status === 'active' && listing.business_status !== 'sold'"
                    type="button"
                    class="my-list-action"
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

      <BaseEmpty
        v-else
        :title="t('common.empty.myListingsTitle')"
        :description="t('common.empty.myListingsDescription')"
      >
        <button
          type="button"
          class="my-list-button my-list-button--primary"
          @click="openCreate"
        >
          <AppIcon
            name="plus-square"
            :size="17"
          />
          <span>{{ t('marketplace.editor.createTitle') }}</span>
        </button>
      </BaseEmpty>
    </article>

    <Teleport to="body">
      <Transition name="my-editor-dialog">
        <div
          v-if="createDialogOpen"
          class="my-editor-dialog"
          role="dialog"
          aria-modal="true"
          :aria-label="t('marketplace.editor.createTitle')"
          @click.self="requestCloseCreate"
        >
          <section class="my-editor-dialog__panel">
            <header class="my-editor-dialog__header">
              <div>
                <p>AJO Living</p>
                <h2>{{ t('marketplace.editor.createTitle') }}</h2>
              </div>
              <button
                type="button"
                class="my-editor-dialog__close"
                :aria-label="t('marketplace.myHub.closeNewListing')"
                @click="requestCloseCreate"
              >
                <AppIcon
                  name="close"
                  :size="18"
                />
              </button>
            </header>

            <div class="my-editor-dialog__body">
              <MarketplaceListingEditorPage
                ref="listingEditorDialog"
                hide-header
                modal-mode
                @close="closeCreate"
                @saved="loadMyListings"
              />
            </div>
          </section>
        </div>
      </Transition>
    </Teleport>
  </main>
</template>

<style scoped>
.marketplace-my-page {
  display: grid;
  width: 100%;
  gap: 1rem;
  margin: 0;
  padding: 0;
  color: rgb(var(--color-text));
}

.my-list-header {
  display: grid;
  gap: 1rem;
}

.my-list-kicker {
  margin: 0;
  color: rgb(var(--color-text-muted));
  font-size: 0.75rem;
  font-weight: 800;
  letter-spacing: 0.1em;
  line-height: 1;
  text-transform: uppercase;
}

.my-list-header h1 {
  margin: 0.35rem 0 0;
  color: rgb(var(--color-text));
  font-family: var(--font-display);
  font-size: clamp(1.65rem, 2.2vw, 2.35rem);
  font-weight: 650;
  line-height: 1.2;
}

.my-list-header p:not(.my-list-kicker) {
  margin: 0.45rem 0 0;
  color: rgb(var(--color-text-muted));
  font-size: 0.9rem;
  line-height: 1.6;
}

.my-list-panel {
  overflow: hidden;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface-raised));
}

.my-list-toolbar {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 0.75rem;
  border-bottom: 1px solid rgb(var(--color-border));
  padding: 1rem;
}

.my-list-search {
  display: inline-flex;
  min-height: 2.75rem;
  align-items: center;
  gap: 0.6rem;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
  padding: 0 0.8rem;
  color: rgb(var(--color-text-muted));
}

.my-list-search input,
.my-list-select {
  width: 100%;
  border: 0;
  background: transparent;
  color: rgb(var(--color-text));
  font-size: 0.95rem;
  font-weight: 650;
  line-height: 1.4;
  outline: none;
}

.my-list-select {
  min-height: 2.75rem;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
  padding: 0 0.75rem;
}

.my-list-button,
.my-list-action {
  display: inline-flex;
  min-height: 2.75rem;
  align-items: center;
  justify-content: center;
  gap: 0.45rem;
  border: 1px solid rgb(var(--color-primary));
  border-radius: 8px;
  background: rgb(var(--color-primary));
  padding: 0 0.95rem;
  color: rgb(var(--color-primary-contrast));
  font-size: 0.9rem;
  font-weight: 750;
  line-height: 1;
  white-space: nowrap;
  transition:
    border-color 0.2s ease,
    background 0.2s ease;
}

.my-list-button:hover,
.my-list-action:hover {
  border-color: rgb(var(--color-text));
  background: rgb(var(--color-text));
}

.my-list-action {
  min-height: 2.25rem;
  border-color: rgb(var(--color-border));
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text));
  font-size: 0.82rem;
}

.my-list-action:hover {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-surface-muted));
}

.my-list-count {
  margin: 0;
  align-self: center;
  color: rgb(var(--color-text-muted));
  font-size: 0.9rem;
  font-weight: 700;
  line-height: 1.5;
}

.my-list-loading {
  padding: 2rem 1rem;
  color: rgb(var(--color-text-muted));
  font-size: 0.95rem;
  font-weight: 650;
  line-height: 1.6;
  text-align: center;
}

.my-table-wrap {
  overflow-x: auto;
}

.my-table {
  width: 100%;
  min-width: 72rem;
  border-collapse: collapse;
  table-layout: fixed;
}

.my-table th,
.my-table td {
  min-width: 0;
  border-bottom: 1px solid rgb(var(--color-border));
  padding: 0.8rem 1rem;
  text-align: left;
  vertical-align: top;
}

.my-table th {
  color: rgb(var(--color-text-muted));
  font-size: 0.75rem;
  font-weight: 800;
  letter-spacing: 0.08em;
  line-height: 1.2;
  overflow: hidden;
  text-overflow: ellipsis;
  text-transform: uppercase;
  white-space: nowrap;
}

.my-table td {
  color: rgb(var(--color-text-muted));
  font-size: 0.9rem;
  line-height: 1.5;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.my-table strong,
.my-table small {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.my-table strong {
  color: rgb(var(--color-text));
  font-size: 0.95rem;
  font-weight: 750;
  line-height: 1.35;
}

.my-table small {
  margin-top: 0.25rem;
  color: rgb(var(--color-text-muted));
  font-size: 0.78rem;
  line-height: 1.45;
}

.my-status-pill {
  display: inline-flex;
  max-width: 100%;
  min-height: 1.75rem;
  align-items: center;
  border-radius: 999px;
  background: rgb(var(--color-primary-soft));
  padding: 0 0.65rem;
  color: rgb(var(--color-primary));
  font-size: 0.78rem;
  font-weight: 800;
  line-height: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.my-action-group {
  display: flex;
  flex-wrap: nowrap;
  gap: 0.45rem;
}

.my-table-actions {
  width: 22rem;
  min-width: 22rem;
  overflow: visible;
}

.my-editor-dialog {
  position: fixed;
  z-index: 90;
  inset: 0;
  display: grid;
  place-items: center;
  background: rgb(15 23 42 / 0.42);
  padding: 1rem;
}

.my-editor-dialog__panel {
  display: flex;
  width: min(100%, 64rem);
  max-height: min(90vh, 60rem);
  flex-direction: column;
  overflow: hidden;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface-raised));
  box-shadow: 0 24px 70px rgb(15 23 42 / 0.22);
}

.my-editor-dialog__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  border-bottom: 1px solid rgb(var(--color-border));
  padding: 1rem;
}

.my-editor-dialog__header p {
  margin: 0;
  color: rgb(var(--color-text-muted));
  font-size: 0.75rem;
  font-weight: 800;
  letter-spacing: 0.08em;
  line-height: 1.2;
  text-transform: uppercase;
}

.my-editor-dialog__header h2 {
  margin: 0.35rem 0 0;
  color: rgb(var(--color-text));
  font-family: var(--font-display);
  font-size: 1.25rem;
  font-weight: 700;
  line-height: 1.25;
}

.my-editor-dialog__close {
  display: inline-flex;
  width: 2.25rem;
  height: 2.25rem;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text));
}

.my-editor-dialog__body {
  overflow: auto;
  padding: 1rem;
}

.my-editor-dialog__body :deep(.listing-editor-workspace) {
  margin-top: 0;
}

.my-editor-dialog-enter-active,
.my-editor-dialog-leave-active {
  transition: opacity 0.18s ease;
}

.my-editor-dialog-enter-active .my-editor-dialog__panel,
.my-editor-dialog-leave-active .my-editor-dialog__panel {
  transition:
    opacity 0.18s ease,
    transform 0.18s ease;
}

.my-editor-dialog-enter-from,
.my-editor-dialog-leave-to {
  opacity: 0;
}

.my-editor-dialog-enter-from .my-editor-dialog__panel,
.my-editor-dialog-leave-to .my-editor-dialog__panel {
  opacity: 0;
  transform: translateY(0.75rem);
}

@media (min-width: 760px) {
  .my-list-header {
    grid-template-columns: minmax(0, 1fr) auto;
    align-items: flex-start;
  }

  .my-list-toolbar {
    grid-template-columns: minmax(0, 1fr) minmax(9rem, 12rem) auto;
  }
}

@media (max-width: 759px) {
  .my-list-button {
    width: 100%;
  }

  .my-list-count {
    text-align: left;
  }
}
</style>

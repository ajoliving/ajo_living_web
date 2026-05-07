<!--
 * 二手交易設定頁。
 * 1. 使用左側設定導航切換發布通知與發現設定。
 * 2. 管理全部帖子狀態、發現頁大推與分類輪播廣告位。
-->
<script setup lang="ts">
import AppGlassSelect from '@/shared/components/base/AppGlassSelect.vue';
import AppIcon from '@/shared/components/base/AppIcon.vue';
import AppUnsavedChangesDialog from '@/shared/components/base/AppUnsavedChangesDialog.vue';
import ListingStatusBadge from '@/shared/components/marketplace/ListingStatusBadge.vue';

import { useMarketplaceSettingsPage } from './settings';

const {
  activeSection,
  assignSelectedToSlot,
  categoryCode,
  categoryOptions,
  clearSlot,
  deactivate,
  formatDate,
  formatPrice,
  getMarketplaceCategoryLabel,
  handleLeavePromptDecision,
  heroSlots,
  isLeavePromptOpen,
  keyword,
  listings,
  listingsLoading,
  loadListings,
  markSold,
  marketplaceCategories,
  noticeActionLabel,
  noticeActionURL,
  noticeBody,
  noticeTitle,
  placementsLoading,
  preferenceStore,
  publishNotice,
  publishingNotice,
  resolveListingCategoryLabel,
  resolveListingCommunityName,
  resolveListingCoverImage,
  resolveListingPrice,
  resolveListingPublishedAt,
  resolveListingStatus,
  resolveListingSummary,
  resolveListingTitle,
  savePlacements,
  savingPlacements,
  selectListing,
  selectedCategoryForSlots,
  selectedListingId,
  setActiveSection,
  settingsNavItems,
  status,
  statusOptions,
  t,
  visibleCategorySlots,
} = useMarketplaceSettingsPage();
</script>

<template>
  <main class="settings-page">
    <aside class="settings-sidebar">
      <div>
        <h1>{{ t('marketplace.settings.title') }}</h1>
        <p class="settings-description">
          {{ t('marketplace.settings.description') }}
        </p>
      </div>

      <nav class="settings-side-nav">
        <button
          v-for="item in settingsNavItems"
          :key="item.key"
          type="button"
          class="settings-side-nav-item"
          :class="activeSection === item.key ? 'settings-side-nav-item-active' : ''"
          @click="setActiveSection(item.key)"
        >
          <AppIcon
            :name="item.icon"
            :size="17"
          />
          <span>{{ item.label }}</span>
        </button>
      </nav>
    </aside>

    <section class="settings-content">
      <section
        v-if="activeSection === 'notice'"
        class="settings-panel"
      >
        <div class="settings-panel-header">
          <div>
            <h2>{{ t('marketplace.settings.noticePublisherTitle') }}</h2>
            <p>{{ t('marketplace.settings.noticePublisherDescription') }}</p>
          </div>
        </div>

        <div class="settings-notice-form">
          <label class="settings-field">
            <span>{{ t('marketplace.settings.noticeTitleField') }}</span>
            <input
              v-model="noticeTitle"
              type="text"
              maxlength="120"
              :placeholder="t('marketplace.settings.noticeTitlePlaceholder')"
            />
          </label>

          <label class="settings-field">
            <span>{{ t('marketplace.settings.noticeBodyField') }}</span>
            <textarea
              v-model="noticeBody"
              maxlength="1000"
              rows="5"
              :placeholder="t('marketplace.settings.noticeBodyPlaceholder')"
            />
          </label>

          <div class="settings-notice-actions">
            <label class="settings-field">
              <span>{{ t('marketplace.settings.noticeActionLabelField') }}</span>
              <input
                v-model="noticeActionLabel"
                type="text"
                maxlength="80"
                :placeholder="t('marketplace.settings.noticeActionLabelPlaceholder')"
              />
            </label>
            <label class="settings-field">
              <span>{{ t('marketplace.settings.noticeActionURLField') }}</span>
              <input
                v-model="noticeActionURL"
                type="text"
                maxlength="500"
                :placeholder="t('marketplace.settings.noticeActionURLPlaceholder')"
              />
            </label>
          </div>

          <div class="settings-notice-footer">
            <button
              type="button"
              class="settings-primary-button"
              :disabled="publishingNotice"
              @click="publishNotice"
            >
              <AppIcon
                name="send"
                :size="16"
              />
              <span>{{ publishingNotice ? t('marketplace.settings.noticePublishing') : t('marketplace.settings.noticePublishAction') }}</span>
            </button>
          </div>
        </div>
      </section>

      <section
        v-else
        class="settings-discover"
      >
        <div class="settings-section-header">
          <div>
            <p class="settings-kicker">
              {{ t('marketplace.settings.discoverSection') }}
            </p>
            <h2>{{ t('marketplace.settings.discoverPlacements') }}</h2>
          </div>
          <button
            type="button"
            class="settings-primary-button"
            :disabled="savingPlacements"
            @click="savePlacements"
          >
            <AppIcon
              name="check-circle"
              :size="17"
            />
            <span>{{ savingPlacements ? t('marketplace.settings.saving') : t('marketplace.settings.savePlacements') }}</span>
          </button>
        </div>

        <div class="settings-grid">
          <aside class="settings-panel settings-listings-panel">
            <div class="settings-panel-header">
              <div>
                <h2>{{ t('marketplace.settings.allListings') }}</h2>
                <p>{{ t('marketplace.filter.resultsFound', { count: listings.length }) }}</p>
              </div>
              <button
                type="button"
                class="settings-icon-button"
                :aria-label="t('marketplace.settings.refresh')"
                @click="loadListings"
              >
                <AppIcon
                  name="reload"
                  :size="16"
                />
              </button>
            </div>

            <div class="settings-filters">
              <label class="settings-search">
                <AppIcon
                  name="search"
                  :size="16"
                />
                <input
                  v-model="keyword"
                  type="search"
                  :placeholder="t('marketplace.settings.searchPlaceholder')"
                  @keyup.enter="loadListings"
                />
              </label>

              <div class="settings-filter-grid">
                <AppGlassSelect
                  v-model="categoryCode"
                  :options="categoryOptions"
                  @change="loadListings"
                />

                <AppGlassSelect
                  v-model="status"
                  :options="statusOptions"
                  @change="loadListings"
                />
              </div>

              <button
                type="button"
                class="settings-secondary-button"
                @click="loadListings"
              >
                <AppIcon
                  name="search"
                  :size="16"
                />
                <span>{{ t('marketplace.list.searchAction') }}</span>
              </button>
            </div>

            <div
              v-if="listingsLoading"
              class="settings-empty"
            >
              {{ t('marketplace.mine.loading') }}
            </div>
            <div
              v-else-if="listings.length === 0"
              class="settings-empty"
            >
              {{ t('marketplace.list.emptyTitle') }}
            </div>
            <div
              v-else
              class="settings-listings"
            >
              <article
                v-for="listing in listings"
                :key="listing.listing_id"
                class="settings-listing-card"
                :class="{ 'settings-listing-card-active': selectedListingId === listing.listing_id }"
                @click="selectListing(listing.listing_id)"
              >
                <div class="settings-listing-media">
                  <img
                    v-if="resolveListingCoverImage(listing)"
                    :src="resolveListingCoverImage(listing)?.url"
                    :alt="resolveListingTitle(listing)"
                  />
                  <AppIcon
                    v-else
                    name="picture"
                    :size="28"
                  />
                </div>
                <div class="min-w-0 flex-1">
                  <div class="mb-2 flex flex-wrap items-center gap-2">
                    <ListingStatusBadge
                      :status="resolveListingStatus(listing)"
                      :visibility="listing.visibility_scope"
                    />
                    <span class="settings-listing-price">
                      {{ formatPrice(resolveListingPrice(listing), preferenceStore.locale) }}
                    </span>
                  </div>
                  <h3 class="settings-listing-title">
                    {{ resolveListingTitle(listing) }}
                  </h3>
                  <p class="settings-listing-summary">
                    {{ resolveListingSummary(listing) }}
                  </p>
                  <div class="settings-listing-meta">
                    <span>{{ resolveListingCategoryLabel(listing, preferenceStore.locale) }}</span>
                    <span>{{ resolveListingCommunityName(listing) }}</span>
                    <span>{{ formatDate(resolveListingPublishedAt(listing), preferenceStore.locale) }}</span>
                  </div>
                  <div class="settings-listing-actions">
                    <button
                      type="button"
                      class="settings-mini-button"
                      :disabled="listing.business_status === 'sold'"
                      @click.stop="markSold(listing.listing_id)"
                    >
                      {{ t('marketplace.mine.markSoldAction') }}
                    </button>
                    <button
                      type="button"
                      class="settings-mini-button"
                      :disabled="listing.publication_status === 'hidden'"
                      @click.stop="deactivate(listing.listing_id)"
                    >
                      {{ t('marketplace.mine.deactivateAction') }}
                    </button>
                  </div>
                </div>
              </article>
            </div>
          </aside>

          <section class="settings-panel settings-slots-panel">
            <div class="settings-panel-header">
              <div>
                <h2>{{ t('marketplace.settings.discoverPlacements') }}</h2>
                <p>{{ t('marketplace.settings.selectedListing', { id: selectedListingId || '-' }) }}</p>
              </div>
            </div>

            <div
              v-if="placementsLoading"
              class="settings-empty"
            >
              {{ t('common.status.loading') }}
            </div>
            <div
              v-else
              class="settings-slot-stack"
            >
              <section class="settings-slot-section">
                <div class="settings-slot-section-header">
                  <h3>{{ t('marketplace.settings.heroSlots') }}</h3>
                </div>
                <div class="settings-slot-grid settings-slot-grid-hero">
                  <article
                    v-for="slot in heroSlots"
                    :key="slot.key"
                    class="settings-slot-card"
                  >
                    <div class="settings-slot-media">
                      <img
                        v-if="slot.listing && resolveListingCoverImage(slot.listing)"
                        :src="resolveListingCoverImage(slot.listing)?.url"
                        :alt="resolveListingTitle(slot.listing)"
                      />
                      <span v-else>{{ slot.slotIndex }}</span>
                    </div>
                    <div class="settings-slot-body">
                      <p class="settings-slot-label">
                        {{ t('marketplace.settings.slotLabel', { index: slot.slotIndex }) }}
                      </p>
                      <h4>{{ slot.listing ? resolveListingTitle(slot.listing) : t('marketplace.settings.emptySlot') }}</h4>
                      <p>{{ slot.listing ? resolveListingCategoryLabel(slot.listing, preferenceStore.locale) : t('marketplace.settings.emptySlotHint') }}</p>
                    </div>
                    <div class="settings-slot-actions">
                      <button
                        type="button"
                        class="settings-mini-button"
                        @click="assignSelectedToSlot(slot)"
                      >
                        {{ t('marketplace.settings.assignSelected') }}
                      </button>
                      <button
                        type="button"
                        class="settings-mini-button"
                        @click="clearSlot(slot)"
                      >
                        {{ t('marketplace.settings.clearSlot') }}
                      </button>
                    </div>
                  </article>
                </div>
              </section>

              <section class="settings-slot-section">
                <div class="settings-slot-section-header">
                  <h3>{{ t('marketplace.settings.categorySlots') }}</h3>
                  <AppGlassSelect
                    v-model="selectedCategoryForSlots"
                    class="settings-category-select"
                    :options="marketplaceCategories.map((category) => ({
                      label: getMarketplaceCategoryLabel(category.value, preferenceStore.locale),
                      value: category.value,
                    }))"
                  />
                </div>
                <div class="settings-slot-grid">
                  <article
                    v-for="slot in visibleCategorySlots"
                    :key="slot.key"
                    class="settings-slot-card"
                  >
                    <div class="settings-slot-media">
                      <img
                        v-if="slot.listing && resolveListingCoverImage(slot.listing)"
                        :src="resolveListingCoverImage(slot.listing)?.url"
                        :alt="resolveListingTitle(slot.listing)"
                      />
                      <span v-else>{{ slot.slotIndex }}</span>
                    </div>
                    <div class="settings-slot-body">
                      <p class="settings-slot-label">
                        {{ t('marketplace.settings.slotLabel', { index: slot.slotIndex }) }}
                      </p>
                      <h4>{{ slot.listing ? resolveListingTitle(slot.listing) : t('marketplace.settings.emptySlot') }}</h4>
                      <p>{{ slot.listing ? resolveListingCommunityName(slot.listing) : t('marketplace.settings.emptySlotHint') }}</p>
                    </div>
                    <div class="settings-slot-actions">
                      <button
                        type="button"
                        class="settings-mini-button"
                        @click="assignSelectedToSlot(slot)"
                      >
                        {{ t('marketplace.settings.assignSelected') }}
                      </button>
                      <button
                        type="button"
                        class="settings-mini-button"
                        @click="clearSlot(slot)"
                      >
                        {{ t('marketplace.settings.clearSlot') }}
                      </button>
                    </div>
                  </article>
                </div>
              </section>
            </div>
          </section>
        </div>
      </section>
    </section>

    <AppUnsavedChangesDialog
      :open="isLeavePromptOpen"
      :title="t('marketplace.settings.unsavedLeaveTitle')"
      :description="t('marketplace.settings.unsavedLeaveDescription')"
      :save-label="savingPlacements ? t('marketplace.settings.saving') : t('marketplace.settings.unsavedLeaveSave')"
      :discard-label="t('marketplace.settings.unsavedLeaveDiscard')"
      :stay-label="t('marketplace.settings.unsavedLeaveStay')"
      :saving="savingPlacements"
      @save="handleLeavePromptDecision('save')"
      @discard="handleLeavePromptDecision('discard')"
      @stay="handleLeavePromptDecision('stay')"
    />
  </main>
</template>

<style scoped>
.settings-page {
  display: grid;
  width: 100%;
  max-width: var(--layout-page-max-width);
  gap: 1.25rem;
  margin: 0 auto;
  padding: 1rem var(--layout-page-padding-inline) 5rem;
  color: #1a1c1b;
}

.settings-sidebar {
  --subroute-nav-active-color: color-mix(in srgb, rgb(var(--color-primary)) 78%, rgb(var(--color-text)) 22%);
  --subroute-nav-active-shadow: 0 0 10px rgb(var(--color-primary) / 0.16);
  --subroute-nav-underline: color-mix(in srgb, rgb(var(--color-primary)) 88%, rgb(var(--color-text)) 12%);
  display: grid;
  align-content: start;
  gap: 1.5rem;
  border: 1px solid rgb(var(--color-border) / 0.3);
  border-radius: 0.75rem;
  background:
    linear-gradient(
      180deg,
      rgb(var(--color-topbar-surface) / 0.76),
      rgb(var(--color-toolbar-surface) / 0.64)
    );
  box-shadow:
    0 16px 40px rgb(15 23 42 / 0.1),
    inset 0 -1px 0 rgb(255 255 255 / 0.06);
  padding: 1rem;
  backdrop-filter: blur(28px) saturate(184%);
  -webkit-backdrop-filter: blur(28px) saturate(184%);
}

.settings-kicker {
  margin: 0;
  color: #717878;
  font-size: 0.75rem;
  font-weight: 800;
  letter-spacing: 0.1em;
  line-height: 1;
  text-transform: uppercase;
}

.settings-sidebar h1 {
  margin: 0.75rem 0 0;
  color: #002727;
  font-family: var(--font-display);
  font-size: 1.75rem;
  font-weight: 500;
  line-height: 1.3;
}

.settings-description {
  margin: 0.65rem 0 0;
  color: #717878;
  font-size: 0.875rem;
  line-height: 1.7;
}

.settings-side-nav {
  display: grid;
  gap: 0.5rem;
}

.settings-side-nav-item {
  position: relative;
  display: inline-flex;
  min-height: 2.75rem;
  width: 100%;
  align-items: center;
  gap: 0.65rem;
  border: 1px solid rgb(var(--color-border) / 0.24);
  border-radius: 0.75rem;
  background:
    linear-gradient(
      180deg,
      rgb(var(--color-topbar-surface) / 0.62),
      rgb(var(--color-toolbar-surface) / 0.48)
    );
  padding: 0.65rem 0.85rem;
  color: rgb(var(--color-text) / 0.78);
  font-size: 0.9375rem;
  font-weight: 700;
  line-height: 1;
  text-align: left;
  box-shadow:
    0 12px 28px rgb(15 23 42 / 0.07),
    inset 0 1px 0 rgb(255 255 255 / 0.08);
  backdrop-filter: blur(18px) saturate(160%);
  -webkit-backdrop-filter: blur(18px) saturate(160%);
  transition:
    border-color 0.2s ease,
    background 0.2s ease,
    color 0.2s ease;
}

.settings-side-nav-item:hover {
  border-color: rgb(var(--color-border) / 0.38);
  color: rgb(var(--color-text));
}

.settings-side-nav-item-active,
.settings-side-nav-item-active:hover {
  border-color: rgb(var(--color-primary) / 0.28);
  background:
    linear-gradient(
      180deg,
      rgb(var(--color-topbar-surface) / 0.86),
      rgb(var(--color-primary-soft) / 0.24)
    );
  color: var(--subroute-nav-active-color);
  box-shadow:
    0 14px 32px rgb(var(--color-primary) / 0.1),
    inset 0 1px 0 rgb(255 255 255 / 0.1);
  text-shadow: var(--subroute-nav-active-shadow);
}

.settings-side-nav-item::after {
  position: absolute;
  left: 0.85rem;
  right: 0.85rem;
  bottom: 0.42rem;
  height: 1.5px;
  border-radius: 999px;
  background: var(--subroute-nav-underline);
  transform: scaleX(0);
  transform-origin: left center;
  opacity: 0;
  transition:
    transform 0.22s ease,
    opacity 0.22s ease;
  content: '';
}

.settings-side-nav-item-active::after {
  transform: scaleX(1);
  opacity: 1;
}

.settings-content {
  min-width: 0;
}

.settings-section-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 1rem;
}

.settings-section-header h2 {
  margin-top: 0.35rem;
  color: #1a1c1b;
  font-family: var(--font-display);
  font-size: clamp(1.5rem, 2vw, 2.25rem);
  font-weight: 650;
  line-height: 1.2;
}

.settings-grid {
  display: grid;
  grid-template-columns: minmax(22rem, 0.85fr) minmax(0, 1.4fr);
  gap: 1rem;
  align-items: start;
}

.settings-panel {
  border: 1px solid #e2e3e1;
  border-radius: 8px;
  background: #fbfbf8;
}

.settings-listings-panel {
  position: sticky;
  top: 7rem;
  max-height: calc(100vh - 8rem);
  overflow: hidden;
}

.settings-panel-header,
.settings-slot-section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 1rem;
  border-bottom: 1px solid #e2e3e1;
}

.settings-panel-header h2,
.settings-slot-section-header h3 {
  font-size: 1rem;
  font-weight: 700;
  color: #1a1c1b;
}

.settings-panel-header p {
  margin-top: 0.2rem;
  font-size: 0.82rem;
  color: #717878;
}

.settings-notice-form,
.settings-filters {
  display: grid;
  gap: 0.75rem;
  padding: 1rem;
}

.settings-filters {
  border-bottom: 1px solid #e2e3e1;
}

.settings-notice-actions,
.settings-filter-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.75rem;
}

.settings-notice-footer {
  display: flex;
  justify-content: flex-end;
}

.settings-field {
  display: grid;
  gap: 0.45rem;
}

.settings-field span {
  color: #545a58;
  font-size: 0.78rem;
  font-weight: 700;
}

.settings-field input,
.settings-field textarea,
.settings-search {
  width: 100%;
  border: 1px solid #d2d5d1;
  border-radius: 8px;
  background: #fff;
  color: #1a1c1b;
  font-size: 0.92rem;
  outline: 0;
}

.settings-field input,
.settings-field textarea {
  padding: 0.75rem;
  line-height: 1.5;
}

.settings-field textarea {
  resize: vertical;
}

.settings-search {
  display: flex;
  align-items: center;
  gap: 0.55rem;
  min-height: 2.75rem;
  padding: 0 0.75rem;
  color: #717878;
}

.settings-search input {
  border: 0;
  background: transparent;
}

.settings-primary-button,
.settings-secondary-button,
.settings-icon-button,
.settings-mini-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.45rem;
  border: 1px solid #1a1c1b;
  border-radius: 8px;
  color: #1a1c1b;
  background: #fff;
  font-size: 0.9rem;
  font-weight: 700;
  transition:
    background 0.2s ease,
    color 0.2s ease,
    border-color 0.2s ease;
}

.settings-primary-button {
  min-height: 2.75rem;
  padding: 0 1rem;
  background: #1a1c1b;
  color: #fff;
}

.settings-secondary-button {
  min-height: 2.65rem;
}

.settings-icon-button {
  width: 2.5rem;
  height: 2.5rem;
}

.settings-mini-button {
  min-height: 2rem;
  padding: 0 0.65rem;
  font-size: 0.78rem;
}

.settings-primary-button:disabled,
.settings-mini-button:disabled {
  cursor: not-allowed;
  opacity: 0.45;
}

.settings-primary-button:not(:disabled):hover,
.settings-secondary-button:hover,
.settings-icon-button:hover,
.settings-mini-button:not(:disabled):hover {
  background: #1a1c1b;
  color: #fff;
}

.settings-empty {
  padding: 1rem;
  color: #717878;
  font-size: 0.9rem;
}

.settings-listings {
  display: grid;
  gap: 0.75rem;
  max-height: calc(100vh - 24rem);
  overflow-y: auto;
  padding: 1rem;
}

.settings-listing-card {
  display: grid;
  grid-template-columns: 5.5rem minmax(0, 1fr);
  gap: 0.85rem;
  border: 1px solid #e2e3e1;
  border-radius: 8px;
  background: #fff;
  padding: 0.75rem;
  cursor: pointer;
  transition:
    border-color 0.2s ease,
    box-shadow 0.2s ease;
}

.settings-listing-card-active {
  border-color: #1a1c1b;
  box-shadow: 0 12px 34px rgb(26 28 27 / 0.1);
}

.settings-listing-media,
.settings-slot-media {
  overflow: hidden;
  border-radius: 8px;
  background: #e2e3e1;
  color: #717878;
}

.settings-listing-media {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 6.25rem;
}

.settings-listing-media img,
.settings-slot-media img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.settings-listing-price {
  font-size: 0.82rem;
  font-weight: 700;
  color: #1a1c1b;
}

.settings-listing-title,
.settings-slot-body h4 {
  overflow: hidden;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  font-weight: 700;
  color: #1a1c1b;
  line-height: 1.35;
}

.settings-listing-summary,
.settings-slot-body p {
  overflow: hidden;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  margin-top: 0.25rem;
  color: #717878;
  font-size: 0.82rem;
  line-height: 1.55;
}

.settings-listing-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 0.45rem;
  margin-top: 0.65rem;
  color: #717878;
  font-size: 0.76rem;
}

.settings-listing-actions,
.settings-slot-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.45rem;
  margin-top: 0.75rem;
}

.settings-slots-panel {
  min-width: 0;
}

.settings-slot-stack {
  display: grid;
  gap: 1rem;
  padding: 1rem;
}

.settings-slot-section {
  border: 1px solid #e2e3e1;
  border-radius: 8px;
  background: #fff;
}

.settings-category-select {
  max-width: 14rem;
}

.settings-slot-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.75rem;
  padding: 1rem;
}

.settings-slot-grid-hero {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.settings-slot-card {
  display: grid;
  grid-template-columns: 7rem minmax(0, 1fr);
  gap: 0.85rem;
  border: 1px solid #e2e3e1;
  border-radius: 8px;
  padding: 0.75rem;
}

.settings-slot-media {
  display: flex;
  align-items: center;
  justify-content: center;
  aspect-ratio: 1 / 1;
  font-family: var(--font-display);
  font-size: 1.75rem;
  font-weight: 700;
}

.settings-slot-label {
  margin-top: 0;
  color: #717878;
  font-size: 0.76rem;
  font-weight: 700;
}

@media (min-width: 1024px) {
  .settings-page {
    grid-template-columns: 15.75rem minmax(0, 1fr);
    align-items: start;
  }

  .settings-sidebar {
    position: sticky;
    top: calc(var(--app-header-offset, 0rem) + 2rem);
  }
}

@media (max-width: 1100px) {
  .settings-grid {
    grid-template-columns: 1fr;
  }

  .settings-listings-panel {
    position: static;
    max-height: none;
  }

  .settings-listings {
    max-height: none;
  }
}

@media (max-width: 767px) {
  .settings-page {
    padding: 1rem var(--layout-page-padding-inline) 4rem;
  }

  .settings-section-header,
  .settings-panel-header,
  .settings-slot-section-header {
    align-items: stretch;
    flex-direction: column;
  }

  .settings-notice-actions,
  .settings-filter-grid,
  .settings-slot-grid,
  .settings-slot-grid-hero {
    grid-template-columns: 1fr;
  }

  .settings-listing-card,
  .settings-slot-card {
    grid-template-columns: 1fr;
  }

  .settings-listing-media {
    height: 12rem;
  }
}
</style>

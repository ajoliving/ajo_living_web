<!--
 * 我的帖子詳情頁。
 * 1. 建立發布者視角的帖子完整檢視、編輯入口與狀態操作。
 * 2. 接入真實帖子狀態、圖片、價格與流程資料。
-->
<script setup lang="ts">
import AppIcon from '@/shared/components/base/AppIcon.vue';

import { useMarketplaceMyListingPreviewPage } from './preview';

const {
  actionLoading,
  canDeactivate,
  canMarkSold,
  canPublish,
  canRepublish,
  canRenew,
  categoryLabel,
  coverImage,
  districtLabel,
  formatAjoPoints,
  listing,
  listingId,
  listingPrice,
  loading,
  openEditor,
  openPublicDetail,
  publishedAt,
  runAction,
  secondhandRenewChargeCost,
  statusLabel,
  t,
} = useMarketplaceMyListingPreviewPage();
</script>

<template>
  <main class="my-preview-page">
    <section
      v-if="loading"
      class="my-preview-summary"
    >
      {{ t('common.status.loading') }}
    </section>

    <section
      v-if="listing"
      class="my-preview-summary"
    >
      <div>
        <p class="preview-kicker">
          {{ t('marketplace.mine.detailKicker') }}
        </p>
        <p class="preview-reference">
          {{ t('marketplace.detail.referenceId', { id: listingId }) }}
        </p>
        <h1>{{ listing.title }}</h1>
        <p>{{ listing.summary }}</p>
      </div>
      <div class="preview-actions">
        <button
          type="button"
          class="preview-action preview-action--primary"
          :disabled="actionLoading"
          @click="openEditor"
        >
          <AppIcon
            name="palette"
            :size="16"
          />
          {{ t('marketplace.mine.editDetailAction') }}
        </button>
        <button
          type="button"
          class="preview-action"
          :disabled="actionLoading"
          @click="openPublicDetail"
        >
          <AppIcon
            name="view"
            :size="16"
          />
          {{ t('marketplace.mine.publicDetailAction') }}
        </button>
        <button
          v-if="canPublish"
          type="button"
          class="preview-action preview-action--primary"
          :disabled="actionLoading"
          @click="runAction('publish')"
        >
          <AppIcon
            name="plus-square"
            :size="16"
          />
          {{ t('marketplace.mine.publishAction') }}
        </button>
        <button
          v-if="canRepublish"
          type="button"
          class="preview-action preview-action--primary"
          :disabled="actionLoading"
          @click="runAction('republish')"
        >
          <AppIcon
            name="reload"
            :size="16"
          />
          {{ t('marketplace.mine.republishAction') }}
        </button>
        <button
          v-if="canRenew"
          type="button"
          class="preview-action preview-action--primary"
          :disabled="actionLoading"
          @click="runAction('renew')"
        >
          <AppIcon
            name="clock"
            :size="16"
          />
          {{ t('marketplace.mine.renewAction') }} · {{ formatAjoPoints(secondhandRenewChargeCost) }}
        </button>
        <button
          v-if="canMarkSold"
          type="button"
          class="preview-action"
          :disabled="actionLoading"
          @click="runAction('mark-sold')"
        >
          <AppIcon
            name="check-circle"
            :size="16"
          />
          {{ t('marketplace.mine.markSoldAction') }}
        </button>
        <button
          v-if="canDeactivate"
          type="button"
          class="preview-action"
          :disabled="actionLoading"
          @click="runAction('deactivate')"
        >
          <AppIcon
            name="close"
            :size="16"
          />
          {{ t('marketplace.mine.deactivateAction') }}
        </button>
      </div>
    </section>

    <section
      v-if="listing"
      class="my-preview-layout"
    >
      <div class="preview-media">
        <img
          v-if="coverImage"
          :src="coverImage.url"
          :alt="coverImage.alt"
        >
        <div
          v-else
          class="flex h-full min-h-[22rem] items-center justify-center text-text-muted"
        >
          <AppIcon
            name="picture"
            :size="44"
          />
        </div>
      </div>

      <article class="preview-detail">
        <div class="preview-price-row">
          <div class="preview-price">{{ listingPrice }}</div>
          <span class="preview-status-pill">{{ statusLabel }}</span>
        </div>
        <dl class="preview-specs">
          <div>
            <dt>{{ t('marketplace.mine.category') }}</dt>
            <dd>{{ categoryLabel }}</dd>
          </div>
          <div>
            <dt>{{ t('marketplace.mine.published') }}</dt>
            <dd>{{ publishedAt }}</dd>
          </div>
          <div>
            <dt>{{ t('marketplace.mine.community') }}</dt>
            <dd>{{ districtLabel }}</dd>
          </div>
        </dl>
      </article>

      <aside class="preview-timeline">
        <p class="preview-kicker">{{ t('marketplace.preview.timeline') }}</p>
        <ol>
          <li>{{ t('marketplace.preview.created') }}</li>
          <li>{{ t('marketplace.preview.published') }}</li>
          <li>{{ t('marketplace.preview.readyForChat') }}</li>
        </ol>
      </aside>
    </section>
  </main>
</template>

<style scoped>
.my-preview-page {
  display: grid;
  width: 100%;
  max-width: min(1240px, calc(100vw - 32px));
  gap: 16px;
  margin: 0 auto;
  padding: 10px 0 72px;
  color: rgb(var(--color-text));
}

.my-preview-summary,
.preview-media,
.preview-detail,
.preview-timeline {
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
}

.my-preview-summary,
.preview-detail,
.preview-timeline {
  padding: 16px;
}

.my-preview-summary {
  display: grid;
  gap: 16px;
}

.preview-kicker {
  margin: 0;
  color: rgb(var(--color-primary));
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 0.14em;
  line-height: 1;
  text-transform: uppercase;
}

.preview-reference {
  margin: 8px 0 0;
  color: rgb(var(--color-text-muted));
  font-size: 12px;
  font-weight: 700;
  line-height: 1.5;
}

.my-preview-summary h1 {
  margin: 8px 0 0;
  color: rgb(var(--color-text));
  font-family: var(--font-display);
  font-size: clamp(1.9rem, 3.2vw, 2.5rem);
  font-weight: 400;
  line-height: 1.15;
}

.my-preview-summary p:not(.preview-kicker) {
  margin-top: 8px;
  max-width: 48rem;
  color: rgb(var(--color-text-muted));
  font-size: 13px;
  line-height: 1.65;
}

.preview-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.preview-action {
  display: inline-flex;
  min-height: 36px;
  align-items: center;
  justify-content: center;
  gap: 6px;
  border: 1px solid rgb(var(--color-border));
  border-radius: 6px;
  padding: 0 12px;
  color: rgb(var(--color-text-muted));
  font-size: 12px;
  font-weight: 700;
  line-height: 1;
  transition:
    border-color 0.2s ease,
    background-color 0.2s ease,
    color 0.2s ease,
    opacity 0.2s ease;
}

.preview-action:hover:not(:disabled) {
  border-color: rgb(var(--color-primary));
  color: rgb(var(--color-primary));
}

.preview-action--primary {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary));
  color: rgb(var(--color-primary-contrast));
}

.preview-action--primary:hover:not(:disabled) {
  border-color: rgb(var(--color-text));
  background: rgb(var(--color-text));
  color: rgb(var(--color-surface));
}

.preview-action:disabled {
  cursor: not-allowed;
  opacity: 0.45;
}

.my-preview-layout {
  display: grid;
  gap: 16px;
}

.preview-media {
  overflow: hidden;
}

.preview-media img {
  width: 100%;
  height: 100%;
  min-height: 22rem;
  object-fit: cover;
}

.preview-price-row {
  display: flex;
  align-items: start;
  justify-content: space-between;
  gap: 12px;
}

.preview-price {
  font-family: var(--font-display);
  font-size: clamp(2rem, 4vw, 2.8rem);
  font-weight: 400;
  color: rgb(var(--color-primary));
}

.preview-status-pill {
  display: inline-flex;
  min-height: 30px;
  align-items: center;
  justify-content: center;
  border: 1px solid rgb(var(--color-primary) / 0.15);
  border-radius: 999px;
  background: rgb(var(--color-primary) / 0.08);
  padding: 0 12px;
  color: rgb(var(--color-primary));
  font-size: 11px;
  font-weight: 700;
}

.preview-specs {
  display: grid;
  gap: 12px;
  margin-top: 16px;
}

.preview-specs dt {
  color: rgb(var(--color-text-muted));
  font-size: 12px;
  font-weight: 700;
}

.preview-specs dd {
  margin: 4px 0 0;
  color: rgb(var(--color-text));
  font-size: 13px;
  font-weight: 700;
}

.preview-timeline ol {
  display: grid;
  gap: 12px;
  margin: 12px 0 0;
  padding-left: 1.25rem;
  color: rgb(var(--color-text-muted));
  font-size: 13px;
  line-height: 1.65;
}

@media (min-width: 768px) {
  .preview-specs {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (min-width: 1024px) {
  .my-preview-layout {
    grid-template-columns: minmax(0, 1fr) 20rem;
  }

  .preview-media {
    grid-row: span 2;
  }
}

@media (max-width: 767px) {
  .my-preview-page {
    max-width: 100%;
    padding-bottom: 96px;
  }
}
</style>

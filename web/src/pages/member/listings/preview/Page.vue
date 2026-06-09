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
      <p class="preview-kicker">
        {{ t('marketplace.mine.detailKicker') }}
      </p>
      <p class="preview-reference">
        {{ t('marketplace.detail.referenceId', { id: listingId }) }}
      </p>
      <h1>{{ listing.title }}</h1>
      <p>{{ listing.summary }}</p>
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
        <div class="preview-price">{{ listingPrice }}</div>
        <dl class="preview-specs">
          <div>
            <dt>{{ t('marketplace.mine.category') }}</dt>
            <dd>{{ categoryLabel }}</dd>
          </div>
          <div>
            <dt>{{ t('common.label.status') }}</dt>
            <dd>{{ statusLabel }}</dd>
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
  max-width: var(--layout-page-max-width);
  gap: 1.5rem;
  margin: 0 auto;
  padding: 1rem var(--layout-page-padding-inline) 5rem;
  color: rgb(var(--color-text));
}

.my-preview-summary,
.preview-media,
.preview-detail,
.preview-timeline {
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.75rem;
  background: rgb(var(--color-surface));
  box-shadow: 0 10px 30px -5px rgb(0 39 39 / 0.05);
}

.my-preview-summary,
.preview-detail,
.preview-timeline {
  padding: 2rem;
}

.preview-kicker {
  margin: 0;
  color: rgb(var(--color-text-muted));
  font-size: 0.75rem;
  font-weight: 600;
  letter-spacing: 0.1em;
  line-height: 1;
  text-transform: uppercase;
}

.preview-reference {
  margin: 0.75rem 0 0;
  color: rgb(var(--color-text-muted));
  font-size: 0.875rem;
  font-weight: 700;
  line-height: 1.5;
}

.my-preview-summary h1 {
  margin: 0.9rem 0 0;
  color: rgb(var(--color-primary));
  font-family: var(--font-display);
  font-size: clamp(2rem, 4vw, 3rem);
  font-weight: 500;
  line-height: 1.2;
}

.my-preview-summary p:not(.preview-kicker) {
  margin-top: 1rem;
  max-width: 44rem;
  color: rgb(var(--color-text-muted));
  line-height: 1.75;
}

.preview-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  margin-top: 1.5rem;
}

.preview-action {
  display: inline-flex;
  min-height: 3.25rem;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  border: 1px solid rgb(var(--color-border));
  border-radius: 999px;
  padding: 0.8rem 1.25rem;
  color: rgb(var(--color-text-muted));
  font-size: 0.875rem;
  font-weight: 800;
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
  gap: 1.5rem;
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

.preview-price {
  font-family: var(--font-display);
  font-size: clamp(2rem, 4vw, 3.2rem);
  font-weight: 600;
  color: rgb(var(--color-primary));
}

.preview-specs {
  display: grid;
  gap: 1rem;
  margin-top: 1.5rem;
}

.preview-specs dt {
  color: rgb(var(--color-text-muted));
  font-size: 0.75rem;
  font-weight: 600;
  letter-spacing: 0.1em;
  text-transform: uppercase;
}

.preview-specs dd {
  margin: 0.45rem 0 0;
  color: rgb(var(--color-text));
  font-weight: 600;
}

.preview-timeline ol {
  display: grid;
  gap: 1rem;
  margin: 1rem 0 0;
  padding-left: 1.25rem;
  color: rgb(var(--color-text-muted));
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
    padding: 1rem var(--layout-page-padding-inline) 4rem;
  }

  .my-preview-summary,
  .preview-detail,
  .preview-timeline {
    padding: 1.5rem;
  }
}
</style>

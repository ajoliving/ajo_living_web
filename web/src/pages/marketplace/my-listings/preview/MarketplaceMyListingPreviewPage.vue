<!--
 * 我的帖子預覽頁。
 * 1. 建立我的帖子完整檢視、狀態操作與流程記錄布局。
 * 2. 接入真實帖子狀態與預覽資料。
-->
<script setup lang="ts">
import AppIcon from '@/shared/components/base/AppIcon.vue';

import { useMarketplaceMyListingPreviewPage } from './preview';

const {
  categoryLabel,
  coverImage,
  districtLabel,
  listing,
  listingId,
  listingPrice,
  loading,
  publishedAt,
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
        {{ t('marketplace.detail.referenceId', { id: listingId }) }}
      </p>
      <h1>{{ listing.title }}</h1>
      <p>{{ listing.summary }}</p>
      <div class="preview-actions">
        <button
          type="button"
          class="preview-action preview-action--primary"
        >
          <AppIcon
            name="reload"
            :size="16"
          />
          {{ t('marketplace.mine.republishAction') }}
        </button>
        <button
          type="button"
          class="preview-action"
        >
          <AppIcon
            name="check-circle"
            :size="16"
          />
          {{ t('marketplace.mine.markSoldAction') }}
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
          class="flex h-full min-h-[22rem] items-center justify-center text-[#717878]"
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
  max-width: 1280px;
  min-height: calc(100vh - var(--app-header-offset, 0rem));
  gap: 1.5rem;
  margin: 0 auto;
  padding: 4rem 2rem;
  color: #1a1c1b;
}

.my-preview-summary,
.preview-media,
.preview-detail,
.preview-timeline {
  border: 1px solid #e2e3e1;
  border-radius: 0.75rem;
  background: #f9f9f7;
  box-shadow: 0 4px 24px rgb(0 0 0 / 0.03);
}

.my-preview-summary,
.preview-detail,
.preview-timeline {
  padding: 1.5rem;
}

.preview-kicker {
  margin: 0;
  color: #717878;
  font-size: 0.75rem;
  font-weight: 600;
  letter-spacing: 0.1em;
  line-height: 1;
  text-transform: uppercase;
}

.my-preview-summary h1 {
  margin: 0.9rem 0 0;
  font-family: var(--font-display);
  font-size: clamp(2.2rem, 5vw, 4rem);
  font-weight: 500;
  line-height: 1.05;
}

.my-preview-summary p:not(.preview-kicker) {
  margin-top: 1rem;
  max-width: 44rem;
  color: #414848;
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
  min-height: 2.75rem;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  border: 1px solid #c1c8c7;
  border-radius: 999px;
  padding: 0.6rem 1rem;
  color: #414848;
  font-weight: 600;
}

.preview-action--primary {
  border-color: #002727;
  background: #002727;
  color: #ffffff;
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
  color: #002727;
}

.preview-specs {
  display: grid;
  gap: 1rem;
  margin-top: 1.5rem;
}

.preview-specs dt {
  color: #717878;
  font-size: 0.75rem;
  font-weight: 600;
  letter-spacing: 0.1em;
  text-transform: uppercase;
}

.preview-specs dd {
  margin: 0.45rem 0 0;
  color: #1a1c1b;
  font-weight: 600;
}

.preview-timeline ol {
  display: grid;
  gap: 1rem;
  margin: 1rem 0 0;
  padding-left: 1.25rem;
  color: #414848;
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
    padding: 2.5rem 1.25rem;
  }
}
</style>

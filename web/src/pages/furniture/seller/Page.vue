<!--
 * 二手交易賣家頁。
 * 1. 建立賣家摘要、信任資訊與帖子列表布局。
 * 2. 先保留靜態骨架，後續再接入賣家公開資料。
-->
<script setup lang="ts">
import { computed } from 'vue';
import { useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';

import AppIcon from '@/shared/components/base/AppIcon.vue';
import { usePreferenceStore } from '@/app/stores/preferences';
import { formatPrice } from '@/shared/utils/format';

const route = useRoute();
const { t } = useI18n();
const preferenceStore = usePreferenceStore();

// 1. 取得目前賣家識別碼
const sellerId = String(route.params.sellerId ?? '');

const sampleItems = [
  '/home-stage/carousel/building.jpeg',
  '/home-stage/carousel/intercom.png',
  '/home-stage/carousel/rant.png',
  '/home-stage/secondhand.webp',
];
const samplePrice = computed(() => formatPrice(680, preferenceStore.locale));
</script>

<template>
  <main class="seller-page">
    <aside class="seller-profile">
      <div class="seller-avatar">
        {{ t('marketplace.seller.avatarFallback') }}
      </div>
      <p class="seller-kicker">
        {{ t('marketplace.seller.referenceId', { id: sellerId || 'seller-demo' }) }}
      </p>
      <h1>{{ t('marketplace.detail.sampleSeller') }}</h1>
      <p>{{ t('marketplace.seller.description') }}</p>
      <div class="seller-stats">
        <div>
          <strong>12</strong>
          <span>{{ t('marketplace.seller.activeListings') }}</span>
        </div>
        <div>
          <strong>4.8</strong>
          <span>{{ t('marketplace.seller.response') }}</span>
        </div>
      </div>
    </aside>

    <section class="seller-results">
      <div class="seller-results__header">
        <div>
          <p class="seller-kicker">{{ t('marketplace.seller.items') }}</p>
          <h2>{{ t('marketplace.seller.otherListings') }}</h2>
        </div>
        <button
          type="button"
          class="seller-contact"
        >
          <AppIcon
            name="message"
            :size="17"
          />
          {{ t('common.action.openChat') }}
        </button>
      </div>

      <div class="seller-grid">
        <article
          v-for="image in sampleItems"
          :key="image"
          class="seller-card group"
        >
          <div class="seller-card__media">
            <img
              :src="image"
              alt=""
            >
          </div>
          <div class="seller-card__body">
            <h3>{{ t('marketplace.detail.sampleMiniTitle') }}</h3>
            <p>{{ t('marketplace.detail.sampleSummary') }}</p>
            <strong>{{ samplePrice }}</strong>
          </div>
        </article>
      </div>
    </section>
  </main>
</template>

<style scoped>
.seller-page {
  display: grid;
  width: 100%;
  max-width: var(--layout-page-max-width);
  min-height: calc(100vh - var(--app-header-offset, 0rem));
  gap: 2rem;
  margin: 0 auto;
  padding: 4rem var(--layout-page-padding-inline);
  color: rgb(var(--color-text));
}

.seller-profile,
.seller-results {
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.75rem;
  background: rgb(var(--color-surface-raised));
  box-shadow: 0 4px 24px rgb(0 0 0 / 0.03);
}

.seller-profile {
  align-content: start;
  padding: 1.5rem;
}

.seller-avatar {
  display: grid;
  width: 4.5rem;
  height: 4.5rem;
  place-items: center;
  border-radius: 999px;
  background: rgb(var(--color-primary));
  color: rgb(var(--color-primary-contrast));
  font-family: var(--font-display);
  font-size: 1.5rem;
}

.seller-kicker {
  margin: 1.5rem 0 0;
  color: rgb(var(--color-text-muted));
  font-size: 0.75rem;
  font-weight: 600;
  letter-spacing: 0.1em;
  line-height: 1;
  text-transform: uppercase;
}

.seller-profile h1,
.seller-results__header h2 {
  margin: 0.75rem 0 0;
  font-family: var(--font-display);
  font-size: clamp(2rem, 4vw, 2.8rem);
  font-weight: 500;
  line-height: 1.1;
}

.seller-profile p:not(.seller-kicker) {
  margin-top: 1rem;
  color: rgb(var(--color-text-muted));
  line-height: 1.75;
}

.seller-stats {
  display: grid;
  gap: 0.75rem;
  margin-top: 1.5rem;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.seller-stats div {
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.75rem;
  padding: 1rem;
}

.seller-stats strong {
  display: block;
  font-family: var(--font-display);
  font-size: 2rem;
  color: rgb(var(--color-primary));
}

.seller-stats span {
  color: rgb(var(--color-text-muted));
  font-size: 0.8rem;
}

.seller-results {
  padding: 1.5rem;
}

.seller-results__header {
  display: grid;
  gap: 1rem;
  align-items: center;
  border-bottom: 1px solid rgb(var(--color-border));
  padding-bottom: 1.5rem;
}

.seller-contact {
  display: inline-flex;
  min-height: 2.75rem;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  border-radius: 999px;
  background: rgb(var(--color-primary));
  padding: 0.6rem 1rem;
  color: rgb(var(--color-primary-contrast));
  font-size: 0.875rem;
  font-weight: 600;
}

.seller-grid {
  display: grid;
  gap: 1.25rem;
  margin-top: 1.5rem;
}

.seller-card {
  overflow: hidden;
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.75rem;
  background: rgb(var(--color-surface-raised));
  box-shadow: 0 4px 24px rgb(0 0 0 / 0.03);
}

.seller-card__media {
  height: 12rem;
  overflow: hidden;
  background: rgb(var(--color-surface-muted));
}

.seller-card__media img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 0.5s ease;
}

.seller-card:hover img {
  transform: scale(1.05);
}

.seller-card__body {
  padding: 1.25rem;
}

.seller-card__body h3 {
  margin: 0;
  font-family: var(--font-display);
  font-size: 1.35rem;
  color: rgb(var(--color-text));
}

.seller-card__body p {
  margin-top: 0.5rem;
  color: rgb(var(--color-text-muted));
  line-height: 1.65;
}

.seller-card__body strong {
  display: block;
  margin-top: 1rem;
  color: rgb(var(--color-primary));
  font-family: var(--font-display);
  font-size: 1.5rem;
}

@media (min-width: 768px) {
  .seller-results__header {
    grid-template-columns: 1fr auto;
  }

  .seller-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (min-width: 1024px) {
  .seller-page {
    grid-template-columns: 20rem minmax(0, 1fr);
  }

  .seller-profile {
    position: sticky;
    top: calc(var(--app-header-offset, 0rem) + 2rem);
  }
}

@media (max-width: 767px) {
  .seller-page {
    padding: 2.5rem var(--layout-page-padding-inline);
  }
}
</style>

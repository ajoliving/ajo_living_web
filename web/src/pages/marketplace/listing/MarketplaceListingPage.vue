<!--
 * 二手帖子詳情頁。
 * 1. 參考 28Hse 商品詳情資訊密度建立圖片、詳情、聯絡與賣家區布局。
 * 2. 接入真實帖子詳情、聯絡權限與聊天入口。
-->
<script setup lang="ts">
import AppIcon from '@/shared/components/base/AppIcon.vue';
import BaseAvatar from '@/shared/components/base/BaseAvatar.vue';

import { useMarketplaceListingPage } from './listing';

const {
  categoryLabel,
  conditionLabel,
  contactPayload,
  coverImage,
  districtLabel,
  formatDeliveryTag,
  galleryImages,
  listing,
  listingId,
  listingPrice,
  loading,
  loadingContact,
  openChat,
  openingChat,
  ownerAvatarUrl,
  ownerName,
  publishedAt,
  revealContact,
  t,
} = useMarketplaceListingPage();
</script>

<template>
  <main class="marketplace-detail-page">
    <section
      v-if="loading"
      class="detail-panel"
    >
      {{ t('common.status.loading') }}
    </section>

    <section
      v-else-if="listing"
      class="detail-main"
    >
      <div class="detail-gallery">
        <div class="detail-gallery__hero">
          <img
            v-if="coverImage"
            :src="coverImage.url"
            :alt="coverImage.alt"
          >
          <div
            v-else
            class="flex h-full items-center justify-center text-text-muted"
          >
            <AppIcon
              name="picture"
              :size="44"
            />
          </div>
        </div>
        <div class="detail-gallery__thumbs">
          <button
            v-for="image in galleryImages"
            :key="image.id"
            type="button"
            class="detail-gallery__thumb"
          >
            <img
              :src="image.url"
              :alt="image.alt"
            >
          </button>
        </div>
      </div>

      <article class="detail-content">
        <div class="detail-kicker">
          {{ t('marketplace.detail.referenceId', { id: listingId }) }}
        </div>
        <div class="detail-heading">
          <div>
            <h1>{{ listing.title }}</h1>
            <p>{{ listing.summary }}</p>
          </div>
          <p class="detail-price">{{ listingPrice }}</p>
        </div>

        <section class="detail-section">
          <h2>{{ t('marketplace.detail.description') }}</h2>
          <p>
            {{ listing.description }}
          </p>
        </section>

        <section class="detail-section">
          <h2>{{ t('marketplace.detail.specification') }}</h2>
          <dl class="detail-spec-grid">
            <div>
              <dt>{{ t('marketplace.mine.category') }}</dt>
              <dd>{{ categoryLabel }}</dd>
            </div>
            <div>
              <dt>{{ t('common.label.condition') }}</dt>
              <dd>{{ conditionLabel }}</dd>
            </div>
            <div>
              <dt>{{ t('marketplace.detail.pickupArea') }}</dt>
              <dd>{{ districtLabel }}</dd>
            </div>
            <div>
              <dt>{{ t('marketplace.detail.publishedAt') }}</dt>
              <dd>{{ publishedAt }}</dd>
            </div>
            <div>
              <dt>{{ t('marketplace.editor.dimensionField') }}</dt>
              <dd>{{ listing.dimension_text || '-' }}</dd>
            </div>
          </dl>
        </section>

        <section class="detail-section">
          <h2>{{ t('marketplace.detail.tradeInfo') }}</h2>
          <div class="detail-tags">
            <span>{{ listing.pickup_location_text }}</span>
            <span
              v-for="tag in listing.delivery_tags"
              :key="tag"
            >
              {{ formatDeliveryTag(tag) }}
            </span>
            <span>{{ listing.visibility_scope === 'public' ? t('marketplace.detail.publicListing') : t('common.state.buildingOnly') }}</span>
          </div>
        </section>
      </article>
    </section>

    <aside
      v-if="listing"
      class="detail-sidebar"
    >
      <section class="detail-panel">
        <p class="detail-panel__label">{{ t('marketplace.detail.seller') }}</p>
        <div class="detail-seller-profile">
          <BaseAvatar
            :src="ownerAvatarUrl"
            :name="ownerName"
            :alt="ownerName"
            :size="56"
          />
          <div>
            <h2>{{ ownerName }}</h2>
            <p>{{ listing.visibility_scope === 'public' ? t('marketplace.detail.publicHint') : t('marketplace.detail.buildingHint') }}</p>
          </div>
        </div>
        <div class="detail-actions">
          <button
            type="button"
            class="detail-action detail-action--primary"
            :disabled="openingChat"
            @click="openChat"
          >
            <AppIcon
              name="message"
              :size="17"
            />
            {{ openingChat ? t('common.status.loading') : t('common.action.openChat') }}
          </button>
          <button
            type="button"
            class="detail-action"
            :disabled="loadingContact"
            @click="revealContact"
          >
            <AppIcon
              name="phone"
              :size="17"
            />
            {{ loadingContact ? t('common.status.loading') : t('common.action.revealContact') }}
          </button>
        </div>
        <dl
          v-if="Object.keys(contactPayload).length > 0"
          class="detail-spec-grid mt-4"
        >
          <div
            v-for="(value, key) in contactPayload"
            :key="key"
          >
            <dt>{{ key }}</dt>
            <dd>{{ value }}</dd>
          </div>
        </dl>
      </section>

      <section class="detail-panel">
        <p class="detail-panel__label">{{ t('marketplace.detail.sellerOtherItems') }}</p>
        <div class="detail-mini-list">
          <article
            v-for="image in galleryImages.slice(1, 3)"
            :key="image.id"
            class="detail-mini-card"
          >
            <img
              :src="image.url"
              :alt="image.alt"
            >
            <div>
              <h3>{{ listing.title }}</h3>
              <p>{{ listingPrice }}</p>
            </div>
          </article>
        </div>
      </section>

      <section class="detail-safety">
        <AppIcon
          name="shield"
          :size="20"
        />
        <p>{{ t('marketplace.detail.safetyNote') }}</p>
      </section>
    </aside>
  </main>
</template>

<style scoped>
.marketplace-detail-page {
  display: grid;
  width: 100%;
  max-width: var(--layout-page-max-width);
  min-height: calc(100vh - var(--app-header-offset, 0rem));
  gap: 2rem;
  margin: 0 auto;
  padding: 4rem var(--layout-page-padding-inline);
  color: rgb(var(--color-text));
}

.detail-main {
  display: grid;
  gap: 2rem;
}

.detail-gallery,
.detail-content,
.detail-panel,
.detail-safety {
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.75rem;
  background: rgb(var(--color-surface-raised));
  box-shadow: 0 4px 24px rgb(0 0 0 / 0.03);
}

.detail-gallery {
  overflow: hidden;
}

.detail-gallery__hero {
  aspect-ratio: 4 / 3;
  background: rgb(var(--color-surface-muted));
}

.detail-gallery__hero img,
.detail-gallery__thumb img,
.detail-mini-card img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.detail-gallery__thumbs {
  display: grid;
  gap: 0.75rem;
  padding: 1rem;
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.detail-gallery__thumb {
  aspect-ratio: 1;
  overflow: hidden;
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.5rem;
  background: rgb(var(--color-surface-muted));
}

.detail-content {
  padding: 1.5rem;
}

.detail-kicker,
.detail-panel__label {
  color: rgb(var(--color-text-muted));
  font-size: 0.75rem;
  font-weight: 600;
  letter-spacing: 0.1em;
  line-height: 1;
  text-transform: uppercase;
}

.detail-heading {
  display: grid;
  gap: 1rem;
  margin-top: 1rem;
}

.detail-heading h1 {
  margin: 0;
  font-family: var(--font-display);
  font-size: clamp(2rem, 4vw, 3rem);
  font-weight: 500;
  line-height: 1.08;
}

.detail-heading p {
  margin-top: 0.75rem;
  color: rgb(var(--color-text-muted));
  font-size: 1rem;
  line-height: 1.7;
}

.detail-price {
  margin: 0;
  font-family: var(--font-display);
  font-size: clamp(2rem, 4vw, 3.25rem);
  font-weight: 600;
  color: rgb(var(--color-primary));
}

.detail-section {
  margin-top: 1.5rem;
  border-top: 1px solid rgb(var(--color-border));
  padding-top: 1.5rem;
}

.detail-section h2 {
  margin: 0 0 1rem;
  color: rgb(var(--color-text-muted));
  font-size: 0.75rem;
  font-weight: 600;
  letter-spacing: 0.1em;
  line-height: 1;
  text-transform: uppercase;
}

.detail-section p {
  color: rgb(var(--color-text-muted));
  line-height: 1.8;
}

.detail-spec-grid {
  display: grid;
  gap: 1rem;
}

.detail-spec-grid dt {
  color: rgb(var(--color-text-muted));
  font-size: 0.75rem;
  letter-spacing: 0.1em;
  text-transform: uppercase;
}

.detail-spec-grid dd {
  margin: 0.5rem 0 0;
  color: rgb(var(--color-text));
  font-weight: 600;
}

.detail-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
}

.detail-tags span {
  border: 1px solid rgb(var(--color-border));
  border-radius: 999px;
  padding: 0.5rem 0.9rem;
  color: rgb(var(--color-text-muted));
  font-size: 0.875rem;
  font-weight: 600;
}

.detail-sidebar {
  display: grid;
  align-content: start;
  gap: 1rem;
}

.detail-panel,
.detail-safety {
  padding: 1.25rem;
}

.detail-seller-profile {
  display: flex;
  align-items: flex-start;
  gap: 0.85rem;
  margin-top: 0.9rem;
}

.detail-panel h2 {
  margin: 0;
  font-family: var(--font-display);
  font-size: 1.6rem;
  color: rgb(var(--color-text));
}

.detail-panel p {
  margin-top: 0.55rem;
  color: rgb(var(--color-text-muted));
  line-height: 1.65;
}

.detail-actions {
  display: grid;
  gap: 0.75rem;
  margin-top: 1.25rem;
}

.detail-action {
  display: inline-flex;
  min-height: 2.75rem;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  border: 1px solid rgb(var(--color-border));
  border-radius: 999px;
  color: rgb(var(--color-text-muted));
  font-size: 0.875rem;
  font-weight: 600;
}

.detail-action--primary {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary));
  color: rgb(var(--color-primary-contrast));
}

.detail-mini-list {
  display: grid;
  gap: 0.9rem;
  margin-top: 1rem;
}

.detail-mini-card {
  display: grid;
  gap: 0.8rem;
  grid-template-columns: 4.5rem 1fr;
}

.detail-mini-card img {
  aspect-ratio: 1;
  border-radius: 0.5rem;
}

.detail-mini-card h3,
.detail-mini-card p {
  margin: 0;
}

.detail-mini-card h3 {
  color: rgb(var(--color-text));
  font-size: 0.95rem;
}

.detail-mini-card p {
  margin-top: 0.4rem;
  color: rgb(var(--color-primary));
  font-weight: 700;
}

.detail-safety {
  display: grid;
  gap: 0.75rem;
  color: rgb(var(--color-text-muted));
  line-height: 1.65;
  grid-template-columns: auto 1fr;
}

@media (min-width: 768px) {
  .detail-spec-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (min-width: 1024px) {
  .marketplace-detail-page {
    grid-template-columns: minmax(0, 1fr) 21rem;
  }

  .detail-main {
    grid-template-columns: minmax(0, 0.9fr) minmax(0, 1.1fr);
    align-items: start;
  }

  .detail-sidebar {
    position: sticky;
    top: calc(var(--app-header-offset, 0rem) + 2rem);
  }
}

@media (max-width: 767px) {
  .marketplace-detail-page {
    padding: 2.5rem var(--layout-page-padding-inline);
  }
}
</style>

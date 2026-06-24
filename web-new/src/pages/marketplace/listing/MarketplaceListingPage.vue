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
  selectedImageIndex,
  isFavorited,
  t,
  toggleFavorite,
  updatingFavorite,
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
            v-for="(image, index) in galleryImages"
            :key="image.id"
            type="button"
            class="detail-gallery__thumb"
            :class="{ 'detail-gallery__thumb--active': selectedImageIndex === index }"
            @click="selectedImageIndex = index"
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
            class="detail-action"
            :class="isFavorited ? 'detail-action--saved' : ''"
            :disabled="updatingFavorite"
            @click="toggleFavorite"
          >
            <AppIcon
              name="star"
              :size="17"
            />
            {{ updatingFavorite ? t('common.status.loading') : isFavorited ? t('marketplace.detail.favoritedAction') : t('marketplace.detail.favoriteAction') }}
          </button>
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
  gap: 16px;
  margin: 0 auto;
  padding: 24px var(--layout-page-padding-inline) 72px;
  color: rgb(var(--color-text));
}

.detail-main {
  display: grid;
  gap: 16px;
}

.detail-gallery,
.detail-content,
.detail-panel,
.detail-safety {
  border: 1px solid rgb(var(--color-border));
  border-radius: 2px;
  background: rgb(var(--color-surface));
  box-shadow: none;
}

.detail-gallery {
  overflow: hidden;
}

.detail-gallery__hero {
  aspect-ratio: 1 / 1;
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
  gap: 8px;
  padding: 10px;
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.detail-gallery__thumb {
  aspect-ratio: 1;
  overflow: hidden;
  border: 1px solid rgb(var(--color-border));
  border-radius: 2px;
  background: rgb(var(--color-surface-muted));
  cursor: pointer;
  transition: border-color 0.16s ease, background-color 0.16s ease;
}

.detail-gallery__thumb:hover,
.detail-gallery__thumb--active {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary-soft));
}

.detail-content {
  padding: 18px;
}

.detail-kicker,
.detail-panel__label {
  color: rgb(var(--color-text-muted));
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0.16em;
  line-height: 1;
  text-transform: uppercase;
}

.detail-heading {
  display: grid;
  gap: 12px;
  margin-top: 12px;
  padding-bottom: 16px;
  border-bottom: 1px solid rgb(var(--color-border));
}

.detail-heading h1 {
  margin: 0;
  font-family: var(--font-sans);
  font-size: 28px;
  font-weight: 500;
  line-height: 1.2;
}

.detail-heading p {
  margin: 8px 0 0;
  color: rgb(var(--color-text-muted));
  font-size: 12px;
  line-height: 1.8;
}

.detail-price {
  margin: 0;
  font-family: var(--font-sans);
  font-size: 22px;
  font-weight: 600;
  color: rgb(var(--color-primary));
}

.detail-section {
  margin-top: 18px;
  padding-top: 18px;
  border-top: 1px solid rgb(var(--color-border));
}

.detail-section h2 {
  margin: 0 0 10px;
  color: rgb(var(--color-text));
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0.16em;
  line-height: 1;
  text-transform: uppercase;
}

.detail-section p {
  color: rgb(var(--color-text));
  font-size: 12px;
  line-height: 1.75;
}

.detail-spec-grid {
  display: grid;
  gap: 12px;
}

.detail-spec-grid dt {
  color: rgb(var(--color-text-muted));
  font-size: 10px;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.detail-spec-grid dd {
  margin: 4px 0 0;
  color: rgb(var(--color-text));
  font-size: 12px;
  font-weight: 600;
}

.detail-spec-grid > div {
  min-width: 0;
  border: 1px solid rgb(var(--color-border));
  border-radius: 2px;
  background: rgb(var(--color-surface-muted));
  padding: 10px 12px;
}

.detail-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.detail-tags span {
  border: 1px solid rgb(var(--color-border));
  border-radius: 2px;
  padding: 5px 8px;
  color: rgb(var(--color-text-muted));
  font-size: 11px;
  font-weight: 600;
}

.detail-sidebar {
  display: grid;
  align-content: start;
  gap: 12px;
}

.detail-panel,
.detail-safety {
  padding: 16px;
}

.detail-seller-profile {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  margin-top: 12px;
}

.detail-panel h2 {
  margin: 0;
  font-family: var(--font-sans);
  font-size: 14px;
  font-weight: 600;
  color: rgb(var(--color-text));
}

.detail-panel p {
  margin-top: 5px;
  color: rgb(var(--color-text-muted));
  font-size: 12px;
  line-height: 1.6;
}

.detail-actions {
  display: grid;
  gap: 8px;
  margin-top: 16px;
}

.detail-action {
  display: inline-flex;
  min-height: 38px;
  align-items: center;
  justify-content: center;
  gap: 6px;
  border: 1px solid rgb(var(--color-border));
  border-radius: 2px;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text-muted));
  font-size: 12px;
  font-weight: 600;
  transition: border-color 0.16s ease, color 0.16s ease, background-color 0.16s ease;
}

.detail-action:hover {
  border-color: rgb(var(--color-primary));
  color: rgb(var(--color-primary));
}

.detail-action--primary {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary));
  color: rgb(var(--color-primary-contrast));
}

.detail-action--saved {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary-soft));
  color: rgb(var(--color-primary));
}

.detail-mini-list {
  display: grid;
  gap: 10px;
  margin-top: 12px;
}

.detail-mini-card {
  display: grid;
  gap: 8px;
  border: 1px solid rgb(var(--color-border));
  border-radius: 2px;
  padding: 8px;
  grid-template-columns: 58px 1fr;
}

.detail-mini-card img {
  aspect-ratio: 1;
  border-radius: 2px;
}

.detail-mini-card h3,
.detail-mini-card p {
  margin: 0;
}

.detail-mini-card h3 {
  color: rgb(var(--color-text));
  font-size: 12px;
  font-weight: 600;
}

.detail-mini-card p {
  margin-top: 4px;
  color: rgb(var(--color-primary));
  font-size: 12px;
  font-weight: 600;
}

.detail-safety {
  display: grid;
  gap: 8px;
  color: rgb(var(--color-text-muted));
  font-size: 12px;
  line-height: 1.6;
  grid-template-columns: auto 1fr;
}

@media (min-width: 768px) {
  .detail-spec-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .detail-heading {
    grid-template-columns: minmax(0, 1fr) auto;
    align-items: start;
  }
}

@media (min-width: 1024px) {
  .marketplace-detail-page {
    grid-template-columns: minmax(0, 1fr) 20rem;
  }

  .detail-main {
    grid-template-columns: minmax(0, 0.92fr) minmax(0, 1.08fr);
    align-items: start;
  }

  .detail-sidebar {
    position: sticky;
    top: 72px;
  }
}

@media (max-width: 767px) {
  .marketplace-detail-page {
    padding: 18px var(--layout-page-padding-inline) 96px;
  }

  .detail-price {
    font-size: 20px;
  }
}
</style>

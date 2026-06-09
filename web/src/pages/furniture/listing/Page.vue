<!--
 * 二手帖子詳情頁。
 * 1. 參考 28Hse 商品詳情資訊密度建立圖片、詳情、聯絡與賣家區布局。
 * 2. 接入真實帖子詳情、聯絡權限與聊天入口。
-->
<script setup lang="ts">
import AppIcon from '@/shared/components/base/AppIcon.vue';
import BaseAvatar from '@/shared/components/base/BaseAvatar.vue';
import AppBreadcrumb from '@/shared/components/navigation/AppBreadcrumb.vue';

import { useFurnitureListingPage } from './listing';

const {
  categoryLabel,
  conditionLabel,
  contactRows,
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
  revealContactLabel,
  revealContact,
  isFavorited,
  t,
  toggleFavorite,
  updatingFavorite,
} = useFurnitureListingPage();
</script>

<template>
  <main class="marketplace-detail-page">
    <AppBreadcrumb
      class="marketplace-detail-breadcrumb"
      :items="[
        { label: t('nav.home'), to: '/' },
        { label: t('nav.furniture'), to: '/furniture' },
        { label: listing?.title || t('marketplace.detail.title') },
      ]"
    />
    <section
      v-if="loading"
      class="detail-panel"
    >
      {{ t('common.status.loading') }}
    </section>

    <template v-else-if="listing">
      <section class="detail-main">
        <section class="detail-hero">
          <div class="detail-heading">
            <div class="detail-kicker">
              {{ t('marketplace.detail.referenceId', { id: listingId }) }}
            </div>
            <h1>{{ listing.title }}</h1>
            <p v-if="listing.summary">{{ listing.summary }}</p>
            <p class="detail-price">{{ listingPrice }}</p>
          </div>

          <div class="detail-gallery">
            <div class="detail-gallery__hero">
              <img
                v-if="coverImage"
                :src="coverImage.url"
                :alt="coverImage.alt"
              >
              <div
                v-else
                class="detail-gallery__placeholder"
              >
                <AppIcon
                  name="picture"
                  :size="44"
                />
              </div>
            </div>
            <div
              v-if="galleryImages.length > 1"
              class="detail-gallery__thumbs"
            >
              <button
                v-for="image in galleryImages.slice(1, 5)"
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
        </section>

        <article class="detail-content">
          <section class="detail-section detail-section--plain">
            <h2>{{ t('marketplace.detail.description') }}</h2>
            <p>
              {{ listing.description }}
            </p>
          </section>

          <section class="detail-section">
            <h2>{{ t('marketplace.detail.specification') }}</h2>
            <div class="detail-info-grid">
              <div class="detail-info-card">
                <span>{{ t('marketplace.mine.category') }}</span>
                <strong>{{ categoryLabel }}</strong>
              </div>
              <div class="detail-info-card">
                <span>{{ t('common.label.condition') }}</span>
                <strong>{{ conditionLabel }}</strong>
              </div>
              <div class="detail-info-card">
                <span>{{ t('marketplace.detail.publishedAt') }}</span>
                <strong>{{ publishedAt }}</strong>
              </div>
              <div class="detail-info-card">
                <span>{{ t('marketplace.editor.dimensionField') }}</span>
                <strong>{{ listing.dimension_text || '-' }}</strong>
              </div>
            </div>
          </section>

          <section class="detail-section">
            <h2>{{ t('marketplace.detail.tradeInfo') }}</h2>
            <div class="detail-trade-list">
              <div>
                <span>{{ t('marketplace.detail.pickupArea') }}</span>
                <strong>{{ districtLabel || '-' }}</strong>
              </div>
              <div v-if="listing.pickup_location_text">
                <span>{{ t('marketplace.editor.tradeNote') }}</span>
                <strong>{{ listing.pickup_location_text }}</strong>
              </div>
              <div v-if="listing.delivery_tags.length > 0">
                <span>{{ t('marketplace.editor.deliveryTags') }}</span>
                <div class="detail-tags">
                  <span
                    v-for="tag in listing.delivery_tags"
                    :key="tag"
                  >
                    {{ formatDeliveryTag(tag) }}
                  </span>
                </div>
              </div>
              <div>
                <span>{{ t('marketplace.editor.visibilityTitle') }}</span>
                <strong>{{ listing.visibility_scope === 'public' ? t('marketplace.detail.publicListing') : t('common.state.buildingOnly') }}</strong>
              </div>
            </div>
          </section>
        </article>

        <section class="detail-safety">
          <AppIcon
            name="shield"
            :size="20"
          />
          <p>{{ t('marketplace.detail.safetyNote') }}</p>
        </section>
      </section>

      <aside class="detail-sidebar">
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
              {{ loadingContact ? t('common.status.loading') : revealContactLabel }}
            </button>
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
          </div>
          <div
            v-if="contactRows.length > 0"
            class="detail-contact-actions"
          >
            <a
              v-for="row in contactRows"
              :key="row.key"
              class="detail-contact-action"
              :class="row.isIconOnly ? 'detail-contact-action--icon' : ''"
              :href="row.href"
              target="_blank"
              rel="noopener noreferrer"
              :aria-label="row.label"
            >
              <AppIcon
                :name="row.icon"
                :size="17"
              />
              <span>{{ row.displayText || row.value }}</span>
            </a>
          </div>
        </section>
      </aside>
    </template>
  </main>
</template>

<style scoped>
.marketplace-detail-page {
  display: grid;
  width: 100%;
  max-width: var(--layout-page-max-width);
  min-height: calc(100vh - var(--app-header-offset, 0rem));
  gap: 1.5rem;
  margin: 0 auto;
  padding: 3rem var(--layout-page-padding-inline) 5rem;
  color: rgb(var(--color-text));
}

.detail-main {
  display: grid;
  align-content: start;
  gap: 1.5rem;
  min-width: 0;
}

.detail-gallery,
.detail-hero,
.detail-content,
.detail-panel,
.detail-safety {
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.5rem;
  background: rgb(var(--color-surface));
  box-shadow: 0 16px 40px rgb(15 23 42 / 0.06);
}

.detail-hero {
  display: grid;
  gap: 1.25rem;
  overflow: hidden;
  padding: 1.25rem;
}

.detail-gallery {
  overflow: hidden;
}

.detail-gallery__hero {
  aspect-ratio: 16 / 9;
  background: rgb(var(--color-surface-muted));
}

.detail-gallery__hero img,
.detail-gallery__thumb img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.detail-gallery__placeholder {
  display: grid;
  height: 100%;
  place-items: center;
  color: rgb(var(--color-text-muted));
}

.detail-gallery__thumbs {
  display: grid;
  gap: 0.25rem;
  padding: 0.25rem;
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.detail-gallery__thumb {
  aspect-ratio: 4 / 3;
  overflow: hidden;
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.25rem;
  background: rgb(var(--color-surface-muted));
}

.detail-content {
  overflow: hidden;
  padding: 1.25rem;
}

.detail-kicker,
.detail-panel__label {
  color: rgb(var(--color-primary));
  font-size: 0.75rem;
  font-weight: 900;
  letter-spacing: 0.12em;
  line-height: 1;
  text-transform: uppercase;
}

.detail-heading {
  display: grid;
  align-content: start;
  gap: 0.9rem;
  min-width: 0;
}

.detail-heading h1 {
  margin: 0;
  font-family: var(--font-display);
  font-size: clamp(2.1rem, 4vw, 3.4rem);
  font-weight: 700;
  line-height: 1.04;
}

.detail-heading p {
  margin: 0;
  color: rgb(var(--color-text-muted));
  font-size: 1rem;
  line-height: 1.7;
}

.detail-price {
  margin: 0;
  font-family: var(--font-display);
  font-size: clamp(1.9rem, 4vw, 2.8rem);
  font-weight: 800;
  color: rgb(var(--color-primary));
}

.detail-section {
  margin-top: 1.5rem;
  border-top: 1px solid rgb(var(--color-border));
  padding-top: 1.25rem;
}

.detail-section--plain {
  margin-top: 0;
  border-top: 0;
  padding-top: 0;
}

.detail-section h2 {
  margin: 0 0 0.9rem;
  color: rgb(var(--color-text));
  font-size: 1rem;
  font-weight: 900;
  line-height: 1.25;
}

.detail-section p {
  margin: 0;
  color: rgb(var(--color-text-muted));
  line-height: 1.8;
}

.detail-info-grid {
  display: grid;
  gap: 0.75rem;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.detail-info-card {
  display: grid;
  min-height: 5.25rem;
  align-content: center;
  gap: 0.4rem;
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.55rem;
  background: rgb(var(--color-surface-raised));
  padding: 0.95rem;
}

.detail-info-card span,
.detail-trade-list > div > span {
  color: rgb(var(--color-text-muted));
  font-size: 0.76rem;
  font-weight: 900;
  line-height: 1.45;
}

.detail-info-card strong {
  min-width: 0;
  color: rgb(var(--color-text));
  font-size: 1rem;
  font-weight: 900;
  line-height: 1.35;
  overflow-wrap: anywhere;
}

.detail-trade-list {
  display: grid;
  overflow: hidden;
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.6rem;
  background: rgb(var(--color-surface));
}

.detail-trade-list > div {
  display: grid;
  gap: 0.35rem;
  padding: 0.95rem 1rem;
}

.detail-trade-list > div + div {
  border-top: 1px solid rgb(var(--color-border));
}

.detail-trade-list > div > strong {
  color: rgb(var(--color-text));
  font-size: 0.95rem;
  font-weight: 900;
  line-height: 1.45;
  overflow-wrap: anywhere;
}

.detail-contact-actions {
  display: grid;
  gap: 0.65rem;
  margin-top: 1rem;
}

.detail-contact-action {
  display: inline-flex;
  min-height: 2.55rem;
  align-items: center;
  justify-content: center;
  gap: 0.45rem;
  border: 1px solid rgb(var(--color-primary));
  border-radius: 0.45rem;
  background: rgb(var(--color-primary));
  padding: 0 0.9rem;
  color: rgb(var(--color-primary-contrast));
  font-size: 0.85rem;
  font-weight: 800;
  text-decoration: none;
  overflow-wrap: anywhere;
}

.detail-contact-action--icon {
  width: 100%;
  padding: 0 0.9rem;
}

.detail-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
}

.detail-tags span {
  border: 1px solid rgb(var(--color-border));
  border-radius: 999px;
  background: rgb(var(--color-surface-raised));
  padding: 0.5rem 0.9rem;
  color: rgb(var(--color-text-muted));
  font-size: 0.875rem;
  font-weight: 850;
}

.detail-sidebar {
  display: grid;
  align-content: start;
  gap: 1.5rem;
  min-width: 0;
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
  font-size: 1.45rem;
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
  border-radius: 0.45rem;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text-muted));
  font-size: 0.875rem;
  font-weight: 850;
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

.detail-safety {
  display: grid;
  gap: 0.75rem;
  color: rgb(var(--color-text-muted));
  line-height: 1.65;
  grid-template-columns: auto 1fr;
}

@media (min-width: 768px) {
  .detail-trade-list > div {
    grid-template-columns: minmax(8rem, 0.35fr) minmax(0, 1fr);
    align-items: center;
  }
}

@media (min-width: 1024px) {
  .marketplace-detail-page {
    grid-template-columns: minmax(0, 1fr) 22rem;
    align-items: start;
  }

  .marketplace-detail-breadcrumb {
    grid-column: 1 / -1;
  }

  .detail-sidebar {
    position: sticky;
    top: calc(var(--app-header-offset, 0rem) + 2rem);
  }
}

@media (max-width: 767px) {
  .marketplace-detail-page {
    gap: 1rem;
    padding: 1.5rem var(--layout-page-padding-inline) 3rem;
  }

  .detail-hero,
  .detail-content,
  .detail-panel,
  .detail-safety {
    padding: 0.95rem;
  }

  .detail-heading h1 {
    font-size: clamp(2rem, 12vw, 3rem);
  }

  .detail-gallery {
    order: -1;
  }

  .detail-gallery__thumbs {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .detail-info-grid {
    grid-template-columns: 1fr;
  }

  .detail-trade-list > div {
    padding: 0.85rem;
  }
}
</style>

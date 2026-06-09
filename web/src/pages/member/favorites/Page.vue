<!--
 * 二手交易商品收藏頁。
 * 1. 查詢會員已收藏的二手帖子。
 * 2. 支援取消收藏並保持列表即時刷新。
-->
<script setup lang="ts">
import axios from 'axios';
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { RouterLink } from 'vue-router';

import {
  fetchMyFavoriteSecondhandListings,
  unfavoriteSecondhandListing,
} from '@/domains/marketplace/api';
import type { SecondhandListingSummaryResponse } from '@/domains/marketplace/model';
import AppIcon from '@/shared/components/base/AppIcon.vue';
import BaseEmpty from '@/shared/components/feedback/BaseEmpty.vue';
import { useFeedbackStore } from '@/app/stores/feedback';
import { usePreferenceStore } from '@/app/stores/preferences';
import { formatDate, formatPrice } from '@/shared/utils/format';
import {
  resolveListingCommunityName,
  resolveListingCoverImage,
} from '@/shared/utils/marketplace';

const { t } = useI18n();
const feedbackStore = useFeedbackStore();
const preferenceStore = usePreferenceStore();
const loading = ref(false);
const removingIds = ref<string[]>([]);
const items = ref<SecondhandListingSummaryResponse[]>([]);
const page = ref(1);
const pageSize = 12;
const total = ref(0);
const hasNext = computed(() => page.value * pageSize < total.value);
const hasPrevious = computed(() => page.value > 1);

// 1. 載入收藏列表
const loadFavorites = async (targetPage = page.value): Promise<void> => {
  loading.value = true;

  try {
    const { data } = await fetchMyFavoriteSecondhandListings({
      page: targetPage,
      page_size: pageSize,
    });
    items.value = data.data.items;
    page.value = data.data.pagination.page;
    total.value = data.data.pagination.total;
  } catch (error) {
    feedbackStore.pushToast(
      axios.isAxiosError(error)
        ? error.response?.data?.message ?? t('marketplace.myHub.favoritesLoadError')
        : t('marketplace.myHub.favoritesLoadError'),
      'error',
    );
  } finally {
    loading.value = false;
  }
};

// 2. 取消收藏
const removeFavorite = async (listingID: string): Promise<void> => {
  if (removingIds.value.includes(listingID)) {
    return;
  }

  removingIds.value = [...removingIds.value, listingID];

  try {
    await unfavoriteSecondhandListing(listingID);
    feedbackStore.pushToast(t('marketplace.detail.favoriteRemoved'), 'success');
    await loadFavorites(page.value);
  } catch (error) {
    feedbackStore.pushToast(
      axios.isAxiosError(error)
        ? error.response?.data?.message ?? t('marketplace.detail.favoriteError')
        : t('marketplace.detail.favoriteError'),
      'error',
    );
  } finally {
    removingIds.value = removingIds.value.filter((id) => id !== listingID);
  }
};

// 3. 格式化收藏卡片價格
const formatListingPrice = (listing: SecondhandListingSummaryResponse): string =>
  listing.price_mode === 'free'
    ? t('common.price.free')
    : formatPrice(listing.price_hkd ?? 0, preferenceStore.locale);

// 4. 判斷帖子是否正在取消收藏
const isRemoving = (listingID: string): boolean => removingIds.value.includes(listingID);

onMounted(() => {
  void loadFavorites(1);
});
</script>

<template>
  <section class="favorite-page">
    <header class="favorite-header">
      <div>
        <p class="favorite-kicker">{{ t('marketplace.myHub.favorites') }}</p>
        <h1>{{ t('marketplace.myHub.favoritesTitle') }}</h1>
        <p>{{ t('marketplace.myHub.favoritesDescription') }}</p>
      </div>
    </header>

    <div
      v-if="loading"
      class="favorite-loading"
    >
      {{ t('common.status.loading') }}
    </div>

    <BaseEmpty
      v-else-if="items.length === 0"
      :title="t('marketplace.myHub.favoritesEmptyTitle')"
      :description="t('marketplace.myHub.favoritesEmptyDescription')"
    >
      <RouterLink
        to="/furniture"
        class="favorite-empty-action"
      >
        {{ t('common.action.browse') }}
      </RouterLink>
    </BaseEmpty>

    <div
      v-else
      class="favorite-grid"
    >
      <article
        v-for="listing in items"
        :key="listing.listing_id"
        class="favorite-card"
      >
        <RouterLink
          :to="`/furniture/listing/${listing.listing_id}`"
          class="favorite-card__media"
        >
          <img
            v-if="resolveListingCoverImage(listing)"
            :src="resolveListingCoverImage(listing)?.url"
            :alt="listing.title"
          />
          <span v-else>
            <AppIcon
              name="picture"
              :size="38"
            />
          </span>
        </RouterLink>
        <div class="favorite-card__body">
          <div>
            <RouterLink
              :to="`/furniture/listing/${listing.listing_id}`"
              class="favorite-card__title"
            >
              {{ listing.title }}
            </RouterLink>
            <p>{{ listing.summary }}</p>
          </div>
          <dl class="favorite-card__meta">
            <div>
              <dt>{{ t('common.label.price') }}</dt>
              <dd>{{ formatListingPrice(listing) }}</dd>
            </div>
            <div>
              <dt>{{ t('common.label.community') }}</dt>
              <dd>{{ resolveListingCommunityName(listing) }}</dd>
            </div>
            <div>
              <dt>{{ t('common.label.publishedAt') }}</dt>
              <dd>{{ formatDate(listing.published_at || listing.updated_at, preferenceStore.locale) }}</dd>
            </div>
          </dl>
          <button
            type="button"
            class="favorite-remove"
            :disabled="isRemoving(listing.listing_id)"
            @click="removeFavorite(listing.listing_id)"
          >
            <AppIcon
              name="star"
              :size="16"
            />
            {{ isRemoving(listing.listing_id) ? t('common.status.loading') : t('marketplace.detail.favoritedAction') }}
          </button>
        </div>
      </article>
    </div>

    <footer
      v-if="items.length > 0"
      class="favorite-pagination"
    >
      <button
        type="button"
        :disabled="!hasPrevious || loading"
        @click="loadFavorites(page - 1)"
      >
        {{ t('marketplace.management.previous') }}
      </button>
      <span>{{ page }}</span>
      <button
        type="button"
        :disabled="!hasNext || loading"
        @click="loadFavorites(page + 1)"
      >
        {{ t('marketplace.management.next') }}
      </button>
    </footer>
  </section>
</template>

<style scoped>
.favorite-page {
  display: grid;
  gap: 1.5rem;
}

.favorite-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
}

.favorite-kicker {
  margin: 0;
  color: rgb(var(--color-text-muted));
  font-size: 0.75rem;
  font-weight: 800;
  letter-spacing: 0.1em;
  line-height: 1;
  text-transform: uppercase;
}

.favorite-header h1 {
  margin: 0.35rem 0 0;
  color: rgb(var(--color-text));
  font-family: var(--font-display);
  font-size: clamp(1.6rem, 2.4vw, 2.25rem);
  font-weight: 650;
  line-height: 1.2;
}

.favorite-header p:not(.favorite-kicker) {
  margin: 0.45rem 0 0;
  color: rgb(var(--color-text-muted));
  font-size: 0.92rem;
  line-height: 1.6;
}

.favorite-loading {
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface-raised));
  padding: 2rem;
  color: rgb(var(--color-text-muted));
  font-weight: 650;
  text-align: center;
}

.favorite-empty-action,
.favorite-pagination button,
.favorite-remove {
  display: inline-flex;
  min-height: 2.5rem;
  align-items: center;
  justify-content: center;
  gap: 0.45rem;
  border: 1px solid rgb(var(--color-primary));
  border-radius: 8px;
  background: rgb(var(--color-primary));
  padding: 0 0.95rem;
  color: rgb(var(--color-primary-contrast));
  font-size: 0.88rem;
  font-weight: 750;
  line-height: 1;
  text-decoration: none;
}

.favorite-grid {
  display: grid;
  gap: 1rem;
}

.favorite-card {
  display: grid;
  grid-template-columns: 13rem minmax(0, 1fr);
  overflow: hidden;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface-raised));
}

.favorite-card__media {
  display: block;
  min-height: 10rem;
  background: rgb(var(--color-surface-muted));
  color: rgb(var(--color-text-muted));
}

.favorite-card__media img,
.favorite-card__media span {
  width: 100%;
  height: 100%;
}

.favorite-card__media img {
  display: block;
  object-fit: cover;
}

.favorite-card__media span {
  display: flex;
  align-items: center;
  justify-content: center;
}

.favorite-card__body {
  display: grid;
  gap: 1rem;
  padding: 1rem;
}

.favorite-card__title {
  display: block;
  color: rgb(var(--color-text));
  font-family: var(--font-display);
  font-size: 1.15rem;
  font-weight: 650;
  line-height: 1.3;
  text-decoration: none;
}

.favorite-card__body p {
  display: -webkit-box;
  overflow: hidden;
  margin: 0.4rem 0 0;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  color: rgb(var(--color-text-muted));
  font-size: 0.9rem;
  line-height: 1.55;
}

.favorite-card__meta {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.75rem;
  margin: 0;
}

.favorite-card__meta dt {
  color: rgb(var(--color-text-muted));
  font-size: 0.72rem;
  font-weight: 800;
  letter-spacing: 0.08em;
  line-height: 1.2;
  text-transform: uppercase;
}

.favorite-card__meta dd {
  overflow: hidden;
  margin: 0.25rem 0 0;
  color: rgb(var(--color-text));
  font-size: 0.9rem;
  font-weight: 700;
  line-height: 1.35;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.favorite-remove {
  justify-self: start;
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary-soft));
  color: rgb(var(--color-primary));
}

.favorite-remove:disabled,
.favorite-pagination button:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

.favorite-pagination {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 0.65rem;
}

.favorite-pagination span {
  color: rgb(var(--color-text-muted));
  font-weight: 750;
}

@media (max-width: 767px) {
  .favorite-card {
    grid-template-columns: 1fr;
  }

  .favorite-card__meta {
    grid-template-columns: 1fr;
  }
}
</style>

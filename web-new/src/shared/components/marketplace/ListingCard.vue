<!--
 * 二手帖子卡片。
 * 1. 承接首頁與列表頁的帖子展示。
 * 2. 統一圖片、價格、狀態與行動按鈕排版。
-->
<script setup lang="ts">
import { computed } from 'vue';
import { RouterLink } from 'vue-router';
import { useI18n } from 'vue-i18n';

import type { Listing } from '@/model/listing';
import type { MarketplaceListingLike } from '@/model/marketplace';
import {
  resolveListingCoverImage,
  resolveListingId,
  resolveListingPrice,
  resolveListingSummary,
  resolveListingTitle,
} from '@/utils/marketplace';
import { usePreferenceStore } from '@/stores/preferences';
import { formatPrice } from '@/utils/format';

interface ListingCardProps {
  listing: Listing | MarketplaceListingLike;
}

const props = defineProps<ListingCardProps>();
const { t } = useI18n();
const preferenceStore = usePreferenceStore();

// 1. 取得封面圖片
const coverImage = computed(() => resolveListingCoverImage(props.listing));

// 2. 格式化價格
const listingPrice = computed(() =>
  formatPrice(resolveListingPrice(props.listing), preferenceStore.locale),
);

// 3. 取得帖子摘要
const listingSummary = computed(() => resolveListingSummary(props.listing));
</script>

<template>
  <RouterLink :to="`/marketplace/listing/${resolveListingId(props.listing)}`">
    <article class="marketplace-post-card group relative h-[22rem] overflow-hidden rounded-feature">
      <img
        v-if="coverImage"
        :src="coverImage.url"
        :alt="coverImage.alt"
        class="absolute inset-0 h-full w-full object-cover transition duration-500 group-hover:scale-[1.04]"
      />
      <div
        v-else
        class="absolute inset-0 flex items-center justify-center bg-surface-raised text-sm text-text-muted"
      >
        {{ t('marketplace.mine.noCover') }}
      </div>
      <div class="marketplace-post-card__shade" />
      <div class="marketplace-post-card__copy">
        <h3 class="line-clamp-2 text-xl font-semibold leading-snug text-white">
          {{ resolveListingTitle(props.listing) }}
        </h3>
        <p class="line-clamp-3 text-sm leading-6 text-white/78">
          {{ listingSummary }}
        </p>
        <p class="font-display text-3xl text-white">
          {{ listingPrice }}
        </p>
      </div>
    </article>
  </RouterLink>
</template>

<style scoped>
.marketplace-post-card {
  border: 1px solid rgb(var(--color-border) / 0.58);
  background: rgb(var(--color-surface-raised));
  transition:
    transform 0.24s ease,
    box-shadow 0.24s ease;
}

.marketplace-post-card:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-floating);
}

.marketplace-post-card__shade {
  position: absolute;
  inset: 0;
  background: linear-gradient(180deg, rgb(0 0 0 / 0.08), rgb(0 0 0 / 0.28) 45%, rgb(0 0 0 / 0.72));
}

.marketplace-post-card__copy {
  position: absolute;
  inset-inline: 0;
  bottom: 0;
  display: flex;
  min-height: 12.5rem;
  flex-direction: column;
  gap: 0.75rem;
  justify-content: flex-end;
  padding: 1.1rem;
}
</style>

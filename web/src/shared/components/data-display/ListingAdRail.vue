<!--
 * 列表頁右側展示廣告欄。
 * 1. 按頻道讀取已配置的 10 個 listing_side 廣告位。
 * 2. 上方顯示 slot 1-3，下方顯示 slot 4-10。
-->
<script setup lang="ts">
import axios from 'axios';
import { computed, onMounted, ref, watch } from 'vue';

import { fetchPublicDisplayAds } from '@/domains/payments/api';
import type { PublicDisplayAdResponse } from '@/domains/payments/model';
import AppIcon from '@/shared/components/base/AppIcon.vue';

type ListingAdChannel = 'property_sale' | 'serviced_apartment' | 'furniture';

const props = withDefaults(defineProps<{
  channel: ListingAdChannel;
  variant?: 'top' | 'bottom';
}>(), {
  variant: 'top',
});

const loading = ref(false);
const ads = ref<PublicDisplayAdResponse[]>([]);

const visibleAds = computed(() =>
  props.variant === 'top'
    ? ads.value.filter((ad) => ad.slot_index >= 1 && ad.slot_index <= 3)
    : ads.value.filter((ad) => ad.slot_index >= 4 && ad.slot_index <= 10),
);

// 1. 讀取公開展示廣告
const loadAds = async (): Promise<void> => {
  loading.value = true;
  try {
    const { data } = await fetchPublicDisplayAds({
      channel: props.channel,
      placement: 'listing_side',
      limit: 10,
    });
    ads.value = data.data.items;
  } catch (error) {
    if (axios.isAxiosError(error)) {
      ads.value = [];
      return;
    }
    ads.value = [];
  } finally {
    loading.value = false;
  }
};

// 2. 取得卡片連結
const resolveAdHref = (ad: PublicDisplayAdResponse): string =>
  ad.target_url.trim() || '#';

// 3. 判斷是否可開啟廣告
const isAdLinkEnabled = (ad: PublicDisplayAdResponse): boolean =>
  ad.target_url.trim().length > 0;

// 4. 取得廣告媒體地址
const resolveAdMediaUrl = (ad: PublicDisplayAdResponse): string =>
  ad.media_url.trim() || ad.cover_url.trim();

// 5. 取得廣告展示文案
const resolveAdDisplayText = (ad: PublicDisplayAdResponse): string =>
  ad.display_text.trim() || ad.summary.trim();

// 6. 取得廣告展示標題
const resolveAdDisplayTitle = (ad: PublicDisplayAdResponse): string =>
  ad.display_title.trim() || ad.title.trim();

onMounted(() => {
  void loadAds();
});

watch(
  () => props.channel,
  () => {
    void loadAds();
  },
);
</script>

<template>
  <aside
    class="listing-ad-rail"
    aria-label="推廣"
  >
    <h2>推廣</h2>

    <div
      v-if="loading"
      class="listing-ad-rail__placeholder"
    >
      載入中...
    </div>

    <div
      v-else-if="visibleAds.length === 0"
      class="listing-ad-rail__placeholder"
    >
      暫未設定展示廣告
    </div>

    <template v-else>
      <a
        v-for="ad in visibleAds"
        :key="`${ad.slot_index}-${ad.task_id}`"
        class="listing-ad-card"
        :class="[
          `listing-ad-card--${ad.display_layout}`,
          variant === 'top' ? 'listing-ad-card--top' : 'listing-ad-card--bottom',
        ]"
        :href="resolveAdHref(ad)"
        :target="isAdLinkEnabled(ad) ? '_blank' : undefined"
        :rel="isAdLinkEnabled(ad) ? 'noopener noreferrer' : undefined"
        @click="!isAdLinkEnabled(ad) && $event.preventDefault()"
      >
        <div class="listing-ad-card__media">
          <img
            v-if="ad.media_type === 'image' && resolveAdMediaUrl(ad)"
            :src="resolveAdMediaUrl(ad)"
            :alt="resolveAdDisplayTitle(ad)"
          />
          <video
            v-else-if="ad.media_type === 'video' && resolveAdMediaUrl(ad)"
            :src="resolveAdMediaUrl(ad)"
            muted
            playsinline
            preload="metadata"
          />
          <div
            v-else
            class="listing-ad-card__empty-media"
          >
            <AppIcon
              name="picture"
              :size="24"
            />
          </div>
        </div>

        <div
          v-if="ad.display_layout !== 'text_compact'"
          class="listing-ad-card__body"
        >
          <span>AD</span>
          <h3>{{ resolveAdDisplayTitle(ad) }}</h3>
          <p v-if="resolveAdDisplayText(ad)">{{ resolveAdDisplayText(ad) }}</p>
        </div>
      </a>
    </template>
  </aside>
</template>

<style scoped>
.listing-ad-rail {
  display: grid;
  align-self: start;
  gap: 0.75rem;
}

.listing-ad-rail h2 {
  margin: 0;
  color: rgb(var(--color-text-muted));
  font-size: 0.75rem;
  font-weight: 700;
  letter-spacing: 0.1em;
  line-height: 1;
  text-transform: uppercase;
}

.listing-ad-rail__placeholder,
.listing-ad-card {
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface-raised));
}

.listing-ad-rail__placeholder {
  padding: 1rem;
  color: rgb(var(--color-text-muted));
  font-size: 0.82rem;
  line-height: 1.5;
}

.listing-ad-card {
  display: grid;
  overflow: hidden;
  color: inherit;
  text-decoration: none;
  transition:
    border-color 0.2s ease,
    box-shadow 0.2s ease;
}

.listing-ad-card:hover {
  border-color: rgb(var(--color-primary) / 0.35);
  box-shadow: 0 8px 28px rgb(15 23 42 / 0.08);
}

.listing-ad-card__media {
  aspect-ratio: 4 / 3;
  background: rgb(var(--color-surface-muted));
}

.listing-ad-card__media img,
.listing-ad-card__media video {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.listing-ad-card__empty-media {
  display: grid;
  height: 100%;
  place-items: center;
  color: rgb(var(--color-text-muted));
}

.listing-ad-card__body {
  display: grid;
  gap: 0.35rem;
  padding: 0.85rem;
}

.listing-ad-card__body span {
  color: rgb(var(--color-primary));
  font-size: 0.68rem;
  font-weight: 800;
  letter-spacing: 0.08em;
  line-height: 1;
}

.listing-ad-card__body h3 {
  display: -webkit-box;
  overflow: hidden;
  margin: 0;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  color: rgb(var(--color-text));
  font-size: 0.95rem;
  font-weight: 700;
  line-height: 1.35;
}

.listing-ad-card__body p {
  display: -webkit-box;
  overflow: hidden;
  margin: 0;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  color: rgb(var(--color-text-muted));
  font-size: 0.8rem;
  line-height: 1.45;
}

.listing-ad-card--image_full .listing-ad-card__body {
  padding-top: 0.7rem;
}

.listing-ad-card--image_text .listing-ad-card__media {
  height: 8rem;
  aspect-ratio: auto;
}

.listing-ad-card--text_compact .listing-ad-card__media {
  aspect-ratio: 16 / 7;
}

.listing-ad-card--bottom {
  grid-template-columns: minmax(0, 1fr);
}
</style>

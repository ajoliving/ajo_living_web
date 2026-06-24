<!--
 * 列表右側展示廣告欄。
 * 1. 依頻道讀取公開 listing_side 展示廣告。
 * 2. 在未設定廣告時保留固定廣告位佔位。
 * 3. 統一樓盤與二手列表右側廣告版面。
-->
<script setup lang="ts">
import axios from 'axios';
import { computed, onMounted, ref, watch } from 'vue';

import { fetchPublicDisplayAds } from '@/httpapis/wallet';
import type { PublicDisplayAdResponse } from '@/model/payments';

const props = defineProps<{
  channel: PublicDisplayAdResponse['display_channel'];
}>();

const ads = ref<PublicDisplayAdResponse[]>([]);
const loading = ref(false);

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
  } catch (error: unknown) {
    if (!axios.isCancel(error)) {
      ads.value = [];
    }
  } finally {
    loading.value = false;
  }
};

// 2. 取得廣告圖片
const resolveImageURL = (ad: PublicDisplayAdResponse): string =>
  ad.cover_url || ad.media_url;

// 3. 取得廣告標題與內容
const resolveTitle = (ad: PublicDisplayAdResponse): string =>
  ad.display_title || ad.title;

const resolveText = (ad: PublicDisplayAdResponse): string =>
  ad.display_text || ad.summary;

const visibleAds = computed(() => ads.value.slice(0, 3));

watch(() => props.channel, loadAds);

onMounted(() => {
  void loadAds();
});
</script>

<template>
  <aside
    class="listing-side-ads"
    aria-label="廣告"
  >
    <div
      v-if="loading"
      class="listing-side-ads__slot listing-side-ads__slot--placeholder listing-side-ads__slot--large"
    >
      <span>廣告</span>
    </div>

    <template v-else-if="visibleAds.length > 0">
      <a
        v-for="(ad, index) in visibleAds"
        :key="`${ad.task_id}-${ad.slot_index}`"
        class="listing-side-ads__slot"
        :class="[
          `listing-side-ads__slot--${ad.display_layout}`,
          { 'listing-side-ads__slot--large': index === 0 },
        ]"
        :href="ad.target_url || undefined"
        :target="ad.target_url ? '_blank' : undefined"
        rel="noopener noreferrer"
      >
        <img
          v-if="resolveImageURL(ad)"
          :src="resolveImageURL(ad)"
          :alt="resolveTitle(ad)"
        />
        <div class="listing-side-ads__body">
          <p>廣告</p>
          <h2>{{ resolveTitle(ad) }}</h2>
          <span v-if="resolveText(ad)">{{ resolveText(ad) }}</span>
        </div>
      </a>
    </template>

    <template v-else>
      <div class="listing-side-ads__slot listing-side-ads__slot--placeholder listing-side-ads__slot--large">
        <span>160 × 600</span>
      </div>
      <div class="listing-side-ads__slot listing-side-ads__slot--placeholder listing-side-ads__slot--small">
        <span>300 × 250</span>
      </div>
    </template>
  </aside>
</template>

<style scoped>
.listing-side-ads {
  display: flex;
  width: 176px;
  flex-direction: column;
  gap: 14px;
  align-items: stretch;
  border-left: 1px solid #e4e4e4;
  background: #ffffff;
  padding: 18px 10px 18px 6px;
}

.listing-side-ads__slot {
  position: relative;
  display: flex;
  min-height: 250px;
  overflow: hidden;
  flex-direction: column;
  justify-content: flex-end;
  border: 1px solid #e4e4e4;
  border-radius: 3px;
  background: #ffffff;
  color: #1a1a1a;
  text-decoration: none;
}

.listing-side-ads__slot::before {
  position: absolute;
  z-index: 2;
  top: 5px;
  right: 8px;
  color: #aaaaaa;
  content: '廣告';
  font-size: 9px;
  letter-spacing: 0.5px;
}

.listing-side-ads__slot--large {
  position: sticky;
  top: calc(var(--app-header-offset, 48px) + 12px);
  min-height: 360px;
}

.listing-side-ads__slot--text_compact {
  min-height: 118px;
}

.listing-side-ads__slot--image_text {
  min-height: 210px;
}

.listing-side-ads__slot--large.listing-side-ads__slot--text_compact,
.listing-side-ads__slot--large.listing-side-ads__slot--image_text,
.listing-side-ads__slot--large.listing-side-ads__slot--image_full {
  min-height: 360px;
}

.listing-side-ads__slot--image_full .listing-side-ads__body {
  background: linear-gradient(180deg, transparent, rgb(0 0 0 / 0.68));
  color: #ffffff;
}

.listing-side-ads__slot img {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.listing-side-ads__body {
  position: relative;
  z-index: 1;
  display: grid;
  gap: 4px;
  background: rgb(255 255 255 / 0.92);
  padding: 14px 12px;
}

.listing-side-ads__body p {
  margin: 0;
  color: rgb(var(--color-primary));
  font-size: 9px;
  font-weight: 700;
  letter-spacing: 0.8px;
}

.listing-side-ads__body h2 {
  margin: 0;
  font-size: 13px;
  font-weight: 700;
  line-height: 1.45;
}

.listing-side-ads__body span {
  display: -webkit-box;
  overflow: hidden;
  color: currentColor;
  font-size: 11px;
  line-height: 1.55;
  opacity: 0.78;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 3;
}

.listing-side-ads__slot--placeholder {
  align-items: center;
  justify-content: center;
  border-style: dashed;
  background: linear-gradient(135deg, #f6f6f6, #efefef);
  color: #aaaaaa;
  font-size: 11px;
  letter-spacing: 0.5px;
}

.listing-side-ads__slot--placeholder::before {
  content: '廣告位';
}

.listing-side-ads__slot--placeholder span {
  border: 1px dashed #cccccc;
  border-radius: 3px;
  padding: 10px 12px;
}

.listing-side-ads__slot--small {
  min-height: 250px;
}

@media (max-width: 1199px) {
  .listing-side-ads {
    display: none;
  }
}
</style>

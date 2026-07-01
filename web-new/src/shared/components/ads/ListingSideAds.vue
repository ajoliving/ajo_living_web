<!--
 * 列表右側展示廣告欄。
 * 1. 依頻道讀取公開 listing_side 展示廣告。
 * 2. 固定 3 個 16:9 短廣告位與 2 個 9:16 長廣告位。
 * 3. 在未設定廣告時保留固定比例佔位。
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

interface ListingAdSlot {
  slotIndex: number;
  kind: 'short' | 'long';
  label: string;
  ad?: PublicDisplayAdResponse;
}

const slotDefinitions: ListingAdSlot[] = [
  { slotIndex: 1, kind: 'short', label: '16:9' },
  { slotIndex: 2, kind: 'short', label: '16:9' },
  { slotIndex: 3, kind: 'short', label: '16:9' },
  { slotIndex: 4, kind: 'long', label: '9:16' },
  { slotIndex: 5, kind: 'long', label: '9:16' },
];

// 1. 讀取公開展示廣告
const loadAds = async (): Promise<void> => {
  loading.value = true;

  try {
    const { data } = await fetchPublicDisplayAds({
      channel: props.channel,
      placement: 'listing_side',
      limit: slotDefinitions.length,
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

// 4. 依固定 slot 回填公開廣告
const displaySlots = computed<ListingAdSlot[]>(() => {
  const adBySlot = new Map<number, PublicDisplayAdResponse>();
  ads.value.forEach((ad, index) => {
    const slotIndex = ad.slot_index || index + 1;
    if (!adBySlot.has(slotIndex)) {
      adBySlot.set(slotIndex, ad);
    }
  });

  return slotDefinitions.map((slot) => ({
    ...slot,
    ad: adBySlot.get(slot.slotIndex),
  }));
});

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
    <template
      v-for="slot in displaySlots"
      :key="slot.slotIndex"
    >
      <a
        v-if="slot.ad && !loading"
        class="listing-side-ads__slot"
        :class="[
          `listing-side-ads__slot--${slot.kind}`,
          `listing-side-ads__slot--${slot.ad.display_layout}`,
        ]"
        :href="slot.ad.target_url || undefined"
        :target="slot.ad.target_url ? '_blank' : undefined"
        rel="noopener noreferrer"
      >
        <img
          v-if="resolveImageURL(slot.ad)"
          :src="resolveImageURL(slot.ad)"
          :alt="resolveTitle(slot.ad)"
        />
        <div class="listing-side-ads__body">
          <p>廣告</p>
          <h2>{{ resolveTitle(slot.ad) }}</h2>
          <span v-if="resolveText(slot.ad)">{{ resolveText(slot.ad) }}</span>
        </div>
      </a>

      <div
        v-else
        class="listing-side-ads__slot listing-side-ads__slot--placeholder"
        :class="`listing-side-ads__slot--${slot.kind}`"
      >
        <span>{{ loading ? '廣告' : slot.label }}</span>
      </div>
    </template>
  </aside>
</template>

<style scoped>
.listing-side-ads {
  display: grid;
  width: 100%;
  gap: 14px;
  align-items: stretch;
}

.listing-side-ads__slot {
  position: relative;
  display: flex;
  overflow: hidden;
  flex-direction: column;
  justify-content: flex-end;
  min-height: 0;
  border: 1px solid var(--bdr, #e4e4e4);
  border-radius: 8px;
  background: var(--sur, #ffffff);
  box-shadow: var(--shadow-sm, 0 1px 2px rgb(0 0 0 / 0.04));
  color: var(--ink, #1a1a1a);
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

.listing-side-ads__slot--short {
  aspect-ratio: 16 / 9;
}

.listing-side-ads__slot--long {
  aspect-ratio: 9 / 16;
}

.listing-side-ads__slot--image_full .listing-side-ads__body {
  background: linear-gradient(180deg, rgb(255 255 255 / 0), rgb(255 255 255 / 0.92) 76%, #fff 100%);
  color: var(--ink, #1a1a1a);
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
  background: linear-gradient(180deg, rgb(255 255 255 / 0), rgb(255 255 255 / 0.94) 72%, #fff 100%);
  padding: 13px;
}

.listing-side-ads__body p {
  margin: 0;
  color: var(--brand, rgb(var(--color-primary)));
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

.listing-side-ads__slot--long .listing-side-ads__body {
  padding: 12px;
}

.listing-side-ads__slot--long .listing-side-ads__body h2 {
  font-size: 12px;
}

.listing-side-ads__slot--long .listing-side-ads__body span {
  font-size: 10px;
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

@media (max-width: 1199px) {
  .listing-side-ads {
    display: none;
  }
}
</style>

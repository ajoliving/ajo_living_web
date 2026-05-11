<!--
 * 首頁入口頁。
 * 1. 以全螢幕 sticky stage 承接二手列表、樓盤放售與服務住宅三個模組。
 * 2. 移除首屏之後的模組說明卡，讓首頁內容保持聚焦。
 * 3. 以 `/channel-home/overview` 與 `/home/content` 的真實資料驅動首頁。
-->
<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';

import { fetchChannelHomeOverview } from '@/httpapis/home';
import { fetchHomeContent } from '@/httpapis/home-content';
import type { HomeChannelEntry, HomeFeaturedSecondhand } from '@/model/home';
import type { HomeCarouselImage, HomeModuleCard } from '@/model/home-content';
import { buildHomeLandingContent } from '@/pages/home/home';
import HomeStageShowcase from '@/pages/home/widgets/HomeStageShowcase.vue';
import { useFeedbackStore } from '@/stores/feedback';

const { t } = useI18n();
const feedbackStore = useFeedbackStore();

const isLoading = ref(false);
const channelItems = ref<HomeChannelEntry[]>([]);
const featuredItems = ref<HomeFeaturedSecondhand[]>([]);
const carouselImages = ref<HomeCarouselImage[]>([]);
const moduleCards = ref<HomeModuleCard[]>([]);

// 1. 組合首頁內容配置
const landingContent = computed(() =>
  buildHomeLandingContent(
    t,
    featuredItems.value.length,
    channelItems.value.length || 3,
    carouselImages.value,
    moduleCards.value,
  ),
);

// 2. 讀取首頁摘要資料
const loadHomeOverview = async () => {
  isLoading.value = true;

  try {
    const [overviewResponse, contentResponse] = await Promise.all([
      fetchChannelHomeOverview(),
      fetchHomeContent(),
    ]);
    channelItems.value = overviewResponse.data.data.channels;
    featuredItems.value = overviewResponse.data.data.featured_secondhand;
    carouselImages.value = contentResponse.data.data.carousel;
    moduleCards.value = contentResponse.data.data.module_cards;
  } catch {
    feedbackStore.pushToast(t('home.loadError'), 'error');
  } finally {
    isLoading.value = false;
  }
};

// 3. 初始化首頁資料
onMounted(async () => {
  await loadHomeOverview();
});
</script>

<template>
  <HomeStageShowcase :content="landingContent" />
</template>

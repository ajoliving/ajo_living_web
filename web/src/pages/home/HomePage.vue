<!--
 * 首頁入口頁。
 * 1. 以全螢幕 sticky stage 承接二手列表、樓盤放售與服務住宅三個模組。
 * 2. 保留模組說明卡，讓首頁在首屏之後直接收束到三個主模組。
 * 3. 以 `/channel-home/overview` 的真實資料驅動首頁統計。
-->
<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';

import { fetchChannelHomeOverview } from '@/httpapis/home';
import type { HomeChannelEntry, HomeFeaturedSecondhand } from '@/model/home';
import { buildHomeLandingContent } from '@/pages/home/home';
import HomeModuleSection from '@/pages/home/widgets/HomeModuleSection.vue';
import HomeStageShowcase from '@/pages/home/widgets/HomeStageShowcase.vue';
import type { HomeModuleCode } from '@/pages/home/home';
import { useFeedbackStore } from '@/stores/feedback';

const { t } = useI18n();
const feedbackStore = useFeedbackStore();

const landingRoot = ref<HTMLElement | null>(null);
const isLoading = ref(false);
const channelItems = ref<HomeChannelEntry[]>([]);
const featuredItems = ref<HomeFeaturedSecondhand[]>([]);
const activeModuleCode = ref<HomeModuleCode>('secondhand');

let sectionObserver: IntersectionObserver | null = null;

// 1. 組合首頁內容配置
const landingContent = computed(() =>
  buildHomeLandingContent(
    t,
    featuredItems.value.length,
    channelItems.value.length || 3,
  ),
);

// 2. 讀取首頁摘要資料
const loadHomeOverview = async () => {
  isLoading.value = true;

  try {
    const { data } = await fetchChannelHomeOverview();
    channelItems.value = data.data.channels;
    featuredItems.value = data.data.featured_secondhand;
  } catch {
    feedbackStore.pushToast(t('home.loadError'), 'error');
  } finally {
    isLoading.value = false;
  }
};

// 3. 註冊首頁模組說明區的 observer
const initializeSectionObserver = async () => {
  await nextTick();

  sectionObserver?.disconnect();

  if (!landingRoot.value) {
    return;
  }

  const sections = Array.from(
    landingRoot.value.querySelectorAll<HTMLElement>('[data-home-module]'),
  );

  if (!sections.length) {
    return;
  }

  sectionObserver = new IntersectionObserver(
    (entries) => {
      const visibleEntries = entries
        .filter((entry) => entry.isIntersecting)
        .sort((left, right) => right.intersectionRatio - left.intersectionRatio);

      const topEntry = visibleEntries[0];
      const moduleCode = topEntry?.target.getAttribute('data-home-module') as HomeModuleCode | null;

      if (moduleCode) {
        activeModuleCode.value = moduleCode;
      }
    },
    {
      threshold: [0.25, 0.45, 0.65],
      rootMargin: '-20% 0px -32% 0px',
    },
  );

  sections.forEach((section) => sectionObserver?.observe(section));
};

// 4. handleStageActiveChange 同步 sticky stage 輸出的 active module
const handleStageActiveChange = (code: HomeModuleCode) => {
  activeModuleCode.value = code;
};

// 5. 初始化首頁資料與模組 observer
onMounted(async () => {
  await loadHomeOverview();
  await initializeSectionObserver();
});

// 6. 離開頁面時釋放 observer
onBeforeUnmount(() => {
  sectionObserver?.disconnect();
});
</script>

<template>
  <div ref="landingRoot">
    <HomeStageShowcase
      :content="landingContent"
      @active-change="handleStageActiveChange"
    />

    <div class="space-y-section pb-8">
      <section class="section-shell">
        <div class="space-y-6">
          <HomeModuleSection
            v-for="module in landingContent.modules"
            :key="module.code"
            :module="module"
            :active="activeModuleCode === module.code"
          />
        </div>
      </section>
    </div>
  </div>
</template>

<style scoped>
.home-section__title {
  margin: 0;
  font-family: var(--font-display);
  font-size: clamp(2rem, 4vw, 3.2rem);
  line-height: 1.04;
  color: rgb(var(--color-text));
}

.home-featured-grid {
  display: grid;
  gap: 1.5rem;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.home-featured-card {
  position: relative;
  overflow: hidden;
}

.home-featured-card::before {
  position: absolute;
  top: -4rem;
  right: -2rem;
  width: 12rem;
  height: 12rem;
  border-radius: 999px;
  background: radial-gradient(circle, rgb(var(--color-primary-soft) / 0.3), transparent 68%);
  content: '';
  filter: blur(10px);
}

.home-featured-card__scope {
  display: inline-flex;
  align-items: center;
  padding: 0.55rem 0.8rem;
  border-radius: 999px;
  background: rgb(var(--color-primary) / 0.1);
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.14em;
  text-transform: uppercase;
  color: rgb(var(--color-primary));
  white-space: nowrap;
}

.home-featured-card__meta {
  display: grid;
  gap: 1rem;
  padding: 1.2rem;
  border: 1px solid rgb(var(--color-border) / 0.68);
  border-radius: 1.6rem;
  background: rgb(var(--color-surface-raised) / 0.82);
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.home-featured-card__meta-label {
  margin: 0;
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: rgb(var(--color-text-muted));
}

.home-featured-card__meta-value {
  margin: 0.55rem 0 0;
  font-family: var(--font-display);
  font-size: 1.4rem;
  line-height: 1.12;
  color: rgb(var(--color-text));
}

.home-featured-card__meta-value--small {
  font-family: var(--font-sans);
  font-size: 0.95rem;
  font-weight: 600;
}

@media (max-width: 1023px) {
  .home-featured-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 639px) {
  .home-section__title {
    font-size: 2.2rem;
  }
}
</style>

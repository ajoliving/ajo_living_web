<!--
 * 首頁入口頁。
 * 1. 依照桌面 HTML 參考還原首頁 Hero、探索服務與精選樓盤區塊。
 * 2. 將原型 onclick 對應為 Vue Router 路由跳轉。
 * 3. 保留首頁摘要 API 讀取作為資料連接狀態。
-->
<script setup lang="ts">
import { computed, onMounted } from 'vue';
import { RouterLink } from 'vue-router';
import { useI18n } from 'vue-i18n';

import { fetchChannelHomeOverview } from '@/httpapis/home';
import { useFeedbackStore } from '@/stores/feedback';

interface HomeServiceLink {
  key: string;
  iconText: string;
  title: string;
  description: string;
  to: string;
}

interface HomeFeaturedProperty {
  key: string;
  title: string;
  price: string;
  suffix: string;
  tags: string[];
  isHot: boolean;
  imageTone: 'neutral' | 'slate' | 'warm';
  imageSize: 'tall' | 'short';
  to: string;
}

const { t } = useI18n();
const feedbackStore = useFeedbackStore();

// 1. 組合首頁服務入口
const serviceLinks = computed<HomeServiceLink[]>(() => [
  {
    key: 'propertySale',
    iconText: '🏢',
    title: t('home.newShell.services.propertySale.title'),
    description: t('home.newShell.services.propertySale.description'),
    to: '/properties',
  },
  {
    key: 'servicedResidence',
    iconText: '🏨',
    title: t('home.newShell.services.servicedResidence.title'),
    description: t('home.newShell.services.servicedResidence.description'),
    to: '/serviced-residences',
  },
  {
    key: 'furniture',
    iconText: '🛋️',
    title: t('home.newShell.services.furniture.title'),
    description: t('home.newShell.services.furniture.description'),
    to: '/furniture',
  },
  {
    key: 'offers',
    iconText: '🎁',
    title: t('home.newShell.services.offers.title'),
    description: t('home.newShell.services.offers.description'),
    to: '/supermarket-offers',
  },
]);

// 2. 組合精選樓盤卡片
const featuredProperties = computed<HomeFeaturedProperty[]>(() => [
  {
    key: 'jordan',
    title: t('home.newShell.featured.jordan.title'),
    price: t('home.newShell.featured.jordan.price'),
    suffix: t('home.newShell.featured.priceSuffix'),
    tags: [
      t('home.newShell.featured.tags.hot'),
      t('home.newShell.featured.tags.kowloon'),
    ],
    isHot: true,
    imageTone: 'neutral',
    imageSize: 'tall',
    to: '/properties',
  },
  {
    key: 'central',
    title: t('home.newShell.featured.central.title'),
    price: t('home.newShell.featured.central.price'),
    suffix: t('home.newShell.featured.priceSuffix'),
    tags: [
      t('home.newShell.featured.tags.hongKongIsland'),
      t('home.newShell.featured.tags.office'),
    ],
    isHot: false,
    imageTone: 'slate',
    imageSize: 'short',
    to: '/properties',
  },
  {
    key: 'shatin',
    title: t('home.newShell.featured.shatin.title'),
    price: t('home.newShell.featured.shatin.price'),
    suffix: t('home.newShell.featured.priceSuffix'),
    tags: [
      t('home.newShell.featured.tags.owner'),
      t('home.newShell.featured.tags.newTerritories'),
    ],
    isHot: true,
    imageTone: 'warm',
    imageSize: 'tall',
    to: '/properties',
  },
]);

// 3. 讀取首頁摘要資料
const loadHomeOverview = async (): Promise<void> => {
  try {
    await fetchChannelHomeOverview();
  } catch {
    feedbackStore.pushToast(t('home.loadError'), 'error');
  }
};

// 4. 初始化首頁資料
onMounted(() => {
  void loadHomeOverview();
});
</script>

<template>
  <main class="home-page">
    <section class="home-hero">
      <div class="home-hero__bg home-pattern">
        <svg
          class="home-hero__skyline"
          viewBox="0 0 800 420"
          preserveAspectRatio="xMidYMid slice"
          aria-hidden="true"
        >
          <rect x="0" y="280" width="60" height="140" />
          <rect x="70" y="200" width="80" height="220" />
          <rect x="160" y="240" width="55" height="180" />
          <rect x="225" y="160" width="100" height="260" />
          <rect x="335" y="210" width="70" height="210" />
          <rect x="415" y="140" width="110" height="280" />
          <rect x="535" y="190" width="75" height="230" />
          <rect x="620" y="220" width="55" height="200" />
          <rect x="685" y="170" width="90" height="250" />
          <rect x="785" y="200" width="15" height="220" />
        </svg>

        <div class="home-hero__content">
          <p class="home-hero__eyebrow">
            {{ t('home.newShell.eyebrow') }}
          </p>
          <h1>
            {{ t('home.newShell.titleLineOne') }}<br />
            {{ t('home.newShell.titleLineTwo') }}
          </h1>
          <p class="home-hero__desc">
            {{ t('home.newShell.subtitle') }}
          </p>
          <div class="home-hero__actions">
            <RouterLink
              to="/properties"
              class="home-button home-button--primary"
            >
              {{ t('home.newShell.primaryAction') }}
            </RouterLink>
            <RouterLink
              to="/serviced-residences"
              class="home-button home-button--ghost"
            >
              {{ t('home.newShell.secondaryAction') }}
            </RouterLink>
          </div>
        </div>
      </div>
    </section>

    <section class="home-section">
      <h2 class="home-section__title">
        {{ t('home.newShell.serviceTitle') }}
      </h2>
      <div class="home-category-grid">
        <RouterLink
          v-for="service in serviceLinks"
          :key="service.key"
          :to="service.to"
          class="home-category-card"
        >
          <span class="home-category-card__icon">
            {{ service.iconText }}
          </span>
          <strong>{{ service.title }}</strong>
          <small>{{ service.description }}</small>
        </RouterLink>
      </div>
    </section>

    <section class="home-section home-section--featured">
      <h2 class="home-section__title">
        {{ t('home.newShell.featuredTitle') }}
      </h2>
      <div class="home-featured-grid">
        <RouterLink
          v-for="item in featuredProperties"
          :key="item.key"
          :to="item.to"
          class="home-featured-card"
        >
          <div
            class="home-featured-card__image home-pattern"
            :class="[
              `home-featured-card__image--${item.imageTone}`,
              `home-featured-card__image--${item.imageSize}`,
            ]"
          />
          <div class="home-featured-card__body">
            <div class="home-tag-row">
              <span
                v-for="tag in item.tags"
                :key="tag"
                class="home-tag"
                :class="{ 'home-tag--dark': item.isHot && tag === item.tags[0] }"
              >
                {{ tag }}
              </span>
            </div>
            <strong>{{ item.title }}</strong>
            <p>
              {{ item.price }}
              <span>{{ item.suffix }}</span>
            </p>
          </div>
        </RouterLink>
      </div>
    </section>
  </main>
</template>

<style scoped>
.home-page {
  background: rgb(var(--color-canvas));
  color: rgb(var(--color-text));
}

.home-pattern {
  position: relative;
  overflow: hidden;
}

.home-pattern::after {
  position: absolute;
  inset: 0;
  background: repeating-linear-gradient(
    45deg,
    transparent,
    transparent 5px,
    rgb(0 0 0 / 0.025) 5px,
    rgb(0 0 0 / 0.025) 10px
  );
  content: '';
  pointer-events: none;
}

.home-hero {
  position: relative;
}

.home-hero__bg {
  display: flex;
  position: relative;
  min-height: 340px;
  align-items: center;
  justify-content: center;
  background: linear-gradient(160deg, #1a1a1a 0%, #2e2e2e 100%);
  padding: 60px 48px;
}

.home-hero__skyline {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  fill: #ffffff;
  opacity: 0.08;
}

.home-hero__content {
  display: grid;
  position: relative;
  z-index: 1;
  max-width: 500px;
  justify-items: center;
  text-align: center;
}

.home-hero__eyebrow {
  margin: 0 0 14px;
  animation: home-hero-enter 0.65s ease both;
  color: #aaaaaa;
  font-size: 10px;
  letter-spacing: 3px;
  text-transform: uppercase;
}

.home-hero h1 {
  margin: 0 0 14px;
  animation: home-hero-enter 0.65s ease 0.08s both;
  color: #ffffff;
  font-family: var(--font-display);
  font-size: 44px;
  font-weight: 400;
  line-height: 1.15;
}

.home-hero__desc {
  max-width: 380px;
  margin: 0 0 24px;
  animation: home-hero-enter 0.65s ease 0.16s both;
  color: #aaaaaa;
  font-size: 13px;
  line-height: 1.8;
}

.home-hero__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  justify-content: center;
  animation: home-hero-enter 0.65s ease 0.24s both;
}

.home-button {
  display: flex;
  align-items: center;
  gap: 6px;
  border-radius: 2px;
  cursor: pointer;
  font-size: 12px;
  letter-spacing: 0.5px;
  padding: 10px 22px;
}

.home-button--primary {
  border: 1px solid rgb(var(--color-primary));
  background: rgb(var(--color-primary));
  color: #ffffff;
}

.home-button--ghost {
  border: 1px solid rgb(255 255 255 / 0.5);
  background: transparent;
  color: #ffffff;
}

.home-button--ghost:hover {
  border-color: rgb(255 255 255 / 0.7);
}

.home-section {
  padding: 32px 40px;
}

.home-section--featured {
  background: #f4f4f4;
  padding-top: 32px;
  padding-bottom: 32px;
}

.home-section__title {
  border-left: 3px solid rgb(var(--color-primary));
  margin: 0 0 16px;
  padding-left: 10px;
  color: rgb(var(--color-text));
  font-size: 13px;
  font-weight: 500;
  letter-spacing: 0.5px;
}

.home-category-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
}

.home-category-card {
  display: grid;
  align-content: start;
  min-height: 132px;
  border: 1px solid #e4e4e4;
  border-radius: 3px;
  background: #ffffff;
  padding: 20px 16px;
  cursor: pointer;
  transition:
    border-color 0.15s ease,
    box-shadow 0.15s ease;
}

.home-category-card:hover {
  border-color: rgb(var(--color-primary));
  box-shadow: 0 3px 14px rgb(240 90 0 / 0.15);
}

.home-category-card__icon {
  display: inline-flex;
  margin-bottom: 8px;
  color: rgb(var(--color-primary));
  font-size: 24px;
  line-height: 1;
}

@keyframes home-hero-enter {
  from {
    opacity: 0;
    transform: translateY(14px);
  }

  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.home-category-card strong {
  margin-bottom: 4px;
  color: #1a1a1a;
  font-size: 13px;
  font-weight: 500;
}

.home-category-card small {
  color: #777777;
  font-size: 11px;
  line-height: 1.5;
}

.home-featured-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
}

.home-featured-card {
  overflow: hidden;
  border: 1px solid #e4e4e4;
  border-radius: 3px;
  background: #ffffff;
  cursor: pointer;
}

.home-featured-card:hover {
  box-shadow: 0 3px 14px rgb(240 90 0 / 0.15);
}

.home-featured-card__image {
  width: 100%;
}

.home-featured-card__image--tall {
  height: 140px;
}

.home-featured-card__image--short {
  height: 100px;
}

.home-featured-card__image--neutral {
  background: linear-gradient(160deg, #e8e8e8, #d0d0d0);
}

.home-featured-card__image--slate {
  background: linear-gradient(160deg, #e0e0e8, #c8c8d8);
}

.home-featured-card__image--warm {
  background: linear-gradient(160deg, #e4dcd8, #ccc0bc);
}

.home-featured-card__body {
  padding: 12px;
}

.home-featured-card__body strong {
  display: block;
  margin-top: 5px;
  color: #1a1a1a;
  font-size: 13px;
  font-weight: 500;
}

.home-featured-card__body p {
  margin: 7px 0 0;
  color: #1a1a1a;
  font-size: 17px;
  font-weight: 300;
  letter-spacing: -0.3px;
}

.home-featured-card__body p span {
  color: #777777;
  font-size: 11px;
  font-weight: 400;
}

.home-tag-row {
  display: flex;
  gap: 3px;
}

.home-tag {
  border: 1px solid #e4e4e4;
  border-radius: 1px;
  color: #777777;
  font-size: 9px;
  letter-spacing: 0.8px;
  padding: 1px 5px;
}

.home-tag--dark {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary));
  color: #ffffff;
  font-weight: 500;
}

@media (max-width: 1023px) {
  .home-hero__bg {
    padding: 48px 24px;
  }

  .home-category-grid,
  .home-featured-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 640px) {
  .home-hero__bg {
    min-height: 330px;
    padding: 42px 20px;
  }

  .home-hero h1 {
    font-size: 38px;
  }

  .home-section {
    padding: 28px 18px;
  }

  .home-category-grid,
  .home-featured-grid {
    grid-template-columns: 1fr;
  }
}
</style>

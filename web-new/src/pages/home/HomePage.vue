<!--
 * 首頁入口頁。
 * 1. HERO 區：暗色漸層背景 + SVG 城市剪影 + 標題與雙按鈕。
 * 2. 探索服務區：4 個分類卡片（樓盤租售、服務式住宅、家具市集、綜合優惠）。
 * 3. 精選樓盤區：3 個樓盤卡片（圖片 + 標籤 + 標題 + 價格 + 面積）。
 * 4. 全部使用靜態 mock 資料，不呼叫 API。
-->
<script setup lang="ts">
import { useRouter } from 'vue-router';

type HomeTarget = 'listing' | 'service' | 'market' | 'offers' | 'detail';

interface CategoryCard {
  key: HomeTarget;
  name: string;
  desc: string;
}

interface FeaturedTag {
  label: string;
  dark: boolean;
}

interface FeaturedCard {
  key: string;
  tags: FeaturedTag[];
  title: string;
  price: string;
  unit: string;
  area: string;
  imgClass: 'tall' | 'short';
  bg: string;
}

// 1. 路由跳轉
const router = useRouter();
const go = (target: HomeTarget): void => {
  const routeMap: Record<HomeTarget, string> = {
    listing: '/properties',
    service: '/serviced-residences',
    market: '/furniture',
    offers: '/supermarket-offers',
    detail: '/properties/1',
  };
  void router.push(routeMap[target]);
};

// 2. 探索服務分類靜態資料
const categories: CategoryCard[] = [
  { key: 'listing', name: '樓盤租售', desc: '住宅、商廈、車位一應俱全' },
  { key: 'service', name: '服務式住宅', desc: '短期靈活入住，設施齊備' },
  { key: 'market', name: '家具市集', desc: '社區二手好物交易平台' },
  { key: 'offers', name: '綜合優惠', desc: '住戶專屬折扣與生活禮遇' },
];

// 3. 精選樓盤靜態資料
const featured: FeaturedCard[] = [
  {
    key: 'feat-1',
    tags: [
      { label: '熱門', dark: true },
      { label: '九龍', dark: false },
    ],
    title: '佐敦 高級住宅',
    price: 'HK$36,000',
    unit: '/ 月',
    area: '實用面積 200呎',
    imgClass: 'tall',
    bg: 'linear-gradient(160deg,#e8e8e8,#d0d0d0)',
  },
  {
    key: 'feat-2',
    tags: [
      { label: '香港島', dark: false },
      { label: '商廈', dark: false },
    ],
    title: '中環甲級寫字樓',
    price: 'HK$120,000',
    unit: '/ 月',
    area: '實用面積 2,400呎',
    imgClass: 'short',
    bg: 'linear-gradient(160deg,#e0e0e8,#c8c8d8)',
  },
  {
    key: 'feat-3',
    tags: [
      { label: '業主盤', dark: true },
      { label: '新界', dark: false },
    ],
    title: '沙田第一城 3房',
    price: 'HK$18,500',
    unit: '/ 月',
    area: '實用面積 650呎',
    imgClass: 'tall',
    bg: 'linear-gradient(160deg,#e4dcd8,#ccc0bc)',
  },
];
</script>

<template>
  <main class="page-home">
    <!-- 1. HERO -->
    <section class="hero">
      <div class="hero-bg pat">
        <svg
          class="hero-skyline"
          viewBox="0 0 800 420"
          preserveAspectRatio="xMidYMid slice"
          aria-hidden="true"
        >
          <rect x="0" y="280" width="60" height="140" fill="white" />
          <rect x="70" y="200" width="80" height="220" fill="white" />
          <rect x="160" y="240" width="55" height="180" fill="white" />
          <rect x="225" y="160" width="100" height="260" fill="white" />
          <rect x="335" y="210" width="70" height="210" fill="white" />
          <rect x="415" y="140" width="110" height="280" fill="white" />
          <rect x="535" y="190" width="75" height="230" fill="white" />
          <rect x="620" y="220" width="55" height="200" fill="white" />
          <rect x="685" y="170" width="90" height="250" fill="white" />
          <rect x="785" y="200" width="15" height="220" fill="white" />
        </svg>
        <div class="hero-content">
          <div class="hero-eyebrow">AJO LIVING</div>
          <h1 class="hero-title">理想生活<br />由此出發</h1>
          <p class="hero-desc">
            全港最大社區生活平台，搜尋住宅、服務式公寓、家具及生活優惠，一站式滿足您的所有需要。
          </p>
          <div class="hero-btns">
            <button class="hbtn-primary" type="button" @click="go('listing')">
              搜尋樓盤
            </button>
            <button class="hbtn-ghost" type="button" @click="go('service')">
              了解服務
            </button>
          </div>
        </div>
      </div>
    </section>

    <!-- 2. 探索服務 -->
    <section class="home-section">
      <div class="home-sec-title">探索服務</div>
      <div class="cat-grid">
        <div
          v-for="cat in categories"
          :key="cat.key"
          class="cat-card"
          role="button"
          tabindex="0"
          @click="go(cat.key)"
          @keyup.enter="go(cat.key)"
        >
          <div class="cat-icon">
            <svg
              viewBox="0 0 32 32"
              width="28"
              height="28"
              aria-hidden="true"
            >
              <rect
                x="4"
                y="10"
                width="24"
                height="18"
                rx="1"
                fill="none"
                stroke="currentColor"
                stroke-width="1.5"
              />
              <path
                d="M8 10V6h16v4"
                fill="none"
                stroke="currentColor"
                stroke-width="1.5"
              />
              <line
                x1="12"
                y1="16"
                x2="20"
                y2="16"
                stroke="currentColor"
                stroke-width="1.5"
              />
              <line
                x1="12"
                y1="20"
                x2="20"
                y2="20"
                stroke="currentColor"
                stroke-width="1.5"
              />
            </svg>
          </div>
          <div class="cat-name">{{ cat.name }}</div>
          <div class="cat-desc">{{ cat.desc }}</div>
        </div>
      </div>
    </section>

    <!-- 3. 精選樓盤 -->
    <section class="home-section home-section--tint">
      <div class="home-sec-title">精選樓盤</div>
      <div class="feat-grid">
        <div
          v-for="feat in featured"
          :key="feat.key"
          class="feat-card"
          role="button"
          tabindex="0"
          @click="go('detail')"
          @keyup.enter="go('detail')"
        >
          <div
            class="feat-img pat"
            :class="feat.imgClass"
            :style="{ background: feat.bg }"
          ></div>
          <div class="feat-body">
            <div class="gtags">
              <span
                v-for="(tag, idx) in feat.tags"
                :key="idx"
                class="gtag"
                :class="{ dark: tag.dark }"
              >{{ tag.label }}</span>
            </div>
            <div class="gtitle">{{ feat.title }}</div>
            <div class="gprice">
              {{ feat.price }} <span>{{ feat.unit }}</span>
            </div>
            <div class="feat-area">{{ feat.area }}</div>
          </div>
        </div>
      </div>
    </section>
  </main>
</template>

<style scoped>
.page-home {
  background: rgb(var(--color-canvas));
  color: rgb(var(--color-text));
}

/* HERO */
.hero {
  position: relative;
}

.hero-bg {
  position: relative;
  min-height: 340px;
  display: flex;
  align-items: center;
  padding: 60px 48px;
  overflow: hidden;
  background: linear-gradient(150deg, rgb(var(--color-text)) 0%, #2d2d2d 100%);
  filter: brightness(0.85);
}

.hero-skyline {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  opacity: 0.08;
  pointer-events: none;
}

.hero-content {
  position: relative;
  z-index: 1;
  max-width: 500px;
}

.hero-eyebrow {
  font-size: 10px;
  letter-spacing: 3px;
  color: rgb(var(--color-ink-4));
  margin-bottom: 14px;
  animation: fadeUp 0.5s ease 0.1s both;
}

.hero-title {
  margin: 0 0 14px;
  font-family: var(--font-display);
  font-size: 44px;
  font-weight: 400;
  line-height: 1.15;
  color: #ffffff;
  animation: fadeUp 0.5s ease 0.25s both;
}

.hero-desc {
  margin: 0 0 24px;
  max-width: 380px;
  font-size: 13px;
  line-height: 1.8;
  color: rgb(var(--color-ink-4));
  animation: fadeUp 0.5s ease 0.4s both;
}

.hero-btns {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  animation: fadeUp 0.5s ease 0.55s both;
}

.hbtn-primary {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 10px 22px;
  font-family: inherit;
  font-size: 12px;
  letter-spacing: 0.5px;
  color: #ffffff;
  background: rgb(var(--color-primary));
  border: 1px solid rgb(var(--color-primary));
  border-radius: 2px;
  cursor: pointer;
  transition:
    background 0.2s ease,
    border-color 0.2s ease;
}

.hbtn-primary:hover {
  background: rgb(var(--color-brand-dark));
  border-color: rgb(var(--color-brand-dark));
}

.hbtn-ghost {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 10px 22px;
  font-family: inherit;
  font-size: 12px;
  letter-spacing: 0.5px;
  color: #ffffff;
  background: rgba(255, 255, 255, 0.08);
  border: 1px solid rgba(255, 255, 255, 0.72);
  border-radius: 2px;
  cursor: pointer;
  transition:
    background 0.2s ease,
    border-color 0.2s ease;
}

.hbtn-ghost:hover {
  background: rgba(255, 255, 255, 0.14);
  border-color: #ffffff;
}

/* HOME SECTION */
.home-section {
  padding: 32px 40px;
}

.home-section--tint {
  padding-top: 32px;
  padding-bottom: 32px;
  background: rgb(var(--color-surface-2));
}

.home-sec-title {
  margin-bottom: 16px;
  padding-left: 10px;
  font-family: var(--font-display);
  font-size: 13px;
  font-weight: 500;
  letter-spacing: 0.5px;
  color: rgb(var(--color-text));
  border-left: 3px solid rgb(var(--color-primary));
}

/* CATEGORIES */
.cat-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
}

.cat-card {
  padding: 20px 16px;
  background: rgb(var(--color-surface));
  border: 1px solid rgb(var(--color-border));
  border-radius: 12px;
  cursor: pointer;
  outline: none;
  transition:
    box-shadow 0.15s ease,
    border-color 0.15s ease;
}

.cat-card:hover,
.cat-card:focus-visible {
  border-color: rgb(var(--color-brand-mid));
  box-shadow: var(--shadow-raised);
}

.cat-icon {
  display: flex;
  align-items: center;
  margin-bottom: 8px;
  color: rgb(var(--color-primary));
}

.cat-name {
  margin-bottom: 4px;
  font-size: 13px;
  font-weight: 500;
  color: rgb(var(--color-text));
}

.cat-desc {
  font-size: 11px;
  line-height: 1.5;
  color: rgb(var(--color-ink-3));
}

/* FEATURED */
.feat-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
}

.feat-card {
  overflow: hidden;
  background: rgb(var(--color-surface));
  border: 1px solid rgb(var(--color-border));
  border-radius: 12px;
  cursor: pointer;
  outline: none;
  transition:
    box-shadow 0.15s ease,
    border-color 0.15s ease;
}

.feat-card:hover,
.feat-card:focus-visible {
  border-color: rgb(var(--color-brand-mid));
  box-shadow: var(--shadow-raised);
}

.feat-img {
  width: 100%;
}

.feat-img.tall {
  height: 140px;
}

.feat-img.short {
  height: 100px;
}

.feat-body {
  padding: 12px;
}

.gtags {
  display: flex;
  gap: 5px;
  margin-bottom: 10px;
}

.gtag {
  padding: 1px 5px;
  font-size: 11px;
  letter-spacing: 0.8px;
  color: rgb(var(--color-ink-3));
  border: 1px solid rgb(var(--color-border));
  border-radius: 1px;
}

.gtag.dark {
  font-weight: 500;
  color: #ffffff;
  background: rgb(var(--color-primary));
  border-color: rgb(var(--color-primary));
}

.gtitle {
  margin-bottom: 3px;
  font-size: 15px;
  font-weight: 600;
  line-height: 1.35;
  color: rgb(var(--color-text));
}

.gprice {
  margin-top: 6px;
  font-size: 20px;
  font-weight: 300;
  letter-spacing: -0.3px;
  color: rgb(var(--color-text));
}

.gprice span {
  font-size: 11px;
  font-weight: 400;
  color: rgb(var(--color-ink-3));
}

.feat-area {
  margin-top: 6px;
  font-size: 10px;
  line-height: 1.4;
  color: rgb(var(--color-ink-3));
}

/* PATTERN OVERLAY */
.pat {
  position: relative;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
}

.pat::after {
  content: '';
  position: absolute;
  inset: 0;
  background: repeating-linear-gradient(
    45deg,
    transparent,
    transparent 5px,
    rgba(0, 0, 0, 0.025) 5px,
    rgba(0, 0, 0, 0.025) 10px
  );
  pointer-events: none;
}

/* ANIMATION */
@keyframes fadeUp {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* RESPONSIVE */
@media (max-width: 900px) {
  .cat-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .feat-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 640px) {
  .hero-bg {
    min-height: 280px;
    padding: 40px 24px;
  }

  .hero-title {
    font-size: 32px;
  }

  .home-section {
    padding: 24px 20px;
  }

  .cat-grid {
    grid-template-columns: 1fr;
  }

  .feat-grid {
    grid-template-columns: 1fr;
  }
}
</style>

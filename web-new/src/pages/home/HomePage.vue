<!--
 * 首頁入口頁。
 * 1. HERO 區：暗色漸層背景 + SVG 城市剪影 + 標題與雙按鈕。
 * 2. 探索服務區：4 個分類卡片（樓盤租售、服務式住宅、家具市集、綜合優惠）。
 * 3. 精選樓盤區：讀取公開樓盤列表並展示真實樓盤資料。
 * 4. 樣式嚴格對齊 HTML 設計稿原生變量。
-->
<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';

import { fetchPropertySaleListings } from '@/httpapis/properties';
import type { PropertyListingSummaryResponse } from '@/model/property';
import {
  resolvePropertyArea,
  resolvePropertyCoverImage,
  resolvePropertyDetailPath,
  resolvePropertyDistrict,
  resolvePropertyPrice,
  resolvePropertyPriceText,
  resolvePropertyPublisherRole,
  resolvePropertyTitle,
  resolvePropertyTransactionType,
  resolvePropertyTypeLabel,
} from '@/utils/property';

type HomeTarget = 'listing' | 'service' | 'market' | 'offers';

interface CategoryCard {
  key: HomeTarget;
  name: string;
  desc: string;
  icon: 'building' | 'service' | 'market' | 'offers';
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
  imageUrl?: string;
  targetPath: string;
}

// 1. 路由跳轉
const router = useRouter();
const go = (target: HomeTarget): void => {
  const routeMap: Record<HomeTarget, string> = {
    listing: '/properties',
    service: '/serviced-residences',
    market: '/marketplace',
    offers: '/supermarket-offers',
  };
  void router.push(routeMap[target]);
};

// 2. 精選樓盤狀態
const propertyListings = ref<PropertyListingSummaryResponse[]>([]);
const loadingFeatured = ref(false);
const featuredError = ref('');

// 3. 精選樓盤資料
const featured = computed<FeaturedCard[]>(() =>
  propertyListings.value.map((listing, index) => toFeaturedCard(listing, index)),
);

// 4. 探索服務分類靜態資料
const categories: CategoryCard[] = [
  { key: 'listing', name: '樓盤租售', desc: '住宅、商廈、車位一應俱全', icon: 'building' },
  { key: 'service', name: '服務式住宅', desc: '短期靈活入住，設施齊備', icon: 'service' },
  { key: 'market', name: '家具市集', desc: '社區二手好物交易平台', icon: 'market' },
  { key: 'offers', name: '綜合優惠', desc: '住戶專屬折扣與生活禮遇', icon: 'offers' },
];

// 5. 讀取真實公開樓盤
const loadFeaturedProperties = async (): Promise<void> => {
  loadingFeatured.value = true;
  featuredError.value = '';

  try {
    const { data } = await fetchPropertySaleListings({
      page: 1,
      page_size: 3,
      sort_by: 'latest',
    });
    propertyListings.value = data.data.items;
  } catch {
    featuredError.value = '暫時無法讀取樓盤。';
    propertyListings.value = [];
  } finally {
    loadingFeatured.value = false;
  }
};

// 6. 轉換首頁樓盤卡片
const toFeaturedCard = (
  listing: PropertyListingSummaryResponse,
  index: number,
): FeaturedCard => {
  const area = resolvePropertyArea(listing);
  const price = resolvePropertyPrice(listing);
  const priceText = resolvePropertyPriceText(listing, 'zh-HK');
  const isRent = resolvePropertyTransactionType(listing) === 'rent';
  const imageUrl = resolvePropertyCoverImage(listing)?.url;

  return {
    key: listing.listing_id,
    tags: [
      { label: resolvePropertyPublisherRole(listing), dark: true },
      { label: resolvePropertyDistrict(listing, 'zh-HK'), dark: false },
      { label: resolvePropertyTypeLabel(listing, 'zh-HK'), dark: false },
    ],
    title: resolvePropertyTitle(listing),
    price: priceText,
    unit: isRent && price > 0 ? '/ 月' : '',
    area: area > 0 ? `實用面積 ${area.toLocaleString('zh-HK')}呎` : '面積待補充',
    imgClass: index % 2 === 1 ? 'short' : 'tall',
    bg: 'linear-gradient(160deg,#e8e8e8,#d0d0d0)',
    imageUrl,
    targetPath: resolvePropertyDetailPath(listing),
  };
};

// 7. 進入樓盤詳情
const openFeatured = (card: FeaturedCard): void => {
  void router.push(card.targetPath);
};

// 8. 隱藏失效樓盤封面
const hideFailedImage = (event: Event): void => {
  (event.currentTarget as HTMLImageElement).style.display = 'none';
};

onMounted(() => {
  void loadFeaturedProperties();
});
</script>

<template>
  <div class="page" id="page-home">
    <!-- 1. HERO -->
    <div class="hero">
      <div class="hero-bg pat" style="background:linear-gradient(160deg,#1a1a1a 0%,#2e2e2e 100%);">
        <svg
          style="position:absolute;inset:0;width:100%;height:100%;opacity:.08"
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
            <button class="hbtn-primary" type="button" @click="go('listing')">搜尋樓盤</button>
            <button class="hbtn-ghost" type="button" @click="go('service')">了解服務</button>
          </div>
        </div>
      </div>
    </div>

    <!-- 2. 探索服務 -->
    <div class="home-section">
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
            <!-- 2.1 樓盤租售圖示 -->
            <svg
              v-if="cat.icon === 'building'"
              viewBox="0 0 32 32"
              width="28"
              height="28"
              aria-hidden="true"
            >
              <rect
                x="5"
                y="8"
                width="22"
                height="20"
                rx="1"
                fill="none"
                stroke="currentColor"
                stroke-width="1.6"
              />
              <rect
                x="9"
                y="12"
                width="4"
                height="4"
                fill="none"
                stroke="currentColor"
                stroke-width="1.4"
              />
              <rect
                x="15"
                y="12"
                width="4"
                height="4"
                fill="none"
                stroke="currentColor"
                stroke-width="1.4"
              />
              <rect
                x="21"
                y="12"
                width="3"
                height="4"
                fill="none"
                stroke="currentColor"
                stroke-width="1.4"
              />
              <rect
                x="9"
                y="19"
                width="4"
                height="4"
                fill="none"
                stroke="currentColor"
                stroke-width="1.4"
              />
              <rect
                x="15"
                y="19"
                width="7"
                height="5"
                fill="none"
                stroke="currentColor"
                stroke-width="1.4"
              />
            </svg>
            <!-- 2.2 服務式住宅圖示 -->
            <svg
              v-else-if="cat.icon === 'service'"
              viewBox="0 0 32 32"
              width="28"
              height="28"
              aria-hidden="true"
            >
              <path
                d="M6 26V12l10-6 10 6v14"
                fill="none"
                stroke="currentColor"
                stroke-width="1.6"
                stroke-linejoin="round"
              />
              <rect
                x="13"
                y="16"
                width="6"
                height="10"
                fill="none"
                stroke="currentColor"
                stroke-width="1.4"
              />
              <line
                x1="6"
                y1="26"
                x2="26"
                y2="26"
                stroke="currentColor"
                stroke-width="1.6"
                stroke-linecap="round"
              />
            </svg>
            <!-- 2.3 家具市集圖示 -->
            <svg
              v-else-if="cat.icon === 'market'"
              viewBox="0 0 32 32"
              width="28"
              height="28"
              aria-hidden="true"
            >
              <path
                d="M5 13h22l-2 13H7L5 13z"
                fill="none"
                stroke="currentColor"
                stroke-width="1.6"
                stroke-linejoin="round"
              />
              <path
                d="M11 13V9a5 5 0 0 1 10 0v4"
                fill="none"
                stroke="currentColor"
                stroke-width="1.6"
              />
            </svg>
            <!-- 2.4 綜合優惠圖示 -->
            <svg
              v-else
              viewBox="0 0 32 32"
              width="28"
              height="28"
              aria-hidden="true"
            >
              <path
                d="M16 4l2.6 5.3 5.9.9-4.3 4.1 1 5.8L16 17.5 10.8 20l1-5.8L7.5 10l5.9-.9L16 4z"
                fill="none"
                stroke="currentColor"
                stroke-width="1.6"
                stroke-linejoin="round"
              />
              <line
                x1="11"
                y1="26"
                x2="21"
                y2="26"
                stroke="currentColor"
                stroke-width="1.6"
                stroke-linecap="round"
              />
            </svg>
          </div>
          <div class="cat-name">{{ cat.name }}</div>
          <div class="cat-desc">{{ cat.desc }}</div>
        </div>
      </div>
    </div>

    <!-- 3. 精選樓盤 -->
    <div class="home-section" style="background:var(--g1);padding-top:32px;padding-bottom:32px;">
      <div class="home-sec-title">精選樓盤</div>
      <div class="feat-grid">
        <div
          v-for="feat in featured"
          :key="feat.key"
          class="feat-card"
          role="button"
          tabindex="0"
          @click="openFeatured(feat)"
          @keyup.enter="openFeatured(feat)"
        >
          <div
            class="feat-img pat"
            :class="feat.imgClass"
            :style="{ background: feat.bg }"
          >
            <img
              v-if="feat.imageUrl"
              :src="feat.imageUrl"
              :alt="feat.title"
              loading="lazy"
              @error="hideFailedImage"
            />
            <span v-else>AJO Living</span>
          </div>
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
            <div class="gprice">{{ feat.price }} <span>{{ feat.unit }}</span></div>
            <div class="feat-area">{{ feat.area }}</div>
          </div>
        </div>
      </div>
      <p
        v-if="loadingFeatured && featured.length === 0"
        class="feat-state"
      >
        正在讀取樓盤。
      </p>
      <p
        v-else-if="featuredError"
        class="feat-state feat-state-error"
      >
        {{ featuredError }}
      </p>
      <p
        v-else-if="featured.length === 0"
        class="feat-state"
      >
        暫時未有公開樓盤。
      </p>
    </div>
  </div>
</template>

<style scoped>
/* 1. HERO 區 */
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
  filter: brightness(0.85);
}

.hero-content {
  position: relative;
  z-index: 1;
  max-width: 500px;
}

.hero-eyebrow {
  font-size: var(--text-xs);
  letter-spacing: 3px;
  color: var(--g3);
  margin-bottom: 14px;
  animation: fadeUp 0.5s ease 0.1s both;
}

.hero-title {
  margin: 0 0 14px;
  font-family: var(--font-serif);
  font-size: var(--text-3xl);
  font-weight: 400;
  line-height: 1.15;
  color: var(--sur);
  animation: fadeUp 0.5s ease 0.25s both;
}

.hero-desc {
  margin: 0 0 24px;
  max-width: 380px;
  font-size: var(--text-base);
  line-height: 1.8;
  color: var(--g3);
  animation: fadeUp 0.5s ease 0.4s both;
}

.hero-btns {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  animation: fadeUp 0.5s ease 0.55s both;
}

/* 1.1 HERO 主按鈕 */
.hbtn-primary {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 10px 22px;
  font-family: var(--font);
  font-size: var(--text-sm);
  letter-spacing: 0.5px;
  font-weight: 500;
  color: var(--sur);
  background: var(--brand);
  border: none;
  border-radius: 2px;
  cursor: pointer;
}

/* 1.2 HERO 次按鈕 */
.hbtn-ghost {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 10px 22px;
  font-family: var(--font);
  font-size: var(--text-sm);
  letter-spacing: 0.5px;
  color: var(--sur);
  background: transparent;
  border: 1px solid rgba(255, 255, 255, 0.5);
  border-radius: 2px;
  cursor: pointer;
  transition: border-color 0.15s ease;
}

.hbtn-ghost:hover {
  border-color: rgba(255, 255, 255, 0.7);
}

/* 2. HOME SECTION 通用 */
.home-section {
  padding: 32px 40px;
  max-width: 1440px;
  margin-left: auto;
  margin-right: auto;
}

.home-sec-title {
  margin-bottom: 16px;
  padding-left: 10px;
  font-family: var(--font-serif);
  font-size: var(--text-base);
  font-weight: 400;
  letter-spacing: 0.5px;
  color: var(--ink);
  border-left: 3px solid var(--brand);
}

/* 3. 探索服務分類卡片 */
.cat-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: var(--sp-3);
}

.cat-card {
  padding: 20px var(--sp-4);
  background: var(--sur);
  border: 1px solid var(--bdr);
  border-radius: var(--r-lg);
  cursor: pointer;
  outline: none;
  transition:
    box-shadow 0.15s ease,
    border-color 0.15s ease;
}

.cat-card:hover,
.cat-card:focus-visible {
  border-color: var(--brand-mid);
  box-shadow: var(--shadow-md);
}

.cat-icon {
  display: flex;
  align-items: center;
  margin-bottom: var(--sp-2);
  color: var(--brand);
}

.cat-name {
  margin-bottom: var(--sp-1);
  font-size: var(--text-base);
  font-weight: 500;
  color: var(--ink);
}

.cat-desc {
  font-size: 11px;
  line-height: 1.5;
  color: var(--g4);
}

/* 4. 精選樓盤卡片 */
.feat-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: var(--sp-3);
}

.feat-card {
  overflow: hidden;
  background: var(--sur);
  border: 1px solid var(--bdr);
  border-radius: var(--r-lg);
  cursor: pointer;
  outline: none;
  transition:
    box-shadow 0.15s ease,
    border-color 0.15s ease;
}

.feat-card:hover,
.feat-card:focus-visible {
  border-color: var(--brand-mid);
  box-shadow: var(--shadow-md);
}

.feat-img {
  width: 100%;
  color: var(--g4);
  font-family: var(--font-serif);
  font-size: var(--text-base);
}

.feat-img.tall {
  height: 140px;
}

.feat-img.short {
  height: 100px;
}

.feat-img img {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.feat-img span {
  position: relative;
  z-index: 1;
}

.feat-body {
  padding: var(--sp-3);
}

.gtags {
  display: flex;
  gap: 5px;
  margin-bottom: 10px;
}

.gtag {
  padding: 3px 7px;
  font-size: 11px;
  letter-spacing: 0.8px;
  color: var(--g4);
  border: 1px solid var(--g2);
  border-radius: 1px;
}

.gtag.dark {
  font-weight: 600;
  letter-spacing: 0.3px;
  border-radius: var(--r-sm);
  color: var(--sur);
  background: var(--brand);
  border-color: var(--brand);
}

.gtitle {
  margin-bottom: 3px;
  font-size: var(--text-md);
  font-weight: 600;
  line-height: 1.35;
  color: var(--ink);
}

.gprice {
  margin-top: 6px;
  font-size: 20px;
  font-weight: 300;
  letter-spacing: -0.3px;
  color: var(--ink);
}

.gprice span {
  font-size: 11px;
  font-weight: 400;
  color: var(--g4);
}

.feat-area {
  margin-top: 6px;
  font-size: var(--text-xs);
  line-height: 1.4;
  color: var(--ink-3);
}

.feat-state {
  margin: var(--sp-4) 0 0;
  padding: var(--sp-4);
  text-align: center;
  font-size: var(--text-sm);
  color: var(--g4);
  background: var(--sur);
  border: 1px solid var(--bdr);
  border-radius: var(--r-lg);
}

.feat-state-error {
  color: #8a2c2c;
}

/* 5. 圖片紋理覆蓋層 */
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

/* 6. 進場動畫 */
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

/* 7. 響應式 */
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
    padding: 40px var(--sp-5);
  }

  .hero-title {
    font-size: var(--text-2xl);
  }

  .home-section {
    padding: var(--sp-5) 20px;
  }

  .cat-grid {
    grid-template-columns: 1fr;
  }

  .feat-grid {
    grid-template-columns: 1fr;
  }
}
</style>

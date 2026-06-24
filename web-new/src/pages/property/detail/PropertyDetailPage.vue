<!--
 * 物業頻道詳情頁。
 * 1. 高保真還原 HTML 設計稿樓盤詳情頁雙欄布局。
 * 2. 左側：標題、價格、統計、設施、描述與大廈資料。
 * 3. 右側：圖集、地圖、代理聯絡卡與相似物件推薦。
 * 4. 全部使用靜態 mock 資料，不呼叫 API。
-->
<script setup lang="ts">
import { ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import AppBreadcrumb from '@/shared/components/navigation/AppBreadcrumb.vue';

// 1. 路由
const route = useRoute();
const router = useRouter();

// 2. 麵包屑項目
const breadcrumbItems = [
  { label: '首頁', to: '/' },
  { label: '樓盤租售', to: '/properties' },
  { label: '佐敦 高級住宅' },
];

// 3. 物件標籤
const tags = ['代理盤', '九龍', '住宅'];

// 4. 物件統計資料
interface DetailStat {
  value: string;
  label: string;
}
const stats: DetailStat[] = [
  { value: '200', label: '實用呎數' },
  { value: '1', label: '睡房' },
  { value: '1', label: '浴室' },
  { value: '高層', label: '樓層' },
];

// 5. 設施配套
const facilities = ['全新裝修', '冷氣', '熱水爐', '電視', '雪櫃', '洗衣機', '寬頻', '升降機'];

// 6. 大廈資料
interface BuildingInfo {
  label: string;
  value: string;
}
const buildingInfo: BuildingInfo[] = [
  { label: '落成年份', value: '2018年' },
  { label: '樓層數目', value: '38層' },
  { label: '管理公司', value: '第一太平' },
  { label: '管理費', value: 'HK$2,800/月' },
];

// 7. 圖集資料
interface GalleryImage {
  id: number;
  background: string;
}
const galleryImages: GalleryImage[] = [
  { id: 1, background: 'linear-gradient(160deg,#e8e8e8,#d0d0d0)' },
  { id: 2, background: 'linear-gradient(160deg,#e0e4e8,#c8d0d8)' },
  { id: 3, background: 'linear-gradient(160deg,#e8e4e0,#d0c8c0)' },
  { id: 4, background: 'linear-gradient(160deg,#e0e8e0,#c8d8c8)' },
  { id: 5, background: 'linear-gradient(160deg,#e8e0e8,#d0c0d0)' },
];
const selectedImageIndex = ref(0);
const extraImageCount = 8;

// 8. 相似推薦物件
interface SimilarListing {
  id: number;
  name: string;
  price: string;
  meta: string;
  background: string;
}
const similarListings: SimilarListing[] = [
  { id: 1, name: '沙田第一城 3房', price: 'HK$18,500/月', meta: '新界 · 650呎', background: 'linear-gradient(135deg,#e4dcd8,#ccc0bc)' },
  { id: 2, name: '旺角商業中心', price: 'HK$28,000/月', meta: '九龍 · 320呎', background: 'linear-gradient(135deg,#d8e0e0,#c0cccc)' },
  { id: 3, name: '銅鑼灣精裝1房', price: 'HK$22,000/月', meta: '香港島 · 280呎', background: 'linear-gradient(135deg,#e0e8e0,#c8d8c8)' },
  { id: 4, name: '尖沙咀海景套房', price: 'HK$32,000/月', meta: '九龍 · 420呎', background: 'linear-gradient(135deg,#e8e0e8,#d0c8d0)' },
  { id: 5, name: '將軍澳新盤2房', price: 'HK$16,500/月', meta: '新界 · 540呎', background: 'linear-gradient(135deg,#e8e8d8,#d0d0b8)' },
];

// 9. 選擇圖片
const selectImage = (index: number): void => {
  selectedImageIndex.value = index;
};

// 10. 跳轉相似物件
const goSimilar = (id: number): void => {
  void router.push(`/properties/${id}`);
};

// 11. 跳轉站內聊天
const goChat = (): void => {
  void router.push('/account/chat');
};

// 12. 取得 listing id（保留路由參數參考）
void route.params.listingId;
</script>

<template>
  <main class="detail-page">
    <div class="detail-wrap">
      <div class="detail-main">
        <!-- 1. 麵包屑 -->
        <AppBreadcrumb :items="breadcrumbItems" />

        <!-- 2. 雙欄布局 -->
        <div class="detail-layout">
          <!-- 2.1 左欄：物件資訊 -->
          <div class="detail-left">
            <div class="detail-tags">
              <span
                v-for="(tag, index) in tags"
                :key="tag"
                :class="{ 'detail-tag--dark': index === 0 }"
                class="detail-tag"
              >
                {{ tag }}
              </span>
            </div>

            <h2 class="detail-title">佐敦 高級住宅</h2>
            <div class="detail-address">油尖旺區 佐敦道 38 號</div>

            <div class="detail-price-row">
              <div class="detail-price-main">
                HK$36,000<span class="detail-price-suffix"> / 月</span>
              </div>
              <div class="detail-price-unit">約 HK$180 / 呎</div>
            </div>

            <div class="detail-stats">
              <div
                v-for="stat in stats"
                :key="stat.label"
                class="detail-stat"
              >
                <div class="detail-stat-val">{{ stat.value }}</div>
                <div class="detail-stat-label">{{ stat.label }}</div>
              </div>
            </div>

            <section class="detail-section">
              <div class="detail-section-title">設施配套</div>
              <div class="detail-pills">
                <span
                  v-for="item in facilities"
                  :key="item"
                  class="detail-pill"
                >
                  {{ item }}
                </span>
              </div>
            </section>

            <section class="detail-section">
              <div class="detail-section-title">物業描述</div>
              <p class="detail-body-text">
                位於佐敦黃金地段，步行3分鐘到佐敦地鐵站，交通四通八達。單位採光充足，全新裝修，廚衛設施齊備，適合單身或情侶入住。大廈管理完善，24小時閉路電視。
              </p>
            </section>

            <section class="detail-section">
              <div class="detail-section-title detail-section-title--label">
                大廈資料
              </div>
              <div class="binfo-grid">
                <div
                  v-for="info in buildingInfo"
                  :key="info.label"
                  class="binfo-card"
                >
                  <div class="binfo-label">{{ info.label }}</div>
                  <div class="binfo-val">{{ info.value }}</div>
                </div>
              </div>
            </section>
          </div>

          <!-- 2.2 右欄：圖集、地圖、代理、推薦 -->
          <div class="detail-right">
            <!-- 2.2.1 圖集 -->
            <div class="detail-right-card">
              <div class="detail-gallery">
                <div
                  class="detail-main-img"
                  :style="{ background: galleryImages[selectedImageIndex].background }"
                >
                  <svg
                    width="120"
                    height="80"
                    viewBox="0 0 120 80"
                    opacity=".15"
                  >
                    <rect
                      x="5"
                      y="10"
                      width="110"
                      height="60"
                      rx="1"
                      stroke="#000"
                      stroke-width="1.5"
                      fill="none"
                    />
                    <rect
                      x="10"
                      y="15"
                      width="30"
                      height="22"
                      rx="1"
                      stroke="#000"
                      fill="none"
                    />
                    <rect
                      x="45"
                      y="15"
                      width="30"
                      height="22"
                      rx="1"
                      stroke="#000"
                      fill="none"
                    />
                    <rect
                      x="80"
                      y="15"
                      width="30"
                      height="22"
                      rx="1"
                      stroke="#000"
                      fill="none"
                    />
                    <rect
                      x="35"
                      y="42"
                      width="50"
                      height="28"
                      rx="1"
                      stroke="#000"
                      fill="none"
                    />
                  </svg>
                </div>
                <div class="detail-thumb-row">
                  <div
                    v-for="(image, index) in galleryImages.slice(0, 3)"
                    :key="image.id"
                    class="detail-thumb"
                    :class="{ 'detail-thumb--active': selectedImageIndex === index }"
                    :style="{ background: image.background }"
                    @click="selectImage(index)"
                  />
                  <div
                    class="detail-thumb detail-thumb-more"
                    :style="{ background: galleryImages[4].background }"
                    @click="selectImage(4)"
                  >
                    +{{ extraImageCount }}
                  </div>
                </div>
              </div>
            </div>

            <!-- 2.2.2 地圖 -->
            <div class="detail-right-card">
              <div class="detail-section-title detail-section-title--label">
                物業位置
              </div>
              <iframe
                class="building-map-frame"
                title="樓盤地圖位置"
                src="https://www.google.com/maps?q=Jordan%20Hong%20Kong&output=embed"
                allowfullscreen
                loading="lazy"
                referrerpolicy="no-referrer-when-downgrade"
              />
            </div>

            <!-- 2.2.3 代理聯絡卡 -->
            <div class="detail-agent-card">
              <div class="detail-agent-profile">
                <div class="detail-agent-avatar">陳</div>
                <div class="detail-agent-copy">
                  <div class="detail-agent-kicker">代理人</div>
                  <div class="detail-agent-name">陳大文</div>
                  <div class="detail-agent-sub">AJO 認證代理 · 立即致電或發送訊息</div>
                </div>
              </div>
              <div class="detail-agent-time">
                <div class="detail-agent-label">預約時間</div>
                <div class="detail-agent-value">今天 14:00 - 18:00</div>
                <div class="detail-agent-note">可另約平日晚上或週末睇樓</div>
              </div>
              <div class="detail-agent-actions">
                <a
                  class="detail-agent-contact"
                  href="tel:+85261234567"
                  aria-label="代理電話"
                >
                  <svg
                    viewBox="0 0 24 24"
                    fill="none"
                    aria-hidden="true"
                  >
                    <path
                      d="M22 16.92v3a2 2 0 0 1-2.18 2A19.8 19.8 0 0 1 11.19 19 19.5 19.5 0 0 1 5 12.81 19.8 19.8 0 0 1 2.08 4.18 2 2 0 0 1 4.06 2h3a2 2 0 0 1 2 1.72c.12.92.33 1.82.62 2.68a2 2 0 0 1-.45 2.11L8 9.73a16 16 0 0 0 6.27 6.27l1.22-1.23a2 2 0 0 1 2.11-.45c.86.29 1.76.5 2.68.62A2 2 0 0 1 22 16.92Z"
                      stroke-width="1.8"
                      stroke-linecap="round"
                      stroke-linejoin="round"
                    />
                  </svg>
                  +852 6123 4567
                </a>
                <a
                  class="detail-agent-contact detail-agent-contact--primary"
                  href="https://wa.me/85261234567"
                  target="_blank"
                  rel="noopener"
                >
                  <svg
                    viewBox="0 0 24 24"
                    fill="none"
                    aria-hidden="true"
                  >
                    <path
                      d="M7.2 20.2 3 21l.9-4A9 9 0 1 1 7.2 20.2Z"
                      stroke-width="1.8"
                      stroke-linecap="round"
                      stroke-linejoin="round"
                    />
                    <path
                      d="M9.2 8.7c.2-.4.4-.5.7-.5h.5c.2 0 .4.1.5.4l.6 1.4c.1.3.1.5-.1.7l-.4.5c.7 1.2 1.6 2.1 2.8 2.8l.5-.4c.2-.2.4-.2.7-.1l1.4.6c.3.1.4.3.4.5v.5c0 .3-.1.5-.5.7-.6.3-1.4.3-2.4-.1-2.4-.8-4.3-2.7-5.1-5.1-.4-1-.4-1.8-.1-2.4Z"
                      stroke-width="1.8"
                      stroke-linecap="round"
                      stroke-linejoin="round"
                    />
                  </svg>
                  WhatsApp
                </a>
                <button
                  class="detail-agent-contact"
                  type="button"
                  @click="goChat"
                >
                  <svg
                    viewBox="0 0 24 24"
                    fill="none"
                    aria-hidden="true"
                  >
                    <path
                      d="M21 11.5a8.4 8.4 0 0 1-9 8.4 8.8 8.8 0 0 1-3.8-.9L3 20l1.1-4.6a8.4 8.4 0 1 1 16.9-3.9Z"
                      stroke-width="1.8"
                      stroke-linecap="round"
                      stroke-linejoin="round"
                    />
                    <path
                      d="M8 11h8M8 14h5"
                      stroke-width="1.8"
                      stroke-linecap="round"
                      stroke-linejoin="round"
                    />
                  </svg>
                  站內聊天
                </button>
              </div>
            </div>

            <!-- 2.2.4 相似推薦 -->
            <div class="similar-section">
              <div class="detail-section-title detail-section-title--label">
                相似物件推薦
              </div>
              <div class="similar-scroll">
                <div
                  v-for="item in similarListings"
                  :key="item.id"
                  class="sim-card"
                  @click="goSimilar(item.id)"
                >
                  <div
                    class="sim-img"
                    :style="{ background: item.background }"
                  />
                  <div class="sim-body">
                    <div class="sim-name">{{ item.name }}</div>
                    <div class="sim-price">{{ item.price }}</div>
                    <div class="sim-meta">{{ item.meta }}</div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </main>
</template>

<style scoped>
/*
 * 樓盤詳情頁樣式。
 * 1. 對齊 HTML 設計稿 #page-detail 結構。
 * 2. CSS 變數映射至專案 token（rgb(var(--color-xxx))）。
 * 3. 雙欄響應式：桌面雙欄、行動單欄。
 */

.detail-page {
  display: block;
  width: 100%;
  min-height: calc(100vh - var(--nav-h));
  background: rgb(var(--color-surface-2));
}

.detail-wrap {
  display: block;
  max-width: 1180px;
  min-height: auto;
  margin: 0 auto;
  padding: 24px;
}

.detail-main {
  max-width: 100%;
  background: rgb(var(--color-surface-2));
  padding: 0;
  overflow: visible;
}

/* 1. 雙欄布局 */
.detail-layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
  gap: 16px;
  align-items: start;
  margin-top: 16px;
}

/* 2. 卡片共用樣式 */
.detail-left,
.detail-right-card,
.detail-agent-card,
.similar-section {
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
  padding: 16px;
}

.detail-right-card,
.detail-agent-card,
.similar-section {
  margin-top: 14px;
}

.detail-right-card:first-of-type {
  margin-top: 0;
}

.detail-left {
  position: sticky;
  top: 72px;
}

/* 3. 標籤 */
.detail-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 5px;
  margin-bottom: 10px;
}

.detail-tag {
  display: inline-flex;
  align-items: center;
  border: 1px solid rgb(var(--color-border));
  border-radius: 999px;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-ink-3));
  font-size: 11px;
  line-height: 1.4;
  padding: 3px 7px;
}

.detail-tag--dark {
  border-color: rgb(var(--color-text));
  background: rgb(var(--color-text));
  color: rgb(var(--color-surface));
}

/* 4. 標題與地址 */
.detail-title {
  margin: 0 0 6px;
  font-family: var(--font-display);
  font-size: 28px;
  font-weight: 500;
  line-height: 1.25;
  color: rgb(var(--color-text));
}

.detail-address {
  margin-bottom: 16px;
  color: rgb(var(--color-ink-3));
  font-size: 14px;
  line-height: 1.6;
}

/* 5. 價格列 */
.detail-price-row {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
  margin-bottom: 16px;
}

.detail-price-main {
  font-size: 34px;
  font-weight: 400;
  letter-spacing: 0;
  color: rgb(var(--color-text));
  line-height: 1.2;
}

.detail-price-suffix {
  font-size: 13px;
  color: rgb(var(--color-ink-3));
  font-weight: 400;
}

.detail-price-unit {
  font-size: 14px;
  color: rgb(var(--color-ink-3));
}

/* 6. 統計區 */
.detail-stats {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
}

.detail-stat {
  min-width: 0;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface-2));
  padding: 12px 8px;
  text-align: center;
}

.detail-stat-val {
  font-size: 21px;
  font-weight: 600;
  line-height: 1.15;
  color: rgb(var(--color-text));
}

.detail-stat-label {
  margin-top: 5px;
  font-size: 12px;
  color: rgb(var(--color-ink-3));
}

/* 7. 區段 */
.detail-section {
  border-top: 1px solid rgb(var(--color-border));
  padding-top: 16px;
  margin-top: 16px;
}

.detail-section-title {
  font-size: 16px;
  font-weight: 700;
  margin-bottom: 12px;
  color: rgb(var(--color-text));
}

.detail-section-title--label {
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: rgb(var(--color-ink-3));
}

/* 8. 設施標籤 */
.detail-pills {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.detail-pill {
  display: inline-flex;
  align-items: center;
  border: 1px solid rgb(var(--color-border));
  border-radius: 999px;
  background: rgb(var(--color-surface-2));
  color: rgb(var(--color-text));
  font-size: 12px;
  line-height: 1.4;
  padding: 5px 9px;
}

/* 9. 描述文字 */
.detail-body-text {
  margin: 0;
  font-size: 15px;
  line-height: 1.85;
  color: rgb(var(--color-text));
}

/* 10. 大廈資料網格 */
.binfo-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.binfo-card {
  min-width: 0;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface-2));
  padding: 12px;
}

.binfo-label {
  font-size: 12px;
  color: rgb(var(--color-ink-3));
}

.binfo-val {
  margin-top: 4px;
  font-size: 15px;
  font-weight: 600;
  line-height: 1.45;
  color: rgb(var(--color-text));
}

/* 11. 圖集 */
.detail-gallery {
  display: grid;
  gap: 8px;
}

.detail-main-img {
  display: flex;
  height: 300px;
  border-radius: 7px;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  background: rgb(var(--color-surface-2));
}

.detail-thumb-row {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 8px;
}

.detail-thumb {
  height: 86px;
  border-radius: 6px;
  cursor: pointer;
  border: 2px solid transparent;
  transition: border-color 0.15s ease;
  background: rgb(var(--color-surface-2));
}

.detail-thumb:hover {
  border-color: rgb(var(--color-brand-mid));
}

.detail-thumb--active {
  border-color: rgb(var(--color-primary));
}

.detail-thumb-more {
  display: flex;
  align-items: center;
  justify-content: center;
  color: rgb(var(--color-text));
  font-size: 16px;
  font-weight: 700;
}

/* 12. 地圖 */
.building-map-frame {
  display: block;
  width: 100%;
  height: 230px;
  border: 0;
  border-radius: 8px;
  background: rgb(var(--color-surface-2));
}

/* 13. 代理卡片 */
.detail-agent-card {
  display: grid;
  grid-template-columns: 1fr;
  gap: 14px;
  align-items: start;
}

.detail-agent-profile {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
  padding-bottom: 14px;
  border-bottom: 1px solid rgb(var(--color-surface-3));
}

.detail-agent-avatar {
  display: flex;
  width: 46px;
  height: 46px;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  background: rgb(var(--color-primary-soft));
  color: rgb(var(--color-primary));
  font-size: 15px;
  font-weight: 800;
  flex: 0 0 auto;
}

.detail-agent-copy {
  min-width: 0;
}

.detail-agent-kicker {
  margin-bottom: 3px;
  color: rgb(var(--color-ink-3));
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 1.2px;
  text-transform: uppercase;
}

.detail-agent-name {
  color: rgb(var(--color-text));
  font-size: 16px;
  font-weight: 700;
  line-height: 1.25;
}

.detail-agent-sub {
  margin-top: 4px;
  color: rgb(var(--color-ink-3));
  font-size: 12px;
  font-weight: 500;
  line-height: 1.45;
}

.detail-agent-time {
  display: grid;
  gap: 4px;
  border: 1px solid rgb(var(--color-border));
  border-radius: 7px;
  background: rgb(var(--color-surface-2));
  padding: 12px;
}

.detail-agent-label {
  color: rgb(var(--color-ink-3));
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.8px;
}

.detail-agent-value {
  color: rgb(var(--color-text));
  font-size: 14px;
  font-weight: 700;
  line-height: 1.35;
}

.detail-agent-note {
  color: rgb(var(--color-ink-3));
  font-size: 12px;
  line-height: 1.45;
}

.detail-agent-actions {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 8px;
}

.detail-agent-contact {
  display: inline-flex;
  min-width: 0;
  min-height: 40px;
  align-items: center;
  justify-content: center;
  gap: 7px;
  border: 1px solid rgb(var(--color-border));
  border-radius: 7px;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text));
  font-family: inherit;
  font-size: 12px;
  font-weight: 700;
  text-decoration: none;
  white-space: nowrap;
  padding: 0 10px;
  cursor: pointer;
  transition: border-color 0.15s ease, color 0.15s ease, background 0.15s ease;
}

.detail-agent-contact svg {
  width: 15px;
  height: 15px;
  flex: 0 0 auto;
  stroke: currentColor;
  fill: none;
}

.detail-agent-contact:hover {
  border-color: rgb(var(--color-brand-mid));
  color: rgb(var(--color-primary));
}

.detail-agent-contact--primary {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary));
  color: rgb(var(--color-primary-contrast));
}

.detail-agent-contact--primary:hover {
  border-color: rgb(var(--color-brand-dark));
  background: rgb(var(--color-brand-dark));
  color: rgb(var(--color-primary-contrast));
}

/* 14. 相似推薦 */
.similar-section {
  padding: 14px;
}

.similar-scroll {
  display: flex;
  gap: 10px;
  max-width: 100%;
  overflow-x: auto;
  overflow-y: hidden;
  min-width: 0;
  padding-bottom: 4px;
  scrollbar-width: thin;
}

.similar-scroll::-webkit-scrollbar {
  height: 6px;
}

.similar-scroll::-webkit-scrollbar-thumb {
  border-radius: 999px;
  background: rgb(var(--color-border-2));
}

.sim-card {
  flex: 0 0 138px;
  min-width: 0;
  overflow: hidden;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
  cursor: pointer;
  transition: border-color 0.15s ease;
}

.sim-card:hover {
  border-color: rgb(var(--color-primary));
}

.sim-img {
  height: 76px;
  border-radius: 7px 7px 0 0;
  background: rgb(var(--color-surface-2));
}

.sim-body {
  min-width: 0;
  padding: 8px 10px 10px;
}

.sim-name,
.sim-price,
.sim-meta {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.sim-name {
  font-size: 12px;
  font-weight: 600;
  line-height: 1.35;
  color: rgb(var(--color-text));
}

.sim-price {
  margin-top: 4px;
  font-size: 13px;
  font-weight: 700;
  color: rgb(var(--color-primary));
}

.sim-meta {
  margin-top: 3px;
  font-size: 10px;
  color: rgb(var(--color-ink-3));
}

/* 15. 響應式 */
@media (max-width: 1023px) {
  .detail-wrap {
    padding: 16px;
  }

  .detail-layout {
    grid-template-columns: 1fr;
  }

  .detail-left {
    position: static;
  }

  .detail-agent-actions {
    grid-template-columns: 1fr;
  }

  .similar-scroll {
    overflow-x: auto;
  }
}

@media (max-width: 640px) {
  .detail-stats {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .binfo-grid {
    grid-template-columns: 1fr;
  }

  .detail-main-img {
    height: 220px;
  }

  .detail-thumb {
    height: 64px;
  }

  .detail-price-main {
    font-size: 28px;
  }
}
</style>

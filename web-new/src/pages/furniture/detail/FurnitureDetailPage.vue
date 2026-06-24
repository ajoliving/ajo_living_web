<!--
 * 家具市集詳情頁。
 * 1. 高保真還原 HTML 設計稿 page-market-detail 雙欄布局。
 * 2. 左側主區：麵包屑、圖集、標籤、標題、價格、統計資料與商品描述。
 * 3. 右側邊欄：賣家聯絡卡片（聯絡賣家、加入收藏）。
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
  { label: '家具市集', to: '/furniture' },
  { label: '北歐實木餐桌' },
];

// 3. 標籤
interface FurnitureTag {
  label: string;
  dark?: boolean;
}
const tags: FurnitureTag[] = [
  { label: '家居傢俱', dark: true },
  { label: '九龍' },
  { label: '近乎全新' },
];

// 4. 統計資料
interface DetailStat {
  value: string;
  label: string;
}
const stats: DetailStat[] = [
  { value: '良好', label: '成色' },
  { value: '餐桌', label: '分類' },
  { value: '自取', label: '交收' },
  { value: '公開', label: '可見範圍' },
];

// 5. 圖集資料
interface GalleryImage {
  id: number;
  background: string;
}
const galleryImages: GalleryImage[] = [
  { id: 1, background: 'linear-gradient(160deg,#eee,#ddd)' },
  { id: 2, background: 'linear-gradient(160deg,#f3f0ec,#ded8d0)' },
  { id: 3, background: 'linear-gradient(160deg,#e8e8e8,#d0d0d0)' },
  { id: 4, background: 'linear-gradient(160deg,#ece8e0,#d4ccc0)' },
];
const selectedImageIndex = ref(0);

// 6. 選擇圖片
const selectImage = (index: number): void => {
  selectedImageIndex.value = index;
};

// 7. 聯絡賣家（跳轉站內聊天）
const handleContactSeller = (): void => {
  void router.push('/account/chat');
};

// 8. 加入收藏（mock）
const isFavorited = ref(false);
const handleToggleFavorite = (): void => {
  isFavorited.value = !isFavorited.value;
};

// 9. 取得 listing id（保留路由參數參考）
void route.params.listingId;
</script>

<template>
  <main class="furniture-detail-page">
    <div class="detail-wrap">
      <!-- 1. 左側主區 -->
      <div class="detail-main">
        <!-- 1.1 麵包屑 -->
        <AppBreadcrumb :items="breadcrumbItems" />

        <!-- 1.2 圖集 -->
        <div class="detail-imgs">
          <div
            class="detail-main-img pat"
            :style="{ background: galleryImages[selectedImageIndex].background }"
          ></div>
          <div class="detail-thumb-row">
            <div
              v-for="(img, index) in galleryImages.slice(1)"
              :key="img.id"
              class="detail-thumb pat"
              :class="{ 'is-active': index + 1 === selectedImageIndex }"
              :style="{ background: img.background }"
              @click="selectImage(index + 1)"
            ></div>
          </div>
        </div>

        <!-- 1.3 詳情資訊 -->
        <div class="detail-info">
          <!-- 1.3.1 標籤 -->
          <div class="gtags">
            <span
              v-for="tag in tags"
              :key="tag.label"
              class="gtag"
              :class="{ 'gtag--dark': tag.dark }"
            >
              {{ tag.label }}
            </span>
          </div>

          <!-- 1.3.2 標題 -->
          <h2 class="detail-title">北歐實木餐桌</h2>
          <div class="detail-sub">旺角 · 可約時間交收 · 同棟優先</div>

          <!-- 1.3.3 價格列 -->
          <div class="detail-price-row">
            <div class="detail-price">HK$2,400</div>
            <div class="detail-published">發布於 2026年6月4日</div>
          </div>

          <!-- 1.3.4 統計資料 -->
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

          <!-- 1.3.5 商品描述 -->
          <div class="detail-desc">
            <div class="detail-desc-title">商品描述</div>
            <p class="body-text">
              實木餐桌保養良好，適合四至六人使用。桌面有正常使用痕跡，不影響日常使用。買家需自行安排搬運，可在晚上或週末交收。
            </p>
          </div>
        </div>
      </div>

      <!-- 2. 右側邊欄 -->
      <aside class="detail-sidebar">
        <div class="detail-contact-card">
          <div class="seller-kicker">賣家</div>
          <div class="seller-name">Admin</div>
          <div class="seller-verified">已通過 AJO Living 帳戶驗證</div>
          <button
            type="button"
            class="hbtn-primary"
            @click="handleContactSeller"
          >
            聯絡賣家
          </button>
          <button
            type="button"
            class="hbtn-ghost"
            :class="{ 'is-favorited': isFavorited }"
            @click="handleToggleFavorite"
          >
            {{ isFavorited ? '已收藏' : '加入收藏' }}
          </button>
        </div>
      </aside>
    </div>
  </main>
</template>

<style scoped>
/* 1. 頁面容器 */
.furniture-detail-page {
  width: 100%;
  background: rgb(var(--color-surface-2));
}

/* 2. 雙欄布局 */
.detail-wrap {
  display: grid;
  grid-template-columns: 1fr 300px;
  gap: 0;
  max-width: 1180px;
  margin: 0 auto;
  min-height: calc(100vh - var(--nav-h, 52px));
  background: rgb(var(--color-surface-2));
}

/* 3. 左側主區 */
.detail-main {
  padding: 20px 28px 40px;
  border-right: 1px solid rgb(var(--color-border));
  overflow-y: auto;
}

/* 4. 圖集 */
.detail-imgs {
  margin-top: 16px;
}

.detail-main-img {
  height: 220px;
  width: 100%;
  border-radius: 3px;
  margin-bottom: 8px;
}

.detail-thumb-row {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 8px;
}

.detail-thumb {
  height: 60px;
  border-radius: 2px;
  cursor: zoom-in;
  border: 2px solid transparent;
  transition: border-color 0.15s ease;
}

.detail-thumb.is-active {
  border-color: rgb(var(--color-primary));
}

/* 5. 斜紋圖案 */
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

/* 6. 詳情資訊區 */
.detail-info {
  margin-top: 20px;
}

/* 7. 標籤 */
.gtags {
  display: flex;
  gap: 3px;
  margin-bottom: 8px;
}

.gtag {
  font-size: 9px;
  letter-spacing: 0.8px;
  color: rgb(var(--color-ink-3));
  border: 1px solid rgb(var(--color-border));
  padding: 1px 5px;
  border-radius: 1px;
}

.gtag--dark {
  background: rgb(var(--color-primary));
  color: rgb(var(--color-surface));
  border-color: rgb(var(--color-primary));
  font-weight: 500;
}

/* 8. 標題與副標題 */
.detail-title {
  font-size: 22px;
  font-weight: 400;
  margin-bottom: 6px;
  color: rgb(var(--color-text));
}

.detail-sub {
  font-size: 12px;
  color: rgb(var(--color-ink-3));
  margin-bottom: 16px;
}

/* 9. 價格列 */
.detail-price-row {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  margin-bottom: 16px;
  padding-bottom: 16px;
  border-bottom: 1px solid rgb(var(--color-border));
}

.detail-price {
  font-size: 28px;
  font-weight: 300;
  letter-spacing: -1px;
  color: rgb(var(--color-text));
}

.detail-published {
  font-size: 12px;
  color: rgb(var(--color-ink-3));
}

/* 10. 統計資料 */
.detail-stats {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 10px;
  margin-bottom: 16px;
}

.detail-stat {
  text-align: center;
  padding: 12px 8px;
  background: rgb(var(--color-surface-2));
  border-radius: 2px;
}

.detail-stat-val {
  font-size: 16px;
  font-weight: 400;
  margin-bottom: 2px;
  color: rgb(var(--color-text));
}

.detail-stat-label {
  font-size: 10px;
  color: rgb(var(--color-ink-3));
}

/* 11. 商品描述 */
.detail-desc {
  border-top: 1px solid rgb(var(--color-border));
  padding-top: 16px;
  margin-top: 16px;
}

.detail-desc-title {
  font-size: 12px;
  font-weight: 500;
  margin-bottom: 8px;
  color: rgb(var(--color-text));
}

.body-text {
  font-size: 13px;
  line-height: 1.8;
  color: rgb(var(--color-ink-2));
}

/* 12. 右側邊欄 */
.detail-sidebar {
  padding: 20px;
  position: sticky;
  top: var(--nav-h, 52px);
  height: calc(100vh - var(--nav-h, 52px));
  overflow-y: auto;
}

/* 13. 賣家聯絡卡片 */
.detail-contact-card {
  border: 1px solid rgb(var(--color-border));
  border-radius: 3px;
  padding: 18px;
  background: rgb(var(--color-surface));
}

.seller-kicker {
  font-size: 11px;
  color: rgb(var(--color-ink-3));
  margin-bottom: 4px;
}

.seller-name {
  font-size: 13px;
  font-weight: 500;
  margin-bottom: 4px;
  color: rgb(var(--color-text));
}

.seller-verified {
  font-size: 11px;
  color: rgb(var(--color-ink-3));
  line-height: 1.6;
  margin-bottom: 14px;
}

/* 14. 按鈕 */
.hbtn-primary {
  width: 100%;
  background: rgb(var(--color-primary));
  color: rgb(var(--color-surface));
  border: none;
  padding: 10px 22px;
  font-size: 12px;
  cursor: pointer;
  font-family: inherit;
  border-radius: 2px;
  letter-spacing: 0.5px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  margin-bottom: 8px;
  transition: background 0.15s ease;
}

.hbtn-primary:hover {
  background: rgb(var(--color-brand-dark));
}

.hbtn-ghost {
  width: 100%;
  background: transparent;
  color: rgb(var(--color-text));
  border: 1px solid rgb(var(--color-border));
  padding: 10px 22px;
  font-size: 12px;
  cursor: pointer;
  font-family: inherit;
  border-radius: 2px;
  letter-spacing: 0.5px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  transition: border-color 0.15s ease, color 0.15s ease;
}

.hbtn-ghost:hover {
  border-color: rgb(var(--color-brand-mid));
  color: rgb(var(--color-primary));
}

.hbtn-ghost.is-favorited {
  border-color: rgb(var(--color-primary));
  color: rgb(var(--color-primary));
}

/* 15. 響應式 */
@media (max-width: 900px) {
  .detail-wrap {
    grid-template-columns: 1fr;
  }

  .detail-main {
    border-right: 0;
    padding: 16px;
  }

  .detail-sidebar {
    position: static;
    height: auto;
    padding: 0 16px 24px;
  }
}

@media (max-width: 600px) {
  .detail-stats {
    grid-template-columns: repeat(2, 1fr);
  }

  .detail-main-img {
    height: 180px;
  }

  .detail-thumb {
    height: 50px;
  }

  .detail-price-row {
    flex-direction: column;
    align-items: flex-start;
    gap: 4px;
  }
}
</style>

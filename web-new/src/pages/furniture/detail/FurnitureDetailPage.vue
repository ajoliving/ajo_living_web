<!--
 * 家具市集詳情頁。
 * 1. 嚴格對齊 HTML 設計稿 #page-market-detail 雙欄布局。
 * 2. 左側主區：麵包屑、圖集、標籤、標題、價格、統計資料與商品描述。
 * 3. 右側邊欄：賣家聯絡卡片（聯絡賣家、加入收藏）。
 * 4. 接入真實二手帖子詳情、聯絡授權、收藏與聊天入口。
-->
<script setup lang="ts">
import { computed } from 'vue';
import { useRouter } from 'vue-router';

import { useFurnitureDetailPage } from './composables/useFurnitureDetailPage';

// 1. 路由
const router = useRouter();
const {
  categoryLabel,
  communityName,
  conditionLabel,
  contactRows,
  coverImage,
  districtLabel,
  formatDeliveryTag,
  galleryImages,
  isFavorited,
  listing,
  listingPrice,
  loading,
  loadingContact,
  openChat,
  openingChat,
  ownerName,
  publishedAt,
  revealContact,
  revealContactLabel,
  selectImage,
  selectedImageIndex,
  t,
  toggleFavorite,
  updatingFavorite,
} = useFurnitureDetailPage();

// 2. 麵包屑導向
const goHome = (): void => {
  void router.push('/');
};
const goMarket = (): void => {
  void router.push('/furniture');
};

// 3. 詳情標籤
const tags = computed(() => {
  if (!listing.value) {
    return [];
  }

  return [
    { label: categoryLabel.value, dark: true },
    { label: districtLabel.value },
    { label: conditionLabel.value },
  ].filter((tag) => tag.label);
});

// 4. 統計資料
const stats = computed(() => [
  { value: conditionLabel.value || '-', label: t('common.label.condition') },
  { value: categoryLabel.value || '-', label: t('marketplace.mine.category') },
  {
    value: listing.value?.delivery_tags.length
      ? listing.value.delivery_tags.map(formatDeliveryTag).join(' / ')
      : t('marketplace.detail.selfPickup'),
    label: t('marketplace.detail.tradeInfo'),
  },
]);
</script>

<template>
  <main class="furniture-detail-page">
    <!-- 1. 麵包屑 -->
    <nav class="breadcrumb">
      <button type="button" class="bc-link" @click="goHome">首頁</button>
      <span class="bc-sep">›</span>
      <button type="button" class="bc-link" @click="goMarket">家具市集</button>
      <span class="bc-sep">›</span>
      <span class="bc-current">{{ listing?.title || t('marketplace.detail.title') }}</span>
    </nav>

    <section
      v-if="loading"
      class="detail-state"
    >
      {{ t('common.status.loading') }}
    </section>

    <section
      v-else-if="!listing"
      class="detail-state"
    >
      {{ t('marketplace.detail.loadError') }}
    </section>

    <div
      v-else
      class="detail-wrap"
    >
      <!-- 2. 左側主區 -->
      <div class="detail-main">
        <!-- 2.1 圖集 -->
        <div class="detail-imgs">
          <div class="detail-main-img" :class="coverImage ? '' : 'pat'">
            <img
              v-if="coverImage"
              :src="coverImage.url"
              :alt="coverImage.alt"
            />
            <span v-else>{{ t('marketplace.filter.noImage') }}</span>
          </div>
          <div
            v-if="galleryImages.length > 1"
            class="detail-thumb-row"
          >
            <button
              v-for="(image, index) in galleryImages"
              :key="image.id"
              type="button"
              class="detail-thumb"
              :class="{ 'is-active': selectedImageIndex === index }"
              @click="selectImage(index)"
            >
              <img
                :src="image.url"
                :alt="image.alt"
              />
            </button>
          </div>
        </div>

        <!-- 2.2 詳情資訊 -->
        <div class="detail-info">
          <!-- 2.2.1 標籤 -->
          <div class="gtags">
            <span
              v-for="tag in tags"
              :key="tag.label"
              class="gtag"
              :class="{ dark: tag.dark }"
            >
              {{ tag.label }}
            </span>
          </div>

          <!-- 2.2.2 標題與副標題 -->
          <h2 class="detail-title">{{ listing.title }}</h2>
          <div class="detail-sub">
            {{ communityName }} · {{ districtLabel }} ·
            {{ listing.visibility_scope === 'building_only' ? t('common.state.buildingOnly') : t('common.state.public') }}
          </div>

          <!-- 2.2.3 價格列 -->
          <div class="detail-price-row">
            <div class="detail-price">{{ listingPrice }}</div>
            <div class="detail-published">{{ t('marketplace.detail.publishedAt') }} {{ publishedAt }}</div>
          </div>

          <!-- 2.2.4 統計資料 -->
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

          <!-- 2.2.5 商品描述 -->
          <div class="detail-desc">
            <div class="detail-desc-title">{{ t('marketplace.detail.description') }}</div>
            <p
              v-if="listing.summary"
              class="body-summary"
            >
              {{ listing.summary }}
            </p>
            <p class="body-text">
              {{ listing.description }}
            </p>
          </div>

          <div class="detail-desc">
            <div class="detail-desc-title">{{ t('marketplace.detail.tradeInfo') }}</div>
            <p class="body-text">
              {{ listing.pickup_location_text || districtLabel }}
            </p>
            <div
              v-if="listing.delivery_tags.length > 0"
              class="detail-delivery-tags"
            >
              <span
                v-for="tag in listing.delivery_tags"
                :key="tag"
              >
                {{ formatDeliveryTag(tag) }}
              </span>
            </div>
          </div>
        </div>
      </div>

      <!-- 3. 右側邊欄 -->
      <aside class="detail-sidebar">
        <div class="detail-contact-card">
          <div class="seller-kicker">{{ t('marketplace.detail.seller') }}</div>
          <div class="seller-name">{{ ownerName }}</div>
          <div class="seller-verified">
            {{ listing.visibility_scope === 'building_only' ? t('marketplace.detail.buildingHint') : t('marketplace.detail.publicHint') }}
          </div>
          <button
            v-if="listing.contact_summary.show_chat"
            type="button"
            class="hbtn-primary"
            :disabled="openingChat"
            @click="openChat"
          >
            {{ openingChat ? t('common.status.loading') : t('common.action.openChat') }}
          </button>
          <button
            v-if="listing.contact_summary.show_phone || listing.contact_summary.show_whatsapp"
            type="button"
            class="hbtn-ghost"
            :disabled="loadingContact"
            @click="revealContact"
          >
            {{ loadingContact ? t('common.status.loading') : revealContactLabel }}
          </button>
          <button
            type="button"
            class="hbtn-ghost"
            :class="{ 'is-favorited': isFavorited }"
            :disabled="updatingFavorite"
            @click="toggleFavorite"
          >
            {{ updatingFavorite ? t('common.status.loading') : isFavorited ? t('marketplace.detail.favoritedAction') : t('marketplace.detail.favoriteAction') }}
          </button>

          <div
            v-if="contactRows.length > 0"
            class="contact-link-list"
          >
            <a
              v-for="row in contactRows"
              :key="row.key"
              class="contact-link"
              :href="row.href"
              target="_blank"
              rel="noopener noreferrer"
            >
              <span>{{ row.label }}</span>
              <strong>{{ row.isWhatsApp ? 'WhatsApp' : row.value }}</strong>
            </a>
          </div>
        </div>
      </aside>
    </div>
  </main>
</template>

<style scoped>
/*
 * 樣式區塊。
 * 1. 頁面容器與雙欄布局。
 * 2. 麵包屑與左側主區：圖集、詳情資訊。
 * 3. 右側邊欄：賣家聯絡卡片。
 * 4. 按鈕與響應式。
 */

/* 1. 頁面容器 */
.furniture-detail-page {
  width: 100%;
  background: var(--sur-2);
}

/* 2. 雙欄布局（對齊設計稿 .detail-wrap + #page-market-detail 覆蓋） */
.detail-wrap {
  display: grid;
  grid-template-columns: 1fr 300px;
  gap: 0;
  max-width: var(--layout-page-max-width);
  min-height: auto;
  margin: 0 auto;
  padding: var(--sp-5);
}

/* 3. 左側主區 */
.detail-main {
  padding: 20px 28px 40px;
  border-right: 1px solid var(--g2);
  overflow-y: auto;
}

/* 4. 麵包屑 */
.breadcrumb {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
  width: 100%;
  margin: 0;
  padding: 12px max(24px, calc((100vw - 1180px) / 2 + 24px));
  border: 1px solid var(--bdr);
  border-right: 0;
  border-left: 0;
  background: var(--sur);
  color: var(--ink-3);
  font-family: var(--font);
  font-size: 12px;
  line-height: 1.5;
}

.bc-link {
  border: 0;
  background: transparent;
  color: var(--ink-3);
  cursor: pointer;
  font-family: var(--font);
  font-size: 12px;
  padding: 0;
}

.bc-link:hover {
  color: var(--brand);
}

.bc-sep {
  color: var(--ink-4);
}

.bc-current {
  color: var(--ink);
  font-weight: 600;
}

.detail-state {
  display: flex;
  align-items: center;
  justify-content: center;
  max-width: var(--layout-page-max-width);
  min-height: 240px;
  margin: var(--sp-5) auto;
  border: 1px solid var(--bdr);
  border-radius: 6px;
  background: var(--sur);
  color: var(--ink-3);
  font-size: var(--text-sm);
}

/* 5. 圖集 */
.detail-main-img {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 220px;
  width: 100%;
  border-radius: 3px;
  margin-bottom: var(--sp-2);
  overflow: hidden;
  background: var(--g1);
  color: var(--g4);
  font-size: var(--text-sm);
}

.detail-main-img img,
.detail-thumb img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.detail-thumb-row {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: var(--sp-2);
}

.detail-thumb {
  height: 60px;
  overflow: hidden;
  border: 1px solid var(--g2);
  border-radius: 2px;
  background: var(--sur);
  cursor: zoom-in;
  padding: 0;
}

.detail-thumb.is-active {
  border-color: var(--accent);
}

/* 6. 斜紋圖案覆蓋層 */
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

/* 7. 詳情資訊區 */
.detail-info {
  margin-top: 20px;
}

/* 8. 標籤 */
.gtags {
  display: flex;
  gap: 3px;
  margin-bottom: var(--sp-2);
}

.gtag {
  font-size: 9px;
  letter-spacing: 0.8px;
  color: var(--g4);
  border: 1px solid var(--g2);
  padding: 1px 5px;
  border-radius: 1px;
}

.gtag.dark {
  background: var(--accent);
  color: var(--white);
  border-color: var(--accent);
  font-weight: 500;
}

/* 9. 標題與副標題 */
.detail-title {
  font-size: 22px;
  font-weight: 400;
  margin-bottom: 6px;
  color: var(--ink);
}

.detail-sub {
  font-size: var(--text-sm);
  color: var(--g4);
  margin-bottom: var(--sp-4);
}

/* 10. 價格列 */
.detail-price-row {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  margin-bottom: var(--sp-4);
  padding-bottom: var(--sp-4);
  border-bottom: 1px solid var(--g2);
}

.detail-price {
  font-size: 28px;
  font-weight: 300;
  letter-spacing: -1px;
  color: var(--ink);
}

.detail-published {
  font-size: var(--text-sm);
  color: var(--g4);
}

/* 11. 統計資料（對齊設計稿 #page-market-detail 三欄覆蓋） */
.detail-stats {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
  margin-bottom: var(--sp-4);
}

.detail-stat {
  text-align: center;
  padding: 12px 8px;
  background: var(--g1);
  border-radius: 2px;
}

.detail-stat-val {
  font-size: 16px;
  font-weight: 400;
  margin-bottom: 2px;
  color: var(--ink);
}

.detail-stat-label {
  font-size: var(--text-xs);
  color: var(--g4);
}

/* 12. 商品描述 */
.detail-desc {
  border-top: 1px solid var(--g2);
  padding-top: var(--sp-4);
  margin-top: var(--sp-4);
}

.detail-desc-title {
  font-size: var(--text-sm);
  font-weight: 500;
  margin-bottom: var(--sp-2);
  color: var(--ink);
}

.body-text {
  font-size: var(--text-base);
  line-height: 1.8;
  color: var(--g5);
}

.body-summary {
  margin: 0 0 var(--sp-2);
  color: var(--ink);
  font-size: var(--text-base);
  line-height: 1.7;
}

.detail-delivery-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: var(--sp-2);
}

.detail-delivery-tags span {
  border: 1px solid var(--g2);
  border-radius: 2px;
  color: var(--g4);
  font-size: var(--text-xs);
  padding: 3px 7px;
}

/* 13. 右側邊欄 */
.detail-sidebar {
  padding: 20px;
  position: sticky;
  top: 48px;
  height: calc(100vh - 48px);
  overflow-y: auto;
}

/* 14. 賣家聯絡卡片 */
.detail-contact-card {
  border: 1px solid var(--g2);
  border-radius: 3px;
  padding: 18px;
}

.seller-kicker {
  font-size: 11px;
  color: var(--g4);
  margin-bottom: 4px;
}

.seller-name {
  font-size: 13px;
  font-weight: 500;
  margin-bottom: 4px;
  color: var(--ink);
}

.seller-verified {
  font-size: 11px;
  color: var(--g4);
  line-height: 1.6;
  margin-bottom: 14px;
}

/* 15. 按鈕 */
.hbtn-primary {
  width: 100%;
  background: var(--accent);
  color: var(--white);
  border: none;
  padding: 10px 22px;
  font-size: var(--text-sm);
  cursor: pointer;
  font-family: inherit;
  border-radius: 2px;
  letter-spacing: 0.5px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  margin-bottom: var(--sp-2);
  transition: background 0.15s ease;
}

.hbtn-primary:hover {
  background: var(--accent-dark);
}

.hbtn-primary:disabled,
.hbtn-ghost:disabled {
  cursor: not-allowed;
  opacity: 0.55;
}

.hbtn-ghost {
  width: 100%;
  background: transparent;
  color: var(--ink);
  border: 1px solid var(--g2);
  padding: 10px 22px;
  font-size: var(--text-sm);
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
  border-color: var(--brand-mid);
  color: var(--accent);
}

.hbtn-ghost.is-favorited {
  border-color: var(--accent);
  color: var(--accent);
}

.contact-link-list {
  display: grid;
  gap: 8px;
  margin-top: var(--sp-3);
}

.contact-link {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  border: 1px solid var(--g2);
  border-radius: 2px;
  color: var(--ink);
  font-size: 12px;
  padding: 9px 10px;
  text-decoration: none;
}

.contact-link:hover {
  border-color: var(--brand-mid);
  color: var(--accent);
}

.contact-link span {
  color: var(--g4);
}

/* 16. 響應式（對齊設計稿 @media max-width:900px） */
@media (max-width: 900px) {
  .breadcrumb {
    padding-right: 24px;
    padding-left: 24px;
  }

  .detail-wrap {
    padding: var(--sp-4);
  }
}

/* 17. 響應式：窄屏堆疊雙欄 */
@media (max-width: 760px) {
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

/* 18. 響應式：極窄屏 */
@media (max-width: 480px) {
  .breadcrumb {
    padding-right: 16px;
    padding-left: 16px;
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

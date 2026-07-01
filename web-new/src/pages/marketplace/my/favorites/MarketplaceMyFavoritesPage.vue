<!--
 * 我的收藏頁。
 * 1. 聚合二手、樓盤與超市優惠真實收藏資料。
 * 2. 支援真實取消收藏與超市商品價格提醒。
 * 3. 使用 router 跳轉至對應詳情頁與搜尋頁。
 * 4. 樣式使用 HTML 原生 CSS 變量（var(--brand)、var(--ink) 等）。
-->
<script setup lang="ts">
import axios from 'axios';
import { computed, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';

import {
  fetchMyFavoriteSecondhandListings,
  unfavoriteSecondhandListing,
} from '@/httpapis/secondhand-listings';
import {
  fetchMyFavoritePropertySales,
  unfavoritePropertySale,
} from '@/httpapis/properties';
import {
  fetchSupermarketFavorites,
  removeSupermarketFavorite,
  saveSupermarketPriceAlert,
} from '@/httpapis/supermarket-offers';
import type { SecondhandListingSummaryResponse } from '@/model/marketplace';
import type { PropertyListingSummaryResponse } from '@/model/property';
import type { SupermarketProduct } from '@/model/supermarket-offers';
import { useFeedbackStore } from '@/stores/feedback';
import { usePreferenceStore } from '@/stores/preferences';
import { formatPrice } from '@/utils/format';
import {
  resolveListingCategoryLabel,
  resolveListingCommunityName,
  resolveListingCoverImage,
  resolveListingPrice,
} from '@/utils/marketplace';
import {
  resolvePropertyArea,
  resolvePropertyCoverImage,
  resolvePropertyDetailPath,
  resolvePropertyDistrict,
  resolvePropertyPrice,
  resolvePropertyPriceText,
  resolvePropertyRooms,
} from '@/utils/property';
import {
  displaySupermarketCategory,
  displaySupermarketStore,
  formatSupermarketHKPrice,
  supermarketPrimaryPrice,
} from '@/utils/supermarket-offers';

type SavedItemType = 'secondhand' | 'property_sale' | 'supermarket_offer';

// 1. 收藏項目資料模型
interface SavedItem {
  id: string;
  sourceId: string;
  type: SavedItemType;
  typeLabel: string;
  name: string;
  price: string;
  priceValue: number;
  area: string;
  areaValue: number;
  rooms: string;
  district: string;
  bg: string;
  imageUrl: string;
  targetPath: string;
  canSetAlert: boolean;
  sortIndex: number;
}

const router = useRouter();
const feedbackStore = useFeedbackStore();
const preferenceStore = usePreferenceStore();
const items = ref<SavedItem[]>([]);
const loading = ref(false);
const actionLoading = ref(false);
const removingIds = ref<string[]>([]);
const sortKey = ref<'date' | 'price-asc' | 'price-desc' | 'area'>('date');
const alertModalOpen = ref(false);
const alertTargetName = ref('');
const alertCurrentPrice = ref('');
const alertProductCode = ref('');
const alertTargetInput = ref<number | null>(null);
const alertSaving = ref(false);

const hasItems = computed(() => items.value.length > 0);

// 2. 排序收藏列表
const sortedItems = computed<SavedItem[]>(() => {
  const list = [...items.value];
  if (sortKey.value === 'price-asc') {
    list.sort((a, b) => a.priceValue - b.priceValue);
  } else if (sortKey.value === 'price-desc') {
    list.sort((a, b) => b.priceValue - a.priceValue);
  } else if (sortKey.value === 'area') {
    list.sort((a, b) => b.areaValue - a.areaValue);
  } else {
    list.sort((a, b) => a.sortIndex - b.sortIndex);
  }
  return list;
});

// 3. 讀取真實收藏列表
const loadFavorites = async (): Promise<void> => {
  loading.value = true;
  try {
    const [secondhandResponse, propertyResponse, supermarketResponse] = await Promise.all([
      fetchMyFavoriteSecondhandListings({ page: 1, page_size: 80 }),
      fetchMyFavoritePropertySales({ page: 1, page_size: 80 }),
      fetchSupermarketFavorites({ page: 1, pageSize: 80 }),
    ]);
    let sortIndex = 0;
    items.value = [
      ...secondhandResponse.data.data.items.map((item) => mapSecondhandFavorite(item, sortIndex++)),
      ...propertyResponse.data.data.items.map((item) => mapPropertyFavorite(item, sortIndex++)),
      ...supermarketResponse.data.data.items.map((item) => mapSupermarketFavorite(item, sortIndex++)),
    ];
  } catch (error: unknown) {
    feedbackStore.pushToast(resolveErrorMessage(error, '收藏列表載入失敗。'), 'error');
    items.value = [];
  } finally {
    loading.value = false;
  }
};

// 4. 移除單筆收藏
const removeItem = async (item: SavedItem): Promise<void> => {
  if (removingIds.value.includes(item.id)) {
    return;
  }

  removingIds.value = [...removingIds.value, item.id];
  try {
    await removeFavoriteByType(item);
    items.value = items.value.filter((current) => current.id !== item.id);
    feedbackStore.pushToast('已取消收藏。', 'success');
  } catch (error: unknown) {
    feedbackStore.pushToast(resolveErrorMessage(error, '取消收藏失敗。'), 'error');
  } finally {
    removingIds.value = removingIds.value.filter((id) => id !== item.id);
  }
};

// 5. 清除目前頁面全部收藏
const clearAll = async (): Promise<void> => {
  if (items.value.length === 0 || actionLoading.value) {
    return;
  }
  if (!window.confirm('確認取消目前顯示的所有收藏？')) {
    return;
  }

  actionLoading.value = true;
  try {
    await Promise.all(items.value.map((item) => removeFavoriteByType(item)));
    items.value = [];
    feedbackStore.pushToast('已清除收藏。', 'success');
  } catch (error: unknown) {
    feedbackStore.pushToast(resolveErrorMessage(error, '清除收藏失敗，請重新整理後再試。'), 'error');
    await loadFavorites();
  } finally {
    actionLoading.value = false;
  }
};

// 6. 跳轉至詳情
const goDetail = (item: SavedItem): void => {
  void router.push(item.targetPath);
};

// 7. 跳轉至搜尋
const goBrowse = (): void => {
  void router.push('/marketplace/filter');
};

// 8. 開啟超市商品目標價提醒 Modal
const openAlertModal = (item: SavedItem): void => {
  if (!item.canSetAlert) {
    return;
  }
  alertTargetName.value = item.name;
  alertCurrentPrice.value = item.price;
  alertProductCode.value = item.sourceId;
  alertTargetInput.value = null;
  alertModalOpen.value = true;
};

// 9. 關閉目標價提醒 Modal
const closeAlertModal = (): void => {
  if (!alertSaving.value) {
    alertModalOpen.value = false;
  }
};

// 10. 儲存超市商品目標價提醒
const saveAlert = async (): Promise<void> => {
  if (!alertProductCode.value || alertTargetInput.value === null || alertTargetInput.value <= 0) {
    feedbackStore.pushToast('請輸入有效目標價。', 'error');
    return;
  }

  alertSaving.value = true;
  try {
    await saveSupermarketPriceAlert({
      productCode: alertProductCode.value,
      targetPrice: alertTargetInput.value,
      priceMode: 'effective',
      offerRequired: false,
      enabled: true,
    });
    feedbackStore.pushToast('價格提醒已設定。', 'success');
    alertModalOpen.value = false;
  } catch (error: unknown) {
    feedbackStore.pushToast(resolveErrorMessage(error, '價格提醒設定失敗。'), 'error');
  } finally {
    alertSaving.value = false;
  }
};

// 11. 判斷移除狀態
const isRemoving = (item: SavedItem): boolean => removingIds.value.includes(item.id);

// 12. 建立卡片圖片樣式
const cardMediaStyle = (item: SavedItem) =>
  item.imageUrl
    ? { backgroundImage: `url("${item.imageUrl}")` }
    : { background: item.bg };

// 13. 映射二手收藏
const mapSecondhandFavorite = (item: SecondhandListingSummaryResponse, sortIndex: number): SavedItem => {
  const cover = resolveListingCoverImage(item);
  const priceValue = resolveListingPrice(item);

  return {
    id: `secondhand-${item.listing_id}`,
    sourceId: item.listing_id,
    type: 'secondhand',
    typeLabel: '二手',
    name: item.title,
    price: item.price_mode === 'free' ? '免費' : formatPrice(priceValue, preferenceStore.locale),
    priceValue,
    area: resolveListingCategoryLabel(item, preferenceStore.locale),
    areaValue: 0,
    rooms: item.condition_level,
    district: resolveListingCommunityName(item),
    bg: 'linear-gradient(160deg,#e8e8e8,#d0d0d0)',
    imageUrl: cover?.url ?? '',
    targetPath: `/marketplace/listing/${item.listing_id}`,
    canSetAlert: false,
    sortIndex,
  };
};

// 14. 映射樓盤收藏
const mapPropertyFavorite = (item: PropertyListingSummaryResponse, sortIndex: number): SavedItem => {
  const cover = resolvePropertyCoverImage(item);
  const area = resolvePropertyArea(item);

  return {
    id: `property-${item.listing_id}`,
    sourceId: item.listing_id,
    type: 'property_sale',
    typeLabel: '樓盤',
    name: item.title,
    price: resolvePropertyPriceText(item, preferenceStore.locale),
    priceValue: resolvePropertyPrice(item),
    area: area > 0 ? `${area}呎` : '-',
    areaValue: area,
    rooms: resolvePropertyRooms(item),
    district: resolvePropertyDistrict(item, preferenceStore.locale),
    bg: 'linear-gradient(160deg,#e4dcd8,#ccc0bc)',
    imageUrl: cover?.url ?? '',
    targetPath: resolvePropertyDetailPath(item),
    canSetAlert: false,
    sortIndex,
  };
};

// 15. 映射超市收藏
const mapSupermarketFavorite = (item: SupermarketProduct, sortIndex: number): SavedItem => {
  const primaryPrice = supermarketPrimaryPrice(item);

  return {
    id: `supermarket-${item.code}`,
    sourceId: item.code,
    type: 'supermarket_offer',
    typeLabel: '優惠',
    name: item.name,
    price: formatSupermarketHKPrice(primaryPrice.effectiveUnitPrice),
    priceValue: primaryPrice.effectiveUnitPrice,
    area: displaySupermarketCategory(item.category1 || item.category2 || item.category3),
    areaValue: 0,
    rooms: displaySupermarketStore(primaryPrice.store || item.bestStore),
    district: item.brand || '綜合優惠',
    bg: 'linear-gradient(160deg,#e0e0e8,#c8c8d8)',
    imageUrl: item.image_url || item.imageUrl || '',
    targetPath: `/supermarket-offers/products/${encodeURIComponent(item.code)}`,
    canSetAlert: true,
    sortIndex,
  };
};

// 16. 按類型取消收藏
const removeFavoriteByType = (item: SavedItem): Promise<unknown> => {
  if (item.type === 'secondhand') {
    return unfavoriteSecondhandListing(item.sourceId);
  }
  if (item.type === 'property_sale') {
    return unfavoritePropertySale(item.sourceId);
  }
  return removeSupermarketFavorite(item.sourceId);
};

// 17. 解析 API 錯誤訊息
const resolveErrorMessage = (error: unknown, fallback: string): string =>
  axios.isAxiosError<{ message?: string }>(error)
    ? error.response?.data?.message ?? fallback
    : fallback;

onMounted(() => {
  void loadFavorites();
});
</script>

<template>
  <div class="saved-page">
    <!-- 1. 工具列：標題與篩選 -->
    <div class="saved-toolbar">
      <div>
        <div class="section-eyebrow">我的收藏</div>
        <div class="saved-toolbar-title">已收藏物件</div>
      </div>
      <div class="saved-filters">
        <select
          v-model="sortKey"
          class="saved-sort"
        >
          <option value="date">最近收藏</option>
          <option value="price-asc">價格低至高</option>
          <option value="price-desc">價格高至低</option>
          <option value="area">面積大至小</option>
        </select>
        <button
          type="button"
          class="nav-login ripple-host"
          style="background:var(--sur)!important;color:var(--ink)!important;border:1px solid var(--bdr)!important;font-size:12px;padding:6px 14px;"
          :disabled="loading || actionLoading || !hasItems"
          @click="clearAll"
        >
          {{ actionLoading ? '處理中' : '清除全部' }}
        </button>
      </div>
    </div>

    <div
      v-if="loading"
      class="saved-loading"
    >
      收藏載入中。
    </div>

    <!-- 2. 收藏卡片網格 -->
    <div
      v-else-if="hasItems"
      class="saved-grid"
    >
      <div
        v-for="item in sortedItems"
        :key="item.id"
        class="saved-card"
      >
        <div
          class="saved-card-img pat"
          :class="{ 'saved-card-img--photo': item.imageUrl }"
          :style="cardMediaStyle(item)"
        >
          <span class="saved-card-type">{{ item.typeLabel }}</span>
          <button
            type="button"
            class="saved-card-remove"
            aria-label="移除收藏"
            :disabled="isRemoving(item)"
            @click="removeItem(item)"
          >
            <svg
              width="12"
              height="12"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2.5"
              stroke-linecap="round"
            >
              <path d="M18 6L6 18M6 6l12 12" />
            </svg>
          </button>
        </div>
        <div class="saved-card-body">
          <div class="saved-card-name">{{ item.name }}</div>
          <div class="saved-card-price">{{ item.price }}</div>
          <div class="saved-card-meta">{{ item.district }} · {{ item.area }} · {{ item.rooms }}</div>
        </div>
        <div class="saved-card-footer">
          <button
            type="button"
            class="saved-card-btn ripple-host"
            :disabled="!item.canSetAlert"
            @click="openAlertModal(item)"
          >
            <svg
              width="11"
              height="11"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
            >
              <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9" />
              <path d="M13.73 21a2 2 0 0 1-3.46 0" />
            </svg>
            {{ item.canSetAlert ? '提醒' : '已收藏' }}
          </button>
          <button
            type="button"
            class="saved-card-btn ripple-host primary"
            @click="goDetail(item)"
          >
            查看詳情
          </button>
        </div>
      </div>
    </div>

    <!-- 3. 空狀態 -->
    <div
      v-else
      class="saved-empty"
    >
      <div class="saved-empty-icon">
        <svg
          width="48"
          height="48"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="1.5"
          stroke-linecap="round"
          stroke-linejoin="round"
        >
          <path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z" />
        </svg>
      </div>
      <div class="saved-empty-title">尚未收藏任何項目</div>
      <div class="saved-empty-desc">瀏覽樓盤、二手商品或綜合優惠時可加入收藏。</div>
      <button
        type="button"
        class="nav-login ripple-host"
        @click="goBrowse"
      >
        去搜尋樓盤
      </button>
    </div>
  </div>

  <!-- 4. 目標價提醒 Modal -->
  <div
    v-if="alertModalOpen"
    class="alert-modal-overlay"
    @click.self="closeAlertModal"
  >
    <div class="alert-modal">
      <div class="alert-modal-header">
        <div class="alert-modal-title">
          <svg
            width="14"
            height="14"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9" />
            <path d="M13.73 21a2 2 0 0 1-3.46 0" />
          </svg>
          設定目標價提醒
        </div>
        <button
          type="button"
          class="alert-close"
          aria-label="關閉"
          @click="closeAlertModal"
        >
          <svg
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
          >
            <path d="M18 6L6 18M6 6l12 12" />
          </svg>
        </button>
      </div>
      <div class="alert-modal-body">
        <div class="alert-prop-name">{{ alertTargetName }}</div>
        <div class="alert-current">目前優惠價：{{ alertCurrentPrice }}</div>
        <label class="alert-label">目標價格 (HK$)</label>
        <div class="alert-input-row">
          <span class="alert-prefix">HK$</span>
          <input
            v-model.number="alertTargetInput"
            class="alert-input"
            type="number"
            placeholder="例如 20"
            min="0.01"
            step="0.01"
          />
        </div>
        <div class="alert-hint">當商品優惠價低於或等於目標價，系統會在通知中心提醒。</div>
        <button
          type="button"
          class="nav-login ripple-host"
          style="width:100%;padding:11px;font-size:13px;"
          :disabled="alertSaving"
          @click="saveAlert"
        >
          {{ alertSaving ? '儲存中' : '確認設定提醒' }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* 1. 頁面容器 */
.saved-page {
  padding: var(--sp-5) var(--sp-6);
}

/* 2. 工具列 */
.saved-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--sp-4);
}

.section-eyebrow {
  font-size: 9px;
  letter-spacing: 2.5px;
  text-transform: uppercase;
  color: var(--accent);
  margin-bottom: 6px;
  font-weight: 500;
}

.saved-toolbar-title {
  font-family: var(--font-serif);
  font-size: var(--text-xl);
  font-weight: 400;
}

.saved-filters {
  display: flex;
  gap: var(--sp-2);
}

.saved-sort {
  border: 1px solid var(--bdr);
  border-radius: var(--r-md);
  padding: 5px 10px;
  font-size: var(--text-sm);
  font-family: var(--font);
  outline: none;
  background: var(--sur);
  color: var(--ink);
  cursor: pointer;
}

.saved-loading {
  border: 1px solid var(--bdr);
  border-radius: var(--r-lg);
  background: var(--sur);
  color: var(--ink-3);
  font-size: 13px;
  padding: 24px;
  text-align: center;
}

/* 4. 收藏卡片網格 */
.saved-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: var(--sp-3);
}

.saved-card {
  background: var(--sur);
  border: 1px solid var(--bdr);
  border-radius: var(--r-lg);
  overflow: hidden;
  position: relative;
  cursor: pointer;
  transition: box-shadow 0.15s, border-color 0.15s;
}

.saved-card:hover {
  box-shadow: var(--shadow-md);
  border-color: var(--brand-mid);
}

/* 5. 卡片圖片區（含 pat 斜紋底紋） */
.saved-card-img {
  height: 110px;
  width: 100%;
  position: relative;
  background-position: center;
  background-size: cover;
}

.saved-card-img--photo::after {
  background: linear-gradient(180deg, rgba(0, 0, 0, 0.18), rgba(0, 0, 0, 0));
}

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
}

.saved-card-type {
  position: absolute;
  left: 8px;
  top: 8px;
  z-index: 1;
  border-radius: var(--r-sm);
  background: rgba(255, 255, 255, 0.9);
  color: var(--ink);
  font-size: 10px;
  font-weight: 600;
  line-height: 1;
  padding: 5px 7px;
}

.saved-card-remove {
  position: absolute;
  top: 8px;
  right: 8px;
  width: 26px;
  height: 26px;
  border-radius: 50%;
  background: rgba(0, 0, 0, 0.5);
  color: #fff;
  border: none;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.15s;
  z-index: 1;
}

.saved-card-remove:hover {
  background: var(--error);
}

/* 6. 卡片內容區 */
.saved-card-body {
  padding: 10px 12px;
}

.saved-card-name {
  font-size: 13px;
  font-weight: 500;
  margin-bottom: 3px;
  color: var(--ink);
}

.saved-card-price {
  font-size: 14px;
  font-weight: 500;
  color: var(--brand);
  margin-bottom: 4px;
}

.saved-card-meta {
  font-size: 10px;
  color: var(--ink-3);
}

/* 7. 卡片底部按鈕 */
.saved-card-footer {
  display: flex;
  gap: 6px;
  padding: 8px 12px;
  border-top: 1px solid var(--sur-3);
}

.saved-card-btn {
  flex: 1;
  font-size: 10px;
  padding: 5px 0;
  border-radius: var(--r-sm);
  border: 1px solid var(--bdr);
  background: var(--sur);
  cursor: pointer;
  font-family: var(--font);
  text-align: center;
  color: var(--ink-2);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 3px;
  transition: background 0.15s;
}

.saved-card-btn:hover {
  background: var(--sur-2);
}

.saved-card-btn:disabled,
.saved-card-remove:disabled,
.nav-login:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

.saved-card-btn.primary {
  background: var(--brand);
  color: #fff;
  border-color: var(--brand);
}

.saved-card-btn.primary:hover {
  background: var(--brand-dark);
}

/* 8. 空狀態 */
.saved-empty {
  text-align: center;
  padding: 80px 40px;
  color: var(--ink-3);
}

.saved-empty-icon {
  font-size: 48px;
  margin-bottom: 16px;
  opacity: 0.4;
  color: var(--ink-3);
  display: flex;
  justify-content: center;
}

.saved-empty-title {
  font-size: 15px;
  font-weight: 500;
  color: var(--ink);
  margin-bottom: 6px;
}

.saved-empty-desc {
  font-size: 12px;
  line-height: 1.6;
  margin-bottom: 20px;
}

/* 9. 目標價提醒 Modal */
.alert-modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.45);
  backdrop-filter: blur(4px);
  z-index: 200;
  display: flex;
  align-items: center;
  justify-content: center;
}

.alert-modal {
  background: var(--sur);
  border-radius: var(--r-xl);
  width: 100%;
  max-width: 420px;
  box-shadow: var(--shadow-lg);
  overflow: hidden;
}

.alert-modal-header {
  padding: 18px 20px;
  border-bottom: 1px solid var(--bdr);
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.alert-modal-title {
  font-size: 15px;
  font-weight: 500;
  color: var(--ink);
  display: flex;
  align-items: center;
  gap: 6px;
}

.alert-close {
  background: none;
  border: none;
  cursor: pointer;
  color: var(--ink-3);
  line-height: 1;
  padding: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.alert-close:hover {
  color: var(--ink);
}

.alert-modal-body {
  padding: 20px;
}

.alert-prop-name {
  font-size: 13px;
  font-weight: 500;
  margin-bottom: 4px;
  color: var(--ink);
}

.alert-current {
  font-size: 12px;
  color: var(--ink-3);
  margin-bottom: 16px;
}

.alert-label {
  font-size: 10px;
  letter-spacing: 1px;
  text-transform: uppercase;
  color: var(--ink-3);
  margin-bottom: 6px;
  display: block;
}

.alert-input-row {
  display: flex;
  gap: 8px;
  align-items: center;
  margin-bottom: 14px;
}

.alert-prefix {
  font-size: 13px;
  color: var(--ink-2);
}

.alert-input {
  flex: 1;
  border: 1px solid var(--bdr);
  border-radius: var(--r-md);
  padding: 9px 12px;
  font-size: 14px;
  font-family: var(--font);
  outline: none;
  background: var(--sur);
  color: var(--ink);
}

.alert-input:focus {
  border-color: var(--brand);
  box-shadow: 0 0 0 3px rgba(240, 90, 0, 0.1);
}

.alert-hint {
  font-size: 11px;
  color: var(--ink-3);
  margin-bottom: 20px;
  line-height: 1.5;
}

.alert-active-list {
  margin-top: 16px;
  border-top: 1px solid var(--bdr);
  padding-top: 14px;
}

.alert-active-title {
  font-size: 10px;
  letter-spacing: 1px;
  text-transform: uppercase;
  color: var(--ink-3);
  margin-bottom: 8px;
}

.alert-active-empty {
  font-size: 11px;
  color: var(--ink-4);
  padding: 8px 0;
}

.alert-active-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 0;
  border-bottom: 1px solid var(--sur-3);
  font-size: 12px;
}

.alert-active-item:last-child {
  border-bottom: none;
}

.alert-active-name {
  flex: 1;
  color: var(--ink-2);
}

.alert-active-price {
  font-weight: 500;
  color: var(--brand);
}

.alert-active-rm {
  background: none;
  border: none;
  color: var(--ink-4);
  cursor: pointer;
  padding: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.alert-active-rm:hover {
  color: var(--error);
}
</style>

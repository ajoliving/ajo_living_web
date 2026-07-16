<!--
 * 我的收藏頁。
 * 1. 聚合二手、樓盤與超市優惠真實收藏資料。
 * 2. 支援真實取消收藏與超市商品價格提醒。
 * 3. 使用 router 跳轉至對應詳情頁與搜尋頁。
 * 4. 樣式使用 HTML 原生 CSS 變量（var(--brand)、var(--ink) 等）。
-->
<script setup lang="ts">
import axios from 'axios';
import { computed, onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
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
import AppIcon from '@/shared/components/base/AppIcon.vue';
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
const { t } = useI18n();
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
    feedbackStore.pushToast(resolveErrorMessage(error, t('marketplace.favoritesPage.loadError')), 'error');
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
    feedbackStore.pushToast(t('marketplace.favoritesPage.removed'), 'success');
  } catch (error: unknown) {
    feedbackStore.pushToast(resolveErrorMessage(error, t('marketplace.favoritesPage.removeError')), 'error');
  } finally {
    removingIds.value = removingIds.value.filter((id) => id !== item.id);
  }
};

// 5. 清除目前頁面全部收藏
const clearAll = async (): Promise<void> => {
  if (items.value.length === 0 || actionLoading.value) {
    return;
  }
  if (!window.confirm(t('marketplace.favoritesPage.clearConfirm'))) {
    return;
  }

  actionLoading.value = true;
  try {
    await Promise.all(items.value.map((item) => removeFavoriteByType(item)));
    items.value = [];
    feedbackStore.pushToast(t('marketplace.favoritesPage.cleared'), 'success');
  } catch (error: unknown) {
    feedbackStore.pushToast(resolveErrorMessage(error, t('marketplace.favoritesPage.clearError')), 'error');
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
    feedbackStore.pushToast(t('marketplace.favoritesPage.invalidTarget'), 'error');
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
    feedbackStore.pushToast(t('marketplace.favoritesPage.alertSaved'), 'success');
    alertModalOpen.value = false;
  } catch (error: unknown) {
    feedbackStore.pushToast(resolveErrorMessage(error, t('marketplace.favoritesPage.alertError')), 'error');
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
    typeLabel: t('marketplace.favoritesPage.typeSecondhand'),
    name: item.title,
    price: item.price_mode === 'free' ? t('marketplace.favoritesPage.free') : formatPrice(priceValue, preferenceStore.locale),
    priceValue,
    area: resolveListingCategoryLabel(item, preferenceStore.locale),
    areaValue: 0,
    rooms: item.condition_level,
    district: resolveListingCommunityName(item, preferenceStore.locale),
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
    typeLabel: t('marketplace.favoritesPage.typeProperty'),
    name: item.title,
    price: resolvePropertyPriceText(item, preferenceStore.locale),
    priceValue: resolvePropertyPrice(item),
    area: area > 0 ? t('marketplace.favoritesPage.squareFeet', { area }) : '-',
    areaValue: area,
    rooms: resolvePropertyRooms(item, preferenceStore.locale),
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
  const primaryPrice = supermarketPrimaryPrice(item, preferenceStore.locale);

  return {
    id: `supermarket-${item.code}`,
    sourceId: item.code,
    type: 'supermarket_offer',
    typeLabel: t('marketplace.favoritesPage.typeOffer'),
    name: item.name,
    price: formatSupermarketHKPrice(primaryPrice.effectiveUnitPrice, preferenceStore.locale),
    priceValue: primaryPrice.effectiveUnitPrice,
    area: displaySupermarketCategory(item.category1 || item.category2 || item.category3, preferenceStore.locale),
    areaValue: 0,
    rooms: displaySupermarketStore(primaryPrice.store || item.bestStore, preferenceStore.locale),
    district: item.brand || t('marketplace.favoritesPage.genericOffers'),
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

watch(
  () => preferenceStore.locale,
  () => {
    void loadFavorites();
  },
);
</script>

<template>
  <div class="saved-page">
    <!-- 1. 工具列：標題與篩選 -->
    <div class="saved-toolbar">
      <div>
        <div class="section-eyebrow">{{ t('marketplace.favoritesPage.kicker') }}</div>
        <div class="saved-toolbar-heading">
          <h1 class="saved-toolbar-title">{{ t('marketplace.favoritesPage.title') }}</h1>
          <span class="saved-count">{{ items.length }}</span>
        </div>
      </div>
      <div class="saved-filters">
        <select
          v-model="sortKey"
          class="saved-sort"
        >
          <option value="date">{{ t('marketplace.favoritesPage.sortRecent') }}</option>
          <option value="price-asc">{{ t('marketplace.favoritesPage.sortPriceAsc') }}</option>
          <option value="price-desc">{{ t('marketplace.favoritesPage.sortPriceDesc') }}</option>
          <option value="area">{{ t('marketplace.favoritesPage.sortAreaDesc') }}</option>
        </select>
        <button
          type="button"
          class="saved-clear-button"
          :disabled="loading || actionLoading || !hasItems"
          @click="clearAll"
        >
          <AppIcon
            name="close"
            :size="15"
          />
          {{ actionLoading ? t('marketplace.favoritesPage.processing') : t('marketplace.favoritesPage.clearAll') }}
        </button>
      </div>
    </div>

    <div
      v-if="loading"
      class="saved-loading"
    >
      {{ t('marketplace.favoritesPage.loading') }}
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
            :aria-label="t('marketplace.favoritesPage.removeAria')"
            :disabled="isRemoving(item)"
            @click="removeItem(item)"
          >
            <AppIcon
              name="close"
              :size="14"
              :stroke-width="2.2"
            />
          </button>
        </div>
        <div class="saved-card-body">
          <button
            type="button"
            class="saved-card-name"
            @click="goDetail(item)"
          >
            {{ item.name }}
          </button>
          <div class="saved-card-price">{{ item.price }}</div>
          <div class="saved-card-meta">{{ item.district }} · {{ item.area }} · {{ item.rooms }}</div>
        </div>
        <div class="saved-card-footer">
          <button
            v-if="item.canSetAlert"
            type="button"
            class="saved-card-btn"
            @click="openAlertModal(item)"
          >
            <AppIcon
              name="bell"
              :size="14"
            />
            {{ t('marketplace.favoritesPage.priceAlert') }}
          </button>
          <button
            type="button"
            class="saved-card-btn primary"
            @click="goDetail(item)"
          >
            <AppIcon
              name="arrow-right"
              :size="14"
            />
            {{ t('marketplace.favoritesPage.viewDetails') }}
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
        <AppIcon
          name="star"
          :size="44"
          :stroke-width="1.4"
        />
      </div>
      <div class="saved-empty-title">{{ t('marketplace.favoritesPage.emptyTitle') }}</div>
      <div class="saved-empty-desc">{{ t('marketplace.favoritesPage.emptyDescription') }}</div>
      <button
        type="button"
        class="saved-primary-action"
        @click="goBrowse"
      >
        <AppIcon
          name="search"
          :size="16"
        />
        {{ t('marketplace.favoritesPage.browseAction') }}
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
          <AppIcon
            name="bell"
            :size="16"
          />
          {{ t('marketplace.favoritesPage.alertTitle') }}
        </div>
        <button
          type="button"
          class="alert-close"
          :aria-label="t('marketplace.favoritesPage.closeAria')"
          @click="closeAlertModal"
        >
          <AppIcon
            name="close"
            :size="18"
          />
        </button>
      </div>
      <div class="alert-modal-body">
        <div class="alert-prop-name">{{ alertTargetName }}</div>
        <div class="alert-current">{{ t('marketplace.favoritesPage.currentPrice', { price: alertCurrentPrice }) }}</div>
        <label class="alert-label">{{ t('marketplace.favoritesPage.targetPrice') }}</label>
        <div class="alert-input-row">
          <span class="alert-prefix">HK$</span>
          <input
            v-model.number="alertTargetInput"
            class="alert-input"
            type="number"
            :placeholder="t('marketplace.favoritesPage.targetPlaceholder')"
            min="0.01"
            step="0.01"
          />
        </div>
        <div class="alert-hint">{{ t('marketplace.favoritesPage.alertHint') }}</div>
        <button
          type="button"
          class="saved-primary-action saved-primary-action--full"
          :disabled="alertSaving"
          @click="saveAlert"
        >
          {{ alertSaving ? t('marketplace.favoritesPage.saving') : t('marketplace.favoritesPage.saveAlert') }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* 1. 頁面容器 */
.saved-page {
  width: 100%;
  min-width: 0;
  padding: 0 0 72px;
}

/* 2. 工具列 */
.saved-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 14px;
  border-bottom: 1px solid var(--bdr);
  padding-bottom: 12px;
}

.section-eyebrow {
  color: var(--accent);
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0;
  margin-bottom: 4px;
  text-transform: uppercase;
}

.saved-toolbar-heading {
  display: flex;
  align-items: center;
  gap: 10px;
}

.saved-toolbar-title {
  margin: 0;
  font-family: var(--font-serif);
  font-size: 26px;
  font-weight: 500;
  line-height: 1.2;
}

.saved-count {
  display: inline-flex;
  min-width: 28px;
  height: 28px;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  background: var(--brand-light);
  color: var(--accent);
  font-size: 12px;
  font-weight: 700;
  padding-inline: 8px;
}

.saved-filters {
  display: flex;
  gap: var(--sp-2);
}

.saved-sort {
  border: 1px solid var(--bdr);
  border-radius: 4px;
  min-height: 36px;
  padding: 0 10px;
  font-size: 12px;
  font-family: var(--font);
  outline: none;
  background: var(--sur);
  color: var(--ink);
  cursor: pointer;
}

.saved-clear-button,
.saved-primary-action {
  display: inline-flex;
  min-height: 36px;
  align-items: center;
  justify-content: center;
  gap: 6px;
  border: 1px solid var(--bdr);
  border-radius: 4px;
  background: var(--sur);
  color: var(--ink);
  cursor: pointer;
  font-family: var(--font);
  font-size: 12px;
  font-weight: 600;
  padding: 0 12px;
}

.saved-clear-button:hover {
  border-color: var(--error);
  color: var(--error);
}

.saved-primary-action {
  border-color: var(--brand);
  background: var(--brand);
  color: #fff;
}

.saved-primary-action--full {
  width: 100%;
  min-height: 42px;
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
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 12px;
}

.saved-card {
  display: grid;
  grid-template-rows: auto minmax(0, 1fr) auto;
  min-width: 0;
  background: var(--sur);
  border: 1px solid var(--bdr);
  border-radius: 6px;
  overflow: hidden;
  position: relative;
  transition: box-shadow 0.15s, border-color 0.15s;
}

.saved-card:hover {
  box-shadow: var(--shadow-md);
  border-color: var(--brand-mid);
}

/* 5. 卡片圖片區（含 pat 斜紋底紋） */
.saved-card-img {
  aspect-ratio: 16 / 10;
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
  border-radius: 3px;
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
  width: 32px;
  height: 32px;
  border-radius: 4px;
  background: rgba(0, 0, 0, 0.58);
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
  display: grid;
  align-content: start;
  gap: 7px;
  min-width: 0;
  padding: 12px;
}

.saved-card-name {
  display: -webkit-box;
  overflow: hidden;
  width: 100%;
  min-height: 2.7em;
  border: 0;
  background: transparent;
  color: var(--ink);
  cursor: pointer;
  font-family: var(--font);
  font-size: 14px;
  font-weight: 600;
  line-height: 1.35;
  padding: 0;
  text-align: left;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.saved-card-name:hover {
  color: var(--brand);
}

.saved-card-price {
  font-family: var(--font-serif);
  font-size: 17px;
  font-weight: 600;
  color: var(--brand);
}

.saved-card-meta {
  overflow: hidden;
  font-size: 11px;
  line-height: 1.5;
  color: var(--ink-3);
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* 7. 卡片底部按鈕 */
.saved-card-footer {
  display: flex;
  gap: 6px;
  padding: 10px 12px;
  border-top: 1px solid var(--sur-3);
}

.saved-card-btn {
  flex: 1;
  min-height: 36px;
  font-size: 11px;
  font-weight: 600;
  padding: 0 8px;
  border-radius: 4px;
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
.saved-clear-button:disabled,
.saved-primary-action:disabled {
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
  border: 1px solid var(--bdr);
  border-radius: 6px;
  background: var(--sur);
  text-align: center;
  padding: 56px 24px;
  color: var(--ink-3);
}

.saved-empty-icon {
  margin-bottom: 14px;
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
  border-radius: 8px;
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

@media (max-width: 767px) {
  .saved-page {
    padding-bottom: calc(var(--app-mobile-content-bottom) + 16px);
  }

  .saved-toolbar {
    align-items: stretch;
  }

  .saved-toolbar-title {
    font-size: 24px;
  }

  .saved-card {
    grid-template-columns: 112px minmax(0, 1fr);
    grid-template-rows: minmax(0, 1fr) auto;
  }

  .saved-card-img {
    grid-row: 1 / 3;
    height: 100%;
    min-height: 164px;
    aspect-ratio: auto;
  }

  .saved-card-body {
    padding: 12px 12px 8px;
  }

  .saved-card-footer {
    padding: 8px 12px 12px;
  }

  .saved-card-btn {
    min-width: 0;
    padding-inline: 6px;
  }

  .saved-empty {
    padding: 40px 20px;
  }

  .alert-modal-overlay {
    align-items: flex-end;
    padding:
      var(--app-safe-top)
      var(--layout-page-padding-inline)
      calc(var(--app-safe-bottom) + 10px);
  }

  .alert-modal {
    max-height: calc(100svh - var(--app-safe-top) - var(--app-safe-bottom) - 20px);
    overflow-y: auto;
    border-radius: 8px 8px 0 0;
  }

  .alert-close {
    width: 44px;
    height: 44px;
  }
}

@media (max-width: 370px) {
  .saved-card {
    grid-template-columns: 96px minmax(0, 1fr);
  }

  .saved-card-footer {
    display: grid;
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>

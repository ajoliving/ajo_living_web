<!--
 * 家具頻道頁。
 * 1. 依照桌面 HTML 參考還原家具市集左側篩選與商品瀑布流。
 * 2. 復用二手帖子 API 查詢家具相關分類。
 * 3. 保持家具入口獨立於完整二手列表。
-->
<script setup lang="ts">
import axios from 'axios';
import { computed, onMounted, ref, watch } from 'vue';
import { RouterLink } from 'vue-router';
import { useI18n } from 'vue-i18n';

import { fetchSecondhandListings } from '@/httpapis/secondhand-listings';
import {
  getMarketplaceCategoryLabel,
  getMarketplaceConditionLabel,
  getMarketplaceDistrictLabel,
  type MarketplaceCategoryCode,
} from '@/constants/marketplace';
import type { SecondhandListingSummaryResponse } from '@/model/marketplace';
import AppIcon from '@/shared/components/base/AppIcon.vue';
import { useFeedbackStore } from '@/stores/feedback';
import { usePreferenceStore } from '@/stores/preferences';
import { formatPrice } from '@/utils/format';

type FurnitureCategoryFilter = 'all' | 'home_furniture' | 'office_furniture' | 'home_decor';

interface FurnitureFilterGroup {
  title: string;
  options: string[];
}

const { t } = useI18n();
const feedbackStore = useFeedbackStore();
const preferenceStore = usePreferenceStore();

const loading = ref(false);
const keyword = ref('');
const selectedCategory = ref<FurnitureCategoryFilter>('all');
const listings = ref<SecondhandListingSummaryResponse[]>([]);

const filterGroups = computed<FurnitureFilterGroup[]>(() => [
  { title: '分類', options: ['全部', '家居傢俱', '家庭電器', '電子產品', 'BB用品', '其他'] },
  { title: '價格範圍', options: ['全部', '$1k以下', '$1k-3k', '$3k-5k', '$10k+'] },
  { title: '成色', options: ['全部', '近乎全新', '良好', '尚可'] },
  { title: '地區', options: ['全部', '香港島', '九龍', '新界', '離島'] },
  { title: '可見範圍', options: ['全部', '公開', '同棟可見'] },
]);

// 1. 讀取家具帖子列表
const loadListings = async (): Promise<void> => {
  loading.value = true;

  try {
    const response = await fetchSecondhandListings({
      page: 1,
      page_size: 24,
      keyword: keyword.value.trim() || undefined,
      category_code: selectedCategory.value === 'all' ? undefined : selectedCategory.value,
      sort_by: 'latest',
    });

    listings.value = response.data.data.items.filter((listing) =>
      ['home_furniture', 'office_furniture', 'home_decor'].includes(listing.category_code),
    );
  } catch (error: unknown) {
    feedbackStore.pushToast(
      axios.isAxiosError<{ message?: string }>(error)
        ? error.response?.data?.message ?? t('channels.furniture.loadError')
        : t('channels.furniture.loadError'),
      'error',
    );
    listings.value = [];
  } finally {
    loading.value = false;
  }
};

// 2. 格式化帖子價格
const formatListingPrice = (listing: SecondhandListingSummaryResponse): string =>
  listing.price_mode === 'free'
    ? t('common.price.free')
    : formatPrice(listing.price_hkd ?? 0, preferenceStore.locale);

// 3. 取得帖子分類文字
const resolveCategoryLabel = (categoryCode: string): string =>
  getMarketplaceCategoryLabel(categoryCode as MarketplaceCategoryCode, preferenceStore.locale);

watch(selectedCategory, () => {
  void loadListings();
});

onMounted(() => {
  void loadListings();
});
</script>

<template>
  <main class="market-page">
    <aside class="market-filter">
      <div class="market-search">
        <input
          v-model="keyword"
          placeholder="搜尋物品..."
          @keyup.enter="loadListings"
        />
        <button
          type="button"
          @click="loadListings"
        >
          搜
        </button>
      </div>

      <section
        v-for="group in filterGroups"
        :key="group.title"
        class="market-filter-section"
      >
        <h2>{{ group.title }}</h2>
        <div class="market-tags">
          <button
            v-for="(option, index) in group.options"
            :key="option"
            type="button"
            class="market-tag"
            :class="{ 'market-tag--active': index === 0 }"
          >
            {{ option }}
          </button>
        </div>
      </section>
    </aside>

    <section class="market-results">
      <div class="market-sort-row">
        <span>{{ loading ? t('common.status.loading') : t('channels.furniture.results', { count: listings.length }) }}</span>
        <select>
          <option>最新發佈</option>
          <option>價格低至高</option>
        </select>
      </div>

      <div
        v-if="!loading && listings.length > 0"
        class="market-grid"
      >
        <RouterLink
          v-for="(listing, index) in listings"
          :key="listing.listing_id"
          :to="`/marketplace/listing/${listing.listing_id}?from=furniture`"
          class="market-card"
        >
          <span
            v-if="index === 0"
            class="market-card__badge"
          >
            全新
          </span>
          <span class="market-card__heart">
            <AppIcon
              name="star"
              :size="13"
            />
          </span>
          <div
            class="market-card__image home-pattern"
            :class="`market-card__image--${(index % 3) + 1}`"
          >
            <img
              v-if="listing.cover_image?.url"
              :src="listing.cover_image.url"
              :alt="listing.title"
            />
            <AppIcon
              v-else
              name="picture"
              :size="30"
            />
          </div>
          <div class="market-card__body">
            <h2>{{ listing.title }}</h2>
            <p>
              {{ getMarketplaceDistrictLabel(listing.district_code, preferenceStore.locale) }}
              ·
              {{ getMarketplaceConditionLabel(listing.condition_level, preferenceStore.locale) }}
            </p>
            <strong>{{ formatListingPrice(listing) }}</strong>
            <div>
              <span>{{ resolveCategoryLabel(listing.category_code) }}</span>
              <span>查看 →</span>
            </div>
          </div>
        </RouterLink>
      </div>

      <div
        v-else
        class="market-empty"
      >
        {{ loading ? t('common.status.loading') : t('channels.furniture.empty') }}
      </div>
    </section>
  </main>
</template>

<style scoped>
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

.market-page {
  display: grid;
  grid-template-columns: 210px 1fr;
  min-height: calc(100vh - var(--app-header-offset, 48px));
  color: #1a1a1a;
}

.market-filter {
  position: sticky;
  top: var(--app-header-offset, 48px);
  height: calc(100vh - var(--app-header-offset, 48px));
  overflow-y: auto;
  border-right: 1px solid #e4e4e4;
  padding: 18px 14px;
}

.market-search {
  display: flex;
  gap: 5px;
  margin-bottom: 14px;
}

.market-search input,
.market-search button,
.market-sort-row select {
  border: 1px solid #e4e4e4;
  border-radius: 2px;
  font: inherit;
  font-size: 11px;
  outline: none;
}

.market-search input {
  min-width: 0;
  flex: 1;
  padding: 7px 9px;
}

.market-search button {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary));
  color: #ffffff;
  cursor: pointer;
  padding: 7px 12px;
}

.market-filter-section {
  margin-bottom: 16px;
}

.market-filter-section h2 {
  margin: 0 0 6px;
  color: #777777;
  font-size: 9px;
  font-weight: 400;
  letter-spacing: 2px;
  text-transform: uppercase;
}

.market-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 3px;
}

.market-tag {
  border: 1px solid #e4e4e4;
  border-radius: 2px;
  background: #ffffff;
  color: #777777;
  cursor: pointer;
  font: inherit;
  font-size: 11px;
  padding: 3px 9px;
}

.market-tag--active {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary));
  color: #ffffff;
}

.market-results {
  min-width: 0;
  padding: 18px 20px 40px;
}

.market-sort-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.market-sort-row span {
  color: #777777;
  font-size: 11px;
}

.market-sort-row select {
  padding: 5px 8px;
}

.market-grid {
  columns: 3;
  column-gap: 10px;
}

.market-card {
  display: inline-block;
  position: relative;
  width: 100%;
  overflow: hidden;
  break-inside: avoid;
  border: 1px solid #e4e4e4;
  border-radius: 3px;
  background: #ffffff;
  margin-bottom: 10px;
  color: inherit;
}

.market-card:hover {
  border-color: rgb(var(--color-primary));
  box-shadow: 0 3px 16px rgb(240 90 0 / 0.15);
}

.market-card__badge {
  position: absolute;
  z-index: 2;
  top: 7px;
  left: 7px;
  background: #1a1a1a;
  color: #ffffff;
  font-size: 9px;
  letter-spacing: 0.3px;
  padding: 1px 5px;
}

.market-card__heart {
  position: absolute;
  z-index: 2;
  top: 7px;
  right: 9px;
  color: #aaaaaa;
  font-size: 13px;
}

.market-card__image {
  display: flex;
  height: 90px;
  align-items: center;
  justify-content: center;
  color: #aaaaaa;
}

.market-card__image--1 {
  background: #f0ece8;
}

.market-card__image--2 {
  background: linear-gradient(135deg, #eeeeee, #dddddd);
}

.market-card__image--3 {
  background: linear-gradient(135deg, #e8eee8, #d0dcd0);
}

.market-card__image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.market-card__body {
  padding: 9px 11px;
}

.market-card h2 {
  margin: 0 0 2px;
  color: #1a1a1a;
  font-size: 12px;
  font-weight: 500;
}

.market-card p {
  margin: 0 0 5px;
  color: #777777;
  font-size: 10px;
}

.market-card strong {
  color: #1a1a1a;
  font-size: 14px;
  font-weight: 300;
}

.market-card__body div {
  display: flex;
  justify-content: space-between;
  border-top: 1px solid #f4f4f4;
  margin-top: 5px;
  padding-top: 5px;
  color: #777777;
  font-size: 10px;
}

.market-card__body div span:last-child {
  color: #000000;
}

.market-empty {
  border: 1px solid #e4e4e4;
  border-radius: 3px;
  background: #ffffff;
  color: #777777;
  font-size: 12px;
  padding: 40px 20px;
  text-align: center;
}

@media (max-width: 1023px) {
  .market-page {
    grid-template-columns: 1fr;
  }

  .market-filter {
    position: static;
    height: auto;
    border-right: 0;
    border-bottom: 1px solid #e4e4e4;
  }

  .market-grid {
    columns: 1;
  }
}
</style>

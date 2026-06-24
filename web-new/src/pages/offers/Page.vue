<!--
 * 綜合優惠頻道頁。
 * 1. 靜態 mock 資料呈現超市格價優惠列表。
 * 2. 提供搜尋、分類篩選、商店篩選與排序控制欄。
 * 3. 4 列卡片網格、網格 / 表格切換、分頁。
-->
<script setup lang="ts">
import { computed, ref } from 'vue';
import { useRouter } from 'vue-router';

import PaginationBar from '@/shared/components/navigation/PaginationBar.vue';

// 1. 分頁按鈕型別（對齊 PaginationBar 元件 PaginationPage 介面）
interface PaginationPage {
  label: string | number;
  active?: boolean;
  key: string | number;
}

// 2. 優惠卡片價格列型別
interface OfferPriceRow {
  store: string;
  offer: string;
  price: string;
  original: string;
}

// 3. 優惠卡片型別
interface OfferCard {
  code: string;
  store: string;
  name: string;
  brand: string;
  badge: string;
  favorite: boolean;
  prices: OfferPriceRow[];
}

// 4. HERO 統計型別
interface HeroStat {
  value: string;
  label: string;
}

const router = useRouter();

const heroStats: HeroStat[] = [
  { value: '2,122', label: '活躍優惠' },
  { value: '9', label: '連鎖商店' },
  { value: '-60%', label: '最高折扣' },
];

const categoryPills = [
  '全部',
  '個人護理',
  '粉麵食品',
  '飲品',
  '零食',
  '家居用品',
  '奶類',
  '清潔用品',
];

const storePills = [
  '全部商店',
  '惠康',
  '百佳',
  'AEON',
  'Market Place',
  'CitySuper',
  'Mannings',
];

const sortOptions = [
  '優惠力度 ↓',
  '優惠後價格',
  '價差',
  '商品名稱',
];

// 5. 靜態優惠卡片資料
const cards: OfferCard[] = [
  {
    code: 'toothbrush-soft-3pk',
    store: 'AEON',
    name: '纖柔牙刷 精巧頭 3支裝',
    brand: '高露潔',
    badge: '-60%',
    favorite: true,
    prices: [
      { store: 'AEON', offer: '買2件 $29.00', price: 'HK$14.50', original: 'HK$35.90' },
      { store: '百佳', offer: '未提供優惠文案', price: 'HK$22.90', original: 'HK$35.90' },
      { store: '惠康', offer: '未提供優惠文案', price: 'HK$26.50', original: 'HK$35.90' },
    ],
  },
  {
    code: 'shiraz-cabernet-750ml',
    store: '惠康 · Market Place',
    name: 'Shiraz Cabernet 紅酒 750ml',
    brand: '羅遜氏',
    badge: '-60%',
    favorite: false,
    prices: [
      { store: '惠康', offer: '$80任揀2件', price: 'HK$40.00', original: 'HK$99.00' },
      { store: 'Market Place', offer: '買6件享85折', price: 'HK$42.00', original: 'HK$99.00' },
    ],
  },
  {
    code: 'seafood-udon-5pk',
    store: '百佳',
    name: '即食海鮮烏冬 5包裝',
    brand: '出前一丁',
    badge: '-57%',
    favorite: false,
    prices: [
      { store: '百佳', offer: '買2件 $25.80', price: 'HK$12.90', original: 'HK$29.90' },
      { store: 'AEON', offer: '會員優惠', price: 'HK$15.20', original: 'HK$29.90' },
    ],
  },
  {
    code: 'tissue-200pull-4box',
    store: 'AEON',
    name: '純棉面巾紙 200抽 4盒裝',
    brand: '維達',
    badge: '-56%',
    favorite: false,
    prices: [
      { store: 'AEON', offer: '買3件 $59.70', price: 'HK$19.90', original: 'HK$45.50' },
      { store: '惠康', offer: '第二件半價', price: 'HK$22.40', original: 'HK$45.50' },
    ],
  },
];

const activeCategory = ref('全部');
const activeStore = ref('全部商店');
const activeSort = ref(sortOptions[0]);
const searchQuery = ref('');
const viewMode = ref<'grid' | 'table'>('grid');
const currentPage = ref(1);
const totalPages = 3;
const totalCount = '2,122';

// 6. 分頁按鈕陣列
const paginationPages = computed<PaginationPage[]>(() => {
  const pages: PaginationPage[] = [
    { label: '上一頁', key: 'prev' },
  ];
  for (let page = 1; page <= totalPages; page += 1) {
    pages.push({
      label: page,
      key: page,
      active: page === currentPage.value,
    });
  }
  pages.push({ label: '下一頁', key: 'next' });
  return pages;
});

const paginationInfo = `第 ${currentPage.value} 頁，共 ${totalPages} 頁 · ${totalCount} 個優惠`;

// 7. 切換分類篩選
const selectCategory = (category: string): void => {
  activeCategory.value = category;
};

// 8. 切換商店篩選
const selectStore = (store: string): void => {
  activeStore.value = store;
};

// 9. 切換收藏
const toggleFavorite = (card: OfferCard): void => {
  card.favorite = !card.favorite;
};

// 10. 切換分頁
const handlePageSelect = (page: PaginationPage): void => {
  if (page.key === 'prev') {
    if (currentPage.value > 1) currentPage.value -= 1;
    return;
  }
  if (page.key === 'next') {
    if (currentPage.value < totalPages) currentPage.value += 1;
    return;
  }
  currentPage.value = Number(page.key);
};

// 11. 開啟商品詳情
const openDetail = (card: OfferCard): void => {
  void router.push({ path: `/supermarket-offers/products/${encodeURIComponent(card.code)}` });
};
</script>

<template>
  <main class="gp-page">
    <!-- 1. 暗色 HERO -->
    <section class="gp-hero">
      <div class="gp-hero-left">
        <div class="gp-hero-label">超市格價</div>
        <div class="gp-hero-title">今日最抵Deal</div>
        <div class="gp-hero-sub">追蹤全港主要超市即時優惠，比較優惠後價格與折扣力度。</div>
      </div>
      <div class="gp-hero-right">
        <template
          v-for="(stat, index) in heroStats"
          :key="stat.label"
        >
          <div class="gp-hstat">
            <span class="gp-hnum">{{ stat.value }}</span>
            <span class="gp-hlabel">{{ stat.label }}</span>
          </div>
          <i
            v-if="index < heroStats.length - 1"
            class="gp-hdiv"
          />
        </template>
      </div>
    </section>

    <!-- 2. 控制欄 -->
    <section class="gp-controls">
      <div class="gp-search-row">
        <div class="gp-search-box">
          <svg
            class="gp-search-ico"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <circle cx="11" cy="11" r="8" />
            <line x1="21" y1="21" x2="16.65" y2="16.65" />
          </svg>
          <input
            v-model="searchQuery"
            class="gp-sinput"
            placeholder="搜尋商品、品牌或分類..."
          >
        </div>
        <button
          type="button"
          class="gp-search-btn"
        >搜尋</button>
        <button
          type="button"
          class="gp-fav-btn"
        >我的收藏</button>
      </div>
      <div class="gp-filter-pills">
        <button
          v-for="category in categoryPills"
          :key="category"
          type="button"
          class="gp-fpill"
          :class="activeCategory === category ? 'on' : ''"
          @click="selectCategory(category)"
        >{{ category }}</button>
      </div>
      <div class="gp-store-pills">
        <button
          v-for="store in storePills"
          :key="store"
          type="button"
          class="gp-spill"
          :class="activeStore === store ? 'on' : ''"
          @click="selectStore(store)"
        >{{ store }}</button>
      </div>
    </section>

    <!-- 3. 內容區 -->
    <section class="gp-content">
      <div class="gp-content-header">
        <span class="gp-count">{{ totalCount }} 個優惠</span>
        <div class="gp-toolbar">
          <div class="gp-view-toggle">
            <button
              type="button"
              class="gp-view-btn"
              :class="viewMode === 'grid' ? 'on' : ''"
              @click="viewMode = 'grid'"
            >網格</button>
            <button
              type="button"
              class="gp-view-btn"
              :class="viewMode === 'table' ? 'on' : ''"
              @click="viewMode = 'table'"
            >表格</button>
          </div>
          <select
            v-model="activeSort"
            class="gp-sort"
          >
            <option
              v-for="option in sortOptions"
              :key="option"
              :value="option"
            >{{ option }}</option>
          </select>
        </div>
      </div>

      <!-- 3.1 網格視圖 -->
      <div
        v-if="viewMode === 'grid'"
        class="gp-grid"
      >
        <article
          v-for="card in cards"
          :key="card.code"
          class="gp-card"
          @click="openDetail(card)"
        >
          <button
            type="button"
            class="gp-card-fav"
            :class="card.favorite ? 'on' : ''"
            @click.stop="toggleFavorite(card)"
          >{{ card.favorite ? '已收藏' : '收藏' }}</button>
          <div class="gp-card-img">
            <div class="gp-card-img-placeholder" />
          </div>
          <div class="gp-card-body">
            <div class="gp-card-heading">
              <div class="gp-card-badge">{{ card.badge }}</div>
              <div class="gp-card-store">{{ card.store }}</div>
              <div class="gp-card-name">{{ card.name }}</div>
              <div class="gp-card-brand">{{ card.brand }}</div>
            </div>
            <div class="gp-card-prices">
              <div
                v-for="(price, index) in card.prices"
                :key="`${card.code}-${price.store}-${index}`"
                class="gp-price-row"
              >
                <div>
                  <strong>{{ price.store }}</strong>
                  <small>{{ price.offer }}</small>
                </div>
                <div>
                  <b>{{ price.price }}</b>
                  <span>{{ price.original }}</span>
                </div>
              </div>
            </div>
          </div>
        </article>
      </div>

      <!-- 3.2 表格視圖 -->
      <div
        v-else
        class="gp-table"
      >
        <table>
          <thead>
            <tr>
              <th>商品</th>
              <th>商店</th>
              <th>優惠</th>
              <th>優惠後</th>
              <th>原價</th>
              <th>收藏</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="card in cards"
              :key="card.code"
              @click="openDetail(card)"
            >
              <td>
                <strong>{{ card.name }}</strong>
                <span>{{ card.brand }}</span>
              </td>
              <td>{{ card.prices[0]?.store ?? '-' }}</td>
              <td>{{ card.prices[0]?.offer ?? '-' }}</td>
              <td>{{ card.prices[0]?.price ?? '-' }}</td>
              <td>{{ card.prices[0]?.original ?? '-' }}</td>
              <td>
                <button
                  type="button"
                  class="gp-table-fav"
                  @click.stop="toggleFavorite(card)"
                >{{ card.favorite ? '已收藏' : '收藏' }}</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- 3.3 分頁 -->
      <PaginationBar
        :info="paginationInfo"
        :pages="paginationPages"
        aria-label="綜合優惠分頁"
        @select="handlePageSelect"
      />

      <div class="gp-updated-bar">
        資料更新：2026年06月05日 · 共監測 2,555 件商品 · 價格只供參考，實際售價以商戶公布為準。
      </div>
    </section>
  </main>
</template>

<style scoped>
/* 1. 頁面容器 */
.gp-page {
  width: 100%;
  max-width: var(--layout-page-max-width);
  margin: 0 auto;
  background: rgb(var(--color-surface-2));
  color: rgb(var(--color-text));
}

/* 2. 暗色 HERO */
.gp-hero {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
  min-height: 164px;
  padding: 34px 38px;
  border-bottom: 3px solid rgb(var(--color-primary));
  background: #1a1a1a;
  color: #ffffff;
}

.gp-hero-label {
  color: rgba(255, 255, 255, 0.58);
  font-size: 12px;
  letter-spacing: 0;
}

.gp-hero-title {
  margin-top: 8px;
  font-family: var(--font-display);
  font-size: 34px;
  font-weight: 400;
  line-height: 1.12;
  color: #ffffff;
}

.gp-hero-sub {
  margin-top: 8px;
  max-width: 420px;
  color: rgba(255, 255, 255, 0.68);
  font-size: 13px;
  line-height: 1.6;
}

.gp-hero-right {
  display: flex;
  align-items: center;
  gap: 0;
  flex-shrink: 0;
}

.gp-hstat {
  min-width: 118px;
  padding: 0 26px;
  text-align: center;
}

.gp-hnum {
  display: block;
  color: #ffffff;
  font-size: 28px;
  font-weight: 300;
  line-height: 1;
}

.gp-hlabel {
  display: block;
  margin-top: 6px;
  color: rgba(255, 255, 255, 0.68);
  font-size: 11px;
}

.gp-hdiv {
  display: block;
  width: 1px;
  height: 42px;
  background: rgba(255, 255, 255, 0.26);
}

/* 3. 控制欄 */
.gp-controls {
  padding: 18px 28px 14px;
  border-bottom: 1px solid rgb(var(--color-border));
  background: rgb(var(--color-surface));
}

.gp-search-row {
  display: flex;
  align-items: stretch;
  gap: 8px;
  margin-bottom: 12px;
}

.gp-search-box {
  display: flex;
  max-width: 520px;
  align-items: center;
  gap: 8px;
  flex: 0 1 520px;
  border: 1px solid rgb(var(--color-border));
  border-radius: 3px;
  background: rgb(var(--color-surface));
  padding: 0 13px;
}

.gp-search-ico {
  width: 18px;
  height: 18px;
  color: rgb(var(--color-ink-3));
  flex-shrink: 0;
}

.gp-sinput {
  flex: 1;
  min-width: 0;
  border: 0;
  background: transparent;
  font: inherit;
  font-size: 13px;
  color: rgb(var(--color-text));
  outline: 0;
  padding: 12px 0;
}

.gp-search-btn,
.gp-fav-btn {
  border: 1px solid rgb(var(--color-border));
  border-radius: 3px;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text));
  cursor: pointer;
  font-family: inherit;
  font-size: 12px;
  font-weight: 600;
  padding: 0 18px;
}

.gp-search-btn {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary));
  color: #ffffff;
}

.gp-fav-btn.on {
  border-color: rgb(var(--color-brand-mid));
  background: rgb(var(--color-primary-soft));
  color: rgb(var(--color-primary));
}

.gp-filter-pills,
.gp-store-pills {
  display: flex;
  flex-wrap: wrap;
  gap: 7px;
  margin-top: 8px;
}

.gp-fpill,
.gp-spill {
  border: 1px solid rgb(var(--color-border));
  background: rgb(var(--color-surface));
  color: rgb(var(--color-ink-3));
  cursor: pointer;
  font-family: inherit;
  font-size: 12px;
  padding: 5px 13px;
}

.gp-fpill {
  border-radius: 999px;
}

.gp-spill {
  border-radius: 3px;
}

.gp-fpill.on,
.gp-spill.on {
  border-color: rgb(var(--color-brand-mid));
  background: rgb(var(--color-primary-soft));
  color: rgb(var(--color-primary));
  font-weight: 500;
}

/* 4. 內容區 */
.gp-content {
  padding: 26px 28px 44px;
  background: rgb(var(--color-surface-2));
}

.gp-content-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 20px;
}

.gp-count {
  color: rgb(var(--color-ink-3));
  font-size: 12px;
  font-weight: 400;
}

.gp-toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
}

.gp-view-toggle {
  display: flex;
  gap: 6px;
}

.gp-view-btn {
  border: 1px solid rgb(var(--color-border));
  border-radius: 2px;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-ink-3));
  cursor: pointer;
  font-family: inherit;
  font-size: 12px;
  font-weight: 600;
  padding: 7px 13px;
}

.gp-view-btn.on {
  border-color: rgb(var(--color-border));
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text));
}

.gp-sort {
  border: 1px solid rgb(var(--color-border));
  border-radius: 2px;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text));
  font-family: inherit;
  font-size: 12px;
  padding: 7px 26px 7px 10px;
  outline: 0;
}

/* 5. 卡片網格 */
.gp-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 14px;
}

.gp-card {
  position: relative;
  min-height: 420px;
  border: 1px solid rgb(var(--color-border));
  border-radius: 3px;
  overflow: hidden;
  background: rgb(var(--color-surface));
  cursor: pointer;
  transition: border-color 0.15s ease, box-shadow 0.15s ease;
}

.gp-card:hover {
  border-color: rgb(var(--color-primary));
  box-shadow: var(--shadow-raised);
}

.gp-card-img {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 140px;
  border-bottom: 1px solid rgb(var(--color-border));
  background: rgb(var(--color-surface-2));
  overflow: hidden;
}

.gp-card-img-placeholder {
  width: 64px;
  height: 64px;
  border-radius: 50%;
  background: linear-gradient(135deg, rgb(var(--color-brand-mid)), rgb(var(--color-primary)));
  opacity: 0.6;
}

.gp-card-fav {
  position: absolute;
  top: 14px;
  right: 14px;
  z-index: 2;
  border: 1px solid rgb(var(--color-border));
  border-radius: 2px;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-ink-3));
  cursor: pointer;
  font-family: inherit;
  font-size: 11px;
  font-weight: 400;
  padding: 6px 10px;
}

.gp-card-fav.on {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary-soft));
  color: rgb(var(--color-primary));
}

.gp-card-body {
  padding: 14px;
}

.gp-card-heading {
  position: relative;
  margin-bottom: 12px;
  padding-right: 0;
}

.gp-card-badge {
  position: absolute;
  top: 0;
  left: 0;
  display: inline-flex;
  align-items: center;
  border-radius: 2px;
  background: rgb(var(--color-primary));
  color: #ffffff;
  font-size: 11px;
  font-weight: 700;
  line-height: 1;
  padding: 5px 7px;
  white-space: nowrap;
}

.gp-card-store {
  margin: 0 0 6px;
  padding-left: 48px;
  color: rgb(var(--color-ink-3));
  font-size: 12px;
  font-weight: 400;
}

.gp-card-name {
  margin: 0;
  color: rgb(var(--color-text));
  font-size: 15px;
  font-weight: 700;
  line-height: 1.28;
}

.gp-card-brand {
  display: none;
}

.gp-card-prices {
  display: grid;
  gap: 0;
  margin-top: 10px;
}

.gp-price-row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 10px;
  border-top: 1px solid rgb(var(--color-surface-3));
  padding: 10px 0;
}

.gp-price-row:first-child {
  border-top: 0;
  padding-top: 0;
}

.gp-price-row strong {
  display: block;
  color: rgb(var(--color-text));
  font-size: 13px;
  font-weight: 700;
}

.gp-price-row small {
  display: block;
  margin-top: 4px;
  color: rgb(var(--color-ink-3));
  font-size: 11px;
  line-height: 1.4;
}

.gp-price-row b {
  display: block;
  color: rgb(var(--color-text));
  font-size: 14px;
  text-align: right;
  white-space: nowrap;
}

.gp-price-row span {
  display: block;
  margin-top: 3px;
  color: rgb(var(--color-primary));
  font-size: 11px;
  text-align: right;
  white-space: nowrap;
}

/* 6. 表格視圖 */
.gp-table {
  overflow: auto;
  border: 1px solid rgb(var(--color-border));
  border-radius: 3px;
  background: rgb(var(--color-surface));
}

.gp-table table {
  width: 100%;
  border-collapse: collapse;
  min-width: 760px;
}

.gp-table th {
  border-bottom: 1px solid rgb(var(--color-border));
  color: rgb(var(--color-ink-3));
  font-size: 12px;
  font-weight: 600;
  padding: 12px;
  text-align: left;
}

.gp-table td {
  border-bottom: 1px solid rgb(var(--color-surface-3));
  color: rgb(var(--color-ink-2));
  font-size: 12px;
  padding: 12px;
}

.gp-table tr {
  cursor: pointer;
}

.gp-table tr:hover td {
  background: rgb(var(--color-primary-soft));
}

.gp-table strong {
  display: block;
  color: rgb(var(--color-text));
  font-size: 13px;
}

.gp-table span {
  display: block;
  margin-top: 3px;
  color: rgb(var(--color-ink-3));
  font-size: 11px;
}

.gp-table-fav {
  border: 1px solid rgb(var(--color-border));
  border-radius: 2px;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-ink-2));
  cursor: pointer;
  font-family: inherit;
  font-size: 11px;
  font-weight: 600;
  padding: 5px 9px;
}

/* 7. 更新資訊列 */
.gp-updated-bar {
  margin-top: 18px;
  color: rgb(var(--color-ink-3));
  font-size: 11px;
  text-align: center;
}

/* 8. 響應式 */
@media (max-width: 1023px) {
  .gp-hero {
    flex-direction: column;
    align-items: flex-start;
    padding: 22px;
  }

  .gp-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 640px) {
  .gp-hero {
    padding: 18px;
  }

  .gp-controls {
    padding: 14px;
  }

  .gp-content {
    padding: 18px 14px 36px;
  }

  .gp-search-row {
    flex-direction: column;
    align-items: stretch;
  }

  .gp-search-box {
    max-width: none;
    flex: 1 1 auto;
  }

  .gp-content-header {
    flex-direction: column;
    align-items: flex-start;
  }

  .gp-grid {
    grid-template-columns: 1fr;
  }

  .gp-hero-right {
    flex-wrap: wrap;
    gap: 12px;
  }

  .gp-hstat {
    min-width: 0;
    padding: 0 12px;
  }

  .gp-hdiv {
    display: none;
  }
}
</style>

<!--
 * 家具市集列表頁。
 * 1. 三欄布局：左側篩選欄 + 中間列表區 + 右側廣告欄。
 * 2. 左側含搜尋框與 5 組篩選標籤（分類、價格範圍、成色、地區、可見範圍）。
 * 3. 中間含排序欄、商品卡片網格（懸停顯示操作按鈕）與分頁。
 * 4. 右側含 3 個 banner 廣告與 2 個 vertical 廣告。
 * 5. 全部使用靜態 mock 資料，不呼叫 API。
-->
<script setup lang="ts">
import { ref } from 'vue';
import { useRouter } from 'vue-router';

import FilterTag from '@/shared/components/base/FilterTag.vue';
import PaginationBar from '@/shared/components/navigation/PaginationBar.vue';
import AdCard from '@/shared/components/marketplace/AdCard.vue';

// 1. 路由
const router = useRouter();

// 2. 搜尋關鍵字
const keyword = ref('');

// 3. 排序選項
const sortBy = ref('latest');

// 4. 篩選群組資料
interface FilterOption {
  label: string;
  value: string;
}
interface FilterGroup {
  key: string;
  title: string;
  options: FilterOption[];
  activeValue: string;
}
const filterGroups = ref<FilterGroup[]>([
  {
    key: 'category',
    title: '分類',
    options: [
      { label: '全部', value: 'all' },
      { label: '家居傢俱', value: 'home_furniture' },
      { label: '家庭電器', value: 'home_appliance' },
      { label: '電子產品', value: 'electronics' },
      { label: 'BB用品', value: 'baby' },
      { label: '其他', value: 'other' },
    ],
    activeValue: 'all',
  },
  {
    key: 'price',
    title: '價格範圍',
    options: [
      { label: '全部', value: 'all' },
      { label: '$1k以下', value: 'under_1k' },
      { label: '$1k–3k', value: '1k_3k' },
      { label: '$3k–5k', value: '3k_5k' },
      { label: '$10k+', value: 'over_10k' },
    ],
    activeValue: 'all',
  },
  {
    key: 'condition',
    title: '成色',
    options: [
      { label: '全部', value: 'all' },
      { label: '近乎全新', value: 'like_new' },
      { label: '良好', value: 'good' },
      { label: '尚可', value: 'fair' },
    ],
    activeValue: 'all',
  },
  {
    key: 'region',
    title: '地區',
    options: [
      { label: '全部', value: 'all' },
      { label: '香港島', value: 'hk_island' },
      { label: '九龍', value: 'kowloon' },
      { label: '新界', value: 'nt' },
      { label: '離島', value: 'islands' },
    ],
    activeValue: 'all',
  },
  {
    key: 'visibility',
    title: '可見範圍',
    options: [
      { label: '全部', value: 'all' },
      { label: '公開', value: 'public' },
      { label: '同棟可見', value: 'building' },
    ],
    activeValue: 'all',
  },
]);

// 5. 家具卡片 mock 資料
interface FurnitureCard {
  id: number;
  isNew?: boolean;
  hasPattern: boolean;
  imageBg: string;
  name: string;
  sub: string;
  price: string;
  foot: string;
}
const cards = ref<FurnitureCard[]>([
  {
    id: 1,
    isNew: true,
    hasPattern: false,
    imageBg: '#f0ece8',
    name: '三色短毛貓領養',
    sub: '香港島 · 黃埔 · 良好',
    price: 'HK$3',
    foot: '其他 · 28/03',
  },
  {
    id: 2,
    hasPattern: true,
    imageBg: 'linear-gradient(135deg,#eee,#ddd)',
    name: '北歐實木餐桌',
    sub: '九龍 · 旺角 · 近乎全新',
    price: 'HK$2,400',
    foot: '傢俱 · 02/04',
  },
  {
    id: 3,
    hasPattern: true,
    imageBg: 'linear-gradient(135deg,#e8eee8,#d0dcd0)',
    name: 'LG 洗衣機 8kg',
    sub: '新界 · 沙田 · 良好',
    price: 'HK$1,800',
    foot: '電器 · 01/04',
  },
]);

// 6. 分頁資料
const paginationPages = ref([
  { label: '上一頁', key: 'prev' },
  { label: 1, active: true, key: 1 },
  { label: 2, key: 2 },
  { label: 3, key: 3 },
  { label: '下一頁', key: 'next' },
]);

// 7. 切換篩選標籤（同組互斥）
const handleFilterToggle = (groupKey: string, optionValue: string) => {
  const group = filterGroups.value.find((g) => g.key === groupKey);
  if (group) {
    group.activeValue = optionValue;
  }
};

// 8. 點擊卡片跳轉詳情
const handleCardClick = (id: number) => {
  void router.push(`/furniture/${id}`);
};

// 9. 點擊分頁
const handlePageSelect = () => {
  // mock：不實作實際分頁邏輯
};
</script>

<template>
  <div class="page">
    <div class="mp">
      <!-- 左側篩選欄 -->
      <aside class="mf">
        <div class="sbar">
          <input
            v-model="keyword"
            class="sinput"
            placeholder="搜尋物品…"
            autocomplete="off"
          />
          <button
            type="button"
            class="sbtn"
          >搜</button>
        </div>

        <section
          v-for="group in filterGroups"
          :key="group.key"
          class="fs"
        >
          <div class="ft-title">{{ group.title }}</div>
          <div class="ftags">
            <FilterTag
              v-for="option in group.options"
              :key="`${group.key}-${option.value}`"
              :label="option.label"
              :active="option.value === group.activeValue"
              @toggle="handleFilterToggle(group.key, option.value)"
            />
          </div>
        </section>
      </aside>

      <!-- 中間列表區 -->
      <main class="mr">
        <button
          type="button"
          class="filter-toggle-btn"
        >
          <svg
            width="14"
            height="14"
            viewBox="0 0 14 14"
            fill="none"
            xmlns="http://www.w3.org/2000/svg"
          >
            <path
              d="M1 2h12M1 7h12M1 12h12"
              stroke="currentColor"
              stroke-width="1.5"
              stroke-linecap="round"
            />
          </svg>
          篩選條件
        </button>

        <div class="sort-row">
          <span class="rn">3 個結果</span>
          <select
            v-model="sortBy"
            class="ssel"
          >
            <option value="latest">最新發佈</option>
            <option value="price_asc">價格低至高</option>
          </select>
        </div>

        <div class="mgrid">
          <div
            v-for="card in cards"
            :key="card.id"
            class="mc"
            @click="handleCardClick(card.id)"
          >
            <span
              v-if="card.isNew"
              class="mnew"
            >全新</span>
            <div
              class="mimg"
              :class="card.hasPattern ? 'pat' : ''"
              :style="{ background: card.imageBg }"
            ></div>
            <div class="mbody">
              <div class="mname">{{ card.name }}</div>
              <div class="msub">{{ card.sub }}</div>
              <div class="mprice">{{ card.price }}</div>
              <div class="mfoot"><span>{{ card.foot }}</span></div>
            </div>
            <div class="mc-actions">
              <button
                type="button"
                class="mc-action-btn"
                @click.stop
              >收藏</button>
              <button
                type="button"
                class="mc-action-btn"
                @click.stop
              >比較</button>
              <button
                type="button"
                class="mc-action-btn primary"
                @click.stop="handleCardClick(card.id)"
              >查看</button>
            </div>
          </div>
        </div>

        <PaginationBar
          info="第 1-3 筆，共 18 筆"
          :pages="paginationPages"
          aria-label="家具市集分頁"
          @select="handlePageSelect"
        />
      </main>

      <!-- 右側廣告欄 -->
      <aside class="market-ad-aside">
        <div class="ad-side-panel">
          <AdCard
            variant="banner"
            visual="market"
            label="16:9"
            title="社區家具回收"
            desc="大件家具回收、清拆與轉售安排。"
          />
          <AdCard
            variant="banner"
            visual="home"
            label="16:9"
            title="精選家居用品"
            desc="收納、餐桌、燈具與日常設備。"
          />
          <AdCard
            variant="banner"
            visual="office"
            label="16:9"
            title="電器保養服務"
            desc="洗衣機、雪櫃與冷氣維修預約。"
          />
          <div class="ad-vertical-grid">
            <AdCard
              variant="vertical"
              visual="service"
              label="9:16"
              title="上門安裝"
              desc="窗簾、層架與燈具安裝。"
            />
            <AdCard
              variant="vertical"
              visual="move"
              label="9:16"
              title="即日配送"
              desc="同區交收與預約送貨。"
            />
          </div>
        </div>
      </aside>
    </div>
  </div>
</template>

<style scoped>
/* 1. 頁面容器 */
.page {
  width: 100%;
}

/* 2. 三欄布局 */
.mp {
  display: grid;
  grid-template-columns: 250px minmax(0, 700px) 360px;
  justify-content: center;
  max-width: 1440px;
  margin: 0 auto;
  background: rgb(var(--color-surface-2));
  min-height: calc(100vh - var(--nav-h, 52px));
}

/* 3. 左側篩選欄 */
.mf {
  background: rgb(var(--color-surface));
  border-right: 1px solid rgb(var(--color-border));
  padding: 18px 18px;
  position: sticky;
  top: var(--nav-h, 52px);
  height: calc(100vh - var(--nav-h, 52px));
  overflow-y: auto;
  min-height: calc(100vh - var(--nav-h, 52px));
  scrollbar-width: none;
}

.mf::-webkit-scrollbar {
  display: none;
}

.sbar {
  display: flex;
  gap: 6px;
  margin-bottom: 14px;
}

.sinput {
  flex: 1;
  min-width: 0;
  width: 100%;
  border: 1px solid rgb(var(--color-border));
  padding: 8px 10px;
  font-size: 12px;
  font-family: inherit;
  outline: none;
  border-radius: 2px;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text));
}

.sinput:focus {
  border-color: rgb(var(--color-primary));
}

.sbtn {
  background: rgb(var(--color-primary));
  color: #fff;
  border: none;
  padding: 8px 12px;
  font-size: 12px;
  cursor: pointer;
  border-radius: 2px;
  font-family: inherit;
}

.fs {
  margin-bottom: 16px;
}

.ft-title {
  font-size: 9px;
  letter-spacing: 2px;
  color: rgb(var(--color-ink-3));
  text-transform: uppercase;
  margin-bottom: 6px;
}

.ftags {
  display: flex;
  flex-wrap: wrap;
  gap: 5px;
  max-width: 100%;
  overflow: hidden;
}

/* 4. 中間列表區 */
.mr {
  background: rgb(var(--color-surface-2));
  padding: 14px 14px 36px;
  min-width: 0;
  min-height: calc(100vh - var(--nav-h, 52px));
}

.filter-toggle-btn {
  display: none;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: rgb(var(--color-text));
  background: rgb(var(--color-surface-2));
  border: 1px solid rgb(var(--color-border));
  padding: 8px 14px;
  border-radius: 2px;
  cursor: pointer;
  font-family: inherit;
  margin-bottom: 12px;
  width: 100%;
}

.sort-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
  max-width: 700px;
  margin-left: 0;
  margin-right: auto;
}

.rn {
  font-size: 11px;
  color: rgb(var(--color-ink-3));
}

.ssel {
  border: 1px solid rgb(var(--color-border));
  padding: 5px 8px;
  font-size: 11px;
  font-family: inherit;
  outline: none;
  border-radius: 2px;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text));
}

/* 5. 商品卡片網格 */
.mgrid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
  max-width: 700px;
  margin-left: 0;
  margin-right: auto;
}

.mc {
  position: relative;
  margin-bottom: 0;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  overflow: hidden;
  cursor: pointer;
  background: rgb(var(--color-surface));
  transition: box-shadow 0.15s, border-color 0.15s;
}

.mc:hover {
  box-shadow: var(--shadow-raised);
  border-color: rgb(var(--color-brand-mid));
}

.mnew {
  position: absolute;
  top: 7px;
  left: 7px;
  z-index: 2;
  background: rgb(var(--color-text));
  color: rgb(var(--color-surface));
  font-size: 9px;
  padding: 1px 5px;
  letter-spacing: 0.3px;
}

.mimg {
  height: 160px;
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* 6. 斜紋圖案 */
.pat {
  position: relative;
  overflow: hidden;
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

.mbody {
  padding: 9px 11px;
  padding-bottom: 9px;
}

.mname {
  font-size: 12px;
  font-weight: 500;
  margin-bottom: 2px;
  color: rgb(var(--color-text));
}

.msub {
  font-size: 10px;
  color: rgb(var(--color-ink-3));
  margin-bottom: 5px;
}

.mprice {
  font-size: 14px;
  font-weight: 300;
  color: rgb(var(--color-text));
}

.mfoot {
  display: none;
}

/* 7. 懸停操作按鈕 */
.mc-actions {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 8px;
  height: auto;
  max-height: 0;
  overflow: hidden;
  opacity: 0;
  pointer-events: none;
  transform: translateY(-2px);
  padding: 0 12px;
  transition: max-height 0.18s ease, opacity 0.15s ease, transform 0.15s ease, padding 0.18s ease;
}

.mc:hover .mc-actions {
  max-height: 54px;
  opacity: 1;
  pointer-events: auto;
  transform: translateY(0);
  padding: 0 12px 12px;
}

.mc-action-btn {
  min-height: 34px;
  border: 1px solid rgb(var(--color-border));
  border-radius: 6px;
  background: #fff;
  color: rgb(var(--color-ink-2));
  cursor: pointer;
  font-family: inherit;
  font-size: 12px;
  font-weight: 500;
}

.mc-action-btn:hover {
  border-color: rgb(var(--color-brand-mid));
  color: rgb(var(--color-primary));
}

.mc-action-btn.primary {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary));
  color: #fff;
}

.mc-action-btn.primary:hover {
  background: rgb(var(--color-brand-dark));
  color: #fff;
}

/* 8. 右側廣告欄 */
.market-ad-aside {
  position: sticky;
  top: var(--nav-h, 52px);
  align-self: start;
  overflow: visible;
  border-left: 1px solid rgb(var(--color-border));
  background: rgb(var(--color-surface-2));
  padding: 18px 18px 40px;
}

.ad-side-panel {
  display: grid;
  gap: 12px;
}

.ad-vertical-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 12px;
}

/* 9. 響應式 */
@media (max-width: 1100px) {
  .mp {
    grid-template-columns: 210px 1fr;
  }

  .market-ad-aside {
    display: none;
  }
}

@media (max-width: 767px) {
  .mp {
    grid-template-columns: 1fr;
    max-width: none;
  }

  .mf {
    position: static;
    height: auto;
    padding: 14px;
  }

  .filter-toggle-btn {
    display: inline-flex;
  }

  .mgrid,
  .sort-row {
    max-width: none;
  }

  .mgrid {
    grid-template-columns: 1fr;
  }
}
</style>

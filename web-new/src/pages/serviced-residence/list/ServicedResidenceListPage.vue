<!--
 * 服務式住宅列表頁。
 * 1. 暗色 HERO 區：標題、描述與品牌定位。
 * 2. 篩選欄：關鍵字搜尋、地區與租期快速篩選、排序。
 * 3. 卡片網格：3 欄卡片，含圖片、標籤、標題、簡介、價格與設施。
 * 4. 分頁：使用 PaginationBar 元件。
 * 5. 全部使用靜態 mock 資料，不呼叫 API。
-->
<script setup lang="ts">
import { ref } from 'vue';
import { useRouter } from 'vue-router';

import PaginationBar from '@/shared/components/navigation/PaginationBar.vue';

// 1. 路由
const router = useRouter();

// 2. 搜尋關鍵字
const keyword = ref('');

// 3. 排序選項
const sortBy = ref('latest');

// 4. 地區快速篩選
interface FilterPill {
  label: string;
  active: boolean;
}
const regionFilters = ref<FilterPill[]>([
  { label: '全港', active: true },
  { label: '香港島', active: false },
  { label: '九龍', active: false },
  { label: '新界', active: false },
  { label: '離島', active: false },
]);

// 5. 租期快速篩選
const leaseFilters = ref<FilterPill[]>([
  { label: '每日', active: false },
  { label: '每週', active: false },
  { label: '每月', active: false },
]);

// 6. 卡片 mock 資料
interface ServicedResidenceCard {
  id: number;
  imageBg: string;
  tags: { label: string; dark?: boolean }[];
  title: string;
  sub: string;
  price: string;
  priceUnit: string;
  pills: string[];
  rooms: string;
}
const cards = ref<ServicedResidenceCard[]>([
  {
    id: 1,
    imageBg: 'linear-gradient(160deg,#dde4e8,#b8c8d0)',
    tags: [
      { label: '每日起租', dark: true },
      { label: '中環' },
    ],
    title: '中環服務式公寓 Superior',
    sub: '780呎 · 獨立廚房 · 健身室 · 24hr禮賓',
    price: 'HK$1,800',
    priceUnit: '/ 晚起',
    pills: ['WiFi', '清潔服務', '洗衣房', '停車場'],
    rooms: '1房 1廁',
  },
  {
    id: 2,
    imageBg: 'linear-gradient(160deg,#e8e4d8,#d0c8b0)',
    tags: [
      { label: '每月起租', dark: true },
      { label: '尖沙咀' },
    ],
    title: '尖沙咀豪華套房 Deluxe',
    sub: '520呎 · 海景 · 游泳池 · 管家服務',
    price: 'HK$28,000',
    priceUnit: '/ 月起',
    pills: ['海景', '泳池', '健身室', '餐廳'],
    rooms: '開放式',
  },
  {
    id: 3,
    imageBg: 'linear-gradient(160deg,#e0e8e0,#c0d4c0)',
    tags: [{ label: '銅鑼灣' }],
    title: '銅鑼灣家庭套房 Family',
    sub: '1,200呎 · 2房 · 近學校 · 超市',
    price: 'HK$45,000',
    priceUnit: '/ 月起',
    pills: ['兒童設施', '學區', '私人陽台'],
    rooms: '2房 2廁',
  },
  {
    id: 4,
    imageBg: 'linear-gradient(160deg,#dcdce8,#c0c0d8)',
    tags: [
      { label: '每日起租', dark: true },
      { label: '上環' },
    ],
    title: '上環精品公寓 Studio',
    sub: '420呎 · 獨立廚房 · 近地鐵站',
    price: 'HK$1,500',
    priceUnit: '/ 晚起',
    pills: ['WiFi', '清潔服務', '健身室'],
    rooms: '開放式',
  },
  {
    id: 5,
    imageBg: 'linear-gradient(160deg,#e8dce0,#d0bcc4)',
    tags: [
      { label: '每月起租', dark: true },
      { label: '旺角' },
    ],
    title: '旺角商務套房 Executive',
    sub: '650呎 · 商務中心 · 會議室 · 快速網路',
    price: 'HK$22,000',
    priceUnit: '/ 月起',
    pills: ['商務中心', '會議室', 'WiFi', '停車場'],
    rooms: '1房 1廁',
  },
  {
    id: 6,
    imageBg: 'linear-gradient(160deg,#dce4e8,#bcccd4)',
    tags: [
      { label: '每月起租', dark: true },
      { label: '將軍澳' },
    ],
    title: '將軍澳海景公寓 Premier',
    sub: '900呎 · 海景 · 會所設施 · 近海濱長廊',
    price: 'HK$30,000',
    priceUnit: '/ 月起',
    pills: ['海景', '會所', '泳池', '健身室'],
    rooms: '2房 2廁',
  },
]);

// 7. 分頁資料
const paginationPages = ref([
  { label: '上一頁', key: 'prev' },
  { label: 1, active: true, key: 1 },
  { label: 2, key: 2 },
  { label: 3, key: 3 },
  { label: '下一頁', key: 'next' },
]);

// 8. 切換地區篩選（互斥）
const handleRegionToggle = (item: FilterPill) => {
  regionFilters.value.forEach((pill) => {
    pill.active = pill.label === item.label;
  });
};

// 9. 切換租期篩選（多選）
const toggleLeaseFilter = (item: FilterPill) => {
  item.active = !item.active;
};

// 10. 點擊卡片跳轉詳情
const handleCardClick = (id: number) => {
  void router.push(`/serviced-residences/${id}`);
};

// 11. 點擊分頁
const handlePageSelect = () => {
  // mock：不實作實際分頁邏輯
};
</script>

<template>
  <div class="sv-page">
    <!-- 1. 暗色 HERO 區 -->
    <section class="sv-hero pat">
      <div class="sv-hero-text">
        <div class="hero-eyebrow">SERVICE APARTMENTS</div>
        <h2 class="sv-hero-title">
          靈活短租<br>全城精選
        </h2>
        <p class="sv-hero-desc">
          酒店式管理，家的感覺。按日、按週、按月靈活租用，適合商務出行及過渡期居住。
        </p>
      </div>
    </section>

    <!-- 2. 篩選欄 -->
    <section class="sv-controls">
      <div class="sv-search-row">
        <div class="sv-search-box">
          <span class="sv-search-ico">
            <svg
              width="16"
              height="16"
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
          </span>
          <input
            v-model="keyword"
            class="sv-search-input"
            placeholder="搜尋地區、屋苑或關鍵字..."
            autocomplete="off"
          />
        </div>
        <button
          type="button"
          class="sv-search-btn"
        >搜尋</button>
      </div>

      <div class="sv-filter-row">
        <div class="sv-filter-group">
          <span class="sv-filter-label">地區</span>
          <div class="sv-pills">
            <span
              v-for="(item, idx) in regionFilters"
              :key="`region-${idx}`"
              class="sv-pill"
              :class="item.active ? 'on' : ''"
              @click="handleRegionToggle(item)"
            >{{ item.label }}</span>
          </div>
        </div>
        <div class="sv-filter-group">
          <span class="sv-filter-label">租期</span>
          <div class="sv-pills">
            <span
              v-for="(item, idx) in leaseFilters"
              :key="`lease-${idx}`"
              class="sv-pill"
              :class="item.active ? 'on' : ''"
              @click="toggleLeaseFilter(item)"
            >{{ item.label }}</span>
          </div>
        </div>
      </div>
    </section>

    <!-- 3. 卡片網格 -->
    <section class="sv-content">
      <div class="sv-content-header">
        <div class="sv-sec-title">精選服務式住宅</div>
        <div class="sv-toolbar">
          <span class="sv-count">共 18 間</span>
          <select
            v-model="sortBy"
            class="sv-sort"
          >
            <option value="latest">最新發布</option>
            <option value="price_asc">價格低至高</option>
            <option value="price_desc">價格高至低</option>
            <option value="area_desc">面積大至小</option>
          </select>
        </div>
      </div>

      <div class="sv-grid">
        <article
          v-for="card in cards"
          :key="card.id"
          class="sv-card"
          @click="handleCardClick(card.id)"
        >
          <div
            class="sv-img pat"
            :style="{ background: card.imageBg }"
          ></div>
          <div class="sv-body">
            <div class="gtags">
              <span
                v-for="(tag, idx) in card.tags"
                :key="idx"
                class="gtag"
                :class="tag.dark ? 'dark' : ''"
              >{{ tag.label }}</span>
            </div>
            <div class="gtitle">{{ card.title }}</div>
            <div class="gsub">{{ card.sub }}</div>
            <div class="gprice">
              {{ card.price }} <span>{{ card.priceUnit }}</span>
            </div>
            <div class="gpills">
              <span
                v-for="(pill, idx) in card.pills"
                :key="idx"
                class="gpill"
              >{{ pill }}</span>
            </div>
          </div>
          <div class="gfoot">
            <span class="grooms">{{ card.rooms }}</span>
            <span class="gview">查看 →</span>
          </div>
        </article>
      </div>

      <!-- 4. 分頁 -->
      <PaginationBar
        info="第 1-6 筆，共 18 筆"
        :pages="paginationPages"
        aria-label="服務式住宅分頁"
        @select="handlePageSelect"
      />
    </section>
  </div>
</template>

<style scoped>
/* 1. 頁面容器 */
.sv-page {
  width: 100%;
  min-height: calc(100vh - var(--nav-h, 52px));
  background: rgb(var(--color-surface-2));
}

/* 2. 暗色 HERO 區 */
.sv-hero {
  min-height: 200px;
  display: flex;
  align-items: flex-end;
  padding: 40px 40px 32px;
  position: relative;
  overflow: hidden;
  background: linear-gradient(160deg, #1a1a1a, #2e2e2e);
}

.sv-hero-text {
  position: relative;
  z-index: 1;
  max-width: 1200px;
  margin: 0 auto;
  width: 100%;
}

.hero-eyebrow {
  font-size: 10px;
  letter-spacing: 3px;
  color: rgb(var(--color-ink-4));
  margin-bottom: 14px;
}

.sv-hero-title {
  font-family: var(--font-display);
  font-size: 36px;
  color: #fff;
  font-weight: 400;
  line-height: 1.2;
  margin: 0 0 12px;
}

.sv-hero-desc {
  font-size: 13px;
  color: rgb(var(--color-ink-4));
  line-height: 1.8;
  max-width: 400px;
  margin: 0;
}

/* 3. 斜紋圖案覆蓋 */
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

/* 4. 篩選欄 */
.sv-controls {
  background: rgb(var(--color-surface));
  border-bottom: 1px solid rgb(var(--color-border));
  padding: 18px 40px;
}

.sv-search-row {
  display: flex;
  align-items: stretch;
  gap: 8px;
  margin-bottom: 14px;
  max-width: 1200px;
  margin-left: auto;
  margin-right: auto;
}

.sv-search-box {
  display: flex;
  flex: 1;
  max-width: 520px;
  align-items: center;
  gap: 8px;
  border: 1px solid rgb(var(--color-border));
  border-radius: 3px;
  background: rgb(var(--color-surface));
  padding: 0 13px;
}

.sv-search-ico {
  color: rgb(var(--color-ink-3));
  display: flex;
  align-items: center;
}

.sv-search-input {
  flex: 1;
  border: 0;
  outline: none;
  box-shadow: none;
  padding: 12px 0;
  font-size: 13px;
  font-family: inherit;
  background: transparent;
  color: rgb(var(--color-text));
}

.sv-search-input::placeholder {
  color: rgb(var(--color-ink-4));
}

.sv-search-btn {
  border: 1px solid rgb(var(--color-primary));
  border-radius: 3px;
  background: rgb(var(--color-primary));
  color: #fff;
  cursor: pointer;
  font-family: inherit;
  font-size: 12px;
  font-weight: 600;
  padding: 0 18px;
  transition: background 0.15s ease;
}

.sv-search-btn:hover {
  background: rgb(var(--color-brand-dark));
  border-color: rgb(var(--color-brand-dark));
}

.sv-filter-row {
  display: flex;
  align-items: flex-start;
  gap: 32px;
  flex-wrap: wrap;
  max-width: 1200px;
  margin: 0 auto;
}

.sv-filter-group {
  display: flex;
  align-items: center;
  gap: 10px;
}

.sv-filter-label {
  font-size: 11px;
  color: rgb(var(--color-ink-3));
  letter-spacing: 0.5px;
  white-space: nowrap;
}

.sv-pills {
  display: flex;
  flex-wrap: wrap;
  gap: 7px;
}

.sv-pill {
  padding: 5px 13px;
  border: 1px solid rgb(var(--color-border));
  border-radius: 999px;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-ink-3));
  font-size: 12px;
  cursor: pointer;
  font-family: inherit;
  transition: all 0.15s ease;
}

.sv-pill:hover {
  border-color: rgb(var(--color-brand-mid));
  color: rgb(var(--color-primary));
}

.sv-pill.on {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary));
  color: #fff;
}

/* 5. 卡片網格區 */
.sv-content {
  padding: 26px 40px 44px;
  max-width: 1200px;
  margin: 0 auto;
}

.sv-content-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 20px;
  flex-wrap: wrap;
}

.sv-sec-title {
  font-size: 13px;
  font-weight: 500;
  letter-spacing: 0.5px;
  color: rgb(var(--color-text));
  border-left: 3px solid rgb(var(--color-primary));
  padding-left: 10px;
}

.sv-toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
}

.sv-count {
  color: rgb(var(--color-ink-3));
  font-size: 12px;
}

.sv-sort {
  border: 1px solid rgb(var(--color-border));
  border-radius: 2px;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text));
  font-family: inherit;
  font-size: 12px;
  padding: 7px 26px 7px 10px;
  outline: none;
  cursor: pointer;
}

.sv-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 14px;
}

/* 6. 卡片樣式 */
.sv-card {
  border: 1px solid rgb(var(--color-border));
  border-radius: 3px;
  overflow: hidden;
  background: rgb(var(--color-surface));
  cursor: pointer;
  transition: border-color 0.15s, box-shadow 0.15s;
}

.sv-card:hover {
  border-color: rgb(var(--color-primary));
  box-shadow: var(--shadow-raised);
}

.sv-img {
  width: 100%;
  height: 140px;
}

.sv-body {
  padding: 12px 14px;
}

.gtags {
  display: flex;
  gap: 3px;
  margin-bottom: 6px;
  flex-wrap: wrap;
}

.gtag {
  font-size: 9px;
  letter-spacing: 0.8px;
  color: rgb(var(--color-ink-3));
  border: 1px solid rgb(var(--color-border));
  padding: 1px 5px;
  border-radius: 1px;
}

.gtag.dark {
  background: rgb(var(--color-primary));
  color: #fff;
  border-color: rgb(var(--color-primary));
  font-weight: 500;
}

.gtitle {
  font-size: 13px;
  font-weight: 500;
  margin-bottom: 3px;
  color: rgb(var(--color-text));
  line-height: 1.35;
}

.gsub {
  font-size: 11px;
  color: rgb(var(--color-ink-3));
  margin-bottom: 7px;
  line-height: 1.55;
}

.gprice {
  font-size: 17px;
  font-weight: 300;
  letter-spacing: -0.3px;
  margin-top: 6px;
  color: rgb(var(--color-text));
}

.gprice span {
  font-size: 11px;
  color: rgb(var(--color-ink-3));
  font-weight: 400;
}

.gpills {
  display: flex;
  gap: 6px;
  margin-top: 10px;
  flex-wrap: wrap;
}

.gpill {
  font-size: 9px;
  padding: 2px 6px;
  background: rgb(var(--color-surface-2));
  color: rgb(var(--color-ink-2));
  border-radius: 2px;
}

.gfoot {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 7px 13px;
  border-top: 1px solid rgb(var(--color-surface-2));
  font-size: 10px;
}

.grooms {
  color: rgb(var(--color-ink-3));
}

.gview {
  color: rgb(var(--color-primary));
  cursor: pointer;
  font-weight: 500;
}

/* 7. 響應式 */
@media (max-width: 1024px) {
  .sv-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .sv-hero,
  .sv-controls,
  .sv-content {
    padding-left: 24px;
    padding-right: 24px;
  }
}

@media (max-width: 767px) {
  .sv-hero {
    min-height: 160px;
    padding: 28px 20px 24px;
  }

  .sv-hero-title {
    font-size: 28px;
  }

  .sv-hero-desc {
    font-size: 12px;
  }

  .sv-controls {
    padding: 16px 20px;
  }

  .sv-search-row {
    flex-direction: column;
    gap: 8px;
  }

  .sv-search-box {
    max-width: 100%;
  }

  .sv-search-btn {
    padding: 10px 18px;
  }

  .sv-filter-row {
    gap: 14px;
  }

  .sv-filter-group {
    flex-direction: column;
    align-items: flex-start;
    gap: 6px;
  }

  .sv-content {
    padding: 20px 20px 36px;
  }

  .sv-content-header {
    flex-direction: column;
    align-items: flex-start;
  }

  .sv-grid {
    grid-template-columns: 1fr;
    gap: 12px;
  }
}
</style>

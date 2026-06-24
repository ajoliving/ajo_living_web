<!--
 * 樓盤租售列表頁。
 * 1. 三欄布局：左側篩選欄 + 中間列表區 + 右側廣告欄。
 * 2. 左側含搜尋框與 6 組篩選標籤（地區、性質、售價、面積、房間、裝修）。
 * 3. 中間含排序欄、4 個樓盤卡片（懸停顯示操作按鈕）與分頁。
 * 4. 右側含 3 個 banner 廣告與 1 個 vertical 廣告。
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
    key: 'region',
    title: '地區',
    options: [
      { label: '全港', value: 'all' },
      { label: '香港島', value: 'hk_island' },
      { label: '九龍', value: 'kowloon' },
      { label: '新界', value: 'nt' },
      { label: '離島', value: 'islands' },
    ],
    activeValue: 'all',
  },
  {
    key: 'transaction',
    title: '性質',
    options: [
      { label: '全部', value: 'all' },
      { label: '出售', value: 'sale' },
      { label: '出租', value: 'rent' },
    ],
    activeValue: 'all',
  },
  {
    key: 'price',
    title: '售價範圍',
    options: [
      { label: '不限', value: 'any' },
      { label: '400萬以下', value: 'under_400w' },
      { label: '400–800萬', value: '400_800w' },
      { label: '800–1200萬', value: '800_1200w' },
      { label: '2000萬+', value: 'over_2000w' },
    ],
    activeValue: 'any',
  },
  {
    key: 'area',
    title: '實用面積',
    options: [
      { label: '不限', value: 'any' },
      { label: '300呎以下', value: 'under_300' },
      { label: '300–500呎', value: '300_500' },
      { label: '500–1000呎', value: '500_1000' },
      { label: '1000呎+', value: 'over_1000' },
    ],
    activeValue: 'any',
  },
  {
    key: 'bedroom',
    title: '房間',
    options: [
      { label: '全部', value: 'all' },
      { label: '開放式', value: 'studio' },
      { label: '1房', value: '1' },
      { label: '2房', value: '2' },
      { label: '3房', value: '3' },
      { label: '4房+', value: '4plus' },
    ],
    activeValue: 'all',
  },
  {
    key: 'renovation',
    title: '裝修',
    options: [
      { label: '全部', value: 'all' },
      { label: '全新', value: 'new' },
      { label: '有裝修', value: 'renovated' },
      { label: '簡潔', value: 'simple' },
      { label: '特色', value: 'special' },
    ],
    activeValue: 'all',
  },
]);

// 5. 頂部快速篩選（多選，不互斥）
const topFilterTypes = ref([
  { label: '住宅', active: false },
  { label: '車位', active: false },
  { label: '工業', active: false },
  { label: '商廈', active: false },
]);
const topFilterPublishers = ref([
  { label: '業主', active: false },
  { label: '代理', active: false },
]);

// 6. 樓盤卡片 mock 資料
interface PropertyCard {
  id: number;
  typeStack: string[];
  imageClass: 'tall' | 'short';
  imageBg: string;
  tags: { label: string; dark?: boolean }[];
  title: string;
  sub: string;
  price: string;
  priceUnit: string;
  area: string;
  pills: string[];
}
const cards = ref<PropertyCard[]>([
  {
    id: 1,
    typeStack: ['住宅', '代理'],
    imageClass: 'tall',
    imageBg: 'linear-gradient(160deg,#e8e8e8,#d0d0d0)',
    tags: [{ label: '九龍' }, { label: '住宅' }],
    title: '佐敦 高級住宅',
    sub: '佐敦站步行3分鐘 · 全新裝修 · 海景',
    price: 'HK$36,000',
    priceUnit: '/ 月',
    area: '實用面積 200呎',
    pills: ['全新', '基本裝修', '廚房'],
  },
  {
    id: 2,
    typeStack: ['住宅', '業主'],
    imageClass: 'short',
    imageBg: 'linear-gradient(160deg,#d8e0e0,#c0cccc)',
    tags: [{ label: '九龍' }, { label: '住宅' }],
    title: '佐敦婚 38 窗 套房',
    sub: '佐敦站步行5分鐘 · 獨立廚房',
    price: 'HK$6,500',
    priceUnit: '/ 月',
    area: '實用面積 100呎',
    pills: [],
  },
  {
    id: 3,
    typeStack: ['商廈', '代理'],
    imageClass: 'tall',
    imageBg: 'linear-gradient(160deg,#e0e0e8,#c8c8d8)',
    tags: [{ label: '香港島' }, { label: '商廈', dark: true }],
    title: '中環甲級寫字樓',
    sub: '中環站步行2分鐘 · 海景 · 全新裝修',
    price: 'HK$120,000',
    priceUnit: '/ 月',
    area: '實用面積 2,400呎',
    pills: ['海景', '全新裝修', '有冷氣'],
  },
  {
    id: 4,
    typeStack: ['住宅', '業主'],
    imageClass: 'short',
    imageBg: 'linear-gradient(160deg,#e4dcd8,#ccc0bc)',
    tags: [{ label: '新界' }],
    title: '沙田第一城 3房',
    sub: '沙田站步行8分鐘 · 會所設施',
    price: 'HK$18,500',
    priceUnit: '/ 月',
    area: '實用面積 650呎',
    pills: [],
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

// 8. 切換篩選標籤（同組互斥）
const handleFilterToggle = (groupKey: string, optionValue: string) => {
  const group = filterGroups.value.find((g) => g.key === groupKey);
  if (group) {
    group.activeValue = optionValue;
  }
};

// 9. 切換頂部快速篩選（多選）
const toggleTopFilter = (item: { active: boolean }) => {
  item.active = !item.active;
};

// 10. 點擊卡片跳轉詳情
const handleCardClick = (id: number) => {
  void router.push(`/properties/${id}`);
};

// 11. 點擊分頁
const handlePageSelect = () => {
  // mock：不實作實際分頁邏輯
};
</script>

<template>
  <div class="page">
    <div class="lp">
      <!-- 左側篩選欄 -->
      <aside class="lf">
        <div class="sbar">
          <input
            v-model="keyword"
            class="sinput"
            placeholder="搜尋樓盤…"
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
      <main class="lr">
        <div class="sort-row listing-sort-row">
          <div class="listing-result-tools">
            <div class="listing-top-filters">
              <span
                v-for="(item, idx) in topFilterTypes"
                :key="`type-${idx}`"
                class="ft"
                :class="item.active ? 'on' : ''"
                @click="toggleTopFilter(item)"
              >{{ item.label }}</span>
              <span class="listing-filter-divider">｜</span>
              <span
                v-for="(item, idx) in topFilterPublishers"
                :key="`pub-${idx}`"
                class="ft"
                :class="item.active ? 'on' : ''"
                @click="toggleTopFilter(item)"
              >{{ item.label }}</span>
            </div>
            <span class="rn">4 個結果</span>
          </div>
          <select
            v-model="sortBy"
            class="ssel"
          >
            <option value="latest">最新發現</option>
            <option value="price_asc">價格低至高</option>
            <option value="area_desc">面積大至小</option>
          </select>
        </div>

        <div class="list-view">
          <div class="grid">
            <div
              v-for="card in cards"
              :key="card.id"
              class="gc"
              @click="handleCardClick(card.id)"
            >
              <div class="listing-card-type-stack">
                <span
                  v-for="(label, idx) in card.typeStack"
                  :key="idx"
                >{{ label }}</span>
              </div>
              <div
                class="gi pat"
                :class="card.imageClass"
                :style="{ background: card.imageBg }"
              ></div>
              <div class="gb">
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
                <div class="garea">{{ card.area }}</div>
                <div
                  v-if="card.pills.length > 0"
                  class="gpills"
                >
                  <span
                    v-for="(pill, idx) in card.pills"
                    :key="idx"
                    class="gpill"
                  >{{ pill }}</span>
                </div>
              </div>
              <div class="gc-actions">
                <button
                  type="button"
                  class="gc-action-btn"
                  @click.stop
                >收藏</button>
                <button
                  type="button"
                  class="gc-action-btn"
                  @click.stop
                >比較</button>
                <button
                  type="button"
                  class="gc-action-btn primary"
                  @click.stop="handleCardClick(card.id)"
                >查看</button>
              </div>
            </div>
          </div>
        </div>

        <PaginationBar
          info="第 1-4 筆，共 24 筆"
          :pages="paginationPages"
          aria-label="樓盤分頁"
          @select="handlePageSelect"
        />
      </main>

      <!-- 右側廣告欄 -->
      <aside class="listing-ad-aside">
        <div class="ad-side-panel">
          <AdCard
            variant="banner"
            visual="office"
            label="16:9"
            title="新盤代理服務"
            desc="樓盤發布、預約睇樓與租售查詢。"
          />
          <AdCard
            variant="banner"
            visual="home"
            label="16:9"
            title="住宅按揭諮詢"
            desc="按揭預批、估價與成交支援。"
          />
          <AdCard
            variant="banner"
            visual="service"
            label="16:9"
            title="搬遷及驗樓服務"
            desc="入住前檢查、清潔及交收安排。"
          />
          <div class="ad-vertical-grid">
            <AdCard
              variant="vertical"
              visual="market"
              label="9:16"
              title="室內設計"
              desc="住宅與商用空間方案。"
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
.lp {
  display: grid;
  grid-template-columns: 240px minmax(0, 760px) 340px;
  max-width: 1440px;
  margin: 0 auto;
  background: rgb(var(--color-surface-2));
  min-height: calc(100vh - var(--nav-h, 52px));
}

/* 3. 左側篩選欄 */
.lf {
  background: rgb(var(--color-surface));
  border-right: 1px solid rgb(var(--color-border));
  padding: 18px 18px;
  position: sticky;
  top: var(--nav-h, 52px);
  height: calc(100vh - var(--nav-h, 52px));
  overflow-y: auto;
  min-height: calc(100vh - var(--nav-h, 52px));
}

.sbar {
  display: flex;
  gap: 5px;
  margin-bottom: 14px;
}

.sinput {
  flex: 1;
  border: 1px solid rgb(var(--color-border));
  padding: 7px 9px;
  font-size: 11px;
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
  padding: 7px 12px;
  font-size: 11px;
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
}

/* 4. 中間列表區 */
.lr {
  background: rgb(var(--color-surface-2));
  padding: 14px 14px 36px;
  min-width: 0;
  min-height: calc(100vh - var(--nav-h, 52px));
}

.sort-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.listing-result-tools {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
  flex-wrap: wrap;
}

.listing-top-filters {
  display: flex;
  align-items: center;
  gap: 4px;
  flex-wrap: wrap;
}

.listing-top-filters .ft {
  font-size: 11px;
  padding: 4px 9px;
  border: 1px solid rgb(var(--color-border));
  color: rgb(var(--color-ink-3));
  cursor: pointer;
  background: rgb(var(--color-surface));
  border-radius: 2px;
  font-family: inherit;
}

.listing-top-filters .ft:hover {
  border-color: rgb(var(--color-primary));
  color: rgb(var(--color-brand-dark));
}

.listing-top-filters .ft.on {
  background: rgb(var(--color-primary));
  color: #fff;
  border-color: rgb(var(--color-primary));
}

.listing-filter-divider {
  color: rgb(var(--color-ink-4));
  font-size: 12px;
  margin: 0 5px;
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

/* 5. 列表卡片 */
.list-view > .grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 12px;
}

.list-view .gc {
  position: relative;
  display: grid;
  grid-template-columns: minmax(260px, 42%) minmax(0, 1fr);
  align-items: stretch;
  margin-bottom: 0;
  border-radius: 8px;
  border: 1px solid rgb(var(--color-border));
  overflow: hidden;
  cursor: pointer;
  background: rgb(var(--color-surface));
  transition: box-shadow 0.15s, border-color 0.15s;
}

.list-view .gc:hover {
  box-shadow: var(--shadow-raised);
  border-color: rgb(var(--color-brand-mid));
}

.list-view .gc .gi {
  grid-column: 1;
  grid-row: 1 / span 2;
  width: 100%;
  height: auto !important;
  min-height: 0;
  aspect-ratio: 600 / 450;
}

.list-view .gc .gb {
  grid-column: 2;
  grid-row: 1;
  display: flex;
  flex-direction: column;
  justify-content: center;
  min-width: 0;
  padding: 16px 18px 12px;
}

.list-view .gc .gtitle {
  font-size: 15px;
  font-weight: 600;
  line-height: 1.35;
}

.list-view .gc .gsub {
  font-size: 12px;
  line-height: 1.55;
}

.list-view .gc .gprice {
  margin-top: 6px;
  font-size: 20px;
}

.list-view .gc .gpills {
  margin-top: 9px;
}

.list-view .gc .gc-actions {
  grid-column: 2;
  grid-row: 2;
}

.gi {
  width: 100%;
  display: block;
}

.gi.tall {
  height: 150px;
}

.gi.short {
  height: 90px;
}

.gb {
  padding: 11px 13px;
}

.gtags {
  display: flex;
  gap: 5px;
  margin-bottom: 10px;
}

.gtag {
  font-size: 11px;
  padding: 3px 7px;
  letter-spacing: 0.8px;
  color: rgb(var(--color-ink-3));
  border: 1px solid rgb(var(--color-border));
  border-radius: 1px;
}

.gtag.dark {
  background: rgb(var(--color-primary));
  color: #fff;
  border-color: rgb(var(--color-primary));
  font-weight: 500;
}

.gtitle {
  font-size: 15px;
  font-weight: 600;
  line-height: 1.35;
  margin-bottom: 3px;
  color: rgb(var(--color-text));
}

.gsub {
  font-size: 12px;
  color: rgb(var(--color-ink-3));
  margin-bottom: 7px;
  line-height: 1.55;
}

.gprice {
  font-size: 20px;
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

.garea {
  margin-top: 6px;
  font-size: 10px;
  color: rgb(var(--color-ink-3));
  line-height: 1.4;
}

.gpills {
  display: flex;
  flex-wrap: wrap;
  gap: 3px;
  margin-top: 9px;
}

.gpill {
  font-size: 9px;
  padding: 2px 6px;
  background: rgb(var(--color-surface-2));
  color: rgb(var(--color-ink-2));
  border-radius: 2px;
}

/* 6. 懸停操作按鈕 */
.gc-actions {
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

.gc:hover .gc-actions {
  max-height: 54px;
  opacity: 1;
  pointer-events: auto;
  transform: translateY(0);
  padding: 0 18px 14px;
}

.gc-action-btn {
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

.gc-action-btn:hover {
  border-color: rgb(var(--color-brand-mid));
  color: rgb(var(--color-primary));
}

.gc-action-btn.primary {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary));
  color: #fff;
}

.gc-action-btn.primary:hover {
  background: rgb(var(--color-brand-dark));
  color: #fff;
}

/* 7. 類型標籤堆疊 */
.listing-card-type-stack {
  position: absolute;
  top: 8px;
  left: 8px;
  display: flex;
  flex-direction: column;
  gap: 3px;
  z-index: 2;
}

.listing-card-type-stack span {
  font-size: 9px;
  padding: 2px 6px;
  background: rgba(255, 255, 255, 0.92);
  color: rgb(var(--color-ink-2));
  border-radius: 2px;
  letter-spacing: 0.3px;
}

/* 8. 斜紋圖案 */
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

/* 9. 右側廣告欄 */
.listing-ad-aside {
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

/* 10. 響應式 */
@media (max-width: 1100px) {
  .lp {
    grid-template-columns: 210px 1fr;
  }

  .listing-ad-aside {
    display: none;
  }
}

@media (max-width: 767px) {
  .lp {
    grid-template-columns: 1fr;
  }

  .lf {
    position: static;
    height: auto;
  }

  .list-view .gc {
    grid-template-columns: 1fr;
  }

  .list-view .gc .gi {
    grid-row: auto;
    aspect-ratio: 600 / 450;
  }

  .list-view .gc .gb {
    grid-column: 1;
    grid-row: auto;
  }

  .list-view .gc .gc-actions {
    grid-column: 1;
    grid-row: auto;
  }
}
</style>

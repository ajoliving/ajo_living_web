<!--
 * 走勢頁 - 香港租金走勢。
 * 1. 頁首標題與區域切換 Tabs。
 * 2. SVG 折線圖呈現過去 6 個月各區平均呎租趨勢。
 * 3. 各區最新呎租卡片網格。
-->
<script setup lang="ts">
import { ref } from 'vue';
import { useRouter } from 'vue-router';

// 1. 區域 Tab 型別
interface TrendTab {
  key: string;
  label: string;
}

// 2. 區域卡片型別
interface DistrictCard {
  code: string;
  name: string;
  price: string;
  change: string;
  direction: 'up' | 'down';
}

// 3. 折線圖資料點型別
interface ChartPoint {
  x: number;
  y: number;
}

// 4. 折線圖系列型別
interface ChartSeries {
  key: string;
  label: string;
  color: string;
  strokeWidth: number;
  dashed: boolean;
  points: ChartPoint[];
  legendX: number;
}

const router = useRouter();

// 5. 區域 Tab 列表
const tabs: TrendTab[] = [
  { key: 'all', label: '全港' },
  { key: 'hk', label: '香港島' },
  { key: 'kln', label: '九龍' },
  { key: 'nt', label: '新界' },
];

const activeTab = ref('all');

// 6. 折線圖 Y 軸標籤
const chartYLabels = [
  { y: 20, text: '$60' },
  { y: 60, text: '$50' },
  { y: 100, text: '$40' },
  { y: 140, text: '$30' },
];

// 7. 折線圖網格線 Y 座標
const chartGridLines = [20, 60, 100, 140];

// 8. 折線圖 X 軸標籤
const chartXLabels = [
  { x: 108, text: '1月' },
  { x: 192, text: '2月' },
  { x: 276, text: '3月' },
  { x: 360, text: '4月' },
  { x: 444, text: '5月' },
  { x: 528, text: '6月' },
];

// 9. 折線圖系列資料
const chartSeries: ChartSeries[] = [
  {
    key: 'hk',
    label: '香港島',
    color: '#f05a00',
    strokeWidth: 2.5,
    dashed: false,
    points: [
      { x: 108, y: 45 },
      { x: 192, y: 40 },
      { x: 276, y: 50 },
      { x: 360, y: 35 },
      { x: 444, y: 38 },
      { x: 528, y: 30 },
    ],
    legendX: 300,
  },
  {
    key: 'kln',
    label: '九龍',
    color: '#1a1a1a',
    strokeWidth: 2.5,
    dashed: false,
    points: [
      { x: 108, y: 75 },
      { x: 192, y: 72 },
      { x: 276, y: 80 },
      { x: 360, y: 68 },
      { x: 444, y: 65 },
      { x: 528, y: 60 },
    ],
    legendX: 360,
  },
  {
    key: 'nt',
    label: '新界',
    color: '#aaa',
    strokeWidth: 2,
    dashed: true,
    points: [
      { x: 108, y: 115 },
      { x: 192, y: 112 },
      { x: 276, y: 118 },
      { x: 360, y: 108 },
      { x: 444, y: 110 },
      { x: 528, y: 105 },
    ],
    legendX: 410,
  },
];

// 10. 各區最新呎租卡片
const districtCards: DistrictCard[] = [
  { code: 'central-admiralty', name: '中環 / 金鐘', price: 'HK$62/呎', change: '+3.2% 本月', direction: 'up' },
  { code: 'causeway-bay', name: '銅鑼灣', price: 'HK$55/呎', change: '+1.8% 本月', direction: 'up' },
  { code: 'tsim-sha-tsui', name: '尖沙咀', price: 'HK$48/呎', change: '+0.9% 本月', direction: 'up' },
  { code: 'mong-kok', name: '旺角', price: 'HK$42/呎', change: '-0.5% 本月', direction: 'down' },
  { code: 'sha-tin', name: '沙田', price: 'HK$32/呎', change: '+1.1% 本月', direction: 'up' },
  { code: 'tseung-kwan-o', name: '將軍澳', price: 'HK$28/呎', change: '-0.3% 本月', direction: 'down' },
];

// 11. 切換區域 Tab
const selectTab = (key: string): void => {
  activeTab.value = key;
};

// 12. 將折線圖系列點轉為 polyline points 字串
const toPolylinePoints = (points: ChartPoint[]): string => {
  return points.map((p) => `${p.x},${p.y}`).join(' ');
};

// 13. 開啟各區詳情
const openDistrict = (card: DistrictCard): void => {
  void router.push({ path: `/trend/district/${encodeURIComponent(card.code)}` });
};
</script>

<template>
  <main class="trend-page">
    <div class="trend-page-wrap">
      <!-- 1. 頁首 -->
      <header class="trend-header">
        <div class="trend-eyebrow">市場數據</div>
        <h1 class="trend-title">香港租金走勢</h1>
        <p class="trend-sub">過去6個月各區平均月租變化（每呎）</p>
      </header>

      <!-- 2. 區域 Tabs -->
      <nav class="trend-tabs">
        <button
          v-for="tab in tabs"
          :key="tab.key"
          type="button"
          class="trend-tab"
          :class="activeTab === tab.key ? 'on' : ''"
          @click="selectTab(tab.key)"
        >{{ tab.label }}</button>
      </nav>

      <!-- 3. 折線圖 -->
      <section class="chart-wrap">
        <div class="chart-title">平均呎租趨勢（HK$/呎）</div>
        <svg
          class="svg-chart"
          viewBox="0 0 600 180"
          role="img"
          aria-label="平均呎租趨勢圖"
        >
          <!-- 網格線 -->
          <line
            v-for="y in chartGridLines"
            :key="`grid-${y}`"
            x1="60"
            :y1="y"
            x2="580"
            :y2="y"
            stroke="#e8e8e8"
            stroke-width="1"
          />
          <!-- Y 軸標籤 -->
          <text
            v-for="label in chartYLabels"
            :key="`y-${label.y}`"
            x="50"
            :y="label.y + 4"
            text-anchor="end"
            font-size="10"
            fill="#aaa"
          >{{ label.text }}</text>
          <!-- X 軸標籤 -->
          <text
            v-for="label in chartXLabels"
            :key="`x-${label.x}`"
            :x="label.x"
            y="165"
            text-anchor="middle"
            font-size="10"
            fill="#aaa"
          >{{ label.text }}</text>
          <!-- 折線 -->
          <polyline
            v-for="series in chartSeries"
            :key="`line-${series.key}`"
            :points="toPolylinePoints(series.points)"
            fill="none"
            :stroke="series.color"
            :stroke-width="series.strokeWidth"
            stroke-linejoin="round"
            :stroke-dasharray="series.dashed ? '4,3' : 'none'"
          />
          <!-- 終點圓點 -->
          <circle
            v-for="series in chartSeries"
            :key="`dot-${series.key}`"
            :cx="series.points[series.points.length - 1].x"
            :cy="series.points[series.points.length - 1].y"
            r="4"
            :fill="series.color"
          />
          <!-- 圖例 -->
          <g
            v-for="series in chartSeries"
            :key="`legend-${series.key}`"
          >
            <rect
              :x="series.legendX"
              y="8"
              width="10"
              height="3"
              :fill="series.color"
              rx="1"
            />
            <text
              :x="series.legendX + 14"
              y="13"
              font-size="10"
              fill="#555"
            >{{ series.label }}</text>
          </g>
        </svg>
      </section>

      <!-- 4. 各區最新呎租 -->
      <div class="trend-section-label">各區最新呎租</div>
      <section class="district-grid">
        <article
          v-for="card in districtCards"
          :key="card.code"
          class="district-card"
          @click="openDistrict(card)"
        >
          <div class="dc-name">{{ card.name }}</div>
          <div class="dc-price">{{ card.price }}</div>
          <div
            class="dc-change"
            :class="card.direction === 'up' ? 'dc-up' : 'dc-dn'"
          >
            <span class="dc-arrow">{{ card.direction === 'up' ? '↑' : '↓' }}</span>
            <span>{{ card.change }}</span>
          </div>
        </article>
      </section>
    </div>
  </main>
</template>

<style scoped>
/* 1. 頁面容器 */
.trend-page {
  width: 100%;
  max-width: var(--layout-page-max-width);
  margin: 0 auto;
  background: rgb(var(--color-surface-2));
  color: rgb(var(--color-text));
}

.trend-page-wrap {
  padding: 24px;
}

/* 2. 頁首 */
.trend-header {
  margin-bottom: 20px;
}

.trend-eyebrow {
  font-size: 9px;
  letter-spacing: 2.5px;
  text-transform: uppercase;
  color: rgb(var(--color-primary));
  margin-bottom: 6px;
  font-weight: 500;
}

.trend-title {
  font-family: var(--font-display);
  font-size: 20px;
  font-weight: 500;
  color: rgb(var(--color-text));
  margin: 0 0 4px;
}

.trend-sub {
  font-size: 12px;
  color: rgb(var(--color-ink-4));
  margin: 0;
}

/* 3. 區域 Tabs */
.trend-tabs {
  display: flex;
  gap: 6px;
  margin-bottom: 20px;
  flex-wrap: wrap;
}

.trend-tab {
  font-family: inherit;
  font-size: 11px;
  padding: 5px 14px;
  border: 1px solid rgb(var(--color-border));
  border-radius: 20px;
  cursor: pointer;
  color: rgb(var(--color-ink-4));
  background: rgb(var(--color-surface));
  transition: all 0.12s;
}

.trend-tab.on {
  background: rgb(var(--color-text));
  color: rgb(var(--color-surface));
  border-color: rgb(var(--color-text));
}

/* 4. 折線圖 */
.chart-wrap {
  background: rgb(var(--color-surface));
  border: 1px solid rgb(var(--color-border));
  border-radius: 4px;
  padding: 20px;
  margin-bottom: 16px;
}

.chart-title {
  font-size: 12px;
  font-weight: 500;
  margin-bottom: 14px;
  color: rgb(var(--color-text));
}

.svg-chart {
  width: 100%;
  height: auto;
  overflow: visible;
}

/* 5. 區段標籤 */
.trend-section-label {
  font-size: 9px;
  letter-spacing: 1.5px;
  text-transform: uppercase;
  color: rgb(var(--color-ink-4));
  font-weight: 500;
  margin-bottom: 12px;
}

/* 6. 各區卡片網格 */
.district-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
  margin-top: 16px;
}

.district-card {
  border: 1px solid rgb(var(--color-border));
  border-radius: 3px;
  padding: 14px;
  background: rgb(var(--color-surface));
  cursor: pointer;
  transition: box-shadow 0.15s, border-color 0.15s;
}

.district-card:hover {
  box-shadow: var(--shadow-raised);
  border-color: rgb(var(--color-brand-mid));
}

.dc-name {
  font-size: 12px;
  font-weight: 500;
  margin-bottom: 4px;
  color: rgb(var(--color-text));
}

.dc-price {
  font-size: 18px;
  font-weight: 300;
  letter-spacing: -0.5px;
  color: rgb(var(--color-text));
}

.dc-change {
  font-size: 11px;
  margin-top: 3px;
  display: flex;
  align-items: center;
  gap: 2px;
}

.dc-up {
  color: rgb(var(--color-success));
}

.dc-dn {
  color: rgb(var(--color-danger));
}

.dc-arrow {
  display: inline-block;
}

/* 7. 響應式設計 */
@media (max-width: 768px) {
  .trend-page-wrap {
    padding: 16px;
  }

  .district-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 480px) {
  .district-grid {
    grid-template-columns: 1fr;
  }
}
</style>

<!--
 * 走勢頁 - 香港住宅租金走勢。
 * 1. 讀取 AJO 後端提供的香港公開租金走勢資料。
 * 2. 以 SVG 折線圖呈現近 6 個月香港島、九龍與新界平均呎租。
 * 3. 提供區域切換、最新值與資料來源狀態。
-->
<script setup lang="ts">
import axios from 'axios';
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';

import { fetchMarketRentTrend } from '@/httpapis/market-trends';
import type { MarketRentTrendRegion, MarketRentTrendResponse } from '@/model/market-trend';
import { usePreferenceStore } from '@/stores/preferences';

// 1. 區域 Tab 型別
interface TrendTab {
  key: string;
  label: string;
}

// 2. 折線圖資料點型別
interface ChartPoint {
  x: number;
  y: number;
}

// 3. 折線圖系列型別
interface ChartSeries {
  key: string;
  label: string;
  color: string;
  strokeWidth: number;
  dashed: boolean;
  points: ChartPoint[];
  values: string[];
}

const activeTab = ref('all');
const { t } = useI18n();
const preferenceStore = usePreferenceStore();
const loading = ref(false);
const errorMessage = ref('');
const rentTrend = ref<MarketRentTrendResponse | null>(null);

const chartWidth = 600;
const chartHeight = 180;
const chartLeft = 64;
const chartRight = 580;
const chartTop = 24;
const chartBottom = 142;
const seriesColors: Record<string, string> = {
  hk: '#b45309',
  kln: '#111827',
  nt: '#64748b',
};

// 4. 區域 Tab 列表
const tabs = computed<TrendTab[]>(() => [
  { key: 'all', label: t('trend.allHongKong') },
  ...trendRegions.value.map((region) => ({
    key: region.key,
    label: resolveRegionLabel(region),
  })),
]);

const trendRegions = computed<MarketRentTrendRegion[]>(() => rentTrend.value?.regions ?? []);
const visibleRegions = computed<MarketRentTrendRegion[]>(() => {
  if (activeTab.value === 'all') {
    return trendRegions.value;
  }

  return trendRegions.value.filter((region) => region.key === activeTab.value);
});

const chartValues = computed<number[]>(() =>
  visibleRegions.value.flatMap((region) =>
    region.points.map((point) => point.value_hkd_per_sqft),
  ),
);

const chartRange = computed(() => {
  if (chartValues.value.length === 0) {
    return { min: 0, max: 1 };
  }

  const minValue = Math.min(...chartValues.value);
  const maxValue = Math.max(...chartValues.value);
  if (minValue === maxValue) {
    return { min: minValue - 1, max: maxValue + 1 };
  }

  const padding = Math.max((maxValue - minValue) * 0.16, 1);
  return {
    min: Math.max(0, Math.floor(minValue - padding)),
    max: Math.ceil(maxValue + padding),
  };
});

const chartGridLines = computed(() => {
  const lines = 4;
  const step = (chartRange.value.max - chartRange.value.min) / (lines - 1);

  return Array.from({ length: lines }, (_, index) => {
    const value = chartRange.value.max - step * index;
    return {
      y: chartTop + ((chartRange.value.max - value) / (chartRange.value.max - chartRange.value.min)) * (chartBottom - chartTop),
      text: formatRentValue(value, 0),
    };
  });
});

const chartXLabels = computed(() => {
  const points = visibleRegions.value[0]?.points ?? [];
  return points.map((point, index) => ({
    x: pointX(index, points.length),
    text: formatTrendMonth(point.month, 'short'),
  }));
});

const chartSeries = computed<ChartSeries[]>(() =>
  visibleRegions.value.map((region, index) => ({
    key: region.key,
    label: resolveRegionLabel(region),
    color: seriesColors[region.key] ?? '#111827',
    strokeWidth: index === 0 ? 2.6 : 2.2,
    dashed: region.key === 'nt',
    points: region.points.map((point, pointIndex) => ({
      x: pointX(pointIndex, region.points.length),
      y: pointY(point.value_hkd_per_sqft),
    })),
    values: region.points.map((point) => formatRentValue(point.value_hkd_per_sqft)),
  })),
);

const updatedMonthLabel = computed(() => {
  if (!rentTrend.value?.updated_month) {
    return '';
  }

  return formatTrendMonth(rentTrend.value.updated_month, 'long');
});

// 5. 取得區域顯示名稱
const resolveRegionLabel = (region: MarketRentTrendRegion): string => {
  const key = `trend.regions.${region.key}`;
  const translated = t(key);
  return translated === key ? region.label : translated;
};

// 6. 格式化趨勢月份
const formatTrendMonth = (value: string, month: 'short' | 'long'): string => {
  const date = new Date(`${value}-01T00:00:00`);
  if (Number.isNaN(date.getTime())) {
    return value;
  }

  return new Intl.DateTimeFormat(preferenceStore.locale, {
    year: month === 'long' ? 'numeric' : undefined,
    month,
  }).format(date);
};

// 7. 格式化每平方呎租金
const formatRentValue = (value: number, maximumFractionDigits = 1): string =>
  new Intl.NumberFormat(preferenceStore.locale, {
    style: 'currency',
    currency: 'HKD',
    minimumFractionDigits: maximumFractionDigits,
    maximumFractionDigits,
  }).format(value);

// 8. 格式化區域最新租金
const formatRegionRent = (region: MarketRentTrendRegion): string =>
  t('trend.pricePerSqft', { value: formatRentValue(region.latest_hkd_per_sqft) });

// 9. 讀取租金走勢
const loadTrend = async (): Promise<void> => {
  loading.value = true;
  errorMessage.value = '';

  try {
    const { data } = await fetchMarketRentTrend();
    rentTrend.value = data.data;
    if (!data.data.regions.some((region) => region.key === activeTab.value)) {
      activeTab.value = 'all';
    }
  } catch (error: unknown) {
    errorMessage.value = resolveTrendError(error);
  } finally {
    loading.value = false;
  }
};

// 10. 切換區域 Tab
const selectTab = (key: string): void => {
  activeTab.value = key;
};

// 11. 將折線圖系列點轉為 polyline points 字串
const toPolylinePoints = (points: ChartPoint[]): string =>
  points.map((point) => `${point.x},${point.y}`).join(' ');

// 12. 計算 X 軸位置
const pointX = (index: number, total: number): number => {
  if (total <= 1) {
    return chartLeft;
  }

  return chartLeft + ((chartRight - chartLeft) / (total - 1)) * index;
};

// 13. 計算 Y 軸位置
const pointY = (value: number): number => {
  const range = chartRange.value.max - chartRange.value.min;
  if (range <= 0) {
    return chartBottom;
  }

  return chartBottom - ((value - chartRange.value.min) / range) * (chartBottom - chartTop);
};

// 14. 格式化月變化
const formatChange = (value: number): string =>
  new Intl.NumberFormat(preferenceStore.locale, {
    style: 'percent',
    signDisplay: 'exceptZero',
    minimumFractionDigits: 1,
    maximumFractionDigits: 1,
  }).format(value / 100);

// 15. 解析錯誤訊息
const resolveTrendError = (error: unknown): string =>
  axios.isAxiosError<{ message?: string }>(error)
    ? error.response?.data?.message ?? t('trend.loadError')
    : t('trend.loadError');

onMounted(() => {
  void loadTrend();
});
</script>

<template>
  <div
    id="page-trend"
    class="page"
  >
    <div class="trend-page-wrap">
      <!-- 1. 頁首 -->
      <div class="trend-header">
        <div class="section-eyebrow">{{ t('trend.eyebrow') }}</div>
        <div class="trend-title">{{ t('trend.title') }}</div>
        <div class="trend-sub">
          {{ t('trend.subtitle') }}
        </div>
      </div>

      <!-- 2. 區域 Tabs -->
      <div class="trend-tabs">
        <button
          v-for="tab in tabs"
          :key="tab.key"
          type="button"
          class="trend-tab"
          :class="{ on: activeTab === tab.key }"
          @click="selectTab(tab.key)"
        >
          {{ tab.label }}
        </button>
      </div>

      <!-- 3. 狀態 -->
      <div
        v-if="loading"
        class="trend-state"
      >
        {{ t('trend.loading') }}
      </div>
      <div
        v-else-if="errorMessage"
        class="trend-state trend-state--error"
      >
        {{ errorMessage }}
        <button
          type="button"
          @click="loadTrend"
        >
          {{ t('trend.retry') }}
        </button>
      </div>

      <template v-else-if="rentTrend">
        <!-- 4. 折線圖 -->
        <div class="chart-wrap">
          <div class="chart-head">
            <div>
              <div class="chart-title">{{ t('trend.chartTitle') }}</div>
              <div class="chart-meta">{{ t('trend.updatedThrough', { month: updatedMonthLabel }) }}</div>
            </div>
            <a
              class="chart-source"
              :href="rentTrend.source_url"
              target="_blank"
              rel="noreferrer"
            >
              DATA.GOV.HK
            </a>
          </div>

          <svg
            class="svg-chart"
            :viewBox="`0 0 ${chartWidth} ${chartHeight}`"
            role="img"
            :aria-label="t('trend.chartAria')"
          >
            <line
              v-for="line in chartGridLines"
              :key="`grid-${line.text}`"
              x1="64"
              :y1="line.y"
              x2="580"
              :y2="line.y"
              stroke="#e8e8e8"
              stroke-width="1"
            />
            <text
              v-for="line in chartGridLines"
              :key="`y-${line.text}`"
              x="54"
              :y="line.y + 4"
              text-anchor="end"
              font-size="10"
              fill="#777"
            >
              {{ line.text }}
            </text>
            <text
              v-for="label in chartXLabels"
              :key="`x-${label.text}`"
              :x="label.x"
              y="165"
              text-anchor="middle"
              font-size="10"
              fill="#777"
            >
              {{ label.text }}
            </text>
            <polyline
              v-for="series in chartSeries"
              :key="`line-${series.key}`"
              :points="toPolylinePoints(series.points)"
              fill="none"
              :stroke="series.color"
              :stroke-width="series.strokeWidth"
              stroke-linecap="round"
              stroke-linejoin="round"
              :stroke-dasharray="series.dashed ? '5,4' : 'none'"
            />
            <g
              v-for="series in chartSeries"
              :key="`dots-${series.key}`"
            >
              <circle
                v-for="(point, index) in series.points"
                :key="`${series.key}-${index}`"
                :cx="point.x"
                :cy="point.y"
                r="3.6"
                :fill="series.color"
              >
                <title>{{ `${series.label} ${series.values[index]}` }}</title>
              </circle>
            </g>
          </svg>

          <div class="chart-legend">
            <div
              v-for="series in chartSeries"
              :key="`legend-${series.key}`"
              class="legend-item"
            >
              <span
                class="legend-line"
                :style="{ background: series.color }"
              />
              {{ series.label }}
            </div>
          </div>
        </div>

        <!-- 5. 各區最新呎租 -->
        <div
          class="label-text"
          style="margin-bottom: 12px;"
        >
          {{ t('trend.latestByRegion') }}
        </div>
        <div class="district-grid">
          <button
            v-for="region in trendRegions"
            :key="region.key"
            type="button"
            class="district-card"
            :class="{ on: activeTab === region.key }"
            @click="selectTab(region.key)"
          >
            <div class="dc-name">{{ resolveRegionLabel(region) }}</div>
            <div class="dc-price">{{ formatRegionRent(region) }}</div>
            <div
              class="dc-change"
              :class="{
                'dc-up': region.direction === 'up',
                'dc-dn': region.direction === 'down',
              }"
            >
              {{ formatChange(region.monthly_change_percent) }} {{ t('trend.monthly') }}
            </div>
          </button>
        </div>

        <!-- 6. 資料說明 -->
        <div class="source-note">
          <span>{{ rentTrend.dataset }}</span>
          <span>{{ t('trend.method') }}</span>
        </div>
      </template>
    </div>
  </div>
</template>

<style scoped>
/*
 * 走勢頁樣式。
 * 1. 頁面容器與頁首。
 * 2. 區域 Tabs 與狀態。
 * 3. 折線圖卡片。
 * 4. 各區呎租卡片網格。
 */

/* 1. 頁面容器 */
.page {
  width: 100%;
  background: var(--sur);
}

.trend-page-wrap {
  padding: 24px;
  background: var(--sur);
}

/* 2. 頁首 */
.trend-header {
  margin-bottom: 20px;
}

.section-eyebrow {
  margin-bottom: 6px;
  color: var(--accent);
  font-size: 9px;
  font-weight: 500;
  letter-spacing: 2.5px;
  text-transform: uppercase;
}

.trend-title {
  margin-bottom: 4px;
  color: var(--ink);
  font-family: var(--font-serif);
  font-size: 20px;
  font-weight: 400;
}

.trend-sub {
  color: var(--g4);
  font-size: 12px;
}

/* 3. 區域 Tabs */
.trend-tabs {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-bottom: 20px;
}

.trend-tab {
  border: 1px solid var(--bdr);
  border-radius: var(--r-pill);
  background: var(--white);
  color: var(--g4);
  cursor: pointer;
  font-family: inherit;
  font-size: var(--text-xs);
  padding: 5px 14px;
  transition:
    background-color 0.12s ease,
    border-color 0.12s ease,
    color 0.12s ease;
}

.trend-tab.on {
  border-color: var(--brand);
  background: var(--brand);
  color: #fff;
}

/* 4. 狀態 */
.trend-state {
  border: 1px solid var(--g2);
  border-radius: 4px;
  background: var(--white);
  color: var(--g4);
  font-size: 13px;
  padding: 18px;
}

.trend-state--error {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  color: var(--error);
}

.trend-state button {
  border: 1px solid var(--bdr);
  border-radius: 4px;
  background: var(--white);
  color: var(--ink);
  cursor: pointer;
  font-family: inherit;
  font-size: 12px;
  padding: 7px 12px;
}

/* 5. 折線圖卡片 */
.chart-wrap {
  border: 1px solid var(--g2);
  border-radius: 4px;
  background: var(--white);
  margin-bottom: 16px;
  padding: 20px;
}

.chart-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 14px;
}

.chart-title {
  color: var(--black);
  font-size: 12px;
  font-weight: 500;
}

.chart-meta {
  color: var(--g4);
  font-size: 11px;
  margin-top: 4px;
}

.chart-source {
  color: var(--accent);
  font-size: 11px;
  text-decoration: none;
}

.svg-chart {
  width: 100%;
  overflow: visible;
}

.chart-legend {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  margin-top: 10px;
}

.legend-item {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: var(--g4);
  font-size: 11px;
}

.legend-line {
  display: inline-block;
  width: 18px;
  height: 3px;
  border-radius: 999px;
}

/* 6. 區段標籤 */
.label-text {
  color: var(--g4);
  font-size: 9px;
  font-weight: 500;
  letter-spacing: 1.5px;
  text-transform: uppercase;
}

/* 7. 各區卡片網格 */
.district-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
  margin-top: 16px;
}

.district-card {
  border: 1px solid var(--g2);
  border-radius: 3px;
  background: var(--white);
  cursor: pointer;
  font-family: inherit;
  padding: 14px;
  text-align: left;
  transition:
    border-color 0.15s ease,
    background-color 0.15s ease;
}

.district-card:hover,
.district-card.on {
  border-color: var(--accent);
}

.district-card.on {
  background: #fffaf5;
}

.dc-name {
  font-size: 12px;
  font-weight: 500;
  margin-bottom: 4px;
}

.dc-price {
  color: var(--ink);
  font-size: 18px;
  font-weight: 300;
}

.dc-change {
  color: var(--g4);
  font-size: 11px;
  margin-top: 3px;
}

.dc-up {
  color: var(--success);
}

.dc-dn {
  color: var(--error);
}

.source-note {
  display: grid;
  gap: 4px;
  color: var(--g4);
  font-size: 11px;
  line-height: 1.6;
  margin-top: 18px;
}

@media (max-width: 767px) {
  .trend-page-wrap {
    padding: 16px;
  }

  .chart-head,
  .trend-state--error {
    align-items: stretch;
    flex-direction: column;
  }

  .district-grid {
    grid-template-columns: 1fr;
  }
}
</style>

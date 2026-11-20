<!--
 * 首頁港鐵即時交通資訊。
 * 1. 顯示最近有效班次並區分大廈附近與公開車站。
 * 2. 依到站時間更新倒數，過時班次移除，不作背景網絡輪詢。
-->
<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';

import type { HomeTransport } from '@/model/home';
import AppIcon from '@/shared/components/base/AppIcon.vue';

import { formatObservationTime } from '../format_observation_time';

const props = defineProps<{ transport?: HomeTransport; loading: boolean; referenceArea?: boolean }>();
const emit = defineEmits<{ retry: [] }>();
const { t, locale } = useI18n();
const now = ref<number>(Date.now());
let timer: ReturnType<typeof setInterval> | undefined;

const updated = computed(() => formatObservationTime(props.transport?.observed_at, locale.value));

// 1. 不保留已過站的班次；只由使用者操作重新抓取。
const arrivals = computed(() => (props.transport?.arrivals ?? [])
  .filter(({ arrival_at: arrivalAt }) => Date.parse(arrivalAt) >= now.value)
  .map(({ destination, arrival_at: arrivalAt }) => ({
    destination,
    arrivalAt,
    minutes: Math.ceil((Date.parse(arrivalAt) - now.value) / 60000),
    time: formatObservationTime(arrivalAt, locale.value),
  })));

const stationScope = computed(() => {
  if (!props.transport?.nearby) return t('home.transport.publicStation');
  return t(props.referenceArea ? 'home.environment.referenceArea' : 'home.transport.nearby');
});

onMounted(() => {
  timer = setInterval(() => { now.value = Date.now(); }, 5000);
});

onUnmounted(() => {
  clearInterval(timer);
});
</script>

<template>
  <section class="home-transport" aria-labelledby="home-transport-title" :aria-busy="loading">
    <!-- 1. 保留城市儀表板查詢與更新操作，無到站資料時仍可使用。 -->
    <header class="transport-heading">
      <h2 id="home-transport-title">{{ t('home.transport.title') }}</h2>
      <div class="transport-actions">
        <a href="https://apis.hk.skylinedances.com/?page=transport" target="_blank" rel="noopener noreferrer">
          {{ t('home.transport.mtr') }} <AppIcon name="arrow-right" :size="16" />
        </a>
        <button
          type="button"
          :title="t('home.transport.refresh')"
          :aria-label="t('home.transport.refresh')"
          :disabled="loading"
          @click="emit('retry')"
        >
          <AppIcon name="reload" :size="18" />
        </button>
      </div>
    </header>
    <div v-if="loading" class="transport-status" role="status">{{ t('home.transport.loading') }}</div>
    <div v-else-if="!transport" class="transport-status" role="status">{{ t('home.transport.unavailable') }}</div>
    <div v-else class="transport-schedule">
      <!-- 2. 車站、資料更新時間及實際到站時間分開顯示。 -->
      <div class="transport-station">
        <div>
          <span class="transport-label">{{ stationScope }}</span>
          <h3>{{ transport.station }}</h3>
          <p class="transport-label">{{ transport.line }}</p>
        </div>
        <span v-if="updated" class="transport-label">{{ t('home.transport.updated', { time: updated }) }}</span>
      </div>
      <p v-if="transport.delayed" class="transport-alert" role="status">{{ t('home.transport.delayed') }}</p>
      <ul v-if="arrivals.length" class="transport-arrivals">
        <li v-for="arrival in arrivals" :key="`${arrival.destination}-${arrival.arrivalAt}`">
          <span class="transport-destination">{{ t('home.transport.boundFor', { destination: arrival.destination }) }}</span>
          <strong>{{ t('home.transport.minutes', { count: arrival.minutes }) }}</strong>
          <span class="transport-label">
            {{ t('home.transport.estimatedArrival') }}
            <time :datetime="arrival.arrivalAt">{{ arrival.time }}</time>
          </span>
        </li>
      </ul>
      <p v-else class="transport-empty" role="status">{{ t('home.transport.noArrivals') }}</p>
    </div>
  </section>
</template>

<style scoped lang="scss">
/* 1. 標題、正式入口與觸控更新按鈕。 */
.transport-heading,
.transport-actions,
.transport-actions a,
.transport-station {
  display: flex;
  align-items: center;
  gap: 8px;
}

.transport-heading {
  justify-content: space-between;
  flex-wrap: wrap;
  margin-bottom: 8px;
}

.transport-heading h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
}

.transport-actions a,
.transport-actions button {
  min-height: 44px;
  color: rgb(var(--color-text-muted));
  font-size: 13px;
}

.transport-actions a:hover {
  text-decoration: underline;
}

.transport-actions button {
  display: grid;
  min-width: 44px;
  place-items: center;
  border: 1px solid rgb(var(--color-border));
  border-radius: 4px;
  cursor: pointer;
}

.transport-actions button:disabled {
  opacity: 0.5;
  cursor: default;
}

.transport-actions :is(a, button):focus-visible {
  outline: 2px solid rgb(var(--color-primary));
  outline-offset: 2px;
}

/* 2. 上下行時間採等寬欄、條目換行。 */
.transport-schedule,
.transport-status {
  border-block: 1px solid rgb(var(--color-border));
}

.transport-station {
  justify-content: space-between;
  align-items: flex-start;
  flex-wrap: wrap;
  padding: 12px;
}

.transport-station h3 {
  margin: 2px 0;
  font-size: 17px;
  font-weight: 600;
}

.transport-label,
.transport-status,
.transport-empty {
  margin: 0;
  color: rgb(var(--color-text-muted));
  font-size: 13px;
}

.transport-status,
.transport-empty {
  padding: 14px 12px;
}

.transport-alert {
  margin: 0;
  padding: 0 12px 12px;
  font-size: 14px;
  font-weight: 600;
}

.transport-arrivals {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  margin: 0;
  padding: 0;
  list-style: none;
}

.transport-arrivals li {
  display: flex;
  min-width: 0;
  min-height: 106px;
  flex-direction: column;
  gap: 2px;
  border-top: 1px solid rgb(var(--color-border));
  padding: 10px 12px;
  overflow-wrap: anywhere;
}

.transport-arrivals li:nth-child(even) {
  border-left: 1px solid rgb(var(--color-border));
}

.transport-destination {
  font-size: 14px;
  font-weight: 600;
}

.transport-arrivals strong {
  font-size: 21px;
  font-weight: 650;
  font-variant-numeric: tabular-nums;
}
</style>

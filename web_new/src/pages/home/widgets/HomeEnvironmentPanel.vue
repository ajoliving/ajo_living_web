<!--
 * 首頁社區環境摘要。
 * 1. 區分大廈與全港資料，標出觀測站或雨量分區。
 * 2. 只顯示可用讀數；載入及重試不影響其他首頁內容。
-->
<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';

import type { HomeEnvironment } from '@/model/home';
import AppIcon from '@/shared/components/base/AppIcon.vue';

import { formatObservationTime } from '../format_observation_time';

const props = defineProps<{ environment: HomeEnvironment | null; loading: boolean }>();
const emit = defineEmits<{ retry: [] }>();
const { t, locale } = useI18n();

const isLocal = computed(() => !!props.environment && props.environment.scope !== 'hong_kong');
const tileCount = computed(() => {
  const { temperature, forecast, rainfall, air_quality: airQuality } = props.environment ?? {};
  return [temperature, forecast, rainfall, airQuality].filter(Boolean).length;
});
// 1. 以香港時間顯示來源的觀測時間，而非瀏覽器的載入時間。
const observationTime = (value: string): string => formatObservationTime(value, locale.value);
</script>

<template>
  <section class="home-environment" aria-labelledby="home-environment-title" :aria-busy="loading">
    <!-- 1. 標題區按已確認位置顯示目前大廈。 -->
    <div class="environment-heading">
      <h2 id="home-environment-title">
        {{ t(isLocal ? 'home.environment.community' : 'home.environment.hongKong') }}
      </h2>
      <span v-if="isLocal && environment?.building_name" class="environment-location">
        <AppIcon name="location" :size="16" />
        {{ environment.building_name }}
        <small v-if="environment.scope === 'reference'">{{ t('home.environment.referenceArea') }}</small>
      </span>
    </div>

    <!-- 2. 各讀數保留原有觀測範圍和時間。 -->
    <div v-if="loading" class="environment-grid" role="status">
      <span class="sr-only">{{ t('home.environment.loading') }}</span>
      <div v-for="slot in 2" :key="slot" class="environment-skeleton" aria-hidden="true"></div>
    </div>
    <div v-else-if="!tileCount" class="environment-unavailable">
      <p role="status">{{ t('home.environment.unavailable') }}</p>
      <button type="button" @click="emit('retry')">
        <AppIcon name="reload" :size="16" />
        {{ t('home.dashboard.retry') }}
      </button>
    </div>
    <dl v-else class="environment-grid" :class="{ 'environment-grid-wide': tileCount > 2 }">
      <div v-if="environment?.temperature" class="environment-metric" data-metric="temperature">
        <dt>{{ t('home.environment.temperature') }}</dt>
        <dd class="environment-value">{{ t('home.environment.celsius', { value: environment.temperature.value }) }}</dd>
        <dd class="environment-detail">
          {{ environment.temperature.station }}
          <time v-if="observationTime(environment.temperature.observed_at)" :datetime="environment.temperature.observed_at">
            {{ observationTime(environment.temperature.observed_at) }}
          </time>
        </dd>
      </div>
      <div v-if="environment?.forecast" class="environment-metric" data-metric="forecast">
        <dt>{{ t('home.environment.forecast') }}</dt>
        <dd class="environment-value">
          {{ t('home.environment.temperatureRange', { min: environment.forecast.min, max: environment.forecast.max }) }}
        </dd>
        <dd v-if="environment.forecast.description" class="environment-detail">
          {{ environment.forecast.description }}
        </dd>
      </div>
      <div v-if="environment?.air_quality" class="environment-metric" data-metric="air-quality">
        <dt>{{ t('home.environment.airQuality') }}</dt>
        <dd class="environment-value">{{ environment.air_quality.value }}</dd>
        <dd class="environment-detail">
          {{ environment.air_quality.station }}
          <time v-if="observationTime(environment.air_quality.observed_at)" :datetime="environment.air_quality.observed_at">
            {{ observationTime(environment.air_quality.observed_at) }}
          </time>
        </dd>
      </div>
      <div v-if="environment?.rainfall" class="environment-metric" data-metric="rainfall">
        <dt>{{ t('home.environment.rainfall') }}</dt>
        <dd class="environment-value">{{ t('home.environment.millimetres', { value: environment.rainfall.value }) }}</dd>
        <dd class="environment-detail">
          {{ environment.rainfall.district }}
          <time v-if="observationTime(environment.rainfall.observed_at)" :datetime="environment.rainfall.observed_at">
            {{ observationTime(environment.rainfall.observed_at) }}
          </time>
        </dd>
      </div>
    </dl>
  </section>
</template>

<style scoped lang="scss">
/* 1. 使用分隔式讀數，手機保持兩欄並容許站名換行。 */
.environment-heading {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 6px 16px;
  margin-bottom: 12px;
}

.environment-heading h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  line-height: 1.4;
}

.environment-location {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  min-width: 0;
  flex-wrap: wrap;
  color: rgb(var(--color-text-muted));
  font-size: 13px;
  overflow-wrap: anywhere;
}

.environment-location svg {
  flex: none;
}

.environment-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  margin: 0;
  border-top: 1px solid rgb(var(--color-border));
  border-bottom: 1px solid rgb(var(--color-border));
}

.environment-grid > :only-child {
  grid-column: 1 / -1;
}

.environment-metric,
.environment-skeleton {
  min-width: 0;
  min-height: 105px;
  padding: 12px;
  overflow-wrap: anywhere;
}

.environment-grid > :nth-child(even) {
  border-left: 1px solid rgb(var(--color-border));
}

.environment-grid > :nth-child(n + 3) {
  border-top: 1px solid rgb(var(--color-border));
}

.environment-metric dt,
.environment-detail {
  color: rgb(var(--color-text-muted));
  font-size: 13px;
  line-height: 1.45;
}

.environment-value {
  margin: 4px 0;
  font-size: 22px;
  font-weight: 650;
  line-height: 1.25;
  font-variant-numeric: tabular-nums;
}

.environment-detail {
  margin: 0;
}

.environment-detail time {
  display: inline-block;
  margin-inline-start: 4px;
}

/* 2. 載入及重試維持清晰的區段和觸控尺寸。 */
.environment-skeleton::before,
.environment-skeleton::after {
  display: block;
  width: 66%;
  height: 13px;
  border-radius: 2px;
  background: rgb(var(--color-surface-muted));
  content: '';
}

.environment-skeleton::after {
  width: 42%;
  height: 22px;
  margin-top: 10px;
}

.environment-unavailable {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  min-height: 80px;
  border-block: 1px solid rgb(var(--color-border));
  padding: 10px 12px;
  color: rgb(var(--color-text-muted));
  font-size: 14px;
}

.environment-unavailable p {
  margin: 0;
}

.environment-unavailable button {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  min-height: 44px;
  padding: 0 8px;
  color: rgb(var(--color-text));
  cursor: pointer;
}

.environment-unavailable button:focus-visible {
  outline: 2px solid rgb(var(--color-primary));
  outline-offset: 2px;
}

@media (min-width: 768px) {
  .environment-grid-wide {
    grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
  }

  .environment-grid-wide > :nth-child(n + 2) {
    border-top: 0;
    border-left: 1px solid rgb(var(--color-border));
  }
}
</style>

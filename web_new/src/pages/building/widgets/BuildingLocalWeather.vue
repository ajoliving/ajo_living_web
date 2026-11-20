<!--
 * 目前大廈天氣及生活提示。
 * 1. 只查詢已選定的大廈，不套用公開參考屋苑。
 * 2. 預報獨立載入，不阻塞大廈服務。
-->
<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { fetchHomeEnvironment } from '@/httpapis/home';
import type { HomeEnvironment } from '@/model/home';

const props = defineProps<{ buildingId: string }>();
const { locale, t } = useI18n();
const data = ref<HomeEnvironment | null>(null);
const loading = ref(true);
const refresh = ref(0);
const dateFormat = computed(() => new Intl.DateTimeFormat(locale.value, {
  timeZone: 'Asia/Hong_Kong', weekday: 'short', month: 'numeric', day: 'numeric',
}));

// 1. 物業及語系切換時取消舊請求。
watch([() => props.buildingId, locale, refresh], async ([buildingId, lang], _, cleanup) => {
  const controller = new AbortController();
  cleanup(() => controller.abort());
  data.value = null;
  loading.value = true;
  try {
    const response = await fetchHomeEnvironment({
      building_id: buildingId, lang: lang === 'en' ? 'en' : 'tc', detail: 'forecast',
    }, controller.signal);
    if (!controller.signal.aborted) data.value = response.data.data;
  } catch {
    // 1.1 不影響大廈服務或物業選擇。
  } finally {
    if (!controller.signal.aborted) loading.value = false;
  }
}, { immediate: true });
</script>

<template>
  <section class="building-local-weather" :aria-busy="loading">
    <h2>{{ t(data?.scope === 'building' ? 'home.environment.community' : 'home.environment.hongKong') }}</h2>
    <p v-if="loading" role="status">{{ t('home.environment.loading') }}</p>
    <template v-else-if="data?.temperature || data?.forecast_days?.length">
      <p v-if="data.temperature" class="weather-current">
        <strong>{{ t('home.environment.celsius', { value: data.temperature.value }) }}</strong>
        <span>{{ data.temperature.station }}</span>
      </p>
      <p v-if="data.humidity != null || data.uv != null" class="weather-current">
        <span v-if="data.humidity != null">{{ t('home.environment.humidity', { value: data.humidity }) }}</span>
        <span v-if="data.uv != null">{{ t('home.environment.uv', { value: data.uv }) }}</span>
      </p>
      <div v-if="data.forecast_days?.length">
        <h3>{{ t('home.environment.nineDay') }}</h3>
        <ul class="weather-days" tabindex="0" :aria-label="t('home.environment.nineDay')">
          <li v-for="day in data.forecast_days" :key="day.date">
            <time :datetime="day.date">{{ dateFormat.format(new Date(day.date + 'T12:00:00+08:00')) }}</time>
            <strong>{{ t('home.environment.temperatureRange', { min: day.min, max: day.max }) }}</strong>
            <span>{{ day.description }}</span>
          </li>
        </ul>
      </div>
      <p v-if="data.emergency?.nearby || data.parking || data.holiday" class="weather-nearby">
        <span v-if="data.emergency?.nearby">{{ data.emergency.hospital }} · {{ t('home.life.otherMedian') }} {{ data.emergency.wait }}</span>
        <span v-if="data.parking?.spaces != null">{{ data.parking.name }} · {{ t('home.life.availableSpaces', { count: data.parking.spaces }) }}</span>
        <span v-if="data.holiday">{{ data.holiday.name }} · {{ data.holiday.date }}</span>
      </p>
    </template>
    <div v-else class="weather-current">
      <p role="status">{{ t('home.environment.unavailable') }}</p>
      <button type="button" @click="refresh++">{{ t('home.dashboard.retry') }}</button>
    </div>
  </section>
</template>

<style scoped>
.building-local-weather { min-width: 0; padding: 16px 0; border-block: 1px solid rgb(var(--color-border)); }
h2 { margin: 0 0 8px; font-size: 18px; font-weight: 650; }
h3 { margin: 12px 0 6px; font-size: 14px; font-weight: 600; }
.weather-current, .weather-nearby { display: flex; flex-wrap: wrap; align-items: center; gap: 6px 14px; margin: 6px 0; font-size: 13px; }
.weather-current strong { font-size: 22px; }
.weather-days { display: grid; grid-auto-flow: column; grid-auto-columns: minmax(108px, 1fr); gap: 8px; overflow-x: auto; margin: 0; padding: 4px 2px 7px; list-style: none; }
.weather-days li { display: grid; align-content: start; gap: 4px; padding: 8px; border: 1px solid rgb(var(--color-border)); border-radius: 6px; font-size: 12px; overflow-wrap: anywhere; }
.weather-days strong { font-size: 14px; }
.weather-days li span, .weather-nearby { color: rgb(var(--color-text-muted)); }
.weather-nearby { margin-top: 12px; }
button { min-height: 44px; padding-inline: 8px; text-decoration: underline; }
</style>

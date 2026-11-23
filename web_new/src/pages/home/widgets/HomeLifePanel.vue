<!--
 * 首頁生活提示。
 * 1. 顯示公眾假期、私家車位及官方急症室輪候統計。
 * 2. 無即時讀數時保留城市儀表板查詢入口，不以其他位置或模擬資料替代。
 * 3. 查詢入口統一置於各項標題右側，摘要顯示於下方。
-->
<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';

import type { HomeEnvironment } from '@/model/home';
import AppIcon from '@/shared/components/base/AppIcon.vue';

import { formatObservationTime } from '../format_observation_time';

const props = defineProps<{ environment: HomeEnvironment | null; loading: boolean }>();
const { t, locale } = useI18n();

// 1. 對齊香港日期及資料更新時間，不將輪候統計視為個人預測。
const holidayDate = computed(() => {
  const date = new Date(`${props.environment?.holiday?.date}T00:00:00+08:00`);
  return Number.isNaN(date.getTime()) ? '' : new Intl.DateTimeFormat(locale.value, {
    dateStyle: 'medium',
    timeZone: 'Asia/Hong_Kong',
  }).format(date);
});
const parkingTime = computed(() => formatObservationTime(props.environment?.parking?.observed_at, locale.value));
const emergencyTime = computed(() => formatObservationTime(props.environment?.emergency?.observed_at, locale.value));
</script>

<template>
  <section class="home-life" aria-labelledby="home-life-title" :aria-busy="loading">
    <header class="home-section-heading">
      <h2 id="home-life-title">{{ t('home.life.title') }}</h2>
    </header>
    <p v-if="loading" class="sr-only" role="status">{{ t('home.life.loading') }}</p>
    <div class="life-grid">
      <!-- 1. 缺少讀數仍可前往城市儀表板。 -->
      <article class="life-item">
        <header class="life-item-heading">
          <h3>
            {{ t(environment?.holiday
              ? environment.holiday.is_today ? 'home.life.todayHoliday' : 'home.life.nextHoliday'
              : 'home.life.publicHoliday') }}
          </h3>
          <a href="https://apis.hk.skylinedances.com/?page=holidays" target="_blank" rel="noopener noreferrer">
            <span>{{ t('home.life.holidaySource') }}</span><AppIcon name="arrow-right" :size="16" />
          </a>
        </header>
        <div v-if="environment?.holiday" class="life-reading">
          <strong>{{ environment.holiday.name }}</strong>
          <time v-if="holidayDate" :datetime="environment.holiday.date">{{ holidayDate }}</time>
        </div>
      </article>

      <!-- 2. 只使用確認位置的車位；零個空位仍屬有效讀數。 -->
      <article class="life-item">
        <header class="life-item-heading">
          <h3>{{ t('home.life.parking') }}</h3>
          <a href="https://apis.hk.skylinedances.com/?page=parking" target="_blank" rel="noopener noreferrer">
            <span>{{ t('home.life.parkingSource') }}</span><AppIcon name="arrow-right" :size="16" />
          </a>
        </header>
        <div v-if="environment?.parking" class="life-reading">
          <strong v-if="environment.parking.closed">{{ t('home.life.parkingClosed') }}</strong>
          <strong v-else-if="environment.parking.spaces != null">{{ t('home.life.availableSpaces', { count: environment.parking.spaces }) }}</strong>
          <p>{{ environment.parking.name }}</p>
          <small v-if="environment.scope === 'reference'">{{ t('home.environment.referenceArea') }}</small>
          <time v-if="parkingTime" :datetime="environment.parking.observed_at">{{ parkingTime }}</time>
        </div>
        <p v-else-if="!loading">{{ t('home.life.nearbyParking') }}</p>
      </article>

      <!-- 3. 依官方分流分類顯示中位數，不推薦就醫地點。 -->
      <article class="life-item">
        <header class="life-item-heading">
          <h3>{{ t('home.life.emergency') }}</h3>
          <a href="https://apis.hk.skylinedances.com/?page=health" target="_blank" rel="noopener noreferrer">
            <span>{{ t('home.life.emergencySource') }}</span><AppIcon name="arrow-right" :size="16" />
          </a>
        </header>
        <div v-if="environment?.emergency" class="life-reading">
          <p>{{ environment.emergency.hospital }}</p>
          <small v-if="environment.scope === 'reference'">{{ t('home.environment.referenceArea') }}</small>
          <small v-else-if="!environment.emergency.nearby">{{ t('home.life.publicWait') }}</small>
          <strong>{{ environment.emergency.wait }}</strong>
          <p>{{ t(environment.emergency.category === '3' ? 'home.life.urgentMedian' : 'home.life.otherMedian') }}</p>
          <time v-if="emergencyTime" :datetime="environment.emergency.observed_at">{{ emergencyTime }}</time>
        </div>
        <p v-else-if="!loading">{{ t('home.life.noWait') }}</p>
      </article>
    </div>
  </section>
</template>

<style scoped lang="scss">
/* 1. 標題列與交通資訊對齊；手機採單欄、桌面三欄。 */
.home-section-heading {
  display: flex;
  align-items: center;
  min-height: 44px;
  margin-bottom: 8px;
}

.home-section-heading h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
}

.life-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  border-block: 1px solid rgb(var(--color-border));
}

.life-item {
  min-width: 0;
  padding: 8px 12px 12px;
  overflow-wrap: anywhere;
}

.life-item + .life-item {
  border-top: 1px solid rgb(var(--color-border));
}

.life-item-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  min-height: 44px;
  margin-bottom: 4px;
}

.life-item h3 {
  min-width: 0;
  margin: 0;
  font-size: 15px;
  font-weight: 600;
  line-height: 1.4;
}

.life-item p,
.life-item small,
.life-item time {
  margin: 0;
  color: rgb(var(--color-text-muted));
  font-size: 13px;
  line-height: 1.45;
}

.life-reading {
  display: grid;
  gap: 4px;
}

.life-reading strong {
  font-size: 16px;
  line-height: 1.4;
}

.life-item a {
  display: inline-flex;
  align-items: center;
  justify-content: flex-end;
  flex-shrink: 0;
  gap: 6px;
  max-width: 60%;
  min-height: 44px;
  color: rgb(var(--color-text-muted));
  font-size: 13px;
  line-height: 1.45;
  text-align: right;
}

.life-item a:hover {
  text-decoration: underline;
}

.life-item a:focus-visible {
  outline: 2px solid rgb(var(--color-primary));
  outline-offset: 2px;
}

.life-item a svg {
  flex: none;
}

@media (min-width: 768px) {
  .life-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .life-item + .life-item {
    border-top: 0;
    border-left: 1px solid rgb(var(--color-border));
  }
}
</style>

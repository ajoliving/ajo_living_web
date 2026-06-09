<!--
 * 廣告任務列表面板。
 * 1. 查詢與展示 Staff 廣告任務。
 * 2. 提供編輯與啟停操作入口。
-->
<script setup lang="ts">
import type { StaffRewardAdResponse } from '@/domains/payments/model';
import AppIcon from '@/shared/components/base/AppIcon.vue';

defineProps<{
  adKeyword: string;
  formatAjoPoints: (value: number) => string;
  loadingAds: boolean;
  rewardAds: StaffRewardAdResponse[];
  t: (key: string, params?: Record<string, unknown>) => string;
}>();

const emit = defineEmits<{
  'update:adKeyword': [value: string];
  edit: [ad: StaffRewardAdResponse];
  load: [];
  toggle: [ad: StaffRewardAdResponse];
}>();
</script>

<template>
  <article class="wallet-settings-panel">
    <div class="wallet-settings-panel-header">
      <div>
        <h2>{{ t('marketplace.management.walletAdsTitle') }}</h2>
        <p>{{ t('marketplace.management.walletAdsDescription') }}</p>
      </div>
      <button
        type="button"
        class="wallet-settings-icon-button"
        :aria-label="t('marketplace.management.refresh')"
        @click="emit('load')"
      >
        <AppIcon
          name="reload"
          :size="16"
        />
      </button>
    </div>

    <div class="wallet-settings-toolbar">
      <label class="wallet-settings-search">
        <AppIcon
          name="search"
          :size="16"
        />
        <input
          :value="adKeyword"
          type="search"
          :placeholder="t('marketplace.management.walletAdSearchPlaceholder')"
          @input="emit('update:adKeyword', ($event.target as HTMLInputElement).value)"
          @keyup.enter="emit('load')"
        />
      </label>
      <button
        type="button"
        class="wallet-settings-secondary-button"
        @click="emit('load')"
      >
        {{ t('marketplace.list.searchAction') }}
      </button>
    </div>

    <div
      v-if="loadingAds"
      class="wallet-settings-empty"
    >
      {{ t('common.status.loading') }}
    </div>
    <div
      v-else-if="rewardAds.length === 0"
      class="wallet-settings-empty"
    >
      {{ t('marketplace.management.walletNoAds') }}
    </div>
    <div
      v-else
      class="wallet-settings-list"
    >
      <article
        v-for="ad in rewardAds"
        :key="ad.task_id"
        class="wallet-settings-ad-card"
      >
        <div class="wallet-settings-ad-cover">
          <video
            v-if="ad.media_type === 'video' && ad.media_url"
            :src="ad.media_url"
            preload="metadata"
          />
          <img
            v-else-if="ad.media_url"
            :src="ad.media_url"
            :alt="ad.title"
          />
          <AppIcon
            v-else
            name="view"
            :size="28"
          />
        </div>
        <div class="wallet-settings-ad-body">
          <div class="wallet-settings-card-title">
            <h3>{{ ad.title }}</h3>
            <span :class="ad.is_active ? 'wallet-status-active' : 'wallet-status-idle'">
              {{ ad.is_active ? t('common.state.active') : t('marketplace.mine.hidden') }}
            </span>
          </div>
          <p>{{ ad.summary || t('marketplace.management.walletAdNoSummary') }}</p>
          <div class="wallet-settings-meta">
            <span>{{ ad.media_type === 'video' ? t('marketplace.management.walletAdMediaTypeVideo') : t('marketplace.management.walletAdMediaTypeImage') }}</span>
            <span>{{ formatAjoPoints(ad.reward_points) }}</span>
            <span>{{ t('marketplace.management.walletAdWatchSecondsValue', { seconds: ad.watch_seconds }) }}</span>
            <span>{{ t('marketplace.management.walletAdWatchCountValue', { count: ad.watch_count }) }}</span>
            <span>{{ t('marketplace.management.walletAdWatchTimeValue', { seconds: ad.total_watch_seconds }) }}</span>
            <span>{{ t('marketplace.management.walletAdClickRateValue', { rate: ad.link_click_rate.toFixed(2) }) }}</span>
            <span v-if="ad.ends_at">{{ t('marketplace.management.walletAdEndsAtValue', { date: ad.ends_at.slice(0, 10) }) }}</span>
          </div>
          <div class="wallet-settings-actions wallet-settings-actions-left">
            <button
              type="button"
              class="wallet-settings-mini-button"
              @click="emit('edit', ad)"
            >
              {{ t('marketplace.management.walletAdEditAction') }}
            </button>
            <button
              type="button"
              class="wallet-settings-mini-button"
              @click="emit('toggle', ad)"
            >
              {{ ad.is_active ? t('marketplace.management.walletAdDisableAction') : t('marketplace.management.walletAdEnableAction') }}
            </button>
          </div>
        </div>
      </article>
    </div>
  </article>
</template>

<!--
 * 廣告列表頁。
 * 1. 列出 Staff 建立的看廣告得積分任務。
 * 2. 提供上架、下架與編輯入口。
-->
<script setup lang="ts">
import RewardAdListPanel from '../wallet/widgets/RewardAdListPanel.vue';
import { useRewardAdListPage } from '../wallet/wallet';

import '../wallet/styles.scss';

const {
  adKeyword,
  editRewardAd,
  formatAjoPoints,
  loadRewardAds,
  loadingAds,
  rewardAds,
  t,
  toggleRewardAd,
} = useRewardAdListPage();
</script>

<template>
  <section class="wallet-settings-page">
    <header class="wallet-settings-header">
      <div>
        <p class="wallet-settings-kicker">
          {{ t('common.brand.pointsName') }}
        </p>
        <h1>{{ t('marketplace.management.walletAdListSection') }}</h1>
        <p>{{ t('marketplace.management.walletAdsDescription') }}</p>
      </div>
    </header>

    <RewardAdListPanel
      v-model:ad-keyword="adKeyword"
      :format-ajo-points="formatAjoPoints"
      :loading-ads="loadingAds"
      :reward-ads="rewardAds"
      :t="t"
      @edit="editRewardAd"
      @load="loadRewardAds"
      @toggle="toggleRewardAd"
    />
  </section>
</template>

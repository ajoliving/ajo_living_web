<!--
 * 廣告發布頁。
 * 1. 建立與編輯看廣告得積分任務。
 * 2. 支援圖片與影片廣告媒體上傳至指定 OSS 目錄。
-->
<script setup lang="ts">
import RewardAdEditorPanel from '../wallet/widgets/RewardAdEditorPanel.vue';
import { useRewardAdEditorPage } from '../wallet/wallet';

import '../wallet/styles.scss';

const {
  adForm,
  isEditingAd,
  loadingAd,
  resetRewardAdForm,
  saveRewardAd,
  savingAd,
  stageRewardAdMedia,
  t,
  uploadingMedia,
} = useRewardAdEditorPage();
</script>

<template>
  <section class="wallet-settings-page">
    <header class="wallet-settings-header">
      <div>
        <p class="wallet-settings-kicker">
          {{ t('common.brand.pointsName') }}
        </p>
        <h1>{{ isEditingAd ? t('marketplace.management.walletAdEditSection') : t('marketplace.management.walletAdPublishSection') }}</h1>
        <p>{{ t('marketplace.management.walletAdEditorDescription') }}</p>
      </div>
    </header>

    <div
      v-if="loadingAd"
      class="wallet-settings-empty"
    >
      {{ t('common.status.loading') }}
    </div>

    <RewardAdEditorPanel
      v-else
      :ad-form="adForm"
      :is-editing-ad="isEditingAd"
      :saving-ad="savingAd || uploadingMedia"
      :t="t"
      @reset="resetRewardAdForm"
      @save="saveRewardAd"
      @stage-media="stageRewardAdMedia"
    />
  </section>
</template>

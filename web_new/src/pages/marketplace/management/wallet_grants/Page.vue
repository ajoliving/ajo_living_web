<!--
 * 積分發放頁。
 * 1. 提供 Staff 會員列表式 AJO Point 發放入口。
 * 2. 將每行會員發放結果寫入不可變積分流水。
-->
<script setup lang="ts">
import WalletGrantPanel from '../wallet/widgets/WalletGrantPanel.vue';
import { useWalletGrantPage } from '../wallet/wallet';

import '../wallet/styles.scss';

const {
  grantRows,
  grantPoints,
  hasNextUsers,
  hasPreviousUsers,
  loadNextGrantUsers,
  loadingUsers,
  loadPreviousGrantUsers,
  searchGrantUsers,
  t,
  userKeyword,
  userPage,
  userPageSize,
  userTotal,
} = useWalletGrantPage();
</script>

<template>
  <section class="wallet-settings-page">
    <header class="wallet-settings-header">
      <div>
        <p class="wallet-settings-kicker">
          {{ t('common.brand.pointsName') }}
        </p>
        <h1>{{ t('marketplace.management.walletGrantSection') }}</h1>
        <p>{{ t('marketplace.management.walletGrantDescription') }}</p>
      </div>
    </header>

    <WalletGrantPanel
      v-model:user-keyword="userKeyword"
      :grant-rows="grantRows"
      :has-next-users="hasNextUsers"
      :has-previous-users="hasPreviousUsers"
      :loading-users="loadingUsers"
      :t="t"
      :user-page="userPage"
      :user-page-size="userPageSize"
      :user-total="userTotal"
      @grant="grantPoints"
      @next="loadNextGrantUsers"
      @previous="loadPreviousGrantUsers"
      @search="searchGrantUsers"
    />
  </section>
</template>

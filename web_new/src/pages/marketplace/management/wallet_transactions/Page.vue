<!--
 * 積分流水頁。
 * 1. 查詢 Staff 可見的平台積分流水。
 * 2. 支援會員與方向篩選。
-->
<script setup lang="ts">
import WalletTransactionPanel from '../wallet/widgets/WalletTransactionPanel.vue';
import { useWalletTransactionPage } from '../wallet/wallet';

import '../wallet/styles.scss';

const {
  formatAjoPoints,
  formatDate,
  loadTransactions,
  loadingTransactions,
  t,
  transactionDirection,
  transactions,
  transactionUserId,
} = useWalletTransactionPage();
</script>

<template>
  <section class="wallet-settings-page">
    <header class="wallet-settings-header">
      <div>
        <p class="wallet-settings-kicker">
          {{ t('common.brand.pointsName') }}
        </p>
        <h1>{{ t('marketplace.management.walletTransactionsSection') }}</h1>
        <p>{{ t('marketplace.management.walletTransactionsDescription') }}</p>
      </div>
    </header>

    <WalletTransactionPanel
      v-model:transaction-direction="transactionDirection"
      v-model:transaction-user-id="transactionUserId"
      :format-ajo-points="formatAjoPoints"
      :format-date="formatDate"
      :loading-transactions="loadingTransactions"
      :t="t"
      :transactions="transactions"
      @load="loadTransactions"
    />
  </section>
</template>

<!--
 * 積分流水列表面板。
 * 1. 展示 Staff 可查看的最近積分流水。
 * 2. 提供會員與方向篩選。
-->
<script setup lang="ts">
import type { StaffWalletTransactionResponse } from '@/model/wallet';
import AppIcon from '@/shared/components/base/AppIcon.vue';
import { formatWalletTransactionSource } from '@/utils/wallet';

defineProps<{
  formatAjoPoints: (value: number) => string;
  formatDate: (value: string) => string;
  loadingTransactions: boolean;
  t: (key: string, params?: Record<string, unknown>) => string;
  transactionDirection: string;
  transactions: StaffWalletTransactionResponse[];
  transactionUserId: string;
}>();

const emit = defineEmits<{
  'update:transactionDirection': [value: string];
  'update:transactionUserId': [value: string];
  load: [];
}>();
</script>

<template>
  <article class="wallet-settings-panel">
    <div class="wallet-settings-toolbar wallet-settings-toolbar-three">
      <label class="wallet-settings-search">
        <AppIcon
          name="user"
          :size="16"
        />
        <input
          :value="transactionUserId"
          type="search"
          :placeholder="t('marketplace.management.walletTransactionUserPlaceholder')"
          @input="emit('update:transactionUserId', ($event.target as HTMLInputElement).value)"
          @keyup.enter="emit('load')"
        />
      </label>
      <select
        :value="transactionDirection"
        class="wallet-settings-select"
        @change="emit('update:transactionDirection', ($event.target as HTMLSelectElement).value); emit('load')"
      >
        <option value="">
          {{ t('marketplace.management.walletDirectionAll') }}
        </option>
        <option value="credit">
          {{ t('account.wallet.credit') }}
        </option>
        <option value="debit">
          {{ t('account.wallet.debit') }}
        </option>
      </select>
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

    <div
      v-if="loadingTransactions"
      class="wallet-settings-empty"
    >
      {{ t('common.status.loading') }}
    </div>
    <div
      v-else-if="transactions.length === 0"
      class="wallet-settings-empty"
    >
      {{ t('account.wallet.noTransactions') }}
    </div>
    <div
      v-else
      class="wallet-settings-transaction-list"
    >
      <article
        v-for="transaction in transactions"
        :key="transaction.transaction_id"
        class="wallet-settings-transaction-row"
      >
        <div>
          <strong>
            {{ transaction.direction === 'credit' ? t('account.wallet.credit') : t('account.wallet.debit') }}
            {{ formatAjoPoints(transaction.amount) }}
          </strong>
          <span>{{ transaction.target_user.display_name || transaction.target_user.user_id }}</span>
          <small>{{ formatWalletTransactionSource(transaction, t) }}</small>
        </div>
        <time>{{ formatDate(transaction.created_at) }}</time>
      </article>
    </div>
  </article>
</template>

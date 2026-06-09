<!--
 * 積分發放面板。
 * 1. 接收 Staff 會員列表與逐行發放狀態。
 * 2. 觸發會員搜尋、分頁與單行發放操作。
-->
<script setup lang="ts">
import AppIcon from '@/shared/components/base/AppIcon.vue';

import type { StaffWalletGrantRow } from '../wallet';

const props = defineProps<{
  grantRows: StaffWalletGrantRow[];
  hasNextUsers: boolean;
  hasPreviousUsers: boolean;
  loadingUsers: boolean;
  t: (key: string, params?: Record<string, unknown>) => string;
  userKeyword: string;
  userPage: number;
  userPageSize: number;
  userTotal: number;
}>();

const emit = defineEmits<{
  grant: [row: StaffWalletGrantRow];
  next: [];
  previous: [];
  search: [];
  'update:userKeyword': [value: string];
}>();

// 1. 格式化會員電話
const formatMemberPhone = (row: StaffWalletGrantRow): string => {
  if (!row.user.phone_number) {
    return '-';
  }
  return `${row.user.phone_country_code} ${row.user.phone_number}`.trim();
};

// 2. 格式化會員名稱
const formatMemberName = (row: StaffWalletGrantRow): string =>
  row.user.display_name || row.user.phone_number || row.user.public_id;

// 3. 格式化會員社區
const formatMemberCommunity = (row: StaffWalletGrantRow): string =>
  row.user.primary_community?.name_zh || row.user.primary_community?.name_en || '';

// 4. 計算目前列表範圍
const formatPaginationRange = (): string => {
  if (props.userTotal <= 0) {
    return props.t('marketplace.management.walletGrantPaginationEmpty');
  }

  const start = (props.userPage - 1) * props.userPageSize + 1;
  const end = Math.min(props.userPage * props.userPageSize, props.userTotal);
  return props.t('marketplace.management.walletGrantPaginationRange', {
    start,
    end,
    total: props.userTotal,
  });
};
</script>

<template>
  <article class="wallet-settings-panel">
    <div class="wallet-settings-panel-header">
      <div>
        <h2>{{ t('marketplace.management.walletGrantTitle') }}</h2>
        <p>{{ t('marketplace.management.walletGrantDescription') }}</p>
      </div>
    </div>

    <div class="wallet-settings-toolbar wallet-settings-grant-toolbar">
      <label class="wallet-settings-search">
        <AppIcon
          name="search"
          :size="16"
        />
        <input
          :value="userKeyword"
          type="search"
          autocomplete="off"
          :placeholder="t('marketplace.management.walletGrantSearchPlaceholder')"
          @input="emit('update:userKeyword', ($event.target as HTMLInputElement).value)"
          @keydown.enter="emit('search')"
        />
      </label>

      <button
        type="button"
        class="wallet-settings-secondary-button"
        :disabled="loadingUsers"
        @click="emit('search')"
      >
        <AppIcon
          name="search"
          :size="16"
        />
        <span>{{ t('marketplace.management.walletGrantSearchAction') }}</span>
      </button>
    </div>

    <div
      v-if="loadingUsers"
      class="wallet-settings-empty"
    >
      {{ t('marketplace.management.walletGrantLoadingMembers') }}
    </div>

    <div
      v-else-if="grantRows.length === 0"
      class="wallet-settings-empty"
    >
      {{ t('marketplace.management.walletGrantEmpty') }}
    </div>

    <div
      v-else
      class="wallet-settings-grant-list"
    >
      <article
        v-for="row in grantRows"
        :key="row.user.public_id"
        class="wallet-settings-grant-row"
      >
        <div class="wallet-settings-member-cell">
          <strong>{{ formatMemberName(row) }}</strong>
          <small>{{ row.user.public_id }}</small>
          <div class="wallet-settings-member-meta">
            <span>{{ formatMemberPhone(row) }}</span>
            <span>{{ row.user.is_staff ? 'staff' : 'user' }}</span>
            <span v-if="row.user.primary_community">
              {{ formatMemberCommunity(row) }}
            </span>
            <span
              v-if="row.user.is_staff"
              class="wallet-settings-member-badge"
            >
              {{ t('marketplace.management.walletGrantStaffBadge') }}
            </span>
          </div>
        </div>

        <label class="wallet-settings-field wallet-settings-compact-field">
          <span>{{ t('marketplace.management.walletAmountField') }}</span>
          <input
            v-model.number="row.amount"
            type="number"
            min="1"
            max="1000000"
            step="1"
          />
        </label>

        <label class="wallet-settings-field wallet-settings-compact-field">
          <span>{{ t('marketplace.management.walletNoteField') }}</span>
          <input
            v-model="row.note"
            type="text"
            maxlength="500"
            :placeholder="t('marketplace.management.walletNotePlaceholder')"
          />
        </label>

        <button
          type="button"
          class="wallet-settings-primary-button"
          :disabled="row.granting"
          @click="emit('grant', row)"
        >
          <AppIcon
            name="star"
            :size="16"
          />
          <span>{{ row.granting ? t('marketplace.management.walletGranting') : t('marketplace.management.walletGrantAction') }}</span>
        </button>
      </article>
    </div>

    <footer class="wallet-settings-grant-footer">
      <span>{{ formatPaginationRange() }}</span>
      <div class="wallet-settings-actions">
        <button
          type="button"
          class="wallet-settings-secondary-button"
          :disabled="loadingUsers || !hasPreviousUsers"
          @click="emit('previous')"
        >
          {{ t('marketplace.management.walletGrantPrevious') }}
        </button>
        <button
          type="button"
          class="wallet-settings-secondary-button"
          :disabled="loadingUsers || !hasNextUsers"
          @click="emit('next')"
        >
          {{ t('marketplace.management.walletGrantNext') }}
        </button>
      </div>
    </footer>
  </article>
</template>

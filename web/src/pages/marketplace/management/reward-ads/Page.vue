<!--
 * 廣告列表頁。
 * 1. 以管理列表版型列出 Staff 建立的看廣告得積分任務。
 * 2. 提供搜尋、分頁、啟停與編輯入口。
-->
<script setup lang="ts">
import { computed, ref } from 'vue';
import { useRouter } from 'vue-router';

import { fetchStaffRewardAds, renewStaffRewardAd, updateStaffRewardAd } from '@/httpapis/wallet';
import type { StaffRewardAdResponse } from '@/model/wallet';
import AppIcon from '@/shared/components/base/AppIcon.vue';
import { formatDate } from '@/utils/format';
import { formatAjoPoints } from '@/utils/wallet';

import {
  type StaffManagementListParams,
  type StaffManagementListFetcher,
  useStaffManagementList,
} from '../composables/useStaffManagementList';
import ManagementPagination from '../widgets/ManagementPagination.vue';
import '../styles.scss';

// 1. 查詢廣告列表
const fetchRewardAdItems: StaffManagementListFetcher<StaffRewardAdResponse> = (
  params: StaffManagementListParams,
) =>
  fetchStaffRewardAds({
    page: params.page,
    page_size: params.page_size,
    keyword: params.keyword,
    is_active: params.status ? params.status === 'active' : undefined,
  });

const router = useRouter();
const {
  hasNext,
  hasPrevious,
  items,
  keyword,
  loading,
  next,
  page,
  pageSize,
  previous,
  runAction,
  search,
  status,
  t,
  total,
} = useStaffManagementList<StaffRewardAdResponse>({
  fetcher: fetchRewardAdItems,
  loadErrorKey: 'marketplace.management.walletLoadAdsError',
  updateErrorKey: 'marketplace.management.walletAdSaveError',
  updateSuccessKey: 'marketplace.management.walletAdStatusUpdated',
});
const renewalDialogAd = ref<StaffRewardAdResponse | null>(null);
const renewalDays = ref(14);
const renewing = ref(false);
const renewalDayOptions = [7, 14, 30, 60, 90];
const canSubmitRenewal = computed(() =>
  Number.isInteger(Number(renewalDays.value)) &&
  Number(renewalDays.value) > 0 &&
  Number(renewalDays.value) <= 365,
);

// 2. 格式化廣告狀態
const formatAdStatus = (ad: StaffRewardAdResponse): string =>
  ad.is_active ? t('common.state.active') : t('common.state.hidden');

// 3. 格式化廣告媒體類型
const formatAdMediaType = (ad: StaffRewardAdResponse): string =>
  ad.media_type === 'video'
    ? t('marketplace.management.walletAdMediaTypeVideo')
    : t('marketplace.management.walletAdMediaTypeImage');

// 4. 格式化積分
const formatPoints = (value: number): string =>
  formatAjoPoints(value, t('common.brand.pointsName'), 'zh-HK');

// 5. 格式化可選日期
const formatOptionalDate = (value?: string | null): string =>
  value ? formatDate(value) : '-';

// 6. 跳轉廣告編輯頁
const editRewardAd = async (ad: StaffRewardAdResponse): Promise<void> => {
  await router.push(`/account/marketplace/management/reward-ad-editor/${ad.task_id}`);
};

// 7. 啟停廣告任務
const toggleRewardAd = async (ad: StaffRewardAdResponse): Promise<void> => {
  await runAction(() => updateStaffRewardAd(ad.task_id, { is_active: !ad.is_active }));
};

// 8. 開啟續期彈窗
const openRenewalDialog = (ad: StaffRewardAdResponse): void => {
  renewalDialogAd.value = ad;
  renewalDays.value = 14;
};

// 9. 關閉續期彈窗
const closeRenewalDialog = (): void => {
  renewalDialogAd.value = null;
  renewalDays.value = 14;
};

// 10. 提交續期設定
const submitRenewal = async (): Promise<void> => {
  if (!renewalDialogAd.value || !canSubmitRenewal.value) {
    return;
  }

  const taskId = renewalDialogAd.value.task_id;
  const targetRenewalDays = Number(renewalDays.value);
  renewing.value = true;
  const renewed = await runAction(() => renewStaffRewardAd(taskId, {
    retention_days: targetRenewalDays,
  }));
  renewing.value = false;
  if (renewed) {
    closeRenewalDialog();
  }
};
</script>

<template>
  <section class="management-list-page">
    <header class="management-list-header">
      <div>
        <p class="management-list-kicker">
          {{ t('common.brand.pointsName') }}
        </p>
        <h1>{{ t('marketplace.management.walletAdListSection') }}</h1>
        <p>{{ t('marketplace.management.walletAdsDescription') }}</p>
      </div>
    </header>

    <article class="management-list-panel">
      <div class="management-list-toolbar">
        <label class="management-list-search">
          <AppIcon
            name="search"
            :size="16"
          />
          <input
            v-model="keyword"
            type="search"
            :placeholder="t('marketplace.management.walletAdSearchPlaceholder')"
            @keyup.enter="search"
          />
        </label>
        <select
          v-model="status"
          class="management-list-select"
        >
          <option value="">
            {{ t('marketplace.management.allStatuses') }}
          </option>
          <option value="active">
            {{ t('common.state.active') }}
          </option>
          <option value="hidden">
            {{ t('common.state.hidden') }}
          </option>
        </select>
        <button
          type="button"
          class="management-list-button"
          @click="search"
        >
          {{ t('marketplace.list.searchAction') }}
        </button>
      </div>

      <div
        v-if="loading"
        class="management-list-loading"
      >
        {{ t('common.status.loading') }}
      </div>

      <div
        v-else-if="items.length === 0"
        class="management-list-empty"
      >
        {{ t('marketplace.management.walletNoAds') }}
      </div>

      <div
        v-else
        class="management-table-wrap"
      >
        <table class="management-table management-table--reward-ads">
          <thead>
            <tr>
              <th>{{ t('marketplace.management.columnTitle') }}</th>
              <th>{{ t('marketplace.management.walletAdMediaTypeField') }}</th>
              <th>{{ t('marketplace.management.walletAdRewardField') }}</th>
              <th>{{ t('common.label.status') }}</th>
              <th>{{ t('marketplace.management.walletAdWatchCountColumn') }}</th>
              <th>{{ t('common.label.expiresAt') }}</th>
              <th>{{ t('marketplace.management.columnUpdatedAt') }}</th>
              <th class="management-table-actions">{{ t('marketplace.management.columnActions') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="ad in items"
              :key="ad.task_id"
            >
              <td>
                <strong>{{ ad.title }}</strong>
                <small>{{ ad.summary || t('marketplace.management.walletAdNoSummary') }}</small>
              </td>
              <td>{{ formatAdMediaType(ad) }}</td>
              <td>{{ formatPoints(ad.reward_points) }}</td>
              <td>
                <span class="management-status-pill">{{ formatAdStatus(ad) }}</span>
              </td>
              <td>{{ t('marketplace.management.walletAdWatchCountValue', { count: ad.watch_count }) }}</td>
              <td>{{ formatOptionalDate(ad.ends_at) }}</td>
              <td>{{ formatDate(ad.updated_at) }}</td>
              <td class="management-table-actions">
                <div class="management-action-group">
                  <button
                    type="button"
                    class="management-list-action"
                    @click="editRewardAd(ad)"
                  >
                    {{ t('marketplace.management.walletAdEditAction') }}
                  </button>
                  <button
                    type="button"
                    class="management-list-action"
                    @click="toggleRewardAd(ad)"
                  >
                    {{ ad.is_active ? t('marketplace.management.walletAdDisableAction') : t('marketplace.management.walletAdEnableAction') }}
                  </button>
                  <button
                    type="button"
                    class="management-list-action"
                    @click="openRenewalDialog(ad)"
                  >
                    {{ t('marketplace.management.renewAction') }}
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <ManagementPagination
        :has-next="hasNext"
        :has-previous="hasPrevious"
        :loading="loading"
        :page="page"
        :page-size="pageSize"
        :t="t"
        :total="total"
        @next="next"
        @previous="previous"
      />
    </article>

    <Teleport to="body">
      <Transition name="management-dialog">
        <div
          v-if="renewalDialogAd"
          class="management-dialog"
          role="dialog"
          aria-modal="true"
          :aria-label="t('marketplace.management.walletAdRenewDialogTitle')"
          @click.self="closeRenewalDialog"
        >
          <section class="management-dialog-panel">
            <header class="management-dialog-header">
              <div>
                <p>{{ t('marketplace.management.renewDialogKicker') }}</p>
                <h2>{{ t('marketplace.management.walletAdRenewDialogTitle') }}</h2>
              </div>
              <button
                type="button"
                class="management-dialog-close"
                :aria-label="t('marketplace.management.renewDialogClose')"
                @click="closeRenewalDialog"
              >
                <AppIcon
                  name="close"
                  :size="18"
                />
              </button>
            </header>

            <div class="management-dialog-body">
              <div class="management-renew-listing">
                <span>{{ t('marketplace.management.walletAdRenewDialogTask') }}</span>
                <strong>{{ renewalDialogAd.title }}</strong>
                <small>{{ renewalDialogAd.task_id }}</small>
              </div>
              <label class="management-dialog-field">
                <span>{{ t('marketplace.management.renewDaysField') }}</span>
                <input
                  v-model.number="renewalDays"
                  type="number"
                  min="1"
                  max="365"
                  step="1"
                  inputmode="numeric"
                />
              </label>
              <div class="management-renew-options">
                <button
                  v-for="dayOption in renewalDayOptions"
                  :key="dayOption"
                  type="button"
                  :class="{ 'is-active': renewalDays === dayOption }"
                  @click="renewalDays = dayOption"
                >
                  {{ t('marketplace.management.renewDaysOption', { days: dayOption }) }}
                </button>
              </div>
              <p
                v-if="!canSubmitRenewal"
                class="management-dialog-error"
              >
                {{ t('marketplace.management.renewDaysInvalid') }}
              </p>
              <p class="management-dialog-hint">
                {{ t('marketplace.management.walletAdRenewDialogHint') }}
              </p>
            </div>

            <footer class="management-dialog-actions">
              <button
                type="button"
                class="management-dialog-button management-dialog-button--secondary"
                @click="closeRenewalDialog"
              >
                {{ t('marketplace.management.cancelAction') }}
              </button>
              <button
                type="button"
                class="management-dialog-button management-dialog-button--primary"
                :disabled="!canSubmitRenewal || renewing"
                @click="submitRenewal"
              >
                {{ renewing ? t('marketplace.management.saving') : t('marketplace.management.renewConfirmAction') }}
              </button>
            </footer>
          </section>
        </div>
      </Transition>
    </Teleport>
  </section>
</template>

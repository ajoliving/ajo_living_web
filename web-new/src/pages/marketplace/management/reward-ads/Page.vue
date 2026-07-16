<!--
 * 廣告列表頁。
 * 1. 以管理列表版型列出 Staff 建立的看廣告得積分任務。
 * 2. 提供搜尋、分頁、啟停與編輯入口。
-->
<script setup lang="ts">
import { computed, ref } from 'vue';
import { useRouter } from 'vue-router';

import { fetchStaffRewardAds, renewStaffRewardAd, updateStaffRewardAd } from '@/httpapis/wallet';
import type { StaffRewardAdResponse, StaffRewardAdType } from '@/model/wallet';
import AppIcon from '@/shared/components/base/AppIcon.vue';
import { usePreferenceStore } from '@/stores/preferences';
import { formatDate } from '@/utils/format';
import { formatAjoPoints } from '@/utils/wallet';

import {
  type StaffManagementListParams,
  type StaffManagementListFetcher,
  useStaffManagementList,
} from '../composables/useStaffManagementList';
import ManagementPagination from '../widgets/ManagementPagination.vue';
import '../styles.scss';

const router = useRouter();
const preferenceStore = usePreferenceStore();
const adTypeFilter = ref<StaffRewardAdType | ''>('');

// 1. 查詢廣告列表
const fetchRewardAdItems: StaffManagementListFetcher<StaffRewardAdResponse> = (
  params: StaffManagementListParams,
) =>
  fetchStaffRewardAds({
    page: params.page,
    page_size: params.page_size,
    keyword: params.keyword,
    is_active: params.status ? params.status === 'active' : undefined,
    ad_type: adTypeFilter.value || undefined,
  });

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

// 2.1 廣告類型篩選選項
const adTypeOptions = computed<Array<{ label: string; value: StaffRewardAdType | '' }>>(() => [
  { label: t('marketplace.management.walletAdTypeAll'), value: '' },
  { label: t('marketplace.management.walletAdTypeReward'), value: 'reward' },
  { label: t('marketplace.management.walletAdTypeDisplayShort'), value: 'display_short' },
  { label: t('marketplace.management.walletAdTypeDisplayLong'), value: 'display_long' },
]);

// 2. 格式化廣告狀態
const formatAdStatus = (ad: StaffRewardAdResponse): string =>
  ad.is_active ? t('common.state.active') : t('common.state.hidden');

// 3. 格式化廣告媒體類型
const formatAdMediaType = (ad: StaffRewardAdResponse): string =>
  ad.media_type === 'video'
    ? t('marketplace.management.walletAdMediaTypeVideo')
    : t('marketplace.management.walletAdMediaTypeImage');

// 4. 格式化廣告用途
const formatAdType = (ad: StaffRewardAdResponse): string => {
  if (ad.ad_type === 'display_short') {
    return t('marketplace.management.walletAdTypeDisplayShort');
  }
  if (ad.ad_type === 'display_long' || ad.ad_type === 'display') {
    return t('marketplace.management.walletAdTypeDisplayLong');
  }
  return t('marketplace.management.walletAdTypeReward');
};

// 5. 格式化積分
const formatPoints = (value: number): string =>
  formatAjoPoints(value, t('common.brand.pointsName'), preferenceStore.locale);

// 6. 跳轉廣告編輯頁
const editRewardAd = async (ad: StaffRewardAdResponse): Promise<void> => {
  await router.push(`/account/marketplace/management/reward-ad-editor/${ad.task_id}`);
};

// 7. 跳轉新增廣告頁
const createRewardAd = async (): Promise<void> => {
  await router.push('/account/marketplace/management/reward-ad-editor');
};

// 8. 啟停廣告任務
const toggleRewardAd = async (ad: StaffRewardAdResponse): Promise<void> => {
  await runAction(() => updateStaffRewardAd(ad.task_id, { is_active: !ad.is_active }));
};

// 9. 開啟續期彈窗
const openRenewalDialog = (ad: StaffRewardAdResponse): void => {
  renewalDialogAd.value = ad;
  renewalDays.value = 14;
};

// 10. 關閉續期彈窗
const closeRenewalDialog = (): void => {
  renewalDialogAd.value = null;
  renewalDays.value = 14;
};

// 11. 提交續期設定
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
  <section class="management-list-page reward-ad-list-page">
    <header class="management-list-header reward-ad-list-header">
      <div class="reward-ad-title-block">
        <p class="management-list-kicker">
          {{ t('common.brand.pointsName') }}
        </p>
        <h1>{{ t('marketplace.management.walletAdListSection') }}</h1>
        <p>{{ t('marketplace.management.walletAdsDescription') }}</p>
      </div>
      <button
        type="button"
        class="management-list-button"
        @click="createRewardAd"
      >
        <AppIcon
          name="plus-square"
          :size="16"
        />
        {{ t('marketplace.management.walletAdNewAction') }}
      </button>
    </header>

    <article class="management-list-panel">
      <div class="management-list-toolbar reward-ad-toolbar">
        <label class="management-list-search reward-ad-search">
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
        <div class="reward-ad-filter-actions">
          <label class="reward-ad-filter-field">
            <span>{{ t('marketplace.management.walletAdTypeField') }}</span>
            <select
              v-model="adTypeFilter"
              class="management-list-select"
              @change="search"
            >
              <option
                v-for="option in adTypeOptions"
                :key="option.value || 'all'"
                :value="option.value"
              >
                {{ option.label }}
              </option>
            </select>
          </label>
          <label class="reward-ad-filter-field">
            <span>{{ t('common.label.status') }}</span>
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
          </label>
          <button
            type="button"
            class="management-list-button reward-ad-search-button"
            @click="search"
          >
            {{ t('marketplace.list.searchAction') }}
          </button>
        </div>
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
              <th>{{ t('marketplace.management.walletAdTypeField') }}</th>
              <th>{{ t('marketplace.management.walletAdMediaTypeField') }}</th>
              <th>{{ t('marketplace.management.walletAdRewardField') }}</th>
              <th>{{ t('common.label.status') }}</th>
              <th>{{ t('marketplace.management.walletAdWatchCountColumn') }}</th>
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
              <td>{{ formatAdType(ad) }}</td>
              <td>{{ formatAdMediaType(ad) }}</td>
              <td>{{ ad.ad_type === 'reward' ? formatPoints(ad.reward_points) : '-' }}</td>
              <td>
                <span class="management-status-pill">{{ formatAdStatus(ad) }}</span>
              </td>
              <td>{{ t('marketplace.management.walletAdWatchCountValue', { count: ad.watch_count }) }}</td>
              <td>{{ formatDate(ad.updated_at, preferenceStore.locale) }}</td>
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

<style scoped lang="scss">
.reward-ad-list-header {
  align-items: stretch;
  border: 1px solid rgb(var(--color-border));
  border-radius: 2px;
  background: rgb(var(--color-surface-raised));
  padding: 1rem;
}

.reward-ad-title-block {
  min-width: 0;
  max-width: 42rem;
}

.reward-ad-list-header .management-list-button {
  align-self: flex-start;
  min-width: 7.5rem;
}

.reward-ad-toolbar {
  grid-template-columns: minmax(18rem, 1fr) auto;
  align-items: end;
  background: rgb(var(--color-surface));
}

.reward-ad-filter-actions {
  display: grid;
  grid-template-columns: repeat(2, minmax(9rem, 12rem)) auto;
  align-items: end;
  gap: 0.75rem;
}

.reward-ad-filter-field {
  display: grid;
  min-width: 0;
  gap: 0.35rem;
}

.reward-ad-filter-field span {
  color: rgb(var(--color-text-muted));
  font-size: 0.72rem;
  font-weight: 800;
  letter-spacing: 0.06em;
  line-height: 1;
  text-transform: uppercase;
}

.reward-ad-filter-field .management-list-select {
  width: 100%;
}

.reward-ad-search,
.reward-ad-search-button {
  align-self: end;
}

@media (max-width: 1023px) {
  .reward-ad-toolbar {
    grid-template-columns: 1fr;
  }

  .reward-ad-filter-actions {
    grid-template-columns: repeat(2, minmax(0, 1fr)) auto;
  }
}

@media (max-width: 767px) {
  .reward-ad-list-header {
    align-items: stretch;
    flex-direction: column;
  }

  .reward-ad-list-header .management-list-button,
  .reward-ad-search-button {
    width: 100%;
  }

  .reward-ad-filter-actions {
    grid-template-columns: 1fr;
  }
}
</style>

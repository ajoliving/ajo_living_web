<!--
 * 廣告列表頁。
 * 1. 以管理列表版型列出 Staff 建立的看廣告得積分任務。
 * 2. 提供搜尋、分頁、啟停與編輯入口。
-->
<script setup lang="ts">
import { computed, ref } from 'vue';

import { fetchStaffRewardAds, renewStaffRewardAd, updateStaffRewardAd } from '@/domains/payments/api';
import type { StaffRewardAdResponse } from '@/domains/payments/model';
import AppUnsavedChangesDialog from '@/shared/components/base/AppUnsavedChangesDialog.vue';
import AppIcon from '@/shared/components/base/AppIcon.vue';
import { formatDate } from '@/shared/utils/format';
import { formatAjoPoints } from '@/shared/utils/wallet';

import {
  type StaffManagementListParams,
  type StaffManagementListFetcher,
  useStaffManagementList,
} from '../composables/useStaffManagementList';
import ManagementPagination from '../widgets/ManagementPagination.vue';
import RewardAdEditorPanel from '../wallet/widgets/RewardAdEditorPanel.vue';
import { useRewardAdEditorPage } from '../wallet/wallet';
import '../styles.scss';
import '../wallet/styles.scss';

const props = withDefaults(defineProps<{
  adType?: 'display' | 'reward';
}>(), {
  adType: 'reward',
});

// 1. 查詢廣告列表
const fetchRewardAdItems: StaffManagementListFetcher<StaffRewardAdResponse> = (
  params: StaffManagementListParams,
) =>
  fetchStaffRewardAds({
    page: params.page,
    page_size: params.page_size,
    keyword: params.keyword,
    is_active: params.status ? params.status === 'active' : undefined,
    ad_type: props.adType,
  });

const {
  hasNext,
  hasPrevious,
  items,
  keyword,
  loadItems,
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
const publishDialogOpen = ref(false);
const publishCloseConfirmOpen = ref(false);
const editDialogOpen = ref(false);
const editCloseConfirmOpen = ref(false);
const {
  adDetail,
  adForm,
  adFormErrors,
  isEditingAd,
  loadRewardAdByTaskId,
  loadingAd,
  resetRewardAdForm,
  saveRewardAd,
  savingAd,
  stageRewardAdMedia,
  submitted,
  uploadingMedia,
} = useRewardAdEditorPage({
  fixedAdType: props.adType,
  afterSave: async () => {
    publishDialogOpen.value = false;
    editDialogOpen.value = false;
    await loadItems(1);
  },
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
const canClosePublishDialog = computed(() => !savingAd.value && !uploadingMedia.value);
const initialAdForm = computed(() => ({
  displayLayout: props.adType === 'display' ? 'image_text' : 'image_text',
  rewardPoints: props.adType === 'display' ? 0 : 50,
  watchSeconds: props.adType === 'display' ? 1 : 30,
}));
const hasPublishDialogDraft = computed(() =>
  adForm.adType !== props.adType ||
  adForm.title.trim().length > 0 ||
  adForm.summary.trim().length > 0 ||
  adForm.mediaURL.trim().length > 0 ||
  Boolean(adForm.mediaFile) ||
  adForm.targetURL.trim().length > 0 ||
  adForm.displayChannel.length > 0 ||
  adForm.displayLayout !== initialAdForm.value.displayLayout ||
  adForm.sortOrder !== 1 ||
  adForm.rewardPoints !== initialAdForm.value.rewardPoints ||
  adForm.watchSeconds !== initialAdForm.value.watchSeconds ||
  adForm.retentionDays !== 30 ||
  adForm.isActive !== true,
);
const hasEditDialogDraft = computed(() => editDialogOpen.value && isEditingAd.value);
const pageTitle = computed(() =>
  props.adType === 'display'
    ? t('marketplace.management.walletDisplayAdListSection')
    : t('marketplace.management.walletRewardAdListSection'),
);
const pageDescription = computed(() =>
  props.adType === 'display'
    ? t('marketplace.management.walletDisplayAdsDescription')
    : t('marketplace.management.walletRewardAdsDescription'),
);
const publishTitle = computed(() =>
  props.adType === 'display'
    ? t('marketplace.management.walletDisplayAdPublishSection')
    : t('marketplace.management.walletRewardAdPublishSection'),
);
const editTitle = computed(() =>
  props.adType === 'display'
    ? t('marketplace.management.walletDisplayAdEditSection')
    : t('marketplace.management.walletRewardAdEditSection'),
);
const valueColumnTitle = computed(() =>
  props.adType === 'display'
    ? t('marketplace.management.walletAdDisplayLayoutField')
    : t('marketplace.management.walletAdRewardField'),
);

// 2. 格式化廣告狀態
const formatAdStatus = (ad: StaffRewardAdResponse): string =>
  ad.is_active ? t('common.state.active') : t('common.state.hidden');

// 3. 格式化廣告媒體類型
const formatAdMediaType = (ad: StaffRewardAdResponse): string =>
  ad.media_type === 'video'
    ? t('marketplace.management.walletAdMediaTypeVideo')
    : t('marketplace.management.walletAdMediaTypeImage');

// 4. 格式化積分或展示樣式
const formatPoints = (ad: StaffRewardAdResponse): string =>
  ad.ad_type === 'display'
    ? t(`marketplace.settings.displayAdsLayout.${ad.display_layout}`)
    : formatAjoPoints(ad.reward_points, t('common.brand.pointsName'), 'zh-HK');

// 5. 格式化可選日期
const formatOptionalDate = (value?: string | null): string =>
  value ? formatDate(value) : '-';

// 6. 開啟發布廣告彈窗
const openPublishDialog = (): void => {
  resetRewardAdForm();
  publishDialogOpen.value = true;
};

// 7. 關閉發布廣告彈窗
const closePublishDialog = (shouldConfirm = false): void => {
  if (!canClosePublishDialog.value) {
    return;
  }
  if (shouldConfirm && hasPublishDialogDraft.value) {
    publishCloseConfirmOpen.value = true;
    return;
  }

  publishCloseConfirmOpen.value = false;
  publishDialogOpen.value = false;
  resetRewardAdForm();
};

// 8. 放棄發布廣告草稿
const discardPublishDialogDraft = (): void => {
  publishCloseConfirmOpen.value = false;
  publishDialogOpen.value = false;
  resetRewardAdForm();
};

// 9. 保存發布廣告草稿
const savePublishDialogDraft = async (): Promise<void> => {
  await saveRewardAd();
  publishCloseConfirmOpen.value = false;
};

// 10. 開啟廣告編輯彈窗
const editRewardAd = async (ad: StaffRewardAdResponse): Promise<void> => {
  await loadRewardAdByTaskId(ad.task_id);
  editDialogOpen.value = true;
};

// 11. 啟停廣告任務
const toggleRewardAd = async (ad: StaffRewardAdResponse): Promise<void> => {
  await runAction(() => updateStaffRewardAd(ad.task_id, { is_active: !ad.is_active }));
};

// 12. 開啟續期彈窗
const openRenewalDialog = (ad: StaffRewardAdResponse): void => {
  renewalDialogAd.value = ad;
  renewalDays.value = 14;
};

// 13. 關閉續期彈窗
const closeRenewalDialog = (): void => {
  renewalDialogAd.value = null;
  renewalDays.value = 14;
};

// 14. 提交續期設定
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

// 15. 關閉編輯廣告彈窗
const closeEditDialog = (shouldConfirm = false): void => {
  if (!canClosePublishDialog.value) {
    return;
  }
  if (shouldConfirm && hasEditDialogDraft.value) {
    editCloseConfirmOpen.value = true;
    return;
  }

  editCloseConfirmOpen.value = false;
  editDialogOpen.value = false;
  resetRewardAdForm();
};

// 16. 放棄編輯廣告草稿
const discardEditDialogDraft = (): void => {
  editCloseConfirmOpen.value = false;
  editDialogOpen.value = false;
  resetRewardAdForm();
};

// 17. 保存編輯廣告草稿
const saveEditDialogDraft = async (): Promise<void> => {
  await saveRewardAd();
  editCloseConfirmOpen.value = false;
};
</script>

<template>
  <section class="management-list-page">
    <header class="management-list-header">
      <div>
        <p class="management-list-kicker">
          {{ t('common.brand.pointsName') }}
        </p>
        <h1>{{ pageTitle }}</h1>
        <p>{{ pageDescription }}</p>
      </div>
      <div class="management-list-header-actions">
        <button
          type="button"
          class="management-list-button"
          @click="openPublishDialog"
        >
          <AppIcon
            name="plus-square"
            :size="16"
          />
          <span>{{ publishTitle }}</span>
        </button>
      </div>
    </header>

    <article class="management-list-panel">
      <div class="management-list-toolbar management-list-toolbar--reward-ads">
        <div class="management-list-filter-group">
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
        </div>
        <div class="management-list-toolbar-actions">
          <button
            type="button"
            class="management-list-button management-list-button--secondary"
            @click="search"
          >
            <AppIcon
              name="search"
              :size="16"
            />
            <span>{{ t('marketplace.list.searchAction') }}</span>
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
              <th>{{ t('marketplace.management.walletAdMediaTypeField') }}</th>
              <th>{{ valueColumnTitle }}</th>
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
              <td>{{ formatPoints(ad) }}</td>
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
          v-if="publishDialogOpen"
          class="management-dialog"
          role="dialog"
          aria-modal="true"
          :aria-label="publishTitle"
          @click.self="closePublishDialog(true)"
        >
          <section class="management-dialog-panel management-dialog-panel--editor">
            <header class="management-dialog-header">
              <div>
                <p>{{ t('common.brand.pointsName') }}</p>
                <h2>{{ publishTitle }}</h2>
              </div>
              <button
                type="button"
                class="management-dialog-close"
                :disabled="!canClosePublishDialog"
                :aria-label="t('marketplace.management.renewDialogClose')"
                @click="closePublishDialog(true)"
              >
                <AppIcon
                  name="close"
                  :size="18"
                />
              </button>
            </header>

            <div class="management-dialog-body management-dialog-body--editor">
              <div
                v-if="loadingAd"
                class="wallet-settings-empty"
              >
                {{ t('common.status.loading') }}
              </div>

              <RewardAdEditorPanel
                v-else
                :ad-detail="adDetail"
                :ad-form="adForm"
                :errors="adFormErrors"
                hide-header
                :is-editing-ad="isEditingAd"
                lock-ad-type
                :saving-ad="savingAd || uploadingMedia"
                :submitted="submitted"
                :t="t"
                @reset="resetRewardAdForm"
                @save="saveRewardAd"
                @stage-media="stageRewardAdMedia"
              />
            </div>
          </section>
        </div>
      </Transition>
    </Teleport>

    <AppUnsavedChangesDialog
      :open="publishCloseConfirmOpen"
      :title="t('marketplace.management.walletAdCloseConfirmTitle')"
      :description="t('marketplace.management.walletAdCloseConfirmDescription')"
      :save-label="t('marketplace.management.walletAdCloseConfirmSave')"
      :discard-label="t('marketplace.management.walletAdCloseConfirmDiscard')"
      :stay-label="t('marketplace.management.walletAdCloseConfirmStay')"
      :saving="savingAd || uploadingMedia"
      @save="savePublishDialogDraft"
      @discard="discardPublishDialogDraft"
      @stay="publishCloseConfirmOpen = false"
    />

    <Teleport to="body">
      <Transition name="management-dialog">
        <div
          v-if="editDialogOpen"
          class="management-dialog"
          role="dialog"
          aria-modal="true"
          :aria-label="editTitle"
          @click.self="closeEditDialog(true)"
        >
          <section class="management-dialog-panel management-dialog-panel--editor">
            <header class="management-dialog-header">
              <div>
                <p>{{ t('common.brand.pointsName') }}</p>
                <h2>{{ editTitle }}</h2>
              </div>
              <button
                type="button"
                class="management-dialog-close"
                :disabled="!canClosePublishDialog"
                :aria-label="t('marketplace.management.renewDialogClose')"
                @click="closeEditDialog(true)"
              >
                <AppIcon
                  name="close"
                  :size="18"
                />
              </button>
            </header>

            <div class="management-dialog-body management-dialog-body--editor">
              <div
                v-if="loadingAd"
                class="wallet-settings-empty"
              >
                {{ t('common.status.loading') }}
              </div>

              <RewardAdEditorPanel
                v-else
                :ad-detail="adDetail"
                :ad-form="adForm"
                :errors="adFormErrors"
                hide-header
                :is-editing-ad="isEditingAd"
                lock-ad-type
                :saving-ad="savingAd || uploadingMedia"
                :submitted="submitted"
                :t="t"
                @reset="resetRewardAdForm"
                @save="saveRewardAd"
                @stage-media="stageRewardAdMedia"
              />
            </div>
          </section>
        </div>
      </Transition>
    </Teleport>

    <AppUnsavedChangesDialog
      :open="editCloseConfirmOpen"
      :title="t('marketplace.management.walletAdEditCloseConfirmTitle')"
      :description="t('marketplace.management.walletAdEditCloseConfirmDescription')"
      :save-label="savingAd || uploadingMedia ? t('marketplace.management.saving') : t('marketplace.management.walletAdUpdateAction')"
      :discard-label="t('marketplace.management.walletAdCloseConfirmDiscard')"
      :stay-label="t('marketplace.management.walletAdCloseConfirmStay')"
      :saving="savingAd || uploadingMedia"
      @save="saveEditDialogDraft"
      @discard="discardEditDialogDraft"
      @stay="editCloseConfirmOpen = false"
    />

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

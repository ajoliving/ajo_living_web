<!--
 * 廣告任務編輯面板。
 * 1. 管理看廣告得積分與列表展示廣告表單。
 * 2. 觸發建立、更新與重置操作。
-->
<script setup lang="ts">
import AppIcon from '@/shared/components/base/AppIcon.vue';
import { formatDate } from '@/shared/utils/format';
import { formatAjoPoints } from '@/shared/utils/wallet';

import type { StaffRewardAdResponse } from '@/domains/payments/model';
import type { RewardAdForm, RewardAdFormErrors } from '../wallet';

const displayLayoutOptions = [
  {
    value: 'text_compact',
    labelKey: 'marketplace.management.walletAdLayoutTextCompact',
    specKey: 'marketplace.management.walletAdLayoutTextCompactSpec',
  },
  {
    value: 'image_text',
    labelKey: 'marketplace.management.walletAdLayoutImageText',
    specKey: 'marketplace.management.walletAdLayoutImageTextSpec',
  },
  {
    value: 'image_full',
    labelKey: 'marketplace.management.walletAdLayoutImageFull',
    specKey: 'marketplace.management.walletAdLayoutImageFullSpec',
  },
] as const;

withDefaults(defineProps<{
  adDetail?: StaffRewardAdResponse | null;
  adForm: RewardAdForm;
  errors?: RewardAdFormErrors;
  hideHeader?: boolean;
  isEditingAd: boolean;
  lockAdType?: boolean;
  savingAd: boolean;
  submitted?: boolean;
  t: (key: string, params?: Record<string, unknown>) => string;
}>(), {
  errors: () => ({}),
  hideHeader: false,
  lockAdType: false,
  submitted: false,
});

const emit = defineEmits<{
  reset: [];
  save: [];
  stageMedia: [file: File];
}>();

// 1. 處理廣告媒體檔案選擇
const handleMediaFileChange = (event: Event): void => {
  const input = event.target as HTMLInputElement;
  const file = input.files?.[0];
  input.value = '';
  if (file) {
    emit('stageMedia', file);
  }
};

// 2. 格式化可選日期
const formatOptionalDate = (value?: string): string =>
  value ? formatDate(value) : '-';

// 3. 判斷欄位錯誤是否需要顯示
const shouldShowError = (message?: string): boolean =>
  Boolean(message);
</script>

<template>
  <article class="wallet-settings-panel">
    <div
      v-if="!hideHeader"
      class="wallet-settings-panel-header"
    >
      <div>
        <h2>{{ t('marketplace.management.walletAdEditorTitle') }}</h2>
        <p>{{ t('marketplace.management.walletAdEditorDescription') }}</p>
      </div>
    </div>

    <div class="wallet-settings-form">
      <div
        v-if="adDetail"
        class="wallet-settings-metrics"
      >
        <div>
          <span>{{ t('marketplace.management.walletAdMetricTaskId') }}</span>
          <strong>{{ adDetail.task_id }}</strong>
        </div>
        <div>
          <span>{{ t('common.label.status') }}</span>
          <strong>{{ adDetail.is_active ? t('common.state.active') : t('common.state.hidden') }}</strong>
        </div>
        <div>
          <span>{{ t('common.label.expiresAt') }}</span>
          <strong>{{ formatOptionalDate(adDetail.ends_at) }}</strong>
        </div>
        <div>
          <span>{{ t('marketplace.management.walletAdWatchCountColumn') }}</span>
          <strong>{{ t('marketplace.management.walletAdWatchCountValue', { count: adDetail.watch_count }) }}</strong>
        </div>
        <div>
          <span>{{ t('marketplace.management.walletAdWatchSecondsField') }}</span>
          <strong>{{ t('marketplace.management.walletAdWatchTimeValue', { seconds: adDetail.total_watch_seconds }) }}</strong>
        </div>
        <div>
          <span>{{ t('marketplace.management.walletAdClickCountField') }}</span>
          <strong>{{ adDetail.link_click_count }}</strong>
        </div>
        <div>
          <span>{{ t('marketplace.management.walletAdClickRateField') }}</span>
          <strong>{{ t('marketplace.management.walletAdClickRateValue', { rate: adDetail.link_click_rate.toFixed(1) }) }}</strong>
        </div>
        <div>
          <span>{{ t('marketplace.management.walletAdRewardField') }}</span>
          <strong>
            {{ adDetail.ad_type === 'reward' ? formatAjoPoints(adDetail.reward_points, t('common.brand.pointsName'), 'zh-HK') : '-' }}
          </strong>
        </div>
        <div>
          <span>{{ t('marketplace.management.walletAdCreatedAtField') }}</span>
          <strong>{{ formatDate(adDetail.created_at) }}</strong>
        </div>
        <div>
          <span>{{ t('marketplace.management.columnUpdatedAt') }}</span>
          <strong>{{ formatDate(adDetail.updated_at) }}</strong>
        </div>
      </div>

      <div class="wallet-settings-two-column wallet-settings-ad-title-row">
        <label class="wallet-settings-field">
          <span>{{ t('marketplace.management.walletAdTypeField') }}</span>
          <select
            v-model="adForm.adType"
            :disabled="lockAdType"
          >
            <option value="reward">
              {{ t('marketplace.management.walletAdTypeReward') }}
            </option>
            <option value="display">
              {{ t('marketplace.management.walletAdTypeDisplay') }}
            </option>
          </select>
        </label>

        <label
          class="wallet-settings-field"
          :class="{ 'wallet-settings-field--invalid': submitted && shouldShowError(errors.title) }"
        >
          <span>
            {{ t('marketplace.management.walletAdTitleField') }}
            <b>*</b>
          </span>
          <input
            v-model="adForm.title"
            type="text"
            maxlength="160"
            :placeholder="t('marketplace.management.walletAdTitlePlaceholder')"
          />
          <small v-if="submitted && errors.title">{{ errors.title }}</small>
        </label>
      </div>

      <div
        v-if="adForm.adType === 'display'"
        class="wallet-settings-two-column wallet-settings-ad-controls"
      >
        <label
          class="wallet-settings-field"
          :class="{ 'wallet-settings-field--invalid': submitted && shouldShowError(errors.displayLayout) }"
        >
          <span>
            {{ t('marketplace.management.walletAdDisplayLayoutField') }}
            <b>*</b>
          </span>
          <select v-model="adForm.displayLayout">
            <option
              v-for="option in displayLayoutOptions"
              :key="option.value"
              :value="option.value"
            >
              {{ t(option.labelKey) }}
            </option>
          </select>
          <small v-if="submitted && errors.displayLayout">{{ errors.displayLayout }}</small>
        </label>
      </div>

      <div
        v-if="adForm.adType === 'display'"
        class="wallet-settings-layout-tags"
      >
        <button
          v-for="option in displayLayoutOptions"
          :key="option.value"
          type="button"
          :class="{ 'is-active': adForm.displayLayout === option.value }"
          @click="adForm.displayLayout = option.value"
        >
          <strong>{{ t(option.labelKey) }}</strong>
          <span v-if="adForm.mediaType === 'image'">{{ t(option.specKey) }}</span>
        </button>
      </div>

      <label
        v-if="adForm.adType === 'reward' || adForm.displayLayout === 'image_text'"
        class="wallet-settings-field"
        :class="{ 'wallet-settings-field--invalid': submitted && shouldShowError(errors.summary) }"
      >
        <span>
          {{ t('marketplace.management.walletAdSummaryField') }}
          <b v-if="adForm.adType === 'reward'">*</b>
        </span>
        <textarea
          v-model="adForm.summary"
          rows="3"
          maxlength="500"
          :placeholder="t('marketplace.management.walletAdSummaryPlaceholder')"
        />
        <small v-if="submitted && errors.summary">{{ errors.summary }}</small>
      </label>

      <div class="wallet-settings-two-column wallet-settings-media-row">
        <label class="wallet-settings-field">
          <span>
            {{ t('marketplace.management.walletAdMediaTypeField') }}
            <b>*</b>
          </span>
          <select v-model="adForm.mediaType">
            <option value="image">
              {{ t('marketplace.management.walletAdMediaTypeImage') }}
            </option>
            <option value="video">
              {{ t('marketplace.management.walletAdMediaTypeVideo') }}
            </option>
          </select>
        </label>

        <label
          class="wallet-settings-field"
          :class="{ 'wallet-settings-field--invalid': submitted && shouldShowError(errors.media) }"
        >
          <span>
            {{ t('marketplace.management.walletAdMediaUploadField') }}
            <b>*</b>
          </span>
          <span class="wallet-settings-upload-control">
            <span class="wallet-settings-upload-status">
              <AppIcon
                name="cloud-upload"
                :size="16"
              />
              <span>
                {{ adForm.mediaFile?.name || (adForm.mediaURL || adForm.mediaPreviewURL ? t('marketplace.management.walletAdMediaSelected') : t('marketplace.management.walletAdMediaEmpty')) }}
              </span>
            </span>
            <span class="wallet-settings-upload-action">
              {{ t('marketplace.management.walletAdMediaChooseAction') }}
            </span>
            <input
              class="wallet-settings-upload-input"
              type="file"
              :accept="adForm.mediaType === 'video' ? 'video/*' : 'image/*'"
              @change="handleMediaFileChange"
            />
          </span>
          <small v-if="submitted && errors.media">{{ errors.media }}</small>
        </label>
      </div>

      <p
        v-if="adForm.adType === 'display' && adForm.mediaType === 'image'"
        class="wallet-settings-media-spec"
      >
        {{ t('marketplace.management.walletAdRecommendedImageSpec') }}
        {{ t(displayLayoutOptions.find((option) => option.value === adForm.displayLayout)?.specKey ?? 'marketplace.management.walletAdLayoutImageTextSpec') }}
      </p>

      <div
        v-if="adForm.mediaPreviewURL || adForm.mediaURL"
        class="wallet-settings-media-preview"
      >
        <img
          v-if="adForm.mediaType === 'image'"
          :src="adForm.mediaPreviewURL || adForm.mediaURL"
          :alt="adForm.title"
        />
        <video
          v-else
          :src="adForm.mediaPreviewURL || adForm.mediaURL"
          controls
          preload="metadata"
        />
      </div>

      <div class="wallet-settings-three-column wallet-settings-ad-controls">
        <label
          v-if="adForm.adType === 'reward'"
          class="wallet-settings-field"
          :class="{ 'wallet-settings-field--invalid': submitted && shouldShowError(errors.rewardPoints) }"
        >
          <span>
            {{ t('marketplace.management.walletAdRewardField') }}
            <b>*</b>
          </span>
          <input
            v-model.number="adForm.rewardPoints"
            type="number"
            min="1"
            step="1"
          />
          <small v-if="submitted && errors.rewardPoints">{{ errors.rewardPoints }}</small>
        </label>

        <label
          v-if="adForm.adType === 'reward'"
          class="wallet-settings-field"
          :class="{ 'wallet-settings-field--invalid': submitted && shouldShowError(errors.watchSeconds) }"
        >
          <span>
            {{ t('marketplace.management.walletAdWatchSecondsField') }}
            <b>*</b>
          </span>
          <input
            v-model.number="adForm.watchSeconds"
            type="number"
            min="1"
            step="1"
          />
          <small v-if="submitted && errors.watchSeconds">{{ errors.watchSeconds }}</small>
        </label>

        <label
          class="wallet-settings-field"
          :class="{ 'wallet-settings-field--invalid': submitted && shouldShowError(errors.retentionDays) }"
        >
          <span>
            {{ isEditingAd ? t('marketplace.management.walletAdRenewDaysField') : t('marketplace.management.walletAdRetentionDaysField') }}
            <b>*</b>
          </span>
          <input
            v-model.number="adForm.retentionDays"
            type="number"
            min="1"
            max="365"
            step="1"
          />
          <small v-if="submitted && errors.retentionDays">{{ errors.retentionDays }}</small>
        </label>

        <label class="wallet-settings-field wallet-settings-switch-field">
          <span>{{ t('marketplace.management.walletAdActiveField') }}</span>
          <span class="wallet-settings-check">
            <input
              v-model="adForm.isActive"
              type="checkbox"
            />
          </span>
        </label>
      </div>

      <label class="wallet-settings-field">
        <span>{{ t('marketplace.management.walletAdTargetField') }}</span>
        <input
          v-model="adForm.targetURL"
          type="url"
          :placeholder="t('marketplace.management.walletAdTargetPlaceholder')"
        />
      </label>

      <div class="wallet-settings-actions">
        <button
          v-if="isEditingAd"
          type="button"
          class="wallet-settings-secondary-button"
          @click="emit('reset')"
        >
          {{ t('marketplace.management.walletAdNewAction') }}
        </button>
        <button
          type="button"
          class="wallet-settings-primary-button"
          :disabled="savingAd"
          @click="emit('save')"
        >
          <AppIcon
            name="check-circle"
            :size="16"
          />
          <span>
            {{ savingAd ? t('marketplace.management.saving') : isEditingAd ? t('marketplace.management.walletAdUpdateAction') : t('marketplace.management.walletAdCreateAction') }}
          </span>
        </button>
      </div>
    </div>
  </article>
</template>

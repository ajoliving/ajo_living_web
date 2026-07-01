<!--
 * 廣告任務編輯面板。
 * 1. 管理單個看廣告得積分任務表單。
 * 2. 觸發建立、更新與重置操作。
-->
<script setup lang="ts">
import AppIcon from '@/shared/components/base/AppIcon.vue';

import {
  displayLayoutForAdType,
  isDisplayRewardAdType,
  type RewardAdDisplayChannel,
  type RewardAdForm,
  type RewardAdType,
} from '../wallet';

const adTypeOptions: Array<{ labelKey: string; value: RewardAdType }> = [
  { labelKey: 'marketplace.management.walletAdTypeReward', value: 'reward' },
  { labelKey: 'marketplace.management.walletAdTypeDisplayShort', value: 'display_short' },
  { labelKey: 'marketplace.management.walletAdTypeDisplayLong', value: 'display_long' },
];

const displayChannelOptions: Array<{ labelKey: string; value: Exclude<RewardAdDisplayChannel, ''> }> = [
  { labelKey: 'marketplace.settings.displayAdChannelPropertySale', value: 'property_sale' },
  { labelKey: 'marketplace.settings.displayAdChannelFurniture', value: 'furniture' },
  { labelKey: 'marketplace.settings.displayAdChannelServicedApartment', value: 'serviced_apartment' },
];

defineProps<{
  adForm: RewardAdForm;
  isEditingAd: boolean;
  savingAd: boolean;
  t: (key: string, params?: Record<string, unknown>) => string;
}>();

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

// 2. 輸出展示廣告尺寸文字
const resolveDisplayAdSizeLabel = (adType: RewardAdType, t: (key: string) => string): string =>
  adType === 'display_long'
    ? t('marketplace.management.walletAdDisplayLongSize')
    : t('marketplace.management.walletAdDisplayShortSize');
</script>

<template>
  <article class="wallet-settings-panel">
    <div class="wallet-settings-panel-header">
      <div>
        <h2>{{ t('marketplace.management.walletAdEditorTitle') }}</h2>
        <p>{{ t('marketplace.management.walletAdEditorDescription') }}</p>
      </div>
    </div>

    <div class="wallet-settings-form">
      <div class="wallet-settings-two-column">
        <label class="wallet-settings-field">
          <span>{{ t('marketplace.management.walletAdTypeField') }}</span>
          <select v-model="adForm.adType">
            <option
              v-for="option in adTypeOptions"
              :key="option.value"
              :value="option.value"
            >
              {{ t(option.labelKey) }}
            </option>
          </select>
        </label>

        <label class="wallet-settings-field">
          <span>{{ t('marketplace.management.walletAdTitleField') }}</span>
          <input
            v-model="adForm.title"
            type="text"
            maxlength="160"
            :placeholder="t('marketplace.management.walletAdTitlePlaceholder')"
          />
        </label>

        <label class="wallet-settings-field">
          <span>{{ t('marketplace.management.walletAdRewardField') }}</span>
          <input
            v-if="adForm.adType === 'reward'"
            v-model.number="adForm.rewardPoints"
            type="number"
            min="1"
            step="1"
          />
          <input
            v-else
            :value="resolveDisplayAdSizeLabel(adForm.adType, t)"
            type="text"
            readonly
          />
        </label>
      </div>

      <label class="wallet-settings-field">
        <span>{{ isDisplayRewardAdType(adForm.adType) ? t('marketplace.settings.displayAdTextField') : t('marketplace.management.walletAdSummaryField') }}</span>
        <textarea
          v-model="adForm.summary"
          rows="3"
          maxlength="500"
          :placeholder="t('marketplace.management.walletAdSummaryPlaceholder')"
        />
      </label>

      <div class="wallet-settings-two-column">
        <label class="wallet-settings-field">
          <span>{{ t('marketplace.management.walletAdMediaTypeField') }}</span>
          <select
            v-model="adForm.mediaType"
            :disabled="isDisplayRewardAdType(adForm.adType)"
          >
            <option value="image">
              {{ t('marketplace.management.walletAdMediaTypeImage') }}
            </option>
            <option value="video">
              {{ t('marketplace.management.walletAdMediaTypeVideo') }}
            </option>
          </select>
        </label>

        <label class="wallet-settings-field">
          <span>{{ t('marketplace.management.walletAdMediaUploadField') }}</span>
          <input
            type="file"
            :accept="adForm.mediaType === 'video' ? 'video/*' : 'image/*'"
            @change="handleMediaFileChange"
          />
        </label>
      </div>

      <div
        v-if="isDisplayRewardAdType(adForm.adType)"
        class="wallet-settings-three-column"
      >
        <label class="wallet-settings-field">
          <span>{{ t('marketplace.management.walletAdDisplayChannelField') }}</span>
          <select v-model="adForm.displayChannel">
            <option
              v-for="option in displayChannelOptions"
              :key="option.value"
              :value="option.value"
            >
              {{ t(option.labelKey) }}
            </option>
          </select>
        </label>

        <label class="wallet-settings-field">
          <span>{{ t('marketplace.management.walletAdDisplayLayoutField') }}</span>
          <input
            :value="t(`marketplace.management.walletAdDisplayLayout.${displayLayoutForAdType(adForm.adType)}`)"
            type="text"
            readonly
          />
        </label>

        <label class="wallet-settings-field">
          <span>{{ t('marketplace.management.walletAdSortOrderField') }}</span>
          <input
            v-model.number="adForm.sortOrder"
            type="number"
            min="1"
            step="1"
          />
        </label>
      </div>

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

      <div
        v-if="adForm.adType === 'reward'"
        class="wallet-settings-three-column wallet-settings-ad-controls"
      >
        <label class="wallet-settings-field">
          <span>{{ t('marketplace.management.walletAdWatchSecondsField') }}</span>
          <input
            v-model.number="adForm.watchSeconds"
            type="number"
            min="1"
            step="1"
          />
        </label>

        <label class="wallet-settings-field">
          <span>{{ isEditingAd ? t('marketplace.management.walletAdRenewDaysField') : t('marketplace.management.walletAdRetentionDaysField') }}</span>
          <input
            v-model.number="adForm.retentionDays"
            type="number"
            min="1"
            max="365"
            step="1"
          />
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

      <div
        v-else
        class="wallet-settings-two-column wallet-settings-ad-controls"
      >
        <label class="wallet-settings-field">
          <span>{{ isEditingAd ? t('marketplace.management.walletAdRenewDaysField') : t('marketplace.management.walletAdRetentionDaysField') }}</span>
          <input
            v-model.number="adForm.retentionDays"
            type="number"
            min="1"
            max="365"
            step="1"
          />
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

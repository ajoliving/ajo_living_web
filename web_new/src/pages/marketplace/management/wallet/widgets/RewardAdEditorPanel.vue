<!--
 * 廣告任務編輯面板。
 * 1. 管理單個看廣告得積分任務表單。
 * 2. 觸發建立、更新與重置操作。
-->
<script setup lang="ts">
import { ref } from 'vue';

import AppIcon from '@/shared/components/base/AppIcon.vue';

import {
  displayLayoutForAdType,
  isDisplayRewardAdType,
  isImageOnlyAdType,
  isPriceBannerAdType,
  type RewardAdDisplayChannel,
  type RewardAdForm,
  type RewardAdType,
} from '../wallet';

const adTypeOptions: Array<{ labelKey: string; value: RewardAdType }> = [
  { labelKey: 'marketplace.management.walletAdTypeReward', value: 'reward' },
  { labelKey: 'marketplace.management.walletAdTypeDisplayShort', value: 'display_short' },
  { labelKey: 'marketplace.management.walletAdTypeDisplayLong', value: 'display_long' },
  { labelKey: 'marketplace.management.walletAdTypePriceDesktop', value: 'price_desktop' },
  { labelKey: 'marketplace.management.walletAdTypePriceMobile', value: 'price_mobile' },
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
  close: [];
  reset: [];
  save: [];
  stageMedia: [file: File];
}>();

const selectedMediaName = ref('');

// 1. 處理廣告媒體檔案選擇
const handleMediaFileChange = (event: Event): void => {
  const input = event.target as HTMLInputElement;
  const file = input.files?.[0];
  input.value = '';
  if (file) {
    selectedMediaName.value = file.name;
    emit('stageMedia', file);
  }
};

</script>

<template>
  <div
    class="management-dialog"
    role="dialog"
    aria-modal="true"
    :aria-label="t(isEditingAd ? 'marketplace.management.walletAdEditSection' : 'marketplace.management.walletAdPublishSection')"
    @click.self="emit('close')"
  >
    <section class="management-dialog-panel management-dialog-panel--editor wallet-ad-dialog">
      <header class="management-dialog-header">
        <div>
          <p>{{ t('marketplace.management.walletAdListSection') }}</p>
          <h2>{{ t(isEditingAd ? 'marketplace.management.walletAdEditSection' : 'marketplace.management.walletAdPublishSection') }}</h2>
        </div>
        <button
          type="button"
          class="management-dialog-close"
          :aria-label="t('marketplace.management.closeDialog')"
          @click="emit('close')"
        >
          <AppIcon name="close" :size="18" />
        </button>
      </header>

      <div class="management-dialog-body management-dialog-body--editor wallet-ad-dialog-body">
    <div class="wallet-ad-purpose-bar" role="group" :aria-label="t('marketplace.management.walletAdTypeField')">
      <label class="wallet-ad-purpose-select">
        <span>{{ t('marketplace.management.walletAdTypeField') }}</span>
        <select v-model="adForm.adType">
          <option v-for="option in adTypeOptions" :key="option.value" :value="option.value">{{ t(option.labelKey) }}</option>
        </select>
      </label>
      <button
        v-for="option in adTypeOptions"
        :key="option.value"
        type="button"
        class="wallet-ad-purpose"
        :aria-pressed="adForm.adType === option.value"
        @click="adForm.adType = option.value"
      >
        {{ t(option.labelKey) }}
      </button>
    </div>

    <div class="wallet-ad-workbench-body">
      <div class="wallet-settings-form wallet-ad-editor-form">
        <label class="wallet-settings-field">
          <span>{{ t(isImageOnlyAdType(adForm.adType) ? 'marketplace.management.walletAdDisplayTitleField' : 'marketplace.management.walletAdTitleField') }}</span>
          <input
            v-model="adForm.title"
            type="text"
            maxlength="160"
            :placeholder="t('marketplace.management.walletAdTitlePlaceholder')"
          />
        </label>

        <label class="wallet-settings-field">
          <span>{{ isImageOnlyAdType(adForm.adType) ? t('marketplace.settings.displayAdTextField') : t('marketplace.management.walletAdSummaryField') }}</span>
          <textarea
            v-model="adForm.summary"
            rows="3"
            maxlength="500"
            :placeholder="t('marketplace.management.walletAdSummaryPlaceholder')"
          />
        </label>

        <div class="wallet-settings-field wallet-ad-media-row">
          <span>{{ t('marketplace.management.walletAdMediaTypeField') }}</span>
          <div class="wallet-ad-media-controls">
            <div class="wallet-ad-media-toggle" role="group" :aria-label="t('marketplace.management.walletAdMediaTypeField')">
              <button
                type="button"
                :aria-pressed="adForm.mediaType === 'image'"
                @click="adForm.mediaType = 'image'"
              >
                {{ t('marketplace.management.walletAdMediaTypeImage') }}
              </button>
              <button
                type="button"
                :aria-pressed="adForm.mediaType === 'video'"
                :disabled="isImageOnlyAdType(adForm.adType)"
                @click="adForm.mediaType = 'video'"
              >
                {{ t('marketplace.management.walletAdMediaTypeVideo') }}
              </button>
            </div>
            <label class="wallet-ad-file-picker">
              <input
                type="file"
                :accept="adForm.mediaType === 'video' ? 'video/*' : 'image/*'"
                :aria-label="t('marketplace.management.walletAdMediaUploadField')"
                @change="handleMediaFileChange"
              />
              <strong>{{ t('marketplace.management.walletAdChooseFile') }}</strong>
              <span>{{ selectedMediaName || t('marketplace.management.walletAdNoFile') }}</span>
            </label>
          </div>
        </div>

        <div
          v-if="isDisplayRewardAdType(adForm.adType) || isPriceBannerAdType(adForm.adType)"
          class="wallet-ad-editor-grid wallet-ad-editor-display-grid"
        >
          <label class="wallet-settings-field">
            <span>{{ t('marketplace.management.walletAdDisplayChannelField') }}</span>
            <input
              v-if="isPriceBannerAdType(adForm.adType)"
              :value="t('marketplace.settings.displayAdChannelSupermarket')"
              type="text"
              readonly
            />
            <select
              v-else
              v-model="adForm.displayChannel"
            >
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
          v-if="adForm.adType === 'reward'"
          class="wallet-ad-editor-grid wallet-ad-editor-reward-grid"
        >
          <label class="wallet-settings-field">
            <span>{{ t('marketplace.management.walletAdRewardField') }}</span>
            <input
              v-model.number="adForm.rewardPoints"
              type="number"
              min="1"
              step="1"
            />
          </label>

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
          class="wallet-ad-editor-grid wallet-ad-editor-publish-grid"
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

      </div>

      <aside class="wallet-ad-preview-pane">
        <span>{{ t('marketplace.management.walletAdPreviewField') }}</span>
        <div class="wallet-settings-media-preview wallet-ad-editor-preview">
          <img
            v-if="(adForm.mediaPreviewURL || adForm.mediaURL) && adForm.mediaType === 'image'"
            :src="adForm.mediaPreviewURL || adForm.mediaURL"
            :alt="adForm.title"
          />
          <video
            v-else-if="adForm.mediaPreviewURL || adForm.mediaURL"
            :src="adForm.mediaPreviewURL || adForm.mediaURL"
            controls
            preload="metadata"
          />
          <strong v-else>{{ t(adTypeOptions.find((option) => option.value === adForm.adType)?.labelKey ?? '') }}</strong>
        </div>
      </aside>
    </div>
      </div>

      <footer class="management-dialog-actions">
        <button type="button" class="management-dialog-button management-dialog-button--secondary" @click="emit('close')">{{ t('marketplace.management.cancelAction') }}</button>
        <button type="button" class="management-dialog-button management-dialog-button--primary" :disabled="savingAd" @click="emit('save')">
          {{ savingAd ? t('marketplace.management.saving') : isEditingAd ? t('marketplace.management.walletAdUpdateAction') : t('marketplace.management.walletAdCreateAction') }}
        </button>
      </footer>
    </section>
  </div>
</template>

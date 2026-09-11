<!--
 * 展示廣告位設定頁。
 * 1. 按頻道讀取可用 display 廣告。
 * 2. 綁定 3 個 16:9 短廣告位與 2 個 9:16 長廣告位。
 * 3. 保存 slot 顯示文案與跳轉 URL。
-->
<script setup lang="ts">
import axios from 'axios';
import { computed, onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

import { fetchDisplayAdSettings, fetchStaffRewardAds, saveDisplayAdSettings } from '@/httpapis/wallet';
import type { DisplayAdSlotSaveItem, PublicDisplayAdResponse, StaffRewardAdResponse } from '@/model/payments';
import AppIcon from '@/shared/components/base/AppIcon.vue';
import { useFeedbackStore } from '@/stores/feedback';
import { isPublicDisplayAdActive } from '@/utils/wallet';

type DisplayAdChannel = PublicDisplayAdResponse['display_channel'];

interface DisplayAdSlotForm {
  slotIndex: number;
  kind: 'short' | 'long';
  adTaskId: string;
  displayTitle: string;
  displayText: string;
  targetURL: string;
}

const { t } = useI18n();
const feedbackStore = useFeedbackStore();
const selectedChannel = ref<DisplayAdChannel>('property_sale');
const loadingSettings = ref(false);
const loadingAds = ref(false);
const saving = ref(false);
const availableAds = ref<StaffRewardAdResponse[]>([]);
const slotForms = ref<DisplayAdSlotForm[]>([]);

const channelOptions = computed<Array<{ label: string; value: DisplayAdChannel }>>(() => [
  { label: t('marketplace.settings.displayAdChannelPropertySale'), value: 'property_sale' },
  { label: t('marketplace.settings.displayAdChannelFurniture'), value: 'furniture' },
  { label: t('marketplace.settings.displayAdChannelServicedApartment'), value: 'serviced_apartment' },
]);

// 1. 建立固定右側廣告位表單
const createEmptySlots = (): DisplayAdSlotForm[] =>
  ([
    { slotIndex: 1, kind: 'short' },
    { slotIndex: 2, kind: 'short' },
    { slotIndex: 3, kind: 'short' },
    { slotIndex: 4, kind: 'long' },
    { slotIndex: 5, kind: 'long' },
  ] as Array<Pick<DisplayAdSlotForm, 'slotIndex' | 'kind'>>).map((slot) => ({
    ...slot,
    adTaskId: '',
    displayTitle: '',
    displayText: '',
    targetURL: '',
  }));

// 2. 讀取可用展示廣告
const loadAvailableAds = async (): Promise<void> => {
  loadingAds.value = true;

  try {
    const { data } = await fetchStaffRewardAds({
      page: 1,
      page_size: 50,
      ad_type: 'display',
      display_channel: selectedChannel.value,
      is_active: true,
    });
    availableAds.value = data.data.items.filter((ad) => isPublicDisplayAdActive(ad));
  } catch (error: unknown) {
    feedbackStore.pushToast(
      axios.isAxiosError<{ message?: string }>(error)
        ? error.response?.data?.message ?? t('marketplace.settings.displayAdLoadAdsError')
        : t('marketplace.settings.displayAdLoadAdsError'),
      'error',
    );
    availableAds.value = [];
  } finally {
    loadingAds.value = false;
  }
};

// 3. 讀取廣告位設定
const loadSettings = async (): Promise<void> => {
  loadingSettings.value = true;

  try {
    const { data } = await fetchDisplayAdSettings(selectedChannel.value);
    const slots = createEmptySlots();
    slotForms.value = slots.map((slot) => {
      const savedSlot = data.data.slots.find((item) => item.slot_index === slot.slotIndex);
      const savedAd = savedSlot?.ads[0];
      if (!savedAd) {
        return slot;
      }

      return {
        ...slot,
        adTaskId: savedAd.task_id,
        displayTitle: savedAd.slot_display_title || savedAd.title,
        displayText: savedAd.display_text || savedAd.summary,
        targetURL: savedAd.slot_target_url || savedAd.target_url,
      };
    });
  } catch (error: unknown) {
    feedbackStore.pushToast(
      axios.isAxiosError<{ message?: string }>(error)
        ? error.response?.data?.message ?? t('marketplace.settings.displayAdLoadSettingsError')
        : t('marketplace.settings.displayAdLoadSettingsError'),
      'error',
    );
    slotForms.value = createEmptySlots();
  } finally {
    loadingSettings.value = false;
  }
};

// 4. 取得廣告位可用素材
const availableAdsForSlot = (slot: DisplayAdSlotForm): StaffRewardAdResponse[] =>
  availableAds.value.filter((ad) => {
    if (slot.kind === 'short') {
      return ad.ad_type === 'display_short' || ad.ad_type === 'display';
    }
    return ad.ad_type === 'display_long' || ad.ad_type === 'display';
  });

// 5. 切換廣告時套用預設文案
const applySelectedAd = (slot: DisplayAdSlotForm): void => {
  const ad = availableAds.value.find((item) => item.task_id === slot.adTaskId);
  if (!ad) {
    slot.displayTitle = '';
    slot.displayText = '';
    slot.targetURL = '';
    return;
  }

  slot.displayTitle = ad.slot_display_title || ad.title;
  slot.displayText = ad.display_text || ad.summary;
  slot.targetURL = ad.slot_target_url || ad.target_url;
};

// 6. 保存廣告位設定
const saveSettings = async (): Promise<void> => {
  const slots: DisplayAdSlotSaveItem[] = slotForms.value.map((slot) => ({
    slot_index: slot.slotIndex,
    ad_task_ids: slot.adTaskId ? [slot.adTaskId] : [],
    ads: slot.adTaskId
      ? [{
          ad_task_id: slot.adTaskId,
          display_title: slot.displayTitle.trim(),
          display_text: slot.displayText.trim(),
          target_url: slot.targetURL.trim(),
        }]
      : [],
  }));

  saving.value = true;
  try {
    await saveDisplayAdSettings(selectedChannel.value, slots);
    feedbackStore.pushToast(t('marketplace.settings.displayAdSaveSuccess'), 'success');
    await loadSettings();
  } catch (error: unknown) {
    feedbackStore.pushToast(
      axios.isAxiosError<{ message?: string }>(error)
        ? error.response?.data?.message ?? t('marketplace.settings.displayAdSaveError')
        : t('marketplace.settings.displayAdSaveError'),
      'error',
    );
  } finally {
    saving.value = false;
  }
};

watch(selectedChannel, () => {
  void Promise.all([loadAvailableAds(), loadSettings()]);
});

onMounted(() => {
  slotForms.value = createEmptySlots();
  void Promise.all([loadAvailableAds(), loadSettings()]);
});
</script>

<template>
  <section class="display-ads-page">
    <header class="display-ads-header">
      <div>
        <p class="display-ads-kicker">
          {{ t('marketplace.settings.displayAdKicker') }}
        </p>
        <h1>{{ t('marketplace.settings.displayAdSection') }}</h1>
        <p>{{ t('marketplace.settings.displayAdDescription') }}</p>
      </div>

      <button
        type="button"
        class="display-ads-button display-ads-button--primary"
        :disabled="saving || loadingSettings"
        @click="saveSettings"
      >
        <AppIcon
          name="check-circle"
          :size="16"
        />
        <span>{{ saving ? t('marketplace.settings.saving') : t('marketplace.settings.displayAdSave') }}</span>
      </button>
    </header>

    <div class="display-ads-tabs">
      <button
        v-for="channel in channelOptions"
        :key="channel.value"
        type="button"
        :class="{ 'display-ads-tab--active': channel.value === selectedChannel }"
        @click="selectedChannel = channel.value"
      >
        {{ channel.label }}
      </button>
    </div>

    <div class="display-ads-grid">
      <section class="display-ads-panel">
        <div class="display-ads-panel__header">
          <h2>{{ t('marketplace.settings.displayAdSlotsTitle') }}</h2>
          <p>{{ t('marketplace.settings.displayAdSlotsDescription') }}</p>
        </div>

        <div
          v-if="loadingSettings"
          class="display-ads-empty"
        >
          {{ t('common.status.loading') }}
        </div>

        <div
          v-else
          class="display-ads-slots"
        >
          <article
            v-for="slot in slotForms"
            :key="slot.slotIndex"
            class="display-ads-slot"
          >
            <div class="display-ads-slot__title">
              <h3>{{ t('marketplace.settings.displayAdSlotLabel', { index: slot.slotIndex }) }}</h3>
              <span>{{ slot.kind === 'short' ? '16:9' : '9:16' }}</span>
            </div>

            <label class="display-ads-field">
              <span>{{ t('marketplace.settings.displayAdSelectField') }}</span>
              <select
                v-model="slot.adTaskId"
                :disabled="loadingAds"
                @change="applySelectedAd(slot)"
              >
                <option value="">
                  {{ t('marketplace.settings.emptySlot') }}
                </option>
                <option
                  v-for="ad in availableAdsForSlot(slot)"
                  :key="ad.task_id"
                  :value="ad.task_id"
                >
                  {{ ad.title }}
                </option>
              </select>
            </label>

            <label class="display-ads-field">
              <span>{{ t('marketplace.settings.displayAdTitleField') }}</span>
              <input
                v-model="slot.displayTitle"
                type="text"
                maxlength="80"
              />
            </label>

            <label class="display-ads-field">
              <span>{{ t('marketplace.settings.displayAdTextField') }}</span>
              <textarea
                v-model="slot.displayText"
                rows="2"
                maxlength="160"
              />
            </label>

            <label class="display-ads-field">
              <span>{{ t('marketplace.settings.displayAdTargetField') }}</span>
              <input
                v-model="slot.targetURL"
                type="url"
                placeholder="https://"
              />
            </label>
          </article>
        </div>
      </section>

      <aside class="display-ads-panel display-ads-preview">
        <div class="display-ads-panel__header">
          <h2>{{ t('marketplace.settings.displayAdPreviewTitle') }}</h2>
          <p>{{ t('marketplace.settings.displayAdPreviewDescription') }}</p>
        </div>

        <div class="display-ads-preview__rail">
          <div
            v-for="slot in slotForms"
            :key="`preview-${slot.slotIndex}`"
            class="display-ads-preview__slot"
            :class="`display-ads-preview__slot--${slot.kind}`"
          >
            <span>{{ t('marketplace.settings.displayAdBadge') }}</span>
            <strong>{{ slot.displayTitle || t('marketplace.settings.emptySlot') }}</strong>
            <p>{{ slot.displayText || t('marketplace.settings.emptySlotHint') }}</p>
          </div>
        </div>
      </aside>
    </div>
  </section>
</template>

<style scoped>
.display-ads-page {
  display: grid;
  gap: 1rem;
}

.display-ads-header,
.display-ads-panel {
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
}

.display-ads-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  padding: 1rem;
}

.display-ads-kicker {
  margin: 0 0 0.35rem;
  color: rgb(var(--color-primary));
  font-size: 0.75rem;
  font-weight: 700;
  letter-spacing: 0.12em;
}

.display-ads-header h1,
.display-ads-panel__header h2 {
  margin: 0;
  color: rgb(var(--color-text));
  font-family: var(--font-display);
  font-size: 1.55rem;
  font-weight: 500;
  line-height: 1.3;
}

.display-ads-header p:not(.display-ads-kicker),
.display-ads-panel__header p {
  margin: 0.5rem 0 0;
  color: rgb(var(--color-text-muted));
  font-size: 0.875rem;
  line-height: 1.7;
}

.display-ads-button {
  display: inline-flex;
  min-height: 2.25rem;
  align-items: center;
  justify-content: center;
  gap: 0.45rem;
  border-radius: 8px;
  font-size: 0.8125rem;
  font-weight: 700;
  padding: 0 1rem;
}

.display-ads-button--primary {
  background: rgb(var(--color-primary));
  color: rgb(var(--color-primary-contrast));
}

.display-ads-button:disabled {
  cursor: wait;
  opacity: 0.62;
}

.display-ads-tabs {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.display-ads-tabs button {
  min-height: 2.25rem;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text-muted));
  font-size: 0.8125rem;
  font-weight: 700;
  padding: 0 0.9rem;
}

.display-ads-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 220px;
  gap: 1rem;
  align-items: start;
}

.display-ads-panel__header {
  border-bottom: 1px solid rgb(var(--color-border));
  padding: 1rem;
}

.display-ads-slots {
  display: grid;
  gap: 1rem;
  padding: 1rem;
}

.display-ads-slot {
  display: grid;
  gap: 0.75rem;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface-muted));
  padding: 1rem;
}

.display-ads-slot__title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

.display-ads-slot__title h3 {
  margin: 0;
  color: rgb(var(--color-text));
  font-size: 0.95rem;
  font-weight: 700;
}

.display-ads-slot__title span {
  color: rgb(var(--color-text-muted));
  font-size: 0.75rem;
}

.display-ads-field {
  display: grid;
  gap: 0.35rem;
}

.display-ads-field span {
  color: rgb(var(--color-text-muted));
  font-size: 0.75rem;
  font-weight: 700;
}

.display-ads-field input,
.display-ads-field select,
.display-ads-field textarea {
  width: 100%;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text));
  font: inherit;
  font-size: 0.875rem;
  outline: none;
  padding: 0.7rem;
}

.display-ads-field textarea {
  resize: vertical;
}

.display-ads-empty {
  color: rgb(var(--color-text-muted));
  font-size: 0.875rem;
  padding: 1rem;
}

.display-ads-preview {
  position: sticky;
  top: calc(var(--app-header-offset, 0rem) + 2rem);
}

.display-ads-preview__rail {
  display: grid;
  gap: 0.875rem;
  padding: 1rem;
}

.display-ads-preview__slot {
  display: grid;
  align-content: end;
  gap: 0.35rem;
  aspect-ratio: 16 / 9;
  border: 1px dashed rgb(var(--color-border));
  border-radius: 3px;
  background: linear-gradient(135deg, rgb(var(--color-surface-muted)), rgb(var(--color-surface-raised)));
  padding: 0.85rem;
}

.display-ads-preview__slot--long {
  aspect-ratio: 9 / 16;
}

.display-ads-preview__slot span {
  color: rgb(var(--color-primary));
  font-size: 0.65rem;
  font-weight: 800;
  letter-spacing: 0.12em;
}

.display-ads-preview__slot strong {
  color: rgb(var(--color-text));
  font-size: 0.95rem;
  line-height: 1.4;
}

.display-ads-preview__slot p {
  margin: 0;
  color: rgb(var(--color-text-muted));
  font-size: 0.75rem;
  line-height: 1.6;
}

@media (max-width: 1023px) {
  .display-ads-header {
    flex-direction: column;
  }

  .display-ads-grid {
    grid-template-columns: 1fr;
  }

  .display-ads-preview {
    position: static;
  }
}

.display-ads-page {
  gap: 0.9rem;
}

.display-ads-header,
.display-ads-panel,
.display-ads-button,
.display-ads-tabs button,
.display-ads-slot,
.display-ads-field input,
.display-ads-field select,
.display-ads-field textarea,
.display-ads-preview__slot {
  border-radius: 2px;
}

.display-ads-header,
.display-ads-panel__header,
.display-ads-slots,
.display-ads-preview__rail {
  padding: 0.9rem;
}

.display-ads-kicker,
.display-ads-slot__title span,
.display-ads-field span {
  font-size: 0.7rem;
}

.display-ads-header h1,
.display-ads-panel__header h2 {
  font-size: clamp(1.4rem, 1.9vw, 2rem);
  font-weight: 600;
  line-height: 1.18;
}

.display-ads-panel__header h2 {
  font-size: 0.95rem;
}

.display-ads-header p:not(.display-ads-kicker),
.display-ads-panel__header p,
.display-ads-empty {
  font-size: 0.8125rem;
  line-height: 1.55;
}

.display-ads-button,
.display-ads-tabs button {
  min-height: 2.45rem;
}

.display-ads-slot {
  padding: 0.85rem;
}

.display-ads-slot__title h3,
.display-ads-preview__slot strong {
  font-size: 0.875rem;
}

.display-ads-preview__slot {
  background: rgb(var(--color-surface-muted));
}

.display-ads-preview__slot span {
  font-size: 0.625rem;
}

.display-ads-preview__slot p {
  font-size: 0.72rem;
}

.display-ads-tabs .display-ads-tab--active {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary-soft));
  color: rgb(var(--color-primary));
}
</style>

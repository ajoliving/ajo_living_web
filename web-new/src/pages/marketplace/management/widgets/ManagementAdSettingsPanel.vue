<!--
 * 管理中心 - 廣告設定面板。
 * 1. 讀取 Staff 廣告素材與展示位設定。
 * 2. 配置樓盤、家具與服務式住宅右側 5 個固定比例廣告位。
 * 3. 保存設定到後端 display-ad-settings 接口。
-->
<script setup lang="ts">
/*
 * 廣告設定面板邏輯。
 * 1. 按頻道載入短廣告、長廣告與現有版位設定。
 * 2. 維護 3 個短廣告位與 2 個長廣告位表單。
 * 3. 組裝後端保存 payload。
 */
import axios from 'axios';
import { computed, onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

import { fetchDisplayAdSettings, fetchStaffRewardAds, saveDisplayAdSettings } from '@/httpapis/wallet';
import type { DisplayAdSlotSaveItem, PublicDisplayAdResponse, StaffRewardAdResponse } from '@/model/payments';
import { useFeedbackStore } from '@/stores/feedback';

type DisplayAdChannel = PublicDisplayAdResponse['display_channel'];
type DisplaySlotKind = 'long' | 'short';

interface DisplayAdSlotForm {
  slotIndex: number;
  kind: DisplaySlotKind;
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

const longAds = computed(() => availableAds.value.filter((ad) => ad.ad_type === 'display_long' || ad.ad_type === 'display'));
const shortAds = computed(() => availableAds.value.filter((ad) => ad.ad_type === 'display_short'));

// 1. 建立固定右側廣告位表單
const createEmptySlots = (): DisplayAdSlotForm[] => [
  {
    slotIndex: 1,
    kind: 'short',
    adTaskId: '',
    displayTitle: '',
    displayText: '',
    targetURL: '',
  },
  {
    slotIndex: 2,
    kind: 'short',
    adTaskId: '',
    displayTitle: '',
    displayText: '',
    targetURL: '',
  },
  {
    slotIndex: 3,
    kind: 'short',
    adTaskId: '',
    displayTitle: '',
    displayText: '',
    targetURL: '',
  },
  {
    slotIndex: 4,
    kind: 'long',
    adTaskId: '',
    displayTitle: '',
    displayText: '',
    targetURL: '',
  },
  {
    slotIndex: 5,
    kind: 'long',
    adTaskId: '',
    displayTitle: '',
    displayText: '',
    targetURL: '',
  },
];

// 2. 讀取可用展示廣告
const loadAvailableAds = async (): Promise<void> => {
  loadingAds.value = true;

  try {
    const { data } = await fetchStaffRewardAds({
      page: 1,
      page_size: 80,
      ad_type: 'display',
      display_channel: selectedChannel.value,
      is_active: true,
    });
    availableAds.value = data.data.items;
  } catch (error: unknown) {
    feedbackStore.pushToast(readErrorMessage(error, t('marketplace.settings.displayAdLoadAdsError')), 'error');
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
    const emptySlots = createEmptySlots();
    slotForms.value = emptySlots.map((slot) => {
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
    feedbackStore.pushToast(readErrorMessage(error, t('marketplace.settings.displayAdLoadSettingsError')), 'error');
    slotForms.value = createEmptySlots();
  } finally {
    loadingSettings.value = false;
  }
};

// 4. 取得目前 slot 可用廣告
const adsForSlot = (slot: DisplayAdSlotForm): StaffRewardAdResponse[] =>
  slot.kind === 'long' ? longAds.value : shortAds.value;

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

// 6. 輸出版位規格
const resolveSlotSize = (slot: DisplayAdSlotForm): string =>
  slot.kind === 'long'
    ? t('marketplace.management.walletAdDisplayLongSize')
    : t('marketplace.management.walletAdDisplayShortSize');

// 7. 保存廣告位設定
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
    feedbackStore.pushToast(readErrorMessage(error, t('marketplace.settings.displayAdSaveError')), 'error');
  } finally {
    saving.value = false;
  }
};

// 8. 讀取錯誤訊息
const readErrorMessage = (error: unknown, fallback: string): string =>
  axios.isAxiosError<{ message?: string }>(error)
    ? error.response?.data?.message ?? fallback
    : fallback;

watch(selectedChannel, () => {
  void Promise.all([loadAvailableAds(), loadSettings()]);
});

onMounted(() => {
  slotForms.value = createEmptySlots();
  void Promise.all([loadAvailableAds(), loadSettings()]);
});
</script>

<template>
  <div class="ad-settings-panel">
    <section class="work-hero">
      <div>
        <div class="work-kicker">Display Ads</div>
        <h2 class="work-title">{{ t('marketplace.management.adSettings') }}</h2>
        <p class="work-desc">{{ t('marketplace.settings.displayAdDescription') }}</p>
      </div>
      <button
        type="button"
        class="work-action"
        :disabled="saving || loadingSettings"
        @click="saveSettings"
      >
        {{ saving ? t('marketplace.settings.saving') : t('marketplace.settings.displayAdSave') }}
      </button>
    </section>

    <div class="display-ads-tabs">
      <button
        v-for="channel in channelOptions"
        :key="channel.value"
        type="button"
        :class="{ 'is-active': channel.value === selectedChannel }"
        @click="selectedChannel = channel.value"
      >
        {{ channel.label }}
      </button>
    </div>

    <section class="admin-ad-settings">
      <div class="work-card">
        <div class="work-card-title">{{ t('marketplace.settings.displayAdsLibraryTitle') }}</div>
        <div
          v-if="loadingAds"
          class="ad-empty"
        >
          {{ t('common.status.loading') }}
        </div>
        <div
          v-else-if="availableAds.length === 0"
          class="ad-empty"
        >
          {{ t('marketplace.settings.displayAdsEmptyLibrary') }}
        </div>
        <div
          v-else
          class="ad-source-list"
        >
          <article
            v-for="item in availableAds"
            :key="item.task_id"
            class="ad-source-item"
          >
            <div class="ad-source-thumb">
              <img
                v-if="item.media_url"
                :src="item.media_url"
                :alt="item.title"
              />
              <span v-else>{{ item.title.slice(0, 1) }}</span>
            </div>
            <div class="ad-source-body">
              <div class="ad-source-title">{{ item.title }}</div>
              <div class="ad-source-meta">
                {{ item.ad_type === 'display_short' ? t('marketplace.management.walletAdTypeDisplayShort') : t('marketplace.management.walletAdTypeDisplayLong') }}
              </div>
              <span class="ad-source-status good">
                {{ item.is_active ? t('common.state.active') : t('common.state.hidden') }}
              </span>
            </div>
          </article>
        </div>
      </div>

      <div class="work-card">
        <div class="work-card-title">{{ t('marketplace.settings.displayAdSlotsTitle') }}</div>
        <div
          v-if="loadingSettings"
          class="ad-empty"
        >
          {{ t('common.status.loading') }}
        </div>
        <div
          v-else
          class="ad-slot-grid"
        >
          <article
            v-for="slot in slotForms"
            :key="slot.slotIndex"
            class="ad-slot-card"
          >
            <div class="ad-slot-head">
              <div>
                <div class="ad-slot-name">{{ t('marketplace.settings.displayAdSlotLabel', { index: slot.slotIndex }) }}</div>
                <strong>{{ resolveSlotSize(slot) }}</strong>
              </div>
              <span>{{ slot.kind === 'long' ? t('marketplace.management.walletAdTypeDisplayLong') : t('marketplace.management.walletAdTypeDisplayShort') }}</span>
            </div>

            <label class="ad-field">
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
                  v-for="ad in adsForSlot(slot)"
                  :key="ad.task_id"
                  :value="ad.task_id"
                >
                  {{ ad.title }}
                </option>
              </select>
            </label>

            <label class="ad-field">
              <span>{{ t('marketplace.settings.displayAdTitleField') }}</span>
              <input
                v-model="slot.displayTitle"
                type="text"
                maxlength="160"
              />
            </label>

            <label class="ad-field">
              <span>{{ t('marketplace.settings.displayAdTextField') }}</span>
              <textarea
                v-model="slot.displayText"
                rows="2"
                maxlength="255"
              />
            </label>

            <label class="ad-field">
              <span>{{ t('marketplace.settings.displayAdTargetField') }}</span>
              <input
                v-model="slot.targetURL"
                type="url"
                placeholder="https://"
              />
            </label>
          </article>
        </div>
      </div>
    </section>
  </div>
</template>

<style scoped>
/* 1. 面板容器 */
.ad-settings-panel {
  display: grid;
  gap: 14px;
}

/* 2. 標題列 */
.work-hero {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 18px;
  border-bottom: 1px solid var(--bdr);
  padding: 0 0 12px;
}

.work-kicker {
  margin-bottom: 6px;
  color: var(--ink-3);
  font-size: 10px;
  letter-spacing: 1.6px;
  text-transform: uppercase;
}

.work-title {
  margin: 0;
  color: var(--ink);
  font-size: 20px;
  font-weight: 600;
  line-height: 1.25;
}

.work-desc {
  max-width: 620px;
  margin: 8px 0 0;
  color: var(--ink-3);
  font-size: 13px;
  line-height: 1.7;
}

.work-action {
  border: 0;
  border-radius: 6px;
  background: var(--accent);
  color: #fff;
  cursor: pointer;
  font-family: inherit;
  font-size: 12px;
  font-weight: 600;
  padding: 10px 14px;
  white-space: nowrap;
}

.work-action:disabled {
  cursor: wait;
  opacity: 0.6;
}

/* 3. 頻道切換 */
.display-ads-tabs {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.display-ads-tabs button {
  min-height: 34px;
  border: 1px solid var(--bdr);
  border-radius: 6px;
  background: #fff;
  color: var(--ink-2);
  cursor: pointer;
  font-family: inherit;
  font-size: 12px;
  font-weight: 700;
  padding: 0 12px;
}

.display-ads-tabs button.is-active {
  border-color: var(--accent);
  background: var(--brand-light);
  color: var(--accent);
}

/* 4. 主設定區 */
.admin-ad-settings {
  display: grid;
  grid-template-columns: 340px minmax(0, 1fr);
  gap: 14px;
}

.work-card {
  border: 1px solid var(--bdr);
  border-radius: 8px;
  background: #fff;
  padding: 16px;
}

.work-card-title {
  margin-bottom: 10px;
  color: var(--ink);
  font-size: 14px;
  font-weight: 600;
}

.ad-empty {
  border: 1px dashed var(--bdr);
  border-radius: 6px;
  color: var(--ink-3);
  font-size: 12px;
  padding: 14px;
}

/* 5. 素材列表 */
.ad-source-list {
  display: grid;
  gap: 10px;
  max-height: 620px;
  overflow: auto;
  padding-right: 4px;
}

.ad-source-item {
  display: grid;
  grid-template-columns: 88px minmax(0, 1fr);
  gap: 12px;
  border: 1px solid var(--bdr);
  border-radius: 8px;
  background: var(--sur);
  padding: 10px;
}

.ad-source-thumb {
  display: grid;
  min-height: 70px;
  overflow: hidden;
  place-items: center;
  border-radius: 6px;
  background: var(--sur-3);
  color: var(--ink-3);
  font-size: 24px;
  font-weight: 700;
}

.ad-source-thumb img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.ad-source-title {
  color: var(--ink);
  font-size: 13px;
  font-weight: 700;
}

.ad-source-meta {
  margin-top: 4px;
  color: var(--ink-3);
  font-size: 11px;
  line-height: 1.5;
}

.ad-source-status {
  display: inline-flex;
  margin-top: 8px;
  border-radius: 999px;
  background: var(--sur-2);
  color: var(--ink-2);
  font-size: 11px;
  font-weight: 700;
  padding: 4px 8px;
}

.ad-source-status.good {
  background: var(--success-bg);
  color: var(--success);
}

/* 6. 廣告位 */
.ad-slot-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.ad-slot-card {
  display: grid;
  gap: 10px;
  border: 1px solid var(--bdr);
  border-radius: 8px;
  background: var(--sur);
  padding: 12px;
}

.ad-slot-head {
  display: flex;
  justify-content: space-between;
  gap: 10px;
}

.ad-slot-name,
.ad-slot-head span,
.ad-field span {
  color: var(--ink-3);
  font-size: 12px;
  font-weight: 700;
}

.ad-slot-head strong {
  display: block;
  margin-top: 3px;
  color: var(--ink);
  font-size: 14px;
}

.ad-field {
  display: grid;
  gap: 5px;
}

.ad-field input,
.ad-field select,
.ad-field textarea {
  width: 100%;
  border: 1px solid var(--bdr);
  border-radius: 6px;
  background: #fff;
  color: var(--ink);
  font: inherit;
  font-size: 12px;
  outline: none;
  padding: 8px 10px;
}

.ad-field textarea {
  resize: vertical;
}

/* 7. 響應式 */
@media (max-width: 980px) {
  .admin-ad-settings {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 560px) {
  .work-hero {
    flex-direction: column;
    align-items: flex-start;
  }

  .ad-slot-grid {
    grid-template-columns: 1fr;
  }
}
</style>

/*
 * 展示廣告設定資料流程。
 * 1. 固定管理樓盤租售、家具與服務式住宅三個展示頻道。
 * 2. 左側載入全部圖片廣告素材列表，右側配置每個頻道 10 個廣告位。
 * 3. 每個廣告位可保存多個廣告，前台每次讀取時由後端隨機挑選一個展示。
 */
import axios from 'axios';
import { computed, onMounted, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

import { fetchDisplayAdSettings, fetchStaffRewardAds, saveDisplayAdSettings } from '@/domains/payments/api';
import type { DisplayAdSlotResponse, StaffRewardAdResponse } from '@/domains/payments/model';
import { useFeedbackStore } from '@/app/stores/feedback';
import { formatDate } from '@/shared/utils/format';

type DisplayAdChannel = 'property_sale' | 'furniture' | 'serviced_apartment';
type DisplayAdLayout = 'image_full' | 'image_text' | 'text_compact';
type DisplayAdLayoutFilter = DisplayAdLayout | '';

interface DisplayAdChannelOption {
  code: DisplayAdChannel;
  labelKey: string;
}

interface DisplayAdSlotView {
  slotIndex: number;
  layout: DisplayAdLayout;
  ads: StaffRewardAdResponse[];
}

export const displayAdChannels: DisplayAdChannelOption[] = [
  { code: 'property_sale', labelKey: 'marketplace.settings.displayAdChannelPropertySale' },
  { code: 'furniture', labelKey: 'marketplace.settings.displayAdChannelFurniture' },
  { code: 'serviced_apartment', labelKey: 'marketplace.settings.displayAdChannelServicedApartment' },
];

interface UseDisplayAdsSettingsPageOptions {
  channel: DisplayAdChannel;
}

// 1. 建立展示廣告設定流程
export const useDisplayAdsSettingsPage = (options: UseDisplayAdsSettingsPageOptions) => {
  const { t } = useI18n();
  const feedbackStore = useFeedbackStore();
  const activeChannel = computed(() => options.channel);
  const loadingAds = ref(false);
  const loadingSlots = ref(false);
  const saving = ref(false);
  const keyword = ref('');
  const layoutFilter = ref<DisplayAdLayoutFilter>('');
  const displayAds = ref<StaffRewardAdResponse[]>([]);
  const selectedAdId = ref('');
  const savedSnapshot = ref('');
  const slots = reactive<DisplayAdSlotView[]>(createEmptySlots());
  const selectedAd = computed(() => displayAds.value.find((ad) => ad.task_id === selectedAdId.value) ?? null);
  const assignedAdIds = computed(() => new Set(slots.flatMap((slot) => slot.ads.map((ad) => ad.task_id))));
  const currentSnapshot = computed(() => buildSlotsSnapshot(slots));
  const hasUnsavedChanges = computed(() => !loadingSlots.value && currentSnapshot.value !== savedSnapshot.value);
  const filteredAds = computed(() => {
    const trimmedKeyword = keyword.value.trim().toLowerCase();
    return displayAds.value.filter((ad) => {
      const matchesLayout = !layoutFilter.value || ad.display_layout === layoutFilter.value;
      const matchesKeyword = !trimmedKeyword ||
        [ad.title, ad.summary, ad.task_id].join(' ').toLowerCase().includes(trimmedKeyword);
      return matchesLayout && matchesKeyword;
    });
  });
  const canSave = computed(() => !loadingSlots.value && !saving.value);

  // 1.1 讀取可選圖片廣告
  const loadAds = async (): Promise<void> => {
    loadingAds.value = true;
    try {
      const { data } = await fetchStaffRewardAds({
        page: 1,
        page_size: 100,
        ad_type: 'display',
      });
      displayAds.value = data.data.items.filter((ad) =>
        ad.media_type === 'image' && ['image_full', 'image_text', 'text_compact'].includes(ad.display_layout),
      );
    } catch (error) {
      feedbackStore.pushToast(resolveErrorMessage(error, t('marketplace.settings.displayAdsLoadAdsError')), 'error');
    } finally {
      loadingAds.value = false;
    }
  };

  // 1.2 讀取目前頻道 slot 設定
  const loadSlots = async (): Promise<void> => {
    loadingSlots.value = true;
    try {
      const { data } = await fetchDisplayAdSettings(activeChannel.value);
      applySlots(slots, data.data.slots);
      savedSnapshot.value = buildSlotsSnapshot(slots);
    } catch (error) {
      resetSlots(slots);
      savedSnapshot.value = buildSlotsSnapshot(slots);
      feedbackStore.pushToast(resolveErrorMessage(error, t('marketplace.settings.displayAdsLoadSettingsError')), 'error');
    } finally {
      loadingSlots.value = false;
    }
  };

  // 1.3 保存目前頻道 slot 設定
  const saveSlots = async (): Promise<boolean> => {
    if (!canSave.value) {
      return false;
    }
    saving.value = true;
    try {
      const { data } = await saveDisplayAdSettings(
        activeChannel.value,
        slots.map((slot) => ({
          slot_index: slot.slotIndex,
          ad_task_ids: slot.ads.map((ad) => ad.task_id),
          ads: slot.ads.map((ad) => ({
            ad_task_id: ad.task_id,
            display_title: ad.slot_display_title.trim(),
            display_text: ad.display_text.trim(),
            target_url: ad.slot_target_url.trim(),
          })),
        })),
      );
      applySlots(slots, data.data.slots);
      savedSnapshot.value = buildSlotsSnapshot(slots);
      feedbackStore.pushToast(t('marketplace.settings.displayAdsSaveSuccess'), 'success');
      return true;
    } catch (error) {
      feedbackStore.pushToast(resolveErrorMessage(error, t('marketplace.settings.displayAdsSaveError')), 'error');
      return false;
    } finally {
      saving.value = false;
    }
  };

  // 1.4 將左側選中廣告加入 slot
  const assignSelectedAd = (slot: DisplayAdSlotView): void => {
    const ad = selectedAd.value;
    if (!ad || slot.ads.some((item) => item.task_id === ad.task_id)) {
      return;
    }
    if (ad.display_layout !== slot.layout) {
      feedbackStore.pushToast(t('marketplace.settings.displayAdsLayoutMismatch'), 'error');
      return;
    }
    slot.ads.push({
      ...ad,
      slot_display_title: ad.title,
      display_text: ad.summary,
      slot_target_url: ad.target_url,
    });
  };

  // 1.4.1 更新 slot 廣告展示標題
  const updateSlotAdDisplayTitle = (ad: StaffRewardAdResponse, value: string): void => {
    ad.slot_display_title = value;
  };

  // 1.4.2 更新 slot 廣告展示文字
  const updateSlotAdDisplayText = (ad: StaffRewardAdResponse, value: string): void => {
    ad.display_text = value;
  };

  // 1.4.3 更新 slot 廣告跳轉連結
  const updateSlotAdTargetURL = (ad: StaffRewardAdResponse, value: string): void => {
    ad.slot_target_url = value;
  };

  // 1.5 從 slot 移除指定廣告
  const removeAdFromSlot = (slot: DisplayAdSlotView, ad: StaffRewardAdResponse): void => {
    slot.ads = slot.ads.filter((item) => item.task_id !== ad.task_id);
  };

  // 1.6 清空 slot
  const clearSlot = (slot: DisplayAdSlotView): void => {
    slot.ads = [];
  };

  // 1.7 選中左側圖片廣告
  const selectAd = (ad: StaffRewardAdResponse): void => {
    selectedAdId.value = ad.task_id;
  };

  watch(activeChannel, async () => {
    resetSlots(slots);
    await Promise.all([loadAds(), loadSlots()]);
  });

  onMounted(async () => {
    await Promise.all([loadAds(), loadSlots()]);
  });

  return {
    activeChannel,
    assignedAdIds,
    canSave,
    clearSlot,
    filteredAds,
    formatDate,
    hasUnsavedChanges,
    keyword,
    layoutFilter,
    loadingAds,
    loadingSlots,
    saveSlots,
    saving,
    selectedAd,
    selectedAdId,
    selectAd,
    assignSelectedAd,
    removeAdFromSlot,
    updateSlotAdDisplayTitle,
    updateSlotAdDisplayText,
    updateSlotAdTargetURL,
    slots,
    t,
  };
};

// 2. 建立固定 10 個空白 slot
const createEmptySlots = (): DisplayAdSlotView[] =>
  Array.from({ length: 10 }, (_, index) => ({
    slotIndex: index + 1,
    layout: resolveSlotLayout(index + 1),
    ads: [],
  }));

// 3. 套用後端 slot 設定
const applySlots = (slots: DisplayAdSlotView[], items: DisplayAdSlotResponse[]): void => {
  const itemMap = new Map(items.map((item) => [item.slot_index, item]));
  slots.forEach((slot) => {
    const item = itemMap.get(slot.slotIndex);
    slot.layout = item?.layout ?? resolveSlotLayout(slot.slotIndex);
    slot.ads = item?.ads ?? [];
  });
};

// 4. 解析 slot 固定樣式
const resolveSlotLayout = (slotIndex: number): DisplayAdLayout => {
  if (slotIndex <= 3) {
    return 'text_compact';
  }
  if (slotIndex <= 6) {
    return 'image_text';
  }
  return 'image_full';
};

// 5. 重置 slot
const resetSlots = (slots: DisplayAdSlotView[]): void => {
  slots.forEach((slot) => {
    slot.layout = resolveSlotLayout(slot.slotIndex);
    slot.ads = [];
  });
};

// 6. 建立 slot 保存狀態快照
const buildSlotsSnapshot = (slots: DisplayAdSlotView[]): string =>
  JSON.stringify(
    slots.map((slot) => ({
      slotIndex: slot.slotIndex,
      ads: slot.ads.map((ad) => ({
        adTaskId: ad.task_id,
        displayTitle: ad.slot_display_title.trim(),
        displayText: ad.display_text.trim(),
        targetURL: ad.slot_target_url.trim(),
      })),
    })),
  );

// 7. 解析 API 錯誤訊息
const resolveErrorMessage = (error: unknown, fallback: string): string =>
  axios.isAxiosError(error) ? error.response?.data?.message ?? fallback : fallback;

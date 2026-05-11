/*
 * 二手交易設定頁 - 狀態與資料流程。
 * 1. 管理全部帖子查詢、售出與下架操作。
 * 2. 管理發現頁大推與分類輪播廣告位。
 */
import axios from 'axios';
import { computed, onMounted, reactive, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { onBeforeRouteLeave } from 'vue-router';

import {
  deactivateSettingsSecondhandListing,
  fetchSettingsDiscoverPlacements,
  fetchSettingsSecondhandListings,
  markSettingsSecondhandListingSold,
  saveSettingsDiscoverPlacements,
} from '@/httpapis/secondhand-listings';
import {
  getMarketplaceCategoryLabel,
  marketplaceCategories,
  type MarketplaceCategoryCode,
} from '@/constants/marketplace';
import type {
  DiscoverPlacementPayload,
  DiscoverPlacementResponse,
  SecondhandListingSummaryResponse,
} from '@/model/marketplace';
import { useFeedbackStore } from '@/stores/feedback';
import { usePreferenceStore } from '@/stores/preferences';
import { formatDate, formatPrice } from '@/utils/format';
import {
  resolveListingCategoryLabel,
  resolveListingCommunityName,
  resolveListingCoverImage,
  resolveListingPrice,
  resolveListingPublishedAt,
  resolveListingStatus,
  resolveListingSummary,
  resolveListingTitle,
} from '@/utils/marketplace';

type SettingsStatusFilter = '' | 'draft' | 'active' | 'hidden' | 'expired' | 'sold';
type SettingsPlacementScene = 'discover_hero' | 'discover_category_carousel';
type SettingsLeaveDecision = 'save' | 'discard' | 'stay';

interface SlotView {
  key: string;
  scene: SettingsPlacementScene;
  categoryCode: MarketplaceCategoryCode | '';
  slotIndex: number;
  listingId: string;
  listing?: SecondhandListingSummaryResponse;
}

// 1. 建立空白 slot
const createSlot = (
  scene: SettingsPlacementScene,
  categoryCode: MarketplaceCategoryCode | '',
  slotIndex: number,
): SlotView => ({
  key: `${scene}-${categoryCode || 'hero'}-${slotIndex}`,
  scene,
  categoryCode,
  slotIndex,
  listingId: '',
});

// 2. 將後端 slot 轉為頁面 slot
const assignSlotListing = (slot: SlotView, placement?: DiscoverPlacementResponse): SlotView => ({
  ...slot,
  listing: placement?.listing,
  listingId: placement?.listing?.listing_id ?? '',
});

// 3. 管理二手交易設定頁資料
export const useMarketplaceSettingsPage = () => {
  const { t } = useI18n();
  const feedbackStore = useFeedbackStore();
  const preferenceStore = usePreferenceStore();
  const listingsLoading = ref(false);
  const placementsLoading = ref(false);
  const savingPlacements = ref(false);
  const keyword = ref('');
  const categoryCode = ref<MarketplaceCategoryCode | ''>('');
  const status = ref<SettingsStatusFilter>('');
  const savedPlacementSnapshot = ref('');
  const isLeavePromptOpen = ref(false);
  const listings = ref<SecondhandListingSummaryResponse[]>([]);
  const selectedListingId = ref('');
  const selectedCategoryForSlots = ref<MarketplaceCategoryCode>('home_furniture');
  let resolveLeavePrompt: ((decision: SettingsLeaveDecision) => void) | null = null;

  const heroSlots = ref<SlotView[]>(
    Array.from({ length: 4 }, (_, index) => createSlot('discover_hero', '', index + 1)),
  );
  const categorySlots = reactive<Record<MarketplaceCategoryCode, SlotView[]>>(
    marketplaceCategories.reduce((result, category) => {
      result[category.value] = Array.from({ length: 10 }, (_, index) =>
        createSlot('discover_category_carousel', category.value, index + 1),
      );
      return result;
    }, {} as Record<MarketplaceCategoryCode, SlotView[]>),
  );

  const categoryOptions = computed(() => [
    { label: t('marketplace.settings.allCategories'), value: '' },
    ...marketplaceCategories.map((category) => ({
      label: getMarketplaceCategoryLabel(category.value, preferenceStore.locale),
      value: category.value,
    })),
  ]);

  const statusOptions = computed<Array<{ label: string; value: SettingsStatusFilter }>>(() => [
    { label: t('marketplace.settings.allStatuses'), value: '' },
    { label: t('marketplace.mine.draft'), value: 'draft' },
    { label: t('common.state.active'), value: 'active' },
    { label: t('marketplace.mine.hidden'), value: 'hidden' },
    { label: t('common.state.expired'), value: 'expired' },
    { label: t('common.state.sold'), value: 'sold' },
  ]);

  const selectedListing = computed(() =>
    listings.value.find((listing) => listing.listing_id === selectedListingId.value),
  );

  const visibleCategorySlots = computed(() => categorySlots[selectedCategoryForSlots.value]);

  // 3.2 建立發現廣告位 payload
  const buildPlacementPayload = (): DiscoverPlacementPayload[] =>
    [
      ...heroSlots.value,
      ...marketplaceCategories.flatMap((category) => categorySlots[category.value]),
    ].map((slot) => ({
      scene: slot.scene,
      category_code: slot.categoryCode,
      slot_index: slot.slotIndex,
      listing_id: slot.listingId,
    }));

  // 3.3 序列化發現廣告位 payload
  const serializePlacementPayload = (payload: DiscoverPlacementPayload[]): string =>
    JSON.stringify(
      payload.map((item) => ({
        scene: item.scene,
        category_code: item.category_code,
        slot_index: item.slot_index,
        listing_id: item.listing_id,
      })),
    );

  // 3.4 更新已保存快照
  const refreshPlacementSnapshot = (): void => {
    savedPlacementSnapshot.value = serializePlacementPayload(buildPlacementPayload());
  };

  const isPlacementsDirty = computed(
    () =>
      savedPlacementSnapshot.value.trim().length > 0 &&
      serializePlacementPayload(buildPlacementPayload()) !== savedPlacementSnapshot.value,
  );

  // 3.7 打開離開確認彈窗
  const requestLeaveDecision = (): Promise<SettingsLeaveDecision> => {
    isLeavePromptOpen.value = true;

    return new Promise((resolve) => {
      resolveLeavePrompt = resolve;
    });
  };

  // 3.8 回應離開確認彈窗
  const handleLeavePromptDecision = (decision: SettingsLeaveDecision): void => {
    isLeavePromptOpen.value = false;
    resolveLeavePrompt?.(decision);
    resolveLeavePrompt = null;
  };

  // 3.9 讀取全部帖子
  const loadListings = async (): Promise<void> => {
    listingsLoading.value = true;

    try {
      const { data } = await fetchSettingsSecondhandListings({
        page: 1,
        page_size: 80,
        keyword: keyword.value.trim() || undefined,
        category_code: categoryCode.value || undefined,
        status: status.value || undefined,
      });
      listings.value = data.data.items;
      if (selectedListingId.value && !listings.value.some((item) => item.listing_id === selectedListingId.value)) {
        selectedListingId.value = '';
      }
    } catch (error) {
      feedbackStore.pushToast(resolveErrorMessage(error, t('marketplace.settings.loadListingsError')), 'error');
    } finally {
      listingsLoading.value = false;
    }
  };

  // 3.10 讀取發現頁廣告位
  const loadPlacements = async (): Promise<void> => {
    placementsLoading.value = true;

    try {
      const { data } = await fetchSettingsDiscoverPlacements();
      heroSlots.value = heroSlots.value.map((slot) =>
        assignSlotListing(slot, data.data.hero.find((placement) => placement.slot_index === slot.slotIndex)),
      );

      marketplaceCategories.forEach((category) => {
        categorySlots[category.value] = categorySlots[category.value].map((slot) =>
          assignSlotListing(
            slot,
            (data.data.categories[category.value] ?? []).find((placement) => placement.slot_index === slot.slotIndex),
          ),
        );
      });
      refreshPlacementSnapshot();
    } catch (error) {
      feedbackStore.pushToast(resolveErrorMessage(error, t('marketplace.settings.loadPlacementsError')), 'error');
    } finally {
      placementsLoading.value = false;
    }
  };

  // 3.11 選擇帖子
  const selectListing = (listingId: string): void => {
    selectedListingId.value = listingId;
  };

  // 3.12 設定 slot 使用目前帖子
  const assignSelectedToSlot = (slot: SlotView): void => {
    if (!selectedListing.value) {
      feedbackStore.pushToast(t('marketplace.settings.selectListingFirst'), 'error');
      return;
    }

    if (slot.scene === 'discover_category_carousel' && selectedListing.value.category_code !== slot.categoryCode) {
      feedbackStore.pushToast(t('marketplace.settings.categoryMismatch'), 'error');
      return;
    }

    slot.listingId = selectedListing.value.listing_id;
    slot.listing = selectedListing.value;
  };

  // 3.13 清空 slot
  const clearSlot = (slot: SlotView): void => {
    slot.listingId = '';
    slot.listing = undefined;
  };

  // 3.14 儲存全部廣告位
  const savePlacements = async (): Promise<boolean> => {
    savingPlacements.value = true;

    try {
      const payload = buildPlacementPayload();

      const { data } = await saveSettingsDiscoverPlacements(payload);
      heroSlots.value = heroSlots.value.map((slot) =>
        assignSlotListing(slot, data.data.hero.find((placement) => placement.slot_index === slot.slotIndex)),
      );
      marketplaceCategories.forEach((category) => {
        categorySlots[category.value] = categorySlots[category.value].map((slot) =>
          assignSlotListing(
            slot,
            (data.data.categories[category.value] ?? []).find((placement) => placement.slot_index === slot.slotIndex),
          ),
        );
      });
      refreshPlacementSnapshot();
      feedbackStore.pushToast(t('marketplace.settings.saveSuccess'), 'success');
      return true;
    } catch (error) {
      feedbackStore.pushToast(resolveErrorMessage(error, t('marketplace.settings.saveError')), 'error');
      return false;
    } finally {
      savingPlacements.value = false;
    }
  };

  // 3.15 標記帖子售出
  const markSold = async (listingId: string): Promise<void> => {
    if (!window.confirm(t('marketplace.settings.confirmSold'))) {
      return;
    }

    try {
      await markSettingsSecondhandListingSold(listingId);
      feedbackStore.pushToast(t('marketplace.mine.statusUpdated'), 'success');
      await Promise.all([loadListings(), loadPlacements()]);
    } catch (error) {
      feedbackStore.pushToast(resolveErrorMessage(error, t('marketplace.mine.updateError')), 'error');
    }
  };

  // 3.16 下架帖子
  const deactivate = async (listingId: string): Promise<void> => {
    if (!window.confirm(t('marketplace.settings.confirmDeactivate'))) {
      return;
    }

    try {
      await deactivateSettingsSecondhandListing(listingId);
      feedbackStore.pushToast(t('marketplace.mine.statusUpdated'), 'success');
      await Promise.all([loadListings(), loadPlacements()]);
    } catch (error) {
      feedbackStore.pushToast(resolveErrorMessage(error, t('marketplace.mine.updateError')), 'error');
    }
  };

  onMounted(() => {
    void Promise.all([loadListings(), loadPlacements()]);
  });

  onBeforeRouteLeave(async () => {
    if (!isPlacementsDirty.value) {
      return true;
    }

    const decision = await requestLeaveDecision();
    if (decision === 'stay') {
      return false;
    }
    if (decision === 'discard') {
      return true;
    }

    return savePlacements();
  });

  return {
    assignSelectedToSlot,
    categoryCode,
    categoryOptions,
    categorySlots,
    clearSlot,
    deactivate,
    formatDate,
    formatPrice,
    getMarketplaceCategoryLabel,
    heroSlots,
    handleLeavePromptDecision,
    isLeavePromptOpen,
    isPlacementsDirty,
    keyword,
    listings,
    listingsLoading,
    loadListings,
    markSold,
    marketplaceCategories,
    placementsLoading,
    preferenceStore,
    resolveListingCategoryLabel,
    resolveListingCommunityName,
    resolveListingCoverImage,
    resolveListingPrice,
    resolveListingPublishedAt,
    resolveListingStatus,
    resolveListingSummary,
    resolveListingTitle,
    savePlacements,
    savingPlacements,
    selectListing,
    selectedCategoryForSlots,
    selectedListingId,
    selectedListing,
    status,
    statusOptions,
    t,
    visibleCategorySlots,
  };
};

// 4. 解析 API 錯誤訊息
const resolveErrorMessage = (error: unknown, fallback: string): string =>
  axios.isAxiosError(error) ? error.response?.data?.message ?? fallback : fallback;

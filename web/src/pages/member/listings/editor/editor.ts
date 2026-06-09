/*
 * 二手帖子發布頁 - 狀態與資料。
 * 1. 管理發布表單、本地圖片暫存、草稿與立即發布流程。
 * 2. 接入真實二手帖子 API，支援建立、編輯回填與發布。
 * 3. 集中處理發布前檢查、預覽資料與錯誤提示。
 */
import axios from 'axios';
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { onBeforeRouteLeave, useRoute, useRouter } from 'vue-router';

import {
  createSecondhandListing,
  fetchSecondhandListingDetail,
  publishSecondhandListing,
  updateSecondhandListing,
} from '@/domains/marketplace/api';
import { fetchPosBuildings } from '@/domains/building/api';
import { completeUpload, createUploadPresign } from '@/domains/media/uploads-api';
import {
  getMarketplaceCategoryLabel,
  getMarketplaceConditionLabel,
  getMarketplaceDistrictLabel,
  getMarketplaceOptionLabel,
  marketplaceCategories,
  marketplaceConditions,
  marketplaceDistricts,
  marketplacePriceModes,
  type MarketplaceCategoryCode,
  type MarketplaceConditionCode,
  type MarketplaceDistrictCode,
  type MarketplacePriceMode,
} from '@/domains/marketplace/constants';
import type {
  MediaAssetResponse,
  SecondhandListingDetailResponse,
  UpsertSecondhandListingPayload,
} from '@/domains/marketplace/model';
import type { PosBuilding } from '@/domains/building/model';
import { useFeedbackStore } from '@/app/stores/feedback';
import { usePreferenceStore } from '@/app/stores/preferences';
import { useSessionStore } from '@/app/stores/session';
import { formatPrice } from '@/shared/utils/format';
import { buildUploadHeaders } from '@/shared/utils/upload';
import { formatAjoPoints, resolveWalletChargeCost, resolveWalletDraftChargeCost } from '@/shared/utils/wallet';

export type ListingEditorVisibility = 'public' | 'building_only';
export type ListingEditorBusinessStatus = 'available' | 'sold';

type EditorStepStatus = 'done' | 'current' | 'idle';
type EditorLeaveDecision = 'save' | 'discard' | 'stay';

export interface MarketplaceListingEditorOptions {
  onClose?: () => void;
  onSaved?: () => void | Promise<void>;
  useModalMode?: boolean;
}

interface EditorSnapshot {
  formState: ListingEditorFormState;
  images: Array<{
    mediaAssetId: string;
    fileName: string;
    fileSize: number;
    fileType: string;
    isCover: boolean;
  }>;
}

export interface EditorOption<TValue extends string = string> {
  label: string;
  value: TValue;
}

export interface EditorWorkflowStep {
  label: string;
  description: string;
  status: EditorStepStatus;
}

export interface EditorMetric {
  label: string;
  value: string;
}

export interface EditorImageSlot {
  id: string;
  label: string;
  url?: string;
  mediaAssetId?: string;
  file?: File;
  fileName?: string;
  isCover: boolean;
  objectUrl?: string;
  uploading: boolean;
}

export interface EditorChecklistItem {
  label: string;
  complete: boolean;
}

export interface ListingEditorFormState {
  title: string;
  categoryCode: MarketplaceCategoryCode;
  priceMode: MarketplacePriceMode;
  price: number;
  isDonation: boolean;
  condition: MarketplaceConditionCode;
  districtCode: MarketplaceDistrictCode;
  visibility: ListingEditorVisibility;
  communityId: string;
  communityName: string;
  summary: string;
  description: string;
  dimensionLength: string;
  dimensionWidth: string;
  dimensionHeight: string;
  dimensionWeight: string;
  phone: string;
  tradeNote: string;
  deliveryTags: string[];
  allowChat: boolean;
  businessStatus: ListingEditorBusinessStatus;
}

const listingObjectPrefix = 'ajo_living/listings';
const maxListingImageSize = 10 * 1024 * 1024;
const imageSlotCount = 4;
const donationDeliveryTag = 'donation_available';
const legacyDonationDeliveryTag = '可捐贈';

// 1. 建立表單初始狀態
const createInitialFormState = (): ListingEditorFormState => ({
  title: '',
  categoryCode: 'home_furniture',
  priceMode: 'fixed',
  price: 0,
  isDonation: false,
  condition: 'used_good',
  districtCode: 'hong_kong_island',
  visibility: 'public',
  communityId: '',
  communityName: '',
  summary: '',
  description: '',
  dimensionLength: '',
  dimensionWidth: '',
  dimensionHeight: '',
  dimensionWeight: '',
  phone: '',
  tradeNote: '',
  deliveryTags: [],
  allowChat: true,
  businessStatus: 'available',
});

// 2. 建立空圖片槽
const createEmptyImageSlots = (): EditorImageSlot[] =>
  Array.from({ length: imageSlotCount }, (_, index) => ({
    id: `slot-${index + 1}`,
    label: '',
    isCover: index === 0,
    uploading: false,
  }));

// 3. 讀取錯誤訊息
const readErrorMessage = (error: unknown, fallback: string): string =>
  axios.isAxiosError(error)
    ? error.response?.data?.message ?? fallback
    : fallback;

// 4. 標準化表單文字值
const normalizeFormText = (value: unknown): string =>
  String(value ?? '').trim();

// 5. 釋放本地圖片預覽 URL
const revokeImageSlotPreview = (slot: EditorImageSlot): void => {
  if (slot.objectUrl) {
    URL.revokeObjectURL(slot.objectUrl);
  }
};

// 6. 拆解既有尺寸文字
const parseDimensionText = (dimensionText: string): Pick<
  ListingEditorFormState,
  'dimensionLength' | 'dimensionWidth' | 'dimensionHeight' | 'dimensionWeight'
> => {
  const text = normalizeFormText(dimensionText);
  const [length = '', width = '', height = ''] = text.match(/\d+(?:\.\d+)?/g) ?? [];
  const weightMatch = text.match(/(?:weight|重量)\s*[:：]?\s*(\d+(?:\.\d+)?)/i);

  return {
    dimensionLength: length,
    dimensionWidth: width,
    dimensionHeight: height,
    dimensionWeight: weightMatch?.[1] ?? '',
  };
};

// 7. 組合尺寸文字
const buildDimensionText = (formState: ListingEditorFormState): string => {
  const length = normalizeFormText(formState.dimensionLength);
  const width = normalizeFormText(formState.dimensionWidth);
  const height = normalizeFormText(formState.dimensionHeight);
  const weight = normalizeFormText(formState.dimensionWeight);
  const items = [
    length ? `L ${length} cm` : '',
    width ? `W ${width} cm` : '',
    height ? `H ${height} cm` : '',
    weight ? `Weight ${weight} kg` : '',
  ].filter(Boolean);

  return items.join(' / ');
};

// 8. 組合香港電話號碼
const buildHongKongPhone = (phone: string): string => {
  const value = normalizeFormText(phone);
  if (!value) {
    return '';
  }
  if (value.startsWith('+')) {
    return value;
  }

  return `+852 ${value}`;
};

// 9. 過濾可捐贈標籤
const withoutDonationTag = (tags: string[]): string[] =>
  tags.filter((tag) => {
    const normalizedTag = normalizeFormText(tag);
    return normalizedTag !== donationDeliveryTag && normalizedTag !== legacyDonationDeliveryTag;
  });

// 10. 轉換交收標籤顯示文字
const resolveDeliveryTagLabel = (tag: string, translate: (key: string) => string): string => {
  const labelMap: Record<string, string> = {
    self_pickup: translate('marketplace.editor.deliveryTagSelfPickup'),
    door_delivery: translate('marketplace.editor.deliveryTagDoorDelivery'),
    free_post: translate('marketplace.editor.deliveryTagFreePost'),
    paid_post: translate('marketplace.editor.deliveryTagPaidPost'),
    face_check: translate('marketplace.editor.deliveryTagFaceCheck'),
    donation_available: translate('marketplace.editor.donationAvailable'),
    '自取': translate('marketplace.editor.deliveryTagSelfPickup'),
    '送貨上門': translate('marketplace.editor.deliveryTagDoorDelivery'),
    '免費郵寄': translate('marketplace.editor.deliveryTagFreePost'),
    '付費郵寄': translate('marketplace.editor.deliveryTagPaidPost'),
    '面對面驗貨': translate('marketplace.editor.deliveryTagFaceCheck'),
    '可捐贈': translate('marketplace.editor.donationAvailable'),
  };

  return labelMap[tag] ?? tag;
};

// 11.0 讀取 POS 大廈 ID
const getBuildingId = (item: PosBuilding): string =>
  String(item.building_id ?? item.id ?? '').trim();

// 11.1 讀取 POS 大廈名稱
const getBuildingName = (item: PosBuilding): string =>
  String(item.buildname_chi ?? item.buildname ?? item.name ?? getBuildingId(item)).trim();

// 11. 將 API 詳情同步到發布表單
const syncDetailToForm = (
  detail: SecondhandListingDetailResponse,
  formState: ListingEditorFormState,
): void => {
  const dimensionFields = parseDimensionText(detail.dimension_text);

  formState.title = detail.title;
  formState.summary = detail.summary;
  formState.description = detail.description;
  formState.categoryCode = detail.category_code as MarketplaceCategoryCode;
  formState.priceMode = detail.price_mode === 'free' ? 'fixed' : detail.price_mode as MarketplacePriceMode;
  formState.price = detail.price_hkd ?? 0;
  formState.isDonation = (
    detail.delivery_tags.includes(donationDeliveryTag) ||
    detail.delivery_tags.includes(legacyDonationDeliveryTag) ||
    detail.price_mode === 'free'
  );
  formState.condition = detail.condition_level as MarketplaceConditionCode;
  formState.districtCode = detail.district_code as MarketplaceDistrictCode;
  formState.visibility = detail.visibility_scope;
  formState.communityId = detail.community?.public_id ?? '';
  formState.communityName =
    detail.community?.name_zh?.trim() ||
    detail.community?.name_en?.trim() ||
    detail.community?.address_text?.trim() ||
    '';
  formState.dimensionLength = dimensionFields.dimensionLength;
  formState.dimensionWidth = dimensionFields.dimensionWidth;
  formState.dimensionHeight = dimensionFields.dimensionHeight;
  formState.dimensionWeight = dimensionFields.dimensionWeight;
  formState.tradeNote = detail.pickup_location_text;
  formState.deliveryTags = withoutDonationTag(detail.delivery_tags);
  formState.allowChat = detail.contact_summary.show_chat;
  formState.businessStatus = detail.business_status === 'sold' ? 'sold' : 'available';
};

// 12. 重置新增帖子表單
const resetFormState = (
  formState: ListingEditorFormState,
  imageSlots: { value: EditorImageSlot[] },
): void => {
  imageSlots.value.forEach(revokeImageSlotPreview);
  Object.assign(formState, createInitialFormState());
  imageSlots.value = createEmptyImageSlots();
};

// 13. 建立表單異動比對快照
const createEditorSnapshot = (
  formState: ListingEditorFormState,
  imageSlots: EditorImageSlot[],
): string => {
  const normalizedFormState: ListingEditorFormState = {
    ...formState,
    title: normalizeFormText(formState.title),
    summary: normalizeFormText(formState.summary),
    description: normalizeFormText(formState.description),
    dimensionLength: normalizeFormText(formState.dimensionLength),
    dimensionWidth: normalizeFormText(formState.dimensionWidth),
    dimensionHeight: normalizeFormText(formState.dimensionHeight),
    dimensionWeight: normalizeFormText(formState.dimensionWeight),
    phone: normalizeFormText(formState.phone),
    tradeNote: normalizeFormText(formState.tradeNote),
    deliveryTags: formState.deliveryTags.map(normalizeFormText).filter(Boolean).sort(),
  };
  const snapshot: EditorSnapshot = {
    formState: normalizedFormState,
    images: imageSlots
      .filter((slot) => Boolean(slot.mediaAssetId || slot.file))
      .map((slot) => ({
        mediaAssetId: slot.mediaAssetId ?? '',
        fileName: slot.fileName ?? slot.file?.name ?? '',
        fileSize: slot.file?.size ?? 0,
        fileType: slot.file?.type ?? '',
        isCover: slot.isCover,
      })),
  };

  return JSON.stringify(snapshot);
};

// 14. 建立帖子圖片 OSS 目錄
const buildListingObjectPrefix = (targetListingId: string): string => {
  const normalizedListingId = normalizeFormText(targetListingId);
  return normalizedListingId
    ? `${listingObjectPrefix}/${normalizedListingId}/`
    : `${listingObjectPrefix}/`;
};

// 15. 管理發布頁資料與動作
export const useMarketplaceListingEditorPage = (options: MarketplaceListingEditorOptions = {}) => {
  const route = useRoute();
  const router = useRouter();
  const { t } = useI18n();
  const feedbackStore = useFeedbackStore();
  const preferenceStore = usePreferenceStore();
  const sessionStore = useSessionStore();
  const useModalMode = options.useModalMode === true;
  const formState = reactive(createInitialFormState());
  const imageSlots = ref<EditorImageSlot[]>(createEmptyImageSlots());
  const listingId = ref(String(route.query.listing_id ?? ''));
  const publicationStatus = ref('');
  const isLoading = ref(false);
  const isSaving = ref(false);
  const isPublishing = ref(false);
  const isLeavePromptOpen = ref(false);
  const buildings = ref<PosBuilding[]>([]);
  const buildingsLoading = ref(false);
  const savedSnapshot = ref('');
  const isProgrammaticNavigation = ref(false);
  let resolveLeavePrompt: ((decision: EditorLeaveDecision) => void) | null = null;

  const isEditing = computed(() => normalizeFormText(listingId.value).length > 0);
  const isDraftListing = computed(() => !isEditing.value || publicationStatus.value === 'draft');

  const categoryOptions = computed<EditorOption<MarketplaceCategoryCode>[]>(() =>
    marketplaceCategories.map((category) => ({
      label: getMarketplaceOptionLabel(category, preferenceStore.locale),
      value: category.value,
    })),
  );

  const conditionOptions = computed<EditorOption<MarketplaceConditionCode>[]>(() =>
    marketplaceConditions.map((condition) => ({
      label: getMarketplaceOptionLabel(condition, preferenceStore.locale),
      value: condition.value,
    })),
  );

  const priceModeOptions = computed<EditorOption<MarketplacePriceMode>[]>(() =>
    marketplacePriceModes
      .filter((mode) => mode.value !== 'free')
      .map((mode) => ({
        label: getMarketplaceOptionLabel(mode, preferenceStore.locale),
        value: mode.value,
      })),
  );

  const areaOptions = computed<EditorOption<MarketplaceDistrictCode>[]>(() =>
    marketplaceDistricts.map((district) => ({
      label: getMarketplaceOptionLabel(district, preferenceStore.locale),
      value: district.value,
    })),
  );

  const visibilityOptions = computed<EditorOption<ListingEditorVisibility>[]>(() => [
    { label: t('marketplace.editor.publicListing'), value: 'public' },
    { label: t('marketplace.editor.buildingOnlyListing'), value: 'building_only' },
  ]);

  const currentCommunityOption = computed<EditorOption[]>(() => {
    const currentCommunity = sessionStore.me?.primary_community;
    const communityName =
      currentCommunity?.name_zh?.trim() ||
      currentCommunity?.name_en?.trim() ||
      currentCommunity?.address_text?.trim() ||
      '';

    if (!currentCommunity?.public_id) {
      return [];
    }

    return [{
      label: communityName || currentCommunity.public_id,
      value: currentCommunity.public_id,
    }];
  });

  const buildingOptions = computed<EditorOption[]>(() => {
    const selectedOption = formState.communityId
      ? [{
          label: normalizeFormText(formState.communityName) || formState.communityId,
          value: formState.communityId,
        }]
      : [];
    const posOptions = buildings.value
      .slice()
      .sort((left, right) => getBuildingName(left).localeCompare(getBuildingName(right), 'en', {
        numeric: true,
        sensitivity: 'base',
      }))
      .map((item) => ({
        label: getBuildingName(item),
        value: getBuildingId(item),
      }))
      .filter((option) => option.value.length > 0);
    const options = new Map<string, EditorOption>();

    [...currentCommunityOption.value, ...posOptions, ...selectedOption].forEach((option) => {
      if (!options.has(option.value)) {
        options.set(option.value, option);
      }
    });

    return [
      {
        label: buildingsLoading.value
          ? t('auth.residenceBuildingLoading')
          : t('marketplace.editor.visibilityCommunityPlaceholder'),
        value: '',
      },
      ...Array.from(options.values()),
    ];
  });

  const selectedCommunityName = computed(() =>
    buildingOptions.value.find((option) => option.value === formState.communityId)?.label ||
    formState.communityName,
  );

  const businessStatusOptions = computed<EditorOption<ListingEditorBusinessStatus>[]>(() => [
    { label: t('marketplace.editor.availableStatus'), value: 'available' },
    { label: t('marketplace.editor.soldStatus'), value: 'sold' },
  ]);

  const coverImage = computed(() =>
    imageSlots.value.find((slot) => slot.isCover && slot.url) ?? imageSlots.value.find((slot) => slot.url),
  );

  const selectedCategoryLabel = computed(() =>
    getMarketplaceCategoryLabel(formState.categoryCode, preferenceStore.locale),
  );

  const selectedAreaLabel = computed(() =>
    getMarketplaceDistrictLabel(formState.districtCode, preferenceStore.locale),
  );

  const selectedConditionLabel = computed(() =>
    getMarketplaceConditionLabel(formState.condition, preferenceStore.locale),
  );

  const previewTitle = computed(() =>
    normalizeFormText(formState.title) || t('marketplace.editor.previewFallbackTitle'),
  );

  const previewPrice = computed(() =>
    formatPrice(Number.isFinite(formState.price) ? formState.price : 0, preferenceStore.locale),
  );

  const previewTagLabels = computed(() => [
    selectedCategoryLabel.value,
    selectedConditionLabel.value,
    selectedAreaLabel.value,
    ...formState.deliveryTags.map((tag) => resolveDeliveryTagLabel(tag, t)),
    formState.isDonation ? t('marketplace.editor.donationAvailable') : '',
  ].filter((label) => normalizeFormText(label).length > 0));

  const selectedImageCount = computed(() =>
    imageSlots.value.filter((slot) => Boolean(slot.mediaAssetId || slot.file)).length,
  );

  const hasValidPrice = computed(() => Number(formState.price) > 0);

  const hasValidContact = computed(() =>
    formState.allowChat || normalizeFormText(formState.phone).length > 0,
  );

  const hasValidVisibility = computed(() =>
    formState.visibility === 'public' ||
    normalizeFormText(formState.communityId).length > 0,
  );

  const checklist = computed<EditorChecklistItem[]>(() => [
    { label: t('marketplace.editor.titleReady'), complete: normalizeFormText(formState.title).length > 0 },
    { label: t('marketplace.editor.categoryReady'), complete: Boolean(formState.categoryCode) },
    { label: t('marketplace.editor.priceReady'), complete: hasValidPrice.value },
    { label: t('marketplace.editor.contactReady'), complete: hasValidContact.value && hasValidVisibility.value },
    { label: t('marketplace.editor.imageReady'), complete: selectedImageCount.value > 0 },
  ]);

  const readyToSaveDraft = computed(() =>
    normalizeFormText(formState.title).length > 0 &&
    Boolean(formState.categoryCode) &&
    hasValidVisibility.value,
  );

  const readyToPublish = computed(() => checklist.value.every((item) => item.complete));
  const chargeCost = computed(() => resolveWalletChargeCost('secondhand'));
  const draftChargeCost = computed(() => resolveWalletDraftChargeCost('secondhand'));
  const walletBalance = computed(() => sessionStore.me?.ajo_balance ?? 0);
  const formatPoints = (value: number): string =>
    formatAjoPoints(value, t('common.brand.pointsName'), preferenceStore.locale);
  const chargeHint = computed(() => {
    const balanceText = `${t('marketplace.editor.walletBalance')} ${formatPoints(walletBalance.value)}`;

    if (!isDraftListing.value) {
      return `${t('marketplace.editor.editChargeHint')} ${formatPoints(chargeCost.value)} · ${balanceText}`;
    }

    return `${t('marketplace.editor.publishChargeHint')} ${formatPoints(chargeCost.value)} · ${t('marketplace.editor.draftChargeHint')} ${formatPoints(draftChargeCost.value)} · ${balanceText}`;
  });
  const canAffordDraftSave = computed(() =>
    !isDraftListing.value || walletBalance.value >= draftChargeCost.value,
  );
  const canAffordPublish = computed(() => walletBalance.value >= chargeCost.value);

  const workflowSteps = computed<EditorWorkflowStep[]>(() => [
    {
      label: t('marketplace.editor.stepContent'),
      description: t('marketplace.editor.stepContentDescription'),
      status: normalizeFormText(formState.title) && normalizeFormText(formState.description) ? 'done' : 'current',
    },
    {
      label: t('marketplace.editor.stepImages'),
      description: t('marketplace.editor.stepImagesDescription'),
      status: selectedImageCount.value > 0 ? 'done' : 'current',
    },
    {
      label: t('marketplace.editor.stepPreview'),
      description: t('marketplace.editor.stepPreviewDescription'),
      status: readyToPublish.value ? 'done' : 'idle',
    },
  ]);

  const metrics = computed<EditorMetric[]>(() => [
    {
      label: t('marketplace.editor.metricPhotos'),
      value: `${selectedImageCount.value}/${imageSlotCount}`,
    },
    {
      label: t('marketplace.editor.metricFields'),
      value: `${checklist.value.filter((item) => item.complete).length}/${checklist.value.length}`,
    },
  ]);

  const currentSnapshot = computed(() => createEditorSnapshot(formState, imageSlots.value));
  const hasUnsavedChanges = computed(() =>
    savedSnapshot.value.length > 0 &&
    currentSnapshot.value !== savedSnapshot.value,
  );

  // 15.1 更新離開提示基準
  const markCurrentStateSaved = (): void => {
    savedSnapshot.value = currentSnapshot.value;
  };

  // 15.1.1 套用會員預設屋苑
  const applyDefaultCommunity = (): void => {
    if (normalizeFormText(formState.communityId)) {
      return;
    }

    const currentCommunity = sessionStore.me?.primary_community;
    if (!currentCommunity?.public_id) {
      return;
    }

    formState.communityId = currentCommunity.public_id;
    formState.communityName =
      currentCommunity.name_zh?.trim() ||
      currentCommunity.name_en?.trim() ||
      currentCommunity.address_text?.trim() ||
      currentCommunity.public_id;
  };

  // 15.1.2 載入 POS 大廈清單
  const loadBuildings = async (): Promise<void> => {
    buildingsLoading.value = true;
    try {
      buildings.value = await fetchPosBuildings();
    } catch {
      buildings.value = [];
    } finally {
      buildingsLoading.value = false;
    }
  };

  // 15.2 打開離開確認彈窗
  const requestLeaveDecision = (): Promise<EditorLeaveDecision> => {
    isLeavePromptOpen.value = true;

    return new Promise((resolve) => {
      resolveLeavePrompt = resolve;
    });
  };

  // 15.3 回應離開確認彈窗
  const handleLeavePromptDecision = (decision: EditorLeaveDecision): void => {
    isLeavePromptOpen.value = false;
    resolveLeavePrompt?.(decision);
    resolveLeavePrompt = null;
  };

  // 15.4 確認是否允許離開編輯頁
  const confirmLeaveEditor = async (): Promise<boolean> => {
    if (isProgrammaticNavigation.value || !hasUnsavedChanges.value) {
      return true;
    }

    const decision = await requestLeaveDecision();
    if (decision === 'stay') {
      return false;
    }
    if (decision === 'discard') {
      return true;
    }

    const savedListingId = await saveDraft({ chargeDraft: true, uploadImages: true });
    if (savedListingId.length > 0) {
      await options.onSaved?.();
      return true;
    }

    return false;
  };

  // 15.4.1 嘗試關閉彈窗編輯器
  const requestCloseEditor = async (): Promise<void> => {
    const canClose = await confirmLeaveEditor();
    if (canClose) {
      options.onClose?.();
    }
  };

  // 15.5 處理瀏覽器關閉或重新整理
  const handleBeforeUnload = (event: BeforeUnloadEvent): void => {
    if (!hasUnsavedChanges.value) {
      return;
    }

    event.preventDefault();
    event.returnValue = '';
  };

  // 15.6 讀取編輯頁詳情
  const loadListingDetail = async (): Promise<void> => {
    if (!listingId.value) {
      return;
    }

    isLoading.value = true;

    try {
      const { data } = await fetchSecondhandListingDetail(listingId.value);
      publicationStatus.value = data.data.publication_status;
      syncDetailToForm(data.data, formState);
      if (formState.visibility === 'building_only') {
        applyDefaultCommunity();
      }
      imageSlots.value.forEach(revokeImageSlotPreview);
      const detailSlots = data.data.images.map<EditorImageSlot>((image, index) => ({
        id: image.media_asset_id,
        label: `${t('marketplace.editor.imageSlot')} ${index + 1}`,
        url: image.url,
        mediaAssetId: image.media_asset_id,
        isCover: image.is_cover || index === 0,
        uploading: false,
      }));
      imageSlots.value = [
        ...detailSlots,
        ...createEmptyImageSlots(),
      ].slice(0, imageSlotCount).map((slot, index) => ({
        ...slot,
        id: slot.mediaAssetId || `slot-${index + 1}`,
      }));
      markCurrentStateSaved();
    } catch (error) {
      feedbackStore.pushToast(readErrorMessage(error, t('marketplace.editor.loadError')), 'error');
    } finally {
      isLoading.value = false;
    }
  };

  // 15.7 設定封面圖片
  const selectCoverImage = (slotId: string): void => {
    imageSlots.value = imageSlots.value.map((slot) => ({
      ...slot,
      isCover: slot.id === slotId && Boolean(slot.mediaAssetId || slot.file),
    }));
  };

  // 15.8 校驗圖片檔案
  const validateImageFile = (file: File): boolean => {
    if (!file.type.startsWith('image/')) {
      feedbackStore.pushToast(t('marketplace.editor.imageFileInvalid'), 'error');
      return false;
    }

    if (file.size > maxListingImageSize) {
      feedbackStore.pushToast(t('marketplace.editor.imageFileTooLarge'), 'error');
      return false;
    }

    return true;
  };

  // 15.9 將圖片暫存在頁面並建立本地預覽
  const stageImageFile = (file: File, slotId: string): void => {
    if (!validateImageFile(file)) {
      return;
    }

    const targetSlot = imageSlots.value.find((slot) => slot.id === slotId);
    if (!targetSlot) {
      return;
    }

    const objectUrl = URL.createObjectURL(file);
    const hasCover = imageSlots.value.some((slot) => slot.isCover && (slot.mediaAssetId || slot.file));
    revokeImageSlotPreview(targetSlot);

    imageSlots.value = imageSlots.value.map((slot) =>
      slot.id === slotId
        ? {
            ...slot,
            label: file.name,
            url: objectUrl,
            file,
            fileName: file.name,
            objectUrl,
            isCover: slot.isCover || !hasCover,
            uploading: false,
          }
        : slot,
    );
  };

  // 15.10 上傳單個暫存圖片到 OSS
  const uploadImageSlot = async (slotId: string, targetListingId: string): Promise<void> => {
    const targetSlot = imageSlots.value.find((slot) => slot.id === slotId);
    if (!targetSlot?.file) {
      return;
    }

    const file = targetSlot.file;
    imageSlots.value = imageSlots.value.map((slot) =>
      slot.id === slotId ? { ...slot, uploading: true } : slot,
    );

    try {
      const presignResponse = await createUploadPresign({
        file_name: file.name,
        mime_type: file.type,
        file_size: file.size,
        object_prefix: buildListingObjectPrefix(targetListingId),
      });
      const presign = presignResponse.data.data;
      const uploadResponse = await fetch(presign.upload_url, {
        method: 'PUT',
        headers: buildUploadHeaders(presign.headers, file.type),
        body: file,
      });

      if (!uploadResponse.ok) {
        throw new Error(`listing image upload failed with status ${uploadResponse.status}`);
      }

      const completeResponse = await completeUpload({
        object_key: presign.object_key,
        mime_type: file.type,
        file_size: file.size,
      });
      const mediaAsset = completeResponse.data.data;
      updateImageSlot(slotId, mediaAsset, file.name);
    } catch (error) {
      imageSlots.value = imageSlots.value.map((slot) =>
        slot.id === slotId ? { ...slot, uploading: false } : slot,
      );
      throw error;
    }
  };

  // 15.11 將上傳結果回寫圖片槽
  const updateImageSlot = (slotId: string, mediaAsset: MediaAssetResponse, fileName: string): void => {
    const hasCover = imageSlots.value.some((slot) => slot.isCover && slot.mediaAssetId);

    imageSlots.value = imageSlots.value.map((slot) => {
      if (slot.id !== slotId) {
        return slot;
      }

      revokeImageSlotPreview(slot);

      return {
        ...slot,
        id: mediaAsset.media_asset_id,
        label: fileName,
        url: mediaAsset.url,
        mediaAssetId: mediaAsset.media_asset_id,
        file: undefined,
        fileName,
        objectUrl: undefined,
        isCover: slot.isCover || !hasCover,
        uploading: false,
      };
    });
  };

  // 15.12 處理圖片 input
  const handleImageFileChange = (event: Event, slotId: string): void => {
    const input = event.target as HTMLInputElement;
    const file = input.files?.[0];
    input.value = '';

    if (file) {
      stageImageFile(file, slotId);
    }
  };

  // 15.13 批量暫存圖片
  const stageImageFiles = (files: File[]): void => {
    for (const file of files) {
      const nextSlot = imageSlots.value.find((slot) => !slot.mediaAssetId && !slot.file && !slot.uploading);
      if (!nextSlot) {
        feedbackStore.pushToast(t('marketplace.editor.imageSlotFull'), 'error');
        return;
      }

      stageImageFile(file, nextSlot.id);
    }
  };

  // 15.14 處理多圖 input
  const handleImageFilesChange = (event: Event): void => {
    const input = event.target as HTMLInputElement;
    const files = Array.from(input.files ?? []);
    input.value = '';

    stageImageFiles(files);
  };

  // 15.15 處理拖拽圖片
  const handleDroppedImageFiles = (files: File[]): void => {
    stageImageFiles(files);
  };

  // 15.16 上傳所有待發布圖片
  const uploadPendingImages = async (targetListingId: string): Promise<void> => {
    const pendingSlots = imageSlots.value.filter((slot) => Boolean(slot.file));
    for (const slot of pendingSlots) {
      await uploadImageSlot(slot.id, targetListingId);
    }
  };

  // 15.17 移除圖片
  const removeImageSlot = (slotId: string): void => {
    const removedSlot = imageSlots.value.find((slot) => slot.id === slotId);
    if (removedSlot) {
      revokeImageSlotPreview(removedSlot);
    }

    imageSlots.value = imageSlots.value.map((slot, index) =>
      slot.id === slotId
        ? {
            id: `slot-${index + 1}`,
            label: '',
            isCover: false,
            uploading: false,
          }
        : slot,
    );

    if (!imageSlots.value.some((slot) => slot.isCover && (slot.mediaAssetId || slot.file))) {
      const firstImageIndex = imageSlots.value.findIndex((slot) => Boolean(slot.mediaAssetId || slot.file));
      if (firstImageIndex >= 0) {
        imageSlots.value[firstImageIndex] = {
          ...imageSlots.value[firstImageIndex],
          isCover: true,
        };
      }
    }
  };

  // 15.18 移動圖片排序
  const moveImageSlot = (slotId: string, offset: -1 | 1): void => {
    const currentIndex = imageSlots.value.findIndex((slot) => slot.id === slotId);
    const nextIndex = currentIndex + offset;
    if (currentIndex < 0 || nextIndex < 0 || nextIndex >= imageSlots.value.length) {
      return;
    }

    const nextSlots = [...imageSlots.value];
    const [current] = nextSlots.splice(currentIndex, 1);
    nextSlots.splice(nextIndex, 0, current);
    imageSlots.value = nextSlots;
  };

  // 15.19 建立提交 payload
  const buildPayload = (): UpsertSecondhandListingPayload => {
    const deliveryTags = [
      ...formState.deliveryTags,
      formState.isDonation ? donationDeliveryTag : '',
    ]
      .map(normalizeFormText)
      .filter(Boolean);
    const communityId = normalizeFormText(formState.communityId);
    const communityName = communityId ? normalizeFormText(selectedCommunityName.value) : '';
    const phone = buildHongKongPhone(normalizeFormText(formState.phone));
    const allowPhone = phone.length > 0;

    return {
      title: normalizeFormText(formState.title),
      summary: normalizeFormText(formState.summary),
      description: normalizeFormText(formState.description),
      district_code: formState.districtCode,
      community_id: communityId,
      community_name: communityName,
      publisher_identity_type: sessionStore.me?.publisher_identity_type || 'owner',
      category_code: formState.categoryCode,
      price_mode: formState.priceMode,
      price_hkd: Number(formState.price),
      condition_level: formState.condition,
      dimension_text: buildDimensionText(formState),
      pickup_region_code: formState.districtCode,
      pickup_location_text: normalizeFormText(formState.tradeNote),
      delivery_tags: deliveryTags,
      visibility_scope: formState.visibility,
      contact_method: formState.allowChat ? 'chat_or_whatsapp' : 'phone',
      business_status: formState.businessStatus,
      images: imageSlots.value
        .filter((slot) => Boolean(slot.mediaAssetId))
        .map((slot, index) => ({
          media_asset_id: slot.mediaAssetId ?? '',
          sort_order: index + 1,
          is_cover: slot.isCover || index === 0,
        })),
      contact: {
        phone,
        whatsapp: '',
        email: sessionStore.me?.email ?? '',
        show_phone: allowPhone,
        show_whatsapp: false,
        show_chat: formState.allowChat,
        show_inquiry_form: false,
      },
    };
  };

  // 15.20 儲存草稿
  const saveDraft = async (options: { chargeDraft?: boolean; uploadImages?: boolean } = {}): Promise<string> => {
    if (!readyToSaveDraft.value) {
      feedbackStore.pushToast(t('marketplace.editor.saveBlocked'), 'error');
      return '';
    }

    isSaving.value = true;

    try {
      const wasEditing = isEditing.value;
      const shouldUploadImages = options.uploadImages ?? wasEditing;
      const shouldChargeDraft = options.chargeDraft === true && isDraftListing.value;
      const shouldChargeOnCreate = shouldChargeDraft && !shouldUploadImages;
      if (shouldChargeDraft && !canAffordDraftSave.value) {
        feedbackStore.pushToast(t('marketplace.editor.draftChargeInsufficient'), 'error');
        return '';
      }

      if (!listingId.value) {
        const createResponse = await createSecondhandListing(buildPayload(), {
          charge_draft: shouldChargeOnCreate,
        });
        listingId.value = createResponse.data.data.listing_id;
        publicationStatus.value = createResponse.data.data.publication_status;
        if (typeof createResponse.data.data.points_balance_after === 'number') {
          await sessionStore.loadCurrentUser();
        }

        if (!shouldUploadImages) {
          markCurrentStateSaved();
          feedbackStore.pushToast(t('marketplace.editor.draftSaved'), 'success');
          return listingId.value;
        }
      }

      if (shouldUploadImages) {
        try {
          await uploadPendingImages(listingId.value);
        } catch (error) {
          feedbackStore.pushToast(readErrorMessage(error, t('marketplace.editor.imageUploadError')), 'error');
          return '';
        }
      }

      const payload = buildPayload();
      const response = await updateSecondhandListing(listingId.value, payload, {
        charge_draft: shouldChargeDraft && (wasEditing || !shouldChargeOnCreate),
      });
      listingId.value = response.data.data.listing_id;
      publicationStatus.value = response.data.data.publication_status;
      if (typeof response.data.data.points_balance_after === 'number') {
        await sessionStore.loadCurrentUser();
      }
      markCurrentStateSaved();
      feedbackStore.pushToast(t(wasEditing ? 'marketplace.editor.updateSaved' : 'marketplace.editor.draftSaved'), 'success');

      return listingId.value;
    } catch (error) {
      feedbackStore.pushToast(readErrorMessage(error, t('marketplace.editor.saveError')), 'error');
      return '';
    } finally {
      isSaving.value = false;
    }
  };

  // 15.21 儲存並立即發布
  const submitListing = async (): Promise<void> => {
    if (!readyToPublish.value) {
      feedbackStore.pushToast(t('marketplace.editor.publishBlocked'), 'error');
      return;
    }

    isPublishing.value = true;

    try {
      if (!canAffordPublish.value) {
        feedbackStore.pushToast(t('marketplace.editor.publishChargeInsufficient'), 'error');
        return;
      }
      const savedListingId = await saveDraft({ uploadImages: true });
      if (!savedListingId) {
        return;
      }

      await publishSecondhandListing(savedListingId);
      await sessionStore.loadCurrentUser();
      feedbackStore.pushToast(t('marketplace.editor.publishSuccess'), 'success');
      isProgrammaticNavigation.value = true;
      await options.onSaved?.();
      if (useModalMode) {
        options.onClose?.();
        return;
      }
      await router.push({ path: '/member', query: { tab: 'listings' } });
    } catch (error) {
      feedbackStore.pushToast(readErrorMessage(error, t('marketplace.editor.publishError')), 'error');
    } finally {
      isPublishing.value = false;
    }
  };

  // 15.22 儲存並返回我的帖子列表
  const saveAndBackToList = async (): Promise<void> => {
    const savedListingId = await saveDraft({ chargeDraft: true, uploadImages: true });
    if (savedListingId) {
      isProgrammaticNavigation.value = true;
      await options.onSaved?.();
      if (useModalMode) {
        options.onClose?.();
        return;
      }
      await router.push({ path: '/member', query: { tab: 'listings' } });
    }
  };

  if (!useModalMode) {
    onBeforeRouteLeave(() => confirmLeaveEditor());
  }

  onMounted(() => {
    window.addEventListener('beforeunload', handleBeforeUnload);
    if (!listingId.value) {
      applyDefaultCommunity();
      markCurrentStateSaved();
    }
    void loadBuildings();
  });

  watch(
    () => formState.visibility,
    (value) => {
      if (value === 'building_only') {
        applyDefaultCommunity();
      }
    },
  );

  watch(
    () => formState.communityId,
    (value) => {
      formState.communityName = value ? selectedCommunityName.value : '';
    },
  );

  watch(
    () => route.query.listing_id,
    (value) => {
      listingId.value = String(value ?? '');
      if (listingId.value) {
        void loadListingDetail();
        return;
      }

      resetFormState(formState, imageSlots);
      applyDefaultCommunity();
      publicationStatus.value = '';
      markCurrentStateSaved();
    },
    { immediate: true },
  );

  onBeforeUnmount(() => {
    window.removeEventListener('beforeunload', handleBeforeUnload);
    imageSlots.value.forEach(revokeImageSlotPreview);
  });

  return {
    areaOptions,
    buildingOptions,
    buildingsLoading,
    businessStatusOptions,
    categoryOptions,
    checklist,
    chargeHint,
    conditionOptions,
    coverImage,
    formState,
    handleLeavePromptDecision,
    handleDroppedImageFiles,
    handleImageFileChange,
    handleImageFilesChange,
    requestCloseEditor,
    imageSlots,
    isEditing,
    isLeavePromptOpen,
    isLoading,
    isPublishing,
    isSaving,
    metrics,
    moveImageSlot,
    previewPrice,
    previewTagLabels,
    previewTitle,
    priceModeOptions,
    readyToPublish,
    readyToSaveDraft,
    removeImageSlot,
    saveDraft,
    saveAndBackToList,
    selectCoverImage,
    selectedAreaLabel,
    selectedCategoryLabel,
    selectedConditionLabel,
    submitListing,
    visibilityOptions,
    workflowSteps,
  };
};

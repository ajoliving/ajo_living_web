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
} from '@/httpapis/secondhand-listings';
import { completeUpload, createUploadPresign } from '@/httpapis/uploads';
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
} from '@/constants/marketplace';
import type {
  MediaAssetResponse,
  SecondhandListingDetailResponse,
  UpsertSecondhandListingPayload,
} from '@/model/marketplace';
import { useFeedbackStore } from '@/stores/feedback';
import { usePreferenceStore } from '@/stores/preferences';
import { useSessionStore } from '@/stores/session';
import { formatPrice } from '@/utils/format';
import { buildUploadHeaders } from '@/utils/upload';
import { formatAjoPoints, resolveWalletChargeCost, resolveWalletDraftChargeCost } from '@/utils/wallet';

export type ListingEditorVisibility = 'public' | 'building_only';
export type ListingEditorBusinessStatus = 'available' | 'sold';

type EditorStepStatus = 'done' | 'current' | 'idle';
type EditorLeaveDecision = 'save' | 'discard' | 'stay';

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
  districtCode: 'eastern',
  visibility: 'public',
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

// 4. 釋放本地圖片預覽 URL
const revokeImageSlotPreview = (slot: EditorImageSlot): void => {
  if (slot.objectUrl) {
    URL.revokeObjectURL(slot.objectUrl);
  }
};

// 5. 拆解既有尺寸文字
const parseDimensionText = (dimensionText: string): Pick<
  ListingEditorFormState,
  'dimensionLength' | 'dimensionWidth' | 'dimensionHeight' | 'dimensionWeight'
> => {
  const text = dimensionText.trim();
  const [length = '', width = '', height = ''] = text.match(/\d+(?:\.\d+)?/g) ?? [];
  const weightMatch = text.match(/(?:weight|重量)\s*[:：]?\s*(\d+(?:\.\d+)?)/i);

  return {
    dimensionLength: length,
    dimensionWidth: width,
    dimensionHeight: height,
    dimensionWeight: weightMatch?.[1] ?? '',
  };
};

// 6. 組合尺寸文字
const buildDimensionText = (formState: ListingEditorFormState): string => {
  const items = [
    formState.dimensionLength.trim() ? `L ${formState.dimensionLength.trim()} cm` : '',
    formState.dimensionWidth.trim() ? `W ${formState.dimensionWidth.trim()} cm` : '',
    formState.dimensionHeight.trim() ? `H ${formState.dimensionHeight.trim()} cm` : '',
    formState.dimensionWeight.trim() ? `Weight ${formState.dimensionWeight.trim()} kg` : '',
  ].filter(Boolean);

  return items.join(' / ');
};

// 7. 組合香港電話號碼
const buildHongKongPhone = (phone: string): string => {
  const value = phone.trim();
  if (!value) {
    return '';
  }
  if (value.startsWith('+')) {
    return value;
  }

  return `+852 ${value}`;
};

// 8. 過濾可捐贈標籤
const withoutDonationTag = (tags: string[]): string[] =>
  tags.filter((tag) => {
    const normalizedTag = tag.trim();
    return normalizedTag !== donationDeliveryTag && normalizedTag !== legacyDonationDeliveryTag;
  });

// 9. 轉換交收標籤顯示文字
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

// 10. 將 API 詳情同步到發布表單
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
  formState.dimensionLength = dimensionFields.dimensionLength;
  formState.dimensionWidth = dimensionFields.dimensionWidth;
  formState.dimensionHeight = dimensionFields.dimensionHeight;
  formState.dimensionWeight = dimensionFields.dimensionWeight;
  formState.tradeNote = detail.pickup_location_text;
  formState.deliveryTags = withoutDonationTag(detail.delivery_tags);
  formState.allowChat = detail.contact_summary.show_chat;
  formState.businessStatus = detail.business_status === 'sold' ? 'sold' : 'available';
};

// 11. 重置新增帖子表單
const resetFormState = (
  formState: ListingEditorFormState,
  imageSlots: { value: EditorImageSlot[] },
): void => {
  imageSlots.value.forEach(revokeImageSlotPreview);
  Object.assign(formState, createInitialFormState());
  imageSlots.value = createEmptyImageSlots();
};

// 12. 建立表單異動比對快照
const createEditorSnapshot = (
  formState: ListingEditorFormState,
  imageSlots: EditorImageSlot[],
): string => {
  const normalizedFormState: ListingEditorFormState = {
    ...formState,
    title: formState.title.trim(),
    summary: formState.summary.trim(),
    description: formState.description.trim(),
    dimensionLength: formState.dimensionLength.trim(),
    dimensionWidth: formState.dimensionWidth.trim(),
    dimensionHeight: formState.dimensionHeight.trim(),
    dimensionWeight: formState.dimensionWeight.trim(),
    phone: formState.phone.trim(),
    tradeNote: formState.tradeNote.trim(),
    deliveryTags: formState.deliveryTags.map((tag) => tag.trim()).filter(Boolean).sort(),
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

// 13. 建立帖子圖片 OSS 目錄
const buildListingObjectPrefix = (targetListingId: string): string => {
  const normalizedListingId = targetListingId.trim();
  return normalizedListingId
    ? `${listingObjectPrefix}/${normalizedListingId}/`
    : `${listingObjectPrefix}/`;
};

// 14. 管理發布頁資料與動作
export const useMarketplaceListingEditorPage = () => {
  const route = useRoute();
  const router = useRouter();
  const { t } = useI18n();
  const feedbackStore = useFeedbackStore();
  const preferenceStore = usePreferenceStore();
  const sessionStore = useSessionStore();
  const formState = reactive(createInitialFormState());
  const imageSlots = ref<EditorImageSlot[]>(createEmptyImageSlots());
  const listingId = ref(String(route.params.listingId ?? ''));
  const publicationStatus = ref('');
  const isLoading = ref(false);
  const isSaving = ref(false);
  const isPublishing = ref(false);
  const isLeavePromptOpen = ref(false);
  const savedSnapshot = ref('');
  const isProgrammaticNavigation = ref(false);
  let resolveLeavePrompt: ((decision: EditorLeaveDecision) => void) | null = null;

  const isEditing = computed(() => listingId.value.trim().length > 0);
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
    formState.title.trim() || t('marketplace.editor.previewFallbackTitle'),
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
  ].filter((label) => label.trim().length > 0));

  const selectedImageCount = computed(() =>
    imageSlots.value.filter((slot) => Boolean(slot.mediaAssetId || slot.file)).length,
  );

  const hasValidPrice = computed(() => Number(formState.price) > 0);

  const hasValidContact = computed(() =>
    formState.allowChat || formState.phone.trim().length > 0,
  );

  const hasValidVisibility = computed(() =>
    formState.visibility === 'public' ||
    sessionStore.currentUser.primary_community.public_id.trim().length > 0,
  );

  const checklist = computed<EditorChecklistItem[]>(() => [
    { label: t('marketplace.editor.titleReady'), complete: formState.title.trim().length > 0 },
    { label: t('marketplace.editor.categoryReady'), complete: Boolean(formState.categoryCode) },
    { label: t('marketplace.editor.priceReady'), complete: hasValidPrice.value },
    { label: t('marketplace.editor.contactReady'), complete: hasValidContact.value && hasValidVisibility.value },
    { label: t('marketplace.editor.imageReady'), complete: selectedImageCount.value > 0 },
  ]);

  const readyToSaveDraft = computed(() =>
    formState.title.trim().length > 0 &&
    Boolean(formState.categoryCode),
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
      status: formState.title.trim() && formState.description.trim() ? 'done' : 'current',
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

  // 14.1 更新離開提示基準
  const markCurrentStateSaved = (): void => {
    savedSnapshot.value = currentSnapshot.value;
  };

  // 14.2 打開離開確認彈窗
  const requestLeaveDecision = (): Promise<EditorLeaveDecision> => {
    isLeavePromptOpen.value = true;

    return new Promise((resolve) => {
      resolveLeavePrompt = resolve;
    });
  };

  // 14.3 回應離開確認彈窗
  const handleLeavePromptDecision = (decision: EditorLeaveDecision): void => {
    isLeavePromptOpen.value = false;
    resolveLeavePrompt?.(decision);
    resolveLeavePrompt = null;
  };

  // 14.4 確認是否允許離開編輯頁
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
    return savedListingId.length > 0;
  };

  // 14.5 處理瀏覽器關閉或重新整理
  const handleBeforeUnload = (event: BeforeUnloadEvent): void => {
    if (!hasUnsavedChanges.value) {
      return;
    }

    event.preventDefault();
    event.returnValue = '';
  };

  // 14.4 讀取編輯頁詳情
  const loadListingDetail = async (): Promise<void> => {
    if (!listingId.value) {
      return;
    }

    isLoading.value = true;

    try {
      const { data } = await fetchSecondhandListingDetail(listingId.value);
      publicationStatus.value = data.data.publication_status;
      syncDetailToForm(data.data, formState);
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

  // 14.5 設定封面圖片
  const selectCoverImage = (slotId: string): void => {
    imageSlots.value = imageSlots.value.map((slot) => ({
      ...slot,
      isCover: slot.id === slotId && Boolean(slot.mediaAssetId || slot.file),
    }));
  };

  // 14.6 校驗圖片檔案
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

  // 14.7 將圖片暫存在頁面並建立本地預覽
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

  // 14.8 上傳單個暫存圖片到 OSS
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

  // 14.9 將上傳結果回寫圖片槽
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

  // 14.10 處理圖片 input
  const handleImageFileChange = (event: Event, slotId: string): void => {
    const input = event.target as HTMLInputElement;
    const file = input.files?.[0];
    input.value = '';

    if (file) {
      stageImageFile(file, slotId);
    }
  };

  // 14.11 批量暫存圖片
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

  // 14.12 處理多圖 input
  const handleImageFilesChange = (event: Event): void => {
    const input = event.target as HTMLInputElement;
    const files = Array.from(input.files ?? []);
    input.value = '';

    stageImageFiles(files);
  };

  // 14.13 處理拖拽圖片
  const handleDroppedImageFiles = (files: File[]): void => {
    stageImageFiles(files);
  };

  // 14.14 上傳所有待發布圖片
  const uploadPendingImages = async (targetListingId: string): Promise<void> => {
    const pendingSlots = imageSlots.value.filter((slot) => Boolean(slot.file));
    for (const slot of pendingSlots) {
      await uploadImageSlot(slot.id, targetListingId);
    }
  };

  // 14.15 移除圖片
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

  // 14.16 移動圖片排序
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

  // 14.17 建立提交 payload
  const buildPayload = (): UpsertSecondhandListingPayload => {
    const deliveryTags = [
      ...formState.deliveryTags,
      formState.isDonation ? donationDeliveryTag : '',
    ]
      .map((item) => item.trim())
      .filter(Boolean);
    const primaryCommunityId =
      formState.visibility === 'building_only'
        ? sessionStore.currentUser.primary_community.public_id
        : sessionStore.currentUser.primary_community.public_id || '';
    const phone = buildHongKongPhone(formState.phone);
    const allowPhone = phone.length > 0;

    return {
      title: formState.title.trim(),
      summary: formState.summary.trim(),
      description: formState.description.trim(),
      district_code: formState.districtCode,
      community_id: primaryCommunityId,
      publisher_identity_type: sessionStore.me?.publisher_identity_type || 'owner',
      category_code: formState.categoryCode,
      price_mode: formState.priceMode,
      price_hkd: Number(formState.price),
      condition_level: formState.condition,
      dimension_text: buildDimensionText(formState),
      pickup_region_code: formState.districtCode,
      pickup_location_text: formState.tradeNote.trim(),
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

  // 14.18 儲存草稿
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
      if (shouldChargeDraft && !canAffordDraftSave.value) {
        feedbackStore.pushToast(t('marketplace.editor.draftChargeInsufficient'), 'error');
        return '';
      }

      if (!listingId.value) {
        const createResponse = await createSecondhandListing(buildPayload(), {
          charge_draft: shouldChargeDraft,
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
        charge_draft: shouldChargeDraft && wasEditing,
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

  // 14.19 儲存並立即發布
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
      await router.push('/account/marketplace/my/listings');
    } catch (error) {
      feedbackStore.pushToast(readErrorMessage(error, t('marketplace.editor.publishError')), 'error');
    } finally {
      isPublishing.value = false;
    }
  };

  // 14.20 儲存並返回我的帖子列表
  const saveAndBackToList = async (): Promise<void> => {
    const savedListingId = await saveDraft({ chargeDraft: true, uploadImages: true });
    if (savedListingId) {
      isProgrammaticNavigation.value = true;
      await router.push('/account/marketplace/my/listings');
    }
  };

  onBeforeRouteLeave(() => confirmLeaveEditor());

  onMounted(() => {
    window.addEventListener('beforeunload', handleBeforeUnload);
  });

  watch(
    () => route.params.listingId,
    (value) => {
      listingId.value = String(value ?? '');
      if (listingId.value) {
        void loadListingDetail();
        return;
      }

      resetFormState(formState, imageSlots);
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

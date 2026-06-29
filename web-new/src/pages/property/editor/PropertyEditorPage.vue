<!--
 * 物業發布編輯頁。
 * 1. 根據頻道建立或更新樓盤放售與服務式住宅草稿。
 * 2. 支援圖片上傳、草稿保存與發布。
 * 3. 支援會員中心內嵌彈窗使用，處理未儲存離開確認。
-->
<script setup lang="ts">
import axios from 'axios';
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue';
import { onBeforeRouteLeave, useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';

import { fetchCommunities } from '@/httpapis/communities';
import {
  fetchStaffPropertySaleDetail,
  fetchStaffServicedApartmentDetail,
  updateStaffPropertySale,
  updateStaffServicedApartment,
} from '@/httpapis/staff';
import {
  createPropertySale,
  createServicedApartment,
  fetchPropertySaleDetail,
  fetchServicedApartmentDetail,
  publishPropertySale,
  publishServicedApartment,
  searchPropertyAddresses,
  updatePropertySale,
  updateServicedApartment,
} from '@/httpapis/properties';
import { completeUpload, createUploadPresign } from '@/httpapis/uploads';
import {
  getPropertyOptionLabel,
  marketplaceDistricts,
  propertyAdPackageOptions,
  propertyAnnualPrepayOptions,
  propertyAreaModeOptions,
  propertyBusinessStatusOptions,
  propertyCookingModeOptions,
  propertyContactMethodOptions,
  propertyFeatureTagOptions,
  propertyFloorZoneOptions,
  propertyKitchenTypeOptions,
  propertyListingCategoryOptions,
  propertyLocationScopeOptions,
  propertyPublisherFilterOptions,
  propertyRenovationFilterOptions,
  propertyTransactionTypeOptions,
  propertyTypeOptions,
  servicedAdPackageOptions,
  servicedFacilityTagOptions,
  servicedServiceTagOptions,
  servicedStayUnitOptions,
} from '@/constants/property';
import type { MetaCommunity } from '@/model/community';
import type { MediaAssetResponse } from '@/model/marketplace';
import type {
  PropertyChannel,
  PropertyAddressSuggestion,
  PropertyImagePayload,
  PropertyListingDetailResponse,
  ServicedApartmentRoomType,
  UpsertPropertySalePayload,
  UpsertServicedApartmentPayload,
} from '@/model/property';
import AppIcon from '@/shared/components/base/AppIcon.vue';
import AppUnsavedChangesDialog from '@/shared/components/base/AppUnsavedChangesDialog.vue';
import { useFeedbackStore } from '@/stores/feedback';
import { usePreferenceStore } from '@/stores/preferences';
import { useSessionStore } from '@/stores/session';
import { formatPrice } from '@/utils/format';
import { buildUploadHeaders } from '@/utils/upload';
import { formatAjoPoints } from '@/utils/wallet';
import { mergePropertyFeatureTags } from '@/utils/property';

void propertyAnnualPrepayOptions;
void propertyAreaModeOptions;
void propertyCookingModeOptions;
void propertyFloorZoneOptions;
void propertyKitchenTypeOptions;
void propertyListingCategoryOptions;
void propertyLocationScopeOptions;
void propertyTransactionTypeOptions;
void servicedAdPackageOptions;
void servicedStayUnitOptions;

const props = withDefaults(defineProps<{
  channel: PropertyChannel;
  returnPath?: string;
  listingId?: string;
  embedded?: boolean;
  hideHeader?: boolean;
  staffMode?: boolean;
}>(), {
  returnPath: '',
  listingId: '',
  embedded: false,
  hideHeader: false,
  staffMode: false,
});

const emit = defineEmits<{
  (event: 'saved', listingId: string): void;
  (event: 'published', listingId: string): void;
  (event: 'cancel'): void;
}>();

type PropertyEditorLeaveDecision = 'save' | 'discard' | 'stay';

interface PropertyEditorImage {
  id: string;
  mediaAssetId?: string;
  url?: string;
  file?: File;
  objectUrl?: string;
  isCover: boolean;
  uploading: boolean;
}

interface PropertyEditorForm {
  title: string;
  titleEn: string;
  summary: string;
  description: string;
  descriptionEn: string;
  districtCode: string;
  communityId: string;
  publisherIdentityType: string;
  businessStatus: 'available' | 'sold';
  contactMethod: 'phone' | 'whatsapp' | 'chat' | 'both' | 'chat_or_whatsapp';
  phone: string;
  whatsapp: string;
  email: string;
  allowPhone: boolean;
  allowWhatsapp: boolean;
  allowChat: boolean;
  adPackageCode: string;
  propertyNo: string;
  transactionType: 'sale' | 'rent';
  locationScope: 'local' | 'overseas';
  listingCategory: string;
  multiUnitProject: boolean;
  propertyType: string;
  rentalType: string;
  renovationType: string;
  agencyCompanyName: string;
  estateName: string;
  addressText: string;
  addressTextEn: string;
  blockName: string;
  unitName: string;
  showUnit: boolean;
  latitude: number;
  longitude: number;
  askingPriceHKD: number;
  monthlyRentHKD: number;
  priceReferenceOnly: boolean;
  priceNegotiable: boolean;
  annualPrepayDiscount: boolean;
  annualPrepayOption: string;
  leaseStartDate: string;
  rentIncluded: string;
  areaMode: 'usable' | 'gross';
  usableAreaSqft: number;
  grossAreaSqft: number;
  bedroomCount: number;
  livingRoomCount: number;
  bathroomCount: number;
  floorLevel: string;
  floorRaw: string;
  floorZone: string;
  totalFloors: number;
  direction: string;
  buildingAge: string;
  completionYear: number;
  buildingTotalFloors: number;
  managementCompany: string;
  kitchenType: string;
  cookingMode: string;
  managementFeeHKD: number;
  videoURL: string;
  vrURL: string;
  privateNote: string;
  featureTags: string[];
  projectName: string;
  projectNameEn: string;
  websiteURL: string;
  serviceWhatsApp: string;
  fax: string;
  serviceIntro: string;
  benefitsText: string;
  extraChargesText: string;
  lowestMonthlyRentHKD: number;
  lowestDailyRentHKD: number;
  priceReferenceOnlyServiced: boolean;
  priceNegotiableServiced: boolean;
  minUsableAreaSqft: number;
  minLeaseMonths: number;
  minStayValue: number;
  minStayUnit: 'month' | 'day';
  facilityTags: string[];
  serviceTags: string[];
  roomTypes: ServicedApartmentRoomType[];
}

const maxImages = 40;
const maxImageSize = 10 * 1024 * 1024;

// 1. 建立服務式住宅房型
const createServicedRoomType = (): ServicedApartmentRoomType => ({
  name: 'Studio',
  room_category: '',
  usable_area_sqft: 0,
  monthly_rent_hkd: 0,
  monthly_rent_min_hkd: 0,
  monthly_rent_max_hkd: 0,
  daily_rent_min_hkd: 0,
  daily_rent_max_hkd: 0,
  included_fees: true,
  included_fee_items: [],
  min_lease_months: 1,
  min_stay_value: 1,
  min_stay_unit: 'month',
  feature_tags: [],
});

const route = useRoute();
const router = useRouter();
const { t } = useI18n();
const feedbackStore = useFeedbackStore();
const preferenceStore = usePreferenceStore();
const sessionStore = useSessionStore();

const listingId = ref(props.listingId || String(route.params.listingId ?? ''));
const loading = ref(false);
const saving = ref(false);
const publishing = ref(false);
const isLeavePromptOpen = ref(false);
const savedSnapshot = ref('');
const isProgrammaticNavigation = ref(false);
const addressSuggestions = ref<PropertyAddressSuggestion[]>([]);
const loadingAddressSuggestions = ref(false);
const addressSearchTimer = ref<ReturnType<typeof window.setTimeout> | null>(null);
const images = ref<PropertyEditorImage[]>([]);
const communities = ref<MetaCommunity[]>([]);
let resolveLeavePrompt: ((decision: PropertyEditorLeaveDecision) => void) | null = null;

const form = reactive<PropertyEditorForm>({
  title: '',
  titleEn: '',
  summary: '',
  description: '',
  descriptionEn: '',
  districtCode: sessionStore.me?.district_code || 'kwun_tong',
  communityId: sessionStore.me?.primary_community?.public_id || '',
  publisherIdentityType: sessionStore.me?.publisher_identity_type || 'owner',
  businessStatus: 'available',
  contactMethod: 'both',
  phone: '',
  whatsapp: '',
  email: sessionStore.me?.email || '',
  allowPhone: true,
  allowWhatsapp: true,
  allowChat: false,
  adPackageCode: 'basic',
  propertyNo: '',
  transactionType: 'sale',
  locationScope: 'local',
  listingCategory: 'standard',
  multiUnitProject: false,
  propertyType: 'private_flat',
  rentalType: '',
  renovationType: '',
  agencyCompanyName: '',
  estateName: '',
  addressText: '',
  addressTextEn: '',
  blockName: '',
  unitName: '',
  showUnit: false,
  latitude: 0,
  longitude: 0,
  askingPriceHKD: 0,
  monthlyRentHKD: 0,
  priceReferenceOnly: false,
  priceNegotiable: false,
  annualPrepayDiscount: false,
  annualPrepayOption: '95_off',
  leaseStartDate: '',
  rentIncluded: '',
  areaMode: 'usable',
  usableAreaSqft: 0,
  grossAreaSqft: 0,
  bedroomCount: 2,
  livingRoomCount: 1,
  bathroomCount: 1,
  floorLevel: '',
  floorRaw: '',
  floorZone: '',
  totalFloors: 0,
  direction: '',
  buildingAge: '',
  completionYear: 0,
  buildingTotalFloors: 0,
  managementCompany: '',
  kitchenType: '',
  cookingMode: '',
  managementFeeHKD: 0,
  videoURL: '',
  vrURL: '',
  privateNote: '',
  featureTags: [],
  projectName: '',
  projectNameEn: '',
  websiteURL: '',
  serviceWhatsApp: '',
  fax: '',
  serviceIntro: '',
  benefitsText: '',
  extraChargesText: '',
  lowestMonthlyRentHKD: 0,
  lowestDailyRentHKD: 0,
  priceReferenceOnlyServiced: false,
  priceNegotiableServiced: false,
  minUsableAreaSqft: 0,
  minLeaseMonths: 1,
  minStayValue: 1,
  minStayUnit: 'month',
  facilityTags: [],
  serviceTags: [],
  roomTypes: [createServicedRoomType()],
});

const isSale = computed(() => props.channel === 'sale');
const isEditing = computed(() => listingId.value.trim().length > 0);
const pageTitle = computed(() =>
  isSale.value ? t('property.sale.publishTitle') : t('property.serviced.publishTitle'),
);
const publisherOptions = computed(() =>
  propertyPublisherFilterOptions.filter((option) => option.value.trim() !== ''),
);
const myPath = computed(() =>
  props.returnPath || (isSale.value ? '/properties/my' : '/serviced-residences/my'),
);
const activeAdPackageOptions = computed(() =>
  isSale.value ? propertyAdPackageOptions : servicedAdPackageOptions,
);
const selectedAdPackage = computed(() =>
  activeAdPackageOptions.value.find((option) => option.value === form.adPackageCode) ??
  activeAdPackageOptions.value[0],
);
const hasImage = computed(() => images.value.some((image) => image.mediaAssetId || image.file));
const salePriceReady = computed(() =>
  form.priceNegotiable ||
  (form.transactionType === 'rent' ? form.monthlyRentHKD > 0 : form.askingPriceHKD > 0),
);
const servicedPriceReady = computed(() =>
  form.priceNegotiableServiced ||
  form.lowestMonthlyRentHKD > 0 ||
  form.lowestDailyRentHKD > 0 ||
  form.roomTypes.some((room) =>
    Number(room.monthly_rent_min_hkd || room.monthly_rent_hkd || room.daily_rent_min_hkd || 0) > 0,
  ),
);
const contactReady = computed(() => form.allowPhone || form.allowWhatsapp || form.allowChat);
const canSave = computed(() =>
  form.title.trim() !== '' &&
  form.summary.trim() !== '' &&
  form.description.trim() !== '' &&
  form.districtCode.trim() !== '' &&
  form.addressText.trim() !== '' &&
  contactReady.value &&
  (isSale.value
    ? salePriceReady.value &&
      form.estateName.trim() !== '' &&
      form.usableAreaSqft > 0 &&
      (form.floorRaw.trim() !== '' || form.floorLevel.trim() !== '')
    : form.projectName.trim() !== '' &&
      servicedPriceReady.value &&
      form.minLeaseMonths > 0 &&
      form.roomTypes.length > 0 &&
      form.roomTypes.every((room) =>
        room.name.trim() !== '' &&
        Number(room.usable_area_sqft || 0) > 0,
      )),
);
const previewPrice = computed(() => {
  if ((isSale.value && form.priceNegotiable) || (!isSale.value && form.priceNegotiableServiced)) {
    return t('property.common.negotiable');
  }
  const value = isSale.value
    ? (form.transactionType === 'rent' ? form.monthlyRentHKD : form.askingPriceHKD)
    : (
        form.lowestMonthlyRentHKD ||
        form.roomTypes[0]?.monthly_rent_min_hkd ||
        form.roomTypes[0]?.monthly_rent_hkd ||
        form.lowestDailyRentHKD ||
        form.roomTypes[0]?.daily_rent_min_hkd ||
        0
      );

  return value > 0 ? formatPrice(value, preferenceStore.locale) : t('property.common.pendingPrice');
});
const chargeCost = computed(() => selectedAdPackage.value.price_points);
const formatPoints = (value: number): string =>
  formatAjoPoints(value, t('common.brand.pointsName'), preferenceStore.locale);
const formatAdPackagePrice = (value: number): string =>
  new Intl.NumberFormat(preferenceStore.locale, { maximumFractionDigits: 0 }).format(value);
const chargeHint = computed(() =>
  `${t('property.editor.chargeHint')} ${formatPoints(chargeCost.value)} · ${t('property.editor.walletBalance')} ${formatPoints(sessionStore.me?.ajo_balance ?? 0)}`,
);
const formatCommunityName = (community: MetaCommunity): string => {
  const primaryName = preferenceStore.locale === 'en'
    ? community.name_en || community.name_zh
    : community.name_zh || community.name_en;

  return primaryName || community.address_text || community.public_id;
};
const currentSnapshot = computed(() => createEditorSnapshot());
const hasUnsavedChanges = computed(() =>
  savedSnapshot.value.length > 0 && currentSnapshot.value !== savedSnapshot.value,
);

// 1. 建立表單異動比對快照
function createEditorSnapshot(): string {
  return JSON.stringify({
    form: {
      ...form,
      title: form.title.trim(),
      titleEn: form.titleEn.trim(),
      summary: form.summary.trim(),
      description: form.description.trim(),
      descriptionEn: form.descriptionEn.trim(),
      privateNote: form.privateNote.trim(),
      featureTags: [...form.featureTags].sort(),
      facilityTags: [...form.facilityTags].sort(),
      serviceTags: [...form.serviceTags].sort(),
    },
    images: images.value
      .filter((image) => Boolean(image.mediaAssetId || image.file))
      .map((image) => ({
        mediaAssetId: image.mediaAssetId ?? '',
        fileName: image.file?.name ?? '',
        fileSize: image.file?.size ?? 0,
        fileType: image.file?.type ?? '',
        isCover: image.isCover,
      })),
  });
}

// 2. 更新離開提示基準
const markCurrentStateSaved = (): void => {
  savedSnapshot.value = currentSnapshot.value;
};

// 3. 打開離開確認彈窗
const requestLeaveDecision = (): Promise<PropertyEditorLeaveDecision> => {
  isLeavePromptOpen.value = true;

  return new Promise((resolve) => {
    resolveLeavePrompt = resolve;
  });
};

// 4. 回應離開確認彈窗
const handleLeavePromptDecision = (decision: PropertyEditorLeaveDecision): void => {
  isLeavePromptOpen.value = false;
  resolveLeavePrompt?.(decision);
  resolveLeavePrompt = null;
};

// 5. 確認是否允許離開編輯頁
const confirmLeaveEditor = async (): Promise<boolean> => {
  if (isProgrammaticNavigation.value || !hasUnsavedChanges.value) {
    return true;
  }
  if (saving.value || publishing.value) {
    return false;
  }

  const decision = await requestLeaveDecision();
  if (decision === 'stay') {
    return false;
  }
  if (decision === 'discard') {
    return true;
  }

  const savedListingId = await saveDraft();
  if (savedListingId && props.embedded) {
    emit('saved', savedListingId);
    return false;
  }

  return savedListingId.length > 0;
};

// 6. 處理瀏覽器關閉或重新整理
const handleBeforeUnload = (event: BeforeUnloadEvent): void => {
  if (!hasUnsavedChanges.value) {
    return;
  }

  event.preventDefault();
  event.returnValue = '';
};

// 7. 由外層彈窗要求關閉編輯器
const requestCloseEditor = async (): Promise<void> => {
  if (await confirmLeaveEditor()) {
    emit('cancel');
  }
};

defineExpose({
  requestCloseEditor,
});

// 8. 讀取錯誤訊息
const readErrorMessage = (error: unknown, fallback: string): string =>
  axios.isAxiosError<{ message?: string }>(error)
    ? error.response?.data?.message ?? fallback
    : fallback;

// 9. 切換標籤
const toggleTag = (target: string[], value: string): void => {
  const index = target.indexOf(value);
  if (index >= 0) {
    target.splice(index, 1);
    return;
  }

  target.push(value);
};

// 10. 建立上傳目錄
const buildObjectPrefix = (targetListingId: string): string =>
  targetListingId.trim()
    ? `ajo_living/listings/${targetListingId.trim()}/`
    : 'ajo_living/listings/';

// 11. 釋放本地圖片預覽
const revokeImagePreview = (image: PropertyEditorImage): void => {
  if (image.objectUrl) {
    URL.revokeObjectURL(image.objectUrl);
  }
};

// 12. 選擇圖片
const handleImageFilesChange = (event: Event): void => {
  const input = event.target as HTMLInputElement;
  const files = Array.from(input.files ?? []);
  input.value = '';

  for (const file of files) {
    if (!file.type.startsWith('image/') || file.size > maxImageSize) {
      feedbackStore.pushToast(t('property.editor.uploadError'), 'error');
      continue;
    }
    if (images.value.length >= maxImages) {
      break;
    }

    const objectUrl = URL.createObjectURL(file);
    images.value.push({
      id: `${Date.now()}-${file.name}-${images.value.length}`,
      file,
      url: objectUrl,
      objectUrl,
      isCover: !images.value.some((image) => image.isCover),
      uploading: false,
    });
  }
};

// 13. 移除圖片
const removeImage = (imageId: string): void => {
  const target = images.value.find((image) => image.id === imageId);
  if (target) {
    revokeImagePreview(target);
  }

  images.value = images.value.filter((image) => image.id !== imageId);
  if (!images.value.some((image) => image.isCover) && images.value[0]) {
    images.value[0].isCover = true;
  }
};

// 14. 選擇封面
const selectCover = (imageId: string): void => {
  images.value = images.value.map((image) => ({
    ...image,
    isCover: image.id === imageId,
  }));
};

// 15. 更新圖片上傳結果
const updateImageAsset = (imageId: string, mediaAsset: MediaAssetResponse): void => {
  images.value = images.value.map((image) => {
    if (image.id !== imageId) {
      return image;
    }

    revokeImagePreview(image);
    return {
      ...image,
      id: mediaAsset.media_asset_id,
      mediaAssetId: mediaAsset.media_asset_id,
      url: mediaAsset.url,
      file: undefined,
      objectUrl: undefined,
      uploading: false,
    };
  });
};

// 16. 上傳單張圖片
const uploadImage = async (image: PropertyEditorImage, targetListingId: string): Promise<void> => {
  if (!image.file) {
    return;
  }

  image.uploading = true;
  const presignResponse = await createUploadPresign({
    file_name: image.file.name,
    mime_type: image.file.type,
    file_size: image.file.size,
    object_prefix: buildObjectPrefix(targetListingId),
  });
  const presign = presignResponse.data.data;
  const uploadResponse = await fetch(presign.upload_url, {
    method: 'PUT',
    headers: buildUploadHeaders(presign.headers, image.file.type),
    body: image.file,
  });
  if (!uploadResponse.ok) {
    throw new Error(`image upload failed with status ${uploadResponse.status}`);
  }

  const completeResponse = await completeUpload({
    object_key: presign.object_key,
    mime_type: image.file.type,
    file_size: image.file.size,
  });
  updateImageAsset(image.id, completeResponse.data.data);
};

// 17. 上傳待處理圖片
const uploadPendingImages = async (targetListingId: string): Promise<void> => {
  const pendingImages = images.value.filter((image) => image.file);
  for (const image of pendingImages) {
    await uploadImage(image, targetListingId);
  }
};

// 18. 建立圖片 payload
const buildImagePayload = (): PropertyImagePayload[] =>
  images.value
    .filter((image) => image.mediaAssetId)
    .map((image, index) => ({
      media_asset_id: image.mediaAssetId as string,
      sort_order: index + 1,
      is_cover: image.isCover || index === 0,
    }));

// 19. 分割文字清單
const splitTextList = (value: string): string[] =>
  value
    .split(/[,，、\n]/)
    .map((item) => item.trim())
    .filter(Boolean);

// 20. 更新服務式住宅房型包含項目
const updateRoomIncludedFeeItems = (room: ServicedApartmentRoomType, event: Event): void => {
  room.included_fee_items = splitTextList((event.target as HTMLInputElement).value);
};

// 21. 整理服務式住宅房型
const normalizeServicedRoomType = (room: ServicedApartmentRoomType): ServicedApartmentRoomType => {
  const stayUnit = room.min_stay_unit === 'day' ? 'day' : 'month';
  const fallbackRent = form.priceReferenceOnlyServiced || form.priceNegotiableServiced ? 1 : 0;
  const monthlyMin = stayUnit === 'day'
    ? 0
    : Number(room.monthly_rent_min_hkd || room.monthly_rent_hkd || fallbackRent);
  const monthlyMax = stayUnit === 'day'
    ? 0
    : Number(room.monthly_rent_max_hkd || monthlyMin || 0);
  const dailyMin = stayUnit === 'day'
    ? Number(room.daily_rent_min_hkd || fallbackRent)
    : Number(room.daily_rent_min_hkd || 0);
  const dailyMax = stayUnit === 'day'
    ? Number(room.daily_rent_max_hkd || dailyMin || 0)
    : Number(room.daily_rent_max_hkd || 0);

  return {
    ...room,
    name: room.name.trim() || 'Studio',
    room_category: room.room_category?.trim() || undefined,
    usable_area_sqft: Number(room.usable_area_sqft || 0),
    monthly_rent_hkd: monthlyMin,
    monthly_rent_min_hkd: monthlyMin,
    monthly_rent_max_hkd: monthlyMax > 0 ? monthlyMax : undefined,
    daily_rent_min_hkd: dailyMin > 0 ? dailyMin : undefined,
    daily_rent_max_hkd: dailyMax > 0 ? dailyMax : undefined,
    included_fees: Boolean(room.included_fees),
    included_fee_items: Array.isArray(room.included_fee_items) ? room.included_fee_items : [],
    min_lease_months: Number(room.min_lease_months || form.minLeaseMonths || 1),
    min_stay_value: Number(room.min_stay_value || room.min_lease_months || form.minStayValue || 1),
    min_stay_unit: stayUnit,
    feature_tags: room.feature_tags ?? [],
  };
};

// 22. 推導服務式住宅項目最短入住
const deriveServicedProjectMinStay = (roomTypes: ServicedApartmentRoomType[]) => {
  const normalizedRooms = roomTypes.length > 0 ? roomTypes : [createServicedRoomType()];
  const dayRooms = normalizedRooms.filter((room) => room.min_stay_unit === 'day');
  if (dayRooms.length > 0) {
    const minStayValue = Math.min(...dayRooms.map((room) => Number(room.min_stay_value || 1)));
    return {
      minLeaseMonths: Math.max(1, Math.ceil(minStayValue / 30)),
      minStayValue,
      minStayUnit: 'day' as const,
    };
  }

  const minStayValue = Math.min(...normalizedRooms.map((room) =>
    Number(room.min_stay_value || room.min_lease_months || form.minStayValue || 1),
  ));

  return {
    minLeaseMonths: minStayValue,
    minStayValue,
    minStayUnit: 'month' as const,
  };
};

// 23. 新增服務式住宅房型
const addServicedRoomType = (): void => {
  form.roomTypes.push(createServicedRoomType());
};

// 24. 移除服務式住宅房型
const removeServicedRoomType = (roomIndex: number): void => {
  if (form.roomTypes.length <= 1) {
    return;
  }

  form.roomTypes.splice(roomIndex, 1);
};

// 25. 讀取屋苑及大廈清單
const loadCommunities = async (): Promise<void> => {
  try {
    const response = await fetchCommunities();
    communities.value = response.data.data.items;
  } catch {
    communities.value = [];
  }
};

// 26. 建立樓盤 payload
const buildSalePayload = (): UpsertPropertySalePayload => ({
  title: form.title.trim(),
  title_en: form.titleEn.trim() || undefined,
  summary: form.summary.trim(),
  description: form.description.trim(),
  description_en: form.descriptionEn.trim() || undefined,
  district_code: form.districtCode,
  community_id: form.communityId,
  publisher_identity_type: form.publisherIdentityType || 'owner',
  property_no: form.propertyNo.trim() || undefined,
  transaction_type: form.transactionType,
  location_scope: form.locationScope,
  listing_category: form.listingCategory,
  multi_unit_project: form.multiUnitProject,
  property_type: form.propertyType,
  rental_type: form.rentalType.trim() || undefined,
  renovation_type: form.renovationType || undefined,
  agency_company_name: form.agencyCompanyName.trim() || undefined,
  estate_name: form.estateName.trim(),
  address_text: form.addressText.trim(),
  address_text_en: form.addressTextEn.trim() || undefined,
  block_name: form.blockName.trim() || undefined,
  unit_name: form.unitName.trim() || undefined,
  show_unit: form.showUnit,
  latitude: form.latitude ? Number(form.latitude) : undefined,
  longitude: form.longitude ? Number(form.longitude) : undefined,
  asking_price_hkd: form.transactionType === 'rent' ? 0 : Number(form.askingPriceHKD),
  monthly_rent_hkd: form.transactionType === 'rent' ? Number(form.monthlyRentHKD) : undefined,
  price_reference_only: form.priceReferenceOnly,
  price_negotiable: form.priceNegotiable,
  annual_prepay_discount: form.transactionType === 'rent' ? form.annualPrepayDiscount : false,
  annual_prepay_option: form.transactionType === 'rent' && form.annualPrepayDiscount
    ? form.annualPrepayOption
    : 'none',
  lease_start_date: form.transactionType === 'rent' ? form.leaseStartDate.trim() || undefined : undefined,
  rent_included: form.transactionType === 'rent' ? form.rentIncluded.trim() || undefined : undefined,
  area_mode: form.areaMode,
  usable_area_sqft: Number(form.usableAreaSqft),
  gross_area_sqft: form.grossAreaSqft > 0 ? Number(form.grossAreaSqft) : undefined,
  bedroom_count: Number(form.bedroomCount),
  living_room_count: Number(form.livingRoomCount),
  bathroom_count: Number(form.bathroomCount),
  floor_level: form.floorLevel.trim() || form.floorRaw.trim(),
  floor_raw: form.floorRaw.trim() || form.floorLevel.trim(),
  floor_zone: form.floorZone || undefined,
  total_floors: form.totalFloors > 0 ? Number(form.totalFloors) : undefined,
  direction: form.direction.trim(),
  building_age: form.buildingAge.trim(),
  completion_year: form.completionYear > 0 ? Number(form.completionYear) : undefined,
  building_total_floors: form.buildingTotalFloors > 0 ? Number(form.buildingTotalFloors) : undefined,
  management_company: form.managementCompany.trim() || undefined,
  kitchen_type: form.kitchenType || undefined,
  cooking_mode: form.cookingMode || undefined,
  management_fee_hkd: form.managementFeeHKD > 0 ? Number(form.managementFeeHKD) : undefined,
  video_url: form.videoURL.trim() || undefined,
  vr_url: form.vrURL.trim() || undefined,
  private_note: form.privateNote.trim() || undefined,
  ad_package_code: form.adPackageCode,
  feature_tags: mergePropertyFeatureTags(form.featureTags, {
    video_url: form.videoURL.trim(),
    vr_url: form.vrURL.trim(),
    annual_prepay_discount: form.transactionType === 'rent' ? form.annualPrepayDiscount : false,
    annual_prepay_option: form.transactionType === 'rent' && form.annualPrepayDiscount
      ? form.annualPrepayOption
      : 'none',
  }),
  contact_method: form.contactMethod,
  business_status: form.businessStatus,
  images: buildImagePayload(),
  contact: {
    phone: form.phone.trim(),
    whatsapp: form.whatsapp.trim(),
    email: form.email.trim(),
    show_phone: form.allowPhone,
    show_whatsapp: form.allowWhatsapp,
    show_chat: form.allowChat,
    show_inquiry_form: false,
  },
});

// 27. 建立服務式住宅 payload
const buildServicedPayload = (): UpsertServicedApartmentPayload => {
  const roomTypes = form.roomTypes.map(normalizeServicedRoomType);
  const monthlyPrices = roomTypes
    .map((room) => Number(room.monthly_rent_min_hkd || room.monthly_rent_hkd || 0))
    .filter((value) => value > 0);
  const dailyPrices = roomTypes
    .map((room) => Number(room.daily_rent_min_hkd || 0))
    .filter((value) => value > 0);
  const areas = roomTypes
    .map((room) => Number(room.usable_area_sqft || 0))
    .filter((value) => value > 0);
  const projectMinStay = deriveServicedProjectMinStay(roomTypes);

  return {
    title: form.title.trim(),
    title_en: form.titleEn.trim() || undefined,
    summary: form.summary.trim(),
    description: form.description.trim(),
    description_en: form.descriptionEn.trim() || undefined,
    district_code: form.districtCode,
    community_id: form.communityId,
    publisher_identity_type: form.publisherIdentityType || 'owner',
    project_name: form.projectName.trim(),
    project_name_en: form.projectNameEn.trim() || undefined,
    address_text: form.addressText.trim(),
    address_text_en: form.addressTextEn.trim() || undefined,
    website_url: form.websiteURL.trim() || undefined,
    whatsapp: form.serviceWhatsApp.trim() || undefined,
    fax: form.fax.trim() || undefined,
    service_intro: form.serviceIntro.trim() || undefined,
    benefits_text: form.benefitsText.trim() || undefined,
    extra_charges_text: form.extraChargesText.trim() || undefined,
    lowest_monthly_rent_hkd: monthlyPrices.length
      ? Math.min(...monthlyPrices)
      : Number(form.lowestMonthlyRentHKD || 0),
    lowest_daily_rent_hkd: dailyPrices.length
      ? Math.min(...dailyPrices)
      : form.lowestDailyRentHKD > 0 ? Number(form.lowestDailyRentHKD) : undefined,
    price_reference_only: form.priceReferenceOnlyServiced,
    price_negotiable: form.priceNegotiableServiced,
    min_usable_area_sqft: areas.length
      ? Math.min(...areas)
      : form.minUsableAreaSqft > 0 ? Number(form.minUsableAreaSqft) : undefined,
    min_lease_months: projectMinStay.minLeaseMonths,
    min_stay_value: projectMinStay.minStayValue,
    min_stay_unit: projectMinStay.minStayUnit,
    location_scope: form.locationScope,
    listing_category: form.listingCategory,
    multi_unit_project: form.multiUnitProject,
    ad_package_code: form.adPackageCode,
    facility_tags: form.facilityTags,
    service_tags: form.serviceTags,
    room_types: roomTypes,
    contact_method: form.contactMethod,
    business_status: form.businessStatus,
    images: buildImagePayload(),
    contact: {
      phone: form.phone.trim(),
      whatsapp: form.whatsapp.trim(),
      email: form.email.trim(),
      show_phone: form.allowPhone,
      show_whatsapp: form.allowWhatsapp,
      show_chat: form.allowChat,
      show_inquiry_form: false,
    },
  };
};

// 28. 清除地址聯想計時器
const clearAddressSearchTimer = (): void => {
  if (addressSearchTimer.value) {
    window.clearTimeout(addressSearchTimer.value);
    addressSearchTimer.value = null;
  }
};

// 29. 取得地址聯想標題
const resolveAddressSuggestionTitle = (suggestion: PropertyAddressSuggestion): string =>
  suggestion.display_name ||
  [
    suggestion.estate_name,
    suggestion.estate_name_en && suggestion.estate_name_en !== suggestion.estate_name
      ? suggestion.estate_name_en
      : '',
    suggestion.district_label ? `(${suggestion.district_label})` : '',
  ].filter(Boolean).join(' ');

// 30. 讀取屋苑或住宅地址聯想
const loadAddressSuggestions = async (): Promise<void> => {
  const keyword = isSale.value ? form.estateName.trim() : form.projectName.trim();
  if (!keyword) {
    addressSuggestions.value = [];
    return;
  }

  loadingAddressSuggestions.value = true;
  try {
    const response = await searchPropertyAddresses({
      keyword,
      district_code: form.districtCode || undefined,
      limit: 12,
    });

    addressSuggestions.value = response.data.data;
  } catch {
    addressSuggestions.value = [];
  } finally {
    loadingAddressSuggestions.value = false;
  }
};

// 31. 延遲讀取地址聯想
const scheduleAddressSuggestions = (): void => {
  clearAddressSearchTimer();
  addressSearchTimer.value = window.setTimeout(() => {
    void loadAddressSuggestions();
  }, 220);
};

// 32. 套用地址聯想
const applyAddressSuggestion = (suggestion: PropertyAddressSuggestion): void => {
  const currentYear = new Date().getFullYear();

  if (isSale.value) {
    form.estateName = resolveAddressSuggestionTitle(suggestion);
  } else {
    form.projectName = suggestion.estate_name || resolveAddressSuggestionTitle(suggestion);
    form.projectNameEn = suggestion.estate_name_en;
  }

  form.addressText = suggestion.address_text;
  form.addressTextEn = suggestion.address_text_en;
  form.blockName = suggestion.block_names[0] ?? form.blockName;
  form.buildingAge = suggestion.completion_year > 0 && suggestion.completion_year <= currentYear
    ? String(currentYear - suggestion.completion_year)
    : form.buildingAge;
  if (suggestion.district_code) {
    form.districtCode = suggestion.district_code;
  }
  addressSuggestions.value = [];
};

// 33. 回填詳情
const applyDetail = (detail: PropertyListingDetailResponse): void => {
  form.title = detail.title;
  form.titleEn = detail.property_sale?.title_en || '';
  form.summary = detail.summary;
  form.description = detail.description;
  form.descriptionEn =
    detail.property_sale?.description_en ||
    detail.serviced_apartment?.description_en ||
    '';
  form.districtCode = detail.district_code;
  form.communityId = detail.community?.public_id || '';
  form.publisherIdentityType = detail.publisher_identity_type;
  form.businessStatus = detail.business_status === 'sold' ? 'sold' : 'available';
  form.allowPhone = detail.contact_summary.show_phone;
  form.allowWhatsapp = detail.contact_summary.show_whatsapp;
  form.allowChat = detail.contact_summary.show_chat;

  if (detail.property_sale) {
    form.propertyNo = detail.property_sale.property_no || '';
    form.transactionType = detail.property_sale.transaction_type;
    form.locationScope = detail.property_sale.location_scope === 'overseas' ? 'overseas' : 'local';
    form.listingCategory = detail.property_sale.listing_category || 'standard';
    form.multiUnitProject = Boolean(detail.property_sale.multi_unit_project);
    form.propertyType = detail.property_sale.property_type;
    form.rentalType = detail.property_sale.rental_type || '';
    form.renovationType = detail.property_sale.renovation_type || '';
    form.agencyCompanyName = detail.property_sale.agency_company_name || '';
    form.estateName = detail.property_sale.estate_name;
    form.addressText = detail.property_sale.address_text;
    form.addressTextEn = detail.property_sale.address_text_en || '';
    form.blockName = detail.property_sale.block_name || '';
    form.unitName = detail.property_sale.unit_name || '';
    form.showUnit = Boolean(detail.property_sale.show_unit);
    form.latitude = detail.property_sale.latitude ?? 0;
    form.longitude = detail.property_sale.longitude ?? 0;
    form.askingPriceHKD = detail.property_sale.asking_price_hkd;
    form.monthlyRentHKD = detail.property_sale.monthly_rent_hkd ?? 0;
    form.priceReferenceOnly = Boolean(detail.property_sale.price_reference_only);
    form.priceNegotiable = Boolean(detail.property_sale.price_negotiable);
    form.annualPrepayDiscount = Boolean(detail.property_sale.annual_prepay_discount);
    form.annualPrepayOption = detail.property_sale.annual_prepay_option || '95_off';
    form.leaseStartDate = detail.property_sale.lease_start_date || '';
    form.rentIncluded = detail.property_sale.rent_included || '';
    form.areaMode = detail.property_sale.area_mode === 'gross' ? 'gross' : 'usable';
    form.usableAreaSqft = detail.property_sale.usable_area_sqft;
    form.grossAreaSqft = detail.property_sale.gross_area_sqft ?? 0;
    form.bedroomCount = detail.property_sale.bedroom_count;
    form.livingRoomCount = detail.property_sale.living_room_count;
    form.bathroomCount = detail.property_sale.bathroom_count;
    form.floorLevel = detail.property_sale.floor_level;
    form.floorRaw = detail.property_sale.floor_raw || detail.property_sale.floor_level;
    form.floorZone = detail.property_sale.floor_zone || '';
    form.totalFloors = detail.property_sale.total_floors ?? 0;
    form.direction = detail.property_sale.direction;
    form.buildingAge = detail.property_sale.building_age;
    form.completionYear = detail.property_sale.completion_year ?? 0;
    form.buildingTotalFloors = detail.property_sale.building_total_floors ?? 0;
    form.managementCompany = detail.property_sale.management_company || '';
    form.kitchenType = detail.property_sale.kitchen_type || '';
    form.cookingMode = detail.property_sale.cooking_mode || '';
    form.managementFeeHKD = detail.property_sale.management_fee_hkd ?? 0;
    form.videoURL = detail.property_sale.video_url || '';
    form.vrURL = detail.property_sale.vr_url || '';
    form.privateNote = detail.property_sale.private_note || '';
    form.adPackageCode = detail.property_sale.ad_package_code || 'basic';
    form.featureTags = [...detail.property_sale.feature_tags];
    form.contactMethod = detail.property_sale.contact_method as PropertyEditorForm['contactMethod'];
  }

  if (detail.serviced_apartment) {
    form.projectName = detail.serviced_apartment.project_name;
    form.projectNameEn = detail.serviced_apartment.project_name_en || '';
    form.addressText = detail.serviced_apartment.address_text;
    form.addressTextEn = detail.serviced_apartment.address_text_en || '';
    form.websiteURL = detail.serviced_apartment.website_url || '';
    form.serviceWhatsApp = detail.serviced_apartment.whatsapp || '';
    form.fax = detail.serviced_apartment.fax || '';
    form.serviceIntro = detail.serviced_apartment.service_intro || '';
    form.benefitsText = detail.serviced_apartment.benefits_text || '';
    form.extraChargesText = detail.serviced_apartment.extra_charges_text || '';
    form.lowestMonthlyRentHKD = detail.serviced_apartment.lowest_monthly_rent_hkd;
    form.lowestDailyRentHKD = detail.serviced_apartment.lowest_daily_rent_hkd ?? 0;
    form.priceReferenceOnlyServiced = Boolean(detail.serviced_apartment.price_reference_only);
    form.priceNegotiableServiced = Boolean(detail.serviced_apartment.price_negotiable);
    form.minUsableAreaSqft = detail.serviced_apartment.min_usable_area_sqft ?? 0;
    form.minLeaseMonths = detail.serviced_apartment.min_lease_months;
    form.minStayValue = detail.serviced_apartment.min_stay_value || detail.serviced_apartment.min_lease_months;
    form.minStayUnit = detail.serviced_apartment.min_stay_unit === 'day' ? 'day' : 'month';
    form.locationScope = detail.serviced_apartment.location_scope === 'overseas' ? 'overseas' : 'local';
    form.listingCategory = detail.serviced_apartment.listing_category || 'standard';
    form.multiUnitProject = Boolean(detail.serviced_apartment.multi_unit_project);
    form.adPackageCode = detail.serviced_apartment.ad_package_code || 'basic';
    form.facilityTags = [...detail.serviced_apartment.facility_tags];
    form.serviceTags = [...detail.serviced_apartment.service_tags];
    form.contactMethod = detail.serviced_apartment.contact_method as PropertyEditorForm['contactMethod'];
    form.roomTypes = detail.serviced_apartment.room_types.length > 0
      ? detail.serviced_apartment.room_types.map((item) => ({
          ...createServicedRoomType(),
          ...item,
          monthly_rent_min_hkd: item.monthly_rent_min_hkd || item.monthly_rent_hkd,
          monthly_rent_max_hkd: item.monthly_rent_max_hkd || item.monthly_rent_min_hkd || item.monthly_rent_hkd,
          daily_rent_max_hkd: item.daily_rent_max_hkd || item.daily_rent_min_hkd || 0,
          min_stay_value: item.min_stay_value || item.min_lease_months || 1,
          min_stay_unit: item.min_stay_unit === 'day' ? 'day' : 'month',
        }))
      : [createServicedRoomType()];

  }

  images.value.forEach(revokeImagePreview);
  images.value = detail.images.map((image, index) => ({
    id: image.media_asset_id,
    mediaAssetId: image.media_asset_id,
    url: image.url,
    isCover: image.is_cover || index === 0,
    uploading: false,
  }));
};

// 34. 載入編輯資料
const loadDetail = async (): Promise<void> => {
  if (!isEditing.value) {
    return;
  }

  loading.value = true;
  try {
    const response = props.staffMode
      ? isSale.value
        ? await fetchStaffPropertySaleDetail(listingId.value)
        : await fetchStaffServicedApartmentDetail(listingId.value)
      : isSale.value
        ? await fetchPropertySaleDetail(listingId.value)
        : await fetchServicedApartmentDetail(listingId.value);

    applyDetail(response.data.data);
    markCurrentStateSaved();
  } catch (error: unknown) {
    feedbackStore.pushToast(readErrorMessage(error, t('property.detail.loadError')), 'error');
  } finally {
    loading.value = false;
  }
};

// 35. 儲存草稿
const saveDraft = async (): Promise<string> => {
  if (!canSave.value) {
    feedbackStore.pushToast(t('property.editor.requiredFields'), 'error');
    return '';
  }

  saving.value = true;
  try {
    if (!listingId.value) {
      if (props.staffMode) {
        feedbackStore.pushToast(t('property.editor.staffCreateDisabled'), 'error');
        return '';
      }

      const response = isSale.value
        ? await createPropertySale({ ...buildSalePayload(), images: [] })
        : await createServicedApartment({ ...buildServicedPayload(), images: [] });

      listingId.value = response.data.data.listing_id;
    }

    await uploadPendingImages(listingId.value);

    if (isSale.value) {
      if (props.staffMode) {
        await updateStaffPropertySale(listingId.value, buildSalePayload());
      } else {
        await updatePropertySale(listingId.value, buildSalePayload());
      }
    } else {
      if (props.staffMode) {
        await updateStaffServicedApartment(listingId.value, buildServicedPayload());
      } else {
        await updateServicedApartment(listingId.value, buildServicedPayload());
      }
    }
    if (!props.staffMode) {
      await sessionStore.loadCurrentUser();
    }
    markCurrentStateSaved();

    feedbackStore.pushToast(t('property.editor.saveSuccess'), 'success');
    return listingId.value;
  } catch (error: unknown) {
    feedbackStore.pushToast(readErrorMessage(error, t('property.editor.saveError')), 'error');
    return '';
  } finally {
    saving.value = false;
  }
};

// 36. 儲存並發布
const saveAndPublish = async (): Promise<void> => {
  if (props.staffMode) {
    const savedListingId = await saveDraft();
    if (savedListingId) {
      emit('saved', savedListingId);
    }
    return;
  }

  if (!hasImage.value) {
    feedbackStore.pushToast(t('property.editor.imageRequired'), 'error');
    return;
  }

  publishing.value = true;
  try {
    const savedListingId = await saveDraft();
    if (!savedListingId) {
      return;
    }

    if (isSale.value) {
      await publishPropertySale(savedListingId);
    } else {
      await publishServicedApartment(savedListingId);
    }
    await sessionStore.loadCurrentUser();

    feedbackStore.pushToast(t('property.editor.publishSuccess'), 'success');
    if (props.embedded) {
      emit('published', savedListingId);
      return;
    }

    isProgrammaticNavigation.value = true;
    await router.push(myPath.value);
  } catch (error: unknown) {
    feedbackStore.pushToast(readErrorMessage(error, t('property.editor.publishError')), 'error');
  } finally {
    publishing.value = false;
  }
};

// 37. 儲存並返回列表
const saveAndReturn = async (): Promise<void> => {
  const savedListingId = await saveDraft();
  if (savedListingId) {
    if (props.embedded) {
      emit('saved', savedListingId);
      return;
    }

    isProgrammaticNavigation.value = true;
    await router.push(myPath.value);
  }
};

onBeforeRouteLeave(() => confirmLeaveEditor());

onMounted(async () => {
  window.addEventListener('beforeunload', handleBeforeUnload);
  await Promise.all([loadCommunities(), loadDetail()]);
  if (!isEditing.value) {
    markCurrentStateSaved();
  }
});

onBeforeUnmount(() => {
  window.removeEventListener('beforeunload', handleBeforeUnload);
  clearAddressSearchTimer();
  images.value.forEach(revokeImagePreview);
});
</script>

<template>
  <main class="property-editor-page">
    <section
      v-if="!props.hideHeader"
      class="property-editor-heading"
    >
      <div>
        <p class="property-kicker">
          {{ isEditing ? 'Edit' : 'Publish' }}
        </p>
        <h1>{{ pageTitle }}</h1>
      </div>
      <div class="property-editor-heading__meta">
        <span>{{ previewPrice }}</span>
      </div>
    </section>

    <section
      v-if="loading"
      class="property-editor-panel"
    >
      {{ t('common.status.loading') }}
    </section>

    <section
      v-else
      class="property-editor-layout"
    >
      <form
        class="property-editor-form"
        @submit.prevent="saveAndReturn"
      >
        <section class="property-editor-panel">
          <h2>{{ isSale ? t('property.editor.stepCategory') : t('property.editor.servicedStepCategory') }}</h2>
          <div class="property-editor-grid">
            <label class="property-input">
              <span>{{ t('property.editor.publisherIdentityField') }}</span>
              <select v-model="form.publisherIdentityType">
                <option
                  v-for="publisher in publisherOptions"
                  :key="publisher.value"
                  :value="publisher.value"
                >
                  {{ publisher.label }}
                </option>
              </select>
            </label>
            <label
              v-if="isSale"
              class="property-input"
            >
              <span>{{ t('property.editor.transactionTypeField') }}</span>
              <select v-model="form.transactionType">
                <option
                  v-for="typeOption in propertyTransactionTypeOptions"
                  :key="typeOption.value"
                  :value="typeOption.value"
                >
                  {{ getPropertyOptionLabel(typeOption, preferenceStore.locale) }}
                </option>
              </select>
            </label>
            <label class="property-input">
              <span>{{ t('property.editor.locationScopeField') }}</span>
              <select v-model="form.locationScope">
                <option
                  v-for="scope in propertyLocationScopeOptions"
                  :key="scope.value"
                  :value="scope.value"
                >
                  {{ getPropertyOptionLabel(scope, preferenceStore.locale) }}
                </option>
              </select>
            </label>
            <label class="property-input">
              <span>{{ t('property.editor.listingCategoryField') }}</span>
              <select v-model="form.listingCategory">
                <option
                  v-for="category in propertyListingCategoryOptions"
                  :key="category.value"
                  :value="category.value"
                >
                  {{ getPropertyOptionLabel(category, preferenceStore.locale) }}
                </option>
              </select>
            </label>
          </div>
          <label class="property-inline-checkbox property-editor-warning">
            <input
              v-model="form.multiUnitProject"
              type="checkbox"
            />
            {{ t('property.editor.multiUnitProjectField') }}
          </label>
          <p class="property-editor-note">
            {{ t('property.editor.multiUnitProjectNote') }}
          </p>
        </section>

        <section class="property-editor-panel">
          <h2>{{ t('property.editor.adPackageField') }}</h2>
          <div class="property-ad-package-grid">
            <button
              v-for="adPackage in activeAdPackageOptions"
              :key="adPackage.value"
              type="button"
              :class="{ 'property-ad-package--active': form.adPackageCode === adPackage.value }"
              @click="form.adPackageCode = adPackage.value"
            >
              <strong>{{ getPropertyOptionLabel(adPackage, preferenceStore.locale) }}</strong>
              <span>
                HKD:{{ formatAdPackagePrice(adPackage.price_hkd) }}
                {{ t('property.editor.orPoints') }}
                {{ formatPoints(adPackage.price_points) }}
              </span>
              <small>
                {{ t('property.editor.adDaysUnit') }}: {{ adPackage.duration_days }}
                · {{ t('property.editor.adWeightField') }}: {{ adPackage.weight }}
              </small>
            </button>
          </div>
        </section>

        <section class="property-editor-panel">
          <h2>{{ t('property.editor.baseInfo') }}</h2>
          <div class="property-editor-grid">
            <label class="property-input">
              <span>{{ t('property.editor.titleField') }}</span>
              <input v-model="form.title" />
            </label>
            <label class="property-input">
              <span>{{ t('property.editor.titleEnField') }}</span>
              <input v-model="form.titleEn" />
            </label>
            <label class="property-input">
              <span>{{ t('property.editor.districtField') }}</span>
              <select v-model="form.districtCode">
                <option
                  v-for="district in marketplaceDistricts"
                  :key="district.value"
                  :value="district.value"
                >
                  {{ getPropertyOptionLabel(district, preferenceStore.locale) }}
                </option>
              </select>
            </label>
            <label class="property-input">
              <span>{{ t('property.editor.communityField') }}</span>
              <select v-model="form.communityId">
                <option value="">
                  {{ t('property.editor.noCommunity') }}
                </option>
                <option
                  v-for="community in communities"
                  :key="community.public_id"
                  :value="community.public_id"
                >
                  {{ formatCommunityName(community) }}
                </option>
              </select>
            </label>
            <label class="property-input">
              <span>{{ t('common.label.status') }}</span>
              <select v-model="form.businessStatus">
                <option
                  v-for="status in propertyBusinessStatusOptions"
                  :key="status.value"
                  :value="status.value"
                >
                  {{ getPropertyOptionLabel(status, preferenceStore.locale) }}
                </option>
              </select>
            </label>
            <label class="property-input property-input--wide">
              <span>{{ t('property.editor.summaryField') }}</span>
              <input v-model="form.summary" />
            </label>
            <label class="property-input property-input--wide">
              <span>{{ t('property.editor.descriptionField') }}</span>
              <textarea
                v-model="form.description"
                rows="5"
              />
            </label>
            <label class="property-input property-input--wide">
              <span>{{ t('property.editor.descriptionEnField') }}</span>
              <textarea
                v-model="form.descriptionEn"
                rows="4"
              />
            </label>
          </div>
        </section>

        <section
          v-if="isSale"
          class="property-editor-panel"
        >
          <h2>{{ t('property.editor.saleTitle') }}</h2>
          <div class="property-editor-grid">
            <label class="property-input">
              <span>{{ t('property.editor.propertyNoField') }}</span>
              <input v-model="form.propertyNo" />
            </label>
            <label class="property-input">
              <span>{{ t('property.sale.typeLabel') }}</span>
              <select v-model="form.propertyType">
                <option
                  v-for="typeOption in propertyTypeOptions"
                  :key="typeOption.value"
                  :value="typeOption.value"
                >
                  {{ getPropertyOptionLabel(typeOption, preferenceStore.locale) }}
                </option>
              </select>
            </label>
            <label class="property-input">
              <span>{{ t('property.editor.renovationField') }}</span>
              <select v-model="form.renovationType">
                <option value="">
                  {{ t('property.editor.notSpecified') }}
                </option>
                <option
                  v-for="option in propertyRenovationFilterOptions.filter((item) => item.value)"
                  :key="option.value"
                  :value="option.value"
                >
                  {{ option.label }}
                </option>
              </select>
            </label>
            <label class="property-input">
              <span>{{ t('property.editor.agencyCompanyField') }}</span>
              <input v-model="form.agencyCompanyName" />
            </label>
            <label class="property-input property-input--wide">
              <span>{{ t('property.editor.estateNameField') }}</span>
              <span class="property-address-input">
                <input
                  v-model="form.estateName"
                  autocomplete="off"
                  @focus="scheduleAddressSuggestions"
                  @input="scheduleAddressSuggestions"
                />
                <i
                  v-if="loadingAddressSuggestions"
                  aria-hidden="true"
                />
                <span
                  v-if="addressSuggestions.length > 0 || loadingAddressSuggestions"
                  class="property-address-suggestions"
                >
                  <span
                    v-if="loadingAddressSuggestions"
                    class="property-address-loading"
                  >
                    {{ t('common.status.loading') }}
                  </span>
                  <button
                    v-for="suggestion in addressSuggestions"
                    :key="suggestion.address_id"
                    type="button"
                    @click="applyAddressSuggestion(suggestion)"
                  >
                    <strong>{{ resolveAddressSuggestionTitle(suggestion) }}</strong>
                    <span>{{ suggestion.address_text }}</span>
                  </button>
                </span>
              </span>
            </label>
            <label class="property-input">
              <span>{{ t('property.editor.blockNameField') }}</span>
              <input v-model="form.blockName" />
            </label>
            <label class="property-input">
              <span>{{ t('property.editor.unitNameField') }}</span>
              <input v-model="form.unitName" />
            </label>
            <label class="property-input property-input--wide">
              <span>{{ t('property.editor.addressField') }}</span>
              <input v-model="form.addressText" />
            </label>
            <label class="property-input property-input--wide">
              <span>{{ t('property.editor.addressEnField') }}</span>
              <input v-model="form.addressTextEn" />
            </label>
            <label class="property-input">
              <span>{{ t('property.editor.latitudeField') }}</span>
              <input
                v-model.number="form.latitude"
                type="number"
                step="0.000001"
              />
            </label>
            <label class="property-input">
              <span>{{ t('property.editor.longitudeField') }}</span>
              <input
                v-model.number="form.longitude"
                type="number"
                step="0.000001"
              />
            </label>
            <label
              v-if="form.transactionType === 'sale'"
              class="property-input"
            >
              <span>{{ t('property.editor.askingPriceField') }}</span>
              <input
                v-model.number="form.askingPriceHKD"
                type="number"
                min="0"
              />
            </label>
            <label
              v-else
              class="property-input"
            >
              <span>{{ t('property.editor.monthlyRentField') }}</span>
              <input
                v-model.number="form.monthlyRentHKD"
                type="number"
                min="0"
              />
            </label>
            <label class="property-input">
              <span>{{ t('property.editor.areaModeField') }}</span>
              <select v-model="form.areaMode">
                <option
                  v-for="mode in propertyAreaModeOptions"
                  :key="mode.value"
                  :value="mode.value"
                >
                  {{ getPropertyOptionLabel(mode, preferenceStore.locale) }}
                </option>
              </select>
            </label>
            <label class="property-input">
              <span>{{ t('property.editor.usableAreaField') }}</span>
              <input
                v-model.number="form.usableAreaSqft"
                type="number"
                min="0"
              />
            </label>
            <label class="property-input">
              <span>{{ t('property.editor.grossAreaField') }}</span>
              <input
                v-model.number="form.grossAreaSqft"
                type="number"
                min="0"
              />
            </label>
            <label class="property-input">
              <span>{{ t('property.editor.bedroomField') }}</span>
              <input
                v-model.number="form.bedroomCount"
                type="number"
                min="0"
              />
            </label>
            <label class="property-input">
              <span>{{ t('property.editor.livingRoomField') }}</span>
              <input
                v-model.number="form.livingRoomCount"
                type="number"
                min="0"
              />
            </label>
            <label class="property-input">
              <span>{{ t('property.editor.bathroomField') }}</span>
              <input
                v-model.number="form.bathroomCount"
                type="number"
                min="0"
              />
            </label>
            <label class="property-input">
              <span>{{ t('property.editor.floorField') }}</span>
              <input v-model="form.floorRaw" />
            </label>
            <label class="property-input">
              <span>{{ t('property.editor.floorZoneField') }}</span>
              <select v-model="form.floorZone">
                <option value="">
                  {{ t('property.editor.notSpecified') }}
                </option>
                <option
                  v-for="zone in propertyFloorZoneOptions"
                  :key="zone.value"
                  :value="zone.value"
                >
                  {{ getPropertyOptionLabel(zone, preferenceStore.locale) }}
                </option>
              </select>
            </label>
            <label class="property-input">
              <span>{{ t('property.editor.totalFloorsField') }}</span>
              <input
                v-model.number="form.totalFloors"
                type="number"
                min="0"
              />
            </label>
            <label class="property-input">
              <span>{{ t('property.editor.directionField') }}</span>
              <input v-model="form.direction" />
            </label>
            <label class="property-input">
              <span>{{ t('property.editor.buildingAgeField') }}</span>
              <input v-model="form.buildingAge" />
            </label>
            <label class="property-input">
              <span>{{ t('property.editor.completionYearField') }}</span>
              <input
                v-model.number="form.completionYear"
                type="number"
                min="0"
              />
            </label>
            <label class="property-input">
              <span>{{ t('property.editor.buildingTotalFloorsField') }}</span>
              <input
                v-model.number="form.buildingTotalFloors"
                type="number"
                min="0"
              />
            </label>
            <label class="property-input">
              <span>{{ t('property.editor.managementCompanyField') }}</span>
              <input v-model="form.managementCompany" />
            </label>
            <label class="property-input">
              <span>{{ t('property.editor.kitchenTypeField') }}</span>
              <select v-model="form.kitchenType">
                <option value="">
                  {{ t('property.editor.notSpecified') }}
                </option>
                <option
                  v-for="kitchen in propertyKitchenTypeOptions"
                  :key="kitchen.value"
                  :value="kitchen.value"
                >
                  {{ getPropertyOptionLabel(kitchen, preferenceStore.locale) }}
                </option>
              </select>
            </label>
            <label class="property-input">
              <span>{{ t('property.editor.cookingModeField') }}</span>
              <select v-model="form.cookingMode">
                <option value="">
                  {{ t('property.editor.notSpecified') }}
                </option>
                <option
                  v-for="mode in propertyCookingModeOptions"
                  :key="mode.value"
                  :value="mode.value"
                >
                  {{ getPropertyOptionLabel(mode, preferenceStore.locale) }}
                </option>
              </select>
            </label>
            <label class="property-input">
              <span>{{ t('property.editor.managementFeeField') }}</span>
              <input
                v-model.number="form.managementFeeHKD"
                type="number"
                min="0"
              />
            </label>
            <label class="property-input">
              <span>{{ t('property.editor.videoUrlField') }}</span>
              <input v-model="form.videoURL" />
            </label>
            <label class="property-input">
              <span>{{ t('property.editor.vrUrlField') }}</span>
              <input v-model="form.vrURL" />
            </label>
            <label class="property-input property-input--wide">
              <span>{{ t('property.editor.privateNoteField') }}</span>
              <textarea
                v-model="form.privateNote"
                rows="3"
              />
            </label>
          </div>
          <div class="property-checkbox-row">
            <label>
              <input
                v-model="form.showUnit"
                type="checkbox"
              />
              {{ t('property.editor.showUnit') }}
            </label>
            <label>
              <input
                v-model="form.priceReferenceOnly"
                type="checkbox"
              />
              {{ t('property.editor.priceReferenceOnly') }}
            </label>
            <label>
              <input
                v-model="form.priceNegotiable"
                type="checkbox"
              />
              {{ t('property.editor.priceNegotiable') }}
            </label>
            <label v-if="form.transactionType === 'rent'">
              <input
                v-model="form.annualPrepayDiscount"
                type="checkbox"
              />
              {{ t('property.editor.annualPrepayDiscount') }}
            </label>
          </div>
          <div
            v-if="form.transactionType === 'rent'"
            class="property-editor-grid property-editor-grid--nested"
          >
            <label class="property-input">
              <span>{{ t('property.editor.annualPrepayOptionField') }}</span>
              <select v-model="form.annualPrepayOption">
                <option
                  v-for="option in propertyAnnualPrepayOptions"
                  :key="option.value"
                  :value="option.value"
                >
                  {{ getPropertyOptionLabel(option, preferenceStore.locale) }}
                </option>
              </select>
            </label>
            <label class="property-input">
              <span>{{ t('property.editor.leaseStartDateField') }}</span>
              <input
                v-model="form.leaseStartDate"
                type="date"
              />
            </label>
            <label class="property-input property-input--wide">
              <span>{{ t('property.editor.rentIncludedField') }}</span>
              <input v-model="form.rentIncluded" />
            </label>
          </div>
          <div class="property-checkbox-row">
            <label
              v-for="tag in propertyFeatureTagOptions"
              :key="tag.value"
            >
              <input
                type="checkbox"
                :checked="form.featureTags.includes(tag.value)"
                @change="toggleTag(form.featureTags, tag.value)"
              />
              {{ getPropertyOptionLabel(tag, preferenceStore.locale) }}
            </label>
          </div>
        </section>

        <section
          v-else
          class="property-editor-panel"
        >
          <h2>{{ t('property.editor.servicedTitle') }}</h2>
          <div class="property-editor-grid">
            <label class="property-input property-input--wide">
              <span>{{ t('property.editor.projectNameField') }}</span>
              <span class="property-address-input">
                <input
                  v-model="form.projectName"
                  autocomplete="off"
                  @focus="scheduleAddressSuggestions"
                  @input="scheduleAddressSuggestions"
                />
                <i
                  v-if="loadingAddressSuggestions"
                  aria-hidden="true"
                />
                <span
                  v-if="addressSuggestions.length > 0 || loadingAddressSuggestions"
                  class="property-address-suggestions"
                >
                  <span
                    v-if="loadingAddressSuggestions"
                    class="property-address-loading"
                  >
                    {{ t('common.status.loading') }}
                  </span>
                  <button
                    v-for="suggestion in addressSuggestions"
                    :key="suggestion.address_id"
                    type="button"
                    @click="applyAddressSuggestion(suggestion)"
                  >
                    <strong>{{ resolveAddressSuggestionTitle(suggestion) }}</strong>
                    <span>{{ suggestion.address_text }}</span>
                  </button>
                </span>
              </span>
            </label>
            <label class="property-input">
              <span>{{ t('property.editor.projectNameEnField') }}</span>
              <input v-model="form.projectNameEn" />
            </label>
            <label class="property-input">
              <span>{{ t('property.editor.monthlyRentField') }}</span>
              <input
                v-model.number="form.lowestMonthlyRentHKD"
                type="number"
                min="0"
              />
            </label>
            <label class="property-input">
              <span>{{ t('property.editor.dailyRentField') }}</span>
              <input
                v-model.number="form.lowestDailyRentHKD"
                type="number"
                min="0"
              />
            </label>
            <label class="property-input property-input--wide">
              <span>{{ t('property.editor.addressField') }}</span>
              <input v-model="form.addressText" />
            </label>
            <label class="property-input property-input--wide">
              <span>{{ t('property.editor.addressEnField') }}</span>
              <input v-model="form.addressTextEn" />
            </label>
            <label class="property-input">
              <span>{{ t('property.editor.websiteField') }}</span>
              <input v-model="form.websiteURL" />
            </label>
            <label class="property-input">
              <span>{{ t('property.editor.serviceWhatsappField') }}</span>
              <input v-model="form.serviceWhatsApp" />
            </label>
            <label class="property-input">
              <span>{{ t('property.editor.faxField') }}</span>
              <input v-model="form.fax" />
            </label>
            <label class="property-input">
              <span>{{ t('property.editor.minLeaseField') }}</span>
              <input
                v-model.number="form.minLeaseMonths"
                type="number"
                min="1"
              />
            </label>
            <label class="property-input">
              <span>{{ t('property.editor.minStayValueField') }}</span>
              <input
                v-model.number="form.minStayValue"
                type="number"
                min="1"
              />
            </label>
            <label class="property-input">
              <span>{{ t('property.editor.minStayUnitField') }}</span>
              <select v-model="form.minStayUnit">
                <option
                  v-for="unit in servicedStayUnitOptions"
                  :key="unit.value"
                  :value="unit.value"
                >
                  {{ getPropertyOptionLabel(unit, preferenceStore.locale) }}
                </option>
              </select>
            </label>
            <label class="property-input">
              <span>{{ t('property.editor.minUsableAreaField') }}</span>
              <input
                v-model.number="form.minUsableAreaSqft"
                type="number"
                min="0"
              />
            </label>
            <section class="property-room-types property-input--wide">
              <header class="property-room-types__header">
                <h3>{{ t('property.editor.roomTypesSection') }}</h3>
                <button
                  type="button"
                  class="property-mini-button"
                  @click="addServicedRoomType"
                >
                  {{ t('property.editor.addRoomType') }}
                </button>
              </header>
              <article
                v-for="(room, roomIndex) in form.roomTypes"
                :key="roomIndex"
                class="property-room-type"
              >
                <div class="property-room-type__title">
                  <strong>{{ t('property.editor.roomTypeName') }} {{ roomIndex + 1 }}</strong>
                  <button
                    type="button"
                    class="property-mini-button property-mini-button--secondary"
                    :disabled="form.roomTypes.length <= 1"
                    @click="removeServicedRoomType(roomIndex)"
                  >
                    {{ t('property.editor.removeRoomType') }}
                  </button>
                </div>
                <div class="property-editor-grid property-editor-grid--nested">
                  <label class="property-input">
                    <span>{{ t('property.editor.roomTypeName') }}</span>
                    <input v-model="room.name" />
                  </label>
                  <label class="property-input">
                    <span>{{ t('property.editor.roomCategoryField') }}</span>
                    <input v-model="room.room_category" />
                  </label>
                  <label class="property-input">
                    <span>{{ t('property.editor.roomTypeArea') }}</span>
                    <input
                      v-model.number="room.usable_area_sqft"
                      type="number"
                      min="0"
                    />
                  </label>
                  <label class="property-input">
                    <span>{{ t('property.editor.roomTypeRent') }}</span>
                    <input
                      v-model.number="room.monthly_rent_min_hkd"
                      type="number"
                      min="0"
                    />
                  </label>
                  <label class="property-input">
                    <span>{{ t('property.editor.roomMonthlyRentMaxField') }}</span>
                    <input
                      v-model.number="room.monthly_rent_max_hkd"
                      type="number"
                      min="0"
                    />
                  </label>
                  <label class="property-input">
                    <span>{{ t('property.editor.roomDailyRentMinField') }}</span>
                    <input
                      v-model.number="room.daily_rent_min_hkd"
                      type="number"
                      min="0"
                    />
                  </label>
                  <label class="property-input">
                    <span>{{ t('property.editor.roomDailyRentMaxField') }}</span>
                    <input
                      v-model.number="room.daily_rent_max_hkd"
                      type="number"
                      min="0"
                    />
                  </label>
                  <label class="property-input">
                    <span>{{ t('property.editor.roomMinLeaseField') }}</span>
                    <input
                      v-model.number="room.min_lease_months"
                      type="number"
                      min="1"
                    />
                  </label>
                  <label class="property-input">
                    <span>{{ t('property.editor.roomMinStayValueField') }}</span>
                    <input
                      v-model.number="room.min_stay_value"
                      type="number"
                      min="1"
                    />
                  </label>
                  <label class="property-input">
                    <span>{{ t('property.editor.roomMinStayUnitField') }}</span>
                    <select v-model="room.min_stay_unit">
                      <option
                        v-for="unit in servicedStayUnitOptions"
                        :key="unit.value"
                        :value="unit.value"
                      >
                        {{ getPropertyOptionLabel(unit, preferenceStore.locale) }}
                      </option>
                    </select>
                  </label>
                  <label class="property-input">
                    <span>{{ t('property.editor.includedFeeItemsField') }}</span>
                    <input
                      :value="room.included_fee_items?.join('、') || ''"
                      @input="updateRoomIncludedFeeItems(room, $event)"
                    />
                  </label>
                  <label class="property-inline-checkbox">
                    <input
                      v-model="room.included_fees"
                      type="checkbox"
                    />
                    {{ t('property.editor.includedFees') }}
                  </label>
                </div>
              </article>
            </section>
            <label class="property-input property-input--wide">
              <span>{{ t('property.editor.serviceIntroField') }}</span>
              <textarea
                v-model="form.serviceIntro"
                rows="3"
              />
            </label>
            <label class="property-input property-input--wide">
              <span>{{ t('property.editor.benefitsField') }}</span>
              <textarea
                v-model="form.benefitsText"
                rows="3"
              />
            </label>
            <label class="property-input property-input--wide">
              <span>{{ t('property.editor.extraChargesField') }}</span>
              <textarea
                v-model="form.extraChargesText"
                rows="3"
              />
            </label>
          </div>
          <div class="property-checkbox-row">
            <label>
              <input
                v-model="form.priceReferenceOnlyServiced"
                type="checkbox"
              />
              {{ t('property.editor.priceReferenceOnly') }}
            </label>
            <label>
              <input
                v-model="form.priceNegotiableServiced"
                type="checkbox"
              />
              {{ t('property.editor.priceNegotiable') }}
            </label>
          </div>

          <div class="property-checkbox-row">
            <label
              v-for="tag in servicedFacilityTagOptions"
              :key="tag.value"
            >
              <input
                type="checkbox"
                :checked="form.facilityTags.includes(tag.value)"
                @change="toggleTag(form.facilityTags, tag.value)"
              />
              {{ getPropertyOptionLabel(tag, preferenceStore.locale) }}
            </label>
          </div>
          <div class="property-checkbox-row">
            <label
              v-for="tag in servicedServiceTagOptions"
              :key="tag.value"
            >
              <input
                type="checkbox"
                :checked="form.serviceTags.includes(tag.value)"
                @change="toggleTag(form.serviceTags, tag.value)"
              />
              {{ getPropertyOptionLabel(tag, preferenceStore.locale) }}
            </label>
          </div>
        </section>

        <section class="property-editor-panel">
          <h2>{{ t('property.editor.media') }}</h2>
          <p class="property-editor-note">
            {{ t('property.editor.imageUploadNote') }}
          </p>
          <label class="property-upload-button">
            <AppIcon
              name="cloud-upload"
              :size="18"
            />
            {{ t('property.editor.selectImages') }}
            <input
              type="file"
              accept="image/*"
              multiple
              @change="handleImageFilesChange"
            />
          </label>
          <div class="property-image-grid">
            <article
              v-for="image in images"
              :key="image.id"
              class="property-image-card"
            >
              <img
                v-if="image.url"
                :src="image.url"
                :alt="form.title"
              />
              <div class="property-image-card__actions">
                <button
                  type="button"
                  :class="{ 'property-image-card__action--active': image.isCover }"
                  @click="selectCover(image.id)"
                >
                  {{ image.isCover ? '封面' : '設為封面' }}
                </button>
                <button
                  type="button"
                  @click="removeImage(image.id)"
                >
                  移除
                </button>
              </div>
            </article>
          </div>
        </section>

        <section class="property-editor-panel">
          <h2>{{ t('property.editor.contact') }}</h2>
          <div class="property-editor-grid">
            <label class="property-input">
              <span>{{ t('property.editor.contactMethodField') }}</span>
              <select v-model="form.contactMethod">
                <option
                  v-for="method in propertyContactMethodOptions"
                  :key="method.value"
                  :value="method.value"
                >
                  {{ getPropertyOptionLabel(method, preferenceStore.locale) }}
                </option>
              </select>
            </label>
            <label class="property-input">
              <span>{{ t('property.editor.phoneField') }}</span>
              <input v-model="form.phone" />
            </label>
            <label class="property-input">
              <span>{{ t('property.editor.whatsappField') }}</span>
              <input v-model="form.whatsapp" />
            </label>
            <label class="property-input">
              <span>{{ t('property.editor.emailField') }}</span>
              <input v-model="form.email" />
            </label>
          </div>
          <div class="property-checkbox-row">
            <label>
              <input
                v-model="form.allowPhone"
                type="checkbox"
              />
              {{ t('property.editor.allowPhone') }}
            </label>
            <label>
              <input
                v-model="form.allowWhatsapp"
                type="checkbox"
              />
              {{ t('property.editor.allowWhatsapp') }}
            </label>
            <label>
              <input
                v-model="form.allowChat"
                type="checkbox"
              />
              {{ t('property.editor.allowChat') }}
            </label>
          </div>
        </section>
      </form>

      <aside class="property-editor-side">
        <section class="property-editor-panel">
          <p class="property-side-kicker">Preview</p>
          <h2>{{ form.title || pageTitle }}</h2>
          <p>{{ form.summary || t('property.editor.requiredFields') }}</p>
          <strong>{{ previewPrice }}</strong>
          <div class="property-side-meta">
            <div>
              <span>{{ t('property.editor.adPackageField') }}</span>
              <strong>{{ getPropertyOptionLabel(selectedAdPackage, preferenceStore.locale) }}</strong>
            </div>
            <div>
              <span>{{ t('property.editor.media') }}</span>
              <strong>{{ images.length }}</strong>
            </div>
          </div>
          <p class="property-editor-charge">
            {{ chargeHint }}
          </p>
          <button
            type="button"
            class="property-editor-action property-editor-action--secondary"
            :disabled="saving || publishing"
            @click="saveAndReturn"
          >
            {{ saving ? t('common.status.loading') : t('property.editor.saveDraft') }}
          </button>
          <button
            type="button"
            class="property-editor-action property-editor-action--primary"
            :disabled="saving || publishing"
            @click="saveAndPublish"
          >
            {{ publishing ? t('common.status.loading') : props.staffMode ? t('property.editor.saveDraft') : t('property.editor.publishNow') }}
          </button>
        </section>
      </aside>
    </section>

    <AppUnsavedChangesDialog
      :open="isLeavePromptOpen"
      :title="t('property.editor.unsavedLeaveTitle')"
      :description="t('property.editor.unsavedLeaveDescription')"
      :save-label="saving ? t('property.editor.savingDraft') : t('property.editor.unsavedLeaveSave')"
      :discard-label="t('property.editor.unsavedLeaveDiscard')"
      :stay-label="t('property.editor.unsavedLeaveStay')"
      :saving="saving"
      @save="handleLeavePromptDecision('save')"
      @discard="handleLeavePromptDecision('discard')"
      @stay="handleLeavePromptDecision('stay')"
    />
  </main>
</template>

<style scoped>
.property-editor-page {
  width: 100%;
  max-width: min(1240px, calc(100vw - 32px));
  margin: 0 auto;
  padding: 10px 0 72px;
  color: rgb(var(--color-text));
}

.property-editor-heading {
  display: flex;
  align-items: end;
  justify-content: space-between;
  gap: 16px;
  border-bottom: 1px solid rgb(var(--color-border));
  margin-bottom: 14px;
  padding-bottom: 18px;
}

.property-kicker {
  margin: 0;
  color: rgb(var(--color-primary));
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0.18em;
  text-transform: uppercase;
}

.property-editor-heading h1 {
  margin: 6px 0 0;
  font-family: var(--font-display);
  font-size: 32px;
  font-weight: 400;
}

.property-editor-heading__meta {
  display: grid;
  gap: 4px;
  min-width: 120px;
  text-align: right;
}

.property-editor-heading__meta span {
  color: rgb(var(--color-primary));
  font-family: var(--font-display);
  font-size: 28px;
  font-weight: 400;
  line-height: 1;
}

.property-editor-layout {
  display: grid;
  gap: 12px;
}

.property-editor-form {
  display: grid;
  gap: 12px;
}

.property-editor-panel {
  border: 1px solid rgb(var(--color-border));
  border-radius: 3px;
  background: rgb(var(--color-surface));
  padding: 12px;
}

.property-editor-panel h2 {
  margin: 0 0 1rem;
  color: rgb(var(--color-primary));
  font-size: 14px;
  font-weight: 600;
}

.property-editor-panel p {
  color: rgb(var(--color-text-muted));
  font-size: 12px;
  line-height: 1.6;
}

.property-editor-panel strong {
  display: block;
  margin: 1rem 0;
  font-family: var(--font-display);
  font-size: 28px;
  font-weight: 400;
}

.property-editor-charge {
  margin: 0 0 1rem;
  color: rgb(var(--color-text));
  font-size: 12px;
  font-weight: 600;
}

.property-editor-grid {
  display: grid;
  gap: 10px;
}

.property-editor-grid--nested,
.property-input--compact {
  margin-top: 10px;
}

.property-editor-note {
  margin: 8px 0 0;
  color: rgb(var(--color-text-muted));
  font-size: 12px;
  line-height: 1.6;
}

.property-ad-package-grid {
  display: grid;
  gap: 8px;
}

.property-ad-package-grid button {
  display: grid;
  gap: 6px;
  min-height: 88px;
  border: 1px solid rgb(var(--color-border));
  border-radius: 3px;
  background: rgb(var(--color-surface));
  padding: 12px;
  color: rgb(var(--color-text));
  text-align: left;
  cursor: pointer;
  transition:
    border-color 0.2s ease,
    background 0.2s ease;
}

.property-ad-package-grid button:hover {
  border-color: rgb(var(--color-primary) / 0.35);
}

.property-ad-package-grid .property-ad-package--active {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary) / 0.08);
}

.property-ad-package-grid strong {
  margin: 0;
  font-family: var(--font-sans);
  font-size: 13px;
  font-weight: 700;
}

.property-ad-package-grid span,
.property-ad-package-grid small {
  color: rgb(var(--color-text-muted));
  font-size: 11px;
  font-weight: 600;
}

.property-input {
  display: grid;
  gap: 6px;
  color: rgb(var(--color-text-muted));
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0.14em;
  text-transform: uppercase;
}

.property-input input,
.property-input select,
.property-input textarea {
  width: 100%;
  min-height: 36px;
  border: 1px solid rgb(var(--color-border));
  border-radius: 2px;
  background: rgb(var(--color-surface));
  padding: 8px 10px;
  color: rgb(var(--color-text));
  font: inherit;
  font-size: 12px;
  letter-spacing: 0;
  text-transform: none;
  outline: 0;
  transition:
    border-color 0.2s ease,
    box-shadow 0.2s ease;
}

.property-input input:focus,
.property-input select:focus,
.property-input textarea:focus {
  border-color: rgb(var(--color-primary));
  box-shadow: 0 0 0 3px rgb(var(--color-primary) / 0.1);
}

.property-input textarea {
  resize: vertical;
}

.property-address-input {
  position: relative;
  display: block;
}

.property-address-input i {
  position: absolute;
  top: 50%;
  right: 0.8rem;
  width: 0.85rem;
  height: 0.85rem;
  border: 2px solid rgb(var(--color-border));
  border-top-color: rgb(var(--color-primary));
  border-radius: 999px;
  animation: property-address-spin 0.8s linear infinite;
  transform: translateY(-50%);
}

.property-address-suggestions {
  position: absolute;
  z-index: 20;
  top: calc(100% + 0.35rem);
  right: 0;
  left: 0;
  display: grid;
  max-height: 16rem;
  overflow-y: auto;
  border: 1px solid rgb(var(--color-border));
  border-radius: 3px;
  background: rgb(var(--color-surface));
  box-shadow: 0 18px 42px rgb(15 23 42 / 0.12);
}

.property-address-loading,
.property-address-suggestions button {
  padding: 9px 10px;
}

.property-address-suggestions button {
  display: grid;
  gap: 4px;
  border: 0;
  border-bottom: 1px solid rgb(var(--color-border));
  background: transparent;
  color: rgb(var(--color-text));
  text-align: left;
  cursor: pointer;
}

.property-address-suggestions button:hover,
.property-address-suggestions button:focus {
  background: rgb(var(--color-surface-raised));
}

.property-address-suggestions strong {
  font-size: 12px;
  font-weight: 700;
}

.property-address-suggestions button span,
.property-address-loading {
  color: rgb(var(--color-text-muted));
  font-size: 11px;
  font-weight: 600;
}

.property-checkbox-row {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 10px;
}

.property-checkbox-row label,
.property-inline-checkbox {
  display: inline-flex;
  min-height: 34px;
  align-items: center;
  gap: 0.45rem;
  border: 1px solid rgb(var(--color-border));
  border-radius: 2px;
  background: rgb(var(--color-surface));
  padding: 0 10px;
  font-size: 11px;
  font-weight: 600;
}

.property-room-types {
  display: grid;
  gap: 10px;
}

.property-room-types__header,
.property-room-type__title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.property-room-types__header h3 {
  margin: 0;
  color: rgb(var(--color-primary));
  font-size: 13px;
  font-weight: 600;
  letter-spacing: 0;
  text-transform: none;
}

.property-room-type {
  display: grid;
  gap: 10px;
  border: 1px solid rgb(var(--color-border));
  border-radius: 3px;
  background: rgb(var(--color-surface-muted));
  padding: 10px;
}

.property-room-type__title strong {
  margin: 0;
  color: rgb(var(--color-text));
  font-family: var(--font-sans);
  font-size: 12px;
  font-weight: 700;
}

.property-mini-button {
  min-height: 30px;
  border: 1px solid rgb(var(--color-primary));
  border-radius: 2px;
  background: rgb(var(--color-primary));
  padding: 0 10px;
  color: rgb(var(--color-primary-contrast));
  font-size: 11px;
  font-weight: 600;
  cursor: pointer;
}

.property-mini-button--secondary {
  border-color: rgb(var(--color-border));
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text));
}

.property-mini-button:disabled {
  cursor: not-allowed;
  opacity: 0.45;
}

.property-upload-button,
.property-editor-action {
  display: inline-flex;
  min-height: 34px;
  align-items: center;
  justify-content: center;
  gap: 6px;
  border-radius: 2px;
  padding: 0 12px;
  font-size: 11px;
  font-weight: 600;
}

.property-upload-button {
  border: 1px solid rgb(var(--color-primary));
  background: rgb(var(--color-surface));
  color: rgb(var(--color-primary));
  cursor: pointer;
}

.property-upload-button input {
  display: none;
}

.property-image-grid {
  display: grid;
  gap: 8px;
  margin-top: 10px;
}

.property-image-card {
  overflow: hidden;
  border: 1px solid rgb(var(--color-border));
  border-radius: 3px;
  background: rgb(var(--color-surface-muted));
}

.property-image-card img {
  width: 100%;
  aspect-ratio: 4 / 3;
  object-fit: cover;
}

.property-image-card__actions {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.property-image-card button {
  min-height: 32px;
  border: 0;
  border-top: 1px solid rgb(var(--color-border));
  background: transparent;
  color: rgb(var(--color-text));
  font-size: 11px;
  font-weight: 600;
}

.property-image-card button + button {
  border-left: 1px solid rgb(var(--color-border));
}

.property-image-card__action--active {
  color: rgb(var(--color-primary));
}

.property-editor-side {
  display: grid;
  align-content: start;
  gap: 10px;
}

.property-side-kicker {
  margin: 0 0 8px;
  color: rgb(var(--color-primary));
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 0.14em;
  text-transform: uppercase;
}

.property-side-meta {
  display: grid;
  gap: 10px;
  margin-bottom: 12px;
}

.property-side-meta div {
  display: grid;
  gap: 4px;
  border-top: 1px solid rgb(var(--color-border));
  padding-top: 10px;
}

.property-side-meta span {
  color: rgb(var(--color-text-muted));
  font-size: 12px;
  font-weight: 600;
}

.property-side-meta strong {
  margin: 0;
  color: rgb(var(--color-text));
  font-family: var(--font-sans);
  font-size: 13px;
  font-weight: 600;
}

.property-editor-action {
  width: 100%;
  border: 1px solid rgb(var(--color-border));
}

.property-editor-action--primary {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary));
  color: rgb(var(--color-primary-contrast));
}

.property-editor-action--secondary {
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text));
}

@keyframes property-address-spin {
  to {
    transform: translateY(-50%) rotate(360deg);
  }
}

@media (min-width: 760px) {
  .property-editor-grid,
  .property-image-grid,
  .property-ad-package-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .property-input--wide {
    grid-column: 1 / -1;
  }
}

@media (min-width: 1080px) {
  .property-editor-layout {
    grid-template-columns: minmax(0, 1fr) 22rem;
    align-items: start;
  }

  .property-editor-side {
    position: sticky;
    top: calc(var(--app-header-offset, 0rem) + 2rem);
  }
}

.property-editor-page {
  padding: 18px var(--layout-page-padding-inline) 72px;
}

.property-editor-heading {
  border-bottom: 1px solid rgb(var(--color-border));
  margin-bottom: 14px;
  padding-bottom: 18px;
}

.property-kicker {
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0.18em;
}

.property-editor-heading h1 {
  font-size: 32px;
  font-weight: 400;
}

.property-editor-heading p:last-child {
  color: rgb(var(--color-primary));
  font-size: 13px;
  font-weight: 600;
}

.property-editor-panel {
  border-radius: 3px;
  box-shadow: none;
  padding: 14px;
}

.property-editor-panel h2 {
  color: rgb(var(--color-primary));
  font-size: 14px;
  font-weight: 600;
}

.property-editor-panel p,
.property-editor-charge,
.property-editor-note {
  font-size: 12px;
  line-height: 1.6;
}

.property-editor-panel strong {
  font-family: var(--font-display);
  font-size: 22px;
  font-weight: 400;
}

@media (min-width: 1080px) {
  .property-editor-side {
    top: 66px;
  }
}

@media (max-width: 767px) {
  .property-editor-heading {
    align-items: start;
    flex-direction: column;
  }

  .property-editor-heading__meta {
    text-align: left;
  }

  .property-editor-page {
    max-width: 100%;
    padding-bottom: 96px;
  }
}
</style>

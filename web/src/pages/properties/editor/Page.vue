<!--
 * 樓盤租售發布編輯頁。
 * 1. 建立或更新樓盤租售草稿。
 * 2. 支援圖片上傳、草稿保存與發布。
-->
<script setup lang="ts">
import axios from 'axios';
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue';
import type { RouteLocationRaw } from 'vue-router';
import { useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';

import { fetchCommunities } from '@/domains/building/api';
import {
  fetchStaffPropertySaleDetail,
  fetchStaffServicedApartmentDetail,
  updateStaffPropertySale,
  updateStaffServicedApartment,
} from '@/domains/management/staff-api';
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
} from '@/domains/property/api';
import { completeUpload, createUploadPresign } from '@/domains/media/uploads-api';
import {
  getPropertyOptionLabel,
  marketplaceDistricts,
  propertyAdPackageOptions,
  propertyContactMethodOptions,
  propertyFloorZoneOptions,
  propertyListingCategoryOptions,
  propertyLocationScopeOptions,
  servicedAdPackageOptions,
  servicedFacilityTagOptions,
  servicedServiceTagOptions,
  propertyTagGroups,
  propertyTransactionTypeOptions,
} from '@/domains/property/constants';
import type { MetaCommunity } from '@/domains/building/model';
import type { MediaAssetResponse } from '@/domains/marketplace/model';
import type {
  PropertyChannel,
  PropertyAddressSuggestion,
  PropertyImagePayload,
  PropertyListingDetailResponse,
  PropertySalePayload,
  ServicedApartmentPayload,
  ServicedApartmentRoomType,
  UpsertPropertySalePayload,
  UpsertServicedApartmentPayload,
} from '@/domains/property/model';
import AppIcon from '@/shared/components/base/AppIcon.vue';
import { useFeedbackStore } from '@/app/stores/feedback';
import { usePreferenceStore } from '@/app/stores/preferences';
import { useSessionStore } from '@/app/stores/session';
import { mergePropertyFeatureTags, resolvePropertyPriceText } from '@/shared/utils/property';
import { buildUploadHeaders } from '@/shared/utils/upload';
import { formatPrice } from '@/shared/utils/format';
import { formatAjoPoints } from '@/shared/utils/wallet';

const props = withDefaults(defineProps<{
  channel: PropertyChannel;
  listingId?: string;
  embedded?: boolean;
  staffMode?: boolean;
}>(), {
  listingId: '',
  embedded: false,
  staffMode: false,
});

const emit = defineEmits<{
  saved: [listingId: string];
  published: [listingId: string];
  cancel: [];
}>();

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
  propertyNo: string;
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
  contactPersonName: string;
  contactPersonNameEn: string;
  phone: string;
  whatsapp: string;
  email: string;
  allowPhone: boolean;
  allowWhatsapp: boolean;
  allowChat: boolean;
  transactionType: 'sale' | 'rent';
  locationScope: 'local' | 'overseas';
  listingCategory: string;
  multiUnitProject: boolean;
  propertyType: string;
  rentalType: string;
  estateName: string;
  addressText: string;
  addressTextEn: string;
  blockName: string;
  unitName: string;
  showUnit: boolean;
  askingPriceHKD: number;
  monthlyRentHKD: number;
  priceReferenceOnly: boolean;
  priceNegotiable: boolean;
  annualPrepayOption: 'none' | '95_off' | '90_off';
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
  floorZone: 'low' | 'middle' | 'high';
  totalFloors: number;
  direction: string;
  buildingAge: string;
  kitchenType: string;
  cookingMode: string;
  managementFeeHKD: number;
  videoUrl: string;
  vrUrl: string;
  privateNote: string;
  adPackageCode: 'basic' | 'featured' | 'premium' | 'fast_sale';
  featureTags: string[];
  projectName: string;
  projectNameEn: string;
  servicedAddressTextEn: string;
  websiteUrl: string;
  servicedWhatsapp: string;
  fax: string;
  serviceIntro: string;
  benefitsText: string;
  extraChargesText: string;
  residenceInfoTags: string[];
  extraChargeTags: string[];
  lowestMonthlyRentHKD: number;
  lowestDailyRentHKD: number;
  minUsableAreaSqft: number;
  minLeaseMonths: number;
  minStayValue: number;
  minStayUnit: 'day' | 'month';
  facilityTags: string[];
  serviceTags: string[];
  roomTypes: ServicedApartmentRoomType[];
}

type PropertyEditorStep = 'category' | 'content' | 'preview' | 'complete';
type AppIconName = InstanceType<typeof AppIcon>['$props']['name'];
type PropertyValidationField =
  | 'publisherIdentityType'
  | 'transactionType'
  | 'locationScope'
  | 'listingCategory'
  | 'adPackageCode'
  | 'estateName'
  | 'addressText'
  | 'addressTextEn'
  | 'askingPriceHKD'
  | 'monthlyRentHKD'
  | 'usableAreaSqft'
  | 'title'
  | 'titleEn'
  | 'description'
  | 'descriptionEn'
  | 'floorRaw'
  | 'projectName'
  | 'roomTypes'
  | 'lowestMonthlyRentHKD';

interface PropertyEditorStepOption {
  value: PropertyEditorStep;
  labelKey: string;
}

const maxImages = 40;
const maxImageSize = 10 * 1024 * 1024;

const route = useRoute();
const router = useRouter();
const { t } = useI18n();
const feedbackStore = useFeedbackStore();
const preferenceStore = usePreferenceStore();
const sessionStore = useSessionStore();

const propertyPublisherOptions = [
  { value: 'owner', label_zh_hk: '業主', label_en: 'Owner' },
  { value: 'agent', label_zh_hk: '地產代理', label_en: 'Estate agent' },
];
const rentIncludedOptions = ['差餉地租', '管理費', '煤氣', '水電'];
const annualPrepayOptions = [
  { value: 'none', label_zh_hk: '無', label_en: 'None' },
  { value: '95_off', label_zh_hk: '95折', label_en: '5% off' },
  { value: '90_off', label_zh_hk: '9折', label_en: '10% off' },
] as const;
const servicedResidenceInfoQuickTags = [
  '每月租金已包水電差餉及管理費',
  '每週一次房間清潔服務（星期日及公眾假期除外）',
  '有上網連接的智能電視',
  '有免費無線上網提供',
  '夾萬',
  '設備齊全的廚房設施',
  '部份單位設有私人露台',
  '每天日報',
  '早餐',
];
const servicedExtraChargeQuickTags = [
  '可攜寵物入住（只需繳付額外一個月按金）',
];

// 1. 取得會員電話預設值
const resolveMemberPhone = (): string => {
  const countryCode = sessionStore.me?.phone_country_code?.trim() ?? '';
  const phoneNumber = sessionStore.me?.phone_number?.trim() ?? '';
  return [countryCode, phoneNumber].filter(Boolean).join(' ');
};

const activeListingId = ref(props.listingId || String(route.params.listingId ?? ''));
const loading = ref(false);
const saving = ref(false);
const publishing = ref(false);
const communities = ref<MetaCommunity[]>([]);
const images = ref<PropertyEditorImage[]>([]);
const addressSuggestions = ref<PropertyAddressSuggestion[]>([]);
const loadingAddressSuggestions = ref(false);
const addressSearchTimer = ref<number | null>(null);
const activeStep = ref<PropertyEditorStep>(props.channel === 'serviced' ? 'content' : 'category');
const completedListingId = ref('');
const validationErrors = reactive<Partial<Record<PropertyValidationField, string>>>({});
const servicedTagDrafts = reactive({
  residenceInfo: '',
  facility: '',
  service: '',
  extraCharge: '',
});

const propertyEditorSteps: PropertyEditorStepOption[] = [
  { value: 'category', labelKey: 'property.editor.stepCategory' },
  { value: 'content', labelKey: 'property.editor.stepContent' },
  { value: 'preview', labelKey: 'property.editor.stepPreview' },
  { value: 'complete', labelKey: 'property.editor.stepComplete' },
];
const servicedEditorSteps: PropertyEditorStepOption[] = [
  { value: 'category', labelKey: 'property.editor.servicedStepCategory' },
  { value: 'content', labelKey: 'property.editor.servicedStepContent' },
  { value: 'preview', labelKey: 'property.editor.stepPreview' },
  { value: 'complete', labelKey: 'property.editor.stepComplete' },
];

const form = reactive<PropertyEditorForm>({
  propertyNo: '',
  title: '',
  titleEn: '',
  summary: '',
  description: '',
  descriptionEn: '',
  districtCode: sessionStore.me?.district_code || 'kowloon',
  communityId: sessionStore.currentUser.primary_community.public_id || '',
  publisherIdentityType: sessionStore.me?.publisher_identity_type || 'owner',
  businessStatus: 'available',
  contactMethod: 'both',
  contactPersonName: sessionStore.me?.display_name || '',
  contactPersonNameEn: '',
  phone: resolveMemberPhone(),
  whatsapp: resolveMemberPhone(),
  email: sessionStore.me?.email || '',
  allowPhone: true,
  allowWhatsapp: true,
  allowChat: false,
  transactionType: 'sale',
  locationScope: 'local',
  listingCategory: 'standard',
  multiUnitProject: false,
  propertyType: 'residential',
  rentalType: '',
  estateName: '',
  addressText: '',
  addressTextEn: '',
  blockName: '',
  unitName: '',
  showUnit: true,
  askingPriceHKD: 0,
  monthlyRentHKD: 0,
  priceReferenceOnly: false,
  priceNegotiable: false,
  annualPrepayOption: 'none',
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
  floorZone: 'middle',
  totalFloors: 0,
  direction: '',
  buildingAge: '',
  kitchenType: 'NA',
  cookingMode: 'NA',
  managementFeeHKD: 0,
  videoUrl: '',
  vrUrl: '',
  privateNote: '',
  adPackageCode: 'basic',
  featureTags: [],
  projectName: '',
  projectNameEn: '',
  servicedAddressTextEn: '',
  websiteUrl: '',
  servicedWhatsapp: resolveMemberPhone(),
  fax: '',
  serviceIntro: '',
  benefitsText: '',
  extraChargesText: '',
  residenceInfoTags: [],
  extraChargeTags: [],
  lowestMonthlyRentHKD: 0,
  lowestDailyRentHKD: 0,
  minUsableAreaSqft: 0,
  minLeaseMonths: 1,
  minStayValue: 1,
  minStayUnit: 'month',
  facilityTags: [],
  serviceTags: [],
  roomTypes: [{
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
  }],
});

const isSale = computed(() => props.channel === 'sale');
const isRentalSaleListing = computed(() => isSale.value && form.transactionType === 'rent');
const isEditing = computed(() => activeListingId.value.trim().length > 0);
const pageTitle = computed(() =>
  isSale.value ? t('property.sale.publishTitle') : t('property.serviced.publishTitle'),
);
const myPath = computed<RouteLocationRaw>(() =>
  isSale.value
    ? { path: '/member', query: { tab: 'properties' } }
    : { path: '/member', query: { tab: 'serviced-residences' } },
);
const hasImage = computed(() => images.value.some((image) => image.mediaAssetId || image.file));
const hasPersistedImage = computed(() => images.value.some((image) => image.mediaAssetId));
const activeAdPackageOptions = computed(() => isSale.value ? propertyAdPackageOptions : servicedAdPackageOptions);
const selectedAdPackage = computed(() =>
  activeAdPackageOptions.value.find((option) => option.value === form.adPackageCode) ?? activeAdPackageOptions.value[0],
);
const visibleEditorSteps = computed(() => isSale.value ? propertyEditorSteps : servicedEditorSteps);
const activeStepIndex = computed(() =>
  Math.max(0, visibleEditorSteps.value.findIndex((step) => step.value === activeStep.value)),
);
const publicFloorPreview = computed(() => {
  const label = propertyFloorZoneOptions.find((option) => option.value === form.floorZone);
  const zoneText = label ? getPropertyOptionLabel(label, preferenceStore.locale) : form.floorZone;
  if (form.showUnit && form.unitName.trim()) {
    return `${form.blockName.trim()}${zoneText}${form.unitName.trim()}`;
  }
  if (!form.showUnit && form.totalFloors > 0) {
    return `${form.blockName.trim()}${zoneText}(${buildFloorRangePreview()})`;
  }
  return `${form.blockName.trim()}${zoneText}`;
});
const buildEmptySalePayload = (): PropertySalePayload => ({
  property_no: '',
  transaction_type: form.transactionType,
  location_scope: form.locationScope,
  listing_category: form.listingCategory,
  multi_unit_project: form.multiUnitProject,
  property_type: form.propertyType,
  rental_type: form.rentalType,
  estate_name: form.estateName,
  address_text: form.addressText,
  address_text_en: form.addressTextEn,
  block_name: form.blockName,
  unit_name: form.unitName,
  show_unit: form.showUnit,
  asking_price_hkd: form.askingPriceHKD,
  monthly_rent_hkd: form.monthlyRentHKD,
  price_reference_only: form.priceReferenceOnly,
  price_negotiable: form.priceNegotiable,
  annual_prepay_discount: false,
  annual_prepay_option: 'none',
  lease_start_date: form.leaseStartDate,
  rent_included: form.rentIncluded,
  area_mode: form.areaMode,
  usable_area_sqft: form.usableAreaSqft,
  gross_area_sqft: form.grossAreaSqft,
  bedroom_count: form.bedroomCount,
  living_room_count: form.livingRoomCount,
  bathroom_count: form.bathroomCount,
  floor_level: form.floorLevel,
  floor_raw: form.floorRaw,
  floor_zone: form.floorZone,
  floor_display_range: '',
  total_floors: form.totalFloors,
  public_location_text: publicFloorPreview.value,
  direction: form.direction,
  building_age: form.buildingAge,
  kitchen_type: form.kitchenType,
  cooking_mode: form.cookingMode,
  management_fee_hkd: form.managementFeeHKD,
  video_url: form.videoUrl,
  vr_url: form.vrUrl,
  private_note: form.privateNote,
  title_en: form.titleEn,
  description_en: form.descriptionEn,
  ad_package_code: form.adPackageCode,
  ad_weight: 0,
  ad_price_hkd: selectedAdPackage.value.price_hkd,
  ad_price_points: selectedAdPackage.value.price_points,
  ad_duration_days: selectedAdPackage.value.duration_days,
  ad_expires_at: null,
  feature_tags: form.featureTags,
  contact_method: form.contactMethod,
  publisher_role_label: '',
});
const buildEmptyServicedPayload = (): ServicedApartmentPayload => {
  const projectMinStay = deriveProjectMinStay(form.roomTypes);

  return {
    project_name: form.projectName,
    project_name_en: form.projectNameEn,
    address_text: form.addressText,
    address_text_en: form.servicedAddressTextEn,
    website_url: form.websiteUrl,
    whatsapp: form.servicedWhatsapp,
    fax: form.fax,
    description_en: form.descriptionEn,
    service_intro: form.serviceIntro,
    benefits_text: form.benefitsText,
    extra_charges_text: form.extraChargesText,
    lowest_monthly_rent_hkd: form.lowestMonthlyRentHKD,
    lowest_daily_rent_hkd: form.lowestDailyRentHKD,
    price_reference_only: form.priceReferenceOnly,
    price_negotiable: form.priceNegotiable,
    min_usable_area_sqft: form.minUsableAreaSqft,
    min_lease_months: projectMinStay.minLeaseMonths,
    min_stay_value: projectMinStay.minStayValue,
    min_stay_unit: projectMinStay.minStayUnit,
    location_scope: form.locationScope,
    listing_category: form.listingCategory,
    multi_unit_project: form.multiUnitProject,
    ad_package_code: form.adPackageCode,
    ad_weight: 0,
    ad_price_hkd: selectedAdPackage.value.price_hkd,
    ad_price_points: selectedAdPackage.value.price_points,
    ad_duration_days: selectedAdPackage.value.duration_days,
    ad_expires_at: null,
    facility_tags: form.facilityTags,
    service_tags: form.serviceTags,
    room_types: form.roomTypes,
    contact_method: form.contactMethod,
    publisher_role_label: '',
  };
};
const createPreviewListing = (): PropertyListingDetailResponse => ({
  listing_id: 'preview',
  module: isSale.value ? 'property_sale' : 'serviced_apartment',
  title: form.title,
  summary: form.summary,
  district_code: form.districtCode,
  publisher_identity_type: form.publisherIdentityType,
  publication_status: 'draft',
  business_status: form.businessStatus,
  updated_at: '',
  description: form.description,
  images: [],
  contact_summary: {
    show_phone: form.allowPhone,
    show_whatsapp: form.allowWhatsapp,
    show_chat: form.allowChat,
    show_inquiry_form: false,
  },
  property_sale: isSale.value
    ? {
      ...buildEmptySalePayload(),
      transaction_type: form.transactionType,
      asking_price_hkd: form.askingPriceHKD,
      monthly_rent_hkd: form.monthlyRentHKD,
      price_reference_only: form.priceReferenceOnly,
      price_negotiable: form.priceNegotiable,
    }
    : null,
  serviced_apartment: isSale.value
    ? null
    : {
      ...buildEmptyServicedPayload(),
      lowest_monthly_rent_hkd: form.lowestMonthlyRentHKD,
      lowest_daily_rent_hkd: form.lowestDailyRentHKD,
      price_reference_only: form.priceReferenceOnly,
      price_negotiable: form.priceNegotiable,
    },
});
const previewPrice = computed(() =>
  resolvePropertyPriceText(createPreviewListing(), preferenceStore.locale),
);
const coverImage = computed(() => images.value.find((image) => image.isCover) ?? images.value[0]);
const galleryPreviewImages = computed(() =>
  coverImage.value
    ? [coverImage.value, ...images.value.filter((image) => image.id !== coverImage.value?.id)]
    : images.value,
);
const servicedPreviewVisibleThumbs = computed(() => galleryPreviewImages.value.slice(1, 9));
const servicedPreviewHiddenThumbCount = computed(() => Math.max(galleryPreviewImages.value.length - 9, 0));
const servicedFacilityIconMap: Record<string, AppIconName> = {
  private_kitchen: 'kitchen',
  front_desk_24h: 'clock',
  broadband: 'globe',
  business_center: 'building',
  parking: 'parking',
  child_care: 'smile',
  gym: 'fitness',
  housekeeping: 'cleaning',
  laundry: 'laundry',
  pay_tv: 'view',
  pet_friendly: 'pet',
  restaurant: 'restaurant',
  shuttle_bus: 'shuttle',
  pool: 'pool',
};
const servicedPreviewFacilityItems = computed(() => {
  const activeTags = new Set([...form.facilityTags, ...form.serviceTags]);
  return servicedFacilityTagOptions.map((option) => ({
    label: getPropertyOptionLabel(option, preferenceStore.locale),
    icon: servicedFacilityIconMap[option.value] ?? 'check-circle',
    active: activeTags.has(option.label_zh_hk) || activeTags.has(option.label_en) || activeTags.has(option.value),
  }));
});
const mediaTitle = computed(() => t('property.editor.media'));
const mediaNote = computed(() => t('property.editor.imageUploadNote'));
const selectMediaText = computed(() => t('property.editor.selectImages'));
const chargeCost = computed(() =>
  isSale.value ? selectedAdPackage.value.price_points : selectedAdPackage.value.price_points,
);
const formatPoints = (value: number): string =>
  formatAjoPoints(value, t('common.brand.pointsName'), preferenceStore.locale);
const chargeHint = computed(() =>
  `${t('property.editor.chargeHint')} ${formatPoints(chargeCost.value)} · ${t('property.editor.walletBalance')} ${formatPoints(sessionStore.me?.ajo_balance ?? 0)}`,
);
const selectedAdPackageText = computed(() =>
  isSale.value
    ? `${getPropertyOptionLabel(selectedAdPackage.value, preferenceStore.locale)} · HKD ${selectedAdPackage.value.price_hkd} · ${formatPoints(selectedAdPackage.value.price_points)} · ${selectedAdPackage.value.duration_days} ${t('property.editor.adDaysUnit')}`
    : `${getPropertyOptionLabel(selectedAdPackage.value, preferenceStore.locale)} · ${formatPoints(selectedAdPackage.value.price_points)} · ${selectedAdPackage.value.duration_days} ${t('property.editor.adDaysUnit')}`,
);
const selectedTransactionText = computed(() => {
  const option = propertyTransactionTypeOptions.find((item) => item.value === form.transactionType);
  return option ? getPropertyOptionLabel(option, preferenceStore.locale) : form.transactionType;
});
const selectedPublisherText = computed(() => {
  const option = propertyPublisherOptions.find((item) => item.value === form.publisherIdentityType);
  return option ? getPropertyOptionLabel(option, preferenceStore.locale) : form.publisherIdentityType;
});
const selectedLocationScopeText = computed(() => {
  const option = propertyLocationScopeOptions.find((item) => item.value === form.locationScope);
  return option ? getPropertyOptionLabel(option, preferenceStore.locale) : form.locationScope;
});
const selectedCategoryText = computed(() => {
  const option = propertyListingCategoryOptions.find((item) => item.value === form.listingCategory);
  return option ? getPropertyOptionLabel(option, preferenceStore.locale) : form.listingCategory;
});
const selectedDistrictText = computed(() => {
  const option = marketplaceDistricts.find((item) => item.value === form.districtCode);
  return option ? getPropertyOptionLabel(option, preferenceStore.locale) : form.districtCode;
});

// 0. 格式化面積
const formatAreaSqft = (value?: number): string =>
  value && value > 0 ? t('common.unit.sqft', { value }) : '-';

// 0. 格式化最短逗留
const formatStay = (value?: number, unit?: string): string =>
  value && value > 0 ? `${value} ${unit === 'day' ? t('property.editor.dayUnit') : t('property.editor.monthUnit')}` : '-';

// 0. 格式化租金範圍
const formatRentRange = (min?: number, max?: number): string => {
  if (!min || min <= 0) {
    return '-';
  }
  if (max && max > min) {
    return `${formatPrice(min, preferenceStore.locale)} - ${formatPrice(max, preferenceStore.locale)}`;
  }
  return `${formatPrice(min, preferenceStore.locale)}起`;
};

// 0. 格式化服務式住宅房型租金
const formatServicedRoomRent = (room: ServicedApartmentRoomType): string =>
  (room.min_stay_unit ?? 'month') === 'day'
    ? formatRentRange(room.daily_rent_min_hkd, room.daily_rent_max_hkd)
    : formatRentRange(room.monthly_rent_min_hkd || room.monthly_rent_hkd, room.monthly_rent_max_hkd);

// 0. 從房型推導項目最短逗留
const deriveProjectMinStay = (roomTypes: ServicedApartmentRoomType[]): {
  minLeaseMonths: number;
  minStayValue: number;
  minStayUnit: 'day' | 'month';
} => {
  const normalizedRooms = roomTypes
    .map((room) => ({
      value: Number(room.min_stay_value || room.min_lease_months || 1),
      unit: room.min_stay_unit ?? 'month',
    }))
    .filter((room) => room.value > 0);
  const monthlyRooms = normalizedRooms.filter((room) => room.unit === 'month');
  const dailyRooms = normalizedRooms.filter((room) => room.unit === 'day');
  const source = monthlyRooms.length > 0
    ? monthlyRooms.reduce((min, room) => (room.value < min.value ? room : min), monthlyRooms[0])
    : dailyRooms.reduce((min, room) => (room.value < min.value ? room : min), dailyRooms[0] ?? { value: 1, unit: 'month' as const });

  return {
    minLeaseMonths: source.unit === 'month' ? source.value : 1,
    minStayValue: source.value,
    minStayUnit: source.unit,
  };
};

// 0. 建立 WhatsApp 連結
const buildWhatsAppHref = (value: string): string => {
  if (value.startsWith('https://wa.me/')) {
    return value;
  }
  const digits = value.replace(/[+\s\-()]/g, '');
  return digits ? `https://wa.me/${digits}` : '';
};

// 0. 建立服務式住宅房型
const createServicedRoomType = (): ServicedApartmentRoomType => ({
  name: '',
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

// 1. 讀取錯誤訊息
const readErrorMessage = (error: unknown, fallback: string): string =>
  axios.isAxiosError<{ message?: string }>(error)
    ? error.response?.data?.message ?? fallback
    : fallback;

// 2. 清空欄位校驗
const clearValidationErrors = (): void => {
  Object.keys(validationErrors).forEach((key) => {
    delete validationErrors[key as PropertyValidationField];
  });
};

// 2.1 設定欄位校驗
const setValidationError = (field: PropertyValidationField): void => {
  validationErrors[field] = t('property.editor.requiredFieldInline');
};

// 2.2 判斷服務式住宅房型是否缺少有效價格
const isServicedRoomPriceMissing = (room: ServicedApartmentRoomType): boolean => {
  if (form.priceReferenceOnly || form.priceNegotiable) return false;
  if ((room.min_stay_unit ?? 'month') === 'day') {
    return Number(room.daily_rent_min_hkd || 0) <= 0;
  }
  return Number(room.monthly_rent_min_hkd || room.monthly_rent_hkd || 0) <= 0;
};

// 2.2 校驗分類步驟
const validateCategoryStep = (): boolean => {
  clearValidationErrors();
  if (!isSale.value) {
    if (!form.adPackageCode.trim()) setValidationError('adPackageCode');
    return Object.keys(validationErrors).length === 0;
  }
  if (!form.publisherIdentityType.trim()) setValidationError('publisherIdentityType');
  if (!form.transactionType.trim()) setValidationError('transactionType');
  if (!form.locationScope.trim()) setValidationError('locationScope');
  if (!form.listingCategory.trim()) setValidationError('listingCategory');
  if (!form.adPackageCode.trim()) setValidationError('adPackageCode');
  return Object.keys(validationErrors).length === 0;
};

// 2.3 校驗內容步驟
const validateContentStep = (): boolean => {
  clearValidationErrors();
  if (isSale.value) {
    if (!form.estateName.trim()) setValidationError('estateName');
    if (!form.addressText.trim()) setValidationError('addressText');
    if (!form.addressTextEn.trim()) setValidationError('addressTextEn');
    if (isRentalSaleListing.value && !form.priceNegotiable) {
      if (form.monthlyRentHKD <= 0) setValidationError('monthlyRentHKD');
    } else if (!isRentalSaleListing.value && !form.priceNegotiable && form.askingPriceHKD <= 0) {
      setValidationError('askingPriceHKD');
    }
    if (form.usableAreaSqft <= 0) setValidationError('usableAreaSqft');
    if (!form.title.trim()) setValidationError('title');
    if (!form.titleEn.trim()) setValidationError('titleEn');
    if (!form.description.trim()) setValidationError('description');
    if (!form.descriptionEn.trim()) setValidationError('descriptionEn');
    if (!form.floorRaw.trim() && !form.floorLevel.trim()) setValidationError('floorRaw');
    return Object.keys(validationErrors).length === 0;
  }

  if (!form.projectName.trim()) setValidationError('projectName');
  if (!form.addressText.trim()) setValidationError('addressText');
  if (!form.districtCode.trim()) setValidationError('addressText');
  if (!form.roomTypes.length || form.roomTypes.some((room) =>
    !room.name.trim() ||
    Number(room.min_stay_value || room.min_lease_months || 0) <= 0 ||
    isServicedRoomPriceMissing(room),
  )) {
    setValidationError('roomTypes');
  }
  return Object.keys(validationErrors).length === 0;
};

// 2.4 切換標籤
const toggleTag = (target: string[], value: string): void => {
  const index = target.indexOf(value);
  if (index >= 0) {
    target.splice(index, 1);
    return;
  }

  target.push(value);
};

// 2.5 新增標籤輸入值
const addTextTag = (target: string[], value: string): void => {
  const normalized = value.trim();
  if (!normalized || target.includes(normalized)) {
    return;
  }
  target.push(normalized);
};

// 2.6 移除標籤輸入值
const removeTextTag = (target: string[], value: string): void => {
  const index = target.indexOf(value);
  if (index >= 0) {
    target.splice(index, 1);
  }
};

// 2.7 提交標籤輸入框
const submitTagDraft = (target: string[], draftKey: keyof typeof servicedTagDrafts): void => {
  addTextTag(target, servicedTagDrafts[draftKey]);
  servicedTagDrafts[draftKey] = '';
};

// 2.8 拆分多行標籤文字
const parseTextTags = (value: string): string[] =>
  value
    .split(/\r?\n/)
    .map((item) => item.replace(/^[-•]\s*/, '').trim())
    .filter(Boolean);

// 2.9 合併標籤為多行文字
const serializeTextTags = (tags: string[]): string =>
  tags.map((tag) => tag.trim()).filter(Boolean).join('\n');

// 2.5 新增服務式住宅房型
const addServicedRoomType = (): void => {
  form.roomTypes.push(createServicedRoomType());
};

// 2.6 移除服務式住宅房型
const removeServicedRoomType = (target: ServicedApartmentRoomType): void => {
  if (form.roomTypes.length <= 1) {
    return;
  }
  form.roomTypes = form.roomTypes.filter((room) => room !== target);
};

// 2.1 切換租金包含項目
const toggleRentIncluded = (value: string): void => {
  const items = form.rentIncluded.split(',').map((item) => item.trim()).filter(Boolean);
  const index = items.indexOf(value);
  if (index >= 0) {
    items.splice(index, 1);
  } else {
    items.push(value);
  }
  form.rentIncluded = items.join(', ');
};

// 2.2 判斷租金包含項目
const hasRentIncluded = (value: string): boolean =>
  form.rentIncluded.split(',').map((item) => item.trim()).includes(value);

// 2.3 建立公開樓層區間預覽
const buildFloorRangePreview = (): string => {
  const floor = Number.parseInt(form.floorRaw.replace(/\D/g, ''), 10);
  if (!Number.isFinite(floor) || floor <= 0 || form.totalFloors <= 0) {
    return '';
  }
  const start = Math.max(1, floor - 4);
  const end = Math.min(form.totalFloors, floor);
  return `${start}-${end}|${form.totalFloors}/F`;
};

// 2.4 查詢地址聯想
const loadAddressSuggestions = async (): Promise<void> => {
  const keyword = isSale.value ? form.estateName.trim() : form.projectName.trim();
  if (keyword.length < 1) {
    addressSuggestions.value = [];
    return;
  }

  loadingAddressSuggestions.value = true;
  try {
    const response = await searchPropertyAddresses({
      keyword,
      limit: 12,
    });
    addressSuggestions.value = response.data.data;
  } catch {
    addressSuggestions.value = [];
  } finally {
    loadingAddressSuggestions.value = false;
  }
};

// 2.5 延遲查詢地址聯想
const scheduleAddressSuggestions = (): void => {
  if (addressSearchTimer.value) {
    window.clearTimeout(addressSearchTimer.value);
  }
  addressSearchTimer.value = window.setTimeout(() => {
    void loadAddressSuggestions();
  }, 220);
};

// 2.6 套用地址聯想
const applyAddressSuggestion = (suggestion: PropertyAddressSuggestion): void => {
  const currentYear = new Date().getFullYear();

  if (isSale.value) {
    form.estateName = resolveAddressSuggestionTitle(suggestion);
  } else {
    form.projectName = suggestion.estate_name || resolveAddressSuggestionTitle(suggestion);
    form.projectNameEn = suggestion.estate_name_en;
  }
  form.addressText = suggestion.address_text;
  if (isSale.value) {
    form.addressTextEn = suggestion.address_text_en;
  } else {
    form.servicedAddressTextEn = suggestion.address_text_en;
  }
  form.buildingAge = suggestion.completion_year > 0 && suggestion.completion_year <= currentYear
    ? String(currentYear - suggestion.completion_year)
    : '';
  form.blockName = suggestion.block_names[0] ?? '';
  if (suggestion.district_code) {
    form.districtCode = suggestion.district_code;
  }
  addressSuggestions.value = [];
};

// 2.7 顯示地址聯想名稱
const resolveAddressSuggestionTitle = (suggestion: PropertyAddressSuggestion): string =>
  suggestion.display_name ||
  [
    suggestion.estate_name,
    suggestion.estate_name_en && suggestion.estate_name_en !== suggestion.estate_name ? suggestion.estate_name_en : '',
    suggestion.district_label ? `(${suggestion.district_label})` : '',
  ].filter(Boolean).join(' ');

// 2.5 切換發布步驟
const goToStep = (step: PropertyEditorStep): void => {
  if (!visibleEditorSteps.value.some((item) => item.value === step)) {
    return;
  }
  clearValidationErrors();
  activeStep.value = step;
};

// 2.6 前往下一步
const goNextStep = (): void => {
  if (activeStep.value === 'category' && !validateCategoryStep()) {
    feedbackStore.pushToast(t('property.editor.requiredFields'), 'error');
    return;
  }
  if (activeStep.value === 'content' && !validateContentStep()) {
    feedbackStore.pushToast(t('property.editor.requiredFields'), 'error');
    return;
  }
  if (activeStep.value === 'preview') {
    void saveAndPublish();
    return;
  }
  const nextStep = visibleEditorSteps.value[activeStepIndex.value + 1];
  if (nextStep) {
    activeStep.value = nextStep.value;
  }
};

// 2.7 返回上一步
const goPreviousStep = (): void => {
  const previousStep = visibleEditorSteps.value[activeStepIndex.value - 1];
  if (previousStep) {
    activeStep.value = previousStep.value;
  }
};

// 3. 建立上傳目錄
const buildObjectPrefix = (targetListingId: string): string =>
  targetListingId.trim()
    ? `ajo_living/listings/${targetListingId.trim()}/`
    : 'ajo_living/listings/';

// 4. 釋放本地圖片預覽
const revokeImagePreview = (image: PropertyEditorImage): void => {
  if (image.objectUrl) {
    URL.revokeObjectURL(image.objectUrl);
  }
};

// 5. 選擇圖片
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

// 6. 移除圖片
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

// 7. 選擇封面
const selectCover = (imageId: string): void => {
  images.value = images.value.map((image) => ({
    ...image,
    isCover: image.id === imageId,
  }));
};

// 8. 更新圖片上傳結果
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

// 9. 上傳單張圖片
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

// 10. 上傳待處理圖片
const uploadPendingImages = async (targetListingId: string): Promise<void> => {
  const pendingImages = images.value.filter((image) => image.file);
  for (const image of pendingImages) {
    await uploadImage(image, targetListingId);
  }
};

// 11. 建立圖片 payload
const buildImagePayload = (): PropertyImagePayload[] =>
  images.value
    .filter((image) => image.mediaAssetId)
    .map((image, index) => ({
      media_asset_id: image.mediaAssetId as string,
      sort_order: index + 1,
      is_cover: image.isCover || index === 0,
    }));

// 12. 建立樓盤 payload
const buildSalePayload = (): UpsertPropertySalePayload => {
  const payload: UpsertPropertySalePayload = {
    property_no: form.propertyNo.trim(),
    title: form.title.trim(),
    title_en: form.titleEn.trim(),
    summary: form.summary.trim(),
    description: form.description.trim(),
    description_en: form.descriptionEn.trim(),
    district_code: form.districtCode,
    community_id: form.communityId,
    publisher_identity_type: form.publisherIdentityType || 'owner',
    transaction_type: form.transactionType,
    location_scope: form.locationScope,
    listing_category: form.listingCategory,
    multi_unit_project: form.multiUnitProject,
    property_type: form.propertyType,
    rental_type: isRentalSaleListing.value ? form.rentalType : '',
    estate_name: form.estateName.trim(),
    address_text: form.addressText.trim(),
    address_text_en: form.addressTextEn.trim(),
    block_name: form.blockName.trim(),
    unit_name: form.unitName.trim(),
    show_unit: form.showUnit,
    asking_price_hkd: isRentalSaleListing.value ? 0 : Number(form.askingPriceHKD),
    monthly_rent_hkd: isRentalSaleListing.value ? Number(form.monthlyRentHKD) : 0,
    price_reference_only: form.priceReferenceOnly,
    price_negotiable: form.priceNegotiable,
    annual_prepay_discount: isRentalSaleListing.value && form.annualPrepayOption !== 'none',
    annual_prepay_option: isRentalSaleListing.value ? form.annualPrepayOption : 'none',
    lease_start_date: isRentalSaleListing.value ? form.leaseStartDate.trim() : '',
    rent_included: isRentalSaleListing.value ? form.rentIncluded.trim() : '',
    area_mode: form.areaMode,
    usable_area_sqft: Number(form.usableAreaSqft),
    gross_area_sqft: form.grossAreaSqft > 0 ? Number(form.grossAreaSqft) : undefined,
    bedroom_count: Number(form.bedroomCount),
    living_room_count: Number(form.livingRoomCount),
    bathroom_count: Number(form.bathroomCount),
    floor_level: form.floorLevel.trim(),
    floor_raw: form.floorRaw.trim(),
    floor_zone: form.floorZone,
    total_floors: Number(form.totalFloors),
    direction: form.direction.trim(),
    building_age: form.buildingAge.trim(),
    kitchen_type: form.kitchenType.trim(),
    cooking_mode: form.cookingMode.trim(),
    management_fee_hkd: Number(form.managementFeeHKD),
    video_url: form.videoUrl.trim(),
    vr_url: form.vrUrl.trim(),
    private_note: isRentalSaleListing.value ? form.privateNote.trim() : '',
    ad_package_code: form.adPackageCode,
    feature_tags: [],
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
  payload.feature_tags = mergePropertyFeatureTags(form.featureTags, payload);
  return payload;
};

// 13. 建立服務式住宅 payload
const buildServicedPayload = (): UpsertServicedApartmentPayload => {
  const roomTypes = form.roomTypes.map((room) => {
    const stayUnit = room.min_stay_unit ?? 'month';
    const fallbackRent = form.priceReferenceOnly || form.priceNegotiable ? 1 : 0;
    const monthlyMin = stayUnit === 'day' ? 0 : Number(room.monthly_rent_min_hkd || room.monthly_rent_hkd || fallbackRent);
    const monthlyMax = stayUnit === 'day' ? 0 : Number(room.monthly_rent_max_hkd || monthlyMin || 0);
    const dailyMin = stayUnit === 'day' ? Number(room.daily_rent_min_hkd || fallbackRent) : 0;
    const dailyMax = stayUnit === 'day' ? Number(room.daily_rent_max_hkd || dailyMin || 0) : 0;
    return {
      ...room,
      name: room.name.trim(),
      room_category: room.room_category?.trim() || '',
      usable_area_sqft: Number(room.usable_area_sqft || 0),
      monthly_rent_hkd: monthlyMin,
      monthly_rent_min_hkd: monthlyMin,
      monthly_rent_max_hkd: monthlyMax,
      daily_rent_min_hkd: dailyMin,
      daily_rent_max_hkd: dailyMax,
      included_fees: Boolean(room.included_fees),
      included_fee_items: room.included_fee_items ?? [],
      min_lease_months: stayUnit === 'month'
        ? Number(room.min_stay_value || room.min_lease_months || 1)
        : Number(room.min_lease_months || 1),
      min_stay_value: Number(room.min_stay_value || room.min_lease_months || 1),
      min_stay_unit: stayUnit,
      feature_tags: room.feature_tags ?? [],
    };
  });
  const monthlyPrices = roomTypes.map((room) => room.monthly_rent_min_hkd ?? 0).filter((value) => value > 0);
  const dailyPrices = roomTypes.map((room) => room.daily_rent_min_hkd ?? 0).filter((value) => value > 0);
  const areas = roomTypes.map((room) => room.usable_area_sqft).filter((value) => value > 0);
  const projectMinStay = deriveProjectMinStay(roomTypes);

  return {
    title: form.projectName.trim(),
    title_en: form.projectNameEn.trim(),
    summary: form.summary.trim(),
    description: serializeTextTags(form.residenceInfoTags),
    description_en: form.descriptionEn.trim(),
    district_code: form.districtCode,
    community_id: form.communityId,
    publisher_identity_type: form.publisherIdentityType || 'owner',
    project_name: form.projectName.trim(),
    project_name_en: form.projectNameEn.trim(),
    address_text: form.addressText.trim(),
    address_text_en: form.servicedAddressTextEn.trim(),
    website_url: form.websiteUrl.trim(),
    whatsapp: form.servicedWhatsapp.trim(),
    fax: form.fax.trim(),
    service_intro: serializeTextTags(form.serviceTags),
    benefits_text: serializeTextTags(form.residenceInfoTags),
    extra_charges_text: serializeTextTags(form.extraChargeTags),
    lowest_monthly_rent_hkd: monthlyPrices.length ? Math.min(...monthlyPrices) : 0,
    lowest_daily_rent_hkd: dailyPrices.length ? Math.min(...dailyPrices) : 0,
    price_reference_only: form.priceReferenceOnly,
    price_negotiable: form.priceNegotiable,
    min_usable_area_sqft: areas.length ? Math.min(...areas) : 0,
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
      whatsapp: form.servicedWhatsapp.trim(),
      email: form.email.trim(),
      show_phone: form.allowPhone,
      show_whatsapp: form.allowWhatsapp,
      show_chat: form.allowChat,
      show_inquiry_form: false,
    },
  };
};

// 14. 載入社區清單
const loadCommunities = async (): Promise<void> => {
  try {
    const response = await fetchCommunities();
    communities.value = response.data.data.items;
  } catch {
    communities.value = [];
  }
};

// 15. 回填詳情
const applyDetail = (detail: PropertyListingDetailResponse): void => {
  form.title = detail.title;
  form.summary = detail.summary;
  form.description = detail.description;
  form.districtCode = detail.district_code;
  form.communityId = detail.community?.public_id || '';
  form.publisherIdentityType = detail.publisher_identity_type;
  form.businessStatus = detail.business_status === 'sold' ? 'sold' : 'available';
  form.allowPhone = detail.contact_summary.show_phone;
  form.allowWhatsapp = detail.contact_summary.show_whatsapp;
  form.allowChat = detail.contact_summary.show_chat;

  if (detail.property_sale) {
    form.propertyNo = detail.property_sale.property_no;
    form.titleEn = detail.property_sale.title_en;
    form.descriptionEn = detail.property_sale.description_en;
    form.transactionType = detail.property_sale.transaction_type;
    form.locationScope = detail.property_sale.location_scope;
    form.listingCategory = detail.property_sale.listing_category;
    form.multiUnitProject = detail.property_sale.multi_unit_project;
    form.propertyType = detail.property_sale.property_type;
    form.rentalType = detail.property_sale.rental_type;
    form.estateName = detail.property_sale.estate_name;
    form.addressText = detail.property_sale.address_text;
    form.addressTextEn = detail.property_sale.address_text_en;
    form.blockName = detail.property_sale.block_name;
    form.unitName = detail.property_sale.unit_name ?? '';
    form.showUnit = detail.property_sale.show_unit;
    form.askingPriceHKD = detail.property_sale.asking_price_hkd;
    form.monthlyRentHKD = detail.property_sale.monthly_rent_hkd;
    form.priceReferenceOnly = detail.property_sale.price_reference_only;
    form.priceNegotiable = detail.property_sale.price_negotiable;
    form.annualPrepayOption = detail.property_sale.annual_prepay_option === '90_off' || detail.property_sale.annual_prepay_option === '95_off'
      ? detail.property_sale.annual_prepay_option
      : detail.property_sale.annual_prepay_discount ? '95_off' : 'none';
    form.leaseStartDate = detail.property_sale.lease_start_date;
    form.rentIncluded = detail.property_sale.rent_included;
    form.areaMode = detail.property_sale.area_mode;
    form.usableAreaSqft = detail.property_sale.usable_area_sqft;
    form.grossAreaSqft = detail.property_sale.gross_area_sqft ?? 0;
    form.bedroomCount = detail.property_sale.bedroom_count;
    form.livingRoomCount = detail.property_sale.living_room_count;
    form.bathroomCount = detail.property_sale.bathroom_count;
    form.floorLevel = detail.property_sale.floor_level;
    form.floorRaw = detail.property_sale.floor_raw ?? '';
    form.floorZone = detail.property_sale.floor_zone;
    form.totalFloors = detail.property_sale.total_floors;
    form.direction = detail.property_sale.direction;
    form.buildingAge = detail.property_sale.building_age;
    form.kitchenType = detail.property_sale.kitchen_type || 'NA';
    form.cookingMode = detail.property_sale.cooking_mode || 'NA';
    form.managementFeeHKD = detail.property_sale.management_fee_hkd;
    form.videoUrl = detail.property_sale.video_url;
    form.vrUrl = detail.property_sale.vr_url;
    form.privateNote = detail.property_sale.private_note;
    form.adPackageCode = detail.property_sale.ad_package_code as PropertyEditorForm['adPackageCode'];
    form.featureTags = [...detail.property_sale.feature_tags];
    form.contactMethod = detail.property_sale.contact_method as PropertyEditorForm['contactMethod'];
  }

  if (detail.serviced_apartment) {
    form.projectName = detail.serviced_apartment.project_name;
    form.projectNameEn = detail.serviced_apartment.project_name_en;
    form.addressText = detail.serviced_apartment.address_text;
    form.servicedAddressTextEn = detail.serviced_apartment.address_text_en;
    form.websiteUrl = detail.serviced_apartment.website_url;
    form.servicedWhatsapp = detail.serviced_apartment.whatsapp;
    form.fax = detail.serviced_apartment.fax;
    form.descriptionEn = detail.serviced_apartment.description_en;
    form.serviceIntro = detail.serviced_apartment.service_intro;
    form.benefitsText = detail.serviced_apartment.benefits_text;
    form.extraChargesText = detail.serviced_apartment.extra_charges_text;
    form.residenceInfoTags = parseTextTags(detail.description || detail.serviced_apartment.benefits_text);
    form.extraChargeTags = parseTextTags(detail.serviced_apartment.extra_charges_text);
    form.lowestMonthlyRentHKD = detail.serviced_apartment.lowest_monthly_rent_hkd;
    form.lowestDailyRentHKD = detail.serviced_apartment.lowest_daily_rent_hkd;
    form.priceReferenceOnly = detail.serviced_apartment.price_reference_only;
    form.priceNegotiable = detail.serviced_apartment.price_negotiable;
    form.minUsableAreaSqft = detail.serviced_apartment.min_usable_area_sqft;
    form.minLeaseMonths = detail.serviced_apartment.min_lease_months;
    form.minStayValue = detail.serviced_apartment.min_stay_value || detail.serviced_apartment.min_lease_months || 1;
    form.minStayUnit = detail.serviced_apartment.min_stay_unit || 'month';
    form.locationScope = detail.serviced_apartment.location_scope;
    form.listingCategory = detail.serviced_apartment.listing_category;
    form.multiUnitProject = detail.serviced_apartment.multi_unit_project;
    form.adPackageCode = detail.serviced_apartment.ad_package_code as PropertyEditorForm['adPackageCode'];
    form.facilityTags = [...detail.serviced_apartment.facility_tags];
    form.serviceTags = detail.serviced_apartment.service_tags.length > 0
      ? [...detail.serviced_apartment.service_tags]
      : parseTextTags(detail.serviced_apartment.service_intro);
    form.contactMethod = detail.serviced_apartment.contact_method as PropertyEditorForm['contactMethod'];
    form.roomTypes = detail.serviced_apartment.room_types.length > 0
      ? detail.serviced_apartment.room_types.map((room) => ({
        ...createServicedRoomType(),
        ...room,
        monthly_rent_min_hkd: room.monthly_rent_min_hkd || room.monthly_rent_hkd,
        monthly_rent_max_hkd: room.monthly_rent_max_hkd || room.monthly_rent_min_hkd || room.monthly_rent_hkd,
        daily_rent_max_hkd: room.daily_rent_max_hkd || room.daily_rent_min_hkd || 0,
        min_stay_value: room.min_stay_value || room.min_lease_months || 1,
        min_stay_unit: room.min_stay_unit || 'month',
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

// 16. 載入編輯資料
const loadDetail = async (): Promise<void> => {
  if (!isEditing.value) {
    return;
  }

  loading.value = true;
  try {
    const response = props.staffMode
      ? isSale.value
        ? await fetchStaffPropertySaleDetail(activeListingId.value)
        : await fetchStaffServicedApartmentDetail(activeListingId.value)
      : isSale.value
        ? await fetchPropertySaleDetail(activeListingId.value)
        : await fetchServicedApartmentDetail(activeListingId.value);

    applyDetail(response.data.data);
  } catch (error: unknown) {
    feedbackStore.pushToast(readErrorMessage(error, t('property.detail.loadError')), 'error');
  } finally {
    loading.value = false;
  }
};

// 17. 套用會員聯絡資料預設值
const applyMemberContactDefaults = (): void => {
  const memberPhone = resolveMemberPhone();
  if (!form.contactPersonName.trim()) form.contactPersonName = sessionStore.me?.display_name || '';
  if (!form.phone.trim()) form.phone = memberPhone;
  if (!form.whatsapp.trim()) form.whatsapp = memberPhone;
  if (!form.servicedWhatsapp.trim()) form.servicedWhatsapp = memberPhone;
  if (!form.email.trim()) form.email = sessionStore.me?.email || '';
};

// 17. 儲存草稿
const saveDraft = async (): Promise<string> => {
  if (!validateContentStep()) {
    feedbackStore.pushToast(t('property.editor.requiredFields'), 'error');
    return '';
  }

  saving.value = true;
  try {
    if (!activeListingId.value) {
      if (props.staffMode) {
        feedbackStore.pushToast(t('property.editor.staffCreateDisabled'), 'error');
        return '';
      }
      const response = isSale.value
        ? await createPropertySale({ ...buildSalePayload(), images: [] })
        : await createServicedApartment({ ...buildServicedPayload(), images: [] });

      activeListingId.value = response.data.data.listing_id;
    }

    await uploadPendingImages(activeListingId.value);

    if (isSale.value) {
      if (props.staffMode) {
        await updateStaffPropertySale(activeListingId.value, buildSalePayload());
      } else {
        await updatePropertySale(activeListingId.value, buildSalePayload());
      }
    } else {
      if (props.staffMode) {
        await updateStaffServicedApartment(activeListingId.value, buildServicedPayload());
      } else {
        await updateServicedApartment(activeListingId.value, buildServicedPayload());
      }
    }
    await sessionStore.loadCurrentUser();

    feedbackStore.pushToast(t('property.editor.saveSuccess'), 'success');
    return activeListingId.value;
  } catch (error: unknown) {
    feedbackStore.pushToast(readErrorMessage(error, t('property.editor.saveError')), 'error');
    return '';
  } finally {
    saving.value = false;
  }
};

// 18. 儲存並發布
const saveAndPublish = async (): Promise<void> => {
  if (props.staffMode) {
    const savedListingId = await saveDraft();
    if (savedListingId) {
      completedListingId.value = savedListingId;
      activeStep.value = 'complete';
      if (props.embedded) {
        emit('saved', savedListingId);
      }
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
    if (!hasPersistedImage.value) {
      feedbackStore.pushToast(t('property.editor.imageUploadRequired'), 'error');
      return;
    }

    if (isSale.value) {
      await publishPropertySale(savedListingId);
    } else {
      await publishServicedApartment(savedListingId);
    }
    await sessionStore.loadCurrentUser();

    feedbackStore.pushToast(t('property.editor.publishSuccess'), 'success');
    completedListingId.value = savedListingId;
    activeStep.value = 'complete';
    if (props.embedded) {
      return;
    }

  } catch (error: unknown) {
    feedbackStore.pushToast(readErrorMessage(error, t('property.editor.publishError')), 'error');
  } finally {
    publishing.value = false;
  }
};

// 19. 儲存並返回列表
const saveAndReturn = async (): Promise<void> => {
  const savedListingId = await saveDraft();
  if (savedListingId) {
    if (props.embedded) {
      emit('saved', savedListingId);
      return;
    }

    await router.push(myPath.value);
  }
};

// 20. 完成後返回列表
const finishEditor = async (): Promise<void> => {
  const listingId = completedListingId.value || activeListingId.value;
  if (props.embedded) {
    emit('published', listingId);
    return;
  }
  await router.push(myPath.value);
};

onMounted(async () => {
  applyMemberContactDefaults();
  await loadCommunities();
  await loadDetail();
  applyMemberContactDefaults();
});

// 21. 取得標籤分組文案
const resolveTagGroupLabel = (group: { label_zh_hk: string; label_en: string }): string =>
  preferenceStore.locale === 'zh-HK' ? group.label_zh_hk : group.label_en;

onBeforeUnmount(() => {
  if (addressSearchTimer.value) {
    window.clearTimeout(addressSearchTimer.value);
  }
  images.value.forEach(revokeImagePreview);
});
</script>

<template>
  <main
    class="property-editor-page"
    :class="{
      'property-editor-page--embedded': props.embedded,
      'property-editor-page--serviced': !isSale,
    }"
  >
    <section class="property-editor-heading">
      <div>
        <p
          v-if="isSale"
          class="property-kicker"
        >
          {{ isEditing ? t('property.editor.editMode') : t('property.editor.publishMode') }}
        </p>
        <h1>{{ pageTitle }}</h1>
        <p v-if="isSale">{{ previewPrice }}</p>
      </div>
      <button
        v-if="props.embedded && isSale"
        type="button"
        class="property-editor-close"
        :aria-label="t('common.action.close')"
        @click="emit('cancel')"
      >
        <AppIcon
          name="close"
          :size="18"
        />
      </button>
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
      <nav class="property-editor-steps" aria-label="property publish steps">
        <button
          v-for="(step, index) in visibleEditorSteps"
          :key="step.value"
          type="button"
          :class="{
            'property-editor-step--active': activeStep === step.value,
            'property-editor-step--done': index < activeStepIndex,
          }"
          @click="goToStep(step.value)"
        >
          <span>{{ index + 1 }}</span>
          {{ t(step.labelKey) }}
        </button>
      </nav>

      <form
        class="property-editor-form"
        @submit.prevent="goNextStep"
      >
        <template v-if="activeStep === 'category'">
          <section class="property-editor-panel">
            <h2>{{ isSale ? t('property.editor.stepCategory') : t('property.editor.servicedStepCategory') }}</h2>
            <div class="property-editor-grid">
              <label
                v-if="isSale"
                class="property-input"
              >
                <span>{{ t('property.editor.publisherIdentityField') }}<b>*</b></span>
                <select v-model="form.publisherIdentityType">
                  <option
                    v-for="publisher in propertyPublisherOptions"
                    :key="publisher.value"
                    :value="publisher.value"
                  >
                    {{ getPropertyOptionLabel(publisher, preferenceStore.locale) }}
                  </option>
                </select>
                <small v-if="validationErrors.publisherIdentityType">{{ validationErrors.publisherIdentityType }}</small>
              </label>
              <label
                v-if="isSale"
                class="property-input"
              >
                <span>{{ t('property.editor.transactionTypeField') }}<b>*</b></span>
                <select v-model="form.transactionType">
                  <option
                    v-for="typeOption in propertyTransactionTypeOptions"
                    :key="typeOption.value"
                    :value="typeOption.value"
                  >
                    {{ getPropertyOptionLabel(typeOption, preferenceStore.locale) }}
                  </option>
                </select>
                <small v-if="validationErrors.transactionType">{{ validationErrors.transactionType }}</small>
              </label>
              <label
                v-if="isSale"
                class="property-input"
              >
                <span>{{ t('property.editor.locationScopeField') }}<b>*</b></span>
                <select v-model="form.locationScope">
                  <option
                    v-for="scope in propertyLocationScopeOptions"
                    :key="scope.value"
                    :value="scope.value"
                  >
                    {{ getPropertyOptionLabel(scope, preferenceStore.locale) }}
                  </option>
                </select>
                <small v-if="validationErrors.locationScope">{{ validationErrors.locationScope }}</small>
              </label>
              <label
                v-if="isSale"
                class="property-input"
              >
                <span>{{ t('property.editor.listingCategoryField') }}<b>*</b></span>
                <select v-model="form.listingCategory">
                  <option
                    v-for="category in propertyListingCategoryOptions"
                    :key="category.value"
                    :value="category.value"
                  >
                    {{ getPropertyOptionLabel(category, preferenceStore.locale) }}
                  </option>
                </select>
                <small v-if="validationErrors.listingCategory">{{ validationErrors.listingCategory }}</small>
              </label>
            </div>
            <label
              v-if="isSale"
              class="property-inline-checkbox property-editor-warning"
            >
              <input
                v-model="form.multiUnitProject"
                type="checkbox"
              />
              {{ t('property.editor.multiUnitProjectField') }}
            </label>
            <p
              v-if="isSale"
              class="property-editor-note"
            >
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
                :class="form.adPackageCode === adPackage.value ? 'property-ad-package--active' : ''"
                @click="form.adPackageCode = adPackage.value"
              >
                <strong>{{ getPropertyOptionLabel(adPackage, preferenceStore.locale) }}</strong>
                <span v-if="isSale">HKD:{{ adPackage.price_hkd }} {{ t('property.editor.orPoints') }} {{ formatPoints(adPackage.price_points) }}</span>
                <span v-else>{{ formatPoints(adPackage.price_points) }}</span>
                <small>{{ t('property.editor.adDaysUnit') }}: {{ adPackage.duration_days }} · {{ t('property.editor.adWeightField') }}: {{ adPackage.weight }}</small>
              </button>
            </div>
          </section>
        </template>

        <template v-if="activeStep === 'content'">
          <template v-if="isSale">
            <section class="property-editor-panel">
              <h2>{{ t('property.editor.basicInfoSection') }}</h2>
              <div class="property-editor-compact-grid">
                <label class="property-input property-input--wide">
                  <span>{{ t('property.editor.estateNameField') }}<b>*</b></span>
                  <span class="property-address-input">
                    <input v-model="form.estateName" autocomplete="off" @focus="scheduleAddressSuggestions" @input="scheduleAddressSuggestions" />
                    <i v-if="loadingAddressSuggestions" aria-hidden="true" />
                    <span v-if="addressSuggestions.length > 0 || loadingAddressSuggestions" class="property-address-suggestions">
                      <span v-if="loadingAddressSuggestions" class="property-address-loading">{{ t('common.status.loading') }}</span>
                      <button v-for="suggestion in addressSuggestions" :key="suggestion.address_id" type="button" @click="applyAddressSuggestion(suggestion)">
                        <strong>{{ resolveAddressSuggestionTitle(suggestion) }}</strong>
                        <span>{{ suggestion.address_text }}</span>
                      </button>
                    </span>
                  </span>
                  <small v-if="validationErrors.estateName">{{ validationErrors.estateName }}</small>
                </label>
                <label class="property-input">
                  <span>{{ t('property.editor.grossAreaField') }}</span>
                  <input v-model.number="form.grossAreaSqft" type="number" min="0" :placeholder="t('property.editor.numberOnly')" />
                </label>
                <label class="property-input">
                  <span>{{ t('property.editor.usableAreaField') }}<b>*</b></span>
                  <input v-model.number="form.usableAreaSqft" type="number" min="0" :placeholder="t('property.editor.numberOnly')" />
                  <small v-if="validationErrors.usableAreaSqft">{{ validationErrors.usableAreaSqft }}</small>
                </label>
                <label class="property-input">
                  <span>{{ t('property.editor.buildingAgeField') }}</span>
                  <input v-model="form.buildingAge" inputmode="numeric" :placeholder="t('property.editor.numberOnly')" />
                </label>
                <label class="property-input">
                  <span>{{ t('property.editor.blockNameField') }}</span>
                  <input v-model="form.blockName" />
                </label>
                <label class="property-input">
                  <span>{{ t('property.editor.unitNameField') }}</span>
                  <input v-model="form.unitName" />
                </label>
                <label class="property-inline-checkbox property-input--wide">
                  <input v-model="form.showUnit" type="checkbox" />
                  {{ t('property.editor.showUnitField') }}
                </label>
                <label class="property-input">
                  <span>{{ t('property.editor.addressField') }}<b>*</b></span>
                  <input v-model="form.addressText" />
                  <small v-if="validationErrors.addressText">{{ validationErrors.addressText }}</small>
                </label>
                <label class="property-input">
                  <span>{{ t('property.editor.addressEnField') }}<b>*</b></span>
                  <input v-model="form.addressTextEn" />
                  <small v-if="validationErrors.addressTextEn">{{ validationErrors.addressTextEn }}</small>
                </label>
                <label class="property-input property-input--wide">
                  <span>{{ t('property.editor.mapLocationField') }}</span>
                  <input :value="form.addressText" disabled />
                </label>
              </div>
            </section>

            <section class="property-editor-panel">
              <h2>{{ isRentalSaleListing ? t('property.editor.rentInfoSection') : t('property.editor.saleInfoSection') }}</h2>
              <div class="property-editor-compact-grid">
                <label v-if="isRentalSaleListing" class="property-input">
                  <span>{{ t('property.editor.monthlyRentField') }}<b v-if="!form.priceNegotiable">*</b></span>
                  <span class="property-money-input">
                    <input v-model.number="form.monthlyRentHKD" type="number" min="0" :placeholder="t('property.editor.numberOnly')" />
                    <small>{{ t('property.editor.monthlyRentUnit') }}</small>
                  </span>
                  <small v-if="validationErrors.monthlyRentHKD">{{ validationErrors.monthlyRentHKD }}</small>
                </label>
                <label v-else class="property-input">
                  <span>{{ t('property.editor.askingPriceField') }}<b v-if="!form.priceNegotiable">*</b></span>
                  <span class="property-money-input">
                    <input v-model.number="form.askingPriceHKD" type="number" min="0" :placeholder="t('property.editor.numberOnly')" />
                    <small>{{ t('property.editor.salePriceUnit') }}</small>
                  </span>
                  <small v-if="validationErrors.askingPriceHKD">{{ validationErrors.askingPriceHKD }}</small>
                </label>
                <div class="property-checkbox-row property-input--wide property-price-options">
                  <label>
                    <input v-model="form.priceReferenceOnly" type="checkbox" />
                    {{ t('property.editor.priceReferenceOnlyOption') }}
                  </label>
                  <label>
                    <input v-model="form.priceNegotiable" type="checkbox" />
                    {{ t('property.editor.priceNegotiableOption') }}
                  </label>
                </div>
                <template v-if="isRentalSaleListing">
                  <label class="property-input">
                    <span>{{ t('property.editor.annualPrepayDiscountField') }}</span>
                    <select v-model="form.annualPrepayOption">
                      <option
                        v-for="option in annualPrepayOptions"
                        :key="option.value"
                        :value="option.value"
                      >
                        {{ getPropertyOptionLabel(option, preferenceStore.locale) }}
                      </option>
                    </select>
                  </label>
                  <label class="property-input">
                    <span>{{ t('property.editor.leaseStartDateField') }}</span>
                    <input v-model="form.leaseStartDate" type="date" />
                  </label>
                  <button type="button" class="property-editor-mini-action" @click="form.leaseStartDate = ''">
                    {{ t('property.editor.clearLeaseStartDate') }}
                  </button>
                  <label class="property-input property-input--wide">
                    <span>{{ t('property.editor.rentIncludedField') }}</span>
                    <input v-model="form.rentIncluded" />
                  </label>
                  <div class="property-checkbox-row property-input--wide">
                    <label v-for="item in rentIncludedOptions" :key="item">
                      <input type="checkbox" :checked="hasRentIncluded(item)" @change="toggleRentIncluded(item)" />
                      {{ item }}
                    </label>
                  </div>
                </template>
              </div>
            </section>

            <section class="property-editor-panel">
              <h2>{{ t('property.editor.tagsField') }}</h2>
              <div class="property-tag-groups">
                <section v-for="group in propertyTagGroups" :key="group.key" class="property-tag-group">
                  <h3>{{ resolveTagGroupLabel(group) }}</h3>
                  <div class="property-checkbox-row">
                    <label v-for="tag in group.options" :key="tag.value">
                      <input type="checkbox" :checked="form.featureTags.includes(tag.value)" @change="toggleTag(form.featureTags, tag.value)" />
                      {{ getPropertyOptionLabel(tag, preferenceStore.locale) }}
                    </label>
                  </div>
                </section>
              </div>
            </section>

            <section class="property-editor-panel">
              <h2>{{ t('property.editor.descriptionInfoSection') }}</h2>
              <div class="property-editor-compact-grid">
                <label class="property-input">
                  <span>{{ t('property.editor.titleField') }}<b>*</b></span>
                  <input v-model="form.title" maxlength="40" />
                  <small v-if="validationErrors.title">{{ validationErrors.title }}</small>
                </label>
                <label class="property-input">
                  <span>{{ t('property.editor.titleEnField') }}<b>*</b></span>
                  <input v-model="form.titleEn" maxlength="100" />
                  <small v-if="validationErrors.titleEn">{{ validationErrors.titleEn }}</small>
                </label>
                <label class="property-input property-input--wide">
                  <span>{{ t('property.editor.descriptionField') }}<b>*</b></span>
                  <textarea v-model="form.description" rows="5" maxlength="1000" />
                  <small v-if="validationErrors.description">{{ validationErrors.description }}</small>
                </label>
                <label class="property-input property-input--wide">
                  <span>{{ t('property.editor.descriptionEnField') }}<b>*</b></span>
                  <textarea v-model="form.descriptionEn" rows="5" maxlength="2000" />
                  <small v-if="validationErrors.descriptionEn">{{ validationErrors.descriptionEn }}</small>
                </label>
                <label class="property-input">
                  <span>{{ t('property.editor.directionField') }}</span>
                  <input v-model="form.direction" />
                </label>
                <label class="property-input">
                  <span>{{ t('property.editor.floorField') }}<b>*</b></span>
                  <input v-model="form.floorRaw" />
                  <small v-if="validationErrors.floorRaw">{{ validationErrors.floorRaw }}</small>
                </label>
                <label class="property-input">
                  <span>{{ t('property.editor.floorZoneField') }}</span>
                  <select v-model="form.floorZone">
                    <option v-for="zone in propertyFloorZoneOptions" :key="zone.value" :value="zone.value">
                      {{ getPropertyOptionLabel(zone, preferenceStore.locale) }}
                    </option>
                  </select>
                </label>
                <label class="property-input">
                  <span>{{ t('property.editor.totalFloorsField') }}</span>
                  <input v-model.number="form.totalFloors" type="number" min="0" />
                </label>
                <p class="property-editor-note property-input--wide">{{ t('property.editor.floorPrivacyNote') }} {{ publicFloorPreview }}</p>
                <label class="property-input">
                  <span>{{ t('property.editor.bedroomField') }}</span>
                  <input v-model.number="form.bedroomCount" type="number" min="0" />
                </label>
                <label class="property-input">
                  <span>{{ t('property.editor.bathroomField') }}</span>
                  <input v-model.number="form.bathroomCount" type="number" min="0" />
                </label>
                <label class="property-input">
                  <span>{{ t('property.editor.kitchenTypeField') }}</span>
                  <input v-model="form.kitchenType" />
                </label>
                <label class="property-input">
                  <span>{{ t('property.editor.cookingModeField') }}</span>
                  <input v-model="form.cookingMode" />
                </label>
                <label class="property-input">
                  <span>{{ t('property.editor.managementFeeField') }}</span>
                  <input v-model.number="form.managementFeeHKD" type="number" min="0" />
                </label>
                <label class="property-input">
                  <span>{{ t('property.editor.videoUrlField') }}</span>
                  <input v-model="form.videoUrl" />
                </label>
                <label class="property-input">
                  <span>{{ t('property.editor.vrUrlField') }}</span>
                  <input v-model="form.vrUrl" />
                </label>
                <label v-if="isRentalSaleListing" class="property-input property-input--wide">
                  <span>{{ t('property.editor.privateNoteField') }}</span>
                  <textarea v-model="form.privateNote" rows="4" />
                </label>
              </div>
            </section>
          </template>

          <template v-else>
            <section class="property-editor-panel">
              <h2>{{ t('property.editor.servicedTitle') }}</h2>
              <div class="property-editor-compact-grid property-serviced-basic-grid">
                <label class="property-input">
                  <span>{{ t('property.editor.residenceNameField') }}<b>*</b></span>
                  <span class="property-address-input">
                    <input
                      v-model="form.projectName"
                      autocomplete="off"
                      @focus="scheduleAddressSuggestions"
                      @input="scheduleAddressSuggestions"
                    />
                    <i v-if="loadingAddressSuggestions" aria-hidden="true" />
                    <span v-if="addressSuggestions.length > 0 || loadingAddressSuggestions" class="property-address-suggestions">
                      <span v-if="loadingAddressSuggestions" class="property-address-loading">{{ t('common.status.loading') }}</span>
                      <button v-for="suggestion in addressSuggestions" :key="suggestion.address_id" type="button" @click="applyAddressSuggestion(suggestion)">
                        <strong>{{ resolveAddressSuggestionTitle(suggestion) }}</strong>
                        <span>{{ suggestion.address_text }}</span>
                      </button>
                    </span>
                  </span>
                  <small v-if="validationErrors.projectName">{{ validationErrors.projectName }}</small>
                </label>
                <label class="property-input">
                  <span>{{ t('property.editor.residenceNameEnField') }}</span>
                  <input v-model="form.projectNameEn" />
                </label>
                <label class="property-input">
                  <span>{{ t('property.editor.addressField') }}<b>*</b></span>
                  <input v-model="form.addressText" />
                  <small v-if="validationErrors.addressText">{{ validationErrors.addressText }}</small>
                </label>
                <label class="property-input">
                  <span>{{ t('property.editor.addressEnField') }}</span>
                  <input v-model="form.servicedAddressTextEn" />
                </label>
                <label class="property-input">
                  <span>{{ t('property.editor.districtField') }}<b>*</b></span>
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
                  <span>{{ t('property.editor.websiteField') }}</span>
                  <input v-model="form.websiteUrl" type="url" />
                </label>
                <label class="property-input">
                  <span>{{ t('property.editor.whatsappField') }}</span>
                  <input
                    v-model="form.servicedWhatsapp"
                    inputmode="tel"
                    :placeholder="t('property.editor.whatsappPhonePlaceholder')"
                  />
                </label>
                <label class="property-input">
                  <span>{{ t('property.editor.faxField') }}</span>
                  <input v-model="form.fax" />
                </label>
              </div>
            </section>

            <section class="property-editor-panel">
              <div class="property-panel-title-row">
                <h2>{{ t('property.editor.roomTypesSection') }}</h2>
                <button type="button" class="property-editor-mini-action" @click="addServicedRoomType">
                  {{ t('property.editor.addRoomType') }}
                </button>
              </div>
              <small v-if="validationErrors.roomTypes" class="property-editor-error">{{ validationErrors.roomTypes }}</small>
              <div class="property-checkbox-row property-price-options">
                <label>
                  <input v-model="form.priceReferenceOnly" type="checkbox" />
                  {{ t('property.editor.priceReferenceOnlyOption') }}
                </label>
                <label>
                  <input v-model="form.priceNegotiable" type="checkbox" />
                  {{ t('property.editor.priceNegotiableOption') }}
                </label>
              </div>
              <div class="property-room-editor-stack">
                <article
                  v-for="(room, roomIndex) in form.roomTypes"
                  :key="roomIndex"
                  class="property-room-editor"
                >
                  <div class="property-editor-compact-grid property-serviced-room-grid">
                    <label class="property-input">
                      <span>{{ t('property.editor.roomTypeName') }}<b>*</b></span>
                      <input v-model="room.name" />
                    </label>
                    <label class="property-input">
                      <span>{{ t('property.editor.roomCategoryField') }}</span>
                      <input v-model="room.room_category" />
                    </label>
                    <label class="property-input">
                      <span>{{ t('property.editor.roomTypeArea') }}</span>
                      <input v-model.number="room.usable_area_sqft" type="number" min="0" />
                    </label>
                    <label v-if="room.min_stay_unit !== 'day'" class="property-input">
                      <span>{{ t('property.editor.monthlyRentMinField') }}<b v-if="!form.priceNegotiable">*</b></span>
                      <input v-model.number="room.monthly_rent_min_hkd" type="number" min="0" />
                    </label>
                    <label v-if="room.min_stay_unit !== 'day'" class="property-input">
                      <span>{{ t('property.editor.monthlyRentMaxField') }}</span>
                      <input v-model.number="room.monthly_rent_max_hkd" type="number" min="0" />
                    </label>
                    <label v-if="room.min_stay_unit === 'day'" class="property-input">
                      <span>{{ t('property.editor.dailyRentMinField') }}<b v-if="!form.priceNegotiable">*</b></span>
                      <input v-model.number="room.daily_rent_min_hkd" type="number" min="0" />
                    </label>
                    <label v-if="room.min_stay_unit === 'day'" class="property-input">
                      <span>{{ t('property.editor.dailyRentMaxField') }}</span>
                      <input v-model.number="room.daily_rent_max_hkd" type="number" min="0" />
                    </label>
                    <label class="property-input">
                      <span>{{ t('property.editor.minStayField') }}<b>*</b></span>
                      <span class="property-money-input">
                        <input v-model.number="room.min_stay_value" type="number" min="1" />
                        <select v-model="room.min_stay_unit">
                          <option value="month">{{ t('property.editor.monthUnit') }}</option>
                          <option value="day">{{ t('property.editor.dayUnit') }}</option>
                        </select>
                      </span>
                    </label>
                  </div>
                  <button
                    type="button"
                    class="property-editor-mini-action"
                    :disabled="form.roomTypes.length <= 1"
                    @click="removeServicedRoomType(room)"
                  >
                    {{ t('property.editor.removeRoomType') }}
                  </button>
                </article>
              </div>
            </section>

            <section class="property-editor-panel">
              <h2>{{ t('property.editor.residenceInfoSection') }}</h2>
              <section class="property-tag-input-block">
                <h3>{{ t('property.editor.residenceInfoTagsField') }}</h3>
                <div class="property-tag-input-row">
                  <input
                    v-model="servicedTagDrafts.residenceInfo"
                    :placeholder="t('property.editor.manualTagPlaceholder')"
                    @keyup.enter.prevent="submitTagDraft(form.residenceInfoTags, 'residenceInfo')"
                  />
                  <button type="button" class="property-editor-mini-action" @click="submitTagDraft(form.residenceInfoTags, 'residenceInfo')">
                    {{ t('property.editor.confirmTagAction') }}
                  </button>
                </div>
                <div class="property-tag-chip-row">
                  <button
                    v-for="tag in form.residenceInfoTags"
                    :key="tag"
                    type="button"
                    class="property-tag-chip"
                  >
                    <span>{{ tag }}</span>
                    <i @click.stop="removeTextTag(form.residenceInfoTags, tag)">×</i>
                  </button>
                </div>
                <div class="property-quick-tag-row">
                  <button
                    v-for="tag in servicedResidenceInfoQuickTags"
                    :key="tag"
                    type="button"
                    @click="addTextTag(form.residenceInfoTags, tag)"
                  >
                    {{ tag }}
                  </button>
                </div>
              </section>

              <section class="property-tag-input-block">
                <h3>{{ t('property.editor.facilityTagsField') }}</h3>
                <div class="property-tag-input-row">
                  <input
                    v-model="servicedTagDrafts.facility"
                    :placeholder="t('property.editor.manualTagPlaceholder')"
                    @keyup.enter.prevent="submitTagDraft(form.facilityTags, 'facility')"
                  />
                  <button type="button" class="property-editor-mini-action" @click="submitTagDraft(form.facilityTags, 'facility')">
                    {{ t('property.editor.confirmTagAction') }}
                  </button>
                </div>
                <div class="property-tag-chip-row">
                  <button
                    v-for="tag in form.facilityTags"
                    :key="tag"
                    type="button"
                    class="property-tag-chip"
                  >
                    <span>{{ tag }}</span>
                    <i @click.stop="removeTextTag(form.facilityTags, tag)">×</i>
                  </button>
                </div>
                <div class="property-quick-tag-row">
                  <button
                    v-for="tag in servicedFacilityTagOptions"
                    :key="tag.value"
                    type="button"
                    @click="addTextTag(form.facilityTags, getPropertyOptionLabel(tag, preferenceStore.locale))"
                  >
                    {{ getPropertyOptionLabel(tag, preferenceStore.locale) }}
                  </button>
                </div>
              </section>

              <section class="property-tag-input-block">
                <h3>{{ t('property.editor.serviceTagsField') }}</h3>
                <div class="property-tag-input-row">
                  <input
                    v-model="servicedTagDrafts.service"
                    :placeholder="t('property.editor.manualTagPlaceholder')"
                    @keyup.enter.prevent="submitTagDraft(form.serviceTags, 'service')"
                  />
                  <button type="button" class="property-editor-mini-action" @click="submitTagDraft(form.serviceTags, 'service')">
                    {{ t('property.editor.confirmTagAction') }}
                  </button>
                </div>
                <div class="property-tag-chip-row">
                  <button
                    v-for="tag in form.serviceTags"
                    :key="tag"
                    type="button"
                    class="property-tag-chip"
                  >
                    <span>{{ tag }}</span>
                    <i @click.stop="removeTextTag(form.serviceTags, tag)">×</i>
                  </button>
                </div>
                <div class="property-quick-tag-row">
                  <button
                    v-for="tag in servicedServiceTagOptions"
                    :key="tag.value"
                    type="button"
                    @click="addTextTag(form.serviceTags, getPropertyOptionLabel(tag, preferenceStore.locale))"
                  >
                    {{ getPropertyOptionLabel(tag, preferenceStore.locale) }}
                  </button>
                </div>
              </section>

              <section class="property-tag-input-block">
                <h3>{{ t('property.editor.extraChargesField') }}</h3>
                <div class="property-tag-input-row">
                  <input
                    v-model="servicedTagDrafts.extraCharge"
                    :placeholder="t('property.editor.manualTagPlaceholder')"
                    @keyup.enter.prevent="submitTagDraft(form.extraChargeTags, 'extraCharge')"
                  />
                  <button type="button" class="property-editor-mini-action" @click="submitTagDraft(form.extraChargeTags, 'extraCharge')">
                    {{ t('property.editor.confirmTagAction') }}
                  </button>
                </div>
                <div class="property-tag-chip-row">
                  <button
                    v-for="tag in form.extraChargeTags"
                    :key="tag"
                    type="button"
                    class="property-tag-chip"
                  >
                    <span>{{ tag }}</span>
                    <i @click.stop="removeTextTag(form.extraChargeTags, tag)">×</i>
                  </button>
                </div>
                <div class="property-quick-tag-row">
                  <button
                    v-for="tag in servicedExtraChargeQuickTags"
                    :key="tag"
                    type="button"
                    @click="addTextTag(form.extraChargeTags, tag)"
                  >
                    {{ tag }}
                  </button>
                </div>
              </section>
            </section>
          </template>

          <section class="property-editor-panel">
            <h2>{{ t('property.editor.contact') }}</h2>
            <div class="property-editor-compact-grid property-contact-grid">
              <label class="property-input"><span>{{ t('property.editor.contactPersonZhField') }}</span><input v-model="form.contactPersonName" /></label>
              <label class="property-input"><span>{{ t('property.editor.contactPersonEnField') }}</span><input v-model="form.contactPersonNameEn" /></label>
              <label class="property-input"><span>{{ t('property.editor.phoneField') }}</span><input v-model="form.phone" /></label>
              <label class="property-input"><span>{{ t('property.editor.phone2Field') }}</span><input /></label>
              <label class="property-input"><span>{{ t('property.editor.wechatField') }}</span><input /></label>
              <label class="property-input"><span>{{ t('property.editor.contactMethodField') }}</span><select v-model="form.contactMethod"><option v-for="method in propertyContactMethodOptions" :key="method.value" :value="method.value">{{ getPropertyOptionLabel(method, preferenceStore.locale) }}</option></select></label>
              <label class="property-input"><span>{{ t('property.editor.emailField') }}</span><input v-model="form.email" /></label>
              <p class="property-editor-note property-input--wide">{{ t('property.editor.publicPhoneNote') }}</p>
            </div>
            <div class="property-checkbox-row">
              <label><input v-model="form.allowPhone" type="checkbox" />{{ t('property.editor.allowPhone') }}</label>
              <label><input v-model="form.allowWhatsapp" type="checkbox" />{{ t('property.editor.allowWhatsapp') }}</label>
              <label><input v-model="form.allowChat" type="checkbox" />{{ t('property.editor.allowChat') }}</label>
            </div>
          </section>

          <section class="property-editor-panel">
            <h2>{{ mediaTitle }}</h2>
            <p class="property-editor-note">{{ mediaNote }}</p>
            <label class="property-upload-button">
              <AppIcon name="cloud-upload" :size="18" />
              {{ selectMediaText }}
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
                <img v-if="image.url" :src="image.url" :alt="form.title" />
                <button type="button" @click="selectCover(image.id)">
                  {{ image.isCover ? t('property.editor.cover') : t('property.editor.setCover') }}
                </button>
                <button type="button" @click="removeImage(image.id)">
                  {{ t('property.editor.removeImage') }}
                </button>
              </article>
            </div>
          </section>
        </template>

        <template v-if="activeStep === 'preview'">
          <section v-if="isSale" class="property-editor-panel property-preview-panel">
            <h2>{{ t('property.editor.stepPreview') }}</h2>
            <dl class="property-preview-list">
              <div v-if="isSale"><dt>{{ t('property.editor.publisherIdentityField') }}</dt><dd>{{ selectedPublisherText }}</dd></div>
              <div v-if="isSale"><dt>{{ t('property.editor.transactionTypeField') }}</dt><dd>{{ selectedTransactionText }}</dd></div>
              <div v-if="isSale"><dt>{{ t('property.editor.locationScopeField') }}</dt><dd>{{ selectedLocationScopeText }}</dd></div>
              <div v-if="isSale"><dt>{{ t('property.editor.listingCategoryField') }}</dt><dd>{{ selectedCategoryText }}</dd></div>
              <div><dt>{{ t('property.editor.adPackageField') }}</dt><dd>{{ selectedAdPackageText }}</dd></div>
              <div v-if="isSale"><dt>{{ t('property.editor.estateNameField') }}</dt><dd>{{ form.estateName }}</dd></div>
              <div v-else><dt>{{ t('property.editor.projectNameField') }}</dt><dd>{{ form.projectName }}</dd></div>
              <div><dt>{{ t('property.editor.addressField') }}</dt><dd>{{ form.addressText }}</dd></div>
              <div v-if="isSale"><dt>{{ t('property.editor.floorZoneField') }}</dt><dd>{{ publicFloorPreview }}</dd></div>
              <div><dt>{{ t('property.editor.usableAreaField') }}</dt><dd>{{ isSale ? form.usableAreaSqft : form.minUsableAreaSqft }}</dd></div>
              <div><dt>{{ t('property.sale.priceLabel') }}</dt><dd>{{ previewPrice }}</dd></div>
            </dl>
          </section>

          <section v-else class="property-editor-panel property-serviced-preview property-serviced-layout">
            <div class="property-serviced-content">
              <section class="property-serviced-hero">
                <div class="property-serviced-hero-copy">
                  <p class="property-serviced-kicker">{{ t('property.serviced.detailTitle') }}</p>
                  <h1>{{ form.projectName }}</h1>
                  <p v-if="form.projectNameEn" class="property-serviced-subtitle">{{ form.projectNameEn }}</p>
                  <div class="property-serviced-address">
                    <AppIcon name="location" :size="18" />
                    <span>{{ [selectedDistrictText, form.addressText].filter(Boolean).join(' · ') }}</span>
                  </div>
                </div>

                <div class="property-serviced-gallery-section">
                  <div class="property-serviced-gallery-main">
                    <img
                      v-if="coverImage?.url"
                      :src="coverImage.url"
                      :alt="form.projectName"
                    />
                    <div v-else class="property-gallery__placeholder">
                      <AppIcon name="picture" :size="46" />
                    </div>
                    <span v-if="galleryPreviewImages.length > 0">{{ `1/${galleryPreviewImages.length}` }}</span>
                  </div>
                  <div v-if="galleryPreviewImages.length > 1" class="property-serviced-thumbs">
                    <div
                      v-for="(image, imageIndex) in servicedPreviewVisibleThumbs"
                      :key="image.id"
                      class="property-serviced-thumb"
                    >
                      <img
                        :src="image.url"
                        :alt="form.projectName"
                      />
                      <span
                        v-if="imageIndex === servicedPreviewVisibleThumbs.length - 1 && servicedPreviewHiddenThumbCount > 0"
                      >
                        +{{ servicedPreviewHiddenThumbCount }}
                      </span>
                    </div>
                  </div>
                </div>
              </section>

              <section class="property-serviced-card">
                <div class="property-serviced-card-title">
                  <div>
                    <span>{{ t('property.serviced.priceLabel') }}</span>
                    <h3>{{ t('property.editor.roomTypesSection') }}</h3>
                  </div>
                  <p>{{ t('property.editor.roomPriceCompanyNote') }}</p>
                </div>
                <div class="property-serviced-room-table">
                  <div class="property-serviced-room-head">
                    <span>{{ `${t('property.editor.roomTypeName')} / ${t('property.editor.roomTypeArea')}` }}</span>
                    <span>{{ `${t('property.editor.roomTypeRent')} (${t('property.editor.minStayField')})` }}</span>
                    <span>{{ t('property.editor.previewInquiryAction') }}</span>
                  </div>
                  <div
                    v-for="(room, roomIndex) in form.roomTypes"
                    :key="roomIndex"
                    class="property-serviced-room-row"
                  >
                    <span class="property-serviced-room-name">
                      <strong>{{ room.name }}</strong>
                      <small>{{ formatAreaSqft(room.usable_area_sqft) }}</small>
                    </span>
                    <span class="property-serviced-room-price">
                      <small>{{ formatStay(room.min_stay_value || room.min_lease_months, room.min_stay_unit || 'month') }}</small>
                      <strong>{{ formatServicedRoomRent(room) }}</strong>
                    </span>
                    <a
                      v-if="form.servicedWhatsapp.trim()"
                      class="property-serviced-room-action"
                      :href="buildWhatsAppHref(form.servicedWhatsapp)"
                      target="_blank"
                      rel="noreferrer"
                      :aria-label="t('property.editor.whatsappInquiryAction')"
                    >
                      <AppIcon name="message" :size="17" />
                    </a>
                    <button
                      v-else
                      type="button"
                      class="property-serviced-room-action"
                      :aria-label="t('property.editor.previewInquiryAction')"
                    >
                      <AppIcon name="message" :size="17" />
                    </button>
                  </div>
                </div>
              </section>

              <section v-if="form.residenceInfoTags.length > 0 || form.serviceTags.length > 0" class="property-serviced-card">
                <div class="property-serviced-card-title">
                  <div>
                    <h3>{{ t('property.editor.residenceInfoSection') }}</h3>
                  </div>
                </div>
                <div class="property-serviced-description">
                  <p v-if="form.residenceInfoTags.length > 0">{{ form.residenceInfoTags.join(' ') }}</p>
                  <p v-if="form.serviceTags.length > 0">{{ form.serviceTags.join(' ') }}</p>
                </div>
              </section>

              <section class="property-serviced-card">
                <div class="property-serviced-card-title">
                  <div>
                    <h3>{{ t('property.editor.facilitiesSection') }}</h3>
                  </div>
                </div>
                <div class="property-serviced-facility-grid">
                  <div
                    v-for="item in servicedPreviewFacilityItems"
                    :key="item.label"
                    :class="item.active ? 'property-serviced-facility--active' : 'property-serviced-facility--inactive'"
                  >
                    <AppIcon :name="item.icon" :size="26" />
                    <span>{{ item.label }}</span>
                  </div>
                </div>
              </section>

              <section
                v-if="form.extraChargeTags.length > 0"
                class="property-serviced-card"
              >
                <div class="property-serviced-card-title">
                  <div>
                    <h3>{{ t('property.editor.extraChargesField') }}</h3>
                  </div>
                </div>
                <ul class="property-serviced-fee-list">
                  <li
                    v-for="item in form.extraChargeTags"
                    :key="item"
                  >
                    {{ item }}
                  </li>
                </ul>
              </section>
            </div>

            <aside class="property-serviced-side">
              <section class="property-serviced-side-card property-serviced-summary-card">
                <div class="property-serviced-card-header">
                  <p class="property-serviced-side-label">{{ t('property.serviced.bookingTitle') }}</p>
                </div>
                <div class="property-serviced-action-stack">
                  <a
                    v-if="form.servicedWhatsapp.trim()"
                    class="property-serviced-action property-serviced-action--primary"
                    :href="buildWhatsAppHref(form.servicedWhatsapp)"
                    target="_blank"
                    rel="noopener noreferrer"
                  >
                    <AppIcon name="message" :size="18" />
                    <span>{{ t('property.editor.whatsappInquiryAction') }}</span>
                  </a>
                  <button
                    v-else
                    type="button"
                    class="property-serviced-action property-serviced-action--primary"
                  >
                    <AppIcon name="message" :size="18" />
                    <span>{{ t('property.editor.whatsappInquiryAction') }}</span>
                  </button>
                  <div class="property-serviced-secondary-actions">
                    <a
                      v-if="form.websiteUrl.trim()"
                      class="property-serviced-action property-serviced-action--secondary-icon"
                      :href="form.websiteUrl"
                      target="_blank"
                      rel="noopener noreferrer"
                      :title="t('property.editor.visitWebsiteAction')"
                    >
                      <AppIcon name="globe" :size="18" />
                    </a>
                  </div>
                </div>
              </section>
            </aside>
          </section>
        </template>

        <template v-if="activeStep === 'complete'">
          <section class="property-editor-panel property-complete-panel">
            <h2>{{ t('property.editor.completeTitle') }}</h2>
            <p>{{ t('property.editor.completeDescription') }}</p>
            <button type="button" class="property-editor-action property-editor-action--primary" @click="finishEditor">
              {{ t('property.editor.completeAction') }}
            </button>
          </section>
        </template>
      </form>

      <aside class="property-editor-side">
        <section class="property-editor-panel">
          <h2>{{ form.title || pageTitle }}</h2>
          <p>{{ form.summary }}</p>
          <strong>{{ previewPrice }}</strong>
          <p class="property-editor-charge">
            {{ chargeHint }}
          </p>
          <p
            v-if="isSale"
            class="property-editor-charge"
          >
            {{ selectedAdPackageText }}
          </p>
          <button
            v-if="activeStepIndex > 0 && activeStep !== 'complete'"
            type="button"
            class="property-editor-action property-editor-action--secondary"
            :disabled="saving || publishing"
            @click="goPreviousStep"
          >
            {{ t('property.editor.previousStep') }}
          </button>
          <button
            v-if="activeStep !== 'preview' && activeStep !== 'complete'"
            type="button"
            class="property-editor-action property-editor-action--secondary"
            :disabled="saving || publishing"
            @click="goNextStep"
          >
            {{ t('property.editor.nextStep') }}
          </button>
          <button
            v-if="activeStep === 'content'"
            type="button"
            class="property-editor-action property-editor-action--secondary"
            :disabled="saving || publishing"
            @click="saveAndReturn"
          >
            {{ saving ? t('common.status.loading') : t('property.editor.saveDraft') }}
          </button>
          <button
            v-if="activeStep === 'preview'"
            type="button"
            class="property-editor-action property-editor-action--primary"
            :disabled="saving || publishing"
            @click="saveAndPublish"
          >
            {{ publishing ? t('common.status.loading') : props.staffMode ? t('property.editor.saveDraft') : t('property.editor.publishNow') }}
          </button>
          <button
            v-if="activeStep === 'complete'"
            type="button"
            class="property-editor-action property-editor-action--primary"
            @click="finishEditor"
          >
            {{ t('property.editor.completeAction') }}
          </button>
        </section>
      </aside>
    </section>
  </main>
</template>

<style scoped>
.property-editor-page {
  width: 100%;
  max-width: var(--layout-page-max-width);
  margin: 0 auto;
  padding: 3rem var(--layout-page-padding-inline) 5rem;
  color: rgb(var(--color-text));
}

.property-editor-page--embedded {
  max-width: none;
  padding: 0;
}

.property-editor-page--serviced {
  --serviced-page-bg: rgb(var(--color-canvas));
  --serviced-panel-bg: rgb(var(--color-surface));
  --serviced-field-bg: rgb(var(--color-field-surface));
  --serviced-raised-bg: rgb(var(--color-surface-raised));
  --serviced-muted-bg: rgb(var(--color-surface-muted));
  --serviced-chip-bg: rgb(var(--color-primary-soft) / 0.48);
  --serviced-primary: rgb(var(--color-primary));
  --serviced-primary-contrast: rgb(var(--color-primary-contrast));
  --serviced-text: rgb(var(--color-text));
  --serviced-muted-text: rgb(var(--color-text-muted));
  --serviced-border: rgb(var(--color-border));
  --serviced-border-strong: rgb(var(--color-border-strong));
  --serviced-danger: rgb(var(--color-danger));
  --serviced-font-body: var(--font-sans);
  --serviced-font-headline: var(--font-display);
  min-height: min(82vh, 58rem);
  background: var(--serviced-page-bg);
  color: var(--serviced-text);
  font-family: var(--serviced-font-body);
}

.property-editor-page--serviced.property-editor-page--embedded {
  padding: 0;
}

.property-editor-page--serviced .property-editor-heading {
  margin-bottom: 2rem;
}

.property-editor-page--serviced .property-editor-heading h1 {
  color: var(--serviced-text);
  font-family: var(--serviced-font-headline);
  font-size: 3rem;
  font-weight: 400;
  letter-spacing: 0;
  line-height: 3.5rem;
}

.property-editor-page--serviced .property-editor-layout {
  gap: 1.5rem;
}

.property-editor-page--serviced .property-editor-steps {
  display: grid;
  gap: 1rem;
  padding-bottom: 0.25rem;
}

.property-editor-page--serviced .property-editor-steps button {
  min-height: 3.15rem;
  border-color: var(--serviced-border);
  border-radius: 4px;
  background: var(--serviced-panel-bg);
  padding: 0.75rem 1rem;
  color: var(--serviced-muted-text);
  font-size: 0.875rem;
  font-weight: 600;
  letter-spacing: 0.02em;
  box-shadow: none;
}

.property-editor-page--serviced .property-editor-steps span {
  width: 1.5rem;
  height: 1.5rem;
  background: var(--serviced-muted-bg);
  color: var(--serviced-muted-text);
  font-size: 0.75rem;
  font-weight: 500;
}

.property-editor-page--serviced .property-editor-steps .property-editor-step--active {
  border-color: var(--serviced-primary);
  background: var(--serviced-page-bg);
  color: var(--serviced-primary);
  box-shadow: var(--shadow-soft);
}

.property-editor-page--serviced .property-editor-steps .property-editor-step--active span,
.property-editor-page--serviced .property-editor-steps .property-editor-step--done span {
  background: var(--serviced-primary);
  color: var(--serviced-primary-contrast);
}

.property-editor-page--serviced .property-editor-form {
  gap: 2rem;
}

.property-editor-page--serviced .property-editor-panel {
  border-color: var(--serviced-border);
  border-radius: 4px;
  background: var(--serviced-panel-bg);
  padding: 1rem;
  box-shadow: none;
}

.property-editor-page--serviced .property-editor-panel h2 {
  margin-bottom: 1rem;
  color: var(--serviced-text);
  font-family: var(--serviced-font-headline);
  font-size: 1.5rem;
  font-weight: 400;
  line-height: 2rem;
}

.property-editor-page--serviced .property-panel-title-row {
  margin-bottom: 1rem;
}

.property-editor-page--serviced .property-editor-compact-grid {
  gap: 1rem;
  align-items: end;
}

.property-editor-page--serviced .property-serviced-basic-grid {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.property-editor-page--serviced .property-serviced-room-grid,
.property-editor-page--serviced .property-contact-grid {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.property-editor-page--serviced .property-input {
  gap: 0.25rem;
  color: var(--serviced-muted-text);
  font-size: 0.875rem;
  font-weight: 600;
  letter-spacing: 0.02em;
}

.property-editor-page--serviced .property-input b {
  color: var(--serviced-danger);
}

.property-editor-page--serviced .property-input input,
.property-editor-page--serviced .property-input select,
.property-editor-page--serviced .property-input textarea,
.property-editor-page--serviced .property-tag-input-row input {
  min-height: 2.65rem;
  border-color: var(--serviced-border);
  border-radius: 4px;
  background: var(--serviced-field-bg);
  padding: 0.5rem;
  color: var(--serviced-text);
  font-size: 1rem;
  font-weight: 400;
  line-height: 1.5rem;
}

.property-editor-page--serviced .property-input input:focus,
.property-editor-page--serviced .property-input select:focus,
.property-editor-page--serviced .property-input textarea:focus,
.property-editor-page--serviced .property-tag-input-row input:focus {
  border-color: var(--serviced-primary);
}

.property-editor-page--serviced .property-money-input {
  grid-template-columns: minmax(0, 1fr) minmax(5rem, auto);
  gap: 0.5rem;
}

.property-editor-page--serviced .property-room-editor {
  border-color: var(--serviced-border);
  border-radius: 4px;
  background: var(--serviced-raised-bg);
  padding: 0.5rem;
}

.property-editor-page--serviced .property-checkbox-row {
  gap: 1rem;
  margin-top: 1rem;
}

.property-editor-page--serviced .property-checkbox-row label,
.property-editor-page--serviced .property-inline-checkbox {
  min-height: 2.25rem;
  border: 0;
  border-radius: 0;
  background: transparent;
  padding: 0;
  color: var(--serviced-muted-text);
  font-size: 1rem;
  font-weight: 400;
}

.property-editor-page--serviced input[type='checkbox'] {
  accent-color: var(--serviced-primary);
}

.property-editor-page--serviced .property-editor-mini-action {
  min-height: 2.5rem;
  border-color: var(--serviced-border);
  border-radius: 4px;
  background: var(--serviced-field-bg);
  color: var(--serviced-text);
  font-size: 0.875rem;
  font-weight: 600;
}

.property-editor-page--serviced .property-tag-input-block {
  border-top: 0;
  padding-top: 0;
}

.property-editor-page--serviced .property-tag-input-block + .property-tag-input-block {
  margin-top: 1.5rem;
}

.property-editor-page--serviced .property-tag-input-block h3 {
  color: var(--serviced-text);
  font-family: var(--serviced-font-headline);
  font-size: 1.5rem;
  font-weight: 400;
  line-height: 2rem;
}

.property-editor-page--serviced .property-tag-input-row {
  gap: 0.5rem;
}

.property-editor-page--serviced .property-tag-input-row:focus-within {
  outline: 1px solid var(--serviced-primary);
  outline-offset: 0;
}

.property-editor-page--serviced .property-tag-chip,
.property-editor-page--serviced .property-quick-tag-row button {
  min-height: 1.9rem;
  border-color: var(--serviced-border);
  border-radius: 4px;
  background: var(--serviced-chip-bg);
  padding: 0.25rem 0.75rem;
  color: var(--serviced-muted-text);
  font-size: 0.875rem;
  font-weight: 400;
}

.property-editor-page--serviced .property-tag-chip i {
  background: transparent;
  color: var(--serviced-muted-text);
}

.property-editor-page--serviced .property-editor-note {
  color: var(--serviced-muted-text);
  font-size: 1rem;
  font-weight: 400;
}

.property-editor-page--serviced .property-upload-button {
  min-height: 2.5rem;
  border-color: var(--serviced-primary);
  border-radius: 4px;
  background: var(--serviced-field-bg);
  color: var(--serviced-primary);
  font-size: 0.875rem;
  font-weight: 600;
}

.property-editor-page--serviced .property-image-card {
  border-color: var(--serviced-border);
  border-radius: 4px;
  background: var(--serviced-field-bg);
}

.property-editor-page--serviced .property-editor-side {
  align-content: start;
}

.property-editor-page--serviced .property-editor-side .property-editor-panel {
  box-shadow: var(--shadow-soft);
}

.property-editor-page--serviced .property-editor-side h2 {
  margin-bottom: 0.5rem;
  font-size: 1.5rem;
}

.property-editor-page--serviced .property-editor-panel strong {
  margin: 0.65rem 0 1rem;
  color: var(--serviced-text);
  font-family: var(--serviced-font-headline);
  font-size: 2rem;
  font-weight: 400;
  line-height: 2.5rem;
}

.property-editor-page--serviced .property-editor-charge {
  color: var(--serviced-muted-text);
  font-size: 1rem;
  font-weight: 400;
}

.property-editor-page--serviced .property-editor-charge strong {
  display: inline;
  margin: 0;
  font-family: inherit;
  font-size: inherit;
  font-weight: 700;
  line-height: inherit;
}

.property-editor-page--serviced .property-editor-action {
  min-height: 3rem;
  border-color: var(--serviced-border);
  border-radius: 4px;
  background: var(--serviced-field-bg);
  color: var(--serviced-text);
  font-size: 0.875rem;
  font-weight: 600;
}

:global(html[data-theme='dark-neutral']) .property-editor-page--serviced {
  --serviced-chip-bg: rgb(var(--color-primary-soft) / 0.58);
}

.property-editor-page--serviced .property-editor-action--primary {
  border-color: var(--serviced-primary);
  background: var(--serviced-primary);
  color: var(--serviced-primary-contrast);
}

.property-editor-heading {
  display: flex;
  align-items: start;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 1.25rem;
}

.property-kicker {
  margin: 0 0 0.6rem;
  color: rgb(var(--color-primary));
  font-size: 0.75rem;
  font-weight: 900;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.property-editor-heading h1 {
  margin: 0;
  font-family: var(--font-display);
  font-size: clamp(2rem, 5vw, 3.8rem);
}

.property-editor-heading p:last-child {
  margin: 0.5rem 0 0;
  color: rgb(var(--color-text-muted));
  font-weight: 900;
}

.property-editor-close {
  display: inline-flex;
  width: 2.4rem;
  height: 2.4rem;
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.5rem;
  background: rgb(var(--color-surface-raised));
  color: rgb(var(--color-text));
}

.property-editor-layout {
  display: grid;
  gap: 1rem;
}

.property-editor-steps {
  display: grid;
  gap: 0.5rem;
}

.property-editor-steps button {
  display: inline-flex;
  min-height: 2.75rem;
  align-items: center;
  gap: 0.65rem;
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.5rem;
  background: rgb(var(--color-surface));
  padding: 0 0.8rem;
  color: rgb(var(--color-text-muted));
  font-weight: 900;
  text-align: left;
}

.property-editor-steps span {
  display: inline-flex;
  width: 1.6rem;
  height: 1.6rem;
  align-items: center;
  justify-content: center;
  border-radius: 999px;
  background: rgb(var(--color-surface-raised));
  color: rgb(var(--color-text));
  font-size: 0.78rem;
}

.property-editor-steps .property-editor-step--active {
  border-color: rgb(var(--color-primary));
  color: rgb(var(--color-primary));
}

.property-editor-steps .property-editor-step--active span,
.property-editor-steps .property-editor-step--done span {
  background: rgb(var(--color-primary));
  color: rgb(var(--color-primary-contrast));
}

.property-editor-form {
  display: grid;
  gap: 1rem;
}

.property-editor-stack {
  display: grid;
  gap: 0.85rem;
}

.property-editor-compact-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.85rem;
  align-items: end;
}

.property-input--wide {
  grid-column: 1 / -1;
}

.property-price-box,
.property-rental-fields {
  display: grid;
  gap: 0.75rem;
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.65rem;
  padding: 0.85rem;
}

.property-price-box--sale {
  background: rgb(var(--color-surface-raised));
}

.property-price-box--rent,
.property-rental-fields {
  background: rgb(var(--color-surface));
}

.property-money-input {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  align-items: center;
  gap: 0.65rem;
}

.property-money-input small {
  color: rgb(var(--color-text));
  font-size: 0.82rem;
  font-weight: 900;
  white-space: nowrap;
}

.property-editor-mini-action {
  min-height: 2.65rem;
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.5rem;
  background: rgb(var(--color-surface-raised));
  padding: 0 0.75rem;
  color: rgb(var(--color-text));
  font-size: 0.82rem;
  font-weight: 900;
}

.property-editor-mini-action:hover {
  border-color: rgb(var(--color-primary));
  color: rgb(var(--color-primary));
}

.property-editor-panel {
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.75rem;
  background: rgb(var(--color-surface));
  padding: 1rem;
  box-shadow: 0 10px 34px rgb(15 23 42 / 0.07);
}

.property-editor-panel h2 {
  margin: 0 0 1rem;
  font-size: 1.05rem;
  font-weight: 900;
}

.property-panel-title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 1rem;
}

.property-panel-title-row h2 {
  margin: 0;
}

.property-editor-error {
  display: block;
  margin-bottom: 0.75rem;
  color: rgb(var(--color-danger));
  font-size: 0.78rem;
  font-weight: 900;
}

.property-room-editor-stack {
  display: grid;
  gap: 0.75rem;
}

.property-room-editor {
  display: grid;
  gap: 0.75rem;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface-raised));
  padding: 0.85rem;
}

.property-editor-panel p {
  color: rgb(var(--color-text-muted));
  line-height: 1.6;
}

.property-editor-panel strong {
  display: block;
  margin: 1rem 0;
  font-family: var(--font-display);
  font-size: 1.8rem;
}

.property-editor-charge {
  margin: 0 0 1rem;
  color: rgb(var(--color-text));
  font-size: 0.85rem;
  font-weight: 900;
}

.property-choice-row,
.property-ad-package-grid {
  display: grid;
  gap: 0.75rem;
}

.property-choice-row button,
.property-ad-package-grid button {
  display: grid;
  gap: 0.35rem;
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.5rem;
  background: rgb(var(--color-surface-raised));
  padding: 0.85rem;
  color: rgb(var(--color-text));
  text-align: left;
}

.property-choice-row .property-choice--active,
.property-ad-package-grid .property-ad-package--active {
  border-color: rgb(var(--color-primary));
  box-shadow: 0 0 0 2px rgb(var(--color-primary) / 0.12);
}

.property-ad-package-grid strong {
  margin: 0;
  font-family: inherit;
  font-size: 1rem;
}

.property-ad-package-grid span,
.property-ad-package-grid small {
  color: rgb(var(--color-text-muted));
  font-weight: 800;
}

.property-preview-list {
  display: grid;
  gap: 0.75rem;
  margin: 0;
}

.property-preview-list div {
  display: grid;
  gap: 0.25rem;
  border-bottom: 1px solid rgb(var(--color-border));
  padding-bottom: 0.75rem;
}

.property-preview-list dt {
  color: rgb(var(--color-text-muted));
  font-size: 0.78rem;
  font-weight: 900;
}

.property-preview-list dd {
  margin: 0;
  color: rgb(var(--color-text));
  font-weight: 900;
}

.property-serviced-preview {
  display: grid;
  gap: 1rem;
}

.property-serviced-layout {
  --serviced-theme-card: 255 255 255;
  --serviced-theme-muted: 247 249 250;
  --serviced-theme-row: 241 245 249;
  --serviced-theme-primary: 18 91 83;
  --serviced-theme-text: 17 24 39;
  --serviced-theme-text-muted: 88 99 115;
  --serviced-theme-border: 222 227 232;
  --serviced-theme-outline: 145 155 166;
  grid-template-columns: minmax(0, 1fr);
  max-width: 100%;
  overflow: hidden;
  border-radius: 1rem;
  background: rgb(var(--serviced-theme-muted));
  padding: 1rem;
}

.property-serviced-content,
.property-serviced-side {
  display: grid;
  align-content: start;
  gap: 1.5rem;
  min-width: 0;
}

.property-serviced-card,
.property-serviced-side-card,
.property-serviced-hero {
  min-width: 0;
  border: 1px solid rgb(var(--serviced-theme-border));
  border-radius: 0.5rem;
  background: rgb(var(--serviced-theme-card));
  box-shadow: 0 16px 40px rgb(15 23 42 / 0.06);
}

.property-serviced-card {
  overflow: hidden;
}

.property-serviced-hero {
  display: grid;
  gap: 1rem;
  padding: 1.25rem;
}

.property-serviced-hero-copy {
  display: grid;
  align-content: start;
  gap: 0.9rem;
}

.property-serviced-kicker,
.property-serviced-side-label,
.property-serviced-card-title span {
  margin: 0;
  color: rgb(var(--serviced-theme-primary));
  font-size: 0.72rem;
  font-weight: 900;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.property-serviced-hero-copy h1 {
  margin: 0;
  color: rgb(var(--serviced-theme-text));
  font-family: var(--font-display);
  font-size: clamp(2.1rem, 4vw, 3.4rem);
  font-weight: 700;
  line-height: 1.03;
}

.property-serviced-subtitle {
  margin: 0;
  color: rgb(var(--serviced-theme-text-muted));
  font-size: 1rem;
  font-weight: 800;
  line-height: 1.45;
}

.property-serviced-address {
  display: flex;
  align-items: flex-start;
  gap: 0.45rem;
  color: rgb(var(--serviced-theme-text-muted));
  font-weight: 800;
  line-height: 1.5;
}

.property-serviced-address .app-icon {
  flex: 0 0 auto;
  margin-top: 0.15rem;
  color: rgb(var(--serviced-theme-primary));
}

.property-serviced-gallery-section {
  display: grid;
  gap: 0.25rem;
  min-width: 0;
  max-width: 100%;
}

.property-serviced-gallery-main {
  position: relative;
  overflow: hidden;
  width: 100%;
  max-width: 100%;
  aspect-ratio: 16 / 9;
  border: 1px solid rgb(var(--serviced-theme-border));
  border-radius: 0.5rem;
  background: rgb(var(--serviced-theme-row));
}

.property-serviced-gallery-main img,
.property-serviced-thumb img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.property-serviced-gallery-main > span {
  position: absolute;
  right: 1rem;
  bottom: 1rem;
  border-radius: 999px;
  background: rgb(17 24 39 / 0.78);
  padding: 0.35rem 0.65rem;
  color: #fff;
  font-size: 0.78rem;
  font-weight: 900;
}

.property-gallery__placeholder {
  display: grid;
  height: 100%;
  place-items: center;
  color: rgb(var(--color-text-muted));
}

.property-serviced-thumbs {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 0.25rem;
  margin-top: 0.25rem;
}

.property-serviced-thumb {
  position: relative;
  overflow: hidden;
  aspect-ratio: 1;
  border: 1px solid rgb(var(--serviced-theme-border));
  border-radius: 0.25rem;
}

.property-serviced-thumb span {
  position: absolute;
  inset: 0;
  display: grid;
  place-items: center;
  background: rgb(17 24 39 / 0.62);
  color: #fff;
  font-weight: 900;
}

.property-serviced-card-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  border-bottom: 1px solid rgb(var(--serviced-theme-border));
  padding: 1rem;
}

.property-serviced-card-title h3,
.property-serviced-card-title p {
  margin: 0;
}

.property-serviced-card-title h3 {
  font-size: 1.05rem;
  font-weight: 900;
}

.property-serviced-card-title p {
  color: rgb(var(--serviced-theme-text-muted));
  font-size: 0.78rem;
  font-weight: 800;
  text-align: right;
}

.property-serviced-room-table {
  display: grid;
}

.property-serviced-room-head,
.property-serviced-room-row {
  display: grid;
  grid-template-columns: minmax(12rem, 1.35fr) minmax(12rem, 1fr) 3rem;
  align-items: center;
  gap: 1rem;
  padding: 0.95rem 1rem;
}

.property-serviced-room-head {
  background: rgb(var(--serviced-theme-muted));
  color: rgb(var(--serviced-theme-text-muted));
  font-size: 0.8rem;
  font-weight: 900;
}

.property-serviced-room-row + .property-serviced-room-row {
  border-top: 1px solid rgb(var(--serviced-theme-border));
}

.property-serviced-room-row > span {
  display: grid;
  gap: 0.25rem;
  color: rgb(var(--serviced-theme-text-muted));
  font-size: 0.9rem;
  font-weight: 850;
}

.property-serviced-room-row strong {
  color: rgb(var(--serviced-theme-text));
  font-weight: 900;
}

.property-serviced-room-row small {
  color: rgb(var(--serviced-theme-text-muted));
  font-size: 0.78rem;
}

.property-serviced-room-action {
  display: inline-flex;
  width: 2.35rem;
  height: 2.35rem;
  align-items: center;
  justify-content: center;
  border: 1px solid #25d366;
  border-radius: 0.45rem;
  background: #25d366;
  color: #fff;
}

.property-serviced-description {
  display: grid;
  gap: 0.9rem;
  padding: 1.25rem;
}

.property-serviced-description p {
  margin: 0;
  color: rgb(var(--serviced-theme-text-muted));
  line-height: 1.85;
}

.property-serviced-facility-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.6rem;
  padding: 1rem;
}

.property-serviced-facility-grid > div {
  display: flex;
  min-height: 4.4rem;
  align-items: center;
  gap: 0.75rem;
  border: 1px solid rgb(var(--serviced-theme-border));
  border-radius: 0.5rem;
  padding: 0.9rem;
}

.property-serviced-facility-grid span {
  font-size: 0.82rem;
  font-weight: 900;
}

.property-serviced-facility--active {
  border-color: rgb(186 218 210);
  background: rgb(244 250 248);
  color: rgb(var(--serviced-theme-primary));
}

.property-serviced-facility--inactive {
  background: rgb(var(--serviced-theme-muted));
  color: rgb(var(--serviced-theme-outline));
  opacity: 0.58;
}

.property-serviced-fee-list {
  margin: 0;
  padding: 1rem 1rem 1rem 2.25rem;
  color: rgb(var(--serviced-theme-text-muted));
  line-height: 1.8;
}

.property-serviced-card-header {
  border-bottom: 1px solid rgb(var(--serviced-theme-border));
  background: rgb(var(--serviced-theme-muted));
  padding: 0.95rem 1.1rem;
}

.property-serviced-action-stack {
  display: grid;
  gap: 0.65rem;
  padding: 1rem;
}

.property-serviced-secondary-actions {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.property-serviced-action {
  display: inline-flex;
  width: 100%;
  min-height: 2.8rem;
  align-items: center;
  justify-content: center;
  gap: 0.6rem;
  border: 0;
  background: transparent;
  color: #fff;
  font-weight: 900;
  text-decoration: none;
}

.property-serviced-action--primary {
  background: #25d366;
}

.property-serviced-action--secondary-icon {
  border: 1px solid rgb(var(--serviced-theme-border));
  background: #fff;
  color: rgb(var(--serviced-theme-text));
}

@media (min-width: 1000px) {
  .property-serviced-layout {
    grid-template-columns: minmax(0, 1fr) 22rem;
    align-items: start;
  }

  .property-serviced-hero {
    grid-template-columns: 1fr;
  }

  .property-serviced-gallery-main {
    min-height: 31rem;
  }

  .property-serviced-side {
    position: sticky;
    top: calc(var(--app-header-offset, 0rem) + 2rem);
  }

  .property-serviced-facility-grid {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }
}

@media (max-width: 760px) {
  .property-serviced-layout {
    padding: 0.75rem;
  }

  .property-serviced-hero {
    padding: 0.9rem;
  }

  .property-serviced-hero-copy {
    order: 2;
  }

  .property-serviced-gallery-section {
    order: 1;
  }

  .property-serviced-room-head {
    display: none;
  }

  .property-serviced-room-row {
    grid-template-columns: minmax(0, 1fr) 2.35rem;
  }

  .property-serviced-room-row > span {
    grid-column: 1 / 2;
  }

  .property-serviced-room-row .property-serviced-room-action {
    grid-column: 2 / 3;
    grid-row: 1 / 2;
    align-self: start;
  }

  .property-serviced-facility-grid {
    grid-template-columns: 1fr;
  }
}

.property-serviced-preview.property-serviced-layout {
  grid-template-columns: 1fr;
}

.property-serviced-preview .property-serviced-gallery-main {
  min-height: 0;
}

.property-serviced-preview .property-serviced-side {
  position: static;
}

.property-serviced-preview .property-serviced-summary-card {
  max-width: none;
}

.property-serviced-preview__card {
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface-raised));
  padding: 1rem;
}

.property-serviced-preview__header {
  display: grid;
  gap: 0.45rem;
}

.property-serviced-preview__header h2 {
  margin: 0;
  font-family: var(--font-display);
  font-size: clamp(1.8rem, 4vw, 3rem);
  line-height: 1.08;
}

.property-serviced-preview__header p {
  margin: 0;
}

.property-serviced-preview__price {
  color: rgb(var(--color-primary));
  font-family: var(--font-display);
  font-size: 1.8rem;
  font-weight: 800;
}

.property-serviced-preview__gallery {
  display: grid;
  gap: 0.65rem;
}

.property-serviced-preview__gallery h2,
.property-serviced-preview__card > h2 {
  margin: 0 0 0.85rem;
  font-size: 1rem;
  font-weight: 900;
}

.property-serviced-preview__gallery > img,
.property-serviced-preview__empty-media {
  width: 100%;
  aspect-ratio: 16 / 9;
  border-radius: 8px;
  object-fit: cover;
}

.property-serviced-preview__empty-media {
  display: grid;
  place-items: center;
  border: 1px dashed rgb(var(--color-border));
  background: rgb(var(--color-surface-raised));
  padding: 1rem;
  color: rgb(var(--color-text-muted));
  font-size: 0.86rem;
  font-weight: 850;
  text-align: center;
}

.property-serviced-preview__thumbs {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 0.5rem;
}

.property-serviced-preview__thumbs img {
  width: 100%;
  aspect-ratio: 4 / 3;
  border-radius: 8px;
  object-fit: cover;
}

.property-detail-section {
  border-top: 1px solid rgb(var(--color-border));
  padding-top: 1.25rem;
}

.property-detail-section--plain {
  border-top: 0;
  padding-top: 0;
}

.property-detail-section h2 {
  margin: 0 0 0.85rem;
  font-size: 1rem;
}

.property-detail-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.property-detail-tags span {
  border-radius: 999px;
  background: rgb(var(--color-primary-soft) / 0.45);
  padding: 0.35rem 0.65rem;
  color: rgb(var(--color-primary));
  font-size: 0.76rem;
  font-weight: 850;
}

.property-room-table {
  display: grid;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  overflow: hidden;
}

.property-room-table__head,
.property-room-table__row {
  display: grid;
  grid-template-columns: minmax(12rem, 1.35fr) minmax(12rem, 1fr) minmax(6rem, 0.55fr);
  gap: 0.75rem;
  align-items: center;
  padding: 0.75rem;
}

.property-room-table__head {
  background: rgb(var(--color-surface-raised));
  color: rgb(var(--color-text-muted));
  font-size: 0.76rem;
  font-weight: 900;
}

.property-room-table__row + .property-room-table__row {
  border-top: 1px solid rgb(var(--color-border));
}

.property-room-table__row span {
  display: grid;
  gap: 0.2rem;
  min-width: 0;
  overflow-wrap: anywhere;
  font-size: 0.86rem;
  font-weight: 800;
}

.property-room-table__row small {
  color: rgb(var(--color-text-muted));
}

.property-serviced-preview__inquiry {
  display: inline-flex;
  min-height: 2.2rem;
  align-items: center;
  justify-content: center;
  border: 1px solid rgb(var(--color-primary));
  border-radius: 0.5rem;
  background: rgb(var(--color-primary));
  padding: 0 0.65rem;
  color: rgb(var(--color-primary-contrast));
  font-size: 0.78rem;
  font-weight: 900;
  text-decoration: none;
}

.property-spec-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.85rem;
  margin: 0;
}

.property-spec-grid dt {
  color: rgb(var(--color-text-muted));
  font-size: 0.75rem;
  font-weight: 900;
}

.property-spec-grid dd {
  margin: 0.25rem 0 0;
  font-weight: 800;
  line-height: 1.4;
  overflow-wrap: anywhere;
}

.property-serviced-preview__three {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 1rem;
}

.property-complete-panel {
  min-height: 18rem;
  align-content: center;
  text-align: center;
}

.property-editor-grid {
  display: grid;
  gap: 0.85rem;
}

.property-input {
  display: grid;
  gap: 0.4rem;
  color: rgb(var(--color-text-muted));
  font-size: 0.78rem;
  font-weight: 900;
}

.property-input b {
  margin-left: 0.2rem;
  color: rgb(var(--color-danger));
  font-weight: 950;
}

.property-input small {
  color: rgb(var(--color-danger));
  font-size: 0.74rem;
  font-weight: 900;
  line-height: 1.35;
}

.property-input input,
.property-input select,
.property-input textarea {
  width: 100%;
  min-height: 2.65rem;
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.5rem;
  background: rgb(var(--color-surface-raised));
  padding: 0.65rem 0.75rem;
  color: rgb(var(--color-text));
  font: inherit;
  outline: 0;
}

.property-input textarea {
  resize: vertical;
}

.property-address-input {
  position: relative;
  display: block;
  z-index: 20;
}

.property-address-input input {
  padding-right: 2.5rem;
}

.property-address-input i {
  position: absolute;
  top: 50%;
  right: 0.85rem;
  width: 1rem;
  height: 1rem;
  border: 2px solid rgb(var(--color-border));
  border-top-color: rgb(var(--color-primary));
  border-radius: 999px;
  transform: translateY(-50%);
  animation: property-spin 0.75s linear infinite;
}

@keyframes property-spin {
  to {
    transform: translateY(-50%) rotate(360deg);
  }
}

.property-address-suggestions {
  position: absolute;
  top: calc(100% + 0.25rem);
  right: 0;
  left: 0;
  z-index: 40;
  display: grid;
  gap: 0;
  max-height: 15rem;
  overflow: auto;
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.5rem;
  background: rgb(var(--color-surface));
  padding: 0.25rem;
  box-shadow: 0 16px 34px rgb(15 23 42 / 0.14);
}

.property-address-loading {
  margin: 0;
  padding: 0.45rem 0.55rem;
  color: rgb(var(--color-text-muted));
  font-size: 0.82rem;
  font-weight: 900;
  line-height: 1.2;
}

.property-address-suggestions button {
  display: grid;
  gap: 0.05rem;
  min-height: 2.3rem;
  border: 0;
  border-radius: 0.35rem;
  background: transparent;
  padding: 0.38rem 0.55rem;
  color: rgb(var(--color-text));
  text-align: left;
}

.property-address-suggestions button:hover,
.property-address-suggestions button:focus {
  background: rgb(var(--color-surface-raised));
}

.property-address-suggestions strong {
  overflow: hidden;
  font-size: 0.86rem;
  line-height: 1.2;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.property-address-suggestions button span {
  overflow: hidden;
  color: rgb(var(--color-text-muted));
  font-size: 0.76rem;
  font-weight: 800;
  line-height: 1.2;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.property-editor-note {
  color: rgb(var(--color-text-muted));
  font-size: 0.82rem;
  font-weight: 800;
}

.property-checkbox-row {
  display: flex;
  flex-wrap: wrap;
  gap: 0.65rem;
  margin-top: 1rem;
}

.property-tag-groups {
  display: grid;
  gap: 1rem;
  margin-top: 1rem;
}

.property-tag-group {
  display: grid;
  gap: 0.65rem;
}

.property-tag-group h3 {
  margin: 0;
  color: rgb(var(--color-text));
  font-size: 0.9rem;
  font-weight: 900;
}

.property-tag-group .property-checkbox-row {
  margin-top: 0;
}

.property-price-options {
  margin-top: 0;
}

.property-tag-input-block {
  display: grid;
  gap: 0.75rem;
  border-top: 1px solid rgb(var(--color-border));
  padding-top: 1rem;
}

.property-tag-input-block:first-of-type {
  border-top: 0;
  padding-top: 0;
}

.property-tag-input-block + .property-tag-input-block {
  margin-top: 1rem;
}

.property-tag-input-block h3 {
  margin: 0;
  font-size: 0.95rem;
  font-weight: 900;
}

.property-tag-input-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 0.65rem;
}

.property-tag-input-row input {
  min-height: 2.65rem;
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.5rem;
  background: rgb(var(--color-surface-raised));
  padding: 0.65rem 0.75rem;
  color: rgb(var(--color-text));
  font: inherit;
  outline: 0;
}

.property-tag-chip-row,
.property-quick-tag-row {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.property-tag-chip,
.property-quick-tag-row button {
  display: inline-flex;
  min-height: 2.2rem;
  align-items: center;
  gap: 0.45rem;
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.5rem;
  background: rgb(var(--color-surface-raised));
  padding: 0.35rem 0.6rem;
  color: rgb(var(--color-text));
  font-size: 0.82rem;
  font-weight: 900;
}

.property-tag-chip i {
  display: inline-flex;
  width: 1.15rem;
  height: 1.15rem;
  align-items: center;
  justify-content: center;
  border-radius: 999px;
  background: rgb(var(--color-danger));
  color: white;
  font-style: normal;
  line-height: 1;
}

.property-quick-tag-row button {
  color: rgb(var(--color-text-muted));
}

.property-checkbox-row label,
.property-inline-checkbox {
  display: inline-flex;
  min-height: 2.25rem;
  align-items: center;
  gap: 0.45rem;
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.5rem;
  background: rgb(var(--color-surface-raised));
  padding: 0.45rem 0.65rem;
  font-size: 0.82rem;
  font-weight: 900;
}

.property-upload-button,
.property-editor-action {
  display: inline-flex;
  min-height: 2.75rem;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  border-radius: 0.5rem;
  padding: 0 1rem;
  font-weight: 900;
}

.property-upload-button {
  border: 1px solid rgb(var(--color-primary));
  color: rgb(var(--color-primary));
  cursor: pointer;
}

.property-upload-button input {
  display: none;
}

.property-image-grid {
  display: grid;
  gap: 0.75rem;
  margin-top: 1rem;
}

.property-image-card {
  overflow: hidden;
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.5rem;
  background: rgb(var(--color-surface-raised));
}

.property-image-card img {
  width: 100%;
  aspect-ratio: 4 / 3;
  object-fit: cover;
}

.property-image-card button {
  width: 50%;
  min-height: 2.25rem;
  border: 0;
  background: transparent;
  color: rgb(var(--color-text));
  font-weight: 900;
}

.property-editor-side {
  display: grid;
  align-content: start;
  gap: 1rem;
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
  background: rgb(var(--color-surface-raised));
  color: rgb(var(--color-text));
}

@media (min-width: 760px) {
  .property-editor-grid,
  .property-image-grid,
  .property-ad-package-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .property-editor-compact-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .property-editor-page--serviced .property-serviced-basic-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .property-editor-steps {
    grid-template-columns: repeat(4, minmax(0, 1fr));
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

  .property-editor-steps {
    grid-column: 1 / -1;
  }

  .property-editor-side {
    position: sticky;
    top: calc(var(--app-header-offset, 0rem) + 2rem);
  }
}

@media (max-width: 759px) {
  .property-editor-page--serviced .property-editor-heading h1 {
    font-size: 2rem;
    line-height: 2.5rem;
  }

  .property-editor-page--serviced .property-serviced-basic-grid,
  .property-editor-page--serviced .property-serviced-room-grid,
  .property-editor-page--serviced .property-contact-grid {
    grid-template-columns: 1fr;
  }
}
</style>

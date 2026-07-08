<!--
 * 物業發布編輯頁。
 * 1. 根據頻道建立或更新樓盤放售與服務式住宅草稿。
 * 2. 支援圖片上傳、草稿保存與發布。
 * 3. 支援會員中心內嵌彈窗使用，處理未儲存離開確認。
-->
<script setup lang="ts">
import axios from 'axios';
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue';
import { onBeforeRouteLeave, useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';

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
  getPropertySaleFieldProfile,
  getPropertyOptionLabel,
  propertyAccountPackageOptions,
  propertyAdPackageOptions,
  propertyApplianceTagOptions,
  propertyBathroomCountOptions,
  propertyCookingModeOptions,
  propertyContactMethodOptions,
  propertyDefaultAvatarOptions,
  propertyDirectionOptions,
  propertyFeatureTagOptions,
  propertyFitoutTagOptions,
  propertyFloorDisplayOptions,
  propertyFurnitureTagOptions,
  propertyKitchenTypeOptions,
  propertyListingCategoryOptions,
  propertyListingTypeOptions,
  propertyLocationAreas,
  propertyLocationDistricts,
  propertyLocationSubdistricts,
  propertyLocationScopeOptions,
  propertyPublisherFilterOptions,
  propertyRentIncludedOptions,
  propertyRoomCountOptions,
  propertyTransactionTypeOptions,
  propertyViewTagOptions,
  industrialSpecialFeatureTagOptions,
  residentialSpecialFeatureTagOptions,
  propertyToiletCountOptions,
  servicedAdPackageOptions,
  servicedFacilityTagOptions,
  servicedRoomCategoryOptions,
  servicedServiceTagOptions,
  servicedStayUnitOptions,
  resolvePropertyLocationSelection,
} from '@/constants/property';
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

void propertyApplianceTagOptions;
void propertyAccountPackageOptions;
void propertyBathroomCountOptions;
void propertyCookingModeOptions;
void propertyDefaultAvatarOptions;
void propertyDirectionOptions;
void propertyFloorDisplayOptions;
void propertyKitchenTypeOptions;
void propertyListingCategoryOptions;
void propertyLocationScopeOptions;
void propertyRentIncludedOptions;
void propertyRoomCountOptions;
void propertyTransactionTypeOptions;

// 電話國家區碼選項
const phoneCountryCodeOptions = [
  { value: '+852', label: '+852 香港' },
  { value: '+86', label: '+86 中國大陸' },
];
void propertyToiletCountOptions;
void servicedAdPackageOptions;
void servicedRoomCategoryOptions;
void servicedStayUnitOptions;

const props = withDefaults(defineProps<{
  channel: PropertyChannel;
  returnPath?: string;
  listingId?: string;
  embedded?: boolean;
  hideHeader?: boolean;
  hideProgress?: boolean;
  staffMode?: boolean;
}>(), {
  returnPath: '',
  listingId: '',
  embedded: false,
  hideHeader: false,
  hideProgress: false,
  staffMode: false,
});

const emit = defineEmits<{
  (event: 'saved', listingId: string): void;
  (event: 'published', listingId: string): void;
  (event: 'cancel'): void;
  (event: 'step-change', payload: { activeIndex: number; total: number }): void;
}>();

type PropertyEditorLeaveDecision = 'save' | 'discard' | 'stay';
type PropertyEditorStepKey = 'category' | 'ad' | 'details' | 'contact';

interface PropertyEditorImage {
  id: string;
  mediaAssetId?: string;
  url?: string;
  file?: File;
  objectUrl?: string;
  isCover: boolean;
  uploading: boolean;
}

interface PropertyEditorStep {
  key: PropertyEditorStepKey;
  label: string;
}

type ResidentialBasicTextFieldKey = 'addressText' | 'addressTextEn';

interface ResidentialBasicTextField {
  key: ResidentialBasicTextFieldKey;
  label: string;
  wide?: boolean;
}

interface PropertyEditorForm {
  title: string;
  titleEn: string;
  summary: string;
  description: string;
  descriptionEn: string;
  locationAreaCode: string;
  locationDistrictCode: string;
  districtCode: string;
  communityId: string;
  publisherIdentityType: string;
  businessStatus: 'available' | 'sold';
  contactMethod: 'phone' | 'whatsapp' | 'chat' | 'both' | 'chat_or_whatsapp';
  contactAttributes: Record<string, string>;
  contactNameZh: string;
  contactNameEn: string;
  phone: string;
  phone2: string;
  whatsapp: string;
  wechat: string;
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
  propertyAttributes: Record<string, string>;
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
  rentIncludedItems: string[];
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
  projectAttributes: Record<string, string>;
  websiteURL: string;
  serviceWhatsApp: string;
  fax: string;
  serviceIntro: string;
  serviceIntroEn: string;
  benefitsText: string;
  benefitsTextEn: string;
  extraChargesText: string;
  extraChargesTextEn: string;
  lowestMonthlyRentHKD: number;
  highestMonthlyRentHKD: number;
  lowestDailyRentHKD: number;
  priceReferenceOnlyServiced: boolean;
  priceNegotiableServiced: boolean;
  minUsableAreaSqft: number;
  maxUsableAreaSqft: number;
  minLeaseMonths: number;
  minStayValue: number;
  minStayUnit: 'day' | 'week' | 'month';
  facilityTags: string[];
  serviceTags: string[];
  roomTypes: ServicedApartmentRoomType[];
}

const maxImages = 40;
const maxImageSize = 10 * 1024 * 1024;

// 1. 建立服務式住宅房型
const createServicedRoomType = (): ServicedApartmentRoomType => ({
  name: 'Studio',
  name_en: '',
  room_category: '',
  usable_area_sqft: 0,
  usable_area_min_sqft: 0,
  usable_area_max_sqft: 0,
  monthly_rent_hkd: 0,
  monthly_rent_min_hkd: 0,
  monthly_rent_max_hkd: 0,
  daily_rent_min_hkd: 0,
  daily_rent_max_hkd: 0,
  rent_unit: 'month',
  rent_suffix_plus: false,
  included_fees: true,
  included_fee_items: [],
  min_lease_months: 1,
  min_stay_value: 1,
  min_stay_unit: 'month',
  feature_tags: [],
  page_url: '',
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
const activeEditorStep = ref<PropertyEditorStepKey>('category');
let resolveLeavePrompt: ((decision: PropertyEditorLeaveDecision) => void) | null = null;

const initialLocationSelection = resolvePropertyLocationSelection(sessionStore.me?.district_code || '');

const form = reactive<PropertyEditorForm>({
  title: '',
  titleEn: '',
  summary: '',
  description: '',
  descriptionEn: '',
  locationAreaCode: initialLocationSelection.areaCode,
  locationDistrictCode: initialLocationSelection.districtCode,
  districtCode: initialLocationSelection.subdistrictCode,
  communityId: sessionStore.me?.primary_community?.public_id || '',
  publisherIdentityType: sessionStore.me?.publisher_identity_type || 'owner',
  businessStatus: 'available',
  contactMethod: 'both',
  contactAttributes: {
    phone_country_code: '+852',
    phone_whatsapp_enabled: '',
    hide_phone_allow_inquiry: '',
    default_avatar_gender: 'male',
    agency_company_profile: '',
    agency_contact_profile: '',
  },
  contactNameZh: '',
  contactNameEn: '',
  phone: '',
  phone2: '',
  whatsapp: '',
  wechat: '',
  email: sessionStore.me?.email || '',
  allowPhone: true,
  allowWhatsapp: true,
  allowChat: true,
  adPackageCode: 'basic',
  propertyNo: '',
  transactionType: 'sale',
  locationScope: 'local',
  listingCategory: 'standard',
  multiUnitProject: false,
  propertyType: 'residential',
  propertyAttributes: {
    account_or_package: 'personal_account',
  },
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
  annualPrepayOption: 'none',
  leaseStartDate: '',
  rentIncluded: '',
  rentIncludedItems: [],
  areaMode: 'usable',
  usableAreaSqft: 0,
  grossAreaSqft: 0,
  bedroomCount: 0,
  livingRoomCount: 0,
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
  kitchenType: 'enclosed',
  cookingMode: 'gas',
  managementFeeHKD: 0,
  videoURL: '',
  vrURL: '',
  privateNote: '',
  featureTags: [],
  projectName: '',
  projectNameEn: '',
  projectAttributes: {
    account_or_package: 'personal_account',
    address_street: '',
    address_doorplate: '',
    facility_custom_text: '',
    service_custom_text: '',
  },
  websiteURL: '',
  serviceWhatsApp: '',
  fax: '',
  serviceIntro: '',
  serviceIntroEn: '',
  benefitsText: '',
  benefitsTextEn: '',
  extraChargesText: '',
  extraChargesTextEn: '',
  lowestMonthlyRentHKD: 0,
  highestMonthlyRentHKD: 0,
  lowestDailyRentHKD: 0,
  priceReferenceOnlyServiced: false,
  priceNegotiableServiced: false,
  minUsableAreaSqft: 0,
  maxUsableAreaSqft: 0,
  minLeaseMonths: 1,
  minStayValue: 1,
  minStayUnit: 'month',
  facilityTags: [],
  serviceTags: [],
  roomTypes: [createServicedRoomType()],
});

const isSale = computed(() => props.channel === 'sale');
const isEditing = computed(() => listingId.value.trim().length > 0);
const isAgentPublisher = computed(() => form.publisherIdentityType === 'agent');
const isSaleOwnerPublisher = computed(() => isSale.value && !isAgentPublisher.value);
const isServicedPublisher = computed(() => !isSale.value);
const pageTitle = computed(() =>
  isSale.value ? t('property.sale.publishTitle') : t('property.serviced.publishTitle'),
);
const publisherOptions = computed(() =>
  propertyPublisherFilterOptions.filter((option) => {
    if (!option.value.trim()) {
      return false;
    }
    return isSale.value
      ? ['owner', 'agent'].includes(option.value)
      : ['owner', 'operator'].includes(option.value);
  }),
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
const publisherIdentityText = computed(() =>
  publisherOptions.value.find((option) => option.value === form.publisherIdentityType)?.label ??
  form.publisherIdentityType,
);
const transactionTypeText = computed(() => {
  const selected = propertyTransactionTypeOptions.find((option) => option.value === form.transactionType);
  return selected ? getPropertyOptionLabel(selected, preferenceStore.locale) : form.transactionType;
});
const propertyTypeText = computed(() => {
  const selected = propertyListingTypeOptions.find((option) => option.value === form.propertyType);
  return selected ? getPropertyOptionLabel(selected, preferenceStore.locale) : form.propertyType;
});
const previewListingCategoryText = computed(() => {
  if (!isSale.value) {
    return [publisherIdentityText.value, t('property.serviced.title')].filter(Boolean).join(' · ');
  }
  const selectedParts = [propertyTypeText.value, transactionTypeText.value].filter(Boolean);
  const categoryText = preferenceStore.locale === 'en' ? selectedParts.join(' ') : selectedParts.join('');
  return [publisherIdentityText.value, categoryText].filter(Boolean).join(' · ');
});
const hasImage = computed(() => images.value.some((image) => image.mediaAssetId || image.file));
const uploadedImageOptions = computed(() => images.value.filter((image) => image.mediaAssetId));
const saleFieldProfile = computed(() => getPropertySaleFieldProfile(form.propertyType));
const locationDistrictOptions = computed(() =>
  propertyLocationDistricts.filter((option) => option.areaCode === form.locationAreaCode),
);
const locationSubdistrictOptions = computed(() =>
  propertyLocationSubdistricts.filter((option) => option.districtCode === form.locationDistrictCode),
);
const isResidentialSaleEditor = computed(() => isSale.value && form.propertyType === 'residential');
const saleRequiresPropertyNo = computed(() => isAgentPublisher.value || form.propertyType === 'land');
const saleUsesToiletLabel = computed(() => ['industrial', 'shop'].includes(form.propertyType));
const saleBathroomCountOptions = computed(() =>
  saleUsesToiletLabel.value ? propertyToiletCountOptions : propertyBathroomCountOptions,
);
const saleCategoryTagOptions = computed(() => saleFieldProfile.value.categoryTags);
const residentialAddressTextFields = computed<ResidentialBasicTextField[]>(() => [
  { key: 'addressText', label: t('property.editor.addressField'), wide: true },
  { key: 'addressTextEn', label: t('property.editor.addressEnField'), wide: true },
]);
const residentialDirectFeatureOptions = computed(() =>
  ['pet_friendly', 'exclusive'].flatMap((value) =>
    propertyFeatureTagOptions.filter((option) => option.value === value),
  ),
);
const residentialFeatureTagGroups = computed(() => {
  const unitFeatureOptions = residentialSpecialFeatureTagOptions.filter((option) => {
    if (['parking_indoor', 'parking_outdoor'].includes(option.value)) {
      return false;
    }
    if (form.transactionType !== 'rent' && option.value === 'feature_student_friendly') {
      return false;
    }
    return true;
  });

  return [
    {
      title: t('property.editor.residentialTagView'),
      options: propertyViewTagOptions,
    },
    {
      title: t('property.editor.residentialTagFitout'),
      options: propertyFitoutTagOptions,
    },
    {
      title: t('property.editor.residentialTagAppliances'),
      options: propertyApplianceTagOptions,
    },
    {
      title: t('property.editor.residentialTagFurniture'),
      options: propertyFurnitureTagOptions,
    },
    {
      title: t('property.editor.residentialTagFeature'),
      options: residentialDirectFeatureOptions.value,
    },
    {
      title: t('property.editor.residentialTagParking'),
      options: residentialSpecialFeatureTagOptions.filter((option) =>
        ['parking_indoor', 'parking_outdoor'].includes(option.value),
      ),
    },
    {
      title: t('property.editor.residentialTagUnitFeature'),
      options: unitFeatureOptions,
    },
  ].filter((group) => group.options.length > 0);
});
const residentialFeatureTagOptions = computed(() =>
  residentialFeatureTagGroups.value.flatMap((group) => group.options),
);
const nonResidentialFeatureTagGroups = computed(() => {
  if (form.propertyType === 'car_park') {
    return [
      {
        title: t('property.editor.tagsField'),
        options: propertyFeatureTagOptions.filter((option) => option.value === 'exclusive'),
      },
    ];
  }
  if (form.propertyType === 'industrial') {
    return [
      {
        title: t('property.editor.residentialTagView'),
        options: propertyViewTagOptions,
      },
      {
        title: t('property.editor.residentialTagFitout'),
        options: propertyFitoutTagOptions,
      },
      {
        title: t('property.editor.residentialTagParking'),
        options: industrialSpecialFeatureTagOptions.filter((option) =>
          ['parking_indoor', 'parking_outdoor'].includes(option.value),
        ),
      },
      {
        title: t('property.editor.tagsField'),
        options: propertyFeatureTagOptions.filter((option) => option.value === 'exclusive'),
      },
      {
        title: t('property.editor.industrialUseTag'),
        options: industrialSpecialFeatureTagOptions.filter((option) =>
          [
            'feature_cctv',
            'feature_24h_access',
            'feature_mailbox',
            'feature_private_toilet',
            'feature_independent_ac',
            'feature_rooftop',
            'feature_flat_roof',
            'feature_whole_floor',
          ].includes(option.value),
        ),
      },
    ].filter((group) => group.options.length > 0);
  }
  if (form.propertyType === 'shop') {
    return [
      {
        title: t('property.editor.residentialTagFitout'),
        options: propertyFitoutTagOptions,
      },
      {
        title: t('property.editor.tagsField'),
        options: propertyFeatureTagOptions.filter((option) => option.value === 'exclusive'),
      },
    ];
  }

  return [];
});
const saleFeatureTagOptions = computed(() =>
  isResidentialSaleEditor.value
    ? residentialFeatureTagOptions.value
    : nonResidentialFeatureTagGroups.value.flatMap((group) => group.options),
);
const saleCategoryRequired = computed(() =>
  saleCategoryTagOptions.value.length > 0 && form.propertyType !== 'industrial',
);
const saleAllowedManualTagValues = computed(() =>
  new Set([
    ...saleCategoryTagOptions.value.map((tag) => tag.value),
    ...saleFeatureTagOptions.value.map((tag) => tag.value),
  ]),
);
const saleCategoryReady = computed(() =>
  !saleCategoryRequired.value ||
  saleCategoryTagOptions.value.some((tag) => form.featureTags.includes(tag.value)),
);
const saleRentIncludedOptions = computed(() => {
  const commonValues = ['rates_government_rent', 'management_fee'];
  const residentialValues = [...commonValues, 'gas', 'utilities'];
  const commercialValues = [...residentialValues, 'air_conditioning_fee'];
  const values = ['industrial', 'shop'].includes(form.propertyType)
    ? commercialValues
    : form.propertyType === 'land' || form.propertyType === 'car_park'
      ? commonValues
      : residentialValues;

  return propertyRentIncludedOptions.filter((option) => values.includes(option.value));
});
const saleAreaReady = computed(() => {
  if (saleFieldProfile.value.requiredArea === 'none') {
    return true;
  }
  if (saleFieldProfile.value.requiredArea === 'gross') {
    return form.grossAreaSqft > 0;
  }

  return form.usableAreaSqft > 0;
});
const saleFloorReady = computed(() =>
  !saleFieldProfile.value.floorRequired ||
  form.floorRaw.trim() !== '' ||
  form.floorLevel.trim() !== '',
);
const salePropertyNoReady = computed(() =>
  !saleRequiresPropertyNo.value ||
  form.propertyNo.trim() !== '',
);
const salePriceReady = computed(() =>
  form.priceNegotiable ||
  (form.transactionType === 'rent' ? form.monthlyRentHKD > 0 : form.askingPriceHKD > 0),
);
const servicedPriceReady = computed(() =>
  form.priceNegotiableServiced ||
  form.lowestMonthlyRentHKD > 0 ||
  form.roomTypes.some((room) =>
    Number(room.monthly_rent_min_hkd || room.monthly_rent_hkd || 0) > 0,
  ),
);
const servicedRoomRentUnitOptions = computed(() =>
  servicedStayUnitOptions.filter((option) => ['week', 'month'].includes(option.value)),
);
const contactChannelReady = computed(() =>
  form.allowPhone ||
  form.allowWhatsapp ||
  form.allowChat ||
  form.contactAttributes.hide_phone_allow_inquiry === 'yes',
);
const contactReady = computed(() => {
  if (isAgentPublisher.value) {
    return form.contactAttributes.agency_company_profile?.trim() !== '' &&
      form.contactAttributes.agency_contact_profile?.trim() !== '';
  }
  if (isServicedPublisher.value) {
    return contactChannelReady.value &&
      [form.phone, form.serviceWhatsApp, form.wechat, form.email].some((value) => value.trim() !== '');
  }

  return form.contactNameZh.trim() !== '' &&
    form.contactNameEn.trim() !== '' &&
    form.phone.trim() !== '';
});
const canSave = computed(() =>
  form.districtCode.trim() !== '' &&
  (isSale.value ? form.addressText.trim() !== '' : resolveServicedAddressText() !== '') &&
  contactReady.value &&
  (isSale.value
    ? form.title.trim() !== '' &&
      form.titleEn.trim() !== '' &&
      form.description.trim() !== '' &&
      form.descriptionEn.trim() !== '' &&
      form.addressTextEn.trim() !== '' &&
      salePriceReady.value &&
      (form.propertyType === 'land' || form.estateName.trim() !== '') &&
      saleAreaReady.value &&
      saleFloorReady.value &&
      salePropertyNoReady.value &&
      saleCategoryReady.value
    : form.projectName.trim() !== '' &&
      form.summary.trim() !== '' &&
      servicedPriceReady.value &&
      form.minStayValue > 0 &&
      form.roomTypes.length > 0 &&
      form.roomTypes.every((room) =>
        room.name.trim() !== '' &&
        Number(room.usable_area_min_sqft || room.usable_area_sqft || 0) > 0 &&
        Number(room.min_stay_value || 0) > 0,
      )),
);

// 必填欄位標記
const estateNameRequired = computed(() => form.propertyType !== 'land');
const grossAreaRequired = computed(() => saleFieldProfile.value.requiredArea === 'gross');
const usableAreaRequired = computed(() => saleFieldProfile.value.requiredArea === 'usable');
const floorFieldRequired = computed(() => saleFieldProfile.value.floorRequired);
const propertyNoFieldRequired = computed(() => saleRequiresPropertyNo.value);
const salePriceFieldRequired = computed(() => !form.priceNegotiable);
const servicedPriceFieldRequired = computed(() => !form.priceNegotiableServiced);
const isOwnerContactRequired = computed(() => isSaleOwnerPublisher.value);
const isAgentContactRequired = computed(() => isAgentPublisher.value);
const isServicedContactRequired = computed(() => isServicedPublisher.value);

const previewPrice = computed(() => {
  if ((isSale.value && form.priceNegotiable) || (!isSale.value && form.priceNegotiableServiced)) {
    return t('property.common.negotiable');
  }
  const value = isSale.value
    ? (form.transactionType === 'rent' ? form.monthlyRentHKD : form.askingPriceHKD * 10000)
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
const editorSteps = computed<PropertyEditorStep[]>(() => {
  const steps: PropertyEditorStep[] = [
    {
      key: 'category',
      label: isSale.value ? t('property.editor.stepCategory') : t('property.editor.servicedStepCategory'),
    },
    {
      key: 'ad',
      label: t('property.editor.adPackageField'),
    },
    {
      key: 'details',
      label: isSale.value ? t('property.editor.saleTitle') : t('property.editor.servicedTitle'),
    },
  ];

  if (isSale.value) {
    steps.push({
      key: 'contact',
      label: isResidentialSaleEditor.value ? t('property.editor.residentialModuleE') : t('property.editor.contact'),
    });
  }

  return steps;
});
const shouldShowMediaSection = computed(() =>
  activeEditorStep.value === 'details',
);
const activeEditorStepIndex = computed(() =>
  Math.max(0, editorSteps.value.findIndex((step) => step.key === activeEditorStep.value)),
);
const isFirstEditorStep = computed(() => activeEditorStepIndex.value <= 0);
const isLastEditorStep = computed(() => activeEditorStepIndex.value >= editorSteps.value.length - 1);
const currentSnapshot = computed(() => createEditorSnapshot());
const hasUnsavedChanges = computed(() =>
  savedSnapshot.value.length > 0 && currentSnapshot.value !== savedSnapshot.value,
);

watch(activeEditorStepIndex, (activeIndex) => {
  emit('step-change', {
    activeIndex,
    total: editorSteps.value.length,
  });
}, { immediate: true });

watch(() => form.locationAreaCode, () => {
  if (!locationDistrictOptions.value.some((option) => option.value === form.locationDistrictCode)) {
    form.locationDistrictCode = '';
  }
  if (!locationSubdistrictOptions.value.some((option) => option.value === form.districtCode)) {
    form.districtCode = '';
  }
});

watch(() => form.locationDistrictCode, () => {
  if (!locationSubdistrictOptions.value.some((option) => option.value === form.districtCode)) {
    form.districtCode = '';
  }
});

watch(() => form.districtCode, (value) => {
  const selection = resolvePropertyLocationSelection(value);
  if (!selection.subdistrictCode) {
    return;
  }
  if (form.locationAreaCode !== selection.areaCode) {
    form.locationAreaCode = selection.areaCode;
  }
  if (form.locationDistrictCode !== selection.districtCode) {
    form.locationDistrictCode = selection.districtCode;
  }
});

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

// 10. 切換發布步驟
const selectEditorStep = (step: PropertyEditorStepKey): void => {
  activeEditorStep.value = step;
};

// 11. 前往上一步
const goPreviousEditorStep = (): void => {
  const previousStep = editorSteps.value[activeEditorStepIndex.value - 1];
  if (previousStep) {
    activeEditorStep.value = previousStep.key;
  }
};

// 12. 前往下一步
const goNextEditorStep = (): void => {
  const nextStep = editorSteps.value[activeEditorStepIndex.value + 1];
  if (nextStep) {
    activeEditorStep.value = nextStep.key;
  }
};

// 13. 建立上傳目錄
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
    upload_token: presign.upload_token,
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

// 20. 格式化可留空數字輸入
const formatOptionalNumberInput = (value: number): string =>
  value > 0 ? String(value) : '';

// 21. 讀取可留空數字輸入
const readOptionalNumberInput = (event: Event): number => {
  const rawValue = (event.target as HTMLInputElement).value.trim();
  if (!rawValue) {
    return 0;
  }
  const numericValue = Number(rawValue);

  return Number.isFinite(numericValue) ? Math.max(0, numericValue) : 0;
};

// 20. 正規化服務式住宅租期單位
const normalizeServicedStayUnit = (value?: string): 'day' | 'week' | 'month' => {
  if (value === 'day' || value === 'week' || value === 'month') {
    return value;
  }

  return 'month';
};

// 21. 估算服務式住宅租期天數
const servicedStayDays = (value: number, unit?: string): number => {
  const safeValue = Math.max(1, Number(value || 1));
  const normalizedUnit = normalizeServicedStayUnit(unit);
  if (normalizedUnit === 'month') {
    return safeValue * 30;
  }
  if (normalizedUnit === 'week') {
    return safeValue * 7;
  }

  return safeValue;
};

// 22. 清理動態屬性
const sanitizeAttributeMap = (source: Record<string, string>): Record<string, string> => {
  const result: Record<string, string> = {};
  Object.entries(source).forEach(([key, value]) => {
    const normalizedKey = key.trim();
    const normalizedValue = String(value ?? '').trim();
    if (normalizedKey && normalizedValue) {
      result[normalizedKey] = normalizedValue;
    }
  });

  return result;
};

// 23. 套用樓盤地點代碼
const applyLocationCode = (value: string, preserveDetailedSelection = false): void => {
  const selection = resolvePropertyLocationSelection(value);
  if (preserveDetailedSelection && form.districtCode && !selection.subdistrictCode) {
    const currentSelection = resolvePropertyLocationSelection(form.districtCode);
    const sameArea = !selection.areaCode || selection.areaCode === currentSelection.areaCode;
    const sameDistrict = !selection.districtCode || selection.districtCode === currentSelection.districtCode;
    if (sameArea && sameDistrict) {
      return;
    }
  }
  form.locationAreaCode = selection.areaCode;
  form.locationDistrictCode = selection.districtCode;
  form.districtCode = selection.subdistrictCode;
};

// 24. 補充樓盤地點層級屬性
const appendLocationAttributes = (attributes: Record<string, string>): Record<string, string> => {
  const area = propertyLocationAreas.find((option) => option.value === form.locationAreaCode);
  const district = propertyLocationDistricts.find((option) => option.value === form.locationDistrictCode);
  const subdistrict = propertyLocationSubdistricts.find((option) => option.value === form.districtCode);
  if (area) {
    attributes.location_area_code = area.value;
    attributes.location_area_label_zh_hk = area.label_zh_hk;
    attributes.location_area_label_en = area.label_en;
  }
  if (district) {
    attributes.location_district_code = district.value;
    attributes.location_district_label_zh_hk = district.label_zh_hk;
    attributes.location_district_label_en = district.label_en;
  }
  if (subdistrict) {
    attributes.location_subdistrict_code = subdistrict.value;
    attributes.location_subdistrict_label_zh_hk = subdistrict.label_zh_hk;
    attributes.location_subdistrict_label_en = subdistrict.label_en;
  }

  return attributes;
};

// 21. 正規化樓盤放售主分類
const normalizeSalePropertyType = (value: string): string => {
  if (['private_flat', 'estate', 'house'].includes(value)) {
    return 'residential';
  }
  if (value === 'office') {
    return 'industrial';
  }

  return value || 'residential';
};

// 22. 建立樓盤分類專屬屬性
const buildSalePropertyAttributes = (): Record<string, string> => {
  const result = sanitizeAttributeMap(form.propertyAttributes);
  if (form.rentIncludedItems.length > 0) {
    result.rent_included_items = form.rentIncludedItems.join(',');
  }

  return appendLocationAttributes(result);
};

// 23. 建立服務式住宅項目屬性
const buildServicedProjectAttributes = (): Record<string, string> =>
  appendLocationAttributes(sanitizeAttributeMap(form.projectAttributes));

// 24. 建立聯絡資料屬性
const buildContactAttributes = (): Record<string, string> =>
  sanitizeAttributeMap(form.contactAttributes);

// 25. 建立租金包含文字
const buildRentIncludedText = (): string => {
  if (form.rentIncludedItems.length === 0) {
    return form.rentIncluded.trim();
  }

  return form.rentIncludedItems
    .flatMap((value) => {
      const option = saleRentIncludedOptions.value.find((item) => item.value === value);
      return option ? [getPropertyOptionLabel(option, preferenceStore.locale)] : [];
    })
    .join('、');
};

// 26. 取得當前樓盤分類允許提交的標籤
const buildSaleFeatureTags = (): string[] =>
  form.featureTags.filter((tag) => saleAllowedManualTagValues.value.has(tag));

// 27. 取得樓盤面積 payload 值
const resolveSaleUsableArea = (): number => {
  if (saleFieldProfile.value.requiredArea === 'none') {
    return Math.max(1, Number(form.usableAreaSqft || 0));
  }

  return Number(form.usableAreaSqft || form.grossAreaSqft || 0);
};

// 28. 取得樓盤樓層 payload 值
const resolveSaleFloorRaw = (): string =>
  form.floorRaw.trim() ||
  form.floorLevel.trim() ||
  saleFieldProfile.value.defaultFloorText;

// 29. 取得樓盤名稱 payload 值
const resolveSaleEstateName = (): string =>
  form.propertyType === 'land'
    ? form.estateName.trim() || form.propertyAttributes.lot_number?.trim() || form.addressText.trim()
    : form.estateName.trim();

// 30. 取得樓盤租售摘要 payload 值
const resolveSaleSummary = (): string =>
  form.description.trim().slice(0, 160);

// 31. 取得服務式住宅標題 payload 值
const resolveServicedTitle = (): string =>
  form.projectName.trim() || form.title.trim();

// 32. 取得服務式住宅描述 payload 值
const resolveServicedDescription = (): string =>
  form.summary.trim() || form.description.trim();

// 33. 取得服務式住宅英文摘要 payload 值
const resolveServicedSummaryEn = (): string =>
  form.projectAttributes.summary_en?.trim() || form.descriptionEn.trim();

// 34. 取得服務式住宅地址 payload 值
const resolveServicedAddressText = (): string => {
  const parts = [
    form.projectAttributes.address_street,
    form.projectAttributes.address_doorplate,
  ].map((part) => String(part ?? '').trim()).filter(Boolean);

  return parts.join(' ') || form.addressText.trim();
};

// 35. 更新服務式住宅房型包含項目
const updateRoomIncludedFeeItems = (room: ServicedApartmentRoomType, event: Event): void => {
  room.included_fee_items = splitTextList((event.target as HTMLInputElement).value);
};

// 30. 切換服務式住宅房型包含設施
const toggleRoomIncludedFeeItem = (room: ServicedApartmentRoomType, value: string): void => {
  if (!Array.isArray(room.included_fee_items)) {
    room.included_fee_items = [];
  }
  toggleTag(room.included_fee_items, value);
};

// 31. 整理服務式住宅房型
const normalizeServicedRoomType = (room: ServicedApartmentRoomType): ServicedApartmentRoomType => {
  const stayUnit = normalizeServicedStayUnit(room.min_stay_unit);
  const rentUnit = normalizeServicedStayUnit(room.rent_unit || stayUnit);
  const fallbackRent = form.priceReferenceOnlyServiced || form.priceNegotiableServiced ? 1 : 0;
  const usableAreaMin = Number(room.usable_area_min_sqft || room.usable_area_sqft || 0);
  const usableAreaMax = Number(room.usable_area_max_sqft || usableAreaMin || 0);
  const monthlyMin = rentUnit === 'day'
    ? 0
    : Number(room.monthly_rent_min_hkd || room.monthly_rent_hkd || fallbackRent);
  const monthlyMax = rentUnit === 'day'
    ? 0
    : Number(room.monthly_rent_max_hkd || monthlyMin || 0);
  const dailyMin = rentUnit === 'day'
    ? Number(room.daily_rent_min_hkd || fallbackRent)
    : Number(room.daily_rent_min_hkd || 0);
  const dailyMax = rentUnit === 'day'
    ? Number(room.daily_rent_max_hkd || dailyMin || 0)
    : Number(room.daily_rent_max_hkd || 0);

  return {
    ...room,
    name: room.name.trim() || 'Studio',
    name_en: room.name_en?.trim() || undefined,
    room_category: room.room_category?.trim() || undefined,
    usable_area_sqft: usableAreaMin,
    usable_area_min_sqft: usableAreaMin,
    usable_area_max_sqft: usableAreaMax > 0 ? usableAreaMax : undefined,
    monthly_rent_hkd: monthlyMin,
    monthly_rent_min_hkd: monthlyMin,
    monthly_rent_max_hkd: monthlyMax > 0 ? monthlyMax : undefined,
    daily_rent_min_hkd: dailyMin > 0 ? dailyMin : undefined,
    daily_rent_max_hkd: dailyMax > 0 ? dailyMax : undefined,
    rent_unit: rentUnit,
    rent_suffix_plus: Boolean(room.rent_suffix_plus),
    included_fees: Boolean(room.included_fees),
    included_fee_items: Array.isArray(room.included_fee_items) ? room.included_fee_items : [],
    min_lease_months: Number(room.min_lease_months || form.minLeaseMonths || 1),
    min_stay_value: Number(room.min_stay_value || room.min_lease_months || form.minStayValue || 1),
    min_stay_unit: stayUnit,
    feature_tags: room.feature_tags ?? [],
  };
};

// 31. 推導服務式住宅項目最短入住
const deriveServicedProjectMinStay = (roomTypes: ServicedApartmentRoomType[]) => {
  const normalizedRooms = roomTypes.length > 0 ? roomTypes : [createServicedRoomType()];
  const shortestRoom = normalizedRooms
    .map((room) => ({
      minStayValue: Number(room.min_stay_value || room.min_lease_months || form.minStayValue || 1),
      minStayUnit: normalizeServicedStayUnit(room.min_stay_unit || form.minStayUnit),
    }))
    .sort((left, right) =>
      servicedStayDays(left.minStayValue, left.minStayUnit) - servicedStayDays(right.minStayValue, right.minStayUnit),
    )[0];
  const minStayValue = shortestRoom.minStayValue;
  const minStayUnit = shortestRoom.minStayUnit;

  return {
    minLeaseMonths: Math.max(1, Math.ceil(servicedStayDays(minStayValue, minStayUnit) / 30)),
    minStayValue,
    minStayUnit,
  };
};

// 32. 新增服務式住宅房型
const addServicedRoomType = (): void => {
  form.roomTypes.push(createServicedRoomType());
};

// 33. 移除服務式住宅房型
const removeServicedRoomType = (roomIndex: number): void => {
  if (form.roomTypes.length <= 1) {
    return;
  }

  form.roomTypes.splice(roomIndex, 1);
};

// 34. 建立樓盤 payload
const buildSalePayload = (): UpsertPropertySalePayload => ({
  title: form.title.trim(),
  title_en: form.titleEn.trim() || undefined,
  summary: resolveSaleSummary(),
  description: form.description.trim(),
  description_en: form.descriptionEn.trim() || undefined,
  district_code: form.districtCode,
  community_id: form.communityId,
  publisher_identity_type: form.publisherIdentityType || 'owner',
  property_no: saleRequiresPropertyNo.value ? form.propertyNo.trim() || undefined : undefined,
  transaction_type: form.transactionType,
  location_scope: form.locationScope,
  listing_category: form.listingCategory,
  multi_unit_project: form.multiUnitProject,
  property_type: normalizeSalePropertyType(form.propertyType),
  property_attributes: buildSalePropertyAttributes(),
  rental_type: form.rentalType.trim() || undefined,
  renovation_type: undefined,
  agency_company_name: isAgentPublisher.value
    ? form.contactAttributes.agency_company_profile?.trim() || form.agencyCompanyName.trim() || undefined
    : undefined,
  estate_name: resolveSaleEstateName(),
  address_text: form.addressText.trim(),
  address_text_en: form.addressTextEn.trim() || undefined,
  block_name: form.blockName.trim() || undefined,
  unit_name: form.unitName.trim() || undefined,
  show_unit: form.showUnit,
  latitude: undefined,
  longitude: undefined,
  asking_price_hkd: form.transactionType === 'rent' ? 0 : Number(form.askingPriceHKD) * 10000,
  monthly_rent_hkd: form.transactionType === 'rent' ? Number(form.monthlyRentHKD) : undefined,
  price_reference_only: form.priceReferenceOnly,
  price_negotiable: form.priceNegotiable,
  annual_prepay_discount: form.transactionType === 'rent' ? form.annualPrepayDiscount : false,
  annual_prepay_option: form.transactionType === 'rent' && form.annualPrepayDiscount
    ? 'provided'
    : 'none',
  lease_start_date: form.transactionType === 'rent' ? form.leaseStartDate.trim() || undefined : undefined,
  rent_included: form.transactionType === 'rent' ? buildRentIncludedText() || undefined : undefined,
  area_mode: saleFieldProfile.value.requiredArea === 'gross' ? 'gross' : 'usable',
  usable_area_sqft: resolveSaleUsableArea(),
  gross_area_sqft: form.grossAreaSqft > 0 ? Number(form.grossAreaSqft) : undefined,
  bedroom_count: saleFieldProfile.value.showRooms ? Number(form.bedroomCount) : 0,
  living_room_count: 0,
  bathroom_count: saleFieldProfile.value.showRooms ? Number(form.bathroomCount) : 0,
  floor_level: resolveSaleFloorRaw(),
  floor_raw: resolveSaleFloorRaw(),
  floor_zone: undefined,
  total_floors: undefined,
  direction: saleFieldProfile.value.showDirection ? form.direction.trim() : '',
  building_age: form.buildingAge.trim(),
  completion_year: undefined,
  building_total_floors: undefined,
  management_company: undefined,
  kitchen_type: saleFieldProfile.value.showKitchen ? form.kitchenType || undefined : undefined,
  cooking_mode: saleFieldProfile.value.showKitchen ? form.cookingMode || undefined : undefined,
  management_fee_hkd: form.managementFeeHKD > 0 ? Number(form.managementFeeHKD) : undefined,
  video_url: form.videoURL.trim() || undefined,
  vr_url: form.vrURL.trim() || undefined,
  private_note: undefined,
  ad_package_code: form.adPackageCode,
  feature_tags: mergePropertyFeatureTags(buildSaleFeatureTags(), {
    video_url: form.videoURL.trim(),
    vr_url: form.vrURL.trim(),
    annual_prepay_discount: form.transactionType === 'rent' ? form.annualPrepayDiscount : false,
    annual_prepay_option: form.transactionType === 'rent' && form.annualPrepayDiscount
      ? 'provided'
      : 'none',
  }),
  contact_method: form.contactMethod,
  business_status: form.businessStatus,
  images: buildImagePayload(),
  contact: {
    contact_name_zh: isSaleOwnerPublisher.value ? form.contactNameZh.trim() || undefined : undefined,
    contact_name_en: isSaleOwnerPublisher.value ? form.contactNameEn.trim() || undefined : undefined,
    phone: isSaleOwnerPublisher.value ? form.phone.trim() : '',
    phone_2: isSaleOwnerPublisher.value ? form.phone2.trim() || undefined : undefined,
    whatsapp: isSaleOwnerPublisher.value && form.contactAttributes.phone_whatsapp_enabled === 'yes'
      ? form.phone.trim()
      : '',
    wechat: isSaleOwnerPublisher.value ? form.wechat.trim() || undefined : undefined,
    email: '',
    contact_attributes: buildContactAttributes(),
    show_phone: isSaleOwnerPublisher.value && form.contactAttributes.hide_phone_allow_inquiry !== 'yes',
    show_whatsapp: isSaleOwnerPublisher.value && form.contactAttributes.phone_whatsapp_enabled === 'yes',
    show_chat: isAgentPublisher.value,
    show_inquiry_form: isSaleOwnerPublisher.value && form.contactAttributes.hide_phone_allow_inquiry === 'yes',
  },
});

// 36. 建立服務式住宅 payload
const buildServicedPayload = (): UpsertServicedApartmentPayload => {
  const roomTypes = form.roomTypes.map(normalizeServicedRoomType);
  const monthlyPrices = roomTypes
    .map((room) => Number(room.monthly_rent_min_hkd || room.monthly_rent_hkd || 0))
    .filter((value) => value > 0);
  const monthlyMaxPrices = roomTypes
    .map((room) => Number(room.monthly_rent_max_hkd || room.monthly_rent_min_hkd || room.monthly_rent_hkd || 0))
    .filter((value) => value > 0);
  const dailyPrices = roomTypes
    .map((room) => Number(room.daily_rent_min_hkd || 0))
    .filter((value) => value > 0);
  const areas = roomTypes
    .map((room) => Number(room.usable_area_min_sqft || room.usable_area_sqft || 0))
    .filter((value) => value > 0);
  const maxAreas = roomTypes
    .map((room) => Number(room.usable_area_max_sqft || room.usable_area_min_sqft || room.usable_area_sqft || 0))
    .filter((value) => value > 0);
  const projectMinStay = deriveServicedProjectMinStay(roomTypes);

  return {
    title: resolveServicedTitle(),
    title_en: undefined,
    summary: form.summary.trim(),
    description: resolveServicedDescription(),
    description_en: resolveServicedSummaryEn() || undefined,
    district_code: form.districtCode,
    community_id: form.communityId,
    publisher_identity_type: form.publisherIdentityType || 'owner',
    project_name: form.projectName.trim(),
    project_name_en: form.projectNameEn.trim() || undefined,
    project_attributes: buildServicedProjectAttributes(),
    address_text: resolveServicedAddressText(),
    address_text_en: undefined,
    website_url: form.websiteURL.trim() || undefined,
    whatsapp: form.serviceWhatsApp.trim() || undefined,
    fax: form.fax.trim() || undefined,
    service_intro: form.serviceIntro.trim() || undefined,
    service_intro_en: form.serviceIntroEn.trim() || undefined,
    benefits_text: form.benefitsText.trim() || undefined,
    benefits_text_en: form.benefitsTextEn.trim() || undefined,
    extra_charges_text: form.extraChargesText.trim() || undefined,
    extra_charges_text_en: form.extraChargesTextEn.trim() || undefined,
    lowest_monthly_rent_hkd: monthlyPrices.length
      ? Math.min(...monthlyPrices)
      : Number(form.lowestMonthlyRentHKD || 0),
    highest_monthly_rent_hkd: monthlyMaxPrices.length
      ? Math.max(...monthlyMaxPrices)
      : form.highestMonthlyRentHKD > 0 ? Number(form.highestMonthlyRentHKD) : undefined,
    lowest_daily_rent_hkd: dailyPrices.length
      ? Math.min(...dailyPrices)
      : form.lowestDailyRentHKD > 0 ? Number(form.lowestDailyRentHKD) : undefined,
    price_reference_only: form.priceReferenceOnlyServiced,
    price_negotiable: form.priceNegotiableServiced,
    min_usable_area_sqft: areas.length
      ? Math.min(...areas)
      : form.minUsableAreaSqft > 0 ? Number(form.minUsableAreaSqft) : undefined,
    max_usable_area_sqft: maxAreas.length
      ? Math.max(...maxAreas)
      : form.maxUsableAreaSqft > 0 ? Number(form.maxUsableAreaSqft) : undefined,
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
      contact_name_zh: undefined,
      contact_name_en: undefined,
      phone: form.phone.trim(),
      phone_2: undefined,
      whatsapp: form.serviceWhatsApp.trim(),
      wechat: form.wechat.trim() || undefined,
      email: form.email.trim(),
      contact_attributes: buildContactAttributes(),
      show_phone: form.allowPhone,
      show_whatsapp: form.allowWhatsapp,
      show_chat: form.allowChat,
      show_inquiry_form: false,
    },
  };
};

// 33. 清除地址聯想計時器
const clearAddressSearchTimer = (): void => {
  if (addressSearchTimer.value) {
    window.clearTimeout(addressSearchTimer.value);
    addressSearchTimer.value = null;
  }
};

// 34. 取得地址聯想標題
const resolveAddressSuggestionTitle = (suggestion: PropertyAddressSuggestion): string =>
  suggestion.display_name ||
  [
    suggestion.estate_name,
    suggestion.estate_name_en && suggestion.estate_name_en !== suggestion.estate_name
      ? suggestion.estate_name_en
      : '',
    suggestion.district_label ? `(${suggestion.district_label})` : '',
  ].filter(Boolean).join(' ');

// 35. 讀取屋苑或住宅地址聯想
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
      district_code: form.locationDistrictCode || form.districtCode || undefined,
      limit: 12,
    });

    addressSuggestions.value = response.data.data;
  } catch {
    addressSuggestions.value = [];
  } finally {
    loadingAddressSuggestions.value = false;
  }
};

// 36. 延遲讀取地址聯想
const scheduleAddressSuggestions = (): void => {
  clearAddressSearchTimer();
  addressSearchTimer.value = window.setTimeout(() => {
    void loadAddressSuggestions();
  }, 220);
};

// 37. 套用地址聯想
const applyAddressSuggestion = (suggestion: PropertyAddressSuggestion): void => {
  const currentYear = new Date().getFullYear();

  if (isSale.value) {
    form.estateName = resolveAddressSuggestionTitle(suggestion);
  } else {
    form.projectName = suggestion.estate_name || resolveAddressSuggestionTitle(suggestion);
    form.projectNameEn = suggestion.estate_name_en;
    form.projectAttributes.address_street = suggestion.address_text;
    form.projectAttributes.address_doorplate = '';
  }

  form.addressText = suggestion.address_text;
  form.addressTextEn = suggestion.address_text_en;
  form.blockName = suggestion.block_names[0] ?? form.blockName;
  form.buildingAge = suggestion.completion_year > 0 && suggestion.completion_year <= currentYear
    ? String(currentYear - suggestion.completion_year)
    : form.buildingAge;
  if (suggestion.district_code) {
    applyLocationCode(suggestion.district_code, true);
  }
  addressSuggestions.value = [];
};

// 38. 回填詳情
const applyDetail = (detail: PropertyListingDetailResponse): void => {
  form.title = detail.title;
  form.titleEn = detail.property_sale?.title_en || '';
  form.summary = detail.summary;
  form.description = detail.description;
  form.descriptionEn =
    detail.property_sale?.description_en ||
    detail.serviced_apartment?.description_en ||
    '';
  applyLocationCode(detail.district_code);
  form.communityId = detail.community?.public_id || '';
  form.publisherIdentityType = detail.publisher_identity_type;
  form.businessStatus = detail.business_status === 'sold' ? 'sold' : 'available';
  form.allowPhone = detail.contact_summary.show_phone;
  form.allowWhatsapp = detail.contact_summary.show_whatsapp;
  form.allowChat = detail.contact_summary.show_chat;
  form.contactAttributes = {
    phone_country_code: '+852',
    phone_whatsapp_enabled: '',
    hide_phone_allow_inquiry: '',
    default_avatar_gender: 'male',
    agency_company_profile: '',
    agency_contact_profile: '',
    ...(detail.contact_summary.contact_attributes ?? {}),
  };

  if (detail.property_sale) {
    form.propertyNo = detail.property_sale.property_no || '';
    form.transactionType = detail.property_sale.transaction_type;
    form.locationScope = detail.property_sale.location_scope === 'overseas' ? 'overseas' : 'local';
    form.listingCategory = detail.property_sale.listing_category || 'standard';
    form.multiUnitProject = Boolean(detail.property_sale.multi_unit_project);
    form.propertyType = normalizeSalePropertyType(detail.property_sale.property_type);
    form.propertyAttributes = {
      account_or_package: 'personal_account',
      ...(detail.property_sale.property_attributes ?? {}),
    };
    form.rentalType = detail.property_sale.rental_type || '';
    form.renovationType = detail.property_sale.renovation_type || '';
    form.agencyCompanyName = detail.property_sale.agency_company_name || '';
    if (!form.contactAttributes.agency_company_profile) {
      form.contactAttributes.agency_company_profile = form.agencyCompanyName;
    }
    form.estateName = detail.property_sale.estate_name;
    form.addressText = detail.property_sale.address_text;
    form.addressTextEn = detail.property_sale.address_text_en || '';
    form.blockName = detail.property_sale.block_name || '';
    form.unitName = detail.property_sale.unit_name || '';
    form.showUnit = Boolean(detail.property_sale.show_unit);
    form.latitude = detail.property_sale.latitude ?? 0;
    form.longitude = detail.property_sale.longitude ?? 0;
    form.askingPriceHKD = detail.property_sale.asking_price_hkd > 0
      ? Number((detail.property_sale.asking_price_hkd / 10000).toFixed(2))
      : 0;
    form.monthlyRentHKD = detail.property_sale.monthly_rent_hkd ?? 0;
    form.priceReferenceOnly = Boolean(detail.property_sale.price_reference_only);
    form.priceNegotiable = Boolean(detail.property_sale.price_negotiable);
    form.annualPrepayDiscount = Boolean(detail.property_sale.annual_prepay_discount);
    form.annualPrepayOption = detail.property_sale.annual_prepay_option || '95_off';
    form.leaseStartDate = detail.property_sale.lease_start_date || '';
    form.rentIncluded = detail.property_sale.rent_included || '';
    form.rentIncludedItems = splitTextList(detail.property_sale.property_attributes?.rent_included_items || '');
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
    form.projectAttributes = {
      account_or_package: 'personal_account',
      address_street: '',
      address_doorplate: '',
      facility_custom_text: '',
      service_custom_text: '',
      ...(detail.serviced_apartment.project_attributes ?? {}),
    };
    form.addressText = detail.serviced_apartment.address_text;
    form.addressTextEn = detail.serviced_apartment.address_text_en || '';
    if (!form.projectAttributes.address_street) {
      form.projectAttributes.address_street = form.addressText;
    }
    form.websiteURL = detail.serviced_apartment.website_url || '';
    form.serviceWhatsApp = detail.serviced_apartment.whatsapp || '';
    form.fax = detail.serviced_apartment.fax || '';
    form.serviceIntro = detail.serviced_apartment.service_intro || '';
    form.serviceIntroEn = detail.serviced_apartment.service_intro_en || '';
    form.benefitsText = detail.serviced_apartment.benefits_text || '';
    form.benefitsTextEn = detail.serviced_apartment.benefits_text_en || '';
    form.extraChargesText = detail.serviced_apartment.extra_charges_text || '';
    form.extraChargesTextEn = detail.serviced_apartment.extra_charges_text_en || '';
    form.lowestMonthlyRentHKD = detail.serviced_apartment.lowest_monthly_rent_hkd;
    form.highestMonthlyRentHKD = detail.serviced_apartment.highest_monthly_rent_hkd ?? 0;
    form.lowestDailyRentHKD = detail.serviced_apartment.lowest_daily_rent_hkd ?? 0;
    form.priceReferenceOnlyServiced = Boolean(detail.serviced_apartment.price_reference_only);
    form.priceNegotiableServiced = Boolean(detail.serviced_apartment.price_negotiable);
    form.minUsableAreaSqft = detail.serviced_apartment.min_usable_area_sqft ?? 0;
    form.maxUsableAreaSqft = detail.serviced_apartment.max_usable_area_sqft ?? 0;
    form.minLeaseMonths = detail.serviced_apartment.min_lease_months;
    form.minStayValue = detail.serviced_apartment.min_stay_value || detail.serviced_apartment.min_lease_months;
    form.minStayUnit = normalizeServicedStayUnit(detail.serviced_apartment.min_stay_unit);
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
          usable_area_min_sqft: item.usable_area_min_sqft || item.usable_area_sqft,
          usable_area_max_sqft: item.usable_area_max_sqft || item.usable_area_min_sqft || item.usable_area_sqft,
          monthly_rent_min_hkd: item.monthly_rent_min_hkd || item.monthly_rent_hkd,
          monthly_rent_max_hkd: item.monthly_rent_max_hkd || item.monthly_rent_min_hkd || item.monthly_rent_hkd,
          daily_rent_max_hkd: item.daily_rent_max_hkd || item.daily_rent_min_hkd || 0,
          rent_unit: normalizeServicedStayUnit(item.rent_unit || item.min_stay_unit),
          min_stay_value: item.min_stay_value || item.min_lease_months || 1,
          min_stay_unit: normalizeServicedStayUnit(item.min_stay_unit),
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

// 39. 載入編輯資料
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

// 40. 儲存草稿
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

// 41. 儲存並發布
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

// 42. 儲存並返回列表
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
  await loadDetail();
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
      <div
        v-if="!props.hideProgress"
        class="property-editor-progress"
        aria-label="發布步驟"
      >
        <button
          v-for="(step, stepIndex) in editorSteps"
          :key="step.key"
          type="button"
          class="property-editor-progress__step"
          :class="{
            'property-editor-progress__step--active': activeEditorStep === step.key,
            'property-editor-progress__step--done': stepIndex < activeEditorStepIndex,
          }"
          :aria-current="activeEditorStep === step.key ? 'step' : undefined"
          :aria-label="step.label"
          @click="selectEditorStep(step.key)"
        >
          <span class="property-editor-progress__number">{{ stepIndex + 1 }}</span>
          <span class="property-editor-progress__label">{{ step.label }}</span>
        </button>
      </div>

      <form
        class="property-editor-form"
        @submit.prevent="saveAndReturn"
      >
        <section
          v-if="activeEditorStep === 'category'"
          class="property-editor-panel"
        >
          <h2>{{ isSale ? t('property.editor.stepCategory') : t('property.editor.servicedStepCategory') }}</h2>
          <div class="property-editor-grid">
            <label class="property-input">
              <span>{{ isSale ? t('property.editor.publisherIdentityField') : t('property.editor.servicedPublisherIdentityField') }}</span>
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
            <label
              v-if="isSale"
              class="property-input"
            >
              <span>{{ t('property.sale.typeLabel') }}</span>
              <select v-model="form.propertyType">
                <option
                  v-for="typeOption in propertyListingTypeOptions"
                  :key="typeOption.value"
                  :value="typeOption.value"
                >
                  {{ getPropertyOptionLabel(typeOption, preferenceStore.locale) }}
                </option>
              </select>
            </label>
            <label
              v-if="isSale"
              class="property-input"
            >
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
            <!-- 服務式住宅暫不開放：本地 / 海外、放盤類別、多於一伙或發展商項目。 -->
            <label
              v-if="isSale"
              class="property-input"
            >
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
            <label
              v-if="isSale"
              class="property-input property-input--wide"
            >
              <span>{{ t('property.editor.accountPackageField') }}</span>
              <select v-model="form.propertyAttributes.account_or_package">
                <option
                  v-for="option in propertyAccountPackageOptions"
                  :key="option.value"
                  :value="option.value"
                >
                  {{ getPropertyOptionLabel(option, preferenceStore.locale) }}
                </option>
              </select>
            </label>
            <label
              v-else
              class="property-input property-input--wide"
            >
              <span>{{ t('property.editor.accountPackageField') }}</span>
              <select v-model="form.projectAttributes.account_or_package">
                <option
                  v-for="option in propertyAccountPackageOptions"
                  :key="option.value"
                  :value="option.value"
                >
                  {{ getPropertyOptionLabel(option, preferenceStore.locale) }}
                </option>
              </select>
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

        <section
          v-if="activeEditorStep === 'ad'"
          class="property-editor-panel"
        >
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

        <section
          v-if="activeEditorStep === 'details' && !isSale"
          class="property-editor-panel"
        >
          <h2>{{ t('property.editor.servicedModuleBasic') }}</h2>
          <div class="property-editor-grid">
            <label class="property-input property-input--wide property-input--required">
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
            <div class="property-location-selectors property-input--wide">
              <label class="property-input property-input--required">
                <span>{{ t('property.editor.locationAreaField') }}</span>
                <select v-model="form.locationAreaCode">
                  <option value="">{{ t('property.editor.locationAreaPlaceholder') }}</option>
                  <option
                    v-for="area in propertyLocationAreas"
                    :key="area.value"
                    :value="area.value"
                  >
                    {{ getPropertyOptionLabel(area, preferenceStore.locale) }}
                  </option>
                </select>
              </label>
              <label class="property-input property-input--required">
                <span>{{ t('property.editor.locationDistrictField') }}</span>
                <select
                  v-model="form.locationDistrictCode"
                  :disabled="!form.locationAreaCode"
                >
                  <option value="">{{ t('property.editor.locationDistrictPlaceholder') }}</option>
                  <option
                    v-for="district in locationDistrictOptions"
                    :key="district.value"
                    :value="district.value"
                  >
                    {{ getPropertyOptionLabel(district, preferenceStore.locale) }}
                  </option>
                </select>
              </label>
              <label class="property-input property-input--required">
                <span>{{ t('property.editor.locationSubdistrictField') }}</span>
                <select
                  v-model="form.districtCode"
                  :disabled="!form.locationDistrictCode"
                >
                  <option value="">{{ t('property.editor.locationSubdistrictPlaceholder') }}</option>
                  <option
                    v-for="subdistrict in locationSubdistrictOptions"
                    :key="subdistrict.value"
                    :value="subdistrict.value"
                  >
                    {{ getPropertyOptionLabel(subdistrict, preferenceStore.locale) }}
                  </option>
                </select>
              </label>
            </div>
            <label class="property-input">
              <span>{{ t('property.editor.addressStreetField') }}</span>
              <input v-model="form.projectAttributes.address_street" />
            </label>
            <label class="property-input">
              <span>{{ t('property.editor.addressDoorplateField') }}</span>
              <input v-model="form.projectAttributes.address_doorplate" />
            </label>
            <label class="property-input property-input--wide property-input--required">
              <span>{{ t('property.editor.summaryField') }}</span>
              <input v-model="form.summary" />
            </label>
            <label class="property-input property-input--wide">
              <span>{{ t('property.editor.summaryEnField') }}</span>
              <input v-model="form.projectAttributes.summary_en" />
            </label>
          </div>
        </section>

        <section
          v-if="activeEditorStep === 'details' && isResidentialSaleEditor"
          class="property-editor-panel"
        >
          <h2>{{ t('property.editor.residentialModuleA') }}</h2>
          <div class="property-editor-grid">
            <label
              v-if="isAgentPublisher"
              :class="['property-input', { 'property-input--required': propertyNoFieldRequired }]"
            >
              <span>{{ t('property.editor.propertyNoField') }}</span>
              <input v-model="form.propertyNo" />
            </label>
            <label
              v-if="isAgentPublisher"
              class="property-input"
            >
              <span>{{ t('property.editor.propertyReferenceNoField') }}</span>
              <input v-model="form.propertyAttributes.prn" />
            </label>
            <label
              :class="['property-input property-input--wide', { 'property-input--required': estateNameRequired }]"
            >
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
            <label :class="['property-input', { 'property-input--required': grossAreaRequired }]">
              <span>{{ t('property.editor.grossAreaField') }}</span>
              <input
                :value="formatOptionalNumberInput(form.grossAreaSqft)"
                type="number"
                min="0"
                @input="form.grossAreaSqft = readOptionalNumberInput($event)"
              />
            </label>
            <div :class="['property-input property-input--with-tools', { 'property-input--required': usableAreaRequired }]">
              <div class="property-input-label-row">
                <span>{{ t('property.editor.usableAreaField') }}</span>
                <label class="property-compact-checkbox">
                  <input
                    v-model="form.propertyAttributes.area_unverified"
                    type="checkbox"
                    true-value="yes"
                    false-value=""
                  />
                  <span>{{ t('property.editor.areaUnverifiedField') }}</span>
                </label>
              </div>
              <input
                :value="formatOptionalNumberInput(form.usableAreaSqft)"
                type="number"
                min="0"
                @input="form.usableAreaSqft = readOptionalNumberInput($event)"
              />
            </div>
            <div class="property-input property-input--with-tools">
              <div class="property-input-label-row">
                <span>{{ t('property.editor.buildingAgeField') }}</span>
                <label class="property-compact-checkbox">
                  <input
                    v-model="form.propertyAttributes.new_completion"
                    type="checkbox"
                    true-value="yes"
                    false-value=""
                  />
                  <span>{{ t('property.editor.newCompletionField') }}</span>
                </label>
              </div>
              <input v-model="form.buildingAge" />
            </div>
            <label class="property-input">
              <span>{{ t('property.editor.blockNameField') }}</span>
              <input v-model="form.blockName" />
            </label>
            <div class="property-input property-input--with-tools">
              <div class="property-input-label-row">
                <span>{{ t('property.editor.unitNameField') }}</span>
                <label class="property-compact-checkbox">
                  <input
                    v-model="form.showUnit"
                    type="checkbox"
                  />
                  <span>{{ t('property.editor.showUnit') }}</span>
                </label>
              </div>
              <input v-model="form.unitName" />
            </div>
            <div class="property-location-selectors property-input--wide">
              <label class="property-input property-input--required">
                <span>{{ t('property.editor.locationAreaField') }}</span>
                <select v-model="form.locationAreaCode">
                  <option value="">{{ t('property.editor.locationAreaPlaceholder') }}</option>
                  <option
                    v-for="area in propertyLocationAreas"
                    :key="area.value"
                    :value="area.value"
                  >
                    {{ getPropertyOptionLabel(area, preferenceStore.locale) }}
                  </option>
                </select>
              </label>
              <label class="property-input property-input--required">
                <span>{{ t('property.editor.locationDistrictField') }}</span>
                <select
                  v-model="form.locationDistrictCode"
                  :disabled="!form.locationAreaCode"
                >
                  <option value="">{{ t('property.editor.locationDistrictPlaceholder') }}</option>
                  <option
                    v-for="district in locationDistrictOptions"
                    :key="district.value"
                    :value="district.value"
                  >
                    {{ getPropertyOptionLabel(district, preferenceStore.locale) }}
                  </option>
                </select>
              </label>
              <label class="property-input property-input--required">
                <span>{{ t('property.editor.locationSubdistrictField') }}</span>
                <select
                  v-model="form.districtCode"
                  :disabled="!form.locationDistrictCode"
                >
                  <option value="">{{ t('property.editor.locationSubdistrictPlaceholder') }}</option>
                  <option
                    v-for="subdistrict in locationSubdistrictOptions"
                    :key="subdistrict.value"
                    :value="subdistrict.value"
                  >
                    {{ getPropertyOptionLabel(subdistrict, preferenceStore.locale) }}
                  </option>
                </select>
              </label>
            </div>
            <label
              v-for="field in residentialAddressTextFields"
              :key="field.key"
              :class="['property-input property-input--required', { 'property-input--wide': field.wide }]"
            >
              <span>{{ field.label }}</span>
              <input v-model="form[field.key]" />
            </label>
          </div>
        </section>

        <section
          v-if="activeEditorStep === 'details' && isResidentialSaleEditor"
          class="property-editor-panel"
        >
          <h2>{{ t('property.editor.residentialModuleB') }}</h2>
          <div class="property-checkbox-row property-checkbox-row--block property-checkbox-row--required">
            <strong>
              <span>{{ t('property.editor.residentialTagRequired') }}</span>
              <span class="property-tag-group-title__meta">{{ t('property.editor.tagRequiredMultiSelect') }}</span>
            </strong>
            <label
              v-for="tag in saleCategoryTagOptions"
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
          <div
            v-for="group in residentialFeatureTagGroups"
            :key="group.title"
            class="property-checkbox-row property-checkbox-row--block"
          >
            <strong>
              <span>{{ group.title }}</span>
              <span class="property-tag-group-title__meta">{{ t('property.editor.tagMultiSelect') }}</span>
            </strong>
            <label
              v-for="tag in group.options"
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
          v-if="activeEditorStep === 'details' && isResidentialSaleEditor"
          class="property-editor-panel"
        >
          <h2>{{ t('property.editor.residentialModuleC') }}</h2>
          <div class="property-editor-grid">
            <label
              v-if="form.transactionType === 'sale'"
              :class="['property-input', { 'property-input--required': salePriceFieldRequired }]"
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
              :class="['property-input', { 'property-input--required': salePriceFieldRequired }]"
            >
              <span>{{ t('property.editor.monthlyRentField') }}</span>
              <input
                v-model.number="form.monthlyRentHKD"
                type="number"
                min="0"
              />
            </label>
            <label
              v-if="form.transactionType === 'rent'"
              class="property-inline-checkbox"
            >
              <input
                v-model="form.annualPrepayDiscount"
                type="checkbox"
              />
              {{ t('property.editor.annualPrepayDiscount') }}
            </label>
            <label
              v-if="form.transactionType === 'rent'"
              class="property-input"
            >
              <span>{{ t('property.editor.leaseStartDateField') }}</span>
              <input
                v-model="form.leaseStartDate"
                type="date"
              />
            </label>
            <label
              v-if="form.transactionType === 'rent'"
              class="property-input property-input--wide"
            >
              <span>{{ t('property.editor.rentIncludedField') }}</span>
              <span class="property-checkbox-row property-checkbox-row--inline">
                <label
                  v-for="option in saleRentIncludedOptions"
                  :key="option.value"
                >
                  <input
                    type="checkbox"
                    :checked="form.rentIncludedItems.includes(option.value)"
                    @change="toggleTag(form.rentIncludedItems, option.value)"
                  />
                  {{ getPropertyOptionLabel(option, preferenceStore.locale) }}
                </label>
              </span>
            </label>
          </div>
          <div
            v-if="form.transactionType === 'sale'"
            class="property-checkbox-row"
          >
            <label>
              <input
                v-model="form.priceReferenceOnly"
                type="checkbox"
              />
              {{ t('property.editor.priceSuffixPlus') }}
            </label>
            <label>
              <input
                v-model="form.priceNegotiable"
                type="checkbox"
              />
              {{ t('property.editor.priceNegotiable') }}
            </label>
          </div>
        </section>

        <section
          v-if="activeEditorStep === 'details' && isResidentialSaleEditor"
          class="property-editor-panel"
        >
          <h2>{{ t('property.editor.residentialModuleD') }}</h2>
          <div class="property-editor-grid">
            <label class="property-input property-input--required">
              <span>{{ t('property.editor.titleField') }}</span>
              <input
                v-model="form.title"
                maxlength="40"
              />
            </label>
            <label class="property-input property-input--required">
              <span>{{ t('property.editor.titleEnField') }}</span>
              <input
                v-model="form.titleEn"
                maxlength="100"
              />
            </label>
            <label class="property-input property-input--wide property-input--required">
              <span>{{ t('property.editor.descriptionField') }}</span>
              <textarea
                v-model="form.description"
                maxlength="1000"
                rows="5"
              />
            </label>
            <label class="property-input property-input--wide property-input--required">
              <span>{{ t('property.editor.descriptionEnField') }}</span>
              <textarea
                v-model="form.descriptionEn"
                maxlength="2000"
                rows="4"
              />
            </label>
            <label class="property-input">
              <span>{{ t('property.editor.directionField') }}</span>
              <select v-model="form.direction">
                <option
                  v-for="option in propertyDirectionOptions"
                  :key="option.value"
                  :value="option.value"
                >
                  {{ getPropertyOptionLabel(option, preferenceStore.locale) }}
                </option>
              </select>
            </label>
            <label :class="['property-input', { 'property-input--required': floorFieldRequired }]">
              <span>{{ t('property.editor.floorField') }}</span>
              <select v-model="form.floorRaw">
                <option
                  v-for="option in propertyFloorDisplayOptions"
                  :key="option.value"
                  :value="option.value"
                >
                  {{ getPropertyOptionLabel(option, preferenceStore.locale) }}
                </option>
              </select>
            </label>
            <label class="property-input">
              <span>{{ t('property.editor.bedroomField') }}</span>
              <select v-model.number="form.bedroomCount">
                <option
                  v-for="option in propertyRoomCountOptions"
                  :key="option.value"
                  :value="Number(option.value)"
                >
                  {{ getPropertyOptionLabel(option, preferenceStore.locale) }}
                </option>
              </select>
            </label>
            <label class="property-input">
              <span>{{ t('property.editor.bathroomField') }}</span>
              <select v-model.number="form.bathroomCount">
                <option
                  v-for="option in propertyBathroomCountOptions"
                  :key="option.value"
                  :value="Number(option.value)"
                >
                  {{ getPropertyOptionLabel(option, preferenceStore.locale) }}
                </option>
              </select>
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
          </div>
          <div class="property-checkbox-row">
            <label>
              <input
                v-model="form.propertyAttributes.extra_bathroom_toilet"
                type="checkbox"
                true-value="yes"
                false-value=""
              />
              {{ t('property.editor.extraBathroomToiletField') }}
            </label>
          </div>
        </section>

        <section
          v-if="activeEditorStep === 'details' && isSale && !isResidentialSaleEditor"
          class="property-editor-panel"
        >
          <h2>{{ t('property.editor.saleModuleA') }}</h2>
          <div class="property-editor-grid">
            <label
              v-if="saleRequiresPropertyNo"
              :class="['property-input', { 'property-input--required': propertyNoFieldRequired }]"
            >
              <span>{{ t('property.editor.propertyNoField') }}</span>
              <input v-model="form.propertyNo" />
            </label>
            <label
              v-if="isAgentPublisher || form.propertyType === 'land'"
              class="property-input"
            >
              <span>{{ t('property.editor.propertyReferenceNoField') }}</span>
              <input v-model="form.propertyAttributes.prn" />
            </label>
            <label
              v-if="saleFieldProfile.attributeKeys.includes('lot_number')"
              class="property-input"
            >
              <span>{{ t('property.editor.lotNumberField') }}</span>
              <input v-model="form.propertyAttributes.lot_number" />
            </label>
            <label
              v-if="form.propertyType !== 'land'"
              class="property-input property-input--wide property-input--required"
            >
              <span>{{ t(saleFieldProfile.estateLabelKey) }}</span>
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
            <label
              v-if="saleFieldProfile.showGrossArea"
              :class="['property-input', { 'property-input--required': grossAreaRequired }]"
            >
              <span>{{ t('property.editor.grossAreaField') }}</span>
              <input
                :value="formatOptionalNumberInput(form.grossAreaSqft)"
                type="number"
                min="0"
                @input="form.grossAreaSqft = readOptionalNumberInput($event)"
              />
            </label>
            <div
              v-if="saleFieldProfile.showUsableArea"
              :class="['property-input property-input--with-tools', { 'property-input--required': usableAreaRequired }]"
            >
              <div class="property-input-label-row">
                <span>{{ t('property.editor.usableAreaField') }}</span>
                <label
                  v-if="saleFieldProfile.attributeKeys.includes('area_unverified')"
                  class="property-compact-checkbox"
                >
                  <input
                    v-model="form.propertyAttributes.area_unverified"
                    type="checkbox"
                    true-value="yes"
                    false-value=""
                  />
                  <span>{{ t('property.editor.areaUnverifiedField') }}</span>
                </label>
              </div>
              <input
                :value="formatOptionalNumberInput(form.usableAreaSqft)"
                type="number"
                min="0"
                @input="form.usableAreaSqft = readOptionalNumberInput($event)"
              />
            </div>
            <div
              v-if="saleFieldProfile.showBuildingDetails"
              class="property-input property-input--with-tools"
            >
              <div class="property-input-label-row">
                <span>{{ t('property.editor.buildingAgeField') }}</span>
                <label class="property-compact-checkbox">
                  <input
                    v-model="form.propertyAttributes.new_completion"
                    type="checkbox"
                    true-value="yes"
                    false-value=""
                  />
                  <span>{{ t('property.editor.newCompletionField') }}</span>
                </label>
              </div>
              <input v-model="form.buildingAge" />
            </div>
            <label
              v-if="saleFieldProfile.showUnitFields"
              class="property-input"
            >
              <span>{{ t('property.editor.blockNameField') }}</span>
              <input v-model="form.blockName" />
            </label>
            <label
              v-if="saleFieldProfile.showUnitFields"
              class="property-input property-input--with-tools"
            >
              <div class="property-input-label-row">
                <span>{{ t('property.editor.unitNameField') }}</span>
                <label class="property-compact-checkbox">
                  <input
                    v-model="form.showUnit"
                    type="checkbox"
                  />
                  <span>{{ t('property.editor.showUnit') }}</span>
                </label>
              </div>
              <input v-model="form.unitName" />
            </label>
            <label class="property-input property-input--wide property-input--required">
              <span>{{ t('property.editor.addressField') }}</span>
              <input v-model="form.addressText" />
            </label>
            <label class="property-input property-input--wide property-input--required">
              <span>{{ t('property.editor.addressEnField') }}</span>
              <input v-model="form.addressTextEn" />
            </label>
            <div class="property-location-selectors property-input--wide">
              <label class="property-input property-input--required">
                <span>{{ t('property.editor.locationAreaField') }}</span>
                <select v-model="form.locationAreaCode">
                  <option value="">{{ t('property.editor.locationAreaPlaceholder') }}</option>
                  <option
                    v-for="area in propertyLocationAreas"
                    :key="area.value"
                    :value="area.value"
                  >
                    {{ getPropertyOptionLabel(area, preferenceStore.locale) }}
                  </option>
                </select>
              </label>
              <label class="property-input property-input--required">
                <span>{{ t('property.editor.locationDistrictField') }}</span>
                <select
                  v-model="form.locationDistrictCode"
                  :disabled="!form.locationAreaCode"
                >
                  <option value="">{{ t('property.editor.locationDistrictPlaceholder') }}</option>
                  <option
                    v-for="district in locationDistrictOptions"
                    :key="district.value"
                    :value="district.value"
                  >
                    {{ getPropertyOptionLabel(district, preferenceStore.locale) }}
                  </option>
                </select>
              </label>
              <label class="property-input property-input--required">
                <span>{{ t('property.editor.locationSubdistrictField') }}</span>
                <select
                  v-model="form.districtCode"
                  :disabled="!form.locationDistrictCode"
                >
                  <option value="">{{ t('property.editor.locationSubdistrictPlaceholder') }}</option>
                  <option
                    v-for="subdistrict in locationSubdistrictOptions"
                    :key="subdistrict.value"
                    :value="subdistrict.value"
                  >
                    {{ getPropertyOptionLabel(subdistrict, preferenceStore.locale) }}
                  </option>
                </select>
              </label>
            </div>
          </div>
        </section>

        <section
          v-if="activeEditorStep === 'details' && isSale && !isResidentialSaleEditor"
          class="property-editor-panel"
        >
          <h2>{{ t('property.editor.saleModuleB') }}</h2>
          <div
            v-if="saleCategoryTagOptions.length > 0"
            :class="['property-checkbox-row property-checkbox-row--block', { 'property-checkbox-row--required': saleCategoryRequired }]"
          >
            <strong>
              <span>{{ t(saleFieldProfile.categoryLabelKey) }}</span>
              <span class="property-tag-group-title__meta">
                {{ saleCategoryRequired ? t('property.editor.tagRequiredMultiSelect') : t('property.editor.tagMultiSelect') }}
              </span>
            </strong>
            <label
              v-for="tag in saleCategoryTagOptions"
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
          <div
            v-for="group in nonResidentialFeatureTagGroups"
            :key="group.title"
            class="property-checkbox-row property-checkbox-row--block"
          >
            <strong>
              <span>{{ group.title }}</span>
              <span class="property-tag-group-title__meta">{{ t('property.editor.tagMultiSelect') }}</span>
            </strong>
            <label
              v-for="tag in group.options"
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
          v-if="activeEditorStep === 'details' && isSale && !isResidentialSaleEditor"
          class="property-editor-panel"
        >
          <h2>{{ t('property.editor.saleModuleC') }}</h2>
          <div class="property-editor-grid">
            <label
              v-if="form.transactionType === 'sale'"
              :class="['property-input', { 'property-input--required': salePriceFieldRequired }]"
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
              :class="['property-input', { 'property-input--required': salePriceFieldRequired }]"
            >
              <span>{{ t('property.editor.monthlyRentField') }}</span>
              <input
                v-model.number="form.monthlyRentHKD"
                type="number"
                min="0"
              />
            </label>
            <label
              v-if="form.transactionType === 'rent'"
              class="property-inline-checkbox"
            >
              <input
                v-model="form.annualPrepayDiscount"
                type="checkbox"
              />
              {{ t('property.editor.annualPrepayDiscount') }}
            </label>
            <label
              v-if="form.transactionType === 'rent'"
              class="property-input"
            >
              <span>{{ t('property.editor.leaseStartDateField') }}</span>
              <input
                v-model="form.leaseStartDate"
                type="date"
              />
            </label>
            <label
              v-if="form.transactionType === 'rent'"
              class="property-input property-input--wide"
            >
              <span>{{ t('property.editor.rentIncludedField') }}</span>
              <span class="property-checkbox-row property-checkbox-row--inline">
                <label
                  v-for="option in saleRentIncludedOptions"
                  :key="option.value"
                >
                  <input
                    type="checkbox"
                    :checked="form.rentIncludedItems.includes(option.value)"
                    @change="toggleTag(form.rentIncludedItems, option.value)"
                  />
                  {{ getPropertyOptionLabel(option, preferenceStore.locale) }}
                </label>
              </span>
            </label>
          </div>
          <div
            v-if="form.transactionType === 'sale'"
            class="property-checkbox-row"
          >
            <label>
              <input
                v-model="form.priceReferenceOnly"
                type="checkbox"
              />
              {{ t('property.editor.priceSuffixPlus') }}
            </label>
            <label>
              <input
                v-model="form.priceNegotiable"
                type="checkbox"
              />
              {{ t('property.editor.priceNegotiable') }}
            </label>
          </div>
        </section>

        <section
          v-if="activeEditorStep === 'details' && isSale && !isResidentialSaleEditor"
          class="property-editor-panel"
        >
          <h2>{{ t('property.editor.saleModuleD') }}</h2>
          <div class="property-editor-grid">
            <label class="property-input property-input--required">
              <span>{{ t('property.editor.titleField') }}</span>
              <input
                v-model="form.title"
                maxlength="40"
              />
            </label>
            <label class="property-input property-input--required">
              <span>{{ t('property.editor.titleEnField') }}</span>
              <input
                v-model="form.titleEn"
                maxlength="100"
              />
            </label>
            <label class="property-input property-input--wide property-input--required">
              <span>{{ t('property.editor.descriptionField') }}</span>
              <textarea
                v-model="form.description"
                maxlength="1000"
                rows="5"
              />
            </label>
            <label class="property-input property-input--wide property-input--required">
              <span>{{ t('property.editor.descriptionEnField') }}</span>
              <textarea
                v-model="form.descriptionEn"
                maxlength="2000"
                rows="4"
              />
            </label>
            <label
              v-if="saleFieldProfile.showDirection"
              class="property-input"
            >
              <span>{{ t('property.editor.directionField') }}</span>
              <select v-model="form.direction">
                <option
                  v-for="option in propertyDirectionOptions"
                  :key="option.value"
                  :value="option.value"
                >
                  {{ getPropertyOptionLabel(option, preferenceStore.locale) }}
                </option>
              </select>
            </label>
            <label
              v-if="saleFieldProfile.showFloor"
              :class="['property-input', { 'property-input--required': floorFieldRequired }]"
            >
              <span>{{ t('property.editor.floorField') }}</span>
              <select v-model="form.floorRaw">
                <option
                  v-for="option in propertyFloorDisplayOptions"
                  :key="option.value"
                  :value="option.value"
                >
                  {{ getPropertyOptionLabel(option, preferenceStore.locale) }}
                </option>
              </select>
            </label>
            <label
              v-if="saleFieldProfile.showRooms"
              class="property-input"
            >
              <span>{{ t('property.editor.bedroomField') }}</span>
              <select v-model.number="form.bedroomCount">
                <option
                  v-for="option in propertyRoomCountOptions"
                  :key="option.value"
                  :value="Number(option.value)"
                >
                  {{ getPropertyOptionLabel(option, preferenceStore.locale) }}
                </option>
              </select>
            </label>
            <label
              v-if="saleFieldProfile.showRooms"
              class="property-input"
            >
              <span>{{ saleUsesToiletLabel ? t('property.editor.toiletField') : t('property.editor.bathroomField') }}</span>
              <select v-model.number="form.bathroomCount">
                <option
                  v-for="option in saleBathroomCountOptions"
                  :key="option.value"
                  :value="Number(option.value)"
                >
                  {{ getPropertyOptionLabel(option, preferenceStore.locale) }}
                </option>
              </select>
            </label>
            <label
              v-if="saleFieldProfile.showRooms"
              class="property-inline-checkbox"
            >
              <input
                v-model="form.propertyAttributes.extra_bathroom_toilet"
                type="checkbox"
                true-value="yes"
                false-value=""
              />
              {{ t('property.editor.extraBathroomToiletField') }}
            </label>
            <label
              v-if="saleFieldProfile.showManagementCompany"
              class="property-input"
            >
              <span>{{ t('property.editor.managementFeeField') }}</span>
              <input
                :value="formatOptionalNumberInput(form.managementFeeHKD)"
                type="number"
                min="0"
                @input="form.managementFeeHKD = readOptionalNumberInput($event)"
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
          </div>
        </section>

        <section
          v-if="activeEditorStep === 'details' && !isSale"
          class="property-editor-panel"
        >
          <h2>{{ t('property.editor.servicedModuleRentArea') }}</h2>
          <div class="property-editor-grid">
            <label class="property-input property-input--required">
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
            <label :class="['property-input', { 'property-input--required': servicedPriceFieldRequired }]">
              <span>{{ t('property.editor.roomRentMinField') }}</span>
              <input
                v-model.number="form.lowestMonthlyRentHKD"
                type="number"
                min="0"
              />
            </label>
            <label class="property-input">
              <span>{{ t('property.editor.roomRentMaxField') }}</span>
              <input
                v-model.number="form.highestMonthlyRentHKD"
                type="number"
                min="0"
              />
            </label>
            <label class="property-input">
              <span>{{ t('property.editor.minUsableAreaField') }}</span>
              <input
                :value="formatOptionalNumberInput(form.minUsableAreaSqft)"
                type="number"
                min="0"
                @input="form.minUsableAreaSqft = readOptionalNumberInput($event)"
              />
            </label>
            <label class="property-input">
              <span>{{ t('property.editor.maxUsableAreaField') }}</span>
              <input
                :value="formatOptionalNumberInput(form.maxUsableAreaSqft)"
                type="number"
                min="0"
                @input="form.maxUsableAreaSqft = readOptionalNumberInput($event)"
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
        </section>

        <section
          v-if="activeEditorStep === 'details' && !isSale"
          class="property-editor-panel"
        >
          <h2>{{ t('property.editor.servicedModuleContact') }}</h2>
          <div class="property-editor-grid">
            <label class="property-input">
              <span>{{ t('property.editor.websiteField') }}</span>
              <input v-model="form.websiteURL" />
            </label>
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
              <span>{{ t('property.editor.servicedPhoneField') }}</span>
              <input v-model="form.phone" />
            </label>
            <label class="property-input">
              <span>{{ t('property.editor.serviceWhatsappField') }}</span>
              <input v-model="form.serviceWhatsApp" />
            </label>
            <label class="property-input">
              <span>{{ t('property.editor.wechatField') }}</span>
              <input v-model="form.wechat" />
            </label>
            <label class="property-input">
              <span>{{ t('property.editor.emailField') }}</span>
              <input v-model="form.email" />
            </label>
            <label class="property-input">
              <span>{{ t('property.editor.faxField') }}</span>
              <input v-model="form.fax" />
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
              {{ t('property.editor.allowWhatsappWechat') }}
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

        <section
          v-if="activeEditorStep === 'details' && !isSale"
          class="property-editor-panel"
        >
          <h2>{{ t('property.editor.servicedModuleRooms') }}</h2>
          <section class="property-room-types">
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
                <label class="property-input property-input--required">
                  <span>{{ t('property.editor.roomTypeName') }}</span>
                  <input v-model="room.name" />
                </label>
                <label class="property-input">
                  <span>{{ t('property.editor.roomTypeNameEn') }}</span>
                  <input v-model="room.name_en" />
                </label>
                <label class="property-input">
                  <span>{{ t('property.editor.roomCategoryField') }}</span>
                  <select v-model="room.room_category">
                    <option value="">
                      {{ t('property.editor.notSpecified') }}
                    </option>
                    <option
                      v-for="option in servicedRoomCategoryOptions"
                      :key="option.value"
                      :value="option.value"
                    >
                      {{ getPropertyOptionLabel(option, preferenceStore.locale) }}
                    </option>
                  </select>
                </label>
                <label class="property-input property-input--required">
                  <span>{{ t('property.editor.roomAreaMinField') }}</span>
                  <input
                    :value="formatOptionalNumberInput(room.usable_area_min_sqft || 0)"
                    type="number"
                    min="0"
                    @input="room.usable_area_min_sqft = readOptionalNumberInput($event)"
                  />
                </label>
                <label class="property-input">
                  <span>{{ t('property.editor.roomAreaMaxField') }}</span>
                  <input
                    :value="formatOptionalNumberInput(room.usable_area_max_sqft || 0)"
                    type="number"
                    min="0"
                    @input="room.usable_area_max_sqft = readOptionalNumberInput($event)"
                  />
                </label>
                <label class="property-input">
                  <span>{{ t('property.editor.roomTypeRent') }}</span>
                  <input
                    :value="formatOptionalNumberInput(room.monthly_rent_min_hkd || 0)"
                    type="number"
                    min="0"
                    @input="room.monthly_rent_min_hkd = readOptionalNumberInput($event)"
                  />
                </label>
                <label class="property-input">
                  <span>{{ t('property.editor.roomRentUnitField') }}</span>
                  <select v-model="room.rent_unit">
                    <option
                      v-for="unit in servicedRoomRentUnitOptions"
                      :key="unit.value"
                      :value="unit.value"
                    >
                      {{ getPropertyOptionLabel(unit, preferenceStore.locale) }}
                    </option>
                  </select>
                </label>
                <label class="property-inline-checkbox">
                  <input
                    v-model="room.rent_suffix_plus"
                    type="checkbox"
                  />
                  {{ t('property.editor.roomRentSuffixPlusField') }}
                </label>
                <label class="property-input">
                  <span>{{ t('property.editor.roomMonthlyRentMaxField') }}</span>
                  <input
                    :value="formatOptionalNumberInput(room.monthly_rent_max_hkd || 0)"
                    type="number"
                    min="0"
                    @input="room.monthly_rent_max_hkd = readOptionalNumberInput($event)"
                  />
                </label>
                <label class="property-input property-input--required">
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
                <div class="property-checkbox-row property-checkbox-row--block property-input--wide">
                  <strong>{{ t('property.editor.roomIncludedFacilitiesField') }}</strong>
                  <label
                    v-for="option in propertyApplianceTagOptions"
                    :key="option.value"
                  >
                    <input
                      type="checkbox"
                      :checked="room.included_fee_items?.includes(option.value)"
                      @change="toggleRoomIncludedFeeItem(room, option.value)"
                    />
                    {{ getPropertyOptionLabel(option, preferenceStore.locale) }}
                  </label>
                </div>
                <label class="property-input property-input--wide">
                  <span>{{ t('property.editor.includedFeeItemsField') }}</span>
                  <input
                    :value="room.included_fee_items?.join('、') || ''"
                    @input="updateRoomIncludedFeeItems(room, $event)"
                  />
                </label>
                <label class="property-input">
                  <span>{{ t('property.editor.roomImageField') }}</span>
                  <select v-model="room.image_media_asset_id">
                    <option value="">
                      {{ t('property.editor.notSpecified') }}
                    </option>
                    <option
                      v-for="(image, imageIndex) in uploadedImageOptions"
                      :key="image.mediaAssetId"
                      :value="image.mediaAssetId"
                    >
                      {{ t('property.editor.media') }} {{ imageIndex + 1 }}
                    </option>
                  </select>
                </label>
                <label class="property-input">
                  <span>{{ t('property.editor.roomPageUrlField') }}</span>
                  <input v-model="room.page_url" />
                </label>
              </div>
            </article>
          </section>
        </section>

        <section
          v-if="activeEditorStep === 'details' && !isSale"
          class="property-editor-panel"
        >
          <h2>{{ t('property.editor.servicedModuleServices') }}</h2>
          <div class="property-editor-grid">
            <label class="property-input property-input--wide">
              <span>{{ t('property.editor.serviceIntroField') }}</span>
              <textarea
                v-model="form.serviceIntro"
                rows="3"
              />
            </label>
            <label class="property-input property-input--wide">
              <span>{{ t('property.editor.serviceIntroEnField') }}</span>
              <textarea
                v-model="form.serviceIntroEn"
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
              <span>{{ t('property.editor.benefitsEnField') }}</span>
              <textarea
                v-model="form.benefitsTextEn"
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
            <label class="property-input property-input--wide">
              <span>{{ t('property.editor.extraChargesEnField') }}</span>
              <textarea
                v-model="form.extraChargesTextEn"
                rows="3"
              />
            </label>
          </div>
          <div class="property-checkbox-row property-checkbox-row--block">
            <strong>
              <span>{{ t('property.editor.publicFacilitiesField') }}</span>
              <span class="property-tag-group-title__meta">{{ t('property.editor.tagMultiSelect') }}</span>
            </strong>
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
          <label class="property-input property-input--wide">
            <span>{{ t('property.editor.facilityCustomField') }}</span>
            <input v-model="form.projectAttributes.facility_custom_text" />
          </label>
          <div class="property-checkbox-row property-checkbox-row--block">
            <strong>
              <span>{{ t('property.editor.roomServicesField') }}</span>
              <span class="property-tag-group-title__meta">{{ t('property.editor.tagMultiSelect') }}</span>
            </strong>
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
          <label class="property-input property-input--wide">
            <span>{{ t('property.editor.serviceCustomField') }}</span>
            <input v-model="form.projectAttributes.service_custom_text" />
          </label>
        </section>

        <section
          v-if="activeEditorStep === 'contact'"
          class="property-editor-panel"
        >
          <h2>{{ isSale ? t('property.editor.residentialModuleE') : t('property.editor.contact') }}</h2>
          <div class="property-editor-grid">
            <label
              v-if="isServicedPublisher"
              class="property-input"
            >
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
            <label
              v-if="isAgentPublisher"
              :class="['property-input', { 'property-input--required': isAgentContactRequired }]"
            >
              <span>{{ t('property.editor.agencyCompanyProfileField') }}</span>
              <input v-model="form.contactAttributes.agency_company_profile" />
            </label>
            <label
              v-if="isAgentPublisher"
              :class="['property-input', { 'property-input--required': isAgentContactRequired }]"
            >
              <span>{{ t('property.editor.agencyContactProfileField') }}</span>
              <input v-model="form.contactAttributes.agency_contact_profile" />
            </label>
            <label
              v-if="isSaleOwnerPublisher"
              :class="['property-input', { 'property-input--required': isOwnerContactRequired }]"
            >
              <span>{{ t('property.editor.contactNameZhField') }}</span>
              <input v-model="form.contactNameZh" />
            </label>
            <label
              v-if="isSaleOwnerPublisher"
              :class="['property-input', { 'property-input--required': isOwnerContactRequired }]"
            >
              <span>{{ t('property.editor.contactNameEnField') }}</span>
              <input v-model="form.contactNameEn" />
            </label>
            <div
              v-if="isSaleOwnerPublisher || isServicedPublisher"
              :class="['property-input', { 'property-input--required': isOwnerContactRequired || isServicedContactRequired }]"
            >
              <span>{{ isSale ? t('property.editor.phoneField') : t('property.editor.servicedPhoneField') }}</span>
              <div class="property-phone-group">
                <select
                  v-if="isSaleOwnerPublisher"
                  v-model="form.contactAttributes.phone_country_code"
                  class="property-phone-group__code"
                >
                  <option
                    v-for="code in phoneCountryCodeOptions"
                    :key="code.value"
                    :value="code.value"
                  >
                    {{ code.label }}
                  </option>
                </select>
                <input
                  v-model="form.phone"
                  class="property-phone-group__number"
                />
              </div>
            </div>
            <label
              v-if="isSaleOwnerPublisher"
              class="property-input"
            >
              <span>{{ t('property.editor.phone2Field') }}</span>
              <input v-model="form.phone2" />
            </label>
            <div
              v-if="isSaleOwnerPublisher"
              class="property-input property-input--wide property-input--inline-tools"
            >
              <div class="property-compact-checkbox-row">
                <label>
                  <input
                    v-model="form.contactAttributes.phone_whatsapp_enabled"
                    type="checkbox"
                    true-value="yes"
                    false-value=""
                  />
                  {{ t('property.editor.phoneWhatsappField') }}
                </label>
              </div>
            </div>
            <label
              v-if="isSaleOwnerPublisher || isServicedPublisher"
              class="property-input"
            >
              <span>{{ t('property.editor.wechatField') }}</span>
              <input v-model="form.wechat" />
            </label>
            <label
              v-if="isSaleOwnerPublisher"
              class="property-inline-checkbox property-input--wide"
            >
              <input
                v-model="form.contactAttributes.hide_phone_allow_inquiry"
                type="checkbox"
                true-value="yes"
                false-value=""
              />
              {{ t('property.editor.hidePhoneAllowInquiryField') }}
            </label>
            <label
              v-if="isServicedPublisher"
              class="property-input"
            >
              <span>{{ t('property.editor.emailField') }}</span>
              <input v-model="form.email" />
            </label>
            <label
              v-if="isSaleOwnerPublisher"
              class="property-input"
            >
              <span>{{ t('property.editor.defaultAvatarField') }}</span>
              <select v-model="form.contactAttributes.default_avatar_gender">
                <option
                  v-for="option in propertyDefaultAvatarOptions"
                  :key="option.value"
                  :value="option.value"
                >
                  {{ getPropertyOptionLabel(option, preferenceStore.locale) }}
                </option>
              </select>
            </label>
          </div>
          <div
            v-if="isServicedPublisher"
            class="property-checkbox-row"
          >
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
              {{ isSale ? t('property.editor.allowWhatsapp') : t('property.editor.allowWhatsappWechat') }}
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

        <section
          v-if="shouldShowMediaSection"
          class="property-editor-panel"
        >
          <h2>{{ t('property.editor.media') }}</h2>
          <div class="property-media-header">
            <p class="property-editor-note property-media-header__note">
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
          </div>
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
              <span
                v-if="image.isCover"
                class="property-image-card__badge"
              >
                {{ t('property.editor.coverImage') }}
              </span>
              <div class="property-image-card__actions">
                <button
                  type="button"
                  :class="{ 'property-image-card__action--active': image.isCover }"
                  @click="selectCover(image.id)"
                >
                  {{ image.isCover ? t('property.editor.coverImage') : t('property.editor.setCover') }}
                </button>
                <button
                  type="button"
                  @click="removeImage(image.id)"
                >
                  {{ t('property.editor.removeImage') }}
                </button>
              </div>
            </article>
          </div>
        </section>
      </form>

      <aside class="property-editor-side">
        <section class="property-editor-panel">
          <div class="property-side-summary">
            <p class="property-side-kicker">Preview</p>
            <h2>{{ (isSale ? form.title : form.projectName) || pageTitle }}</h2>
            <p>{{ (isSale ? resolveSaleSummary() : form.summary) || t('property.editor.requiredFields') }}</p>
            <strong>{{ previewPrice }}</strong>
          </div>

          <div class="property-side-section">
            <span class="property-side-section__title">{{ isSale ? t('property.editor.stepCategory') : t('property.editor.servicedStepCategory') }}</span>
            <p class="property-side-selected-category">{{ previewListingCategoryText }}</p>
          </div>

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
          <div class="property-editor-step-actions">
            <button
              type="button"
              class="property-editor-action property-editor-action--secondary"
              :disabled="isFirstEditorStep || saving || publishing"
              @click="goPreviousEditorStep"
            >
              上一步
            </button>
            <button
              v-if="!isLastEditorStep"
              type="button"
              class="property-editor-action property-editor-action--primary"
              :disabled="saving || publishing"
              @click="goNextEditorStep"
            >
              下一步
            </button>
          </div>
          <button
            v-if="isLastEditorStep"
            type="button"
            class="property-editor-action property-editor-action--secondary"
            :disabled="saving || publishing"
            @click="saveAndReturn"
          >
            {{ saving ? t('common.status.loading') : t('property.editor.saveDraft') }}
          </button>
          <button
            v-if="isLastEditorStep"
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
  letter-spacing: 0;
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
  gap: 16px;
}

.property-editor-progress {
  position: relative;
  display: grid;
  grid-column: 1 / -1;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  align-items: start;
  padding: 4px 0 10px;
}

.property-editor-progress::before {
  position: absolute;
  top: 19px;
  right: 20px;
  left: 20px;
  height: 1px;
  background: rgb(var(--color-border));
  content: "";
}

.property-editor-progress__step {
  position: relative;
  z-index: 1;
  display: grid;
  min-width: 0;
  justify-items: center;
  gap: 7px;
  border: 0;
  background: transparent;
  color: rgb(var(--color-text-muted));
  cursor: pointer;
}

.property-editor-progress__number {
  display: inline-flex;
  width: 30px;
  height: 30px;
  align-items: center;
  justify-content: center;
  border: 1px solid rgb(var(--color-border));
  border-radius: 999px;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text-muted));
  font-size: 12px;
  font-weight: 800;
  line-height: 1;
  transition:
    border-color 0.2s ease,
    background 0.2s ease,
    box-shadow 0.2s ease,
    color 0.2s ease;
}

.property-editor-progress__label {
  overflow: hidden;
  max-width: min(8rem, 100%);
  color: currentColor;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0;
  line-height: 1.3;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.property-editor-progress__step:hover .property-editor-progress__number,
.property-editor-progress__step--active .property-editor-progress__number {
  border-color: rgb(var(--color-primary));
  color: rgb(var(--color-primary));
}

.property-editor-progress__step--active .property-editor-progress__number,
.property-editor-progress__step--done .property-editor-progress__number {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary));
  color: rgb(var(--color-primary-contrast));
}

.property-editor-progress__step--active .property-editor-progress__number {
  box-shadow: 0 0 0 4px rgb(var(--color-primary) / 0.12);
}

.property-editor-progress__step--active .property-editor-progress__label,
.property-editor-progress__step--done .property-editor-progress__label {
  color: rgb(var(--color-primary));
}

.property-editor-form {
  display: grid;
  gap: 14px;
}

.property-editor-panel {
  border: 1px solid rgb(var(--color-border) / 0.8);
  border-radius: 8px;
  background: rgb(var(--color-surface));
  padding: 18px;
  box-shadow: 0 2px 8px rgb(15 23 42 / 0.03);
}

.property-editor-panel h2 {
  position: relative;
  margin: 0 0 14px;
  padding-left: 10px;
  color: rgb(var(--color-primary));
  font-size: 14px;
  font-weight: 700;
  letter-spacing: 0;
  line-height: 1.35;
}

.property-editor-panel h2::before {
  position: absolute;
  top: 0.16rem;
  bottom: 0.16rem;
  left: 0;
  width: 3px;
  border-radius: 999px;
  background: rgb(var(--color-primary));
  content: "";
}

.property-editor-panel p {
  color: rgb(var(--color-text-muted));
  font-size: 11px;
  line-height: 1.6;
}

.property-editor-panel strong {
  display: block;
  margin: 0.8rem 0;
  font-family: var(--font-display);
  font-size: 24px;
  font-weight: 400;
}

.property-editor-charge {
  margin: 0 0 0.8rem;
  color: rgb(var(--color-text));
  font-size: 11px;
  font-weight: 600;
}

.property-editor-grid {
  display: grid;
  gap: 12px;
}

.property-editor-grid--nested,
.property-input--compact {
  margin-top: 10px;
}

.property-editor-note {
  margin: 6px 0 10px;
  color: rgb(var(--color-text-muted));
  font-size: 11px;
  line-height: 1.6;
}

.property-ad-package-grid {
  display: grid;
  gap: 10px;
}

.property-ad-package-grid button {
  display: grid;
  gap: 6px;
  min-height: 82px;
  border: 1px solid rgb(var(--color-border) / 0.7);
  border-radius: 6px;
  background: rgb(var(--color-surface));
  padding: 14px;
  color: rgb(var(--color-text));
  text-align: left;
  cursor: pointer;
  transition:
    border-color 0.2s ease,
    background 0.2s ease,
    box-shadow 0.2s ease;
}

.property-ad-package-grid button:hover {
  border-color: rgb(var(--color-primary) / 0.4);
  box-shadow: 0 4px 12px rgb(15 23 42 / 0.05);
}

.property-ad-package-grid .property-ad-package--active {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary) / 0.06);
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
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0;
  text-transform: none;
}

.property-input > span:first-child {
  display: flex;
  min-height: 24px;
  align-items: center;
  line-height: 1.2;
}

.property-input--required > span:first-child::before,
.property-input--required > .property-input-label-row > span:first-child::before,
.property-checkbox-row--required > strong > span:first-child::before {
  content: '*';
  margin-right: 3px;
  color: rgb(220 38 38);
  font-weight: 700;
}

.property-input input,
.property-input select,
.property-input textarea {
  box-sizing: border-box;
  width: 100%;
  border: 1px solid rgb(var(--color-border));
  border-radius: 4px;
  background: rgb(var(--color-surface));
  padding: 8px 12px;
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

.property-input input:not([type='checkbox']),
.property-input select {
  height: 44px;
  min-height: 44px;
  padding-top: 0;
  padding-bottom: 0;
  line-height: 44px;
}

.property-phone-group__code,
.property-phone-group__number {
  height: 44px;
  min-height: 44px;
  padding-top: 0;
  padding-bottom: 0;
  line-height: 44px;
}

.property-input input:hover,
.property-input select:hover,
.property-input textarea:hover {
  border-color: rgb(var(--color-border-strong, 180 180 180));
}

.property-input input:focus,
.property-input select:focus,
.property-input textarea:focus {
  border-color: rgb(var(--color-primary));
  box-shadow: 0 0 0 3px rgb(var(--color-primary) / 0.08);
}

.property-input textarea {
  resize: vertical;
  min-height: 80px;
  line-height: 1.6;
}

.property-input select {
  appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 24 24' fill='none' stroke='%23999' stroke-width='2'%3E%3Cpolyline points='6 9 12 15 18 9'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 10px center;
  padding-right: 30px;
}

.property-input select:disabled {
  background-color: rgb(var(--color-muted) / 0.35);
  color: rgb(var(--color-text-muted));
  cursor: not-allowed;
}

.property-location-selectors {
  display: grid;
  gap: 10px;
}

@media (min-width: 760px) {
  .property-location-selectors {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
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
  top: calc(100% + 0.3rem);
  right: 0;
  left: 0;
  display: grid;
  max-height: 16rem;
  overflow-y: auto;
  border: 1px solid rgb(var(--color-border) / 0.7);
  border-radius: 6px;
  background: rgb(var(--color-surface));
  box-shadow: 0 8px 24px rgb(15 23 42 / 0.1);
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

.property-editor-panel .property-checkbox-row--block {
  align-items: flex-start;
  border: 1px solid rgb(var(--color-border) / 0.7);
  border-radius: 6px;
  background: rgb(var(--color-surface-muted) / 0.5);
  gap: 8px;
  padding: 12px 14px;
}

.property-editor-panel .property-checkbox-row--block strong {
  display: flex;
  flex: 0 0 100%;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px;
  margin: 0;
  padding-bottom: 6px;
  border-bottom: 1px solid rgb(var(--color-border) / 0.5);
  color: rgb(var(--color-text));
  font-family: var(--font-sans);
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0;
  line-height: 1.25;
}

.property-tag-group-title__meta {
  display: inline-flex;
  min-height: 16px;
  align-items: center;
  border: 0;
  border-radius: 999px;
  background: rgb(var(--color-primary) / 0.08);
  padding: 0 7px;
  color: rgb(var(--color-primary));
  font-size: 10px;
  font-weight: 600;
  line-height: 1;
}

.property-checkbox-row--inline {
  margin-top: 0;
}

.property-input--with-tools {
  align-content: start;
  gap: 6px;
}

.property-input-label-row {
  display: flex;
  min-height: 24px;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  line-height: 1.2;
}

.property-input-label-row > span:first-child {
  min-width: 0;
}

.property-compact-checkbox {
  display: inline-flex;
  min-height: 24px;
  align-items: center;
  gap: 6px;
  color: rgb(var(--color-text-muted));
  font-size: 11px;
  font-weight: 600;
  line-height: 1;
  white-space: nowrap;
  cursor: pointer;
  transition: color 0.15s ease;
}

.property-compact-checkbox:hover {
  color: rgb(var(--color-text));
}

.property-compact-checkbox input[type='checkbox'] {
  width: 18px;
  height: 18px;
  flex: 0 0 18px;
  margin: 0;
  border: 2px solid rgb(var(--color-border));
  border-radius: 4px;
  accent-color: rgb(var(--color-primary));
  cursor: pointer;
}

.property-input--inline-tools {
  align-content: center;
  gap: 0;
  min-height: 0;
  margin-top: -4px;
  padding-left: 10px;
  border-left: 2px solid rgb(var(--color-border) / 0.6);
}

.property-compact-checkbox-row {
  display: flex;
  min-height: 0;
  align-items: center;
  flex-wrap: wrap;
  gap: 4px 10px;
}

.property-compact-checkbox-row label {
  display: inline-flex;
  min-height: 26px;
  align-items: center;
  gap: 7px;
  border: 0;
  border-radius: 3px;
  background: transparent;
  padding: 0 2px;
  color: rgb(var(--color-text-muted));
  font-size: 11px;
  font-weight: 500;
  line-height: 1;
  cursor: pointer;
  transition: color 0.15s ease;
}

.property-compact-checkbox-row label:hover {
  color: rgb(var(--color-text));
}

.property-compact-checkbox-row input[type='checkbox'] {
  width: 18px;
  height: 18px;
  margin: 0;
  border: 1px solid rgb(var(--color-border));
  border-radius: 4px;
  accent-color: rgb(var(--color-primary));
  cursor: pointer;
}

.property-checkbox-row label,
.property-inline-checkbox {
  display: inline-flex;
  min-height: 28px;
  align-items: center;
  gap: 5px;
  border: 1px solid rgb(var(--color-border) / 0.7);
  border-radius: 4px;
  background: rgb(var(--color-surface));
  padding: 0 10px;
  color: rgb(var(--color-text-muted));
  font-size: 11px;
  font-weight: 500;
  cursor: pointer;
  transition:
    border-color 0.15s ease,
    color 0.15s ease,
    background 0.15s ease;
}

.property-inline-checkbox {
  align-self: center;
  min-height: 0;
  border: 0;
  background: transparent;
  padding: 0;
  color: rgb(var(--color-text-muted));
  font-size: 11px;
  font-weight: 500;
}

.property-checkbox-row label:hover {
  border-color: rgb(var(--color-primary) / 0.4);
  color: rgb(var(--color-text));
}

.property-inline-checkbox:hover {
  border: 0;
  color: rgb(var(--color-text));
}

.property-checkbox-row input[type='checkbox'],
.property-inline-checkbox input[type='checkbox'] {
  width: 13px;
  height: 13px;
  margin: 0;
  accent-color: rgb(var(--color-primary));
  cursor: pointer;
}

.property-editor-panel .property-checkbox-row--block label {
  min-height: 26px;
  padding: 0 8px;
  font-size: 11px;
  font-weight: 500;
  letter-spacing: 0;
  line-height: 1;
}

.property-checkbox-row--divider {
  margin-left: 6px;
  padding-left: 10px;
  border-left: 1px solid rgb(var(--color-border) / 0.6);
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
  border: 1px solid rgb(var(--color-border) / 0.7);
  border-radius: 6px;
  background: rgb(var(--color-surface-muted) / 0.5);
  padding: 14px;
}

.property-room-type__title {
  padding-bottom: 8px;
  border-bottom: 1px solid rgb(var(--color-border) / 0.5);
}

.property-room-type__title strong {
  margin: 0;
  color: rgb(var(--color-text));
  font-family: var(--font-sans);
  font-size: 12px;
  font-weight: 700;
}

.property-mini-button {
  min-height: 28px;
  border: 1px solid rgb(var(--color-primary));
  border-radius: 4px;
  background: rgb(var(--color-primary));
  padding: 0 12px;
  color: rgb(var(--color-primary-contrast));
  font-size: 11px;
  font-weight: 600;
  cursor: pointer;
  transition: opacity 0.15s ease;
}

.property-mini-button:hover {
  opacity: 0.85;
}

.property-mini-button--secondary {
  border-color: rgb(var(--color-border));
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text-muted));
}

.property-mini-button--secondary:hover {
  border-color: rgb(var(--color-text-muted));
  color: rgb(var(--color-text));
  opacity: 1;
}

.property-mini-button:disabled {
  cursor: not-allowed;
  opacity: 0.4;
}

.property-phone-group {
  display: flex;
  gap: 6px;
}

.property-phone-group__code {
  flex: 0 0 130px;
  box-sizing: border-box;
  height: 44px;
  min-height: 44px;
  border: 1px solid rgb(var(--color-border));
  border-radius: 4px;
  background: rgb(var(--color-surface));
  padding: 0 30px 0 10px;
  color: rgb(var(--color-text));
  font: inherit;
  font-size: 12px;
  line-height: 44px;
  outline: 0;
  appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 24 24' fill='none' stroke='%23999' stroke-width='2'%3E%3Cpolyline points='6 9 12 15 18 9'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 8px center;
  cursor: pointer;
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
}

.property-phone-group__code:hover {
  border-color: rgb(var(--color-border-strong, 180 180 180));
}

.property-phone-group__code:focus {
  border-color: rgb(var(--color-primary));
  box-shadow: 0 0 0 3px rgb(var(--color-primary) / 0.08);
}

.property-phone-group__number {
  flex: 1 1 0;
  min-width: 0;
}

.property-phone-whatsapp {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  margin-top: 4px;
  color: rgb(var(--color-text-muted));
  font-size: 11px;
  font-weight: 500;
  cursor: pointer;
}

.property-phone-whatsapp input[type='checkbox'] {
  width: 13px;
  height: 13px;
  margin: 0;
  accent-color: rgb(var(--color-primary));
  cursor: pointer;
}

.property-upload-button,
.property-editor-action {
  display: inline-flex;
  min-height: 38px;
  align-items: center;
  justify-content: center;
  gap: 6px;
  border-radius: 6px;
  padding: 0 14px;
  font-size: 12px;
  font-weight: 600;
  transition:
    border-color 0.15s ease,
    background 0.15s ease,
    box-shadow 0.15s ease;
}

.property-upload-button {
  border: 1px dashed rgb(var(--color-primary) / 0.5);
  background: rgb(var(--color-primary) / 0.04);
  color: rgb(var(--color-primary));
  cursor: pointer;
}

.property-upload-button:hover {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary) / 0.08);
  box-shadow: 0 2px 8px rgb(var(--color-primary) / 0.1);
}

.property-upload-button input {
  display: none;
}

.property-media-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
}

.property-media-header__note {
  flex: 1 1 240px;
  margin: 0;
  align-self: center;
}

.property-image-grid {
  display: grid;
  gap: 10px;
  margin-top: 12px;
}

.property-image-card__badge {
  position: absolute;
  top: 6px;
  left: 6px;
  z-index: 1;
  border-radius: 3px;
  background: rgb(var(--color-primary));
  padding: 2px 8px;
  color: rgb(var(--color-primary-contrast));
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 0.03em;
}

.property-image-card {
  position: relative;
  overflow: hidden;
  border: 1px solid rgb(var(--color-border) / 0.7);
  border-radius: 6px;
  background: rgb(var(--color-surface-muted));
  transition:
    border-color 0.15s ease,
    box-shadow 0.15s ease;
}

.property-image-card:hover {
  border-color: rgb(var(--color-primary) / 0.3);
  box-shadow: 0 4px 12px rgb(15 23 42 / 0.06);
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
  min-height: 30px;
  border: 0;
  border-top: 1px solid rgb(var(--color-border) / 0.6);
  background: transparent;
  color: rgb(var(--color-text-muted));
  font-size: 11px;
  font-weight: 500;
  cursor: pointer;
  transition: background 0.15s ease, color 0.15s ease;
}

.property-image-card button:hover {
  background: rgb(var(--color-surface-muted));
  color: rgb(var(--color-text));
}

.property-image-card button + button {
  border-left: 1px solid rgb(var(--color-border) / 0.6);
}

.property-image-card__action--active {
  color: rgb(var(--color-primary));
  font-weight: 600;
}

.property-editor-side {
  display: grid;
  align-content: start;
  gap: 10px;
}

.property-editor-side .property-editor-panel {
  position: relative;
  overflow: hidden;
  padding: 20px;
}

.property-side-summary {
  display: grid;
  gap: 8px;
  border-bottom: 1px solid rgb(var(--color-border) / 0.6);
  margin-bottom: 14px;
  padding-bottom: 14px;
}

.property-side-summary h2 {
  margin: 0;
  padding-left: 0;
  color: rgb(var(--color-text));
  font-size: 18px;
  font-weight: 600;
  line-height: 1.35;
}

.property-side-summary h2::before {
  content: none;
}

.property-side-summary p {
  margin: 0;
  color: rgb(var(--color-text-muted));
  font-size: 12px;
  line-height: 1.6;
}

.property-side-summary strong {
  margin: 2px 0 0;
  color: rgb(var(--color-primary));
  font-size: 22px;
}

.property-side-kicker {
  margin: 0 0 6px;
  color: rgb(var(--color-primary));
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 0.05em;
  text-transform: uppercase;
}

.property-side-section {
  display: grid;
  gap: 6px;
  margin-bottom: 14px;
}

.property-side-section__title {
  color: rgb(var(--color-text-muted));
  font-size: 11px;
  font-weight: 600;
}

.property-side-selected-category {
  display: flex;
  min-height: 34px;
  align-items: center;
  border: 1px solid rgb(var(--color-border) / 0.7);
  border-radius: 5px;
  background: rgb(var(--color-surface-muted) / 0.5);
  margin: 0;
  padding: 0 12px;
  color: rgb(var(--color-text));
  font-size: 12px;
  font-weight: 600;
  line-height: 1.4;
}

.property-side-meta {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px;
  margin-bottom: 14px;
}

.property-side-meta div {
  display: grid;
  gap: 3px;
  border: 1px solid rgb(var(--color-border) / 0.6);
  border-radius: 5px;
  background: rgb(var(--color-surface-muted) / 0.4);
  padding: 8px 10px;
}

.property-side-meta span {
  color: rgb(var(--color-text-muted));
  font-size: 10px;
  font-weight: 600;
}

.property-side-meta strong {
  margin: 0;
  color: rgb(var(--color-text));
  font-family: var(--font-sans);
  font-size: 13px;
  font-weight: 600;
}

.property-editor-step-actions {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px;
  margin-bottom: 8px;
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

.property-editor-action--primary:hover {
  opacity: 0.9;
}

.property-editor-action--secondary {
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text));
}

.property-editor-action--secondary:hover {
  border-color: rgb(var(--color-text-muted));
}

.property-editor-action:disabled {
  cursor: not-allowed;
  opacity: 0.4;
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
    grid-template-columns: minmax(0, 1fr) minmax(18rem, 22rem);
    align-items: start;
    gap: 16px;
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
  letter-spacing: 0;
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
  border-radius: 6px;
  box-shadow: 0 12px 28px rgb(15 23 42 / 0.05);
  padding: 16px;
}

.property-editor-panel h2 {
  color: rgb(var(--color-primary));
  font-size: 15px;
  font-weight: 700;
  letter-spacing: 0;
}

.property-editor-panel p,
.property-editor-charge,
.property-editor-note {
  font-size: 12px;
  line-height: 1.6;
}

.property-editor-panel strong {
  font-family: var(--font-display);
  font-size: 24px;
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

  .property-editor-progress {
    padding-bottom: 4px;
  }

  .property-editor-progress__label {
    display: none;
  }

  .property-side-meta {
    grid-template-columns: 1fr;
  }
}
</style>

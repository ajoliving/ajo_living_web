<!--
 * 物業發布編輯頁。
 * 1. 根據頻道建立或更新樓盤放售與服務式住宅草稿。
 * 2. 支援圖片上傳、草稿保存與發布。
-->
<script setup lang="ts">
import axios from 'axios';
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';

import { fetchCommunities } from '@/httpapis/communities';
import {
  createPropertySale,
  createServicedApartment,
  fetchPropertySaleDetail,
  fetchServicedApartmentDetail,
  publishPropertySale,
  publishServicedApartment,
  updatePropertySale,
  updateServicedApartment,
} from '@/httpapis/properties';
import { completeUpload, createUploadPresign } from '@/httpapis/uploads';
import {
  getPropertyOptionLabel,
  marketplaceDistricts,
  propertyBusinessStatusOptions,
  propertyContactMethodOptions,
  propertyFeatureTagOptions,
  propertyTypeOptions,
  servicedFacilityTagOptions,
  servicedServiceTagOptions,
} from '@/constants/property';
import type { MetaCommunity } from '@/model/community';
import type { MediaAssetResponse } from '@/model/marketplace';
import type {
  PropertyChannel,
  PropertyImagePayload,
  PropertyListingDetailResponse,
  UpsertPropertySalePayload,
  UpsertServicedApartmentPayload,
} from '@/model/property';
import AppIcon from '@/shared/components/base/AppIcon.vue';
import { useFeedbackStore } from '@/stores/feedback';
import { usePreferenceStore } from '@/stores/preferences';
import { useSessionStore } from '@/stores/session';
import { formatPrice } from '@/utils/format';
import { buildUploadHeaders } from '@/utils/upload';
import { formatAjoPoints, resolveWalletChargeCost } from '@/utils/wallet';

const props = defineProps<{
  channel: PropertyChannel;
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
  title: string;
  summary: string;
  description: string;
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
  propertyType: string;
  estateName: string;
  addressText: string;
  askingPriceHKD: number;
  usableAreaSqft: number;
  grossAreaSqft: number;
  bedroomCount: number;
  livingRoomCount: number;
  bathroomCount: number;
  floorLevel: string;
  direction: string;
  buildingAge: string;
  featureTags: string[];
  projectName: string;
  lowestMonthlyRentHKD: number;
  minLeaseMonths: number;
  facilityTags: string[];
  serviceTags: string[];
  roomName: string;
  roomAreaSqft: number;
  roomMonthlyRentHKD: number;
  roomIncludedFees: boolean;
  roomMinLeaseMonths: number;
}

const maxImages = 8;
const maxImageSize = 10 * 1024 * 1024;

const route = useRoute();
const router = useRouter();
const { t } = useI18n();
const feedbackStore = useFeedbackStore();
const preferenceStore = usePreferenceStore();
const sessionStore = useSessionStore();

const listingId = ref(String(route.params.listingId ?? ''));
const loading = ref(false);
const saving = ref(false);
const publishing = ref(false);
const communities = ref<MetaCommunity[]>([]);
const images = ref<PropertyEditorImage[]>([]);

const form = reactive<PropertyEditorForm>({
  title: '',
  summary: '',
  description: '',
  districtCode: sessionStore.me?.district_code || 'kwun_tong',
  communityId: sessionStore.currentUser.primary_community.public_id || '',
  publisherIdentityType: sessionStore.me?.publisher_identity_type || 'owner',
  businessStatus: 'available',
  contactMethod: 'both',
  phone: '',
  whatsapp: '',
  email: sessionStore.me?.email || '',
  allowPhone: true,
  allowWhatsapp: true,
  allowChat: false,
  propertyType: 'private_flat',
  estateName: '',
  addressText: '',
  askingPriceHKD: 0,
  usableAreaSqft: 0,
  grossAreaSqft: 0,
  bedroomCount: 2,
  livingRoomCount: 1,
  bathroomCount: 1,
  floorLevel: '',
  direction: '',
  buildingAge: '',
  featureTags: [],
  projectName: '',
  lowestMonthlyRentHKD: 0,
  minLeaseMonths: 1,
  facilityTags: [],
  serviceTags: [],
  roomName: 'Studio',
  roomAreaSqft: 0,
  roomMonthlyRentHKD: 0,
  roomIncludedFees: true,
  roomMinLeaseMonths: 1,
});

const isSale = computed(() => props.channel === 'sale');
const isEditing = computed(() => listingId.value.trim().length > 0);
const pageTitle = computed(() =>
  isSale.value ? t('property.sale.publishTitle') : t('property.serviced.publishTitle'),
);
const myPath = computed(() =>
  isSale.value ? '/properties/my' : '/serviced-residences/my',
);
const hasImage = computed(() => images.value.some((image) => image.mediaAssetId || image.file));
const canSave = computed(() =>
  form.title.trim() !== '' &&
  form.summary.trim() !== '' &&
  form.description.trim() !== '' &&
  form.districtCode.trim() !== '' &&
  form.addressText.trim() !== '' &&
  (isSale.value
    ? form.askingPriceHKD > 0 && form.usableAreaSqft > 0
    : form.projectName.trim() !== '' && form.lowestMonthlyRentHKD > 0 && form.minLeaseMonths > 0),
);
const previewPrice = computed(() =>
  formatPrice(isSale.value ? form.askingPriceHKD : form.lowestMonthlyRentHKD, preferenceStore.locale),
);
const chargeCost = computed(() =>
  resolveWalletChargeCost(isSale.value ? 'property_sale' : 'serviced_apartment'),
);
const formatPoints = (value: number): string =>
  formatAjoPoints(value, t('common.brand.pointsName'), preferenceStore.locale);
const chargeHint = computed(() =>
  `${t('property.editor.chargeHint')} ${formatPoints(chargeCost.value)} · ${t('property.editor.walletBalance')} ${formatPoints(sessionStore.me?.ajo_balance ?? 0)}`,
);

// 1. 讀取錯誤訊息
const readErrorMessage = (error: unknown, fallback: string): string =>
  axios.isAxiosError<{ message?: string }>(error)
    ? error.response?.data?.message ?? fallback
    : fallback;

// 2. 切換標籤
const toggleTag = (target: string[], value: string): void => {
  const index = target.indexOf(value);
  if (index >= 0) {
    target.splice(index, 1);
    return;
  }

  target.push(value);
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
const buildSalePayload = (): UpsertPropertySalePayload => ({
  title: form.title.trim(),
  summary: form.summary.trim(),
  description: form.description.trim(),
  district_code: form.districtCode,
  community_id: form.communityId,
  publisher_identity_type: form.publisherIdentityType || 'owner',
  property_type: form.propertyType,
  estate_name: form.estateName.trim(),
  address_text: form.addressText.trim(),
  asking_price_hkd: Number(form.askingPriceHKD),
  usable_area_sqft: Number(form.usableAreaSqft),
  gross_area_sqft: form.grossAreaSqft > 0 ? Number(form.grossAreaSqft) : undefined,
  bedroom_count: Number(form.bedroomCount),
  living_room_count: Number(form.livingRoomCount),
  bathroom_count: Number(form.bathroomCount),
  floor_level: form.floorLevel.trim(),
  direction: form.direction.trim(),
  building_age: form.buildingAge.trim(),
  feature_tags: form.featureTags,
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

// 13. 建立服務式住宅 payload
const buildServicedPayload = (): UpsertServicedApartmentPayload => ({
  title: form.title.trim(),
  summary: form.summary.trim(),
  description: form.description.trim(),
  district_code: form.districtCode,
  community_id: form.communityId,
  publisher_identity_type: form.publisherIdentityType || 'owner',
  project_name: form.projectName.trim(),
  address_text: form.addressText.trim(),
  lowest_monthly_rent_hkd: Number(form.lowestMonthlyRentHKD),
  min_lease_months: Number(form.minLeaseMonths),
  facility_tags: form.facilityTags,
  service_tags: form.serviceTags,
  room_types: [
    {
      name: form.roomName.trim() || 'Studio',
      usable_area_sqft: Number(form.roomAreaSqft || form.usableAreaSqft),
      monthly_rent_hkd: Number(form.roomMonthlyRentHKD || form.lowestMonthlyRentHKD),
      included_fees: form.roomIncludedFees,
      min_lease_months: Number(form.roomMinLeaseMonths || form.minLeaseMonths),
      feature_tags: [],
    },
  ],
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
    form.propertyType = detail.property_sale.property_type;
    form.estateName = detail.property_sale.estate_name;
    form.addressText = detail.property_sale.address_text;
    form.askingPriceHKD = detail.property_sale.asking_price_hkd;
    form.usableAreaSqft = detail.property_sale.usable_area_sqft;
    form.grossAreaSqft = detail.property_sale.gross_area_sqft ?? 0;
    form.bedroomCount = detail.property_sale.bedroom_count;
    form.livingRoomCount = detail.property_sale.living_room_count;
    form.bathroomCount = detail.property_sale.bathroom_count;
    form.floorLevel = detail.property_sale.floor_level;
    form.direction = detail.property_sale.direction;
    form.buildingAge = detail.property_sale.building_age;
    form.featureTags = [...detail.property_sale.feature_tags];
    form.contactMethod = detail.property_sale.contact_method as PropertyEditorForm['contactMethod'];
  }

  if (detail.serviced_apartment) {
    const room = detail.serviced_apartment.room_types[0];
    form.projectName = detail.serviced_apartment.project_name;
    form.addressText = detail.serviced_apartment.address_text;
    form.lowestMonthlyRentHKD = detail.serviced_apartment.lowest_monthly_rent_hkd;
    form.minLeaseMonths = detail.serviced_apartment.min_lease_months;
    form.facilityTags = [...detail.serviced_apartment.facility_tags];
    form.serviceTags = [...detail.serviced_apartment.service_tags];
    form.contactMethod = detail.serviced_apartment.contact_method as PropertyEditorForm['contactMethod'];
    if (room) {
      form.roomName = room.name;
      form.roomAreaSqft = room.usable_area_sqft;
      form.roomMonthlyRentHKD = room.monthly_rent_hkd;
      form.roomIncludedFees = room.included_fees;
      form.roomMinLeaseMonths = room.min_lease_months;
    }
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
    const response = isSale.value
      ? await fetchPropertySaleDetail(listingId.value)
      : await fetchServicedApartmentDetail(listingId.value);

    applyDetail(response.data.data);
  } catch (error: unknown) {
    feedbackStore.pushToast(readErrorMessage(error, t('property.detail.loadError')), 'error');
  } finally {
    loading.value = false;
  }
};

// 17. 儲存草稿
const saveDraft = async (): Promise<string> => {
  if (!canSave.value) {
    feedbackStore.pushToast(t('property.editor.requiredFields'), 'error');
    return '';
  }

  saving.value = true;
  try {
    if (!listingId.value) {
      const response = isSale.value
        ? await createPropertySale({ ...buildSalePayload(), images: [] })
        : await createServicedApartment({ ...buildServicedPayload(), images: [] });

      listingId.value = response.data.data.listing_id;
    }

    await uploadPendingImages(listingId.value);

    if (isSale.value) {
      await updatePropertySale(listingId.value, buildSalePayload());
    } else {
      await updateServicedApartment(listingId.value, buildServicedPayload());
    }
    await sessionStore.loadCurrentUser();

    feedbackStore.pushToast(t('property.editor.saveSuccess'), 'success');
    return listingId.value;
  } catch (error: unknown) {
    feedbackStore.pushToast(readErrorMessage(error, t('property.editor.saveError')), 'error');
    return '';
  } finally {
    saving.value = false;
  }
};

// 18. 儲存並發布
const saveAndPublish = async (): Promise<void> => {
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
    await router.push(myPath.value);
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
    await router.push(myPath.value);
  }
};

onMounted(async () => {
  await loadCommunities();
  await loadDetail();
});

onBeforeUnmount(() => {
  images.value.forEach(revokeImagePreview);
});
</script>

<template>
  <main class="property-editor-page">
    <section class="property-editor-heading">
      <p class="property-kicker">
        {{ isEditing ? 'Edit' : 'Publish' }}
      </p>
      <h1>{{ pageTitle }}</h1>
      <p>{{ previewPrice }}</p>
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
          <h2>{{ t('property.editor.baseInfo') }}</h2>
          <div class="property-editor-grid">
            <label class="property-input">
              <span>{{ t('property.editor.titleField') }}</span>
              <input v-model="form.title" />
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
                  {{ community.name_zh || community.name_en || community.address_text }}
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
          </div>
        </section>

        <section
          v-if="isSale"
          class="property-editor-panel"
        >
          <h2>{{ t('property.editor.saleTitle') }}</h2>
          <div class="property-editor-grid">
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
              <span>{{ t('property.editor.estateNameField') }}</span>
              <input v-model="form.estateName" />
            </label>
            <label class="property-input property-input--wide">
              <span>{{ t('property.editor.addressField') }}</span>
              <input v-model="form.addressText" />
            </label>
            <label class="property-input">
              <span>{{ t('property.editor.askingPriceField') }}</span>
              <input
                v-model.number="form.askingPriceHKD"
                type="number"
                min="0"
              />
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
              <input v-model="form.floorLevel" />
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
            <label class="property-input">
              <span>{{ t('property.editor.projectNameField') }}</span>
              <input v-model="form.projectName" />
            </label>
            <label class="property-input">
              <span>{{ t('property.editor.monthlyRentField') }}</span>
              <input
                v-model.number="form.lowestMonthlyRentHKD"
                type="number"
                min="0"
              />
            </label>
            <label class="property-input property-input--wide">
              <span>{{ t('property.editor.addressField') }}</span>
              <input v-model="form.addressText" />
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
              <span>{{ t('property.editor.roomTypeName') }}</span>
              <input v-model="form.roomName" />
            </label>
            <label class="property-input">
              <span>{{ t('property.editor.roomTypeArea') }}</span>
              <input
                v-model.number="form.roomAreaSqft"
                type="number"
                min="0"
              />
            </label>
            <label class="property-input">
              <span>{{ t('property.editor.roomTypeRent') }}</span>
              <input
                v-model.number="form.roomMonthlyRentHKD"
                type="number"
                min="0"
              />
            </label>
          </div>
          <label class="property-inline-checkbox">
            <input
              v-model="form.roomIncludedFees"
              type="checkbox"
            />
            {{ t('property.editor.includedFees') }}
          </label>

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
              <button
                type="button"
                @click="selectCover(image.id)"
              >
                {{ image.isCover ? 'Cover' : 'Set cover' }}
              </button>
              <button
                type="button"
                @click="removeImage(image.id)"
              >
                {{ t('common.action.retry') }}
              </button>
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
          <h2>{{ form.title || pageTitle }}</h2>
          <p>{{ form.summary }}</p>
          <strong>{{ previewPrice }}</strong>
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
            {{ publishing ? t('common.status.loading') : t('property.editor.publishNow') }}
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

.property-editor-heading {
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

.property-editor-layout {
  display: grid;
  gap: 1rem;
}

.property-editor-form {
  display: grid;
  gap: 1rem;
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

.property-checkbox-row {
  display: flex;
  flex-wrap: wrap;
  gap: 0.65rem;
  margin-top: 1rem;
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
  .property-image-grid {
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
</style>

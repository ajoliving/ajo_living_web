<!--
 * 樓盤租售詳情頁。
 * 1. 讀取樓盤買賣租賃詳情。
 * 2. 展示圖集、規格、發布者與聯絡方式解鎖。
-->
<script setup lang="ts">
import axios from 'axios';
import { computed, onMounted, ref } from 'vue';
import { RouterLink, useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';

import {
  fetchPropertySaleContactAccess,
  fetchPropertySaleDetail,
  fetchServicedApartmentContactAccess,
  fetchServicedApartmentDetail,
} from '@/domains/property/api';
import type {
  ContactAccessResult,
  PropertyChannel,
  PropertyListingDetailResponse,
  ServicedApartmentRoomType,
} from '@/domains/property/model';
import AppIcon from '@/shared/components/base/AppIcon.vue';
import BaseAvatar from '@/shared/components/base/BaseAvatar.vue';
import AppBreadcrumb from '@/shared/components/navigation/AppBreadcrumb.vue';
import { useFeedbackStore } from '@/app/stores/feedback';
import { usePreferenceStore } from '@/app/stores/preferences';
import { formatDate, formatPrice } from '@/shared/utils/format';
import {
  getPropertyOptionLabel,
  servicedFacilityTagOptions,
} from '@/domains/property/constants';
import {
  resolvePropertyArea,
  resolvePropertyCommunityName,
  resolvePropertyDistrict,
  resolvePropertyImages,
  resolvePropertyPriceText,
  resolvePropertyPublisherRole,
  resolvePropertyRooms,
  resolvePropertyTransactionType,
  resolvePropertyTypeLabel,
} from '@/shared/utils/property';

import {
  demoContactAccess,
  demoPropertyListings,
  getDemoPropertyDetail,
  isDemoPropertyListing,
  propertyAgentProfile,
  propertyNearbyPlaces,
} from '@/domains/property/demo';
import ServicedContactCard from './widgets/ServicedContactCard.vue';

const props = defineProps<{
  channel: PropertyChannel;
}>();

const route = useRoute();
const { t } = useI18n();
const feedbackStore = useFeedbackStore();
const preferenceStore = usePreferenceStore();
type AppIconName = InstanceType<typeof AppIcon>['$props']['name'];

interface ContactEntry {
  label: string;
  value: string;
  href: string;
  icon: AppIconName;
  displayText: string;
  isIconOnly: boolean;
  isAction: boolean;
}

interface DetailTableItem {
  label: string;
  value: string;
}

interface DetailTableRow {
  first: DetailTableItem;
  second?: DetailTableItem;
}

const listing = ref<PropertyListingDetailResponse | null>(null);
const contactAccess = ref<ContactAccessResult | null>(null);
const loading = ref(false);
const loadingContact = ref(false);

const listingId = computed(() => String(route.params.listingId ?? ''));
const galleryImages = computed(() => listing.value ? resolvePropertyImages(listing.value) : []);
const coverImage = computed(() => galleryImages.value[0]);
const servicedVisibleThumbs = computed(() => galleryImages.value.slice(1, 9));
const servicedHiddenThumbCount = computed(() => Math.max(galleryImages.value.length - 9, 0));
const priceText = computed(() =>
  listing.value ? resolvePropertyPriceText(listing.value, preferenceStore.locale) : '',
);
const publishedAt = computed(() =>
  listing.value
    ? formatDate(listing.value.published_at || listing.value.updated_at, preferenceStore.locale)
    : '',
);
const pageTitle = computed(() =>
  props.channel === 'sale' ? t('property.sale.detailTitle') : t('property.serviced.detailTitle'),
);
const priceLabel = computed(() =>
  props.channel === 'sale' && listing.value && resolvePropertyTransactionType(listing.value) === 'rent'
    ? t('property.sale.rentLabel')
    : props.channel === 'sale' ? t('property.sale.priceLabel') : t('property.serviced.priceLabel'),
);
const areaLabel = computed(() =>
  props.channel === 'sale' ? t('property.sale.areaLabel') : t('property.serviced.areaLabel'),
);
const formatAreaSqft = (value: number): string => t('common.unit.sqft', { value });
const formatMoney = (value?: number): string => value && value > 0 ? formatPrice(value, preferenceStore.locale) : '-';
const formatStay = (value?: number, unit?: string): string =>
  value && value > 0 ? `${value} ${unit === 'day' ? t('property.editor.dayUnit') : t('property.editor.monthUnit')}` : '-';
const formatRentRange = (min?: number, max?: number): string => {
  if (!min || min <= 0) {
    return '-';
  }
  if (max && max > min) {
    return `${formatPrice(min, preferenceStore.locale)} - ${formatPrice(max, preferenceStore.locale)}`;
  }
  return `${formatPrice(min, preferenceStore.locale)}起`;
};
const formatRoomMonthlyRent = (room: ServicedApartmentRoomType): string =>
  formatRentRange(room.monthly_rent_min_hkd || room.monthly_rent_hkd, room.monthly_rent_max_hkd);
const formatRoomDailyRent = (room: ServicedApartmentRoomType): string =>
  formatRentRange(room.daily_rent_min_hkd, room.daily_rent_max_hkd);
const formatServicedRoomRent = (room: ServicedApartmentRoomType): string =>
  (room.min_stay_unit ?? 'month') === 'day'
    ? formatRoomDailyRent(room)
    : formatRoomMonthlyRent(room);
const parseTextTags = (value?: string): string[] =>
  (value ?? '')
    .split(/\r?\n/)
    .map((item) => item.replace(/^[-•]\s*/, '').trim())
    .filter(Boolean);
const servicedResidenceTags = computed(() =>
  listing.value ? parseTextTags(listing.value.description || listing.value.serviced_apartment?.benefits_text) : [],
);
const servicedServiceTags = computed(() =>
  listing.value?.serviced_apartment
    ? [
      ...listing.value.serviced_apartment.service_tags,
      ...parseTextTags(listing.value.serviced_apartment.service_intro),
    ].filter((item, index, items) => items.indexOf(item) === index)
    : [],
);
const servicedExtraChargeTags = computed(() =>
  listing.value ? parseTextTags(listing.value.serviced_apartment?.extra_charges_text) : [],
);
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
const servicedFacilityDisplayItems = computed(() => {
  const activeTags = new Set([
    ...(listing.value?.serviced_apartment?.facility_tags ?? []),
    ...(listing.value?.serviced_apartment?.service_tags ?? []),
  ]);
  return servicedFacilityTagOptions.map((option) => ({
    label: getPropertyOptionLabel(option, preferenceStore.locale),
    icon: servicedFacilityIconMap[option.value] ?? 'check-circle',
    active: activeTags.has(option.label_zh_hk) || activeTags.has(option.label_en) || activeTags.has(option.value),
  }));
});
const resolveAnnualPrepayText = (value?: string): string => {
  if (value === '95_off') {
    return '95折';
  }
  if (value === '90_off') {
    return '9折';
  }
  return t('property.editor.notProvided');
};
const hasEnglishListingContent = computed(() =>
  Boolean(listing.value?.property_sale?.title_en || listing.value?.property_sale?.description_en),
);
const contactPayload = computed<Record<string, string>>(() => contactAccess.value?.contact_payload ?? {});
const buildPhoneHref = (phone: string): string => {
  const normalizedPhone = phone.replace(/\s+/g, '');
  return normalizedPhone ? `tel:${normalizedPhone}` : '';
};
const buildWhatsAppHref = (value: string): string => {
  if (value.startsWith('https://wa.me/')) {
    return value;
  }

  const digits = value.replace(/[+\s\-()]/g, '');
  return digits ? `https://wa.me/${digits}` : '';
};
const servicedPrimaryWhatsAppHref = computed(() => {
  const rawValue = listing.value?.serviced_apartment?.whatsapp || contactPayload.value.whatsapp_url || contactPayload.value.whatsapp || '';
  return rawValue ? buildWhatsAppHref(rawValue) : '';
});
const servicedVisibleWeChatValue = computed(() => contactPayload.value.wechat || contactPayload.value.wechat_id || '');
const servicedVisiblePhone = computed(() => contactPayload.value.phone || '');
const contactEntries = computed<ContactEntry[]>(() =>
  Object.entries(contactPayload.value).map(([key, value]) => {
    if (key === 'phone') {
      return {
        label: t('property.editor.phoneField'),
        value,
        href: buildPhoneHref(value),
        icon: 'phone' as const,
        displayText: '',
        isIconOnly: true,
        isAction: true,
      };
    }
    if (key === 'whatsapp_url' || key === 'whatsapp') {
      return {
        label: t('property.editor.whatsappField'),
        value,
        href: buildWhatsAppHref(value),
        icon: 'message' as const,
        displayText: 'WhatsApp',
        isIconOnly: false,
        isAction: true,
      };
    }
    if (key === 'email') {
      return { label: t('property.editor.emailField'), value, href: '', icon: 'message' as const, displayText: value, isIconOnly: false, isAction: false };
    }
    if (key === 'agent') {
      return { label: t('property.detail.contactPerson'), value, href: '', icon: 'user' as const, displayText: value, isIconOnly: false, isAction: false };
    }

    return { label: key, value, href: '', icon: 'message' as const, displayText: value, isIconOnly: false, isAction: false };
  }),
);
const salePhoneContact = computed(() =>
  contactEntries.value.find((entry) => entry.label === t('property.editor.phoneField') && entry.value),
);
const saleWhatsAppContact = computed(() =>
  contactEntries.value.find((entry) => entry.label === t('property.editor.whatsappField') && entry.href),
);
const ownerName = computed(() =>
  listing.value?.owner?.display_name || (listing.value ? resolvePropertyPublisherRole(listing.value) : ''),
);
const similarListings = computed(() =>
  demoPropertyListings.filter((item) => item.listing_id !== listing.value?.listing_id).slice(0, 3),
);
const groupedNearbyPlaces = computed(() =>
  propertyNearbyPlaces.reduce<Record<string, typeof propertyNearbyPlaces>>((result, place) => {
    result[place.categoryLabel] = [...(result[place.categoryLabel] ?? []), place];
    return result;
  }, {}),
);
const saleAddressText = computed(() => {
  if (!listing.value?.property_sale) {
    return '';
  }

  const district = resolvePropertyDistrict(listing.value, preferenceStore.locale);
  const address = preferenceStore.locale === 'en' && listing.value.property_sale.address_text_en
    ? listing.value.property_sale.address_text_en
    : listing.value.property_sale.address_text;
  return [district, address].filter(Boolean).join(' · ');
});
const saleHeroMetrics = computed(() => {
  if (!listing.value) {
    return [];
  }

  return [
    { label: priceLabel.value, value: priceText.value || '-' },
    { label: areaLabel.value, value: formatAreaSqft(resolvePropertyArea(listing.value)) },
    { label: t('property.filter.rooms'), value: resolvePropertyRooms(listing.value) },
    { label: t('property.detail.publishedAt'), value: publishedAt.value || '-' },
  ];
});
const buildDetailRows = (items: DetailTableItem[]): DetailTableRow[] =>
  items
    .filter((item) => item.value && item.value !== '-')
    .reduce<DetailTableRow[]>((rows, item, index) => {
      if (index % 2 === 0) {
        rows.push({ first: item });
      } else {
        rows[rows.length - 1].second = item;
      }
      return rows;
    }, []);
const saleSpecRows = computed<DetailTableRow[]>(() => {
  if (!listing.value) {
    return [];
  }

  const sale = listing.value.property_sale;
  return buildDetailRows([
    { label: priceLabel.value, value: priceText.value || '-' },
    { label: areaLabel.value, value: formatAreaSqft(resolvePropertyArea(listing.value)) },
    { label: t('property.editor.grossAreaField'), value: sale?.gross_area_sqft ? formatAreaSqft(sale.gross_area_sqft) : '-' },
    { label: t('property.filter.rooms'), value: resolvePropertyRooms(listing.value) },
    { label: t('property.sale.estateLabel'), value: resolvePropertyCommunityName(listing.value) },
    { label: t('property.sale.typeLabel'), value: resolvePropertyTypeLabel(listing.value, preferenceStore.locale) },
    { label: t('property.list.district'), value: resolvePropertyDistrict(listing.value, preferenceStore.locale) },
    { label: t('property.detail.address'), value: sale?.address_text || '-' },
    { label: t('property.editor.addressEnField'), value: sale?.address_text_en || '-' },
    { label: t('property.editor.floorZoneField'), value: sale?.public_location_text || '-' },
    { label: t('property.editor.blockNameField'), value: sale?.block_name || '-' },
    { label: t('property.editor.unitNameField'), value: sale?.unit_name || '-' },
    { label: t('property.editor.buildingAgeField'), value: sale?.building_age || '-' },
    { label: t('property.editor.managementFeeField'), value: formatMoney(sale?.management_fee_hkd) },
    { label: t('property.detail.publishedAt'), value: publishedAt.value || '-' },
  ]);
});
const saleContentRows = computed<DetailTableRow[]>(() => {
  const sale = listing.value?.property_sale;
  if (!sale) {
    return [];
  }

  return buildDetailRows([
    { label: t('property.editor.directionField'), value: sale.direction || '-' },
    { label: t('property.editor.kitchenTypeField'), value: sale.kitchen_type || '-' },
    { label: t('property.editor.cookingModeField'), value: sale.cooking_mode || '-' },
    { label: t('property.editor.videoUrlField'), value: sale.video_url || '-' },
    { label: t('property.editor.vrUrlField'), value: sale.vr_url || '-' },
    { label: t('property.editor.annualPrepayDiscountField'), value: sale.transaction_type === 'rent' ? resolveAnnualPrepayText(sale.annual_prepay_option) : '-' },
    { label: t('property.editor.leaseStartDateField'), value: sale.lease_start_date || '-' },
    { label: t('property.editor.rentIncludedField'), value: sale.rent_included || '-' },
  ]);
});
const servicedTitle = computed(() =>
  listing.value?.serviced_apartment?.project_name || listing.value?.title || '',
);
const servicedSubtitle = computed(() =>
  listing.value?.serviced_apartment?.project_name_en || '',
);
const servicedAddressText = computed(() => {
  if (!listing.value?.serviced_apartment) {
    return '';
  }

  const district = resolvePropertyDistrict(listing.value, preferenceStore.locale);
  const address = preferenceStore.locale === 'en' && listing.value.serviced_apartment.address_text_en
    ? listing.value.serviced_apartment.address_text_en
    : listing.value.serviced_apartment.address_text;
  return [district, address].filter(Boolean).join(' · ');
});
const breadcrumbItems = computed(() => [
  { label: t('nav.home'), to: '/' },
  {
    label: props.channel === 'sale' ? t('nav.properties') : t('nav.servicedResidences'),
    to: props.channel === 'sale' ? '/properties' : '/serviced-residences',
  },
  {
    label: listing.value
      ? listing.value.serviced_apartment?.project_name || listing.value.property_sale?.estate_name || listing.value.title
      : pageTitle.value,
  },
]);

// 1. 讀取詳情
const loadDetail = async (): Promise<void> => {
  if (!listingId.value) {
    return;
  }

  loading.value = true;
  try {
    if (props.channel === 'sale' && isDemoPropertyListing(listingId.value)) {
      listing.value = getDemoPropertyDetail(listingId.value) ?? null;
      return;
    }

    const response = props.channel === 'sale'
      ? await fetchPropertySaleDetail(listingId.value)
      : await fetchServicedApartmentDetail(listingId.value);

    listing.value = response.data.data;
  } catch (error: unknown) {
    feedbackStore.pushToast(
      axios.isAxiosError<{ message?: string }>(error)
        ? error.response?.data?.message ?? t('property.detail.loadError')
        : t('property.detail.loadError'),
      'error',
    );
    if (props.channel === 'sale') {
      listing.value = getDemoPropertyDetail(listingId.value) ?? null;
    }
  } finally {
    loading.value = false;
  }
};

// 2. 解鎖聯絡方式
const revealContact = async (): Promise<void> => {
  if (!listing.value) {
    return;
  }

  loadingContact.value = true;
  try {
    if (props.channel === 'sale' && isDemoPropertyListing(listing.value.listing_id)) {
      contactAccess.value = {
        ...demoContactAccess,
        listing_id: listing.value.listing_id,
      };
      return;
    }

    const response = props.channel === 'sale'
      ? await fetchPropertySaleContactAccess(listing.value.listing_id)
      : await fetchServicedApartmentContactAccess(listing.value.listing_id);

    contactAccess.value = response.data.data;
  } catch (error: unknown) {
    feedbackStore.pushToast(
      axios.isAxiosError<{ message?: string }>(error)
        ? error.response?.data?.message ?? t('property.detail.contactError')
        : t('property.detail.contactError'),
      'error',
    );
  } finally {
    loadingContact.value = false;
  }
};

onMounted(() => {
  void loadDetail();
});
</script>

<template>
  <main class="property-detail-page">
    <AppBreadcrumb
      :items="breadcrumbItems"
      class="property-detail-breadcrumb"
    />
    <section
      v-if="loading"
      class="property-detail-panel"
    >
      {{ t('common.status.loading') }}
    </section>

    <template v-else-if="listing">
      <template v-if="listing.serviced_apartment">
        <section class="property-serviced-layout">
          <div class="property-serviced-content">
            <section class="property-serviced-hero">
              <div class="property-serviced-hero-copy">
                <p class="property-serviced-kicker">{{ t('property.serviced.detailTitle') }}</p>
                <h1>{{ servicedTitle }}</h1>
                <p v-if="servicedSubtitle" class="property-serviced-subtitle">{{ servicedSubtitle }}</p>
                <div class="property-serviced-address">
                  <AppIcon name="location" :size="18" />
                  <span>{{ servicedAddressText }}</span>
                </div>
              </div>

              <div class="property-serviced-gallery-section">
                <div class="property-serviced-gallery-main">
                  <img
                    v-if="coverImage"
                    :src="coverImage.url"
                    :alt="listing.title"
                  />
                  <div v-else class="property-gallery__placeholder">
                    <AppIcon name="picture" :size="46" />
                  </div>
                  <span v-if="galleryImages.length > 0">{{ `1/${galleryImages.length}` }}</span>
                </div>
                <div v-if="galleryImages.length > 1" class="property-serviced-thumbs">
                  <div
                    v-for="(image, imageIndex) in servicedVisibleThumbs"
                    :key="image.id"
                    class="property-serviced-thumb"
                  >
                    <img
                      :src="image.url"
                      :alt="image.alt"
                    />
                    <span
                      v-if="imageIndex === servicedVisibleThumbs.length - 1 && servicedHiddenThumbCount > 0"
                    >
                      +{{ servicedHiddenThumbCount }}
                    </span>
                  </div>
                </div>
              </div>
            </section>

            <section
              v-if="listing.serviced_apartment.room_types.length"
              class="property-serviced-card"
            >
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
                  <span>{{ t('property.editor.roomActionColumn') }}</span>
                </div>
                <div
                  v-for="room in listing.serviced_apartment.room_types"
                  :key="room.name"
                  class="property-serviced-room-row"
                >
                  <span class="property-serviced-room-name">
                    <strong>{{ room.name }}</strong>
                    <small>{{ room.usable_area_sqft ? formatAreaSqft(room.usable_area_sqft) : '-' }}</small>
                  </span>
                  <span class="property-serviced-room-price">
                    <small>{{ formatStay(room.min_stay_value || room.min_lease_months, room.min_stay_unit || 'month') }}</small>
                    <strong>{{ formatServicedRoomRent(room) }}</strong>
                  </span>
                  <button
                    type="button"
                    class="property-serviced-room-action"
                    @click="revealContact"
                    :aria-label="t('property.editor.whatsappInquiryAction')"
                  >
                    <AppIcon name="message" :size="17" />
                  </button>
                </div>
              </div>
            </section>

            <section class="property-serviced-card">
              <div class="property-serviced-card-title">
                <div>
                  <h3>{{ t('property.editor.residenceInfoSection') }}</h3>
                </div>
              </div>
              <div class="property-serviced-description">
                <p v-if="servicedResidenceTags.length > 0">{{ servicedResidenceTags.join(' ') }}</p>
                <p v-if="servicedServiceTags.length > 0">{{ servicedServiceTags.join(' ') }}</p>
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
                  v-for="item in servicedFacilityDisplayItems"
                  :key="item.label"
                  :class="item.active ? 'property-serviced-facility--active' : 'property-serviced-facility--inactive'"
                >
                  <AppIcon :name="item.icon" :size="26" />
                  <span>{{ item.label }}</span>
                </div>
              </div>
            </section>

            <section
              v-if="servicedExtraChargeTags.length > 0"
              class="property-serviced-card"
            >
              <div class="property-serviced-card-title">
                <div>
                  <h3>{{ t('property.editor.extraChargesField') }}</h3>
                </div>
              </div>
              <ul class="property-serviced-fee-list">
                <li
                  v-for="item in servicedExtraChargeTags"
                  :key="item"
                >
                  {{ item }}
                </li>
              </ul>
            </section>
          </div>

          <aside class="property-serviced-side">
            <ServicedContactCard
              :primary-whats-app-href="servicedPrimaryWhatsAppHref"
              :website-url="listing.serviced_apartment.website_url"
              :we-chat-value="servicedVisibleWeChatValue"
              :phone-value="servicedVisiblePhone"
              :has-contact-access="Boolean(contactAccess)"
              :loading-contact="loadingContact"
              @reveal-contact="revealContact"
            />
          </aside>

          <div class="property-serviced-mobile-action">
            <a
              v-if="servicedPrimaryWhatsAppHref"
              class="property-serviced-whatsapp-action"
              :href="servicedPrimaryWhatsAppHref"
              target="_blank"
              rel="noopener noreferrer"
            >
              <AppIcon name="message" :size="18" />
              {{ t('property.editor.whatsappInquiryAction') }}
            </a>
            <button
              v-else
              type="button"
              class="property-serviced-whatsapp-action"
              @click="revealContact"
            >
              <AppIcon name="message" :size="18" />
              {{ t('property.editor.whatsappInquiryAction') }}
            </button>
          </div>
        </section>
      </template>

      <section v-else class="property-detail-main property-sale-main">
        <section class="property-sale-hero">
          <div class="property-sale-hero-copy">
            <p class="property-kicker">
              {{ pageTitle }}
            </p>
            <h1>{{ listing.title }}</h1>
            <div class="property-sale-address">
              <AppIcon name="location" :size="18" />
              <span>{{ saleAddressText }}</span>
            </div>
            <p v-if="listing.summary" class="property-detail-summary">
              {{ listing.summary }}
            </p>
            <div class="property-sale-metrics">
              <div
                v-for="item in saleHeroMetrics"
                :key="item.label"
              >
                <span>{{ item.label }}</span>
                <strong>{{ item.value }}</strong>
              </div>
            </div>
          </div>

          <div class="property-gallery property-sale-gallery">
            <div class="property-gallery__hero">
              <img
                v-if="coverImage"
                :src="coverImage.url"
                :alt="listing.title"
              />
              <div
                v-else
                class="property-gallery__placeholder"
              >
                <AppIcon
                  name="picture"
                  :size="46"
                />
              </div>
            </div>

            <div
              v-if="galleryImages.length > 1"
              class="property-gallery__thumbs"
            >
              <img
                v-for="image in galleryImages.slice(1, 5)"
                :key="image.id"
                :src="image.url"
                :alt="image.alt"
              />
            </div>
          </div>
        </section>

        <article class="property-detail-card">
          <section class="property-detail-section">
            <h2>{{ t('property.detail.description') }}</h2>
            <p>{{ listing.description }}</p>
          </section>

          <section class="property-detail-section">
            <h2>{{ t('property.detail.specification') }}</h2>
            <div class="property-info-table">
              <div
                v-for="row in saleSpecRows"
                :key="row.first.label"
                class="property-info-table__row"
              >
                <span class="property-info-table__label">{{ row.first.label }}</span>
                <strong>{{ row.first.value }}</strong>
                <span class="property-info-table__label">{{ row.second?.label || '' }}</span>
                <strong>{{ row.second?.value || '' }}</strong>
              </div>
            </div>
          </section>

          <section
            v-if="saleContentRows.length > 0"
            class="property-detail-section"
          >
            <h2>{{ t('property.editor.descriptionInfoSection') }}</h2>
            <div class="property-info-table">
              <div
                v-for="row in saleContentRows"
                :key="row.first.label"
                class="property-info-table__row"
              >
                <span class="property-info-table__label">{{ row.first.label }}</span>
                <strong>{{ row.first.value }}</strong>
                <span class="property-info-table__label">{{ row.second?.label || '' }}</span>
                <strong>{{ row.second?.value || '' }}</strong>
              </div>
            </div>
          </section>

          <section
            v-if="hasEnglishListingContent && listing.property_sale"
            class="property-detail-section"
          >
            <h2>English</h2>
            <dl class="property-spec-grid">
              <div v-if="listing.property_sale.title_en">
                <dt>{{ t('property.editor.titleEnField') }}</dt>
                <dd>{{ listing.property_sale.title_en }}</dd>
              </div>
              <div v-if="listing.property_sale.description_en">
                <dt>{{ t('property.editor.descriptionEnField') }}</dt>
                <dd>{{ listing.property_sale.description_en }}</dd>
              </div>
            </dl>
          </section>

          <section
            v-if="channel === 'sale'"
            class="property-detail-section"
          >
            <h2>{{ t('property.detail.mapAndNearby') }}</h2>
            <div class="property-map-panel">
              <div class="property-map-canvas">
                <span>{{ resolvePropertyCommunityName(listing) }}</span>
                <small>{{ listing.property_sale?.address_text }}</small>
              </div>
              <div class="property-nearby-grid">
                <article
                  v-for="(places, category) in groupedNearbyPlaces"
                  :key="category"
                  class="property-nearby-card"
                >
                  <h3>{{ category }}</h3>
                  <div
                    v-for="place in places.slice(0, 3)"
                    :key="`${place.category}-${place.name}`"
                    class="property-nearby-row"
                  >
                    <span>{{ place.name }}（{{ place.type }}）</span>
                    <strong>{{ place.walkMinutes }}{{ t('property.detail.walkMinute') }}</strong>
                  </div>
                </article>
              </div>
            </div>
          </section>

          <section
            v-if="channel === 'sale'"
            class="property-detail-section"
          >
            <h2>{{ t('property.detail.similarUnits') }}</h2>
            <div class="property-similar-grid">
              <RouterLink
                v-for="item in similarListings"
                :key="item.listing_id"
                :to="`/properties/${item.listing_id}`"
                class="property-similar-card"
              >
                <img
                  v-if="item.cover_image"
                  :src="item.cover_image.url"
                  :alt="item.title"
                />
                <div>
                  <span>{{ resolvePropertyCommunityName(item) }}</span>
                  <strong>{{ resolvePropertyPriceText(item, preferenceStore.locale) }}</strong>
                </div>
              </RouterLink>
            </div>
          </section>
        </article>
      </section>

      <aside
        v-if="!listing.serviced_apartment"
        class="property-detail-sidebar"
      >
        <section class="property-detail-panel property-sale-contact-card">
          <div class="property-sale-card-header">
            <p class="property-panel-label">
              {{ t('property.detail.owner') }}
            </p>
          </div>
          <div class="property-owner">
            <BaseAvatar
              :src="listing.owner?.avatar_url || ''"
              :name="ownerName"
              :alt="ownerName"
              :size="54"
            />
            <div>
              <h2>{{ ownerName }}</h2>
              <p>{{ resolvePropertyPublisherRole(listing) }}</p>
            </div>
          </div>

          <div class="property-contact-actions">
            <a
              v-if="saleWhatsAppContact"
              class="property-contact-action property-contact-action--whatsapp"
              :href="saleWhatsAppContact.href"
              target="_blank"
              rel="noopener noreferrer"
              :aria-label="saleWhatsAppContact.label"
            >
              <AppIcon
                name="message"
                :size="18"
              />
              <span>{{ saleWhatsAppContact.displayText || t('property.editor.whatsappField') }}</span>
            </a>
            <button
              v-else
              type="button"
              class="property-contact-action property-contact-action--whatsapp"
              :disabled="loadingContact"
              @click="revealContact"
            >
              <AppIcon
                name="message"
                :size="18"
              />
              <span>{{ loadingContact ? t('property.detail.loadingContact') : t('property.editor.whatsappField') }}</span>
            </button>

            <div
              v-if="salePhoneContact"
              class="property-contact-info"
            >
              <AppIcon
                name="phone"
                :size="16"
              />
              <span>{{ salePhoneContact.value }}</span>
            </div>
            <button
              v-else
              type="button"
              class="property-contact-reveal"
              :disabled="loadingContact"
              @click="revealContact"
            >
              <AppIcon
                name="phone"
                :size="16"
              />
              <span>{{ loadingContact ? t('property.detail.loadingContact') : t('property.detail.revealContact') }}</span>
            </button>
          </div>
        </section>

        <section
          v-if="channel === 'sale'"
          class="property-detail-panel"
        >
          <p class="property-panel-label">
            {{ t('property.detail.agentCompany') }}
          </p>
          <div class="property-agent-card">
            <h2>{{ propertyAgentProfile.companyName }}</h2>
            <p>{{ propertyAgentProfile.officeAddress }}</p>
            <dl>
              <div>
                <dt>{{ t('property.detail.companyLicense') }}</dt>
                <dd>{{ propertyAgentProfile.companyLicense }}</dd>
              </div>
              <div>
                <dt>{{ t('property.detail.agentLicense') }}</dt>
                <dd>{{ propertyAgentProfile.agentLicense }}</dd>
              </div>
            </dl>
            <div class="property-agent-badges">
              <span
                v-for="badge in propertyAgentProfile.verifiedBadges"
                :key="badge"
              >
                {{ badge }}
              </span>
            </div>
          </div>
        </section>

      </aside>
    </template>
  </main>
</template>

<style scoped>
.property-detail-page {
  display: grid;
  width: 100%;
  max-width: var(--layout-page-max-width);
  gap: 1.25rem;
  margin: 0 auto;
  padding: 3rem var(--layout-page-padding-inline) 5rem;
  color: rgb(var(--color-text));
}

.property-detail-main {
  display: grid;
  gap: 1.25rem;
  min-width: 0;
}

.property-detail-breadcrumb {
  min-width: 0;
}

.property-gallery,
.property-detail-card,
.property-detail-panel {
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.75rem;
  background: rgb(var(--color-surface));
  box-shadow: 0 10px 34px rgb(15 23 42 / 0.07);
}

.property-gallery {
  overflow: hidden;
}

.property-gallery__hero {
  aspect-ratio: 4 / 3;
  background: rgb(var(--color-surface-raised));
}

.property-gallery__hero img,
.property-gallery__thumbs img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.property-gallery__placeholder {
  display: grid;
  height: 100%;
  place-items: center;
  color: rgb(var(--color-text-muted));
}

.property-gallery__thumbs {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 0.75rem;
  padding: 0.75rem;
}

.property-gallery__thumbs img {
  aspect-ratio: 4 / 3;
  border-radius: 0.5rem;
}

.property-detail-card,
.property-detail-panel {
  padding: 1.25rem;
}

.property-kicker,
.property-panel-label {
  margin: 0 0 0.6rem;
  color: rgb(var(--color-primary));
  font-size: 0.75rem;
  font-weight: 900;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.property-detail-card h1 {
  margin: 0;
  font-family: var(--font-display);
  font-size: clamp(2rem, 5vw, 3.5rem);
  font-weight: 600;
  line-height: 1.12;
}

.property-detail-card > h2 {
  margin: 0 0 1rem;
  font-size: 1.05rem;
  font-weight: 900;
}

.property-detail-summary {
  max-width: 45rem;
  margin: 1rem 0 0;
  color: rgb(var(--color-text-muted));
  font-size: 1rem;
  line-height: 1.7;
}

.property-detail-price {
  margin: 1rem 0 0;
  font-family: var(--font-display);
  font-size: 2rem;
  font-weight: 700;
}

.property-detail-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin-top: 1rem;
}

.property-detail-tags span {
  border-radius: 999px;
  background: rgb(var(--color-primary-soft) / 0.45);
  padding: 0.35rem 0.65rem;
  color: rgb(var(--color-primary));
  font-size: 0.76rem;
  font-weight: 850;
}

.property-detail-section {
  margin-top: 1.5rem;
  border-top: 1px solid rgb(var(--color-border));
  padding-top: 1.25rem;
}

.property-detail-section--plain {
  margin-top: 0;
  border-top: 0;
  padding-top: 0;
}

.property-detail-section h2 {
  margin: 0 0 0.8rem;
  font-size: 1rem;
  font-weight: 900;
}

.property-detail-section p {
  margin: 0;
  color: rgb(var(--color-text-muted));
  line-height: 1.8;
}

.property-info-table {
  display: grid;
  overflow: hidden;
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.6rem;
  background: rgb(var(--color-surface));
}

.property-info-table__row {
  display: grid;
  grid-template-columns: minmax(7rem, 0.75fr) minmax(0, 1.25fr) minmax(7rem, 0.75fr) minmax(0, 1.25fr);
  min-height: 3.1rem;
}

.property-info-table__row + .property-info-table__row {
  border-top: 1px solid rgb(var(--color-border));
}

.property-info-table__row > span,
.property-info-table__row > strong {
  display: flex;
  align-items: center;
  min-width: 0;
  padding: 0.8rem 0.9rem;
  overflow-wrap: anywhere;
  line-height: 1.45;
}

.property-info-table__row > * + * {
  border-left: 1px solid rgb(var(--color-border));
}

.property-info-table__label {
  background: rgb(var(--color-surface-raised));
  color: rgb(var(--color-text-muted));
  font-size: 0.78rem;
  font-weight: 900;
}

.property-info-table__row > strong {
  color: rgb(var(--color-text));
  font-size: 0.9rem;
  font-weight: 850;
}

.property-sale-main {
  gap: 1.5rem;
}

.property-sale-hero {
  display: grid;
  gap: 1.25rem;
  overflow: hidden;
  border: 1px solid rgb(222 227 232);
  border-radius: 0.5rem;
  background: rgb(var(--color-surface));
  padding: 1.25rem;
  box-shadow: 0 20px 46px rgb(15 23 42 / 0.08);
}

.property-sale-hero-copy {
  display: grid;
  align-content: start;
  gap: 0.9rem;
  min-width: 0;
}

.property-sale-hero h1 {
  margin: 0;
  font-family: var(--font-display);
  font-size: clamp(2.1rem, 4vw, 3.4rem);
  font-weight: 700;
  line-height: 1.05;
}

.property-sale-address {
  display: flex;
  align-items: flex-start;
  gap: 0.45rem;
  color: rgb(var(--color-text-muted));
  font-weight: 850;
  line-height: 1.5;
}

.property-sale-address .app-icon {
  flex: 0 0 auto;
  margin-top: 0.15rem;
  color: rgb(var(--color-primary));
}

.property-sale-metrics {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 0.75rem;
}

.property-sale-metrics div {
  display: grid;
  min-height: 5rem;
  align-content: center;
  gap: 0.3rem;
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.5rem;
  background: rgb(var(--color-surface-raised));
  padding: 0.8rem;
}

.property-sale-metrics span {
  color: rgb(var(--color-text-muted));
  font-size: 0.74rem;
  font-weight: 900;
}

.property-sale-metrics strong {
  color: rgb(var(--color-text));
  font-size: 1rem;
  font-weight: 950;
  line-height: 1.2;
  overflow-wrap: anywhere;
}

.property-sale-gallery {
  border: 0;
  border-radius: 0.5rem;
  background: transparent;
  box-shadow: none;
}

.property-sale-gallery .property-gallery__hero {
  min-height: 0;
  aspect-ratio: 16 / 9;
  border-radius: 0.5rem;
  overflow: hidden;
}

.property-serviced-layout {
  --serviced-theme-card: 255 255 255;
  --serviced-theme-muted: 247 249 250;
  --serviced-theme-row: 241 245 249;
  --serviced-theme-primary: 18 91 83;
  --serviced-theme-primary-strong: 12 72 66;
  --serviced-theme-text: 17 24 39;
  --serviced-theme-text-muted: 88 99 115;
  --serviced-theme-border: 222 227 232;
  --serviced-theme-outline: 145 155 166;
  display: grid;
  grid-column: 1 / -1;
  gap: 1.5rem;
  min-width: 0;
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
.property-serviced-side-card {
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
  overflow: hidden;
  border: 1px solid rgb(var(--serviced-theme-border));
  border-radius: 0.5rem;
  background: rgb(var(--serviced-theme-card));
  padding: 1.25rem;
  box-shadow: 0 20px 46px rgb(15 23 42 / 0.08);
}

.property-serviced-hero-copy {
  display: grid;
  align-content: start;
  gap: 0.9rem;
  min-width: 0;
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
  max-width: 44rem;
  color: rgb(var(--serviced-theme-text-muted));
  font-weight: 800;
  line-height: 1.5;
}

.property-serviced-address .app-icon {
  flex: 0 0 auto;
  margin-top: 0.15rem;
  color: rgb(var(--serviced-theme-primary));
}

.property-serviced-metrics {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 0.75rem;
  margin-top: 0.25rem;
}

.property-serviced-metrics div {
  display: grid;
  min-height: 5rem;
  align-content: center;
  gap: 0.3rem;
  border: 1px solid rgb(var(--serviced-theme-border));
  border-radius: 0.5rem;
  background: rgb(var(--serviced-theme-muted));
  padding: 0.8rem;
}

.property-serviced-metrics span {
  color: rgb(var(--serviced-theme-text-muted));
  font-size: 0.74rem;
  font-weight: 900;
}

.property-serviced-metrics strong {
  color: rgb(var(--serviced-theme-text));
  font-size: 1rem;
  font-weight: 950;
  line-height: 1.2;
  overflow-wrap: anywhere;
}

.property-serviced-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 0.45rem;
}

.property-serviced-tags span {
  border: 1px solid rgb(var(--serviced-theme-border));
  border-radius: 999px;
  background: #fff;
  padding: 0.38rem 0.65rem;
  color: rgb(var(--serviced-theme-text-muted));
  font-size: 0.78rem;
  font-weight: 850;
}

.property-serviced-gallery-section {
  display: grid;
  gap: 0.25rem;
}

.property-serviced-gallery-main {
  position: relative;
  overflow: hidden;
  aspect-ratio: 16 / 9;
  border: 1px solid rgb(var(--serviced-theme-border));
  border-radius: 0.5rem;
  background: rgb(var(--serviced-theme-row));
}

.property-serviced-gallery-main img,
.property-serviced-thumb img {
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

.property-serviced-gallery-main .property-gallery__placeholder {
  min-height: 100%;
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
  font-size: 1rem;
  font-weight: 900;
}

.property-serviced-card-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  border-bottom: 1px solid rgb(var(--serviced-theme-border));
  background: rgb(var(--serviced-theme-card));
  padding: 1rem;
  color: rgb(var(--serviced-theme-text));
}

.property-serviced-card-title div {
  display: grid;
  gap: 0.25rem;
}

.property-serviced-card-title h3 {
  margin: 0;
  font-size: 1.05rem;
  font-weight: 900;
}

.property-serviced-card-title p {
  margin: 0;
  color: rgb(var(--serviced-theme-text-muted));
  font-size: 0.78rem;
  font-weight: 800;
  line-height: 1.45;
  text-align: right;
}

.property-serviced-room-table {
  display: grid;
  margin: 0;
  border: 1px solid rgb(var(--serviced-theme-border));
  border-width: 0;
  border-radius: 0;
  overflow: hidden;
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

.property-serviced-room-row {
  background: rgb(var(--serviced-theme-card));
  transition: background 160ms ease;
}

.property-serviced-room-row:hover {
  background: rgb(244 250 248);
}

.property-serviced-room-row + .property-serviced-room-row {
  border-top: 1px solid rgb(var(--serviced-theme-border));
}

.property-serviced-room-row > span {
  display: grid;
  gap: 0.25rem;
  min-width: 0;
  overflow-wrap: anywhere;
  color: rgb(var(--serviced-theme-text-muted));
  font-size: 0.9rem;
  font-weight: 850;
}

.property-serviced-room-row strong {
  color: rgb(var(--serviced-theme-text));
  font-weight: 900;
}

.property-serviced-room-row small {
  display: block;
  color: rgb(var(--serviced-theme-text-muted));
  font-size: 0.78rem;
  line-height: 1.5;
}

.property-serviced-room-name strong,
.property-serviced-room-price strong {
  line-height: 1.35;
}

.property-serviced-room-name small {
  margin-top: 0.15rem;
}

.property-serviced-room-price {
  align-items: start;
}

.property-serviced-room-action {
  display: inline-flex;
  width: 2.35rem;
  min-width: 2.35rem;
  height: 2.35rem;
  min-height: 2.35rem;
  align-items: center;
  justify-content: center;
  border: 1px solid #25d366;
  border-radius: 0.45rem;
  background: #25d366;
  padding: 0;
  color: #fff;
  font-weight: 900;
  white-space: nowrap;
}

.property-serviced-footnote {
  margin: 0 1rem 1rem;
  color: rgb(var(--color-text-muted));
  font-size: 0.78rem;
  font-weight: 800;
  text-align: right;
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
  justify-content: flex-start;
  gap: 0.75rem;
  border-radius: 0.5rem;
  padding: 0.9rem;
  text-align: left;
}

.property-serviced-facility-grid span {
  font-size: 0.82rem;
  font-weight: 900;
}

.property-serviced-facility-grid > div,
.property-serviced-side-card {
  border: 1px solid rgb(var(--serviced-theme-border));
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

.property-serviced-side-list {
  display: grid;
  gap: 0.8rem;
  margin: 1rem 0 0;
  border-top: 1px solid rgb(var(--serviced-theme-border));
  padding-top: 1rem;
}

.property-serviced-side-list div {
  display: grid;
  gap: 0.25rem;
}

.property-serviced-side-list dt {
  color: rgb(var(--serviced-theme-text-muted));
  font-size: 0.75rem;
  font-weight: 900;
}

.property-serviced-side-list dd {
  margin: 0;
  font-weight: 850;
  line-height: 1.5;
  overflow-wrap: anywhere;
}

.property-contact-button--serviced {
  margin-top: 0;
  border-color: rgb(var(--serviced-theme-primary));
  background: rgb(var(--serviced-theme-primary));
}

.property-serviced-mobile-action {
  display: none;
}

.property-map-panel {
  display: grid;
  gap: 1rem;
}

.property-map-canvas {
  display: grid;
  min-height: 18rem;
  align-content: center;
  justify-items: center;
  gap: 0.5rem;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background:
    linear-gradient(90deg, rgb(var(--color-border) / 0.3) 1px, transparent 1px),
    linear-gradient(rgb(var(--color-border) / 0.3) 1px, transparent 1px),
    rgb(var(--color-surface-raised));
  background-size: 2.5rem 2.5rem;
  color: rgb(var(--color-primary));
  text-align: center;
}

.property-map-canvas span {
  border-radius: 999px;
  background: rgb(var(--color-primary));
  padding: 0.55rem 0.85rem;
  color: rgb(var(--color-primary-contrast));
  font-weight: 900;
}

.property-map-canvas small {
  color: rgb(var(--color-text-muted));
  font-weight: 800;
}

.property-nearby-grid,
.property-similar-grid {
  display: grid;
  gap: 0.75rem;
}

.property-nearby-card,
.property-similar-card {
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface-raised));
  padding: 0.85rem;
}

.property-nearby-card h3 {
  margin: 0 0 0.75rem;
  font-size: 0.95rem;
  font-weight: 900;
}

.property-nearby-row {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  color: rgb(var(--color-text-muted));
  font-size: 0.86rem;
  line-height: 1.5;
}

.property-nearby-row + .property-nearby-row {
  margin-top: 0.5rem;
}

.property-nearby-row strong {
  flex-shrink: 0;
  color: rgb(var(--color-primary));
}

.property-similar-card {
  display: grid;
  grid-template-columns: 6rem minmax(0, 1fr);
  gap: 0.75rem;
  color: inherit;
}

.property-similar-card img {
  width: 100%;
  aspect-ratio: 4 / 3;
  border-radius: 6px;
  object-fit: cover;
}

.property-similar-card div {
  display: grid;
  align-content: center;
  min-width: 0;
  gap: 0.25rem;
}

.property-similar-card span,
.property-similar-card strong {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.property-similar-card span {
  color: rgb(var(--color-text-muted));
  font-size: 0.82rem;
  font-weight: 850;
}

.property-similar-card strong {
  font-weight: 900;
}

.property-spec-grid {
  display: grid;
  gap: 0.85rem;
  margin: 0;
}

.property-spec-grid {
  grid-template-columns: repeat(2, minmax(0, 1fr));
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

.property-room-table {
  display: grid;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  overflow: hidden;
}

.property-room-table__head,
.property-room-table__row {
  display: grid;
  grid-template-columns: 1.35fr 0.8fr 1fr 1fr 0.8fr;
  gap: 0.75rem;
  align-items: center;
  padding: 0.75rem;
}

.property-room-table__head--compact,
.property-room-table__row--compact {
  grid-template-columns: 1.35fr 0.8fr 0.8fr 1fr;
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

.property-detail-sidebar {
  display: grid;
  align-content: start;
  gap: 1rem;
}

.property-sale-contact-card {
  overflow: hidden;
  padding: 0;
}

.property-sale-card-header {
  border-bottom: 1px solid rgb(var(--color-border));
  background: rgb(var(--color-surface-raised));
  padding: 0.95rem 1.1rem;
}

.property-sale-card-header .property-panel-label {
  margin: 0;
}

.property-owner {
  display: flex;
  align-items: center;
  gap: 0.85rem;
  padding: 1rem 1.1rem;
}

.property-owner h2 {
  margin: 0;
  font-size: 1rem;
  font-weight: 900;
}

.property-owner p {
  margin: 0.2rem 0 0;
  color: rgb(var(--color-text-muted));
  font-size: 0.85rem;
}

.property-contact-actions {
  display: grid;
  gap: 0;
  border-top: 1px solid rgb(var(--color-border));
}

.property-contact-action,
.property-contact-reveal,
.property-contact-info {
  display: flex;
  width: 100%;
  min-height: 2.75rem;
  align-items: center;
  justify-content: center;
  gap: 0.65rem;
  border: 0;
  border-bottom: 1px solid rgb(var(--color-border));
  padding: 0 1rem;
  font-weight: 900;
  text-decoration: none;
}

.property-contact-actions > :last-child {
  border-bottom: 0;
}

.property-contact-action--whatsapp {
  background: #25d366;
  color: #fff;
  transition: background 0.2s ease-in-out;
}

.property-contact-action--whatsapp:hover {
  background: #1cb14f;
}

.property-contact-action--whatsapp:disabled,
.property-contact-reveal:disabled {
  cursor: not-allowed;
  opacity: 0.65;
}

.property-contact-reveal {
  background: rgb(var(--color-surface-raised));
  color: rgb(var(--color-primary));
  transition: background 0.2s ease-in-out;
}

.property-contact-reveal:hover {
  background: rgb(var(--color-primary-soft) / 0.24);
}

.property-contact-info {
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text));
  font-size: 0.9rem;
}

.property-contact-info .app-icon {
  flex: 0 0 auto;
  color: rgb(var(--color-primary));
}

.property-contact-info span {
  min-width: 0;
  overflow: hidden;
  text-align: center;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.property-room-stack {
  display: grid;
  gap: 0.65rem;
}

.property-room-card {
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.5rem;
  padding: 0.8rem;
}

.property-room-card h3,
.property-room-card p {
  margin: 0;
}

.property-room-card p {
  margin-top: 0.3rem;
  font-weight: 900;
}

.property-room-card span {
  display: block;
  margin-top: 0.25rem;
  color: rgb(var(--color-text-muted));
  font-size: 0.82rem;
}

.property-agent-card {
  display: grid;
  gap: 0.75rem;
}

.property-agent-card h2,
.property-agent-card p,
.property-agent-card dl {
  margin: 0;
}

.property-agent-card h2 {
  font-size: 1rem;
  font-weight: 900;
  line-height: 1.4;
}

.property-agent-card p {
  color: rgb(var(--color-text-muted));
  line-height: 1.6;
}

.property-agent-card dl {
  display: grid;
  gap: 0.65rem;
}

.property-agent-card dt {
  color: rgb(var(--color-text-muted));
  font-size: 0.75rem;
  font-weight: 900;
}

.property-agent-card dd {
  margin: 0.2rem 0 0;
  font-weight: 900;
}

.property-agent-badges {
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem;
}

.property-agent-badges span {
  border-radius: 999px;
  background: rgb(var(--color-primary-soft) / 0.5);
  padding: 0.35rem 0.55rem;
  color: rgb(var(--color-primary));
  font-size: 0.72rem;
  font-weight: 900;
}

@media (min-width: 1000px) {
  .property-detail-page {
    grid-template-columns: minmax(0, 1fr) 24rem;
  }

  .property-detail-breadcrumb {
    grid-column: 1 / -1;
  }

  .property-detail-main {
    grid-column: 1 / 2;
  }

  .property-detail-sidebar {
    grid-column: 2 / 3;
  }

  .property-serviced-layout {
    grid-column: 1 / -1;
    grid-template-columns: minmax(0, 1fr) 22rem;
    align-items: start;
  }

  .property-serviced-hero {
    grid-template-columns: 1fr;
  }

  .property-serviced-gallery-section {
    align-content: stretch;
  }

  .property-serviced-gallery-main {
    min-height: 31rem;
  }

  .property-nearby-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .property-detail-sidebar {
    position: sticky;
    top: calc(var(--app-header-offset, 0rem) + 2rem);
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
  .property-detail-page {
    padding-top: 1.25rem;
    padding-bottom: 6rem;
  }

  .property-sale-hero {
    padding: 0.9rem;
  }

  .property-sale-hero-copy {
    order: 2;
  }

  .property-sale-gallery {
    order: 1;
  }

  .property-sale-metrics {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .property-info-table__row {
    grid-template-columns: minmax(6.5rem, 0.8fr) minmax(0, 1.2fr);
  }

  .property-info-table__row > :nth-child(3),
  .property-info-table__row > :nth-child(4) {
    border-top: 1px solid rgb(var(--color-border));
  }

  .property-info-table__row > :nth-child(3) {
    border-left: 0;
  }

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

  .property-serviced-hero-copy h1 {
    font-size: clamp(2rem, 13vw, 3.15rem);
  }

  .property-serviced-metrics {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .property-serviced-card-title {
    align-items: flex-start;
    flex-direction: column;
  }

  .property-serviced-card-title p {
    text-align: left;
  }

  .property-serviced-room-head {
    display: none;
  }

  .property-serviced-room-row {
    grid-template-columns: minmax(0, 1fr) 2.35rem;
    align-items: stretch;
  }

  .property-serviced-room-row > span {
    grid-column: 1 / 2;
  }

  .property-serviced-room-row button {
    grid-column: 2 / 3;
    grid-row: 1 / 2;
    align-self: start;
  }

  .property-serviced-thumbs {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .property-serviced-facility-grid {
    grid-template-columns: 1fr;
  }

  .property-serviced-side {
    position: static;
  }

  .property-serviced-mobile-action {
    position: fixed;
    right: var(--layout-page-padding-inline, 1rem);
    bottom: 1rem;
    left: var(--layout-page-padding-inline, 1rem);
    z-index: 30;
    display: flex;
    gap: 0.5rem;
    border: 1px solid rgb(255 255 255 / 0.75);
    border-radius: 0.75rem;
    background: rgb(255 255 255 / 0.92);
    padding: 0.5rem;
    box-shadow: 0 16px 36px rgb(15 23 42 / 0.18);
    backdrop-filter: blur(14px);
  }

  .property-serviced-mobile-action .property-serviced-whatsapp-action {
    flex: 1;
    min-height: 2.8rem;
    background: #25d366;
    color: #fff;
    border: none;
    border-radius: 0.6rem;
    padding: 0 1rem;
    font-size: 0.9rem;
  }
}
</style>

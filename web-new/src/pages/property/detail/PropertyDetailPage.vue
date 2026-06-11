<!--
 * 物業頻道詳情頁。
 * 1. 根據頻道讀取樓盤或服務式住宅詳情。
 * 2. 展示圖集、規格、發布者與聯絡方式解鎖。
-->
<script setup lang="ts">
import axios from 'axios';
import { computed, onMounted, ref, watch } from 'vue';
import { RouterLink, useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';

import {
  fetchPropertySaleContactAccess,
  fetchPropertySaleDetail,
  fetchServicedApartmentContactAccess,
  fetchServicedApartmentDetail,
} from '@/httpapis/properties';
import type {
  ContactAccessResult,
  PropertyChannel,
  PropertyListingDetailResponse,
} from '@/model/property';
import AppIcon from '@/shared/components/base/AppIcon.vue';
import BaseAvatar from '@/shared/components/base/BaseAvatar.vue';
import { useFeedbackStore } from '@/stores/feedback';
import { usePreferenceStore } from '@/stores/preferences';
import { formatDate, formatPrice } from '@/utils/format';
import {
  resolvePropertyArea,
  resolvePropertyCommunityName,
  resolvePropertyDistrict,
  resolvePropertyImages,
  resolvePropertyPrice,
  resolvePropertyPublisherRole,
  resolvePropertyRooms,
  resolvePropertyTypeLabel,
} from '@/utils/property';

const props = defineProps<{
  channel: PropertyChannel;
}>();

const route = useRoute();
const { t } = useI18n();
const feedbackStore = useFeedbackStore();
const preferenceStore = usePreferenceStore();

const listing = ref<PropertyListingDetailResponse | null>(null);
const contactAccess = ref<ContactAccessResult | null>(null);
const loading = ref(false);
const loadingContact = ref(false);
const selectedImageIndex = ref(0);

const listingId = computed(() => String(route.params.listingId ?? ''));
const galleryImages = computed(() => listing.value ? resolvePropertyImages(listing.value) : []);
const coverImage = computed(() => galleryImages.value[selectedImageIndex.value] ?? galleryImages.value[0]);
const isSale = computed(() => props.channel === 'sale');
const salePayload = computed(() => listing.value?.property_sale ?? null);
const servicedPayload = computed(() => listing.value?.serviced_apartment ?? null);
const listPath = computed(() => (isSale.value ? '/properties' : '/serviced-residences'));
const listLabel = computed(() => (isSale.value ? t('nav.properties') : t('nav.servicedResidences')));
const priceText = computed(() => {
  if (!listing.value) {
    return '';
  }
  if (salePayload.value?.price_negotiable || servicedPayload.value?.price_negotiable) {
    return '面議';
  }
  const price = salePayload.value?.transaction_type === 'rent'
    ? salePayload.value.monthly_rent_hkd ?? 0
    : resolvePropertyPrice(listing.value);
  if (price <= 0) {
    return '待定';
  }

  return salePayload.value?.price_reference_only || servicedPayload.value?.price_reference_only
    ? `${formatPrice(price, preferenceStore.locale)} 起`
    : formatPrice(price, preferenceStore.locale);
});
const priceUnit = computed(() => {
  if (servicedPayload.value) {
    return '/ 月起';
  }

  return salePayload.value?.transaction_type === 'rent' ? '/ 月' : '';
});
const publishedAt = computed(() =>
  listing.value
    ? formatDate(listing.value.published_at || listing.value.updated_at, preferenceStore.locale)
    : '',
);
const pageTitle = computed(() =>
  props.channel === 'sale' ? t('property.sale.detailTitle') : t('property.serviced.detailTitle'),
);
const priceLabel = computed(() =>
  props.channel === 'sale' ? t('property.sale.priceLabel') : t('property.serviced.priceLabel'),
);
const areaLabel = computed(() =>
  props.channel === 'sale' ? t('property.sale.areaLabel') : t('property.serviced.areaLabel'),
);
const formatAreaSqft = (value: number): string => t('common.unit.sqft', { value });
const formatRoomSpec = (area: number, months: number): string =>
  t('common.unit.roomSpec', { area, months });
const contactPayload = computed(() => contactAccess.value?.contact_payload ?? {});
const contactPayloadEntries = computed(() =>
  Object.entries(contactPayload.value).map(([key, value]) => ({
    key,
    label: resolveContactLabel(key),
    text: key === 'whatsapp_url' ? 'WhatsApp' : value,
    href: key === 'whatsapp_url' ? value : key === 'phone' ? `tel:${value}` : '',
  })),
);
const contactChannels = computed(() => {
  if (!listing.value) {
    return [];
  }
  return [
    listing.value.contact_summary.show_phone ? '電話' : '',
    listing.value.contact_summary.show_whatsapp ? 'WhatsApp' : '',
    listing.value.contact_summary.show_chat ? '站內聊天' : '',
    listing.value.contact_summary.show_inquiry_form ? '查詢表格' : '',
  ].filter(Boolean);
});
const ownerName = computed(() =>
  listing.value?.owner?.display_name ||
  (listing.value ? resolvePropertyPublisherRole(listing.value) : ''),
);
const detailTags = computed(() => {
  if (!listing.value) {
    return [];
  }
  return [
    resolvePropertyPublisherRole(listing.value),
    resolvePropertyDistrict(listing.value, preferenceStore.locale),
    resolvePropertyTypeLabel(listing.value, preferenceStore.locale),
    salePayload.value?.transaction_type === 'rent' ? '出租' : salePayload.value ? '出售' : '',
  ].filter(Boolean);
});
const specRows = computed(() => {
  if (!listing.value) {
    return [];
  }
  const rows = salePayload.value
    ? [
      { label: '參考編號', value: salePayload.value.property_no || listing.value.listing_id },
      { label: priceLabel.value, value: `${priceText.value}${priceUnit.value}` },
      { label: areaLabel.value, value: formatAreaSqft(resolvePropertyArea(listing.value)) },
      { label: '實用 / 建築', value: formatAreaPair() },
      { label: t('property.detail.address'), value: salePayload.value.public_location_text || salePayload.value.address_text },
      { label: t('property.list.district'), value: resolvePropertyDistrict(listing.value, preferenceStore.locale) },
      { label: t('property.sale.estateLabel'), value: resolvePropertyCommunityName(listing.value) },
      { label: t('property.sale.typeLabel'), value: resolvePropertyTypeLabel(listing.value, preferenceStore.locale) },
      { label: '間隔', value: resolvePropertyRooms(listing.value) },
      { label: '樓層', value: salePayload.value.floor_display_range || salePayload.value.floor_level },
      { label: '座向', value: salePayload.value.direction },
      { label: '樓齡', value: salePayload.value.building_age },
      { label: '管理費', value: salePayload.value.management_fee_hkd ? formatPrice(salePayload.value.management_fee_hkd, preferenceStore.locale) : '' },
      { label: '特色', value: salePayload.value.feature_tags.join('、') },
      { label: t('property.detail.publishedAt'), value: publishedAt.value },
    ]
    : [
      { label: priceLabel.value, value: `${priceText.value}${priceUnit.value}` },
      { label: areaLabel.value, value: formatAreaSqft(resolvePropertyArea(listing.value)) },
      { label: t('property.sale.estateLabel'), value: resolvePropertyCommunityName(listing.value) },
      { label: t('property.detail.address'), value: servicedPayload.value?.address_text },
      { label: t('property.list.district'), value: resolvePropertyDistrict(listing.value, preferenceStore.locale) },
      { label: '最短入住', value: formatStayText() },
      { label: '設施', value: servicedPayload.value?.facility_tags.join('、') },
      { label: '服務', value: servicedPayload.value?.service_tags.join('、') },
      { label: '網站', value: servicedPayload.value?.website_url },
      { label: t('property.detail.publishedAt'), value: publishedAt.value },
    ];

  return rows.filter((row) => String(row.value ?? '').trim());
});

// 1. 取得聯絡欄位標籤
const resolveContactLabel = (key: string): string => {
  const labels: Record<string, string> = {
    phone: '電話',
    whatsapp_url: 'WhatsApp',
    email: 'Email',
  };

  return labels[key] ?? key;
};

// 2. 格式化面積組合
const formatAreaPair = (): string => {
  const sale = salePayload.value;
  if (!sale) {
    return '';
  }
  const usable = sale.usable_area_sqft > 0 ? formatAreaSqft(sale.usable_area_sqft) : '';
  const gross = sale.gross_area_sqft && sale.gross_area_sqft > 0
    ? formatAreaSqft(sale.gross_area_sqft)
    : '';

  return [usable, gross].filter(Boolean).join(' / ');
};

// 3. 格式化最短入住
const formatStayText = (): string => {
  const serviced = servicedPayload.value;
  if (!serviced) {
    return '';
  }
  const value = serviced.min_stay_value || serviced.min_lease_months;
  const unit = serviced.min_stay_unit === 'day' ? '日' : '個月';

  return value > 0 ? `${value}${unit}` : '';
};

// 4. 格式化房型價格
const formatRoomPrice = (monthlyRent: number): string =>
  monthlyRent > 0 ? formatPrice(monthlyRent, preferenceStore.locale) : '面議';

// 5. 讀取詳情
const loadDetail = async (): Promise<void> => {
  if (!listingId.value) {
    return;
  }

  loading.value = true;
  try {
    const response = props.channel === 'sale'
      ? await fetchPropertySaleDetail(listingId.value)
      : await fetchServicedApartmentDetail(listingId.value);

    listing.value = response.data.data;
    selectedImageIndex.value = 0;
    contactAccess.value = null;
  } catch (error: unknown) {
    feedbackStore.pushToast(
      axios.isAxiosError<{ message?: string }>(error)
        ? error.response?.data?.message ?? t('property.detail.loadError')
        : t('property.detail.loadError'),
      'error',
    );
  } finally {
    loading.value = false;
  }
};

// 6. 解鎖聯絡方式
const revealContact = async (): Promise<void> => {
  if (!listing.value) {
    return;
  }

  loadingContact.value = true;
  try {
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

watch(listingId, () => {
  void loadDetail();
});
</script>

<template>
  <main class="property-detail-page">
    <nav class="property-breadcrumb">
      <RouterLink to="/">
        {{ t('nav.home') }}
      </RouterLink>
      <span>/</span>
      <RouterLink :to="listPath">
        {{ listLabel }}
      </RouterLink>
      <span>/</span>
      <strong>{{ listing?.title || pageTitle }}</strong>
    </nav>

    <section
      v-if="loading"
      class="property-detail-panel"
    >
      {{ t('common.status.loading') }}
    </section>

    <template v-else-if="listing">
      <section class="property-detail-main">
        <div class="property-gallery">
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
            <button
              v-for="(image, index) in galleryImages.slice(0, 5)"
              :key="image.id"
              type="button"
              :class="{ 'property-gallery__thumb--active': selectedImageIndex === index }"
              @click="selectedImageIndex = index"
            >
              <img
                :src="image.url"
                :alt="image.alt"
              />
            </button>
          </div>
        </div>

        <article class="property-detail-card">
          <p class="property-kicker">
            {{ pageTitle }}
          </p>
          <div class="property-detail-tags">
            <span
              v-for="tag in detailTags"
              :key="tag"
            >
              {{ tag }}
            </span>
          </div>
          <h1>{{ listing.title }}</h1>
          <p class="property-detail-summary">
            {{ listing.summary }}
          </p>
          <p class="property-detail-price">
            {{ priceText }}
            <small v-if="priceUnit">{{ priceUnit }}</small>
          </p>

          <section class="property-detail-section">
            <h2>{{ t('property.detail.description') }}</h2>
            <p>{{ listing.description }}</p>
          </section>

          <section class="property-detail-section">
            <h2>{{ t('property.detail.specification') }}</h2>
            <dl class="property-spec-grid">
              <div
                v-for="row in specRows"
                :key="row.label"
              >
                <dt>{{ row.label }}</dt>
                <dd>{{ row.value }}</dd>
              </div>
            </dl>
          </section>

          <section
            v-if="salePayload?.video_url || salePayload?.vr_url || servicedPayload?.service_intro || servicedPayload?.benefits_text || servicedPayload?.extra_charges_text"
            class="property-detail-section"
          >
            <h2>補充資料</h2>
            <p v-if="servicedPayload?.service_intro">{{ servicedPayload.service_intro }}</p>
            <p v-if="servicedPayload?.benefits_text">{{ servicedPayload.benefits_text }}</p>
            <p v-if="servicedPayload?.extra_charges_text">{{ servicedPayload.extra_charges_text }}</p>
            <div class="property-link-row">
              <a
                v-if="salePayload?.video_url"
                :href="salePayload.video_url"
                target="_blank"
                rel="noreferrer"
              >
                影片
              </a>
              <a
                v-if="salePayload?.vr_url"
                :href="salePayload.vr_url"
                target="_blank"
                rel="noreferrer"
              >
                VR
              </a>
            </div>
          </section>
        </article>
      </section>

      <aside class="property-detail-sidebar">
        <section class="property-detail-panel">
          <p class="property-panel-label">
            {{ t('property.detail.owner') }}
          </p>
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

          <div
            v-if="contactChannels.length"
            class="property-contact-channels"
          >
            <span
              v-for="channel in contactChannels"
              :key="channel"
            >
              {{ channel }}
            </span>
          </div>

          <button
            type="button"
            class="property-contact-button"
            :disabled="loadingContact"
            @click="revealContact"
          >
            <AppIcon
              name="phone"
              :size="17"
            />
            {{ loadingContact ? t('property.detail.loadingContact') : t('property.detail.revealContact') }}
          </button>

          <dl
            v-if="contactPayloadEntries.length > 0"
            class="property-contact-list"
          >
            <div
              v-for="item in contactPayloadEntries"
              :key="item.key"
            >
              <dt>{{ item.label }}</dt>
              <dd>
                <a
                  v-if="item.href"
                  :href="item.href"
                  target="_blank"
                  rel="noreferrer"
                >
                  {{ item.text }}
                </a>
                <span v-else>{{ item.text }}</span>
              </dd>
            </div>
          </dl>
        </section>

        <section
          v-if="listing.serviced_apartment?.room_types.length"
          class="property-detail-panel"
        >
          <p class="property-panel-label">
            {{ t('property.editor.roomTypeName') }}
          </p>
          <div class="property-room-stack">
            <article
              v-for="room in listing.serviced_apartment.room_types"
              :key="room.name"
              class="property-room-card"
            >
              <h3>{{ room.name }}</h3>
              <p>{{ formatRoomPrice(room.monthly_rent_hkd) }}</p>
              <span>{{ formatRoomSpec(room.usable_area_sqft, room.min_lease_months) }}</span>
            </article>
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
  gap: 14px;
  margin: 0 auto;
  padding: 18px var(--layout-page-padding-inline) 72px;
  color: rgb(var(--color-text));
}

.property-breadcrumb {
  display: flex;
  min-width: 0;
  grid-column: 1 / -1;
  flex-wrap: wrap;
  align-items: center;
  gap: 7px;
  color: rgb(var(--color-text-muted));
  font-size: 11px;
  line-height: 1.6;
}

.property-breadcrumb a {
  color: rgb(var(--color-text-muted));
  text-decoration: none;
}

.property-breadcrumb a:hover {
  color: rgb(var(--color-primary));
}

.property-breadcrumb strong {
  max-width: 26rem;
  overflow: hidden;
  color: rgb(var(--color-text));
  font-weight: 500;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.property-detail-page > .property-detail-panel {
  grid-column: 1 / -1;
}

.property-detail-main {
  display: grid;
  gap: 14px;
  min-width: 0;
}

.property-gallery,
.property-detail-card,
.property-detail-panel {
  border: 1px solid rgb(var(--color-border));
  border-radius: 3px;
  background: rgb(var(--color-surface));
  box-shadow: none;
}

.property-gallery {
  overflow: hidden;
}

.property-gallery__hero {
  aspect-ratio: 4 / 3;
  background: rgb(var(--color-surface-muted));
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
  gap: 8px;
  padding: 8px;
}

.property-gallery__thumbs button {
  overflow: hidden;
  border: 1px solid transparent;
  border-radius: 2px;
  background: transparent;
  cursor: pointer;
  padding: 0;
}

.property-gallery__thumb--active {
  border-color: rgb(var(--color-primary)) !important;
}

.property-gallery__thumbs img {
  aspect-ratio: 4 / 3;
  display: block;
}

.property-detail-card,
.property-detail-panel {
  padding: 14px;
}

.property-kicker,
.property-panel-label {
  margin: 0 0 8px;
  color: rgb(var(--color-primary));
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0.18em;
  text-transform: uppercase;
}

.property-detail-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  margin-bottom: 8px;
}

.property-detail-tags span {
  border: 1px solid rgb(var(--color-border));
  border-radius: 2px;
  color: rgb(var(--color-text-muted));
  font-size: 10px;
  line-height: 1.4;
  padding: 2px 6px;
}

.property-detail-tags span:first-child {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary));
  color: rgb(var(--color-primary-contrast));
}

.property-detail-card h1 {
  margin: 0;
  font-family: var(--font-display);
  font-size: 32px;
  font-weight: 400;
  line-height: 1.16;
}

.property-detail-summary {
  max-width: 600px;
  margin: 8px 0 0;
  color: rgb(var(--color-text-muted));
  font-size: 13px;
  line-height: 1.75;
}

.property-detail-price {
  margin: 12px 0 0;
  color: rgb(var(--color-primary));
  font-family: var(--font-sans);
  font-size: 24px;
  font-weight: 500;
}

.property-detail-price small {
  margin-left: 5px;
  color: rgb(var(--color-text-muted));
  font-size: 12px;
  font-weight: 400;
}

.property-detail-section {
  margin-top: 16px;
  border-top: 1px solid rgb(var(--color-border));
  padding-top: 14px;
}

.property-detail-section h2 {
  margin: 0 0 10px;
  color: rgb(var(--color-text-muted));
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0.16em;
  text-transform: uppercase;
}

.property-detail-section p {
  margin: 0;
  color: rgb(var(--color-text-muted));
  font-size: 13px;
  line-height: 1.75;
}

.property-detail-section p + p {
  margin-top: 8px;
}

.property-link-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 10px;
}

.property-link-row a {
  border: 1px solid rgb(var(--color-primary));
  border-radius: 2px;
  color: rgb(var(--color-primary));
  font-size: 12px;
  font-weight: 600;
  padding: 5px 10px;
  text-decoration: none;
}

.property-spec-grid,
.property-contact-list {
  display: grid;
  gap: 10px;
  margin: 0;
}

.property-spec-grid {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.property-spec-grid dt,
.property-contact-list dt {
  color: rgb(var(--color-text-muted));
  font-size: 10px;
  font-weight: 500;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.property-spec-grid dd,
.property-contact-list dd {
  margin: 3px 0 0;
  font-size: 12px;
  font-weight: 600;
  line-height: 1.4;
  overflow-wrap: anywhere;
}

.property-detail-sidebar {
  display: grid;
  align-content: start;
  gap: 10px;
}

.property-owner {
  display: flex;
  align-items: center;
  gap: 10px;
}

.property-owner h2 {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
}

.property-owner p {
  margin: 3px 0 0;
  color: rgb(var(--color-text-muted));
  font-size: 12px;
}

.property-contact-button {
  display: inline-flex;
  width: 100%;
  min-height: 36px;
  align-items: center;
  justify-content: center;
  gap: 6px;
  margin-top: 14px;
  border: 1px solid rgb(var(--color-primary));
  border-radius: 2px;
  background: rgb(var(--color-primary));
  color: rgb(var(--color-primary-contrast));
  font-size: 12px;
  font-weight: 600;
}

.property-contact-channels {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  margin-top: 12px;
}

.property-contact-channels span {
  border-radius: 2px;
  background: rgb(var(--color-surface-muted));
  color: rgb(var(--color-text-muted));
  font-size: 10px;
  padding: 3px 7px;
}

.property-contact-list {
  margin-top: 12px;
  border-top: 1px solid rgb(var(--color-border));
  padding-top: 12px;
}

.property-contact-list a {
  color: rgb(var(--color-primary));
  text-decoration: none;
}

.property-room-stack {
  display: grid;
  gap: 8px;
}

.property-room-card {
  border: 1px solid rgb(var(--color-border));
  border-radius: 2px;
  padding: 10px;
}

.property-room-card h3,
.property-room-card p {
  margin: 0;
}

.property-room-card p {
  margin-top: 4px;
  font-size: 13px;
  font-weight: 600;
}

.property-room-card span {
  display: block;
  margin-top: 3px;
  color: rgb(var(--color-text-muted));
  font-size: 11px;
}

@media (min-width: 1000px) {
  .property-detail-page {
    grid-template-columns: minmax(0, 1fr) 22rem;
  }

  .property-detail-sidebar {
    position: sticky;
    top: 66px;
  }
}

@media (max-width: 767px) {
  .property-detail-page {
    padding: 18px var(--layout-page-padding-inline) 96px;
  }

  .property-spec-grid {
    grid-template-columns: 1fr;
  }
}
</style>

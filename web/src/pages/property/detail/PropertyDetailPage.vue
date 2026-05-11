<!--
 * 物業頻道詳情頁。
 * 1. 根據頻道讀取樓盤或服務式住宅詳情。
 * 2. 展示圖集、規格、發布者與聯絡方式解鎖。
-->
<script setup lang="ts">
import axios from 'axios';
import { computed, onMounted, ref } from 'vue';
import { useRoute } from 'vue-router';
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

const listingId = computed(() => String(route.params.listingId ?? ''));
const galleryImages = computed(() => listing.value ? resolvePropertyImages(listing.value) : []);
const coverImage = computed(() => galleryImages.value[0]);
const priceText = computed(() =>
  listing.value ? formatPrice(resolvePropertyPrice(listing.value), preferenceStore.locale) : '',
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
  props.channel === 'sale' ? t('property.sale.priceLabel') : t('property.serviced.priceLabel'),
);
const areaLabel = computed(() =>
  props.channel === 'sale' ? t('property.sale.areaLabel') : t('property.serviced.areaLabel'),
);
const formatAreaSqft = (value: number): string => t('common.unit.sqft', { value });
const formatRoomSpec = (area: number, months: number): string => t('common.unit.roomSpec', { area, months });
const contactPayload = computed(() => contactAccess.value?.contact_payload ?? {});
const ownerName = computed(() =>
  listing.value?.owner?.display_name ||
  resolvePropertyPublisherRole(listing.value as PropertyListingDetailResponse),
);

// 1. 讀取詳情
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

// 2. 解鎖聯絡方式
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
</script>

<template>
  <main class="property-detail-page">
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
            <img
              v-for="image in galleryImages.slice(1, 5)"
              :key="image.id"
              :src="image.url"
              :alt="image.alt"
            />
          </div>
        </div>

        <article class="property-detail-card">
          <p class="property-kicker">
            {{ pageTitle }}
          </p>
          <h1>{{ listing.title }}</h1>
          <p class="property-detail-summary">
            {{ listing.summary }}
          </p>
          <p class="property-detail-price">
            {{ priceText }}
          </p>

          <section class="property-detail-section">
            <h2>{{ t('property.detail.description') }}</h2>
            <p>{{ listing.description }}</p>
          </section>

          <section class="property-detail-section">
            <h2>{{ t('property.detail.specification') }}</h2>
            <dl class="property-spec-grid">
              <div>
                <dt>{{ priceLabel }}</dt>
                <dd>{{ priceText }}</dd>
              </div>
              <div>
                <dt>{{ areaLabel }}</dt>
                <dd>{{ formatAreaSqft(resolvePropertyArea(listing)) }}</dd>
              </div>
              <div>
                <dt>{{ t('property.detail.address') }}</dt>
                <dd>{{ listing.property_sale?.address_text || listing.serviced_apartment?.address_text }}</dd>
              </div>
              <div>
                <dt>{{ t('property.list.district') }}</dt>
                <dd>{{ resolvePropertyDistrict(listing, preferenceStore.locale) }}</dd>
              </div>
              <div>
                <dt>{{ t('property.sale.estateLabel') }}</dt>
                <dd>{{ resolvePropertyCommunityName(listing) }}</dd>
              </div>
              <div>
                <dt>{{ t('property.sale.typeLabel') }}</dt>
                <dd>{{ resolvePropertyTypeLabel(listing, preferenceStore.locale) }}</dd>
              </div>
              <div>
                <dt>{{ t('property.sale.typeLabel') }}</dt>
                <dd>{{ resolvePropertyRooms(listing) }}</dd>
              </div>
              <div>
                <dt>{{ t('property.detail.publishedAt') }}</dt>
                <dd>{{ publishedAt }}</dd>
              </div>
            </dl>
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
            v-if="Object.keys(contactPayload).length > 0"
            class="property-contact-list"
          >
            <div
              v-for="(value, key) in contactPayload"
              :key="key"
            >
              <dt>{{ key }}</dt>
              <dd>{{ value }}</dd>
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
              <p>{{ formatPrice(room.monthly_rent_hkd, preferenceStore.locale) }}</p>
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

.property-detail-section {
  margin-top: 1.5rem;
  border-top: 1px solid rgb(var(--color-border));
  padding-top: 1.25rem;
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

.property-spec-grid,
.property-contact-list {
  display: grid;
  gap: 0.85rem;
  margin: 0;
}

.property-spec-grid {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.property-spec-grid dt,
.property-contact-list dt {
  color: rgb(var(--color-text-muted));
  font-size: 0.75rem;
  font-weight: 900;
}

.property-spec-grid dd,
.property-contact-list dd {
  margin: 0.25rem 0 0;
  font-weight: 800;
  line-height: 1.4;
  overflow-wrap: anywhere;
}

.property-detail-sidebar {
  display: grid;
  align-content: start;
  gap: 1rem;
}

.property-owner {
  display: flex;
  align-items: center;
  gap: 0.85rem;
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

.property-contact-button {
  display: inline-flex;
  width: 100%;
  min-height: 2.75rem;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  margin-top: 1rem;
  border: 1px solid rgb(var(--color-primary));
  border-radius: 0.5rem;
  background: rgb(var(--color-primary));
  color: rgb(var(--color-primary-contrast));
  font-weight: 900;
}

.property-contact-list {
  margin-top: 1rem;
  border-top: 1px solid rgb(var(--color-border));
  padding-top: 1rem;
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

@media (min-width: 1000px) {
  .property-detail-page {
    grid-template-columns: minmax(0, 1fr) 22rem;
  }

  .property-detail-sidebar {
    position: sticky;
    top: calc(var(--app-header-offset, 0rem) + 2rem);
  }
}
</style>

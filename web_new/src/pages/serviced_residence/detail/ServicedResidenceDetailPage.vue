<!--
 * 服務式住宅詳情頁。
 * 1. 讀取後端服務式住宅詳情。
 * 2. 展示真實圖片、房型、項目資料、服務與聯絡入口。
 * 3. 僅使用 API 資料展示公開內容與聯絡方式。
-->
<script setup lang="ts">
import axios from 'axios';
import { computed, onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { RouterLink, useRoute, useRouter } from 'vue-router';

import { createOrReusePropertyChat } from '@/httpapis/chats';
import {
  fetchServicedApartmentContactAccess,
  fetchServicedApartmentDetail,
  fetchServicedApartmentListings,
} from '@/httpapis/properties';
import {
  getPropertyOptionLabel,
  servicedRoomCategoryOptions,
} from '@/constants/property';
import type { ContactAccessResult } from '@/model/marketplace';
import type {
  PropertyListingDetailResponse,
  PropertyListingSummaryResponse,
  ServicedApartmentRoomType,
} from '@/model/property';
import { useFeedbackStore } from '@/stores/feedback';
import { usePreferenceStore } from '@/stores/preferences';
import { useSessionStore } from '@/stores/session';
import { formatPrice } from '@/utils/format';
import {
  resolvePropertyArea,
  resolvePropertyDistrict,
  resolvePropertyImages,
  resolvePropertyPriceText,
  resolvePropertyTagLabels,
  resolvePropertyTitle,
} from '@/utils/property';

interface DetailStat {
  value: string;
  label: string;
}

interface ContactRow {
  key: string;
  label: string;
  value: string;
  href: string;
}

const route = useRoute();
const router = useRouter();
const { t } = useI18n();
const feedbackStore = useFeedbackStore();
const preferenceStore = usePreferenceStore();
const sessionStore = useSessionStore();
const listing = ref<PropertyListingDetailResponse | null>(null);
const similarResidences = ref<PropertyListingSummaryResponse[]>([]);
const contactAccess = ref<ContactAccessResult | null>(null);
const loading = ref(false);
const loadingSimilar = ref(false);
const loadingContact = ref(false);
const openingChat = ref(false);
const selectedImageIndex = ref(0);

const listingId = computed(() => String(route.params.listingId ?? ''));
const serviced = computed(() => listing.value?.serviced_apartment ?? null);
const galleryImages = computed(() => listing.value ? resolvePropertyImages(listing.value) : []);
const selectedImage = computed(() => galleryImages.value[selectedImageIndex.value] ?? galleryImages.value[0]);
const contactSummary = computed(() => listing.value?.contact_summary);
const contactPayload = computed(() => contactAccess.value?.contact_payload ?? {});
const canRevealDirectContact = computed(() =>
  Boolean(contactSummary.value?.show_phone || contactSummary.value?.show_whatsapp),
);
const canOpenChat = computed(() => Boolean(contactSummary.value?.show_chat));
const contactRows = computed<ContactRow[]>(() =>
  Object.entries(contactPayload.value)
    .filter(([, value]) => Boolean(value))
    .map(([key, value]) => ({
      key,
      label: resolveContactLabel(key),
      value,
      href: buildContactHref(key, value),
    })),
);
const serviceContentSections = computed(() => {
  if (!serviced.value) {
    return [];
  }

  return [
    {
      title: t('servicedResidence.detail.serviceIntro'),
      body: preferenceStore.locale === 'en'
        ? serviced.value.service_intro_en || serviced.value.service_intro
        : serviced.value.service_intro || serviced.value.service_intro_en,
    },
    {
      title: t('servicedResidence.detail.benefits'),
      body: preferenceStore.locale === 'en'
        ? serviced.value.benefits_text_en || serviced.value.benefits_text
        : serviced.value.benefits_text || serviced.value.benefits_text_en,
    },
    {
      title: t('servicedResidence.detail.extraCharges'),
      body: preferenceStore.locale === 'en'
        ? serviced.value.extra_charges_text_en || serviced.value.extra_charges_text
        : serviced.value.extra_charges_text || serviced.value.extra_charges_text_en,
    },
  ].filter((item) => item.body);
});
const tags = computed(() => {
  if (!listing.value) {
    return [];
  }

  return [
    t('servicedResidence.detail.title'),
    resolvePropertyDistrict(listing.value, preferenceStore.locale),
    ...resolvePropertyTagLabels(listing.value, preferenceStore.locale, 4),
  ].filter(Boolean);
});
const stats = computed<DetailStat[]>(() => {
  if (!listing.value || !serviced.value) {
    return [];
  }

  return [
    {
      value: formatProjectArea(),
      label: t('servicedResidence.detail.areaFrom'),
    },
    {
      value: String(serviced.value.room_types.length),
      label: t('servicedResidence.detail.roomType'),
    },
    {
      value: formatMinimumStay(serviced.value.min_stay_value, serviced.value.min_stay_unit),
      label: t('servicedResidence.detail.minimumStay'),
    },
    {
      value: serviced.value.location_scope === 'overseas'
        ? t('servicedResidence.detail.overseas')
        : t('servicedResidence.detail.hongKong'),
      label: t('servicedResidence.detail.region'),
    },
  ];
});
const facilityLabels = computed(() => {
  if (!listing.value) {
    return [];
  }

  const projectAttributes = serviced.value?.project_attributes ?? {};
  return [
    ...resolvePropertyTagLabels(listing.value, preferenceStore.locale, 20),
    projectAttributes.facility_custom_text,
    projectAttributes.service_custom_text,
  ].filter((item): item is string => Boolean(item));
});
const hasDailyRent = computed(() => Number(serviced.value?.lowest_daily_rent_hkd || 0) > 0);
const hasMonthlyRentRange = computed(() =>
  Number(serviced.value?.highest_monthly_rent_hkd || 0) >
  Number(serviced.value?.lowest_monthly_rent_hkd || 0),
);
const projectName = computed(() => {
  if (!serviced.value) {
    return '';
  }

  return preferenceStore.locale === 'en'
    ? serviced.value.project_name_en || serviced.value.project_name
    : serviced.value.project_name || serviced.value.project_name_en || '';
});
const propertyAddress = computed(() => {
  if (!serviced.value) {
    return '';
  }

  return preferenceStore.locale === 'en'
    ? serviced.value.address_text_en || serviced.value.address_text
    : serviced.value.address_text || serviced.value.address_text_en || '';
});
const residenceDescription = computed(() => {
  if (!listing.value || !serviced.value) {
    return '';
  }

  return preferenceStore.locale === 'en'
    ? serviced.value.description_en || listing.value.description || listing.value.summary
    : listing.value.description || listing.value.summary || serviced.value.description_en || '';
});
const mapSrc = computed(() => {
  const address = propertyAddress.value || resolvePropertyTitle(
    listing.value as PropertyListingSummaryResponse,
    preferenceStore.locale,
  );

  return `https://www.google.com/maps?q=${encodeURIComponent(address)}&output=embed`;
});

// 1. 解析房型分類
const resolveRoomCategory = (category = ''): string => {
  const option = servicedRoomCategoryOptions.find((item) => item.value === category);
  return option
    ? getPropertyOptionLabel(option, preferenceStore.locale)
    : category || t('servicedResidence.detail.roomType');
};

// 2. 取得房型名稱
const resolveRoomName = (room: ServicedApartmentRoomType): string =>
  preferenceStore.locale === 'en' ? room.name_en || room.name : room.name || room.name_en || '';

// 2. 建立電話連結
const buildPhoneHref = (phone: string): string => {
  const normalizedPhone = phone.replace(/\s+/g, '');

  return normalizedPhone ? `tel:${normalizedPhone}` : '';
};

// 2. 建立 WhatsApp 連結
const buildWhatsAppHref = (value: string): string => {
  if (value.startsWith('https://wa.me/')) {
    return value;
  }

  const digits = value.replace(/[+\s\-()]/g, '');

  return digits ? `https://wa.me/${digits}` : '';
};

// 3. 判斷 WhatsApp 欄位
const isWhatsAppContact = (key: string): boolean =>
  key === 'whatsapp_url' || key === 'whatsapp';

// 4. 建立聯絡欄位連結
const buildContactHref = (key: string, value: string): string => {
  if (isWhatsAppContact(key)) {
    return buildWhatsAppHref(value);
  }
  if (key === 'phone' || key === 'phone_2') {
    return buildPhoneHref(value);
  }

  return '';
};

// 5. 輸出聯絡欄位標籤
const resolveContactLabel = (key: string): string => {
  if (key === 'contact_name_zh') {
    return t('servicedResidence.detail.contactNameZh');
  }
  if (key === 'contact_name_en') {
    return t('servicedResidence.detail.contactNameEn');
  }
  if (key === 'phone') {
    return t('servicedResidence.detail.phone');
  }
  if (key === 'phone_2') {
    return t('servicedResidence.detail.phone2');
  }
  if (key === 'wechat') {
    return 'WeChat';
  }
  if (isWhatsAppContact(key)) {
    return 'WhatsApp';
  }

  return key;
};

// 6. 格式化最短入住
const formatMinimumStay = (value?: number, unit?: string): string => {
  const safeValue = Number(value || 1);
  if (unit === 'week') {
    return t('servicedResidence.detail.minimumStayWeek', { value: safeValue });
  }

  return unit === 'day'
    ? t('servicedResidence.detail.minimumStayDay', { value: safeValue })
    : t('servicedResidence.detail.minimumStayMonth', { value: safeValue });
};

// 7. 格式化項目面積
const formatProjectArea = (): string => {
  const minArea = Number(serviced.value?.min_usable_area_sqft || 0);
  const maxArea = Number(serviced.value?.max_usable_area_sqft || 0);
  if (minArea <= 0) {
    return '-';
  }
  if (maxArea > minArea) {
    return `${minArea}-${maxArea}`;
  }

  return String(minArea);
};

// 8. 格式化房型租金
const formatRoomPrice = (room: ServicedApartmentRoomType): string => {
  const dailyMin = Number(room.daily_rent_min_hkd || 0);
  const dailyMax = Number(room.daily_rent_max_hkd || 0);
  const monthlyMin = Number(room.monthly_rent_min_hkd || room.monthly_rent_hkd || 0);
  const monthlyMax = Number(room.monthly_rent_max_hkd || 0);
  const rentUnitLabel = room.rent_unit === 'week'
    ? t('servicedResidence.detail.weekly')
    : t('servicedResidence.detail.monthly');
  const rentSuffix = room.rent_suffix_plus ? '+' : '';

  if (dailyMin > 0) {
    return dailyMax > dailyMin
      ? t('servicedResidence.detail.priceRangePerUnit', {
          min: formatPrice(dailyMin, preferenceStore.locale),
          max: formatPrice(dailyMax, preferenceStore.locale),
          unit: t('servicedResidence.detail.daily'),
        })
      : t('servicedResidence.detail.priceFromPerUnit', {
          price: formatPrice(dailyMin, preferenceStore.locale),
          suffix: rentSuffix,
          unit: t('servicedResidence.detail.daily'),
          from: rentSuffix ? '' : t('servicedResidence.detail.from'),
        });
  }
  if (monthlyMin > 0) {
    return monthlyMax > monthlyMin
      ? t('servicedResidence.detail.priceRangePerUnit', {
          min: formatPrice(monthlyMin, preferenceStore.locale),
          max: formatPrice(monthlyMax, preferenceStore.locale),
          unit: rentUnitLabel,
        })
      : t('servicedResidence.detail.priceFromPerUnit', {
          price: formatPrice(monthlyMin, preferenceStore.locale),
          suffix: rentSuffix,
          unit: rentUnitLabel,
          from: rentSuffix ? '' : t('servicedResidence.detail.from'),
        });
  }

  return serviced.value?.price_negotiable
    ? t('servicedResidence.detail.negotiable')
    : t('servicedResidence.detail.referencePrice');
};

// 9. 格式化房型面積
const formatRoomArea = (room: ServicedApartmentRoomType): string => {
  const minArea = Number(room.usable_area_min_sqft || room.usable_area_sqft || 0);
  const maxArea = Number(room.usable_area_max_sqft || 0);
  if (minArea <= 0) {
    return '-';
  }
  if (maxArea > minArea) {
    return t('servicedResidence.detail.sqftRange', { min: minArea, max: maxArea });
  }

  return t('servicedResidence.detail.sqft', { area: minArea });
};

// 10. 讀取詳情
const loadDetail = async (): Promise<void> => {
  if (!listingId.value) {
    return;
  }

  loading.value = true;
  contactAccess.value = null;

  try {
    const response = await fetchServicedApartmentDetail(listingId.value);
    listing.value = response.data.data;
    selectedImageIndex.value = 0;
    await loadSimilarResidences(response.data.data);
  } catch (error: unknown) {
    feedbackStore.pushToast(
      axios.isAxiosError<{ message?: string }>(error)
        ? error.response?.data?.message ?? t('servicedResidence.detail.loadError')
        : t('servicedResidence.detail.loadError'),
      'error',
    );
    listing.value = null;
    similarResidences.value = [];
  } finally {
    loading.value = false;
  }
};

// 11. 讀取同區推薦
const loadSimilarResidences = async (detail: PropertyListingDetailResponse): Promise<void> => {
  loadingSimilar.value = true;

  try {
    const response = await fetchServicedApartmentListings({
      page: 1,
      page_size: 5,
      district_code: detail.district_code,
    });
    similarResidences.value = response.data.data.items.filter(
      (item) => item.listing_id !== detail.listing_id,
    ).slice(0, 4);
  } catch {
    similarResidences.value = [];
  } finally {
    loadingSimilar.value = false;
  }
};

// 12. 選擇圖片
const selectImage = (index: number): void => {
  selectedImageIndex.value = index;
};

// 13. 解鎖聯絡方式
const revealContact = async (): Promise<void> => {
  if (!listing.value || loadingContact.value || contactRows.value.length > 0) {
    return;
  }
  if (!sessionStore.isAuthenticated) {
    await router.push({
      path: '/login',
      query: { redirect: route.fullPath },
    });
    return;
  }

  loadingContact.value = true;

  try {
    const response = await fetchServicedApartmentContactAccess(listing.value.listing_id);
    contactAccess.value = response.data.data;
    feedbackStore.pushToast(t('servicedResidence.detail.contactUnlockedToast'), 'success');
  } catch (error: unknown) {
    feedbackStore.pushToast(
      axios.isAxiosError<{ message?: string }>(error)
        ? error.response?.data?.message ?? t('servicedResidence.detail.contactError')
        : t('servicedResidence.detail.contactError'),
      'error',
    );
  } finally {
    loadingContact.value = false;
  }
};

// 14. 建立或開啟服務住宅聊天
const openChat = async (): Promise<void> => {
  if (!listing.value || openingChat.value) {
    return;
  }
  if (!sessionStore.isAuthenticated) {
    await router.push({
      path: '/login',
      query: { redirect: route.fullPath },
    });
    return;
  }

  openingChat.value = true;

  try {
    const response = await createOrReusePropertyChat('serviced', listing.value.listing_id);
    await router.push(`/account/chat/${response.data.data.chat_id}`);
  } catch (error: unknown) {
    feedbackStore.pushToast(
      axios.isAxiosError<{ message?: string }>(error)
        ? error.response?.data?.message ?? t('servicedResidence.detail.chatError')
        : t('servicedResidence.detail.chatError'),
      'error',
    );
  } finally {
    openingChat.value = false;
  }
};

// 15. 前往相似住宅
const goSimilar = async (id: string): Promise<void> => {
  await router.push(`/serviced-residences/${id}`);
};

watch(listingId, () => {
  void loadDetail();
});

onMounted(() => {
  void loadDetail();
});
</script>

<template>
  <div class="page-service-detail">
    <nav class="breadcrumb">
      <RouterLink
        class="bc-link"
        to="/"
      >
        {{ t('servicedResidence.detail.home') }}
      </RouterLink>
      <span class="bc-sep">›</span>
      <RouterLink
        class="bc-link"
        to="/serviced-residences"
      >
        {{ t('servicedResidence.detail.title') }}
      </RouterLink>
      <span class="bc-sep">›</span>
      <span class="bc-current">{{ listing
        ? resolvePropertyTitle(listing, preferenceStore.locale)
        : t('servicedResidence.detail.detail') }}</span>
    </nav>

    <section
      v-if="loading"
      class="detail-state"
    >
      {{ t('servicedResidence.detail.loading') }}
    </section>

    <section
      v-else-if="!listing || !serviced"
      class="detail-state"
    >
      {{ t('servicedResidence.detail.loadStateError') }}
    </section>

    <div
      v-else
      class="detail-wrap"
    >
      <div class="detail-main">
        <div class="detail-layout">
          <div class="detail-left">
            <div class="gtags">
              <span
                v-for="(tag, index) in tags"
                :key="tag"
                class="gtag"
                :class="{ dark: index === 0 }"
              >
                {{ tag }}
              </span>
            </div>

            <h2>{{ resolvePropertyTitle(listing, preferenceStore.locale) }}</h2>
            <div class="detail-subtitle">
              {{ projectName }} · {{ resolvePropertyDistrict(listing, preferenceStore.locale) }}
            </div>

            <div class="detail-price-row">
              <div class="detail-price-main">
                {{ resolvePropertyPriceText(listing, preferenceStore.locale) }}
                <span class="detail-price-suffix">{{ hasMonthlyRentRange
                  ? t('servicedResidence.detail.monthlyPrice')
                  : t('servicedResidence.detail.monthlyPriceFrom') }}</span>
              </div>
              <div
                v-if="hasDailyRent"
                class="detail-price-unit"
              >
                {{ resolvePropertyPriceText(listing, preferenceStore.locale, 'daily') }}
                {{ t('servicedResidence.detail.dailyPriceFrom') }}
              </div>
            </div>

            <div class="detail-stats">
              <div
                v-for="stat in stats"
                :key="stat.label"
                class="detail-stat"
              >
                <div class="detail-stat-val">{{ stat.value }}</div>
                <div class="detail-stat-label">{{ stat.label }}</div>
              </div>
            </div>

            <section
              v-if="facilityLabels.length > 0"
              class="detail-section"
            >
              <div class="detail-section-title">{{ t('servicedResidence.detail.facilitiesServices') }}</div>
              <div class="gpills">
                <span
                  v-for="item in facilityLabels"
                  :key="item"
                  class="gpill"
                >
                  {{ item }}
                </span>
              </div>
            </section>

            <section class="detail-section">
              <div class="service-room-tabs">
                <div class="service-room-title">{{ t('servicedResidence.detail.roomType') }}</div>
                <span class="service-room-currency">HKD</span>
              </div>
              <div
                class="service-room-table"
                role="table"
                :aria-label="t('servicedResidence.detail.roomType')"
              >
                <div
                  class="service-room-row service-room-row-head"
                  role="row"
                >
                  <div>{{ t('servicedResidence.detail.roomArea') }}</div>
                  <div>{{ t('servicedResidence.detail.rent') }}</div>
                  <div>{{ t('servicedResidence.detail.minimumStay') }}</div>
                </div>
                <div
                  v-for="room in serviced.room_types"
                  :key="`${room.name}-${room.usable_area_sqft}-${room.min_stay_value}`"
                  class="service-room-row"
                  role="row"
                >
                  <div>
                    <div class="service-room-type">{{ resolveRoomName(room) }}</div>
                    <div class="service-room-area">
                      {{ resolveRoomCategory(room.room_category) }} · {{ formatRoomArea(room) }}
                    </div>
                  </div>
                  <div class="service-room-price">{{ formatRoomPrice(room) }}</div>
                  <div class="service-room-price">
                    {{ formatMinimumStay(room.min_stay_value || room.min_lease_months, room.min_stay_unit) }}
                  </div>
                </div>
              </div>
            </section>

            <section class="detail-section">
              <div class="detail-section-title">{{ t('servicedResidence.detail.residenceDescription') }}</div>
              <p class="body-text">{{ residenceDescription }}</p>
            </section>

            <section
              v-if="serviceContentSections.length > 0"
              class="detail-section"
            >
              <div class="detail-section-title">{{ t('servicedResidence.detail.serviceFacilities') }}</div>
              <div class="detail-copy-grid">
                <article
                  v-for="section in serviceContentSections"
                  :key="section.title"
                >
                  <strong>{{ section.title }}</strong>
                  <p v-if="section.body">{{ section.body }}</p>
                </article>
              </div>
            </section>
          </div>

          <div class="detail-right">
            <div class="detail-right-card">
              <div
                v-if="galleryImages.length > 0"
                class="detail-gallery"
              >
                <img
                  class="detail-main-img"
                  :src="selectedImage.url"
                  :alt="selectedImage.alt"
                />
                <div class="detail-thumb-row">
                  <button
                    v-for="(image, index) in galleryImages.slice(0, 4)"
                    :key="image.id"
                    class="detail-thumb"
                    :class="{ 'detail-thumb--active': selectedImageIndex === index }"
                    type="button"
                    @click="selectImage(index)"
                  >
                    <img
                      :src="image.url"
                      :alt="image.alt"
                    />
                  </button>
                </div>
              </div>
              <div
                v-else
                class="detail-main-img detail-main-img--empty pat"
              >
                AJO Living
              </div>
            </div>

            <div class="detail-right-card">
              <div class="label-text">{{ t('servicedResidence.detail.propertyLocation') }}</div>
              <iframe
                class="building-map-frame"
                :title="t('servicedResidence.detail.mapTitle')"
                :src="mapSrc"
                allowfullscreen
                loading="lazy"
                referrerpolicy="no-referrer-when-downgrade"
              />
              <p class="detail-location-text">{{ propertyAddress }}</p>
            </div>

            <div class="detail-agent-card">
              <div class="detail-agent-profile">
                <div class="detail-agent-avatar">A</div>
                <div class="detail-agent-copy">
                  <div class="detail-agent-kicker">{{ t('servicedResidence.detail.contactKicker') }}</div>
                  <div class="detail-agent-name">
                    {{ serviced.publisher_role_label || listing.owner?.display_name || 'AJO Living' }}
                  </div>
                  <div class="detail-agent-sub">
                    {{ t('servicedResidence.detail.contactHint') }}
                  </div>
                </div>
              </div>

              <div
                v-if="contactRows.length > 0"
                class="detail-contact-list"
              >
                <template
                  v-for="row in contactRows"
                  :key="row.key"
                >
                  <a
                    v-if="row.href"
                    class="detail-contact-row"
                    :href="row.href"
                    :target="isWhatsAppContact(row.key) ? '_blank' : undefined"
                    :rel="isWhatsAppContact(row.key) ? 'noopener' : undefined"
                  >
                    <span>{{ row.label }}</span>
                    <strong>{{ row.label === 'WhatsApp' ? t('servicedResidence.detail.openWhatsApp') : row.value }}</strong>
                  </a>
                  <div
                    v-else
                    class="detail-contact-row"
                  >
                    <span>{{ row.label }}</span>
                    <strong>{{ row.value }}</strong>
                  </div>
                </template>
              </div>

              <div class="detail-agent-actions">
                <button
                  v-if="canRevealDirectContact"
                  class="detail-agent-contact primary"
                  type="button"
                  :disabled="loadingContact"
                  @click="revealContact"
                >
                  {{ contactRows.length > 0
                    ? t('servicedResidence.detail.contactUnlocked')
                    : loadingContact
                      ? t('servicedResidence.detail.unlocking')
                      : t('servicedResidence.detail.unlockContact') }}
                </button>
                <button
                  v-if="canOpenChat"
                  class="detail-agent-contact"
                  type="button"
                  :disabled="openingChat"
                  @click="openChat"
                >
                  {{ openingChat ? t('servicedResidence.detail.opening') : t('servicedResidence.detail.message') }}
                </button>
              </div>
            </div>

            <div
              v-if="similarResidences.length > 0 || loadingSimilar"
              class="similar-section"
            >
              <div class="label-text">{{ t('servicedResidence.detail.similar') }}</div>
              <div class="similar-scroll">
                <button
                  v-for="item in similarResidences"
                  :key="item.listing_id"
                  class="sim-card"
                  type="button"
                  @click="goSimilar(item.listing_id)"
                >
                  <div class="sim-body">
                    <div class="sim-name">{{ resolvePropertyTitle(item, preferenceStore.locale) }}</div>
                    <div class="sim-price">{{ resolvePropertyPriceText(item, preferenceStore.locale) }}</div>
                    <div class="sim-meta">
                      {{ resolvePropertyDistrict(item, preferenceStore.locale) }} ·
                      {{ resolvePropertyArea(item) > 0
                        ? t('servicedResidence.list.sqftFrom', { area: resolvePropertyArea(item) })
                        : t('servicedResidence.detail.servicedFallback') }}
                    </div>
                  </div>
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.page-service-detail {
  display: block;
  width: 100%;
  min-height: calc(100svh - var(--nav-h));
  background: var(--sur);
}

.breadcrumb {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
  width: 100%;
  margin: 0;
  padding: 12px max(24px, calc((100vw - 1180px) / 2 + 24px));
  border: 1px solid var(--bdr);
  border-right: 0;
  border-left: 0;
  background: var(--sur);
  color: var(--ink-3);
  font-family: var(--font);
  font-size: 12px;
  line-height: 1.5;
}

.bc-link {
  border: 0;
  background: transparent;
  color: var(--ink-3);
  cursor: pointer;
  font-family: var(--font);
  font-size: 12px;
  padding: 0;
  text-decoration: none;
}

.bc-link:hover {
  color: var(--brand);
}

.bc-sep {
  color: var(--ink-4);
}

.bc-current {
  color: var(--ink);
  font-weight: 600;
}

.detail-state {
  display: grid;
  min-height: 360px;
  place-items: center;
  color: var(--ink-3);
}

.detail-wrap {
  display: block;
  max-width: var(--layout-page-max-width);
  min-height: auto;
  margin: 0 auto;
  padding: 24px;
}

.detail-main {
  max-width: 100%;
  background: var(--sur);
  padding: 0;
  overflow: visible;
}

.detail-layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
  gap: 16px;
  align-items: start;
  margin-top: 16px;
}

.detail-left,
.detail-right-card,
.detail-agent-card,
.similar-section {
  border: 1px solid var(--bdr);
  border-radius: 8px;
  background: var(--sur);
  padding: 16px;
}

.detail-right-card,
.detail-agent-card,
.similar-section {
  margin-top: 14px;
}

.detail-right-card:first-of-type {
  margin-top: 0;
}

.detail-left {
  position: sticky;
  top: 72px;
}

.gtags {
  display: flex;
  flex-wrap: wrap;
  gap: 5px;
  margin-bottom: 10px;
}

.gtag {
  display: inline-flex;
  align-items: center;
  border: 1px solid var(--bdr);
  border-radius: 999px;
  background: var(--sur);
  color: var(--ink-3);
  font-size: 11px;
  line-height: 1.4;
  padding: 3px 7px;
}

.gtag.dark {
  border-color: var(--brand);
  background: var(--brand);
  color: var(--sur);
  font-weight: 500;
}

.detail-left h2 {
  margin: 0 0 6px;
  font-size: 28px;
  font-weight: 500;
  line-height: 1.25;
  color: var(--ink);
}

.detail-subtitle {
  margin-bottom: 16px;
  font-size: 14px;
  line-height: 1.6;
  color: var(--ink-3);
}

.detail-price-row {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
  margin-bottom: 16px;
}

.detail-price-main {
  font-size: 34px;
  font-weight: 400;
  letter-spacing: 0;
  color: var(--ink);
  line-height: 1.2;
}

.detail-price-suffix {
  font-size: 13px;
  color: var(--ink-3);
  font-weight: 400;
}

.detail-price-unit {
  font-size: 14px;
  color: var(--ink-3);
}

.detail-stats {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
}

.detail-stat {
  min-width: 0;
  border: 1px solid var(--bdr);
  border-radius: 8px;
  background: var(--sur-2);
  padding: 12px 8px;
  text-align: center;
}

.detail-stat-val {
  font-size: 21px;
  font-weight: 600;
  line-height: 1.15;
  color: var(--ink);
}

.detail-stat-label {
  margin-top: 5px;
  font-size: 12px;
  color: var(--ink-3);
}

.detail-section {
  border-top: 1px solid var(--g2);
  padding-top: 16px;
  margin-top: 16px;
}

.detail-section-title {
  font-size: 16px;
  font-weight: 700;
  margin-bottom: 12px;
  color: var(--ink);
}

.label-text {
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--ink-3);
  margin-bottom: 10px;
}

.gpills {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.gpill {
  display: inline-flex;
  align-items: center;
  border: 1px solid var(--bdr);
  border-radius: 999px;
  background: var(--sur-2);
  color: var(--ink);
  font-size: 12px;
  line-height: 1.4;
  padding: 5px 9px;
}

.service-room-tabs {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
}

.service-room-title {
  color: var(--ink);
  font-size: 15px;
  font-weight: 700;
  line-height: 1.3;
}

.service-room-currency {
  border: 1px solid var(--bdr);
  border-radius: 999px;
  background: var(--sur);
  color: var(--ink-3);
  font-size: 11px;
  font-weight: 800;
  padding: 6px 10px;
}

.service-room-table {
  border: 1px solid var(--bdr);
  border-radius: 8px;
  background: var(--sur);
  overflow: hidden;
}

.service-room-row {
  display: grid;
  grid-template-columns: minmax(0, 1.35fr) minmax(150px, 0.95fr) minmax(90px, 0.55fr);
  gap: 12px;
  align-items: center;
  border-top: 1px solid var(--sur-3);
  padding: 12px 14px;
  font-size: 13px;
}

.service-room-row:first-child {
  border-top: 0;
}

.service-room-row-head {
  background: var(--sur-2);
  color: var(--ink-3);
  font-size: 11px;
  font-weight: 800;
}

.service-room-type {
  color: var(--ink);
  font-size: 14px;
  font-weight: 800;
  line-height: 1.45;
}

.service-room-area {
  margin-top: 3px;
  color: var(--ink-3);
  font-size: 12px;
  font-weight: 700;
}

.service-room-price {
  color: var(--ink);
  font-size: 14px;
  font-weight: 800;
}

.body-text {
  margin: 0;
  font-size: 15px;
  line-height: 1.85;
  color: var(--ink);
}

.detail-copy-grid {
  display: grid;
  gap: 12px;
}

.detail-copy-grid article {
  border: 1px solid var(--bdr);
  border-radius: 8px;
  background: var(--sur-2);
  padding: 12px;
}

.detail-copy-grid strong {
  display: block;
  margin-bottom: 6px;
  color: var(--ink);
}

.detail-copy-grid p {
  margin: 0;
  color: var(--ink-3);
  line-height: 1.7;
}

.detail-gallery {
  display: grid;
  gap: 8px;
}

.detail-main-img {
  display: block;
  width: 100%;
  height: 300px;
  border: 0;
  border-radius: 7px;
  object-fit: cover;
  background: var(--sur-2);
}

.detail-main-img--empty {
  color: var(--ink-3);
  font-size: 13px;
  font-weight: 800;
  letter-spacing: 0.1em;
}

.pat {
  position: relative;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
}

.pat::after {
  content: '';
  position: absolute;
  inset: 0;
  background: repeating-linear-gradient(
    45deg,
    transparent,
    transparent 5px,
    rgba(0, 0, 0, 0.025) 5px,
    rgba(0, 0, 0, 0.025) 10px
  );
  pointer-events: none;
}

.detail-thumb-row {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 8px;
}

.detail-thumb {
  height: 86px;
  border-radius: 6px;
  cursor: pointer;
  border: 2px solid transparent;
  overflow: hidden;
  padding: 0;
  background: var(--sur-2);
}

.detail-thumb img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.detail-thumb:hover {
  border-color: var(--brand-mid);
}

.detail-thumb--active {
  border-color: var(--brand);
}

.building-map-frame {
  display: block;
  width: 100%;
  height: 230px;
  border: 0;
  border-radius: 8px;
  background: var(--sur-2);
}

.detail-location-text {
  margin: 10px 0 0;
  color: var(--ink-3);
  font-size: 12px;
  line-height: 1.6;
}

.detail-agent-card {
  display: grid;
  grid-template-columns: 1fr;
  gap: 14px;
  align-items: start;
}

.detail-agent-profile {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
  padding-bottom: 14px;
  border-bottom: 1px solid var(--sur-3);
}

.detail-agent-avatar {
  display: flex;
  width: 46px;
  height: 46px;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  background: var(--brand-light);
  color: var(--brand);
  font-size: 15px;
  font-weight: 800;
  flex: 0 0 auto;
}

.detail-agent-copy {
  min-width: 0;
}

.detail-agent-kicker {
  margin-bottom: 3px;
  color: var(--ink-3);
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 1.2px;
  text-transform: uppercase;
}

.detail-agent-name {
  color: var(--ink);
  font-size: 16px;
  font-weight: 700;
  line-height: 1.25;
}

.detail-agent-sub {
  margin-top: 4px;
  color: var(--ink-3);
  font-size: 12px;
  font-weight: 500;
  line-height: 1.45;
}

.detail-contact-list {
  display: grid;
  gap: 8px;
}

.detail-contact-row {
  display: grid;
  gap: 4px;
  border: 1px solid var(--bdr);
  border-radius: 7px;
  background: var(--sur-2);
  color: var(--ink);
  padding: 10px 12px;
  text-decoration: none;
}

.detail-contact-row span {
  color: var(--ink-3);
  font-size: 11px;
  font-weight: 700;
}

.detail-contact-row strong {
  font-size: 13px;
}

.detail-agent-actions {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px;
}

.detail-agent-contact {
  display: inline-flex;
  min-width: 0;
  min-height: 40px;
  align-items: center;
  justify-content: center;
  gap: 7px;
  border: 1px solid var(--bdr);
  border-radius: 7px;
  background: var(--sur);
  color: var(--ink);
  font-family: inherit;
  font-size: 12px;
  font-weight: 700;
  text-decoration: none;
  white-space: nowrap;
  padding: 0 10px;
  cursor: pointer;
}

.detail-agent-contact:disabled {
  cursor: not-allowed;
  opacity: 0.55;
}

.detail-agent-contact:hover {
  border-color: var(--brand-mid);
  color: var(--brand);
}

.detail-agent-contact.primary {
  border-color: var(--brand);
  background: var(--brand);
  color: var(--sur);
}

.similar-scroll {
  display: grid;
  gap: 8px;
}

.sim-card {
  display: block;
  width: 100%;
  border: 1px solid var(--bdr);
  border-radius: 7px;
  background: var(--sur);
  color: inherit;
  cursor: pointer;
  padding: 12px;
  text-align: left;
}

.sim-name {
  color: var(--ink);
  font-size: 13px;
  font-weight: 700;
  line-height: 1.4;
}

.sim-price {
  margin-top: 4px;
  color: var(--brand);
  font-size: 13px;
  font-weight: 800;
}

.sim-meta {
  margin-top: 3px;
  color: var(--ink-3);
  font-size: 12px;
}

@media (max-width: 1023px) {
  .detail-layout {
    grid-template-columns: 1fr;
  }

  .detail-left {
    position: static;
  }
}

@media (max-width: 640px) {
  .detail-wrap {
    padding: 16px;
  }

  .detail-stats,
  .detail-agent-actions {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .service-room-row {
    grid-template-columns: 1fr;
  }

  .detail-main-img {
    height: 230px;
  }
}
</style>

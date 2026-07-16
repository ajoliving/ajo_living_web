<!--
 * 服務式住宅列表頁。
 * 1. 讀取後端公開服務式住宅列表。
 * 2. 還原服務式住宅高保真展示版面。
 * 3. 僅保留列表、分頁與詳情入口。
-->
<script setup lang="ts">
import axios from 'axios';
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

import { fetchServicedApartmentListings } from '@/httpapis/properties';
import type { PaginationMeta } from '@/model/api';
import type { PropertyListParams, PropertyListingSummaryResponse } from '@/model/property';
import { useFeedbackStore } from '@/stores/feedback';
import { usePreferenceStore } from '@/stores/preferences';
import {
  resolvePropertyArea,
  resolvePropertyCoverImage,
  resolvePropertyDistrict,
  resolvePropertyPriceText,
  resolvePropertyRooms,
  resolvePropertySummary,
  resolvePropertyTagLabels,
  resolvePropertyTitle,
} from '@/utils/property';

const router = useRouter();
const { t } = useI18n();
const feedbackStore = useFeedbackStore();
const preferenceStore = usePreferenceStore();
const loading = ref(false);
const listings = ref<PropertyListingSummaryResponse[]>([]);
const pagination = ref<PaginationMeta>({
  page: 1,
  page_size: 12,
  total: 0,
});

const hasPrevious = computed(() => pagination.value.page > 1);
const hasNext = computed(() => pagination.value.page * pagination.value.page_size < pagination.value.total);
const resultText = computed(() => {
  if (loading.value) {
    return t('servicedResidence.list.loading');
  }
  if (pagination.value.total <= 0) {
    return t('servicedResidence.list.empty');
  }

  const start = (pagination.value.page - 1) * pagination.value.page_size + 1;
  const end = Math.min(pagination.value.page * pagination.value.page_size, pagination.value.total);
  return t('servicedResidence.list.resultRange', {
    start,
    end,
    total: pagination.value.total,
  });
});

// 1. 建立列表查詢參數
const buildQueryParams = (targetPage: number): PropertyListParams => ({
  page: targetPage,
  page_size: pagination.value.page_size,
  sort_by: 'latest',
});

// 2. 讀取服務式住宅列表
const loadListings = async (targetPage = 1): Promise<void> => {
  loading.value = true;

  try {
    const response = await fetchServicedApartmentListings(buildQueryParams(targetPage));
    listings.value = response.data.data.items;
    pagination.value = response.data.data.pagination;
  } catch (error: unknown) {
    feedbackStore.pushToast(
      axios.isAxiosError<{ message?: string }>(error)
        ? error.response?.data?.message ?? t('servicedResidence.list.loadError')
        : t('servicedResidence.list.loadError'),
      'error',
    );
    listings.value = [];
    pagination.value = {
      ...pagination.value,
      page: targetPage,
      total: 0,
    };
  } finally {
    loading.value = false;
  }
};

// 3. 切換頁碼
const loadPage = async (page: number): Promise<void> => {
  if (loading.value || page < 1) {
    return;
  }
  await loadListings(page);
};

// 4. 進入詳情
const openDetail = async (listing: PropertyListingSummaryResponse): Promise<void> => {
  await router.push(`/serviced-residences/${listing.listing_id}`);
};

// 5. 取得卡片封面樣式
const resolveCoverStyle = (listing: PropertyListingSummaryResponse) => {
  const cover = resolvePropertyCoverImage(listing);

  return cover
    ? { backgroundImage: `url("${cover.url}")` }
    : {};
};

// 6. 取得卡片標籤
const resolveCardTags = (listing: PropertyListingSummaryResponse): string[] => [
  resolvePropertyDistrict(listing, preferenceStore.locale),
  ...resolvePropertyTagLabels(listing, preferenceStore.locale, 3),
].filter(Boolean);

// 7. 取得卡片租金模式
const resolveCardPriceMode = (listing: PropertyListingSummaryResponse): 'daily' | 'monthly' =>
  Number(listing.serviced_apartment?.lowest_daily_rent_hkd || 0) > 0 ? 'daily' : 'monthly';

// 8. 取得卡片面積或房型摘要
const resolveCardMeta = (listing: PropertyListingSummaryResponse): string => {
  const area = resolvePropertyArea(listing);
  if (area > 0) {
    return t('servicedResidence.list.sqftFrom', { area });
  }

  const roomCount = listing.serviced_apartment?.room_types.length ?? 0;
  return roomCount > 0
    ? t('servicedResidence.detail.roomTypeCount', { count: roomCount }, roomCount)
    : resolvePropertyRooms(listing, preferenceStore.locale);
};

onMounted(() => {
  void loadListings(1);
});
</script>

<template>
  <div class="sv-page">
    <section class="sv-hero pat">
      <div class="sv-hero-text">
        <div class="hero-eyebrow">{{ t('servicedResidence.list.heroEyebrow') }}</div>
        <h2 class="sv-hero-title">
          {{ t('servicedResidence.list.heroTitleTop') }}<br />{{ t('servicedResidence.list.heroTitleBottom') }}
        </h2>
        <p class="sv-hero-desc">
          {{ t('servicedResidence.list.heroDescription') }}
        </p>
      </div>
    </section>

    <section class="sv-section">
      <div class="sv-section-head">
        <div class="home-sec-title">{{ t('servicedResidence.list.featured') }}</div>
        <div class="sv-result-count">{{ resultText }}</div>
      </div>

      <section
        v-if="loading"
        class="sv-state"
      >
        {{ t('servicedResidence.list.loading') }}
      </section>

      <section
        v-else-if="listings.length === 0"
        class="sv-state"
      >
        {{ t('servicedResidence.list.empty') }}
      </section>

      <section
        v-else
        class="sv-grid"
      >
        <article
          v-for="listing in listings"
          :key="listing.listing_id"
          class="sv-card"
          @click="openDetail(listing)"
        >
          <div
            class="sv-img pat"
            :class="{ 'sv-img--empty': !resolvePropertyCoverImage(listing) }"
            :style="resolveCoverStyle(listing)"
          >
            <span v-if="!resolvePropertyCoverImage(listing)">AJO Living</span>
          </div>
          <div class="sv-body">
            <div class="gtags">
              <span class="gtag dark">
                {{ listing.serviced_apartment?.min_stay_unit === 'day'
                  ? t('servicedResidence.list.shortStay')
                  : t('servicedResidence.list.monthly') }}
              </span>
              <span
                v-for="tag in resolveCardTags(listing)"
                :key="tag"
                class="gtag"
              >
                {{ tag }}
              </span>
            </div>
            <div class="gtitle">{{ resolvePropertyTitle(listing, preferenceStore.locale) }}</div>
            <div class="gsub">{{ resolvePropertySummary(listing, preferenceStore.locale) }}</div>
            <div class="gprice">
              {{ resolvePropertyPriceText(listing, preferenceStore.locale, resolveCardPriceMode(listing)) }}
              <span>{{ resolveCardPriceMode(listing) === 'daily'
                ? t('servicedResidence.list.perDayFrom')
                : t('servicedResidence.list.perMonthFrom') }}</span>
            </div>
            <div class="gpills">
              <span
                v-for="tag in resolvePropertyTagLabels(listing, preferenceStore.locale, 5)"
                :key="tag"
                class="gpill"
              >
                {{ tag }}
              </span>
            </div>
          </div>
          <div class="gfoot">
            <span class="grooms">
              {{ resolveCardMeta(listing) }}
            </span>
            <span class="gview">{{ t('servicedResidence.list.view') }} →</span>
          </div>
        </article>
      </section>

      <div class="sv-pagination">
        <button
          class="sv-button sv-button--secondary"
          type="button"
          :disabled="!hasPrevious || loading"
          @click="loadPage(pagination.page - 1)"
        >
          {{ t('servicedResidence.list.previous') }}
        </button>
        <span>{{ pagination.page }}</span>
        <button
          class="sv-button sv-button--secondary"
          type="button"
          :disabled="!hasNext || loading"
          @click="loadPage(pagination.page + 1)"
        >
          {{ t('servicedResidence.list.next') }}
        </button>
      </div>
    </section>
  </div>
</template>

<style scoped>
.sv-page {
  --brand: #F05A00;
  --brand-dark: #C04600;
  --brand-light: #FFF0E6;
  --brand-mid: #FDA96A;
  --ink: #1A1A1A;
  --ink-2: #444444;
  --ink-3: #777777;
  --ink-4: #AAAAAA;
  --sur: #FFFFFF;
  --sur-2: #F6F6F6;
  --sur-3: #EFEFEF;
  --bdr: #E4E4E4;
  --bdr-2: #CCCCCC;
  --font: 'DM Sans', 'Noto Sans TC', sans-serif;
  --font-serif: 'DM Serif Display', serif;
  --g1: var(--sur-2);
  --g2: var(--bdr);
  --g3: var(--ink-4);
  --g4: var(--ink-3);
  --g5: var(--ink-2);
  --white: var(--sur);
  --accent: var(--brand);
  --accent-light: var(--brand-light);
  --accent-dark: var(--brand-dark);
  --nav-h: 52px;

  width: 100%;
  min-height: calc(100svh - var(--nav-h));
  background: var(--sur-2);
  font-family: var(--font);
  font-size: 13px;
  color: var(--ink);
}

.sv-hero {
  min-height: 200px;
  display: flex;
  align-items: flex-end;
  padding: 40px 40px 32px;
  position: relative;
  overflow: hidden;
  max-width: 1440px;
  margin: 0 auto;
  background: linear-gradient(160deg, #1a1a1a, #2e2e2e);
}

.sv-hero-text {
  position: relative;
  z-index: 1;
}

.hero-eyebrow {
  font-size: 10px;
  letter-spacing: 3px;
  color: var(--g3);
  margin-bottom: 14px;
}

.sv-hero-title {
  font-family: var(--font-serif);
  font-size: 36px;
  color: white;
  font-weight: 400;
  line-height: 1.2;
  margin: 0 0 12px;
}

.sv-hero-desc {
  font-size: 13px;
  color: var(--g3);
  line-height: 1.8;
  max-width: 400px;
  margin: 0;
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

.sv-section {
  padding: 32px 40px;
  max-width: 1440px;
  margin: 0 auto;
}

.sv-button {
  display: inline-flex;
  min-height: 38px;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  cursor: pointer;
  font: inherit;
  font-weight: 800;
  padding: 0 14px;
}

.sv-button:disabled {
  cursor: not-allowed;
  opacity: 0.55;
}

.sv-button--primary {
  border: 1px solid var(--accent);
  background: var(--accent);
  color: var(--white);
}

.sv-button--secondary {
  border: 1px solid var(--g2);
  background: var(--white);
  color: var(--ink);
}

.sv-section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 16px;
}

.home-sec-title {
  font-size: 13px;
  font-weight: 500;
  letter-spacing: 0.5px;
  border-left: 3px solid var(--brand);
  padding-left: 10px;
}

.sv-result-count {
  color: var(--g4);
  font-size: 12px;
}

.sv-state {
  display: grid;
  min-height: 220px;
  place-items: center;
  border: 1px solid var(--g2);
  border-radius: 6px;
  background: var(--white);
  color: var(--g4);
}

.sv-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 14px;
}

.sv-card {
  border: 1px solid var(--g2);
  border-radius: 3px;
  overflow: hidden;
  background: var(--white);
  cursor: pointer;
}

.sv-card:hover {
  box-shadow: 0 3px 14px rgba(0, 0, 0, 0.07);
}

.sv-img {
  width: 100%;
  height: 140px;
  background-size: cover;
  background-position: center;
}

.sv-img--empty {
  background: linear-gradient(160deg, #dde4e8, #b8c8d0);
  color: rgba(26, 26, 26, 0.45);
  font-size: 12px;
  font-weight: 800;
  letter-spacing: 0.08em;
}

.sv-body {
  padding: 12px 14px;
}

.gtags {
  display: flex;
  gap: 5px;
  margin-bottom: 6px;
  flex-wrap: wrap;
}

.gtag {
  font-size: 11px;
  letter-spacing: 0.8px;
  color: var(--g4);
  border: 1px solid var(--g2);
  padding: 3px 7px;
  border-radius: 1px;
}

.gtag.dark {
  background: var(--accent);
  color: var(--white);
  border-color: var(--accent);
  font-weight: 500;
}

.gtitle {
  font-size: 14px;
  font-weight: 700;
  margin-bottom: 5px;
}

.gsub {
  min-height: 34px;
  font-size: 12px;
  color: var(--g4);
  line-height: 1.45;
  margin-bottom: 8px;
}

.gprice {
  font-size: 17px;
  font-weight: 500;
  margin-top: 6px;
}

.gprice span {
  font-size: 11px;
  color: var(--g4);
  font-weight: 400;
}

.gpills {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 10px;
}

.gpill {
  font-size: 10px;
  padding: 3px 7px;
  background: var(--g1);
  color: var(--g5);
  border-radius: 2px;
}

.gfoot {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 13px;
  border-top: 1px solid var(--g1);
  font-size: 11px;
}

.grooms {
  color: var(--g4);
}

.gview {
  color: var(--accent);
  font-weight: 700;
}

.sv-pagination {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 10px;
  margin-top: 18px;
}

@media (min-width: 1440px) {
  .sv-hero,
  .sv-section {
    max-width: 1480px;
  }
}

@media (max-width: 1024px) {
  .sv-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 767px) {
  .sv-hero {
    min-height: 160px;
    padding: 28px 20px 24px;
  }

  .sv-hero-title {
    font-size: 28px;
  }

  .sv-section {
    padding: 20px 20px 36px;
  }

  .sv-grid {
    grid-template-columns: 1fr;
  }

  .sv-section-head {
    align-items: stretch;
    flex-direction: column;
  }
}
</style>

<!--
 * 物業頻道列表頁。
 * 1. 根據頻道展示樓盤放售或服務式住宅列表。
 * 2. 提供關鍵字、地區、排序與發布入口。
-->
<script setup lang="ts">
import axios from 'axios';
import { computed, onMounted, ref, watch } from 'vue';
import { RouterLink } from 'vue-router';
import { useI18n } from 'vue-i18n';

import {
  fetchPropertySaleListings,
  fetchServicedApartmentListings,
} from '@/httpapis/properties';
import {
  getPropertyOptionLabel,
  marketplaceDistricts,
} from '@/constants/property';
import type {
  PropertyChannel,
  PropertyListParams,
  PropertyListingSummaryResponse,
} from '@/model/property';
import AppIcon from '@/shared/components/base/AppIcon.vue';
import { useFeedbackStore } from '@/stores/feedback';
import { usePreferenceStore } from '@/stores/preferences';
import { formatPrice } from '@/utils/format';
import {
  resolvePropertyArea,
  resolvePropertyCommunityName,
  resolvePropertyCoverImage,
  resolvePropertyDetailPath,
  resolvePropertyDistrict,
  resolvePropertyPrice,
  resolvePropertyPublisherRole,
  resolvePropertyRooms,
  resolvePropertySummary,
  resolvePropertyTitle,
  resolvePropertyTypeLabel,
} from '@/utils/property';

const props = defineProps<{
  channel: PropertyChannel;
}>();

const { t } = useI18n();
const feedbackStore = useFeedbackStore();
const preferenceStore = usePreferenceStore();

const loading = ref(false);
const keyword = ref('');
const districtCode = ref('');
const sortBy = ref<PropertyListParams['sort_by']>('latest');
const items = ref<PropertyListingSummaryResponse[]>([]);

const pageTitle = computed(() =>
  props.channel === 'sale' ? t('property.sale.title') : t('property.serviced.title'),
);
const pageSubtitle = computed(() =>
  props.channel === 'sale' ? t('property.sale.subtitle') : t('property.serviced.subtitle'),
);
const publishPath = computed(() =>
  props.channel === 'sale' ? '/properties/my/new' : '/serviced-residences/my/new',
);
const minePath = computed(() =>
  props.channel === 'sale' ? '/properties/my' : '/serviced-residences/my',
);
const priceLabel = computed(() =>
  props.channel === 'sale' ? t('property.sale.priceLabel') : t('property.serviced.priceLabel'),
);
const areaLabel = computed(() =>
  props.channel === 'sale' ? t('property.sale.areaLabel') : t('property.serviced.areaLabel'),
);
const formatAreaSqft = (value: number): string => t('common.unit.sqft', { value });

// 1. 讀取列表資料
const loadListings = async (): Promise<void> => {
  loading.value = true;

  try {
    const params: PropertyListParams = {
      page: 1,
      page_size: 30,
      keyword: keyword.value.trim(),
      district_code: districtCode.value,
      sort_by: sortBy.value,
    };
    const response = props.channel === 'sale'
      ? await fetchPropertySaleListings(params)
      : await fetchServicedApartmentListings(params);

    items.value = response.data.data.items;
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

// 2. 格式化價格
const formatListingPrice = (listing: PropertyListingSummaryResponse): string =>
  formatPrice(resolvePropertyPrice(listing), preferenceStore.locale);

watch([districtCode, sortBy], () => {
  void loadListings();
});

onMounted(() => {
  void loadListings();
});
</script>

<template>
  <main class="property-list-page">
    <section class="property-list-hero">
      <div>
        <p class="property-kicker">
          AJO Living
        </p>
        <h1>{{ pageTitle }}</h1>
        <p>{{ pageSubtitle }}</p>
      </div>
      <div class="property-hero-actions">
        <RouterLink
          :to="minePath"
          class="property-button property-button--secondary"
        >
          <AppIcon
            name="user"
            :size="17"
          />
          {{ t('property.list.manageMine') }}
        </RouterLink>
        <RouterLink
          :to="publishPath"
          class="property-button property-button--primary"
        >
          <AppIcon
            name="plus-square"
            :size="17"
          />
          {{ t('property.list.publish') }}
        </RouterLink>
      </div>
    </section>

    <section class="property-filter-bar">
      <label class="property-field property-field--search">
        <span>{{ t('property.list.keyword') }}</span>
        <div class="property-search-input">
          <AppIcon
            name="search"
            :size="17"
          />
          <input
            v-model="keyword"
            type="search"
            :placeholder="t('property.list.keywordPlaceholder')"
            @keyup.enter="loadListings"
          />
        </div>
      </label>

      <label class="property-field">
        <span>{{ t('property.list.district') }}</span>
        <select v-model="districtCode">
          <option value="">
            {{ t('property.list.allDistricts') }}
          </option>
          <option
            v-for="district in marketplaceDistricts"
            :key="district.value"
            :value="district.value"
          >
            {{ getPropertyOptionLabel(district, preferenceStore.locale) }}
          </option>
        </select>
      </label>

      <label class="property-field">
        <span>{{ t('property.list.sort') }}</span>
        <select v-model="sortBy">
          <option value="latest">
            {{ t('property.list.latest') }}
          </option>
          <option value="price_asc">
            {{ t('property.list.priceAsc') }}
          </option>
          <option value="price_desc">
            {{ t('property.list.priceDesc') }}
          </option>
        </select>
      </label>

      <button
        type="button"
        class="property-button property-button--primary"
        @click="loadListings"
      >
        <AppIcon
          name="search"
          :size="17"
        />
        {{ t('common.action.browse') }}
      </button>
    </section>

    <div class="property-results-line">
      {{ t('property.list.results', { count: items.length }) }}
    </div>

    <section
      v-if="loading"
      class="property-empty"
    >
      {{ t('common.status.loading') }}
    </section>

    <section
      v-else-if="items.length === 0"
      class="property-empty"
    >
      {{ t('property.list.noListings') }}
    </section>

    <section
      v-else
      class="property-grid"
    >
      <RouterLink
        v-for="listing in items"
        :key="listing.listing_id"
        :to="resolvePropertyDetailPath(listing)"
        class="property-card group"
      >
        <div class="property-card__media">
          <img
            v-if="resolvePropertyCoverImage(listing)"
            :src="resolvePropertyCoverImage(listing)?.url"
            :alt="resolvePropertyTitle(listing)"
          />
          <div
            v-else
            class="property-card__placeholder"
          >
            <AppIcon
              name="picture"
              :size="42"
            />
          </div>
        </div>

        <div class="property-card__body">
          <div class="property-card__meta">
            <span>{{ resolvePropertyDistrict(listing, preferenceStore.locale) }}</span>
            <span>{{ resolvePropertyPublisherRole(listing) }}</span>
          </div>
          <h2>{{ resolvePropertyTitle(listing) }}</h2>
          <p>{{ resolvePropertySummary(listing) }}</p>
          <dl class="property-card__specs">
            <div>
              <dt>{{ priceLabel }}</dt>
              <dd>{{ formatListingPrice(listing) }}</dd>
            </div>
            <div>
              <dt>{{ areaLabel }}</dt>
              <dd>{{ formatAreaSqft(resolvePropertyArea(listing)) }}</dd>
            </div>
            <div>
              <dt>{{ t('property.sale.estateLabel') }}</dt>
              <dd>{{ resolvePropertyCommunityName(listing) }}</dd>
            </div>
            <div>
              <dt>{{ t('property.sale.typeLabel') }}</dt>
              <dd>{{ resolvePropertyTypeLabel(listing, preferenceStore.locale) }}</dd>
            </div>
          </dl>
          <div class="property-card__footer">
            <span>{{ resolvePropertyRooms(listing) }}</span>
            <span>{{ t('common.action.viewDetail') }}</span>
          </div>
        </div>
      </RouterLink>
    </section>
  </main>
</template>

<style scoped>
.property-list-page {
  width: 100%;
  max-width: var(--layout-page-max-width);
  margin: 0 auto;
  padding: 3rem var(--layout-page-padding-inline) 5rem;
  color: rgb(var(--color-text));
}

.property-list-hero {
  display: grid;
  gap: 1.5rem;
  align-items: end;
  margin-bottom: 1.5rem;
}

.property-kicker {
  margin: 0 0 0.6rem;
  color: rgb(var(--color-primary));
  font-size: 0.75rem;
  font-weight: 800;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.property-list-hero h1 {
  margin: 0;
  font-family: var(--font-display);
  font-size: clamp(2.25rem, 5vw, 4.5rem);
  font-weight: 600;
  line-height: 1.05;
}

.property-list-hero p:last-child {
  max-width: 42rem;
  margin: 1rem 0 0;
  color: rgb(var(--color-text-muted));
  font-size: 1rem;
  line-height: 1.7;
}

.property-hero-actions,
.property-filter-bar {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
}

.property-button {
  display: inline-flex;
  min-height: 2.75rem;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  border-radius: 0.5rem;
  padding: 0 1rem;
  font-size: 0.9rem;
  font-weight: 800;
}

.property-button--primary {
  border: 1px solid rgb(var(--color-primary));
  background: rgb(var(--color-primary));
  color: rgb(var(--color-primary-contrast));
}

.property-button--secondary {
  border: 1px solid rgb(var(--color-border));
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text));
}

.property-filter-bar {
  align-items: end;
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.75rem;
  background: rgb(var(--color-surface));
  padding: 1rem;
}

.property-field {
  display: grid;
  min-width: 12rem;
  gap: 0.45rem;
  color: rgb(var(--color-text-muted));
  font-size: 0.78rem;
  font-weight: 800;
}

.property-field--search {
  flex: 1 1 18rem;
}

.property-field select,
.property-search-input {
  min-height: 2.75rem;
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.5rem;
  background: rgb(var(--color-surface-raised));
  color: rgb(var(--color-text));
}

.property-field select {
  padding: 0 0.75rem;
}

.property-search-input {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0 0.75rem;
}

.property-search-input input {
  width: 100%;
  border: 0;
  background: transparent;
  color: inherit;
  outline: 0;
}

.property-results-line {
  margin: 1.25rem 0;
  color: rgb(var(--color-text-muted));
  font-size: 0.9rem;
  font-weight: 700;
}

.property-grid {
  display: grid;
  gap: 1rem;
}

.property-card {
  display: grid;
  overflow: hidden;
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.75rem;
  background: rgb(var(--color-surface));
  color: inherit;
  box-shadow: 0 10px 34px rgb(15 23 42 / 0.07);
}

.property-card__media {
  aspect-ratio: 4 / 3;
  background: rgb(var(--color-surface-raised));
}

.property-card__media img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 0.35s ease;
}

.property-card:hover .property-card__media img {
  transform: scale(1.03);
}

.property-card__placeholder {
  display: grid;
  height: 100%;
  place-items: center;
  color: rgb(var(--color-text-muted));
}

.property-card__body {
  display: grid;
  gap: 0.9rem;
  padding: 1rem;
}

.property-card__meta,
.property-card__footer {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  color: rgb(var(--color-text-muted));
  font-size: 0.78rem;
  font-weight: 800;
}

.property-card h2 {
  margin: 0;
  font-size: 1.15rem;
  font-weight: 800;
  line-height: 1.35;
}

.property-card p {
  display: -webkit-box;
  min-height: 3rem;
  margin: 0;
  overflow: hidden;
  color: rgb(var(--color-text-muted));
  font-size: 0.9rem;
  line-height: 1.65;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.property-card__specs {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.75rem;
  margin: 0;
}

.property-card__specs dt {
  color: rgb(var(--color-text-muted));
  font-size: 0.72rem;
  font-weight: 800;
}

.property-card__specs dd {
  margin: 0.2rem 0 0;
  font-size: 0.92rem;
  font-weight: 800;
  line-height: 1.35;
}

.property-card__footer {
  justify-content: space-between;
  border-top: 1px solid rgb(var(--color-border));
  padding-top: 0.85rem;
}

.property-empty {
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.75rem;
  background: rgb(var(--color-surface));
  padding: 2rem;
  color: rgb(var(--color-text-muted));
  text-align: center;
}

@media (min-width: 760px) {
  .property-list-hero {
    grid-template-columns: minmax(0, 1fr) auto;
  }

  .property-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (min-width: 1120px) {
  .property-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}
</style>

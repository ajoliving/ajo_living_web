<!--
 * 物業頻道列表頁。
 * 1. 依照桌面 HTML 參考還原樓盤租售與服務式住宅列表。
 * 2. 接入公開樓盤與服務住宅 API 查詢、篩選、排序與分頁。
 * 3. 保持樓盤價格、租金與詳情入口對應真實資料。
-->
<script setup lang="ts">
import axios from 'axios';
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { RouterLink } from 'vue-router';
import { useI18n } from 'vue-i18n';

import {
  propertyAreaRangeFilterOptions,
  propertyBedroomFilterOptions,
  propertyPriceRangeFilterOptions,
  propertyPublisherFilterOptions,
  propertyRegionFilterOptions,
  propertyRenovationFilterOptions,
  propertyTransactionTypeFilterOptions,
  propertyTypeFilterOptions,
  type PropertyFilterOption,
  type PropertyRangeOption,
} from '@/constants/property';
import {
  fetchPropertySaleListings,
  fetchServicedApartmentListings,
} from '@/httpapis/properties';
import type {
  PropertyChannel,
  PropertyListParams,
  PropertyListingSummaryResponse,
  PropertyTransactionType,
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

type PropertySortOption = NonNullable<PropertyListParams['sort_by']>;
type ListingFilterOption = PropertyFilterOption | PropertyRangeOption;

interface FilterGroup {
  key: string;
  title: string;
  options: ListingFilterOption[];
  activeValue: string;
  onSelect: (value: string) => void;
}

const { t } = useI18n();
const feedbackStore = useFeedbackStore();
const preferenceStore = usePreferenceStore();
const pageSize = ref(24);
const page = ref(1);
const total = ref(0);
const loading = ref(false);
const keyword = ref('');
const selectedRegion = ref('');
const selectedTransactionType = ref<PropertyTransactionType | ''>('sale');
const selectedPropertyType = ref('');
const selectedPriceRange = ref('');
const selectedAreaRange = ref('');
const selectedBedroom = ref('');
const selectedRenovationTag = ref('');
const publisherIdentityType = ref('');
const sortBy = ref<PropertySortOption>('latest');
const items = ref<PropertyListingSummaryResponse[]>([]);
const filterReloadTimer = ref<ReturnType<typeof window.setTimeout> | null>(null);

const propertyRentRangeFilterOptions: PropertyRangeOption[] = [
  { value: '', label: '不限' },
  { value: 'under_10k', label: '1萬以下', max: 10_000 },
  { value: '10k_20k', label: '1萬-2萬', min: 10_000, max: 20_000 },
  { value: '20k_40k', label: '2萬-4萬', min: 20_000, max: 40_000 },
  { value: 'over_40k', label: '4萬+', min: 40_000 },
];

const isSale = computed(() => props.channel === 'sale');
const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize.value)));
const resultLabel = computed(() => t('property.list.results', { count: total.value }));
const activePriceRangeOptions = computed<PropertyRangeOption[]>(() =>
  selectedTransactionType.value === 'rent'
    ? propertyRentRangeFilterOptions
    : propertyPriceRangeFilterOptions,
);
const selectedPriceRangeOption = computed(() =>
  activePriceRangeOptions.value.find((option) => option.value === selectedPriceRange.value),
);
const selectedAreaRangeOption = computed(() =>
  propertyAreaRangeFilterOptions.find((option) => option.value === selectedAreaRange.value),
);
const activeFilterCount = computed(() =>
  [
    keyword.value.trim(),
    selectedRegion.value,
    selectedTransactionType.value,
    selectedPropertyType.value,
    selectedPriceRange.value,
    selectedAreaRange.value,
    selectedBedroom.value,
    selectedRenovationTag.value,
    publisherIdentityType.value,
  ].filter(Boolean).length,
);
const pageNumbers = computed(() => {
  const current = Math.min(totalPages.value, Math.max(1, page.value));
  const start = Math.max(1, Math.min(current - 2, totalPages.value - 4));
  const end = Math.min(totalPages.value, start + 4);

  return Array.from({ length: end - start + 1 }, (_, index) => start + index);
});
const saleFilterGroups = computed<FilterGroup[]>(() => [
  {
    key: 'region',
    title: '地區',
    options: propertyRegionFilterOptions,
    activeValue: selectedRegion.value,
    onSelect: (value) => {
      selectedRegion.value = value;
    },
  },
  {
    key: 'transactionType',
    title: '性質',
    options: propertyTransactionTypeFilterOptions,
    activeValue: selectedTransactionType.value,
    onSelect: (value) => {
      selectedTransactionType.value = value as PropertyTransactionType | '';
      selectedPriceRange.value = '';
    },
  },
  {
    key: 'propertyType',
    title: '類型',
    options: propertyTypeFilterOptions,
    activeValue: selectedPropertyType.value,
    onSelect: (value) => {
      selectedPropertyType.value = value;
    },
  },
  {
    key: 'priceRange',
    title: selectedTransactionType.value === 'rent' ? '租金範圍' : '售價範圍',
    options: activePriceRangeOptions.value,
    activeValue: selectedPriceRange.value,
    onSelect: (value) => {
      selectedPriceRange.value = value;
    },
  },
  {
    key: 'areaRange',
    title: '實用面積',
    options: propertyAreaRangeFilterOptions,
    activeValue: selectedAreaRange.value,
    onSelect: (value) => {
      selectedAreaRange.value = value;
    },
  },
  {
    key: 'bedroom',
    title: '房間',
    options: propertyBedroomFilterOptions,
    activeValue: selectedBedroom.value,
    onSelect: (value) => {
      selectedBedroom.value = value;
    },
  },
  {
    key: 'renovation',
    title: '裝修',
    options: propertyRenovationFilterOptions,
    activeValue: selectedRenovationTag.value,
    onSelect: (value) => {
      selectedRenovationTag.value = value;
    },
  },
  {
    key: 'publisher',
    title: '業主類型',
    options: propertyPublisherFilterOptions,
    activeValue: publisherIdentityType.value,
    onSelect: (value) => {
      publisherIdentityType.value = value;
    },
  },
]);
const serviceHeroTitle = computed(() =>
  preferenceStore.locale === 'zh-HK' ? '靈活短租\n全城精選' : 'Flexible Stays\nCitywide Picks',
);
const serviceHeroDescription = computed(() =>
  preferenceStore.locale === 'zh-HK'
    ? '酒店式管理，家的感覺。按日、按週、按月靈活租用，適合商務出行及過渡期居住。'
    : 'Hotel-style management with the comfort of home. Daily, weekly, and monthly stays for business travel and transition periods.',
);

// 1. 建立列表查詢參數
const buildListParams = (): PropertyListParams => {
  const params: PropertyListParams = {
    page: page.value,
    page_size: pageSize.value,
    sort_by: sortBy.value,
  };
  const normalizedKeyword = keyword.value.trim();
  if (normalizedKeyword) {
    params.keyword = normalizedKeyword;
  }
  if (!isSale.value) {
    return params;
  }
  if (selectedRegion.value) {
    params.region_code = selectedRegion.value;
  }
  if (selectedTransactionType.value) {
    params.transaction_type = selectedTransactionType.value;
  }
  if (selectedPropertyType.value) {
    params.property_type = selectedPropertyType.value;
  }
  if (selectedPriceRangeOption.value?.min !== undefined) {
    params.min_price_hkd = selectedPriceRangeOption.value.min;
  }
  if (selectedPriceRangeOption.value?.max !== undefined) {
    params.max_price_hkd = selectedPriceRangeOption.value.max;
  }
  if (selectedAreaRangeOption.value?.min !== undefined) {
    params.min_area_sqft = selectedAreaRangeOption.value.min;
  }
  if (selectedAreaRangeOption.value?.max !== undefined) {
    params.max_area_sqft = selectedAreaRangeOption.value.max;
  }
  if (selectedBedroom.value) {
    params.bedroom_count = Number(selectedBedroom.value);
  }
  if (selectedRenovationTag.value) {
    params.feature_tags = selectedRenovationTag.value;
  }
  if (publisherIdentityType.value) {
    params.publisher_identity_type = publisherIdentityType.value;
  }

  return params;
};

// 2. 清除延遲查詢計時器
const clearFilterReloadTimer = (): void => {
  if (filterReloadTimer.value) {
    window.clearTimeout(filterReloadTimer.value);
    filterReloadTimer.value = null;
  }
};

// 3. 延遲重新查詢列表
const scheduleListingsReload = (): void => {
  page.value = 1;
  clearFilterReloadTimer();
  filterReloadTimer.value = window.setTimeout(() => {
    void loadListings();
  }, 120);
};

// 4. 讀取列表資料
const loadListings = async (): Promise<void> => {
  clearFilterReloadTimer();
  loading.value = true;

  try {
    const response = isSale.value
      ? await fetchPropertySaleListings(buildListParams())
      : await fetchServicedApartmentListings(buildListParams());
    const result = response.data.data;

    items.value = result.items;
    page.value = result.pagination.page;
    pageSize.value = result.pagination.page_size;
    total.value = result.pagination.total;
  } catch (error: unknown) {
    feedbackStore.pushToast(
      axios.isAxiosError<{ message?: string }>(error)
        ? error.response?.data?.message ?? t('property.detail.loadError')
        : t('property.detail.loadError'),
      'error',
    );
    items.value = [];
    total.value = 0;
  } finally {
    loading.value = false;
  }
};

// 5. 提交搜尋
const runSearch = (): void => {
  page.value = 1;
  void loadListings();
};

// 6. 重設篩選
const resetFilters = (): void => {
  keyword.value = '';
  selectedRegion.value = '';
  selectedTransactionType.value = isSale.value ? 'sale' : '';
  selectedPropertyType.value = '';
  selectedPriceRange.value = '';
  selectedAreaRange.value = '';
  selectedBedroom.value = '';
  selectedRenovationTag.value = '';
  publisherIdentityType.value = '';
  sortBy.value = 'latest';
  page.value = 1;
  void loadListings();
};

// 7. 切換分頁
const goToPage = (targetPage: number): void => {
  const nextPage = Math.min(totalPages.value, Math.max(1, targetPage));
  if (nextPage === page.value || loading.value) {
    return;
  }
  page.value = nextPage;
  void loadListings();
};

// 8. 格式化價格
const formatListingPrice = (listing: PropertyListingSummaryResponse): string => {
  const sale = listing.property_sale;
  if (sale?.price_negotiable) {
    return '面議';
  }
  const price = sale?.transaction_type === 'rent'
    ? sale.monthly_rent_hkd ?? 0
    : resolvePropertyPrice(listing);
  if (price <= 0) {
    return '待定';
  }

  return sale?.price_reference_only
    ? `${formatPrice(price, preferenceStore.locale)} 起`
    : formatPrice(price, preferenceStore.locale);
};

// 9. 取得價格單位
const resolvePriceUnit = (listing: PropertyListingSummaryResponse): string => {
  if (listing.serviced_apartment) {
    return '/ 月起';
  }
  return listing.property_sale?.transaction_type === 'rent' ? '/ 月' : '';
};

// 10. 取得樓盤地點文字
const resolveLocationText = (listing: PropertyListingSummaryResponse): string =>
  listing.property_sale?.public_location_text?.trim() ||
  listing.property_sale?.address_text?.trim() ||
  resolvePropertyDistrict(listing, preferenceStore.locale);

// 11. 取得樓盤編號
const resolveListingNumber = (listing: PropertyListingSummaryResponse): string =>
  listing.property_sale?.property_no?.trim() || listing.listing_id;

// 12. 取得樓盤性質標籤
const resolveTransactionLabel = (listing: PropertyListingSummaryResponse): string =>
  listing.property_sale?.transaction_type === 'rent' ? '出租' : '出售';

// 13. 格式化面積
const formatArea = (listing: PropertyListingSummaryResponse): string => {
  const area = resolvePropertyArea(listing);
  return area > 0 ? t('common.unit.sqft', { value: area }) : '-';
};

// 14. 取得卡片圖片高度樣式
const resolveImageClass = (index: number): string =>
  index % 3 === 1 ? 'listing-card__image--short' : 'listing-card__image--tall';

watch([
  selectedRegion,
  selectedTransactionType,
  selectedPropertyType,
  selectedPriceRange,
  selectedAreaRange,
  selectedBedroom,
  selectedRenovationTag,
  publisherIdentityType,
  sortBy,
], scheduleListingsReload);

watch(() => props.channel, () => {
  resetFilters();
});

onMounted(() => {
  void loadListings();
});

onBeforeUnmount(() => {
  clearFilterReloadTimer();
});
</script>

<template>
  <main
    v-if="isSale"
    class="listing-page"
  >
    <aside class="listing-filter">
      <form
        class="listing-search"
        @submit.prevent="runSearch"
      >
        <input
          v-model="keyword"
          placeholder="搜尋樓盤..."
          autocomplete="off"
        />
        <button
          type="submit"
          aria-label="搜尋樓盤"
        >
          <AppIcon
            name="search"
            :size="13"
          />
        </button>
      </form>

      <div class="filter-heading">
        <span>篩選</span>
        <button
          type="button"
          :disabled="activeFilterCount === 0 && sortBy === 'latest'"
          @click="resetFilters"
        >
          重設
        </button>
      </div>

      <section
        v-for="group in saleFilterGroups"
        :key="group.key"
        class="filter-section"
      >
        <h2>{{ group.title }}</h2>
        <div class="filter-tags">
          <button
            v-for="option in group.options"
            :key="`${group.key}-${option.value}`"
            type="button"
            class="filter-tag"
            :class="{ 'filter-tag--active': option.value === group.activeValue }"
            @click="group.onSelect(option.value)"
          >
            {{ option.label }}
          </button>
        </div>
      </section>
    </aside>

    <section class="listing-results">
      <div class="sort-row">
        <div>
          <span>{{ loading ? t('common.status.loading') : resultLabel }}</span>
          <small v-if="activeFilterCount > 0">
            已套用 {{ activeFilterCount }} 項篩選
          </small>
        </div>
        <div class="sort-actions">
          <RouterLink
            class="publish-link"
            to="/properties/my/new"
          >
            發布樓盤
          </RouterLink>
          <select v-model="sortBy">
            <option value="latest">最新發佈</option>
            <option value="price_asc">價格低至高</option>
            <option value="price_desc">價格高至低</option>
            <option value="usable_area_desc">實用面積由大至小</option>
          </select>
        </div>
      </div>

      <div
        v-if="loading"
        class="listing-grid"
      >
        <div
          v-for="item in 6"
          :key="item"
          class="listing-card listing-card--loading"
        >
          <div class="listing-card__image listing-card__image--short home-pattern"></div>
          <div class="listing-card__body">
            <span></span>
            <h2></h2>
            <p></p>
            <strong></strong>
          </div>
        </div>
      </div>

      <div
        v-else-if="items.length > 0"
        class="listing-grid"
      >
        <RouterLink
          v-for="(listing, index) in items"
          :key="listing.listing_id"
          :to="resolvePropertyDetailPath(listing)"
          class="listing-card"
        >
          <div
            class="listing-card__image home-pattern"
            :class="resolveImageClass(index)"
          >
            <img
              v-if="resolvePropertyCoverImage(listing)"
              :src="resolvePropertyCoverImage(listing)?.url"
              :alt="resolvePropertyTitle(listing)"
            />
            <AppIcon
              v-else
              name="building"
              :size="42"
            />
          </div>
          <div class="listing-card__body">
            <div class="tag-row">
              <span class="tag-row__dark">{{ resolvePropertyPublisherRole(listing) }}</span>
              <span>{{ resolvePropertyDistrict(listing, preferenceStore.locale) }}</span>
              <span>{{ resolvePropertyTypeLabel(listing, preferenceStore.locale) }}</span>
              <span>{{ resolveTransactionLabel(listing) }}</span>
            </div>
            <h2>{{ resolvePropertyTitle(listing) }}</h2>
            <p>{{ resolveLocationText(listing) }}</p>
            <strong>
              {{ formatListingPrice(listing) }}
              <small v-if="resolvePriceUnit(listing)">{{ resolvePriceUnit(listing) }}</small>
            </strong>
            <div class="pill-row">
              <span>{{ resolvePropertyRooms(listing) }}</span>
              <span>{{ formatArea(listing) }}</span>
              <span>{{ resolvePropertyCommunityName(listing) }}</span>
            </div>
          </div>
          <div class="listing-card__footer">
            <span>{{ resolveListingNumber(listing) }}</span>
            <span>查看詳情</span>
          </div>
        </RouterLink>
      </div>

      <div
        v-else
        class="empty-panel"
      >
        {{ t('property.list.noListings') }}
      </div>

      <nav
        v-if="totalPages > 1"
        class="pagination-row"
        aria-label="樓盤分頁"
      >
        <button
          type="button"
          :disabled="page <= 1 || loading"
          @click="goToPage(page - 1)"
        >
          上一頁
        </button>
        <button
          v-for="item in pageNumbers"
          :key="item"
          type="button"
          :class="{ 'pagination-row__active': item === page }"
          :disabled="loading"
          @click="goToPage(item)"
        >
          {{ item }}
        </button>
        <button
          type="button"
          :disabled="page >= totalPages || loading"
          @click="goToPage(page + 1)"
        >
          下一頁
        </button>
      </nav>
    </section>
  </main>

  <main
    v-else
    class="service-page"
  >
    <section class="service-hero home-pattern">
      <div>
        <p>SERVICE APARTMENTS</p>
        <h1>{{ serviceHeroTitle }}</h1>
        <span>{{ serviceHeroDescription }}</span>
      </div>
    </section>

    <section class="service-section">
      <div class="service-toolbar">
        <h2 class="section-title">精選服務式住宅</h2>
        <form
          class="service-search"
          @submit.prevent="runSearch"
        >
          <input
            v-model="keyword"
            placeholder="搜尋服務式住宅..."
            autocomplete="off"
          />
          <button
            type="submit"
            aria-label="搜尋服務式住宅"
          >
            <AppIcon
              name="search"
              :size="13"
            />
          </button>
        </form>
      </div>
      <div
        v-if="!loading && items.length > 0"
        class="service-grid"
      >
        <RouterLink
          v-for="(listing, index) in items"
          :key="listing.listing_id"
          :to="resolvePropertyDetailPath(listing)"
          class="service-card"
        >
          <div
            class="service-card__image home-pattern"
            :class="`service-card__image--${(index % 3) + 1}`"
          >
            <img
              v-if="resolvePropertyCoverImage(listing)"
              :src="resolvePropertyCoverImage(listing)?.url"
              :alt="resolvePropertyTitle(listing)"
            />
            <AppIcon
              v-else
              name="building"
              :size="36"
            />
          </div>
          <div class="service-card__body">
            <div class="tag-row">
              <span class="tag-row__dark">{{ resolvePropertyPublisherRole(listing) }}</span>
              <span>{{ resolvePropertyDistrict(listing, preferenceStore.locale) }}</span>
            </div>
            <h2>{{ resolvePropertyTitle(listing) }}</h2>
            <p>{{ resolvePropertySummary(listing) }}</p>
            <strong>
              {{ formatListingPrice(listing) }}
              <small>{{ resolvePriceUnit(listing) }}</small>
            </strong>
            <div class="pill-row">
              <span>{{ resolvePropertyRooms(listing) }}</span>
              <span>{{ formatArea(listing) }}</span>
              <span>{{ resolvePropertyCommunityName(listing) }}</span>
            </div>
          </div>
          <div class="listing-card__footer">
            <span>{{ listing.listing_id }}</span>
            <span>查看詳情</span>
          </div>
        </RouterLink>
      </div>

      <div
        v-else
        class="empty-panel"
      >
        {{ loading ? t('common.status.loading') : t('property.list.noListings') }}
      </div>

      <nav
        v-if="totalPages > 1"
        class="pagination-row"
        aria-label="服務式住宅分頁"
      >
        <button
          type="button"
          :disabled="page <= 1 || loading"
          @click="goToPage(page - 1)"
        >
          上一頁
        </button>
        <button
          v-for="item in pageNumbers"
          :key="item"
          type="button"
          :class="{ 'pagination-row__active': item === page }"
          :disabled="loading"
          @click="goToPage(item)"
        >
          {{ item }}
        </button>
        <button
          type="button"
          :disabled="page >= totalPages || loading"
          @click="goToPage(page + 1)"
        >
          下一頁
        </button>
      </nav>
    </section>
  </main>
</template>

<style scoped>
.home-pattern {
  position: relative;
  overflow: hidden;
}

.home-pattern::after {
  position: absolute;
  inset: 0;
  background: repeating-linear-gradient(
    45deg,
    transparent,
    transparent 5px,
    rgb(0 0 0 / 0.025) 5px,
    rgb(0 0 0 / 0.025) 10px
  );
  content: '';
  pointer-events: none;
}

.listing-page {
  display: grid;
  grid-template-columns: 216px minmax(0, 1fr);
  min-height: calc(100vh - var(--app-header-offset, 48px));
  background: #ffffff;
  color: #1a1a1a;
}

.listing-filter {
  position: sticky;
  top: var(--app-header-offset, 48px);
  height: calc(100vh - var(--app-header-offset, 48px));
  overflow-y: auto;
  border-right: 1px solid #e4e4e4;
  background: #ffffff;
  padding: 18px 14px;
}

.listing-search,
.service-search {
  display: flex;
  gap: 5px;
}

.listing-search {
  margin-bottom: 14px;
}

.listing-search input,
.listing-search button,
.service-search input,
.service-search button,
.sort-row select {
  border: 1px solid #e4e4e4;
  border-radius: 2px;
  background: #ffffff;
  font: inherit;
  font-size: 11px;
  outline: none;
}

.listing-search input,
.service-search input {
  min-width: 0;
  flex: 1;
  padding: 7px 9px;
}

.listing-search input:focus,
.service-search input:focus {
  border-color: rgb(var(--color-primary));
}

.listing-search button,
.service-search button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary));
  color: #ffffff;
  cursor: pointer;
  padding: 7px 11px;
}

.filter-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid #f0f0f0;
  margin-bottom: 14px;
  padding-bottom: 10px;
}

.filter-heading span {
  color: #333333;
  font-size: 11px;
  font-weight: 500;
  letter-spacing: 1.5px;
}

.filter-heading button {
  border: 0;
  background: transparent;
  color: rgb(var(--color-primary));
  cursor: pointer;
  font: inherit;
  font-size: 10px;
}

.filter-heading button:disabled {
  color: #b6b6b6;
  cursor: default;
}

.filter-section {
  margin-bottom: 16px;
}

.filter-section h2 {
  margin: 0 0 6px;
  color: #777777;
  font-size: 9px;
  font-weight: 400;
  letter-spacing: 2px;
  text-transform: uppercase;
}

.filter-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 3px;
}

.filter-tag {
  border: 1px solid #e4e4e4;
  border-radius: 2px;
  background: #ffffff;
  color: #666666;
  cursor: pointer;
  font: inherit;
  font-size: 11px;
  padding: 3px 9px;
  transition:
    border-color 0.16s ease,
    background 0.16s ease,
    color 0.16s ease;
}

.filter-tag:hover {
  border-color: rgb(var(--color-primary));
  color: rgb(var(--color-primary));
}

.filter-tag--active {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary));
  color: #ffffff;
}

.filter-tag--active:hover {
  color: #ffffff;
}

.listing-results {
  min-width: 0;
  padding: 18px 20px 40px;
}

.sort-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
  margin-bottom: 12px;
}

.sort-row span {
  display: block;
  color: #777777;
  font-size: 11px;
}

.sort-row small {
  display: block;
  margin-top: 3px;
  color: #aaaaaa;
  font-size: 10px;
}

.sort-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.publish-link {
  border: 1px solid rgb(var(--color-primary));
  border-radius: 2px;
  background: rgb(var(--color-primary));
  color: #ffffff;
  font-size: 11px;
  padding: 6px 10px;
  text-decoration: none;
  white-space: nowrap;
}

.sort-row select {
  color: #333333;
  padding: 5px 8px;
}

.listing-grid {
  columns: 2;
  column-gap: 10px;
}

.listing-card,
.service-card {
  display: inline-block;
  width: 100%;
  overflow: hidden;
  break-inside: avoid;
  border: 1px solid #e4e4e4;
  border-radius: 3px;
  background: #ffffff;
  margin-bottom: 10px;
  color: inherit;
  cursor: pointer;
  text-decoration: none;
  transition:
    border-color 0.16s ease,
    box-shadow 0.16s ease;
}

.listing-card:hover,
.service-card:hover {
  border-color: rgb(var(--color-primary));
  box-shadow: 0 3px 16px rgb(240 90 0 / 0.15);
}

.listing-card__image,
.service-card__image {
  display: flex;
  width: 100%;
  align-items: center;
  justify-content: center;
  background: linear-gradient(160deg, #e8e8e8, #d0d0d0);
  color: #aaaaaa;
}

.listing-card__image--tall {
  height: 150px;
}

.listing-card__image--short {
  height: 94px;
}

.listing-card__image img,
.service-card__image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.listing-card__body,
.service-card__body {
  padding: 11px 13px;
}

.tag-row {
  display: flex;
  flex-wrap: wrap;
  gap: 3px;
  margin-bottom: 5px;
}

.tag-row span {
  border: 1px solid #e4e4e4;
  border-radius: 1px;
  color: #777777;
  font-size: 9px;
  letter-spacing: 0.8px;
  padding: 1px 5px;
}

.tag-row span.tag-row__dark {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary));
  color: #ffffff;
  font-weight: 500;
}

.listing-card h2,
.service-card h2 {
  margin: 0 0 4px;
  color: #1a1a1a;
  font-size: 13px;
  font-weight: 500;
  line-height: 1.35;
}

.listing-card p,
.service-card p {
  display: -webkit-box;
  margin: 0 0 7px;
  overflow: hidden;
  color: #777777;
  font-size: 11px;
  line-height: 1.55;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.listing-card strong,
.service-card strong {
  display: block;
  color: #1a1a1a;
  font-size: 17px;
  font-weight: 300;
  letter-spacing: 0;
}

.listing-card strong small,
.service-card strong small {
  margin-left: 4px;
  color: #777777;
  font-size: 11px;
  font-weight: 400;
}

.pill-row {
  display: flex;
  flex-wrap: wrap;
  gap: 3px;
  margin-top: 6px;
}

.pill-row span {
  max-width: 100%;
  overflow: hidden;
  border-radius: 2px;
  background: #f4f4f4;
  color: #333333;
  font-size: 9px;
  padding: 2px 6px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.listing-card__footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  border-top: 1px solid #f4f4f4;
  padding: 7px 13px;
  color: #777777;
  font-size: 10px;
}

.listing-card__footer span:first-child {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.listing-card__footer span:last-child {
  flex: 0 0 auto;
  color: rgb(var(--color-primary));
  font-weight: 500;
}

.listing-card--loading {
  pointer-events: none;
}

.listing-card--loading .listing-card__image,
.listing-card--loading .listing-card__body span,
.listing-card--loading .listing-card__body h2,
.listing-card--loading .listing-card__body p,
.listing-card--loading .listing-card__body strong {
  background: linear-gradient(90deg, #f4f4f4, #ececec, #f4f4f4);
  background-size: 200% 100%;
  animation: listing-loading 1.1s ease-in-out infinite;
}

.listing-card--loading .listing-card__body span,
.listing-card--loading .listing-card__body h2,
.listing-card--loading .listing-card__body p,
.listing-card--loading .listing-card__body strong {
  display: block;
  border-radius: 2px;
}

.listing-card--loading .listing-card__body span {
  width: 40%;
  height: 12px;
  margin-bottom: 8px;
}

.listing-card--loading .listing-card__body h2 {
  width: 80%;
  height: 16px;
}

.listing-card--loading .listing-card__body p {
  width: 100%;
  height: 28px;
}

.listing-card--loading .listing-card__body strong {
  width: 52%;
  height: 22px;
}

.pagination-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: center;
  gap: 5px;
  margin-top: 18px;
}

.pagination-row button {
  min-width: 32px;
  border: 1px solid #e4e4e4;
  border-radius: 2px;
  background: #ffffff;
  color: #555555;
  cursor: pointer;
  font: inherit;
  font-size: 11px;
  padding: 6px 9px;
}

.pagination-row button:disabled {
  color: #b6b6b6;
  cursor: default;
}

.pagination-row__active {
  border-color: rgb(var(--color-primary)) !important;
  background: rgb(var(--color-primary)) !important;
  color: #ffffff !important;
}

.service-page {
  background: #ffffff;
  color: #1a1a1a;
}

.service-hero {
  display: flex;
  min-height: 200px;
  align-items: flex-end;
  background: linear-gradient(160deg, #1a1a1a, #2e2e2e);
  padding: 40px 40px 32px;
}

.service-hero p {
  margin: 0 0 14px;
  color: #aaaaaa;
  font-size: 10px;
  letter-spacing: 3px;
}

.service-hero h1 {
  margin: 0 0 12px;
  white-space: pre-line;
  color: #ffffff;
  font-family: var(--font-display, Georgia, serif);
  font-size: 36px;
  font-weight: 400;
  line-height: 1.2;
}

.service-hero span {
  display: block;
  max-width: 400px;
  color: #aaaaaa;
  font-size: 13px;
  line-height: 1.8;
}

.service-section {
  padding: 32px 40px;
}

.service-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 16px;
}

.section-title {
  border-left: 3px solid rgb(var(--color-primary));
  margin: 0;
  padding-left: 10px;
  color: #1a1a1a;
  font-size: 13px;
  font-weight: 500;
  letter-spacing: 0.5px;
}

.service-search {
  width: min(320px, 100%);
}

.service-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 14px;
}

.service-card {
  display: block;
  margin-bottom: 0;
}

.service-card__image {
  height: 140px;
}

.service-card__image--1 {
  background: linear-gradient(160deg, #dde4e8, #b8c8d0);
}

.service-card__image--2 {
  background: linear-gradient(160deg, #e8e4d8, #d0c8b0);
}

.service-card__image--3 {
  background: linear-gradient(160deg, #e0e8e0, #c0d4c0);
}

.empty-panel {
  border: 1px solid #e4e4e4;
  border-radius: 3px;
  background: #ffffff;
  color: #777777;
  font-size: 12px;
  padding: 40px 20px;
  text-align: center;
}

@keyframes listing-loading {
  0% {
    background-position: 200% 0;
  }

  100% {
    background-position: -200% 0;
  }
}

@media (max-width: 1023px) {
  .listing-page {
    grid-template-columns: 1fr;
  }

  .listing-filter {
    position: static;
    height: auto;
    border-right: 0;
    border-bottom: 1px solid #e4e4e4;
  }

  .listing-grid {
    columns: 1;
  }

  .service-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 640px) {
  .listing-filter,
  .listing-results,
  .service-section {
    padding-right: 14px;
    padding-left: 14px;
  }

  .sort-row,
  .sort-actions,
  .service-toolbar {
    align-items: stretch;
    flex-direction: column;
  }

  .publish-link,
  .sort-row select,
  .service-search {
    width: 100%;
  }

  .service-hero {
    min-height: 180px;
    padding: 30px 20px 24px;
  }

  .service-hero h1 {
    font-size: 30px;
  }

  .service-grid {
    grid-template-columns: 1fr;
  }
}
</style>

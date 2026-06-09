<!--
 * 樓盤租售列表頁。
 * 1. 展示樓盤買賣租賃列表。
 * 2. 提供地區、租售、價格、面積、房數與標籤篩選。
-->
<script setup lang="ts">
import axios from 'axios';
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { RouterLink, useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';

import {
  fetchPropertySaleListings,
  fetchServicedApartmentListings,
} from '@/domains/property/api';
import {
  getPropertyOptionLabel,
  marketplaceDistricts,
  propertyAreaRangeOptions,
  propertyBedroomOptions,
  propertyPriceRangeOptions,
  propertyRentalTypeOptions,
  propertySalePriceRangeOptions,
  servicedAreaRangeOptions,
  servicedFacilityTagOptions,
  servicedPriceRangeOptions,
  propertyTagGroups,
  propertyTransactionTypeOptions,
  propertyTypeOptions,
  type PropertyRangeOption,
} from '@/domains/property/constants';
import type { MarketplaceLabelledOption } from '@/domains/marketplace/constants';
import type {
  PropertyChannel,
  PropertyListParams,
  PropertyListingSummaryResponse,
  PropertyTransactionType,
} from '@/domains/property/model';
import ListingAdRail from '@/shared/components/data-display/ListingAdRail.vue';
import AppGlassSelect from '@/shared/components/base/AppGlassSelect.vue';
import AppIcon from '@/shared/components/base/AppIcon.vue';
import { useFeedbackStore } from '@/app/stores/feedback';
import { usePreferenceStore } from '@/app/stores/preferences';
import {
  resolvePropertyArea,
  resolvePropertyCommunityName,
  resolvePropertyCoverImage,
  resolvePropertyDetailPath,
  resolvePropertyDistrict,
  resolvePropertyPrice,
  resolvePropertyPriceText,
  resolvePropertyPublisherRole,
  resolvePropertyRooms,
  resolvePropertySummary,
  resolvePropertyTagLabels,
  resolvePropertyTitle,
  resolvePropertyTransactionType,
  resolvePropertyTypeLabel,
  mergePropertyFeatureTags,
} from '@/shared/utils/property';

import { demoPropertyListings } from '@/domains/property/demo';

type PropertyResultViewMode = 'list' | 'grid';
type PropertySortBy = NonNullable<PropertyListParams['sort_by']>;
type ServicedPriceMode = 'monthly' | 'daily';
type PropertyConditionFilter = 'all' | 'new' | 'secondhand';
type ActiveConditionFilter = Exclude<PropertyConditionFilter, 'all'>;

interface PropertySortOption {
  value: PropertySortBy;
  labelKey: string;
}

const props = defineProps<{
  channel: PropertyChannel;
}>();

const route = useRoute();
const { t } = useI18n();
const feedbackStore = useFeedbackStore();
const preferenceStore = usePreferenceStore();

const propertyPublisherOptions = [
  { value: 'owner', label_zh_hk: '業主', label_en: 'Owner' },
  { value: 'agent', label_zh_hk: '代理', label_en: 'Agent' },
];
const propertySortOptions: PropertySortOption[] = [
  { value: 'latest', labelKey: 'property.filter.sortPresetLatest' },
  { value: 'price_desc', labelKey: 'property.filter.sortPresetPriceDesc' },
  { value: 'price_asc', labelKey: 'property.filter.sortPresetPriceAsc' },
  { value: 'usable_area_desc', labelKey: 'property.filter.sortPresetUsableAreaDesc' },
  { value: 'gross_area_desc', labelKey: 'property.filter.sortPresetGrossAreaDesc' },
  { value: 'usable_unit_price_desc', labelKey: 'property.filter.sortPresetUsableUnitPriceDesc' },
  { value: 'gross_unit_price_desc', labelKey: 'property.filter.sortPresetGrossUnitPriceDesc' },
];
const propertyDistrictSelectOptions = computed(() => [
  { label: t('property.list.allDistricts'), value: '' },
  ...marketplaceDistricts.map((district) => ({
    label: getPropertyOptionLabel(district, preferenceStore.locale),
    value: district.value,
  })),
]);
const propertySortSelectOptions = computed(() =>
  propertySortOptions.map((option) => ({
    label: t(option.labelKey),
    value: option.value,
  })),
);
const activePriceRangeOptions = computed(() =>
  isSaleChannel.value
    ? !(selectedTransactionTypes.value.length === 1 && selectedTransactionTypes.value[0] === 'rent')
      ? propertySalePriceRangeOptions
      : propertyPriceRangeOptions
    : servicedPriceRangeOptions,
);
const activeAreaRangeOptions = computed(() => isSaleChannel.value ? propertyAreaRangeOptions : servicedAreaRangeOptions);

const maxBoundPrice = 80000000;
const maxBoundArea = 5000;
const loading = ref(false);
const filterReloadTimer = ref<number | null>(null);
const keyword = ref('');
const districtCodes = ref<string[]>([]);
const sortBy = ref<PropertySortBy>('latest');
const selectedConditionFilters = ref<ActiveConditionFilter[]>([]);
const selectedTransactionTypes = ref<PropertyTransactionType[]>([]);
const selectedPropertyTypes = ref<string[]>([]);
const selectedRentalTypes = ref<string[]>([]);
const selectedBedrooms = ref<string[]>([]);
const minPrice = ref('0');
const maxPrice = ref(String(maxBoundPrice));
const minAreaSqft = ref('0');
const maxAreaSqft = ref(String(maxBoundArea));
const selectedPriceRange = ref('');
const selectedAreaRange = ref('');
const selectedTags = ref<string[]>([]);
const servicedPriceMode = ref<ServicedPriceMode>('monthly');
const publisherIdentityTypes = ref<string[]>([]);
const viewMode = ref<PropertyResultViewMode>('list');
const favoritesOnly = ref(false);
const historyOnly = ref(false);
const items = ref<PropertyListingSummaryResponse[]>([]);

const isSaleChannel = computed(() => props.channel === 'sale');
const maxPriceRangeValue = computed(() => toBoundedPrice(maxPrice.value, maxBoundPrice));
const maxAreaRangeValue = computed(() => toBoundedArea(maxAreaSqft.value, maxBoundArea));
const pageTitle = computed(() => isSaleChannel.value ? t('property.sale.title') : t('property.serviced.title'));
const adChannel = computed<'property_sale' | 'serviced_apartment'>(() =>
  isSaleChannel.value ? 'property_sale' : 'serviced_apartment',
);
const priceLabel = computed(() =>
  isSaleChannel.value ? t('property.sale.priceLabel') : t('property.serviced.priceLabel'),
);
const areaLabel = computed(() =>
  isSaleChannel.value ? t('property.sale.areaLabel') : t('property.serviced.areaLabel'),
);
const displayedSourceItems = computed(() =>
  items.value.length > 0 ? items.value : isSaleChannel.value ? demoPropertyListings : [],
);
const displayedItems = computed(() => {
  const keywordText = keyword.value.trim().toLowerCase();
  const minPriceValue = resolveMinPriceValue();
  const maxPriceValue = resolveMaxPriceValue();
  const minAreaValue = resolveMinAreaValue();
  const maxAreaValue = resolveMaxAreaValue();

  const filteredItems = displayedSourceItems.value.filter((listing) => {
    const titleText = resolvePropertyTitle(listing).toLowerCase();
    const summaryText = resolvePropertySummary(listing).toLowerCase();
    const communityText = resolvePropertyCommunityName(listing).toLowerCase();
    const districtText = resolvePropertyDistrict(listing, preferenceStore.locale).toLowerCase();
    const priceValue = resolvePropertyPrice(listing, servicedPriceMode.value);
    const areaValue = resolveListingArea(listing);
    const sale = listing.property_sale;
    const tags = mergePropertyFeatureTags(sale?.feature_tags ?? [], sale);
    const servicedTags = [...(listing.serviced_apartment?.facility_tags ?? []), ...(listing.serviced_apartment?.service_tags ?? [])];

    const matchesKeyword = !keywordText || [titleText, summaryText, communityText, districtText].some((value) => value.includes(keywordText));
    const matchesDistrict = districtCodes.value.length === 0 || districtCodes.value.includes(listing.district_code);
    const matchesMinPrice = minPriceValue === undefined || priceValue >= minPriceValue;
    const matchesMaxPrice = maxPriceValue === undefined || priceValue <= maxPriceValue;
    const matchesMinArea = minAreaValue === undefined || areaValue >= minAreaValue;
    const matchesMaxArea = maxAreaValue === undefined || areaValue <= maxAreaValue;
    const matchesTransaction = !isSaleChannel.value || selectedTransactionTypes.value.length === 0 || selectedTransactionTypes.value.includes(resolvePropertyTransactionType(listing));
    const matchesPropertyType = !isSaleChannel.value || selectedPropertyTypes.value.length === 0 || selectedPropertyTypes.value.includes(sale?.property_type ?? '');
    const matchesRentalType = !isSaleChannel.value || selectedRentalTypes.value.length === 0 || selectedRentalTypes.value.includes(sale?.rental_type ?? '');
    const matchesPublisher = publisherIdentityTypes.value.length === 0 || publisherIdentityTypes.value.includes(listing.publisher_identity_type);
    const matchesBedroom = !isSaleChannel.value || selectedBedrooms.value.length === 0 || selectedBedrooms.value.some((bedroom) => matchesBedroomCount(sale?.bedroom_count ?? 0, bedroom));
    const matchesCondition = !isSaleChannel.value || matchesConditionFilter(tags);
    const matchesTags = isSaleChannel.value
      ? selectedTags.value.every((tag) => tags.includes(tag))
      : selectedTags.value.every((tag) => servicedTags.includes(tag));

    return matchesKeyword &&
      matchesDistrict &&
      matchesMinPrice &&
      matchesMaxPrice &&
      matchesMinArea &&
      matchesMaxArea &&
      matchesTransaction &&
      matchesPropertyType &&
      matchesRentalType &&
      matchesPublisher &&
      matchesBedroom &&
      matchesCondition &&
      matchesTags;
  });

  return sortDisplayedItems(filteredItems);
});
// 1. 取得查詢字串值
const readQueryValue = (value: unknown): string => {
  if (Array.isArray(value)) {
    return String(value[0] ?? '');
  }

  return typeof value === 'string' ? value : '';
};

// 2. 轉換正數查詢值
const toPositiveNumber = (value: string): number | undefined => {
  const parsed = Number(value);

  return Number.isFinite(parsed) && parsed > 0 ? parsed : undefined;
};

// 3. 轉換價格滑桿數值
const toBoundedPrice = (value: string, fallback: number): number => {
  const parsed = Number(value);
  if (!Number.isFinite(parsed)) {
    return fallback;
  }

  return Math.min(maxBoundPrice, Math.max(0, Math.round(parsed)));
};

// 4. 轉換面積滑桿數值
const toBoundedArea = (value: string, fallback: number): number => {
  const parsed = Number(value);
  if (!Number.isFinite(parsed)) {
    return fallback;
  }

  return Math.min(maxBoundArea, Math.max(0, Math.round(parsed)));
};

// 5. 建立初始標籤條件
const buildInitialTags = (): string[] => {
  const feature = readQueryValue(route.query.feature);
  const featureTags = readQueryValue(route.query.feature_tags)
    .split(',')
    .map((tag) => tag.trim())
    .filter(Boolean);

  return [...new Set([feature, ...featureTags].filter(Boolean))];
};

// 6. 套用路由預設條件
const applyRoutePreset = (): void => {
  const initialTags = buildInitialTags();
  selectedConditionFilters.value = initialTags.includes('brand_new') ? ['new'] : [];
  selectedTags.value = initialTags.filter((tag) => tag !== 'brand_new');
  selectedBedrooms.value = readQueryValue(route.query.bedrooms)
    ? [readQueryValue(route.query.bedrooms)]
    : [];
  publisherIdentityTypes.value = readQueryValue(route.query.publisher)
    ? [readQueryValue(route.query.publisher)]
    : [];
};

// 9. 取得下拉選項文案
const resolveOptionText = (options: MarketplaceLabelledOption[], value: string): string => {
  const option = options.find((item) => item.value === value);

  return option ? getPropertyOptionLabel(option, preferenceStore.locale) : t('property.filter.any');
};

// 9.1 取得廣告等級標籤
const resolveAdPackageBadge = (listing: PropertyListingSummaryResponse): string => {
  const code = listing.property_sale?.ad_package_code;
  if (code === 'premium') {
    return '黃金';
  }
  if (code === 'featured') {
    return '置頂';
  }
  if (code === 'fast_sale') {
    return '即走';
  }
  return '';
};

// 10. 判斷房數篩選
const matchesBedroomCount = (bedroomCount: number, value: string): boolean => {
  const target = Number(value);
  return target >= 5 ? bedroomCount >= 5 : bedroomCount === target;
};

// 11. 判斷房源類型
const matchesConditionFilter = (tags: string[]): boolean => {
  if (selectedConditionFilters.value.length === 0) {
    return true;
  }
  if (selectedConditionFilters.value.includes('new') && selectedConditionFilters.value.includes('secondhand')) {
    return true;
  }
  if (selectedConditionFilters.value.includes('new')) {
    return tags.includes('brand_new');
  }

  return !tags.includes('brand_new');
};

// 12. 切換多選值
const toggleFilterValue = <T extends string>(values: T[], value: T): T[] =>
  values.includes(value)
    ? values.filter((item) => item !== value)
    : [...values, value];

// 13. 取得列表面積值
const resolveListingArea = (listing: PropertyListingSummaryResponse): number =>
  resolvePropertyArea(listing);

// 14. 取得價格下限
const resolveMinPriceValue = (): number | undefined =>
  toPositiveNumber(minPrice.value);

// 15. 取得價格上限
const resolveMaxPriceValue = (): number | undefined =>
  maxPriceRangeValue.value >= maxBoundPrice ? undefined : toPositiveNumber(maxPrice.value);

// 16. 取得面積下限
const resolveMinAreaValue = (): number | undefined =>
  toPositiveNumber(minAreaSqft.value);

// 17. 取得面積上限
const resolveMaxAreaValue = (): number | undefined =>
  maxAreaRangeValue.value >= maxBoundArea ? undefined : toPositiveNumber(maxAreaSqft.value);

// 18. 取得建築面積值
const resolveGrossAreaValue = (listing: PropertyListingSummaryResponse): number =>
  listing.property_sale?.gross_area_sqft ||
  listing.property_sale?.usable_area_sqft ||
  listing.serviced_apartment?.room_types?.[0]?.usable_area_sqft ||
  0;

// 19. 取得尺價
const resolveUnitPrice = (listing: PropertyListingSummaryResponse, mode: 'usable' | 'gross'): number => {
  const price = resolvePropertyPrice(listing);
  const area = mode === 'gross' ? resolveGrossAreaValue(listing) : resolveListingArea(listing);

  return area > 0 ? price / area : 0;
};

// 20. 排序畫面結果
const sortDisplayedItems = (source: PropertyListingSummaryResponse[]): PropertyListingSummaryResponse[] => {
  const itemsToSort = [...source];

  if (sortBy.value === 'usable_area_desc') {
    return itemsToSort.sort((first, second) => resolveListingArea(second) - resolveListingArea(first));
  }
  if (sortBy.value === 'gross_area_desc') {
    return itemsToSort.sort((first, second) => resolveGrossAreaValue(second) - resolveGrossAreaValue(first));
  }
  if (sortBy.value === 'usable_unit_price_desc') {
    return itemsToSort.sort((first, second) => resolveUnitPrice(second, 'usable') - resolveUnitPrice(first, 'usable'));
  }
  if (sortBy.value === 'gross_unit_price_desc') {
    return itemsToSort.sort((first, second) => resolveUnitPrice(second, 'gross') - resolveUnitPrice(first, 'gross'));
  }

  return itemsToSort;
};

// 21. 切換標籤
const toggleTag = (value: string): void => {
  const index = selectedTags.value.indexOf(value);
  if (index >= 0) {
    selectedTags.value.splice(index, 1);
    return;
  }

  selectedTags.value.push(value);
};

// 22. 判斷標籤分組是否未篩選
const isTagGroupClear = (options: MarketplaceLabelledOption[]): boolean =>
  options.every((option) => !selectedTags.value.includes(option.value));

// 23. 清除標籤分組
const clearTagGroup = (options: MarketplaceLabelledOption[]): void => {
  const values = new Set(options.map((option) => option.value));
  selectedTags.value = selectedTags.value.filter((tag) => !values.has(tag));
};

// 24. 套用價格區間
const applyPriceRange = (option?: PropertyRangeOption): void => {
  selectedPriceRange.value = option?.value ?? '';
  minPrice.value = String(option?.min ?? 0);
  maxPrice.value = String(option?.max ?? maxBoundPrice);
};

// 25. 套用面積區間
const applyAreaRange = (option?: PropertyRangeOption): void => {
  selectedAreaRange.value = option?.value ?? '';
  minAreaSqft.value = String(option?.min ?? 0);
  maxAreaSqft.value = String(option?.max ?? maxBoundArea);
};

// 26. 取得區間選項文案
const resolveRangeOptionText = (option: PropertyRangeOption): string =>
  getPropertyOptionLabel(option, preferenceStore.locale);

// 27. 更新排序
const updateSortBy = (value: PropertySortBy): void => {
  sortBy.value = value;
};

// 27.1 從自繪下拉更新排序
const handleSortChange = (value: string): void => {
  updateSortBy(value as PropertySortBy);
};

// 28. 排程刷新篩選結果
const scheduleFilterReload = (): void => {
  if (filterReloadTimer.value) {
    window.clearTimeout(filterReloadTimer.value);
  }

  filterReloadTimer.value = window.setTimeout(() => {
    void loadListings();
    filterReloadTimer.value = null;
  }, 80);
};

// 29. 清除篩選
const clearFilters = (): void => {
  keyword.value = '';
  districtCodes.value = [];
  sortBy.value = 'latest';
  selectedConditionFilters.value = [];
  selectedTransactionTypes.value = [];
  selectedPropertyTypes.value = [];
  selectedRentalTypes.value = [];
  selectedBedrooms.value = [];
  minPrice.value = '0';
  maxPrice.value = String(maxBoundPrice);
  minAreaSqft.value = '0';
  maxAreaSqft.value = String(maxBoundArea);
  selectedPriceRange.value = '';
  selectedAreaRange.value = '';
  selectedTags.value = [];
  servicedPriceMode.value = 'monthly';
  publisherIdentityTypes.value = [];
  favoritesOnly.value = false;
  historyOnly.value = false;
  void loadListings();
};

// 30. 讀取列表資料
const loadListings = async (): Promise<void> => {
  loading.value = true;

  try {
    const params: PropertyListParams = {
      page: 1,
      page_size: 30,
      keyword: keyword.value.trim(),
      district_code: districtCodes.value[0],
      sort_by: sortBy.value,
      min_price_hkd: resolveMinPriceValue(),
      max_price_hkd: resolveMaxPriceValue(),
      publisher_identity_type: publisherIdentityTypes.value[0],
    };

    if (isSaleChannel.value) {
      params.transaction_type = selectedTransactionTypes.value[0];
      params.property_type = selectedPropertyTypes.value[0];
      params.rental_type = selectedRentalTypes.value[0];
      params.min_area_sqft = resolveMinAreaValue();
      params.max_area_sqft = resolveMaxAreaValue();
      params.bedroom_count = selectedBedrooms.value[0] ? Number(selectedBedrooms.value[0]) : undefined;
      params.feature_tags = selectedTags.value.length > 0 ? selectedTags.value.join(',') : undefined;
	      params.is_new = selectedConditionFilters.value.length === 1
	        ? selectedConditionFilters.value[0] === 'new'
	        : undefined;
	    } else {
	      params.price_mode = servicedPriceMode.value;
	      params.min_area_sqft = resolveMinAreaValue();
	      params.max_area_sqft = resolveMaxAreaValue();
	      params.facility_tags = selectedTags.value.length > 0 ? selectedTags.value.join(',') : undefined;
	    }

    const response = isSaleChannel.value
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
    items.value = isSaleChannel.value ? demoPropertyListings : [];
  } finally {
    loading.value = false;
  }
};

// 31. 格式化價格
const formatListingPrice = (listing: PropertyListingSummaryResponse): string =>
  resolvePropertyPriceText(listing, preferenceStore.locale, servicedPriceMode.value);

// 32. 取得單張卡片價格標籤
const resolveListingPriceLabel = (listing: PropertyListingSummaryResponse): string =>
  isSaleChannel.value && resolvePropertyTransactionType(listing) === 'rent'
    ? t('property.sale.rentLabel')
    : isSaleChannel.value ? priceLabel.value : servicedPriceMode.value === 'daily' ? t('property.filter.dailyRent') : priceLabel.value;

// 33. 取得標籤分組文案
const resolveTagGroupLabel = (group: { label_zh_hk: string; label_en: string }): string =>
  preferenceStore.locale === 'zh-HK' ? group.label_zh_hk : group.label_en;

watch(
  [
    districtCodes,
    sortBy,
    selectedConditionFilters,
    selectedTransactionTypes,
    selectedPropertyTypes,
    selectedRentalTypes,
    selectedBedrooms,
    publisherIdentityTypes,
    selectedTags,
    servicedPriceMode,
  ],
  scheduleFilterReload,
  { deep: true },
);

watch(
  () => route.fullPath,
  () => {
    applyRoutePreset();
    void loadListings();
  },
);

watch(selectedTransactionTypes, () => {
  minPrice.value = '0';
  maxPrice.value = String(maxBoundPrice);
  if (!selectedTransactionTypes.value.includes('rent')) {
    selectedRentalTypes.value = [];
  }
});

onMounted(() => {
  applyRoutePreset();
  void loadListings();
});

onBeforeUnmount(() => {
  if (filterReloadTimer.value) {
    window.clearTimeout(filterReloadTimer.value);
  }
});
</script>

<template>
  <main class="property-list-page">
    <section class="property-filter-layout">
      <section class="property-top-layout">
        <section class="property-filter-panel">
        <label class="property-field">
          <span>{{ t('property.list.district') }}</span>
          <div class="property-option-list property-option-list--inline">
            <button
              v-for="district in propertyDistrictSelectOptions"
              :key="district.value"
              type="button"
              class="property-option-row property-option-row--inline"
              :class="{ 'property-option-row--active': district.value === '' ? districtCodes.length === 0 : districtCodes.includes(district.value) }"
              @click="districtCodes = district.value === '' ? [] : toggleFilterValue(districtCodes, district.value)"
            >
              <span class="property-option-box" />
              <span>{{ district.label }}</span>
            </button>
          </div>
        </label>

        <template v-if="isSaleChannel">
          <section class="property-filter-section">
            <h3>{{ t('property.filter.conditionType') }}</h3>
            <div class="property-option-list property-option-list--inline">
              <button
                type="button"
                class="property-option-row property-option-row--inline"
                :class="{ 'property-option-row--active': selectedConditionFilters.length === 0 }"
                @click="selectedConditionFilters = []"
              >
                <span class="property-option-box" />
                <span>{{ t('property.filter.all') }}</span>
              </button>
              <button
                type="button"
                class="property-option-row property-option-row--inline"
                :class="{ 'property-option-row--active': selectedConditionFilters.includes('new') }"
                @click="selectedConditionFilters = toggleFilterValue(selectedConditionFilters, 'new')"
              >
                <span class="property-option-box" />
                <span>{{ t('property.filter.conditionNew') }}</span>
              </button>
              <button
                type="button"
                class="property-option-row property-option-row--inline"
                :class="{ 'property-option-row--active': selectedConditionFilters.includes('secondhand') }"
                @click="selectedConditionFilters = toggleFilterValue(selectedConditionFilters, 'secondhand')"
              >
                <span class="property-option-box" />
                <span>{{ t('property.filter.conditionSecondhand') }}</span>
              </button>
            </div>
          </section>

          <section class="property-filter-section">
            <h3>{{ t('property.filter.transactionType') }}</h3>
            <div class="property-option-list property-option-list--inline">
              <button
                type="button"
                class="property-option-row property-option-row--inline"
                :class="{ 'property-option-row--active': selectedTransactionTypes.length === 0 }"
                @click="selectedTransactionTypes = []"
              >
                <span class="property-option-box" />
                <span>{{ t('property.filter.all') }}</span>
              </button>
              <button
                v-for="typeOption in propertyTransactionTypeOptions"
                :key="typeOption.value"
                type="button"
                class="property-option-row property-option-row--inline"
                :class="{ 'property-option-row--active': selectedTransactionTypes.includes(typeOption.value) }"
                @click="selectedTransactionTypes = toggleFilterValue(selectedTransactionTypes, typeOption.value)"
              >
                <span class="property-option-box" />
                <span>{{ getPropertyOptionLabel(typeOption, preferenceStore.locale) }}</span>
              </button>
            </div>
          </section>

          <section class="property-filter-section">
            <h3>{{ t('property.sale.typeLabel') }}</h3>
            <div class="property-option-list property-option-list--inline">
              <button
                type="button"
                class="property-option-row property-option-row--inline"
                :class="{ 'property-option-row--active': selectedPropertyTypes.length === 0 }"
                @click="selectedPropertyTypes = []"
              >
                <span class="property-option-box" />
                <span>{{ t('property.filter.all') }}</span>
              </button>
              <button
                v-for="typeOption in propertyTypeOptions"
                :key="typeOption.value"
                type="button"
                class="property-option-row property-option-row--inline"
                :class="{ 'property-option-row--active': selectedPropertyTypes.includes(typeOption.value) }"
                @click="selectedPropertyTypes = toggleFilterValue(selectedPropertyTypes, typeOption.value)"
              >
                <span class="property-option-box" />
                <span>{{ getPropertyOptionLabel(typeOption, preferenceStore.locale) }}</span>
              </button>
            </div>
          </section>

          <section
            v-if="selectedTransactionTypes.includes('rent')"
            class="property-filter-section"
          >
            <h3>{{ t('property.filter.rentalType') }}</h3>
            <div class="property-option-list property-option-list--inline">
              <button
                type="button"
                class="property-option-row property-option-row--inline"
                :class="{ 'property-option-row--active': selectedRentalTypes.length === 0 }"
                @click="selectedRentalTypes = []"
              >
                <span class="property-option-box" />
                <span>{{ t('property.filter.all') }}</span>
              </button>
              <button
                v-for="typeOption in propertyRentalTypeOptions"
                :key="typeOption.value"
                type="button"
                class="property-option-row property-option-row--inline"
                :class="{ 'property-option-row--active': selectedRentalTypes.includes(typeOption.value) }"
                @click="selectedRentalTypes = toggleFilterValue(selectedRentalTypes, typeOption.value)"
              >
                <span class="property-option-box" />
                <span>{{ getPropertyOptionLabel(typeOption, preferenceStore.locale) }}</span>
              </button>
            </div>
          </section>
        </template>

        <section class="property-filter-section property-filter-section--price">
          <h3>{{ t('property.filter.priceRange') }}</h3>
          <div :class="isSaleChannel ? 'property-price-filter' : 'property-price-filter property-price-filter--serviced'">
            <div
              v-if="!isSaleChannel"
              class="property-option-list property-option-list--inline property-price-filter__mode"
            >
              <button
                type="button"
                class="property-option-row property-option-row--inline"
                :class="{ 'property-option-row--active': servicedPriceMode === 'monthly' }"
                @click="servicedPriceMode = 'monthly'"
              >
                <span class="property-option-box" />
                <span>{{ t('property.filter.monthlyRent') }}</span>
              </button>
              <button
                type="button"
                class="property-option-row property-option-row--inline"
                :class="{ 'property-option-row--active': servicedPriceMode === 'daily' }"
                @click="servicedPriceMode = 'daily'"
              >
                <span class="property-option-box" />
                <span>{{ t('property.filter.dailyRent') }}</span>
              </button>
            </div>
            <div class="property-option-list property-option-list--inline property-price-filter__range">
              <button
                type="button"
                class="property-option-row property-option-row--inline"
                :class="{ 'property-option-row--active': selectedPriceRange === '' }"
                @click="applyPriceRange()"
              >
                <span class="property-option-box" />
                <span>{{ t('property.filter.any') }}</span>
              </button>
              <button
                v-for="option in activePriceRangeOptions"
                :key="option.value"
                type="button"
                class="property-option-row property-option-row--inline"
                :class="{ 'property-option-row--active': selectedPriceRange === option.value }"
                @click="applyPriceRange(option)"
              >
                <span class="property-option-box" />
                <span>{{ resolveRangeOptionText(option) }}</span>
              </button>
            </div>
          </div>
        </section>

        <template v-if="isSaleChannel">
          <section class="property-filter-section property-filter-section--area">
            <h3>{{ t('property.filter.areaRange') }}</h3>
            <div class="property-option-list property-option-list--inline">
              <button
                type="button"
                class="property-option-row property-option-row--inline"
                :class="{ 'property-option-row--active': selectedAreaRange === '' }"
                @click="applyAreaRange()"
              >
                <span class="property-option-box" />
                <span>{{ t('property.filter.any') }}</span>
              </button>
              <button
                v-for="option in activeAreaRangeOptions"
                :key="option.value"
                type="button"
                class="property-option-row property-option-row--inline"
                :class="{ 'property-option-row--active': selectedAreaRange === option.value }"
                @click="applyAreaRange(option)"
              >
                <span class="property-option-box" />
                <span>{{ resolveRangeOptionText(option) }}</span>
              </button>
            </div>
          </section>

          <section class="property-filter-section">
            <h3>{{ t('property.filter.rooms') }}</h3>
            <div class="property-option-list property-option-list--inline">
              <button
                type="button"
                class="property-option-row property-option-row--inline"
                :class="{ 'property-option-row--active': selectedBedrooms.length === 0 }"
                @click="selectedBedrooms = []"
              >
                <span class="property-option-box" />
                <span>{{ t('property.filter.all') }}</span>
              </button>
              <button
                v-for="bedroom in propertyBedroomOptions"
                :key="bedroom.value"
                type="button"
                class="property-option-row property-option-row--inline"
                :class="{ 'property-option-row--active': selectedBedrooms.includes(bedroom.value) }"
                @click="selectedBedrooms = toggleFilterValue(selectedBedrooms, bedroom.value)"
              >
                <span class="property-option-box" />
                <span>{{ getPropertyOptionLabel(bedroom, preferenceStore.locale) }}</span>
              </button>
            </div>
          </section>

          <section class="property-filter-section">
            <h3>{{ t('property.filter.publisher') }}</h3>
            <div class="property-option-list property-option-list--inline">
              <button
                type="button"
                class="property-option-row property-option-row--inline"
                :class="{ 'property-option-row--active': publisherIdentityTypes.length === 0 }"
                @click="publisherIdentityTypes = []"
              >
                <span class="property-option-box" />
                <span>{{ t('property.filter.all') }}</span>
              </button>
              <button
                v-for="publisher in propertyPublisherOptions"
                :key="publisher.value"
                type="button"
                class="property-option-row property-option-row--inline"
                :class="{ 'property-option-row--active': publisherIdentityTypes.includes(publisher.value) }"
                @click="publisherIdentityTypes = toggleFilterValue(publisherIdentityTypes, publisher.value)"
              >
                <span class="property-option-box" />
                <span>{{ getPropertyOptionLabel(publisher, preferenceStore.locale) }}</span>
              </button>
            </div>
          </section>
        </template>

        <section
          v-else
          class="property-filter-section property-filter-section--area"
        >
          <h3>{{ t('property.filter.areaRange') }}</h3>
          <div class="property-option-list property-option-list--inline">
            <button
              type="button"
              class="property-option-row property-option-row--inline"
              :class="{ 'property-option-row--active': selectedAreaRange === '' }"
              @click="applyAreaRange()"
            >
              <span class="property-option-box" />
              <span>{{ t('property.filter.any') }}</span>
            </button>
            <button
              v-for="option in activeAreaRangeOptions"
              :key="option.value"
              type="button"
              class="property-option-row property-option-row--inline"
              :class="{ 'property-option-row--active': selectedAreaRange === option.value }"
              @click="applyAreaRange(option)"
            >
              <span class="property-option-box" />
              <span>{{ resolveRangeOptionText(option) }}</span>
            </button>
          </div>
        </section>

        <div
          v-if="isSaleChannel"
          class="property-tag-stack"
        >
          <section
            v-for="group in propertyTagGroups"
            :key="group.key"
            class="property-tag-group"
          >
            <h3>{{ resolveTagGroupLabel(group) }}</h3>
            <div class="property-option-list">
              <button
                type="button"
                class="property-option-row"
                :class="{ 'property-option-row--active': isTagGroupClear(group.options) }"
                @click="clearTagGroup(group.options)"
              >
                <span class="property-option-box" />
                <span>{{ t('property.filter.all') }}</span>
              </button>
              <button
                v-for="tag in group.options"
                :key="tag.value"
                type="button"
                class="property-option-row"
                :class="{ 'property-option-row--active': selectedTags.includes(tag.value) }"
                @click="toggleTag(tag.value)"
              >
                <span class="property-option-box" />
                <span>{{ getPropertyOptionLabel(tag, preferenceStore.locale) }}</span>
              </button>
            </div>
          </section>
        </div>
        <section
          v-else
          class="property-filter-section"
        >
          <h3>{{ t('property.editor.facilityTagsField') }}</h3>
          <div class="property-option-list property-option-list--inline">
            <button
              type="button"
              class="property-option-row property-option-row--inline"
              :class="{ 'property-option-row--active': selectedTags.length === 0 }"
              @click="selectedTags = []"
            >
              <span class="property-option-box" />
              <span>{{ t('property.filter.any') }}</span>
            </button>
            <button
              v-for="tag in servicedFacilityTagOptions"
              :key="tag.value"
              type="button"
              class="property-option-row property-option-row--inline"
              :class="{ 'property-option-row--active': selectedTags.includes(tag.value) }"
              @click="selectedTags = toggleFilterValue(selectedTags, tag.value)"
            >
              <span class="property-option-box" />
              <span>{{ getPropertyOptionLabel(tag, preferenceStore.locale) }}</span>
            </button>
          </div>
        </section>
        <div class="property-search-toolbar">
          <span>{{ t('property.filter.searchLabel') }}</span>
          <div class="property-search-input">
            <AppIcon
              name="search"
              :size="16"
            />
            <input
              v-model="keyword"
              type="search"
              :placeholder="t('property.list.keywordPlaceholder')"
              @keyup.enter="loadListings"
            />
          </div>
          <button
            type="button"
            class="property-toolbar-button property-toolbar-button--primary"
            @click="loadListings"
          >
            {{ t('property.filter.searchAction') }}
          </button>
          <button
            type="button"
            class="property-toolbar-icon-button"
            :class="{ 'property-toolbar-icon-button--active': favoritesOnly }"
            :aria-label="t('property.filter.favorites')"
            :title="t('property.filter.favorites')"
            :data-tooltip="t('property.filter.favorites')"
            @click="favoritesOnly = !favoritesOnly"
          >
            <AppIcon
              name="star"
              :size="17"
            />
          </button>
          <button
            type="button"
            class="property-toolbar-icon-button"
            :class="{ 'property-toolbar-icon-button--active': historyOnly }"
            :aria-label="t('property.filter.history')"
            :title="t('property.filter.history')"
            :data-tooltip="t('property.filter.history')"
            @click="historyOnly = !historyOnly"
          >
            <AppIcon
              name="clock"
              :size="17"
            />
          </button>
          <button
            type="button"
            class="property-toolbar-icon-button"
            :class="{ 'property-toolbar-icon-button--active': viewMode === 'grid' }"
            :aria-label="t('marketplace.filter.gridView')"
            :title="t('marketplace.filter.gridView')"
            :data-tooltip="t('marketplace.filter.gridView')"
            @click="viewMode = viewMode === 'grid' ? 'list' : 'grid'"
          >
            <AppIcon
              name="layout-grid"
              :size="17"
            />
          </button>
          <button
            type="button"
            class="property-toolbar-icon-button"
            :aria-label="t('property.filter.clearAll')"
            :title="t('property.filter.clearAll')"
            :data-tooltip="t('property.filter.clearAll')"
            @click="clearFilters"
          >
            <AppIcon
              name="trash"
              :size="17"
            />
          </button>
          <AppGlassSelect
            class="property-sort-menu"
            :model-value="sortBy"
            :options="propertySortSelectOptions"
            panel-max-height="14rem"
            @change="handleSortChange"
          />
        </div>
        </section>

        <ListingAdRail
          :channel="adChannel"
          variant="top"
        />
      </section>

      <section class="property-content-layout">
        <section class="property-filter-results">
        <div class="property-results-topline">
          <div class="property-results-heading">
            <div>
              <p>{{ loading ? t('common.status.loading') : t('property.list.results', { count: displayedItems.length }) }}</p>
              <h2>{{ pageTitle }}</h2>
            </div>
          </div>
        </div>

        <section
          v-if="loading"
          class="property-empty"
        >
          {{ t('common.status.loading') }}
        </section>

        <section
          v-else-if="displayedItems.length === 0"
          class="property-empty"
        >
          {{ t('property.list.noListings') }}
        </section>

        <section
          v-else
          class="property-grid"
          :class="viewMode === 'grid' ? 'property-grid--grid' : 'property-grid--list'"
        >
          <RouterLink
            v-for="listing in displayedItems"
            :key="listing.listing_id"
            :to="resolvePropertyDetailPath(listing)"
            class="property-card group"
            :class="{ 'property-card--list': viewMode === 'list' }"
          >
            <div class="property-card__media">
              <span
                v-if="resolveAdPackageBadge(listing)"
                class="property-card__ad-badge"
                :class="{ 'property-card__ad-badge--gold': listing.property_sale?.ad_package_code === 'premium' }"
              >
                {{ resolveAdPackageBadge(listing) }}
              </span>
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
                <span v-if="isSaleChannel">
                  {{ resolveOptionText(propertyTransactionTypeOptions, resolvePropertyTransactionType(listing)) }}
                </span>
              </div>
              <h2>{{ resolvePropertyTitle(listing) }}</h2>
              <p>{{ resolvePropertySummary(listing) }}</p>
              <dl class="property-card__specs">
                <div>
                  <dt>{{ resolveListingPriceLabel(listing) }}</dt>
                  <dd>{{ formatListingPrice(listing) }}</dd>
                </div>
                <div>
                  <dt>{{ areaLabel }}</dt>
                  <dd>{{ t('common.unit.sqft', { value: resolveListingArea(listing) }) }}</dd>
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
              <div
                v-if="resolvePropertyTagLabels(listing, preferenceStore.locale).length > 0"
                class="property-card__tags"
              >
                <span
                  v-for="tag in resolvePropertyTagLabels(listing, preferenceStore.locale)"
                  :key="tag"
                >
                  {{ tag }}
                </span>
              </div>
              <div class="property-card__footer">
                <span>{{ resolvePropertyRooms(listing) }}</span>
                <span>{{ t('common.action.viewDetail') }}</span>
              </div>
            </div>
          </RouterLink>
        </section>
      </section>

      <ListingAdRail
        :channel="adChannel"
        variant="bottom"
      />
      </section>
    </section>
  </main>
</template>

<style scoped>
.property-list-page {
  display: block;
  max-width: var(--layout-page-max-width);
  min-height: calc(100vh - var(--app-header-offset, 0rem));
  margin: 0 auto;
  padding: 1.75rem var(--layout-page-padding-inline) 4rem;
  color: rgb(var(--color-text));
  overflow: visible;
}

.property-filter-layout {
  display: grid;
  gap: 2rem;
}

.property-top-layout,
.property-content-layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr) clamp(13rem, 18vw, 17rem);
  align-items: start;
  gap: 1.5rem;
}

.property-filter-panel {
  display: flex;
  flex-direction: column;
  gap: 0;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface-raised));
  overflow: visible;
  padding: 0.85rem 1rem;
}

.property-filter-sidebar__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid rgb(var(--color-border));
  padding-bottom: 0.75rem;
  gap: 1rem;
}

.property-filter-sidebar__header h2 {
  margin: 0;
  color: rgb(var(--color-text));
  font-family: var(--font-display);
  font-size: 1.25rem;
  font-weight: 500;
  line-height: 1.4;
}

.property-tag-group h3,
.property-filter-section h3,
.property-field > span {
  margin: 0;
  color: rgb(var(--color-text-muted));
  font-size: 0.8rem;
  font-weight: 600;
  letter-spacing: 0.1em;
  line-height: 1;
  padding-top: 0.5rem;
  text-transform: uppercase;
}

.property-filter-sidebar__header button {
  color: rgb(var(--color-primary));
  font-size: 0.875rem;
  font-weight: 500;
  letter-spacing: 0.02em;
}

.property-filter-section--area,
.property-filter-section--price {
  padding-top: 0.35rem;
}

.property-filter-panel > .property-field,
.property-filter-panel > .property-filter-section,
.property-search-toolbar {
  display: grid;
  grid-template-columns: minmax(4.75rem, 6.5rem) minmax(0, 1fr);
  align-items: start;
  gap: 0.35rem 0.45rem;
  padding: 0.35rem 0;
}

.property-search-toolbar {
  grid-template-columns: minmax(4.75rem, 6.5rem) minmax(10rem, 1fr) auto auto auto auto auto minmax(11rem, 15rem);
}

.property-filter-panel > .property-filter-section--price,
.property-filter-panel > .property-filter-section--area {
  align-items: start;
}

.property-filter-results {
  min-width: 0;
  overflow: visible;
}

.property-results-topline {
  display: grid;
  gap: 1.25rem;
  margin-bottom: 2rem;
}

.property-results-topline p,
.property-results-topline h2 {
  margin: 0;
}

.property-results-topline p {
  font-size: 1.125rem;
  line-height: 1.6;
  color: rgb(var(--color-text-muted));
}

.property-results-topline h2 {
  display: none;
}

.property-results-heading {
  display: grid;
  gap: 1rem;
}

.property-results-side {
  display: grid;
  justify-items: start;
  gap: 0.75rem;
}

.property-results-toolbar {
  display: flex;
  min-width: 0;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.75rem 1rem;
}

.property-results-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
}

.property-results-actions .property-button {
  min-height: 2.25rem;
  border-radius: 999px;
  font-size: 0.82rem;
}

.property-control-group {
  display: inline-flex;
  overflow: hidden;
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.75rem;
  background: rgb(var(--color-surface-raised));
}

.property-control-button {
  display: inline-flex;
  min-height: 2.25rem;
  align-items: center;
  justify-content: center;
  padding: 0 0.85rem;
  color: rgb(var(--color-text-muted));
  font-size: 0.82rem;
  font-weight: 600;
  letter-spacing: 0.02em;
  line-height: 1;
  white-space: nowrap;
  transition:
    background-color 0.2s ease,
    color 0.2s ease;
}

.property-control-button--icon {
  display: inline-flex;
  width: 2.25rem;
  height: 2.25rem;
  min-height: 2.25rem;
  align-items: center;
  justify-content: center;
  padding: 0;
}

.property-control-button:focus-visible {
  outline: 2px solid rgb(var(--color-primary));
  outline-offset: 2px;
}

.property-control-button--active {
  background: rgb(var(--color-primary));
  color: rgb(var(--color-primary-contrast));
}

.property-control-button--idle:hover {
  color: rgb(var(--color-primary));
}

.property-control-button--icon,
.property-control-group--icon .property-control-button {
  transition:
    background-color 0.2s ease,
    color 0.2s ease;
}

.property-chip-grid,
.property-card__tags {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.property-card__tags {
  max-height: 4.5rem;
  overflow: hidden;
}

.property-card__tags span {
  border-radius: 999px;
  border: 1px solid rgb(var(--color-border));
  background: rgb(var(--color-surface-raised));
  padding: 0.35rem 0.65rem;
  color: rgb(var(--color-primary));
  font-size: 0.74rem;
  font-weight: 600;
}

.property-button,
.property-filter-chip {
  display: inline-flex;
  min-height: 2.5rem;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  border-radius: 999px;
  padding: 0 0.85rem;
  font-size: 0.875rem;
  font-weight: 500;
  letter-spacing: 0.02em;
}

.property-button--primary {
  border: 1px solid rgb(var(--color-primary));
  background: rgb(var(--color-primary));
  color: rgb(var(--color-primary-contrast));
}

.property-button--secondary,
.property-filter-chip {
  border: 1px solid rgb(var(--color-border));
  background: transparent;
  color: rgb(var(--color-text));
}

.property-filter-chip {
  min-height: 2rem;
  padding: 0 0.8rem;
  color: rgb(var(--color-text-muted));
  font-size: 0.78rem;
}

.property-filter-chip--active {
  border-color: rgb(var(--color-primary));
  background: rgb(0 39 39 / 0.05);
  color: rgb(var(--color-primary));
}

.property-filter-section {
  display: grid;
  gap: 0.35rem;
}

.property-option-list {
  display: grid;
  gap: 0.22rem;
}

.property-option-list--inline {
  display: flex;
  flex-wrap: wrap;
  gap: 0.12rem;
}

.property-option-row {
  display: flex;
  width: 100%;
  cursor: pointer;
  align-items: center;
  gap: 0.5rem;
  color: rgb(var(--color-text-muted));
  font-size: 0.95rem;
  line-height: 1.35;
  text-align: left;
  transition:
    background-color 0.2s ease,
    border-color 0.2s ease,
    color 0.2s ease;
}

.property-option-row--inline {
  width: auto;
}

.property-option-list--inline .property-option-row {
  min-height: 1.75rem;
  width: auto;
  justify-content: center;
  gap: 0;
  border: 1px solid transparent;
  border-radius: 999px;
  padding: 0 0.38rem;
  color: rgb(var(--color-text-muted));
  font-size: 0.84rem;
  font-weight: 600;
  line-height: 1;
}

.property-option-list--inline .property-option-box {
  display: none;
}

.property-price-filter {
  min-width: 0;
}

.property-price-filter--serviced {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr);
  align-items: start;
  gap: 0.2rem 0.45rem;
}

.property-price-filter__mode,
.property-price-filter__range {
  min-width: 0;
}

.property-price-filter__mode {
  flex-wrap: nowrap;
}

.property-price-filter__range {
  align-content: start;
}

.property-price-filter__mode .property-option-row,
.property-price-filter__range .property-option-row {
  white-space: nowrap;
}

.property-option-list--inline .property-option-row:hover {
  color: rgb(var(--color-text));
}

.property-option-list--inline .property-option-row:focus-visible {
  border-color: rgb(var(--color-primary));
  outline: 2px solid rgb(var(--color-primary));
  outline-offset: 2px;
}

.property-option-list--inline .property-option-row--active {
  border-color: rgb(var(--color-primary));
  background: rgb(0 39 39 / 0.05);
  color: rgb(var(--color-primary));
}

.property-option-row:not(.property-option-row--inline):hover,
.property-option-row--active:not(.property-option-row--inline) {
  color: rgb(var(--color-text));
}

.property-option-box {
  display: inline-flex;
  width: 1rem;
  height: 1rem;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.25rem;
  background: rgb(var(--color-surface-raised));
}

.property-option-row--active .property-option-box {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary));
  box-shadow: inset 0 0 0 3px rgb(var(--color-surface-raised));
}

.property-field {
  display: grid;
  gap: 0.55rem;
}

.property-field select,
.property-field input,
.property-search-input {
  min-height: 2rem;
  border: 0;
  border-radius: 0.25rem;
  background: rgb(var(--color-surface-muted));
  color: rgb(var(--color-text));
  font-size: 0.95rem;
  line-height: 1.35;
}

.property-field select,
.property-field input {
  padding: 0 0.7rem;
}

.property-field input:disabled {
  opacity: 0.55;
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

.property-search-toolbar > span {
  color: rgb(var(--color-text-muted));
  font-size: 0.88rem;
  font-weight: 700;
  white-space: nowrap;
}

.property-toolbar-button,
.property-toolbar-icon-button {
  display: inline-flex;
  min-height: 2rem;
  align-items: center;
  justify-content: center;
  border-radius: 0.25rem;
  font-size: 0.88rem;
  font-weight: 700;
  line-height: 1;
  white-space: nowrap;
}

.property-toolbar-button {
  padding: 0 0.85rem;
}

.property-toolbar-button--primary {
  background: rgb(var(--color-primary));
  color: rgb(var(--color-primary-contrast));
}

.property-toolbar-icon-button {
  position: relative;
  width: 2rem;
  border: 1px solid rgb(var(--color-border));
  color: rgb(var(--color-text-muted));
}

.property-toolbar-icon-button:hover,
.property-toolbar-icon-button:focus-visible,
.property-toolbar-icon-button--active {
  border-color: rgb(var(--color-primary));
  background: rgb(0 39 39 / 0.05);
  color: rgb(var(--color-primary));
}

.property-toolbar-icon-button:focus-visible {
  outline: 2px solid rgb(var(--color-primary));
  outline-offset: 2px;
}

.property-toolbar-icon-button::after {
  position: absolute;
  bottom: calc(100% + 0.45rem);
  left: 50%;
  z-index: 10;
  max-width: 8rem;
  border-radius: 0.25rem;
  background: rgb(var(--color-text));
  padding: 0.35rem 0.5rem;
  color: rgb(var(--color-surface-raised));
  content: attr(data-tooltip);
  font-size: 0.72rem;
  font-weight: 700;
  line-height: 1.2;
  opacity: 0;
  pointer-events: none;
  text-align: center;
  transform: translate(-50%, 0.2rem);
  transition:
    opacity 0.16s ease,
    transform 0.16s ease;
  white-space: nowrap;
}

.property-toolbar-icon-button:hover::after,
.property-toolbar-icon-button:focus-visible::after {
  opacity: 1;
  transform: translate(-50%, 0);
}

.property-sort-menu {
  position: relative;
  display: flex;
  min-width: 0;
  align-items: center;
}

.property-sort-menu :deep(.app-glass-select__trigger) {
  width: 100%;
  min-height: 2rem;
  cursor: pointer;
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.25rem;
  background: rgb(var(--color-surface-muted));
  padding: 0 1.9rem 0 0.65rem;
  color: rgb(var(--color-text));
  font-size: 0.88rem;
  font-weight: 700;
  line-height: 1;
  box-shadow: none;
}

.property-sort-menu :deep(.app-glass-select__trigger:hover),
.property-sort-menu :deep(.app-glass-select__trigger[aria-expanded='true']) {
  border-color: rgb(var(--color-border));
  box-shadow: none;
}

.property-sort-menu :deep(.app-glass-select__trigger svg) {
  position: absolute;
  right: 0.55rem;
  color: rgb(var(--color-text-muted));
  pointer-events: none;
}

.property-sort-menu :deep(.app-glass-select__trigger span) {
  font-size: 0.88rem;
  font-weight: 700;
  line-height: 1;
}

.property-tag-stack {
  display: grid;
  gap: 0;
  border-top: 1px solid rgb(var(--color-border));
  padding-top: 0.2rem;
}

.property-tag-group {
  display: grid;
  grid-template-columns: minmax(4.75rem, 6.5rem) minmax(0, 1fr);
  gap: 0.35rem 0.45rem;
  align-items: start;
  padding: 0.35rem 0;
}

.property-tag-group .property-option-list {
  display: flex;
  flex-wrap: wrap;
  gap: 0.12rem;
}

.property-tag-group .property-option-row {
  min-height: 1.75rem;
  width: auto;
  justify-content: center;
  gap: 0;
  border: 1px solid transparent;
  border-radius: 999px;
  padding: 0 0.38rem;
  color: rgb(var(--color-text-muted));
  font-size: 0.84rem;
  font-weight: 600;
  line-height: 1;
}

.property-tag-group .property-option-box {
  display: none;
}

.property-tag-group .property-option-row:hover {
  color: rgb(var(--color-text));
}

.property-tag-group .property-option-row--active {
  border-color: rgb(var(--color-primary));
  background: rgb(0 39 39 / 0.05);
  color: rgb(var(--color-primary));
}

.property-grid {
  display: grid;
  gap: 2rem;
}

.property-grid--list {
  grid-template-columns: minmax(0, 1fr);
  gap: 1.25rem;
}

.property-card {
  display: grid;
  overflow: hidden;
  height: 25.125rem;
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.75rem;
  background: rgb(var(--color-surface-raised));
  color: inherit;
  box-shadow: 0 4px 24px rgb(0 0 0 / 0.03);
  transition: box-shadow 0.3s ease;
}

.property-card:hover {
  box-shadow: 0 8px 32px rgb(0 0 0 / 0.06);
}

.property-card__media {
  position: relative;
  aspect-ratio: 4 / 3;
  min-height: 0;
  overflow: hidden;
  background: rgb(var(--color-surface-raised));
}

.property-card__ad-badge {
  position: absolute;
  top: 0.65rem;
  left: 0.65rem;
  z-index: 2;
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.35rem;
  background: rgb(var(--color-surface) / 0.92);
  padding: 0.25rem 0.45rem;
  color: rgb(var(--color-text));
  font-size: 0.72rem;
  font-weight: 950;
  line-height: 1;
}

.property-card__ad-badge--gold {
  border-color: rgb(190 137 25 / 0.55);
  background: rgb(255 243 199 / 0.96);
  color: rgb(126 74 10);
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
  grid-template-rows: auto auto auto auto minmax(0, 1fr) auto;
  min-width: 0;
  min-height: 0;
  gap: 0.75rem;
  padding: 0.875rem 1rem;
}

.property-card__meta,
.property-card__footer {
  display: flex;
  flex-wrap: wrap;
  min-width: 0;
  min-height: 0;
  gap: 0.45rem;
  color: rgb(var(--color-text-muted));
  font-size: 0.75rem;
  font-weight: 800;
}

.property-card__meta span,
.property-card__footer span {
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.property-card h2 {
  margin: 0;
  overflow: hidden;
  color: rgb(var(--color-text));
  font-family: var(--font-display);
  font-size: 1.125rem;
  font-weight: 500;
  line-height: 1.28;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.property-card p {
  display: -webkit-box;
  min-height: 2.7rem;
  margin: 0;
  overflow: hidden;
  color: rgb(var(--color-text-muted));
  font-size: 0.875rem;
  line-height: 1.45;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.property-card__specs {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  min-height: 0;
  gap: 0.65rem;
  margin: 0;
  overflow: hidden;
}

.property-card__specs dt {
  color: rgb(var(--color-text-muted));
  font-size: 0.7rem;
  font-weight: 600;
}

.property-card__specs dd {
  margin: 0.18rem 0 0;
  overflow: hidden;
  font-size: 0.84rem;
  font-weight: 700;
  line-height: 1.35;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.property-card__footer {
  align-self: end;
  justify-content: space-between;
  border-top: 1px solid rgb(var(--color-border));
  padding-top: 0.75rem;
}

.property-empty {
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface-raised));
  padding: 2rem;
  color: rgb(var(--color-text-muted));
  text-align: center;
}

@media (max-width: 767px) {
  .property-list-page {
    padding: 1.25rem var(--layout-page-padding-inline) 3rem;
  }

  .property-top-layout,
  .property-content-layout {
    grid-template-columns: minmax(0, 1fr);
  }

  .property-filter-panel > .property-field,
  .property-filter-panel > .property-filter-section,
  .property-tag-group,
  .property-search-toolbar {
    grid-template-columns: minmax(0, 1fr);
  }

  .property-price-filter--serviced {
    grid-template-columns: minmax(0, 1fr);
  }

}

@media (min-width: 640px) {
  .property-grid--grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (min-width: 768px) {
  .property-results-heading {
    grid-template-columns: minmax(0, 1fr) auto;
    align-items: center;
  }

  .property-results-side {
    justify-items: end;
  }

  .property-results-toolbar {
    justify-content: flex-end;
  }

  .property-card--list {
    height: auto;
    min-height: 6.875rem;
    grid-template-columns: 10.25rem minmax(0, 1fr);
  }

  .property-card--list .property-card__media {
    height: 100%;
    min-height: 6.875rem;
    align-self: stretch;
    aspect-ratio: auto;
    border-radius: 0.75rem 0 0 0.75rem;
  }

  .property-card--list .property-card__body {
    grid-template-rows: auto;
    min-height: 6.875rem;
    gap: 0.24rem;
    padding: 0.42rem 0.62rem;
  }

  .property-card--list p {
    display: -webkit-box;
    min-height: 0;
    -webkit-line-clamp: 1;
  }

  .property-card--list .property-card__tags {
    max-height: 1.45rem;
  }

  .property-card--list .property-card__specs {
    gap: 0.18rem 0.4rem;
  }

  .property-card--list .property-card__footer {
    padding-top: 0.28rem;
  }
}

@media (min-width: 1280px) {
  .property-grid--grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}
</style>

/*
 * 二手交易篩選 - 狀態與資料。
 * 1. 讀取真實公開帖子列表並同步 URL query。
 * 2. 統一分類、地區、價格、成色、排序與分頁條件。
 */
import axios from 'axios';
import { computed, onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import { fetchSecondhandListings } from '@/httpapis/secondhand-listings';
import {
  buildMarketplaceAreaFilterOptions,
  getMarketplaceCategoryLabel,
  getMarketplaceConditionLabel,
  getMarketplaceDistrictLabel,
  getMarketplaceOptionLabel,
  marketplaceCategories,
  marketplaceConditions,
  type MarketplaceCategoryCode,
  type MarketplaceConditionCode,
} from '@/constants/marketplace';
import type { ListingListParams, SecondhandListingSummaryResponse } from '@/model/marketplace';
import { useFeedbackStore } from '@/stores/feedback';
import type { AppLocale } from '@/stores/preferences';
import { usePreferenceStore } from '@/stores/preferences';
import { formatPrice } from '@/utils/format';

interface FilterCategoryItem {
  key: MarketplaceCategoryCode;
  label: string;
  count: number;
}

interface FilterListingCard {
  id: string;
  title: string;
  summary: string;
  price: string;
  rawPrice: number;
  categoryCode: string;
  conditionLevel: string;
  districtCode: string;
  districtLabel: string;
  imageUrl?: string;
}

type FilterSortValue = 'latest' | 'price_asc' | 'price_desc';
type FilterAreaValue = string;

const defaultPageSize = 12;

// 1. 讀取單一 query 字串
const readQueryString = (value: unknown): string =>
  typeof value === 'string' ? value : Array.isArray(value) && typeof value[0] === 'string' ? value[0] : '';

// 2. 建立帖子卡片資料
const buildListingCard = (
  listing: SecondhandListingSummaryResponse,
  locale: AppLocale,
  translate: (key: string) => string,
): FilterListingCard => {
  const price = listing.price_mode === 'free'
    ? translate('common.price.free')
    : formatPrice(listing.price_hkd ?? 0, locale);

  return {
    id: listing.listing_id,
    title: listing.title,
    summary: listing.summary,
    price,
    rawPrice: listing.price_hkd ?? 0,
    categoryCode: listing.category_code,
    conditionLevel: listing.condition_level,
    districtCode: listing.district_code,
    districtLabel: getMarketplaceDistrictLabel(listing.district_code, locale),
    imageUrl: listing.cover_image?.url,
  };
};

// 3. 管理篩選頁資料與動作
export const useMarketplaceFilterPage = () => {
  const route = useRoute();
  const router = useRouter();
  const { t } = useI18n();
  const feedbackStore = useFeedbackStore();
  const preferenceStore = usePreferenceStore();
  const loading = ref(false);
  const keyword = ref('');
  const minPrice = ref('');
  const maxPrice = ref('');
  const sortBy = ref<FilterSortValue>('latest');
  const area = ref<FilterAreaValue>('all');
  const withPhotos = ref(true);
  const selectedConditions = ref<MarketplaceConditionCode[]>([]);
  const selectedCategoryKeys = ref<MarketplaceCategoryCode[]>([]);
  const sourceListings = ref<SecondhandListingSummaryResponse[]>([]);
  const page = ref(1);
  const totalResults = ref(0);
  const totalPages = computed(() => Math.max(1, Math.ceil(totalResults.value / defaultPageSize)));

  const sortOptions = computed<Array<{ label: string; value: FilterSortValue }>>(() => [
    { label: t('marketplace.filter.sortNewest'), value: 'latest' },
    { label: t('marketplace.filter.sortPriceAsc'), value: 'price_asc' },
    { label: t('marketplace.filter.sortPriceDesc'), value: 'price_desc' },
  ]);

  const areaOptions = computed(() =>
    buildMarketplaceAreaFilterOptions(preferenceStore.locale, t('marketplace.filter.areaAll')).map((option) => ({
      label: option.label_zh_hk,
      value: option.value,
      filterType: option.filterType,
      regionCode: option.regionCode,
      districtCode: option.districtCode,
    })),
  );

  const conditionOptions = computed<Array<{ label: string; value: MarketplaceConditionCode }>>(() =>
    marketplaceConditions.map((condition) => ({
      label: getMarketplaceOptionLabel(condition, preferenceStore.locale),
      value: condition.value,
    })),
  );

  const listings = computed(() =>
    sourceListings.value
      .map((listing) => buildListingCard(listing, preferenceStore.locale, t))
      .filter((listing) => !withPhotos.value || Boolean(listing.imageUrl)),
  );

  const categories = computed<FilterCategoryItem[]>(() =>
    marketplaceCategories.map((category) => ({
      key: category.value,
      label: getMarketplaceCategoryLabel(category.value, preferenceStore.locale),
      count: sourceListings.value.filter((listing) => listing.category_code === category.value).length,
    })),
  );

  // 3.1 從 URL query 同步本地條件
  const syncStateFromQuery = (): void => {
    keyword.value = readQueryString(route.query.keyword);
    minPrice.value = readQueryString(route.query.min_price_hkd);
    maxPrice.value = readQueryString(route.query.max_price_hkd);
    sortBy.value = (readQueryString(route.query.sort_by) || 'latest') as FilterSortValue;
    page.value = Number(readQueryString(route.query.page) || 1);
    selectedCategoryKeys.value = readQueryString(route.query.category_code)
      ? [readQueryString(route.query.category_code) as MarketplaceCategoryCode]
      : [];
    selectedConditions.value = readQueryString(route.query.condition_level)
      ? [readQueryString(route.query.condition_level) as MarketplaceConditionCode]
      : [];

    const regionCode = readQueryString(route.query.region_code);
    const districtCode = readQueryString(route.query.district_code);
    area.value = districtCode || regionCode || 'all';
  };

  // 3.2 建立 API 查詢參數
  const buildListParams = (): ListingListParams => {
    const selectedArea = areaOptions.value.find((option) => option.value === area.value);
    const minPriceNumber = Number(minPrice.value);
    const maxPriceNumber = Number(maxPrice.value);

    return {
      page: page.value,
      page_size: defaultPageSize,
      keyword: keyword.value.trim() || undefined,
      category_code: selectedCategoryKeys.value[0],
      region_code: selectedArea?.filterType === 'region' ? selectedArea.regionCode : undefined,
      district_code: selectedArea?.filterType === 'district' ? selectedArea.districtCode : undefined,
      condition_level: selectedConditions.value[0],
      min_price_hkd: Number.isFinite(minPriceNumber) && minPrice.value ? minPriceNumber : undefined,
      max_price_hkd: Number.isFinite(maxPriceNumber) && maxPrice.value ? maxPriceNumber : undefined,
      sort_by: sortBy.value,
    };
  };

  // 3.3 讀取公開帖子
  const loadListings = async (): Promise<void> => {
    loading.value = true;

    try {
      const { data } = await fetchSecondhandListings(buildListParams());
      sourceListings.value = data.data.items;
      totalResults.value = data.data.pagination.total;
    } catch (error) {
      feedbackStore.pushToast(
        axios.isAxiosError(error)
          ? error.response?.data?.message ?? t('marketplace.filter.loadError')
          : t('marketplace.filter.loadError'),
        'error',
      );
      sourceListings.value = [];
      totalResults.value = 0;
    } finally {
      loading.value = false;
    }
  };

  // 3.4 將目前條件同步到 URL
  const syncFiltersToQuery = async (nextPage = 1): Promise<void> => {
    const selectedArea = areaOptions.value.find((option) => option.value === area.value);

    await router.replace({
      query: {
        ...route.query,
        keyword: keyword.value.trim() || undefined,
        category_code: selectedCategoryKeys.value[0] || undefined,
        condition_level: selectedConditions.value[0] || undefined,
        region_code: selectedArea?.filterType === 'region' ? selectedArea.regionCode : undefined,
        district_code: selectedArea?.filterType === 'district' ? selectedArea.districtCode : undefined,
        min_price_hkd: minPrice.value || undefined,
        max_price_hkd: maxPrice.value || undefined,
        sort_by: sortBy.value === 'latest' ? undefined : sortBy.value,
        page: nextPage > 1 ? String(nextPage) : undefined,
      },
    });
  };

  // 3.5 切換主分類勾選
  const toggleCategory = async (key: MarketplaceCategoryCode): Promise<void> => {
    selectedCategoryKeys.value = selectedCategoryKeys.value.includes(key) ? [] : [key];
    await syncFiltersToQuery();
  };

  // 3.6 切換成色條件
  const toggleCondition = async (value: MarketplaceConditionCode): Promise<void> => {
    selectedConditions.value = selectedConditions.value.includes(value) ? [] : [value];
    await syncFiltersToQuery();
  };

  // 3.7 清除目前篩選條件
  const clearFilters = async (): Promise<void> => {
    keyword.value = '';
    minPrice.value = '';
    maxPrice.value = '';
    selectedCategoryKeys.value = [];
    selectedConditions.value = [];
    area.value = 'all';
    sortBy.value = 'latest';
    withPhotos.value = true;
    await syncFiltersToQuery();
  };

  // 3.8 切換分頁
  const setPage = async (value: number): Promise<void> => {
    page.value = Math.min(Math.max(value, 1), totalPages.value);
    await syncFiltersToQuery(page.value);
  };

  watch(
    () => route.query,
    () => {
      syncStateFromQuery();
      void loadListings();
    },
    { immediate: true },
  );

  onMounted(() => {
    syncStateFromQuery();
  });

  return {
    area,
    areaOptions,
    categories,
    clearFilters,
    conditionOptions,
    getMarketplaceConditionLabel,
    keyword,
    listings,
    loading,
    maxPrice,
    minPrice,
    page,
    selectedCategoryKeys,
    selectedConditions,
    setPage,
    sortBy,
    sortOptions,
    syncFiltersToQuery,
    t,
    toggleCategory,
    toggleCondition,
    totalPages,
    totalResults,
    withPhotos,
  };
};

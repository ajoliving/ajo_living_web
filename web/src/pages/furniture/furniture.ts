/*
 * 二手交易篩選 - 狀態與資料。
 * 1. 讀取真實公開帖子列表並同步 URL query。
 * 2. 統一分類、地區、價格、成色、排序與分頁條件。
 */
import axios from 'axios';
import { computed, onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import {
  favoriteSecondhandListing,
  fetchSecondhandListings,
  unfavoriteSecondhandListing,
} from '@/domains/marketplace/api';
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
} from '@/domains/marketplace/constants';
import type { ListingListParams, SecondhandListingSummaryResponse } from '@/domains/marketplace/model';
import { useFeedbackStore } from '@/app/stores/feedback';
import type { AppLocale } from '@/app/stores/preferences';
import { usePreferenceStore } from '@/app/stores/preferences';
import { useSessionStore } from '@/app/stores/session';
import { formatPrice } from '@/shared/utils/format';

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
  categoryLabel: string;
  conditionLevel: string;
  conditionLabel: string;
  districtCode: string;
  districtLabel: string;
  imageUrl?: string;
  isFavorited: boolean;
  publishedAt?: string | null;
  updatedAt: string;
}

type FilterSortValue = 'latest' | 'price_asc' | 'price_desc';
type FilterAreaValue = string;
type FilterVisibilityValue = 'all' | 'public' | 'building_only';

const defaultPageSize = 12;
const maxFilterPrice = 20000;

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
    categoryLabel: getMarketplaceCategoryLabel(listing.category_code, locale),
    conditionLevel: listing.condition_level,
    conditionLabel: getMarketplaceConditionLabel(listing.condition_level, locale),
    districtCode: listing.district_code,
    districtLabel: getMarketplaceDistrictLabel(listing.district_code, locale),
    imageUrl: listing.cover_image?.url,
    isFavorited: listing.is_favorited,
    publishedAt: listing.published_at,
    updatedAt: listing.updated_at,
  };
};

// 3. 管理篩選頁資料與動作
export const useFurniturePage = () => {
  const route = useRoute();
  const router = useRouter();
  const { t } = useI18n();
  const feedbackStore = useFeedbackStore();
  const preferenceStore = usePreferenceStore();
  const sessionStore = useSessionStore();
  const loading = ref(false);
  const favoriteUpdatingIds = ref<string[]>([]);
  const keyword = ref('');
  const minPrice = ref('0');
  const maxPrice = ref(String(maxFilterPrice));
  const sortBy = ref<FilterSortValue>('latest');
  const area = ref<FilterAreaValue>('all');
  const visibility = ref<FilterVisibilityValue>('all');
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

  const visibilityOptions = computed<Array<{ label: string; value: FilterVisibilityValue }>>(() => [
    { label: t('marketplace.filter.visibilityAll'), value: 'all' },
    { label: t('marketplace.filter.visibilityPublic'), value: 'public' },
    { label: t('marketplace.filter.visibilityBuildingOnly'), value: 'building_only' },
  ]);

  const listings = computed(() =>
    sourceListings.value.map((listing) => buildListingCard(listing, preferenceStore.locale, t)),
  );
  const isFavoriteUpdating = (listingID: string): boolean =>
    favoriteUpdatingIds.value.includes(listingID);

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
    minPrice.value = readQueryString(route.query.min_price_hkd) || '0';
    maxPrice.value = readQueryString(route.query.max_price_hkd) || String(maxFilterPrice);
    sortBy.value = (readQueryString(route.query.sort_by) || 'latest') as FilterSortValue;
    page.value = Number(readQueryString(route.query.page) || 1);
    const visibilityScope = readQueryString(route.query.visibility_scope);
    visibility.value = visibilityScope === 'public' || visibilityScope === 'building_only'
      ? visibilityScope
      : 'all';
    withPhotos.value = readQueryString(route.query.has_media) === 'true';
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
      min_price_hkd: Number.isFinite(minPriceNumber) && minPriceNumber > 0 ? minPriceNumber : undefined,
      max_price_hkd: Number.isFinite(maxPriceNumber) && maxPriceNumber < maxFilterPrice ? maxPriceNumber : undefined,
      has_media: withPhotos.value ? true : undefined,
      visibility_scope: visibility.value === 'all' ? undefined : visibility.value,
      only_building: visibility.value === 'building_only' ? true : undefined,
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

  // 3.4 切換帖子收藏狀態
  const toggleFavorite = async (listingID: string): Promise<void> => {
    const target = sourceListings.value.find((listing) => listing.listing_id === listingID);
    if (!target || isFavoriteUpdating(listingID)) {
      return;
    }
    if (!sessionStore.isAuthenticated) {
      await router.push({
        path: '/login',
        query: { redirect: route.fullPath },
      });
      return;
    }

    favoriteUpdatingIds.value = [...favoriteUpdatingIds.value, listingID];

    try {
      const { data } = target.is_favorited
        ? await unfavoriteSecondhandListing(listingID)
        : await favoriteSecondhandListing(listingID);
      sourceListings.value = sourceListings.value.map((listing) =>
        listing.listing_id === listingID
          ? { ...listing, is_favorited: data.data.is_favorited }
          : listing,
      );
      feedbackStore.pushToast(
        data.data.is_favorited
          ? t('marketplace.detail.favoriteAdded')
          : t('marketplace.detail.favoriteRemoved'),
        'success',
      );
    } catch (error) {
      feedbackStore.pushToast(
        axios.isAxiosError(error)
          ? error.response?.data?.message ?? t('marketplace.detail.favoriteError')
          : t('marketplace.detail.favoriteError'),
        'error',
      );
    } finally {
      favoriteUpdatingIds.value = favoriteUpdatingIds.value.filter((id) => id !== listingID);
    }
  };

  // 3.5 將目前條件同步到 URL
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
        min_price_hkd: Number(minPrice.value) > 0 ? minPrice.value : undefined,
        max_price_hkd: Number(maxPrice.value) < maxFilterPrice ? maxPrice.value : undefined,
        has_media: withPhotos.value ? 'true' : undefined,
        visibility_scope: visibility.value === 'all' ? undefined : visibility.value,
        only_building: undefined,
        sort_by: sortBy.value === 'latest' ? undefined : sortBy.value,
        page: nextPage > 1 ? String(nextPage) : undefined,
      },
    });
  };

  // 3.6 切換主分類勾選
  const toggleCategory = async (key: MarketplaceCategoryCode): Promise<void> => {
    selectedCategoryKeys.value = selectedCategoryKeys.value.includes(key) ? [] : [key];
    await syncFiltersToQuery();
  };

  // 3.7 切換成色條件
  const toggleCondition = async (value: MarketplaceConditionCode): Promise<void> => {
    selectedConditions.value = selectedConditions.value.includes(value) ? [] : [value];
    await syncFiltersToQuery();
  };

  // 3.8 清除成色條件
  const clearConditions = async (): Promise<void> => {
    selectedConditions.value = [];
    await syncFiltersToQuery();
  };

  // 3.9 清除目前篩選條件
  const clearFilters = async (): Promise<void> => {
    keyword.value = '';
    minPrice.value = '0';
    maxPrice.value = String(maxFilterPrice);
    selectedCategoryKeys.value = [];
    selectedConditions.value = [];
    area.value = 'all';
    visibility.value = 'all';
    sortBy.value = 'latest';
    withPhotos.value = false;
    await syncFiltersToQuery();
  };

  // 3.10 切換分頁
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
    clearConditions,
    conditionOptions,
    getMarketplaceConditionLabel,
    isFavoriteUpdating,
    keyword,
    listings,
    loading,
    maxPrice,
    maxPriceLimit: maxFilterPrice,
    minPrice,
    page,
    selectedCategoryKeys,
    selectedConditions,
    setPage,
    sortBy,
    sortOptions,
    syncFiltersToQuery,
    t,
    toggleFavorite,
    toggleCategory,
    toggleCondition,
    totalPages,
    totalResults,
    visibility,
    visibilityOptions,
    withPhotos,
  };
};

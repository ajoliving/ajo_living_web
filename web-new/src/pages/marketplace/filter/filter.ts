/*
 * 二手交易篩選 - 狀態與資料。
 * 1. 讀取真實公開帖子列表並同步 URL query。
 * 2. 統一分類、地區、價格、成色、排序與分頁條件。
 */
import axios from 'axios';
import { computed, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import {
  favoriteSecondhandListing,
  fetchSecondhandListings,
  unfavoriteSecondhandListing,
} from '@/httpapis/secondhand-listings';
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
import { useSessionStore } from '@/stores/session';
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
  isFavorited: boolean;
}

type FilterSortValue = 'latest' | 'price_asc' | 'price_desc';
type FilterAreaValue = string;

const defaultPageSize = 12;

// 1. 讀取單一 query 字串
const readQueryString = (value: unknown): string =>
  typeof value === 'string' ? value : Array.isArray(value) && typeof value[0] === 'string' ? value[0] : '';

// 2. 檢查 query enum 是否仍然可用
const isMarketplaceCategoryCode = (value: string): value is MarketplaceCategoryCode =>
  marketplaceCategories.some((category) => category.value === value);

const isMarketplaceConditionCode = (value: string): value is MarketplaceConditionCode =>
  marketplaceConditions.some((condition) => condition.value === value);

// 3. 建立帖子卡片資料
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
    isFavorited: listing.is_favorited,
  };
};

// 4. 管理篩選頁資料與動作
export const useMarketplaceFilterPage = () => {
  const route = useRoute();
  const router = useRouter();
  const { t } = useI18n();
  const feedbackStore = useFeedbackStore();
  const preferenceStore = usePreferenceStore();
  const sessionStore = useSessionStore();
  const loading = ref(false);
  const favoriteUpdatingIds = ref<string[]>([]);
  const keyword = ref('');
  const minPrice = ref('');
  const maxPrice = ref('');
  const sortBy = ref<FilterSortValue>('latest');
  const area = ref<FilterAreaValue>('all');
  const withPhotos = ref(false);
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

  // 4.1 從 URL query 同步本地條件
  const syncStateFromQuery = (): void => {
    const categoryCode = readQueryString(route.query.category_code);
    const conditionLevel = readQueryString(route.query.condition_level);

    keyword.value = readQueryString(route.query.keyword);
    minPrice.value = readQueryString(route.query.min_price_hkd);
    maxPrice.value = readQueryString(route.query.max_price_hkd);
    sortBy.value = (readQueryString(route.query.sort_by) || 'latest') as FilterSortValue;
    page.value = Number(readQueryString(route.query.page) || 1);
    selectedCategoryKeys.value = isMarketplaceCategoryCode(categoryCode) ? [categoryCode] : [];
    selectedConditions.value = isMarketplaceConditionCode(conditionLevel) ? [conditionLevel] : [];

    const regionCode = readQueryString(route.query.region_code);
    const districtCode = readQueryString(route.query.district_code);
    area.value = districtCode || regionCode || 'all';
    withPhotos.value = readQueryString(route.query.has_media) === 'true';
  };

  // 4.2 建立 API 查詢參數
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
      has_media: withPhotos.value || undefined,
      sort_by: sortBy.value,
    };
  };

  // 4.3 讀取公開帖子
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

  // 4.4 切換帖子收藏狀態
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

  // 4.5 將目前條件同步到 URL
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
        has_media: withPhotos.value ? 'true' : undefined,
        sort_by: sortBy.value === 'latest' ? undefined : sortBy.value,
        page: nextPage > 1 ? String(nextPage) : undefined,
      },
    });
  };

  // 4.6 切換主分類勾選
  const toggleCategory = async (key: MarketplaceCategoryCode): Promise<void> => {
    selectedCategoryKeys.value = selectedCategoryKeys.value.includes(key) ? [] : [key];
    await syncFiltersToQuery();
  };

  // 4.7 切換成色條件
  const toggleCondition = async (value: MarketplaceConditionCode): Promise<void> => {
    selectedConditions.value = selectedConditions.value.includes(value) ? [] : [value];
    await syncFiltersToQuery();
  };

  // 4.8 清除目前篩選條件
  const clearFilters = async (): Promise<void> => {
    keyword.value = '';
    minPrice.value = '';
    maxPrice.value = '';
    selectedCategoryKeys.value = [];
    selectedConditions.value = [];
    area.value = 'all';
    sortBy.value = 'latest';
    withPhotos.value = false;
    await syncFiltersToQuery();
  };

  // 4.9 切換分頁
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

  return {
    area,
    areaOptions,
    categories,
    clearFilters,
    conditionOptions,
    getMarketplaceConditionLabel,
    isFavoriteUpdating,
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
    toggleFavorite,
    toggleCategory,
    toggleCondition,
    totalPages,
    totalResults,
    withPhotos,
  };
};

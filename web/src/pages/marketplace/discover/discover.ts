/*
 * 二手交易 Discover - 展示資料。
 * 1. 讀取真實公開帖子並建立推薦好物。
 * 2. 依統一分類輸出輪播圖與橫向帖子列表。
 */
import axios from 'axios';
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';

import { fetchSecondhandListings } from '@/httpapis/secondhand-listings';
import {
  getMarketplaceCategoryLabel,
  marketplaceCategories,
  type MarketplaceCategoryCode,
} from '@/constants/marketplace';
import type { SecondhandListingSummaryResponse } from '@/model/marketplace';
import { useFeedbackStore } from '@/stores/feedback';
import { usePreferenceStore } from '@/stores/preferences';
import { formatPrice } from '@/utils/format';

interface DiscoverListingPreview {
  key: string;
  title: string;
  summary: string;
  price: string;
  imageUrl: string;
  imageAlt: string;
  to: string;
}

interface DiscoverCategoryRow {
  key: MarketplaceCategoryCode;
  label: string;
  count: number;
  filterTo: string;
  heroSlides: DiscoverListingPreview[];
  listings: DiscoverListingPreview[];
}

// 1. 建立展示卡片資料
const buildListingPreview = (
  listing: SecondhandListingSummaryResponse,
  locale: string,
): DiscoverListingPreview => ({
  key: listing.listing_id,
  title: listing.title,
  summary: listing.summary,
  price: listing.price_mode === 'free' ? locale === 'zh-HK' ? '免費' : 'Free' : formatPrice(listing.price_hkd ?? 0, locale),
  imageUrl: listing.cover_image?.url ?? '',
  imageAlt: listing.title,
  to: `/marketplace/listing/${listing.listing_id}`,
});

// 2. 管理 Discover 頁展示狀態
export const useMarketplaceDiscoverPage = () => {
  const { t } = useI18n();
  const feedbackStore = useFeedbackStore();
  const preferenceStore = usePreferenceStore();
  const activeSlideIndexByCategory = ref<Record<string, number>>({});
  const loading = ref(false);
  const sourceListings = ref<SecondhandListingSummaryResponse[]>([]);

  const categoryRows = computed<DiscoverCategoryRow[]>(() =>
    marketplaceCategories
      .map((category) => {
        const source = sourceListings.value.filter((listing) => listing.category_code === category.value);
        const listings = source.map((listing) => buildListingPreview(listing, preferenceStore.locale));

        return {
          key: category.value,
          label: getMarketplaceCategoryLabel(category.value, preferenceStore.locale),
          count: source.length,
          filterTo: `/marketplace/filter?category_code=${category.value}`,
          heroSlides: listings.slice(0, 5),
          listings: listings.slice(0, 15),
        };
      })
      .filter((row) => row.listings.length > 0),
  );

  const recommendedItems = computed(() =>
    sourceListings.value.slice(0, 4).map((listing) => buildListingPreview(listing, preferenceStore.locale)),
  );

  const primaryRecommendation = computed(() => recommendedItems.value[0]);

  // 2.1 讀取 Discover 真實資料
  const loadDiscoverListings = async (): Promise<void> => {
    loading.value = true;

    try {
      const { data } = await fetchSecondhandListings({
        page: 1,
        page_size: 60,
        sort_by: 'latest',
      });
      sourceListings.value = data.data.items;
    } catch (error) {
      feedbackStore.pushToast(
        axios.isAxiosError(error)
          ? error.response?.data?.message ?? t('marketplace.discover.loadError')
          : t('marketplace.discover.loadError'),
        'error',
      );
      sourceListings.value = [];
    } finally {
      loading.value = false;
    }
  };

  // 2.2 取得分類目前輪播圖
  const getActiveCategorySlide = (row: DiscoverCategoryRow): DiscoverListingPreview =>
    row.heroSlides[activeSlideIndexByCategory.value[row.key] ?? 0] ?? row.heroSlides[0];

  // 2.3 切換分類輪播圖
  const updateCategorySlide = (row: DiscoverCategoryRow, offset: number): void => {
    const currentIndex = activeSlideIndexByCategory.value[row.key] ?? 0;
    const nextIndex = (currentIndex + offset + row.heroSlides.length) % row.heroSlides.length;

    activeSlideIndexByCategory.value = {
      ...activeSlideIndexByCategory.value,
      [row.key]: nextIndex,
    };
  };

  // 2.4 切到上一張分類輪播圖
  const showPreviousCategorySlide = (row: DiscoverCategoryRow): void => {
    updateCategorySlide(row, -1);
  };

  // 2.5 切到下一張分類輪播圖
  const showNextCategorySlide = (row: DiscoverCategoryRow): void => {
    updateCategorySlide(row, 1);
  };

  onMounted(() => {
    void loadDiscoverListings();
  });

  return {
    categoryRows,
    getActiveCategorySlide,
    loading,
    primaryRecommendation,
    recommendedItems,
    showNextCategorySlide,
    showPreviousCategorySlide,
  };
};

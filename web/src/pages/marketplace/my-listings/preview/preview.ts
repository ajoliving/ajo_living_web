/*
 * 我的帖子預覽頁 - 狀態與資料。
 * 1. 讀取真實帖子詳情。
 * 2. 格式化預覽頁圖片、價格與狀態資料。
 */
import axios from 'axios';
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute } from 'vue-router';

import { fetchSecondhandListingDetail } from '@/httpapis/secondhand-listings';
import {
  getMarketplaceCategoryLabel,
  getMarketplaceDistrictLabel,
} from '@/constants/marketplace';
import type { SecondhandListingDetailResponse } from '@/model/marketplace';
import { useFeedbackStore } from '@/stores/feedback';
import { usePreferenceStore } from '@/stores/preferences';
import { formatDate, formatPrice } from '@/utils/format';
import { resolveListingCoverImage, resolveListingStatus } from '@/utils/marketplace';

// 1. 管理我的帖子預覽頁資料
export const useMarketplaceMyListingPreviewPage = () => {
  const route = useRoute();
  const { t } = useI18n();
  const feedbackStore = useFeedbackStore();
  const preferenceStore = usePreferenceStore();
  const loading = ref(false);
  const listing = ref<SecondhandListingDetailResponse | null>(null);
  const listingId = computed(() => String(route.params.listingId ?? ''));
  const coverImage = computed(() => listing.value ? resolveListingCoverImage(listing.value) : undefined);
  const listingPrice = computed(() => {
    if (!listing.value) {
      return '';
    }

    return listing.value.price_mode === 'free'
      ? t('marketplace.editor.freePrice')
      : formatPrice(listing.value.price_hkd ?? 0, preferenceStore.locale);
  });
  const categoryLabel = computed(() =>
    listing.value ? getMarketplaceCategoryLabel(listing.value.category_code, preferenceStore.locale) : '',
  );
  const districtLabel = computed(() =>
    listing.value ? getMarketplaceDistrictLabel(listing.value.district_code, preferenceStore.locale) : '',
  );
  const statusLabel = computed(() => listing.value ? t(`common.state.${resolveListingStatus(listing.value)}`) : '');
  const publishedAt = computed(() =>
    listing.value ? formatDate(listing.value.published_at || listing.value.updated_at, preferenceStore.locale) : '',
  );

  // 1.1 讀取帖子詳情
  const loadPreview = async (): Promise<void> => {
    if (!listingId.value) {
      return;
    }

    loading.value = true;

    try {
      const { data } = await fetchSecondhandListingDetail(listingId.value);
      listing.value = data.data;
    } catch (error) {
      feedbackStore.pushToast(
        axios.isAxiosError(error)
          ? error.response?.data?.message ?? t('marketplace.detail.loadError')
          : t('marketplace.detail.loadError'),
        'error',
      );
    } finally {
      loading.value = false;
    }
  };

  onMounted(() => {
    void loadPreview();
  });

  return {
    categoryLabel,
    coverImage,
    districtLabel,
    listing,
    listingId,
    listingPrice,
    loading,
    publishedAt,
    statusLabel,
    t,
  };
};

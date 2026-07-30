/*
 * 我的帖子預覽頁 - 狀態與資料。
 * 1. 讀取真實帖子詳情。
 * 2. 格式化我的帖子詳情頁圖片、價格與狀態資料。
 * 3. 提供發布者編輯、查看公開頁與狀態操作。
 */
import axios from 'axios';
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import {
  deactivateSecondhandListing,
  fetchSecondhandListingDetail,
  markSecondhandListingSold,
  publishSecondhandListing,
  republishSecondhandListing,
  renewSecondhandListing,
} from '@/httpapis/secondhand-listings';
import {
  getMarketplaceCategoryLabel,
  getMarketplaceDistrictLabel,
} from '@/constants/marketplace';
import type { SecondhandListingDetailResponse } from '@/model/marketplace';
import { useFeedbackStore } from '@/stores/feedback';
import { usePreferenceStore } from '@/stores/preferences';
import { useSessionStore } from '@/stores/session';
import { formatDate, formatPrice } from '@/utils/format';
import { resolveListingCoverImage, resolveListingStatus } from '@/utils/marketplace';
import { formatAjoPoints, resolveWalletChargeCost, resolveWalletRenewChargeCost } from '@/utils/wallet';

type MyListingDetailAction = 'publish' | 'republish' | 'renew' | 'mark-sold' | 'deactivate';

// 1. 管理我的帖子預覽頁資料
export const useMarketplaceMyListingPreviewPage = () => {
  const route = useRoute();
  const router = useRouter();
  const { t } = useI18n();
  const feedbackStore = useFeedbackStore();
  const preferenceStore = usePreferenceStore();
  const sessionStore = useSessionStore();
  const loading = ref(false);
  const actionLoading = ref(false);
  const listing = ref<SecondhandListingDetailResponse | null>(null);
  const pendingAction = ref<MyListingDetailAction | null>(null);
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
  const canPublish = computed(() => listing.value?.publication_status === 'draft');
  const canRepublish = computed(() => listing.value?.publication_status === 'expired');
  const canRenew = computed(() => {
    const currentListing = listing.value;
    return currentListing?.publication_status === 'active' && currentListing.business_status !== 'sold';
  });
  const canMarkSold = computed(() => {
    const currentListing = listing.value;
    return currentListing?.publication_status === 'active' && currentListing.business_status !== 'sold';
  });
  const canDeactivate = computed(() => {
    const currentListing = listing.value;
    return currentListing?.publication_status === 'active' && currentListing.business_status !== 'sold';
  });
  const formatPoints = (value: number): string =>
    formatAjoPoints(value, t('common.brand.pointsName'), preferenceStore.locale);
  const secondhandChargeCost = resolveWalletChargeCost('secondhand');
  const secondhandRenewChargeCost = resolveWalletRenewChargeCost('secondhand');

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

  // 1.2 導向帖子編輯頁
  const openEditor = async (): Promise<void> => {
    if (!listingId.value) {
      return;
    }

    await router.push(`/account/listings/editor/${listingId.value}`);
  };

  // 1.3 導向公開帖子詳情
  const openPublicDetail = async (): Promise<void> => {
    if (!listingId.value) {
      return;
    }

    await router.push(`/marketplace/listing/${listingId.value}`);
  };

  const confirmDescription = computed(() => {
    const action = pendingAction.value;
    if (!action) {
      return '';
    }

    if (action === 'publish' || action === 'republish' || action === 'renew') {
      return `${t(resolveChargeConfirmKey(action))} ${formatPoints(resolveActionChargeCost(action))}`;
    }

    return t(action === 'mark-sold' ? 'marketplace.mine.confirmSold' : 'marketplace.mine.confirmDeactivate');
  });

  // 1.4 開啟帖子狀態操作確認
  const requestAction = (action: MyListingDetailAction): void => {
    if (listingId.value && !actionLoading.value) {
      pendingAction.value = action;
    }
  };

  // 1.5 執行已確認的帖子狀態操作
  const confirmAction = async (): Promise<void> => {
    const action = pendingAction.value;
    if (!listingId.value || !action || actionLoading.value) {
      return;
    }

    pendingAction.value = null;
    actionLoading.value = true;

    try {
      if (action === 'publish') {
        await publishSecondhandListing(listingId.value);
      }

      if (action === 'republish') {
        await republishSecondhandListing(listingId.value);
      }

      if (action === 'renew') {
        await renewSecondhandListing(listingId.value);
      }

      if (action === 'mark-sold') {
        await markSecondhandListingSold(listingId.value);
      }

      if (action === 'deactivate') {
        await deactivateSecondhandListing(listingId.value);
      }

      feedbackStore.pushToast(t('marketplace.mine.statusUpdated'), 'success');
      if (action === 'publish' || action === 'republish' || action === 'renew') {
        await sessionStore.loadCurrentUser();
      }
      await loadPreview();
    } catch (error) {
      feedbackStore.pushToast(
        axios.isAxiosError(error)
          ? error.response?.data?.message ?? t('marketplace.mine.updateError')
          : t('marketplace.mine.updateError'),
        'error',
      );
    } finally {
      actionLoading.value = false;
    }
  };

  // 1.6 關閉帖子狀態操作確認
  const cancelAction = (): void => {
    pendingAction.value = null;
  };

  // 1.7 輸出扣費確認文案 key
  const resolveChargeConfirmKey = (action: MyListingDetailAction): string => {
    if (action === 'publish') {
      return 'marketplace.mine.confirmPublishCharge';
    }
    if (action === 'republish') {
      return 'marketplace.mine.confirmRepublishCharge';
    }
    return 'marketplace.mine.confirmRenewCharge';
  };

  // 1.8 輸出指定動作扣費
  const resolveActionChargeCost = (action: MyListingDetailAction): number =>
    action === 'renew' ? secondhandRenewChargeCost : secondhandChargeCost;

  onMounted(() => {
    void loadPreview();
  });

  return {
    actionLoading,
    cancelAction,
    canDeactivate,
    canMarkSold,
    canPublish,
    canRepublish,
    canRenew,
    categoryLabel,
    confirmAction,
    confirmDescription,
    coverImage,
    districtLabel,
    listing,
    listingId,
    listingPrice,
    loading,
    openEditor,
    openPublicDetail,
    publishedAt,
    pendingAction,
    requestAction,
    secondhandChargeCost,
    secondhandRenewChargeCost,
    formatAjoPoints: formatPoints,
    statusLabel,
    t,
  };
};

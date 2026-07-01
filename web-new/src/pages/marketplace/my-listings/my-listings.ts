/*
 * 我的帖子 - 狀態與資料流程。
 * 1. 讀取我的帖子列表並依狀態分頁。
 * 2. 管理搜尋、狀態篩選與帖子狀態操作。
 */
import axios from 'axios';
import { computed, onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

import {
  deactivateSecondhandListing,
  fetchMySecondhandListings,
  markSecondhandListingSold,
  publishSecondhandListing,
  republishSecondhandListing,
  renewSecondhandListing,
} from '@/httpapis/secondhand-listings';
import type { SecondhandListingSummaryResponse } from '@/model/marketplace';
import { useFeedbackStore } from '@/stores/feedback';
import { usePreferenceStore } from '@/stores/preferences';
import { useSessionStore } from '@/stores/session';
import { formatDate, formatPrice } from '@/utils/format';
import {
  resolveListingCategoryLabel,
  resolveListingCommunityName,
  resolveListingCoverImage,
  resolveListingPrice,
  resolveListingPublishedAt,
  resolveListingStatus,
  resolveListingSummary,
  resolveListingTitle,
  resolveListingVisibility,
} from '@/utils/marketplace';
import { formatAjoPoints, resolveWalletChargeCost, resolveWalletRenewChargeCost } from '@/utils/wallet';

type MyListingsTab = 'all' | 'draft' | 'active' | 'hidden' | 'expired' | 'sold';
type MyListingsAction = 'publish' | 'republish' | 'renew' | 'mark-sold' | 'deactivate';

// 1. 管理我的帖子頁資料與動作
export const useMarketplaceMyListingsPage = () => {
  const { t } = useI18n();
  const router = useRouter();
  const feedbackStore = useFeedbackStore();
  const preferenceStore = usePreferenceStore();
  const sessionStore = useSessionStore();
  const loading = ref(false);
  const searchQuery = ref('');
  const activeTab = ref<MyListingsTab>('all');
  const items = ref<SecondhandListingSummaryResponse[]>([]);
  const formatPoints = (value: number): string =>
    formatAjoPoints(value, t('common.brand.pointsName'), preferenceStore.locale);

  const tabOptions = computed(() => [
    { label: t('marketplace.mine.totalListings'), value: 'all' },
    { label: t('marketplace.mine.draft'), value: 'draft' },
    { label: t('common.state.active'), value: 'active' },
    { label: t('marketplace.mine.hidden'), value: 'hidden' },
    { label: t('common.state.expired'), value: 'expired' },
    { label: t('common.state.sold'), value: 'sold' },
  ]);

  // 1.1 讀取我的帖子
  const loadMyListings = async (): Promise<void> => {
    loading.value = true;

    try {
      const { data } = await fetchMySecondhandListings({
        status: activeTab.value === 'all' ? '' : activeTab.value,
        page: 1,
        page_size: 50,
      });
      items.value = data.data.items;
    } catch (error) {
      feedbackStore.pushToast(
        axios.isAxiosError(error)
          ? error.response?.data?.message ?? t('marketplace.mine.loadError')
          : t('marketplace.mine.loadError'),
        'error',
      );
    } finally {
      loading.value = false;
    }
  };

  const filteredItems = computed(() =>
    items.value.filter((listing) => {
      const keyword = searchQuery.value.trim().toLowerCase();
      if (!keyword) {
        return true;
      }

      return [
        resolveListingTitle(listing),
        resolveListingSummary(listing),
        resolveListingCategoryLabel(listing),
        resolveListingCommunityName(listing),
      ]
        .join(' ')
        .toLowerCase()
        .includes(keyword);
    }),
  );

  // 1.2 切換帖子狀態分頁
  const setActiveTab = (value: string | number): void => {
    activeTab.value = value as MyListingsTab;
  };

  // 1.3 判斷帖子是否有可執行操作
  const hasListingActions = (listing: SecondhandListingSummaryResponse): boolean =>
    listing.publication_status === 'draft' ||
    listing.publication_status === 'expired' ||
    (listing.publication_status === 'active' && listing.business_status !== 'sold');

  // 1.4 執行帖子狀態操作
  const runAction = async (action: MyListingsAction, listingId: string): Promise<void> => {
    if ((action === 'publish' || action === 'republish' || action === 'renew') &&
      !window.confirm(`${t(resolveChargeConfirmKey(action))} ${formatPoints(resolveActionChargeCost(action))}`)) {
      return;
    }
    if (
      (action === 'mark-sold' || action === 'deactivate') &&
      !window.confirm(t(action === 'mark-sold' ? 'marketplace.mine.confirmSold' : 'marketplace.mine.confirmDeactivate'))
    ) {
      return;
    }

    try {
      if (action === 'publish') {
        await publishSecondhandListing(listingId);
      }

      if (action === 'republish') {
        await republishSecondhandListing(listingId);
      }

      if (action === 'renew') {
        await renewSecondhandListing(listingId);
      }

      if (action === 'mark-sold') {
        await markSecondhandListingSold(listingId);
      }

      if (action === 'deactivate') {
        await deactivateSecondhandListing(listingId);
      }

      feedbackStore.pushToast(t('marketplace.mine.statusUpdated'), 'success');
      if (action === 'publish' || action === 'republish' || action === 'renew') {
        await sessionStore.loadCurrentUser();
      }
      await loadMyListings();
    } catch (error) {
      feedbackStore.pushToast(
        axios.isAxiosError(error)
          ? error.response?.data?.message ?? t('marketplace.mine.updateError')
          : t('marketplace.mine.updateError'),
        'error',
      );
    }
  };

  // 1.5 導向新增帖子頁
  const openCreate = async (): Promise<void> => {
    await router.push('/account/listings/new');
  };

  // 1.6 導向帖子編輯頁
  const openEditor = async (listingId: string): Promise<void> => {
    await router.push(`/account/listings/editor/${listingId}`);
  };

  // 1.7 導向我的帖子管理詳情頁
  const openManagedDetail = async (listingId: string): Promise<void> => {
    await router.push(`/account/listings/item/${listingId}`);
  };

  // 1.8 導向公開詳情頁
  const openPublicDetail = async (listingId: string): Promise<void> => {
    await router.push(`/furniture/${listingId}`);
  };

  // 1.9 輸出扣費確認文案 key
  const resolveChargeConfirmKey = (action: MyListingsAction): string => {
    if (action === 'publish') {
      return 'marketplace.mine.confirmPublishCharge';
    }
    if (action === 'republish') {
      return 'marketplace.mine.confirmRepublishCharge';
    }
    return 'marketplace.mine.confirmRenewCharge';
  };

  // 1.10 輸出指定動作扣費
  const resolveActionChargeCost = (action: MyListingsAction): number =>
    action === 'renew'
      ? resolveWalletRenewChargeCost('secondhand')
      : resolveWalletChargeCost('secondhand');

  watch(activeTab, () => {
    void loadMyListings();
  });

  onMounted(() => {
    void loadMyListings();
  });

  return {
    activeTab,
    filteredItems,
    formatDate,
    formatPrice,
    hasListingActions,
    loading,
    openCreate,
    openEditor,
    openManagedDetail,
    openPublicDetail,
    preferenceStore,
    resolveListingCategoryLabel,
    resolveListingCommunityName,
    resolveListingCoverImage,
    resolveListingPrice,
    resolveListingPublishedAt,
    resolveListingStatus,
    resolveListingSummary,
    resolveListingTitle,
    resolveListingVisibility,
    runAction,
    searchQuery,
    formatAjoPoints: formatPoints,
    secondhandChargeCost: resolveWalletChargeCost('secondhand'),
    secondhandRenewChargeCost: resolveWalletRenewChargeCost('secondhand'),
    setActiveTab,
    t,
    tabOptions,
  };
};

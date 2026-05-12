/*
 * 二手帖子詳情頁 - 狀態與資料。
 * 1. 讀取真實帖子詳情並格式化展示欄位。
 * 2. 處理聯絡方式授權與建立聊天會話。
 */
import axios from 'axios';
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import { createOrReuseChat } from '@/httpapis/chats';
import {
  favoriteSecondhandListing,
  fetchListingContactAccess,
  fetchSecondhandListingDetail,
  unfavoriteSecondhandListing,
} from '@/httpapis/secondhand-listings';
import {
  getMarketplaceCategoryLabel,
  getMarketplaceConditionLabel,
  getMarketplaceDistrictLabel,
} from '@/constants/marketplace';
import type { ContactAccessResult, SecondhandListingDetailResponse } from '@/model/marketplace';
import { useFeedbackStore } from '@/stores/feedback';
import { useSessionStore } from '@/stores/session';
import { usePreferenceStore } from '@/stores/preferences';
import { formatDate, formatPrice } from '@/utils/format';
import {
  getDeliveryTagTranslationKey,
  resolveListingImages,
  resolveListingOwnerName,
} from '@/utils/marketplace';

// 1. 管理帖子詳情頁資料與動作
export const useMarketplaceListingPage = () => {
  const route = useRoute();
  const router = useRouter();
  const { t } = useI18n();
  const feedbackStore = useFeedbackStore();
  const preferenceStore = usePreferenceStore();
  const sessionStore = useSessionStore();
  const listing = ref<SecondhandListingDetailResponse | null>(null);
  const contactAccess = ref<ContactAccessResult | null>(null);
  const loading = ref(false);
  const loadingContact = ref(false);
  const openingChat = ref(false);
  const updatingFavorite = ref(false);

  const listingId = computed(() => String(route.params.listingId ?? ''));
  const galleryImages = computed(() => listing.value ? resolveListingImages(listing.value) : []);
  const coverImage = computed(() => galleryImages.value[0]);
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
  const conditionLabel = computed(() =>
    listing.value ? getMarketplaceConditionLabel(listing.value.condition_level, preferenceStore.locale) : '',
  );
  const districtLabel = computed(() =>
    listing.value ? getMarketplaceDistrictLabel(listing.value.district_code, preferenceStore.locale) : '',
  );
  const ownerName = computed(() => listing.value ? resolveListingOwnerName(listing.value) : '');
  const ownerAvatarUrl = computed(() => listing.value?.owner?.avatar_url ?? '');
  const publishedAt = computed(() =>
    listing.value ? formatDate(listing.value.published_at || listing.value.updated_at, preferenceStore.locale) : '',
  );
  const contactPayload = computed(() => contactAccess.value?.contact_payload ?? {});
  const isFavorited = computed(() => Boolean(listing.value?.is_favorited));
  const formatDeliveryTag = (tag: string): string => {
    const translationKey = getDeliveryTagTranslationKey(tag);
    return translationKey ? t(translationKey) : tag;
  };

  // 1.1 讀取帖子詳情
  const loadListing = async (): Promise<void> => {
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

  // 1.2 取得可展示聯絡方式
  const revealContact = async (): Promise<void> => {
    if (!listing.value) {
      return;
    }

    loadingContact.value = true;

    try {
      const { data } = await fetchListingContactAccess(listing.value.listing_id);
      contactAccess.value = data.data;
    } catch (error) {
      feedbackStore.pushToast(
        axios.isAxiosError(error)
          ? error.response?.data?.message ?? t('marketplace.detail.contactError')
          : t('marketplace.detail.contactError'),
        'error',
      );
    } finally {
      loadingContact.value = false;
    }
  };

  // 1.3 切換收藏狀態
  const toggleFavorite = async (): Promise<void> => {
    if (!listing.value || updatingFavorite.value) {
      return;
    }
    if (!sessionStore.isAuthenticated) {
      await router.push({
        path: '/login',
        query: { redirect: route.fullPath },
      });
      return;
    }

    updatingFavorite.value = true;

    try {
      const { data } = isFavorited.value
        ? await unfavoriteSecondhandListing(listing.value.listing_id)
        : await favoriteSecondhandListing(listing.value.listing_id);
      listing.value = {
        ...listing.value,
        is_favorited: data.data.is_favorited,
      };
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
      updatingFavorite.value = false;
    }
  };

  // 1.4 建立或重用聊天
  const openChat = async (): Promise<void> => {
    if (!listing.value) {
      return;
    }

    openingChat.value = true;

    try {
      const { data } = await createOrReuseChat(listing.value.listing_id);
      await router.push(`/account/marketplace/my/chat/${data.data.chat_id}`);
    } catch (error) {
      feedbackStore.pushToast(
        axios.isAxiosError(error)
          ? error.response?.data?.message ?? t('marketplace.detail.chatError')
          : t('marketplace.detail.chatError'),
        'error',
      );
    } finally {
      openingChat.value = false;
    }
  };

  onMounted(() => {
    void loadListing();
  });

  return {
    categoryLabel,
    conditionLabel,
    contactPayload,
    coverImage,
    districtLabel,
    galleryImages,
    formatDeliveryTag,
    listing,
    listingId,
    listingPrice,
    loading,
    loadingContact,
    openChat,
    openingChat,
    ownerAvatarUrl,
    ownerName,
    publishedAt,
    revealContact,
    isFavorited,
    t,
    toggleFavorite,
    updatingFavorite,
  };
};

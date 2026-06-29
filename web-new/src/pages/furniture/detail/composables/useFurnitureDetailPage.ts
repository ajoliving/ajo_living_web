/*
 * 家具詳情 - 狀態與資料。
 * 1. 讀取真實二手帖子詳情並輸出家具詳情展示欄位。
 * 2. 管理圖片選取、收藏、聯絡方式授權與站內聊天入口。
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
import { usePreferenceStore } from '@/stores/preferences';
import { useSessionStore } from '@/stores/session';
import { formatDate, formatPrice } from '@/utils/format';
import {
  getDeliveryTagTranslationKey,
  resolveListingCommunityName,
  resolveListingImages,
  resolveListingOwnerName,
} from '@/utils/marketplace';

interface ContactDisplayRow {
  key: string;
  label: string;
  value: string;
  href: string;
  isWhatsApp: boolean;
}

// 1. 建立電話撥號連結
const buildPhoneHref = (phone: string): string => {
  const normalizedPhone = phone.replace(/\s+/g, '');

  return normalizedPhone ? `tel:${normalizedPhone}` : '';
};

// 2. 建立 WhatsApp 跳轉連結
const buildWhatsAppHref = (value: string): string => {
  if (value.startsWith('https://wa.me/')) {
    return value;
  }

  const digits = value.replace(/[+\s\-()]/g, '');

  return digits ? `https://wa.me/${digits}` : '';
};

// 3. 判斷聯絡方式類型
const isWhatsAppContact = (key: string): boolean =>
  key === 'whatsapp_url' || key === 'whatsapp';

// 4. 管理家具詳情資料與動作
export const useFurnitureDetailPage = () => {
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
  const selectedImageIndex = ref(0);

  const listingId = computed(() => String(route.params.listingId ?? ''));
  const galleryImages = computed(() => listing.value ? resolveListingImages(listing.value) : []);
  const coverImage = computed(() => galleryImages.value[selectedImageIndex.value] ?? galleryImages.value[0]);
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
  const communityName = computed(() =>
    listing.value ? resolveListingCommunityName(listing.value) : '',
  );
  const ownerName = computed(() => listing.value ? resolveListingOwnerName(listing.value) : '');
  const publishedAt = computed(() =>
    listing.value ? formatDate(listing.value.published_at || listing.value.updated_at, preferenceStore.locale) : '',
  );
  const isFavorited = computed(() => Boolean(listing.value?.is_favorited));
  const contactPayload = computed(() => contactAccess.value?.contact_payload ?? {});
  const contactRows = computed<ContactDisplayRow[]>(() =>
    Object.entries(contactPayload.value)
      .filter(([, value]) => Boolean(value))
      .map(([key, value]) => ({
        key,
        label: resolveContactLabel(key),
        value,
        href: isWhatsAppContact(key) ? buildWhatsAppHref(value) : buildPhoneHref(value),
        isWhatsApp: isWhatsAppContact(key),
      })),
  );
  const revealContactLabel = computed(() =>
    contactRows.value.length > 0
      ? t('marketplace.detail.contactUnlockedAction')
      : t('common.action.revealContact'),
  );

  const formatDeliveryTag = (tag: string): string => {
    const translationKey = getDeliveryTagTranslationKey(tag);

    return translationKey ? t(translationKey) : tag;
  };

  // 4.1 輸出聯絡方式欄位標籤
  const resolveContactLabel = (key: string): string => {
    if (key === 'phone') {
      return t('marketplace.detail.contactPhone');
    }
    if (key === 'whatsapp_url' || key === 'whatsapp') {
      return t('marketplace.detail.contactWhatsApp');
    }

    return key;
  };

  // 4.2 讀取帖子詳情
  const loadListing = async (): Promise<void> => {
    if (!listingId.value) {
      return;
    }

    loading.value = true;

    try {
      const { data } = await fetchSecondhandListingDetail(listingId.value);
      listing.value = data.data;
      selectedImageIndex.value = 0;
    } catch (error) {
      feedbackStore.pushToast(
        axios.isAxiosError<{ message?: string }>(error)
          ? error.response?.data?.message ?? t('marketplace.detail.loadError')
          : t('marketplace.detail.loadError'),
        'error',
      );
    } finally {
      loading.value = false;
    }
  };

  // 4.3 取得可展示聯絡方式
  const revealContact = async (): Promise<void> => {
    if (!listing.value || loadingContact.value || contactRows.value.length > 0) {
      return;
    }
    if (!sessionStore.isAuthenticated) {
      await router.push({
        path: '/login',
        query: { redirect: route.fullPath },
      });
      return;
    }

    loadingContact.value = true;

    try {
      const { data } = await fetchListingContactAccess(listing.value.listing_id);
      contactAccess.value = data.data;
      feedbackStore.pushToast(t('marketplace.detail.contactUnlockedSuccess'), 'success');
    } catch (error) {
      feedbackStore.pushToast(
        axios.isAxiosError<{ message?: string }>(error)
          ? error.response?.data?.message ?? t('marketplace.detail.contactError')
          : t('marketplace.detail.contactError'),
        'error',
      );
    } finally {
      loadingContact.value = false;
    }
  };

  // 4.4 切換收藏狀態
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
        axios.isAxiosError<{ message?: string }>(error)
          ? error.response?.data?.message ?? t('marketplace.detail.favoriteError')
          : t('marketplace.detail.favoriteError'),
        'error',
      );
    } finally {
      updatingFavorite.value = false;
    }
  };

  // 4.5 建立或重用聊天
  const openChat = async (): Promise<void> => {
    if (!listing.value || openingChat.value) {
      return;
    }
    if (!sessionStore.isAuthenticated) {
      await router.push({
        path: '/login',
        query: { redirect: route.fullPath },
      });
      return;
    }

    openingChat.value = true;

    try {
      const { data } = await createOrReuseChat(listing.value.listing_id);
      await router.push(`/account/chat/${data.data.chat_id}`);
    } catch (error) {
      feedbackStore.pushToast(
        axios.isAxiosError<{ message?: string }>(error)
          ? error.response?.data?.message ?? t('marketplace.detail.chatError')
          : t('marketplace.detail.chatError'),
        'error',
      );
    } finally {
      openingChat.value = false;
    }
  };

  // 4.6 切換主圖
  const selectImage = (index: number): void => {
    selectedImageIndex.value = index;
  };

  onMounted(() => {
    void loadListing();
  });

  return {
    categoryLabel,
    communityName,
    conditionLabel,
    contactRows,
    coverImage,
    districtLabel,
    formatDeliveryTag,
    galleryImages,
    isFavorited,
    listing,
    listingId,
    listingPrice,
    loading,
    loadingContact,
    openChat,
    openingChat,
    ownerName,
    publishedAt,
    revealContact,
    revealContactLabel,
    selectImage,
    selectedImageIndex,
    t,
    toggleFavorite,
    updatingFavorite,
  };
};

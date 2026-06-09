/*
 * 二手帖子詳情頁 - 狀態與資料。
 * 1. 讀取真實帖子詳情並格式化展示欄位。
 * 2. 處理聯絡方式授權與建立聊天會話。
 */
import axios from 'axios';
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import { createOrReuseChat } from '@/domains/communications/api';
import {
  favoriteSecondhandListing,
  fetchListingContactAccess,
  fetchSecondhandListingDetail,
  unfavoriteSecondhandListing,
} from '@/domains/marketplace/api';
import {
  getMarketplaceCategoryLabel,
  getMarketplaceConditionLabel,
  getMarketplaceDistrictLabel,
} from '@/domains/marketplace/constants';
import type { ContactAccessResult, SecondhandListingDetailResponse } from '@/domains/marketplace/model';
import { useFeedbackStore } from '@/app/stores/feedback';
import { useSessionStore } from '@/app/stores/session';
import { usePreferenceStore } from '@/app/stores/preferences';
import { formatDate, formatPrice } from '@/shared/utils/format';
import {
  getDeliveryTagTranslationKey,
  resolveListingImages,
  resolveListingOwnerName,
} from '@/shared/utils/marketplace';
import { formatAjoPoints, resolveWalletContactAccessCost } from '@/shared/utils/wallet';

interface ContactDisplayRow {
  key: string;
  label: string;
  value: string;
  href: string;
  icon: 'phone' | 'message';
  displayText: string;
  isIconOnly: boolean;
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

// 3. 判斷聯絡方式是否應以 WhatsApp 按鈕展示
const isWhatsAppContact = (key: string): boolean =>
  key === 'whatsapp_url' || key === 'whatsapp';

// 1. 管理帖子詳情頁資料與動作
export const useFurnitureListingPage = () => {
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
  const contactRows = computed<ContactDisplayRow[]>(() =>
    Object.entries(contactPayload.value).map(([key, value]) => ({
      key,
      label: resolveContactLabel(key),
      value,
      href: isWhatsAppContact(key) ? buildWhatsAppHref(value) : buildPhoneHref(value),
      icon: isWhatsAppContact(key) ? 'message' : 'phone',
      displayText: isWhatsAppContact(key) ? 'WhatsApp' : '',
      isIconOnly: !isWhatsAppContact(key),
    })),
  );
  const contactAccessCost = computed(() =>
    formatAjoPoints(resolveWalletContactAccessCost('secondhand'), t('common.brand.pointsName'), preferenceStore.locale),
  );
  const isOwnListing = computed(() =>
    Boolean(listing.value?.owner?.public_id && sessionStore.me?.public_id && listing.value.owner.public_id === sessionStore.me.public_id),
  );
  const hasPaidContactChannel = computed(() =>
    Boolean(listing.value?.contact_summary.show_phone || listing.value?.contact_summary.show_whatsapp),
  );
  const revealContactLabel = computed(() =>
    contactRows.value.length > 0
      ? t('marketplace.detail.contactUnlockedAction')
      : isOwnListing.value || !hasPaidContactChannel.value
        ? t('common.action.revealContact')
      : t('marketplace.detail.revealContactWithCost', { cost: contactAccessCost.value }),
  );
  const isFavorited = computed(() => Boolean(listing.value?.is_favorited));
  const formatDeliveryTag = (tag: string): string => {
    const translationKey = getDeliveryTagTranslationKey(tag);
    return translationKey ? t(translationKey) : tag;
  };

  // 1.1 輸出聯絡方式欄位標籤
  const resolveContactLabel = (key: string): string => {
    if (key === 'phone') {
      return t('marketplace.detail.contactPhone');
    }
    if (key === 'whatsapp_url' || key === 'whatsapp') {
      return t('marketplace.detail.contactWhatsApp');
    }
    return key;
  };

  // 1.2 讀取帖子詳情
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

  // 1.3 取得可展示聯絡方式
  const revealContact = async (): Promise<void> => {
    if (!listing.value) {
      return;
    }
    if (contactRows.value.length > 0) {
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
      if (typeof data.data.points_balance_after === 'number' && sessionStore.me) {
        sessionStore.me = {
          ...sessionStore.me,
          ajo_balance: data.data.points_balance_after,
        };
      }
      feedbackStore.pushToast(
        (data.data.points_charged ?? 0) > 0
          ? t('marketplace.detail.contactChargedSuccess', {
            cost: formatAjoPoints(data.data.points_charged ?? 0, t('common.brand.pointsName'), preferenceStore.locale),
          })
          : t('marketplace.detail.contactUnlockedSuccess'),
        'success',
      );
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

  // 1.4 切換收藏狀態
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

  // 1.5 建立或重用聊天
  const openChat = async (): Promise<void> => {
    if (!listing.value) {
      return;
    }

    openingChat.value = true;

    try {
      const { data } = await createOrReuseChat(listing.value.listing_id);
      await router.push({ path: '/member', query: { tab: 'chat', chat_id: data.data.chat_id } });
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
    contactRows,
    revealContactLabel,
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

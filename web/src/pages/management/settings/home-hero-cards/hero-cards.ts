/*
 * 首頁三大圖設定資料流程。
 * 1. 固定管理二手、樓盤放售與服務住宅三個首頁模組。
 * 2. 上傳 PNG 圖片並保存主標題、副標題與內容。
 */
import axios from 'axios';
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue';
import { useI18n } from 'vue-i18n';

import { fetchHomeModuleCardSettings, saveHomeModuleCardSettings } from '@/domains/home/content-api';
import { completeUpload, createUploadPresign } from '@/domains/media/uploads-api';
import type { HomeContentModuleCode, HomeModuleCard } from '@/domains/home/content-model';
import { useFeedbackStore } from '@/app/stores/feedback';
import { buildUploadHeaders } from '@/shared/utils/upload';

const homeCarouselObjectPrefix = 'ajo_living/eng/home-carousel/';

interface HomeModuleCardView {
  moduleCode: HomeContentModuleCode;
  mediaAssetId: string;
  url: string;
  objectKey: string;
  title: string;
  subtitle: string;
  body: string;
  objectUrl?: string;
  uploading: boolean;
}

interface HomeModuleOption {
  code: HomeContentModuleCode;
  labelKey: string;
}

const moduleOptions: HomeModuleOption[] = [
  { code: 'secondhand', labelKey: 'marketplace.settings.homeModuleSecondhand' },
  { code: 'property_sale', labelKey: 'marketplace.settings.homeModulePropertySale' },
  { code: 'serviced_apartment', labelKey: 'marketplace.settings.homeModuleServicedApartment' },
];

// 1. 建立三大圖設定流程
export const useHomeHeroCardsSettingsPage = () => {
  const { t } = useI18n();
  const feedbackStore = useFeedbackStore();
  const loading = ref(false);
  const saving = ref(false);
  const cards = reactive<Record<HomeContentModuleCode, HomeModuleCardView>>(createEmptyCards());

  const orderedCards = computed(() => moduleOptions.map((module) => cards[module.code]));
  const canSave = computed(
    () =>
      !loading.value &&
      !saving.value &&
      orderedCards.value.every((card) => !card.uploading && card.mediaAssetId.trim().length > 0),
  );

  // 1.1 讀取三大圖設定
  const loadCards = async (): Promise<void> => {
    loading.value = true;

    try {
      const { data } = await fetchHomeModuleCardSettings();
      applyCards(data.data.cards);
    } catch (error) {
      feedbackStore.pushToast(resolveErrorMessage(error, t('marketplace.settings.homeHeroCardsLoadError')), 'error');
    } finally {
      loading.value = false;
    }
  };

  // 1.2 上傳單個模組圖片
  const handleFileSelected = async (moduleCode: HomeContentModuleCode, event: Event): Promise<void> => {
    const input = event.target as HTMLInputElement;
    const file = input.files?.[0];
    input.value = '';
    if (!file) {
      return;
    }
    if (file.type !== 'image/png') {
      feedbackStore.pushToast(t('marketplace.settings.homePNGOnly'), 'error');
      return;
    }

    const card = cards[moduleCode];
    if (card.objectUrl) {
      URL.revokeObjectURL(card.objectUrl);
    }
    const objectUrl = URL.createObjectURL(file);
    card.url = objectUrl;
    card.objectUrl = objectUrl;
    card.uploading = true;

    try {
      const { data } = await createUploadPresign({
        file_name: file.name,
        mime_type: file.type,
        file_size: file.size,
        object_prefix: homeCarouselObjectPrefix,
      });
      const presign = data.data;
      const uploadResponse = await fetch(presign.upload_url, {
        method: 'PUT',
        headers: buildUploadHeaders(presign.headers, file.type),
        body: file,
      });
      if (!uploadResponse.ok) {
        throw new Error(`home module image upload failed with status ${uploadResponse.status}`);
      }
      const completeResponse = await completeUpload({
        object_key: presign.object_key,
        mime_type: file.type,
        file_size: file.size,
      });
      const media = completeResponse.data.data;
      if (card.objectUrl) {
        URL.revokeObjectURL(card.objectUrl);
      }
      card.mediaAssetId = media.media_asset_id;
      card.url = media.url;
      card.objectKey = media.object_key;
      card.objectUrl = undefined;
      card.uploading = false;
    } catch (error) {
      card.mediaAssetId = '';
      card.url = '';
      card.objectKey = '';
      card.uploading = false;
      if (card.objectUrl) {
        URL.revokeObjectURL(card.objectUrl);
        card.objectUrl = undefined;
      }
      feedbackStore.pushToast(resolveErrorMessage(error, t('marketplace.settings.homeHeroCardsUploadError')), 'error');
    }
  };

  // 1.3 保存三大圖設定
  const saveCards = async (): Promise<void> => {
    if (!canSave.value) {
      return;
    }

    saving.value = true;

    try {
      const { data } = await saveHomeModuleCardSettings(
        orderedCards.value.map((card) => ({
          module_code: card.moduleCode,
          media_asset_id: card.mediaAssetId,
          title: card.title.trim(),
          subtitle: card.subtitle.trim(),
          body: card.body.trim(),
        })),
      );
      applyCards(data.data.cards);
      feedbackStore.pushToast(t('marketplace.settings.homeHeroCardsSaveSuccess'), 'success');
    } catch (error) {
      feedbackStore.pushToast(resolveErrorMessage(error, t('marketplace.settings.homeHeroCardsSaveError')), 'error');
    } finally {
      saving.value = false;
    }
  };

  // 1.4 將 API 資料套回固定三模組
  const applyCards = (items: HomeModuleCard[]): void => {
    const itemMap = new Map(items.map((item) => [item.module_code, item]));
    moduleOptions.forEach((module) => {
      const item = itemMap.get(module.code);
      const card = cards[module.code];
      if (card.objectUrl) {
        URL.revokeObjectURL(card.objectUrl);
      }
      card.mediaAssetId = item?.media_asset_id ?? '';
      card.url = item?.url ?? '';
      card.objectKey = item?.object_key ?? '';
      card.title = item?.title ?? '';
      card.subtitle = item?.subtitle ?? '';
      card.body = item?.body ?? '';
      card.objectUrl = undefined;
      card.uploading = false;
    });
  };

  // 1.5 釋放本地圖片 URL
  const revokeObjectURLs = (): void => {
    orderedCards.value.forEach((card) => {
      if (card.objectUrl) {
        URL.revokeObjectURL(card.objectUrl);
      }
    });
  };

  onMounted(() => {
    void loadCards();
  });

  onBeforeUnmount(() => {
    revokeObjectURLs();
  });

  return {
    canSave,
    handleFileSelected,
    loading,
    moduleOptions,
    orderedCards,
    saveCards,
    saving,
    t,
  };
};

// 2. 建立三個空白模組設定
const createEmptyCards = (): Record<HomeContentModuleCode, HomeModuleCardView> =>
  moduleOptions.reduce((result, module) => {
    result[module.code] = {
      moduleCode: module.code,
      mediaAssetId: '',
      url: '',
      objectKey: '',
      title: '',
      subtitle: '',
      body: '',
      uploading: false,
    };
    return result;
  }, {} as Record<HomeContentModuleCode, HomeModuleCardView>);

// 3. 解析 API 錯誤訊息
const resolveErrorMessage = (error: unknown, fallback: string): string =>
  axios.isAxiosError(error) ? error.response?.data?.message ?? fallback : fallback;

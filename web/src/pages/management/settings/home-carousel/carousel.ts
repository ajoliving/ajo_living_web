/*
 * 首頁輪播設定資料流程。
 * 1. 讀取、上傳、排序與保存首頁輪播圖片。
 * 2. 沿用 OSS 兩段式上傳並限定首頁目錄。
 */
import axios from 'axios';
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';

import { fetchHomeCarouselSettings, saveHomeCarouselSettings } from '@/domains/home/content-api';
import { completeUpload, createUploadPresign } from '@/domains/media/uploads-api';
import type { HomeCarouselImage } from '@/domains/home/content-model';
import { useFeedbackStore } from '@/app/stores/feedback';
import { buildUploadHeaders } from '@/shared/utils/upload';

const homeCarouselObjectPrefix = 'ajo_living/eng/home-carousel/';

interface HomeCarouselImageView extends HomeCarouselImage {
  key: string;
  file?: File;
  objectUrl?: string;
  uploading: boolean;
}

// 1. 建立首頁輪播設定流程
export const useHomeCarouselSettingsPage = () => {
  const { t } = useI18n();
  const feedbackStore = useFeedbackStore();
  const loading = ref(false);
  const saving = ref(false);
  const images = ref<HomeCarouselImageView[]>([]);

  const canSave = computed(() => !loading.value && !saving.value && images.value.every((item) => !item.uploading));

  // 1.1 讀取輪播設定
  const loadCarousel = async (): Promise<void> => {
    loading.value = true;

    try {
      const { data } = await fetchHomeCarouselSettings();
      revokeObjectURLs();
      images.value = data.data.items.map((item) => ({
        ...item,
        key: item.media_asset_id,
        uploading: false,
      }));
    } catch (error) {
      feedbackStore.pushToast(resolveErrorMessage(error, t('marketplace.settings.homeCarouselLoadError')), 'error');
    } finally {
      loading.value = false;
    }
  };

  // 1.2 選擇圖片並立即上傳
  const handleFilesSelected = async (event: Event): Promise<void> => {
    const input = event.target as HTMLInputElement;
    const files = Array.from(input.files ?? []);
    input.value = '';

    for (const file of files) {
      await addAndUploadFile(file);
    }
  };

  // 1.3 新增並上傳單張圖片
  const addAndUploadFile = async (file: File): Promise<void> => {
    if (!file.type.startsWith('image/')) {
      feedbackStore.pushToast(t('marketplace.settings.imageOnly'), 'error');
      return;
    }

    const key = `${Date.now()}-${file.name}`;
    const objectUrl = URL.createObjectURL(file);
    images.value = [
      ...images.value,
      {
        key,
        media_asset_id: '',
        url: objectUrl,
        object_key: '',
        sort_order: images.value.length + 1,
        file,
        objectUrl,
        uploading: true,
      },
    ];

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
        throw new Error(`home carousel upload failed with status ${uploadResponse.status}`);
      }
      const completeResponse = await completeUpload({
        object_key: presign.object_key,
        mime_type: file.type,
        file_size: file.size,
      });
      const media = completeResponse.data.data;
      images.value = images.value.map((item) => {
        if (item.key !== key) {
          return item;
        }
        if (item.objectUrl) {
          URL.revokeObjectURL(item.objectUrl);
        }
        return {
          key: media.media_asset_id,
          media_asset_id: media.media_asset_id,
          url: media.url,
          object_key: media.object_key,
          sort_order: item.sort_order,
          uploading: false,
        };
      });
    } catch (error) {
      images.value = images.value.filter((item) => item.key !== key);
      URL.revokeObjectURL(objectUrl);
      feedbackStore.pushToast(resolveErrorMessage(error, t('marketplace.settings.homeCarouselUploadError')), 'error');
    }
  };

  // 1.4 移動圖片排序
  const moveImage = (key: string, direction: -1 | 1): void => {
    const index = images.value.findIndex((item) => item.key === key);
    const targetIndex = index + direction;
    if (index < 0 || targetIndex < 0 || targetIndex >= images.value.length) {
      return;
    }

    const nextImages = [...images.value];
    const [item] = nextImages.splice(index, 1);
    nextImages.splice(targetIndex, 0, item);
    images.value = normalizeSortOrders(nextImages);
  };

  // 1.5 移除圖片
  const removeImage = (key: string): void => {
    const target = images.value.find((item) => item.key === key);
    if (target?.objectUrl) {
      URL.revokeObjectURL(target.objectUrl);
    }
    images.value = normalizeSortOrders(images.value.filter((item) => item.key !== key));
  };

  // 1.6 保存輪播設定
  const saveCarousel = async (): Promise<void> => {
    if (!canSave.value) {
      return;
    }

    saving.value = true;

    try {
      const { data } = await saveHomeCarouselSettings(
        images.value
          .filter((item) => item.media_asset_id)
          .map((item, index) => ({
            media_asset_id: item.media_asset_id,
            sort_order: index + 1,
          })),
      );
      images.value = data.data.items.map((item) => ({
        ...item,
        key: item.media_asset_id,
        uploading: false,
      }));
      feedbackStore.pushToast(t('marketplace.settings.homeCarouselSaveSuccess'), 'success');
    } catch (error) {
      feedbackStore.pushToast(resolveErrorMessage(error, t('marketplace.settings.homeCarouselSaveError')), 'error');
    } finally {
      saving.value = false;
    }
  };

  // 1.7 釋放本地預覽 URL
  const revokeObjectURLs = (): void => {
    images.value.forEach((item) => {
      if (item.objectUrl) {
        URL.revokeObjectURL(item.objectUrl);
      }
    });
  };

  onMounted(() => {
    void loadCarousel();
  });

  onBeforeUnmount(() => {
    revokeObjectURLs();
  });

  return {
    canSave,
    handleFilesSelected,
    images,
    loadCarousel,
    loading,
    moveImage,
    removeImage,
    saveCarousel,
    saving,
    t,
  };
};

// 2. 重新整理排序
const normalizeSortOrders = (items: HomeCarouselImageView[]): HomeCarouselImageView[] =>
  items.map((item, index) => ({ ...item, sort_order: index + 1 }));

// 3. 解析 API 錯誤訊息
const resolveErrorMessage = (error: unknown, fallback: string): string =>
  axios.isAxiosError(error) ? error.response?.data?.message ?? fallback : fallback;

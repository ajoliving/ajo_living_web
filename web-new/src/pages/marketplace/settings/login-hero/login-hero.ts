/*
 * 登入背景圖設定資料流程。
 * 1. 讀取、上傳與保存登入頁左側背景圖片。
 * 2. 管理圖片來源與地點文字。
 */
import axios from 'axios';
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';

import { fetchLoginHeroSettings, saveLoginHeroSettings } from '@/httpapis/home-content';
import { completeUpload, createUploadPresign } from '@/httpapis/uploads';
import { useFeedbackStore } from '@/stores/feedback';
import { buildUploadHeaders } from '@/utils/upload';

const loginHeroObjectPrefix = 'ajo_living/login_bag/';
const loginHeroMaxImages = 3;

interface LoginHeroFormState {
  key: string;
  mediaAssetId: string;
  url: string;
  objectKey: string;
  author: string;
  location: string;
  objectUrl?: string;
  uploading: boolean;
}

const createEmptyHero = (index: number): LoginHeroFormState => ({
  key: `login-hero-${Date.now()}-${index}`,
  mediaAssetId: '',
  url: '',
  objectKey: '',
  author: '',
  location: '',
  uploading: false,
});

// 1. 建立登入背景圖設定流程
export const useLoginHeroSettingsPage = () => {
  const { t } = useI18n();
  const feedbackStore = useFeedbackStore();
  const loading = ref(false);
  const saving = ref(false);
  const images = ref<LoginHeroFormState[]>([createEmptyHero(1)]);

  const canSave = computed(
    () =>
      !loading.value &&
      !saving.value &&
      images.value.length > 0 &&
      images.value.every((item) => !item.uploading && item.mediaAssetId.trim().length > 0),
  );
  const canAddImage = computed(() => images.value.length < loginHeroMaxImages);

  // 1.1 讀取登入背景圖設定
  const loadHero = async (): Promise<void> => {
    loading.value = true;

    try {
      const { data } = await fetchLoginHeroSettings();
      revokeObjectURLs();
      const loadedImages = data.data.items
        .slice()
        .sort((left, right) => left.sort_order - right.sort_order)
        .slice(0, loginHeroMaxImages)
        .map((item, index) => ({
          key: item.media_asset_id || `login-hero-${index + 1}`,
          mediaAssetId: item.media_asset_id,
          url: item.url,
          objectKey: item.object_key,
          author: item.author,
          location: item.location,
          uploading: false,
        }));
      images.value = loadedImages.length > 0 ? loadedImages : [createEmptyHero(1)];
    } catch (error) {
      feedbackStore.pushToast(resolveErrorMessage(error, t('marketplace.settings.loginHeroLoadError')), 'error');
    } finally {
      loading.value = false;
    }
  };

  // 1.2 選擇並上傳登入背景圖
  const handleFileSelected = async (targetKey: string, event: Event): Promise<void> => {
    const input = event.target as HTMLInputElement;
    const file = input.files?.[0];
    input.value = '';
    if (!file) {
      return;
    }
    if (!file.type.startsWith('image/')) {
      feedbackStore.pushToast(t('marketplace.settings.imageOnly'), 'error');
      return;
    }

    const target = images.value.find((item) => item.key === targetKey);
    if (!target) {
      return;
    }
    if (target.objectUrl) {
      URL.revokeObjectURL(target.objectUrl);
    }
    const objectUrl = URL.createObjectURL(file);
    target.url = objectUrl;
    target.objectUrl = objectUrl;
    target.uploading = true;

    try {
      const { data } = await createUploadPresign({
        file_name: file.name,
        mime_type: file.type,
        file_size: file.size,
        object_prefix: loginHeroObjectPrefix,
      });
      const presign = data.data;
      const uploadResponse = await fetch(presign.upload_url, {
        method: 'PUT',
        headers: buildUploadHeaders(presign.headers, file.type),
        body: file,
      });
      if (!uploadResponse.ok) {
        throw new Error(`login hero upload failed with status ${uploadResponse.status}`);
      }
      const completeResponse = await completeUpload({
        object_key: presign.object_key,
        upload_token: presign.upload_token,
        mime_type: file.type,
        file_size: file.size,
      });
      const media = completeResponse.data.data;
      if (target.objectUrl) {
        URL.revokeObjectURL(target.objectUrl);
      }
      target.mediaAssetId = media.media_asset_id;
      target.url = media.url;
      target.objectKey = media.object_key;
      target.objectUrl = undefined;
      target.uploading = false;
    } catch (error) {
      target.mediaAssetId = '';
      target.url = '';
      target.objectKey = '';
      target.uploading = false;
      if (target.objectUrl) {
        URL.revokeObjectURL(target.objectUrl);
        target.objectUrl = undefined;
      }
      feedbackStore.pushToast(resolveErrorMessage(error, t('marketplace.settings.loginHeroUploadError')), 'error');
    }
  };

  // 1.3 保存登入背景圖設定
  const saveHero = async (): Promise<void> => {
    if (!canSave.value) {
      return;
    }

    saving.value = true;

    try {
      const { data } = await saveLoginHeroSettings(
        images.value.map((item, index) => ({
          media_asset_id: item.mediaAssetId,
          author: item.author.trim(),
          location: item.location.trim(),
          sort_order: index + 1,
        })),
      );
      images.value = data.data.items
        .slice()
        .sort((left, right) => left.sort_order - right.sort_order)
        .map((item, index) => ({
          key: item.media_asset_id || `login-hero-${index + 1}`,
          mediaAssetId: item.media_asset_id,
          url: item.url,
          objectKey: item.object_key,
          author: item.author,
          location: item.location,
          uploading: false,
        }));
      feedbackStore.pushToast(t('marketplace.settings.loginHeroSaveSuccess'), 'success');
    } catch (error) {
      feedbackStore.pushToast(resolveErrorMessage(error, t('marketplace.settings.loginHeroSaveError')), 'error');
    } finally {
      saving.value = false;
    }
  };

  // 1.4 新增登入背景圖位置
  const addImage = (): void => {
    if (!canAddImage.value) {
      return;
    }
    images.value = [...images.value, createEmptyHero(images.value.length + 1)];
  };

  // 1.5 移除登入背景圖位置
  const removeImage = (targetKey: string): void => {
    if (images.value.length <= 1) {
      return;
    }
    const target = images.value.find((item) => item.key === targetKey);
    if (target?.objectUrl) {
      URL.revokeObjectURL(target.objectUrl);
    }
    images.value = images.value.filter((item) => item.key !== targetKey);
  };

  // 1.6 移動登入背景圖排序
  const moveImage = (targetKey: string, direction: -1 | 1): void => {
    const index = images.value.findIndex((item) => item.key === targetKey);
    const targetIndex = index + direction;
    if (index < 0 || targetIndex < 0 || targetIndex >= images.value.length) {
      return;
    }
    const nextImages = [...images.value];
    const [item] = nextImages.splice(index, 1);
    nextImages.splice(targetIndex, 0, item);
    images.value = nextImages;
  };

  // 1.7 釋放本地圖片 URL
  const revokeObjectURLs = (): void => {
    images.value.forEach((item) => {
      if (item.objectUrl) {
        URL.revokeObjectURL(item.objectUrl);
      }
    });
  };

  onMounted(() => {
    void loadHero();
  });

  onBeforeUnmount(() => {
    revokeObjectURLs();
  });

  return {
    addImage,
    canAddImage,
    canSave,
    handleFileSelected,
    images,
    loadHero,
    loading,
    moveImage,
    removeImage,
    saveHero,
    saving,
    t,
  };
};

// 2. 解析 API 錯誤訊息
const resolveErrorMessage = (error: unknown, fallback: string): string =>
  axios.isAxiosError(error) ? error.response?.data?.message ?? fallback : fallback;

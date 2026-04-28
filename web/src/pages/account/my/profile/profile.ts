/*
 * 會員中心個人資料 - 狀態與資料流程。
 * 1. 讀取並同步會員資料表單。
 * 2. 上傳 OSS 頭像並同步會員狀態。
 * 3. 儲存會員資料並同步全域會員狀態。
 */
import { computed, onMounted, reactive, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

import { updateMe } from '@/httpapis/me';
import { completeUpload, createUploadPresign } from '@/httpapis/uploads';
import { useFeedbackStore } from '@/stores/feedback';
import { useSessionStore } from '@/stores/session';

const avatarObjectPrefix = 'ajo_living/account/';
const maxAvatarFileSize = 5 * 1024 * 1024;
const blockedUploadHeaders = new Set(['host', 'content-length']);

// 1. 建立 OSS 直傳 headers
const buildUploadHeaders = (headers: Record<string, string>, mimeType: string): Headers => {
  const result = new Headers();

  Object.entries(headers).forEach(([key, value]) => {
    if (!blockedUploadHeaders.has(key.toLowerCase())) {
      result.set(key, value);
    }
  });

  if (!result.has('Content-Type')) {
    result.set('Content-Type', mimeType);
  }

  return result;
};

// 2. 管理個人資料頁資料與動作
export const useAccountProfilePage = () => {
  const { t } = useI18n();
  const router = useRouter();
  const feedbackStore = useFeedbackStore();
  const sessionStore = useSessionStore();
  const isSaving = ref(false);
  const isLoading = ref(false);
  const isUploadingAvatar = ref(false);
  const isSigningOut = ref(false);

  const formState = reactive({
    display_name: '',
    email: '',
    phone: '',
    district_code: '',
    publisher_identity_type: '',
  });

  // 2.1 顯示帳戶電話資料
  const phoneDisplay = computed(() => {
    if (sessionStore.me?.phone_country_code === 'email') {
      return t('account.profile.phoneUnavailable');
    }

    return formState.phone || t('account.profile.phoneUnavailable');
  });

  // 2.2 同步表單內容
  const syncFormState = (): void => {
    formState.display_name = sessionStore.me?.display_name ?? sessionStore.currentUser.display_name;
    formState.email = sessionStore.me?.email ?? '';
    formState.phone = `${sessionStore.me?.phone_country_code ?? '+852'} ${sessionStore.me?.phone_number ?? ''}`.trim();
    formState.district_code = sessionStore.me?.district_code ?? '';
    formState.publisher_identity_type = sessionStore.me?.publisher_identity_type ?? '';
  };

  // 2.3 讀取會員資料
  const loadProfile = async (): Promise<void> => {
    if (sessionStore.me) {
      syncFormState();
      return;
    }

    isLoading.value = true;

    try {
      await sessionStore.loadCurrentUser();
      syncFormState();
    } catch (error) {
      console.error(error);
      feedbackStore.pushToast(t('account.profile.loadError'), 'error');
    } finally {
      isLoading.value = false;
    }
  };

  // 2.4 上傳頭像並更新會員資料
  const handleAvatarFileChange = async (event: Event): Promise<void> => {
    const input = event.target as HTMLInputElement;
    const file = input.files?.[0];
    input.value = '';

    if (!file) {
      return;
    }
    if (!file.type.startsWith('image/')) {
      feedbackStore.pushToast(t('account.profile.avatarFileInvalid'), 'error');
      return;
    }
    if (file.size > maxAvatarFileSize) {
      feedbackStore.pushToast(t('account.profile.avatarFileTooLarge'), 'error');
      return;
    }

    isUploadingAvatar.value = true;

    try {
      const presignResponse = await createUploadPresign({
        file_name: file.name,
        mime_type: file.type,
        file_size: file.size,
        object_prefix: avatarObjectPrefix,
      });
      const presign = presignResponse.data.data;

      const uploadResponse = await fetch(presign.upload_url, {
        method: 'PUT',
        headers: buildUploadHeaders(presign.headers, file.type),
        body: file,
      });
      if (!uploadResponse.ok) {
        throw new Error(`avatar upload failed with status ${uploadResponse.status}`);
      }

      const completeResponse = await completeUpload({
        object_key: presign.object_key,
        mime_type: file.type,
        file_size: file.size,
      });
      const profileResponse = await updateMe({
        display_name: formState.display_name.trim(),
        publisher_identity_type: formState.publisher_identity_type.trim(),
        district_code: formState.district_code.trim(),
        avatar_asset_id: completeResponse.data.data.media_asset_id,
      });

      sessionStore.me = profileResponse.data.data;
      syncFormState();
      feedbackStore.pushToast(t('account.profile.avatarUploadSuccess'), 'success');
    } catch (error) {
      console.error(error);
      feedbackStore.pushToast(t('account.profile.avatarUploadError'), 'error');
    } finally {
      isUploadingAvatar.value = false;
    }
  };

  // 2.5 儲存會員資料
  const handleSaveProfile = async (): Promise<void> => {
    isSaving.value = true;

    try {
      const { data } = await updateMe({
        display_name: formState.display_name.trim(),
        publisher_identity_type: formState.publisher_identity_type.trim(),
        district_code: formState.district_code.trim(),
      });

      sessionStore.me = data.data;
      syncFormState();
      feedbackStore.pushToast(t('account.profile.updateSuccess'), 'success');
    } catch (error) {
      console.error(error);
      feedbackStore.pushToast(t('account.profile.updateError'), 'error');
    } finally {
      isSaving.value = false;
    }
  };

  // 2.6 執行登出
  const handleSignOut = async (): Promise<void> => {
    isSigningOut.value = true;

    try {
      await sessionStore.signOut();
      await router.push('/login');
    } finally {
      isSigningOut.value = false;
    }
  };

  // 2.7 初始化會員資料
  onMounted(() => {
    void loadProfile();
  });

  return {
    formState,
    handleAvatarFileChange,
    handleSaveProfile,
    handleSignOut,
    isLoading,
    isSaving,
    isSigningOut,
    isUploadingAvatar,
    phoneDisplay,
    sessionStore,
    t,
  };
};

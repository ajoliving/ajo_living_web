/*
 * 積分營運設定 - 共享資料流程。
 * 1. 提供積分發放、廣告發布、廣告列表與積分流水頁共用型別。
 * 2. 封裝 Staff 錢包與廣告任務 API 操作。
 * 3. 管理廣告圖片與影片本地暫存及保存時 OSS 上傳流程。
 */
import axios from 'axios';
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import { fetchStaffUsers } from '@/httpapis/staff';
import { completeUpload, createUploadPresign } from '@/httpapis/uploads';
import {
  createStaffRewardAd,
  fetchStaffRewardAd,
  fetchStaffRewardAds,
  fetchStaffWalletTransactions,
  grantStaffWalletPoints,
  updateStaffRewardAd,
} from '@/httpapis/wallet';
import type {
  StaffRewardAdPayload,
  StaffRewardAdResponse,
  StaffWalletTransactionResponse,
} from '@/model/wallet';
import type { StaffUserSummary } from '@/model/user';
import { useFeedbackStore } from '@/stores/feedback';
import { formatDate } from '@/utils/format';
import { buildUploadHeaders } from '@/utils/upload';
import { formatAjoPoints } from '@/utils/wallet';

type WalletDirectionFilter = '' | 'credit' | 'debit';
export type RewardAdMediaType = 'image' | 'video';

export interface GrantForm {
  userId: string;
  amount: number;
  note: string;
}

export interface StaffWalletGrantRow {
  user: StaffUserSummary;
  amount: number;
  note: string;
  granting: boolean;
}

export interface RewardAdForm {
  taskId: string;
  title: string;
  summary: string;
  mediaURL: string;
  mediaType: RewardAdMediaType;
  mediaFile?: File;
  mediaPreviewURL: string;
  targetURL: string;
  rewardPoints: number;
  watchSeconds: number;
  retentionDays: number;
  isActive: boolean;
}

const advertisementImageObjectPrefix = 'ajo_living/advertisements/images/';
const advertisementVideoObjectPrefix = 'ajo_living/advertisements/video/';

// 1. 建立空廣告表單
export const emptyRewardAdForm = (): RewardAdForm => ({
  taskId: '',
  title: '',
  summary: '',
  mediaURL: '',
  mediaType: 'image',
  mediaFile: undefined,
  mediaPreviewURL: '',
  targetURL: '',
  rewardPoints: 50,
  watchSeconds: 30,
  retentionDays: 30,
  isActive: true,
});

// 2. 建立積分發放頁流程
export const useWalletGrantPage = () => {
  const { t } = useI18n();
  const feedbackStore = useFeedbackStore();
  const loadingUsers = ref(false);
  const userKeyword = ref('');
  const userPage = ref(1);
  const userPageSize = 20;
  const userTotal = ref(0);
  const grantRows = ref<StaffWalletGrantRow[]>([]);
  const hasPreviousUsers = computed(() => userPage.value > 1);
  const hasNextUsers = computed(() => userPage.value * userPageSize < userTotal.value);

  const loadGrantUsers = async (page = userPage.value): Promise<void> => {
    loadingUsers.value = true;

    try {
      const previousRows = new Map(grantRows.value.map((row) => [row.user.public_id, row]));
      const { data } = await fetchStaffUsers({
        page,
        page_size: userPageSize,
        keyword: userKeyword.value.trim() || undefined,
      });

      userPage.value = data.data.pagination.page;
      userTotal.value = data.data.pagination.total;
      grantRows.value = data.data.items.map((user) => {
        const existingRow = previousRows.get(user.public_id);
        return {
          user,
          amount: existingRow?.amount ?? 100,
          note: existingRow?.note ?? '',
          granting: false,
        };
      });
    } catch (error) {
      feedbackStore.pushToast(readErrorMessage(error, t('marketplace.management.walletLoadMembersError')), 'error');
    } finally {
      loadingUsers.value = false;
    }
  };

  const searchGrantUsers = async (): Promise<void> => {
    await loadGrantUsers(1);
  };

  const loadPreviousGrantUsers = async (): Promise<void> => {
    if (!hasPreviousUsers.value) {
      return;
    }
    await loadGrantUsers(userPage.value - 1);
  };

  const loadNextGrantUsers = async (): Promise<void> => {
    if (!hasNextUsers.value) {
      return;
    }
    await loadGrantUsers(userPage.value + 1);
  };

  const grantPoints = async (row: StaffWalletGrantRow): Promise<void> => {
    const userId = row.user.public_id.trim();
    const note = row.note.trim();
    if (!userId || !Number.isFinite(row.amount) || row.amount <= 0 || !note) {
      feedbackStore.pushToast(t('marketplace.management.walletGrantRequired'), 'error');
      return;
    }

    row.granting = true;
    try {
      await grantStaffWalletPoints({ user_id: userId, amount: row.amount, note });
      row.amount = 100;
      row.note = '';
      feedbackStore.pushToast(t('marketplace.management.walletGrantSuccess'), 'success');
    } catch (error) {
      feedbackStore.pushToast(readErrorMessage(error, t('marketplace.management.walletGrantError')), 'error');
    } finally {
      row.granting = false;
    }
  };

  onMounted(() => {
    void loadGrantUsers(1);
  });

  return {
    grantRows,
    grantPoints,
    hasNextUsers,
    hasPreviousUsers,
    loadGrantUsers,
    loadNextGrantUsers,
    loadingUsers,
    loadPreviousGrantUsers,
    searchGrantUsers,
    t,
    userKeyword,
    userPage,
    userPageSize,
    userTotal,
  };
};

// 3. 建立廣告發布與編輯頁流程
export const useRewardAdEditorPage = () => {
  const route = useRoute();
  const router = useRouter();
  const { t } = useI18n();
  const feedbackStore = useFeedbackStore();
  const loadingAd = ref(false);
  const savingAd = ref(false);
  const uploadingMedia = ref(false);
  const adForm = reactive<RewardAdForm>(emptyRewardAdForm());
  const taskId = computed(() => String(route.params.taskId ?? '').trim());
  const isEditingAd = computed(() => taskId.value.length > 0 || adForm.taskId.trim().length > 0);

  const loadRewardAd = async (): Promise<void> => {
    if (!taskId.value) {
      revokeRewardAdPreview(adForm);
      Object.assign(adForm, emptyRewardAdForm());
      return;
    }

    loadingAd.value = true;
    try {
      const { data } = await fetchStaffRewardAd(taskId.value);
      applyRewardAdToForm(data.data, adForm);
    } catch (error) {
      feedbackStore.pushToast(readErrorMessage(error, t('marketplace.management.walletLoadAdsError')), 'error');
    } finally {
      loadingAd.value = false;
    }
  };

  const resetRewardAdForm = (): void => {
    revokeRewardAdPreview(adForm);
    Object.assign(adForm, emptyRewardAdForm());
  };

  const stageRewardAdMedia = (file: File): void => {
    const expectedType = adForm.mediaType;
    if (expectedType === 'image' && !file.type.startsWith('image/')) {
      feedbackStore.pushToast(t('marketplace.management.walletAdImageRequired'), 'error');
      return;
    }
    if (expectedType === 'video' && !file.type.startsWith('video/')) {
      feedbackStore.pushToast(t('marketplace.management.walletAdVideoRequired'), 'error');
      return;
    }

    revokeRewardAdPreview(adForm);
    adForm.mediaFile = file;
    adForm.mediaPreviewURL = URL.createObjectURL(file);
    adForm.mediaURL = '';
  };

  const saveRewardAd = async (): Promise<void> => {
    if (!isRewardAdFormValid(adForm, isEditingAd.value)) {
      feedbackStore.pushToast(t('marketplace.management.walletAdRequired'), 'error');
      return;
    }

    savingAd.value = true;
    try {
      uploadingMedia.value = Boolean(adForm.mediaFile);
      const mediaURL = adForm.mediaFile
        ? await uploadAdvertisementMedia(adForm.mediaFile, adForm.mediaType)
        : adForm.mediaURL.trim();
      uploadingMedia.value = false;
      const payload = buildRewardAdPayload(adForm, mediaURL);

      if (isEditingAd.value) {
        await updateStaffRewardAd(adForm.taskId || taskId.value, payload);
      } else {
        await createStaffRewardAd(payload);
      }
      revokeRewardAdPreview(adForm);
      feedbackStore.pushToast(t('marketplace.management.walletAdSaveSuccess'), 'success');
      await router.push('/account/marketplace/management/reward-ads');
    } catch (error) {
      uploadingMedia.value = false;
      feedbackStore.pushToast(readErrorMessage(error, t('marketplace.management.walletAdSaveError')), 'error');
    } finally {
      savingAd.value = false;
    }
  };

  watch(taskId, () => {
    void loadRewardAd();
  });

  watch(
    () => adForm.mediaType,
    () => {
      if (loadingAd.value) {
        return;
      }
      revokeRewardAdPreview(adForm);
      adForm.mediaFile = undefined;
      adForm.mediaURL = '';
    },
    { flush: 'sync' },
  );

  onMounted(() => {
    void loadRewardAd();
  });

  onBeforeUnmount(() => {
    revokeRewardAdPreview(adForm);
  });

  return {
    adForm,
    formatAjoPoints,
    isEditingAd,
    loadingAd,
    resetRewardAdForm,
    saveRewardAd,
    savingAd,
    stageRewardAdMedia,
    t,
    uploadingMedia,
  };
};

// 4. 建立廣告列表頁流程
export const useRewardAdListPage = () => {
  const router = useRouter();
  const { t } = useI18n();
  const feedbackStore = useFeedbackStore();
  const loadingAds = ref(false);
  const adKeyword = ref('');
  const rewardAds = ref<StaffRewardAdResponse[]>([]);
  const formatPoints = (value: number): string =>
    formatAjoPoints(value, t('common.brand.pointsName'), 'zh-HK');

  const loadRewardAds = async (): Promise<void> => {
    loadingAds.value = true;

    try {
      const { data } = await fetchStaffRewardAds({
        page: 1,
        page_size: 20,
        keyword: adKeyword.value.trim() || undefined,
      });
      rewardAds.value = data.data.items;
    } catch (error) {
      feedbackStore.pushToast(readErrorMessage(error, t('marketplace.management.walletLoadAdsError')), 'error');
    } finally {
      loadingAds.value = false;
    }
  };

  const editRewardAd = async (ad: StaffRewardAdResponse): Promise<void> => {
    await router.push(`/account/marketplace/management/reward-ad-editor/${ad.task_id}`);
  };

  const toggleRewardAd = async (ad: StaffRewardAdResponse): Promise<void> => {
    try {
      await updateStaffRewardAd(ad.task_id, { is_active: !ad.is_active });
      await loadRewardAds();
      feedbackStore.pushToast(t('marketplace.management.walletAdStatusUpdated'), 'success');
    } catch (error) {
      feedbackStore.pushToast(readErrorMessage(error, t('marketplace.management.walletAdSaveError')), 'error');
    }
  };

  onMounted(() => {
    void loadRewardAds();
  });

  return {
    adKeyword,
    editRewardAd,
    formatAjoPoints: formatPoints,
    loadRewardAds,
    loadingAds,
    rewardAds,
    t,
    toggleRewardAd,
  };
};

// 5. 建立積分流水頁流程
export const useWalletTransactionPage = () => {
  const { t } = useI18n();
  const feedbackStore = useFeedbackStore();
  const loadingTransactions = ref(false);
  const transactionUserId = ref('');
  const transactionDirection = ref<WalletDirectionFilter>('');
  const transactions = ref<StaffWalletTransactionResponse[]>([]);
  const formatPoints = (value: number): string =>
    formatAjoPoints(value, t('common.brand.pointsName'), 'zh-HK');

  const loadTransactions = async (): Promise<void> => {
    loadingTransactions.value = true;

    try {
      const { data } = await fetchStaffWalletTransactions({
        page: 1,
        page_size: 20,
        user_id: transactionUserId.value.trim() || undefined,
        direction: transactionDirection.value || undefined,
      });
      transactions.value = data.data.items;
    } catch (error) {
      feedbackStore.pushToast(readErrorMessage(error, t('marketplace.management.walletLoadTransactionsError')), 'error');
    } finally {
      loadingTransactions.value = false;
    }
  };

  onMounted(() => {
    void loadTransactions();
  });

  return {
    formatAjoPoints: formatPoints,
    formatDate,
    loadTransactions,
    loadingTransactions,
    t,
    transactionDirection,
    transactions,
    transactionUserId,
  };
};

// 6. 將廣告資料寫入表單
const applyRewardAdToForm = (ad: StaffRewardAdResponse, form: RewardAdForm): void => {
  revokeRewardAdPreview(form);
  form.taskId = ad.task_id;
  form.title = ad.title;
  form.summary = ad.summary;
  form.mediaURL = ad.media_url;
  form.mediaType = ad.media_type;
  form.mediaFile = undefined;
  form.mediaPreviewURL = '';
  form.targetURL = ad.target_url;
  form.rewardPoints = ad.reward_points;
  form.watchSeconds = ad.watch_seconds;
  form.retentionDays = resolveRetentionDays(ad.ends_at);
  form.isActive = ad.is_active;
};

// 7. 建立廣告保存 payload
const buildRewardAdPayload = (form: RewardAdForm, mediaURL: string): StaffRewardAdPayload => ({
  title: form.title.trim(),
  summary: form.summary.trim(),
  cover_url: '',
  media_url: mediaURL.trim(),
  media_type: form.mediaType,
  target_url: form.targetURL.trim(),
  reward_points: Number(form.rewardPoints),
  watch_seconds: Number(form.watchSeconds),
  total_budget: 0,
  retention_days: Number(form.retentionDays),
  is_active: form.isActive,
});

// 8. 判斷廣告表單是否可提交
const isRewardAdFormValid = (form: RewardAdForm, isEditingAd: boolean): boolean =>
  form.title.trim().length > 0 &&
  form.summary.trim().length > 0 &&
  Number.isFinite(Number(form.rewardPoints)) &&
  Number(form.rewardPoints) > 0 &&
  Number.isFinite(Number(form.watchSeconds)) &&
  Number(form.watchSeconds) > 0 &&
  Number.isFinite(Number(form.retentionDays)) &&
  Number(form.retentionDays) > 0 &&
  (Boolean(form.mediaFile) || (isEditingAd && form.mediaURL.trim().length > 0));

// 9. 上傳廣告媒體到 OSS
const uploadAdvertisementMedia = async (file: File, mediaType: RewardAdMediaType): Promise<string> => {
  const presignResponse = await createUploadPresign({
    file_name: file.name,
    mime_type: file.type,
    file_size: file.size,
    object_prefix: mediaType === 'video' ? advertisementVideoObjectPrefix : advertisementImageObjectPrefix,
  });
  const presign = presignResponse.data.data;
  const uploadResponse = await fetch(presign.upload_url, {
    method: 'PUT',
    headers: buildUploadHeaders(presign.headers, file.type),
    body: file,
  });

  if (!uploadResponse.ok) {
    throw new Error(`advertisement media upload failed with status ${uploadResponse.status}`);
  }

  const completeResponse = await completeUpload({
    object_key: presign.object_key,
    mime_type: file.type,
    file_size: file.size,
  });

  return completeResponse.data.data.url;
};

// 10. 釋放廣告媒體本地預覽
const revokeRewardAdPreview = (form: RewardAdForm): void => {
  if (form.mediaPreviewURL) {
    URL.revokeObjectURL(form.mediaPreviewURL);
    form.mediaPreviewURL = '';
  }
};

// 11. 從到期時間換算剩餘保留天數
const resolveRetentionDays = (endsAt?: string): number => {
  if (!endsAt) {
    return 30;
  }

  const endTime = new Date(endsAt).getTime();
  if (!Number.isFinite(endTime)) {
    return 30;
  }

  const dayMs = 24 * 60 * 60 * 1000;
  return Math.max(1, Math.ceil((endTime - Date.now()) / dayMs));
};

// 12. 解析 API 錯誤訊息
const readErrorMessage = (error: unknown, fallback: string): string =>
  axios.isAxiosError<{ message?: string }>(error)
    ? error.response?.data?.message ?? fallback
    : fallback;

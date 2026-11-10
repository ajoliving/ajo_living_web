/*
 * 發布通知設定資料流程。
 * 1. 管理通知表單輸入與基礎校驗。
 * 2. 調用 staff 通知接口發布站內通知。
 */
import axios from 'axios';
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';

import { publishSystemNotice } from '@/httpapis/notifications';
import { useFeedbackStore } from '@/stores/feedback';

// 1. 建立發布通知流程
export const useMarketplaceSettingsNoticePage = () => {
  const { t } = useI18n();
  const feedbackStore = useFeedbackStore();
  const publishingNotice = ref(false);
  const noticeTitle = ref('');
  const noticeBody = ref('');

  // 1.1 發布系統通知
  const publishNotice = async (): Promise<void> => {
    const title = noticeTitle.value.trim();
    const body = noticeBody.value.trim();
    if (!title || !body) {
      feedbackStore.pushToast(t('marketplace.settings.noticeRequired'), 'error');
      return;
    }

    publishingNotice.value = true;

    try {
      const { data } = await publishSystemNotice({
        title,
        body,
      });
      noticeTitle.value = '';
      noticeBody.value = '';
      feedbackStore.pushToast(
        t('marketplace.settings.noticePublishSuccess', { count: data.data.delivered_count }),
        'success',
      );
    } catch (error) {
      feedbackStore.pushToast(resolveErrorMessage(error, t('marketplace.settings.noticePublishError')), 'error');
    } finally {
      publishingNotice.value = false;
    }
  };

  return {
    noticeBody,
    noticeTitle,
    publishNotice,
    publishingNotice,
    t,
  };
};

// 2. 解析 API 錯誤訊息
const resolveErrorMessage = (error: unknown, fallback: string): string =>
  axios.isAxiosError(error) ? error.response?.data?.message ?? fallback : fallback;

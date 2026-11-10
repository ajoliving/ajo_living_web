<!--
 * 會員中心提示設定頁。
 * 1. 讀取並更新 iSmart 大廈通告電郵設定。
-->
<script setup lang="ts">
/*
 * 會員中心提示設定頁邏輯。
 * 1. 提供頁面所需的多語系文字。
 */
import { onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import axios from 'axios';
import { fetchMemberIsmartNotificationSettings, updateMemberIsmartNotificationSettings } from '@/httpapis/me';
import { useFeedbackStore } from '@/stores/feedback';

const { t } = useI18n();
const feedbackStore = useFeedbackStore();
const isLoading = ref(true);
const isSavingEmail = ref(false);
const receiveEmail = ref(true);

const readError = (error: unknown, fallback: string): string =>
  axios.isAxiosError<{ message?: string }>(error) ? error.response?.data?.message ?? fallback : fallback;

const loadSettings = async (): Promise<void> => {
  isLoading.value = true;
  try {
    const { data } = await fetchMemberIsmartNotificationSettings();
    receiveEmail.value = data.data.is_receive_email ?? data.data.blg_notice_email;
  } catch (error) {
    feedbackStore.pushToast(readError(error, t('account.preferences.loadError')), 'error');
  } finally {
    isLoading.value = false;
  }
};

const saveEmailSetting = async (): Promise<void> => {
  if (isSavingEmail.value) return;
  isSavingEmail.value = true;
  try {
    const { data } = await updateMemberIsmartNotificationSettings(receiveEmail.value);
    receiveEmail.value = data.data.is_receive_email ?? data.data.blg_notice_email;
    feedbackStore.pushToast(t('account.preferences.emailSaved'), 'success');
  } catch (error) {
    feedbackStore.pushToast(readError(error, t('account.preferences.emailSaveError')), 'error');
  } finally {
    isSavingEmail.value = false;
  }
};

onMounted(() => { void loadSettings(); });
</script>

<template>
  <section class="preferences-page">
    <header class="preferences-page__header">
      <h2>{{ t('account.center.nav.preferences') }}</h2>
    </header>

    <section class="preferences-page__card" :aria-busy="isLoading">
      <label class="preferences-page__option">
        <input v-model="receiveEmail" type="checkbox" :disabled="isLoading || isSavingEmail">
        <span>{{ t('account.center.account.receiveNoticeEmail') }}</span>
      </label>
      <button type="button" class="preferences-page__submit" :disabled="isLoading || isSavingEmail" @click="saveEmailSetting">
        {{ isSavingEmail ? t('account.center.common.saving') : t('account.center.common.submit') }}
      </button>
    </section>
  </section>
</template>

<style scoped>
/*
 * 會員中心提示設定頁樣式。
 * 1. 對齊會員中心卡片與操作控件層級。
 * 2. 保持窄螢幕下的垂直排列與可操作性。
 */

/* 1. 頁面結構 */
.preferences-page {
  display: grid;
  gap: 14px;
}

.preferences-page__header {
  border-bottom: 1px solid var(--bdr);
  padding-bottom: 12px;
}

.preferences-page__header h2 {
  margin: 0;
  color: var(--ink);
  font-size: 20px;
  font-weight: 600;
  line-height: 1.25;
}

/* 2. 提示設定卡片 */
.preferences-page__card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
  border: 1px solid var(--bdr);
  border-radius: 8px;
  background: #fff;
  padding: 14px;
}

.preferences-page__option {
  display: flex;
  align-items: center;
  gap: 10px;
  color: var(--ink);
  font-size: 14px;
  font-weight: 700;
}

.preferences-page__option input {
  width: 16px;
  height: 16px;
  accent-color: var(--brand);
}

.preferences-page__submit {
  min-height: 38px;
  border: 0;
  border-radius: 6px;
  background: var(--accent);
  color: #fff;
  cursor: pointer;
  font-family: inherit;
  font-size: 13px;
  font-weight: 700;
  padding: 9px 14px;
  white-space: nowrap;
}

.preferences-page__submit:disabled {
  cursor: not-allowed;
  opacity: 0.55;
}

.preferences-page__submit:focus-visible {
  outline: 2px solid var(--brand);
  outline-offset: 2px;
}

/* 3. 行動裝置佈局 */
@media (max-width: 560px) {
  .preferences-page__card {
    align-items: stretch;
    flex-direction: column;
  }

  .preferences-page__submit {
    width: 100%;
  }
}
</style>

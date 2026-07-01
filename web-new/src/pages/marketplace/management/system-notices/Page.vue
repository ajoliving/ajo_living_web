<!--
 * 管理 - 系統通知頁。
 * 1. 提供 Staff 發布全站系統通知表單。
 * 2. 調用 staff 系統通知接口並回饋送達人數。
-->
<script setup lang="ts">
import axios from 'axios';
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';

import { publishSystemNotice } from '@/httpapis/notifications';
import AppIcon from '@/shared/components/base/AppIcon.vue';
import { useFeedbackStore } from '@/stores/feedback';

import '../styles.scss';

const { t } = useI18n();
const feedbackStore = useFeedbackStore();
const publishing = ref(false);
const noticeTitle = ref('');
const noticeBody = ref('');
const canPublish = computed(() =>
  noticeTitle.value.trim().length > 0 &&
  noticeBody.value.trim().length > 0 &&
  !publishing.value,
);

// 1. 發布系統通知
const publishNotice = async (): Promise<void> => {
  const title = noticeTitle.value.trim();
  const body = noticeBody.value.trim();

  if (!title || !body) {
    feedbackStore.pushToast(t('marketplace.settings.noticeRequired'), 'error');
    return;
  }

  publishing.value = true;
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
  } catch (error: unknown) {
    feedbackStore.pushToast(resolveErrorMessage(error), 'error');
  } finally {
    publishing.value = false;
  }
};

// 2. 解析接口錯誤訊息
const resolveErrorMessage = (error: unknown): string =>
  axios.isAxiosError<{ message?: string }>(error)
    ? error.response?.data?.message ?? t('marketplace.settings.noticePublishError')
    : t('marketplace.settings.noticePublishError');
</script>

<template>
  <section class="management-list-page">
    <header class="management-list-header">
      <div>
        <p class="management-list-kicker">
          Staff
        </p>
        <h1>{{ t('marketplace.management.systemNotices') }}</h1>
        <p>{{ t('marketplace.management.systemNoticesDescription') }}</p>
      </div>
    </header>

    <article class="management-list-panel">
      <div class="system-notice-form">
        <label class="system-notice-field">
          <span>{{ t('marketplace.settings.noticeTitleField') }}</span>
          <input
            v-model="noticeTitle"
            type="text"
            maxlength="120"
            :placeholder="t('marketplace.settings.noticeTitlePlaceholder')"
          />
        </label>

        <label class="system-notice-field">
          <span>{{ t('marketplace.settings.noticeBodyField') }}</span>
          <textarea
            v-model="noticeBody"
            maxlength="1000"
            rows="7"
            :placeholder="t('marketplace.settings.noticeBodyPlaceholder')"
          />
        </label>

        <div class="system-notice-footer">
          <button
            type="button"
            class="management-list-button"
            :disabled="!canPublish"
            @click="publishNotice"
          >
            <AppIcon
              name="send"
              :size="16"
            />
            <span>{{ publishing ? t('marketplace.settings.noticePublishing') : t('marketplace.settings.noticePublishAction') }}</span>
          </button>
        </div>
      </div>
    </article>
  </section>
</template>

<style scoped>
.system-notice-form {
  display: grid;
  gap: 0.85rem;
  padding: 1rem;
}

.system-notice-field {
  display: grid;
  gap: 0.45rem;
}

.system-notice-field span {
  color: rgb(var(--color-text-muted));
  font-size: 0.78rem;
  font-weight: 700;
}

.system-notice-field input,
.system-notice-field textarea {
  width: 100%;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text));
  font-size: 0.92rem;
  line-height: 1.5;
  outline: 0;
  padding: 0.75rem;
}

.system-notice-field textarea {
  resize: vertical;
}

.system-notice-footer {
  display: flex;
  justify-content: flex-end;
}

@media (max-width: 767px) {
  .system-notice-footer .management-list-button {
    width: 100%;
  }
}
</style>

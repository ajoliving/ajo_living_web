<!--
 * 發布通知設定頁。
 * 1. 提供站內通知標題、內容與可選行動按鈕表單。
 * 2. 發布後清空表單並提示送達人數。
-->
<script setup lang="ts">
import AppIcon from '@/shared/components/base/AppIcon.vue';

import { useManagementSettingsNoticePage } from './notice';

const {
  noticeActionLabel,
  noticeActionURL,
  noticeBody,
  noticeTitle,
  publishNotice,
  publishingNotice,
  t,
} = useManagementSettingsNoticePage();
</script>

<template>
  <section class="notice-page">
    <div class="notice-header">
      <div>
        <p class="notice-kicker">
          {{ t('marketplace.settings.noticeSection') }}
        </p>
        <h2>{{ t('marketplace.settings.noticePublisherTitle') }}</h2>
        <p>{{ t('marketplace.settings.noticePublisherDescription') }}</p>
      </div>
    </div>

    <section class="notice-panel">
      <div class="notice-form">
        <label class="notice-field">
          <span>{{ t('marketplace.settings.noticeTitleField') }}</span>
          <input
            v-model="noticeTitle"
            type="text"
            maxlength="120"
            :placeholder="t('marketplace.settings.noticeTitlePlaceholder')"
          />
        </label>

        <label class="notice-field">
          <span>{{ t('marketplace.settings.noticeBodyField') }}</span>
          <textarea
            v-model="noticeBody"
            maxlength="1000"
            rows="7"
            :placeholder="t('marketplace.settings.noticeBodyPlaceholder')"
          />
        </label>

        <div class="notice-action-grid">
          <label class="notice-field">
            <span>{{ t('marketplace.settings.noticeActionLabelField') }}</span>
            <input
              v-model="noticeActionLabel"
              type="text"
              maxlength="80"
              :placeholder="t('marketplace.settings.noticeActionLabelPlaceholder')"
            />
          </label>
          <label class="notice-field">
            <span>{{ t('marketplace.settings.noticeActionURLField') }}</span>
            <input
              v-model="noticeActionURL"
              type="text"
              maxlength="500"
              :placeholder="t('marketplace.settings.noticeActionURLPlaceholder')"
            />
          </label>
        </div>

        <div class="notice-footer">
          <button
            type="button"
            class="notice-button notice-button--primary"
            :disabled="publishingNotice"
            @click="publishNotice"
          >
            <AppIcon
              name="send"
              :size="16"
            />
            <span>{{ publishingNotice ? t('marketplace.settings.noticePublishing') : t('marketplace.settings.noticePublishAction') }}</span>
          </button>
        </div>
      </div>
    </section>
  </section>
</template>

<style scoped>
.notice-page {
  display: grid;
  gap: 1rem;
}

.notice-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
}

.notice-kicker {
  margin: 0;
  color: rgb(var(--color-text-muted));
  font-size: 0.75rem;
  font-weight: 800;
  letter-spacing: 0.1em;
  line-height: 1;
  text-transform: uppercase;
}

.notice-header h2 {
  margin: 0.35rem 0 0;
  color: rgb(var(--color-text));
  font-family: var(--font-display);
  font-size: clamp(1.5rem, 2vw, 2.25rem);
  font-weight: 650;
  line-height: 1.2;
}

.notice-header p:not(.notice-kicker) {
  margin: 0.45rem 0 0;
  color: rgb(var(--color-text-muted));
  font-size: 0.9rem;
  line-height: 1.7;
}

.notice-panel {
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface-raised));
}

.notice-form {
  display: grid;
  gap: 0.85rem;
  padding: 1rem;
}

.notice-action-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.75rem;
}

.notice-field {
  display: grid;
  gap: 0.45rem;
}

.notice-field span {
  color: rgb(var(--color-text-muted));
  font-size: 0.78rem;
  font-weight: 700;
}

.notice-field input,
.notice-field textarea {
  width: 100%;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
  padding: 0.75rem;
  color: rgb(var(--color-text));
  font-size: 0.92rem;
  line-height: 1.5;
  outline: 0;
}

.notice-field textarea {
  resize: vertical;
}

.notice-footer {
  display: flex;
  justify-content: flex-end;
}

.notice-button {
  display: inline-flex;
  min-height: 2.75rem;
  align-items: center;
  justify-content: center;
  gap: 0.45rem;
  border: 1px solid rgb(var(--color-text));
  border-radius: 8px;
  padding: 0 1rem;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text));
  font-size: 0.9rem;
  font-weight: 700;
}

.notice-button--primary {
  background: rgb(var(--color-text));
  color: rgb(var(--color-surface));
}

.notice-button:disabled {
  cursor: not-allowed;
  opacity: 0.45;
}

@media (max-width: 767px) {
  .notice-header {
    flex-direction: column;
  }

  .notice-action-grid {
    grid-template-columns: 1fr;
  }

  .notice-button {
    width: 100%;
  }
}
</style>

<!--
 * 聊天訊息氣泡。
 * 1. 區分自己與對方的訊息樣式。
 * 2. 統一時間與氣泡留白節奏。
-->
<script setup lang="ts">
import { computed } from 'vue';
import { RouterLink } from 'vue-router';

import type { ChatMessageView } from '@/domains/communications/model';
import { usePreferenceStore } from '@/app/stores/preferences';

interface MessageBubbleProps {
  message: ChatMessageView;
}

const props = defineProps<MessageBubbleProps>();
const preferenceStore = usePreferenceStore();

// 1. 判斷訊息方向
const isSelf = computed(() => props.message.sender_role === 'self');
const isNoticeCard = computed(() => props.message.message_type === 'notice_card');
const noticeLines = computed(() => props.message.body.split('\n').filter(Boolean));

// 2. 格式化訊息時間
const formattedTime = computed(() =>
  new Intl.DateTimeFormat(preferenceStore.locale, {
    hour: '2-digit',
    minute: '2-digit',
  }).format(new Date(props.message.sent_at)),
);
</script>

<template>
  <div
    class="message-bubble-row"
    :class="isNoticeCard ? 'message-bubble-row--notice' : isSelf ? 'message-bubble-row--self' : 'message-bubble-row--peer'"
  >
    <div
      v-if="isNoticeCard"
      class="notice-card"
    >
      <p class="notice-card__title">
        {{ noticeLines[0] }}
      </p>
      <p
        v-if="noticeLines[1]"
        class="notice-card__body"
      >
        {{ noticeLines[1] }}
      </p>
      <RouterLink
        v-if="props.message.action_label && props.message.action_url"
        :to="props.message.action_url"
        class="notice-card__action"
      >
        {{ props.message.action_label }}
      </RouterLink>
      <p class="notice-card__time">
        {{ formattedTime }}
      </p>
    </div>

    <div
      v-else
      class="message-bubble"
      :class="isSelf ? 'message-bubble--self' : 'message-bubble--peer'"
    >
      <p
        class="message-bubble__time"
        :class="isSelf ? 'message-bubble__time--self' : 'message-bubble__time--peer'"
      >
        {{ formattedTime }}
      </p>
      <p class="message-bubble__body">
        {{ props.message.body }}
      </p>
    </div>
  </div>
</template>

<style scoped>
.message-bubble-row {
  display: flex;
  width: 100%;
}

.message-bubble-row--notice {
  justify-content: center;
}

.message-bubble-row--self {
  justify-content: flex-end;
}

.message-bubble-row--peer {
  justify-content: flex-start;
}

.message-bubble {
  width: fit-content;
  max-width: min(76%, 34rem);
  border-radius: 0.75rem;
  box-shadow: 0 12px 28px rgb(15 23 42 / 0.08);
  padding: 0.72rem 0.9rem 0.82rem;
}

.message-bubble--self {
  background: rgb(var(--color-primary));
  color: rgb(var(--color-primary-contrast));
}

.message-bubble--peer {
  border: 1px solid rgb(var(--color-border) / 0.78);
  background: rgb(var(--color-surface-raised));
  color: rgb(var(--color-text));
}

.message-bubble__time {
  margin: 0 0 0.35rem;
  font-size: 0.7rem;
  font-weight: 700;
  line-height: 1.2;
  text-align: left;
}

.message-bubble__time--self {
  color: rgb(var(--color-primary-contrast) / 0.72);
}

.message-bubble__time--peer {
  color: rgb(var(--color-text-muted));
}

.message-bubble__body {
  margin: 0;
  overflow-wrap: anywhere;
  white-space: pre-wrap;
  font-size: 0.9rem;
  line-height: 1.65;
}

.notice-card {
  width: min(100%, 24rem);
  border: 1px solid rgb(var(--color-border) / 0.72);
  border-radius: 0.75rem;
  background: rgb(var(--color-surface));
  padding: 1rem;
  text-align: left;
  box-shadow: 0 14px 36px rgb(15 23 42 / 0.08);
}

.notice-card__title {
  margin: 0;
  color: rgb(var(--color-text));
  font-size: 1rem;
  font-weight: 800;
  line-height: 1.5;
}

.notice-card__body {
  margin: 0.45rem 0 0;
  color: rgb(var(--color-text-muted));
  font-size: 0.9rem;
  line-height: 1.65;
}

.notice-card__action {
  display: inline-flex;
  margin-top: 0.85rem;
  color: rgb(var(--color-primary));
  font-size: 0.9rem;
  font-weight: 800;
}

.notice-card__time {
  margin: 0.8rem 0 0;
  color: rgb(var(--color-text-muted));
  font-size: 0.74rem;
  text-align: right;
}

@media (max-width: 640px) {
  .message-bubble {
    max-width: 88%;
  }
}
</style>

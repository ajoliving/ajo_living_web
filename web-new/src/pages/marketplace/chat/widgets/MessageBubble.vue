<!--
 * 聊天訊息氣泡。
 * 1. 區分自己與對方的訊息樣式。
 * 2. 統一時間與氣泡留白節奏。
-->
<script setup lang="ts">
import { computed } from 'vue';

import type { ChatMessageView } from '@/model/chat';
import { usePreferenceStore } from '@/stores/preferences';

interface MessageBubbleProps {
  message: ChatMessageView;
}

const props = defineProps<MessageBubbleProps>();
const preferenceStore = usePreferenceStore();

// 1. 判斷訊息方向
const isSelf = computed(() => props.message.sender_role === 'self');

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
    :class="isSelf ? 'message-bubble-row--self' : 'message-bubble-row--peer'"
  >
    <div
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

.message-bubble-row--self {
  justify-content: flex-end;
}

.message-bubble-row--peer {
  justify-content: flex-start;
}

.message-bubble {
  width: fit-content;
  max-width: min(76%, 34rem);
  border-radius: 2px;
  padding: 0.65rem 0.8rem 0.72rem;
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
  font-size: 0.875rem;
  line-height: 1.6;
}

@media (max-width: 640px) {
  .message-bubble {
    max-width: 88%;
  }
}
</style>

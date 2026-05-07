<!--
 * 聊天訊息氣泡。
 * 1. 區分自己與對方的訊息樣式。
 * 2. 統一時間與氣泡留白節奏。
-->
<script setup lang="ts">
import { computed } from 'vue';
import { RouterLink } from 'vue-router';

import type { ChatMessageView } from '@/model/chat';
import { usePreferenceStore } from '@/stores/preferences';

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
    class="flex"
    :class="isNoticeCard ? 'justify-center' : isSelf ? 'justify-end' : 'justify-start'"
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
      class="max-w-[80%] rounded-panel px-4 py-3 shadow-soft"
      :class="
        isSelf
          ? 'bg-primary text-white'
          : 'border border-border/80 bg-surface-raised text-text'
      "
    >
      <p class="text-sm leading-6">
        {{ props.message.body }}
      </p>
      <p
        class="mt-2 text-right text-[11px]"
        :class="isSelf ? 'text-white/70' : 'text-text-muted'"
      >
        {{ formattedTime }}
      </p>
    </div>
  </div>
</template>

<style scoped>
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
</style>

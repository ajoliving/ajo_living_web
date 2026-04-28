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
    class="flex"
    :class="isSelf ? 'justify-end' : 'justify-start'"
  >
    <div
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

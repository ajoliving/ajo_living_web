<!--
 * 聊天訊息氣泡。
 * 1. 區分自己與對方的訊息樣式。
 * 2. 群聊訊息顯示發送者名稱，統一時間與氣泡留白節奏。
-->
<script setup lang="ts">
import { computed } from 'vue';

import type { ChatMessageView } from '@/model/chat';
import { usePreferenceStore } from '@/stores/preferences';
import { useI18n } from 'vue-i18n';

interface MessageBubbleProps {
  message: ChatMessageView;
  showSenderName: boolean;
}

const props = withDefaults(defineProps<MessageBubbleProps>(), {
  showSenderName: false,
});
const preferenceStore = usePreferenceStore();
const { t } = useI18n();

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
        v-if="showSenderName && !isSelf && props.message.sender_name"
        class="message-bubble__sender"
      >
        {{ props.message.sender_name }}
      </p>
      <p
        class="message-bubble__time"
        :class="isSelf ? 'message-bubble__time--self' : 'message-bubble__time--peer'"
      >
        {{ formattedTime }}
      </p>
      <p class="message-bubble__body">
        {{ props.message.body }}
      </p>
      <div v-if="props.message.attachments.length" class="message-bubble__attachments">
        <template v-for="attachment in props.message.attachments" :key="attachment.media_asset_id">
          <template v-if="attachment.processing_status === 'rejected' || attachment.scan_status === 'rejected'">
            <span class="message-bubble__file">{{ t('chat.attachmentRejected') }}</span>
          </template>
          <template v-else-if="attachment.processing_status === 'processing' || attachment.scan_status === 'pending'">
            <span class="message-bubble__file">{{ t('chat.attachmentProcessing') }}</span>
          </template>
          <img v-else-if="attachment.mime_type.startsWith('image/')" :src="attachment.url" class="message-bubble__image" loading="lazy" />
          <video v-else-if="attachment.mime_type.startsWith('video/')" :src="attachment.url" class="message-bubble__video" controls preload="metadata" />
          <a v-else :href="attachment.url" target="_blank" rel="noreferrer" class="message-bubble__file">{{ attachment.mime_type }}</a>
        </template>
      </div>
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

.message-bubble__sender {
  margin: 0 0 0.35rem;
  overflow: hidden;
  color: rgb(var(--color-primary));
  font-size: 0.72rem;
  font-weight: 700;
  line-height: 1.2;
  text-overflow: ellipsis;
  white-space: nowrap;
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

.message-bubble__attachments { display: grid; gap: 0.45rem; margin-top: 0.55rem; }
.message-bubble__image, .message-bubble__video { display: block; max-width: 18rem; max-height: 14rem; border-radius: 2px; object-fit: cover; }
.message-bubble__file { color: inherit; text-decoration: underline; font-size: 0.8rem; }

@media (max-width: 640px) {
  .message-bubble {
    max-width: 88%;
  }
}
</style>

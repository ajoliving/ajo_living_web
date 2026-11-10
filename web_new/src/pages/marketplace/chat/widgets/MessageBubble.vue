<!--
 * 聊天訊息氣泡。
 * 1. 區分自己與對方的訊息樣式。
 * 2. 群聊訊息顯示發送者名稱；已撤回訊息只顯示佔位文字。
 * 3. 附件與時間收在氣泡內容區，空白文字不佔位。
 * 4. 右鍵或長按開啟回覆、複製與撤回選單。
-->
<script setup lang="ts">
import { computed, onBeforeUnmount, ref } from 'vue';

import type { ChatMessageView } from '@/model/chat';
import { usePreferenceStore } from '@/stores/preferences';
import { useI18n } from 'vue-i18n';

interface MessageBubbleProps {
  message: ChatMessageView;
  showSenderName: boolean;
  canRecall: boolean;
  recalling: boolean;
}

const props = withDefaults(defineProps<MessageBubbleProps>(), {
  showSenderName: false,
  canRecall: false,
  recalling: false,
});
const emit = defineEmits<{
  (event: 'reply'): void;
  (event: 'recall'): void;
  (event: 'quote-click', messageId: string): void;
}>();
const preferenceStore = usePreferenceStore();
const { t } = useI18n();
const menuOpen = ref(false);
const menuX = ref(0);
const menuY = ref(0);
const longPressTimer = ref<number | null>(null);
const suppressClick = ref(false);

// 1. 判斷訊息方向
const isSelf = computed(() => props.message.sender_role === 'self');

// 2. 判斷是否已撤回
const isRecalled = computed(() => props.message.status === 'recalled');

// 3. 格式化訊息時間
const formattedTime = computed(() =>
  new Intl.DateTimeFormat(preferenceStore.locale, {
    hour: '2-digit',
    minute: '2-digit',
  }).format(new Date(props.message.sent_at)),
);

// 4. 附件檔案標籤
const fileLabel = (mimeType: string): string => {
  if (mimeType === 'application/pdf') {
    return 'PDF';
  }
  if (mimeType === 'text/plain') {
    return 'TXT';
  }
  return mimeType;
};

// 5. 複製用文字
const copyText = computed(() => {
  const body = props.message.body.trim();
  if (isRecalled.value || !body) {
    return isRecalled.value ? t('chat.recalledMessage') : t('chat.attachmentMessagePreview');
  }
  return body;
});

// 6. 關閉操作選單
const closeMenu = (): void => {
  menuOpen.value = false;
};

// 7. 在點擊位置開啟操作選單
const openMenuAt = (clientX: number, clientY: number): void => {
  if (isRecalled.value) {
    return;
  }
  const menuWidth = 160;
  const menuHeight = props.canRecall ? 148 : 88;
  menuX.value = Math.min(Math.max(8, clientX), window.innerWidth - menuWidth - 8);
  menuY.value = Math.min(Math.max(8, clientY), window.innerHeight - menuHeight - 8);
  menuOpen.value = true;
};

// 8. 桌面右鍵開啟選單
const handleContextMenu = (event: MouseEvent): void => {
  if (isRecalled.value) {
    return;
  }
  event.preventDefault();
  openMenuAt(event.clientX, event.clientY);
};

// 9. 取消尚未觸發的長按
const clearLongPress = (): void => {
  if (longPressTimer.value !== null) {
    window.clearTimeout(longPressTimer.value);
    longPressTimer.value = null;
  }
};

// 10. 觸控長按開啟選單
const handleTouchStart = (event: TouchEvent): void => {
  if (isRecalled.value || event.touches.length !== 1) {
    return;
  }
  const touch = event.touches[0];
  clearLongPress();
  longPressTimer.value = window.setTimeout(() => {
    suppressClick.value = true;
    openMenuAt(touch.clientX, touch.clientY);
    longPressTimer.value = null;
  }, 500);
};

// 11. 觸控移動或結束時取消長按
const handleTouchCancel = (): void => {
  clearLongPress();
};

// 12. 回覆目前訊息
const handleReply = (): void => {
  closeMenu();
  emit('reply');
};

// 13. 複製訊息文字
const handleCopy = async (): Promise<void> => {
  closeMenu();
  try {
    await navigator.clipboard.writeText(copyText.value);
  } catch {
    return;
  }
};

// 14. 撤回目前訊息
const handleRecall = (): void => {
  if (!props.canRecall || props.recalling) {
    return;
  }
  closeMenu();
  emit('recall');
};

// 15. 點擊引用預覽
const handleQuoteClick = (): void => {
  const reply = props.message.reply;
  if (!reply?.message_id) {
    return;
  }
  emit('quote-click', reply.message_id);
};

// 16. 長按後忽略下一次點擊，避免誤開附件
const handleRowClick = (event: MouseEvent): void => {
  if (!suppressClick.value) {
    return;
  }
  event.preventDefault();
  event.stopPropagation();
  suppressClick.value = false;
};

onBeforeUnmount(() => {
  clearLongPress();
});
</script>

<template>
  <div
    class="message-bubble-row"
    :class="isSelf ? 'message-bubble-row--self' : 'message-bubble-row--peer'"
    :data-message-id="props.message.id"
    @contextmenu="handleContextMenu"
    @touchstart.passive="handleTouchStart"
    @touchend="handleTouchCancel"
    @touchmove="handleTouchCancel"
    @touchcancel="handleTouchCancel"
    @click.capture="handleRowClick"
  >
    <div
      class="message-bubble"
      :class="[
        isSelf ? 'message-bubble--self' : 'message-bubble--peer',
        isRecalled ? 'message-bubble--recalled' : '',
      ]"
    >
      <p
        v-if="showSenderName && !isSelf && props.message.sender_name"
        class="message-bubble__sender"
      >
        {{ props.message.sender_name }}
      </p>
      <button
        v-if="props.message.reply"
        type="button"
        class="message-bubble__quote"
        @click.stop="handleQuoteClick"
      >
        <span class="message-bubble__quote-name">{{ props.message.reply.sender_name }}</span>
        <span class="message-bubble__quote-text">{{ props.message.reply.content }}</span>
      </button>
      <p
        v-if="isRecalled || props.message.body"
        class="message-bubble__body"
      >
        {{ isRecalled ? t('chat.recalledMessage') : props.message.body }}
      </p>
      <div
        v-if="!isRecalled && props.message.attachments.length"
        class="message-bubble__attachments"
      >
        <template v-for="attachment in props.message.attachments" :key="attachment.media_asset_id">
          <template v-if="attachment.processing_status === 'rejected' || attachment.scan_status === 'rejected'">
            <span class="message-bubble__file">{{ t('chat.attachmentRejected') }}</span>
          </template>
          <template v-else-if="attachment.processing_status === 'processing' || attachment.scan_status === 'pending'">
            <span class="message-bubble__file">{{ t('chat.attachmentProcessing') }}</span>
          </template>
          <img
            v-else-if="attachment.mime_type.startsWith('image/')"
            :src="attachment.url"
            class="message-bubble__image"
            loading="lazy"
          >
          <video
            v-else-if="attachment.mime_type.startsWith('video/')"
            :src="attachment.url"
            class="message-bubble__video"
            controls
            preload="metadata"
          />
          <a
            v-else
            :href="attachment.url"
            target="_blank"
            rel="noreferrer"
            class="message-bubble__file"
          >{{ fileLabel(attachment.mime_type) }}</a>
        </template>
      </div>
      <div class="message-bubble__meta">
        <span
          class="message-bubble__time"
          :class="isSelf ? 'message-bubble__time--self' : 'message-bubble__time--peer'"
        >
          {{ formattedTime }}
        </span>
      </div>
    </div>
  </div>

  <Teleport to="body">
    <div
      v-if="menuOpen"
      class="message-menu-backdrop"
      @click="closeMenu"
      @contextmenu.prevent="closeMenu"
    >
      <div
        class="message-menu"
        :style="{ left: `${menuX}px`, top: `${menuY}px` }"
        @click.stop
      >
        <button
          type="button"
          class="message-menu__item"
          @click="handleReply"
        >
          {{ t('chat.reply') }}
        </button>
        <button
          type="button"
          class="message-menu__item"
          @click="handleCopy"
        >
          {{ t('chat.copy') }}
        </button>
        <button
          v-if="canRecall"
          type="button"
          class="message-menu__item"
          :disabled="recalling"
          @click="handleRecall"
        >
          {{ t('chat.recall') }}
        </button>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.message-bubble-row {
  display: flex;
  width: 100%;
  -webkit-touch-callout: none;
  -webkit-user-select: none;
  user-select: none;
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
  border-radius: 8px;
  padding: 0.7rem 0.85rem 0.6rem;
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

.message-bubble--recalled {
  background: rgb(var(--color-surface-raised));
  color: rgb(var(--color-text-muted));
  border: 1px solid rgb(var(--color-border) / 0.78);
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

.message-bubble__quote {
  display: block;
  width: 100%;
  margin: 0 0 0.45rem;
  border: 0;
  border-left: 3px solid currentColor;
  border-radius: 4px;
  background: rgb(255 255 255 / 0.16);
  color: inherit;
  cursor: pointer;
  font: inherit;
  padding: 6px 8px;
  text-align: left;
}

.message-bubble--peer .message-bubble__quote {
  background: rgb(var(--color-surface-muted));
  border-left-color: rgb(var(--color-primary));
}

.message-bubble__quote-name {
  display: block;
  font-size: 0.7rem;
  font-weight: 700;
  opacity: 0.9;
}

.message-bubble__quote-text {
  display: block;
  overflow: hidden;
  font-size: 0.78rem;
  line-height: 1.4;
  opacity: 0.82;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.message-bubble__meta {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 0.4rem;
}

.message-bubble__time {
  margin: 0;
  font-size: 0.7rem;
  font-weight: 600;
  line-height: 1.2;
}

.message-bubble__time--self {
  color: rgb(var(--color-primary-contrast) / 0.72);
}

.message-bubble--recalled .message-bubble__time--self,
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

.message-bubble--recalled .message-bubble__body {
  font-style: italic;
}

.message-bubble__attachments {
  display: grid;
  gap: 0.45rem;
  margin-top: 0.45rem;
}

.message-bubble__body + .message-bubble__attachments {
  margin-top: 0.5rem;
}

.message-bubble__image,
.message-bubble__video {
  display: block;
  max-width: 18rem;
  max-height: 14rem;
  border-radius: 6px;
  object-fit: cover;
}

.message-bubble__file {
  display: inline-flex;
  align-items: center;
  min-height: 32px;
  border: 1px solid rgb(255 255 255 / 0.22);
  border-radius: 6px;
  padding: 0 10px;
  color: inherit;
  text-decoration: none;
  font-size: 0.8rem;
  font-weight: 700;
}

.message-bubble--peer .message-bubble__file {
  border-color: rgb(var(--color-border));
  background: rgb(var(--color-surface-muted));
}

@media (max-width: 640px) {
  .message-bubble {
    max-width: 88%;
  }
}
</style>

<style>
.message-menu-backdrop {
  position: fixed;
  inset: 0;
  z-index: 80;
}

.message-menu {
  position: fixed;
  z-index: 81;
  min-width: 148px;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface-raised));
  box-shadow: 0 8px 24px rgb(15 23 42 / 0.12);
  padding: 4px;
}

.message-menu__item {
  display: block;
  width: 100%;
  min-height: 36px;
  border: 0;
  border-radius: 6px;
  background: transparent;
  color: rgb(var(--color-text));
  cursor: pointer;
  font: inherit;
  font-size: 0.875rem;
  padding: 0 12px;
  text-align: left;
}

.message-menu__item:hover,
.message-menu__item:focus-visible {
  background: rgb(var(--color-primary) / 0.08);
  outline: none;
}

.message-menu__item:disabled {
  cursor: not-allowed;
  opacity: 0.45;
}
</style>

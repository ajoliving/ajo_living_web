<!--
 * 聊天表情選擇器。
 * 1. 顯示常用表情列表。
 * 2. 觸發表情插入與面板開關。
-->
<script setup lang="ts">
import AppIcon from '@/shared/components/base/AppIcon.vue';

import type { ChatEmojiOption } from '../chat';

defineProps<{
  emojis: ChatEmojiOption[];
  open: boolean;
  t: (key: string, params?: Record<string, unknown>) => string;
}>();

const emit = defineEmits<{
  select: [emoji: string];
  toggle: [];
}>();
</script>

<template>
  <div class="chat-emoji-picker">
    <button
      type="button"
      class="chat-emoji-toggle"
      :aria-expanded="open"
      :aria-label="t('chat.emojiToggle')"
      @click="emit('toggle')"
    >
      <AppIcon
        name="smile"
        :size="18"
      />
    </button>

    <div
      v-if="open"
      class="chat-emoji-panel"
      role="listbox"
      :aria-label="t('chat.emojiPanelLabel')"
    >
      <button
        v-for="emoji in emojis"
        :key="emoji.key"
        type="button"
        class="chat-emoji-option"
        @click="emit('select', emoji.value)"
      >
        {{ emoji.value }}
      </button>
    </div>
  </div>
</template>

<style scoped lang="scss">
.chat-emoji-picker {
  position: relative;
}

.chat-emoji-toggle {
  display: inline-flex;
  width: 2.75rem;
  height: 2.75rem;
  align-items: center;
  justify-content: center;
  border: 1px solid rgb(var(--color-border) / 0.8);
  border-radius: 999px;
  background: rgb(var(--color-surface-raised));
  color: rgb(var(--color-text));
  transition:
    border-color 0.2s ease,
    background 0.2s ease;
}

.chat-emoji-toggle:hover {
  border-color: rgb(var(--color-primary) / 0.4);
  background: rgb(var(--color-primary) / 0.08);
}

.chat-emoji-panel {
  position: absolute;
  right: 0;
  bottom: calc(100% + 0.75rem);
  z-index: 20;
  display: grid;
  width: min(18rem, calc(100vw - 2rem));
  grid-template-columns: repeat(8, minmax(0, 1fr));
  gap: 0.35rem;
  border: 1px solid rgb(var(--color-border) / 0.9);
  border-radius: 1rem;
  background: rgb(var(--color-surface));
  box-shadow: 0 18px 48px rgb(15 23 42 / 0.16);
  padding: 0.65rem;
}

.chat-emoji-option {
  display: inline-flex;
  aspect-ratio: 1;
  align-items: center;
  justify-content: center;
  border: 0;
  border-radius: 0.75rem;
  background: transparent;
  font-size: 1.25rem;
  line-height: 1;
  transition: background 0.2s ease;
}

.chat-emoji-option:hover {
  background: rgb(var(--color-primary) / 0.1);
}

@media (max-width: 640px) {
  .chat-emoji-panel {
    right: auto;
    left: 0;
    grid-template-columns: repeat(6, minmax(0, 1fr));
  }
}
</style>

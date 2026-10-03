<!--
 * 聊天表情選擇器。
 * 1. 提供一組常用 Unicode 表情，插入目前輸入草稿。
 * 2. 不引入貼圖包或額外資源。
-->
<script setup lang="ts">
const emit = defineEmits<{
  (event: 'select', emoji: string): void;
}>();

const emojiCodePoints: number[] = [
  0x1f600, 0x1f603, 0x1f604, 0x1f601, 0x1f605, 0x1f602, 0x1f923, 0x1f60a,
  0x1f607, 0x1f642, 0x1f609, 0x1f60d, 0x1f618, 0x1f61c, 0x1f92a, 0x1f914,
  0x1f644, 0x1f622, 0x1f62d, 0x1f621, 0x1f44d, 0x1f44e, 0x1f44f, 0x1f64f,
  0x1f91d, 0x1f44b, 0x2764, 0x1f525, 0x1f389, 0x1f4aa, 0x1f3c6, 0x1f3e0,
  0x1f697, 0x2600, 0x1f319, 0x2615, 0x1f37a, 0x1f354, 0x1f382, 0x1f4ac,
];

// 1. 將碼點轉成可插入的字元
const emojiList = emojiCodePoints.map((codePoint) => String.fromCodePoint(codePoint));

// 2. 選擇一個表情
const handleSelect = (emoji: string): void => {
  emit('select', emoji);
};
</script>

<template>
  <div class="emoji-picker">
    <button
      v-for="emoji in emojiList"
      :key="emoji"
      type="button"
      class="emoji-picker__item"
      @click="handleSelect(emoji)"
    >
      {{ emoji }}
    </button>
  </div>
</template>

<style scoped>
.emoji-picker {
  display: grid;
  grid-template-columns: repeat(8, minmax(0, 1fr));
  gap: 4px;
  width: min(100%, 20rem);
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
  padding: 8px;
  box-shadow: 0 8px 24px rgb(0 0 0 / 0.12);
}

.emoji-picker__item {
  display: inline-flex;
  width: 100%;
  aspect-ratio: 1;
  align-items: center;
  justify-content: center;
  border: 0;
  border-radius: 6px;
  background: transparent;
  cursor: pointer;
  font-size: 1.15rem;
  line-height: 1;
}

.emoji-picker__item:hover,
.emoji-picker__item:focus-visible {
  background: rgb(var(--color-primary) / 0.08);
  outline: none;
}
</style>

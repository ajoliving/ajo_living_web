<!--
 * 篩選標籤元件。
 * 1. 單個篩選標籤，支援選中態（.on）。
 * 2. 點擊觸發 toggle 事件，由父元件管理選中邏輯。
 * 3. 樣式還原 HTML 設計稿 .ft 規則。
-->
<script setup lang="ts">
interface FilterTagProps {
  label: string;
  active?: boolean;
}

const props = withDefaults(defineProps<FilterTagProps>(), {
  active: false,
});

const emit = defineEmits<{
  (event: 'toggle'): void;
}>();

// 1. 點擊觸發 toggle 事件
const handleToggle = () => {
  emit('toggle');
};
</script>

<template>
  <button
    type="button"
    class="ft"
    :class="props.active ? 'on' : ''"
    @click="handleToggle"
  >
    {{ props.label }}
  </button>
</template>

<style scoped>
.ft {
  font-size: 11px;
  padding: 4px 9px;
  border: 1px solid var(--color-border);
  color: var(--color-ink-3);
  cursor: pointer;
  background: var(--color-surface);
  border-radius: 2px;
  font-family: inherit;
}

.ft:hover {
  border-color: var(--color-primary);
  color: var(--color-brand-dark);
}

.ft.on {
  background: var(--color-primary);
  color: var(--color-surface);
  border-color: var(--color-primary);
}
</style>

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
  min-height: 28px;
  border: 1px solid var(--bdr);
  border-radius: var(--r-pill);
  background: var(--sur-2);
  padding: 4px 11px;
  color: var(--ink-3);
  cursor: pointer;
  font-family: inherit;
  line-height: 1;
  transition:
    border-color 0.15s ease,
    background 0.15s ease,
    color 0.15s ease;
}

.ft:hover {
  border-color: var(--brand);
  color: var(--brand-dark);
}

.ft.on {
  border-color: var(--brand);
  background: var(--brand);
  color: #fff;
}
</style>

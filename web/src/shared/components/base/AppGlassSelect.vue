<!--
 * 玻璃質感下拉選擇器。
 * 1. 提供可控位置的自繪單選彈層。
 * 2. 統一磨砂玻璃質感、選中狀態、禁用狀態與外部關閉行為。
-->
<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';

import AppIcon from '@/shared/components/base/AppIcon.vue';

interface GlassSelectOption {
  label: string;
  value: string;
}

interface AppGlassSelectProps {
  modelValue: string;
  options: GlassSelectOption[];
  panelMaxHeight?: string;
  panelMinWidth?: string;
  disabled?: boolean;
}

const props = withDefaults(defineProps<AppGlassSelectProps>(), {
  panelMaxHeight: '14rem',
  panelMinWidth: '100%',
  disabled: false,
});

const emit = defineEmits<{
  (event: 'update:modelValue', value: string): void;
  (event: 'change', value: string): void;
}>();

const rootRef = ref<HTMLElement | null>(null);
const isOpen = ref(false);

const currentOption = computed(
  () => props.options.find((option) => option.value === props.modelValue) ?? props.options[0],
);

const panelStyle = computed(() => ({
  maxHeight: props.panelMaxHeight,
  minWidth: props.panelMinWidth,
}));

// 1. 切換下拉開合狀態
const toggleOpen = (): void => {
  if (props.disabled) {
    return;
  }

  isOpen.value = !isOpen.value;
};

// 2. 選擇選項並同步給父層
const handleSelect = (value: string): void => {
  if (props.disabled) {
    return;
  }

  emit('update:modelValue', value);
  emit('change', value);
  isOpen.value = false;
};

// 3. 點擊外部時關閉下拉
const handlePointerDown = (event: PointerEvent): void => {
  const target = event.target;

  if (!(target instanceof Node)) {
    return;
  }

  if (!rootRef.value?.contains(target)) {
    isOpen.value = false;
  }
};

// 4. 使用 Esc 關閉下拉
const handleKeydown = (event: KeyboardEvent): void => {
  if (event.key === 'Escape') {
    isOpen.value = false;
  }
};

onMounted(() => {
  window.addEventListener('pointerdown', handlePointerDown);
  window.addEventListener('keydown', handleKeydown);
});

onBeforeUnmount(() => {
  window.removeEventListener('pointerdown', handlePointerDown);
  window.removeEventListener('keydown', handleKeydown);
});
</script>

<template>
  <div
    ref="rootRef"
    class="app-glass-select"
  >
    <button
      type="button"
      class="app-glass-select__trigger"
      aria-haspopup="listbox"
      :aria-expanded="isOpen"
      :disabled="props.disabled"
      @click="toggleOpen"
    >
      <span>{{ currentOption?.label }}</span>
      <AppIcon
        name="chevron-down"
        :size="16"
      />
    </button>

    <Transition name="app-glass-select">
      <div
        v-if="isOpen"
        class="app-glass-select__menu"
        role="listbox"
        :style="panelStyle"
      >
        <button
          v-for="option in props.options"
          :key="String(option.value)"
          type="button"
          class="app-glass-select__option"
          :class="{ 'app-glass-select__option--active': option.value === props.modelValue }"
          role="option"
          :aria-selected="option.value === props.modelValue"
          @click="handleSelect(option.value)"
        >
          {{ option.label }}
        </button>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.app-glass-select {
  position: relative;
  width: 100%;
  min-width: 0;
}

.app-glass-select__trigger {
  display: flex;
  width: 100%;
  min-width: 0;
  min-height: 3rem;
  cursor: pointer;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.5rem;
  background: rgb(var(--color-surface));
  padding: 0 1rem;
  color: rgb(var(--color-text));
  font-size: 0.9375rem;
  line-height: 1.6;
  outline: none;
  transition:
    background-color 0.2s ease,
    border-color 0.2s ease,
    box-shadow 0.2s ease;
}

.app-glass-select__trigger:hover,
.app-glass-select__trigger[aria-expanded='true'] {
  border-color: rgb(var(--color-primary));
  box-shadow: 0 0 0 3px rgb(var(--color-primary) / 0.16);
}

.app-glass-select__trigger:disabled {
  cursor: not-allowed;
  border-color: rgb(var(--color-border) / 0.72);
  background: color-mix(in srgb, rgb(var(--color-surface)) 82%, rgb(var(--color-page-tint)) 18%);
  color: rgb(var(--color-text-muted));
  box-shadow: none;
}

.app-glass-select__trigger span {
  min-width: 0;
  overflow: hidden;
  color: inherit;
  font-size: inherit;
  font-weight: 600;
  letter-spacing: 0;
  line-height: inherit;
  text-overflow: ellipsis;
  text-transform: none;
  white-space: nowrap;
}

.app-glass-select__trigger svg {
  flex-shrink: 0;
}

.app-glass-select__menu {
  position: absolute;
  z-index: 40;
  top: 100%;
  right: 0;
  left: 0;
  display: grid;
  gap: 0;
  overflow-y: auto;
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.25rem;
  background: rgb(var(--color-surface));
  padding: 0.25rem 0;
  box-shadow: 0 8px 18px rgb(15 23 42 / 0.12);
  scrollbar-color: rgb(var(--color-border)) transparent;
  scrollbar-width: thin;
}

.app-glass-select__menu::-webkit-scrollbar {
  width: 0.45rem;
}

.app-glass-select__menu::-webkit-scrollbar-thumb {
  border-radius: 9999px;
  background: rgb(var(--color-border));
}

.app-glass-select__option {
  display: flex;
  width: 100%;
  cursor: pointer;
  align-items: center;
  border: 0;
  border-radius: 0;
  background: transparent;
  padding: 0.48rem 0.7rem;
  color: rgb(var(--color-text));
  font-size: 0.875rem;
  font-weight: 500;
  line-height: 1.3;
  text-align: left;
  transition:
    background-color 0.2s ease,
    color 0.2s ease;
}

.app-glass-select__option:hover {
  background: rgb(var(--color-surface-muted));
}

.app-glass-select__option--active {
  background: rgb(var(--color-primary-soft) / 0.32);
  color: rgb(var(--color-text));
  font-weight: 700;
}

.app-glass-select-enter-active,
.app-glass-select-leave-active {
  transition:
    opacity 0.18s ease,
    transform 0.18s ease;
}

.app-glass-select-enter-from,
.app-glass-select-leave-to {
  opacity: 0;
  transform: translateY(-2px);
}
</style>

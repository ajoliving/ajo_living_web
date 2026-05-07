<!--
 * 玻璃質感下拉選擇器。
 * 1. 提供可控位置的自繪單選彈層。
 * 2. 統一磨砂玻璃質感、選中狀態與外部關閉行為。
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
}

const props = withDefaults(defineProps<AppGlassSelectProps>(), {
  panelMaxHeight: '14rem',
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
}));

// 1. 切換下拉開合狀態
const toggleOpen = (): void => {
  isOpen.value = !isOpen.value;
};

// 2. 選擇選項並同步給父層
const handleSelect = (value: string): void => {
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
  border-radius: 8px;
  background: #ffffff;
  padding: 0 0.875rem;
  color: rgb(var(--color-text));
  font-size: 0.9375rem;
  line-height: 1.4;
  outline: none;
  transition:
    border-color 0.2s ease,
    box-shadow 0.2s ease,
    background 0.2s ease;
}

.app-glass-select__trigger:hover,
.app-glass-select__trigger[aria-expanded='true'] {
  border-color: rgb(var(--color-primary));
  box-shadow: 0 0 0 3px rgb(var(--color-primary) / 0.16);
}

.app-glass-select__trigger span {
  min-width: 0;
  overflow: hidden;
  color: inherit;
  font-size: inherit;
  font-weight: 650;
  line-height: inherit;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.app-glass-select__trigger svg {
  flex-shrink: 0;
}

.app-glass-select__menu {
  position: absolute;
  z-index: 40;
  top: calc(100% - 1px);
  right: 0;
  left: 0;
  display: grid;
  gap: 0.25rem;
  overflow-y: auto;
  border: 1px solid rgb(var(--color-border) / 0.62);
  border-radius: 0 0 8px 8px;
  background:
    linear-gradient(135deg, rgb(255 255 255 / 0.34), rgb(255 255 255 / 0.08) 46%, transparent 100%),
    linear-gradient(
      180deg,
      rgb(var(--color-dropdown-surface) / 0.52),
      rgb(var(--color-surface) / 0.28)
    );
  padding: 0.45rem;
  box-shadow:
    0 20px 48px rgb(15 23 42 / 0.18),
    inset 0 1px 0 rgb(255 255 255 / 0.42),
    inset 0 -1px 0 rgb(255 255 255 / 0.16);
  backdrop-filter: blur(34px) saturate(190%);
  -webkit-backdrop-filter: blur(34px) saturate(190%);
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
  border-radius: 6px;
  background: transparent;
  padding: 0.65rem 0.7rem;
  color: rgb(var(--color-text));
  font-size: 0.875rem;
  font-weight: 700;
  line-height: 1.3;
  text-align: left;
  transition:
    background-color 0.2s ease,
    color 0.2s ease;
}

.app-glass-select__option:hover {
  background: rgb(255 255 255 / 0.28);
}

.app-glass-select__option--active {
  background: rgb(var(--color-primary-soft) / 0.42);
  color: rgb(var(--color-primary));
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
  transform: translateY(-4px);
}
</style>

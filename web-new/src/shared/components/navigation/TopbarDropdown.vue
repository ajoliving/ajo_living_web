<!--
 * 頂部導覽文字下拉元件。
 * 1. 提供文字按鈕式觸發器。
 * 2. 提供可主題化的展開面板與選項切換。
 * 3. 處理點擊外部與 Esc 關閉行為。
-->
<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';

interface DropdownOption {
  label: string;
  value: string | number;
}

interface TopbarDropdownProps {
  modelValue: string | number;
  options: DropdownOption[];
  align?: 'left' | 'right';
  panelWidth?: string;
}

const props = withDefaults(defineProps<TopbarDropdownProps>(), {
  align: 'right',
  panelWidth: '9rem',
});

const emit = defineEmits<{
  (event: 'update:modelValue', value: string | number): void;
}>();

const rootRef = ref<HTMLElement | null>(null);
const isOpen = ref(false);

const currentOption = computed(
  () => props.options.find((item) => item.value === props.modelValue) ?? props.options[0],
);

const panelPositionClassName = computed(() =>
  props.align === 'left' ? 'left-0 origin-top-left' : 'right-0 origin-top-right',
);

const panelStyle = computed(() => ({
  minWidth: props.panelWidth,
}));

// 1. 切換展開狀態
const toggleOpen = () => {
  isOpen.value = !isOpen.value;
};

// 2. 點擊選項後同步值並關閉面板
const handleSelect = (value: string | number) => {
  emit('update:modelValue', value);
  isOpen.value = false;
};

// 3. 點擊外部時自動關閉
const handlePointerDown = (event: PointerEvent) => {
  const target = event.target;

  if (!(target instanceof Node)) {
    return;
  }

  if (!rootRef.value?.contains(target)) {
    isOpen.value = false;
  }
};

// 4. 使用 Esc 關閉面板
const handleKeydown = (event: KeyboardEvent) => {
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
    class="relative"
  >
    <button
      type="button"
      class="topbar-dropdown-trigger"
      :class="{ 'topbar-dropdown-trigger-open': isOpen }"
      @click="toggleOpen"
    >
      {{ currentOption?.label }}
    </button>

    <Transition name="topbar-dropdown">
      <div
        v-if="isOpen"
        class="topbar-dropdown-panel"
        :class="panelPositionClassName"
        :style="panelStyle"
      >
        <button
          v-for="option in props.options"
          :key="String(option.value)"
          type="button"
          class="topbar-dropdown-option"
          :class="{ 'topbar-dropdown-option-active': option.value === props.modelValue }"
          @click="handleSelect(option.value)"
        >
          {{ option.label }}
        </button>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.topbar-dropdown-trigger {
  border: none;
  background: transparent;
  font-family: var(--font-display);
  color: rgb(var(--color-text) / 0.98);
  padding: 0.35rem 0;
  font-size: 0.875rem;
  font-weight: 600;
  line-height: 1.2;
  white-space: nowrap;
  cursor: pointer;
  transition:
    color 0.2s ease,
    transform 0.2s ease;
}

.topbar-dropdown-trigger:hover,
.topbar-dropdown-trigger-open {
  color: rgb(var(--color-text));
}

.topbar-dropdown-trigger-open {
  transform: translateY(-1px);
}

.topbar-dropdown-panel {
  position: absolute;
  top: calc(100% + 0.55rem);
  z-index: 80;
  display: flex;
  flex-direction: column;
  gap: 0.12rem;
  border: var(--border-subtle);
  border-radius: 3px;
  background: rgb(var(--color-dropdown-surface));
  padding: 0.35rem;
  box-shadow: var(--shadow-floating);
  overflow: hidden;
}

.topbar-dropdown-option {
  border: none;
  border-radius: var(--radius-md);
  background: transparent;
  font-family: var(--font-display);
  color: rgb(var(--color-text-muted));
  padding: 0.62rem 0.8rem;
  text-align: left;
  font-size: 0.8125rem;
  font-weight: 600;
  line-height: 1.2;
  cursor: pointer;
  transition:
    background-color 0.2s ease,
    color 0.2s ease,
    transform 0.2s ease;
}

.topbar-dropdown-option:hover,
.topbar-dropdown-option-active {
  background: rgb(var(--color-toolbar-surface) / 0.9);
  color: rgb(var(--color-text));
}

.topbar-dropdown-option-active {
  background: rgb(var(--color-primary-soft) / 0.4);
  color: rgb(var(--color-primary));
}

.topbar-dropdown-option:hover {
  transform: translateY(-1px);
}

.topbar-dropdown-enter-active,
.topbar-dropdown-leave-active {
  transition:
    opacity 0.2s ease,
    transform 0.2s ease;
}

.topbar-dropdown-enter-from,
.topbar-dropdown-leave-to {
  opacity: 0;
  transform: translateY(-6px) scale(0.985);
}

</style>

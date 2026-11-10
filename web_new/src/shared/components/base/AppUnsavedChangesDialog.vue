<!--
 * 未儲存離開確認彈窗。
 * 1. 提供通用的保存、放棄與留在頁面三段式確認。
 * 2. 使用穩定面板承載跨頁面離開提示。
-->
<script setup lang="ts">
import { useDialogBackdropClose } from '@/shared/composables/useDialogBackdropClose';

interface AppUnsavedChangesDialogProps {
  open: boolean;
  title: string;
  description: string;
  saveLabel: string;
  discardLabel: string;
  stayLabel: string;
  saving?: boolean;
}

const props = withDefaults(defineProps<AppUnsavedChangesDialogProps>(), {
  saving: false,
});

const emit = defineEmits<{
  (event: 'save'): void;
  (event: 'discard'): void;
  (event: 'stay'): void;
}>();

// 1. 只在完整點擊遮罩時留在編輯頁
const {
  handleBackdropPointerCancel,
  handleBackdropPointerDown,
  handleBackdropPointerUp,
} = useDialogBackdropClose(() => emit('stay'));
</script>

<template>
  <Teleport to="body">
    <Transition name="unsaved-dialog">
      <div
        v-if="props.open"
        class="unsaved-dialog"
        role="dialog"
        aria-modal="true"
        :aria-label="props.title"
        @pointercancel="handleBackdropPointerCancel"
        @pointerdown="handleBackdropPointerDown"
        @pointerup="handleBackdropPointerUp"
      >
        <div class="unsaved-dialog__panel">
          <div class="unsaved-dialog__header">
            <span class="unsaved-dialog__mark" />
            <div>
              <h2>{{ props.title }}</h2>
              <p>{{ props.description }}</p>
            </div>
          </div>

          <div class="unsaved-dialog__actions">
            <button
              type="button"
              class="unsaved-dialog__button unsaved-dialog__button--secondary"
              @click="emit('stay')"
            >
              {{ props.stayLabel }}
            </button>
            <button
              type="button"
              class="unsaved-dialog__button unsaved-dialog__button--secondary"
              @click="emit('discard')"
            >
              {{ props.discardLabel }}
            </button>
            <button
              type="button"
              class="unsaved-dialog__button unsaved-dialog__button--primary"
              :disabled="props.saving"
              @click="emit('save')"
            >
              {{ props.saveLabel }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.unsaved-dialog {
  position: fixed;
  z-index: 120;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgb(15 23 42 / 0.22);
  padding: 1rem;
}

.unsaved-dialog__panel {
  width: min(100%, 30rem);
  border: 1px solid rgb(var(--color-border));
  border-radius: 3px;
  background: rgb(var(--color-surface));
  padding: 1.25rem;
  box-shadow: 0 18px 44px rgb(15 23 42 / 0.16);
  color: rgb(var(--color-text));
}

.unsaved-dialog__header {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr);
  gap: 0.9rem;
  align-items: start;
}

.unsaved-dialog__mark {
  display: inline-flex;
  width: 0.65rem;
  height: 2.25rem;
  border-radius: 999px;
  background: rgb(var(--color-primary-soft) / 0.42);
  box-shadow:
    inset 0 1px 0 rgb(255 255 255 / 0.42),
    0 0 20px rgb(var(--color-primary) / 0.16);
}

.unsaved-dialog__panel h2 {
  margin: 0;
  color: rgb(var(--color-text));
  font-family: var(--font-display);
  font-size: 1.15rem;
  font-weight: 700;
  line-height: 1.35;
}

.unsaved-dialog__panel p {
  margin: 0.65rem 0 0;
  color: rgb(var(--color-text-muted));
  font-size: 0.92rem;
  line-height: 1.7;
}

.unsaved-dialog__actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 0.65rem;
  margin-top: 1.25rem;
}

.unsaved-dialog__button {
  display: inline-flex;
  min-height: 2.65rem;
  cursor: pointer;
  align-items: center;
  justify-content: center;
  border: 0;
  border-radius: 6px;
  background: transparent;
  padding: 0 0.95rem;
  color: rgb(var(--color-text));
  font-size: 0.875rem;
  font-weight: 700;
  line-height: 1.2;
  transition:
    background-color 0.2s ease,
    color 0.2s ease,
    transform 0.2s ease;
}

.unsaved-dialog__button--secondary {
  background: rgb(255 255 255 / 0.34);
}

.unsaved-dialog__button--primary {
  background: rgb(var(--color-primary-soft) / 0.68);
  color: rgb(var(--color-primary));
}

.unsaved-dialog__button:hover:not(:disabled) {
  transform: translateY(-1px);
}

.unsaved-dialog__button--secondary:hover:not(:disabled) {
  background: rgb(255 255 255 / 0.48);
}

.unsaved-dialog__button--primary:hover:not(:disabled) {
  background: rgb(var(--color-primary-soft) / 0.78);
}

.unsaved-dialog__button:disabled {
  cursor: not-allowed;
  opacity: 0.52;
}

.unsaved-dialog-enter-active,
.unsaved-dialog-leave-active {
  transition: opacity 0.18s ease;
}

.unsaved-dialog-enter-active .unsaved-dialog__panel,
.unsaved-dialog-leave-active .unsaved-dialog__panel {
  transition:
    opacity 0.18s ease,
    transform 0.18s ease;
}

.unsaved-dialog-enter-from,
.unsaved-dialog-leave-to {
  opacity: 0;
}

.unsaved-dialog-enter-from .unsaved-dialog__panel,
.unsaved-dialog-leave-to .unsaved-dialog__panel {
  opacity: 0;
  transform: translateY(0.5rem) scale(0.985);
}

@media (max-width: 640px) {
  .unsaved-dialog__actions {
    flex-direction: column-reverse;
  }

  .unsaved-dialog__button {
    width: 100%;
  }
}
</style>

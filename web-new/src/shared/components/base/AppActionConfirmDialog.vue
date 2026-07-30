<!--
 * 操作確認彈窗。
 * 1. 以專案內彈窗取代瀏覽器原生二次確認。
 * 2. 提供取消、確認與處理中狀態。
 -->
<script setup lang="ts">
interface AppActionConfirmDialogProps {
  open: boolean;
  title: string;
  description: string;
  cancelLabel: string;
  confirmLabel: string;
  confirming?: boolean;
}

const props = withDefaults(defineProps<AppActionConfirmDialogProps>(), {
  confirming: false,
});

const emit = defineEmits<{
  (event: 'cancel'): void;
  (event: 'confirm'): void;
}>();
</script>

<template>
  <Teleport to="body">
    <Transition name="action-confirm-dialog">
      <div
        v-if="props.open"
        class="action-confirm-dialog"
        role="alertdialog"
        aria-modal="true"
        :aria-label="props.title"
        @click.self="emit('cancel')"
      >
        <div class="action-confirm-dialog__panel">
          <div class="action-confirm-dialog__header">
            <span
              class="action-confirm-dialog__mark"
              aria-hidden="true"
            />
            <div>
              <h2>{{ props.title }}</h2>
              <p>{{ props.description }}</p>
            </div>
          </div>

          <div class="action-confirm-dialog__actions">
            <button
              type="button"
              class="action-confirm-dialog__button action-confirm-dialog__button--secondary"
              :disabled="props.confirming"
              @click="emit('cancel')"
            >
              {{ props.cancelLabel }}
            </button>
            <button
              type="button"
              class="action-confirm-dialog__button action-confirm-dialog__button--primary"
              :disabled="props.confirming"
              @click="emit('confirm')"
            >
              {{ props.confirmLabel }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.action-confirm-dialog {
  position: fixed;
  z-index: 120;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgb(15 23 42 / 0.32);
  padding: 1rem;
}

.action-confirm-dialog__panel {
  width: min(100%, 30rem);
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
  padding: 1.25rem;
  box-shadow: 0 24px 60px rgb(15 23 42 / 0.2);
  color: rgb(var(--color-text));
}

.action-confirm-dialog__header {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr);
  gap: 0.9rem;
  align-items: start;
}

.action-confirm-dialog__mark {
  display: inline-flex;
  width: 0.65rem;
  height: 2.25rem;
  border-radius: 999px;
  background: rgb(var(--color-primary));
}

.action-confirm-dialog__panel h2 {
  margin: 0;
  color: rgb(var(--color-text));
  font-family: var(--font-display);
  font-size: 1.15rem;
  font-weight: 600;
  line-height: 1.35;
}

.action-confirm-dialog__panel p {
  margin: 0.55rem 0 0;
  color: rgb(var(--color-text-muted));
  font-size: 0.92rem;
  line-height: 1.65;
}

.action-confirm-dialog__actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 0.65rem;
  margin-top: 1.25rem;
}

.action-confirm-dialog__button {
  display: inline-flex;
  min-height: 2.65rem;
  cursor: pointer;
  align-items: center;
  justify-content: center;
  border: 1px solid transparent;
  border-radius: 6px;
  padding: 0 0.95rem;
  font-size: 0.875rem;
  font-weight: 700;
  line-height: 1.2;
}

.action-confirm-dialog__button--secondary {
  border-color: rgb(var(--color-border));
  background: rgb(var(--color-surface-raised));
  color: rgb(var(--color-text));
}

.action-confirm-dialog__button--primary {
  background: rgb(var(--color-primary));
  color: rgb(var(--color-primary-contrast));
}

.action-confirm-dialog__button:disabled {
  cursor: not-allowed;
  opacity: 0.52;
}

.action-confirm-dialog-enter-active,
.action-confirm-dialog-leave-active {
  transition: opacity 0.18s ease;
}

.action-confirm-dialog-enter-active .action-confirm-dialog__panel,
.action-confirm-dialog-leave-active .action-confirm-dialog__panel {
  transition: opacity 0.18s ease, transform 0.18s ease;
}

.action-confirm-dialog-enter-from,
.action-confirm-dialog-leave-to {
  opacity: 0;
}

.action-confirm-dialog-enter-from .action-confirm-dialog__panel,
.action-confirm-dialog-leave-to .action-confirm-dialog__panel {
  opacity: 0;
  transform: translateY(0.5rem) scale(0.985);
}

@media (max-width: 640px) {
  .action-confirm-dialog__actions {
    flex-direction: column-reverse;
  }

  .action-confirm-dialog__button {
    width: 100%;
  }
}
</style>

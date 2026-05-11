<!--
 * 發布頁側欄。
 * 1. 顯示發布流程、完成度與可見範圍設定。
 * 2. 對齊二手交易篩選頁的側欄密度與主色。
-->
<script setup lang="ts">
import { useI18n } from 'vue-i18n';

import AppIcon from '@/shared/components/base/AppIcon.vue';

import type {
  EditorMetric,
  EditorOption,
  EditorWorkflowStep,
  ListingEditorVisibility,
} from '../editor';

interface EditorSidebarProps {
  isEditing: boolean;
  metrics: EditorMetric[];
  selectedVisibility: ListingEditorVisibility;
  steps: EditorWorkflowStep[];
  visibilityOptions: EditorOption<ListingEditorVisibility>[];
}

const props = defineProps<EditorSidebarProps>();

const emit = defineEmits<{
  'update:selectedVisibility': [value: ListingEditorVisibility];
}>();

const { t } = useI18n();

// 1. 更新可見範圍
const updateVisibility = (value: ListingEditorVisibility): void => {
  emit('update:selectedVisibility', value);
};
</script>

<template>
  <aside class="editor-sidebar">
    <div class="border-b border-border pb-4">
      <p class="editor-kicker">
        {{ t('marketplace.editor.workflow') }}
      </p>
      <h1>
        {{ props.isEditing ? t('marketplace.editor.editTitle') : t('marketplace.editor.createTitle') }}
      </h1>
      <p class="editor-sidebar__description">
        {{ t('marketplace.editor.description') }}
      </p>
    </div>

    <section class="space-y-4">
      <h2 class="editor-section-title">
        {{ t('marketplace.editor.progressTitle') }}
      </h2>
      <div class="editor-step-list">
        <div
          v-for="step in props.steps"
          :key="step.label"
          class="editor-step"
          :class="`editor-step--${step.status}`"
        >
          <span class="editor-step__icon">
            <AppIcon
              :name="step.status === 'done' ? 'check-circle' : 'clock'"
              :size="16"
            />
          </span>
          <span class="min-w-0">
            <strong>{{ step.label }}</strong>
            <small>{{ step.description }}</small>
          </span>
        </div>
      </div>
    </section>

    <section class="space-y-4 border-t border-border pt-6">
      <h2 class="editor-section-title">
        {{ t('marketplace.editor.visibilityTitle') }}
      </h2>
      <div class="flex flex-wrap gap-2">
        <button
          v-for="option in props.visibilityOptions"
          :key="option.value"
          type="button"
          class="editor-chip"
          :class="props.selectedVisibility === option.value ? 'editor-chip-active' : 'editor-chip-idle'"
          @click="updateVisibility(option.value)"
        >
          {{ option.label }}
        </button>
      </div>
    </section>

    <section class="grid grid-cols-2 gap-3 border-t border-border pt-6">
      <div
        v-for="metric in props.metrics"
        :key="metric.label"
        class="editor-metric"
      >
        <p>{{ metric.value }}</p>
        <span>{{ metric.label }}</span>
      </div>
    </section>
  </aside>
</template>

<style scoped>
.editor-sidebar {
  display: flex;
  width: 16rem;
  height: 100%;
  flex-shrink: 0;
  flex-direction: column;
  gap: 2rem;
  overflow-y: auto;
  overscroll-behavior: contain;
  padding-right: 2rem;
  scrollbar-width: none;
  scrollbar-gutter: stable;
}

.editor-sidebar::-webkit-scrollbar {
  display: none;
}

.editor-kicker,
.editor-section-title {
  color: rgb(var(--color-text-muted));
  font-size: 0.75rem;
  font-weight: 600;
  letter-spacing: 0.1em;
  line-height: 1;
  text-transform: uppercase;
}

.editor-sidebar h1 {
  margin: 0.75rem 0 0;
  color: rgb(var(--color-text));
  font-family: var(--font-display);
  font-size: 1.5rem;
  font-weight: 500;
  line-height: 1.4;
}

.editor-sidebar__description {
  margin: 0.5rem 0 0;
  color: rgb(var(--color-text-muted));
  font-size: 0.875rem;
  line-height: 1.7;
}

.editor-step-list {
  display: grid;
  gap: 0.75rem;
}

.editor-step {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr);
  gap: 0.75rem;
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.75rem;
  background: rgb(var(--color-surface-raised));
  padding: 0.9rem;
  box-shadow: 0 4px 24px rgb(0 0 0 / 0.02);
}

.editor-step__icon {
  display: inline-flex;
  height: 1.75rem;
  width: 1.75rem;
  align-items: center;
  justify-content: center;
  border-radius: 9999px;
  background: rgb(var(--color-surface-muted));
  color: rgb(var(--color-text-muted));
}

.editor-step--done .editor-step__icon,
.editor-step--current .editor-step__icon {
  background: rgb(0 39 39 / 0.07);
  color: rgb(var(--color-primary));
}

.editor-step strong {
  display: block;
  color: rgb(var(--color-text));
  font-size: 0.875rem;
  font-weight: 700;
  line-height: 1.3;
}

.editor-step small {
  display: block;
  margin-top: 0.35rem;
  color: rgb(var(--color-text-muted));
  font-size: 0.75rem;
  line-height: 1.45;
}

.editor-chip {
  border-radius: 9999px;
  border-width: 1px;
  padding: 0.5rem 1rem;
  font-size: 0.875rem;
  font-weight: 500;
  letter-spacing: 0.02em;
  line-height: 1;
  transition:
    border-color 0.2s ease,
    background-color 0.2s ease,
    color 0.2s ease;
}

.editor-chip-active {
  border-color: rgb(var(--color-primary));
  background: rgb(0 39 39 / 0.05);
  color: rgb(var(--color-primary));
}

.editor-chip-idle {
  border-color: rgb(var(--color-border));
  color: rgb(var(--color-text-muted));
}

.editor-chip-idle:hover {
  border-color: rgb(var(--color-text-muted));
  color: rgb(var(--color-text));
}

.editor-metric {
  min-height: 5.6rem;
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.75rem;
  background: rgb(var(--color-surface-raised));
  padding: 0.9rem;
  box-shadow: 0 4px 24px rgb(0 0 0 / 0.02);
}

.editor-metric p {
  margin: 0;
  color: rgb(var(--color-primary));
  font-family: var(--font-display);
  font-size: 2rem;
  line-height: 1;
}

.editor-metric span {
  display: block;
  margin-top: 0.75rem;
  color: rgb(var(--color-text-muted));
  font-size: 0.75rem;
  font-weight: 600;
  line-height: 1.45;
}

@media (max-width: 767px) {
  .editor-sidebar {
    width: 100%;
    height: auto;
    overflow: visible;
    padding-right: 0;
  }
}
</style>

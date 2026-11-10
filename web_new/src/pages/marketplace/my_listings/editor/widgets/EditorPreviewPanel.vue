<!--
 * 發布頁預覽區。
 * 1. 顯示商品卡片、發布檢查與主要操作。
 * 2. 對齊新增帖子頁右側 sticky preview 版式。
-->
<script setup lang="ts">
import { useI18n } from 'vue-i18n';

import AppIcon from '@/shared/components/base/AppIcon.vue';

import type {
  EditorChecklistItem,
  EditorImageSlot,
} from '../editor';

interface EditorPreviewPanelProps {
  checklist: EditorChecklistItem[];
  chargeHint: string;
  coverImage?: EditorImageSlot;
  isFirstStep: boolean;
  isLastStep: boolean;
  isEditing: boolean;
  isPublishing: boolean;
  isSaving: boolean;
  previewPrice: string;
  previewTagLabels: string[];
  previewTitle: string;
  readyToPublish: boolean;
  readyToSaveDraft: boolean;
}

const props = defineProps<EditorPreviewPanelProps>();

const emit = defineEmits<{
  nextStep: [];
  publish: [];
  previousStep: [];
  saveDraft: [];
}>();

const { t } = useI18n();

// 1. 觸發儲存草稿
const saveDraft = (): void => {
  emit('saveDraft');
};

// 2. 觸發下一步
const goNextStep = (): void => {
  emit('nextStep');
};

// 3. 觸發上一步
const goPreviousStep = (): void => {
  emit('previousStep');
};

// 4. 觸發主要操作
const runPrimaryAction = (): void => {
  if (props.isEditing) {
    emit('saveDraft');
    return;
  }

  emit('publish');
};
</script>

<template>
  <aside class="editor-preview-panel">
    <p class="editor-preview-kicker">
      {{ t('marketplace.editor.livePreview') }}
    </p>

    <div class="editor-preview-card group">
      <div class="editor-preview-card__media">
        <img
          v-if="props.coverImage?.url"
          :src="props.coverImage.url"
          :alt="props.previewTitle"
          class="transition-transform duration-500 group-hover:scale-105"
        />
        <div
          v-else
          class="editor-preview-card__empty"
        >
          <AppIcon
            name="picture"
            :size="44"
          />
        </div>
      </div>

      <div class="editor-preview-card__body">
        <h2>{{ props.previewTitle }}</h2>

        <div class="editor-preview-meta-line">
          {{ props.previewTagLabels.join(' · ') }}
        </div>

        <strong>{{ props.previewPrice }}</strong>
      </div>
    </div>

    <section class="editor-checklist">
      <p class="editor-preview-kicker">
        {{ t('marketplace.editor.checklistTitle') }}
      </p>
      <div class="editor-checklist__items">
        <div
          v-for="item in props.checklist"
          :key="item.label"
          class="editor-checklist__item"
          :class="item.complete ? 'editor-checklist__item--complete' : ''"
        >
          <AppIcon
            :name="item.complete ? 'check-circle' : 'clock'"
            :size="16"
          />
          <span>{{ item.label }}</span>
        </div>
      </div>
    </section>

    <div class="editor-actions">
      <p class="editor-charge-hint">
        {{ props.chargeHint }}
      </p>
      <div class="editor-step-actions">
        <button
          type="button"
          class="editor-action editor-action--secondary"
          :disabled="props.isFirstStep || props.isSaving || props.isPublishing"
          @click="goPreviousStep"
        >
          {{ t('marketplace.editor.previousStep') }}
        </button>
        <button
          v-if="!props.isLastStep"
          type="button"
          class="editor-action editor-action--primary"
          :disabled="props.isSaving || props.isPublishing"
          @click="goNextStep"
        >
          {{ t('marketplace.editor.nextStep') }}
        </button>
      </div>
      <button
        v-if="props.isLastStep"
        type="button"
        class="editor-action editor-action--primary"
        :disabled="props.isEditing ? !props.readyToSaveDraft || props.isSaving || props.isPublishing : !props.readyToPublish || props.isSaving || props.isPublishing"
        @click="runPrimaryAction"
      >
        <AppIcon
          :name="props.isEditing ? 'check-circle' : 'send'"
          :size="17"
        />
        <span>
          {{
            props.isSaving || props.isPublishing
              ? t('common.status.loading')
              : props.isEditing
                ? t('marketplace.editor.saveChanges')
                : t('common.action.publish')
          }}
        </span>
      </button>
      <button
        v-if="props.isLastStep && !props.isEditing"
        type="button"
        class="editor-action editor-action--secondary"
        :disabled="props.isSaving || props.isPublishing"
        @click="saveDraft"
      >
        {{ props.isSaving ? t('common.status.loading') : t('common.action.saveDraft') }}
      </button>
    </div>
  </aside>
</template>

<style scoped>
.editor-preview-panel {
  position: sticky;
  top: calc(var(--app-header-offset, 0rem) + 1rem);
  display: grid;
  align-content: start;
  gap: 0.75rem;
  min-width: 0;
}

.editor-preview-kicker {
  margin: 0;
  color: rgb(var(--color-text-muted));
  font-size: 0.6875rem;
  font-weight: 700;
  letter-spacing: 0.12em;
  line-height: 1;
  text-transform: uppercase;
}

.editor-preview-card,
.editor-checklist {
  overflow: hidden;
  border: 1px solid rgb(var(--color-border));
  border-radius: 2px;
  background: rgb(var(--color-surface));
  box-shadow: none;
}

.editor-preview-card__media {
  position: relative;
  aspect-ratio: 4 / 3;
  overflow: hidden;
  background: rgb(var(--color-surface-muted));
}

.editor-preview-card__media img {
  height: 100%;
  width: 100%;
  object-fit: cover;
}

.editor-preview-card__empty {
  display: grid;
  height: 100%;
  place-items: center;
  color: rgb(var(--color-text-muted));
}

.editor-preview-card__body {
  display: grid;
  gap: 0.5rem;
  min-width: 0;
  padding: 0.9rem;
}

.editor-preview-card__body h2 {
  margin: 0;
  min-width: 0;
  overflow: hidden;
  color: rgb(var(--color-text));
  font-family: var(--font-display);
  font-size: 1rem;
  font-weight: 500;
  line-height: 1.25;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.editor-preview-card__body strong {
  display: block;
  color: rgb(var(--color-primary));
  font-size: 1.05rem;
  font-weight: 800;
  line-height: 1.2;
}

.editor-preview-meta-line {
  min-width: 0;
  overflow: hidden;
  color: rgb(var(--color-text-muted));
  font-size: 0.8125rem;
  font-weight: 700;
  line-height: 1.3;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.editor-checklist {
  padding: 0.9rem;
}

.editor-checklist__items {
  display: grid;
  gap: 0.55rem;
  margin-top: 0.75rem;
}

.editor-checklist__item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: rgb(var(--color-text-muted));
  font-size: 0.875rem;
  font-weight: 600;
}

.editor-checklist__item--complete {
  color: rgb(var(--color-primary));
}

.editor-actions {
  display: grid;
  gap: 0.6rem;
  margin-top: 0.25rem;
}

.editor-step-actions {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.5rem;
}

.editor-charge-hint {
  margin: 0;
  color: rgb(var(--color-text-muted));
  font-size: 0.82rem;
  font-weight: 700;
  line-height: 1.45;
}

.editor-action {
  display: inline-flex;
  min-height: 2.75rem;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  border: 1px solid rgb(var(--color-primary));
  border-radius: 2px;
  padding: 0.7rem 1rem;
  font-size: 0.875rem;
  font-weight: 800;
  letter-spacing: 0.02em;
  line-height: 1;
  transition:
    border-color 0.2s ease,
    background-color 0.2s ease,
    color 0.2s ease,
    opacity 0.2s ease;
}

.editor-action--primary {
  background: rgb(var(--color-primary));
  color: rgb(var(--color-primary-contrast));
}

.editor-action--primary:hover:not(:disabled) {
  border-color: rgb(var(--color-text));
  background: rgb(var(--color-text));
}

.editor-action--secondary {
  background: transparent;
  color: rgb(var(--color-primary));
}

.editor-action--secondary:hover:not(:disabled) {
  background: rgb(var(--color-surface-muted));
}

.editor-action:disabled {
  cursor: not-allowed;
  opacity: 0.45;
}

@media (max-width: 1023px) {
  .editor-preview-panel {
    position: static;
  }
}
</style>

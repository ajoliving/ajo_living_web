<!--
 * 發布頁預覽區。
 * 1. 顯示商品卡片、發布檢查與主要操作。
 * 2. 對齊新增帖子頁右側 sticky preview 版式。
-->
<script setup lang="ts">
import { useI18n } from 'vue-i18n';

import AppIcon from '@/shared/components/base/AppIcon.vue';

import type { EditorChecklistItem, EditorImageSlot } from '../editor';

interface EditorPreviewPanelProps {
  checklist: EditorChecklistItem[];
  coverImage?: EditorImageSlot;
  isPublishing: boolean;
  isSaving: boolean;
  previewPrice: string;
  previewTagLabels: string[];
  previewTitle: string;
  readyToPublish: boolean;
}

const props = defineProps<EditorPreviewPanelProps>();

const emit = defineEmits<{
  publish: [];
  saveDraft: [];
}>();

const { t } = useI18n();

// 1. 觸發儲存草稿
const saveDraft = (): void => {
  emit('saveDraft');
};

// 2. 觸發發布
const publish = (): void => {
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
      <button
        type="button"
        class="editor-action editor-action--primary"
        :disabled="!props.readyToPublish || props.isSaving || props.isPublishing"
        @click="publish"
      >
        <AppIcon
          name="send"
          :size="17"
        />
        <span>{{ props.isPublishing ? t('common.status.loading') : t('common.action.publish') }}</span>
      </button>
      <button
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
  top: calc(var(--app-header-offset, 0rem) + 2rem);
  display: grid;
  align-content: start;
  gap: 1rem;
  min-width: 0;
}

.editor-preview-kicker {
  margin: 0;
  color: #717878;
  font-size: 0.75rem;
  font-weight: 700;
  letter-spacing: 0.1em;
  line-height: 1;
  text-transform: uppercase;
}

.editor-preview-card,
.editor-checklist {
  overflow: hidden;
  border: 1px solid #e2e3e1;
  border-radius: 0.75rem;
  background: #ffffff;
  box-shadow: 0 10px 30px -5px rgb(0 39 39 / 0.05);
}

.editor-preview-card__media {
  position: relative;
  aspect-ratio: 4 / 3;
  overflow: hidden;
  background: #f4f4f2;
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
  color: #717878;
}

.editor-preview-card__body {
  display: grid;
  gap: 0.65rem;
  min-width: 0;
  padding: 1.125rem;
}

.editor-preview-card__body h2 {
  margin: 0;
  min-width: 0;
  overflow: hidden;
  color: #1a1c1b;
  font-family: var(--font-display);
  font-size: 1.15rem;
  font-weight: 500;
  line-height: 1.25;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.editor-preview-card__body strong {
  display: block;
  color: #002727;
  font-size: 1.2rem;
  font-weight: 800;
  line-height: 1.2;
}

.editor-preview-meta-line {
  min-width: 0;
  overflow: hidden;
  color: #414848;
  font-size: 0.8125rem;
  font-weight: 700;
  line-height: 1.3;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.editor-checklist {
  padding: 1.25rem;
}

.editor-checklist__items {
  display: grid;
  gap: 0.75rem;
  margin-top: 1rem;
}

.editor-checklist__item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: #717878;
  font-size: 0.875rem;
  font-weight: 600;
}

.editor-checklist__item--complete {
  color: #002727;
}

.editor-actions {
  display: grid;
  gap: 0.75rem;
  margin-top: 0.5rem;
}

.editor-action {
  display: inline-flex;
  min-height: 3.25rem;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  border: 1px solid #002727;
  border-radius: 9999px;
  padding: 0.8rem 1.25rem;
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
  background: #002727;
  color: #ffffff;
}

.editor-action--primary:hover:not(:disabled) {
  border-color: #1a1c1b;
  background: #1a1c1b;
}

.editor-action--secondary {
  background: transparent;
  color: #002727;
}

.editor-action--secondary:hover:not(:disabled) {
  background: #f4f4f2;
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

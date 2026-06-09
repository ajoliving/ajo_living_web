<!--
 * 我的帖子發布頁。
 * 1. 組裝發布表單、媒體上傳與即時預覽。
 * 2. 使用左側步驟表單與右側 sticky preview 的發布體驗。
-->
<script setup lang="ts">
import AppUnsavedChangesDialog from '@/shared/components/base/AppUnsavedChangesDialog.vue';

import { useMarketplaceListingEditorPage } from './editor';
import EditorFormPanel from './widgets/EditorFormPanel.vue';
import EditorPreviewPanel from './widgets/EditorPreviewPanel.vue';

const props = withDefaults(defineProps<{
  hideHeader?: boolean;
  modalMode?: boolean;
}>(), {
  hideHeader: false,
  modalMode: false,
});

const emit = defineEmits<{
  (event: 'close'): void;
  (event: 'saved'): void;
}>();

const {
  areaOptions,
  buildingOptions,
  buildingsLoading,
  businessStatusOptions,
  categoryOptions,
  chargeHint,
  checklist,
  conditionOptions,
  coverImage,
  formState,
  handleLeavePromptDecision,
  handleDroppedImageFiles,
  handleImageFileChange,
  handleImageFilesChange,
  imageSlots,
  isEditing,
  isLeavePromptOpen,
  isLoading,
  isPublishing,
  isSaving,
  moveImageSlot,
  previewPrice,
  previewTagLabels,
  previewTitle,
  priceModeOptions,
  readyToPublish,
  readyToSaveDraft,
  removeImageSlot,
  requestCloseEditor,
  saveAndBackToList,
  selectCoverImage,
  submitListing,
  visibilityOptions,
} = useMarketplaceListingEditorPage({
  useModalMode: props.modalMode,
  onClose: () => emit('close'),
  onSaved: () => emit('saved'),
});

defineExpose<{
  requestCloseEditor: () => Promise<void>;
}>({
  requestCloseEditor,
});
</script>

<template>
  <main class="listing-editor-page">
    <section
      v-if="!props.hideHeader"
      class="listing-editor-heading"
    >
      <p class="listing-editor-kicker">
        {{ $t('marketplace.editor.workflow') }}
      </p>
      <h1>
        {{ isEditing ? $t('marketplace.editor.editTitle') : $t('marketplace.editor.createTitle') }}
      </h1>
    </section>

    <div
      v-if="isLoading"
      class="listing-editor-loading"
    >
      {{ $t('common.status.loading') }}
    </div>

    <section class="listing-editor-workspace">
      <EditorFormPanel
        :area-options="areaOptions"
        :building-options="buildingOptions"
        :buildings-loading="buildingsLoading"
        :business-status-options="businessStatusOptions"
        :category-options="categoryOptions"
        :condition-options="conditionOptions"
        :form-state="formState"
        :image-slots="imageSlots"
        :price-mode-options="priceModeOptions"
        :visibility-options="visibilityOptions"
        @images-change="handleImageFilesChange"
        @images-drop="handleDroppedImageFiles"
        @image-change="handleImageFileChange"
        @move-image="moveImageSlot"
        @remove-image="removeImageSlot"
        @select-cover="selectCoverImage"
      />

      <EditorPreviewPanel
        :checklist="checklist"
        :charge-hint="chargeHint"
        :cover-image="coverImage"
        :is-publishing="isPublishing"
        :is-saving="isSaving"
        :is-editing="isEditing"
        :preview-price="previewPrice"
        :preview-tag-labels="previewTagLabels"
        :preview-title="previewTitle"
        :ready-to-publish="readyToPublish"
        :ready-to-save-draft="readyToSaveDraft"
        @publish="submitListing"
        @save-draft="saveAndBackToList"
      />
    </section>

    <AppUnsavedChangesDialog
      :open="isLeavePromptOpen"
      :title="$t('marketplace.editor.unsavedLeaveTitle')"
      :description="$t('marketplace.editor.unsavedLeaveDescription')"
      :save-label="isSaving ? $t('marketplace.editor.savingDraft') : $t('marketplace.editor.unsavedLeaveSave')"
      :discard-label="$t('marketplace.editor.unsavedLeaveDiscard')"
      :stay-label="$t('marketplace.editor.unsavedLeaveStay')"
      :saving="isSaving"
      @save="handleLeavePromptDecision('save')"
      @discard="handleLeavePromptDecision('discard')"
      @stay="handleLeavePromptDecision('stay')"
    />
  </main>
</template>

<style scoped>
.listing-editor-page {
  width: 100%;
  max-width: none;
  margin: 0;
  padding: 0;
  color: rgb(var(--color-text));
}

.listing-editor-heading {
  margin-top: 0;
}

.listing-editor-kicker {
  margin: 0 0 0.7rem;
  color: rgb(var(--color-text-muted));
  font-size: 0.75rem;
  font-weight: 700;
  letter-spacing: 0.1em;
  line-height: 1;
  text-transform: uppercase;
}

.listing-editor-heading h1 {
  margin: 0;
  color: rgb(var(--color-primary));
  font-family: var(--font-display);
  font-size: clamp(2rem, 4vw, 3rem);
  font-weight: 500;
  line-height: 1.2;
}

.listing-editor-loading {
  margin-top: 1.5rem;
  border: 1px solid rgb(var(--color-border));
  border-radius: 0.75rem;
  background: rgb(var(--color-surface-raised));
  padding: 1rem 1.25rem;
  color: rgb(var(--color-text-muted));
  font-size: 0.875rem;
}

.listing-editor-workspace {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 2rem;
  align-items: start;
  margin-top: 2rem;
  min-width: 0;
}

@media (max-width: 767px) {
  .listing-editor-page {
    padding: 0;
  }
}

@media (min-width: 1024px) {
  .listing-editor-workspace {
    grid-template-columns: minmax(0, 2fr) minmax(20rem, 1fr);
  }
}
</style>

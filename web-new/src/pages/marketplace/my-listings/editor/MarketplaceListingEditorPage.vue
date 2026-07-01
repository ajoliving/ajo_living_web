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

const {
  areaOptions,
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
  saveAndBackToList,
  selectCoverImage,
  submitListing,
  visibilityOptions,
} = useMarketplaceListingEditorPage();
</script>

<template>
  <main class="listing-editor-page">
    <section class="listing-editor-heading">
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
  display: grid;
  gap: 1rem;
  width: 100%;
  color: rgb(var(--color-text));
}

.listing-editor-heading {
  display: grid;
  gap: 0.4rem;
  margin: 0;
  border: 1px solid rgb(var(--color-border));
  border-radius: 2px;
  background: rgb(var(--color-surface));
  padding: 1rem;
}

.listing-editor-kicker {
  margin: 0;
  color: rgb(var(--color-text-muted));
  font-size: 0.6875rem;
  font-weight: 700;
  letter-spacing: 0.12em;
  line-height: 1;
  text-transform: uppercase;
}

.listing-editor-heading h1 {
  margin: 0;
  color: rgb(var(--color-primary));
  font-family: var(--font-display);
  font-size: clamp(1.45rem, 2.2vw, 1.9rem);
  font-weight: 500;
  line-height: 1.22;
}

.listing-editor-loading {
  margin-top: 1rem;
  border: 1px solid rgb(var(--color-border));
  border-radius: 2px;
  background: rgb(var(--color-surface-raised));
  padding: 0.85rem 1rem;
  color: rgb(var(--color-text-muted));
  font-size: 0.875rem;
}

.listing-editor-workspace {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 1rem;
  align-items: start;
  margin-top: 1rem;
  min-width: 0;
}

@media (max-width: 767px) {
  .listing-editor-page {
    max-width: none;
    padding: 1rem var(--layout-page-padding-inline) 4rem;
  }
}

@media (min-width: 1024px) {
  .listing-editor-workspace {
    grid-template-columns: minmax(0, 2fr) minmax(20rem, 1fr);
  }
}
</style>

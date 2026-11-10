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
  activeEditorStep,
  activeEditorStepIndex,
  areaOptions,
  businessStatusOptions,
  categoryOptions,
  chargeHint,
  checklist,
  conditionOptions,
  coverImage,
  clearPublishValidationError,
  editorSteps,
  formState,
  goNextEditorStep,
  goPreviousEditorStep,
  handleLeavePromptDecision,
  handleDroppedImageFiles,
  handleImageFileChange,
  handleImageFilesChange,
  imageSlots,
  isEditing,
  isFirstEditorStep,
  isLastEditorStep,
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
  selectEditorStep,
  selectCoverImage,
  submitListing,
  validationErrors,
  validationItems,
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
      <div
        class="listing-editor-progress"
        :aria-label="$t('marketplace.editor.publishSteps')"
      >
        <button
          v-for="(step, stepIndex) in editorSteps"
          :key="step.key"
          type="button"
          class="listing-editor-progress__step"
          :class="{
            'listing-editor-progress__step--active': activeEditorStep === step.key,
            'listing-editor-progress__step--done': stepIndex < activeEditorStepIndex,
          }"
          :aria-current="activeEditorStep === step.key ? 'step' : undefined"
          @click="selectEditorStep(step.key)"
        >
          {{ stepIndex + 1 }}
        </button>
      </div>

      <EditorFormPanel
        :active-step="activeEditorStep"
        :area-options="areaOptions"
        :business-status-options="businessStatusOptions"
        :category-options="categoryOptions"
        :condition-options="conditionOptions"
        :form-state="formState"
        :image-slots="imageSlots"
        :price-mode-options="priceModeOptions"
        :validation-errors="validationErrors"
        :validation-items="validationItems"
        :visibility-options="visibilityOptions"
        @clear-validation-error="clearPublishValidationError"
        @images-change="handleImageFilesChange"
        @images-drop="handleDroppedImageFiles"
        @image-change="handleImageFileChange"
        @move-image="moveImageSlot"
        @remove-image="removeImageSlot"
        @select-cover="selectCoverImage"
        @select-validation-step="selectEditorStep"
      />

      <EditorPreviewPanel
        :checklist="checklist"
        :charge-hint="chargeHint"
        :cover-image="coverImage"
        :is-first-step="isFirstEditorStep"
        :is-last-step="isLastEditorStep"
        :is-publishing="isPublishing"
        :is-saving="isSaving"
        :is-editing="isEditing"
        :preview-price="previewPrice"
        :preview-tag-labels="previewTagLabels"
        :preview-title="previewTitle"
        :ready-to-publish="readyToPublish"
        :ready-to-save-draft="readyToSaveDraft"
        @next-step="goNextEditorStep"
        @publish="submitListing"
        @previous-step="goPreviousEditorStep"
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

.listing-editor-progress {
  position: relative;
  display: grid;
  grid-column: 1 / -1;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  align-items: center;
  padding: 0.25rem 0.375rem;
}

.listing-editor-progress::before {
  position: absolute;
  top: 50%;
  right: 1.25rem;
  left: 1.25rem;
  height: 1px;
  background: rgb(var(--color-border));
  content: "";
  transform: translateY(-50%);
}

.listing-editor-progress__step {
  position: relative;
  z-index: 1;
  display: inline-flex;
  width: 1.75rem;
  height: 1.75rem;
  align-items: center;
  justify-content: center;
  justify-self: center;
  border: 1px solid rgb(var(--color-border));
  border-radius: 9999px;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text-muted));
  font-size: 0.75rem;
  font-weight: 700;
  line-height: 1;
  cursor: pointer;
}

.listing-editor-progress__step:hover,
.listing-editor-progress__step--active {
  border-color: rgb(var(--color-primary));
  color: rgb(var(--color-primary));
}

.listing-editor-progress__step--active,
.listing-editor-progress__step--done {
  background: rgb(var(--color-primary));
  color: rgb(var(--color-primary-contrast));
}

.listing-editor-progress__step--active:hover,
.listing-editor-progress__step--done:hover {
  color: rgb(var(--color-primary-contrast));
}

@media (max-width: 767px) {
  .listing-editor-page {
    max-width: none;
    padding: 1rem var(--layout-page-padding-inline) calc(var(--app-mobile-content-bottom) + 1rem);
  }
}

@media (min-width: 1024px) {
  .listing-editor-workspace {
    grid-template-columns: minmax(0, 2fr) minmax(20rem, 1fr);
  }
}
</style>

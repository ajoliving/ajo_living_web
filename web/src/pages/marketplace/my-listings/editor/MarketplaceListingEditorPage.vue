<!--
 * 我的帖子發布頁。
 * 1. 組裝發布表單、媒體上傳與即時預覽。
 * 2. 使用左側步驟表單與右側 sticky preview 的發布體驗。
-->
<script setup lang="ts">
import { useMarketplaceListingEditorPage } from './editor';
import EditorFormPanel from './widgets/EditorFormPanel.vue';
import EditorPreviewPanel from './widgets/EditorPreviewPanel.vue';

const {
  areaOptions,
  businessStatusOptions,
  categoryOptions,
  checklist,
  conditionOptions,
  coverImage,
  formState,
  handleDroppedImageFiles,
  handleImageFileChange,
  handleImageFilesChange,
  imageSlots,
  isEditing,
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
  </main>
</template>

<style scoped>
.listing-editor-page {
  width: 100%;
  max-width: 1280px;
  margin: 0 auto;
  padding: 1rem 2rem 5rem;
  color: #1a1c1b;
}

.listing-editor-heading {
  margin-top: 0;
}

.listing-editor-kicker {
  margin: 0 0 0.7rem;
  color: #717878;
  font-size: 0.75rem;
  font-weight: 700;
  letter-spacing: 0.1em;
  line-height: 1;
  text-transform: uppercase;
}

.listing-editor-heading h1 {
  margin: 0;
  color: #002727;
  font-family: var(--font-display);
  font-size: clamp(2rem, 4vw, 3rem);
  font-weight: 500;
  line-height: 1.2;
}

.listing-editor-loading {
  margin-top: 1.5rem;
  border: 1px solid #e2e3e1;
  border-radius: 0.75rem;
  background: #f9f9f7;
  padding: 1rem 1.25rem;
  color: #717878;
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
    padding: 2.5rem 1.25rem;
  }
}

@media (min-width: 1024px) {
  .listing-editor-workspace {
    grid-template-columns: minmax(0, 2fr) minmax(20rem, 1fr);
  }
}
</style>

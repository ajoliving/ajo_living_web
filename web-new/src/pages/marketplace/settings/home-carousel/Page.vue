<!--
 * 首頁輪播設定頁。
 * 1. 上傳首頁橫向輪播 PNG 圖片。
 * 2. 管理圖片排序、移除與保存。
-->
<script setup lang="ts">
import AppIcon from '@/shared/components/base/AppIcon.vue';

import { useHomeCarouselSettingsPage } from './carousel';

const {
  canSave,
  handleFilesSelected,
  images,
  loadCarousel,
  loading,
  moveImage,
  removeImage,
  saveCarousel,
  saving,
  t,
} = useHomeCarouselSettingsPage();
</script>

<template>
  <section class="home-setting-page">
    <div class="home-setting-header">
      <div>
        <p class="home-setting-kicker">
          {{ t('marketplace.settings.homeContentKicker') }}
        </p>
        <h2>{{ t('marketplace.settings.homeCarouselTitle') }}</h2>
        <p>{{ t('marketplace.settings.homeCarouselDescription') }}</p>
      </div>
      <div class="home-setting-actions">
        <button
          type="button"
          class="home-setting-button home-setting-button--secondary"
          :disabled="loading"
          @click="loadCarousel"
        >
          <AppIcon
            name="reload"
            :size="16"
          />
          <span>{{ t('marketplace.settings.refresh') }}</span>
        </button>
        <button
          type="button"
          class="home-setting-button home-setting-button--primary"
          :disabled="!canSave"
          @click="saveCarousel"
        >
          <AppIcon
            name="check-circle"
            :size="16"
          />
          <span>{{ saving ? t('marketplace.settings.saving') : t('marketplace.settings.homeCarouselSave') }}</span>
        </button>
      </div>
    </div>

    <label class="home-upload-panel">
      <input
        type="file"
        accept="image/png"
        multiple
        @change="handleFilesSelected"
      />
      <AppIcon
        name="cloud-upload"
        :size="28"
      />
      <span>{{ t('marketplace.settings.homeCarouselUploadTitle') }}</span>
      <small>{{ t('marketplace.settings.homeCarouselUploadHint') }}</small>
    </label>

    <div
      v-if="loading"
      class="home-setting-empty"
    >
      {{ t('common.status.loading') }}
    </div>

    <div
      v-else-if="images.length === 0"
      class="home-setting-empty"
    >
      {{ t('marketplace.settings.homeCarouselEmpty') }}
    </div>

    <div
      v-else
      class="home-carousel-grid"
    >
      <article
        v-for="image in images"
        :key="image.key"
        class="home-carousel-card"
      >
        <div class="home-carousel-card__media">
          <img
            :src="image.url"
            :alt="t('marketplace.settings.homeCarouselImageAlt', { index: image.sort_order })"
          />
          <span
            v-if="image.uploading"
            class="home-carousel-card__uploading"
          >
            {{ t('marketplace.settings.uploading') }}
          </span>
        </div>

        <div class="home-carousel-card__body">
          <p>{{ t('marketplace.settings.slotLabel', { index: image.sort_order }) }}</p>
          <small>{{ t('marketplace.settings.homeCarouselImageSpec') }}</small>
          <code>{{ image.object_key || image.file?.name }}</code>
        </div>

        <div class="home-carousel-card__actions">
          <button
            type="button"
            class="home-setting-mini-button"
            :disabled="image.sort_order === 1 || image.uploading"
            @click="moveImage(image.key, -1)"
          >
            {{ t('marketplace.settings.moveUp') }}
          </button>
          <button
            type="button"
            class="home-setting-mini-button"
            :disabled="image.sort_order === images.length || image.uploading"
            @click="moveImage(image.key, 1)"
          >
            {{ t('marketplace.settings.moveDown') }}
          </button>
          <button
            type="button"
            class="home-setting-mini-button"
            :disabled="image.uploading"
            @click="removeImage(image.key)"
          >
            {{ t('marketplace.editor.removeImage') }}
          </button>
        </div>
      </article>
    </div>
  </section>
</template>

<style scoped>
.home-setting-page {
  display: grid;
  gap: 1rem;
}

.home-setting-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
}

.home-setting-kicker {
  margin: 0;
  color: rgb(var(--color-text-muted));
  font-size: 0.75rem;
  font-weight: 800;
  letter-spacing: 0.1em;
  line-height: 1;
  text-transform: uppercase;
}

.home-setting-header h2 {
  margin: 0.35rem 0 0;
  color: rgb(var(--color-text));
  font-family: var(--font-display);
  font-size: clamp(1.5rem, 2vw, 2.25rem);
  font-weight: 650;
  line-height: 1.2;
}

.home-setting-header p:not(.home-setting-kicker) {
  margin: 0.45rem 0 0;
  color: rgb(var(--color-text-muted));
  font-size: 0.9rem;
  line-height: 1.7;
}

.home-setting-actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 0.65rem;
}

.home-setting-button,
.home-setting-mini-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.45rem;
  border: 1px solid rgb(var(--color-text));
  border-radius: 8px;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text));
  font-weight: 700;
}

.home-setting-button {
  min-height: 2.75rem;
  padding: 0 1rem;
  font-size: 0.9rem;
}

.home-setting-button--primary {
  background: rgb(var(--color-text));
  color: rgb(255 255 255);
}

.home-setting-mini-button {
  min-height: 2rem;
  padding: 0 0.65rem;
  font-size: 0.78rem;
}

.home-setting-button:disabled,
.home-setting-mini-button:disabled {
  cursor: not-allowed;
  opacity: 0.45;
}

.home-upload-panel {
  display: grid;
  justify-items: center;
  gap: 0.45rem;
  border: 1px dashed rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface-raised));
  padding: 1.25rem;
  color: rgb(var(--color-text));
  cursor: pointer;
}

.home-upload-panel input {
  display: none;
}

.home-upload-panel span {
  font-size: 1rem;
  font-weight: 700;
}

.home-upload-panel small {
  color: rgb(var(--color-text-muted));
  font-size: 0.82rem;
}

.home-setting-empty {
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface-raised));
  padding: 1rem;
  color: rgb(var(--color-text-muted));
  font-size: 0.9rem;
}

.home-carousel-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 1rem;
}

.home-carousel-card {
  overflow: hidden;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
}

.home-carousel-card__media {
  position: relative;
  aspect-ratio: 16 / 10;
  overflow: hidden;
  background: rgb(var(--color-border));
}

.home-carousel-card__media img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.home-carousel-card__uploading {
  position: absolute;
  inset: auto 0 0;
  background: rgb(26 28 27 / 0.76);
  padding: 0.55rem;
  color: rgb(255 255 255);
  font-size: 0.82rem;
  font-weight: 700;
  text-align: center;
}

.home-carousel-card__body {
  display: grid;
  gap: 0.35rem;
  padding: 0.85rem;
}

.home-carousel-card__body p {
  margin: 0;
  color: rgb(var(--color-text));
  font-weight: 700;
}

.home-carousel-card__body small {
  color: rgb(var(--color-text-muted));
  font-size: 0.78rem;
  line-height: 1.55;
}

.home-carousel-card__body code {
  overflow: hidden;
  color: rgb(var(--color-text-muted));
  font-size: 0.76rem;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.home-carousel-card__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.45rem;
  padding: 0 0.85rem 0.85rem;
}

@media (max-width: 767px) {
  .home-setting-header {
    flex-direction: column;
  }

  .home-setting-actions,
  .home-setting-button {
    width: 100%;
  }

  .home-carousel-grid {
    grid-template-columns: 1fr;
  }
}

.home-setting-page {
  gap: 0.9rem;
}

.home-setting-header h2 {
  font-size: clamp(1.4rem, 1.9vw, 2rem);
  font-weight: 600;
  line-height: 1.18;
}

.home-setting-header p:not(.home-setting-kicker),
.home-upload-panel small,
.home-setting-empty {
  font-size: 0.8125rem;
  line-height: 1.55;
}

.home-setting-kicker {
  font-size: 0.7rem;
}

.home-setting-button,
.home-setting-mini-button,
.home-upload-panel,
.home-setting-empty,
.home-carousel-card {
  border-radius: 2px;
}

.home-setting-button {
  min-height: 2.45rem;
  font-size: 0.8125rem;
}

.home-setting-mini-button {
  min-height: 1.9rem;
  font-size: 0.72rem;
}

.home-upload-panel {
  padding: 1rem;
}

.home-upload-panel span,
.home-carousel-card__body p {
  font-size: 0.875rem;
}

.home-carousel-card__body {
  padding: 0.75rem;
}

.home-carousel-card__body small,
.home-carousel-card__body code {
  font-size: 0.72rem;
}

.home-carousel-card__actions {
  padding: 0 0.75rem 0.75rem;
}
</style>

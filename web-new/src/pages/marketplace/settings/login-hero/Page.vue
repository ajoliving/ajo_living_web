<!--
 * 登入背景圖設定頁。
 * 1. 上傳最多三張登入頁左側背景圖片。
 * 2. 設定圖片來源、地點資訊與輪換順序。
-->
<script setup lang="ts">
import AppIcon from '@/shared/components/base/AppIcon.vue';

import { useLoginHeroSettingsPage } from './login-hero';

const {
  addImage,
  canAddImage,
  canSave,
  handleFileSelected,
  images,
  loadHero,
  loading,
  moveImage,
  removeImage,
  saveHero,
  saving,
  t,
} = useLoginHeroSettingsPage();
</script>

<template>
  <section class="login-hero-page">
    <div class="login-hero-header">
      <div>
        <p class="login-hero-kicker">
          {{ t('marketplace.settings.loginHeroSection') }}
        </p>
        <h2>{{ t('marketplace.settings.loginHeroTitle') }}</h2>
        <p>{{ t('marketplace.settings.loginHeroDescription') }}</p>
      </div>
      <div class="login-hero-actions">
        <button
          type="button"
          class="login-hero-button login-hero-button--secondary"
          :disabled="loading"
          @click="loadHero"
        >
          <AppIcon
            name="reload"
            :size="16"
          />
          <span>{{ t('marketplace.settings.refresh') }}</span>
        </button>
        <button
          type="button"
          class="login-hero-button login-hero-button--primary"
          :disabled="!canSave"
          @click="saveHero"
        >
          <AppIcon
            name="check-circle"
            :size="16"
          />
          <span>{{ saving ? t('marketplace.settings.saving') : t('marketplace.settings.loginHeroSave') }}</span>
        </button>
      </div>
    </div>

    <div class="login-hero-list">
      <section
        v-for="(image, index) in images"
        :key="image.key"
        class="login-hero-panel"
      >
        <div class="login-hero-preview">
          <img
            v-if="image.url"
            :src="image.url"
            :alt="t('marketplace.settings.loginHeroImageAlt', { index: index + 1 })"
          />
          <div
            v-else
            class="login-hero-placeholder"
          >
            <AppIcon
              name="picture"
              :size="28"
            />
            <span>{{ t('marketplace.settings.loginHeroImageRequired') }}</span>
          </div>
          <span
            v-if="image.uploading"
            class="login-hero-uploading"
          >
            {{ t('marketplace.settings.uploading') }}
          </span>
        </div>

        <div class="login-hero-form">
          <div class="login-hero-form__top">
            <p>{{ t('marketplace.settings.slotLabel', { index: index + 1 }) }}</p>
            <div class="login-hero-mini-actions">
              <button
                type="button"
                class="login-hero-mini-button"
                :disabled="index === 0 || image.uploading"
                @click="moveImage(image.key, -1)"
              >
                {{ t('marketplace.settings.moveUp') }}
              </button>
              <button
                type="button"
                class="login-hero-mini-button"
                :disabled="index === images.length - 1 || image.uploading"
                @click="moveImage(image.key, 1)"
              >
                {{ t('marketplace.settings.moveDown') }}
              </button>
              <button
                type="button"
                class="login-hero-mini-button"
                :disabled="images.length <= 1 || image.uploading"
                @click="removeImage(image.key)"
              >
                {{ t('marketplace.editor.removeImage') }}
              </button>
            </div>
          </div>

          <label class="login-hero-upload">
            <input
              type="file"
              accept="image/*"
              @change="handleFileSelected(image.key, $event)"
            />
            <AppIcon
              name="cloud-upload"
              :size="16"
            />
            <span>{{ t('marketplace.settings.loginHeroUpload') }}</span>
          </label>
          <p class="login-hero-spec">
            {{ t('marketplace.settings.loginHeroImageSpec') }}
          </p>

          <label class="login-hero-field">
            <span>{{ t('marketplace.settings.loginHeroAuthorField') }}</span>
            <input
              v-model="image.author"
              type="text"
              maxlength="120"
              :placeholder="t('marketplace.settings.loginHeroAuthorPlaceholder')"
            />
          </label>

          <label class="login-hero-field">
            <span>{{ t('marketplace.settings.loginHeroLocationField') }}</span>
            <input
              v-model="image.location"
              type="text"
              maxlength="160"
              :placeholder="t('marketplace.settings.loginHeroLocationPlaceholder')"
            />
          </label>

          <div class="login-hero-object">
            <span>{{ t('marketplace.settings.loginHeroObjectKey') }}</span>
            <code>{{ image.objectKey || '-' }}</code>
          </div>
        </div>
      </section>
    </div>

    <button
      v-if="canAddImage"
      type="button"
      class="login-hero-button login-hero-button--secondary login-hero-add"
      @click="addImage"
    >
      <AppIcon
        name="plus-square"
        :size="16"
      />
      <span>{{ t('marketplace.settings.loginHeroAdd') }}</span>
    </button>
  </section>
</template>

<style scoped>
.login-hero-page {
  display: grid;
  gap: 1rem;
}

.login-hero-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
}

.login-hero-kicker {
  margin: 0;
  color: rgb(var(--color-text-muted));
  font-size: 0.75rem;
  font-weight: 800;
  letter-spacing: 0.1em;
  line-height: 1;
  text-transform: uppercase;
}

.login-hero-header h2 {
  margin: 0.35rem 0 0;
  color: rgb(var(--color-text));
  font-family: var(--font-display);
  font-size: clamp(1.5rem, 2vw, 2.25rem);
  font-weight: 650;
  line-height: 1.2;
}

.login-hero-header p:not(.login-hero-kicker) {
  margin: 0.45rem 0 0;
  color: rgb(var(--color-text-muted));
  font-size: 0.9rem;
  line-height: 1.7;
}

.login-hero-actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 0.65rem;
}

.login-hero-list {
  display: grid;
  gap: 1rem;
}

.login-hero-panel {
  display: grid;
  grid-template-columns: minmax(18rem, 0.9fr) minmax(0, 1.1fr);
  overflow: hidden;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
}

.login-hero-preview {
  position: relative;
  min-height: 28rem;
  background: rgb(var(--color-border));
}

.login-hero-preview img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.login-hero-placeholder {
  display: grid;
  min-height: 28rem;
  place-items: center;
  align-content: center;
  gap: 0.55rem;
  color: rgb(var(--color-text-muted));
  font-size: 0.9rem;
  font-weight: 700;
}

.login-hero-uploading {
  position: absolute;
  inset: auto 0 0;
  background: rgb(26 28 27 / 0.76);
  padding: 0.55rem;
  color: rgb(255 255 255);
  font-size: 0.82rem;
  font-weight: 700;
  text-align: center;
}

.login-hero-form {
  display: grid;
  align-content: start;
  gap: 0.85rem;
  padding: 1rem;
}

.login-hero-form__top {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.75rem;
}

.login-hero-form__top p {
  margin: 0;
  color: rgb(var(--color-text));
  font-size: 1rem;
  font-weight: 800;
}

.login-hero-button,
.login-hero-upload,
.login-hero-mini-button {
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

.login-hero-button {
  min-height: 2.75rem;
  padding: 0 1rem;
  font-size: 0.9rem;
}

.login-hero-button--primary {
  background: rgb(var(--color-text));
  color: rgb(255 255 255);
}

.login-hero-button:disabled {
  cursor: not-allowed;
  opacity: 0.45;
}

.login-hero-upload {
  min-height: 2.35rem;
  width: fit-content;
  padding: 0 0.75rem;
  cursor: pointer;
  font-size: 0.85rem;
}

.login-hero-spec {
  margin: -0.35rem 0 0;
  color: rgb(var(--color-text-muted));
  font-size: 0.78rem;
  line-height: 1.6;
}

.login-hero-mini-actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 0.45rem;
}

.login-hero-mini-button {
  min-height: 2rem;
  padding: 0 0.65rem;
  font-size: 0.78rem;
}

.login-hero-mini-button:disabled {
  cursor: not-allowed;
  opacity: 0.45;
}

.login-hero-add {
  width: fit-content;
}

.login-hero-upload input {
  display: none;
}

.login-hero-field,
.login-hero-object {
  display: grid;
  gap: 0.45rem;
}

.login-hero-field span,
.login-hero-object span {
  color: rgb(var(--color-text-muted));
  font-size: 0.78rem;
  font-weight: 700;
}

.login-hero-field input {
  width: 100%;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
  padding: 0.75rem;
  color: rgb(var(--color-text));
  font-size: 0.92rem;
  line-height: 1.5;
  outline: 0;
}

.login-hero-object code {
  overflow: hidden;
  color: rgb(var(--color-text-muted));
  font-size: 0.76rem;
  text-overflow: ellipsis;
  white-space: nowrap;
}

@media (max-width: 1023px) {
  .login-hero-panel {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 767px) {
  .login-hero-header {
    flex-direction: column;
  }

  .login-hero-actions,
  .login-hero-button,
  .login-hero-add {
    width: 100%;
  }
}

.login-hero-page,
.login-hero-list {
  gap: 0.9rem;
}

.login-hero-header h2 {
  font-size: clamp(1.4rem, 1.9vw, 2rem);
  font-weight: 600;
  line-height: 1.18;
}

.login-hero-header p:not(.login-hero-kicker),
.login-hero-spec {
  font-size: 0.8125rem;
  line-height: 1.55;
}

.login-hero-kicker {
  font-size: 0.7rem;
}

.login-hero-panel,
.login-hero-button,
.login-hero-upload,
.login-hero-mini-button,
.login-hero-field input {
  border-radius: 2px;
}

.login-hero-form {
  padding: 0.9rem;
}

.login-hero-button {
  min-height: 2.45rem;
  font-size: 0.8125rem;
}

.login-hero-upload {
  min-height: 2.25rem;
  font-size: 0.78rem;
}

.login-hero-mini-button {
  min-height: 1.9rem;
  font-size: 0.72rem;
}

.login-hero-form__top p {
  font-size: 0.875rem;
}

.login-hero-field span,
.login-hero-object span,
.login-hero-object code {
  font-size: 0.72rem;
}

.login-hero-field input {
  padding: 0.65rem;
  font-size: 0.875rem;
}
</style>

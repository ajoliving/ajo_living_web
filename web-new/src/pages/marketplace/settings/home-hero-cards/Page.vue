<!--
 * 首頁三大圖設定頁。
 * 1. 固定管理二手、樓盤放售與服務住宅三個主模組。
 * 2. 設定 PNG 圖片、主標題、副標題與內容。
-->
<script setup lang="ts">
import AppIcon from '@/shared/components/base/AppIcon.vue';

import { useHomeHeroCardsSettingsPage } from './hero-cards';

const {
  canSave,
  handleFileSelected,
  loading,
  moduleOptions,
  orderedCards,
  saveCards,
  saving,
  t,
} = useHomeHeroCardsSettingsPage();
</script>

<template>
  <section class="home-hero-page">
    <div class="home-hero-header">
      <div>
        <p class="home-hero-kicker">
          {{ t('marketplace.settings.homeContentKicker') }}
        </p>
        <h2>{{ t('marketplace.settings.homeHeroCardsTitle') }}</h2>
        <p>{{ t('marketplace.settings.homeHeroCardsDescription') }}</p>
      </div>
      <button
        type="button"
        class="home-hero-button home-hero-button--primary"
        :disabled="!canSave"
        @click="saveCards"
      >
        <AppIcon
          name="check-circle"
          :size="16"
        />
        <span>{{ saving ? t('marketplace.settings.saving') : t('marketplace.settings.homeHeroCardsSave') }}</span>
      </button>
    </div>

    <div
      v-if="loading"
      class="home-hero-empty"
    >
      {{ t('common.status.loading') }}
    </div>

    <div
      v-else
      class="home-hero-card-grid"
    >
      <article
        v-for="(card, index) in orderedCards"
        :key="card.moduleCode"
        class="home-hero-card"
      >
        <div class="home-hero-card__media">
          <img
            v-if="card.url"
            :src="card.url"
            :alt="t(moduleOptions[index].labelKey)"
          />
          <div
            v-else
            class="home-hero-card__placeholder"
          >
            <AppIcon
              name="picture"
              :size="28"
            />
            <span>{{ t('marketplace.settings.homeHeroCardsImageRequired') }}</span>
          </div>
          <span
            v-if="card.uploading"
            class="home-hero-card__uploading"
          >
            {{ t('marketplace.settings.uploading') }}
          </span>
        </div>

        <div class="home-hero-card__body">
          <div>
            <p class="home-hero-card__module">
              {{ t(moduleOptions[index].labelKey) }}
            </p>
            <label class="home-hero-upload">
              <input
                type="file"
                accept="image/png"
                @change="handleFileSelected(card.moduleCode, $event)"
              />
              <AppIcon
                name="cloud-upload"
                :size="16"
              />
              <span>{{ t('marketplace.settings.homeHeroCardsUpload') }}</span>
            </label>
            <p class="home-hero-card__spec">
              {{ t('marketplace.settings.homeHeroCardsImageSpec') }}
            </p>
          </div>

          <label class="home-hero-field">
            <span>{{ t('marketplace.settings.homeHeroCardsTitleField') }}</span>
            <input
              v-model="card.title"
              type="text"
              maxlength="160"
              :placeholder="t('marketplace.settings.homeHeroCardsTitlePlaceholder')"
            />
          </label>

          <label class="home-hero-field">
            <span>{{ t('marketplace.settings.homeHeroCardsSubtitleField') }}</span>
            <input
              v-model="card.subtitle"
              type="text"
              maxlength="240"
              :placeholder="t('marketplace.settings.homeHeroCardsSubtitlePlaceholder')"
            />
          </label>

          <label class="home-hero-field">
            <span>{{ t('marketplace.settings.homeHeroCardsBodyField') }}</span>
            <textarea
              v-model="card.body"
              rows="4"
              maxlength="1000"
              :placeholder="t('marketplace.settings.homeHeroCardsBodyPlaceholder')"
            />
          </label>
        </div>
      </article>
    </div>
  </section>
</template>

<style scoped>
.home-hero-page {
  display: grid;
  gap: 1rem;
}

.home-hero-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
}

.home-hero-kicker {
  margin: 0;
  color: rgb(var(--color-text-muted));
  font-size: 0.75rem;
  font-weight: 800;
  letter-spacing: 0.1em;
  line-height: 1;
  text-transform: uppercase;
}

.home-hero-header h2 {
  margin: 0.35rem 0 0;
  color: rgb(var(--color-text));
  font-family: var(--font-display);
  font-size: clamp(1.5rem, 2vw, 2.25rem);
  font-weight: 650;
  line-height: 1.2;
}

.home-hero-header p:not(.home-hero-kicker) {
  margin: 0.45rem 0 0;
  color: rgb(var(--color-text-muted));
  font-size: 0.9rem;
  line-height: 1.7;
}

.home-hero-button,
.home-hero-upload {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.45rem;
  border: 1px solid rgb(var(--color-text));
  border-radius: 8px;
  font-weight: 700;
}

.home-hero-button {
  min-height: 2.75rem;
  padding: 0 1rem;
  font-size: 0.9rem;
}

.home-hero-button--primary {
  background: rgb(var(--color-text));
  color: rgb(255 255 255);
}

.home-hero-button:disabled {
  cursor: not-allowed;
  opacity: 0.45;
}

.home-hero-card-grid {
  display: grid;
  gap: 1rem;
}

.home-hero-card {
  display: grid;
  grid-template-columns: minmax(16rem, 0.85fr) minmax(0, 1.15fr);
  overflow: hidden;
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface));
}

.home-hero-card__media {
  position: relative;
  min-height: 18rem;
  background: rgb(var(--color-border));
}

.home-hero-card__media img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.home-hero-card__placeholder {
  display: grid;
  height: 100%;
  min-height: 18rem;
  place-items: center;
  align-content: center;
  gap: 0.55rem;
  color: rgb(var(--color-text-muted));
  font-size: 0.9rem;
  font-weight: 700;
}

.home-hero-card__uploading {
  position: absolute;
  inset: auto 0 0;
  background: rgb(26 28 27 / 0.76);
  padding: 0.55rem;
  color: rgb(255 255 255);
  font-size: 0.82rem;
  font-weight: 700;
  text-align: center;
}

.home-hero-card__body {
  display: grid;
  align-content: start;
  gap: 0.85rem;
  padding: 1rem;
}

.home-hero-card__module {
  margin: 0 0 0.6rem;
  color: rgb(var(--color-text));
  font-size: 1rem;
  font-weight: 800;
}

.home-hero-card__spec {
  margin: 0.45rem 0 0;
  color: rgb(var(--color-text-muted));
  font-size: 0.78rem;
  line-height: 1.6;
}

.home-hero-upload {
  min-height: 2.35rem;
  padding: 0 0.75rem;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-text));
  cursor: pointer;
  font-size: 0.85rem;
}

.home-hero-upload input {
  display: none;
}

.home-hero-field {
  display: grid;
  gap: 0.45rem;
}

.home-hero-field span {
  color: rgb(var(--color-text-muted));
  font-size: 0.78rem;
  font-weight: 700;
}

.home-hero-field input,
.home-hero-field textarea {
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

.home-hero-field textarea {
  resize: vertical;
}

.home-hero-empty {
  border: 1px solid rgb(var(--color-border));
  border-radius: 8px;
  background: rgb(var(--color-surface-raised));
  padding: 1rem;
  color: rgb(var(--color-text-muted));
  font-size: 0.9rem;
}

@media (max-width: 900px) {
  .home-hero-card {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 767px) {
  .home-hero-header {
    flex-direction: column;
  }

  .home-hero-button {
    width: 100%;
  }
}
</style>

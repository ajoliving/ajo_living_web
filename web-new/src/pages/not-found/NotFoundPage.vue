<!--
 * 404 頁面。
 * 1. 提供基礎未找到頁面。
 * 2. 導回首頁或二手列表頁。
-->
<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { RouterLink } from 'vue-router';

import AppIcon from '@/shared/components/base/AppIcon.vue';
import BaseButton from '@/shared/components/base/BaseButton.vue';

const { t } = useI18n();
</script>

<template>
  <main class="not-found-page">
    <section class="not-found-panel">
      <p class="not-found-kicker">
        404
      </p>
      <h1>
        {{ t('common.notFound.title') }}
      </h1>
      <p class="not-found-description">
        {{ t('common.notFound.description') }}
      </p>

      <div class="not-found-actions">
        <RouterLink to="/">
          <BaseButton
            variant="primary"
            size="lg"
          >
            <template #leading>
              <AppIcon
                name="home"
                :size="18"
              />
            </template>
            {{ t('common.notFound.goHome') }}
          </BaseButton>
        </RouterLink>
      </div>
    </section>
  </main>
</template>

<style scoped>
.not-found-page {
  display: grid;
  width: 100%;
  min-height: calc(100vh - var(--app-header-offset, 0rem));
  place-items: center;
  padding: 4rem var(--layout-page-padding-inline);
  color: rgb(var(--color-text));
}

.not-found-panel {
  display: grid;
  width: min(100%, 42rem);
  justify-items: center;
  border: 1px solid rgb(var(--color-border));
  border-radius: 2px;
  background: rgb(var(--color-surface));
  padding: 3rem 2rem;
  text-align: center;
}

.not-found-kicker {
  margin: 0;
  color: rgb(var(--color-text-muted));
  font-size: 0.75rem;
  font-weight: 800;
  letter-spacing: 0.18em;
  line-height: 1;
  text-transform: uppercase;
}

.not-found-panel h1 {
  margin: 1rem 0 0;
  color: rgb(var(--color-text));
  font-family: var(--font-display);
  font-size: clamp(2rem, 4vw, 3rem);
  font-weight: 500;
  line-height: 1.14;
}

.not-found-description {
  max-width: 32rem;
  margin: 1rem 0 0;
  color: rgb(var(--color-text-muted));
  font-size: 0.9rem;
  line-height: 1.7;
}

.not-found-actions {
  margin-top: 1.5rem;
}

@media (max-width: 767px) {
  .not-found-page {
    padding: 2rem var(--layout-page-padding-inline) 4rem;
  }

  .not-found-panel {
    padding: 2rem 1rem;
  }
}
</style>

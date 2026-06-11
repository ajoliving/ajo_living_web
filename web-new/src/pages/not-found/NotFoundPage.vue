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
  <div class="section-shell py-20">
    <div class="panel-surface px-6 py-14 text-center sm:px-10">
      <p class="text-kicker">
        404
      </p>
      <h1 class="display-title mt-4 text-5xl text-text sm:text-6xl">
        {{ t('common.notFound.title') }}
      </h1>
      <p class="mx-auto mt-4 max-w-2xl text-base leading-8 text-text-muted">
        {{ t('common.notFound.description') }}
      </p>

      <div class="mt-8 flex justify-center">
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
    </div>
  </div>
</template>

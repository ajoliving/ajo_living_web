<!--
 * 應用程式根元件。
 * 1. 同步全站主題與語系偏好。
 * 2. 載入應用程式外框與全域提示層。
-->
<script setup lang="ts">
import { watch } from 'vue';
import { useRoute } from 'vue-router';

import AppShell from '@/shared/components/layout/AppShell.vue';
import { applyLocale } from '@/i18n';
import { updateDocumentTitle } from '@/router';
import { usePreferenceStore } from '@/stores/preferences';
import { useSessionStore } from '@/stores/session';

const preferenceStore = usePreferenceStore();
const sessionStore = useSessionStore();
const route = useRoute();

// 1. 啟動時還原工作階段
void sessionStore.hydrateSession();

// 2. 將 Pinia 語系同步回 vue-i18n 與文件標題
watch(
  () => preferenceStore.locale,
  (locale) => {
    applyLocale(locale);
    updateDocumentTitle(typeof route.meta.titleKey === 'string' ? route.meta.titleKey : undefined);
  },
  { immediate: true },
);
</script>

<template>
  <AppShell />
</template>

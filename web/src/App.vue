<!--
 * 應用程式根元件。
 * 1. 同步全站主題與語系偏好。
 * 2. 載入應用程式外框與全域提示層。
-->
<script setup lang="ts">
import { watch } from 'vue';

import AppShell from '@/shared/components/layout/AppShell.vue';
import i18n from '@/i18n';
import { usePreferenceStore } from '@/stores/preferences';
import { useSessionStore } from '@/stores/session';

const preferenceStore = usePreferenceStore();
const sessionStore = useSessionStore();

// 1. 啟動時還原主題與語系偏好
preferenceStore.hydratePreferences();
void sessionStore.hydrateSession();

// 2. 將 Pinia 語系同步回 vue-i18n
watch(
  () => preferenceStore.locale,
  (locale) => {
    i18n.global.locale.value = locale;
    document.documentElement.lang = locale;
  },
  { immediate: true },
);
</script>

<template>
  <AppShell />
</template>

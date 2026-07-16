/*
 * 前端應用程式入口。
 * 1. 掛載 Pinia、Router 與 i18n。
 * 2. 載入全域樣式與設計 token。
 */
import { createApp } from 'vue';

import App from '@/App.vue';
import { AUTH_SESSION_EXPIRED_EVENT } from '@/httpapis/auth-session';
import i18n, { applyLocale } from '@/i18n';
import { pinia } from '@/pinia';
import router from '@/router';
import { usePreferenceStore } from '@/stores/preferences';
import { useSessionStore } from '@/stores/session';

import '@/styles/index.css';

// 1. 在路由啟動前還原語系，避免首個文件標題與 HTML lang 使用錯誤語系
const preferenceStore = usePreferenceStore(pinia);
preferenceStore.hydratePreferences();
applyLocale(preferenceStore.locale);

// 2. 建立並掛載 Vue 應用程式
const app = createApp(App);

app.use(pinia);
app.use(router);
app.use(i18n);
app.mount('#app');

// 3. 監聽 HTTP 層登入失效事件並同步 UI 登入狀態
if (typeof window !== 'undefined') {
  window.addEventListener(AUTH_SESSION_EXPIRED_EVENT, () => {
    const sessionStore = useSessionStore(pinia);
    sessionStore.clearSession();

    if (router.currentRoute.value.meta.requiresAuth) {
      void router.push({
        path: '/login',
        query: { redirect: router.currentRoute.value.fullPath },
      });
    }
  });
}

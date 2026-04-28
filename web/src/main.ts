/*
 * 前端應用程式入口。
 * 1. 掛載 Pinia、Router 與 i18n。
 * 2. 載入全域樣式與設計 token。
 */
import { createApp } from 'vue';

import App from '@/App.vue';
import i18n from '@/i18n';
import { pinia } from '@/pinia';
import router from '@/router';

import '@/styles/index.css';

// 1. 建立並掛載 Vue 應用程式
const app = createApp(App);

app.use(pinia);
app.use(router);
app.use(i18n);
app.mount('#app');

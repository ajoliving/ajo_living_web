/*
 * 首頁路由。
 * 1. 定義首頁主路由。
 * 2. 對外輸出首頁路由陣列供總路由組裝。
 */
import type { RouteRecordRaw } from 'vue-router';

import HomePage from '@/pages/home/HomePage.vue';

// 1. 輸出首頁路由
export const homeRoutes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'Home',
    component: HomePage,
    meta: { titleKey: 'nav.home', headerMode: 'transparent' },
  },
];

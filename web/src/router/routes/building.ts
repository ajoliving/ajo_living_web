/*
 * 樓盤放售路由。
 * 1. 定義樓盤放售主路由。
 * 2. 對外輸出樓盤放售路由陣列供總路由組裝。
 */
import type { RouteRecordRaw } from 'vue-router';

import BuildingPage from '@/pages/building/BuildingPage.vue';

// 1. 輸出樓盤放售路由
export const buildingRoutes: RouteRecordRaw[] = [
  {
    path: '/properties',
    name: 'Properties',
    component: BuildingPage,
    meta: { titleKey: 'nav.properties' },
  },
];

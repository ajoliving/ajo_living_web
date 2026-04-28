/*
 * 服務住宅路由。
 * 1. 定義服務住宅主路由。
 * 2. 對外輸出服務住宅路由陣列供總路由組裝。
 */
import type { RouteRecordRaw } from 'vue-router';

import FirstPage from '@/pages/first/FirstPage.vue';

// 1. 輸出服務住宅路由
export const firstRoutes: RouteRecordRaw[] = [
  {
    path: '/serviced-residences',
    name: 'ServicedResidences',
    component: FirstPage,
    meta: { titleKey: 'nav.servicedResidences' },
  },
];

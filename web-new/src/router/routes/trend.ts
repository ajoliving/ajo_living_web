/*
 * 走勢頁路由。
 * 1. 定義走勢主路由。
 * 2. 對外輸出走勢路由陣列供總路由組裝。
 */
import type { RouteRecordRaw } from 'vue-router';

import TrendPage from '@/pages/trend/TrendPage.vue';

// 1. 輸出走勢路由
export const trendRoutes: RouteRecordRaw[] = [
  {
    path: '/trend',
    name: 'Trend',
    component: TrendPage,
    meta: { titleKey: 'nav.trend' },
  },
  {
    path: '/trend/district/:district',
    redirect: '/trend',
  },
];

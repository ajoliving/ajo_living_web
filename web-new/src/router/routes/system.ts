/*
 * 系統兜底路由。
 * 1. 提供 404 頁面路由。
 * 2. 對外輸出兜底路由供總路由組裝。
 */
import type { RouteRecordRaw } from 'vue-router';

const NotFoundPage = () => import('@/pages/not-found/NotFoundPage.vue');

// 1. 輸出系統兜底路由
export const systemRoutes: RouteRecordRaw[] = [
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: NotFoundPage,
    meta: { titleKey: 'common.action.retry' },
  },
];

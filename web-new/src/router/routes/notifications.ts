/*
 * 通知中心路由。
 * 1. 定義平台級通知中心入口。
 * 2. 保留舊 notif 路徑兼容跳轉。
 */
import type { RouteRecordRaw } from 'vue-router';

const NotificationsPage = () => import('@/pages/notifications/Page.vue');

// 1. 輸出通知中心路由
export const notificationRoutes: RouteRecordRaw[] = [
  {
    path: '/notif',
    redirect: '/notifications',
  },
  {
    path: '/notifications',
    name: 'Notifications',
    component: NotificationsPage,
    meta: { titleKey: 'nav.notifications', requiresAuth: true },
  },
];

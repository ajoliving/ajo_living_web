/*
 * 通知中心路由。
 * 1. 定義會員通知中心主入口。
 * 2. 承載全部通知、優惠提醒與已讀操作 UI。
 */
import type { RouteRecordRaw } from 'vue-router';

import NotificationsPage from '@/pages/notifications/Page.vue';

// 1. 輸出通知中心路由
export const notificationRoutes: RouteRecordRaw[] = [
  {
    path: '/notifications',
    name: 'Notifications',
    component: NotificationsPage,
    meta: { titleKey: 'nav.notifications' },
  },
];

/*
 * 通訊路由。
 * 1. 保留站內信、通知與公告模組路由邊界。
 * 2. 會員中心內嵌頁面暫不建立獨立公開路由。
 */
import type { RouteRecordRaw } from 'vue-router';

import CommunicationsPage from '@/pages/communications/Page.vue';
import CommunicationsMessagesPage from '@/pages/communications/messages/Page.vue';
import CommunicationsNoticesPage from '@/pages/communications/notices/Page.vue';

// 1. 輸出通訊路由
export const communicationsRoutes: RouteRecordRaw[] = [
  {
    path: '/communications',
    name: 'Communications',
    component: CommunicationsPage,
    meta: { titleKey: 'nav.communications', requiresAuth: true },
  },
  {
    path: '/communications/notices',
    name: 'CommunicationsNotices',
    component: CommunicationsNoticesPage,
    meta: { titleKey: 'nav.communicationsNotices', requiresAuth: true },
  },
  {
    path: '/communications/messages',
    name: 'CommunicationsMessages',
    component: CommunicationsMessagesPage,
    meta: { titleKey: 'nav.communicationsMessages', requiresAuth: true },
  },
];

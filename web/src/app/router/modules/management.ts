/*
 * 管理端路由。
 * 1. 定義 AJO 管理端與內容設定入口。
 * 2. 與 marketplace 頁面邊界分離。
 */
import type { RouteRecordRaw } from 'vue-router';

import ManagementIntegrationsPage from '@/pages/management/integrations/Page.vue';
import ManagementPage from '@/pages/management/Page.vue';
import ManagementContentPage from '@/pages/management/content/Page.vue';
import ManagementUsersPage from '@/pages/management/users/Page.vue';

// 1. 輸出管理端路由
export const managementRoutes: RouteRecordRaw[] = [
  {
    path: '/settings',
    name: 'Settings',
    component: ManagementContentPage,
    meta: { titleKey: 'nav.settings', requiresAuth: true, requiresStaff: true },
  },
  {
    path: '/management',
    name: 'Management',
    component: ManagementPage,
    meta: { titleKey: 'nav.management', requiresAuth: true, requiresStaff: true },
  },
  {
    path: '/management/users',
    name: 'ManagementUsers',
    component: ManagementUsersPage,
    meta: { titleKey: 'nav.management', requiresAuth: true, requiresStaff: true },
  },
  {
    path: '/management/content',
    name: 'ManagementContent',
    component: ManagementContentPage,
    meta: { titleKey: 'nav.management', requiresAuth: true, requiresStaff: true },
  },
  {
    path: '/management/integrations',
    name: 'ManagementIntegrations',
    component: ManagementIntegrationsPage,
    meta: { titleKey: 'nav.management', requiresAuth: true, requiresStaff: true },
  },
];

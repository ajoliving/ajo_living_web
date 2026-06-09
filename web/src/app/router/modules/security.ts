/*
 * 安防路由。
 * 1. 保留 CCTV、門禁、智能門鎖與對講模組路由邊界。
 * 2. 後續接入舊系統時再落實具體頁面。
 */
import type { RouteRecordRaw } from 'vue-router';

import SecurityPage from '@/pages/security/Page.vue';
import SecurityAccessControlPage from '@/pages/security/access-control/Page.vue';
import SecurityCctvPage from '@/pages/security/cctv/Page.vue';
import SecurityIntercomPage from '@/pages/security/intercom/Page.vue';

// 1. 輸出安防路由
export const securityRoutes: RouteRecordRaw[] = [
  {
    path: '/security',
    name: 'Security',
    component: SecurityPage,
    meta: { titleKey: 'nav.security', requiresAuth: true },
  },
  {
    path: '/security/cctv',
    name: 'SecurityCctv',
    component: SecurityCctvPage,
    meta: { titleKey: 'nav.securityCctv', requiresAuth: true },
  },
  {
    path: '/security/access-control',
    name: 'SecurityAccessControl',
    component: SecurityAccessControlPage,
    meta: { titleKey: 'nav.securityAccessControl', requiresAuth: true },
  },
  {
    path: '/security/intercom',
    name: 'SecurityIntercom',
    component: SecurityIntercomPage,
    meta: { titleKey: 'nav.securityIntercom', requiresAuth: true },
  },
];

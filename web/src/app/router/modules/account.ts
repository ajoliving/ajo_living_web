/*
 * 帳戶路由。
 * 1. 定義登入與會員中心入口。
 * 2. 保持帳戶頁面路由集中管理。
 */
import type { RouteRecordRaw } from 'vue-router';

import LoginPage from '@/pages/account/login/Page.vue';
import MemberPage from '@/pages/member/Page.vue';

// 1. 輸出帳戶路由
export const accountRoutes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: LoginPage,
    meta: { titleKey: 'nav.login' },
  },
  {
    path: '/member',
    name: 'Member',
    component: MemberPage,
    meta: { titleKey: 'nav.member', requiresAuth: true },
  },
];

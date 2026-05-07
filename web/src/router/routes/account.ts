/*
 * 會員中心路由。
 * 1. 定義登入頁、會員資料頁與通知兼容跳轉路由。
 * 2. 對外輸出會員中心路由陣列供總路由組裝。
 */
import type { RouteRecordRaw } from 'vue-router';

import LoginPage from '@/pages/account/login/LoginPage.vue';
import AccountMyPage from '@/pages/account/my/AccountMyPage.vue';
import AccountNotificationsPage from '@/pages/account/my/notifications/AccountNotificationsPage.vue';
import AccountProfilePage from '@/pages/account/my/profile/AccountProfilePage.vue';

// 1. 輸出會員中心路由
export const accountRoutes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: LoginPage,
    meta: { titleKey: 'nav.login' },
  },
  {
    path: '/account',
    component: AccountMyPage,
    meta: { requiresAuth: true },
    children: [
      {
        path: '',
        redirect: '/account/profile',
      },
      {
        path: 'profile',
        name: 'AccountProfile',
        component: AccountProfilePage,
        meta: { titleKey: 'account.profile.title', requiresAuth: true },
      },
      {
        path: 'notifications',
        name: 'AccountNotificationsRedirect',
        component: AccountNotificationsPage,
        meta: { titleKey: 'chat.systemNoticeTitle', requiresAuth: true },
      },
    ],
  },
];

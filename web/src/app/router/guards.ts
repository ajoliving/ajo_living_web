/*
 * 路由守衛設定。
 * 1. 集中處理登入態與 Staff 權限守衛。
 * 2. 集中同步文件標題。
 */
import type { Router } from 'vue-router';

import i18n from '@/app/i18n';
import { pinia } from '@/app/stores';
import { useSessionStore } from '@/app/stores/session';

// 1. 掛載全域路由守衛
export const setupRouterGuards = (router: Router): void => {
  router.beforeEach(async (to) => {
    const sessionStore = useSessionStore(pinia);

    if (to.meta.requiresAuth && !sessionStore.isAuthenticated) {
      return {
        path: '/login',
        query: { redirect: to.fullPath },
      };
    }

    if (to.meta.requiresAuth && sessionStore.isAuthenticated && !sessionStore.me) {
      try {
        await sessionStore.loadCurrentUser();
      } catch {
        sessionStore.clearSession();
        return {
          path: '/login',
          query: { redirect: to.fullPath },
        };
      }
    }

    if (to.meta.requiresStaff && !sessionStore.currentUser.is_staff) {
      return { path: '/member' };
    }

    return true;
  });

  router.afterEach((to) => {
    const titleKey =
      typeof to.meta.titleKey === 'string' ? to.meta.titleKey : 'common.brand.name';

    document.title = `${i18n.global.t(titleKey)} | AJO Living`;
  });
};

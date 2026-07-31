/*
 * 路由設定。
 * 1. 組裝各業務模組輸出的路由陣列。
 * 2. 集中處理登入守衛與文件標題。
 */
import { createRouter, createWebHistory } from 'vue-router';

import i18n from '@/i18n';
import { accountRoutes } from '@/router/routes/account';
import { buildingRoutes } from '@/router/routes/building';
import { channelRoutes } from '@/router/routes/channels';
import { firstRoutes } from '@/router/routes/first';
import { homeRoutes } from '@/router/routes/home';
import { marketplaceRoutes } from '@/router/routes/marketplace';
import { notificationRoutes } from '@/router/routes/notifications';
import { paymentRoutes } from '@/router/routes/payments';
import { systemRoutes } from '@/router/routes/system';
import { trendRoutes } from '@/router/routes/trend';
import { pinia } from '@/pinia';
import { useSessionStore } from '@/stores/session';

// 1. 解析代理帳戶受限頁面跳轉
export const resolveAgencyAccountRedirect = (
  accountType: string,
  memberStatus: string,
  targetPath: string,
  targetRequiresAuth: boolean,
): string | null => {
  const agencyProfilePath = '/account/profile/agency-profile';
  const isAgencyOwner = ['individual_agent', 'agency_company'].includes(accountType);
  if (
    isAgencyOwner
    && targetRequiresAuth
    && ['pending_profile', 'pending_review', 'rejected'].includes(memberStatus)
    && targetPath !== agencyProfilePath
  ) {
    return agencyProfilePath;
  }
  if (targetPath === agencyProfilePath && !isAgencyOwner) {
    return '/account/profile';
  }
  return null;
};

// 2. 組裝模組路由
const router = createRouter({
  history: createWebHistory(),
  routes: [
    ...homeRoutes,
    ...buildingRoutes,
    ...firstRoutes,
    ...channelRoutes,
    ...trendRoutes,
    ...notificationRoutes,
    ...paymentRoutes,
    ...marketplaceRoutes,
    ...accountRoutes,
    ...systemRoutes,
  ],
  scrollBehavior() {
    return { top: 0 };
  },
});

// 2. 統一處理登入與管理端守衛
router.beforeEach(async (to) => {
  const sessionStore = useSessionStore(pinia);

  if (!sessionStore.isLoaded) {
    await sessionStore.hydrateSession();
  }

  if (to.meta.requiresAuth && !sessionStore.isAuthenticated) {
    return {
      path: '/login',
      query: { redirect: to.fullPath },
    };
  }

  const agencyRedirect = sessionStore.isAuthenticated
    ? resolveAgencyAccountRedirect(
      sessionStore.me?.account_type ?? '',
      sessionStore.me?.member_status ?? '',
      to.path,
      Boolean(to.meta.requiresAuth),
    )
    : null;
  if (agencyRedirect) {
    return { path: agencyRedirect };
  }

  if (to.meta.requiresStaff && !sessionStore.currentUser.is_staff) {
    return { path: '/profile' };
  }

  if (to.path === '/login' && sessionStore.isAuthenticated) {
    const restrictedRedirect = resolveAgencyAccountRedirect(
      sessionStore.me?.account_type ?? '', sessionStore.me?.member_status ?? '', '/',
      true,
    );
    if (restrictedRedirect) {
      return { path: restrictedRedirect };
    }
    const redirect = typeof to.query.redirect === 'string' ? to.query.redirect : '/';
    return redirect === '/login' ? '/' : redirect;
  }

  return true;
});

// 3. 統一同步文件標題，供切頁與語系切換共用
export const updateDocumentTitle = (titleKey?: string): void => {
  const resolvedTitleKey = titleKey || 'common.brand.name';
  document.title = `${i18n.global.t(resolvedTitleKey)} | AJO Living`;
};

// 4. 切頁時同步更新文件標題
router.afterEach((to) => {
  updateDocumentTitle(typeof to.meta.titleKey === 'string' ? to.meta.titleKey : undefined);
});

export default router;

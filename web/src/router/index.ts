/*
 * 路由設定。
 * 1. 組裝各業務模組輸出的路由陣列。
 * 2. 集中處理登入守衛與文件標題。
 */
import { createRouter, createWebHistory } from 'vue-router';

import i18n from '@/i18n';
import { accountRoutes } from '@/router/routes/account';
import { buildingRoutes } from '@/router/routes/building';
import { firstRoutes } from '@/router/routes/first';
import { homeRoutes } from '@/router/routes/home';
import { marketplaceRoutes } from '@/router/routes/marketplace';
import { systemRoutes } from '@/router/routes/system';
import { pinia } from '@/pinia';
import { useSessionStore } from '@/stores/session';

// 1. 組裝模組路由
const router = createRouter({
  history: createWebHistory(),
  routes: [
    ...homeRoutes,
    ...buildingRoutes,
    ...firstRoutes,
    ...marketplaceRoutes,
    ...accountRoutes,
    ...systemRoutes,
  ],
  scrollBehavior() {
    return { top: 0 };
  },
});

// 2. 集中處理需登入頁面的導航守衛
router.beforeEach((to) => {
  const sessionStore = useSessionStore(pinia);

  if (to.meta.requiresAuth && !sessionStore.isAuthenticated) {
    return {
      path: '/login',
      query: { redirect: to.fullPath },
    };
  }

  return true;
});

// 3. 切頁時同步更新文件標題
router.afterEach((to) => {
  const titleKey =
    typeof to.meta.titleKey === 'string' ? to.meta.titleKey : 'common.brand.name';

  document.title = `${i18n.global.t(titleKey)} | AJO Living`;
});

export default router;

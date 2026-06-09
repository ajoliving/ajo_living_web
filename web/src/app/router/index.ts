/*
 * 路由設定。
 * 1. 組裝各業務模組路由。
 * 2. 掛載全域導航守衛。
 */
import { createRouter, createWebHistory } from 'vue-router';

import { setupRouterGuards } from '@/app/router/guards';
import { accountRoutes } from '@/app/router/modules/account';
import { propertyRoutes } from '@/app/router/modules/properties';
import { communicationsRoutes } from '@/app/router/modules/communications';
import { homeRoutes } from '@/app/router/modules/home';
import { managementRoutes } from '@/app/router/modules/management';
import { furnitureRoutes } from '@/app/router/modules/furniture';
import { paymentsRoutes } from '@/app/router/modules/payments';
import { securityRoutes } from '@/app/router/modules/security';
import { supermarketOfferRoutes } from '@/app/router/modules/supermarket-offers';
import { systemRoutes } from '@/app/router/modules/system';

// 1. 組裝模組路由
const router = createRouter({
  history: createWebHistory(),
  routes: [
    ...homeRoutes,
    ...propertyRoutes,
    ...furnitureRoutes,
    ...supermarketOfferRoutes,
    ...accountRoutes,
    ...communicationsRoutes,
    ...paymentsRoutes,
    ...securityRoutes,
    ...managementRoutes,
    ...systemRoutes,
  ],
  scrollBehavior() {
    return { top: 0 };
  },
});

// 2. 掛載全域守衛
setupRouterGuards(router);

export default router;

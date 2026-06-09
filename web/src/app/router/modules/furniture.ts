/*
 * 二手傢俬路由。
 * 1. 定義家具公開列表、詳情與賣家路由。
 * 2. 取消舊 marketplace discover/filter 公開入口。
 */
import type { RouteRecordRaw } from 'vue-router';

import FurniturePage from '@/pages/furniture/Page.vue';
import FurnitureChatPage from '@/pages/furniture/chat/Page.vue';
import FurnitureListingPage from '@/pages/furniture/listing/Page.vue';
import FurnitureSellerPage from '@/pages/furniture/seller/Page.vue';

// 1. 輸出家具公開路由
export const furnitureRoutes: RouteRecordRaw[] = [
  {
    path: '/furniture',
    name: 'Furniture',
    component: FurniturePage,
    meta: { titleKey: 'nav.furniture' },
  },
  {
    path: '/furniture/listing/:listingId',
    name: 'FurnitureListingDetail',
    component: FurnitureListingPage,
    meta: { titleKey: 'marketplace.detail.title' },
  },
  {
    path: '/furniture/chat',
    name: 'FurnitureChat',
    component: FurnitureChatPage,
    meta: { titleKey: 'nav.communicationsMessages', requiresAuth: true },
  },
  {
    path: '/furniture/seller/:sellerId',
    name: 'FurnitureSeller',
    component: FurnitureSellerPage,
    meta: { titleKey: 'marketplace.seller.title' },
  },
];

/*
 * 新前台頻道路由。
 * 1. 定義家具與超市優惠入口。
 * 2. 定義超市優惠商品詳情路由。
 */
import type { RouteRecordRaw } from 'vue-router';

import FurniturePage from '@/pages/furniture/Page.vue';
import SupermarketOffersPage from '@/pages/offers/Page.vue';

// 1. 輸出新前台頻道路由
export const channelRoutes: RouteRecordRaw[] = [
  {
    path: '/furniture',
    name: 'Furniture',
    component: FurniturePage,
    meta: { titleKey: 'nav.furniture' },
  },
  {
    path: '/supermarket-offers',
    name: 'SupermarketOffers',
    component: SupermarketOffersPage,
    meta: { titleKey: 'nav.supermarketOffers' },
  },
  {
    path: '/supermarket-offers/products/:code',
    name: 'SupermarketOfferDetail',
    component: SupermarketOffersPage,
    meta: { titleKey: 'nav.supermarketOffers' },
  },
];

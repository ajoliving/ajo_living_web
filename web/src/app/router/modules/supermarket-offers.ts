/*
 * 超市優惠路由。
 * 1. 定義公開超市優惠主入口。
 * 2. 商品詳情由主入口內的商品詳情分頁承接。
 */
import type { RouteRecordRaw } from 'vue-router';

import SupermarketOffersPage from '@/pages/supermarket-offers/Page.vue';

// 1. 輸出超市優惠公開路由
export const supermarketOfferRoutes: RouteRecordRaw[] = [
  {
    path: '/supermarket-offers',
    name: 'SupermarketOffers',
    component: SupermarketOffersPage,
    meta: { titleKey: 'nav.supermarketOffers' },
  },
];

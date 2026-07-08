/*
 * 新前台頻道路由。
 * 1. 定義家具與超市優惠入口。
 * 2. 定義超市優惠商品詳情路由。
 */
import type { RouteRecordRaw } from 'vue-router';

const FurnitureDetailPage = () => import('@/pages/furniture/detail/FurnitureDetailPage.vue');
const FurniturePage = () => import('@/pages/furniture/Page.vue');
const SupermarketOfferDetailPage = () => import('@/pages/offers/detail/Page.vue');
const SupermarketOffersPage = () => import('@/pages/offers/Page.vue');

// 1. 輸出新前台頻道路由
export const channelRoutes: RouteRecordRaw[] = [
  {
    path: '/market',
    redirect: '/furniture',
  },
  {
    path: '/market-detail',
    redirect: '/furniture/1',
  },
  {
    path: '/furniture',
    name: 'Furniture',
    component: FurniturePage,
    meta: { titleKey: 'nav.furniture' },
  },
  {
    path: '/furniture/listing/:listingId',
    redirect: (to) => ({
      path: `/furniture/${String(to.params.listingId)}`,
      query: to.query,
      hash: to.hash,
    }),
  },
  {
    path: '/furniture/:listingId',
    name: 'FurnitureDetail',
    component: FurnitureDetailPage,
    meta: { titleKey: 'nav.furniture' },
  },
  {
    path: '/supermarket-offers',
    name: 'SupermarketOffers',
    component: SupermarketOffersPage,
    meta: { titleKey: 'nav.supermarketOffers' },
  },
  {
    path: '/offers',
    redirect: '/supermarket-offers',
  },
  {
    path: '/supermarket-offers/products/:code',
    name: 'SupermarketOfferDetail',
    component: SupermarketOfferDetailPage,
    meta: { titleKey: 'nav.supermarketOffers' },
  },
];

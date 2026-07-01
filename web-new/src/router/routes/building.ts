/*
 * 樓盤放售路由。
 * 1. 定義樓盤放售主路由。
 * 2. 對外輸出樓盤放售路由陣列供總路由組裝。
 */
import type { RouteRecordRaw } from 'vue-router';

import BuildingPage from '@/pages/building/BuildingPage.vue';
import PropertyDetailPage from '@/pages/property/detail/PropertyDetailPage.vue';
import PropertyListPage from '@/pages/property/list/PropertyListPage.vue';

// 1. 輸出樓盤放售路由
export const buildingRoutes: RouteRecordRaw[] = [
  {
    path: '/affairs',
    redirect: '/building',
  },
  {
    path: '/building',
    name: 'Building',
    component: BuildingPage,
    meta: { titleKey: 'nav.building' },
  },
  {
    path: '/building/:section',
    redirect: '/building',
  },
  {
    path: '/listing',
    redirect: '/properties',
  },
  {
    path: '/detail',
    redirect: '/properties/1',
  },
  {
    path: '/properties',
    name: 'Properties',
    component: PropertyListPage,
    props: { channel: 'sale' },
    meta: { titleKey: 'nav.properties' },
  },
  {
    path: '/properties/my',
    redirect: '/account/properties/sale',
  },
  {
    path: '/properties/my/new',
    redirect: '/account/properties/sale/new',
  },
  {
    path: '/properties/my/editor/:listingId?',
    redirect: (to) => ({
      path: to.params.listingId
        ? `/account/properties/sale/editor/${String(to.params.listingId)}`
        : '/account/properties/sale/new',
      query: to.query,
      hash: to.hash,
    }),
  },
  {
    path: '/properties/:listingId',
    name: 'PropertySaleDetail',
    component: PropertyDetailPage,
    props: { channel: 'sale' },
    meta: { titleKey: 'property.sale.detailTitle' },
  },
];

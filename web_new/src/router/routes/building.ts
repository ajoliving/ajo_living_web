/*
 * 我的大廈與樓盤放售路由。
 * 1. 定義我的大廈主路由。
 * 2. 保留樓盤放售相關公開路由。
 */
import type { RouteRecordRaw } from 'vue-router';

const BuildingPage = () => import('@/pages/building/BuildingPage.vue');
const PropertyDetailPage = () => import('@/pages/property/detail/PropertyDetailPage.vue');
const PropertyListPage = () => import('@/pages/property/list/PropertyListPage.vue');

// 1. 輸出我的大廈與樓盤放售路由
export const buildingRoutes: RouteRecordRaw[] = [
  {
    path: '/affairs',
    redirect: '/building',
  },
  {
    path: '/building',
    name: 'Building',
    component: BuildingPage,
    meta: { titleKey: 'nav.building', requiresAuth: true },
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
    redirect: '/properties',
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

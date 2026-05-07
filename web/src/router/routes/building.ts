/*
 * 樓盤放售路由。
 * 1. 定義樓盤放售主路由。
 * 2. 對外輸出樓盤放售路由陣列供總路由組裝。
 */
import type { RouteRecordRaw } from 'vue-router';

import PropertyDetailPage from '@/pages/property/detail/PropertyDetailPage.vue';
import PropertyEditorPage from '@/pages/property/editor/PropertyEditorPage.vue';
import PropertyListPage from '@/pages/property/list/PropertyListPage.vue';
import PropertyMyPage from '@/pages/property/my/PropertyMyPage.vue';

// 1. 輸出樓盤放售路由
export const buildingRoutes: RouteRecordRaw[] = [
  {
    path: '/properties',
    name: 'Properties',
    component: PropertyListPage,
    props: { channel: 'sale' },
    meta: { titleKey: 'nav.properties' },
  },
  {
    path: '/properties/my',
    name: 'PropertySaleMy',
    component: PropertyMyPage,
    props: { channel: 'sale' },
    meta: { titleKey: 'property.sale.myTitle', requiresAuth: true },
  },
  {
    path: '/properties/my/new',
    name: 'PropertySalePublish',
    component: PropertyEditorPage,
    props: { channel: 'sale' },
    meta: { titleKey: 'property.sale.publishTitle', requiresAuth: true },
  },
  {
    path: '/properties/my/editor/:listingId?',
    name: 'PropertySaleEditor',
    component: PropertyEditorPage,
    props: { channel: 'sale' },
    meta: { titleKey: 'property.sale.publishTitle', requiresAuth: true },
  },
  {
    path: '/properties/:listingId',
    name: 'PropertySaleDetail',
    component: PropertyDetailPage,
    props: { channel: 'sale' },
    meta: { titleKey: 'property.sale.detailTitle' },
  },
];

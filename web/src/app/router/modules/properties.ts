/*
 * 樓盤與服務式住宅路由。
 * 1. 定義樓盤租售與服務式住宅主路由。
 * 2. 保留大廈相關會員上下文接入邊界。
 */
import type { RouteRecordRaw } from 'vue-router';

import BuildingPage from '@/pages/building/Page.vue';
import BuildingDetailPage from '@/pages/building/detail/Page.vue';
import PropertyDetailPage from '@/pages/properties/detail/Page.vue';
import PropertyListPage from '@/pages/properties/Page.vue';
import ServicedResidenceDetailPage from '@/pages/serviced-residences/detail/Page.vue';
import ServicedResidencesPage from '@/pages/serviced-residences/Page.vue';

// 1. 輸出樓盤與服務式住宅路由
export const propertyRoutes: RouteRecordRaw[] = [
  {
    path: '/building',
    name: 'Building',
    component: BuildingPage,
    meta: { titleKey: 'nav.building', requiresAuth: true },
  },
  {
    path: '/building/:buildingId',
    name: 'BuildingDetail',
    component: BuildingDetailPage,
    meta: { titleKey: 'nav.buildingDetail', requiresAuth: true },
  },
  {
    path: '/properties',
    name: 'Properties',
    component: PropertyListPage,
    props: { channel: 'sale' },
    meta: { titleKey: 'nav.properties' },
  },
  {
    path: '/properties/:listingId',
    name: 'PropertySaleDetail',
    component: PropertyDetailPage,
    props: { channel: 'sale' },
    meta: { titleKey: 'property.sale.detailTitle' },
  },
  {
    path: '/serviced-residences',
    name: 'ServicedResidences',
    component: ServicedResidencesPage,
    meta: { titleKey: 'nav.servicedResidences' },
  },
  {
    path: '/serviced-residences/my',
    redirect: '/member',
  },
  {
    path: '/serviced-residences/my/new',
    redirect: (to) => ({
      path: '/member',
      query: { ...to.query, tab: 'serviced-residences', propertyEditor: 'new' },
      hash: to.hash,
    }),
  },
  {
    path: '/serviced-residences/my/editor/:listingId?',
    redirect: (to) => ({
      path: '/member',
      query: {
        ...to.query,
        tab: 'serviced-residences',
        propertyEditor: to.params.listingId ? String(to.params.listingId) : 'new',
      },
      hash: to.hash,
    }),
  },
  {
    path: '/serviced-residences/:listingId',
    name: 'ServicedResidenceDetail',
    component: ServicedResidenceDetailPage,
    meta: { titleKey: 'property.serviced.detailTitle' },
  },
];

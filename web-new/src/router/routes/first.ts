/*
 * 服務住宅路由。
 * 1. 定義服務住宅主路由。
 * 2. 對外輸出服務住宅路由陣列供總路由組裝。
 */
import type { RouteRecordRaw } from 'vue-router';

import PropertyDetailPage from '@/pages/property/detail/PropertyDetailPage.vue';
import PropertyListPage from '@/pages/property/list/PropertyListPage.vue';

// 1. 輸出服務住宅路由
export const firstRoutes: RouteRecordRaw[] = [
  {
    path: '/serviced-residences',
    name: 'ServicedResidences',
    component: PropertyListPage,
    props: { channel: 'serviced' },
    meta: { titleKey: 'nav.servicedResidences' },
  },
  {
    path: '/serviced-residences/my',
    redirect: '/account/properties/serviced-residences',
  },
  {
    path: '/serviced-residences/my/new',
    redirect: '/account/properties/serviced-residences/new',
  },
  {
    path: '/serviced-residences/my/editor/:listingId?',
    redirect: (to) => ({
      path: to.params.listingId
        ? `/account/properties/serviced-residences/editor/${String(to.params.listingId)}`
        : '/account/properties/serviced-residences/new',
      query: to.query,
      hash: to.hash,
    }),
  },
  {
    path: '/serviced-residences/:listingId',
    name: 'ServicedResidenceDetail',
    component: PropertyDetailPage,
    props: { channel: 'serviced' },
    meta: { titleKey: 'property.serviced.detailTitle' },
  },
];

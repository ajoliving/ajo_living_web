/*
 * 服務住宅路由。
 * 1. 定義服務住宅主路由。
 * 2. 對外輸出服務住宅路由陣列供總路由組裝。
 */
import type { RouteRecordRaw } from 'vue-router';

import PropertyDetailPage from '@/pages/property/detail/PropertyDetailPage.vue';
import PropertyEditorPage from '@/pages/property/editor/PropertyEditorPage.vue';
import PropertyListPage from '@/pages/property/list/PropertyListPage.vue';
import PropertyMyPage from '@/pages/property/my/PropertyMyPage.vue';

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
    name: 'ServicedResidenceMy',
    component: PropertyMyPage,
    props: { channel: 'serviced' },
    meta: { titleKey: 'property.serviced.myTitle', requiresAuth: true },
  },
  {
    path: '/serviced-residences/my/new',
    name: 'ServicedResidencePublish',
    component: PropertyEditorPage,
    props: { channel: 'serviced' },
    meta: { titleKey: 'property.serviced.publishTitle', requiresAuth: true },
  },
  {
    path: '/serviced-residences/my/editor/:listingId?',
    name: 'ServicedResidenceEditor',
    component: PropertyEditorPage,
    props: { channel: 'serviced' },
    meta: { titleKey: 'property.serviced.publishTitle', requiresAuth: true },
  },
  {
    path: '/serviced-residences/:listingId',
    name: 'ServicedResidenceDetail',
    component: PropertyDetailPage,
    props: { channel: 'serviced' },
    meta: { titleKey: 'property.serviced.detailTitle' },
  },
];

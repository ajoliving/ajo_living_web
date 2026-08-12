/*
 * 大廈副戶授權路由。
 * 1. 提供公開受邀帳戶啟用頁。
 */
import type { RouteRecordRaw } from 'vue-router';

const BuildingAuthorizationAcceptPage = () => import('@/pages/account/building-authorization-accept/Page.vue');

export const buildingAuthorizationRoutes: RouteRecordRaw[] = [
  {
    path: '/auth/building-authorizations/accept',
    name: 'BuildingAuthorizationAccept',
    component: BuildingAuthorizationAcceptPage,
    meta: { titleKey: 'building.authorizationAccept.title' },
  },
];

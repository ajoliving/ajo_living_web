/*
 * 大廈副戶授權路由。
 * 1. 提供公開受邀帳戶啟用頁。
 * 2. 提供 iSmart 副戶授權確認頁（電郵連結與站內通知共用）。
 */
import type { RouteRecordRaw } from 'vue-router';

const BuildingAuthorizationAcceptPage = () => import('@/pages/account/building-authorization-accept/Page.vue');
const IsmartAuthorizationConfirmPage = () => import('@/pages/account/ismart-authorization-confirm/Page.vue');

export const buildingAuthorizationRoutes: RouteRecordRaw[] = [
  {
    path: '/auth/building-authorizations/accept',
    name: 'BuildingAuthorizationAccept',
    component: BuildingAuthorizationAcceptPage,
    meta: { titleKey: 'building.authorizationAccept.title' },
  },
  {
    path: '/auth/ismart-subaccount-authorizations/confirm',
    name: 'IsmartSubaccountAuthorizationConfirm',
    component: IsmartAuthorizationConfirmPage,
    meta: { titleKey: 'building.ismartAuthorizationConfirm.title' },
  },
];

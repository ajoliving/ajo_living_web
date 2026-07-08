/*
 * 支付路由。
 * 1. 定義 AJO Pay 首頁與付款頁兩個正式入口。
 * 2. 保留舊支付子路徑兼容跳轉，避免外部鏈接失效。
 */
import type { RouteRecordRaw } from 'vue-router';

const PaymentPayPage = () => import('@/pages/payments/pay/Page.vue');
const PaymentsPage = () => import('@/pages/payments/Page.vue');

// 1. 輸出支付路由
export const paymentRoutes: RouteRecordRaw[] = [
  {
    path: '/payment',
    redirect: '/payments',
  },
  {
    path: '/payment/unit',
    redirect: '/account/profile',
  },
  {
    path: '/payment/order',
    redirect: (to) => ({
      path: '/payments/pay',
      query: to.query,
    }),
  },
  {
    path: '/payment/result',
    redirect: (to) => ({
      path: '/payments/pay',
      query: to.query,
    }),
  },
  {
    path: '/payment/h5/redirect',
    redirect: (to) => ({
      path: '/payments/pay',
      query: to.query,
    }),
  },
  {
    path: '/payments',
    name: 'Payments',
    component: PaymentsPage,
    meta: { titleKey: 'nav.payments', requiresAuth: true },
  },
  {
    path: '/payments/pay',
    name: 'PaymentPay',
    component: PaymentPayPage,
    meta: { titleKey: 'nav.payments', requiresAuth: true },
  },
  {
    path: '/payments/overview',
    redirect: '/payments',
  },
  {
    path: '/payments/bills',
    redirect: '/payments/pay',
  },
  {
    path: '/payments/cart',
    redirect: '/payments/pay',
  },
  {
    path: '/payments/orders',
    redirect: (to) => ({
      path: '/payments/pay',
      query: to.query,
    }),
  },
  {
    path: '/payments/records',
    redirect: '/payments/pay',
  },
  {
    path: '/payments/history',
    redirect: '/payments/pay',
  },
  {
    path: '/payments/accounting',
    redirect: '/payments',
  },
  {
    path: '/payments/units',
    redirect: '/account/profile',
  },
];

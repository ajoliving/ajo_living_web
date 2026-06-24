/*
 * 支付路由。
 * 1. 定義支付中心主路由。
 * 2. 對外輸出支付路由陣列供總路由組裝。
 */
import type { RouteRecordRaw } from 'vue-router';

import PaymentsPage from '@/pages/payments/Page.vue';
import PaymentAccountingPage from '@/pages/payments/accounting/Page.vue';
import PaymentBillsPage from '@/pages/payments/bills/Page.vue';
import PaymentCartPage from '@/pages/payments/cart/Page.vue';
import PaymentHistoryPage from '@/pages/payments/history/Page.vue';
import PaymentOverviewPage from '@/pages/payments/overview/Page.vue';
import PaymentOrdersPage from '@/pages/payments/orders/Page.vue';
import PaymentUnitsPage from '@/pages/payments/units/Page.vue';

// 1. 輸出支付路由
export const paymentRoutes: RouteRecordRaw[] = [
  {
    path: '/payment/unit',
    redirect: (to) => ({
      path: '/payments/units',
      query: to.query,
    }),
  },
  {
    path: '/payment/order',
    redirect: (to) => ({
      path: '/payments/orders',
      query: to.query,
    }),
  },
  {
    path: '/payment/result',
    redirect: (to) => ({
      path: '/payments/orders',
      query: to.query,
    }),
  },
  {
    path: '/payment/h5/redirect',
    redirect: (to) => ({
      path: '/payments/orders',
      query: to.query,
    }),
  },
  {
    path: '/payments',
    name: 'Payments',
    component: PaymentsPage,
    meta: { titleKey: 'nav.payments', requiresAuth: true },
    children: [
      {
        path: '',
        name: 'PaymentsOverview',
        component: PaymentOverviewPage,
        meta: { titleKey: 'nav.payments', requiresAuth: true },
      },
      {
        path: 'units',
        name: 'PaymentUnits',
        component: PaymentUnitsPage,
        meta: { titleKey: 'nav.paymentUnits', requiresAuth: true },
      },
      {
        path: 'bills',
        name: 'PaymentBills',
        component: PaymentBillsPage,
        meta: { titleKey: 'nav.paymentBills', requiresAuth: true },
      },
      {
        path: 'cart',
        name: 'PaymentCart',
        component: PaymentCartPage,
        meta: { titleKey: 'nav.paymentCart', requiresAuth: true },
      },
      {
        path: 'records',
        name: 'PaymentRecords',
        redirect: '/payments/history',
      },
      {
        path: 'orders',
        name: 'PaymentOrders',
        component: PaymentOrdersPage,
        meta: { titleKey: 'nav.paymentOrders', requiresAuth: true },
      },
      {
        path: 'accounting',
        name: 'PaymentAccounting',
        component: PaymentAccountingPage,
        meta: { titleKey: 'nav.paymentAccounting', requiresAuth: true },
      },
      {
        path: 'history',
        name: 'PaymentHistory',
        component: PaymentHistoryPage,
        meta: { titleKey: 'nav.paymentHistory', requiresAuth: true },
      },
    ],
  },
];

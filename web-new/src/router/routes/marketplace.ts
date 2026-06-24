/*
 * 二手交易路由。
 * 1. 定義主路由與公開瀏覽路由。
 * 2. 保留舊會員路徑的兼容跳轉。
 */
import type { RouteRecordRaw } from 'vue-router';

import MarketplacePage from '@/pages/marketplace/MarketplacePage.vue';
import MarketplaceDiscoverPage from '@/pages/marketplace/discover/MarketplaceDiscoverPage.vue';
import MarketplaceFilterPage from '@/pages/marketplace/filter/MarketplaceFilterPage.vue';
import MarketplaceListingPage from '@/pages/marketplace/listing/MarketplaceListingPage.vue';
import MarketplaceSellerPage from '@/pages/marketplace/seller/MarketplaceSellerPage.vue';

// 1. 取得舊路徑剩餘片段
const normalizeLegacyRestPath = (value: string | string[] | undefined): string =>
  Array.isArray(value) ? value.join('/') : value || '';

// 1. 輸出二手交易路由
export const marketplaceRoutes: RouteRecordRaw[] = [
  {
    path: '/marketplace',
    component: MarketplacePage,
    meta: { titleKey: 'nav.marketplace' },
    children: [
      {
        path: '',
        redirect: '/marketplace/discover',
      },
      {
        path: 'discover',
        name: 'MarketplaceDiscover',
        component: MarketplaceDiscoverPage,
        meta: { titleKey: 'nav.discover' },
      },
      {
        path: 'filter',
        name: 'MarketplaceFilter',
        component: MarketplaceFilterPage,
        meta: { titleKey: 'nav.filter' },
      },
      {
        path: 'settings/:rest(.*)*',
        redirect: (to) => {
          const restPath = normalizeLegacyRestPath(to.params.rest);

          return {
            path: restPath
              ? `/account/marketplace/settings/${restPath}`
              : '/account/marketplace/settings',
            query: to.query,
            hash: to.hash,
          };
        },
      },
      {
        path: 'publish',
        redirect: (to) => ({
          path: '/account/listings/new',
          query: to.query,
          hash: to.hash,
        }),
      },
      {
        path: 'my-listings',
        redirect: (to) => ({
          path: '/account/listings',
          query: to.query,
          hash: to.hash,
        }),
      },
      {
        path: 'my-listings/editor/:listingId?',
        redirect: (to) => ({
          path: to.params.listingId
            ? `/account/listings/editor/${String(to.params.listingId)}`
            : '/account/listings/new',
          query: to.query,
          hash: to.hash,
        }),
      },
      {
        path: 'my-listings/preview/:listingId',
        redirect: (to) => ({
          path: `/account/listings/item/${String(to.params.listingId)}`,
          query: to.query,
          hash: to.hash,
        }),
      },
      {
        path: 'chat/:conversationId?',
        redirect: (to) => ({
          path: to.params.conversationId
            ? `/account/chat/${String(to.params.conversationId)}`
            : '/account/chat',
          query: to.query,
          hash: to.hash,
        }),
      },
      {
        path: 'my/:rest(.*)*',
        redirect: (to) => {
          const restPath = normalizeLegacyRestPath(to.params.rest);

          return {
            path: restPath
              ? `/account/marketplace/my/${restPath}`
              : '/account/marketplace/my',
            query: to.query,
            hash: to.hash,
          };
        },
      },
      {
        path: 'listing/:listingId',
        name: 'MarketplaceListingDetail',
        component: MarketplaceListingPage,
        meta: { titleKey: 'marketplace.detail.title' },
      },
      {
        path: 'seller/:sellerId',
        name: 'MarketplaceSeller',
        component: MarketplaceSellerPage,
        meta: { titleKey: 'marketplace.seller.title' },
      },
    ],
  },
];

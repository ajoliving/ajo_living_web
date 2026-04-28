/*
 * 二手交易路由。
 * 1. 定義主路由與有效子路由。
 * 2. 對外輸出二手交易路由陣列供總路由組裝。
 */
import type { RouteRecordRaw } from 'vue-router';

import MarketplacePage from '@/pages/marketplace/MarketplacePage.vue';
import MarketplaceDiscoverPage from '@/pages/marketplace/discover/MarketplaceDiscoverPage.vue';
import MarketplaceFilterPage from '@/pages/marketplace/filter/MarketplaceFilterPage.vue';
import MarketplaceMyPage from '@/pages/marketplace/my/MarketplaceMyPage.vue';
import MarketplaceMyFavoritesPage from '@/pages/marketplace/my/favorites/MarketplaceMyFavoritesPage.vue';
import MarketplaceMyOrdersPage from '@/pages/marketplace/my/orders/MarketplaceMyOrdersPage.vue';
import MarketplaceMyProfilePage from '@/pages/marketplace/my/profile/MarketplaceMyProfilePage.vue';
import MarketplaceMyListingsPage from '@/pages/marketplace/my-listings/MarketplaceMyListingsPage.vue';
import MarketplaceListingEditorPage from '@/pages/marketplace/my-listings/editor/MarketplaceListingEditorPage.vue';
import MarketplaceMyListingPreviewPage from '@/pages/marketplace/my-listings/preview/MarketplaceMyListingPreviewPage.vue';
import MarketplaceChatPage from '@/pages/marketplace/chat/MarketplaceChatPage.vue';
import MarketplaceListingPage from '@/pages/marketplace/listing/MarketplaceListingPage.vue';
import MarketplaceSellerPage from '@/pages/marketplace/seller/MarketplaceSellerPage.vue';

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
        path: 'publish',
        redirect: '/marketplace/my/new',
      },
      {
        path: 'my-listings',
        redirect: '/marketplace/my/listings',
      },
      {
        path: 'my-listings/editor/:listingId?',
        redirect: (to) => ({
          path: to.params.listingId
            ? `/marketplace/my/editor/${String(to.params.listingId)}`
            : '/marketplace/my/new',
        }),
      },
      {
        path: 'my-listings/preview/:listingId',
        redirect: (to) => ({
          path: `/marketplace/my/preview/${String(to.params.listingId)}`,
        }),
      },
      {
        path: 'chat/:conversationId?',
        redirect: (to) => ({
          path: to.params.conversationId
            ? `/marketplace/my/chat/${String(to.params.conversationId)}`
            : '/marketplace/my/chat',
        }),
      },
      {
        path: 'my',
        component: MarketplaceMyPage,
        meta: { titleKey: 'nav.my', requiresAuth: true },
        children: [
          {
            path: '',
            redirect: '/marketplace/my/listings',
          },
          {
            path: 'listings',
            name: 'MarketplaceMyListings',
            component: MarketplaceMyListingsPage,
            meta: { titleKey: 'nav.myListings', requiresAuth: true },
          },
          {
            path: 'new',
            name: 'MarketplacePublish',
            component: MarketplaceListingEditorPage,
            meta: { titleKey: 'marketplace.publish.title', requiresAuth: true },
          },
          {
            path: 'editor/:listingId?',
            name: 'MarketplaceListingEditor',
            component: MarketplaceListingEditorPage,
            meta: { titleKey: 'marketplace.editor.title', requiresAuth: true },
          },
          {
            path: 'listing/:listingId',
            name: 'MarketplaceMyListingDetail',
            component: MarketplaceMyListingPreviewPage,
            meta: { titleKey: 'marketplace.mine.detailTitle', requiresAuth: true },
          },
          {
            path: 'preview/:listingId',
            redirect: (to) => ({
              path: `/marketplace/my/listing/${String(to.params.listingId)}`,
            }),
          },
          {
            path: 'orders',
            name: 'MarketplaceMyOrders',
            component: MarketplaceMyOrdersPage,
            meta: { titleKey: 'marketplace.myHub.orders', requiresAuth: true },
          },
          {
            path: 'favorites',
            name: 'MarketplaceMyFavorites',
            component: MarketplaceMyFavoritesPage,
            meta: { titleKey: 'marketplace.myHub.favorites', requiresAuth: true },
          },
          {
            path: 'profile',
            name: 'MarketplaceMyProfile',
            component: MarketplaceMyProfilePage,
            meta: { titleKey: 'marketplace.myHub.profile', requiresAuth: true },
          },
          {
            path: 'chat/:conversationId?',
            name: 'MarketplaceChat',
            component: MarketplaceChatPage,
            meta: { titleKey: 'nav.chat', requiresAuth: true },
          },
        ],
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

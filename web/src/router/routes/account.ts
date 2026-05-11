/*
 * 會員中心路由。
 * 1. 定義登入頁、會員資料頁、二手交易會員頁與通知兼容跳轉路由。
 * 2. 對外輸出會員中心路由陣列供總路由組裝。
 */
import type { RouteRecordRaw } from 'vue-router';

import LoginPage from '@/pages/account/login/LoginPage.vue';
import AccountMyPage from '@/pages/account/my/AccountMyPage.vue';
import AccountNotificationsPage from '@/pages/account/my/notifications/AccountNotificationsPage.vue';
import ProfileShellPage from '@/pages/account/my/profile/Page.vue';
import AccountProfilePage from '@/pages/account/my/profile/info/Page.vue';
import AccountWalletPage from '@/pages/account/my/profile/wallet/Page.vue';
import MarketplaceChatPage from '@/pages/marketplace/chat/MarketplaceChatPage.vue';
import MarketplaceMyFavoritesPage from '@/pages/marketplace/my/favorites/MarketplaceMyFavoritesPage.vue';
import MarketplaceMyOrdersPage from '@/pages/marketplace/my/orders/MarketplaceMyOrdersPage.vue';
import MarketplaceMyProfilePage from '@/pages/marketplace/my/profile/MarketplaceMyProfilePage.vue';
import MarketplaceMyPage from '@/pages/marketplace/my/MarketplaceMyPage.vue';
import MarketplaceMyListingsPage from '@/pages/marketplace/my-listings/MarketplaceMyListingsPage.vue';
import MarketplaceListingEditorPage from '@/pages/marketplace/my-listings/editor/MarketplaceListingEditorPage.vue';
import MarketplaceMyListingPreviewPage from '@/pages/marketplace/my-listings/preview/MarketplaceMyListingPreviewPage.vue';
import MarketplaceSettingsPage from '@/pages/marketplace/settings/Page.vue';
import MarketplaceSettingsDiscoverPage from '@/pages/marketplace/settings/discover/Page.vue';
import MarketplaceSettingsHomeCarouselPage from '@/pages/marketplace/settings/home-carousel/Page.vue';
import MarketplaceSettingsHomeHeroCardsPage from '@/pages/marketplace/settings/home-hero-cards/Page.vue';
import MarketplaceSettingsLoginHeroPage from '@/pages/marketplace/settings/login-hero/Page.vue';
import MarketplaceSettingsNoticePage from '@/pages/marketplace/settings/notice/Page.vue';
import MarketplaceManagementPage from '@/pages/marketplace/management/Page.vue';
import MarketplaceManagementMembersPage from '@/pages/marketplace/management/members/Page.vue';
import MarketplaceManagementPropertySalesPage from '@/pages/marketplace/management/property-sales/Page.vue';
import MarketplaceManagementRewardAdEditorPage from '@/pages/marketplace/management/reward-ad-editor/Page.vue';
import MarketplaceManagementRewardAdsPage from '@/pages/marketplace/management/reward-ads/Page.vue';
import MarketplaceManagementSecondhandListingsPage from '@/pages/marketplace/management/secondhand-listings/Page.vue';
import MarketplaceManagementServicedApartmentsPage from '@/pages/marketplace/management/serviced-apartments/Page.vue';
import MarketplaceManagementWalletGrantsPage from '@/pages/marketplace/management/wallet-grants/Page.vue';
import MarketplaceManagementWalletTransactionsPage from '@/pages/marketplace/management/wallet-transactions/Page.vue';

// 1. 輸出會員中心路由
export const accountRoutes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: LoginPage,
    meta: { titleKey: 'nav.login' },
  },
  {
    path: '/account',
    component: AccountMyPage,
    meta: { requiresAuth: true },
    children: [
      {
        path: '',
        redirect: '/account/profile',
      },
      {
        path: 'profile',
        component: ProfileShellPage,
        meta: { titleKey: 'account.profile.title', requiresAuth: true },
        children: [
          {
            path: '',
            redirect: '/account/profile/info',
          },
          {
            path: 'info',
            name: 'AccountProfile',
            component: AccountProfilePage,
            meta: { titleKey: 'account.profile.title', requiresAuth: true },
          },
          {
            path: 'wallet',
            name: 'AccountWallet',
            component: AccountWalletPage,
            meta: { titleKey: 'account.wallet.title', requiresAuth: true },
          },
        ],
      },
      {
        path: 'wallet',
        redirect: '/account/profile/wallet',
      },
      {
        path: 'notifications',
        name: 'AccountNotificationsRedirect',
        component: AccountNotificationsPage,
        meta: { titleKey: 'chat.systemNoticeTitle', requiresAuth: true },
      },
      {
        path: 'marketplace/my',
        component: MarketplaceMyPage,
        meta: { titleKey: 'nav.marketplaceMy', requiresAuth: true },
        children: [
          {
            path: '',
            redirect: '/account/marketplace/my/listings',
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
              path: `/account/marketplace/my/listing/${String(to.params.listingId)}`,
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
        path: 'marketplace/settings',
        name: 'MarketplaceSettings',
        component: MarketplaceSettingsPage,
        meta: { titleKey: 'nav.marketplaceSettings', requiresAuth: true, requiresStaff: true },
        children: [
          {
            path: '',
            redirect: '/account/marketplace/settings/discover',
          },
          {
            path: 'wallet-grants',
            redirect: '/account/marketplace/management/wallet-grants',
          },
          {
            path: 'reward-ad-editor/:taskId?',
            redirect: (to) => ({
              path: to.params.taskId
                ? `/account/marketplace/management/reward-ad-editor/${String(to.params.taskId)}`
                : '/account/marketplace/management/reward-ad-editor',
              query: to.query,
              hash: to.hash,
            }),
          },
          {
            path: 'reward-ads',
            redirect: '/account/marketplace/management/reward-ads',
          },
          {
            path: 'wallet-transactions',
            redirect: '/account/marketplace/management/wallet-transactions',
          },
          {
            path: 'wallet',
            redirect: '/account/marketplace/management/wallet-grants',
          },
          {
            path: 'notice',
            name: 'MarketplaceSettingsNotice',
            component: MarketplaceSettingsNoticePage,
            meta: { titleKey: 'marketplace.settings.noticeSection', requiresAuth: true, requiresStaff: true },
          },
          {
            path: 'discover',
            name: 'MarketplaceSettingsDiscover',
            component: MarketplaceSettingsDiscoverPage,
            meta: { titleKey: 'marketplace.settings.discoverSection', requiresAuth: true, requiresStaff: true },
          },
          {
            path: 'home-carousel',
            name: 'MarketplaceSettingsHomeCarousel',
            component: MarketplaceSettingsHomeCarouselPage,
            meta: { titleKey: 'marketplace.settings.homeCarouselSection', requiresAuth: true, requiresStaff: true },
          },
          {
            path: 'home-hero-cards',
            name: 'MarketplaceSettingsHomeHeroCards',
            component: MarketplaceSettingsHomeHeroCardsPage,
            meta: { titleKey: 'marketplace.settings.homeHeroCardsSection', requiresAuth: true, requiresStaff: true },
          },
          {
            path: 'login-hero',
            name: 'MarketplaceSettingsLoginHero',
            component: MarketplaceSettingsLoginHeroPage,
            meta: { titleKey: 'marketplace.settings.loginHeroSection', requiresAuth: true, requiresStaff: true },
          },
        ],
      },
      {
        path: 'marketplace/management',
        name: 'MarketplaceManagement',
        component: MarketplaceManagementPage,
        meta: { titleKey: 'nav.marketplaceManagement', requiresAuth: true, requiresStaff: true },
        children: [
          {
            path: '',
            redirect: '/account/marketplace/management/secondhand-listings',
          },
          {
            path: 'wallet-grants',
            name: 'MarketplaceManagementWalletGrants',
            component: MarketplaceManagementWalletGrantsPage,
            meta: { titleKey: 'marketplace.management.walletGrantSection', requiresAuth: true, requiresStaff: true },
          },
          {
            path: 'reward-ad-editor/:taskId?',
            name: 'MarketplaceManagementRewardAdEditor',
            component: MarketplaceManagementRewardAdEditorPage,
            meta: { titleKey: 'marketplace.management.walletAdPublishSection', requiresAuth: true, requiresStaff: true },
          },
          {
            path: 'reward-ads',
            name: 'MarketplaceManagementRewardAds',
            component: MarketplaceManagementRewardAdsPage,
            meta: { titleKey: 'marketplace.management.walletAdListSection', requiresAuth: true, requiresStaff: true },
          },
          {
            path: 'wallet-transactions',
            name: 'MarketplaceManagementWalletTransactions',
            component: MarketplaceManagementWalletTransactionsPage,
            meta: { titleKey: 'marketplace.management.walletTransactionsSection', requiresAuth: true, requiresStaff: true },
          },
          {
            path: 'secondhand-listings',
            name: 'MarketplaceManagementSecondhandListings',
            component: MarketplaceManagementSecondhandListingsPage,
            meta: { titleKey: 'marketplace.management.secondhandListings', requiresAuth: true, requiresStaff: true },
          },
          {
            path: 'property-sales',
            name: 'MarketplaceManagementPropertySales',
            component: MarketplaceManagementPropertySalesPage,
            meta: { titleKey: 'marketplace.management.propertySales', requiresAuth: true, requiresStaff: true },
          },
          {
            path: 'members',
            name: 'MarketplaceManagementMembers',
            component: MarketplaceManagementMembersPage,
            meta: { titleKey: 'marketplace.management.members', requiresAuth: true, requiresStaff: true },
          },
          {
            path: 'serviced-apartments',
            name: 'MarketplaceManagementServicedApartments',
            component: MarketplaceManagementServicedApartmentsPage,
            meta: { titleKey: 'marketplace.management.servicedApartments', requiresAuth: true, requiresStaff: true },
          },
        ],
      },
    ],
  },
];

/*
 * 會員中心路由。
 * 1. 定義登入頁、會員資料頁、二手交易會員頁與舊路徑兼容跳轉路由。
 * 2. 對外輸出會員中心路由陣列供總路由組裝。
 */
import type { RouteRecordRaw } from 'vue-router';

const AccountMyPage = () => import('@/pages/account/my/AccountMyPage.vue');
const AccountWalletPage = () => import('@/pages/account/my/profile/wallet/Page.vue');
const ForgotPasswordPage = () => import('@/pages/account/forgot-password/Page.vue');
const LoginPage = () => import('@/pages/account/login/LoginPage.vue');
const MarketplaceChatPage = () => import('@/pages/marketplace/chat/MarketplaceChatPage.vue');
const MarketplaceListingEditorPage = () =>
  import('@/pages/marketplace/my-listings/editor/MarketplaceListingEditorPage.vue');
const MarketplaceManagementPage = () => import('@/pages/marketplace/management/Page.vue');
const MarketplaceManagementRewardAdEditorPage = () =>
  import('@/pages/marketplace/management/reward-ad-editor/Page.vue');
const MarketplaceMyFavoritesPage = () =>
  import('@/pages/marketplace/my/favorites/MarketplaceMyFavoritesPage.vue');
const MarketplaceMyListingPreviewPage = () =>
  import('@/pages/marketplace/my-listings/preview/MarketplaceMyListingPreviewPage.vue');
const MarketplaceMyListingsPage = () =>
  import('@/pages/marketplace/my-listings/MarketplaceMyListingsPage.vue');
const MarketplaceMyPage = () => import('@/pages/marketplace/my/MarketplaceMyPage.vue');
const MarketplaceMyProfilePage = () =>
  import('@/pages/marketplace/my/profile/MarketplaceMyProfilePage.vue');
const MarketplaceSettingsDisplayAdsPage = () =>
  import('@/pages/marketplace/settings/display-ads/Page.vue');
const MarketplaceSettingsHomeHeroCardsPage = () =>
  import('@/pages/marketplace/settings/home-hero-cards/Page.vue');
const MarketplaceSettingsLoginHeroPage = () =>
  import('@/pages/marketplace/settings/login-hero/Page.vue');
const MarketplaceSettingsNoticePage = () => import('@/pages/marketplace/settings/notice/Page.vue');
const MarketplaceSettingsPage = () => import('@/pages/marketplace/settings/Page.vue');
const PropertyEditorPage = () => import('@/pages/property/editor/PropertyEditorPage.vue');
const PropertyMyPage = () => import('@/pages/property/my/PropertyMyPage.vue');

// 1. 管理中心舊子路徑對應真實管理分頁
const managementTabRoute = (tab: string) => ({
  path: '/account/marketplace/management',
  query: { tab },
});

// 2. 輸出會員中心路由
export const accountRoutes: RouteRecordRaw[] = [
  {
    path: '/chat',
    redirect: '/account/chat',
  },
  {
    path: '/saved',
    redirect: '/account/favorites',
  },
  {
    path: '/profile',
    name: 'Profile',
    component: AccountMyPage,
    meta: { titleKey: 'nav.memberCenter', requiresAuth: true },
  },
  {
    path: '/login',
    name: 'Login',
    component: LoginPage,
    meta: { titleKey: 'nav.login' },
  },
  {
    path: '/forgot-password',
    name: 'ForgotPassword',
    component: ForgotPasswordPage,
    meta: { titleKey: 'auth.forgotPassword' },
  },
  {
    path: '/account/marketplace/management',
    name: 'MarketplaceManagement',
    component: MarketplaceManagementPage,
    meta: { titleKey: 'nav.marketplaceManagement', requiresAuth: true, requiresStaff: true },
  },
  {
    path: '/account/marketplace/management/secondhand-listings',
    redirect: () => managementTabRoute('admin-furniture'),
  },
  {
    path: '/account/marketplace/management/property-sales',
    redirect: () => managementTabRoute('admin-properties'),
  },
  {
    path: '/account/marketplace/management/members',
    redirect: () => managementTabRoute('admin-members'),
  },
  {
    path: '/account/marketplace/management/system-notices',
    redirect: () => managementTabRoute('admin-notices'),
  },
  {
    path: '/account/marketplace/management/serviced-apartments',
    redirect: () => managementTabRoute('admin-homes'),
  },
  {
    path: '/account/marketplace/management/reward-ad-editor/:taskId?',
    name: 'MarketplaceManagementRewardAdEditor',
    component: MarketplaceManagementRewardAdEditorPage,
    meta: { titleKey: 'marketplace.management.walletAdPublishSection', requiresAuth: true, requiresStaff: true },
  },
  {
    path: '/account/marketplace/management/reward-ads',
    redirect: () => managementTabRoute('admin-ads'),
  },
  {
    path: '/account/marketplace/management/wallet-transactions',
    redirect: () => managementTabRoute('admin-points'),
  },
  {
    path: '/account/marketplace/management/wallet-grants',
    redirect: () => managementTabRoute('admin-members'),
  },
  {
    path: '/account/marketplace/settings',
    component: MarketplaceSettingsPage,
    meta: { titleKey: 'nav.marketplaceSettings', requiresAuth: true, requiresStaff: true },
    children: [
      {
        path: '',
        name: 'MarketplaceSettings',
        redirect: '/account/marketplace/settings/home-hero-cards',
      },
      {
        path: 'notice',
        name: 'MarketplaceSettingsNotice',
        component: MarketplaceSettingsNoticePage,
        meta: { titleKey: 'marketplace.settings.noticeSection', requiresAuth: true, requiresStaff: true },
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
      {
        path: 'display-ads',
        name: 'MarketplaceSettingsDisplayAds',
        component: MarketplaceSettingsDisplayAdsPage,
        meta: { titleKey: 'marketplace.settings.displayAdSection', requiresAuth: true, requiresStaff: true },
      },
    ],
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
        name: 'AccountProfile',
        component: AccountMyPage,
        meta: { titleKey: 'account.profile.title', requiresAuth: true },
      },
      {
        path: 'profile/info',
        redirect: '/account/profile',
      },
      {
        path: 'profile/wallet',
        name: 'AccountWallet',
        component: AccountWalletPage,
        meta: { titleKey: 'account.wallet.title', requiresAuth: true },
      },
      {
        path: 'wallet',
        redirect: '/account/profile/wallet',
      },
      {
        path: 'notifications',
        redirect: '/notifications',
      },
      {
        path: 'chat/:conversationId?',
        name: 'AccountChat',
        component: MarketplaceChatPage,
        meta: { titleKey: 'nav.chat', requiresAuth: true },
      },
      {
        path: 'properties',
        redirect: '/account/properties/sale',
      },
      {
        path: 'properties/sale',
        name: 'AccountPropertySaleMy',
        component: PropertyMyPage,
        props: { channel: 'sale', basePath: '/account/properties/sale' },
        meta: { titleKey: 'property.sale.myTitle', requiresAuth: true },
      },
      {
        path: 'properties/sale/new',
        name: 'AccountPropertySalePublish',
        component: PropertyEditorPage,
        props: { channel: 'sale', returnPath: '/account/properties/sale' },
        meta: { titleKey: 'property.sale.publishTitle', requiresAuth: true },
      },
      {
        path: 'properties/sale/editor/:listingId?',
        name: 'AccountPropertySaleEditor',
        component: PropertyEditorPage,
        props: { channel: 'sale', returnPath: '/account/properties/sale' },
        meta: { titleKey: 'property.sale.publishTitle', requiresAuth: true },
      },
      {
        path: 'properties/serviced-residences',
        name: 'AccountServicedResidenceMy',
        component: PropertyMyPage,
        props: { channel: 'serviced', basePath: '/account/properties/serviced-residences' },
        meta: { titleKey: 'property.serviced.myTitle', requiresAuth: true },
      },
      {
        path: 'properties/serviced-residences/new',
        name: 'AccountServicedResidencePublish',
        component: PropertyEditorPage,
        props: { channel: 'serviced', returnPath: '/account/properties/serviced-residences' },
        meta: { titleKey: 'property.serviced.publishTitle', requiresAuth: true },
      },
      {
        path: 'properties/serviced-residences/editor/:listingId?',
        name: 'AccountServicedResidenceEditor',
        component: PropertyEditorPage,
        props: { channel: 'serviced', returnPath: '/account/properties/serviced-residences' },
        meta: { titleKey: 'property.serviced.publishTitle', requiresAuth: true },
      },
      {
        path: 'listings',
        name: 'AccountSecondhandListings',
        component: MarketplaceMyListingsPage,
        meta: { titleKey: 'nav.secondhandFurniture', requiresAuth: true },
      },
      {
        path: 'listings/new',
        name: 'AccountSecondhandPublish',
        component: MarketplaceListingEditorPage,
        meta: { titleKey: 'marketplace.publish.title', requiresAuth: true },
      },
      {
        path: 'listings/editor/:listingId?',
        name: 'AccountSecondhandListingEditor',
        component: MarketplaceListingEditorPage,
        meta: { titleKey: 'marketplace.editor.title', requiresAuth: true },
      },
      {
        path: 'listings/item/:listingId',
        name: 'AccountSecondhandListingDetail',
        component: MarketplaceMyListingPreviewPage,
        meta: { titleKey: 'marketplace.mine.detailTitle', requiresAuth: true },
      },
      {
        path: 'listings/preview/:listingId',
        redirect: (to) => ({
          path: `/account/listings/item/${String(to.params.listingId)}`,
        }),
      },
      {
        path: 'favorites',
        name: 'AccountProductFavorites',
        component: MarketplaceMyFavoritesPage,
        meta: { titleKey: 'nav.productFavorites', requiresAuth: true },
      },
      {
        path: 'marketplace/my',
        component: MarketplaceMyPage,
        meta: { titleKey: 'nav.marketplaceMy', requiresAuth: true },
        children: [
          {
            path: '',
            redirect: '/account/listings',
          },
          {
            path: 'listings',
            redirect: '/account/listings',
          },
          {
            path: 'new',
            redirect: '/account/listings/new',
          },
          {
            path: 'editor/:listingId?',
            redirect: (to) => ({
              path: to.params.listingId
                ? `/account/listings/editor/${String(to.params.listingId)}`
                : '/account/listings/new',
              query: to.query,
              hash: to.hash,
            }),
          },
          {
            path: 'listing/:listingId',
            redirect: (to) => ({
              path: `/account/listings/item/${String(to.params.listingId)}`,
              query: to.query,
              hash: to.hash,
            }),
          },
          {
            path: 'preview/:listingId',
            redirect: (to) => ({
              path: `/account/listings/item/${String(to.params.listingId)}`,
              query: to.query,
              hash: to.hash,
            }),
          },
          {
            path: 'orders',
            redirect: '/account/listings',
          },
          {
            path: 'favorites',
            redirect: '/account/favorites',
          },
          {
            path: 'profile',
            name: 'MarketplaceMyProfile',
            component: MarketplaceMyProfilePage,
            meta: { titleKey: 'marketplace.myHub.profile', requiresAuth: true },
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
        ],
      },
    ],
  },
];

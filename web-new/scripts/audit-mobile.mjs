/*
 * 手機端瀏覽器審計工具。
 * 1. 使用 Playwright 以 iPhone 14 Pro Max 尺寸檢查主要路由。
 * 2. 驗證橫向溢出、頂部與底部導航、手機端頁腳隱藏、登入可見性、運行時錯誤、網絡錯誤、觸控目標、底部可操作內容與主要操作流。
 * 3. 預設 mock API，避免依賴本機後端；可用 `--mock-api=false` 改跑真實 API，或用 `--section` 分段定位。
 */
import { chromium } from '@playwright/test';

const DEFAULT_BASE_URL = 'http://localhost:5173';
const now = '2026-07-09T12:00:00+08:00';
const pagination = { page: 1, page_size: 10, total: 1 };
const stats = {
  records: 1,
  products: 1,
  stores: 1,
  offers: 0,
  parsedOffers: 0,
  unsupportedOffers: 0,
  ambiguousOffers: 0,
  uncalculatedOffers: 0,
};

// 1. 解析命令列參數
const readOption = (name, fallback) => {
  const prefix = `--${name}=`;
  const option = process.argv.find((arg) => arg.startsWith(prefix));

  return option ? option.slice(prefix.length) : fallback;
};

const normalizeBaseUrl = (value) => value.replace(/\/+$/, '');
const baseUrl = normalizeBaseUrl(readOption('base', process.env.MOBILE_AUDIT_BASE_URL ?? DEFAULT_BASE_URL));
const shouldMockApi = readOption('mock-api', process.env.MOBILE_AUDIT_MOCK_API ?? 'true') !== 'false';
const shouldShowSamples = readOption('samples', 'false') === 'true';
const auditSection = readOption('section', process.env.MOBILE_AUDIT_SECTION ?? 'all');
const allowLiveMutationFlows = readOption('allow-live-mutations', process.env.MOBILE_AUDIT_ALLOW_LIVE_MUTATIONS ?? 'false') === 'true';
const allowedAuditSections = ['all', 'guest', 'staff', 'member-staff', 'viewport'];
const mockSession = { accessToken: 'mock-access', refreshToken: 'mock-refresh' };
const liveStaffSession = {
  accessToken: process.env.MOBILE_AUDIT_STAFF_ACCESS_TOKEN ?? process.env.MOBILE_AUDIT_ACCESS_TOKEN ?? '',
  refreshToken: process.env.MOBILE_AUDIT_STAFF_REFRESH_TOKEN ?? process.env.MOBILE_AUDIT_REFRESH_TOKEN ?? '',
};
const liveMemberSession = {
  accessToken: process.env.MOBILE_AUDIT_MEMBER_ACCESS_TOKEN ?? process.env.MOBILE_AUDIT_ACCESS_TOKEN ?? '',
  refreshToken: process.env.MOBILE_AUDIT_MEMBER_REFRESH_TOKEN ?? process.env.MOBILE_AUDIT_REFRESH_TOKEN ?? '',
};
const mobileViewportScenarios = [
  { name: 'iphone-14-pro-max-portrait', viewport: { width: 430, height: 932 } },
  { name: 'iphone-14-pro-max-short', viewport: { width: 430, height: 740 } },
  { name: 'iphone-14-pro-max-landscape', viewport: { width: 932, height: 430 } },
];

const api = (data) => ({ code: 'OK', message: 'ok', data });
const paginated = (items = []) => ({ items, pagination });

const gpProduct = {
  code: 'P001',
  name: '示範商品',
  brand: 'Demo',
  image_url: 'https://images.unsplash.com/photo-1604719312566-8912e9227c6a?auto=format&fit=crop&w=800&q=80',
  imageUrl: 'https://images.unsplash.com/photo-1604719312566-8912e9227c6a?auto=format&fit=crop&w=800&q=80',
  category1: '食品',
  category2: '日用品',
  category3: '測試',
  minPrice: 12,
  maxPrice: 16,
  priceDiff: 4,
  diffPercent: 25,
  listPrice: 16,
  effectiveUnitPrice: 12,
  bestEffectiveUnitPrice: 12,
  discountAmount: 4,
  discountRate: 0.25,
  stores: ['Demo Store'],
  storePrices: [],
  hasOffer: true,
  bestOffer: '會員價',
  bestStore: 'Demo Store',
  offerType: 'discount',
  offerPattern: 'member',
  parseStatus: 'ok',
  isFavorite: false,
};

const community = {
  public_id: 'community-1',
  community_type: 'building',
  name_zh: '示範大廈',
  name_en: 'Demo Building',
  district_code: 'central',
  address_text: 'Hong Kong',
};

const member = {
  public_id: 'staff-member',
  email: 'staff@example.com',
  phone_country_code: '+852',
  phone_number: '60000000',
  member_status: 'active',
  member_type: 'owner',
  is_staff: true,
  role: 'staff',
  roles: ['staff'],
  permissions: ['marketplace:manage'],
  display_name: 'Staff Member',
  avatar_url: '',
  publisher_identity_type: 'owner',
  district_code: 'central',
  profile_completed: true,
  ajo_balance: 1200,
  primary_community: community,
  bound_building_ids: ['B001'],
  bound_flat_unit_ids: ['U001'],
  residence_floor: '12',
  residence_unit: 'A',
  created_at: now,
  updated_at: now,
  ismart_linked: true,
  ismart_username: 'staff',
  ismart_msg: {
    user_id: 1,
    username: 'staff',
    is_staff: true,
    building: ['B001'],
  },
};

const regularMember = {
  ...member,
  public_id: 'member',
  email: 'member@example.com',
  is_staff: false,
  role: 'member',
  roles: ['member'],
  permissions: [],
  display_name: 'Member',
  ismart_username: 'member',
  ismart_msg: {
    ...member.ismart_msg,
    username: 'member',
    is_staff: false,
  },
};

const image = {
  media_asset_id: 'asset-1',
  url: 'https://images.unsplash.com/photo-1560185007-c5ca9d2c014d?auto=format&fit=crop&w=900&q=80',
  sort_order: 0,
  is_cover: true,
};

const contactSummary = {
  show_phone: true,
  show_whatsapp: true,
  show_chat: true,
  show_inquiry_form: true,
  contact_attributes: {
    phone: '60000000',
    email: 'agent@example.com',
  },
};

const owner = {
  user_id: 'seller-1',
  public_id: 'seller-1',
  display_name: 'Demo Seller',
  avatar_url: '',
  publisher_identity_type: 'owner',
};

const secondhand = {
  listing_id: 'listing-1',
  title: '示範家具',
  summary: '保養良好，可安排交收。',
  description: '示範詳情',
  district_code: 'central',
  category_code: 'home_furniture',
  price_mode: 'fixed',
  price_hkd: 1200,
  condition_level: 'used_excellent',
  visibility_scope: 'public',
  contact_method: 'chat',
  is_free_giveaway: false,
  publisher_identity_type: 'owner',
  publication_status: 'published',
  business_status: 'available',
  updated_at: now,
  published_at: now,
  expire_at: null,
  is_favorited: false,
  cover_image: image,
  images: [image],
  media_assets: [image],
  community,
  owner,
  seller: owner,
  dimension_text: '120 x 60 cm',
  pickup_region_code: 'hk_island',
  pickup_location_text: 'Central',
  delivery_tags: ['self_pickup'],
  contact_summary: contactSummary,
};

const propertySale = {
  property_no: 'AJO-001',
  transaction_type: 'sale',
  property_type: 'residential',
  estate_name: '示範花園',
  address_text: '中環示範街 1 號',
  asking_price_hkd: 6800000,
  monthly_rent_hkd: 18000,
  usable_area_sqft: 480,
  gross_area_sqft: 620,
  bedroom_count: 2,
  living_room_count: 1,
  bathroom_count: 1,
  floor_level: 'middle',
  direction: 'south',
  building_age: '10',
  feature_tags: [],
  contact_method: 'chat',
  publisher_role_label: '代理盤',
  agency_company_name: 'Demo Agency',
};

const roomType = {
  name: 'Studio',
  usable_area_sqft: 260,
  monthly_rent_hkd: 22000,
  included_fees: true,
  min_lease_months: 1,
  feature_tags: [],
};

const servicedPayload = {
  project_name: '服務住宅示範套房',
  address_text: '灣仔示範道 8 號',
  lowest_monthly_rent_hkd: 22000,
  min_usable_area_sqft: 260,
  min_lease_months: 1,
  facility_tags: [],
  service_tags: [],
  room_types: [roomType],
  contact_method: 'chat',
  publisher_role_label: '服務住宅',
};

const property = {
  listing_id: 'property-1',
  module: 'property_sale',
  title: '中環示範單位',
  summary: '兩房實用單位。',
  description: '示範樓盤詳情',
  district_code: 'central',
  publisher_identity_type: 'agent',
  publication_status: 'published',
  business_status: 'available',
  updated_at: now,
  published_at: now,
  expire_at: null,
  community,
  owner,
  cover_image: image,
  images: [image],
  property_sale: propertySale,
  serviced_apartment: null,
  contact_summary: contactSummary,
  is_favorite: false,
};

const serviced = {
  ...property,
  listing_id: 'serviced-1',
  module: 'serviced_apartment',
  title: '服務住宅示範套房',
  property_sale: null,
  serviced_apartment: servicedPayload,
};

const chat = {
  chat_id: 'chat-1',
  listing_id: 'listing-1',
  listing_title: '示範家具',
  chat_type: 'direct_listing_chat',
  biz_module: 'secondhand',
  last_message_preview: '你好，請問仍可交收嗎？',
  last_message_at: now,
  unread_count: 1,
  created_at: now,
  participants: [{ user_id: 'staff-member', role_in_chat: 'buyer', display_name: 'Staff Member' }],
  peer: { user_id: 'seller-1', public_id: 'seller-1', display_name: 'Demo Seller' },
  listing: secondhand,
};

const walletOrder = {
  order_id: 'wallet-order-1',
  mch_order_no: 'WALLET-001',
  pay_order_id: 'gateway-wallet-1',
  amount_hkd: 100,
  amount_cents: 10000,
  points_amount: 10000,
  pay_channel: 'WX_QR',
  pay_region: 'HK',
  pay_data_type: 'codeurl',
  pay_data: 'https://example.com/pay',
  state: 'PAYING',
  state_label: '待支付',
  gateway_state_code: 0,
  gateway_message: '',
  currency: 'HKD',
  created_at: now,
  updated_at: now,
  expire_time: now,
  paid_at: '',
  credited_at: '',
};

const paymentBill = {
  invoice_no: 'INV-001',
  item_name: '管理費',
  bill_dt: '2026-07-01',
  trs_to: '2026-07',
  flat_code: '12A',
  unit_name: '12A',
  net_amount: 1200,
};

const publicPageRoutes = [
  '/',
  '/login',
  '/forgot-password',
  '/properties',
  '/properties/property-1',
  '/furniture',
  '/furniture/listing-1',
  '/marketplace',
  '/marketplace/filter',
  '/marketplace/listing/listing-1',
  '/marketplace/seller/seller-1',
  '/supermarket-offers',
  '/supermarket-offers/products/P001',
  '/serviced-residences',
  '/serviced-residences/serviced-1',
  '/trend',
];

const publicRedirectRoutes = [
  '/listing',
  '/detail',
  '/service',
  '/service-detail',
  '/serviced-residence',
  '/serviced-residence/booking',
  '/serviced-residence/serviced-1',
  '/market',
  '/market-detail',
  '/furniture/listing/listing-1',
  '/offers',
  '/trend/district/central',
];

const managementTabRoutes = [
  '/account/marketplace/management?tab=admin-furniture',
  '/account/marketplace/management?tab=admin-properties',
  '/account/marketplace/management?tab=admin-homes',
  '/account/marketplace/management?tab=admin-icctv',
  '/account/marketplace/management?tab=admin-members',
  '/account/marketplace/management?tab=admin-notices',
  '/account/marketplace/management?tab=admin-ads',
  '/account/marketplace/management?tab=admin-ad-settings',
  '/account/marketplace/management?tab=admin-points',
];

const publicRoutes = [
  ...publicPageRoutes,
  ...publicRedirectRoutes,
  '/not-existing-mobile-audit',
];

const authPageRoutes = [
  '/profile',
  '/account',
  '/account/profile',
  '/account/profile/wallet',
  '/account/chat',
  '/account/chat/chat-1',
  '/account/properties/sale',
  '/account/properties/sale/new',
  '/account/properties/sale/editor/property-1',
  '/account/properties/serviced-residences',
  '/account/properties/serviced-residences/new',
  '/account/properties/serviced-residences/editor/serviced-1',
  '/account/listings',
  '/account/listings/new',
  '/account/listings/editor/listing-1',
  '/account/listings/item/listing-1',
  '/account/favorites',
  '/account/marketplace/my',
  '/account/marketplace/my/profile',
  '/account/marketplace/management',
  '/account/marketplace/management/reward-ad-editor',
  '/account/marketplace/management/reward-ad-editor/task-1',
  '/account/marketplace/settings',
  '/account/marketplace/settings/notice',
  '/account/marketplace/settings/home-hero-cards',
  '/account/marketplace/settings/login-hero',
  '/account/marketplace/settings/display-ads',
  '/notifications',
  '/payments',
  '/payments/pay',
  '/building',
];

const authRedirectRoutes = [
  '/affairs',
  '/building/finance',
  '/chat',
  '/saved',
  '/payment',
  '/payment/unit',
  '/payment/order?mch_order_no=PAY-001',
  '/payment/result?mch_order_no=PAY-001',
  '/payment/h5/redirect?mch_order_no=PAY-001',
  '/payments/overview',
  '/payments/bills',
  '/payments/cart',
  '/payments/orders?mch_order_no=PAY-001',
  '/payments/records',
  '/payments/history',
  '/payments/accounting',
  '/payments/units',
  '/properties/my',
  '/properties/my/new',
  '/properties/my/editor/property-1',
  '/serviced-residences/my',
  '/serviced-residences/my/new',
  '/serviced-residences/my/editor/serviced-1',
  '/marketplace/publish',
  '/marketplace/my-listings',
  '/marketplace/my-listings/editor/listing-1',
  '/marketplace/my-listings/preview/listing-1',
  '/marketplace/chat/chat-1',
  '/marketplace/my',
  '/marketplace/my/profile',
  '/marketplace/settings/login-hero',
  '/management',
  '/management/property-sales',
  '/marketplace/management',
  '/marketplace/management/members',
  '/account/profile/info',
  '/account/wallet',
  '/account/notifications',
  '/account/properties',
  '/account/listings/preview/listing-1',
  '/account/marketplace/my/listings',
  '/account/marketplace/my/new',
  '/account/marketplace/my/editor/listing-1',
  '/account/marketplace/my/listing/listing-1',
  '/account/marketplace/my/preview/listing-1',
  '/account/marketplace/my/orders',
  '/account/marketplace/my/favorites',
  '/account/marketplace/my/chat/chat-1',
  '/account/marketplace/management/secondhand-listings',
  '/account/marketplace/management/property-sales',
  '/account/marketplace/management/members',
  '/account/marketplace/management/system-notices',
  '/account/marketplace/management/serviced-apartments',
  '/account/marketplace/management/reward-ads',
  '/account/marketplace/management/wallet-transactions',
  '/account/marketplace/management/wallet-grants',
];

const authRoutes = [
  ...authPageRoutes,
  ...managementTabRoutes,
  ...authRedirectRoutes,
];

const focusRoutes = [
  '/login',
  '/forgot-password',
  '/properties',
  '/furniture',
  '/supermarket-offers',
  '/account/profile/wallet',
  '/account/chat/chat-1',
  '/account/properties/sale/new',
  '/account/listings/new',
  '/account/marketplace/management/reward-ad-editor',
  '/account/marketplace/settings/notice',
  '/account/marketplace/settings/home-hero-cards',
  '/account/marketplace/settings/login-hero',
  '/account/marketplace/settings/display-ads',
  ...managementTabRoutes,
  '/payments/pay',
];

const touchScrollProbeRoutes = new Set([
  '/',
  '/properties',
  '/furniture',
  '/supermarket-offers',
  '/serviced-residences',
  '/trend',
  '/profile',
  '/building',
  '/payments',
  '/account/marketplace/management',
]);

const alternateViewportRoutes = [
  { path: '/', authenticated: false },
  { path: '/login', authenticated: false },
  { path: '/properties', authenticated: false },
  { path: '/furniture', authenticated: false },
  { path: '/supermarket-offers', authenticated: false },
  { path: '/serviced-residences', authenticated: false },
  { path: '/profile' },
  { path: '/building' },
  { path: '/payments/pay' },
  { path: '/account/chat/chat-1' },
  { path: '/account/properties/sale/new' },
  { path: '/account/listings/new' },
  { path: '/account/marketplace/management' },
  { path: '/account/marketplace/settings/login-hero' },
];

const alternateViewportFilterSheetRoutes = [
  { path: '/properties', sheetSelector: '.lf.is-filter-open', label: 'property mobile filter' },
  { path: '/furniture', sheetSelector: '.mf.is-filter-open', label: 'furniture mobile filter' },
];

const guestProtectedRoutes = [
  '/profile',
  '/account',
  '/account/profile/wallet',
  '/account/chat',
  '/account/properties/sale/new',
  '/account/listings/new',
  '/building',
  '/notifications',
  '/payments',
  '/payments/pay',
  '/account/marketplace/management',
  '/account/marketplace/settings/login-hero',
];

const memberStaffRoutes = [
  '/account/marketplace/management',
  '/account/marketplace/management?tab=admin-furniture',
  '/account/marketplace/management?tab=admin-members',
  '/account/marketplace/management/reward-ad-editor',
  '/account/marketplace/management/reward-ad-editor/task-1',
  '/account/marketplace/settings',
  '/account/marketplace/settings/notice',
  '/account/marketplace/settings/home-hero-cards',
  '/account/marketplace/settings/login-hero',
  '/account/marketplace/settings/display-ads',
  '/management',
  '/marketplace/management',
  '/marketplace/management/members',
  '/marketplace/settings/login-hero',
  '/account/marketplace/management/members',
  '/account/marketplace/management/system-notices',
  '/account/marketplace/management/reward-ads',
];

// 1.1 讀取真實 API 的第一條列表資料
const fetchLiveFirstItem = async (path) => {
  if (shouldMockApi) {
    return null;
  }

  try {
    const response = await fetch(`${baseUrl}${path}`);
    if (!response.ok) {
      return null;
    }

    const body = await response.json();
    return body?.data?.items?.[0] ?? null;
  } catch {
    return null;
  }
};

// 1.2 依真實資料生成公開路由
const resolveLivePublicRoutes = async () => {
  if (shouldMockApi) {
    return { pages: publicPageRoutes, redirects: publicRedirectRoutes };
  }

  const [propertyItem, secondhandItem, servicedItem, supermarketItem] = await Promise.all([
    fetchLiveFirstItem('/api/v1/property-sales?page=1&page_size=1'),
    fetchLiveFirstItem('/api/v1/listings?page=1&page_size=1'),
    fetchLiveFirstItem('/api/v1/serviced-apartments?page=1&page_size=1'),
    fetchLiveFirstItem('/api/v1/supermarket-offers/search?page=1&pageSize=1'),
  ]);
  const propertyID = propertyItem?.listing_id ? encodeURIComponent(propertyItem.listing_id) : '';
  const secondhandID = secondhandItem?.listing_id ? encodeURIComponent(secondhandItem.listing_id) : '';
  const sellerID = secondhandItem?.owner?.public_id ? encodeURIComponent(secondhandItem.owner.public_id) : '';
  const servicedID = servicedItem?.listing_id ? encodeURIComponent(servicedItem.listing_id) : '';
  const supermarketCode = supermarketItem?.code ? encodeURIComponent(supermarketItem.code) : '';

  return {
    pages: [
      '/',
      '/login',
      '/forgot-password',
      '/properties',
      propertyID ? `/properties/${propertyID}` : '',
      '/furniture',
      secondhandID ? `/furniture/${secondhandID}` : '',
      '/marketplace',
      '/marketplace/filter',
      secondhandID ? `/marketplace/listing/${secondhandID}` : '',
      sellerID ? `/marketplace/seller/${sellerID}` : '',
      '/supermarket-offers',
      supermarketCode ? `/supermarket-offers/products/${supermarketCode}` : '',
      '/serviced-residences',
      servicedID ? `/serviced-residences/${servicedID}` : '',
      '/trend',
    ].filter(Boolean),
    redirects: [
      '/listing',
      '/detail',
      '/service',
      '/service-detail',
      '/serviced-residence',
      '/serviced-residence/booking',
      servicedID ? `/serviced-residence/${servicedID}` : '',
      '/market',
      '/market-detail',
      secondhandID ? `/furniture/listing/${secondhandID}` : '',
      '/offers',
      '/trend/district/central',
    ].filter(Boolean),
  };
};

const privateGuestNavigationLabels = [
  'AJO Pay',
  '我的大廈',
  '會員中心',
  '通知中心',
  '管理',
];
const privateMemberNavigationLabels = privateGuestNavigationLabels.filter((label) => label !== '管理');

const interactionFlows = [
  {
    name: 'login-password-submit',
    path: '/login',
    authenticated: false,
    action: async (page) => {
      await fillFirstVisible(page, 'input[type="tel"]', '60000000');
      await fillFirstVisible(page, 'input[type="password"]', 'Password123!');
      await clickFirstVisible(page, '.login-submit-button', 'login submit');
    },
  },
  {
    name: 'forgot-password-submit',
    path: '/forgot-password',
    authenticated: false,
    action: async (page) => {
      await fillFirstVisible(page, 'input[type="email"]', 'resident@example.com');
      await clickFirstVisible(page, '.forgot-password-secondary-button', 'request reset code');
      await fillFirstVisible(page, 'input[inputmode="numeric"]', '123456');
      await fillAllVisible(page, 'input[type="password"]', ['Password123!', 'Password123!']);
      await clickFirstVisible(page, '.forgot-password-panel button[type="submit"]', 'reset password submit');
    },
  },
  {
    name: 'chat-send-message',
    path: '/account/chat/chat-1',
    action: async (page) => {
      await fillFirstVisible(page, '.chat-composer textarea', '手機端測試訊息');
      await clickFirstVisible(page, '.chat-send-button', 'chat send');
    },
  },
  {
    name: 'payment-confirm',
    path: '/payments/pay',
    action: async (page) => {
      await page.waitForSelector('.ajo-pay-checkout__confirm:not([disabled])', { timeout: 5000 });
      await clickFirstVisible(page, '.ajo-pay-checkout__confirm:not(:disabled)', 'payment confirm');
    },
  },
  {
    name: 'wallet-recharge-dialog',
    path: '/account/profile/wallet',
    action: async (page) => {
      await clickFirstVisible(page, '.wallet-recharge-panel .wallet-recharge-submit', 'open recharge dialog');
      await fillFirstVisible(page, '.wallet-recharge-dialog input[type="number"]', '100');
      await clickFirstVisible(page, '.wallet-recharge-dialog .wallet-recharge-submit:not(:disabled)', 'create recharge order');
    },
  },
  {
    name: 'property-editor-next-step',
    path: '/account/properties/sale/new',
    action: async (page) => {
      await clickFirstVisible(page, '.property-editor-action--primary:not(:disabled)', 'property editor primary action');
    },
  },
  {
    name: 'secondhand-editor-next-step',
    path: '/account/listings/new',
    action: async (page) => {
      await clickFirstVisible(page, '.editor-action--primary:not(:disabled)', 'secondhand editor primary action');
    },
  },
  {
    name: 'notice-publish',
    path: '/account/marketplace/settings/notice',
    action: async (page) => {
      await fillFirstVisible(page, '.notice-field input', '手機端通知測試');
      await fillFirstVisible(page, '.notice-field textarea', '這是一則手機端發布測試。');
      await clickFirstVisible(page, '.notice-button--primary:not(:disabled)', 'publish notice');
    },
  },
  {
    name: 'login-hero-settings-action',
    path: '/account/marketplace/settings/login-hero',
    action: async (page) => {
      await clickFirstVisible(page, '.login-hero-add', 'add login hero item');
    },
  },
  {
    name: 'display-ads-save',
    path: '/account/marketplace/settings/display-ads',
    action: async (page) => {
      await clickFirstVisible(page, '.display-ads-tabs button:nth-child(2)', 'display ads tab');
      await clickFirstVisible(page, '.display-ads-button--primary:not(:disabled)', 'display ads save');
    },
  },
  {
    name: 'bottom-nav-primary-switching',
    path: '/',
    liveSafe: true,
    action: async (page) => {
      await clickVisibleSequence(page, '.bottom-nav .bnav-item', 'bottom navigation');
    },
  },
  {
    name: 'management-tab-switching',
    path: '/account/marketplace/management',
    liveSafe: true,
    action: async (page) => {
      await clickVisibleSequence(page, '.work-nav-item', 'management tabs');
    },
  },
  {
    name: 'guest-mobile-drawer-close',
    path: '/',
    authenticated: false,
    liveSafe: true,
    action: async (page) => {
      await openMobileDrawer(page);
      await clickFirstVisible(page, '.mobile-close', 'mobile drawer close');
      await waitForMobileDrawerHidden(page);
    },
  },
  {
    name: 'guest-mobile-drawer-backdrop-close',
    path: '/',
    authenticated: false,
    liveSafe: true,
    action: async (page) => {
      await openMobileDrawer(page);
      const [drawerBox, panelBox] = await Promise.all([
        page.locator('.mobile-menu').boundingBox(),
        page.locator('.mobile-menu-panel').boundingBox(),
      ]);

      if (!drawerBox || !panelBox || panelBox.width >= drawerBox.width) {
        throw new Error('Mobile drawer should leave a clickable backdrop');
      }

      await page.mouse.click(
        Math.max(panelBox.x + panelBox.width + 12, drawerBox.x + drawerBox.width - 20),
        drawerBox.y + Math.min(drawerBox.height / 2, 180),
      );
      await waitForMobileDrawerHidden(page);
    },
  },
  {
    name: 'guest-mobile-drawer-login-route',
    path: '/',
    authenticated: false,
    liveSafe: true,
    action: async (page) => {
      await openMobileDrawer(page);
      await clickVisibleByText(page, '.mobile-menu .mnl', '登入', 'mobile drawer login');
      await page.waitForURL(/\/login(?:$|\?)/, { timeout: 5000 });
    },
  },
  {
    name: 'guest-mobile-header-login-route',
    path: '/',
    authenticated: false,
    liveSafe: true,
    action: async (page) => {
      await clickFirstVisible(page, '.mobile-login', 'mobile header login');
      await page.waitForURL(/\/login(?:$|\?)/, { timeout: 5000 });
    },
  },
  {
    name: 'guest-mobile-header-locale-toggle',
    path: '/',
    authenticated: false,
    liveSafe: true,
    action: async (page) => {
      const initialLocale = await page.locator('html').getAttribute('lang');
      if (!['zh-HK', 'en'].includes(initialLocale ?? '')) {
        throw new Error(`Unsupported initial document locale: ${initialLocale ?? 'missing'}`);
      }

      const initialNavigationText = ((await page.locator('.nav-links').textContent()) ?? '').replace(/\s+/g, ' ').trim();
      await clickFirstVisible(page, '.locale-toggle', 'mobile header locale toggle');
      await page.waitForFunction((locale) => document.documentElement.lang !== locale, initialLocale, { timeout: 5000 });

      const switchedLocale = await page.locator('html').getAttribute('lang');
      const switchedNavigationText = ((await page.locator('.nav-links').textContent()) ?? '').replace(/\s+/g, ' ').trim();
      const storedSwitchedLocale = await page.evaluate(() => localStorage.getItem('ajoliving.locale'));
      if (!switchedLocale || switchedLocale === initialLocale) {
        throw new Error('Locale toggle did not produce a supported language change');
      }
      if (!switchedNavigationText || switchedNavigationText === initialNavigationText) {
        throw new Error('Locale toggle changed document.lang without translating navigation labels');
      }
      if (storedSwitchedLocale !== switchedLocale) {
        throw new Error(`Locale preference was not persisted: expected ${switchedLocale}, got ${storedSwitchedLocale ?? 'missing'}`);
      }

      await clickFirstVisible(page, '.locale-toggle', 'mobile header locale reset');
      await page.waitForFunction((locale) => document.documentElement.lang === locale, initialLocale, { timeout: 5000 });
      const storedResetLocale = await page.evaluate(() => localStorage.getItem('ajoliving.locale'));
      if (storedResetLocale !== initialLocale) {
        throw new Error(`Locale reset was not persisted: expected ${initialLocale}, got ${storedResetLocale ?? 'missing'}`);
      }
    },
  },
  {
    name: 'guest-property-mobile-filter-sheet',
    path: '/properties',
    authenticated: false,
    liveSafe: true,
    action: async (page) => {
      await clickFirstVisible(page, '.mobile-filter-button', 'property mobile filter');
      await page.waitForSelector('.lf.is-filter-open', { state: 'visible', timeout: 5000 });
      await assertMobileFilterSheetBounds(page, '.lf.is-filter-open', 'property mobile filter');
      await clickFirstVisible(page, '.filter-sheet-close', 'property filter close');
      await page.waitForSelector('.lf.is-filter-open', { state: 'hidden', timeout: 5000 });
    },
  },
  {
    name: 'guest-furniture-mobile-filter-sheet',
    path: '/furniture',
    authenticated: false,
    liveSafe: true,
    action: async (page) => {
      await clickFirstVisible(page, '.mobile-filter-button', 'furniture mobile filter');
      await page.waitForSelector('.mf.is-filter-open', { state: 'visible', timeout: 5000 });
      await assertMobileFilterSheetBounds(page, '.mf.is-filter-open', 'furniture mobile filter');
      await clickFirstVisible(page, '.filter-sheet-close', 'furniture filter close');
      await page.waitForSelector('.mf.is-filter-open', { state: 'hidden', timeout: 5000 });
    },
  },
  {
    name: 'staff-mobile-drawer-menu-route',
    path: '/',
    liveSafe: true,
    action: async (page) => {
      await openMobileDrawer(page);
      await clickVisibleByText(page, '.mobile-menu .mnl', '我的大廈', 'mobile drawer building entry');
      await page.waitForURL(/\/building(?:$|\?)/, { timeout: 5000 });
    },
  },
  {
    name: 'staff-mobile-drawer-theme-toggle',
    path: '/',
    liveSafe: true,
    action: async (page) => {
      await openMobileDrawer(page);
      const beforeTheme = await page.locator('html').getAttribute('data-theme');
      await clickFirstVisible(page, '.mobile-menu-theme', 'mobile drawer theme toggle');
      await page.waitForFunction((theme) => document.documentElement.getAttribute('data-theme') !== theme, beforeTheme, { timeout: 5000 });
      await clickFirstVisible(page, '.mobile-close', 'mobile drawer close after theme toggle');
      await waitForMobileDrawerHidden(page);
    },
  },
  {
    name: 'staff-mobile-drawer-signout',
    path: '/profile',
    action: async (page) => {
      await openMobileDrawer(page);
      await clickFirstVisible(page, '.mobile-menu-signout', 'mobile drawer sign out');
      await page.waitForURL(/\/login(?:$|\?)/, { timeout: 5000 });
    },
  },
];

// 2. 回傳能讓主要頁面穩定渲染的最小 API 資料
const mockData = (url, method, currentMember = member) => {
  const { pathname } = new URL(url);

  if (pathname === '/api/v1/me' || pathname === '/api/v1/me/profile' || pathname === '/api/v1/me/ismart/bind') return currentMember;
  if (pathname.includes('/auth/') && pathname.includes('/request')) return { sent: true, expires_in: 60 };
  if (pathname.includes('/auth/') && (pathname.includes('/login') || pathname.includes('/verify') || pathname.includes('/register') || pathname.includes('/reset'))) return { access_token: 'mock-access', refresh_token: 'mock-refresh', member: currentMember };
  if (pathname === '/api/v1/auth/logout') return { logged_out: true };
  if (pathname === '/api/v1/meta/communities') return { items: [community] };
  if (pathname === '/api/v1/channel-home/overview') return {
    channels: [
      { code: 'properties', title: '樓盤放售', description: '正式放盤與會員查詢' },
      { code: 'furniture', title: '二手家具', description: '住戶可用二手交易' },
    ],
    featured_secondhand: [secondhand],
  };
  if (pathname === '/api/v1/home/settings/module-cards') return { cards: [] };
  if (pathname === '/api/v1/home/login-hero' || pathname === '/api/v1/home/settings/login-hero') return { items: [] };
  if (pathname === '/api/v1/market-trends/rent') return { series: [], districts: [] };
  if (pathname === '/api/v1/notifications/unread-count') return { unread_count: 1 };
  if (pathname === '/api/v1/notifications') return { items: [], pagination };
  if (pathname === '/api/v1/staff/system-notices') return { delivered_count: 1 };
  if (pathname === '/api/v1/listings') return method === 'GET' ? paginated([secondhand]) : secondhand;
  if (pathname.match(/^\/api\/v1\/listings\/[^/]+$/)) return secondhand;
  if (pathname.includes('/listings/') && pathname.endsWith('/chats')) return { chat_id: 'chat-1', is_new: false };
  if (pathname.includes('/listings/') && pathname.endsWith('/contact-access')) return { listing_id: 'listing-1', allowed_channels: { phone: true, whatsapp: true, chat: true }, contact_payload: { phone: '60000000', email: 'seller@example.com' } };
  if (pathname.includes('/listings/') && pathname.endsWith('/favorite')) return { listing_id: 'listing-1', is_favorited: true };
  if (pathname.startsWith('/api/v1/me/secondhand/listings')) return paginated([secondhand]);
  if (pathname.startsWith('/api/v1/me/secondhand/favorites')) return paginated([secondhand]);
  if (pathname.startsWith('/api/v1/secondhand/settings/listings')) return paginated([secondhand]);
  if (pathname === '/api/v1/property-sales') return method === 'GET' ? paginated([property]) : property;
  if (pathname.match(/^\/api\/v1\/property-sales\/[^/]+$/)) return property;
  if (pathname.includes('/property-sales/') && pathname.endsWith('/chats')) return { chat_id: 'chat-1', is_new: false };
  if (pathname.includes('/property-sales/') && pathname.endsWith('/contact-access')) return { listing_id: 'property-1', allowed_channels: { phone: true, whatsapp: true, chat: true }, contact_payload: { phone: '60000000', email: 'agent@example.com' } };
  if (pathname.startsWith('/api/v1/me/property-sales')) return paginated([property]);
  if (pathname.startsWith('/api/v1/property-sales/featured')) return { items: [property] };
  if (pathname.startsWith('/api/v1/serviced-apartments')) {
    if (pathname.match(/^\/api\/v1\/serviced-apartments\/[^/]+$/)) return serviced;
    return method === 'GET' ? paginated([serviced]) : serviced;
  }
  if (pathname.startsWith('/api/v1/me/serviced-apartments')) return paginated([serviced]);
  if (pathname === '/api/v1/staff/property-sales') return paginated([property]);
  if (pathname.startsWith('/api/v1/staff/property-sales/')) return property;
  if (pathname === '/api/v1/staff/serviced-apartments') return paginated([serviced]);
  if (pathname.startsWith('/api/v1/staff/serviced-apartments/')) return serviced;
  if (pathname.startsWith('/api/v1/staff/secondhand/listings')) return paginated([secondhand]);
  if (pathname === '/api/v1/staff/users') return paginated([member]);
  if (pathname === '/api/v1/staff/roles') return { items: [{ code: 'staff', scope: 'global', name: 'Staff', description: 'Staff', permissions: ['marketplace:manage'] }] };
  if (pathname === '/api/v1/chats/chat-1/messages' && method === 'POST') return {
    message_id: 'msg-2',
    sender_user_id: 'staff-member',
    sender_display_name: 'Staff Member',
    content: '手機端測試訊息',
    message_type: 'text',
    status: 'sent',
    created_at: now,
  };
  if (pathname.startsWith('/api/v1/chats/chat-1/messages')) return {
    items: [{
      message_id: 'msg-1',
      sender_user_id: 'seller-1',
      sender_display_name: 'Demo Seller',
      content: '你好',
      message_type: 'text',
      status: 'sent',
      created_at: now,
    }],
    pagination,
  };
  if (pathname === '/api/v1/chats') return paginated([chat]);
  if (pathname === '/api/v1/chats/chat-1') return chat;
  if (pathname.endsWith('/read')) return { read: true, chat_id: 'chat-1', notification_id: 'notice-1' };
  if (pathname === '/api/v1/messages') return { message_id: 'msg-2', body: 'ok', created_at: now };
  if (pathname.startsWith('/api/v1/supermarket-offers/summary')) return { stats, categories: [], stores: [], cheapest: [gpProduct], bestDiscounts: [gpProduct], biggestDiffs: [gpProduct], offers: [gpProduct] };
  if (pathname.startsWith('/api/v1/supermarket-offers/search')) return { items: [gpProduct], total: 1, page: 1, pageSize: 10, stats, categories: ['食品'], brands: ['Demo'], stores: ['Demo Store'] };
  if (pathname.startsWith('/api/v1/supermarket-offers/products/')) return { product: gpProduct, summary: { days: 90, lowestList: 12, highestList: 16, lowestDeal: 12, highestDeal: 16 }, stores: [], history: [], sameBrand: [], sameCategory: [], alertRule: null, isFavorite: false };
  if (pathname.startsWith('/api/v1/me/supermarket-offers/favorites')) return method === 'GET' ? { items: [], pagination } : { ...gpProduct, isFavorite: true };
  if (pathname.startsWith('/api/v1/me/supermarket-offers/price-alerts')) return method === 'GET' ? { items: [] } : { id: 1, productCode: 'P001', productName: '示範商品', priceMode: 'effective', offerRequired: false, enabled: true, createdAt: now, updatedAt: now };
  if (pathname === '/api/v1/me/wallet') return { balance: 1200, available_balance: 1200, frozen_balance: 0, lifetime_earned: 2000, lifetime_spent: 800 };
  if (pathname === '/api/v1/me/wallet/transactions') return paginated([]);
  if (pathname === '/api/v1/me/wallet/ad-tasks') return { items: [] };
  if (pathname.includes('/me/wallet/ad-tasks/')) return { claim_id: 'claim-1', points: 10, status: 'completed' };
  if (pathname === '/api/v1/me/wallet/recharge-orders') return walletOrder;
  if (pathname.startsWith('/api/v1/me/wallet/recharge-orders/')) return walletOrder;
  if (pathname === '/api/v1/staff/wallet/transactions') return paginated([]);
  if (pathname === '/api/v1/staff/wallet/grants') return { transaction_id: 'grant-1', amount: 10 };
  if (pathname === '/api/v1/staff/wallet/reward-ads') return method === 'GET' ? paginated([]) : { task_id: 'task-1', title: '示範廣告', is_active: true };
  if (pathname.startsWith('/api/v1/staff/wallet/reward-ads/')) return { task_id: 'task-1', title: '示範廣告', is_active: true };
  if (pathname === '/api/v1/staff/wallet/display-ad-settings') return { channel: 'furniture', slots: [] };
  if (pathname === '/api/v1/public/ads') return { items: [] };
  if (pathname === '/api/v1/me/orders') return { items: [], pagination };
  if (pathname.startsWith('/api/v1/orders/')) return { order_id: 'order-1', status: 'pending', items: [] };
  if (pathname === '/api/v1/me/payments/pos/bills') return {
    context: {
      building_id: 'B001',
      building_name: '示範大廈',
      unit_id: 'U001',
      unit_label: '12A',
      floor: '12',
      unit: 'A',
    },
    items: [paymentBill],
    building_options: ['B001'],
    selected_building_id: 'B001',
  };
  if (pathname === '/api/v1/me/payments/pos/fees') return {
    items: [{ pay_type: 'POS_WECHAT', markup: '0%' }],
    building_options: ['B001'],
    selected_building_id: 'B001',
  };
  if (pathname === '/api/v1/me/payments/pos/bank-accounts') return {
    items: [{ title: 'Demo Bank', account_no: '123-456-789' }],
    building_options: ['B001'],
    selected_building_id: 'B001',
  };
  if (pathname === '/api/v1/me/payments/pos/orders' && method === 'POST') return {
    mch_order_no: 'PAY-001',
    amount_hkd: 1200,
    final_amount: 120000,
    state: 'pending',
    pay_data: 'https://example.com/payments/PAY-001',
    pay_data_type: 'payurl',
  };
  if (pathname.startsWith('/api/v1/me/payments/pos/')) return { items: [], building_options: ['B001'], selected_building_id: 'B001' };
  if (pathname === '/api/v1/me/ismart/buildings/info') return { building_options: ['B001'], selected_building_id: 'B001', building: { building_id: 'B001', buildname_chi: '示範大廈' }, building_info: {}, documents: { forms: [], floorplans: [], audit_reports: [], financial_reports: [] } };
  if (pathname === '/api/v1/me/ismart/management-fees' || pathname === '/api/v1/me/ismart/other-fees') return { building_options: ['B001'], items: [] };
  if (pathname === '/api/v1/me/ismart/building-notices') return { building_options: ['B001'], notices: [] };
  if (pathname === '/api/v1/me/ismart/building-access') return { building_options: ['B001'], doors: [], recent_records: [] };
  if (pathname === '/api/v1/me/security/icctv/public-cameras') return { building_options: ['B001'], selected_building_id: 'B001', orangepis: [], cameras: [] };
  if (pathname === '/api/v1/me/ismart/subaccounts') return { building_options: ['B001'], items: [], count: 0 };
  if (pathname === '/api/v1/oss/assets') return paginated([]);
  if (pathname === '/api/v1/oss/presign') return { upload_url: 'https://example.com/upload', object_key: 'mock/object.jpg', upload_token: 'token', headers: {} };
  if (pathname === '/api/v1/oss/complete') return { ...image, storage_provider: 'oss', bucket_name: 'mock', object_key: 'mock/object.jpg', mime_type: 'image/jpeg', file_size: 1000, checksum_sha256: 'abc', in_use: true, created_at: now };

  return { items: [], pagination };
};

// 3. 準備瀏覽器與 mock API
const createMobileContext = (browser, scenario = mobileViewportScenarios[0]) =>
  browser.newContext({
    viewport: scenario.viewport,
    isMobile: true,
    hasTouch: true,
    deviceScaleFactor: 3,
    userAgent: 'Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1',
  });

const attachMockApi = async (page, currentMember = member) => {
  if (!shouldMockApi) {
    return;
  }

  await page.route('**/api/v1/**', async (route) => {
    const request = route.request();

    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(api(mockData(request.url(), request.method(), currentMember))),
    });
  });
};

const setStaffSession = async (page) => {
  await setSessionTokens(page, shouldMockApi ? mockSession : liveStaffSession);
};

const setMemberSession = async (page) => {
  await setSessionTokens(page, shouldMockApi ? mockSession : liveMemberSession);
};

const setSessionTokens = async (page, session) => {
  await page.addInitScript((tokens) => {
    localStorage.setItem('ajoliving.access-token', tokens.accessToken);
    if (tokens.refreshToken) {
      localStorage.setItem('ajoliving.refresh-token', tokens.refreshToken);
    }
  }, session);
};

// 4. 判斷是否屬於需要審計的網絡請求
const isApiRequest = (url) => {
  try {
    return new URL(url).pathname.startsWith('/api/v1/');
  } catch {
    return false;
  }
};

const isAuditedFailedRequest = (url, resourceType) => {
  if (isApiRequest(url)) {
    return true;
  }

  if (!['document', 'fetch', 'xhr', 'script', 'stylesheet'].includes(resourceType)) {
    return false;
  }

  try {
    return new URL(url).origin === new URL(baseUrl).origin;
  } catch {
    return false;
  }
};

const shortenUrl = (url) => {
  try {
    const parsed = new URL(url);
    const base = new URL(baseUrl);
    return parsed.origin === base.origin ? `${parsed.pathname}${parsed.search}` : url;
  } catch {
    return url;
  }
};

const isSameOriginUrl = (url) => {
  try {
    return new URL(url).origin === new URL(baseUrl).origin;
  } catch {
    return false;
  }
};

const isExternalResourceConsole = (message) => {
  const location = message.location();
  return message.text().startsWith('Failed to load resource:') &&
    Boolean(location.url) &&
    !isSameOriginUrl(location.url);
};

// 4.1 導航到審計頁面
const gotoAuditPage = async (page, path) => {
  await page.goto(`${baseUrl}${path}`, { waitUntil: 'domcontentloaded', timeout: 20000 });
  await page.waitForLoadState('networkidle', { timeout: 8000 }).catch(() => {});
};

// 4.2 按真實 API 模式篩選互動流
const selectInteractionFlows = (authenticated) =>
  interactionFlows.filter((item) => {
    const authMatches = authenticated
      ? item.authenticated !== false
      : item.authenticated === false;

    if (!authMatches) {
      return false;
    }

    return shouldMockApi || allowLiveMutationFlows || item.liveSafe === true;
  });

// 5. 建立運行時診斷收集器
const createRuntimeCapture = (page) => {
  const diagnostics = {
    pageErrors: [],
    consoleMessages: [],
    requestFailures: [],
    responseFailures: [],
  };

  const onPageError = (error) => {
    diagnostics.pageErrors.push(error.message);
  };
  const onConsole = (message) => {
    if (!['error', 'warning'].includes(message.type())) {
      return;
    }

    if (isExternalResourceConsole(message)) {
      return;
    }

    const location = message.location();
    diagnostics.consoleMessages.push({
      type: message.type(),
      text: message.text().slice(0, 300),
      url: location.url ? shortenUrl(location.url) : '',
      lineNumber: location.lineNumber,
    });
  };
  const onRequestFailed = (request) => {
    if (!isAuditedFailedRequest(request.url(), request.resourceType())) {
      return;
    }

    diagnostics.requestFailures.push({
      method: request.method(),
      resourceType: request.resourceType(),
      url: shortenUrl(request.url()),
      errorText: request.failure()?.errorText ?? 'unknown',
    });
  };
  const onResponse = (response) => {
    if (response.status() < 400 || !isApiRequest(response.url())) {
      return;
    }

    diagnostics.responseFailures.push({
      method: response.request().method(),
      status: response.status(),
      url: shortenUrl(response.url()),
    });
  };

  page.on('pageerror', onPageError);
  page.on('console', onConsole);
  page.on('requestfailed', onRequestFailed);
  page.on('response', onResponse);

  return {
    diagnostics,
    detach: () => {
      page.off('pageerror', onPageError);
      page.off('console', onConsole);
      page.off('requestfailed', onRequestFailed);
      page.off('response', onResponse);
    },
  };
};

// 6. 讀取目前頁面的手機端版面指標
const readMobileMetrics = async (page) =>
  page.evaluate(() => {
    const viewportWidth = window.innerWidth;
    const viewportHeight = window.innerHeight;
    const root = document.documentElement;
    const body = document.body;
    const bottomNav = document.querySelector('.bottom-nav');
    const footer = document.querySelector('.app-footer');
    const header = document.querySelector('.nav');
    const hamburger = document.querySelector('.hamburger');
    const logo = document.querySelector('.nav-logo');
    const themeToggle = document.querySelector('.theme-toggle');
    const localeToggle = document.querySelector('.locale-toggle');
    const navRect = bottomNav?.getBoundingClientRect();
    const footerRect = footer?.getBoundingClientRect();
    const footerStyle = footer ? window.getComputedStyle(footer) : null;
    const visibleText = (body.innerText || '').trim().length;
    const width = Math.max(root.scrollWidth, body.scrollWidth);
    const height = Math.max(root.scrollHeight, body.scrollHeight);
    const overflowX = width - viewportWidth;
    const bottomNavVisible = Boolean(
      navRect &&
      navRect.width > 0 &&
      navRect.height > 0 &&
      navRect.bottom <= viewportHeight + 1 &&
      navRect.top >= viewportHeight - 130,
    );
    const mobileFooterVisible = Boolean(
      viewportWidth <= 1023 &&
      footerRect &&
      footerStyle &&
      footerStyle.display !== 'none' &&
      footerStyle.visibility !== 'hidden' &&
      Number.parseFloat(footerStyle.opacity || '1') > 0 &&
      footerRect.width > 0 &&
      footerRect.height > 0,
    );
    const mobileHeaderLayoutFailures = [];

    if (viewportWidth <= 1023) {
      const headerNodes = [
        ['header', header],
        ['hamburger', hamburger],
        ['logo', logo],
        ['theme', themeToggle],
        ['locale', localeToggle],
      ];
      const nodeRects = new Map();

      headerNodes.forEach(([name, node]) => {
        if (!node) {
          mobileHeaderLayoutFailures.push(`missing ${name}`);
          return;
        }

        const style = window.getComputedStyle(node);
        const rect = node.getBoundingClientRect();
        const isVisible = style.display !== 'none' &&
          style.visibility !== 'hidden' &&
          Number.parseFloat(style.opacity || '1') > 0 &&
          rect.width > 0 &&
          rect.height > 0;

        if (!isVisible) {
          mobileHeaderLayoutFailures.push(`hidden ${name}`);
          return;
        }

        nodeRects.set(name, rect);
      });

      const menuRect = nodeRects.get('hamburger');
      const logoRect = nodeRects.get('logo');
      const themeRect = nodeRects.get('theme');
      const localeRect = nodeRects.get('locale');

      if (menuRect && logoRect && menuRect.right > logoRect.left) {
        mobileHeaderLayoutFailures.push('hamburger is not left of logo');
      }

      if (logoRect && themeRect && themeRect.left < logoRect.right) {
        mobileHeaderLayoutFailures.push('theme is not right of logo');
      }

      if (logoRect && localeRect && localeRect.left < logoRect.right) {
        mobileHeaderLayoutFailures.push('locale is not right of logo');
      }

      if (logoRect && Math.abs((logoRect.left + logoRect.right) / 2 - viewportWidth / 2) > 4) {
        mobileHeaderLayoutFailures.push('logo is not centered');
      }
    }

    return {
      href: location.pathname + location.search,
      viewportWidth,
      viewportHeight,
      width,
      height,
      overflowX,
      visibleText,
      bottomNavVisible,
      mobileFooterVisible,
      mobileHeaderLayoutFailures,
    };
  });

// 7. 讀取手機端觸控目標是否過小
const readTouchTargetFailures = async (page) =>
  page.evaluate(() => {
    const selector = [
      'button',
      'input',
      'select',
      'textarea',
      '[role="button"]',
      '.app-button',
      'a.bnav-item',
      'a.mnl',
      'a[class*="button"]',
      'a[class*="card"]',
      'a[class*="item"]',
      'a[class*="link"]',
    ].join(',');

    const nodes = [...document.querySelectorAll(selector)]
      .filter((node) => !node.classList.contains('sr-only'))
      .filter((node) => !node.closest('[aria-hidden="true"]'))
      .filter((node) => {
        const style = window.getComputedStyle(node);
        const measuringNode = node.matches('input[type="checkbox"], input[type="radio"]')
          ? node.closest('label') || node
          : node;
        const rect = measuringNode.getBoundingClientRect();
        const disabled = node.disabled || node.getAttribute('aria-disabled') === 'true';

        return !disabled &&
          style.display !== 'none' &&
          style.visibility !== 'hidden' &&
          style.pointerEvents !== 'none' &&
          rect.width > 0 &&
          rect.height > 0;
      });

    const failures = [];
    for (const node of nodes) {
      const measuringNode = node.matches('input[type="checkbox"], input[type="radio"]')
        ? node.closest('label') || node
        : node;
      const rect = measuringNode.getBoundingClientRect();
      const className = typeof node.className === 'string' ? node.className : '';
      const isIconOnly = (node.textContent || '').trim().length <= 2 && !node.querySelector('input, select, textarea');
      const minHeight = node.closest('.bottom-nav, .mobile-menu') ? 44 : 36;
      const minWidth = isIconOnly || node.closest('.bottom-nav, .mobile-menu') ? 36 : 28;

      if (rect.height < minHeight || rect.width < minWidth) {
        failures.push({
          tag: node.tagName,
          className: className.split(/\s+/).filter(Boolean).slice(0, 3).join(' '),
          text: (node.textContent || node.getAttribute('aria-label') || '').trim().slice(0, 40),
          width: Math.round(rect.width),
          height: Math.round(rect.height),
          minWidth,
          minHeight,
        });
      }
    }

    return failures.slice(0, 10);
  });

// 8. 讀取頁面底部可操作內容是否被底部導航遮擋
const readBottomClearanceFailures = async (page) =>
  page.evaluate(async () => {
    const nav = document.querySelector('.bottom-nav')?.getBoundingClientRect();
    if (!nav) {
      return [];
    }

    window.scrollTo({ top: document.documentElement.scrollHeight, behavior: 'instant' });
    await new Promise((resolve) => requestAnimationFrame(resolve));
    await new Promise((resolve) => requestAnimationFrame(resolve));

    const selector = [
      'button',
      'a[href]',
      'input',
      'select',
      'textarea',
      '[role="button"]',
      '.app-button',
    ].join(',');
    const failures = [...document.querySelectorAll(selector)]
      .filter((node) => !node.closest('.bottom-nav, [aria-hidden="true"]'))
      .filter((node) => {
        const style = window.getComputedStyle(node);
        const rect = node.getBoundingClientRect();
        const disabled = node.disabled || node.getAttribute('aria-disabled') === 'true';

        return !disabled &&
          style.display !== 'none' &&
          style.visibility !== 'hidden' &&
          style.pointerEvents !== 'none' &&
          rect.width > 0 &&
          rect.height > 0 &&
          rect.bottom > nav.top - 2 &&
          rect.top < window.innerHeight;
      })
      .map((node) => {
        const rect = node.getBoundingClientRect();
        const className = typeof node.className === 'string' ? node.className : '';

        return {
          tag: node.tagName,
          className: className.split(/\s+/).filter(Boolean).slice(0, 3).join(' '),
          text: (node.textContent || node.getAttribute('aria-label') || '').trim().slice(0, 40),
          bottom: Math.round(rect.bottom),
          navTop: Math.round(nav.top),
        };
      });

    return failures.slice(0, 8);
  });

// 9. 讀取手機端真實觸摸拖動是否能推動頁面
const readTouchScrollFailures = async (page) => {
  const before = await page.evaluate(async () => {
    window.scrollTo({ top: 0, behavior: 'instant' });
    await new Promise((resolve) => requestAnimationFrame(resolve));
    await new Promise((resolve) => requestAnimationFrame(resolve));

    const root = document.documentElement;
    const body = document.body;
    const maxScroll = Math.max(root.scrollHeight, body.scrollHeight) - window.innerHeight;
    const bottomNav = document.querySelector('.bottom-nav')?.getBoundingClientRect();
    const startX = Math.round(window.innerWidth / 2);
    const startY = Math.round(Math.min(window.innerHeight - (bottomNav?.height ?? 0) - 80, 760));
    const target = document.elementFromPoint(startX, startY);

    return {
      scrollY: window.scrollY,
      maxScroll,
      startX,
      startY,
      target: target?.tagName ?? '',
      targetClass: typeof target?.className === 'string' ? target.className.split(/\s+/).slice(0, 3).join(' ') : '',
      bodyOverscrollY: window.getComputedStyle(body).overscrollBehaviorY,
      rootOverscrollY: window.getComputedStyle(root).overscrollBehaviorY,
    };
  });

  if (before.maxScroll < 160) {
    return [];
  }

  const client = await page.context().newCDPSession(page);
  try {
    await client.send('Input.dispatchTouchEvent', {
      type: 'touchStart',
      touchPoints: [{ x: before.startX, y: before.startY, radiusX: 4, radiusY: 4, id: 1 }],
    });

    const moveSteps = [60, 120, 180, 240, 300, 360, 420, 480]
      .map((offset) => Math.max(80, before.startY - offset));
    for (const y of moveSteps) {
      await client.send('Input.dispatchTouchEvent', {
        type: 'touchMove',
        touchPoints: [{ x: before.startX, y, radiusX: 4, radiusY: 4, id: 1 }],
      });
      await page.waitForTimeout(24);
    }

    await client.send('Input.dispatchTouchEvent', { type: 'touchEnd', touchPoints: [] });
  } finally {
    await client.detach().catch(() => {});
  }

  await page.waitForTimeout(280);
  const after = await page.evaluate(() => window.scrollY);

  if (after > 24) {
    return [];
  }

  return [{
    startX: before.startX,
    startY: before.startY,
    maxScroll: Math.round(before.maxScroll),
    beforeScrollY: Math.round(before.scrollY),
    afterScrollY: Math.round(after),
    target: before.target,
    targetClass: before.targetClass,
    bodyOverscrollY: before.bodyOverscrollY,
    rootOverscrollY: before.rootOverscrollY,
  }];
};

// 10. 判斷是否需要在此路由執行真實觸摸拖動探針
const shouldProbeTouchScroll = (path, viewportScenario) =>
  viewportScenario === mobileViewportScenarios[0].name &&
  touchScrollProbeRoutes.has(path.split('?')[0]);

// 11. 點擊第一個可用元素
const clickFirstVisible = async (page, selector, label) => {
  const locator = page.locator(selector);
  const count = await locator.count();

  for (let index = 0; index < count; index += 1) {
    const target = locator.nth(index);
    if (await target.isVisible().catch(() => false)) {
      await target.scrollIntoViewIfNeeded();
      await target.click();
      await page.waitForTimeout(160);
      return;
    }
  }

  throw new Error(`Cannot find visible control: ${label}`);
};

// 12. 點擊第一個指定文字的可用元素
const clickVisibleByText = async (page, selector, text, label) => {
  const locator = page.locator(selector, { hasText: text });
  const count = await locator.count();

  for (let index = 0; index < count; index += 1) {
    const target = locator.nth(index);
    if (await target.isVisible().catch(() => false)) {
      await target.scrollIntoViewIfNeeded();
      await target.click();
      await page.waitForLoadState('networkidle', { timeout: 5000 }).catch(() => {});
      await page.waitForTimeout(160);
      return;
    }
  }

  throw new Error(`Cannot find visible text control: ${label}`);
};

// 13. 開啟手機端抽屜導航
const openMobileDrawer = async (page) => {
  await clickFirstVisible(page, '.hamburger', 'mobile drawer trigger');
  await page.waitForSelector('.mobile-menu', { state: 'visible', timeout: 5000 });
};

// 14. 等待手機端抽屜導航關閉
const waitForMobileDrawerHidden = async (page) => {
  await page.waitForSelector('.mobile-menu', { state: 'hidden', timeout: 5000 });
};

// 14.1 驗證手機端篩選面板貼齊可視區底部且操作列完整可見
const assertMobileFilterSheetBounds = async (page, selector, label) => {
  await page.waitForFunction((sheetSelector) => {
    const sheet = document.querySelector(sheetSelector);
    const sheetRect = sheet?.getBoundingClientRect();

    return Boolean(sheetRect)
      && sheetRect.top >= -1
      && Math.abs(sheetRect.bottom - window.innerHeight) <= 1;
  }, selector, { timeout: 1500 });

  const failure = await page.evaluate((sheetSelector) => {
    const sheet = document.querySelector(sheetSelector);
    const actions = sheet?.querySelector('.filter-sheet-actions');
    const sheetStyle = sheet ? window.getComputedStyle(sheet) : null;
    const sheetRect = sheet?.getBoundingClientRect();
    const actionsRect = actions?.getBoundingClientRect();

    if (!sheet || !sheetStyle || !sheetRect || !actionsRect) {
      return 'missing sheet or action controls';
    }

    if (sheetStyle.position !== 'fixed') {
      return `expected fixed sheet, received ${sheetStyle.position}`;
    }

    if (sheetRect.top < -1 || sheetRect.bottom > window.innerHeight + 1) {
      return `sheet bounds ${Math.round(sheetRect.top)}-${Math.round(sheetRect.bottom)} exceed viewport ${window.innerHeight}`;
    }

    if (Math.abs(sheetRect.bottom - window.innerHeight) > 1) {
      return `sheet bottom ${Math.round(sheetRect.bottom)} is not anchored to viewport ${window.innerHeight}`;
    }

    if (actionsRect.top < sheetRect.top - 1 || actionsRect.bottom > sheetRect.bottom + 1) {
      return 'sheet action controls are outside the sheet bounds';
    }

    return '';
  }, selector);

  if (failure) {
    throw new Error(`${label}: ${failure}`);
  }
};

// 15. 填寫第一個可用輸入
const fillFirstVisible = async (page, selector, value) => {
  const locator = page.locator(selector);
  const count = await locator.count();

  for (let index = 0; index < count; index += 1) {
    const target = locator.nth(index);
    if (await target.isVisible().catch(() => false)) {
      await target.scrollIntoViewIfNeeded();
      await target.fill(value);
      return;
    }
  }

  throw new Error(`Cannot find visible input: ${selector}`);
};

// 16. 依序填寫可用輸入
const fillAllVisible = async (page, selector, values) => {
  const locator = page.locator(selector);
  const count = await locator.count();
  let filled = 0;

  for (let index = 0; index < count && filled < values.length; index += 1) {
    const target = locator.nth(index);
    if (await target.isVisible().catch(() => false)) {
      await target.scrollIntoViewIfNeeded();
      await target.fill(values[filled]);
      filled += 1;
    }
  }

  if (filled < values.length) {
    throw new Error(`Cannot fill enough visible inputs: ${selector}`);
  }
};

// 17. 依序點擊一組手機端可見入口
const clickVisibleSequence = async (page, selector, label) => {
  const locator = page.locator(selector);
  const count = await locator.count();
  let clicked = 0;

  for (let index = 0; index < count; index += 1) {
    const target = locator.nth(index);
    if (await target.isVisible().catch(() => false)) {
      await target.scrollIntoViewIfNeeded();
      await target.click();
      await page.waitForLoadState('networkidle', { timeout: 5000 }).catch(() => {});
      await page.waitForTimeout(160);
      clicked += 1;
    }
  }

  if (clicked === 0) {
    throw new Error(`Cannot find visible sequence controls: ${label}`);
  }
};

// 18. 審計單一路由
const auditRoute = async (page, path, viewportScenario = mobileViewportScenarios[0].name) => {
  const runtime = createRuntimeCapture(page);

  await gotoAuditPage(page, path);
  await page.waitForTimeout(120);

  const [result, touchFailures] = await Promise.all([
    readMobileMetrics(page),
    readTouchTargetFailures(page),
  ]);
  const touchScrollFailures = shouldProbeTouchScroll(path, viewportScenario)
    ? await readTouchScrollFailures(page)
    : [];
  const bottomClearanceFailures = await readBottomClearanceFailures(page);
  runtime.detach();

  return {
    path,
    viewportScenario,
    errors: runtime.diagnostics.pageErrors,
    consoleMessages: runtime.diagnostics.consoleMessages,
    requestFailures: runtime.diagnostics.requestFailures,
    responseFailures: runtime.diagnostics.responseFailures,
    touchFailures,
    touchScrollFailures,
    bottomClearanceFailures,
    ...result,
  };
};

// 18.1 審計短視口與橫屏中的公開篩選面板邊界
const auditFilterSheetViewport = async (page, filterSheet, viewportScenario) => {
  const runtime = createRuntimeCapture(page);

  try {
    await gotoAuditPage(page, filterSheet.path);
    await page.waitForTimeout(120);
    await clickFirstVisible(page, '.mobile-filter-button', filterSheet.label);
    await page.waitForSelector(filterSheet.sheetSelector, { state: 'visible', timeout: 5000 });
    await assertMobileFilterSheetBounds(page, filterSheet.sheetSelector, filterSheet.label);
    const result = await readMobileMetrics(page);

    return {
      path: filterSheet.path,
      viewportScenario,
      errors: runtime.diagnostics.pageErrors,
      consoleMessages: runtime.diagnostics.consoleMessages,
      requestFailures: runtime.diagnostics.requestFailures,
      responseFailures: runtime.diagnostics.responseFailures,
      touchFailures: [],
      touchScrollFailures: [],
      bottomClearanceFailures: [],
      ...result,
    };
  } catch (error) {
    return {
      path: filterSheet.path,
      viewportScenario,
      errors: [
        ...runtime.diagnostics.pageErrors,
        error instanceof Error ? error.message : String(error),
      ],
      consoleMessages: runtime.diagnostics.consoleMessages,
      requestFailures: runtime.diagnostics.requestFailures,
      responseFailures: runtime.diagnostics.responseFailures,
      touchFailures: [],
      touchScrollFailures: [],
      bottomClearanceFailures: [],
      overflowX: 0,
      visibleText: 8,
      bottomNavVisible: true,
      mobileFooterVisible: false,
      mobileHeaderLayoutFailures: [],
    };
  } finally {
    runtime.detach();
  }
};

// 19. 讀取未登入狀態下是否洩漏私有手機導航入口
const readGuestNavigationLeakage = async (page) => {
  await openMobileDrawer(page);
  const leakage = await page.evaluate((labels) => {
    const drawerText = document.querySelector('.mobile-menu')?.textContent ?? '';
    const bottomText = document.querySelector('.bottom-nav')?.textContent ?? '';
    const visibleLabels = labels.filter((label) => drawerText.includes(label) || bottomText.includes(label));
    const hasLoginEntry = drawerText.includes('登入') || bottomText.includes('登入');

    return { visibleLabels, hasLoginEntry };
  }, privateGuestNavigationLabels);
  await clickFirstVisible(page, '.mobile-close', 'guest protected drawer close');
  await waitForMobileDrawerHidden(page);

  return leakage;
};

// 20. 讀取普通會員手機導航是否隱藏 staff 入口並保留會員入口
const readMemberNavigationVisibility = async (page) => {
  await openMobileDrawer(page);
  const visibility = await page.evaluate((labels) => {
    const drawerText = document.querySelector('.mobile-menu')?.textContent ?? '';
    const bottomText = document.querySelector('.bottom-nav')?.textContent ?? '';
    const missingLabels = labels.filter((label) => !drawerText.includes(label) && !bottomText.includes(label));
    const hasManagementEntry = drawerText.includes('管理') || bottomText.includes('管理');

    return { missingLabels, hasManagementEntry };
  }, privateMemberNavigationLabels);
  await clickFirstVisible(page, '.mobile-close', 'member protected drawer close');
  await waitForMobileDrawerHidden(page);

  return visibility;
};

// 21. 審計未登入訪問私有手機路由
const auditGuestProtectedRoute = async (page, path) => {
  const runtime = createRuntimeCapture(page);

  try {
    await gotoAuditPage(page, '/');
    await page.evaluate(() => localStorage.clear());
    await gotoAuditPage(page, path);
    await page.waitForTimeout(160);

    const [result, touchFailures, bottomClearanceFailures, navigationLeakage] = await Promise.all([
      readMobileMetrics(page),
      readTouchTargetFailures(page),
      readBottomClearanceFailures(page),
      readGuestNavigationLeakage(page),
    ]);
    const redirectToLogin = result.href.startsWith('/login');

    return {
      path,
      redirectToLogin,
      navigationLeakage,
      errors: runtime.diagnostics.pageErrors,
      consoleMessages: runtime.diagnostics.consoleMessages,
      requestFailures: runtime.diagnostics.requestFailures,
      responseFailures: runtime.diagnostics.responseFailures,
      touchFailures,
      bottomClearanceFailures,
      ...result,
    };
  } catch (error) {
    return {
      path,
      redirectToLogin: false,
      navigationLeakage: { visibleLabels: [], hasLoginEntry: false },
      errors: [
        ...runtime.diagnostics.pageErrors,
        error instanceof Error ? error.message : String(error),
      ],
      consoleMessages: runtime.diagnostics.consoleMessages,
      requestFailures: runtime.diagnostics.requestFailures,
      responseFailures: runtime.diagnostics.responseFailures,
      touchFailures: [],
      bottomClearanceFailures: [],
    };
  } finally {
    runtime.detach();
  }
};

// 22. 審計普通會員訪問 staff 手機路由
const auditMemberStaffRoute = async (page, path) => {
  const runtime = createRuntimeCapture(page);

  try {
    await gotoAuditPage(page, path);
    await page.waitForTimeout(160);

    const [result, touchFailures, bottomClearanceFailures, navigationVisibility] = await Promise.all([
      readMobileMetrics(page),
      readTouchTargetFailures(page),
      readBottomClearanceFailures(page),
      readMemberNavigationVisibility(page),
    ]);
    const redirectToProfile = result.href === '/profile';

    return {
      path,
      redirectToProfile,
      navigationVisibility,
      errors: runtime.diagnostics.pageErrors,
      consoleMessages: runtime.diagnostics.consoleMessages,
      requestFailures: runtime.diagnostics.requestFailures,
      responseFailures: runtime.diagnostics.responseFailures,
      touchFailures,
      bottomClearanceFailures,
      ...result,
    };
  } catch (error) {
    return {
      path,
      redirectToProfile: false,
      navigationVisibility: { missingLabels: [], hasManagementEntry: false },
      errors: [
        ...runtime.diagnostics.pageErrors,
        error instanceof Error ? error.message : String(error),
      ],
      consoleMessages: runtime.diagnostics.consoleMessages,
      requestFailures: runtime.diagnostics.requestFailures,
      responseFailures: runtime.diagnostics.responseFailures,
      touchFailures: [],
      bottomClearanceFailures: [],
    };
  } finally {
    runtime.detach();
  }
};

// 23. 判斷路由或操作流是否存在手機端失敗
const hasMobileAuditFailure = (item) =>
  item.errors.length > 0 ||
  (item.consoleMessages?.length ?? 0) > 0 ||
  (item.requestFailures?.length ?? 0) > 0 ||
  (item.responseFailures?.length ?? 0) > 0 ||
  (item.focusFailures?.length ?? 0) > 0 ||
  (item.touchFailures?.length ?? 0) > 0 ||
  (item.touchScrollFailures?.length ?? 0) > 0 ||
  (item.bottomClearanceFailures?.length ?? 0) > 0 ||
  (item.mobileHeaderLayoutFailures?.length ?? 0) > 0 ||
  item.overflowX > 1 ||
  item.visibleText < 8 ||
  item.mobileFooterVisible ||
  !item.bottomNavVisible;

// 24. 判斷未登入私有路由是否存在手機端失敗
const hasGuestProtectedFailure = (item) =>
  hasMobileAuditFailure(item) ||
  !item.redirectToLogin ||
  !item.navigationLeakage.hasLoginEntry ||
  item.navigationLeakage.visibleLabels.length > 0;

// 25. 判斷普通會員訪問 staff 路由是否存在手機端失敗
const hasMemberStaffFailure = (item) =>
  hasMobileAuditFailure(item) ||
  !item.redirectToProfile ||
  item.navigationVisibility.hasManagementEntry ||
  item.navigationVisibility.missingLabels.length > 0;

// 26. 讀取目前頁面控件滾動後是否仍被底部導航遮擋
const readFocusFailures = async (page) =>
  page.evaluate(async () => {
    const nav = document.querySelector('.bottom-nav')?.getBoundingClientRect();
    if (!nav) {
      return [];
    }

    const candidates = [...document.querySelectorAll('button, [role="button"], input, select, textarea, .app-button')]
      .filter((node) => !node.closest('.bottom-nav'))
      .filter((node) => {
        const style = window.getComputedStyle(node);
        const rect = node.getBoundingClientRect();

        return style.display !== 'none' &&
          style.visibility !== 'hidden' &&
          !node.disabled &&
          rect.width > 8 &&
          rect.height > 8;
      })
      .slice(0, 18);

    const failures = [];
    for (const node of candidates) {
      node.scrollIntoView({ block: 'end', inline: 'nearest' });
      await new Promise((resolve) => requestAnimationFrame(resolve));
      const rect = node.getBoundingClientRect();

      if (rect.bottom > nav.top - 2) {
        failures.push({
          tag: node.tagName,
          text: (node.textContent || node.getAttribute('aria-label') || '').trim().slice(0, 40),
          bottom: rect.bottom,
          navTop: nav.top,
        });
      }
    }

    return failures.slice(0, 6);
  });

// 27. 審計重點控件滾動後是否仍被底部導航遮擋
const auditFocusReachability = async (page, path) => {
  await gotoAuditPage(page, path);
  await page.waitForTimeout(120);

  return readFocusFailures(page);
};

// 28. 審計主要操作流
const auditInteractionFlow = async (page, flow) => {
  const runtime = createRuntimeCapture(page);

  try {
    if (flow.authenticated === false) {
      await gotoAuditPage(page, '/');
      await page.evaluate(() => localStorage.clear());
    }

    await gotoAuditPage(page, flow.path);
    await page.waitForTimeout(160);
    await flow.action(page);
    await page.waitForTimeout(260);

    return {
      name: flow.name,
      path: flow.path,
      errors: runtime.diagnostics.pageErrors,
      consoleMessages: runtime.diagnostics.consoleMessages,
      requestFailures: runtime.diagnostics.requestFailures,
      responseFailures: runtime.diagnostics.responseFailures,
      focusFailures: await readFocusFailures(page),
      touchFailures: await readTouchTargetFailures(page),
      bottomClearanceFailures: await readBottomClearanceFailures(page),
      ...await readMobileMetrics(page),
    };
  } catch (error) {
    return {
      name: flow.name,
      path: flow.path,
      errors: [
        ...runtime.diagnostics.pageErrors,
        error instanceof Error ? error.message : String(error),
      ],
      consoleMessages: runtime.diagnostics.consoleMessages,
      requestFailures: runtime.diagnostics.requestFailures,
      responseFailures: runtime.diagnostics.responseFailures,
      focusFailures: [],
      touchFailures: [],
      bottomClearanceFailures: [],
    };
  } finally {
    runtime.detach();
  }
};

// 29. 審計短視口與橫屏重點路由
const auditAlternateViewports = async (browser) => {
  const viewportResults = [];
  const alternateScenarios = mobileViewportScenarios.slice(1);

  for (const scenario of alternateScenarios) {
    const guestContext = await createMobileContext(browser, scenario);
    const guestPage = await guestContext.newPage();
    await attachMockApi(guestPage);

    for (const route of alternateViewportRoutes.filter((item) => item.authenticated === false)) {
      viewportResults.push(await auditRoute(guestPage, route.path, scenario.name));
    }

    for (const filterSheet of alternateViewportFilterSheetRoutes) {
      viewportResults.push(await auditFilterSheetViewport(guestPage, filterSheet, scenario.name));
    }
    await guestContext.close();

    const staffContext = await createMobileContext(browser, scenario);
    const staffPage = await staffContext.newPage();
    await attachMockApi(staffPage);
    await setStaffSession(staffPage);

    for (const route of alternateViewportRoutes.filter((item) => item.authenticated !== false)) {
      viewportResults.push(await auditRoute(staffPage, route.path, scenario.name));
    }
    await staffContext.close();
  }

  return viewportResults;
};

// 30. 建立空審計結果
const createAuditBuckets = () => ({
  results: [],
  focusFailures: [],
  interactions: [],
  guestProtectedResults: [],
  memberStaffResults: [],
  alternateViewportResults: [],
});

// 31. 審計公開路由與未登入私有入口
const auditGuestSection = async (browser, buckets) => {
  const guestContext = await createMobileContext(browser);
  const guestPage = await guestContext.newPage();
  await attachMockApi(guestPage);
  const guestRoutes = await resolveLivePublicRoutes();

  for (const path of [...guestRoutes.pages, ...guestRoutes.redirects, '/not-existing-mobile-audit']) {
    buckets.results.push(await auditRoute(guestPage, path));
  }

  for (const flow of selectInteractionFlows(false)) {
    buckets.interactions.push(await auditInteractionFlow(guestPage, flow));
  }

  for (const path of guestProtectedRoutes) {
    buckets.guestProtectedResults.push(await auditGuestProtectedRoute(guestPage, path));
  }
  await guestContext.close();
};

// 32. 審計 staff 路由與主要操作
const auditStaffSection = async (browser, buckets) => {
  const staffContext = await createMobileContext(browser);
  const staffPage = await staffContext.newPage();
  await attachMockApi(staffPage);
  await setStaffSession(staffPage);

  for (const path of authRoutes) {
    buckets.results.push(await auditRoute(staffPage, path));
  }

  for (const path of focusRoutes) {
    const failures = await auditFocusReachability(staffPage, path);
    if (failures.length > 0) {
      buckets.focusFailures.push({ path, failures });
    }
  }

  for (const flow of selectInteractionFlows(true)) {
    buckets.interactions.push(await auditInteractionFlow(staffPage, flow));
  }
  await staffContext.close();
};

// 31. 審計普通會員訪問 staff 入口
const auditMemberStaffSection = async (browser, buckets) => {
  const memberContext = await createMobileContext(browser);
  const memberPage = await memberContext.newPage();
  await attachMockApi(memberPage, regularMember);
  await setMemberSession(memberPage);

  for (const path of memberStaffRoutes) {
    buckets.memberStaffResults.push(await auditMemberStaffRoute(memberPage, path));
  }
  await memberContext.close();
};

// 32. 組裝審計輸出
const buildAuditSummary = (buckets) => {
  const {
    results,
    focusFailures,
    interactions,
    guestProtectedResults,
    memberStaffResults,
    alternateViewportResults,
  } = buckets;
  const routeFailures = results.filter(hasMobileAuditFailure);
  const interactionFailures = interactions.filter(hasMobileAuditFailure);
  const guestProtectedFailures = guestProtectedResults.filter(hasGuestProtectedFailure);
  const memberStaffFailures = memberStaffResults.filter(hasMemberStaffFailure);
  const alternateViewportFailures = alternateViewportResults.filter(hasMobileAuditFailure);

  return {
    ok: routeFailures.length === 0 &&
      focusFailures.length === 0 &&
      interactionFailures.length === 0 &&
      guestProtectedFailures.length === 0 &&
      memberStaffFailures.length === 0 &&
      alternateViewportFailures.length === 0,
    checked: results.length,
    alternateViewportChecked: alternateViewportResults.length,
    guestProtectedChecked: guestProtectedResults.length,
    memberStaffChecked: memberStaffResults.length,
    interactionsChecked: interactions.length,
    routeCoverage: {
      publicPages: publicPageRoutes.length,
      publicRedirects: publicRedirectRoutes.length,
      authenticatedPages: authPageRoutes.length,
      managementTabs: managementTabRoutes.length,
      authenticatedRedirects: authRedirectRoutes.length,
      focusRoutes: focusRoutes.length,
      guestProtectedRoutes: guestProtectedRoutes.length,
      memberStaffRoutes: memberStaffRoutes.length,
      alternateViewportRoutes: alternateViewportRoutes.length,
      alternateViewportScenarios: mobileViewportScenarios.length - 1,
    },
    routeFailures,
    focusFailures,
    interactionFailures,
    guestProtectedFailures,
    memberStaffFailures,
    alternateViewportFailures,
    sample: shouldShowSamples ? results.slice(0, 5) : undefined,
  };
};

// 33. 檢查真實 API 模式所需登入 token
const assertLiveSessionReady = () => {
  if (!allowedAuditSections.includes(auditSection)) {
    throw new Error(`Unsupported mobile audit section: ${auditSection}`);
  }

  if (shouldMockApi) {
    return;
  }

  if (['all', 'staff', 'viewport'].includes(auditSection) && !liveStaffSession.accessToken) {
    throw new Error('MOBILE_AUDIT_STAFF_ACCESS_TOKEN or MOBILE_AUDIT_ACCESS_TOKEN is required when auditing staff routes with --mock-api=false');
  }

  if (['all', 'member-staff'].includes(auditSection) && !liveMemberSession.accessToken) {
    throw new Error('MOBILE_AUDIT_MEMBER_ACCESS_TOKEN or MOBILE_AUDIT_ACCESS_TOKEN is required when auditing member staff-route interception with --mock-api=false');
  }
};

// 34. 執行審計
const runAudit = async () => {
  assertLiveSessionReady();

  const browser = await chromium.launch({ headless: true });
  const buckets = createAuditBuckets();

  try {
    if (auditSection === 'all' || auditSection === 'guest') {
      await auditGuestSection(browser, buckets);
    }

    if (auditSection === 'all' || auditSection === 'staff') {
      await auditStaffSection(browser, buckets);
    }

    if (auditSection === 'all' || auditSection === 'member-staff') {
      await auditMemberStaffSection(browser, buckets);
    }

    if (auditSection === 'all' || auditSection === 'viewport') {
      buckets.alternateViewportResults = await auditAlternateViewports(browser);
    }
  } finally {
    await browser.close();
  }

  return buildAuditSummary(buckets);
};

try {
  const result = await runAudit();
  console.log(JSON.stringify(result, null, 2));

  if (!result.ok) {
    process.exitCode = 1;
  }
} catch (error) {
  console.error(error instanceof Error ? error.message : String(error));
  process.exitCode = 1;
}

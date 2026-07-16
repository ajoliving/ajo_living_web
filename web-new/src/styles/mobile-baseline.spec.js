/*
 * 手機端樣式基線測試。
 * 1. 鎖定 iPhone 安全區、底部導航留白與觸控尺寸規則。
 * 2. 鎖定管理、會員、聊天、彈窗與表格在手機端的共用覆蓋。
 */
import { readdirSync, readFileSync, statSync } from 'node:fs';
import { join } from 'node:path';
import { describe, expect, it } from 'vitest';

import { accountRoutes } from '@/router/routes/account';
import { buildingRoutes } from '@/router/routes/building';
import { channelRoutes } from '@/router/routes/channels';
import { firstRoutes } from '@/router/routes/first';
import { homeRoutes } from '@/router/routes/home';
import { marketplaceRoutes } from '@/router/routes/marketplace';
import { notificationRoutes } from '@/router/routes/notifications';
import { paymentRoutes } from '@/router/routes/payments';
import { systemRoutes } from '@/router/routes/system';
import { trendRoutes } from '@/router/routes/trend';

// 1. 讀取全域樣式與 token 檔案
const styles = readFileSync('src/styles/index.css', 'utf8');
const tokens = readFileSync('src/styles/tokens.css', 'utf8');
const appShell = readFileSync('src/shared/components/layout/AppShell.vue', 'utf8');
const mobileAudit = readFileSync('scripts/audit-mobile.mjs', 'utf8');
const propertyListPage = readFileSync('src/pages/property/list/PropertyListPage.vue', 'utf8');
const furnitureListPage = readFileSync('src/pages/furniture/Page.vue', 'utf8');
const offersPage = readFileSync('src/pages/offers/Page.vue', 'utf8');
const collectSourceFiles = (directory) =>
  readdirSync(directory).flatMap((entry) => {
    const entryPath = join(directory, entry);
    const stats = statSync(entryPath);

    if (stats.isDirectory()) {
      return collectSourceFiles(entryPath);
    }

    return /\.(vue|scss|css)$/.test(entryPath) ? [entryPath] : [];
  });
const mobileStyleSource = collectSourceFiles('src/pages')
  .concat(collectSourceFiles('src/shared'))
  .concat(collectSourceFiles('src/styles'))
  .map((filePath) => readFileSync(filePath, 'utf8'))
  .join('\n');
const allRoutes = [
  ...homeRoutes,
  ...buildingRoutes,
  ...firstRoutes,
  ...channelRoutes,
  ...trendRoutes,
  ...notificationRoutes,
  ...paymentRoutes,
  ...marketplaceRoutes,
  ...accountRoutes,
  ...systemRoutes,
];

// 2. 展開真實路由路徑
const normalizeRoutePath = (parentPath, routePath) => {
  if (routePath.startsWith('/')) {
    return routePath;
  }

  const joinedPath = parentPath === '/' ? `/${routePath}` : `${parentPath}/${routePath}`;
  const normalizedPath = joinedPath.replaceAll(/\/+/g, '/');

  return normalizedPath === '' ? '/' : normalizedPath;
};

const collectPageRoutePaths = (routes, parentPath = '') =>
  routes.flatMap((route) => {
    const routePath = normalizeRoutePath(parentPath, route.path);
    const ownPaths = route.component ? [routePath] : [];
    const childPaths = route.children ? collectPageRoutePaths(route.children, routePath) : [];

    return [...ownPaths, ...childPaths];
  });

const pageRoutePaths = new Set(collectPageRoutePaths(allRoutes));

// 3. 鎖定主要功能路由與手機端基線選擇器
const routeMobileContracts = [
  { path: '/', selectors: ['.section-shell', '.home-section'] },
  { path: '/login', selectors: ['.login-shell', '.login-hero-panel'] },
  { path: '/forgot-password', selectors: ['.forgot-password-page'] },
  { path: '/account/profile/agency-profile', selectors: ['.agency-profile-page'] },
  { path: '/properties', selectors: ['.property-list-layout'] },
  { path: '/properties/:listingId', selectors: ['.detail-layout'] },
  { path: '/furniture', selectors: ['.channel-page', '.mgrid'] },
  { path: '/furniture/:listingId', selectors: ['.furniture-detail-page'] },
  { path: '/supermarket-offers', selectors: ['.gp-page'] },
  { path: '/supermarket-offers/products/:code', selectors: ['.offers-detail-page'] },
  { path: '/serviced-residences', selectors: ['.sv-page'] },
  { path: '/serviced-residences/:listingId', selectors: ['.sv-content'] },
  { path: '/trend', selectors: ['.trend-page-wrap'] },
  { path: '/marketplace', selectors: ['.marketplace-layout', '.breadcrumb'] },
  { path: '/marketplace/filter', selectors: ['.marketplace-filter-page'] },
  { path: '/marketplace/listing/:listingId', selectors: ['.marketplace-detail-page'] },
  { path: '/marketplace/seller/:sellerId', selectors: ['.seller-page'] },
  { path: '/notifications', selectors: ['.work-shell', '.work-nav'] },
  { path: '/payments', selectors: ['.ajo-pay-home__hero'] },
  { path: '/payments/pay', selectors: ['.ajo-pay-checkout__offline', '.ajo-pay-checkout__confirm'] },
  { path: '/building', selectors: ['.work-shell', '.building-summary-grid'] },
  { path: '/account', selectors: ['.work-shell'] },
  { path: '/profile', selectors: ['.work-shell'] },
  { path: '/account/profile', selectors: ['.work-shell'] },
  { path: '/account/profile/wallet', selectors: ['.wallet-page', '.wallet-recharge-dialog'] },
  { path: '/account/chat/:conversationId?', selectors: ['.marketplace-chat-page', '.chat-layout'] },
  { path: '/account/properties/sale', selectors: ['.property-my-page', '.property-table-wrap'] },
  { path: '/account/properties/sale/new', selectors: ['.property-editor-grid'] },
  { path: '/account/properties/sale/editor/:listingId?', selectors: ['.property-editor-grid'] },
  { path: '/account/properties/serviced-residences', selectors: ['.property-my-page'] },
  { path: '/account/properties/serviced-residences/new', selectors: ['.property-editor-grid'] },
  { path: '/account/properties/serviced-residences/editor/:listingId?', selectors: ['.property-editor-grid'] },
  { path: '/account/listings', selectors: ['.marketplace-my-page', '.my-table-wrap'] },
  { path: '/account/listings/new', selectors: ['.listing-editor-workspace'] },
  { path: '/account/listings/editor/:listingId?', selectors: ['.listing-editor-workspace'] },
  { path: '/account/listings/item/:listingId', selectors: ['.my-preview-page'] },
  { path: '/account/favorites', selectors: ['.saved-page'] },
  { path: '/account/marketplace/my', selectors: ['.marketplace-my-page'] },
  { path: '/account/marketplace/my/profile', selectors: ['.marketplace-profile-grid'] },
  { path: '/account/marketplace/management', selectors: ['.management-table-wrap', '.management-list-toolbar'] },
  { path: '/account/marketplace/management/reward-ad-editor/:taskId?', selectors: ['.wallet-settings-page'] },
  { path: '/account/marketplace/settings', selectors: ['.settings-shell'] },
  { path: '/account/marketplace/settings/notice', selectors: ['.settings-shell', '.notice-page'] },
  { path: '/account/marketplace/settings/home-hero-cards', selectors: ['.home-hero-page'] },
  { path: '/account/marketplace/settings/login-hero', selectors: ['.login-hero-page'] },
  { path: '/account/marketplace/settings/display-ads', selectors: ['.display-ads-page'] },
  { path: '/:pathMatch(.*)*', selectors: ['.not-found-page'] },
];

describe('mobile baseline styles', () => {
  it('keeps safe-area and bottom navigation spacing tokens', () => {
    expect(tokens).toContain('--app-safe-bottom: env(safe-area-inset-bottom, 0px)');
    expect(tokens).toContain('--app-mobile-bottom-nav-height');
    expect(tokens).toContain('--app-mobile-content-bottom');
    expect(styles).toContain('100svh');
    expect(styles).toContain('@media (max-width: 1023px)');
    expect(styles).toContain('html #app .app-main');
    expect(styles).toContain('padding-bottom: var(--app-mobile-content-bottom)');
  });

  it('keeps mobile layouts usable across page families', () => {
    expect(styles).toContain('@media (max-width: 767px)');
    expect(styles).toContain('.work-shell');
    expect(styles).toContain('.settings-shell');
    expect(styles).toContain('.marketplace-chat-page');
    expect(styles).toContain('.wallet-page');
    expect(styles).toContain('.display-ads-page');
    expect(styles).toContain('grid-template-columns: minmax(0, 1fr)');
  });

  it('keeps scrollable tables and mobile dialogs reachable', () => {
    expect(styles).toContain('.management-table-wrap');
    expect(styles).toContain('.wallet-transaction-table-wrap');
    expect(styles).toContain('-webkit-overflow-scrolling: touch');
    expect(styles).toContain('.management-dialog');
    expect(styles).toContain('.wallet-recharge-dialog');
    expect(styles).toContain('.account-profile-modal');
    expect(styles).toContain('.unsaved-dialog');
    expect(styles).toContain('.property-editor-dialog');
    expect(styles).toContain('.alert-modal-overlay');
    expect(styles).toContain('html body .management-dialog-panel');
    expect(styles).toContain('max-height: calc(100svh');
    expect(styles).toContain("html #app :where(button, [role='button'], input, select, textarea, .app-button)");
    expect(styles).toContain(":where(button, [role='button'], input, select, textarea)");
    expect(styles).toContain('scroll-margin-bottom: calc(var(--app-mobile-content-bottom) + 12px)');
  });

  it('keeps document-level touch scrolling enabled', () => {
    expect(styles).not.toContain('overscroll-behavior-y: none');
    expect(styles).not.toContain('overscroll-behavior: none');
  });

  it('does not place hidden public filter sheets in mobile document flow', () => {
    expect(styles).not.toContain('html #app :where(.lf,');
    expect(styles).not.toContain('html #app :where(.mf,');
  });

  it('keeps supermarket search controls compact when stacked on mobile', () => {
    expect(offersPage).toMatch(/@media \(max-width: (?:900|1023)px\)[\s\S]*?\.gp-search-box \{[\s\S]*?flex: 1 1 auto;/);
    expect(offersPage).toContain('class="gp-mobile-filter-select gp-mobile-sort-select"');
    expect(offersPage.indexOf('gp-mobile-sort-select')).toBeLessThan(offersPage.indexOf(":aria-label=\"t('offers.list.category')\""));
    expect(offersPage.indexOf('gp-mobile-sort-select')).toBeLessThan(offersPage.indexOf('class="gp-mobile-filter-button mobile-filter-button"'));
    expect(offersPage).toMatch(/@media \(max-width: 1023px\)[\s\S]*?\.gp-toolbar \{[\s\S]*?display: none;/);
  });

  it('hides the desktop footer throughout mobile navigation layouts', () => {
    expect(appShell).toMatch(/@media \(max-width: 1023px\) \{[\s\S]*?\.app-footer \{\s*display: none;\s*\}/);
  });

  it('keeps public list filters inside compact mobile sheets', () => {
    [propertyListPage, furnitureListPage].forEach((source) => {
      expect(source).toContain('class="mobile-search-toolbar"');
      expect(source).toContain('class="mobile-listing-controls"');
      expect(source).toContain('class="mobile-filter-rail"');
      expect(source).toContain('class="mobile-filter-select"');
      expect(source).toContain('class="mobile-filter-select mobile-sort-select"');
      expect(source).toContain('class="mobile-filter-button"');
      expect(source).toContain('class="mobile-search-form"');
      expect(source).toContain('class="filter-backdrop"');
      expect(source).toContain('class="filter-sheet-close"');
      expect(source).toContain('max-height: min(82svh, 720px)');
      expect(source).toMatch(/@media \(max-width: 1023px\)[\s\S]*?position: fixed;[\s\S]*?top: auto;[\s\S]*?bottom: 0;/);
      expect(source.indexOf('class="mobile-search-form"')).toBeLessThan(source.indexOf('class="mobile-filter-rail"'));
      expect(source.indexOf('mobile-sort-select')).toBeLessThan(source.indexOf('class="mobile-filter-button"'));
      expect(source.indexOf('class="mobile-filter-rail"')).toBeLessThan(source.indexOf('class="sort-row'));
      expect(source).toContain('overflow-x: auto');
    });
  });

  it('keeps member and work pages compact across portrait and landscape mobile layouts', () => {
    expect(styles).toMatch(/@media \(max-width: 1023px\)[\s\S]*?\.work-account-actions \{[\s\S]*?grid-template-columns: repeat\(3, minmax\(0, 1fr\)\)/);
    expect(styles).toMatch(/@media \(max-width: 1023px\)[\s\S]*?\.wallet-summary-grid \{[\s\S]*?grid-template-columns: repeat\(2, minmax\(0, 1fr\)\)/);
    expect(styles).toMatch(/@media \(max-width: 1023px\)[\s\S]*?:where\(\.my-filter-row, \.property-tab-row\) \{[\s\S]*?overflow-x: auto/);
    expect(styles).toMatch(/:where\(\.my-filter-row, \.property-tab-row\) :where\(\.my-filter-chip, \.property-tab\) \{[\s\S]*?border-radius: 999px/);
    expect(styles).toMatch(/@media \(max-width: 1023px\)[\s\S]*?\.work-sidebar[\s\S]*?border-bottom: 1px solid var\(--bdr\)/);
  });

  it('keeps the requested property and furniture filters in the mobile quick rail', () => {
    expect(propertyListPage).toContain("['region', 'property_type', 'bedroom', 'renovation']");
    expect(furnitureListPage).toContain("@change=\"handleMobileFilterChange('area', $event)\"");
    expect(furnitureListPage).toContain("@change=\"handleMobileFilterChange('condition', $event)\"");
  });

  it('waits for a mobile filter sheet to finish its entry transition before checking its bounds', () => {
    expect(mobileAudit).toMatch(/const assertMobileFilterSheetBounds = async \(page, selector, label\) => \{[\s\S]*?await page\.waitForFunction[\s\S]*?sheetRect\.bottom/);
  });

  it('keeps mobile coverage tied to real functional route families', () => {
    const coveredRoutePaths = new Set(routeMobileContracts.map(({ path }) => path));
    const uncoveredRoutePaths = [...pageRoutePaths].filter((path) => !coveredRoutePaths.has(path));

    expect(uncoveredRoutePaths).toEqual([]);

    routeMobileContracts.forEach(({ path, selectors }) => {
      expect(pageRoutePaths.has(path), `${path} route should exist`).toBe(true);

      selectors.forEach((selector) => {
        expect(mobileStyleSource, `${path} should keep ${selector} in the mobile baseline`).toContain(selector);
      });
    });
  });

  it('does not use forceful overrides in the global mobile baseline', () => {
    expect(styles).not.toContain('!important');
  });

  it('keeps page-level mobile viewport and bottom spacing tied to app tokens', () => {
    expect(mobileStyleSource).not.toContain('100vh');
    expect(mobileStyleSource).not.toMatch(/\b(?:min-h|h|max-h)-screen\b/);
    expect(mobileStyleSource).not.toMatch(/padding-bottom:\s*(72px|80px|88px|90px|96px|100px|110px|120px)/);
  });
});

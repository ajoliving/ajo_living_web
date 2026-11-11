/*
 * My Building responsive regression audit.
 * 1. Render compiled pages or an existing frontend without starting a server.
 * 2. Mock all APIs and external media; never perform live mutations.
 * 3. Check the mobile navigation home, content overflow, controls and CCTV geometry.
 */
import { mkdir, readFile, realpath, writeFile } from 'node:fs/promises';
import { dirname, extname, isAbsolute, join, relative, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

import { chromium } from '@playwright/test';

// 1. Read optional command-line settings.
const option = (name, fallback) =>
  process.argv.find((argument) => argument.startsWith(`--${name}=`))?.slice(name.length + 3) ?? fallback;
const webRoot = resolve(dirname(fileURLToPath(import.meta.url)), '..');
const staticMode = process.argv.includes('--static');
const base = staticMode ? 'http://ajo.preview' : option('base', 'http://localhost:5173').replace(/\/+$/, '');
const output = resolve(option('output', join(webRoot, '..', 'outputs', 'mobile-layout', 'building')));
const buildingID = '0999900';
const unitID = '09999000012';
const contextData = { selected_building_id: buildingID, building_options: [buildingID] };
const longTitle = '公共地方維修工程及升降機定期保養安排 / Common Area Maintenance and Lift Inspection';
const documentFile = {
  id: 'file-1', title: longTitle, type: 'PDF',
  file_url: `${base}/mock/document.pdf`, file_date: '2026-10-20',
  created_date: '2026-10-21', file_month: '2026-10',
};
const member = {
  public_id: 'member', display_name: 'Member', email: 'member@example.com',
  phone_country_code: '+852', phone_number: '60000000', member_status: 'active',
  member_type: 'owner', is_staff: false, role: 'member', roles: ['member'], permissions: [],
  publisher_identity_type: 'owner', profile_completed: true,
  primary_community: {
    public_id: buildingID, name_zh: '海景大廈', name_en: 'Seaview Building',
    address_text: 'Hong Kong', community_type: 'building',
  },
  bound_building_ids: [buildingID], bound_flat_unit_ids: [unitID],
  residence_floor: 'G', residence_unit: 'B', ismart_linked: true, ismart_username: 'member',
  ismart_msg: {
    user_id: 1, username: 'member', is_staff: false,
    client_building_permissions: [buildingID], client_building_flat_units_permissions: [unitID],
  },
};
const modules = [
  ['affairs-notices', '.notice-table tbody tr:not(:has(.building-empty-row))'],
  ['affairs-building', '.building-file-table tbody tr:not(:has(.building-empty-row))'],
  ['affairs-finance', '.finance-management-table tbody td[data-label]'],
  ['affairs-owner-account', '.owner-account-table tbody td[data-label]'],
  ['affairs-forms', '.stack-table tbody td[data-label]'],
  ['affairs-feedback', '.stack-table tbody td[data-label]'],
  ['affairs-access', '.access-door-table tbody td[data-label]'],
  ['affairs-icctv', '.icctv-camera-title'],
  ['affairs-resident-authorizations', '.authorization-record'],
];
const mimeTypes = {
  '.html': 'text/html', '.js': 'text/javascript', '.css': 'text/css',
  '.json': 'application/json', '.svg': 'image/svg+xml', '.png': 'image/png',
  '.jpg': 'image/jpeg', '.jpeg': 'image/jpeg', '.webp': 'image/webp',
  '.woff': 'font/woff', '.woff2': 'font/woff2', '.ico': 'image/x-icon',
};

// 2. Provide populated read-only responses using the current API contracts.
const mockData = (path) => {
  if (path === '/api/v1/me') return member;
  if (path === '/api/v1/me/pos/buildings' || path === '/api/v1/pos/buildings') {
    return { items: [{ building_id: buildingID, buildname_chi: '海景大廈', buildname: 'Seaview Building' }] };
  }
  if (/\/pos\/buildings\/[^/]+\/units$/.test(path)) return { items: [{ unit_id: unitID, floor: 'G', unit: 'B' }] };
  if (path === '/api/v1/me/ismart/building-info') return {
    ...contextData,
    building: { building_id: buildingID, buildname_chi: '海景大廈', buildname: 'Seaview Building' },
    building_info: {
      year_built: 1998, total_floor: 28, total_unit: 112, total_carpark: 36,
      owners_corporation_name: '海景大廈業主立案法團',
      management_company_name: 'Seaview Property Management Company Limited',
      management_office_phone: '2123 4567', management_company_phone: '2123 4568',
      management_company_email: 'property.management@example.com',
      management_company_fax: '2123 4569', home_affairs_department_phone: '2835 2500',
    },
    documents: {
      forms: [documentFile], building_info_files: [documentFile], floorplans: [documentFile],
      audit_reports: [documentFile], financial_reports: [documentFile],
    },
  };
  if (path === '/api/v1/me/ismart/notices') return {
    ...contextData,
    result: [
      { id: 1, mess_code: 'BLDG-NOTICE-2026-1029-001', mess_title: longTitle, mess_type: '維修通告 / Maintenance', mess_date: '2026-10-20', mess_down: '2026-12-31', mess_file: documentFile.file_url },
      { id: 2, mess_code: 'BLDG-NOTICE-2026-1029-002', mess_title: '大廈清潔及公用設施例行檢查安排', mess_type: '一般通告', mess_date: '2026-10-22', mess_down: '2026-12-31', mess_file: documentFile.file_url },
    ],
  };
  if (path === '/api/v1/me/ismart/management-fees') return {
    ...contextData,
    result: [{
      flat_code: 'G/B', invoice_no: 'MANAGEMENT-2026-10-0001', item_id: 'Monthly Management Fee',
      trs_to: '2026/10', trs_val: 1200, remark: longTitle,
      'Outstanding Balance Carried Forward': 1800,
    }],
  };
  if (path === '/api/v1/me/ismart/other-fees') return {
    ...contextData,
    result: [{ invoice_no: 'OTHER-2026-10-0001', flat_code: 'G/B', item_id: '公共地方維修費', trs_to: '2026/10', trs_val: 300, remark: longTitle }],
  };
  if (path === '/api/v1/me/payments/pos/bills') return {
    items: [{ invoice_no: 'INVOICE-2026-10-0001', flat_code: unitID, item_id: 'Monthly Management Fee', trs_to: '2026/10', bill_dt: '2026-10-01', net_amount: 1200, remark: longTitle }],
  };
  if (path === '/api/v1/me/payments/pos/history') return {
    items: [{
      payment_id: 'PAYMENT-2026-10-0001', receipt_id: 'RECEIPT-2026-10-0001',
      tran_time: '2026-10-21 15:30:00', input_time: '2026-10-21 15:32:00',
      pay_type: 'POS_CHEQUE', status: 'completed',
      payment_detail_objs: [{ floor: 'G', unit: 'B', flat_code: unitID, item_id: 'Monthly Management Fee', term: '2026/10', trs_val: 1200, remark: longTitle }],
    }],
  };
  if (path === '/api/v1/me/ismart/building-access') return {
    ...contextData,
    doors: [{
      door_id: 'door-1', title: '地下大堂主要出入口 / Ground Floor Main Entrance',
      serial: 'DOOR-0001', door_no: '1', building_id: buildingID,
      is_public: true, is_qrcode_enabled: true, has_permission: true,
      camera: { id: 'door-camera', title: 'Ground Floor Entrance', url: `${base}/mock/camera`, source: 'mock' },
      password: { record_id: 1, value: '123456', start_time: '2026-10-01', end_time: '2026-12-31' },
      qrcode: { record_id: 1, start_time: '2026-10-01', end_time: '2026-12-31' },
    }],
    recent_records: [{ door: { id: 'door-1', title: 'Ground Floor Main Entrance' }, records: [{ open_time: '2026-10-29 09:30:00', open_type: 'Remote', is_success: true }] }],
  };
  if (path === '/api/v1/me/security/icctv/public-cameras') return {
    ...contextData,
    orangepis: [{ orangepi_id: 1, orangepi_name: 'device-1', is_active: true, camera_count: 1 }],
    cameras: [{ id: 'camera-1', title: '地下大堂主要出入口 / Ground Floor Main Entrance', channel: 'CH-01', url: `${base}/mock/camera`, orangepi_id: 1, orangepi_name: 'device-1', is_active: true }],
  };
  if (path === '/api/v1/me/ismart/subaccounts') return {
    ...contextData, count: 1,
    items: [{ relation_info_id: 1, unit_id: unitID, target_user_id: 2, target_username: 'resident', target_name: 'Alexandra Resident', target_phone: '+85260000001', target_email: 'alexandra.resident.long.address@example.com', role: 'subaccount', status: 'active' }],
  };
  if (path === '/api/v1/me/ismart/subaccount-authorizations') return {
    ...contextData, count: 1,
    items: [{ authorization_id: 'pending-1', unit_id: unitID, phone: '+85260000002', email: 'another.resident.long.address@example.com', target_name: 'Another Resident', status: 'awaiting_confirmation' }],
  };
  if (path === '/api/v1/me/ismart/service-cases') return {
    ...contextData, total: 1, returned: 1,
    status_choices: [{ code: 'processing', label: '處理中 / Processing' }],
    cases: [{ case_id: 'case-1', case_no: 'SERVICE-2026-10-0001', building_id: buildingID, unit_id: unitID, subject: longTitle, request_type: 'repair', category_label: '公共設施維修 / Common Facilities', status: 'processing', status_label: '處理中 / Processing', created_at: '2026-10-21T15:30:00+08:00', updated_at: '2026-10-22T15:30:00+08:00' }],
  };
  if (path === '/api/v1/me/ismart/service-cases/case-1') return {
    case_id: 'case-1', subject: longTitle, content: longTitle,
    messages: [{ message_id: 'reply-1', author_username: 'Management Office', body: longTitle, created_at: '2026-10-22T15:30:00+08:00' }],
  };
  if (path.endsWith('/unread-count')) return { unread_count: 0 };
  if (path === '/api/v1/chats' || path === '/api/v1/notifications' || path === '/api/v1/public/ads') return { items: [], pagination: { page: 1, page_size: 20, total: 0 } };
  return undefined;
};

// 3. Serve only files inside dist, including symlink containment checks.
const serveAsset = async (route, pathname, distRoot) => {
  const requested = pathname === '/' || !extname(pathname) ? 'index.html' : `.${decodeURIComponent(pathname)}`;
  const candidate = resolve(distRoot, requested);
  const allowed = (path) => {
    const difference = relative(distRoot, path);
    return difference !== '..' && !difference.startsWith(`..${process.platform === 'win32' ? '\\' : '/'}`) && !isAbsolute(difference);
  };
  if (!allowed(candidate)) return route.fulfill({ status: 403, body: 'Forbidden' });
  try {
    const canonical = await realpath(candidate);
    if (!allowed(canonical)) return route.fulfill({ status: 403, body: 'Forbidden' });
    await route.fulfill({ status: 200, contentType: mimeTypes[extname(canonical)] ?? 'application/octet-stream', body: await readFile(canonical) });
  } catch {
    await route.fulfill({ status: 404, body: 'Not found' });
  }
};

// 4. Measure visible content, including horizontally scrolling inner containers.
const inspectLayout = async (page) => page.evaluate(() => {
  const visible = (element) => {
    const header = element.closest('.stack-table thead');
    // 4.1 Mobile table headers remain accessible but are visually clipped.
    const headerStyle = header && getComputedStyle(header);
    const clippedHeader = header && headerStyle.clip !== 'auto' && headerStyle.position === 'absolute'
      && headerStyle.overflow === 'hidden' && header.clientWidth <= 1 && header.clientHeight <= 1;
    return element.getClientRects().length > 0 && !clippedHeader;
  };
  const describe = (element) => `${element.tagName.toLowerCase()}${element.className ? `.${String(element.className).trim().replace(/\s+/g, '.')}` : ''}`;
  const root = document.querySelector('#page-affairs');
  const overflows = [...root.querySelectorAll('div,section,main,aside,nav,table,tbody,tr,td,label,p,h1,h2,h3,strong,span')]
    .filter((element) => visible(element) && !element.matches('.work-nav,.acct-tabs')
      && element.clientWidth > 0 && element.scrollWidth - element.clientWidth > 1)
    .map((element) => ({ element: describe(element), client: element.clientWidth, scroll: element.scrollWidth }));
  const controls = [...root.querySelectorAll('button,input,select,textarea')]
    .filter((element) => visible(element) && !element.closest('.work-nav,.acct-tabs'))
    .flatMap((button) => {
      const rect = button.getBoundingClientRect();
      const container = button.closest('td,.work-card,.work-hero,.work-sidebar');
      const parent = container?.getBoundingClientRect();
      return rect.left < -1 || rect.right > innerWidth + 1 || (parent && (rect.left < parent.left - 1 || rect.right > parent.right + 1))
        ? [{ tag: button.tagName, text: button.id || button.textContent.trim(), left: rect.left, right: rect.right, parentLeft: parent?.left, parentRight: parent?.right }]
        : [];
    });
  const nav = root.querySelector('.work-nav');
  const navRect = nav.getBoundingClientRect();
  const tabs = [...nav.querySelectorAll('.work-nav-item')];
  return {
    documentOverflow: document.documentElement.scrollWidth - innerWidth,
    overflows, controls,
    selectedTabIndex: tabs.findIndex((tab) => tab.getAttribute('aria-current') === 'page'),
    navigation: {
      tabCount: tabs.length, hasDropdown: Boolean(root.querySelector('.building-module-select')),
      visible: visible(nav), backVisible: visible(root.querySelector('.building-mobile-back')),
      visibleTabs: tabs.filter(visible).length,
      iconCount: tabs.filter((tab) => visible(tab.querySelector('.work-nav-icon'))).length,
      left: navRect.left, right: navRect.right, scrollLeft: nav.scrollLeft,
      clientWidth: nav.clientWidth, scrollWidth: nav.scrollWidth, overflowX: getComputedStyle(nav).overflowX,
    },
    visiblePanels: [...root.querySelectorAll('[data-work-panel]')].filter(visible).map((element) => element.dataset.workPanel),
  };
});

// 5. Exercise every module and populated finance/account subtab.
await mkdir(output, { recursive: true });
const browser = await chromium.launch({ headless: true });
const report = {
  mode: staticMode ? 'compiled-static' : 'existing-server',
  cctvPlayback: 'Mock frame only; live playback not verified.',
  overflowScope: 'Mobile checks visible content containers; finance and account subtabs may scroll horizontally. Desktop preserves existing internal table scrolling.',
  headerExclusion: 'Only descendants of .stack-table thead with absolute position, hidden overflow, a CSS clip and at most 1px width/height are excluded as visually clipped accessibility headers.',
  skippedActions: ['Door opening and QR generation require POST and are not invoked.', 'Door camera is a text-only field in the current UI; no camera popup is available.'],
  results: [], errors: [], blockedExternalRequests: [],
};
try {
  const distRoot = staticMode ? await realpath(join(webRoot, 'dist')) : null;
  const scenarios = [
    { width: 320, height: 740, locale: 'zh-HK' },
    { width: 390, height: 844, locale: 'zh-HK' }, { width: 430, height: 932, locale: 'zh-HK' },
    { width: 768, height: 1024, locale: 'zh-HK' }, { width: 932, height: 430, locale: 'zh-HK' },
    { width: 1440, height: 1000, locale: 'zh-HK' },
    { width: 390, height: 844, locale: 'en' }, { width: 430, height: 932, locale: 'en' },
  ];
  for (const scenario of scenarios) {
    const context = await browser.newContext({ viewport: { width: scenario.width, height: scenario.height }, deviceScaleFactor: 1 });
    await context.addInitScript((locale) => {
      localStorage.setItem('ajoliving.access-token', 'mock-access');
      localStorage.setItem('ajoliving.refresh-token', 'mock-refresh');
      localStorage.setItem('ajoliving.locale', locale);
    }, scenario.locale);
    const page = await context.newPage();
    const blockedExternal = new Set();
    page.on('pageerror', (error) => report.errors.push({ scenario, type: 'pageerror', message: error.message }));
    page.on('console', (message) => {
      if (['warning', 'error'].includes(message.type()) && !blockedExternal.has(message.location().url)) {
        report.errors.push({ scenario, type: `console-${message.type()}`, message: message.text(), location: message.location() });
      }
    });
    page.on('requestfailed', (request) => {
      if (!blockedExternal.has(request.url())) report.errors.push({ scenario, type: 'requestfailed', url: request.url(), failure: request.failure() });
    });
    await context.route('**/*', async (route) => {
      const request = route.request();
      const url = new URL(request.url());
      if (url.pathname.startsWith('/api/v1/')) {
        if (request.method() !== 'GET') {
          report.errors.push({ scenario, type: 'unexpected-mutation', path: url.pathname, method: request.method() });
          return route.fulfill({ status: 405, json: { code: 'AUDIT_READ_ONLY', message: 'Mutations are disabled.' } });
        }
        const data = mockData(url.pathname);
        if (data === undefined) report.errors.push({ scenario, type: 'unknown-api', path: url.pathname });
        return route.fulfill({ status: 200, json: { code: 'OK', message: 'ok', data: data ?? { items: [] } } });
      }
      if (url.pathname === '/mock/camera') {
        return route.fulfill({ contentType: 'text/html', body: '<!doctype html><html><body style="margin:0;background:#151515;height:100svh;display:grid;place-items:center"><video style="width:100%;height:100%;object-fit:contain" controls></video></body></html>' });
      }
      if (url.origin === new URL(base).origin) {
        if (staticMode) return serveAsset(route, url.pathname, distRoot);
        return route.continue();
      }
      blockedExternal.add(request.url());
      report.blockedExternalRequests.push({ scenario, url: request.url() });
      return route.abort();
    });
    await page.goto(`${base}/building`, { waitUntil: 'networkidle' });
    if (scenario.width < 1024) {
      await page.locator('.work-nav-item').first().waitFor();
      const home = await inspectLayout(page);
      report.results.push({
        ...scenario, module: 'home', state: 'default',
        pass: home.documentOverflow <= 1 && home.navigation.visible && home.navigation.visibleTabs === modules.length
          && home.navigation.iconCount === modules.length && home.visiblePanels.length === 0
          && home.navigation.scrollWidth <= home.navigation.clientWidth + 1,
        ...home,
      });
      if (scenario.locale === 'zh-HK' && scenario.width === 390) {
        await page.screenshot({ path: join(output, 'mobile-building-home-390-after.png') });
      }
    } else {
      await page.locator('.notice-table td[data-label]').first().waitFor();
    }
    for (const [module, populatedSelector] of modules) {
      const index = modules.findIndex(([name]) => name === module);
      if (scenario.width < 1024 && index > 0) {
        await page.locator('.building-mobile-back').click();
        await page.locator('.work-nav-item').first().waitFor();
      }
      await page.locator('.work-nav-item').nth(index).click();
      const panel = page.locator(`[data-work-panel="${module}"]`);
      await panel.locator(populatedSelector).first().waitFor();
      await page.waitForTimeout(200);
      if (scenario.width < 1024 && index === 0) {
        await page.goBack({ waitUntil: 'domcontentloaded' });
        await page.locator('.work-nav-item').first().waitFor();
        const returnedHome = await inspectLayout(page);
        report.results.push({
          ...scenario, module: 'home', state: 'history-back',
          pass: returnedHome.navigation.visible && returnedHome.visiblePanels.length === 0,
          ...returnedHome,
        });
        await page.goForward({ waitUntil: 'domcontentloaded' });
        await panel.locator(populatedSelector).first().waitFor();
      }
      const capture = async (state) => {
        const layout = await inspectLayout(page);
        const navigationFits = layout.navigation.tabCount === modules.length && !layout.navigation.hasDropdown
          && (scenario.width < 1024
            ? !layout.navigation.visible && layout.navigation.backVisible
            : layout.navigation.left >= -1 && layout.navigation.right <= scenario.width + 1);
        const pass = layout.documentOverflow <= 1 && (scenario.width >= 1024 || layout.overflows.length === 0) && layout.controls.length === 0
          && navigationFits && layout.selectedTabIndex === index && layout.visiblePanels.length === 1 && layout.visiblePanels[0] === module;
        report.results.push({ ...scenario, module, state, pass, ...layout });
      };
      await capture('default');
      if (module === 'affairs-finance') {
        for (const index of [1, 2]) {
          await panel.locator('.acct-tab').nth(index).click();
          await panel.locator('.acct-subpanel.on .building-file-table td[data-label]').first().waitFor();
          await capture(index === 1 ? 'financial-reports' : 'audit-reports');
        }
      }
      if (module === 'affairs-owner-account') {
        await panel.locator('.acct-tab').nth(1).click();
        await panel.locator('.owner-record-table td[data-label]').first().waitFor();
        await capture('payment-history');
        await panel.locator('input[type="date"]').nth(0).fill('2026-10-01');
        await panel.locator('input[type="date"]').nth(1).fill('2026-10-29');
        await panel.locator('.acct-filter-row button[type="submit"]').click();
        await panel.locator('.owner-record-table td[data-label]').first().waitFor();
        await capture('payment-date-filter');
      }
      if (module === 'affairs-feedback') {
        await panel.locator('tbody .work-mini-btn').first().click();
        await panel.locator('.affairs-record-detail:visible .affairs-record-message').waitFor();
        await capture('service-case-detail');
        for (const [index, mode] of ['repair', 'feedback'].entries()) {
          await panel.locator('.affairs-entry-card').nth(index).click();
          const category = panel.locator(`#affairs-${mode}-category`);
          const categoryValue = await category.locator('option').nth(1).getAttribute('value');
          await category.selectOption(categoryValue);
          await panel.locator('.affairs-guided-actions button').nth(1).click();
          await panel.locator(`#affairs-${mode}-content`).fill(longTitle.repeat(3));
          await panel.locator('#affairs-location').fill('地下大堂主要出入口 / Ground Floor Main Entrance');
          await capture(`${mode}-content-form`);
          await panel.locator('.affairs-guided-actions button').nth(1).click();
          await panel.locator('.affairs-review-list').waitFor();
          await capture(`${mode}-review`);
        }
      }
      if (module === 'affairs-access') {
        await panel.locator('.access-table-actions button').nth(1).click();
        await capture('visible-password');
      }
      if (module === 'affairs-icctv') {
        await panel.locator('.icctv-row-actions .primary').first().click();
        const iframe = panel.locator('.icctv-expanded-row iframe');
        await iframe.waitFor();
        const geometry = await iframe.evaluate((frame) => {
          const frameRect = frame.getBoundingClientRect();
          const tableRect = frame.closest('table').getBoundingClientRect();
          const viewer = frame.parentElement.getBoundingClientRect();
          return { width: frameRect.width, height: frameRect.height, viewerWidth: viewer.width, tableWidth: tableRect.width };
        });
        const mobileGeometry = scenario.width >= 1024 || (
          geometry.width >= geometry.tableWidth * 0.85
          && Math.abs(geometry.width / geometry.height - 16 / 9) < 0.03
        );
        report.results.push({ ...scenario, module, state: 'expanded-geometry', pass: geometry.width > 100 && Math.abs(geometry.width - geometry.viewerWidth) < 2 && mobileGeometry, geometry });
        await capture('expanded');
      }
      const mainScreenshot = ['affairs-notices', 'affairs-icctv'].includes(module);
      if (scenario.locale === 'zh-HK' && (scenario.width === 390 || (scenario.width === 430 && mainScreenshot))) {
        await panel.evaluate((element, moduleName) => {
          const top = moduleName === 'affairs-notices' ? 0 : element.getBoundingClientRect().top + scrollY - 70;
          window.scrollTo({ top, behavior: 'instant' });
        }, module);
        await page.waitForTimeout(200);
        const name = module === 'affairs-icctv' ? 'cctv' : module.replace('affairs-', '');
        await page.screenshot({ path: join(output, `mobile-building-${name}-${scenario.width}-after.png`) });
        if (scenario.width === 430) await page.screenshot({ path: join(output, `mobile-building-${name}-after.png`) });
      }
    }
    await context.close();
  }
} catch (error) {
  report.errors.push({ type: 'audit-error', message: error.stack ?? error.message });
} finally {
  await browser.close();
}
report.passed = report.errors.length === 0 && report.results.length > 0 && report.results.every((result) => result.pass);
await writeFile(join(output, 'building-mobile-audit.json'), `${JSON.stringify(report, null, 2)}\n`);
console.log(JSON.stringify({ passed: report.passed, checks: report.results.length, failures: report.results.filter((result) => !result.pass), errors: report.errors, output }, null, 2));
if (!report.passed) process.exitCode = 1;

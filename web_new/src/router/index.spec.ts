/*
 * 路由多語系測試。
 * 1. 驗證所有路由文件標題 key 同時存在於 zh-HK 與 en。
 * 2. 驗證文件標題會使用目前語系即時更新。
 * 3. 驗證 Staff 守衛只用於管理中心與管理設定頁。
 */
import { afterEach, describe, expect, it } from 'vitest';

import i18n, { applyLocale } from '@/i18n';

import router, { resolveAgencyAccountRedirect, updateDocumentTitle } from './index';

describe('router locale titles', () => {
  afterEach(() => {
    applyLocale('zh-HK');
  });

  it('defines every route title in both supported locales', () => {
    const titleKeys = [...new Set(router.getRoutes()
      .map((route) => route.meta.titleKey)
      .filter((titleKey): titleKey is string => typeof titleKey === 'string'))];

    titleKeys.forEach((titleKey) => {
      expect(i18n.global.te(titleKey, 'zh-HK'), `${titleKey} in zh-HK`).toBe(true);
      expect(i18n.global.te(titleKey, 'en'), `${titleKey} in en`).toBe(true);
    });
  });

  it('updates the current document title after a locale change', () => {
    applyLocale('en');
    updateDocumentTitle('nav.home');
    expect(document.title).toBe('Home | AJO Living');

    applyLocale('zh-HK');
    updateDocumentTitle('nav.home');
    expect(document.title).toBe('首頁 | AJO Living');
  });
});

describe('agency account access', () => {
  it('keeps unapproved agents out of protected member operations', () => {
    expect(resolveAgencyAccountRedirect('individual_agent', 'pending_review', '/account/properties/sale', true))
      .toBe('/account/profile/agency-profile');
    expect(resolveAgencyAccountRedirect('agency_company', 'rejected', '/login', true))
      .toBe('/account/profile/agency-profile');
  });

  it('allows unapproved agents to browse public pages', () => {
    expect(resolveAgencyAccountRedirect('individual_agent', 'pending_review', '/', false)).toBeNull();
    expect(resolveAgencyAccountRedirect('agency_company', 'rejected', '/properties', false)).toBeNull();
  });

  it('keeps personal and company subaccounts out of the editable agency profile page', () => {
    expect(resolveAgencyAccountRedirect('personal', 'active', '/account/profile/agency-profile', true))
      .toBe('/account/profile');
    expect(resolveAgencyAccountRedirect('agency_company_subaccount', 'active', '/account/profile/agency-profile', true))
      .toBe('/account/profile');
  });
});

describe('staff route boundary', () => {
  it('limits staff guards to management surfaces', () => {
    const guardedRoutes = router.getRoutes().filter((route) => route.meta.requiresStaff);

    expect(guardedRoutes.length).toBeGreaterThan(0);
    guardedRoutes.forEach((route) => {
      const isManagementRoute = route.path.startsWith('/account/marketplace/management')
        || route.path.startsWith('/account/marketplace/settings');
      expect(isManagementRoute, `${route.path} must not require Staff`).toBe(true);
    });
  });
});

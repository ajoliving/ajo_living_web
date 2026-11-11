/*
 * 主題 Token 測試。
 * 1. 驗證主題方案與可選主題。
 * 2. 驗證主題色會同步到既有語意變量與維持白色內容面。
 */
import { afterEach, describe, expect, it } from 'vitest';
import { compileStyle, parse } from '@vue/compiler-sfc';

import { DEFAULT_THEME_NAME, THEME_OPTIONS, applyTheme } from './theme';
import accountPage from '@/pages/account/my/AccountMyPage.vue?raw';
import offersPage from '@/pages/offers/Page.vue?raw';
import offersDetailPage from '@/pages/offers/detail/Page.vue?raw';

describe('theme tokens', () => {
  afterEach(() => {
    applyTheme(DEFAULT_THEME_NAME);
  });

  it('provides the default, cyan, and night choices', () => {
    expect(THEME_OPTIONS.filter((theme) => theme.group === 'orange')).toHaveLength(40);
    expect(THEME_OPTIONS.map((theme) => theme.value)).toContain('cyan');
    expect(THEME_OPTIONS.map((theme) => theme.value)).toContain('night');
  });

  it('uses a white canvas with the selected primary action color', () => {
    applyTheme('orange-05');

    expect(document.documentElement.style.getPropertyValue('--color-canvas')).toBe('255 255 255');
    expect(document.documentElement.style.getPropertyValue('--color-primary')).toBe('255 90 0');
    expect(document.documentElement.style.getPropertyValue('--color-primary-contrast')).toBe('255 255 255');
    expect(document.documentElement.style.getPropertyValue('--color-surface-muted')).toBe('247 248 250');
  });

  it('synchronizes a non-orange theme with legacy color variables', () => {
    applyTheme('emerald');

    expect(document.documentElement.getAttribute('data-theme')).toBe('emerald');
    expect(document.documentElement.style.getPropertyValue('--brand')).toBe('rgb(23 130 74)');
    expect(document.documentElement.style.getPropertyValue('--ink')).toBe('rgb(26 26 26)');
    expect(document.documentElement.style.getPropertyValue('--sur')).toBe('rgb(255 255 255)');
    expect(document.documentElement.style.getPropertyValue('--color-primary-contrast')).toBe('255 255 255');
  });

  it('preserves dark foreground for lighter legacy themes', () => {
    applyTheme('orange-14');
    expect(document.documentElement.style.getPropertyValue('--color-primary-contrast')).toBe('26 26 26');
  });
  it('applies dark surfaces for night mode', () => {
    applyTheme('night');

    expect(document.documentElement.getAttribute('data-theme')).toBe('night');
    expect(document.documentElement.style.colorScheme).toBe('dark');
    expect(document.documentElement.style.getPropertyValue('--sur')).toBe('rgb(28 32 38)');
  });

  it('uses theme surfaces for account cards and night-mode offer cards', () => {
    expect(accountPage).not.toMatch(/background:\s*#fff(?:fff)?\s*;/i);
    expect(accountPage).toMatch(/\.work-card\s*\{[^}]*background:\s*var\(--sur\)/s);
    const style = parse(offersPage).descriptor.styles[0];
    const compiled = compileStyle({ source: style.content, filename: 'Page.vue', id: 'data-v-test', scoped: true }).code;
    expect(compiled).toContain('html[data-theme="night"] #page-offers .gp-card');
    expect(compiled).toContain('html[data-theme="night"] #page-offers .gp-card-name-text');
    expect(compiled).not.toMatch(/html\[data-theme="night"\]\s*\{/);
  });

  it('scopes the detail chart night colors to the detail page', () => {
    const style = parse(offersDetailPage).descriptor.styles[0];
    const compiled = compileStyle({ source: style.content, filename: 'DetailPage.vue', id: 'data-v-test', scoped: true }).code;
    expect(compiled).toContain('html[data-theme="night"] .offers-detail-page .gp-chart');
    expect(compiled).toContain('--gp-series-3: #e1e5ea');
    expect(compiled).toMatch(/\.gp-chart text[^{]*\{[^}]*fill:\s*var\(--ink-3\)/s);
    expect(compiled).not.toMatch(/html\[data-theme="night"\]\s*\{/);
  });
});

/*
 * 主題 Token 測試。
 * 1. 驗證六十組主題方案與四組十色亮橙細分。
 * 2. 驗證主題色會同步到既有語意變量與維持白色內容面。
 */
import { afterEach, describe, expect, it } from 'vitest';

import { DEFAULT_THEME_NAME, THEME_OPTIONS, applyTheme } from './theme';

describe('theme tokens', () => {
  afterEach(() => {
    applyTheme(DEFAULT_THEME_NAME);
  });

  it('provides forty bright orange themes and twenty other formal themes', () => {
    expect(THEME_OPTIONS).toHaveLength(60);
    expect(THEME_OPTIONS.filter((theme) => theme.group === 'orange')).toHaveLength(40);
    expect(THEME_OPTIONS.filter((theme) => theme.group === 'other')).toHaveLength(20);
    expect(THEME_OPTIONS.filter((theme) => theme.orangeTone === 'flame')).toHaveLength(10);
    expect(THEME_OPTIONS.filter((theme) => theme.orangeTone === 'classic')).toHaveLength(10);
    expect(THEME_OPTIONS.filter((theme) => theme.orangeTone === 'solar')).toHaveLength(10);
    expect(THEME_OPTIONS.filter((theme) => theme.orangeTone === 'peach')).toHaveLength(10);
  });

  it('uses a white canvas with the selected primary action color', () => {
    applyTheme('orange-05');

    expect(document.documentElement.style.getPropertyValue('--color-canvas')).toBe('255 255 255');
    expect(document.documentElement.style.getPropertyValue('--color-primary')).toBe('255 90 0');
    expect(document.documentElement.style.getPropertyValue('--color-surface-muted')).toBe('247 248 250');
  });

  it('synchronizes a non-orange theme with legacy color variables', () => {
    applyTheme('emerald');

    expect(document.documentElement.getAttribute('data-theme')).toBe('emerald');
    expect(document.documentElement.style.getPropertyValue('--brand')).toBe('rgb(23 130 74)');
    expect(document.documentElement.style.getPropertyValue('--ink')).toBe('rgb(26 26 26)');
    expect(document.documentElement.style.getPropertyValue('--sur')).toBe('rgb(255 255 255)');
  });
});

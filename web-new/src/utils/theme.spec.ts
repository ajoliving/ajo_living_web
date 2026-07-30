/*
 * 主題 Token 測試。
 * 1. 驗證頁面底色固定為白色。
 * 2. 驗證鮮橙主操作色會同步到既有語意變量。
 */
import { afterEach, describe, expect, it } from 'vitest';

import { applyTheme } from './theme';

describe('theme tokens', () => {
  afterEach(() => {
    applyTheme();
  });

  it('uses a white canvas with a bright orange primary action color', () => {
    applyTheme();

    expect(document.documentElement.style.getPropertyValue('--color-canvas')).toBe('255 255 255');
    expect(document.documentElement.style.getPropertyValue('--color-primary')).toBe('255 106 0');
  });

  it('synchronizes legacy color variables with the formal color tokens', () => {
    applyTheme();

    expect(document.documentElement.style.getPropertyValue('--brand')).toBe('rgb(255 106 0)');
    expect(document.documentElement.style.getPropertyValue('--ink')).toBe('rgb(26 26 26)');
    expect(document.documentElement.style.getPropertyValue('--sur')).toBe('rgb(255 255 255)');
  });
});

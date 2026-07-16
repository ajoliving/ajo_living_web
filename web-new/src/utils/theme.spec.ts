/*
 * 主題 Token 測試。
 * 1. 驗證深色主題的主色淺底保持可辨識對比。
 * 2. 驗證新舊語意色彩變量會由同一主題配置同步。
 */
import { afterEach, describe, expect, it } from 'vitest';

import { applyTheme } from './theme';

describe('theme tokens', () => {
  afterEach(() => {
    applyTheme('default');
  });

  it('keeps the dark primary-soft surface distinct from the primary text color', () => {
    applyTheme('dark-neutral');

    expect(document.documentElement.style.getPropertyValue('--color-primary-soft')).toBe('72 33 10');
    expect(document.documentElement.style.getPropertyValue('--color-primary')).toBe('240 90 0');
  });

  it('synchronizes legacy color variables with the active semantic tokens', () => {
    applyTheme('dark-neutral');

    expect(document.documentElement.style.getPropertyValue('--brand')).toBe('rgb(240 90 0)');
    expect(document.documentElement.style.getPropertyValue('--ink')).toBe('rgb(244 246 251)');
    expect(document.documentElement.style.getPropertyValue('--sur')).toBe('rgb(17 24 39)');
  });
});

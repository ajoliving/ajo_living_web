/*
 * 偏好設定語系與主題測試。
 * 1. 驗證語系切換會保存到 localStorage。
 * 2. 驗證啟動還原只接受支援的語系值。
 * 3. 驗證主題選擇器關閉時固定使用預設品牌色。
 */
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';

import { usePreferenceStore } from './preferences';

describe('preference locale persistence', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    window.localStorage.clear();
  });

  afterEach(() => {
    vi.unstubAllEnvs();
  });

  it('persists a selected locale and hydrates it in a fresh store', () => {
    const preferenceStore = usePreferenceStore();

    preferenceStore.setLocale('en');
    expect(window.localStorage.getItem('ajoliving.locale')).toBe('en');

    setActivePinia(createPinia());
    const restoredStore = usePreferenceStore();
    restoredStore.hydratePreferences();

    expect(restoredStore.locale).toBe('en');
  });

  it('falls back to zh-HK when localStorage contains an unsupported locale', () => {
    window.localStorage.setItem('ajoliving.locale', 'fr');

    const preferenceStore = usePreferenceStore();
    preferenceStore.hydratePreferences();

    expect(preferenceStore.locale).toBe('zh-HK');
  });

  it('locks the default brand theme when the theme picker is disabled', () => {
    vi.stubEnv('VITE_ENABLE_THEME_PICKER', '0');
    window.localStorage.setItem('ajoliving.theme', 'graphite');

    const preferenceStore = usePreferenceStore();
    preferenceStore.hydratePreferences();

    expect(preferenceStore.theme).toBe('orange-05');
    expect(document.documentElement.style.getPropertyValue('--color-primary')).toBe('255 90 0');
  });

  it('restores only the default, cyan, or night theme', () => {
    vi.stubEnv('VITE_ENABLE_THEME_PICKER', '1');
    window.localStorage.setItem('ajoliving.theme', 'cyan');

    const preferenceStore = usePreferenceStore();
    preferenceStore.hydratePreferences();

    expect(preferenceStore.theme).toBe('cyan');
  });

  it('falls back to the default theme when a retired palette choice is stored', () => {
    vi.stubEnv('VITE_ENABLE_THEME_PICKER', '1');
    window.localStorage.setItem('ajoliving.theme', 'graphite');

    const preferenceStore = usePreferenceStore();
    preferenceStore.hydratePreferences();

    expect(preferenceStore.theme).toBe('orange-05');
  });
});

/*
 * 偏好設定語系測試。
 * 1. 驗證語系切換會保存到 localStorage。
 * 2. 驗證啟動還原只接受支援的語系值。
 */
import { beforeEach, describe, expect, it } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';

import { usePreferenceStore } from './preferences';

describe('preference locale persistence', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    window.localStorage.clear();
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
});

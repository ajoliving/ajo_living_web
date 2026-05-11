/*
 * 偏好設定狀態。
 * 1. 管理主題與語系切換。
 * 2. 將設定持久化到瀏覽器本地儲存。
 */
import { defineStore } from 'pinia';

import type { AppThemeName } from '@/utils/theme';
import { DEFAULT_THEME_NAME, applyTheme } from '@/utils/theme';

// 1. 定義語系型別
export type AppLocale = 'zh-HK' | 'en';

// 2. 定義可選語系清單
export const LOCALE_OPTIONS: Array<{ labelKey: string; value: AppLocale }> = [
  { labelKey: 'common.locale.zhHk', value: 'zh-HK' },
  { labelKey: 'common.locale.en', value: 'en' },
];

const THEME_STORAGE_KEY = 'ajoliving.theme';
const THEME_STORAGE_VERSION_KEY = 'ajoliving.theme.version';
const CURRENT_THEME_STORAGE_VERSION = '4';
const LOCALE_STORAGE_KEY = 'ajoliving.locale';

const isValidTheme = (value: string | null): value is AppThemeName =>
  value === 'default' ||
  value === 'copper-sun' ||
  value === 'dark-neutral';

const isValidLocale = (value: string | null): value is AppLocale =>
  value === 'zh-HK' || value === 'en';

// 3. 遷移舊版主題儲存值，收斂為三套主題
const migrateStoredTheme = (storedTheme: string | null, storageVersion: string | null): AppThemeName | null => {
  if (storageVersion !== CURRENT_THEME_STORAGE_VERSION) {
    const migratedTheme =
      storedTheme === 'copper-sun' || storedTheme === 'default'
        ? 'copper-sun'
        : storedTheme === 'harbour-blue' ||
            storedTheme === 'warm' ||
            storedTheme === 'amber-glow' ||
            storedTheme === 'citrus-mist' ||
            storedTheme === 'daylight'
          ? 'default'
          : storedTheme === 'dark-neutral'
            ? 'dark-neutral'
            : null;

    if (migratedTheme) {
      window.localStorage.setItem(THEME_STORAGE_KEY, migratedTheme);
      window.localStorage.setItem(THEME_STORAGE_VERSION_KEY, CURRENT_THEME_STORAGE_VERSION);
      return migratedTheme;
    }
  }

  if (isValidTheme(storedTheme)) {
    window.localStorage.setItem(THEME_STORAGE_VERSION_KEY, CURRENT_THEME_STORAGE_VERSION);
    return storedTheme;
  }

  return null;
};

// 4. 建立偏好設定 Store
export const usePreferenceStore = defineStore('preferences', {
  state: () => ({
    theme: DEFAULT_THEME_NAME as AppThemeName,
    locale: 'zh-HK' as AppLocale,
    mobile_menu_open: false,
  }),
  actions: {
    // 5. 還原已保存的語系與主題偏好
    hydratePreferences() {
      if (typeof window === 'undefined') {
        return;
      }

      const storedTheme = window.localStorage.getItem(THEME_STORAGE_KEY);
      const themeStorageVersion = window.localStorage.getItem(THEME_STORAGE_VERSION_KEY);
      const storedLocale = window.localStorage.getItem(LOCALE_STORAGE_KEY);
      const migratedTheme = migrateStoredTheme(storedTheme, themeStorageVersion);

      if (migratedTheme) {
        this.theme = migratedTheme;
      }

      if (isValidLocale(storedLocale)) {
        this.locale = storedLocale;
      }

      applyTheme(this.theme);
    },

    // 6. 更新主題並寫入本地儲存
    setTheme(theme: AppThemeName) {
      this.theme = theme;
      applyTheme(theme);
      window.localStorage.setItem(THEME_STORAGE_KEY, theme);
      window.localStorage.setItem(THEME_STORAGE_VERSION_KEY, CURRENT_THEME_STORAGE_VERSION);
    },

    // 7. 更新語系並寫入本地儲存
    setLocale(locale: AppLocale) {
      this.locale = locale;
      window.localStorage.setItem(LOCALE_STORAGE_KEY, locale);
    },

    // 8. 更新 mobile 導航開關狀態
    setMobileMenuOpen(open: boolean) {
      this.mobile_menu_open = open;
    },
  },
});

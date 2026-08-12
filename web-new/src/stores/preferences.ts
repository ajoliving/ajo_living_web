/*
 * 偏好設定狀態。
 * 1. 管理可持久化的全站主題色與語系。
 * 2. 管理 mobile 導航偏好。
 */
import { defineStore } from 'pinia';

import { DEFAULT_THEME_NAME, applyTheme, getThemeOption, type AppThemeName } from '@/utils/theme';

// 1. 定義語系型別
export type AppLocale = 'zh-HK' | 'en';

// 2. 定義可選語系清單
export const LOCALE_OPTIONS: Array<{ labelKey: string; value: AppLocale }> = [
  { labelKey: 'common.locale.zhHk', value: 'zh-HK' },
  { labelKey: 'common.locale.en', value: 'en' },
];

const THEME_STORAGE_KEY = 'ajoliving.theme';
const THEME_STORAGE_VERSION_KEY = 'ajoliving.theme.version';
const CURRENT_THEME_STORAGE_VERSION = '8';
const LOCALE_STORAGE_KEY = 'ajoliving.locale';

const isValidLocale = (value: string | null): value is AppLocale =>
  value === 'zh-HK' || value === 'en';

// 3. 讀取有效主題，舊版名稱與未知值安全回落至預設方案
const getStoredTheme = (): AppThemeName => {
  const storedTheme = window.localStorage.getItem(THEME_STORAGE_KEY);
  return getThemeOption(storedTheme).value;
};

// 4. 建立偏好設定 Store
export const usePreferenceStore = defineStore('preferences', {
  state: () => ({
    locale: 'zh-HK' as AppLocale,
    theme: DEFAULT_THEME_NAME as AppThemeName,
    mobile_menu_open: false,
  }),
  actions: {
    // 4. 還原已保存的語系與主題
    hydratePreferences() {
      if (typeof window === 'undefined') {
        return;
      }

      const storedLocale = window.localStorage.getItem(LOCALE_STORAGE_KEY);
      this.theme = getStoredTheme();
      window.localStorage.setItem(THEME_STORAGE_KEY, this.theme);
      window.localStorage.setItem(THEME_STORAGE_VERSION_KEY, CURRENT_THEME_STORAGE_VERSION);

      if (isValidLocale(storedLocale)) {
        this.locale = storedLocale;
      }

      applyTheme(this.theme);
    },

    // 5. 更新語系並寫入本地儲存
    setLocale(locale: AppLocale) {
      this.locale = locale;
      window.localStorage.setItem(LOCALE_STORAGE_KEY, locale);
    },

    // 6. 更新主題並即時同步全站 token
    setTheme(theme: AppThemeName) {
      this.theme = getThemeOption(theme).value;
      window.localStorage.setItem(THEME_STORAGE_KEY, this.theme);
      window.localStorage.setItem(THEME_STORAGE_VERSION_KEY, CURRENT_THEME_STORAGE_VERSION);
      applyTheme(this.theme);
    },

    // 7. 更新 mobile 導航開關狀態
    setMobileMenuOpen(open: boolean) {
      this.mobile_menu_open = open;
    },
  },
});

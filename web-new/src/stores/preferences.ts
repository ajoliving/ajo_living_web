/*
 * 偏好設定狀態。
 * 1. 固定套用白色與鮮橙正式視覺基線。
 * 2. 管理語系與 mobile 導航偏好。
 */
import { defineStore } from 'pinia';

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
const CURRENT_THEME_STORAGE_VERSION = '6';
const LOCALE_STORAGE_KEY = 'ajoliving.locale';

const isValidLocale = (value: string | null): value is AppLocale =>
  value === 'zh-HK' || value === 'en';

// 3. 將舊版主題偏好收斂至白色與鮮橙正式皮膚
const migrateStoredTheme = (): void => {
  window.localStorage.setItem(THEME_STORAGE_KEY, DEFAULT_THEME_NAME);
  window.localStorage.setItem(THEME_STORAGE_VERSION_KEY, CURRENT_THEME_STORAGE_VERSION);
};

// 4. 建立偏好設定 Store
export const usePreferenceStore = defineStore('preferences', {
  state: () => ({
    locale: 'zh-HK' as AppLocale,
    mobile_menu_open: false,
  }),
  actions: {
    // 5. 還原已保存的語系與套用正式視覺基線
    hydratePreferences() {
      if (typeof window === 'undefined') {
        return;
      }

      const storedLocale = window.localStorage.getItem(LOCALE_STORAGE_KEY);
      migrateStoredTheme();

      if (isValidLocale(storedLocale)) {
        this.locale = storedLocale;
      }

      applyTheme();
    },

    // 6. 更新語系並寫入本地儲存
    setLocale(locale: AppLocale) {
      this.locale = locale;
      window.localStorage.setItem(LOCALE_STORAGE_KEY, locale);
    },

    // 7. 更新 mobile 導航開關狀態
    setMobileMenuOpen(open: boolean) {
      this.mobile_menu_open = open;
    },
  },
});

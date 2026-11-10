/*
 * 會員中心模組型別。
 * 1. 定義會員中心導覽、提醒與隱私設定結構。
 * 2. 為帳戶頁面與元件提供穩定型別。
 */
import type { AppLocale } from '@/stores/preferences';
import type { AppThemeName } from '@/utils/theme';

export type AccountPrivacyScope = 'public' | 'community' | 'private';

// 1. 會員中心側欄項目
export interface AccountNavItem {
  key: string;
  label: string;
  to: string;
  exact?: boolean;
}

// 2. 偏好設定表單
export interface AccountPreferenceForm {
  locale: AppLocale;
  theme: AppThemeName;
  email_updates: boolean;
  chat_alerts: boolean;
  listing_reminders: boolean;
}

// 3. 通知偏好設定
export interface AccountNotificationPreference {
  key: keyof Pick<
    AccountPreferenceForm,
    'email_updates' | 'chat_alerts' | 'listing_reminders'
  >;
  title: string;
  description: string;
}

// 4. 隱私設定項目
export interface AccountPrivacyPreference {
  key: string;
  title: string;
  description: string;
  value: AccountPrivacyScope;
}

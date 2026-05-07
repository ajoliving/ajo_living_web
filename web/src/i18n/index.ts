/*
 * 多語系初始化。
 * 1. 組裝 zh-HK 與 en 語系內容。
 * 2. 提供全站共用的 vue-i18n 實例。
 */
import { createI18n } from 'vue-i18n';

import accountEn from '@/i18n/locales/en/account';
import authEn from '@/i18n/locales/en/auth';
import chatEn from '@/i18n/locales/en/chat';
import commonEn from '@/i18n/locales/en/common';
import homeEn from '@/i18n/locales/en/home';
import marketplaceEn from '@/i18n/locales/en/marketplace';
import navEn from '@/i18n/locales/en/nav';
import propertyEn from '@/i18n/locales/en/property';
import accountZhHk from '@/i18n/locales/zh-HK/account';
import authZhHk from '@/i18n/locales/zh-HK/auth';
import chatZhHk from '@/i18n/locales/zh-HK/chat';
import commonZhHk from '@/i18n/locales/zh-HK/common';
import homeZhHk from '@/i18n/locales/zh-HK/home';
import marketplaceZhHk from '@/i18n/locales/zh-HK/marketplace';
import navZhHk from '@/i18n/locales/zh-HK/nav';
import propertyZhHk from '@/i18n/locales/zh-HK/property';

const messages = {
  'zh-HK': {
    account: accountZhHk,
    auth: authZhHk,
    chat: chatZhHk,
    common: commonZhHk,
    home: homeZhHk,
    marketplace: marketplaceZhHk,
    nav: navZhHk,
    property: propertyZhHk,
  },
  en: {
    account: accountEn,
    auth: authEn,
    chat: chatEn,
    common: commonEn,
    home: homeEn,
    marketplace: marketplaceEn,
    nav: navEn,
    property: propertyEn,
  },
};

// 1. 建立 i18n 實例
const i18n = createI18n({
  legacy: false,
  locale: 'zh-HK',
  fallbackLocale: 'en',
  messages: messages as Record<string, any>,
});

export default i18n;

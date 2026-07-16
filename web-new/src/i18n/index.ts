/*
 * 多語系初始化。
 * 1. 組裝 zh-HK 與 en 語系內容。
 * 2. 提供全站共用的 vue-i18n 實例。
 */
import { createI18n } from 'vue-i18n';

import type { AppLocale } from '@/stores/preferences';
import accountEn from '@/i18n/locales/en/account';
import authEn from '@/i18n/locales/en/auth';
import buildingEn from '@/i18n/locales/en/building';
import chatEn from '@/i18n/locales/en/chat';
import channelsEn from '@/i18n/locales/en/channels';
import commonEn from '@/i18n/locales/en/common';
import homeEn from '@/i18n/locales/en/home';
import marketplaceEn from '@/i18n/locales/en/marketplace';
import navEn from '@/i18n/locales/en/nav';
import offersEn from '@/i18n/locales/en/offers';
import propertyEn from '@/i18n/locales/en/property';
import servicedResidenceEn from '@/i18n/locales/en/servicedResidence';
import trendEn from '@/i18n/locales/en/trend';
import accountZhHk from '@/i18n/locales/zh-HK/account';
import authZhHk from '@/i18n/locales/zh-HK/auth';
import buildingZhHk from '@/i18n/locales/zh-HK/building';
import chatZhHk from '@/i18n/locales/zh-HK/chat';
import channelsZhHk from '@/i18n/locales/zh-HK/channels';
import commonZhHk from '@/i18n/locales/zh-HK/common';
import homeZhHk from '@/i18n/locales/zh-HK/home';
import marketplaceZhHk from '@/i18n/locales/zh-HK/marketplace';
import navZhHk from '@/i18n/locales/zh-HK/nav';
import offersZhHk from '@/i18n/locales/zh-HK/offers';
import propertyZhHk from '@/i18n/locales/zh-HK/property';
import servicedResidenceZhHk from '@/i18n/locales/zh-HK/servicedResidence';
import trendZhHk from '@/i18n/locales/zh-HK/trend';

const messages = {
  'zh-HK': {
    account: accountZhHk,
    payments: accountZhHk.payments,
    auth: authZhHk,
    building: buildingZhHk,
    chat: chatZhHk,
    channels: channelsZhHk,
    common: commonZhHk,
    home: homeZhHk,
    marketplace: marketplaceZhHk,
    nav: navZhHk,
    offers: offersZhHk,
    property: propertyZhHk,
    servicedResidence: servicedResidenceZhHk,
    trend: trendZhHk,
  },
  en: {
    account: accountEn,
    payments: accountEn.payments,
    auth: authEn,
    building: buildingEn,
    chat: chatEn,
    channels: channelsEn,
    common: commonEn,
    home: homeEn,
    marketplace: marketplaceEn,
    nav: navEn,
    offers: offersEn,
    property: propertyEn,
    servicedResidence: servicedResidenceEn,
    trend: trendEn,
  },
};

// 1. 建立 i18n 實例
const i18n = createI18n({
  legacy: false,
  locale: 'zh-HK',
  fallbackLocale: 'en',
  messages,
});

// 2. 同步語系與瀏覽器文件語言，供入口與根元件共用
export const applyLocale = (locale: AppLocale): void => {
  i18n.global.locale.value = locale;

  if (typeof document !== 'undefined') {
    document.documentElement.lang = locale;
  }
};

export default i18n;

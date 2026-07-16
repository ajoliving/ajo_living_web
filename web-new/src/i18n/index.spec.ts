/*
 * 多語系底座測試。
 * 1. 驗證 English 與繁體中文語系的訊息鍵完整對齊。
 * 2. 驗證 fallback 設定與語系同步會更新文件語言。
 */
import { afterEach, describe, expect, it } from 'vitest';

import accountEn from '@/i18n/locales/en/account';
import authEn from '@/i18n/locales/en/auth';
import buildingEn from '@/i18n/locales/en/building';
import channelsEn from '@/i18n/locales/en/channels';
import chatEn from '@/i18n/locales/en/chat';
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
import channelsZhHk from '@/i18n/locales/zh-HK/channels';
import chatZhHk from '@/i18n/locales/zh-HK/chat';
import commonZhHk from '@/i18n/locales/zh-HK/common';
import homeZhHk from '@/i18n/locales/zh-HK/home';
import marketplaceZhHk from '@/i18n/locales/zh-HK/marketplace';
import navZhHk from '@/i18n/locales/zh-HK/nav';
import offersZhHk from '@/i18n/locales/zh-HK/offers';
import propertyZhHk from '@/i18n/locales/zh-HK/property';
import servicedResidenceZhHk from '@/i18n/locales/zh-HK/servicedResidence';
import trendZhHk from '@/i18n/locales/zh-HK/trend';

import i18n, { applyLocale } from './index';

const localeModules = {
  account: { en: accountEn, 'zh-HK': accountZhHk },
  auth: { en: authEn, 'zh-HK': authZhHk },
  building: { en: buildingEn, 'zh-HK': buildingZhHk },
  channels: { en: channelsEn, 'zh-HK': channelsZhHk },
  chat: { en: chatEn, 'zh-HK': chatZhHk },
  common: { en: commonEn, 'zh-HK': commonZhHk },
  home: { en: homeEn, 'zh-HK': homeZhHk },
  marketplace: { en: marketplaceEn, 'zh-HK': marketplaceZhHk },
  nav: { en: navEn, 'zh-HK': navZhHk },
  offers: { en: offersEn, 'zh-HK': offersZhHk },
  property: { en: propertyEn, 'zh-HK': propertyZhHk },
  servicedResidence: { en: servicedResidenceEn, 'zh-HK': servicedResidenceZhHk },
  trend: { en: trendEn, 'zh-HK': trendZhHk },
} as const;

const localeMessageRoots = {
  ...localeModules,
  payments: { en: accountEn.payments, 'zh-HK': accountZhHk.payments },
} as const;

const staticTranslationKeyPattern = /\b(?:\$t|t)\(\s*(['\"])([A-Za-z][A-Za-z0-9_.-]*)\1/g;
const translationSourceModules = import.meta.glob(
  ['../**/*.{ts,vue}', '!../i18n/locales/**'],
  { eager: true, query: '?raw', import: 'default' },
) as Record<string, string>;

// 1. 收集巢狀物件與陣列中的所有訊息葉節點
const collectMessagePaths = (value: unknown, prefix = ''): string[] => {
  if (Array.isArray(value)) {
    return value.flatMap((item, index) => collectMessagePaths(item, `${prefix}[${index}]`));
  }

  if (value !== null && typeof value === 'object') {
    return Object.entries(value).flatMap(([key, item]) =>
      collectMessagePaths(item, prefix ? `${prefix}.${key}` : key));
  }

  return prefix ? [prefix] : [];
};

// 2. 收集直接呼叫 t() 或 $t() 的固定翻譯 key
const collectStaticTranslationKeys = (): string[] => {
  const keys = new Set<string>();

  Object.values(translationSourceModules).forEach((source) => {
    staticTranslationKeyPattern.lastIndex = 0;
    let match = staticTranslationKeyPattern.exec(source);

    while (match) {
      keys.add(match[2]);
      match = staticTranslationKeyPattern.exec(source);
    }
  });

  return [...keys].sort();
};

describe('i18n foundation', () => {
  afterEach(() => {
    applyLocale('zh-HK');
  });

  it('keeps English and zh-HK message keys in parity for every module', () => {
    Object.entries(localeModules).forEach(([moduleName, locales]) => {
      const englishPaths = collectMessagePaths(locales.en).sort();
      const traditionalChinesePaths = collectMessagePaths(locales['zh-HK']).sort();

      expect(traditionalChinesePaths, `${moduleName} locale keys`).toEqual(englishPaths);
    });
  });

  it('resolves every static translation key used by the frontend', () => {
    const availableKeys = new Set(
      Object.entries(localeMessageRoots).flatMap(([moduleName, locales]) =>
        collectMessagePaths(locales.en).map((key) => `${moduleName}.${key}`)),
    );
    const missingKeys = collectStaticTranslationKeys().filter((key) => !availableKeys.has(key));

    expect(missingKeys).toEqual([]);
  });

  it('exposes both supported locales and an English fallback', () => {
    expect(i18n.global.availableLocales.sort()).toEqual(['en', 'zh-HK']);
    expect(i18n.global.fallbackLocale.value).toBe('en');
  });

  it('updates vue-i18n and document language together', () => {
    applyLocale('en');

    expect(i18n.global.locale.value).toBe('en');
    expect(document.documentElement.lang).toBe('en');

    applyLocale('zh-HK');

    expect(i18n.global.locale.value).toBe('zh-HK');
    expect(document.documentElement.lang).toBe('zh-HK');
  });

  it('renders a literal at sign in property unit prices', () => {
    applyLocale('en');
    expect(i18n.global.t('property.publicList.unitPrice', { price: '10,000' })).toBe('@10,000/sqft');

    applyLocale('zh-HK');
    expect(i18n.global.t('property.publicList.unitPrice', { price: '10,000' })).toBe('@10,000/呎');
  });

  it('localizes the serviced-residence channel entry content', () => {
    applyLocale('en');
    expect(i18n.global.t('servicedResidence.channel.openList')).toBe('View listings');
    expect(i18n.global.t('servicedResidence.channel.publicTitle')).toBe('Short and monthly stays');

    applyLocale('zh-HK');
    expect(i18n.global.t('servicedResidence.channel.openList')).toBe('進入列表');
    expect(i18n.global.t('servicedResidence.channel.publicTitle')).toBe('短租與月租列表');
  });
});

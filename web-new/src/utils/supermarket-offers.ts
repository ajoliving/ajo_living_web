/*
 * 超市優惠顯示工具。
 * 1. 統一商店、分類、日期與港幣格式。
 * 2. 統一商品價格列、最低價與折扣計算。
 * 3. 供綜合優惠列表與商品詳情頁共用。
 */
import type { SupermarketProduct, SupermarketStorePrice } from '@/model/supermarket-offers';
import type { AppLocale } from '@/stores/preferences';

const STORE_LABELS: Record<AppLocale, Record<string, string>> = {
  'zh-HK': {
    AEON: 'AEON',
    WELLCOME: '惠康',
    PARKNSHOP: '百佳',
    JASONS: 'Market Place',
    LUNGFUNG: '龍豐',
    DCHFOOD: '大昌食品',
    WATSONS: '屈臣氏',
    MANNINGS: '萬寧',
    SASA: '莎莎',
  },
  en: {
    AEON: 'AEON',
    WELLCOME: 'Wellcome',
    PARKNSHOP: 'PARKnSHOP',
    JASONS: 'Market Place',
    LUNGFUNG: 'Lung Fung',
    DCHFOOD: 'DCH Food Mart',
    WATSONS: 'Watsons',
    MANNINGS: 'Mannings',
    SASA: 'Sa Sa',
  },
};

// 1. 標準化超市代碼
export const normalizeSupermarketStoreCode = (value: string): string => value.trim().toUpperCase();

// 2. 顯示超市名稱
export const displaySupermarketStore = (value: string, locale: AppLocale = 'zh-HK'): string => {
  const label = STORE_LABELS[locale][normalizeSupermarketStoreCode(value)] ?? value.trim();
  return label || (locale === 'en' ? 'Store not provided' : '未提供商店');
};

// 3. 顯示商品分類
export const displaySupermarketCategory = (value: string, locale: AppLocale = 'zh-HK'): string => {
  const normalized = value.split('/')[0]?.trim() ?? value;
  const labels = locale === 'en'
    ? {
        personalCare: 'Personal care',
        drinks: 'Drinks / Water',
        snacks: 'Snacks / Food',
        babyFormula: 'Baby formula',
        household: 'Household goods',
        groceries: 'Rice, oil and groceries',
        noodles: 'Noodles and pasta',
        fallback: 'Category not provided',
      }
    : {
        personalCare: '個人護理',
        drinks: '飲品 / 水',
        snacks: '零食 / 食品',
        babyFormula: '奶粉嬰兒',
        household: '家居用品',
        groceries: '米油雜貨',
        noodles: '粉麵食品',
        fallback: '未提供分類',
      };
  if (normalized.includes('個人護理')) return labels.personalCare;
  if (normalized.includes('飲品')) return labels.drinks;
  if (normalized.includes('糖果')) return labels.snacks;
  if (normalized.includes('奶粉')) return labels.babyFormula;
  if (normalized.includes('家居')) return labels.household;
  if (normalized.includes('米')) return labels.groceries;
  if (normalized.includes('粉麵')) return labels.noodles;
  return normalized || labels.fallback;
};

// 4. 格式化港幣
export const formatSupermarketHKPrice = (
  value: number | null | undefined,
  locale: AppLocale = 'zh-HK',
): string => {
  const numericValue = Number(value ?? 0);
  const safeValue = Number.isFinite(numericValue) ? numericValue : 0;
  return `HK$ ${safeValue.toLocaleString(locale, {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  })}`;
};

// 5. 格式化日期
export const formatSupermarketDate = (value: string | undefined, locale: AppLocale = 'zh-HK'): string => {
  if (!value) return '';
  const date = new Date(`${value}T00:00:00`);
  if (Number.isNaN(date.getTime())) return value;
  return new Intl.DateTimeFormat(locale, {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  }).format(date);
};

// 6. 格式化日期時間
export const formatSupermarketDateTime = (value: string | undefined, locale: AppLocale = 'zh-HK'): string => {
  if (!value) return '';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return date.toLocaleString(locale, {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  });
};

// 7. 取得商品分類文字
export const supermarketCategoryText = (product: SupermarketProduct): string =>
  [product.category1, product.category2, product.category3].filter(Boolean).join(' / ');

// 8. 建立後備價格列
export const supermarketFallbackStorePrice = (product: SupermarketProduct): SupermarketStorePrice => ({
  store: product.bestStore || product.stores?.[0] || '',
  listPrice: Number(product.listPrice ?? product.maxPrice ?? 0),
  effectiveUnitPrice: Number(product.effectiveUnitPrice ?? product.bestEffectiveUnitPrice ?? product.minPrice ?? 0),
  offer: product.bestOffer ?? '',
  parseStatus: product.parseStatus ?? '',
  snapshotDate: '',
});

// 9. 標準化商品各商店價格
export const supermarketStorePrices = (
  product: SupermarketProduct,
  locale: AppLocale = 'zh-HK',
): SupermarketStorePrice[] => {
  if (product.storePrices && product.storePrices.length > 0) {
    return [...product.storePrices].sort((a, b) => {
      if (a.effectiveUnitPrice === b.effectiveUnitPrice) {
        return displaySupermarketStore(a.store, locale).localeCompare(displaySupermarketStore(b.store, locale), locale);
      }
      return a.effectiveUnitPrice - b.effectiveUnitPrice;
    });
  }

  const stores = product.stores && product.stores.length > 0 ? product.stores : [product.bestStore || ''];
  const prices = stores.filter(Boolean).map((store) => ({
    store,
    listPrice: store === product.bestStore ? Number(product.listPrice ?? 0) : Number(product.maxPrice ?? product.listPrice ?? 0),
    effectiveUnitPrice:
      store === product.bestStore
        ? Number(product.effectiveUnitPrice ?? product.bestEffectiveUnitPrice ?? product.minPrice ?? 0)
        : Number(product.maxPrice ?? product.listPrice ?? 0),
    offer: store === product.bestStore ? product.bestOffer ?? '' : '',
    parseStatus: store === product.bestStore ? product.parseStatus ?? '' : 'none',
    snapshotDate: '',
  }));

  return prices.length > 0 ? prices : [supermarketFallbackStorePrice(product)];
};

// 10. 取得目前商店價格排序
export const supermarketCurrentStorePrices = (
  stores: SupermarketStorePrice[],
  locale: AppLocale = 'zh-HK',
): SupermarketStorePrice[] =>
  [...stores].sort((a, b) => {
    if (a.effectiveUnitPrice === b.effectiveUnitPrice) {
      return displaySupermarketStore(a.store, locale).localeCompare(displaySupermarketStore(b.store, locale), locale);
    }
    return a.effectiveUnitPrice - b.effectiveUnitPrice;
  });

// 11. 計算單一價格折扣率
export const supermarketPriceDiscountRate = (price: SupermarketStorePrice): number =>
  price.listPrice > 0 ? Math.max(0, ((price.listPrice - price.effectiveUnitPrice) / price.listPrice) * 100) : 0;

// 12. 取得有折扣的商店價格排序
export const supermarketDiscountStorePrices = (
  stores: SupermarketStorePrice[],
  locale: AppLocale = 'zh-HK',
): SupermarketStorePrice[] =>
  supermarketCurrentStorePrices(stores, locale)
    .filter((item) => supermarketPriceDiscountRate(item) > 0)
    .sort((a, b) => {
      const rateDiff = supermarketPriceDiscountRate(b) - supermarketPriceDiscountRate(a);
      if (rateDiff !== 0) return rateDiff;
      const savingDiff = (b.listPrice - b.effectiveUnitPrice) - (a.listPrice - a.effectiveUnitPrice);
      if (savingDiff !== 0) return savingDiff;
      return a.effectiveUnitPrice - b.effectiveUnitPrice;
    });

// 13. 取得最低價商店
export const supermarketBestStorePrices = (
  product: SupermarketProduct,
  locale: AppLocale = 'zh-HK',
): SupermarketStorePrice[] => {
  const prices = supermarketStorePrices(product, locale);
  const lowestPrice = Math.min(...prices.map((item) => item.effectiveUnitPrice));
  return prices.filter((item) => item.effectiveUnitPrice === lowestPrice);
};

// 14. 取得主要顯示價格
export const supermarketPrimaryPrice = (
  product: SupermarketProduct,
  locale: AppLocale = 'zh-HK',
): SupermarketStorePrice =>
  supermarketBestStorePrices(product, locale)[0] ?? supermarketFallbackStorePrice(product);

// 15. 取得最優惠商店
export const supermarketBestDealStorePrices = (
  stores: SupermarketStorePrice[],
  locale: AppLocale = 'zh-HK',
): SupermarketStorePrice[] => {
  const discountedStores = supermarketDiscountStorePrices(stores, locale);
  const primaryStore = discountedStores[0];
  if (!primaryStore) {
    return supermarketCurrentStorePrices(stores, locale).slice(0, 1);
  }

  const bestRate = supermarketPriceDiscountRate(primaryStore);
  return discountedStores.filter((item) => Math.abs(supermarketPriceDiscountRate(item) - bestRate) < 0.0001);
};

// 16. 取得商品優惠文字
export const supermarketOfferTexts = (
  product: SupermarketProduct,
  locale: AppLocale = 'zh-HK',
): string[] => {
  const values = supermarketBestStorePrices(product, locale).map((price) => price.offer).filter(Boolean);
  if (values.length === 0 && product.bestOffer) {
    values.push(product.bestOffer);
  }
  return Array.from(new Set(values.map((value) => value.trim()).filter(Boolean)));
};

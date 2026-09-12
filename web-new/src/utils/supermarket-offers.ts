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
  const stores = product.stores && product.stores.length > 0 ? product.stores : [product.bestStore || ''];
  const pricesByStore = new Map<string, SupermarketStorePrice>();
  (product.storePrices ?? []).forEach((price) => {
    const storeCode = normalizeSupermarketStoreCode(price.store);
    if (storeCode) {
      pricesByStore.set(storeCode, price);
    }
  });

  stores.filter(Boolean).forEach((store) => {
    const storeCode = normalizeSupermarketStoreCode(store);
    if (pricesByStore.has(storeCode)) {
      return;
    }
    pricesByStore.set(storeCode, {
      store,
      listPrice: store === product.bestStore ? Number(product.listPrice ?? 0) : Number(product.maxPrice ?? product.listPrice ?? 0),
      effectiveUnitPrice:
        store === product.bestStore
          ? Number(product.effectiveUnitPrice ?? product.bestEffectiveUnitPrice ?? product.minPrice ?? 0)
          : Number(product.maxPrice ?? product.listPrice ?? 0),
      offer: store === product.bestStore ? product.bestOffer ?? '' : '',
      parseStatus: store === product.bestStore ? product.parseStatus ?? '' : 'none',
      snapshotDate: '',
    });
  });

  const prices = [...pricesByStore.values()];

  return prices.length > 0
    ? supermarketCurrentStorePrices(prices, locale)
    : [supermarketFallbackStorePrice(product)];
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

// 12. 計算最低價相對下一個不同價格的優勢百分比。
export const supermarketSecondPriceAdvantageRate = (stores: SupermarketStorePrice[]): number => {
  const prices = [...new Set(
    supermarketCurrentStorePrices(stores)
      .map((item) => item.effectiveUnitPrice)
      .filter((value) => Number.isFinite(value) && value > 0),
  )].sort((a, b) => a - b);
  if (prices.length < 2) return 0;
  return ((prices[1] - prices[0]) / prices[1]) * 100;
};

// 13. 取得可顯示第二低價優勢的最低價門店。
export const supermarketSecondPriceAdvantageStores = (
  stores: SupermarketStorePrice[],
  locale: AppLocale = 'zh-HK',
): SupermarketStorePrice[] => {
  const sortedStores = supermarketCurrentStorePrices(stores, locale)
    .filter((item) => Number.isFinite(item.effectiveUnitPrice) && item.effectiveUnitPrice > 0);
  if (Math.round(supermarketSecondPriceAdvantageRate(sortedStores)) < 1) {
    return [];
  }
  const lowestPrice = sortedStores[0]?.effectiveUnitPrice;
  return sortedStores.filter((item) => item.effectiveUnitPrice === lowestPrice);
};

// 13. 回傳其中一個可顯示第二低價優勢的最低價門店。
export const supermarketSecondPriceAdvantageStore = (
  stores: SupermarketStorePrice[],
  locale: AppLocale = 'zh-HK',
): SupermarketStorePrice | null =>
  supermarketSecondPriceAdvantageStores(stores, locale)[0] ?? null;

// 14. 取得有折扣的商店價格排序
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

// 15. 取得最低價商店
export const supermarketBestStorePrices = (
  product: SupermarketProduct,
  locale: AppLocale = 'zh-HK',
): SupermarketStorePrice[] => {
  const prices = supermarketStorePrices(product, locale);
  const lowestPrice = Math.min(...prices.map((item) => item.effectiveUnitPrice));
  return prices.filter((item) => item.effectiveUnitPrice === lowestPrice);
};

// 16. 取得主要顯示價格
export const supermarketPrimaryPrice = (
  product: SupermarketProduct,
  locale: AppLocale = 'zh-HK',
): SupermarketStorePrice =>
  supermarketBestStorePrices(product, locale)[0] ?? supermarketFallbackStorePrice(product);

// 17. 取得最優惠商店
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

// 18. 取得商品優惠文字
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

// 19. 判斷純減價或原價貨品標籤。
export const supermarketIsSimpleOffer = (value: string): boolean => {
  const normalized = value.trim().toLocaleLowerCase();
  return normalized.includes('純減價')
    || normalized.includes('原價貨品')
    || normalized.includes('pure discount')
    || normalized.includes('regular price');
};

// 20. 移除沒有實際資訊的價格說明文字。
export const supermarketOfferDisplayText = (value: string): string => value
  .split(/[\r\n|]+/)
  .map((item) => item.trim())
  .filter(Boolean)
  .filter((item) => {
    const normalized = item.toLocaleLowerCase();
    return !normalized.includes('純減價')
      && !normalized.includes('原價貨品')
      && !normalized.includes('沒有額外條款')
      && !normalized.includes('目前售價')
      && !normalized.includes('pure discount')
      && !normalized.includes('regular price')
      && !normalized.includes('no additional terms')
      && !normalized.includes('current price');
  })
  .join(' / ');

// 19. 取得商品規格
const resolveProductUnit = (product: SupermarketProduct): string =>
  Object.values(product.variantLabels ?? {}).map((value) => value.trim()).filter(Boolean).join(' / ')
  || product.subtitle?.trim()
  || '';

// 20. 取得商品名稱並移除名稱末尾重複規格
export const resolveSupermarketProductName = (product: SupermarketProduct): string => {
  const name = product.name?.trim() || product.code;
  const unit = resolveProductUnit(product);
  if (!unit || !name.endsWith(unit)) return name;
  return name.slice(0, -unit.length).trim().replace(/[-–—|:/]+$/, '').trim() || name;
};

// 21. 取得商品品牌
export const resolveSupermarketProductBrand = (product: SupermarketProduct): string =>
  product.brand?.trim() || '';

// 22. 取得商品規格
export const resolveSupermarketProductUnit = (product: SupermarketProduct): string =>
  resolveProductUnit(product);

// 23. 取得商品完整標題
export const resolveSupermarketProductFullTitle = (product: SupermarketProduct): string =>
  [resolveSupermarketProductBrand(product), resolveSupermarketProductName(product), resolveSupermarketProductUnit(product)]
    .filter(Boolean)
    .join(' ');

// 23. 取得商品完整分類
export const resolveSupermarketProductCategory = (product: SupermarketProduct): string =>
  [product.category1, product.category2, product.category3].map((value) => value?.trim()).filter(Boolean).join(' / ');

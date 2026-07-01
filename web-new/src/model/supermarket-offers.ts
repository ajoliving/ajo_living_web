/*
 * 超市優惠型別。
 * 1. 定義 good-price 公開商品、摘要與詳情資料。
 * 2. 定義 AJO 會員收藏與價格提示資料。
 * 3. 供綜合優惠列表與詳情頁共用。
 */

export interface SupermarketDatasetStats {
  records: number;
  products: number;
  stores: number;
  offers: number;
  parsedOffers: number;
  unsupportedOffers: number;
  ambiguousOffers: number;
  uncalculatedOffers: number;
}

export interface SupermarketValueCount {
  value: string;
  count: number;
}

export interface SupermarketDatasetMetadata {
  latestSnapshotDate?: string;
  latestUpdatedAt?: string;
}

export interface SupermarketStorePrice {
  store: string;
  listPrice: number;
  effectiveUnitPrice: number;
  offer: string;
  parseStatus: string;
  snapshotDate: string;
}

export interface SupermarketProduct {
  code: string;
  name: string;
  brand: string;
  image_url?: string;
  imageUrl?: string;
  category1: string;
  category2: string;
  category3: string;
  minPrice: number;
  maxPrice: number;
  priceDiff: number;
  diffPercent: number;
  listPrice: number;
  effectiveUnitPrice: number;
  bestEffectiveUnitPrice: number;
  discountAmount: number;
  discountRate: number;
  stores: string[];
  storePrices?: SupermarketStorePrice[];
  hasOffer: boolean;
  bestOffer?: string;
  bestStore: string;
  offerType: string;
  offerPattern: string;
  parseStatus: string;
  isFavorite?: boolean;
}

export interface SupermarketSummary {
  stats: SupermarketDatasetStats;
  metadata?: SupermarketDatasetMetadata;
  categories: SupermarketValueCount[];
  stores: SupermarketValueCount[];
  cheapest: SupermarketProduct[];
  bestDiscounts: SupermarketProduct[];
  biggestDiffs: SupermarketProduct[];
  offers: SupermarketProduct[];
}

export interface SupermarketSearchResult {
  items: SupermarketProduct[];
  total: number;
  page: number;
  pageSize: number;
  stats: SupermarketDatasetStats;
  categories: string[];
  brands: string[];
  stores: string[];
}

export interface SupermarketHistorySummary {
  days: number;
  lowestList: number;
  highestList: number;
  lowestDeal: number;
  highestDeal: number;
}

export interface SupermarketDailyStorePrice {
  date: string;
  store: string;
  listPrice: number;
  effectiveUnitPrice: number;
  offer?: string;
}

export interface SupermarketPriceAlert {
  id: number;
  productCode: string;
  productName: string;
  targetPrice?: number;
  priceMode: 'list' | 'effective';
  offerRequired: boolean;
  enabled: boolean;
  lastTriggeredAt?: string;
  createdAt: string;
  updatedAt: string;
}

export interface SupermarketProductDetail {
  product: SupermarketProduct;
  summary: SupermarketHistorySummary;
  stores: SupermarketStorePrice[];
  history: SupermarketDailyStorePrice[];
  sameBrand: SupermarketProduct[];
  sameCategory: SupermarketProduct[];
  alertRule?: SupermarketPriceAlert | null;
  isFavorite: boolean;
}

export interface SupermarketSearchParams {
  q?: string;
  category?: string;
  brand?: string;
  store?: string;
  offerOnly?: boolean;
  sort?: string;
  page?: number;
  pageSize?: number;
}

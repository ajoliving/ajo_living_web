/*
 * 超市優惠顯示工具測試。
 * 1. 確認已記錄門店不會因部分價格資料缺失而消失。
 * 2. 確認第二低價優勢只標記唯一最低價門店。
 */
import { describe, expect, it } from 'vitest';

import type { SupermarketProduct, SupermarketStorePrice } from '@/model/supermarket-offers';

import {
  supermarketSecondPriceAdvantageRate,
  supermarketSecondPriceAdvantageStore,
  supermarketSecondPriceAdvantageStores,
  supermarketStorePrices,
} from './supermarket-offers';

// 1. 補齊已有門店的後備價格列。
describe('supermarketStorePrices', () => {
  it('keeps every recorded store when explicit prices cover only one store', () => {
    const product: SupermarketProduct = {
      code: 'P000000001',
      name: '測試商品',
      brand: '測試品牌',
      category1: '',
      category2: '',
      category3: '',
      minPrice: 20,
      maxPrice: 30,
      priceDiff: 10,
      diffPercent: 50,
      listPrice: 25,
      effectiveUnitPrice: 20,
      bestEffectiveUnitPrice: 20,
      discountAmount: 5,
      discountRate: 20,
      stores: ['WELLCOME', 'PARKNSHOP'],
      storePrices: [{
        store: 'WELLCOME',
        listPrice: 25,
        effectiveUnitPrice: 20,
        offer: '優惠',
        parseStatus: 'parsed',
        snapshotDate: '2026-08-17',
      }],
      hasOffer: true,
      bestOffer: '優惠',
      bestStore: 'WELLCOME',
      offerType: 'discount',
      offerPattern: '',
      parseStatus: 'parsed',
    };

    expect(supermarketStorePrices(product).map((item) => item.store)).toEqual(['WELLCOME', 'PARKNSHOP']);
  });
});

// 2. 驗證第二低價優勢標籤對應最低價門店。
describe('supermarketSecondPriceAdvantageStore', () => {
  const prices: SupermarketStorePrice[] = [
    {
      store: 'LUNGFUNG', listPrice: 178, effectiveUnitPrice: 178, offer: '', parseStatus: 'none', snapshotDate: '2026-08-29',
    },
    {
      store: 'SASA', listPrice: 188, effectiveUnitPrice: 188, offer: '', parseStatus: 'none', snapshotDate: '2026-08-29',
    },
    {
      store: 'WATSONS', listPrice: 449.25, effectiveUnitPrice: 449.25, offer: '', parseStatus: 'none', snapshotDate: '2026-08-29',
    },
    {
      store: 'MANNINGS', listPrice: 599, effectiveUnitPrice: 599, offer: '', parseStatus: 'none', snapshotDate: '2026-08-29',
    },
  ];

  it('returns only the lowest-priced store and calculates its saving against the second-lowest price', () => {
    expect(supermarketSecondPriceAdvantageStore(prices)?.store).toBe('LUNGFUNG');
    expect(supermarketSecondPriceAdvantageRate(prices)).toBeCloseTo((10 / 188) * 100, 8);
  });

  it('still marks every lowest-priced store when the cheapest price is tied', () => {
    const tiedPrices = [
      ...prices,
      {
        store: 'AEON', listPrice: 178, effectiveUnitPrice: 178, offer: '', parseStatus: 'none', snapshotDate: '2026-08-29',
      },
    ];
    expect(supermarketSecondPriceAdvantageStores(tiedPrices).map((item) => item.store).sort()).toEqual(['AEON', 'LUNGFUNG']);
    expect(supermarketSecondPriceAdvantageRate(tiedPrices)).toBeCloseTo((10 / 188) * 100, 8);
  });

  it('does not mark a store when every current price is the same', () => {
    expect(supermarketSecondPriceAdvantageStores([
      {
        store: 'JASONS', listPrice: 43.5, effectiveUnitPrice: 43.5, offer: '', parseStatus: 'none', snapshotDate: '2026-09-12',
      },
      {
        store: 'PARKNSHOP', listPrice: 43.5, effectiveUnitPrice: 43.5, offer: '', parseStatus: 'none', snapshotDate: '2026-09-12',
      },
      {
        store: 'WELLCOME', listPrice: 43.5, effectiveUnitPrice: 43.5, offer: '', parseStatus: 'none', snapshotDate: '2026-09-12',
      },
    ])).toEqual([]);
  });

  it('ignores invalid zero prices when comparing current stores', () => {
    expect(supermarketSecondPriceAdvantageStore([
      {
        store: 'AEON', listPrice: 0, effectiveUnitPrice: 0, offer: '', parseStatus: 'none', snapshotDate: '2026-09-12',
      },
      {
        store: 'LUNGFUNG', listPrice: 92, effectiveUnitPrice: 92, offer: '', parseStatus: 'none', snapshotDate: '2026-09-12',
      },
      {
        store: 'MANNINGS', listPrice: 340, effectiveUnitPrice: 340, offer: '', parseStatus: 'none', snapshotDate: '2026-09-12',
      },
    ])?.store).toBe('LUNGFUNG');
  });
});

/*
 * 超市優惠顯示工具測試。
 * 1. 確認已記錄門店不會因部分價格資料缺失而消失。
 */
import { describe, expect, it } from 'vitest';

import type { SupermarketProduct } from '@/model/supermarket-offers';

import { supermarketStorePrices } from './supermarket-offers';

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

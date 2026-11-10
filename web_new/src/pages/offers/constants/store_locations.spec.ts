/*
 * 超市門店座標測試。
 * 1. 缺少座標的門店不參與距離計算。
 * 2. 有效座標可計算附近門店。
 */
import { describe, expect, it } from 'vitest';
import { nearestSupermarketStoreLocations, supermarketStoreLocations } from './store_locations';

describe('supermarket store locations', () => {
  it('excludes stores with missing coordinates instead of treating them as zero', () => {
    expect(supermarketStoreLocations.some((store) =>
      store.address === '新界上水新康街66號A - D號',
    )).toBe(false);
    expect(supermarketStoreLocations.every((store) =>
      store.latitude !== 0 && store.longitude !== 0,
    )).toBe(true);
  });

  it('calculates distance using valid store coordinates', () => {
    const nearby = nearestSupermarketStoreLocations(22.5043543, 114.1288134, ['LUNGFUNG'], 1);

    expect(nearby[0]?.name).toBe('龍豐Mall');
    expect(nearby[0]?.distanceKm).toBeCloseTo(0);
  });
});

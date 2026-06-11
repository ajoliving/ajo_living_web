/*
 * 超市優惠門店位置資料。
 * 1. 匯入 good-price 參考頁同一批門店地址文字。
 * 2. 解析地址、營業時間與座標。
 * 3. 提供最近門店與 Google Maps 路線工具。
 * 4. 提供門店地址公開來源連結。
 */
import aeonStoreAddressText from '../assets/aeon_store_addresses.txt?raw';
import lungFungStoreAddressText from '../assets/lungfung_store_addresses.txt?raw';
import manningsStoreAddressText from '../assets/mannings_store_addresses.txt?raw';
import marketplaceStoreAddressText from '../assets/marketplace_store_addresses.txt?raw';
import pnsStoreAddressText from '../assets/pns_store_addresses.txt?raw';
import sasaStoreAddressText from '../assets/sasa_store_addresses.txt?raw';
import watsonsStoreAddressText from '../assets/watsons_store_addresses.txt?raw';
import wellcomeStoreAddressText from '../assets/wellcome_store_addresses.txt?raw';

export interface SupermarketStoreLocation {
  id: string;
  storeCode: string;
  name: string;
  address: string;
  phone: string;
  hours: string;
  latitude: number;
  longitude: number;
}

export interface SupermarketNearbyStoreLocation extends SupermarketStoreLocation {
  distanceKm: number;
  originLatitude: number;
  originLongitude: number;
}

export interface SupermarketStoreAddressSourceLink {
  name: string;
  url: string;
}

interface StoreAddressSource {
  storeCode: string;
  rawText: string;
}

export const SUPERMARKET_NEARBY_STORE_LIMIT = 3;
export const SUPERMARKET_GEOLOCATION_CACHE_MAX_AGE_MS = 5 * 60 * 1000;
export const supermarketStoreAddressSourceLinks: SupermarketStoreAddressSourceLink[] = [
  { name: '龍豐 Lung Fung', url: 'https://www.lungfung.hk/location/' },
  { name: 'Wellcome 惠康', url: 'https://www.wellcome.com.hk/zh-hant/store_locator' },
  { name: 'PARKnSHOP 百佳', url: 'https://www.pns.hk/zh-hk/store-finder' },
  { name: 'Market Place / JASONS', url: 'https://www.marketplacehk.com/zh-hant/store_locator' },
  { name: 'SaSa 莎莎', url: 'https://www.sasa.com.hk/v2/Shop/StoreList/17?lang=zh-HK' },
  { name: 'AEON 綜合百貨', url: 'https://www.aeonstores.com.hk/shop_info?q%5Bshop_info_category_shop_info_category_type_master_id_eq%5D=6' },
  { name: 'AEON 獨立超級市場', url: 'https://www.aeonstores.com.hk/shop_info?q%5Bshop_info_category_shop_info_category_type_master_id_eq%5D=5' },
  { name: 'Mannings 萬寧', url: 'https://www.mannings.com.hk/store-finder' },
  { name: 'Watsons 屈臣氏', url: 'https://www.watsons.com.hk/zh-hk/store-finder' },
];

const storeAddressSources: StoreAddressSource[] = [
  { storeCode: 'AEON', rawText: aeonStoreAddressText },
  { storeCode: 'WELLCOME', rawText: wellcomeStoreAddressText },
  { storeCode: 'PARKNSHOP', rawText: pnsStoreAddressText },
  { storeCode: 'JASONS', rawText: marketplaceStoreAddressText },
  { storeCode: 'LUNGFUNG', rawText: lungFungStoreAddressText },
  { storeCode: 'WATSONS', rawText: watsonsStoreAddressText },
  { storeCode: 'MANNINGS', rawText: manningsStoreAddressText },
  { storeCode: 'SASA', rawText: sasaStoreAddressText },
];

export const supermarketStoreLocations = storeAddressSources.flatMap((source) =>
  parseStoreLocations(source.rawText, source.storeCode),
);

// 1. 解析單一商戶門店文字
function parseStoreLocations(rawText: string, storeCode: string): SupermarketStoreLocation[] {
  const stores: SupermarketStoreLocation[] = [];
  let current: Partial<SupermarketStoreLocation> | null = null;

  const commitCurrent = (): void => {
    const latitude = current?.latitude;
    const longitude = current?.longitude;
    if (
      !current?.name ||
      !current.address ||
      latitude === undefined ||
      longitude === undefined ||
      !Number.isFinite(latitude) ||
      !Number.isFinite(longitude)
    ) {
      return;
    }

    stores.push({
      id: current.id || `${current.name}-${latitude}-${longitude}`,
      storeCode,
      name: current.name,
      address: current.address,
      phone: current.phone ?? '',
      hours: current.hours ?? '',
      latitude,
      longitude,
    });
  };

  rawText.split(/\r?\n/).forEach((line) => {
    const trimmed = line.trim();
    const titleMatch = trimmed.match(/^\d+\.\s+(.+)$/);
    if (titleMatch) {
      commitCurrent();
      current = {
        id: '',
        storeCode,
        name: titleMatch[1].trim(),
        address: '',
        phone: '',
        hours: '',
        latitude: Number.NaN,
        longitude: Number.NaN,
      };
      return;
    }

    if (!current) {
      return;
    }

    if (trimmed.startsWith('地址：')) {
      current.address = trimmed.slice('地址：'.length).trim();
    }
    if (trimmed.startsWith('電話：')) {
      current.phone = trimmed.slice('電話：'.length).trim();
    }
    if (trimmed.startsWith('營業時間：')) {
      current.hours = trimmed.slice('營業時間：'.length).trim();
    }
    if (trimmed.startsWith('門店營業時間：')) {
      current.hours = trimmed.slice('門店營業時間：'.length).trim();
    }
    if (trimmed.startsWith('緯度：')) {
      current.latitude = Number(trimmed.slice('緯度：'.length).trim());
    }
    if (trimmed.startsWith('經度：')) {
      current.longitude = Number(trimmed.slice('經度：'.length).trim());
    }
    if (trimmed.startsWith('門店 ID：')) {
      current.id = trimmed.slice('門店 ID：'.length).trim();
    }
    if (trimmed.startsWith('SAP ID：') && !current.id) {
      current.id = trimmed.slice('SAP ID：'.length).trim();
    }
    if (trimmed.startsWith('地圖：')) {
      const coordinates = coordinatesFromMapUrl(trimmed.slice('地圖：'.length).trim());
      if (coordinates) {
        current.latitude = coordinates.latitude;
        current.longitude = coordinates.longitude;
      }
    }
  });

  commitCurrent();
  return stores;
}

// 2. 從 Google Maps URL 解析座標
function coordinatesFromMapUrl(url: string): { latitude: number; longitude: number } | null {
  const match = url.match(/@(-?\d+(?:\.\d+)?),(-?\d+(?:\.\d+)?)/);
  if (!match) {
    return null;
  }

  const latitude = Number(match[1]);
  const longitude = Number(match[2]);
  if (!Number.isFinite(latitude) || !Number.isFinite(longitude)) {
    return null;
  }

  return { latitude, longitude };
}

// 3. 取得指定商戶最近門店
export function nearestSupermarketStoreLocations(
  latitude: number,
  longitude: number,
  storeCodes: string[],
  limit = SUPERMARKET_NEARBY_STORE_LIMIT,
): SupermarketNearbyStoreLocation[] {
  const eligibleStoreCodes = new Set(storeCodes.map(normalizeSupermarketStoreCode));
  return supermarketStoreLocations
    .filter((storeItem) => eligibleStoreCodes.has(normalizeSupermarketStoreCode(storeItem.storeCode)))
    .map((storeItem) => ({
      ...storeItem,
      distanceKm: haversineDistanceKm(latitude, longitude, storeItem.latitude, storeItem.longitude),
      originLatitude: latitude,
      originLongitude: longitude,
    }))
    .filter((storeItem) => Number.isFinite(storeItem.distanceKm))
    .sort((a, b) => a.distanceKm - b.distanceKm)
    .slice(0, limit);
}

export function normalizeSupermarketStoreCode(storeCode: string): string {
  return storeCode.trim().toUpperCase();
}

// 4. 建立 Google Maps 步行路線 URL
export function supermarketDirectionsUrl(store: SupermarketNearbyStoreLocation): string {
  const params = new URLSearchParams({
    api: '1',
    origin: `${store.originLatitude},${store.originLongitude}`,
    destination: `${store.latitude},${store.longitude}`,
    travelmode: 'walking',
  });
  return `https://www.google.com/maps/dir/?${params.toString()}`;
}

// 5. 格式化門店距離
export function formatSupermarketDistanceKm(distanceKm: number): string {
  if (distanceKm < 1) {
    return `${Math.round(distanceKm * 1000)} m`;
  }
  return `${distanceKm.toFixed(1)} km`;
}

// 6. 計算兩個座標之間的距離
function haversineDistanceKm(
  fromLatitude: number,
  fromLongitude: number,
  toLatitude: number,
  toLongitude: number,
): number {
  const earthRadiusKm = 6371;
  const latitudeDelta = degreesToRadians(toLatitude - fromLatitude);
  const longitudeDelta = degreesToRadians(toLongitude - fromLongitude);
  const fromLatitudeRadians = degreesToRadians(fromLatitude);
  const toLatitudeRadians = degreesToRadians(toLatitude);
  const value = Math.sin(latitudeDelta / 2) ** 2 +
    Math.cos(fromLatitudeRadians) * Math.cos(toLatitudeRadians) * Math.sin(longitudeDelta / 2) ** 2;
  return earthRadiusKm * 2 * Math.atan2(Math.sqrt(value), Math.sqrt(1 - value));
}

// 7. 角度轉弧度
function degreesToRadians(value: number): number {
  return (value * Math.PI) / 180;
}

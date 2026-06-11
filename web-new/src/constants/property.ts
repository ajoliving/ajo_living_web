/*
 * 物業頻道常量。
 * 1. 統一樓盤放售與服務式住宅表單選項。
 * 2. 復用二手頻道地區碼以保持後端校驗一致。
 */
import type { AppLocale } from '@/stores/preferences';
import {
  getMarketplaceDistrictLabel,
  getMarketplaceOptionLabel,
  marketplaceDistricts,
  type MarketplaceDistrictCode,
  type MarketplaceLabelledOption,
} from '@/constants/marketplace';

export type PropertyDistrictCode = MarketplaceDistrictCode;
export type PropertyTypeCode = 'private_flat' | 'estate' | 'house' | 'office' | 'shop' | 'car_park' | 'industrial';
export type PropertyContactMethod = 'phone' | 'whatsapp' | 'chat' | 'both' | 'chat_or_whatsapp';
export type PropertyBusinessStatus = 'available' | 'sold';

export interface PropertyRangeOption {
  value: string;
  label: string;
  min?: number;
  max?: number;
}

export interface PropertyFilterOption<TValue extends string = string> {
  value: TValue;
  label: string;
}

export const propertyRegionFilterOptions: PropertyFilterOption[] = [
  { value: '', label: '全港' },
  { value: 'hong_kong_island', label: '香港島' },
  { value: 'kowloon', label: '九龍' },
  { value: 'new_territories', label: '新界' },
  { value: 'outlying_islands', label: '離島' },
];

export const propertyTransactionTypeFilterOptions: PropertyFilterOption[] = [
  { value: '', label: '全部' },
  { value: 'sale', label: '出售' },
  { value: 'rent', label: '出租' },
];

export const propertyTypeFilterOptions: PropertyFilterOption[] = [
  { value: '', label: '全部' },
  { value: 'residential', label: '住宅' },
  { value: 'car_park', label: '車位' },
  { value: 'industrial', label: '工業' },
  { value: 'office', label: '商廈' },
];

export const propertyPriceRangeFilterOptions: PropertyRangeOption[] = [
  { value: '', label: '不限' },
  { value: 'under_4m', label: '400萬以下', max: 4_000_000 },
  { value: '4m_8m', label: '400-800萬', min: 4_000_000, max: 8_000_000 },
  { value: '8m_12m', label: '800-1200萬', min: 8_000_000, max: 12_000_000 },
  { value: 'over_20m', label: '2000萬+', min: 20_000_000 },
];

export const propertyAreaRangeFilterOptions: PropertyRangeOption[] = [
  { value: '', label: '不限' },
  { value: 'under_300', label: '300呎以下', max: 300 },
  { value: '300_500', label: '300-500呎', min: 300, max: 500 },
  { value: '500_1000', label: '500-1000呎', min: 500, max: 1000 },
  { value: 'over_1000', label: '1000呎+', min: 1000 },
];

export const propertyBedroomFilterOptions: PropertyFilterOption[] = [
  { value: '', label: '全部' },
  { value: '0', label: '開放式' },
  { value: '1', label: '1房' },
  { value: '2', label: '2房' },
  { value: '3', label: '3房' },
  { value: '4', label: '4房+' },
];

export const propertyRenovationFilterOptions: PropertyFilterOption[] = [
  { value: '', label: '全部' },
  { value: 'brand_new', label: '全新' },
  { value: 'renovated', label: '有裝修' },
  { value: 'simple', label: '簡潔' },
  { value: 'special', label: '特色' },
];

export const propertyPublisherFilterOptions: PropertyFilterOption[] = [
  { value: '', label: '全部' },
  { value: 'owner', label: '業主' },
  { value: 'agent', label: '代理' },
];

export const propertyTypeOptions: MarketplaceLabelledOption<PropertyTypeCode>[] = [
  { value: 'private_flat', label_zh_hk: '私人住宅', label_en: 'Private flat' },
  { value: 'estate', label_zh_hk: '屋苑單位', label_en: 'Estate flat' },
  { value: 'house', label_zh_hk: '洋房', label_en: 'House' },
  { value: 'office', label_zh_hk: '寫字樓', label_en: 'Office' },
  { value: 'shop', label_zh_hk: '商舖', label_en: 'Shop' },
  { value: 'car_park', label_zh_hk: '車位', label_en: 'Car park' },
];

export const propertyContactMethodOptions: MarketplaceLabelledOption<PropertyContactMethod>[] = [
  { value: 'both', label_zh_hk: '電話及 WhatsApp', label_en: 'Phone and WhatsApp' },
  { value: 'whatsapp', label_zh_hk: 'WhatsApp', label_en: 'WhatsApp' },
  { value: 'phone', label_zh_hk: '電話', label_en: 'Phone' },
  { value: 'chat', label_zh_hk: '站內聊天', label_en: 'In-app chat' },
  { value: 'chat_or_whatsapp', label_zh_hk: '聊天或 WhatsApp', label_en: 'Chat or WhatsApp' },
];

export const propertyBusinessStatusOptions: MarketplaceLabelledOption<PropertyBusinessStatus>[] = [
  { value: 'available', label_zh_hk: '可放售', label_en: 'Available' },
  { value: 'sold', label_zh_hk: '已成交', label_en: 'Sold' },
];

export const propertyFeatureTagOptions: MarketplaceLabelledOption[] = [
  { value: 'near_mtr', label_zh_hk: '近港鐵', label_en: 'Near MTR' },
  { value: 'sea_view', label_zh_hk: '海景', label_en: 'Sea view' },
  { value: 'renovated', label_zh_hk: '已裝修', label_en: 'Renovated' },
  { value: 'clubhouse', label_zh_hk: '會所', label_en: 'Clubhouse' },
  { value: 'balcony', label_zh_hk: '露台', label_en: 'Balcony' },
  { value: 'pet_friendly', label_zh_hk: '可養寵物', label_en: 'Pet friendly' },
];

export const servicedFacilityTagOptions: MarketplaceLabelledOption[] = [
  { value: 'gym', label_zh_hk: '健身室', label_en: 'Gym' },
  { value: 'laundry', label_zh_hk: '洗衣房', label_en: 'Laundry' },
  { value: 'workspace', label_zh_hk: '共享工作區', label_en: 'Workspace' },
  { value: 'lounge', label_zh_hk: '住客休息室', label_en: 'Resident lounge' },
  { value: 'parking', label_zh_hk: '泊車', label_en: 'Parking' },
];

export const servicedServiceTagOptions: MarketplaceLabelledOption[] = [
  { value: 'housekeeping', label_zh_hk: '房務清潔', label_en: 'Housekeeping' },
  { value: 'wifi', label_zh_hk: 'Wi-Fi', label_en: 'Wi-Fi' },
  { value: 'utilities', label_zh_hk: '水電煤', label_en: 'Utilities' },
  { value: 'front_desk', label_zh_hk: '前台服務', label_en: 'Front desk' },
  { value: 'linen', label_zh_hk: '床品更換', label_en: 'Linen service' },
];

// 1. 取得物業選項標籤
export const getPropertyOptionLabel = (
  option: MarketplaceLabelledOption,
  locale: AppLocale,
) => getMarketplaceOptionLabel(option, locale);

// 2. 取得樓盤類型標籤
export const getPropertyTypeLabel = (value: string, locale: AppLocale) => {
  const option = propertyTypeOptions.find((item) => item.value === value);

  return option ? getPropertyOptionLabel(option, locale) : value;
};

// 3. 取得地區標籤
export const getPropertyDistrictLabel = (value: string, locale: AppLocale) =>
  getMarketplaceDistrictLabel(value, locale);

export { marketplaceDistricts };

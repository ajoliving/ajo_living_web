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
  marketplaceRegions,
  type MarketplaceDistrictCode,
  type MarketplaceLabelledOption,
} from '@/constants/marketplace';

export type PropertyDistrictCode = MarketplaceDistrictCode;
export type PropertyTypeCode = 'private_flat' | 'estate' | 'house' | 'office' | 'shop' | 'car_park' | 'industrial';
export type PropertyContactMethod = 'phone' | 'whatsapp' | 'chat' | 'both' | 'chat_or_whatsapp';
export type PropertyBusinessStatus = 'available' | 'sold';
export type PropertyTransactionTypeCode = 'sale' | 'rent';
export type PropertyAreaModeCode = 'usable' | 'gross';
export type PropertyFloorZoneCode = 'low' | 'middle' | 'high';
export type PropertyKitchenTypeCode = 'enclosed' | 'open' | 'pantry' | 'none';
export type PropertyCookingModeCode = 'gas' | 'electric' | 'induction' | 'no_cooking';
export type PropertyAdPackageCode = 'basic' | 'featured' | 'premium' | 'fast_sale';
export type PropertyLocationScopeCode = 'local' | 'overseas';
export type ServicedStayUnitCode = 'month' | 'day';

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

export interface PropertyAdPackageOption extends MarketplaceLabelledOption<PropertyAdPackageCode> {
  weight: number;
  price_hkd: number;
  price_points: number;
  duration_days: number;
}

export const propertyRegionFilterOptions: PropertyFilterOption[] = [
  { value: '', label: '全港' },
  { value: 'hong_kong_island', label: '香港島' },
  { value: 'kowloon', label: '九龍' },
  { value: 'new_territories', label: '新界' },
  { value: 'outlying_islands', label: '離島' },
];

const propertyRegionDisplayOptions: MarketplaceLabelledOption[] = [
  { value: 'hong_kong_island', label_zh_hk: '香港島', label_en: 'Hong Kong Island' },
  { value: 'kowloon', label_zh_hk: '九龍', label_en: 'Kowloon' },
  { value: 'new_territories', label_zh_hk: '新界', label_en: 'New Territories' },
  { value: 'outlying_islands', label_zh_hk: '離島', label_en: 'Outlying Islands' },
];

export const propertyTransactionTypeFilterOptions: PropertyFilterOption[] = [
  { value: '', label: '全部' },
  { value: 'sale', label: '出售' },
  { value: 'rent', label: '出租' },
];

export const propertyTransactionTypeOptions: MarketplaceLabelledOption<PropertyTransactionTypeCode>[] = [
  { value: 'sale', label_zh_hk: '放售', label_en: 'Sale' },
  { value: 'rent', label_zh_hk: '放租', label_en: 'Rent' },
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
  { value: 'industrial', label_zh_hk: '工業', label_en: 'Industrial' },
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

export const propertyAreaModeOptions: MarketplaceLabelledOption<PropertyAreaModeCode>[] = [
  { value: 'usable', label_zh_hk: '實用面積', label_en: 'Usable area' },
  { value: 'gross', label_zh_hk: '建築面積', label_en: 'Gross area' },
];

export const propertyLocationScopeOptions: MarketplaceLabelledOption<PropertyLocationScopeCode>[] = [
  { value: 'local', label_zh_hk: '本地', label_en: 'Local' },
  { value: 'overseas', label_zh_hk: '海外', label_en: 'Overseas' },
];

export const propertyListingCategoryOptions: MarketplaceLabelledOption[] = [
  { value: 'standard', label_zh_hk: '一般放盤', label_en: 'Standard listing' },
  { value: 'new_development', label_zh_hk: '新盤', label_en: 'New development' },
  { value: 'developer_project', label_zh_hk: '發展商項目', label_en: 'Developer project' },
  { value: 'multi_unit', label_zh_hk: '多於一伙', label_en: 'Multiple units' },
];

export const propertyFloorZoneOptions: MarketplaceLabelledOption<PropertyFloorZoneCode>[] = [
  { value: 'low', label_zh_hk: '低層', label_en: 'Low floor' },
  { value: 'middle', label_zh_hk: '中層', label_en: 'Middle floor' },
  { value: 'high', label_zh_hk: '高層', label_en: 'High floor' },
];

export const propertyKitchenTypeOptions: MarketplaceLabelledOption<PropertyKitchenTypeCode>[] = [
  { value: 'enclosed', label_zh_hk: '獨立廚房', label_en: 'Enclosed kitchen' },
  { value: 'open', label_zh_hk: '開放式廚房', label_en: 'Open kitchen' },
  { value: 'pantry', label_zh_hk: '茶水間', label_en: 'Pantry' },
  { value: 'none', label_zh_hk: '無廚房', label_en: 'No kitchen' },
];

export const propertyCookingModeOptions: MarketplaceLabelledOption<PropertyCookingModeCode>[] = [
  { value: 'gas', label_zh_hk: '明火煮食', label_en: 'Gas cooking' },
  { value: 'electric', label_zh_hk: '電煮食', label_en: 'Electric cooking' },
  { value: 'induction', label_zh_hk: '電磁爐', label_en: 'Induction' },
  { value: 'no_cooking', label_zh_hk: '不可煮食', label_en: 'No cooking' },
];

export const propertyAdPackageOptions: PropertyAdPackageOption[] = [
  { value: 'basic', label_zh_hk: '普通', label_en: 'Basic', weight: 0, price_hkd: 600, price_points: 600, duration_days: 30 },
  { value: 'featured', label_zh_hk: '置頂', label_en: 'Featured', weight: 1, price_hkd: 800, price_points: 800, duration_days: 30 },
  { value: 'premium', label_zh_hk: '黃金置頂', label_en: 'Premium featured', weight: 2, price_hkd: 1500, price_points: 1500, duration_days: 30 },
  { value: 'fast_sale', label_zh_hk: '即走盤', label_en: 'Fast sale', weight: 3, price_hkd: 1200, price_points: 1200, duration_days: 15 },
];

export const servicedAdPackageOptions: PropertyAdPackageOption[] = propertyAdPackageOptions.map((option) => ({
  ...option,
  price_points: option.price_points + 200,
}));

export const propertyAnnualPrepayOptions: MarketplaceLabelledOption[] = [
  { value: '95_off', label_zh_hk: '年繳 95 折', label_en: '5% annual discount' },
  { value: '90_off', label_zh_hk: '年繳 9 折', label_en: '10% annual discount' },
];

export const servicedStayUnitOptions: MarketplaceLabelledOption<ServicedStayUnitCode>[] = [
  { value: 'month', label_zh_hk: '個月', label_en: 'Month' },
  { value: 'day', label_zh_hk: '日', label_en: 'Day' },
];

export const propertyFeatureTagOptions: MarketplaceLabelledOption[] = [
  { value: 'brand_new', label_zh_hk: '全新', label_en: 'Brand new' },
  { value: 'view', label_zh_hk: '有景觀', label_en: 'Open view' },
  { value: 'renovated', label_zh_hk: '有裝修', label_en: 'Renovated' },
  { value: 'appliances', label_zh_hk: '連電器', label_en: 'Appliances included' },
  { value: 'furnished', label_zh_hk: '連傢俬', label_en: 'Furnished' },
  { value: 'exclusive', label_zh_hk: '獨家盤', label_en: 'Exclusive' },
  { value: 'pet_friendly', label_zh_hk: '可養寵物', label_en: 'Pet friendly' },
  { value: 'parking_included', label_zh_hk: '連車位', label_en: 'Parking included' },
  { value: 'special_unit', label_zh_hk: '單位特色', label_en: 'Special unit' },
  { value: 'commercial_use', label_zh_hk: '工商專用', label_en: 'Commercial use' },
];

export const propertyAutoTagOptions: MarketplaceLabelledOption[] = [
  { value: 'vr_video', label_zh_hk: '有 VR 或影片', label_en: 'VR or video' },
  { value: 'prepay_discount', label_zh_hk: '預繳全年租金優惠', label_en: 'Annual prepay discount' },
];

export const servicedFacilityTagOptions: MarketplaceLabelledOption[] = [
  { value: 'gym', label_zh_hk: '健身室', label_en: 'Gym' },
  { value: 'laundry', label_zh_hk: '洗衣房', label_en: 'Laundry' },
  { value: 'workspace', label_zh_hk: '共享工作區', label_en: 'Workspace' },
  { value: 'lounge', label_zh_hk: '住客休息室', label_en: 'Resident lounge' },
  { value: 'parking', label_zh_hk: '泊車', label_en: 'Parking' },
  { value: 'restaurant', label_zh_hk: '餐廳', label_en: 'Restaurant' },
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

// 3. 取得樓盤標籤
export const getPropertyTagLabel = (value: string, locale: AppLocale) => {
  const options = [
    ...propertyFeatureTagOptions,
    ...propertyAutoTagOptions,
    ...servicedFacilityTagOptions,
    ...servicedServiceTagOptions,
  ];
  const option = options.find((item) => item.value === value);

  return option ? getPropertyOptionLabel(option, locale) : value;
};

// 4. 取得地區標籤
export const getPropertyDistrictLabel = (value: string, locale: AppLocale) => {
  const regionOption = propertyRegionDisplayOptions.find((item) => item.value === value) ||
    marketplaceRegions.find((item) => item.value === value);
  if (regionOption) {
    return getPropertyOptionLabel(regionOption, locale);
  }

  return getMarketplaceDistrictLabel(value, locale);
};

export { marketplaceDistricts };

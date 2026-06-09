/*
 * 物業頻道常量。
 * 1. 統一樓盤放售與服務式住宅表單選項。
 * 2. 復用二手頻道地區碼以保持後端校驗一致。
 */
import type { AppLocale } from '@/app/stores/preferences';
import {
  getMarketplaceDistrictLabel,
  getMarketplaceOptionLabel,
  marketplaceDistricts,
  type MarketplaceDistrictCode,
  type MarketplaceLabelledOption,
} from '@/domains/marketplace/constants';

export type PropertyDistrictCode = MarketplaceDistrictCode;
export type PropertyTransactionTypeCode = 'sale' | 'rent';
export type PropertyTypeCode =
  | 'residential'
  | 'car_park'
  | 'commercial'
  | 'shop'
  | 'land'
  | 'private_flat'
  | 'estate'
  | 'house';
export type PropertyRentalTypeCode = 'subdivided' | 'shared' | 'short_term';
export type PropertyAreaModeCode = 'usable' | 'gross';
export type PropertyContactMethod = 'phone' | 'whatsapp' | 'chat' | 'both' | 'chat_or_whatsapp';
export type PropertyBusinessStatus = 'available' | 'sold';
export type PropertyLocationScope = 'local' | 'overseas';
export type PropertyAdPackageCode = 'basic' | 'featured' | 'premium' | 'fast_sale';
export type PropertyFloorZone = 'low' | 'middle' | 'high';

export interface PropertyRangeOption extends MarketplaceLabelledOption {
  min?: number;
  max?: number;
}

export interface PropertyTagGroup {
  key: string;
  label_zh_hk: string;
  label_en: string;
  options: MarketplaceLabelledOption[];
}

export interface PropertyAdPackageOption extends MarketplaceLabelledOption<PropertyAdPackageCode> {
  weight: number;
  price_hkd: number;
  price_points: number;
  duration_days: number;
}

export const propertyTransactionTypeOptions: MarketplaceLabelledOption<PropertyTransactionTypeCode>[] = [
  { value: 'sale', label_zh_hk: '放售', label_en: 'For sale' },
  { value: 'rent', label_zh_hk: '放租', label_en: 'For rent' },
];

export const propertyTypeOptions: MarketplaceLabelledOption<PropertyTypeCode>[] = [
  { value: 'residential', label_zh_hk: '住宅', label_en: 'Residential' },
  { value: 'car_park', label_zh_hk: '車位', label_en: 'Car park' },
  { value: 'commercial', label_zh_hk: '工商', label_en: 'Commercial' },
  { value: 'shop', label_zh_hk: '商舖', label_en: 'Shop' },
  { value: 'land', label_zh_hk: '土地', label_en: 'Land' },
];

const legacyPropertyTypeOptions: MarketplaceLabelledOption<PropertyTypeCode>[] = [
  { value: 'private_flat', label_zh_hk: '住宅', label_en: 'Residential' },
  { value: 'estate', label_zh_hk: '住宅', label_en: 'Residential' },
  { value: 'house', label_zh_hk: '住宅', label_en: 'Residential' },
];

export const propertyRentalTypeOptions: MarketplaceLabelledOption<PropertyRentalTypeCode>[] = [
  { value: 'subdivided', label_zh_hk: '分間單位', label_en: 'Subdivided unit' },
  { value: 'shared', label_zh_hk: '分租', label_en: 'Shared rent' },
  { value: 'short_term', label_zh_hk: '短租', label_en: 'Short-term rent' },
];

export const propertyAreaModeOptions: MarketplaceLabelledOption<PropertyAreaModeCode>[] = [
  { value: 'usable', label_zh_hk: '實用面積', label_en: 'Usable area' },
  { value: 'gross', label_zh_hk: '建築面積', label_en: 'Gross area' },
];

export const propertyPriceRangeOptions: PropertyRangeOption[] = [
  { value: 'under_5000', label_zh_hk: '5000 元以下', label_en: 'Under HK$5,000', max: 5000 },
  { value: '5000_10000', label_zh_hk: '5000 元-10000 元', label_en: 'HK$5,000 - HK$10,000', min: 5000, max: 10000 },
  { value: '10000_15000', label_zh_hk: '10000 元-15000 元', label_en: 'HK$10,000 - HK$15,000', min: 10000, max: 15000 },
  { value: '15000_20000', label_zh_hk: '15000 元-20000 元', label_en: 'HK$15,000 - HK$20,000', min: 15000, max: 20000 },
  { value: '20000_40000', label_zh_hk: '20000 元-40000 元', label_en: 'HK$20,000 - HK$40,000', min: 20000, max: 40000 },
  { value: 'over_40000', label_zh_hk: '40000 元以上', label_en: 'Over HK$40,000', min: 40000 },
];

export const propertySalePriceRangeOptions: PropertyRangeOption[] = [
  { value: 'under_4000000', label_zh_hk: '400 萬以下', label_en: 'Under HK$4M', max: 4000000 },
  { value: '4000000_8000000', label_zh_hk: '400 萬-800 萬', label_en: 'HK$4M - HK$8M', min: 4000000, max: 8000000 },
  { value: '8000000_12000000', label_zh_hk: '800 萬-1200 萬', label_en: 'HK$8M - HK$12M', min: 8000000, max: 12000000 },
  { value: '12000000_20000000', label_zh_hk: '1200 萬-2000 萬', label_en: 'HK$12M - HK$20M', min: 12000000, max: 20000000 },
  { value: 'over_20000000', label_zh_hk: '2000 萬以上', label_en: 'Over HK$20M', min: 20000000 },
];

export const propertyAreaRangeOptions: PropertyRangeOption[] = [
  { value: 'under_300', label_zh_hk: '300呎以下', label_en: 'Under 300 sqft', max: 300 },
  { value: '300_500', label_zh_hk: '300-500呎', label_en: '300 - 500 sqft', min: 300, max: 500 },
  { value: '500_1000', label_zh_hk: '500-1000呎', label_en: '500 - 1000 sqft', min: 500, max: 1000 },
  { value: '1000_2000', label_zh_hk: '1000-2000呎', label_en: '1,000 - 2,000 sqft', min: 1000, max: 2000 },
  { value: 'over_2000', label_zh_hk: '2000呎以上', label_en: 'Over 2,000 sqft', min: 2000 },
];

export const propertyBedroomOptions: MarketplaceLabelledOption[] = [
  { value: '0', label_zh_hk: '開放式間隔', label_en: 'Studio' },
  { value: '1', label_zh_hk: '1房', label_en: '1 bedroom' },
  { value: '2', label_zh_hk: '2房', label_en: '2 bedrooms' },
  { value: '3', label_zh_hk: '3房', label_en: '3 bedrooms' },
  { value: '4', label_zh_hk: '4房', label_en: '4 bedrooms' },
  { value: '5', label_zh_hk: '5房以上', label_en: '5+ bedrooms' },
];

export const propertyContactMethodOptions: MarketplaceLabelledOption<PropertyContactMethod>[] = [
  { value: 'both', label_zh_hk: '電話及 WhatsApp', label_en: 'Phone and WhatsApp' },
  { value: 'whatsapp', label_zh_hk: 'WhatsApp', label_en: 'WhatsApp' },
  { value: 'phone', label_zh_hk: '電話', label_en: 'Phone' },
  { value: 'chat', label_zh_hk: '站內聊天', label_en: 'In-app chat' },
  { value: 'chat_or_whatsapp', label_zh_hk: '聊天或 WhatsApp', label_en: 'Chat or WhatsApp' },
];

export const propertyBusinessStatusOptions: MarketplaceLabelledOption<PropertyBusinessStatus>[] = [
  { value: 'available', label_zh_hk: '上架中', label_en: 'Available' },
  { value: 'sold', label_zh_hk: '已成交', label_en: 'Sold' },
];

export const propertyLocationScopeOptions: MarketplaceLabelledOption<PropertyLocationScope>[] = [
  { value: 'local', label_zh_hk: '本地', label_en: 'Local' },
  { value: 'overseas', label_zh_hk: '海外', label_en: 'Overseas' },
];

export const propertyListingCategoryOptions: MarketplaceLabelledOption[] = [
  { value: 'standard', label_zh_hk: '一般放盤', label_en: 'Standard listing' },
  { value: 'new_development', label_zh_hk: '新盤', label_en: 'New development' },
  { value: 'developer_project', label_zh_hk: '發展商項目', label_en: 'Developer project' },
  { value: 'multi_unit', label_zh_hk: '多於一伙', label_en: 'Multiple units' },
];

export const propertyFloorZoneOptions: MarketplaceLabelledOption<PropertyFloorZone>[] = [
  { value: 'low', label_zh_hk: '低層', label_en: 'Low floor' },
  { value: 'middle', label_zh_hk: '中層', label_en: 'Middle floor' },
  { value: 'high', label_zh_hk: '高層', label_en: 'High floor' },
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

export const propertyFeatureTagOptions: MarketplaceLabelledOption[] = [
  { value: 'brand_new', label_zh_hk: '全新', label_en: 'Brand new' },
  { value: 'view', label_zh_hk: '有景觀', label_en: 'Open view' },
  { value: 'renovated', label_zh_hk: '有裝修', label_en: 'Renovated' },
  { value: 'appliances', label_zh_hk: '連電器', label_en: 'Appliances included' },
  { value: 'furnished', label_zh_hk: '連傢俬', label_en: 'Furnished' },
  { value: 'exclusive', label_zh_hk: '獨家盤', label_en: 'Exclusive' },
  { value: 'pet_friendly', label_zh_hk: '可養貓狗', label_en: 'Pet friendly' },
  { value: 'parking_included', label_zh_hk: '連車位', label_en: 'Parking included' },
  { value: 'special_unit', label_zh_hk: '單位特色', label_en: 'Special unit' },
  { value: 'commercial_use', label_zh_hk: '工商專用', label_en: 'Commercial use' },
];

export const propertyAutoTagOptions: MarketplaceLabelledOption[] = [
  { value: 'vr_video', label_zh_hk: '有VR或影片', label_en: 'VR or video' },
  { value: 'prepay_discount', label_zh_hk: '預繳全年租金享有折扣優惠', label_en: 'Annual prepay discount' },
];

export const propertyRenovationTagOptions: MarketplaceLabelledOption[] = [
  { value: 'basic_renovation', label_zh_hk: '基本裝修', label_en: 'Basic renovation' },
  { value: 'elegant_renovation', label_zh_hk: '雅緻裝修', label_en: 'Elegant renovation' },
  { value: 'luxury_renovation', label_zh_hk: '豪華裝修', label_en: 'Luxury renovation' },
];

export const propertyFurnitureTagOptions: MarketplaceLabelledOption[] = [
  { value: 'tv_cabinet', label_zh_hk: '電視櫃', label_en: 'TV cabinet' },
  { value: 'wardrobe', label_zh_hk: '衣櫃', label_en: 'Wardrobe' },
  { value: 'table', label_zh_hk: '枱', label_en: 'Table' },
  { value: 'sofa', label_zh_hk: '梳化', label_en: 'Sofa' },
  { value: 'bed', label_zh_hk: '床', label_en: 'Bed' },
  { value: 'chair', label_zh_hk: '椅子', label_en: 'Chair' },
  { value: 'coffee_table', label_zh_hk: '茶几', label_en: 'Coffee table' },
  { value: 'storage_cabinet', label_zh_hk: '儲物櫃', label_en: 'Storage cabinet' },
  { value: 'dressing_table', label_zh_hk: '梳妝台', label_en: 'Dressing table' },
  { value: 'bookshelf', label_zh_hk: '書架', label_en: 'Bookshelf' },
  { value: 'full_furniture', label_zh_hk: '包全屋傢俬', label_en: 'Full furniture' },
  { value: 'partial_furniture', label_zh_hk: '包部份傢俬', label_en: 'Partial furniture' },
];

export const propertyMoreTagOptions: MarketplaceLabelledOption[] = [
  { value: 'living_room_direction', label_zh_hk: '座向(客廳)', label_en: 'Living room direction' },
  { value: 'owner_or_agent', label_zh_hk: '業主或代理', label_en: 'Owner or agent' },
  { value: 'estate_age', label_zh_hk: '屋苑樓齡', label_en: 'Estate age' },
  { value: 'floor_level_tag', label_zh_hk: '樓層', label_en: 'Floor' },
  { value: 'kitchen_type', label_zh_hk: '廚房類型', label_en: 'Kitchen type' },
  { value: 'cooking_mode', label_zh_hk: '廚房煮食模式', label_en: 'Cooking mode' },
  { value: 'developer', label_zh_hk: '發展商', label_en: 'Developer' },
  { value: 'more_options', label_zh_hk: '更多選項', label_en: 'More options' },
];

export const propertyTagGroups: PropertyTagGroup[] = [
  { key: 'features', label_zh_hk: '標籤', label_en: 'Tags', options: propertyFeatureTagOptions },
  { key: 'renovation', label_zh_hk: '裝修', label_en: 'Renovation', options: propertyRenovationTagOptions },
  { key: 'furniture', label_zh_hk: '傢俬', label_en: 'Furniture', options: propertyFurnitureTagOptions },
  { key: 'more', label_zh_hk: '更多', label_en: 'More', options: propertyMoreTagOptions },
];

export const servicedFacilityTagOptions: MarketplaceLabelledOption[] = [
  { value: 'private_kitchen', label_zh_hk: '獨立廚房', label_en: 'Private kitchen' },
  { value: 'front_desk_24h', label_zh_hk: '24小時接待', label_en: '24-hour reception' },
  { value: 'broadband', label_zh_hk: '寬頻上網', label_en: 'Broadband' },
  { value: 'business_center', label_zh_hk: '商業中心', label_en: 'Business centre' },
  { value: 'child_care', label_zh_hk: '幼兒護理', label_en: 'Child care' },
  { value: 'gym', label_zh_hk: '健身室', label_en: 'Gym' },
  { value: 'laundry', label_zh_hk: '洗衣房', label_en: 'Laundry' },
  { value: 'parking', label_zh_hk: '泊車', label_en: 'Parking' },
  { value: 'housekeeping', label_zh_hk: '房務', label_en: 'Housekeeping' },
  { value: 'pay_tv', label_zh_hk: '收費電視', label_en: 'Pay TV' },
  { value: 'pet_friendly', label_zh_hk: '寵物', label_en: 'Pet friendly' },
  { value: 'restaurant', label_zh_hk: '餐廳', label_en: 'Restaurant' },
  { value: 'shuttle_bus', label_zh_hk: '接駁巴士', label_en: 'Shuttle bus' },
  { value: 'pool', label_zh_hk: '泳池', label_en: 'Pool' },
];

export const servicedServiceTagOptions: MarketplaceLabelledOption[] = [
  { value: 'housekeeping', label_zh_hk: '房務清潔', label_en: 'Housekeeping' },
  { value: 'wifi', label_zh_hk: 'Wi-Fi', label_en: 'Wi-Fi' },
  { value: 'utilities', label_zh_hk: '水電煤', label_en: 'Utilities' },
  { value: 'front_desk', label_zh_hk: '前台服務', label_en: 'Front desk' },
  { value: 'linen', label_zh_hk: '床品更換', label_en: 'Linen service' },
];

export const servicedPriceRangeOptions: PropertyRangeOption[] = [
  { value: 'under_5000', label_zh_hk: '5000 以下', label_en: 'Under HK$5,000', max: 5000 },
  { value: '5000_10000', label_zh_hk: '5000-10000', label_en: 'HK$5,000 - HK$10,000', min: 5000, max: 10000 },
  { value: '10000_20000', label_zh_hk: '10000-20000', label_en: 'HK$10,000 - HK$20,000', min: 10000, max: 20000 },
  { value: '20000_50000', label_zh_hk: '20000-50000', label_en: 'HK$20,000 - HK$50,000', min: 20000, max: 50000 },
  { value: 'over_50000', label_zh_hk: '50000 以上', label_en: 'Over HK$50,000', min: 50000 },
];

export const servicedAreaRangeOptions: PropertyRangeOption[] = [
  { value: 'under_100', label_zh_hk: '100呎以下', label_en: 'Under 100 sqft', max: 100 },
  { value: '100_500', label_zh_hk: '100-500呎', label_en: '100 - 500 sqft', min: 100, max: 500 },
  { value: '500_1000', label_zh_hk: '500-1000呎', label_en: '500 - 1,000 sqft', min: 500, max: 1000 },
  { value: '1000_2000', label_zh_hk: '1000-2000呎', label_en: '1,000 - 2,000 sqft', min: 1000, max: 2000 },
  { value: 'over_2000', label_zh_hk: '2000呎以上', label_en: 'Over 2,000 sqft', min: 2000 },
];

// 1. 取得物業選項標籤
export const getPropertyOptionLabel = (
  option: MarketplaceLabelledOption,
  locale: AppLocale,
) => getMarketplaceOptionLabel(option, locale);

// 2. 取得樓盤類型標籤
export const getPropertyTypeLabel = (value: string, locale: AppLocale) => {
  const option = [...propertyTypeOptions, ...legacyPropertyTypeOptions].find((item) => item.value === value);

  return option ? getPropertyOptionLabel(option, locale) : value;
};

// 3. 取得樓盤標籤
export const getPropertyTagLabel = (value: string, locale: AppLocale) => {
  const options = [
    ...propertyTagGroups.flatMap((group) => group.options),
    ...propertyAutoTagOptions,
    ...servicedFacilityTagOptions,
    ...servicedServiceTagOptions,
  ];
  const option = options.find((item) => item.value === value);

  return option ? getPropertyOptionLabel(option, locale) : value;
};

// 4. 取得地區標籤
export const getPropertyDistrictLabel = (value: string, locale: AppLocale) =>
  getMarketplaceDistrictLabel(value, locale);

export { marketplaceDistricts };

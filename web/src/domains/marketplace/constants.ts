/*
 * 二手交易選項常量。
 * 1. 統一分類、地區、價格、成色與聯絡方式 code。
 * 2. 供發布、篩選、Discover 與詳情頁共用。
 */
import type { AppLocale } from '@/app/stores/preferences';

export type MarketplaceCategoryCode =
  | 'home_furniture'
  | 'home_appliance'
  | 'electronics'
  | 'music'
  | 'baby_goods'
  | 'office_furniture'
  | 'home_decor'
  | 'other';

export type MarketplaceRegionCode =
  | 'hong_kong_island'
  | 'kowloon'
  | 'new_territories'
  | 'outlying_islands';

export type MarketplaceDistrictCode =
  | MarketplaceRegionCode
  | 'primary_school_net'
  | 'secondary_school_net'
  | 'tertiary_institution';

export type MarketplaceConditionCode =
  | 'brand_new'
  | 'used_excellent'
  | 'used_good'
  | 'used_fair';

export type MarketplacePriceMode = 'fixed' | 'negotiable' | 'free';

export interface MarketplaceLabelledOption<TValue extends string = string> {
  value: TValue;
  label_zh_hk: string;
  label_en: string;
}

export interface MarketplaceDistrictOption extends MarketplaceLabelledOption<MarketplaceDistrictCode> {
  region?: MarketplaceRegionCode;
}

export interface MarketplaceAreaFilterOption extends MarketplaceLabelledOption {
  filterType: 'all' | 'region' | 'district';
  regionCode?: MarketplaceRegionCode;
  districtCode?: MarketplaceDistrictCode;
}

export const marketplaceCategories: MarketplaceLabelledOption<MarketplaceCategoryCode>[] = [
  { value: 'home_furniture', label_zh_hk: '家居傢俬', label_en: 'Home Furniture' },
  { value: 'home_appliance', label_zh_hk: '家庭電器', label_en: 'Home Appliances' },
  { value: 'electronics', label_zh_hk: '電子產品', label_en: 'Electronics' },
  { value: 'music', label_zh_hk: '樂器', label_en: 'Musical Instruments' },
  { value: 'baby_goods', label_zh_hk: 'BB 用品', label_en: 'Baby Goods' },
  { value: 'office_furniture', label_zh_hk: '辦公室傢俬', label_en: 'Office Furniture' },
  { value: 'home_decor', label_zh_hk: '家居裝飾', label_en: 'Home Decor' },
  { value: 'other', label_zh_hk: '其他', label_en: 'Others' },
];

export const marketplaceRegions: MarketplaceLabelledOption<MarketplaceRegionCode>[] = [
  { value: 'hong_kong_island', label_zh_hk: '香港島', label_en: 'Hong Kong Island' },
  { value: 'kowloon', label_zh_hk: '九龍', label_en: 'Kowloon' },
  { value: 'new_territories', label_zh_hk: '新界', label_en: 'New Territories' },
  { value: 'outlying_islands', label_zh_hk: '離島', label_en: 'Outlying Islands' },
];

export const marketplaceDistricts: MarketplaceDistrictOption[] = [
  { value: 'hong_kong_island', region: 'hong_kong_island', label_zh_hk: '香港島', label_en: 'Hong Kong Island' },
  { value: 'kowloon', region: 'kowloon', label_zh_hk: '九龍', label_en: 'Kowloon' },
  { value: 'new_territories', region: 'new_territories', label_zh_hk: '新界', label_en: 'New Territories' },
  { value: 'outlying_islands', region: 'outlying_islands', label_zh_hk: '離島', label_en: 'Outlying Islands' },
  { value: 'primary_school_net', label_zh_hk: '小學校網', label_en: 'Primary School Net' },
  { value: 'secondary_school_net', label_zh_hk: '中學校網', label_en: 'Secondary School Net' },
  { value: 'tertiary_institution', label_zh_hk: '大學 / 大專院校', label_en: 'University / Tertiary Institution' },
];

export const marketplaceConditions: MarketplaceLabelledOption<MarketplaceConditionCode>[] = [
  { value: 'brand_new', label_zh_hk: '全新', label_en: 'Brand new' },
  { value: 'used_excellent', label_zh_hk: '近乎全新', label_en: 'Excellent' },
  { value: 'used_good', label_zh_hk: '良好', label_en: 'Good' },
  { value: 'used_fair', label_zh_hk: '可用', label_en: 'Fair' },
];

export const marketplacePriceModes: MarketplaceLabelledOption<MarketplacePriceMode>[] = [
  { value: 'fixed', label_zh_hk: '固定價格', label_en: 'Fixed price' },
  { value: 'negotiable', label_zh_hk: '可議價', label_en: 'Negotiable' },
  { value: 'free', label_zh_hk: '免費送出', label_en: 'Free giveaway' },
];

export const getMarketplaceOptionLabel = (
  option: MarketplaceLabelledOption,
  locale: AppLocale,
) => (locale === 'zh-HK' ? option.label_zh_hk : option.label_en);

export const getMarketplaceCategoryLabel = (
  value: string,
  locale: AppLocale,
) =>
  getMarketplaceOptionLabel(
    marketplaceCategories.find((option) => option.value === value) ?? marketplaceCategories[marketplaceCategories.length - 1],
    locale,
  );

export const getMarketplaceDistrictLabel = (
  value: string,
  locale: AppLocale,
) => {
  const option = marketplaceDistricts.find((item) => item.value === value);

  return option ? getMarketplaceOptionLabel(option, locale) : value;
};

export const getMarketplaceConditionLabel = (
  value: string,
  locale: AppLocale,
) => {
  const option = marketplaceConditions.find((item) => item.value === value);

  return option ? getMarketplaceOptionLabel(option, locale) : value;
};

export const buildMarketplaceAreaFilterOptions = (
  locale: AppLocale,
  allLabel: string,
): MarketplaceAreaFilterOption[] => [
  {
    value: 'all',
    label_zh_hk: allLabel,
    label_en: allLabel,
    filterType: 'all' as const,
  },
  ...marketplaceDistricts.map((district) => ({
    ...district,
    filterType: 'district' as const,
    districtCode: district.value,
  })),
].map((option) => ({
  ...option,
  label_zh_hk: locale === 'zh-HK' ? option.label_zh_hk : option.label_en,
}));

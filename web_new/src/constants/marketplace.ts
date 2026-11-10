/*
 * 二手交易選項常量。
 * 1. 統一分類、地區、價格、成色與聯絡方式 code。
 * 2. 供發布、篩選與詳情頁共用。
 */
import type { AppLocale } from '@/stores/preferences';

export type MarketplaceCategoryCode =
  | 'home_furniture'
  | 'home_appliance'
  | 'electronics'
  | 'baby_goods'
  | 'other';

export type MarketplaceRegionCode =
  | 'hong_kong_island'
  | 'kowloon'
  | 'new_territories'
  | 'outlying_islands';

export type MarketplaceDistrictCode =
  | 'central_western'
  | 'wan_chai'
  | 'eastern'
  | 'southern'
  | 'yau_tsim_mong'
  | 'sham_shui_po'
  | 'kowloon_city'
  | 'wong_tai_sin'
  | 'kwun_tong'
  | 'kwai_tsing'
  | 'tsuen_wan'
  | 'tuen_mun'
  | 'yuen_long'
  | 'north'
  | 'tai_po'
  | 'sha_tin'
  | 'sai_kung'
  | 'islands';

export type MarketplaceConditionCode =
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
  region: MarketplaceRegionCode;
}

export interface MarketplaceAreaFilterOption extends MarketplaceLabelledOption {
  filterType: 'all' | 'region' | 'district';
  regionCode?: MarketplaceRegionCode;
  districtCode?: MarketplaceDistrictCode;
}

export const marketplaceCategories: MarketplaceLabelledOption<MarketplaceCategoryCode>[] = [
  { value: 'home_furniture', label_zh_hk: '家居傢俱', label_en: 'Home Furniture' },
  { value: 'home_appliance', label_zh_hk: '家庭電器', label_en: 'Home Appliances' },
  { value: 'electronics', label_zh_hk: '電子產品', label_en: 'Electronics' },
  { value: 'baby_goods', label_zh_hk: 'BB 用品', label_en: 'Baby Goods' },
  { value: 'other', label_zh_hk: '其他', label_en: 'Others' },
];

export const marketplaceRegions: MarketplaceLabelledOption<MarketplaceRegionCode>[] = [
  { value: 'hong_kong_island', label_zh_hk: '香港島', label_en: 'Hong Kong Island' },
  { value: 'kowloon', label_zh_hk: '九龍', label_en: 'Kowloon' },
  { value: 'new_territories', label_zh_hk: '新界', label_en: 'New Territories' },
  { value: 'outlying_islands', label_zh_hk: '離島', label_en: 'Outlying Islands' },
];

export const marketplaceDistricts: MarketplaceDistrictOption[] = [
  { value: 'central_western', region: 'hong_kong_island', label_zh_hk: '中西區', label_en: 'Central and Western' },
  { value: 'wan_chai', region: 'hong_kong_island', label_zh_hk: '灣仔區', label_en: 'Wan Chai' },
  { value: 'eastern', region: 'hong_kong_island', label_zh_hk: '東區', label_en: 'Eastern' },
  { value: 'southern', region: 'hong_kong_island', label_zh_hk: '南區', label_en: 'Southern' },
  { value: 'yau_tsim_mong', region: 'kowloon', label_zh_hk: '油尖旺區', label_en: 'Yau Tsim Mong' },
  { value: 'sham_shui_po', region: 'kowloon', label_zh_hk: '深水埗區', label_en: 'Sham Shui Po' },
  { value: 'kowloon_city', region: 'kowloon', label_zh_hk: '九龍城區', label_en: 'Kowloon City' },
  { value: 'wong_tai_sin', region: 'kowloon', label_zh_hk: '黃大仙區', label_en: 'Wong Tai Sin' },
  { value: 'kwun_tong', region: 'kowloon', label_zh_hk: '觀塘區', label_en: 'Kwun Tong' },
  { value: 'kwai_tsing', region: 'new_territories', label_zh_hk: '葵青區', label_en: 'Kwai Tsing' },
  { value: 'tsuen_wan', region: 'new_territories', label_zh_hk: '荃灣區', label_en: 'Tsuen Wan' },
  { value: 'tuen_mun', region: 'new_territories', label_zh_hk: '屯門區', label_en: 'Tuen Mun' },
  { value: 'yuen_long', region: 'new_territories', label_zh_hk: '元朗區', label_en: 'Yuen Long' },
  { value: 'north', region: 'new_territories', label_zh_hk: '北區', label_en: 'North' },
  { value: 'tai_po', region: 'new_territories', label_zh_hk: '大埔區', label_en: 'Tai Po' },
  { value: 'sha_tin', region: 'new_territories', label_zh_hk: '沙田區', label_en: 'Sha Tin' },
  { value: 'sai_kung', region: 'new_territories', label_zh_hk: '西貢區', label_en: 'Sai Kung' },
  { value: 'islands', region: 'outlying_islands', label_zh_hk: '離島區', label_en: 'Islands' },
];

export const marketplaceConditions: MarketplaceLabelledOption<MarketplaceConditionCode>[] = [
  { value: 'used_excellent', label_zh_hk: '近乎全新', label_en: 'Excellent' },
  { value: 'used_good', label_zh_hk: '良好', label_en: 'Good' },
  { value: 'used_fair', label_zh_hk: '尚可', label_en: 'Fair' },
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
  const option =
    marketplaceRegions.find((item) => item.value === value) ??
    marketplaceDistricts.find((item) => item.value === value);

  return option ? getMarketplaceOptionLabel(option, locale) : value;
};

export const getMarketplaceConditionLabel = (
  value: string,
  locale: AppLocale,
) => {
  const option = marketplaceConditions.find((item) => item.value === value);

  return option ? getMarketplaceOptionLabel(option, locale) : value;
};

export const normalizeMarketplaceRegionCode = (value: string): MarketplaceRegionCode | '' => {
  const region = marketplaceRegions.find((item) => item.value === value);
  if (region) {
    return region.value;
  }

  return marketplaceDistricts.find((item) => item.value === value)?.region ?? '';
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
  ...marketplaceRegions.map((region) => ({
    ...region,
    filterType: 'region' as const,
    regionCode: region.value,
  })),
  ...marketplaceDistricts.map((district) => ({
    ...district,
    filterType: 'district' as const,
    districtCode: district.value,
  })),
].map((option) => ({
  ...option,
  label_zh_hk: locale === 'zh-HK' ? option.label_zh_hk : option.label_en,
}));

export const buildMarketplaceRegionFilterOptions = (
  locale: AppLocale,
  allLabel: string,
): MarketplaceAreaFilterOption[] => [
  {
    value: 'all',
    label_zh_hk: allLabel,
    label_en: allLabel,
    filterType: 'all' as const,
  },
  ...marketplaceRegions.map((region) => ({
    ...region,
    filterType: 'region' as const,
    regionCode: region.value,
  })),
].map((option) => ({
  ...option,
  label_zh_hk: locale === 'zh-HK' ? option.label_zh_hk : option.label_en,
}));

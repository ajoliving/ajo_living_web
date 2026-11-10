/*
 * 物業頻道常量。
 * 1. 統一樓盤放售與服務式住宅表單選項。
 * 2. 復用二手頻道地區碼以保持後端校驗一致。
 * 3. 提供樓盤發布用區域、地區、分區三級地點選項。
 */
import type { AppLocale } from '@/stores/preferences';
import {
  getMarketplaceDistrictLabel,
  getMarketplaceOptionLabel,
  marketplaceDistricts,
  marketplaceRegions,
  type MarketplaceLabelledOption,
} from '@/constants/marketplace';

export type PropertyDistrictCode = string;
export type PropertyLocationAreaCode = 'hong_kong_island' | 'kowloon' | 'new_territories';
export type PropertyTypeCode = 'residential' | 'private_flat' | 'estate' | 'house' | 'office' | 'shop' | 'car_park' | 'industrial' | 'land';
export type PropertyContactMethod = 'phone' | 'whatsapp' | 'chat' | 'both' | 'chat_or_whatsapp';
export type PropertyBusinessStatus = 'available' | 'sold';
export type PropertyTransactionTypeCode = 'sale' | 'rent';
export type PropertyAreaModeCode = 'usable' | 'gross';
export type PropertyFloorZoneCode = 'low' | 'middle' | 'high';
export type PropertyKitchenTypeCode = 'enclosed' | 'open' | 'pantry' | 'none';
export type PropertyCookingModeCode = 'gas' | 'electric' | 'induction' | 'no_cooking';
export type PropertyAdPackageCode = 'basic' | 'featured' | 'premium';
export type PropertyLocationScopeCode = 'local' | 'overseas';
export type ServicedStayUnitCode = 'day' | 'week' | 'month';
export type PropertyDefaultAvatarCode = 'male' | 'female';

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

export interface PropertyLocationDistrictOption extends MarketplaceLabelledOption {
  areaCode: PropertyLocationAreaCode;
}

export interface PropertyLocationSubdistrictOption extends MarketplaceLabelledOption {
  areaCode: PropertyLocationAreaCode;
  districtCode: string;
}

export interface PropertyLocationSelection {
  areaCode: string;
  districtCode: string;
  subdistrictCode: string;
  area?: MarketplaceLabelledOption<PropertyLocationAreaCode>;
  district?: PropertyLocationDistrictOption;
  subdistrict?: PropertyLocationSubdistrictOption;
}

export interface PropertyAdPackageOption extends MarketplaceLabelledOption<PropertyAdPackageCode> {
  weight: number;
  price_hkd: number;
  price_points: number;
  duration_days: number;
}

export interface PropertySaleFieldProfile {
  value: PropertyTypeCode;
  estateLabelKey: string;
  categoryLabelKey: string;
  requiredArea: 'usable' | 'gross' | 'none';
  showAreaMode: boolean;
  showUsableArea: boolean;
  showGrossArea: boolean;
  showBuildingDetails: boolean;
  showUnitFields: boolean;
  showDirection: boolean;
  showRooms: boolean;
  showLivingRoom: boolean;
  showKitchen: boolean;
  showManagementCompany: boolean;
  showFloor: boolean;
  floorRequired: boolean;
  defaultFloorText: string;
  attributeKeys: string[];
  categoryTags: MarketplaceLabelledOption[];
  featureTags: MarketplaceLabelledOption[];
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
  { value: 'hk_east', label_zh_hk: '港島東', label_en: 'Hong Kong Island East' },
  { value: 'hk_west', label_zh_hk: '港島西', label_en: 'Hong Kong Island West' },
  { value: 'hk_south', label_zh_hk: '港島南', label_en: 'Hong Kong Island South' },
  { value: 'hk_central', label_zh_hk: '港島中', label_en: 'Hong Kong Island Central' },
  { value: 'kowloon', label_zh_hk: '九龍', label_en: 'Kowloon' },
  { value: 'hk_kowloon', label_zh_hk: '九龍', label_en: 'Kowloon' },
  { value: 'new_territories', label_zh_hk: '新界', label_en: 'New Territories' },
  { value: 'outlying_islands', label_zh_hk: '離島', label_en: 'Outlying Islands' },
];

export const propertyLocationAreas: MarketplaceLabelledOption<PropertyLocationAreaCode>[] = [
  { value: 'hong_kong_island', label_zh_hk: '港島', label_en: 'Hong Kong Island' },
  { value: 'kowloon', label_zh_hk: '九龍', label_en: 'Kowloon' },
  { value: 'new_territories', label_zh_hk: '新界', label_en: 'New Territories' },
];

export const propertyLocationDistricts: PropertyLocationDistrictOption[] = [
  { value: 'central_western', areaCode: 'hong_kong_island', label_zh_hk: '中西區', label_en: 'Central and Western' },
  { value: 'wan_chai', areaCode: 'hong_kong_island', label_zh_hk: '灣仔', label_en: 'Wan Chai' },
  { value: 'eastern', areaCode: 'hong_kong_island', label_zh_hk: '東區', label_en: 'Eastern' },
  { value: 'southern', areaCode: 'hong_kong_island', label_zh_hk: '南區', label_en: 'Southern' },
  { value: 'yau_tsim_mong', areaCode: 'kowloon', label_zh_hk: '油尖旺', label_en: 'Yau Tsim Mong' },
  { value: 'sham_shui_po', areaCode: 'kowloon', label_zh_hk: '深水埗', label_en: 'Sham Shui Po' },
  { value: 'kowloon_city', areaCode: 'kowloon', label_zh_hk: '九龍城', label_en: 'Kowloon City' },
  { value: 'wong_tai_sin', areaCode: 'kowloon', label_zh_hk: '黃大仙', label_en: 'Wong Tai Sin' },
  { value: 'kwun_tong', areaCode: 'kowloon', label_zh_hk: '觀塘', label_en: 'Kwun Tong' },
  { value: 'kwai_tsing', areaCode: 'new_territories', label_zh_hk: '葵青', label_en: 'Kwai Tsing' },
  { value: 'tsuen_wan', areaCode: 'new_territories', label_zh_hk: '荃灣', label_en: 'Tsuen Wan' },
  { value: 'tuen_mun', areaCode: 'new_territories', label_zh_hk: '屯門', label_en: 'Tuen Mun' },
  { value: 'yuen_long', areaCode: 'new_territories', label_zh_hk: '元朗', label_en: 'Yuen Long' },
  { value: 'north', areaCode: 'new_territories', label_zh_hk: '北區', label_en: 'North' },
  { value: 'tai_po', areaCode: 'new_territories', label_zh_hk: '大埔', label_en: 'Tai Po' },
  { value: 'sha_tin', areaCode: 'new_territories', label_zh_hk: '沙田', label_en: 'Sha Tin' },
  { value: 'sai_kung', areaCode: 'new_territories', label_zh_hk: '西貢', label_en: 'Sai Kung' },
  { value: 'islands', areaCode: 'new_territories', label_zh_hk: '離島', label_en: 'Islands' },
];

export const propertyLocationSubdistricts: PropertyLocationSubdistrictOption[] = [
  { value: 'kennedy_town', areaCode: 'hong_kong_island', districtCode: 'central_western', label_zh_hk: '堅尼地城', label_en: 'Kennedy Town' },
  { value: 'shek_tong_tsui', areaCode: 'hong_kong_island', districtCode: 'central_western', label_zh_hk: '石塘咀', label_en: 'Shek Tong Tsui' },
  { value: 'sai_ying_pun', areaCode: 'hong_kong_island', districtCode: 'central_western', label_zh_hk: '西營盤', label_en: 'Sai Ying Pun' },
  { value: 'sheung_wan', areaCode: 'hong_kong_island', districtCode: 'central_western', label_zh_hk: '上環', label_en: 'Sheung Wan' },
  { value: 'central', areaCode: 'hong_kong_island', districtCode: 'central_western', label_zh_hk: '中環', label_en: 'Central' },
  { value: 'admiralty', areaCode: 'hong_kong_island', districtCode: 'central_western', label_zh_hk: '金鐘', label_en: 'Admiralty' },
  { value: 'mid_levels', areaCode: 'hong_kong_island', districtCode: 'central_western', label_zh_hk: '半山區', label_en: 'Mid-levels' },
  { value: 'peak', areaCode: 'hong_kong_island', districtCode: 'central_western', label_zh_hk: '山頂', label_en: 'Peak' },
  { value: 'wan_chai', areaCode: 'hong_kong_island', districtCode: 'wan_chai', label_zh_hk: '灣仔', label_en: 'Wan Chai' },
  { value: 'causeway_bay', areaCode: 'hong_kong_island', districtCode: 'wan_chai', label_zh_hk: '銅鑼灣', label_en: 'Causeway Bay' },
  { value: 'happy_valley', areaCode: 'hong_kong_island', districtCode: 'wan_chai', label_zh_hk: '跑馬地', label_en: 'Happy Valley' },
  { value: 'tai_hang', areaCode: 'hong_kong_island', districtCode: 'wan_chai', label_zh_hk: '大坑', label_en: 'Tai Hang' },
  { value: 'so_kon_po', areaCode: 'hong_kong_island', districtCode: 'wan_chai', label_zh_hk: '掃桿埔', label_en: 'So Kon Po' },
  { value: 'jardines_lookout', areaCode: 'hong_kong_island', districtCode: 'wan_chai', label_zh_hk: '渣甸山', label_en: "Jardine's Lookout" },
  { value: 'tin_hau', areaCode: 'hong_kong_island', districtCode: 'eastern', label_zh_hk: '天后', label_en: 'Tin Hau' },
  { value: 'braemar_hill', areaCode: 'hong_kong_island', districtCode: 'eastern', label_zh_hk: '寶馬山', label_en: 'Braemar Hill' },
  { value: 'north_point', areaCode: 'hong_kong_island', districtCode: 'eastern', label_zh_hk: '北角', label_en: 'North Point' },
  { value: 'quarry_bay', areaCode: 'hong_kong_island', districtCode: 'eastern', label_zh_hk: '鰂魚涌', label_en: 'Quarry Bay' },
  { value: 'sai_wan_ho', areaCode: 'hong_kong_island', districtCode: 'eastern', label_zh_hk: '西灣', label_en: 'Sai Wan Ho' },
  { value: 'shau_kei_wan', areaCode: 'hong_kong_island', districtCode: 'eastern', label_zh_hk: '筲箕灣', label_en: 'Shau Kei Wan' },
  { value: 'chai_wan', areaCode: 'hong_kong_island', districtCode: 'eastern', label_zh_hk: '柴灣', label_en: 'Chai Wan' },
  { value: 'siu_sai_wan', areaCode: 'hong_kong_island', districtCode: 'eastern', label_zh_hk: '小西灣', label_en: 'Siu Sai Wan' },
  { value: 'pok_fu_lam', areaCode: 'hong_kong_island', districtCode: 'southern', label_zh_hk: '薄扶林', label_en: 'Pok Fu Lam' },
  { value: 'aberdeen', areaCode: 'hong_kong_island', districtCode: 'southern', label_zh_hk: '香港仔', label_en: 'Aberdeen' },
  { value: 'ap_lei_chau', areaCode: 'hong_kong_island', districtCode: 'southern', label_zh_hk: '鴨脷洲', label_en: 'Ap Lei Chau' },
  { value: 'wong_chuk_hang', areaCode: 'hong_kong_island', districtCode: 'southern', label_zh_hk: '黃竹坑', label_en: 'Wong Chuk Hang' },
  { value: 'shouson_hill', areaCode: 'hong_kong_island', districtCode: 'southern', label_zh_hk: '壽臣山', label_en: 'Shouson Hill' },
  { value: 'repulse_bay', areaCode: 'hong_kong_island', districtCode: 'southern', label_zh_hk: '淺水灣', label_en: 'Repulse Bay' },
  { value: 'chung_hom_kok', areaCode: 'hong_kong_island', districtCode: 'southern', label_zh_hk: '舂磡角', label_en: 'Chung Hom Kok' },
  { value: 'stanley', areaCode: 'hong_kong_island', districtCode: 'southern', label_zh_hk: '赤柱', label_en: 'Stanley' },
  { value: 'tai_tam', areaCode: 'hong_kong_island', districtCode: 'southern', label_zh_hk: '大潭', label_en: 'Tai Tam' },
  { value: 'shek_o', areaCode: 'hong_kong_island', districtCode: 'southern', label_zh_hk: '石澳', label_en: 'Shek O' },
  { value: 'tsim_sha_tsui', areaCode: 'kowloon', districtCode: 'yau_tsim_mong', label_zh_hk: '尖沙咀', label_en: 'Tsim Sha Tsui' },
  { value: 'yau_ma_tei', areaCode: 'kowloon', districtCode: 'yau_tsim_mong', label_zh_hk: '油麻地', label_en: 'Yau Ma Tei' },
  { value: 'west_kowloon_reclamation', areaCode: 'kowloon', districtCode: 'yau_tsim_mong', label_zh_hk: '西九龍填海區', label_en: 'West Kowloon Reclamation' },
  { value: 'kings_park', areaCode: 'kowloon', districtCode: 'yau_tsim_mong', label_zh_hk: '京士柏', label_en: "King's Park" },
  { value: 'mong_kok', areaCode: 'kowloon', districtCode: 'yau_tsim_mong', label_zh_hk: '旺角', label_en: 'Mong Kok' },
  { value: 'tai_kok_tsui', areaCode: 'kowloon', districtCode: 'yau_tsim_mong', label_zh_hk: '大角咀', label_en: 'Tai Kok Tsui' },
  { value: 'mei_foo', areaCode: 'kowloon', districtCode: 'sham_shui_po', label_zh_hk: '美孚', label_en: 'Mei Foo' },
  { value: 'lai_chi_kok', areaCode: 'kowloon', districtCode: 'sham_shui_po', label_zh_hk: '荔枝角', label_en: 'Lai Chi Kok' },
  { value: 'cheung_sha_wan', areaCode: 'kowloon', districtCode: 'sham_shui_po', label_zh_hk: '長沙灣', label_en: 'Cheung Sha Wan' },
  { value: 'sham_shui_po', areaCode: 'kowloon', districtCode: 'sham_shui_po', label_zh_hk: '深水埗', label_en: 'Sham Shui Po' },
  { value: 'shek_kip_mei', areaCode: 'kowloon', districtCode: 'sham_shui_po', label_zh_hk: '石硤尾', label_en: 'Shek Kip Mei' },
  { value: 'yau_yat_tsuen', areaCode: 'kowloon', districtCode: 'sham_shui_po', label_zh_hk: '又一村', label_en: 'Yau Yat Tsuen' },
  { value: 'tai_wo_ping', areaCode: 'kowloon', districtCode: 'sham_shui_po', label_zh_hk: '大窩坪', label_en: 'Tai Wo Ping' },
  { value: 'stonecutters_island', areaCode: 'kowloon', districtCode: 'sham_shui_po', label_zh_hk: '昂船洲', label_en: 'Stonecutters Island' },
  { value: 'hung_hom', areaCode: 'kowloon', districtCode: 'kowloon_city', label_zh_hk: '紅磡', label_en: 'Hung Hom' },
  { value: 'to_kwa_wan', areaCode: 'kowloon', districtCode: 'kowloon_city', label_zh_hk: '土瓜灣', label_en: 'To Kwa Wan' },
  { value: 'ma_tau_kok', areaCode: 'kowloon', districtCode: 'kowloon_city', label_zh_hk: '馬頭角', label_en: 'Ma Tau Kok' },
  { value: 'ma_tau_wai', areaCode: 'kowloon', districtCode: 'kowloon_city', label_zh_hk: '馬頭圍', label_en: 'Ma Tau Wai' },
  { value: 'kai_tak', areaCode: 'kowloon', districtCode: 'kowloon_city', label_zh_hk: '啟德', label_en: 'Kai Tak' },
  { value: 'kowloon_city', areaCode: 'kowloon', districtCode: 'kowloon_city', label_zh_hk: '九龍城', label_en: 'Kowloon City' },
  { value: 'ho_man_tin', areaCode: 'kowloon', districtCode: 'kowloon_city', label_zh_hk: '何文田', label_en: 'Ho Man Tin' },
  { value: 'kowloon_tong', areaCode: 'kowloon', districtCode: 'kowloon_city', label_zh_hk: '九龍塘', label_en: 'Kowloon Tong' },
  { value: 'beacon_hill', areaCode: 'kowloon', districtCode: 'kowloon_city', label_zh_hk: '筆架山', label_en: 'Beacon Hill' },
  { value: 'san_po_kong', areaCode: 'kowloon', districtCode: 'wong_tai_sin', label_zh_hk: '新蒲崗', label_en: 'San Po Kong' },
  { value: 'wong_tai_sin', areaCode: 'kowloon', districtCode: 'wong_tai_sin', label_zh_hk: '黃大仙', label_en: 'Wong Tai Sin' },
  { value: 'tung_tau', areaCode: 'kowloon', districtCode: 'wong_tai_sin', label_zh_hk: '東頭', label_en: 'Tung Tau' },
  { value: 'wang_tau_hom', areaCode: 'kowloon', districtCode: 'wong_tai_sin', label_zh_hk: '橫頭磡', label_en: 'Wang Tau Hom' },
  { value: 'lok_fu', areaCode: 'kowloon', districtCode: 'wong_tai_sin', label_zh_hk: '樂富', label_en: 'Lok Fu' },
  { value: 'diamond_hill', areaCode: 'kowloon', districtCode: 'wong_tai_sin', label_zh_hk: '鑽石山', label_en: 'Diamond Hill' },
  { value: 'tsz_wan_shan', areaCode: 'kowloon', districtCode: 'wong_tai_sin', label_zh_hk: '慈雲山', label_en: 'Tsz Wan Shan' },
  { value: 'ngau_chi_wan', areaCode: 'kowloon', districtCode: 'wong_tai_sin', label_zh_hk: '牛池灣', label_en: 'Ngau Chi Wan' },
  { value: 'ping_shek', areaCode: 'kowloon', districtCode: 'kwun_tong', label_zh_hk: '坪石', label_en: 'Ping Shek' },
  { value: 'kowloon_bay', areaCode: 'kowloon', districtCode: 'kwun_tong', label_zh_hk: '九龍灣', label_en: 'Kowloon Bay' },
  { value: 'ngau_tau_kok', areaCode: 'kowloon', districtCode: 'kwun_tong', label_zh_hk: '牛頭角', label_en: 'Ngau Tau Kok' },
  { value: 'jordan_valley', areaCode: 'kowloon', districtCode: 'kwun_tong', label_zh_hk: '佐敦谷', label_en: 'Jordan Valley' },
  { value: 'kwun_tong', areaCode: 'kowloon', districtCode: 'kwun_tong', label_zh_hk: '觀塘', label_en: 'Kwun Tong' },
  { value: 'sau_mau_ping', areaCode: 'kowloon', districtCode: 'kwun_tong', label_zh_hk: '秀茂坪', label_en: 'Sau Mau Ping' },
  { value: 'lam_tin', areaCode: 'kowloon', districtCode: 'kwun_tong', label_zh_hk: '藍田', label_en: 'Lam Tin' },
  { value: 'yau_tong', areaCode: 'kowloon', districtCode: 'kwun_tong', label_zh_hk: '油塘', label_en: 'Yau Tong' },
  { value: 'lei_yue_mun', areaCode: 'kowloon', districtCode: 'kwun_tong', label_zh_hk: '鯉魚門', label_en: 'Lei Yue Mun' },
  { value: 'kwai_chung', areaCode: 'new_territories', districtCode: 'kwai_tsing', label_zh_hk: '葵涌', label_en: 'Kwai Chung' },
  { value: 'tsing_yi', areaCode: 'new_territories', districtCode: 'kwai_tsing', label_zh_hk: '青衣', label_en: 'Tsing Yi' },
  { value: 'tsuen_wan', areaCode: 'new_territories', districtCode: 'tsuen_wan', label_zh_hk: '荃灣', label_en: 'Tsuen Wan' },
  { value: 'lei_muk_shue', areaCode: 'new_territories', districtCode: 'tsuen_wan', label_zh_hk: '梨木樹', label_en: 'Lei Muk Shue' },
  { value: 'ting_kau', areaCode: 'new_territories', districtCode: 'tsuen_wan', label_zh_hk: '汀九', label_en: 'Ting Kau' },
  { value: 'sham_tseng', areaCode: 'new_territories', districtCode: 'tsuen_wan', label_zh_hk: '深井', label_en: 'Sham Tseng' },
  { value: 'tsing_lung_tau', areaCode: 'new_territories', districtCode: 'tsuen_wan', label_zh_hk: '青龍頭', label_en: 'Tsing Lung Tau' },
  { value: 'ma_wan', areaCode: 'new_territories', districtCode: 'tsuen_wan', label_zh_hk: '馬灣', label_en: 'Ma Wan' },
  { value: 'sunny_bay', areaCode: 'new_territories', districtCode: 'tsuen_wan', label_zh_hk: '欣澳', label_en: 'Sunny Bay' },
  { value: 'tai_lam_chung', areaCode: 'new_territories', districtCode: 'tuen_mun', label_zh_hk: '大欖涌', label_en: 'Tai Lam Chung' },
  { value: 'so_kwun_wat', areaCode: 'new_territories', districtCode: 'tuen_mun', label_zh_hk: '掃管笏', label_en: 'So Kwun Wat' },
  { value: 'tuen_mun', areaCode: 'new_territories', districtCode: 'tuen_mun', label_zh_hk: '屯門', label_en: 'Tuen Mun' },
  { value: 'lam_tei', areaCode: 'new_territories', districtCode: 'tuen_mun', label_zh_hk: '藍地', label_en: 'Lam Tei' },
  { value: 'hung_shui_kiu', areaCode: 'new_territories', districtCode: 'yuen_long', label_zh_hk: '洪水橋', label_en: 'Hung Shui Kiu' },
  { value: 'ha_tsuen', areaCode: 'new_territories', districtCode: 'yuen_long', label_zh_hk: '廈村', label_en: 'Ha Tsuen' },
  { value: 'lau_fau_shan', areaCode: 'new_territories', districtCode: 'yuen_long', label_zh_hk: '流浮山', label_en: 'Lau Fau Shan' },
  { value: 'tin_shui_wai', areaCode: 'new_territories', districtCode: 'yuen_long', label_zh_hk: '天水圍', label_en: 'Tin Shui Wai' },
  { value: 'yuen_long', areaCode: 'new_territories', districtCode: 'yuen_long', label_zh_hk: '元朗', label_en: 'Yuen Long' },
  { value: 'san_tin', areaCode: 'new_territories', districtCode: 'yuen_long', label_zh_hk: '新田', label_en: 'San Tin' },
  { value: 'lok_ma_chau', areaCode: 'new_territories', districtCode: 'yuen_long', label_zh_hk: '落馬洲', label_en: 'Lok Ma Chau' },
  { value: 'kam_tin', areaCode: 'new_territories', districtCode: 'yuen_long', label_zh_hk: '錦田', label_en: 'Kam Tin' },
  { value: 'shek_kong', areaCode: 'new_territories', districtCode: 'yuen_long', label_zh_hk: '石崗', label_en: 'Shek Kong' },
  { value: 'pat_heung', areaCode: 'new_territories', districtCode: 'yuen_long', label_zh_hk: '八鄉', label_en: 'Pat Heung' },
  { value: 'fanling', areaCode: 'new_territories', districtCode: 'north', label_zh_hk: '粉嶺', label_en: 'Fanling' },
  { value: 'luen_wo_hui', areaCode: 'new_territories', districtCode: 'north', label_zh_hk: '聯和墟', label_en: 'Luen Wo Hui' },
  { value: 'sheung_shui', areaCode: 'new_territories', districtCode: 'north', label_zh_hk: '上水', label_en: 'Sheung Shui' },
  { value: 'shek_wu_hui', areaCode: 'new_territories', districtCode: 'north', label_zh_hk: '石湖墟', label_en: 'Shek Wu Hui' },
  { value: 'sha_tau_kok', areaCode: 'new_territories', districtCode: 'north', label_zh_hk: '沙頭角', label_en: 'Sha Tau Kok' },
  { value: 'luk_keng', areaCode: 'new_territories', districtCode: 'north', label_zh_hk: '鹿頸', label_en: 'Luk Keng' },
  { value: 'wu_kau_tang', areaCode: 'new_territories', districtCode: 'north', label_zh_hk: '烏蛟騰', label_en: 'Wu Kau Tang' },
  { value: 'tai_po_market', areaCode: 'new_territories', districtCode: 'tai_po', label_zh_hk: '大埔墟', label_en: 'Tai Po Market' },
  { value: 'tai_po', areaCode: 'new_territories', districtCode: 'tai_po', label_zh_hk: '大埔', label_en: 'Tai Po' },
  { value: 'tai_po_kau', areaCode: 'new_territories', districtCode: 'tai_po', label_zh_hk: '大埔滘', label_en: 'Tai Po Kau' },
  { value: 'tai_mei_tuk', areaCode: 'new_territories', districtCode: 'tai_po', label_zh_hk: '大尾篤', label_en: 'Tai Mei Tuk' },
  { value: 'shuen_wan', areaCode: 'new_territories', districtCode: 'tai_po', label_zh_hk: '船灣', label_en: 'Shuen Wan' },
  { value: 'cheung_muk_tau', areaCode: 'new_territories', districtCode: 'tai_po', label_zh_hk: '樟木頭', label_en: 'Cheung Muk Tau' },
  { value: 'kei_ling_ha', areaCode: 'new_territories', districtCode: 'tai_po', label_zh_hk: '企嶺下', label_en: 'Kei Ling Ha' },
  { value: 'tai_wai', areaCode: 'new_territories', districtCode: 'sha_tin', label_zh_hk: '大圍', label_en: 'Tai Wai' },
  { value: 'sha_tin', areaCode: 'new_territories', districtCode: 'sha_tin', label_zh_hk: '沙田', label_en: 'Sha Tin' },
  { value: 'fo_tan', areaCode: 'new_territories', districtCode: 'sha_tin', label_zh_hk: '火炭', label_en: 'Fo Tan' },
  { value: 'ma_liu_shui', areaCode: 'new_territories', districtCode: 'sha_tin', label_zh_hk: '馬料水', label_en: 'Ma Liu Shui' },
  { value: 'wu_kai_sha', areaCode: 'new_territories', districtCode: 'sha_tin', label_zh_hk: '烏溪沙', label_en: 'Wu Kai Sha' },
  { value: 'ma_on_shan', areaCode: 'new_territories', districtCode: 'sha_tin', label_zh_hk: '馬鞍山', label_en: 'Ma On Shan' },
  { value: 'clear_water_bay', areaCode: 'new_territories', districtCode: 'sai_kung', label_zh_hk: '清水灣', label_en: 'Clear Water Bay' },
  { value: 'sai_kung', areaCode: 'new_territories', districtCode: 'sai_kung', label_zh_hk: '西貢', label_en: 'Sai Kung' },
  { value: 'tai_mong_tsai', areaCode: 'new_territories', districtCode: 'sai_kung', label_zh_hk: '大網仔', label_en: 'Tai Mong Tsai' },
  { value: 'tseung_kwan_o', areaCode: 'new_territories', districtCode: 'sai_kung', label_zh_hk: '將軍澳', label_en: 'Tseung Kwan O' },
  { value: 'hang_hau', areaCode: 'new_territories', districtCode: 'sai_kung', label_zh_hk: '坑口', label_en: 'Hang Hau' },
  { value: 'tiu_keng_leng', areaCode: 'new_territories', districtCode: 'sai_kung', label_zh_hk: '調景嶺', label_en: 'Tiu Keng Leng' },
  { value: 'ma_yau_tong', areaCode: 'new_territories', districtCode: 'sai_kung', label_zh_hk: '馬游塘', label_en: 'Ma Yau Tong' },
  { value: 'cheung_chau', areaCode: 'new_territories', districtCode: 'islands', label_zh_hk: '長洲', label_en: 'Cheung Chau' },
  { value: 'peng_chau', areaCode: 'new_territories', districtCode: 'islands', label_zh_hk: '坪洲', label_en: 'Peng Chau' },
  { value: 'lantau_island_including_tung_chung', areaCode: 'new_territories', districtCode: 'islands', label_zh_hk: '大嶼山（包括東涌）', label_en: 'Lantau Island (including Tung Chung)' },
  { value: 'lamma_island', areaCode: 'new_territories', districtCode: 'islands', label_zh_hk: '南丫島', label_en: 'Lamma Island' },
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
  { value: 'industrial', label: '工商' },
  { value: 'office', label: '商廈' },
  { value: 'shop', label: '店鋪' },
  { value: 'land', label: '土地' },
];

export const propertySalePriceRangeFilterOptions: PropertyRangeOption[] = [
  { value: '', label: '不限' },
  { value: 'under_4m', label: '400萬以下', max: 4_000_000 },
  { value: '4m_8m', label: '400-800萬', min: 4_000_000, max: 8_000_000 },
  { value: '8m_12m', label: '800-1200萬', min: 8_000_000, max: 12_000_000 },
  { value: 'over_20m', label: '2000萬+', min: 20_000_000 },
];

export const propertyRentPriceRangeFilterOptions: PropertyRangeOption[] = [
  { value: '', label: '不限' },
  { value: 'under_5000', label: '5000元以下', max: 5_000 },
  { value: '5000_10000', label: '5000-10000元', min: 5_000, max: 10_000 },
  { value: '10000_15000', label: '10000-15000元', min: 10_000, max: 15_000 },
  { value: '15000_20000', label: '15000-20000元', min: 15_000, max: 20_000 },
  { value: '20000_40000', label: '20000-40000元', min: 20_000, max: 40_000 },
  { value: 'over_40000', label: '40000元以上', min: 40_000 },
];

export const propertyAreaModeFilterOptions: PropertyFilterOption[] = [
  { value: 'usable', label: '實用面積' },
  { value: 'gross', label: '建築面積' },
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
  { value: 'simple', label: '簡約裝修' },
  { value: 'special', label: '特色裝修' },
];

export const propertyTagFilterOptions: PropertyFilterOption[] = [
  { value: '', label: '全部' },
  { value: 'brand_new', label: '全新' },
  { value: 'view', label: '有景觀' },
  { value: 'renovated', label: '有裝修' },
  { value: 'furnished', label: '連傢俬' },
  { value: 'exclusive', label: '獨家盤' },
  { value: 'pet_friendly', label: '可養貓狗' },
];

export const propertyPublisherFilterOptions: PropertyFilterOption[] = [
  { value: '', label: '全部' },
  { value: 'owner', label: '業主' },
  { value: 'agent', label: '代理' },
  { value: 'operator', label: '營運商' },
];

export const propertyTypeOptions: MarketplaceLabelledOption<PropertyTypeCode>[] = [
  { value: 'residential', label_zh_hk: '住宅', label_en: 'Residential' },
  { value: 'private_flat', label_zh_hk: '私人住宅', label_en: 'Private flat' },
  { value: 'estate', label_zh_hk: '屋苑單位', label_en: 'Estate flat' },
  { value: 'house', label_zh_hk: '洋房', label_en: 'House' },
  { value: 'office', label_zh_hk: '寫字樓', label_en: 'Office' },
  { value: 'shop', label_zh_hk: '商舖', label_en: 'Shop' },
  { value: 'car_park', label_zh_hk: '車位', label_en: 'Car park' },
  { value: 'industrial', label_zh_hk: '工商', label_en: 'Commercial / Industrial' },
  { value: 'land', label_zh_hk: '土地', label_en: 'Land' },
];

export const propertyListingTypeOptions: MarketplaceLabelledOption<PropertyTypeCode>[] = [
  { value: 'residential', label_zh_hk: '住宅', label_en: 'Residential' },
  { value: 'car_park', label_zh_hk: '車位', label_en: 'Car park' },
  { value: 'industrial', label_zh_hk: '工商', label_en: 'Commercial / Industrial' },
  { value: 'shop', label_zh_hk: '店鋪', label_en: 'Shop' },
  { value: 'land', label_zh_hk: '土地', label_en: 'Land' },
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

export const propertyAccountPackageOptions: MarketplaceLabelledOption[] = [
  { value: 'personal_account', label_zh_hk: '個人帳戶', label_en: 'Personal account' },
  { value: 'joined_package', label_zh_hk: '已加入的套餐', label_en: 'Joined package' },
];

export const propertyFloorZoneOptions: MarketplaceLabelledOption<PropertyFloorZoneCode>[] = [
  { value: 'low', label_zh_hk: '低層', label_en: 'Low floor' },
  { value: 'middle', label_zh_hk: '中層', label_en: 'Middle floor' },
  { value: 'high', label_zh_hk: '高層', label_en: 'High floor' },
];

export const propertyKitchenTypeOptions: MarketplaceLabelledOption<PropertyKitchenTypeCode>[] = [
  { value: 'enclosed', label_zh_hk: '梗廚', label_en: 'Enclosed kitchen' },
  { value: 'open', label_zh_hk: '開放式廚房', label_en: 'Open kitchen' },
  { value: 'none', label_zh_hk: '沒有廚房', label_en: 'No kitchen' },
];

export const propertyCookingModeOptions: MarketplaceLabelledOption<PropertyCookingModeCode>[] = [
  { value: 'gas', label_zh_hk: '明火煮食', label_en: 'Gas cooking' },
];

export const propertyDirectionOptions: MarketplaceLabelledOption[] = [
  { value: 'N/A', label_zh_hk: 'N/A', label_en: 'N/A' },
  { value: '東', label_zh_hk: '東', label_en: 'East' },
  { value: '東南', label_zh_hk: '東南', label_en: 'Southeast' },
  { value: '南', label_zh_hk: '南', label_en: 'South' },
  { value: '西南', label_zh_hk: '西南', label_en: 'Southwest' },
  { value: '西', label_zh_hk: '西', label_en: 'West' },
  { value: '西北', label_zh_hk: '西北', label_en: 'Northwest' },
  { value: '北', label_zh_hk: '北', label_en: 'North' },
  { value: '東北', label_zh_hk: '東北', label_en: 'Northeast' },
];

export const propertyFloorDisplayOptions: MarketplaceLabelledOption[] = [
  { value: 'N/A', label_zh_hk: 'N/A', label_en: 'N/A' },
  { value: '全幢', label_zh_hk: '全幢', label_en: 'Whole building' },
  { value: '高層', label_zh_hk: '高層', label_en: 'High floor' },
  { value: '中層', label_zh_hk: '中層', label_en: 'Middle floor' },
  { value: '低層', label_zh_hk: '低層', label_en: 'Low floor' },
  { value: '地下', label_zh_hk: '地下', label_en: 'Ground floor' },
  { value: '地庫', label_zh_hk: '地庫', label_en: 'Basement' },
];

export const propertyRoomCountOptions: MarketplaceLabelledOption[] = [
  { value: '-1', label_zh_hk: 'N/A', label_en: 'N/A' },
  { value: '0', label_zh_hk: '開放式間隔', label_en: 'Studio' },
  { value: '1', label_zh_hk: '1房', label_en: '1 bedroom' },
  { value: '2', label_zh_hk: '2房', label_en: '2 bedrooms' },
  { value: '3', label_zh_hk: '3房', label_en: '3 bedrooms' },
  { value: '4', label_zh_hk: '4房', label_en: '4 bedrooms' },
  { value: '5', label_zh_hk: '5+房', label_en: '5+ bedrooms' },
];

export const propertyBathroomCountOptions: MarketplaceLabelledOption[] = [
  { value: '0', label_zh_hk: 'N/A', label_en: 'N/A' },
  { value: '1', label_zh_hk: '1浴室', label_en: '1 bathroom' },
  { value: '2', label_zh_hk: '2浴室', label_en: '2 bathrooms' },
  { value: '3', label_zh_hk: '3浴室', label_en: '3 bathrooms' },
  { value: '4', label_zh_hk: '4浴室', label_en: '4 bathrooms' },
  { value: '5', label_zh_hk: '5+浴室', label_en: '5+ bathrooms' },
];

export const propertyToiletCountOptions: MarketplaceLabelledOption[] = [
  { value: '0', label_zh_hk: 'N/A', label_en: 'N/A' },
  { value: '1', label_zh_hk: '1廁所', label_en: '1 toilet' },
  { value: '2', label_zh_hk: '2廁所', label_en: '2 toilets' },
  { value: '3', label_zh_hk: '3廁所', label_en: '3 toilets' },
  { value: '4', label_zh_hk: '4廁所', label_en: '4 toilets' },
  { value: '5', label_zh_hk: '5+廁所', label_en: '5+ toilets' },
];

export const propertyRentIncludedOptions: MarketplaceLabelledOption[] = [
  { value: 'rates_government_rent', label_zh_hk: '差餉地租', label_en: 'Rates and government rent' },
  { value: 'management_fee', label_zh_hk: '管理費', label_en: 'Management fee' },
  { value: 'gas', label_zh_hk: '煤氣', label_en: 'Gas' },
  { value: 'utilities', label_zh_hk: '水電', label_en: 'Water and electricity' },
  { value: 'air_conditioning_fee', label_zh_hk: '冷氣費', label_en: 'Air-conditioning fee' },
];

export const propertyDefaultAvatarOptions: MarketplaceLabelledOption<PropertyDefaultAvatarCode>[] = [
  { value: 'male', label_zh_hk: '男', label_en: 'Male' },
  { value: 'female', label_zh_hk: '女', label_en: 'Female' },
];

export const propertyAdPackageOptions: PropertyAdPackageOption[] = [
  { value: 'basic', label_zh_hk: '普通', label_en: 'Basic', weight: 0, price_hkd: 600, price_points: 600, duration_days: 30 },
  { value: 'featured', label_zh_hk: '置頂', label_en: 'Featured', weight: 1, price_hkd: 800, price_points: 800, duration_days: 30 },
  { value: 'premium', label_zh_hk: '黃金置頂', label_en: 'Premium featured', weight: 2, price_hkd: 1500, price_points: 1500, duration_days: 30 },
];

export const servicedAdPackageOptions: PropertyAdPackageOption[] = [
  { value: 'basic', label_zh_hk: '普通', label_en: 'Basic', weight: 0, price_hkd: 600, price_points: 800, duration_days: 30 },
];

/*
 * 暫不開放：xlsx 服務式住宅保留「置頂」、「黃金置頂」、「即走盤」廣告等級。
 * { value: 'featured', label_zh_hk: '置頂', label_en: 'Featured', weight: 1, price_hkd: 800, price_points: 1000, duration_days: 30 }
 * { value: 'premium', label_zh_hk: '黃金置頂', label_en: 'Premium featured', weight: 2, price_hkd: 1500, price_points: 1700, duration_days: 30 }
 * { value: 'quick_sale', label_zh_hk: '即走盤', label_en: 'Quick move-out', weight: 3, price_hkd: 1200, price_points: 1400, duration_days: 15 }
 */

export const propertyAnnualPrepayOptions: MarketplaceLabelledOption[] = [
  { value: '95_off', label_zh_hk: '年繳 95 折', label_en: '5% annual discount' },
  { value: '90_off', label_zh_hk: '年繳 9 折', label_en: '10% annual discount' },
];

export const servicedStayUnitOptions: MarketplaceLabelledOption<ServicedStayUnitCode>[] = [
  { value: 'day', label_zh_hk: '日', label_en: 'Day' },
  { value: 'week', label_zh_hk: '周', label_en: 'Week' },
  { value: 'month', label_zh_hk: '個月', label_en: 'Month' },
];

export const servicedRoomCategoryOptions: MarketplaceLabelledOption[] = [
  { value: 'studio', label_zh_hk: '開放式', label_en: 'Studio' },
  { value: 'one_bedroom', label_zh_hk: '1房', label_en: '1 bedroom' },
  { value: 'two_bedroom', label_zh_hk: '2房', label_en: '2 bedrooms' },
  { value: 'three_bedroom', label_zh_hk: '3房', label_en: '3 bedrooms' },
  { value: 'four_bedroom_plus', label_zh_hk: '4房+', label_en: '4+ bedrooms' },
];

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

export const residentialCategoryTagOptions: MarketplaceLabelledOption[] = [
  { value: 'residential_private_estate', label_zh_hk: '私人屋苑', label_en: 'Private estate' },
  { value: 'residential_hos', label_zh_hk: '居屋', label_en: 'HOS flat' },
  { value: 'residential_village_house', label_zh_hk: '村屋', label_en: 'Village house' },
  { value: 'residential_detached_house', label_zh_hk: '獨立屋', label_en: 'Detached house' },
  { value: 'residential_public_housing', label_zh_hk: '公屋', label_en: 'Public housing' },
  { value: 'residential_tong_lau', label_zh_hk: '唐樓', label_en: 'Tong lau' },
  { value: 'residential_mansion', label_zh_hk: '洋樓', label_en: 'Mansion' },
  { value: 'residential_single_block', label_zh_hk: '單幢式大廈', label_en: 'Single block building' },
];

export const carParkCategoryTagOptions: MarketplaceLabelledOption[] = [
  { value: 'car_park_lorry', label_zh_hk: '貨車車位', label_en: 'Lorry parking' },
  { value: 'car_park_motorcycle', label_zh_hk: '電單車位', label_en: 'Motorcycle parking' },
  { value: 'car_park_commercial', label_zh_hk: '工商車位', label_en: 'Commercial parking' },
  { value: 'car_park_residential', label_zh_hk: '住宅車位', label_en: 'Residential parking' },
];

export const shopCategoryTagOptions: MarketplaceLabelledOption[] = [
  { value: 'shop_mall', label_zh_hk: '商場鋪位', label_en: 'Mall shop' },
  { value: 'shop_street', label_zh_hk: '地鋪', label_en: 'Street shop' },
  { value: 'shop_upper_floor', label_zh_hk: '樓上鋪', label_en: 'Upper floor shop' },
  { value: 'shop_business_transfer', label_zh_hk: '生意頂讓', label_en: 'Business transfer' },
];

export const landCategoryTagOptions: MarketplaceLabelledOption[] = [
  { value: 'land_private_residential', label_zh_hk: '私人住宅土地', label_en: 'Private residential land' },
  { value: 'land_village_house', label_zh_hk: '村屋土地', label_en: 'Village house land' },
  { value: 'land_farmland', label_zh_hk: '農地', label_en: 'Farmland' },
  { value: 'land_storage', label_zh_hk: '倉地', label_en: 'Storage land' },
  { value: 'land_warehouse', label_zh_hk: '貸倉', label_en: 'Warehouse land' },
  { value: 'land_recreation', label_zh_hk: '康樂地', label_en: 'Recreation land' },
];

export const propertyViewTagOptions: MarketplaceLabelledOption[] = [
  { value: 'view_mountain', label_zh_hk: '望山景', label_en: 'Mountain view' },
  { value: 'view_garden', label_zh_hk: '望園景', label_en: 'Garden view' },
  { value: 'view_open', label_zh_hk: '望開揚景', label_en: 'Open view' },
  { value: 'view_building', label_zh_hk: '望樓景', label_en: 'Building view' },
  { value: 'view_sea', label_zh_hk: '望海景', label_en: 'Sea view' },
  { value: 'view_river', label_zh_hk: '望河景', label_en: 'River view' },
  { value: 'view_pool', label_zh_hk: '望泳池景', label_en: 'Pool view' },
];

export const propertyFitoutTagOptions: MarketplaceLabelledOption[] = [
  { value: 'fitout_basic', label_zh_hk: '基本裝修', label_en: 'Basic fit-out' },
  { value: 'fitout_nice', label_zh_hk: '雅緻裝修', label_en: 'Nice fit-out' },
  { value: 'fitout_luxury', label_zh_hk: '豪華裝修', label_en: 'Luxury fit-out' },
];

export const propertyApplianceTagOptions: MarketplaceLabelledOption[] = [
  { value: 'appliance_full', label_zh_hk: '包全屋家電', label_en: 'Full appliances' },
  { value: 'appliance_tv', label_zh_hk: '電視', label_en: 'TV' },
  { value: 'appliance_microwave', label_zh_hk: '微波爐', label_en: 'Microwave' },
  { value: 'appliance_air_conditioner', label_zh_hk: '冷氣機', label_en: 'Air conditioner' },
  { value: 'appliance_washer', label_zh_hk: '洗衣機', label_en: 'Washing machine' },
  { value: 'appliance_water_heater', label_zh_hk: '熱水爐', label_en: 'Water heater' },
  { value: 'appliance_fridge', label_zh_hk: '雪櫃', label_en: 'Fridge' },
  { value: 'appliance_steam_oven', label_zh_hk: '蒸焗爐', label_en: 'Steam oven' },
  { value: 'appliance_induction', label_zh_hk: '電磁爐', label_en: 'Induction cooker' },
  { value: 'appliance_hood', label_zh_hk: '抽油煙機', label_en: 'Range hood' },
  { value: 'appliance_oven', label_zh_hk: '焗爐', label_en: 'Oven' },
  { value: 'appliance_dishwasher', label_zh_hk: '洗碗碟機', label_en: 'Dishwasher' },
  { value: 'appliance_dryer', label_zh_hk: '乾衣機', label_en: 'Dryer' },
  { value: 'appliance_wine_cellar', label_zh_hk: '酒櫃', label_en: 'Wine cabinet' },
];

export const propertyFurnitureTagOptions: MarketplaceLabelledOption[] = [
  { value: 'furniture_tv_cabinet', label_zh_hk: '電視櫃', label_en: 'TV cabinet' },
  { value: 'furniture_wardrobe', label_zh_hk: '衣櫃', label_en: 'Wardrobe' },
  { value: 'furniture_table', label_zh_hk: '枱', label_en: 'Table' },
  { value: 'furniture_sofa', label_zh_hk: '梳化', label_en: 'Sofa' },
  { value: 'furniture_bed', label_zh_hk: '床', label_en: 'Bed' },
  { value: 'furniture_chair', label_zh_hk: '椅子', label_en: 'Chair' },
  { value: 'furniture_coffee_table', label_zh_hk: '茶几', label_en: 'Coffee table' },
  { value: 'furniture_storage', label_zh_hk: '儲物櫃', label_en: 'Storage cabinet' },
  { value: 'furniture_dressing_table', label_zh_hk: '梳妝台', label_en: 'Dressing table' },
  { value: 'furniture_bookshelf', label_zh_hk: '書架', label_en: 'Bookshelf' },
  { value: 'furniture_full', label_zh_hk: '包全屋傢俬', label_en: 'Full furniture' },
  { value: 'furniture_partial', label_zh_hk: '包部份傢俬', label_en: 'Partial furniture' },
];

export const propertySpecialFeatureTagOptions: MarketplaceLabelledOption[] = [
  { value: 'parking_indoor', label_zh_hk: '室內車位', label_en: 'Indoor parking' },
  { value: 'parking_outdoor', label_zh_hk: '露天車位', label_en: 'Outdoor parking' },
  { value: 'feature_garden', label_zh_hk: '連花園', label_en: 'Garden' },
  { value: 'feature_tenancy', label_zh_hk: '連租約', label_en: 'With tenancy' },
  { value: 'feature_balcony', label_zh_hk: '有露台', label_en: 'Balcony' },
  { value: 'feature_rooftop', label_zh_hk: '連天台', label_en: 'Rooftop' },
  { value: 'feature_flat_roof', label_zh_hk: '連平台', label_en: 'Flat roof' },
  { value: 'feature_ensuite', label_zh_hk: '連套房', label_en: 'Ensuite' },
  { value: 'feature_maid_room', label_zh_hk: '有工人房', label_en: 'Maid room' },
  { value: 'feature_duplex', label_zh_hk: '複式單位', label_en: 'Duplex' },
  { value: 'feature_whole_floor', label_zh_hk: '全層單位', label_en: 'Whole floor' },
  { value: 'feature_clubhouse', label_zh_hk: '有會所', label_en: 'Clubhouse' },
  { value: 'feature_mtr', label_zh_hk: '近地鐵站', label_en: 'Near MTR' },
  { value: 'feature_mall', label_zh_hk: '迎大型商場', label_en: 'Near shopping mall' },
  { value: 'feature_school_net', label_zh_hk: '優質校網', label_en: 'School net' },
  { value: 'feature_student_friendly', label_zh_hk: '歡迎大學生', label_en: 'Student friendly' },
  { value: 'feature_cctv', label_zh_hk: '24小時閉路電視系統', label_en: '24-hour CCTV' },
  { value: 'feature_24h_access', label_zh_hk: '24小時出入', label_en: '24-hour access' },
  { value: 'feature_mailbox', label_zh_hk: '獨立信箱', label_en: 'Private mailbox' },
  { value: 'feature_private_toilet', label_zh_hk: '獨立洗手間', label_en: 'Private toilet' },
  { value: 'feature_independent_ac', label_zh_hk: '獨立冷氣機', label_en: 'Independent air conditioner' },
];

export const residentialSpecialFeatureTagOptions: MarketplaceLabelledOption[] =
  propertySpecialFeatureTagOptions.filter((option) =>
    ![
      'feature_cctv',
      'feature_24h_access',
      'feature_mailbox',
      'feature_private_toilet',
      'feature_independent_ac',
    ].includes(option.value),
  );

export const industrialSpecialFeatureTagOptions: MarketplaceLabelledOption[] =
  propertySpecialFeatureTagOptions.filter((option) =>
    [
      'parking_indoor',
      'parking_outdoor',
      'feature_cctv',
      'feature_24h_access',
      'feature_mailbox',
      'feature_private_toilet',
      'feature_independent_ac',
      'feature_rooftop',
      'feature_flat_roof',
      'feature_whole_floor',
    ].includes(option.value),
  );

export const industrialCategoryTagOptions: MarketplaceLabelledOption[] = [
  { value: 'industrial_building', label_zh_hk: '工商大廈', label_en: 'Industrial building' },
  { value: 'office', label_zh_hk: '寫字樓', label_en: 'Office' },
];

export const propertySaleFieldProfiles: Record<string, PropertySaleFieldProfile> = {
  residential: {
    value: 'residential',
    estateLabelKey: 'property.editor.estateNameField',
    categoryLabelKey: 'property.editor.residentialCategoryField',
    requiredArea: 'usable',
    showAreaMode: true,
    showUsableArea: true,
    showGrossArea: true,
    showBuildingDetails: true,
    showUnitFields: true,
    showDirection: true,
    showRooms: true,
    showLivingRoom: true,
    showKitchen: true,
    showManagementCompany: true,
    showFloor: true,
    floorRequired: false,
    defaultFloorText: 'N/A',
    attributeKeys: ['area_unverified'],
    categoryTags: residentialCategoryTagOptions,
    featureTags: [
      ...propertyViewTagOptions,
      ...propertyFitoutTagOptions,
      ...propertyApplianceTagOptions,
      ...propertyFurnitureTagOptions,
      ...residentialSpecialFeatureTagOptions,
      ...propertyFeatureTagOptions.filter((option) => option.value !== 'commercial_use'),
    ],
  },
  car_park: {
    value: 'car_park',
    estateLabelKey: 'property.editor.estateNameField',
    categoryLabelKey: 'property.editor.carParkCategoryField',
    requiredArea: 'none',
    showAreaMode: false,
    showUsableArea: false,
    showGrossArea: false,
    showBuildingDetails: true,
    showUnitFields: true,
    showDirection: false,
    showRooms: false,
    showLivingRoom: false,
    showKitchen: false,
    showManagementCompany: true,
    showFloor: true,
    floorRequired: false,
    defaultFloorText: 'N/A',
    attributeKeys: [],
    categoryTags: carParkCategoryTagOptions,
    featureTags: [
      { value: 'exclusive', label_zh_hk: '獨家盤', label_en: 'Exclusive' },
    ],
  },
  industrial: {
    value: 'industrial',
    estateLabelKey: 'property.editor.commercialBuildingNameField',
    categoryLabelKey: 'property.editor.industrialCategoryField',
    requiredArea: 'gross',
    showAreaMode: true,
    showUsableArea: true,
    showGrossArea: true,
    showBuildingDetails: true,
    showUnitFields: true,
    showDirection: true,
    showRooms: true,
    showLivingRoom: false,
    showKitchen: false,
    showManagementCompany: true,
    showFloor: true,
    floorRequired: false,
    defaultFloorText: 'N/A',
    attributeKeys: ['area_unverified'],
    categoryTags: industrialCategoryTagOptions,
    featureTags: [
      ...propertyViewTagOptions,
      ...propertyFitoutTagOptions,
      { value: 'parking_included', label_zh_hk: '連車位', label_en: 'Parking included' },
      { value: 'exclusive', label_zh_hk: '獨家盤', label_en: 'Exclusive' },
      { value: 'commercial_use', label_zh_hk: '工商專用', label_en: 'Commercial use' },
      ...industrialSpecialFeatureTagOptions,
    ],
  },
  shop: {
    value: 'shop',
    estateLabelKey: 'property.editor.commercialBuildingNameField',
    categoryLabelKey: 'property.editor.shopCategoryField',
    requiredArea: 'gross',
    showAreaMode: true,
    showUsableArea: true,
    showGrossArea: true,
    showBuildingDetails: true,
    showUnitFields: true,
    showDirection: false,
    showRooms: true,
    showLivingRoom: false,
    showKitchen: false,
    showManagementCompany: true,
    showFloor: true,
    floorRequired: false,
    defaultFloorText: 'N/A',
    attributeKeys: ['area_unverified'],
    categoryTags: shopCategoryTagOptions,
    featureTags: [
      ...propertyFitoutTagOptions,
      { value: 'renovated', label_zh_hk: '有裝修', label_en: 'Renovated' },
      { value: 'exclusive', label_zh_hk: '獨家盤', label_en: 'Exclusive' },
    ],
  },
  land: {
    value: 'land',
    estateLabelKey: 'property.editor.lotNameField',
    categoryLabelKey: 'property.editor.landCategoryField',
    requiredArea: 'gross',
    showAreaMode: true,
    showUsableArea: true,
    showGrossArea: true,
    showBuildingDetails: false,
    showUnitFields: false,
    showDirection: false,
    showRooms: false,
    showLivingRoom: false,
    showKitchen: false,
    showManagementCompany: false,
    showFloor: false,
    floorRequired: false,
    defaultFloorText: 'N/A',
    attributeKeys: ['lot_number', 'area_unverified'],
    categoryTags: landCategoryTagOptions,
    featureTags: [],
  },
};

export const getPropertySaleFieldProfile = (propertyType: string): PropertySaleFieldProfile =>
  propertySaleFieldProfiles[propertyType] ?? propertySaleFieldProfiles.residential;

export const propertyAutoTagOptions: MarketplaceLabelledOption[] = [
  { value: 'vr_video', label_zh_hk: '有 VR 或影片', label_en: 'VR or video' },
  { value: 'prepay_discount', label_zh_hk: '預繳全年租金優惠', label_en: 'Annual prepay discount' },
];

export const servicedFacilityTagOptions: MarketplaceLabelledOption[] = [
  { value: 'private_kitchen', label_zh_hk: '獨立廚房', label_en: 'Private kitchen' },
  { value: 'front_desk_24h', label_zh_hk: '24小時接待', label_en: '24-hour reception' },
  { value: 'broadband', label_zh_hk: '寬頻上網', label_en: 'Broadband' },
  { value: 'business_center', label_zh_hk: '商業中心', label_en: 'Business centre' },
  { value: 'child_care', label_zh_hk: '幼兒護理', label_en: 'Child care' },
  { value: 'gym', label_zh_hk: '健身室', label_en: 'Gym' },
  { value: 'laundry', label_zh_hk: '洗衣房', label_en: 'Laundry' },
  { value: 'workspace', label_zh_hk: '共享工作區', label_en: 'Workspace' },
  { value: 'lounge', label_zh_hk: '住客休息室', label_en: 'Resident lounge' },
  { value: 'parking', label_zh_hk: '泊車', label_en: 'Parking' },
  { value: 'pay_tv', label_zh_hk: '收費電視', label_en: 'Pay TV' },
  { value: 'pet_friendly', label_zh_hk: '可養貓狗', label_en: 'Pet friendly' },
  { value: 'restaurant', label_zh_hk: '餐廳', label_en: 'Restaurant' },
  { value: 'shuttle_bus', label_zh_hk: '接駁巴士', label_en: 'Shuttle bus' },
  { value: 'pool', label_zh_hk: '泳池', label_en: 'Pool' },
];

export const servicedServiceTagOptions: MarketplaceLabelledOption[] = [
  { value: 'housekeeping', label_zh_hk: '房務清潔', label_en: 'Housekeeping' },
  { value: 'wifi', label_zh_hk: '無線網絡', label_en: 'Wi-Fi' },
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
    ...residentialCategoryTagOptions,
    ...carParkCategoryTagOptions,
    ...industrialCategoryTagOptions,
    ...shopCategoryTagOptions,
    ...landCategoryTagOptions,
    ...propertyViewTagOptions,
    ...propertyFitoutTagOptions,
    ...propertyApplianceTagOptions,
    ...propertyFurnitureTagOptions,
    ...propertySpecialFeatureTagOptions,
    ...propertyAutoTagOptions,
    ...servicedFacilityTagOptions,
    ...servicedServiceTagOptions,
  ];
  const option = options.find((item) => item.value === value);

  return option ? getPropertyOptionLabel(option, locale) : value;
};

// 4. 解析樓盤地點選擇
export const resolvePropertyLocationSelection = (value: string): PropertyLocationSelection => {
  const normalizedValue = value.trim();
  const subdistrict = propertyLocationSubdistricts.find((item) => item.value === normalizedValue);
  if (subdistrict) {
    const district = propertyLocationDistricts.find((item) => item.value === subdistrict.districtCode);
    const area = propertyLocationAreas.find((item) => item.value === subdistrict.areaCode);
    return {
      areaCode: subdistrict.areaCode,
      districtCode: subdistrict.districtCode,
      subdistrictCode: subdistrict.value,
      area,
      district,
      subdistrict,
    };
  }

  const district = propertyLocationDistricts.find((item) => item.value === normalizedValue);
  if (district) {
    const area = propertyLocationAreas.find((item) => item.value === district.areaCode);
    return {
      areaCode: district.areaCode,
      districtCode: district.value,
      subdistrictCode: '',
      area,
      district,
    };
  }

  const area = propertyLocationAreas.find((item) => item.value === normalizedValue);
  if (area) {
    return {
      areaCode: area.value,
      districtCode: '',
      subdistrictCode: '',
      area,
    };
  }

  return {
    areaCode: '',
    districtCode: '',
    subdistrictCode: '',
  };
};

// 5. 取得地區標籤
export const getPropertyDistrictLabel = (value: string, locale: AppLocale) => {
  const subdistrictOption = propertyLocationSubdistricts.find((item) => item.value === value);
  if (subdistrictOption) {
    return getPropertyOptionLabel(subdistrictOption, locale);
  }

  const districtOption = propertyLocationDistricts.find((item) => item.value === value);
  if (districtOption) {
    return getPropertyOptionLabel(districtOption, locale);
  }

  const areaOption = propertyLocationAreas.find((item) => item.value === value);
  if (areaOption) {
    return getPropertyOptionLabel(areaOption, locale);
  }

  const regionOption = propertyRegionDisplayOptions.find((item) => item.value === value) ||
    marketplaceRegions.find((item) => item.value === value);
  if (regionOption) {
    return getPropertyOptionLabel(regionOption, locale);
  }

  return getMarketplaceDistrictLabel(value, locale);
};

export { marketplaceDistricts };

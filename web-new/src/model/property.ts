/*
 * 物業頻道型別。
 * 1. 對齊樓盤放售與服務式住宅後端接口。
 * 2. 提供列表、詳情、發布、我的列表與聯絡授權型別。
 */
import type {
  ContactAccessResult,
  ListingContactSummary,
  ListingImageResponse,
  MarketplaceCommunityResponse,
  MarketplaceOwnerResponse,
} from '@/model/marketplace';

// 1. 定義物業頻道代碼
export type PropertyChannel = 'sale' | 'serviced';
export type PropertyTransactionType = 'sale' | 'rent';
export type PropertyAreaMode = 'usable' | 'gross';
export type PropertySortBy =
  | 'latest'
  | 'price_asc'
  | 'price_desc'
  | 'usable_area_desc'
  | 'gross_area_desc'
  | 'usable_unit_price_desc'
  | 'gross_unit_price_desc';

// 2. 定義物業列表查詢參數
export interface PropertyListParams {
  page?: number;
  page_size?: number;
  keyword?: string;
  region_code?: string;
  district_code?: string;
  status?: string;
  transaction_type?: PropertyTransactionType;
  property_type?: string;
  rental_type?: string;
  area_mode?: PropertyAreaMode;
  min_price_hkd?: number;
  max_price_hkd?: number;
  min_area_sqft?: number;
  max_area_sqft?: number;
  bedroom_count?: number;
  feature_tags?: string;
  publisher_identity_type?: string;
  has_media?: boolean;
  is_new?: boolean;
  price_mode?: 'monthly' | 'daily';
  facility_tags?: string;
  service_tags?: string;
  sort_by?: PropertySortBy;
}

// 3. 定義放售樓盤欄位
export interface PropertySalePayload {
  property_no?: string;
  transaction_type: PropertyTransactionType;
  location_scope?: string;
  listing_category?: string;
  multi_unit_project?: boolean;
  rental_type?: string;
  property_type: string;
  property_attributes?: Record<string, string>;
  renovation_type?: string;
  agency_company_name?: string;
  estate_name: string;
  address_text: string;
  address_text_en?: string;
  block_name?: string;
  unit_name?: string;
  show_unit?: boolean;
  public_location_text?: string;
  latitude?: number | null;
  longitude?: number | null;
  asking_price_hkd: number;
  monthly_rent_hkd?: number;
  price_reference_only?: boolean;
  price_negotiable?: boolean;
  annual_prepay_discount?: boolean;
  annual_prepay_option?: string;
  lease_start_date?: string;
  rent_included?: string;
  area_mode?: PropertyAreaMode;
  usable_area_sqft: number;
  gross_area_sqft?: number | null;
  bedroom_count: number;
  living_room_count: number;
  bathroom_count: number;
  floor_level: string;
  floor_raw?: string;
  floor_zone?: string;
  floor_display_range?: string;
  total_floors?: number;
  direction: string;
  building_age: string;
  completion_year?: number;
  building_total_floors?: number;
  management_company?: string;
  kitchen_type?: string;
  cooking_mode?: string;
  management_fee_hkd?: number;
  video_url?: string;
  vr_url?: string;
  title_en?: string;
  description_en?: string;
  private_note?: string;
  ad_package_code?: string;
  ad_weight?: number;
  ad_price_hkd?: number;
  ad_price_points?: number;
  ad_duration_days?: number;
  ad_expires_at?: string | null;
  feature_tags: string[];
  contact_method: string;
  publisher_role_label: string;
  view_count?: number;
  inquiry_count?: number;
}

// 4.1 已通過代理資料的公開安全快照
export interface PropertyAgentSnapshot {
  profile_type: 'individual' | 'company';
  name: string;
  name_zh: string;
  name_en: string;
  license_number: string;
  default_avatar: 'male' | 'female' | 'custom' | '';
  avatar_url: string;
  signature_zh: string;
  signature_en: string;
  company_card_url: string;
  wechat_url: string;
  wechat_qr_url: string;
}

// 4. 定義樓盤地址聯想結果
export interface PropertyAddressSuggestion {
  address_id: string;
  estate_name: string;
  estate_name_en: string;
  display_name: string;
  address_text: string;
  address_text_en: string;
  district_code: string;
  district_label: string;
  region_code: string;
  block_names: string[];
  completion_year: number;
  remark: string;
}

// 5. 定義服務式住宅房型
export interface ServicedApartmentRoomType {
  name: string;
  name_en?: string;
  room_category?: string;
  usable_area_sqft: number;
  usable_area_min_sqft?: number;
  usable_area_max_sqft?: number;
  monthly_rent_hkd: number;
  monthly_rent_min_hkd?: number;
  monthly_rent_max_hkd?: number;
  daily_rent_min_hkd?: number;
  daily_rent_max_hkd?: number;
  rent_unit?: string;
  rent_suffix_plus?: boolean;
  included_fees: boolean;
  included_fee_items?: string[];
  min_lease_months: number;
  min_stay_value?: number;
  min_stay_unit?: string;
  feature_tags: string[];
  image_media_asset_id?: string;
  page_url?: string;
}

// 6. 定義服務式住宅欄位
export interface ServicedApartmentPayload {
  project_name: string;
  project_name_en?: string;
  project_attributes?: Record<string, string>;
  address_text: string;
  address_text_en?: string;
  website_url?: string;
  whatsapp?: string;
  fax?: string;
  description_en?: string;
  service_intro?: string;
  service_intro_en?: string;
  benefits_text?: string;
  benefits_text_en?: string;
  extra_charges_text?: string;
  extra_charges_text_en?: string;
  lowest_monthly_rent_hkd: number;
  highest_monthly_rent_hkd?: number;
  lowest_daily_rent_hkd?: number;
  price_reference_only?: boolean;
  price_negotiable?: boolean;
  min_usable_area_sqft?: number;
  max_usable_area_sqft?: number;
  min_lease_months: number;
  min_stay_value?: number;
  min_stay_unit?: string;
  location_scope?: string;
  listing_category?: string;
  multi_unit_project?: boolean;
  ad_package_code?: string;
  ad_weight?: number;
  ad_price_hkd?: number;
  ad_price_points?: number;
  ad_duration_days?: number;
  ad_expires_at?: string | null;
  facility_tags: string[];
  service_tags: string[];
  room_types: ServicedApartmentRoomType[];
  contact_method: string;
  publisher_role_label: string;
}

// 7. 定義物業列表項
export interface PropertyListingSummaryResponse {
  listing_id: string;
  module: 'property_sale' | 'serviced_apartment';
  title: string;
  summary: string;
  district_code: string;
  publisher_identity_type: string;
  publication_status: string;
  business_status: string;
  expire_at?: string | null;
  published_at?: string | null;
  updated_at: string;
  community?: MarketplaceCommunityResponse | null;
  owner?: MarketplaceOwnerResponse | null;
  cover_image?: ListingImageResponse | null;
  property_sale?: PropertySalePayload | null;
  serviced_apartment?: ServicedApartmentPayload | null;
  is_favorite?: boolean;
  agent_snapshot?: PropertyAgentSnapshot | null;
}

// 8. 定義物業詳情
export interface PropertyListingDetailResponse extends PropertyListingSummaryResponse {
  description: string;
  images: ListingImageResponse[];
  contact_summary: ListingContactSummary;
  points_charged?: number;
  points_balance_after?: number | null;
  points_transaction_id?: string;
}

// 9. 定義共用聯絡輸入
export interface PropertyContactPayload {
  contact_name_zh?: string;
  contact_name_en?: string;
  phone: string;
  phone_2?: string;
  whatsapp: string;
  wechat?: string;
  email: string;
  contact_attributes?: Record<string, string>;
  show_phone: boolean;
  show_whatsapp: boolean;
  show_chat: boolean;
  show_inquiry_form: boolean;
}

// 10. 定義共用圖片輸入
export interface PropertyImagePayload {
  media_asset_id: string;
  sort_order: number;
  is_cover: boolean;
}

// 11. 定義樓盤放售儲存請求
export interface UpsertPropertySalePayload {
  title: string;
  title_en?: string;
  summary: string;
  description: string;
  description_en?: string;
  district_code: string;
  community_id: string;
  property_no?: string;
  transaction_type?: PropertyTransactionType;
  location_scope?: string;
  listing_category?: string;
  multi_unit_project?: boolean;
  property_type: string;
  property_attributes?: Record<string, string>;
  rental_type?: string;
  renovation_type?: string;
  agency_company_name?: string;
  estate_name: string;
  address_text: string;
  address_text_en?: string;
  block_name?: string;
  unit_name?: string;
  show_unit?: boolean;
  latitude?: number | null;
  longitude?: number | null;
  asking_price_hkd: number;
  monthly_rent_hkd?: number;
  price_reference_only?: boolean;
  price_negotiable?: boolean;
  annual_prepay_discount?: boolean;
  annual_prepay_option?: string;
  lease_start_date?: string;
  rent_included?: string;
  area_mode?: PropertyAreaMode;
  usable_area_sqft: number;
  gross_area_sqft?: number;
  bedroom_count: number;
  living_room_count: number;
  bathroom_count: number;
  floor_level: string;
  floor_raw?: string;
  floor_zone?: string;
  total_floors?: number;
  direction: string;
  building_age: string;
  completion_year?: number;
  building_total_floors?: number;
  management_company?: string;
  kitchen_type?: string;
  cooking_mode?: string;
  management_fee_hkd?: number;
  video_url?: string;
  vr_url?: string;
  private_note?: string;
  ad_package_code?: string;
  feature_tags: string[];
  contact_method: string;
  business_status?: 'available' | 'sold';
  images: PropertyImagePayload[];
  contact: PropertyContactPayload;
}

export interface PropertyActionResult {
  listing_id: string;
  is_favorite?: boolean;
}

export interface PropertyAppointmentPayload {
  contact_name: string;
  contact_phone: string;
  preferred_time?: string;
  message?: string;
  appointment_type?: string;
}

export interface PropertyAppointmentResponse {
  appointment_id: string;
  listing_id: string;
  status: string;
}

export interface PropertyReportPayload {
  reason: string;
  message?: string;
}

export interface PropertyReportResponse {
  report_id: string;
  listing_id: string;
  review_status: string;
}

// 12. 定義樓盤內容翻譯請求與回應
export interface PropertyContentTranslationPayload {
  title: string;
  description: string;
}

export interface PropertyContentTranslationResponse {
  title_en: string;
  description_en: string;
}

// 13. 定義服務式住宅儲存請求
export interface UpsertServicedApartmentPayload {
  title: string;
  title_en?: string;
  summary: string;
  description: string;
  description_en?: string;
  district_code: string;
  community_id: string;
  publisher_identity_type: string;
  project_name: string;
  project_name_en?: string;
  project_attributes?: Record<string, string>;
  address_text: string;
  address_text_en?: string;
  website_url?: string;
  whatsapp?: string;
  fax?: string;
  service_intro?: string;
  service_intro_en?: string;
  benefits_text?: string;
  benefits_text_en?: string;
  extra_charges_text?: string;
  extra_charges_text_en?: string;
  lowest_monthly_rent_hkd: number;
  highest_monthly_rent_hkd?: number;
  lowest_daily_rent_hkd?: number;
  price_reference_only?: boolean;
  price_negotiable?: boolean;
  min_usable_area_sqft?: number;
  max_usable_area_sqft?: number;
  min_lease_months: number;
  min_stay_value?: number;
  min_stay_unit?: string;
  location_scope?: string;
  listing_category?: string;
  multi_unit_project?: boolean;
  ad_package_code?: string;
  facility_tags: string[];
  service_tags: string[];
  room_types: ServicedApartmentRoomType[];
  contact_method: string;
  business_status?: 'available' | 'sold';
  images: PropertyImagePayload[];
  contact: PropertyContactPayload;
}

export type {
  ContactAccessResult,
  ListingImageResponse,
  MarketplaceCommunityResponse,
  MarketplaceOwnerResponse,
};

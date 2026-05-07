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

// 2. 定義物業列表查詢參數
export interface PropertyListParams {
  page?: number;
  page_size?: number;
  keyword?: string;
  district_code?: string;
  status?: string;
  min_price_hkd?: number;
  max_price_hkd?: number;
  sort_by?: 'latest' | 'price_asc' | 'price_desc';
}

// 3. 定義放售樓盤欄位
export interface PropertySalePayload {
  property_type: string;
  estate_name: string;
  address_text: string;
  asking_price_hkd: number;
  usable_area_sqft: number;
  gross_area_sqft?: number | null;
  bedroom_count: number;
  living_room_count: number;
  bathroom_count: number;
  floor_level: string;
  direction: string;
  building_age: string;
  feature_tags: string[];
  contact_method: string;
  publisher_role_label: string;
}

// 4. 定義服務式住宅房型
export interface ServicedApartmentRoomType {
  name: string;
  usable_area_sqft: number;
  monthly_rent_hkd: number;
  included_fees: boolean;
  min_lease_months: number;
  feature_tags: string[];
  image_media_asset_id?: string;
}

// 5. 定義服務式住宅欄位
export interface ServicedApartmentPayload {
  project_name: string;
  address_text: string;
  lowest_monthly_rent_hkd: number;
  min_lease_months: number;
  facility_tags: string[];
  service_tags: string[];
  room_types: ServicedApartmentRoomType[];
  contact_method: string;
  publisher_role_label: string;
}

// 6. 定義物業列表項
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
}

// 7. 定義物業詳情
export interface PropertyListingDetailResponse extends PropertyListingSummaryResponse {
  description: string;
  images: ListingImageResponse[];
  contact_summary: ListingContactSummary;
}

// 8. 定義共用聯絡輸入
export interface PropertyContactPayload {
  phone: string;
  whatsapp: string;
  email: string;
  show_phone: boolean;
  show_whatsapp: boolean;
  show_chat: boolean;
  show_inquiry_form: boolean;
}

// 9. 定義共用圖片輸入
export interface PropertyImagePayload {
  media_asset_id: string;
  sort_order: number;
  is_cover: boolean;
}

// 10. 定義樓盤放售儲存請求
export interface UpsertPropertySalePayload {
  title: string;
  summary: string;
  description: string;
  district_code: string;
  community_id: string;
  publisher_identity_type: string;
  property_type: string;
  estate_name: string;
  address_text: string;
  asking_price_hkd: number;
  usable_area_sqft: number;
  gross_area_sqft?: number;
  bedroom_count: number;
  living_room_count: number;
  bathroom_count: number;
  floor_level: string;
  direction: string;
  building_age: string;
  feature_tags: string[];
  contact_method: string;
  business_status?: 'available' | 'sold';
  images: PropertyImagePayload[];
  contact: PropertyContactPayload;
}

// 11. 定義服務式住宅儲存請求
export interface UpsertServicedApartmentPayload {
  title: string;
  summary: string;
  description: string;
  district_code: string;
  community_id: string;
  publisher_identity_type: string;
  project_name: string;
  address_text: string;
  lowest_monthly_rent_hkd: number;
  min_lease_months: number;
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

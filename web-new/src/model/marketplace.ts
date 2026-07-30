/*
 * 二手交易模組型別。
 * 1. 對齊後端二手帖子、聯絡權限與上傳契約。
 * 2. 為列表、詳情、我的發布與發布頁提供穩定型別。
 */
import type { Listing } from '@/model/listing';
import type {
  MarketplaceCategoryCode,
  MarketplaceConditionCode,
  MarketplaceRegionCode,
} from '@/constants/marketplace';

// 1. 定義帖子圖片回應
export interface ListingImageResponse {
  media_asset_id: string;
  url: string;
  sort_order: number;
  is_cover: boolean;
}

// 2. 定義帖子社區回應
export interface MarketplaceCommunityResponse {
  public_id: string;
  community_type: string;
  name_zh: string;
  name_en: string;
  district_code: string;
  address_text: string;
}

// 3. 定義帖子擁有者回應
export interface MarketplaceOwnerResponse {
  user_id: string;
  public_id?: string;
  display_name?: string;
  avatar_url?: string;
  publisher_identity_type?: string;
}

// 4. 定義帖子聯絡摘要
export interface ListingContactSummary {
  show_phone: boolean;
  show_whatsapp: boolean;
  show_chat: boolean;
  show_inquiry_form: boolean;
  contact_attributes?: Record<string, string>;
  editable_contact?: ListingEditableContact;
}

// 4.1 定義僅供發布者或授權管理者編輯時回填的聯絡資料
export interface ListingEditableContact {
  contact_name_zh: string;
  contact_name_en: string;
  phone: string;
  phone_2: string;
  whatsapp: string;
  wechat: string;
  email: string;
}

// 5. 定義帖子列表項回應
export interface SecondhandListingSummaryResponse {
  listing_id: string;
  title: string;
  summary: string;
  district_code: string;
  category_code: MarketplaceCategoryCode;
  price_mode: string;
  price_hkd?: number | null;
  condition_level: MarketplaceConditionCode;
  visibility_scope: 'public' | 'building_only';
  contact_method: string;
  is_free_giveaway: boolean;
  publisher_identity_type: string;
  publication_status: string;
  business_status: string;
  expire_at?: string | null;
  published_at?: string | null;
  updated_at: string;
  is_favorited: boolean;
  cover_image?: ListingImageResponse;
  community?: MarketplaceCommunityResponse | null;
  owner?: MarketplaceOwnerResponse | null;
}

// 6. 定義帖子詳情回應
export interface SecondhandListingDetailResponse extends SecondhandListingSummaryResponse {
  description: string;
  dimension_text: string;
  pickup_region_code: string;
  pickup_location_text: string;
  delivery_tags: string[];
  images: ListingImageResponse[];
  contact_summary: ListingContactSummary;
  points_charged?: number;
  points_balance_after?: number | null;
  points_transaction_id?: string;
}

// 7. 定義聯絡方式授權結果
export interface ContactAccessResult {
  listing_id: string;
  allowed_channels: Record<string, boolean>;
  contact_payload?: Record<string, string>;
}

// 8. 定義收藏狀態結果
export interface FavoriteResult {
  listing_id: string;
  is_favorited: boolean;
}

// 9. 定義帖子列表查詢參數
export interface ListingListParams {
  page?: number;
  page_size?: number;
  keyword?: string;
  category_code?: MarketplaceCategoryCode;
  region_code?: MarketplaceRegionCode;
  district_code?: string;
  condition_level?: MarketplaceConditionCode;
  visibility_scope?: 'public' | 'building_only';
  price_mode?: string;
  min_price_hkd?: number;
  max_price_hkd?: number;
  only_building?: boolean;
  only_free?: boolean;
  exclude_free?: boolean;
  has_media?: boolean;
  sort_by?: 'latest' | 'price_asc' | 'price_desc';
}

// 10. 定義我的帖子查詢參數
export interface MyListingListParams {
  page?: number;
  page_size?: number;
  keyword?: string;
  category_code?: string;
  status?: string;
}

// 11. 定義設定頁全部帖子查詢參數
export interface SettingsListingListParams {
  page?: number;
  page_size?: number;
  keyword?: string;
  category_code?: string;
  status?: string;
}

// 12. 定義帖子建立與更新請求
export interface UpsertSecondhandListingPayload {
  title: string;
  summary: string;
  description: string;
  district_code: string;
  community_id: string;
  publisher_identity_type: string;
  category_code: MarketplaceCategoryCode;
  price_mode: string;
  price_hkd?: number;
  condition_level: MarketplaceConditionCode;
  dimension_text: string;
  pickup_region_code: string;
  pickup_location_text: string;
  delivery_tags: string[];
  visibility_scope: 'public' | 'building_only';
  contact_method: string;
  images: Array<{
    media_asset_id: string;
    sort_order: number;
    is_cover: boolean;
  }>;
  business_status?: 'available' | 'sold';
  contact: {
    phone: string;
    whatsapp: string;
    email: string;
    show_phone: boolean;
    show_whatsapp: boolean;
    show_chat: boolean;
    show_inquiry_form: boolean;
  };
}

// 13. 定義上傳預簽名請求
export interface PresignUploadPayload {
  file_name: string;
  mime_type: string;
  file_size: number;
  purpose?: string;
  object_prefix?: string;
}

// 14. 定義上傳預簽名結果
export interface PresignUploadResult {
  upload_url: string;
  object_key: string;
  upload_token: string;
  headers: Record<string, string>;
}

// 15. 定義媒體資產回應
export interface MediaAssetResponse {
  media_asset_id: string;
  storage_provider: string;
  bucket_name: string;
  object_key: string;
  mime_type: string;
  width?: number | null;
  height?: number | null;
  file_size: number;
  checksum_sha256: string;
  url: string;
  in_use: boolean;
  created_at: string;
}

// 16. 定義前端卡片可接受的帖子資料
export type MarketplaceListingLike =
  | Listing
  | SecondhandListingSummaryResponse
  | SecondhandListingDetailResponse;

/*
 * 二手帖子資料型別。
 * 1. 描述列表、詳情、發布與我的發布所需欄位。
 * 2. 對齊一期可見性、狀態與聯絡能力。
 */
import type { Community } from '@/domains/building/model';
import type { UserSummary } from '@/domains/account/model';

// 1. 定義帖子可見性
export type ListingVisibility = 'public' | 'building_only';

// 2. 定義帖子狀態
export type ListingStatus = 'active' | 'expired' | 'sold';

// 3. 定義商品類型
export type ListingItemType = 'new' | 'used' | 'personal';

// 4. 定義帖子圖片
export interface ListingImage {
  id: string;
  url: string;
  alt: string;
  is_cover: boolean;
}

// 5. 定義聯絡資訊展示結構
export interface ListingContact {
  phone: string;
  whatsapp: string;
  allow_chat: boolean;
  gate_hint: string;
}

// 6. 定義帖子資料主體
export interface Listing {
  id: number;
  slug: string;
  title: string;
  category: string;
  category_group_key: string;
  category_key: string;
  listing_type: ListingItemType;
  condition_label: string;
  price_hkd: number;
  summary: string;
  description: string;
  visibility: ListingVisibility;
  status: ListingStatus;
  published_at: string;
  expires_at: string;
  tags: string[];
  featured: boolean;
  community: Community;
  seller: UserSummary;
  images: ListingImage[];
  contact: ListingContact;
  stats: {
    view_count: number;
    save_count: number;
    chat_count: number;
  };
}

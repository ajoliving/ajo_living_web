/*
 * 會員資料型別。
 * 1. 提供公開賣家資訊與當前會員資訊結構。
 * 2. 定義會員個人資料與首頁導覽所需結構。
 * 3. 與真實登入、會員中心 API 保持一致命名方向。
 */
import type { Community, MetaCommunity } from '@/domains/building/model';
import type { IsmartMessage } from '@/domains/account/model';

// 1. 定義公開會員摘要
export interface UserSummary {
  id: number;
  display_name: string;
  avatar_url: string;
  member_since: string;
  community_id: number;
}

// 2. 定義當前會員資料
export interface UserProfile extends UserSummary {
  phone: string;
  email: string;
  preferred_locale: 'zh-HK' | 'en';
  preferred_theme: 'default' | 'copper-sun' | 'dark-neutral';
  primary_community: Community;
}

// 3. 定義當前會員社區資料
export interface CurrentMemberCommunity extends MetaCommunity {}

// 4. 定義 iSmart 相關物業資料
export interface IsmartRelatedProperty {
  property_name: string;
  status: string;
}

// 5. 定義 iSmart 舊系統帳號資料
export interface IsmartAccountProfile {
  account_code: string;
  account_phone: string;
  account_email: string;
  owner_name_en: string;
  owner_name_zh: string;
  identity_number: string;
  legal_entity: string;
  gender: string;
  birth_date: string;
  contact_name: string;
  contact_phone: string;
  billing_email: string;
  billing_address: string;
  properties: IsmartRelatedProperty[];
}

// 6. 定義當前會員資料
export interface CurrentMemberProfile {
  public_id: string;
  email: string;
  phone_country_code: string;
  phone_number: string;
  member_status: string;
  member_type: string;
  is_staff: boolean;
  role?: string;
  roles?: string[];
  permissions?: string[];
  display_name: string;
  avatar_url: string;
  publisher_identity_type: string;
  district_code: string;
  residence_floor: string;
  residence_unit: string;
  bound_building_ids: string[];
  bound_flat_unit_ids: string[];
  primary_community?: CurrentMemberCommunity;
  profile_completed: boolean;
  ajo_balance?: number;
  ismart_linked: boolean;
  ismart_username: string;
  ismart_bound_phone: string;
  ismart_password: string;
  local_password: string;
  ismart_msg?: IsmartMessage;
  ismart_account_profile?: IsmartAccountProfile;
}

// 7. 定義 Staff 會員列表資料
export interface StaffUserSummary {
  public_id: string;
  email: string;
  phone_country_code: string;
  phone_number: string;
  member_status: string;
  member_type: string;
  is_staff: boolean;
  role?: string;
  roles?: string[];
  permissions?: string[];
  display_name: string;
  publisher_identity_type: string;
  district_code: string;
  residence_floor: string;
  residence_unit: string;
  primary_community?: CurrentMemberCommunity;
  ajo_balance: number;
  created_at: string;
  updated_at: string;
}

// 8. 定義 Staff 新建會員帳戶請求
export interface StaffUserCreatePayload {
  email: string;
  password: string;
  display_name: string;
  phone_country_code: string;
  phone_number: string;
  publisher_identity_type?: string;
  primary_community_id?: string;
  primary_community_name?: string;
  residence_floor?: string;
  residence_unit?: string;
  district_code?: string;
  is_staff?: boolean;
}

// 9. 定義前端導覽用會員資料
export interface SessionUserView {
  public_id: string;
  display_name: string;
  avatar_url: string;
  community_id: number;
  primary_community: {
    public_id: string;
    name: string;
  };
  is_staff: boolean;
  ajo_balance: number;
}

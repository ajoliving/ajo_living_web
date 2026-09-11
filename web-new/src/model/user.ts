/*
 * 會員資料型別。
 * 1. 提供公開賣家資訊與當前會員資訊結構。
 * 2. 定義會員個人資料與首頁導覽所需結構。
 * 3. 與真實登入、會員中心 API 保持一致命名方向。
 */
import type { Community, MetaCommunity } from '@/model/community';
import type { AppThemeName } from '@/utils/theme';

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
  preferred_theme: AppThemeName;
  primary_community: Community;
}

// 3. 定義當前會員社區資料
export interface CurrentMemberCommunity extends MetaCommunity {}

// 4. 定義 iSmart 相關物業資料
export interface IsmartRelatedPropertyProfile {
  property_name: string;
  status: string;
}

// 5. 定義 iSmart 帳號資料
export interface IsmartAccountProfile {
  account_code: string;
  account_phone: string;
  account_email: string;
  owner_name_en: string;
  owner_name_zh: string;
  account_name: string;
  identity_number: string;
  legal_entity: string;
  client_type: string;
  gender: string;
  birth_date: string;
  contact_name: string;
  contact_phone: string;
  contact_email?: string;
  emergency_contact_name: string;
  emergency_contact_phone: string;
  billing_phone: string;
  billing_email: string;
  billing_address: string;
  billing_address_en: string;
  billing_address_zh: string;
  properties: IsmartRelatedPropertyProfile[];
}

// 6. 定義當前會員資料
export interface CurrentMemberProfile {
  public_id: string;
  email: string;
  phone_country_code: string;
  phone_number: string;
  member_status: string;
  account_type?: 'personal' | 'individual_agent' | 'agency_company' | 'agency_company_subaccount';
  member_type: string;
  is_staff: boolean;
  role: string;
  roles: string[];
  permissions: string[];
  display_name: string;
  avatar_url: string;
  publisher_identity_type: string;
  district_code: string;
  primary_community?: CurrentMemberCommunity;
  profile_completed: boolean;
  ajo_balance?: number;
  bound_building_ids?: string[];
  bound_flat_unit_ids?: string[];
  residence_floor?: string;
  residence_unit?: string;
  residence_binding_status?: 'pending' | 'approved' | '';
  ismart_linked?: boolean;
  ismart_username?: string;
  ismart_bound_phone?: string;
  ismart_msg?: {
    user_id?: number;
    username?: string;
    email?: string;
    phone?: string;
    is_staff?: boolean;
    building?: string[];
    staff_building_permissions?: string[];
    client_building_permissions?: string[];
    client_building_flat_units_permissions?: string[];
  };
  ismart_raw?: Record<string, unknown>;
  ismart_account_profile?: IsmartAccountProfile;
}

// 7. 定義 Staff 會員列表資料
export interface StaffUserSummary {
  public_id: string;
  email: string;
  phone_country_code: string;
  phone_number: string;
  member_status: string;
  account_type?: 'personal' | 'individual_agent' | 'agency_company' | 'agency_company_subaccount';
  member_type: string;
  is_staff: boolean;
  role: string;
  roles: string[];
  permissions: string[];
  display_name: string;
  publisher_identity_type: string;
  district_code: string;
  primary_community?: CurrentMemberCommunity;
  created_at: string;
  updated_at: string;
}

// 8. 定義 Staff 新建管理員帳戶請求
export interface StaffUserCreatePayload {
  email: string;
  password: string;
  display_name: string;
  phone_country_code: string;
  phone_number: string;
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
  role: string;
  roles: string[];
  permissions: string[];
  member_type: string;
  is_staff: boolean;
  ajo_balance: number;
}

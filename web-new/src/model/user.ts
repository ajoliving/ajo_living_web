/*
 * 會員資料型別。
 * 1. 提供公開賣家資訊與當前會員資訊結構。
 * 2. 定義會員個人資料與首頁導覽所需結構。
 * 3. 與真實登入、會員中心 API 保持一致命名方向。
 */
import type { Community, MetaCommunity } from '@/model/community';

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
  preferred_theme: 'default' | 'html-fidelity' | 'copper-sun' | 'dark-neutral';
  primary_community: Community;
}

// 3. 定義當前會員社區資料
export interface CurrentMemberCommunity extends MetaCommunity {}

// 4. 定義當前會員資料
export interface CurrentMemberProfile {
  public_id: string;
  email: string;
  phone_country_code: string;
  phone_number: string;
  member_status: string;
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
  ismart_linked?: boolean;
  ismart_username?: string;
  local_password?: string;
  ismart_msg?: {
    username?: string;
    email?: string;
    is_staff?: boolean;
    building?: string[];
    staff_building_permissions?: string[];
    client_building_permissions?: string[];
    client_building_flat_units_permissions?: string[];
  };
}

// 5. 定義 Staff 會員列表資料
export interface StaffUserSummary {
  public_id: string;
  email: string;
  phone_country_code: string;
  phone_number: string;
  member_status: string;
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

// 6. 定義 Staff 角色目錄項目
export interface RoleCatalogItem {
  code: string;
  scope: string;
  name: string;
  description: string;
  permissions: string[];
}

// 7. 定義 Staff 新建會員帳戶請求
export interface StaffUserCreatePayload {
  email: string;
  password: string;
  display_name: string;
  phone_country_code: string;
  phone_number: string;
  member_type: string;
  role_codes: string[];
}

// 8. 定義前端導覽用會員資料
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

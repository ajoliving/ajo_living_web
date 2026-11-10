/*
 * 地產代理資料型別。
 * 1. 定義個人代理、代理公司、審核版本及媒體資料。
 * 2. 定義代理公司子帳戶及 Staff 審核接口契約。
 */

export type AccountType = 'personal' | 'individual_agent' | 'agency_company';
export type AgencyProfileType = 'individual' | 'company';
export type AgencyProfileStatus = 'draft' | 'pending' | 'approved' | 'rejected';
export type AgencyAvatarType = 'male' | 'female' | 'custom';

// 1. 代理媒體資料
export interface AgencyAsset {
  media_asset_id: string;
  url: string;
  mime_type: string;
}

// 2. 代理資料版本
export interface AgencyProfile {
  profile_id: string;
  profile_type: AgencyProfileType;
  name_zh: string;
  name_en: string;
  address_zh: string;
  address_en: string;
  license_number: string;
  is_overseas: boolean;
  is_big_four: boolean;
  phone_1_country_code: string;
  phone_1_number: string;
  phone_1_whatsapp: boolean;
  phone_2_country_code: string;
  phone_2_number: string;
  phone_2_whatsapp: boolean;
  wechat_id: string;
  wechat_url: string;
  signature_zh: string;
  signature_en: string;
  default_avatar: AgencyAvatarType | '';
  avatar_asset?: AgencyAsset | null;
  wechat_qr_asset?: AgencyAsset | null;
  logo_asset?: AgencyAsset | null;
  eaa_license_asset?: AgencyAsset | null;
  business_registration_asset?: AgencyAsset | null;
  company_card_asset?: AgencyAsset | null;
  status: AgencyProfileStatus;
  review_note?: string | null;
  submitted_at?: string | null;
  reviewed_at?: string | null;
  updated_at: string;
  next_editable_at?: string | null;
}

// 3. 會員代理資料回應
export interface CurrentAgencyProfileResponse {
  account_type: AccountType;
  member_status: string;
  active_profile: AgencyProfile | null;
  revision: AgencyProfile | null;
  next_editable_at?: string | null;
}

// 4. 代理資料儲存請求
export interface AgencyProfileUpsertPayload {
  profile_type: AgencyProfileType;
  name_zh: string;
  name_en: string;
  address_zh: string;
  address_en: string;
  license_number: string;
  is_overseas: boolean;
  is_big_four: boolean;
  phone_1_country_code: string;
  phone_1_number: string;
  phone_1_whatsapp: boolean;
  phone_2_country_code: string;
  phone_2_number: string;
  phone_2_whatsapp: boolean;
  wechat_id: string;
  wechat_url: string;
  signature_zh: string;
  signature_en: string;
  default_avatar: AgencyAvatarType | '';
  avatar_asset_id?: string | null;
  wechat_qr_asset_id?: string | null;
  logo_asset_id?: string | null;
  eaa_license_asset_id?: string | null;
  business_registration_asset_id?: string | null;
  company_card_asset_id?: string | null;
}

// 5. 公司子帳戶
export type AgencySubaccountStatus = 'active' | 'disabled';
export type AgencySubaccountPermission = 'property_publish' | 'property_manage';

export interface AgencySubaccount {
  public_id: string;
  user_public_id: string;
  display_name: string;
  phone_country_code: string;
  phone_number: string;
  email: string;
  permissions: AgencySubaccountPermission[];
  status: AgencySubaccountStatus;
  created_at: string;
}

export interface CreateAgencySubaccountPayload {
  display_name: string;
  phone_country_code: string;
  phone_number: string;
  email: string;
  password: string;
  permissions: AgencySubaccountPermission[];
}

// 6. Staff 審核資料
export interface StaffAgencyProfile extends AgencyProfile {
  owner: { display_name?: string; email?: string; public_id?: string };
}

export interface StaffAgencyProfileListParams {
  page?: number;
  page_size?: number;
  keyword?: string;
  profile_type?: AgencyProfileType;
  status?: AgencyProfileStatus | 'all';
}

export interface ReviewAgencyProfilePayload {
  approved: boolean;
  review_note: string;
}

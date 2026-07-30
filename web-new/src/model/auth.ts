/*
 * 認證資料型別。
 * 1. 對齊手機與郵箱 OTP 申請、驗證回應。
 * 2. 保留郵箱密碼、手機密碼、ismart 與住戶註冊流程型別。
 * 3. 對齊登入回應中的 AJO 會員摘要與 POS 權限快照。
 */

// 1. OTP 申請結果
export interface RequestOtpResult {
  expires_in: number;
}

// 2. POS 登入資料快照
export interface IsmartLoginMessage {
  user_id: number;
  username: string;
  email?: string;
  phone?: string;
  is_staff: boolean;
  building: string[];
  staff_building_permissions: string[];
  client_building_permissions: string[];
  client_building_flat_units_permissions: string[];
}

// 3. iSmart 原始業務資料
export type IsmartRawData = Record<string, unknown>;

// 4. 登入回應會員摘要
export interface AuthLoginUser {
  public_id: string;
  member_status: string;
  member_type: string;
  is_staff: boolean;
  role: string;
  roles: string[];
  permissions: string[];
  profile_completed: boolean;
  account_type: 'personal' | 'individual_agent' | 'agency_company' | 'agency_company_subaccount';
  ismart_msg?: IsmartLoginMessage;
  ismart_raw?: IsmartRawData;
}

// 5. OTP 驗證結果
export interface VerifyOtpResult {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  user: AuthLoginUser;
}

// 6. 郵箱驗證碼認證請求
export interface EmailOtpPayload {
  email: string;
  scene?: string;
  code?: string;
  display_name?: string;
}

// 4. 郵箱密碼認證請求
export interface EmailPasswordPayload {
  email: string;
  password: string;
  eng_name?: string;
  display_name?: string;
  phone_country_code?: string;
  phone_number?: string;
  username?: string;
  account_type?: 'personal' | 'individual_agent' | 'agency_company';
  primary_community_id?: string;
  primary_community_name?: string;
  residence_floor?: string;
  residence_unit?: string;
}

// 4.1 電郵密碼註冊請求
export interface RegisterEmailAccountPayload {
  email: string;
  password: string;
  username?: string;
  eng_name: string;
  chi_name?: string;
  phone_country_code: string;
  phone_number: string;
  account_type: 'personal' | 'individual_agent' | 'agency_company';
  id_card?: string;
  remark?: string;
  gender?: 'M' | 'F';
  is_receive_email?: boolean;
  primary_community_id?: string;
  primary_community_name?: string;
  residence_floor?: string;
  residence_unit?: string;
}

// 4.2 註冊資料可用性檢查請求
export interface RegistrationAvailabilityPayload {
	phone_country_code: string;
	phone_number: string;
	email?: string;
}

// 4.3 註冊資料可用性檢查結果
export interface RegistrationAvailabilityResult {
	email_available: boolean;
	phone_available: boolean;
}

// 5. 重設密碼請求
export interface EmailPasswordResetPayload {
  email: string;
  code: string;
  password: string;
}

// 6. 重設密碼結果
export interface EmailPasswordResetResult {
  password_reset: boolean;
}

// 7. 統一帳戶識別登入請求
export interface IdentifierPasswordPayload {
  identifier: string;
  password: string;
}

// 8. 用戶名密碼認證請求
export interface UsernamePasswordPayload {
  username: string;
  password: string;
}

// 9. 手機密碼認證請求
export interface PhonePasswordPayload {
  phone_country_code: string;
  phone_number: string;
  password: string;
}

// 10. ismart 帳戶登入請求
export interface IsmartLoginPayload {
  account: string;
  password: string;
  phone?: string;
  email?: string;
}

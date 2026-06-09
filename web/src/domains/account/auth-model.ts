/*
 * 認證資料型別。
 * 1. 對齊手機與郵箱 OTP 申請、驗證回應。
 * 2. 保留郵箱密碼、手機密碼流程型別供既有接口使用。
 */

// 1. OTP 申請結果
export interface RequestOtpResult {
  expires_in: number;
  mock_code?: string;
}

// 2. OTP 驗證結果
export interface VerifyOtpResult {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  user?: AuthUserResponse;
}

// 3. ismart 帳戶資料
export interface IsmartMessage {
  user_id: number;
  username: string;
  phone?: string;
  password?: string;
  is_staff: boolean;
  building: string[];
  staff_building_permissions: string[];
  client_building_permissions: string[];
  client_building_flat_units_permissions: string[];
}

// 4. 登入後會員摘要
export interface AuthUserResponse {
  public_id: string;
  member_status: string;
  member_type: string;
  is_staff: boolean;
  role: string;
  roles: string[];
  permissions: string[];
  profile_completed: boolean;
  ismart_msg?: IsmartMessage;
}

// 5. 郵箱驗證碼認證請求
export interface EmailOtpPayload {
  email: string;
  scene?: string;
  code?: string;
  display_name?: string;
}

// 6. 郵箱密碼認證請求
export interface EmailPasswordPayload {
  email?: string;
  password: string;
  display_name?: string;
  phone_country_code?: string;
  phone_number?: string;
  username?: string;
  primary_community_id?: string;
  primary_community_name?: string;
  residence_floor?: string;
  residence_unit?: string;
}

// 7. 手機密碼認證請求
export interface PhonePasswordPayload {
  phone_country_code: string;
  phone_number: string;
  password: string;
}

// 8. ismart 登入請求
export interface IsmartLoginPayload {
  account: string;
  password: string;
  phone?: string;
  email?: string;
}

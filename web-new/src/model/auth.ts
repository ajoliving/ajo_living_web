/*
 * 認證資料型別。
 * 1. 對齊手機與郵箱 OTP 申請、驗證回應。
 * 2. 保留郵箱密碼、手機密碼、ismart 與住戶註冊流程型別。
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
}

// 3. 郵箱驗證碼認證請求
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
  display_name?: string;
  phone_country_code?: string;
  phone_number?: string;
  username?: string;
  primary_community_id?: string;
  primary_community_name?: string;
  residence_floor?: string;
  residence_unit?: string;
}

// 5. 手機密碼認證請求
export interface PhonePasswordPayload {
  phone_country_code: string;
  phone_number: string;
  password: string;
}

// 6. ismart 帳戶登入請求
export interface IsmartLoginPayload {
  account: string;
  password: string;
  phone?: string;
  email?: string;
}

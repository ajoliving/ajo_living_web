/*
 * 認證資料型別。
 * 1. 對齊手機與郵箱 OTP 申請、驗證回應。
 * 2. 保留郵箱密碼流程型別供既有接口使用。
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
}

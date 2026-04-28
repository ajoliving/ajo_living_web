/*
 * 認證資料型別。
 * 1. 對齊 OTP 申請與驗證回應。
 * 2. 為真實登入流程提供清晰型別。
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

// 3. 郵箱密碼認證請求
export interface EmailPasswordPayload {
  email: string;
  password: string;
  display_name?: string;
}

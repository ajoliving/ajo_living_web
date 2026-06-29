/*
 * 認證 API。
 * 1. 串接手機 OTP、郵箱 OTP、郵箱密碼、重設密碼、用戶名密碼、手機密碼與登出接口。
 * 2. 與後端真實認證流程保持一致。
 */
import httpClient from '@/httpapis';
import type { ApiResponse } from '@/model/api';
import type { EmailOtpPayload, EmailPasswordPayload, EmailPasswordResetPayload, EmailPasswordResetResult, IsmartLoginPayload, PhonePasswordPayload, RequestOtpResult, UsernamePasswordPayload, VerifyOtpResult } from '@/model/auth';

interface RequestOtpPayload {
  phone_country_code: string;
  phone_number: string;
  scene?: string;
}

interface VerifyOtpPayload {
  phone_country_code: string;
  phone_number: string;
  scene?: string;
  code: string;
}

// 1. 請求 OTP 驗證碼
export const requestOtp = (payload: RequestOtpPayload) =>
  httpClient.post<ApiResponse<RequestOtpResult>>('/auth/otp/request', payload);

// 2. 驗證 OTP 並登入
export const verifyOtp = (payload: VerifyOtpPayload) =>
  httpClient.post<ApiResponse<VerifyOtpResult>>('/auth/otp/verify', payload);

// 3. 請求郵箱驗證碼
export const requestEmailOtp = (payload: EmailOtpPayload) =>
  httpClient.post<ApiResponse<RequestOtpResult>>('/auth/email/otp/request', payload);

// 4. 驗證郵箱驗證碼並登入
export const verifyEmailOtp = (payload: EmailOtpPayload) =>
  httpClient.post<ApiResponse<VerifyOtpResult>>('/auth/email/otp/verify', payload);

// 5. 註冊郵箱密碼帳戶
export const registerWithEmail = (payload: EmailPasswordPayload) =>
  httpClient.post<ApiResponse<VerifyOtpResult>>('/auth/email/register', payload);

// 6. 請求電郵重設密碼驗證碼
export const requestEmailPasswordReset = (payload: Pick<EmailOtpPayload, 'email'>) =>
  httpClient.post<ApiResponse<RequestOtpResult>>('/auth/password/email/request', payload);

// 7. 使用電郵驗證碼重設密碼
export const resetPasswordWithEmail = (payload: EmailPasswordResetPayload) =>
  httpClient.post<ApiResponse<EmailPasswordResetResult>>('/auth/password/email/reset', payload);

// 8. 使用郵箱密碼登入
export const loginWithEmail = (payload: EmailPasswordPayload) =>
  httpClient.post<ApiResponse<VerifyOtpResult>>('/auth/email/login', payload);

// 9. 使用用戶名密碼登入
export const loginWithUsername = (payload: UsernamePasswordPayload) =>
  httpClient.post<ApiResponse<VerifyOtpResult>>('/auth/username/login', payload);

// 10. 使用手機密碼登入
export const loginWithPhone = (payload: PhonePasswordPayload) =>
  httpClient.post<ApiResponse<VerifyOtpResult>>('/auth/phone/login', payload);

// 11. 使用 ismart 帳戶登入
export const loginWithIsmart = (payload: IsmartLoginPayload) =>
  httpClient.post<ApiResponse<VerifyOtpResult>>('/auth/ismart/login', payload);

// 12. 登出目前會員
export const logout = () =>
  httpClient.post<ApiResponse<{ logged_out: boolean }>>('/auth/logout');

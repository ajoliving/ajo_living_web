/*
 * 認證 API。
 * 1. 串接手機 OTP、郵箱 OTP、郵箱密碼、手機密碼與登出接口。
 * 2. 與後端真實認證流程保持一致。
 */
import httpClient from '@/httpapis';
import type { ApiResponse } from '@/model/api';
import type { EmailOtpPayload, EmailPasswordPayload, IsmartLoginPayload, PhonePasswordPayload, RequestOtpResult, VerifyOtpResult } from '@/model/auth';

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

// 6. 使用郵箱密碼登入
export const loginWithEmail = (payload: EmailPasswordPayload) =>
  httpClient.post<ApiResponse<VerifyOtpResult>>('/auth/email/login', payload);

// 7. 使用手機密碼登入
export const loginWithPhone = (payload: PhonePasswordPayload) =>
  httpClient.post<ApiResponse<VerifyOtpResult>>('/auth/phone/login', payload);

// 8. 使用 ismart 帳戶登入
export const loginWithIsmart = (payload: IsmartLoginPayload) =>
  httpClient.post<ApiResponse<VerifyOtpResult>>('/auth/ismart/login', payload);

// 9. 登出目前會員
export const logout = () =>
  httpClient.post<ApiResponse<{ logged_out: boolean }>>('/auth/logout');

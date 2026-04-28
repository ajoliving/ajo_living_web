/*
 * 認證 API。
 * 1. 串接 OTP 申請、驗證與登出接口。
 * 2. 與後端真實認證流程保持一致。
 */
import httpClient from '@/httpapis';
import type { ApiResponse } from '@/model/api';
import type { EmailPasswordPayload, RequestOtpResult, VerifyOtpResult } from '@/model/auth';

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

// 3. 註冊郵箱密碼帳戶
export const registerWithEmail = (payload: EmailPasswordPayload) =>
  httpClient.post<ApiResponse<VerifyOtpResult>>('/auth/email/register', payload);

// 4. 使用郵箱密碼登入
export const loginWithEmail = (payload: EmailPasswordPayload) =>
  httpClient.post<ApiResponse<VerifyOtpResult>>('/auth/email/login', payload);

// 5. 登出目前會員
export const logout = () =>
  httpClient.post<ApiResponse<{ logged_out: boolean }>>('/auth/logout');

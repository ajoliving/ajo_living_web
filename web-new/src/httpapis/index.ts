/*
 * Axios 實例設定。
 * 1. 統一處理 baseURL、Authorization 與 401 響應。
 * 2. 供後續真實 API 接入時直接復用。
 */
import axios from 'axios';

import {
  clearStoredTokens,
  emitAuthSessionExpired,
  readStoredAccessToken,
} from '@/httpapis/auth-session';

const BASE_URL = import.meta.env.VITE_API_BASE_URL ?? '/api/v1';

// 1. 建立 HTTP 用戶端實例
const httpClient = axios.create({
  baseURL: BASE_URL,
  timeout: 10000,
});

// 2. 判斷是否為 POS/iSmart 業務授權失效
export const isIntegrationBusinessAuthError = (error: unknown): boolean => {
  if (!axios.isAxiosError(error)) {
    return false;
  }
  const url = String(error.config?.url ?? '');
  const data = error.response?.data as { code?: string; message?: string } | undefined;
  const message = String(data?.message ?? '').trim().toLowerCase();
  return (
    (
      url.includes('/me/payments/pos/')
      || url.includes('/me/pos/')
      || url.includes('/me/ismart/')
      || url.includes('/me/security/icctv/')
    ) &&
    data?.code === 'AUTH_REQUIRED' &&
    (message === 'ismart login is required' || message === 'pos token expired')
  );
};

// 3. 統一注入登入 token
httpClient.interceptors.request.use((config) => {
  const token = readStoredAccessToken();

  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }

  return config;
});

// 4. 統一處理 401 響應
httpClient.interceptors.response.use(
  (response) => response,
  async (error) => {
    if (error.response?.status === 401 && !isIntegrationBusinessAuthError(error)) {
      clearStoredTokens();
      emitAuthSessionExpired();
    }

    return Promise.reject(error);
  },
);

export default httpClient;

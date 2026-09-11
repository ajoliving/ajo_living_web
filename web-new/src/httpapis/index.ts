/*
 * Axios 實例設定。
 * 1. 統一處理 baseURL、Authorization 與 401 響應。
 * 2. 供後續真實 API 接入時直接復用。
 */
import axios, { type InternalAxiosRequestConfig } from 'axios';

import {
  clearStoredTokens,
  emitAuthSessionExpired,
  readStoredAccessToken,
  readStoredRefreshToken,
  writeStoredTokens,
} from '@/httpapis/auth-session';

const BASE_URL = import.meta.env.VITE_API_BASE_URL ?? '/api/v1';

// 1. 建立 HTTP 用戶端實例
const httpClient = axios.create({
  baseURL: BASE_URL,
  timeout: 10000,
});

interface SessionRefreshPayload {
  access_token?: string;
  refresh_token?: string;
}

interface RetriableAxiosRequestConfig extends InternalAxiosRequestConfig {
  _sessionRefreshRetried?: boolean;
}

let refreshingSession: Promise<string> | null = null;

// 1.1 Refreshes one expired access token for all concurrent browser requests.
const refreshSession = async (): Promise<string> => {
  if (refreshingSession) {
    return refreshingSession;
  }

  const refreshToken = readStoredRefreshToken();
  if (!refreshToken) {
    return Promise.reject(new Error('refresh token is unavailable'));
  }

  refreshingSession = axios.post<{ data?: SessionRefreshPayload }>(`${BASE_URL}/auth/refresh`, {
    refresh_token: refreshToken,
  })
    .then((response) => {
      const payload = response.data.data;
      const accessToken = payload?.access_token?.trim() || '';
      const nextRefreshToken = payload?.refresh_token?.trim() || '';
      if (!accessToken || !nextRefreshToken) {
        throw new Error('session refresh response is invalid');
      }
      writeStoredTokens(accessToken, nextRefreshToken);
      return accessToken;
    })
    .finally(() => {
      refreshingSession = null;
    });

  return refreshingSession;
};

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
    const request = error.config as RetriableAxiosRequestConfig | undefined;
    const isRefreshRequest = request?.url?.includes('/auth/refresh');
    if (error.response?.status === 401 && !isIntegrationBusinessAuthError(error)) {
      if (request && !request._sessionRefreshRetried && !isRefreshRequest) {
        request._sessionRefreshRetried = true;
        try {
          const accessToken = await refreshSession();
          request.headers.Authorization = `Bearer ${accessToken}`;
          return httpClient.request(request);
        } catch {
          // The terminal session-expiry handling below clears stale local state.
        }
      }
      clearStoredTokens();
      emitAuthSessionExpired();
    }

    return Promise.reject(error);
  },
);

export default httpClient;

/*
 * 認證狀態工具。
 * 1. 統一管理登入 token 的本地儲存鍵。
 * 2. 在 HTTP 層與應用入口之間廣播登入失效事件。
 */

export const ACCESS_TOKEN_STORAGE_KEY = 'ajoliving.access-token';
export const REFRESH_TOKEN_STORAGE_KEY = 'ajoliving.refresh-token';
export const AUTH_SESSION_EXPIRED_EVENT = 'ajoliving:auth-session-expired';

// 1. 讀取本地儲存值
const getStoredValue = (key: string): string => {
  if (typeof window === 'undefined') {
    return '';
  }

  return window.localStorage.getItem(key) ?? '';
};

// 2. 讀取 access token
export const readStoredAccessToken = (): string => getStoredValue(ACCESS_TOKEN_STORAGE_KEY);

// 3. 讀取 refresh token
export const readStoredRefreshToken = (): string => getStoredValue(REFRESH_TOKEN_STORAGE_KEY);

// 4. 寫入登入 token
export const writeStoredTokens = (accessToken: string, refreshToken: string): void => {
  if (typeof window === 'undefined') {
    return;
  }

  window.localStorage.setItem(ACCESS_TOKEN_STORAGE_KEY, accessToken);
  window.localStorage.setItem(REFRESH_TOKEN_STORAGE_KEY, refreshToken);
};

// 5. 清除登入 token
export const clearStoredTokens = (): void => {
  if (typeof window === 'undefined') {
    return;
  }

  window.localStorage.removeItem(ACCESS_TOKEN_STORAGE_KEY);
  window.localStorage.removeItem(REFRESH_TOKEN_STORAGE_KEY);
};

// 6. isAccessTokenExpired safely reads an optional JWT exp claim without trusting token contents.
export const isAccessTokenExpired = (accessToken: string): boolean => {
  const payload = accessToken.trim().split('.')[1];
  if (!payload || typeof window === 'undefined') {
    return false;
  }

  try {
    const normalizedPayload = payload.replace(/-/g, '+').replace(/_/g, '/');
    const decodedPayload = window.atob(normalizedPayload.padEnd(Math.ceil(normalizedPayload.length / 4) * 4, '='));
    const decodedBytes = Uint8Array.from(decodedPayload, (character) => character.charCodeAt(0));
    const parsedPayload = JSON.parse(new TextDecoder().decode(decodedBytes)) as { exp?: unknown };
    const expiresAt = Number(parsedPayload.exp);
    return Number.isFinite(expiresAt) && expiresAt > 0 && Date.now() >= expiresAt * 1000;
  } catch {
    return false;
  }
};

// 7. 廣播登入失效事件
export const emitAuthSessionExpired = (): void => {
  if (typeof window === 'undefined') {
    return;
  }

  window.dispatchEvent(new Event(AUTH_SESSION_EXPIRED_EVENT));
};

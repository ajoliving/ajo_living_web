/*
 * 會員登入狀態。
 * 1. 保存真實登入 token 與當前會員資料。
 * 2. 提供 OTP 登入、當前會員查詢與登出流程。
 */
import { defineStore } from 'pinia';

import { loginWithEmail, logout, registerWithEmail, requestOtp, verifyOtp } from '@/httpapis/auth';
import { fetchMe } from '@/httpapis/me';
import type { RequestOtpResult, VerifyOtpResult } from '@/model/auth';
import type { CurrentMemberProfile, SessionUserView } from '@/model/user';

const ACCESS_TOKEN_STORAGE_KEY = 'ajoliving.access-token';
const REFRESH_TOKEN_STORAGE_KEY = 'ajoliving.refresh-token';

const getStoredValue = (key: string) => {
  if (typeof window === 'undefined') {
    return '';
  }

  return window.localStorage.getItem(key) ?? '';
};

const buildSessionUser = (member: CurrentMemberProfile | null): SessionUserView => ({
  public_id: member?.public_id ?? '',
  display_name: member?.display_name?.trim() || 'Guest Member',
  avatar_url: member?.avatar_url ?? '',
  community_id: 0,
  primary_community: {
    public_id: member?.primary_community?.public_id ?? '',
    name:
      member?.primary_community?.name_zh?.trim() ||
      member?.primary_community?.name_en?.trim() ||
      member?.primary_community?.address_text?.trim() ||
      'Community not set',
  },
  role: member?.role ?? 'guest',
  roles: member?.roles ?? [],
  permissions: member?.permissions ?? [],
  member_type: member?.member_type ?? '',
  is_staff: member?.is_staff ?? false,
});

// 1. 建立會員登入狀態 Store
export const useSessionStore = defineStore('session', {
  state: () => ({
    me: null as CurrentMemberProfile | null,
    accessToken: getStoredValue(ACCESS_TOKEN_STORAGE_KEY),
    refreshToken: getStoredValue(REFRESH_TOKEN_STORAGE_KEY),
    isHydrating: false,
    isLoaded: false,
  }),
  getters: {
    // 2. 判斷是否已登入
    isAuthenticated: (state) => state.accessToken.trim().length > 0,
    // 3. 輸出導覽層可直接使用的會員資料
    currentUser: (state) => buildSessionUser(state.me),
  },
  actions: {
    // 4. 保存 token
    setTokens(accessToken: string, refreshToken: string) {
      this.accessToken = accessToken;
      this.refreshToken = refreshToken;

      if (typeof window !== 'undefined') {
        window.localStorage.setItem(ACCESS_TOKEN_STORAGE_KEY, accessToken);
        window.localStorage.setItem(REFRESH_TOKEN_STORAGE_KEY, refreshToken);
      }
    },

    // 5. 清除登入狀態
    clearSession() {
      this.me = null;
      this.accessToken = '';
      this.refreshToken = '';

      if (typeof window !== 'undefined') {
        window.localStorage.removeItem(ACCESS_TOKEN_STORAGE_KEY);
        window.localStorage.removeItem(REFRESH_TOKEN_STORAGE_KEY);
      }
    },

    // 6. 啟動時還原登入狀態
    async hydrateSession() {
      if (this.isLoaded || this.isHydrating) {
        return;
      }

      this.isHydrating = true;

      try {
        if (this.accessToken) {
          await this.loadCurrentUser();
        }
      } catch {
        this.clearSession();
      } finally {
        this.isHydrating = false;
        this.isLoaded = true;
      }
    },

    // 7. 請求 OTP 驗證碼
    async sendOtp(phoneCountryCode: string, phoneNumber: string): Promise<RequestOtpResult> {
      const { data } = await requestOtp({
        phone_country_code: phoneCountryCode,
        phone_number: phoneNumber,
        scene: 'login',
      });

      return data.data;
    },

    // 8. 驗證 OTP 並登入
    async signInWithOtp(phoneCountryCode: string, phoneNumber: string, code: string): Promise<VerifyOtpResult> {
      const { data } = await verifyOtp({
        phone_country_code: phoneCountryCode,
        phone_number: phoneNumber,
        scene: 'login',
        code,
      });

      this.setTokens(data.data.access_token, data.data.refresh_token);

      try {
        await this.loadCurrentUser();
      } catch (error) {
        this.clearSession();
        throw error;
      }

      return data.data;
    },

    // 9. 使用郵箱密碼登入
    async signInWithEmail(email: string, password: string): Promise<VerifyOtpResult> {
      const { data } = await loginWithEmail({ email, password });
      this.setTokens(data.data.access_token, data.data.refresh_token);

      try {
        await this.loadCurrentUser();
      } catch (error) {
        this.clearSession();
        throw error;
      }

      return data.data;
    },

    // 10. 註冊郵箱密碼帳戶
    async registerEmailAccount(email: string, password: string, displayName: string): Promise<VerifyOtpResult> {
      const { data } = await registerWithEmail({
        email,
        password,
        display_name: displayName,
      });
      this.setTokens(data.data.access_token, data.data.refresh_token);

      try {
        await this.loadCurrentUser();
      } catch (error) {
        this.clearSession();
        throw error;
      }

      return data.data;
    },

    // 11. 讀取目前會員資料
    async loadCurrentUser() {
      const { data } = await fetchMe();
      this.me = data.data;
      return data.data;
    },

    // 12. 執行登出
    async signOut() {
      try {
        if (this.accessToken) {
          await logout();
        }
      } finally {
        this.clearSession();
      }
    },
  },
});

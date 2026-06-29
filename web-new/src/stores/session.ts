/*
 * 會員登入狀態。
 * 1. 保存真實登入 token 與當前會員資料。
 * 2. 提供手機 OTP、郵箱 OTP、郵箱密碼、重設密碼、用戶名密碼、手機密碼、ismart、當前會員查詢與登出流程。
 */
import { defineStore } from 'pinia';

import {
  loginWithEmail,
  loginWithIsmart,
  loginWithPhone,
  loginWithUsername,
  logout,
  registerWithEmail,
  requestEmailPasswordReset,
  requestEmailOtp,
  requestOtp,
  resetPasswordWithEmail,
  verifyEmailOtp,
  verifyOtp,
} from '@/httpapis/auth';
import {
  clearStoredTokens,
  readStoredAccessToken,
  readStoredRefreshToken,
  writeStoredTokens,
} from '@/httpapis/auth-session';
import { fetchMe } from '@/httpapis/me';
import type { RequestOtpResult, VerifyOtpResult } from '@/model/auth';
import type { CurrentMemberProfile, SessionUserView } from '@/model/user';

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
  ajo_balance: member?.ajo_balance ?? 0,
});

// 1. 建立會員登入狀態 Store
export const useSessionStore = defineStore('session', {
  state: () => ({
    me: null as CurrentMemberProfile | null,
    accessToken: readStoredAccessToken(),
    refreshToken: readStoredRefreshToken(),
    isHydrating: false,
    isLoaded: false,
  }),
  getters: {
    // 2. 判斷是否已登入
    isAuthenticated: (state) => Boolean(state.accessToken),
    // 3. 輸出導覽層可直接使用的會員資料
    currentUser: (state) => buildSessionUser(state.me),
  },
  actions: {
    // 4. 保存 token
    setTokens(accessToken: string, refreshToken: string) {
      this.accessToken = accessToken;
      this.refreshToken = refreshToken;
      writeStoredTokens(accessToken, refreshToken);
    },

    // 5. 清除登入狀態
    clearSession() {
      this.me = null;
      this.accessToken = '';
      this.refreshToken = '';
      clearStoredTokens();
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

    // 9. 請求郵箱驗證碼
    async sendEmailOtp(email: string, scene = 'login'): Promise<RequestOtpResult> {
      const { data } = await requestEmailOtp({
        email,
        scene,
      });

      return data.data;
    },

    // 10. 驗證郵箱驗證碼並登入
    async signInWithEmailOtp(
      email: string,
      code: string,
      displayName: string,
      scene = 'login',
    ): Promise<VerifyOtpResult> {
      const { data } = await verifyEmailOtp({
        email,
        scene,
        code,
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

    // 11. 請求電郵重設密碼驗證碼
    async sendPasswordResetEmail(email: string): Promise<RequestOtpResult> {
      const { data } = await requestEmailPasswordReset({
        email,
      });

      return data.data;
    },

    // 12. 使用電郵驗證碼重設密碼
    async resetPasswordByEmail(email: string, code: string, password: string): Promise<boolean> {
      const { data } = await resetPasswordWithEmail({
        email,
        code,
        password,
      });

      return data.data.password_reset;
    },

    // 13. 使用郵箱密碼登入
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

    // 14. 使用手機密碼登入
    async signInWithPhone(phoneCountryCode: string, phoneNumber: string, password: string): Promise<VerifyOtpResult> {
      const { data } = await loginWithPhone({
        phone_country_code: phoneCountryCode,
        phone_number: phoneNumber,
        password,
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

    // 15. 使用用戶名密碼登入
    async signInWithUsername(username: string, password: string): Promise<VerifyOtpResult> {
      const { data } = await loginWithUsername({
        username,
        password,
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

    // 16. 註冊郵箱與手機密碼帳戶
    async registerEmailAccount(
      email: string,
      password: string,
      displayName: string,
      phoneCountryCode: string,
      phoneNumber: string,
      username: string,
      publisherIdentityType = '',
      primaryCommunityID = '',
      primaryCommunityName = '',
      residenceFloor = '',
      residenceUnit = '',
    ): Promise<VerifyOtpResult> {
      const trimmedEmail = email.trim();
      const { data } = await registerWithEmail({
        email: trimmedEmail,
        password,
        display_name: displayName,
        phone_country_code: phoneCountryCode,
        phone_number: phoneNumber,
        username,
        publisher_identity_type: publisherIdentityType,
        primary_community_id: primaryCommunityID,
        primary_community_name: primaryCommunityName,
        residence_floor: residenceFloor,
        residence_unit: residenceUnit,
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

    // 17. 使用 ismart 帳戶登入
    async signInWithIsmart(account: string, password: string, phone?: string, email?: string): Promise<VerifyOtpResult> {
      const { data } = await loginWithIsmart({
        account,
        password,
        phone,
        email,
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

    // 18. 讀取目前會員資料
    async loadCurrentUser() {
      if (!this.accessToken) {
        this.me = null;
        return this.me;
      }
      const { data } = await fetchMe();
      this.me = data.data;
      return data.data;
    },

    // 19. 執行登出
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

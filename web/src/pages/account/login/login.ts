/*
 * 登入頁 - 狀態與資料流程。
 * 1. 管理登入表單狀態、模式切換與顯示文字。
 * 2. 處理郵箱密碼登入、手機密碼登入、郵箱驗證碼登入、住戶註冊與登出。
 * 3. 統一錯誤提示與登入成功後跳轉。
 */
import axios from 'axios';
import { computed, onMounted, reactive, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import { fetchLoginHero } from '@/httpapis/home-content';
import type { RequestOtpResult } from '@/model/auth';
import type { LoginHeroImageSetting } from '@/model/home-content';
import { useFeedbackStore } from '@/stores/feedback';
import { useSessionStore } from '@/stores/session';

export type LoginAuthMode = 'email' | 'phone';

export type LoginEmailAction = 'login' | 'register';

export type LoginEmailMethod = 'password' | 'code';

export interface LoginFormState {
  email: string;
  password: string;
  displayName: string;
  phone: string;
  otp: string;
}

export interface LoginHeroImage {
  readonly src: string;
  readonly author: string;
  readonly location: string;
}

interface ParsedPhoneInput {
  phoneCountryCode: string;
  phoneNumber: string;
}

const LOGIN_HERO_MAX_IMAGES = 3;
const EMAIL_PLACEHOLDER = 'name@example.com';

const LOGIN_HERO_IMAGES: readonly LoginHeroImage[] = [
  {
    src: '/images/pexels-jimmy-teoh-294331-35774007.jpg',
    author: '@jimmy teoh',
    location: 'Victoria Harbour, HK',
  },
  {
    src: '/images/pexels-kseniya-kobi-3624194-7820979.jpg',
    author: '@kseniya kobi',
    location: 'Hong Kong Residence',
  },
];

// 1. 建立登入表單初始狀態
const createInitialFormState = (): LoginFormState => ({
  email: '',
  password: '',
  displayName: '',
  phone: '',
  otp: '',
});

// 2. 解析手機輸入
const parsePhoneInput = (rawValue: string): ParsedPhoneInput => {
  const compactValue = rawValue.replace(/[()-]/g, ' ').trim();
  const parts = compactValue.split(/\s+/).filter(Boolean);

  if (parts.length >= 2) {
    return {
      phoneCountryCode: parts[0].startsWith('+') ? parts[0] : `+${parts[0]}`,
      phoneNumber: parts.slice(1).join('').replace(/\D/g, ''),
    };
  }

  const noSpaces = compactValue.replace(/\s+/g, '');
  if (noSpaces.startsWith('+852') && noSpaces.length > 4) {
    return {
      phoneCountryCode: '+852',
      phoneNumber: noSpaces.slice(4),
    };
  }

  return {
    phoneCountryCode: '',
    phoneNumber: '',
  };
};

// 3. 檢查郵箱格式
const isValidEmailInput = (email: string): boolean => {
  const value = email.trim();
  return value.includes('@') && value.includes('.') && value.length <= 255;
};

// 4. 組合錯誤訊息
const readErrorMessage = (error: unknown): string =>
  axios.isAxiosError(error)
    ? error.response?.data?.message ?? error.message
    : 'Request failed.';

// 5. 取得隨機登入頁主視覺
const getRandomHeroImage = (): LoginHeroImage => {
  const selectedIndex = Math.floor(Math.random() * LOGIN_HERO_IMAGES.length);
  return LOGIN_HERO_IMAGES[selectedIndex] ?? LOGIN_HERO_IMAGES[0];
};

// 6. 將後台登入背景圖轉為頁面主視覺列表
const buildConfiguredHeroImages = (items: LoginHeroImageSetting[]): LoginHeroImage[] =>
  items
    .slice()
    .sort((left, right) => left.sort_order - right.sort_order)
    .filter((item) => item.url.trim().length > 0)
    .slice(0, LOGIN_HERO_MAX_IMAGES)
    .map((item) => ({
      src: item.url,
      author: item.author || 'AJO Living',
      location: item.location || 'Hong Kong',
    }));

// 7. 管理登入頁資料與動作
export const useLoginPage = () => {
  const route = useRoute();
  const router = useRouter();
  const { t } = useI18n();
  const feedbackStore = useFeedbackStore();
  const sessionStore = useSessionStore();
  const formState = reactive(createInitialFormState());
  const authMode = ref<LoginAuthMode>('email');
  const emailAction = ref<LoginEmailAction>('login');
  const emailLoginMethod = ref<LoginEmailMethod>('password');
  const requestingOtp = ref(false);
  const submitting = ref(false);
  const rememberMe = ref(true);
  const selectedHero = ref(getRandomHeroImage());

  const isAuthenticated = computed(() => sessionStore.isAuthenticated);

  const emailSubmitLabel = computed(() => {
    if (submitting.value) {
      return t('auth.loading');
    }

    return emailAction.value === 'register' ? t('auth.registerSubmit') : t('auth.submit');
  });

  const submitLabel = computed(() => {
    return authMode.value === 'email' || emailAction.value === 'register'
      ? emailSubmitLabel.value
      : submitting.value ? t('auth.loading') : t('auth.submit');
  });

  const otpRequestLabel = computed(() =>
    requestingOtp.value ? t('auth.sending') : t('auth.requestOtp'),
  );

  const emailActionSwitchLabel = computed(() =>
    emailAction.value === 'register' ? t('auth.emailLogin') : t('auth.emailRegister'),
  );

  const footerPrompt = computed(() =>
    emailAction.value === 'register' ? t('auth.alreadyHaveAccount') : t('auth.dontHaveAccount'),
  );

  // 7.1 載入後台登入背景圖
  const loadConfiguredHero = async (): Promise<void> => {
    try {
      const { data } = await fetchLoginHero();
      const configuredHeroes = buildConfiguredHeroImages(data.data.items);
      if (configuredHeroes.length > 0) {
        const selectedIndex = Math.floor(Math.random() * configuredHeroes.length);
        selectedHero.value = configuredHeroes[selectedIndex] ?? configuredHeroes[0];
      }
    } catch {
      selectedHero.value = selectedHero.value || getRandomHeroImage();
    }
  };

  // 7.2 請求驗證碼
  const handleRequestOtp = async (): Promise<void> => {
    let result: RequestOtpResult;
    requestingOtp.value = true;

    try {
      if (!isValidEmailInput(formState.email)) {
        feedbackStore.pushToast(t('auth.invalidEmail'), 'error');
        return;
      }

      result = await sessionStore.sendEmailOtp(
        formState.email.trim(),
        emailAction.value === 'register' ? 'register' : 'login',
      );

      feedbackStore.pushToast(
        result.mock_code ? t('auth.otpPreview', { code: result.mock_code }) : t('auth.otpSent'),
        'info',
      );
    } catch (error) {
      feedbackStore.pushToast(readErrorMessage(error), 'error');
    } finally {
      requestingOtp.value = false;
    }
  };

  // 7.3 完成登入後跳轉
  const redirectAfterSignIn = async (): Promise<void> => {
    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/marketplace';
    await router.push(redirect);
  };

  // 7.4 執行郵箱登入或註冊
  const handleEmailSubmit = async (): Promise<void> => {
    if (!isValidEmailInput(formState.email)) {
      feedbackStore.pushToast(t('auth.invalidEmail'), 'error');
      return;
    }

    if (emailAction.value === 'register') {
      const { phoneCountryCode, phoneNumber } = parsePhoneInput(formState.phone);
      if (!formState.displayName.trim() || !phoneCountryCode || !phoneNumber || formState.password.trim().length < 8) {
        feedbackStore.pushToast(t('auth.registerRequiredFields'), 'error');
        return;
      }
    } else if (emailLoginMethod.value === 'code') {
      if (formState.otp.trim().length === 0) {
        feedbackStore.pushToast(t('auth.emailOtpRequiredFields'), 'error');
        return;
      }
    } else if (formState.password.trim().length === 0) {
      feedbackStore.pushToast(t('auth.passwordRequired'), 'error');
      return;
    }

    submitting.value = true;

    try {
      if (emailAction.value === 'register') {
        const { phoneCountryCode, phoneNumber } = parsePhoneInput(formState.phone);
        await sessionStore.registerEmailAccount(
          formState.email.trim(),
          formState.password,
          formState.displayName.trim(),
          phoneCountryCode,
          phoneNumber,
        );
        feedbackStore.pushToast(t('auth.registerSuccess'), 'success');
      } else if (emailLoginMethod.value === 'code') {
        await sessionStore.signInWithEmailOtp(
          formState.email.trim(),
          formState.otp.trim(),
          '',
          'login',
        );
        feedbackStore.pushToast(t('auth.signInSuccess'), 'success');
      } else {
        await sessionStore.signInWithEmail(formState.email.trim(), formState.password);
        feedbackStore.pushToast(t('auth.signInSuccess'), 'success');
      }

      await redirectAfterSignIn();
    } catch (error) {
      feedbackStore.pushToast(readErrorMessage(error), 'error');
    } finally {
      submitting.value = false;
    }
  };

  // 7.5 執行手機密碼登入
  const handlePhoneSubmit = async (): Promise<void> => {
    const { phoneCountryCode, phoneNumber } = parsePhoneInput(formState.phone);
    if (!phoneCountryCode || !phoneNumber || formState.password.trim().length === 0) {
      feedbackStore.pushToast(t('auth.phonePasswordRequired'), 'error');
      return;
    }

    submitting.value = true;

    try {
      await sessionStore.signInWithPhone(phoneCountryCode, phoneNumber, formState.password);
      feedbackStore.pushToast(t('auth.signInSuccess'), 'success');
      await redirectAfterSignIn();
    } catch (error) {
      feedbackStore.pushToast(readErrorMessage(error), 'error');
    } finally {
      submitting.value = false;
    }
  };

  // 7.6 按目前模式提交登入表單
  const handleSubmit = async (): Promise<void> => {
    if (authMode.value === 'email' || emailAction.value === 'register') {
      await handleEmailSubmit();
      return;
    }

    await handlePhoneSubmit();
  };

  // 7.7 執行登出
  const handleSignOut = async (): Promise<void> => {
    await sessionStore.signOut();
    feedbackStore.pushToast(t('auth.signOutSuccess'), 'success');
  };

  // 7.8 設定登入模式
  const setAuthMode = (mode: LoginAuthMode): void => {
    authMode.value = mode;
    emailAction.value = 'login';
    formState.otp = '';
  };

  // 7.9 切換郵箱登入與註冊模式
  const toggleEmailAction = (): void => {
    emailAction.value = emailAction.value === 'register' ? 'login' : 'register';
    authMode.value = 'email';
    formState.otp = '';
  };

  // 7.10 切換郵箱登入方式
  const setEmailLoginMethod = (method: LoginEmailMethod): void => {
    authMode.value = 'email';
    emailAction.value = 'login';
    emailLoginMethod.value = method;
    formState.otp = '';
  };

  onMounted(() => {
    void loadConfiguredHero();
  });

  return {
    authMode,
    emailAction,
    emailActionSwitchLabel,
    emailLoginMethod,
    emailPlaceholder: EMAIL_PLACEHOLDER,
    footerPrompt,
    formState,
    handleRequestOtp,
    handleSignOut,
    handleSubmit,
    isAuthenticated,
    otpRequestLabel,
    rememberMe,
    requestingOtp,
    selectedHero,
    setAuthMode,
    setEmailLoginMethod,
    submitting,
    submitLabel,
    toggleEmailAction,
  };
};

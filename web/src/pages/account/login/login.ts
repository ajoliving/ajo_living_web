/*
 * 登入頁 - 狀態與資料流程。
 * 1. 管理登入表單狀態、模式切換與顯示文字。
 * 2. 處理郵箱登入、郵箱註冊、短訊 OTP 登入與登出。
 * 3. 統一錯誤提示與登入成功後跳轉。
 */
import axios from 'axios';
import { computed, reactive, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import { useFeedbackStore } from '@/stores/feedback';
import { useSessionStore } from '@/stores/session';

export type LoginAuthMode = 'email' | 'otp';

export type LoginEmailAction = 'login' | 'register';

export type LoginAuthModeSwitchIcon = 'phone' | 'user';

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

const DEFAULT_LOGIN_EMAIL = 'admin@admin.com';
const DEFAULT_LOGIN_PASSWORD = 'admin123';
const EMAIL_PLACEHOLDER = DEFAULT_LOGIN_EMAIL;

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
  email: DEFAULT_LOGIN_EMAIL,
  password: DEFAULT_LOGIN_PASSWORD,
  displayName: 'Admin',
  phone: '+852 6123 4567',
  otp: '246810',
});

// 2. 解析手機輸入
const parsePhoneInput = (rawValue: string): ParsedPhoneInput => {
  const compactValue = rawValue.replace(/-/g, ' ').trim();
  const parts = compactValue.split(/\s+/).filter(Boolean);

  if (parts.length >= 2) {
    return {
      phoneCountryCode: parts[0].startsWith('+') ? parts[0] : `+${parts[0]}`,
      phoneNumber: parts.slice(1).join(''),
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

// 3. 組合錯誤訊息
const readErrorMessage = (error: unknown): string =>
  axios.isAxiosError(error)
    ? error.response?.data?.message ?? error.message
    : 'Request failed.';

// 4. 取得隨機登入頁主視覺
const getRandomHeroImage = (): LoginHeroImage => {
  const selectedIndex = Math.floor(Math.random() * LOGIN_HERO_IMAGES.length);
  return LOGIN_HERO_IMAGES[selectedIndex] ?? LOGIN_HERO_IMAGES[0];
};

// 5. 管理登入頁資料與動作
export const useLoginPage = () => {
  const route = useRoute();
  const router = useRouter();
  const { t } = useI18n();
  const feedbackStore = useFeedbackStore();
  const sessionStore = useSessionStore();
  const formState = reactive(createInitialFormState());
  const authMode = ref<LoginAuthMode>('email');
  const emailAction = ref<LoginEmailAction>('login');
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
    if (authMode.value === 'email') {
      return emailSubmitLabel.value;
    }

    return submitting.value ? t('auth.loading') : t('auth.submit');
  });

  const otpRequestLabel = computed(() =>
    requestingOtp.value ? t('auth.sending') : t('auth.requestOtp'),
  );

  const authModeSwitchIcon = computed<LoginAuthModeSwitchIcon>(() =>
    authMode.value === 'email' ? 'phone' : 'user',
  );

  const authModeSwitchLabel = computed(() =>
    authMode.value === 'email' ? t('auth.otpMode') : t('auth.emailMode'),
  );

  const emailActionSwitchLabel = computed(() =>
    emailAction.value === 'register' ? t('auth.emailLogin') : t('auth.emailRegister'),
  );

  const footerPrompt = computed(() =>
    emailAction.value === 'register' ? t('auth.alreadyHaveAccount') : t('auth.dontHaveAccount'),
  );

  const footerActionLabel = computed(() =>
    emailAction.value === 'register' ? t('auth.emailLogin') : t('auth.emailRegister'),
  );

  // 5.1 請求 OTP
  const handleRequestOtp = async (): Promise<void> => {
    const { phoneCountryCode, phoneNumber } = parsePhoneInput(formState.phone);
    if (!phoneCountryCode || !phoneNumber) {
      feedbackStore.pushToast(t('auth.invalidPhone'), 'error');
      return;
    }

    requestingOtp.value = true;

    try {
      const result = await sessionStore.sendOtp(phoneCountryCode, phoneNumber);
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

  // 5.2 完成登入後跳轉
  const redirectAfterSignIn = async (): Promise<void> => {
    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/marketplace';
    await router.push(redirect);
  };

  // 5.3 執行郵箱密碼登入或註冊
  const handleEmailSubmit = async (): Promise<void> => {
    if (!formState.email.trim() || formState.password.trim().length < 8) {
      feedbackStore.pushToast(t('auth.emailRequiredFields'), 'error');
      return;
    }

    submitting.value = true;

    try {
      if (emailAction.value === 'register') {
        await sessionStore.registerEmailAccount(
          formState.email.trim(),
          formState.password,
          formState.displayName.trim(),
        );
        feedbackStore.pushToast(t('auth.registerSuccess'), 'success');
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

  // 5.4 執行 OTP 登入
  const handleOtpSubmit = async (): Promise<void> => {
    const { phoneCountryCode, phoneNumber } = parsePhoneInput(formState.phone);
    if (!phoneCountryCode || !phoneNumber || formState.otp.trim().length === 0) {
      feedbackStore.pushToast(t('auth.requiredFields'), 'error');
      return;
    }

    submitting.value = true;

    try {
      await sessionStore.signInWithOtp(phoneCountryCode, phoneNumber, formState.otp.trim());
      feedbackStore.pushToast(t('auth.signInSuccess'), 'success');
      await redirectAfterSignIn();
    } catch (error) {
      feedbackStore.pushToast(readErrorMessage(error), 'error');
    } finally {
      submitting.value = false;
    }
  };

  // 5.5 按目前模式提交登入表單
  const handleSubmit = async (): Promise<void> => {
    if (authMode.value === 'email') {
      await handleEmailSubmit();
      return;
    }

    await handleOtpSubmit();
  };

  // 5.6 執行登出
  const handleSignOut = async (): Promise<void> => {
    await sessionStore.signOut();
    feedbackStore.pushToast(t('auth.signOutSuccess'), 'success');
  };

  // 5.7 切換登入模式
  const toggleAuthMode = (): void => {
    authMode.value = authMode.value === 'email' ? 'otp' : 'email';
  };

  // 5.8 切換郵箱登入與註冊模式
  const toggleEmailAction = (): void => {
    emailAction.value = emailAction.value === 'register' ? 'login' : 'register';
    authMode.value = 'email';
  };

  return {
    authMode,
    authModeSwitchIcon,
    authModeSwitchLabel,
    emailAction,
    emailActionSwitchLabel,
    emailPlaceholder: EMAIL_PLACEHOLDER,
    footerActionLabel,
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
    submitting,
    submitLabel,
    toggleAuthMode,
    toggleEmailAction,
  };
};

/*
 * 登入頁 - 狀態與資料流程。
 * 1. 管理登入表單狀態、模式切換與顯示文字。
 * 2. 處理郵箱密碼登入、手機密碼登入、郵箱驗證碼登入、ismart 帳戶登入與住戶註冊。
 * 3. 統一錯誤提示與登入成功後跳轉。
 */
import axios from 'axios';
import { computed, onMounted, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import { fetchPosBuildings, fetchPosBuildingUnits } from '@/httpapis/building';
import { fetchLoginHero } from '@/httpapis/home-content';
import type { RequestOtpResult } from '@/model/auth';
import type { PosBuilding, PosBuildingUnit } from '@/model/community';
import type { LoginHeroImageSetting } from '@/model/home-content';
import { useFeedbackStore } from '@/stores/feedback';
import { useSessionStore } from '@/stores/session';

export type LoginAuthMode = 'email' | 'phone' | 'username';

export type LoginEmailAction = 'login' | 'register';

export type LoginEmailMethod = 'password' | 'code';

export type LoginAccountSource = 'local' | 'ismart';

export interface LoginFormState {
  email: string;
  password: string;
  displayName: string;
  phone: string;
  otp: string;
  ismartAccount: string;
  primaryCommunityID: string;
  residenceFloor: string;
  residenceUnit: string;
}

export interface LoginHeroImage {
  readonly src: string;
  readonly author: string;
  readonly location: string;
}

export interface LoginSelectOption {
  label: string;
  value: string;
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
  ismartAccount: '',
  primaryCommunityID: '',
  residenceFloor: '',
  residenceUnit: '',
});

// 2. 解析手機輸入
const parsePhoneInput = (rawValue: string): ParsedPhoneInput => {
  const compactValue = rawValue.replace(/[()-]/g, ' ').trim();
  const parts = compactValue.split(/\s+/).filter(Boolean);
  const normalizedValue = compactValue.replace(/\s+/g, '');

  if (!normalizedValue) {
    return {
      phoneCountryCode: '',
      phoneNumber: '',
    };
  }

  if (parts.length >= 2 && /^\+?\d{1,4}$/.test(parts[0])) {
    return {
      phoneCountryCode: parts[0].startsWith('+') ? parts[0] : `+${parts[0]}`,
      phoneNumber: parts.slice(1).join('').replace(/\D/g, ''),
    };
  }

  if (normalizedValue.startsWith('+852') && normalizedValue.length > 4) {
    return {
      phoneCountryCode: '+852',
      phoneNumber: normalizedValue.slice(4).replace(/\D/g, ''),
    };
  }

  if (normalizedValue.startsWith('+')) {
    const matched = normalizedValue.match(/^(\+\d{1,4})(\d{4,32})$/);
    return {
      phoneCountryCode: matched?.[1] ?? '',
      phoneNumber: matched?.[2] ?? '',
    };
  }

  const digits = normalizedValue.replace(/\D/g, '');
  if (digits.startsWith('852') && digits.length > 10) {
    return {
      phoneCountryCode: '+852',
      phoneNumber: digits.slice(3),
    };
  }

  if (digits.length >= 4) {
    return {
      phoneCountryCode: '+852',
      phoneNumber: digits,
    };
  }

  return {
    phoneCountryCode: '',
    phoneNumber: '',
  };
};

// 2.1 檢查手機格式是否符合後端要求
const isValidParsedPhone = ({ phoneCountryCode, phoneNumber }: ParsedPhoneInput): boolean => {
  const countryCodeDigits = phoneCountryCode.replace(/^\+/, '');
  return (
    phoneCountryCode.startsWith('+') &&
    phoneCountryCode.length <= 8 &&
    phoneNumber.length >= 4 &&
    phoneNumber.length <= 32 &&
    /^\d+$/.test(countryCodeDigits) &&
    /^\d+$/.test(phoneNumber)
  );
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

// 7. 讀取 POS 大廈 ID
const getBuildingId = (item: PosBuilding): string =>
  String(item.building_id ?? item.id ?? '').trim();

// 8. 讀取 POS 大廈名稱
const getBuildingName = (item: PosBuilding): string =>
  String(item.buildname_chi ?? item.buildname ?? item.name ?? getBuildingId(item)).trim();

// 9. 讀取 POS 單位 ID
const getUnitId = (item: PosBuildingUnit): string =>
  String(item.unit_id ?? item.id ?? '').trim();

// 10. 讀取 POS 單位樓層
const getRawUnitFloor = (item: PosBuildingUnit): string =>
  String(item.floor ?? '').trim();

// 11. 讀取 POS 單位名稱
const getRawUnitName = (item: PosBuildingUnit): string =>
  String(item.unit ?? item.unit_name ?? item.name ?? '').trim();

// 12. 讀取 POS 單位顯示名稱
const getUnitName = (item: PosBuildingUnit): string =>
  getRawUnitName(item) || getUnitId(item);

// 13. 取出數字字串
const digitsOnly = (value: unknown): string => String(value ?? '').replace(/\D/g, '');

// 14. 排除 POS 回傳的樓宇本身佔位資料
const isSelectableUnit = (buildingID: string, item: PosBuildingUnit): boolean => {
  const normalizedBuildingID = digitsOnly(buildingID).slice(0, 7);
  const normalizedUnitID = digitsOnly(getUnitId(item));
  const hasFloor = getRawUnitFloor(item).length > 0;
  const hasUnit = getRawUnitName(item).length > 0;

  if (
    normalizedBuildingID &&
    normalizedUnitID === normalizedBuildingID &&
    !hasFloor &&
    !hasUnit
  ) {
    return false;
  }

  return true;
};

// 15. 取得樓層與單位排序分組
const getDisplaySortBucket = (value: string): number => {
  const normalized = value.trim().toUpperCase();
  if (!normalized) {
    return 2;
  }
  if (/^[A-Z]/.test(normalized)) {
    return 0;
  }
  if (/^\d/.test(normalized)) {
    return 1;
  }
  return 2;
};

// 16. 依顯示規則排序樓層與單位
const compareDisplayCodes = (left: string, right: string): number => {
  const bucketDiff = getDisplaySortBucket(left) - getDisplaySortBucket(right);
  if (bucketDiff !== 0) {
    return bucketDiff;
  }

  return left.localeCompare(right, 'en', {
    numeric: true,
    sensitivity: 'base',
  });
};

// 17. 管理登入頁資料與動作
export const useLoginPage = () => {
  const route = useRoute();
  const router = useRouter();
  const { t } = useI18n();
  const feedbackStore = useFeedbackStore();
  const sessionStore = useSessionStore();
  const formState = reactive(createInitialFormState());
  const authMode = ref<LoginAuthMode>('username');
  const emailAction = ref<LoginEmailAction>('login');
  const emailLoginMethod = ref<LoginEmailMethod>('password');
  const accountSource = ref<LoginAccountSource>('local');
  const buildings = ref<PosBuilding[]>([]);
  const buildingUnits = ref<PosBuildingUnit[]>([]);
  const buildingsLoading = ref(false);
  const unitsLoading = ref(false);
  const requestingOtp = ref(false);
  const submitting = ref(false);
  const rememberMe = ref(true);
  const selectedHero = ref(getRandomHeroImage());
  let latestUnitRequestID = 0;

  const isAuthenticated = computed(() => sessionStore.isAuthenticated);

  const emailSubmitLabel = computed(() => {
    if (submitting.value) {
      return t('auth.loading');
    }

    return emailAction.value === 'register' ? t('auth.registerSubmit') : t('auth.submit');
  });

  const submitLabel = computed(() => {
    if (emailAction.value === 'login' && authMode.value === 'username') {
      return submitting.value ? t('auth.loading') : t('auth.submit');
    }

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

  const selectedIsmartAccount = computed(() => formState.ismartAccount.trim());
  const selectedBuildingName = computed(() =>
    buildingOptions.value.find((option) => option.value === formState.primaryCommunityID)?.label ?? '',
  );

  const buildingOptions = computed<LoginSelectOption[]>(() => {
    const placeholder = buildingsLoading.value
      ? t('auth.residenceBuildingLoading')
      : t('auth.residenceBuildingPlaceholder');

    return [
      { label: placeholder, value: '' },
      ...buildings.value
        .slice()
        .sort((left, right) => getBuildingName(left).localeCompare(getBuildingName(right), 'en', {
          numeric: true,
          sensitivity: 'base',
        }))
        .map((item) => ({
          label: getBuildingName(item),
          value: getBuildingId(item),
        }))
        .filter((option) => option.value.length > 0),
    ];
  });

  const residenceFloorOptions = computed<LoginSelectOption[]>(() => {
    const placeholder = unitsLoading.value
      ? t('auth.residenceUnitLoading')
      : t('auth.residenceFloorPlaceholder');
    const floors = Array.from(
      new Set(
        buildingUnits.value
          .map((item) => getRawUnitFloor(item) || t('auth.residenceUnassignedFloor'))
          .filter(Boolean),
      ),
    ).sort(compareDisplayCodes);

    return [
      { label: placeholder, value: '' },
      ...floors.map((floor) => ({
        label: floor,
        value: floor,
      })),
    ];
  });

  const residenceUnitOptions = computed<LoginSelectOption[]>(() => {
    const placeholder = unitsLoading.value
      ? t('auth.residenceUnitLoading')
      : t('auth.residenceUnitPlaceholder');

    return [
      { label: placeholder, value: '' },
      ...buildingUnits.value
        .filter((item) => {
          const floor = getRawUnitFloor(item) || t('auth.residenceUnassignedFloor');
          return floor === formState.residenceFloor;
        })
        .slice()
        .sort((left, right) => compareDisplayCodes(getUnitName(left), getUnitName(right)))
        .map((item) => ({
          label: getUnitName(item),
          value: getUnitName(item),
        }))
        .filter((option) => option.value.length > 0),
    ];
  });

  // 18. 重置樓層與單位級聯選擇
  watch(
    () => formState.primaryCommunityID,
    (nextValue, previousValue) => {
      if (nextValue !== previousValue) {
        formState.residenceFloor = '';
        formState.residenceUnit = '';
        buildingUnits.value = [];
        if (nextValue.trim()) {
          void loadUnitsForBuilding(nextValue.trim());
        }
      }
    },
  );

  // 19. 重置單位選擇
  watch(
    () => formState.residenceFloor,
    (nextValue, previousValue) => {
      if (nextValue !== previousValue) {
        formState.residenceUnit = '';
      }
    },
  );

  // 20. 載入後台登入背景圖
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

  // 21. 載入住戶註冊大廈選項
  const loadBuildings = async (): Promise<void> => {
    buildingsLoading.value = true;
    try {
      buildings.value = await fetchPosBuildings();
    } catch {
      buildings.value = [];
    } finally {
      buildingsLoading.value = false;
    }
  };

  // 22. 載入指定大廈單位清單
  const loadUnitsForBuilding = async (buildingID: string): Promise<void> => {
    const requestID = ++latestUnitRequestID;
    unitsLoading.value = true;
    try {
      const units = await fetchPosBuildingUnits(buildingID);
      if (requestID !== latestUnitRequestID) {
        return;
      }

      buildingUnits.value = units.filter((item) => isSelectableUnit(buildingID, item));
    } catch (error) {
      if (requestID !== latestUnitRequestID) {
        return;
      }

      buildingUnits.value = [];
      feedbackStore.pushToast(readErrorMessage(error), 'error');
    } finally {
      if (requestID === latestUnitRequestID) {
        unitsLoading.value = false;
      }
    }
  };

  // 23. 請求郵箱驗證碼
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

  // 24. 完成登入後跳轉
  const redirectAfterSignIn = async (): Promise<void> => {
    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/marketplace';
    await router.push(redirect);
  };

  // 25. 執行郵箱登入或註冊
  const handleEmailSubmit = async (): Promise<void> => {
    if (emailAction.value === 'register') {
      if (accountSource.value === 'ismart') {
        await handleIsmartSubmit();
        return;
      }

      const optionalEmail = formState.email.trim();
      if (optionalEmail && !isValidEmailInput(optionalEmail)) {
        feedbackStore.pushToast(t('auth.invalidEmail'), 'error');
        return;
      }

      const { phoneCountryCode, phoneNumber } = parsePhoneInput(formState.phone);
      if (!formState.displayName.trim() || !isValidParsedPhone({ phoneCountryCode, phoneNumber }) || formState.password.trim().length < 8) {
        feedbackStore.pushToast(t('auth.registerRequiredFields'), 'error');
        return;
      }
    } else if (!isValidEmailInput(formState.email)) {
      feedbackStore.pushToast(t('auth.invalidEmail'), 'error');
      return;
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
          formState.displayName.trim(),
          formState.primaryCommunityID.trim(),
          selectedBuildingName.value.trim(),
          formState.residenceFloor.trim(),
          formState.residenceUnit.trim(),
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

  // 26. 執行手機密碼登入
  const handlePhoneSubmit = async (): Promise<void> => {
    const { phoneCountryCode, phoneNumber } = parsePhoneInput(formState.phone);
    if (!isValidParsedPhone({ phoneCountryCode, phoneNumber }) || formState.password.trim().length === 0) {
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

  // 27. 使用 ismart 帳戶登入或綁定
  const handleIsmartSubmit = async (): Promise<void> => {
    if (selectedIsmartAccount.value.length === 0 || formState.password.trim().length === 0) {
      feedbackStore.pushToast(t('auth.ismartRequiredFields'), 'error');
      return;
    }

    submitting.value = true;

    try {
      await sessionStore.signInWithIsmart(
        selectedIsmartAccount.value,
        formState.password,
        undefined,
        formState.email.trim() || undefined,
      );
      feedbackStore.pushToast(t('auth.signInSuccess'), 'success');
      await redirectAfterSignIn();
    } catch (error) {
      feedbackStore.pushToast(readErrorMessage(error), 'error');
    } finally {
      submitting.value = false;
    }
  };

  // 28. 按目前模式提交登入表單
  const handleSubmit = async (): Promise<void> => {
    if (emailAction.value === 'login' && authMode.value === 'username') {
      await handleIsmartSubmit();
      return;
    }

    if (authMode.value === 'email' || emailAction.value === 'register') {
      await handleEmailSubmit();
      return;
    }

    await handlePhoneSubmit();
  };

  // 29. 執行登出
  const handleSignOut = async (): Promise<void> => {
    await sessionStore.signOut();
    feedbackStore.pushToast(t('auth.signOutSuccess'), 'success');
  };

  // 30. 設定登入模式
  const setAuthMode = (mode: LoginAuthMode): void => {
    authMode.value = mode;
    emailAction.value = 'login';
    formState.otp = '';
  };

  // 31. 設定註冊帳戶來源
  const setAccountSource = (source: LoginAccountSource): void => {
    accountSource.value = source;
    emailAction.value = 'register';
    authMode.value = 'email';
    formState.otp = '';
  };

  // 32. 切換郵箱登入與註冊模式
  const toggleEmailAction = (): void => {
    accountSource.value = 'local';
    emailAction.value = emailAction.value === 'register' ? 'login' : 'register';
    authMode.value = emailAction.value === 'register' ? 'email' : 'phone';
    formState.otp = '';
  };

  // 33. 切換郵箱登入方式
  const setEmailLoginMethod = (method: LoginEmailMethod): void => {
    authMode.value = 'email';
    emailAction.value = 'login';
    emailLoginMethod.value = method;
    formState.otp = '';
  };

  onMounted(() => {
    void loadConfiguredHero();
    void loadBuildings();
  });

  return {
    accountSource,
    authMode,
    buildingOptions,
    buildingsLoading,
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
    residenceFloorOptions,
    residenceUnitOptions,
    selectedHero,
    setAccountSource,
    setAuthMode,
    setEmailLoginMethod,
    submitting,
    submitLabel,
    toggleEmailAction,
    unitsLoading,
  };
};

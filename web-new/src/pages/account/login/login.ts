/*
 * 登入頁 - 狀態與資料流程。
 * 1. 管理統一帳戶登入與用戶註冊表單狀態。
 * 2. 登入只提交一次帳戶識別與密碼，由後端辨識電話、電郵或 username 並處理回退。
 * 3. 統一錯誤提示與登入成功後跳轉。
 */
import axios from 'axios';
import { computed, onMounted, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import { fetchPosBuildings, fetchPosBuildingUnits } from '@/httpapis/building';
import { fetchLoginHero } from '@/httpapis/home-content';
import type { PosBuilding, PosBuildingUnit } from '@/model/community';
import type { LoginHeroImageSetting } from '@/model/home-content';
import type { RegisterEmailAccountPayload } from '@/model/auth';
import { useFeedbackStore } from '@/stores/feedback';
import { useSessionStore } from '@/stores/session';

export type LoginEmailAction = 'login' | 'register';

export interface LoginFormState {
  email: string;
  password: string;
  confirmPassword: string;
  username: string;
  engName: string;
  chiName: string;
  phone: string;
  phoneCountryCode: string;
  publisherIdentityType: 'personal' | 'individual_agent' | 'agency_company';
  otp: string;
  ismartAccount: string;
  primaryCommunityID: string;
  residenceFloor: string;
  residenceUnit: string;
  idCard: string;
  shouldBindResidence: boolean;
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

export type RegistrationValidationErrorKey =
  | 'auth.invalidEmail'
  | 'auth.registerUsernameInvalid'
  | 'auth.registerEnglishNameInvalid'
  | 'auth.idCardRequired'
  | 'auth.invalidPhone'
  | 'auth.registerPasswordTooShort'
  | 'auth.registerPasswordMismatch'
  | 'auth.residenceBindingRequired';

export type LoginFormField = 'email' | 'username' | 'engName' | 'idCard' | 'ismartAccount' | 'password' | 'confirmPassword' | 'phone' | 'residence';

export type LoginValidationErrorKey = RegistrationValidationErrorKey | 'auth.accountRequired' | 'auth.passwordRequired';

export type LoginValidationErrors = Partial<Record<LoginFormField, LoginValidationErrorKey>>;

const LOGIN_HERO_MAX_IMAGES = 3;
const EMAIL_PLACEHOLDER = 'name@example.com';
const DEFAULT_PHONE_COUNTRY_CODE = '+852';
const DEFAULT_ACCOUNT_TYPE = 'personal';
const SUPPORTED_PHONE_COUNTRY_CODES = ['+852', '+86'] as const;

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
  confirmPassword: '',
  username: '',
  engName: '',
  chiName: '',
  phone: '',
  phoneCountryCode: DEFAULT_PHONE_COUNTRY_CODE,
  publisherIdentityType: DEFAULT_ACCOUNT_TYPE,
  otp: '',
  ismartAccount: '',
  primaryCommunityID: '',
  residenceFloor: '',
  residenceUnit: '',
  idCard: '',
  shouldBindResidence: false,
});

// 2. 解析手提電話輸入
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

  const supportedCountryCode = SUPPORTED_PHONE_COUNTRY_CODES.find((countryCode) =>
    normalizedValue.startsWith(countryCode) && normalizedValue.length > countryCode.length,
  );
  if (supportedCountryCode) {
    return {
      phoneCountryCode: supportedCountryCode,
      phoneNumber: normalizedValue.slice(supportedCountryCode.length).replace(/\D/g, ''),
    };
  }

  if (parts.length >= 2 && /^\+?\d{1,4}$/.test(parts[0])) {
    const phoneCountryCode = parts[0].startsWith('+') ? parts[0] : `+${parts[0]}`;
    if (!SUPPORTED_PHONE_COUNTRY_CODES.includes(phoneCountryCode as (typeof SUPPORTED_PHONE_COUNTRY_CODES)[number])) {
      return {
        phoneCountryCode: '',
        phoneNumber: '',
      };
    }

    return {
      phoneCountryCode,
      phoneNumber: parts.slice(1).join('').replace(/\D/g, ''),
    };
  }

  if (normalizedValue.startsWith('+')) {
    return {
      phoneCountryCode: '',
      phoneNumber: '',
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

// 2.1 標準化手提電話區號
const normalizePhoneCountryCode = (value: string): string =>
  SUPPORTED_PHONE_COUNTRY_CODES.includes(value as (typeof SUPPORTED_PHONE_COUNTRY_CODES)[number])
    ? value
    : DEFAULT_PHONE_COUNTRY_CODE;

// 2.2 解析手提電話表單輸入
export const parsePhoneFormInput = (phoneCountryCode: string, rawValue: string): ParsedPhoneInput => {
  const trimmedValue = rawValue.trim();
  if (trimmedValue.startsWith('+')) {
    return parsePhoneInput(trimmedValue);
  }

  const normalizedCountryCode = normalizePhoneCountryCode(phoneCountryCode);
  const countryDigits = normalizedCountryCode.replace(/^\+/, '');
  const phoneDigits = trimmedValue.replace(/\D/g, '');
  const normalizedPhoneNumber =
    phoneDigits.startsWith(countryDigits) && phoneDigits.length > countryDigits.length + 4
      ? phoneDigits.slice(countryDigits.length)
      : phoneDigits;

  return parsePhoneInput(`${normalizedCountryCode} ${normalizedPhoneNumber}`);
};

// 2.3 檢查手提電話格式是否符合後端要求
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

// 3. 檢查電郵格式
const isValidEmailInput = (email: string): boolean => {
  const value = email.trim();
  return value.includes('@') && value.includes('.') && value.length <= 255;
};

// 4. 組合錯誤訊息
const readErrorMessage = (error: unknown): string =>
  axios.isAxiosError(error)
    ? error.response?.data?.message ?? error.message
    : 'Request failed.';

// 5. 檢查英文姓名格式
const isValidEngNameInput = (value: string): boolean => {
  const engName = value.trim();
  return engName.length >= 2 && engName.length <= 120;
};

// 4.1 檢查用戶名稱格式
const isValidUsernameInput = (value: string): boolean => {
  const username = value.trim();
  return username.length >= 2 && username.length <= 120 && !/\s/.test(username);
};

// 5.1 逐項判斷註冊資料錯誤，避免使用無法定位問題的合併提示
export const resolveRegistrationValidationError = (
  email: string,
  username: string,
  engName: string,
  phoneCountryCode: string,
  phone: string,
  password: string,
  confirmPassword: string,
  shouldBindResidence = false,
  primaryCommunityID = '',
  residenceFloor = '',
  residenceUnit = '',
  idCard = '',
): RegistrationValidationErrorKey | null => {
  if (!isValidUsernameInput(username)) {
    return 'auth.registerUsernameInvalid';
  }

  if (!isValidEmailInput(email)) {
    return 'auth.invalidEmail';
  }

  if (shouldBindResidence && !isValidEngNameInput(engName)) {
    return 'auth.registerEnglishNameInvalid';
  }

  if (shouldBindResidence && !idCard.trim()) {
    return 'auth.idCardRequired';
  }

  if (!isValidParsedPhone(parsePhoneFormInput(phoneCountryCode, phone))) {
    return 'auth.invalidPhone';
  }

  if (password.trim().length < 8) {
    return 'auth.registerPasswordTooShort';
  }

  if (password !== confirmPassword) {
    return 'auth.registerPasswordMismatch';
  }

  if (shouldBindResidence && (!primaryCommunityID.trim() || !residenceFloor.trim() || !residenceUnit.trim())) {
    return 'auth.residenceBindingRequired';
  }

  return null;
};

// 5.2 取得註冊欄位的完整校驗結果
export const resolveRegistrationValidationErrors = (
  email: string,
  username: string,
  engName: string,
  phoneCountryCode: string,
  phone: string,
  password: string,
  confirmPassword: string,
  shouldBindResidence = false,
  primaryCommunityID = '',
  residenceFloor = '',
  residenceUnit = '',
  idCard = '',
): LoginValidationErrors => {
  const errors: LoginValidationErrors = {};

  if (!isValidUsernameInput(username)) {
    errors.username = 'auth.registerUsernameInvalid';
  }
  if (!isValidEmailInput(email)) {
    errors.email = 'auth.invalidEmail';
  }
  if (shouldBindResidence && !isValidEngNameInput(engName)) {
    errors.engName = 'auth.registerEnglishNameInvalid';
  }
  if (shouldBindResidence && !idCard.trim()) {
    errors.idCard = 'auth.idCardRequired';
  }
  if (!isValidParsedPhone(parsePhoneFormInput(phoneCountryCode, phone))) {
    errors.phone = 'auth.invalidPhone';
  }
  if (password.trim().length < 8) {
    errors.password = 'auth.registerPasswordTooShort';
  }
  if (password !== confirmPassword) {
    errors.confirmPassword = 'auth.registerPasswordMismatch';
  }
  if (shouldBindResidence && (!primaryCommunityID.trim() || !residenceFloor.trim() || !residenceUnit.trim())) {
    errors.residence = 'auth.residenceBindingRequired';
  }

  return errors;
};

// 5.3 取得統一登入欄位的完整校驗結果
export const resolveLoginValidationErrors = (
  accountInput: string,
  password: string,
): LoginValidationErrors => {
  const errors: LoginValidationErrors = {};

  if (accountInput.trim() === '') {
    errors.ismartAccount = 'auth.accountRequired';
  }

  if (password.trim() === '') {
    errors.password = 'auth.passwordRequired';
  }

  return errors;
};

// 8. 取得隨機登入頁主視覺
const getRandomHeroImage = (): LoginHeroImage => {
  const selectedIndex = Math.floor(Math.random() * LOGIN_HERO_IMAGES.length);
  return LOGIN_HERO_IMAGES[selectedIndex] ?? LOGIN_HERO_IMAGES[0];
};

// 9. 將後台登入背景圖轉為頁面主視覺列表
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

// 9. 讀取 POS 大廈 ID
const getBuildingId = (item: PosBuilding): string =>
  String(item.building_id ?? item.id ?? '').trim();

// 10. 讀取 POS 大廈名稱
const getBuildingName = (item: PosBuilding): string =>
  String(item.buildname_chi ?? item.buildname ?? item.name ?? getBuildingId(item)).trim();

// 11. 讀取 POS 單位 ID
const getUnitId = (item: PosBuildingUnit): string =>
  String(item.unit_id ?? item.id ?? '').trim();

// 12. 讀取 POS 單位樓層
const getRawUnitFloor = (item: PosBuildingUnit): string =>
  String(item.floor ?? '').trim();

// 13. 讀取 POS 單位名稱
const getRawUnitName = (item: PosBuildingUnit): string =>
  String(item.unit ?? item.unit_name ?? item.name ?? '').trim();

// 14. 讀取 POS 單位顯示名稱
const getUnitName = (item: PosBuildingUnit): string =>
  getRawUnitName(item) || getUnitId(item);

// 15. 取出數字字串
const digitsOnly = (value: unknown): string => String(value ?? '').replace(/\D/g, '');

// 16. 排除 POS 回傳的樓宇本身佔位資料
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

// 17. 取得樓層與單位排序分組
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

// 18. 依顯示規則排序樓層與單位
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

// 19. 管理登入頁資料與動作
export const useLoginPage = () => {
  const route = useRoute();
  const router = useRouter();
  const { t } = useI18n();
  const feedbackStore = useFeedbackStore();
  const sessionStore = useSessionStore();
  const formState = reactive(createInitialFormState());
  const emailAction = ref<LoginEmailAction>('login');
  const buildings = ref<PosBuilding[]>([]);
  const buildingUnits = ref<PosBuildingUnit[]>([]);
  const buildingsLoading = ref(false);
  const unitsLoading = ref(false);
  const submitting = ref(false);
  const rememberMe = ref(true);
  const validationErrors = ref<LoginValidationErrors>({});
  const selectedHero = ref(getRandomHeroImage());
  let latestUnitRequestID = 0;
  let buildingsRequested = false;

  const isAuthenticated = computed(() => sessionStore.isAuthenticated);

  // 20. 清除已修正表單的校驗提示
  watch(
    () => [
      formState.email,
      formState.username,
      formState.engName,
      formState.ismartAccount,
      formState.password,
      formState.confirmPassword,
      formState.phone,
      formState.phoneCountryCode,
      formState.shouldBindResidence,
      formState.primaryCommunityID,
      formState.residenceFloor,
      formState.residenceUnit,
      formState.idCard,
    ],
    () => {
      validationErrors.value = {};
    },
  );

  const submitLabel = computed(() => {
    if (submitting.value) {
      return t('auth.loading');
    }

    return emailAction.value === 'register' ? t('auth.registerSubmit') : t('auth.submit');
  });

  const emailActionSwitchLabel = computed(() =>
    emailAction.value === 'register' ? t('auth.emailLogin') : t('auth.emailRegister'),
  );

  const footerPrompt = computed(() =>
    emailAction.value === 'register' ? t('auth.alreadyHaveAccount') : t('auth.dontHaveAccount'),
  );

  const selectedAccountInput = computed(() => formState.ismartAccount.trim());
  const selectedBuildingName = computed(() =>
    buildingOptions.value.find((option) => option.value === formState.primaryCommunityID)?.label ?? '',
  );
  const publisherIdentityOptions = computed<LoginSelectOption[]>(() => [
    { label: t('auth.accountTypePersonal'), value: 'personal' },
    { label: t('auth.accountTypeIndividualAgent'), value: 'individual_agent' },
    { label: t('auth.accountTypeAgencyCompany'), value: 'agency_company' },
  ]);

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

  // 20. 重置樓層與單位級聯選擇
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

  // 21. 代理帳戶不綁定住戶大廈及單位
  watch(
    () => formState.publisherIdentityType,
    (accountType) => {
      if (accountType !== 'personal') {
        formState.shouldBindResidence = false;
        formState.primaryCommunityID = '';
        formState.residenceFloor = '';
        formState.residenceUnit = '';
        buildingUnits.value = [];
      }
    },
  );

  // 21.1 大廈綁定只在用戶主動勾選後載入並保留資料
  watch(
    () => formState.shouldBindResidence,
    (shouldBindResidence) => {
      if (shouldBindResidence) {
        void ensureBuildingsLoaded();
        return;
      }

      formState.primaryCommunityID = '';
      formState.residenceFloor = '';
      formState.residenceUnit = '';
      buildingUnits.value = [];
    },
  );

  // 22. 重置單位選擇
  watch(
    () => formState.residenceFloor,
    (nextValue, previousValue) => {
      if (nextValue !== previousValue) {
        formState.residenceUnit = '';
      }
    },
  );

  // 22. 載入後台登入背景圖
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

  // 23. 載入用戶註冊大廈選項
  const loadBuildings = async (): Promise<boolean> => {
    buildingsLoading.value = true;
    try {
      buildings.value = await fetchPosBuildings();
      return true;
    } catch (error) {
      buildings.value = [];
      feedbackStore.pushToast(readErrorMessage(error), 'error');
      return false;
    } finally {
      buildingsLoading.value = false;
    }
  };

  // 24. 按需載入用戶註冊大廈選項
  const ensureBuildingsLoaded = async (): Promise<void> => {
    if (buildingsRequested || buildingsLoading.value) {
      return;
    }

    buildingsRequested = true;
    const loaded = await loadBuildings();
    if (!loaded) {
      buildingsRequested = false;
    }
  };

  // 25. 載入指定大廈單位清單
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

  // 25. 完成登入後跳轉
  const redirectAfterSignIn = async (): Promise<void> => {
    const restrictedStatuses = ['pending_profile', 'pending_review', 'rejected'];
    const redirect = restrictedStatuses.includes(sessionStore.me?.member_status ?? '')
      ? '/account/profile/agency-profile'
      : typeof route.query.redirect === 'string' ? route.query.redirect : '/';
    await router.push(redirect);
  };

  // 26. 執行用戶註冊
  const handleRegisterSubmit = async (): Promise<void> => {
    validationErrors.value = resolveRegistrationValidationErrors(
      formState.email,
      formState.username,
      formState.engName,
      formState.phoneCountryCode,
      formState.phone,
      formState.password,
      formState.confirmPassword,
      formState.shouldBindResidence,
      formState.primaryCommunityID,
      formState.residenceFloor,
      formState.residenceUnit,
    );
    if (Object.keys(validationErrors.value).length > 0) {
      feedbackStore.pushToast(t('auth.formInvalid'), 'error');
      return;
    }

    submitting.value = true;

    try {
      const { phoneCountryCode, phoneNumber } = parsePhoneFormInput(formState.phoneCountryCode, formState.phone);
      const registrationPayload: RegisterEmailAccountPayload = {
        email: formState.email.trim(),
        password: formState.password,
        username: formState.username.trim(),
        eng_name: formState.shouldBindResidence ? formState.engName.trim() : formState.username.trim(),
        phone_country_code: phoneCountryCode,
        phone_number: phoneNumber,
        account_type: formState.publisherIdentityType,
      };
      const chiName = formState.chiName.trim();
      const idCard = formState.idCard.trim();
      if (chiName) {
        registrationPayload.chi_name = chiName;
      }
      if (idCard) {
        registrationPayload.id_card = idCard;
      }
      if (formState.shouldBindResidence) {
        registrationPayload.primary_community_id = formState.primaryCommunityID.trim();
        registrationPayload.primary_community_name = selectedBuildingName.value.trim();
        registrationPayload.residence_floor = formState.residenceFloor.trim();
        registrationPayload.residence_unit = formState.residenceUnit.trim();
      }
      await sessionStore.registerEmailAccount(registrationPayload);
      feedbackStore.pushToast(t('auth.registerSuccess'), 'success');
      await redirectAfterSignIn();
    } catch (error) {
      feedbackStore.pushToast(readErrorMessage(error), 'error');
    } finally {
      submitting.value = false;
    }
  };

  // 27. 執行統一帳戶登入
  const handleAccountSubmit = async (): Promise<void> => {
    const accountInput = selectedAccountInput.value;
    validationErrors.value = resolveLoginValidationErrors(accountInput, formState.password);
    if (Object.keys(validationErrors.value).length > 0) {
      feedbackStore.pushToast(t('auth.formInvalid'), 'error');
      return;
    }

    submitting.value = true;

    try {
      await sessionStore.signInWithIdentifier(accountInput, formState.password);
      feedbackStore.pushToast(t('auth.signInSuccess'), 'success');
      await redirectAfterSignIn();
    } catch (error) {
      feedbackStore.pushToast(readErrorMessage(error), 'error');
    } finally {
      submitting.value = false;
    }
  };

  // 28. 按目前操作提交登入或註冊表單
  const handleSubmit = async (): Promise<void> => {
    if (emailAction.value === 'register') {
      await handleRegisterSubmit();
      return;
    }

    await handleAccountSubmit();
  };

  // 29. 執行登出
  const handleSignOut = async (): Promise<void> => {
    await sessionStore.signOut();
    feedbackStore.pushToast(t('auth.signOutSuccess'), 'success');
  };

  // 30. 前往忘記密碼流程
  const handleForgotPassword = async (): Promise<void> => {
    await router.push('/forgot-password');
  };

  // 31. 切換登入與註冊模式
  const toggleEmailAction = (): void => {
    emailAction.value = emailAction.value === 'register' ? 'login' : 'register';
    formState.otp = '';
    validationErrors.value = {};
  };

  onMounted(() => {
    void loadConfiguredHero();
  });

  return {
    buildingOptions,
    buildingsLoading,
    emailAction,
    emailActionSwitchLabel,
    emailPlaceholder: EMAIL_PLACEHOLDER,
    footerPrompt,
    formState,
    handleForgotPassword,
    handleSignOut,
    handleSubmit,
    isAuthenticated,
    rememberMe,
    publisherIdentityOptions,
    residenceFloorOptions,
    residenceUnitOptions,
    selectedHero,
    submitting,
    submitLabel,
    ensureBuildingsLoaded,
    toggleEmailAction,
    unitsLoading,
    validationErrors,
  };
};

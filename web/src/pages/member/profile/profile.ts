/*
 * 會員中心個人資料 - 狀態與資料流程。
 * 1. 讀取並同步會員資料表單。
 * 2. 上傳 OSS 頭像並同步會員狀態。
 * 3. 儲存會員資料並同步全域會員狀態。
 */
import axios from 'axios';
import { computed, onMounted, reactive, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

import { fetchPosBuildings, fetchPosBuildingUnits } from '@/domains/building/api';
import { bindCurrentUserIsmart } from '@/domains/account/api';
import { requestEmailOtp } from '@/domains/account/api';
import { updateMe } from '@/domains/account/api';
import { completeUpload, createUploadPresign } from '@/domains/media/uploads-api';
import type { PosBuilding, PosBuildingUnit } from '@/domains/building/model';
import { useFeedbackStore } from '@/app/stores/feedback';
import { useSessionStore } from '@/app/stores/session';

interface ProfileInfoRow {
  key: string;
  label: string;
  value: string;
}

const avatarObjectPrefix = 'ajo_living/account/';
const maxAvatarFileSize = 5 * 1024 * 1024;
const blockedUploadHeaders = new Set(['host', 'content-length']);

// 1.0 讀取 POS 大廈 ID
const getBuildingId = (item: PosBuilding): string =>
  String(item.building_id ?? item.id ?? '').trim();

// 1.1 讀取 POS 大廈名稱
const getBuildingName = (item: PosBuilding): string =>
  String(item.buildname_chi ?? item.buildname ?? item.name ?? getBuildingId(item)).trim();

// 1.2 讀取 POS 單位 ID
const getUnitId = (item: PosBuildingUnit): string =>
  String(item.unit_id ?? item.id ?? '').trim();

// 1.3 讀取 POS 單位樓層
const getRawUnitFloor = (item: PosBuildingUnit): string =>
  String(item.floor ?? '').trim();

// 1.4 讀取 POS 單位名稱
const getRawUnitName = (item: PosBuildingUnit): string =>
  String(item.unit ?? item.unit_name ?? item.name ?? '').trim();

// 1.5 讀取 POS 單位顯示標籤
const getUnitLabel = (item: PosBuildingUnit): string => {
  const floor = getRawUnitFloor(item);
  const unit = getRawUnitName(item);
  return [floor, unit].filter(Boolean).join(' / ');
};

// 1.5.1 正規化 POS 權限 ID 列表
const normalizeTextList = (values: string[] | undefined): string[] => {
  const result: string[] = [];
  values?.forEach((item) => {
    String(item ?? '')
      .split(/[,，\n\r]+/)
      .forEach((part) => {
        const value = part.trim();
        if (value && !result.includes(value)) {
          result.push(value);
        }
      });
  });
  return result;
};

// 1.5.2 顯示大廈權限摘要
const formatBuildingListDisplay = (values: string[], isStaff: boolean): string => {
  const names = values.filter((value) => !/^\d+$/.test(value.trim()));
  if (names.length === 0) {
    return '';
  }
  if (isStaff) {
    return names.join(', ');
  }
  if (names.length > 3) {
    return `已綁定 ${names.length} 個大廈`;
  }
  return names.join(', ');
};

// 1.5.3 顯示單位權限摘要
const formatUnitListDisplay = (values: string[]): string => {
  if (values.length === 0) {
    return '';
  }
  if (values.length > 3) {
    return `已綁定 ${values.length} 個單位`;
  }
  return values.join(', ');
};

// 1.5.4 從 POS 單位 ID 推導大廈 ID
const getUnitBuildingID = (unitID: string): string =>
  digitsOnly(unitID).slice(0, 7);

// 1.6 取出數字字串
const digitsOnly = (value: unknown): string => String(value ?? '').replace(/\D/g, '');

// 1.7 排除 POS 回傳的樓宇本身佔位資料
const isSelectableUnit = (buildingID: string, item: PosBuildingUnit): boolean => {
  const normalizedBuildingID = digitsOnly(buildingID).slice(0, 7);
  const normalizedUnitID = digitsOnly(getUnitId(item));
  const hasFloor = getRawUnitFloor(item).length > 0;
  const hasUnit = getRawUnitName(item).length > 0;

  if (normalizedBuildingID && normalizedUnitID === normalizedBuildingID && !hasFloor && !hasUnit) {
    return false;
  }

  return true;
};

// 1.8 取得樓層與單位排序分組
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

// 1.9 依 POS 顯示規則排序樓層與單位
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

// 1. 建立 OSS 直傳 headers
const buildUploadHeaders = (headers: Record<string, string>, mimeType: string): Headers => {
  const result = new Headers();

  Object.entries(headers).forEach(([key, value]) => {
    if (!blockedUploadHeaders.has(key.toLowerCase())) {
      result.set(key, value);
    }
  });

  if (!result.has('Content-Type')) {
    result.set('Content-Type', mimeType);
  }

  return result;
};

// 2. 管理個人資料頁資料與動作
export const useAccountProfilePage = () => {
  const { t } = useI18n();
  const router = useRouter();
  const feedbackStore = useFeedbackStore();
  const sessionStore = useSessionStore();
  const isSaving = ref(false);
  const isLoading = ref(false);
  const isEditModalOpen = ref(false);
  const isIsmartModalOpen = ref(false);
  const isUploadingAvatar = ref(false);
  const isSigningOut = ref(false);
  const isRequestingEmailOTP = ref(false);
  const isBindingIsmart = ref(false);
  const buildings = ref<PosBuilding[]>([]);
  const buildingUnits = ref<PosBuildingUnit[]>([]);
  const buildingsLoading = ref(false);
  const unitsLoading = ref(false);

  const formState = reactive({
    display_name: '',
    email: '',
    email_otp_code: '',
    phone_country_code: '+852',
    phone_number: '',
    password: '',
    district_code: '',
  });
  const ismartFormState = reactive({
    account: '',
    password: '',
  });

  // 2.1 判斷是否為系統佔位電話帳號
  const isSystemPhonePlaceholder = (countryCode?: string): boolean =>
    ['email', 'ismart'].includes((countryCode ?? '').trim().toLowerCase());

  // 2.2 計算可提交的電話欄位
  const buildPhonePayload = (): { phone_country_code?: string; phone_number?: string } => {
    if (isSystemPhonePlaceholder(formState.phone_country_code)) {
      return {};
    }

    const phoneNumber = formState.phone_number.trim();
    if (!phoneNumber) {
      return {};
    }

    return {
      phone_country_code: formState.phone_country_code.trim() || '+852',
      phone_number: phoneNumber,
    };
  };

  // 2.3 讀取 API 錯誤訊息
  const readErrorMessage = (error: unknown, fallback: string): string =>
    axios.isAxiosError<{ message?: string }>(error)
      ? error.response?.data?.message ?? fallback
      : fallback;

  // 2.4 輸出缺省顯示文案
  const fallbackValue = computed(() => t('marketplace.myProfile.emptyValue'));

  // 2.4.1 判斷郵箱是否被修改
  const hasEmailChanged = computed(() =>
    formState.email.trim().toLowerCase() !== (sessionStore.me?.email ?? '').trim().toLowerCase(),
  );

  // 2.5 顯示帳戶電話資料
  const phoneDisplay = computed(() => {
    if (isSystemPhonePlaceholder(sessionStore.me?.phone_country_code)) {
      return t('account.profile.phoneUnavailable');
    }

    const countryCode = formState.phone_country_code.trim();
    const phoneNumber = formState.phone_number.trim();
    return phoneNumber ? `${countryCode || '+852'} ${phoneNumber}`.trim() : t('account.profile.phoneUnavailable');
  });

  // 2.5.1 顯示 ismart 綁定狀態
  const ismartStatusDisplay = computed(() =>
    sessionStore.me?.ismart_linked
      ? sessionStore.me.ismart_username || sessionStore.me.ismart_msg?.username || fallbackValue.value
      : t('account.profile.ismartUnlinked'),
  );

  // 2.6 顯示本地主要屋苑資料
  const profileCommunityDisplay = computed(() => {
    const community = sessionStore.me?.primary_community;
    const rawValue =
      community?.name_zh?.trim() ||
      community?.name_en?.trim() ||
      community?.address_text?.trim() ||
      '';
    const normalizedValues = normalizeTextList(rawValue ? [rawValue] : []);
    if (normalizedValues.length > 1) {
      return formatBuildingListDisplay(normalizedValues, false) || t('marketplace.myProfile.noCommunity');
    }

    return rawValue || t('marketplace.myProfile.noCommunity');
  });

  // 2.6.1 判斷 POS Staff 身份
  const isPOSStaff = computed(() =>
    Boolean(sessionStore.me?.is_staff || sessionStore.me?.ismart_msg?.is_staff),
  );

  // 2.6.2 取得 POS 權限大廈
  const permissionBuildingIDs = computed(() => {
    const message = sessionStore.me?.ismart_msg;
    if (message) {
      if (isPOSStaff.value) {
        const staffBuildings = normalizeTextList(message.staff_building_permissions);
        return staffBuildings.length > 0 ? staffBuildings : normalizeTextList(message.building);
      }

      return normalizeTextList(message.client_building_permissions);
    }

    return normalizeTextList(sessionStore.me?.bound_building_ids);
  });

  // 2.6.3 取得 POS 權限單位
  const permissionUnitIDs = computed(() => {
    if (isPOSStaff.value) {
      return [];
    }

    const message = sessionStore.me?.ismart_msg;
    return normalizeTextList(
      message?.client_building_flat_units_permissions ?? sessionStore.me?.bound_flat_unit_ids,
    );
  });

  // 2.6.4 建立 POS 大廈名稱索引
  const buildingNameMap = computed(() => {
    const map = new Map<string, string>();
    buildings.value.forEach((item) => {
      const id = getBuildingId(item);
      const name = getBuildingName(item);
      if (id && name) {
        map.set(id, name);
        map.set(digitsOnly(id), name);
        map.set(name, name);
      }
    });

    const community = sessionStore.me?.primary_community;
    const communityID = community?.public_id?.trim() ?? '';
    if (communityID && profileCommunityDisplay.value !== t('marketplace.myProfile.noCommunity')) {
      map.set(communityID, profileCommunityDisplay.value);
    }

    return map;
  });

  // 2.6.5 顯示 POS 權限大廈名稱
  const buildingPermissionsDisplay = computed(() => {
    const names: string[] = [];
    permissionBuildingIDs.value.forEach((buildingID) => {
      const value = buildingID.trim();
      const resolved =
        buildingNameMap.value.get(value) ||
        buildingNameMap.value.get(digitsOnly(value)) ||
        (/^\d+$/.test(value) ? '' : value);
      if (resolved && !names.includes(resolved)) {
        names.push(resolved);
      }
    });

    return formatBuildingListDisplay(names, isPOSStaff.value);
  });

  // 2.6.6 顯示主要屋苑資料
  const communityDisplay = computed(() =>
    buildingPermissionsDisplay.value || profileCommunityDisplay.value,
  );

  // 2.6.7 建立 POS 單位名稱索引
  const unitLabelMap = computed(() => {
    const map = new Map<string, string>();
    buildingUnits.value.forEach((item) => {
      const unitID = getUnitId(item);
      if (!unitID) {
        return;
      }

      const buildingName = buildingNameMap.value.get(getUnitBuildingID(unitID)) ?? '';
      const unitLabel = getUnitLabel(item);
      map.set(unitID, [buildingName, unitLabel].filter(Boolean).join(' / '));
    });
    return map;
  });

  // 2.7 顯示居住單位資料
  const residenceUnitDisplay = computed(() => {
    if (isPOSStaff.value) {
      return '';
    }

    const unitLabels = permissionUnitIDs.value
      .map((unitID) => unitLabelMap.value.get(unitID) ?? '')
      .filter(Boolean)
      .sort(compareDisplayCodes);
    if (unitLabels.length > 0) {
      return formatUnitListDisplay(unitLabels);
    }

    const floor = sessionStore.me?.residence_floor?.trim() ?? '';
    const unit = sessionStore.me?.residence_unit?.trim() ?? '';
    if (!floor && !unit) {
      return fallbackValue.value;
    }

    return [floor, unit].filter(Boolean).join(' / ');
  });

  // 2.7 顯示帳戶資料行
  const profileRows = computed<ProfileInfoRow[]>(() => {
    const rows: ProfileInfoRow[] = [
      {
        key: 'display_name',
        label: t('account.profile.displayName'),
        value: sessionStore.me?.display_name?.trim() || sessionStore.currentUser.display_name || fallbackValue.value,
      },
      {
        key: 'ismart_status',
        label: t('account.profile.ismartAccount'),
        value: ismartStatusDisplay.value,
      },
      {
        key: 'email',
        label: t('account.profile.email'),
        value: formState.email || t('account.profile.emailUnavailable'),
      },
      {
        key: 'phone',
        label: t('account.profile.phone'),
        value: phoneDisplay.value,
      },
      {
        key: 'local_password',
        label: t('account.profile.localPassword'),
        value: sessionStore.me?.local_password?.trim() || fallbackValue.value,
      },
      {
        key: 'district_code',
        label: t('account.profile.districtCode'),
        value: sessionStore.me?.district_code?.trim() || fallbackValue.value,
      },
    ];

    if (isPOSStaff.value) {
      rows.push({
        key: 'managed_buildings',
        label: t('account.profile.managedBuildings'),
        value: buildingPermissionsDisplay.value || fallbackValue.value,
      });
    } else {
      rows.push({
        key: 'resident_building',
        label: t('account.profile.residentBuilding'),
        value: communityDisplay.value || fallbackValue.value,
      });
      rows.push({
        key: 'residence_unit',
        label: t('account.profile.residentUnit'),
        value: residenceUnitDisplay.value,
      });
    }

    return rows;
  });

  // 2.8 顯示身份狀態資料行
  const accountRows = computed<ProfileInfoRow[]>(() => [
    {
      key: 'public_id',
      label: t('marketplace.myProfile.memberId'),
      value: sessionStore.me?.public_id?.trim() || fallbackValue.value,
    },
    {
      key: 'member_status',
      label: t('marketplace.myProfile.memberStatus'),
      value: sessionStore.me?.member_status?.trim() || fallbackValue.value,
    },
    {
      key: 'account_type',
      label: t('marketplace.myProfile.memberType'),
      value: sessionStore.me?.is_staff ? 'staff' : 'user',
    },
    {
      key: 'profile_completed',
      label: t('marketplace.myProfile.profileCompleted'),
      value: sessionStore.me?.profile_completed ? t('marketplace.myProfile.completed') : t('marketplace.myProfile.incomplete'),
    },
    {
      key: 'is_staff',
      label: t('marketplace.myProfile.staffAccess'),
      value: sessionStore.me?.is_staff ? t('marketplace.myProfile.enabled') : t('marketplace.myProfile.disabled'),
    },
  ]);

  // 2.9 同步表單內容
  const syncFormState = (): void => {
    formState.display_name = sessionStore.me?.display_name ?? sessionStore.currentUser.display_name;
    formState.email = sessionStore.me?.email ?? '';
    formState.email_otp_code = '';
    formState.phone_country_code = isSystemPhonePlaceholder(sessionStore.me?.phone_country_code)
      ? '+852'
      : sessionStore.me?.phone_country_code || '+852';
    formState.phone_number = isSystemPhonePlaceholder(sessionStore.me?.phone_country_code)
      ? ''
      : sessionStore.me?.phone_number ?? '';
    formState.password = sessionStore.me?.local_password ?? '';
    ismartFormState.account = sessionStore.me?.ismart_username ?? sessionStore.me?.ismart_msg?.username ?? '';
    ismartFormState.password = '';
    formState.district_code = sessionStore.me?.district_code ?? '';
  };

  // 2.10.1 載入 POS 大廈清單
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

  // 2.10.2 載入 POS 權限單位顯示資料
  const loadPermissionUnits = async (): Promise<void> => {
    const buildingIDs = permissionBuildingIDs.value;
    if (isPOSStaff.value || buildingIDs.length === 0) {
      buildingUnits.value = [];
      return;
    }

    unitsLoading.value = true;
    try {
      const result = await Promise.all(
        buildingIDs.map(async (buildingID) => {
          const units = await fetchPosBuildingUnits(buildingID);
          return units.filter((item) => isSelectableUnit(buildingID, item));
        }),
      );
      const merged = new Map<string, PosBuildingUnit>();
      result.flat().forEach((item) => {
        const unitID = getUnitId(item);
        if (unitID) {
          merged.set(unitID, item);
        }
      });
      buildingUnits.value = Array.from(merged.values());
    } catch (error) {
      buildingUnits.value = [];
      feedbackStore.pushToast(readErrorMessage(error, t('account.profile.loadError')), 'error');
    } finally {
      unitsLoading.value = false;
    }
  };

  // 2.11 讀取會員資料
  const loadProfile = async (): Promise<void> => {
    if (sessionStore.me) {
      syncFormState();
      return;
    }

    isLoading.value = true;

    try {
      await sessionStore.loadCurrentUser();
      syncFormState();
    } catch (error) {
      console.error(error);
      feedbackStore.pushToast(readErrorMessage(error, t('account.profile.loadError')), 'error');
    } finally {
      isLoading.value = false;
    }
  };

  // 2.12 開啟編輯彈窗
  const openEditModal = (): void => {
    syncFormState();
    isEditModalOpen.value = true;
  };

  // 2.13 關閉編輯彈窗
  const closeEditModal = (): void => {
    if (!isSaving.value && !isUploadingAvatar.value && !isRequestingEmailOTP.value) {
      isEditModalOpen.value = false;
    }
  };

  // 2.12.1 開啟 ismart 綁定彈窗
  const openIsmartModal = (): void => {
    ismartFormState.account = sessionStore.me?.ismart_username ?? sessionStore.me?.ismart_msg?.username ?? '';
    ismartFormState.password = '';
    isIsmartModalOpen.value = true;
  };

  // 2.12.2 關閉 ismart 綁定彈窗
  const closeIsmartModal = (): void => {
    if (!isBindingIsmart.value) {
      isIsmartModalOpen.value = false;
    }
  };

  // 2.13.1 請求郵箱修改驗證碼
  const handleRequestEmailOTP = async (): Promise<void> => {
    const email = formState.email.trim();
    if (!email || !email.includes('@') || !email.includes('.')) {
      feedbackStore.pushToast(t('auth.invalidEmail'), 'error');
      return;
    }
    if (!hasEmailChanged.value) {
      feedbackStore.pushToast(t('account.profile.emailUnchanged'), 'info');
      return;
    }

    isRequestingEmailOTP.value = true;
    try {
      const { data } = await requestEmailOtp({
        email,
        scene: 'profile_email_update',
      });
      feedbackStore.pushToast(
        data.data.mock_code ? t('auth.otpPreview', { code: data.data.mock_code }) : t('auth.otpSent'),
        'info',
      );
    } catch (error) {
      feedbackStore.pushToast(readErrorMessage(error, t('account.profile.emailOtpError')), 'error');
    } finally {
      isRequestingEmailOTP.value = false;
    }
  };

  // 2.14 上傳頭像並更新會員資料
  const handleAvatarFileChange = async (event: Event): Promise<void> => {
    const input = event.target as HTMLInputElement;
    const file = input.files?.[0];
    input.value = '';

    if (!file) {
      return;
    }
    if (!file.type.startsWith('image/')) {
      feedbackStore.pushToast(t('account.profile.avatarFileInvalid'), 'error');
      return;
    }
    if (file.size > maxAvatarFileSize) {
      feedbackStore.pushToast(t('account.profile.avatarFileTooLarge'), 'error');
      return;
    }
    if (hasEmailChanged.value && formState.email_otp_code.trim().length === 0) {
      feedbackStore.pushToast(t('account.profile.emailOtpRequired'), 'error');
      return;
    }

    isUploadingAvatar.value = true;

    try {
      const presignResponse = await createUploadPresign({
        file_name: file.name,
        mime_type: file.type,
        file_size: file.size,
        object_prefix: avatarObjectPrefix,
      });
      const presign = presignResponse.data.data;

      const uploadResponse = await fetch(presign.upload_url, {
        method: 'PUT',
        headers: buildUploadHeaders(presign.headers, file.type),
        body: file,
      });
      if (!uploadResponse.ok) {
        throw new Error(`avatar upload failed with status ${uploadResponse.status}`);
      }

      const completeResponse = await completeUpload({
        object_key: presign.object_key,
        mime_type: file.type,
        file_size: file.size,
      });
      const profileResponse = await updateMe({
        display_name: formState.display_name.trim(),
        email: formState.email.trim(),
        email_otp_code: hasEmailChanged.value ? formState.email_otp_code.trim() : '',
        ...buildPhonePayload(),
        password: formState.password.trim(),
        primary_community_id: sessionStore.me?.primary_community?.public_id ?? '',
        primary_community_name: profileCommunityDisplay.value === t('marketplace.myProfile.noCommunity')
          ? ''
          : profileCommunityDisplay.value,
        residence_floor: sessionStore.me?.residence_floor ?? '',
        residence_unit: sessionStore.me?.residence_unit ?? '',
        district_code: formState.district_code.trim(),
        avatar_asset_id: completeResponse.data.data.media_asset_id,
      });

      sessionStore.me = profileResponse.data.data;
      syncFormState();
      feedbackStore.pushToast(t('account.profile.avatarUploadSuccess', { points: 50 }), 'success');
    } catch (error) {
      console.error(error);
      feedbackStore.pushToast(readErrorMessage(error, t('account.profile.avatarUploadError')), 'error');
    } finally {
      isUploadingAvatar.value = false;
    }
  };

  // 2.15 儲存會員資料
  const handleSaveProfile = async (): Promise<void> => {
    if (hasEmailChanged.value) {
      const email = formState.email.trim();
      if (!email || !email.includes('@') || !email.includes('.')) {
        feedbackStore.pushToast(t('auth.invalidEmail'), 'error');
        return;
      }
      if (formState.email_otp_code.trim().length === 0) {
        feedbackStore.pushToast(t('account.profile.emailOtpRequired'), 'error');
        return;
      }
    }

    isSaving.value = true;

    try {
      const { data } = await updateMe({
        display_name: formState.display_name.trim(),
        email: formState.email.trim(),
        email_otp_code: hasEmailChanged.value ? formState.email_otp_code.trim() : '',
        ...buildPhonePayload(),
        password: formState.password.trim(),
        primary_community_id: sessionStore.me?.primary_community?.public_id ?? '',
        primary_community_name: profileCommunityDisplay.value === t('marketplace.myProfile.noCommunity')
          ? ''
          : profileCommunityDisplay.value,
        residence_floor: sessionStore.me?.residence_floor ?? '',
        residence_unit: sessionStore.me?.residence_unit ?? '',
        district_code: formState.district_code.trim(),
      });

      sessionStore.me = data.data;
      syncFormState();
      isEditModalOpen.value = false;
      feedbackStore.pushToast(t('account.profile.updateSuccess'), 'success');
    } catch (error) {
      console.error(error);
      feedbackStore.pushToast(readErrorMessage(error, t('account.profile.updateError')), 'error');
    } finally {
      isSaving.value = false;
    }
  };

  // 2.15.1 綁定 ismart 帳戶
  const handleBindIsmart = async (): Promise<void> => {
    const account = ismartFormState.account.trim();
    const password = ismartFormState.password.trim();
    if (!account || !password) {
      feedbackStore.pushToast(t('auth.ismartRequiredFields'), 'error');
      return;
    }

    isBindingIsmart.value = true;
    try {
      const { data } = await bindCurrentUserIsmart({
        account,
        password,
        phone: formState.phone_number.trim() || undefined,
        email: formState.email.trim() || undefined,
      });
      sessionStore.me = data.data;
      syncFormState();
      void loadPermissionUnits();
      isIsmartModalOpen.value = false;
      feedbackStore.pushToast(t('account.profile.ismartBindSuccess'), 'success');
    } catch (error) {
      feedbackStore.pushToast(readErrorMessage(error, t('account.profile.ismartBindError')), 'error');
    } finally {
      isBindingIsmart.value = false;
    }
  };

  // 2.16 執行登出
  const handleSignOut = async (): Promise<void> => {
    isSigningOut.value = true;

    try {
      await sessionStore.signOut();
      await router.push('/login');
    } finally {
      isSigningOut.value = false;
    }
  };

  // 2.17 初始化會員資料
  onMounted(async () => {
    await loadProfile();
    await loadBuildings();
    await loadPermissionUnits();
  });

  return {
    accountRows,
    buildingsLoading,
    closeEditModal,
    closeIsmartModal,
    communityDisplay,
    formState,
    handleAvatarFileChange,
    handleRequestEmailOTP,
    handleBindIsmart,
    handleSaveProfile,
    handleSignOut,
    isEditModalOpen,
    isIsmartModalOpen,
    isLoading,
    isRequestingEmailOTP,
    isBindingIsmart,
    ismartFormState,
    isSaving,
    isSigningOut,
    isUploadingAvatar,
    openEditModal,
    openIsmartModal,
    phoneDisplay,
    profileRows,
    hasEmailChanged,
    sessionStore,
    t,
    unitsLoading,
  };
};

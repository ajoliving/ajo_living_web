/*
 * 會員中心個人資料 - 狀態與資料流程。
 * 1. 讀取並同步會員資料表單。
 * 2. 上傳 OSS 頭像並同步會員狀態。
 * 3. 儲存會員資料並同步全域會員狀態。
 */
import axios from 'axios';
import { computed, onMounted, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

import { fetchMemberPosBuildings, fetchMemberPosBuildingUnits } from '@/httpapis/building';
import { bindCurrentUserIsmart, updateMe } from '@/httpapis/me';
import { completeUpload, createUploadPresign } from '@/httpapis/uploads';
import type { PosBuilding, PosBuildingUnit } from '@/model/community';
import { useFeedbackStore } from '@/stores/feedback';
import { useSessionStore } from '@/stores/session';

interface ProfileInfoRow {
  key: string;
  label: string;
  value: string;
}

interface ProfileUnitOption {
  label: string;
  value: string;
}

const avatarObjectPrefix = 'ajo_living/account/';
const maxAvatarFileSize = 5 * 1024 * 1024;
const blockedUploadHeaders = new Set(['host', 'content-length']);
const unassignedFloor = '未指定樓層';

// 1.0 讀取 POS 大廈 ID
const getBuildingID = (item: PosBuilding): string =>
  String(item.building_id ?? item.id ?? '').trim();

// 1.1 讀取 POS 大廈名稱
const getBuildingName = (item: PosBuilding): string =>
  String(item.buildname_chi ?? item.buildname ?? item.name ?? getBuildingID(item)).trim();

// 1.2 讀取 POS 單位 ID
const getUnitID = (item: PosBuildingUnit): string =>
  String(item.unit_id ?? item.id ?? '').trim();

// 1.3 讀取 POS 單位樓層
const getUnitFloor = (item: PosBuildingUnit): string =>
  String(item.floor ?? '').trim();

// 1.4 讀取 POS 單位樓層顯示值
const getDisplayUnitFloor = (item: PosBuildingUnit): string =>
  getUnitFloor(item) || unassignedFloor;

// 1.5 讀取 POS 單位名稱
const getUnitName = (item: PosBuildingUnit): string =>
  String(item.unit ?? item.unit_name ?? item.name ?? '').trim();

// 1.6 取出數字字串
const digitsOnly = (value: unknown): string =>
  String(value ?? '').replace(/\D/g, '');

// 1.7 建立 POS 大廈 ID 查找鍵
const buildBuildingIDKeys = (value: string): string[] => {
  const rawValue = value.trim();
  const result: string[] = [];
  if (rawValue) {
    result.push(rawValue);
  }
  if (!/^\d+$/.test(rawValue)) {
    return result;
  }

  const paddedValue = rawValue.padStart(7, '0');
  const unpaddedValue = rawValue.replace(/^0+/, '') || '0';
  [paddedValue, unpaddedValue].forEach((item) => {
    if (!result.includes(item)) {
      result.push(item);
    }
  });
  return result;
};

// 1.8 正規化 POS 權限 ID 列表
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

// 1.9 排除 POS 回傳的樓宇佔位資料
const isSelectableUnit = (buildingID: string, item: PosBuildingUnit): boolean => {
  const normalizedBuildingID = digitsOnly(buildingID).slice(0, 7);
  const normalizedUnitID = digitsOnly(getUnitID(item));
  const hasFloor = getUnitFloor(item).length > 0;
  const hasUnit = getUnitName(item).length > 0;

  return !(normalizedBuildingID && normalizedUnitID === normalizedBuildingID && !hasFloor && !hasUnit);
};

// 1.10 由 POS 權限 ID 建立可選單位
const buildUnitsFromFlatUnitPermissions = (
  buildingID: string,
  flatUnitPermissions: string[],
): PosBuildingUnit[] => {
  const normalizedBuildingID = digitsOnly(buildingID).slice(0, 7);
  if (!normalizedBuildingID || flatUnitPermissions.length === 0) {
    return [];
  }

  const seen = new Set<string>();
  const units: PosBuildingUnit[] = [];
  flatUnitPermissions.forEach((rawPermission) => {
    const permission = digitsOnly(rawPermission);
    let normalizedUnitID = '';
    if (permission.startsWith(normalizedBuildingID) && permission.length >= 11) {
      normalizedUnitID = permission.slice(0, 11);
    } else if (permission.startsWith(normalizedBuildingID.slice(0, 6)) && permission.length > 6) {
      const suffix = permission.slice(6);
      if (suffix.length > 0 && suffix.length <= 4) {
        normalizedUnitID = `${normalizedBuildingID}${suffix.padStart(4, '0')}`;
      }
    }
    if (!normalizedUnitID || seen.has(normalizedUnitID) || !normalizedUnitID.startsWith(normalizedBuildingID)) {
      return;
    }

    const floorCode = normalizedUnitID.slice(7, 9);
    const unitCode = normalizedUnitID.slice(9, 11);
    const unitName = unitCode.replace(/^0+/, '') || unitCode;
    seen.add(normalizedUnitID);
    units.push({
      unit_id: normalizedUnitID,
      floor: floorCode === '00' ? '' : floorCode,
      unit: unitName,
      unit_name: unitName,
    });
  });

  return units.sort((left, right) => getUnitID(left).localeCompare(getUnitID(right), 'en', { numeric: true }));
};

// 1.11 轉換 POS 權限碼
const toTwoDigitCode = (value: unknown): string => {
  const digits = digitsOnly(value);
  return digits ? digits.slice(-2).padStart(2, '0') : '';
};

// 1.12 取得單位權限碼
const getUnitPermissionCode = (buildingID: string, item: PosBuildingUnit): string => {
  const normalizedBuildingID = digitsOnly(buildingID).slice(0, 7);
  const unitID = digitsOnly(getUnitID(item));
  if (unitID.startsWith(normalizedBuildingID) && unitID.length >= 11) {
    return unitID.slice(0, 11);
  }

  const floorCode = toTwoDigitCode(getUnitFloor(item));
  const unitCode = toTwoDigitCode(getUnitName(item));
  return normalizedBuildingID && floorCode && unitCode ? `${normalizedBuildingID}${floorCode}${unitCode}` : '';
};

// 1.13 判斷單位是否符合 POS 權限
const matchesUnitPermission = (
  buildingID: string,
  item: PosBuildingUnit,
  flatUnitPermissions: string[],
): boolean => {
  if (flatUnitPermissions.length === 0) {
    return true;
  }

  const permissionCode = getUnitPermissionCode(buildingID, item);
  if (!permissionCode) {
    return false;
  }

  const normalizedBuildingID = digitsOnly(buildingID).slice(0, 7);
  const compactUnitID = permissionCode.replace(/0/g, '');
  return flatUnitPermissions.some((rawPermission) => {
    const permission = digitsOnly(rawPermission);
    if (!permission) {
      return false;
    }
    const compactPermission = permission.replace(/0/g, '');
    return (
      permission === normalizedBuildingID ||
      permissionCode === permission ||
      compactUnitID === compactPermission ||
      permissionCode.startsWith(permission) ||
      compactUnitID.startsWith(compactPermission)
    );
  });
};

// 1.14 按權限過濾 POS 單位
const filterUnitsByPermission = (
  buildingID: string,
  units: PosBuildingUnit[],
  flatUnitPermissions: string[],
): PosBuildingUnit[] =>
  units
    .filter((item) => isSelectableUnit(buildingID, item))
    .filter((item) => matchesUnitPermission(buildingID, item, flatUnitPermissions));

// 1.15 取得 POS 顯示排序分組
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

// 1.16 按 POS 顯示規則排序
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
  const isBindingIsmart = ref(false);
  const isBuildingsLoading = ref(false);
  const isUnitsLoading = ref(false);
  const isSavingUnit = ref(false);
  const buildings = ref<PosBuilding[]>([]);
  const selectedBuildingUnits = ref<PosBuildingUnit[]>([]);
  const memberVisibleUnits = ref<PosBuildingUnit[]>([]);
  const selectedBuildingID = ref('');
  const selectedFloor = ref('');
  const selectedUnitID = ref('');
  let isSyncingUnitSelection = false;

  const formState = reactive({
    display_name: '',
    email: '',
    phone_country_code: '+852',
    phone_number: '',
    district_code: '',
    publisher_identity_type: '',
  });
  const ismartFormState = reactive({
    account: '',
    password: '',
  });

  // 2.1 計算可提交的電話欄位
  const buildPhonePayload = (): { phone_country_code?: string; phone_number?: string } => {
    const phoneNumber = formState.phone_number.trim();
    if (!phoneNumber) {
      return {};
    }

    return {
      phone_country_code: formState.phone_country_code.trim() || '+852',
      phone_number: phoneNumber,
    };
  };

  // 2.2 讀取 API 錯誤訊息
  const readErrorMessage = (error: unknown, fallback: string): string =>
    axios.isAxiosError<{ message?: string }>(error)
      ? error.response?.data?.message ?? fallback
      : fallback;

  // 2.3 輸出缺省顯示文案
  const fallbackValue = computed(() => t('marketplace.myProfile.emptyValue'));

  // 2.3.1 判斷 POS Staff 身份
  const isPOSStaff = computed(() =>
    Boolean(sessionStore.me?.is_staff || sessionStore.me?.ismart_msg?.is_staff),
  );

  // 2.3.2 取得員工可管理大廈
  const staffBuildingIDs = computed(() => {
    const message = sessionStore.me?.ismart_msg;
    if (!message) {
      return normalizeTextList(sessionStore.me?.bound_building_ids);
    }

    const staffBuildings = normalizeTextList(message.staff_building_permissions);
    return staffBuildings.length > 0 ? staffBuildings : normalizeTextList(message.building);
  });

  // 2.3.3 取得住戶可見大廈
  const clientBuildingIDs = computed(() => {
    const message = sessionStore.me?.ismart_msg;
    const values = message
      ? normalizeTextList(message.client_building_permissions)
      : normalizeTextList(sessionStore.me?.bound_building_ids);
    const primaryCommunityID = sessionStore.me?.primary_community?.public_id?.trim() ?? '';
    return values.length > 0 || !primaryCommunityID ? values : [primaryCommunityID];
  });

  // 2.3.4 取得住戶可見單位
  const clientUnitIDs = computed(() => {
    const message = sessionStore.me?.ismart_msg;
    return message
      ? normalizeTextList(message.client_building_flat_units_permissions)
      : normalizeTextList(sessionStore.me?.bound_flat_unit_ids);
  });

  // 2.3.5 取得目前可選大廈
  const allowedBuildingIDs = computed(() =>
    isPOSStaff.value ? staffBuildingIDs.value : clientBuildingIDs.value,
  );

  // 2.3.6 建立 POS 大廈名稱索引
  const buildingNameMap = computed(() => {
    const map = new Map<string, string>();
    buildings.value.forEach((item) => {
      const id = getBuildingID(item);
      const name = getBuildingName(item);
      if (id && name) {
        buildBuildingIDKeys(id).forEach((key) => {
          map.set(key, name);
        });
      }
    });

    const community = sessionStore.me?.primary_community;
    const communityID = community?.public_id?.trim() ?? '';
    const communityName =
      community?.name_zh?.trim() ||
      community?.name_en?.trim() ||
      community?.address_text?.trim() ||
      '';
    if (communityID && communityName) {
      buildBuildingIDKeys(communityID).forEach((key) => {
        if (!map.has(key)) {
          map.set(key, communityName);
        }
      });
    }

    return map;
  });

  // 2.3.7 讀取大廈顯示名稱
  const resolveBuildingName = (buildingID: string): string => {
    const matchedKey = buildBuildingIDKeys(buildingID).find((key) => buildingNameMap.value.has(key));
    return matchedKey ? buildingNameMap.value.get(matchedKey) ?? buildingID : buildingID;
  };

  // 2.3.8 顯示員工管理屋苑
  const managedBuildingsDisplay = computed(() => {
    const names = staffBuildingIDs.value
      .map((buildingID) => resolveBuildingName(buildingID))
      .filter(Boolean);
    return names.length > 0 ? names.join(', ') : fallbackValue.value;
  });

  // 2.3.9 取得目前大廈可見單位
  const visibleSelectedUnits = computed(() => {
    if (isPOSStaff.value) {
      return selectedBuildingUnits.value;
    }

    return selectedBuildingUnits.value.filter((item) =>
      matchesUnitPermission(selectedBuildingID.value, item, clientUnitIDs.value),
    );
  });

  // 2.3.10 顯示住戶所屬大廈 / 樓層 / 單位
  const residentUnitDisplay = computed(() => {
    if (isPOSStaff.value) {
      return '';
    }

    const labels = memberVisibleUnits.value
      .filter((item) => matchesUnitPermission(digitsOnly(getUnitID(item)).slice(0, 7), item, clientUnitIDs.value))
      .map((item) => {
        const buildingID = digitsOnly(getUnitID(item)).slice(0, 7);
        return [
          resolveBuildingName(buildingID),
          getDisplayUnitFloor(item),
          getUnitName(item),
        ].filter(Boolean).join(' / ');
      })
      .filter(Boolean)
      .sort(compareDisplayCodes);

    if (labels.length > 0) {
      return labels.join(', ');
    }

    const community = sessionStore.me?.primary_community;
    const buildingName =
      community?.name_zh?.trim() ||
      community?.name_en?.trim() ||
      community?.address_text?.trim() ||
      '';
    const floor = sessionStore.me?.residence_floor?.trim() ?? '';
    const unit = sessionStore.me?.residence_unit?.trim() ?? '';
    return [buildingName, floor, unit].filter(Boolean).join(' / ') || fallbackValue.value;
  });

  // 2.3.11 建立綁定大廈選項
  const buildingOptions = computed<ProfileUnitOption[]>(() => [
    {
      label: isBuildingsLoading.value ? '載入中' : '請選擇大廈',
      value: '',
    },
    ...allowedBuildingIDs.value
      .map((buildingID) => ({
        label: resolveBuildingName(buildingID),
        value: buildingID,
      }))
      .filter((item, index, list) =>
        item.value && list.findIndex((candidate) => candidate.value === item.value) === index,
      )
      .sort((left, right) => compareDisplayCodes(left.label, right.label)),
  ]);

  // 2.3.12 建立綁定樓層選項
  const floorOptions = computed<ProfileUnitOption[]>(() => {
    const floors = Array.from(
      new Set(visibleSelectedUnits.value.map((item) => getDisplayUnitFloor(item)).filter(Boolean)),
    ).sort(compareDisplayCodes);

    return [
      {
        label: isUnitsLoading.value ? '載入中' : '請選擇樓層',
        value: '',
      },
      ...floors.map((floor) => ({
        label: floor,
        value: floor,
      })),
    ];
  });

  // 2.3.13 建立綁定單位選項
  const unitOptions = computed<ProfileUnitOption[]>(() => [
    {
      label: isUnitsLoading.value ? '載入中' : '請選擇單位',
      value: '',
    },
    ...visibleSelectedUnits.value
      .filter((item) => getDisplayUnitFloor(item) === selectedFloor.value)
      .slice()
      .sort((left, right) => compareDisplayCodes(getUnitName(left), getUnitName(right)))
      .map((item) => ({
        label: getUnitName(item),
        value: getUnitID(item),
      }))
      .filter((item) => item.label && item.value),
  ]);

  // 2.3.14 取得目前選中單位
  const selectedUnit = computed(() =>
    visibleSelectedUnits.value.find((item) => getUnitID(item) === selectedUnitID.value),
  );

  // 2.3.15 顯示目前選中單位
  const selectedUnitDisplay = computed(() => {
    const buildingName = selectedBuildingID.value
      ? buildingOptions.value.find((item) => item.value === selectedBuildingID.value)?.label ?? ''
      : '';
    const unitName = selectedUnit.value ? getUnitName(selectedUnit.value) : '';
    return [buildingName, selectedFloor.value, unitName].filter(Boolean).join(' / ') || '尚未選擇單位';
  });

  // 2.3.16 判斷能否儲存綁定單位
  const canSaveUnit = computed(() =>
    selectedBuildingID.value.length > 0 &&
    selectedFloor.value.length > 0 &&
    selectedUnitID.value.length > 0 &&
    Boolean(selectedUnit.value),
  );

  // 2.4 顯示帳戶電話資料
  const phoneDisplay = computed(() => {
    if (sessionStore.me?.phone_country_code === 'email') {
      return t('account.profile.phoneUnavailable');
    }

    const countryCode = formState.phone_country_code.trim();
    const phoneNumber = formState.phone_number.trim();
    return phoneNumber ? `${countryCode || '+852'} ${phoneNumber}`.trim() : t('account.profile.phoneUnavailable');
  });

  // 2.4.1 顯示 ismart 帳戶狀態
  const ismartStatusDisplay = computed(() =>
    sessionStore.me?.ismart_linked
      ? sessionStore.me.ismart_username || sessionStore.me.ismart_msg?.username || fallbackValue.value
      : t('account.profile.ismartUnlinked'),
  );

  // 2.6 顯示帳戶資料行
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
        key: 'publisher_identity_type',
        label: t('account.profile.publisherIdentity'),
        value: sessionStore.me?.publisher_identity_type?.trim() || fallbackValue.value,
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
        label: '員工管理屋苑',
        value: managedBuildingsDisplay.value,
      });
      return rows;
    }

    rows.push({
      key: 'resident_units',
      label: '所屬大廈 / 樓層 / 單位',
      value: residentUnitDisplay.value,
    });
    return rows;
  });

  // 2.7 顯示身份狀態資料行
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
      key: 'member_type',
      label: t('marketplace.myProfile.memberType'),
      value: sessionStore.me?.member_type?.trim() || fallbackValue.value,
    },
    {
      key: 'role',
      label: t('marketplace.myProfile.primaryRole'),
      value: sessionStore.me?.role?.trim() || fallbackValue.value,
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

  // 2.8 顯示角色與權限
  const roleChips = computed<string[]>(() =>
    sessionStore.me?.roles?.length ? sessionStore.me.roles : [t('marketplace.myProfile.noRoles')],
  );
  const permissionChips = computed<string[]>(() =>
    sessionStore.me?.permissions?.length ? sessionStore.me.permissions : [t('marketplace.myProfile.noPermissions')],
  );

  // 2.9 同步表單內容
  const syncFormState = (): void => {
    formState.display_name = sessionStore.me?.display_name ?? sessionStore.currentUser.display_name;
    formState.email = sessionStore.me?.email ?? '';
    formState.phone_country_code = sessionStore.me?.phone_country_code === 'email'
      ? '+852'
      : sessionStore.me?.phone_country_code || '+852';
    formState.phone_number = sessionStore.me?.phone_country_code === 'email'
      ? ''
      : sessionStore.me?.phone_number ?? '';
    formState.district_code = sessionStore.me?.district_code ?? '';
    formState.publisher_identity_type = sessionStore.me?.publisher_identity_type ?? '';
  };

  // 2.10 讀取會員資料
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

  // 2.11 開啟編輯彈窗
  const openEditModal = (): void => {
    syncFormState();
    isEditModalOpen.value = true;
  };

  // 2.12 關閉編輯彈窗
  const closeEditModal = (): void => {
    if (!isSaving.value && !isUploadingAvatar.value) {
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

  // 2.13 上傳頭像並更新會員資料
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
        ...buildPhonePayload(),
        publisher_identity_type: formState.publisher_identity_type.trim(),
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

  // 2.14 儲存會員資料
  const handleSaveProfile = async (): Promise<void> => {
    isSaving.value = true;

    try {
      const { data } = await updateMe({
        display_name: formState.display_name.trim(),
        ...buildPhonePayload(),
        publisher_identity_type: formState.publisher_identity_type.trim(),
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

  // 2.15 載入 POS 大廈清單
  const loadBuildings = async (): Promise<void> => {
    isBuildingsLoading.value = true;
    try {
      buildings.value = await fetchMemberPosBuildings();
    } catch (error) {
      const fallbackBuildings = allowedBuildingIDs.value.map((buildingID) => ({
        building_id: buildingID,
        buildname: resolveBuildingName(buildingID),
      }));
      buildings.value = fallbackBuildings;
      if (fallbackBuildings.length === 0) {
        feedbackStore.pushToast(readErrorMessage(error, '大廈資料載入失敗。'), 'error');
      }
    } finally {
      isBuildingsLoading.value = false;
    }
  };

  // 2.16 載入住戶權限單位詳情
  const loadMemberVisibleUnits = async (): Promise<void> => {
    if (isPOSStaff.value) {
      memberVisibleUnits.value = [];
      return;
    }

    const buildingIDs = clientBuildingIDs.value;
    if (buildingIDs.length === 0) {
      memberVisibleUnits.value = [];
      return;
    }

    const result = await Promise.allSettled(
      buildingIDs.map(async (buildingID) => {
        const units = await fetchMemberPosBuildingUnits(buildingID);
        return filterUnitsByPermission(buildingID, units, clientUnitIDs.value);
      }),
    );
    const merged = new Map<string, PosBuildingUnit>();
    result.forEach((item, index) => {
      const buildingID = buildingIDs[index] ?? '';
      const units = item.status === 'fulfilled'
        ? item.value
        : buildUnitsFromFlatUnitPermissions(buildingID, clientUnitIDs.value);
      units.forEach((unit) => {
        const unitID = getUnitID(unit);
        if (unitID) {
          merged.set(unitID, unit);
        }
      });
    });
    memberVisibleUnits.value = Array.from(merged.values());
  };

  // 2.17 同步目前已保存繳費單位
  const syncUnitSelection = (): void => {
    const profileBuildingID = sessionStore.me?.primary_community?.public_id?.trim() ?? '';
    selectedBuildingID.value = allowedBuildingIDs.value.includes(profileBuildingID)
      ? profileBuildingID
      : allowedBuildingIDs.value[0] ?? '';
    selectedFloor.value = sessionStore.me?.residence_floor?.trim() ?? '';
    selectedUnitID.value = '';
  };

  // 2.18 按目前會員資料匹配單位 ID
  const syncSelectedUnitFromProfile = (): void => {
    const floor = sessionStore.me?.residence_floor?.trim() ?? '';
    const unitName = sessionStore.me?.residence_unit?.trim() ?? '';
    if (!floor || !unitName) {
      selectedUnitID.value = '';
      return;
    }

    const matchedUnit = visibleSelectedUnits.value.find((item) =>
      getDisplayUnitFloor(item) === floor && getUnitName(item) === unitName,
    );
    selectedUnitID.value = matchedUnit ? getUnitID(matchedUnit) : '';
  };

  // 2.19 載入目前大廈的可選單位
  const loadSelectedBuildingUnits = async (buildingID = selectedBuildingID.value): Promise<void> => {
    const value = buildingID.trim();
    selectedBuildingUnits.value = [];
    selectedUnitID.value = '';
    if (!value) {
      return;
    }

    isUnitsLoading.value = true;
    try {
      const units = await fetchMemberPosBuildingUnits(value);
      const filteredUnits = filterUnitsByPermission(value, units, clientUnitIDs.value);
      const fallbackUnits = buildUnitsFromFlatUnitPermissions(value, clientUnitIDs.value);
      selectedBuildingUnits.value = filteredUnits.length > 0 ? filteredUnits : fallbackUnits;
      syncSelectedUnitFromProfile();
    } catch (error) {
      const fallbackUnits = buildUnitsFromFlatUnitPermissions(value, clientUnitIDs.value);
      selectedBuildingUnits.value = fallbackUnits;
      syncSelectedUnitFromProfile();
      if (fallbackUnits.length === 0) {
        feedbackStore.pushToast(readErrorMessage(error, '單位資料載入失敗。'), 'error');
      }
    } finally {
      isUnitsLoading.value = false;
    }
  };

  // 2.20 變更綁定大廈
  const handleUnitBuildingChange = (value: string | number): void => {
    selectedBuildingID.value = String(value);
    selectedFloor.value = '';
    selectedUnitID.value = '';
  };

  // 2.21 變更綁定樓層
  const handleUnitFloorChange = (value: string | number): void => {
    selectedFloor.value = String(value);
    selectedUnitID.value = '';
  };

  // 2.22 儲存繳費單位
  const handleSaveUnit = async (): Promise<void> => {
    if (!canSaveUnit.value || !selectedUnit.value) {
      feedbackStore.pushToast('請先選擇大廈、樓層與單位。', 'error');
      return;
    }

    const buildingName = buildingOptions.value.find((item) => item.value === selectedBuildingID.value)?.label ?? '';
    isSavingUnit.value = true;
    try {
      const { data } = await updateMe({
        display_name: formState.display_name.trim() || sessionStore.currentUser.display_name,
        ...buildPhonePayload(),
        publisher_identity_type: formState.publisher_identity_type.trim(),
        primary_community_id: selectedBuildingID.value,
        primary_community_name: buildingName,
        bound_building_ids: [selectedBuildingID.value],
        bound_flat_unit_ids: [selectedUnitID.value],
        district_code: formState.district_code.trim(),
        residence_floor: selectedFloor.value,
        residence_unit: getUnitName(selectedUnit.value),
      });

      sessionStore.me = data.data;
      syncFormState();
      isSyncingUnitSelection = true;
      syncUnitSelection();
      await loadMemberVisibleUnits();
      await loadSelectedBuildingUnits();
      feedbackStore.pushToast('繳費單位已更新。', 'success');
    } catch (error) {
      feedbackStore.pushToast(readErrorMessage(error, '繳費單位儲存失敗，請稍後再試。'), 'error');
    } finally {
      isSyncingUnitSelection = false;
      isSavingUnit.value = false;
    }
  };

  // 2.23 綁定 ismart 帳戶
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
      await loadBuildings();
      await loadMemberVisibleUnits();
      isSyncingUnitSelection = true;
      syncUnitSelection();
      if (selectedBuildingID.value) {
        await loadSelectedBuildingUnits(selectedBuildingID.value);
      }
      isIsmartModalOpen.value = false;
      feedbackStore.pushToast(t('account.profile.ismartBindSuccess'), 'success');
    } catch (error) {
      feedbackStore.pushToast(readErrorMessage(error, t('account.profile.ismartBindError')), 'error');
    } finally {
      isSyncingUnitSelection = false;
      isBindingIsmart.value = false;
    }
  };

  // 2.24 執行登出
  const handleSignOut = async (): Promise<void> => {
    isSigningOut.value = true;

    try {
      await sessionStore.signOut();
      await router.push('/login');
    } finally {
      isSigningOut.value = false;
    }
  };

  // 2.25 初始化會員資料
  onMounted(async () => {
    try {
      await loadProfile();
      await loadBuildings();
      await loadMemberVisibleUnits();
      isSyncingUnitSelection = true;
      syncUnitSelection();
      if (selectedBuildingID.value) {
        await loadSelectedBuildingUnits(selectedBuildingID.value);
      }
    } finally {
      isSyncingUnitSelection = false;
    }
  });

  watch(selectedBuildingID, (nextValue, previousValue) => {
    if (nextValue === previousValue) {
      return;
    }
    if (isSyncingUnitSelection) {
      return;
    }
    selectedFloor.value = '';
    selectedUnitID.value = '';
    void loadSelectedBuildingUnits(nextValue);
  });

  watch(selectedFloor, (nextValue, previousValue) => {
    if (nextValue !== previousValue && !isSyncingUnitSelection) {
      selectedUnitID.value = '';
    }
  });

  return {
    accountRows,
    buildingOptions,
    canSaveUnit,
    closeEditModal,
    closeIsmartModal,
    floorOptions,
    formState,
    handleAvatarFileChange,
    handleBindIsmart,
    handleSaveUnit,
    handleSaveProfile,
    handleSignOut,
    handleUnitBuildingChange,
    handleUnitFloorChange,
    isBuildingsLoading,
    isBindingIsmart,
    isEditModalOpen,
    isIsmartModalOpen,
    isLoading,
    isSaving,
    isSavingUnit,
    isSigningOut,
    isUnitsLoading,
    isUploadingAvatar,
    ismartFormState,
    openEditModal,
    openIsmartModal,
    permissionChips,
    phoneDisplay,
    profileRows,
    roleChips,
    selectedBuildingID,
    selectedFloor,
    selectedUnitDisplay,
    selectedUnitID,
    sessionStore,
    t,
    unitOptions,
  };
};

<!--
 * 會員中心頁。
 * 1. 承載左側模組導覽與右側內容面板，支援帳號管理、授權副戶、物業綁定、AJO 錢包、我的樓盤、我的住宅、我的家具與我的收藏等面板切換。
 * 2. 綁定單位讀取 POS 大廈與單位資料，帳號管理讀取目前會員資料。
 * 3. CSS 變量與樣式嚴格對齊 HTML 設計稿 ajo_living_desktop_20260624(3)(10).html 的原生變量名。
 * 4. 響應式設計：桌面雙欄、行動單欄（900px / 560px 斷點）。
-->
<script setup lang="ts">
/*
 * 會員中心頁邏輯。
 * 1. 面板索引型別與導覽項目定義。
 * 2. 當前面板狀態與切換方法。
 * 3. 帳號管理、授權副戶、物業綁定、錢包、樓盤入口、住宅、家具與收藏資料。
 * 4. 退出登入後返回登入頁。
 */
import axios from 'axios';
import { computed, onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { RouterView, useRoute, useRouter } from 'vue-router';

import { isIntegrationBusinessAuthError } from '@/httpapis';
import {
  fetchMemberIsmartSubaccounts,
  fetchMemberPosBuildings,
  fetchMemberPosBuildingUnits,
  fetchPosBuildings,
  fetchPosBuildingUnits,
  grantMemberIsmartSubaccount,
  revokeMemberIsmartSubaccount,
  submitMemberIsmartOwnerBindingRequest,
  type IsmartSubaccountRow,
} from '@/httpapis/building';
import { updateMe } from '@/httpapis/me';
import type { PosBuilding, PosBuildingUnit } from '@/model/community';
import { useFeedbackStore } from '@/stores/feedback';
import { usePreferenceStore } from '@/stores/preferences';
import { useSessionStore } from '@/stores/session';
import AccountWalletPage from '@/pages/account/my/profile/wallet/Page.vue';
import PropertyMyPage from '@/pages/property/my/PropertyMyPage.vue';
import { resolveResidentUnitDisplays } from './composables/unit-display';

const route = useRoute();
const router = useRouter();
const { t } = useI18n();
const feedbackStore = useFeedbackStore();
const preferenceStore = usePreferenceStore();
const sessionStore = useSessionStore();

// 1. 面板索引類型
type PanelKey =
  | 'profile-account'
  | 'profile-agency-company'
  | 'profile-subaccounts'
  | 'profile-property-binding'
  | 'profile-wallet'
  | 'profile-chat'
  | 'profile-properties'
  | 'profile-homes'
  | 'profile-furniture'
  | 'profile-saved';

// 2. 導覽項目（對齊 HTML data-work-target）
const isRestrictedAgencyAccount = computed(() =>
  ['individual_agent', 'agency_company'].includes(sessionStore.me?.account_type ?? '')
  && ['pending_profile', 'pending_review', 'rejected'].includes(sessionStore.me?.member_status ?? ''),
);

const navItems = computed<Array<{ key: PanelKey; label: string; needsApi?: boolean }>>(() => isRestrictedAgencyAccount.value
  ? [{ key: 'profile-agency-company', label: t('account.center.nav.agencyCompany') }]
  : [
  { key: 'profile-account', label: t('account.center.nav.account') },
  ...(['individual_agent', 'agency_company'].includes(sessionStore.me?.account_type ?? '')
    ? [{ key: 'profile-agency-company' as PanelKey, label: t('account.center.nav.agencyCompany') }]
    : []),
  { key: 'profile-subaccounts', label: t('account.center.nav.subaccounts') },
  { key: 'profile-property-binding', label: t('account.center.nav.propertyBinding') },
  { key: 'profile-wallet', label: t('account.center.nav.wallet') },
  { key: 'profile-chat', label: t('account.center.nav.messages') },
  { key: 'profile-properties', label: t('account.center.nav.properties') },
  { key: 'profile-homes', label: t('account.center.nav.homes') },
  { key: 'profile-furniture', label: t('account.center.nav.furniture') },
  { key: 'profile-saved', label: t('account.center.nav.favorites') },
  ]);

// 3. 當前面板
const activePanel = ref<PanelKey>('profile-account');
const isSigningOut = ref(false);

const routePanelPaths = [
  '/account/profile/wallet',
  '/account/profile/agency-profile',
  '/account/chat',
  '/account/properties',
  '/account/listings',
  '/account/favorites',
  '/account/marketplace/my',
];

// 4. 判斷是否在原會員中心外殼內渲染子路由內容
const shouldRenderRoutePanel = computed(() =>
  routePanelPaths.some((path) => route.path === path || route.path.startsWith(`${path}/`)),
);

// 5. 切換面板
const switchPanel = (key: PanelKey) => {
  const routeMap: Partial<Record<PanelKey, string>> = {
    'profile-agency-company': '/account/profile/agency-profile',
    'profile-wallet': '/account/profile/wallet',
    'profile-chat': '/account/chat',
    'profile-properties': '/account/properties/sale',
    'profile-homes': '/account/properties/serviced-residences',
    'profile-furniture': '/account/listings',
    'profile-saved': '/account/favorites',
  };
  const nextRoute = routeMap[key];
  if (nextRoute) {
    void router.push(nextRoute);
    return;
  }

  activePanel.value = key;
  void router.push('/account/profile');
};

// 6. 根據路由同步會員中心面板
const syncActivePanelFromRoute = (path: string) => {
  if (path.startsWith('/account/profile/agency-profile')) {
    activePanel.value = 'profile-agency-company';
    return;
  }
  if (path.startsWith('/account/profile/wallet')) {
    activePanel.value = 'profile-wallet';
    return;
  }
  if (path.startsWith('/account/chat')) {
    activePanel.value = 'profile-chat';
    return;
  }
  if (path.startsWith('/account/properties/serviced-residences')) {
    activePanel.value = 'profile-homes';
    return;
  }
  if (path.startsWith('/account/properties')) {
    activePanel.value = 'profile-properties';
    return;
  }
  if (path.startsWith('/account/listings')) {
    activePanel.value = 'profile-furniture';
    return;
  }
  if (path.startsWith('/account/favorites')) {
    activePanel.value = 'profile-saved';
    return;
  }
  if (path === '/account/profile' || path === '/account/profile/info') {
    activePanel.value = route.query.panel === 'property-binding'
      ? 'profile-property-binding'
      : 'profile-account';
  }
};

interface AccountDisplayField {
  label: string;
  value: string;
}

const unsetText = computed(() => t('account.center.common.unset'));

// 7. 格式化會員資料顯示值
const displayText = (value: unknown, fallback = unsetText.value): string => {
  const text = String(value ?? '').trim();
  return text || fallback;
};

// 8. 遮罩證件編號
const maskIdentityNumber = (value: unknown): string => {
  const text = String(value ?? '').trim();
  return text ? '********' : unsetText.value;
};

const ismartProfile = computed(() => sessionStore.me?.ismart_account_profile ?? null);
const accountDisplayName = computed(() => displayText(sessionStore.me?.display_name, sessionStore.currentUser.display_name));
const accountAvatarText = computed(() => accountDisplayName.value.trim().slice(0, 1).toUpperCase() || 'A');
const memberPhoneText = computed(() => {
  const member = sessionStore.me;
  if (!member) {
    return unsetText.value;
  }
  if (isSystemPhonePlaceholder(member.phone_country_code)) {
    return displayText(member.ismart_bound_phone || ismartProfile.value?.account_phone || member.ismart_msg?.phone);
  }
  const countryCode = displayText(member.phone_country_code, '');
  const phoneNumber = displayText(member.phone_number, '');
  return displayText([countryCode, phoneNumber].filter(Boolean).join(' '));
});

// 9. 帳號管理 AJO 專屬資料
const ajoAccountData = computed<AccountDisplayField[]>(() => [
  {
    label: t('account.profile.localPassword'),
    value: sessionStore.me ? t('account.center.account.passwordManaged') : unsetText.value,
  },
  {
    label: t('auth.publisherIdentityType'),
    value: t(`auth.accountType${({
      personal: 'Personal',
      individual_agent: 'IndividualAgent',
      agency_company: 'AgencyCompany',
      agency_company_subaccount: 'AgencyCompanySubaccount',
    } as Record<string, string>)[sessionStore.me?.account_type ?? 'personal'] ?? 'Personal'}`),
  },
  { label: t('account.profile.districtCode'), value: displayText(sessionStore.me?.district_code) },
  { label: t('account.center.account.residentUnits'), value: residentUnitNameText.value },
]);

const identityStatus = computed(() => ({
  memberNo: displayText(sessionStore.me?.public_id),
  memberStatus: displayText(sessionStore.me?.member_status),
  memberType: displayText(sessionStore.me?.member_type),
  primaryRole: displayText(sessionStore.me?.role),
  profileCompleteness: sessionStore.me?.profile_completed
    ? t('account.center.account.profileComplete')
    : t('account.center.account.profileIncomplete'),
  staffPermission: sessionStore.me?.is_staff
    ? t('account.center.account.staffEnabled')
    : t('account.center.account.staffDisabled'),
}));
const accountRoles = computed(() => {
  const roles = sessionStore.me?.roles?.filter(Boolean) ?? [];
  return roles.length > 0 ? roles : [displayText(sessionStore.me?.role, 'user')];
});
const accountPermissionText = computed(() =>
  (sessionStore.me?.permissions?.length ?? 0) > 0
    ? t('account.center.account.permissionCount', { count: sessionStore.me?.permissions?.length ?? 0 })
    : t('account.center.account.noPermissions'),
);

interface BindOption {
  label: string;
  value: string;
}

interface BindPropertyOption extends BindOption {
  buildingID: string;
  floor: string;
  unit: string;
}

const bindBuildings = ref<PosBuilding[]>([]);
const bindUnits = ref<PosBuildingUnit[]>([]);
const residentUnitDetails = ref<PosBuildingUnit[]>([]);
const bindBuildingID = ref('');
const bindFloor = ref('');
const bindUnitID = ref('');
const bindLoading = ref(false);
const bindUnitsLoading = ref(false);
const bindSaving = ref(false);
const integrationAuthUnavailable = ref(false);
let latestBindUnitsRequestID = 0;
let isSyncingBindProfile = false;
const bindUnassignedFloor = '__unassigned__';
const memberUnitCache = new Map<string, PosBuildingUnit[]>();
const memberUnitRequests = new Map<string, Promise<PosBuildingUnit[]>>();

// 6.1 判斷是否為系統佔位電話帳號
const isSystemPhonePlaceholder = (countryCode?: string): boolean =>
  ['email', 'ismart'].includes((countryCode ?? '').trim().toLowerCase());

// 6.2 讀取 POS 大廈與單位欄位
const getBindBuildingID = (item: PosBuilding): string =>
  String(item.building_id ?? item.id ?? '').trim();
const getBindBuildingName = (item: PosBuilding): string => {
  const buildingID = getBindBuildingID(item);
  const name = String(
    preferenceStore.locale === 'en'
      ? item.buildname ?? item.name ?? item.buildname_chi ?? ''
      : item.buildname_chi ?? item.buildname ?? item.name ?? '',
  ).trim();
  if (name && name !== buildingID) {
    return name;
  }

  const community = sessionStore.me?.primary_community;
  if (community?.public_id?.trim() !== buildingID) {
    return '';
  }
  const communityName = String(
    preferenceStore.locale === 'en'
      ? community.name_en || community.name_zh || community.address_text || ''
      : community.name_zh || community.name_en || community.address_text || '',
  ).trim();
  return communityName === buildingID ? '' : communityName;
};
const getBindBuildingLabel = (item: PosBuilding): string =>
  getBindBuildingName(item) || t('account.center.common.buildingCodeLabel', { id: getBindBuildingID(item) });
const getBindUnitID = (item: PosBuildingUnit): string =>
  String(item.unit_id ?? item.id ?? '').trim();
const getBindUnitFloor = (item: PosBuildingUnit): string =>
  String(item.floor ?? '').trim();
const getBindDisplayUnitFloor = (item: PosBuildingUnit): string =>
  getBindUnitFloor(item) || bindUnassignedFloor;
const formatBindFloorLabel = (value: string): string =>
  value === bindUnassignedFloor ? t('account.center.common.unassignedFloor') : value;
const getBindUnitName = (item: PosBuildingUnit): string =>
  String(item.unit ?? item.unit_name ?? item.name ?? '').trim();
const getBindDigitsOnly = (value: unknown): string =>
  String(value ?? '').replace(/\D/g, '');

// 6.3 正規化 POS 權限 ID 列表
const normalizeBindList = (values: string[] | undefined): string[] => {
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

// 6.4 正規化可顯示的 POS 單位權限碼
const normalizeBindFlatUnitPermissions = (values: string[] | undefined): string[] => {
  const result: string[] = [];
  normalizeBindList(values).forEach((value) => {
    const unitID = getBindDigitsOnly(value);
    if (unitID.length < 11) {
      return;
    }

    const normalizedUnitID = unitID.slice(0, 11);
    if (!result.includes(normalizedUnitID)) {
      result.push(normalizedUnitID);
    }
  });
  return result;
};

// 6.6 取得住戶有效單位權限
const bindAllowedUnitIDs = computed(() =>
  normalizeBindFlatUnitPermissions(
    sessionStore.me?.ismart_msg?.client_building_flat_units_permissions ?? sessionStore.me?.bound_flat_unit_ids,
  ),
);

// 6.7 從住戶單位權限取得所屬屋苑
const bindResidentBuildingIDs = computed(() => {
  const result: string[] = [];
  bindAllowedUnitIDs.value.forEach((unitID) => {
    const buildingID = unitID.slice(0, 7);
    if (buildingID && !result.includes(buildingID)) {
      result.push(buildingID);
    }
  });
  return result;
});

// 6.8 建立大廈 ID 與名稱索引
const bindBuildingNameMap = computed<Record<string, string>>(() =>
  bindBuildings.value.reduce<Record<string, string>>((result, item) => {
    const buildingID = getBindBuildingID(item);
    if (buildingID) {
      result[buildingID] = getBindBuildingName(item);
    }
    return result;
  }, {}),
);

// 6.11 顯示住戶所屬屋苑、樓層與單位
const residentUnitNameText = computed(() =>
  displayText(
    resolveResidentUnitDisplays(bindAllowedUnitIDs.value, residentUnitDetails.value).map(({ buildingID, floor, unit }) => {
      const floorLabel = floor || t('account.center.common.unassignedFloor');
      const buildingName = bindBuildingNameMap.value[buildingID]
        || t('account.center.common.buildingCodeLabel', { id: buildingID });
      return [buildingName, floorLabel, unit].filter(Boolean).join(' / ');
    }).join(', '),
  ),
);

// 6.14 判斷 POS 單位是否可選
const isSelectableBindUnit = (buildingID: string, item: PosBuildingUnit): boolean => {
  const normalizedBuildingID = getBindDigitsOnly(buildingID).slice(0, 7);
  const normalizedUnitID = getBindDigitsOnly(getBindUnitID(item));
  const hasFloor = getBindUnitFloor(item).length > 0;
  const hasUnit = getBindUnitName(item).length > 0;
  return !(normalizedBuildingID && normalizedUnitID === normalizedBuildingID && !hasFloor && !hasUnit);
};

// 6.15 由 POS 權限 ID 建立可選單位
const buildBindUnitsFromFlatUnitPermissions = (
  buildingID: string,
  flatUnitPermissions: string[],
): PosBuildingUnit[] => {
  const normalizedBuildingID = getBindDigitsOnly(buildingID).slice(0, 7);
  if (!normalizedBuildingID || flatUnitPermissions.length === 0) {
    return [];
  }

  const seen = new Set<string>();
  const units: PosBuildingUnit[] = [];
  flatUnitPermissions.forEach((rawPermission) => {
    const permission = getBindDigitsOnly(rawPermission);
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

  return units.sort((left, right) => getBindUnitID(left).localeCompare(getBindUnitID(right), 'en', { numeric: true }));
};

// 6.16 轉換 POS 權限碼
const toBindTwoDigitCode = (value: unknown): string => {
  const digits = getBindDigitsOnly(value);
  return digits ? digits.slice(-2).padStart(2, '0') : '';
};

// 6.17 取得單位權限碼
const getBindUnitPermissionCode = (buildingID: string, item: PosBuildingUnit): string => {
  const normalizedBuildingID = getBindDigitsOnly(buildingID).slice(0, 7);
  const unitID = getBindDigitsOnly(getBindUnitID(item));
  if (unitID.startsWith(normalizedBuildingID) && unitID.length >= 11) {
    return unitID.slice(0, 11);
  }

  const floorCode = toBindTwoDigitCode(getBindUnitFloor(item));
  const unitCode = toBindTwoDigitCode(getBindUnitName(item));
  return normalizedBuildingID && floorCode && unitCode ? `${normalizedBuildingID}${floorCode}${unitCode}` : '';
};

// 6.18 判斷單位是否符合 POS 權限
const matchesBindUnitPermission = (
  buildingID: string,
  item: PosBuildingUnit,
  flatUnitPermissions: string[],
): boolean => {
  if (flatUnitPermissions.length === 0) {
    return true;
  }

  const permissionCode = getBindUnitPermissionCode(buildingID, item);
  if (!permissionCode) {
    return false;
  }

  const normalizedBuildingID = getBindDigitsOnly(buildingID).slice(0, 7);
  const compactUnitID = permissionCode.replace(/0/g, '');
  return flatUnitPermissions.some((rawPermission) => {
    const permission = getBindDigitsOnly(rawPermission);
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

// 6.19 按權限過濾 POS 單位
const filterBindUnitsByPermission = (
  buildingID: string,
  units: PosBuildingUnit[],
  flatUnitPermissions: string[],
): PosBuildingUnit[] =>
  units
    .filter((item) => isSelectableBindUnit(buildingID, item))
    .filter((item) => matchesBindUnitPermission(buildingID, item, flatUnitPermissions));

// 6.20 在頁面生命週期內合併相同大廈的會員單位請求
const fetchCachedMemberUnits = async (buildingID: string): Promise<PosBuildingUnit[]> => {
  const value = buildingID.trim();
  const cached = memberUnitCache.get(value);
  if (cached) {
    return cached;
  }

  const pending = memberUnitRequests.get(value);
  if (pending) {
    return pending;
  }

  const request = fetchMemberPosBuildingUnits(value)
    .then((units) => {
      memberUnitCache.set(value, units);
      return units;
    })
    .finally(() => {
      memberUnitRequests.delete(value);
    });
  memberUnitRequests.set(value, request);
  return request;
};

// 6.21 載入住戶權限單位詳情
const loadResidentUnitDetails = async (): Promise<void> => {
  if (bindResidentBuildingIDs.value.length === 0) {
    residentUnitDetails.value = [];
    return;
  }

  const results = await Promise.allSettled(
    bindResidentBuildingIDs.value.map(async (buildingID) => {
      const fallbackUnits = buildBindUnitsFromFlatUnitPermissions(buildingID, bindAllowedUnitIDs.value);
      const units = await fetchCachedMemberUnits(buildingID);
      const actualUnits = filterBindUnitsByPermission(buildingID, units, bindAllowedUnitIDs.value);
      const actualUnitMap = new Map(
        actualUnits.map((item) => [getBindDigitsOnly(getBindUnitID(item)).slice(0, 11), item]),
      );
      return fallbackUnits.map((item) =>
        actualUnitMap.get(getBindDigitsOnly(getBindUnitID(item)).slice(0, 11)) ?? item,
      );
    }),
  );
  const unitMap = new Map<string, PosBuildingUnit>();
  results.forEach((result, index) => {
    const buildingID = bindResidentBuildingIDs.value[index] ?? '';
    const units = result.status === 'fulfilled'
      ? result.value
      : buildBindUnitsFromFlatUnitPermissions(buildingID, bindAllowedUnitIDs.value);
    units.forEach((item) => {
      const unitID = getBindDigitsOnly(getBindUnitID(item)).slice(0, 11);
      if (unitID) {
        unitMap.set(unitID, item);
      }
    });
  });
  residentUnitDetails.value = Array.from(unitMap.values());
};

// 6.22 POS 顯示排序
const compareBindCodes = (left: string, right: string): number =>
  left.localeCompare(right, 'en', {
    numeric: true,
    sensitivity: 'base',
  });

const bindAllowedBuildingIDs = computed(() => {
  const message = sessionStore.me?.ismart_msg;
  const primaryCommunityID = sessionStore.me?.primary_community?.public_id?.trim() ?? '';
  if (message) {
    return bindResidentBuildingIDs.value;
  }
  return bindResidentBuildingIDs.value.length > 0 || !primaryCommunityID
    ? bindResidentBuildingIDs.value
    : [primaryCommunityID];
});
const visibleBindBuildings = computed(() => {
  const allowed = new Set(bindAllowedBuildingIDs.value);
  return bindBuildings.value.filter((item) => allowed.has(getBindBuildingID(item)));
});
const visibleBindUnits = computed(() => {
  return bindUnits.value.filter((item) =>
    matchesBindUnitPermission(bindBuildingID.value, item, bindAllowedUnitIDs.value),
  );
});
const bindPropertyOptions = computed<BindPropertyOption[]>(() => {
  const details = new Map(
    residentUnitDetails.value.map((item) => [getBindDigitsOnly(getBindUnitID(item)).slice(0, 11), item]),
  );
  return bindAllowedUnitIDs.value
    .map((value) => {
      const unit = details.get(value);
      if (!unit) return null;
      const buildingID = value.slice(0, 7);
      const buildingName = bindBuildingNameMap.value[buildingID]
        || t('account.center.common.buildingCodeLabel', { id: buildingID });
      const floor = getBindDisplayUnitFloor(unit);
      const unitName = getBindUnitName(unit);
      return {
        value,
        buildingID,
        floor,
        unit: unitName,
        label: [buildingName, formatBindFloorLabel(floor), unitName].filter(Boolean).join(' / '),
      };
    })
    .filter((item): item is BindPropertyOption => Boolean(item?.unit))
    .sort((left, right) => compareBindCodes(left.label, right.label));
});
const bindSelectedUnit = computed(() =>
  residentUnitDetails.value.find((item) => getBindDigitsOnly(getBindUnitID(item)).slice(0, 11) === bindUnitID.value),
);
const bindSelectedProperty = computed(() =>
  bindPropertyOptions.value.find((item) => item.value === bindUnitID.value),
);
const bindCurrent = computed(() =>
  bindSelectedProperty.value?.label || t('account.center.common.noUnitSelected'),
);
const canSaveBindUnit = computed(() => Boolean(bindSelectedProperty.value && bindSelectedUnit.value));

// 6.6 同步會員已保存單位
const syncBindFromProfile = (): void => {
  isSyncingBindProfile = true;
  const profileBuildingID = sessionStore.me?.primary_community?.public_id?.trim() ?? '';
  bindBuildingID.value = bindAllowedBuildingIDs.value.includes(profileBuildingID)
    ? profileBuildingID
    : bindAllowedBuildingIDs.value[0] ?? '';
  const residenceFloor = sessionStore.me?.residence_floor?.trim() ?? '';
  const residenceUnit = sessionStore.me?.residence_unit?.trim() ?? '';
  bindFloor.value = residenceFloor || (residenceUnit ? bindUnassignedFloor : '');
  bindUnitID.value = '';
};

// 6.7 按已保存樓層與單位名稱同步 POS 單位
const syncBindUnitFromProfile = (): void => {
  const floor = sessionStore.me?.residence_floor?.trim() || bindUnassignedFloor;
  const unitName = sessionStore.me?.residence_unit?.trim() ?? '';
  if (!unitName) {
    bindUnitID.value = '';
    return;
  }

  const matchedUnit = visibleBindUnits.value.find((item) =>
    getBindDisplayUnitFloor(item) === floor && getBindUnitName(item) === unitName,
  );
  bindUnitID.value = matchedUnit ? getBindUnitID(matchedUnit) : '';
};

// 6.8 載入 POS 大廈與單位資料
const loadBindBuildings = async (): Promise<void> => {
  bindLoading.value = true;
  try {
    const memberBuildings = await fetchMemberPosBuildings();
    integrationAuthUnavailable.value = false;
    const memberBuildingMap = new Map(memberBuildings.map((item) => [getBindBuildingID(item), item]));
    const hasMissingBuildingName = bindAllowedBuildingIDs.value.some((buildingID) => {
      const item = memberBuildingMap.get(buildingID);
      return !item || !getBindBuildingName(item);
    });
    let publicBuildings: PosBuilding[] = [];
    if (hasMissingBuildingName) {
      try {
        publicBuildings = await fetchPosBuildings();
      } catch (error) {
        console.error(error);
      }
    }
    const publicBuildingMap = new Map(publicBuildings.map((item) => [getBindBuildingID(item), item]));
    const mergedBuildingMap = new Map(memberBuildingMap);
    bindAllowedBuildingIDs.value.forEach((buildingID) => {
      const current = mergedBuildingMap.get(buildingID);
      if (!current || !getBindBuildingName(current)) {
        mergedBuildingMap.set(buildingID, publicBuildingMap.get(buildingID) ?? current ?? { building_id: buildingID });
      }
    });
    bindBuildings.value = Array.from(mergedBuildingMap.values());
  } catch (error) {
    integrationAuthUnavailable.value = isIntegrationBusinessAuthError(error);
    let publicBuildings: PosBuilding[] = [];
    try {
      publicBuildings = await fetchPosBuildings();
    } catch (publicError) {
      console.error(publicError);
    }
    const allowedBuildingIDs = new Set(bindAllowedBuildingIDs.value);
    const visiblePublicBuildings = publicBuildings.filter((item) => allowedBuildingIDs.has(getBindBuildingID(item)));
    const loadedBuildingIDs = new Set(visiblePublicBuildings.map((item) => getBindBuildingID(item)));
    const fallbackBuildings = [
      ...visiblePublicBuildings,
      ...bindAllowedBuildingIDs.value
        .filter((buildingID) => !loadedBuildingIDs.has(buildingID))
        .map((buildingID) => ({ building_id: buildingID })),
    ];
    bindBuildings.value = fallbackBuildings;
    if (fallbackBuildings.length === 0) {
      feedbackStore.pushToast(t('account.center.account.buildingLoadError'), 'error');
    }
  } finally {
    bindLoading.value = false;
  }
};

const loadBindUnits = async (buildingID: string): Promise<void> => {
  const requestID = ++latestBindUnitsRequestID;
  const value = buildingID.trim();
  if (!value) {
    bindUnits.value = [];
    return;
  }

  if (integrationAuthUnavailable.value) {
    bindUnits.value = buildBindUnitsFromFlatUnitPermissions(value, bindAllowedUnitIDs.value);
    if (isSyncingBindProfile) {
      syncBindUnitFromProfile();
    }
    return;
  }

  bindUnitsLoading.value = true;
  try {
    const result = await fetchCachedMemberUnits(value);
    if (requestID !== latestBindUnitsRequestID) {
      return;
    }
    const filteredUnits = filterBindUnitsByPermission(value, result, bindAllowedUnitIDs.value);
    const fallbackUnits = buildBindUnitsFromFlatUnitPermissions(value, bindAllowedUnitIDs.value);
    bindUnits.value = filteredUnits.length > 0 ? filteredUnits : fallbackUnits;
    if (isSyncingBindProfile) {
      syncBindUnitFromProfile();
      return;
    }
    if (bindUnitID.value && !visibleBindUnits.value.some((item) => getBindUnitID(item) === bindUnitID.value)) {
      bindUnitID.value = '';
    }
  } catch {
    if (requestID !== latestBindUnitsRequestID) {
      return;
    }
    const fallbackUnits = buildBindUnitsFromFlatUnitPermissions(value, bindAllowedUnitIDs.value);
    bindUnits.value = fallbackUnits;
    if (fallbackUnits.length === 0) {
      feedbackStore.pushToast(t('account.center.account.unitLoadError'), 'error');
    }
  } finally {
    if (requestID === latestBindUnitsRequestID) {
      bindUnitsLoading.value = false;
    }
  }
};

// 6.9 儲存會員繳費單位
const handleSaveBindUnit = async (): Promise<void> => {
  if (!canSaveBindUnit.value || !bindSelectedProperty.value) {
    feedbackStore.pushToast(t('account.center.account.selectUnitRequired'), 'error');
    return;
  }

  bindSaving.value = true;
  try {
    const selectedProperty = bindSelectedProperty.value;
    const buildingName = bindBuildingNameMap.value[selectedProperty.buildingID]
      || selectedProperty.buildingID;
    const { data } = await updateMe({
      display_name: sessionStore.me?.display_name ?? sessionStore.currentUser.display_name,
      ...(sessionStore.me?.email ? { email: sessionStore.me.email } : {}),
      ...(isSystemPhonePlaceholder(sessionStore.me?.phone_country_code)
        ? {}
        : {
            phone_country_code: sessionStore.me?.phone_country_code ?? '',
            phone_number: sessionStore.me?.phone_number ?? '',
          }),
      primary_community_id: selectedProperty.buildingID,
      primary_community_name: buildingName,
      bound_building_ids: [selectedProperty.buildingID],
      bound_flat_unit_ids: [bindUnitID.value],
      residence_floor: selectedProperty.floor === bindUnassignedFloor ? '' : selectedProperty.floor,
      residence_unit: selectedProperty.unit,
      district_code: sessionStore.me?.district_code ?? '',
    });
    sessionStore.me = data.data;
    feedbackStore.pushToast(t('account.center.account.unitSaved'), 'success');
  } catch {
    feedbackStore.pushToast(t('account.center.account.unitSaveError'), 'error');
  } finally {
    bindSaving.value = false;
  }
};

const ismartAccountData = computed<AccountDisplayField[]>(() => [
  {
    label: t('account.center.account.accountCode'),
    value: displayText(ismartProfile.value?.account_code || sessionStore.me?.ismart_username || sessionStore.me?.ismart_msg?.username),
  },
  {
    label: t('account.center.account.accountPhone'),
    value: displayText(ismartProfile.value?.account_phone || sessionStore.me?.ismart_bound_phone || sessionStore.me?.ismart_msg?.phone),
  },
  {
    label: t('account.center.account.accountEmail'),
    value: displayText(ismartProfile.value?.account_email || sessionStore.me?.ismart_msg?.email),
  },
  { label: t('account.center.account.ownerNameEnglish'), value: displayText(ismartProfile.value?.owner_name_en) },
  { label: t('account.center.account.ownerNameChinese'), value: displayText(ismartProfile.value?.owner_name_zh) },
  { label: t('account.center.account.identityNumber'), value: maskIdentityNumber(ismartProfile.value?.identity_number) },
]);

const ismartHouseholdData = computed<AccountDisplayField[]>(() => [
  { label: t('account.center.account.legalEntity'), value: displayText(ismartProfile.value?.legal_entity) },
  { label: t('account.center.account.gender'), value: displayText(ismartProfile.value?.gender) },
  { label: t('account.center.account.birthDate'), value: displayText(ismartProfile.value?.birth_date) },
  { label: t('account.center.account.contactName'), value: displayText(ismartProfile.value?.contact_name) },
  { label: t('account.center.account.contactPhone'), value: displayText(ismartProfile.value?.contact_phone) },
  { label: t('account.center.account.billingEmail'), value: displayText(ismartProfile.value?.billing_email) },
  { label: t('account.center.account.billingAddress'), value: displayText(ismartProfile.value?.billing_address) },
]);

interface SubaccountDisplayRow {
  key: string;
  relationInfoID: string;
  location: string;
  user: string;
  meta: string;
  role: string;
  historyCount: number;
  targetUserID: string;
}

interface SubaccountGroup {
  unitID: string;
  buildingID: string;
  name: string;
  rows: SubaccountDisplayRow[];
  grantTargetUserID: string;
  grantRemark: string;
  isGranting: boolean;
  error: string;
}

interface SubaccountUnitContext {
  unitID: string;
  buildingID: string;
  name: string;
}

const subaccountGroups = ref<SubaccountGroup[]>([]);
const subaccountsLoading = ref(false);
const subaccountsLoaded = ref(false);
const subaccountsError = ref('');
const activeSubaccountGrantUnitID = ref('');
const revokingSubaccountKey = ref('');

// 7.1 讀取 API 錯誤訊息
const readAccountApiErrorMessage = (error: unknown, fallback: string): string =>
  axios.isAxiosError<{ message?: string }>(error)
    ? error.response?.data?.message ?? fallback
    : fallback;

// 7.2 判斷是否為無業主權限單位
const isForbiddenRequest = (error: unknown): boolean =>
  axios.isAxiosError(error) && error.response?.status === 403;

// 7.3 轉換授權副戶顯示文字
const toSubaccountText = (value: unknown, fallback = ''): string => {
  const text = String(value ?? '').trim();
  return text || fallback;
};

// 7.4 組裝授權副戶聯絡資料
const buildSubaccountMeta = (row: IsmartSubaccountRow): string =>
  [row.target_phone, row.target_email]
    .map((item) => toSubaccountText(item))
    .filter(Boolean)
    .join(' · ');

// 7.5 轉換授權副戶列表列
const buildSubaccountDisplayRow = (
  row: IsmartSubaccountRow,
  groupName: string,
  index: number,
): SubaccountDisplayRow => {
  const targetUserID = toSubaccountText(row.target_user_id);
  const historyCount = Number(row.history_count ?? 0);
  const rowKey =
    toSubaccountText(row.relation_info_id) ||
    [row.unit_id, targetUserID, index].map((item) => toSubaccountText(item)).filter(Boolean).join('-');

  return {
    key: rowKey || `subaccount-${index}`,
    relationInfoID: toSubaccountText(row.relation_info_id, '-'),
    location: toSubaccountText(row.unit_display, groupName),
    user: toSubaccountText(row.target_name || row.target_username || row.target_client_id || targetUserID),
    meta: buildSubaccountMeta(row),
    role: toSubaccountText(row.role),
    historyCount: Number.isFinite(historyCount) ? historyCount : 0,
    targetUserID,
  };
};

// 7.6 建立授權副戶分組
const createSubaccountGroup = (unit: SubaccountUnitContext, rows: IsmartSubaccountRow[] = []): SubaccountGroup => ({
  unitID: unit.unitID,
  buildingID: unit.buildingID,
  name: unit.name,
  rows: rows.map((row, index) => buildSubaccountDisplayRow(row, unit.name, index)),
  grantTargetUserID: '',
  grantRemark: '',
  isGranting: false,
  error: '',
});

// 7.7 建立授權副戶單位名稱
const buildSubaccountUnitName = (buildingID: string, item: PosBuildingUnit): string => {
  const buildingName = bindBuildingNameMap.value[buildingID]
    || t('account.center.common.buildingCodeLabel', { id: buildingID });
  const floor = getBindUnitFloor(item);
  const unitName = getBindUnitName(item);
  return [buildingName, floor, unitName].filter(Boolean).join(' ') || getBindUnitID(item);
};

// 7.8 讀取會員可見單位作為授權副戶分組
const loadSubaccountUnitContexts = async (): Promise<SubaccountUnitContext[]> => {
  const buildingIDs = bindAllowedBuildingIDs.value;
  const unitPermissions = bindAllowedUnitIDs.value;
  if (buildingIDs.length === 0 || unitPermissions.length === 0) {
    return [];
  }

  if (residentUnitDetails.value.length === 0) {
    await loadResidentUnitDetails();
  }
  const unitMap = new Map<string, SubaccountUnitContext>();
  buildingIDs.forEach((buildingID) => {
    residentUnitDetails.value
      .filter((unit) => getBindDigitsOnly(getBindUnitID(unit)).startsWith(buildingID))
      .forEach((unit) => {
        const unitID = getBindUnitID(unit);
        if (!unitID || unitMap.has(unitID)) {
          return;
        }
        unitMap.set(unitID, {
          unitID,
          buildingID,
          name: buildSubaccountUnitName(buildingID, unit),
        });
      });
  });

  return Array.from(unitMap.values()).sort((left, right) =>
    left.name.localeCompare(right.name, 'en', { numeric: true, sensitivity: 'base' }),
  );
};

// 7.9 按單位讀取授權副戶
const loadSubaccountRowsForGroup = async (group: SubaccountGroup): Promise<SubaccountGroup | null> => {
  try {
    const result = await fetchMemberIsmartSubaccounts(group.unitID);
    return {
      ...group,
      rows: (result.items ?? []).map((row, index) => buildSubaccountDisplayRow(row, group.name, index)),
      error: '',
    };
  } catch (error) {
    if (isForbiddenRequest(error)) {
      return null;
    }
    return {
      ...group,
      rows: [],
      error: readAccountApiErrorMessage(error, t('account.center.subaccounts.loadError')),
    };
  }
};

// 7.10 使用後端匯總資料建立授權副戶分組
const loadAggregateSubaccountGroups = async (): Promise<void> => {
  const result = await fetchMemberIsmartSubaccounts();
  const groups = new Map<string, SubaccountGroup>();
  (result.items ?? []).forEach((row, index) => {
    const unitID = toSubaccountText(row.unit_id, `aggregate-${index}`);
    const buildingName = toSubaccountText(row.building_name || row.building_id);
    const unitName = toSubaccountText(row.unit_display || row.unit_id, unitID);
    const groupName = [buildingName, unitName].filter(Boolean).join(' ') || unitID;
    const existingGroup = groups.get(unitID);
    if (existingGroup) {
      existingGroup.rows.push(buildSubaccountDisplayRow(row, groupName, existingGroup.rows.length));
      return;
    }
    groups.set(unitID, createSubaccountGroup({
      unitID,
      buildingID: toSubaccountText(row.building_id),
      name: groupName,
    }, [row]));
  });
  subaccountGroups.value = Array.from(groups.values());
};

// 7.11 載入授權副戶資料
const loadSubaccountGroups = async (): Promise<void> => {
  subaccountsLoading.value = true;
  subaccountsError.value = '';
  try {
    if (integrationAuthUnavailable.value) {
      subaccountGroups.value = [];
      subaccountsError.value = t('account.center.subaccounts.reloginIsmart');
      return;
    }
    const unitContexts = await loadSubaccountUnitContexts();
    if (unitContexts.length === 0) {
      await loadAggregateSubaccountGroups();
      return;
    }

    const groups = unitContexts.map((item) => createSubaccountGroup(item));
    const settledGroups = await Promise.all(groups.map((group) => loadSubaccountRowsForGroup(group)));
    subaccountGroups.value = settledGroups.filter((group): group is SubaccountGroup => Boolean(group));
  } catch (error) {
    subaccountGroups.value = [];
    subaccountsError.value = readAccountApiErrorMessage(error, t('account.center.subaccounts.loadError'));
  } finally {
    subaccountsLoading.value = false;
    subaccountsLoaded.value = true;
  }
};

// 7.12 重新整理單一授權副戶分組
const refreshSubaccountGroup = async (group: SubaccountGroup): Promise<void> => {
  const refreshedGroup = await loadSubaccountRowsForGroup(group);
  if (!refreshedGroup) {
    subaccountGroups.value = subaccountGroups.value.filter((item) => item.unitID !== group.unitID);
    return;
  }
  Object.assign(group, {
    rows: refreshedGroup.rows,
    error: refreshedGroup.error,
  });
};

// 7.13 切換新增授權表單
const toggleSubaccountGrantForm = (unitID: string): void => {
  activeSubaccountGrantUnitID.value = activeSubaccountGrantUnitID.value === unitID ? '' : unitID;
};

// 7.14 判斷新增授權是否可提交
const canSubmitSubaccountGrant = (group: SubaccountGroup): boolean =>
  /^\d+$/.test(group.grantTargetUserID.trim()) && !group.isGranting;

// 7.15 新增授權副戶
const handleGrantSubaccount = async (group: SubaccountGroup): Promise<void> => {
  const targetUserID = Number(group.grantTargetUserID.trim());
  if (!Number.isSafeInteger(targetUserID) || targetUserID <= 0) {
    feedbackStore.pushToast(t('account.center.subaccounts.invalidUserId'), 'error');
    return;
  }

  group.isGranting = true;
  try {
    await grantMemberIsmartSubaccount({
      unit_id: group.unitID,
      target_user_id: targetUserID,
      remark: group.grantRemark.trim() || undefined,
    });
    group.grantTargetUserID = '';
    group.grantRemark = '';
    activeSubaccountGrantUnitID.value = '';
    await refreshSubaccountGroup(group);
    feedbackStore.pushToast(t('account.center.subaccounts.grantSuccess'), 'success');
  } catch (error) {
    feedbackStore.pushToast(readAccountApiErrorMessage(error, t('account.center.subaccounts.grantError')), 'error');
  } finally {
    group.isGranting = false;
  }
};

// 7.16 撤銷授權副戶
const handleRevokeSubaccount = async (group: SubaccountGroup, row: SubaccountDisplayRow): Promise<void> => {
  const targetUserID = Number(row.targetUserID);
  if (!Number.isSafeInteger(targetUserID) || targetUserID <= 0) {
    feedbackStore.pushToast(t('account.center.subaccounts.invalidTarget'), 'error');
    return;
  }
  if (!window.confirm(t('account.center.subaccounts.revokeConfirm'))) {
    return;
  }

  revokingSubaccountKey.value = row.key;
  try {
    await revokeMemberIsmartSubaccount({
      unit_id: group.unitID,
      target_user_id: targetUserID,
    });
    await refreshSubaccountGroup(group);
    feedbackStore.pushToast(t('account.center.subaccounts.revokeSuccess'), 'success');
  } catch (error) {
    feedbackStore.pushToast(readAccountApiErrorMessage(error, t('account.center.subaccounts.revokeError')), 'error');
  } finally {
    revokingSubaccountKey.value = '';
  }
};

interface BindingStatusItem {
  key: string;
  title: string;
  meta: string;
  chip: string;
  chipType: 'good' | 'warn' | 'brand';
}

// 8. 物業綁定申請
const bindingSteps = computed(() => [
  {
    num: '1',
    title: t('account.center.binding.selectUnitStep'),
    desc: t('account.center.binding.selectUnitStepDescription'),
  },
  {
    num: '2',
    title: t('account.center.binding.detailsStep'),
    desc: t('account.center.binding.detailsStepDescription'),
  },
  {
    num: '3',
    title: t('account.center.binding.approvalStep'),
    desc: t('account.center.binding.approvalStepDescription'),
  },
]);

const bindingDocs = computed(() => [
  {
    title: t('account.center.binding.identityDocument'),
    desc: t('account.center.binding.identityDocumentDescription'),
  },
  {
    title: t('account.center.binding.relationshipDocument'),
    desc: t('account.center.binding.relationshipDocumentDescription'),
  },
  {
    title: t('account.center.binding.recentBill'),
    desc: t('account.center.binding.recentBillDescription'),
  },
  {
    title: t('account.center.binding.supportingDocument'),
    desc: t('account.center.binding.supportingDocumentDescription'),
  },
]);

const bindingRoleOptions = computed(() => [
  { value: '業主', label: t('account.center.binding.ownerRole') },
  { value: '住戶代表', label: t('account.center.binding.residentRole') },
  { value: '公司授權人', label: t('account.center.binding.companyRole') },
]);
const bindingBuildings = ref<PosBuilding[]>([]);
const bindingUnits = ref<PosBuildingUnit[]>([]);
const bindingBuildingID = ref('');
const bindingFloor = ref('');
const bindingUnitID = ref('');
const bindingRole = ref('業主');
const bindingApplicantName = ref('');
const bindingPhone = ref('');
const bindingEmail = ref('');
const bindingNote = ref('');
const bindingReceiveEmail = ref(true);
const bindingBuildingsLoading = ref(false);
const bindingUnitsLoading = ref(false);
const bindingSubmitting = ref(false);
const bindingStatusList = ref<BindingStatusItem[]>([]);
let latestBindingUnitsRequestID = 0;
let ownerBindingLoaded = false;
let ownerBindingLoadPromise: Promise<void> | null = null;
let isInitializingOwnerBinding = false;

// 8.1 取得物業綁定大廈選項
const bindingBuildingOptions = computed<BindOption[]>(() =>
  bindingBuildings.value
    .slice()
    .sort((left, right) => getBindBuildingLabel(left).localeCompare(getBindBuildingLabel(right), 'en', {
      numeric: true,
      sensitivity: 'base',
    }))
    .map((item) => ({ label: getBindBuildingLabel(item), value: getBindBuildingID(item) }))
    .filter((item) => item.label.length > 0 && item.value.length > 0),
);

// 8.2 取得物業綁定可選單位
const visibleBindingUnits = computed(() => bindingUnits.value);

// 8.3 取得物業綁定樓層與單位選項
const bindingFloorOptions = computed<BindOption[]>(() =>
  Array.from(new Set(visibleBindingUnits.value.map((item) => getBindDisplayUnitFloor(item)).filter(Boolean)))
    .sort(compareBindCodes)
    .map((floor) => ({ label: formatBindFloorLabel(floor), value: floor })),
);
const bindingUnitOptions = computed<BindOption[]>(() =>
  visibleBindingUnits.value
    .filter((item) => getBindDisplayUnitFloor(item) === bindingFloor.value)
    .slice()
    .sort((left, right) => compareBindCodes(getBindUnitName(left), getBindUnitName(right)))
    .map((item) => ({ label: getBindUnitName(item), value: getBindUnitID(item) }))
    .filter((item) => item.label.length > 0 && item.value.length > 0),
);

// 8.4 取得物業綁定目前選擇
const bindingSelectedBuildingName = computed(() =>
  bindingBuildingOptions.value.find((item) => item.value === bindingBuildingID.value)?.label ?? '',
);
const bindingSelectedUnit = computed(() =>
  visibleBindingUnits.value.find((item) => getBindUnitID(item) === bindingUnitID.value),
);
const bindingSelectedUnitName = computed(() =>
  bindingSelectedUnit.value ? getBindUnitName(bindingSelectedUnit.value) : '',
);
const bindingCurrent = computed(() =>
  [
    bindingSelectedBuildingName.value,
    bindingFloor.value ? formatBindFloorLabel(bindingFloor.value) : '',
    bindingSelectedUnitName.value,
  ].filter(Boolean).join(' / ') || t('account.center.common.noUnitSelected'),
);
const bindingRoleLabel = computed(() =>
  bindingRoleOptions.value.find((role) => role.value === bindingRole.value)?.label ?? bindingRole.value,
);
const hasPendingResidenceBinding = computed(() => sessionStore.me?.residence_binding_status === 'pending');
const canSubmitOwnerBinding = computed(() =>
  bindingBuildingID.value.length > 0 &&
  bindingFloor.value.length > 0 &&
  bindingUnitID.value.length > 0 &&
  bindingRole.value.trim().length > 0 &&
  bindingApplicantName.value.trim().length > 0 &&
  bindingPhone.value.trim().length > 0 &&
  !hasPendingResidenceBinding.value &&
  !bindingSubmitting.value,
);

// 8.5 同步物業綁定預設聯絡資料
const syncOwnerBindingContactFromProfile = (): void => {
  if (!bindingApplicantName.value.trim()) {
    bindingApplicantName.value = accountDisplayName.value === unsetText.value ? '' : accountDisplayName.value;
  }
  if (!bindingPhone.value.trim()) {
    bindingPhone.value = memberPhoneText.value === unsetText.value ? '' : memberPhoneText.value;
  }
  if (!bindingEmail.value.trim()) {
    bindingEmail.value = sessionStore.me?.email?.trim()
      || ismartProfile.value?.account_email?.trim()
      || sessionStore.me?.ismart_msg?.email?.trim()
      || '';
  }
};

// 8.6 同步物業綁定預設大廈
const syncOwnerBindingBuildingFromOptions = (): void => {
  if (bindingBuildingID.value || bindingBuildingOptions.value.length === 0) {
    return;
  }
  const preferredBuildingID = bindBuildingID.value && bindingBuildingOptions.value.some((item) => item.value === bindBuildingID.value)
    ? bindBuildingID.value
    : bindingBuildingOptions.value[0]?.value ?? '';
  bindingBuildingID.value = preferredBuildingID;
};

// 8.7 載入物業綁定大廈
const loadOwnerBindingBuildings = async (): Promise<void> => {
  bindingBuildingsLoading.value = true;
  try {
    bindingBuildings.value = await fetchPosBuildings();
  } catch {
    bindingBuildings.value = visibleBindBuildings.value;
    if (bindingBuildings.value.length === 0) {
      feedbackStore.pushToast(t('account.center.binding.buildingLoadError'), 'error');
    }
  } finally {
    bindingBuildingsLoading.value = false;
  }
};

// 8.8 載入物業綁定單位
const loadOwnerBindingUnits = async (buildingID: string): Promise<void> => {
  const requestID = ++latestBindingUnitsRequestID;
  const value = buildingID.trim();
  if (!value) {
    bindingUnits.value = [];
    return;
  }

  bindingUnitsLoading.value = true;
  try {
    const result = await fetchPosBuildingUnits(value);
    if (requestID !== latestBindingUnitsRequestID) {
      return;
    }
    bindingUnits.value = result.filter((item) => isSelectableBindUnit(value, item));
    if (bindingFloor.value && !bindingFloorOptions.value.some((item) => item.value === bindingFloor.value)) {
      bindingFloor.value = '';
      bindingUnitID.value = '';
    }
    if (bindingUnitID.value && !visibleBindingUnits.value.some((item) => getBindUnitID(item) === bindingUnitID.value)) {
      bindingUnitID.value = '';
    }
  } catch {
    if (requestID !== latestBindingUnitsRequestID) {
      return;
    }
    bindingUnits.value = [];
    feedbackStore.pushToast(t('account.center.binding.unitLoadError'), 'error');
  } finally {
    if (requestID === latestBindingUnitsRequestID) {
      bindingUnitsLoading.value = false;
    }
  }
};

// 8.9 首次進入物業綁定面板時才載入公共大廈與單位
const ensureOwnerBindingLoaded = async (): Promise<void> => {
  if (ownerBindingLoaded) {
    return;
  }
  if (ownerBindingLoadPromise) {
    await ownerBindingLoadPromise;
    return;
  }

  ownerBindingLoadPromise = (async () => {
    isInitializingOwnerBinding = true;
    try {
      syncOwnerBindingContactFromProfile();
      await loadOwnerBindingBuildings();
      syncOwnerBindingBuildingFromOptions();
      if (bindingBuildingID.value) {
        await loadOwnerBindingUnits(bindingBuildingID.value);
      }
      ownerBindingLoaded = true;
    } finally {
      isInitializingOwnerBinding = false;
      ownerBindingLoadPromise = null;
    }
  })();
  await ownerBindingLoadPromise;
};

// 8.10 轉換物業綁定可選欄位
const optionalBindingValue = (value: string): string | undefined => {
  const text = value.trim();
  return text || undefined;
};

// 8.11 提交物業綁定申請
const handleSubmitOwnerBindingRequest = async (): Promise<void> => {
  if (hasPendingResidenceBinding.value) {
    feedbackStore.pushToast(t('account.center.binding.pendingSubmitBlocked'), 'error');
    return;
  }

  if (!canSubmitOwnerBinding.value) {
    feedbackStore.pushToast(t('account.center.binding.requiredFields'), 'error');
    return;
  }

  bindingSubmitting.value = true;
  try {
    await submitMemberIsmartOwnerBindingRequest({
      building_id: bindingBuildingID.value,
      ownedflat: [bindingUnitID.value],
      cli_role: bindingRole.value.trim(),
      ownernote: optionalBindingValue(bindingNote.value),
      is_receive_email: bindingReceiveEmail.value,
      reg_tel: optionalBindingValue(bindingPhone.value),
      reg_email: optionalBindingValue(bindingEmail.value),
      cli_name: optionalBindingValue(bindingApplicantName.value),
      cli_tel: optionalBindingValue(bindingPhone.value),
    });

    const submittedAt = new Date().toLocaleDateString(preferenceStore.locale, {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    });
    const pendingStatus: BindingStatusItem = {
      key: `${bindingUnitID.value}-${Date.now()}`,
      title: [bindingSelectedBuildingName.value || bindingBuildingID.value, bindingFloor.value, bindingSelectedUnitName.value || bindingUnitID.value]
        .filter(Boolean)
        .join(' / '),
      meta: t('account.center.binding.statusMeta', { role: bindingRoleLabel.value, date: submittedAt }),
      chip: t('account.center.binding.pending'),
      chipType: 'warn',
    };
    bindingStatusList.value = [pendingStatus, ...bindingStatusList.value].slice(0, 5);
    feedbackStore.pushToast(t('account.center.binding.submitSuccess'), 'success');
  } catch (error) {
    feedbackStore.pushToast(readAccountApiErrorMessage(error, t('account.center.binding.submitError')), 'error');
  } finally {
    bindingSubmitting.value = false;
  }
};

// 9. 住宅 mock 資料
const homeListings = computed(() => [
  {
    name: '康睦庭園第二座 / 02 / D',
    identity: t('marketplace.management.staffKicker'),
    status: t('account.center.homes.linked'),
    chipType: 'good',
    updatedAt: new Intl.DateTimeFormat(preferenceStore.locale, { dateStyle: 'medium' }).format(new Date('2026-06-05')),
  },
  {
    name: 'Harbour Residence',
    identity: t('account.center.homes.applicant'),
    status: t('account.center.homes.pending'),
    chipType: 'warn',
    updatedAt: new Intl.DateTimeFormat(preferenceStore.locale, { dateStyle: 'medium' }).format(new Date('2026-05-22')),
  },
]);

// 10. 收藏 mock 資料
const savedItems = computed(() => [
  {
    name: '佐敦高級住宅',
    category: t('account.center.favorites.property'),
    price: t('account.center.favorites.monthlyPrice', {
      price: new Intl.NumberFormat(preferenceStore.locale, { style: 'currency', currency: 'HKD' }).format(36000),
    }),
    savedAt: t('account.center.favorites.today'),
  },
  {
    name: '纖柔牙刷 精巧頭 3支裝',
    category: t('account.center.favorites.offers'),
    price: new Intl.NumberFormat(preferenceStore.locale, {
      style: 'currency',
      currency: 'HKD',
      minimumFractionDigits: 2,
    }).format(14.5),
    savedAt: t('account.center.favorites.yesterday'),
  },
]);

// 11. 退出登入
const handleLogout = async (): Promise<void> => {
  if (isSigningOut.value) {
    return;
  }

  isSigningOut.value = true;

  try {
    await sessionStore.signOut();
    await router.push('/login');
  } finally {
    isSigningOut.value = false;
  }
};

// 12. 前往真實我的家具列表
const openFurnitureListings = () => {
  router.push('/account/listings');
};

// 13. 前往家具發布頁
const openFurnitureCreate = () => {
  router.push('/account/listings/new');
};

// 13. 按目前面板首次載入非帳號管理資料
const ensureActivePanelLoaded = async (): Promise<void> => {
  if (activePanel.value === 'profile-subaccounts' && !subaccountsLoaded.value && !subaccountsLoading.value) {
    await loadSubaccountGroups();
    return;
  }
  if (activePanel.value === 'profile-property-binding') {
    await ensureOwnerBindingLoaded();
  }
};

onMounted(async () => {
  try {
    if (!sessionStore.me) {
      await sessionStore.loadCurrentUser();
    }
    syncBindFromProfile();
    await loadBindBuildings();
    await loadResidentUnitDetails();
    if (bindBuildingID.value) {
      await loadBindUnits(bindBuildingID.value);
      syncBindUnitFromProfile();
    }
    await ensureActivePanelLoaded();
  } finally {
    isSyncingBindProfile = false;
  }
});

watch(bindBuildingID, (nextValue, previousValue) => {
  if (nextValue === previousValue) {
    return;
  }
  if (isSyncingBindProfile) {
    return;
  }
  bindFloor.value = '';
  bindUnitID.value = '';
  void loadBindUnits(nextValue);
});

watch(bindFloor, (nextValue, previousValue) => {
  if (nextValue !== previousValue && !isSyncingBindProfile) {
    bindUnitID.value = '';
  }
});

watch(bindingBuildingID, (nextValue, previousValue) => {
  if (nextValue === previousValue) {
    return;
  }
  bindingFloor.value = '';
  bindingUnitID.value = '';
  if (isInitializingOwnerBinding) {
    return;
  }
  void loadOwnerBindingUnits(nextValue);
});

watch(bindingFloor, (nextValue, previousValue) => {
  if (nextValue !== previousValue) {
    bindingUnitID.value = '';
  }
});

watch(
  bindingBuildingOptions,
  () => {
    syncOwnerBindingBuildingFromOptions();
  },
);

watch(
  [() => route.path, () => route.query.panel],
  ([path]) => {
    syncActivePanelFromRoute(path);
  },
  { immediate: true },
);

watch(activePanel, () => {
  void ensureActivePanelLoaded();
});
</script>

<template>
  <div
    class="page"
    id="page-profile"
  >
    <div class="work-shell">
      <!-- 1. 左側導覽 -->
      <aside class="work-sidebar">
        <h1>{{ t('account.center.nav.title') }}</h1>
        <nav class="work-nav">
          <button
            v-for="item in navItems"
            :key="item.key"
            type="button"
            class="work-nav-item"
            :class="{ on: activePanel === item.key }"
            @click="switchPanel(item.key)"
          >
            <span class="work-nav-label">{{ item.label }}</span>
            <span v-if="item.needsApi" class="work-nav-note">{{ t('account.center.nav.apiRequired') }}</span>
          </button>
        </nav>
      </aside>

      <!-- 2. 右側內容 -->
      <main class="work-main">
        <!-- 2.1 帳號管理 -->
        <div v-show="activePanel === 'profile-account'" class="work-panel on" data-work-panel="profile-account">
          <section class="work-hero">
            <div class="work-account-topbar">
              <div class="work-account-head">
                <div class="work-account-user">
                  <div class="work-account-avatar">{{ accountAvatarText }}</div>
                  <div>
                    <div class="work-account-name">{{ accountDisplayName }}</div>
                    <div class="work-account-sub">{{ t('account.center.account.personalProfile') }}</div>
                  </div>
                </div>
              </div>
              <div class="work-account-actions">
                <button
                  type="button"
                  class="work-account-btn outline"
                  :disabled="isSigningOut"
                  @click="handleLogout"
                >
                  {{ isSigningOut ? t('account.center.account.signingOut') : t('account.actions.signOut') }}
                </button>
                <button type="button" class="work-account-btn primary">{{ t('account.center.account.editProfile') }}</button>
                <button type="button" class="work-account-btn">{{ t('account.center.account.updateIsmart') }}</button>
              </div>
            </div>
          </section>

          <section class="work-account-grid">
            <div class="work-account-card">
              <div class="work-card-title">{{ t('account.center.account.accountInfo') }}</div>
              <div
                v-for="item in ajoAccountData"
                :key="item.label"
                class="work-account-field"
              >
                <span>{{ item.label }}</span>
                <strong>{{ item.value }}</strong>
              </div>
            </div>
            <div class="work-account-card">
              <div class="work-card-title">{{ t('account.center.account.identityStatus') }}</div>
              <div class="work-account-field"><span>{{ t('marketplace.myProfile.memberId') }}</span><strong>{{ identityStatus.memberNo }}</strong></div>
              <div class="work-account-field"><span>{{ t('marketplace.myProfile.memberStatus') }}</span><strong>{{ identityStatus.memberStatus }}</strong></div>
              <div class="work-account-field"><span>{{ t('marketplace.myProfile.memberType') }}</span><strong>{{ identityStatus.memberType }}</strong></div>
              <div class="work-account-field"><span>{{ t('marketplace.myProfile.primaryRole') }}</span><strong>{{ identityStatus.primaryRole }}</strong></div>
              <div class="work-account-field"><span>{{ t('marketplace.myProfile.profileCompleted') }}</span><strong>{{ identityStatus.profileCompleteness }}</strong></div>
              <div class="work-account-field"><span>{{ t('marketplace.myProfile.staffAccess') }}</span><strong>{{ identityStatus.staffPermission }}</strong></div>
              <div class="work-role-strip">
                <span
                  v-for="role in accountRoles"
                  :key="role"
                  class="work-role-badge"
                >
                  {{ role }}
                </span>
                <span class="work-role-tag">{{ accountPermissionText }}</span>
              </div>
            </div>
          </section>

          <section class="work-card" style="margin-top:14px;">
            <div class="work-card-title">{{ t('account.center.account.linkedUnit') }}</div>
            <div class="work-bind-wrap">
              <div class="work-bind-current">{{ bindCurrent }}</div>
              <div class="work-bind-control-row">
                <div class="work-bind-grid work-bind-grid--single">
                  <div class="work-bind-field">
                    <label class="work-bind-label">{{ t('building.context.property') }}</label>
                    <select
                      v-model="bindUnitID"
                      class="work-bind-select"
                      :disabled="bindLoading || bindUnitsLoading || bindPropertyOptions.length === 0"
                    >
                      <option value="">{{ bindLoading || bindUnitsLoading ? t('account.center.common.loading') : t('building.context.selectProperty') }}</option>
                      <option
                        v-for="item in bindPropertyOptions"
                        :key="item.value"
                        :value="item.value"
                      >
                        {{ item.label }}
                      </option>
                    </select>
                  </div>
                </div>
                <div class="work-bind-actions">
                  <button
                    type="button"
                    class="work-action work-compact-action"
                    :disabled="!canSaveBindUnit || bindSaving"
                    @click="handleSaveBindUnit"
                  >
                    {{ bindSaving ? t('account.center.common.saving') : t('account.profile.saveUnit') }}
                  </button>
                </div>
              </div>
            </div>
          </section>

          <section class="work-card" style="margin-top:14px;">
            <div class="work-card-title">{{ t('account.center.account.ismartData') }}</div>
            <div class="work-ismart-grid">
              <div class="work-ismart-card">
                <div class="work-card-sub">{{ t('account.center.account.ismartAccountData') }}</div>
                <div v-for="item in ismartAccountData" :key="item.label" class="work-row">
                  <div>
                    <strong>{{ item.label }}</strong>
                    <span>{{ item.value }}</span>
                  </div>
                </div>
              </div>
              <div class="work-ismart-card">
                <div class="work-card-sub">{{ t('account.center.account.ismartHouseholdData') }}</div>
                <div v-for="item in ismartHouseholdData" :key="item.label" class="work-row">
                  <div>
                    <strong>{{ item.label }}</strong>
                    <span>{{ item.value }}</span>
                  </div>
                </div>
              </div>
            </div>
          </section>

          <section class="work-card" style="margin-top:14px;">
            <div class="work-card-title">{{ t('account.center.account.reminderSettings') }}</div>
            <div class="work-setting-box">
              <label class="work-setting-check">
                <input type="checkbox" checked>
                <span>{{ t('account.center.account.receiveNoticeEmail') }}</span>
              </label>
              <button type="button" class="work-action work-compact-action">{{ t('account.center.common.submit') }}</button>
            </div>
          </section>
        </div>

        <!-- 2.2 地產代理公司 -->
        <div v-show="activePanel === 'profile-agency-company'" class="work-panel on" data-work-panel="profile-agency-company">
          <RouterView v-if="activePanel === 'profile-agency-company'" />
        </div>

        <!-- 2.3 授權副戶 -->
        <div v-show="activePanel === 'profile-subaccounts'" class="work-panel on" data-work-panel="profile-subaccounts">
          <section class="work-hero">
            <div>
              <div class="work-kicker">{{ t('account.center.subaccounts.kicker') }}</div>
              <h2 class="work-title">{{ t('account.center.subaccounts.title') }}</h2>
              <p class="work-desc">{{ t('account.center.subaccounts.description') }}</p>
            </div>
          </section>
          <section class="work-subaccount-group">
            <div v-if="subaccountsLoading" class="work-subaccount-state">{{ t('account.center.subaccounts.loading') }}</div>
            <div v-else-if="subaccountsError" class="work-subaccount-state error">
              <span>{{ subaccountsError }}</span>
              <button type="button" class="work-action secondary" @click="loadSubaccountGroups">{{ t('account.center.subaccounts.retry') }}</button>
            </div>
            <div v-else-if="subaccountsLoaded && subaccountGroups.length === 0" class="work-subaccount-state">
              {{ t('account.center.subaccounts.emptyGroups') }}
            </div>
            <template v-else>
              <div v-for="group in subaccountGroups" :key="group.unitID" class="work-subaccount-card">
                <div class="work-subaccount-head">
                  <div class="work-subaccount-name">{{ group.name }}</div>
                  <button
                    type="button"
                    class="work-action"
                    @click="toggleSubaccountGrantForm(group.unitID)"
                  >
                    {{
                      activeSubaccountGrantUnitID === group.unitID
                        ? t('account.center.subaccounts.collapse')
                        : t('account.center.subaccounts.addGrant')
                    }}
                  </button>
                </div>
                <div v-if="activeSubaccountGrantUnitID === group.unitID" class="work-subaccount-form">
                  <label>
                    <span>{{ t('account.center.subaccounts.userId') }}</span>
                    <input
                      v-model="group.grantTargetUserID"
                      type="number"
                      min="1"
                      inputmode="numeric"
                      :placeholder="t('account.center.subaccounts.userIdPlaceholder')"
                    >
                  </label>
                  <label>
                    <span>{{ t('account.center.subaccounts.remark') }}</span>
                    <input
                      v-model="group.grantRemark"
                      type="text"
                      maxlength="120"
                      :placeholder="t('account.center.subaccounts.optional')"
                    >
                  </label>
                  <div class="work-subaccount-form-actions">
                    <button
                      type="button"
                      class="work-action"
                      :disabled="!canSubmitSubaccountGrant(group)"
                      @click="handleGrantSubaccount(group)"
                    >
                      {{ group.isGranting ? t('account.center.common.submitting') : t('account.center.subaccounts.submitGrant') }}
                    </button>
                    <button
                      type="button"
                      class="work-action secondary"
                      :disabled="group.isGranting"
                      @click="toggleSubaccountGrantForm(group.unitID)"
                    >
                      {{ t('account.center.common.cancel') }}
                    </button>
                  </div>
                </div>
                <table class="work-table">
                  <thead>
                    <tr>
                      <th>{{ t('account.center.subaccounts.location') }}</th>
                      <th>{{ t('account.center.subaccounts.user') }}</th>
                      <th>{{ t('account.center.subaccounts.role') }}</th>
                      <th>{{ t('account.center.subaccounts.history') }}</th>
                      <th>{{ t('account.center.subaccounts.actions') }}</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="row in group.rows" :key="row.key">
                      <td>{{ row.location }}</td>
                      <td>
                        <strong>{{ row.user || t('account.center.subaccounts.unnamedUser') }}</strong>
                        <span v-if="row.meta" class="work-subaccount-user-meta">{{ row.meta }}</span>
                      </td>
                      <td>{{ row.role || t('account.center.subaccounts.authorisedUser') }}</td>
                      <td>{{ t('account.center.subaccounts.historyCount', { count: row.historyCount }) }}</td>
                      <td>
                        <button
                          type="button"
                          class="work-mini-btn"
                          :disabled="revokingSubaccountKey === row.key"
                          @click="handleRevokeSubaccount(group, row)"
                        >
                          {{
                            revokingSubaccountKey === row.key
                              ? t('account.center.subaccounts.revoking')
                              : t('account.center.subaccounts.revoke')
                          }}
                        </button>
                      </td>
                    </tr>
                    <tr v-if="group.error">
                      <td colspan="5" class="work-subaccount-empty">{{ group.error }}</td>
                    </tr>
                    <tr v-else-if="group.rows.length === 0">
                      <td colspan="5" class="work-subaccount-empty">{{ t('account.center.subaccounts.emptyRows') }}</td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </template>
          </section>
        </div>

        <!-- 2.3 物業綁定 -->
        <div v-show="activePanel === 'profile-property-binding'" class="work-panel on" data-work-panel="profile-property-binding">
          <section class="work-hero">
            <div>
              <div class="work-kicker">{{ t('account.center.binding.kicker') }}</div>
              <h2 class="work-title">{{ t('account.center.binding.title') }}</h2>
              <p class="work-desc">{{ t('account.center.binding.description') }}</p>
            </div>
          </section>
          <section class="work-card">
            <div class="work-card-title">{{ t('account.center.binding.processTitle') }}</div>
            <div class="binding-flow">
              <div v-for="step in bindingSteps" :key="step.num" class="binding-step">
                <div class="binding-step-num">{{ step.num }}</div>
                <div class="binding-step-title">{{ step.title }}</div>
                <div class="binding-step-desc">{{ step.desc }}</div>
              </div>
            </div>
          </section>
          <section class="work-card">
            <div class="work-card-title">{{ t('account.center.binding.reviewTitle') }}</div>
            <div class="binding-review-box">
              <strong>{{ t('account.center.binding.reviewResponsibility') }}</strong>
              <span>{{ t('account.center.binding.reviewDescription') }}</span>
            </div>
          </section>
          <section class="work-card">
            <div class="work-card-title">{{ t('account.center.binding.applicationTitle') }}</div>
            <div v-if="hasPendingResidenceBinding" class="binding-review-box binding-pending-state">
              <strong>{{ t('account.center.binding.pendingTitle') }}</strong>
              <span>{{ t('account.center.binding.pendingDescription') }}</span>
            </div>
            <div class="binding-platform-panel">
              <div class="binding-subtitle">{{ t('account.center.binding.applicationUnit') }}</div>
              <div class="work-bind-current">{{ bindingCurrent }}</div>
              <div class="acct-form-grid">
                <div class="acct-field">
                  <label>{{ t('account.center.common.building') }}</label>
                  <select
                    v-model="bindingBuildingID"
                    class="acct-select"
                    :disabled="hasPendingResidenceBinding || bindingBuildingsLoading || bindingBuildingOptions.length === 0"
                  >
                    <option value="">{{ bindingBuildingsLoading ? t('account.center.common.loading') : t('account.center.common.selectBuilding') }}</option>
                    <option
                      v-for="item in bindingBuildingOptions"
                      :key="item.value"
                      :value="item.value"
                    >
                      {{ item.label }}
                    </option>
                  </select>
                </div>
                <div class="acct-field">
                  <label>{{ t('account.center.common.floor') }}</label>
                  <select
                    v-model="bindingFloor"
                    class="acct-select"
                    :disabled="hasPendingResidenceBinding || !bindingBuildingID || bindingUnitsLoading || bindingFloorOptions.length === 0"
                  >
                    <option value="">{{ bindingUnitsLoading ? t('account.center.common.loading') : t('account.center.common.selectFloor') }}</option>
                    <option
                      v-for="item in bindingFloorOptions"
                      :key="item.value"
                      :value="item.value"
                    >
                      {{ item.label }}
                    </option>
                  </select>
                </div>
                <div class="acct-field">
                  <label>{{ t('account.center.common.unit') }}</label>
                  <select
                    v-model="bindingUnitID"
                    class="acct-select"
                    :disabled="hasPendingResidenceBinding || !bindingFloor || bindingUnitsLoading || bindingUnitOptions.length === 0"
                  >
                    <option value="">{{ bindingUnitsLoading ? t('account.center.common.loading') : t('account.center.common.selectUnit') }}</option>
                    <option
                      v-for="item in bindingUnitOptions"
                      :key="item.value"
                      :value="item.value"
                    >
                      {{ item.label }}
                    </option>
                  </select>
                </div>
                <div class="acct-field">
                  <label>{{ t('account.center.binding.applicantRole') }}</label>
                  <select v-model="bindingRole" class="acct-select" :disabled="hasPendingResidenceBinding">
                    <option
                      v-for="role in bindingRoleOptions"
                      :key="role.value"
                      :value="role.value"
                    >
                      {{ role.label }}
                    </option>
                  </select>
                </div>
              </div>
            </div>
            <div class="acct-form-grid binding-common-grid">
              <div class="acct-field">
                <label>{{ t('account.center.binding.applicantName') }}</label>
                <input v-model="bindingApplicantName" class="acct-input" type="text" autocomplete="name" :readonly="hasPendingResidenceBinding">
              </div>
              <div class="acct-field">
                <label>{{ t('account.center.binding.phone') }}</label>
                <input v-model="bindingPhone" class="acct-input" type="tel" autocomplete="tel" :readonly="hasPendingResidenceBinding">
              </div>
              <div class="acct-field full">
                <label>{{ t('account.center.binding.email') }}</label>
                <input v-model="bindingEmail" class="acct-input" type="email" autocomplete="email" :readonly="hasPendingResidenceBinding">
              </div>
              <div class="acct-field full">
                <label>{{ t('account.center.binding.remark') }}</label>
                <textarea
                  v-model="bindingNote"
                  class="acct-textarea"
                  :placeholder="t('account.center.binding.remarkPlaceholder')"
                  :readonly="hasPendingResidenceBinding"
                ></textarea>
              </div>
            </div>
            <label class="binding-check">
              <input v-model="bindingReceiveEmail" type="checkbox" :disabled="hasPendingResidenceBinding">
              <span>{{ t('account.center.binding.receiveEmail') }}</span>
            </label>
            <div class="work-bind-actions work-doc-actions">
              <button
                type="button"
                class="work-action"
                :disabled="!canSubmitOwnerBinding"
                @click="handleSubmitOwnerBindingRequest"
              >
                {{ bindingSubmitting
                  ? t('account.center.common.submitting')
                  : hasPendingResidenceBinding
                    ? t('account.center.binding.pending')
                    : t('account.center.binding.submitApproval') }}
              </button>
            </div>
          </section>
          <section class="work-card">
            <div class="work-card-title">{{ t('account.center.binding.requiredDocuments') }}</div>
            <div class="binding-doc-grid">
              <div v-for="doc in bindingDocs" :key="doc.title" class="binding-doc-card">
                <div class="binding-doc-title">{{ doc.title }}</div>
                <div class="binding-doc-desc">{{ doc.desc }}</div>
              </div>
            </div>
          </section>
          <section class="work-card">
            <div class="work-card-title">{{ t('account.center.binding.applicationStatus') }}</div>
            <div class="binding-status-list">
              <div v-if="hasPendingResidenceBinding" class="binding-status-item">
                <div>
                  <div class="binding-status-title">{{ t('account.center.binding.pendingTitle') }}</div>
                  <div class="binding-status-meta">{{ t('account.center.binding.pendingDescription') }}</div>
                </div>
                <span class="work-chip warn">{{ t('account.center.binding.pending') }}</span>
              </div>
              <div v-else-if="bindingStatusList.length === 0" class="binding-empty-state">
                {{ t('account.center.binding.noSubmission') }}
              </div>
              <div v-for="item in bindingStatusList" :key="item.key" class="binding-status-item">
                <div>
                  <div class="binding-status-title">{{ item.title }}</div>
                  <div class="binding-status-meta">{{ item.meta }}</div>
                </div>
                <span class="work-chip" :class="item.chipType">{{ item.chip }}</span>
              </div>
            </div>
          </section>
        </div>

        <!-- 2.4 AJO 錢包 -->
        <div v-show="activePanel === 'profile-wallet'" class="work-panel on" data-work-panel="profile-wallet">
          <RouterView v-if="shouldRenderRoutePanel && activePanel === 'profile-wallet'" />
          <AccountWalletPage v-else-if="activePanel === 'profile-wallet'" />
        </div>

        <!-- 2.5 訊息管理 -->
        <div v-show="activePanel === 'profile-chat'" class="work-panel on" data-work-panel="profile-chat">
          <RouterView v-if="activePanel === 'profile-chat'" />
        </div>

        <!-- 2.6 我的樓盤 -->
        <div v-show="activePanel === 'profile-properties'" class="work-panel on" data-work-panel="profile-properties">
          <RouterView v-if="shouldRenderRoutePanel && activePanel === 'profile-properties'" />
          <PropertyMyPage
            v-else-if="activePanel === 'profile-properties'"
            channel="sale"
            base-path="/account/properties/sale"
          />
        </div>

        <!-- 2.7 我的住宅 -->
        <div v-show="activePanel === 'profile-homes'" class="work-panel on" data-work-panel="profile-homes">
          <RouterView v-if="shouldRenderRoutePanel && activePanel === 'profile-homes'" />
          <template v-else>
          <section class="work-hero">
            <div>
              <div class="work-kicker">{{ t('account.center.homes.kicker') }}</div>
              <h2 class="work-title">{{ t('account.center.homes.title') }}</h2>
              <p class="work-desc">{{ t('account.center.homes.description') }}</p>
            </div>
          </section>
          <section class="work-card">
            <div class="work-card-title">{{ t('account.center.homes.listTitle') }}</div>
            <table class="work-table">
              <thead>
                <tr>
                  <th>{{ t('account.center.homes.residence') }}</th>
                  <th>{{ t('account.center.homes.identity') }}</th>
                  <th>{{ t('account.center.homes.status') }}</th>
                  <th>{{ t('account.center.homes.updatedAt') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(item, idx) in homeListings" :key="idx">
                  <td>{{ item.name }}</td>
                  <td>{{ item.identity }}</td>
                  <td><span class="work-chip" :class="item.chipType">{{ item.status }}</span></td>
                  <td>{{ item.updatedAt }}</td>
                </tr>
              </tbody>
            </table>
          </section>
          </template>
        </div>

        <!-- 2.8 我的家具 -->
        <div v-show="activePanel === 'profile-furniture'" class="work-panel on" data-work-panel="profile-furniture">
          <RouterView v-if="shouldRenderRoutePanel && activePanel === 'profile-furniture'" />
          <template v-else>
          <section class="work-hero">
            <div>
              <div class="work-kicker">{{ t('account.center.furniture.kicker') }}</div>
              <h2 class="work-title">{{ t('account.center.furniture.title') }}</h2>
              <p class="work-desc">{{ t('account.center.furniture.description') }}</p>
            </div>
            <button type="button" class="work-action" @click="openFurnitureCreate">{{ t('account.center.furniture.add') }}</button>
          </section>
          <section class="work-card">
            <div class="work-card-title">{{ t('account.center.furniture.listTitle') }}</div>
            <div class="work-account-field">
              <span>{{ t('account.center.furniture.dataSource') }}</span>
              <strong>{{ t('account.center.furniture.dataSourceDescription') }}</strong>
            </div>
            <div class="work-table-actions">
              <button type="button" class="work-mini-btn primary" @click="openFurnitureListings">{{ t('account.center.furniture.viewList') }}</button>
              <button type="button" class="work-mini-btn" @click="openFurnitureCreate">{{ t('account.center.furniture.add') }}</button>
            </div>
          </section>
          </template>
        </div>

        <!-- 2.9 我的收藏 -->
        <div v-show="activePanel === 'profile-saved'" class="work-panel on" data-work-panel="profile-saved">
          <RouterView v-if="shouldRenderRoutePanel && activePanel === 'profile-saved'" />
          <template v-else>
          <section class="work-hero">
            <div>
              <div class="work-kicker">{{ t('account.center.favorites.kicker') }}</div>
              <h2 class="work-title">{{ t('account.center.favorites.title') }}</h2>
              <p class="work-desc">{{ t('account.center.favorites.description') }}</p>
            </div>
          </section>
          <section class="work-card">
            <div class="work-card-title">{{ t('account.center.favorites.listTitle') }}</div>
            <table class="work-table">
              <thead>
                <tr>
                  <th>{{ t('account.center.favorites.item') }}</th>
                  <th>{{ t('account.center.favorites.category') }}</th>
                  <th>{{ t('account.center.favorites.price') }}</th>
                  <th>{{ t('account.center.favorites.savedAt') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(item, idx) in savedItems" :key="idx">
                  <td>{{ item.name }}</td>
                  <td>{{ item.category }}</td>
                  <td>{{ item.price }}</td>
                  <td>{{ item.savedAt }}</td>
                </tr>
              </tbody>
            </table>
          </section>
          </template>
        </div>

      </main>
    </div>
  </div>
</template>

<style scoped>
/*
 * 會員中心頁樣式。
 * 1. CSS 變量嚴格對齊 HTML 設計稿原生變量名（var(--brand)、var(--ink)、var(--bdr)、var(--sur) 等）。
 * 2. 桌面雙欄佈局，行動單欄響應式。
 */

/* 1. 頁面容器 */
.page {
  min-height: calc(100svh - var(--nav-h, 52px));
  background: var(--sur-2);
  color: var(--ink);
}

/* 2. 雙欄佈局 */
.work-shell {
  display: grid;
  grid-template-columns: 240px minmax(0, 1fr);
  gap: 16px;
  max-width: var(--layout-page-max-width);
  margin: 0 auto;
  padding: 12px 24px 16px;
  color: var(--ink);
}

/* 3. 左側導覽 */
.work-sidebar {
  position: sticky;
  top: 72px;
  align-self: start;
  border: 1px solid var(--bdr);
  border-radius: 8px;
  background: #fff;
  padding: 16px;
}

.work-sidebar h1 {
  margin: 0;
  color: var(--accent);
  font-family: var(--font-serif);
  font-size: 26px;
  font-weight: 400;
  line-height: 1.2;
}

.work-nav {
  display: grid;
  gap: 2px;
  margin-top: 18px;
}

.work-nav-item {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 2px 6px;
  min-height: 34px;
  border: 0;
  border-radius: 0;
  background: transparent;
  color: var(--ink-2);
  cursor: pointer;
  font-family: inherit;
  font-size: 13px;
  font-weight: 500;
  padding: 8px 2px;
  text-align: left;
  transition: background-color 0.15s ease, color 0.15s ease;
}

.work-nav-label {
  min-width: 0;
}

.work-nav-note {
  flex: 0 0 auto;
  color: var(--ink-3);
  font-size: 10px;
  font-weight: 600;
  line-height: 1.2;
  white-space: nowrap;
}

.work-nav-item::after {
  content: '';
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  height: 2px;
  background: var(--brand);
  transform: scaleX(0);
  transform-origin: left center;
  transition: transform 0.24s ease;
}

.work-nav-item:hover {
  background: var(--brand-light);
  color: var(--ink);
}

.work-nav-item.on,
.work-nav-item:focus-visible {
  background: transparent;
  color: var(--accent);
  font-weight: 700;
  outline: none;
}

.work-nav-item.on:hover,
.work-nav-item:focus-visible:hover {
  background: var(--brand-light);
}

.work-nav-item.on::after,
.work-nav-item:focus-visible::after {
  transform: scaleX(1);
}

/* 4. 右側主內容 */
.work-main {
  display: grid;
  gap: 12px;
  min-width: 0;
  align-content: start;
}

.work-panel {
  display: none;
}

.work-panel.on {
  display: grid;
  gap: 14px;
  align-content: start;
}

/* 5. 區段標題 */
.work-hero {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 18px;
  border: 0;
  border-bottom: 1px solid var(--bdr);
  border-radius: 0;
  background: transparent;
  margin: 0;
  padding: 0 0 12px;
}

.work-kicker {
  margin-bottom: 4px;
  color: var(--ink-3);
  font-size: 9px;
  letter-spacing: 1.4px;
  text-transform: uppercase;
}

.work-title {
  margin: 0;
  color: var(--ink);
  font-size: 20px;
  font-weight: 600;
  line-height: 1.25;
}

.work-desc {
  max-width: 560px;
  margin: 5px 0 0;
  color: var(--ink-3);
  font-size: 13px;
  line-height: 1.5;
}

/* 5.1 特定面板 work-hero 調整（對齊 HTML 設計稿 #page-profile 覆蓋） */
[data-work-panel="profile-account"] .work-hero {
  align-items: center;
  padding: 10px 0 12px;
}

[data-work-panel="profile-properties"] .work-hero,
[data-work-panel="profile-homes"] .work-hero,
[data-work-panel="profile-furniture"] .work-hero,
[data-work-panel="profile-saved"] .work-hero {
  padding-bottom: 10px;
}

[data-work-panel="profile-saved"] .work-hero {
  display: none;
}

[data-work-panel="profile-saved"] .work-card {
  margin-top: 0;
}

/* 6. 帳號頂欄 */
.work-account-topbar {
  display: flex;
  width: 100%;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
  flex-wrap: wrap;
}

.work-account-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
}

.work-account-user {
  display: flex;
  align-items: center;
  gap: 10px;
}

.work-account-avatar {
  display: flex;
  width: 42px;
  height: 42px;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  background: #fff0eb;
  border: 1px solid #eadbd4;
  color: var(--brand);
  font-size: 14px;
  font-weight: 800;
}

.work-account-name {
  color: var(--ink);
  font-size: 14px;
  font-weight: 800;
  line-height: 1.2;
}

.work-account-sub {
  margin-top: 3px;
  color: var(--ink-3);
  font-size: 12px;
  font-weight: 700;
}

.work-account-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.work-account-btn {
  display: inline-flex;
  min-height: 34px;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--bdr);
  border-radius: 6px;
  background: #fff;
  color: var(--ink);
  cursor: pointer;
  font-family: inherit;
  font-size: 12px;
  font-weight: 700;
  padding: 0 12px;
  line-height: 1;
}

.work-account-btn.outline {
  border-color: #f6c7b3;
  box-shadow: inset 0 0 0 1px #f6c7b3;
}

.work-account-btn.primary {
  border-color: var(--brand);
  background: var(--brand);
  color: #fff;
}

/* 7. 帳號資料網格 */
.work-account-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0;
  border: 1px solid var(--bdr);
  border-radius: 8px;
  background: #fff;
  overflow: hidden;
  margin-top: 12px;
}

.work-account-card {
  background: #fff;
  overflow: hidden;
}

.work-account-card:first-child {
  border-right: 1px solid var(--bdr);
}

.work-account-card .work-card-title {
  margin: 0;
  padding: 14px 16px 12px;
  color: var(--brand);
  font-size: 14px;
  font-weight: 800;
}

.work-account-field {
  display: grid;
  grid-template-columns: 150px minmax(0, 1fr);
  gap: 12px;
  border-top: 1px solid var(--sur-3);
  padding: 10px 16px;
}

.work-account-field:first-child {
  border-top: 0;
}

.work-account-field span {
  color: var(--ink-3);
  font-size: 12px;
  font-weight: 700;
  line-height: 1.5;
  letter-spacing: 0;
}

.work-account-field strong {
  color: var(--ink);
  font-size: 13px;
  font-weight: 700;
  line-height: 1.5;
  word-break: break-word;
}

.work-role-strip {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  padding: 10px 16px;
  border-top: 1px solid var(--sur-3);
}

.work-role-badge {
  display: inline-flex;
  align-items: center;
  border-radius: 4px;
  background: var(--brand);
  color: #fff;
  font-size: 12px;
  font-weight: 800;
  padding: 7px 10px;
}

.work-role-tag {
  display: inline-flex;
  align-items: center;
  border: 1px solid var(--bdr);
  border-radius: 4px;
  background: #fff;
  color: var(--ink-2);
  font-size: 12px;
  font-weight: 700;
  padding: 7px 10px;
}

/* 8. 綁定單位 */
.work-bind-wrap {
  display: grid;
  gap: 12px;
}

.work-bind-current {
  color: var(--ink);
  font-size: 14px;
  font-weight: 800;
}

.work-bind-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
}

.work-bind-grid--single {
  grid-template-columns: minmax(0, 1fr);
}

.work-bind-control-row {
  display: grid;
  grid-template-columns: minmax(0, 680px) auto;
  align-items: end;
  gap: 12px;
}

.work-bind-field {
  display: grid;
  gap: 8px;
}

.work-bind-label {
  color: var(--ink-3);
  font-size: 12px;
  font-weight: 700;
}

.work-bind-select {
  width: 100%;
  border: 1px solid var(--bdr);
  border-radius: 8px;
  background: #fff;
  color: var(--ink);
  font-family: inherit;
  font-size: 14px;
  font-weight: 700;
  padding: 14px 16px;
}

.work-bind-actions {
  display: flex;
  justify-content: flex-end;
}

/* 9. iSmart 卡片 */
.work-ismart-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
  margin-top: 12px;
}

.work-ismart-card {
  border: 1px solid var(--bdr);
  border-radius: 8px;
  background: var(--sur);
  padding: 14px;
}

.work-ismart-card .work-card-sub {
  font-size: 13px;
  font-weight: 800;
  color: var(--ink);
  margin-bottom: 6px;
}

.work-ismart-card .work-row {
  padding: 10px 0;
}

.work-ismart-card .work-row > div {
  display: grid;
  grid-template-columns: 150px minmax(0, 1fr);
  gap: 12px;
  width: 100%;
}

.work-ismart-card .work-row strong {
  font-size: 12px;
  font-weight: 700;
  color: var(--ink-3);
  line-height: 1.5;
}

.work-ismart-card .work-row span {
  margin-top: 0;
  font-size: 13px;
  font-weight: 700;
  color: var(--ink);
  line-height: 1.5;
  word-break: break-word;
}

/* 10. 通用行 */
.work-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  border-top: 1px solid var(--sur-3);
  padding: 12px 0;
}

.work-row:first-child {
  border-top: 0;
  padding-top: 0;
}

.work-row:last-child {
  padding-bottom: 0;
}

.work-row strong {
  display: block;
  color: var(--ink);
  font-size: 13px;
  font-weight: 600;
}

.work-row span {
  display: block;
  margin-top: 3px;
  color: var(--ink-3);
  font-size: 12px;
  line-height: 1.5;
}

/* 11. 通用按鈕 */
.work-action {
  border: 0;
  border-radius: 6px;
  background: var(--accent);
  color: #fff;
  cursor: pointer;
  font-family: inherit;
  font-size: 12px;
  font-weight: 600;
  padding: 10px 14px;
  white-space: nowrap;
}

.work-action.secondary {
  border: 1px solid var(--bdr);
  background: #fff;
  color: var(--ink);
}

.work-action:disabled,
.work-mini-btn:disabled {
  cursor: not-allowed;
  opacity: 0.55;
}

.work-compact-action {
  min-height: 38px;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 700;
  padding: 9px 14px;
}

.work-mini-btn {
  border: 1px solid var(--bdr);
  border-radius: 6px;
  background: var(--sur);
  color: var(--ink);
  cursor: pointer;
  font-family: inherit;
  font-size: 12px;
  font-weight: 600;
  padding: 7px 10px;
}

.work-mini-btn.primary {
  border-color: var(--brand);
  background: var(--brand);
  color: #fff;
}

/* 12. 卡片 */
.work-card {
  border: 1px solid var(--bdr);
  border-radius: 8px;
  background: #fff;
  padding: 16px;
}

.work-card-title {
  margin-bottom: 10px;
  color: var(--ink);
  font-size: 14px;
  font-weight: 600;
}

.work-card-sub {
  color: var(--ink-3);
  font-size: 12px;
  line-height: 1.7;
}

/* 13. 統計 */
.work-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
}

.work-stat {
  color: var(--accent);
  font-size: 24px;
  font-weight: 700;
  line-height: 1.1;
}

.work-stat-label {
  margin-top: 6px;
  color: var(--ink-3);
  font-size: 12px;
}

/* 14. 狀態標籤 */
.work-chip {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  background: var(--sur-2);
  color: var(--ink-2);
  font-size: 11px;
  font-weight: 600;
  padding: 5px 9px;
  white-space: nowrap;
}

.work-chip.good {
  background: var(--success-bg);
  color: var(--success);
}

.work-chip.warn {
  background: var(--warning-bg);
  color: var(--warning);
}

.work-chip.brand {
  background: var(--brand-light);
  color: var(--accent);
}

/* 15. 表格 */
.work-table-wrap {
  overflow: auto;
}

.work-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 12px;
}

.work-table th {
  border-bottom: 1px solid var(--bdr);
  color: var(--ink-3);
  font-weight: 500;
  padding: 10px;
  text-align: left;
}

.work-table td {
  border-bottom: 1px solid var(--sur-3);
  padding: 12px 10px;
  color: var(--ink-2);
}

.work-table-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.work-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
}

.work-search {
  flex: 1;
  border: 1px solid var(--bdr);
  border-radius: 6px;
  background: var(--sur);
  color: var(--ink);
  font-family: inherit;
  font-size: 13px;
  padding: 10px 12px;
}

/* 16. 授權副戶 */
.work-subaccount-group {
  display: grid;
  gap: 14px;
}

.work-subaccount-card {
  border: 1px solid var(--bdr);
  border-radius: 8px;
  background: var(--sur);
  padding: 14px;
}

.work-subaccount-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
}

.work-subaccount-name {
  color: var(--ink);
  font-size: 18px;
  font-weight: 700;
}

.work-subaccount-state {
  border: 1px solid var(--bdr);
  border-radius: 8px;
  background: var(--sur);
  color: var(--ink-3);
  font-size: 13px;
  font-weight: 600;
  padding: 18px;
}

.work-subaccount-state.error {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  color: var(--danger);
}

.work-subaccount-form {
  display: grid;
  grid-template-columns: minmax(160px, 220px) minmax(180px, 1fr) auto;
  align-items: end;
  gap: 10px;
  border: 1px solid var(--bdr);
  border-radius: 8px;
  background: #fff;
  padding: 12px;
  margin-bottom: 12px;
}

.work-subaccount-form label {
  display: grid;
  gap: 6px;
  color: var(--ink-3);
  font-size: 12px;
  font-weight: 700;
}

.work-subaccount-form input {
  width: 100%;
  border: 1px solid var(--bdr);
  border-radius: 6px;
  background: var(--sur);
  color: var(--ink);
  font: inherit;
  padding: 9px 10px;
}

.work-subaccount-form-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.work-subaccount-user-meta {
  display: block;
  margin-top: 4px;
  color: var(--ink-3);
  font-size: 11px;
  line-height: 1.4;
}

.work-subaccount-empty {
  color: var(--ink-3);
  font-size: 12px;
  font-weight: 600;
  text-align: center;
  padding: 18px 12px;
}

/* 17. 提示設定 */
.work-setting-box {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
  border: 1px solid var(--bdr);
  border-radius: 8px;
  background: var(--sur);
  padding: 14px;
  margin-top: 12px;
}

.work-setting-check {
  display: flex;
  align-items: center;
  gap: 10px;
  color: var(--ink);
  font-size: 14px;
  font-weight: 700;
}

.work-setting-check input {
  width: 16px;
  height: 16px;
  accent-color: var(--brand);
}

/* 18. 物業綁定流程 */
.binding-flow {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 10px;
}

.binding-step {
  border: 1px solid var(--bdr);
  border-radius: 8px;
  background: #fff;
  padding: 14px;
}

.binding-step-num {
  display: flex;
  width: 26px;
  height: 26px;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  background: var(--brand);
  color: #fff;
  font-size: 12px;
  font-weight: 800;
  margin-bottom: 10px;
}

.binding-step-title {
  color: var(--ink);
  font-size: 13px;
  font-weight: 800;
  margin-bottom: 5px;
}

.binding-step-desc {
  color: var(--ink-3);
  font-size: 12px;
  line-height: 1.6;
}

/* 19. 審批說明 */
.binding-review-box {
  display: grid;
  gap: 6px;
  border: 1px solid var(--bdr);
  border-radius: 8px;
  background: var(--sur-2);
  padding: 12px 14px;
}

.binding-review-box strong {
  color: var(--ink);
  font-size: 12px;
  font-weight: 800;
}

.binding-review-box span {
  color: var(--ink-3);
  font-size: 12px;
  line-height: 1.6;
}

/* 20. 綁定申請表單 */
.binding-platform-panel {
  display: block;
}

.binding-subtitle {
  color: var(--ink);
  font-size: 12px;
  font-weight: 800;
  margin: 0 0 10px;
}

.binding-subtitle:not(:first-child) {
  margin-top: 16px;
}

.binding-common-grid {
  margin-top: 14px;
}

.binding-check {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: 14px;
  color: var(--ink);
  font-size: 13px;
  font-weight: 700;
}

.binding-check input {
  width: 16px;
  height: 16px;
  accent-color: var(--brand);
}

/* 21. 文件說明 */
.binding-doc-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.binding-doc-card {
  border: 1px dashed var(--bdr-2);
  border-radius: 8px;
  background: var(--sur-2);
  padding: 14px;
}

.binding-doc-title {
  color: var(--ink);
  font-size: 13px;
  font-weight: 800;
  margin-bottom: 5px;
}

.binding-doc-desc {
  color: var(--ink-3);
  font-size: 12px;
  line-height: 1.6;
  margin-bottom: 10px;
}

.work-doc-actions {
  margin-top: 14px;
}

/* 22. 申請狀態 */
.binding-status-list {
  display: grid;
  gap: 10px;
}

.binding-status-item {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 12px;
  align-items: center;
  border: 1px solid var(--bdr);
  border-radius: 8px;
  background: #fff;
  padding: 12px 14px;
}

.binding-status-title {
  color: var(--ink);
  font-size: 13px;
  font-weight: 800;
}

.binding-status-meta {
  color: var(--ink-3);
  font-size: 12px;
  margin-top: 4px;
}

.binding-empty-state {
  border: 1px dashed var(--bdr-2);
  border-radius: 8px;
  background: var(--sur-2);
  color: var(--ink-3);
  font-size: 12px;
  font-weight: 700;
  padding: 14px;
}

/* 23. 表單元件 */
.acct-form-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px;
}

.acct-field {
  display: grid;
  gap: 7px;
}

.acct-field.full {
  grid-column: 1 / -1;
}

.acct-field label {
  color: var(--ink);
  font-size: 12px;
  font-weight: 800;
}

.acct-input,
.acct-select,
.acct-textarea {
  width: 100%;
  border: 1px solid var(--bdr);
  border-radius: 6px;
  background: #fff;
  color: var(--ink);
  font-family: inherit;
  font-size: 13px;
  font-weight: 600;
  padding: 10px 12px;
}

.acct-textarea {
  min-height: 86px;
  resize: vertical;
  line-height: 1.6;
}

/* 24. 響應式 - 平板 */
@media (max-width: 1023px) {
  .work-shell {
    grid-template-columns: 1fr;
    padding: 14px;
  }

  .work-sidebar {
    position: static;
  }

  .work-nav {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .work-grid,
  .work-ismart-grid,
  .work-account-grid,
  .work-bind-grid,
  .binding-flow,
  .binding-doc-grid {
    grid-template-columns: 1fr;
  }

  .work-account-card:first-child {
    border-right: 0;
    border-bottom: 1px solid var(--bdr);
  }

  .acct-form-grid {
    grid-template-columns: 1fr;
  }

  .work-subaccount-form {
    grid-template-columns: 1fr;
  }
}

/* 25. 響應式 - 行動 */
@media (max-width: 560px) {
  .work-hero {
    align-items: flex-start;
    flex-direction: column;
  }

  .work-nav {
    grid-template-columns: 1fr;
  }

  .work-bind-control-row {
    grid-template-columns: 1fr;
  }

  .work-bind-actions,
  .work-bind-actions .work-action {
    width: 100%;
  }

}
</style>

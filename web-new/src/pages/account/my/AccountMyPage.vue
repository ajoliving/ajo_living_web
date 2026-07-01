<!--
 * 會員中心頁。
 * 1. 承載左側模組導覽與右側內容面板，支援帳號管理、授權副戶、物業綁定、AJO 錢包、我的樓盤、我的住宅、我的家具與我的收藏等面板切換。
 * 2. 綁定單位讀取 POS 大廈與單位資料，其他面板沿用展示資料。
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
import { computed, onMounted, ref, watch } from 'vue';
import { RouterView, useRoute, useRouter } from 'vue-router';

import { fetchMemberPosBuildings, fetchMemberPosBuildingUnits } from '@/httpapis/building';
import { updateMe } from '@/httpapis/me';
import type { PosBuilding, PosBuildingUnit } from '@/model/community';
import { useFeedbackStore } from '@/stores/feedback';
import { useSessionStore } from '@/stores/session';
import AccountWalletPage from '@/pages/account/my/profile/wallet/Page.vue';
import PropertyMyPage from '@/pages/property/my/PropertyMyPage.vue';

const route = useRoute();
const router = useRouter();
const feedbackStore = useFeedbackStore();
const sessionStore = useSessionStore();

// 1. 面板索引類型
type PanelKey =
  | 'profile-account'
  | 'profile-subaccounts'
  | 'profile-property-binding'
  | 'profile-wallet'
  | 'profile-chat'
  | 'profile-properties'
  | 'profile-homes'
  | 'profile-furniture'
  | 'profile-saved';

// 2. 導覽項目（對齊 HTML data-work-target）
const navItems: { key: PanelKey; label: string; needsApi?: boolean }[] = [
  { key: 'profile-account', label: '帳號管理' },
  { key: 'profile-subaccounts', label: '授權副戶', needsApi: true },
  { key: 'profile-property-binding', label: '物業綁定', needsApi: true },
  { key: 'profile-wallet', label: 'AJO 錢包' },
  { key: 'profile-chat', label: '訊息管理' },
  { key: 'profile-properties', label: '我的樓盤' },
  { key: 'profile-homes', label: '我的住宅' },
  { key: 'profile-furniture', label: '我的家具' },
  { key: 'profile-saved', label: '我的收藏' },
];

// 3. 當前面板
const activePanel = ref<PanelKey>('profile-account');
const isSigningOut = ref(false);

const routePanelPaths = [
  '/account/profile/wallet',
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
    activePanel.value = 'profile-account';
  }
};

// 7. 帳號管理 mock 資料
const accountProfile = {
  displayName: 'patrick',
  ismartAccount: 'patrick',
  email: '尚未綁定電郵',
  phone: 'ismart 10',
  password: '未設定',
  publisherRole: '未設定',
  regionCode: 'unknown',
  managedBuildings:
    '康睦庭園第二座, 萬高大廈B座, 僑偉大廈, 利來大廈, 仁美大廈, 豐富大廈, 鴻英大廈, 東昇樓, 華園, 麗麗大廈, 協和大廈, 景暉閣, 新萬利大廈, 東山臺24號, 仁英大廈, 浣紗大廈, 東南大樓(77號), 東南大樓(75號), 華興大廈(269號), 華興大廈(271號), 得運大廈, 仁文大廈, 富澤軒, 天富大廈, 榮森工業第二大廈, 仁利大廈, 嘉樂苑, 南昌苑, 測試2大廈, 測試1大廈, 坪麗苑, 玉桂園(車位), 玉桂園(1座), 玉桂園(2座), 玉桂園(3座), 玉桂園(4座), 玉桂園(5座), 玉桂園(6座), 玉桂園(7座), 玉桂園(8座), 玉桂園(9座), 玉桂園(10座), 玉桂園(11座)',
};

const identityStatus = {
  memberNo: '01KRWKDTQY6DY1H650GB63AC66',
  memberStatus: 'active',
  memberType: 'user',
  primaryRole: 'staff',
  profileCompleteness: '已完成',
  staffPermission: '已啟用',
};

interface BindOption {
  label: string;
  value: string;
}

const bindBuildings = ref<PosBuilding[]>([]);
const bindUnits = ref<PosBuildingUnit[]>([]);
const bindBuildingID = ref('');
const bindFloor = ref('');
const bindUnitID = ref('');
const bindLoading = ref(false);
const bindUnitsLoading = ref(false);
const bindSaving = ref(false);
let latestBindUnitsRequestID = 0;
let isSyncingBindProfile = false;
const bindUnassignedFloor = '未指定樓層';

// 6.1 判斷是否為系統佔位電話帳號
const isSystemPhonePlaceholder = (countryCode?: string): boolean =>
  ['email', 'ismart'].includes((countryCode ?? '').trim().toLowerCase());

// 6.2 讀取 POS 大廈與單位欄位
const getBindBuildingID = (item: PosBuilding): string =>
  String(item.building_id ?? item.id ?? '').trim();
const getBindBuildingName = (item: PosBuilding): string =>
  String(item.buildname_chi ?? item.buildname ?? item.name ?? getBindBuildingID(item)).trim();
const getBindUnitID = (item: PosBuildingUnit): string =>
  String(item.unit_id ?? item.id ?? '').trim();
const getBindUnitFloor = (item: PosBuildingUnit): string =>
  String(item.floor ?? '').trim();
const getBindDisplayUnitFloor = (item: PosBuildingUnit): string =>
  getBindUnitFloor(item) || bindUnassignedFloor;
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

// 6.4 判斷 POS 單位是否可選
const isSelectableBindUnit = (buildingID: string, item: PosBuildingUnit): boolean => {
  const normalizedBuildingID = getBindDigitsOnly(buildingID).slice(0, 7);
  const normalizedUnitID = getBindDigitsOnly(getBindUnitID(item));
  const hasFloor = getBindUnitFloor(item).length > 0;
  const hasUnit = getBindUnitName(item).length > 0;
  return !(normalizedBuildingID && normalizedUnitID === normalizedBuildingID && !hasFloor && !hasUnit);
};

// 6.5 由 POS 權限 ID 建立可選單位
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

// 6.6 轉換 POS 權限碼
const toBindTwoDigitCode = (value: unknown): string => {
  const digits = getBindDigitsOnly(value);
  return digits ? digits.slice(-2).padStart(2, '0') : '';
};

// 6.7 取得單位權限碼
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

// 6.8 判斷單位是否符合 POS 權限
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

// 6.9 按權限過濾 POS 單位
const filterBindUnitsByPermission = (
  buildingID: string,
  units: PosBuildingUnit[],
  flatUnitPermissions: string[],
): PosBuildingUnit[] =>
  units
    .filter((item) => isSelectableBindUnit(buildingID, item))
    .filter((item) => matchesBindUnitPermission(buildingID, item, flatUnitPermissions));

// 6.10 POS 顯示排序
const compareBindCodes = (left: string, right: string): number =>
  left.localeCompare(right, 'en', {
    numeric: true,
    sensitivity: 'base',
  });

const bindIsPOSStaff = computed(() =>
  Boolean(sessionStore.me?.is_staff || sessionStore.me?.ismart_msg?.is_staff),
);
const bindAllowedBuildingIDs = computed(() => {
  const message = sessionStore.me?.ismart_msg;
  const primaryCommunityID = sessionStore.me?.primary_community?.public_id?.trim() ?? '';
  if (message) {
    if (bindIsPOSStaff.value) {
      const staffBuildings = normalizeBindList(message.staff_building_permissions);
      return staffBuildings.length > 0 ? staffBuildings : normalizeBindList(message.building);
    }
    const values = normalizeBindList(message.client_building_permissions);
    return values.length > 0 || !primaryCommunityID ? values : [primaryCommunityID];
  }
  const values = normalizeBindList(sessionStore.me?.bound_building_ids);
  return values.length > 0 || !primaryCommunityID ? values : [primaryCommunityID];
});
const bindAllowedUnitIDs = computed(() =>
  bindIsPOSStaff.value
    ? []
    : normalizeBindList(
      sessionStore.me?.ismart_msg?.client_building_flat_units_permissions ?? sessionStore.me?.bound_flat_unit_ids,
    ),
);
const visibleBindBuildings = computed(() => {
  const allowed = new Set(bindAllowedBuildingIDs.value);
  return bindBuildings.value.filter((item) => allowed.has(getBindBuildingID(item)));
});
const visibleBindUnits = computed(() => {
  if (bindIsPOSStaff.value) {
    return bindUnits.value;
  }
  return bindUnits.value.filter((item) =>
    matchesBindUnitPermission(bindBuildingID.value, item, bindAllowedUnitIDs.value),
  );
});
const bindBuildingOptions = computed<BindOption[]>(() =>
  visibleBindBuildings.value
    .slice()
    .sort((left, right) => getBindBuildingName(left).localeCompare(getBindBuildingName(right), 'en', {
      numeric: true,
      sensitivity: 'base',
    }))
    .map((item) => ({ label: getBindBuildingName(item), value: getBindBuildingID(item) })),
);
const bindFloorOptions = computed<BindOption[]>(() =>
  Array.from(new Set(visibleBindUnits.value.map((item) => getBindDisplayUnitFloor(item)).filter(Boolean)))
    .sort(compareBindCodes)
    .map((floor) => ({ label: floor, value: floor })),
);
const bindUnitOptions = computed<BindOption[]>(() =>
  visibleBindUnits.value
    .filter((item) => getBindDisplayUnitFloor(item) === bindFloor.value)
    .slice()
    .sort((left, right) => compareBindCodes(getBindUnitName(left), getBindUnitName(right)))
    .map((item) => ({ label: getBindUnitName(item), value: getBindUnitID(item) }))
    .filter((item) => item.label.length > 0 && item.value.length > 0),
);
const bindSelectedBuildingName = computed(() =>
  bindBuildingOptions.value.find((item) => item.value === bindBuildingID.value)?.label ?? '',
);
const bindSelectedUnit = computed(() =>
  visibleBindUnits.value.find((item) => getBindUnitID(item) === bindUnitID.value),
);
const bindSelectedUnitName = computed(() =>
  bindSelectedUnit.value ? getBindUnitName(bindSelectedUnit.value) : '',
);
const bindCurrent = computed(() =>
  [bindSelectedBuildingName.value, bindFloor.value, bindSelectedUnitName.value].filter(Boolean).join(' / ') || '尚未選擇單位',
);
const canSaveBindUnit = computed(() =>
  bindBuildingID.value.length > 0 &&
  bindFloor.value.length > 0 &&
  bindUnitID.value.length > 0 &&
  Boolean(bindSelectedUnit.value),
);

// 6.6 同步會員已保存單位
const syncBindFromProfile = (): void => {
  isSyncingBindProfile = true;
  const profileBuildingID = sessionStore.me?.primary_community?.public_id?.trim() ?? '';
  bindBuildingID.value = bindAllowedBuildingIDs.value.includes(profileBuildingID)
    ? profileBuildingID
    : bindAllowedBuildingIDs.value[0] ?? '';
  bindFloor.value = sessionStore.me?.residence_floor || '';
  bindUnitID.value = '';
};

// 6.7 按已保存樓層與單位名稱同步 POS 單位
const syncBindUnitFromProfile = (): void => {
  const floor = sessionStore.me?.residence_floor?.trim() ?? '';
  const unitName = sessionStore.me?.residence_unit?.trim() ?? '';
  if (!floor || !unitName) {
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
    bindBuildings.value = await fetchMemberPosBuildings();
  } catch {
    const fallbackBuildings = bindAllowedBuildingIDs.value.map((buildingID) => ({
      building_id: buildingID,
      buildname: buildingID,
    }));
    bindBuildings.value = fallbackBuildings;
    if (fallbackBuildings.length === 0) {
      feedbackStore.pushToast('大廈資料載入失敗。', 'error');
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

  bindUnitsLoading.value = true;
  try {
    const result = await fetchMemberPosBuildingUnits(value);
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
      feedbackStore.pushToast('單位資料載入失敗。', 'error');
    }
  } finally {
    if (requestID === latestBindUnitsRequestID) {
      bindUnitsLoading.value = false;
    }
  }
};

// 6.9 儲存會員繳費單位
const handleSaveBindUnit = async (): Promise<void> => {
  if (!canSaveBindUnit.value) {
    feedbackStore.pushToast('請先選擇大廈、樓層與單位。', 'error');
    return;
  }

  bindSaving.value = true;
  try {
    const { data } = await updateMe({
      display_name: sessionStore.me?.display_name ?? sessionStore.currentUser.display_name,
      ...(sessionStore.me?.email ? { email: sessionStore.me.email } : {}),
      ...(isSystemPhonePlaceholder(sessionStore.me?.phone_country_code)
        ? {}
        : {
            phone_country_code: sessionStore.me?.phone_country_code ?? '',
            phone_number: sessionStore.me?.phone_number ?? '',
          }),
      primary_community_id: bindBuildingID.value,
      primary_community_name: bindSelectedBuildingName.value,
      bound_building_ids: [bindBuildingID.value],
      bound_flat_unit_ids: [bindUnitID.value],
      residence_floor: bindFloor.value,
      residence_unit: bindSelectedUnitName.value,
      district_code: sessionStore.me?.district_code ?? '',
    });
    sessionStore.me = data.data;
    feedbackStore.pushToast('繳費單位已更新。', 'success');
  } catch {
    feedbackStore.pushToast('繳費單位儲存失敗，請稍後再試。', 'error');
  } finally {
    bindSaving.value = false;
  }
};

const ismartAccountData = [
  { label: '帳戶編號', value: 'SAWYER' },
  { label: '帳戶電話', value: '90771352' },
  { label: '帳戶電郵', value: 'sawyer@seventy2.hk' },
  { label: '業戶名稱(英)', value: 'CHEUNG CHUN HO' },
  { label: '業戶名稱(中)', value: '未設定' },
  { label: '證件編號', value: '********' },
];

const ismartHouseholdData = [
  { label: '法體類型', value: '自然人' },
  { label: '性別', value: 'M' },
  { label: '出生日期', value: '未設定' },
  { label: '聯絡人名稱', value: 'S' },
  { label: '聯絡人電話', value: '90771352' },
  { label: '帳單電郵', value: 'sawyer@seventy2.hk' },
  { label: '帳單地址', value: 'test@123' },
];

const relatedProperties = [
  { name: '界限大廈', status: '登記業主' },
  { name: '協和大廈', status: '法團' },
  { name: '華興大廈(271號)', status: '登記業主' },
  { name: '華興大廈(269號)', status: '登記業主' },
  { name: '仁英大廈 G 02', status: '登記業主' },
];

// 7. 授權副戶 mock 資料
const subaccountGroups = [
  {
    name: '華興大廈(271號)',
    rows: [
      { location: '華興大廈(271號)', user: 'The Incorporated Owners of No. 269, 271 Temple Street', type: '法團', permission: '只讀' },
    ],
  },
  {
    name: '仁英大廈 G 02',
    rows: [
      { location: '仁英大廈 G 02', user: '住客帳戶', type: '住客', permission: '只讀' },
    ],
  },
  { name: '仁英大廈 07 B', rows: [] },
  { name: '仁英大廈 08 C', rows: [] },
];

// 8. 物業綁定 mock 資料
const bindingSteps = [
  { num: '1', title: '選擇平台', desc: '先選擇 AJO PM 大廈平台或 AJO Rent 租務平台。' },
  { num: '2', title: '填寫聯絡資料', desc: '提供申請人姓名、電話及電郵，便於管理處核對。' },
  { num: '3', title: '上載證明文件', desc: '按身份提交業權、租約或授權文件。' },
  { num: '4', title: '等待審批', desc: '審批通過後，該物業會加入帳戶可見範圍。' },
];

const bindingPlatform = ref<'pm' | 'rent'>('pm');

const bindingDocs = [
  { title: '身份證明', desc: '身份證、護照或公司授權人身份文件。' },
  { title: '物業關係證明', desc: '業權文件、租約、住戶證明或授權書。' },
  { title: '最近賬單或收據', desc: '管理費賬單、水電煤賬單或管理處認可文件。' },
  { title: '補充文件', desc: '如管理處要求，可上載其他補充證明。' },
];

const bindingStatusList = [
  { title: '康睦庭園第二座 / 02 / D', meta: '業主身份 · 已於 2026年6月5日完成審批', chip: '已綁定', chipType: 'good' },
  { title: '仁英大廈 / 07 / B', meta: '租客身份 · 等待管理處核對文件', chip: '審批中', chipType: 'warn' },
];

// 9. 住宅 mock 資料
const homeListings = [
  { name: '康睦庭園第二座 / 02 / D', identity: 'staff', status: '已綁定', chipType: 'good', updatedAt: '2026年6月5日' },
  { name: 'Harbour Residence', identity: '申請人', status: '待確認', chipType: 'warn', updatedAt: '2026年5月22日' },
];

// 10. 收藏 mock 資料
const savedItems = [
  { name: '佐敦高級住宅', category: '樓盤', price: 'HK$36,000/月', savedAt: '今天 12:30' },
  { name: '纖柔牙刷 精巧頭 3支裝', category: '綜合優惠', price: 'HK$14.50', savedAt: '昨天 17:20' },
];

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

onMounted(async () => {
  try {
    if (!sessionStore.me) {
      await sessionStore.loadCurrentUser();
    }
    syncBindFromProfile();
    await loadBindBuildings();
    if (bindBuildingID.value) {
      await loadBindUnits(bindBuildingID.value);
      syncBindUnitFromProfile();
    }
  } finally {
    isSyncingBindProfile = false;
  }
});

watch(bindBuildingID, (nextValue, previousValue) => {
  if (nextValue === previousValue) {
    return;
  }
  if (!isSyncingBindProfile) {
    bindFloor.value = '';
    bindUnitID.value = '';
  }
  void loadBindUnits(nextValue);
});

watch(bindFloor, (nextValue, previousValue) => {
  if (nextValue !== previousValue && !isSyncingBindProfile) {
    bindUnitID.value = '';
  }
});

watch(
  () => route.path,
  (path) => {
    syncActivePanelFromRoute(path);
  },
  { immediate: true },
);
</script>

<template>
  <div
    class="page"
    id="page-profile"
  >
    <div class="work-shell">
      <!-- 1. 左側導覽 -->
      <aside class="work-sidebar">
        <h1>會員中心</h1>
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
            <span v-if="item.needsApi" class="work-nav-note">（需要接口）</span>
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
                  <div class="work-account-avatar">P</div>
                  <div>
                    <div class="work-account-name">patrick</div>
                    <div class="work-account-sub">個人資料</div>
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
                  {{ isSigningOut ? '退出中' : '退出登入' }}
                </button>
                <button type="button" class="work-account-btn primary">編輯資料</button>
                <button type="button" class="work-account-btn">更新 iSmart</button>
              </div>
            </div>
          </section>

          <section class="work-account-grid">
            <div class="work-account-card">
              <div class="work-card-title">帳戶資料</div>
              <div class="work-account-field"><span>顯示名稱</span><strong>{{ accountProfile.displayName }}</strong></div>
              <div class="work-account-field"><span>ISMART 帳戶</span><strong>{{ accountProfile.ismartAccount }}</strong></div>
              <div class="work-account-field"><span>電郵地址</span><strong>{{ accountProfile.email }}</strong></div>
              <div class="work-account-field"><span>電話號碼</span><strong>{{ accountProfile.phone }}</strong></div>
              <div class="work-account-field"><span>登入密碼</span><strong>{{ accountProfile.password }}</strong></div>
              <div class="work-account-field"><span>發布者身份</span><strong>{{ accountProfile.publisherRole }}</strong></div>
              <div class="work-account-field"><span>地區代碼</span><strong>{{ accountProfile.regionCode }}</strong></div>
              <div class="work-account-field"><span>員工管理屋苑</span><strong>{{ accountProfile.managedBuildings }}</strong></div>
            </div>
            <div class="work-account-card">
              <div class="work-card-title">身份狀態</div>
              <div class="work-account-field"><span>會員編號</span><strong>{{ identityStatus.memberNo }}</strong></div>
              <div class="work-account-field"><span>會員狀態</span><strong>{{ identityStatus.memberStatus }}</strong></div>
              <div class="work-account-field"><span>會員類型</span><strong>{{ identityStatus.memberType }}</strong></div>
              <div class="work-account-field"><span>主要角色</span><strong>{{ identityStatus.primaryRole }}</strong></div>
              <div class="work-account-field"><span>資料完整度</span><strong>{{ identityStatus.profileCompleteness }}</strong></div>
              <div class="work-account-field"><span>員工權限</span><strong>{{ identityStatus.staffPermission }}</strong></div>
              <div class="work-role-strip">
                <span class="work-role-badge">staff</span>
                <span class="work-role-tag">尚未分配權限</span>
              </div>
            </div>
          </section>

          <section class="work-card" style="margin-top:14px;">
            <div class="work-card-title">綁定單位</div>
            <div class="work-bind-wrap">
              <div class="work-bind-current">{{ bindCurrent }}</div>
              <div class="work-bind-grid">
                <div class="work-bind-field">
                  <label class="work-bind-label">大廈</label>
                  <select
                    v-model="bindBuildingID"
                    class="work-bind-select"
                    :disabled="bindLoading || bindBuildingOptions.length === 0"
                  >
                    <option value="">{{ bindLoading ? '載入中' : '請選擇大廈' }}</option>
                    <option
                      v-for="item in bindBuildingOptions"
                      :key="item.value"
                      :value="item.value"
                    >
                      {{ item.label }}
                    </option>
                  </select>
                </div>
                <div class="work-bind-field">
                  <label class="work-bind-label">樓層</label>
                  <select
                    v-model="bindFloor"
                    class="work-bind-select"
                    :disabled="!bindBuildingID || bindUnitsLoading || bindFloorOptions.length === 0"
                  >
                    <option value="">{{ bindUnitsLoading ? '載入中' : '請選擇樓層' }}</option>
                    <option
                      v-for="item in bindFloorOptions"
                      :key="item.value"
                      :value="item.value"
                    >
                      {{ item.label }}
                    </option>
                  </select>
                </div>
                <div class="work-bind-field">
                  <label class="work-bind-label">單位</label>
                  <select
                    v-model="bindUnitID"
                    class="work-bind-select"
                    :disabled="!bindFloor || bindUnitsLoading || bindUnitOptions.length === 0"
                  >
                    <option value="">{{ bindUnitsLoading ? '載入中' : '請選擇單位' }}</option>
                    <option
                      v-for="item in bindUnitOptions"
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
                  {{ bindSaving ? '儲存中' : '儲存單位' }}
                </button>
              </div>
            </div>
          </section>

          <section class="work-card" style="margin-top:14px;">
            <div class="work-card-title">iSmart 帳號資料</div>
            <div class="work-ismart-grid">
              <div class="work-ismart-card">
                <div class="work-card-sub">帳號資料</div>
                <div v-for="item in ismartAccountData" :key="item.label" class="work-row">
                  <div>
                    <strong>{{ item.label }}</strong>
                    <span>{{ item.value }}</span>
                  </div>
                </div>
              </div>
              <div class="work-ismart-card">
                <div class="work-card-sub">業戶資料</div>
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
            <div class="work-card-title">相關物業</div>
            <div class="work-table-wrap">
              <table class="work-table">
                <thead><tr><th>物業</th><th>狀態</th></tr></thead>
                <tbody>
                  <tr v-for="item in relatedProperties" :key="item.name">
                    <td>{{ item.name }}</td>
                    <td>{{ item.status }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </section>

          <section class="work-card" style="margin-top:14px;">
            <div class="work-card-title">提示設定</div>
            <div class="work-setting-box">
              <label class="work-setting-check">
                <input type="checkbox" checked>
                <span>接收大廈通告電郵提示</span>
              </label>
              <button type="button" class="work-action work-compact-action">提交</button>
            </div>
          </section>
        </div>

        <!-- 2.2 授權副戶 -->
        <div v-show="activePanel === 'profile-subaccounts'" class="work-panel on" data-work-panel="profile-subaccounts">
          <section class="work-hero">
            <div>
              <div class="work-kicker">Subaccounts</div>
              <h2 class="work-title">授權副戶</h2>
              <p class="work-desc">只讀顯示各物業的授權副戶資料，後續再接同步與管理操作。</p>
            </div>
          </section>
          <section class="work-subaccount-group">
            <div v-for="group in subaccountGroups" :key="group.name" class="work-subaccount-card">
              <div class="work-subaccount-head">
                <div class="work-subaccount-name">{{ group.name }}</div>
                <button type="button" class="work-action">新增授權</button>
              </div>
              <table class="work-table">
                <thead><tr><th>地點</th><th>用戶</th><th>類型</th><th>權限</th></tr></thead>
                <tbody>
                  <tr v-for="row in group.rows" :key="row.location">
                    <td>{{ row.location }}</td>
                    <td>{{ row.user }}</td>
                    <td>{{ row.type }}</td>
                    <td>{{ row.permission }}</td>
                  </tr>
                  <tr v-if="group.rows.length === 0">
                    <td colspan="4" class="work-subaccount-empty">目前未有授權副戶資料</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </section>
        </div>

        <!-- 2.3 物業綁定 -->
        <div v-show="activePanel === 'profile-property-binding'" class="work-panel on" data-work-panel="profile-property-binding">
          <section class="work-hero">
            <div>
              <div class="work-kicker">Property Binding</div>
              <h2 class="work-title">物業綁定</h2>
              <p class="work-desc">提交大廈與單位資料，並上載指定文件予管理處審批。</p>
            </div>
          </section>
          <section class="work-card">
            <div class="work-card-title">申請流程</div>
            <div class="binding-flow">
              <div v-for="step in bindingSteps" :key="step.num" class="binding-step">
                <div class="binding-step-num">{{ step.num }}</div>
                <div class="binding-step-title">{{ step.title }}</div>
                <div class="binding-step-desc">{{ step.desc }}</div>
              </div>
            </div>
          </section>
          <section class="work-card">
            <div class="work-card-title">選擇平台</div>
            <div class="binding-platform-grid">
              <button
                type="button"
                class="binding-platform-card"
                :class="{ on: bindingPlatform === 'pm' }"
                @click="bindingPlatform = 'pm'"
              >
                <strong>AJO PM 大廈平台</strong>
                <span>適用於已接入 iSmart 或由物業管理公司審批的大廈。</span>
              </button>
              <button
                type="button"
                class="binding-platform-card"
                :class="{ on: bindingPlatform === 'rent' }"
                @click="bindingPlatform = 'rent'"
              >
                <strong>AJO Rent 租務平台</strong>
                <span>適用於租務住宅或由 AJO 確認的租住申請。</span>
              </button>
            </div>
            <div class="binding-review-box">
              <strong>審批責任</strong>
              <span v-if="bindingPlatform === 'pm'">AJO PM 大廈平台由物業管理公司審批。</span>
              <span v-else>AJO Rent 租務平台由 AJO 根據租務資料確認。</span>
            </div>
          </section>
          <section class="work-card">
            <div class="work-card-title">綁定申請</div>
            <div v-show="bindingPlatform === 'pm'" class="binding-platform-panel" :class="{ on: bindingPlatform === 'pm' }">
              <div class="binding-subtitle">已支援大廈</div>
              <div class="acct-form-grid">
                <div class="acct-field full">
                  <label>大廈</label>
                  <select class="acct-select">
                    <option>請選擇已支援大廈</option>
                    <option>時安大廈</option>
                    <option>仁英大廈</option>
                    <option>康睦庭園第二座</option>
                    <option>協和大廈</option>
                  </select>
                </div>
              </div>
              <div class="binding-subtitle">未收錄大廈資料</div>
              <div class="acct-form-grid">
                <div class="acct-field"><label>大廈名稱</label><input class="acct-input" type="text" placeholder="請輸入大廈名稱"></div>
                <div class="acct-field"><label>大廈地址</label><input class="acct-input" type="text" placeholder="請輸入完整地址"></div>
                <div class="acct-field"><label>樓層</label><input class="acct-input" type="text" placeholder="例如 07"></div>
                <div class="acct-field"><label>單位</label><input class="acct-input" type="text" placeholder="例如 B"></div>
                <div class="acct-field"><label>申請身份</label><select class="acct-select"><option>業主</option><option>租客</option><option>住戶代表</option><option>公司授權人</option></select></div>
              </div>
            </div>
            <div v-show="bindingPlatform === 'rent'" class="binding-platform-panel" :class="{ on: bindingPlatform === 'rent' }">
              <div class="binding-subtitle">租務單位資料</div>
              <div class="acct-form-grid">
                <div class="acct-field"><label>租務大廈或項目</label><input class="acct-input" type="text" placeholder="請輸入大廈或項目名稱"></div>
                <div class="acct-field"><label>申請身份</label><select class="acct-select"><option>業主</option><option>租客</option><option>住戶代表</option><option>公司授權人</option></select></div>
                <div class="acct-field"><label>樓層</label><input class="acct-input" type="text" placeholder="例如 07"></div>
                <div class="acct-field"><label>單位</label><input class="acct-input" type="text" placeholder="例如 B"></div>
              </div>
            </div>
            <div class="acct-form-grid binding-common-grid">
              <div class="acct-field"><label>申請人姓名</label><input class="acct-input" type="text"></div>
              <div class="acct-field"><label>聯絡電話</label><input class="acct-input" type="text"></div>
              <div class="acct-field full"><label>聯絡電郵</label><input class="acct-input" type="email"></div>
              <div class="acct-field full"><label>備註</label><textarea class="acct-textarea" placeholder="可補充與審批人核對所需資料"></textarea></div>
            </div>
          </section>
          <section class="work-card">
            <div class="work-card-title">所需文件</div>
            <div class="binding-doc-grid">
              <div v-for="doc in bindingDocs" :key="doc.title" class="binding-doc-card">
                <div class="binding-doc-title">{{ doc.title }}</div>
                <div class="binding-doc-desc">{{ doc.desc }}</div>
                <input class="acct-file" type="file">
              </div>
            </div>
            <div class="work-bind-actions work-doc-actions">
              <button type="button" class="work-action">提交審批</button>
            </div>
          </section>
          <section class="work-card">
            <div class="work-card-title">申請狀態</div>
            <div class="binding-status-list">
              <div v-for="item in bindingStatusList" :key="item.title" class="binding-status-item">
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
              <div class="work-kicker">Residential</div>
              <h2 class="work-title">我的住宅</h2>
              <p class="work-desc">查看已綁定住宅與服務式住宅申請。</p>
            </div>
          </section>
          <section class="work-card">
            <div class="work-card-title">住宅列表</div>
            <table class="work-table">
              <thead><tr><th>住宅</th><th>身份</th><th>狀態</th><th>更新時間</th></tr></thead>
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
              <div class="work-kicker">Furniture</div>
              <h2 class="work-title">我的家具</h2>
              <p class="work-desc">管理二手家具發布、上架、續期與交易狀態。</p>
            </div>
            <button type="button" class="work-action" @click="openFurnitureCreate">新增家具</button>
          </section>
          <section class="work-card">
            <div class="work-card-title">家具列表</div>
            <div class="work-account-field">
              <span>資料來源</span>
              <strong>我的家具列表已接入真實二手帖子資料。</strong>
            </div>
            <div class="work-table-actions">
              <button type="button" class="work-mini-btn primary" @click="openFurnitureListings">查看家具列表</button>
              <button type="button" class="work-mini-btn" @click="openFurnitureCreate">新增家具</button>
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
              <div class="work-kicker">Saved</div>
              <h2 class="work-title">我的收藏</h2>
              <p class="work-desc">查看已收藏樓盤、家具與優惠商品。</p>
            </div>
          </section>
          <section class="work-card">
            <div class="work-card-title">收藏列表</div>
            <table class="work-table">
              <thead><tr><th>項目</th><th>分類</th><th>價格</th><th>收藏時間</th></tr></thead>
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
  min-height: calc(100vh - var(--nav-h, 52px));
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

/* 19. 平台選擇 */
.binding-platform-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.binding-platform-card {
  border: 1px solid var(--bdr);
  border-radius: 8px;
  background: #fff;
  color: var(--ink);
  cursor: pointer;
  font-family: inherit;
  text-align: left;
  padding: 16px;
  transition: border-color 0.15s, background 0.15s, box-shadow 0.15s;
}

.binding-platform-card:hover {
  border-color: var(--brand-mid);
  box-shadow: var(--shadow-sm);
}

.binding-platform-card.on {
  border-color: var(--brand);
  background: var(--brand-light);
}

.binding-platform-card strong {
  display: block;
  color: var(--ink);
  font-size: 14px;
  font-weight: 800;
  margin-bottom: 6px;
}

.binding-platform-card span {
  display: block;
  color: var(--ink-3);
  font-size: 12px;
  line-height: 1.6;
}

.binding-review-box {
  display: grid;
  gap: 6px;
  margin-top: 12px;
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

/* 21. 文件上傳 */
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

.acct-file {
  flex: 1;
  min-width: 260px;
  border: 1px dashed var(--bdr-2);
  border-radius: 6px;
  background: var(--sur-2);
  padding: 12px;
  color: var(--ink-3);
  font-size: 12px;
  font-weight: 700;
}

/* 24. 響應式 - 平板 */
@media (max-width: 900px) {
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
  .binding-platform-grid,
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

}
</style>

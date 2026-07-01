<!--
 * 我的大廈頁。
 * 1. 高保真還原 HTML 設計稿 page-affairs 雙欄布局（左側大廈導航 + 右側面板）。
 * 2. 提供最新通告、大廈資料、大廈財務、業戶帳目、申請表格、意見提供/維修報修、智能門禁、視像監控、設備監測九個面板。
 * 3. 意見提供面板包含最近記錄（可展開內容）、入口卡片、引導式四步提交流程。
 * 4. 大廈資料讀取目前會員 iSmart 綁定大廈。
-->
<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useRouter } from 'vue-router';

import QrCodeImage from '@/shared/components/base/QrCodeImage.vue';
import { useSessionStore } from '@/stores/session';
import {
  fetchMemberICCTVPublicCameras,
  fetchMemberIsmartBuildingAccess,
  fetchMemberIsmartBuildingInfo,
  generateMemberIsmartDoorQRCode,
  openMemberIsmartDoor,
  type ICCTVCameraSummary,
  type IsmartAccessDoor,
  type IsmartAccessRecord,
  type IsmartBuildingDocument,
  type IsmartBuildingAccessResponse,
  type IsmartBuildingInfoResponse,
  type IsmartRecentAccessGroup,
} from '@/httpapis/building';

// 1. 型別定義
type AffairsTab =
  | 'affairs-notices'
  | 'affairs-building'
  | 'affairs-finance'
  | 'affairs-owner-account'
  | 'affairs-forms'
  | 'affairs-feedback'
  | 'affairs-access'
  | 'affairs-icctv'
  | 'affairs-equipment';

type FinanceSubTab = 'acct-overview' | 'acct-finance' | 'acct-audit';
type AffairsMode = 'repair' | 'feedback';
type PayStatus = 'paid' | 'due';

interface NavItem {
  target: AffairsTab;
  label: string;
  needsApi?: boolean;
}

interface NoticeRow {
  code: string;
  title: string;
  type: string;
  publishDate: string;
  expireDate: string;
}

interface BuildingField {
  label: string;
  value: string;
}

interface BuildingDocCard {
  title: string;
  desc: string;
  count: string;
  empty: string;
}

interface FloorPlanRow {
  title: string;
  date: string;
  month: string;
  url: string;
}

interface BuildingFileRow {
  key: string;
  title: string;
  date: string;
  month: string;
  url: string;
}

interface FinanceOverRow {
  unit: string;
  m11: string;
  m10: string;
  m09: string;
  mBefore: string;
  m11Status: PayStatus;
  m10Status: PayStatus;
  m09Status: PayStatus;
  mBeforeStatus: PayStatus;
}

interface FinanceReportRow {
  date: string;
  title: string;
}

interface AuditRow {
  date: string;
  item: string;
}

interface FeedbackRecord {
  type: string;
  building: string;
  subject: string;
  category: string;
  status: 'warn' | 'good' | '';
  statusText: string;
  updatedAt: string;
  content: string;
}

interface EquipmentRow {
  device: string;
  location: string;
  status: 'good' | 'warn';
  statusText: string;
  updatedAt: string;
}

interface AccessQRPanel {
  doorID: string;
  doorTitle: string;
  value: string;
  expiresAt: string;
  term: string;
}

interface AccessRecordRow {
  key: string;
  doorTitle: string;
  openTime: string;
  openType: string;
  status: 'good' | 'warn';
  statusText: string;
}

// 2. 路由
const router = useRouter();
const sessionStore = useSessionStore();

// 3. 主面板與子面板狀態
const activeTab = ref<AffairsTab>('affairs-feedback');
const financeSubTab = ref<FinanceSubTab>('acct-overview');
const buildingInfoLoading = ref(false);
const buildingInfoError = ref('');
const selectedBuildingID = ref('');
const ismartBuildingProfile = ref<IsmartBuildingInfoResponse | null>(null);
const accessLoading = ref(false);
const accessError = ref('');
const accessMessage = ref('');
const ismartAccessProfile = ref<IsmartBuildingAccessResponse | null>(null);
const visiblePasswordDoorIDs = ref<string[]>([]);
const openingDoorID = ref('');
const qrLoadingDoorID = ref('');
const accessQRPanel = ref<AccessQRPanel | null>(null);
const icctvLoading = ref(false);
const icctvError = ref('');
const icctvProfile = ref<Awaited<ReturnType<typeof fetchMemberICCTVPublicCameras>> | null>(null);
const expandedICCTVCameraIDs = ref<string[]>([]);

// 4. 意見提供引導式表單狀態
const affairsMode = ref<AffairsMode>('repair');
const affairsStep = ref(1);
const expandedRecord = ref<number | null>(null);
const selectedFeedbackBuilding = ref('協和大廈');
const repairCategory = ref('');
const repairSubcategory = ref('');
const feedbackCategory = ref('');
const feedbackSubcategory = ref('');
const repairContent = ref('');
const feedbackContent = ref('');
const mediaFileName = ref('');

// 5. 左側導航項目
const navItems: NavItem[] = [
  { target: 'affairs-notices', label: '最新通告', needsApi: true },
  { target: 'affairs-building', label: '大廈資料' },
  { target: 'affairs-finance', label: '大廈財務', needsApi: true },
  { target: 'affairs-owner-account', label: '業戶帳目', needsApi: true },
  { target: 'affairs-forms', label: '申請表格' },
  { target: 'affairs-feedback', label: '意見提供/維修報修', needsApi: true },
  { target: 'affairs-access', label: '智能門禁' },
  { target: 'affairs-icctv', label: '視像監控' },
  { target: 'affairs-equipment', label: '設備監測', needsApi: true },
];

// 6. 最新通告 mock 資料
const noticeBuildings = ['仁美大廈', '康睦庭園第二座'];
const selectedNoticeBuilding = ref('仁美大廈');
const notices: NoticeRow[] = [
  {
    code: 'bulk/0145100_90c1340f10e849bd9c69',
    title: '清潔員大年初一休假一天',
    type: '一般',
    publishDate: 'Jan. 24, 2025',
    expireDate: 'Jan. 31, 2025',
  },
];

// 7. 大廈資料
const textValue = (value: string | number | null | undefined): string => {
  if (value === null || value === undefined) return '-';
  const text = String(value).trim();
  return text || '-';
};

const fileRows = (rows: IsmartBuildingDocument[] | undefined): BuildingFileRow[] =>
  (rows ?? []).map((item, index) => ({
    key: String(item.id ?? `${item.title ?? 'file'}-${index}`),
    title: textValue(item.title),
    date: textValue(item.file_date),
    month: textValue(item.file_month),
    url: String(item.file_url ?? '').trim(),
  }));

const normalizeMapEmbedURL = (value: string | null | undefined): string => {
  const raw = String(value ?? '').trim();
  if (!raw) return '';
  const iframeMatch = raw.match(/src=["']([^"']+)["']/i);
  const url = iframeMatch?.[1] ?? raw;
  if (!/^https?:\/\//i.test(url)) return '';
  if (url.includes('/maps/embed')) return url;
  if (url.includes('google.com/maps')) {
    const separator = url.includes('?') ? '&' : '?';
    return `${url}${separator}output=embed`;
  }
  return '';
};

const currentBuilding = computed(() => ismartBuildingProfile.value?.building ?? {});
const currentBuildingInfo = computed(() => ismartBuildingProfile.value?.building_info ?? {});
const buildingName = computed(() => textValue(currentBuilding.value.buildname_chi || currentBuilding.value.buildname || selectedBuildingID.value));
const organizationName = computed(() => textValue(currentBuildingInfo.value.owners_corporation_name));
const buildingForms = computed(() => fileRows(ismartBuildingProfile.value?.documents?.forms));
const buildingInfoFiles = computed(() => fileRows(ismartBuildingProfile.value?.documents?.building_info_files));
const floorPlans = computed<FloorPlanRow[]>(() => fileRows(ismartBuildingProfile.value?.documents?.floorplan));
const buildingMapURL = computed(() => String(currentBuildingInfo.value.google_map_url ?? '').trim());
const buildingMapEmbedURL = computed(() => normalizeMapEmbedURL(currentBuildingInfo.value.google_map_url));
const buildingFileCount = computed(() => buildingForms.value.length + buildingInfoFiles.value.length + floorPlans.value.length);
const buildingStatusText = computed(() => {
  if (buildingInfoLoading.value) return '載入中';
  if (buildingInfoError.value) return '未能載入';
  return ismartBuildingProfile.value ? '已有資料記錄' : '未有資料記錄';
});
const buildingStatusDetail = computed(() => {
  if (buildingInfoLoading.value) return '正在讀取已綁定大廈資料';
  if (buildingInfoError.value) return buildingInfoError.value;
  return ismartBuildingProfile.value ? '資料由 iSmart 同步顯示' : '目前未有可顯示的大廈資料';
});
const buildingStatusClass = computed(() => (buildingInfoError.value ? 'warn' : 'good'));

const buildingFields = computed<BuildingField[]>(() => [
  { label: '落成年份', value: textValue(currentBuildingInfo.value.year_built) },
  { label: '樓層總數', value: textValue(currentBuildingInfo.value.total_floor) },
  { label: '單位總數', value: textValue(currentBuildingInfo.value.total_unit) },
  { label: '車位總數', value: textValue(currentBuildingInfo.value.total_carpark) },
  { label: '法團名稱', value: textValue(currentBuildingInfo.value.owners_corporation_name) },
  { label: '管理處電話', value: textValue(currentBuildingInfo.value.management_office_phone) },
  { label: '管理公司名稱', value: textValue(currentBuildingInfo.value.management_company_name) },
  { label: '管理公司電話', value: textValue(currentBuildingInfo.value.management_company_phone) },
  { label: '管理公司電郵', value: textValue(currentBuildingInfo.value.management_company_email) },
  { label: '管理公司傳真', value: textValue(currentBuildingInfo.value.management_company_fax) },
  { label: '民政事務處電話', value: textValue(currentBuildingInfo.value.home_affairs_department_phone) },
  { label: '資料來源', value: ismartBuildingProfile.value ? 'iSmart 基本資料' : '-' },
]);

const buildingDocCards = computed<BuildingDocCard[]>(() => [
  {
    title: '表格',
    desc: '住戶常用或職員常用的基本表格文件。',
    count: `${buildingForms.value.length} 份`,
    empty: buildingForms.value.length > 0 ? '已有可下載表格。' : '目前未有表格。',
  },
  {
    title: '大廈資訊',
    desc: '對外發佈或內部參考的大廈介紹與基本資訊附件。',
    count: `${buildingInfoFiles.value.length} 份`,
    empty: buildingInfoFiles.value.length > 0 ? '已有大廈資訊文件。' : '目前未有大廈資訊。',
  },
  {
    title: '平面圖',
    desc: '平面圖、設施位置圖及相關圖則文件。',
    count: `${floorPlans.value.length} 份`,
    empty: floorPlans.value.length > 0 ? '已有平面圖資料。' : '目前未有平面圖。',
  },
]);

// 7.1 智能門禁資料
const accessBuildingName = computed(() => textValue(
  ismartAccessProfile.value?.building?.buildname_chi
  || ismartAccessProfile.value?.building?.buildname
  || buildingName.value,
));
const accessDoors = computed<IsmartAccessDoor[]>(() => ismartAccessProfile.value?.doors ?? []);
const accessRecordGroups = computed<IsmartRecentAccessGroup[]>(() => ismartAccessProfile.value?.recent_records ?? []);
const accessAllowedDoorCount = computed(() => accessDoors.value.filter((door) => door.has_permission).length);
const accessQRCodeDoorCount = computed(() => accessDoors.value.filter((door) => door.is_qrcode_enabled && door.qrcode?.record_id).length);
const accessStatusText = computed(() => {
  if (accessLoading.value) return '載入中';
  if (accessError.value) return '未能載入';
  return ismartAccessProfile.value ? '已連接' : '未載入';
});
const accessStatusClass = computed(() => (accessError.value ? 'warn' : 'good'));

// 7.2 取得門禁顯示值
const accessDoorID = (door: IsmartAccessDoor): string => String(door.door_id ?? '').trim();
const accessDoorNumber = (door: IsmartAccessDoor): number => Number(door.door_id ?? 0);
const accessQRCodeRecordNumber = (door: IsmartAccessDoor): number => Number(door.qrcode?.record_id ?? 0);
const accessDoorTitle = (door: IsmartAccessDoor): string => {
  const doorID = accessDoorID(door);
  return textValue(door.title || (doorID ? `門禁 ${doorID}` : '未命名門禁'));
};
const accessDoorPasswordVisible = (door: IsmartAccessDoor): boolean => visiblePasswordDoorIDs.value.includes(accessDoorID(door));
const accessDoorPasswordText = (door: IsmartAccessDoor): string => {
  if (!door.password?.value) return '-';
  return accessDoorPasswordVisible(door) ? door.password.value : '******';
};
const accessTimeRange = (start: string | undefined, end: string | undefined): string => {
  const startText = textValue(start);
  const endText = textValue(end);
  if (startText === '-' && endText === '-') return '-';
  return `${startText} 至 ${endText}`;
};
const accessOpenTypeText = (value: string | undefined): string => {
  const text = String(value ?? '').trim();
  if (text === 'remote') return '遠端開門';
  if (text === 'qrcode') return '二維碼';
  return text || '-';
};
const accessRecordSuccess = (record: IsmartAccessRecord): boolean => {
  const value = record.is_success;
  return value === true || value === 1 || value === '1' || String(value).toLowerCase() === 'true';
};
const accessRecentRows = computed<AccessRecordRow[]>(() =>
  accessRecordGroups.value.flatMap((group, groupIndex) =>
    (group.records ?? []).map((record, recordIndex) => {
      const success = accessRecordSuccess(record);
      return {
        key: `${group.door?.id ?? groupIndex}-${record.open_time ?? recordIndex}-${recordIndex}`,
        doorTitle: textValue(group.door?.title),
        openTime: textValue(record.open_time),
        openType: accessOpenTypeText(record.open_type),
        status: success ? 'good' : 'warn',
        statusText: success ? '成功' : '未成功',
      };
    }),
  ),
);

// 7.3 視像監控資料
const icctvCameras = computed<ICCTVCameraSummary[]>(() => icctvProfile.value?.cameras ?? []);
const icctvOrangePis = computed(() => icctvProfile.value?.orangepis ?? []);
const memberBoundCommunityName = computed(() => String(
  sessionStore.me?.primary_community?.name_zh
  || sessionStore.me?.primary_community?.name_en
  || sessionStore.me?.primary_community?.address_text
  || '',
).trim());
const icctvBuildingTitle = computed(() => {
  const parts = [
    memberBoundCommunityName.value,
    String(sessionStore.me?.residence_floor ?? '').trim(),
    String(sessionStore.me?.residence_unit ?? '').trim(),
  ].filter(Boolean);
  if (parts.length > 0) return parts.join(' / ');

  return buildingName.value;
});
const icctvEnabled = computed(() => icctvOrangePis.value.some((item) => item.is_active));
const icctvStatusText = computed(() => {
  if (icctvLoading.value) return '載入中';
  if (icctvError.value) return '未能載入';
  return icctvEnabled.value ? '已啟用' : '未啟用';
});
const icctvCameraName = (camera: ICCTVCameraSummary, index: number): string => {
  const match = String(camera.channel ?? '').match(/^channel(\d+)$/i);
  return `鏡頭 ${match?.[1] ?? index + 1}`;
};
const icctvCameraStatusText = (camera: ICCTVCameraSummary): string => (camera.is_active && camera.url ? '可查看' : '不可查看');
const icctvCameraFrameTitle = (camera: ICCTVCameraSummary, index: number): string => `${icctvCameraName(camera, index)} 即時監控`;
const isICCTVCameraExpanded = (cameraID: string): boolean => expandedICCTVCameraIDs.value.includes(cameraID);

// 8. 大廈財務 mock 資料（管理費總覽）
const financeOverview: FinanceOverRow[] = [
  { unit: 'G樓 01', m11: '-463.0', m10: '已付', m09: '已付', mBefore: '已付', m11Status: 'due', m10Status: 'paid', m09Status: 'paid', mBeforeStatus: 'paid' },
  { unit: 'G樓 02', m11: '-463.0', m10: '已付', m09: '已付', mBefore: '已付', m11Status: 'due', m10Status: 'paid', m09Status: 'paid', mBeforeStatus: 'paid' },
  { unit: 'G樓 03', m11: '-463.0', m10: '已付', m09: '已付', mBefore: '已付', m11Status: 'due', m10Status: 'paid', m09Status: 'paid', mBeforeStatus: 'paid' },
  { unit: 'G樓 04', m11: '-463.0', m10: '已付', m09: '已付', mBefore: '已付', m11Status: 'due', m10Status: 'paid', m09Status: 'paid', mBeforeStatus: 'paid' },
  { unit: 'G樓 5A', m11: '已付', m10: '已付', m09: '已付', mBefore: '已付', m11Status: 'paid', m10Status: 'paid', m09Status: 'paid', mBeforeStatus: 'paid' },
  { unit: 'G樓 5B', m11: '-543.0', m10: '已付', m09: '已付', mBefore: '已付', m11Status: 'due', m10Status: 'paid', m09Status: 'paid', mBeforeStatus: 'paid' },
  { unit: '01樓 A', m11: '-476.0', m10: '已付', m09: '已付', mBefore: '已付', m11Status: 'due', m10Status: 'paid', m09Status: 'paid', mBeforeStatus: 'paid' },
  { unit: '01樓 B', m11: '-476.0', m10: '已付', m09: '已付', mBefore: '已付', m11Status: 'due', m10Status: 'paid', m09Status: 'paid', mBeforeStatus: 'paid' },
  { unit: '01樓 E', m11: '-1427.0', m10: '已付', m09: '已付', mBefore: '已付', m11Status: 'due', m10Status: 'paid', m09Status: 'paid', mBeforeStatus: 'paid' },
  { unit: '02樓 B', m11: '-476.0', m10: '-476.0', m09: '已付', mBefore: '已付', m11Status: 'due', m10Status: 'due', m09Status: 'paid', mBeforeStatus: 'paid' },
  { unit: '02樓 E', m11: '-1427.0', m10: '-1427.0', m09: '-1427.0', mBefore: '-1427.0', m11Status: 'due', m10Status: 'due', m09Status: 'due', mBeforeStatus: 'due' },
  { unit: '03樓 A', m11: '-619.0', m10: '-619.0', m09: '已付', mBefore: '已付', m11Status: 'due', m10Status: 'due', m09Status: 'paid', mBeforeStatus: 'paid' },
  { unit: '04樓 A', m11: '-1162.0', m10: '-1162.0', m09: '已付', mBefore: '已付', m11Status: 'due', m10Status: 'due', m09Status: 'paid', mBeforeStatus: 'paid' },
  { unit: '05樓 D', m11: '-1006.0', m10: '-1006.0', m09: '已付', mBefore: '已付', m11Status: 'due', m10Status: 'due', m09Status: 'paid', mBeforeStatus: 'paid' },
  { unit: '06樓 A', m11: '-1006.0', m10: '-1006.0', m09: '-1006.0', mBefore: '已付', m11Status: 'due', m10Status: 'due', m09Status: 'due', mBeforeStatus: 'paid' },
  { unit: '07樓 D', m11: '-1006.0', m10: '-1006.0', m09: '-1006.0', mBefore: '-5030.0', m11Status: 'due', m10Status: 'due', m09Status: 'due', mBeforeStatus: 'due' },
  { unit: '09樓 C', m11: '-1006.0', m10: '-1006.0', m09: '-1006.0', mBefore: '-2012.0', m11Status: 'due', m10Status: 'due', m09Status: 'due', mBeforeStatus: 'due' },
  { unit: '10樓 C', m11: '-1006.0', m10: '-1006.0', m09: '-1006.0', mBefore: '-6038.0', m11Status: 'due', m10Status: 'due', m09Status: 'due', mBeforeStatus: 'due' },
  { unit: '11樓 D', m11: '-1006.0', m10: '-1006.0', m09: '-1006.0', mBefore: '-6036.0', m11Status: 'due', m10Status: 'due', m09Status: 'due', mBeforeStatus: 'due' },
  { unit: '13樓 D', m11: '-1006.0', m10: '-1006.0', m09: '-1006.0', mBefore: '-7042.0', m11Status: 'due', m10Status: 'due', m09Status: 'due', mBeforeStatus: 'due' },
  { unit: '14樓 C', m11: '-1006.0', m10: '-1006.0', m09: '-1006.0', mBefore: '-7042.0', m11Status: 'due', m10Status: 'due', m09Status: 'due', mBeforeStatus: 'due' },
  { unit: '15樓 B', m11: '-1006.0', m10: '-1006.0', m09: '-1006.0', mBefore: '-1006.0', m11Status: 'due', m10Status: 'due', m09Status: 'due', mBeforeStatus: 'due' },
  { unit: '18樓 A', m11: '-1006.0', m10: '-1006.0', m09: '已付', mBefore: '已付', m11Status: 'due', m10Status: 'due', m09Status: 'paid', mBeforeStatus: 'paid' },
  { unit: '20樓 B', m11: '-697.0', m10: '-697.0', m09: '-697.0', mBefore: '-3485.0', m11Status: 'due', m10Status: 'due', m09Status: 'due', mBeforeStatus: 'due' },
  { unit: '21樓 C', m11: '-619.0', m10: '-619.0', m09: '-619.0', mBefore: '-16713.0', m11Status: 'due', m10Status: 'due', m09Status: 'due', mBeforeStatus: 'due' },
  { unit: '22樓 B', m11: '-1006.0', m10: '已付', m09: '已付', mBefore: '已付', m11Status: 'due', m10Status: 'paid', m09Status: 'paid', mBeforeStatus: 'paid' },
];

const financeReports: FinanceReportRow[] = [
  { date: '2027-07', title: 'YIG 財務報告 2022-07' },
  { date: '2023-03', title: 'YIG 財務報告 2023-03' },
  { date: '2023-03', title: 'YIG 財務報告 2023-03' },
  { date: '2023-02', title: 'YIG 財務報告 2023-02' },
  { date: '2023-02', title: 'YIG 財務報告 2023-01' },
  { date: '2022-12', title: 'YIG 財務報告 2022-12' },
  { date: '2022-11', title: 'YIG 財務報告 2022-11' },
];

const auditReports: AuditRow[] = [
  { date: '2021-11', item: 'YIG 財務報告 2021-11' },
  { date: '2018', item: '仁英大廈核數報告 2018' },
  { date: '2017', item: '仁英大廈核數報告 2017' },
  { date: '2016', item: '仁英大廈核數報告 2016' },
  { date: '2015', item: '仁英大廈核數報告 2015' },
];

// 9. 申請表格 mock 資料
const selectedFormOrg = computed(() => organizationName.value);
const selectedFormBuilding = computed(() => buildingName.value);
const forms = computed(() => buildingForms.value);

// 10. 意見提供 mock 資料
const feedbackBuildings = [
  '仁英大廈', '協和大廈', '華興大廈(269號)', '華興大廈(271號)',
  '豐富大廈', '得運大廈', '天富大廈', '新萬利大廈', '萬高大廈B座',
  '時安大廈', '東昇樓', '華園', '榮森工業第二大廈', '康睦庭園第二座',
  '南昌苑', '東南大樓(77號)', '東南大樓(75號)', '測試1大廈',
  '利來大廈', '仁文大廈', '榮昇閣', '仁利大廈', '麗麗大廈',
  '玉桂園(1座)', '玉桂園(2座)', '玉桂園(3座)', '玉桂園(4座)',
  '玉桂園(5座)', '玉桂園(6座)', '玉桂園(7座)', '玉桂園(8座)',
  '玉桂園(9座)', '玉桂園(10座)', '玉桂園(11座)',
];

const repairCategoryMap: Record<string, string[]> = {
  '電力與燈光': ['走廊照明', '大堂照明', '電制故障', '其他燈光'],
  '結構與門窗': ['門鎖', '窗戶', '牆身', '天花板', '其他結構'],
  '環境與衛生': ['清潔', '積水', '蟲鼠', '其他衛生'],
  '升降機': ['升降機故障', '升降機清潔', '其他升降機'],
  '水務': ['食水', '沖廁水', '漏水', '其他水務'],
  '其他維修': ['其他'],
};

const feedbackCategoryMap: Record<string, string[]> = {
  '環境與衛生': ['清潔建議', '環境改善', '其他衛生'],
  '公共設施': ['設施建議', '設施損壞', '其他設施'],
  '管理服務': ['管理服務建議', '職員表現', '其他管理'],
  '系統與平台功能': ['平台功能', '系統問題', '其他系統'],
  '其他意見': ['其他'],
};

const repairCategories = Object.keys(repairCategoryMap);
const feedbackCategories = Object.keys(feedbackCategoryMap);
const repairSubcategories = computed(() => {
  if (!repairCategory.value) return [];
  return repairCategoryMap[repairCategory.value] || [];
});
const feedbackSubcategories = computed(() => {
  if (!feedbackCategory.value) return [];
  return feedbackCategoryMap[feedbackCategory.value] || [];
});

const feedbackRecords: FeedbackRecord[] = [
  {
    type: '維修報修',
    building: '協和大廈',
    subject: '公共走廊照明檢查',
    category: '電力與燈光',
    status: 'warn',
    statusText: '處理中',
    updatedAt: '今天 10:20',
    content: '12樓公共走廊近升降機位置照明不穩，晚間出現閃爍，請安排檢查燈泡及電制。',
  },
  {
    type: '維修報修',
    building: '協和大廈',
    subject: '地下門鎖檢查',
    category: '結構與門窗',
    status: '',
    statusText: '待跟進',
    updatedAt: '昨天 09:30',
    content: '地下大門門鎖開合不順，住戶進出時需要多次嘗試，請安排師傅檢查。',
  },
  {
    type: '意見反映',
    building: '協和大廈',
    subject: '大堂清潔建議',
    category: '環境與衛生',
    status: 'good',
    statusText: '已完成',
    updatedAt: '昨天 16:45',
    content: '大堂入口雨天較易積水，建議加密清潔及放置防滑提示牌，管理處已完成跟進。',
  },
  {
    type: '意見反映',
    building: '時安大廈',
    subject: '平台功能建議',
    category: '系統與平台功能',
    status: 'warn',
    statusText: '處理中',
    updatedAt: '2026年6月4日',
    content: '希望日後可在平台查看管理處回覆進度及補充相片，方便住戶追蹤事項。',
  },
];

// 11. 設備監測 mock 資料
const equipmentRows: EquipmentRow[] = [
  { device: '升降機 1 號', location: '大堂', status: 'good', statusText: '正常', updatedAt: '今天 10:20' },
  { device: '水泵房', location: '地庫', status: 'warn', statusText: '需檢查', updatedAt: '今天 09:40' },
  { device: '照明系統', location: '公共走廊', status: 'good', statusText: '正常', updatedAt: '昨天 18:10' },
];

// 12. 讀取目前會員綁定大廈資料
const loadBuildingInfo = async (buildingID = selectedBuildingID.value) => {
  buildingInfoLoading.value = true;
  buildingInfoError.value = '';
  try {
    const result = await fetchMemberIsmartBuildingInfo(buildingID || undefined);
    ismartBuildingProfile.value = result;
    selectedBuildingID.value = result.selected_building_id || result.building?.building_id || buildingID || '';
  } catch (error) {
    console.error(error);
    buildingInfoError.value = '大廈資料載入失敗';
  } finally {
    buildingInfoLoading.value = false;
  }
};

// 13. 讀取目前會員智能門禁資料
const loadBuildingAccess = async (clearMessage = true) => {
  accessLoading.value = true;
  accessError.value = '';
  if (clearMessage) {
    accessMessage.value = '';
  }
  try {
    const result = await fetchMemberIsmartBuildingAccess(selectedBuildingID.value || undefined);
    ismartAccessProfile.value = result;
    selectedBuildingID.value = result.selected_building_id
      || result.building?.access_building_id
      || result.building?.requested_building_id
      || selectedBuildingID.value
      || '';
  } catch (error) {
    console.error(error);
    accessError.value = '門禁資料載入失敗';
  } finally {
    accessLoading.value = false;
  }
};

// 14. 讀取目前會員視像監控資料
const loadICCTV = async () => {
  icctvLoading.value = true;
  icctvError.value = '';
  try {
    const result = await fetchMemberICCTVPublicCameras(selectedBuildingID.value || undefined);
    icctvProfile.value = result;
    selectedBuildingID.value = result.selected_building_id || selectedBuildingID.value || '';
    expandedICCTVCameraIDs.value = [];
  } catch (error) {
    console.error(error);
    icctvError.value = '視像監控資料載入失敗';
    icctvProfile.value = null;
    expandedICCTVCameraIDs.value = [];
  } finally {
    icctvLoading.value = false;
  }
};

// 15. 切換主面板
const switchTab = (target: AffairsTab) => {
  activeTab.value = target;
  if (target === 'affairs-access' && !ismartAccessProfile.value && !accessLoading.value) {
    void loadBuildingAccess();
  }
  if (target === 'affairs-icctv' && !icctvProfile.value && !icctvLoading.value) {
    void loadICCTV();
  }
};

// 16. 取得門禁操作用大廈 ID
const accessPayloadBuildingID = (door?: IsmartAccessDoor): string | undefined => {
  const buildingID = selectedBuildingID.value
    || ismartAccessProfile.value?.selected_building_id
    || door?.building_id
    || '';
  return buildingID || undefined;
};

// 17. 切換門禁密碼可見狀態
const toggleAccessPassword = (door: IsmartAccessDoor) => {
  const doorID = accessDoorID(door);
  if (!doorID) return;
  visiblePasswordDoorIDs.value = visiblePasswordDoorIDs.value.includes(doorID)
    ? visiblePasswordDoorIDs.value.filter((id) => id !== doorID)
    : [...visiblePasswordDoorIDs.value, doorID];
};

// 18. 發送開門指令
const openAccessDoor = async (door: IsmartAccessDoor) => {
  const doorID = accessDoorID(door);
  const doorNumber = accessDoorNumber(door);
  if (!door.has_permission || doorNumber <= 0) return;
  if (!window.confirm(`確認開啟「${accessDoorTitle(door)}」？`)) return;

  openingDoorID.value = doorID;
  accessMessage.value = '';
  accessError.value = '';
  try {
    const result = await openMemberIsmartDoor({
      building_id: accessPayloadBuildingID(door),
      door_id: doorNumber,
    });
    const resultMessage = result.is_success === false ? '開門指令未成功' : '已發送開門指令';
    await loadBuildingAccess(false);
    accessMessage.value = resultMessage;
  } catch (error) {
    console.error(error);
    accessError.value = '開門指令發送失敗';
  } finally {
    openingDoorID.value = '';
  }
};

// 19. 生成門禁二維碼
const generateAccessQRCode = async (door: IsmartAccessDoor) => {
  const doorID = accessDoorID(door);
  const recordID = accessQRCodeRecordNumber(door);
  if (!door.is_qrcode_enabled || recordID <= 0) return;

  qrLoadingDoorID.value = doorID;
  accessMessage.value = '';
  accessError.value = '';
  try {
    const result = await generateMemberIsmartDoorQRCode({
      building_id: accessPayloadBuildingID(door),
      qrcode_record_id: recordID,
      term: 'dynamic',
    });
    if (!result.qrcode_value) {
      accessError.value = '二維碼內容為空';
      accessQRPanel.value = null;
      return;
    }
    accessQRPanel.value = {
      doorID,
      doorTitle: accessDoorTitle(door),
      value: result.qrcode_value,
      expiresAt: textValue(result.expires_at),
      term: textValue(result.term || 'dynamic'),
    };
  } catch (error) {
    console.error(error);
    accessError.value = '二維碼生成失敗';
  } finally {
    qrLoadingDoorID.value = '';
  }
};

// 20. 切換財務子面板
const switchFinanceSub = (target: FinanceSubTab) => {
  financeSubTab.value = target;
};

// 21. 切換意見提供模式
const setAffairsMode = (mode: AffairsMode) => {
  affairsMode.value = mode;
  affairsStep.value = 1;
};

// 22. 切換意見提供步驟
const setAffairsStep = (step: number) => {
  if (step < 1) return;
  if (step > 4) return;
  affairsStep.value = step;
};

// 23. 下一步
const nextAffairsStep = () => {
  if (affairsStep.value < 4) {
    affairsStep.value += 1;
  }
};

// 24. 展開/收起最近記錄
const toggleRecord = (index: number) => {
  expandedRecord.value = expandedRecord.value === index ? null : index;
};

// 25. 選擇檔案
const onMediaChange = (e: Event) => {
  const target = e.target as HTMLInputElement;
  if (!target.files || target.files.length === 0) {
    mediaFileName.value = '';
    return;
  }
  mediaFileName.value = Array.from(target.files)
    .map((f) => f.name)
    .join(', ');
};

// 26. 提交意見提供
const submitAffairsFeedback = () => {
  affairsMode.value = 'repair';
  affairsStep.value = 1;
  repairCategory.value = '';
  repairSubcategory.value = '';
  feedbackCategory.value = '';
  feedbackSubcategory.value = '';
  repairContent.value = '';
  feedbackContent.value = '';
  mediaFileName.value = '';
};

// 27. 確認提交摘要
const reviewMode = computed(() => (affairsMode.value === 'repair' ? '維修報修' : '意見反映'));
const reviewCategory = computed(() =>
  affairsMode.value === 'repair'
    ? repairCategory.value || '請選擇大類'
    : feedbackCategory.value || '請選擇大類',
);
const reviewSubcategory = computed(() =>
  affairsMode.value === 'repair'
    ? repairSubcategory.value || '請選擇次分類'
    : feedbackSubcategory.value || '請選擇次分類',
);
const reviewContent = computed(() =>
  affairsMode.value === 'repair'
    ? repairContent.value || '請填寫內容'
    : feedbackContent.value || '請填寫內容',
);
const reviewMedia = computed(() => mediaFileName.value || '未選擇檔案');

// 28. 重新整理通告
const refreshNotices = () => {
  // 靜態 mock，無需操作
};

// 29. 展開或收起視像監控鏡頭
const toggleICCTVCamera = (cameraID: string) => {
  expandedICCTVCameraIDs.value = isICCTVCameraExpanded(cameraID)
    ? expandedICCTVCameraIDs.value.filter((id) => id !== cameraID)
    : [...expandedICCTVCameraIDs.value, cameraID];
};

// 30. 開啟視像監控新窗口
const openIcctvWindow = (camera: ICCTVCameraSummary) => {
  if (!camera.url) return;
  window.open(camera.url, '_blank', 'noopener');
};

// 31. 重新整理設備監測
const refreshEquipment = () => {
  // 靜態 mock，無需操作
};

// 32. 跳轉至通告詳情
const goNoticeDetail = () => {
  router.push('/building/notices');
};

onMounted(() => {
  void loadBuildingInfo();
});
</script>

<template>
  <div
    id="page-affairs"
    class="page"
  >
    <div class="work-shell">
      <!-- 左側大廈導航 -->
      <aside class="work-sidebar">
        <h1>我的大廈</h1>
        <nav class="work-nav">
          <button
            v-for="item in navItems"
            :key="item.target"
            type="button"
            class="work-nav-item"
            :class="{ on: activeTab === item.target }"
            @click="switchTab(item.target)"
          >
            <span class="work-nav-label">{{ item.label }}</span>
            <span v-if="item.needsApi" class="work-nav-note">（需要接口）</span>
          </button>
        </nav>
      </aside>

      <!-- 右側主內容 -->
      <main class="work-main">
        <!-- 最新通告 -->
        <div
          v-show="activeTab === 'affairs-notices'"
          class="work-panel"
          :class="{ on: activeTab === 'affairs-notices' }"
          data-work-panel="affairs-notices"
        >
          <section class="work-hero">
            <div>
              <div class="work-kicker">Building Notices</div>
              <h2 class="work-title">最新通告</h2>
            </div>
          </section>
          <section class="notice-admin-grid">
            <div class="notice-admin-card">
              <h3>大廈選擇</h3>
              <label class="notice-admin-label">選擇大廈:</label>
              <select
                v-model="selectedNoticeBuilding"
                class="notice-select"
              >
                <option
                  v-for="b in noticeBuildings"
                  :key="b"
                  :value="b"
                >
                  {{ b }}
                </option>
              </select>
            </div>
          </section>
          <section class="work-card">
            <div class="notice-current">
              <span>目前顯示: 所有通告</span>
              <button
                type="button"
                class="notice-action-btn"
                @click="refreshNotices"
              >
                重新整理
              </button>
            </div>
            <table class="work-table notice-table">
              <thead>
                <tr>
                  <th>編號</th>
                  <th>標題</th>
                  <th>類型</th>
                  <th>發佈日期</th>
                  <th>下架日期</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="n in notices"
                  :key="n.code"
                >
                  <td>{{ n.code }}</td>
                  <td>{{ n.title }}</td>
                  <td>{{ n.type }}</td>
                  <td>{{ n.publishDate }}</td>
                  <td>{{ n.expireDate }}</td>
                  <td>
                    <div class="notice-table-actions">
                      <button
                        type="button"
                        class="notice-action-btn"
                        @click="goNoticeDetail"
                      >
                        查閱
                      </button>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </section>
        </div>

        <!-- 大廈資料 -->
        <div
          v-show="activeTab === 'affairs-building'"
          class="work-panel"
          :class="{ on: activeTab === 'affairs-building' }"
          data-work-panel="affairs-building"
        >
          <section class="work-hero">
            <div>
              <div class="work-kicker">Building Profile</div>
              <h2 class="work-title">大廈基本資料</h2>
              <p class="work-desc">查看已綁定大廈的基本資料、常用表格、大廈資訊及平面圖。</p>
            </div>
            <button
              type="button"
              class="work-action"
              :disabled="buildingInfoLoading"
              @click="loadBuildingInfo()"
            >
              重新整理
            </button>
          </section>
          <section class="building-summary-grid">
            <div class="work-card">
              <div class="work-card-title">機構</div>
              <div class="work-card-sub">{{ organizationName }}</div>
            </div>
            <div class="work-card">
              <div class="work-card-title">目前大廈</div>
              <div class="work-stat-label">{{ buildingName }}</div>
              <div class="work-stat">{{ buildingFileCount }}</div>
              <div class="work-stat-label">基本資料文件</div>
            </div>
            <div class="work-card">
              <div class="work-card-title">資料狀態</div>
              <div class="work-row">
                <div>
                  <strong>{{ buildingStatusText }}</strong>
                  <span>{{ buildingStatusDetail }}</span>
                </div>
                <span
                  class="work-chip"
                  :class="buildingStatusClass"
                >
                  {{ buildingInfoLoading ? '載入中' : buildingInfoError ? '異常' : '正常' }}
                </span>
              </div>
            </div>
            <div class="work-card">
              <div class="work-card-title">大廈</div>
              <div class="work-card-sub">{{ buildingName }}</div>
            </div>
          </section>
          <section class="work-card">
            <div class="work-card-title">基本欄位</div>
            <div class="building-field-grid">
              <div
                v-for="f in buildingFields"
                :key="f.label"
                class="building-field"
              >
                <span>{{ f.label }}</span>
                <strong>{{ f.value }}</strong>
              </div>
            </div>
          </section>
          <section class="work-card">
            <div class="work-card-title">地圖網址</div>
            <iframe
              v-if="buildingMapEmbedURL"
              class="building-map-frame"
              :src="buildingMapEmbedURL"
              allowfullscreen
              loading="lazy"
              referrerpolicy="no-referrer-when-downgrade"
            />
            <div
              v-else-if="buildingMapURL"
              class="building-map-fallback"
            >
              <a
                class="building-link"
                :href="buildingMapURL"
                target="_blank"
                rel="noopener"
              >
                查看地圖
              </a>
            </div>
            <div
              v-else
              class="building-empty-row"
            >
              未提供地圖網址。
            </div>
          </section>
          <section class="work-card">
            <div class="work-card-title">基本文件區</div>
            <div class="building-doc-grid">
              <div
                v-for="d in buildingDocCards"
                :key="d.title"
                class="building-doc-card"
              >
                <div class="work-card-title">{{ d.title }}</div>
                <div class="work-card-sub">{{ d.desc }}</div>
                <div class="building-doc-count">{{ d.count }}</div>
                <div class="building-doc-empty">{{ d.empty }}</div>
              </div>
            </div>
          </section>
          <section class="work-card">
            <div class="building-section-head">
              <div class="work-card-title">表格</div>
            </div>
            <table class="work-table building-file-table">
              <colgroup>
                <col class="building-file-title-col">
                <col class="building-file-date-col">
                <col class="building-file-month-col">
                <col class="building-file-action-col">
              </colgroup>
              <thead>
                <tr>
                  <th>標題</th>
                  <th>日期</th>
                  <th>月份</th>
                  <th>下載</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="file in buildingForms"
                  :key="file.key"
                >
                  <td class="building-file-title-cell">{{ file.title }}</td>
                  <td class="building-file-meta-cell">{{ file.date }}</td>
                  <td class="building-file-meta-cell">{{ file.month }}</td>
                  <td class="building-file-action-cell">
                    <a
                      v-if="file.url"
                      class="building-link"
                      :href="file.url"
                      target="_blank"
                      rel="noopener"
                    >查看</a>
                    <span v-else>-</span>
                  </td>
                </tr>
                <tr v-if="buildingForms.length === 0">
                  <td
                    class="building-empty-row"
                    colspan="4"
                  >
                    目前未有表格。
                  </td>
                </tr>
              </tbody>
            </table>
          </section>
          <section class="work-card">
            <div class="building-section-head">
              <div class="work-card-title">大廈資訊</div>
            </div>
            <table class="work-table building-file-table">
              <colgroup>
                <col class="building-file-title-col">
                <col class="building-file-date-col">
                <col class="building-file-month-col">
                <col class="building-file-action-col">
              </colgroup>
              <thead>
                <tr>
                  <th>標題</th>
                  <th>日期</th>
                  <th>月份</th>
                  <th>下載</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="file in buildingInfoFiles"
                  :key="file.key"
                >
                  <td class="building-file-title-cell">{{ file.title }}</td>
                  <td class="building-file-meta-cell">{{ file.date }}</td>
                  <td class="building-file-meta-cell">{{ file.month }}</td>
                  <td class="building-file-action-cell">
                    <a
                      v-if="file.url"
                      class="building-link"
                      :href="file.url"
                      target="_blank"
                      rel="noopener"
                    >查看</a>
                    <span v-else>-</span>
                  </td>
                </tr>
                <tr v-if="buildingInfoFiles.length === 0">
                  <td
                    class="building-empty-row"
                    colspan="4"
                  >
                    目前未有大廈資訊。
                  </td>
                </tr>
              </tbody>
            </table>
          </section>
          <section class="work-card">
            <div class="building-section-head">
              <div class="work-card-title">平面圖</div>
            </div>
            <table class="work-table building-file-table">
              <colgroup>
                <col class="building-file-title-col">
                <col class="building-file-date-col">
                <col class="building-file-month-col">
                <col class="building-file-action-col">
              </colgroup>
              <thead>
                <tr>
                  <th>標題</th>
                  <th>日期</th>
                  <th>月份</th>
                  <th>下載</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="p in floorPlans"
                  :key="p.title"
                >
                  <td class="building-file-title-cell">{{ p.title }}</td>
                  <td class="building-file-meta-cell">{{ p.date }}</td>
                  <td class="building-file-meta-cell">{{ p.month }}</td>
                  <td class="building-file-action-cell">
                    <a
                      v-if="p.url"
                      class="building-link"
                      :href="p.url"
                      target="_blank"
                      rel="noopener"
                    >查看</a>
                    <span v-else>-</span>
                  </td>
                </tr>
                <tr v-if="floorPlans.length === 0">
                  <td
                    class="building-empty-row"
                    colspan="4"
                  >
                    目前未有平面圖。
                  </td>
                </tr>
              </tbody>
            </table>
          </section>
        </div>

        <!-- 大廈財務 -->
        <div
          v-show="activeTab === 'affairs-finance'"
          class="work-panel"
          :class="{ on: activeTab === 'affairs-finance' }"
          data-work-panel="affairs-finance"
        >
          <section class="work-hero">
            <div>
              <div class="work-kicker">Building Finance</div>
              <h2 class="work-title">大廈財務</h2>
              <p class="work-desc">按大廈查看管理費、財務報表及核數報告。</p>
            </div>
          </section>
          <section class="work-card acct-card-wrap">
            <div class="acct-tabs">
              <button
                type="button"
                class="acct-tab"
                :class="{ on: financeSubTab === 'acct-overview' }"
                @click="switchFinanceSub('acct-overview')"
              >
                管理費總覽
              </button>
              <button
                type="button"
                class="acct-tab"
                :class="{ on: financeSubTab === 'acct-finance' }"
                @click="switchFinanceSub('acct-finance')"
              >
                財務報表
              </button>
              <button
                type="button"
                class="acct-tab"
                :class="{ on: financeSubTab === 'acct-audit' }"
                @click="switchFinanceSub('acct-audit')"
              >
                核數報告
              </button>
            </div>
            <div class="acct-body">
              <!-- 管理費總覽 -->
              <div
                v-show="financeSubTab === 'acct-overview'"
                class="acct-subpanel"
                :class="{ on: financeSubTab === 'acct-overview' }"
              >
                <div class="acct-note">選用「電子付款」繳交管理費住戶，於3小時內即可結算，其他支付方式由於核對需時未能即時更新，敬請見諒。</div>
                <div class="acct-toolbar">
                  <div class="work-card-title acct-toolbar-title">管理費總覽</div>
                  <input
                    class="acct-search"
                    placeholder="Search.."
                  >
                </div>
                <div class="acct-table-wrap">
                  <table class="acct-table">
                    <thead>
                      <tr>
                        <th>單位</th>
                        <th>2025/11</th>
                        <th>2025/10</th>
                        <th>2025/09</th>
                        <th>2025/08月前</th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr
                        v-for="r in financeOverview"
                        :key="r.unit"
                      >
                        <td>{{ r.unit }}</td>
                        <td :class="r.m11Status === 'due' ? 'acct-due' : 'acct-paid'">
                          {{ r.m11 }}
                        </td>
                        <td :class="r.m10Status === 'due' ? 'acct-due' : 'acct-paid'">
                          {{ r.m10 }}
                        </td>
                        <td :class="r.m09Status === 'due' ? 'acct-due' : 'acct-paid'">
                          {{ r.m09 }}
                        </td>
                        <td :class="r.mBeforeStatus === 'due' ? 'acct-due' : 'acct-paid'">
                          {{ r.mBefore }}
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>
              <!-- 財務報表 -->
              <div
                v-show="financeSubTab === 'acct-finance'"
                class="acct-subpanel"
                :class="{ on: financeSubTab === 'acct-finance' }"
              >
                <div class="acct-toolbar">
                  <div class="work-card-title acct-toolbar-title">法團財務報表</div>
                  <input
                    class="acct-search"
                    placeholder="Search.."
                  >
                </div>
                <div class="acct-table-wrap">
                  <table class="acct-table">
                    <thead>
                      <tr>
                        <th>日期</th>
                        <th>標題</th>
                        <th>查閱</th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr
                        v-for="(r, i) in financeReports"
                        :key="i"
                      >
                        <td>{{ r.date }}</td>
                        <td>{{ r.title }}</td>
                        <td>
                          <button
                            type="button"
                            class="acct-download"
                          >
                            查閱
                          </button>
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>
              <!-- 核數報告 -->
              <div
                v-show="financeSubTab === 'acct-audit'"
                class="acct-subpanel"
                :class="{ on: financeSubTab === 'acct-audit' }"
              >
                <div class="acct-toolbar">
                  <div class="work-card-title acct-toolbar-title">核數報告</div>
                  <input
                    class="acct-search"
                    placeholder="Search.."
                  >
                </div>
                <div class="acct-table-wrap">
                  <table class="acct-table">
                    <thead>
                      <tr>
                        <th>日期</th>
                        <th>項目</th>
                        <th>查閱</th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr
                        v-for="(r, i) in auditReports"
                        :key="i"
                      >
                        <td>{{ r.date }}</td>
                        <td>{{ r.item }}</td>
                        <td>
                          <button
                            type="button"
                            class="acct-download"
                          >
                            查閱
                          </button>
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>
            </div>
          </section>
        </div>

        <!-- 業戶帳目 -->
        <div
          v-show="activeTab === 'affairs-owner-account'"
          class="work-panel"
          :class="{ on: activeTab === 'affairs-owner-account' }"
          data-work-panel="affairs-owner-account"
        >
          <section class="work-hero">
            <div>
              <div class="work-kicker">Owner Account</div>
              <h2 class="work-title">業戶帳目</h2>
              <p class="work-desc">查看業戶管理費繳款記錄。</p>
            </div>
          </section>
          <section class="work-card acct-card-wrap">
            <div class="acct-tabs">
              <button
                type="button"
                class="acct-tab on"
              >
                管理費繳款記錄
              </button>
            </div>
            <div class="acct-body">
              <div class="acct-subpanel on">
                <div class="acct-dev">正在開發</div>
              </div>
            </div>
          </section>
        </div>

        <!-- 申請表格 -->
        <div
          v-show="activeTab === 'affairs-forms'"
          class="work-panel"
          :class="{ on: activeTab === 'affairs-forms' }"
          data-work-panel="affairs-forms"
        >
          <section class="work-hero">
            <div>
              <div class="work-kicker">Forms</div>
              <h2 class="work-title">申請表格</h2>
              <p class="work-desc">選擇大廈後查看可下載的住戶及工程申請表格。</p>
            </div>
          </section>
          <section class="notice-admin-grid">
            <div class="notice-admin-card">
              <h3>機構</h3>
              <div class="notice-admin-label">目前機構</div>
              <div class="notice-select building-readonly-select">{{ selectedFormOrg }}</div>
            </div>
            <div class="notice-admin-card">
              <h3>大廈</h3>
              <div class="notice-admin-label">目前大廈</div>
              <div class="notice-select building-readonly-select">{{ selectedFormBuilding }}</div>
            </div>
          </section>
          <section class="work-card">
            <table class="work-table">
              <thead>
                <tr>
                  <th>標題</th>
                  <th>查閱</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="f in forms"
                  :key="f.key"
                >
                  <td>{{ f.title }}</td>
                  <td>
                    <a
                      v-if="f.url"
                      class="building-link"
                      :href="f.url"
                      target="_blank"
                      rel="noopener"
                    >填寫表格</a>
                    <span v-else>-</span>
                  </td>
                </tr>
                <tr v-if="forms.length === 0">
                  <td
                    class="building-empty-row"
                    colspan="2"
                  >
                    目前未有申請表格。
                  </td>
                </tr>
              </tbody>
            </table>
          </section>
        </div>

        <!-- 意見提供/維修報修 -->
        <div
          v-show="activeTab === 'affairs-feedback'"
          class="work-panel"
          :class="{ on: activeTab === 'affairs-feedback' }"
          data-work-panel="affairs-feedback"
        >
          <section class="work-hero">
            <div>
              <div class="work-kicker">Feedback & Maintenance</div>
              <h2 class="work-title">意見提供/維修報修</h2>
              <p class="work-desc">選擇大廈及事項分類，提交後由管理處跟進。</p>
            </div>
          </section>
          <section class="work-card">
            <div class="work-card-title">最近記錄</div>
            <table class="work-table">
              <thead>
                <tr>
                  <th>類型</th>
                  <th>大廈</th>
                  <th>事項</th>
                  <th>分類</th>
                  <th>狀態</th>
                  <th>更新時間</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
              <template
                v-for="(r, i) in feedbackRecords"
                :key="i"
              >
                <tr>
                  <td>{{ r.type }}</td>
                  <td>{{ r.building }}</td>
                  <td>{{ r.subject }}</td>
                  <td>{{ r.category }}</td>
                  <td>
                    <span
                      class="work-chip"
                      :class="r.status"
                    >{{ r.statusText }}</span>
                  </td>
                  <td>{{ r.updatedAt }}</td>
                  <td>
                    <button
                      type="button"
                      class="work-mini-btn"
                      @click="toggleRecord(i)"
                    >
                      查看內容
                    </button>
                  </td>
                </tr>
                <tr
                  v-show="expandedRecord === i"
                  class="affairs-record-detail"
                >
                  <td colspan="7">
                    <div class="affairs-record-card">
                      <strong>內容</strong>
                      {{ r.content }}
                    </div>
                  </td>
                </tr>
              </template>
              </tbody>
            </table>
          </section>
          <section
            class="affairs-entry-grid"
            aria-label="意見提供與維修報修入口"
          >
            <button
              type="button"
              class="affairs-entry-card"
              :class="{ on: affairsMode === 'repair' }"
              @click="setAffairsMode('repair')"
            >
              <span
                class="affairs-entry-icon"
                aria-hidden="true"
              >
                <svg
                  viewBox="0 0 24 24"
                  fill="none"
                >
                  <path
                    d="m14.7 6.3 3-3a4 4 0 0 1 2.7 5.1l-4.7 4.7-4.8 4.8-3.3 3.3a2 2 0 0 1-2.8-2.8l3.3-3.3 4.8-4.8 4.7-4.7"
                    stroke-width="1.8"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                  />
                  <path
                    d="m8.1 15.1 2.8 2.8"
                    stroke-width="1.8"
                    stroke-linecap="round"
                  />
                </svg>
              </span>
              <span>
                <strong>提交維修報修</strong>
                <span>提交水務、電力、門窗、升降機及設施維修事項。</span>
              </span>
            </button>
            <button
              type="button"
              class="affairs-entry-card"
              :class="{ on: affairsMode === 'feedback' }"
              @click="setAffairsMode('feedback')"
            >
              <span
                class="affairs-entry-icon"
                aria-hidden="true"
              >
                <svg
                  viewBox="0 0 24 24"
                  fill="none"
                >
                  <path
                    d="M21 11.5a8.4 8.4 0 0 1-9 8.4 8.8 8.8 0 0 1-3.8-.9L3 20l1.1-4.6a8.4 8.4 0 1 1 16.9-3.9Z"
                    stroke-width="1.8"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                  />
                  <path
                    d="M8 10h8M8 13h5"
                    stroke-width="1.8"
                    stroke-linecap="round"
                  />
                </svg>
              </span>
              <span>
                <strong>意見反映</strong>
                <span>提交環境、管理服務、公共設施及平台功能意見。</span>
              </span>
            </button>
          </section>
          <section class="work-card affairs-feedback-form">
            <div class="work-card-title">選擇大廈</div>
            <div class="affairs-field">
              <label for="affairs-building">大廈</label>
              <select
                id="affairs-building"
                v-model="selectedFeedbackBuilding"
                class="affairs-select"
              >
                <option value="">---------</option>
                <option
                  v-for="b in feedbackBuildings"
                  :key="b"
                  :value="b"
                >
                  {{ b }}
                </option>
              </select>
            </div>
          </section>
          <section class="work-card affairs-guided-shell">
            <div
              class="affairs-stepper"
              aria-label="意見及維修提交流程"
            >
              <span
                class="affairs-step-indicator"
                :class="{ on: affairsStep === 1 }"
              >
                <b>1</b>選擇分類
              </span>
              <span
                class="affairs-step-indicator"
                :class="{ on: affairsStep === 2 }"
              >
                <b>2</b>填寫內容
              </span>
              <span
                class="affairs-step-indicator"
                :class="{ on: affairsStep === 3 }"
              >
                <b>3</b>上載圖片/影片
              </span>
              <span
                class="affairs-step-indicator"
                :class="{ on: affairsStep === 4 }"
              >
                <b>4</b>確認提交
              </span>
            </div>

            <!-- 步驟 1：選擇分類 -->
            <div
              v-show="affairsStep === 1"
              class="affairs-guided-step"
            >
              <div class="work-card-title">選擇分類</div>
              <div
                v-if="affairsMode === 'repair'"
                class="affairs-step-hint"
              >
                請選擇維修大類及次分類，方便管理處安排合適人員跟進。
              </div>
              <div
                v-else
                class="affairs-step-hint"
              >
                請選擇意見大類及次分類，方便管理處按事項性質處理。
              </div>
              <div
                v-if="affairsMode === 'repair'"
                class="affairs-field-grid"
              >
                <div class="affairs-field">
                  <label for="affairs-repair-category">維修大類</label>
                  <select
                    id="affairs-repair-category"
                    v-model="repairCategory"
                    class="affairs-select"
                  >
                    <option value="">---------</option>
                    <option
                      v-for="c in repairCategories"
                      :key="c"
                      :value="c"
                    >
                      {{ c }}
                    </option>
                  </select>
                </div>
                <div class="affairs-field">
                  <label for="affairs-repair-subcategory">維修次分類</label>
                  <select
                    id="affairs-repair-subcategory"
                    v-model="repairSubcategory"
                    class="affairs-select"
                  >
                    <option value="">---------</option>
                    <option
                      v-for="s in repairSubcategories"
                      :key="s"
                      :value="s"
                    >
                      {{ s }}
                    </option>
                  </select>
                </div>
              </div>
              <div
                v-else
                class="affairs-field-grid"
              >
                <div class="affairs-field">
                  <label for="affairs-feedback-category">意見大類</label>
                  <select
                    id="affairs-feedback-category"
                    v-model="feedbackCategory"
                    class="affairs-select"
                  >
                    <option value="">---------</option>
                    <option
                      v-for="c in feedbackCategories"
                      :key="c"
                      :value="c"
                    >
                      {{ c }}
                    </option>
                  </select>
                </div>
                <div class="affairs-field">
                  <label for="affairs-feedback-subcategory">意見次分類</label>
                  <select
                    id="affairs-feedback-subcategory"
                    v-model="feedbackSubcategory"
                    class="affairs-select"
                  >
                    <option value="">---------</option>
                    <option
                      v-for="s in feedbackSubcategories"
                      :key="s"
                      :value="s"
                    >
                      {{ s }}
                    </option>
                  </select>
                </div>
              </div>
            </div>

            <!-- 步驟 2：填寫內容 -->
            <div
              v-show="affairsStep === 2"
              class="affairs-guided-step"
            >
              <div class="work-card-title">填寫內容</div>
              <div
                v-if="affairsMode === 'repair'"
                class="affairs-field"
              >
                <label for="affairs-repair-content">內容</label>
                <textarea
                  id="affairs-repair-content"
                  v-model="repairContent"
                  class="affairs-textarea"
                  placeholder="請描述維修位置、故障情況及需要跟進的內容"
                />
              </div>
              <div
                v-else
                class="affairs-field"
              >
                <label for="affairs-feedback-content">內容</label>
                <textarea
                  id="affairs-feedback-content"
                  v-model="feedbackContent"
                  class="affairs-textarea"
                  placeholder="請描述事項位置、情況及需要跟進的內容"
                />
              </div>
            </div>

            <!-- 步驟 3：上載圖片/影片 -->
            <div
              v-show="affairsStep === 3"
              class="affairs-guided-step"
            >
              <div class="work-card-title">上載圖片/影片</div>
              <div class="affairs-step-hint">
                如有相關相片或影片，可一併提交予管理處參考。
              </div>
              <div class="affairs-upload-box">
                <strong>提交圖片/影片</strong>
                <span>支援選擇多個圖片或影片檔案，實際上載限制可按後台設定調整。</span>
                <input
                  type="file"
                  accept="image/*,video/*"
                  multiple
                  @change="onMediaChange"
                >
              </div>
            </div>

            <!-- 步驟 4：確認提交 -->
            <div
              v-show="affairsStep === 4"
              class="affairs-guided-step"
            >
              <div class="work-card-title">確認提交</div>
              <div
                v-if="affairsMode === 'repair'"
                class="affairs-step-hint"
              >
                請確認資料無誤。提交後由管理處安排跟進。
              </div>
              <div
                v-else
                class="affairs-step-hint"
              >
                閣下提供之意見將絕對保密。
              </div>
              <div class="affairs-review-list">
                <div class="affairs-review-row">
                  <span>類型</span>
                  <strong>{{ reviewMode }}</strong>
                </div>
                <div class="affairs-review-row">
                  <span>大廈</span>
                  <strong>{{ selectedFeedbackBuilding || '---------' }}</strong>
                </div>
                <div class="affairs-review-row">
                  <span>大類</span>
                  <strong>{{ reviewCategory }}</strong>
                </div>
                <div class="affairs-review-row">
                  <span>次分類</span>
                  <strong>{{ reviewSubcategory }}</strong>
                </div>
                <div class="affairs-review-row">
                  <span>內容</span>
                  <strong>{{ reviewContent }}</strong>
                </div>
                <div class="affairs-review-row">
                  <span>附件</span>
                  <strong>{{ reviewMedia }}</strong>
                </div>
              </div>
            </div>

            <div class="affairs-guided-actions">
              <button
                type="button"
                class="work-action secondary"
                :disabled="affairsStep === 1"
                @click="setAffairsStep(affairsStep - 1)"
              >
                上一步
              </button>
              <button
                v-if="affairsStep < 4"
                type="button"
                class="work-action"
                @click="nextAffairsStep"
              >
                下一步
              </button>
              <button
                v-else
                type="button"
                class="work-action"
                @click="submitAffairsFeedback"
              >
                提交
              </button>
            </div>
          </section>
        </div>

        <!-- 智能門禁 -->
        <div
          v-show="activeTab === 'affairs-access'"
          class="work-panel"
          :class="{ on: activeTab === 'affairs-access' }"
          data-work-panel="affairs-access"
        >
          <section class="work-hero">
            <div>
              <div class="work-kicker">Access</div>
              <h2 class="work-title">智能門禁</h2>
              <p class="work-desc">查看已綁定大廈的門禁權限、通行密碼、二維碼及最近開門記錄。</p>
            </div>
            <button
              type="button"
              class="work-action"
              :disabled="accessLoading"
              @click="loadBuildingAccess()"
            >
              {{ accessLoading ? '載入中' : '重新整理' }}
            </button>
          </section>

          <section class="access-summary-grid">
            <div class="access-summary-card">
              <div class="work-card-title">目前大廈</div>
              <div class="access-stat-value">{{ accessBuildingName }}</div>
              <div class="work-stat-label">依登入會員綁定資料顯示</div>
            </div>
            <div class="access-summary-card">
              <div class="work-card-title">門禁數量</div>
              <div class="work-stat">{{ accessDoors.length }}</div>
              <div class="work-stat-label">目前可見門禁</div>
            </div>
            <div class="access-summary-card">
              <div class="work-card-title">可開門</div>
              <div class="work-stat">{{ accessAllowedDoorCount }}</div>
              <div class="work-stat-label">已授權門禁</div>
            </div>
            <div class="access-summary-card">
              <div class="work-card-title">資料狀態</div>
              <span
                class="work-chip"
                :class="accessStatusClass"
              >{{ accessStatusText }}</span>
              <div class="work-stat-label">{{ accessQRCodeDoorCount }} 個門禁支援二維碼</div>
            </div>
          </section>

          <div
            v-if="accessError || accessMessage"
            class="access-banner"
            :class="{ warn: accessError, good: accessMessage && !accessError }"
          >
            {{ accessError || accessMessage }}
          </div>

          <section class="access-door-section">
            <div class="building-section-head">
              <div class="work-card-title">門禁列表</div>
              <span class="work-chip">{{ accessLoading ? '載入中' : `${accessDoors.length} 個門禁` }}</span>
            </div>

            <div
              v-if="accessLoading && accessDoors.length === 0"
              class="access-empty"
            >
              正在讀取門禁資料。
            </div>
            <div
              v-else-if="accessDoors.length === 0"
              class="access-empty"
            >
              目前未有可顯示的門禁資料。
            </div>
            <div
              v-else
              class="access-table-wrap"
            >
              <table class="work-table access-door-table">
                <thead>
                  <tr>
                    <th>門禁</th>
                    <th>門號</th>
                    <th>所屬大廈</th>
                    <th>鏡頭</th>
                    <th>權限</th>
                    <th>通行密碼</th>
                    <th>有效期</th>
                    <th>操作</th>
                  </tr>
                </thead>
                <tbody>
                  <tr
                    v-for="door in accessDoors"
                    :key="accessDoorID(door) || accessDoorTitle(door)"
                  >
                    <td class="access-door-title-cell">
                      <strong>{{ accessDoorTitle(door) }}</strong>
                      <span>{{ textValue(door.serial) }}</span>
                    </td>
                    <td>{{ textValue(door.door_no) }}</td>
                    <td>{{ textValue(door.building_id) }}</td>
                    <td>{{ textValue(door.camera?.title) }}</td>
                    <td>
                      <div class="access-table-tags">
                        <span
                          class="work-chip"
                          :class="door.has_permission ? 'good' : 'warn'"
                        >{{ door.has_permission ? '可開門' : '未授權' }}</span>
                        <span
                          v-if="door.is_public"
                          class="work-chip brand"
                        >公共門</span>
                        <span
                          v-if="door.is_qrcode_enabled"
                          class="work-chip"
                        >二維碼</span>
                      </div>
                    </td>
                    <td>{{ accessDoorPasswordText(door) }}</td>
                    <td class="access-door-period-cell">
                      <div>
                        <span>密碼</span>
                        <strong>{{ accessTimeRange(door.password?.start_time, door.password?.end_time) }}</strong>
                      </div>
                      <div>
                        <span>二維碼</span>
                        <strong>{{ accessTimeRange(door.qrcode?.start_time, door.qrcode?.end_time) }}</strong>
                      </div>
                    </td>
                    <td>
                      <div class="access-table-actions">
                        <button
                          type="button"
                          class="work-mini-btn primary"
                          :disabled="!door.has_permission || openingDoorID === accessDoorID(door)"
                          @click="openAccessDoor(door)"
                        >
                          {{ openingDoorID === accessDoorID(door) ? '處理中' : '開門' }}
                        </button>
                        <button
                          v-if="door.password?.value"
                          type="button"
                          class="work-mini-btn"
                          @click="toggleAccessPassword(door)"
                        >
                          {{ accessDoorPasswordVisible(door) ? '隱藏密碼' : '查看密碼' }}
                        </button>
                        <button
                          v-if="door.is_qrcode_enabled && door.qrcode?.record_id"
                          type="button"
                          class="work-mini-btn"
                          :disabled="qrLoadingDoorID === accessDoorID(door)"
                          @click="generateAccessQRCode(door)"
                        >
                          {{ qrLoadingDoorID === accessDoorID(door) ? '生成中' : '二維碼' }}
                        </button>
                      </div>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </section>

          <section
            v-if="accessQRPanel"
            class="work-card access-qr-section"
          >
            <div class="access-qr-layout">
              <div class="access-qr-image">
                <QrCodeImage
                  :text="accessQRPanel.value"
                  :size="220"
                  alt="Door access QR code"
                />
              </div>
              <div>
                <div class="work-card-title">{{ accessQRPanel.doorTitle }}</div>
                <div class="work-card-sub">請於有效時間內使用此二維碼通行。</div>
                <div class="access-qr-meta">
                  <div>
                    <span>類型</span>
                    <strong>{{ accessQRPanel.term }}</strong>
                  </div>
                  <div>
                    <span>有效期至</span>
                    <strong>{{ accessQRPanel.expiresAt }}</strong>
                  </div>
                </div>
              </div>
            </div>
          </section>

          <section class="work-card">
            <div class="building-section-head">
              <div class="work-card-title">最近開門記錄</div>
            </div>
            <table class="work-table access-record-table">
              <thead>
                <tr>
                  <th>門禁</th>
                  <th>時間</th>
                  <th>方式</th>
                  <th>狀態</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="row in accessRecentRows"
                  :key="row.key"
                >
                  <td>{{ row.doorTitle }}</td>
                  <td>{{ row.openTime }}</td>
                  <td>{{ row.openType }}</td>
                  <td>
                    <span
                      class="work-chip"
                      :class="row.status"
                    >{{ row.statusText }}</span>
                  </td>
                </tr>
                <tr v-if="accessRecentRows.length === 0">
                  <td
                    class="building-empty-row"
                    colspan="4"
                  >
                    目前未有最近開門記錄。
                  </td>
                </tr>
              </tbody>
            </table>
          </section>
        </div>

        <!-- 視像監控 -->
        <div
          v-show="activeTab === 'affairs-icctv'"
          class="work-panel"
          :class="{ on: activeTab === 'affairs-icctv' }"
          data-work-panel="affairs-icctv"
        >
          <section class="work-hero">
            <div>
              <div class="work-kicker">ICCTV</div>
              <h2 class="work-title">視像監控</h2>
              <p class="work-desc">即時查看已授權的大廈鏡頭。</p>
            </div>
            <button
              type="button"
              class="work-action"
              :disabled="icctvLoading"
              @click="loadICCTV()"
            >
              {{ icctvLoading ? '載入中' : '重新整理' }}
            </button>
          </section>
          <section class="work-card icctv-panel-wide icctv-table-card">
            <div class="icctv-profile-head">
              <div>
                <h3>{{ icctvBuildingTitle }}</h3>
                <p>目前會員已綁定單位</p>
              </div>
              <span
                class="icctv-enabled-badge"
                :class="{ off: !icctvEnabled && !icctvLoading }"
              >
                {{ icctvStatusText }}
              </span>
            </div>

            <div class="icctv-table-wrap">
              <table class="work-table icctv-table">
                <thead>
                  <tr>
                    <th>鏡頭</th>
                    <th>狀態</th>
                    <th>操作</th>
                  </tr>
                </thead>
                <tbody>
                  <template
                    v-for="(camera, index) in icctvCameras"
                    :key="camera.id"
                  >
                    <tr :class="{ expanded: isICCTVCameraExpanded(camera.id) }">
                      <td>
                        <div class="icctv-camera-title">{{ icctvCameraName(camera, index) }}</div>
                        <div class="icctv-camera-channel">{{ camera.channel }}</div>
                      </td>
                      <td>
                        <span
                          class="icctv-table-status"
                          :class="camera.is_active && camera.url ? 'on' : 'off'"
                        >
                          {{ icctvCameraStatusText(camera) }}
                        </span>
                      </td>
                      <td>
                        <div class="icctv-row-actions">
                          <button
                            type="button"
                            class="work-mini-btn primary"
                            :disabled="!camera.url"
                            @click="toggleICCTVCamera(camera.id)"
                          >
                            {{ isICCTVCameraExpanded(camera.id) ? '收起' : '查看' }}
                          </button>
                          <button
                            type="button"
                            class="work-mini-btn"
                            :disabled="!camera.url"
                            @click="openIcctvWindow(camera)"
                          >
                            新窗口
                          </button>
                        </div>
                      </td>
                    </tr>
                    <tr
                      v-if="isICCTVCameraExpanded(camera.id)"
                      class="icctv-expanded-row"
                    >
                      <td colspan="3">
                        <div class="icctv-inline-viewer">
                          <iframe
                            :key="`${camera.id}-frame`"
                            :src="camera.url"
                            :title="icctvCameraFrameTitle(camera, index)"
                            allow="autoplay; fullscreen; encrypted-media; picture-in-picture"
                          />
                        </div>
                      </td>
                    </tr>
                  </template>
                  <tr v-if="icctvCameras.length === 0">
                    <td
                      class="building-empty-row"
                      colspan="3"
                    >
                      {{ icctvLoading ? '正在載入鏡頭資料。' : icctvError || '目前大廈未有鏡頭資料。' }}
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </section>
        </div>

        <!-- 設備監測 -->
        <div
          v-show="activeTab === 'affairs-equipment'"
          class="work-panel"
          :class="{ on: activeTab === 'affairs-equipment' }"
          data-work-panel="affairs-equipment"
        >
          <section class="work-hero">
            <div>
              <div class="work-kicker">Equipment</div>
              <h2 class="work-title">設備監測</h2>
              <p class="work-desc">查看大廈設備狀態、巡檢紀錄與異常提醒。</p>
            </div>
            <button
              type="button"
              class="work-action secondary"
              @click="refreshEquipment"
            >
              重新整理
            </button>
          </section>
          <section class="work-card">
            <div class="work-card-title">設備狀態</div>
            <table class="work-table">
              <thead>
                <tr>
                  <th>設備</th>
                  <th>位置</th>
                  <th>狀態</th>
                  <th>更新時間</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="(r, i) in equipmentRows"
                  :key="i"
                >
                  <td>{{ r.device }}</td>
                  <td>{{ r.location }}</td>
                  <td>
                    <span
                      class="work-chip"
                      :class="r.status"
                    >{{ r.statusText }}</span>
                  </td>
                  <td>{{ r.updatedAt }}</td>
                </tr>
              </tbody>
            </table>
          </section>
        </div>
      </main>
    </div>
  </div>
</template>

<style scoped>
/* 1. 頁面容器與雙欄布局 */
#page-affairs {
  min-height: calc(100vh - var(--nav-h, 52px));
  background: var(--sur-2);
}

.work-shell {
  display: grid;
  grid-template-columns: 240px minmax(0, 1fr);
  gap: 16px;
  max-width: var(--layout-page-max-width);
  margin: 0 auto;
  padding: 12px 24px 16px;
  color: var(--ink);
}

/* 2. 左側大廈導航 */
.work-sidebar {
  position: sticky;
  top: 72px;
  align-self: start;
  border: 1px solid var(--bdr);
  border-radius: 8px;
  background: var(--sur);
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

.work-nav-item.on {
  background: transparent;
  color: var(--brand);
  font-weight: 700;
  outline: none;
}

.work-nav-item.on:hover {
  background: var(--brand-light);
}

.work-nav-item.on::after {
  transform: scaleX(1);
}

/* 3. 右側主內容 */
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

/* 4. Hero 區 */
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
  padding: 0 0 10px;
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

.work-action {
  border: 0;
  border-radius: 6px;
  background: var(--brand);
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
  background: var(--sur);
  color: var(--ink);
}

.work-action:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* 5. Card */
.work-card {
  border: 1px solid var(--bdr);
  border-radius: 8px;
  background: var(--sur);
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

.work-stat {
  color: var(--brand);
  font-size: 24px;
  font-weight: 700;
  line-height: 1.1;
}

.work-stat-label {
  margin-top: 6px;
  color: var(--ink-3);
  font-size: 12px;
}

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

/* 6. Chip */
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

/* 7. Table */
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

.work-table tr:last-child td {
  border-bottom: 0;
}

.notice-table-actions {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
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

.work-mini-btn:disabled,
.work-mini-btn.primary:disabled {
  border-color: var(--bdr);
  background: var(--sur-3);
  color: var(--ink-3);
  cursor: not-allowed;
  opacity: 1;
}

/* 8. Notice admin */
.notice-admin-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.notice-admin-card {
  border: 1px solid var(--bdr);
  border-radius: 8px;
  background: var(--sur);
  padding: 14px;
}

.notice-admin-card h3 {
  margin: 0 0 10px;
  color: var(--ink);
  font-size: 14px;
  font-weight: 700;
}

.notice-admin-label {
  display: block;
  margin-bottom: 6px;
  color: var(--ink-3);
  font-size: 11px;
  font-weight: 700;
}

.notice-select {
  width: 100%;
  height: 36px;
  border: 1px solid var(--bdr);
  border-radius: 6px;
  background: var(--sur);
  color: var(--ink);
  font-family: inherit;
  font-size: 13px;
  font-weight: 700;
  padding: 0 10px;
}

.building-readonly-select {
  display: flex;
  align-items: center;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.notice-current {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 10px;
  color: var(--ink-3);
  font-size: 12px;
  font-weight: 700;
}

.notice-action-btn {
  border: 1px solid var(--bdr);
  border-radius: 6px;
  background: var(--sur);
  color: var(--ink);
  cursor: pointer;
  font-family: inherit;
  font-size: 12px;
  font-weight: 700;
  padding: 6px 10px;
}

.notice-action-btn:hover {
  border-color: var(--brand-mid);
  color: var(--brand);
}

.notice-table {
  min-width: 780px;
}

.notice-table td:first-child {
  max-width: 230px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

#page-affairs [data-work-panel="affairs-notices"] .work-card {
  overflow: auto;
}

/* 9. Building profile */
.building-summary-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
}

.building-field-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0;
  border: 1px solid var(--bdr);
  border-radius: 8px;
  background: var(--sur);
  overflow: hidden;
}

.building-field {
  display: grid;
  grid-template-columns: 150px minmax(0, 1fr);
  gap: 12px;
  border-right: 1px solid var(--sur-3);
  border-bottom: 1px solid var(--sur-3);
  padding: 13px 14px;
}

.building-field:nth-child(2n) {
  border-right: 0;
}

.building-field span:first-child {
  color: var(--ink-3);
  font-size: 12px;
  font-weight: 700;
}

.building-field strong {
  color: var(--ink);
  font-size: 13px;
  font-weight: 700;
  line-height: 1.5;
  word-break: break-word;
}

.building-map-frame {
  display: block;
  width: 100%;
  height: 260px;
  border: 0;
  border-radius: 8px;
  background: var(--sur-2);
}

.building-doc-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
}

.building-doc-card {
  border: 1px solid var(--bdr);
  border-radius: 8px;
  background: var(--sur);
  padding: 14px;
}

.building-doc-card .work-card-title {
  margin-bottom: 6px;
}

.building-doc-count {
  margin-top: 10px;
  color: var(--brand);
  font-size: 22px;
  font-weight: 800;
  line-height: 1;
}

.building-doc-empty {
  margin-top: 10px;
  color: var(--ink-3);
  font-size: 12px;
  line-height: 1.6;
}

.building-section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
}

.building-section-head .work-card-title {
  margin: 0;
}

.building-file-table {
  table-layout: fixed;
}

.building-file-title-col {
  width: auto;
}

.building-file-date-col,
.building-file-month-col {
  width: 128px;
}

.building-file-action-col {
  width: 96px;
}

.building-file-table th:nth-child(2),
.building-file-table th:nth-child(3),
.building-file-meta-cell {
  text-align: center;
}

.building-file-table th:nth-child(4),
.building-file-action-cell {
  text-align: center;
}

.building-file-title-cell {
  overflow-wrap: anywhere;
  word-break: break-word;
}

.building-file-action-cell {
  white-space: nowrap;
}

.building-empty-row {
  color: var(--ink-3);
  font-size: 13px;
  font-weight: 600;
  text-align: center;
}

.building-link {
  color: var(--brand);
  font-weight: 700;
  text-decoration: none;
}

.building-link:hover {
  text-decoration: underline;
}

/* 10. 智能門禁 */
.access-summary-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
}

.access-summary-card {
  min-width: 0;
  border: 1px solid var(--bdr);
  border-radius: 8px;
  background: var(--sur);
  padding: 14px;
}

.access-stat-value {
  color: var(--ink);
  font-size: 16px;
  font-weight: 800;
  line-height: 1.35;
  overflow-wrap: anywhere;
}

.access-banner {
  border: 1px solid var(--success);
  border-radius: 8px;
  background: var(--success-bg);
  color: var(--success);
  font-size: 13px;
  font-weight: 700;
  padding: 11px 14px;
}

.access-banner.warn {
  border-color: var(--warning);
  background: var(--warning-bg);
  color: var(--warning);
}

.access-door-section {
  border: 1px solid var(--bdr);
  border-radius: 8px;
  background: var(--sur);
  padding: 16px;
}

.access-empty {
  display: grid;
  min-height: 120px;
  place-items: center;
  border: 1px dashed var(--bdr-2);
  border-radius: 8px;
  background: var(--sur-2);
  color: var(--ink-3);
  font-size: 13px;
  font-weight: 700;
}

.access-table-wrap {
  overflow-x: auto;
  border: 1px solid var(--bdr);
  border-radius: 8px;
  background: #fff;
}

.access-door-table {
  min-width: 1040px;
  table-layout: fixed;
}

.access-door-table th:nth-child(1) {
  width: 180px;
}

.access-door-table th:nth-child(2) {
  width: 74px;
}

.access-door-table th:nth-child(3),
.access-door-table th:nth-child(4) {
  width: 110px;
}

.access-door-table th:nth-child(5) {
  width: 160px;
}

.access-door-table th:nth-child(6) {
  width: 96px;
}

.access-door-table th:nth-child(7) {
  width: 210px;
}

.access-door-table th:nth-child(8) {
  width: 160px;
}

.access-door-title-cell strong {
  display: block;
  color: var(--ink);
  font-size: 13px;
  font-weight: 800;
  line-height: 1.4;
  overflow-wrap: anywhere;
}

.access-door-title-cell span {
  display: block;
  margin-top: 3px;
  color: var(--ink-3);
  font-size: 11px;
  font-weight: 700;
  overflow-wrap: anywhere;
}

.access-table-tags,
.access-table-actions {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

.access-door-period-cell {
  display: grid;
  gap: 7px;
}

.access-door-period-cell span,
.access-qr-meta span {
  display: block;
  color: var(--ink-3);
  font-size: 11px;
  font-weight: 700;
}

.access-door-period-cell strong,
.access-qr-meta strong {
  display: block;
  margin-top: 4px;
  color: var(--ink);
  font-size: 13px;
  font-weight: 700;
  line-height: 1.45;
  overflow-wrap: anywhere;
}

.access-qr-section {
  overflow: hidden;
}

.access-qr-layout {
  display: grid;
  grid-template-columns: 240px minmax(0, 1fr);
  gap: 18px;
  align-items: center;
}

.access-qr-image {
  display: grid;
  min-height: 240px;
  place-items: center;
  border: 1px solid var(--bdr);
  border-radius: 8px;
  background: #fff;
}

.access-qr-image img {
  display: block;
  width: 220px;
  height: 220px;
}

.access-qr-meta {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
  margin-top: 16px;
}

.access-record-table {
  table-layout: fixed;
}

.access-record-table th:nth-child(4),
.access-record-table td:nth-child(4) {
  width: 110px;
  text-align: center;
}

/* 11. Accounting */
.acct-card-wrap {
  padding: 0;
  overflow: hidden;
}

/* 1. 帳目 tab 列 */
.acct-tabs {
  display: flex;
  align-items: center;
  gap: 0;
  border-bottom: 1px solid var(--bdr);
  background: #fff;
}

.acct-tab {
  border: 0;
  border-right: 1px solid var(--bdr);
  background: #fff;
  color: var(--brand);
  cursor: pointer;
  font-family: inherit;
  font-size: 13px;
  font-weight: 800;
  padding: 13px 22px;
}

.acct-tab.on {
  background: var(--sur);
  color: var(--ink);
  box-shadow: inset 0 3px 0 var(--brand);
}

.acct-body {
  padding: 16px;
}

.acct-subpanel {
  display: none;
}

.acct-subpanel.on {
  display: grid;
  gap: 12px;
}

/* 2. 帳目提示框 */
.acct-note {
  border: 1px solid var(--bdr);
  border-left: 3px solid var(--brand);
  border-radius: 6px;
  background: #fff;
  color: var(--ink-2);
  font-size: 12px;
  font-weight: 600;
  line-height: 1.7;
  padding: 12px 14px;
}

.acct-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.acct-toolbar-title {
  margin: 0;
}

/* 3. 帳目搜尋框 */
.acct-search {
  max-width: 320px;
  border: 1px solid var(--bdr);
  border-radius: 6px;
  background: #fff;
  color: var(--ink);
  font-family: inherit;
  font-size: 13px;
  font-weight: 600;
  padding: 9px 12px;
}

/* 4. 帳目表格 */
.acct-table-wrap {
  overflow: auto;
  border: 1px solid var(--bdr);
  border-radius: 8px;
  background: #fff;
}

.acct-table {
  width: 100%;
  min-width: 760px;
  border-collapse: collapse;
  font-size: 13px;
}

.acct-table th {
  position: sticky;
  top: 0;
  background: #C7EAF0;
  color: var(--ink);
  font-size: 13px;
  font-weight: 900;
  padding: 11px 12px;
  text-align: left;
  z-index: 1;
}

.acct-table td {
  border-top: 1px solid var(--sur-3);
  color: var(--ink-2);
  font-weight: 700;
  padding: 10px 12px;
  white-space: nowrap;
}

.acct-table tr:nth-child(even) td {
  background: #FAFAFA;
}

/* 5. 帳目狀態色 */
.acct-paid {
  color: var(--success);
}

.acct-due {
  color: var(--error);
}

/* 6. 帳目下載按鈕 */
.acct-download {
  border: 0;
  border-radius: 6px;
  background: #18BFE0;
  color: #041016;
  cursor: pointer;
  font-family: inherit;
  font-size: 12px;
  font-weight: 900;
  padding: 7px 12px;
}

/* 7. 開發中佔位 */
.acct-dev {
  display: flex;
  min-height: 280px;
  align-items: center;
  justify-content: center;
  border: 1px dashed var(--bdr-2);
  border-radius: 8px;
  background: #fff;
  color: var(--ink-3);
  font-size: 16px;
  font-weight: 800;
}

/* 12. 意見提供入口卡片 */
.affairs-entry-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.affairs-entry-card {
  display: grid;
  grid-template-columns: 44px minmax(0, 1fr);
  gap: 12px;
  align-items: center;
  border: 1px solid var(--bdr);
  border-radius: 8px;
  background: var(--sur);
  color: var(--ink);
  cursor: pointer;
  font-family: inherit;
  padding: 16px;
  text-align: left;
}

.affairs-entry-card:hover,
.affairs-entry-card.on {
  border-color: var(--brand);
  background: var(--brand-light);
}

.affairs-entry-icon {
  display: flex;
  width: 44px;
  height: 44px;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  background: var(--sur-2);
  color: var(--brand);
}

.affairs-entry-icon svg {
  width: 24px;
  height: 24px;
  stroke: currentColor;
}

.affairs-entry-card > span:last-child strong {
  display: block;
  font-size: 14px;
  font-weight: 800;
  line-height: 1.4;
}

.affairs-entry-card > span:last-child span {
  display: block;
  margin-top: 4px;
  color: var(--ink-3);
  font-size: 12px;
  line-height: 1.5;
}

/* 13. 意見提供引導式表單 */
.affairs-feedback-form {
  display: grid;
  gap: 18px;
}

.affairs-field {
  display: grid;
  gap: 7px;
  max-width: 560px;
}

.affairs-field label {
  color: var(--ink);
  font-size: 13px;
  font-weight: 700;
}

.affairs-select,
.affairs-textarea {
  width: 100%;
  border: 1px solid var(--bdr);
  border-radius: 6px;
  background: #fff;
  color: var(--ink);
  font-family: inherit;
  font-size: 13px;
  padding: 10px 12px;
}

.affairs-textarea {
  min-height: 160px;
  resize: vertical;
  line-height: 1.6;
}

.affairs-field-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px;
}

.affairs-guided-shell {
  display: grid;
  gap: 16px;
}

.affairs-stepper {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 8px;
}

.affairs-step-indicator {
  display: flex;
  align-items: center;
  gap: 7px;
  border: 1px solid var(--bdr);
  border-radius: 8px;
  background: var(--sur-2);
  color: var(--ink-3);
  font-size: 12px;
  font-weight: 700;
  padding: 9px 10px;
}

.affairs-step-indicator b {
  display: inline-flex;
  width: 22px;
  height: 22px;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  background: var(--sur);
  color: var(--ink-3);
  font-size: 11px;
}

.affairs-step-indicator.on {
  border-color: var(--brand);
  background: var(--brand-light);
  color: var(--brand);
}

.affairs-step-indicator.on b {
  background: var(--brand);
  color: #fff;
}

.affairs-guided-step {
  display: grid;
  gap: 14px;
  min-height: 220px;
}

.affairs-step-hint {
  border-left: 3px solid var(--brand);
  background: var(--brand-light);
  color: var(--ink-2);
  font-size: 12px;
  font-weight: 600;
  line-height: 1.6;
  padding: 10px 12px;
}

.affairs-guided-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  border-top: 1px solid var(--sur-3);
  padding-top: 12px;
}

.affairs-review-list {
  display: grid;
  gap: 8px;
  font-size: 13px;
}

.affairs-review-row {
  display: grid;
  grid-template-columns: 110px minmax(0, 1fr);
  gap: 10px;
  border-bottom: 1px solid var(--sur-3);
  padding-bottom: 8px;
  color: var(--ink-2);
}

.affairs-review-row span:first-child {
  color: var(--ink-3);
  font-weight: 700;
}

.affairs-upload-box {
  display: grid;
  gap: 10px;
  border: 1px dashed var(--bdr-2);
  border-radius: 8px;
  background: var(--sur-2);
  padding: 16px;
}

.affairs-upload-box strong {
  color: var(--ink);
  font-size: 13px;
  font-weight: 800;
}

.affairs-upload-box span {
  color: var(--ink-3);
  font-size: 12px;
  line-height: 1.6;
}

.affairs-upload-box input {
  width: 100%;
  border: 1px solid var(--bdr);
  border-radius: 6px;
  background: var(--sur);
  color: var(--ink-2);
  font-family: inherit;
  font-size: 12px;
  padding: 9px;
}

.affairs-record-card {
  border: 1px solid var(--bdr);
  border-radius: 8px;
  background: var(--sur-2);
  padding: 12px 14px;
  color: var(--ink-2);
  font-size: 12px;
  line-height: 1.7;
}

.affairs-record-card strong {
  display: block;
  margin-bottom: 4px;
  color: var(--ink);
  font-size: 13px;
}

/* 14. ICCTV */
.icctv-panel-wide {
  grid-column: 1 / -1;
}

.icctv-table-card {
  padding: 0;
  overflow: hidden;
}

.icctv-profile-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 18px;
  border-bottom: 1px solid var(--bdr);
  background: #fff;
  padding: 18px;
}

.icctv-profile-head h3 {
  margin: 0;
  color: var(--ink);
  font-size: 18px;
  font-weight: 900;
  line-height: 1.35;
}

.icctv-profile-head p {
  margin: 6px 0 0;
  color: var(--ink-3);
  font-size: 13px;
  font-weight: 700;
  line-height: 1.5;
}

.icctv-enabled-badge {
  display: inline-flex;
  min-height: 30px;
  align-items: center;
  border-radius: 999px;
  background: var(--success-bg);
  color: var(--success);
  font-size: 12px;
  font-weight: 900;
  padding: 6px 12px;
  white-space: nowrap;
}

.icctv-enabled-badge.off {
  background: var(--sur-2);
  color: var(--ink-4);
}

.staff-select {
  height: 36px;
  border: 1px solid var(--bdr);
  border-radius: 6px;
  background: var(--sur);
  color: var(--ink);
  font-family: inherit;
  font-size: 13px;
  font-weight: 700;
  padding: 0 12px;
}

.icctv-table-wrap {
  overflow-x: auto;
}

.icctv-table {
  min-width: 760px;
  table-layout: fixed;
}

.icctv-table th:nth-child(1),
.icctv-table td:nth-child(1) {
  width: 44%;
}

.icctv-table th:nth-child(2),
.icctv-table td:nth-child(2) {
  width: 20%;
}

.icctv-table th:nth-child(3),
.icctv-table td:nth-child(3) {
  width: 36%;
}

.icctv-table tbody tr.expanded > td {
  background: var(--brand-light);
}

.icctv-camera-title {
  color: var(--ink);
  font-size: 13px;
  font-weight: 900;
}

.icctv-camera-channel {
  margin-top: 4px;
  color: var(--ink-3);
  font-size: 12px;
  font-weight: 700;
}

.icctv-table-status {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  font-size: 11px;
  font-weight: 900;
  padding: 5px 9px;
  white-space: nowrap;
}

.icctv-table-status.on {
  background: var(--success-bg);
  color: var(--success);
}

.icctv-table-status.off {
  background: var(--sur-2);
  color: var(--ink-4);
}

.icctv-row-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.icctv-expanded-row td {
  background: #fff;
  padding: 0;
}

.icctv-inline-viewer {
  border-top: 1px solid var(--bdr);
  background: #0F172A;
}

.icctv-inline-viewer iframe {
  display: block;
  width: 100%;
  height: 520px;
  border: 0;
  background: #0F172A;
}

/* 15. 響應式設計 */
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

  .affairs-entry-grid,
  .affairs-field-grid,
  .affairs-stepper {
    grid-template-columns: 1fr;
  }

  .building-summary-grid,
  .building-doc-grid,
  .access-summary-grid,
  .building-field-grid {
    grid-template-columns: 1fr;
  }

  .building-field {
    grid-template-columns: 1fr;
    border-right: 0;
  }

  .access-qr-layout,
  .access-qr-meta {
    grid-template-columns: 1fr;
  }

  .acct-tabs {
    overflow: auto;
  }

  .acct-tab {
    white-space: nowrap;
  }
}

@media (max-width: 767px) {
  .work-shell {
    padding: 12px;
    gap: 12px;
  }

  .work-nav {
    grid-template-columns: 1fr;
  }

  .work-hero {
    align-items: flex-start;
    flex-direction: column;
  }

  .access-summary-grid {
    grid-template-columns: 1fr;
  }

  .icctv-profile-head {
    align-items: stretch;
    flex-direction: column;
  }

  .icctv-inline-viewer iframe {
    height: 360px;
  }

  .notice-admin-grid {
    grid-template-columns: 1fr;
  }

  .acct-search {
    width: 100%;
    max-width: none;
  }

  .acct-toolbar {
    flex-direction: column;
    align-items: stretch;
  }

  .affairs-guided-actions {
    flex-direction: column-reverse;
    align-items: stretch;
  }

  .affairs-guided-actions .work-action {
    width: 100%;
  }
}
</style>

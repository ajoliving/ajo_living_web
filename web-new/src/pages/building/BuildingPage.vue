<!--
 * 我的大廈頁。
 * 1. 高保真還原 HTML 設計稿 page-affairs 雙欄布局（左側大廈導航 + 右側面板）。
 * 2. 提供最新通告、大廈資料、大廈財務、業戶帳目、申請表格、意見提供/維修報修、智能門禁、視像監控八個面板。
 * 3. 意見提供面板包含最近記錄（可展開內容）、入口卡片、引導式四步提交流程。
 * 4. 大廈資料讀取目前會員 iSmart 綁定大廈。
-->
<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';

import QrCodeImage from '@/shared/components/base/QrCodeImage.vue';
import { useSessionStore } from '@/stores/session';
import {
  fetchMemberICCTVPublicCameras,
  fetchMemberIsmartBuildingAccess,
  fetchMemberIsmartBuildingInfo,
  fetchMemberIsmartManagementFees,
  fetchMemberIsmartBuildingNotices,
  fetchMemberIsmartOtherFees,
  fetchMemberPosBuildings,
  fetchPosBuildings,
  generateMemberIsmartDoorQRCode,
  openMemberIsmartDoor,
  type ICCTVCameraSummary,
  type IsmartAccessDoor,
  type IsmartAccessRecord,
  type IsmartBuildingDocument,
  type IsmartBuildingAccessResponse,
  type IsmartBuildingInfoResponse,
  type IsmartBuildingNotice,
  type IsmartBuildingNoticesResponse,
  type IsmartManagementFeeTableRow,
  type IsmartOtherFeeRow,
  type IsmartRecentAccessGroup,
} from '@/httpapis/building';
import type { PosBuilding } from '@/model/community';
import {
  fetchPOSIntegrationTransactionsByDate,
  fetchPOSIntegrationTransactionsByUnit,
  fetchPOSIntegrationUnpaidInvoices,
} from '@/httpapis/payments';
import type {
  POSIntegrationPaymentDetail,
  POSIntegrationPaymentTransaction,
  POSIntegrationUnpaidInvoice,
} from '@/model/payments';
import {
  buildBuildingNameMap,
  indexedBuildingName,
  memberCommunityName,
  posBuildingID,
  posBuildingName,
  preferredMemberBuildingID,
  usableBuildingName,
} from './composables/building-display';

// 1. 型別定義
type AffairsTab =
  | 'affairs-notices'
  | 'affairs-building'
  | 'affairs-finance'
  | 'affairs-owner-account'
  | 'affairs-forms'
  | 'affairs-feedback'
  | 'affairs-access'
  | 'affairs-icctv';

type AffairsMode = 'repair' | 'feedback';
type FinanceSubTab = 'management-overview' | 'financial-reports' | 'audit-reports';
type OwnerAccountTab = 'owner-unpaid' | 'owner-records';

interface NavItem {
  target: AffairsTab;
  label: string;
  needsApi?: boolean;
}

interface SelectOption {
  value: string;
  label: string;
}

interface NoticeRow {
  key: string;
  code: string;
  title: string;
  type: string;
  publishDate: string;
  expireDate: string;
  fileUrl: string;
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

interface BuildingFileRow {
  key: string;
  title: string;
  date: string;
  month: string;
  url: string;
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

interface OwnerUnitContext {
  buildingID: string;
  buildingName: string;
  floor: string;
  unit: string;
  unitID: string;
}

interface OwnerPaymentRecordRow {
  key: string;
  receiptID: string;
  paymentID: string;
  inputTime: string;
  tranTime: string;
  unit: string;
  item: string;
  term: string;
  amount: number;
  payType: string;
  status: string;
  statusClass: string;
  remark: string;
}

// 2. 會員狀態
const sessionStore = useSessionStore();
const { t, locale } = useI18n();

// 2.1 建立日期查詢預設值
const formatDateInputValue = (date: Date): string => {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  return `${year}-${month}-${day}`;
};

const currentMonthStart = (): string => {
  const date = new Date();
  return formatDateInputValue(new Date(date.getFullYear(), date.getMonth(), 1));
};

const currentDateValue = (): string => formatDateInputValue(new Date());

// 2.2 按目前語系格式化日期、數字及金額
const displayLocale = computed(() => (locale.value === 'en' ? 'en-HK' : 'zh-HK'));
const translateMessage = (key: string): string => (key ? t(key) : '');

const formatLocaleNumber = (value: number): string =>
  new Intl.NumberFormat(displayLocale.value).format(value);

const parseDisplayDate = (value: unknown): { date: Date; hasTime: boolean } | null => {
  const raw = String(value ?? '').trim();
  if (!raw || raw === '-') return null;
  const match = raw.match(
    /^(\d{4})[-/.年](\d{1,2})(?:[-/.月](\d{1,2}))?日?(?:[ T](\d{1,2}):(\d{2})(?::(\d{2}))?)?/,
  );
  if (!match) return null;

  const year = Number(match[1]);
  const month = Number(match[2]);
  const day = Number(match[3] ?? 1);
  const hour = Number(match[4] ?? 0);
  const minute = Number(match[5] ?? 0);
  const second = Number(match[6] ?? 0);
  const date = new Date(year, month - 1, day, hour, minute, second);
  if (Number.isNaN(date.getTime())) return null;
  return { date, hasTime: Boolean(match[4]) };
};

const formatLocaleDateValue = (value: unknown): string => {
  const raw = String(value ?? '').trim();
  if (!raw) return '-';
  const parsed = parseDisplayDate(raw);
  if (!parsed) return raw;
  return new Intl.DateTimeFormat(displayLocale.value, parsed.hasTime
    ? { year: 'numeric', month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' }
    : { year: 'numeric', month: 'short', day: 'numeric' }).format(parsed.date);
};

const formatLocaleMonthValue = (value: unknown): string => {
  const raw = String(value ?? '').trim();
  if (!raw) return '-';
  const parsed = parseDisplayDate(raw);
  if (!parsed) return raw;
  return new Intl.DateTimeFormat(displayLocale.value, { year: 'numeric', month: 'short' }).format(parsed.date);
};

// 3. 主面板與子面板狀態
const activeTab = ref<AffairsTab>('affairs-notices');
const buildingInfoLoading = ref(false);
const buildingInfoError = ref('');
const selectedBuildingID = ref('');
const buildingDirectory = ref<PosBuilding[]>([]);
const ismartBuildingProfile = ref<IsmartBuildingInfoResponse | null>(null);
const financeSubTab = ref<FinanceSubTab>('management-overview');
const financeReceivableLoading = ref(false);
const financeReceivableLoaded = ref(false);
const financeReceivableError = ref('');
const managementFeeRows = ref<IsmartManagementFeeTableRow[]>([]);
const otherFeeRows = ref<IsmartOtherFeeRow[]>([]);
const noticeLoading = ref(false);
const noticeError = ref('');
const selectedNoticeBuildingID = ref('');
const ismartNoticeProfile = ref<IsmartBuildingNoticesResponse | null>(null);
const ownerAccountTab = ref<OwnerAccountTab>('owner-unpaid');
const ownerUnpaidLoading = ref(false);
const ownerUnpaidLoaded = ref(false);
const ownerUnpaidError = ref('');
const ownerUnpaidInvoices = ref<POSIntegrationUnpaidInvoice[]>([]);
const ownerRecordsLoading = ref(false);
const ownerRecordsLoaded = ref(false);
const ownerRecordsError = ref('');
const ownerRecordsDateSearched = ref(false);
const ownerPaymentRecords = ref<POSIntegrationPaymentTransaction[]>([]);
const ownerRecordFromDate = ref(currentMonthStart());
const ownerRecordToDate = ref(currentDateValue());
const ownerRecordDateType = ref<'input_date' | 'tran_date'>('input_date');
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
const selectedFeedbackBuilding = ref('harmony');
const repairCategory = ref('');
const repairSubcategory = ref('');
const feedbackCategory = ref('');
const feedbackSubcategory = ref('');
const repairContent = ref('');
const feedbackContent = ref('');
const mediaFileName = ref('');

// 5. 左側導航項目
const navItems = computed<NavItem[]>(() => [
  { target: 'affairs-notices', label: t('building.nav.notices') },
  { target: 'affairs-building', label: t('building.nav.building') },
  { target: 'affairs-finance', label: t('building.nav.finance') },
  { target: 'affairs-owner-account', label: t('building.nav.ownerAccount') },
  { target: 'affairs-forms', label: t('building.nav.forms') },
  { target: 'affairs-feedback', label: t('building.nav.feedback'), needsApi: true },
  { target: 'affairs-access', label: t('building.nav.access') },
  { target: 'affairs-icctv', label: t('building.nav.icctv') },
]);

// 6. 最新通告資料
const rawNotices = computed<IsmartBuildingNotice[]>(() => ismartNoticeProfile.value?.result ?? []);
const memberBoundCommunityName = computed(() => memberCommunityName(sessionStore.me?.primary_community, locale.value));
const buildingNameMap = computed(() => buildBuildingNameMap(
  buildingDirectory.value,
  sessionStore.me?.primary_community,
  locale.value,
));
const resolveIndexedBuildingName = (buildingID: string): string =>
  indexedBuildingName(buildingNameMap.value, String(buildingID ?? '').trim());
const noticeBuildingOptions = computed<string[]>(() => {
  const options = [
    ...(ismartNoticeProfile.value?.building_options ?? []),
    ...(ismartBuildingProfile.value?.building_options ?? []),
    ...(ismartAccessProfile.value?.building_options ?? []),
    ...(icctvProfile.value?.building_options ?? []),
    selectedNoticeBuildingID.value,
    selectedBuildingID.value,
  ];
  return Array.from(new Set(options.map((item) => String(item ?? '').trim()).filter(Boolean)));
});

const noticeBuildingLabel = (buildingID: string): string => {
  const id = String(buildingID ?? '').trim();
  if (!id) return t('building.notices.noBuildingSelected');
  const name = resolveIndexedBuildingName(id)
    || (id === selectedBuildingID.value && buildingName.value !== '-' ? buildingName.value : '');
  if (name) {
    return t('building.notices.buildingWithId', { building: name, id });
  }
  return t('building.common.buildingCodeLabel', { id });
};

const currentNoticeBuildingText = computed(() => {
  const buildingID = selectedNoticeBuildingID.value || ismartNoticeProfile.value?.selected_building_id || selectedBuildingID.value;
  return buildingID ? noticeBuildingLabel(buildingID) : t('building.notices.noBuildingSelected');
});

const notices = computed<NoticeRow[]>(() =>
  rawNotices.value.map((item, index) => {
    const id = textValue(item.id);
    const code = textValue(item.mess_code);
    return {
      key: `${id}-${code}-${index}`,
      code,
      title: textValue(item.mess_title),
      type: textValue(item.mess_type),
      publishDate: formatLocaleDateValue(item.mess_date),
      expireDate: formatLocaleDateValue(item.mess_down),
      fileUrl: String(item.mess_file ?? '').trim(),
    };
  }),
);

const noticeEmptyText = computed(() => {
  if (noticeLoading.value) return t('building.notices.loadingEmpty');
  if (noticeError.value) return translateMessage(noticeError.value);
  if (!selectedNoticeBuildingID.value && noticeBuildingOptions.value.length === 0) {
    return t('building.notices.noAvailableBuilding');
  }
  return t('building.notices.empty');
});

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
    date: formatLocaleDateValue(item.file_date),
    month: formatLocaleMonthValue(item.file_month),
    url: String(item.file_url ?? '').trim(),
  }));

const parseReportDate = (value: string | null | undefined): number => {
  const raw = String(value ?? '').trim();
  if (!raw || raw === '-') return 0;
  const match = raw.match(/(19\d{2}|20\d{2})(?:[-/.年](0?[1-9]|1[0-2]))?(?:[-/.月](0?[1-9]|[12]\d|3[01]))?/);
  if (!match) return 0;
  const year = Number(match[1]);
  const month = match[2] ? Number(match[2]) : 12;
  const day = match[3] ? Number(match[3]) : new Date(Date.UTC(year, month, 0)).getUTCDate();
  return Date.UTC(year, month - 1, day);
};

const reportSortTime = (item: IsmartBuildingDocument): number =>
  parseReportDate(item.file_month)
  || parseReportDate(item.file_date)
  || parseReportDate(item.created_date)
  || parseReportDate(item.title);

const sortReportDocuments = (rows: IsmartBuildingDocument[] | undefined): IsmartBuildingDocument[] =>
  [...(rows ?? [])].sort((a, b) => {
    const timeDiff = reportSortTime(b) - reportSortTime(a);
    if (timeDiff !== 0) return timeDiff;
    return Number(b.id ?? 0) - Number(a.id ?? 0);
  });

const firstDocumentGroup = (...groups: Array<IsmartBuildingDocument[] | undefined>): IsmartBuildingDocument[] | undefined =>
  groups.find((group) => Array.isArray(group) && group.length > 0)
  ?? groups.find((group) => Array.isArray(group));

const documentRows = (...groups: Array<IsmartBuildingDocument[] | undefined>): BuildingFileRow[] => {
  return fileRows(firstDocumentGroup(...groups));
};

const reportDocumentRows = (...groups: Array<IsmartBuildingDocument[] | undefined>): BuildingFileRow[] =>
  fileRows(sortReportDocuments(firstDocumentGroup(...groups)));

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
const hasPendingResidenceBinding = computed(() => sessionStore.me?.residence_binding_status === 'pending');
const hasLinkedBuilding = computed(() => {
  const ismartMessage = sessionStore.me?.ismart_msg;
  const permissionBuildingIDs = [
    ...(ismartMessage?.building ?? []),
    ...(ismartMessage?.staff_building_permissions ?? []),
    ...(ismartMessage?.client_building_permissions ?? []),
  ];
  const profileBuildingIDs = [
    selectedBuildingID.value,
    selectedNoticeBuildingID.value,
    ismartBuildingProfile.value?.selected_building_id,
    ismartBuildingProfile.value?.building?.building_id,
    ismartNoticeProfile.value?.selected_building_id,
    ismartAccessProfile.value?.selected_building_id,
    icctvProfile.value?.selected_building_id,
  ];

  return sessionStore.me?.residence_binding_status === 'approved'
    || [...permissionBuildingIDs, ...profileBuildingIDs].some((buildingID) => String(buildingID ?? '').trim().length > 0);
});
const buildingName = computed(() => {
  const buildingID = String(currentBuilding.value.building_id || selectedBuildingID.value || '').trim();
  const upstreamName = usableBuildingName(
    locale.value === 'en'
      ? currentBuilding.value.buildname || currentBuilding.value.buildname_chi
      : currentBuilding.value.buildname_chi || currentBuilding.value.buildname,
    buildingID,
  );
  const profileName = preferredMemberBuildingID(sessionStore.me) === buildingID
    ? memberBoundCommunityName.value
    : '';
  return textValue(upstreamName || resolveIndexedBuildingName(buildingID) || profileName);
});
const organizationName = computed(() => textValue(currentBuildingInfo.value.owners_corporation_name));
const buildingForms = computed(() => fileRows(ismartBuildingProfile.value?.documents?.forms));
const buildingInfoFiles = computed(() => fileRows(ismartBuildingProfile.value?.documents?.building_info_files));
const floorPlans = computed(() => documentRows(
  ismartBuildingProfile.value?.documents?.floorplans,
  ismartBuildingProfile.value?.documents?.floorplan,
));
const financialReports = computed(() => reportDocumentRows(
  ismartBuildingProfile.value?.documents?.financial_reports,
  ismartBuildingProfile.value?.documents?.mfinreport,
));
const auditReports = computed(() => reportDocumentRows(
  ismartBuildingProfile.value?.documents?.audit_reports,
  ismartBuildingProfile.value?.documents?.auditreport,
  ismartBuildingProfile.value?.documents?.audition,
  ismartBuildingProfile.value?.documents?.audit_report,
  ismartBuildingProfile.value?.documents?.auditreports,
  ismartBuildingProfile.value?.documents?.auditions,
));
const buildingMapURL = computed(() => String(currentBuildingInfo.value.google_map_url ?? '').trim());
const buildingMapEmbedURL = computed(() => normalizeMapEmbedURL(currentBuildingInfo.value.google_map_url));
const buildingFileCount = computed(() => buildingForms.value.length + buildingInfoFiles.value.length + floorPlans.value.length);
const financeDocumentCount = computed(() => financialReports.value.length + auditReports.value.length);
const buildingStatusText = computed(() => {
  if (buildingInfoLoading.value) return t('building.profile.status.loading');
  if (buildingInfoError.value) return t('building.profile.status.loadFailed');
  return ismartBuildingProfile.value
    ? t('building.profile.status.hasData')
    : t('building.profile.status.noData');
});
const buildingStatusDetail = computed(() => {
  if (buildingInfoLoading.value) return t('building.profile.status.loadingDetail');
  if (buildingInfoError.value) return translateMessage(buildingInfoError.value);
  return ismartBuildingProfile.value
    ? t('building.profile.status.syncedDetail')
    : t('building.profile.status.noDataDetail');
});
const buildingStatusClass = computed(() => (buildingInfoError.value ? 'warn' : 'good'));

const buildingFields = computed<BuildingField[]>(() => [
  { label: t('building.profile.fields.yearBuilt'), value: textValue(currentBuildingInfo.value.year_built) },
  { label: t('building.profile.fields.totalFloors'), value: textValue(currentBuildingInfo.value.total_floor) },
  { label: t('building.profile.fields.totalUnits'), value: textValue(currentBuildingInfo.value.total_unit) },
  { label: t('building.profile.fields.totalCarparks'), value: textValue(currentBuildingInfo.value.total_carpark) },
  { label: t('building.profile.fields.ownersCorporation'), value: textValue(currentBuildingInfo.value.owners_corporation_name) },
  { label: t('building.profile.fields.managementOfficePhone'), value: textValue(currentBuildingInfo.value.management_office_phone) },
  { label: t('building.profile.fields.managementCompany'), value: textValue(currentBuildingInfo.value.management_company_name) },
  { label: t('building.profile.fields.managementCompanyPhone'), value: textValue(currentBuildingInfo.value.management_company_phone) },
  { label: t('building.profile.fields.managementCompanyEmail'), value: textValue(currentBuildingInfo.value.management_company_email) },
  { label: t('building.profile.fields.managementCompanyFax'), value: textValue(currentBuildingInfo.value.management_company_fax) },
  { label: t('building.profile.fields.homeAffairsPhone'), value: textValue(currentBuildingInfo.value.home_affairs_department_phone) },
  {
    label: t('building.profile.fields.dataSource'),
    value: ismartBuildingProfile.value ? t('building.profile.fields.dataSourceValue') : '-',
  },
]);

const managementOverviewFields = computed<BuildingField[]>(() => [
  { label: t('building.profile.fields.buildingName'), value: buildingName.value },
  { label: t('building.profile.fields.buildingId'), value: textValue(currentBuilding.value.building_id || selectedBuildingID.value) },
  { label: t('building.profile.fields.buildingType'), value: textValue(currentBuilding.value.building_type) },
  { label: t('building.profile.fields.area'), value: textValue(currentBuilding.value.area) },
  { label: t('building.profile.fields.district'), value: textValue(currentBuilding.value.district) },
  { label: t('building.profile.fields.street'), value: textValue(currentBuilding.value.street) },
  { label: t('building.profile.fields.streetNumber'), value: textValue(currentBuilding.value.street_no) },
  { label: t('building.profile.fields.block'), value: textValue(currentBuilding.value.block || currentBuilding.value.court) },
  { label: t('building.profile.fields.ownersCorporation'), value: textValue(currentBuildingInfo.value.owners_corporation_name) },
  { label: t('building.profile.fields.managementOfficePhone'), value: textValue(currentBuildingInfo.value.management_office_phone) },
  { label: t('building.profile.fields.managementCompany'), value: textValue(currentBuildingInfo.value.management_company_name) },
  { label: t('building.profile.fields.managementCompanyPhone'), value: textValue(currentBuildingInfo.value.management_company_phone) },
  { label: t('building.profile.fields.managementCompanyEmail'), value: textValue(currentBuildingInfo.value.management_company_email) },
  { label: t('building.profile.fields.managementCompanyFax'), value: textValue(currentBuildingInfo.value.management_company_fax) },
  { label: t('building.profile.fields.homeAffairsPhone'), value: textValue(currentBuildingInfo.value.home_affairs_department_phone) },
  {
    label: t('building.profile.fields.documentTotal'),
    value: t('building.common.fileCount', { count: formatLocaleNumber(financeDocumentCount.value) }),
  },
]);

const financeLoading = computed(() => buildingInfoLoading.value || financeReceivableLoading.value);
const financeReceivableStatusText = computed(() => {
  if (financeReceivableLoading.value) return t('building.finance.status.loading');
  if (financeReceivableError.value) return t('building.finance.status.partialFailure');
  return financeReceivableLoaded.value
    ? t('building.finance.status.synced')
    : t('building.finance.status.notLoaded');
});
const managementFeePreviewRows = computed(() => managementFeeRows.value.slice(0, 10));
const otherFeePreviewRows = computed(() => otherFeeRows.value.slice(0, 10));
const managementFeeColumns = computed(() => {
  const keys: string[] = [];
  managementFeePreviewRows.value.forEach((row) => {
    Object.keys(row).forEach((key) => {
      if (!keys.includes(key)) {
        keys.push(key);
      }
    });
  });

  const unitColumns = ['單位', 'flat_code', 'unit_id'].filter((key) => keys.includes(key));
  const otherColumns = keys.filter((key) => !unitColumns.includes(key));
  return [...unitColumns, ...otherColumns];
});
const managementFeeColumnSpan = computed(() => Math.max(managementFeeColumns.value.length, 1));
const financeColumnTranslationKeys: Record<string, string> = {
  '單位': 'building.finance.columns.unit',
  flat_code: 'building.finance.columns.unit',
  unit_id: 'building.finance.columns.unit',
  invoice_no: 'building.finance.columns.invoiceNumber',
  item_id: 'building.finance.columns.item',
  trs_to: 'building.finance.columns.term',
  trs_val: 'building.finance.columns.amount',
  remark: 'building.finance.columns.remark',
};
const financeColumnLabel = (column: string): string => {
  const translationKey = financeColumnTranslationKeys[column];
  return translationKey ? t(translationKey) : column;
};
const otherFeeTotal = computed(() =>
  otherFeeRows.value.reduce((sum, item) => sum + ownerAmountValue(item.trs_val), 0),
);
const financeReceivableFields = computed<BuildingField[]>(() => [
  {
    label: t('building.finance.fields.managementRows'),
    value: t('building.common.itemCount', { count: formatLocaleNumber(managementFeeRows.value.length) }),
  },
  {
    label: t('building.finance.fields.otherItems'),
    value: t('building.common.itemCount', { count: formatLocaleNumber(otherFeeRows.value.length) }),
  },
  { label: t('building.finance.fields.otherTotal'), value: formatOwnerHKD(otherFeeTotal.value) },
  { label: t('building.finance.fields.receivableStatus'), value: financeReceivableStatusText.value },
]);
const managementFeeEmptyText = computed(() => {
  if (financeReceivableLoading.value) return t('building.finance.managementLoading');
  if (financeReceivableError.value && managementFeeRows.value.length === 0) {
    return translateMessage(financeReceivableError.value);
  }
  return t('building.finance.managementEmpty');
});
const otherFeeEmptyText = computed(() => {
  if (financeReceivableLoading.value) return t('building.finance.otherLoading');
  if (financeReceivableError.value && otherFeeRows.value.length === 0) {
    return translateMessage(financeReceivableError.value);
  }
  return t('building.finance.otherEmpty');
});
const financeCellText = (value: unknown): string => {
  if (value === null || value === undefined) return '-';
  if (typeof value === 'boolean') return value ? t('building.common.yes') : t('building.common.no');
  const text = String(value).trim();
  return text || '-';
};
const financeRowKeyText = (values: unknown[]): string => {
  const key = values.map((item) => financeCellText(item)).join('|');
  return key.replace(/[\s|-]/g, '') ? key : JSON.stringify(values);
};
const managementFeeRowKey = (row: IsmartManagementFeeTableRow): string =>
  financeRowKeyText(managementFeeColumns.value.map((column) => row[column]));
const otherFeeRowKey = (row: IsmartOtherFeeRow): string =>
  financeRowKeyText([row.invoice_no, row.flat_code, row.item_id, row.trs_to, row.trs_val, row.remark]);

const buildingDocCards = computed<BuildingDocCard[]>(() => [
  {
    title: t('building.profile.documents.forms.title'),
    desc: t('building.profile.documents.forms.description'),
    count: t('building.common.fileCount', { count: formatLocaleNumber(buildingForms.value.length) }),
    empty: buildingForms.value.length > 0
      ? t('building.profile.documents.forms.available')
      : t('building.profile.documents.forms.empty'),
  },
  {
    title: t('building.profile.documents.information.title'),
    desc: t('building.profile.documents.information.description'),
    count: t('building.common.fileCount', { count: formatLocaleNumber(buildingInfoFiles.value.length) }),
    empty: buildingInfoFiles.value.length > 0
      ? t('building.profile.documents.information.available')
      : t('building.profile.documents.information.empty'),
  },
  {
    title: t('building.profile.documents.floorPlans.title'),
    desc: t('building.profile.documents.floorPlans.description'),
    count: t('building.common.fileCount', { count: formatLocaleNumber(floorPlans.value.length) }),
    empty: floorPlans.value.length > 0
      ? t('building.profile.documents.floorPlans.available')
      : t('building.profile.documents.floorPlans.empty'),
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
  if (accessLoading.value) return t('building.access.status.loading');
  if (accessError.value) return t('building.access.status.loadFailed');
  return ismartAccessProfile.value
    ? t('building.access.status.connected')
    : t('building.access.status.notLoaded');
});
const accessStatusClass = computed(() => (accessError.value ? 'warn' : 'good'));

// 7.2 取得門禁顯示值
const accessDoorID = (door: IsmartAccessDoor): string => String(door.door_id ?? '').trim();
const accessDoorNumber = (door: IsmartAccessDoor): number => Number(door.door_id ?? 0);
const accessQRCodeRecordNumber = (door: IsmartAccessDoor): number => Number(door.qrcode?.record_id ?? 0);
const accessDoorTitle = (door: IsmartAccessDoor): string => {
  const doorID = accessDoorID(door);
  return textValue(door.title || (doorID
    ? t('building.access.defaultDoorName', { id: doorID })
    : t('building.access.unnamedDoor')));
};
const accessDoorPasswordVisible = (door: IsmartAccessDoor): boolean => visiblePasswordDoorIDs.value.includes(accessDoorID(door));
const accessDoorPasswordText = (door: IsmartAccessDoor): string => {
  if (!door.password?.value) return '-';
  return accessDoorPasswordVisible(door) ? door.password.value : '******';
};
const accessTimeRange = (start: string | undefined, end: string | undefined): string => {
  const startText = formatLocaleDateValue(start);
  const endText = formatLocaleDateValue(end);
  if (startText === '-' && endText === '-') return '-';
  return t('building.access.timeRange', { start: startText, end: endText });
};
const accessOpenTypeText = (value: string | undefined): string => {
  const text = String(value ?? '').trim();
  if (text === 'remote') return t('building.access.remoteOpen');
  if (text === 'qrcode') return t('building.access.qrCode');
  return text || '-';
};
const accessTermText = (value: string | undefined): string => {
  const text = String(value ?? '').trim();
  if (text === 'dynamic') return t('building.access.dynamicQr');
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
        openTime: formatLocaleDateValue(record.open_time),
        openType: accessOpenTypeText(record.open_type),
        status: success ? 'good' : 'warn',
        statusText: success ? t('building.access.success') : t('building.access.unsuccessful'),
      };
    }),
  ),
);

// 7.3 視像監控資料
const icctvCameras = computed<ICCTVCameraSummary[]>(() => icctvProfile.value?.cameras ?? []);
const icctvOrangePis = computed(() => icctvProfile.value?.orangepis ?? []);
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
  if (icctvLoading.value) return t('building.icctv.statusText.loading');
  if (icctvError.value) return t('building.icctv.statusText.loadFailed');
  return icctvEnabled.value
    ? t('building.icctv.statusText.enabled')
    : t('building.icctv.statusText.disabled');
});
const icctvCameraName = (camera: ICCTVCameraSummary, index: number): string => {
  const match = String(camera.channel ?? '').match(/^channel(\d+)$/i);
  return t('building.icctv.cameraName', { number: match?.[1] ?? formatLocaleNumber(index + 1) });
};
const icctvCameraStatusText = (camera: ICCTVCameraSummary): string => (camera.is_active && camera.url
  ? t('building.icctv.available')
  : t('building.icctv.unavailable'));
const icctvCameraFrameTitle = (camera: ICCTVCameraSummary, index: number): string =>
  t('building.icctv.cameraFrameTitle', { camera: icctvCameraName(camera, index) });
const isICCTVCameraExpanded = (cameraID: string): boolean => expandedICCTVCameraIDs.value.includes(cameraID);

// 8. 大廈財務子面板
const switchFinanceSub = (target: FinanceSubTab) => {
  financeSubTab.value = target;
};

// 8.1 業戶帳目資料
const normalizeOwnerTextList = (values: string[] | undefined): string[] => {
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

const ownerDigitsOnly = (value: unknown): string =>
  String(value ?? '').replace(/\D/g, '');

const normalizeOwnerCompareValue = (value: unknown): string => {
  const raw = String(value ?? '').trim();
  if (!raw) return '';
  const digits = ownerDigitsOnly(raw);
  if (digits && digits.length === raw.replace(/\s/g, '').length) {
    return digits.replace(/^0+/, '') || '0';
  }
  return raw.toUpperCase();
};

const ownerTextValue = (value: unknown): string => {
  if (value === null || value === undefined) return '-';
  const text = String(value).trim();
  return text || '-';
};

const ownerAmountValue = (value: unknown): number => {
  const numeric = Number(String(value ?? '').replace(/,/g, ''));
  return Number.isFinite(numeric) ? numeric : 0;
};

const formatOwnerHKD = (value: number): string =>
  new Intl.NumberFormat(displayLocale.value, {
    style: 'currency',
    currency: 'HKD',
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(value);

const readOwnerPaymentText = (row: Partial<Record<string, unknown>>, keys: string[]): string => {
  for (const key of keys) {
    const value = row[key];
    if (value !== undefined && value !== null && String(value).trim() !== '') {
      return String(value).trim();
    }
  }
  return '';
};

const ownerBoundUnitIDs = computed(() => {
  const savedUnits = normalizeOwnerTextList(sessionStore.me?.bound_flat_unit_ids);
  if (savedUnits.length > 0) {
    return savedUnits;
  }
  return normalizeOwnerTextList(sessionStore.me?.ismart_msg?.client_building_flat_units_permissions);
});
const ownerUnitContext = computed<OwnerUnitContext>(() => {
  const unitID = ownerBoundUnitIDs.value[0] ?? '';
  const unitDigits = ownerDigitsOnly(unitID);
  const primaryBuildingID = String(sessionStore.me?.primary_community?.public_id ?? '').trim();
  const buildingID = unitDigits.length >= 7 ? unitDigits.slice(0, 7) : primaryBuildingID;
  const buildingNameText = memberBoundCommunityName.value || buildingName.value;
  const fallbackFloor = unitDigits.length >= 9 ? unitDigits.slice(7, 9) : '';
  const fallbackUnit = unitDigits.length >= 11 ? unitDigits.slice(9, 11).replace(/^0+/, '') : '';

  return {
    buildingID,
    buildingName: buildingNameText && buildingNameText !== '-' ? buildingNameText : buildingID,
    floor: String(sessionStore.me?.residence_floor ?? '').trim() || fallbackFloor,
    unit: String(sessionStore.me?.residence_unit ?? '').trim() || fallbackUnit,
    unitID,
  };
});
const ownerHasBoundUnit = computed(() => ownerBoundUnitIDs.value.length > 0 && ownerUnitContext.value.unitID !== '');
const ownerBoundUnitLabel = computed(() => {
  if (!ownerHasBoundUnit.value) {
    return t('building.ownerAccount.unboundUnit');
  }

  const context = ownerUnitContext.value;
  return [
    context.buildingName,
    context.floor,
    context.unit,
  ].filter(Boolean).join(' / ');
});
const ownerAccountLoading = computed(() => ownerUnpaidLoading.value || ownerRecordsLoading.value);
const ownerUnpaidTotal = computed(() =>
  ownerUnpaidInvoices.value.reduce((sum, item) => sum + ownerAmountValue(item.net_amount), 0),
);
const ownerUnpaidEmptyText = computed(() => {
  if (ownerUnpaidLoading.value) return t('building.ownerAccount.unpaidLoading');
  if (ownerUnpaidError.value) return translateMessage(ownerUnpaidError.value);
  if (!ownerHasBoundUnit.value) return t('building.ownerAccount.saveBoundUnitFirst');
  return t('building.ownerAccount.unpaidEmpty');
});

const ownerUnpaidInvoiceKey = (invoice: POSIntegrationUnpaidInvoice, index: number): string =>
  [
    invoice.invoice_no,
    invoice.flat_code,
    invoice.item_id,
    invoice.trs_to,
    invoice.bill_dt,
    index,
  ].map((item) => ownerTextValue(item)).join('|');

const ownerPaymentStatusText = (value: string): string => {
  const status = value.trim();
  if (!status) return '-';
  if (status === 'confirmed' || status === 'validated_by_ismart' || status === 'payment_captured') {
    return t('building.ownerAccount.statuses.confirmed');
  }
  if (status === 'in_cashier') return t('building.ownerAccount.statuses.inCashier');
  if (status === 'pending_validation') return t('building.ownerAccount.statuses.pendingValidation');
  if (status === 'init') return t('building.ownerAccount.statuses.init');
  return status;
};

const ownerPaymentStatusClass = (value: string): string => {
  const status = value.trim();
  if (status === 'confirmed' || status === 'validated_by_ismart' || status === 'payment_captured') return 'acct-paid';
  if (status === 'in_cashier' || status === 'pending_validation' || status === 'init') return 'acct-pending';
  return '';
};

const ownerPaymentDetailBelongsToBoundUnit = (detail: POSIntegrationPaymentDetail): boolean => {
  const context = ownerUnitContext.value;
  const detailUnitID = readOwnerPaymentText(detail, ['flat_code', 'unit_id', 'unitID']);
  if (detailUnitID && context.unitID) {
    return ownerDigitsOnly(detailUnitID) === ownerDigitsOnly(context.unitID);
  }

  const detailFloor = readOwnerPaymentText(detail, ['floor']);
  const detailUnit = readOwnerPaymentText(detail, ['unit', 'unit_name']);
  return (
    normalizeOwnerCompareValue(detailFloor) === normalizeOwnerCompareValue(context.floor) &&
    normalizeOwnerCompareValue(detailUnit) === normalizeOwnerCompareValue(context.unit)
  );
};

const filterOwnerTransactionsForBoundUnit = (
  records: POSIntegrationPaymentTransaction[],
): POSIntegrationPaymentTransaction[] =>
  records.flatMap((record) => {
    const details = record.payment_detail_objs ?? [];
    if (details.length === 0) {
      return [];
    }

    const matchedDetails = details.filter(ownerPaymentDetailBelongsToBoundUnit);
    return matchedDetails.length > 0
      ? [{ ...record, payment_detail_objs: matchedDetails }]
      : [];
  });

const ownerPaymentRecordRows = computed<OwnerPaymentRecordRow[]>(() =>
  ownerPaymentRecords.value.flatMap((record, recordIndex) => {
    const details = record.payment_detail_objs && record.payment_detail_objs.length > 0
      ? record.payment_detail_objs
      : [undefined];

    return details.map((detail, detailIndex) => {
      const detailRow = detail ?? {};
      const floor = readOwnerPaymentText(detailRow, ['floor']);
      const unit = readOwnerPaymentText(detailRow, ['unit', 'unit_name']);
      const paymentID = ownerTextValue(record.payment_id);
      const receiptID = ownerTextValue(record.receipt_id);
      return {
        key: `${paymentID}-${receiptID}-${recordIndex}-${detailIndex}`,
        receiptID,
        paymentID,
        inputTime: formatLocaleDateValue(record.input_time),
        tranTime: formatLocaleDateValue(record.tran_time),
        unit: [floor, unit].filter(Boolean).join(' / ') || ownerBoundUnitLabel.value,
        item: ownerTextValue(readOwnerPaymentText(detailRow, ['item_id', 'item_name', 'name'])),
        term: ownerTextValue(readOwnerPaymentText(detailRow, ['term', 'trs_to', 'period'])),
        amount: ownerAmountValue(detail?.trs_val ?? record.trs_val),
        payType: ownerTextValue(record.pay_type),
        status: ownerPaymentStatusText(String(record.status ?? '')),
        statusClass: ownerPaymentStatusClass(String(record.status ?? '')),
        remark: ownerTextValue(readOwnerPaymentText(detailRow, ['remark'])),
      };
    });
  }),
);
const ownerPaymentRecordTotal = computed(() =>
  ownerPaymentRecordRows.value.reduce((sum, row) => sum + row.amount, 0),
);
const ownerRecordsEmptyText = computed(() => {
  if (ownerRecordsLoading.value) return t('building.ownerAccount.recordsLoading');
  if (ownerRecordsError.value) return translateMessage(ownerRecordsError.value);
  if (!ownerHasBoundUnit.value) return t('building.ownerAccount.saveBoundUnitFirst');
  if (ownerRecordsDateSearched.value) return t('building.ownerAccount.recordsDateEmpty');
  return t('building.ownerAccount.recordsEmpty');
});

// 9. 申請表格 mock 資料
const selectedFormOrg = computed(() => organizationName.value);
const selectedFormBuilding = computed(() => buildingName.value);
const forms = computed(() => buildingForms.value);

// 10. 意見提供示例資料
const feedbackBuildingKeys = [
  'yanYing', 'harmony', 'wahHing269', 'wahHing271', 'rich', 'luck', 'tinFu',
  'newManLee', 'manKoB', 'siOn', 'tungSing', 'wahYuen', 'wingSumIndustrial2',
  'hongMukGarden2', 'namCheongCourt', 'tungNam77', 'tungNam75', 'test1', 'leeLoi',
  'yanMan', 'wingSingCourt', 'yanLee', 'laiLai', 'jadeCinnamon1', 'jadeCinnamon2',
  'jadeCinnamon3', 'jadeCinnamon4', 'jadeCinnamon5', 'jadeCinnamon6', 'jadeCinnamon7',
  'jadeCinnamon8', 'jadeCinnamon9', 'jadeCinnamon10', 'jadeCinnamon11',
];
const feedbackBuildings = computed<SelectOption[]>(() => feedbackBuildingKeys.map((value) => ({
  value,
  label: t(`building.feedback.buildings.${value}`),
})));
const selectedFeedbackBuildingLabel = computed(() =>
  feedbackBuildings.value.find((item) => item.value === selectedFeedbackBuilding.value)?.label
  || t('building.common.emptySelection'));

const repairCategoryMap: Record<string, string[]> = {
  electrical: ['corridorLighting', 'lobbyLighting', 'switchFault', 'otherLighting'],
  structure: ['doorLock', 'window', 'wall', 'ceiling', 'otherStructure'],
  environment: ['cleaning', 'standingWater', 'pests', 'otherHygiene'],
  lift: ['fault', 'cleaning', 'other'],
  water: ['freshWater', 'flushingWater', 'leakage', 'other'],
  other: ['other'],
};

const feedbackCategoryMap: Record<string, string[]> = {
  environment: ['cleaningSuggestion', 'environmentImprovement', 'otherHygiene'],
  facilities: ['suggestion', 'damage', 'other'],
  management: ['serviceSuggestion', 'staffPerformance', 'other'],
  platform: ['feature', 'issue', 'other'],
  other: ['other'],
};

const repairCategoryLabel = (category: string): string =>
  category ? t(`building.feedback.categories.repair.${category}.label`) : '';
const repairSubcategoryLabel = (category: string, subcategory: string): string =>
  category && subcategory ? t(`building.feedback.categories.repair.${category}.${subcategory}`) : '';
const feedbackCategoryLabel = (category: string): string =>
  category ? t(`building.feedback.categories.feedback.${category}.label`) : '';
const feedbackSubcategoryLabel = (category: string, subcategory: string): string =>
  category && subcategory ? t(`building.feedback.categories.feedback.${category}.${subcategory}`) : '';

const repairCategories = computed<SelectOption[]>(() => Object.keys(repairCategoryMap).map((value) => ({
  value,
  label: repairCategoryLabel(value),
})));
const feedbackCategories = computed<SelectOption[]>(() => Object.keys(feedbackCategoryMap).map((value) => ({
  value,
  label: feedbackCategoryLabel(value),
})));
const repairSubcategories = computed<SelectOption[]>(() => {
  if (!repairCategory.value) return [];
  return (repairCategoryMap[repairCategory.value] ?? []).map((value) => ({
    value,
    label: repairSubcategoryLabel(repairCategory.value, value),
  }));
});
const feedbackSubcategories = computed<SelectOption[]>(() => {
  if (!feedbackCategory.value) return [];
  return (feedbackCategoryMap[feedbackCategory.value] ?? []).map((value) => ({
    value,
    label: feedbackSubcategoryLabel(feedbackCategory.value, value),
  }));
});

const feedbackRecords = computed<FeedbackRecord[]>(() => [
  {
    type: t('building.feedback.mode.repair'),
    building: t('building.feedback.buildings.harmony'),
    subject: t('building.feedback.records.corridorLighting.subject'),
    category: repairCategoryLabel('electrical'),
    status: 'warn',
    statusText: t('building.feedback.records.processing'),
    updatedAt: t('building.feedback.records.todayAt', { time: '10:20' }),
    content: t('building.feedback.records.corridorLighting.content'),
  },
  {
    type: t('building.feedback.mode.repair'),
    building: t('building.feedback.buildings.harmony'),
    subject: t('building.feedback.records.entranceLock.subject'),
    category: repairCategoryLabel('structure'),
    status: '',
    statusText: t('building.feedback.records.pending'),
    updatedAt: t('building.feedback.records.yesterdayAt', { time: '09:30' }),
    content: t('building.feedback.records.entranceLock.content'),
  },
  {
    type: t('building.feedback.mode.feedback'),
    building: t('building.feedback.buildings.harmony'),
    subject: t('building.feedback.records.lobbyCleaning.subject'),
    category: feedbackCategoryLabel('environment'),
    status: 'good',
    statusText: t('building.feedback.records.completed'),
    updatedAt: t('building.feedback.records.yesterdayAt', { time: '16:45' }),
    content: t('building.feedback.records.lobbyCleaning.content'),
  },
  {
    type: t('building.feedback.mode.feedback'),
    building: t('building.feedback.buildings.siOn'),
    subject: t('building.feedback.records.platformFeature.subject'),
    category: feedbackCategoryLabel('platform'),
    status: 'warn',
    statusText: t('building.feedback.records.processing'),
    updatedAt: formatLocaleDateValue('2026-06-04'),
    content: t('building.feedback.records.platformFeature.content'),
  },
]);

// 12. 讀取目前會員綁定大廈資料
const loadBuildingInfo = async (buildingID = selectedBuildingID.value) => {
  buildingInfoLoading.value = true;
  buildingInfoError.value = '';
  try {
    const result = await fetchMemberIsmartBuildingInfo(buildingID || undefined);
    ismartBuildingProfile.value = result;
    selectedBuildingID.value = result.selected_building_id || result.building?.building_id || buildingID || '';
    if (!selectedNoticeBuildingID.value) {
      selectedNoticeBuildingID.value = selectedBuildingID.value;
    }
  } catch (error) {
    console.error(error);
    buildingInfoError.value = 'building.profile.loadError';
    ismartBuildingProfile.value = null;
  } finally {
    buildingInfoLoading.value = false;
  }
};

// 12.0 讀取大廈名稱目錄
const loadBuildingDirectory = async (): Promise<void> => {
  try {
    const memberBuildings = await fetchMemberPosBuildings();
    if (!memberBuildings.some((item) => !posBuildingName(item, locale.value))) {
      buildingDirectory.value = memberBuildings;
      return;
    }
    try {
      const publicBuildings = await fetchPosBuildings();
      const publicBuildingMap = new Map(publicBuildings.map((item) => [posBuildingID(item), item]));
      buildingDirectory.value = memberBuildings.map((item) =>
        posBuildingName(item, locale.value) ? item : publicBuildingMap.get(posBuildingID(item)) ?? item,
      );
    } catch (publicError) {
      console.error(publicError);
      buildingDirectory.value = memberBuildings;
    }
  } catch {
    try {
      buildingDirectory.value = await fetchPosBuildings();
    } catch (publicError) {
      console.error(publicError);
      buildingDirectory.value = [];
    }
  }
};

// 12.1 讀取目前會員大廈應收資料
const loadBuildingFinanceReceivables = async (buildingID = selectedBuildingID.value) => {
  financeReceivableLoading.value = true;
  financeReceivableError.value = '';
  try {
    const targetBuildingID = String(buildingID || selectedBuildingID.value).trim();
    if (!targetBuildingID) {
      managementFeeRows.value = [];
      otherFeeRows.value = [];
      financeReceivableLoaded.value = true;
      financeReceivableError.value = 'building.finance.missingBuildingId';
      return;
    }

    const [managementResult, otherResult] = await Promise.allSettled([
      fetchMemberIsmartManagementFees(targetBuildingID, 'table'),
      fetchMemberIsmartOtherFees(targetBuildingID, 'list'),
    ]);

    if (managementResult.status === 'fulfilled') {
      managementFeeRows.value = managementResult.value.result ?? [];
    } else {
      console.error(managementResult.reason);
      managementFeeRows.value = [];
    }
    if (otherResult.status === 'fulfilled') {
      otherFeeRows.value = otherResult.value.result ?? [];
    } else {
      console.error(otherResult.reason);
      otherFeeRows.value = [];
    }

    financeReceivableLoaded.value = true;
    if (managementResult.status === 'rejected' && otherResult.status === 'rejected') {
      financeReceivableError.value = 'building.finance.receivableLoadError';
    } else if (managementResult.status === 'rejected' || otherResult.status === 'rejected') {
      financeReceivableError.value = 'building.finance.partialReceivableLoadError';
    }
  } finally {
    financeReceivableLoading.value = false;
  }
};

// 12.2 讀取大廈財務整合資料
const loadBuildingFinance = async () => {
  await loadBuildingInfo();
  await loadBuildingFinanceReceivables(selectedBuildingID.value);
};

// 12.3 讀取目前會員大廈有效通告
const loadBuildingNotices = async (buildingID = selectedNoticeBuildingID.value || selectedBuildingID.value) => {
  noticeLoading.value = true;
  noticeError.value = '';
  try {
    const result = await fetchMemberIsmartBuildingNotices(buildingID || undefined);
    ismartNoticeProfile.value = result;
    selectedNoticeBuildingID.value = result.selected_building_id || buildingID || selectedBuildingID.value || '';
    if (!selectedBuildingID.value) {
      selectedBuildingID.value = selectedNoticeBuildingID.value;
    }
  } catch (error) {
    console.error(error);
    noticeError.value = 'building.notices.loadError';
    ismartNoticeProfile.value = null;
  } finally {
    noticeLoading.value = false;
  }
};

// 13. 讀取目前會員綁定單位未繳賬單
const loadOwnerUnpaidInvoices = async () => {
  ownerUnpaidLoading.value = true;
  ownerUnpaidError.value = '';
  try {
    if (!ownerHasBoundUnit.value) {
      ownerUnpaidInvoices.value = [];
      ownerUnpaidLoaded.value = true;
      return;
    }

    const result = await Promise.all(
      ownerBoundUnitIDs.value.map((unitID) => fetchPOSIntegrationUnpaidInvoices(unitID)),
    );
    ownerUnpaidInvoices.value = result.flat();
    ownerUnpaidLoaded.value = true;
  } catch (error) {
    console.error(error);
    ownerUnpaidInvoices.value = [];
    ownerUnpaidLoaded.value = true;
    ownerUnpaidError.value = 'building.ownerAccount.unpaidLoadError';
  } finally {
    ownerUnpaidLoading.value = false;
  }
};

// 13.1 讀取目前會員綁定單位繳費記錄
const loadOwnerPaymentRecords = async () => {
  ownerRecordsLoading.value = true;
  ownerRecordsError.value = '';
  ownerRecordsDateSearched.value = false;
  try {
    if (!ownerHasBoundUnit.value) {
      ownerPaymentRecords.value = [];
      ownerRecordsLoaded.value = true;
      return;
    }

    const result = await fetchPOSIntegrationTransactionsByUnit(ownerBoundUnitIDs.value);
    ownerPaymentRecords.value = result.payment_objs ?? [];
    ownerRecordsLoaded.value = true;
  } catch (error) {
    console.error(error);
    ownerPaymentRecords.value = [];
    ownerRecordsLoaded.value = true;
    ownerRecordsError.value = 'building.ownerAccount.recordsLoadError';
  } finally {
    ownerRecordsLoading.value = false;
  }
};

// 13.2 按日期查詢目前會員綁定單位繳費記錄
const searchOwnerPaymentRecordsByDate = async () => {
  ownerRecordsError.value = '';
  if (!ownerHasBoundUnit.value) {
    ownerPaymentRecords.value = [];
    ownerRecordsLoaded.value = true;
    ownerRecordsError.value = 'building.ownerAccount.saveBoundUnitFirst';
    return;
  }
  if (!ownerRecordFromDate.value || !ownerRecordToDate.value) {
    ownerRecordsError.value = 'building.ownerAccount.selectDateRange';
    return;
  }
  if (ownerRecordFromDate.value > ownerRecordToDate.value) {
    ownerRecordsError.value = 'building.ownerAccount.invalidDateRange';
    return;
  }
  if (!ownerUnitContext.value.buildingID) {
    ownerRecordsError.value = 'building.ownerAccount.missingBuildingId';
    return;
  }

  ownerRecordsLoading.value = true;
  try {
    const result = await fetchPOSIntegrationTransactionsByDate({
      building_id: ownerUnitContext.value.buildingID,
      from_date: ownerRecordFromDate.value,
      to_date: ownerRecordToDate.value,
      date_type: ownerRecordDateType.value,
      pay_method: 'all',
    }, ownerBoundUnitIDs.value);
    ownerPaymentRecords.value = filterOwnerTransactionsForBoundUnit(result.payment_objs ?? []);
    ownerRecordsDateSearched.value = true;
    ownerRecordsLoaded.value = true;
  } catch (error) {
    console.error(error);
    ownerPaymentRecords.value = [];
    ownerRecordsLoaded.value = true;
    ownerRecordsError.value = 'building.ownerAccount.dateSearchError';
  } finally {
    ownerRecordsLoading.value = false;
  }
};

// 13.3 讀取業戶帳目資料
const loadOwnerAccount = async () => {
  await Promise.all([
    loadOwnerUnpaidInvoices(),
    loadOwnerPaymentRecords(),
  ]);
};

// 13.4 切換業戶帳目子面板
const switchOwnerAccountTab = (target: OwnerAccountTab) => {
  ownerAccountTab.value = target;
  if (target === 'owner-unpaid' && !ownerUnpaidLoaded.value && !ownerUnpaidLoading.value) {
    void loadOwnerUnpaidInvoices();
  }
  if (target === 'owner-records' && !ownerRecordsLoaded.value && !ownerRecordsLoading.value) {
    void loadOwnerPaymentRecords();
  }
};

// 13.5 重新整理目前業戶帳目子面板
const refreshOwnerAccount = () => {
  if (ownerAccountTab.value === 'owner-unpaid') {
    void loadOwnerUnpaidInvoices();
    return;
  }
  if (ownerRecordsDateSearched.value) {
    void searchOwnerPaymentRecordsByDate();
    return;
  }
  void loadOwnerPaymentRecords();
};

// 13.6 顯示全部繳費記錄
const resetOwnerPaymentRecordSearch = () => {
  ownerRecordFromDate.value = currentMonthStart();
  ownerRecordToDate.value = currentDateValue();
  void loadOwnerPaymentRecords();
};

// 14. 讀取目前會員智能門禁資料
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
    accessError.value = 'building.access.loadError';
  } finally {
    accessLoading.value = false;
  }
};

// 15. 讀取目前會員視像監控資料
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
    icctvError.value = 'building.icctv.loadError';
    icctvProfile.value = null;
    expandedICCTVCameraIDs.value = [];
  } finally {
    icctvLoading.value = false;
  }
};

// 16. 切換主面板
const switchTab = (target: AffairsTab) => {
  activeTab.value = target;
  if (target === 'affairs-notices' && !ismartNoticeProfile.value && !noticeLoading.value) {
    void loadBuildingNotices();
  }
  if (
    target === 'affairs-finance'
    && (!ismartBuildingProfile.value || buildingInfoError.value || !financeReceivableLoaded.value || financeReceivableError.value)
    && !financeReceivableLoading.value
  ) {
    void loadBuildingFinance();
  }
  if (target === 'affairs-owner-account' && (!ownerUnpaidLoaded.value || !ownerRecordsLoaded.value) && !ownerAccountLoading.value) {
    void loadOwnerAccount();
  }
  if (target === 'affairs-access' && !ismartAccessProfile.value && !accessLoading.value) {
    void loadBuildingAccess();
  }
  if (target === 'affairs-icctv' && !icctvProfile.value && !icctvLoading.value) {
    void loadICCTV();
  }
};

// 17. 取得門禁操作用大廈 ID
const accessPayloadBuildingID = (door?: IsmartAccessDoor): string | undefined => {
  const buildingID = selectedBuildingID.value
    || ismartAccessProfile.value?.selected_building_id
    || door?.building_id
    || '';
  return buildingID || undefined;
};

// 18. 切換門禁密碼可見狀態
const toggleAccessPassword = (door: IsmartAccessDoor) => {
  const doorID = accessDoorID(door);
  if (!doorID) return;
  visiblePasswordDoorIDs.value = visiblePasswordDoorIDs.value.includes(doorID)
    ? visiblePasswordDoorIDs.value.filter((id) => id !== doorID)
    : [...visiblePasswordDoorIDs.value, doorID];
};

// 19. 發送開門指令
const openAccessDoor = async (door: IsmartAccessDoor) => {
  const doorID = accessDoorID(door);
  const doorNumber = accessDoorNumber(door);
  if (!door.has_permission || doorNumber <= 0) return;
  if (!window.confirm(t('building.access.confirmOpen', { door: accessDoorTitle(door) }))) return;

  openingDoorID.value = doorID;
  accessMessage.value = '';
  accessError.value = '';
  try {
    const result = await openMemberIsmartDoor({
      building_id: accessPayloadBuildingID(door),
      door_id: doorNumber,
    });
    const resultMessage = result.is_success === false
      ? 'building.access.openFailed'
      : 'building.access.openSent';
    await loadBuildingAccess(false);
    accessMessage.value = resultMessage;
  } catch (error) {
    console.error(error);
    accessError.value = 'building.access.openRequestError';
  } finally {
    openingDoorID.value = '';
  }
};

// 20. 生成門禁二維碼
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
      accessError.value = 'building.access.qrEmptyError';
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
    accessError.value = 'building.access.qrGenerateError';
  } finally {
    qrLoadingDoorID.value = '';
  }
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
const reviewMode = computed(() => (affairsMode.value === 'repair'
  ? t('building.feedback.mode.repair')
  : t('building.feedback.mode.feedback')));
const reviewCategory = computed(() =>
  affairsMode.value === 'repair'
    ? repairCategoryLabel(repairCategory.value) || t('building.feedback.review.selectCategory')
    : feedbackCategoryLabel(feedbackCategory.value) || t('building.feedback.review.selectCategory'),
);
const reviewSubcategory = computed(() =>
  affairsMode.value === 'repair'
    ? repairSubcategoryLabel(repairCategory.value, repairSubcategory.value)
      || t('building.feedback.review.selectSubcategory')
    : feedbackSubcategoryLabel(feedbackCategory.value, feedbackSubcategory.value)
      || t('building.feedback.review.selectSubcategory'),
);
const reviewContent = computed(() =>
  affairsMode.value === 'repair'
    ? repairContent.value || t('building.feedback.review.enterContent')
    : feedbackContent.value || t('building.feedback.review.enterContent'),
);
const reviewMedia = computed(() => mediaFileName.value || t('building.feedback.review.noFile'));

// 28. 重新整理通告
const refreshNotices = () => {
  void loadBuildingNotices();
};

// 28.1 切換通告大廈
const handleNoticeBuildingChange = () => {
  void loadBuildingNotices(selectedNoticeBuildingID.value);
};

// 29. 展開或收起視像監控鏡頭
const toggleICCTVCamera = (cameraID: string) => {
  expandedICCTVCameraIDs.value = isICCTVCameraExpanded(cameraID)
    ? expandedICCTVCameraIDs.value.filter((id) => id !== cameraID)
    : [...expandedICCTVCameraIDs.value, cameraID];
};

// 30. 開啟視像監控新窗口
const openIcctvWindow = (camera: ICCTVCameraSummary) => {
  if (!camera.is_active || !camera.url) return;
  window.open(camera.url, '_blank', 'noopener');
};

// 32. 開啟通告文件
const goNoticeDetail = (notice: NoticeRow) => {
  if (!notice.fileUrl) return;
  window.open(notice.fileUrl, '_blank', 'noopener');
};

// 31. 以最新會員綁定初始化我的大廈
const initializeBuildingPage = async (): Promise<void> => {
  try {
    await sessionStore.loadCurrentUser();
  } catch (error) {
    console.error(error);
  }

  const buildingID = preferredMemberBuildingID(sessionStore.me);
  selectedBuildingID.value = buildingID;
  selectedNoticeBuildingID.value = buildingID;
  await loadBuildingDirectory();
  await Promise.all([
    loadBuildingInfo(buildingID),
    loadBuildingNotices(buildingID),
  ]);
};

onMounted(() => {
  void initializeBuildingPage();
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
        <h1>{{ t('building.pageTitle') }}</h1>
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
            <span v-if="item.needsApi" class="work-nav-note">{{ t('building.nav.needsApi') }}</span>
          </button>
        </nav>
      </aside>

      <!-- 右側主內容 -->
      <main class="work-main">
        <section
          v-if="hasPendingResidenceBinding"
          class="work-card building-binding-empty"
        >
          <div>
            <div class="work-card-title">{{ t('building.binding.pendingTitle') }}</div>
            <p class="work-card-sub">{{ t('building.binding.pendingDescription') }}</p>
          </div>
          <RouterLink
            class="work-action"
            to="/account/profile?panel=property-binding"
          >
            {{ t('building.binding.pendingAction') }}
          </RouterLink>
        </section>
        <section
          v-else-if="!hasLinkedBuilding && !buildingInfoLoading && !noticeLoading"
          class="work-card building-binding-empty"
        >
          <div>
            <div class="work-card-title">{{ t('building.binding.title') }}</div>
            <p class="work-card-sub">{{ t('building.binding.description') }}</p>
          </div>
          <RouterLink
            class="work-action"
            to="/account/profile?panel=property-binding"
          >
            {{ t('building.binding.action') }}
          </RouterLink>
        </section>

        <!-- 最新通告 -->
        <div
          v-show="activeTab === 'affairs-notices'"
          class="work-panel"
          :class="{ on: activeTab === 'affairs-notices' }"
          data-work-panel="affairs-notices"
        >
          <section class="work-hero">
            <div>
              <div class="work-kicker">{{ t('building.notices.kicker') }}</div>
              <h2 class="work-title">{{ t('building.notices.title') }}</h2>
            </div>
          </section>
          <section class="notice-admin-grid">
            <div class="notice-admin-card">
              <h3>{{ t('building.notices.buildingSelection') }}</h3>
              <label class="notice-admin-label">{{ t('building.notices.selectBuilding') }}</label>
              <select
                v-model="selectedNoticeBuildingID"
                class="notice-select"
                :disabled="noticeLoading || noticeBuildingOptions.length === 0"
                @change="handleNoticeBuildingChange"
              >
                <option
                  v-if="noticeBuildingOptions.length === 0"
                  value=""
                >
                  {{ t('building.notices.noBuildingOption') }}
                </option>
                <option
                  v-for="b in noticeBuildingOptions"
                  :key="b"
                  :value="b"
                >
                  {{ noticeBuildingLabel(b) }}
                </option>
              </select>
            </div>
          </section>
          <section class="work-card">
            <div class="notice-current">
              <span>{{ t('building.notices.currentDisplay', { building: currentNoticeBuildingText }) }}</span>
              <button
                type="button"
                class="notice-action-btn"
                :disabled="noticeLoading"
                @click="refreshNotices"
              >
                {{ noticeLoading ? t('building.common.loading') : t('building.common.refresh') }}
              </button>
            </div>
            <table class="work-table notice-table">
              <thead>
                <tr>
                  <th>{{ t('building.notices.code') }}</th>
                  <th>{{ t('building.common.title') }}</th>
                  <th>{{ t('building.common.type') }}</th>
                  <th>{{ t('building.notices.publishDate') }}</th>
                  <th>{{ t('building.notices.expireDate') }}</th>
                  <th>{{ t('building.common.action') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-if="noticeLoading || noticeError || notices.length === 0">
                  <td
                    class="building-empty-row"
                    colspan="6"
                  >
                    {{ noticeEmptyText }}
                  </td>
                </tr>
                <template v-else>
                  <tr
                    v-for="n in notices"
                    :key="n.key"
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
                          :disabled="!n.fileUrl"
                          @click="goNoticeDetail(n)"
                        >
                          {{ t('building.common.view') }}
                        </button>
                      </div>
                    </td>
                  </tr>
                </template>
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
              <div class="work-kicker">{{ t('building.profile.kicker') }}</div>
              <h2 class="work-title">{{ t('building.profile.title') }}</h2>
              <p class="work-desc">{{ t('building.profile.description') }}</p>
            </div>
            <button
              type="button"
              class="work-action"
              :disabled="buildingInfoLoading"
              @click="loadBuildingInfo()"
            >
              {{ t('building.common.refresh') }}
            </button>
          </section>
          <section class="building-summary-grid">
            <div class="work-card">
              <div class="work-card-title">{{ t('building.profile.organization') }}</div>
              <div class="work-card-sub">{{ organizationName }}</div>
            </div>
            <div class="work-card">
              <div class="work-card-title">{{ t('building.profile.currentBuilding') }}</div>
              <div class="work-stat-label">{{ buildingName }}</div>
              <div class="work-stat">{{ formatLocaleNumber(buildingFileCount) }}</div>
              <div class="work-stat-label">{{ t('building.profile.basicDocumentCount') }}</div>
            </div>
            <div class="work-card">
              <div class="work-card-title">{{ t('building.profile.dataStatus') }}</div>
              <div class="work-row">
                <div>
                  <strong>{{ buildingStatusText }}</strong>
                  <span>{{ buildingStatusDetail }}</span>
                </div>
                <span
                  class="work-chip"
                  :class="buildingStatusClass"
                >
                  {{ buildingInfoLoading
                    ? t('building.common.loading')
                    : buildingInfoError
                      ? t('building.common.abnormal')
                      : t('building.common.normal') }}
                </span>
              </div>
            </div>
            <div class="work-card">
              <div class="work-card-title">{{ t('building.profile.building') }}</div>
              <div class="work-card-sub">{{ buildingName }}</div>
            </div>
          </section>
          <section class="work-card">
            <div class="work-card-title">{{ t('building.profile.basicFields') }}</div>
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
            <div class="work-card-title">{{ t('building.profile.mapUrl') }}</div>
            <iframe
              v-if="buildingMapEmbedURL"
              class="building-map-frame"
              :src="buildingMapEmbedURL"
              :title="t('building.profile.mapFrameTitle', { building: buildingName })"
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
                {{ t('building.profile.viewMap') }}
              </a>
            </div>
            <div
              v-else
              class="building-empty-row"
            >
              {{ t('building.profile.noMap') }}
            </div>
          </section>
          <section class="work-card">
            <div class="work-card-title">{{ t('building.profile.basicDocuments') }}</div>
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
              <div class="work-card-title">{{ t('building.profile.documents.forms.title') }}</div>
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
                  <th>{{ t('building.common.title') }}</th>
                  <th>{{ t('building.common.date') }}</th>
                  <th>{{ t('building.common.month') }}</th>
                  <th>{{ t('building.common.download') }}</th>
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
                    >{{ t('building.common.view') }}</a>
                    <span v-else>-</span>
                  </td>
                </tr>
                <tr v-if="buildingForms.length === 0">
                  <td
                    class="building-empty-row"
                    colspan="4"
                  >
                    {{ t('building.profile.documents.forms.empty') }}
                  </td>
                </tr>
              </tbody>
            </table>
          </section>
          <section class="work-card">
            <div class="building-section-head">
              <div class="work-card-title">{{ t('building.profile.documents.information.title') }}</div>
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
                  <th>{{ t('building.common.title') }}</th>
                  <th>{{ t('building.common.date') }}</th>
                  <th>{{ t('building.common.month') }}</th>
                  <th>{{ t('building.common.download') }}</th>
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
                    >{{ t('building.common.view') }}</a>
                    <span v-else>-</span>
                  </td>
                </tr>
                <tr v-if="buildingInfoFiles.length === 0">
                  <td
                    class="building-empty-row"
                    colspan="4"
                  >
                    {{ t('building.profile.documents.information.empty') }}
                  </td>
                </tr>
              </tbody>
            </table>
          </section>
          <section class="work-card">
            <div class="building-section-head">
              <div class="work-card-title">{{ t('building.profile.documents.floorPlans.title') }}</div>
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
                  <th>{{ t('building.common.title') }}</th>
                  <th>{{ t('building.common.date') }}</th>
                  <th>{{ t('building.common.month') }}</th>
                  <th>{{ t('building.common.download') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="p in floorPlans"
                  :key="p.key"
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
                    >{{ t('building.common.view') }}</a>
                    <span v-else>-</span>
                  </td>
                </tr>
                <tr v-if="floorPlans.length === 0">
                  <td
                    class="building-empty-row"
                    colspan="4"
                  >
                    {{ t('building.profile.documents.floorPlans.empty') }}
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
              <div class="work-kicker">{{ t('building.finance.kicker') }}</div>
              <h2 class="work-title">{{ t('building.finance.title') }}</h2>
              <p class="work-desc">{{ t('building.finance.description') }}</p>
            </div>
            <button
              type="button"
              class="work-action"
              :disabled="financeLoading"
              @click="loadBuildingFinance"
            >
              {{ financeLoading ? t('building.common.loading') : t('building.common.refresh') }}
            </button>
          </section>
          <section class="work-card acct-card-wrap">
            <div class="acct-tabs">
              <button
                type="button"
                class="acct-tab"
                :class="{ on: financeSubTab === 'management-overview' }"
                @click="switchFinanceSub('management-overview')"
              >
                {{ t('building.finance.tabs.overview') }}
              </button>
              <button
                type="button"
                class="acct-tab"
                :class="{ on: financeSubTab === 'financial-reports' }"
                @click="switchFinanceSub('financial-reports')"
              >
                {{ t('building.finance.tabs.financialReports', { count: formatLocaleNumber(financialReports.length) }) }}
              </button>
              <button
                type="button"
                class="acct-tab"
                :class="{ on: financeSubTab === 'audit-reports' }"
                @click="switchFinanceSub('audit-reports')"
              >
                {{ t('building.finance.tabs.auditReports', { count: formatLocaleNumber(auditReports.length) }) }}
              </button>
            </div>
            <div class="acct-body">
              <!-- 管理處總覽 -->
              <div
                v-show="financeSubTab === 'management-overview'"
                class="acct-subpanel"
                :class="{ on: financeSubTab === 'management-overview' }"
              >
                <div class="acct-note">{{ t('building.finance.syncNote') }}</div>
                <div
                  v-if="buildingInfoError"
                  class="acct-banner warn"
                >
                  {{ translateMessage(buildingInfoError) }}
                </div>
                <div class="acct-toolbar">
                  <div class="work-card-title acct-toolbar-title">{{ t('building.finance.tabs.overview') }}</div>
                  <div class="acct-total">{{ buildingName }}</div>
                </div>
                <div class="building-field-grid">
                  <div
                    v-for="f in managementOverviewFields"
                    :key="f.label"
                    class="building-field"
                  >
                    <span>{{ f.label }}</span>
                    <strong>{{ f.value }}</strong>
                  </div>
                </div>
                <div
                  v-if="financeReceivableError"
                  class="acct-banner warn"
                >
                  {{ translateMessage(financeReceivableError) }}
                </div>
                <div class="finance-summary-grid">
                  <div
                    v-for="f in financeReceivableFields"
                    :key="f.label"
                    class="finance-summary-item"
                  >
                    <span>{{ f.label }}</span>
                    <strong>{{ f.value }}</strong>
                  </div>
                </div>
                <div class="acct-toolbar">
                  <div class="work-card-title acct-toolbar-title">{{ t('building.finance.managementReceivable') }}</div>
                  <div class="acct-total">
                    {{ t('building.common.itemCount', { count: formatLocaleNumber(managementFeeRows.length) }) }}
                  </div>
                </div>
                <div class="acct-table-wrap">
                  <table class="acct-table finance-management-table">
                    <thead>
                      <tr>
                        <th
                          v-for="column in managementFeeColumns"
                          :key="column"
                        >
                          {{ financeColumnLabel(column) }}
                        </th>
                        <th v-if="managementFeeColumns.length === 0">{{ t('building.finance.data') }}</th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr
                        v-for="row in managementFeePreviewRows"
                        :key="managementFeeRowKey(row)"
                      >
                        <td
                          v-for="column in managementFeeColumns"
                          :key="column"
                        >
                          {{ financeCellText(row[column]) }}
                        </td>
                      </tr>
                      <tr v-if="managementFeePreviewRows.length === 0">
                        <td
                          class="building-empty-row"
                          :colspan="managementFeeColumnSpan"
                        >
                          {{ managementFeeEmptyText }}
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </div>
                <div class="acct-toolbar">
                  <div class="work-card-title acct-toolbar-title">{{ t('building.finance.otherReceivable') }}</div>
                  <div class="acct-total">
                    {{ t('building.common.total', { amount: formatOwnerHKD(otherFeeTotal) }) }}
                  </div>
                </div>
                <div class="acct-table-wrap">
                  <table class="acct-table finance-other-fee-table">
                    <thead>
                      <tr>
                        <th>{{ t('building.finance.columns.invoiceNumber') }}</th>
                        <th>{{ t('building.finance.columns.unit') }}</th>
                        <th>{{ t('building.finance.columns.item') }}</th>
                        <th>{{ t('building.finance.columns.term') }}</th>
                        <th>{{ t('building.finance.columns.amount') }}</th>
                        <th>{{ t('building.finance.columns.remark') }}</th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr
                        v-for="fee in otherFeePreviewRows"
                        :key="otherFeeRowKey(fee)"
                      >
                        <td>{{ financeCellText(fee.invoice_no) }}</td>
                        <td>{{ financeCellText(fee.flat_code) }}</td>
                        <td>{{ financeCellText(fee.item_id) }}</td>
                        <td>{{ financeCellText(fee.trs_to) }}</td>
                        <td class="acct-due">{{ formatOwnerHKD(ownerAmountValue(fee.trs_val)) }}</td>
                        <td>{{ financeCellText(fee.remark) }}</td>
                      </tr>
                      <tr v-if="otherFeePreviewRows.length === 0">
                        <td
                          class="building-empty-row"
                          colspan="6"
                        >
                          {{ otherFeeEmptyText }}
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>

              <!-- 財務報表 -->
              <div
                v-show="financeSubTab === 'financial-reports'"
                class="acct-subpanel"
                :class="{ on: financeSubTab === 'financial-reports' }"
              >
                <div
                  v-if="buildingInfoError"
                  class="acct-banner warn"
                >
                  {{ translateMessage(buildingInfoError) }}
                </div>
                <div class="acct-toolbar">
                  <div class="work-card-title acct-toolbar-title">{{ t('building.finance.financialReports') }}</div>
                  <div class="acct-total">
                    {{ t('building.common.fileCount', { count: formatLocaleNumber(financialReports.length) }) }}
                  </div>
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
                      <th>{{ t('building.common.title') }}</th>
                      <th>{{ t('building.common.date') }}</th>
                      <th>{{ t('building.common.month') }}</th>
                      <th>{{ t('building.common.download') }}</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr
                      v-for="report in financialReports"
                      :key="report.key"
                    >
                      <td class="building-file-title-cell">{{ report.title }}</td>
                      <td class="building-file-meta-cell">{{ report.date }}</td>
                      <td class="building-file-meta-cell">{{ report.month }}</td>
                      <td class="building-file-action-cell">
                        <a
                          v-if="report.url"
                          class="building-link"
                          :href="report.url"
                          target="_blank"
                          rel="noopener"
                        >{{ t('building.common.view') }}</a>
                        <span v-else>-</span>
                      </td>
                    </tr>
                    <tr v-if="financialReports.length === 0">
                      <td
                        class="building-empty-row"
                        colspan="4"
                      >
                        <span v-if="buildingInfoLoading">{{ t('building.finance.financialReportsLoading') }}</span>
                        <span v-else>{{ t('building.finance.financialReportsEmpty') }}</span>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>

              <!-- 核數報表 -->
              <div
                v-show="financeSubTab === 'audit-reports'"
                class="acct-subpanel"
                :class="{ on: financeSubTab === 'audit-reports' }"
              >
                <div
                  v-if="buildingInfoError"
                  class="acct-banner warn"
                >
                  {{ translateMessage(buildingInfoError) }}
                </div>
                <div class="acct-toolbar">
                  <div class="work-card-title acct-toolbar-title">{{ t('building.finance.auditReports') }}</div>
                  <div class="acct-total">
                    {{ t('building.common.fileCount', { count: formatLocaleNumber(auditReports.length) }) }}
                  </div>
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
                      <th>{{ t('building.common.title') }}</th>
                      <th>{{ t('building.common.date') }}</th>
                      <th>{{ t('building.common.month') }}</th>
                      <th>{{ t('building.common.download') }}</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr
                      v-for="report in auditReports"
                      :key="report.key"
                    >
                      <td class="building-file-title-cell">{{ report.title }}</td>
                      <td class="building-file-meta-cell">{{ report.date }}</td>
                      <td class="building-file-meta-cell">{{ report.month }}</td>
                      <td class="building-file-action-cell">
                        <a
                          v-if="report.url"
                          class="building-link"
                          :href="report.url"
                          target="_blank"
                          rel="noopener"
                        >{{ t('building.common.view') }}</a>
                        <span v-else>-</span>
                      </td>
                    </tr>
                    <tr v-if="auditReports.length === 0">
                      <td
                        class="building-empty-row"
                        colspan="4"
                      >
                        <span v-if="buildingInfoLoading">{{ t('building.finance.auditReportsLoading') }}</span>
                        <span v-else>{{ t('building.finance.auditReportsEmpty') }}</span>
                      </td>
                    </tr>
                  </tbody>
                </table>
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
              <div class="work-kicker">{{ t('building.ownerAccount.kicker') }}</div>
              <h2 class="work-title">{{ t('building.ownerAccount.title') }}</h2>
              <p class="work-desc">{{ t('building.ownerAccount.description') }}</p>
            </div>
            <button
              type="button"
              class="work-action"
              :disabled="ownerAccountLoading"
              @click="refreshOwnerAccount"
            >
              {{ ownerAccountLoading ? t('building.common.loading') : t('building.common.refresh') }}
            </button>
          </section>
          <section class="work-card acct-card-wrap">
            <div class="acct-tabs">
              <button
                type="button"
                class="acct-tab"
                :class="{ on: ownerAccountTab === 'owner-unpaid' }"
                @click="switchOwnerAccountTab('owner-unpaid')"
              >
                {{ t('building.ownerAccount.unpaidTab') }}
              </button>
              <button
                type="button"
                class="acct-tab"
                :class="{ on: ownerAccountTab === 'owner-records' }"
                @click="switchOwnerAccountTab('owner-records')"
              >
                {{ t('building.ownerAccount.recordsTab') }}
              </button>
            </div>
            <div class="acct-body">
              <div
                v-show="ownerAccountTab === 'owner-unpaid'"
                class="acct-subpanel"
                :class="{ on: ownerAccountTab === 'owner-unpaid' }"
              >
                <div class="acct-note">
                  {{ t('building.ownerAccount.boundUnit', { unit: ownerBoundUnitLabel }) }}
                </div>
                <div
                  v-if="ownerUnpaidError"
                  class="acct-banner warn"
                >
                  {{ translateMessage(ownerUnpaidError) }}
                </div>
                <div class="acct-toolbar">
                  <div class="work-card-title acct-toolbar-title">{{ t('building.ownerAccount.unpaidTab') }}</div>
                  <div class="acct-total">
                    {{ t('building.common.total', { amount: formatOwnerHKD(ownerUnpaidTotal) }) }}
                  </div>
                </div>
                <div class="acct-table-wrap">
                  <table class="acct-table owner-account-table">
                    <thead>
                      <tr>
                        <th>{{ t('building.ownerAccount.invoiceNumber') }}</th>
                        <th>{{ t('building.common.unit') }}</th>
                        <th>{{ t('building.common.item') }}</th>
                        <th>{{ t('building.common.term') }}</th>
                        <th>{{ t('building.ownerAccount.invoiceDate') }}</th>
                        <th>{{ t('building.common.amount') }}</th>
                        <th>{{ t('building.common.remark') }}</th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr
                        v-for="(invoice, invoiceIndex) in ownerUnpaidInvoices"
                        :key="ownerUnpaidInvoiceKey(invoice, invoiceIndex)"
                      >
                        <td>{{ ownerTextValue(invoice.invoice_no) }}</td>
                        <td>{{ ownerTextValue(invoice.flat_code) }}</td>
                        <td>{{ ownerTextValue(invoice.item_id) }}</td>
                        <td>{{ ownerTextValue(invoice.trs_to) }}</td>
                        <td>{{ formatLocaleDateValue(invoice.bill_dt) }}</td>
                        <td class="acct-due">{{ formatOwnerHKD(ownerAmountValue(invoice.net_amount)) }}</td>
                        <td>{{ ownerTextValue(invoice.remark) }}</td>
                      </tr>
                      <tr v-if="ownerUnpaidInvoices.length === 0">
                        <td
                          class="building-empty-row"
                          colspan="7"
                        >
                          {{ ownerUnpaidEmptyText }}
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>

              <div
                v-show="ownerAccountTab === 'owner-records'"
                class="acct-subpanel"
                :class="{ on: ownerAccountTab === 'owner-records' }"
              >
                <div class="acct-note">
                  {{ t('building.ownerAccount.boundUnit', { unit: ownerBoundUnitLabel }) }}
                </div>
                <form
                  class="acct-filter-row"
                  @submit.prevent="searchOwnerPaymentRecordsByDate"
                >
                  <label class="acct-filter-field">
                    <span>{{ t('building.ownerAccount.fromDate') }}</span>
                    <input
                      v-model="ownerRecordFromDate"
                      type="date"
                    >
                  </label>
                  <label class="acct-filter-field">
                    <span>{{ t('building.ownerAccount.toDate') }}</span>
                    <input
                      v-model="ownerRecordToDate"
                      type="date"
                    >
                  </label>
                  <label class="acct-filter-field">
                    <span>{{ t('building.ownerAccount.dateType') }}</span>
                    <select v-model="ownerRecordDateType">
                      <option value="input_date">{{ t('building.ownerAccount.inputDate') }}</option>
                      <option value="tran_date">{{ t('building.ownerAccount.transactionDate') }}</option>
                    </select>
                  </label>
                  <button
                    type="submit"
                    class="notice-action-btn"
                    :disabled="ownerRecordsLoading"
                  >
                    {{ ownerRecordsLoading ? t('building.ownerAccount.searching') : t('building.ownerAccount.search') }}
                  </button>
                  <button
                    type="button"
                    class="acct-secondary-action"
                    :disabled="ownerRecordsLoading"
                    @click="resetOwnerPaymentRecordSearch"
                  >
                    {{ t('building.ownerAccount.showAll') }}
                  </button>
                </form>
                <div
                  v-if="ownerRecordsError"
                  class="acct-banner warn"
                >
                  {{ translateMessage(ownerRecordsError) }}
                </div>
                <div class="acct-toolbar">
                  <div class="work-card-title acct-toolbar-title">{{ t('building.ownerAccount.recordsTab') }}</div>
                  <div class="acct-total">
                    {{ t('building.common.total', { amount: formatOwnerHKD(ownerPaymentRecordTotal) }) }}
                  </div>
                </div>
                <div class="acct-table-wrap">
                  <table class="acct-table owner-record-table">
                    <thead>
                      <tr>
                        <th>{{ t('building.ownerAccount.receiptNumber') }}</th>
                        <th>{{ t('building.ownerAccount.paymentNumber') }}</th>
                        <th>{{ t('building.ownerAccount.inputTime') }}</th>
                        <th>{{ t('building.ownerAccount.transactionTime') }}</th>
                        <th>{{ t('building.common.unit') }}</th>
                        <th>{{ t('building.common.item') }}</th>
                        <th>{{ t('building.common.term') }}</th>
                        <th>{{ t('building.common.amount') }}</th>
                        <th>{{ t('building.ownerAccount.paymentMethod') }}</th>
                        <th>{{ t('building.common.status') }}</th>
                        <th>{{ t('building.common.remark') }}</th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr
                        v-for="row in ownerPaymentRecordRows"
                        :key="row.key"
                      >
                        <td>{{ row.receiptID }}</td>
                        <td>{{ row.paymentID }}</td>
                        <td>{{ row.inputTime }}</td>
                        <td>{{ row.tranTime }}</td>
                        <td>{{ row.unit }}</td>
                        <td>{{ row.item }}</td>
                        <td>{{ row.term }}</td>
                        <td>{{ formatOwnerHKD(row.amount) }}</td>
                        <td>{{ row.payType }}</td>
                        <td :class="row.statusClass">{{ row.status }}</td>
                        <td>{{ row.remark }}</td>
                      </tr>
                      <tr v-if="ownerPaymentRecordRows.length === 0">
                        <td
                          class="building-empty-row"
                          colspan="11"
                        >
                          {{ ownerRecordsEmptyText }}
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </div>
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
              <div class="work-kicker">{{ t('building.forms.kicker') }}</div>
              <h2 class="work-title">{{ t('building.forms.title') }}</h2>
              <p class="work-desc">{{ t('building.forms.description') }}</p>
            </div>
          </section>
          <section class="notice-admin-grid">
            <div class="notice-admin-card">
              <h3>{{ t('building.forms.organization') }}</h3>
              <div class="notice-admin-label">{{ t('building.forms.currentOrganization') }}</div>
              <div class="notice-select building-readonly-select">{{ selectedFormOrg }}</div>
            </div>
            <div class="notice-admin-card">
              <h3>{{ t('building.forms.building') }}</h3>
              <div class="notice-admin-label">{{ t('building.forms.currentBuilding') }}</div>
              <div class="notice-select building-readonly-select">{{ selectedFormBuilding }}</div>
            </div>
          </section>
          <section class="work-card">
            <table class="work-table">
              <thead>
                <tr>
                  <th>{{ t('building.common.title') }}</th>
                  <th>{{ t('building.forms.view') }}</th>
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
                    >{{ t('building.forms.fill') }}</a>
                    <span v-else>-</span>
                  </td>
                </tr>
                <tr v-if="forms.length === 0">
                  <td
                    class="building-empty-row"
                    colspan="2"
                  >
                    {{ t('building.forms.empty') }}
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
              <div class="work-kicker">{{ t('building.feedback.kicker') }}</div>
              <h2 class="work-title">{{ t('building.feedback.title') }}</h2>
              <p class="work-desc">{{ t('building.feedback.description') }}</p>
            </div>
          </section>
          <section class="work-card">
            <div class="work-card-title">{{ t('building.feedback.recentRecords') }}</div>
            <table class="work-table">
              <thead>
                <tr>
                  <th>{{ t('building.common.type') }}</th>
                  <th>{{ t('building.common.building') }}</th>
                  <th>{{ t('building.feedback.subject') }}</th>
                  <th>{{ t('building.feedback.category') }}</th>
                  <th>{{ t('building.common.status') }}</th>
                  <th>{{ t('building.feedback.updatedAt') }}</th>
                  <th>{{ t('building.common.action') }}</th>
                </tr>
              </thead>
              <tbody>
              <template
                v-for="(r, i) in feedbackRecords"
                :key="`${r.type}-${r.subject}`"
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
                      {{ t('building.feedback.viewContent') }}
                    </button>
                  </td>
                </tr>
                <tr
                  v-show="expandedRecord === i"
                  class="affairs-record-detail"
                >
                  <td colspan="7">
                    <div class="affairs-record-card">
                      <strong>{{ t('building.common.content') }}</strong>
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
            :aria-label="t('building.feedback.entryAria')"
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
                <strong>{{ t('building.feedback.repairEntryTitle') }}</strong>
                <span>{{ t('building.feedback.repairEntryDescription') }}</span>
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
                <strong>{{ t('building.feedback.feedbackEntryTitle') }}</strong>
                <span>{{ t('building.feedback.feedbackEntryDescription') }}</span>
              </span>
            </button>
          </section>
          <section class="work-card affairs-feedback-form">
            <div class="work-card-title">{{ t('building.feedback.selectBuilding') }}</div>
            <div class="affairs-field">
              <label for="affairs-building">{{ t('building.common.building') }}</label>
              <select
                id="affairs-building"
                v-model="selectedFeedbackBuilding"
                class="affairs-select"
              >
                <option value="">{{ t('building.common.emptySelection') }}</option>
                <option
                  v-for="b in feedbackBuildings"
                  :key="b.value"
                  :value="b.value"
                >
                  {{ b.label }}
                </option>
              </select>
            </div>
          </section>
          <section class="work-card affairs-guided-shell">
            <div
              class="affairs-stepper"
              :aria-label="t('building.feedback.flowAria')"
            >
              <span
                class="affairs-step-indicator"
                :class="{ on: affairsStep === 1 }"
              >
                <b>1</b>{{ t('building.feedback.steps.category') }}
              </span>
              <span
                class="affairs-step-indicator"
                :class="{ on: affairsStep === 2 }"
              >
                <b>2</b>{{ t('building.feedback.steps.content') }}
              </span>
              <span
                class="affairs-step-indicator"
                :class="{ on: affairsStep === 3 }"
              >
                <b>3</b>{{ t('building.feedback.steps.upload') }}
              </span>
              <span
                class="affairs-step-indicator"
                :class="{ on: affairsStep === 4 }"
              >
                <b>4</b>{{ t('building.feedback.steps.confirm') }}
              </span>
            </div>

            <!-- 步驟 1：選擇分類 -->
            <div
              v-show="affairsStep === 1"
              class="affairs-guided-step"
            >
              <div class="work-card-title">{{ t('building.feedback.steps.category') }}</div>
              <div
                v-if="affairsMode === 'repair'"
                class="affairs-step-hint"
              >
                {{ t('building.feedback.repairCategoryHint') }}
              </div>
              <div
                v-else
                class="affairs-step-hint"
              >
                {{ t('building.feedback.feedbackCategoryHint') }}
              </div>
              <div
                v-if="affairsMode === 'repair'"
                class="affairs-field-grid"
              >
                <div class="affairs-field">
                  <label for="affairs-repair-category">{{ t('building.feedback.repairCategory') }}</label>
                  <select
                    id="affairs-repair-category"
                    v-model="repairCategory"
                    class="affairs-select"
                  >
                    <option value="">{{ t('building.common.emptySelection') }}</option>
                    <option
                      v-for="c in repairCategories"
                      :key="c.value"
                      :value="c.value"
                    >
                      {{ c.label }}
                    </option>
                  </select>
                </div>
                <div class="affairs-field">
                  <label for="affairs-repair-subcategory">{{ t('building.feedback.repairSubcategory') }}</label>
                  <select
                    id="affairs-repair-subcategory"
                    v-model="repairSubcategory"
                    class="affairs-select"
                  >
                    <option value="">{{ t('building.common.emptySelection') }}</option>
                    <option
                      v-for="s in repairSubcategories"
                      :key="s.value"
                      :value="s.value"
                    >
                      {{ s.label }}
                    </option>
                  </select>
                </div>
              </div>
              <div
                v-else
                class="affairs-field-grid"
              >
                <div class="affairs-field">
                  <label for="affairs-feedback-category">{{ t('building.feedback.feedbackCategory') }}</label>
                  <select
                    id="affairs-feedback-category"
                    v-model="feedbackCategory"
                    class="affairs-select"
                  >
                    <option value="">{{ t('building.common.emptySelection') }}</option>
                    <option
                      v-for="c in feedbackCategories"
                      :key="c.value"
                      :value="c.value"
                    >
                      {{ c.label }}
                    </option>
                  </select>
                </div>
                <div class="affairs-field">
                  <label for="affairs-feedback-subcategory">{{ t('building.feedback.feedbackSubcategory') }}</label>
                  <select
                    id="affairs-feedback-subcategory"
                    v-model="feedbackSubcategory"
                    class="affairs-select"
                  >
                    <option value="">{{ t('building.common.emptySelection') }}</option>
                    <option
                      v-for="s in feedbackSubcategories"
                      :key="s.value"
                      :value="s.value"
                    >
                      {{ s.label }}
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
              <div class="work-card-title">{{ t('building.feedback.steps.content') }}</div>
              <div
                v-if="affairsMode === 'repair'"
                class="affairs-field"
              >
                <label for="affairs-repair-content">{{ t('building.common.content') }}</label>
                <textarea
                  id="affairs-repair-content"
                  v-model="repairContent"
                  class="affairs-textarea"
                  :placeholder="t('building.feedback.repairPlaceholder')"
                />
              </div>
              <div
                v-else
                class="affairs-field"
              >
                <label for="affairs-feedback-content">{{ t('building.common.content') }}</label>
                <textarea
                  id="affairs-feedback-content"
                  v-model="feedbackContent"
                  class="affairs-textarea"
                  :placeholder="t('building.feedback.feedbackPlaceholder')"
                />
              </div>
            </div>

            <!-- 步驟 3：上載圖片/影片 -->
            <div
              v-show="affairsStep === 3"
              class="affairs-guided-step"
            >
              <div class="work-card-title">{{ t('building.feedback.steps.upload') }}</div>
              <div class="affairs-step-hint">
                {{ t('building.feedback.uploadHint') }}
              </div>
              <div class="affairs-upload-box">
                <strong>{{ t('building.feedback.uploadTitle') }}</strong>
                <span>{{ t('building.feedback.uploadDescription') }}</span>
                <input
                  type="file"
                  accept="image/*,video/*"
                  multiple
                  :aria-label="t('building.feedback.uploadTitle')"
                  @change="onMediaChange"
                >
              </div>
            </div>

            <!-- 步驟 4：確認提交 -->
            <div
              v-show="affairsStep === 4"
              class="affairs-guided-step"
            >
              <div class="work-card-title">{{ t('building.feedback.steps.confirm') }}</div>
              <div
                v-if="affairsMode === 'repair'"
                class="affairs-step-hint"
              >
                {{ t('building.feedback.repairConfirmHint') }}
              </div>
              <div
                v-else
                class="affairs-step-hint"
              >
                {{ t('building.feedback.privacyHint') }}
              </div>
              <div class="affairs-review-list">
                <div class="affairs-review-row">
                  <span>{{ t('building.common.type') }}</span>
                  <strong>{{ reviewMode }}</strong>
                </div>
                <div class="affairs-review-row">
                  <span>{{ t('building.common.building') }}</span>
                  <strong>{{ selectedFeedbackBuildingLabel }}</strong>
                </div>
                <div class="affairs-review-row">
                  <span>{{ t('building.feedback.majorCategory') }}</span>
                  <strong>{{ reviewCategory }}</strong>
                </div>
                <div class="affairs-review-row">
                  <span>{{ t('building.feedback.subcategory') }}</span>
                  <strong>{{ reviewSubcategory }}</strong>
                </div>
                <div class="affairs-review-row">
                  <span>{{ t('building.common.content') }}</span>
                  <strong>{{ reviewContent }}</strong>
                </div>
                <div class="affairs-review-row">
                  <span>{{ t('building.feedback.attachment') }}</span>
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
                {{ t('building.feedback.previous') }}
              </button>
              <button
                v-if="affairsStep < 4"
                type="button"
                class="work-action"
                @click="nextAffairsStep"
              >
                {{ t('building.feedback.next') }}
              </button>
              <button
                v-else
                type="button"
                class="work-action"
                @click="submitAffairsFeedback"
              >
                {{ t('building.feedback.submit') }}
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
              <div class="work-kicker">{{ t('building.access.kicker') }}</div>
              <h2 class="work-title">{{ t('building.access.title') }}</h2>
              <p class="work-desc">{{ t('building.access.description') }}</p>
            </div>
            <button
              type="button"
              class="work-action"
              :disabled="accessLoading"
              @click="loadBuildingAccess()"
            >
              {{ accessLoading ? t('building.common.loading') : t('building.common.refresh') }}
            </button>
          </section>

          <section class="access-summary-grid">
            <div class="access-summary-card">
              <div class="work-card-title">{{ t('building.access.currentBuilding') }}</div>
              <div class="access-stat-value">{{ accessBuildingName }}</div>
              <div class="work-stat-label">{{ t('building.access.boundDataHint') }}</div>
            </div>
            <div class="access-summary-card">
              <div class="work-card-title">{{ t('building.access.doorCount') }}</div>
              <div class="work-stat">{{ formatLocaleNumber(accessDoors.length) }}</div>
              <div class="work-stat-label">{{ t('building.access.visibleDoors') }}</div>
            </div>
            <div class="access-summary-card">
              <div class="work-card-title">{{ t('building.access.allowedDoors') }}</div>
              <div class="work-stat">{{ formatLocaleNumber(accessAllowedDoorCount) }}</div>
              <div class="work-stat-label">{{ t('building.access.authorizedDoors') }}</div>
            </div>
            <div class="access-summary-card">
              <div class="work-card-title">{{ t('building.access.dataStatus') }}</div>
              <span
                class="work-chip"
                :class="accessStatusClass"
              >{{ accessStatusText }}</span>
              <div class="work-stat-label">
                {{ t('building.access.qrDoorCount', { count: formatLocaleNumber(accessQRCodeDoorCount) }) }}
              </div>
            </div>
          </section>

          <div
            v-if="accessError || accessMessage"
            class="access-banner"
            :class="{ warn: accessError, good: accessMessage && !accessError }"
          >
            {{ translateMessage(accessError || accessMessage) }}
          </div>

          <section class="access-door-section">
            <div class="building-section-head">
              <div class="work-card-title">{{ t('building.access.doorList') }}</div>
              <span class="work-chip">
                {{ accessLoading
                  ? t('building.common.loading')
                  : t('building.common.doorCount', { count: formatLocaleNumber(accessDoors.length) }) }}
              </span>
            </div>

            <div
              v-if="accessLoading && accessDoors.length === 0"
              class="access-empty"
            >
              {{ t('building.access.loadingDoors') }}
            </div>
            <div
              v-else-if="accessDoors.length === 0"
              class="access-empty"
            >
              {{ t('building.access.emptyDoors') }}
            </div>
            <div
              v-else
              class="access-table-wrap"
            >
              <table class="work-table access-door-table">
                <thead>
                  <tr>
                    <th>{{ t('building.access.door') }}</th>
                    <th>{{ t('building.access.doorNumber') }}</th>
                    <th>{{ t('building.access.building') }}</th>
                    <th>{{ t('building.access.camera') }}</th>
                    <th>{{ t('building.access.permission') }}</th>
                    <th>{{ t('building.access.password') }}</th>
                    <th>{{ t('building.access.validPeriod') }}</th>
                    <th>{{ t('building.common.action') }}</th>
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
                        >{{ door.has_permission ? t('building.access.canOpen') : t('building.access.unauthorized') }}</span>
                        <span
                          v-if="door.is_public"
                          class="work-chip brand"
                        >{{ t('building.access.publicDoor') }}</span>
                        <span
                          v-if="door.is_qrcode_enabled"
                          class="work-chip"
                        >{{ t('building.access.qrCode') }}</span>
                      </div>
                    </td>
                    <td>{{ accessDoorPasswordText(door) }}</td>
                    <td class="access-door-period-cell">
                      <div>
                        <span>{{ t('building.access.passwordShort') }}</span>
                        <strong>{{ accessTimeRange(door.password?.start_time, door.password?.end_time) }}</strong>
                      </div>
                      <div>
                        <span>{{ t('building.access.qrCode') }}</span>
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
                          {{ openingDoorID === accessDoorID(door)
                            ? t('building.access.processing')
                            : t('building.access.openDoor') }}
                        </button>
                        <button
                          v-if="door.password?.value"
                          type="button"
                          class="work-mini-btn"
                          @click="toggleAccessPassword(door)"
                        >
                          {{ accessDoorPasswordVisible(door)
                            ? t('building.access.hidePassword')
                            : t('building.access.viewPassword') }}
                        </button>
                        <button
                          v-if="door.is_qrcode_enabled && door.qrcode?.record_id"
                          type="button"
                          class="work-mini-btn"
                          :disabled="qrLoadingDoorID === accessDoorID(door)"
                          @click="generateAccessQRCode(door)"
                        >
                          {{ qrLoadingDoorID === accessDoorID(door)
                            ? t('building.access.generating')
                            : t('building.access.qrCode') }}
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
                  :alt="t('building.access.qrAlt')"
                />
              </div>
              <div>
                <div class="work-card-title">{{ accessQRPanel.doorTitle }}</div>
                <div class="work-card-sub">{{ t('building.access.qrHint') }}</div>
                <div class="access-qr-meta">
                  <div>
                    <span>{{ t('building.common.type') }}</span>
                    <strong>{{ accessTermText(accessQRPanel.term) }}</strong>
                  </div>
                  <div>
                    <span>{{ t('building.access.expiresAt') }}</span>
                    <strong>{{ formatLocaleDateValue(accessQRPanel.expiresAt) }}</strong>
                  </div>
                </div>
              </div>
            </div>
          </section>

          <section class="work-card">
            <div class="building-section-head">
              <div class="work-card-title">{{ t('building.access.recentRecords') }}</div>
            </div>
            <table class="work-table access-record-table">
              <thead>
                <tr>
                  <th>{{ t('building.access.door') }}</th>
                  <th>{{ t('building.common.time') }}</th>
                  <th>{{ t('building.access.method') }}</th>
                  <th>{{ t('building.common.status') }}</th>
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
                    {{ t('building.access.emptyRecords') }}
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
              <div class="work-kicker">{{ t('building.icctv.kicker') }}</div>
              <h2 class="work-title">{{ t('building.icctv.title') }}</h2>
              <p class="work-desc">{{ t('building.icctv.description') }}</p>
            </div>
            <button
              type="button"
              class="work-action"
              :disabled="icctvLoading"
              @click="loadICCTV()"
            >
              {{ icctvLoading ? t('building.common.loading') : t('building.common.refresh') }}
            </button>
          </section>
          <section class="work-card icctv-panel-wide icctv-table-card">
            <div class="icctv-profile-head">
              <div>
                <h3>{{ icctvBuildingTitle }}</h3>
                <p>{{ t('building.icctv.boundUnitHint') }}</p>
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
                    <th>{{ t('building.icctv.camera') }}</th>
                    <th>{{ t('building.icctv.status') }}</th>
                    <th>{{ t('building.icctv.action') }}</th>
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
                            :disabled="!camera.is_active || !camera.url"
                            @click="toggleICCTVCamera(camera.id)"
                          >
                            {{ isICCTVCameraExpanded(camera.id)
                              ? t('building.icctv.collapse')
                              : t('building.icctv.view') }}
                          </button>
                          <button
                            type="button"
                            class="work-mini-btn"
                            :disabled="!camera.is_active || !camera.url"
                            @click="openIcctvWindow(camera)"
                          >
                            {{ t('building.icctv.newWindow') }}
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
                      {{ icctvLoading
                        ? t('building.icctv.loadingCameras')
                        : translateMessage(icctvError) || t('building.icctv.emptyCameras') }}
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </section>
        </div>

      </main>
    </div>
  </div>
</template>

<style scoped>
/* 1. 頁面容器與雙欄布局 */
#page-affairs {
  min-height: calc(100svh - var(--nav-h, 52px));
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

.building-binding-empty {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 16px;
}

.building-binding-empty .work-card-title {
  margin-bottom: 4px;
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

.notice-select:disabled {
  color: var(--ink-3);
  cursor: not-allowed;
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

.notice-action-btn:disabled,
.notice-action-btn:disabled:hover {
  border-color: var(--bdr);
  background: var(--sur-3);
  color: var(--ink-3);
  cursor: not-allowed;
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
  background: rgb(var(--color-surface));
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
  background: rgb(var(--color-surface));
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
  background: rgb(var(--color-surface));
}

.acct-tab {
  border: 0;
  border-right: 1px solid var(--bdr);
  background: rgb(var(--color-surface));
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
  background: rgb(var(--color-surface));
  color: var(--ink-2);
  font-size: 12px;
  font-weight: 600;
  line-height: 1.7;
  padding: 12px 14px;
}

.acct-banner {
  border: 1px solid var(--success);
  border-radius: 8px;
  background: var(--success-bg);
  color: var(--success);
  font-size: 13px;
  font-weight: 700;
  padding: 11px 14px;
}

.acct-banner.warn {
  border-color: var(--warning);
  background: var(--warning-bg);
  color: var(--warning);
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

.acct-total {
  color: var(--ink);
  font-size: 13px;
  font-weight: 900;
  white-space: nowrap;
}

/* 3. 帳目搜尋框 */
.acct-search {
  max-width: 320px;
  border: 1px solid var(--bdr);
  border-radius: 6px;
  background: rgb(var(--color-surface));
  color: var(--ink);
  font-family: inherit;
  font-size: 13px;
  font-weight: 600;
  padding: 9px 12px;
}

.acct-filter-row {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  gap: 10px;
  border: 1px solid var(--bdr);
  border-radius: 8px;
  background: var(--sur);
  padding: 12px;
}

.acct-filter-field {
  display: grid;
  min-width: 150px;
  gap: 5px;
}

.acct-filter-field span {
  color: var(--ink-3);
  font-size: 12px;
  font-weight: 800;
}

.acct-filter-field input,
.acct-filter-field select {
  min-height: 32px;
  border: 1px solid var(--bdr);
  border-radius: 6px;
  background: rgb(var(--color-surface));
  color: var(--ink);
  font-family: inherit;
  font-size: 13px;
  font-weight: 700;
  padding: 6px 9px;
}

.acct-secondary-action {
  border: 1px solid var(--bdr);
  border-radius: 6px;
  background: rgb(var(--color-surface));
  color: var(--ink-2);
  cursor: pointer;
  font-family: inherit;
  font-size: 12px;
  font-weight: 800;
  min-height: 32px;
  padding: 6px 10px;
}

.acct-secondary-action:hover {
  border-color: var(--brand-mid);
  color: var(--brand);
}

.acct-secondary-action:disabled,
.acct-secondary-action:disabled:hover {
  border-color: var(--bdr);
  background: var(--sur-3);
  color: var(--ink-3);
  cursor: not-allowed;
}

/* 4. 帳目表格 */
.acct-table-wrap {
  overflow: auto;
  border: 1px solid var(--bdr);
  border-radius: 8px;
  background: rgb(var(--color-surface));
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

.owner-account-table {
  min-width: 900px;
}

.owner-record-table {
  min-width: 1280px;
}

.finance-management-table {
  min-width: 980px;
}

.finance-other-fee-table {
  min-width: 900px;
}

.finance-summary-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 10px;
}

.finance-summary-item {
  display: grid;
  gap: 6px;
  border: 1px solid var(--bdr);
  border-radius: 8px;
  background: rgb(var(--color-surface));
  padding: 12px;
}

.finance-summary-item span {
  color: var(--ink-3);
  font-size: 12px;
  font-weight: 800;
}

.finance-summary-item strong {
  color: var(--ink);
  font-size: 15px;
  font-weight: 900;
}

/* 5. 帳目狀態色 */
.acct-unit-cell {
  color: var(--ink);
}

.acct-paid {
  color: var(--success);
}

.acct-pending {
  color: var(--warning);
}

.acct-error,
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
  background: rgb(var(--color-surface));
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
  background: rgb(var(--color-surface));
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
  background: rgb(var(--color-surface));
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
  background: rgb(var(--color-surface));
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
  .finance-summary-grid,
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

  .building-binding-empty {
    align-items: stretch;
    flex-direction: column;
  }

  .building-binding-empty .work-action {
    align-self: flex-start;
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

  .acct-filter-row {
    align-items: stretch;
    flex-direction: column;
  }

  .acct-filter-field {
    min-width: 0;
  }

  .acct-filter-row .notice-action-btn,
  .acct-secondary-action {
    width: 100%;
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

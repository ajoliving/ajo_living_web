<!--
 * AJO Pay 付款頁。
 * 1. 高保真還原 HTML 原型的四步付款頁布局。
 * 2. 第一步直接展示目前會員單位的所有待繳賬單，不再依賴購物車頁。
 * 3. 第四步承載線上訂單狀態與最近訂單展示。
-->
<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import {
  createPOSPaymentOrder,
  fetchPOSPaymentBankAccounts,
  fetchPOSPaymentBills,
  fetchPOSPaymentFees,
  fetchPOSPaymentOrder,
  fetchPOSPaymentOrders,
  queryPOSPaymentOrder,
  reportPOSPayment,
} from '@/httpapis/payments';
import type {
  POSPaymentChannel,
  POSPaymentContext,
  POSPaymentMethodKey,
  POSPaymentRow,
} from '@/model/payments';
import QrCodeImage from '@/shared/components/base/QrCodeImage.vue';
import AppBreadcrumb from '@/shared/components/navigation/AppBreadcrumb.vue';
import { useFeedbackStore } from '@/stores/feedback';
import { usePreferenceStore } from '@/stores/preferences';

type CheckoutMethodKey =
  | 'POS_WECHAT'
  | 'POS_ALIPAY_CN'
  | 'POS_ALIPAY_HK'
  | 'POS_YSF_QR'
  | 'POS_BANK'
  | 'POS_CHEQUE';

interface CheckoutMethod {
  key: CheckoutMethodKey;
  labelKey: string;
  shortLabel: string;
  category: 'online' | 'offline';
  payType: POSPaymentMethodKey;
  walletType?: 'CN' | 'HK';
  variant?: 'wechat' | 'alipay' | 'ysf';
  feeRate: number;
}

interface BankAccountOption {
  label: string;
  value: string;
}

interface VoucherImage {
  name: string;
  data: string;
  previewUrl: string;
}

interface PaymentAPIError {
  response?: {
    status?: number;
    data?: {
      code?: string;
      message?: string;
    };
  };
}

const route = useRoute();
const router = useRouter();
const { t } = useI18n();
const feedbackStore = useFeedbackStore();
const preferenceStore = usePreferenceStore();
const rows = ref<POSPaymentRow[]>([]);
const orders = ref<POSPaymentRow[]>([]);
const context = ref<POSPaymentContext | undefined>();
const selectedBillKeys = ref<string[]>([]);
const methods = ref<CheckoutMethod[]>([]);
const bankAccounts = ref<BankAccountOption[]>([]);
const selectedMethodKey = ref<CheckoutMethodKey>('POS_WECHAT');
const isLoading = ref(false);
const isSettingsLoading = ref(false);
const isSubmitting = ref(false);
const isRefreshingOrder = ref(false);
const profileRequired = ref(false);
const posLoginRequired = ref(false);
const resultOrder = ref<POSPaymentRow | null>(null);
const resultReceipt = ref('');
const resultMessage = ref('');
const voucherInput = ref<HTMLInputElement | null>(null);
const voucherImages = ref<VoucherImage[]>([]);
const posPaymentExpireSeconds = 180;
const breadcrumbItems = computed(() => [
  { label: t('account.payments.checkout.breadcrumbHome'), to: '/' },
  { label: t('account.payments.checkout.breadcrumbPayment'), to: '/payments' },
  { label: t('account.payments.checkout.breadcrumbCheckout') },
]);

const offlineForm = reactive({
  tranDateTime: '',
  tranRefNo: '',
  remark: '',
  bankAccountNo: '',
});

// 1. 讀取字串值
const readPaymentText = (row: POSPaymentRow, keys: string[]): string => {
  for (const key of keys) {
    const value = row[key];
    if (value !== undefined && value !== null && String(value).trim() !== '') {
      return String(value).trim();
    }
  }
  return '';
};

// 2. 讀取金額值
const readPaymentAmountValue = (row: POSPaymentRow): number => {
  const raw = readPaymentText(row, [
    'net_amount',
    'amount',
    'final_amount',
    'paid_amount',
    'trs_val',
    'total',
    'total_amount',
    'payable',
  ]);
  const numeric = Number(raw);
  return Number.isFinite(numeric) ? numeric : 0;
};

// 3. 格式化港幣金額
const formatHKD = (value: number): string =>
  new Intl.NumberFormat(preferenceStore.locale, {
    style: 'currency',
    currency: 'HKD',
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(value);

// 4. 讀取賬單標題
const readBillTitle = (row: POSPaymentRow, fallback = t('account.payments.checkout.billFallback')): string =>
  readPaymentText(row, [
    'invoice_no',
    'bill_no',
    'bill_number',
    'bill_id',
    'item_name',
    'name',
    'item_id',
    'id',
  ]) || fallback;

// 5. 讀取賬單名稱
const readBillName = (row: POSPaymentRow): string =>
  readPaymentText(row, ['item_name', 'item_id', 'name', 'fee_name', 'item']) || t('account.payments.checkout.defaultBillName');

// 6. 讀取賬單日期
const readBillDate = (row: POSPaymentRow): string =>
  readPaymentText(row, ['bill_dt', 'bill_date', 'tran_time', 'tran_datetime', 'input_time', 'created_at', 'date']) || '-';

// 7. 建立賬單唯一鍵
const billKey = (row: POSPaymentRow, index = 0): string =>
  readPaymentText(row, ['invoice_no', 'bill_no', 'bill_number', 'bill_id', 'id']) ||
  [
    readBillName(row),
    readPaymentText(row, ['trs_to', 'term', 'period']),
    readBillDate(row),
    String(readPaymentAmountValue(row)),
    String(index),
  ].join('|');

// 8. 建立不依賴列表位置的賬單匹配鍵
const billIdentityKey = (row: POSPaymentRow): string =>
  readPaymentText(row, ['invoice_no', 'bill_no', 'bill_number', 'bill_id', 'id']) ||
  [
    readBillName(row),
    readPaymentText(row, ['trs_to', 'term', 'period']),
    readBillDate(row),
    String(readPaymentAmountValue(row)),
  ].join('|');

// 9. 判斷是否為管理費
const isManagementBill = (row: POSPaymentRow): boolean => {
  const label = readBillName(row).toLowerCase();
  return label.includes('管理') || label.includes('management');
};

const activeBuildingID = computed(() => context.value?.building_id ?? '');
const activeUnitID = computed(() => context.value?.unit_id ?? '');
const paymentQuery = computed(() => ({
  building_id: activeBuildingID.value || undefined,
  unit_id: activeUnitID.value || undefined,
}));
const payableRows = computed(() => rows.value.filter((row) => readPaymentAmountValue(row) > 0));
const selectedBillKeySet = computed(() => new Set(selectedBillKeys.value));
const selectedRows = computed(() =>
  payableRows.value.filter((row, index) => selectedBillKeySet.value.has(billKey(row, index))),
);
const allPayableBillsSelected = computed(() =>
  payableRows.value.length > 0 &&
  payableRows.value.every((row, index) => selectedBillKeySet.value.has(billKey(row, index))),
);
const selectedAmount = computed(() =>
  selectedRows.value.reduce((sum, row) => sum + readPaymentAmountValue(row), 0),
);
const managementAmount = computed(() =>
  selectedRows.value.filter(isManagementBill).reduce((sum, row) => sum + readPaymentAmountValue(row), 0),
);
const otherAmount = computed(() =>
  selectedRows.value.filter((row) => !isManagementBill(row)).reduce((sum, row) => sum + readPaymentAmountValue(row), 0),
);
const selectedMethod = computed(() =>
  methods.value.find((item) => item.key === selectedMethodKey.value) ?? null,
);
const feeAmount = computed(() =>
  selectedRows.value.reduce((sum, row) => sum + calculateFeeAmount(readPaymentAmountValue(row), selectedMethod.value?.feeRate ?? 0), 0),
);
const finalAmount = computed(() => selectedAmount.value + feeAmount.value);
const rewardCoins = computed(() => Math.floor(finalAmount.value / 100));
const hasContext = computed(() => activeBuildingID.value !== '' && activeUnitID.value !== '');
const isBankTransferMethod = computed(() => selectedMethod.value?.key === 'POS_BANK');
const isChequeMethod = computed(() => selectedMethod.value?.key === 'POS_CHEQUE');
const requiresBankAccount = computed(() => isBankTransferMethod.value);
const contextLabel = computed(() => {
  if (!context.value) {
    if (profileRequired.value) {
      return t('account.payments.checkout.contextBindUnit');
    }
    if (posLoginRequired.value) {
      return t('account.payments.checkout.contextBindIsmart');
    }
    return t('account.payments.checkout.contextUnavailable');
  }
  return [
    context.value.building_name || context.value.building_id,
    context.value.unit_label || [context.value.floor, context.value.unit].filter(Boolean).join(' / '),
  ].filter(Boolean).join(' · ');
});
const activeStepIndex = computed(() => {
  if (resultOrder.value || resultReceipt.value) {
    return 4;
  }
  if (selectedAmount.value > 0 && selectedMethod.value) {
    return 3;
  }
  if (selectedAmount.value > 0) {
    return 2;
  }
  return 1;
});
const canSubmit = computed(() => (
  hasContext.value &&
  selectedRows.value.length > 0 &&
  finalAmount.value > 0 &&
  selectedMethod.value !== null &&
  (!requiresBankAccount.value || offlineForm.bankAccountNo !== '') &&
  (!isChequeMethod.value || offlineForm.tranRefNo.trim() !== '') &&
  !isSubmitting.value
));
const confirmButtonLabel = computed(() => {
  if (isSubmitting.value) {
    return t('account.payments.checkout.submitting');
  }
  if (selectedRows.value.length === 0) {
    return t('account.payments.checkout.selectItemsFirst');
  }
  if (!selectedMethod.value) {
    return t('account.payments.checkout.selectMethodFirst');
  }
  return t('account.payments.checkout.confirmAmount', { amount: formatHKD(finalAmount.value) });
});
const resultPayData = computed(() =>
  resultOrder.value ? readPaymentText(resultOrder.value, ['pay_data', 'payData']) : '',
);
const resultPayDataType = computed(() =>
  resultOrder.value ? readPaymentText(resultOrder.value, ['pay_data_type', 'payDataType']).toLowerCase() : '',
);
const resultQrImageUrl = computed(() =>
  resultPayDataType.value === 'codeimgurl' ? resultPayData.value : '',
);
const resultQrText = computed(() =>
  ['codeurl', 'payurl', ''].includes(resultPayDataType.value) ? resultPayData.value : '',
);
const displayedOrders = computed(() => {
  const resultKey = resultOrder.value ? readOrderKey(resultOrder.value) : '';
  const list = resultOrder.value ? [resultOrder.value] : [];
  orders.value.forEach((row) => {
    if (!resultKey || readOrderKey(row) !== resultKey) {
      list.push(row);
    }
  });
  return list.slice(0, 4);
});

// 10. 建立本地日期時間
const nowLocalDateTime = (): string => {
  const now = new Date();
  const pad = (value: number): string => String(value).padStart(2, '0');
  return `${now.getFullYear()}-${pad(now.getMonth() + 1)}-${pad(now.getDate())}T${pad(now.getHours())}:${pad(now.getMinutes())}`;
};

// 11. 轉成 POS API 日期時間
const toAPIDateTime = (value: string): string =>
  value ? `${value.replace('T', ' ')}:00` : `${nowLocalDateTime().replace('T', ' ')}:00`;

// 12. 讀取 API 錯誤
const readPaymentAPIError = (error: unknown): PaymentAPIError['response'] =>
  (error as PaymentAPIError).response;

// 13. 讀取手續費率
const normalizeFeeRate = (value: unknown): number => {
  const raw = String(value ?? '').trim();
  if (!raw) {
    return 0;
  }
  const numeric = Number(raw.replace(/[％%]/g, '').trim());
  if (!Number.isFinite(numeric) || numeric < 0) {
    return 0;
  }
  return raw.includes('%') || raw.includes('％') || numeric > 1 ? numeric / 100 : numeric;
};

// 14. 格式化費率
const formatFeeRate = (feeRate: number): string =>
  new Intl.NumberFormat(preferenceStore.locale, {
    style: 'percent',
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(feeRate);

// 14.1 格式化整數
const formatInteger = (value: number): string =>
  new Intl.NumberFormat(preferenceStore.locale, { maximumFractionDigits: 0 }).format(value);

// 15. 計算單筆手續費
const calculateFeeAmount = (amount: number, feeRate: number): number => {
  if (!Number.isFinite(amount) || amount <= 0 || !Number.isFinite(feeRate) || feeRate <= 0) {
    return 0;
  }
  return Math.ceil(amount * feeRate * 100 - 1e-8) / 100;
};

// 16. 讀取瀏覽器裝置資訊
const readBrowserUserAgent = (): string =>
  typeof navigator === 'undefined' ? '' : navigator.userAgent.toLowerCase();

// 17. 判斷目前是否優先展示 QR 付款
const prefersQRPayment = (): boolean => {
  const userAgent = readBrowserUserAgent();
  const isMobile = /(iphone|ipod|windows phone|iemobile|blackberry|bb10|opera mini|mobile)/i.test(userAgent);
  const isAndroidTablet = userAgent.includes('android') && !userAgent.includes('mobile');
  const isTablet = userAgent.includes('ipad') || /(tablet|playbook|silk)/i.test(userAgent) || isAndroidTablet;
  return !isMobile && !isTablet;
};

// 18. 建立預設支付方式
const fallbackMethods = (): CheckoutMethod[] => [
  { key: 'POS_WECHAT', labelKey: 'account.payments.methods.wechat', shortLabel: 'WX', category: 'online', payType: 'POS_WECHAT', variant: 'wechat', feeRate: 0.027 },
  { key: 'POS_ALIPAY_CN', labelKey: 'account.payments.methods.mainlandAlipay', shortLabel: 'AP', category: 'online', payType: 'POS_ALIPAY', variant: 'alipay', walletType: 'CN', feeRate: 0.027 },
  { key: 'POS_ALIPAY_HK', labelKey: 'account.payments.methods.alipayHK', shortLabel: 'HK', category: 'online', payType: 'POS_ALIPAY', variant: 'alipay', walletType: 'HK', feeRate: 0.027 },
  { key: 'POS_YSF_QR', labelKey: 'account.payments.methods.unionpay', shortLabel: 'YF', category: 'online', payType: 'POS_YSF_QR', variant: 'ysf', feeRate: 0.027 },
  { key: 'POS_BANK', labelKey: 'account.payments.methods.bankTransfer', shortLabel: 'BT', category: 'offline', payType: 'POS_BANK', feeRate: 0 },
  { key: 'POS_CHEQUE', labelKey: 'account.payments.methods.cheque', shortLabel: 'CQ', category: 'offline', payType: 'POS_CHEQUE', feeRate: 0 },
];

// 19. 按 POS pay_type 映射支付方式
const mapFeeRowsToMethods = (feeRows: POSPaymentRow[]): CheckoutMethod[] => {
  const mapped: CheckoutMethod[] = [];
  feeRows.forEach((row) => {
    const code = readPaymentText(row, ['pay_type', 'method']).toUpperCase();
    const feeRate = normalizeFeeRate(row.markup);
    if (code === 'POS_ALIWE' || code === 'WEBPOS_WECHAT' || code === 'WEBPOS_ALIPAY') {
      if (code !== 'WEBPOS_ALIPAY') {
        mapped.push({ key: 'POS_WECHAT', labelKey: 'account.payments.methods.wechat', shortLabel: 'WX', category: 'online', payType: 'POS_WECHAT', variant: 'wechat', feeRate });
      }
      if (code !== 'WEBPOS_WECHAT') {
        mapped.push(
          { key: 'POS_ALIPAY_CN', labelKey: 'account.payments.methods.mainlandAlipay', shortLabel: 'AP', category: 'online', payType: 'POS_ALIPAY', variant: 'alipay', walletType: 'CN', feeRate },
          { key: 'POS_ALIPAY_HK', labelKey: 'account.payments.methods.alipayHK', shortLabel: 'HK', category: 'online', payType: 'POS_ALIPAY', variant: 'alipay', walletType: 'HK', feeRate },
        );
      }
    }
    if (code === 'POS_YSF_QR' || code === 'WEBPOS_YSF') {
      mapped.push({ key: 'POS_YSF_QR', labelKey: 'account.payments.methods.unionpay', shortLabel: 'YF', category: 'online', payType: 'POS_YSF_QR', variant: 'ysf', feeRate });
    }
    if (code === 'POS_BANK' || code === 'WEBPOS_BANK') {
      mapped.push({ key: 'POS_BANK', labelKey: 'account.payments.methods.bankTransfer', shortLabel: 'BT', category: 'offline', payType: 'POS_BANK', feeRate: 0 });
    }
    if (code === 'POS_CHEQUE' || code === 'WEBPOS_CHEQUE') {
      mapped.push({ key: 'POS_CHEQUE', labelKey: 'account.payments.methods.cheque', shortLabel: 'CQ', category: 'offline', payType: 'POS_CHEQUE', feeRate: 0 });
    }
  });

  const order: CheckoutMethodKey[] = ['POS_WECHAT', 'POS_ALIPAY_CN', 'POS_ALIPAY_HK', 'POS_YSF_QR', 'POS_BANK', 'POS_CHEQUE'];
  const seen = new Set<CheckoutMethodKey>();
  return mapped
    .filter((item) => {
      if (seen.has(item.key)) {
        return false;
      }
      seen.add(item.key);
      return true;
    })
    .sort((left, right) => order.indexOf(left.key) - order.indexOf(right.key));
};

// 20. 解析線上支付通道
const resolveOnlineChannel = (method: CheckoutMethod): POSPaymentChannel => {
  if (method.variant === 'wechat') {
    return prefersQRPayment() ? 'WX_QR' : 'WX_H5';
  }
  if (method.variant === 'alipay') {
    return prefersQRPayment() && !method.walletType ? 'ALI_QR' : 'ALI_H5';
  }
  if (method.variant === 'ysf') {
    return 'YSF_QR';
  }
  return 'WX_H5';
};

// 21. 選取或取消賬單
const toggleBill = (row: POSPaymentRow, index: number): void => {
  const key = billKey(row, index);
  if (selectedBillKeySet.value.has(key)) {
    selectedBillKeys.value = selectedBillKeys.value.filter((item) => item !== key);
    return;
  }
  selectedBillKeys.value = [...selectedBillKeys.value, key];
};

// 22. 切換全部賬單選取狀態
const toggleAllBills = (): void => {
  selectedBillKeys.value = allPayableBillsSelected.value
    ? []
    : payableRows.value.map((row, index) => billKey(row, index));
};

// 23. 載入支付設定
const loadPaymentSettings = async (): Promise<void> => {
  if (!hasContext.value) {
    return;
  }
  isSettingsLoading.value = true;
  try {
    const [feeResponse, bankResponse, orderResponse] = await Promise.all([
      fetchPOSPaymentFees(paymentQuery.value),
      fetchPOSPaymentBankAccounts(paymentQuery.value),
      fetchPOSPaymentOrders(paymentQuery.value),
    ]);
    const mappedMethods = mapFeeRowsToMethods(feeResponse.data.data.items ?? []);
    methods.value = mappedMethods.length > 0 ? mappedMethods : fallbackMethods();
    bankAccounts.value = (bankResponse.data.data.items ?? [])
      .map((item) => ({
        label: readPaymentText(item, ['title', 'account_no', 'bank_name']) || t('account.payments.checkout.bankAccountFallback'),
        value: readPaymentText(item, ['account_no', 'id']),
      }))
      .filter((item) => item.value !== '');
    orders.value = orderResponse.data.data.items ?? [];
  } catch {
    methods.value = fallbackMethods();
    bankAccounts.value = [];
    orders.value = [];
  } finally {
    selectedMethodKey.value = methods.value.some((item) => item.key === selectedMethodKey.value)
      ? selectedMethodKey.value
      : methods.value[0]?.key ?? 'POS_WECHAT';
    offlineForm.bankAccountNo = bankAccounts.value[0]?.value ?? '';
    isSettingsLoading.value = false;
  }
};

// 24. 載入待繳賬單
const loadBills = async (): Promise<void> => {
  isLoading.value = true;
  profileRequired.value = false;
  posLoginRequired.value = false;
  try {
    const response = await fetchPOSPaymentBills();
    rows.value = response.data.data.items ?? [];
    context.value = response.data.data.context;
    const availableKeys = payableRows.value.map((row, index) => billKey(row, index));
    selectedBillKeys.value = selectedBillKeys.value.filter((key) => availableKeys.includes(key));
    if (selectedBillKeys.value.length === 0 && availableKeys.length > 0) {
      selectedBillKeys.value = [availableKeys[0]];
    }
  } catch (error: unknown) {
    rows.value = [];
    context.value = undefined;
    const response = readPaymentAPIError(error);
    const message = String(response?.data?.message ?? '').toLowerCase();
    profileRequired.value = response?.status === 400 || message.includes('profile') || message.includes('residence');
    posLoginRequired.value = response?.status === 401 || message.includes('ismart');
    if (!profileRequired.value && !posLoginRequired.value) {
      feedbackStore.pushToast(t('account.payments.checkout.billLoadError'), 'error');
    }
  } finally {
    isLoading.value = false;
  }
};

// 25. 提交前重新校驗賬單
const validateSelectedBills = async (): Promise<boolean> => {
  const response = await fetchPOSPaymentBills(paymentQuery.value);
  const latestRows = response.data.data.items ?? [];
  const latestAmountMap = new Map<string, number>();
  latestRows.forEach((row) => {
    latestAmountMap.set(billIdentityKey(row), readPaymentAmountValue(row));
  });

  for (const row of selectedRows.value) {
    const latestAmount = latestAmountMap.get(billIdentityKey(row));
    if (latestAmount === undefined) {
      feedbackStore.pushToast(t('account.payments.checkout.billChanged'), 'error');
      return false;
    }
    if (readPaymentAmountValue(row) - latestAmount > 0.01) {
      feedbackStore.pushToast(t('account.payments.checkout.billAmountChanged'), 'error');
      return false;
    }
  }
  return true;
};

// 26. 組裝 POS 賬單資料
const buildBillObjects = (feeRate: number): POSPaymentRow[] => selectedRows.value.map((row) => {
  const amountCents = Math.round(readPaymentAmountValue(row) * 100);
  const feeCents = Math.round(calculateFeeAmount(readPaymentAmountValue(row), feeRate) * 100);
  return {
    flat_code: readPaymentText(row, ['flat_code', 'unit', 'unit_name']),
    unit_name: readPaymentText(row, ['unit_name', 'unit', 'flat_code']),
    item_id: readPaymentText(row, ['item_name', 'item_id', 'name']) || t('account.payments.checkout.defaultBillItem'),
    trs_to: readPaymentText(row, ['trs_to', 'term', 'period']),
    bill_dt: readBillDate(row),
    net_amount: amountCents,
    invoice_no: readPaymentText(row, ['invoice_no', 'bill_no', 'bill_number']),
    transfer_fee: feeCents,
  };
});

// 27. 組裝手續費資料
const buildFeeObjects = (feeRate: number): POSPaymentRow[] => selectedRows.value.map((row) => ({
  invoice_no: readPaymentText(row, ['invoice_no', 'bill_no', 'bill_number']),
  flat_code: readPaymentText(row, ['flat_code', 'unit', 'unit_name']),
  handling_fee: Math.round(calculateFeeAmount(readPaymentAmountValue(row), feeRate) * 100),
}));

// 28. 讀取訂單資料
const readOrderData = (payload: Record<string, unknown>): POSPaymentRow => {
  const nested = payload.data;
  if (nested && typeof nested === 'object' && !Array.isArray(nested)) {
    return nested as POSPaymentRow;
  }
  return payload;
};

// 29. 讀取訂單 key
const readOrderKey = (row: POSPaymentRow): string =>
  readPaymentText(row, ['mch_order_no', 'mchOrderNo', 'pay_order_id', 'payOrderId', 'order_no', 'id']);

// 30. 讀取訂單金額
const readOrderAmount = (row: POSPaymentRow): string => {
  const direct = Number(readPaymentText(row, ['amount_hkd', 'paid_amount_hkd']));
  if (Number.isFinite(direct) && direct > 0) {
    return formatHKD(direct);
  }
  const cents = Number(readPaymentText(row, ['final_amount', 'amount', 'paid_amount', 'total_amount']));
  if (Number.isFinite(cents) && cents > 0) {
    return formatHKD(cents / 100);
  }
  return '-';
};

// 31. 讀取訂單狀態
const readOrderStatus = (row: POSPaymentRow): string => {
  const state = readPaymentText(row, ['state', 'status', 'gateway_state_code', 'business_state']).toLowerCase();
  if (['success', 'succeeded', 'paid', '2'].includes(state)) {
    return t('account.payments.checkout.statusSuccess');
  }
  if (['failed', 'expired', 'revoked'].includes(state)) {
    return t('account.payments.checkout.statusFailed');
  }
  if (['closed', 'cancelled', 'canceled'].includes(state)) {
    return t('account.payments.checkout.statusClosed');
  }
  return state
    ? t('account.payments.checkout.statusProcessing')
    : t('account.payments.checkout.statusPending');
};

// 32. 讀取圖片
const readFileAsDataURL = (file: File): Promise<string> =>
  new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => resolve(String(reader.result ?? ''));
    reader.onerror = () => reject(new Error('voucher image read failed'));
    reader.readAsDataURL(file);
  });

// 33. 選擇憑證圖片
const onVoucherChange = async (event: Event): Promise<void> => {
  const input = event.target as HTMLInputElement;
  const files = Array.from(input.files ?? []);
  if (files.length === 0) {
    return;
  }
  try {
    const mapped = await Promise.all(files.map(async (file) => ({
      name: file.name,
      data: await readFileAsDataURL(file),
      previewUrl: URL.createObjectURL(file),
    })));
    voucherImages.value = [...voucherImages.value, ...mapped];
  } catch {
    feedbackStore.pushToast(t('account.payments.checkout.voucherReadError'), 'error');
  } finally {
    input.value = '';
  }
};

// 34. 移除憑證圖片
const removeVoucher = (index: number): void => {
  const target = voucherImages.value[index];
  if (!target) {
    return;
  }
  URL.revokeObjectURL(target.previewUrl);
  voucherImages.value.splice(index, 1);
};

// 35. 提交線上支付
const submitOnline = async (method: CheckoutMethod): Promise<void> => {
  const payChannel = resolveOnlineChannel(method);
  const response = await createPOSPaymentOrder({
    scene: 'billing',
    building_id: activeBuildingID.value,
    unit_id: activeUnitID.value,
    pay_channel: payChannel,
    expire_seconds: posPaymentExpireSeconds,
    final_amount: Math.round(finalAmount.value * 100),
    handle_fee_amount: Math.round(feeAmount.value * 100),
    bill_objs: buildBillObjects(method.feeRate),
    handle_fee_obj: buildFeeObjects(method.feeRate),
    return_path: `${window.location.origin}/payments/pay`,
    remark: offlineForm.remark,
    ...(method.walletType ? { gateway_request_overrides: { walletType: method.walletType } } : {}),
  });
  resultOrder.value = readOrderData(response.data.data);
  resultReceipt.value = '';
  resultMessage.value = t('account.payments.checkout.orderCreated');
  await loadPaymentSettings();
};

// 36. 提交線下繳費
const submitOffline = async (method: CheckoutMethod): Promise<void> => {
  const entryDateTime = toAPIDateTime(nowLocalDateTime());
  const response = await reportPOSPayment({
    building_id: activeBuildingID.value,
    unit_id: activeUnitID.value,
    FINAL_AMOUNT: Math.round(finalAmount.value * 100),
    ENTRY_DATETIME: entryDateTime,
    TRAN_DATETIME: toAPIDateTime(offlineForm.tranDateTime),
    TRAN_REF_NO: offlineForm.tranRefNo || `${method.key}-${Date.now()}`,
    COMMENT: offlineForm.remark,
    PIC_FILENAME: voucherImages.value.map((item) => item.name),
    PIC_DATA: voucherImages.value.map((item) => item.data),
    BILL_OBJS: buildBillObjects(0),
    PAY_METHOD: method.payType,
    BLG_ID: activeBuildingID.value,
    UNIT_ID: activeUnitID.value,
    bank_account_received: requiresBankAccount.value ? offlineForm.bankAccountNo : '',
  });
  const result = response.data.data;
  resultReceipt.value = readPaymentText(result, ['receipt_id', 'receipt_no']) ||
    readPaymentText((result.ismart_receipt_no as POSPaymentRow | undefined) ?? {}, ['receipt_id', 'receipt_no']);
  resultOrder.value = null;
  resultMessage.value = resultReceipt.value
    ? t('account.payments.checkout.paymentReportedWithReceipt', { receipt: resultReceipt.value })
    : t('account.payments.checkout.paymentReported');
  voucherImages.value.forEach((item) => URL.revokeObjectURL(item.previewUrl));
  voucherImages.value = [];
  selectedBillKeys.value = [];
  await loadBills();
};

// 37. 提交付款
const submitPayment = async (): Promise<void> => {
  const method = selectedMethod.value;
  if (!method || !canSubmit.value) {
    return;
  }
  isSubmitting.value = true;
  try {
    const isValid = await validateSelectedBills();
    if (!isValid) {
      return;
    }
    if (method.category === 'online') {
      await submitOnline(method);
    } else {
      await submitOffline(method);
    }
    feedbackStore.pushToast(t('account.payments.checkout.submitSuccess'), 'success');
  } catch {
    feedbackStore.pushToast(t('account.payments.checkout.submitError'), 'error');
  } finally {
    isSubmitting.value = false;
  }
};

// 38. 載入路由指定訂單
const loadRouteOrder = async (): Promise<void> => {
  const mchOrderNo = String(route.query.mch_order_no ?? route.query.mchOrderNo ?? '').trim();
  const payOrderId = String(route.query.pay_order_id ?? route.query.payOrderId ?? '').trim();
  if (!hasContext.value || (!mchOrderNo && !payOrderId)) {
    return;
  }
  try {
    const payload = mchOrderNo
      ? {
          mch_order_no: mchOrderNo,
          building_id: activeBuildingID.value,
          unit_id: activeUnitID.value,
          retry_business: true,
          real_gateway: true,
        }
      : {
          pay_order_id: payOrderId,
          building_id: activeBuildingID.value,
          unit_id: activeUnitID.value,
          retry_business: true,
          real_gateway: true,
        };
    resultOrder.value = readOrderData((await queryPOSPaymentOrder(payload)).data.data);
    resultMessage.value = t('account.payments.checkout.orderLoaded');
  } catch {
    feedbackStore.pushToast(t('account.payments.checkout.orderQueryError'), 'error');
  }
};

// 39. 刷新結果訂單
const refreshResultOrder = async (): Promise<void> => {
  if (!resultOrder.value) {
    return;
  }
  const mchOrderNo = readPaymentText(resultOrder.value, ['mch_order_no', 'mchOrderNo']);
  const payOrderId = readPaymentText(resultOrder.value, ['pay_order_id', 'payOrderId']);
  if (!mchOrderNo && !payOrderId) {
    return;
  }
  isRefreshingOrder.value = true;
  try {
    resultOrder.value = mchOrderNo
      ? readOrderData((await fetchPOSPaymentOrder(mchOrderNo, { ...paymentQuery.value, detail: true, refresh: true })).data.data)
      : readOrderData((await queryPOSPaymentOrder({ ...paymentQuery.value, pay_order_id: payOrderId, retry_business: true, real_gateway: true })).data.data);
    await loadPaymentSettings();
  } catch {
    feedbackStore.pushToast(t('account.payments.checkout.orderRefreshError'), 'error');
  } finally {
    isRefreshingOrder.value = false;
  }
};

// 40. 返回會員中心
const goProfile = async (): Promise<void> => {
  await router.push('/account/profile');
};

onMounted(async () => {
  offlineForm.tranDateTime = nowLocalDateTime();
  await loadBills();
  await loadPaymentSettings();
  await loadRouteOrder();
});

watch(hasContext, (ready) => {
  if (ready) {
    void loadPaymentSettings();
    void loadRouteOrder();
  }
});
</script>

<template>
  <main class="ajo-pay-checkout">
    <AppBreadcrumb :items="breadcrumbItems" />

    <section class="ajo-pay-checkout__header">
      <h1>{{ t('account.payments.checkout.title') }}</h1>
      <p>{{ contextLabel }}</p>

      <div class="ajo-pay-checkout__steps">
        <div
          v-for="step in [1, 2, 3, 4]"
          :key="step"
          class="ajo-pay-checkout__step"
          :class="{ 'ajo-pay-checkout__step--active': activeStepIndex >= step }"
        >
          <span>{{ step }}</span>
          <strong>
            {{
              step === 1
                ? t('account.payments.checkout.stepItems')
                : step === 2
                  ? t('account.payments.checkout.stepMethod')
                  : step === 3
                    ? t('account.payments.checkout.stepConfirm')
                    : t('account.payments.checkout.stepOrder')
            }}
          </strong>
        </div>
      </div>
    </section>

    <section
      v-if="profileRequired || posLoginRequired"
      class="ajo-pay-checkout__notice"
    >
      <p>
        {{
          profileRequired
            ? t('account.payments.checkout.bindUnitNotice')
            : t('account.payments.checkout.bindIsmartNotice')
        }}
      </p>
      <button
        type="button"
        @click="goProfile"
      >
        {{ t('account.payments.checkout.accountAction') }}
      </button>
    </section>

    <div class="ajo-pay-checkout__layout">
      <section class="ajo-pay-checkout__section">
        <div class="ajo-pay-checkout__section-head">
          <span>1</span>
          <strong>{{ t('account.payments.checkout.paymentItems') }}</strong>
        </div>
        <div class="ajo-pay-checkout__card">
          <template v-if="payableRows.length > 0">
            <button
              v-for="(row, index) in payableRows"
              :key="billKey(row, index)"
              type="button"
              class="ajo-pay-checkout__bill"
              :class="{ 'ajo-pay-checkout__bill--selected': selectedBillKeySet.has(billKey(row, index)) }"
              @click="toggleBill(row, index)"
            >
              <span class="ajo-pay-checkout__bill-icon">
                {{
                  isManagementBill(row)
                    ? t('account.payments.checkout.managementIcon')
                    : t('account.payments.checkout.feeIcon')
                }}
              </span>
              <span class="ajo-pay-checkout__bill-copy">
                <strong>{{ readBillName(row) }}</strong>
                <small>{{ readBillTitle(row) }} · {{ readBillDate(row) }}</small>
              </span>
              <b>{{ formatHKD(readPaymentAmountValue(row)) }}</b>
              <i>✓</i>
            </button>
          </template>
          <div
            v-else
            class="ajo-pay-checkout__empty"
          >
            {{ isLoading ? t('account.payments.checkout.loading') : t('account.payments.checkout.noBills') }}
          </div>
          <div class="ajo-pay-checkout__total">
            <span>{{ t('account.payments.checkout.total') }}</span>
            <strong>{{ formatHKD(selectedAmount) }}</strong>
          </div>
          <button
            v-if="payableRows.length > 1"
            type="button"
            class="ajo-pay-checkout__link-button"
            @click="toggleAllBills"
          >
            {{
              allPayableBillsSelected
                ? t('account.payments.checkout.deselectAll')
                : t('account.payments.checkout.selectAll')
            }}
          </button>
        </div>
      </section>

      <section class="ajo-pay-checkout__section">
        <div class="ajo-pay-checkout__section-head">
          <span>2</span>
          <strong>{{ t('account.payments.checkout.paymentMethod') }}</strong>
        </div>
        <button
          v-for="method in methods"
          :key="method.key"
          type="button"
          class="ajo-pay-checkout__method"
          :class="{ 'ajo-pay-checkout__method--selected': method.key === selectedMethodKey }"
          :disabled="isSettingsLoading"
          @click="selectedMethodKey = method.key"
        >
          <span class="ajo-pay-checkout__radio">
            <i v-if="method.key === selectedMethodKey"></i>
          </span>
          <b :class="{ 'ajo-pay-checkout__method-logo--fps': method.key === 'POS_YSF_QR' }">
            {{ method.shortLabel }}
          </b>
          <strong>{{ t(method.labelKey) }}</strong>
          <small>{{ t('account.payments.checkout.feeRate', { rate: formatFeeRate(method.feeRate) }) }}</small>
        </button>
      </section>

      <section class="ajo-pay-checkout__section">
        <div class="ajo-pay-checkout__section-head">
          <span>3</span>
          <strong>{{ t('account.payments.checkout.confirmPayment') }}</strong>
        </div>
        <div class="ajo-pay-checkout__coin-box">
          <div>
            <span>{{ t('account.payments.checkout.rewardRate') }}</span>
            <strong>{{ t('account.payments.checkout.rewardProgram') }}</strong>
            <small>{{ t('account.payments.checkout.estimatedReward') }}</small>
            <b>{{ t('account.payments.checkout.rewardCoins', { count: formatInteger(rewardCoins) }) }}</b>
          </div>
          <i>A</i>
        </div>

        <div class="ajo-pay-checkout__summary-box">
          <div>
            <span>{{ t('account.payments.checkout.managementFee') }}</span>
            <strong>{{ formatHKD(managementAmount) }}</strong>
          </div>
          <div>
            <span>{{ t('account.payments.checkout.otherFees') }}</span>
            <strong>{{ formatHKD(otherAmount) }}</strong>
          </div>
          <div>
            <span>{{ t('account.payments.checkout.handlingFee') }}</span>
            <strong>{{ formatHKD(feeAmount) }}</strong>
          </div>
        </div>

        <div
          v-if="selectedMethod?.category === 'offline'"
          class="ajo-pay-checkout__offline"
        >
          <label>
            <span>{{ t('account.payments.checkout.transactionTime') }}</span>
            <input
              v-model="offlineForm.tranDateTime"
              type="datetime-local"
            >
          </label>
          <label>
            <span>
              {{
                isChequeMethod
                  ? t('account.payments.checkout.chequeNumber')
                  : t('account.payments.checkout.transactionReference')
              }}
            </span>
            <input
              v-model="offlineForm.tranRefNo"
              type="text"
              :placeholder="isChequeMethod
                ? t('account.payments.checkout.chequePlaceholder')
                : t('account.payments.checkout.referencePlaceholder')"
            >
          </label>
          <label v-if="requiresBankAccount">
            <span>{{ t('account.payments.checkout.bankAccount') }}</span>
            <select v-model="offlineForm.bankAccountNo">
              <option
                v-for="account in bankAccounts"
                :key="account.value"
                :value="account.value"
              >
                {{ account.label }}
              </option>
            </select>
          </label>
          <label>
            <span>{{ t('account.payments.checkout.remark') }}</span>
            <textarea
              v-model="offlineForm.remark"
              rows="3"
            ></textarea>
          </label>
          <div
            v-if="isBankTransferMethod"
            class="ajo-pay-checkout__voucher"
          >
            <input
              ref="voucherInput"
              type="file"
              accept="image/*"
              multiple
              @change="onVoucherChange"
            >
            <button
              type="button"
              @click="voucherInput?.click()"
            >
              {{ t('account.payments.checkout.uploadVoucher') }}
            </button>
            <span>
              {{
                voucherImages.length === 0
                  ? t('account.payments.checkout.noVoucher')
                  : t('account.payments.checkout.voucherCount', { count: voucherImages.length })
              }}
            </span>
          </div>
          <div
            v-if="voucherImages.length > 0"
            class="ajo-pay-checkout__voucher-list"
          >
            <article
              v-for="(image, index) in voucherImages"
              :key="image.previewUrl"
            >
              <img
                :src="image.previewUrl"
                :alt="image.name"
              >
              <span>{{ image.name }}</span>
              <button
                type="button"
                @click="removeVoucher(index)"
              >
                {{ t('account.payments.checkout.removeVoucher') }}
              </button>
            </article>
          </div>
        </div>

        <button
          type="button"
          class="ajo-pay-checkout__confirm"
          :disabled="!canSubmit"
          @click="submitPayment"
        >
          {{ confirmButtonLabel }}
        </button>
        <p class="ajo-pay-checkout__hint">
          {{ t('account.payments.checkout.submitHint') }}
        </p>
      </section>

      <section class="ajo-pay-checkout__section">
        <div class="ajo-pay-checkout__section-head">
          <span>4</span>
          <strong>{{ t('account.payments.checkout.order') }}</strong>
        </div>
        <div class="ajo-pay-checkout__complete-note">
          {{ resultMessage || t('account.payments.checkout.orderIntro') }}
        </div>

        <article
          v-if="resultOrder"
          class="ajo-pay-checkout__order-result"
        >
          <div>
            <span>{{ t('account.payments.checkout.paymentOrder') }}</span>
            <strong>{{ readOrderKey(resultOrder) || t('account.payments.checkout.paymentOrder') }}</strong>
            <small>{{ readOrderStatus(resultOrder) }} · {{ readOrderAmount(resultOrder) }}</small>
          </div>
          <button
            type="button"
            :disabled="isRefreshingOrder"
            @click="refreshResultOrder"
          >
            {{
              isRefreshingOrder
                ? t('account.payments.checkout.refreshing')
                : t('account.payments.checkout.refreshStatus')
            }}
          </button>
        </article>

        <div
          v-if="resultQrImageUrl || resultQrText"
          class="ajo-pay-checkout__qr"
        >
          <img
            v-if="resultQrImageUrl"
            :src="resultQrImageUrl"
            :alt="t('account.payments.checkout.qrAlt')"
          >
          <QrCodeImage
            v-else
            :text="resultQrText"
            :alt="t('account.payments.checkout.qrAlt')"
            :size="220"
          />
          <a
            v-if="resultPayDataType === 'payurl' && resultPayData"
            :href="resultPayData"
            target="_blank"
            rel="noreferrer"
          >
            {{ t('account.payments.checkout.openPaymentLink') }}
          </a>
        </div>

        <div
          v-if="resultReceipt"
          class="ajo-pay-checkout__receipt"
        >
          <strong>{{ t('account.payments.checkout.submitted') }}</strong>
          <span>{{ resultReceipt }}</span>
        </div>

        <div class="ajo-pay-checkout__record-title">{{ t('account.payments.checkout.paymentOrder') }}</div>
        <div class="ajo-pay-checkout__records">
          <article
            v-for="order in displayedOrders"
            :key="readOrderKey(order)"
            class="ajo-pay-checkout__record"
          >
            <span>{{ readOrderKey(order) || t('account.payments.checkout.paymentOrder') }}</span>
            <strong>{{ readOrderAmount(order) }}</strong>
            <small>{{ readOrderStatus(order) }}</small>
          </article>
          <div
            v-if="displayedOrders.length === 0"
            class="ajo-pay-checkout__empty"
          >
            {{ t('account.payments.checkout.noOrders') }}
          </div>
        </div>
      </section>
    </div>
  </main>
</template>

<style scoped>
.ajo-pay-checkout {
  --pay-brand: rgb(var(--color-primary));
  --pay-brand-dark: rgb(var(--color-brand-dark));
  --pay-brand-light: rgb(var(--color-primary-soft));
  --pay-brand-mid: rgb(var(--color-brand-mid));
  --pay-ink: rgb(var(--color-text));
  --pay-ink-2: rgb(var(--color-ink-2));
  --pay-ink-3: rgb(var(--color-ink-3));
  --pay-surface: rgb(var(--color-surface));
  --pay-surface-2: rgb(var(--color-surface-2));
  --pay-border: rgb(var(--color-border));
  --pay-border-2: rgb(var(--color-border-2));
  --pay-info-surface: rgb(var(--color-surface-3));
  --pay-success: rgb(var(--color-success));
  display: grid;
  gap: 0;
  min-height: calc(100svh - var(--nav-h, 52px));
  background: var(--pay-surface-2);
  color: var(--pay-ink);
  padding: 40px;
}

.ajo-pay-checkout__header {
  max-width: 760px;
  margin: 0 auto;
  text-align: center;
  width: 100%;
}

.ajo-pay-checkout__header h1 {
  margin: 0 0 6px;
  color: var(--pay-ink);
  font-family: 'Noto Sans TC', var(--font-sans);
  font-size: 24px;
  font-weight: 700;
  line-height: 1.3;
}

.ajo-pay-checkout__header p {
  margin: 0 0 28px;
  color: var(--pay-ink-3);
  font-size: 13px;
  line-height: 1.5;
}

.ajo-pay-checkout__steps {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
  margin-bottom: 28px;
}

.ajo-pay-checkout__step {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  color: var(--pay-ink-3);
  font-size: 12px;
  font-weight: 600;
  text-align: center;
}

.ajo-pay-checkout__step:not(:last-child)::after {
  position: absolute;
  top: 15px;
  left: calc(50% + 24px);
  right: calc(-50% + 24px);
  height: 1px;
  background: var(--pay-border-2);
  content: '';
}

.ajo-pay-checkout__step span {
  position: relative;
  z-index: 1;
  display: grid;
  width: 30px;
  height: 30px;
  place-items: center;
  border: 1px solid var(--pay-border-2);
  border-radius: 50%;
  background: var(--pay-surface);
  color: var(--pay-ink-3);
  font-size: 13px;
  font-weight: 800;
}

.ajo-pay-checkout__step strong {
  font-weight: 600;
}

.ajo-pay-checkout__step--active {
  color: var(--pay-brand);
}

.ajo-pay-checkout__step--active span {
  border-color: var(--pay-brand);
  background: var(--pay-brand);
  color: #ffffff;
}

.ajo-pay-checkout__notice {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  max-width: 760px;
  width: 100%;
  margin: 0 auto 18px;
  border: 0.5px solid var(--pay-brand-mid);
  border-radius: 9px;
  background: var(--pay-brand-light);
  padding: 13px 16px;
}

.ajo-pay-checkout__notice p {
  margin: 0;
  color: var(--pay-ink-2);
  font-size: 13px;
  line-height: 1.5;
}

.ajo-pay-checkout__notice button,
.ajo-pay-checkout__link-button,
.ajo-pay-checkout__order-result button,
.ajo-pay-checkout__voucher button,
.ajo-pay-checkout__voucher-list button {
  border: 0;
  border-radius: 9px;
  background: var(--pay-brand);
  color: #ffffff;
  cursor: pointer;
  font-family: inherit;
  font-size: 12px;
  font-weight: 700;
  padding: 9px 14px;
}

.ajo-pay-checkout__layout {
  display: flex;
  flex-direction: column;
  gap: 18px;
  max-width: 760px;
  width: 100%;
  margin: 0 auto;
}

.ajo-pay-checkout__section {
  border: 0.5px solid var(--pay-border);
  border-radius: 14px;
  background: var(--pay-surface);
  padding: 20px;
}

.ajo-pay-checkout__section-head {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 14px;
}

.ajo-pay-checkout__section-head span {
  display: grid;
  width: 30px;
  height: 30px;
  place-items: center;
  border-radius: 50%;
  background: var(--pay-brand);
  color: #ffffff;
  font-size: 13px;
  font-weight: 800;
}

.ajo-pay-checkout__section-head strong,
.ajo-pay-checkout__record-title {
  color: var(--pay-ink-3);
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 0.08em;
  line-height: 1.3;
  text-transform: uppercase;
}

.ajo-pay-checkout__card {
  overflow: hidden;
  border: 0.5px solid var(--pay-border);
  border-radius: 14px;
  background: var(--pay-surface);
}

.ajo-pay-checkout__bill {
  display: flex;
  width: 100%;
  align-items: center;
  gap: 14px;
  border: 0;
  border-left: 3px solid transparent;
  background: var(--pay-surface);
  cursor: pointer;
  filter: grayscale(1);
  font-family: inherit;
  opacity: 0.48;
  padding: 16px 20px;
  text-align: left;
  transition: background 0.12s ease, opacity 0.12s ease;
}

.ajo-pay-checkout__bill + .ajo-pay-checkout__bill {
  border-top: 0.5px solid var(--pay-border);
}

.ajo-pay-checkout__bill--selected {
  border-left-color: var(--pay-brand);
  background: var(--pay-brand-light);
  filter: none;
  opacity: 1;
}

.ajo-pay-checkout__bill-icon {
  display: grid;
  width: 42px;
  height: 42px;
  flex: 0 0 auto;
  place-items: center;
  border-radius: 11px;
  background: var(--pay-info-surface);
  color: var(--pay-ink);
  font-size: 18px;
  font-weight: 800;
}

.ajo-pay-checkout__bill-copy {
  display: grid;
  flex: 1;
  min-width: 0;
  gap: 2px;
}

.ajo-pay-checkout__bill-copy strong {
  overflow: hidden;
  color: var(--pay-ink);
  font-size: 14px;
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ajo-pay-checkout__bill-copy small {
  overflow: hidden;
  color: var(--pay-ink-3);
  font-size: 12px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ajo-pay-checkout__bill b {
  color: var(--pay-ink);
  font-size: 14px;
  font-weight: 700;
  white-space: nowrap;
}

.ajo-pay-checkout__bill i {
  display: grid;
  width: 24px;
  height: 24px;
  flex: 0 0 auto;
  place-items: center;
  border: 0.5px solid var(--pay-border);
  border-radius: 7px;
  background: var(--pay-surface-2);
  color: transparent;
  font-style: normal;
  font-size: 13px;
  font-weight: 800;
}

.ajo-pay-checkout__bill--selected i {
  border-color: var(--pay-brand-mid);
  background: var(--pay-brand-light);
  color: var(--pay-brand);
}

.ajo-pay-checkout__empty {
  padding: 18px 20px;
  color: var(--pay-ink-3);
  font-size: 13px;
  line-height: 1.6;
  text-align: center;
}

.ajo-pay-checkout__total {
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-top: 0.5px solid var(--pay-border-2);
  padding: 16px 20px;
}

.ajo-pay-checkout__total span {
  color: var(--pay-ink);
  font-size: 15px;
  font-weight: 600;
}

.ajo-pay-checkout__total strong {
  color: var(--pay-ink);
  font-size: 22px;
  font-weight: 800;
}

.ajo-pay-checkout__link-button {
  width: calc(100% - 40px);
  margin: 0 20px 16px;
}

.ajo-pay-checkout__method {
  display: flex;
  width: 100%;
  align-items: center;
  gap: 12px;
  border: 1.5px solid var(--pay-border);
  border-radius: 12px;
  background: var(--pay-surface);
  cursor: pointer;
  filter: grayscale(1);
  font-family: inherit;
  margin-bottom: 10px;
  opacity: 0.52;
  padding: 14px 16px;
  text-align: left;
  transition: border-color 0.15s ease, background 0.15s ease, opacity 0.15s ease;
}

.ajo-pay-checkout__method--selected {
  border-color: var(--pay-brand);
  background: var(--pay-brand-light);
  filter: none;
  opacity: 1;
}

.ajo-pay-checkout__method:hover {
  border-color: var(--pay-brand-mid);
}

.ajo-pay-checkout__radio {
  display: grid;
  width: 18px;
  height: 18px;
  flex: 0 0 auto;
  place-items: center;
  border: 1.5px solid var(--pay-border-2);
  border-radius: 50%;
}

.ajo-pay-checkout__method--selected .ajo-pay-checkout__radio {
  border-color: var(--pay-brand);
  background: var(--pay-brand);
}

.ajo-pay-checkout__radio i {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: #ffffff;
}

.ajo-pay-checkout__method b {
  border-radius: 6px;
  background: var(--pay-ink);
  color: #ffffff;
  font-size: 11px;
  font-weight: 800;
  line-height: 1;
  padding: 5px 8px;
}

.ajo-pay-checkout__method b.ajo-pay-checkout__method-logo--fps {
  background: #e31837;
}

.ajo-pay-checkout__method strong {
  flex: 1;
  min-width: 0;
  color: var(--pay-ink);
  font-size: 13px;
  font-weight: 600;
}

.ajo-pay-checkout__method small {
  color: var(--pay-ink-3);
  font-size: 12px;
  font-weight: 700;
  white-space: nowrap;
}

.ajo-pay-checkout__method--selected small {
  color: var(--pay-brand);
}

.ajo-pay-checkout__coin-box {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  border: 0.5px solid var(--pay-brand-mid);
  border-radius: 14px;
  background: var(--pay-brand-light);
  margin-bottom: 16px;
  padding: 18px 20px;
}

.ajo-pay-checkout__coin-box div {
  display: grid;
  gap: 4px;
}

.ajo-pay-checkout__coin-box span {
  width: max-content;
  border-radius: 100px;
  background: var(--pay-brand);
  color: #ffffff;
  font-size: 10px;
  font-weight: 800;
  padding: 3px 9px;
}

.ajo-pay-checkout__coin-box strong {
  color: var(--pay-ink);
  font-size: 14px;
  font-weight: 700;
}

.ajo-pay-checkout__coin-box small {
  color: var(--pay-ink-2);
  font-size: 12px;
}

.ajo-pay-checkout__coin-box b {
  color: var(--pay-brand);
  font-size: 22px;
  font-weight: 800;
}

.ajo-pay-checkout__coin-box i {
  color: var(--pay-brand);
  font-style: normal;
  font-size: 34px;
  font-weight: 800;
}

.ajo-pay-checkout__summary-box {
  display: grid;
  gap: 1px;
  overflow: hidden;
  border: 0.5px solid var(--pay-border);
  border-radius: 9px;
  background: var(--pay-border);
  margin-bottom: 16px;
}

.ajo-pay-checkout__summary-box div,
.ajo-pay-checkout__record,
.ajo-pay-checkout__receipt {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  background: var(--pay-surface);
  padding: 13px 16px;
}

.ajo-pay-checkout__summary-box span,
.ajo-pay-checkout__record span {
  color: var(--pay-ink-2);
  font-size: 13px;
}

.ajo-pay-checkout__summary-box strong,
.ajo-pay-checkout__record strong {
  color: var(--pay-ink);
  font-size: 13px;
  font-weight: 700;
}

.ajo-pay-checkout__offline {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
  border: 0.5px solid var(--pay-border);
  border-radius: 9px;
  background: var(--pay-surface-2);
  margin-bottom: 16px;
  padding: 14px;
}

.ajo-pay-checkout__offline label {
  display: grid;
  gap: 6px;
}

.ajo-pay-checkout__offline label:last-of-type {
  grid-column: 1 / -1;
}

.ajo-pay-checkout__offline span {
  color: var(--pay-ink-3);
  font-size: 11px;
  font-weight: 700;
}

.ajo-pay-checkout__offline input,
.ajo-pay-checkout__offline select,
.ajo-pay-checkout__offline textarea {
  width: 100%;
  border: 0.5px solid var(--pay-border-2);
  border-radius: 9px;
  background: var(--pay-surface);
  color: var(--pay-ink);
  font: inherit;
  font-size: 13px;
  outline: none;
  padding: 10px 12px;
}

.ajo-pay-checkout__voucher {
  display: flex;
  grid-column: 1 / -1;
  align-items: center;
  gap: 10px;
}

.ajo-pay-checkout__voucher input {
  display: none;
}

.ajo-pay-checkout__voucher-list {
  display: grid;
  grid-column: 1 / -1;
  gap: 8px;
}

.ajo-pay-checkout__voucher-list article {
  display: grid;
  grid-template-columns: 48px minmax(0, 1fr) auto;
  align-items: center;
  gap: 10px;
}

.ajo-pay-checkout__voucher-list img {
  width: 48px;
  height: 48px;
  border-radius: 8px;
  object-fit: cover;
}

.ajo-pay-checkout__confirm {
  width: 100%;
  border: 0;
  border-radius: 9px;
  background: var(--pay-brand);
  color: #ffffff;
  cursor: pointer;
  font-family: 'Noto Sans TC', var(--font-sans);
  font-size: 15px;
  font-weight: 700;
  padding: 16px;
  transition: background 0.15s ease;
}

.ajo-pay-checkout__confirm:hover {
  background: var(--pay-brand-dark);
}

.ajo-pay-checkout__confirm:disabled {
  background: var(--pay-border-2);
  color: var(--pay-ink-3);
  cursor: not-allowed;
}

.ajo-pay-checkout__hint {
  margin: 8px 0 0;
  color: var(--pay-ink-3);
  font-size: 11px;
  line-height: 1.6;
  text-align: center;
}

.ajo-pay-checkout__complete-note {
  border: 0.5px solid var(--pay-brand-mid);
  border-radius: 9px;
  background: var(--pay-brand-light);
  color: var(--pay-ink-2);
  font-size: 12px;
  line-height: 1.6;
  margin-bottom: 14px;
  padding: 13px 16px;
}

.ajo-pay-checkout__order-result {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
  border: 0.5px solid var(--pay-border);
  border-radius: 14px;
  margin-bottom: 14px;
  padding: 16px;
}

.ajo-pay-checkout__order-result div {
  display: grid;
  gap: 4px;
  min-width: 0;
}

.ajo-pay-checkout__order-result span,
.ajo-pay-checkout__order-result small {
  color: var(--pay-ink-3);
  font-size: 12px;
}

.ajo-pay-checkout__order-result strong {
  overflow: hidden;
  color: var(--pay-ink);
  font-size: 14px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ajo-pay-checkout__qr {
  display: grid;
  justify-items: center;
  gap: 12px;
  border: 0.5px solid var(--pay-border);
  border-radius: 14px;
  margin-bottom: 14px;
  padding: 18px;
}

.ajo-pay-checkout__qr img,
.ajo-pay-checkout__qr :deep(img) {
  width: 220px;
  height: 220px;
  object-fit: contain;
}

.ajo-pay-checkout__qr a {
  color: var(--pay-brand);
  font-size: 13px;
  font-weight: 700;
  text-decoration: none;
}

.ajo-pay-checkout__receipt {
  border: 0.5px solid var(--pay-border);
  border-radius: 9px;
  margin-bottom: 14px;
}

.ajo-pay-checkout__receipt strong {
  color: var(--pay-success);
}

.ajo-pay-checkout__record-title {
  margin: 12px 0 10px;
}

.ajo-pay-checkout__records {
  display: grid;
  gap: 8px;
}

.ajo-pay-checkout__record {
  border: 0.5px solid var(--pay-border);
  border-radius: 9px;
  opacity: 0.72;
}

.ajo-pay-checkout__record small {
  color: var(--pay-success);
  font-size: 11px;
  font-weight: 700;
}

@media (max-width: 767px) {
  .ajo-pay-checkout {
    padding: 20px var(--layout-page-padding-inline) calc(var(--app-mobile-content-bottom) + 24px);
  }

  .ajo-pay-checkout__steps {
    gap: 8px;
    overflow-x: auto;
    padding-bottom: 4px;
  }

  .ajo-pay-checkout__step strong {
    font-size: 11px;
  }

  .ajo-pay-checkout__bill,
  .ajo-pay-checkout__method,
  .ajo-pay-checkout__summary-box div,
  .ajo-pay-checkout__record {
    align-items: flex-start;
    flex-wrap: wrap;
  }

  .ajo-pay-checkout__offline {
    grid-template-columns: 1fr;
  }

  .ajo-pay-checkout__notice button,
  .ajo-pay-checkout__link-button,
  .ajo-pay-checkout__order-result button,
  .ajo-pay-checkout__voucher button,
  .ajo-pay-checkout__voucher-list button {
    width: 100%;
  }
}
</style>

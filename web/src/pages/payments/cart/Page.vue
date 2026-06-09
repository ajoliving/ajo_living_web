<!--
 * POS 購物車頁。
 * 1. 展示已加入的物業費賬單、總額與手續費。
 * 2. 支援 H5 / QR 線上支付與現金、支票、銀行轉賬線下繳費。
-->
<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import { useFeedbackStore } from '@/app/stores/feedback';
import { usePOSPaymentCartStore } from '@/app/stores/pos-payment-cart';
import {
  fetchPOSPaymentBills,
  createPOSPaymentOrder,
  fetchPOSPaymentBankAccounts,
  fetchPOSPaymentFees,
  reportPOSPayment,
} from '@/domains/payments/api';
import type { POSPaymentChannel, POSPaymentMethodKey, POSPaymentRow } from '@/domains/payments/model';
import BaseButton from '@/shared/components/base/BaseButton.vue';
import BaseSelect from '@/shared/components/base/BaseSelect.vue';

import {
  readPaymentDate,
  readPaymentText,
  readPaymentTitle,
  usePOSPaymentPage,
} from '../composables/usePOSPaymentPage';

type CheckoutMethodKey = 'POS_WECHAT' | 'POS_ALIPAY_CN' | 'POS_ALIPAY_HK' | 'POS_BANK' | 'POS_CHEQUE' | 'POS_CASH';

interface CheckoutMethod {
  key: CheckoutMethodKey;
  label: string;
  category: 'online' | 'offline';
  payType: POSPaymentMethodKey;
  walletType?: 'CN' | 'HK';
  variant?: 'wechat' | 'alipay';
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

const route = useRoute();
const router = useRouter();
const feedbackStore = useFeedbackStore();
const cartStore = usePOSPaymentCartStore();
const {
  activeBuildingID,
  activeUnitID,
  contextLabel,
  goPaymentUnit,
  isLoading,
  posLoginRequired,
  profileRequired,
} = usePOSPaymentPage('bills');

const cartContext = computed(() => ({
  buildingID: activeBuildingID.value,
  unitID: activeUnitID.value,
}));
const hasCartContext = computed(() => activeBuildingID.value !== '' && activeUnitID.value !== '');
const currentCartItems = computed(() => hasCartContext.value ? cartStore.itemsForContext(cartContext.value) : []);
const currentCartCount = computed(() => hasCartContext.value ? cartStore.countForContext(cartContext.value) : 0);
const singleCheckoutKey = computed(() => String(route.query.checkout_bill_key ?? '').trim());
const selectedBillKeys = ref<string[]>([]);
const checkoutDialogOpen = ref(false);
const checkoutStepIndex = ref(0);
const hasAppliedSingleCheckout = ref(false);
const hasInitializedSelection = ref(false);
const currentCartKeys = computed(() => currentCartItems.value.map((row) => cartStore.billKeyForContext(row, cartContext.value)).filter(Boolean));
const selectedBillKeySet = computed(() => new Set(selectedBillKeys.value));
const checkoutItems = computed(() => currentCartItems.value.filter((row) => selectedBillKeySet.value.has(cartStore.billKeyForContext(row, cartContext.value))));
const checkoutCount = computed(() => checkoutItems.value.length);
const methods = ref<CheckoutMethod[]>([]);
const bankAccounts = ref<BankAccountOption[]>([]);
const selectedMethodKey = ref<CheckoutMethodKey>('POS_WECHAT');
const hasUserSelectedMethod = ref(false);
const isSubmitting = ref(false);
const voucherInput = ref<HTMLInputElement | null>(null);
const voucherImages = ref<VoucherImage[]>([]);
const posPaymentExpireSeconds = 180;

const form = reactive({
  tranDateTime: '',
  tranRefNo: '',
  remark: '',
  bankAccountNo: '',
  receivedAmount: '',
  chequeDeposited: '',
});

const selectedMethod = computed(() => methods.value.find((item) => item.key === selectedMethodKey.value) ?? methods.value[0] ?? null);
const isCashMethod = computed(() => selectedMethod.value?.key === 'POS_CASH');
const isChequeMethod = computed(() => selectedMethod.value?.key === 'POS_CHEQUE');
const isBankTransferMethod = computed(() => selectedMethod.value?.key === 'POS_BANK');
const receivedAmount = computed(() => {
  const numeric = Number(form.receivedAmount);
  return Number.isFinite(numeric) ? numeric : 0;
});
const feeAmount = computed(() =>
  checkoutItems.value.reduce((sum, row) => sum + calculateFeeAmount(readCheckoutAmountValue(row), selectedMethod.value?.feeRate ?? 0), 0),
);
const finalAmount = computed(() => checkoutAmount.value + feeAmount.value);
const changeAmount = computed(() => Math.round((receivedAmount.value - finalAmount.value) * 100) / 100);
const isCashEnough = computed(() => !isCashMethod.value || changeAmount.value >= 0);
const canSubmit = computed(() => (
  checkoutCount.value > 0 &&
  selectedMethod.value !== null &&
  finalAmount.value > 0 &&
  (!isCashMethod.value || (form.receivedAmount !== '' && isCashEnough.value)) &&
  (!isChequeMethod.value || (form.tranRefNo !== '' && form.chequeDeposited !== '')) &&
  (!requiresBankAccount.value || form.bankAccountNo !== '') &&
  !isSubmitting.value
));
const checkoutSteps = ['選擇付款方式', '確認金額', '備註與提交'];
const activeCheckoutStep = computed(() => checkoutSteps[checkoutStepIndex.value] ?? checkoutSteps[0]);
const isAllSelected = computed(() => currentCartKeys.value.length > 0 && currentCartKeys.value.every((key) => selectedBillKeySet.value.has(key)));
const isPartiallySelected = computed(() => selectedBillKeys.value.length > 0 && !isAllSelected.value);
const allSelectionAriaChecked = computed(() => isPartiallySelected.value ? 'mixed' : isAllSelected.value);
const chequeStatusOptions = [
  { label: '請選擇', value: '' },
  { label: '已存入銀行', value: 'deposited' },
  { label: '未存入銀行', value: 'undeposited' },
];
const requiresBankAccount = computed(() =>
  isBankTransferMethod.value || (isChequeMethod.value && form.chequeDeposited === 'deposited'),
);

// 1. 讀取結賬金額
const readCheckoutAmountValue = (row: POSPaymentRow): number => {
  const raw = readPaymentText(row, [
    'net_amount',
    'amount',
    'final_amount',
    'trs_val',
    'total',
    'total_amount',
    'payable',
  ]);
  const numeric = Number(raw);
  return Number.isFinite(numeric) ? numeric : 0;
};

// 2. 讀取結賬金額文字
const readCheckoutAmount = (row: POSPaymentRow): string => {
  const amount = readCheckoutAmountValue(row);
  if (amount > 0) {
    return amount.toFixed(2);
  }
  return readPaymentText(row, [
    'net_amount',
    'amount',
    'final_amount',
    'trs_val',
    'total',
    'total_amount',
    'payable',
  ]) || '-';
};

const checkoutAmount = computed(() => checkoutItems.value.reduce((sum, row) => sum + readCheckoutAmountValue(row), 0));

// 3. 判斷賬單是否已選取
const isBillSelected = (row: POSPaymentRow): boolean =>
  selectedBillKeySet.value.has(cartStore.billKeyForContext(row, cartContext.value));

// 4. 切換單筆賬單選取狀態
const toggleBillSelection = (row: POSPaymentRow, checked?: boolean): void => {
  const key = cartStore.billKeyForContext(row, cartContext.value);
  if (!key) {
    return;
  }
  const nextSelected = checked ?? !selectedBillKeySet.value.has(key);
  if (nextSelected) {
    selectedBillKeys.value = Array.from(new Set([...selectedBillKeys.value, key]));
    return;
  }
  selectedBillKeys.value = selectedBillKeys.value.filter((item) => item !== key);
};

// 5. 接收單筆賬單勾選事件
const handleBillSelectionChange = (row: POSPaymentRow, event: Event): void => {
  toggleBillSelection(row, (event.target as HTMLInputElement).checked);
};

// 6. 切換目前購物車全部選取狀態
const toggleAllSelection = (event: Event): void => {
  const checked = (event.target as HTMLInputElement).checked;
  selectedBillKeys.value = checked ? currentCartKeys.value : [];
};

// 7. 選取全部賬單
const selectAllBills = (): void => {
  selectedBillKeys.value = currentCartKeys.value;
};

// 8. 清空選取賬單
const clearSelectedBills = (): void => {
  selectedBillKeys.value = [];
};

// 9. 開啟結賬彈窗
const openCheckoutDialog = (): void => {
  if (checkoutCount.value === 0) {
    feedbackStore.pushToast('請先選擇賬單', 'error');
    return;
  }
  if (!form.tranDateTime) {
    form.tranDateTime = nowLocalDateTime();
  }
  checkoutStepIndex.value = 0;
  checkoutDialogOpen.value = true;
};

// 10. 關閉結賬彈窗
const closeCheckoutDialog = (): void => {
  if (isSubmitting.value) {
    return;
  }
  checkoutDialogOpen.value = false;
};

// 11. 前往下一步
const goNextCheckoutStep = (): void => {
  checkoutStepIndex.value = Math.min(checkoutStepIndex.value + 1, checkoutSteps.length - 1);
};

// 12. 返回上一步
const goPreviousCheckoutStep = (): void => {
  checkoutStepIndex.value = Math.max(checkoutStepIndex.value - 1, 0);
};

// 13. 同步購物車選取狀態
const syncSelectedBillKeys = (): void => {
  const availableKeys = currentCartKeys.value;
  const availableKeySet = new Set(availableKeys);
  const singleKey = singleCheckoutKey.value;
  if (singleKey && availableKeySet.has(singleKey) && !hasAppliedSingleCheckout.value) {
    selectedBillKeys.value = [singleKey];
    hasAppliedSingleCheckout.value = true;
    hasInitializedSelection.value = true;
    openCheckoutDialog();
    return;
  }
  selectedBillKeys.value = selectedBillKeys.value.filter((key) => availableKeySet.has(key));
  if (!hasInitializedSelection.value && selectedBillKeys.value.length === 0 && availableKeys.length > 0) {
    selectedBillKeys.value = availableKeys;
    hasInitializedSelection.value = true;
  }
};

// 14. 讀取訂單資料
const readOrderData = (payload: Record<string, unknown>): Record<string, unknown> => {
  const nested = payload.data;
  if (nested && typeof nested === 'object' && !Array.isArray(nested)) {
    return nested as Record<string, unknown>;
  }
  return payload;
};

// 15. 建立日期時間
const nowLocalDateTime = (): string => {
  const now = new Date();
  const pad = (value: number): string => String(value).padStart(2, '0');
  return `${now.getFullYear()}-${pad(now.getMonth() + 1)}-${pad(now.getDate())}T${pad(now.getHours())}:${pad(now.getMinutes())}`;
};

// 16. 轉成 POS API 日期時間
const toAPIDateTime = (value: string): string =>
  value ? `${value.replace('T', ' ')}:00` : `${nowLocalDateTime().replace('T', ' ')}:00`;

// 17. 讀取圖片
const readFileAsDataURL = (file: File): Promise<string> =>
  new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => resolve(String(reader.result ?? ''));
    reader.onerror = () => reject(new Error('圖片讀取失敗'));
    reader.readAsDataURL(file);
  });

// 18. 選擇憑證圖片
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
    feedbackStore.pushToast('圖片讀取失敗', 'error');
  } finally {
    input.value = '';
  }
};

// 19. 移除憑證圖片
const removeVoucher = (index: number): void => {
  const target = voucherImages.value[index];
  if (!target) {
    return;
  }
  URL.revokeObjectURL(target.previewUrl);
  voucherImages.value.splice(index, 1);
};

// 20. 讀取手續費率
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

// 20.1 格式化費率
const formatFeeRate = (feeRate: number): string =>
  `${(feeRate * 100).toFixed(2)}%`;

// 20.2 計算單筆手續費
const calculateFeeAmount = (amount: number, feeRate: number): number => {
  if (!Number.isFinite(amount) || amount <= 0 || !Number.isFinite(feeRate) || feeRate <= 0) {
    return 0;
  }
  return Math.ceil(amount * feeRate * 100 - 1e-8) / 100;
};

// 20.1 讀取瀏覽器裝置資訊
const readBrowserUserAgent = (): string =>
  typeof navigator === 'undefined' ? '' : navigator.userAgent.toLowerCase();

// 20.2 判斷 iPadOS 觸控裝置
const isIPadOS = (): boolean =>
  typeof navigator !== 'undefined' &&
  navigator.platform === 'MacIntel' &&
  navigator.maxTouchPoints > 1;

// 20.3 判斷平板裝置
const isTabletPaymentDevice = (): boolean => {
  const userAgent = readBrowserUserAgent();
  if (userAgent.includes('ipad') || isIPadOS()) {
    return true;
  }
  if (/(tablet|playbook|silk)/i.test(userAgent)) {
    return true;
  }
  return userAgent.includes('android') && !userAgent.includes('mobile');
};

// 20.4 判斷手機裝置
const isMobilePhonePaymentDevice = (): boolean => {
  const userAgent = readBrowserUserAgent();
  if (!userAgent) {
    return false;
  }
  return /(iphone|ipod|windows phone|iemobile|blackberry|bb10|opera mini|mobile)/i.test(userAgent);
};

// 20.5 判斷目前是否優先展示 QR 付款
const prefersQRPayment = (): boolean =>
  !isMobilePhonePaymentDevice() && !isTabletPaymentDevice();

const onlineMethodOrder: CheckoutMethodKey[] = ['POS_WECHAT', 'POS_ALIPAY_CN', 'POS_ALIPAY_HK'];
const offlineMethodOrder: CheckoutMethodKey[] = ['POS_CHEQUE', 'POS_CASH', 'POS_BANK'];

// 20.6 取得目前裝置的付款方式排序
const currentMethodOrder = (): CheckoutMethodKey[] => [
  ...onlineMethodOrder,
  ...offlineMethodOrder,
];

// 20.7 按裝置偏好排序付款方式
const sortMethodsForDevice = (items: CheckoutMethod[]): CheckoutMethod[] => {
  const order = currentMethodOrder();
  return items.slice().sort((left, right) => order.indexOf(left.key) - order.indexOf(right.key));
};

// 20.8 取得預設付款方式
const preferredMethodKey = (items: CheckoutMethod[]): CheckoutMethodKey =>
  currentMethodOrder().find((key) => items.some((item) => item.key === key)) ?? items[0]?.key ?? 'POS_WECHAT';

// 20.9 解析線上支付通道
const resolveOnlineChannel = (method: CheckoutMethod): POSPaymentChannel => {
  if (method.variant === 'wechat') {
    return prefersQRPayment() ? 'WX_QR' : 'WX_H5';
  }
  if (method.variant === 'alipay') {
    return prefersQRPayment() && !method.walletType ? 'ALI_QR' : 'ALI_H5';
  }
  return 'WX_H5';
};

// 20.10 讀取提交按鈕文案
const checkoutSubmitLabel = computed(() => {
  if (isSubmitting.value) {
    return '提交中';
  }
  if (selectedMethod.value?.category === 'online') {
    return prefersQRPayment() ? '生成二維碼' : '前往支付';
  }
  return '提交結賬';
});

// 21. 建立預設支付方式
const fallbackMethods = (): CheckoutMethod[] => sortMethodsForDevice([
  { key: 'POS_WECHAT', label: '微信支付', category: 'online', payType: 'POS_WECHAT', variant: 'wechat', feeRate: 0 },
  { key: 'POS_ALIPAY_CN', label: '支付寶（大陸）', category: 'online', payType: 'POS_ALIPAY', variant: 'alipay', walletType: 'CN', feeRate: 0 },
  { key: 'POS_ALIPAY_HK', label: '支付寶香港', category: 'online', payType: 'POS_ALIPAY', variant: 'alipay', walletType: 'HK', feeRate: 0 },
  { key: 'POS_CHEQUE', label: '支票', category: 'offline', payType: 'POS_CHEQUE', feeRate: 0 },
  { key: 'POS_CASH', label: '現金', category: 'offline', payType: 'POS_CASH', feeRate: 0 },
  { key: 'POS_BANK', label: '銀行轉賬', category: 'offline', payType: 'POS_BANK', feeRate: 0 },
]);

// 22. 按 POS pay_type 映射支付方式
const mapFeeRowsToMethods = (rows: POSPaymentRow[]): CheckoutMethod[] => {
  const mapped: CheckoutMethod[] = [];
  rows.forEach((row) => {
    const code = String(row.pay_type ?? row.method ?? '').trim().toUpperCase();
    const feeRate = normalizeFeeRate(row.markup);
    if (code === 'POS_ALIWE' || code === 'WEBPOS_WECHAT' || code === 'WEBPOS_ALIPAY') {
      if (code !== 'WEBPOS_ALIPAY') {
        mapped.push({ key: 'POS_WECHAT', label: '微信支付', category: 'online', payType: 'POS_WECHAT', variant: 'wechat', feeRate });
      }
      if (code !== 'WEBPOS_WECHAT') {
        mapped.push(
          { key: 'POS_ALIPAY_CN', label: '支付寶（大陸）', category: 'online', payType: 'POS_ALIPAY', variant: 'alipay', walletType: 'CN', feeRate },
          { key: 'POS_ALIPAY_HK', label: '支付寶香港', category: 'online', payType: 'POS_ALIPAY', variant: 'alipay', walletType: 'HK', feeRate },
        );
      }
    }
    if (code === 'POS_BANK' || code === 'WEBPOS_BANK') {
      mapped.push({ key: 'POS_BANK', label: '銀行轉賬', category: 'offline', payType: 'POS_BANK', feeRate: 0 });
    }
    if (code === 'POS_CHEQUE' || code === 'WEBPOS_CHEQUE') {
      mapped.push({ key: 'POS_CHEQUE', label: '支票', category: 'offline', payType: 'POS_CHEQUE', feeRate: 0 });
    }
    if (code === 'POS_CASH') {
      mapped.push({ key: 'POS_CASH', label: '現金', category: 'offline', payType: 'POS_CASH', feeRate: 0 });
    }
  });

  const seen = new Set<string>();
  return mapped
    .filter((item) => {
      if (seen.has(item.key)) {
        return false;
      }
      seen.add(item.key);
      return true;
    })
    .sort((left, right) => currentMethodOrder().indexOf(left.key) - currentMethodOrder().indexOf(right.key));
};

// 23. 載入支付設定
const loadPaymentSettings = async (): Promise<void> => {
  if (profileRequired.value || posLoginRequired.value) {
    return;
  }
  const query = {
    building_id: activeBuildingID.value || undefined,
    unit_id: activeUnitID.value || undefined,
  };
  try {
    const [feeResponse, bankResponse] = await Promise.all([
      fetchPOSPaymentFees(query),
      fetchPOSPaymentBankAccounts(query),
    ]);
    const mapped = mapFeeRowsToMethods(feeResponse.data.data.items ?? []);
    methods.value = mapped.length > 0 ? mapped : fallbackMethods();
    bankAccounts.value = (bankResponse.data.data.items ?? []).map((item) => ({
      label: readPaymentText(item, ['title', 'account_no', 'bank_name']) || '銀行戶口',
      value: readPaymentText(item, ['account_no', 'id']),
    })).filter((item) => item.value !== '');
  } catch {
    methods.value = fallbackMethods();
    bankAccounts.value = [];
  }
  const selectedMethodAvailable = methods.value.some((item) => item.key === selectedMethodKey.value);
  if (!selectedMethodAvailable || !hasUserSelectedMethod.value) {
    selectedMethodKey.value = preferredMethodKey(methods.value);
  }
  form.bankAccountNo = bankAccounts.value[0]?.value ?? '';
};

// 24. 提交前重新校驗賬單
const validateCheckoutBills = async (): Promise<boolean> => {
  const query = {
    building_id: activeBuildingID.value || undefined,
    unit_id: activeUnitID.value || undefined,
  };
  try {
    const response = await fetchPOSPaymentBills(query);
    const latestRows = response.data.data.items ?? [];
    const latestAmountMap = new Map<string, number>();
    latestRows.forEach((row) => {
      const key = cartStore.billKeyForContext(row, cartContext.value);
      if (key) {
        latestAmountMap.set(key, readCheckoutAmountValue(row));
      }
    });

    for (const row of checkoutItems.value) {
      const key = cartStore.billKeyForContext(row, cartContext.value);
      const latestAmount = latestAmountMap.get(key);
      if (latestAmount === undefined) {
        feedbackStore.pushToast('賬單已更新，請重新整理購物車', 'error');
        return false;
      }
      if (readCheckoutAmountValue(row) - latestAmount > 0.01) {
        feedbackStore.pushToast('賬單金額已更新，請重新整理購物車', 'error');
        return false;
      }
    }
    return true;
  } catch {
    feedbackStore.pushToast('賬單校驗失敗，請稍後重試', 'error');
    return false;
  }
};

// 25. 組裝 POS 賬單資料
const buildBillObjects = (feeRate: number): POSPaymentRow[] => checkoutItems.value.map((row) => {
  const amountCents = Math.round(readCheckoutAmountValue(row) * 100);
  const feeCents = Math.round(calculateFeeAmount(readCheckoutAmountValue(row), feeRate) * 100);
  return {
    flat_code: readPaymentText(row, ['flat_code', 'unit', 'unit_name']),
    unit_name: readPaymentText(row, ['unit_name', 'unit', 'flat_code']),
    item_id: readPaymentText(row, ['item_name', 'item_id', 'name']) || '賬單項目',
    trs_to: readPaymentText(row, ['trs_to', 'term', 'period']),
    bill_dt: readPaymentDate(row),
    net_amount: amountCents,
    invoice_no: readPaymentText(row, ['invoice_no', 'bill_no', 'bill_number']),
    transfer_fee: feeCents,
  };
});

// 26. 組裝手續費資料
const buildFeeObjects = (feeRate: number): POSPaymentRow[] => checkoutItems.value.map((row) => ({
  invoice_no: readPaymentText(row, ['invoice_no', 'bill_no', 'bill_number']),
  flat_code: readPaymentText(row, ['flat_code', 'unit', 'unit_name']),
  handling_fee: Math.round(calculateFeeAmount(readCheckoutAmountValue(row), feeRate) * 100),
}));

// 27. 跳轉或展示 H5 / QR 訂單
const handleCreatedOrder = async (payload: Record<string, unknown>, channel: POSPaymentChannel, method: CheckoutMethod): Promise<void> => {
  const order = readOrderData(payload);
  const payData = String(order.pay_data ?? '').trim();
  const payDataType = String(order.pay_data_type ?? '').trim().toLowerCase();
  const mchOrderNo = String(order.mch_order_no ?? order.mchOrderNo ?? '').trim();
  const payOrderId = String(order.pay_order_id ?? order.payOrderId ?? '').trim();
  const isRedirect = payDataType === '' || payDataType === 'payurl';

  if (payData && isRedirect && channel.endsWith('_H5') && !prefersQRPayment()) {
    window.location.assign(payData);
    return;
  }

  await router.push({
    path: '/payments/orders',
    query: {
      ...(mchOrderNo ? { mch_order_no: mchOrderNo } : {}),
      ...(payOrderId ? { pay_order_id: payOrderId } : {}),
      pay_channel: channel,
      payment_method_key: method.key,
      payment_label: method.label,
    },
  });
};

// 27.1 選擇結賬付款方式
const selectCheckoutMethod = (key: CheckoutMethodKey): void => {
  selectedMethodKey.value = key;
  hasUserSelectedMethod.value = true;
};

// 28. 提交線上支付
const submitOnline = async (method: CheckoutMethod): Promise<void> => {
  const payChannel = resolveOnlineChannel(method);
  const response = await createPOSPaymentOrder({
    scene: 'cart',
    building_id: activeBuildingID.value,
    unit_id: activeUnitID.value,
    pay_channel: payChannel,
    expire_seconds: posPaymentExpireSeconds,
    final_amount: Math.round(finalAmount.value * 100),
    handle_fee_amount: Math.round(feeAmount.value * 100),
    bill_objs: buildBillObjects(method.feeRate),
    handle_fee_obj: buildFeeObjects(method.feeRate),
    return_path: `${window.location.origin}/payments/orders`,
    remark: form.remark,
    ...(method.walletType ? { gateway_request_overrides: { walletType: method.walletType } } : {}),
  });
  await handleCreatedOrder(response.data.data, payChannel, method);
};

// 29. 完成已入賬支付
const finishReportedPayment = async (payload: Record<string, unknown>): Promise<void> => {
  const result = payload;
  const receipt = String((result.ismart_receipt_no as Record<string, unknown> | undefined)?.receipt_id ?? result.receipt_id ?? '').trim();
  voucherImages.value.forEach((item) => URL.revokeObjectURL(item.previewUrl));
  voucherImages.value = [];
  cartStore.removeByKeys(checkoutItems.value.map((row) => cartStore.billKeyForContext(row, cartContext.value)));
  feedbackStore.pushToast(receipt ? `繳費已提交：${receipt}` : '繳費已提交', 'success');
  await router.push('/payments/history');
};

// 31. 提交線下繳費
const submitOffline = async (method: CheckoutMethod): Promise<void> => {
  const entryDateTime = toAPIDateTime(nowLocalDateTime());
  const response = await reportPOSPayment({
    building_id: activeBuildingID.value,
    unit_id: activeUnitID.value,
    FINAL_AMOUNT: Math.round(finalAmount.value * 100),
    ENTRY_DATETIME: entryDateTime,
    TRAN_DATETIME: toAPIDateTime(form.tranDateTime),
    TRAN_REF_NO: form.tranRefNo || `${method.key}-${Date.now()}`,
    COMMENT: form.remark,
    PIC_FILENAME: voucherImages.value.map((item) => item.name),
    PIC_DATA: voucherImages.value.map((item) => item.data),
    BILL_OBJS: buildBillObjects(0),
    PAY_METHOD: method.payType,
    BLG_ID: activeBuildingID.value,
    UNIT_ID: activeUnitID.value,
    bank_account_received: requiresBankAccount.value ? form.bankAccountNo : '',
  });
  await finishReportedPayment(response.data.data);
};

// 32. 提交結賬
const submitCheckout = async (): Promise<void> => {
  const method = selectedMethod.value;
  if (!method || !canSubmit.value || !hasCartContext.value) {
    return;
  }
  if (requiresBankAccount.value && !form.bankAccountNo) {
    feedbackStore.pushToast('請先選擇銀行戶口', 'error');
    return;
  }
  if (isCashMethod.value && !isCashEnough.value) {
    feedbackStore.pushToast('實收現金不足', 'error');
    return;
  }
  if (isChequeMethod.value && form.chequeDeposited === '') {
    feedbackStore.pushToast('請選擇支票是否已存入銀行', 'error');
    return;
  }
  isSubmitting.value = true;
  try {
    const isValid = await validateCheckoutBills();
    if (!isValid) {
      return;
    }
    if (method.category === 'online') {
      await submitOnline(method);
    } else {
      await submitOffline(method);
    }
  } catch {
    feedbackStore.pushToast('結賬失敗，請稍後重試', 'error');
  } finally {
    isSubmitting.value = false;
  }
};

// 33. 移除賬單
const removeItem = (row: POSPaymentRow): void => {
  cartStore.remove(row, cartContext.value);
};

onMounted(() => {
  form.tranDateTime = nowLocalDateTime();
  void loadPaymentSettings();
});

watch([activeBuildingID, activeUnitID], () => {
  if (!isLoading.value) {
    void loadPaymentSettings();
  }
});

watch(currentCartKeys, syncSelectedBillKeys, { immediate: true });

watch(singleCheckoutKey, () => {
  hasAppliedSingleCheckout.value = false;
  syncSelectedBillKeys();
});
</script>

<template>
  <main class="pos-page">
    <section class="pos-page__header">
      <div>
        <p>Payments / Cart</p>
        <h1>購物車</h1>
        <span>{{ contextLabel }}</span>
      </div>
      <BaseButton
        variant="secondary"
        size="md"
        :disabled="currentCartCount === 0"
        @click="cartStore.clearContext(cartContext)"
      >
        清空
      </BaseButton>
    </section>

    <section
      v-if="profileRequired || posLoginRequired"
      class="pos-page__notice"
    >
      <p>{{ profileRequired ? '請先選擇支付單元。' : '請先使用 ismart 帳戶登入以同步 POS 權限。' }}</p>
      <BaseButton
        v-if="profileRequired"
        variant="primary"
        size="md"
        @click="goPaymentUnit"
      >
        選擇單元
      </BaseButton>
    </section>

    <section
      v-else-if="currentCartCount === 0"
      class="pos-page__empty"
    >
      暫時沒有已加入賬單
    </section>

    <template v-else>
      <section class="pos-page__toolbar">
        <div>
          <strong>已選 {{ checkoutCount }} 筆賬單</strong>
          <span>結賬金額 HKD {{ finalAmount.toFixed(2) }}</span>
        </div>
        <div class="pos-page__toolbar-actions">
          <BaseButton
            variant="secondary"
            size="md"
            :disabled="checkoutCount === currentCartCount"
            @click="selectAllBills"
          >
            全選
          </BaseButton>
          <BaseButton
            variant="ghost"
            size="md"
            :disabled="checkoutCount === 0"
            @click="clearSelectedBills"
          >
            取消選取
          </BaseButton>
          <BaseButton
            variant="primary"
            size="md"
            :disabled="checkoutCount === 0"
            @click="openCheckoutDialog"
          >
            結賬
          </BaseButton>
        </div>
      </section>

      <section class="pos-table-wrap">
        <table class="pos-table pos-table--cart">
          <thead>
            <tr>
              <th class="pos-table__check-cell">
                <input
                  class="pos-table__checkbox"
                  type="checkbox"
                  :checked="isAllSelected"
                  :aria-checked="allSelectionAriaChecked"
                  @change="toggleAllSelection"
                >
              </th>
              <th>賬單</th>
              <th>期數</th>
              <th>賬單日</th>
              <th>金額</th>
              <th class="pos-table__actions">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="(row, index) in currentCartItems"
              :key="cartStore.billKeyForContext(row, cartContext) || String(index)"
              class="pos-table__select-row"
              :class="isBillSelected(row) ? 'pos-table__select-row--active' : ''"
            >
              <td class="pos-table__check-cell">
                <input
                  class="pos-table__checkbox"
                  type="checkbox"
                  :checked="isBillSelected(row)"
                  @change="handleBillSelectionChange(row, $event)"
                >
              </td>
              <td>
                <span class="pos-table__primary">
                  <strong>{{ readPaymentTitle(row, '賬單') }}</strong>
                  <small>{{ readPaymentText(row, ['item_name', 'item_id', 'name']) || '物業費' }}</small>
                </span>
              </td>
              <td>{{ readPaymentText(row, ['trs_to', 'term', 'period']) || '-' }}</td>
              <td class="pos-table__nowrap">{{ readPaymentDate(row) }}</td>
              <td class="pos-table__amount">HKD {{ readCheckoutAmount(row) }}</td>
              <td class="pos-table__actions">
                <div class="pos-table__action-group">
                  <BaseButton
                    variant="danger"
                    size="sm"
                    @click="removeItem(row)"
                  >
                    刪除
                  </BaseButton>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </section>

      <div
        v-if="checkoutDialogOpen"
        class="pos-checkout-modal"
        role="dialog"
        aria-modal="true"
        aria-labelledby="cart-checkout-title"
      >
        <section class="pos-checkout-modal__panel">
          <header class="pos-checkout-modal__header">
            <div>
              <p>Checkout</p>
              <h2 id="cart-checkout-title">
                {{ activeCheckoutStep }}
              </h2>
              <span>{{ checkoutCount }} 筆賬單 · HKD {{ finalAmount.toFixed(2) }}</span>
            </div>
            <button
              class="pos-checkout-modal__close"
              type="button"
              :disabled="isSubmitting"
              aria-label="關閉結賬"
              @click="closeCheckoutDialog"
            >
              ×
            </button>
          </header>

          <nav class="pos-checkout-steps">
            <button
              v-for="(step, index) in checkoutSteps"
              :key="step"
              type="button"
              :class="index === checkoutStepIndex ? 'pos-checkout-steps__item--active' : ''"
              @click="checkoutStepIndex = index"
            >
              <span>{{ index + 1 }}</span>
              <strong>{{ step }}</strong>
            </button>
          </nav>

          <section class="pos-checkout-modal__body">
            <div
              v-if="checkoutStepIndex === 0"
              class="pos-checkout-methods"
            >
              <button
                v-for="method in methods"
                :key="method.key"
                type="button"
                class="pos-checkout-method"
                :class="method.key === selectedMethodKey ? 'pos-checkout-method--active' : ''"
                :disabled="isSubmitting"
                @click="selectCheckoutMethod(method.key)"
              >
                <strong>{{ method.label }}</strong>
                <span>
                  {{ method.category === 'online' ? (prefersQRPayment() ? '生成二維碼' : 'H5 支付') : '線下繳費' }}
                  · 費率 {{ formatFeeRate(method.feeRate) }}
                </span>
              </button>
            </div>

            <div
              v-else-if="checkoutStepIndex === 1"
              class="pos-checkout-review"
            >
              <article
                v-for="(row, index) in checkoutItems"
                :key="cartStore.billKeyForContext(row, cartContext) || String(index)"
              >
                <div>
                  <strong>{{ readPaymentTitle(row, '賬單') }}</strong>
                  <span>{{ readPaymentText(row, ['trs_to', 'term', 'period']) || '-' }} · {{ readPaymentDate(row) }}</span>
                </div>
                <b>HKD {{ readCheckoutAmountValue(row).toFixed(2) }}</b>
              </article>
              <div class="pos-summary">
                <div>
                  <span>賬單金額</span>
                  <strong>HKD {{ checkoutAmount.toFixed(2) }}</strong>
                </div>
                <div>
                  <span>費率</span>
                  <strong>{{ formatFeeRate(selectedMethod?.feeRate ?? 0) }}</strong>
                </div>
                <div>
                  <span>手續費</span>
                  <strong>HKD {{ feeAmount.toFixed(2) }}</strong>
                </div>
                <div>
                  <span>結賬金額</span>
                  <strong>HKD {{ finalAmount.toFixed(2) }}</strong>
                </div>
              </div>
            </div>

            <div
              v-else
              class="pos-checkout-final"
            >
              <label class="pos-checkout-remark">
                <span>備註</span>
                <textarea
                  v-model="form.remark"
                  class="pos-textarea"
                  rows="3"
                />
              </label>
              <section
                v-if="selectedMethod?.category === 'offline'"
                class="pos-checkout-grid"
              >
                <label>
                  <span>交易時間</span>
                  <input
                    v-model="form.tranDateTime"
                    class="pos-input"
                    type="datetime-local"
                  >
                </label>
                <label>
                  <span>{{ isChequeMethod ? '支票號碼' : '交易參考' }}</span>
                  <input
                    v-model="form.tranRefNo"
                    class="pos-input"
                    type="text"
                    :placeholder="isChequeMethod ? 'Cheque no.' : 'Reference no.'"
                  >
                </label>
                <label v-if="isCashMethod">
                  <span>實收現金</span>
                  <input
                    v-model="form.receivedAmount"
                    class="pos-input"
                    type="number"
                    min="0"
                    step="0.01"
                    inputmode="decimal"
                    placeholder="0.00"
                  >
                </label>
                <label v-if="isChequeMethod">
                  <span>支票狀態</span>
                  <BaseSelect
                    v-model="form.chequeDeposited"
                    :options="chequeStatusOptions"
                    :disabled="isSubmitting"
                  />
                </label>
                <label v-if="requiresBankAccount">
                  <span>銀行戶口</span>
                  <BaseSelect
                    v-model="form.bankAccountNo"
                    :options="bankAccounts"
                    :disabled="bankAccounts.length === 0 || isSubmitting"
                  />
                </label>
                <section
                  v-if="isCashMethod"
                  class="pos-inline-status"
                  :class="isCashEnough ? 'pos-inline-status--success' : 'pos-inline-status--danger'"
                >
                  <span>找回金額</span>
                  <strong>{{ isCashEnough ? `HKD ${Math.max(changeAmount, 0).toFixed(2)}` : '實收現金不足' }}</strong>
                </section>
              </section>
              <section
                v-if="isBankTransferMethod"
                class="pos-voucher-panel"
              >
                <div class="pos-section-head">
                  <div>
                    <strong>交易憑證</strong>
                    <span>{{ voucherImages.length === 0 ? '未上傳任何圖片' : `${voucherImages.length} 張圖片` }}</span>
                  </div>
                  <BaseButton
                    variant="secondary"
                    size="sm"
                    :disabled="isSubmitting"
                    @click="voucherInput?.click()"
                  >
                    選擇圖片
                  </BaseButton>
                </div>
                <input
                  ref="voucherInput"
                  class="pos-hidden-input"
                  type="file"
                  accept="image/*"
                  multiple
                  @change="onVoucherChange"
                >
                <div
                  v-if="voucherImages.length > 0"
                  class="pos-voucher-grid"
                >
                  <article
                    v-for="(image, index) in voucherImages"
                    :key="image.previewUrl"
                    class="pos-voucher-card"
                  >
                    <img
                      :src="image.previewUrl"
                      :alt="image.name"
                    >
                    <span>{{ image.name }}</span>
                    <BaseButton
                      variant="ghost"
                      size="sm"
                      :disabled="isSubmitting"
                      @click="removeVoucher(index)"
                    >
                      移除
                    </BaseButton>
                  </article>
                </div>
              </section>
              <div class="pos-summary">
                <div>
                  <span>付款方式</span>
                  <strong>{{ selectedMethod?.label || '-' }}</strong>
                </div>
                <div>
                  <span>選取賬單</span>
                  <strong>{{ checkoutCount }} 筆</strong>
                </div>
                <div>
                  <span>結賬金額</span>
                  <strong>HKD {{ finalAmount.toFixed(2) }}</strong>
                </div>
              </div>
            </div>
          </section>

          <footer class="pos-checkout-modal__footer">
            <BaseButton
              variant="secondary"
              size="md"
              :disabled="checkoutStepIndex === 0 || isSubmitting"
              @click="goPreviousCheckoutStep"
            >
              上一步
            </BaseButton>
            <BaseButton
              v-if="checkoutStepIndex < checkoutSteps.length - 1"
              variant="primary"
              size="md"
              :disabled="selectedMethod === null || checkoutCount === 0"
              @click="goNextCheckoutStep"
            >
              下一步
            </BaseButton>
            <BaseButton
              v-else
              variant="primary"
              size="md"
              :disabled="!canSubmit"
              @click="submitCheckout"
            >
              {{ checkoutSubmitLabel }}
            </BaseButton>
          </footer>
        </section>
      </div>
    </template>
  </main>
</template>

<style scoped>
@import '../payment-page.css';
</style>

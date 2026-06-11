/*
 * POS 物業繳費頁面狀態。
 * 1. 讀取繳費概覽、賬單、訂單、會計與歷史資料。
 * 2. 提供正式簡潔的金額、日期與欄位顯示方法。
 */
import { computed, onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

import { useFeedbackStore } from '@/stores/feedback';
import { useSessionStore } from '@/stores/session';
import {
  fetchPOSPaymentBills,
  fetchPOSPaymentHistory,
  fetchPOSPaymentOrders,
  fetchPOSPaymentOverview,
} from '@/httpapis/payments';
import type {
  POSPaymentContext,
  POSPaymentListPayload,
  POSPaymentOverview,
  POSPaymentRow,
} from '@/model/payments';

interface PaymentSelectOption {
  label: string;
  value: string;
}

export type POSPaymentPageKind = 'bills' | 'orders' | 'accounting' | 'history';

const emptyListPayload: POSPaymentListPayload = {
  items: [],
};

// 1. 讀取字串值
export const readPaymentText = (row: POSPaymentRow, keys: string[]): string => {
  for (const key of keys) {
    const value = row[key];
    if (value !== undefined && value !== null && String(value).trim() !== '') {
      return String(value).trim();
    }
  }
  return '';
};

// 2. 讀取金額值
export const readPaymentAmountValue = (row: POSPaymentRow): number => {
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

// 3. 讀取金額文字
export const readPaymentAmount = (row: POSPaymentRow): string => {
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
  if (Number.isFinite(numeric)) {
    return numeric.toFixed(2);
  }
  return raw || '-';
};

// 4. 讀取分制金額文字
export const readPaymentCentsAmount = (row: POSPaymentRow): string => {
  const amountText = readPaymentText(row, ['amount_hkd']);
  const amountValue = Number(amountText);
  if (Number.isFinite(amountValue)) {
    return amountValue.toFixed(2);
  }

  const centsText = readPaymentText(row, [
    'final_amount',
    'amount',
    'paid_amount',
    'amount_cents',
    'finalAmount',
    'paidAmount',
  ]);
  const centsValue = Number(centsText);
  if (Number.isFinite(centsValue)) {
    return (centsValue / 100).toFixed(2);
  }
  return centsText || '-';
};

// 5. 讀取日期值
export const readPaymentDate = (row: POSPaymentRow): string =>
  readPaymentText(row, [
    'bill_dt',
    'bill_date',
    'tran_time',
    'tran_datetime',
    'input_time',
    'created_at',
    'updated_at',
    'date',
  ]) || '-';

// 6. 讀取主標題
export const readPaymentTitle = (row: POSPaymentRow, fallback: string): string =>
  readPaymentText(row, [
    'invoice_no',
    'bill_no',
    'bill_number',
    'bill_id',
    'mch_order_no',
    'pay_order_id',
    'payment_id',
    'receipt_id',
    'receiptNo',
    'id',
  ]) || fallback;

// 7. 管理 POS 物業繳費頁面資料
export const usePOSPaymentPage = (kind: POSPaymentPageKind) => {
  const { t } = useI18n();
  const router = useRouter();
  const feedbackStore = useFeedbackStore();
  const sessionStore = useSessionStore();
  const overview = ref<POSPaymentOverview | null>(null);
  const payload = ref<POSPaymentListPayload>(emptyListPayload);
  const isLoading = ref(false);
  const selectedBuildingID = ref('');
  const selectedUnitID = ref('');

  const isStaff = computed(() => Boolean(sessionStore.me?.is_staff || overview.value?.is_staff));
  const context = computed<POSPaymentContext | undefined>(() => payload.value.context ?? overview.value?.context);
  const rows = computed(() => payload.value.items ?? []);
  const buildingOptions = computed<PaymentSelectOption[]>(() => (
    overview.value?.building_options ?? payload.value.building_options ?? sessionStore.me?.bound_building_ids ?? []
  ).map((item) => ({ label: item, value: item })));
  const unitOptions = computed<PaymentSelectOption[]>(() => (
    overview.value?.unit_options ?? payload.value.unit_options ?? sessionStore.me?.bound_flat_unit_ids ?? []
  )
    .filter((item) => !selectedBuildingID.value || item.startsWith(selectedBuildingID.value))
    .map((item) => ({ label: item, value: item })));
  const contextLabel = computed(() => {
    if (!context.value) {
      return t('payments.common.noUnit');
    }
    return [
      context.value.building_name || context.value.building_id,
      context.value.unit_label || [context.value.floor, context.value.unit].filter(Boolean).join(' / '),
    ].filter(Boolean).join(' · ');
  });
  const profileRequired = computed(() => Boolean(overview.value?.profile_required));
  const posLoginRequired = computed(() => Boolean(overview.value?.pos_login_required));
  const canViewAccounting = computed(() => kind !== 'accounting' || isStaff.value);
  const activeBuildingID = computed(() =>
    selectedBuildingID.value || context.value?.building_id || '',
  );
  const activeUnitID = computed(() =>
    selectedUnitID.value || context.value?.unit_id || '',
  );
  const paymentQuery = computed(() => ({
    building_id: activeBuildingID.value || undefined,
    unit_id: activeUnitID.value || undefined,
  }));

  // 7.1 前往支付單元選擇頁
  const goPaymentUnit = async (): Promise<void> => {
    await router.push('/payment/unit');
  };

  // 7.2 載入繳費概覽
  const loadOverview = async (): Promise<void> => {
    const { data } = await fetchPOSPaymentOverview({
      ...paymentQuery.value,
      ...(kind === 'accounting' ? { summary: false } : {}),
    });
    overview.value = data.data;
    if (data.data.profile_required && !data.data.pos_login_required) {
      await router.replace('/payment/unit');
      return;
    }
    if (!selectedBuildingID.value && data.data.context?.building_id) {
      selectedBuildingID.value = data.data.context.building_id;
    }
    if (!selectedUnitID.value && data.data.context?.unit_id) {
      selectedUnitID.value = data.data.context.unit_id;
    }
  };

  // 7.3 載入當前頁面列表
  const loadRows = async (): Promise<void> => {
    if (kind === 'bills') {
      const response = await fetchPOSPaymentBills(paymentQuery.value);
      payload.value = response.data.data;
      return;
    }
    if (kind === 'orders') {
      const response = await fetchPOSPaymentOrders(paymentQuery.value);
      payload.value = response.data.data;
      return;
    }
    const response = await fetchPOSPaymentHistory(paymentQuery.value);
    payload.value = response.data.data;
  };

  // 7.4 重新載入資料
  const reload = async (): Promise<void> => {
    isLoading.value = true;
    try {
      await loadOverview();
      const canLoadCurrentKind = kind === 'accounting'
        ? canViewAccounting.value && !posLoginRequired.value
        : !profileRequired.value && !posLoginRequired.value;
      if (canLoadCurrentKind && kind !== 'accounting') {
        await loadRows();
      }
    } catch {
      payload.value = emptyListPayload;
      feedbackStore.pushToast(t('payments.common.loadError'), 'error');
    } finally {
      isLoading.value = false;
    }
  };

  onMounted(() => {
    void reload();
  });

  watch(selectedBuildingID, (nextValue, previousValue) => {
    if (nextValue !== previousValue) {
      selectedUnitID.value = '';
      if (!isLoading.value) {
        void reload();
      }
    }
  });

  watch(selectedUnitID, (nextValue, previousValue) => {
    if (nextValue !== previousValue) {
      if (!isLoading.value) {
        void reload();
      }
    }
  });

  return {
    buildingOptions,
    activeBuildingID,
    activeUnitID,
    canViewAccounting,
    contextLabel,
    goPaymentUnit,
    isLoading,
    isStaff,
    overview,
    posLoginRequired,
    profileRequired,
    readPaymentAmount,
    readPaymentAmountValue,
    readPaymentCentsAmount,
    readPaymentDate,
    readPaymentText,
    readPaymentTitle,
    reload,
    rows,
    selectedBuildingID,
    selectedUnitID,
    t,
    unitOptions,
  };
};

<!--
 * POS 歷史頁。
 * 1. 支援日期範圍、日期類型與收款方式篩選。
 * 2. Staff 可按可見單位批量查詢，普通會員查目前選中單位。
 * 3. 提供交易詳情展示。
-->
<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue';

import { fetchPOSPaymentHistory, fetchPOSPaymentHistoryDetail } from '@/httpapis/payments';
import { resolvePOSPaymentMethodI18nKey } from '@/model/payment-constants';
import type { POSPaymentHistoryQuery, POSPaymentRow } from '@/model/payments';
import { useFeedbackStore } from '@/stores/feedback';
import BaseButton from '@/shared/components/base/BaseButton.vue';
import BaseSelect from '@/shared/components/base/BaseSelect.vue';

import { usePOSPaymentPage } from '../composables/usePOSPaymentPage';

interface DetailField {
  label: string;
  value: string;
}

type DateFilterKey = 'fromDate' | 'toDate';

interface CalendarDay {
  value: string;
  label: string;
  inCurrentMonth: boolean;
  isToday: boolean;
  isSelected: boolean;
}

const feedbackStore = useFeedbackStore();
const {
  activeBuildingID,
  activeUnitID,
  contextLabel,
  goPaymentUnit,
  isLoading,
  isStaff,
  posLoginRequired,
  profileRequired,
  readPaymentAmount,
  readPaymentDate,
  readPaymentText,
  readPaymentTitle,
  t,
  unitOptions,
} = usePOSPaymentPage('history');

const today = new Date();
const formatDate = (value: Date): string => {
  const pad = (part: number): string => String(part).padStart(2, '0');
  return `${value.getFullYear()}-${pad(value.getMonth() + 1)}-${pad(value.getDate())}`;
};
const todayText = formatDate(today);
const daysAgo = new Date(today);
daysAgo.setDate(today.getDate() - 6);

const filters = reactive({
  fromDate: formatDate(daysAgo),
  toDate: formatDate(today),
  dateType: 'input_date',
  payMethod: 'all',
  unitIDs: '',
});

const localRows = ref<POSPaymentRow[]>([]);
const localLoading = ref(false);
const detailRow = ref<POSPaymentRow | null>(null);
const activeDateField = ref<DateFilterKey | ''>('');
const calendarCursor = ref(new Date(today.getFullYear(), today.getMonth(), 1));

const dateTypeOptions = [
  { label: '入賬日期', value: 'input_date' },
  { label: '交易日期', value: 'tran_date' },
];
const payMethodOptions = computed(() => [
  { label: t('payments.methods.all'), value: 'all' },
  { label: t('payments.methods.wechatAlipay'), value: 'POS_ALIWE' },
  { label: t('payments.methods.unionpay'), value: 'POS_YSF_QR' },
  { label: t('payments.methods.bankTransfer'), value: 'POS_BANK' },
  { label: t('payments.methods.cheque'), value: 'POS_CHEQUE' },
  { label: t('payments.methods.cash'), value: 'POS_CASH' },
]);

const unitIDList = computed(() => filters.unitIDs
  .split(',')
  .map((item) => item.trim())
  .filter(Boolean));
const selectedUnitIDSet = computed(() => new Set(unitIDList.value));

// 1. 讀取交易詳情字段
const historyDetailFields = computed<DetailField[]>(() => {
  const row = detailRow.value;
  if (!row) {
    return [];
  }
  return [
    { label: '交易編號', value: readPaymentTitle(row, '交易') },
    { label: 'Payment ID', value: readPaymentText(row, ['payment_id', 'id']) || '-' },
    { label: '收據', value: readPaymentText(row, ['receipt_id', 'receiptNo', 'receipt_no', 'ref_no']) || '-' },
    { label: '大廈', value: readPaymentText(row, ['building_name', 'blg_id', 'BLG_ID', 'building_id']) || '-' },
    { label: '單位', value: readPaymentText(row, ['unit_name', 'flat_code', 'UNIT_ID', 'unit_id']) || '-' },
    { label: '收款方式', value: readPaymentMethod(row) },
    { label: '狀態', value: readPaymentText(row, ['status', 'state']) || '-' },
    { label: '金額', value: `HKD ${readPaymentAmount(row)}` },
    { label: '交易日期', value: readPaymentText(row, ['tran_datetime', 'TRAN_DATETIME', 'tran_time', 'tran_date']) || '-' },
    { label: '入賬日期', value: readPaymentText(row, ['input_time', 'input_date', 'ENTRY_DATETIME', 'created_at']) || '-' },
    { label: '賬單期數', value: readPaymentText(row, ['trs_to', 'term', 'period']) || '-' },
    { label: '賬單日期', value: readPaymentText(row, ['bill_dt', 'bill_date']) || '-' },
    { label: '交易參考', value: readPaymentText(row, ['tran_ref_no', 'TRAN_REF_NO', 'ref_no']) || '-' },
    { label: '操作人', value: readPaymentText(row, ['operator', 'login_name', 'created_by', 'staff_name']) || '-' },
    { label: '銀行戶口', value: readPaymentText(row, ['bank_account_received', 'bank_account_no', 'account_no']) || '-' },
    { label: '備註', value: readPaymentText(row, ['comment', 'COMMENT', 'remark']) || '-' },
  ];
});

// 1.1 讀取付款方式
const readPaymentMethod = (row: POSPaymentRow): string => {
  const rawMethod = readPaymentText(row, [
    'payment_method_key',
    'payment_label',
    'pay_channel',
    'pay_method_text',
    'pay_method_label',
    'method_label',
    'pay_type_label',
    'pay_type',
    'pay_method',
    'method',
    'PAY_METHOD',
  ]);
  const methodKey = resolvePOSPaymentMethodI18nKey(rawMethod);

  return methodKey ? t(`payments.methods.${methodKey}`) : rawMethod || '-';
};

// 1.2 讀取歷史詳情明細
const readRowList = (row: POSPaymentRow, keys: string[]): POSPaymentRow[] => {
  for (const key of keys) {
    const value = row[key];
    if (Array.isArray(value)) {
      return value.filter((item): item is POSPaymentRow =>
        item !== null && typeof item === 'object' && !Array.isArray(item),
      );
    }
  }
  return [];
};

const historyDetailItems = computed(() => {
  const row = detailRow.value;
  if (!row) {
    return [];
  }
  return readRowList(row, ['payment_detail_objs', 'items', 'payment_objs', 'payment_items', 'details', 'transactions']);
});

const weekdayLabels = ['日', '一', '二', '三', '四', '五', '六'];

const calendarMonthLabel = computed(() =>
  `${calendarCursor.value.getFullYear()}-${String(calendarCursor.value.getMonth() + 1).padStart(2, '0')}`,
);

const calendarDays = computed<CalendarDay[]>(() => {
  const monthStart = new Date(calendarCursor.value.getFullYear(), calendarCursor.value.getMonth(), 1);
  const gridStart = new Date(monthStart);
  const selectedDate = activeDateField.value ? filters[activeDateField.value] : '';

  gridStart.setDate(monthStart.getDate() - monthStart.getDay());

  return Array.from({ length: 42 }, (_, index) => {
    const date = new Date(gridStart);
    date.setDate(gridStart.getDate() + index);

    const value = formatDate(date);

    return {
      value,
      label: String(date.getDate()),
      inCurrentMonth: date.getMonth() === calendarCursor.value.getMonth(),
      isToday: value === todayText,
      isSelected: value === selectedDate,
    };
  });
});

// 2. 解析日期字串
const parseFilterDate = (value: string): Date => {
  const [year, month, day] = value.split('-').map((item) => Number(item));
  if (!Number.isFinite(year) || !Number.isFinite(month) || !Number.isFinite(day)) {
    return today;
  }

  return new Date(year, month - 1, day);
};

// 3. 打開日期彈窗
const openDatePicker = (field: DateFilterKey): void => {
  const selectedDate = parseFilterDate(filters[field]);

  activeDateField.value = field;
  calendarCursor.value = new Date(selectedDate.getFullYear(), selectedDate.getMonth(), 1);
};

// 4. 關閉日期彈窗
const closeDatePicker = (): void => {
  activeDateField.value = '';
};

// 5. 切換日曆月份
const changeCalendarMonth = (step: number): void => {
  const nextDate = new Date(calendarCursor.value);
  nextDate.setMonth(nextDate.getMonth() + step);
  calendarCursor.value = new Date(nextDate.getFullYear(), nextDate.getMonth(), 1);
};

// 6. 選擇日期
const selectCalendarDate = (value: string): void => {
  if (!activeDateField.value) {
    return;
  }
  filters[activeDateField.value] = value;
  closeDatePicker();
};

// 7. 選擇今日
const selectToday = (): void => {
  selectCalendarDate(todayText);
};

// 8. 切換 Staff 查詢單位
const toggleHistoryUnit = (unitID: string): void => {
  const next = new Set(unitIDList.value);
  if (next.has(unitID)) {
    next.delete(unitID);
  } else {
    next.add(unitID);
  }
  filters.unitIDs = Array.from(next).join(',');
};

// 9. 全選 Staff 可見單位
const toggleAllHistoryUnits = (): void => {
  if (unitIDList.value.length === unitOptions.value.length) {
    filters.unitIDs = '';
    return;
  }
  filters.unitIDs = unitOptions.value.map((item) => String(item.value)).join(',');
};

// 10. 建立查詢參數
const buildQuery = (): POSPaymentHistoryQuery => {
  const selectedStaffUnits = unitIDList.value;
  return {
    building_id: activeBuildingID.value || undefined,
    ...(!isStaff.value ? { unit_id: activeUnitID.value || undefined } : {}),
    from_date: filters.fromDate,
    to_date: filters.toDate,
    date_type: filters.dateType as 'input_date' | 'tran_date',
    pay_method: filters.payMethod,
    ...(isStaff.value && selectedStaffUnits.length > 0 ? { unit_ids: selectedStaffUnits.join(',') } : {}),
  };
};

// 11. 查詢歷史
const queryHistory = async (): Promise<void> => {
  if (profileRequired.value || posLoginRequired.value) {
    return;
  }
  if (filters.fromDate > filters.toDate) {
    feedbackStore.pushToast('開始日期不能晚於結束日期', 'error');
    return;
  }

  localLoading.value = true;
  try {
    const response = await fetchPOSPaymentHistory(buildQuery());
    localRows.value = response.data.data.items ?? [];
  } catch {
    localRows.value = [];
    feedbackStore.pushToast('查詢歷史失敗', 'error');
  } finally {
    localLoading.value = false;
  }
};

// 12. 查看交易詳情
const showDetail = async (row: POSPaymentRow): Promise<void> => {
  const paymentID = readPaymentText(row, ['payment_id', 'receipt_id', 'receiptNo', 'ref_no', 'id']);
  if (!paymentID) {
    detailRow.value = row;
    return;
  }
  try {
    const response = await fetchPOSPaymentHistoryDetail(paymentID, buildQuery());
    detailRow.value = response.data.data;
  } catch {
    detailRow.value = row;
    feedbackStore.pushToast('載入交易詳情失敗', 'error');
  }
};

watch([activeBuildingID, activeUnitID, isStaff], () => {
  void queryHistory();
}, { immediate: true });
</script>

<template>
  <main class="pos-page">
    <section class="pos-page__header">
      <div>
        <p>Payments / History</p>
        <h1>歷史</h1>
        <span>{{ contextLabel }}</span>
      </div>
      <BaseButton
        variant="secondary"
        size="sm"
        :disabled="isLoading || localLoading"
        @click="queryHistory"
      >
        {{ localLoading ? '載入中' : '查詢' }}
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
        size="sm"
        @click="goPaymentUnit"
      >
        選擇單元
      </BaseButton>
    </section>

    <template v-else>
      <section class="pos-page__panel">
        <div class="pos-checkout-grid">
          <label>
            <span>開始日期</span>
            <button
              type="button"
              class="pos-date-trigger"
              @click="openDatePicker('fromDate')"
            >
              {{ filters.fromDate }}
            </button>
          </label>
          <label>
            <span>結束日期</span>
            <button
              type="button"
              class="pos-date-trigger"
              @click="openDatePicker('toDate')"
            >
              {{ filters.toDate }}
            </button>
          </label>
          <label>
            <span>日期類型</span>
            <BaseSelect
              v-model="filters.dateType"
              :options="dateTypeOptions"
            />
          </label>
          <label>
            <span>收款方式</span>
            <BaseSelect
              v-model="filters.payMethod"
              :options="payMethodOptions"
            />
          </label>
        </div>
        <div
          v-if="activeDateField"
          class="pos-date-popover-backdrop"
          @click.self="closeDatePicker"
        >
          <section
            class="pos-date-popover"
            aria-label="日期選擇"
          >
            <header class="pos-date-popover__header">
              <button
                type="button"
                class="pos-date-popover__nav"
                aria-label="上一個月"
                @click="changeCalendarMonth(-1)"
              >
                ‹
              </button>
              <strong>{{ calendarMonthLabel }}</strong>
              <button
                type="button"
                class="pos-date-popover__nav"
                aria-label="下一個月"
                @click="changeCalendarMonth(1)"
              >
                ›
              </button>
            </header>
            <div class="pos-date-popover__weekdays">
              <span
                v-for="weekday in weekdayLabels"
                :key="weekday"
              >
                {{ weekday }}
              </span>
            </div>
            <div class="pos-date-popover__grid">
              <button
                v-for="day in calendarDays"
                :key="day.value"
                type="button"
                class="pos-date-popover__day"
                :class="{
                  'pos-date-popover__day--muted': !day.inCurrentMonth,
                  'pos-date-popover__day--today': day.isToday,
                  'pos-date-popover__day--selected': day.isSelected,
                }"
                @click="selectCalendarDate(day.value)"
              >
                {{ day.label }}
              </button>
            </div>
            <footer class="pos-date-popover__footer">
              <button
                type="button"
                @click="selectToday"
              >
                今天
              </button>
              <button
                type="button"
                @click="closeDatePicker"
              >
                關閉
              </button>
            </footer>
          </section>
        </div>
        <section
          v-if="isStaff"
          class="pos-unit-picker"
        >
          <div class="pos-section-head">
            <div>
              <strong>單位列表</strong>
              <span>{{ unitIDList.length === 0 ? '未指定單位' : `${unitIDList.length} 個單位` }}</span>
            </div>
            <BaseButton
              variant="secondary"
              size="sm"
              :disabled="unitOptions.length === 0"
              @click="toggleAllHistoryUnits"
            >
              {{ unitIDList.length === unitOptions.length ? '取消全選' : '全選單位' }}
            </BaseButton>
          </div>
          <div
            v-if="unitOptions.length === 0"
            class="pos-page__empty"
          >
            暫時沒有可選單位
          </div>
          <div
            v-else
            class="pos-unit-chip-grid"
          >
            <button
              v-for="option in unitOptions"
              :key="String(option.value)"
              type="button"
              class="pos-unit-chip"
              :class="selectedUnitIDSet.has(String(option.value)) ? 'pos-unit-chip--active' : ''"
              @click="toggleHistoryUnit(String(option.value))"
            >
              {{ option.label }}
            </button>
          </div>
        </section>
        <BaseButton
          variant="primary"
          size="sm"
          :disabled="localLoading"
          @click="queryHistory"
        >
          查詢歷史
        </BaseButton>
      </section>

      <section
        v-if="localRows.length === 0"
        class="pos-page__empty"
      >
        {{ localLoading ? '載入中' : '暫時沒有交易歷史' }}
      </section>

      <section
        v-else
        class="pos-table-wrap"
      >
        <table class="pos-table pos-table--history">
          <thead>
            <tr>
              <th>交易</th>
              <th>收據</th>
              <th>狀態</th>
              <th>時間</th>
              <th>金額</th>
              <th class="pos-table__actions">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="(row, index) in localRows"
              :key="readPaymentTitle(row, String(index))"
            >
              <td>
                <span class="pos-table__primary">
                  <strong>{{ readPaymentTitle(row, '交易') }}</strong>
                  <small>{{ readPaymentMethod(row) }}</small>
                </span>
              </td>
              <td>{{ readPaymentText(row, ['receipt_id', 'receiptNo', 'ref_no']) || '-' }}</td>
              <td>
                <span class="pos-table__status">{{ readPaymentText(row, ['status', 'state']) || '-' }}</span>
              </td>
              <td class="pos-table__nowrap">{{ readPaymentDate(row) }}</td>
              <td class="pos-table__amount">HKD {{ readPaymentAmount(row) }}</td>
              <td class="pos-table__actions">
                <div class="pos-table__action-group">
                  <BaseButton
                    variant="secondary"
                    size="sm"
                    @click="showDetail(row)"
                  >
                    詳情
                  </BaseButton>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </section>

      <div
        v-if="detailRow"
        class="pos-detail-backdrop"
        @click.self="detailRow = null"
      >
        <section class="pos-detail-modal">
          <div class="pos-page__header pos-page__header--compact">
            <div>
              <p>Transaction Detail</p>
              <h1>交易詳情</h1>
            </div>
            <BaseButton
              variant="secondary"
              size="sm"
              @click="detailRow = null"
            >
              關閉詳情
            </BaseButton>
          </div>
          <dl class="pos-detail-list">
            <div
              v-for="field in historyDetailFields"
              :key="field.label"
            >
              <dt>{{ field.label }}</dt>
              <dd>{{ field.value }}</dd>
            </div>
          </dl>
          <section
            v-if="historyDetailItems.length > 0"
            class="pos-detail-section"
          >
            <div class="pos-section-head">
              <div>
                <strong>賬單明細</strong>
                <span>{{ historyDetailItems.length }} 筆</span>
              </div>
            </div>
            <article
              v-for="(item, index) in historyDetailItems"
              :key="readPaymentText(item, ['invoice_no', 'bill_no', 'bill_id', 'item_id', 'id']) || String(index)"
              class="pos-mini-row"
            >
              <strong>{{ readPaymentText(item, ['item_id', 'item_name', 'name']) || '賬單項目' }}</strong>
              <dl>
                <div>
                  <dt>樓層</dt>
                  <dd>{{ readPaymentText(item, ['floor', 'FLOOR']) || '-' }}</dd>
                </div>
                <div>
                  <dt>單位</dt>
                  <dd>{{ readPaymentText(item, ['unit', 'unit_name', 'UNIT_ID']) || '-' }}</dd>
                </div>
                <div>
                  <dt>期數</dt>
                  <dd>{{ readPaymentText(item, ['term', 'trs_to', 'period']) || '-' }}</dd>
                </div>
                <div>
                  <dt>金額</dt>
                  <dd>HKD {{ readPaymentAmount(item) }}</dd>
                </div>
                <div>
                  <dt>備註</dt>
                  <dd>{{ readPaymentText(item, ['remark', 'comment', 'COMMENT']) || '-' }}</dd>
                </div>
              </dl>
            </article>
          </section>
        </section>
      </div>
    </template>
  </main>
</template>

<style scoped>
@import '../payment-page.css';
</style>

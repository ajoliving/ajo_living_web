<!--
 * POS 會計頁。
 * 1. 只對 Staff 顯示現金、支票待清機交易。
 * 2. 支援預設全選、多選清機、清機歷史與清機詳情。
-->
<script setup lang="ts">
import { computed, ref, watch } from 'vue';

import {
  clearPOSPaymentAccounting,
  fetchPOSPaymentAccounting,
  fetchPOSPaymentAccountingRecord,
} from '@/httpapis/payments';
import { resolvePOSPaymentMethodI18nKey } from '@/model/payment-constants';
import type { POSPaymentAccountingPayload, POSPaymentRow } from '@/model/payments';
import { useFeedbackStore } from '@/stores/feedback';
import BaseButton from '@/shared/components/base/BaseButton.vue';

import { usePOSPaymentPage } from '../composables/usePOSPaymentPage';

interface DetailField {
  label: string;
  value: string;
}

const feedbackStore = useFeedbackStore();
const {
  activeBuildingID,
  canViewAccounting,
  contextLabel,
  isLoading,
  readPaymentAmount,
  readPaymentDate,
  readPaymentText,
  readPaymentTitle,
  t,
} = usePOSPaymentPage('accounting');

const accounting = ref<POSPaymentAccountingPayload>({
  building_options: [],
  unit_options: [],
  cash_items: [],
  cheque_items: [],
  history_items: [],
  items: [],
});
const checkedCash = ref<string[]>([]);
const checkedCheque = ref<string[]>([]);
const localLoading = ref(false);
const clearingType = ref<'cash' | 'cheque' | ''>('');
const detailRecord = ref<POSPaymentRow | null>(null);

const cashTotal = computed(() => sumRows(accounting.value.cash_items));
const chequeTotal = computed(() => sumRows(accounting.value.cheque_items));

// 1. 讀取 payment id
const readPaymentID = (row: POSPaymentRow): string =>
  readPaymentText(row, ['payment_id', 'receipt_id', 'id']);

// 1.1 讀取清機記錄 id
const readAccountingRecordID = (row: POSPaymentRow): string =>
  readPaymentText(row, ['record_id', 'id', 'accounting_record_id']);

// 1.2 讀取清機記錄日期
const readAccountingRecordDate = (row: POSPaymentRow): string =>
  readPaymentText(row, ['time', 'record_time', 'created_at', 'input_time', 'date']) || readPaymentDate(row);

// 1.3 讀取付款方式
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

// 2. 讀取列表或字串字段
const readJoinedPaymentText = (row: POSPaymentRow, keys: string[]): string => {
  for (const key of keys) {
    const value = row[key];
    if (Array.isArray(value) && value.length > 0) {
      return value.map((item) => String(item).trim()).filter(Boolean).join(', ');
    }
    if (value !== undefined && value !== null && String(value).trim() !== '') {
      return String(value).trim();
    }
  }
  return '';
};

// 3. 讀取列表字段
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

// 4. 讀取清機詳情字段
const accountingDetailFields = computed<DetailField[]>(() => {
  const row = detailRecord.value;
  if (!row) {
    return [];
  }
  return [
    { label: '記錄編號', value: readPaymentTitle(row, '清機記錄') },
    { label: 'Record ID', value: readAccountingRecordID(row) || '-' },
    { label: '日期', value: readAccountingRecordDate(row) },
    { label: '大廈', value: readPaymentText(row, ['building_name', 'blg_id', 'BLG_ID', 'building_id']) || '-' },
    { label: '收款方式', value: readPaymentMethod(row) },
    { label: '狀態', value: readPaymentText(row, ['status', 'state']) || '-' },
    { label: '金額', value: `HKD ${readPaymentAmount(row)}` },
    { label: '現金總額', value: `HKD ${readPaymentText(row, ['cash_amount', 'cash_total']) || '-'}` },
    { label: '支票總額', value: `HKD ${readPaymentText(row, ['cheque_amount', 'cheque_total']) || '-'}` },
    { label: '交易數量', value: readPaymentText(row, ['payment_count', 'count', 'total_count']) || '-' },
    { label: 'Payment ID List', value: readJoinedPaymentText(row, ['payment_id_list', 'payment_ids', 'payment_id']) || '-' },
    { label: '操作人', value: readPaymentText(row, ['operator', 'login_name', 'created_by', 'staff_name']) || '-' },
    { label: '收據', value: readPaymentText(row, ['receipt_id', 'receiptNo', 'receipt_no', 'ref_no']) || '-' },
    { label: '單位', value: readPaymentText(row, ['unit_name', 'flat_code', 'UNIT_ID', 'unit_id']) || '-' },
    { label: '備註', value: readPaymentText(row, ['comment', 'COMMENT', 'remark']) || '-' },
  ];
});

// 5. 讀取清機交易明細
const accountingDetailItems = computed(() => {
  const row = detailRecord.value;
  if (!row) {
    return [];
  }
  return readRowList(row, ['payment_detail_objs', 'items', 'payment_objs', 'payment_items', 'details', 'transactions']);
});

// 6. 統計金額
const sumRows = (rows: POSPaymentRow[]): number =>
  rows.reduce((sum, row) => sum + Number(readPaymentText(row, ['amount', 'trs_val', 'total']) || 0), 0);

// 7. 載入會計資料
const loadAccounting = async (): Promise<void> => {
  if (!canViewAccounting.value || !activeBuildingID.value) {
    return;
  }
  localLoading.value = true;
  try {
    const response = await fetchPOSPaymentAccounting({
      building_id: activeBuildingID.value || undefined,
    });
    accounting.value = response.data.data;
    checkedCash.value = accounting.value.cash_items.map(readPaymentID).filter(Boolean);
    checkedCheque.value = accounting.value.cheque_items.map(readPaymentID).filter(Boolean);
  } catch {
    feedbackStore.pushToast('載入會計資料失敗', 'error');
  } finally {
    localLoading.value = false;
  }
};

// 8. 切換待清機選取
const toggleChecked = (type: 'cash' | 'cheque', paymentID: string): void => {
  const target = type === 'cash' ? checkedCash : checkedCheque;
  if (target.value.includes(paymentID)) {
    target.value = target.value.filter((item) => item !== paymentID);
    return;
  }
  target.value = [...target.value, paymentID];
};

// 9. 清機
const clearAccounting = async (type: 'cash' | 'cheque'): Promise<void> => {
  const paymentIDs = type === 'cash' ? checkedCash.value : checkedCheque.value;
  if (paymentIDs.length === 0) {
    feedbackStore.pushToast('請至少選擇一筆交易', 'error');
    return;
  }

  clearingType.value = type;
  try {
    await clearPOSPaymentAccounting({
      building_id: activeBuildingID.value,
      payment_id_list: paymentIDs,
    });
    feedbackStore.pushToast('清機已提交', 'success');
    await loadAccounting();
  } catch {
    feedbackStore.pushToast('清機提交失敗', 'error');
  } finally {
    clearingType.value = '';
  }
};

// 10. 查看清機詳情
const showRecordDetail = async (row: POSPaymentRow): Promise<void> => {
  const recordID = readAccountingRecordID(row);
  if (!recordID) {
    detailRecord.value = row;
    return;
  }
  try {
    const response = await fetchPOSPaymentAccountingRecord(recordID, {
      building_id: activeBuildingID.value,
    });
    detailRecord.value = response.data.data;
  } catch {
    detailRecord.value = row;
    feedbackStore.pushToast('載入清機詳情失敗', 'error');
  }
};

// 11. 切換全選
const toggleAll = (type: 'cash' | 'cheque'): void => {
  const rows = type === 'cash' ? accounting.value.cash_items : accounting.value.cheque_items;
  const target = type === 'cash' ? checkedCash : checkedCheque;
  const ids = rows.map(readPaymentID).filter(Boolean);
  target.value = target.value.length === ids.length ? [] : ids;
};

watch([canViewAccounting, activeBuildingID], () => {
  void loadAccounting();
}, { immediate: true });
</script>

<template>
  <main class="pos-page">
    <section class="pos-page__header">
      <div>
        <p>Payments / Accounting</p>
        <h1>會計</h1>
        <span>{{ contextLabel }}</span>
      </div>
      <BaseButton
        variant="secondary"
        size="md"
        :disabled="isLoading || localLoading || !canViewAccounting"
        @click="loadAccounting"
      >
        {{ localLoading ? '載入中' : '刷新' }}
      </BaseButton>
    </section>

    <section
      v-if="!canViewAccounting"
      class="pos-page__notice"
    >
      <p>此頁面只開放 Staff 查看。</p>
    </section>

    <template v-else>
      <section class="pos-accounting-grid">
        <article class="pos-page__panel">
          <div class="pos-section-head">
            <div>
              <strong>現金待清機</strong>
              <span>HKD {{ cashTotal.toFixed(2) }}</span>
            </div>
            <div class="pos-page__toolbar-actions">
              <BaseButton
                variant="secondary"
                size="sm"
                @click="toggleAll('cash')"
              >
                全選
              </BaseButton>
              <BaseButton
                variant="primary"
                size="sm"
                :disabled="checkedCash.length === 0 || clearingType !== ''"
                @click="clearAccounting('cash')"
              >
                {{ clearingType === 'cash' ? '提交中' : '清機' }}
              </BaseButton>
            </div>
          </div>
          <div
            v-if="accounting.cash_items.length === 0"
            class="pos-page__empty"
          >
            暫時沒有現金待清機交易
          </div>
          <div
            v-else
            class="pos-table-wrap"
          >
            <table class="pos-table pos-table--accounting">
              <thead>
                <tr>
                  <th>選取</th>
                  <th>交易</th>
                  <th>日期</th>
                  <th>金額</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="(row, index) in accounting.cash_items"
                  :key="readPaymentID(row) || String(index)"
                  class="pos-table__select-row"
                  :class="checkedCash.includes(readPaymentID(row)) ? 'pos-table__select-row--active' : ''"
                  @click="toggleChecked('cash', readPaymentID(row))"
                >
                  <td>
                    <input
                      type="checkbox"
                      class="pos-table__checkbox"
                      :checked="checkedCash.includes(readPaymentID(row))"
                      @click.stop="toggleChecked('cash', readPaymentID(row))"
                    >
                  </td>
                  <td>
                    <strong>{{ readPaymentTitle(row, '交易') }}</strong>
                  </td>
                  <td class="pos-table__nowrap">{{ readPaymentDate(row) }}</td>
                  <td class="pos-table__amount">HKD {{ readPaymentAmount(row) }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </article>

        <article class="pos-page__panel">
          <div class="pos-section-head">
            <div>
              <strong>支票待清機</strong>
              <span>HKD {{ chequeTotal.toFixed(2) }}</span>
            </div>
            <div class="pos-page__toolbar-actions">
              <BaseButton
                variant="secondary"
                size="sm"
                @click="toggleAll('cheque')"
              >
                全選
              </BaseButton>
              <BaseButton
                variant="primary"
                size="sm"
                :disabled="checkedCheque.length === 0 || clearingType !== ''"
                @click="clearAccounting('cheque')"
              >
                {{ clearingType === 'cheque' ? '提交中' : '清機' }}
              </BaseButton>
            </div>
          </div>
          <div
            v-if="accounting.cheque_items.length === 0"
            class="pos-page__empty"
          >
            暫時沒有支票待清機交易
          </div>
          <div
            v-else
            class="pos-table-wrap"
          >
            <table class="pos-table pos-table--accounting">
              <thead>
                <tr>
                  <th>選取</th>
                  <th>交易</th>
                  <th>日期</th>
                  <th>金額</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="(row, index) in accounting.cheque_items"
                  :key="readPaymentID(row) || String(index)"
                  class="pos-table__select-row"
                  :class="checkedCheque.includes(readPaymentID(row)) ? 'pos-table__select-row--active' : ''"
                  @click="toggleChecked('cheque', readPaymentID(row))"
                >
                  <td>
                    <input
                      type="checkbox"
                      class="pos-table__checkbox"
                      :checked="checkedCheque.includes(readPaymentID(row))"
                      @click.stop="toggleChecked('cheque', readPaymentID(row))"
                    >
                  </td>
                  <td>
                    <strong>{{ readPaymentTitle(row, '交易') }}</strong>
                  </td>
                  <td class="pos-table__nowrap">{{ readPaymentDate(row) }}</td>
                  <td class="pos-table__amount">HKD {{ readPaymentAmount(row) }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </article>
      </section>

      <section class="pos-page__panel">
        <div class="pos-section-head">
          <div>
            <strong>清機歷史</strong>
            <span>{{ accounting.history_items.length }} 筆記錄</span>
          </div>
        </div>
        <div
          v-if="accounting.history_items.length === 0"
          class="pos-page__empty"
        >
          暫時沒有清機歷史
        </div>
        <div
          v-else
          class="pos-table-wrap"
        >
          <table class="pos-table pos-table--history">
            <thead>
              <tr>
                <th>序號</th>
                <th>時間</th>
                <th>付款方式</th>
                <th>金額</th>
                <th class="pos-table__actions">操作</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="(row, index) in accounting.history_items"
                :key="readPaymentTitle(row, String(index))"
              >
                <td>
                  <strong>{{ readAccountingRecordID(row) || String(index + 1) }}</strong>
                </td>
                <td class="pos-table__nowrap">{{ readAccountingRecordDate(row) }}</td>
                <td>{{ readPaymentMethod(row) }}</td>
                <td class="pos-table__amount">HKD {{ readPaymentAmount(row) }}</td>
                <td class="pos-table__actions">
                  <div class="pos-table__action-group">
                    <BaseButton
                      variant="secondary"
                      size="sm"
                      @click="showRecordDetail(row)"
                    >
                      詳情
                    </BaseButton>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      <div
        v-if="detailRecord"
        class="pos-detail-backdrop"
        @click.self="detailRecord = null"
      >
        <section class="pos-detail-modal">
          <div class="pos-page__header pos-page__header--compact">
            <div>
              <p>Accounting Detail</p>
              <h1>清機詳情</h1>
            </div>
            <BaseButton
              variant="secondary"
              size="sm"
              @click="detailRecord = null"
            >
              關閉詳情
            </BaseButton>
          </div>
          <dl class="pos-detail-list">
            <div
              v-for="field in accountingDetailFields"
              :key="field.label"
            >
              <dt>{{ field.label }}</dt>
              <dd>{{ field.value }}</dd>
            </div>
          </dl>
          <section
            v-if="accountingDetailItems.length > 0"
            class="pos-detail-section"
          >
            <div class="pos-section-head">
              <div>
                <strong>交易明細</strong>
                <span>{{ accountingDetailItems.length }} 筆</span>
              </div>
            </div>
            <article
              v-for="(item, index) in accountingDetailItems"
              :key="readPaymentText(item, ['payment_id', 'receipt_id', 'id']) || String(index)"
              class="pos-mini-row"
            >
              <strong>{{ readPaymentText(item, ['payment_id', 'receipt_id', 'id']) || '交易' }}</strong>
              <dl>
                <div>
                  <dt>收據</dt>
                  <dd>{{ readPaymentText(item, ['receipt_id', 'receiptNo', 'receipt_no']) || '-' }}</dd>
                </div>
                <div>
                  <dt>方式</dt>
                  <dd>{{ readPaymentMethod(item) }}</dd>
                </div>
                <div>
                  <dt>時間</dt>
                  <dd>{{ readPaymentDate(item) }}</dd>
                </div>
                <div>
                  <dt>金額</dt>
                  <dd>HKD {{ readPaymentAmount(item) }}</dd>
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

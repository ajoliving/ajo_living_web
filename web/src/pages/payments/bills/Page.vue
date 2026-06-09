<!--
 * 賬單頁。
 * 1. 讀取目前會員單位的 POS 物業費賬單。
 * 2. 提供線上繳費入口所需的賬單狀態展示。
-->
<script setup lang="ts">
import { computed, ref } from 'vue';
import { useRouter } from 'vue-router';

import { useFeedbackStore } from '@/app/stores/feedback';
import { usePOSPaymentCartStore } from '@/app/stores/pos-payment-cart';
import type { POSPaymentRow } from '@/domains/payments/model';
import BaseButton from '@/shared/components/base/BaseButton.vue';

import { usePOSPaymentPage } from '../composables/usePOSPaymentPage';

const {
  activeBuildingID,
  activeUnitID,
  contextLabel,
  goPaymentUnit,
  isLoading,
  posLoginRequired,
  profileRequired,
  readPaymentAmount,
  readPaymentAmountValue,
  readPaymentDate,
  readPaymentText,
  readPaymentTitle,
  reload,
  rows,
  t,
} = usePOSPaymentPage('bills');

const router = useRouter();
const feedbackStore = useFeedbackStore();
const cartStore = usePOSPaymentCartStore();
const checkoutKey = ref('');
const cartContext = computed(() => ({
  buildingID: activeBuildingID.value,
  unitID: activeUnitID.value,
}));
const hasCartContext = computed(() => activeBuildingID.value !== '' && activeUnitID.value !== '');
const currentCartCount = computed(() => hasCartContext.value ? cartStore.countForContext(cartContext.value) : 0);
const currentCartAmount = computed(() => hasCartContext.value ? cartStore.amountForContext(cartContext.value) : 0);
const selectableRows = computed(() => hasCartContext.value
  ? rows.value.filter((row) => cartStore.billKeyForContext(row, cartContext.value) !== '')
  : []);
const allRowsInCart = computed(() => (
  selectableRows.value.length > 0 && selectableRows.value.every((row) => cartStore.has(row, cartContext.value))
));

// 1. 讀取賬單的單筆支付金額
const readBillAmountCents = (row: POSPaymentRow): number =>
  Math.round(readPaymentAmountValue(row) * 100);

// 2. 切換購物車狀態
const toggleCart = (row: POSPaymentRow): void => {
  const key = cartStore.billKeyForContext(row, cartContext.value);
  if (!key) {
    return;
  }
  if (cartStore.has(row, cartContext.value)) {
    cartStore.remove(row, cartContext.value);
    feedbackStore.pushToast('已取消加入購物車', 'success');
    return;
  }
  cartStore.add(row, cartContext.value);
  feedbackStore.pushToast('已加入購物車', 'success');
};

// 3. 全選或取消全部賬單
const toggleAllRows = (): void => {
  if (allRowsInCart.value) {
    cartStore.removeByKeys(selectableRows.value.map((row) => cartStore.billKeyForContext(row, cartContext.value)));
    feedbackStore.pushToast('已取消全部賬單', 'success');
    return;
  }
  selectableRows.value.forEach((row) => cartStore.add(row, cartContext.value));
  feedbackStore.pushToast('已將全部賬單加入購物車', 'success');
};

// 4. 前往購物車
const goCart = async (): Promise<void> => {
  await router.push('/payments/cart');
};

// 5. 進入單筆結賬
const handlePayBill = async (row: POSPaymentRow, index: number): Promise<void> => {
  const title = readPaymentTitle(row, String(index));
  const key = cartStore.billKeyForContext(row, cartContext.value);
  const amountCents = readBillAmountCents(row);
  if (!key || amountCents <= 0) {
    return;
  }

  checkoutKey.value = title;
  try {
    if (!cartStore.has(row, cartContext.value)) {
      cartStore.add(row, cartContext.value);
    }
    await router.push({
      path: '/payments/cart',
      query: { checkout_bill_key: key },
    });
  } catch {
    feedbackStore.pushToast(t('payments.common.createOrderError'), 'error');
  } finally {
    checkoutKey.value = '';
  }
};
</script>

<template>
  <main class="pos-page">
    <section class="pos-page__header">
      <div>
        <p>Payments / Bills</p>
        <h1>賬單</h1>
        <span>{{ contextLabel }}</span>
      </div>
      <BaseButton
        variant="secondary"
        size="md"
        :disabled="isLoading"
        @click="reload"
      >
        {{ isLoading ? '載入中' : '刷新' }}
      </BaseButton>
    </section>

    <section
      v-if="!profileRequired && !posLoginRequired && rows.length > 0"
      class="pos-page__toolbar"
    >
      <div>
        <strong>{{ currentCartCount }} 筆已加入購物車</strong>
        <span>HKD {{ currentCartAmount.toFixed(2) }}</span>
      </div>
      <div class="pos-page__toolbar-actions">
        <BaseButton
          variant="secondary"
          size="sm"
          :disabled="selectableRows.length === 0"
          @click="toggleAllRows"
        >
          {{ allRowsInCart ? '取消全部' : '全選加入' }}
        </BaseButton>
        <BaseButton
          variant="primary"
          size="sm"
          :disabled="currentCartCount === 0"
          @click="goCart"
        >
          前往購物車
        </BaseButton>
      </div>
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
      v-else-if="rows.length === 0"
      class="pos-page__empty"
    >
      {{ isLoading ? '載入中' : '暫時沒有待繳賬單' }}
    </section>

    <section
      v-else
      class="pos-table-wrap"
    >
      <table class="pos-table pos-table--bills">
        <thead>
          <tr>
            <th>賬單</th>
            <th>期數</th>
            <th>賬單日</th>
            <th>狀態</th>
            <th>金額</th>
            <th class="pos-table__actions">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="(row, index) in rows"
            :key="readPaymentTitle(row, String(index))"
          >
            <td>
              <span class="pos-table__primary">
                <strong>{{ readPaymentTitle(row, '賬單') }}</strong>
                <small>{{ readPaymentText(row, ['item_name', 'item_id', 'name']) || '物業費' }}</small>
              </span>
            </td>
            <td>{{ readPaymentText(row, ['trs_to', 'term', 'period']) || '-' }}</td>
            <td class="pos-table__nowrap">{{ readPaymentDate(row) }}</td>
            <td>
              <span class="pos-table__status">{{ readPaymentText(row, ['status', 'state']) || 'pending' }}</span>
            </td>
            <td class="pos-table__amount">HKD {{ readPaymentAmount(row) }}</td>
            <td class="pos-table__actions">
              <div class="pos-table__action-group">
                <BaseButton
                  variant="secondary"
                  size="sm"
                  :disabled="cartStore.billKeyForContext(row, cartContext) === ''"
                  @click="toggleCart(row)"
                >
                  {{ cartStore.has(row, cartContext) ? '取消加入' : '加入購物車' }}
                </BaseButton>
                <BaseButton
                  variant="primary"
                  size="sm"
                  :disabled="checkoutKey !== '' || readBillAmountCents(row) <= 0"
                  @click="handlePayBill(row, index)"
                >
                  {{ checkoutKey === readPaymentTitle(row, String(index)) ? '前往結賬中' : '單筆結賬' }}
                </BaseButton>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </section>
  </main>
</template>

<style scoped>
@import '../payment-page.css';
</style>

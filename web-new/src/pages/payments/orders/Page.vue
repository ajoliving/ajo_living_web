<!--
 * POS 訂單頁。
 * 1. 展示線上繳費訂單的支付方式、金額、狀態與建立時間。
 * 2. 支援付款二維碼、付款鏈接與自動刷新支付狀態。
-->
<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue';
import { useRoute } from 'vue-router';

import { useFeedbackStore } from '@/stores/feedback';
import {
  fetchPOSPaymentOrder,
  queryPOSPaymentOrder,
} from '@/httpapis/payments';
import type { POSPaymentRow } from '@/model/payments';
import BaseButton from '@/shared/components/base/BaseButton.vue';
import QrCodeImage from '@/shared/components/base/QrCodeImage.vue';

import { usePOSPaymentPage } from '../composables/usePOSPaymentPage';

const route = useRoute();
const feedbackStore = useFeedbackStore();
const {
  activeBuildingID,
  activeUnitID,
  contextLabel,
  goPaymentUnit,
  isLoading,
  posLoginRequired,
  profileRequired,
  readPaymentCentsAmount,
  readPaymentDate,
  readPaymentText,
  readPaymentTitle,
  reload,
  rows,
} = usePOSPaymentPage('orders');

const focusedOrder = ref<POSPaymentRow | null>(null);
const expandedQrOrderKey = ref('');
const actionOrderNo = ref('');
let orderPollingTimer: number | null = null;

const routeMchOrderNo = computed(() => String(route.query.mch_order_no ?? route.query.mchOrderNo ?? '').trim());
const routePayOrderId = computed(() => String(route.query.pay_order_id ?? route.query.payOrderId ?? '').trim());
const routePayChannel = computed(() => String(route.query.pay_channel ?? route.query.payChannel ?? '').trim());
const routePaymentMethodKey = computed(() => String(route.query.payment_method_key ?? route.query.paymentMethodKey ?? '').trim());
const routePaymentLabel = computed(() => String(route.query.payment_label ?? route.query.paymentLabel ?? '').trim());

const paymentLabelMap: Record<string, string> = {
  POS_WECHAT: '微信支付',
  WX_H5: '微信支付',
  WX_QR: '微信支付',
  POS_ALIPAY: '支付寶',
  POS_ALIPAY_CN: '支付寶（大陸）',
  POS_ALIPAY_HK: '支付寶香港',
  ALI_H5: '支付寶',
  ALI_QR: '支付寶',
  POS_YSF_QR: '雲閃付',
  YSF_QR: '雲閃付',
};

const paymentShortLabelMap: Record<string, string> = {
  POS_WECHAT: '微信',
  WX_H5: '微信',
  WX_QR: '微信',
  POS_ALIPAY: '支付寶',
  POS_ALIPAY_CN: '支付寶',
  POS_ALIPAY_HK: 'AlipayHK',
  ALI_H5: '支付寶',
  ALI_QR: '支付寶',
  POS_YSF_QR: '雲閃付',
  YSF_QR: '雲閃付',
};

// 1. 讀取訂單資料
const readOrderData = (payload: Record<string, unknown>): POSPaymentRow => {
  const nested = payload.data;
  if (nested && typeof nested === 'object' && !Array.isArray(nested)) {
    return nested as POSPaymentRow;
  }
  return payload;
};

// 2. 讀取訂單號
const readOrderNo = (row: POSPaymentRow): string =>
  readPaymentText(row, ['mch_order_no', 'mchOrderNo']) || readPaymentTitle(row, '');

// 3. 讀取支付單號
const readPayOrderId = (row: POSPaymentRow): string =>
  readPaymentText(row, ['pay_order_id', 'payOrderId']);

// 4. 讀取訂單行 key
const readOrderRowKey = (row: POSPaymentRow, index = 0): string =>
  readOrderNo(row) || readPayOrderId(row) || readPaymentTitle(row, String(index));

// 5. 讀取支付資料類型
const readPayDataType = (row: POSPaymentRow): string =>
  readPaymentText(row, ['pay_data_type', 'payDataType']).toLowerCase();

// 6. 讀取支付資料
const readPayData = (row: POSPaymentRow): string =>
  readPaymentText(row, ['pay_data', 'payData']);

// 7. 正規化支付方式 key
const normalizePaymentKey = (value: unknown): string =>
  String(value ?? '').trim().toUpperCase();

// 8. 判斷路由查詢是否指向目前訂單
const isRouteOrder = (row: POSPaymentRow): boolean => {
  const orderNo = readOrderNo(row);
  const payOrderId = readPayOrderId(row);
  return Boolean(
    (routeMchOrderNo.value && orderNo === routeMchOrderNo.value) ||
    (routePayOrderId.value && payOrderId === routePayOrderId.value),
  );
};

// 9. 讀取支付方式 key
const readOrderPaymentKey = (row: POSPaymentRow): string => {
  if (isRouteOrder(row) && routePaymentMethodKey.value) {
    return normalizePaymentKey(routePaymentMethodKey.value);
  }
  const rowKey = readPaymentText(row, ['payment_method_key', 'paymentMethodKey', 'pay_channel', 'payChannel']);
  if (rowKey) {
    return normalizePaymentKey(rowKey);
  }
  if (isRouteOrder(row)) {
    return normalizePaymentKey(routePaymentMethodKey.value || routePayChannel.value);
  }
  return '';
};

// 10. 讀取支付方式文字
const readOrderPaymentLabel = (row: POSPaymentRow): string => {
  if (isRouteOrder(row) && routePaymentLabel.value) {
    return routePaymentLabel.value;
  }
  const rawLabel = readPaymentText(row, ['payment_label', 'paymentLabel', 'pay_method_text', 'pay_method_label']);
  if (rawLabel) {
    return rawLabel;
  }
  const key = readOrderPaymentKey(row);
  const routeKey = isRouteOrder(row) ? normalizePaymentKey(routePayChannel.value) : '';
  return paymentLabelMap[key] ?? paymentLabelMap[routeKey] ?? '線上支付';
};

// 11. 讀取支付方式短文字
const readOrderPaymentShortLabel = (row: POSPaymentRow): string => {
  const key = readOrderPaymentKey(row);
  if (paymentShortLabelMap[key]) {
    return paymentShortLabelMap[key];
  }
  const label = readOrderPaymentLabel(row);
  if (label.includes('微信')) {
    return '微信';
  }
  if (label.includes('香港') || label.toLowerCase().includes('alipayhk')) {
    return 'AlipayHK';
  }
  if (label.includes('支付寶')) {
    return '支付寶';
  }
  return label;
};

// 12. 讀取訂單狀態 key
const readOrderStateKey = (row: POSPaymentRow): string =>
  readPaymentText(row, ['state', 'status', 'gateway_state_code', 'business_state']).toLowerCase();

// 13. 判斷訂單是否支付中
const isPendingOrder = (row: POSPaymentRow): boolean => {
  const state = readOrderStateKey(row);
  const businessState = readPaymentText(row, ['business_state']).toLowerCase();
  if (['success', 'succeeded', 'paid', 'closed', 'cancelled', 'canceled', 'failed'].includes(state)) {
    return false;
  }
  return state === '' || ['paying', 'pending', 'created', '0'].includes(state) || businessState === 'pending';
};

// 14. 讀取訂單狀態文字
const readOrderStatusLabel = (row: POSPaymentRow): string => {
  const state = readOrderStateKey(row);
  if (['success', 'succeeded', 'paid'].includes(state)) {
    return '支付成功';
  }
  if (isPendingOrder(row)) {
    return '支付中';
  }
  return readPaymentText(row, ['state_label', 'status_label', 'state', 'status']) || '待確認';
};

// 15. 讀取系統訊息
const readOrderSystemMessage = (row: POSPaymentRow): string => {
  const direct = readPaymentText(row, ['gateway_message', 'gatewayMessage', 'business_message', 'businessMessage', 'message', 'msg']);
  if (direct) {
    return direct;
  }
  for (const key of ['gateway_create_response', 'gateway_query_response', 'business_response']) {
    const value = row[key];
    if (value && typeof value === 'object' && !Array.isArray(value)) {
      const nested = value as POSPaymentRow;
      const nestedText = readPaymentText(nested, ['message', 'msg', 'retCode', 'ret_code', 'code', 'result_code', 'errCode', 'err_code', 'state', 'status']);
      if (nestedText) {
        return nestedText;
      }
    }
  }
  return '';
};

// 16. 格式化訂單時間
const formatOrderDate = (row: POSPaymentRow): string => {
  const raw = readPaymentText(row, ['created_at', 'createdAt', 'tran_time', 'tran_datetime']) || readPaymentDate(row);
  const date = new Date(raw);
  if (Number.isNaN(date.getTime())) {
    return raw || '-';
  }
  const pad = (value: number): string => String(value).padStart(2, '0');
  return `${pad(date.getDate())}/${pad(date.getMonth() + 1)}/${date.getFullYear()} ${pad(date.getHours())}:${pad(date.getMinutes())}`;
};

// 17. 判斷是否支付寶訂單
const isAlipayOrder = (row: POSPaymentRow): boolean => {
  const key = readOrderPaymentKey(row);
  const channel = normalizePaymentKey(readPaymentText(row, ['pay_channel', 'payChannel']) || routePayChannel.value);
  return key.includes('ALIPAY') || key.startsWith('ALI_') || channel.startsWith('ALI_');
};

// 18. 讀取二維碼圖片
const readQrImageUrl = (row: POSPaymentRow): string => {
  if (!isPendingOrder(row)) {
    return '';
  }
  return readPayDataType(row) === 'codeimgurl' ? readPayData(row) : '';
};

// 19. 讀取二維碼鏈接
const readQrLinkUrl = (row: POSPaymentRow): string => {
  if (!isPendingOrder(row)) {
    return '';
  }
  const payDataType = readPayDataType(row);
  if (payDataType === 'codeurl') {
    return readPayData(row);
  }
  if (payDataType === 'payurl' && isAlipayOrder(row)) {
    return readPayData(row);
  }
  return '';
};

// 20. 判斷是否有二維碼面板
const hasQrPanel = (row: POSPaymentRow): boolean =>
  Boolean(readQrImageUrl(row) || readQrLinkUrl(row));

// 21. 判斷二維碼面板是否展開
const shouldShowQrPanel = (row: POSPaymentRow, index: number): boolean =>
  expandedQrOrderKey.value === readOrderRowKey(row, index) && hasQrPanel(row);

// 22. 展開或收起二維碼
const toggleQrPanel = (row: POSPaymentRow, index: number): void => {
  const key = readOrderRowKey(row, index);
  expandedQrOrderKey.value = expandedQrOrderKey.value === key ? '' : key;
};

// 23. 查詢最新訂單
const queryLatestOrder = async (row: POSPaymentRow): Promise<POSPaymentRow | null> => {
  const mchOrderNo = readOrderNo(row);
  const payOrderId = readPayOrderId(row);
  if (!mchOrderNo && !payOrderId) {
    return null;
  }
  const params = {
    building_id: activeBuildingID.value,
    unit_id: activeUnitID.value,
  };
  const response = mchOrderNo
    ? await fetchPOSPaymentOrder(mchOrderNo, { ...params, detail: true, refresh: true, retry_business: true, real_gateway: true })
    : await queryPOSPaymentOrder({ ...params, pay_order_id: payOrderId, retry_business: true, real_gateway: true });
  return readOrderData(response.data.data);
};

// 24. 載入路由指定訂單
const loadRouteOrder = async (): Promise<void> => {
  if (!routeMchOrderNo.value && !routePayOrderId.value) {
    focusedOrder.value = null;
    return;
  }
  try {
    const payload = routeMchOrderNo.value
      ? {
          mch_order_no: routeMchOrderNo.value,
          building_id: activeBuildingID.value,
          unit_id: activeUnitID.value,
          retry_business: true,
          real_gateway: true,
        }
      : {
          pay_order_id: routePayOrderId.value,
          building_id: activeBuildingID.value,
          unit_id: activeUnitID.value,
          retry_business: true,
          real_gateway: true,
        };
    const order = readOrderData((await queryPOSPaymentOrder(payload)).data.data);
    focusedOrder.value = order;
    if (hasQrPanel(order)) {
      expandedQrOrderKey.value = readOrderRowKey(order);
    }
  } catch {
    feedbackStore.pushToast('訂單狀態查詢失敗', 'error');
  }
};

// 25. 刷新單一訂單
const refreshOrder = async (row: POSPaymentRow, silent = false): Promise<void> => {
  const orderNo = readOrderNo(row) || readPayOrderId(row);
  actionOrderNo.value = silent ? '' : orderNo;
  try {
    const latest = await queryLatestOrder(row);
    if (latest) {
      focusedOrder.value = latest;
      if (hasQrPanel(latest) && !expandedQrOrderKey.value) {
        expandedQrOrderKey.value = readOrderRowKey(latest);
      }
    }
    await reload();
    if (!silent) {
      feedbackStore.pushToast('訂單狀態已刷新', 'success');
    }
  } catch {
    if (!silent) {
      feedbackStore.pushToast('刷新訂單失敗', 'error');
    }
  } finally {
    actionOrderNo.value = '';
  }
};

// 26. 靜默刷新待支付訂單
const refreshPendingOrders = async (): Promise<void> => {
  if (actionOrderNo.value) {
    return;
  }
  const pendingOrder = displayRows.value.find(isPendingOrder);
  if (!pendingOrder) {
    return;
  }
  await refreshOrder(pendingOrder, true);
};

// 27. 渲染訂單列表
const displayRows = computed(() => {
  const orderNo = focusedOrder.value ? readOrderNo(focusedOrder.value) : '';
  if (!focusedOrder.value || !orderNo) {
    return rows.value;
  }
  return [
    focusedOrder.value,
    ...rows.value.filter((row) => readOrderNo(row) !== orderNo),
  ];
});

onMounted(() => {
  void loadRouteOrder();
  orderPollingTimer = window.setInterval(() => {
    void refreshPendingOrders();
  }, 8000);
});

onUnmounted(() => {
  if (orderPollingTimer) {
    window.clearInterval(orderPollingTimer);
    orderPollingTimer = null;
  }
});

watch([routeMchOrderNo, routePayOrderId, activeBuildingID, activeUnitID], () => {
  void loadRouteOrder();
});
</script>

<template>
  <main class="pos-page">
    <section class="pos-page__header">
      <div>
        <p>Payments / Orders</p>
        <h1>訂單</h1>
        <span>{{ contextLabel }}</span>
      </div>
      <BaseButton
        variant="secondary"
        size="sm"
        :disabled="isLoading"
        @click="reload"
      >
        {{ isLoading ? '載入中' : '刷新' }}
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

    <section
      v-else-if="displayRows.length === 0"
      class="pos-page__empty"
    >
      {{ isLoading ? '載入中' : '暫時沒有線上繳費訂單' }}
    </section>

    <section
      v-else
      class="pos-order-list"
    >
      <article
        v-for="(row, index) in displayRows"
        :key="readOrderRowKey(row, index)"
        class="pos-order-card"
      >
        <div class="pos-order-card__top">
          <div>
            <strong>{{ readOrderPaymentLabel(row) }}</strong>
          </div>
          <div class="pos-order-card__amount">
            <span>金額</span>
            <b>HKD {{ readPaymentCentsAmount(row) }}</b>
          </div>
        </div>

        <div class="pos-order-card__meta">
          <span class="pos-table__status">{{ readOrderStatusLabel(row) }}</span>
          <span
            v-if="isPendingOrder(row)"
            class="pos-order-card__pill"
          >
            3 分鐘未支付自動刪除
          </span>
          <time>{{ formatOrderDate(row) }}</time>
        </div>

        <section
          v-if="readOrderSystemMessage(row)"
          class="pos-order-card__system"
        >
          <span>系統資訊</span>
          <strong>{{ readOrderSystemMessage(row) }}</strong>
        </section>

        <section
          v-if="shouldShowQrPanel(row, index)"
          class="pos-order-card__qr"
        >
          <strong>{{ readOrderPaymentLabel(row) }}付款二維碼</strong>
          <img
            v-if="readQrImageUrl(row)"
            :src="readQrImageUrl(row)"
            :alt="`${readOrderPaymentLabel(row)}二維碼`"
          >
          <QrCodeImage
            v-else
            :text="readQrLinkUrl(row)"
            :alt="`${readOrderPaymentLabel(row)}付款二維碼`"
            :size="260"
          />
          <a
            v-if="readQrLinkUrl(row)"
            :href="readQrLinkUrl(row)"
            target="_blank"
            rel="noreferrer"
          >
            複製 / 打開付款鏈接
          </a>
          <p>請使用手機上的 {{ readOrderPaymentShortLabel(row) }} 掃碼付款，支付後可點擊刷新狀態確認結果。</p>
        </section>

        <div class="pos-order-card__actions">
          <BaseButton
            v-if="hasQrPanel(row)"
            variant="secondary"
            size="sm"
            :disabled="actionOrderNo !== ''"
            @click="toggleQrPanel(row, index)"
          >
            {{ shouldShowQrPanel(row, index) ? '收起二維碼' : '查看二維碼' }}
          </BaseButton>
          <BaseButton
            variant="secondary"
            size="sm"
            :disabled="actionOrderNo !== ''"
            @click="refreshOrder(row)"
          >
            刷新
          </BaseButton>
        </div>
      </article>
    </section>
  </main>
</template>

<style scoped>
@import '../payment-page.css';
</style>

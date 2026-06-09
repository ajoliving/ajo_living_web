/*
 * POS 物業費購物車狀態。
 * 1. 使用 Pinia 保存已加入的 POS 賬單。
 * 2. 使用目前大廈與單位隔離購物車資料。
 * 3. 使用 localStorage 在刷新頁面後保留購物車。
 * 4. 提供金額、手續費與移除操作。
 */
import { computed, ref } from 'vue';
import { defineStore } from 'pinia';

import type { POSPaymentRow } from '@/domains/payments/model';

const STORAGE_KEY = 'ajo_pos_payment_cart';
const CART_BUILDING_KEY = '_ajo_payment_building_id';
const CART_UNIT_KEY = '_ajo_payment_unit_id';

interface POSPaymentCartContext {
  buildingID?: string;
  unitID?: string;
}

// 1. 讀取本地購物車
const readStoredItems = (): POSPaymentRow[] => {
  if (typeof window === 'undefined') {
    return [];
  }
  try {
    const raw = window.localStorage.getItem(STORAGE_KEY);
    if (!raw) {
      return [];
    }
    const parsed = JSON.parse(raw);
    return Array.isArray(parsed) ? parsed : [];
  } catch {
    return [];
  }
};

// 2. 正規化賬單 key
const normalizeKey = (value: unknown): string => String(value ?? '').trim();

// 3. 組合賬單備用 key
const joinBillKeyParts = (parts: unknown[]): string =>
  parts.map(normalizeKey).filter(Boolean).join(':');

// 4. 讀取賬單原始 key
const readRawBillKey = (item: POSPaymentRow): string =>
  normalizeKey(item.invoice_no ?? item.bill_no ?? item.bill_number ?? item.bill_id ?? item.id) ||
  joinBillKeyParts([
    item.item_id ?? item.item_name ?? item.name,
    item.trs_to ?? item.term ?? item.period,
    item.bill_dt ?? item.bill_date,
    item.net_amount ?? item.amount ?? item.total ?? item.total_amount ?? item.payable,
  ]);

// 5. 讀取賬單所屬大廈
const readCartBuildingID = (item: POSPaymentRow): string =>
  normalizeKey(item[CART_BUILDING_KEY] ?? item.building_id ?? item.BLG_ID ?? item.blg_id);

// 6. 讀取賬單所屬單位
const readCartUnitID = (item: POSPaymentRow): string =>
  normalizeKey(item[CART_UNIT_KEY] ?? item.unit_id ?? item.UNIT_ID);

// 7. 讀取賬單 key
const readCartBillKey = (item: POSPaymentRow): string => {
  const billKey = readRawBillKey(item);
  if (!billKey) {
    return '';
  }
  return [
    readCartBuildingID(item),
    readCartUnitID(item),
    billKey,
  ].map(normalizeKey).join('|');
};

// 8. 加入目前單位上下文
const attachCartContext = (item: POSPaymentRow, context?: POSPaymentCartContext): POSPaymentRow => ({
  ...item,
  [CART_BUILDING_KEY]: normalizeKey(context?.buildingID ?? readCartBuildingID(item)),
  [CART_UNIT_KEY]: normalizeKey(context?.unitID ?? readCartUnitID(item)),
});

// 9. 判斷賬單是否屬於目前單位
const isCartItemForContext = (item: POSPaymentRow, context?: POSPaymentCartContext): boolean => {
  const buildingID = normalizeKey(context?.buildingID);
  const unitID = normalizeKey(context?.unitID);
  if (!buildingID && !unitID) {
    return true;
  }
  return readCartBuildingID(item) === buildingID && readCartUnitID(item) === unitID;
};

// 10. 讀取賬單金額
const readCartAmount = (item: POSPaymentRow): number => {
  const value = Number(item.net_amount ?? item.amount ?? item.total ?? item.total_amount ?? item.payable ?? 0);
  return Number.isFinite(value) && value > 0 ? value : 0;
};

// 11. 建立 POS 購物車 store
export const usePOSPaymentCartStore = defineStore('posPaymentCart', () => {
  const items = ref<POSPaymentRow[]>(readStoredItems());

  const count = computed(() => items.value.length);
  const amount = computed(() => items.value.reduce((sum, item) => sum + readCartAmount(item), 0));

  // 11.1 持久化購物車
  const persist = (): void => {
    if (typeof window === 'undefined') {
      return;
    }
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify(items.value));
  };

  // 11.2 讀取目前單位賬單
  const itemsForContext = (context?: POSPaymentCartContext): POSPaymentRow[] =>
    items.value.filter((item) => isCartItemForContext(item, context));

  // 11.3 計算目前單位金額
  const amountForContext = (context?: POSPaymentCartContext): number =>
    itemsForContext(context).reduce((sum, item) => sum + readCartAmount(item), 0);

  // 11.4 計算目前單位數量
  const countForContext = (context?: POSPaymentCartContext): number =>
    itemsForContext(context).length;

  // 11.5 讀取目前單位賬單 key
  const billKeyForContext = (item: POSPaymentRow, context?: POSPaymentCartContext): string =>
    readCartBillKey(attachCartContext(item, context));

  // 11.6 判斷賬單是否已加入
  const has = (item: POSPaymentRow, context?: POSPaymentCartContext): boolean =>
    Boolean(billKeyForContext(item, context)) &&
    items.value.some((current) => readCartBillKey(current) === billKeyForContext(item, context));

  // 11.7 加入賬單
  const add = (item: POSPaymentRow, context?: POSPaymentCartContext): void => {
    const nextItem = attachCartContext(item, context);
    const key = readCartBillKey(nextItem);
    if (!key || has(item, context)) {
      return;
    }
    items.value = [...items.value, nextItem];
    persist();
  };

  // 11.8 移除賬單
  const removeByKey = (key: string): void => {
    const normalized = normalizeKey(key);
    if (!normalized) {
      return;
    }
    items.value = items.value.filter((item) => readCartBillKey(item) !== normalized);
    persist();
  };

  // 11.9 批量移除賬單
  const removeByKeys = (keys: string[]): void => {
    const keySet = new Set(keys.map(normalizeKey).filter(Boolean));
    if (keySet.size === 0) {
      return;
    }
    items.value = items.value.filter((item) => !keySet.has(readCartBillKey(item)));
    persist();
  };

  // 11.10 移除指定單位賬單
  const remove = (item: POSPaymentRow, context?: POSPaymentCartContext): void => {
    removeByKey(billKeyForContext(item, context));
  };

  // 11.11 清空購物車
  const clear = (): void => {
    items.value = [];
    persist();
  };

  // 11.12 清空目前單位購物車
  const clearContext = (context?: POSPaymentCartContext): void => {
    items.value = items.value.filter((item) => !isCartItemForContext(item, context));
    persist();
  };

  // 11.13 計算手續費
  const feeAmount = (feeRate: number, context?: POSPaymentCartContext): number => {
    if (!Number.isFinite(feeRate) || feeRate <= 0) {
      return 0;
    }
    return Math.ceil(amountForContext(context) * feeRate * 100 - 1e-8) / 100;
  };

  return {
    add,
    amount,
    amountForContext,
    billKey: readCartBillKey,
    billKeyForContext,
    clear,
    clearContext,
    count,
    countForContext,
    feeAmount,
    has,
    items,
    itemsForContext,
    remove,
    removeByKey,
    removeByKeys,
  };
});

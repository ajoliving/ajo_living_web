/*
 * POS 物業繳費型別。
 * 1. 定義 AJO 後端聚合後的 POS 單位上下文。
 * 2. 定義賬單、購物車、訂單、會計與歷史列表的鬆散資料結構。
 */

// 1. POS 單位上下文
export interface POSPaymentContext {
  building_id: string;
  building_name: string;
  unit_id: string;
  floor: string;
  unit: string;
  unit_label: string;
}

// 2. POS 繳費概覽
export interface POSPaymentOverviewSummary {
  pending_bill_count: number;
  pending_bill_amount: number;
  order_count: number;
  pending_order_count: number;
  abnormal_order_count: number;
}

// 3. POS 繳費概覽
export interface POSPaymentOverview {
  context?: POSPaymentContext;
  building_options: string[];
  unit_options: string[];
  is_staff: boolean;
  profile_required: boolean;
  pos_login_required: boolean;
  terminal_proxy_enabled: boolean;
  summary?: POSPaymentOverviewSummary;
}

// 4. POS 通用資料行
export type POSPaymentRow = Record<string, unknown>;

// 5. POS 通用列表回應
export interface POSPaymentListPayload {
  context?: POSPaymentContext;
  building_options?: string[];
  unit_options?: string[];
  items: POSPaymentRow[];
}

// 6. POS H5 支付訂單建立請求
export interface POSPaymentOrderCreatePayload {
  scene?: 'billing' | 'cart';
  building_id?: string;
  unit_id?: string;
  pay_channel: POSPaymentChannel;
  expire_seconds?: number;
  final_amount: number;
  handle_fee_amount?: number;
  bill_objs: POSPaymentRow[];
  handle_fee_obj?: POSPaymentRow[];
  return_path?: string;
  remark?: string;
  gateway_request_overrides?: POSPaymentRow;
}

// 7. POS 支付通道
export type POSPaymentChannel = 'WX_H5' | 'ALI_H5' | 'WX_QR' | 'ALI_QR' | 'YSF_QR';

// 8. POS 支付方式
export type POSPaymentMethodKey =
  | 'POS_CARD'
  | 'POS_WECHAT'
  | 'POS_ALIPAY'
  | 'POS_YSF_QR'
  | 'POS_BANK'
  | 'POS_CHEQUE'
  | 'POS_CASH';

// 9. POS 歷史查詢
export interface POSPaymentHistoryQuery {
  building_id?: string;
  unit_id?: string;
  from_date?: string;
  to_date?: string;
  date_type?: 'input_date' | 'tran_date';
  pay_method?: string;
  unit_ids?: string;
  unit_id_list?: string[];
}

// 10. POS 線下繳費請求
export interface POSPaymentReportPayload {
  building_id?: string;
  unit_id?: string;
  FINAL_AMOUNT: number;
  ENTRY_DATETIME: string;
  TRAN_DATETIME: string;
  TRAN_REF_NO?: string;
  COMMENT?: string;
  PIC_FILENAME?: string[];
  PIC_DATA?: string[];
  BILL_OBJS: POSPaymentRow[];
  PAY_METHOD: string;
  BLG_ID?: string;
  UNIT_ID?: string;
  bank_account_received?: string;
}

// 11. POS 機收款請求
export interface POSTerminalPaymentPayload {
  building_id?: string;
  unit_id?: string;
  pay_type: 'POS_CARD';
  final_amount: number;
  bill_objs: POSPaymentRow[];
  handle_fee_obj?: POSPaymentRow[];
  remark?: string;
}

// 12. POS 會計資料
export interface POSPaymentAccountingPayload {
  context?: POSPaymentContext;
  building_options: string[];
  unit_options: string[];
  cash_items: POSPaymentRow[];
  cheque_items: POSPaymentRow[];
  history_items: POSPaymentRow[];
  items: POSPaymentRow[];
}

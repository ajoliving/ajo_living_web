/*
 * POS 物業繳費型別。
 * 1. 定義 AJO 後端聚合後的 POS 單位上下文。
 * 2. 定義 AJO Pay 首頁與付款頁需要的賬單、付款設定與訂單資料結構。
 * 3. 定義 iSmart integration 未繳賬單與交易紀錄資料結構。
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

// 2. POS 通用資料行
export type POSPaymentRow = Record<string, unknown>;

// 3. POS 通用列表回應
export interface POSPaymentListPayload {
  context?: POSPaymentContext;
  building_options?: string[];
  unit_options?: string[];
  items: POSPaymentRow[];
}

// 4. POS H5 支付訂單建立請求
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

// 5. POS 支付通道
export type POSPaymentChannel = 'WX_H5' | 'ALI_H5' | 'WX_QR' | 'ALI_QR' | 'YSF_QR';

// 6. POS 支付方式
export type POSPaymentMethodKey =
  | 'POS_CARD'
  | 'POS_WECHAT'
  | 'POS_ALIPAY'
  | 'POS_YSF_QR'
  | 'POS_BANK'
  | 'POS_CHEQUE'
  | 'POS_CASH';

// 7. POS 線下繳費請求
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

// 8. iSmart integration 未繳賬單
export interface POSIntegrationUnpaidInvoice {
  invoice_no?: string;
  flat_code?: string;
  item_id?: string;
  trs_to?: string;
  bill_dt?: string;
  net_amount?: number;
  remark?: string;
  [field: string]: string | number | null | undefined;
}

// 9. iSmart integration 交易明細
export interface POSIntegrationPaymentDetail {
  floor?: string;
  unit?: string;
  flat_code?: string;
  item_id?: string;
  term?: string;
  trs_val?: number;
  remark?: string;
  [field: string]: string | number | null | undefined;
}

// 10. iSmart integration 交易紀錄
export interface POSIntegrationPaymentTransaction {
  payment_id?: string;
  input_time?: string;
  tran_time?: string;
  trs_val?: number;
  receipt_id?: string;
  ref_no?: string;
  pay_type?: string;
  status?: string;
  payment_detail_objs?: POSIntegrationPaymentDetail[];
  [field: string]: string | number | POSIntegrationPaymentDetail[] | null | undefined;
}

// 11. iSmart integration 交易紀錄回應
export interface POSIntegrationTransactionsPayload {
  payment_objs?: POSIntegrationPaymentTransaction[];
}

// 12. iSmart integration 日期查詢條件
export interface POSIntegrationTransactionDateQuery {
  building_id: string;
  from_date: string;
  to_date: string;
  date_type: 'input_date' | 'tran_date';
  pay_method?: 'all' | 'pos';
}

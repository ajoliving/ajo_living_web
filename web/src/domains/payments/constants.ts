/*
 * 支付 domain 常量入口。
 * 1. 保持物業費、賬單、訂單與 POS 展示常量集中出口。
 * 2. 提供 POS 支付方式代碼與顯示文字的語系映射。
 */

export type POSPaymentMethodI18nKey =
  | 'card'
  | 'unionpay'
  | 'wechat'
  | 'alipay'
  | 'mainlandAlipay'
  | 'alipayHK'
  | 'wechatAlipay'
  | 'bankTransfer'
  | 'cheque'
  | 'cash';

const posPaymentMethodCodeMap: Record<string, POSPaymentMethodI18nKey> = {
  POS_CARD: 'card',
  WEBPOS_CARD: 'card',
  CARD: 'card',
  POS_YSF_QR: 'unionpay',
  YSF_QR: 'unionpay',
  POS_UP_OP: 'unionpay',
  UP_OP: 'unionpay',
  WEBPOS_CARD_UP: 'unionpay',
  WEBPOS_UP_OP: 'unionpay',
  POS_WECHAT: 'wechat',
  WEBPOS_WECHAT: 'wechat',
  WX_H5: 'wechat',
  WX_QR: 'wechat',
  WECHAT: 'wechat',
  POS_ALIPAY: 'alipay',
  WEBPOS_ALIPAY: 'alipay',
  ALI_H5: 'alipay',
  ALI_QR: 'alipay',
  ALIPAY: 'alipay',
  POS_ALIPAY_CN: 'mainlandAlipay',
  ALIPAY_CN: 'mainlandAlipay',
  POS_ALIPAY_HK: 'alipayHK',
  ALIPAY_HK: 'alipayHK',
  POS_ALIWE: 'wechatAlipay',
  ALIWE: 'wechatAlipay',
  POS_BANK: 'bankTransfer',
  WEBPOS_BANK: 'bankTransfer',
  BANK: 'bankTransfer',
  POS_CHEQUE: 'cheque',
  WEBPOS_CHEQUE: 'cheque',
  CHEQUE: 'cheque',
  POS_CASH: 'cash',
  WEBPOS_CASH: 'cash',
  CASH: 'cash',
};

const posPaymentMethodTextMap: Record<string, POSPaymentMethodI18nKey> = {
  '銀行卡': 'card',
  '刷卡': 'card',
  'bank card': 'card',
  'card': 'card',
  '雲閃付': 'unionpay',
  'unionpay': 'unionpay',
  'unionpay qr': 'unionpay',
  '微信支付': 'wechat',
  '微信': 'wechat',
  'wechat pay': 'wechat',
  'wechat': 'wechat',
  '支付寶': 'alipay',
  'alipay': 'alipay',
  '支付寶（大陸）': 'mainlandAlipay',
  '大陸支付寶': 'mainlandAlipay',
  'mainland alipay': 'mainlandAlipay',
  '支付寶香港': 'alipayHK',
  '香港支付寶': 'alipayHK',
  'alipayhk': 'alipayHK',
  'alipay hk': 'alipayHK',
  '微信 / 支付寶': 'wechatAlipay',
  '微信/支付寶': 'wechatAlipay',
  'wechat pay / alipay': 'wechatAlipay',
  'wechat/alipay': 'wechatAlipay',
  '銀行轉賬': 'bankTransfer',
  '銀行轉帳': 'bankTransfer',
  '轉賬': 'bankTransfer',
  'bank transfer': 'bankTransfer',
  'transfer': 'bankTransfer',
  '支票': 'cheque',
  'cheque': 'cheque',
  'check': 'cheque',
  '現金': 'cash',
  'cash': 'cash',
};

// 1. 正規化 POS 支付方式代碼
const normalizePOSPaymentMethodCode = (value: unknown): string =>
  String(value ?? '').trim().toUpperCase();

// 2. 正規化 POS 支付方式顯示文字
const normalizePOSPaymentMethodText = (value: unknown): string =>
  String(value ?? '').trim().replace(/\s+/g, ' ').toLowerCase();

// 3. 解析 POS 支付方式語系鍵
export const resolvePOSPaymentMethodI18nKey = (value: unknown): POSPaymentMethodI18nKey | '' => {
  const codeKey = posPaymentMethodCodeMap[normalizePOSPaymentMethodCode(value)];
  if (codeKey) {
    return codeKey;
  }

  return posPaymentMethodTextMap[normalizePOSPaymentMethodText(value)] ?? '';
};

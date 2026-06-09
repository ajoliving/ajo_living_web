/*
 * 支付中心語系。
 * 1. 提供 POS 物業繳費頁面的共用文案。
 * 2. 保持支付模組文案正式且簡潔。
 */
export default {
  common: {
    noUnit: '尚未選擇單位',
    building: '屋苑',
    unit: '單位',
    loadError: '繳費資料載入失敗，請稍後再試。',
    createOrderError: '線上繳費訂單建立失敗，請稍後再試。',
    payNow: '線上繳費',
  },
  methods: {
    all: '全部方式',
    card: '銀行卡',
    unionpay: '雲閃付',
    wechat: '微信支付',
    alipay: '支付寶',
    mainlandAlipay: '支付寶（大陸）',
    alipayHK: '支付寶香港',
    wechatAlipay: '微信 / 支付寶',
    bankTransfer: '銀行轉賬',
    cheque: '支票',
    cash: '現金',
  },
};

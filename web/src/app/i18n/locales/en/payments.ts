/*
 * Payment center locale.
 * 1. Provide common copy for POS property payment pages.
 * 2. Keep payment module copy concise and formal.
 */
export default {
  common: {
    noUnit: 'No unit selected',
    building: 'Building',
    unit: 'Unit',
    loadError: 'Failed to load payment data. Please try again later.',
    createOrderError: 'Failed to create the payment order. Please try again later.',
    payNow: 'Pay online',
  },
  methods: {
    all: 'All methods',
    card: 'Bank Card',
    unionpay: 'UnionPay',
    wechat: 'WeChat Pay',
    alipay: 'Alipay',
    mainlandAlipay: 'Mainland Alipay',
    alipayHK: 'AlipayHK',
    wechatAlipay: 'WeChat Pay / Alipay',
    bankTransfer: 'Bank Transfer',
    cheque: 'Cheque',
    cash: 'Cash',
  },
};

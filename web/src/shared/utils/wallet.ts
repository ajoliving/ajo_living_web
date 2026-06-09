/*
 * AJO Point 錢包工具。
 * 1. 集中提供三模組扣費規則與積分顯示。
 * 2. 避免頁面重複硬編碼扣費金額。
 */

export const WALLET_CHARGE_COSTS = {
  secondhand: 100,
  property_sale: 1000,
  serviced_apartment: 800,
} as const;

export const WALLET_DRAFT_CHARGE_COSTS = {
  secondhand: WALLET_CHARGE_COSTS.secondhand / 2,
} as const;

export const WALLET_RENEW_CHARGE_COSTS = {
  secondhand: WALLET_CHARGE_COSTS.secondhand / 2,
} as const;

export const WALLET_CONTACT_ACCESS_COSTS = {
  secondhand: 5,
} as const;

export type WalletBizModule = keyof typeof WALLET_CHARGE_COSTS;
export type WalletChargeAction = 'publish' | 'edit' | 'republish' | 'renew';

// 1. 格式化積分數字
export const formatAjoPoints = (value: number, pointName: string, locale: string): string =>
  `${new Intl.NumberFormat(locale, { maximumFractionDigits: 0 }).format(value)} ${pointName}`;

// 2. 取得模組扣費
export const resolveWalletChargeCost = (module: WalletBizModule): number =>
  WALLET_CHARGE_COSTS[module];

// 3. 取得草稿保存扣費
export const resolveWalletDraftChargeCost = (module: 'secondhand'): number =>
  WALLET_DRAFT_CHARGE_COSTS[module];

// 4. 取得續期扣費
export const resolveWalletRenewChargeCost = (module: 'secondhand'): number =>
  WALLET_RENEW_CHARGE_COSTS[module];

// 5. 取得聯絡方式查看扣費
export const resolveWalletContactAccessCost = (module: 'secondhand'): number =>
  WALLET_CONTACT_ACCESS_COSTS[module];

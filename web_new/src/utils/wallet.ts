/*
 * AJO Point 錢包工具。
 * 1. 集中提供三模組扣費規則與積分顯示。
 * 2. 提供積分流水來源的本地化顯示。
 * 3. 避免頁面重複硬編碼扣費金額。
 * 4. 提供展示廣告公開投放時段判斷。
 */

import type { StaffRewardAdResponse, WalletTransactionResponse } from '@/model/wallet';

export const WALLET_CHARGE_COSTS = {
  secondhand: 100,
  property_sale: 1000,
  serviced_apartment: 800,
} as const;

export const WALLET_DRAFT_CHARGE_COSTS = {
  secondhand: WALLET_CHARGE_COSTS.secondhand / 2,
  property_sale: 600,
} as const;

export const WALLET_RENEW_CHARGE_COSTS = {
  secondhand: WALLET_CHARGE_COSTS.secondhand / 2,
} as const;

export type WalletBizModule = keyof typeof WALLET_CHARGE_COSTS;
export type WalletChargeAction = 'publish' | 'edit' | 'republish' | 'renew';
type WalletTransactionSource = Pick<WalletTransactionResponse, 'biz_module' | 'action_type'>;
type WalletTranslator = (key: string) => string;

const walletTransactionModules: Record<string, string> = {
  wallet: 'wallet',
  secondhand: 'secondhand',
  property_sale: 'propertySale',
  serviced_apartment: 'servicedApartment',
  payment: 'payment',
  profile: 'profile',
};

const walletTransactionActions: Record<string, string> = {
  save_draft: 'saveDraft',
  publish: 'publish',
  edit: 'edit',
  republish: 'republish',
  renew: 'renew',
  avatar_update: 'avatarUpdate',
  ad_reward: 'adReward',
  contact_access: 'contactAccess',
  recharge: 'recharge',
  pos_payment_reward: 'posPaymentReward',
  operator_grant: 'operatorGrant',
  draft_charge_correction: 'draftChargeCorrection',
};

// 1. 格式化積分數字
export const formatAjoPoints = (value: number, pointName: string, locale: string): string =>
  `${new Intl.NumberFormat(locale, { maximumFractionDigits: 0 }).format(value)} ${pointName}`;

// 2. 輸出本地化積分流水來源
export const formatWalletTransactionSource = (transaction: WalletTransactionSource, t: WalletTranslator): string => {
  const moduleKey = walletTransactionModules[transaction.biz_module];
  const actionKey = walletTransactionActions[transaction.action_type];
  const moduleLabel = moduleKey ? t(`account.wallet.transactionModules.${moduleKey}`) : transaction.biz_module;
  const actionLabel = actionKey ? t(`account.wallet.transactionActions.${actionKey}`) : transaction.action_type;
  return `${moduleLabel} · ${actionLabel}`;
};

// 3. 取得模組扣費
export const resolveWalletChargeCost = (module: WalletBizModule): number =>
  WALLET_CHARGE_COSTS[module];

// 4. 取得草稿保存扣費
export const resolveWalletDraftChargeCost = (module: keyof typeof WALLET_DRAFT_CHARGE_COSTS): number =>
  WALLET_DRAFT_CHARGE_COSTS[module];

// 5. 取得續期扣費
export const resolveWalletRenewChargeCost = (module: 'secondhand'): number =>
  WALLET_RENEW_CHARGE_COSTS[module];

// 6. 判斷展示廣告是否符合公開投放時段
export const isPublicDisplayAdActive = (
  ad: Pick<StaffRewardAdResponse, 'ends_at' | 'is_active' | 'starts_at'>,
  now = Date.now(),
): boolean => {
  if (!ad.is_active) {
    return false;
  }

  const startsAt = ad.starts_at ? Date.parse(ad.starts_at) : Number.NaN;
  if (Number.isFinite(startsAt) && startsAt > now) {
    return false;
  }

  const endsAt = ad.ends_at ? Date.parse(ad.ends_at) : Number.NaN;
  return !Number.isFinite(endsAt) || endsAt >= now;
};

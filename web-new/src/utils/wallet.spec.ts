/*
 * AJO Point 錢包工具測試。
 * 1. 驗證積分流水來源的繁中與英文顯示。
 * 2. 驗證未知歷史值保留原始代碼。
 */
import { describe, expect, it } from 'vitest';

import { formatWalletTransactionSource } from './wallet';

const zhHKMessages: Record<string, string> = {
  'account.wallet.transactionModules.propertySale': '樓盤放售',
  'account.wallet.transactionModules.wallet': '錢包',
  'account.wallet.transactionActions.saveDraft': '儲存草稿',
  'account.wallet.transactionActions.operatorGrant': '管理員發放',
};

const enMessages: Record<string, string> = {
  'account.wallet.transactionModules.propertySale': 'Property sale',
  'account.wallet.transactionActions.saveDraft': 'Save draft',
};

// 1. translate returns a test locale string while preserving unknown keys.
const translate = (messages: Record<string, string>) => (key: string): string => messages[key] ?? key;

describe('formatWalletTransactionSource', () => {
  it('formats known zh-HK transaction codes', () => {
    expect(formatWalletTransactionSource({ biz_module: 'property_sale', action_type: 'save_draft' }, translate(zhHKMessages)))
      .toBe('樓盤放售 · 儲存草稿');
    expect(formatWalletTransactionSource({ biz_module: 'wallet', action_type: 'operator_grant' }, translate(zhHKMessages)))
      .toBe('錢包 · 管理員發放');
  });

  it('formats known English transaction codes', () => {
    expect(formatWalletTransactionSource({ biz_module: 'property_sale', action_type: 'save_draft' }, translate(enMessages)))
      .toBe('Property sale · Save draft');
  });

  it('keeps unknown historical transaction codes readable', () => {
    expect(formatWalletTransactionSource({ biz_module: 'legacy_module', action_type: 'legacy_action' }, translate(zhHKMessages)))
      .toBe('legacy_module · legacy_action');
  });
});

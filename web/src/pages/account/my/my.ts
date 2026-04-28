/*
 * 會員中心 - 導覽狀態。
 * 1. 組裝會員中心側欄入口項目。
 * 2. 集中管理帳戶模組的有效頁面導覽結構。
 */
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';

import type { AccountNavItem } from '@/model/account';

// 1. 輸出會員中心導覽項目
export const useAccountNavigation = () => {
  const { t } = useI18n();

  return computed<AccountNavItem[]>(() => [
    { key: 'profile', label: t('account.sections.profile'), to: '/account/profile' },
  ]);
};

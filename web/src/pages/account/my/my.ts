/*
 * 會員中心 - 導覽狀態。
 * 1. 組裝會員中心與二手交易會員入口項目。
 * 2. 集中管理帳戶模組的有效頁面導覽結構。
 */
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';

import type { AccountNavItem } from '@/model/account';
import { useSessionStore } from '@/stores/session';

// 1. 輸出會員中心導覽項目
export const useAccountNavigation = () => {
  const { t } = useI18n();
  const sessionStore = useSessionStore();

  return computed<AccountNavItem[]>(() => {
    const items: AccountNavItem[] = [
      { key: 'profile', label: t('account.sections.profile'), to: '/account/profile' },
      { key: 'marketplace-my', label: t('nav.marketplaceMy'), to: '/account/marketplace/my' },
    ];

    if (sessionStore.currentUser.is_staff) {
      items.push({
        key: 'marketplace-settings',
        label: t('nav.marketplaceSettings'),
        to: '/account/marketplace/settings',
      });
      items.push({
        key: 'marketplace-management',
        label: t('nav.marketplaceManagement'),
        to: '/account/marketplace/management',
      });
    }

    return items;
  });
};

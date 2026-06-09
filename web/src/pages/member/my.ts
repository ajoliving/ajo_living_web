/*
 * 會員中心 - 導覽狀態。
 * 1. 組裝會員中心入口項目。
 * 2. 集中管理帳戶模組的有效頁面導覽結構。
 */
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';

import type { AccountNavItem } from '@/domains/account/model';
import { useSessionStore } from '@/app/stores/session';

// 1. 輸出會員中心導覽項目
export const useAccountNavigation = () => {
  const { t } = useI18n();
  const sessionStore = useSessionStore();

  return computed<AccountNavItem[]>(() => {
    const items: AccountNavItem[] = [
      { key: 'member', label: t('nav.member'), to: '/member' },
    ];

    if (sessionStore.currentUser.is_staff) {
      items.push({
        key: 'settings',
        label: t('nav.settings'),
        to: '/settings',
      });
      items.push({
        key: 'management',
        label: t('nav.management'),
        to: '/management',
      });
    }

    return items;
  });
};

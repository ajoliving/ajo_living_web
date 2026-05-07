/*
 * 全域頂部第二層子導航。
 * 1. 根據當前主路由輸出對應子導航項目。
 * 2. 統一管理 marketplace、properties、serviced residences 與 account 的第二層入口。
 */
import { computed } from 'vue';
import { useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';

interface HeaderSubnavItem {
  key: string;
  label: string;
  to: string;
  match: string[];
  exact?: boolean;
}

// 1. 輸出全域第二層子導航
export const useHeaderSubnav = () => {
  const route = useRoute();
  const { t } = useI18n();

  const subnavItems = computed<HeaderSubnavItem[]>(() => {
    if (route.path.startsWith('/marketplace')) {
      return [
        { key: 'marketplace-discover', label: t('nav.discover'), to: '/marketplace/discover', match: ['/marketplace/discover'] },
        { key: 'marketplace-filter', label: t('nav.filter'), to: '/marketplace/filter', match: ['/marketplace/filter'] },
        { key: 'marketplace-my', label: t('nav.my'), to: '/marketplace/my', match: ['/marketplace/my', '/marketplace/my-listings', '/marketplace/publish', '/marketplace/chat'] },
        { key: 'marketplace-settings', label: t('nav.settings'), to: '/marketplace/settings', match: ['/marketplace/settings'] },
      ];
    }

    if (route.path.startsWith('/properties')) {
      return [
        { key: 'properties-discover', label: t('nav.propertiesDiscover'), to: '/properties', match: ['/properties'], exact: true },
        { key: 'properties-sale', label: t('nav.propertiesMap'), to: '/properties/my', match: ['/properties/my'], exact: true },
        { key: 'properties-guide', label: t('nav.propertiesGuide'), to: '/properties/my/new', match: ['/properties/my/new', '/properties/my/editor'] },
      ];
    }

    if (route.path.startsWith('/serviced-residences')) {
      return [
        { key: 'serviced-discover', label: t('nav.servicedDiscover'), to: '/serviced-residences', match: ['/serviced-residences'], exact: true },
        { key: 'serviced-stay', label: t('nav.servicedStay'), to: '/serviced-residences/my', match: ['/serviced-residences/my'], exact: true },
        { key: 'serviced-guide', label: t('nav.servicedGuide'), to: '/serviced-residences/my/new', match: ['/serviced-residences/my/new', '/serviced-residences/my/editor'] },
      ];
    }

    if (route.path.startsWith('/account')) {
      return [
        { key: 'account-profile', label: t('nav.accountProfile'), to: '/account/profile', match: ['/account/profile'] },
      ];
    }

    return [];
  });

  // 2. 判斷當前第二層子導航是否啟用
  const isSubnavActive = (item: HeaderSubnavItem) =>
    item.match.some((path) =>
      item.exact
        ? route.path === path
        : route.path === path || route.path.startsWith(`${path}/`),
    );

  return {
    subnavItems,
    isSubnavActive,
  };
};

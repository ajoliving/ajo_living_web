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
      ];
    }

    if (route.path.startsWith('/properties')) {
      return [
        { key: 'properties-discover', label: t('nav.propertiesDiscover'), to: '/properties', match: ['/properties'] },
        { key: 'properties-sale', label: t('nav.propertiesSale'), to: '/properties', match: ['/properties'] },
        { key: 'properties-map', label: t('nav.propertiesMap'), to: '/properties', match: ['/properties'] },
        { key: 'properties-guide', label: t('nav.propertiesGuide'), to: '/properties', match: ['/properties'] },
      ];
    }

    if (route.path.startsWith('/serviced-residences')) {
      return [
        { key: 'serviced-discover', label: t('nav.servicedDiscover'), to: '/serviced-residences', match: ['/serviced-residences'] },
        { key: 'serviced-all', label: t('nav.servicedAll'), to: '/serviced-residences', match: ['/serviced-residences'] },
        { key: 'serviced-stay', label: t('nav.servicedStay'), to: '/serviced-residences', match: ['/serviced-residences'] },
        { key: 'serviced-guide', label: t('nav.servicedGuide'), to: '/serviced-residences', match: ['/serviced-residences'] },
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
    item.match.some((path) => route.path === path || route.path.startsWith(`${path}/`) || route.path.startsWith(path));

  return {
    subnavItems,
    isSubnavActive,
  };
};

<!--
 * 管理中心頁。
 * 1. 提供 Staff 管理中心左側導航，對齊參考 HTML 的 work-shell 結構。
 * 2. 承載 web_new 既有真實管理功能頁與 ICCTV、廣告設定面板。
 * 3. 透過 query tab 切換：家具、樓盤、住宅、ICCTV、會員、系統通知、廣告列表、廣告設定、積分流水。
-->
<script setup lang="ts">
/*
 * 管理中心頁邏輯。
 * 1. 定義管理分頁索引與導覽項目。
 * 2. 解析 query tab 並決定當前面板。
 * 3. 依分頁載入對應元件。
 * 4. 切換分頁時更新 query。
 */
import { computed, provide } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';

import ManagementMembersPage from './members/Page.vue';
import ManagementAgencyCompaniesPage from './agency_companies/Page.vue';
import ManagementPropertySalesPage from './property_sales/Page.vue';
import ManagementRewardAdsPage from './reward_ads/Page.vue';
import ManagementSecondhandListingsPage from './secondhand_listings/Page.vue';
import ManagementServicedApartmentsPage from './serviced_apartments/Page.vue';
import ManagementSystemNoticesPage from './system_notices/Page.vue';
import ManagementWalletTransactionsPage from './wallet_transactions/Page.vue';
import ManagementBuildingChatsPage from './building_chats/Page.vue';
import ManagementIcctvPanel from './widgets/ManagementIcctvPanel.vue';
import ManagementAdSettingsPanel from './widgets/ManagementAdSettingsPanel.vue';
import AppIcon from '@/shared/components/base/AppIcon.vue';
import AccountMobileBack from '@/shared/components/navigation/AccountMobileBack.vue';

// 1. 管理分頁索引
type ManagementTab =
  | 'admin-furniture'
  | 'admin-properties'
  | 'admin-homes'
  | 'admin-icctv'
  | 'admin-members'
  | 'admin-agency-companies'
  | 'admin-building-chats'
  | 'admin-notices'
  | 'admin-ads'
  | 'admin-ad-settings'
  | 'admin-points';

interface ManagementTabNavItem {
  key: ManagementTab;
  label: string;
}

const route = useRoute();
const router = useRouter();
const { t } = useI18n();
const defaultTab: ManagementTab = 'admin-furniture';
const validTabs: ManagementTab[] = [
  'admin-furniture',
  'admin-properties',
  'admin-homes',
  'admin-icctv',
  'admin-members',
  'admin-agency-companies',
  'admin-building-chats',
  'admin-notices',
  'admin-ads',
  'admin-ad-settings',
  'admin-points',
];

// 2. 建立管理中心導覽項目（順序對齊參考 HTML）
const navItems = computed<ManagementTabNavItem[]>(() => [
  { key: 'admin-furniture', label: t('marketplace.management.secondhandListings') },
  { key: 'admin-properties', label: t('marketplace.management.propertySales') },
  { key: 'admin-homes', label: t('marketplace.management.servicedApartments') },
  { key: 'admin-icctv', label: t('marketplace.management.icctv') },
  { key: 'admin-members', label: t('marketplace.management.members') },
  { key: 'admin-agency-companies', label: t('marketplace.management.agencyCompany.title') },
  { key: 'admin-building-chats', label: t('marketplace.management.buildingChats.navTitle') },
  { key: 'admin-notices', label: t('marketplace.management.systemNotices') },
  { key: 'admin-ads', label: t('marketplace.management.walletAdListSection') },
  { key: 'admin-ad-settings', label: t('marketplace.management.adSettings') },
  { key: 'admin-points', label: t('marketplace.management.walletTransactionsSection') },
]);

// 3. 解析 query tab
const resolveManagementTab = (value: unknown): ManagementTab => {
  const tab = typeof value === 'string' ? value : '';

  return validTabs.includes(tab as ManagementTab) ? tab as ManagementTab : defaultTab;
};

const activeTab = computed<ManagementTab>(() => resolveManagementTab(route.query.tab));
const mobileHome = computed(() => !route.query.tab);
const navIcons: Record<ManagementTab, 'browse' | 'home' | 'building' | 'view' | 'user' | 'settings' | 'message' | 'inbox' | 'wallet'> = {
  'admin-furniture': 'browse',
  'admin-properties': 'home',
  'admin-homes': 'building',
  'admin-icctv': 'view',
  'admin-members': 'user',
  'admin-agency-companies': 'building',
  'admin-building-chats': 'message',
  'admin-notices': 'inbox',
  'admin-ads': 'wallet',
  'admin-ad-settings': 'settings',
  'admin-points': 'wallet',
};

// 4. 依分頁載入元件
const activeComponent = computed(() => {
  switch (activeTab.value) {
    case 'admin-properties':
      return ManagementPropertySalesPage;
    case 'admin-homes':
      return ManagementServicedApartmentsPage;
    case 'admin-icctv':
      return ManagementIcctvPanel;
    case 'admin-members':
      return ManagementMembersPage;
    case 'admin-agency-companies':
      return ManagementAgencyCompaniesPage;
    case 'admin-building-chats':
      return ManagementBuildingChatsPage;
    case 'admin-notices':
      return ManagementSystemNoticesPage;
    case 'admin-ads':
      return ManagementRewardAdsPage;
    case 'admin-ad-settings':
      return ManagementAdSettingsPanel;
    case 'admin-points':
      return ManagementWalletTransactionsPage;
    case 'admin-furniture':
    default:
      return ManagementSecondhandListingsPage;
  }
});

// 5. 切換管理分頁
const activateNavItem = async (item: ManagementTabNavItem): Promise<void> => {
  const isMobile = typeof window.matchMedia === 'function' && window.matchMedia('(max-width: 1023px)').matches;
  const destination = {
    path: '/account/marketplace/management',
    query: item.key === defaultTab && !isMobile ? {} : { tab: item.key },
  };
  if (mobileHome.value && isMobile) {
    await router.push({ ...destination, state: { managementMobileHome: true } });
    return;
  }
  await router.replace(destination);
};

// 6. 判斷導覽啟用狀態
const isActiveNavItem = (item: ManagementTabNavItem): boolean =>
  activeTab.value === item.key;

const returnToMobileHome = (): void => {
  if (window.history.state?.managementMobileHome === true) {
    router.back();
    return;
  }
  void router.replace('/account/marketplace/management');
};
provide('managementReturnToHome', returnToMobileHome);
</script>

<template>
  <main class="work-shell management-shell" :class="{ 'is-mobile-home': mobileHome }">
    <aside class="work-sidebar">
      <h1>{{ t('marketplace.management.title') }}</h1>
      <p class="management-sidebar-description">{{ t('marketplace.management.description') }}</p>

      <nav class="work-nav">
        <button
          v-for="item in navItems"
          :key="item.key"
          type="button"
          class="work-nav-item"
          :class="{ on: isActiveNavItem(item) }"
          @click="activateNavItem(item)"
        >
          <AppIcon class="work-nav-icon" :name="navIcons[item.key]" :size="19" />
          {{ item.label }}
          <AppIcon class="work-nav-arrow" name="arrow-right" :size="16" />
        </button>
      </nav>
    </aside>

    <section class="work-main">
      <div v-if="!mobileHome" class="management-mobile-back">
        <AccountMobileBack label-key="marketplace.management.backToHome" />
      </div>
      <Transition
        name="subroute-slide"
        mode="out-in"
      >
        <component
          :is="activeComponent"
          :key="activeTab"
          class="subroute-transition-shell"
        />
      </Transition>
    </section>
  </main>
</template>

<style scoped>
/*
 * 管理中心頁樣式。
 * 1. 嚴格對齊參考 HTML styles.css 的 work-shell / work-sidebar / work-nav-item 樣式。
 * 2. 使用全域 tokens.css 提供的 --brand / --ink / --bdr / --sur / --font-serif 變量。
 */

/* 1. 殼層 */
.work-shell {
  display: grid;
  grid-template-columns: 240px minmax(0, 1fr);
  gap: 16px;
  max-width: var(--layout-page-max-width);
  margin: 0 auto;
  padding: 12px var(--layout-page-padding-inline) 16px;
  color: var(--ink);
}

/* 2. 側欄 */
.work-sidebar {
  position: sticky;
  top: 72px;
  align-self: start;
  border: 1px solid var(--bdr);
  border-radius: 8px;
  background: rgb(var(--color-surface));
  padding: 16px;
}

.work-sidebar h1 {
  margin: 0;
  color: var(--accent);
  font-family: var(--font-serif);
  font-size: 26px;
  font-weight: 400;
  line-height: 1.2;
}

.work-sidebar p {
  margin: 8px 0 0;
  color: var(--ink-3);
  font-size: 12px;
  line-height: 1.7;
}

/* 3. 導覽 */
.work-nav {
  display: grid;
  gap: 2px;
  margin-top: 18px;
}

.work-nav-item {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 34px;
  border: 0;
  border-radius: 0;
  background: transparent;
  color: var(--ink-2);
  cursor: pointer;
  font-family: inherit;
  font-size: 13px;
  font-weight: 500;
  padding: 8px 2px;
  text-align: left;
  transition:
    background-color 0.15s ease,
    color 0.15s ease;
}

.work-nav-item::after {
  content: '';
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  height: 2px;
  background: var(--brand);
  transform: scaleX(0);
  transform-origin: left center;
  transition: transform 0.24s ease;
}

.work-nav-item:hover {
  background: var(--brand-light);
  color: var(--ink);
}

.work-nav-item.on,
.work-nav-item:focus-visible {
  background: transparent;
  color: var(--accent);
  font-weight: 700;
  outline: none;
}

.work-nav-icon,
.work-nav-arrow,
.management-mobile-back {
  display: none;
}

.work-nav-item.on:hover,
.work-nav-item:focus-visible:hover {
  background: var(--brand-light);
}

.work-nav-item.on::after,
.work-nav-item:focus-visible::after {
  transform: scaleX(1);
}

/* 4. 主內容 */
.work-main {
  display: grid;
  gap: 12px;
  min-width: 0;
  align-content: start;
}

/* 5. 響應式 */
@media (max-width: 1023px) {
  .work-shell {
    grid-template-columns: 1fr;
    padding: 14px var(--layout-page-padding-inline);
  }

  .work-sidebar {
    position: static;
  }

  html #app .management-shell.is-mobile-home .work-main,
  html #app .management-shell:not(.is-mobile-home) .work-sidebar {
    display: none;
  }

  html #app .management-shell.is-mobile-home .work-nav {
    display: grid;
    grid-template-columns: minmax(0, 1fr);
    gap: 0;
    margin: 10px 0 0;
    padding: 0;
    overflow: visible;
  }

  html #app .management-shell.is-mobile-home .work-nav-item {
    display: flex;
    justify-content: flex-start;
    min-height: 52px;
    align-items: center;
    gap: 12px;
    border: 0;
    border-bottom: 1px solid var(--bdr);
    border-radius: 0;
    background: transparent;
    color: var(--ink);
    padding: 10px 4px;
    white-space: normal;
  }

  .management-shell.is-mobile-home .work-nav-icon,
  .management-shell.is-mobile-home .work-nav-arrow {
    display: block;
    flex: 0 0 auto;
  }

  .management-shell.is-mobile-home .work-nav-arrow {
    margin-left: auto;
    color: var(--ink-3);
  }

  .management-shell .management-sidebar-description {
    display: none;
  }

  .management-shell:not(.is-mobile-home) .work-main {
    position: relative;
  }

  .management-shell:not(.is-mobile-home) .management-mobile-back {
    display: block;
    position: absolute;
    top: 0;
    left: 0;
    z-index: 1;
  }

  .management-shell :deep(.management-list-header h1),
  .management-shell :deep(.wallet-settings-header h1),
  .management-shell :deep(.staff-list-head .staff-title),
  .management-shell :deep(.work-hero .work-title) {
    min-height: 44px;
    margin: 0;
    padding: 8px 0 0 44px;
  }

  .management-shell :deep(.furniture-list-header),
  .management-shell :deep(.reward-ad-list-header),
  .management-shell :deep(.building-chats-header) {
    border: 0;
    background: transparent;
    padding: 0;
  }

  .management-shell :deep(.building-chats-page) {
    background: transparent;
    padding: 0;
  }

  .management-shell :deep(.management-list-header .management-list-kicker),
  .management-shell :deep(.wallet-settings-header .wallet-settings-kicker),
  .management-shell :deep(.staff-list-head .staff-kicker),
  .management-shell :deep(.work-hero .work-kicker) {
    display: none;
  }
}

@media (max-width: 560px) {
  .work-shell {
    padding: 12px var(--layout-page-padding-inline);
  }
}
</style>

<!--
 * 二手交易設定外框頁。
 * 1. 提供設定頁左側本地 tab 導航。
 * 2. 承載廣告、首頁內容、發布通知與登入背景圖設定頁。
-->
<script setup lang="ts">
import { computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';

import ManagementSettingsDisplayAdsPage from '@/pages/management/settings/display-ads/Page.vue';
import ManagementSettingsHomeCarouselPage from '@/pages/management/settings/home-carousel/Page.vue';
import ManagementSettingsHomeHeroCardsPage from '@/pages/management/settings/home-hero-cards/Page.vue';
import ManagementSettingsLoginHeroPage from '@/pages/management/settings/login-hero/Page.vue';
import ManagementSettingsNoticePage from '@/pages/management/settings/notice/Page.vue';
import AppIcon from '@/shared/components/base/AppIcon.vue';

type SettingsIconName = 'palette' | 'picture' | 'layout-grid' | 'send' | 'login';
type SettingsTab =
  | 'property-ads'
  | 'furniture-ads'
  | 'serviced-residence-ads'
  | 'home-carousel'
  | 'home-hero-cards'
  | 'login-hero'
  | 'notice';

interface SettingsRouteItem {
  label: string;
  tab: SettingsTab;
  icon: SettingsIconName;
}

const route = useRoute();
const router = useRouter();
const { t } = useI18n();
const activeTab = computed<SettingsTab>(() => resolveSettingsTab(route.query.tab));
const displayAdsPageProps = computed(() => {
  switch (activeTab.value) {
    case 'furniture-ads':
      return {
        channel: 'furniture',
        titleKey: 'marketplace.settings.displayAdChannelFurniture',
      };
    case 'serviced-residence-ads':
      return {
        channel: 'serviced_apartment',
        titleKey: 'marketplace.settings.displayAdChannelServicedApartment',
      };
    default:
      return {
        channel: 'property_sale',
        titleKey: 'marketplace.settings.displayAdChannelPropertySale',
      };
  }
});
const activeComponent = computed(() => {
  switch (activeTab.value) {
    case 'home-carousel':
      return ManagementSettingsHomeCarouselPage;
    case 'home-hero-cards':
      return ManagementSettingsHomeHeroCardsPage;
    case 'login-hero':
      return ManagementSettingsLoginHeroPage;
    case 'notice':
      return ManagementSettingsNoticePage;
    case 'property-ads':
    case 'furniture-ads':
    case 'serviced-residence-ads':
    default:
      return ManagementSettingsDisplayAdsPage;
  }
});

// 1. 建立設定子路由導航
const settingsRoutes = computed<SettingsRouteItem[]>(() => [
  {
    label: t('marketplace.settings.displayAdChannelPropertySale'),
    tab: 'property-ads',
    icon: 'palette',
  },
  {
    label: t('marketplace.settings.displayAdChannelFurniture'),
    tab: 'furniture-ads',
    icon: 'palette',
  },
  {
    label: t('marketplace.settings.displayAdChannelServicedApartment'),
    tab: 'serviced-residence-ads',
    icon: 'palette',
  },
  {
    label: t('marketplace.settings.homeCarouselSection'),
    tab: 'home-carousel',
    icon: 'picture',
  },
  {
    label: t('marketplace.settings.homeHeroCardsSection'),
    tab: 'home-hero-cards',
    icon: 'layout-grid',
  },
  {
    label: t('marketplace.settings.loginHeroSection'),
    tab: 'login-hero',
    icon: 'login',
  },
  {
    label: t('marketplace.settings.noticeSection'),
    tab: 'notice',
    icon: 'send',
  },
]);

// 2. 判斷目前子路由是否啟用
const isActiveRoute = (tab: SettingsTab): boolean =>
  activeTab.value === tab;

// 3. 讀取設定 tab
const resolveSettingsTab = (value: unknown): SettingsTab => {
  const tab = typeof value === 'string' ? value : '';
  if (tab === 'ads') {
    return 'property-ads';
  }
  return ['property-ads', 'furniture-ads', 'serviced-residence-ads', 'home-carousel', 'home-hero-cards', 'login-hero', 'notice'].includes(tab)
    ? tab as SettingsTab
    : 'property-ads';
};

// 4. 切換設定 tab
const setActiveRoute = async (tab: SettingsTab): Promise<void> => {
  await router.push({ path: '/settings', query: { tab } });
};
</script>

<template>
  <main class="settings-shell">
    <aside class="settings-shell__sidebar">
      <div>
        <h1>{{ t('marketplace.settings.title') }}</h1>
        <p>{{ t('marketplace.settings.description') }}</p>
      </div>

      <nav class="settings-shell__nav">
        <button
          v-for="item in settingsRoutes"
          :key="item.tab"
          type="button"
          class="settings-shell__nav-item"
          :class="{ 'settings-shell__nav-item--active': isActiveRoute(item.tab) }"
          @click="setActiveRoute(item.tab)"
        >
          <AppIcon
            :name="item.icon"
            :size="17"
          />
          <span>{{ item.label }}</span>
        </button>
      </nav>
    </aside>

    <section class="settings-shell__content">
      <Transition
        name="subroute-slide"
        mode="out-in"
      >
        <component
          :is="activeComponent"
          :key="activeTab"
          v-bind="activeComponent === ManagementSettingsDisplayAdsPage ? displayAdsPageProps : {}"
          class="subroute-transition-shell"
        />
      </Transition>
    </section>
  </main>
</template>

<style scoped>
.settings-shell {
  display: grid;
  width: 100%;
  max-width: var(--layout-page-max-width);
  gap: 1.25rem;
  margin: 0 auto;
  padding: 1rem var(--layout-page-padding-inline) 5rem;
  color: rgb(var(--color-text));
}

.settings-shell__sidebar {
  display: grid;
  align-content: start;
  gap: 1.5rem;
  border: 1px solid rgb(var(--color-border) / 0.3);
  border-radius: 8px;
  background:
    linear-gradient(
      180deg,
      rgb(var(--color-topbar-surface) / 0.76),
      rgb(var(--color-toolbar-surface) / 0.64)
    );
  box-shadow:
    0 16px 40px rgb(15 23 42 / 0.1),
    inset 0 -1px 0 rgb(255 255 255 / 0.06);
  padding: 1rem;
  backdrop-filter: blur(28px) saturate(184%);
  -webkit-backdrop-filter: blur(28px) saturate(184%);
}

.settings-shell__sidebar h1 {
  margin: 0;
  color: rgb(var(--color-primary));
  font-family: var(--font-display);
  font-size: 1.75rem;
  font-weight: 500;
  line-height: 1.3;
}

.settings-shell__sidebar p {
  margin: 0.65rem 0 0;
  color: rgb(var(--color-text-muted));
  font-size: 0.875rem;
  line-height: 1.7;
}

.settings-shell__nav {
  display: grid;
  gap: 0.5rem;
}

.settings-shell__nav-item {
  position: relative;
  display: inline-flex;
  min-height: 2.75rem;
  width: 100%;
  align-items: center;
  gap: 0.65rem;
  border: 1px solid rgb(var(--color-border) / 0.24);
  border-radius: 8px;
  background:
    linear-gradient(
      180deg,
      rgb(var(--color-topbar-surface) / 0.62),
      rgb(var(--color-toolbar-surface) / 0.48)
    );
  padding: 0.65rem 0.85rem;
  color: rgb(var(--color-text) / 0.78);
  font-size: 0.9375rem;
  font-weight: 700;
  line-height: 1;
  text-decoration: none;
  box-shadow:
    0 12px 28px rgb(15 23 42 / 0.07),
    inset 0 1px 0 rgb(255 255 255 / 0.08);
  transition:
    border-color 0.2s ease,
    background 0.2s ease,
    color 0.2s ease;
}

.settings-shell__nav-item:hover {
  border-color: rgb(var(--color-border) / 0.38);
  color: rgb(var(--color-text));
}

.settings-shell__nav-item--active,
.settings-shell__nav-item--active:hover {
  border-color: rgb(var(--color-primary) / 0.28);
  background:
    linear-gradient(
      180deg,
      rgb(var(--color-topbar-surface) / 0.86),
      rgb(var(--color-primary-soft) / 0.24)
    );
  color: color-mix(in srgb, rgb(var(--color-primary)) 78%, rgb(var(--color-text)) 22%);
  box-shadow:
    0 14px 32px rgb(var(--color-primary) / 0.1),
    inset 0 1px 0 rgb(255 255 255 / 0.1);
}

.settings-shell__content {
  min-width: 0;
}

@media (min-width: 1024px) {
  .settings-shell {
    grid-template-columns: 15.75rem minmax(0, 1fr);
    align-items: start;
  }

  .settings-shell__sidebar {
    position: sticky;
    top: calc(var(--app-header-offset, 0rem) + 2rem);
  }
}

@media (max-width: 767px) {
  .settings-shell {
    padding: 1rem var(--layout-page-padding-inline) 4rem;
  }
}
</style>

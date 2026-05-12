<!--
 * 二手交易管理外框頁。
 * 1. 提供 Staff 管理子路由左側導航。
 * 2. 承載營運操作與四類管理列表頁。
-->
<script setup lang="ts">
import { computed } from 'vue';
import { RouterLink, RouterView, useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';

import AppIcon from '@/shared/components/base/AppIcon.vue';

type ManagementIconName = 'layout-list' | 'plus-square' | 'browse' | 'home' | 'user';

interface ManagementRouteItem {
  label: string;
  to: string;
  icon: ManagementIconName;
}

const route = useRoute();
const { t } = useI18n();

// 1. 建立管理子路由導航
const managementRoutes = computed<ManagementRouteItem[]>(() => [
  {
    label: t('marketplace.management.secondhandListings'),
    to: '/account/marketplace/management/secondhand-listings',
    icon: 'layout-list',
  },
  {
    label: t('marketplace.management.propertySales'),
    to: '/account/marketplace/management/property-sales',
    icon: 'home',
  },
  {
    label: t('marketplace.management.members'),
    to: '/account/marketplace/management/members',
    icon: 'user',
  },
  {
    label: t('marketplace.management.servicedApartments'),
    to: '/account/marketplace/management/serviced-apartments',
    icon: 'browse',
  },
  {
    label: t('marketplace.management.walletAdPublishSection'),
    to: '/account/marketplace/management/reward-ad-editor',
    icon: 'plus-square',
  },
  {
    label: t('marketplace.management.walletAdListSection'),
    to: '/account/marketplace/management/reward-ads',
    icon: 'browse',
  },
  {
    label: t('marketplace.management.walletTransactionsSection'),
    to: '/account/marketplace/management/wallet-transactions',
    icon: 'layout-list',
  },
]);

// 2. 判斷目前子路由是否啟用
const isActiveRoute = (path: string): boolean =>
  route.path === path || route.path.startsWith(`${path}/`);
</script>

<template>
  <main class="management-shell">
    <aside class="management-shell__sidebar">
      <div>
        <h1>{{ t('marketplace.management.title') }}</h1>
        <p>{{ t('marketplace.management.description') }}</p>
      </div>

      <nav class="management-shell__nav">
        <RouterLink
          v-for="item in managementRoutes"
          :key="item.to"
          :to="item.to"
          class="management-shell__nav-item"
          :class="{ 'management-shell__nav-item--active': isActiveRoute(item.to) }"
        >
          <AppIcon
            :name="item.icon"
            :size="17"
          />
          <span>{{ item.label }}</span>
        </RouterLink>
      </nav>
    </aside>

    <section class="management-shell__content">
      <RouterView v-slot="{ Component }">
        <Transition
          name="subroute-slide"
          mode="out-in"
        >
          <component
            :is="Component"
            class="subroute-transition-shell"
          />
        </Transition>
      </RouterView>
    </section>
  </main>
</template>

<style scoped>
.management-shell {
  display: grid;
  width: 100%;
  max-width: var(--layout-page-max-width);
  gap: 1.25rem;
  margin: 0 auto;
  padding: 1rem var(--layout-page-padding-inline) 5rem;
  color: rgb(var(--color-text));
}

.management-shell__sidebar {
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

.management-shell__sidebar h1 {
  margin: 0;
  color: rgb(var(--color-primary));
  font-family: var(--font-display);
  font-size: 1.75rem;
  font-weight: 500;
  line-height: 1.3;
}

.management-shell__sidebar p {
  margin: 0.65rem 0 0;
  color: rgb(var(--color-text-muted));
  font-size: 0.875rem;
  line-height: 1.7;
}

.management-shell__nav {
  display: grid;
  gap: 0.5rem;
}

.management-shell__nav-item {
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

.management-shell__nav-item:hover {
  border-color: rgb(var(--color-border) / 0.38);
  color: rgb(var(--color-text));
}

.management-shell__nav-item--active,
.management-shell__nav-item--active:hover {
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

.management-shell__content {
  min-width: 0;
}

@media (min-width: 1024px) {
  .management-shell {
    grid-template-columns: 15.75rem minmax(0, 1fr);
    align-items: start;
  }

  .management-shell__sidebar {
    position: sticky;
    top: calc(var(--app-header-offset, 0rem) + 2rem);
  }
}

@media (max-width: 767px) {
  .management-shell {
    padding: 1rem var(--layout-page-padding-inline) 4rem;
  }
}
</style>

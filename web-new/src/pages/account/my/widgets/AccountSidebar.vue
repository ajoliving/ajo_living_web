<!--
 * 會員中心側欄。
 * 1. 顯示當前會員摘要與模組導覽。
 * 2. 提供帳戶模組各子頁的固定入口。
-->
<script setup lang="ts">
import { computed } from 'vue';
import { RouterLink, useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';

import AppIcon from '@/shared/components/base/AppIcon.vue';
import { useSessionStore } from '@/stores/session';

import { useAccountNavigation } from '../my';

type AccountSidebarIconName =
  | 'building'
  | 'home'
  | 'layout-list'
  | 'message'
  | 'palette'
  | 'star'
  | 'user'
  | 'wallet';

const route = useRoute();
const { t } = useI18n();
const sessionStore = useSessionStore();
const navigationItems = useAccountNavigation();

// 1. 判斷目前導覽是否為啟用狀態
const isActive = (to: string, exact = false) =>
  exact ? route.path === to : route.path === to || route.path.startsWith(`${to}/`);

// 2. 組合側欄眉標
const accountRoleLabel = computed(() =>
  sessionStore.currentUser.member_type || t('account.overview.memberLabel'),
);

// 3. 取得導覽項目圖示
const resolveNavigationIcon = (key: string): AccountSidebarIconName => {
  const iconMap: Record<string, AccountSidebarIconName> = {
    profile: 'user',
    wallet: 'wallet',
    chat: 'message',
    'property-sales': 'home',
    'serviced-residences': 'building',
    listings: 'layout-list',
    favorites: 'star',
    'marketplace-settings': 'palette',
    'marketplace-management': 'building',
  };

  return iconMap[key] ?? 'layout-list';
};
</script>

<template>
  <aside class="account-sidebar">
    <div class="account-sidebar__header">
      <h1>{{ t('account.title') }}</h1>
      <p>{{ sessionStore.currentUser.display_name }}</p>

      <div class="account-sidebar__meta">
        <span>{{ t('account.overview.memberLabel') }}</span>
        <strong>{{ accountRoleLabel }}</strong>
      </div>
    </div>

    <nav class="account-sidebar__nav">
      <RouterLink
        v-for="item in navigationItems"
        :key="item.key"
        :to="item.to"
        class="account-sidebar__nav-item"
        :class="{ 'account-sidebar__nav-item--active': isActive(item.to, item.exact) }"
      >
        <AppIcon
          :name="resolveNavigationIcon(item.key)"
          :size="17"
        />
        <span>{{ item.label }}</span>
      </RouterLink>
    </nav>
  </aside>
</template>

<style scoped>
.account-sidebar {
  display: grid;
  align-content: start;
  gap: 1rem;
  min-height: calc(100vh - var(--app-header-offset, 48px));
  border-right: 1px solid rgb(var(--color-border));
  background: rgb(var(--color-surface));
  padding: 1.5rem 1rem 1rem 0;
}

.account-sidebar__header h1 {
  margin: 0;
  color: rgb(var(--color-text));
  font-family: var(--font-display);
  font-size: 1.8rem;
  font-weight: 400;
  line-height: 1.1;
}

.account-sidebar__header p {
  margin: 0.55rem 0 0;
  color: rgb(var(--color-text-muted));
  font-size: 0.8rem;
  line-height: 1.65;
}

.account-sidebar__meta {
  display: grid;
  gap: 0.35rem;
  margin-top: 1rem;
  border-left: 3px solid rgb(var(--color-primary));
  background: rgb(var(--color-surface-muted));
  padding: 0.8rem 0.85rem;
}

.account-sidebar__meta span {
  color: rgb(var(--color-text-muted));
  font-size: 0.75rem;
  font-weight: 700;
  line-height: 1;
}

.account-sidebar__meta strong {
  color: rgb(var(--color-text));
  font-size: 0.9375rem;
  font-weight: 700;
  line-height: 1.4;
}

.account-sidebar__nav {
  display: grid;
  gap: 0.5rem;
}

.account-sidebar__nav-item {
  display: inline-flex;
  min-height: 2.6rem;
  width: 100%;
  align-items: center;
  gap: 0.65rem;
  border: 1px solid transparent;
  border-radius: 2px;
  background: rgb(var(--color-surface));
  padding: 0.62rem 0.8rem;
  color: rgb(var(--color-text-muted));
  font-size: 0.82rem;
  font-weight: 500;
  line-height: 1;
  text-decoration: none;
  transition:
    background 0.2s ease,
    border-color 0.2s ease,
    color 0.2s ease,
    transform 0.18s ease;
}

.account-sidebar__nav-item span {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.account-sidebar__nav-item:hover {
  border-color: rgb(var(--color-border));
  color: rgb(var(--color-text));
  transform: translateX(2px);
}

.account-sidebar__nav-item--active,
.account-sidebar__nav-item--active:hover {
  border-color: rgb(var(--color-text));
  background: rgb(var(--color-text));
  color: #ffffff;
  transform: none;
}

@media (min-width: 1024px) {
  .account-sidebar {
    position: sticky;
    top: var(--app-header-offset, 48px);
  }
}

@media (max-width: 767px) {
  .account-sidebar {
    min-height: auto;
    gap: 1rem;
    padding: 1rem 0 0;
    border-right: 0;
    border-bottom: 1px solid rgb(var(--color-border));
  }

  .account-sidebar__nav {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .account-sidebar__nav-item {
    min-height: 2.5rem;
  }
}
</style>

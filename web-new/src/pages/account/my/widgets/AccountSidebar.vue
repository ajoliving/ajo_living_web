<!--
 * 會員中心側欄。
 * 1. 顯示當前會員摘要與模組導覽。
 * 2. 提供帳戶模組各子頁的固定入口。
-->
<script setup lang="ts">
import { computed } from 'vue';
import { RouterLink, useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';

import BaseAvatar from '@/shared/components/base/BaseAvatar.vue';
import { useSessionStore } from '@/stores/session';

import { useAccountNavigation } from '../my';

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
</script>

<template>
  <aside class="account-sidebar">
    <div class="border-b border-border pb-5">
      <div class="flex items-center gap-4">
        <BaseAvatar
          :src="sessionStore.currentUser.avatar_url"
          :name="sessionStore.currentUser.display_name"
          :size="56"
        />
        <div class="min-w-0">
          <p class="truncate font-display text-[24px] font-medium leading-[1.4] text-text">
            {{ sessionStore.currentUser.display_name }}
          </p>
          <p class="truncate text-[14px] leading-[1.6] text-text-muted">
            {{ t('nav.memberCenter') }}
          </p>
        </div>
      </div>

      <div class="mt-5 border-l-2 border-primary bg-surface-muted px-4 py-3">
        <p class="text-[12px] font-semibold uppercase leading-none tracking-[0.1em] text-text-muted">
          {{ t('account.overview.memberLabel') }}
        </p>
        <p class="mt-2 text-[15px] font-semibold leading-[1.5] text-text">
          {{ accountRoleLabel }}
        </p>
      </div>
    </div>

    <nav class="space-y-3">
      <RouterLink
        v-for="item in navigationItems"
        :key="item.key"
        :to="item.to"
        class="account-sidebar-link"
        :class="isActive(item.to, item.exact) ? 'account-sidebar-link-active' : 'account-sidebar-link-idle'"
      >
        {{ item.label }}
      </RouterLink>
    </nav>
  </aside>
</template>

<style scoped>
.account-sidebar {
  --subroute-nav-active-color: color-mix(in srgb, rgb(var(--color-primary)) 78%, rgb(var(--color-text)) 22%);
  --subroute-nav-active-shadow: 0 0 10px rgb(var(--color-primary) / 0.16);
  --subroute-nav-underline: color-mix(in srgb, rgb(var(--color-primary)) 88%, rgb(var(--color-text)) 12%);
  display: flex;
  width: 16rem;
  flex-shrink: 0;
  flex-direction: column;
  gap: 2rem;
  border: 1px solid rgb(var(--color-border));
  border-radius: 3px;
  background: rgb(var(--color-surface));
  padding: 1rem;
  overflow: visible;
}

.account-sidebar-link {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: space-between;
  border: 1px solid rgb(var(--color-border));
  border-radius: 3px;
  background: rgb(var(--color-surface-raised));
  padding: 0.9rem 0.85rem;
  font-size: 1rem;
  font-weight: 600;
  line-height: 1.6;
  transition:
    background 0.2s ease,
    border-color 0.2s ease,
    color 0.2s ease,
    padding-left 0.2s ease;
}

.account-sidebar-link-active {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary-soft));
  color: var(--subroute-nav-active-color);
  text-shadow: var(--subroute-nav-active-shadow);
}

.account-sidebar-link-idle {
  color: rgb(var(--color-text) / 0.78);
}

.account-sidebar-link-idle:hover {
  border-color: rgb(var(--color-border) / 0.38);
  color: rgb(var(--color-text));
  padding-left: 1rem;
}

.account-sidebar-link::after {
  position: absolute;
  left: 0.85rem;
  right: 0.85rem;
  bottom: 0.58rem;
  height: 1.5px;
  border-radius: 999px;
  background: var(--subroute-nav-underline);
  transform: scaleX(0);
  transform-origin: left center;
  opacity: 0;
  transition:
    transform 0.22s ease,
    opacity 0.22s ease;
  content: '';
}

.account-sidebar-link-active::after {
  transform: scaleX(1);
  opacity: 1;
}

@media (max-width: 767px) {
  .account-sidebar {
    width: 100%;
    overflow: visible;
    padding-right: 0;
  }
}
</style>

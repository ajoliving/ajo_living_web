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
    <div class="border-b border-[#E2E3E1] pb-5">
      <div class="flex items-center gap-4">
        <BaseAvatar
          :src="sessionStore.currentUser.avatar_url"
          :name="sessionStore.currentUser.display_name"
          :size="56"
        />
        <div class="min-w-0">
          <p class="truncate font-display text-[24px] font-medium leading-[1.4] text-[#1A1C1B]">
            {{ sessionStore.currentUser.display_name }}
          </p>
          <p class="truncate text-[14px] leading-[1.6] text-[#717878]">
            {{ sessionStore.currentUser.primary_community.name }}
          </p>
        </div>
      </div>

      <div class="mt-5 border-l-2 border-[#002727] bg-[#F4F4F2] px-4 py-3">
        <p class="text-[12px] font-semibold uppercase leading-none tracking-[0.1em] text-[#717878]">
          {{ t('account.overview.memberLabel') }}
        </p>
        <p class="mt-2 text-[15px] font-semibold leading-[1.5] text-[#1A1C1B]">
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
  display: flex;
  width: 16rem;
  height: 100%;
  flex-shrink: 0;
  flex-direction: column;
  gap: 2rem;
  overflow-y: auto;
  overscroll-behavior: contain;
  padding-right: 0.5rem;
  scrollbar-width: none;
  scrollbar-gutter: stable;
}

.account-sidebar::-webkit-scrollbar {
  display: none;
}

.account-sidebar-link {
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-top: 1px solid #e2e3e1;
  padding: 1rem 0;
  font-size: 1rem;
  font-weight: 600;
  line-height: 1.6;
  transition:
    color 0.2s ease,
    padding-left 0.2s ease;
}

.account-sidebar-link-active {
  color: #002727;
  padding-left: 0.5rem;
}

.account-sidebar-link-idle {
  color: #414848;
}

.account-sidebar-link-idle:hover {
  color: #1a1c1b;
  padding-left: 0.5rem;
}

@media (max-width: 767px) {
  .account-sidebar {
    width: 100%;
    height: auto;
    overflow: visible;
    padding-right: 0;
  }
}
</style>

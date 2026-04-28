<!--
 * 二手交易我的中心頁。
 * 1. 提供我的模組左側內置導航。
 * 2. 承載我的帖子、新增帖子、訂單、收藏、個人資料與聊天子頁。
-->
<script setup lang="ts">
import { computed } from 'vue';
import { RouterLink, RouterView, useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';

import AppIcon from '@/shared/components/base/AppIcon.vue';

interface MyNavItem {
  key: string;
  label: string;
  to: string;
  icon: 'browse' | 'plus-square' | 'inbox' | 'check-circle' | 'user' | 'message';
  match: string[];
}

const route = useRoute();
const { t } = useI18n();

const navItems = computed<MyNavItem[]>(() => [
  {
    key: 'listings',
    label: t('marketplace.myHub.publishedListings'),
    to: '/marketplace/my/listings',
    icon: 'browse',
    match: ['/marketplace/my/listings', '/marketplace/my/editor', '/marketplace/my/listing', '/marketplace/my/preview'],
  },
  {
    key: 'new',
    label: t('marketplace.myHub.newListing'),
    to: '/marketplace/my/new',
    icon: 'plus-square',
    match: ['/marketplace/my/new'],
  },
  {
    key: 'orders',
    label: t('marketplace.myHub.orders'),
    to: '/marketplace/my/orders',
    icon: 'inbox',
    match: ['/marketplace/my/orders'],
  },
  {
    key: 'favorites',
    label: t('marketplace.myHub.favorites'),
    to: '/marketplace/my/favorites',
    icon: 'check-circle',
    match: ['/marketplace/my/favorites'],
  },
  {
    key: 'profile',
    label: t('marketplace.myHub.profile'),
    to: '/marketplace/my/profile',
    icon: 'user',
    match: ['/marketplace/my/profile'],
  },
  {
    key: 'chat',
    label: t('marketplace.myHub.chat'),
    to: '/marketplace/my/chat',
    icon: 'message',
    match: ['/marketplace/my/chat'],
  },
]);

// 1. 判斷左側導航是否啟用
const isNavActive = (item: MyNavItem): boolean =>
  item.match.some((path) => route.path === path || route.path.startsWith(`${path}/`));
</script>

<template>
  <main class="marketplace-my-hub">
    <aside class="marketplace-my-hub__sidebar">
      <div>
        <p class="marketplace-my-hub__kicker">
          {{ t('marketplace.myHub.title') }}
        </p>
        <h1>{{ t('marketplace.myHub.title') }}</h1>
        <p class="marketplace-my-hub__description">
          {{ t('marketplace.myHub.subtitle') }}
        </p>
      </div>

      <nav class="marketplace-my-hub__nav">
        <RouterLink
          v-for="item in navItems"
          :key="item.key"
          :to="item.to"
          class="marketplace-my-hub__nav-item"
          :class="isNavActive(item) ? 'marketplace-my-hub__nav-item--active' : ''"
        >
          <AppIcon
            :name="item.icon"
            :size="17"
          />
          <span>{{ item.label }}</span>
        </RouterLink>
      </nav>
    </aside>

    <section class="marketplace-my-hub__content">
      <RouterView />
    </section>
  </main>
</template>

<style scoped>
.marketplace-my-hub {
  display: grid;
  width: 100%;
  max-width: 1280px;
  gap: 1.25rem;
  margin: 0 auto;
  padding: 1rem 2rem 5rem;
  color: #1a1c1b;
}

.marketplace-my-hub__sidebar {
  display: grid;
  align-content: start;
  gap: 1.5rem;
  padding: 0.25rem 0.75rem 0.25rem 0;
}

.marketplace-my-hub__kicker {
  margin: 0;
  color: #717878;
  font-size: 0.75rem;
  font-weight: 800;
  letter-spacing: 0.1em;
  line-height: 1;
  text-transform: uppercase;
}

.marketplace-my-hub__sidebar h1 {
  margin: 0.75rem 0 0;
  color: #002727;
  font-family: var(--font-display);
  font-size: 1.75rem;
  font-weight: 500;
  line-height: 1.3;
}

.marketplace-my-hub__description {
  margin: 0.65rem 0 0;
  color: #717878;
  font-size: 0.875rem;
  line-height: 1.7;
}

.marketplace-my-hub__nav {
  display: grid;
  gap: 0.5rem;
}

.marketplace-my-hub__nav-item {
  display: inline-flex;
  min-height: 2.75rem;
  align-items: center;
  gap: 0.65rem;
  border: 1px solid transparent;
  border-radius: 0.75rem;
  padding: 0.65rem 0.85rem;
  color: #414848;
  font-size: 0.9375rem;
  font-weight: 700;
  line-height: 1;
  transition:
    border-color 0.2s ease,
    background-color 0.2s ease,
    color 0.2s ease;
}

.marketplace-my-hub__nav-item:hover {
  border-color: rgb(var(--color-border));
  background: rgb(var(--color-surface));
  color: #002727;
}

.marketplace-my-hub__nav-item--active {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary));
  color: #ffffff;
}

.marketplace-my-hub__content {
  min-width: 0;
}

@media (min-width: 1024px) {
  .marketplace-my-hub {
    grid-template-columns: 15.75rem minmax(0, 1fr);
    align-items: start;
  }

  .marketplace-my-hub__sidebar {
    position: sticky;
    top: calc(var(--app-header-offset, 0rem) + 2rem);
  }
}

@media (max-width: 767px) {
  .marketplace-my-hub {
    padding: 1rem 1.25rem 4rem;
  }
}
</style>

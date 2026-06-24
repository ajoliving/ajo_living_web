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
  icon: 'browse' | 'plus-square' | 'inbox' | 'star' | 'user' | 'message';
  match: string[];
}

const route = useRoute();
const { t } = useI18n();

const navItems = computed<MyNavItem[]>(() => [
  {
    key: 'chat',
    label: t('marketplace.myHub.chat'),
    to: '/account/marketplace/my/chat',
    icon: 'message',
    match: ['/account/marketplace/my/chat'],
  },
  {
    key: 'listings',
    label: t('marketplace.myHub.publishedListings'),
    to: '/account/marketplace/my/listings',
    icon: 'browse',
    match: ['/account/marketplace/my/listings', '/account/marketplace/my/editor', '/account/marketplace/my/listing', '/account/marketplace/my/preview'],
  },
  {
    key: 'new',
    label: t('marketplace.myHub.newListing'),
    to: '/account/marketplace/my/new',
    icon: 'plus-square',
    match: ['/account/marketplace/my/new'],
  },
  {
    key: 'orders',
    label: t('marketplace.myHub.orders'),
    to: '/account/marketplace/my/orders',
    icon: 'inbox',
    match: ['/account/marketplace/my/orders'],
  },
  {
    key: 'favorites',
    label: t('marketplace.myHub.favorites'),
    to: '/account/marketplace/my/favorites',
    icon: 'star',
    match: ['/account/marketplace/my/favorites'],
  },
  {
    key: 'profile',
    label: t('marketplace.myHub.profile'),
    to: '/account/marketplace/my/profile',
    icon: 'user',
    match: ['/account/marketplace/my/profile'],
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
.marketplace-my-hub {
  display: grid;
  width: 100%;
  max-width: 1180px;
  gap: 1rem;
  margin: 0 auto;
  padding: 0.75rem var(--layout-page-padding-inline) 4rem;
  color: rgb(var(--color-text));
}

.marketplace-my-hub__sidebar {
  --subroute-nav-active-color: color-mix(in srgb, rgb(var(--color-primary)) 78%, rgb(var(--color-text)) 22%);
  display: grid;
  align-content: start;
  gap: 1.25rem;
  border: 1px solid rgb(var(--color-border));
  border-radius: 2px;
  background: rgb(var(--color-surface));
  padding: 0.95rem;
}

.marketplace-my-hub__kicker {
  margin: 0;
  color: rgb(var(--color-text-muted));
  font-size: 0.7rem;
  font-weight: 800;
  letter-spacing: 0.1em;
  line-height: 1;
  text-transform: uppercase;
}

.marketplace-my-hub__sidebar h1 {
  margin: 0.75rem 0 0;
  color: rgb(var(--color-primary));
  font-family: var(--font-display);
  font-size: 1.55rem;
  font-weight: 500;
  line-height: 1.22;
}

.marketplace-my-hub__description {
  margin: 0.55rem 0 0;
  color: rgb(var(--color-text-muted));
  font-size: 0.8125rem;
  line-height: 1.65;
}

.marketplace-my-hub__nav {
  display: grid;
  gap: 0.4rem;
}

.marketplace-my-hub__nav-item {
  position: relative;
  display: inline-flex;
  min-height: 2.45rem;
  width: 100%;
  align-items: center;
  gap: 0.55rem;
  border: 1px solid rgb(var(--color-border));
  border-radius: 2px;
  background: rgb(var(--color-surface-raised));
  padding: 0.55rem 0.72rem;
  color: rgb(var(--color-text) / 0.78);
  font-size: 0.875rem;
  font-weight: 700;
  line-height: 1;
  transition:
    border-color 0.2s ease,
    background 0.2s ease,
    color 0.2s ease;
}

.marketplace-my-hub__nav-item:hover {
  border-color: rgb(var(--color-border) / 0.38);
  color: rgb(var(--color-text));
}

.marketplace-my-hub__nav-item--active,
.marketplace-my-hub__nav-item--active:hover {
  border-color: rgb(var(--color-primary));
  background: rgb(var(--color-primary-soft));
  color: var(--subroute-nav-active-color);
}

.marketplace-my-hub__content {
  min-width: 0;
}

@media (min-width: 1024px) {
  .marketplace-my-hub {
    grid-template-columns: 15rem minmax(0, 1fr);
    align-items: start;
  }

  .marketplace-my-hub__sidebar {
    position: sticky;
    top: calc(var(--app-header-offset, 0rem) + 1rem);
  }
}

@media (max-width: 767px) {
  .marketplace-my-hub {
    padding: 1rem var(--layout-page-padding-inline) 4rem;
  }
}
</style>
